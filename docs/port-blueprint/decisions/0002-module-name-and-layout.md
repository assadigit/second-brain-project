# ADR-0002 — Go module name, repository placement, and version policy

- **Status**: accepted
- **Date**: blueprint session (2026-10)

## Context

The repository is public at `https://github.com/assadigit/GitCurator` (README.md:14-15), currently
containing the Python app under `app/`, the Cloudflare Worker under `app/cloudflare-bot/`, docs at
root and under `docs/`. The Go port must coexist with the Python app during the transition — the
Python app stays the released product until the Go app reaches parity (see roadmap).

## Problem

Pick a Go module path and directory layout that (a) matches the repo's public URL, (b) does not
disturb the existing tree, (c) reads clearly next to `app/`, and (d) keeps a sane package structure.

## Options considered

1. Module at repo root (`module github.com/assadigit/GitCurator`) — would put `go.mod` at the root
   next to Python docs/config; risks tool confusion and pollutes the root.
2. Module under `app/` — mixes Go sources into the Python app's directory.
3. **Module under a new top-level `go-desktop/` directory.**

## Decision

- New top-level directory **`go-desktop/`** in the repository, with:

  ```
  go.mod → module github.com/assadigit/GitCurator/go-desktop
  ```

- Rationale: preserves the public-URL-derived path (Go module paths are case-sensitive and must match
  the repo path when fetched over HTTP), keeps the Go tree separate from `app/`, and the name says
  what it is. The existing Python release zip pipeline (`app/gitcurator/tools/build_zip.py`) is
  untouched.
- **Go version policy**: use the latest stable Go toolchain available at implementation time;
  minimum supported version **Go 1.22**. Pin `go 1.22` (or newer) in `go.mod`; do not use unstable
  stdlib features. Record the exact toolchain in the build docs (see `10-build-release-packaging.md`).
- **Package naming**: lowercase single words (`links`, `notebuilder`, `vaultscan`, …) mirroring the
  Python module names where practical to ease traceability (see `05-python-to-go-mapping.md`).

## Consequences

- The future CI can run Go tests and Python tests side by side without interference.
- If the project later wants the Go app to *become* the repo root deliverable, moving the module is a
  one-line `go.mod` change plus path rewrites (`gofmt -r` / `gofmt` after `go mod edit -module`).

## Evidence

- Repo remote: `git remote -v` → `https://github.com/assadigit/GitCurator` (verified in session)
- Existing layout constraints: `Port-investigation.md` §3 (repository safety — do not restructure the
  existing tree in *this* session; the next session adds only the new directory).

## Amendment 1 — Implementation home moves to a new private repo (2026-10, owner decision)

The owner created **`assadigit/second-brain-project`** (private, id 1420105183; codename until a
branded name is chosen) as the Go implementation's home.

- **Module path**: `module github.com/assadigit/second-brain-project`, repo-root layout
  (`cmd/`, `internal/`, `webui/`, `docs/`, `assets/`). The `go-desktop/` subdirectory plan is
  superseded; the package tree in `04-target-go-architecture.md` is otherwise unchanged.
- **GitCurator (this repo) remains the Python live-product repo** — Python releases, the Worker,
  and the Python zip ritual continue here until Go GUI parity (roadmap Phase 6). Nothing Go is
  added to this repository.
- **Differential testing**: the unmodified Python reference stays a pinned clone of
  `assadigit/GitCurator`; the harness records the exact git rev per run (at blueprint time
  v0.91.0 / `8fd4179`).
- **The blueprint is mirrored into the new repo** at `docs/port-blueprint/` (implementation-
  canonical copy; this copy remains the audit-of-record snapshot). `Port-investigation.md` is
  mirrored alongside it.
- **Releases during development** attach to the private repo (owner-only). The branded/public
  decision (rename, make public, stay private) is deferred to Phase 7/8.
- **A8 is retired**: no Go tree inside GitCurator → zero interaction with `build_zip.py`.
- A module-path rename at branding time is mechanical (`go mod edit -module` + import rewrites) —
  acceptable pre-launch.

## Amendment 2 — The go directive rides tls-client's floor (2026-10-16, SESSION-012, T-31)

**Status: applied with an owner-acknowledgement request** (the ADR-0006
Am.1 precedent: proceed with the newer owner-acknowledged pin, record
the correction, request acknowledgement).

The version policy above says "Pin `go 1.22` (or newer) in `go.mod`".
T-31 (internal/webfetch) imports `bogdanfinn/tls-client v1.16.0` — the
impersonation door's owner-acknowledged pin (ADR-0006 Am.1+2) — into
the main module, and that pin's own go.mod declares `go 1.24.1`, so
the main module's directive MUST be ≥1.24.x (Go's module-graph rule;
verified empirically: `go build` refuses, `go mod tidy` rewrites).

- **Applied**: `go 1.24.1` in the main go.mod (the pin's own
  declaration — the conservative graph-minimum choice).
