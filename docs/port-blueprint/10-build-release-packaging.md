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

> **Dated note (2026-10-20, T-80/SESSION-048 — the [verify]s settled +
> the GUI command's reality):** (a) the ldflags/Wails interplay is settled:
> `wails build -platform windows/amd64 -trimpath -ldflags "-X
> main.version=$(cat VERSION)"` over the NESTED webui module is the GUI
> command (the spec's `./cmd/gitcurator` path predates the T-72
> nested-module decision; the stub stays, the wails build is the real
> door) — CGO-free, no webkit tree needed for the windows leg. (b) The
> production build carries `-w -s` (wails' own production default); the AV
> verdict is the owner's clean-VM checklist item (T-82). (c) EMPIRICAL LAW
> (pinned): the linker DROPS an unreferenced var's `-X` bytes under
> `-trimpath` — the version var is referenced by the reference's own boot
> print (app.py:30-38's `[main] LOADED version …`), and mkzip's
> version-stamp guard asserts the bytes in both exes. The ritual itself:
> `build/release.sh` (both exes + the zip ×2 + a cold-cache leg; the
> evidence: docs/poc-artifacts/t-80-reproducible-packaging.md).

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

> **Dated note (2026-10-22, SESSION-063 — the §4 [verify] on the
> WebView2 row RESOLVED; the recorded "bootstrapper leg" debt's
> preparable half landed).** The row's promise — "the installer/
> `1-INSTALL` equivalent checks and bootstraps the evergreen installer
> [verify] the Wails bootstrapper path" — is now the shipped shape,
> split into its two honest halves:
>
> 1. **The installer half (LANDED, this session)**: `1-INSTALL.bat`
>    gained the WebView2 leg — a READ-ONLY registry probe of the
>    Evergreen runtime's own key (per-user `HKCU\…\Clients\{GUID}` then
>    per-machine `HKLM\…\WOW6432Node\…\Clients\{GUID}`, value `pv`),
>    GUID `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}` settled at byte
>    level against Microsoft's distribution docs (the t-70 poc plan's
>    recorded GUID was a corruption — corrected the same session). The
>    found leg prints the version; the absent leg points honestly (the
>    desktop app offers the one-click download at its first launch; or
>    install now from the official page). The checker still MUTATES
>    nothing (reg query is the one allowed reg verb; reg add/reg
>    delete/setx/msiexec/powershell all stay banned). Pinned BOTH
>    sides: `TestInstallBatChecksGitAndWebview2` (the self-side laws +
>    the GUID guard — the corrupted shape mutated in fails exactly,
>    reverts clean) and the golden differential's 1-INSTALL goOnly
>    contract (`internal/cli/goldencli_test.go`).
> 2. **The runtime half (SETTLED at source level; the LIVE leg is the
>    owner's, riding T-82 item 1)**: the built GUI exe carries wails
>    v2.16.0's DEFAULT strategy — `download` (byte-settled in the
>    pinned CLI's source: `cmd/wails/flags/build.go:71`; the strategy
>    compiles through `internal/wv2installer/download.go`'s build
>    tags): at first launch, when the runtime is missing, the exe asks
>    (one confirm dialog), then fetches Microsoft's Evergreen
>    bootstrapper (`go.microsoft.com/fwlink/p/?LinkId=2124703`) and
>    installs it — the QUICKSTART's "one internet connection, then
>    never again". No ritual change needed (the flag equals the
>    default under the pin; the proven release command stands). The
>    live verification on a real WebView2-less machine — ADR-0001
>    Am.1's scenario, the owner's machine — remains T-82 item 1's
>    check ("installer checks WebView2 + git and bootstraps/points as
>    designed" — now true by design).

## 5. Assets, icons, versioning

- `go:embed` the prompts, taxonomy defaults, Inter font, wav sounds, Lucide SVGs; icon set from
  `app/assets/` + a Windows `.ico` for the exe (generate from the logo — **[verify]** source
  asset; the repo has `public/logo.svg` at the web-dashboard level; an app icon must be chosen —
  open question OQ-9).
  *Resolved at T-80: the source asset is the seed commit's brand icon
  (`assets/icons/gitcurator.ico`); `webui/build/windows/icon.ico` is now
  byte-equal to it (the T-70-era file was wails' regeneration of the DEFAULT
  template appicon — the find + the proof:
  docs/poc-artifacts/t-80-reproducible-packaging.md §2), and appicon.png is
  the brand artwork's 256px entry. OQ-9's choosing is done (the seed's own
  brand icon); a future re-brand is a deliberate commit + pin update.*
- Version: read repo `VERSION`; `-X` stamp into `main.version`; UA string uses it (doc 08 §6);
  `--version` flag prints it. Milestones: `0.92.0-go.1` … until parity, then continue semver.
  *Reality note (2026-10-20, T-80): the UA half of this line does NOT ride the
  stamp — the reference's own Worker UA is HARDCODED at `GitCurator/0.91.0`
  (cloudflare_sync.py:59, with the "keep in sync with VERSION at release
  time" comment) and the Go twin hardcodes the same string (the cross-binary
  law: the Worker must see both binaries speak identically). The stamp's
  homes: the CLI's `--version`, the GUI's boot print, the exe's version
  resource (wails.json Info, pinned to VERSION by test), and the zip's
  stamped VERSION member.*
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

> **Dated note (2026-10-20, T-81/SESSION-049 — the workflow LANDED at
> `.github/workflows/release.yml`):** the tag-only door (`v*`), the
> toolchain pins (go 1.27.2 + the wails CLI v2.16.0, cross-checked by
> test), the ritual wrap (the job runs `build/release.sh` — the single
> owner of the build commands), THE TAG LAW (a real release tag must
> equal `v$(cat VERSION)`; the reserved dry-run prefix `v0.0.0-` rides
> with the assertion waived → a draft+prerelease `[DRY-RUN]` release),
> and the AC-23 notes (the parity table + the zip numbers + the
> reproduce recipe, asserted-consistent before the attach). The
> dry-run's runner legs are BLOCKED by the account's Actions billing
> (free-plan minutes exhausted — the refusal predates T-81: every recent
> `go-ci` run failed the same way; run `39512685975` awaits the owner's
> billing fix + re-run; the recipe:
> `docs/poc-artifacts/t-81-release-workflow.md` §3). **THE TAG-STATE
> BUILD LAW (T-81's find)**: go1.24+ embeds the VCS-derived module
> version in buildinfo — a tagged checkout and an untagged one of the
> same commit produce DIFFERENT exe bytes (the release numbers live at
> the tagged checkout; the wails-built GUI exe is exempt — `(devel)`);
> byte-settled in the evidence doc §2.2.

> **Dated note (2026-10-20, later the same day — SESSION-050: the
> dry-run COMPLETED; T-81's acceptance GREEN).** The owner's fix for
> the billing block: **the repo went PUBLIC** ("Make it public, so you
> can have unlimited CI/CD" — public repos get free unlimited Actions
> minutes; a full-history secret scan ran before the flip: 1,415 blobs,
> zero token-shaped strings). The re-run of `39512685975` (attempt 2)
> went **green end-to-end on a real runner** (all 8 steps; the ritual's
> own PASS line on the runner), created the `[DRY-RUN]` draft+prerelease
> release with the zip asset, and the **cross-machine sha256 comparison
> proved byte-identical reproduction**: the runner's numbers == a full
> sandbox rebuild at the same tagged tree state (CLI `f2ef8362…`,
> GUI `99cc22e6…`, zip `6409352a…`, 13 files / 56,908,881 in /
> 27,225,658 zipped) — the evidence doc §2.1's comparison target had
> gone stale (the tag moved `b5c9a27`→`0961ad9` after that measurement;
> recorded in the evidence doc §2.1's correction + §5). The refusal leg
> (`v9.9.9-t81-refusal`,
> run `39515838232`) failed EXACTLY at the tag-law step and created NO
> release. The test release + both test tags deleted; only the real
> milestone release `v0.92.0-go.1` remains. The full record: the
> evidence doc §5.

## 9. Antivirus false-positive notes

Go exes occasionally trip heuristic scanners (packed strings). Mitigations, in order: keep
build simple (`-trimpath`, no obfuscation), avoid `-ldflags -s -w` if it worsens scores,
consider an EV/OV code-signing cert later (owner decision — cost), document the sha256 in
release notes so the owner can verify downloads.

## 10. Reference release artifact (v0.91.0 — the parity target for the Go zip ritual)

The Python app's last release (the artifact the Go port must eventually match in discipline):
`GitCurator-v0.91.0-windows.zip` — **157 files, 1,395,452 bytes, sha256
`44834bbaa936a41c955576ff7707cfb78b6fc3d98b31f6ea5a283ce13b15cee4`** — GitHub release 413382822,
tag `v0.91.0` at `a0ccf34`, published 2026-10-15. The Go deterministic-zip builder (§2) uses this
as the working reference: same fixed-timestamp discipline, same tracked-files-only law, double-build
sha256 identity check. (Recorded 2026-10-15 from the release API + the wrap-up deploy record at
`main` `98f4fc7`.)
