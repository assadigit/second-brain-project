// settings.js — T-73: the Settings window's frontend half
// (dialogs.py's SettingsDialog + ui.py's six pages A, the HTML twin).
// The Go half (webui/settingsui.go → internal/app/settings.go) owns
// every engine law; this file renders the pages from the snapshot and
// wires THE SAVE-ON-CHANGE LAW: editingFinished/toggled → the page
// savers (one commit = one group = one merge-based SaveSettings call).
//
// The window (the reference's shape): the header (Settings title + the
// section hint + Close), the 190px sidebar nav (one row per page — the
// icon pack's section glyphs, two-mode tint via currentColor), and the
// master-detail stack. Non-modal in spirit: the desk keeps living
// underneath (the overlay covers it, the transport keeps flowing).
import {
  Snapshot, VaultStatus, SaveSettings, RemoveVault,
  TestGitHubToken, GenerateAboutMe, UserConfigStatusLine,
  ShowUserConfigFile, RefreshOllamaModels, StartOllamaServer,
  DetectLlamaCpp, RefreshLlamaCppModels, QuickDetect, ApplyQuickDetect,
  PreviewSound, SelectImportFile, BrowseVault,
} from './wailsjs/go/main/SettingsUI.js';
import {
  SaveBotGroup, CheckBotQueue, ProcessBotQueue, ProcessSourcesUrls, DirectPair,
  RefreshDashboard, ClearQuarantine, SaveQuarantineGroup,
  ReadUndoLastBatch, ApplyUndoLastBatch, FetchFromSources,
  SaveBackupGroup, RefreshBackupLadders, BackupNowGuards, BackupNow,
  SealNow, ResyncMirror, PublishNow, DetectRepos, FastSync,
  OpenDashboard, RestoreBackupsList, RestoreBackupRun, CopyBackupTo,
} from './wailsjs/go/main/PagesBUI.js';
import {
  Refresh as SyncRefresh, StageDifferences, MarkIntentional,
  ReopenItem, DeferredRows, LedgerCount, StageTables,
  CheckBotQueue as SyncCheckQueue, SealNow as SyncSealNow,
  PublishNow as SyncPublishNow, OpenNote as OpenNoteDoor,
} from './wailsjs/go/main/SyncUI.js';
import { EventsOn } from './wailsjs/runtime/runtime.js';
import { svgEl } from './icons.js';
import { iconify } from './emojicons.js';

// The section table (dialogs.py's TAB_INFO — labels, glyphs, hints;
// the six T-73 pages in the reference's own order; the later pages
// join at their owning tasks).
const SECTIONS = [
  { id: 'credentials', label: 'Credentials', icon: 'key-round',
    hint: 'API keys and tokens — Telegram, GitHub, the worker URL' },
  { id: 'proxy', label: 'Proxy', icon: 'globe',
    hint: 'The SOCKS/HTTP proxy for Telegram and web fetches' },
  { id: 'vault', label: 'Vault', icon: 'folder',
    hint: 'Vault folders, backup repos, the law and pipeline switches' },
  { id: 'llm', label: 'LLM', icon: 'brain',
    hint: 'Local engines or any cloud API — provider, limits, detection' },
  { id: 'input', label: 'Input', icon: 'inbox',
    hint: 'Import a .txt or .md file of links, one address per line' },
  { id: 'dashboard', label: 'Dashboard', icon: 'bar-chart-3',
    hint: 'Vault statistics, the 404 quarantine and batch undo' },
  { id: 'sync', label: 'Sync', icon: 'refresh',
    hint: "The four elements' flow + the stage-by-stage gap matrix — what is waiting, and every gap fixed or explained" },
  { id: 'bot', label: 'Bot', icon: 'bot',
    hint: 'The Telegram bot queue — check, process, verify' },
  { id: 'sources', label: 'Sources', icon: 'satellite-dish',
    hint: 'GitHub links from RSS feeds and Reddit, no API key needed' },
  { id: 'backup', label: 'Backup', icon: 'archive',
    hint: 'Folder backups, VaultSeal and Good Repos publishing' },
  { id: 'sound', label: 'Sound', icon: 'volume-2',
    hint: 'The batch-finish chime — on/off, volume, both previews' },
];

const view = document.getElementById('settings-view');
const hintEl = document.getElementById('settings-hint');
const navEl = document.getElementById('settings-nav');
const stackEl = document.getElementById('settings-stack');
let snap = null;          // the Snapshot payload
let quickRunning = false;  // the Detect & Set double-click guard

// ---------------------------------------------------------------------------
// tiny DOM helpers
// ---------------------------------------------------------------------------
function el(tag, cls, text) {
  const n = document.createElement(tag);
  if (cls) n.className = cls;
  if (text !== undefined) n.textContent = text;
  return n;
}
function input(id, value, opts = {}) {
  const i = el('input', 'form-input');
  i.id = id;
  i.value = value == null ? '' : String(value);
  if (opts.placeholder) i.placeholder = opts.placeholder;
  if (opts.title) i.title = opts.title;
  if (opts.password) i.type = 'password';
  return i;
}
function label(text, title) {
  const l = el('label', 'form-label', text);
  if (title) l.title = title;
  return l;
}
function row(labelEl, ...controls) {
  const r = el('div', 'form-row');
  r.append(labelEl, ...controls);
  return r;
}
function btn(text, kind, title, onClick) {
  const b = el('button', 'btn-' + kind, text);
  if (title) b.title = title;
  if (onClick) b.addEventListener('click', onClick);
  iconify(b);                 // a label's emoji rides as a real glyph
  return b;
}
function helpButton(text) {
  // the '?' glyph: the dense page's law — every explanation wall hides
  // behind one glyph (v0.66.0); the tooltip IS the page's teaching.
  const b = el('button', 'help-btn', '?');
  b.type = 'button';
  b.title = text;
  b.setAttribute('aria-label', 'help');
  return b;
}
function muted(text) { return el('p', 'muted-note', text); }

// the live status label (the ladder's rung: semantic class + ● text).
function statusLabel() {
  const s = el('span', 'vault-status muted-strong', '● —');
  return s;
}

// ---------------------------------------------------------------------------
// the save-on-change law's commit savers (one page group per commit)
// ---------------------------------------------------------------------------
function saveCredentials() {
  return SaveSettings({
    credentials: {
      apiId: $('#set-api-id').value,
      apiHash: $('#set-api-hash').value,
      phone: $('#set-phone').value,
      githubToken: $('#set-github-token').value,
    },
  });
}
function saveProxy() {
  return SaveSettings({
    proxy: {
      enabled: $('#set-proxy-enabled').checked,
      type: $('#set-proxy-type').value,
      host: $('#set-proxy-host').value,
      port: $('#set-proxy-port').value,
      useForWeb: $('#set-proxy-web').checked,
    },
  });
}
function vaultHistory() {
  // the combo's law (on_vault_changed): the current path rides at the
  // front; older picks keep their order; empties never enter.
  const cur = ($('#set-vault-combo').value || '').trim();
  const hist = [];
  if (cur) hist.push(cur);
  for (const h of historyCache) {
    if (h && h !== cur && !hist.includes(h)) hist.push(h);
  }
  return hist;
}
function saveVaultPage() {
  return SaveSettings({
    vault: {
      vaultPath: ($('#set-vault-combo').value || '').trim(),
      history: vaultHistory(),
      websiteVaultPath: $('#set-website-vault').value,
      websiteRepoName: $('#set-website-repo').value,
      manualVaultPath: $('#set-manual-vault').value,
      manualRepoName: $('#set-manual-repo').value,
      webBlocked: $('#set-web-blocked').value,
      webSelf: $('#set-web-self').value,
      pipelineGithub: $('#set-pipe-github').checked,
      pipelineWebsites: $('#set-pipe-websites').checked,
    },
  });
}
function saveLLMPage() {
  return SaveSettings({
    llm: {
      provider: currentProvider(),
      numCtx: $('#set-num-ctx').value,
      maxOutputTokens: $('#set-max-output').value,
      ollamaUrl: $('#set-ollama-url').value,
      ollamaModel: $('#set-ollama-model').value,
      llamaCppUrl: $('#set-llamacpp-url').value,
      llamaCppKey: $('#set-llamacpp-key').value,
      llamaCppModel: $('#set-llamacpp-model').value,
      cloudUrl: $('#set-cloud-url').value,
      cloudKey: $('#set-cloud-key').value,
      cloudModel: $('#set-cloud-model').value,
    },
  });
}
function saveSoundPage() {
  return SaveSettings({
    sound: {
      enabled: $('#set-sound-enabled').checked,
      volume: Number($('#set-sound-volume').value),
    },
  });
}

function $(sel) { return document.querySelector(sel); }

// ---------------------------------------------------------------------------
// the pages
// ---------------------------------------------------------------------------
function buildCredentials() {
  const p = el('div', 'page');
  const f = el('div', 'form');
  const apiId = input('set-api-id', snap.telegramApiId);
  const apiHash = input('set-api-hash', snap.telegramApiHash);
  const phone = input('set-phone', snap.telegramPhone);
  const token = input('set-github-token', snap.githubToken, { password: true });
  f.append(
    row(label('API ID:'), apiId),
    row(label('API Hash:'), apiHash),
    row(label('Phone:'), phone),
    row(label('GitHub Token (optional):'), token),
    row(label(''), btn('Test GitHub Token', 'secondary',
      'Validate the token and show the GitHub account it belongs to',
      () => TestGitHubToken($('#set-github-token').value))),
    row(label(''), btn('About Me Wizard', 'secondary',
      'Generate about_me.md to give the LLM context about who you are',
      openAboutMeWizard)),
  );
  p.appendChild(f);

  // v0.69.0 — the credential memory's quiet truth + the show door.
  const status = el('p', 'muted-note wordwrap', snap.userConfigStatus);
  status.id = 'set-user-config-status';
  p.appendChild(status);
  p.appendChild(row(label(''), btn('Show the credentials file', 'secondary',
    'Open the folder that holds the user config file (Documents/.gitcurator) — the credential memory that carries your logins and tokens over to every new app version automatically.',
    () => ShowUserConfigFile().then(() => {
      UserConfigStatusLine().then((line) => {
        status.textContent = line;
      });
    }))));

  for (const i of [apiId, apiHash, phone, token]) {
    i.addEventListener('change', saveCredentials);
  }
  return p;
}

function buildProxy() {
  const p = el('div', 'page');
  const f = el('div', 'form');
  const enabled = el('input', 'form-check');
  enabled.id = 'set-proxy-enabled';
  enabled.type = 'checkbox';
  enabled.checked = !!snap.proxyEnabled;

  const type = el('select', 'form-input');
  type.id = 'set-proxy-type';
  for (const t of ['socks5', 'socks4', 'http']) {
    const o = el('option', '', t);
    o.value = t;
    type.appendChild(o);
  }
  type.value = snap.proxyType;

  const host = input('set-proxy-host', snap.proxyHost);
  const port = input('set-proxy-port', snap.proxyPort);
  const useWeb = el('input', 'form-check');
  useWeb.id = 'set-proxy-web';
  useWeb.type = 'checkbox';
  useWeb.checked = !!snap.proxyUseForWeb;
  useWeb.title = 'When ON, website fetches try the proxy FIRST and the direct line SECOND (v0.43.0 both doors): if the proxy path fails, times out, or is refused with 403/405/429/451 — a wall aimed at that route\'s IP — the same URL is retried once direct, so a datacenter-exit 403 cannot strand a site your own line can reach (and x.com / YouTube still work through the tunnel). When OFF, web fetches never touch the proxy. Loopback (Ollama / llama.cpp) is NEVER proxied. Telegram is unchanged: it always uses the proxy when one is enabled.';

  const useWebLabel = el('label', 'form-check-label',
    'Use this proxy for web fetches too (Websites pipeline — x.com / YouTube need it; the direct line stays the fallback)');
  useWebLabel.prepend(useWeb);

  f.append(
    row(el('span', ''), enabledLabel('Enable Proxy', enabled)),
    row(label('Type:'), type),
    row(label('Host:'), host),
    row(label('Port:'), port),
    useWebLabel,
  );
  p.appendChild(f);
  for (const c of [enabled, type, host, port, useWeb]) {
    c.addEventListener('change', saveProxy);
  }
  return p;
}

function enabledLabel(text, box) {
  const l = el('label', 'form-check-label', text);
  l.prepend(box);
  return l;
}

let historyCache = [];

