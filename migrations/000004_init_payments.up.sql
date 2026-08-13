CREATE SCHEMA IF NOT EXISTS payments;

CREATE TYPE payments.payment_method AS ENUM ('STARS', 'SBP');
CREATE TYPE payments.payment_status AS ENUM ('PENDING', 'PROCESSING', 'PAID', 'FAILED', 'REFUNDED', 'CANCELLED');

CREATE TABLE payments.payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    user_id UUID NOT NULL,
    method payments.payment_method NOT NULL,
    amount_kopecks BIGINT NOT NULL CHECK (amount_kopecks > 0),
    status payments.payment_status NOT NULL DEFAULT 'PENDING',
    provider_payment_id VARCHAR(255),
    idempotency_key VARCHAR(255) NOT NULL UNIQUE,
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_payments_order_id ON payments.payments(order_id);
CREATE INDEX idx_payments_user_id ON payments.payments(user_id);
CREATE INDEX idx_payments_status ON payments.payments(status);
CREATE INDEX idx_payments_provider_id ON payments.payments(provider_payment_id);

CREATE TABLE payments.payment_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    payment_id UUID NOT NULL REFERENCES payments.payments(id),
    event_type VARCHAR(100) NOT NULL,
    provider_event_id VARCHAR(255),
    raw_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(provider_event_id)
);

CREATE TABLE payments.outbox_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    aggregate_type VARCHAR(100) NOT NULL,
    aggregate_id VARCHAR(255) NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    published_at TIMESTAMPTZ
);

CREATE INDEX idx_payments_outbox_unpublished ON payments.outbox_events(published_at) WHERE published_at IS NULL;
