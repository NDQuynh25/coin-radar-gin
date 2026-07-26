package models

import "time"

// Các model bên dưới ánh xạ tới hypertable của TimescaleDB để truy vấn bằng GORM.
// Việc tạo hypertable, continuous aggregate, chính sách lưu trữ và ghi dữ liệu
// vẫn do SQL migration hoặc pgx đảm nhiệm.

// Trade biểu diễn một giao dịch đã được khớp trên sàn.
type Trade struct {
	ID        string     `gorm:"column:id;type:uuid;primaryKey"` // UUID định danh duy nhất của bản ghi.
	Time      time.Time  `gorm:"column:time"`                    // Thời điểm giao dịch được khớp.
	Exchange  string     `gorm:"column:exchange"`                // Sàn giao dịch, ví dụ: Binance.
	TradeID   string     `gorm:"column:trade_id"`                // Mã giao dịch do sàn cung cấp.
	Symbol    string     `gorm:"column:symbol"`                  // Cặp giao dịch, ví dụ: BTCUSDT.
	Price     float64    `gorm:"column:price"`                   // Giá khớp lệnh.
	Qty       float64    `gorm:"column:qty"`                     // Khối lượng tài sản được giao dịch.
	Side      string     `gorm:"column:side"`                    // Phía giao dịch: buy hoặc sell.
	CreatedAt time.Time  `gorm:"column:created_at"`              // Thời điểm tạo bản ghi.
	UpdatedAt time.Time  `gorm:"column:updated_at"`              // Thời điểm cập nhật bản ghi gần nhất.
	DeletedAt *time.Time `gorm:"column:deleted_at"`              // Thời điểm xóa mềm; nil nếu chưa bị xóa.
}

func (Trade) TableName() string { return "trades" }

// Kline biểu diễn một cây nến OHLCV trong một khung thời gian.
type Kline struct {
	Time       time.Time `gorm:"column:time;primaryKey"`     // Thời điểm bắt đầu cây nến.
	Exchange   string    `gorm:"column:exchange;primaryKey"` // Sàn cung cấp dữ liệu.
	Symbol     string    `gorm:"column:symbol;primaryKey"`   // Cặp giao dịch.
	Interval   string    `gorm:"column:interval;primaryKey"` // Khung thời gian, ví dụ: 1m, 1h, 1d.
	Open       float64   `gorm:"column:open"`                // Giá mở cửa.
	High       float64   `gorm:"column:high"`                // Giá cao nhất.
	Low        float64   `gorm:"column:low"`                 // Giá thấp nhất.
	Close      float64   `gorm:"column:close"`               // Giá đóng cửa.
	Volume     float64   `gorm:"column:volume"`              // Tổng khối lượng giao dịch.
	TradeCount *int      `gorm:"column:trade_count"`         // Số giao dịch trong cây nến; có thể không có dữ liệu.
}

func (Kline) TableName() string { return "klines" }

// Derivative chứa dữ liệu thị trường phái sinh tại một thời điểm.
type Derivative struct {
	Time            time.Time  `gorm:"column:time"`              // Thời điểm ghi nhận dữ liệu.
	Exchange        string     `gorm:"column:exchange"`          // Sàn giao dịch phái sinh.
	Symbol          string     `gorm:"column:symbol"`            // Mã hợp đồng hoặc cặp giao dịch.
	FundingRate     *float64   `gorm:"column:funding_rate"`      // Tỷ lệ funding hiện tại.
	NextFundingTime *time.Time `gorm:"column:next_funding_time"` // Thời điểm thanh toán funding tiếp theo.
	OpenInterest    *float64   `gorm:"column:open_interest"`     // Tổng khối lượng hoặc giá trị vị thế đang mở.
	MarkPrice       *float64   `gorm:"column:mark_price"`        // Giá đánh dấu dùng để tính PnL và thanh lý.
	IndexPrice      *float64   `gorm:"column:index_price"`       // Giá chỉ số tham chiếu từ thị trường cơ sở.
}

func (Derivative) TableName() string { return "derivatives" }