function buildVault() {
  const p = el('div', 'page');
  historyCache = (snap.vaultChoices || []).slice();

  // --- GitHub vault (the combo + Browse + Remove) ---
  const combo = el('input', 'form-input');
  combo.id = 'set-vault-combo';
  combo.setAttribute('list', 'set-vault-choices');
  combo.value = snap.vaultPath || '';
  combo.title = 'The GitHub projects vault: every repo link becomes a curated note in this vault\'s category folders.\nPick from the list, browse for a folder, or type a path.';
  const list = el('datalist');
  list.id = 'set-vault-choices';
  for (const v of snap.vaultChoices || []) {
    const o = el('option', '');
    o.value = v;
    list.appendChild(o);
  }
  const browse = btn('Browse...', 'secondary', '', async () => {
    const path = await BrowseVault('github');
    if (path) {
      combo.value = path;
      onVaultChanged();
    }
  });
  const remove = btn('Remove', 'danger', '', () => confirmRemoveVault(combo.value));
  const vaultRow = row(label('GitHub vault:'), combo, browse, remove,
    helpButton('GitHub vault — where the repo notes live.\nPick an Obsidian vault (discovered automatically), Browse for its folder, or type the path.\nRemove takes it off the list only — the folder itself is never deleted.'));
  p.append(vaultRow, list);

  // --- Websites vault group ---
  const webGroup = el('div', 'settings-group');
  webGroup.appendChild(el('div', 'group-title', 'Websites vault'));
  const webVault = input('set-website-vault', snap.websiteVaultPath,
    { placeholder: 'path to the Websites vault' });
  webVault.title = 'The Websites vault: non-GitHub links become curated website notes here.\nThe folder does not need to exist yet — the pipeline creates it on the first run.';
  const webStatus = statusLabel();
  webStatus.id = 'set-website-vault-status';
  webGroup.appendChild(row(webVault, btn('Browse...', 'secondary', '', async () => {
    const path = await BrowseVault('websites');
    if (path) { webVault.value = path; saveVaultPage(); }
  }), helpButton('Websites vault — where website notes live.\nThe folder does not need to exist yet; the pipeline creates it on the first run. Empty = the Websites pipeline has nowhere to write (links keep going to the _inbox tables).')));
  webGroup.appendChild(webStatus);
  const webRepo = input('set-website-repo', snap.websiteRepoName,
    { placeholder: 'private GitHub backup repo' });
  webRepo.title = 'The private GitHub repo the Websites vault is mirrored to after every run. Starts EMPTY — no default, nothing is created automatically; type the name once and it is remembered like the credentials.';
  webGroup.appendChild(row(label('Backup repo:'), webRepo, helpButton(
    'Backup repo — the private GitHub repo the Websites vault is mirrored to after every run (full history; restore with a git clone). Empty = NOT configured: nothing is pushed and no repo is ever created automatically (v0.83.0).')));
  const blocked = input('set-web-blocked', snap.webBlocked,
    { placeholder: 'extra domains to ban (the always-banned law list applies)' });
  blocked.title = 'EXTRA domains the Websites pipeline refuses, on top of the always-banned law list (behind the \'?\' here). Links on them are recorded in the _inbox platform tables only — never fetched, never turned into notes, never retried. Subdomains count (www.x.com matches x.com).';
  webGroup.appendChild(row(label('Blocked domains:'), blocked, helpButton(
    '🔒 Always banned (the law — cannot be removed):\n' + (snap.blockedLawDomains || []).join(', ') +
    '\n\nLinks on these domains never enter the Websites vault: never fetched, never noted, never retried. Existing notes for them are swept to the vault\'s .trash on the next run. The record lives on: the _inbox platform tables and the bot\'s ledger keep every link.\n\nThe \'Blocked domains\' field only ADDS more; the law list always applies.')));
  const self = input('set-web-self', snap.webSelf,
    { placeholder: "the app's own hosts — never fetched (empty = none)" });
  self.title = 'Links on these domains belong to this deployment (the bot\'s auth/OAuth handoff links) — never fetched, never turned into notes; the _inbox row keeps the record with secret query values scrubbed. Subdomains count. Empty field = no self domains.';
  webGroup.appendChild(row(label('Self domains:'), self, helpButton(
    'Self domains — hosts that belong to THIS deployment (the Telegram bot\'s own worker).\nIts auth links (…/auth/?token=…) land in the same chat the curator reads; they are never fetched, never noted; the _inbox row (token scrubbed) is the record. Default: the bot\'s workers.dev URL. Empty = none.')));
  p.appendChild(webGroup);

  // --- Manual Notes vault group ---
  const manualGroup = el('div', 'settings-group');
  manualGroup.appendChild(el('div', 'group-title', 'Manual Notes vault'));
  const manualVault = input('set-manual-vault', snap.manualVaultPath,
    { placeholder: 'path to your Manual Notes vault' });
  manualVault.title = 'Your own vault: the app writes only the read-only Library/ mirror there — never your notes.';
  const manualStatus = statusLabel();
  manualStatus.id = 'set-manual-vault-status';
  manualGroup.appendChild(row(manualVault, btn('Browse...', 'secondary', '', async () => {
    const path = await BrowseVault('manual');
    if (path) { manualVault.value = path; saveVaultPage(); }
  }), helpButton('Manual Notes vault — yours; the app writes only the read-only Library/ mirror there (a synced copy of the curated notes), never your own writing.\nLibrary mirror: run tools/mirror_manual.py — dry-run first, then --apply. Only Library/ is ever touched.')));
  manualGroup.appendChild(manualStatus);
  const manualRepo = input('set-manual-repo', snap.manualRepoName,
    { placeholder: 'private GitHub backup repo (auto-detected)' });
  manualRepo.title = 'The private GitHub repo your Personal vault is backed up to. Starts EMPTY — no default; the repo radar detects the repo that already mirrors this vault (its git remote, its name, or a content fingerprint) and fills this once. Type a name yourself and it is remembered — your word always wins over detection.';
  manualGroup.appendChild(row(label('Backup repo:'), manualRepo, helpButton(
    'Backup repo — the private GitHub repo the Personal vault is mirrored to (the vault\'s whole content, full history; restore with a git clone).\n\nEmpty = not configured: the radar (v0.84.0) looks for the repo that already backs up this vault — its own git remote first, then the vault folder\'s name, then a content fingerprint — and fills this field automatically. Nothing is ever created (v0.83.0 law); an ambiguous case stays empty for you to type once.')));
  p.appendChild(manualGroup);

  // --- Pipelines group ---
  const pipesGroup = el('div', 'settings-group');
  pipesGroup.appendChild(el('div', 'group-title', 'Pipelines'));
  const pipeGithub = el('input', 'form-check');
  pipeGithub.id = 'set-pipe-github';
  pipeGithub.type = 'checkbox';
  pipeGithub.checked = !!snap.pipelineGithub;
  pipeGithub.title = 'Repo links become curated notes in the GitHub vault.';
  const pipeGithubLabel = enabledLabel('GitHub projects', pipeGithub);
  const pipeWeb = el('input', 'form-check');
  pipeWeb.id = 'set-pipe-websites';
  pipeWeb.type = 'checkbox';
  pipeWeb.checked = !!snap.pipelineWebsites;
  pipeWeb.title = 'Non-GitHub links are fetched and become classified notes in the Websites vault. OFF = they keep going to the _inbox tables.';
  const pipeWebLabel = enabledLabel('Websites (non-GitHub links become notes)', pipeWeb);
  pipesGroup.append(pipeGithubLabel, pipeWebLabel, helpButton(
    '💡 Pipelines — what a SYNC processes.\nGitHub (ON by default): repo links become curated notes in the GitHub vault — today\'s behavior exactly.\nWebsites: non-GitHub links are fetched and become classified notes in the Websites vault (OFF = they keep going to the _inbox tables, like before).\nBoth OFF = a SYNC processes nothing.'));
  p.appendChild(pipesGroup);

  // --- the save-on-change wiring (ui.py:576-586) ---
  const refreshStatuses = () => {
    VaultStatus(webVault.value, 'websites').then((st) => paintStatus(webStatus, st));
    VaultStatus(manualVault.value, 'manual').then((st) => paintStatus(manualStatus, st));
  };
  const paintStatus = (node, st) => {
    node.textContent = st.text;
    node.className = 'vault-status status-' + st.state;
  };
  for (const i of [webVault, manualVault]) i.addEventListener('input', refreshStatuses);
  for (const i of [webRepo, blocked, self, manualRepo]) i.addEventListener('change', saveVaultPage);
  for (const i of [webVault, manualVault]) i.addEventListener('change', saveVaultPage);
  for (const c of [pipeGithub, pipeWeb]) c.addEventListener('change', saveVaultPage);
  combo.addEventListener('change', onVaultChanged);
  refreshStatuses();
  return p;

  function onVaultChanged() {
    saveVaultPage().then(() => {
      const cur = (combo.value || '').trim();
      if (cur && !historyCache.includes(cur)) historyCache.unshift(cur);
    });
  }
}

function confirmRemoveVault(path) {
  path = (path || '').trim();
  if (!path) return;
  miniDialog('Remove Vault',
    'Remove \'' + path + '\' from the list? (This will not delete the folder.).',
    ['Remove', 'Cancel'], (answer) => {
      if (answer === 'Remove') {
        RemoveVault(path).then(() => {
          historyCache = historyCache.filter((h) => h !== path);
          const combo = $('#set-vault-combo');
          if ((combo.value || '').trim() === path) combo.value = '';
        });
      }
    });
}

function currentProvider() {
  const host = document.querySelector('input[name=llm-host]:checked');
  const engine = document.querySelector('input[name=llm-engine]:checked');
  if (host && host.value === 'cloud') return 'cloud';
  if (engine && engine.value === 'llamacpp') return 'llamacpp';
  return 'ollama';
}

