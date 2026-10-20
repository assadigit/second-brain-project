// heroflow.js — T-77: the main window chrome's frontend half —
//   · THE HERO BUTTON MACHINE's DOM half: the hero:state crossings
//     render the ONE button (SYNC → FETCHING… → PROCESS (n) → Stop /
//     Stopping… — the label, the glyph, the disabled flag, the danger
//     fill, the tooltip); the click door runs the whole flow through
//     HeroUI.HeroClick (the guards' boxes render here, the answers
//     ride the answer doors);
//   · THE PROGRESS ROW's live formats: the hero payload's format +
//     count, the bus:status texts (the "Processing…" family), the
//     batch-finished terminal marks;
//   · THE SENTINEL LABELS (Net · AI · Proxy): the bus:sentinel
//     snapshot paints the dot + the word + the tooltip (never bare
//     dashes; color never carries the state alone), and the offline
//     law's warn-now opens the ONE themed box;
//   · THE BATTERY MODAL (the ConnectionTestDialog twin): the four
//     rows ( Vaults / LLM / GitHub / the Worker), the spinner while a
//     section runs, the compact one-line items with the full text on
//     the tooltip, and Start Syncing — born disabled, green only when
//     every row is ok;
//   · THE MORE MENU: the reference's own action list, verbatim — the
//     live doors wired (Test Connection, Open Dashboard, Dark Mode,
//     Settings), the not-yet-landed engines honestly disabled;
//   · THE BATCH CHIME: the batch-finished crossing's chime payload
//     plays through <audio> (the QSoundEffect twin — the exact batch
//     gate lives in Go).
import { HeroClick, HeroAnswerOffline, HeroRunQueue, HeroState, TestConnection,
         MoreOpenDashboard, MoreDarkMode, MoreRetryBacklog, MoreRetryMaster,
         MoreTestTelegramGitHub, MoreTestProxy, MoreTestOllama,
         MoreTestCloud, MoreTestLlamaCpp,
         ChromeBoot } from './wailsjs/go/main/HeroUI.js';
import { FastSync } from './wailsjs/go/main/PagesBUI.js';
import { EventsOn } from './wailsjs/runtime/runtime.js';
import { svgEl } from './icons.js';

// ---------------------------------------------------------------------------
// tiny DOM helpers (the scanflow.js kit's local twins).
// ---------------------------------------------------------------------------
function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined && text !== null) n.textContent = text;
  return n;
}
function btn(text, kind, title, onClick) {
  const b = el('button', 'btn-' + kind, text);
  b.type = 'button';
  if (title) b.title = title;
  if (onClick) b.addEventListener('click', onClick);
  return b;
}

// one overlay at a time (the modal stack law).
let overlay = null;
function openOverlay(width) {
  closeOverlay();
  overlay = el('div', 'dialog-overlay heroflow-overlay');
  const box = el('div', 'dialog-box heroflow-box');
  if (width) box.style.minWidth = width + 'px';
  overlay.appendChild(box);
  overlay.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.stopPropagation(); closeOverlay(); }
  });
  document.getElementById('app').appendChild(overlay);
  return box;
}
function closeOverlay() {
  if (overlay && overlay.parentNode) overlay.remove();
  overlay = null;
}

// themedBox is the _show_custom_message_box / _show_custom_question
// twin (the error boxes, the guard questions — the hero flow's own).
function themedBox(box, buttons) {
  const b = openOverlay(460);
  b.setAttribute('role', 'dialog');
  b.setAttribute('aria-label', box.title);
  b.appendChild(el('div', 'card-heading', box.title));
  const body = el('div', 'dialog-body');
  body.appendChild(el('p', 'dialog-message', box.body));
  b.appendChild(body);
  const btnRow = el('div', 'dialog-buttons');
  for (const [text, kind, fn] of buttons) {
    btnRow.appendChild(btn(text, kind, '', fn));
  }
  b.appendChild(btnRow);
}

