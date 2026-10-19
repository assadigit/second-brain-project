// scanflow.js — T-76: the Scan flow's frontend half (the three UI
// legs' renderers — vault_pick_dialog.py + scan_plan_dialog.py +
// scan_result_dialog.py, the HTML twins over internal/app's pure
// models). The engine laws live in Go (scandialogs.go — the models,
// the picking semantics, the field tags mirror the Python dict keys:
// status_text, vault_label, low_confidence…); this file only renders
// and answers:
//   · THE VAULT PICKER — the three rows (nothing pre-checked; the two
//     machine vaults tickable, Personal read-only with its why), the
//     live Scan-button arithmetic, Cancel/Esc/✕ = the safe cancel;
//   · THE PLAN MODAL — the four section cards (the washes), the
//     master "All" boxes, the low-confidence pre-unchecked rows, THE
//     FRESH-CLICK DOOR (350 ms input swallow, both verdict buttons
//     born disabled — a fossil click is not consent), the answering
//     window's countdown, and the four-key selection riding ONLY a
//     confirmed verdict;
//   · THE SCORECARD — the landed counts (K of N in the warning tone),
//     the per-vault sections, the failed moves by name, the
//     recoverability footnote, the 12 s self-close.
// The Scan CTA (ui.py's scan_btn — its own machine, NOT the hero's)
// opens the picker; the flow's one door is ScanUI.StartScan (the
// guards → the start line → the job → the finished routing — the desk
// lines ride the bus underneath the whole way).
import { PickVaults, StartScan, ScanAnswer, ProbeAskSample, ProbeScorecardSample } from './wailsjs/go/main/ScanUI.js';
import { EventsOn } from './wailsjs/runtime/runtime.js';

// the reference's own numbers (the clocks are the law both sides read).
const ARM_MS = 350;          // _ARM_MS — the fresh-click door's beat
const AUTO_CLOSE_S = 12;     // AUTO_CLOSE_S — the scorecard's window

// ---------------------------------------------------------------------------
// tiny DOM helpers (the settings.js kit's local twins)
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

// elideMiddle is the HTML twin of Qt's ElideMiddle (the parity law:
// ONE line, elided MIDDLE, the full story on the tooltip): a char
// budget with the … in the middle — the honest browser-side shape.
function elideMiddle(text, max) {
  if (text.length <= max) return text;
  const keep = Math.max(max - 1, 3);
  const head = Math.ceil(keep / 2);
  const tail = keep - head;
  return text.slice(0, head) + '…' + text.slice(text.length - tail);
}

// one overlay at a time (the modal stack law: the flow's modals never
// stack — the picker closes before the job starts, the plan modal
// stands alone over the desk, the scorecard opens after the story
// ends).
let overlay = null;
function openOverlay(width) {
  closeOverlay();
  overlay = el('div', 'dialog-overlay scanflow-overlay');
  const box = el('div', 'dialog-box scanflow-box');
  if (width) box.style.minWidth = width + 'px';
  overlay.appendChild(box);
  overlay.addEventListener('keydown', (e) => {
    if (e.key === 'Escape') { e.stopPropagation(); onOverlayEscape(); }
  });
  document.getElementById('app').appendChild(overlay);
  return box;
}
function closeOverlay() {
  if (overlay && overlay.parentNode) overlay.remove();
  overlay = null;
}
let escapeHandler = null;
function onOverlayEscape() {
  if (escapeHandler) escapeHandler();
}

// ---------------------------------------------------------------------------
// THE VAULT PICKER (vault_pick_dialog.py's renderer).
// ---------------------------------------------------------------------------
const pickState = { boxes: new Map(), open: false };

