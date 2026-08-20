import base64
import logging
from io import BytesIO

from aiogram import Bot

from bot.catalog_client import CatalogClient

logger = logging.getLogger(__name__)


def _optional_int(value):
    return int(value) if value is not None else None


async def _preview_base64(bot: Bot, sticker) -> str | None:
    if not sticker:
        return None
    thumb = getattr(sticker, "thumbnail", None) or getattr(sticker, "thumb", None)
    target = thumb or sticker
    try:
        buf = BytesIO()
        await bot.download(target, destination=buf)
        data = buf.getvalue()
        if not data:
            return None
        return base64.b64encode(data).decode("ascii")
    except Exception as exc:
        logger.debug("gift preview download failed: %s", exc)
        return None


async def sync_telegram_gifts(bot: Bot, catalog: CatalogClient) -> dict | None:
    try:
        gifts_result = await bot.get_available_gifts()
    except Exception as exc:
        logger.warning("get_available_gifts failed: %s", exc)
        return None

    payload: list[dict] = []
    for gift in gifts_result.gifts:
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

        payload.append(
            {
                "telegram_gift_id": gift.id,
                "star_count": gift.star_count,
                "emoji": emoji,
                "sticker_thumb_file_id": thumb_file_id,
                "sticker_file_id": sticker_file_id,
                "image_base64": await _preview_base64(bot, sticker),
                "total_count": _optional_int(gift.total_count),
                "remaining_count": _optional_int(gift.remaining_count),
            }
        )

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
