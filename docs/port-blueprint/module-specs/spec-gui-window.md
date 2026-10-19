# Module Spec — `gui/main_window/window.py` + `gui/main_window/lifecycle.py`

- **LOC**: 756 + 282 · **Go home**: `internal/app/lifecycle.go` (startup/shutdown orchestration)
  + `ui/` shell · **Status**: rehost (the startup/shutdown *sequence* and cadences are the
  contract)
- **Responsibility**: the MainWindow mixin composition + startup passes + sentinel painters;
  the about-me wizard + the full shutdown chain.

## window.py (`gui/main_window/window.py:`)

- **Mixin MRO** (93): `LifecycleMixin, BackupSealMixin, BotQueueMixin, DashboardMixin,
  SyncFlowMixin, ProcessingControlMixin, TelegramUiMixin, InputProxyMixin, VaultConfigMixin,
  TestConnectionModalMixin, VaultScanUiMixin, SyncViewMixin, LinkingToolsMixin, LlamaCppMixin,
  ConnectionTestsMixin, HeroMixin, UiMixin, ThemeMixin, QMainWindow` (file↔domain table in
  `02-current-architecture.md` §2).
- Class-level queued signals: `_llamacpp_autodetect_signal(dict)`, `_sentinel_signal(dict)`,
  `_sync_matrix_signal(dict)`.
- `__init__` (115-451): config load → `initUI()` → retry-queue truth pass (stale rows ♻,
  missing rows cry only post-settlement) → LinkTracker manifest reconciliation (healed hits ✓;
  failed links ⚠ max 5 listed; set-aside websites/settled honest lines) → `_refresh_retry_banner`
  → timers/single-shots: **sentinel 60 s + first pass @1200 ms**; `_sync_matrix_signal` connect;
  tg watchdog 60 s; llamacpp autodetect @1500 ms; review-backlog @2500 ms; quiet vault heal
  @2000 ms; repo radar @3500 ms.
- `_run_sentinel_pass` (544-594): raw daemon `threading.Thread(name='gitcurator-sentinel')`
  (sockets never on the GUI thread — the v22 freeze fix); results painted via queued signal.
- `_apply_sentinel` (596-660): `offline_verdict` transitions — warn-now → ONE modal at open;
  fell-offline/back-online → log lines only. `_apply_net/ai/proxy_status` paint the top-bar
  labels (words + colors, never bare dashes).
- Startup heal/radar deferred when a batch/scan is running.

## lifecycle.py (`gui/main_window/lifecycle.py:`)

- `show_about_me_wizard` (58-171): first-run modal (name/role/interests/objectives/tech/vault
  usage) writing `about_me.md` in CWD.
- `closeEvent` (173-261): while `worker.is_running` → terminate-confirmation modal BEFORE
  `_closing`; then: stop timers → `save_config()` + `_save_window_size()` (QSettings
  `GitCurator/MainWindow`, key `main_window/size`) → unblock worker (`resolve_llm_failure('stop')`,
  `provide_code('')`, `_disk_full_paused=False`) → `worker.stop()` + `wait(5000)` → each test
  worker `wait(2000)` → kill orphan Telethon subprocesses → force-release the telegram lock →
  accept → `QApplication.quit()` (the hidden-SettingsDialog zombie fix)
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** the `kill orphan Telethon subprocesses`
  and `force-release the telegram lock` shutdown steps fall away in the Go port (no MTProto
  client, no lock — ADR-0004 Am.3); the rest of the sequence (unblock, stop, wait, quit) stays.
- `_unblock_worker_for_shutdown` (263-281): the three unblock doors above.

## Go parity requirements

1. Same startup order and **exact cadences** (1200/1500/2000/2500/3500 ms single-shots; 60 s
   periodic), same quiet-open law, same offline warn-once behavior.
2. Same shutdown chain and grace windows (5 s batch, 2 s jobs), child killing, lock release.
3. Window geometry persistence (size key) — via the UI framework's settings store.
4. The about-me wizard writes the same file.

## Tests

Startup cadence test (events observed at the right delays, nothing network before sentinel);
shutdown chain test (worker blocked on auth → close → unblocked → stopped within windows);
offline warn-once; wizard file write.

## Dated landing note — T-77 (2026-10-20, SESSION-040)

The sentinel pass's chrome half landed (the labels + the offline law):

1. **The snapshot emission**: the pass now runs the FULL
   `SentinelSnapshot` (net + ai + proxy — the reference's
   `_sentinel_signal` dict) and crosses it as the `Sentinel` bus
   event; the Net · AI · Proxy labels paint from it (the paint laws
   live in Go — see spec-gui-ui's note). The offline verdict
   transitions (warn-now / fell-offline / back-online) speak the
   reference's lines; the warn-now MODAL rides the verdict word to the
   frontend (the reference's own box text).
2. **TWO reality fixes, found re-reading the pass** (the law pays in
   bugs): the old 🛰️ line dereferenced `net.LatencyMs` unconditionally
   — `ProbeInternet` leaves it nil on the offline path, so an offline
   pass PANICKED into the crash-isolation and the labels never painted
   (the latency is nil-safe "n/a" now); and the sentinelOnline seat
   RACED (the lifecycle's recurring tick + the save law's 250 ms
   re-arm can pass concurrently — the read-transition-write is one
   unit under sentinelMu, mirroring the reference's GUI-thread
   serialization).
3. **The startup cadence acceptance**: the lifecycle's own battery
   (the cadences, the order, the crash isolation, the cancellation)
   re-verified green + the pass's own emission test
   (TestSentinelPassEmitsSnapshotAndOfflineLaw, over the hermetic
   SentinelProbe seam per the HERMETIC-SEAM law).
4. **The shutdown chain**: unchanged (the drain order tested at T-71);
   the hero's mirror loop stops first at shutdown.