async function openScanPicker() {
  const model = await PickVaults();
  if (!model) return;
  pickState.open = true;
  pickState.boxes = new Map();
  const box = openOverlay(560);
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', model.title);
  box.appendChild(el('div', 'card-heading', model.title));
  const body = el('div', 'dialog-body');
  const head = el('p', 'dialog-message pick-header', model.header);
  body.appendChild(head);

  const card = el('div', 'sync_card pick-card');
  for (const row of model.rows || []) {
    const name = (row.icon || '🗂️') + ' ' + row.label;
    let sub;
    if (row.selectable) {
      const line = el('label', 'pick-row');
      const box2 = el('input');
      box2.type = 'checkbox';
      box2.checked = false;            // NOTHING pre-checked (the owner's ask)
      box2.title = row.path + '\nTick to scan this library this click — its plan opens for your review before anything is touched.';
      box2.addEventListener('change', refreshPickScanBtn);
      line.appendChild(box2);
      line.appendChild(el('span', 'pick-name', name));
      card.appendChild(line);
      pickState.boxes.set(row.kind, box2);
      sub = row.status_text + ' · ' + row.path;
    } else {
      card.appendChild(el('div', 'cc_row_name pick-name-readonly', name));
      sub = row.status_text + (row.why ? ' · ' + row.why : '');
    }
    card.appendChild(el('div', 'cc_item pick-sub', sub));
  }
  body.appendChild(card);
  if (model.hint) body.appendChild(el('p', 'help_box pick-hint', model.hint));

  const btnRow = el('div', 'dialog-buttons');
  const cancelBtn = btn('Cancel', 'secondary',
    'Nothing is scanned and nothing is touched — the Scan button stays ready for the next click.',
    () => finishPick(null));
  const scanBtn = btn('🔍 Scan', 'primary',
    'Scan ONLY the ticked vault(s): the hygiene pass still heals both machine vaults quietly, then the ticked library\'s plan opens for your review — nothing is deleted or moved until you answer it.',
    () => {
      const picked = [];
      for (const [kind, b] of pickState.boxes) if (b.checked) picked.push(kind);
      if (!picked.length) return;     // never fires unarmed
      finishPick(picked);
    });
  scanBtn.disabled = true;            // armed only when a connected vault is ticked
  btnRow.append(cancelBtn, scanBtn);
  body.appendChild(btnRow);
  box.appendChild(body);
  pickState.scanBtn = scanBtn;
  escapeHandler = () => finishPick(null);
}

function refreshPickScanBtn() {
  const b = pickState.scanBtn;
  if (!b) return;
  let n = 0;
  for (const [, box2] of pickState.boxes) if (box2.checked) n++;
  b.disabled = n === 0;
  b.textContent = n === 0 ? '🔍 Scan' : (n === 1 ? '🔍 Scan this vault' : '🔍 Scan both vaults');
}

function finishPick(picked) {
  if (!pickState.open) return;
  pickState.open = false;
  closeOverlay();
  // the choice (or the cancel) rides the one flow door — the quiet
  // cancel line, the guards, and the whole job live on the Go side.
  runScanFlow(picked);
}

// ---------------------------------------------------------------------------
// THE SCAN CTA — the button's own machine (ui.py's scan_btn: its own
// state, never the hero's).
// ---------------------------------------------------------------------------
const scanBtnEl = document.getElementById('btn-scan');
let scanFlowBusy = false;

async function runScanFlow(picked) {
  if (scanFlowBusy) return;
  scanFlowBusy = true;
  scanBtnEl.disabled = true;
  scanBtnEl.textContent = 'Scanning…';
  try {
    await StartScan(picked || null);
  } catch (e) {
    // the binding itself never rejects in production; the honest guard
    // is the Go side's refusal lines on the desk.
  }
  // the story ended — the button re-arms (the reference's own law).
  scanBtnEl.disabled = false;
  scanBtnEl.textContent = 'Scan';
  scanFlowBusy = false;
}

if (scanBtnEl) {
  scanBtnEl.disabled = false;         // T-76: the Scan CTA's own machine is live
  scanBtnEl.title = '🔍 Vault scan — first the picker names your detected\n' +
    'vaults (v0.84.0: Websites, GitHub Repos, Personal — with\n' +
    'their note counts); tick the vault(s) this pass should\n' +
    'read, and ONLY those are scanned (the Personal vault is\n' +
    'owner-owned: it is shown read-only — the curator mirrors\n' +
    'Library/ into it and never proposes changes there):\n' +
    '① your delete marks are found — 🗑️ / delete / auto_delete\n' +
    '   tags, the master-table gesture, or a move into the root\n' +
    '   Trash folder (the move IS the verdict);\n' +
    '② the LLM proposes folders and moves for the orphaned /\n' +
    '   uncategorized / too-broadly-filed notes (repo notes ride\n' +
    '   their own evidence — stars, language, topics);\n' +
    '③ EACH scanned vault\'s plan opens in its own review modal\n' +
    '   — 🗂️ Apply plan / ✋ Keep everything (or Telegram, if\n' +
    '   configured) — nothing is deleted or moved until you\n' +
    '   answer; no answer in 300s = nothing done.\n' +
    'Moves never rewrite a note — files only change folders.';
  scanBtnEl.addEventListener('click', openScanPicker);
}

