// Định nghĩa toàn bộ endpoint REST của các sàn cần test (T1.1).
// Mỗi endpoint: { name, url, check } — check(json) trả về chuỗi tóm tắt data hoặc throw nếu sai.
// Không cần API key — tất cả đều là public market data.

const SYMBOL = 'BTCUSDT';
const OKX_SWAP = 'BTC-USDT-SWAP';

const num = (v) => (v === undefined || v === null ? NaN : Number(v));
const must = (cond, msg) => { if (!cond) throw new Error(msg); };

const exchanges = {
  // ─────────────────────────────── BINANCE ───────────────────────────────
  binance: {
    label: 'Binance',
    rest: [
      {
        name: 'Spot price',
        url: `https://api.binance.com/api/v3/ticker/price?symbol=${SYMBOL}`,
        check: (j) => { must(num(j.price) > 0, 'price <= 0'); return `giá=${j.price}`; },
      },
      {
        name: 'Spot 24h ticker',
        url: `https://api.binance.com/api/v3/ticker/24hr?symbol=${SYMBOL}`,
        check: (j) => { must(num(j.lastPrice) > 0, 'no lastPrice'); return `last=${j.lastPrice} vol=${j.volume} change=${j.priceChangePercent}%`; },
      },
      {
        name: 'Spot klines 1m',
        url: `https://api.binance.com/api/v3/klines?symbol=${SYMBOL}&interval=1m&limit=3`,
        check: (j) => { must(Array.isArray(j) && j.length === 3, 'klines len != 3'); return `${j.length} nến, nến cuối close=${j[2][4]}`; },
      },
      {
        name: 'Spot recent trades',
        url: `https://api.binance.com/api/v3/trades?symbol=${SYMBOL}&limit=5`,
        check: (j) => { must(Array.isArray(j) && j.length > 0, 'no trades'); return `${j.length} trades, giá cuối=${j[j.length-1].price}`; },
      },
      {
        name: 'Spot orderbook depth',
        url: `https://api.binance.com/api/v3/depth?symbol=${SYMBOL}&limit=5`,
        check: (j) => { must(j.bids && j.asks, 'no book'); return `bid top=${j.bids[0][0]} ask top=${j.asks[0][0]}`; },
      },
      {
        name: 'Futures funding + mark',
        url: `https://fapi.binance.com/fapi/v1/premiumIndex?symbol=${SYMBOL}`,
        check: (j) => { must(j.lastFundingRate !== undefined, 'no funding'); return `funding=${(num(j.lastFundingRate)*100).toFixed(4)}% mark=${j.markPrice}`; },
      },
      {
        name: 'Futures open interest',
        url: `https://fapi.binance.com/fapi/v1/openInterest?symbol=${SYMBOL}`,
        check: (j) => { must(num(j.openInterest) > 0, 'no OI'); return `OI=${j.openInterest} BTC`; },
      },
      {
        name: 'Futures funding history',
        url: `https://fapi.binance.com/fapi/v1/fundingRate?symbol=${SYMBOL}&limit=3`,
        check: (j) => { must(Array.isArray(j) && j.length > 0, 'no funding hist'); return `${j.length} mốc, gần nhất=${(num(j[j.length-1].fundingRate)*100).toFixed(4)}%`; },
      },
      {
        name: 'Futures OI history (5m)',
        url: `https://fapi.binance.com/futures/data/openInterestHist?symbol=${SYMBOL}&period=5m&limit=3`,
        check: (j) => { must(Array.isArray(j) && j.length > 0, 'no OI hist'); return `${j.length} mốc, OI cuối=${j[j.length-1].sumOpenInterest}`; },
      },
      {
        name: 'Futures Long/Short ratio',
        url: `https://fapi.binance.com/futures/data/globalLongShortAccountRatio?symbol=${SYMBOL}&period=5m&limit=1`,
        check: (j) => { must(Array.isArray(j) && j.length > 0, 'no LS ratio'); return `long/short=${j[0].longShortRatio}`; },
      },
    ],
  },

  // ─────────────────────────────── BYBIT ───────────────────────────────
  bybit: {
    label: 'Bybit',
    rest: [
      {
        name: 'Linear ticker (gộp giá+funding+OI)',
        url: `https://api.bybit.com/v5/market/tickers?category=linear&symbol=${SYMBOL}`,
        check: (j) => { must(j.retCode === 0, `retCode=${j.retCode}`); const t = j.result.list[0]; return `giá=${t.lastPrice} funding=${(num(t.fundingRate)*100).toFixed(4)}% OI=${t.openInterest}`; },
      },
      {
        name: 'Kline 1m',
        url: `https://api.bybit.com/v5/market/kline?category=linear&symbol=${SYMBOL}&interval=1&limit=3`,
        check: (j) => { must(j.retCode === 0, `retCode=${j.retCode}`); return `${j.result.list.length} nến, close=${j.result.list[0][4]}`; },
      },
      {
        name: 'Recent trades',
        url: `https://api.bybit.com/v5/market/recent-trade?category=linear&symbol=${SYMBOL}&limit=5`,
        check: (j) => { must(j.retCode === 0, `retCode=${j.retCode}`); return `${j.result.list.length} trades, giá=${j.result.list[0].price}`; },
      },
      {
        name: 'Open interest history',
        url: `https://api.bybit.com/v5/market/open-interest?category=linear&symbol=${SYMBOL}&intervalTime=5min&limit=3`,
        check: (j) => { must(j.retCode === 0, `retCode=${j.retCode}`); return `${j.result.list.length} mốc OI, cuối=${j.result.list[0].openInterest}`; },
      },
      {
        name: 'Funding rate history',
        url: `https://api.bybit.com/v5/market/funding/history?category=linear&symbol=${SYMBOL}&limit=3`,
        check: (j) => { must(j.retCode === 0, `retCode=${j.retCode}`); return `${j.result.list.length} mốc funding, cuối=${(num(j.result.list[0].fundingRate)*100).toFixed(4)}%`; },
      },
      {
        name: 'Orderbook',
        url: `https://api.bybit.com/v5/market/orderbook?category=linear&symbol=${SYMBOL}&limit=5`,
        check: (j) => { must(j.retCode === 0, `retCode=${j.retCode}`); return `bid=${j.result.b[0][0]} ask=${j.result.a[0][0]}`; },
      },
    ],
  },

  // ─────────────────────────────── OKX ───────────────────────────────
  okx: {
    label: 'OKX',
    rest: [
      {
        name: 'Swap ticker',
        url: `https://www.okx.com/api/v5/market/ticker?instId=${OKX_SWAP}`,
        check: (j) => { must(j.code === '0', `code=${j.code}`); return `giá=${j.data[0].last} vol24h=${j.data[0].vol24h}`; },
      },
      {
        name: 'Funding rate',
        url: `https://www.okx.com/api/v5/public/funding-rate?instId=${OKX_SWAP}`,
        check: (j) => { must(j.code === '0', `code=${j.code}`); return `funding=${(num(j.data[0].fundingRate)*100).toFixed(4)}%`; },
      },
      {
        name: 'Open interest',
        url: `https://www.okx.com/api/v5/public/open-interest?instId=${OKX_SWAP}`,
        check: (j) => { must(j.code === '0', `code=${j.code}`); return `OI=${j.data[0].oi} (${j.data[0].oiCcy} BTC)`; },
      },
      {
        name: 'Candles 1m',
        url: `https://www.okx.com/api/v5/market/candles?instId=${OKX_SWAP}&bar=1m&limit=3`,
        check: (j) => { must(j.code === '0', `code=${j.code}`); return `${j.data.length} nến, close=${j.data[0][4]}`; },
      },
      {
        name: 'Recent trades',
        url: `https://www.okx.com/api/v5/market/trades?instId=${OKX_SWAP}&limit=5`,
        check: (j) => { must(j.code === '0', `code=${j.code}`); return `${j.data.length} trades, giá=${j.data[0].px}`; },
      },
      {
        name: 'Funding rate history',
        url: `https://www.okx.com/api/v5/public/funding-rate-history?instId=${OKX_SWAP}&limit=3`,
        check: (j) => { must(j.code === '0', `code=${j.code}`); return `${j.data.length} mốc funding lịch sử`; },
      },
    ],
  },
};

