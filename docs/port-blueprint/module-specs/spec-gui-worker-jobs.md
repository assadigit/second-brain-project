# Module Spec — `gui/worker_jobs.py`

- **LOC**: 1,287 · **Go home**: `internal/app/commands.go` (the job registry) ·
  **Status**: port (pure halves) / rehost (threading)
- **Responsibility**: the background-job registry run inside TestWorker threads — the
  heal-job pattern home. Every job: pure-ish callable taking (args…, log signal), never
  raises, streams via signal, returns a plain dict.

## Jobs (with their result contracts)

| Job | Lines | Contract |
|---|---|---|
| `_run_telegram_worker(config, log_signal, code_callback)` | 50-73 | delegates to subprocess_runner (legacy timeout ignored) |
| `_telegram_test_job` | 75-91 | connection battery telegram leg |
| `_bot_queue_job(config, …)` | 93-391 | fetch via Telethon subprocess, then **worker-thread vault filtering**: GitHub VaultIndex rebuild, dead set, one-time `settle_existing_repos(queue history)`; classification buckets (blocked/self/in_vault/decommissioned/settled/pending) with the PLATFORM half of the law for repos; websites twin (WebsiteStateDB probes + `settle_the_ledger(queue_urls)`); result gains `pending_urls`, `pending_website_urls` + ~12 count fields |
| `_connection_battery_job` | 393-434 | runs `connection_check.run_local_checks` with section/result callbacks |
| `_quick_detect_job(provider, config, log)` | 436-497 | detect → model list → pick |
| `_scan_llm_call(config, log)` | 500-559 | provider router with `models.vaultscan` task override |
| `_sync_display_job(config, log_signal)` | 562-614 | READ-ONLY: `collect_local_facts` always; bot queue + published index only because Refresh was clicked; each leg degrades to its honest error state |
| `_vault_heal_job` | 617-644 | hygiene both machine vaults |
| `_repo_radar_job(config, log_signal)` | 647-683 | read-only detection on a config COPY |
| `_vault_scan_job(config, log_signal, confirm_gui, progress, confirm_selection, vaults)` | 686-1287 | one-open NoteReadCache (logs "⚡ N note open(s) — M served from the last scan's memory"); hygiene both vaults; scope picker (`vaults=None`=both / list=chosen alone — a chosen-but-not-connected vault speaks its own name); per-vault story via `_scan_one_vault`: build plan → log proposal lines → confirm door (`_channel_for`: GUI ScanPlanDialog when injected and `scan_confirm_door != 'telegram'`, else Telegram round-trip; 300 s window) → picked counts → `apply_scan_plan`; merged verdict; DB close honesty line |

## Go mapping

Each job becomes a command method (doc 04 §5) with the same result fields (the GUI fallbacks and
tests read them). The heal-job pattern becomes: command goroutine + bus events; the ask-gate via
channel.

## Tests

Per-job result-shape tests (fake fetchers/LLM); scan scope law (None vs list); confirm-door
routing; queue-job classification counts vs known fixture vaults; read-only guarantees of
`_sync_display_job` (no writes asserted).
