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