// ---------------------------------------------------------------------------
// THE HERO BUTTON MACHINE — the DOM half.
// ---------------------------------------------------------------------------
const heroBtn = document.getElementById('btn-sync');
const fastBtn = document.getElementById('btn-fast');
const testBtn = document.getElementById('btn-test');
const barFill = document.getElementById('bar-fill');
const barLabel = document.getElementById('bar-label');

// the live import-file fact (the reference reads the Input page's
// widget; the port's save-on-change law makes the field the settings
// module's — read it live at click time).
function liveImportFile() {
  const f = document.getElementById('set-import-file');
  return f ? f.value.trim() : '';
}

const HERO_GLYPHS = {
  refresh: 'refresh-cw', loader: 'loader', play: 'play', stop: 'square',
};

function renderHero(payload) {
  if (!heroBtn || !payload || !payload.render) return;
  const r = payload.render;
  heroBtn.textContent = r.label;
  heroBtn.disabled = !r.enabled;
  heroBtn.classList.toggle('btn-hero-danger', !!r.danger);
  if (r.tooltip) heroBtn.title = r.tooltip;
  // the glyph rides the lucide pack (currentColor tint)
  const old = heroBtn.querySelector('svg');
  if (old) old.remove();
  const name = HERO_GLYPHS[r.glyph] || 'refresh-cw';
  const g = svgEl(name, 'hero-glyph' + (r.glyph === 'loader' ? ' hero-glyph-spin' : ''));
  if (g) heroBtn.prepend(g);
  // the progress row's live formats ride the same crossing.
  if (payload.format) {
    barLabel.textContent = payload.format;
    barLabel.className = 'progress-count';
    if (payload.format === 'Ready') barFill.style.width = '0%';
  }
  if (payload.count) {
    barLabel.textContent = payload.count;
  }
}

EventsOn('hero:state', (payload) => { renderHero(payload); });

if (heroBtn) {
  heroBtn.disabled = false;          // T-77: the hero machine is live
  heroBtn.addEventListener('click', async () => {
    let res;
    try {
      res = await HeroClick(liveImportFile());
    } catch (e) {
      return;                        // the binding never rejects in production
    }
    handleHeroResult(res);
  });
}

function handleHeroResult(res) {
  if (!res) return;
  if (res.render) renderHero(res);
  if (res.Box) {
    // a guard box is the whole answer (an error box — OK closes).
    themedBox(res.Box, [['OK', 'primary', () => closeOverlay()]]);
    return;
  }
  if (res.Offline) {
    // the continue-anyway question (v0.85.0's advisory — always escapable).
    themedBox({
      title: '⚠️ No internet connection',
      body: "The 'Net' indicator says you are offline.\n\n" +
        'SYNC fetches undone items from your Telegram bot (through your ' +
        'proxy) and the batch needs GitHub and web fetches — all of them ' +
        'need the internet.\n\nContinue anyway?',
    }, [
      ['Continue anyway', 'primary', async () => {
        closeOverlay();
        handleHeroResult(await HeroAnswerOffline(true, liveImportFile()).catch(() => null));
      }],
      ['Cancel', 'secondary', () => {
        closeOverlay();
        // stay in SYNC; retry when back online.
        HeroAnswerOffline(false, '').catch(() => {});
      }],
    ]);
    return;
  }
  if (res.Confirm) {
    // the >10 confirm gate (v31.1: the exact item count, stated).
    themedBox(res.Confirm, [
      ['Continue', 'primary', async () => {
        closeOverlay();
        await HeroRunQueue().catch(() => {});
      }],
      ['Cancel', 'secondary', () => closeOverlay()],
    ]);
  }
}

