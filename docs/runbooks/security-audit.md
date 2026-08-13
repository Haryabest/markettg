# Security Audit Checklist

## Authentication
- [x] Telegram initData HMAC-SHA256 validation
- [x] Admin JWT with role-based access
- [x] Bot internal secret for service-to-service
- [ ] TOTP 2FA for admin (interface ready, enable per user)

## Payments
- [x] Never trust frontend prices
- [x] Idempotency keys on payment creation
- [x] Webhook signature verification (SBP)
- [x] Payment event deduplication

## Delivery
- [x] Idempotent delivery jobs (unique order_item_id)
- [x] Redis distributed locks
- [x] Retry with exponential backoff
- [x] processed_events table

## Infrastructure
- [x] Secrets in .env only
- [x] Internal services on Docker network
- [x] Rate limiting on gateway
- [x] Structured JSON logging with request_id

## Recommendations for Production
1. Rotate JWT secrets regularly
2. Enable Cloudflare WAF
3. Use TLS everywhere (Caddy auto HTTPS)
4. Restrict admin panel by IP if possible
5. Set `SBP_MOCK_MODE=false` with real provider