// Liquidation biểu diễn một sự kiện thanh lý vị thế.
type Liquidation struct {
	Time     time.Time `gorm:"column:time"`      // Thời điểm xảy ra thanh lý.
	Exchange string    `gorm:"column:exchange"`  // Sàn xảy ra sự kiện thanh lý.
	Symbol   string    `gorm:"column:symbol"`    // Cặp giao dịch hoặc hợp đồng bị thanh lý.
	Side     string    `gorm:"column:side"`      // Phía của lệnh thanh lý: buy hoặc sell.
	Price    float64   `gorm:"column:price"`     // Giá thực hiện thanh lý.
	Qty      float64   `gorm:"column:qty"`       // Khối lượng bị thanh lý.
	ValueUSD *float64  `gorm:"column:value_usd"` // Giá trị thanh lý quy đổi sang USD.
}

func (Liquidation) TableName() string { return "liquidations" }

// OrderbookSnapshot lưu trạng thái sổ lệnh tại một thời điểm.
type OrderbookSnapshot struct {
	Time     time.Time `gorm:"column:time"`            // Thời điểm chụp sổ lệnh.
	Exchange string    `gorm:"column:exchange"`        // Sàn cung cấp sổ lệnh.
	Symbol   string    `gorm:"column:symbol"`          // Cặp giao dịch.
	Bids     []byte    `gorm:"column:bids;type:jsonb"` // Danh sách lệnh mua, lưu dưới dạng JSONB.
	Asks     []byte    `gorm:"column:asks;type:jsonb"` // Danh sách lệnh bán, lưu dưới dạng JSONB.
}

func (OrderbookSnapshot) TableName() string { return "orderbook_snapshots" }

// WhaleTransfer biểu diễn một giao dịch on-chain có giá trị lớn.
type WhaleTransfer struct {
	Time         time.Time `gorm:"column:time"`                // Thời điểm giao dịch được ghi nhận trên blockchain.
	Chain        string    `gorm:"column:chain"`               // Mạng blockchain, ví dụ: Ethereum.
	TxHash       string    `gorm:"column:tx_hash"`             // Hash định danh giao dịch on-chain.
	TokenAddress *string   `gorm:"column:token_address"`       // Địa chỉ smart contract; nil đối với native coin.
	TokenSymbol  *string   `gorm:"column:token_symbol"`        // Ký hiệu token, ví dụ: USDT.
	FromAddress  string    `gorm:"column:from_address"`        // Địa chỉ ví gửi.
	ToAddress    string    `gorm:"column:to_address"`          // Địa chỉ ví nhận.
	Amount       *string   `gorm:"column:amount;type:numeric"` // Số lượng token; dùng string để tránh mất độ chính xác.
	ValueUSD     *float64  `gorm:"column:value_usd"`           // Giá trị giao dịch quy đổi sang USD.
	Direction    *string   `gorm:"column:direction"`           // Hướng dòng tiền, ví dụ: inflow hoặc outflow.
}

func (WhaleTransfer) TableName() string { return "whale_transfers" }

// Signal biểu diễn một tín hiệu cảnh báo được hệ thống phát hiện.
type Signal struct {
	Time      time.Time `gorm:"column:time;primaryKey"`    // Thời điểm tín hiệu được tạo.
	ID        int64     `gorm:"column:id;primaryKey"`      // ID tín hiệu; kết hợp với Time tạo khóa chính.
	Type      string    `gorm:"column:type"`               // Loại tín hiệu, ví dụ: price_spike.
	Exchange  *string   `gorm:"column:exchange"`           // Sàn liên quan; nil nếu không thuộc sàn cụ thể.
	Symbol    string    `gorm:"column:symbol"`             // Tài sản hoặc cặp giao dịch liên quan.
	Severity  string    `gorm:"column:severity"`           // Mức độ quan trọng: low, medium hoặc high.
	Value     *float64  `gorm:"column:value"`              // Giá trị thực tế làm phát sinh tín hiệu.
	Threshold *float64  `gorm:"column:threshold"`          // Ngưỡng dùng để kích hoạt tín hiệu.
	Payload   []byte    `gorm:"column:payload;type:jsonb"` // Dữ liệu chi tiết bổ sung, lưu dưới dạng JSONB.
}

func (Signal) TableName() string { return "signals" }