// ---------------------------------------------------------------------------
// THE PROGRESS ROW — the status texts + the terminal marks (the bus's
// own payloads; the hero formats ride hero:state above).
// ---------------------------------------------------------------------------
EventsOn('bus:status', (st) => {
  if (!st || !st.text) return;
  if (st.text !== 'Ready') {
    barLabel.textContent = st.text;
    barLabel.className = 'progress-count';
  }
});
EventsOn('bus:progress', (p) => {
  if (!p || !p.total) return;
  barFill.style.width = ((p.pos / p.total) * 100).toFixed(1) + '%';
  barLabel.textContent = p.pos + ' / ' + p.total;
});
EventsOn('bus:job-progress', (jp) => {
  if (!jp || !(jp.max > 0)) return;
  barFill.style.width = ((jp.value / jp.max) * 100).toFixed(1) + '%';
  barLabel.textContent = jp.label + ' — ' + jp.value + ' / ' + jp.max;
});

// THE BATCH CHIME (the exact batch gate lives in Go; the <audio> is
// the QSoundEffect twin — a machine without an audio backend accepts
// the play and stays silent, never an error).
EventsOn('bus:batch-finished', (bf) => {
  if (!bf) return;
  barLabel.textContent = bf.ok ? '✅ Done' : '❌ Failed';
  barLabel.className = 'progress-count ' + (bf.ok ? 'done' : 'failed');
  if (bf.chime && bf.chime.play && bf.chime.dataUrl) {
    try {
      const audio = new Audio(bf.chime.dataUrl);
      audio.volume = Math.min(1, Math.max(0, bf.chime.volume || 0.8));
      const p = audio.play();
      if (p && p.catch) p.catch(() => { /* never an error */ });
    } catch (e) { /* never an error */ }
  }
});

// ---------------------------------------------------------------------------
// ⚡ FAST SYNC — the triad's own button (vault → GitHub → the website,
// no fetch, no LLM): the existing PagesB door with a busy render.
// ---------------------------------------------------------------------------
if (fastBtn) {
  fastBtn.disabled = false;
  fastBtn.title = '⚡ Fast Sync — the no-AI mirror of the triad: seal the ' +
    'GitHub vault, mirror the Websites vault, publish the website ' +
    'directory (no Telegram fetch, no LLM — v0.90.0).';
  fastBtn.addEventListener('click', async () => {
    if (fastBtn.disabled) return;
    fastBtn.disabled = true;
    fastBtn.textContent = '⚡ Fast Syncing…';
    try {
      await FastSync();
    } catch (e) { /* the verdict lines ride the desk */ }
    fastBtn.disabled = false;
    fastBtn.textContent = '⚡ Fast Sync';
  });
}

// ---------------------------------------------------------------------------
// THE SENTINEL LABELS — Net · AI · Proxy (the paint laws).
// ---------------------------------------------------------------------------
const SPIN = ['⠋', '⠙', '⠹', '⠸', '⠼', '⠴', '⠦', '⠧', '⠇', '⠏'];
let spinPos = 0;
setInterval(() => { spinPos = (spinPos + 1) % SPIN.length; }, 100);

function sentinelRow(id) {
  // the row: [dot][word] — built once, painted per snapshot.
  let row = document.getElementById(id);
  if (row) return row;
  row = el('span', 'sentinel sentinel-live');
  row.id = id;
  const dot = el('span', 'sentinel-dot');
  const word = el('span', 'sentinel-word', '…');
  row.append(dot, word);
  document.querySelector('.sentinels').appendChild(row);
  return row;
}

function paintSentinel(id, cue, checking) {
  const row = sentinelRow(id);
  const dot = row.querySelector('.sentinel-dot');
  const word = row.querySelector('.sentinel-word');
  if (checking) {
    dot.style.background = 'var(--text-muted)';
    word.textContent = SPIN[spinPos];
    return;
  }
  dot.style.background = cue.dot_color || 'var(--text-muted)';
  word.textContent = cue.word || '';
  word.className = 'sentinel-word sentinel-' + (cue.semantic || 'muted');
  row.title = cue.tooltip || cue.word || '';
}

