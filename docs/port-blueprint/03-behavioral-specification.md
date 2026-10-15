# 03 — Behavioral Specification

> What the application *does*, expressed as observable behavior. Evidence tags `app/…:line` at
> v0.91.0. Deep per-module contracts live in `module-specs/`; this document is the workflow-level
> layer the Go port must reproduce. Terminology follows the repo's own laws (AGENTS.md, SPEC.md).

## 1. The core value loop (owner-visible)

1. **Send** — forward any link to `@githubfetcherbot`. GitHub links get an instant reply with stars
   + description; other websites get "🌐 Received"; banned domains (THE LAW: x/Twitter, GitHub
   group, HuggingFace, Instagram, Facebook, LinkedIn + config extras) get an honest 🚫 and are
   never fetched (`core/links.py:298-335`; Worker enforces the same law).
2. **SYNC** (GUI button or `--sync` CLI door) — fetch undone bot items, classify, and report.
3. **PROCESS** — the batch: each link is fetched, LLM-curated, and written as an Obsidian note.
4. **Seal & publish** — vaults commit/push to private GitHub mirrors; the Websites vault's
   public-ready notes publish to the live directory site.
5. **Learn from the owner** — moving a note to a different folder is recorded as a *correction*
   (category updates + lock + few-shot example); the app never moves it back.

## 2. User workflows (must-preserve list)

### 2.1 W1 — First run / setup
- `1-INSTALL.bat` creates `.venv` + installs requirements; `GitCurator.bat` launches the GUI;
  `GitCurator-DRY-RUN.bat` rehearses with **zero writes**.
- First-run wizard (`about_me.md` modal) when missing (`gui/main_window/lifecycle.py:58-171`).
- Settings pages (Credentials, Proxy, Vault, LLM, Input, Dashboard, Sync, Bot, Sources, Backup,
  Sound) write `config.json` with merge semantics; credentials also mirror to
  `Documents/.gitcurator` so they survive version updates (`core/user_config.py`).
- CLI alternative: `--init` wizard (with repo-radar pre-run) and `--login`.
- **Test Connection** modal: four subsystem rows (Vaults/LLM/GitHub/Telegram) spin → settle on
  ✅/⚠️/❌ with check lines; **Start Syncing stays disabled until all four are ok**
  (`gui/dialogs.py:250-406`).

### 2.2 W2 — The SYNC door (GUI)
- One hero button, four states: SYNC → FETCHING → PROCESS(n) → STOP
  (`gui/main_window/hero.py:423-489`).
- Offline guard: if the sentinel says offline, an advisory "Continue anyway?" modal before the
  fetch burns its timeout (`hero.py:140-155`).
- Queue report (Settings → Bot / log desk): counts for total/in-vault/decommissioned/settled/
  blocked/self/pending, per pipeline, plus numbered pending lists (`gui/worker_jobs.py:93-391`).
- Batches >10 items ask for confirmation (`hero.py:47-55`).
- Caught-up sync: honest "Nothing new from Telegram", master-table waiting rows reported (never
  auto-refetched), 🖐 hand rows scraped via Chrome, redo pass, then the **drift seal** (seals dirty
  vaults + publishes; quiet when clean — `hero.py:345-373`).
- **Fast Sync** (v0.90.0): the triad-only mirror (3 seals + directory publish; no Telegram, no LLM,
  git+HTTP only); always-publish + always-speak laws (`gui/main_window/backup_seal.py:89-198`).

### 2.3 W3 — Import from file
- `.txt`/`.md`, one address per line; markdown links and bare domains recognized; duplicates
  counted; unparsable lines listed (never silently dropped) — the faithful grammar is pinned in
  `spec-core-links.md` (`core/links.py:678-798`).

