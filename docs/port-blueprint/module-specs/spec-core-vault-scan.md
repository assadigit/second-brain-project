# Module Spec — `app/gitcurator/core/vault_scan.py`

- **LOC**: 3,353 · **Go home**: `internal/vaultscan` · **Status**: port
- **Responsibility**: THE VAULT SCAN — the LLM-assisted refiling engine with the deterministic
  "regex brain" (declared-home, sibling-domain, taxonomy-tag passes), weighed-evidence scoring,
  chunked LLM filing, validated plan building, and the confirm-then-apply machinery. `kind`
  door: same grammar for Websites and GitHub Repos vaults.

## Constants (`vault_scan.py:`)

`SCAN_PREFIX "🔍 Vault scan:"` · prompts `s01_vaultscan.txt` / `s02_reposscan.txt` (read
directly) · `DEFAULT_TOO_BROAD 40` · `DEFAULT_MAX_NOTES 300` · `DEFAULT_LLM_BATCH 60` ·
`_PROTECTED_FOLDERS ('_review','_moc','_inbox')` (+ `_missing` for repos) ·
`_FRONTMATTER_READ_BYTES 16_000` · `WEIGHT_TOP_K 3` · `WEIGHT_TIE_MARGIN 0.08` ·
`STRONG_EVIDENCE_WEIGHT 0.55` · `OVERRIDE_MARGIN 0.15` · `_SAFE_FOLDER_NAME
^[A-Za-z0-9][A-Za-z0-9 _\-&+()']{0,39}$` · `_SAFE_NOTE_NAME ^[^/\\:*?"<>|\r\n]+$` · lock keys
`('locked','category_locked')` + folded tags lock/locked · curated key/tag · ~70 taxonomy
stopwords · `VAULT_KIND_WEBSITES/'repos'`.

## Pipeline (build_scan_plan 2323-2678)

1. **Inventory** (`scan_vault_inventory` 548-752): `os.walk` with sorted dirs/files; skip
   `_inbox` + dot-folders; per note: bounded 16 KB head read (via NoteReadCache when injected);
   rows carry `file/path/rel/title/source/domain/category/subcategory/tags/fetch_status/
   app_owned/folder/in_root/core/locked/curated/stars/primary_language/org/protected`;
   progress every 25 notes.
2. **Candidate radar** (`_move_candidates`): app-owned, not protected/core/locked/curated, and
   (in_root OR uncategorized OR stray `'/' not in folder` OR too-broad folder). Consolidation
   merge-notes excluded from the filing chain.