// Định nghĩa WebSocket streams cần test
const websockets = {
  binance: {
    label: 'Binance',
    url: 'wss://stream.binance.com:9443/ws/btcusdt@aggTrade',
    onMessage: (raw) => { const t = JSON.parse(raw); return `trade giá=${t.p} qty=${t.q}`; },
    subscribe: null, // Binance: stream nằm sẵn trong URL, không cần gửi subscribe
  },
  bybit: {
    label: 'Bybit',
    url: 'wss://stream.bybit.com/v5/public/linear',
    subscribe: JSON.stringify({ op: 'subscribe', args: ['publicTrade.BTCUSDT'] }),
    onMessage: (raw) => {
      const m = JSON.parse(raw);
      if (m.op === 'subscribe' || !m.data) return null; // bỏ qua ack
      return `trade giá=${m.data[0].p} qty=${m.data[0].v}`;
    },
  },
  okx: {
    label: 'OKX',
    url: 'wss://ws.okx.com:8443/ws/v5/public',
    subscribe: JSON.stringify({ op: 'subscribe', args: [{ channel: 'trades', instId: 'BTC-USDT-SWAP' }] }),
    onMessage: (raw) => {
      const m = JSON.parse(raw);
      if (m.event || !m.data) return null; // bỏ qua ack
      return `trade giá=${m.data[0].px} qty=${m.data[0].sz}`;
    },
  },
};

module.exports = { exchanges, websockets, SYMBOL };
