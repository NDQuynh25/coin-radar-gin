# Bộ Test API Các Sàn (T1.1)

Kiểm chứng toàn bộ endpoint REST + WebSocket của Binance, Bybit, OKX dùng cho Crypto Data Platform.
Tất cả là **public market data — không cần API key**.

## Yêu cầu
- Node.js >= 21 (có `WebSocket` và `fetch` built-in toàn cục). Đã test trên Node v24.

## Cách chạy

```bash
cd tools/exchange-api-test

node test-all.js            # test tất cả sàn (REST + WebSocket)
node test-all.js binance    # chỉ test 1 sàn (binance | bybit | okx)
node test-all.js --rest     # bỏ qua WebSocket, chỉ test REST
```

Exit code: `0` nếu tất cả pass, `1` nếu có lỗi (tiện cho CI/CD sau này).

## Cấu trúc

| File | Vai trò |
|---|---|
| `endpoints.js` | Định nghĩa mọi endpoint REST + cấu hình WebSocket mỗi sàn, kèm hàm `check()` xác thực data |
| `test-all.js` | Runner: gọi từng endpoint, đo thời gian, in pass/fail + tóm tắt data |
| `README.md` | Tài liệu này |

## Endpoint được test

| Sàn | REST | WebSocket |
|---|---|---|
| **Binance** | spot price, 24h ticker, klines, trades, depth, **funding+mark**, **open interest**, funding history, OI history, long/short ratio | aggTrade stream |
| **Bybit** | linear ticker (gộp giá+funding+OI), kline, trades, OI history, funding history, orderbook | publicTrade stream |
| **OKX** | swap ticker, **funding rate**, **open interest**, candles, trades, funding history | trades channel |

## Ghi chú
- Test chạy **tuần tự** để tôn trọng rate limit của sàn.
- Các endpoint phái sinh (funding, OI, liquidation) là **phần giá trị cốt lõi** của sản phẩm — được ưu tiên test kỹ.
- Đây là tool validation/research (Node). Code ingest thật trong sản phẩm sẽ viết bằng **Go** (`internal/exchange/`).
