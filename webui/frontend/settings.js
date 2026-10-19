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
  SaveBotGroup, CheckBotQueue, ProcessBotQueue, ProcessSourcesUrls,
  RefreshDashboard, ClearQuarantine, SaveQuarantineGroup,
  ReadUndoLastBatch, ApplyUndoLastBatch, FetchFromSources,
  SaveBackupGroup, RefreshBackupLadders, BackupNowGuards, BackupNow,
  SealNow, ResyncMirror, PublishNow, DetectRepos, FastSync,
  OpenDashboard,
} from './wailsjs/go/main/PagesBUI.js';
import { EventsOn } from './wailsjs/runtime/runtime.js';
import { svgEl } from './icons.js';

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
    stackEl.appendChild(page);
  }
  selectSection('credentials');
}

function selectSection(id) {
  for (const sec of SECTIONS) {
    if (sec.id === id) hintEl.textContent = sec.hint;
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
  return p;
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
    'Pick a backup ZIP and extract it to a NEW folder next to the current vault — the current vault is never touched\n(lands with the packaging phase — the safety-net restore engine)', null);
  restore.disabled = true;
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