function buildLLM() {
  const p = el('div', 'page');

  // === Card 1 — provider (the host radios) ===
  const providerCard = el('div', 'settings-card');
  const providerRow = el('div', 'card-row');
  providerRow.appendChild(el('span', 'card-heading', 'LLM provider'));
  const hostLocal = el('input');
  hostLocal.type = 'radio'; hostLocal.name = 'llm-host'; hostLocal.value = 'local';
  hostLocal.id = 'set-host-local';
  hostLocal.title = 'Run the model on YOUR machine — no data leaves it.\nTwo engines: Ollama (http://localhost:11434) or a llama.cpp llama-server (http://127.0.0.1:8080). Both are DETECTED — the Detect & Set buttons find the running server, list its models and configure everything in one click.';
  const hostCloud = el('input');
  hostCloud.type = 'radio'; hostCloud.name = 'llm-host'; hostCloud.value = 'cloud';
  hostCloud.id = 'set-host-cloud';
  hostCloud.title = 'Any OpenAI-compatible endpoint: OpenAI, OpenRouter, Together, vLLM, LM Studio, Cloudflare Workers AI…\nAND Anthropic Claude — api.anthropic.com URLs automatically use the Claude Messages API (x-api-key + anthropic-version).\nThe URL decides the wire format; the same Model field takes \'gpt-4o-mini\' or \'claude-sonnet-4-5\' alike.';
  if (snap.llmProvider === 'cloud') hostCloud.checked = true; else hostLocal.checked = true;
  providerRow.append(hostLocal, checkLabel('Locally hosted LLM model', hostLocal),
    hostCloud, checkLabel('Cloud API model', hostCloud));
  providerCard.append(providerRow, muted(
    'Where the model runs. Local engines keep every byte on your machine; the cloud API reaches any OpenAI-compatible endpoint or Anthropic Claude — the URL decides the wire format.'));

  // === Card 2 — limits ===
  const limitsCard = el('div', 'settings-card');
  limitsCard.appendChild(el('span', 'card-heading', 'Limits'));
  const numCtx = input('set-num-ctx', snap.llmNumCtx);
  numCtx.style.width = '160px';
  limitsCard.append(
    row(label('Model max context window (tokens):'), numCtx),
    muted('What the model can READ — the total context window. Ollama: sent as num_ctx with every call (its defaults are small and truncate silently). OpenAI-compatible servers: the window is fixed at server launch (llama.cpp -c / vLLM --max-model-len) — this value only powers the over-budget warning; Claude: the same warning. 0 = not sent — the server\'s own window is used.'));
  const maxOut = input('set-max-output', snap.llmMaxOutputTokens);
  maxOut.style.width = '160px';
  limitsCard.append(
    row(label('Output max tokens:'), maxOut),
    muted('What the model can WRITE — the cap on the model\'s answer. Ollama: sent as options.num_predict. OpenAI-compatible: sent as max_tokens. Claude: max_tokens is REQUIRED — the configured value is sent, with a 4096-token fallback when unset. 0 = not sent — the server\'s default cap is used.'));

  // === Card 3 — local engine ===
  const localCard = el('div', 'settings-card');
  localCard.id = 'set-local-card';
  const engineRow = el('div', 'card-row');
  engineRow.appendChild(el('span', 'card-heading', 'Local engine'));
  const engineOllama = el('input');
  engineOllama.type = 'radio'; engineOllama.name = 'llm-engine'; engineOllama.value = 'ollama';
  engineOllama.id = 'set-engine-ollama';
  engineOllama.title = 'Use a local Ollama server (http://localhost:11434 by default).\nNo API key required — runs entirely on your machine.';
  const engineLlama = el('input');
  engineLlama.type = 'radio'; engineLlama.name = 'llm-engine'; engineLlama.value = 'llamacpp';
  engineLlama.id = 'set-engine-llamacpp';
  engineLlama.title = 'A local llama.cpp server (llama-server, http://127.0.0.1:8080 by default). Detected like Ollama: \'Detect\' finds the server and its loaded model automatically (llama.cpp /props + /v1/models).\nNo API key unless the server was started with --api-key.';
  if (snap.llmProvider === 'llamacpp') engineLlama.checked = true; else engineOllama.checked = true;
  engineRow.append(engineOllama, checkLabel('Ollama', engineOllama),
    engineLlama, checkLabel('llama.cpp', engineLlama));
  localCard.append(engineRow, muted(
    'Quick switch — one click per engine: probe the running server, pick the model when several are installed, set the provider + model + URL and save.'));
  const quickRow = el('div', 'card-row');
  quickRow.append(
    btn('Detect & Set Ollama', 'secondary',
      'Probe the Ollama server, pick the model when several are installed, set the provider + model + URL and save',
      () => runQuickDetect('ollama')),
    btn('Detect & Set llama.cpp', 'secondary',
      'Find the running llama-server (process ports + common ports), pick the model when several are advertised, set the provider + model + URL and save',
      () => runQuickDetect('llamacpp')),
  );
  localCard.appendChild(quickRow);

  // --- Ollama section ---
  const ollamaGroup = el('div', 'settings-group');
  ollamaGroup.id = 'set-ollama-group';
  ollamaGroup.appendChild(el('div', 'group-title', 'Ollama'));
  const ollamaURL = input('set-ollama-url', snap.ollamaUrl);
  const ollamaModel = input('set-ollama-model', snap.ollamaModel);
  const modelRow = el('div', 'inline-row');
  modelRow.append(ollamaModel,
    btn('Refresh', 'secondary', 'Reload the model list from the Ollama server', () => {
      RefreshOllamaModels($('#set-ollama-url').value, $('#set-ollama-model').value)
        .then((r) => { if (r && r.ok) ollamaModel.value = r.model; });
    }),
    btn('Start Server', 'secondary', 'Start the local Ollama server (ollama serve)', () => {
      StartOllamaServer().then((r) => {
        if (r && r.ok) setTimeout(() => {
          RefreshOllamaModels($('#set-ollama-url').value, $('#set-ollama-model').value)
            .then((r2) => { if (r2 && r2.ok) ollamaModel.value = r2.model; });
        }, 3000);  // QTimer.singleShot(3000, refresh)'s twin
      });
    }));
  ollamaGroup.append(row(label('Ollama URL:'), ollamaURL), row(label('Model:'), modelRow));

  // --- llama.cpp section ---
  const llamaGroup = el('div', 'settings-group');
  llamaGroup.id = 'set-llamacpp-group';
  llamaGroup.appendChild(el('div', 'group-title', 'llama.cpp'));
  const llamaURL = input('set-llamacpp-url', snap.llamaCppUrl);
  const llamaKey = input('set-llamacpp-key', snap.llamaCppKey,
    { password: true, placeholder: '(empty — only needed when llama-server was started with --api-key)' });
  const llamaModel = input('set-llamacpp-model', snap.llamaCppModel,
    { placeholder: '(auto-detected from the server)' });
  const llamaModelRow = el('div', 'inline-row');
  llamaModelRow.append(llamaModel,
    btn('Detect', 'primary',
      'Find the running llama-server automatically: its PROCESS\'s listening ports first (any --port), then the common ports (8080 first). Positively identifies llama.cpp, then fills this URL and the model automatically.',
      () => DetectLlamaCpp($('#set-llamacpp-key').value).then((r) => {
        if (r && r.found) {
          llamaURL.value = r.url;
          llamaModel.value = r.model;
        }
      })),
    btn('Refresh', 'secondary', 'Reload the model list from the llama.cpp server at the URL above',
      () => RefreshLlamaCppModels($('#set-llamacpp-url').value, $('#set-llamacpp-key').value)
        .then((r) => { if (r && r.found) llamaModel.value = r.model; })));
  llamaGroup.append(
    row(label('Server URL:'), llamaURL),
    row(label('API key:'), llamaKey),
    row(label('Model:'), llamaModelRow));
  localCard.append(ollamaGroup, llamaGroup);

  // === The Cloud API card ===
  const cloudCard = el('div', 'settings-card');
  cloudCard.id = 'set-cloud-card';
  cloudCard.appendChild(el('span', 'card-heading', 'Cloud API — OpenAI-compatible endpoint'));
  const cloudURL = input('set-cloud-url', snap.cloudUrl,
    { placeholder: 'https://api.openai.com/v1 — or https://api.anthropic.com/v1 for Claude' });
  cloudURL.title = 'OpenAI-compatible: https://api.openai.com/v1, OpenRouter, Together, vLLM, LM Studio, a local server…\nClaude: https://api.anthropic.com/v1 — the URL decides the wire format automatically (Messages API, x-api-key header).';
  const cloudKey = input('set-cloud-key', snap.cloudKey,
    { password: true, placeholder: 'sk-… / sk-ant-… (kept locally in config.json)' });
  const cloudModel = input('set-cloud-model', snap.cloudModel,
    { placeholder: 'gpt-4o-mini / claude-sonnet-4-5 / …' });
  cloudCard.append(
    row(label('API URL:'), cloudURL),
    row(label('API Key:'), cloudKey),
    row(label('Model:'), cloudModel),
    muted('Any OpenAI-compatible endpoint (OpenAI, OpenRouter, Together, vLLM, LM Studio, Cloudflare Workers AI…) or Anthropic Claude — the URL decides the wire format.'));

  p.append(providerCard, limitsCard, localCard, cloudCard);

  // --- the radio visibility law (_toggle_llm_provider): radios flip
  // VISIBILITY at once, and a change (the user's own click) commits the
  // LLM page through the same save-on-change law.
  const paintVisibility = () => {
    const isLocal = hostLocal.checked;
    const isOllama = engineOllama.checked;
    localCard.hidden = !isLocal;
    cloudCard.hidden = hostCloud.checked;
    ollamaGroup.hidden = !isOllama;
    llamaGroup.hidden = isOllama;
  };
  const onRadioChange = () => { paintVisibility(); saveLLMPage(); };
  for (const r of [hostLocal, hostCloud, engineOllama, engineLlama]) {
    r.addEventListener('change', onRadioChange);
  }
  paintVisibility();  // the initial paint never saves
  // the fields' commit savers
  for (const i of [numCtx, maxOut, ollamaURL, ollamaModel, llamaURL, llamaKey,
    llamaModel, cloudURL, cloudKey, cloudModel]) {
    i.addEventListener('change', saveLLMPage);
  }
  return p;
}

function checkLabel(text, box) {
  const l = el('label', 'form-check-label', text);
  l.prepend(box);
  return l;
}

function runQuickDetect(provider) {
  if (quickRunning) return;
  quickRunning = true;
  const liveURL = provider === 'ollama'
    ? $('#set-ollama-url').value : $('#set-llamacpp-url').value;
  const liveKey = $('#set-llamacpp-key').value;
  QuickDetect(provider, liveURL, liveKey).then((r) => {
    quickRunning = false;
    if (!r || !r.ok || !r.found) return;
    if (r.applied) {
      paintQuickApplied(r);
      return;
    }
    // several models → the menu ("a menu like the current one")
    const current = provider === 'ollama'
      ? $('#set-ollama-model').value : $('#set-llamacpp-model').value;
    const items = (r.models || []).slice();
    miniSelect('Select the ' + (provider === 'ollama' ? '🧠 Ollama' : '🦙 llama.cpp') + ' model',
      items.length + ' models detected at ' + r.baseUrl + ' — pick the one to use:',
      items, current, (choice) => {
        if (!choice) return;  // the honest cancel line already spoke
        ApplyQuickDetect(provider, r.baseUrl, items, choice, liveLLMValues()).then(() => {
          paintQuickApplied({ provider, baseUrl: r.baseUrl, choice });
        });
      });
  }).catch(() => { quickRunning = false; });
}

function paintQuickApplied(r) {
  if (r.provider === 'ollama') {
    $('#set-host-local').checked = true;
    $('#set-engine-ollama').checked = true;
    $('#set-ollama-url').value = r.baseUrl;
    $('#set-ollama-model').value = r.choice;
  } else {
    $('#set-host-local').checked = true;
    $('#set-engine-llamacpp').checked = true;
    $('#set-llamacpp-url').value = r.baseUrl.replace(/\/$/, '') + '/v1';
    $('#set-llamacpp-model').value = r.choice;
  }
  toggleAfterRadio();
}

function liveLLMValues() {
  return {
    provider: currentProvider(),
    numCtx: $('#set-num-ctx').value,
    maxOutputTokens: $('#set-max-output').value,
    ollamaUrl: $('#set-ollama-url').value,
    ollamaModel: $('#set-ollama-model').value,
    llamaCppUrl: $('#set-llamacpp-url').value,
    llamaCppKey: $('#set-llamacpp-key').value,
    llamaCppModel: $('#set-llamacpp-model').value,
    cloudUrl: $('#set-cloud-url').value,
    cloudKey: $('#set-cloud-key').value,
    cloudModel: $('#set-cloud-model').value,
  };
}

function toggleAfterRadio() {
  // re-run the visibility law without re-saving (the paint is a refresh).
  const hostLocal = $('#set-host-local');
  const engineOllama = $('#set-engine-ollama');
  $('#set-local-card').hidden = !hostLocal.checked;
  $('#set-cloud-card').hidden = hostLocal.checked;
  $('#set-ollama-group').hidden = !engineOllama.checked;
  $('#set-llamacpp-group').hidden = engineOllama.checked;
}

function buildInput() {
  const p = el('div', 'page');
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'Import txt file'));
  const file = input('set-import-file', '',
    { placeholder: 'Path to a .txt or .md file — one URL per line' });
  file.title = 'A plain-text or Markdown file with one address per line.\nMarkdown links, bullets and trailing notes are fine — the address is read out of them. Lines starting with # are comments; blank lines are skipped; duplicates are ignored (reported); lines with no address are listed in the log.\nGitHub repos go to the GitHub pipeline, every other website to the Websites pipeline — exactly like a fetched batch.';
  const pick = btn('Select…', 'secondary', 'Pick the .txt / .md file to import', async () => {
    const path = await SelectImportFile();
    if (path) file.value = path;
  });
  group.append(row(file, pick), el('p', 'info-note',
    'PROCESS (main view) imports the file: GitHub repos are noted into the GitHub vault, every other website into the Websites vault. Use SYNC instead to fetch the Telegram bot queue.'));
  p.appendChild(group);
  return p;
}

function buildSound() {
  const p = el('div', 'page');
  p.appendChild(el('p', 'info-header',
    'Sound\nWhat the app plays when a batch runs to its natural end: a rising arpeggio for a clean finish, a softer descending tone when links failed and need retry.'));
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'Batch Completion'));
  const enabled = el('input', 'form-check');
  enabled.id = 'set-sound-enabled';
  enabled.type = 'checkbox';
  enabled.checked = !!snap.soundEnabled;
  enabled.title = 'sound_enabled — when OFF, every batch finishes in silence (the scorecard modal still opens).';
  group.appendChild(enabledLabel('Play the chime when a batch finishes', enabled));

  const volRow = el('div', 'form-row');
  volRow.appendChild(label('Volume:'));
  const slider = el('input', 'form-slider');
  slider.id = 'set-sound-volume';
  slider.type = 'range';
  slider.min = '0'; slider.max = '100';
  slider.step = '1';
  slider.value = String(snap.soundVolume);
  slider.title = 'sound_volume — 0 to 100 here, saved as 0.0–1.0 and clamped into Qt\'s [0, 1] range.';
  const volLabel = el('span', 'cc-item', snap.soundVolume + '%');
  volRow.append(slider, volLabel);
  group.appendChild(volRow);

  const previewRow = el('div', 'card-row');
  previewRow.append(
    btn('Preview — Clean Finish', 'secondary',
      'Play the success chime (success.wav) through the exact batch path: the switch and the volume above apply.',
      () => playPreview(false)),
    btn('Preview — Needs Retry', 'secondary',
      'Play the softer needs-retry tone (retry.wav) through the exact batch path — what a batch with failed links sounds like.',
      () => playPreview(true)));
  group.appendChild(previewRow);
  group.appendChild(el('p', 'info-note',
    '💡 Both previews use the exact batch path (same switch, same volume). The chime is best-effort by design: a machine without an audio backend stays silent — one notice in the log, never an error.'));
  p.appendChild(group);

  // save-on-change: the toggle at once; the slider on RELEASE (the
  // valueChanged-label-only law); the % label follows every tick.
  enabled.addEventListener('change', saveSoundPage);
  slider.addEventListener('input', () => { volLabel.textContent = slider.value + '%'; });
  slider.addEventListener('change', saveSoundPage);
  return p;
}

function playPreview(needsRetry) {
  PreviewSound(needsRetry).then((r) => {
    if (!r || !r.play) return;
    // the QSoundEffect twin: the audio element never raises; a machine
    // without an audio backend accepts the play and stays silent.
    const audio = new Audio(r.dataUrl);
    audio.volume = Math.max(0, Math.min(1, r.volume));
    audio.play().catch(() => { /* silent accept — rule 4 */ });
  });
}

