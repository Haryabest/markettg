import base64
import logging
from io import BytesIO

from telethon import TelegramClient
from telethon.sessions import MemorySession, StringSession
from telethon.tl.functions.payments import GetResaleStarGiftsRequest, GetStarGiftsRequest

from bot.catalog_client import CatalogClient
from bot.config import settings

logger = logging.getLogger(__name__)

RESALE_PAGE_SIZE = 100


def _stars_value(gift) -> int:
    for attr in ("resell_stars", "stars", "upgrade_stars", "convert_stars"):
        value = getattr(gift, attr, None)
        if value:
            return int(value)
    return 0


async def _preview_base64(client: TelegramClient, gift) -> str | None:
    for attr in getattr(gift, "attributes", None) or []:
        document = getattr(attr, "document", None)
        if not document:
            continue
        try:
            buf = BytesIO()
            await client.download_media(document, file=buf)
            data = buf.getvalue()
            if data:
                return base64.b64encode(data).decode("ascii")
        except Exception as exc:
            logger.debug("nft preview download failed: %s", exc)
    sticker = getattr(gift, "sticker", None)
    if sticker:
        try:
            buf = BytesIO()
            await client.download_media(sticker, file=buf)
            data = buf.getvalue()
            if data:
                return base64.b64encode(data).decode("ascii")
        except Exception as exc:
            logger.debug("nft sticker download failed: %s", exc)
    return None


async def _connect_nft_client() -> TelegramClient | None:
    if not settings.telegram_api_id or not settings.telegram_api_hash:
        logger.info("nft sync skipped: TELEGRAM_API_ID/HASH missing")
        return None

    if settings.telegram_session:
        client = TelegramClient(
            StringSession(settings.telegram_session),
            settings.telegram_api_id,
            settings.telegram_api_hash,
        )
        await client.connect()
        if not await client.is_user_authorized():
            logger.warning("nft sync skipped: TELEGRAM_SESSION is invalid")
            await client.disconnect()
            return None
        return client

    if settings.bot_token:
        client = TelegramClient(
            MemorySession(),
            settings.telegram_api_id,
            settings.telegram_api_hash,
        )
        await client.start(bot_token=settings.bot_token)
        return client

    logger.info("nft sync skipped: TELEGRAM_SESSION or bot token missing")
    return None


async def sync_nft_gifts(catalog: CatalogClient) -> dict | None:
    client = await _connect_nft_client()
    if not client:
        return None

    using_user = bool(settings.telegram_session)
    payload: list[dict] = []
    try:
        catalog_result = await client(GetStarGiftsRequest(hash=0))
        for star_gift in catalog_result.gifts:
            if not getattr(star_gift, "availability_resale", None):
                continue
            if not using_user:
                logger.warning(
                    "nft resale sync needs TELEGRAM_SESSION (bot cannot call GetResaleStarGifts)"
                )
                return None

            offset = ""
            while True:
                resale = await client(
                    GetResaleStarGiftsRequest(
                        gift_id=star_gift.id,
                        offset=offset,
                        limit=RESALE_PAGE_SIZE,
                        stars_only=True,
                    )
                )
                if not resale.gifts:
                    break

                for item in resale.gifts:
                    slug = getattr(item, "slug", None)
                    if not slug:
                        continue
                    stars = _stars_value(item)
                    if stars <= 0:
                        continue
                    payload.append(
                        {
                            "nft_slug": slug,
                            "title": getattr(item, "title", None) or slug,
                            "gift_num": int(getattr(item, "num", 0) or 0),
                            "base_gift_id": int(star_gift.id),
                            "star_count": stars,
                            "image_base64": None,
                        }
                    )

                next_offset = getattr(resale, "next_offset", None) or ""
                if not next_offset or next_offset == offset:
                    break
                offset = next_offset
    except Exception as exc:
        logger.warning("nft sync failed: %s", exc)
        return None
    finally:
        await client.disconnect()

    if not payload:
        logger.info("no nft resale items found")
        return {"upserted": 0, "disabled": 0}

    result = await catalog.sync_nft_gifts(payload)
    logger.info(
        "nft gifts synced: upserted=%s disabled=%s",
        result.get("upserted"),
        result.get("disabled"),
    )
    return result