### 2.4 W4 — The GitHub pipeline (per repo)
Order and gates (`gui/processing_worker.py:947-1719`): rate-limit guard (pause under 50 remaining;
  403-rate → 3,605 s wait) → canonicalize → dead-set → dismissed → settled → invalid →
  **vault-index dedupe BEFORE any API call** → `get_repo` (404 → strike + `_missing/` placeholder
  note; 401 → token dropped → anonymous retry) → metadata/topics/org-rep/90-day commits → README
  (≤1,500 chars) → banner (async thread, throttles, 24 h failure cache) → LLM analyze
  (3-attempt ladder with `<<<README_BEGIN>>>` injection guard) → category validation
  (confidence <50 → Uncategorized) → note build (sanitized frontmatter; `spec-core-note-builder.md`)
  → atomic write (disk-full protocol) → cache + note_state + manifest updates → 0.5 s delay.
- 404 quarantine: N strikes (default 3, config `notfound_strike_threshold`) → confirmed dead,
  skipped everywhere; delete the placeholder note to re-check.

### 2.5 W5 — The Websites pipeline (per link)
Full state machine in `spec-core-website-pipeline.md`; the owner-visible guarantees:
- Gates: banished-list (60 s cache) → blocked domain → self domain → dismissed → already-in-vault
  (with failed-`_review` upgrade and 🖐 exemptions) → **one note per site** consolidation →
  retry gate (max 3).
- Fetch ladder: hand-delivered page → stdlib-through-proxy → direct fallback → curl_cffi
  impersonation (only on 403/405/429/451 or TLS markers) → (async) hand-delivery queue → Chrome
  tab retry. Dead/paywalled/refused are dismissed with an auto-verdict; healable failures requeue
  with `2×attempts` day backoff (`core/website_state.py:152-172`).
- Classification: two passes (w01 category, w02 subcategory) against the taxonomy file with
  exact-name validation, 1+2 corrective retries, low-confidence → `_review/`.
- Analysis (w03) → note with the web-v2 body (`## What it does`, `## Best used for` always
  present, practical details, watch out, similar tools, source footer, delete-hint block).
- Banishment: 🗑️ verdict (frontmatter keys, tags list, body inline tags, master-table gesture) →
  **confirmation gate over Telegram** (Delete all / Keep all; timeout = safe defer) → move to
  `.trash/banished`, dismiss URL forever, twins swept in the other vault, run tally always spoken
  (`core/website_pipeline.py:2243-3052`).
- After every batch: the **Website Directory note** (`000 📚 Website Directory.md`) regenerates;
  the master table `_review/DECOMMISSIONED.md` refreshes (waiting/stored/hand rows, pruning).

### 2.6 W6 — Moves are corrections (run start, both machine vaults)
Baseline on first run → classify (moved/edited/deleted/duplicates/unmanaged/unmapped) →
apply: moved notes get frontmatter updated, `category_locked: true`, correction logged, never
moved back; deleted → dismissed forever; edited/dupes/unmapped → report only
(`core/note_state.py:477-842`).

### 2.7 W7 — The vault Scan (GUI 🔍 button or Telegram door)
- Vault picker modal (both machine vaults tickable; Personal shown read-only, never scanned)
  (`gui/vault_pick_dialog.py`).
- Scan engine: inventory walk (one-open cache) → stray/hygiene pass → candidate radar
  (declared-home → sibling-domain → taxonomy-tag regex passes) → weighed evidence shortlist →
  chunked LLM filing (60/chunk, ≤300 shown) → validated plan (deletions, consolidations, new
  folders, moves with `_v` name-bending and landable names) (`core/vault_scan.py`).
- **Nothing moves until the owner confirms**: GUI plan modal (checkbox rows, masters, live Apply
  count, 350 ms fresh-click consent guard, 300 s window) or Telegram round-trip
  (`gui/scan_plan_dialog.py`, `integrations/scan_confirm.py`).
