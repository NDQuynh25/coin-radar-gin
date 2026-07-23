CREATE MATERIALIZED VIEW klines_1m WITH (timescaledb.continuous) AS
SELECT time_bucket('1 minute', time) AS bucket, exchange, symbol,
       first(price, time) AS open, max(price) AS high, min(price) AS low,
       last(price, time) AS close, sum(qty) AS volume, count(*) AS trade_count
FROM trades GROUP BY bucket, exchange, symbol WITH NO DATA;
SELECT add_continuous_aggregate_policy('klines_1m', start_offset => INTERVAL '10 minutes',
    end_offset => INTERVAL '1 minute', schedule_interval => INTERVAL '1 minute');
