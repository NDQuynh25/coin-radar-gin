# Kế Hoạch Phát Triển — Crypto Data Platform

> Cụ thể hóa lộ trình 6 tháng (T1.1 → T4.4) thành **sprint + task code + Definition of Done (DoD)**.
> Stack: **Go + Gin + Redis/asynq + TimescaleDB**. Kiến trúc: **Modular Monolith + Multi-binary**.
> Team: **Dev A (Backend/Infra)** · **Dev B (Frontend/Analyst)** · Sprint = **2 tuần**.

---

## 0. Quy Ước Làm Việc

- **Branch model:** `main` (ổn định) ← `dev` ← `feature/*`. PR review chéo trước khi merge.
- **Mỗi feature:** code + test + cập nhật migration/docs trước khi đóng task.
- **DoD chung:** code chạy được local (`docker-compose up`), test pass, không lỗi `go vet`/`golangci-lint`, có log rõ ràng.
- **Ưu tiên:** Pipeline dữ liệu chạy thật > tính năng đẹp. "Chất lượng thuật toán phân tích = giá trị cốt lõi."

---

## GIAI ĐOẠN 1 — NGHIÊN CỨU & ĐỊNH HÌNH (Tháng 1)

### Sprint 1 (Tuần 1-2) — Khảo sát & Nền tảng Go
**Mục tiêu:** Hiểu API các sàn + dựng được skeleton chạy "hello".

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 1.1 Khảo sát API/WS các sàn | A+B | Liệt kê endpoint Binance/Bybit/OKX cần (trades, klines, funding, OI, liquidation). Lưu vào `docs/api-research.md` | Có bảng endpoint + rate limit + format payload mẫu |
| 1.2 Học nền tảng Go | A+B | Tour of Go + goroutines/channels + context. Viết 1 demo goroutine đọc WS thử | Hiểu concurrency, chạy được demo |
| 1.3 Khởi tạo repo + skeleton | A | `go mod init`, cấu trúc `cmd/` + `internal/`, `docker-compose.yml` (Timescale+Redis), Makefile | `docker-compose up` lên DB+Redis, build 4 binary rỗng OK |
| 1.4 Wireframe Dashboard + Bot | B | Vẽ mockup Dashboard (chart + bảng signal) + kịch bản lệnh Telegram Bot | File mockup + flow trong `docs/ux/` |

**Deliverable:** Repo skeleton + tài liệu khảo sát API + wireframe.

### Sprint 2 (Tuần 3-4) — Schema DB & Môi trường
**Mục tiêu:** Database sẵn sàng + CI/CD cơ bản.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 2.1 Viết migrations | A | 8 file theo `database_design.md` phần 11 (hypertable, continuous aggregate, retention) | `migrate up` tạo đủ bảng, hypertable hoạt động |
| 2.2 Setup sqlc + pgx | A | Cấu hình `sqlc.yaml`, viết query mẫu, generate code Go type-safe vào `internal/storage/timescale` | `sqlc generate` ra code, query test pass |
| 2.3 Redis wrapper + queue interface | A | `internal/cache` (go-redis) + `internal/queue` interface Publisher/Subscriber (asynq impl) | Test publish/consume 1 message OK |
| 2.4 CI/CD cơ bản | A | GitHub Actions: lint + test + build. Dockerfile multi-stage | Push → CI xanh, image build OK |
| 2.5 Hoàn thiện UX flow | B | Chốt luồng UI/UX Dashboard + cấu trúc lệnh Bot | Tài liệu UX final |

**Deliverable:** DB migrate được + sqlc + queue + CI/CD xanh. → **Sẵn sàng code core.**

---

## GIAI ĐOẠN 2 — CORE SYSTEM & DATA PIPELINE (Tháng 2-3)

### Sprint 3 (Tuần 5-6) — Ingestor WebSocket (T2.1)
**Mục tiêu:** Hứng được data realtime từ Binance, ghi vào DB.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 3.1 Exchange interface | A | `internal/exchange`: interface `Connector` (Subscribe, ReadLoop) + model chuẩn hóa | Interface rõ ràng, có mock test |
| 3.2 Binance connector | A | coder/websocket: stream trades + tickers + klines. Reconnect + backoff. Chuẩn hóa về model | Nhận tick realtime, tự reconnect khi rớt |
| 3.3 cmd/ingestor | A | Wire connector → publish vào queue → consumer batch insert (`pgx.CopyFrom`) vào `trades` | Data Binance chảy vào DB liên tục, không mất kết nối |
| 3.4 Lib thuật toán TA | B | Khảo sát `cinar/indicator`, viết wrapper RSI/EMA/MACD trong `internal/indicator` | Hàm tính RSI/EMA chạy đúng trên data mẫu |

**Deliverable:** Binance realtime → DB. **Cột mốc quan trọng nhất GĐ2.**

