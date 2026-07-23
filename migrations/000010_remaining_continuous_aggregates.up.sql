CREATE MATERIALIZED VIEW klines_5m
WITH (timescaledb.continuous) AS
SELECT time_bucket('5 minutes', bucket) AS bucket,
       exchange,
       symbol,
       first(open, bucket) AS open,
       max(high) AS high,
       min(low) AS low,
       last(close, bucket) AS close,
       sum(volume) AS volume,
       sum(trade_count) AS trade_count
FROM klines_1m
GROUP BY 1, exchange, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'klines_5m',
    start_offset => INTERVAL '1 hour',
    end_offset => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes'
);

CREATE MATERIALIZED VIEW klines_1h
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', bucket) AS bucket,
       exchange,
       symbol,
       first(open, bucket) AS open,
       max(high) AS high,
       min(low) AS low,
       last(close, bucket) AS close,
       sum(volume) AS volume,
       sum(trade_count) AS trade_count
FROM klines_5m
GROUP BY 1, exchange, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'klines_1h',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);

CREATE MATERIALIZED VIEW funding_1h
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', time) AS bucket,
       exchange,
       symbol,
       avg(funding_rate) AS funding_rate_avg,
       min(funding_rate) AS funding_rate_min,
       max(funding_rate) AS funding_rate_max,
       count(funding_rate) AS sample_count
FROM derivatives
GROUP BY 1, exchange, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'funding_1h',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);

CREATE MATERIALIZED VIEW oi_1h
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 hour', time) AS bucket,
       exchange,
       symbol,
       first(open_interest, time) AS open_interest_open,
       max(open_interest) AS open_interest_high,
       min(open_interest) AS open_interest_low,
       last(open_interest, time) AS open_interest_close,
       avg(open_interest) AS open_interest_avg,
       count(open_interest) AS sample_count
FROM derivatives
GROUP BY 1, exchange, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy(
    'oi_1h',
    start_offset => INTERVAL '1 day',
    end_offset => INTERVAL '1 hour',
    schedule_interval => INTERVAL '1 hour'
);
