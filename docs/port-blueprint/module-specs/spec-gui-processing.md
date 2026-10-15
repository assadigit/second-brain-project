# Module Spec — `gui/processing_worker.py` + `gui/worker/*` (the batch machinery)

- **LOC**: 2,113 + 2,725 (7 mixin files) · **Go home**: `internal/batch` +
  `internal/githubpipeline` · **Status**: port (logic) / rehost (threading → goroutines/events)
- **Responsibility**: the batch QThread shell + the full GitHub per-repo pipeline + the websites
  phase wiring + LLM interaction handling + reports + the generic TestWorker job runner.

## ProcessingWorker (shell)

- Signals (101-112): `progress_updated(int,int)`, `status_updated(str)`, `log_message(str,str)`,
  `finished_signal(bool,str)`, `code_requested(str)` ("CODE"/"PASSWORD"), `disk_full_signal(str)`,
  `llm_failed_signal(str)`, `model_changed(str,str)`.
- `run` (199-237): exception-proof wrapper + dry-run enable/disable + **finished_signal always
  emitted exactly once** (lock-leak fix).
- Interactive doors (1959-2113): `threading.Event` round-trips — `request_code` (300 s),
  `resolve_llm_failure` (600 s), `request_scan_confirm` (300 s, verdict + selection);
  headless=True → fail-fast defaults.
- `TestWorker(fn, name, *args)`: runs `self._fn(*args)` in a BaseException-safe wrapper,
  **always** emits `finished_signal(name, result|{'success':False,'error':…})` — the heal-job
  pattern's engine.

## The GitHub per-repo loop (`_run_impl` 239-1889; order is the contract)

position++ → emit status+progress → manifest `mark_processing` → rate-limit check (pause when
<50 remaining until exact reset time, chunked 5 s sleeps; 403+rate → wait 3,605 s; 401 → drop
token → anonymous retry) → `clean_url` + canonicalize via `links.extract_github_urls` → gates in
order: non-GitHub → dead-set → dismissed (note_state) → settled ledger → invalid → **vault-index
dedupe BEFORE any API call** (v0.38.0) → `g.get_repo` (404 → `record_404` + `_record_missing_repo`
note + notfound-links.md append; other errors → failed) → success → `reset_dead_links` → cache
dedupe (note-valid? skip : remove_entry reprocess) → fork→parent → stars/forks/description/topics
→ org reputation → 90-day commits → README (≤1,500 chars) → banner inner thread (daemon, joined
15 s) → model re-read from config per URL → `_llm_analyze` → category (confidence <50 →
Uncategorized) → tags merge (≤12) → languages (≤5) → latest release → credibility → folder
(confidence <60 or low-quality → `_review`) → banner join+move → `_build_note` → atomic write
(disk-full protocol) → `cache.add_processed` + log "✅ [Vault] owner/repo processed and stored →
relpath" → `note_state.record_note` → `mark_failed_resolved` → `vault_index.add_url` →
processed++ → manifest mark → sleep `delay_between_api_calls` (0.5 s) + `large_batch_extra_delay`
(1.5 s when total>50).

After the loop: `_run_website_phase` → stores close → batch-undo diff (`_undo_last_batch.txt`) →
progress 100% → LinkTracker verify → `_generate_summary_log` → `_vault_hygiene_pass` →
`_generate_final_report` → `batch_summary` → `finished_signal(True, msg)` + aggregate lines.

## Worker mixins

- **llm.py** (720): `_pick_best_model` (deterministic stand-in: drop embedders → same family →
  same size → biggest); `resolve_llm_failure` + `_wait_for_llm_decision` (600 s; CLI console
  menu hook); `_apply_model_choice` (mutates config in place + emits `model_changed`);
  `_autostart_ollama` (16 × 1.5 s polls); `_call_cloud_llm` (verify_tls from config);
  `_llm_analyze` (439-720): prompt with about_me + README wrapped in
  `<<<README_BEGIN>>>…<<<README_END>>>` injection guard; **3-attempt ladder** (JSON → stricter →
  2-sentence simple fallback with minimal result); per-attempt provider routing with
  `models.analyze` override; failure path: connection-error hint → decision event → stop/retry/
  retry-with-model/skip-fallback.
- **github_meta.py** (321): `_get_org_reputation` (major=10/mid=7 lists);
  `_download_banner` (`https://opengraph.githubassets.com/1/{owner}/{repo}` →
  `<vault>/attachments/banners/{safe}_banner.png`; throttles every 10→2 s / 50→5 s; 24 h
  failed-marker cache; 3 retries 5/10/15 s on 429; atomic write; TLS context from netctx);
  `_calculate_credibility` (0.4·org + 0.25·stars + 0.2·commits + 0.15·doc, ×10, famous-org
  floor 90); `_backfill_missing_notes`.
- **inputs.py** (232): `_fetch_from_telegram` (session .bak; subprocess config; non-GitHub →
  `_create_inbox_notes` + `_non_github_urls`); `_fetch_from_import` (the one link grammar;
  duplicates and unparsed reported, ≤20 listed).
- **notes.py** (198): `_build_note` (delegates to note_builder); `_record_missing_repo`
  (`_missing/` placeholder + immediate `confirm_dead`, idempotent set); `_create_inbox_notes`
  (bypassed when websites pipeline ON; banned never collected).
- **reports.py** (634): `_reports_dir` (`app/reports/` + config override);
  `_vault_hygiene_pass` (delegates); `_generate_final_report`
  (`_processing_report_<ts>.md`: summary table, note-state section, verification table,
  per-category, all-processed, failed, non-GitHub, websites section);
  `_generate_summary_log` (`processing_summary_<ts>.txt` + `summary_keep_last` rotation,
  default 10, 0 = keep all).
- **website_phase.py** (515): constructs `WebsitePipeline` with injected `llm_call` router
  (`models.classify/analyze`), `fetch_fn`, `banish_confirm` (Telegram channel),
  `vault_index_has`, `site_note_for`; blocked/self intake drops; shadow WebsiteStateDB in
  dry-run; `settle_the_ledger`; websites VaultIndex with the website normalizer;
  `split_dismissed_links` pre-drop; `state.due_retries()`; **bar re-base**
  `total = github_done + len(links) + len(due)` with `_wp_advance` closure; backlog modes
  replace normal passes; manifest marking per outcome; `build_website_directory` rebuild;
  `prune_inbox_tables`; summary dict.

## Go mapping

- One `internal/batch.Runner` with the same per-item order, gates, delays, banner goroutine,
  and completion chain; events emitted via the bus; ask-gates via channels with the same
  timeouts; headless fail-fast doors preserved for CLI.

## Tests

Batch summary goldens (fake repo + fake LLM + fake banner); rate-limit law (fake clock);
404-strike flow; 401-anonymous fallback; injection-guard prompt snapshot; 3-attempt ladder;
reports file formats; undo diff correctness; websites-phase wiring (pipeline constructor args
asserted).
