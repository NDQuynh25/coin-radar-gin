# Coin Radar API

Backend độc lập cho Coin Radar, gồm REST API và các worker Go.

## Chạy local

1. Cài Go 1.26 trở lên và Docker Desktop.
2. Khởi động các dịch vụ phụ trợ:

   ```bash
   docker compose up -d
   ```

3. Generate Prisma client:

   ```bash
   go run github.com/steebchen/prisma-client-go generate
   ```

4. Cập nhật `config.yaml` (đặc biệt là `auth.jwt_secret` và Telegram nếu dùng), sau đó chạy API:

   ```bash
   go run ./cmd/api
   ```

API mặc định chạy tại `http://localhost:9000`, với endpoint health là
`GET /health` và API là `/api/v1`.

## Các lệnh hữu ích

```bash
go test ./...
make build
make run-ingestor
make run-aggregator
make run-bot
```

## Kết nối frontend

Frontend là một dự án riêng tại `C:\Users\ndquynh\Documents\coin-radar-next`.
Cấu hình biến `API_URL` và
`NEXT_PUBLIC_API_URL` của frontend thành URL của API này, ví dụ:
`http://localhost:9000/api/v1`.
