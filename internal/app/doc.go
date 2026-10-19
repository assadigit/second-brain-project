// Package app is the composition layer (doc 04 §2/§5 — the Qt
// replacement seam): the App struct (the config source, the injected
// engine doors), the typed event bus with the ask-gates, the command
// surface both the GUI and the CLI drive, and the lifecycle
// (single-instance lock, the one-batch-at-a-time op lock + watchdog,
// startup passes, shutdown graces).
//
// It is the production evolution of the guibridge/ seam (T-70, the
// Phase-6 entry gate's real-job wiring) and webui/deskbus (the
// coalescing log desk's Go half); guibridge RETIRED at T-72
// (SESSION-035, deleted with its tests): the shell binds this
// package DIRECTLY from the nested webui module (ADR-0002 Am.4's
// probe). The full dissolution into the main module is BLOCKED by the
// authorized-pins law — wails v2.16.0's own go.mod demands x/net
// v0.56.0 + the 1.25.0 directive; the probe record + the owner's
// pending call live in ADR-0002 Am.5.
//
// The files:
//
//   - events.go — the Bus (the 8 queued Qt signals' transport) + the
//     three ask-gates with the reference's identical timeouts.
//   - commands.go — the command surface (doc 04 §5's T-71 table) over
//     injected seams; the crash-safety wrapper.
//   - settings.go — the Settings pages' command surface (T-73): the
//     LIVE-CONFIG law (the session-held map, the GUI load path at
//     construction, in-place saves) + the six pages' engines (the ONE
//     merge save with all four v30 laws, the ladder, the token test,
//     the wizard, the Ollama/llama.cpp detect family, the quick-switch
//     fast lane, the sound gate) over injected SettingsDeps.
//   - settingsb.go — the Settings pages B's engines (T-74): the queue
//     report rendering + the badge + the log family (the canned-result
//     law), the dashboard engines (the frontmatter scan, the quarantine
//     table/reset, the undo engine), the sources fetch over its seam,
//     and the Backup page's ladders + save groups.
//   - batchwire.go — THE CLI-PARITY SHELL WIRING (T-74, the debt T-73
//     carried closed): engine.go's runImpl composition re-hosted on
//     this surface — the warmup family, the inputs fetchers, the full
//     intake, the banishment + settled-ledger + backfill + §4.4
//     passes, the websites phase, the finish seals + THE PHASE-5
//     CLEAR + the dry-run law.
//   - lock.go — the OpLock (the TelegramLockManager port — the
//     one-batch law).
//   - lifecycle.go — the single-instance app.lock + PID healing, the
//     watchdog, the startup passes, the shutdown graces, the sentinel
//     re-arm door; process_posix/windows.go — the PID liveness probe
//   - the detached `ollama serve` spawn twins.
//   - app.go — the DI root (New wires the production seams; the
//     batch composition; the startup passes' engines; the settings
//     surface's production seams; the live-config load + the fill
//     line).
//   - queue.go — the _bot_queue_job classification half.
//   - scanjob.go — the _vault_scan_job engine (both confirm doors).
//   - backupjob.go — the BackupWorker port (the vault zip).
//
// Spec: docs/port-blueprint/04-target-go-architecture.md §2/§5; task
// T-71 (Phase 6), landed SESSION-034; task T-73 landed SESSION-036.
package app
