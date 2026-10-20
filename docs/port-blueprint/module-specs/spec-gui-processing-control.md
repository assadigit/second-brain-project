# Module Spec — `gui/main_window/processing_control.py`

- **LOC**: 2,746 · **Go home**: `internal/app/commands.go` + `ui/` bindings · **Status**: rehost
  (completion chain is the contract)
- **Responsibility**: PROCESS dispatch, worker lifecycle, progress/status rendering, the whole
  end-of-batch chain (seals, publish, Phase-5 CLEAR, fanfare), the LLM-failure and disk-full
  dialogs, and the websites retry drivers.

## Key flows

> **Dated note (2026-10-20, T-79f/SESSION-047 — the fanfare chain's
> three automatic triggers landed; Phase 6 CLOSED):**
> `_maybe_deliver_hand_queue` (475) + `_scan_master_redo` (524) +
> `_maybe_redo_hand_notes` (560) + `_maybe_offer_chrome_tab_retry`
> (1197) + hero.py's caught-up hand scrape (280-305) + the
> caught-up redo check (316-328) + the work-started holds (329-330)
> live in Go as `internal/app/fanfaredoors.go` (the Settings-method
> engines over the landed scans) + `webui/heroui.go`'s `onBatchFanfare`
> (the chain at the batch-finished crossing, in the reference's
> order, gated on Natural + not-Stopped — the reference's
> `if _summary and not _summary.get('stopped')`) + the caught-up
> block's scrape/redo/holds + `startRedoRerun` (the reference's
> worker-thread shape) + the chrome:offer crossing +
> `ChromeOfferAnswer` (the DECLINED line's door vs the manual door's
> CANCELLED one) + **THE FLAG-CLEAR LAW** in SyncRunBatch (the flag
> clears before the finished signal — the reference's
> processing_finished opening). The end-of-run offer's data rides
> BatchFinished's new WebsiteResults facts (the runner's summary
> results now POPULATED, the full Result-dict mirror + THE FIELD-TAG
> LAW's tags). The stash interplay (`_stashed_batch_summary` +
> `_delivery_rerun_pending`) stays unlanded — its consumer, the Batch
> Complete modal, is no T-79 row's (T-79d's divergence #2 record);
> the six recorded divergences:
> docs/poc-artifacts/t-79f-fanfare-triggers.md.