// ---------------------------------------------------------------------------
// the About Me Wizard (lifecycle.py's dialog, the HTML twin).
// ---------------------------------------------------------------------------
function openAboutMeWizard() {
  const fields = [
    ['Who are you:', 'set-am-name', 'e.g. Software Engineer, Researcher, Student'],
    ['Your role:', 'set-am-role', 'e.g. AI/ML Engineer, Full-stack Developer'],
    ['Your interests:', 'set-am-interests', 'e.g. AI agents, developer tools, self-hosted software, automation...'],
    ['Your objectives:', 'set-am-objectives', 'e.g. Find tools for my workflow, learn new frameworks, build a knowledge base...'],
    ['Technologies you use:', 'set-am-tech', 'e.g. Python, Rust, TypeScript, Docker, Kubernetes'],
    ['How you use this vault:', 'set-am-vault', 'e.g. I organize projects by domain (AI, Tools, Infrastructure) and use notes for quick recall.'],
  ];
  const body = el('div');
  body.appendChild(el('p', 'info-header',
    'About Me Wizard\n\nThis generates an \'about_me.md\' file that gives the LLM context about you.\nThe LLM uses this to personalize its analysis and explain how each project\nmight specifically help YOU.\n\nFill in the fields below (leave blank if you prefer not to answer):'));
  for (const [lab, id, ph] of fields) {
    let i = input(id, '', { placeholder: ph });
    if (lab.startsWith('Your interests') || lab.startsWith('Your objectives') || lab.startsWith('How you')) {
      i = el('textarea', 'form-input form-area');
      i.id = id;
      i.placeholder = ph;
    }
    body.appendChild(row(label(lab), i));
  }
  miniCustom('About Me Wizard', body, [['Generate about_me.md', 'primary'], ['Cancel', 'secondary']],
    (answer) => {
      if (answer !== 'Generate about_me.md') return;
      GenerateAboutMe({
        name: val('#set-am-name'),
        role: val('#set-am-role'),
        interests: val('#set-am-interests'),
        objectives: val('#set-am-objectives'),
        tech: val('#set-am-tech'),
        vaultUse: val('#set-am-vault'),
      });
    });
  function val(sel) { const n = document.querySelector(sel); return n ? n.value : ''; }
}

// ---------------------------------------------------------------------------
// the mini-dialog kit (the themed confirm/select/custom — the
// _show_custom_question family's HTML twin; the full dialogs polish is
// T-76/T-77's, these are the pages' own small modals).
// ---------------------------------------------------------------------------
let dialogOverlay = null;

function miniDialog(title, message, buttons, onAnswer) {
  const body = el('div');
  body.appendChild(el('p', 'dialog-message', message));
  miniCustom(title, body, buttons.map((b, i) => [b, i === 0 ? 'primary' : 'secondary']), onAnswer);
}

function miniSelect(title, message, items, current, onChoice) {
  const body = el('div');
  body.appendChild(el('p', 'dialog-message', message));
  const select = el('select', 'form-input');
  select.style.minWidth = '280px';
  for (const it of items) {
    const o = el('option', '', it);
    o.value = it;
    select.appendChild(o);
  }
  if (current && items.includes(current)) select.value = current;
  body.appendChild(row(label(''), select));
  miniCustom(title, body, [['Use this model', 'primary'], ['Cancel', 'secondary']], (answer) => {
    onChoice(answer === 'Use this model' ? select.value : '');
  });
}

function miniCustom(title, bodyEl, buttons, onAnswer) {
  closeMiniDialog();
  dialogOverlay = el('div', 'dialog-overlay');
  const box = el('div', 'dialog-box');
  box.setAttribute('role', 'dialog');
  box.setAttribute('aria-label', title);
  box.appendChild(el('div', 'card-heading', title));
  const body = el('div', 'dialog-body');
  body.appendChild(bodyEl);
  box.appendChild(body);
  const btnRow = el('div', 'dialog-buttons');
  for (const [text, kind] of buttons) {
    btnRow.appendChild(btn(text, kind, '', () => {
      closeMiniDialog();
      onAnswer(text);
    }));
  }
  box.appendChild(btnRow);
  iconify(box);                // the dialog's emojis ride as real glyphs
  dialogOverlay.appendChild(box);
  view.appendChild(dialogOverlay);
}

function closeMiniDialog() {
  if (dialogOverlay && dialogOverlay.parentNode) dialogOverlay.remove();
  dialogOverlay = null;
}

// ---------------------------------------------------------------------------
// the window assembly (SettingsDialog's nav + stack + hint).
// ---------------------------------------------------------------------------
const builders = {
  credentials: buildCredentials,
  proxy: buildProxy,
  vault: buildVault,
  llm: buildLLM,
  input: buildInput,
  dashboard: buildDashboard,
  sync: buildSync,
  bot: buildBot,
  sources: buildSources,
  backup: buildBackup,
  sound: buildSound,
};

// the queue's last check (the Process All door's payload)
let queuePayload = null;

function buildWindow() {
  navEl.textContent = '';
  stackEl.textContent = '';
  for (const sec of SECTIONS) {
    const item = el('button', 'settings-nav-item');
    item.type = 'button';
    item.appendChild(svgEl(sec.icon, 'settings-nav-icon'));
    item.appendChild(el('span', '', sec.label));
    item.title = sec.hint;
    item.addEventListener('click', () => selectSection(sec.id));
    item.dataset.section = sec.id;
    navEl.appendChild(item);

    const page = builders[sec.id]();
    page.dataset.page = sec.id;
    page.hidden = true;
    iconify(page);              // the page's emojis ride as real glyphs
    stackEl.appendChild(page);
  }
  selectSection('credentials');
}

let syncShownOnce = false;  // the first-show door (ONE automatic refresh)

function selectSection(id) {
  for (const sec of SECTIONS) {
    if (sec.id === id) hintEl.textContent = sec.hint;
  }
  if (id === 'sync' && !syncShownOnce) {
    syncShownOnce = true;  // exactly once per session (the owner's Q2 answer)
    refreshSyncView();
  }
  for (const item of navEl.querySelectorAll('.settings-nav-item')) {
    item.classList.toggle('settings-nav-active', item.dataset.section === id);
  }
  for (const page of stackEl.children) {
    page.hidden = page.dataset.page !== id;
  }
}

async function openSettings() {
  snap = await Snapshot();
  buildWindow();
  view.hidden = false;
}

function closeSettings() {
  view.hidden = true;
  closeMiniDialog();
}

// the doors
const gear = document.getElementById('btn-settings');
gear.appendChild(svgEl('settings', 'chrome-icon'));
gear.addEventListener('click', openSettings);
document.getElementById('btn-settings-close').addEventListener('click', closeSettings);
EventsOn('autorun:open-settings', async (section) => {
  await openSettings();
  if (section) selectSection(String(section));
});

// ---------------------------------------------------------------------------
// T-74 — the pages B (dashboard.py + bot_queue.py + backup_seal.py +
// the Sources page, the HTML twins over internal/app's engines).
// ---------------------------------------------------------------------------

function buildDashboard() {
  const p = el('div', 'page');

  // The button row: the ONE filled primary + the filled danger.
  const btnRow = el('div', 'card-row');
  btnRow.append(
    btn('Refresh Dashboard', 'primary', '', () => refreshDashboard()),
    btn('Undo Last Batch', 'danger',
      'v22 Feature 6: deletes the .md files written by the most recent batch (listed in <vault>/_undo_last_batch.txt).',
      () => undoLastBatchFlow()));
  p.appendChild(btnRow);

  // The results panel — the tab's ONE growable region.
  const dashText = el('textarea', 'form-area');
  dashText.id = 'set-dash-text';
  dashText.readOnly = true;
  dashText.style.minHeight = '120px';
  dashText.style.flex = '1';
  dashText.placeholder = "Click 'Refresh Dashboard' to scan the vault and view statistics.";
  p.appendChild(dashText);

  // The 404 quarantine group (the merged lineage design).
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'Deleted Repos — 404 Quarantine'));
  const ctrlRow = el('div', 'form-row');
  ctrlRow.appendChild(label('Confirm dead after'));
  const spin = el('input', 'form-input');
  spin.id = 'set-strike-threshold';
  spin.type = 'number';
  spin.min = '2'; spin.max = '10'; spin.step = '1';
  spin.style.width = '72px';
  spin.value = String(snap.strikeThreshold || 3);
  spin.title = 'How many consecutive 404s (counted across sessions) before a repo is quarantined (auto-ignored). A successful fetch resets its counter. Minimum 2, default 3.';
  ctrlRow.append(spin, label('consecutive 404s'));
  ctrlRow.append(btn('Refresh', 'secondary',
    'Reload the 404 quarantine table from cache.db', () => refreshDashboard()));
  ctrlRow.appendChild(btn('Reset Quarantine', 'secondary',
    'Reset ALL 404 attempt counters — quarantined repos are re-checked on the next run instead of being auto-ignored.',
    () => {
      miniDialog('♻️ Reset 404 Quarantine',
        'Reset the attempt counter for EVERY recorded URL?\nQuarantined repos will be re-checked on the next run instead of being auto-ignored.',
        ['Reset', 'Cancel'], (answer) => {
          if (answer !== 'Reset') return;
          ClearQuarantine().then((r) => {
            if (r && r.quarantine) $('#set-quarantine-text').value = r.quarantine;
          });
        });
    }));
  group.appendChild(ctrlRow);
  const quarText = el('textarea', 'form-area');
  quarText.id = 'set-quarantine-text';
  quarText.readOnly = true;
  quarText.style.minHeight = '110px';
  quarText.placeholder = 'No 404 attempts recorded yet — deleted repos will appear here with their attempt counts after a run.';
  group.appendChild(quarText);
  // The valueChanged save (persist immediately — the merge law).
  spin.addEventListener('change', () => {
    SaveQuarantineGroup({ strikeThreshold: parseInt(spin.value, 10) || 3 });
  });
  p.appendChild(group);
  return p;
}

function refreshDashboard() {
  RefreshDashboard().then((r) => {
    if (!r) return;
    $('#set-dash-text').value = r.text || '';
    $('#set-quarantine-text').value = r.quarantine || '';
  });
}

function undoLastBatchFlow() {
  ReadUndoLastBatch().then((res) => {
    if (!res) return;
    if (res.stage === 'ok') {
      miniDialog('↩️ Undo Last Batch', res.confirmMessage, ['Delete the files', 'Cancel'],
        (answer) => {
          if (answer !== 'Delete the files') return;
          ApplyUndoLastBatch().then(() => refreshDashboard());
        });
      return;
    }
    // The early doors' own dialogs (the reference's information boxes).
    miniDialog(res.stage === 'gone' ? 'Nothing to Undo' : 'No Batch to Undo',
      res.message || 'Nothing to undo.', ['OK'], () => {});
  });
}

function buildBot() {
  const p = el('div', 'page');
  p.appendChild(el('p', 'info-header',
    '🤖 Bot Queue\nForward GitHub repo messages to your dedicated bot (@githubfetcherbot).\nClick \'Check Queue\' to fetch pending repos, then \'Process All\'.'));

  const group = el('div', 'settings-group');
  const userRow = el('div', 'form-row');
  userRow.appendChild(label('Bot Username:'));
  const user = input('set-bot-username', snap.botUsername || 'githubfetcherbot',
    { placeholder: 'e.g. githubfetcherbot' });
  userRow.appendChild(user);
  group.appendChild(userRow);

  const tokenRow = el('div', 'form-row');
  tokenRow.appendChild(label('Bot Token:'));
  const token = input('set-bot-token', snap.botToken || '',
    { placeholder: 'e.g. 123456789:AAF... (optional)', password: true });
  const saveBtn = btn('Save', 'secondary', '', () => saveBotFields());
  tokenRow.append(token, saveBtn);
  group.appendChild(tokenRow);
  // save-on-change (editingFinished → the page saver).
  user.addEventListener('change', saveBotFields);
  token.addEventListener('change', saveBotFields);
  p.appendChild(group);

  // Queue controls + the pending badge (zinc ⏳ routine / mint ✅ zero).
  const queueRow = el('div', 'card-row');
  queueRow.appendChild(btn('Check Queue', 'secondary', '', () => checkQueueFlow()));
  const badge = el('span', 'pending-badge', '');
  badge.id = 'set-queue-badge';
  badge.hidden = true;
  queueRow.appendChild(badge);
  queueRow.appendChild(btn('Process All', 'primary', '', () => processQueueFlow()));
  p.appendChild(queueRow);

  p.appendChild(el('label', 'form-label', 'Pending Repos:'));
  const display = el('textarea', 'form-area');
  display.id = 'set-queue-display';
  display.readOnly = true;
  display.style.minHeight = '120px';
  display.style.flex = '1';
  display.placeholder = "Click 'Check Queue' to fetch pending repos from your bot...";
  p.appendChild(display);

  // THE DIRECT-PAIR SECTION (the owner's SESSION-079 directive — the
  // advanced door for owners whose ISP blocks Telegram): paste the
  // Worker URL + install_id + shared_secret from a pre-paired D1
  // entry, click Direct Pair, and the app is paired without needing
  // the Telegram /pair code.
  p.appendChild(el('h4', 'form-heading',
    '⚙️ Direct Pair (advanced — no Telegram code needed)'));
  p.appendChild(el('p', 'info-header',
    'If your ISP blocks Telegram (the /pair command can\'t reach the bot), ' +
    'paste the pre-paired credentials here. The Worker URL + install_id + ' +
    'shared_secret are written directly to config.json + cloudflare_enabled ' +
    'is turned ON. Get these from your Worker\'s D1 database or your ' +
    'Cloudflare dashboard.'));

  const dpGroup = el('div', 'settings-group');
  const dpUrlRow = el('div', 'form-row');
  dpUrlRow.appendChild(label('Worker URL:'));
  const dpUrl = input('set-dp-worker-url', '',
    { placeholder: 'https://github-to-obsidian-bot.your-subdomain.workers.dev' });
  dpUrlRow.appendChild(dpUrl);
  dpGroup.appendChild(dpUrlRow);

  const dpIdRow = el('div', 'form-row');
  dpIdRow.appendChild(label('Install ID:'));
  const dpId = input('set-dp-install-id', '',
    { placeholder: 'e.g. 3a7facc8-028e-4223-9911-289f0862fc9b' });
  dpIdRow.appendChild(dpId);
  dpGroup.appendChild(dpIdRow);

  const dpSecretRow = el('div', 'form-row');
  dpSecretRow.appendChild(label('Shared Secret:'));
  const dpSecret = input('set-dp-shared-secret', '',
    { placeholder: 'e.g. 903638261d464e2c...', password: true });
  const dpBtn = btn('Direct Pair', 'primary', '', () => doDirectPair());
  dpSecretRow.append(dpSecret, dpBtn);
  dpGroup.appendChild(dpSecretRow);
  p.appendChild(dpGroup);
  return p;
}

