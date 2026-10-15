# 10 — Build, Release, and Packaging (Windows-first)

## 1. Build toolchain

- Go ≥ 1.22 (latest stable at implementation; record the exact version in the release notes).
- No CGO anywhere (ADRs 0003/0006 keep the build toolchain-free):
  `CGO_ENABLED=0` for all Windows builds.
- Wails CLI pinned version for the GUI build (`wails build`); the frontend is a plain
  `npm/pnpm` build step.

## 2. Reproducible build commands (to validate in implementation)

```bash
# CLI (console subsystem)
GOFLAGS=-trimpath CGO_ENABLED=0 go build -o bin/GitCurator-cli.exe ./cmd/gitcurator-cli

# GUI (no console window)
GOFLAGS=-trimpath CGO_ENABLED=0 go build -trimpath \
  -ldflags "-H=windowsgui -s -w -X main.version=$(cat ../VERSION)" \
  -o bin/GitCurator.exe ./cmd/gitcurator

# Package (deterministic zip — port of tools/build_zip.py laws)
go run ./build/mkzip -version $(cat ../VERSION) -out reports/dist/
```

**[verify]** exact ldflags/Wails interplay; **[verify]** `-s -w` vs AV-heuristics tradeoff
(stripped binaries sometimes score *worse* with antivirus scanners — keep symbols if flagged).

Determinism laws ported from `tools/build_zip.py:126-217`: fixed entry timestamps
(2026-09-29 00:00:00), sorted arcnames, `0o644` external attrs, DEFLATE level 9, tracked-files
only, banned-secrets guard (refuse to package `config.json`, `*.session`, `cache.db`, `.env*`),
launcher flattening to zip root, `.bat` CRLF + ASCII-only assertion, VERSION stamped from the repo
file.

## 3. Windows architecture and subsystem

- Target: `windows/amd64` (64-bit; the owner's Windows 10). `windows/arm64` is a future option —
  the architecture keeps it feasible (no CGO, no native deps beyond WebView2/git).
- GUI subsystem: `-H=windowsgui` (no console flash). CLI ships as its own console exe.
  A single exe with a `--cli` flag can flip subsystems only per-build; two exes is the
  honest solution (documented in ADR-0010).

## 4. Runtime requirements (document, don't hide)

| Component | Needed by | Mitigation |
|---|---|---|
| **WebView2 Runtime** | GUI (Wails) | Preinstalled on updated Win10/11; the installer/`1-INSTALL` equivalent checks and bootstraps the evergreen installer **[verify]** the Wails bootstrapper path |
| **git.exe on PATH** | VaultSeal, GoodRepos, radar, mirror | Same requirement as the Python release; `1-INSTALL.bat` equivalent checks `git --version` |
| Local Chrome (optional) | fifth door | Detected at the standard paths (`hand_delivery.find_chrome`) |
| Local LLM server (optional) | Ollama/llama.cpp | Auto-detect/auto-start preserved |
| Telethon-free | — | gotd embedded; no Python anywhere |

## 5. Assets, icons, versioning

- `go:embed` the prompts, taxonomy defaults, Inter font, wav sounds, Lucide SVGs; icon set from
  `app/assets/` + a Windows `.ico` for the exe (generate from the logo — **[verify]** source
  asset; the repo has `public/logo.svg` at the web-dashboard level; an app icon must be chosen —
  open question OQ-9).
- Version: read repo `VERSION`; `-X` stamp into `main.version`; UA string uses it (doc 08 §6);
  `--version` flag prints it. Milestones: `0.92.0-go.1` … until parity, then continue semver.
- App manifest: DPI-aware (`PerMonitorV2`), supported OS Win10; long-path aware optional.

## 6. Installation vs portable

- **Portable zip first** (owner habit; no installer risk). Layout:
  `GitCurator/GitCurator.exe`, `GitCurator-cli.exe`, `.bat` launchers (ASCII, CRLF, walk-up cd
  removed — exes sit at root), `WINDOWS-QUICKSTART.md`, `config.example.json`, `VERSION`.
- Optional `1-INSTALL.bat` (checks WebView2 + git; prints next steps). No registry writes, no
  Program Files install in this phase.

## 7. Clean-machine validation checklist (release gate)

1. Fresh Windows 10 VM (no dev tools): unzip → `1-INSTALL.bat` → launch GUI.
2. Cold start ≤ 1.5 s to first frame; window 1000×750; theme correct.
3. CLI: `GitCurator-cli.exe --status` completes ≤ 0.5 s with the config card.
4. Dry-run: `GitCurator-DRY-RUN.bat` equivalent completes; dry-run report written; **zero** vault
   writes (verify with a snapshot diff).
5. Test Connection modal against the owner's real services (or fakes in the VM).
6. A real 5-link batch against a temp vault; seals against a temp GitHub repo; publish against the
   staging Worker door (or the loopback rig).
7. Verify: no console window on GUI launch; sounds play; dark/light toggle; log flood (10k lines)
   stays interactive.
8. AV check: upload the unsigned exe to a scanner service; if heavily flagged, document
   mitigations ( signing cert is the owner's call — cost/benefit noted).
9. sha256 of the zip recorded in the release notes; rebuild reproducibility spot-check.

## 8. Release automation

- GitHub Actions job (new workflow file in `go-desktop/.github/` or an addition under a new path —
  the existing `ci.yml` is untouched per safety rules; the Go workflow is a new file):
  tag push → build both exes → mkzip → attach `GitCurator-v<version>-windows.zip` to the GitHub
  release (the owner's law: a release without its zip is not done).
- Future platforms: linux (exe + same zip), macOS (needs notarization; later phase) — architecture
  permits; explicitly out of scope now.

## 9. Antivirus false-positive notes

Go exes occasionally trip heuristic scanners (packed strings). Mitigations, in order: keep
build simple (`-trimpath`, no obfuscation), avoid `-ldflags -s -w` if it worsens scores,
consider an EV/OV code-signing cert later (owner decision — cost), document the sha256 in
release notes so the owner can verify downloads.