3. **Regex brain** (websites): `_declared_home_moves` (note's own cat/sub → existing folder,
   folded identity, `_disk_spelling` canonicalization, `_fold` skip-if-equal) →
   `_sibling_domain_moves` (domain's settled notes agree on ONE folded home) →
   `_taxonomy_tag_moves` (matched tags' home-sets intersect to one; home exists on disk).
   Repos: declared-home only (CATEGORY_FOLDERS key → folder).
4. **Weighed evidence** (`weigh_candidate_homes` 1489-1694) per depth-legal folder:
   `+0.30×(votes/dom_total)` siblings (repos: language agreement, same arithmetic), `+0.10`
   unanimous; `+min(0.06×shared_tags, 0.18)`; `+0.10` taxonomy word; `+min(0.05×title_overlap,
   0.10)`; `+0.06` folder-name match; declared echoes `+0.12` category / `+0.15` subcategory
   (repos `+0.25` folder match); **prior only when w>0**: `+min(0.05, bit_length(count+1)/24)`;
   too-broad `−0.10`; final `round(min(w,1.0), 2)` (banker's rounding!); top 3 by
   `(-weight, folder)`; `homes_close` when top two within 0.08.
5. **Chunked LLM filing** (`_llm_propose` 1906-2059): ≤300 candidates, chunks of 60; digest
   lines with weighed shortlist + CLOSE CALL marker; system prompt s01/s02; expected JSON
   `{"new_folders":[], "moves":[{"note","to","reason","confidence"}], "summary"}`; per-chunk
   failure drops only that chunk; un-proposed notes in landed chunks → `reviewed` (the census).
6. **Validation** (`_validate_moves` 2126-2320): deterministic moves ride FIRST; only listed
   candidates; destinations existing-folder / declared-home extra / validated new folder; depth
   law (`'/' in dest`); fresh-eyes absorption (`_absorb_near_duplicate`); intra-plan folded
   dedupe; never own-folder (folded); never twice (`taken`); `_bend_path` pre-bends colliding
   names (`Name_v1.md`, row carries `as`); LLM `confidence:"low"` → `low_confidence`;
   evidence-override: top ≥0.55 and chosen unlisted or ≥0.15 lighter → flagged low.

## Apply (`apply_scan_plan` 2848-3353)

Consolidations (keeper unlocked; checked merges: URLs appended via `add_site_links_to_note`,
merged files → `.trash/consolidated`, ledger re-points) → picked new folders (dryrun.makedirs)
→ moves (lock re-read at apply; `_bend_path` `as` honored; verified landings — vanished source
or did-not-land → `failed_moves` with named files) → census stamps (`curated: true` on moved
notes at new home, on declined rows when a selection spoke, on `reviewed`) → deletions (repos:
`repo_banish.banish_marked_repos(items=picked)`; websites: `consume_decommission_table(apply_banish=True,
skip_canonicals=…)` + `banish_marked_notes(skip_canonicals=…)` — unchecked/mid-plan-locked
canonicals held out).

Selection API: `normalize_selection` / `picked_counts` (the GUI picking door's checked rows).

## Migration-sensitive semantics

- Folded identity = `lower()` + strip `[^a-z0-9]+` (not casefold) — pin Unicode tests.
- `round(…, 2)` banker's rounding; `(count+1).bit_length()` → `bits.Len(uint(count+1))`.
- Sorted walk (deterministic) — Go lexical walk matches.
- 16 KB head read (bytes here, unlike note_state's 800 chars — both preserved as-is).

## Tests

Port selections from `test_vaultscan` (91), `test_weighhomes` (49), `test_onespelling`,
`test_censusstamp` (44), `test_reposcan` (38), `test_picklock` (45): radar pass triggers,
weighing arithmetic table, validation refusals, bending, absorption, census stamps, apply
failures ledger, repos-kind parity cases.

---

## Shipped record — T-45, SESSION-020 (2026-10-17)

**Landed** (`internal/vaultscan`, 8 files + the golden + three
behavior batteries): the radar (`scan.go` — the constants, the kind
door, the truthy/lock/curated predicates, the direct frontmatter
parser with vault_scan's OWN grammar preserved beside note_reads'
unified parse exactly as the reference keeps both, the CPython
urlparse netloc/path shapes, the core-note law, the inventory walk,
the move candidates), ONE NOTE PER SITE (`duplicates.go` — the
magnet, the locked-canonicals door, the tree/digest lines, the fold
laws, the sanitizer, the disk spelling + the shared resolver), the
regex brain (`regexbrain.go` — declared home / sibling domain /
taxonomy, each with the settled teaching set), THE WEIGHED EVIDENCE
(`weighing.go` — the census, the scale with banker's rounding and the
0.05-capped prior, the CLOSE CALL attachment, the homes cell), the
chunked LLM pass (`llmpass.go` — the exact user message, per-chunk
failure isolation, the honest review list), the validation door
(`validate.go` — fresh eyes incl. intra-plan dedupe, depth, taken,
the landable move with the pre-computed `_v` bend carried as `as`,
the low-confidence flag + the evidence-override check), the plan
(`plan.go`), and the apply (`apply.go` — consolidations, folders,
moves with the lock re-read and the spoken failure ledger, the census
stamps, the deletions through the machinery's own skip doors, the
repos ledger re-point).

**The repos banishment floor came first** (in `internal/repobanish`):
vaultscan's kind door consumes `scan_pending_repo_banishments` and
`banish_marked_repos`, and T-45's acceptance (the ported test_reposcan
selections) exercises them end to end — so those two functions plus
the note-finder landed WITH T-45 (their battery rides vaultscan's
suites); revive + the orchestration + test_repobanish stay T-46's.

**Tests** (215 new, 1347 race-clean at the close): the ported
test_vaultscan / test_censusstamp / test_picklock core selections
(`vaultscan_test.go` — the inventory, the candidates, the grammar
reuse, the LLM pass, the three laws, the apply, THE LANDABLE MOVE
battery incl. the `_windows_eyes` replay through the new stat seam,
the regex brain, the consolidations apply, the declared-home pass,
the chunked filing, the sibling/taxonomy passes, the one-open scan,
the census grammar/radar/writer/apply/reviewed/second-scan, the lock
predicate/parse/radar/apply, the skip door, the picking door, the
confidence law, the selection law); the test_weighhomes selections
(`weighhomes_test.go` — the word tokens, the teaching set, the
census, the weighing folder set, the scale signal-by-signal, the
attachment, the homes cell, the alignment door, the wiring); the
test_reposcan selections (`reposcan_test.go` — the inventory/plan/
apply kind doors, ONE NOTE PER REPO, the repos digest + prompt, the
repos regex brain + weighed evidence, the round trip, the repos
dry-run rehearsal); and THE GOLDEN (21 sections — kind doors,
predicates, core-note, frontmatter, pyurl, fold laws, sanitize,
candidate reason, digest lines, homes line, the full mini-vault
inventory walk, settled homes, folder profiles, weighing set, the
weigh arithmetic table, the attachment, absorption, the bend, the
disk spelling, the selection arithmetic, the validation door)
GENERATED by `testdiff/goldens/dump_vaultscan.py` from the unmodified
reference @ 98f4fc7 (paths folded to `<vault>/…` — no per-run facts).
The GUI legs (the Scan channel's Telegram round trip, the ScanPlan
checkbox battery, the ask-gate, the two-vault worker job, the
scorecard models) stay with the future GUI task; the
release-bookkeeping legs are the Python repo's own CI facts; the
hygiene leg rides T-46's suite (test_onespelling's state-boundary
selections already live in T-23's port).

**Go-reality notes** (the SESSION-020 amendments):

- **THE POINTER-THREADING LAW**: the reference threads Python dict
  references from the candidates list through
  `attach_weighed_homes` into `_validate_moves` — the validator
  reads the homes the scale attached. Go struct copies would break
  that aliasing silently (found live: the evidence-override flag
  never fired); the chain threads `[]*NoteRow` end to end, and
  `MoveCandidates`' value rows are re-taken by path from the
  inventory where the aliasing matters.
- **The stat seam** (`fsExists`/`fsIsDir`/`fsIsFile`): the reference
  suite's `_windows_eyes` patches os.path's exists/isdir/isfile to
  replay the owner's case-insensitive Windows disk on POSIX; the Go
  door routes every os.path exists/isdir/isfile call in this module
  through the seam so the same replay rides the ported tests.
- **The urlparse shapes**: the inventory's domain cell is
  `urlparse(src).netloc` (the \t\r\n removal, case-insensitive
  scheme, netloc case-PRESERVED, the bracket law's raise reading as
  "no parse") and the keeper rank is `urlparse(src).path` (query and
  fragment stripped, `_splitparams` cutting the last segment's
  `;params`, the except sentinel "z") — hand-rolled in `scan.go`
  (Go's net/url is not CPython's urlparse; the pyurl precedent law).
- **Python truthiness at the JSON door**: the guest's
  `summary`/`note`/`confidence` cells fold through Python truthiness
  (`pyTruthyAny`) — a JSON `false` summary is NOT set, an empty
  string is not.
- **The frontmatter head read is 16,000 CHARACTERS** (Python
  text-mode `read(n)` counts chars; the constant is named BYTES in
  the reference — the spec's migration note preserved the read, the
  Go constant is named `FrontmatterReadRunes` to say what it is).
- **The selection's junk tolerance**: the reference's junk shapes
  (strings/ints as section values) are type-impossible at the Go
  door's `map[string][]string` — every unrepresentable shape reads
  as its section's nil, the same verdict the reference's junk law
  gives them (documented in the golden test).

**Dependency posture**: zero new (go.mod/go.sum untouched,
`git rev-parse HEAD:go.mod` = 5c0d4e4a, the same bytes as the
SESSION-019 close).
