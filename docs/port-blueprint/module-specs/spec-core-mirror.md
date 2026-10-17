# Module Spec — `app/gitcurator/core/mirror.py` + `mirror_keys.py` (+ Phase-6 group: `linking.py`, `embeddings.py`, `recall.py`)

- **LOC**: 701 + 26 + 624 + 393 + 386 · **Go home**: `internal/mirror` (one package:
  `mirror.go`, `linking.go`, `embeddings.go`, `recall.go`) · **Status**: port (merged group —
  the linking layer is the mirror's Phase-6 extension; `mirror_keys` exists only to keep the
  Python import graph acyclic)
- **Responsibility**: the read-only `Library/` mirror of machine-library notes into the Manual
  vault; recall hooks; local embeddings; suggested links with owner approval; Related (auto)
  blocks in mirror copies only.

## mirror.py (`mirror.py:`)

- `LIBRARY_FOLDER "Library"`, trees `GitHub Projects` / `Websites`; MIRROR_BANNERS (read-only
  warning callouts); skip marks `('_moc','_inbox','_missing','attachments','.obsidian','.trash')`
  (substring on path).
- `_safe_join(root, rel)` (118-144): rejects absolute rel, backslashes→`/`, empty/`.`/`..`,
  `:` (drive letters); realpath containment re-verified; `MirrorError("unsafe mirror path …")`.
- `check_vault_paths(...)` (147-191): overlap refusal in both directions.
- Sources = `.md` with a `source:` line in the first **800 chars**; mirror tree indexed by
  `mirror_of:` value; unmarked files never touched (conflicts counted).
- `build_mirror_note` (321-361): inserts `mirror_of: <raw source>` before closing `---`;
  read-only banner as first body line; strips banner refs (`cover:` line + `![banner]` embed);
  preserves the note's own line ending; carries the Related (auto) block via
  `preserve_related_block`.
- Plan/apply: matching by RAW source → moves become remove+create; updates only when content
  differs; deletions before writes; empty dirs cleaned bottom-up; dry-run by default
  (`run_mirror(..., apply=False)`); idempotent.

## linking.py (`linking.py:`)

- `RELATED_START/END` markers; caps `LINKS_CAP 7`, `LINKS_TARGET 5`, `DEFAULT_TOP_K 8`,
  `DEFAULT_MIN_SCORE 0.30`; statuses pending/approved/rejected.
- `collect_note_fields(vaults)` (75-115): managed notes; **first hit wins** on duplicate
  sources (walk-order dependent!); stem = filename minus `.md`.
- `find_candidate_pairs` (122-149): O(n²) cosine, per-note top-k, pair canonicalized
  `tuple(sorted)`, best score per pair, final sort desc.
- `build_confirm_prompt` (l01 slots, each `[:300]` with fallbacks) / `parse_confirm_reply`
  (`related` bool or true/yes/false/no; unparseable → NOT related) / `confirm_pair`.
- `LinkStore` (205-316): table `link_suggestions (key_a, key_b, score, status, reason,
  created_at, decided_at, PRIMARY KEY(key_a,key_b))`; `suggest` (known pair → existing status),
  `decide` (pending→terminal only; invalid status error), `approved_links` (both directions,
  per-note sort `(-score, other)`, cap 7).
- `build_suggestions_note` / `collect_decisions` (checkbox lines with two backticked URLs on the
  next ≤2 lines; approved = `x/X` not struck; rejected = struck) / `apply_decisions`.
- `apply_related_blocks(library_root, …)`: mirror copies only (identified via `mirror_of:`);
  insert/replace/remove the delimited block; containment guard `_is_inside_library`.

## embeddings.py (`embeddings.py:`)

- `embed_text_for(fields)`: name + one_line + recall + `"tags: …"` joined `\n`, cap 1200 chars.
- Providers: Ollama `/api/embed` with legacy `/api/embeddings` fallback (per-text prompt);
  OpenAI-compatible `/embeddings` (honors per-item `index`); `build_embedding_call` provider
  routing (llamacpp/cloud/ollama with model fallback chains); loopback never proxied.
