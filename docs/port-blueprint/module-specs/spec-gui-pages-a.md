# Module Spec — Settings pages A: `hero.py`, `dashboard.py`, `bot_queue.py`, `telegram_ui.py`, `connection_tests.py`, `test_connection_modal.py`, `vaults_config.py`, `input_proxy.py`, `llamacpp.py`, `linking_tools.py`, `vault_scan_ui.py`

> **RE-SCOPED (2026-10-15, ADR-0004 Am.3 — owner stamp "OPTION B CONFIRMED", rider R-0):**
> the Go app drops its direct Telegram (MTProto) leg; the bot queue rides the Cloudflare Worker
> HTTP API only. Ripples in this spec, recorded as dated notes at their sections below:
> (a) `bot_queue.py`'s data source becomes the Worker's `/api/pending` + `get_pending_result`
> path (the Telethon fetch, its lock, and `_mark_bot_messages_read` fall away);
> (b) the battery modal's live Telegram leg is dropped; (c) `telegram_ui.py`'s lock/login UX
> falls away entirely (no MTProto client ships). The queue still *originates* on Telegram
> (the bot webhook → Worker path), so Telegram-worded observable strings (e.g. hero's honest
> empty state) stay accurate and are ported verbatim.
>
> **Pointer note (same date, R-0):** the R-0 instruction labeled this edit "spec-gui-pages-b
> (bot-queue data source = Worker pending)" — that label follows the task-list page grouping
> (doc 13, T-74 "pages B: Bot, Dashboard, …"). The bot-queue page spec actually lives **here**
> (spec-gui-pages-a.md, `bot_queue.py`); `spec-gui-pages-b.md` covers only `backup_seal.py`
> and has no Option-B ripple of its own. A clarifying note on the T-73/T-74 grouping lands in
> doc 13 in the same commit.

- **LOC**: 489+838+1,586+217+568+267+539+154+687+124+332 · **Go home**: `ui/frontend/src/pages/*`
  + command bindings · **Status**: rehost (contracts below; Telegram legs re-scoped per the
  2026-10-15 banner above)
- **Responsibility**: the hero SYNC state machine; Dashboard page; the bot-queue view; Telegram
  lock/code UX; per-subsystem test buttons; the connection battery modal; config load/save;
  proxy page; llama.cpp page; Phase-6 tools; the Scan CTA wiring.

## hero.py — the two-stage SYNC flow

- State machine `sync → fetching → process → running` on ONE button; label/glyph/kind per state;
  progress formats ("Fetching undone items…", "{n} ready to process", …); running = Stop
  (hero_danger, stop glyph).
- Offline guard: advisory "Continue anyway?" BEFORE the fetch (140-155). No bot + no import ⇒
  "Nothing to sync" teaching modal.
- `_after_sync_fetch` decision tree: pending ⇒ PROCESS; caught-up + import ⇒ PROCESS; caught-up ⇒
  waiting rows report + 🖐 scrape via Chrome + redo pass + **drift seal** + honest empty state
  ("Nothing new from Telegram · Last sync …").
- `_confirm_batch(>10)`.

## dashboard.py

