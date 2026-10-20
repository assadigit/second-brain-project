# SESSION-048 — 2026-10-20 — Phase 7 (opening row): T-80 COMPLETE (the deterministic zip builder + version stamping + manifest/icon + the standing assignment) + the 44th-wipe repair; 2371 main-battery tests over 36 packages + 89 webui over 4

## Goal (from NEXT-SESSION.md at start)

T-80 (M): the deterministic zip builder + version stamping + manifest/icon
(doc 10 §2/§5): the reproducible packaging — the zip's byte-identical
rebuild (the acceptance: sha256 equal on two builds from the same commit).
The export zip + the Backup page's Restore button ride the same packaging
composition (the standing assignment — SESSION-040's record + the
honestly-disabled More-menu row + the Backup page's disabled Restore note).

## What was done (task IDs, files, commits — cite hashes)

- **The forty-fourth full wipe repaired BEFORE work**: the ENTIRE environment
  was gone at boot (both repos, the toolchain, the venv, the wails CLI).
  Everything rebuilt from the committed recipes: both repos re-cloned with
  the ephemeral GIT_CONFIG_COUNT PAT form (scrubbed; `git remote -v` clean);
  the reference at `98f4fc71cfb862cee7225d864636c08062724348` =
  testdiff/PINNED-GITCURATOR-REV exactly; go1.27.2 from go.dev with the
  local sha256 matching TOOLCHAIN.md Am.2's committed 64-char pin
  byte-for-byte (`ecbadb99091a…`); the venv (Python 3.12.15, the plain
  ensurepip form); the wails CLI v2.16.0. **No webkit rebuild was needed** —
  the windows cross-compile leg is CGO-free (the WebView2 loader is pure
  Go), and no linux GUI build was required this session. The FULL battery
  re-verified green pre-work: **2356 main + 84 webui, plain + race, exactly
  the committed numbers**.
- **T-80 core (commit `cb84491`)**:
  - `build/mkzip` — the FULL builder: the two-exe composition (`-gui-exe`
    joins `-exe`: GitCurator.exe + GitCurator-cli.exe + the flattened
    launchers + the GUI-side pair + the assets + the stamped VERSION —
    13 files), **THE VERSION-STAMP GUARD** (both exes must carry the
    -version stamp's bytes — a forgot-the-ldflags build exits 2) and **THE
    BRAND-ICON GUARD** (the GUI exe must carry webui/build/windows/icon.ico's
    bytes); +5 new law tests (the two guards, TestManifestLaws,
    TestIconLaws, TestWailsInfoLaws — the one-version-truth: wails.json's
    Info.productVersion pinned to the repo VERSION file).
  - `build/release.sh` — **THE DOUBLE-BUILD RITUAL** (AC-22's shape): both
    exes + the zip built TWICE + one cold-GOCACHE leg per exe (the
    fresh-machine story), sha256-compared; the preflights (the
    one-version-truth drift guard + the clean-tree law with --allow-dirty
    for dev probes); the reproduce-from-tag recipe in the header.
  - The manifest laws (doc 10 §5): supportedOS Win10
    (`{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}`) + longPathAware beside the
    wails template's DPI-permonitorv2 section.
  - The version stamping: `webui/main.go`'s `var version = "dev"` (the -X
    target) + the reference's own boot print (app.py:30-38's
    `[main] LOADED version …` stderr line — the reference's own shape) +
    wails.json's Info block (ProductVersion 0.92.0-go.1 → the exe's version
    resource) + the quickstart's two-exe story + `/reports/` gitignored.
- **T-80's standing assignment (commit `20356d6`)**:
  - `internal/app/exportdoors.go` — the ExportWorker's data half
    (backup_seal.py:1274-1344: the exclude laws, the arc-root law, the
    ledger rider, the done message) + the restore family (1356-1441: the
    newest-first list, the ⚠️ confirm body, the extract-to-
    vault-restored-<ts> safety net with the path-escape belt, the copy2
    twin) + the guards' exact boxes; +10 engine tests (THE
    REFERENCE'S-OWN-ARITHMETIC law paid once: the rider's arcname is at
    the ZIP ROOT — `zf.write(_ledger, 'GitCurator-data/sync_review.db')`
    — the first test expectation was wrong, not the engine).
  - **THE RIDER FIX (reality wins)**: `backupjob.go` gains the v0.86.0
    deferred-ledger rider — T-71 had recorded its absence against the
    then-placeholder syncreview seat; the ledger is real since Phase 3/4,
    so BOTH zips (backup + export) carry it now (pinned by
    TestBackupRiderLanded).
  - The shell doors: `HeroUI.MoreExportBackupZip` (the No Vault guard, the
    wails save-dialog seam with the default name's stamp law, the engine
    call, the exact 📤/❌ desk lines + the Export Complete/Failed boxes; a
    cancel is a silent no-op) + `PagesBUI.RestoreBackupsList /
    RestoreBackupRun / CopyBackupTo` (the guards, the rows payload, the
    extract, the native copy dialog).
  - The frontend: the More menu's 📤 row FLIPS LIVE (the in-flight
    Exporting… belt — the reference's setEnabled/setText law) + the Backup
    page's Restore button FLIPS LIVE (the save-first order, the guard
    boxes, the selectable list dialog with the three buttons pinned
    OUTSIDE the scrolling body, the exact confirm question, the result
    boxes) + the restore-list CSS + **THE BINDING-IMPORT LAW honored**
    (MoreExportBackupZip + the three restore doors joined the import
    statements, pinned by tests) + the two stale tail-pins updated honestly
    (the More menu has NO disabled row left) + the wailsjs bindings
    regenerated over the windows cross-compile leg (content-only; the 755
    baseline intact — `git diff --raw` verified).
