# Release v0.92.0-go.13 — THE D1-FALLBACK-DIAGNOSTIC release (SESSION-081)

> go.12 fixed the D1 row-shape bug, but the owner's go.12 re-run STILL
> hit "❌ Bot queue check failed: the Worker queue poll failed (network
> down or the Worker answered an error)" — in **<1 second**. The <1s
> timing was the key evidence: if the D1 fallback had RUN, the D1 API
> call would take a few seconds. The <1s failure meant the fallback
> was **SKIPPED** — the 3 Cloudflare-API creds were missing on the
> owner's machine (the fallback short-circuited at the disabled-when-
> creds-missing law; `GetPendingResult` returned `("error", nil)`
> with NO indication the fallback was even attempted). go.13 closes
> the diagnostic gap: the error message now reports whether the
> fallback is ARMED or DORMANT + what to fix.

## The fix (THE D1-FALLBACK-DIAGNOSTIC law)

- `internal/workerclient/d1fallback.go` — a new `D1FallbackStatus()`
  method on the Sync reports the fallback's arming state (a pure
  config read, no HTTP):
  - **DORMANT** — names WHICH of the 3 creds is missing + tells the
    owner to open Settings → Direct Pair + fill the 3 Cloudflare-API
    fields (API Token + Account ID + Database ID) to arm it.
  - **ARMED** — the 3 creds are set; the direct Worker request failed
    + the fallback should have run; if the error persists, the D1 API
    call itself failed (verify the token has D1 read access + the
    account_id + database_id are correct).
- `internal/app/commands.go` — `SyncFetchQueue`'s `"error"` branch
  (the GUI SYNC path the owner hit) now appends `D1FallbackStatus()`
  to the error message. Never a bare "queue poll failed" again.
- `internal/workerclient/d1fallback_test.go` — a hermetic test pins
  the arming law (DORMANT names the 3 missing keys; ARMED when all
  set; partial → DORMANT names the missing one).

## The owner's path (the root cause of the <1s failure)

The owner had the Worker pairing (so `IsEnabled` was true + the
direct leg ran + failed fast on the ISP block) but NOT the 3 D1 creds
(likely from a go.9-era Direct Pair that pre-dated the D1 fields, OR
the pre-filled reference config wasn't dropped in). The fallback was
skipped → `("error", nil)` → the generic error. go.13's diagnostic
now TELLS the owner this.

## The owner's fix (works with go.12 OR go.13)

1. Open **Settings → Direct Pair**.
2. Fill the 3 new optional Cloudflare-API fields (API Token + Account
   ID + Database ID — the values are in the reference config at
   `/home/z/my-project/public/gitcurator-config.json`, downloadable
   via the Preview Panel).
3. Click **Direct Pair** (saves the 3 creds to config.json).
4. Click **SYNC** — the direct leg fails (ISP blocks workers.dev) →
   the D1 fallback fetches the test GitHub-URL pending row live.

## The battery at the tag

- main plain over 36 packages GREEN (2403 + 7 D1-fallback tests, the
  new D1FallbackStatus arming test) / webui 4 packages GREEN / gofmt
  + vet + build clean both modules / race clean on workerclient.

## The workflows + the artifact (appended after the live verification)

> The release-record law: the workflow run IDs + the sha256 + the
> GitHub release ID are appended here after the go-ci + release
> workflows complete + the live asset is re-downloaded fresh.
