# Thiết Kế Kiến Trúc — Crypto Data Platform

> Nền tảng thu thập & phân tích dữ liệu crypto realtime.
> Mô hình: **Lean MVP / SaaS** · Team: **2 người** · VPS: **~50 USD/tháng** · Thời gian: **6 tháng**

---

## 1. Quyết Định Công Nghệ (Đã Chốt)

| Hạng mục | Lựa chọn | Lý do |
|---|---|---|
| Ngôn ngữ Backend | **Go (Golang)** | Hiệu năng cao, RAM thấp, goroutines hợp xử lý nhiều WebSocket song song |
| Web/API framework | **Gin** | Phổ biến nhất, nhiều tài liệu, dùng `net/http` chuẩn, hợp người mới học Go |
| WebSocket client | **coder/websocket** | Hiện đại, context-aware, không phụ thuộc framework |
| Message Queue | **Redis + asynq** | API gần giống BullMQ, đơn giản, đủ tải cho MVP |
| Cache | **Redis** | Dùng chung Redis với queue cho gọn hạ tầng |
| Time-series DB | **TimescaleDB** (+ pgx + sqlc) | Chuyên dụng chuỗi thời gian, continuous aggregates rollup nến tự động |
| Frontend | **Next.js + TailwindCSS** | SSR/ISR tốt cho SEO, UI mượt |
| Visualization | **Lightweight Charts (TradingView)** | Render mượt hàng vạn nến |
| DevOps | **Docker + GitHub Actions + VPS** | Image Go ~15-20MB, deploy gọn |
| **Kiến trúc** | **Modular Monolith + Multi-binary** | KHÔNG microservice — hợp team nhỏ + VPS rẻ |

---

## 2. Nguyên Tắc Kiến Trúc

1. **Modular Monolith, không Microservice.** Một codebase, nhiều binary, share chung `internal/`. Tách process độc lập để scale riêng nhưng không gánh chi phí vận hành của microservice (service discovery, gRPC, distributed tracing).
2. **Giao tiếp qua Queue, không qua RPC.** Các process nói chuyện gián tiếp qua Redis/asynq → loose coupling, không cần network API giữa các service.
3. **Tính toán streaming trong RAM, không query DB mỗi tick.** Dùng rolling window (ring buffer) cập nhật incremental O(1).
4. **Interface-first.** Mọi điểm có thể đổi (queue, exchange adapter, detector) đều ẩn sau interface → dễ thay thế/test, dễ tách service sau này nếu cần.
5. **Đừng tối ưu cho tải chưa có.** Bắt đầu đơn giản, chỉ phức tạp hóa khi gặp giới hạn thật.

---

## 3. Sơ Đồ Kiến Trúc Tổng Thể

```
                        SÀN / NGUỒN DỮ LIỆU
   Binance   Bybit   OKX      ETH/BSC RPC      GeckoTerminal   X(Twitter)
     │         │       │            │                │             │
     │  WebSocket / REST / JSON-RPC                                │
     ▼         ▼       ▼            ▼                ▼             ▼
┌─────────────────────────────────────────────────────────────────────┐
│  cmd/ingestor   (1..N process, mỗi sàn/nhóm 1 goroutine kết nối)     │
│  - Hứng tick realtime (trades, tickers, klines, funding, OI)        │
│  - Chuẩn hóa về model chung → đẩy vào Queue                          │
└───────────────────────────────┬─────────────────────────────────────┘
                                 │ publish
                                 ▼
                        ┌──────────────────┐
                        │   Redis + asynq  │  ◄── Message Queue + Cache
                        └────────┬─────────┘
              consume            │            consume
        ┌──────────────────────┼───────────────────────┐
        ▼                       ▼                        ▼
┌────────────────┐   ┌─────────────────────┐   ┌──────────────────┐
│ cmd/aggregator │   │  Indicator Engine   │   │  (ghi raw data)  │
│ Cronjob gộp    │   │  (trong ingestor /  │   │                  │
│ nến 1m/5m/1h   │   │   worker riêng)     │   │                  │
│ + rollup       │   │ Detector → Signal   │   │                  │
└───────┬────────┘   └──────────┬──────────┘   └────────┬─────────┘
        │                       │ signal                 │
        │ write                 │ publish alert          │ write
        ▼                       ▼                        ▼
┌─────────────────────────────────────────────────────────────────────┐
│                 TimescaleDB (PostgreSQL + Timescale)                 │
│   hypertables: trades, klines, funding, open_interest, signals      │
│   continuous aggregates: klines_1m / 5m / 1h                        │
└─────────────────────────────────────────────────────────────────────┘
        ▲                                              │
        │ query (qua cache Redis)                      │ alert event
        │                                              ▼
┌────────────────┐                          ┌─────────────────────┐
│   cmd/api      │  REST + WS               │     cmd/bot         │
│   (Gin)        │ ───────────────►         │  Telegram Bot       │
│  trả phân tích │                          │  gửi cảnh báo        │
└───────┬────────┘                          └─────────────────────┘
        │ HTTP/WS
        ▼
┌────────────────┐
│  Frontend      │  Next.js + Lightweight Charts
│  Dashboard     │
└────────────────┘
```

