# Từ Điển Thuật Ngữ Crypto — Tài Liệu Tra Cứu

> Tài liệu nền tảng cho team hiểu domain trước/trong khi code Crypto Data Platform.
> Mỗi thuật ngữ kèm: định nghĩa dễ hiểu + ví dụ + liên hệ với dự án.

---

## 0. Sản Phẩm Của Chúng Ta Làm Gì?

Xây **"radar cảnh báo sớm"** cho dân giao dịch crypto.

```
Hứng dữ liệu thô (trade, funding, OI, whale...) từ các sàn
        ↓
Tính toán → phát hiện điều bất thường
   ("Funding tăng vọt!", "Cá mập vừa chuyển 5 triệu $!", "Hàng loạt lệnh bị thanh lý!")
        ↓
Gửi cảnh báo qua Telegram + hiện trên Dashboard
        ↓
Trader đọc → ra quyết định mua/bán tốt hơn → trả tiền cho mình (Premium)
```

Giá trị nằm ở chỗ **phát hiện tín hiệu nhanh và chính xác** trước khi giá biến động.
Chúng ta **không giao dịch hộ** — chỉ cung cấp dữ liệu & tín hiệu.

---

## 1. Khái Niệm Nền Tảng

### Crypto (Tiền mã hóa)
Tiền số như Bitcoin (BTC), Ethereum (ETH)... Mua đi bán lại để kiếm lời, giống cổ phiếu/vàng nhưng giao dịch **24/7** (không nghỉ cuối tuần).

### Spot (Giao dịch giao ngay)
Mua/bán **coin thật** — bạn thực sự sở hữu coin sau khi mua. Đơn giản nhất.

### Derivatives (Phái sinh)
Cá cược giá lên/xuống mà **không cần sở hữu coin thật**. Phức tạp hơn nhưng là nơi sinh ra nhiều tín hiệu giá trị (funding, OI, liquidation).

### Symbol / Cặp giao dịch
Ký hiệu cặp tiền. VD: `BTCUSDT` = giá Bitcoin tính theo USDT. `BTC` là base asset, `USDT` là quote asset.

---

## 2. Các Loại Sàn / Nguồn Dữ Liệu

| Loại | Là gì | Ví dụ |
|---|---|---|
| **CEX** (Centralized Exchange) | Sàn tập trung, có công ty quản lý — như "sàn chứng khoán" của crypto | **Binance, Bybit, OKX** |
| **DEX** (Decentralized Exchange) | Sàn phi tập trung, giao dịch trực tiếp trên blockchain | Uniswap, PancakeSwap |
| **On-chain** | Dữ liệu ghi trực tiếp trên blockchain (ai chuyển tiền cho ai) | Ethereum, BSC |

---

## 3. Dữ Liệu Thị Trường

### Trade (Giao dịch)
Một lần mua/bán xảy ra. VD: "Ai đó vừa mua 0.5 BTC giá 60,000$". Mỗi giây có hàng nghìn trade → **dữ liệu thô realtime** hứng qua WebSocket. Lưu ở bảng `trades`.

### Kline / Candle (Nến)
Tóm tắt giá trong 1 khoảng thời gian. Một cây nến cho biết 4 giá — gọi là **OHLC**:
- **O**pen (mở cửa) · **H**igh (cao nhất) · **L**ow (thấp nhất) · **C**lose (đóng cửa)

Ghép nhiều nến → biểu đồ giá (như trên TradingView). Lưu ở bảng `klines`.

> **Aggregation (Gộp nến):** từ hàng nghìn trade lẻ → nén thành 1 cây nến 1m/5m/1h cho gọn. Dùng continuous aggregate của TimescaleDB.

### Orderbook (Sổ lệnh)
Danh sách tất cả lệnh **mua đang chờ (bid)** và **bán đang chờ (ask)** ở từng mức giá. Cho biết áp lực mua/bán. Lưu ở bảng `orderbook_snapshots`.

### Volume (Khối lượng)
Tổng số lượng giao dịch trong 1 khoảng thời gian. **Volume Spike** (khối lượng tăng đột biến) thường báo hiệu biến động sắp tới → là 1 detector của dự án.

---

## 4. Giao Dịch Phái Sinh (Derivatives) — Nơi Có Tín Hiệu Giá Trị Nhất

