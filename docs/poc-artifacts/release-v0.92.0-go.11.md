# Release v0.92.0-go.11 — THE D1-API-FALLBACK release (SESSION-081)

> The owner's ISP blocks workers.dev at the IP level — even the DoH
> tier can't reach a black-holed anycast. THE D1-API FALLBACK closes
> the last gap: when the direct Worker request fails, the honest poll
> falls back to GetPendingViaD1(), querying the D1 database directly
> via api.cloudflare.com (a DIFFERENT edge the ISP does NOT block).
> Read-only: no ack — the ack rides the direct path when the network
> recovers. Cut under the standing fix-cycle directive, riding the
> proven release ritual verbatim (T-80/T-81: the bump → the tag on
> the bump commit → the workflows green → the release verified).

## The law (THE D1-API FALLBACK)

`internal/workerclient/d1fallback.go` (NEW) — `GetPendingViaD1()`
queries the D1 database directly via the Cloudflare REST API
(api.cloudflare.com/client/v4/accounts/{account_id}/d1/database/
{database_id}/query) when the direct Worker request fails. The SQL
mirrors the Worker's `handleGetPending` exactly (the documented
`/api/pending` contract): `ledger LEFT JOIN vault_mirror`,
`forgotten=0`, mirror status `NULL/'pending'` (the items the desktop
hasn't acknowledged yet), not decommissioned, not a blocked-domain
dead letter, `ORDER BY first_seen_at ASC`, `LIMIT 100`. The
`github_metadata` column rides as JSON text in D1; the fallback
parses it back to the nested object the `/api/pending` response
carries. Read-only: no ack, no `sync_state.desktop_last_poll` side
effect — that's the Worker's own leg.

THE REACH-CLOUDFLARE LAW (carried): the D1 API call rides
`directClient` — the SAME DoH-fallback dialer the sync channel uses
— so a poisoned local resolver can't kill this leg either. THE
SAME-DECODER LAW: the envelope rides `storage.DecodeJSONValue` (the
same decoder `MakeRequest` uses for the Worker response), so integral
columns surface as int64 — the `/api/pending` shape the downstream
consumers already handle.

## The wiring (the fallback + the UI + the config)

- `internal/workerclient/endpoints.go` — `GetPendingResult()` now
  falls back to `GetPendingViaD1()` after the direct request fails.
  A new `d1FallbackConfig()` method reads the 3 config keys; a new
  `d1FallbackMissing()` method names WHICH key is missing (the
  diagnostic law).
- `internal/workerclient/config.go` — `ConfigKeys` gains
  `cloudflare_api_token` + `cloudflare_account_id` +
  `cloudflare_d1_database_id`.
- `internal/constants/constants.go` — `ExampleTree()` gains the 3
  keys (the defaults contract; the fresh-version load path fills
  empty for owners who haven't paired the fallback — the fallback
  stays dormant).
- `internal/app/settingsb.go` — `DirectPairValues` struct +
  `DirectPair` method extended with `APIToken` + `AccountID` +
  `DatabaseID` fields.
- `webui/frontend/settings.js` — the Direct Pair UI has 3 new
  OPTIONAL fields (API Token + Account ID + Database ID) with the
  explanatory info-header.
- `webui/frontend/wailsjs/go/models.ts` — `DirectPairValues` class
  updated.

## The pinning tests (the hermetic suite — every law pinned)

`internal/workerclient/d1fallback_test.go` (NEW) — six hermetic
pinning tests (a loopback httptest server fakes the D1 REST API; the
`d1APIBase` seam swaps to it; no network):

1. **THE DISABLED-WHEN-CREDS-MISSING LAW** — any of the 3 creds
   empty → `(false, nil)`; NO HTTP leg runs (the law's own guard).
2. **THE DIAGNOSTIC LAW** — `d1FallbackMissing` names WHICH key is
   empty (the owner sees exactly what to fix).
3. **THE ROW-SHAPING LAW** — columns→dict + `github_metadata`
   parsed back to a nested object; integers ride as int64.
4. **THE ENVELOPE-FAILURE LAW** — `success:false` at the envelope
   level → `(false, nil)`.
5. **THE QUERY-FAILURE LAW** — `result[0].success:false` (a SQL
   error) → `(false, nil)`.
6. **THE FALLBACK-WIRING LAW** (SESSION-081 itself) — when the
   direct Worker request fails (the worker_url is dead),
   `GetPendingResult` falls back to `GetPendingViaD1` and returns
   `("ok", list)`.

## The goldens (updated for the 3 new keys)

- `internal/constants/testdata/config_example.golden.json` (sorted)
  + `internal/config/testdata/config_example_ordered.golden.json`
  (insertion order) — the 3 new keys added.
- `internal/config/testdata/config_corpus.golden.json` — the
  `example_keys` section gains the 3 new keys (in insertion order).

## The live-D1 verification (SESSION-081's own evidence)

A throwaway build-tagged test (`d1live`, deleted before the commit;
credentials via ENV vars so no secret bytes touched git) exercised
`GetPendingViaD1` against the owner's LIVE D1 database. Findings:

- The Cloudflare REST API call **works**: transport OK, envelope
  decode OK, the SQL **parses** (D1 returned "no such table: ledger"
  — a schema-state issue, NOT a code bug).
- The account has exactly ONE D1 database — `curator-bot` — and it
  is **EMPTY** (0 tables). The `database_id` is correct (the right
  `curator-bot` DB), but the schema hasn't been created yet — the bot hasn't received/processed any GitHub URLs
  (the Worker creates the ledger table lazily on first write). The
  fallback will fetch live the moment the bot processes a URL and the
  `ledger` table exists.
- The code's transport + envelope + SQL-parse are all verified
  against the live D1; the column-name match is per the documented
  `/api/pending` contract (the Worker source is not in this repo).
  If the Worker's actual column names diverge from the spec, the
  owner surfaces it after the bot processes a URL + the fallback
  runs — the SQL is a one-line edit at that seat.

## The battery at the tag

- main plain over 36 packages GREEN (2403 + 6 new D1-fallback tests)
  / webui 4 packages GREEN / gofmt + vet + build clean both modules /
  race clean on the workerclient package (the new code's seat).

## The workflows + the artifact (verified three ways)

- **The workflows**: go-ci run **#170** (head `f87d7fc`, the bump)
  GREEN + the release run **#13** (head `f87d7fc`) GREEN (both
  watched to completion).
- **The release**: GitHub release **426430005**, LIVE, published
  2026-10-28T08:46:09Z; the asset
  `GitCurator-v0.92.0-go.11-windows.zip` (**27,452,357 bytes**,
  state uploaded).
- **The fresh download**: re-downloaded once; byte-count matches the
  release asset exactly (27,452,357); sha256
  `2b3b82ac3dab3427c11de62168c26895793f39e43a28b7ab47e98297c1a13f88`.
- **The zip**: 14 files (0-START-HERE.txt first in the folder, both
  exes, VERSION, the launchers, the optional git+WebView2 checker).
- **THE ONE-VERSION-TRUTH** (the release-record law): the tag
  `v0.92.0-go.11` + the zip's `GitCurator/VERSION` member (13 bytes,
  `0.92.0-go.11\n`) + the GUI's `ProductVersion` resource + the
  CLI's buildinfo mod — all four carry one number, verified.

