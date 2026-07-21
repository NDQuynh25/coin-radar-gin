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

type okx struct{ cfg config.ExchangeConfig }

func newOKX(cfg config.ExchangeConfig) Connector { return &okx{cfg: cfg} }
func (o *okx) Name() string                      { return "okx" }
func (o *okx) Run(ctx context.Context, out chan<- *market.Trade) error {
	return reconnect(ctx, o.Name(), func(ctx context.Context) error {
		endpoint := o.cfg.URL
		if endpoint == "" {
			endpoint = "wss://ws.okx.com:8443/ws/v5/public"
		}
		conn, _, err := websocket.Dial(ctx, endpoint, nil)
		if err != nil {
			return err
		}
		defer conn.CloseNow()
		args := make([]map[string]string, 0, len(o.cfg.Symbols))
		for _, s := range o.cfg.Symbols {
			args = append(args, map[string]string{"channel": "trades", "instId": strings.ToUpper(s)})
		}
		if err := wsjson.Write(ctx, conn, map[string]any{"op": "subscribe", "args": args}); err != nil {
			return err
		}
		for {
			var msg struct {
				Data []struct {
					ID     string `json:"tradeId"`
					Symbol string `json:"instId"`
					Side   string `json:"side"`
					Qty    string `json:"sz"`
					Price  string `json:"px"`
					Time   string `json:"ts"`
				} `json:"data"`
			}
			if err := wsjson.Read(ctx, conn, &msg); err != nil {
				return err
			}
			for _, d := range msg.Data {
				price, e1 := strconv.ParseFloat(d.Price, 64)
				qty, e2 := strconv.ParseFloat(d.Qty, 64)
				millis, e3 := strconv.ParseInt(d.Time, 10, 64)
				if e1 != nil || e2 != nil || e3 != nil {
					continue
				}
				if err := publish(ctx, out, &market.Trade{TradeID: d.ID, Time: time.UnixMilli(millis).UTC(), Exchange: o.Name(), Symbol: d.Symbol, Price: price, Qty: qty, Side: strings.ToLower(d.Side)}); err != nil {
					return err
				}
			}
		}
	})
}
