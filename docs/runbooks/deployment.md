# Deployment Runbook

## Prerequisites

- Docker & Docker Compose
- Domain with Cloudflare (production)
- Telegram Bot Token
- VPS: 4 vCPU, 8 GB RAM, NVMe

## Local Development

```bash
cp .env.example .env
docker compose up --build
```

## Production

1. Clone repo on VPS
2. Configure `.env` with production secrets
3. Point domain to VPS IP via Cloudflare
4. Update Caddyfile with your domain
5. Run: `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d --build`

## PostgreSQL Backup

```bash
docker compose exec postgres pg_dump -U markettg markettg > backup_$(date +%Y%m%d).sql
```

Cron daily at 3 AM recommended.

## Rollback

```bash
docker compose pull
docker compose up -d --no-build
```

## Health Checks

- Gateway: `curl http://localhost:8080/health`
- All services expose `/health`