### Long / Short
- **Long** = cược giá **lên** (mua, kỳ vọng tăng) — phe "bò" 🐂
- **Short** = cược giá **xuống** (bán khống, kỳ vọng giảm) — phe "gấu" 🐻

### Leverage (Đòn bẩy)
Vay tiền của sàn để cược lớn hơn vốn thật. VD đòn bẩy 10x: có 1,000$ nhưng cược như 10,000$. Lời to nhưng **lỗ cũng to** → dễ bị thanh lý.

### Perpetual Futures (Hợp đồng vĩnh cửu)
Hợp đồng phái sinh phổ biến nhất, **không có ngày hết hạn** (giữ bao lâu cũng được). Để giá hợp đồng bám sát giá Spot, sinh ra cơ chế **Funding Rate**.

### ⭐ Funding Rate (Phí tài trợ) — TÍN HIỆU QUAN TRỌNG NHẤT
Phí định kỳ (**mỗi 8 giờ**) giữa phe Long và Short, dùng để cân bằng giá hợp đồng với giá Spot.

| Tình huống | Funding | Ai trả ai? | Ý nghĩa |
|---|---|---|---|
| Long đông (giá HĐ > Spot) | **Dương (+)** | Long trả cho Short | Đám đông tham lam → dễ đảo chiều **giảm** |
| Short đông (giá HĐ < Spot) | **Âm (−)** | Short trả cho Long | Đám đông sợ hãi → dễ đảo chiều **tăng** |

**Ví dụ:** Long 10,000$, funding +0.01% → trả 1$/8h (bình thường).
Nhưng funding vọt +0.3% → trả 30$/8h = 90$/ngày = ~27% vốn/tháng (rất tốn) → dấu hiệu hưng phấn quá mức.

> **Quy tắc vàng:** Funding cao bất thường = đám đông dồn 1 phía = thị trường dễ **ĐẢO CHIỀU**. Trader chuyên nghiệp đi ngược đám đông (contrarian). Detector của mình bắt và cảnh báo tự động.

### Open Interest (OI — Số hợp đồng mở)
Tổng số hợp đồng phái sinh đang mở (chưa đóng). Kết hợp OI + hướng giá:

| OI | Giá | Ý nghĩa |
|---|---|---|
| Tăng | Tăng | Tiền **mới** vào, xu hướng tăng **khỏe** |
| Tăng | Giảm | Mở Short mạnh, xu hướng giảm khỏe |
| Giảm | — | Tiền rút ra, xu hướng yếu dần |

→ Detector **OI Delta** theo dõi thay đổi OI.

### 💀 Liquidation (Thanh lý)
Khi trader chơi đòn bẩy mà giá đi **ngược** quá nhiều → sàn tự động đóng lệnh, họ **mất sạch tiền cược**. Lưu ở bảng `liquidations`.

- **Long Squeeze:** quá nhiều Long bị thanh lý cùng lúc → giá giảm mạnh thêm.
- **Short Squeeze:** quá nhiều Short bị thanh lý → giá bật tăng mạnh.

> **Liquidation Spike** (thanh lý hàng loạt trong thời gian ngắn) = biến động mạnh, thường báo đảo chiều ngắn hạn → là detector giá trị cao của dự án.

### Long/Short Ratio (Tỷ lệ Long/Short)
Tỷ lệ giữa số người Long và Short. Khi đám đông lệch hẳn 1 phía → tín hiệu **ngược** (contrarian).

### Mark Price / Index Price
- **Index Price:** giá tham chiếu trung bình từ nhiều sàn Spot.
- **Mark Price:** giá dùng để tính lãi/lỗ và thanh lý (tránh bị thao túng giá tức thời).

---

## 5. Dữ Liệu On-Chain

### 🐋 Whale (Cá mập)
Ví có **rất nhiều tiền**. Khi cá mập chuyển lượng lớn coin (vd > 1 triệu $) → có thể sắp có biến động lớn.

### Whale Tracking (Theo dõi cá mập)
Lắng nghe sự kiện **Transfer** trên blockchain, lọc giao dịch giá trị lớn → cảnh báo. Lưu ở bảng `whale_transfers`, dùng bảng `known_wallets` để gán nhãn ví.

### Exchange Inflow / Outflow (Dòng tiền vào/ra sàn)
- **Inflow** (tiền chảy **vào** ví sàn) = có thể chuẩn bị **bán** → áp lực giảm.
- **Outflow** (tiền chảy **ra** khỏi sàn) = tích lũy, cất giữ → tín hiệu tích cực.

