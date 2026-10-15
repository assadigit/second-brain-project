# ADR-0005 — Scope: the Cloudflare Worker and web apps are NOT migrated

- **Status**: accepted
- **Date**: blueprint session (2026-10)

## Context

The repository contains four deployable artifacts:

1. The **Python desktop app** (`app/`) — the migration target.
2. The **Cloudflare Worker bot** (`app/cloudflare-bot/src/*.js`, 8,243 LOC JS) — live in production
   ("worker 0.40.0 LIVE", AGENTS.md), speaks the Telegram webhook, D1/KV/Queues, the `/api/*`
   desktop contract, the directory publish door.
3. The **Next.js directory site** (`app/cloudflare-bot/directory/`) — the public website (Cloudflare Pages).
4. The **Next.js bot dashboard** (`app/cloudflare-bot/dashboard/`) — bot web dashboard (a parallel
   surface; the Worker itself also serves an inline HTML dashboard).

The mission title says "Python-to-Go migration". The Worker and web apps are **JavaScript**, not
Python.

## Problem

Define the migration boundary so the future agent does not attempt to rewrite the Worker (huge risk,
zero requirement) and does not accidentally break the live system's wire contract.

## Decision

1. **The Go migration covers only the Python desktop application** (`app/gitcurator/`, `app/main.py`,
   its config, prompts, taxonomy, assets, launchers). The Worker and both Next.js apps remain
   untouched and live.
2. The blueprint documents the **full wire contract** the Go desktop must honor (HMAC scheme, 1010
   UA law, endpoints, D1 semantic model, directory publish schema) — see
   `08-data-and-compatibility.md` §Worker contract and module spec `spec-worker-contract.md`
   (contract-only, no port).
3. The Go app must **not** depend on any Worker version newer than the currently expected
   `EXPECTED_WORKER_VERSION = "0.40.0"` (`app/gitcurator/constants.py:313`).
4. Any future Worker change is out of scope and requires its own spec/deploy round per the repo's
   existing process (`app/cloudflare-bot/DEPLOYMENT.md`).

## Consequences

- Parity surface shrinks by ~12,500 LOC (JS + TS web apps) — the migration becomes tractable.
- The `test_synclive`-style live-loopback test rig (a real local HTTP server speaking the Worker
  protocol, `app/tests/test_synclive.py:29-120`) is the model for Go integration tests.
- Risk: contract drift. Mitigation: golden contract tests pin the exact request/response shapes
  (see `09-testing-and-parity.md`).

## Evidence

- Worker live status: AGENTS.md ("worker 0.40.0 stays LIVE (app-only round)"), wrangler.toml.
- Desktop↔Worker contract: `cloud/cloudflare_sync.py` + `app/cloudflare-bot/src/api.js:249-296`
  (investigation reports 1-f/1-g).
