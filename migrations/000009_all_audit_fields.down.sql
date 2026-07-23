ALTER TABLE known_wallets DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at;
ALTER TABLE symbols DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE watchlists DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at;
ALTER TABLE payments DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at;
ALTER TABLE signals DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE whale_transfers DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE orderbook_snapshots DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE liquidations DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE derivatives DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;

SELECT remove_compression_policy('klines', if_exists => TRUE);
ALTER TABLE klines SET (timescaledb.compress = FALSE);
ALTER TABLE klines DROP COLUMN IF EXISTS updated_by, DROP COLUMN IF EXISTS created_by, DROP COLUMN IF EXISTS deleted_at, DROP COLUMN IF EXISTS updated_at, DROP COLUMN IF EXISTS created_at;
ALTER TABLE klines SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'symbol, interval'
);
SELECT add_compression_policy('klines', INTERVAL '30 days');
