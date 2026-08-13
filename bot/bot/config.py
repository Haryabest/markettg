import os
from dataclasses import dataclass


@dataclass
class Settings:
    bot_token: str
    gateway_url: str
    bot_secret: str
    mini_app_url: str


settings = Settings(
    bot_token=os.getenv("TELEGRAM_BOT_TOKEN", ""),
    gateway_url=os.getenv("GATEWAY_URL", "http://gateway:8080"),
    bot_secret=os.getenv("BOT_INTERNAL_SECRET", "bot-secret"),
    mini_app_url=os.getenv("MINI_APP_URL", "http://localhost:3000"),
)