function paintSentinels(ev) {
  if (!ev) return;
  // the paint laws (window.py's _apply_net/ai/proxy_status — the words
  // + dots + semantics live in Go; this only renders).
  paintSentinel('sentinel-net', {
    dot_color: ev.net && ev.net.online ? '#42B36B' : '#E85D75',
    word: 'Net: ' + (ev.net && ev.net.online ? 'online' : 'offline'),
    semantic: ev.net && ev.net.online ? 'success' : 'error',
    tooltip: (ev.net && ev.net.detail) || '',
  });
  const ai = ev.ai || {};
  let aiWord = 'AI: not configured';
  let aiSem = 'muted';
  let aiDot = '#8E8A90';
  if (ai.state === 'on') {
    aiWord = 'AI: ' + (ai.kind === 'cloud' ? 'on — cloud'
      : 'on — local (' + (ai.engine || 'local') + ')');
    aiSem = 'success'; aiDot = '#42B36B';
  } else if (ai.state === 'warn') {
    aiWord = 'AI: ' + (ai.short && ai.short.startsWith('warn') ? ai.short : 'warn — ' + (ai.short || 'warn'));
    aiSem = 'warning'; aiDot = '#D9A036';
  } else if (ai.state === 'off') {
    aiWord = 'AI: ' + (ai.short && ai.short.startsWith('off') ? ai.short : 'off — ' + (ai.short || 'off'));
    aiSem = 'error'; aiDot = '#E85D75';
  }
  paintSentinel('sentinel-ai', {
    dot_color: aiDot, word: aiWord, semantic: aiSem, tooltip: ai.detail || '',
  });
  const px = ev.proxy || {};
  let pxWord = 'Proxy: unreachable', pxSem = 'error', pxDot = '#E85D75';
  if (px.state === 'connected') { pxWord = 'Proxy: connected'; pxSem = 'success'; pxDot = '#42B36B'; }
  else if (px.state === 'disabled') { pxWord = 'Proxy: off'; pxSem = 'muted'; pxDot = '#8E8A90'; }
  paintSentinel('sentinel-proxy', {
    dot_color: pxDot, word: pxWord, semantic: pxSem, tooltip: px.detail || '',
  });
}

EventsOn('bus:sentinel', (ev) => {
  paintSentinels(ev);
  // THE OFFLINE LAW's one modal at open (warn-now) — the reference's
  // own text; later transitions spoke their lines on the desk already.
  if (ev.verdict === 'warn-now') {
    themedBox({
      title: 'No internet connection',
      body: 'GitCurator could not reach the internet (the bot service, ' +
        'GitHub and a generic endpoint all failed to answer).\n\n' +
        'This app needs the internet for syncing: fetching the Telegram ' +
        'bot queue, GitHub repo data, backup vault seals and the cloud ' +
        'LLM all require it.\n\n' +
        'What still works offline: local vault scans and hygiene, local ' +
        'LLM engines (Ollama / llama.cpp), and reading your notes.\n\n' +
        "The 'Net' indicator at the top will turn green as soon as you're " +
        'back online — no restart needed.',
    }, [['OK', 'primary', () => closeOverlay()]]);
  }
});

// ---------------------------------------------------------------------------
// THE BATTERY MODAL — the ConnectionTestDialog twin.
// ---------------------------------------------------------------------------
const CC_ROWS = [
  { icon: '📁', title: 'Vaults', tip: 'found + writable (ready to receive notes)' },
  { icon: '🧠', title: 'LLM', tip: 'the ACTIVE provider: cloud / Ollama / llama.cpp' },
  { icon: '🐙', title: 'GitHub', tip: 'token valid + the backup repos ready' },
  { icon: '☁️', title: 'Worker', tip: 'the sync queue reachable through the Cloudflare Worker' },
];
const CC_MARK = { ok: '✅', warn: '⚠️', error: '❌' };
const CC_LABEL = { ok: 'Connected', warn: 'Connected (warnings)', error: 'Not connected' };
let ccOpen = false;

