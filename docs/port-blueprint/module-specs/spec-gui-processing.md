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
  duplicates and unparsed reported, ≤20 listed)
  — **Dated note (2026-10-15, ADR-0004 Am.3 / R-0):** `_fetch_from_telegram` (the MTProto
  fetch) falls away in the Go port — the queue source is the Worker `/api/pending` +
  `get_pending_result` path; the non-GitHub inbox-note creation and `_fetch_from_import`
  (import grammar) stay and are ported verbatim.
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
  (`banish_confirm` (Telegram channel) is **Worker-mediated** —
  `make_telegram_confirm` → `/api/banish/propose` trio, server-side Telegram ask — and is
  **unaffected** by ADR-0004 Am.3/Option B; verified in the Python reference, R-0.)
  `split_dismissed_links` pre-drop; `state.due_retries()`; **bar re-base**
  `total = github_done + len(links) + len(due)` with `_wp_advance` closure; backlog modes
  replace normal passes; manifest marking per outcome; `build_website_directory` rebuild;
  `prune_inbox_tables`; summary dict.

## Go mapping

- One `internal/batch.Runner` with the same per-item order, gates, delays, banner goroutine,
  and completion chain; events emitted via the bus; ask-gates via channels with the same
  timeouts; headless fail-fast doors preserved for CLI.

**Amended at T-42 implementation time (reality-wins, in the port commit):**
the Runner lives in `internal/githubpipeline`, not `internal/batch` —
T-25's LinkTracker + VaultIndex already live in `internal/batch`, and a
batch-side Runner importing the GitHub loop would form an import cycle
(githubpipeline needs batch.LinkTracker/batch.VaultIndex). The layering
is githubpipeline → batch (one direction). `internal/batch` gains the
task's "GitHub side": `summary.go` (the v0.39.0 scorecard,
`_build_batch_summary`) — extend, never duplicate.

## SESSION-017 shipped record (T-42, 2026-10-17)

- **internal/githubpipeline**: `ghclient.go` (the RepoFetcher seam +
  the production REST client), `runner.go` (the per-repo loop with the
  gates in order, the get_repo error paths incl. the 403+rate 3605 s
  wait and the 401-anonymous fallback, the two-layer dedupe,
  fork→parent, the enrichment, the banner goroutine joined 15 s, the
  model re-read, the review/quality routing, the disk-full protocol,
  the bookkeeping chain, the delays), `meta.go` (org reputation,
  credibility, the _missing placeholder, the backfill), `banner.go`
  (throttle/429 ladder/24 h marker/atomic write, the netctx verify_ssl
  split), `llmanalyze.go` (the 3-attempt ladder, the README injection
  guard, the per-provider routing, the failure decision door,
  PickBestModel/ApplyModelChoice/LooksLikeConnectionError),
  `reports.go` (the summary log + rotation, the final report),
  `undo.go` (the snapshot, the diff, `_undo_last_batch.txt`, the
  Finish completion chain).
- **internal/batch/summary.go**: the scorecard.
- **THE GOLDEN**: `testdata/githubpipeline.golden.json` (12 sections)
  generated by `testdiff/goldens/dump_githubpipeline.py` from the
  unmodified reference — the reference's GUI-layer imports (PyQt6 /
  PyGithub / ollama) are satisfied by minimal sys.modules stubs (the
  pinned venv is stdlib-only; the reference code is unmodified).
  The reports' file formats byte-compare; the prompt battery pins the
  injection guard + about_me context; the ladder, credibility,
  reputation, and model-picker batteries replay.
- **Reality-wins fixes this task forced**:
  1. notebuilder's `cred_score` became a FLOAT (`round(x, 1)` — every
     real note renders "90.0/100"); the notebuilder golden was
     regenerated with the production calling convention (float values,
     incl. a 55.5 fraction + a 12500-stars comma-grouping case).
  2. The empty-attempt-2 comment in llm.py lies about its own code —
     the raise escapes to the OUTER failure door (the port follows the
     code; see the package doc.go's divergence record).
  3. The GitHub-level error branch marks the MANIFEST only (no
     retry-queue row) — the reference's per-URL except owns add_failed.
- **Deliberately deferred** (the GUI shell's later task): the worker
  shell (`run`/`_run_impl`'s mode wiring, the Ollama warmup family,
  the inputs fetchers, the interactive doors' GUI halves), the
  websites-phase wiring (website_phase.py — the Runner exposes the
  WebsitePhase seam + the WebsiteSummary shape), and the hygiene pass
  (the HygienePass seam; core/vault_hygiene is T-46).

## Tests

Batch summary goldens (fake repo + fake LLM + fake banner); rate-limit law (fake clock);
404-strike flow; 401-anonymous fallback; injection-guard prompt snapshot; 3-attempt ladder;
reports file formats; undo diff correctness; websites-phase wiring (pipeline constructor args
asserted).