// ---------------------------------------------------------------------------
// THE PLAN MODAL (scan_plan_dialog.py's renderer — the picking door).
// ---------------------------------------------------------------------------
const planState = {
  open: false, model: null, timeoutS: 300, remaining: 0,
  clockTimer: 0, armTimer: 0, guard: null, armed: false,
  sections: [],       // [{key, rows: [{id, checked, low}], master}]
};

EventsOn('scan:ask', (payload) => { openPlanModal(payload); });

function openPlanModal(payload) {
  const model = payload && payload.model;
  if (!model) return;
  planState.open = true;
  planState.model = model;
  planState.timeoutS = (payload.timeout_s && payload.timeout_s > 0) ? payload.timeout_s : 300;
  planState.remaining = Math.max(Math.floor(planState.timeoutS), 1);
  planState.armed = false;
  planState.sections = [];

  const box = openOverlay(640);
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', model.title);
  box.appendChild(el('div', 'card-heading', model.title));
  const body = el('div', 'dialog-body plan-body');

  body.appendChild(el('p', 'dialog-message plan-header', model.header));
  const card = el('div', 'sync_card plan-card');
  card.appendChild(el('div', 'plan_heading plan-heading-neutral', model.library_line));
  if (model.trash_line) card.appendChild(el('div', 'cc_row_name', model.trash_line));
  if (model.locked_line) card.appendChild(el('div', 'cc_row_name', model.locked_line));
  if (model.kept_line) card.appendChild(el('div', 'cc_row_name', model.kept_line));
  if (model.low_confidence_line) card.appendChild(el('div', 'cc_item', model.low_confidence_line));

  for (const sec of model.sections || []) {
    if (sec.total <= 0) continue;
    const card2 = el('div', 'plan_section plan-' + washOf(sec.key));
    const headRow = el('div', 'plan_section_head');
    headRow.appendChild(el('div', 'plan_heading plan-heading-' + sec.tone, sec.heading));
    const section = { key: sec.key, rows: [], master: null, head: null };
    if (sec.rows && sec.rows.length) {
      const master = el('input');
      master.type = 'checkbox';
      const masterLine = el('label', 'plan_master');
      master.title = 'Tick or untick every row in this panel at once';
      master.addEventListener('click', () => toggleSection(section));
      masterLine.append(master, el('span', '', 'All'));
      headRow.appendChild(masterLine);
      section.master = master;
    }
    card2.appendChild(headRow);
    if (sec.rows && sec.rows.length) {
      const list = el('div', 'plan_list');
      for (const row of sec.rows) {
        const line = el('label', 'plan_row');
        const box2 = el('input');
        box2.type = 'checkbox';
        box2.checked = !row.low_confidence;    // the guest's own doubt is a default
        let tip = row.full && row.full !== row.text ? row.full : row.text;
        if (row.low_confidence) {
          tip = 'the LLM itself called this move low-confidence — unticked for your eye\n' + (tip || '');
        }
        if (tip) line.title = tip;
        box2.addEventListener('change', () => { refreshApply(section); });
        line.appendChild(box2);
        line.appendChild(el('span', 'plan_row_text', elideMiddle(row.text, 88)));
        list.appendChild(line);
        section.rows.push({ id: row.id, box: box2 });
      }
      card2.appendChild(list);
    }
    if (sec.more > 0) card2.appendChild(el('div', 'cc_item', '   … +' + sec.more + ' more (counted, not listed)'));
    card.appendChild(card2);
    planState.sections.push(section);
  }
  if (model.summary) card.appendChild(el('p', 'help_box plan-summary', model.summary));
  body.appendChild(card);

  const clock = el('div', 'cc_item plan-clock', clockText());
  planState.clockEl = clock;

  const btnRow = el('div', 'dialog-buttons');
  const keepBtn = btn('✋ Keep everything', 'secondary',
    'The safe default — nothing is deleted, nothing moves, no folder is created; the vault stays exactly as it is (I\'ll propose again on the next scan).',
    () => answerPlan('declined', null));
  const applyBtn = btn('🗂️ Apply plan', 'primary',
    'Apply ONLY the ticked rows: merge the ticked duplicate notes into their site\'s note (links appended, duplicates resting in .trash/consolidated — never blacklisted), create the ticked folders, move the ticked notes (files only — never a byte rewritten), and run the ticked marked deletions through the same machinery a confirmed Telegram ask uses (.trash/banished + never-fetch blacklist). Unticked rows wait — the next scan proposes them again.',
    () => {
      if (!planState.armed) return;   // the spine: no path to CONFIRMED before the door is live
      answerPlan('confirmed', planSelection());
    });
  keepBtn.disabled = true;            // born dead (the fresh-click door)
  applyBtn.disabled = true;
  btnRow.append(keepBtn, applyBtn);

  // THE PINNED VERDICT ROW: the body scrolls; the clock + the two
  // doors stay on screen whatever the pile (the reference's Qt modal
  // kept its buttons visible — its lists scroll, never the dialog).
  box.append(body, clock, btnRow);
  planState.keepBtn = keepBtn;
  planState.applyBtn = applyBtn;

  // ---- THE FRESH-CLICK DOOR: swallow every input for the arm beat ----
  const guard = (e) => { e.stopPropagation(); e.preventDefault(); };
  overlay.addEventListener('click', guard, true);
  overlay.addEventListener('keydown', guard, true);
  overlay.addEventListener('keyup', guard, true);
  planState.guard = guard;
  planState.armTimer = setTimeout(() => armPlanDoors(), ARM_MS);

  planState.clockTimer = setInterval(() => {
    planState.remaining--;
    if (planState.remaining <= 0) {
      answerPlan('timeout', null);    // the safe defer — the same word a silent ask gets
      return;
    }
    if (planState.clockEl) planState.clockEl.textContent = clockText();
  }, 1000);

  escapeHandler = () => answerPlan('declined', null);   // Esc/✕ = declined
  refreshApplyAll();
}