function doDirectPair() {
  const vals = {
    workerUrl: ($('#set-dp-worker-url') || {}).value || '',
    installId: ($('#set-dp-install-id') || {}).value || '',
    sharedSecret: ($('#set-dp-shared-secret') || {}).value || '',
  };
  if (!vals.workerUrl || !vals.installId || !vals.sharedSecret) {
    miniDialog('Direct Pair — missing fields',
      'All three fields are required (Worker URL + Install ID + Shared Secret).',
      ['OK'], () => {});
    return;
  }
  DirectPair(vals).then((r) => {
    if (!r) return;
    if (r.error) {
      miniDialog('Direct Pair failed', r.error, ['OK'], () => {});
      return;
    }
    miniDialog('✅ Direct Pair — success',
      'The Worker pairing is saved + cloudflare_enabled is ON.\n\n' +
      'Click SYNC now — the app will fetch the pending items from your ' +
      'Telegram bot queue through the Cloudflare Worker (the DoH ' +
      'fallback dialer handles DNS censorship on your ISP).',
      ['OK'], () => {});
  }).catch((e) => {
    miniDialog('Direct Pair error',
      (e && e.message) ? e.message : 'unknown error', ['OK'], () => {});
  });
}

function saveBotFields() {
  return SaveBotGroup({
    username: $('#set-bot-username').value,
    token: $('#set-bot-token').value,
  });
}

function checkQueueFlow() {
  CheckBotQueue($('#set-bot-username').value).then((r) => {
    if (!r) return;
    const badge = $('#set-queue-badge');
    if (r.refusal || r.error) {
      badge.hidden = true;
      $('#set-queue-display').value = r.report || '';
      queuePayload = null;
      return;
    }
    $('#set-queue-display').value = r.report || '';
    if (r.badge) {
      badge.hidden = false;
      badge.textContent = r.badge.text;
      badge.classList.toggle('badge-ok', !r.badge.pending);
      badge.classList.toggle('badge-pending', !!r.badge.pending);
    }
    queuePayload = r;
  });
}

function processQueueFlow() {
  if (!queuePayload || (!queuePayload.pendingUrls && !queuePayload.pendingWebsites)) {
    // the honest line (the reference's empty-queue answer)
    miniDialog('Queue Empty',
      'No items in queue. Click \'Check Queue\' first.', ['OK'], () => {});
    return;
  }
  const total = queuePayload.pendingUrls.length + queuePayload.pendingWebsites.length;
  if (total > 10) {
    miniDialog('Confirm Batch',
      `Process ${total} items from the bot queue?\nThis will fetch and process every pending link.`,
      ['Process All', 'Cancel'], (answer) => {
        if (answer === 'Process All') runProcessQueue();
      });
    return;
  }
  runProcessQueue();
}

function runProcessQueue() {
  ProcessBotQueue(queuePayload.pendingUrls || [], queuePayload.pendingWebsites || [],
    queuePayload.duplicates || 0, queuePayload.rawCount || 0);
}

function buildSources() {
  const p = el('div', 'page');
  p.appendChild(el('p', 'info-header',
    '📡 Additional Sources\nFetch GitHub URLs from RSS feeds or Reddit (no API key needed).\nReddit uses the free .json endpoint (e.g. https://reddit.com/r/programming.json)'));

  const group = el('div', 'settings-group');
  const urlRow = el('div', 'form-row');
  urlRow.appendChild(label('URL:'));
  const url = input('set-sources-url', '',
    { placeholder: 'https://reddit.com/r/programming.json  OR  https://hnrss.org/frontpage' });
  urlRow.append(url, btn('Fetch URLs', 'secondary', '', () => fetchSourcesFlow()));
  group.appendChild(urlRow);

  group.appendChild(el('label', 'form-label', 'Fetched GitHub URLs:'));
  const results = el('textarea', 'form-area');
  results.id = 'set-sources-results';
  results.readOnly = true;
  results.style.minHeight = '120px';
  results.style.flex = '1';
  group.appendChild(results);

  group.appendChild(btn('Process Fetched URLs', 'primary', '', () => processSourcesFlow()));
  p.appendChild(group);
  return p;
}

let sourcesURLs = null;

function fetchSourcesFlow() {
  FetchFromSources($('#set-sources-url').value).then((r) => {
    if (!r) return;
    $('#set-sources-results').value = r.body || '';
    sourcesURLs = r.urls || null;
  });
}

function processSourcesFlow() {
  if (!sourcesURLs || !sourcesURLs.length) {
    miniDialog('No URLs', 'No URLs to process. Fetch from a source first.', ['OK'], () => {});
    return;
  }
  if (sourcesURLs.length > 10) {
    miniDialog('Confirm Batch',
      `Process ${sourcesURLs.length} URLs from sources?\nThis will fetch and process every fetched link.`,
      ['Process Fetched URLs', 'Cancel'], (answer) => {
        if (answer === 'Process Fetched URLs') ProcessSourcesUrls(sourcesURLs);
      });
    return;
  }
  ProcessSourcesUrls(sourcesURLs);
}

function buildBackup() {
  const p = el('div', 'page');
  p.appendChild(buildVaultBackupGroup());
  p.appendChild(buildSealGroup());
  p.appendChild(buildGoodReposGroup());
  p.appendChild(buildWorkerDashboardGroup());
  return p;
}

// the live field reader (the ladders' keystroke law reads unsaved edits)
function backupFieldValues() {
  return {
    folder: val('#set-backup-folder'),
    enabled: checked('#set-backup-enabled'),
    sealEnabled: checked('#set-seal-enabled'),
    sealPush: checked('#set-seal-push'),
    goodEnabled: checked('#set-good-enabled'),
    goodPush: checked('#set-good-push'),
  };
}
function val(sel) { const n = document.querySelector(sel); return n ? n.value : ''; }
function checked(sel) { const n = document.querySelector(sel); return !!(n && n.checked); }

function paintBackupLadders() {
  const v = backupFieldValues();
  RefreshBackupLadders(v.folder, v.enabled, v.sealEnabled, v.sealPush,
    v.goodEnabled, v.goodPush).then((r) => {
    if (!r) return;
    paintDot('#set-backup-dot', r.backup);
    paintDot('#set-seal-dot', r.vaultSeal);
    paintDot('#set-good-dot', r.goodRepos);
  });
}
function paintDot(sel, dot) {
  const n = document.querySelector(sel);
  if (!n || !dot) return;
  n.textContent = dot.text;
  n.className = 'vault-status status-' + (dot.state || 'muted');
}

function buildVaultBackupGroup() {
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'Vault Backup'));

  const folderRow = el('div', 'form-row');
  folderRow.appendChild(label('Backup folder:'));
  const folder = input('set-backup-folder', snap.backupFolder || '',
    { placeholder: 'where the vault zips land' });
  folder.addEventListener('change', saveBackupFields);
  folderRow.append(folder,
    btn('Browse', 'secondary', '', async () => {
      const path = await BrowseVault('backup-folder');
      if (path) { folder.value = path; saveBackupFields(); paintBackupLadders(); }
    }),
    helpButton('💡 Backup folder — where the timestamped vault zips land.\nAny folder works: a OneDrive/Dropbox/Drive sync folder uploads your zips to the cloud automatically. Zero setup, no OAuth.'));
  group.appendChild(folderRow);

  const keepRow = el('div', 'form-row');
  keepRow.appendChild(label('Keep last:'));
  const keep = input('set-backup-max', snap.backupMax || '10', { });
  keep.style.maxWidth = '60px';
  keep.addEventListener('change', saveBackupFields);
  keepRow.append(keep, label('backups'),
    helpButton('Keep last N backups — FIFO rotation: when a new zip lands and more than N exist, the OLDEST is removed. 0 or less keeps everything.'));
  const dot = el('span', 'vault-status muted-strong', '● Disabled');
  dot.id = 'set-backup-dot';
  keepRow.appendChild(dot);
  group.appendChild(keepRow);

  const btnRow = el('div', 'card-row');
  btnRow.appendChild(btn('Backup Now', 'primary', '', () => backupNowFlow()));
  const restore = btn('Restore', 'secondary',
    'Pick a backup ZIP and extract it to a NEW folder next to the current vault — the current vault is never touched',
    () => restoreFlow());
  btnRow.appendChild(restore);
  const enabledBox = el('input', 'form-check');
  enabledBox.id = 'set-backup-enabled';
  enabledBox.type = 'checkbox';
  enabledBox.checked = !!snap.backupEnabled;
  enabledBox.title = 'Create a timestamped vault ZIP after every processing batch';
  enabledBox.addEventListener('change', () => { saveBackupFields(); paintBackupLadders(); });
  const enabledLabel = el('label', 'form-label', 'Enable auto-backup after each batch');
  enabledLabel.prepend(enabledBox);
  btnRow.appendChild(enabledLabel);
  group.appendChild(btnRow);
  return group;
}

function buildSealGroup() {
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'VaultSeal — GitHub Mirror'));

  const checksRow = el('div', 'form-row');
  const sealEnabled = el('input', 'form-check');
  sealEnabled.id = 'set-seal-enabled';
  sealEnabled.type = 'checkbox';
  sealEnabled.checked = snap.vaultSealEnabled !== false;
  sealEnabled.title = 'Commit the whole vault to git after every run and push it to a PRIVATE GitHub repository';
  sealEnabled.addEventListener('change', () => { saveBackupFields(); paintBackupLadders(); });
  const sealLabel = el('label', 'form-label', 'Seal after every run');
  sealLabel.prepend(sealEnabled);
  checksRow.appendChild(sealLabel);

  const sealPush = el('input', 'form-check');
  sealPush.id = 'set-seal-push';
  sealPush.type = 'checkbox';
  sealPush.checked = snap.vaultSealPush !== false;
  sealPush.title = 'Push the seal to GitHub (uses the GitHub token from Settings)';
  sealPush.addEventListener('change', saveBackupFields);
  const pushLabel = el('label', 'form-label', 'Push to GitHub');
  pushLabel.prepend(sealPush);
  checksRow.append(pushLabel,
    helpButton('💡 VaultSeal — Obsidian\'s free tier has no sync.\nVaultSeal commits the whole vault after every run and pushes it to a PRIVATE repository (full history, restore with git clone). Machine state (workspace.json, .trash) is excluded; an unchanged vault is a no-op.'));
  group.appendChild(checksRow);

  const repoRow = el('div', 'form-row');
  repoRow.appendChild(label('Backup repo:'));
  const repo = input('set-seal-repo', snap.vaultSealRepo || '',
    { placeholder: 'required — the private backup repo name (no default; remembered like credentials once you type it)' });
  repo.addEventListener('change', saveBackupFields);
  repoRow.append(repo,
    btn('📡 Detect Repos', 'secondary',
      'THE REPO RADAR (v0.84.0) — find the backup repos that ALREADY exist for the three vaults and set them automatically:\n① the GitHub vault\'s repo (this field), ② the Websites vault\'s repo, ③ the Personal vault\'s repo.\n\nEvidence ladder: the vault\'s own git remote → a repo named like the vault folder → a content fingerprint (blob-SHA match against every candidate repo).\n\nRead-only: nothing is created, pushed, or overwritten — names you have typed yourself always win; ambiguous cases are reported, never guessed.',
      () => DetectRepos()));
  group.appendChild(repoRow);

  const btnRow = el('div', 'card-row');
  btnRow.append(
    btn('Seal Now', 'secondary', '', () => SealNow('repos')),
    btn('Resync Mirror', 'secondary',
      'Make the GitHub backup repos match the vaults EXACTLY — files you deleted in Obsidian (e.g. wrongly-created notes) are deleted from the mirror too. History is kept; nothing is force-pushed. Asks for confirmation first.',
      () => {
        miniDialog('Resync Mirror',
          'Make the GitHub backup repos match the vaults EXACTLY?\nFiles you deleted in Obsidian are deleted from the mirror too. History is kept; nothing is force-pushed.',
          ['Resync', 'Cancel'], (answer) => {
            if (answer === 'Resync') ResyncMirror();
          });
      }));
  const dot = el('span', 'vault-status muted-strong', '● —');
  dot.id = 'set-seal-dot';
  btnRow.appendChild(dot);
  group.appendChild(btnRow);
  return group;
}

