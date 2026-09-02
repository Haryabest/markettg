"""Telegram Bot — thin client over Gateway API. No business logic."""

import asyncio
import logging

from aiogram import Bot, Dispatcher, F
from aiogram.client.default import DefaultBotProperties
from aiogram.enums import ParseMode
from aiogram.filters import Command, CommandStart
from aiogram.types import BotCommand, MenuButtonWebApp, Message, WebAppInfo
from aiohttp import web

from bot.api import GatewayClient
from bot.catalog_client import CatalogClient
from bot.config import settings
from bot.gifts_sync import sync_telegram_gifts
from bot.keyboards import catalog_keyboard, main_menu_keyboard, profile_keyboard, shop_keyboard, support_keyboard

try:
    from bot.nft_sync import sync_nft_gifts
except ImportError:
    sync_nft_gifts = None
    logging.getLogger(__name__).warning("NFT sync unavailable — update telethon to a recent version")

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger(__name__)

bot = Bot(token=settings.bot_token, default=DefaultBotProperties(parse_mode=ParseMode.HTML))
dp = Dispatcher()
gateway = GatewayClient(settings.gateway_url, settings.bot_secret)
catalog = CatalogClient(settings.catalog_url, settings.bot_secret)

WELCOME_TEXT = (
    "👋 <b>Добро пожаловать в MarketTG!</b>\n\n"
    "Здесь можно купить <b>Telegram Stars</b>, <b>Premium</b> и <b>подарки</b> "
    "с доставкой на ваш аккаунт сразу после оплаты.\n\n"
    "⚡️ Оплата Stars или СБП\n"
    "🔐 Авторизация через Telegram — без паролей\n"
    "🎁 Промокод <code>WELCOME10</code> — 10% на первый заказ\n\n"
    "Нажмите «Открыть магазин» или выберите действие ниже."
)

REFERRAL_SUFFIX = (
    "\n\n🎁 Вы перешли по реферальной ссылке — после входа в магазин "
    "друг получит <b>5%</b> на первый заказ, а пригласивший — <b>100 ₽</b> бонусом."
)

HELP_TEXT = (
    "<b>Команды бота</b>\n\n"
    "/start — главное меню и магазин\n"
    "/catalog — каталог Stars, Premium и подарков\n"
    "/orders — история ваших заказов\n"
    "/profile — корзина, избранное и профиль\n"
    "/support — связаться с поддержкой\n"
    "/legal — соглашение и политика конфиденциальности\n"
    "/help — эта справка\n\n"
    "<b>Как купить</b>\n"
    "1. Откройте магазин\n"
    "2. Выберите товар и оформите заказ\n"
    "3. Оплатите Stars или СБП\n"
    "4. Доставка придёт на ваш Telegram-аккаунт автоматически\n\n"
    "Промокод <code>WELCOME10</code> — 10% на первый заказ.\n"
    "По вопросам оплаты и доставки — /support"
)

LEGAL_TEXT = (
    "<b>Правовая информация</b>\n\n"
    "Перед покупкой ознакомьтесь с документами:\n"
    "• Пользовательское соглашение — условия покупки и реферальной программы\n"
    "• Политика конфиденциальности — какие данные мы обрабатываем\n\n"
    "Откройте нужный документ кнопками в главном меню (/start)."
)


@dp.message(CommandStart())
async def cmd_start(message: Message):
    text = WELCOME_TEXT
    start_param = (message.text or "").split(maxsplit=1)
    if len(start_param) > 1 and start_param[1].strip().startswith("ref_"):
        text += REFERRAL_SUFFIX
    await message.answer(text, reply_markup=main_menu_keyboard())


@dp.message(Command("catalog"))
async def cmd_catalog(message: Message):
    await message.answer(
        "⭐ <b>Каталог MarketTG</b>\n\n"
        "Stars, Premium и подарки — с мгновенной доставкой после оплаты.",
        reply_markup=catalog_keyboard(),
    )


