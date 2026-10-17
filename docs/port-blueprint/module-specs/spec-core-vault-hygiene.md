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

## Shipped record — T-46, SESSION-021 (2026-10-17)

**Landed** (`internal/vaulthygiene`, hygiene.go + the golden generator
+ the golden + the behavior battery): the classifiers (`IsEmptyStub`,
`IsAppOwnedBodyless` — the ownership proof with the `---`/`...` fence
law and the unterminated-block door, `IsLockedFrontmatter` — the
truthy `locked:`/`category_locked:` doors with the flow/block tags
fold and the `...`-is-not-a-fence law, `HasNoteNameShape`,
`NormalizeIdentity`), the harvesters (`HarvestLinkTargets` — the
[[wiki-link]] eyes with anchors/aliases stripped, path-shaped targets'
last segment, embeds; the one walk's head-identity harvest — stems,
first H1 only, inline + block `aliases:`), and `RunVaultHygiene` —
THE ONE WALK (each .md opened and read exactly once: the size gate,
the shared notereads cache or the direct bounded read with the OSError
skip door, the ghost map with self-links excluded, the stub /
locked-falls-through-to-harvest / app-owned-bodyless / real-note
ladder), the three fingerprints (shape / name-twin / linked-ghost,
retirement needs ≥1), the retirement to `.trash/empty-stubs/` through
`storage.UniquePath` + the dry-run seam, the master-index scaffold
removal (`_index.md` + `_moc/*.md` proven by `type:`, the folder
prune with the owner-file law), the legacy root reports to
`.trash/retired-reports/`, the honest summary (the names cell, the
linked-ghost / app-owned why cells, the clean line),
`HealMachineVaults` (both machine vaults, the vault-labeled progress
wrapper) and `HygieneTotal`.

**Tests** (76 new module tests; 1464 race-clean at the close): the
ported test_strayheal selections (`vaulthygiene_test.go` — the
identity harvest through the one walk [the dead collector's legs ride
the twins they vouch for], the stray retirement incl. the owner's
exact scenario + recoverability + idempotence + the story + the
dry-run rehearsal + the frontmatter/body-text guards, THE
LINKED-GHOST DOOR incl. the app-owned bodyless stray + the unlinked
sacred law + the self-link trap, heal_machine_vaults + hygiene_total,
the pure classifiers), the ported test_vaulthygiene selections (the
hygiene pass's laws through `RunVaultHygiene` directly — the reference
drives them through the GUI worker's `_vault_hygiene_pass`, whose Go
half is the future GUI task's), and test_picklock's locked-bodyless
law; plus THE GOLDEN (9 sections — empty_stub, app_owned_bodyless,
locked_frontmatter, note_name_shape, normalize_identity,
harvest_link_targets, head_identities, frontmatter_type, run_hygiene)
GENERATED by `testdiff/goldens/dump_vaulthygiene.py` from the
unmodified reference @ 98f4fc7 (paths folded to `<vault>/…`; the
pass's counts pinned as ints with `stub_names` compared SORTED — no
per-run facts, no walk-order facts). test_vaultscan's one-open
hygiene-sharing leg landed in `internal/vaultscan`'s suite (the cache
is shared across hygiene + both plan walks: 4 opens for 4 notes). The
GUI legs (the scan door's both-vault heal, the startup heal, the
scorecard rows, the quiet-open timer) stay with the future GUI task;
the resolvable-link law legs are core/linking.py's own suite; the
release-bookkeeping legs are the Python repo's own CI facts.

**Go-reality notes** (the SESSION-021 amendments):

- **The walk-order fold**: the reference's hygiene walk rides
  os.listdir order (UNSORTED — unlike vault_scan's own sorted walks);
  the Go house walk rides ReadDir's sorted order. The retirement
  decisions are order-independent (the identity and link sets complete
  before any decision), and the story's names cell is the only
  order-bearing surface — the golden folds `stub_names` sorted, and
  the battery pins the story by substrings (the reference's own test
  law).
- **The OSError door**: the reference wraps getsize + the read in one
  `except OSError: continue` — the Go direct read keeps the error door
  (an unreadable file is skipped). Through the CACHE the house API
  cannot error (an unreadable file reads `''`) — a documented
  convergence at a door the reference's own tests never exercise.
- **The never-raises wrapper**: the reference's whole-pass
  `except Exception` (the "pass skipped" warning) is unreachable in
  Go — no path can abort mid-walk; the section writes land at the same
  points so the partial-counts law holds.
- **The None doors type-narrow**: Python `None` content reads True
  (empty stub) / '' (identity) / False (bodyless, locked) — Go's zero
  string carries every one of those verdicts; the reference's junk
  non-str shapes are type-impossible at the Go door (documented in the
  batteries).
- **The defined-but-unused constants**: `STUB_MAX_BYTES` /
  `APP_OWNED_BODYLESS_MAX_BYTES` judge nothing in the classifiers
  (shape is the law) — kept exported as the GUI mixin's parity
  surface, exactly as the reference keeps them.
- **The bounded read counts runes** (the T-45 lesson): the walk's cap
  is `walkReadMaxRunes` (1,048,576) + 1 at the call site, the size
  gate reads `st.Size()` (bytes) — a file at or under the byte cap is
  read whole, a bigger one contributes its stem only.
