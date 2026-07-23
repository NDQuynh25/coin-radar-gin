CREATE TABLE symbols (
    symbol TEXT PRIMARY KEY, base_asset TEXT NOT NULL, quote_asset TEXT NOT NULL, exchange TEXT NOT NULL,
    market_type TEXT NOT NULL, is_active BOOLEAN NOT NULL DEFAULT true, listed_at TIMESTAMPTZ
);
CREATE TABLE known_wallets (
    address TEXT PRIMARY KEY, chain TEXT NOT NULL, label TEXT, category TEXT,
    is_exchange BOOLEAN NOT NULL DEFAULT false, created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_known_wallets_category ON known_wallets (category);
