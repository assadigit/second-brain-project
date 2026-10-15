# ADR-0009 — Git operations stay CLI-subprocess, never a library

- **Status**: accepted
- **Date**: blueprint session (2026-10)

## Context

VaultSeal (`integrations/vaultseal.py`) and GoodRepos (`integrations/goodrepos.py`) drive **the git
CLI as a subprocess** with a precise command ladder:

- `git init -b main` (fallback `git init` + `symbolic-ref HEAD refs/heads/main`)
- `git add -A -- .`, `status --porcelain`, `rev-parse`, `commit -m "seal: …"`,
- divergence ladder: plain push → fetch → `merge-base --is-ancestor` → rebase →
  `merge --allow-unrelated-histories` → rescue branch `seal-rescue/<ts>` (main is **never**
  force-pushed),
- token passed **only** inside one-time push/fetch URLs (`https://x-access-token:<tok>@github.com/…`),
  never in `.git/config`, `FETCH_HEAD` sanitized after every fetch
  (`vaultseal.py:606-623, 643-713`),
- environment scrubbing (`GIT_TERMINAL_PROMPT=0`, strip `GIT_DIR`/`GIT_WORK_TREE`/…) and identity
  `GitCurator VaultSeal <vaultseal@users.noreply.github.com>` (`vaultseal.py:731-738`).

## Problem

Choose between shelling out to git (same as Python) vs a Go git library (go-git).

## Options considered

| Option | Assessment |
|---|---|
| **A. git CLI subprocess (status quo)** | Identical behavior & fingerprints; requires git installed (already required today — the Python app shells out to git everywhere); trivial parity |
| **B. go-git (pure Go)** | No external git needed, but: rebase/merge-unrelated-histories support is weak/partial, performance on big repos is worse, and byte-level behavioral parity with git CLI (pack negotiation, FETCH_HEAD handling) is unproven |
| **C. libgit2/git2go** | CGO — conflicts with the no-CGO build policy (ADR-0003/0006) |

## Decision

**Keep the git CLI subprocess approach.** Port `vaultseal`/`goodrepos`/`repo_radar`'s git usage as
`exec.Command("git", …)` sequences with the exact same arguments, env scrubbing, redaction, and the
same output parsing (`--porcelain` lines, `%h%x00%s%x00%aI` log format). Windows requirement: git
for Windows on PATH — same as today (documented in `10-build-release-packaging.md` runtime deps).

## Consequences

- Behavioral parity is near-mechanical; the divergence ladder and token laws port verbatim.
- Runtime dependency on git.exe remains (acceptable: the current release already requires it; the
  quickstart documents it).
- Windows quirks to replicate: `CREATE_NO_WINDOW`-equivalent (`SysProcAttr{HideWindow: true}`),
  killing hung git children, `.git/FETCH_HEAD` rewriting.

## Evidence

- Full ladder & token law: investigation report 1-f §D.1; `integrations/vaultseal.py:528-641`
- Env scrubbing: `integrations/goodrepos.py:461-468`
- Stuck-subprocess killing precedent: `integrations/subprocess_runner.py` (kill + wait(10))
