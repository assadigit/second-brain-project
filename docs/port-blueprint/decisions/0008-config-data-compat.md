# ADR-0008 — Configuration, data, and on-disk compatibility policy

- **Status**: accepted
- **Date**: blueprint session (2026-10)

## Context

The Python app's user-visible persistence:

- `app/config.json` — 109-key JSON config (template `app/config.example.json`; defaults in
  `gitcurator/constants.py:122-305` `CONFIG_EXAMPLE`), saved with a **merge-never-rebuild** law
  (`gui/main_window/vaults_config.py:260-281` — unknown keys survive).
- `Documents/.gitcurator` — the user-level credential/vault memory (fill + mirror, chmod 600,
  `core/user_config.py`), env override `GITCURATOR_USER_CONFIG`.
- `app/cache.db`, `app/sync_review.db` — SQLite ledgers (ADR-0003).
- Vaults (Obsidian markdown, three of them), `links_manifest.json`, `_inbox` platform tables,
  `DECOMMISSIONED.md` master table, hand-delivery queue, reports, undo files.
- Windows conventions: Known-Folder Documents via `SHGetKnownFolderPath`, `.bat` launchers, no
  symlinks (SPEC.md §0: "open every file with explicit UTF-8 … do not use symlinks").

## Problem

The Go app must run against the **same** user data — the owner's real vaults (1,309 notes), real
config, real cache.db — without a destructive migration and without silent divergence.

## Decision

1. **Same file paths**: the Go app resolves `config.json`, `cache.db`, prompts, taxonomy, sounds,
   fonts, `session` relative to the executable's directory exactly like `APP_DIR`
   (`gitcurator/constants.py:28-30`). The Go exe can be dropped into the same `app/` folder
   (side-by-side install during transition) or pointed at the folder via flag/env.
2. **config.json is read/written with the same merge law** (unknown keys preserved; nested dicts
   merged key-by-key; tuning keys only defaulted when absent — algorithm pinned in
   `spec-core-storage.md`). `json.MarshalIndent` with 2-space indent and **no HTML escaping**
   (Go's `json.Marshal` escapes `<`, `>`, `&` by default — must use an encoder with
   `SetEscapeHTML(false)` to keep files byte-comparable; flagged as a parity test).
3. **User-config memory (`Documents/.gitcurator`)**: same filename, same JSON marker key
   `_gitcurator_user_config`, same fill/mirror semantics and `0o600` permission on Windows
   (`os.Chmod` is advisory there; keep the call for POSIX parity).
4. **SQLite files opened in place** (ADR-0003) — schema-compatible, no export/import step.
5. **Dry-run shadow cache** behavior preserved (`core/dryrun.py` — temp-dir copy incl. WAL
   sidecars).
6. **Telethon session is the one accepted break** (re-login; ADR-0004). Everything else must be
   non-breaking.
7. **Version stamping**: the Go app reports the same semantic version lineage (starts at the next
   minor after 0.91.0 during development, e.g. `0.92.0-go.1` milestones — exact scheme in
   `10-build-release-packaging.md`), and its UA string format stays
   `GitCurator/<version> (+https://github.com/assadigit/GitCurator)`.

## Consequences

- Differential testing can run Python and Go against copies of the same fixtures/config and diff
  outputs (see `09-testing-and-parity.md`).
- Side-by-side operation means **cache.db contention** if both apps run at once — the Python app's
  single-instance lock (`app.lock`, `gui/app.py:393-440`) is file-based; the Go app must implement
  the same lock file semantics (same filename, PID liveness healing) so the two apps mutually
  exclude. Documented as a hard requirement.

## Evidence

- Merge law: `gui/main_window/vaults_config.py:260-281, 452-474`
- Single-instance lock: `gui/app.py:393-440` (PID check, zombie-handle retries)
- APP_DIR anchoring: `gitcurator/constants.py:28-30`; `core/note_state.py:188-191`