### Sprint 4 (Tuần 7-8) — Queue chịu tải + Partitioning (T2.2, T2.3)
**Mục tiêu:** Pipeline chịu tải khi thị trường biến động.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 4.1 asynq tải nặng | A | Cấu hình worker pool, retry, dead-letter. Test với burst tick (mô phỏng tải cao) | Không drop message khi tải cao, không tràn RAM |
| 4.2 Bybit + OKX connector | A | Thêm 2 connector (Bybit: liquidation; OKX: orderbook) vào cùng interface | 3 sàn cùng chảy vào DB |
| 4.3 Partitioning/chunk tuning | A | Tinh chỉnh `chunk_time_interval`, index, kiểm tra plan query | Query 1 symbol/1 ngày < 50ms |
| 4.4 Rolling window engine | B | `internal/indicator/rolling.go`: ring buffer O(1) cho trung bình động | Test rolling mean/std đúng, O(1) |

**Deliverable:** 3 sàn realtime + pipeline chịu tải.

### Sprint 5 (Tuần 9-10) — Aggregation (T2.4)
**Mục tiêu:** Gộp nến tự động + quản lý ổ cứng.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 5.1 Continuous aggregates | A | Bật `klines_1m/5m/1h` + refresh policy | Nến tự rollup, query nến nhanh |
| 5.2 cmd/aggregator | A | Cronjob bổ sung (gộp funding/OI theo giờ, dọn dẹp) | Chạy định kỳ ổn định |
| 5.3 Nén + retention | A | Bật compression + retention policy theo `database_design.md` phần 8 | Dung lượng raw được kiểm soát, nén hoạt động |
| 5.4 Volume/RSI detector | B | Detector Volume Spike + RSI extreme dùng rolling window | Sinh signal đúng trên data thật |

**Deliverable:** Pipeline đầy đủ: ingest → aggregate → nén. Ổ cứng được kiểm soát.

### Sprint 6 (Tuần 11-12) — Thuật toán cốt lõi (T2.5)
**Mục tiêu:** Các detector tạo giá trị — phần đáng tiền nhất.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 6.1 Detector interface + Rule Engine | B | `internal/indicator/detector.go` + `internal/alert` (so ngưỡng, cooldown) | Kiến trúc 3 lớp Detect→Rule→Dispatch chạy |
| 6.2 Funding Rate detector | B | So funding hiện tại vs trung bình 7-30 ngày → signal khi vượt ngưỡng | Phát hiện funding bất thường chính xác |
| 6.3 OI Delta + Liquidation Spike | B | OI delta kết hợp hướng giá; đếm liquidation theo cửa sổ | 2 detector sinh signal đúng |
| 6.4 Lưu signals + test backtest | A+B | Ghi `signals` vào DB, viết script backtest trên data lịch sử | Signal lưu DB, backtest cho kết quả hợp lý |

**Deliverable:** Bộ detector phái sinh (Funding/OI/Liquidation) + TA cơ bản chạy thật. → **Lõi giá trị hoàn thành.**

---

## GIAI ĐOẠN 3 — PHÁT HÀNH MVP (Tháng 4)

### Sprint 7 (Tuần 13-14) — REST API (T3.1)
**Mục tiêu:** API trả kết quả phân tích cho Frontend/Bot.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 7.1 cmd/api (Gin) | A | Router, middleware (log, recover, CORS, rate-limit), health check | Server chạy, `/health` OK |
| 7.2 Endpoints dữ liệu | A | `GET /klines`, `/signals`, `/funding`, `/symbols` — đọc qua cache Redis | Trả data đúng, có cache |
| 7.3 Auth bằng token | A | API key / JWT, gắn user, phân quyền free/premium | Endpoint bảo mật, phân quyền hoạt động |
| 7.4 WS server realtime | A | WebSocket push signal mới + giá realtime cho Dashboard | Client nhận push realtime |

### Sprint 8 (Tuần 15-16) — Bot + Dashboard + Deploy (T3.2, T3.3, T3.4)
**Mục tiêu:** Sản phẩm hoàn chỉnh, deploy, alpha test.

| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 8.1 cmd/bot Telegram | B | Bot gửi signal tự động + lệnh `/price /watch /alerts` | Bot gửi cảnh báo realtime |
| 8.2 Dashboard Next.js | B | Chart Lightweight-Charts realtime + bảng signal, kết nối WS | Dashboard hiển thị chart + signal live |
| 8.3 Deploy VPS | A | docker-compose lên Hetzner/DigitalOcean, domain, HTTPS, monitoring cơ bản | Hệ thống chạy 24/7 trên VPS |
| 8.4 Alpha test | A+B | Mời 10-20 trader test, thu feedback | Có báo cáo feedback + bug list |

**Deliverable:** 🚀 **MVP LIVE** — API + Bot + Dashboard chạy trên VPS thật.

