# Module Spec — `app/gitcurator/core/website_pipeline.py` (with `website_state.py`)

- **LOC**: 5,338 + 755 · **Go home**: `internal/websitepipeline` (+ `state.go`) ·
  **Status**: port (the pipeline owns its state; the directory-note
  regeneration lives in its own spec —
  `spec-core-website-directory.md` — since T-41)
- **Responsibility**: the complete Websites pipeline: per-link gates → five-door fetch →
  extraction → two-pass taxonomy classification → analysis → note write; plus the verdict
  grammar (master table, banishment + Telegram gate, consolidation, settled ledger) and the four
  batch drivers. `llm_call`, `vault_index_has`, `site_note_for`, `banish_confirm`, `fetch_fn`
  are **injected** (hermetic law).

## Module constants (`website_pipeline.py:`)

`WEBSITE_PROMPT_VERSION "web-v2"` · `REVIEW_FOLDER "_review"` · `FETCH_TIMEOUT_S 20` ·
`FETCH_MAX_BYTES 2_000_000` · `DOMAIN_DELAY_S 2.0` · `CLASSIFY_RETRIES 2` ·
`LLM_EXCERPT_CHARS 4000` · `MIN_TEXT_FOR_ANALYSIS 80` · `MAX_FETCH_RETRIES 3` /
`RETRY_BACKOFF_DAYS 2` (from website_state) · `DEAD_MARKERS`/`REVIVE_MARKERS`/`REVIEWED_MARKERS`
(616-629) · `BANISH_EMOJI "🗑️"`, `BANISH_WORDS` 6, `BANISH_KEYS` 3 (666-669) ·
`BANISH_QUARANTINE_RELPATH ".trash/banished"` · `BANISH_REASON_PREFIX "banished by owner"` ·
`BANISH_TALLY_PREFIX "🗑️ Run tally:"` · `BANISH_GATE_ENDPOINT "/api/banish"` · canonical master
header `| # | Date | URL | Domain | Source | Status | Notes |` (7 cells).

## Per-link state machine (`_process_link_inner` 4370-4875)

```
QUEUED URL → canonicalize (normalize_website_url)
GATES (order): directory banish list (60 s cache, fail-open) → blocked domain → self domain →
  dismissed (reason-coded error) → already-in-vault (unless failed-_review prior → UPGRADE, or
  🖐 gesture + app _review note → UPGRADE) → ONE-NOTE-PER-SITE consolidation (site note append +
  placeholder sweep + retry resolve + harvest) → retry gate (attempts ≥ 3: 🖐 reborn resets,
  else skipped "no more retries")
FETCH: hand-delivery (allow_consumed on gesture) → fetch_fn ladder
  failure cat ∈ {dead, paywalled, refused, bad_url} → resolve_retry + dismiss("auto-verdict: …")
  healable → enqueue_retry (attempts+1; next = now + 2×attempts days)
  always write/refresh failed _review note + _record('failed') → outcome 'review'
  archived rescue → partial note, retry resolved
EXTRACT: full | partial (js_shell/paywall → partial)
CLASSIFY: locked-row override → w01 (category; invalid ×(1+2 corrective) or low-confidence →
  _review) → w02 (subcategory; "none" → cat-only; low → sub dropped)
ANALYZE: w03 → failure → _review note "Analysis failed: …"
WRITE: unique_path <Cat>/<Sub>/<Name>.md (name fallback chain) → atomic write → upgrade cleanup
  (fingerprint-proven) → mark_processed → resolve_retry → 🖐 harvest (only when properly stored)
  → log "✅ [VaultName] <name> processed and stored → <rel>"
process_link wraps all: unexpected exception → _review "Pipeline error: …" note; never raises.
```

Batch drivers: `run(urls, should_continue, on_progress)`, `run_due_retries`,
`retry_review_backlog`, `retry_master_waiting`; all end with `_refresh_master_table` (waiting
rows `unreviewed` → stamp `📁 stored` → harvest hand rows → prune terminal rows).

