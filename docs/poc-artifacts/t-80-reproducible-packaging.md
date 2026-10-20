# T-80 — the reproducible packaging evidence (SESSION-048, 2026-10-20)

> The deterministic zip builder + version stamping + manifest/icon — doc 10
> §2/§5, AC-22's acceptance ("the zip rebuilds byte-identical, sha256 match on
> two builds from the same commit"). Everything below is byte-level evidence,
> reproducible from this commit with `build/release.sh`.

## 1. THE DOUBLE-BUILD RITUAL (the acceptance proof)

`build/release.sh` at commit `20356d6` (the clean-tree law — the ritual ran
on a committed tree), toolchain go1.27.2, wails CLI v2.16.0:

| Artifact | sha256 (A = warm build 1) | sha256 (B = warm build 2) | sha256 (cold-cache leg) |
|---|---|---|---|
| `GitCurator-cli.exe` | `82bfa4ecf575ead9bdcc3bf6dde57d8573ae7f9e45c0b71977d34da8325610e0` | **identical** | **identical** |
| `GitCurator.exe` (GUI) | `99cc22e6810f6881a19dc5bd6f702d9d8e653a4a1eae7994a705e95bbd002422` | **identical** | **identical** |
| `GitCurator-v0.92.0-go.1-windows.zip` | `61b4cf2adc2eec96c70ae7962ac5bd89f70ac4d644eedefdb0b5b02cda527ff4` | **identical** | — |

**AC-22 GREEN**: the two-exe zip rebuilds byte-identical — warm×2 AND a
cold-GOCACHE leg per exe (the fresh-machine story behind "reproduce from its
tag"). The ritual's build commands:

```bash
# the CLI exe (doc 10 §2's command, the windows cross-compile):
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 go build -trimpath \
    -ldflags "-X main.version=$(cat VERSION)" -o bin/GitCurator-cli.exe ./cmd/gitcurator-cli

# the GUI exe (the NESTED webui module; CGO-free — the WebView2 loader is
# pure Go; no webkit tree needed for the windows leg):
cd webui && CGO_ENABLED=0 wails build -platform windows/amd64 -trimpath \
    -ldflags "-X main.version=$(cat ../VERSION)"

# the zip (the full composition + the guards):
go run ./build/mkzip -version "$(cat VERSION)" -exe … -gui-exe …
```

The zip: **13 files, 56,908,881 bytes in, 27,226,166 bytes zipped, sha256
`61b4cf2adc2eec96c70ae7962ac5bd89f70ac4d644eedefdb0b5b02cda527ff4`** —
`GitCurator.exe` (27,693,056) + `GitCurator-cli.exe` (29,193,216) + the six
launchers + `config.example.json` + `WINDOWS-QUICKSTART.md` + the stamped
`VERSION` (`0.92.0-go.1`). (T-65's CLI-milestone zip was 12 files /
15,988,399 bytes — the GUI exe is the thirteenth member now.)

## 2. THE FINDS (byte-level, each settled before the fix)

**THE -X-UNDER-TRIMPATH DROP (new law, pinned by test + the print).** The
version stamp did NOT land in the first `-trimpath` builds: the linker DROPS
an unreferenced package-level var's `-X` bytes under `-trimpath` (isolated
empirically: unused var + trimpath → the value is absent from the exe;
referenced var + trimpath → present; without trimpath → present either way).
The fix is the reference's own shape: `gui/app.py:30-38` prints
`[main] LOADED version …` to stderr at GUI startup — the Go twin
(`webui/main.go`) prints the stamped version the same way, which references
the var and keeps the bytes alive. mkzip's version-stamp guard then asserts
the stamp's byte-presence in BOTH exes (a forgot-the-flag build exits 2).
Proof: `b"0.92.0-go.1" in exe` → `True` post-fix; the pre-fix exe carried
only the manifest's winres-normalized `version="0.92.0.1"` string.

