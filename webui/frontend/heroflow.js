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
//     live doors wired (Test Connection, the five per-test doors, the
//     five linking-tools doors — recall hooks' dry-run, link
//     suggestions, Validate/Verify Vault, Recategorize — the three
//     bot-queue tool doors + the manual-resolve dialog, the four
//     Chrome doors, the decommission picker (T-79e — the tail is
//     EMPTY now, Phase 6's close), Open Dashboard, Dark Mode); the
//     📤 Export Backup ZIP row alone stays honestly disabled (T-80's
//     standing assignment — the packaging composition);
//   · THE BATCH CHIME: the batch-finished crossing's chime payload
//     plays through <audio> (the QSoundEffect twin — the exact batch
//     gate lives in Go).
import { HeroClick, HeroAnswerOffline, HeroRunQueue, HeroState, TestConnection,
         MoreOpenDashboard, MoreDarkMode, MoreRetryBacklog, MoreRetryMaster,
         MoreTestTelegramGitHub, MoreTestProxy, MoreTestOllama,
         MoreTestCloud, MoreTestLlamaCpp,
         MoreValidateVault, MoreVerifyVault, MoreRecallHooksDryRun,
         MoreBuildLinkSuggestions, MoreRecategorizeOpen, MoreRecategorizeApply,
         MoreExportAllBotLinks, MoreVerifyAllBotLinks, MoreRetryFailed,
         MoreResolveMarkProcessed, MoreResolveDecommission,
         MoreResolveMarkAll, MoreResolveProcess,
         MoreHandDeliverWalled, MoreHandDeliverQueue, MoreOpenHandFolder,
         MoreHandRowsDeliver, MoreChromeTabRetry, MoreChromeTabRetryRun,
         MoreAttachMyChrome,
         MoreDecommissionScan, MoreDecommissionBury, MoreOpenDecommissionTable,
         ChromeBoot, ChromeOfferAnswer } from './wailsjs/go/main/HeroUI.js';
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
function themedBox(box, buttons, opts) {
  const width = (opts && opts.width) || 460;
  const b = openOverlay(width);
  // T-79b: the long-body law — a scroll flag pins the action buttons
  // OUTSIDE the scrolling body (the verify report's own shape).
  if (opts && opts.scroll) b.classList.add('linkdoor-box');
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
  return b;
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
// T-79d: the Chrome doors' dialogs — the reference's own picker shapes
// (_pick_hand_links_dialog 1095-1191 + _chrome_tab_retry_dialog
// 1266-1349): checkable lists (all ticked by default), the wall/error
// riding under each URL, the folder label, and the buttons pinned
// OUTSIDE the scrolling body (the long-body law).
// ---------------------------------------------------------------------------

// pickRowEl builds one checkable row: the URL with the wall/error
// under it (the reference's "url   —   wall" shape, 110-rune capped,
// the full text on the tooltip).
function pickRowEl(url, note) {
  const shown = note ? (note.length > 110 ? note.slice(0, 110) : note) : '';
  const row = el('label', 'form-check-label chrome-pick-row');
  const box = el('input', 'form-check');
  box.type = 'checkbox';
  box.checked = true;
  row.appendChild(box);
  const text = el('span', 'chrome-pick-text');
  const name = el('span', 'chrome-pick-url', url);
  text.appendChild(name);
  if (shown) text.appendChild(el('span', 'chrome-pick-note', '   —   ' + shown));
  row.title = url + (shown ? '   —   ' + shown : '');
  row.appendChild(text);
  return { row, box };
}

// openHandDeliverDialog — 🖐 Hand-deliver walled links' picker
// (processing_control.py:1095): tick the ones worth saving; 🖐 Queue
// & open in Chrome runs the queue door with the checked rows; Cancel
// speaks the reference's cancel line (an empty list).
function openHandDeliverDialog(model) {
  const box = openOverlay(680);
  box.classList.add('linkdoor-box');
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', 'Hand-deliver walled links');
  const header = el('div', 'chrome-pick-header');
  header.appendChild(el('span', 'chrome-glyph', '🖐'));
  header.appendChild(el('span', 'card-heading', 'Hand-deliver walled links'));
  box.appendChild(header);
  const body = el('div', 'dialog-body chrome-pick-body');
  body.appendChild(el('p', 'dialog-message',
    "These links answered every machine door with a wall (403 / bot defense / TLS fingerprint — the wall rides under each URL). Tick the ones worth saving: each opens in your real Chrome, is queued with a suggested filename, and the page you save (Ctrl+S, 'Webpage, HTML Only') becomes a REAL fetch on the next batch. Nothing is retired — the links keep waiting until their page is delivered."));
  const list = el('div', 'chrome-pick-list');
  const rows = [];
  for (const link of model.links) {
    const r = pickRowEl(link.url, link.wall);
    rows.push({ r, link });
    list.appendChild(r.row);
  }
  body.appendChild(list);
  body.appendChild(el('p', 'chrome-pick-folder',
    'Deliver into: ' + model.folder + ' (the README regenerates there with every link\'s suggested filename)'));
  box.appendChild(body);
  const btnRow = el('div', 'dialog-buttons');
  btnRow.appendChild(btn('Open the folder', 'secondary',
    'Reveal the hand-delivered folder in your file manager — saved pages land there.',
    async () => { await MoreOpenHandFolder().catch(() => {}); }));
  const spacer = el('span', 'chrome-pick-spacer');
  btnRow.appendChild(spacer);
  btnRow.appendChild(btn('Cancel', 'secondary',
    'Nothing is queued — the walled links keep waiting.',
    async () => { closeOverlay(); await MoreHandDeliverQueue([]).catch(() => {}); }));
  btnRow.appendChild(btn('🖐 Queue & open in Chrome', 'primary',
    "Queue every ticked link for hand-delivery and open it in your real Chrome — the next SYNC batch consumes every delivered page as a REAL fetch.",
    async () => {
      const picked = rows.filter((x) => x.r.box.checked).map((x) => x.link);
      closeOverlay();
      const res = await MoreHandDeliverQueue(picked).catch(() => null);
      if (res && res.title) {
        themedBox({ title: res.title, body: res.body },
          [['OK', 'primary', () => closeOverlay()]], { width: 560, scroll: true });
      }
    }));
  box.appendChild(btnRow);
}

// openChromeRetryDialog — 🤖 Chrome tab-retry's offer modal
// (processing_control.py:1266): the failed links with their last error
// under each URL, the question, and the two answers. ONE dialog,
// TWO answer doors (the reference's own two call sites): the manual
// More door answers through MoreChromeTabRetryRun (its decline speaks
// the CANCELLED line), the end-of-run offer through ChromeOfferAnswer
// (its decline speaks the DECLINED line — the reference's 1258-1261).
function openChromeRetryDialog(links, answer) {
  const box = openOverlay(720);
  box.classList.add('linkdoor-box');
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', 'Retry in your real Chrome?');
  const header = el('div', 'chrome-pick-header');
  header.appendChild(el('span', 'chrome-glyph', '🤖'));
  header.appendChild(el('span', 'card-heading',
    links.length + " link(s) couldn't generate content"));
  box.appendChild(header);
  const body = el('div', 'dialog-body chrome-pick-body');
  body.appendChild(el('p', 'dialog-message',
    'The machine fetch failed for ' + links.length + ' link(s) this run — no content was generated, each one\'s last error rides under its URL (they are the " - " rows in the master table).\n\nWant to retry them in your real Google Chrome? GitCurator opens a fresh Chrome window (your own Chrome, a new session — your running window is never touched), starts one tab per link, waits for every page to load, and takes the content from the live pages itself — no manual saving. Links that still fail keep waiting; their automatic retries continue.'));
  const list = el('div', 'chrome-pick-list');
  const rows = [];
  for (const link of links) {
    const r = pickRowEl(link.url, link.error);
    rows.push({ r, link });
    list.appendChild(r.row);
  }
  body.appendChild(list);
  box.appendChild(body);
  const btnRow = el('div', 'dialog-buttons');
  const spacer = el('span', 'chrome-pick-spacer');
  btnRow.appendChild(spacer);
  btnRow.appendChild(btn('Not now', 'secondary',
    'The failed links keep waiting — their automatic retries continue, and the master table\'s Status cell decides their fate.',
    async () => { closeOverlay(); await answer([]).catch(() => {}); }));
  btnRow.appendChild(btn('🤖 Retry in Chrome now', 'primary',
    'Opens your own Google Chrome (a fresh session), one tab per link, and takes the content from the live pages.',
    async () => {
      const picked = rows.filter((x) => x.r.box.checked).map((x) => x.link);
      closeOverlay();
      await answer(picked).catch(() => {});
    }));
  box.appendChild(btnRow);
}

// the end-of-run offer (the chrome:offer event — the batch-scorecard
// fanfare's automatic modal, T-79f): ONE listener, registered once at
// boot (the ONE-OWNER-PER-SURFACE law). The SAME dialog the manual
// door renders; the answers ride ChromeOfferAnswer (the DECLINED
// line's door — the offered set marks only the picked links).
EventsOn('chrome:offer', (payload) => {
  if (!payload || !payload.links || !payload.links.length) return;
  openChromeRetryDialog(payload.links, (picked) => ChromeOfferAnswer(picked));
});

// openDecommissionDialog — 🪦 Decommission dead links' picker
// (processing_control.py:910): tick the GENUINELY dead ones (all rows
// start UNCHECKED — the reference's setCheckState(Unchecked) law — a
// burial is never a default); 🪦 Decommission selected writes the
// SAME Status cells + enforces immediately; Cancel speaks the
// reference's cancel line (an empty list); Open the table reveals the
// ledger in the file manager.
function openDecommissionDialog(model) {
  const box = openOverlay(680);
  box.classList.add('linkdoor-box');
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', 'Decommission dead links');
  const header = el('div', 'chrome-pick-header');
  header.appendChild(el('span', 'chrome-glyph', '🪦'));
  header.appendChild(el('span', 'card-heading', 'Decommission dead links'));
  box.appendChild(header);
  const body = el('div', 'dialog-body chrome-pick-body');
  body.appendChild(el('p', 'dialog-message',
    "Tick the links that are GENUINELY dead — a real 404, a lost page, an abandoned domain. A buried link is never fetched, never retried, never noticed again; its _review placeholder is swept. (Unticked links keep waiting — ♻️ revived in the table brings a buried link back.)"));
  const list = el('div', 'chrome-pick-list');
  const boxes = [];
  for (const url of model.links) {
    const row = el('div', 'chrome-pick-row');
    const box2 = el('input');
    box2.type = 'checkbox';
    // the reference's law: every row starts UNCHECKED (a burial is
    // never a default — the owner's explicit hand).
    box2.checked = false;
    row.appendChild(box2);
    const text = el('span', 'chrome-pick-text');
    text.appendChild(el('span', 'chrome-pick-url', url));
    row.title = url;
    row.appendChild(text);
    boxes.push({ box: box2, url });
    list.appendChild(row);
  }
  body.appendChild(list);
  body.appendChild(el('p', 'chrome-pick-folder', 'The ledger: ' + model.table));
  box.appendChild(body);
  const btnRow = el('div', 'dialog-buttons');
  btnRow.appendChild(btn('Open the table', 'secondary',
    'Reveal the graveyard table in your file manager — hand-edit the Status column in Obsidian (🪦 dead / ♻️ revived).',
    async () => { await MoreOpenDecommissionTable().catch(() => {}); }));
  const spacer = el('span', 'chrome-pick-spacer');
  btnRow.appendChild(spacer);
  btnRow.appendChild(btn('Cancel', 'secondary',
    'Nothing is buried — the candidate rows wait in the table.',
    async () => { closeOverlay(); await MoreDecommissionBury([]).catch(() => {}); }));
  btnRow.appendChild(btn('🪦 Decommission selected', 'danger',
    'Write 🪦 dead into the chosen rows\' Status cells (the same cells a hand-edit writes) and enforce the burials now: dismissed, retry-queue rows dropped, placeholders swept.',
    async () => {
      const picked = boxes.filter((x) => x.box.checked).map((x) => x.url);
      closeOverlay();
      const res = await MoreDecommissionBury(picked).catch(() => null);
      if (res && res.title) {
        themedBox({ title: res.title, body: res.body },
          [['OK', 'primary', () => closeOverlay()]], { width: 560, scroll: true });
      }
    }));
  box.appendChild(btnRow);
}

// the delivery's failure modal (the chrome:failed event — _on_delivery_
// failed's twin): ONE listener, registered once at boot (the
// ONE-OWNER-PER-SURFACE law).
EventsOn('chrome:failed', (payload) => {
  if (!payload || !payload.title) return;
  themedBox({ title: payload.title, body: payload.body },
    [['OK', 'primary', () => closeOverlay()]], { width: 560, scroll: true });
});

// ---------------------------------------------------------------------------
// THE MORE MENU — the reference's own action list, verbatim (ui.py
// 1849-1946). The live doors are wired (the Path-B tail T-79a…T-79e
// is COMPLETE — every door family landed); the 📤 Export Backup ZIP
// row alone renders HONESTLY DISABLED with its landing note (T-80's
// packaging composition — the standing assignment, spec-gui-pages-b's
// dated note).
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
  { label: '🪝 Recall hooks (dry-run)', live: 'recall',
    tip: 'The Phase-6 step-1 tool in its default dry-run mode — nothing is written, the diff lands in the log + app/reports/recall/. Approve samples and apply from the command line (the SPEC flow).' },
  { label: '🔗 Build link suggestions', live: 'links',
    tip: 'Embeddings → neighbors → LLM confirm → the Suggestions note under the manual vault\'s Library/ (its only write). Tick boxes there, then run the tool with --collect on the command line to apply.' },
  { sep: true },
  { label: '✅ Validate Vault', live: 'validate',
    tip: 'Validate the selected vault path (fast, inline) — the .obsidian check.' },
  { label: '🔍 Verify Vault', live: 'verify',
    tip: 'The manifest verification report — every GitHub link\'s note on disk, every non-GitHub link in the inbox tables; failed repos join the retry queue.' },
  { label: '📁 Recategorize Notes', live: 'recategorize',
    tip: 'A table of the vault\'s notes — pick new categories in bulk; Apply rewrites the frontmatter and moves the files (locked notes are never touched).' },
  { sep: true },
  { label: '✅ Verify All Processed', live: 'verifyAll',
    tip: 'Every GitHub link the queue holds, checked against the vault + the settled ledger — the report box + the permanent _verification_report file; missing links open the manual-resolve dialog.' },
  { label: '📋 Export All Links', live: 'exportAll',
    tip: 'All the queue\'s links saved to a file in the vault for manual verification — compare what the app found with what you actually forwarded to the bot.' },
  { label: '🔄 Retry Failed', live: 'retryFailed',
    tip: 'The retry queue\'s unresolved links re-processed through the full batch (settled repos are never re-fetched); a retry batch never clears the bot queue.' },
  { label: '🚫 View 404 Quarantine', live: 'dashboard', tip: 'The confirmed-dead list — the Dashboard page\'s quarantine manager (Settings ▸ Dashboard).' },
  { label: '🔁 Retry _review backlog', live: 'retryReview', tip: 'Re-fetch the app-owned fetch-failed placeholders under _review (v0.41\'s browser-grade presentation).' },
  { label: '🔁 Retry the table\'s \' - \' rows', live: 'retryMaster', tip: 'The master table\'s " - " rows — valid links whose fetches failed — fetched again before "everything is up to date" may be said.' },
  { label: '🪦 Decommission dead links', live: 'decommission',
    tip: 'The graveyard procedure\'s in-app picker — tick the genuinely dead links (a real 404, a lost page, an abandoned domain); the picker writes the same Status cells a hand-edit writes, then enforces the burials immediately (dismissed + retry-queue rows dropped + placeholders swept). ♻️ revived on a row brings a link back.' },
  { label: '🖐 Hand-deliver walled links', live: 'handDeliver',
    tip: "The links every machine door failed to open (403 / bot defense / TLS fingerprint) — the picker queues the chosen ones for your real Chrome; the saved page becomes a real fetch on the next batch." },
  { label: '🖐 Scrape hand rows via Chrome', live: 'handRows',
    tip: "The master table's 🖐 hand rows scraped in your real Chrome NOW — one tab per link, the live pages become real fetches." },
  { label: '🤖 Chrome tab-retry failed links', live: 'chromeRetry',
    tip: 'Every link waiting in the websites retry queue re-fetched through your own Chrome — one tab per link, the live DOM taken as the content.' },
  { label: '🪄 Attach to my Chrome', live: 'attach',
    tip: "Restart your Chrome WITH its DevTools port open so the fifth door drives YOUR instance — your profile, your cookies, your logins (Windows; the .bat + instructions)." },
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
  if (act.live === 'validate' || act.live === 'recall' ||
      act.live === 'links') {
    // T-79b: the log-only linking doors — the engines stream their own
    // lines (the tool doors' captured tail lands at completion).
    const fn = act.live === 'validate' ? MoreValidateVault
      : act.live === 'recall' ? MoreRecallHooksDryRun
      : MoreBuildLinkSuggestions;
    await fn().catch(() => null);
    return;
  }
  if (act.live === 'verify') {
    // T-79b: Verify Vault — the report box (the long-body law: the
    // buttons stay pinned under the scrolling report).
    const res = await MoreVerifyVault().catch(() => null);
    if (res && res.title) {
      themedBox({ title: res.title, body: res.body },
        [['OK', 'primary', () => closeOverlay()]], { width: 560, scroll: true });
    }
    return;
  }
  if (act.live === 'recategorize') {
    // T-79b: the two-door dialog — the open walk, then the table
    // modal; Apply rides the apply door.
    const model = await MoreRecategorizeOpen().catch(() => null);
    if (!model) return;
    if (model.boxTitle) {
      themedBox({ title: model.boxTitle, body: model.boxBody },
        [['OK', 'primary', () => closeOverlay()]]);
      return;
    }
    if (model.notes && model.notes.length) openRecategorizeDialog(model);
    return;
  }
  if (act.live === 'verifyAll') {
    // T-79c: Verify All Processed — the report box (the long-body
    // law); missing links open the manual-resolve dialog through the
    // box's own Resolve button (the one-overlay-at-a-time law).
    const res = await MoreVerifyAllBotLinks().catch(() => null);
    if (res && res.title) {
      const buttons = [];
      if (res.missing && res.missing.length) {
        buttons.push(['Resolve Missing Links', 'primary',
          () => { closeOverlay(); openBotResolveDialog(res.missing); }]);
        buttons.push(['OK', 'secondary', () => closeOverlay()]);
      } else {
        buttons.push(['OK', 'primary', () => closeOverlay()]);
      }
      themedBox({ title: res.title, body: res.body }, buttons,
        { width: 560, scroll: true });
    }
    return;
  }
  if (act.live === 'exportAll') {
    // T-79c: Export All Links — the 📋 Links Exported box (a refusal
    // or failure spoke on the desk).
    const res = await MoreExportAllBotLinks().catch(() => null);
    if (res && res.title) {
      themedBox({ title: res.title, body: res.body },
        [['OK', 'primary', () => closeOverlay()]]);
    }
    return;
  }
  if (act.live === 'retryFailed') {
    // T-79c: Retry Failed — the two-phase confirm (a >10 kept list
    // asks; Continue calls the confirmed pass; the vault-error and
    // empty-queue boxes render the verdict; a ≤10 list already ran —
    // the batch's story is on the desk).
    const res = await MoreRetryFailed(false).catch(() => null);
    if (res && res.confirm) {
      themedBox({ title: res.confirm.title, body: res.confirm.body }, [
        ['Continue', 'primary',
          async () => { closeOverlay(); await MoreRetryFailed(true).catch(() => {}); }],
        ['Cancel', 'secondary', () => closeOverlay()],
      ]);
      return;
    }
    if (res && res.box) {
      themedBox({ title: res.box.title, body: res.box.body },
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
  if (act.live === 'decommission') {
    // T-79e: 🪦 Decommission dead links — the guards'/empty boxes, or
    // the picker dialog (the reference's own words, every row
    // UNCHECKED — a burial is never a default); the answer doors
    // write the same Status cells + enforce immediately.
    const model = await MoreDecommissionScan().catch(() => null);
    if (!model) return;
    if (model.box) {
      themedBox({ title: model.box.title, body: model.box.body },
        [['OK', 'primary', () => closeOverlay()]]);
      return;
    }
    if (model.links && model.links.length) openDecommissionDialog(model);
    return;
  }
  if (act.live === 'handDeliver') {
    // T-79d: 🖐 Hand-deliver walled links — the picker dialog (the
    // reference's own words; the checked rows queue + open in Chrome).
    const model = await MoreHandDeliverWalled().catch(() => null);
    if (!model) return;
    if (model.box) {
      themedBox({ title: model.box.title, body: model.box.body },
        [['OK', 'primary', () => closeOverlay()]]);
      return;
    }
    if (model.links && model.links.length) openHandDeliverDialog(model);
    return;
  }
  if (act.live === 'handRows') {
    // T-79d: 🖐 Scrape hand rows via Chrome — the busy/empty boxes,
    // the >10 confirm (Continue calls the confirmed pass), or the
    // delivery started (the desk speaks its own story).
    const res = await MoreHandRowsDeliver(false).catch(() => null);
    if (!res) return;
    if (res.box) {
      themedBox({ title: res.box.title, body: res.box.body },
        [['OK', 'primary', () => closeOverlay()]]);
      return;
    }
    if (res.confirm) {
      themedBox({ title: res.confirm.title, body: res.confirm.body }, [
        ['Continue', 'primary',
          async () => { closeOverlay(); await MoreHandRowsDeliver(true).catch(() => {}); }],
        ['Cancel', 'secondary', () => closeOverlay()],
      ]);
      return;
    }
    return;
  }
  if (act.live === 'chromeRetry') {
    // T-79d: 🤖 Chrome tab-retry failed links — the offer dialog (the
    // reference's own words + the two answers); the MANUAL door's
    // answers ride MoreChromeTabRetryRun (the CANCELLED line).
    const res = await MoreChromeTabRetry().catch(() => null);
    if (!res) return;
    if (res.box) {
      themedBox({ title: res.box.title, body: res.box.body },
        [['OK', 'primary', () => closeOverlay()]]);
      return;
    }
    if (res.links && res.links.length) {
      openChromeRetryDialog(res.links,
        (picked) => MoreChromeTabRetryRun(picked));
    }
    return;
  }
  if (act.live === 'attach') {
    // T-79d: 🪄 Attach to my Chrome — the ladder's boxes; the ONE
    // explicit OK (Windows) rides the confirm (the default answer
    // is No — his browser is never touched on a doubt).
    const res = await MoreAttachMyChrome(false).catch(() => null);
    if (!res) return;
    if (res.box) {
      themedBox({ title: res.box.title, body: res.box.body },
        [['OK', 'primary', () => closeOverlay()]], { width: 560 });
      return;
    }
    if (res.confirm) {
      const overlay = themedBox({ title: res.confirm.title, body: res.confirm.body }, [
        ['No', 'secondary', () => closeOverlay()],
        ['Yes', 'primary',
          async () => { closeOverlay(); await MoreAttachMyChrome(true).catch(() => {}); }],
      ], { width: 560 });
      // the reference's default-button law: No is the default.
      const noBtn = overlay ? overlay.querySelector('.btn-secondary') : null;
      if (noBtn) noBtn.focus();
      return;
    }
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

// openRecategorizeDialog — the reference's table dialog (dashboard.py
// 548-660): the notes table (Note | Current Category | New Category),
// per-row selects over the category list (the current selected when
// present), Apply + Cancel right-aligned — the buttons pinned OUTSIDE
// the scrolling body (the long-body law).
function openRecategorizeDialog(model) {
  const box = openOverlay(640);
  box.classList.add('linkdoor-box');
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', 'Recategorize Notes');
  box.appendChild(el('div', 'card-heading',
    '📁 Recategorize Notes (' + model.notes.length + ' notes)'));
  const body = el('div', 'dialog-body recat-body');
  const table = el('table', 'recat-table');
  const thead = el('thead');
  const headRow = el('tr');
  for (const h of ['Note', 'Current Category', 'New Category']) {
    headRow.appendChild(el('th', '', h));
  }
  thead.appendChild(headRow);
  table.appendChild(thead);
  const tbody = el('tbody');
  for (const note of model.notes) {
    const tr = el('tr');
    const name = el('td', 'recat-name', note.name);
    name.title = note.path;
    tr.appendChild(name);
    tr.appendChild(el('td', 'recat-cat', note.category));
    const td = el('td');
    const select = el('select', 'recat-select');
    select.title = 'Pick the new category — Apply rewrites the frontmatter and moves the note to its category folder.';
    for (const cat of (model.categories || [])) {
      const opt = el('option', '', cat);
      opt.value = cat;
      if (cat === note.category) opt.selected = true;
      select.appendChild(opt);
    }
    td.appendChild(select);
    tr.appendChild(td);
    tbody.appendChild(tr);
  }
  table.appendChild(tbody);
  body.appendChild(table);
  const btnRow = el('div', 'dialog-buttons');
  btnRow.appendChild(btn('Cancel', 'secondary',
    'Leave the notes where they are — nothing is rewritten, nothing moved.',
    () => closeOverlay()));
  btnRow.appendChild(btn('Apply Changes', 'primary',
    'Rewrite the category frontmatter of the changed rows and move those notes to their category folders (locked notes are never touched).',
    async () => {
      const rows = [...tbody.querySelectorAll('tr')];
      const changes = rows.map((tr, i) => ({
        path: model.notes[i].path,
        oldCategory: model.notes[i].category,
        category: tr.querySelector('select').value,
      }));
      closeOverlay();
      await MoreRecategorizeApply(changes).catch(() => null);
    }));
  box.append(body, btnRow);
}

// openBotResolveDialog — the verify door's manual-resolve dialog
// (bot_queue.py:1368-1496): the header (⏳ — pending work, not a
// failure), the missing-URLs list (click a row to select it — the
// QListWidget twin; the actions no-op with nothing selected, the
// reference's own silent return), and the action row — ONE filled
// primary (Process Selected), outlined secondaries (Mark as
// Processed / Mark ALL), ONE filled danger (Decommission) — plus
// Close; the buttons pinned OUTSIDE the scrolling body.
function openBotResolveDialog(missing) {
  let list = missing.slice();
  const box = openOverlay(640);
  box.classList.add('linkdoor-box');
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', 'Manual Resolve — Missing Links');
  box.appendChild(el('div', 'card-heading',
    'Manual Resolve — Missing Links (' + list.length + ')'));
  const body = el('div', 'dialog-body resolve-body');
  body.appendChild(el('p', 'dialog-message',
    '⏳ ' + list.length + ' GitHub link(s) are missing from the vault.\n' +
    'For each URL, choose an action:'));
  const ul = el('ul', 'resolve-list');
  ul.setAttribute('role', 'listbox');
  let selected = null;
  const rowEls = new Map();
  const paint = () => {
    for (const [url, li] of rowEls) {
      li.classList.toggle('selected', url === selected);
      li.setAttribute('aria-selected', url === selected ? 'true' : 'false');
    }
  };
  for (const url of list) {
    const li = el('li', 'resolve-row', url);
    li.title = url;
    li.setAttribute('role', 'option');
    li.tabIndex = 0;
    li.addEventListener('click', () => { selected = url; paint(); });
    li.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); selected = url; paint(); }
    });
    rowEls.set(url, li);
    ul.appendChild(li);
  }
  body.appendChild(ul);
  const removeRow = (url) => {
    const li = rowEls.get(url);
    if (li) { li.remove(); rowEls.delete(url); }
    list = list.filter((u) => u !== url);
    selected = null;
    paint();
  };
  const btnRow = el('div', 'dialog-buttons');
  btnRow.appendChild(btn('Process Selected', 'primary',
    'Fetch + curate the selected URL into the vault — the batch runs now (bot source).',
    async () => {
      if (!selected) return; // the reference's silent no-selection return
      closeOverlay();
      await MoreResolveProcess([selected]).catch(() => null);
    }));
  btnRow.appendChild(btn('Mark as Processed', 'secondary',
    "The owner's word: this repo is addressed — recorded processed + settled, never fetched or counted again (♻️ Reset 404 Quarantine is the door back).",
    async () => {
      if (!selected) return;
      const url = selected;
      const res = await MoreResolveMarkProcessed(url, list).catch(() => null);
      removeRow(url);
      if (res && res.allResolved) closeOverlay();
    }));
  btnRow.appendChild(btn('Decommission', 'danger',
    'Bury the selected link — the 404 quarantine + the settled ledger, never fetched again.',
    async () => {
      if (!selected) return;
      const url = selected;
      const res = await MoreResolveDecommission(url, list).catch(() => null);
      removeRow(url);
      if (res && res.allResolved) closeOverlay();
    }));
  btnRow.appendChild(btn('Mark ALL as Processed', 'secondary',
    'Every missing URL on the owner\'s word — processed + settled in one pass.',
    async () => {
      await MoreResolveMarkAll(list).catch(() => null);
      closeOverlay();
    }));
  btnRow.appendChild(btn('Close', 'secondary',
    'Leave the links pending — nothing is written.', () => closeOverlay()));
  box.append(body, btnRow);
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
    // own honest failure lines. 'recall' (T-79b) clicks the tool
    // door: with no vault configured the runner speaks its whole
    // grammar — the opening, the captured refusal, the exit-code
    // finish line. 'export' (T-79c) clicks the bot-queue export
    // door: with no bot username configured the door speaks the
    // reference's own username-refusal line. 'hand' (T-79d) clicks
    // the fourth door's scan: with the gate's unconfigured fixture
    // (no Websites pipeline, no Websites vault), the door answers the
    // reference's own "Websites pipeline is off" guard box — the
    // collaborator law's honest refusal, no Chrome anywhere near
    // (a pure config read). 'grave' (T-79e) clicks the decommission
    // picker's scan: the same unconfigured fixture answers the same
    // guard box (the reference's own remedy line) — the door is a
    // pure config + file read, no network anywhere near.
    await new Promise((res) => setTimeout(res, 400));
    const pick = door === 'ollama' ? 'Test Ollama'
      : door === 'recall' ? 'Recall hooks (dry-run)'
      : door === 'export' ? 'Export All Links'
      : door === 'hand' ? 'Hand-deliver walled links'
      : door === 'grave' ? 'Decommission dead links' : '';
    if (pick) {
      const item = [...document.querySelectorAll('#more-menu .more-item')]
        .find((n) => n.textContent.includes(pick));
      if (item && !item.disabled) item.click();
    }
  }
});

// ---------------------------------------------------------------------------
// THE BOOT PAINT — the persisted dark_mode + the hero's first render.
// ---------------------------------------------------------------------------
ChromeBoot().then((boot) => {
  if (boot && boot.dark_mode) applyTheme('dark');
}).catch(() => {});
HeroState().then((payload) => { renderHero(payload); }).catch(() => {});
