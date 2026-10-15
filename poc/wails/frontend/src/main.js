// T-70 PoC frontend — wires the Go surface (App bindings + events) to the
// themed desk. Batch DOM appends (one fragment per log:batch event), DOM
// row cap 5000 (trim oldest), so the 10k flood stays interactive.
import { MarkFirstFrame, RunJob, FloodLines, Stats } from '../wailsjs/go/main/App';
import { EventsOn } from '../wailsjs/runtime/runtime';
import './app.css';

const desk = document.getElementById('log-desk');
const countEl = document.getElementById('log-count');
const barFill = document.getElementById('bar-fill');
const barLabel = document.getElementById('bar-label');
const sentinel = document.getElementById('sentinel-startup');
const statsEl = document.getElementById('stats');
const btnJob = document.getElementById('btn-job');
const btnFlood = document.getElementById('btn-flood');
const btnStats = document.getElementById('btn-stats');

const MAX_DOM_ROWS = 5000;
let rendered = 0;

function appendRows(rows) {
  const frag = document.createDocumentFragment();
  for (const r of rows) {
    const row = document.createElement('div');
    row.className = 'log-row ' + (r.level || 'info');
    const ts = document.createElement('span');
    ts.className = 'log-ts';
    ts.textContent = r.ts;
    const msg = document.createElement('span');
    msg.className = 'log-msg';
    msg.textContent = r.msg;
    row.append(ts, msg);
    frag.appendChild(row);
  }
  desk.appendChild(frag);
  rendered += rows.length;
  // trim oldest rows past the cap (desk stays light during floods)
  if (rendered > MAX_DOM_ROWS) {
    const excess = rendered - MAX_DOM_ROWS;
    for (let i = 0; i < excess; i++) desk.removeChild(desk.firstChild);
    rendered = MAX_DOM_ROWS;
  }
  countEl.textContent = rendered + ' rows';
  desk.scrollTop = desk.scrollHeight;
}

EventsOn('log:batch', (rows) => appendRows(rows));

EventsOn('job:progress', (p) => {
  barFill.style.width = ((p.pos / p.total) * 100).toFixed(1) + '%';
  barLabel.textContent = p.pos + ' / ' + p.total;
});

btnJob.addEventListener('click', async () => {
  btnJob.disabled = true;
  barLabel.textContent = 'starting…';
  try {
    const summary = await RunJob(50);
    statsEl.textContent = 'job: ' + summary;
  } catch (e) {
    statsEl.textContent = 'job error: ' + e;
  } finally {
    btnJob.disabled = false;
  }
});

btnFlood.addEventListener('click', async () => {
  try {
    const s = await FloodLines(10000);
    statsEl.textContent = 'flood: ' + s;
  } catch (e) {
    statsEl.textContent = 'flood error: ' + e;
  }
});

btnStats.addEventListener('click', async () => {
  statsEl.textContent = 'stats: ' + (await Stats());
});

// cold-start measurement: DOMContentLoaded → MarkFirstFrame round-trip
window.addEventListener('DOMContentLoaded', async () => {
  try {
    const d = await MarkFirstFrame();
    sentinel.textContent = 'startup→frame: ' + d;
    sentinel.classList.add('pass');
  } catch (e) {
    sentinel.textContent = 'startup measure failed: ' + e;
  }
});
