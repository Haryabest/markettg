# Events

## Redis Streams

| Stream | Events |
|--------|--------|
| `events:payments` | PaymentCreated, PaymentSucceeded, PaymentFailed |
| `events:orders` | OrderCreated, OrderCompleted, OrderCancelled |
| `events:delivery` | DeliveryStarted, DeliveryCompleted, DeliveryFailed |

## Consumer Groups

- `delivery-workers` — Delivery Service
- `notification-workers` — Notification Service

## Event Payloads

### PaymentSucceeded
```json
{
  "payment_id": "uuid",
  "order_id": "uuid",
  "user_id": "uuid",
  "amount_kopecks": 9900,
  "method": "STARS"
}
```

### DeliveryCompleted
```json
{
  "order_id": "uuid",
  "order_item_id": "uuid",
  "user_id": "uuid"
}
```

## Idempotency

All consumers check `processed_events` table before handling.
Delivery uses Redis lock `lock:delivery:{order_id}`.
