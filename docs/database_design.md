# Thiết Kế Cơ Sở Dữ Liệu — Crypto Data Platform

> DBMS: **PostgreSQL 16 + TimescaleDB Extension**
> Driver Go: **pgx** · Code gen: **sqlc** · Migration: **golang-migrate** hoặc **goose**
> Liên quan task: **T1.2** (DB Schema), **T2.3** (Partitioning), **T2.4** (Aggregation)

---

## 1. Tổng Quan

Database chia làm **3 nhóm** theo tính chất dữ liệu:

| Nhóm | Đặc điểm | Bảng | Loại |
|---|---|---|---|
| **Market Data** | Ghi rất nhiều, append-only, chuỗi thời gian | `trades`, `klines`, `derivatives`, `orderbook_snapshots` | **Hypertable** |
| **On-chain & Signals** | Sự kiện rời rạc, query nhiều | `whale_transfers`, `signals` | **Hypertable** |
| **Application / SaaS** | Ít record, đọc/ghi cân bằng, quan hệ | `users`, `subscriptions`, `alert_rules`, `watchlists`, `known_wallets` | **Bảng thường** |

**Nguyên tắc:**
- Dữ liệu thị trường = **hypertable** (Timescale tự phân vùng theo `time`).
- Dữ liệu nghiệp vụ = bảng quan hệ chuẩn (có khóa ngoại, ràng buộc).
- Tách rõ 2 nhóm để retention/nén chỉ áp dụng cho market data, không đụng nghiệp vụ.

---

## 2. Sơ Đồ Quan Hệ (ERD rút gọn)

```
┌──────────────┐        ┌────────────────┐        ┌──────────────┐
│    users     │1──────*│ subscriptions  │        │  alert_rules │
│──────────────│        │────────────────│        │──────────────│
│ id (PK)      │        │ id (PK)        │   ┌────*│ id (PK)      │
│ telegram_id  │        │ user_id (FK)   │   │     │ user_id (FK) │
│ email        │        │ plan           │   │     │ type         │
│ created_at   │        │ status         │   │     │ symbol       │
└──────┬───────┘        │ expires_at     │   │     │ condition    │
       │1               └────────────────┘   │     │ threshold    │
       │                                      │     │ is_active    │
       ├──────────────────────────────────────┘     └──────────────┘
       │*
┌──────────────┐
│  watchlists  │
│──────────────│
│ id (PK)      │
│ user_id (FK) │
│ symbol       │
└──────────────┘

────────── MARKET DATA (hypertable, không FK) ──────────
 trades   klines   derivatives   orderbook_snapshots   whale_transfers   signals

────────── REFERENCE ──────────
 symbols (danh mục cặp giao dịch)     known_wallets (nhãn ví on-chain)
```

> Market data **không dùng khóa ngoại** tới `symbols` — vì hypertable insert tốc độ cao, FK sẽ làm chậm. Dùng `symbol TEXT` + validate ở tầng app.

---

## 3. Nhóm Market Data (Hypertables)

### 3.1. `trades` — Giao dịch thô realtime

```sql
CREATE TABLE trades (
    time        TIMESTAMPTZ      NOT NULL,
    exchange    TEXT             NOT NULL,   -- binance / bybit / okx
    symbol      TEXT             NOT NULL,   -- BTCUSDT
    trade_id    BIGINT,
    price       DOUBLE PRECISION NOT NULL,
    qty         DOUBLE PRECISION NOT NULL,
    quote_qty   DOUBLE PRECISION,            -- price * qty
    side        TEXT             NOT NULL     -- buy / sell (taker side)
);

SELECT create_hypertable('trades', 'time', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX idx_trades_symbol_time ON trades (symbol, time DESC);
CREATE INDEX idx_trades_exchange_time ON trades (exchange, time DESC);
```

### 3.2. `klines` — Nến (OHLCV)

```sql
CREATE TABLE klines (
    time        TIMESTAMPTZ      NOT NULL,   -- open time của nến
    exchange    TEXT             NOT NULL,
    symbol      TEXT             NOT NULL,
    interval    TEXT             NOT NULL,   -- 1m / 5m / 15m / 1h / 4h / 1d
    open        DOUBLE PRECISION NOT NULL,
    high        DOUBLE PRECISION NOT NULL,
    low         DOUBLE PRECISION NOT NULL,
    close       DOUBLE PRECISION NOT NULL,
    volume      DOUBLE PRECISION NOT NULL,
    trade_count INTEGER,
    PRIMARY KEY (time, exchange, symbol, interval)
);

SELECT create_hypertable('klines', 'time', chunk_time_interval => INTERVAL '7 days');

CREATE INDEX idx_klines_sym_int_time ON klines (symbol, interval, time DESC);
```