function buildGoodReposGroup() {
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'Good Repos — Public Directory'));

  const checksRow = el('div', 'form-row');
  const goodEnabled = el('input', 'form-check');
  goodEnabled.id = 'set-good-enabled';
  goodEnabled.type = 'checkbox';
  goodEnabled.checked = snap.goodReposEnabled !== false;
  goodEnabled.title = 'Publish the curated directory to a PUBLIC GitHub repo after every run';
  goodEnabled.addEventListener('change', () => { saveBackupFields(); paintBackupLadders(); });
  const goodLabel = el('label', 'form-label', 'Publish after every run');
  goodLabel.prepend(goodEnabled);
  checksRow.appendChild(goodLabel);

  const goodPush = el('input', 'form-check');
  goodPush.id = 'set-good-push';
  goodPush.type = 'checkbox';
  goodPush.checked = snap.goodReposPush !== false;
  goodPush.title = 'Push the directory to GitHub (uses the GitHub token from Settings)';
  goodPush.addEventListener('change', saveBackupFields);
  const pushLabel = el('label', 'form-label', 'Push to GitHub');
  pushLabel.prepend(goodPush);
  checksRow.append(pushLabel,
    helpButton('💡 Good Repos — every curated repo becomes an entry in a browsable, emoji-rich README directory (organized like AI → Skills → …) with the full notes mirrored into category folders.\nPUBLIC by design — everyone can browse and benefit from your curation.'));
  group.appendChild(checksRow);

  const repoRow = el('div', 'form-row');
  repoRow.appendChild(label('Directory repo:'));
  const repo = input('set-good-repo', snap.goodReposRepo || '',
    { placeholder: 'required — empty = publishing is off (no default, no auto-creation)' });
  repo.addEventListener('change', saveBackupFields);
  repoRow.appendChild(repo);
  group.appendChild(repoRow);

  const btnRow = el('div', 'card-row');
  btnRow.appendChild(btn('Publish Now', 'secondary', '', () => PublishNow()));
  const dot = el('span', 'vault-status muted-strong', '● —');
  dot.id = 'set-good-dot';
  btnRow.appendChild(dot);
  group.appendChild(btnRow);
  return group;
}

function buildWorkerDashboardGroup() {
  const group = el('div', 'settings-group');
  group.appendChild(el('div', 'group-title', 'Dashboard'));

  const dashRow = el('div', 'form-row');
  dashRow.appendChild(label('Worker URL:'));
  const url = input('set-worker-url', snap.workerUrl || '',
    { placeholder: 'https://github-to-obsidian-bot.your-subdomain.workers.dev' });
  url.addEventListener('change', saveBackupFields);
  dashRow.append(url,
    btn('Open Dashboard', 'secondary', '', () => {
      OpenDashboard($('#set-worker-url').value).then((r) => {
        // the wails multi-return lands as an array
        const title = Array.isArray(r) ? r[0] : (r && r[0]);
        const body = Array.isArray(r) ? r[1] : (r && r[1]);
        if (title) miniDialog(title, body || '', ['OK'], () => {});
      });
    }),
    helpButton('📊 Dashboard — bot stats, pending links and recent activity, served by your Cloudflare Worker (no separate deployment needed).'));
  group.appendChild(dashRow);
  return group;
}

function saveBackupFields() {
  return SaveBackupGroup({
    backupFolder: val('#set-backup-folder'),
    backupMax: val('#set-backup-max'),
    backupEnabled: checked('#set-backup-enabled'),
    workerUrl: val('#set-worker-url'),
    vaultSealEnabled: checked('#set-seal-enabled'),
    vaultSealPush: checked('#set-seal-push'),
    vaultSealRepo: val('#set-seal-repo'),
    goodReposEnabled: checked('#set-good-enabled'),
    goodReposPush: checked('#set-good-push'),
    goodReposRepo: val('#set-good-repo'),
  });
}

function backupNowFlow() {
  BackupNowGuards().then((r) => {
    const title = Array.isArray(r) ? r[0] : (r && r[0]);
    const body = Array.isArray(r) ? r[1] : (r && r[1]);
    if (title) { miniDialog(title, body || '', ['OK'], () => {}); return; }
    BackupNow();
  });
}

// ---------------------------------------------------------------------------
// THE RESTORE DIALOG (T-80's standing assignment — backup_seal.py's
// _backup_restore: the page saves first, then the list, the ⚠⃣ confirm,
// the extract-to-new-folder safety net, the copy door).
// ---------------------------------------------------------------------------

function restoreFlow() {
  // The reference's own order (_backup_restore opens with
  // _backup_save_config — the page's fields are committed before the
  // backup folder is read).
  saveBackupFields().then(() => {
    RestoreBackupsList().then((m) => {
      if (!m) return;
      if (m.boxTitle) {
        miniDialog(m.boxTitle, m.boxBody || '', ['OK'], () => {});
        return;
      }
      if (m.backups && m.backups.length) openRestoreDialog(m.backups);
    });
  });
}

function openRestoreDialog(backups) {
  const body = el('div');
  body.appendChild(el('p', 'dialog-message',
    'Found ' + backups.length + ' backup(s):'));
  const list = el('div', 'restore-list');
  list.style.maxHeight = '260px';
  list.style.overflowY = 'auto';
  let selected = -1;
  backups.forEach((b, i) => {
    const row = el('button', 'restore-row', b.name + '    [ ' + b.size + ' ]');
    row.type = 'button';
    row.addEventListener('click', () => {
      selected = i;
      list.querySelectorAll('.restore-row').forEach((n) =>
        n.classList.remove('selected'));
      row.classList.add('selected');
    });
    list.appendChild(row);
  });
  body.appendChild(list);
  miniCustom('Restore from Backup', body,
    [['Restore to New Folder', 'primary'], ['Copy ZIP to...', 'secondary'], ['Close', 'secondary']],
    (answer) => {
      if (selected < 0) return;                 // the reference's currentRow < 0 law
      const pick = backups[selected].name;
      if (answer === 'Restore to New Folder') {
        miniDialog('⚠️ Restore',
          'Restore from:\n' + pick + '\n\n' +
          'This will extract to a NEW folder next to your current vault\n' +
          '(current vault is NOT touched). Continue?',
          ['Yes', 'Cancel'], (yesNo) => {
            if (yesNo !== 'Yes') return;
            RestoreBackupRun(pick).then((res) => {
              if (res && res.title) miniDialog(res.title, res.body, ['OK'], () => {});
            });
          });
      } else if (answer === 'Copy ZIP to...') {
        CopyBackupTo(pick).then((res) => {
          if (res && res.title) miniDialog(res.title, res.body, ['OK'], () => {});
        });
      }
    });
}

// ---------------------------------------------------------------------------
// T-75 — the Sync page (sync_flow.py + sync_view.py, the HTML twin):
// THE VISUAL SYNC DISPLAY (the four-element flow cards, two per row)
// above THE DEFERRED LEDGER (the stage-by-stage gap matrix) + the
// differences modal. ONE Refresh measures both desks (v0.87.0's GRAND
// CONVERGENCE); read-only by law — the click is the consent for every
// network leg.
// ---------------------------------------------------------------------------

const SYNC_DIFF_CAP_ROWS = 400;   // _DIFF_LIST_CAP
const SYNC_CARD_MAX_W = 300;      // SYNC_CARD_MAX_W (the 2×2 grid law)

// the stage tables (fetched once at build; the reference's own shape)
let syncTables = null;
let syncPayload = null;           // the last Refresh's payload
let syncBusy = false;             // one measurement at a time
const syncDiff = { open: false, key: '', stage: null, checked: new Set(), reason: '' };

function buildSync() {
  const p = el('div', 'page');

  const header = el('p', 'info-header wordwrap',
    "Sync — the pipeline at a glance. The four elements your library " +
    "flows through, two per row — each card speaks its own truth: what " +
    "is waiting, what is stale, what is not set. Read-only: nothing is " +
    "sealed, published or moved from here (the hints name the doors).");
  p.appendChild(header);

  // -- the Refresh row (the consent for every network leg) ----------
  const refreshRow = el('div', 'card-row');
  const refreshBtn = btn('🔄 Refresh', 'primary',
    "Measure the WHOLE page again — the flow's four elements AND the " +
    "stage matrix below (v0.87.0's one consent door: the local walks — " +
    "vaults, git seal statuses, the public-ready count — always run; " +
    "the live doors (the bot queue, the published site) are asked ONLY " +
    "because you clicked. Nothing is ever changed by a refresh.)",
    () => refreshSyncView());
  refreshBtn.id = 'sync-refresh-btn';
  const measured = el('span', 'muted-note', 'Not measured yet');
  measured.id = 'sync-measured-note';
  refreshRow.append(refreshBtn, measured);
  p.appendChild(refreshRow);

  // -- THE 2×2 COMPACT GRID (v0.89.0 — the owner's ask: two per row;
  // the Z-reading IS the pipeline) -----------------------------------
  const grid = el('div', 'sync-grid');
  const pairs = [
    ['bot', 'vault', 'data flows this way'],
    ['github', 'website', "the vault's public-ready notes publish to the site"],
  ];
  for (const [a, b, tip] of pairs) {
    for (const key of [a, b]) {
      const card = el('div', 'sync-card');
      card.id = 'sync-card-' + key;
      card.style.maxWidth = SYNC_CARD_MAX_W + 'px';
      grid.appendChild(card);
    }
    const arrow = el('span', 'sync-arrow', '→');
    arrow.title = tip;
    grid.appendChild(arrow);
  }
  p.appendChild(grid);

  // -- the calm line + the legend -------------------------------------
  const calm = el('p', 'sync-calm');
  calm.id = 'sync-calm';
  calm.hidden = true;
  p.appendChild(calm);
  p.appendChild(el('p', 'muted-note wordwrap',
    "Links arrive at the bot → become notes in the vault → seal to " +
    "GitHub → publish to the website."));

  // -- THE DEFERRED LEDGER: the matrix section (the page's SECOND desk)
  const matrix = el('div', 'settings-group');
  matrix.appendChild(el('div', 'group-title', 'The chain, stage by stage'));
  matrix.appendChild(el('p', 'info-header wordwrap',
    "The chain, stage by stage — Telegram queue → vaults → backups → " +
    "the public website. Every cell speaks its number AND what the " +
    "number is; a gap is not bad: it either gets fixed, or it gets " +
    "reviewed and explained. Reviewed gaps never warn again; they stay " +
    "visible as a calm count."));

  const banner = el('div', 'sync-banner');
  banner.dataset.tone = 'muted';
  const bannerIcon = el('span', 'sync-banner-icon');
  const bannerText = el('span', 'sync-banner-text', 'Not measured yet');
  banner.append(bannerIcon, bannerText);
  banner.id = 'sync-banner';
  matrix.appendChild(banner);

  const table = el('div', 'sync-matrix');
  table.id = 'sync-matrix';
  matrix.appendChild(table);

  matrix.appendChild(el('p', 'muted-note wordwrap',
    "How to read a cell: the big number is the count that reached the " +
    "target; the small line under it names both sides (e.g. “720 in " +
    "the vault · 710 on the site”). ✓ Match — counts agree · ⚠ Behind " +
    "— unreviewed gap (click to review) · ⓘ Reviewed — explained, " +
    "calm count (click to revisit) · — Unknown — the reason is written " +
    "on the cell; click it for its door."));
  const lastLine = el('p', 'muted-note',
    'Not measured yet — open this page or click Refresh to measure.');
  lastLine.id = 'sync-last-line';
  matrix.appendChild(lastLine);
  p.appendChild(matrix);

  // the placeholder paint (never a fabricated count)
  StageTables().then((t) => { syncTables = t; renderSyncMatrix(); });
  renderSyncPlaceholder();
  return p;
}

// the never-refreshed state: honest "not measured yet" cards
function renderSyncPlaceholder() {
  for (const key of ['bot', 'vault', 'github', 'website']) {
    renderSyncCard(key, {
      key, icon: '·', name: key.charAt(0).toUpperCase() + key.slice(1),
      dotTone: 'muted', statusText: 'not measured yet',
      hero: { value: '—', label: 'not measured yet', tone: 'muted' },
      stats: [], foot: [], lines: ['click Refresh to measure'], hint: '',
    });
  }
}

// one card from the model's dict (the Qt layer's twin — paint only)
function renderSyncCard(key, model) {
  const card = document.getElementById('sync-card-' + key);
  if (!card) return;
  card.textContent = '';
  const head = el('div', 'sync-card-head');
  const dot = el('span', 'sync-dot');
  dot.dataset.tone = model.dotTone || model.dot_tone || 'muted';
  dot.title = model.statusText || model.status_text || '';
  head.append(dot, el('span', 'sync-card-name',
    (model.icon || '🗂️') + ' ' + (model.name || 'Element')));
  head.lastChild.title = model.statusText || model.status_text || '';
  card.appendChild(head);

  const heroRow = el('div', 'sync-hero-row');
  const heroVal = el('span', 'sync-hero', model.hero ? model.hero.value : '—');
  heroVal.dataset.tone = (model.hero && model.hero.tone) || 'muted';
  const heroLabel = el('span', 'sync-hero-label',
    model.hero ? model.hero.label : '');
  heroRow.append(heroVal, heroLabel);
  card.appendChild(heroRow);

  for (const stat of (model.stats || [])) {
    const row = el('div', 'sync-stat-row');
    const val = el('span', 'sync-stat-value', String(stat.value));
    if (stat.tone) val.dataset.tone = stat.tone;
    row.append(el('span', 'sync-stat-label', stat.label || ''), val);
    card.appendChild(row);
    if (stat.note) card.appendChild(el('p', 'sync-stat-note', stat.note));
  }
  for (const line of (model.foot || [])) {
    card.appendChild(el('p', 'sync-foot wordwrap', line));
  }
  if (model.hint) card.appendChild(el('p', 'muted-note wordwrap', model.hint));
}

