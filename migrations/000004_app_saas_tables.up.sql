CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), telegram_id BIGINT UNIQUE, email TEXT UNIQUE,
    username TEXT, password_hash TEXT, created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(), deleted_at TIMESTAMPTZ,
    created_by UUID, updated_by UUID
);
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan TEXT NOT NULL DEFAULT 'free', status TEXT NOT NULL DEFAULT 'active',
    started_at TIMESTAMPTZ NOT NULL DEFAULT now(), expires_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ, created_by UUID, updated_by UUID
);
CREATE INDEX idx_subscriptions_user_id ON subscriptions (user_id);
CREATE INDEX idx_subscriptions_status_expires_at ON subscriptions (status, expires_at);
CREATE TABLE payments (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount_usd DOUBLE PRECISION NOT NULL, currency TEXT NOT NULL DEFAULT 'USDT', network TEXT NOT NULL,
    tx_hash TEXT UNIQUE, status TEXT NOT NULL DEFAULT 'pending', created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirmed_at TIMESTAMPTZ
);
CREATE INDEX idx_payments_user_created_at ON payments (user_id, created_at DESC);
