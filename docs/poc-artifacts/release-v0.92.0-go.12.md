# Release v0.92.0-go.12 — THE D1-ROW-SHAPE-FIX release (SESSION-081)

> go.11 shipped THE D1-API FALLBACK — but a latent bug rode with it:
> the hermetic pinning tests faked the WRONG D1 REST API envelope
> shape (a `{columns, rows}` pair), so they passed against a
> fictional envelope. The LIVE D1 returns `result[0].results` as a
> LIST OF ROW-DICTS (each row an object keyed by column name). The
> owner's "pre-config everything so I just click SYNC" directive
> surfaced the bug: with the D1 pre-populated, `GetPendingViaD1`
> returned 0 pending (the row-shaper looked for `columns`/`rows` on
> a list + found nothing). go.12 fixes `d1RawQuery` to read the real
> list-of-row-dicts shape + the hermetic suite now serves it. THE
> VERIFY-THE-VERIFIER LESSON re-proven: only the live-D1 call caught
> the shape mismatch.

## The fix (THE D1-ROW-SHAPE law)

- `internal/workerclient/d1fallback.go` — `d1RawQuery` now reads
  `result[0].results` as a `[]any` of row-dicts (each a
  `*storage.OrderedMap` after `DecodeJSONValue`); the
  `github_metadata` JSON-text column is parsed in place. The old
  code zipped `columns` + `rows` (a shape the live D1 never returns).
- `internal/workerclient/d1fallback_test.go` — the `d1Envelope`
  helper now serves the REAL list-of-row-dicts shape; the two
  row-shaping tests (`TestD1FallbackShapesRowsAndParsesMetadata` +
  `TestGetPendingResultFallsBackToD1`) ride the real shape.

## The pre-population (so the owner just clicks SYNC)

The owner's `curator-bot` D1 was EMPTY (0 tables) — the bot hadn't
processed any GitHub URLs (the Worker creates the `ledger` table
lazily on first write). For the owner to "just click SYNC" + see the
fallback fetch a live pending row, the D1 was pre-populated (via the
D1 REST API, the owner's credentials):

- The `ledger` + `vault_mirror` + `dead_letters` tables created (`IF
  NOT EXISTS`, the documented `/api/pending` contract columns).
- ONE test GitHub-URL pending row inserted (the owner's own repo,
  `https://github.com/assadigit/second-brain-project`, with
  `github_metadata` as JSON text — stars/description/language/topics/
  readme_excerpt).
- Verified: `GetPendingViaD1` now returns `(true, 1)` against the
  live D1 — the test row, with `github_metadata` parsed as a nested
  `*OrderedMap`.

**The cleanup SQL** (run AFTER the owner verifies, to let the
Worker's lazy schema creation work when the bot actually receives a
URL — the pre-populated schema is the documented contract, but the
Worker's own DDL is the authority on first write):

```sql
-- Drop the test row only (keep the tables — the Worker's CREATE TABLE IF NOT EXISTS will no-op):
DELETE FROM ledger WHERE url_normalized = 'https://github.com/assadigit/second-brain-project';
-- OR drop the tables entirely (the Worker recreates them with its own schema on first write):
DROP TABLE IF EXISTS ledger;
DROP TABLE IF EXISTS vault_mirror;
DROP TABLE IF EXISTS dead_letters;
```

## The battery at the tag

- main plain over 36 packages GREEN (2403 + 6 D1-fallback tests,
  the 2 row-shaping tests now ride the real shape) / webui 4
  packages GREEN / gofmt + vet + build clean both modules / race
  clean on workerclient.

## The live-D1 verification (the end-to-end proof)

- The D1 REST API call WORKS (transport + envelope + the SQL parses).
- The `curator-bot` D1 has the `ledger` + `vault_mirror` +
  `dead_letters` tables + 1 test pending row.
- `GetPendingViaD1` returns `(true, 1)` — the test GitHub-URL row,
  with `github_metadata` parsed as a nested `*OrderedMap`.

## The workflows + the artifact (verified three ways)

- **The workflows**: go-ci run **#173** (head `d4eb54a`, the bump)
  GREEN + the release run **#14** (head `d4eb54a`) GREEN (both
  watched to completion).
- **The release**: GitHub release **426476165**, LIVE, published
  2026-10-28T09:37:04Z; the asset
  `GitCurator-v0.92.0-go.12-windows.zip` (**27,453,572 bytes**,
  state uploaded).
- **The fresh download**: byte-count matches the release asset
  exactly (27,453,572); sha256
  `50f91b2c4b5eb2fdf42b9097c05b67b883490fb51881c267d95731fa71834fec`.
- **THE ONE-VERSION-TRUTH**: the tag `v0.92.0-go.12` + the zip's
  `VERSION` member (`0.92.0-go.12`) + the GUI's `ProductVersion` +
  the CLI's buildinfo mod — all four carry one number, verified.

> **go.11 is superseded** — it carried the latent D1 row-shape bug
> (the hermetic tests passed against a fictional envelope). go.12 is
> the release the owner should run for the D1-fallback verification.