## Banishment machinery

`scan_banished_notes` (both doors: frontmatter keys, tags list, body inline tags — via
note_reads unified parse; deduped canonical) + `scan_decommission_table` gestures → count spoken
→ injected `banish_confirm(items, log)` (verdict `confirmed/declined/timeout/defer/auto`;
config `banish_confirm:false` = auto) → `apply_banish` gates `consume_decommission_table` and
`banish_marked_notes` (move to `.trash/banished`, dismiss, twins sweep in the other vault) →
tally line always spoken; `prune_decommission_table` keeps only pending rows (compact table).

## Note formats

**Website note** (`build_website_note` 97-217): frontmatter `source, aliases: [], tags[…],
category, subcategory, fetch_status, pricing (free|freemium|paid|unknown), login_required
(yes|no|unknown), date_processed, managed_by: "gitcurator", schema_version: "1",
prompt_version: "web-v2"`; body `## What it does` → `## Best used for` (**always present** — the
recall regex needs the heading) → `## Practical details` → `## Watch out` → `## Similar tools`
→ `*Source: [url](url)*` → v0.60.0 delete-hint block. Sanitizer caps: name 120, one_line 2000,
what_it_does items 200, best_used_for 400, practical 4×200, watch_out 2×200, tags ≤8, similar ≤5.
Gists get tag `snippet`.

**Review note** (`build_review_note` 220-264): `> [!warning] Needs review — <status>` callout +
one `> <reason>` line. **Consolidation** (`add_site_links_to_note` 294-418): the ONE sanctioned
rewrite — `site_links: [..]` frontmatter line after `source:` + `## Links on this site` section
before the footer; idempotent.

## WebsiteStateDB (`website_state.py:71-755`)

Tables (94-132): `websites_processed(url PK, note_path, category, subcategory, fetch_status,
processed_at)` · `website_retry_queue(url PK, attempts, last_error, next_attempt_at,
first_failed_at, last_failed_at)` · `dismissed_urls(url PK, reason, dismissed_at)` ·
`website_state_meta(key, value)` · `websites_settled(url PK, settled_at)`.
Key methods: `mark_processed` (INSERT OR REPLACE, naive-local `isoformat(seconds)`),
`enqueue_retry` (attempts+1; `next = now + timedelta(days=2*attempts)`; `last_error[:500]`),
`resolve_retry`, `dismiss(url, reason)` (reason grammars: `"banished by owner — 🗑️ …"`,
`"auto-verdict: …"`, `"decommissioned by owner — graveyard table (v0.44.0)"`,
`"retired by owner (reviewed ✅) …"`, `"blocked domain (v0.20.0 purge)"`), `undismiss`,
`forget_failed_row` (failed+_review only), `forget_row`, `settle_existing`,
`settle_queue_history`, `normalize_ledger_keys` (healing re-key; preferences: live note beats
dead, more attempts wins, first by `ORDER BY url` for dismissed/settled), `reset_retry_attempts`,
`rearm_retries`, `purge_blocked_domains`, `due_retries` (attempts<3 ∧ next ≤ now; unparseable
date counts as due). `_key()` = `normalize_website_url` at the DB boundary (ONE SPELLING).
Dead: `settle_urls` (no callers).

## website_directory.py

**Moved to its own spec at T-41 (SESSION-016, reality-wins): the task list and
the state card had always referenced a standalone `spec-core-website-directory`
that never existed — the content below was this file's whole coverage of the
module. The full spec (sweep + directory + the light parse + the walk + the
golden contract + the T-41 shipped record) now lives in
`spec-core-website-directory.md`; `websitepipeline/directorynote.go` is its
port.**

## Tests

