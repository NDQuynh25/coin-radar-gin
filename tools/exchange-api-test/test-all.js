#!/usr/bin/env node
// Bộ test toàn bộ API bên thứ 3 của các sàn (REST + WebSocket).
// Chạy:  node test-all.js            -> test tất cả
//        node test-all.js binance    -> chỉ test 1 sàn
//        node test-all.js --rest      -> chỉ REST (bỏ WebSocket)
//
// Không cần API key. Yêu cầu Node >= 21 (có WebSocket built-in).

const { exchanges, websockets } = require('./endpoints');

const C = { green: '\x1b[32m', red: '\x1b[31m', yellow: '\x1b[33m', dim: '\x1b[2m', reset: '\x1b[0m', bold: '\x1b[1m' };
const ok = (s) => `${C.green}✅ ${s}${C.reset}`;
const fail = (s) => `${C.red}❌ ${s}${C.reset}`;
const REQUEST_TIMEOUT = 15000;
const WS_WAIT = 12000;

const args = process.argv.slice(2);
const onlyExchange = args.find((a) => !a.startsWith('--'));
const restOnly = args.includes('--rest');

const results = { pass: 0, fail: 0, failures: [] };

// ───────────────────────── REST ─────────────────────────
async function fetchJSON(url) {
  const ctrl = new AbortController();
  const timer = setTimeout(() => ctrl.abort(), REQUEST_TIMEOUT);
  try {
    const res = await fetch(url, { signal: ctrl.signal, headers: { 'User-Agent': 'crypto-data-platform-test/1.0' } });
    if (!res.ok) throw new Error(`HTTP ${res.status}`);
    return await res.json();
  } finally {
    clearTimeout(timer);
  }
}

async function testRestEndpoint(exId, ep) {
  const t0 = Date.now();
  try {
    const json = await fetchJSON(ep.url);
    const summary = ep.check(json);
    const ms = Date.now() - t0;
    console.log(`  ${ok(ep.name.padEnd(38))} ${C.dim}${String(ms).padStart(5)}ms${C.reset}  ${summary}`);
    results.pass++;
  } catch (e) {
    console.log(`  ${fail(ep.name.padEnd(38))} ${C.red}${e.message}${C.reset}`);
    results.fail++;
    results.failures.push(`${exId} REST / ${ep.name}: ${e.message}`);
  }
}

async function testRestForExchange(exId) {
  const ex = exchanges[exId];
  console.log(`\n${C.bold}━━━ ${ex.label} — REST (${ex.rest.length} endpoints) ━━━${C.reset}`);
  for (const ep of ex.rest) {
    await testRestEndpoint(exId, ep); // tuần tự để tôn trọng rate limit
  }
}

// ───────────────────────── WebSocket ─────────────────────────
function testWebSocket(exId) {
  return new Promise((resolve) => {
    const cfg = websockets[exId];
    const t0 = Date.now();
    let received = 0;
    let settled = false;

    const finish = (success, detail) => {
      if (settled) return;
      settled = true;
      try { ws.close(); } catch {}
      clearTimeout(timer);
      if (success) {
        console.log(`  ${ok(('WS ' + cfg.label).padEnd(38))} ${C.dim}${Date.now() - t0}ms${C.reset}  ${detail}`);
        results.pass++;
      } else {
        console.log(`  ${fail(('WS ' + cfg.label).padEnd(38))} ${C.red}${detail}${C.reset}`);
        results.fail++;
        results.failures.push(`${exId} WS: ${detail}`);
      }
      resolve();
    };

    let ws;
    try {
      ws = new WebSocket(cfg.url);
    } catch (e) {
      return finish(false, 'không tạo được WebSocket: ' + e.message);
    }

    const timer = setTimeout(
      () => finish(received > 0, received > 0 ? `nhận ${received} msg` : 'timeout, không nhận được data'),
      WS_WAIT
    );

    ws.addEventListener('open', () => {
      if (cfg.subscribe) ws.send(cfg.subscribe);
    });
    ws.addEventListener('message', (e) => {
      try {
        const summary = cfg.onMessage(typeof e.data === 'string' ? e.data : e.data.toString());
        if (summary === null) return; // bỏ qua ack / heartbeat
        received++;
        if (received === 1) finish(true, `kết nối OK, msg đầu: ${summary}`);
      } catch (err) {
        finish(false, 'lỗi parse message: ' + err.message);
      }
    });
    ws.addEventListener('error', (e) => finish(false, 'WS error: ' + (e.message || 'unknown')));
    ws.addEventListener('close', () => { if (!settled) finish(received > 0, received > 0 ? `nhận ${received} msg` : 'đóng sớm'); });
  });
}

// ───────────────────────── Main ─────────────────────────
async function main() {
  const ids = onlyExchange ? [onlyExchange] : Object.keys(exchanges);
  for (const id of ids) {
    if (!exchanges[id]) { console.log(fail(`Không có sàn '${id}'. Chọn: ${Object.keys(exchanges).join(', ')}`)); process.exit(1); }
  }

  console.log(`${C.bold}🔌 TEST API CÁC SÀN — ${ids.map((i) => exchanges[i].label).join(', ')}${C.reset}`);
  console.log(`${C.dim}Node ${process.version} · ${restOnly ? 'chỉ REST' : 'REST + WebSocket'} · không cần API key${C.reset}`);

  for (const id of ids) await testRestForExchange(id);

  if (!restOnly) {
    console.log(`\n${C.bold}━━━ WebSocket (realtime trade stream) ━━━${C.reset}`);
    for (const id of ids) await testWebSocket(id);
  }

  // Tổng kết
  const total = results.pass + results.fail;
  console.log(`\n${C.bold}═══════════════ TỔNG KẾT ═══════════════${C.reset}`);
  console.log(`  ${C.green}Pass: ${results.pass}${C.reset} / ${total}    ${results.fail ? C.red : C.dim}Fail: ${results.fail}${C.reset}`);
  if (results.failures.length) {
    console.log(`\n  ${C.red}Chi tiết lỗi:${C.reset}`);
    results.failures.forEach((f) => console.log(`   • ${f}`));
  }
  process.exit(results.fail ? 1 : 0);
}

main().catch((e) => { console.error(fail('Lỗi runner: ' + e.message)); process.exit(1); });
