CREATE TABLE whale_transfers (
    time TIMESTAMPTZ NOT NULL, chain TEXT NOT NULL, tx_hash TEXT NOT NULL, token_address TEXT,
    token_symbol TEXT, from_address TEXT NOT NULL, to_address TEXT NOT NULL, amount NUMERIC,
    value_usd DOUBLE PRECISION, direction TEXT
);
SELECT create_hypertable('whale_transfers', by_range('time', INTERVAL '1 day'));
CREATE INDEX idx_whale_transfers_token_time ON whale_transfers (token_symbol, time DESC);
CREATE INDEX idx_whale_transfers_value_time ON whale_transfers (value_usd DESC, time DESC);

CREATE TABLE signals (
    time TIMESTAMPTZ NOT NULL, id BIGINT GENERATED ALWAYS AS IDENTITY, type TEXT NOT NULL,
    exchange TEXT, symbol TEXT NOT NULL, severity TEXT NOT NULL, value DOUBLE PRECISION,
    threshold DOUBLE PRECISION, payload JSONB, PRIMARY KEY (time, id)
);
SELECT create_hypertable('signals', by_range('time', INTERVAL '7 days'));
CREATE INDEX idx_signals_type_time ON signals (type, time DESC);
CREATE INDEX idx_signals_symbol_time ON signals (symbol, time DESC);
