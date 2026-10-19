// app.js — the production desk's frontend half: THE COALESCED LOG contract
// (spec-gui-ui, ui.py:65-86 + 2374-2794) as the HTML twin —
//   · one log:batch event = ONE DocumentFragment append + ONE tail scroll
//     (the single cursor-edit-block law);
//   · the cache is the truth (the _all_log_entries twin, capped at
//     _LOG_CAP 1000): filters/search re-render FROM CACHE, never by
//     re-parsing the document;
//   · the document mirrors the cache's rendered subset: head-trim keeps
//     the desk at ≤1000 rows whatever the flood;
//   · the row: level glyph (shape-coded, level-colored; the message ink is
//     ONE muted tone — the v0.32 law) · [ts] · elided message, with the
//     full text on the row tooltip.
// The shell (T-72) binds internal/app directly: the typed events cross
// as bus:* wails events; the desk's rows ride the deskbus's coalesced
// log:batch batches. The glyphs ride the lucide-static pack (icons.js —
// the reference's _LOG_LEVEL_ICONS law: info=dot, success=check,
// warning=triangle-alert, error=circle-x; tint via currentColor).
// The flood probe is the perf-test instrument (the log perf test's
// interactivity evidence — handler-time + the task-gap sampler with
// the Xvfb timer-clamp caveat).
import { MarkFirstFrame, FloodLines, Stats, ReportFloodUI, ClearDesk } from './wailsjs/go/main/Shell.js';
import { EventsOn } from './wailsjs/runtime/runtime.js';
import { LEVEL_ICONS, svgEl } from './icons.js';

const desk = document.getElementById('log-desk');
const logEmpty = document.getElementById('log-empty');
const countEl = document.getElementById('log-count');
const cacheEl = document.getElementById('log-cache');
const sentinels = document.querySelector('.sentinels');
const search = document.getElementById('log-search');

const LOG_CAP = 1000;          // _LOG_CAP (ui.py:86)
let cache = [];                // _all_log_entries twin — THE TRUTH
let renderedCount = 0;         // _log_visible_count twin
let chip = 'all';
let filterText = '';

// the lucide-static v0.544.0 glyph pack rides icons.js (golden-locked);
// the desk's level map is the reference's own law (ui.py:2144).
function glyphEl(level) {
  const known = LEVEL_ICONS[level] !== undefined;
  const name = LEVEL_ICONS[level] || 'dot';
  return svgEl(name, 'glyph glyph-' + (known ? level : 'info'));
}

// _log_entry_would_render twin: the LIVE chip + search decide what lands.
function wouldRender(e) {
  if (chip !== 'all' && e.level !== chip) return false;
  if (filterText && !String(e.msg).toLowerCase().includes(filterText)) return false;
  return true;
}

function hideEmpty() { if (logEmpty.parentNode) logEmpty.remove(); }
function showEmpty() {
  hideEmpty();
  desk.textContent = '';
  desk.appendChild(logEmpty);
  renderedCount = 0;
}

function updateCounters() {
  countEl.textContent = renderedCount + ' rows';
  cacheEl.textContent = 'cache ' + cache.length;
}

// ONE batch = ONE fragment append + ONE tail scroll (the flush law).
function appendToDesk(rows) {
  hideEmpty();
  const frag = document.createDocumentFragment();
  for (const r of rows) {
    const row = document.createElement('div');
    row.className = 'log-row';
    row.title = '[' + r.ts + '] ' + r.msg;      // the tooltip = full text
    const ts = document.createElement('span');
    ts.className = 'log-ts';
    ts.textContent = '[' + r.ts + '] ';
    const msg = document.createElement('span');
    msg.className = 'log-msg';
    msg.textContent = r.msg;                    // CSS elides right (NoWrap)
    row.append(glyphEl(r.level), ts, msg);
    frag.appendChild(row);
  }
  desk.appendChild(frag);
  renderedCount += rows.length;
  // head-trim: the document mirrors the cache's rendered subset (≤ LOG_CAP)
  if (renderedCount > LOG_CAP) {
    const excess = renderedCount - LOG_CAP;
    for (let i = 0; i < excess; i++) desk.removeChild(desk.firstChild);
    renderedCount = LOG_CAP;
  }
  desk.scrollTop = desk.scrollHeight;           // ONE tail scroll per batch
}

function appendRows(rows) {
  for (const r of rows) cache.push(r);
  if (cache.length > LOG_CAP) cache = cache.slice(cache.length - LOG_CAP);
  const renderable = rows.filter(wouldRender);
  if (renderable.length) appendToDesk(renderable);
  if (cache.length === 0) showEmpty();
  updateCounters();
}

