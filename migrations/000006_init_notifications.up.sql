CREATE SCHEMA IF NOT EXISTS notifications;

CREATE TYPE notifications.notification_status AS ENUM ('PENDING', 'SENT', 'FAILED');
CREATE TYPE notifications.notification_channel AS ENUM ('TELEGRAM', 'WEBSOCKET', 'EMAIL');

CREATE TABLE notifications.notifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL,
    type VARCHAR(100) NOT NULL,
    channel notifications.notification_channel NOT NULL,
    payload JSONB NOT NULL DEFAULT '{}',
    status notifications.notification_status NOT NULL DEFAULT 'PENDING',
    sent_at TIMESTAMPTZ,
    error_message TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_notifications_user_id ON notifications.notifications(user_id);
CREATE INDEX idx_notifications_status ON notifications.notifications(status);

CREATE TABLE notifications.processed_events (
    event_id VARCHAR(255) PRIMARY KEY,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