- **Unchanged**: `modernc.org/sqlite v1.36.1` exactly; our own source
  keeps the 1.22 language floor; the build toolchain stays go1.27.2.
- **Side effects** (all forced by the authorized pin's graph):
  `golang.org/x/net v0.48.0` and `golang.org/x/text v0.32.0` become
  direct requires (x/net/proxy is ADR-0006 §1's named SOCKS5 dialer;
  x/text carries the charset ladder), `golang.org/x/sys` bumps
  v0.30.0 → v0.39.0, plus tls-client's transitive set (fhttp, utls
  forks, quic-go-utls, brotli, circl, … — all recorded in go.sum).
- **The rollback path if the owner declines**: downgrade the tls-client
  pin to a ≤go-1.22 era version (needs a re-PoC of the fingerprint
  substitution) or drop the in-process door (contradicts ADR-0006 §4).
  One line from the owner resolves it.

## Amendment 3 — The Wails build-time toolchain + the webui nested module (2026-10-19, SESSION-033, T-70)

**Status: applied — surfaced BEFORE the first `wails` invocation per the session
mandate ("surface the pinned-version + frontend-step + CGO-story record
BEFORE adding anything").**

Phase 6's entry (T-70, the PoC C GUI-shell gate) needs the Wails toolchain.
The record:

- **The wails CLI (build-time only)**:
  `GOBIN=/home/z/.local/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`
  with the session toolchain go1.27.2 (installed into the toolchain tree next
  to go itself — rebuilt per wipe from this record; the sha256-verified
  go1.27.2 tarball is TOOLCHAIN.md Am.2's). v2.17.0 exists on the proxy; the
  pin deliberately stays **v2.16.0** — ADR-0001 Am.2's
  confirmed-by-working-PoC version (the same stability-over-novelty law as
  every other pin in this repo). The CLI is a build tool; it is NEVER a
  runtime dependency of the main module.
- **The runtime dependency story (the authorized-four law holds
  byte-level)**: the PoC shell lives at `webui/` as a NESTED Go module
  (`module github.com/assadigit/second-brain-project/webui`) whose go.mod
  requires `github.com/wailsapp/wails/v2 v2.16.0` (+ its transitive closure —
  authorized by ADR-0001, the accepted framework decision, and Am.2's pin)
  with `replace github.com/assadigit/second-brain-project => ../`. The MAIN
  module's go.mod stays EXACTLY the authorized four (modernc.org/sqlite
  v1.36.1, bogdanfinn/tls-client v1.16.0, golang.org/x/net v0.48.0,
  golang.org/x/text v0.32.0) — unchanged this session. Go's module-graph rule
  forces the webui module's `go` directive to ride wails v2.16.0's own floor
  (`go 1.25.0`, verified from the module proxy's published .mod) — the
  tls-client/1.24.1 precedent (Am.2) applied again at the module boundary.
- **The internal-package boundary + the interim seam**: Go forbids the nested
  webui module from importing the main module's `internal/` packages, so the
  PoC's real job is exposed through `guibridge/` — a NON-internal main-module
  package shaped like doc 04 §2's event seam (LogLine/Progress callbacks +
  one command method). It RETIRES at T-72, when the production shell lands:
  webui/ dissolves into the main module and webui/bindings.go imports
  internal/app directly — the wails require THEN enters the main go.mod via
  its own amendment (the recorded next decision; deliberately not taken now).
- **The frontend step (the PoC: none)**: `webui/frontend/` is vanilla
  HTML/CSS/JS served as-is (wails.json's `frontend:install`/`frontend:build`
  are empty strings — the wails CLI skips the frontend build). The
  SESSION-001 PoC had used npm+vite; the re-gate goes vanilla so the build
  stays hermetic across the wipe-rebuild cycle (no node_modules fetch per
  wipe; node exists in the sandbox but is NOT part of the build). The
  vanilla-vs-vite choice for the PRODUCTION frontend (T-72+) stays OPEN for
  the owner — both work under wails v2.
- **The CGO story**: Linux dev legs need CGO + webkit2gtk-4.1 + gtk3 via
  SESSION-001's rootless recipe (apt --download-only + dpkg -x →
  ~/.local/lib-wails, the .pc prefix rewrites, the runtime LD_PRELOAD shim
  for WebKit's compile-time libexec path, WEBKIT_DISABLE_COMPOSITING_MODE=1
  under Xvfb — the full recipe: docs/poc-artifacts/t-70-wails-poc.md). The
  WINDOWS production exe is CGO-FREE: wails' WebView2 loader is pure Go, so
  `wails build -platform windows/amd64` cross-compiles from this Linux box
  (proven in SESSION-001; re-proven this session).
- **The nested module and the battery**: `go test ./...` at the repo root
  does not descend into nested modules — the main battery stays the main
  battery; webui's own tests run from webui/ (recorded in the session
  report).