- Apply: verified landings, census stamp `curated: true`, failure ledger with named files,
  scorecard modal with **landed counts** (never the plan's promises), self-closing 12 s
  (`gui/scan_result_dialog.py`).

### 2.8 W8 — Seals, mirrors, publish
- VaultSeal per vault (creation guard: never auto-creates; explicit opt-in creates **private**);
  divergence ladder: push → fetch → rebase → merge-unrelated → rescue branch; main never
  force-pushed; token never in `.git/config`, FETCH_HEAD sanitized (`integrations/vaultseal.py`).
- **Local-only belt**: private category folders + configured stems are gitignored + untracked —
  files never leave the disk, mirrors lose them on the next seal (`cloud/directory_sync.py:478-567`).
- GoodRepos public directory publish (staging tree + README; same creation guard).
- Website Directory publish to the Worker: `build_index` (tree = vault folder tree, private belt,
  tombstones) → `POST /api/directory/sync` (Bearer) → read-back door `GET /api/directory/index`
  (honest 404 = never published).
- Backup tab: local zip backups (FIFO rotation, excludes), resync (mirror follows vault exactly,
  deletions included, history kept, never forced), repo radar (read-only detection, never
  overwrites owner-typed names).

### 2.9 W9 — The Sync view (Settings → Sync)
- Four-element flow cards (🤖 bot → 📓 vault → ☁️ GitHub / 🌐 site) with status dots
  (green counted / grey not-configured / red error spoken), hero numbers with micro-copy, stat
  rows, last-sync times, where-to-go hints; Refresh is **on-demand consent** for all network legs
  plus ONE automatic refresh on first open (`gui/main_window/sync_flow.py`, `gui/sync_display.py`).
- The §5 arithmetic: `V>P → "V−P need publishing"`; `V<P → "−N since the last publish"` (never a
  negative needs-publishing); never-published = "never".
- 3×3 gap matrix (Repos/Websites/Personal × Queue→Vault / Vault→Website / Vault→Backup) in four
  states — Match / Behind / Reviewed / Unknown (split "not set up" vs "not measured", v0.91.0
  named verdicts) — plus the differences modal with the four reason chips and the never-delete
  Deferred log (`core/sync_review.py`, `gui/main_window/sync_view.py`).

### 2.10 W10 — The five fetch doors (websites)
Door 0 hand-delivered page → Door 1 stdlib via proxy → Door 2 direct fallback → Door 3 curl_cffi
impersonation → Door 4 hand-delivery queue (owner saves pages with Ctrl+S into
`_review/hand-delivered/`, README teaches it) → Door 5 Chrome tab retry (attach to the owner's
running Chrome on 9222 → real-profile junction → dedicated persistent profile; wave engine of 8
tabs; `page_is_real` grammar rejects error/crash/challenge pages; delivered pages land in the
hand folder stamped `door:auto` and re-verified at consume) — `core/hand_delivery.py`,
`core/chrome_tabs.py`.

### 2.11 W11 — LLM backends
- Three providers behind one router: Ollama (auto-detect, auto-start, model stand-in
  recommendations), llama.cpp (process-port detection ~1.5 s after launch, `/props` config grab →
  fills `llm_num_ctx`/`llm_max_output_tokens`), cloud (OpenAI-compatible or Anthropic by URL;
  `/v1/models` pre-flight warns, never blocks; JSON mode with one memoized clean retry).
- Failure UX: the LLM-failure modal (model dropdown, Remember, Skip/RetrySame/RetryWith/Stop,
  600 s window); per-task model overrides (`models.classify/analyze/vaultscan`).
- The owner's past corrections ride the classifier prompts as few-shot examples.

### 2.12 W12 — CLI doors
`--init`, `--login`, `--status`, `--test-connection`, `--detect-llm`, `--list-dead/--reset-dead`,
`--hand-delivery`, `--chrome-retry`, `--auto` (preflight → fetch → classify → guard → confirm →
batch → seals → CLEAR), `--sync` (check → Enter → sync; `--yes` = cron shape), `--retry-failed`,
import/range/offset/single-id modes, `--dry-run` everywhere. Full flag table:
`spec-cli.md`.

## 3. Honesty laws (cross-cutting, user-observable)

1. **Never a zero that lies** — unconfigured/unmeasured shows "—" with its reason, never 0.
2. **Icon + text + color together** (WCAG) — every dot has a word.
3. **Unconfigured = silence** (no backup-repo row, no error to fix).
4. **Settled = addressed, never re-fetched** — 🤝 lines at every door.
5. **Pending = zinc ⏳, not red** (red reserved for failures).
6. **K-of-N shortfalls speak** (landed counts, never promises).
7. **Quiet-when-clean automatics; always-speak manual doors** (drift seal vs Fast Sync).
8. **Every dot/label carries the full story on hover** (tooltips as second layer).
9. **Countdown honesty**: "at zero nothing is done" (300 s scan window); scorecard self-closes 12 s.
10. **Never-raise integration boundaries**; one failing item never stops the batch.

## 4. Failure behavior (observable defaults)

- Telethon subprocess: idle 180 s / hard cap 30 min / auth grace 360 s; kill reasons surfaced
  verbatim; watchdog force-releases the lock at 35 min.
- LLM timeout: `TimeoutError "LLM call did not respond within {n}s"` → the failure modal (GUI) or
  console menu (CLI); headless fails fast.
- Disk full: GUI modal + paused spin until resolved or stopped; headless skips the note.
- Worker down: network halves of the Sync view degrade to honest error states; local halves stay
  measured; restart recovers the numbers (pinned by `test_synclive` outage cases).
- Git failures: reported verbatim with the reconciliation story; curation never fails because a
  backup failed.

## 5. Sensitive semantics (accidental-change hotspots)

Pinned in `05-python-to-go-mapping.md` §3 with per-item tests; headline list:

- URL canonicalization rules (both normalizers) — byte-exact identity keys.
- Dict **insertion order** as observable behavior (dedupe winners, platform first-match,
  approved-links ordering).
- `str.lower()` vs Go casing on edge codepoints; folded-identity grammars.
- Code-point slicing caps (1200/600/300/220/200/80 chars).
- `round(x, 2)` half-to-even vs Go rounding (vault-scan weights).
- Naive-local ISO timestamps (most tables) vs UTC-aware (sync_deferrals); `%Y-%m-%d` note dates.
- `os.walk` unsorted vs Go lexical walk (duplicate winners, scan order).
- `rstrip` character-set semantics; `…` U+2026 token scrubbing; `🗑️` with VS16 substring checks.
- YAML scalars: always-double-quoted, `\` then `"` escaping, caps.
- JSON: `ensure_ascii=False` persistence; Go's default HTML escaping must be disabled.
- SQLite: same DDL, same naive-string dates, WAL, busy timeouts.

## 6. Behavior verified by tests vs inferred

- **Test-pinned** (highest confidence): URL grammar, note formats, sanitizers, moves-as-corrections,
  retry/backoff, banishment grammar + gates, scan plan validation/apply, sync-review state model,
  Worker HMAC protocol (loopback rig), log coalescing, fast-sync laws, drift-seal laws,
  local-only belt, creation guards, one-spelling, census stamp, hygiene fingerprints.
- **Inferred from implementation** (documented, lower tier): exact Windows-only behaviors we could
  not execute here (junction creation, taskkill paths, Known-Folder resolution), live-LLM
  interaction feel, Chrome attach behavior against a real browser, timing envelopes.
- **Unverifiable in this session**: GUI pixel rendering (PyQt6 unavailable in the sandbox — the
  pure-stdlib test modules pass: `tests.test_core` 34/34, `tests.test_notereads` 16/16,
  `tests.test_webnote_v2` 10/10; GUI modules need PyQt6, not installed here).
