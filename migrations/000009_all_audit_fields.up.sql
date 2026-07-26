SELECT remove_compression_policy('klines', if_exists => TRUE);
ALTER TABLE klines SET (timescaledb.compress = FALSE);

ALTER TABLE klines
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE derivatives
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE liquidations
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE orderbook_snapshots
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE whale_transfers
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE signals
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE payments
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE watchlists
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE symbols
    ADD COLUMN created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;
ALTER TABLE known_wallets
    ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    ADD COLUMN deleted_at TIMESTAMPTZ,
    ADD COLUMN created_by UUID,
    ADD COLUMN updated_by UUID;

ALTER TABLE klines SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, interval'
);
SELECT add_compression_policy('klines', INTERVAL '30 days');
