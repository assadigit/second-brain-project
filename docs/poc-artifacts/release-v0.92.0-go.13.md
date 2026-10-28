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

## The workflows + the artifact (verified three ways)

- **The workflows**: go-ci run **#175** (head `8ffc9dc`, the bump)
  GREEN + the release run **#15** (head `8ffc9dc`) GREEN.
- **The release**: GitHub release **426518397**, LIVE, published
  2026-10-28T10:20:42Z; the asset
  `GitCurator-v0.92.0-go.13-windows.zip` (**27,454,969 bytes**,
  state uploaded).
- **The fresh download**: byte-count matches (27,454,969); sha256
  `b4b8a4f9f56a3dee8340c1e4ceb5fac008704d37a79904bfddc6ee941d60cf4d`.
- **THE ONE-VERSION-TRUTH**: the tag `v0.92.0-go.13` + the zip's
  `VERSION` member + the GUI's `ProductVersion` + the CLI's
  buildinfo mod — all carry one number, verified.

> **go.11 + go.12 are superseded** — go.11 carried the latent D1
> row-shape bug; go.12 fixed it but the error message didn't tell the
> owner the fallback was DORMANT. go.13 is the release the owner
> runs — the error message now self-diagnoses.