- `EmbeddingStore`: table `embeddings (vault, source_url, model, dim, vec JSON, text_hash,
  updated_at, PRIMARY KEY(vault, source_url, model))`; `refresh_embeddings` batches of 16,
  stale-skip by text_hash, failures never abort.
- `cosine` pure float64; `hash_embed_text` sha256.

## recall.py (`recall.py:`)

- Block markers from note_state; `RECALL_PLACEHOLDER "Use when you need to… (not captured — see
  the source link)."`; sentence must start `"use when you need to"` (ci), one line, ≤220 chars
  (code points), else placeholder.
- `insert_recall_block`: in-place replace (idempotent) or append; CRLF detection
  (`count('\r\n') > count('\n')/2`).
- Field extractors: GitHub (H1, TL;DR, `## What is it?`, `## Key Features & Technologies`,
  tags) and websites (name, one_line, `## Best used for`); `recall_fields_for` uniform shape.
- `plan_recall_hooks` (read-only) / `run_recall_hooks` (LLM via r01, sanitize, write
  atomically; per-note failures captured, never abort); reports under `APP_DIR/reports/recall`.

## Tests

Port `test_phase5` (core) + `test_phase6` (core) selections: block surgery idempotence (both
blocks), CRLF preservation, containment guards, suggestion lifecycle
(pending→approved/rejected, never re-suggested), decisions grammar, embeddings refresh
staleness, recall sanitize rules (prefix/plurals/220-cap), first-wins dedupe documented.

**Reality-wins correction (2026-10-17, SESSION-025, same commit as the port)**: this section
previously read "Port `test_phase6` selections + `test_noteredo` (48)". Evidence at hash level:
`app/tests/test_noteredo.py` @ 98f4fc7 is the **v0.57.0 website-pipeline redo suite** — the ✅-
earned-by-the-note machinery (`note_is_properly_stored`, `take_hand_delivered(allow_consumed)`,
`scan_master_redo_rows`, `process_link` redo, CLI routing, GUI wiring) — it contains ZERO
mirror/linking/embeddings/recall tests. The parity doc's own mapping (09-testing-and-parity.md)
already routes it elsewhere ("notereads cache laws" — also a mislabel). The T-40 port landed the
machinery (`internal/websitepipeline/decommission.go`: `NoteIsProperlyStored`,
`ScanMasterRedoRows`; `pipeline_process.go`: the redo-consume) with golden coverage
(`note_is_properly_stored`, `take_hand_delivered_redo` sections); the suite's remaining
un-ported selections are recorded as deferred test debt on the books (see NEXT-SESSION).
The binding acceptance for T-50 (task list) is: block-surgery idempotence; containment guards;
the suggestion lifecycle — all test_phase5/test_phase6 territory, all landed.

## SESSION-025 shipped record (T-50 — landed)

`internal/mirror` — one package, four files + the folded mirror_keys leaf:

- **mirror.go** — `mirror.py` + `mirror_keys.py` (the Python leaf exists only to dodge the
  import cycle; Go has none): `LibraryFolder`/tree names, `MirrorBanners`, `SkippedFolderMarks`,
  `sourceRE`/`mirrorRE` (Python `\s` → the house `pyS` Unicode class), `MirrorError`,
  `isInside`/`pyRealpath` (commonpath semantics; component-wise symlink resolution with a
  bounded hop count), `SafeJoin` (every traversal defense; the exact refusal messages),
  `CheckVaultPaths` (both-direction overlap refusal), `scanSourceNotes` (sorted-then-deduped,
  first-wins by rel order), `scanMirrorTree` (no skip rules — orphan detection sees every
  marker), `stripBannerRefs` + `lineEndingOf`, `BuildMirrorNote` (the front-matter insert
  before the closing `---`, the fresh-block wrap, the blank-line guarantees), `TreePlan`/
  `MirrorPlan` + `RenderReport` byte-faithful, `planTree` (RAW-source matching; moves become
  remove+create; `doomed`; duplicate-marker collapse), `RunMirror` (dry-run default;
  deletions-before-writes; the marker re-checks; `dryrun.Remove`/`dryrun.MkdirAll` seams;
  `storage.AtomicWriteText`; bottom-up empty-dir cleanup; the tree/Library rmdir attempts).