// THE ONE CONSENT DOOR — one Refresh measures both desks
function refreshSyncView() {
  if (syncBusy) return;  // a second click while measuring is swallowed
  syncBusy = true;
  const btn = document.getElementById('sync-refresh-btn');
  if (btn) btn.disabled = true;
  // the transient measuring state (the honest "· · ·")
  for (const key of ['bot', 'vault', 'github', 'website']) {
    renderSyncCard(key, {
      key, icon: '·', name: key.charAt(0).toUpperCase() + key.slice(1),
      dotTone: 'muted', statusText: 'measuring…',
      hero: { value: '· · ·', label: 'measuring', tone: 'muted' },
      stats: [], foot: [],
      lines: ['the local walks run first, then the live doors (only because you clicked)'],
      hint: '',
    });
  }
  SyncRefresh().then((payload) => {
    syncBusy = false;
    if (btn) btn.disabled = false;
    if (!payload) return;
    syncPayload = payload;
    for (const cardModel of (payload.model && payload.model.cards) || []) {
      renderSyncCard(cardModel.key, cardModel);
    }
    const calm = document.getElementById('sync-calm');
    if (calm) {
      calm.textContent = payload.model ? payload.model.calmLine || payload.model.calm_line : '';
      calm.hidden = !calm.textContent;
    }
    renderSyncMatrix();
    const measured = document.getElementById('sync-measured-note');
    const stamp = payload.measuredAt || payload.measured_at || '';
    if (measured && stamp) {
      measured.textContent =
        `measured at ${stamp.substring(11, 19)} — the numbers are as fresh as this line`;
    }
    const syncPage = document.querySelector('[data-page="sync"]');
    if (syncPage) iconify(syncPage);   // the re-rendered emojis ride as glyphs
  }).catch(() => {
    syncBusy = false;
    if (btn) btn.disabled = false;
    const measured = document.getElementById('sync-measured-note');
    if (measured) {
      measured.textContent = 'measurement failed — the cards keep their last truth';
    }
  });
}

// the matrix paint (the model's snapshot → the six cells)
function renderSyncMatrix() {
  const table = document.getElementById('sync-matrix');
  if (!table || !syncTables) return;
  table.textContent = '';
  const stages = (syncPayload && syncPayload.snapshot) ? syncPayload.snapshot.stages : null;

  const headRow = el('div', 'sync-matrix-row sync-matrix-head');
  headRow.appendChild(el('span', 'sync-matrix-corner', 'Stream'));
  for (const col of syncTables.columns) {
    headRow.appendChild(el('span', 'sync-matrix-colhead', col));
  }
  table.appendChild(headRow);

  for (const stream of syncTables.streams) {
    const row = el('div', 'sync-matrix-row');
    row.appendChild(el('span', 'sync-matrix-rowhead', stream));
    for (const key of syncTables.layout[stream] || []) {
      if (!key) {
        const na = el('span', 'sync-cell-na', '—');
        na.title = 'not a stage for this stream (the queue feeds the ' +
          'GitHub and Websites pipelines; the website publishes the ' +
          'Websites vault; every vault seals to its own mirror)';
        row.appendChild(na);
        continue;
      }
      const stage = stages ? stages[key] : null;
      row.appendChild(buildSyncCell(key, stream, stage));
    }
    table.appendChild(row);
  }
  renderSyncBanner();
  renderSyncLastLine();
}

// one matrix cell — the big number + the micro-copy + the state wash
function buildSyncCell(key, stream, stage) {
  const state = stage ? stage.state : 'unknown';
  const glyph = state === 'match' ? '✓ ' :
    state === 'behind' ? '⚠ ' :
    state === 'reviewed' ? 'ⓘ ' : '— ';
  let num = '—', cap = 'not measured yet';
  if (stage) {
    // the caption's decision lives in the engine (syncreview's law);
    // the paint rides the same branches the core computed
    num = cellNumOf(stage);
    cap = cellCapOf(stage);
  }
  const cell = el('button', 'sync-cell');
  cell.dataset.cellState = state;
  cell.dataset.key = key;
  const row = el('div', 'sync-cell-row');
  const numEl = el('span', 'sync-cell-num', num);
  numEl.dataset.cellState = state;
  row.append(numEl, el('span', 'sync-cell-icon',
    state === 'behind' ? '⚠' : state === 'reviewed' ? 'ⓘ' :
    state === 'match' ? '✓' : ''));
  cell.appendChild(row);
  const capEl = el('span', 'sync-cell-cap', glyph + cap);
  capEl.dataset.cellState = state;
  cell.appendChild(capEl);
  cell.title = cellTitleOf(stream, stage, state, num, cap);
  cell.setAttribute('aria-label', `${stream} ${stage ? stage.stage : ''} — ${cap}`);
  if (state === 'match') {
    cell.disabled = true;
  } else {
    cell.addEventListener('click', () => onSyncCellClicked(key, state));
  }
  return cell;
}

// the caption's pure branches (the engine's own twin — the frontend
// renders what the core decided; these two helpers mirror
// syncreview.CellCaption for the paint only, byte-for-byte)
function cellNumOf(stage) {
  if (stage.state === 'unknown') return '—';
  const disp = stageNum(stage, 'displayNumber', 'display_number');
  return disp == null ? '—' : String(disp);
}

function stageNum(stage, camel, snake) {
  const v = stage[camel] != null ? stage[camel] : stage[snake];
  return v != null ? v : null;
}

function cellCapOf(stage) {
  if (stage.state === 'unknown') {
    const reason = (stage.reason || '').trim();
    return reason || 'not set up or not measured';
  }
  const gap = stage.gapCount || stage.gap_count || 0;
  const deferred = stage.deferredGapCount || stage.deferred_gap_count || 0;
  const src = stageNum(stage, 'sourceCount', 'source_count');
  const tgt = stageNum(stage, 'targetCount', 'target_count');
  const v = src == null ? '—' : String(src);
  const p = tgt == null ? '—' : String(tgt);
  const disp = stageNum(stage, 'displayNumber', 'display_number');
  const inVault = disp == null ? 0 : disp;
  if (stage.key.endsWith(':queue_to_vault')) {
    if (stage.state === 'behind') return `${gap} waiting in the queue · ${inVault} already in the vault`;
    if (stage.state === 'reviewed') return `${deferred} deferred (explained) · ${inVault} in the vault`;
    return `${inVault} in the vault · ${gap} waiting in the queue`;
  }
  if (stage.key.endsWith(':vault_to_site')) {
    if (stage.state === 'behind') return `${gap} to publish · ${v} in the vault · ${p} on the site`;
    if (stage.state === 'reviewed') return `${deferred} deferred (explained) · ${v} in the vault · ${p} on the site`;
    return `${v} in the vault = ${p} on the site`;
  }
  const sealed = disp == null ? 0 : disp;
  const total = sealed + gap;
  if (stage.state === 'behind') return `${gap} unsealed · ${total} in the vault · ${sealed} sealed in the mirror`;
  if (stage.state === 'reviewed') return `${deferred} deferred (explained) · ${total} in the vault · ${sealed} sealed in the mirror`;
  return `${total} in the vault = ${sealed} in the mirror`;
}

function cellTitleOf(stream, stage, state, num, cap) {
  const st = stage || { stage: '', stream, sourceLabel: 'the source', targetLabel: 'the target' };
  if (state === 'match') {
    return `${st.stream} · ${st.stage}: counts agree (${st.sourceCount != null ? st.sourceCount : st.source_count} in ${st.sourceLabel || st.source_label}, ${st.targetCount != null ? st.targetCount : st.target_count} in ${st.targetLabel || st.target_label}).`;
  }
  if (state === 'behind') {
    const unexplained = st.unexplainedCount || st.unexplained_count || 0;
    return `${st.stream} · ${st.stage}: ${unexplained} unexplained item${unexplained === 1 ? '' : 's'} — click to review (explain) or fix. ${st.setUpHint || ''}`;
  }
  if (state === 'reviewed') {
    const n = st.deferredGapCount || st.deferred_gap_count || 0;
    return `${st.stream} · ${st.stage}: ${n} item${n === 1 ? '' : 's'} deferred — every gap explained (click to revisit the log; Reopen brings any item back).`;
  }
  return `${st.stream} · ${st.stage}: ${cap} — click to act. ${st.setUpHint || st.set_up_hint || 'not measured yet'}`;
}

// the banner's verdict (the engine's own — the paint rides the payload)
function renderSyncBanner() {
  const banner = document.getElementById('sync-banner');
  if (!banner) return;
  const verdict = syncPayload ? syncPayload.banner : null;
  const tone = verdict ? verdict.tone : 'muted';
  const text = verdict ? verdict.text : 'Not measured yet';
  banner.dataset.tone = tone;
  banner.textContent = '';
  const icon = el('span', 'sync-banner-icon',
    tone === 'success' ? '✓' : tone === 'warning' ? '⚠' : '●');
  banner.append(icon, el('span', 'sync-banner-text', text));
  const tips = {
    success: 'Healthy: every stage matches or is fully explained. A deferred count never disappears — it is the calm record of your reviews.',
    warning: 'At least one stage has an unreviewed gap — the banner names each one; open the cell to review or fix it. After review it never warns again.',
    muted: 'Not set up (a Settings door is missing — e.g. the Cloudflare Worker URL, which no LLM / scan / fetch ever sets) or not measured (a check that has not run this session, e.g. the bot queue). The banner names each stage and its reason — click the cell to act.',
  };
  banner.title = tips[tone] || '';
}

function renderSyncLastLine() {
  const line = document.getElementById('sync-last-line');
  if (!line) return;
  if (!syncPayload || !syncPayload.snapshot) return;
  const shown = (syncPayload.snapshot.generatedAt || syncPayload.snapshot.generated_at || '').substring(0, 19).replace('T', ' ');
  LedgerCount().then((n) => {
    line.textContent =
      `Measured ${shown || '—'} (UTC) · the deferred log remembers ${n} review${n === 1 ? '' : 's'} — nothing is ever deleted.`;
  });
}

// the cell click: Behind/Reviewed open the modal; Unknown runs the
// stage's set-up door (measure / fix / teach)
function onSyncCellClicked(key, state) {
  if (state === 'behind' || state === 'reviewed') {
    openSyncDifferences(key);
    return;
  }
  runSyncFixDoor(key);
}

// the FIX door for one stage (the "it gets fixed" half of the law)
function runSyncFixDoor(key) {
  const stage = syncPayload && syncPayload.snapshot ?
    syncPayload.snapshot.stages[key] : null;
  const action = stage ? stage.fixAction || stage.fix_action : '';
  if (action === 'check_queue') {
    SyncCheckQueue().then((r) => {
      if (r && r.report && !r.refusal && !r.error) refreshSyncView();
    });
    return;
  }
  if (action === 'publish') {
    SyncPublishNow().then(() => refreshSyncView());
    return;
  }
  if (action === 'seal') {
    const kind = key.startsWith('repos:') ? 'repos' :
      key.startsWith('websites:') ? 'websites' : 'personal';
    SyncSealNow(kind).then(() => {
      // the repos seal re-measure arming (singleShot(4000)'s twin)
      setTimeout(() => refreshSyncView(), 4000);
    });
    return;
  }
  refreshSyncView();  // unknown / no fix action: measure + teach
}

// ---------------------------------------------------------------------------
// THE DIFFERENCES MODAL — where gaps get resolved
// ---------------------------------------------------------------------------

function openSyncDifferences(key) {
  StageDifferences(key).then((payload) => {
    if (!payload || !payload.found) return;
    syncDiff.open = true;
    syncDiff.key = key;
    syncDiff.stage = payload.stage;
    syncDiff.checked = new Set();
    syncDiff.reason = '';
    renderSyncDiffModal(payload);
  });
}

function syncModalRoot() {
  let root = document.getElementById('sync-diff-modal');
  if (!root) {
    root = el('div', 'dialog-overlay');
    root.id = 'sync-diff-modal';
    root.hidden = true;
    document.getElementById('app').appendChild(root);
  }
  return root;
}

