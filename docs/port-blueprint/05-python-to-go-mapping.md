# 05 — Python-to-Go Mapping

> Every production Python module → its Go destination. Status legend: **port** (faithful port),
> **merge** (combined with sibling(s)), **rehost** (logic ported, host replaced — GUI),
> **contract-only** (external system kept, client ported), **drop** (dead/vestigial, with
> evidence). Grouped exactly like `module-specs/`.

## 1. Mapping table

| Python module (LOC) | Go package / file | Status | Parity anchor (tests) |
|---|---|---|---|
| `core/constants.py` (366) | `internal/constants` | port | golden: defaults, CATEGORY_FOLDERS, law lists |
| `core/links.py` (798) | `internal/links` | port | `test_core`-equivalent grammar tests + golden URLs |
| `core/storage.py` (182) | `internal/storage` | port | atomic-write + safe_filename tests |
| `core/dryrun.py` (304) | `internal/dryrun` | port | recorder + shadow-cache tests |
| `core/prompts.py` (98) | `internal/prompts` (+ embedded files) | port | slot/refusal tests |
| `core/note_builder.py` (400) | `internal/notebuilder` | port | YAML sanitizer + note-format goldens |
| `core/note_state.py` (842) | `internal/notestate` | port | moves-as-corrections suite |
| `core/note_reads.py` (663) | `internal/notereads` | port (fix noted) | one-open/fingerprint/version tests |
| `core/taxonomy.py` (342) | `internal/taxonomy` | port | parser grammar tests |
| `core/user_config.py` (481) | `internal/config/usermemory.go` | port | fill/mirror/unset-grammar tests |
| `core/llm_client.py` (1,459) | `internal/llmclient` | port | payload-shape goldens + extract_json |
| `core/web_fetch.py` (1,888) | `internal/webfetch` | port | ladder/failure-category tests (fake server) |
| `core/web_extract.py` (227) | `internal/webextract` | port | fixture-based extraction tests |
| `core/chrome_tabs.py` (2,583) | `internal/chrometabs` | port | CDP protocol unit tests + manual rig |
| `core/hand_delivery.py` (919) | `internal/handdelivery` | port | queue/consume/harvest tests |
| `core/website_pipeline.py` (5,338) | `internal/websitepipeline` (+ `state.go`, `directory.go`) | merge (with website_state, website_directory) | per-link state machine tests |
| `core/website_state.py` (755) | `internal/websitepipeline/state.go` | merge | retry/backoff tests |
| `core/website_directory.py` (432) | `internal/websitepipeline/directorynote.go` | merge | output-format golden |
| `core/vault_scan.py` (3,353) | `internal/vaultscan` | port | radar/weighing/validation tests |
| `core/vault_hygiene.py` (812) | `internal/vaulthygiene` | port | stray-fingerprint tests |
| `core/repo_banish.py` (659) | `internal/repobanish` | port | banish-gate tests |
| `core/sync_review.py` (968) | `internal/syncreview` | port | truth-table + ledger lifecycle tests |
| `core/mirror.py` + `mirror_keys.py` (727) | `internal/mirror` | merge | plan/apply + containment tests |
| `core/linking.py`/`embeddings.py`/`recall.py` (1,403) | `internal/mirror/linking.go` (+ `embeddings.go`, `recall.go`) | merge (Phase-6 group) | block-surgery + suggestion tests |
| `core/connection_check.py`/`net_guard.py`/`netctx.py` (1,282) | `internal/connectioncheck` | merge | battery result-shape tests |
| `gui/app.py` (474) | `cmd/gitcurator/main.go` + `internal/app/lifecycle.go` | rehost (facade → entry) | single-instance lock tests |
| `gui/headless.py` (282) | `internal/cli` (fail-fast doors) | merge | headless timeout tests |
| `gui/log_bridge.py` (84) | `internal/app/events.go` (slog handler) | merge | — |
| `gui/telegram_lazy.py` (63) | (unnecessary — Go imports are cheap) | drop | — |
| `gui/cache_db.py` (876) | `internal/store` + `internal/store/repos.go` | port | DDL + behavior tests |
| `gui/link_tracker.py` (692) | `internal/batch/linktracker.go` | port | manifest accounting tests |
| `gui/vault_index.py` (294) | `internal/batch/vaultindex.go` | port | rebuild/site-map tests |
| `gui/link_helpers.py`/`dead_links.py` (132) | `internal/links` + `internal/config` | merge (delegates) | — |
| `gui/theme.py`/`main_window/theme.py`/`icons.py`/`sound.py`/`gui/constants.py` (1,889) | `ui/frontend/src/theme/` (CSS tokens) + `internal/app/sound.go` | rehost | token-table golden (CSS var dump) |
| `gui/dialogs.py` + `scan_plan_dialog.py` + `scan_result_dialog.py` + `vault_pick_dialog.py` (2,374) | `ui/frontend/src/dialogs/*` | rehost | workflow tests (GUI driver) |
| `gui/main_window/window.py`/`lifecycle.py` (1,038) | `internal/app/lifecycle.go` + `ui/` shell | rehost | startup-sequence tests |
| `gui/main_window/ui.py` (3,032) | `ui/frontend/src/pages/*` + `internal/app` commands | rehost | log-coalescing perf test |
| `gui/processing_worker.py` + `gui/worker/*` (4,783) | `internal/batch` (+ `githubpipeline`) | port (logic) / rehost (threading) | batch summary goldens |
| `gui/worker_jobs.py` (1,287) | `internal/app/commands.go` | port | per-job result-shape tests |
| `gui/main_window/processing_control.py` (2,746) | `internal/app/commands.go` + `ui/` | rehost | completion-chain tests |
| `gui/main_window/{hero,dashboard,bot_queue,telegram_ui,connection_tests,test_connection_modal,vaults_config,input_proxy,llamacpp,linking_tools,vault_scan_ui}.py` (5,767) | `ui/frontend/src/pages/*` + command bindings | rehost | per-page workflow tests |
| `gui/main_window/{backup_seal,sync_view,sync_flow}.py` + `gui/sync_display.py` (3,488) | `internal/app/commands.go` (drift/fast-sync) + `ui/frontend/src/sync/*` | rehost | drift/fast-sync law tests (pure halves ported) |
| `gui/platform_intake.py` (442) | `internal/batch/intake.go` | port | platform-table tests |
| `gui/telegram_lock.py` (181) | `internal/app/tglock.go` | port | lock/watchdog tests |
| `integrations/telegram_fetch_worker.py` + `telethon_fetcher.py` + `subprocess_runner.py` (1,629) | `internal/telegram` (gotd/td; no subprocess) | **replace** (ADR-0004) | queue-fetch contract tests |
| `integrations/vaultseal.py` (1,301) | `internal/vaultseal` | port | ladder tests (real git temp repos) |
| `integrations/goodrepos.py` + `repo_radar.py` (1,786) | `internal/goodrepos` (+ radar.go) | merge | publish/README + fingerprint tests |
| `integrations/{backfill_manager,error_reporter,banish_confirm,scan_confirm}.py` (936) | `internal/workerclient/{backfill,errors,confirm}.go` | merge | protocol tests |
| `cloud/cloudflare_sync.py` (735) | `internal/workerclient` | port | HMAC golden + loopback rig |
| `cloud/directory_sync.py` (1,826) | `internal/directorysync` | port | index golden + command-cycle tests |
| `cloud/{cloudflare_manager,cloudflare_gui,gdrive_backup,gdrive_gui}.py` (2,484) | — | drop (dormant; documented) | — |
| `cli/*.py` + shims (3,596) | `internal/cli` + `cmd/gitcurator-cli` | port | flag-surface + output tests |
| `tools/*.py` (~1,300) | `go-desktop/tools/` (Go cmds) or kept Python | port selectively | see `spec-tools.md` |
| `app/cloudflare-bot/*` (JS) | — | **contract-only** (ADR-0005) | `spec-worker-contract.md` |

