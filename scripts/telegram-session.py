#!/usr/bin/env python3
"""Generate TELEGRAM_SESSION for NFT resale sync (one-time setup)."""

import asyncio
import os
from pathlib import Path

from dotenv import load_dotenv
from telethon import TelegramClient
from telethon.sessions import StringSession

load_dotenv(Path(__file__).resolve().parents[1] / ".env")

API_ID = int(os.getenv("TELEGRAM_API_ID", "0") or "0")
API_HASH = os.getenv("TELEGRAM_API_HASH", "")


async def main() -> None:
    if not API_ID or not API_HASH:
        raise SystemExit("Set TELEGRAM_API_ID and TELEGRAM_API_HASH in .env first.")

    client = TelegramClient(StringSession(), API_ID, API_HASH)
    await client.start()
    session = client.session.save()
    await client.disconnect()

    print("\nAdd this line to .env:\n")
    print(f"TELEGRAM_SESSION={session}")
    print("\nThen restart bot: docker compose up -d --build bot")


if __name__ == "__main__":
    asyncio.run(main())