function renderSyncDiffModal(payload) {
  const root = syncModalRoot();
  root.textContent = '';
  root.hidden = false;
  const stage = payload.stage;

  const dlg = el('div', 'sync-diff-dialog');
  dlg.setAttribute('role', 'dialog');
  dlg.setAttribute('aria-modal', 'true');
  dlg.setAttribute('aria-label', `${stage.stream} · ${stage.stage}`);

  // the header row: the story + the stage's fix door
  const head = el('div', 'sync-diff-head');
  head.appendChild(el('p', 'sync-diff-heading wordwrap', payload.heading));
  const fixAction = stage.fixAction || stage.fix_action || '';
  if (fixAction && fixAction !== 'none') {
    head.appendChild(btn(stage.fixLabel || stage.fix_label || 'Fix now', 'secondary',
      "The FIX door — how this stage's gap gets closed (the other half " +
      "of the law: a gap either gets fixed, or it gets reviewed and " +
      "explained).", () => { closeSyncDiff(); runSyncFixDoor(syncDiff.key); }));
  }
  dlg.appendChild(head);

  // the tabs
  const tabs = el('div', 'sync-tabs');
  const tabBtns = el('div', 'sync-tab-btns');
  const diffTab = btn(`Differences (${(stage.items || []).length})`, 'secondary', '', () => selectSyncTab('diff'));
  const logTab = btn(`Deferred log (${payload.logCount || payload.log_count})`, 'secondary', '', () => selectSyncTab('log'));
  diffTab.classList.add('sync-tab-btn');
  logTab.classList.add('sync-tab-btn');
  tabBtns.append(diffTab, logTab);
  tabs.appendChild(tabBtns);

  const diffPage = buildSyncDiffTab(stage);
  const logPage = buildSyncLogTab(payload);
  tabs.append(diffPage, logPage);
  dlg.appendChild(tabs);

  // the bottom row: the confirm line + Close
  const bottom = el('div', 'sync-diff-bottom');
  const confirmLine = el('p', 'sync-confirm-line wordwrap', '');
  confirmLine.id = 'sync-confirm-line';
  bottom.append(confirmLine,
    btn('Close', 'secondary', 'Close the differences dialog', closeSyncDiff));
  dlg.appendChild(bottom);
  root.appendChild(dlg);

  // a Reviewed cell opens onto the LOG (the story is the explanation)
  selectSyncTab(stage.state === 'reviewed' ? 'log' : 'diff');
  refreshSyncMarkButton();
}

function selectSyncTab(which) {
  for (const b of document.querySelectorAll('.sync-tab-btn')) {
    b.classList.toggle('sync-tab-active', b.textContent.startsWith(
      which === 'diff' ? 'Differences' : 'Deferred log'));
  }
  const diffPage = document.getElementById('sync-diff-page');
  const logPage = document.getElementById('sync-log-page');
  if (diffPage) diffPage.hidden = which !== 'diff';
  if (logPage) logPage.hidden = which !== 'log';
}

// the Differences tab: Select all + the rows + the WHY bar
function buildSyncDiffTab(stage) {
  const page = el('div', 'sync-tab-page');
  page.id = 'sync-diff-page';

  const master = el('label', 'sync-master');
  const masterBox = el('input');
  masterBox.type = 'checkbox';
  master.append(masterBox, el('span', '', 'Select all'));
  master.title = 'Tick or untick every row at once — then pick ONE reason for all of them';
  masterBox.addEventListener('click', () => {
    // the plan_master law: any unticked row ticks the whole list
    const rows = document.querySelectorAll('.sync-row-check');
    const target = !Array.from(rows).every((r) => r.checked);
    for (const r of rows) r.checked = target;
    syncDiff.checked = new Set();
    if (target) {
      for (const it of stage.items || []) syncDiff.checked.add(it.id);
    }
    refreshSyncMarkButton();
  });
  page.appendChild(master);

  const list = el('div', 'sync-diff-list');
  const items = (stage.items || []).slice(0, SYNC_DIFF_CAP_ROWS);
  for (const item of items) {
    list.appendChild(buildSyncDiffRow(item, stage));
  }
  if (items.length === 0) {
    list.appendChild(el('p', 'muted-note wordwrap',
      'Nothing waiting here — every gap item is explained (see the Deferred log tab).'));
  }
  page.appendChild(list);

  // the WHY bar: four chips + note + Mark (the spec's law)
  const whyRow = el('div', 'card-row');
  whyRow.appendChild(el('span', 'sync-why', 'Why?'));
  const chipTips = {
    not_ready: 'keeping it for later — it will ride the next batch/publish once ready',
    private: 'this one must never publish publicly',
    duplicate: 'the same item already exists on the target side under another name',
    other: 'anything else — write the one-line note',
  };
  const chipLabels = { not_ready: 'Not ready', private: 'Private', duplicate: 'Duplicate', other: 'Other' };
  for (const reason of ['not_ready', 'private', 'duplicate', 'other']) {
    const chip = btn(chipLabels[reason], 'secondary',
      chipTips[reason] || '', () => {
        // exclusive single choice (the chip group's law)
        syncDiff.reason = syncDiff.reason === reason ? '' : reason;
        for (const c of document.querySelectorAll('.sync-chip')) {
          c.classList.toggle('sync-chip-on',
            c.dataset.reason === syncDiff.reason);
        }
        refreshSyncMarkButton();
      });
    chip.classList.add('sync-chip');
    chip.dataset.reason = reason;
    chip.setAttribute('aria-label', 'Reason: ' + chipLabels[reason]);
    whyRow.appendChild(chip);
  }
  page.appendChild(whyRow);

  const noteRow = el('div', 'card-row');
  const noteInput = el('input', 'form-input');
  noteInput.id = 'sync-note-input';
  noteInput.placeholder = 'short note (optional — one line, remembered in the log)';
  noteInput.addEventListener('input', () => refreshSyncMarkButton());
  const markBtn = btn('Mark as intentional', 'primary',
    'Apply the chosen reason (and note) to every checked item — they ' +
    'leave the Differences list, the cell turns Reviewed, and the ' +
    'reason is remembered (Reopen brings any of them back).',
    () => syncMarkIntentional());
  markBtn.id = 'sync-mark-btn';
  markBtn.disabled = true;   // THE disabled-until law
  noteRow.append(noteInput, markBtn);
  page.appendChild(noteRow);
  return page;
}

// one differences row: checkbox + name + relative time + (deferred
// chip) + Open note — one line always, the full story on the hover
function buildSyncDiffRow(item, stage) {
  const row = el('div', 'sync-row');
  const box = el('input', 'sync-row-check');
  box.type = 'checkbox';
  box.dataset.itemId = item.id;
  const name = item.name || item.id || '';
  box.title = name;
  const label = el('label', 'sync-row-name wordwrap');
  const short = name.length > 48 ? name.slice(0, 23) + '…' + name.slice(-22) : name;
  label.append(box, el('span', '', short));
  row.appendChild(label);
  const ago = el('span', 'sync-row-ago', item.addedAt ? agoOf(item.addedAt) : agoOf(item.added_at));
  ago.title = 'added at ' + (item.addedAt || item.added_at || '—');
  row.appendChild(ago);
  const reviewed = syncPayload && syncPayload.snapshot &&
    syncPayload.snapshot.stages[stage.key] &&
    (syncPayload.snapshot.stages[stage.key].deferred_gap_ids || []).includes(item.id);
  if (reviewed) {
    const chip = el('span', 'sync-row-chip', 'deferred');
    chip.title = 'already explained';
    row.appendChild(chip);
  }
  if (item.notePath || item.note_path) {
    row.appendChild(btn('Open note', 'secondary',
      'Open the note to verify before excusing — that is the point of a review.',
      () => OpenNoteDoor(item.notePath || item.note_path)));
  }
  box.addEventListener('change', () => {
    if (box.checked) syncDiff.checked.add(item.id);
    else syncDiff.checked.delete(item.id);
    refreshSyncMarkButton();
  });
  return row;
}

// the compact relative form (the engine's own grammar, the paint twin)
function agoOf(stamp) {
  const m = /^(\d{4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2})(?::(\d{2}))?)?/.exec(String(stamp || '').trim());
  if (!m) return '—';
  const then = Date.UTC(+m[1], +m[2] - 1, +m[3], +(m[4] || 0), +(m[5] || 0), +(m[6] || 0));
  let seconds = (Date.now() - then) / 1000;
  if (seconds < 0) seconds = 0;
  if (seconds < 60) return 'just now';
  const minutes = Math.floor(seconds / 60);
  if (minutes < 60) return `${minutes} m ago`;
  const hours = Math.floor(minutes / 60);
  if (hours < 24) return `${hours} h ago`;
  const days = Math.floor(hours / 24);
  if (days < 30) return `${days} d ago`;
  const months = Math.floor(days / 30);
  if (months < 12) return `${months} mo ago`;
  return `${Math.floor(months / 12)} y ago`;
}

// THE disabled-until law: ≥1 checked item AND a chip chosen
function refreshSyncMarkButton() {
  const markBtn = document.getElementById('sync-mark-btn');
  if (!markBtn) return;
  const ids = Array.from(syncDiff.checked);
  const reason = syncDiff.reason;
  markBtn.disabled = !(ids.length >= 1 &&
    ['not_ready', 'private', 'duplicate', 'other'].includes(reason));
}

function syncMarkIntentional() {
  const ids = Array.from(syncDiff.checked);
  const noteEl = document.getElementById('sync-note-input');
  const note = noteEl ? noteEl.value.trim() : '';
  const names = {};
  const stage = syncDiff.stage;
  for (const it of (stage && stage.items) || []) names[it.id] = it.name || it.id;
  MarkIntentional(syncDiff.key, ids, syncDiff.reason, note, names).then((r) => {
    if (!r || !r.marked) return;
    const confirmLine = document.getElementById('sync-confirm-line');
    if (confirmLine) confirmLine.textContent = r.line;
    if (noteEl) noteEl.value = '';
    syncDiff.checked = new Set();
    syncDiff.reason = '';
    for (const c of document.querySelectorAll('.sync-chip')) {
      c.classList.remove('sync-chip-on');
    }
    // the modal re-opens its own fresh data (the rows left the list;
    // the log tab grew) — pure local, instant
    StageDifferences(syncDiff.key).then((payload) => {
      if (payload && payload.found) renderSyncDiffModal(payload);
    });
    refreshSyncView();  // the cell + banner repaint behind the modal
  });
}

// the Deferred log tab: search + the 6-column table + the note
function buildSyncLogTab(payload) {
  const page = el('div', 'sync-tab-page');
  page.id = 'sync-log-page';
  page.hidden = true;

  const searchRow = el('div', 'card-row');
  const search = el('input', 'form-input');
  search.id = 'sync-log-search';
  search.placeholder = 'search the log — item, reason or note';
  search.addEventListener('input', () => renderSyncLogRows());
  searchRow.appendChild(search);
  page.appendChild(searchRow);

  const table = el('div', 'sync-log-table');
  table.id = 'sync-log-table';
  page.appendChild(table);
  page.appendChild(el('p', 'muted-note wordwrap',
    'Nothing is ever deleted from the log — status changes only. ' +
    'Resolving happens automatically when the item later appears on ' +
    'the website; Reopen puts an item back in the Differences tab.'));
  renderSyncLogRows();
  return page;
}

const SYNC_LOG_CAP = 500;  // the log tab's paint cap

function renderSyncLogRows() {
  const table = document.getElementById('sync-log-table');
  if (!table) return;
  const search = document.getElementById('sync-log-search');
  const q = search ? search.value : '';
  DeferredRows(syncDiff.key, q).then((rows) => {
    table.textContent = '';
    const head = el('div', 'sync-log-row sync-log-head');
    for (const h of ['Item', 'Reason', 'Note', 'Date reviewed', 'Status', '']) {
      head.appendChild(el('span', '', h));
    }
    table.appendChild(head);
    const labels = { not_ready: 'Not ready', private: 'Private', duplicate: 'Duplicate', other: 'Other' };
    const statusShown = { active: 'Active', resolved: 'Resolved', reopened: 'Reopened' };
    for (const rec of (rows || []).slice(0, SYNC_LOG_CAP)) {
      const row = el('div', 'sync-log-row');
      row.append(
        el('span', 'sync-log-item', rec.displayName || rec.display_name || rec.item_id || rec.itemId || rec.item_id || ''),
        el('span', '', labels[rec.reason] || rec.reason || ''),
        el('span', '', rec.note || ''),
        el('span', '', (rec.dateReviewed || rec.date_reviewed || '').substring(0, 19).replace('T', ' ')),
        el('span', '', statusShown[rec.status] || rec.status || ''),
      );
      if (rec.status === 'active') {
        const reopen = btn('Reopen', 'secondary',
          'Put this item back in the Differences tab — the deferral is ' +
          'withdrawn (the row stays in the log, reopen counted).',
          () => {
            ReopenItem(syncDiff.key, rec.itemId || rec.item_id).then((r) => {
              if (!r || !r.ok) return;
              const confirmLine = document.getElementById('sync-confirm-line');
              if (confirmLine) confirmLine.textContent = r.line;
              renderSyncLogRows();
              refreshSyncView();
            });
          });
        reopen.classList.add('sync-reopen-btn');
        row.appendChild(reopen);
      } else {
        const dash = el('span', 'sync-log-dash', '—');
        dash.title = 'only Active rows reopen — a resolved row already ' +
          'left the gap; a reopened row is already back in Differences';
        row.appendChild(dash);
      }
      table.appendChild(row);
    }
  });
}

function closeSyncDiff() {
  const root = document.getElementById('sync-diff-modal');
  if (root) root.hidden = true;
  syncDiff.open = false;
}
