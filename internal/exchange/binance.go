package exchange

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/modules/market"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type binance struct{ cfg config.ExchangeConfig }

func newBinance(cfg config.ExchangeConfig) Connector { return &binance{cfg: cfg} }
func (b *binance) Name() string                      { return "binance" }
func (b *binance) Run(ctx context.Context, out chan<- *market.Trade) error {
	return reconnect(ctx, b.Name(), func(ctx context.Context) error {
		streams := make([]string, 0, len(b.cfg.Symbols))
		for _, symbol := range b.cfg.Symbols {
			streams = append(streams, strings.ToLower(symbol)+"@trade")
		}
		endpoint := b.cfg.URL
		if endpoint == "" {
			endpoint = "wss://stream.binance.com:9443/stream?streams=" + strings.Join(streams, "/")
		}
		conn, _, err := websocket.Dial(ctx, endpoint, nil)
		if err != nil {
			return err
		}
		defer conn.CloseNow()
		for {
			var msg struct {
				Data struct {
					ID         int64  `json:"t"`
					Symbol     string `json:"s"`
					Price      string `json:"p"`
					Qty        string `json:"q"`
					Time       int64  `json:"T"`
					BuyerMaker bool   `json:"m"`
				} `json:"data"`
			}
			if err := wsjson.Read(ctx, conn, &msg); err != nil {
				return err
			}
			price, err := strconv.ParseFloat(msg.Data.Price, 64)
			if err != nil {
				continue
			}
			qty, err := strconv.ParseFloat(msg.Data.Qty, 64)
			if err != nil {
				continue
			}
			side := "buy"
			if msg.Data.BuyerMaker {
				side = "sell"
			}
			trade := &market.Trade{TradeID: fmt.Sprint(msg.Data.ID), Time: time.UnixMilli(msg.Data.Time).UTC(), Exchange: b.Name(), Symbol: msg.Data.Symbol, Price: price, Qty: qty, Side: side}
			if err := publish(ctx, out, trade); err != nil {
				return err
			}
		}
	})
}