**THE DEFAULT-ICON EXE (the icon find).** `webui/build/appicon.png` was
byte-identical to the wails v2.16.0 TEMPLATE default (sha256
`28332a3b130d1970d208e40b7ab2c84f9ead10199f8c058e08e07fe21fe7ea49` — the
packager's `generateIcoFile` regenerates `icon.ico` from it), so every exe
since T-70 shipped the WAILS icon, not the brand. Fixed: `icon.ico` is now
the brand icon (byte-equal to `assets/icons/gitcurator.ico`, 7 PNG-compressed
sizes, 20,664 bytes) + `appicon.png` is the brand artwork's 256px entry
(8,780 bytes, sha256 `e7323fae5385b977ee4e592cf3e8375f712ff1069c6c28c4d6466f27247a402c`).
The exe-side proof: the 256px entry's head AND tail AND the 128px entry's
head are byte-present in the built exe; the exe's sha256 CHANGED with the
icon swap (`84e64b60…` → `a97687fd…`) — the icon rides the `.syso`. The
wails packager accepts the PNG-compressed ICO (`winres.LoadICO`) — the build
links clean.

**The winres version normalization (recorded, cosmetic).** The manifest's
`assemblyIdentity version="{{.Info.ProductVersion}}.0"` (= `0.92.0-go.1.0`)
is parsed by winres into the dotted quad `0.92.0.1` inside the exe's manifest
resource — a cosmetic normalization in wails' own pipeline. The HUMAN-visible
version resource carries the full stamp: `ProductVersion "0.92.0-go.1"` +
`ProductName "GitCurator"` + the copyright + comments (UTF-16 byte-probes
all `True` in the built exe). The wails.json `Info.productVersion` is pinned
to the repo `VERSION` file by test (the one-version-truth law) + the ritual's
preflight refuses drift.

## 3. THE MANIFEST LAWS (doc 10 §5, pinned by TestManifestLaws)

`webui/build/windows/wails.exe.manifest` now carries, beside the wails
template's own DPI section (`dpiAwareness` `permonitorv2,permonitor`):

- **supportedOS Win10/11**: `{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}` (the
  compatibility section);
- **longPathAware**: `true` (the optional law, taken — vault paths under
  long trees stay openable).

## 4. THE STANDING ASSIGNMENT (the rider that rode the same slice)

The 📤 **Export Backup ZIP** More-menu row + the Backup page's **Restore**
button — both born honestly-disabled in Path B — are LIVE doors now
(`internal/app/exportdoors.go` + the shell doors + the frontend; the wiring
pinned by tests, THE BINDING-IMPORT LAW honored, the wailsjs bindings
regenerated content-only at the 755 baseline). The rider fix that rode with
them: **the v0.86.0 deferred-ledger rider** now lands on BOTH zips (backup +
export) as `GitCurator-data/sync_review.db` at the ZIP ROOT — the reference's
own arcname (`zf.write(_ledger, 'GitCurator-data/sync_review.db')`) —
retiring T-71's placeholder-era note (the syncreview ledger is real since
Phase 3/4; `TestBackupRiderLanded` + `TestRunVaultExportLaws` pin it).

## 5. The honest boundary

- The **live Windows runtime legs** (unzip → launch → the clean-VM checklist,
  doc 10 §7) are the owner's Phase-7 items (T-82) — this sandbox has no
  Windows box (the standing constraint since SESSION-001). The cross-compile
  legs above are the sandbox's proof.
- The release automation (tag → build → attach) is **T-81** — `build/
  release.sh` is its reusable core; no tag, no GitHub release was cut this
  session (per-session: commit + push + reports only).
- The `-s -w` AV question (doc 10 §2's [verify]): the wails production
  default includes `-w -s` (the exe rides it); the AV scan verdict is T-82's
  checklist item on the owner's side.

## 6. Reproduce-from-tag recipe (the release-notes text)

From a clean checkout of this commit, with go ≥ 1.24 + the wails CLI v2.16.0
on PATH:

```bash
./build/release.sh          # builds both exes + the zip twice + cold legs,
                            # proves the sha256s, prints the release numbers
```

The expected numbers: the table in §1 (the zip's sha256 is the release-notes
record — AC-23's shape).
