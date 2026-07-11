# Kiến Trúc Hệ Thống — Coin Radar (Clean Architecture - DDD)

Tài liệu này đặc tả kiến trúc mã nguồn của Coin Radar. Hệ thống được triển khai dưới dạng **Modular Monolith + Multi-binary** để bảo đảm tính đơn giản trong vận hành nhưng hiệu quả trong hiệu năng và chia nhỏ tài nguyên.

---

## 1. Thiết Kế Các Lớp (Clean Architecture Layers)

Mã nguồn nghiệp vụ cốt lõi nằm trong thư mục `internal/` và tuân thủ mô hình 4 lớp Clean Architecture:

```
┌──────────────────────────────────────────────────────────┐
│                     4. Interfaces                        │ (HTTP, WebSockets, Telegram Bot)
│   ┌──────────────────────────────────────────────────┐   │
│   │                 3. Infrastructure                │   │ (DB, Redis Client, Queue wrapper)
│   │   ┌──────────────────────────────────────────┐   │   │
│   │   │             2. Application               │   │   │ (Use cases / Business services)
│   │   │   ┌──────────────────────────────────┐   │   │   │
│   │   │   │           1. Domain              │   │   │   │ (Core business rules: Entity & Interface)
│   │   │   └──────────────────────────────────┘   │   │   │
│   │   └──────────────────────────────────────────┘   │   │
│   └──────────────────────────────────────────────────┘   │
└──────────────────────────────────────────────────────────┘
```

### Lớp 1: Domain (`internal/domain/`)
* **Nhiệm vụ:** Chứa các thực thể (Entities), cấu trúc dữ liệu dùng chung (Models) và định nghĩa các giao tiếp (Interfaces) cho tầng lưu trữ.
* **Quy tắc:** Lớp độc lập cao nhất, **không import** bất kỳ package nào từ các lớp ngoài.
* **Thành phần:**
  * `internal/domain/model/`: Định nghĩa các cấu trúc dữ liệu như `Trade`, `Kline`, `Signal`, `AlertRule`, `User`.
  * `internal/domain/repository/`: Định nghĩa các interface truy xuất DB như `UserRepository`, `SignalRepository`.
  * `internal/domain/cache/`: Định nghĩa các interface cache như `PriceCache`.

### Lớp 2: Application (`internal/application/`)
* **Nhiệm vụ:** Chứa logic nghiệp vụ cốt lõi (Use Cases). Điều phối dữ liệu từ Domain, gọi các Repository/Cache để hoàn thành luồng công việc.
* **Quy tắc:** Chỉ phụ thuộc vào các interface định nghĩa ở lớp Domain. Không phụ thuộc vào MySQL, PostgreSQL hay Redis cụ thể.
* **Thành phần:**
  * `indicator_service.go`: Xử lý thuật toán nến, RSI, EMA, Rolling Window.
  * `alert_service.go`: So khớp cảnh báo của người dùng, thực thi Rule Engine.
  * `ingest_service.go`: Gom luồng dữ liệu chuẩn hóa đẩy đi xử lý.

### Lớp 3: Infrastructure (`internal/infrastructure/`)
* **Nhiệm vụ:** Triển khai kỹ thuật cụ thể cho các interface ở lớp Domain.
* **Quy tắc:** Phụ thuộc vào thư viện bên ngoài (pgx, redis, asynq) để thực hiện kết nối vật lý.
* **Thành phần:**
  * `internal/infrastructure/storage/`: Thiết lập connection pool (`timescale/`, `redis/`).
  * `internal/infrastructure/repository/`: Hiện thực các SQL query cụ thể (sử dụng code sinh ra bởi sqlc).
  * `internal/infrastructure/cache/`: Hiện thực cache cụ thể bằng Redis.
  * `internal/infrastructure/queue/`: Hiện thực việc gửi & nhận job bất đồng bộ qua Redis asynq.

### Lớp 4: Interfaces / Delivery (`internal/interfaces/`)
* **Nhiệm vụ:** Tiếp xúc với thế giới bên ngoài. Tiếp nhận request, kiểm tra định dạng và gọi Service ở lớp Application xử lý, sau đó định dạng lại response.
* **Thành phần:**
  * `internal/interfaces/http/`: REST API Router và Handlers viết bằng Gin.
  * `internal/interfaces/ws/`: Kết nối WebSocket phục vụ đẩy dữ liệu live cho Dashboard.
  * `internal/interfaces/telegram/`: Xử lý các lệnh chat bot Telegram.

---

## 2. Mô hình Đa Binary (Multi-binary Entrypoints)

Tất cả các dịch vụ độc lập đều nằm trong thư mục `cmd/`. Khi biên dịch, mỗi thư mục sẽ tạo ra một file thực thi riêng:

1. **`cmd/ingestor`**:
   * **Mục tiêu:** Mở kết nối WebSocket tới các sàn (Binance, Bybit...), chuẩn hóa payload nhận được thành cấu trúc `model.Trade` và đẩy vào Queue (asynq) hoặc Batch Insert vào DB.
2. **`cmd/api`**:
   * **Mục tiêu:** REST API phục vụ UI đọc dữ liệu lịch sử nến, quản lý rule alert của user. Đồng thời chạy WebSocket server để push tín hiệu.
3. **`cmd/bot`**:
   * **Mục tiêu:** Chạy listener nhận tin nhắn từ Telegram. Gửi tin nhắn thông báo khi Rule Engine phát hiện tín hiệu.
4. **`cmd/aggregator`**:
   * **Mục tiêu:** Cronjob chạy định kỳ để dọn dẹp data cũ, ép nén dung lượng TimescaleDB.

---

## 3. Quy tắc Vàng khi Mở Rộng
1. **Interface First:** Trước khi kết nối với bất kỳ dịch vụ hay viết truy vấn nào, hãy khai báo Interface trong `internal/domain/`.
2. **Không Viết Trực Tiếp SQL Trong Controller:** Mọi thao tác ghi/đọc dữ liệu đều phải đi qua Service $\rightarrow$ Repository.
3. **Log & Graceful Shutdown:** Mọi tiến trình chạy nền trong `cmd/` đều cần cài đặt bắt tín hiệu ngắt OS (`SIGINT`, `SIGTERM`) để đóng kết nối DB, Redis an toàn trước khi dừng.
