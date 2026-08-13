# MarketTG Monorepo

Telegram Marketplace — Mini App + Bot + Admin Panel + Microservices backend.

## Architecture

```
Mini App / Admin / Bot → Caddy → API Gateway → Go Microservices
                                              ↓
                                    PostgreSQL + Redis + MinIO
                                              ↓
                                    Redis Streams (events)
```

### Services

| Service | Port | Responsibility |
|---------|------|----------------|
| gateway | 8080 | Routing, auth, rate limiting |
| user-service | 8081 | Users, favorites, admin auth |
| catalog-service | 8082 | Products, categories, search, S3 uploads |
| order-service | 8083 | Cart (Redis), orders, promo codes |
| payment-service | 8084 | Telegram Stars, SBP |
| delivery-service | 8085 | Stars (Bot API), Premium/Gifts (Fragment) |
| notification-service | 8086 | WebSocket, Telegram notifications |
| bot | 8090 | aiogram 3, Gateway client only |
| mini-app | 3000 | Next.js Telegram Mini App |
| admin | 3001 | Next.js admin panel |

## Quick Start

```bash
cp .env.example .env
# Set TELEGRAM_BOT_TOKEN in .env

docker compose up --build
```

- Mini App: http://localhost:3000
- Admin: http://localhost:3001 (admin@markettg.local / password)
- API: http://localhost:8080
- Grafana: http://localhost:3002 (admin / admin)
- MinIO Console: http://localhost:9001 (minioadmin / minioadmin)

## Environment Variables

See [.env.example](.env.example) for full list.

Critical:
- `TELEGRAM_BOT_TOKEN` — Bot API + initData validation
- `FRAGMENT_API_TOKEN` — Premium & Gifts delivery
- `ADMIN_JWT_SECRET` / `JWT_SECRET` — Auth tokens
- `S3_*` — MinIO/S3 for product images

## Migrations

Migrations run automatically via `migrate` service on `docker compose up`.

Manual:
```bash
docker compose run --rm migrate -path /migrations -database "postgres://markettg:markettg@postgres:5432/markettg?sslmode=disable" up
```

## Create Admin

Default admin is seeded: `admin@markettg.local` / `password`

Change password in production via direct DB update or admin UI.

## Telegram Bot Setup

1. Create bot via [@BotFather](https://t.me/BotFather)
2. Set `TELEGRAM_BOT_TOKEN` in `.env`
3. Configure Mini App URL: `/setmenubutton` → `http://localhost:3000`
4. Set webhook for payments to `https://your-domain/api/v1/payments/webhooks/telegram`

## Telegram Stars

Payment flow uses Bot API `createInvoiceLink`. Configure bot for payments in BotFather.

## SBP

`SBP_MOCK_MODE=true` for local dev. Set `SBP_PROVIDER_*` when provider is chosen.

## Production Deployment

1. VPS: 4 vCPU, 8 GB RAM recommended
2. Cloudflare → Caddy → Docker Compose
3. Set all secrets in `.env`
4. Use `docker compose -f docker-compose.yml -f docker-compose.prod.yml up -d`
5. Enable PostgreSQL backups: `pg_dump` cron daily

## Monitoring

- Prometheus: http://localhost:9090
- Grafana dashboards in `infrastructure/grafana/dashboards/`
- Loki logs via Promtail

## Troubleshooting

| Issue | Solution |
|-------|----------|
| initData invalid | Check TELEGRAM_BOT_TOKEN matches bot |
| Cart empty after restart | Redis must be running |
| Images not loading | Check MinIO bucket `catalog` exists |
| Payment webhook fails | Verify webhook URL is HTTPS in production |

See [docs/runbooks/](docs/runbooks/) for detailed guides.