---

## 4. Các Thành Phần (Multi-binary)

Mỗi `cmd/*` build ra **một binary độc lập**, deploy & scale riêng, nhưng share chung code trong `internal/`.

| Binary | Nhiệm vụ | Scale |
|---|---|---|
| **ingestor** | Kết nối WebSocket/API các sàn, hứng tick realtime, chuẩn hóa, đẩy vào queue | Chạy nhiều bản theo nhóm sàn nếu tải cao |
| **aggregator** | Cronjob nén/gộp dữ liệu thô → nến 1m/5m/1h, rollup chỉ số dài hạn | 1 bản |
| **api** | REST API (Gin) + WebSocket server trả kết quả phân tích cho Frontend/Bot | Scale theo lượng user |
| **bot** | Telegram Bot gửi tín hiệu cảnh báo tự động | 1 bản |

> Indicator Engine (phát hiện tín hiệu) có thể nằm chung trong `ingestor` ở giai đoạn đầu, tách thành `cmd/analyzer` riêng khi tải lớn.

---

## 5. Cấu Trúc Thư Mục (Clean Architecture - DDD)

```
coin-radar/
├── cmd/                         # Entry points cho các binary độc lập
│   ├── api/                     # REST API (Gin) & WebSocket Server (T3.1)
│   │   └── main.go
│   ├── ingestor/                # Worker WebSocket nhận dữ liệu sàn -> Queue/DB (T2.1)
│   │   └── main.go
│   ├── aggregator/              # Worker chạy các tác vụ định kỳ & nén dữ liệu DB (T2.4)
│   │   └── main.go
│   └── bot/                     # Telegram Bot gửi cảnh báo realtime (T3.2)
│       └── main.go
├── internal/                    # Logic nghiệp vụ & code private
│   ├── domain/                  # Lớp 1 - Domain (Không phụ thuộc gì khác, chứa interface/entity lõi)
│   │   ├── model/               # Struct dữ liệu dùng chung (trades, klines, signals)
│   │   ├── repository/          # Interface kết nối DB/Storage
│   │   └── cache/               # Interface lưu trữ Cache
│   ├── application/             # Lớp 2 - Use Cases (Logic xử lý nghiệp vụ)
│   │   ├── alert_service.go     # Xử lý luật cảnh báo (Rule Engine)
│   │   ├── indicator_service.go # Tính toán chỉ báo (RSI, Funding...)
│   │   └── ingest_service.go    # Điều phối nhận & lưu trữ dữ liệu
│   ├── infrastructure/          # Lớp 3 - Triển khai kỹ thuật cụ thể (Database, Redis, Queue adapter)
│   │   ├── storage/             # TimescaleDB & Redis connections
│   │   │   ├── timescale/       # Code Go sinh tự động từ sqlc
│   │   │   └── redis/           # go-redis client wrapper
│   │   ├── repository/          # Implements domain/repository interfaces
│   │   ├── cache/               # Implements domain/cache interfaces
│   │   └── queue/               # Cài đặt Publisher/Subscriber (asynq wrapper)
│   ├── interfaces/              # Lớp 4 - Tầng giao tiếp (Giao diện với bên ngoài)
│   │   ├── http/                # REST API handlers & router (Gin)
│   │   │   ├── handlers/
│   │   │   └── routes.go
│   │   ├── ws/                  # WebSocket streaming handlers cho Dashboard
│   │   └── telegram/            # Bot handler xử lý lệnh & push notifications
│   └── config/                  # Load file config (Viper wrapper)
│       └── config.go
├── migrations/                  # Tệp Migration SQL (TimescaleDB schema)
├── sqlc.yaml                    # Cấu hình SQL compiler (sqlc)
├── Makefile                     # Shortcut commands (run, build, migrate, gen-sqlc)
├── docker-compose.yml           # local dev: TimescaleDB + Redis
├── config.yaml                  # File config dev
├── go.mod                       # Go module file
└── README.md
```

