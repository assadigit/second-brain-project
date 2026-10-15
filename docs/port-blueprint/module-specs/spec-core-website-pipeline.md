# Module Spec — `app/gitcurator/core/website_pipeline.py` (with `website_state.py`, `website_directory.py`)

- **LOC**: 5,338 + 755 + 432 · **Go home**: `internal/websitepipeline` (+ `state.go`,
  `directorynote.go`) · **Status**: port (merged group — the pipeline owns its state and the
  directory-note regeneration)
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

`sweep_banned_notes` (170-272): app-owned notes on banned domains → `.trash/banned-domains/`;
platform `_inbox` tables quarantined when all platform domains banned. `build_website_directory`
(308-432): regenerates `000 📚 Website Directory.md` — frontmatter `managed_by/kind: "directory"/
generated`; `# 📚 Website Directory`; info callout with counts; `## Cat (n)` / `### Sub (n)`;
entries `- [[path|alias ≤80]] — TL;DR ≤140 — pricing — [↗](url)` joined `" — "`, sorted by
`(title.lower(), relpath.lower())`; taxonomy order then leftovers; `_review` todo callout;
closing regeneration line. Read caps 512,000 bytes/note; walk skips `_inbox/_review/_missing/
_moc/attachments/.obsidian/.trash/.git`.

## Tests

State-machine table tests (every gate/outcome); retry arithmetic; dismissal reason grammars;
banishment flows incl. gate verdicts; consolidation idempotence; master-table lifecycle; settle
passes; directory-note golden; pipeline constructor's gate sequence with fake confirm channels.
Port from `test_banishment`, `test_phase*` selections, `test_websitesqueuefix`,
`test_reviewretry`, `test_mastertable`/`test_masterretry`, `test_onespelling`.
