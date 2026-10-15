# ADR-0010 — Packaging: single Windows exe, GUI subsystem, no console window

- **Status**: accepted (shape); exact tooling verified in implementation
- **Date**: blueprint session (2026-10)

## Context

The owner launches the app by double-clicking `.bat` files on Windows 10 and receives a
deterministic zip per release (`tools/build_zip.py`: fixed timestamps, tracked files only, secrets
banned, byte-identical rebuilds; AGENTS.md "a release without its zip is not done"). The Python GUI
keeps a console window behind it for log lines.

## Problem

Package the Go app as a reliable, ordinary-user-launchable Windows executable with minimal setup,
while keeping the release ritual (tag → deterministic artifact → GitHub release asset).

## Decision

1. **Deliverable**: `GitCurator.exe` (amd64), built with:
   - `-ldflags "-H=windowsgui …"` for the GUI build so **no console window** appears;
   - a separate console-subsystem build `GitCurator-cli.exe` for CLI/automation usage (the `.bat`
     wrappers call the CLI exe; a `--cli` flag on the GUI exe is also acceptable if the subsystem
     choice is per-build).
2. **Assets**: embed prompts, taxonomy defaults, sounds, fonts, icons via `go:embed`
   (the Python zip ships `prompts/`, `taxonomy/`, `assets/` — same files, embedded).
   The taxonomy file remains overridable via config `taxonomy_path` for owner edits.
3. **Runtime dependencies to document** (not eliminable): WebView2 runtime (GUI; bootstrappable),
   git.exe on PATH (seals/publish), optional local Chrome (fifth door), optional local
   LLM server. No Python, no CGO DLLs.
4. **Distribution**: keep the zip ritual — a `build` target produces
   `GitCurator-v<version>-windows.zip` containing the exe(s), `.bat` launchers (ASCII-only law,
   same walk-up discovery adapted to the exe), `WINDOWS-QUICKSTART.md`, `config.example.json`.
   Determinism: fixed file timestamps in the zip writer (Go's `archive/zip` accepts explicit
   `ZipInfo`-equivalent metadata); record sha256. The Python zip tool stays the reference until the
   Go artifact replaces it.
5. **Versioning**: continue semver; Go milestones carry `-go.N` pre-release suffixes until parity,
   then continue the normal lineage. VERSION file in repo remains the single source.
6. **Installer vs portable**: portable zip first (matches owner habit); optional
   quick-start `1-INSTALL.bat` equivalent that only checks WebView2/git presence.

## Consequences

- Single-exe GUI with embedded assets ≈ 15–25 MB (Wails + assets) — well within norms.
- Console-subsystem split avoids the classic "GUI app flashes a black window" problem while keeping
  CLI output visible when run from a terminal.
- Antivirus false positives on unsigned Go exes are a known nuisance; mitigations documented
  (signing optional, `go build -trimpath`, reproducible builds) — see
  `10-build-release-packaging.md` §AV.

## Evidence

- Current zip algorithm: `app/gitcurator/tools/build_zip.py:126-217`
- Windows launch contract: `app/1-INSTALL.bat`, `app/GitCurator.bat`, CLI `.bat` launchers
  (walk-up cd, venv discovery, ASCII-only/no-chcp law)
- Mission constraints: `Port-investigation.md` §2 (Windows-first), §10

## Amendment 1 — Owner Q&A (2026-10): dual artifact, bootstrapper chain, icon, signing

- **Two artifacts from the same build** (owner Q5 = "both"):
  1. `GitCurator-v<version>-windows-setup.exe` — proper installer (NSIS or WiX, chosen in
     implementation): Start-menu shortcut, uninstaller, embedded app icon, and a **silent
     WebView2 Evergreen bootstrapper chain** when the runtime is missing (owner accepted the
     one-time silent install; always online at install time — see ADR-0001 Am.1).
  2. `GitCurator-v<version>-windows-portable.zip` — the deterministic zip ritual as specified
     above (remains the byte-reproducible artifact of record). First run performs the WebView2
     check and offers a one-click silent fetch. The exe is directly double-clickable; `.bat`
     wrappers are no longer needed (the exe replaces the venv walk-up dance entirely).
- The GUI exe and CLI exe contents are identical across both artifacts; only packaging differs.
- **App icon**: owner will upload a `.ico` to the repo (path pending; suggested
  `app/assets/icons/gitcurator.ico`) — wired into the exe resource (rsrc/winres), window and
  taskbar icon, installer, Start-menu entry (OQ-9 resolved pending the file).
- **Code signing**: deferred per owner Q8 — unsigned first; sign only if AV false positives
  actually appear (mitigations doc 10 §9).
- **Cross-platform**: v1 remains Windows-only; the build setup keeps macOS/Linux targets feasible
  (Wails — ADR-0001 Am.1) but produces none in this phase.

## Amendment 2 — App icon resolved (2026-10): ready-made brand mark

The owner chose a ready-made icon for now: the repo's own brand mark — the **emerald rounded tile
with the white feather** (the Website Directory's favicon artwork, Lucide's ISC-licensed feather
path — `app/cloudflare-bot/directory/public/favicon.svg`; rendered `icon-512.png`). Converted to a
multi-size Windows `.ico` (16–256 px, 20,664 bytes) living at **`assets/icons/gitcurator.ico` in
the implementation repo** (ADR-0002 Am.1). Swapping in a custom icon later = replacing that one
file and rebuilding. **OQ-9 resolved.**
