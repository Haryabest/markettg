import asyncio
import base64
import logging
from io import BytesIO

from aiogram import Bot
from telethon import TelegramClient
from telethon.sessions import MemorySession, StringSession
from telethon.tl.functions.payments import GetStarGiftsRequest

from bot.catalog_client import CatalogClient
from bot.config import settings

logger = logging.getLogger(__name__)

IMAGE_BATCH_SIZE = 20
IMAGE_CONCURRENCY = 4


def _optional_int(value):
    return int(value) if value is not None else None


def _gift_title(star_gift) -> str | None:
    title = getattr(star_gift, "title", None)
    if title and str(title).strip():
        return str(title).strip()
    for attr in getattr(getattr(star_gift, "sticker", None), "attributes", None) or []:
        alt = getattr(attr, "alt", None)
        if alt and str(alt).strip():
            return str(alt).strip()
    return None


def _sticker_emoji(sticker) -> str | None:
    if not sticker:
        return None
    for attr in getattr(sticker, "attributes", None) or []:
        alt = getattr(attr, "alt", None)
        if alt:
            return str(alt)
    return None


async def _preview_base64_bot(bot: Bot, sticker) -> str | None:
    if not sticker:
        return None
    try:
        buf = BytesIO()
        await bot.download(sticker, destination=buf)
        data = buf.getvalue()
        if data:
            return base64.b64encode(data).decode("ascii")
    except Exception:
        pass
    thumb = getattr(sticker, "thumbnail", None) or getattr(sticker, "thumb", None)
    if thumb:
        try:
            buf = BytesIO()
            await bot.download(thumb, destination=buf)
            data = buf.getvalue()
            if data:
                return base64.b64encode(data).decode("ascii")
        except Exception as exc:
            logger.debug("gift preview download failed: %s", exc)
    return None


async def _preview_base64_telethon(client: TelegramClient, star_gift) -> str | None:
    sticker = getattr(star_gift, "sticker", None)
    if not sticker:
        return None
    try:
        buf = BytesIO()
        await client.download_media(sticker, file=buf, thumb=-1)
        data = buf.getvalue()
        if data:
            return base64.b64encode(data).decode("ascii")
    except Exception as exc:
        logger.debug("gift preview thumb failed: %s", exc)
    return None


async def _sticker_base64_telethon(client: TelegramClient, star_gift) -> str | None:
    sticker = getattr(star_gift, "sticker", None)
    if not sticker:
        return None
    try:
        buf = BytesIO()
        await client.download_media(sticker, file=buf)
        data = buf.getvalue()
        if data:
            return base64.b64encode(data).decode("ascii")
    except Exception as exc:
        logger.debug("gift sticker download failed: %s", exc)
    return None


def _star_count(gift) -> int:
    for attr in ("stars", "convert_stars", "resell_min_stars", "upgrade_stars"):
        value = getattr(gift, attr, None)
        if value:
            return int(value)
    return 0


def _include_gift(star_gift) -> bool:
    if getattr(star_gift, "sold_out", False) or getattr(star_gift, "limited", False):
        return True
    return _star_count(star_gift) > 0


def _gift_entry(star_gift, *, image_base64=None, sticker_base64=None) -> dict:
    gift_id = str(star_gift.id)
    title = _gift_title(star_gift)
    emoji = _sticker_emoji(getattr(star_gift, "sticker", None))
    entry = {
        "telegram_gift_id": gift_id,
        "star_count": _star_count(star_gift),
        "emoji": emoji,
        "sticker_thumb_file_id": None,
        "sticker_file_id": None,
        "image_base64": image_base64,
        "sticker_base64": sticker_base64,
        "total_count": _optional_int(getattr(star_gift, "availability_total", None)),
        "remaining_count": _optional_int(getattr(star_gift, "availability_remains", None)),
    }
    if title:
        entry["title"] = title
    return entry


async def _catalog_payload(client: TelegramClient) -> tuple[dict[str, dict], dict[str, object]]:
    catalog_result = await client(GetStarGiftsRequest(hash=0))
    gifts_by_id: dict[str, object] = {str(g.id): g for g in catalog_result.gifts}
    payload: dict[str, dict] = {}
    for gift_id, star_gift in gifts_by_id.items():
        if not _include_gift(star_gift):
            continue
        entry = _gift_entry(star_gift)
        if entry["star_count"] <= 0:
            entry["star_count"] = 1
        payload[gift_id] = entry
    return payload, gifts_by_id


