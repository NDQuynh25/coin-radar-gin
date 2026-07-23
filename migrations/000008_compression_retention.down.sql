SELECT remove_compression_policy('klines', if_exists => TRUE);
ALTER TABLE klines SET (timescaledb.compress = FALSE);
SELECT remove_retention_policy('orderbook_snapshots', if_exists => TRUE);
SELECT remove_retention_policy('trades', if_exists => TRUE);
SELECT remove_compression_policy('trades', if_exists => TRUE);
ALTER TABLE trades SET (timescaledb.compress = FALSE);
