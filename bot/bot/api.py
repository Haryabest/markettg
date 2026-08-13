import httpx


class GatewayClient:
    def __init__(self, base_url: str, bot_secret: str):
        self.base_url = base_url.rstrip("/")
        self.bot_secret = bot_secret
        self.client = httpx.AsyncClient(timeout=30.0)

    def _headers(self, telegram_id: int) -> dict:
        return {
            "X-Bot-Secret": self.bot_secret,
            "X-Telegram-User-Id": str(telegram_id),
        }

    async def get_orders(self, telegram_id: int) -> list:
        resp = await self.client.get(
            f"{self.base_url}/api/v1/internal/bot/orders",
            headers=self._headers(telegram_id),
        )
        resp.raise_for_status()
        data = resp.json()
        return data.get("orders", [])

    async def close(self):
        await self.client.aclose()