### Smart Money (Tiền thông minh)
Các ví của nhà đầu tư giỏi/tổ chức đã được biết đến. Theo dõi động thái của riêng họ.

### Các thuật ngữ on-chain khác
- **RPC Node:** điểm kết nối để đọc dữ liệu blockchain (JSON-RPC / Web3).
- **Token Transfer:** sự kiện chuyển token giữa các ví.
- **Meme coin:** token mới, đầu cơ cao, rủi ro lớn (theo dõi qua GeckoTerminal/DexScreener).

---

## 6. Chỉ Báo Kỹ Thuật (Technical Analysis — TA)

Tính trên dữ liệu nến. Dùng thư viện Go `cinar/indicator` (không tự viết).

| Chỉ báo | Là gì |
|---|---|
| **EMA / SMA** | Đường trung bình động (giá trung bình N nến). Crossover (cắt nhau) báo xu hướng |
| **RSI** | Đo "quá mua/quá bán" (0-100). > 70 = quá mua, < 30 = quá bán |
| **MACD** | Đo động lượng xu hướng |
| **Bollinger Bands** | Dải biến động quanh giá trung bình |
| **VWAP** | Giá trung bình theo khối lượng |

---

## 7. Social Data

### Social Listening
Theo dõi tần suất một coin được nhắc đến trên mạng xã hội (X/Twitter), đặc biệt từ **KOL** (người có ảnh hưởng). Tăng đột biến = sự chú ý tăng → có thể biến động.

### Sentiment (Tâm lý)
Phân tích cảm xúc tích cực/tiêu cực của cộng đồng. (Nâng cao, làm sau — P2.)

---

## 8. Hệ Thống Tín Hiệu Của Dự Án

### Signal (Tín hiệu)
Kết quả khi detector phát hiện điều bất thường. Lưu ở bảng `signals` với:
- **type:** loại (funding_spike / oi_delta / liquidation_spike / volume_spike / whale_alert...)
- **severity:** mức độ (info / warning / critical)

### Detector
Module phát hiện 1 loại tín hiệu. Mỗi loại implement interface `Detect(tick) *Signal`.

### Alert / Custom Alert (Cảnh báo tùy chỉnh)
Người dùng tự đặt điều kiện nhận cảnh báo (vd "báo tôi khi funding BTC > 0.1%"). Lưu ở bảng `alert_rules`. Có **cooldown** chống spam.

### Rolling Window (Cửa sổ trượt)
Kỹ thuật tính trung bình động trong RAM bằng **ring buffer**, cập nhật O(1) mỗi tick — không cần query DB mỗi lần.

---

## 9. Bảng Tra Nhanh — Tín Hiệu & Ý Nghĩa

| Tín hiệu | Khi nào kích hoạt | Trader hiểu là |
|---|---|---|
| **Funding Spike** | Funding cao/thấp bất thường | Đám đông lệch 1 phía → dễ đảo chiều |
| **OI Delta** | OI thay đổi mạnh | Tiền vào/ra thị trường, xu hướng khỏe/yếu |
| **Liquidation Spike** | Thanh lý hàng loạt | Biến động mạnh, đảo chiều ngắn hạn |
| **Volume Spike** | Khối lượng tăng vọt | Có động thái lớn sắp xảy ra |
| **Whale Alert** | Cá mập chuyển tiền lớn | Có thể sắp biến động |
| **RSI Extreme** | RSI > 70 hoặc < 30 | Quá mua / quá bán |

---

## 10. Độ Ưu Tiên Triển Khai (theo giá trị / công sức)

| # | Nhóm | Tín hiệu | Độ khó |
|---|---|---|---|
| 🥇 | Phái sinh | Funding Rate, OI Delta, Liquidation Spike, Long/Short Ratio | Dễ, data sẵn |
| 🥈 | Kỹ thuật | Volume Spike, RSI/EMA/MACD | Dễ (dùng lib) |
| 🥉 | On-chain | Whale Transfer, Inflow/Outflow | Khó vừa (cần map ví) |
| 4 | Social | Mention spike | Làm sau (P2) |

---

## Liên Kết Tài Liệu

- [desgin.md](desgin.md) — Kiến trúc hệ thống tổng thể
- [database_design.md](database_design.md) — Thiết kế cơ sở dữ liệu
- [development_plan.md](development_plan.md) — Kế hoạch phát triển 6 tháng