---

## 6. Luồng Dữ Liệu (Data Flow)

1. **Ingest:** `ingestor` mở WebSocket tới sàn → mỗi kết nối 1 goroutine → nhận tick (trade/ticker/funding/OI).
2. **Normalize:** chuyển payload riêng của từng sàn về `model` chung (cùng schema bất kể nguồn).
3. **Publish:** đẩy tick chuẩn hóa vào Redis/asynq.
4. **Persist:** consumer ghi raw vào hypertable TimescaleDB.
5. **Analyze (streaming):** Indicator Engine cập nhật rolling window trong RAM → khi vượt ngưỡng sinh `Signal`.
6. **Aggregate (batch):** `aggregator` định kỳ gộp nến + dùng continuous aggregates của Timescale rollup nến 1m/5m/1h.
7. **Alert:** `Signal` → Rule Engine so ngưỡng (cả ngưỡng custom của user) → Dispatcher đẩy ra Telegram + WS Dashboard.
8. **Serve:** `api` (Gin) đọc kết quả (ưu tiên qua cache Redis) → trả Frontend/Bot.

---

## 7. Lược Đồ Cơ Sở Dữ Liệu (TimescaleDB)

Các bảng chính dùng **hypertable** (phân vùng theo thời gian) + **partitioning** (T2.3):

```sql
-- Trades thô realtime
CREATE TABLE trades (
    time        TIMESTAMPTZ      NOT NULL,
    exchange    TEXT             NOT NULL,
    symbol      TEXT             NOT NULL,
    price       DOUBLE PRECISION NOT NULL,
    qty         DOUBLE PRECISION NOT NULL,
    side        TEXT             NOT NULL  -- buy/sell
);
SELECT create_hypertable('trades', 'time');

-- Nến (klines) — thô, aggregator/continuous aggregate sinh 1m/5m/1h
CREATE TABLE klines (
    time        TIMESTAMPTZ      NOT NULL,
    exchange    TEXT             NOT NULL,
    symbol      TEXT             NOT NULL,
    interval    TEXT             NOT NULL,
    open        DOUBLE PRECISION,
    high        DOUBLE PRECISION,
    low         DOUBLE PRECISION,
    close       DOUBLE PRECISION,
    volume      DOUBLE PRECISION
);
SELECT create_hypertable('klines', 'time');

-- Funding rate & Open Interest (phái sinh)
CREATE TABLE derivatives (
    time           TIMESTAMPTZ NOT NULL,
    exchange       TEXT        NOT NULL,
    symbol         TEXT        NOT NULL,
    funding_rate   DOUBLE PRECISION,
    open_interest  DOUBLE PRECISION
);
SELECT create_hypertable('derivatives', 'time');

-- Tín hiệu phát hiện (signals/alerts)
CREATE TABLE signals (
    time       TIMESTAMPTZ NOT NULL,
    type       TEXT        NOT NULL,  -- funding_spike / oi_delta / liquidation / whale ...
    symbol     TEXT        NOT NULL,
    severity   TEXT        NOT NULL,  -- info / warning / critical
    payload    JSONB
);
SELECT create_hypertable('signals', 'time');

-- Continuous aggregate: nến 1m tự rollup (T2.4)
CREATE MATERIALIZED VIEW klines_1m
WITH (timescaledb.continuous) AS
SELECT time_bucket('1 minute', time) AS bucket,
       exchange, symbol,
       first(price, time) AS open,
       max(price)         AS high,
       min(price)         AS low,
       last(price, time)  AS close,
       sum(qty)           AS volume
FROM trades
GROUP BY bucket, exchange, symbol;

-- Cơ chế giải phóng dung lượng ổ cứng: nén + xóa dữ liệu thô cũ
SELECT add_retention_policy('trades', INTERVAL '7 days');
```

