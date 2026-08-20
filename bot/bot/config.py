import os
from dataclasses import dataclass


@dataclass
class Settings:
    bot_token: str
    gateway_url: str
    bot_secret: str
    mini_app_url: str
    catalog_url: str
    gifts_sync_interval_sec: int
    telegram_api_id: int
    telegram_api_hash: str


settings = Settings(
    bot_token=os.getenv("TELEGRAM_BOT_TOKEN", ""),
    gateway_url=os.getenv("GATEWAY_URL", "http://gateway:8080"),
    bot_secret=os.getenv("BOT_INTERNAL_SECRET", "bot-secret"),
    mini_app_url=os.getenv("MINI_APP_URL", "http://localhost:3000"),
    catalog_url=os.getenv("CATALOG_SERVICE_URL", "http://catalog-service:8082"),
    gifts_sync_interval_sec=int(os.getenv("GIFTS_SYNC_INTERVAL_SEC", "600")),
    telegram_api_id=int(os.getenv("TELEGRAM_API_ID", "0") or "0"),
    telegram_api_hash=os.getenv("TELEGRAM_API_HASH", ""),
)
