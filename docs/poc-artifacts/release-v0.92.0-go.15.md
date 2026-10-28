# Release v0.92.0-go.15 — THE DRIFT-SEAL READY-MADE AUTO-RECOVERY release (SESSION-081)

> go.14's D1 fallback WORKED (the queue fetched + settled in 8s), but
> the drift-seal then failed with "git commit failed: On branch main
> Changes not staged for commit" — the pre-existing vault-repo
> git-state issue, unmasked because the sync now reaches the
> drift-seal step (before, the queue check failed first). The owner
> wanted it ready-made: "I just want to click sync." go.15 delivers:
> the drift-seal now AUTO-RECOVERS from a stale `.git/index.lock`
> (the common Windows class) + surfaces the real git-add error for
> other failures — no manual `del .git\index.lock` needed.

## The root cause (the swallowed `git add` return code)

`internal/vaultseal/seal.go`'s `sealInternal` ran `git add -A -- .`
(line 54) but **ignored its return code**. A stale `.git/index.lock`
(the common Windows class — a previously-crashed git process leaves
the lock) made `git add` fail silently, so `git commit` saw nothing
staged + failed with the confusing "Changes not staged for commit"
message. The owner had no idea the lock was the cause.

## The fix (THE READY-MADE AUTO-RECOVERY law)

- `internal/vaultseal/seal.go` — a new `stageAll()` helper runs
  `git add -A -- .` + AUTO-RECOVERS from a stale `.git/index.lock`:
  on a `git add` failure, if the error mentions `index.lock` OR the
  lock file exists, the helper deletes the lock + retries the stage
  (safe because at drift-seal time the app is the only writer, so the
  lock is stale). The real `git add` error surfaces when the stage
  fails even after the recovery (never the confusing "Changes not
  staged" message). `sealInternal` now calls `stageAll()` + surfaces
  its error instead of swallowing it.
- `internal/vaultseal/laws_test.go` — a hermetic test pins the law:
  a clean stage succeeds with no recovery; a stale lock is
  auto-cleared + retried + the recovery warning logged + the file
  staged.

## The owner's path (just click SYNC)

1. Download `v0.92.0-go.15` + extract (copy your `config.json` +
   `d1creds.json` from the go.14 folder).
2. Click **SYNC**. If a stale `.git/index.lock` blocks the
   drift-seal's stage, it's auto-cleared + retried → the vault-seal
   commits + pushes. No manual `del .git\index.lock`.

## The battery at the tag

- main plain over 36 packages GREEN (2403 + 9 D1-fallback/drop-in/
  drift-seal tests) / webui 4 packages GREEN / gofmt + vet + build
  clean both modules / race clean on vaultseal + workerclient.

## The workflows + the artifact (appended after the live verification)

> The release-record law: the workflow run IDs + the sha256 + the
> GitHub release ID are appended here after the go-ci + release
> workflows complete + the live asset is re-downloaded fresh.