State-machine table tests (every gate/outcome); retry arithmetic; dismissal reason grammars;
banishment flows incl. gate verdicts; consolidation idempotence; master-table lifecycle; settle
passes; directory-note golden; pipeline constructor's gate sequence with fake confirm channels.
Port from `test_banishment`, `test_phase*` selections, `test_websitesqueuefix`,
`test_reviewretry`, `test_mastertable`/`test_masterretry`, `test_onespelling`.

> **SESSION-015 (T-40) — what actually shipped, and the layering
> record.** The full pipeline landed: the note builders, the
> consolidation, the review scans, the table grammar + writers, the
> banishment + its confirmation gate + the tally, the settlement, the
> splits, the master-table refresh/stamp/harvest/prune, and the
> WebsitePipeline class (the constructor's gate sequence, the
> per-link state machine, the four drivers). The acceptance's three
> stones: the state-machine table (pipeline_test.go — every
> gate/outcome), the banishment flows confirmed/declined/timeout/
> defer/auto/broken-channel (banishgate_test.go), and THE DIFFERENTIAL
> GOLDEN (testdata/websitepipeline_pure.golden.json — 51 sections,
> 78,496 bytes, generated by testdiff/goldens/dump_websitepipeline.py
> from the unmodified reference with every clock frozen at
> 2026-10-17 12:00:00; no per-run facts inside). 38 new top-level
> tests; 910 race-clean at the close.
>
> Go's no-deferred-imports law forced a layering the reference
> dissolves with Python's function-level imports: the table grammar
> (parse + status predicates + canonical header) lives in the new
> `internal/websitepipeline/decomtable` subpackage, below both
> websitepipeline and handdelivery (whose stamps consume it). The
> handdelivery surface the pipeline needs (gestures, queue read/write,
> Collect/Take, the harvest stamp) shipped into `internal/handdelivery`
> as T-40's dependency floor — T-43 completes that module (the Chrome
> launch family, the batch-level pickers, its own battery); the
> directorysync banish-gate helpers (LoadBanishList/IsBanished) ship
> likewise ahead of T-49. The v0.28.0 banned-domains DISK sweep was a
> documented seam (sweepBannedNotesSeam) until T-41 landed
> directorynote (SESSION-016) — the constructor now runs the REAL
> guarded sweep; the DB-side purge is enforced as before. The third door's
> constructor ladder collapses to "armed" (tls-client is always
> compiled in — ADR-0006 §4; the auto-pip leg is obsolete by
> construction). state.go gained only the clock seam (nowSeconds reads
> the package clockNow — the golden's frozen-clock need, T-23's shapes
> unchanged); notebuilder gained SanitizeTagsMax (the reference's own
> max_items parameter); storage.OrderedMap gained Delete (dict.pop —
> the hand-delivery queue heal's own primitive).

> **SESSION-016 (T-41) — the seam closed, and the golden it shipped
> forced three reality-wins fixes in this package's shared
> primitives.** directorynote.go landed (its own spec now:
> spec-core-website-directory.md): the constructor's sweep seam became
> the REAL guarded sweep, and the directorynote differential golden
> (testdata/directorynote.golden.json) caught latent divergences the
> T-40 fixtures never exercised: readHeadRunes now translates universal
> newlines before the rune cap (helpers.PyNewlines — every reference
> text-mode read of a note file does; CR-only and CRLF notes parse
> through \r→\n translation), decodeUTF8Replace now collapses each
> MAXIMAL SUBPART into one U+FFFD (notestate's own law; the previous
> version emitted one per byte), and walkVault now yields a directory's
> files before ANY of its subdirectories' files (os.walk's top-down
> order — WalkDir interleaves lexically; the banishment scans' list
> order rides the walk order). T-40's own golden recorded no U+FFFD and
> no interleaving-case orders, so it passes unchanged on the fixed
> primitives. 938 race-clean at the close.

