# Module Spec — `app/gitcurator/core/vault_hygiene.py`

- **LOC**: 812 · **Go home**: `internal/vaulthygiene` · **Status**: port
- **Responsibility**: stray/empty-notes cleanup for any vault: retires content-free stubs and
  app-owned bodyless notes to `.trash/empty-stubs/`, removes the retired master-index scaffold,
  retires legacy root reports — under the sacred law (never touch anything content-bearing or
  unproven; locks always win).

## Constants (`vault_hygiene.py:`)

`HYGIENE_SKIP_DIRS ('.obsidian','.git','.trash','_missing','_moc','_inbox','attachments',
'Library','__pycache__')` · `STUB_MAX_BYTES 256` · `APP_OWNED_BODYLESS_MAX_BYTES 16_384` ·
`_WALK_READ_MAX_BYTES 1_048_576` · `_PROGRESS_EVERY 25` · `_WIKILINK_RE
r'\[\[([^\[\]<>|]+?)(?:#[^\[\]<>|]*)?(?:\|[^\[\]]*)?\]\]'`.

## Symbols

| Symbol | Lines | Contract |
|---|---|---|
| `is_empty_stub(content)` | 141-163 | no frontmatter; only blanks + ≤1 heading |
| `is_app_owned_bodyless(content)` | 166-208 | frontmatter with `managed_by: gitcurator`, body = blanks + ≤1 heading |
| `is_locked_frontmatter(content)` | 211-257 | truthy `locked`/`category_locked` or folded tag lock/locked |
| `has_note_name_shape(fname)` | 272-285 | stem contains a CATEGORY_KEYS segment (the `<repo>_<Category>_<tag>.md` shape) |
| `normalize_identity(text)` | 288-296 | `re.sub(r'[^a-z0-9]+','',text.lower())` |
| `harvest_link_targets(content)` | 312-334 | wiki-link targets, last path segment, identity-normalized |
| `run_vault_hygiene(vault, log, progress, read_cache)` | 458-745 | ONE walk, ONE bounded read per file (or cache): harvests identities (stem, first H1, frontmatter aliases), link targets, candidates; retirement needs ≥1 fingerprint: app name shape / name-twin (stem ∈ identities) / linked-ghost (stem ∈ linked map, self-links excluded); app-owned bodyless door; locks never retire (stem still vouches); huge files contribute stem identity only; retires to `.trash/empty-stubs/`; removes `_index.md` + `_moc/*.md` with `type: master-index\|moc` + prunes empty `_moc/`; legacy root reports (`_processing_report_*.md`, `processing_summary_*.txt`) → `.trash/retired-reports/` |
| `heal_machine_vaults(config, …)` | 752-796 | runs over both machine vaults |
| `hygiene_total(healed)` | 799-812 | counter sum |

Dead: `collect_real_note_identities` (341-389, superseded — not ported).

## Tests

Port `test_strayheal` (65) + `test_vaulthygiene` selections: each fingerprint (shape/name-twin/
linked-ghost/click-birth), self-twin trap (stubs never vouch), lock guard, master-index
scaffold, legacy reports, one-open law via injected cache.