### 3.3. `derivatives` — Funding Rate & Open Interest

```sql
CREATE TABLE derivatives (
    time              TIMESTAMPTZ      NOT NULL,
    exchange          TEXT             NOT NULL,
    symbol            TEXT             NOT NULL,
    funding_rate      DOUBLE PRECISION,        -- vd 0.0001 = 0.01%
    next_funding_time TIMESTAMPTZ,
    open_interest     DOUBLE PRECISION,        -- giá trị OI (USD hoặc contracts)
    mark_price        DOUBLE PRECISION,
    index_price       DOUBLE PRECISION
);

SELECT create_hypertable('derivatives', 'time', chunk_time_interval => INTERVAL '1 day');

CREATE INDEX idx_deriv_sym_time ON derivatives (symbol, time DESC);
```

### 3.4. `liquidations` — Lệnh thanh lý (Bybit/Binance)

```sql
CREATE TABLE liquidations (
    time      TIMESTAMPTZ      NOT NULL,
    exchange  TEXT             NOT NULL,
    symbol    TEXT             NOT NULL,
    side      TEXT             NOT NULL,   -- long / short bị thanh lý
    price     DOUBLE PRECISION NOT NULL,
    qty       DOUBLE PRECISION NOT NULL,
    value_usd DOUBLE PRECISION
);

SELECT create_hypertable('liquidations', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_liq_sym_time ON liquidations (symbol, time DESC);
```

### 3.5. `orderbook_snapshots` — Ảnh chụp sổ lệnh (OKX, tùy chọn)

```sql
CREATE TABLE orderbook_snapshots (
    time      TIMESTAMPTZ NOT NULL,
    exchange  TEXT        NOT NULL,
    symbol    TEXT        NOT NULL,
    bids      JSONB       NOT NULL,   -- [[price, qty], ...] top N
    asks      JSONB       NOT NULL
);

SELECT create_hypertable('orderbook_snapshots', 'time', chunk_time_interval => INTERVAL '6 hours');
```

---

## 4. Nhóm On-chain & Tín Hiệu (Hypertables)

### 4.1. `whale_transfers` — Giao dịch cá mập

```sql
CREATE TABLE whale_transfers (
    time           TIMESTAMPTZ      NOT NULL,
    chain          TEXT             NOT NULL,   -- ethereum / bsc
    tx_hash        TEXT             NOT NULL,
    token_address  TEXT,
    token_symbol   TEXT,
    from_address   TEXT             NOT NULL,
    to_address     TEXT             NOT NULL,
    amount         NUMERIC,                      -- số token (precision cao)
    value_usd      DOUBLE PRECISION,
    direction      TEXT                          -- inflow_exchange / outflow_exchange / wallet
);

SELECT create_hypertable('whale_transfers', 'time', chunk_time_interval => INTERVAL '1 day');
CREATE INDEX idx_whale_token_time ON whale_transfers (token_symbol, time DESC);
CREATE INDEX idx_whale_value ON whale_transfers (value_usd DESC, time DESC);
```

### 4.2. `signals` — Tín hiệu phát hiện bởi Indicator Engine

```sql
CREATE TABLE signals (
    time       TIMESTAMPTZ      NOT NULL,
    id         BIGINT GENERATED ALWAYS AS IDENTITY,
    type       TEXT             NOT NULL,   -- funding_spike / oi_delta / liquidation_spike
                                            -- / volume_spike / whale_alert / rsi_extreme
    exchange   TEXT,
    symbol     TEXT             NOT NULL,
    severity   TEXT             NOT NULL,   -- info / warning / critical
    value      DOUBLE PRECISION,            -- giá trị đo được
    threshold  DOUBLE PRECISION,            -- ngưỡng đã vượt
    payload    JSONB,                       -- dữ liệu phụ (chi tiết tín hiệu)
    PRIMARY KEY (time, id)
);

SELECT create_hypertable('signals', 'time', chunk_time_interval => INTERVAL '7 days');
CREATE INDEX idx_signals_type_time ON signals (type, time DESC);
CREATE INDEX idx_signals_symbol_time ON signals (symbol, time DESC);
```

---

## 5. Nhóm Application / SaaS (Bảng Quan Hệ)

+
### 5.1. `users`

```sql
CREATE TABLE users (
    id           BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    telegram_id  BIGINT UNIQUE,
    email        TEXT UNIQUE,
    username     TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
```

### 5.2. `subscriptions` — Gói Premium (T4.3)

