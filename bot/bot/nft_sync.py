import base64
import logging
from io import BytesIO

from telethon import TelegramClient
from telethon.sessions import MemorySession
from telethon.tl.functions.payments import GetResaleStarGiftsRequest, GetStarGiftsRequest

from bot.catalog_client import CatalogClient
from bot.config import settings

logger = logging.getLogger(__name__)

MAX_NFT_ITEMS = 80
MAX_COLLECTIONS = 12


def _stars_value(gift) -> int:
    for attr in ("resell_stars", "stars", "upgrade_stars"):
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


async def sync_nft_gifts(catalog: CatalogClient) -> dict | None:
    if not settings.bot_token or not settings.telegram_api_id or not settings.telegram_api_hash:
        logger.info("nft sync skipped: TELEGRAM_API_ID/HASH or bot token missing")
        return None

    client = TelegramClient(MemorySession(), settings.telegram_api_id, settings.telegram_api_hash)
    await client.start(bot_token=settings.bot_token)

    payload: list[dict] = []
    try:
        catalog_result = await client(GetStarGiftsRequest(hash=0))
        collections = 0
        for star_gift in catalog_result.gifts:
            if collections >= MAX_COLLECTIONS or len(payload) >= MAX_NFT_ITEMS:
                break
            if not getattr(star_gift, "availability_resale", False):
                continue
            collections += 1

            resale = await client(
                GetResaleStarGiftsRequest(
                    gift_id=star_gift.id,
                    offset="",
                    limit=20,
                    stars_only=True,
                )
            )

            for item in resale.gifts:
                if len(payload) >= MAX_NFT_ITEMS:
                    break
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
                        "image_base64": await _preview_base64(client, item),
                    }
                )
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