- **recall.go** — the block surgery (`StripRecallBlock`/`HasRecallBlock`/
  `GetRecallSentence`/`RenderRecallBlock`/`InsertRecallBlock` with the CRLF-dominance law),
  `SanitizeRecallSentence` (quote-strip, one-line over Python's splitlines boundary set,
  ci-prefix, bare-prefix refusal, the 220-rune cut at the last word boundary with the
  `',;:'` char-set rstrip), the field extractors (the section regexes with `$\r?\n`
  anchors over `(?ms)`), `RecallFieldsFor`, `BuildRecallPrompt` (r01 slots with the
  `or` fallbacks), `PlanRecallHooks` (notestate.ScanVault; unmanaged skipped), `RunRecallHooks`
  (per-note failure isolation; the dry-run/apply/unchanged doors; the writer seam),
  `RecallReport`, `ReportDir`.
- **embeddings.go** — `EmbedTextFor` (cap 1200 RUNES), `HashEmbedText` (sha256 over the
  encode-replace twin), `Cosine` (pure float64), `postJSON` (loopback never proxied — the
  v0.15.1 rule; non-JSON body → the repr-prefixed EmbeddingError), `OllamaEmbed` (batched
  /api/embed with the detected legacy fallback), `OpenAIEmbed` (index-honoring reordering,
  count-mismatch + missing-index errors), `BuildEmbeddingCall` (llamacpp/cloud/ollama routing
  with the `or`-falsy doors over `float(cfg.get('llm_timeout_s', 300) or 300)`), `EmbeddingStore`
  (own table on `store.Open`, the "cache.db"→APP_DIR anchor, json.dumps vec bytes), 
  `RefreshEmbeddings` (batches of 16, stale-skip by text_hash, failures never abort).
- **linking.go** — `CollectNoteFields` (ordered `VaultPath` slice — the reference dict's
  insertion order drives the first-wins dedupe), `FindCandidatePairs` (ordered `VecEntry`
  slice — Python dict order is load-bearing for tie stability; pair canonicalization; the
  best-per-pair + score-desc stable sorts), `BuildConfirmPrompt` (l01 slots `[:300]` with
  fallbacks), `ParseConfirmReply` (nil = unparseable → NOT related), `ConfirmPair`,
  `LinkStore` (PRIMARY KEY (key_a,key_b); suggest/decide with decided_at;
  never-resuggest; approved_links capped 7 sorted (-score, other)), `BuildSuggestionsNote`
  (the exact note bytes; the frozen-clock Generated line), `CollectDecisions` (the checkbox/
  strike grammar with the backtick URL pairs), `ApplyDecisions`, the Related-block surgery
  (render/insert/extract + the removal pass), `ApplyRelatedBlocks` (mirror copies ONLY —
  the `mirror_of` keying through `links.NormalizeURL`; the containment guard),
  `PreserveRelatedBlock` (the Phase-5↔6 carry-over hook).