- **The evidence doc**: `docs/poc-artifacts/t-80-reproducible-packaging.md`
  (the ritual's table, the finds' byte-level proofs, the manifest laws, the
  honest boundary, the reproduce-from-tag recipe).
- **The spec updates (reality wins, same commits)**: doc 10 §2's [verify]s
  settled with dated notes (the GUI command's reality — the wails build
  over the NESTED webui module, not the spec's `./cmd/gitcurator` path;
  the -w -s note; the -X-under-trimpath law) + doc 10 §5's UA reality
  note (the reference's own Worker UA is hardcoded at 0.91.0 — the Go twin
  hardcodes the same string, THE CROSS-BINARY DISAGREEMENT LAW; the
  stamp's homes are --version, the boot print, the version resource, and
  the zip's VERSION member) + the icon [verify]/OQ-9 resolution (the seed
  commit's brand icon is the chosen source asset).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-80's acceptance = "rebuild byte-identical (sha256 twice)" —
**GREEN, evidence-first** (the ritual ran on the committed tree `20356d6`,
toolchain go1.27.2, wails CLI v2.16.0):

| Artifact | sha256 A (warm) | B (warm) | cold leg |
|---|---|---|---|
| GitCurator-cli.exe | `82bfa4ecf575ead9bdcc3bf6dde57d8573ae7f9e45c0b71977d34da8325610e0` | identical | identical |
| GitCurator.exe | `99cc22e6810f6881a19dc5bd6f702d9d8e653a4a1eae7994a705e95bbd002422` | identical | identical |
| the zip | `61b4cf2adc2eec96c70ae7962ac5bd89f70ac4d644eedefdb0b5b02cda527ff4` | identical | — |

- **The release zip**: `GitCurator-v0.92.0-go.1-windows.zip` — 13 files,
  56,908,881 bytes in, 27,226,166 bytes zipped (T-65's CLI-milestone zip
  was 12 files / 15,988,399 — the GUI exe is the thirteenth member).
- **The version stamping proven byte-level**: the stamp's bytes present in
  BOTH exes (`0.92.0-go.1` in the binaries post-fix; the version resource
  carries ProductVersion/ProductName/copyright/comments — UTF-16 probes
  all True); the zip's VERSION member stamps `0.92.0-go.1`.
- **The icon proven byte-level**: the brand ico's 256px head+tail and the
  128px head are present in the built exe; the exe's sha256 CHANGED with
  the icon swap (`84e64b60…` → `a97687fd…`) — the icon rides the .syso.
- **The standing assignment**: the export/restore doors pinned by 15 new
  main-battery tests + 5 new webui tests (the wiring, the field tags, the
  binding imports, the exact reference strings); the ledger rider on both
  zips pinned.
- **The battery at the close: 2371 main-battery tests race-clean over 36
  packages (2356 + 15) + 89 webui tests over 4 (84 + 5)**; the touched
  packages (internal/app, build/mkzip, webui) 3× race-clean (real
  executions, -count=1); gofmt/vet clean over both modules; the MAIN
  go.mod + go.sum byte-unchanged (verified after every wails build — the
  tidy runs are no-ops); the wailsjs 755 baseline intact (content-only
  diffs).
- **The honest boundary**: the live Windows runtime legs + the clean-VM
  checklist are the owner's Phase-7 items (T-82; ADR-0001 Am.2's
  register); the release automation is T-81 (`build/release.sh` is its
  reusable core); no tag, no GitHub release was cut this session (the
  per-session law: commit + push + reports only).

