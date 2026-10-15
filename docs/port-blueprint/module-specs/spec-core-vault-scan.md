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