function washOf(key) {
  return { deletions: 'delete_card', consolidations: 'consolidate_card',
    new_folders: 'create_card', moves: 'move_card' }[key] || 'move_card';
}

function clockText() {
  const m = Math.floor(Math.max(planState.remaining, 0) / 60);
  const s = String(Math.max(planState.remaining, 0) % 60).padStart(2, '0');
  return '⏱ Answering window: ' + m + ':' + s +
    ' — at zero nothing is done (the same safe defer as a silent Telegram ask)';
}

function armPlanDoors() {
  planState.armed = true;
  if (planState.guard && overlay) {
    overlay.removeEventListener('click', planState.guard, true);
    overlay.removeEventListener('keydown', planState.guard, true);
    overlay.removeEventListener('keyup', planState.guard, true);
    planState.guard = null;
  }
  refreshApplyAll();
}

// the master sweep (the ONE gesture): any unticked row ticks the whole
// panel; an all-ticked panel unticks entirely — then ONE recount.
function toggleSection(section) {
  if (!section.rows.length) return;
  let all = true;
  for (const r of section.rows) if (!r.box.checked) { all = false; break; }
  for (const r of section.rows) r.box.checked = !all;
  refreshApply(section);
}

function pickedTotals() {
  let total = 0, picked = 0;
  for (const sec of planState.sections) {
    for (const r of sec.rows) { total++; if (r.box.checked) picked++; }
  }
  return { picked, total };
}

function refreshApply(section) {
  // the master mirrors its panel (checked / unchecked / partial).
  if (section && section.master && section.rows.length) {
    let all = true, any = false;
    for (const r of section.rows) { if (r.box.checked) any = true; else all = false; }
    section.master.checked = all;
    section.master.indeterminate = any && !all;
  }
  refreshApplyAll();
}

function refreshApplyAll() {
  const { picked, total } = pickedTotals();
  const applyBtn = planState.applyBtn;
  if (!applyBtn) return;
  applyBtn.textContent = total > 0
    ? '🗂️ Apply plan (' + picked + ' of ' + total + ')' : '🗂️ Apply plan';
  // the enable law: armed AND (something picked, or nothing to pick).
  applyBtn.disabled = !(planState.armed && (picked > 0 || total === 0));
  if (planState.keepBtn) planState.keepBtn.disabled = !planState.armed;
}

// the four-key selection (only a CONFIRMED verdict ever carries it —
// the shape law the bus itself enforces).
function planSelection() {
  const out = { deletions: [], consolidations: [], new_folders: [], moves: [] };
  for (const sec of planState.sections) {
    const ids = [];
    for (const r of sec.rows) if (r.box.checked) ids.push(r.id);
    out[sec.key] = ids;
  }
  return out;
}