@dp.message(Command("orders"))
async def cmd_orders(message: Message):
    try:
        orders = await gateway.get_orders(message.from_user.id)
        if not orders:
            await message.answer(
                "📦 У вас пока нет заказов.\n\nОткройте каталог и оформите первую покупку.",
                reply_markup=shop_keyboard(),
            )
            return
        lines = [
            f"#{o['id'][:8]} — {o['status']} — {o['total_kopecks'] / 100:.0f} ₽"
            for o in orders[:10]
        ]
        await message.answer(
            "<b>Ваши заказы</b>\n\n" + "\n".join(lines) + "\n\nПодробности — в Mini App.",
            reply_markup=main_menu_keyboard(),
        )
    except Exception as exc:
        logger.error("orders error: %s", exc)
        await message.answer(
            "Не удалось загрузить заказы. Попробуйте позже или откройте магазин.",
            reply_markup=shop_keyboard(),
        )


@dp.message(Command("support"))
async def cmd_support(message: Message):
    await message.answer(
        "💬 <b>Поддержка MarketTG</b>\n\n"
        "Поможем с заказом, оплатой, доставкой Stars / Premium / подарков "
        "и удалением персональных данных.\n\n"
        "Напишите нам — ответим в Telegram.",
        reply_markup=support_keyboard(),
    )


@dp.message(Command("profile"))
async def cmd_profile(message: Message):
    await message.answer(
        "👤 <b>Профиль MarketTG</b>\n\n"
        "Заказы, корзина и избранное привязаны к вашему Telegram.\n"
        "Пригласите друга — получите <b>100 ₽</b> бонусом после его первой покупки.",
        reply_markup=profile_keyboard(),
    )


@dp.message(Command("legal"))
async def cmd_legal(message: Message):
    await message.answer(LEGAL_TEXT, reply_markup=main_menu_keyboard())


@dp.message(Command("help"))
async def cmd_help(message: Message):
    await message.answer(HELP_TEXT, reply_markup=main_menu_keyboard())


@dp.message(F.text)
async def fallback_message(message: Message):
    await message.answer(
        "Я помогу с покупкой Stars, Premium и подарков.\n"
        "Нажмите /start для главного меню или /help для справки.",
        reply_markup=shop_keyboard(),
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


async def setup_bot_profile() -> bool:
    if not settings.mini_app_url.startswith("https://"):
        logger.warning("MINI_APP_URL is not HTTPS — menu button and WebApp may not work in Telegram")
        return False

    await bot.set_my_commands(
        [
            BotCommand(command="start", description="Главное меню и магазин"),
            BotCommand(command="catalog", description="Каталог товаров"),
            BotCommand(command="orders", description="Мои заказы"),
            BotCommand(command="profile", description="Профиль, корзина, избранное"),
            BotCommand(command="support", description="Связаться с поддержкой"),
            BotCommand(command="legal", description="Соглашение и конфиденциальность"),
            BotCommand(command="help", description="Справка"),
        ]
    )
    await bot.set_chat_menu_button(
        menu_button=MenuButtonWebApp(
            text="Магазин",
            web_app=WebAppInfo(url=settings.mini_app_url),
        )
    )
    logger.info("Bot profile updated, Mini App: %s", settings.mini_app_url)
    return True


async def setup_bot_profile_with_retry():
    for attempt in range(1, 6):
        try:
            await setup_bot_profile()
            return
        except Exception as exc:
            logger.warning("Bot profile setup attempt %s failed: %s", attempt, exc)
            await asyncio.sleep(min(attempt * 3, 15))
    logger.error("Could not configure bot menu/commands — bot will still poll for messages")


async def gifts_sync_worker():
    if not settings.bot_token:
        logger.warning("TELEGRAM_BOT_TOKEN missing — gifts sync disabled")
        return
    while True:
        try:
            await sync_telegram_gifts(bot, catalog)
            if sync_nft_gifts and settings.telegram_api_id and settings.telegram_api_hash:
                await sync_nft_gifts(catalog)
        except Exception as exc:
            logger.error("gifts sync error: %s", exc)
        await asyncio.sleep(settings.gifts_sync_interval_sec)


async def main():
    await start_http_server()
    await setup_bot_profile_with_retry()
    asyncio.create_task(gifts_sync_worker())

    while True:
        try:
            logger.info("Bot polling started")
            await dp.start_polling(bot)
        except Exception as exc:
            logger.error("Polling stopped: %s — retry in 10s", exc)
            await asyncio.sleep(10)


if __name__ == "__main__":
    asyncio.run(main())