---

## 8. Thuật Toán Phân Tích (Core Value — T2.5)

Tách 3 lớp để dễ mở rộng và phục vụ luôn **Custom Alerts (T4.2)**:

```
Tick → [Detector] → Signal → [Rule Engine] → [Dispatcher] → Telegram / WS
```

- **Detector:** mỗi tín hiệu là struct implement `interface Detector { Detect(tick) *Signal }`.
- **Rule Engine:** so `Signal` với ngưỡng mặc định + ngưỡng người dùng tự đặt.
- **Dispatcher:** đẩy alert ra các kênh qua queue.

**Ưu tiên triển khai (giá trị / công sức):**

| # | Nhóm | Tín hiệu | Độ khó |
|---|---|---|---|
| 🥇 | Phái sinh | Funding Rate bất thường, OI Delta, Liquidation Spike, Long/Short Ratio | Dễ, data sẵn |
| 🥈 | Kỹ thuật | Volume Spike, RSI/EMA/MACD (dùng lib `cinar/indicator`) | Dễ |
| 🥉 | On-chain | Whale Transfer Alert, Exchange Inflow/Outflow | Khó vừa (cần map ví) |
| 4 | Social | Mention frequency spike | Làm sau (P2) |

**Kỹ thuật tính toán:**
- **Rolling window (RAM):** ring buffer cập nhật O(1) cho trung bình động (funding, volume, OI delta) — không query DB mỗi tick.
- **Batch dài hạn:** so sánh với lịch sử 7-30 ngày → dùng `time_bucket()` + continuous aggregates của Timescale.

---

## 9. Nguồn Dữ Liệu Cần Kết Nối

| Loại | Đối tác | Phương thức | Dữ liệu | Ưu tiên |
|---|---|---|---|---|
| CEX | Binance | WebSocket Streams | Ticker, Klines, Trades, Funding Rate, Open Interest | **P0** |
| CEX | Bybit | REST + WebSocket | Phái sinh, Liquidation (Long/Short) | P1 |
| CEX | OKX | WebSocket Streams | Orderbook depth, dòng tiền phái sinh | P2 |
| DEX | ETH/BSC RPC | JSON-RPC / Web3 | Whale transfer tracking | P1 |
| DEX | GeckoTerminal / DexScreener | REST | Giá & volume token mới list | P1 |
| Social | X (Twitter) API v2 | REST | Social listening KOLs | P2 |

---

## 10. Triển Khai (Deployment)

- **Local dev:** `docker-compose up` → Postgres+TimescaleDB + Redis.
- **Build:** multi-stage Dockerfile → mỗi binary 1 image nhỏ (~15-20MB).
- **CI/CD:** GitHub Actions build + test + push image → deploy lên VPS (Hetzner/DigitalOcean).
- **Production (VPS 50$):** chạy tất cả binary + Timescale + Redis trên 1 VPS bằng docker-compose. Tách scale `ingestor` khi tải cao.

---

## 11. Lộ Trình (tóm tắt 6 tháng)

| GĐ | Tháng | Trọng tâm |
|---|---|---|
| 1 | Tháng 1 | Nghiên cứu API, thiết kế DB schema, wireframe, setup môi trường (15%) |
| 2 | Tháng 2-3 | Core: ingestor WebSocket, queue, partitioning, aggregation, thuật toán (40%) |
| 3 | Tháng 4 | MVP: REST API, Telegram Bot, Dashboard Web, deploy + alpha test (25%) |
| 4 | Tháng 5-6 | Tối ưu, Custom Alerts, thanh toán crypto Premium, marketing (20%) |

---

## 12. Quy Tắc Vàng

> **Bắt đầu bằng monolith module hóa tốt. Chỉ tách microservice KHI có vấn đề thật buộc phải tách.**
> Tập trung 100% công sức vào **chất lượng thuật toán phân tích** — đó mới là thứ tạo ra tiền.
