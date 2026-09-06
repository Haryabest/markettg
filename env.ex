# Скопируйте в .env в корне репозитория и заполните значения.

# Обязательно для реальных подарков в каталоге
TELEGRAM_BOT_TOKEN=8648803947:AAFnKhb5_dwV0X9gp8O_pNynfq5vacYX-2E

# Для NFT (перепродажа) — получить на https://my.telegram.org
TELEGRAM_API_ID=
TELEGRAM_API_HASH=
# User session для NFT resale (python scripts/telegram-session.py)
TELEGRAM_SESSION=

BOT_INTERNAL_SECRET=bot-secret
STAR_KOPECKS_RATE=180
GIFTS_SYNC_INTERVAL_SEC=600
# Telegram ID администраторов через запятую — им доступна команда /balance
ADMIN_TELEGRAM_IDS=

# Продажа подарков со Stars-баланса бота
# Наценка к цене подарка в Stars, % (0 = продаём по себестоимости)
STARS_MARKUP_PERCENT=0
# Оплачивать ли апгрейд подарка до коллекционного (дороже)
GIFT_PAY_FOR_UPGRADE=false
# Писать ли покупателю в бота после успешной отправки подарка
GIFT_NOTIFY_USER=true

# Mini App
MINI_APP_URL=http://localhost:3000
NEXT_PUBLIC_API_URL=http://localhost:8080
NEXT_PUBLIC_SUPPORT_URL=https://t.me/markettg_support
TELEGRAM_BOT_USERNAME=testmarkettg674598bot
