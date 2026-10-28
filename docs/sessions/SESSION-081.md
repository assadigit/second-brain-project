# SESSION-081 — 2026-10-28 — Phase 7 (fix cycle) — THE D1-API-FALLBACK release shipped + verified three ways

## Goal (from NEXT-SESSION.md at start)

The state card (`docs/NEXT-SESSION.md`) was STALE at the v77
(SESSION-077 post-fix decision-tree) form — SESSION-077 through
SESSION-080 happened (releases `go.5`…`go.10`) but none rewrote the
card. The actual session goal rode the owner's prompt verbatim:
**implement THE D1-API FALLBACK** (the owner's ISP blocks workers.dev
at the IP level; when the direct Worker request fails, the honest poll
falls back to a direct D1 query via api.cloudflare.com), then run the
full verify cycle (gofmt + vet + build + test + golden updates),
commit + push + tag `v0.92.0-go.11`, wait for go-ci + the release
workflow, and verify the release is live.

> **Honest note on the "already written" claim:** the prompt said the
> code changes were "already written to the working tree at
> /tmp/second-brain-project." They were NOT — the tree was clean at
> `bbd42ec` (v0.92.0-go.10) when the session opened (the prior
> session's tooling went down before the writes persisted). This
> session IMPLEMENTED the D1-API fallback from the spec the prompt
> carried, then ran the verify/commit/tag/release cycle.

## What was done (task IDs, files, commits — cite hashes)

**Commit `f48160a`** — SESSION-081 the D1-API fallback (11 files,
+627/-8):
- `internal/workerclient/d1fallback.go` (NEW, 235 lines) —
  `GetPendingViaD1()` queries the D1 database directly via the
  Cloudflare REST API (`api.cloudflare.com/client/v4/accounts/{acct}/
  d1/database/{db}/query`) when the direct Worker request fails. The
  SQL `d1PendingSQL` mirrors the Worker's `handleGetPending` exactly
  (the documented `/api/pending` contract): `ledger LEFT JOIN
  vault_mirror`, `forgotten=0`, mirror status `NULL/'pending'`, not
  decommissioned, not a blocked-domain dead letter, `ORDER BY
  first_seen_at ASC`, `LIMIT 100`. Read-only: no ack, no
  `sync_state.desktop_last_poll` side effect. THE REACH-CLOUDFLARE
  LAW carried (rides `directClient` — the same DoH-fallback dialer).
  THE SAME-DECODER LAW (the envelope rides `storage.DecodeJSONValue`;
  integers surface as int64 — the `/api/pending` shape). The
  `d1APIBase` seam is a `var` (the hermetic-suite law); `d1RawQuery`
  is the reusable read-only helper (rides both `GetPendingViaD1` and
  the hermetic suite).
- `internal/workerclient/d1fallback_test.go` (NEW, 314 lines) —
  six hermetic pinning tests (a loopback httptest server fakes the D1
  REST API; the `d1APIBase` seam swaps to it; no network): the
  disabled-when-creds-missing law, the diagnostic law, the row-shaping
  law (columns→dict + `github_metadata` parsed), the envelope-failure
  law, the query-failure law, and THE FALLBACK-WIRING LAW itself
  (`GetPendingResult` falls back to D1 when the direct Worker leg
  fails → `("ok", list)`).
- `internal/workerclient/endpoints.go` — `GetPendingResult()` now
  falls back to `GetPendingViaD1()` after the direct request fails;
  the `d1FallbackConfig()` + `d1FallbackMissing()` methods (the
  diagnostic law) live in `d1fallback.go`.
- `internal/workerclient/config.go` — `ConfigKeys` gains
  `cloudflare_api_token` + `cloudflare_account_id` +
  `cloudflare_d1_database_id`.
- `internal/constants/constants.go` — `ExampleTree()` gains the 3
  keys (the defaults contract; empty by default — the fallback stays
  dormant).
- `internal/app/settingsb.go` — `DirectPairValues` struct +
  `DirectPair` method extended with `APIToken` + `AccountID` +
  `DatabaseID` fields.
- `webui/frontend/settings.js` — the Direct Pair UI has 3 new
  OPTIONAL fields (API Token + Account ID + Database ID) + the
  explanatory info-header (the placeholders are generic fakes — the
  owner's real account_id/database_id redacted from the committed
  placeholders; the pre-existing install_id/shared-secret-prefix
  placeholders from SESSION-079 ride unchanged — a debt item).
- `webui/frontend/wailsjs/go/models.ts` — `DirectPairValues` class
  updated (6 lines, tabs preserved via a Python byte-replace after
  the Edit tool tab-corrupted the whole file — see Blockers).
- 3 golden files updated: `internal/constants/testdata/
  config_example.golden.json` (sorted) + `internal/config/testdata/
  config_example_ordered.golden.json` (insertion order) +
  `internal/config/testdata/config_corpus.golden.json` (the
  `example_keys` section).

**Commit `f87d7fc`** — SESSION-081 release bump (5 files, +140/-10):
- `VERSION` `0.92.0-go.10` → `0.92.0-go.11`
- `webui/wails.json` `Info.productVersion` → `0.92.0-go.11`
  (THE ONE-VERSION-TRUTH)
- `WINDOWS-QUICKSTART.md` live-usage version refs ride the bump
  (lines 1, 24, 28, 155; the historical line 148 stays)
- `README.md` front-page live-release pointer rides the bump (the
  go.11 D1-API-fallback clause; the go.4-go.6 history stays
  historical)
- `docs/poc-artifacts/release-v0.92.0-go.11.md` (NEW) — the release
  report (the law + the wiring + the pinning tests + the live-D1
  verification + the battery; the workflow/sha256/release-ID section
  filled post-verification in the next commit)
- **The tag `v0.92.0-go.11` rides THIS commit** (THE TAG-STATE BUILD
  LAW).

**Commit `8734593`** — SESSION-081 close: the release verified three
ways (docs-only; the report's verification section filled with the
real numbers).

## Gate Evaluation (pass/fail/partial + EVIDENCE)

- **gofmt**: clean (`gofmt -l` empty on all changed packages).
- **go vet**: clean (main + webui).
- **go build**: clean (main + webui).
- **main tests**: 36 packages GREEN (2403 + 6 new D1-fallback tests).
- **webui tests**: 4 packages GREEN.
- **race**: clean on the workerclient package (the new code's seat).
- **go-ci run #170** (head `f87d7fc`): GREEN.
- **release run #13** (head `f87d7fc`): GREEN.
- **The release** `v0.92.0-go.11`: LIVE — GitHub release **426430005**,
  published 2026-10-28T08:46:09Z; the asset
  `GitCurator-v0.92.0-go.11-windows.zip` (**27,452,357 bytes**,
  state uploaded).
- **The fresh download**: byte-count matches (27,452,357); sha256
  `2b3b82ac3dab3427c11de62168c26895793f39e43a28b7ab47e98297c1a13f88`.
- **THE ONE-VERSION-TRUTH**: the tag + the zip's `VERSION` member
  (13 bytes, `0.92.0-go.11\n`) + the GUI's `ProductVersion` + the
  CLI's buildinfo mod — all four carry one number, verified.

**THE LIVE-D1 VERIFICATION (the honest finding):** a throwaway
build-tagged test (`d1live`, deleted before the commit; credentials
via ENV vars so no secret bytes touched git) exercised
`GetPendingViaD1` against the owner's LIVE D1. Findings:
- The Cloudflare REST API call WORKS: transport OK, envelope decode
  OK, the SQL PARSES (D1 returned "no such table: ledger" — a
  schema-state issue, NOT a code bug).
- The account has exactly ONE D1 database — `curator-bot` — and it
  is EMPTY (0 tables). The `database_id` is correct (the right
  `curator-bot` DB), but the schema hasn't been created yet — the
  bot hasn't received/processed any GitHub URLs (the Worker creates
  the ledger table lazily on first write). The fallback fetches live
  the moment the bot processes a URL + the `ledger` table exists.
- The code's transport + envelope + SQL-parse are all verified
  against the live D1; the column-name match is per the documented
  `/api/pending` contract (the Worker source is not in this repo).

**Phase-7 exit evaluation**: UNCHANGED — still NOT GREEN honestly.
The owner's go.4 re-run + the sign-off call remain the exit gate's
last leg. The fix-cycle debt (the live-D1 schema) is a NEW owner
action (the bot must process at least one URL for the fallback's
happy path to be live-verifiable).

## Spec updates made

- `docs/poc-artifacts/release-v0.92.0-go.11.md` (NEW) — the release
  report documenting the D1-API fallback law + the wiring + the six
  pinning tests + the live-D1 verification + the verified-three-ways
  numbers.
- `docs/sessions/SESSION-081.md` (THIS file) — the session report.
- `docs/NEXT-SESSION.md` — rewritten for SESSION-082 (the state card
  v82).
- The two goldens + the corpus `example_keys` section (the 3 new
  config keys).

## Blockers & decisions needed

- **THE LIVE-D1 SCHEMA** (the owner's `curator-bot` D1 is empty — 0
  tables). The D1-API fallback code is verified correct (transport +
  envelope + SQL-parse), but the happy path (fetching live pending
  rows) can't be end-to-end-verified until the bot processes at least
  one GitHub URL (the Worker creates the `ledger` table lazily on
  first write). DECISION NEEDED: the owner either (a) sends a test
  GitHub URL to the bot (the Worker writes the ledger + the
  fallback's happy path goes live), OR (b) accepts the hermetic
  pinning tests as the code's verification (the SQL parses; the
  column names are per the documented contract). The owner's call.
- **THE COLUMN-NAME RISK** — the SQL mirrors the documented
  `/api/pending` contract (the Worker source is not in this repo). If
  the Worker's actual column names diverge from the spec, the owner
  surfaces it after the bot processes a URL + the fallback runs — the
  SQL is a one-line edit at `d1PendingSQL` (the single construction
  site).
- **THE PRE-EXISTING OWNER-DATA PLACEHOLDERS** (SESSION-079's
  install_id + shared-secret-prefix in the Direct Pair UI examples)
  are already in git (since go.9); redacting now has no security
  benefit (the values are in git history forever) and changes UI
  behavior. A future cleanup may genericize them; out of scope here.

## Owner actions requested

- **OWNER ACTION #1 (the live-D1 happy-path verification, optional)**:
  send a test GitHub URL to the bot (e.g. forward a repo link to the
  Telegram bot). The Worker writes the `ledger` row + creates the
  `ledger` table (lazy on first write); then SYNC — the direct leg
  fails (ISP blocks workers.dev) + the D1 fallback fetches the row
  live (the happy path proven end-to-end on the owner's machine). If
  the SQL's column names diverge from the Worker's actual schema,
  surface the exact D1 error + the column list (the SQL is a one-line
  edit).
- **OWNER ACTION #2 (the Direct Pair UI fill)**: open Settings →
  Direct Pair, paste the 3 new optional fields (API Token +
  Account ID + Database ID — the values are in the reference config
  at `/home/z/my-project/public/gitcurator-config.json`), click Direct
  Pair. The fallback is now armed.
- **OWNER ACTION #3 (the go.4 re-run + the sign-off)** — unchanged
  from the v77 card; the Phase-7 exit gate's last leg.
- **OWNER ACTION #4 (the config reference is web-served)** — the
  reference config at `/home/z/my-project/public/gitcurator-config.json`
  is served by the Next.js dev server (the preview panel); treat it
  as ephemeral (delete after copying into the GitCurator install).
  The Next.js `.gitignore` excludes it from commits.

## Handoff

- Next session: **SESSION-082** = the post-D1-fallback session (the
  decision tree in priority order: (a) the owner's go.11 D1-fallback
  re-run results (the happy-path verification after the bot processes
  a URL), (b) the standing go.4 re-run + the Phase-7 sign-off, (c)
  the Phase-8 menu pick — only when the fix cycle closes).
- NEXT-SESSION.md rewritten: **yes** (the state card v82).
- The next prompt: see `docs/NEXT-SESSION.md` §5 (verbatim).
