# ADR-0007 — Migration order: core + CLI first, GUI second

- **Status**: accepted
- **Date**: blueprint session (2026-10)

## Context

The Python app already has a **CLI twin** of the GUI: `python main.py --cli --auto --yes` runs the
exact same pipeline (same config.json, same cache.db, same vaults) through a `QCoreApplication`
(`app/gitcurator/cli/cli_run.py:239-504`). The CLI surface (`--init`, `--sync`, `--status`,
`--test-connection`, `--detect-llm`, `--auto`, `--import-file`, `--dry-run`, …) covers the *entire*
functional core: queue fetch, both pipelines, seals, publish, retries, quarantine, dry-run.

The mission requires a "reliable functional foundation" before any UI work, and forbids a UI redesign
in this phase.

## Problem

Sequence the migration so risk is retired early and the GUI never blocks proof of correctness.

## Options considered

1. **GUI-first** — highest-risk surface first; parity unprovable until late. Rejected.
2. **Strangler pattern inside one process** — not applicable (different language runtimes).
3. **Core + CLI first, then GUI** on top of the same command layer.

## Decision

1. **Phase order (see `12-migration-roadmap.md`)**: foundations (links/storage/config/DB) →
   pipelines (GitHub, Websites) → integrations (Telegram, seals, Worker client, publish) →
   **CLI parity release** (a usable Go CLI that the owner can run alongside the Python app;
   golden/differential tests green) → GUI (Wails shell reusing the command layer) → packaging.
2. The Go **command layer** (`internal/app` façade in the target architecture,
   `04-target-go-architecture.md`) exposes every operation the GUI will need (start batch, progress
   events, log events, confirm-gates, scan flow, sync display facts, …) as an event-emitting service
   used by both the CLI and the GUI. This mirrors the Python split where the CLI already drives
   `ProcessingWorker` (`cli/cli_run.py:318-341`).
3. The owner's `.bat` launchers get Go equivalents late in the roadmap (`10-build-release-packaging.md`).
4. The Python app remains the released product until Go GUI parity is demonstrated
   (acceptance criteria in `15-acceptance-criteria.md`).

## Consequences

- Correctness (notes, ledgers, seals, publish) is provable long before any UI exists — matching the
  repo's own test-first culture (3192 tests, AGENTS.md).
- The riskiest external contracts (Telethon→gotd, curl_cffi→tls-client) land inside the CLI phase
  where they are testable headlessly.
- GUI work starts with an already-validated backend; the Wails PoC (ADR-0001) can proceed in
  parallel *after* the CLI milestone, not before.

## Evidence

- CLI drives the same worker: `app/gitcurator/cli/cli_run.py:239-504`, `gui/headless.py:55-244`
- Launchers: `app/gitcurator/cli/*.bat` (walk-up cd, venv discovery — see
  `10-build-release-packaging.md` for the Go equivalent)

## Amendment 1 — Owner Q&A re-scope (2026-10)

The owner revealed (round-1 Q5) that the Python CLI was built **only to escape the sluggish
Python GUI** — "if GO is fast enough, it's not needed at least at this phase."

**Re-scope:**

1. The *engineering order* is unchanged — core + CLI first, GUI second. The CLI remains the
   engine's headless proving ground, the differential-harness driver, and the emergency
   fallback channel. This is internal discipline, not a product.
2. **The owner-facing CLI milestone is CUT**: no polished terminal paint, no `.bat` launcher
   parity work, no tagged CLI release (old Phase-5 exit / M5 as a *product* event). The CLI
   surface is built only as deep as tests and emergency use require.
3. The **GUI becomes the first owner-visible release milestone** (roadmap Amendment 1 in
   `12-migration-roadmap.md`); the tag + dual-artifact ritual moves there.
4. Python app remains the released product until Go GUI parity (unchanged).

## Amendment 2 — Owner re-instatement of the CLI-facing legs (2026-10-19, SESSION-031)

Am.1 cut the owner-facing CLI polish (terminal paint, `.bat` launcher parity,
the tagged CLI release). The practice has since superseded that cut, on the
owner's own instruction line, and this amendment records the reversal so the
record stops contradicting itself:

1. **The paint work was never skipped**: SESSION-027 landed the full paint box
   (terminal.go — the C ANSI system with NO_COLOR/TTY/--no-color, level styles,
   term_width, mask/rule, cli_print, banner, render_bar, StatusLine, the
   65001 hardening) because the differential golden itself needs the paint
   shapes byte-compared (colors ON and OFF); the "polish" is load-bearing test
   surface, not decoration.
2. **The owner's SESSION-027..031 line re-instated the launcher + release
   legs**: the task list kept T-61 (terminal paint) / T-63 (the launchers) /
   T-64 (the differential corpus run) / T-65 (the `v0.92.0-go.1` tag + zip)
   as Phase 5's remaining books; `docs/NEXT-SESSION.md` v31 (rewritten at
   SESSION-030's close) names T-63 the natural next task; the SESSION-031
   kickoff (the owner's instruction) commands T-61's close-out audit + T-63
   the launchers explicitly.
3. **The CLI remains the emergency-fallback + proving-ground channel**
   (Am.1's engineering rationale stands); what changed back is the *product*
   question — the owner wants the launchers and the parity tag after all
   (the emergency-fallback channel is only real if a non-engineer can
   double-click it).
4. The GUI (Phase 6) stays the first *GUI* milestone; the dual-artifact ritual
   there is unaffected.