### The finds (byte-settled, each recorded in the evidence doc)

1. **THE -X-UNDER-TRIMPATH DROP** — the linker drops an UNREFERENCED
   package-level var's -X bytes under `-trimpath` (isolated empirically:
   unused+trimpath → absent; referenced+trimpath → present; no-trimpath →
   present either way). The fix: the reference's own boot print
   (app.py:30-38) references the var — the stamp lands. This is a NEW
   standing law for any future -X stamping.
2. **THE DEFAULT-ICON EXE** — `webui/build/appicon.png` was byte-identical
   to the wails v2.16.0 TEMPLATE default (sha256 `28332a3b…`), so every
   exe since T-70 shipped the wails icon, not the brand. Fixed + pinned.
3. **The winres version normalization** — the manifest's
   `assemblyIdentity version="{{.Info.ProductVersion}}.0"` is parsed by
   winres into the dotted quad `0.92.0.1` inside the exe (cosmetic; the
   human-visible version resource carries the full stamp). Recorded.
4. **The renderer ate characters TWICE on the verification itself** — the
   Edit tool's space-vs-tab corruption struck webui/main.go (gofmt -w
   restored; the law held), AND a tool-output rendering of `[main]` as
   `ain]` nearly caused a phantom "fix" of a CORRECT file — the hex dump
   proved the file right (the standing settle-with-bytes lesson, now with
   a sharper edge: verify the VERIFIER too).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/10-build-release-packaging.md` — §2's dated note
  (the [verify]s settled: the GUI command's reality, -w -s, the
  -X-under-trimpath law, the ritual pointer) + §5's UA reality note (the
  hardcoded truth — reality wins) + §5's icon [verify]/OQ-9 resolution.
- `docs/port-blueprint/13-implementation-task-list.md` — the T-80 row
  COMPLETE with the full record.
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-048 entry.
- `docs/port-blueprint/module-specs/spec-gui-pages-b.md` — the dated note
  superseded (the Restore button + the export row LANDED; the rider's
  zip-root law recorded).
- `README.md` — the SESSION-048 status block.
- `webui/README.md` — the icon/manifest/stamp record (the tail's
  brand-mark TODO resolved) + the heroflow.js/settings.js map entries
  extended with the new surfaces.
- `docs/poc-artifacts/t-80-reproducible-packaging.md` — the evidence doc.
- `docs/NEXT-SESSION.md` — rewritten for SESSION-049 (the state card v49).
- `docs/sessions/SESSION-048.md` — this report.

## Blockers & decisions needed

- **No new blockers.** T-80 closed clean, acceptance GREEN.
- Carried, unchanged: the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (now the
  thirty-ninth session carrying it); x/net/html keep-or-swap; the
  answering-window belt + the T-71 merged-verdict question
  (SESSION-039's record); the deferred test debt; the batch-scorecard
  modal's open question (the owner may name it a Phase-7 row or leave the
  bar+chime fanfare as the release shape).
- **The v0.91.0 zip parity question (the owner's, NEW, minor)**: doc 10
  §10 registers the Python release artifact as "the parity target for the
  Go zip ritual" (the DISCIPLINE, not the bytes — the compositions differ
  by design: 157 Python files vs the Go two-exe portable shape). The
  ritual now matches the discipline (double-build sha256 identity +
  tracked-only + the fixed stamps). Recorded here for the owner's
  visibility; no action needed unless the owner wants a different
  composition reading.

## Owner actions requested

- None urgent. When convenient: the standing calls (dissolution,
  vanilla-vs-vite, the scorecard-modal question) + a look at doc 10's
  amended notes (the GUI command's reality + the UA truth) at review
  time.

## Handoff

- **SESSION-049 = T-81 (S): the release workflow** (doc 10 §8: the tag →
  build → mkzip → attach-zip automation over `build/release.sh`; the
  dry-run on a test tag is the row's acceptance). T-82 (the clean-VM
  checklist + the AV note) + T-83 (the docs) follow — one slice done
  properly with proof beats two half-done.
- NEXT-SESSION.md rewritten: yes (the state card v49, the SESSION-049
  prompt inside).
- Push state: all work committed + pushed at the close (the ephemeral
  PAT; never stored, never written).
