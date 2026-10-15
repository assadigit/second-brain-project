# 00 — Executive Summary

> Status: blueprint complete, awaiting implementation. Written from a full read of the repository at
> v0.91.0 (HEAD `8fd4179`, branch `main`, clean tree). Every claim in the blueprint cites
> `path:line` evidence from that commit.

## What the application is

**GitCurator** is a personal "second brain" library builder for a single owner (a designer, not a
programmer) on Windows 10:

1. The owner forwards links (GitHub repos, websites, anything) to a Telegram bot
   (`@githubfetcherbot`) backed by a **Cloudflare Worker** (JS, D1, KV, Queues — already live,
   **not** part of this migration).
2. The **Python desktop app** (PyQt6 GUI + visual CLI) fetches the queue, routes GitHub repos to the
   GitHub pipeline and everything else to the Websites pipeline.
3. An **LLM** (Ollama / llama.cpp / Anthropic / any OpenAI-compatible endpoint) curates each link
   into a structured Obsidian note in one of **three vaults** (GitHub Projects, Websites —
   machine-written; Manual Notes — owner-only with a read-only `Library/` mirror).
4. Every run **seals** the vaults to private GitHub mirrors (git CLI subprocess ladder) and
   optionally **publishes** a public Website Directory to the Worker (which serves the live site).
5. A fifth "door" (real Chrome via raw-socket CDP) rescues pages the machine doors cannot fetch.

The design laws that must survive the port: local-first · private by default · no link left behind ·
the owner's moves are corrections (never damage) · existing notes are never rewritten · every risky
operation has a dry-run.

## Scale (measured, not estimated)

| Artifact | Size |
|---|---|
| Production Python (`app/gitcurator/` + `app/main.py`) | **123 files, 75,605 LOC** |
| Python tests (`app/tests/`) | 85 files, 58,823 LOC, **3,192 test methods** |
| Cloudflare Worker JS (`app/cloudflare-bot/src/`) | 8,243 LOC (+ 4,201 LOC tests, 217 cases) |
| Two Next.js web apps (directory site, bot dashboard) | ~50 TSX/TS source files |
| Release ritual | deterministic Windows zip per version since v0.14.1 |

**This is 3–4× larger than the mission brief's "fewer than 20,000 lines" estimate.** The migration
is a multi-session engineering project, not a weekend port; the roadmap (doc 12) is phased
accordingly and the CLI-first strategy (ADR-0007) exists precisely because of this scale.

## Main architectural findings

- **Layering is already good**: `core/` is pure-stdlib, Qt-free, and dependency-injected
  (`WebsitePipeline` takes `llm_call`, `fetch_fn`, `banish_confirm` as arguments — hermetic by
  construction). The GUI is a shell over the same engine the CLI drives. A Go port can follow the
  same seams (`04-target-go-architecture.md`).
- **The GUI is the biggest and most Qt-bound chunk** (~25.6k LOC, 18 mixins, measured-height
  wordwrap hacks, 350 ms input guards, coalesced log). It cannot be transliterated; it must be
  re-hosted on a GUI framework using the app's unusually complete design-token documentation.
- **Behavior is pinned by tests to an unusual degree** (3,192 hermetic cases; a live loopback
  protocol rig in `test_synclive.py`). This is the parity backbone for the port
  (`09-testing-and-parity.md`).
- **Performance history is documented in-repo**: the app *was* catastrophically slow in specific
  ways (1.7 s per log line past 1,000 entries; 2 s GUI-thread socket freezes; 0.5–5 s queue-check
  freezes) and those were fixed in Python. The remaining sluggishness is dominated by Python
  startup, GIL-bound vault walks/parsing, and Qt re-render costs — all of which Go addresses — but
  LLM/network latency (the 300 s timeouts, per-domain politeness) is irreducible and must not be
  mistaken for a regression (`07-ui-and-performance-assessment.md`).

## Recommended target architecture

- **Go module** `github.com/assadigit/GitCurator/go-desktop` in a new top-level `go-desktop/`
  directory (ADR-0002); pure-Go packages mirroring `core/` (links, notebuilder, vaultscan, …), a
  `store` package for the SQLite ledgers (modernc.org/sqlite, ADR-0003), `integrations` for
  Telegram (gotd/td, ADR-0004) and git (CLI subprocess, ADR-0009), `cloud` for the Worker client,
  one `internal/app` command/event façade shared by CLI and GUI.
- **GUI: Wails (Go + WebView2)** replicating the current layout from the documented token tables —
  no redesign; **Fyne** as fallback (ADR-0001, provisional on a PoC).
- **Migration order: foundations → pipelines → integrations → CLI parity milestone → GUI →
  packaging** (ADR-0007). The Python app stays released until Go parity is demonstrated.

## Major risks (top 7)

1. **GUI replication effort** — the honest unknown; mitigated by CLI-first and the token tables.
2. **Telethon → gotd/td** — auth UX and bot-chat iteration must be proven early (PoC) or the
   fallback (Worker-only queue) triggers a documented behavior change.
3. **Chrome-TLS impersonation replacement** (`curl_cffi` → `tls-client`/utls) — library maturity and
   fingerprint quality need a PoC against real walled sites.
4. **Semantic drift in the "small" stuff** — Python's ordered dicts, `str.lower()` vs Go casing,
   code-point slicing, `round()` half-even, naive-local vs UTC timestamps, `os.walk` ordering. A
   dedicated semantics chapter (`05-python-to-go-mapping.md` §3) and differential tests cover this.
5. **Data compatibility** — the Go app must open the owner's real `cache.db`/`config.json`
   side-by-side with Python (merge laws, lock file, WAL) — ADR-0008.
6. **Chrome CDP fifth door** — hand-rolled WebSocket + DevTools; must be ported faithfully
   (attach/launch, wave engine, `page_is_real` grammar).
7. **Scope creep** — the Worker, dashboards, and any visual redesign are explicitly out of scope
   (ADR-0005; mission §2).

## Decisions needing attention in implementation

See `14-risks-assumptions-open-questions.md` for the full register. The headline open items:
Wails-version pinning + WebView2 on the owner's machine; `tls-client` license/API verification;
gotd login PoC; modernc DSN pragma syntax; JSON HTML-escaping parity; replicate-or-fix calls on
three known Python quirks (`stars` dropped in `note_reads` DB hits, character-count RAM budget,
`os.walk` ordering).

## Reading guide for the next session

Start at `README.md` (ordered reading list), then this file, then `12-migration-roadmap.md` and
`13-implementation-task-list.md`. Module-level truth lives in `module-specs/`; consequential choices
in `decisions/`; the parity plan in `09-testing-and-parity.md`.