**Coverage check**: every file in `01-repository-inventory.md` §2 appears exactly once above
(including the drops, each with evidence).

## 2. Retain / replace / consolidate / redesign summary

- **Retained (port)**: the entire `core/` engine, `cache_db`/`link_tracker`/`vault_index`, the
  integrations' behavior, the CLI surface, the Worker client contract.
- **Replaced (different mechanism, same behavior)**: Telethon→gotd (ADR-0004); curl_cffi→tls-client
  (ADR-0006); QThread/signals→goroutines/event bus; the ThreadPoolExecutor timeout→contexts;
  `os.walk`→`filepath.WalkDir`; `html.parser`→`x/net/html` (with conservative flags to keep the
  same extraction results); Telethon subprocess isolation→goroutine + lock.
- **Consolidated**: website_pipeline+state+directory (one package); mirror+linking+embeddings+recall;
  connection_check+net_guard+netctx; goodrepos+repo_radar; confirm gates into workerclient;
  worker mixins into `internal/batch`.
- **Redesigned internally (behavior preserved)**: the 18-mixin MainWindow → `app` commands + Wails
  UI; theme/icons/sound → CSS tokens + embedded assets; the coalesced log → ring buffer + ticker.
- **Dropped (with evidence)**: dormant cloud quartet, dead code items in inventory §6, the Python
  subprocess protocol itself, `gui/constants.py` duplicate, unprompted prompt files.

## 3. Semantic conversion risks (the compatibility hotspots)

Ordered by blast radius. Each requires a dedicated parity test (see doc 09).

