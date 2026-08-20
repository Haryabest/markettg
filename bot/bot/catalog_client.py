import httpx


class CatalogClient:
    def __init__(self, base_url: str, internal_secret: str):
        self.base_url = base_url.rstrip("/")
        self.internal_secret = internal_secret
        self.client = httpx.AsyncClient(timeout=120.0)

    async def sync_telegram_gifts(self, gifts: list[dict]) -> dict:
        resp = await self.client.post(
            f"{self.base_url}/api/v1/internal/gifts/sync",
            headers={"X-Internal-Secret": self.internal_secret},
            json={"gifts": gifts},
        )
        resp.raise_for_status()
        return resp.json()

    async def sync_nft_gifts(self, items: list[dict]) -> dict:
        resp = await self.client.post(
            f"{self.base_url}/api/v1/internal/nft/sync",
            headers={"X-Internal-Secret": self.internal_secret},
            json={"items": items},
        )
        resp.raise_for_status()
        return resp.json()

    async def close(self):
        await self.client.aclose()