```sql
CREATE TABLE subscriptions (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan        TEXT   NOT NULL DEFAULT 'free',   -- free / premium / pro
    status      TEXT   NOT NULL DEFAULT 'active', -- active / expired / canceled
    started_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    expires_at  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_subs_user ON subscriptions (user_id);
CREATE INDEX idx_subs_status ON subscriptions (status, expires_at);
```

### 5.3. `payments` — Thanh toán crypto (USDT BSC/TRON)

```sql
CREATE TABLE payments (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    amount_usd  DOUBLE PRECISION NOT NULL,
    currency    TEXT NOT NULL DEFAULT 'USDT',
    network     TEXT NOT NULL,              -- BSC / TRON
    tx_hash     TEXT UNIQUE,
    status      TEXT NOT NULL DEFAULT 'pending', -- pending / confirmed / failed
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    confirmed_at TIMESTAMPTZ
);
CREATE INDEX idx_payments_user ON payments (user_id, created_at DESC);
```

### 5.4. `alert_rules` — Custom Alerts (T4.2)

```sql
CREATE TABLE alert_rules (
    id          BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id     BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    type        TEXT   NOT NULL,            -- price / funding / volume / whale ...
    symbol      TEXT   NOT NULL,
    condition   TEXT   NOT NULL,            -- gt / lt / cross_up / cross_down
    threshold   DOUBLE PRECISION NOT NULL,
    channel     TEXT   NOT NULL DEFAULT 'telegram',
    is_active   BOOLEAN NOT NULL DEFAULT true,
    cooldown_s  INTEGER NOT NULL DEFAULT 300,  -- chống spam: giây giữa 2 lần báo
    last_fired  TIMESTAMPTZ,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_rules_active ON alert_rules (is_active, symbol, type);
CREATE INDEX idx_rules_user ON alert_rules (user_id);
```

### 5.5. `watchlists` — Danh sách theo dõi

```sql
CREATE TABLE watchlists (
    id         BIGINT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    symbol     TEXT   NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE (user_id, symbol)
);
```

---

## 6. Bảng Tham Chiếu (Reference)

### 6.1. `symbols` — Danh mục cặp giao dịch

```sql
CREATE TABLE symbols (
    symbol        TEXT PRIMARY KEY,          -- BTCUSDT
    base_asset    TEXT NOT NULL,             -- BTC
    quote_asset   TEXT NOT NULL,             -- USDT
    exchange      TEXT NOT NULL,
    market_type   TEXT NOT NULL,             -- spot / futures
    is_active     BOOLEAN NOT NULL DEFAULT true,
    listed_at     TIMESTAMPTZ
);
```

### 6.2. `known_wallets` — Nhãn ví on-chain (cho Whale Tracking)

```sql
CREATE TABLE known_wallets (
    address     TEXT PRIMARY KEY,
    chain       TEXT NOT NULL,
    label       TEXT,                        -- "Binance Hot Wallet", "Whale #1"
    category    TEXT,                        -- exchange / whale / smart_money / contract
    is_exchange BOOLEAN NOT NULL DEFAULT false,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX idx_wallets_category ON known_wallets (category);
```

---

## 7. Continuous Aggregates (T2.4 — Rollup Nến Tự Động)

TimescaleDB tự rollup, **không cần cronjob thủ công**:

```sql
-- Nến 1 phút từ trades thô
CREATE MATERIALIZED VIEW klines_1m
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 minute', time) AS bucket,
       exchange, symbol,
       first(price, time) AS open,
       max(price)         AS high,
       min(price)         AS low,
       last(price, time)  AS close,
       sum(qty)           AS volume,
       count(*)           AS trade_count
FROM trades
GROUP BY bucket, exchange, symbol
WITH NO DATA;

-- Tự refresh: cập nhật dữ liệu 1m gần nhất mỗi phút
SELECT add_continuous_aggregate_policy('klines_1m',
    start_offset => INTERVAL '10 minutes',
    end_offset   => INTERVAL '1 minute',
    schedule_interval => INTERVAL '1 minute');

-- Nến 5m & 1h: rollup TIẾP từ klines_1m (hierarchical aggregate) cho nhẹ
CREATE MATERIALIZED VIEW klines_5m
WITH (timescaledb.continuous) AS
SELECT time_bucket('5 minutes', bucket) AS bucket,
       exchange, symbol,
       first(open, bucket) AS open,
       max(high)           AS high,
       min(low)            AS low,
       last(close, bucket) AS close,
       sum(volume)         AS volume
FROM klines_1m
GROUP BY 1, exchange, symbol
WITH NO DATA;

SELECT add_continuous_aggregate_policy('klines_5m',
    start_offset => INTERVAL '1 hour',
    end_offset   => INTERVAL '5 minutes',
    schedule_interval => INTERVAL '5 minutes');
```

