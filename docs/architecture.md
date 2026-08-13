# Architecture

## Overview

MarketTG is an event-driven microservices platform for selling Telegram digital goods.

## Service Boundaries

- **Gateway**: No business logic. Auth, routing, rate limiting only.
- **User Service**: `users` schema — identity, favorites, admin users, audit log.
- **Catalog Service**: `catalog` schema — products, FTS search, MinIO images.
- **Order Service**: `orders` schema + Redis cart — checkout with price snapshots.
- **Payment Service**: `payments` schema — never trusts frontend prices.
- **Delivery Service**: `delivery` schema — idempotent handlers per product type.
- **Notification Service**: `notifications` schema — WebSocket + bot messages.

## Event Flow

```
OrderCreated → Redis Stream
PaymentSucceeded → Delivery Service + Notification Service (parallel)
DeliveryCompleted → OrderCompleted → Notification
```

## Outbox Pattern

Payment and Order services write to `outbox_events` in the same DB transaction.
Background publisher pushes to Redis Streams — no lost events.

## Security

- Telegram `initData` HMAC validation on every authenticated request
- Admin JWT + RBAC
- Internal services not exposed publicly
- Idempotency keys on payments and orders

## Data Storage

| Store | Purpose |
|-------|---------|
| PostgreSQL | Source of truth (schema-per-service) |
| Redis | Cart, cache, rate limits, streams, locks |
| MinIO | Product/promo images (S3-compatible) |

## Scaling Path

Services can be split to separate PostgreSQL instances and VPS nodes without code changes.
