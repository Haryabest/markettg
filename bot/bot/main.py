"""Telegram Bot — thin client over Gateway API. No business logic."""

import asyncio
import logging
import os

from aiogram import Bot, Dispatcher
from aiogram.filters import Command
from aiogram.types import MenuButtonWebApp, Message, WebAppInfo
from aiogram.utils.keyboard import InlineKeyboardBuilder
from aiohttp import web

from bot.api import GatewayClient
from bot.catalog_client import CatalogClient
from bot.config import settings
from bot.gifts_sync import sync_telegram_gifts
from bot.nft_sync import sync_nft_gifts

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

bot = Bot(token=settings.bot_token)
dp = Dispatcher()
gateway = GatewayClient(settings.gateway_url, settings.bot_secret)
catalog = CatalogClient(settings.catalog_url, settings.bot_secret)


@dp.message(Command("start"))
async def cmd_start(message: Message):
    parts = (message.text or "").split(maxsplit=1)
    start_param = parts[1].strip() if len(parts) > 1 else ""

    builder = InlineKeyboardBuilder()
    builder.button(text="🛍 Открыть магазин", web_app=WebAppInfo(url=settings.mini_app_url))

    text = (
        "Добро пожаловать в MarketTG!\n\n"
        "Покупайте Telegram Stars, Premium и подарки."
    )
    if start_param.startswith("ref_"):
        text += (
            "\n\n🎁 Вы перешли по реферальной ссылке — после входа в магазин "
            "друг получит 5% на первый заказ, а пригласивший — 100 ₽ бонусом."
        )

    await message.answer(text, reply_markup=builder.as_markup())


@dp.message(Command("orders"))
async def cmd_orders(message: Message):
    try:
        orders = await gateway.get_orders(message.from_user.id)
        if not orders:
            await message.answer("У вас пока нет заказов.")
            return
        lines = [f"#{o['id'][:8]} — {o['status']} — {o['total_kopecks'] / 100:.0f} ₽" for o in orders[:10]]
        await message.answer("Ваши заказы:\n\n" + "\n".join(lines))
    except Exception as e:
        logger.error("orders error: %s", e)
        await message.answer("Не удалось загрузить заказы. Попробуйте позже.")


@dp.message(Command("help"))
async def cmd_help(message: Message):
    await message.answer(
        "/start — открыть магазин\n"
        "/orders — история заказов\n"
        "/help — справка"
    )


async def notify_handler(request: web.Request):
    secret = request.headers.get("X-Bot-Secret", "")
    if secret != settings.bot_secret:
        return web.json_response({"error": "unauthorized"}, status=401)
    data = await request.json()
    telegram_id = data.get("telegram_id")
    text = data.get("message", "")
    if telegram_id and text:
        await bot.send_message(telegram_id, text)
    return web.json_response({"ok": True})


async def health_handler(request: web.Request):
    return web.json_response({"status": "ok", "service": "bot"})


async def start_http_server():
    app = web.Application()
    app.router.add_post("/internal/notify", notify_handler)
    app.router.add_get("/health", health_handler)
    runner = web.AppRunner(app)
    await runner.setup()
    site = web.TCPSite(runner, "0.0.0.0", 8090)
    await site.start()
    logger.info("HTTP server started on :8090")


async def gifts_sync_worker():
    if not settings.bot_token:
        logger.warning("TELEGRAM_BOT_TOKEN missing — gifts sync disabled")
        return
    while True:
        try:
            await sync_telegram_gifts(bot, catalog)
            await sync_nft_gifts(catalog)
        except Exception as exc:
            logger.error("gifts sync error: %s", exc)
        await asyncio.sleep(settings.gifts_sync_interval_sec)


async def main():
    await start_http_server()
    if settings.mini_app_url:
        await bot.set_chat_menu_button(
            menu_button=MenuButtonWebApp(text="Магазин", web_app=WebAppInfo(url=settings.mini_app_url))
        )
        logger.info("Mini App menu button set to %s", settings.mini_app_url)
    asyncio.create_task(gifts_sync_worker())
    await dp.start_polling(bot)


if __name__ == "__main__":
    asyncio.run(main())