> **SESSION-026 — the Phase-4 closer landed in this package: the
> 30-link offline golden runner equivalent (the exit gate's last named
> artifact).** testdiff/goldens/dump_goldenwebsites.py drives the
> UNMODIFIED reference's own `tools/run_golden_websites.py --offline`
> (CI's mode) over its own tests/golden/websites.json corpus — the
> fake LLM answering from the golden file, canned pages feeding the
> fetch, a throwaway vault, every clock frozen at 2026-10-17 12:00:00
> local, temp roots + the reference app dir folded by REPLACE — into
> testdata/goldenwebsites.golden.json (94,979 bytes; the generator
> runs its payload twice and refuses to write unless the two dumps are
> byte-identical — the determinism proof is part of every
> regeneration). goldenrunner_test.go replays the SAME corpus through
> this package with the same seams and byte-compares: the canned pages
> (Python str.capitalize's rest-lowercased quirk included — "AI
> illustration maker — …" titles as "Ai illustration maker — …"), the
> 90-call LLM choreography (classify, classify, analyze per link; the
> analysis prompt carries no `Website: ` line, so the fake's URL scan
> misses and every analysis answer is the DEFAULT dict — the quirk
> every golden note is named "Golden test site" by), the first three
> prompts, the per-link results, the 30 produced notes (the
> unique_path v-suffixes included), the report (the runner's own
> renderer ported byte-exactly; the taxonomy summary path folds
> `<repo>/assets` → `<app>` on the Go side), the console (the log
> lines + the runner's stats prints), and the exit-code law (0 iff no
> invalid taxonomy answer). ZERO port fixes were needed — T-40's
> machinery replayed the reference's own end-to-end run green on the
> first try; the mutation test (one byte in the golden's report
> section) fails with a precise diff. 1828 race-clean at the close.
> Phase 4's exit gate: CLOSED (the stamp + the stone table citation:
> 12-migration-roadmap.md §Phase 4).

> **SESSION-059 — the delivered-law correction (the noteredo
> core-level port found a REAL parity gap; reality wins, fixed in the
> same commit).** The deferred test debt's "TestRedoScan shapes" port
> (`internal/websitepipeline/noteredo_test.go` — 11 selections + the
> iconcolumn §7 never-re-ask leg) failed 4 selections on its first
> run, and the byte-level read of both sides settled why:
> `scan_master_hand_rows` + `scan_master_redo_rows` INLINE their own
> delivered probe in the reference (a queue entry marked consumed, OR
> its suggested page file present — `os.path.isfile`, NO size check),
> but the Go port had reused the BATCH-side collector
> (`handdelivery.CollectDelivered` — skip-consumed + getsize>0), so a
> CONSUMED row was invisible to both Go scans: the redo scan missed
> THE OWNER'S EXACT CASE (consumed + a half-fetched `_review` note is
> the very false-success the v0.57.0 feature exists to redo) and the
> hand scan re-asked a consumed row for a Chrome tab (the reference:
> "the record stays, the link is not re-scraped"). The 0-byte page
> file and the nil-meta queue row diverged likewise. The fix:
> `decommission.go`'s own `deliveredURLSet` — the reference's inlined
> law byte-for-byte (MetaStr/MetaTruthy exported from handdelivery as
> the queue-meta readers; `(meta or {})` nil-handling; a broken queue
> degrades to the empty set behind the never-raise law) — now feeds
> BOTH scans. The same port also forced the probe-ladder twin: the
> reference wraps `is_dismissed` + `processed_row` in ONE try/except
> ("a broken probe redoes nothing"), and the Go scan handled errors
> but PANICKED on the closed-DB probe shape (`store.DB.Close()` nils
> the handle by its own law; the reference's `_Broken` raises the
> Python-class twin) — the scan now recovers the panic as the broken
> answer. The harvest (`HarvestHandRows`) carries the reference's
> same try/except and still lacks the panic twin — recorded on the
> sharpened remainder, not fixed here (one slice, done properly).
> All 12 ported selections green post-fix; the reference's own suite
> green in the venv both before and after (the authority); the
> battery 2394 + 89 at the close (2382 + 12).