| # | Python semantic | Where | Go hazard | Required behavior |
|---|---|---|---|---|
| 1 | Insertion-ordered dicts (first-wins dedupes, platform first-match, approved-links order) | links, note_state, linking, vault_scan | Go maps unordered | ordered set type (slice + index map) or explicit sort matching Python's observed order |
| 2 | `str.lower()` casing (not casefold) | everywhere folded matching | `strings.ToLower` differs on e.g. U+0130 | use `strings.ToLower` but pin folded-identity tests with Unicode edge fixtures; `specialLower` helper if needed |
| 3 | Code-point slicing `[:N]` | recall 220, excerpt 4000, one_line 300, embed text 1200, head reads 800/4000/16000, log 160 | byte slicing cuts mid-rune | rune-slice helpers `cutRunes(s, n)` |
| 4 | `round(x, 2)` banker's rounding | vault_scan weights | `math.Round` half-away | decimal-aware round-half-even helper; golden weight tests |
| 5 | Naive-local ISO timestamps vs UTC-aware | note_state/website_state (local) vs sync_review (UTC) | `time.Time` always carries zone | format helpers producing the exact legacy strings (`2006-01-02T15:04:05` local; UTC ISO with offset for deferrals) |
| 6 | `os.walk` unsorted iteration | note_state.scan_vault, linking, mirror | `filepath.WalkDir` is lexical | sort or accept + document; duplicate-winner tests pin the choice |
| 7 | `str.rstrip(chars)` character-set semantics | links (trailing punct, all trailing slashes), tags strip | suffix-strip mindset | implement `rstripChars`/`stripChars` helpers |
| 8 | `re` vs RE2 | all regexes (list in specs) | RE2 lacks backrefs/lazy semantics differ slightly | every pattern is RE2-safe (verified in specs); port with `regexp.MustCompile` + `(?m)`/`(?s)` flags as needed; `fullmatch` → `^…$` |
| 9 | JSON `ensure_ascii=False` + Go HTML escaping | config, note_reads, embeddings, queue.json | `json.Marshal` escapes `<>&` | encoder with `SetEscapeHTML(false)`; parity byte-goldens |
| 10 | YAML always-double-quoted scalars | note_builder | hand-rolled YAML | port `yamlScalar` byte-exactly (escape `\` then `"`, caps) |
| 11 | Substring emoji checks (`🗑️` incl. VS16, `…` U+2026) | banish grammar, token scrub | byte-equal `strings.Contains` works | keep exact byte sequences; fixture tests |
| 12 | `f"{n:,}"` comma grouping | note_builder stars, dryrun sizes | no builtin | `humanize.Comma` or local helper — must match exactly incl. negatives (not used) |
| 13 | `os.path.normcase` (Windows lower+slash-fold) | note_state moved test | Go has none | explicit `normcase` helper used only where Python used it |
| 14 | `time.monotonic` deadlines | banish cache 60 s, rate limiter | `time.Since` on a monotonic clock — Go's clock is already monotonic | direct |
| 15 | SQLite naive text dates + `fromisoformat` parsing | due_retries | Go `time.Parse` layout | tolerant parser accepting both `T` and space separators (Python 3.11+ does) |
| 16 | `dict.fromkeys` ordered dedupe | sync_review stage_cell | map loses order | ordered dedupe helper |
| 17 | Exception-driven control flow (`try/except` as gates) | everywhere | error values | keep the same *user-visible* outcomes; internal structure free |
| 18 | Thread leaks by design (LLM timeout executor) | llm_client | contexts cancel for real | behavior superset; no parity test needed beyond timeouts firing |
| 19 | `os.replace` atomic write + no dir fsync | storage | `os.Rename` same semantics on Windows/NTFS | port exactly; no chmod added |
| 20 | Percent-format widths / `strftime` | many | `time.Format` layouts | helpers with Go layouts pinned by goldens |

## 4. Type translation notes (per group)

- `FetchResult` → struct with `Body []byte`, `Text string`, plus `Status/Category/Reason/...`
  mirroring `core/web_fetch.py:331-372` slots.
- Python "result dicts" (batch results, scan plans, seals) → named structs **plus** a
  `map[string]any` escape hatch only where the Python code treated keys dynamically (plan rows
  carry `as`, `low_confidence`, `weighed_homes`… — keep them as struct fields, they are finite and
  enumerated in the specs).
- SQLite rows ↔ structs: 1:1 with the DDL in specs; `Nullable` columns as pointer/`sql.Null*`.
- Config: `map[string]any`-style tree retained intentionally (merge law needs arbitrary keys) —
  typed accessors on top (`cfg.Str("vault_path")`, `cfg.Proxy()`), mirroring Python's tolerant
  `.get` grammar including the **unset-value grammar** (`user_config.is_unset_value` — `0` is
  unset for ints!).