- **THE GOLDEN** (39 sections, 37,117 bytes, `testdata/mirror.golden.json`): GENERATED by
  `testdiff/goldens/dump_mirror.py` from the unmodified reference @ 98f4fc7 — the generator
  imports the reference's own test_phase5 + test_phase6 fixtures via importlib (zero
  transcription risk); run twice byte-identical. Sections: mirror_keys + the unquote matrix,
  the seven build_mirror_note shapes (exact bytes), strip_banner_refs, line_ending_of,
  safe_join (the refusal messages), check_vault_paths (the nine-case refusal matrix),
  the recall block surgery + the 16-case sanitize matrix, the field extractors, the filled
  r01, embed_text + hashes, cosine, the embedding store + refresh staleness, the candidates +
  the confirm parsing (10 reply shapes), the link store lifecycle + the capped approved links,
  the Suggestions note bytes (incl. the stem|alias resolvable-link shape), the decisions
  grammar (4 variants), apply_decisions, the Related blocks + apply_related_blocks (mirror
  copies only + the removal pass), collect_note_fields, THE MIRROR FLOW (dry → apply →
  idempotent → move → orphan → conflict → duplicate over the reference's fixture trio —
  plan stats, logs, reports, the exact tree bytes), the recall flow (dry/apply/unchanged/
  garbage + the report), and THE PHASE-5↔6 RESYNC INTEGRATION (the approved block survives
  the rebuild; an unchanged note is a plain keep).
- **The batteries** (90 new top-level tests): the ported test_phase5 core selections (39 —
  TestBuildMirrorNote 6, TestSafetyRefusals 8, TestSafeJoin 5, TestNothingOutsideLibrary 5,
  TestSync 15) + the ported test_phase6 core selections (40 — TestRecallBlock 9,
  TestRecallPipeline 8, TestEmbedMath 3, TestEmbedEndpoints 5, TestEmbeddingStore 2,
  TestCandidates 2, TestConfirm 3, TestLinkStore 2, TestSuggestionsNote 2, TestRelatedBlocks
  4, TestCollectFields 1 — the loopback fake-embed server is the reference's _FakeEmbedHandler
  port, NO real network) + the golden differential (11 replay tests). 1826/1826 race-clean
  over 30 packages.
- The tool/CLI legs (mirror_manual, add_recall_hooks, build_links, the CLI status row, the
  GUI wiring) are Phase-5 CLI-parity territory (T-60), as the reference's own suite splits
  them; the tools' core surfaces are all landed here.

## SESSION-025 Go-note amendments

- **The map-order law bites the LOGS**: `plan.Trees` is a Go map — the reference's
  `for tree in plan.trees.values()` rides Python dict insertion order (GitHub Projects,
  Websites). RunMirror's say-loops must iterate the fixed `treeSpecs` order (the golden
  caught it live: the log lines came out in random order across runs).
- **The []-vs-null byte law (T-49) extends to `.get()` dumps**: a golden section dumping
  `r.get('written')`/`r.get('error')` renders NULL when the key was never set — the Go
  replay must emit nil for never-set fields, not the zero value.
- **The house JSON decoder returns int64 for integral literals** (storage.decodeNumber):
  any float conversion door (`float(x)` in the reference) must accept int64/int/json.Number.
- **`str.splitlines`' boundary set** is wider than `\n` (\v, \f, \x1c-\x1e, \x85, U+2028/9,
  \r\n as ONE) — `pySplitLines` is the twin; `firstPyLine` cuts at the first boundary.
- **The reference's exception-text rendering** (`f"{type(e).__name__}: {e}"`) stays
  unpinnable in Go (the T-40 law): the pinned contract is the SHAPE (the entry degrades,
  the run continues), recorded in doc.go.
- **Walk-order independence is earned, not assumed**: every mirror output is sorted
  afterward (notes by rel, deletes sorted, rels sorted); the order-dependent surfaces
  (collect_note_fields first-wins, apply_related_blocks' updated list, run_recall_hooks'
  results) pin sorted in the golden — the reference's own os.walk/listdir order is
  filesystem-dependent and unpinnable. The first-wins law's CHOICE is walk-order dependent
  in the reference; the Go walk is lexical (documented divergence — the house ScanVault
  law), and the golden's duplicate fixtures use identical content so the choice converges.
- **The clock seam**: linking + embeddings stamp `time.strftime` — the package `clockNow`
  var is the freeze point (the generator patches `lk.time`/`eb.time` with a frozen
  strftime on the Python side).
- **postJSON's wire bytes** ride `storage.PyDumpsCompact` (json.dumps default separators);
  the loopback-vs-proxy door rides `llmclient.IsLoopbackURL`.