// the filter/search doors re-render FROM CACHE (never re-parse the doc).
function rerenderFromCache() {
  desk.textContent = '';
  renderedCount = 0;
  if (cache.length === 0) {
    showEmpty();
    updateCounters();
    return;
  }
  appendToDesk(cache.filter(wouldRender));
  updateCounters();
}

// the desk's rows: the deskbus's coalesced batches (log:batch).
EventsOn('log:batch', (rows) => {
  measureHandler(rows || [], () => appendRows(rows || []));
});

// the typed bus events (the production transport): the desk's own log
// rows. T-77: the progress BAR + the batch chime moved to heroflow.js
// (the chrome's module — one owner per surface; the tags both read are
// pinned by TestCrossedPayloadFieldTags).
EventsOn('bus:status', (st) => {
  if (st && st.text) appendRows([{ ts: nowTs(), level: 'info', msg: st.text }]);
});
EventsOn('bus:batch-finished', (bf) => {
  if (!bf) return;
  appendRows([{ ts: nowTs(), level: bf.ok ? 'success' : 'error', msg: 'batch: ' + (bf.summary || '') }]);
});

// the flood probe — TWO honest instruments:
//   · the handler timer: every log:batch event's own handling time (the
//     DOM append work the flood costs the main thread — THE interactivity
//     evidence, measured where the work happens);
//   · a setTimeout-chain sampler: the event loop's task-queue latency.
//     CAVEAT (measured, not guessed): under a WM-less Xvfb WebKit treats
//     the page as background and CLAMPS timers to ~1 Hz — the chain's gap
//     numbers read the throttling, not blockage; the verdict cites the
//     handler numbers + the rendered desk, and says so.
let probe = null;
function startProbe() {
  probe = {
    samples: 0, maxGap: 0, gapsOver100: 0, last: 0, timer: 0,
    handlerBatches: 0, handlerMsTotal: 0, handlerMsMax: 0, rowsAppended: 0,
  };
  const tick = () => {
    const now = performance.now();
    if (probe.last > 0) {
      const gap = now - probe.last;
      if (gap > probe.maxGap) probe.maxGap = gap;
      if (gap > 100) probe.gapsOver100++;
    }
    probe.last = now;
    probe.samples++;
    probe.timer = setTimeout(tick, 0);
  };
  tick();
}
function measureHandler(rows, fn) {
  if (!probe) {
    fn();
    return;
  }
  const t0 = performance.now();
  fn();
  const dt = performance.now() - t0;
  probe.handlerBatches++;
  probe.handlerMsTotal += dt;
  if (dt > probe.handlerMsMax) probe.handlerMsMax = dt;
  probe.rowsAppended += rows.length;
}
EventsOn('flood:start', () => startProbe());
EventsOn('flood:done', () => {
  if (probe) {
    clearTimeout(probe.timer);
    const p = probe;
    probe = null;
    ReportFloodUI({
      samples: p.samples,
      max_task_gap_ms: Math.round(p.maxGap),
      gaps_over_100ms: p.gapsOver100,
      handler_batches: p.handlerBatches,
      handler_ms_total: Math.round(p.handlerMsTotal * 10) / 10,
      handler_ms_max: Math.round(p.handlerMsMax * 10) / 10,
      rows_appended: p.rowsAppended,
    });
  }
});

// the first-frame badge (the cold-start clock's face)
MarkFirstFrame().then((d) => {
  const badge = document.createElement('span');
  badge.className = 'sentinel';
  badge.id = 'sentinel-frame';
  badge.textContent = 'startup→frame: ' + d;
  sentinels.appendChild(badge);
});

// the chips + search + clear doors
document.querySelectorAll('.chip').forEach((c) => {
  c.addEventListener('click', () => {
    document.querySelectorAll('.chip').forEach((x) => x.classList.remove('chip-active'));
    c.classList.add('chip-active');
    chip = c.dataset.filter;
    rerenderFromCache();
  });
});

search.addEventListener('input', () => {
  filterText = search.value.trim().toLowerCase();
  rerenderFromCache();
});

document.getElementById('btn-clear').addEventListener('click', async () => {
  try { await ClearDesk(); } catch (e) { /* the local wipe still happens */ }
  cache = [];
  showEmpty();
  updateCounters();
});

function nowTs() {
  const d = new Date();
  const p = (n) => String(n).padStart(2, '0');
  return p(d.getHours()) + ':' + p(d.getMinutes()) + ':' + p(d.getSeconds());
}
