# second-brain-project (codename)

> **Private workspace for the Go port of GitCurator** — the personal second-brain library
> builder (Python/PyQt6 → Go). Codename until the branded name is chosen; stays private
> until launch.

**Status: Phase 0 CLOSED** — all three PoC verdicts final (Wails ✅ · tls-client ✅ ·
gotd/td ✅ proven on the owner's account, then **demoted to dormant by the owner's
Option-B stamp**: the queue rides the Worker only — ADR-0004 Am.3); harness skeleton +
Python baselines archived; SESSION-001/002 closed. **Phase 1 CLOSED** (SESSION-003–005):
T-10…T-17 ALL DONE with generated-golden parity — the characterization suites
(test_core incl. prompts/usermemory, taxonomy, note formats), the import-grammar
differential, and the byte-stable config round-trip are green. **Phase 2 CLOSED
(SESSION-006–010; all six tasks)**: T-20 `internal/store` (modernc.org/sqlite
v1.36.1 pinned to the go 1.22 floor; the ADR-0003 DSN pragmas; **ADR-0003 Am.1's
CAST read law**); T-21 `internal/notestate` (the §4.4 moves-as-corrections engine;
+ the SESSION-008 CRLF fingerprint fidelity fix); T-22 `internal/notereads` (v0.71.0
THE ONE-OPEN LAW — the second scan of unchanged notes opens NOTHING:
reference-counted opens 4→0, db_hits 6); **T-23 `internal/websitepipeline/state`
(SESSION-010) — WebsiteStateDB: THE ONE SPELLING boundary, the 2×attempts backoff,
both settlements, the v0.64.0 healing pass; T-24 `internal/store/repos.go` —
CacheDB: the retry queue's no-stacking/normalized-key laws, THE SETTLED REPOS
LEDGER + the ♻️ revive door, the fail-closed record_404, the open-time migrations
(and the CAST read law GENUINELY LIVE on the TIMESTAMP-declared columns);
T-25 `internal/batch` — LinkTracker (the 5-phase manifest with FULL ACCOUNTING +
the json.dump(indent=2) byte-parity writer) + VaultIndex (the reconciliation
ground truth: skip sets, the 4000-char read, the site-map gates). SESSION-009's
T-23/T-24 were lost at an environment wipe while push-pending — re-ported
independently in SESSION-010, which now pushes after every task. 425 tests
race-clean.**
**Phase 3 OPEN (SESSION-011–012)**: T-30 `internal/llmclient` landed
(commit `f2073b5`) — the three providers on Python-json.dumps-identical
wire bytes (the payload goldens), the JSON-mode memoized 400-fallback,
extract_json, the llama.cpp detection engine + the /props grab, the
autodetect policy; and **T-31 `internal/webfetch` landed (commit
`4d93877`) — the whole fetch ladder: the routes (SOCKS5 rdns, the
http-proxy absolute URI, loopback/direct never proxied), the OWN
redirect loop (hop cookies, 308, the Referer law), the Chrome-131
header set, the rate limiter with penalties, the caps + capped
decompression, the failure taxonomy, the DoH verdicts, the dead-page
rescue + the Wayback courtesy, and the tls-client impersonation door
(ships armed); 676 tests race-clean; the thirteenth golden generator;
the go directive rides tls-client's floor (1.24.1 — ADR-0002 Am.2
records the correction; sqlite stays v1.36.1)**; and **T-32
`internal/webextract` landed (commit `833b21b`; SESSION-013) — the
HTML extraction port on the HAND-ROLLED CPython-3.12 html.parser
tokenizer (option b: the owner's x/net/html verdict never arrived, so
zero new imports — the swap stays one ADR away), the _PageParser
laws, the js-shell/paywall heuristics, the charset ladder with four
GENERATED data tables embedded (entities, single-byte charsets,
cp949, DBCS acceptance bitmaps — from the reference runtime's own
stdlib, sha-pinned by the golden); 713 tests race-clean; the
fourteenth golden generator; the goldens caught five real divergences
before landing**. Next: T-33 connectioncheck + the sentinel, then
T-34 workerclient (the gate's last stone).

## Map

| What | Where |
|---|---|
| **HOW SESSIONS WORK (start here)** | [`docs/SESSION-PROTOCOL.md`](docs/SESSION-PROTOCOL.md) — the loop every session follows |
| **The state card / next prompt** | [`docs/NEXT-SESSION.md`](docs/NEXT-SESSION.md) — rewritten at every session close |
| **Session reports (the history)** | [`docs/sessions/`](docs/sessions/) — numbered SESSION-NNN.md files |
| The complete migration blueprint (73 docs) | [`docs/port-blueprint/`](docs/port-blueprint/) — start at its `README.md`, then `00-executive-summary.md` |
| The mission spec that produced the blueprint | [`docs/Port-investigation.md`](docs/Port-investigation.md) |
| App icon (ready-made brand mark, multi-size) | [`assets/icons/gitcurator.ico`](assets/icons/gitcurator.ico) — swap the file later to rebrand |
| Owner decisions (all Q&A rounds) | `docs/port-blueprint/14-risks-assumptions-open-questions.md` §6 |

> **Starting a new work session?** Open a fresh AI chat and paste the prompt from
> `docs/NEXT-SESSION.md` §5. That's the whole workflow.

## The goal in one paragraph

A single Windows executable (Go engine + Wails GUI) that behaves byte-identically to the
Python app, opens the owner's real data in place, and fixes the sluggishness: cold start
≤ 1.5 s, vault scan ≤ 50% of Python, idle memory ≤ 200 MB. Delivered as both a
`setup.exe` (silent WebView2 bootstrap when missing) and a deterministic portable zip.
The unmodified Python reference repo (`assadigit/GitCurator`, pinned rev — at blueprint
time v0.91.0 / `8fd4179`) stays the live product until Go GUI parity, and is never
modified by this migration; differential tests run both apps side by side.

## Laws (binding, from the blueprint)

- Reality wins over specs: when source code contradicts a spec, fix the spec in the same commit.
- No visual redesign during the port — replicate the documented design (ADR-0001).
- Preserve observable behavior, including quirks; only the five pre-classified fixes are allowed.
- Never commit credentials, session files, or the owner's real `config.json` / `cache.db`.
- Tests are never weakened, skipped, or deleted to get green.