async def _enrich_images(
    client: TelegramClient,
    gifts_by_id: dict[str, object],
    gift_ids: list[str],
) -> dict[str, dict]:
    sem = asyncio.Semaphore(IMAGE_CONCURRENCY)
    enriched: dict[str, dict] = {}

    async def process_one(gift_id: str) -> tuple[str, dict | None]:
        star_gift = gifts_by_id.get(gift_id)
        if not star_gift:
            return gift_id, None
        async with sem:
            preview = await _preview_base64_telethon(client, star_gift)
            sticker = await _sticker_base64_telethon(client, star_gift)
        if not preview and not sticker:
            return gift_id, None
        return gift_id, _gift_entry(star_gift, image_base64=preview, sticker_base64=sticker)

    for start in range(0, len(gift_ids), IMAGE_BATCH_SIZE):
        batch_ids = gift_ids[start : start + IMAGE_BATCH_SIZE]
        entries = await asyncio.gather(*(process_one(gid) for gid in batch_ids))
        for gift_id, entry in entries:
            if entry:
                enriched[gift_id] = entry
        logger.info("gift images prepared batch %s-%s", start + 1, start + len(batch_ids))

    return enriched


async def sync_telegram_gifts(bot: Bot, catalog: CatalogClient) -> dict | None:
    payload_by_id: dict[str, dict] = {}
    gifts_by_id: dict[str, object] = {}

    if settings.telegram_api_id and settings.telegram_api_hash and (settings.telegram_session or settings.bot_token):
        session = StringSession(settings.telegram_session) if settings.telegram_session else MemorySession()
        client = TelegramClient(session, settings.telegram_api_id, settings.telegram_api_hash)
        try:
            if settings.telegram_session:
                await client.start()
            else:
                await client.start(bot_token=settings.bot_token)
            payload_by_id, gifts_by_id = await _catalog_payload(client)
            logger.info("star gifts catalog fetched: %s items", len(payload_by_id))
            if payload_by_id:
                enriched = await _enrich_images(client, gifts_by_id, list(payload_by_id.keys()))
                for gift_id, entry in enriched.items():
                    payload_by_id[gift_id] = {**payload_by_id.get(gift_id, {}), **entry}
        except Exception as exc:
            logger.warning("GetStarGifts failed: %s", exc)
        finally:
            if client.is_connected():
                await client.disconnect()

    try:
        gifts_result = await bot.get_available_gifts()
        for gift in gifts_result.gifts:
            gift_id = str(gift.id)
            sticker = gift.sticker
            thumb_file_id = None
            sticker_file_id = None
            emoji = None
            if sticker:
                emoji = sticker.emoji
                sticker_file_id = sticker.file_id
                thumb = getattr(sticker, "thumbnail", None) or getattr(sticker, "thumb", None)
                if thumb is not None:
                    thumb_file_id = getattr(thumb, "file_id", None)

            available = {
                "telegram_gift_id": gift_id,
                "star_count": gift.star_count,
                "emoji": emoji,
                "sticker_thumb_file_id": thumb_file_id,
                "sticker_file_id": sticker_file_id,
                "image_base64": await _preview_base64_bot(bot, sticker),
                "total_count": _optional_int(gift.total_count),
                "remaining_count": _optional_int(gift.remaining_count),
            }
            existing = payload_by_id.get(gift_id, {})
            merged = {**existing, **{k: v for k, v in available.items() if v is not None}}
            if not merged.get("image_base64"):
                merged["image_base64"] = existing.get("image_base64")
            payload_by_id[gift_id] = merged
    except Exception as exc:
        logger.warning("get_available_gifts failed: %s", exc)

    payload = list(payload_by_id.values())
    if not payload:
        logger.info("telegram returned no gifts to sync")
        return {"upserted": 0, "disabled": 0}

    result = await catalog.sync_telegram_gifts(payload)
    logger.info(
        "telegram gifts synced: upserted=%s disabled=%s",
        result.get("upserted"),
        result.get("disabled"),
    )
    return result
