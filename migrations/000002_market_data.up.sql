CREATE TABLE trades (
    id UUID NOT NULL DEFAULT gen_random_uuid(), time TIMESTAMPTZ NOT NULL,
    exchange TEXT NOT NULL, trade_id TEXT NOT NULL, symbol TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL, qty DOUBLE PRECISION NOT NULL, side TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(), updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    deleted_at TIMESTAMPTZ, created_by UUID, updated_by UUID,
    PRIMARY KEY (time, id), UNIQUE (time, exchange, trade_id)
);
SELECT create_hypertable('trades', by_range('time', INTERVAL '1 day'));
CREATE INDEX idx_trades_symbol_time ON trades (symbol, time DESC);
CREATE INDEX idx_trades_exchange_time ON trades (exchange, time DESC);

CREATE TABLE klines (
    time TIMESTAMPTZ NOT NULL, exchange TEXT NOT NULL, symbol TEXT NOT NULL, interval TEXT NOT NULL,
    open DOUBLE PRECISION NOT NULL, high DOUBLE PRECISION NOT NULL, low DOUBLE PRECISION NOT NULL,
    close DOUBLE PRECISION NOT NULL, volume DOUBLE PRECISION NOT NULL, trade_count INTEGER,
    PRIMARY KEY (time, exchange, symbol, interval)
);
SELECT create_hypertable('klines', by_range('time', INTERVAL '7 days'));
CREATE INDEX idx_klines_symbol_interval_time ON klines (symbol, interval, time DESC);

CREATE TABLE derivatives (
    time TIMESTAMPTZ NOT NULL, exchange TEXT NOT NULL, symbol TEXT NOT NULL,
    funding_rate DOUBLE PRECISION, next_funding_time TIMESTAMPTZ, open_interest DOUBLE PRECISION,
    mark_price DOUBLE PRECISION, index_price DOUBLE PRECISION
);
SELECT create_hypertable('derivatives', by_range('time', INTERVAL '1 day'));
CREATE INDEX idx_derivatives_symbol_time ON derivatives (symbol, time DESC);

CREATE TABLE liquidations (
    time TIMESTAMPTZ NOT NULL, exchange TEXT NOT NULL, symbol TEXT NOT NULL, side TEXT NOT NULL,
    price DOUBLE PRECISION NOT NULL, qty DOUBLE PRECISION NOT NULL, value_usd DOUBLE PRECISION
);
SELECT create_hypertable('liquidations', by_range('time', INTERVAL '1 day'));
CREATE INDEX idx_liquidations_symbol_time ON liquidations (symbol, time DESC);

CREATE TABLE orderbook_snapshots (
    time TIMESTAMPTZ NOT NULL, exchange TEXT NOT NULL, symbol TEXT NOT NULL,
    bids JSONB NOT NULL, asks JSONB NOT NULL
);
SELECT create_hypertable('orderbook_snapshots', by_range('time', INTERVAL '6 hours'));