function answerPlan(verdict, selection) {
  if (!planState.open) return;
  planState.open = false;
  clearInterval(planState.clockTimer);
  clearTimeout(planState.armTimer);
  closeOverlay();
  ScanAnswer(verdict, selection).catch(() => { /* the bus's own window is the belt */ });
}

// ---------------------------------------------------------------------------
// THE SCORECARD (scan_result_dialog.py's renderer).
// ---------------------------------------------------------------------------
EventsOn('scan:finished', (payload) => {
  if (payload && payload.scorecard) openScorecard(payload.scorecard);
});

function openScorecard(model) {
  if (!model) return;
  const box = openOverlay(480);
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', 'Vault scan — applied');
  box.appendChild(el('div', 'card-heading', 'Vault scan — applied'));
  const body = el('div', 'dialog-body score-body');

  const head = el('div', 'score-head');
  head.appendChild(el('span', 'msg_glyph', '✅'));
  head.appendChild(el('span', 'msg_heading score-heading', ' ' + (model.heading || 'The plan is applied')));
  body.appendChild(head);

  const card = el('div', 'sync_card score-card');
  const sections = (model.vault_sections || []).filter((s) => s.rows && s.rows.length);
  if (model.nothing_landed) {
    card.appendChild(el('p', 'cc_item score-nothing', model.nothing_line));
  } else if (sections.length) {
    for (const sec of sections) {
      const line = el('div', 'score-section-line');
      const label = el('span', 'vault_section_label', sec.label);
      label.title = sec.tip || '';
      const changed = el('span', 'vault_section_count', sec.changed + ' change(s)');
      changed.title = sec.tip || '';
      line.append(label, changed);
      card.appendChild(line);
      for (const row of sec.rows) {
        const rowLine = el('div', 'score-row');
        const n = el('span', 'cc_item', row.name);
        const v = el('span', 'cc_row_name score-value tone-' + (row.tone || 'success'), row.value);
        if (row.tip) v.title = row.tip;   // the why rides the hover — one glance stays one glance
        rowLine.append(n, v);
        card.appendChild(rowLine);
      }
    }
  } else {
    for (const row of model.rows || []) {
      const line = el('div', 'score-row');
      const n = el('span', 'cc_item', row.name);
      const v = el('span', 'cc_row_name score-value tone-' + (row.tone || 'success'), row.value);
      if (row.tip) v.title = row.tip;
      line.append(n, v);
      card.appendChild(line);
    }
  }
  body.appendChild(card);
  body.appendChild(el('p', 'info_note score-footnote', model.footnote));

  // BRIEF (the owner's own word): the self-closing window.
  let remaining = AUTO_CLOSE_S;
  const clock = el('div', 'cc_item score-clock',
    '⏱ This scorecard closes itself in ' + remaining + 's — the full story stays in the log');
  const timer = setInterval(() => {
    remaining--;
    if (remaining <= 0) { clearInterval(timer); closeOverlay(); return; }
    clock.textContent = '⏱ This scorecard closes itself in ' + remaining + 's — the full story stays in the log';
  }, 1000);
  escapeHandler = () => { clearInterval(timer); closeOverlay(); };

  const btnRow = el('div', 'dialog-buttons');
  btnRow.appendChild(btn('OK', 'primary',
    'Dismiss the scorecard — the full story stays in the log (and it closes itself in a moment anyway).',
    () => { clearInterval(timer); closeOverlay(); }));
  // the clock + OK stay pinned under the scrolling body.
  box.append(body, clock, btnRow);
}

// ---------------------------------------------------------------------------
// the evidence probes (the gate leg's visual doors — SB_POC_OPEN_SCAN).
// ---------------------------------------------------------------------------
EventsOn('autorun:open-scan', async (v) => {
  const mode = String(v || '');
  if (mode === 'picker') {
    await openScanPicker();
  } else if (mode === 'plan') {
    await ProbeAskSample();
  } else if (mode === 'scorecard') {
    openScorecard(await ProbeScorecardSample());
  } else if (mode === 'flow') {
    // the REAL flow over the live config (the honest miss lines when
    // no vault is set — the machinery proven end-to-end).
    await StartScan(['repos']);
  }
});
