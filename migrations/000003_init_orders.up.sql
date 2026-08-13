CREATE SCHEMA IF NOT EXISTS orders;

CREATE TYPE orders.order_status AS ENUM (
    'CREATED',
    'PAYMENT_PENDING',
    'PAID',
    'DELIVERY_PENDING',
    'DELIVERING',
    'COMPLETED',
    'DELIVERY_FAILED',
    'CANCELLED',
    'REFUNDED'
);

CREATE TYPE orders.discount_type AS ENUM ('PERCENT', 'FIXED');

CREATE TABLE orders.promo_codes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    code VARCHAR(50) NOT NULL UNIQUE,
    discount_type orders.discount_type NOT NULL,
    discount_value BIGINT NOT NULL CHECK (discount_value > 0),
    max_uses INT,
    used_count INT NOT NULL DEFAULT 0,
    valid_from TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    valid_to TIMESTAMPTZ,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE orders.orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    status orders.order_status NOT NULL DEFAULT 'CREATED',
    total_kopecks BIGINT NOT NULL CHECK (total_kopecks >= 0),
    discount_kopecks BIGINT NOT NULL DEFAULT 0,
    promo_code_id UUID REFERENCES orders.promo_codes(id),
    idempotency_key VARCHAR(255) UNIQUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_orders_user_id ON orders.orders(user_id);
CREATE INDEX idx_orders_status ON orders.orders(status);
CREATE INDEX idx_orders_created ON orders.orders(created_at DESC);

CREATE TABLE orders.order_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL REFERENCES orders.orders(id) ON DELETE CASCADE,
    product_id UUID NOT NULL,
    name VARCHAR(255) NOT NULL,
    price_kopecks BIGINT NOT NULL,
    delivery_config JSONB NOT NULL DEFAULT '{}',
    product_type VARCHAR(20) NOT NULL,
    quantity INT NOT NULL DEFAULT 1 CHECK (quantity > 0),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_order_items_order_id ON orders.order_items(order_id);

CREATE TABLE orders.promo_code_usages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    promo_code_id UUID NOT NULL REFERENCES orders.promo_codes(id),
    user_id UUID NOT NULL,
    order_id UUID NOT NULL REFERENCES orders.orders(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(promo_code_id, order_id)
);

CREATE TABLE orders.outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_orders_outbox_unpublished ON orders.outbox_events(published_at) WHERE published_at IS NULL;
