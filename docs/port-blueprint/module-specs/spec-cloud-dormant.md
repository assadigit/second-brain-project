# Module Spec — Dormant cloud quartet: `cloudflare_manager.py`, `cloudflare_gui.py`, `gdrive_backup.py`, `gdrive_gui.py`

- **LOC**: 605 + 531 + 736 + 612 (≈2,584) · **Go home**: none in Phase 1 · **Status**: drop
  (dormant by owner decision "fixed, not deleted" — AGENTS.md; documented here for a possible
  future revival)
- **Evidence of dormancy**: `gui/main_window/window.py:176` sets `self.cf_manager = None` and
  nothing ever assigns a `CloudflareManager`; `cloudflare_gui`/`gdrive_gui` panels are referenced
  only by `tools/integration_snippet.py` (a documentation snippet); `gdrive_backup` imported
  under try/except but never instantiated.

## Rough API surfaces (for a future revival spec only)

- **cloudflare_manager.py** — `CloudflareManager(QThread)` with signals
  `pending_links_received(list)`, `status_updated(str)`, `error_occurred(str,str)`,
  `backup_completed(str)`, `backup_failed(str)`, `auth_required(str)`, `sync_completed(dict)`,
  `poll_completed(int)`; `run()` 1 s loop polling `/api/pending` every 300 s, pushing errors
  every 60 s, debounced batch push (5 s), debounced GDrive backup (5 s);
  `on_batch_complete` also fires `directory_sync.publish_after_batch` on a daemon thread;
  `on_decommission` real-time push; start_backfill / complete_cutover / stop.
- **cloudflare_gui.py** — `CloudflareSettingsPanel` (pair form + enable toggle + 5 s status
  refresh + unpair + force sync + backfill dialog).
- **gdrive_backup.py** — `GDriveBackup(config)`: OAuth 2.0 desktop flow (loopback
  `localhost:8765/callback`, 5 min timeout; scope `drive.file`; `access_type=offline`,
  `prompt=consent`), token refresh, Drive v3 REST (multipart upload, search, delete, download);
  full-vault zip backup with the same exclusions as local backup into folder
  `Obsidian-Vault-Backups`, FIFO keep 10; `restore_to_new_folder` / `replace_current_vault`
  (with safety-net) / `download_backup` / `cleanup_safety_nets`; pushes
  `sync.push_backup_status` after upload. Config keys `gdrive_*` (71-81).
- **gdrive_gui.py** — settings panel + restore dialog + OAuth/ListBackups/Restore QThreads.

## Disposition in the Go port

1. **Not ported** in the migration phases (no behavior is lost — nothing wires them today).
2. The `cloudflare_*` and `gdrive_*` **config keys keep round-tripping** (merge law preserves
   unknown keys; the user-memory prefix families include them).
3. If the owner requests revival later, write a new spec from this file first; the Worker-side
   endpoints (`/api/gdrive/backup_status`, `/api/errors`) already exist and are documented in
   `spec-worker-contract.md`.