function openBatteryModal() {
  const b = openOverlay(560);
  b.setAttribute('role', 'dialog');
  b.setAttribute('aria-label', 'Test Connection');
  b.appendChild(el('div', 'card-heading', 'Test Connection'));
  const body = el('div', 'dialog-body cc-body');
  body.appendChild(el('p', 'dialog-message',
    'Checking every subsystem — vaults, LLM, GitHub and the sync queue…'));
  const card = el('div', 'sync_card cc-rows-card');
  const rows = [];
  for (const r of CC_ROWS) {
    const row = el('div', 'cc-row');
    const head = el('div', 'cc-row-head');
    const name = el('span', 'cc_row_name', r.icon + ' ' + r.title);
    name.title = r.tip;
    const status = el('span', 'cc_row_status', '⏳ Waiting…');
    head.append(name, status);
    const items = el('div', 'cc-items');
    row.append(head, items);
    card.appendChild(row);
    rows.push({ status, items, done: false, spinning: false });
  }
  body.appendChild(card);
  b.appendChild(body);
  const btnRow = el('div', 'dialog-buttons');
  const closeB = btn('Close', 'secondary',
    'Close this dialog (the checks keep running in the log)', () => {
      ccOpen = false;
      closeOverlay();
    });
  const startB = btn('Start Syncing', 'primary',
    'Enabled when every subsystem is connected — starts the main view\'s SYNC flow (fetch the bot queue, then PROCESS).',
    () => {
      ccOpen = false;
      closeOverlay();
      // the hero's SYNC flow, one tick later (the modal fully closes
      // first — the reference's own deferred handoff).
      setTimeout(() => { HeroClick(liveImportFile()).then(handleHeroResult).catch(() => {}); }, 0);
    });
  startB.disabled = true;           // born dead — green only when all-ok
  btnRow.append(closeB, startB);
  b.appendChild(btnRow);
  return { rows, startB };
}

function ccSetChecking(idx) {
  if (!ccState || !ccState.rows[idx] || ccState.rows[idx].done) return;
  ccState.rows[idx].spinning = true;
  for (const it of ccState.rows[idx].items.children) it.style.display = 'none';
}

let ccState = null;
let ccSpinTimer = 0;

function ccAddItem(idx, short, full) {
  if (!ccState || !ccState.rows[idx]) return;
  const item = el('div', 'cc_item cc-item-line', short);
  item.title = full || short;
  ccState.rows[idx].items.appendChild(item);
}

function ccFinalize(idx, verdict) {
  if (!ccState || !ccState.rows[idx]) return;
  const r = ccState.rows[idx];
  r.spinning = false;
  r.done = true;
  r.verdict = verdict;
  r.status.textContent = (CC_MARK[verdict] || '•') + ' ' + (CC_LABEL[verdict] || verdict);
  r.status.className = 'cc_row_status cc-status-' + verdict;
}

function ccFinishAll() {
  if (!ccState) return;
  const all = ccState.rows.every((r) => r.done && r.verdict === 'ok');
  ccState.startB.disabled = !all;
  if (all) {
    ccState.startB.title = 'Every subsystem is connected — click to start SYNC.';
  }
}

EventsOn('cc:section', (ev) => { if (ccOpen) ccSetChecking((ev.idx || 1) - 1); });
EventsOn('cc:result', (ev) => { if (ccOpen) ccAddItem((ev.idx || 1) - 1, ev.short, ev.full); });
EventsOn('cc:done', (res) => {
  if (!ccOpen || !res) return;
  for (let i = 0; i < (res.rows || []).length; i++) {
    ccFinalize(i, res.rows[i].verdict);
  }
  ccFinishAll();
});