- `update_dashboard` (61-155): vault stats from first-1000-chars frontmatter reads (category,
  credibility, language counts; `_review` count; banner count).
  — **Dated note (2026-10-20, T-74/SESSION-037):** the page's engines landed
  (`internal/app/settingsb.go`: ScanDashboardStats + RenderDashboardText +
  QuarantineView/ClearQuarantine + the threshold group's merge save + the undo
  engine with the top-10 confirm + the processed_repos row removal + the once
  law); verify_vault/recategorize_notes/dead-links viewer/retry launchers stay
  T-76/T-77's (the More-menu family — the page's own rows are live).
  — **Dated note (2026-10-20, T-79c/SESSION-044, the completion half):** the
    failed-repos retry launcher landed at T-79c — `dashboard.py:770`'s flow
    over `internal/app/botdoors.go`'s scan + the More menu's two-phase shell
    door (the vault-error box, the settled-skip layer, the >10 confirm, the
    pending-lists clearing, the DIRECT batch). The reconciliation-links
    retry (the banner's own action) remains the Sync view's door; the
    record: `docs/poc-artifacts/t-79c-bot-queue-doors.md`.
- Quarantine view/reset (`refresh_quarantine_view` top 30 rows + "… and N more";
  `clear_all_quarantine`; threshold spinbox saves `notfound_strike_threshold`).
- `undo_last_batch` (228-311): reads `<vault>/_undo_last_batch.txt`, deletes files best-effort,
  deletes processed rows by note_path.
- `verify_vault` (313-498): enqueues only GitHub-type failed links (v0.63.1), skips settled
  (v0.63.3); `recategorize_notes` modal (QTableWidget + per-row category combo; locked notes
  skipped); dead-links viewer + reset; retry launchers (reconciliation links, failed repos) with
  the >10 confirm gate and bot-queue list clearing.

## bot_queue.py

- `check_bot_queue` (141-635): acquire lock → save config → TestWorker(_bot_queue_job) with a
  re-bound closure injecting law-split domain lists → on finish: write `_inbox` platform tables;
  read worker-side classification (GUI fallback re-runs it if fields missing) → pending badge
  `⏳ N pending` (zinc) / `✅ 0 pending` (mint) → on_done
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** the **bot-queue data source in the Go
  port is the Worker pending queue** (`/api/pending` + the existing `get_pending_result`
  path — T-34 scope): no Telethon fetch, no acquire-lock, no `TestWorker(_bot_queue_job)`
  fetch leg. The classification buckets (blocked/self/in_vault/decommissioned/settled/pending
  + websites twin), the report line names, the badges, and the `_inbox` platform tables stay
  — they are computed from the Worker pending list + vault state, and are ported verbatim.
- Report "columns" = labeled count lines (Total GitHub URLs in bot / ✅ Already in vault /
  🗑️ Decommissioned / 🤝 Repos settled / 🚫 Blocked / 🔒 Self / ⏳ Pending / 🔗 Non-GitHub +
  websites sub-table / 🔄 Duplicates removed) + numbered pending lists (websites capped 50).
  — **Dated note (2026-10-20, T-74/SESSION-037 — the port's Check Queue door):** the
  page's engine is `internal/app/settingsb.go`'s RenderQueueReport over a classified
  QueueReport (the canned-result law, pinned by tests) + the badge's two-state law;
  the page keeps the username guard (its refusal line is observable) and drops only
  the Telethon credential + proxy-warning guards; **Process New and Mark All as Read
  do not render** (the cursor + read-acknowledge are Option B's accepted losses — the
  ledger + the mirror push are the truth); the Worker-pending fetch itself rides
  Commands.SyncFetchQueue, the render + the badge ride the page engine, and the
  port's Check Queue saves the config before the fetch exactly like the reference.
- `process_bot_queue` (settled-link exclusion layers + confirm) / `process_new_bot_queue`
  (incremental by `last_processed_msg_id`); `export_all_bot_links`; `verify_all_bot_links`
  (+ fuzzy match + manual-resolve dialog + permanent `_verification_report_*.md`);
  `clear_bot_queue` (via `QTimer.singleShot(0, …)` to break the signal-ordering self-deadlock);
  `_mark_bot_messages_read`
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** `_mark_bot_messages_read` is dropped in
  the Go port (read-acknowledge is an accepted loss; mirror-push is the ack —
  spec-worker-contract §8.4). `process_new_bot_queue`'s incremental `last_processed_msg_id`
  cursor falls away with the Telethon fetch; the Worker queue is ledger-state-based. The
  settled-link exclusion layers and confirm gates stay.
  — **Dated note (2026-10-20, T-79c/SESSION-044 — the three bot-queue tool doors
    LIVE):** the More menu's third family landed (`internal/app/botdoors.go`
    + the `HeroUI` wrappers + the heroflow.js wiring + the
    manual-resolve dialog): ✅ Verify All Processed (the reference's
    own per-URL ladder — in-vault → decommissioned → cache-processed
    → fuzzy → settled → missing — over the Worker pending poll +
    the byte-parity VERIFICATION REPORT + the permanent
    `_verification_report_*.md` + the resolve payload) + 📋 Export
    All Links (the routed lists + the byte-parity
    `_bot_links_export_*.md` + the 📋 Links Exported box) + 🔄 Retry
    Failed (the two-phase scan: the vault-error box, the read-failure
    line, the empty verdict, the settled-skip layer, the >10 confirm
    over "the retry queue"; the post-confirm leg clears the session's
    pending lists — the reference's `_bot_queue_urls` law — and runs
    the DIRECT batch) + the manual-resolve engines (Mark as
    Processed / Decommission / Mark ALL settle the ledger too —
    v0.63.3's law; Process Selected runs the bot-source batch). ONE
    new hermetic `BotQueueFetch` seam (the SAME a.queueFetch
    composition the queue command rides). THE GUARD LAW carried
    exactly: NO busy guard on export/verify (the reference's
    TELEGRAM lock was their only guard — Option B's accepted loss,
    the doors run alongside a batch, the reference's own concurrency
    semantics); the retry batch rides SyncRunBatch's one-at-a-time
    refusal. The recorded divergences (the export's message count =
    Σ forward_count; the save-before-verify rides the save-on-change
    law; the mark-all-read question dropped; the report rides the
    More door's themed box; the resolve dialog opens through the
    report box's own button; the one-connection cache blocks; the
    sorted fuzzy iteration; the plain-open file writes; repo_id 0
    replaces): `docs/poc-artifacts/t-79c-bot-queue-doors.md`.

## telegram_ui.py

- Lock lifecycle: `_acquire_telegram_lock(owner)` (busy message names holder + age; >300 s stuck
  hint); `_keep_worker(worker, owner)` (GC guard; `_cleanup` connected LAST; owner-scoped
  release); watchdog 35 min force-release; login/2FA modal (Enter submits; Cancel feeds `""`)
  — **DROPPED in the Go port (2026-10-15, ADR-0004 Am.3 / R-0):** no MTProto client ships, so
  there is no lock to acquire, no code/2FA modal, and no worker to keep. The Telegram config
  keys keep being *read* for compatibility (the compat law); the engine never uses them. See
  spec-gui-intake.md (telegram_lock.py) for the same re-scope.

## connection_tests.py + test_connection_modal.py

- Per-button tests: Telegram+GitHub, GitHub token (401 → actionable message; empty → anonymous
  warning 60 vs 5000 req/h), proxy TCP, vault, Ollama (refresh models / **detached server spawn**
  + 3000 ms re-probe / inference test "What is 2 + 3?" with options retry), lenient
  `llm_num_ctx`/`llm_max_output_tokens` readers, cloud test (`/v1/models` pre-flight warn-only;
  flavor by URL; empty-200 vs connection-failed modals).
- Battery modal: snapshot saved config + live credential widgets; rows spin → settle; **crash ⇒
  never "ALL SYSTEMS READY" from an empty result**; live telegram leg (bot queue when configured
  else Saved-Messages preview); Start Syncing gated on all-ok
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** the live Telegram leg is dropped from the
  Go battery modal; the remaining legs (vault, LLM, GitHub, Worker version/reachability —
  spec-core-connection) gate "Start Syncing". The crash-honesty law (never "ALL SYSTEMS READY"
  from an empty result) stays.
  — **Dated note (2026-10-20, T-79a/SESSION-042 — the five per-test doors LIVE):** the
  More menu's per-test doors landed (`internal/app/pertests.go` + the `HeroUI` wrappers +
    the heroflow.js wiring): Test Telegram & GitHub (the Option-B re-scope — the GitHub
    half verbatim + the Worker queue leg in the Telegram half's seat), Test Proxy
    Connection (the re-scope — the sentinel's ProxyReachable TCP verdict; the Telegram
    round trip is the accepted loss), Test Ollama (models + the real "What is 2 + 3?"
    ping), Test Cloud API (the /models pre-flight + the "Say hello" ping + the three
    verdict boxes), Test llama.cpp (the /props probe + the pick + the "Say hello" ping +
    the three verdict boxes). Five new hermetic `SettingsDeps` seams
    (OllamaChatPing/CloudChatPing/PreflightCloud/WorkerCheck/ProxyCheck); the guard
    difference recorded (the reference's telegram-lock doors guarded the TELEGRAM worker,
    dropped under Option B — these doors are read-only probes, no busy guard, exactly like
    the reference's own Ollama/cloud/llama.cpp doors); the divergences recorded in
    `docs/poc-artifacts/t-79a-per-test-doors.md` (the options-retry vocabulary — the
    port's num_ctx/num_predict wire shape carries no temperature; the raw-response-type
    line skipped; the model-seat write holds the Settings mutex; the port's own
    Worker-not-configured info line; the two re-scoped doors' opening lines speak the
    port's truth). The evidence: 27 engine tests + 3 shell tests; the VLM-verified
    menu + click-leg screenshots; the full record at the artifact above.

## vaults_config.py

- `load_config` (209-258): CONFIG_EXAMPLE deep-copy base + user-config fill + credential strip
  for display.
- `save_config` (260-485): **four laws** — MERGE never whitelist-rebuild; tuning keys only
  defaulted; **`self.config` mutated IN PLACE** (clear+update) so running workers see saves;
  atomic write. Plus: user-config mirror; sentinel re-arm `singleShot(250 ms)`; vault combo
  history; pickers' status ladder (`● not set` / `● found` / `● will be created` /
  `● missing on disk` / `● not found — you create this vault yourself`); no pre-selection law;
  volume slider label-only during drag.
- Sound page handlers.
- **Dated note (2026-10-20, T-77/SESSION-040 — the chime assets SHIP):** the deferred-debt
  note ("the preview's missing-asset notice is the parity behavior until T-77/T-80 ship
  the files") closes: success.wav + retry.wav + the README landed in `assets/sounds/`
  (sha-identical copies of the reference's synthesized, license-free files), embedded in
  the assets package. The lookup is DISK-FIRST (an owner's APP_DIR/assets/sounds copy
  wins — the swap-the-chime law) with the embed as the fallback; the preview's
  missing-asset 🔕 notice now fires only when BOTH are absent. The batch-finish play
  rides the same gate (the quiet `soundFor` — `_play_batch_sound` speaks nothing on a
  switched-off chime; the preview keeps its own 🔇 line), crossed with the
  batch-finished event for the frontend's `<audio>`.
- **Dated note (2026-10-19, T-73/SESSION-036 — the port's commit cadence):** the Go port's
  six Settings pages commit per page on `editingFinished`/`toggled` (the established
  save-on-change pattern applied uniformly — the task brief's own wording), where the
  reference saves the Credentials/Proxy/LLM plain fields only at window close
  (`closeEvent` → `save_config`). Every reference-persisted value persists in the port;
  some persist a commit earlier (a crashed window can no longer lose an unsaved credential
  edit). The four save laws, the ladder strings, and the mirror are byte-parity, pinned by
  the save-law tests. The engine-action saves (Detect & Set, the props grab, autodetect)
  keep the reference's own best-effort save cadence verbatim.

## input_proxy.py / llamacpp.py / linking_tools.py / vault_scan_ui.py

- input_proxy: import picker (.txt/.md); `_get_proxy_dict`; proxy status label (words, WCAG
  label==accessible-name); the 2 s TCP probe lives on the sentinel thread (v0.85).
- llamacpp: startup autodetect daemon thread → queued signal; **/props config grab** fills the
  limits fields ("🦙 llama.cpp server config grabbed: context window N tokens · output cap M
  tokens"); Detect/Refresh/Test; quick-detect dialog (combo min-contents-length law); apply
  law (in-place + merge save; cancel ⇒ "the LLM provider was NOT changed (still {provider})");
  loading model ⇒ "⏳ model still LOADING…".
- linking_tools: Phase-6 tool runners inside TestWorker with stdout capture, last 40 lines
  streamed; guard: no tool while a batch runs.
  — **Dated note (2026-10-20, T-79b/SESSION-043 — the five linking doors LIVE):** the
    More menu's linking-tools family landed (`internal/app/linkdoors.go` + the `HeroUI`
    wrappers + the heroflow.js wiring + the recategorize table modal): 🪝 Recall hooks
    (dry-run) + 🔗 Build link suggestions over the TestWorker law (the runner: the guard
    → the "⏳ {label} running… (the GUI stays usable)" opening → the captured
    stdout+stderr run → the last-40-lines streaming at completion → the ✅/❌ finish
    verdict), ✅ Validate Vault (connection_tests.py:219 verbatim), 🔍 Verify Vault
    (the report box + the enqueue legs), 📁 Recategorize Notes (the two-door flow —
    the open walk + the table modal + the apply). THE GUARD LAW widened per the task
    row's own naming: all five doors carry the batch-busy guard with the reference's
    own line (the reference guards only the two tool runners — the recorded
    divergence, docs/poc-artifacts/t-79b-linking-doors.md's list). The other recorded
    divergences: the provider abort speaks its message into the captured stderr and
    exits 2 (the reference's GUI worker crashed on the string SystemExit code and
    lost the message — a display bug not copied); the 💥 crash line renders the
    recovered panic's own text (the T-40 house law); Verify Vault's report renders in
    the More door's themed box (the reference's dashboard text area is a main-window
    page the port does not carry — the port's Dashboard is a Settings page) with the
    long-body scroll law; the recategorize cache connection opens once per apply; the
    vectors' iteration order pinned to the collect order + sorted stale keys
    (deterministic, the reference rides the SQLite row order). The seams:
    LinkOllamaChat/LinkCloudChat/LinkEmbedCall on SettingsDeps (faked at
    construction). The evidence: 32 engine tests + 4 shell tests + the VLM-verified
    menu + the recall-door click leg; the full record at
    `docs/poc-artifacts/t-79b-linking-doors.md`.
- vault_scan_ui: Scan CTA → picker (None = safe cancel, quiet line); worker wiring (confirm-gui
  door, progress emitter, selection door, vaults scope); finished routing: clean (+healed strays
  ⇒ scorecard), confirmed ⇒ scorecard (repos sub-story line), error ⇒ 💥; scorecard skipped
  when closing/hidden (logged instead).
  — **Dated note (2026-10-20, T-76/SESSION-039 — the Scan flow landed):**
    the mixin's whole flow is ONE binding door (webui/scanui.go's
    `StartScan`): the guards verbatim (the batch-busy check over
    `Commands.BatchBusy()` + the one-scan flag), the picker's safe
    cancel spoken as the quiet line by the door itself (the frontend
    picker answers through it — Cancel/Esc/✕ all ride `StartScan(nil)`),
    the start line with the chosen-vault names, the blocking job over
    the chosen scope (the desk events flow on their own transport
    while the JS promise waits — the TestWorker law's transport twin),
    and `_vault_scan_finished`'s routing verbatim (clean+healed ⇒ the
    scorecard; confirmed ⇒ the complete line + the repos sub-story
    line + the scorecard; error ⇒ 💥; declined/timeout already spoke
    in the job). The progress emitter rides the bus's JobProgress →
    `bus:job-progress` (landed T-72). The "scorecard skipped when
    closing/hidden" law is the WINDOW's law — it rides T-77's window
    lifecycle (the port's shell never blocks on the scorecard: it is
    an overlay with a 12 s self-close, never a gate).

## Go parity requirements

All labels, badge colors (zinc pending), report line names, ladder strings, modal texts, and
cadences (250 ms re-arm, 3000 ms re-probe) ported verbatim — they are observable behavior.

## Tests

Hero state machine; queue report rendering from a canned job result; lock owner-scope; save-law
(merge + in-place mutation observed by a "running worker" fake); vault status ladder; scan
safe-cancel; battery crash honesty.
