from urllib.parse import quote

from aiogram.types import InlineKeyboardButton, InlineKeyboardMarkup, WebAppInfo
from aiogram.utils.keyboard import InlineKeyboardBuilder

from bot.config import settings


def _app(path: str = "", start_param: str = "") -> str:
    base = settings.mini_app_url.rstrip("/")
    if path:
        base = f"{base}/{path.lstrip('/')}"
    if start_param:
        sep = "&" if "?" in base else "?"
        base = f"{base}{sep}tgWebAppStartParam={quote(start_param, safe='')}"
    return base


def main_menu_keyboard(start_param: str = "") -> InlineKeyboardMarkup:
    builder = InlineKeyboardBuilder()
    builder.row(
        InlineKeyboardButton(
            text="🛍 Открыть магазин",
            web_app=WebAppInfo(url=_app(start_param=start_param)),
        )
    )
    builder.row(
        InlineKeyboardButton(
            text="⭐ Каталог",
            web_app=WebAppInfo(url=_app("catalog")),
        ),
        InlineKeyboardButton(
            text="📦 Заказы",
            web_app=WebAppInfo(url=_app("orders")),
        ),
    )
    builder.row(
        InlineKeyboardButton(
            text="🛒 Корзина",
            web_app=WebAppInfo(url=_app("cart")),
        ),
        InlineKeyboardButton(
            text="❤️ Избранное",
            web_app=WebAppInfo(url=_app("favorites")),
        ),
    )
    builder.row(
        InlineKeyboardButton(
            text="👥 Пригласи друга",
            web_app=WebAppInfo(url=_app("profile/referral")),
        )
    )
    builder.row(
        InlineKeyboardButton(
            text="💬 Поддержка",
            url=settings.support_url,
        )
    )
    builder.row(
        InlineKeyboardButton(
            text="📄 Соглашение",
            web_app=WebAppInfo(url=_app("legal/terms")),
        ),
        InlineKeyboardButton(
            text="🔒 Конфиденциальность",
            web_app=WebAppInfo(url=_app("legal/privacy")),
        ),
    )
    return builder.as_markup()


def shop_keyboard() -> InlineKeyboardMarkup:
    builder = InlineKeyboardBuilder()
    builder.button(
        text="🛍 Открыть магазин",
        web_app=WebAppInfo(url=_app()),
    )
    return builder.as_markup()


def catalog_keyboard() -> InlineKeyboardMarkup:
    builder = InlineKeyboardBuilder()
    builder.button(
        text="⭐ Открыть каталог",
        web_app=WebAppInfo(url=_app("catalog")),
    )
    return builder.as_markup()


def support_keyboard() -> InlineKeyboardMarkup:
    builder = InlineKeyboardBuilder()
    builder.button(text="💬 Написать в поддержку", url=settings.support_url)
    builder.row(
        InlineKeyboardButton(
            text="🛍 Магазин",
            web_app=WebAppInfo(url=_app()),
        )
    )
    return builder.as_markup()


def profile_keyboard() -> InlineKeyboardMarkup:
    builder = InlineKeyboardBuilder()
    builder.row(
        InlineKeyboardButton(
            text="📦 Мои заказы",
            web_app=WebAppInfo(url=_app("orders")),
        ),
        InlineKeyboardButton(
            text="🛒 Корзина",
            web_app=WebAppInfo(url=_app("cart")),
        ),
    )
    builder.row(
        InlineKeyboardButton(
            text="❤️ Избранное",
            web_app=WebAppInfo(url=_app("favorites")),
        ),
        InlineKeyboardButton(
            text="👤 Профиль",
            web_app=WebAppInfo(url=_app("profile")),
        ),
    )
    return builder.as_markup()