if (testBtn) {
  testBtn.disabled = false;
  testBtn.title = '🔌 Test Connection — the four-subsystem readiness check: ' +
    'vaults, the ACTIVE LLM provider, GitHub and the sync queue ' +
    '(the Worker). A modal opens with live per-subsystem status; ' +
    'Start Syncing turns green once everything is connected.';
  testBtn.addEventListener('click', async () => {
    if (ccOpen) return;
    ccOpen = true;
    ccState = openBatteryModal();
    clearInterval(ccSpinTimer);
    ccSpinTimer = setInterval(() => {
      if (!ccState) return;
      for (const r of ccState.rows) {
        if (r.spinning) r.status.textContent = SPIN[spinPos] + ' Checking…';
      }
    }, 100);
    try {
      const res = await TestConnection();
      if (res && res.refusal) {
        // the busy refusal — the desk already spoke the line.
        ccOpen = false;
        closeOverlay();
        return;
      }
      // the events drove the rows live; the settled verdicts land now.
      if (res && res.rows) {
        for (let i = 0; i < res.rows.length; i++) ccFinalize(i, res.rows[i].verdict);
        ccFinishAll();
      }
    } catch (e) { /* the lines ride the desk */ }
    clearInterval(ccSpinTimer);
  });
}

// ---------------------------------------------------------------------------
// THE MORE MENU — the reference's own action list, verbatim (ui.py
// 1849-1946). The live doors are wired; the engines that have not
// landed yet render HONESTLY DISABLED with their landing note (the
// same convention the CTA band rode until this task).
// ---------------------------------------------------------------------------
const MORE_ACTIONS = [
  { label: '🔌 Test Connection (all systems)', live: true, tip: 'The four-subsystem readiness check — the battery modal.' },
  { sep: true },
  { label: '🔑 Test Telegram & GitHub', live: 'testTgGh',
    tip: 'The GitHub token check + the sync queue (the Worker) — Option B re-scope: the Telegram leg rides the Worker queue (ADR-0004 Am.3).' },
  { label: '🌐 Test Proxy Connection', live: 'testProxy',
    tip: 'The configured proxy\'s TCP verdict — the same probe the sentinel paints (the deep Telegram round trip is Option B\'s accepted loss).' },
  { label: '🧠 Test Ollama', live: 'testOllama',
    tip: 'Connection + the model list + a real "What is 2 + 3?" inference test.' },
  { label: '🔌 Test Cloud API', live: 'testCloud',
    tip: 'The /models pre-flight (warn-only) + a "Say hello" chat ping — the wire format follows the URL.' },
  { label: '🦙 Test llama.cpp', live: 'testLlama',
    tip: 'The /props probe + a real one-token chat ping through the same OpenAI-compatible path the batch uses.' },
  { sep: true },
  { label: '🪝 Recall hooks (dry-run)', tip: 'The linking tools land with their own task (recall hooks in dry-run — diff only).' },
  { label: '🔗 Build link suggestions', tip: 'The linking tools land with their own task (suggest mode — the one write is the Suggestions note).' },
  { sep: true },
  { label: '✅ Validate Vault', tip: 'The vault validation door lands with the linking-tools task.' },
  { label: '🔍 Verify Vault', tip: 'The vault verification door lands with the linking-tools task.' },
  { label: '📁 Recategorize Notes', tip: 'The recategorize door lands with the linking-tools task.' },
  { sep: true },
  { label: '✅ Verify All Processed', tip: 'The bot-queue tools land with the queue-tools task.' },
  { label: '📋 Export All Links', tip: 'The bot-queue tools land with the queue-tools task.' },
  { label: '🔄 Retry Failed', tip: 'The bot-queue tools land with the queue-tools task (the retry queue rides the Dashboard page\'s undo door today).' },
  { label: '🚫 View 404 Quarantine', live: 'dashboard', tip: 'The confirmed-dead list — the Dashboard page\'s quarantine manager (Settings ▸ Dashboard).' },
  { label: '🔁 Retry _review backlog', live: 'retryReview', tip: 'Re-fetch the app-owned fetch-failed placeholders under _review (v0.41\'s browser-grade presentation).' },
  { label: '🔁 Retry the table\'s \' - \' rows', live: 'retryMaster', tip: 'The master table\'s " - " rows — valid links whose fetches failed — fetched again before "everything is up to date" may be said.' },
  { label: '🪦 Decommission dead links', tip: 'The graveyard procedure lands with the master-table tools task (the picker writes the same Status cells in-app).' },
  { label: '🖐 Hand-deliver walled links', tip: 'The Chrome doors land with the Chrome-doors GUI task (the engines live — internal/chrometabs).' },
  { label: '🖐 Scrape hand rows via Chrome', tip: 'The Chrome doors land with the Chrome-doors GUI task.' },
  { label: '🤖 Chrome tab-retry failed links', tip: 'The Chrome doors land with the Chrome-doors GUI task.' },
  { label: '🪄 Attach to my Chrome', tip: 'The Chrome doors land with the Chrome-doors GUI task (restart Chrome WITH its DevTools port open).' },
  { label: '📤 Export Backup ZIP', tip: 'The export-zip engine rides the packaging phase (the recorded plan — spec-gui-pages-b\'s dated note).' },
  { sep: true },
  { label: '📊 Open Dashboard', live: 'dashboard', tip: 'Open the website dashboard (the Worker URL from Settings ▸ Backup).' },
  { label: '⚙️ Settings', submenu: [{ label: '🌙 Dark Mode', live: 'dark',
      tip: 'The plum-night theme — a display preference, saved like every setting.' }] },
];

