# Module Spec — Settings pages B: `gui/main_window/backup_seal.py`

- **LOC**: 1,477 · **Go home**: `internal/app` (drift/fast-sync commands — the **pure halves
  port**) + `ui/frontend/src/pages/backup` · **Status**: split port/rehost
- **Responsibility**: the Backup settings page (local zip backup, VaultSeal GitHub mirror,
  GoodRepos public directory, dashboard link), all seal/publish/resync/export/restore actions on
  QThreads, the v0.89.0 **drift seal** and the v0.90.0 **⚡ Fast Sync** door.

## The pure half (ports to Go logic — headless-testable)

- `_DRIFT_SEAL_LEGS (64-74)`: repos/websites/personal guards mirroring the batch-end conditions.
- `_drift_seal_plan(config) (77-86)`.
- `_run_drift_seal(config, log, manual) (89-198)`:
  - **THE QUIET LAW** (automatic): a leg that seals nothing and fails nothing stays SILENT
    (returns `('', 'info')`); one line per leg when the chain moves.
  - **THE MANUAL DOOR** (Fast Sync, `manual=True`): ALWAYS-SPEAK ("the triad is already in
    sync — nothing to seal, nothing to publish") and ALWAYS-PUBLISH (the manual publish runs
    whenever the websites leg is enabled — a clean vault with a stale site is exactly the drift
    the fast sync exists to catch; the publish is idempotent).
  - Guards: never while a batch runs; never while another seal is in flight.
  - Verdict strings: ⚡ fast-sync summary / 🛡 drift-seal line + publish line.

## The page (rehost contracts)

- Four QGroupBoxes (Vault Backup / VaultSeal — GitHub Mirror / Good Repos — Public Directory /
  Dashboard), compacted rows, status dots `● Ready` / `● Local-only (no GitHub token…)` /
  `● No vault selected` / `● Disabled`, `?` help glyphs.
- `_backup_now` (548-640): inner BackupWorker QThread — zip with
  `EXCLUDE={'config.json','.git','__pycache__'}`, suffix `.lock`, prefixes
  `.obsidian/workspace`/`.obsidian/app.json`, `compresslevel=6`, FIFO rotation `backup_max`
  (default 10), plus **`GitCurator-data/sync_review.db` inside every zip** (v0.86.0).
- Seal starters (`_start_vault_seal` / `_start_websites_seal` / `_start_manual_seal`,
  652-784): buttons disable + relabel ("Sealing…"); `done(bool,str)` results.
- `_start_drift_seal (790-862)` + `_drift_seal_result`; **`_on_fast_sync_clicked` (892-938)**:
  "Fast Syncing…" until the verdict; refuses while batch/seal runs.
- Repo radar (`_radar_detect/_radar_run_worker/_radar_finished`, 944-1051): read-only detection
  on a config copy; typed names never overwritten; ambiguous reported.
- `_vaultseal_now` / `_vaultseal_resync` (warned modal: deletions included, history kept, never
  force-pushed) / `_vaultseal_refresh_status` (status ladder: Ready / Local-only / No vault /
  Disabled).
- GoodRepos publish/result/status (1196-1272); `_backup_export_zip` / `_backup_restore`
  (restore with safety-net backup); dashboard openers.

## Go parity requirements

1. The pure `_run_drift_seal` semantics (quiet vs manual laws, guards, leg order) port exactly —
   they are already covered by Python tests (`test_driftseal`, `test_fastsync`) to re-port.
2. Backup zip law: the same EXCLUDE set, compression level, FIFO rotation, and the
   sync_review.db rider.
3. Button state machines and status dot ladders.

## Tests

Port `test_driftseal` (24) + `test_fastsync` (27): quiet-when-clean, one-line-per-moving-leg,
always-publish (stale-site case), always-speak, not-landed publish warning without doubling a
raised exception, deletions ride the same legs, automatic-path regression pins; backup zip
content tests (excludes + ledger rider).