---

## GIAI ĐOẠN 4 — TỐI ƯU, THƯƠNG MẠI HÓA & MARKETING (Tháng 5-6)

### Sprint 9 (Tuần 17-18) — Tối ưu & Ổn định (T4.1)
| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 9.1 Profiling & fix memory leak | A | `pprof` tìm rò rỉ goroutine/heap, fix | Chạy 72h không tăng RAM bất thường |
| 9.2 Tối ưu hiệu năng | A | Tối ưu query chậm, batch size, pool config | Latency API giảm, throughput tăng |
| 9.3 Whale Transfer detector | B | Lắng nghe event Transfer ETH/BSC, map `known_wallets`, signal whale | Phát hiện whale > ngưỡng USD |
| 9.4 Hardening | A | Graceful shutdown, alert khi service chết, backup DB | Hệ thống tự phục hồi, có backup |

### Sprint 10 (Tuần 19-20) — Custom Alerts (T4.2)
| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 10.1 CRUD alert_rules | A | API tạo/sửa/xóa rule, validate | User quản lý rule qua API/Bot |
| 10.2 Rule matching engine | B | Khớp rule user với signal realtime + cooldown chống spam | Alert đúng điều kiện user đặt |
| 10.3 UI quản lý alert | B | Màn hình tạo alert trên Dashboard + Bot | User tự đặt alert dễ dàng |

### Sprint 11 (Tuần 21-22) — Thương mại hóa (T4.3)
| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 11.1 Cổng thanh toán crypto | A | Nhận USDT BSC/TRON, verify tx_hash on-chain, kích hoạt subscription | Thanh toán → tự nâng gói Premium |
| 11.2 Phân tầng tính năng | A | Giới hạn free vs premium (số alert, độ trễ data, symbol) | Free/Premium khác biệt rõ |
| 11.3 Quản lý subscription | B | Màn hình gói + lịch sử thanh toán + nhắc gia hạn | User xem/mua/gia hạn được |

### Sprint 12 (Tuần 23-24) — Marketing & Hoàn thiện (T4.4)
| Task | Người | Chi tiết | DoD |
|---|---|---|---|
| 12.1 Auto-share insight | B | Bot tự đăng insight onchain/funding lên FB/X | Nội dung tự đăng định kỳ |
| 12.2 Landing page | B | Trang giới thiệu + đăng ký | Landing live, có CTA |
| 12.3 Onboarding & polish | A+B | Tài liệu user, fix bug còn lại, tối ưu cuối | Sản phẩm sẵn sàng public |

**Deliverable:** 💰 **Sản phẩm thương mại** — có doanh thu (Premium) + kênh marketing.

---

## Bảng Tổng Hợp Cột Mốc

| Sprint | Tuần | Cột mốc | Trạng thái |
|---|---|---|---|
| 1-2 | 1-4 | Repo + DB + CI/CD sẵn sàng | ⬜ |
| 3 | 5-6 | **Binance realtime → DB** | ⬜ |
| 4-5 | 7-10 | 3 sàn + pipeline chịu tải + aggregation | ⬜ |
| 6 | 11-12 | **Bộ detector cốt lõi hoàn thành** | ⬜ |
| 7-8 | 13-16 | 🚀 **MVP LIVE trên VPS** | ⬜ |
| 9-10 | 17-20 | Ổn định 24/7 + Custom Alerts | ⬜ |
| 11-12 | 21-24 | 💰 **Thương mại hóa + Marketing** | ⬜ |

---

## Rủi Ro & Giảm Thiểu

| Rủi ro | Mức độ | Giảm thiểu |
|---|---|---|
| Học Go chậm hơn dự kiến | Cao | Vừa học vừa làm task thật; Sprint 1-2 buffer cho học |
| Rate limit / ban IP từ sàn | TB | Tôn trọng limit, dùng WS thay REST polling, có backoff |
| Tải WebSocket vượt VPS 50$ | TB | Queue + batch insert + nén; scale ingestor riêng nếu cần |
| Map ví whale khó (data label) | TB | Bootstrap từ Etherscan labels; làm sau ở GĐ4 |
| Memory leak khi chạy 24/7 | Cao | pprof từ sớm, graceful shutdown, Sprint 9 dành riêng |
| 2 người quá tải | Cao | Ưu tiên nghiệt ngã: pipeline + detector trước, đẹp sau |

---

## Nguyên Tắc Ưu Tiên (khi thiếu thời gian)

1. **Pipeline dữ liệu chạy thật** > mọi thứ khác.
2. **Detector phái sinh** (Funding/OI/Liquidation) > TA > Whale > Social.
3. **Bot Telegram** (kênh tiếp cận nhanh) > Dashboard đẹp.
4. **Ổn định 24/7** > thêm tính năng mới.
5. **Cắt scope, không cắt chất lượng lõi.**