let moreOpen = false;

function closeMoreMenu() {
  moreOpen = false;
  const m = document.getElementById('more-menu');
  if (m) m.remove();
}

function openMoreMenu(anchor) {
  closeMoreMenu();
  moreOpen = true;
  const menu = el('div', 'more-menu');
  menu.id = 'more-menu';
  menu.setAttribute('role', 'menu');
  for (const act of MORE_ACTIONS) {
    if (act.sep) {
      menu.appendChild(el('div', 'more-sep'));
      continue;
    }
    if (act.submenu) {
      const item = el('button', 'more-item more-item-parent', act.label + ' ▸');
      item.type = 'button';
      item.setAttribute('role', 'menuitem');
      item.setAttribute('aria-haspopup', 'true');
      item.title = 'The display preference lives here (v31.1 — not a batch action).';
      const sub = el('div', 'more-submenu');
      for (const subAct of act.submenu) {
        const si = el('button', 'more-item' + (subAct.live ? '' : ' more-item-soon'), subAct.label);
        si.type = 'button';
        si.setAttribute('role', 'menuitem');
        si.title = subAct.tip || '';
        if (!subAct.live) si.disabled = true;
        si.addEventListener('click', (e) => {
          e.stopPropagation();
          closeMoreMenu();
          runMoreAction(subAct);
        });
        sub.appendChild(si);
      }
      item.appendChild(sub);
      item.addEventListener('click', () => { sub.classList.toggle('open'); });
      menu.appendChild(item);
      continue;
    }
    const item = el('button', 'more-item' + (act.live ? '' : ' more-item-soon'), act.label);
    item.type = 'button';
    item.setAttribute('role', 'menuitem');
    item.title = act.tip || '';
    if (!act.live) item.disabled = true;
    item.addEventListener('click', () => { closeMoreMenu(); runMoreAction(act); });
    menu.appendChild(item);
  }
  document.getElementById('app').appendChild(menu);
  const r = anchor.getBoundingClientRect();
  menu.style.top = (r.bottom + 6) + 'px';
  menu.style.left = Math.max(8, r.right - menu.offsetWidth) + 'px';
  setTimeout(() => {
    const off = (e) => {
      if (!menu.contains(e.target) && e.target !== anchor) {
        closeMoreMenu();
        document.removeEventListener('mousedown', off);
      }
    };
    document.addEventListener('mousedown', off);
  }, 0);
}

