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
