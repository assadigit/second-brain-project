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

Port `test_phase6` selections + `test_noteredo` (48): block surgery idempotence (both blocks),
CRLF preservation, containment guards, suggestion lifecycle (pending→approved/rejected, never
re-suggested), decisions grammar, embeddings refresh staleness, recall sanitize rules
(prefix/plurals/220-cap), first-wins dedupe documented.