async function runMoreAction(act) {
  if (act.live === true) {
    // Test Connection (all systems) — the battery modal.
    if (testBtn) testBtn.click();
    return;
  }
  if (act.live === 'testTgGh' || act.live === 'testProxy' ||
      act.live === 'testOllama' || act.live === 'testCloud' ||
      act.live === 'testLlama') {
    // T-79a: the five per-test doors — the engines stream their own
    // line grammar onto the desk while the promise waits; a verdict
    // box (title set) renders in the same themed box the dashboard
    // door uses; the log-only doors return an empty title.
    const fn = act.live === 'testTgGh' ? MoreTestTelegramGitHub
      : act.live === 'testProxy' ? MoreTestProxy
      : act.live === 'testOllama' ? MoreTestOllama
      : act.live === 'testCloud' ? MoreTestCloud
      : MoreTestLlamaCpp;
    const res = await fn().catch(() => null);
    if (res && res.title) {
      themedBox({ title: res.title, body: res.body },
        [['OK', 'primary', () => closeOverlay()]]);
    }
    return;
  }
  if (act.live === 'dashboard') {
    const res = await MoreOpenDashboard().catch(() => null);
    if (res && res.title) {
      themedBox({ title: res.title, body: res.body },
        [['OK', 'primary', () => closeOverlay()]]);
    }
    return;
  }
  if (act.live === 'retryReview') {
    await MoreRetryBacklog().catch(() => {});
    return;
  }
  if (act.live === 'retryMaster') {
    await MoreRetryMaster().catch(() => {});
    return;
  }
  if (act.live === 'dark') {
    const on = !document.documentElement.dataset.theme ||
      document.documentElement.dataset.theme !== 'dark';
    applyTheme(on ? 'dark' : 'light');
    await MoreDarkMode(on).catch(() => {});
    return;
  }
}

function applyTheme(mode) {
  document.documentElement.dataset.theme = mode;
}

// the More button itself (ui.py's more_btn — "More ▾").
let moreBtnEl = null;
const ctaBand = document.querySelector('.cta');
if (ctaBand) {
  const moreBtn = btn('More ▾', 'secondary',
    'Tests, verification, export, retry and settings', (e) => {
    e.stopPropagation();
    if (moreOpen) closeMoreMenu();
    else openMoreMenu(moreBtn);
  });
  ctaBand.appendChild(moreBtn);
  moreBtnEl = moreBtn;
}

// the evidence probe (T-79a's gate leg — the T-70 law applied to the
// More menu): open the menu over its own anchor so the screenshot can
// prove the five per-test doors LIVE (clickable, no 'soon' wash).
EventsOn('autorun:open-more', async (v) => {
  if (!moreBtnEl) return;
  openMoreMenu(moreBtnEl);
  const door = String(v || '');
  if (door) {
    // the click leg: run ONE door end-to-end (the click → the
    // binding → the engine → the bus → the desk). 'ollama' is the
    // hermetic pick — a box without an Ollama server refuses the
    // loopback probe instantly and the desk speaks the reference's
    // own honest failure lines.
    await new Promise((res) => setTimeout(res, 400));
    const item = [...document.querySelectorAll('#more-menu .more-item')]
      .find((n) => n.textContent.includes('Test Ollama'));
    if (item && !item.disabled) item.click();
  }
});

// ---------------------------------------------------------------------------
// THE BOOT PAINT — the persisted dark_mode + the hero's first render.
// ---------------------------------------------------------------------------
ChromeBoot().then((boot) => {
  if (boot && boot.dark_mode) applyTheme('dark');
}).catch(() => {});
HeroState().then((payload) => { renderHero(payload); }).catch(() => {});
