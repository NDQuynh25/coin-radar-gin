ALTER TABLE trades SET (timescaledb.compress, timescaledb.compress_segmentby = 'exchange, symbol', timescaledb.compress_orderby = 'time DESC');
SELECT add_compression_policy('trades', INTERVAL '1 day');
SELECT add_retention_policy('trades', INTERVAL '7 days');
SELECT add_retention_policy('orderbook_snapshots', INTERVAL '2 days');
ALTER TABLE klines SET (timescaledb.compress, timescaledb.compress_segmentby = 'symbol, interval');
SELECT add_compression_policy('klines', INTERVAL '30 days');
