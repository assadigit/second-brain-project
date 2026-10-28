# Release v0.92.0-go.14 — THE d1creds.json PRE-CONFIG DROP-IN release (SESSION-081)

> go.13's diagnostic told the owner the D1-API fallback was DORMANT
> (the 3 Cloudflare-API creds were missing), but the owner wanted
> **zero manual field-entry**. go.14 closes that gap: the app now
> reads an OPTIONAL `d1creds.json` file (next to `config.json`) on
> startup + merges the 3 Cloudflare-API creds into the live config.
> The owner downloads ONE small file (just the 3 creds) + drops it in
> the GitCurator folder; the app arms the fallback automatically. No
> Direct Pair field entry. The drop-in MERGES (never overwrites the
> owner's other settings). The no-secrets law holds: the creds ride
> the file the owner provides, never the public release.

## The fix (THE PRE-CONFIG DROP-IN law)

- `internal/config/config.go` — a new `applyD1CredsDropin(config,
  configPath)` reads `d1creds.json` (next to `config.json`) on
  startup + merges the 3 Cloudflare-API creds
  (`cloudflare_api_token` + `cloudflare_account_id` +
  `cloudflare_d1_database_id`) into the live config when present +
  non-empty. Called from `LoadAppConfig` after `ApplyGlobalCredentials`.
  No-op when the file is absent (the common case). Only non-empty
  trimmed values merge (a broken or partial drop-in is ignored, never
  poisons the config). The OTHER config keys are untouched (the
  drop-in MERGES, not overwrites).
- `internal/workerclient/d1fallback.go` — the `D1FallbackStatus()`
  DORMANT message now mentions BOTH arming paths: (a) the Direct Pair
  UI, OR (b) the `d1creds.json` drop-in.
- `internal/config/config_test.go` — a hermetic test pins every law:
  absent → no-op; full → all 3 merge; the other config keys untouched;
  partial → only present creds; empty-value → doesn't overwrite;
  unparseable → no-op.

## The owner's path (the 2-step zero-field-entry flow)

1. **Download `d1creds.json`** — from the Preview Panel at
   `/d1creds.json` (the file is served by the Next.js dev server +
   gitignored from the Next.js repo; it carries just the 3 creds).
2. **Drop it in the GitCurator folder** (next to `GitCurator.exe`,
   where `config.json` lives).
3. **Click SYNC** — the app reads `d1creds.json` on startup + arms
   the fallback; the direct Worker leg fails (ISP blocks workers.dev)
   → the D1 fallback queries the D1 → fetches the test GitHub-URL
   pending row live. ✅

(The Direct Pair UI path still works too — fill the 3 fields +
click Direct Pair. The drop-in is the zero-field-entry alternative.)

## The battery at the tag

- main plain over 36 packages GREEN (2403 + 8 D1-fallback/drop-in
  tests) / webui 4 packages GREEN / gofmt + vet + build clean both
  modules / race clean on workerclient.

## The workflows + the artifact (appended after the live verification)

> The release-record law: the workflow run IDs + the sha256 + the
> GitHub release ID are appended here after the go-ci + release
> workflows complete + the live asset is re-downloaded fresh.