> Tương tự tạo `klines_1h`, `funding_1h` (avg funding theo giờ), `oi_1h` (max OI theo giờ) để phục vụ phân tích lịch sử dài.

---

## 8. Nén & Giải Phóng Dung Lượng (T2.4 — Quản Lý Ổ Cứng)

VPS rẻ → **bắt buộc** nén dữ liệu cũ và xóa raw không cần thiết.

```sql
-- Bật nén cho trades (dữ liệu cũ hơn 1 ngày được nén ~90%)
ALTER TABLE trades SET (
    timescaledb.compress,
    timescaledb.compress_segmentby = 'exchange, symbol',
    timescaledb.compress_orderby   = 'time DESC'
);
SELECT add_compression_policy('trades', INTERVAL '1 day');

-- Xóa raw trades sau 7 ngày (nến đã rollup vào continuous aggregate, giữ lại nến)
SELECT add_retention_policy('trades', INTERVAL '7 days');

-- Orderbook snapshots rất nặng → giữ ngắn
SELECT add_retention_policy('orderbook_snapshots', INTERVAL '2 days');

-- Nến giữ lâu (nhẹ), funding/OI giữ lâu để phân tích
ALTER TABLE klines SET (timescaledb.compress, timescaledb.compress_segmentby = 'symbol, interval');
SELECT add_compression_policy('klines', INTERVAL '30 days');
```

**Chiến lược retention tổng quát:**

| Bảng | Nén sau | Xóa raw sau | Ghi chú |
|---|---|---|---|
| `trades` | 1 ngày | 7 ngày | Đã rollup thành nến |
| `orderbook_snapshots` | — | 2 ngày | Rất nặng |
| `klines` | 30 ngày | giữ vô hạn | Nhẹ, giá trị lịch sử cao |
| `derivatives` | 7 ngày | giữ ~1 năm | Phân tích funding lịch sử |
| `liquidations` | 7 ngày | giữ ~1 năm | |
| `signals` | 30 ngày | giữ ~1 năm | Audit/backtest |
| `whale_transfers` | 30 ngày | giữ vô hạn | Giá trị insight cao |

---

## 9. Chiến Lược Index

| Pattern truy vấn | Index |
|---|---|
| Lấy data 1 symbol theo thời gian (phổ biến nhất) | `(symbol, time DESC)` trên mọi hypertable |
| Lọc theo loại tín hiệu | `(type, time DESC)` trên `signals` |
| Whale theo giá trị lớn | `(value_usd DESC, time DESC)` |
| Rule đang active để khớp realtime | `(is_active, symbol, type)` |
| Sub sắp hết hạn (cronjob check) | `(status, expires_at)` |

> Timescale tự index theo `time` (chunk). Chỉ thêm index phụ cho cột hay lọc — đừng over-index vì làm chậm insert.

---

## 10. Lưu Ý Vận Hành & Hiệu Năng

1. **Batch insert** market data (gom 100-1000 row/lần) qua `pgx.CopyFrom` thay vì insert từng dòng → nhanh gấp nhiều lần.
2. **Không dùng FK trên hypertable** — làm chậm insert tốc độ cao.
3. **`NUMERIC` chỉ cho `amount` on-chain** (cần precision tuyệt đối); giá/volume dùng `DOUBLE PRECISION` cho nhanh.
4. **Connection pool:** dùng `pgxpool`, tách pool đọc (api) và pool ghi (ingestor) nếu cần.
5. **Cache lớp Redis:** giá mới nhất / kết quả phân tích hot → đọc từ Redis, tránh đụng DB cho mỗi request Dashboard.
6. **Migration versioned:** mọi thay đổi schema qua file migration đánh số, không sửa tay DB production.

---

## 11. Thứ Tự Migration Đề Xuất

```
migrations/
├── 001_enable_timescaledb.sql        -- CREATE EXTENSION timescaledb
├── 002_market_data_hypertables.sql   -- trades, klines, derivatives, liquidations
├── 003_onchain_signals.sql           -- whale_transfers, signals
├── 004_app_saas_tables.sql           -- users, subscriptions, payments
├── 005_alert_watchlist.sql           -- alert_rules, watchlists
├── 006_reference_tables.sql          -- symbols, known_wallets
├── 007_continuous_aggregates.sql     -- klines_1m / 5m / 1h
└── 008_compression_retention.sql     -- policies nén + xóa
```
