CREATE SCHEMA IF NOT EXISTS delivery;

CREATE TYPE delivery.job_status AS ENUM ('PENDING', 'PROCESSING', 'COMPLETED', 'FAILED', 'CANCELLED');

CREATE TABLE delivery.delivery_jobs (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    order_id UUID NOT NULL,
    order_item_id UUID NOT NULL,
    user_id UUID NOT NULL,
    telegram_id BIGINT NOT NULL,
    handler_type VARCHAR(50) NOT NULL,
    delivery_config JSONB NOT NULL DEFAULT '{}',
    status delivery.job_status NOT NULL DEFAULT 'PENDING',
    attempts INT NOT NULL DEFAULT 0,
    max_attempts INT NOT NULL DEFAULT 5,
    last_error TEXT,
    completed_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    UNIQUE(order_item_id)
);

CREATE INDEX idx_delivery_jobs_order_id ON delivery.delivery_jobs(order_id);
CREATE INDEX idx_delivery_jobs_status ON delivery.delivery_jobs(status);

CREATE TABLE delivery.delivery_attempts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    job_id UUID NOT NULL REFERENCES delivery.delivery_jobs(id) ON DELETE CASCADE,
    attempt_no INT NOT NULL,
    request_payload JSONB,
    response_payload JSONB,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE delivery.processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
