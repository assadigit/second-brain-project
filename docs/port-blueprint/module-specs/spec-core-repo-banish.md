# Module Spec — `app/gitcurator/core/repo_banish.py`

- **LOC**: 659 · **Go home**: `internal/repobanish` · **Status**: port
- **Responsibility**: the GitHub-vault twin of banishment — detect repo notes marked for deletion
  (same grammar, filtered to `github.com/` URLs), confirm over the injected Telegram channel,
  enforce DB-first, then move files and write `DECOMMISSIONED.md` rows.

## Symbols (`repo_banish.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `REPO_BANISH_GATE_PREFIX "🗑️ Repo vault scan:"` / `REPO_BANISH_TALLY_PREFIX` | 94-96 | spoken prefixes |
| `scan_pending_repo_banishments(vault, note_state_db, log, read_cache)` | 110-209 | reuses website_pipeline scans; canonical = `normalize_url`; already-dismissed URLs never re-asked |
| `banish_marked_repos(cache, note_state_db, vault, log, items)` | 337-462 | per canonical: `_banish_repo_db` (274-334: note_state dismiss [skipped under dry-run]; `cache.reset_dead_links` FIRST (a leftover quarantine row would hand the reset door the power to un-do); `cache.settle_repos`; `cache.resolve_failed_urls`; `cache.forget_processed_url`) → `_wp._banish_note_file` move → record rows (`source='repo note tag'/'repo Trash folder'/'repo table gesture'`, status `🗑️ banished — confirmed <date>`) |
| `revive_banished_repos` | 475-526 | ♻ rows → clear dismissed + unsettle |
| `run_repo_banishment_pass(cache, note_state_db, vault, log, config, confirm_factory)` | 529-659 | revive (ungated) → scan → count → ask → verdict gates burial (auto/confirmed only) → tally (zero run answers too) |

Never raises; idempotent (`.trash` never walked; dismissed never re-asked); hand-written marks
kept + warned.

## Notes

- Private-API coupling to `website_pipeline._banish_note_file` / `_parse_review_frontmatter` —
  in Go these become shared helpers in `internal/websitepipeline` (exported).
- DB-first ordering is behavioral (the 404-reset trap) — preserve exactly.

## Tests

Port `test_repobanish` (46) + `test_decommission` selections: detection doors, gate verdicts,
DB-first ordering, revive flow, dry-run laws, twin hygiene.

## Shipped record — T-46, SESSION-021 (2026-10-17)

**Landed** (`internal/repobanish/pass.go` — the completion; the gate's
eyes + the burial + the note-finder landed with T-45 in
repobanish.go): `ReviveBanishedRepos` (THE ♻️ DOOR BACK — every
♻-marked repo row in the vault's own DECOMMISSIONED.md clears the
note_state dismissal [dry-run gated, exactly as the reference gates
it] and un-settles the ledger row [riding the given cache — the
shadow cache in a rehearsal]; idempotent; the "fetched like new"
line), and `RunRepoBanishmentPass` (the confirmation-gate
orchestration: the ungated revive FIRST, the detection scan, the count
SPOKEN before the ask, the ask over the INJECTED channel factory, the
verdict gates — 'auto'/'confirmed' bury, 'declined'/'timeout'/'defer'
keep with the honest lines — THE TALLY with the zero-run answer and
the why map), plus `RepoBanishGatePrefix` and the channel types
(`RepoBanishConfirmResult` / `RepoBanishChannelFn` /
`RepoBanishConfirmFactory`) and the config `banish_confirm` opt-out
door (str() through helpers.PyStr — `0` fires, `0.0` does not).

**Tests** (40 new module tests; 1464 race-clean at the close): the
ported test_repobanish selections (`repobanish_test.go` — the gate's
eyes [the mark grammar incl. the body hashtag + the frontmatter key +
the block style + the trash-folder move, the trailing-slash twin, the
dismissed gate, the .trash invisibility, the table gestures incl. the
confirmed-history and ♻️-revival and websites-row doors], the burial
[every never-fetch gate, the handwritten-kept sacred law, the
table-gesture doors incl. the DB-only close, the leftover-quarantine
subsumption, the dry-run rehearsal, idempotence], THE ♻️ DOOR BACK
[clears dismissal + settlement, idempotent, websites rows ignored],
the pass [the quiet clean noop, the count spoken before the ask, the
verdict gates, the config opt-out, the ungated revival under a
decline, the broken channel / broken scan honesty], the never-fetch
promise [the batch loop's own loads, the queue-door ladder through a
fresh batch.VaultIndex, the revive-then-rebanish full cycle], and the
no-op-settle never-locks-the-other-connection regression pin — two
connections over one cache.db, the reference's own test_repobanish
find). The three CacheDB helper legs (processed_row_for
found/normalized/missing, forget true-then-false) already live in
internal/store's suites (T-24); the source-contract legs (the GUI
worker's wiring, the hermetic-law source pin, build_note's teaching
line) are the Python repo's own CI facts — the note-teaching line
rides notebuilder's port.

**Go-reality notes** (the SESSION-021 amendments):

- **The bare-string channel door type-narrowed**: the reference's
  channel answers a dict `{'verdict': …}` OR a bare string; Go's
  `RepoBanishChannelFn` returns only the result struct — the same
  narrowing the websites twin (T-40) made (the Telegram round-trip's
  own contract is the dict shape). A raising channel maps to a nil
  return; a panicking one is survived through recover (the house
  pattern).
- **The factory's raise maps to an error**: `banish_confirm_factory`
  builds the channel or raises — Go's `RepoBanishConfirmFactory`
  returns `(nil, error)`, the warning line + the defer unchanged.
- **The revive's table iteration order**: the reference iterates the
  table dict in insertion (row) order; Go iterates sorted keys — the
  revived count, the DB effects and the per-row log lines are
  order-independent (multi-row log order is the only surface, and no
  reference test pins it); the same deterministic-order choice
  T-45's ScanPendingRepoBanishments made.
- **The dry-run asymmetry is preserved exactly**: only the note_state
  dismissal clear is dry-run gated in `ReviveBanishedRepos`; the
  unsettle rides the given cache (the caller passes the shadow cache
  in a rehearsal) — the reference's own gating.
