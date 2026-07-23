CREATE TABLE alert_rules (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type TEXT NOT NULL, symbol TEXT NOT NULL, condition TEXT NOT NULL, threshold DOUBLE PRECISION NOT NULL,
    channel TEXT NOT NULL DEFAULT 'telegram', is_active BOOLEAN NOT NULL DEFAULT true,
    cooldown_s INTEGER NOT NULL DEFAULT 300, last_fired TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ, created_by UUID, updated_by UUID
);
CREATE INDEX idx_alert_rules_active_symbol_type ON alert_rules (is_active, symbol, type);
CREATE INDEX idx_alert_rules_user_id ON alert_rules (user_id);
CREATE TABLE watchlists (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(), user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol TEXT NOT NULL, created_at TIMESTAMPTZ NOT NULL DEFAULT now(), UNIQUE (user_id, symbol)
);
