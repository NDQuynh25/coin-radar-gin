package exchange

import (
	"context"
	"strconv"
	"strings"
	"time"

	"coin-radar-gin/config"
	"coin-radar-gin/internal/modules/market"

	"github.com/coder/websocket"
	"github.com/coder/websocket/wsjson"
)

type bybit struct{ cfg config.ExchangeConfig }

func newBybit(cfg config.ExchangeConfig) Connector { return &bybit{cfg: cfg} }
func (b *bybit) Name() string                      { return "bybit" }
func (b *bybit) Run(ctx context.Context, out chan<- *market.Trade) error {
	return reconnect(ctx, b.Name(), func(ctx context.Context) error {
		endpoint := b.cfg.URL
		if endpoint == "" {
			endpoint = "wss://stream.bybit.com/v5/public/" + marketName(b.cfg.Market)
		}
		conn, _, err := websocket.Dial(ctx, endpoint, nil)
		if err != nil {
			return err
		}
		defer conn.CloseNow()
		args := make([]string, 0, len(b.cfg.Symbols))
		for _, s := range b.cfg.Symbols {
			args = append(args, "publicTrade."+strings.ToUpper(s))
		}
		if err := wsjson.Write(ctx, conn, map[string]any{"op": "subscribe", "args": args}); err != nil {
			return err
		}
		for {
			var msg struct {
				Data []struct {
					ID     string `json:"i"`
					Symbol string `json:"s"`
					Side   string `json:"S"`
					Qty    string `json:"v"`
					Price  string `json:"p"`
					Time   int64  `json:"T"`
				} `json:"data"`
			}
			if err := wsjson.Read(ctx, conn, &msg); err != nil {
				return err
			}
			for _, d := range msg.Data {
				price, e1 := strconv.ParseFloat(d.Price, 64)
				qty, e2 := strconv.ParseFloat(d.Qty, 64)
				if e1 != nil || e2 != nil {
					continue
				}
				if err := publish(ctx, out, &market.Trade{TradeID: d.ID, Time: time.UnixMilli(d.Time).UTC(), Exchange: b.Name(), Symbol: d.Symbol, Price: price, Qty: qty, Side: strings.ToLower(d.Side)}); err != nil {
					return err
				}
			}
		}
	})
}

func marketName(name string) string {
	if name == "linear" || name == "inverse" {
		return name
	}
	return "spot"
}
