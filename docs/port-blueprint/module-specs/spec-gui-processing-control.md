# Module Spec — `gui/main_window/processing_control.py`

- **LOC**: 2,746 · **Go home**: `internal/app/commands.go` + `ui/` bindings · **Status**: rehost
  (completion chain is the contract)
- **Responsibility**: PROCESS dispatch, worker lifecycle, progress/status rendering, the whole
  end-of-batch chain (seals, publish, Phase-5 CLEAR, fanfare), the LLM-failure and disk-full
  dialogs, and the websites retry drivers.

## Key flows

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
mark-read + msg-id advance; retry drivers wiring; modal option outcomes.
