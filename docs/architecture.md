# Kiến trúc Coin Radar

Coin Radar dùng **modular monolith với nhiều binary**. Dự án chỉ áp dụng các nguyên tắc cần thiết của Clean Architecture, không bắt buộc mọi chức năng phải có đủ nhiều lớp hoặc interface.

## Mục tiêu

- Dễ tìm và sửa code theo nghiệp vụ.
- Nghiệp vụ không phụ thuộc trực tiếp vào HTTP hoặc database.
- API, bot, ingestor và aggregator có thể dùng chung logic.
- Tránh abstraction và interface khi chưa có nhu cầu thực tế.

## Cấu trúc hiện tại

```text
cmd/
  api/          composition root và HTTP server
  bot/          Telegram worker
  ingestor/     market-data ingestion worker
  aggregator/   aggregation worker

internal/
  user/         model, service, repository, auth và HTTP handler
  market/       model, repository, cache contract và HTTP handler
  signal/       model, repository, cache contract và HTTP handler
  alert/        model và repository contract
  platform/
    web/          Gin bootstrap, health, request và response dùng chung
    postgres/
      prisma/     relational client do Prisma generate
    timescale/    pgx connection và sqlc queries
  config/       cấu hình ứng dụng
```

Mỗi package nghiệp vụ chứa các thành phần liên quan đến chính feature đó. Chỉ tách package con khi feature thực sự lớn hoặc xuất hiện dependency cycle.

## Luồng phụ thuộc

Luồng xử lý thông thường:

```text
HTTP/worker -> service -> repository interface -> repository implementation
```

Quy tắc:

1. Handler chỉ đọc request, gọi service và tạo response.
2. Service chứa logic nghiệp vụ và phụ thuộc vào contract cần thiết.
3. Repository chịu trách nhiệm lưu trữ và truy vấn dữ liệu.
4. Dependency được khởi tạo trong `cmd/<binary>/main.go`, không khởi tạo trong router hoặc handler.
5. Chỉ tạo interface khi có boundary cần test/thay thế hoặc có nhiều implementation.

## Phân chia database

- Prisma: user, subscription, alert rule và dữ liệu relational.
- pgx/sqlc + TimescaleDB: trade, candle, tick và dữ liệu time-series khối lượng lớn.
- Redis: cache, stream hoặc queue khi các luồng xử lý thực sự cần đến.

Service không được phụ thuộc trực tiếp vào generated Prisma/sqlc code. Generated client được bọc bởi repository implementation.

## Generated code

Sau khi clone hoặc thay đổi `prisma/schema.prisma`, chạy:

```powershell
go run github.com/steebchen/prisma-client-go generate
```

Generated Prisma files không được commit. CI phải generate chúng trước khi build hoặc test.

## Nguyên tắc mở rộng

- Ưu tiên code đơn giản hơn kiến trúc “đúng sách”.
- Không thêm layer chỉ để chuyển tiếp một lời gọi.
- Không truy vấn database trực tiếp từ handler.
- Mỗi binary phải hỗ trợ graceful shutdown cho tài nguyên mà nó mở.
- Một feature nhỏ có thể nằm trong một package; chỉ tách nhỏ khi package thực sự khó quản lý.