> **Dated note (2026-10-20, T-79e/SESSION-046 — the decommission
> picker landed):** `decommission_dead_links_now` (807-908) +
> `_pick_dead_links_dialog` (910-998) + `_scan_backlog_alive`
> (152-173) live in Go as `internal/app/decomdoors.go` (the
> Settings-method engines): DecommissionScan (the guards — NO busy
> guard, the reference's own shape — + the backlog-alive scan with
> the is_dismissed probe + the pure-file fallback + the sorted-dedup
> law + the empty box + the candidates pre-fill) +
> DecommissionBury (the picked URLs marked dead through the SAME
> Status-cell writer a hand-edit rides — the table stays the ONE
> ledger — then the burials enforced immediately through the same
> consume pass a batch runs: dismiss + retry-row drop + placeholder
> sweep + the confirmed stamp; the exact success line + box; the
> cancel line; the unopenable-DB deferral law) +
> OpenDecommissionTable (the reveal over the OpenFileLocation
> seam) + `webui/heroui.go`'s three More doors + the frontend's
> picker dialog (the reference's own words; every row UNCHECKED — a
> burial is never a default; the danger bury button outside the
> scrolling body; THE BINDING-IMPORT pin) + the probe's 'grave'
> click pick. The six recorded divergences:
> docs/poc-artifacts/t-79e-decommission-doors.md. The unlanded
> companions named above (the end-of-run offer + the caught-up hand
> scrape + the redo pass) are now **T-79f — the Phase-6 closure
> row** on doc 13 (SESSION-046's scoping decision).

> **Dated note (2026-10-20, T-79d/SESSION-045 — the four Chrome doors
> landed as More-menu doors):** `hand_deliver_walled_links_now` (1000)
> + `hand_rows_deliver_now` (615) + `chrome_tab_retry_now` (1601) +
> `attach_my_chrome_now` (669) + the shared `_start_chrome_tab_retry`
> (1351-1519) + `_on_delivery_failed` (1521-1566) live in Go as
> `internal/app/chromedoors.go` (the Settings-method engines over six
> new hermetic SettingsDeps seams — ChromeFind / ChromeUserDataDir /
> ChromeAttachBat / ChromeAttachRun / THE ChromeDeliver CDP/DevTools
> seam / HandEnqueue; the busy ladders carried in the reference's own
> order, guards → busy → scan, the flags passed in as parameters) +
> `webui/heroui.go`'s seven More doors + the `startChromeTabRetry`
> runner (ONE delivery at a time with the backlog merge; the failure
> story FIRST over the chrome:failed event; the delivered pages
> re-run as the DIRECT NonGithub batch — the v0.57 routing law) +
> the frontend's two picker dialogs + THE BINDING-IMPORT pin. The
> unlanded companions — `_maybe_offer_chrome_tab_retry` (the
> end-of-run offer), hero.py's caught-up hand scrape, and
> `_maybe_redo_hand_notes` — remain the honest boundary: they ride
> the same runner, so their landing is a follow-up wiring. The six
> recorded divergences: docs/poc-artifacts/t-79d-chrome-doors.md.

> **Dated note (2026-10-15, ADR-0004 Am.3 — owner stamp "OPTION B CONFIRMED", rider R-0):**
> the `lock only for telegram modes` clause below and the mark-read/msg-id steps it guarded
> fall away with the MTProto client — no Telegram lock exists in the Go port; the queue source
> is the Worker `/api/pending` (spec-gui-pages-a, bot_queue.py note). The batch-running state
> machine itself stays.

- `start_processing` (61-121) → `_start_worker_with_urls` (123-148) → `_start_worker`
  (1686-1765): lock only for telegram modes; `_batch_running=True`; hero→'running' (Stop);
  bar reset "Processing…"; `_clear_log()`; `ProcessingWorker(...)` with bot source/intake
  counters injected BEFORE start; the 8 queued signal connections.
- `update_progress` (1777-1812) / `update_status` (1814-1843): bar + count + pretty
  `owner/repo`; `_schedule_progress_hide(2500 ms)` flashes the result then resets to Ready.
- **`processing_finished` (1870-2230) — the completion chain, in order**: hero→SYNC + retry
  banner + `_record_last_sync` + counter refresh → owner-scoped lock release → optional
  folder backup → **GitHub vault seal** (VaultSealWorker QThread) → **GoodRepos publish** →
  **Websites seal** → **Personal seal** → elapsed → success path: log + tip + "✅ Done" bar +
  pipeline state done + hide timer; then **Phase-5 CLEAR** (bot batch only): manifest
  `get_all_clear()` → `_mark_bot_messages_read()` else pending-links listing (max 5 +
  websites-waiting line); `last_processed_msg_id` advanced only if all_clear; then fanfare when
  `batch_summary` and not stopped: hand-queue delivery → hand-note redo → Chrome tab-retry
  offer → scorecard (stashed if Chrome gathering; merged on the delivery re-run) with
  needs_retry chime routing. Failure path: ❌ log, "❌ Failed" bar, error box.
- Retry drivers: `_startup_review_backlog_check` (175-243), `_scan_master_waiting` (340-390) +
  `_start_master_retry`, `_maybe_deliver_hand_queue` (475-520), `_start_review_retry`
  (1673-1684).
- Modals: `_confirm_batch` (>10 items); the **LLM-failure modal** (2510-2697): model dropdown +
  Remember + Skip / RetrySame / RetryWith / Stop, feeds `resolve_llm_failure`; **disk-full**
  (2495-2508): question modal + paused spin until resolved/stopped; **Batch Complete
  scorecard** (2309-2422+): summary table + needs_retry chime.
- Sounds: `_celebrate_batch` / `_play_batch_sound` (2236-2288) — success.wav natural /
  retry.wav failures.

## Go parity requirements

1. The completion chain **order** is behavioral (seals only after success; CLEAR only on
   all-clear; msg id advance gated).
2. Progress formats and hide-flash timing (2500 ms) preserved.
3. The three interactive modals with the same options and timeouts (LLM 600 s).

## Tests

Completion-chain order test (fake workers asserting call sequence); all-clear gating of
mark-read + msg-id advance
(re-scoped 2026-10-15 per ADR-0004 Am.3/R-0: no mark-read/msg-id advance exists in Go — the
all-clear gate's Phase-5 re-spec is an open item for T-62, see spec-cli's Phase-5 CLEAR note);
retry drivers wiring; modal option outcomes.
