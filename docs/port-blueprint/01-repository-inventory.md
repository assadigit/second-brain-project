# 01 — Repository Inventory

> Evidence: `git ls-files` at v0.91.0 (417 tracked files); LOC by `wc -l`. All counts measured in
> this session. Line references are `repo-relative:path:line` at HEAD `8fd4179`.

## 1. Top-level layout

| Path | What it is | Size |
|---|---|---|
| `app/` | **The application** (migration target) | 123 production .py files, 75,605 LOC |
| `app/cloudflare-bot/` | Telegram bot Worker (JS) + 2 Next.js apps — **not migrated** (ADR-0005) | 8,243 LOC JS src + web apps |
| `docs/` | Archived session history, phase reports, design tokens | 18 files |
| `Port-investigation.md` | This mission's specification | 746 lines |
| `README.md`, `app/README.md` | Owner-facing manual (**stale: claims v0.26.0/837 tests**; reality 0.91.0/3192) | 339 / 572 lines |
| `AGENTS.md` | Agent briefing: test run line, architecture map, house rules | 421 lines |
| `SPEC.md` | Original product spec (six phases, non-negotiables) | 386 lines |
| `CHANGELOG.md` | Plain-language version history | 10,488 lines (~170 KB+) |
| `SWOT-ANALYSIS.md` | Engineering risk picture (v0.26 era) | 218 lines |
| `visual-sync-display-spec.md`, `website-directory-spec.md`, `website-directory-ui-spec (1).md` | Behavioral specs for the Sync view and the Website Directory | 231/308/324 lines |
| `VERSION` | `0.91.0` | — |
| `.github/workflows/ci.yml` | The CI gate (compile 107 modules + 3,192 unittest cases + 217 Worker tests + offline golden run) | 870 lines |

## 2. Production Python module inventory (the parity surface)

All files below are specified in `module-specs/`. Grouped exactly as the specs group them.

### 2.1 Entry point and package root

| File | LOC | Role |
|---|---|---|
| `app/main.py` | 50 | Launcher shim: `--cli` dispatch before GUI import; else `gitcurator.gui.app.main` (`main.py:43-50`) |

### 2.2 `app/gitcurator/core/` — pure-stdlib engine (no Qt; 28 files)

| File | LOC | Spec |
|---|---|---|
| `constants.py` | 366 | `spec-core-constants.md` |
| `links.py` | 798 | `spec-core-links.md` |
| `storage.py` | 182 | `spec-core-storage.md` |
| `dryrun.py` | 304 | `spec-core-dryrun.md` |
| `prompts.py` | 98 | `spec-core-prompts.md` |
| `note_builder.py` | 400 | `spec-core-note-builder.md` |
| `note_state.py` | 842 | `spec-core-note-state.md` |
| `note_reads.py` | 663 | `spec-core-note-reads.md` |
| `taxonomy.py` | 342 | `spec-core-taxonomy.md` |
| `user_config.py` | 481 | `spec-core-user-config.md` |
| `llm_client.py` | 1,459 | `spec-core-llm-client.md` |
| `web_fetch.py` | 1,888 | `spec-core-web-fetch.md` |
| `web_extract.py` | 227 | `spec-core-web-extract.md` |
| `chrome_tabs.py` | 2,583 | `spec-core-chrome-tabs.md` |
| `hand_delivery.py` | 919 | `spec-core-hand-delivery.md` |
| `website_pipeline.py` | 5,338 | `spec-core-website-pipeline.md` |
| `website_state.py` | 755 | `spec-core-website-state.md` |
| `website_directory.py` | 432 | `spec-core-website-directory.md` |
| `vault_scan.py` | 3,353 | `spec-core-vault-scan.md` |
| `vault_hygiene.py` | 812 | `spec-core-vault-hygiene.md` |
| `repo_banish.py` | 659 | `spec-core-repo-banish.md` |
| `sync_review.py` | 968 | `spec-core-sync-review.md` |
| `mirror.py` | 701 | `spec-core-mirror.md` (with `mirror_keys.py`) |
| `mirror_keys.py` | 26 | ″ |
| `linking.py` | 624 | `spec-core-linking.md` (Phase-6 group) |
| `embeddings.py` | 393 | ″ |
| `recall.py` | 386 | ″ |
| `connection_check.py` | 789 | `spec-core-connection.md` |
| `net_guard.py` | 450 | ″ (with `netctx.py`) |
| `netctx.py` | 43 | ″ |
| `__init__.py` | — | empty marker |

### 2.3 `app/gitcurator/gui/` — PyQt6 layer (41 files)

| File | LOC | Spec |
|---|---|---|
| `app.py` | 474 | `spec-gui-app-shell.md` (facade + `main()` + single-instance lock) |
| `headless.py` | 282 | ″ |
| `log_bridge.py` | 84 | ″ |
| `telegram_lazy.py` | 63 | ″ |
| `cache_db.py` | 876 | `spec-gui-cache-db.md` |
| `link_tracker.py` | 692 | ″ |
| `vault_index.py` | 294 | ″ |
| `link_helpers.py` | 62 | ″ |
| `dead_links.py` | 70 | ″ |
| `theme.py` | 681 | `spec-gui-theme-icons.md` |
| `icons.py` | 178 | ″ |
| `sound.py` | 168 | ″ |
| `constants.py` | 258 | ″ (stale duplicate of core constants — test-only importers) |
| `dialogs.py` | 406 | `spec-gui-dialogs.md` |
| `scan_plan_dialog.py` | 930 | ″ |
| `scan_result_dialog.py` | 618 | ″ |
| `vault_pick_dialog.py` | 420 | ″ |
| `processing_worker.py` | 2,113 | `spec-gui-processing.md` |
| `worker_jobs.py` | 1,287 | `spec-gui-worker-jobs.md` |
| `platform_intake.py` | 442 | `spec-gui-intake.md` |
| `telegram_lock.py` | 181 | ″ |
| `sync_display.py` | 679 | `spec-gui-sync-views.md` |
| `main_window/window.py` | 756 | `spec-gui-window.md` |
| `main_window/lifecycle.py` | 282 | ″ |
| `main_window/ui.py` | 3,032 | `spec-gui-ui.md` |
| `main_window/log_view.py` | 142 | `spec-gui-ui.md` (the LogView list widget — head-trim, tooltips, resize debounce) |
| `main_window/processing_control.py` | 2,746 | `spec-gui-processing-control.md` |
| `main_window/theme.py` | 528 | `spec-gui-theme-icons.md` (router mixin) |
| `main_window/hero.py` | 489 | `spec-gui-pages-a.md` |
| `main_window/dashboard.py` | 838 | ″ |
| `main_window/bot_queue.py` | 1,586 | ″ |
| `main_window/telegram_ui.py` | 217 | ″ |
| `main_window/connection_tests.py` | 568 | ″ |
| `main_window/test_connection_modal.py` | 267 | ″ |
| `main_window/vaults_config.py` | 539 | ″ |
| `main_window/input_proxy.py` | 154 | ″ |
| `main_window/llamacpp.py` | 687 | ″ |
| `main_window/linking_tools.py` | 124 | ″ |
| `main_window/vault_scan_ui.py` | 332 | ″ |
| `main_window/backup_seal.py` | 1,477 | `spec-gui-pages-b.md` |
| `main_window/sync_view.py` | 1,453 | `spec-gui-sync-views.md` |
| `main_window/sync_flow.py` | 558 | ″ |
| `main_window/__init__.py`, `worker/__init__.py` | — | markers |
| `worker/llm.py` | 720 | `spec-gui-processing.md` (with the other worker mixins) |
| `worker/github_meta.py` | 321 | ″ |
| `worker/inputs.py` | 232 | ″ |
| `worker/notes.py` | 198 | ″ |
| `worker/reports.py` | 634 | ″ |
| `worker/website_phase.py` | 515 | ″ |

### 2.4 `app/gitcurator/integrations/`, `cloud/`, `cli/`, `tools/`

| File | LOC | Spec |
|---|---|---|
| `integrations/telegram_fetch_worker.py` | 741 | `spec-integrations-telegram.md` |
| `integrations/telethon_fetcher.py` | 432 | ″ (legacy in-process path) |
| `integrations/subprocess_runner.py` | 456 | ″ |
| `integrations/vaultseal.py` | 1,301 | `spec-integrations-vaultseal.md` |
| `integrations/goodrepos.py` | 1,125 | `spec-integrations-goodrepos.md` (with repo_radar) |
| `integrations/repo_radar.py` | 661 | ″ |
| `integrations/backfill_manager.py` | 364 | `spec-integrations-misc.md` |
| `integrations/error_reporter.py` | 288 | ″ |
| `integrations/banish_confirm.py` | 131 | ″ |
| `integrations/scan_confirm.py` | 153 | ″ |
| `integrations/subprocess_runner.py` (dup row above) | — | — |
| `cloud/cloudflare_sync.py` | 735 | `spec-cloud-worker-client.md` |
| `cloud/directory_sync.py` | 1,826 | `spec-cloud-directory-sync.md` |
| `cloud/cloudflare_manager.py` | 605 | `spec-cloud-dormant.md` (dormant) |
| `cloud/cloudflare_gui.py` | 531 | ″ (dormant) |
| `cloud/gdrive_backup.py` | 736 | ″ (dormant) |
| `cloud/gdrive_gui.py` | 612 | ″ (dormant) |
| `cli/cli.py` | 1,777 | `spec-cli.md` |
| `cli/cli_run.py` | 719 | ″ |
| `cli/cli_sync.py` | 381 | ″ |
| `cli/cli_settings.py` | 211 | ″ |
| `cli/cli_terminal.py` | 321 | ″ |
| `cli_run.py` / `cli_settings.py` / `cli_terminal.py` (top-level shims) | 33 total | ″ |
| `tools/build_zip.py` | 285 | `spec-tools.md` |
| `tools/pick_golden_links.py`, `tools/run_golden_websites.py` | ~200+120 | ″ (golden harness) |
| `tools/scan_vault_edits.py`, `tools/snapshot_vault.py` | — | ″ (safety scanners) |
| `tools/backfill_websites.py`, `tools/build_links.py`, `tools/mirror_manual.py`, `tools/add_recall_hooks.py`, `tools/diagnose_code.py`, `tools/integration_snippet.py`, `tools/test.py` | — | ″ (utilities) |

### 2.5 Data/assets shipped in `app/`

- `app/config.example.json` — 109-key config template.
- `app/prompts/*.txt` — 10 prompt files (live: `w01/w02/w03`, `s01/s02` read directly by vault_scan,
  `l01`, `r01`; **dead**: `01/02/03_*` — no production consumer, investigation 1-a).
- `app/taxonomy/website-library-categories.md` (the owner-editable category file),
  `website-directory-config.json`, `search-synonyms.json`.
- `app/assets/fonts/Inter-Bold.ttf`; `app/assets/sounds/success.wav`, `retry.wav`.
- `app/about_me.md`, `app/system_prompt.txt` (GitHub analysis prompt — actually
  `constants.DEFAULT_SYSTEM_PROMPT` is the live one, `constants.py:344-365`).
- `.bat` launchers: `app/1-INSTALL.bat`, `app/GitCurator.bat`, `app/GitCurator-DRY-RUN.bat`,
  plus 5 CLI launchers in `app/gitcurator/cli/` and `gitcurator-cli.sh`.

## 3. External integrations (the outside world)

1. **Cloudflare Worker HTTP API** — pairing, HMAC-signed desktop endpoints, Bearer directory
   endpoints, public index (full contract: `08-data-and-compatibility.md` §5, spec
   `spec-worker-contract.md`).
2. **Telegram MTProto** via Telethon subprocess (bot chat queue, Saved Messages, read-acknowledge).
3. **GitHub REST** — PyGithub (repo metadata/README), raw urllib in goodrepos/vaultseal/repo_radar
   (`Authorization: Bearer`, `Accept: application/vnd.github+json`).
4. **LLM endpoints** — Ollama native API, OpenAI-compatible `/chat/completions`+`/models`,
   Anthropic Messages API, llama.cpp `/props`+`/health`.
5. **Web fetches** — arbitrary sites with Chrome-131 UA ladder; DoH at `cloudflare-dns.com`;
   Wayback at `archive.org`.
6. **Local Chrome DevTools** — raw WebSocket CDP (fifth door).
7. **git CLI** — seals, publishes, radar fingerprints.
8. **PySocks SOCKS5/4 + HTTP proxies** — Telegram and web fetches (DNS at exit).

## 4. Tests, fixtures, golden data

- 83 test modules, 3,192 methods, zero real network (loopback-only exception in `test_synclive`).
- Fixtures: `app/tests/fixtures/web/*.html` (article/landing/latin1/js_shell/paywall),
  `fixtures/website-library-categories.md`.
- Golden: `app/tests/golden/websites.json` (30 links with expected category/subcategory),
  `websites_candidates.json`.
- Worker: 10 `.test.js` files (~217 cases) with `fake-db.mjs`.
- Out-of-gate tools: `tests/timing_runner.py`, `tests/smoke_v007.py`.

## 5. Build/CI/release infrastructure

- `.github/workflows/ci.yml`: Python 3.12 + Qt offscreen; compile 107 modules; run the 84-name
  unittest line; worker `npm test`; offline golden run (`ci.yml:733-870`).
- `tools/build_zip.py`: deterministic zip (fixed timestamps, tracked-files-only, banned-secrets
  guard, launcher flattening, `.bat` CRLF/ASCII law).
- Release ritual: tag → build zip from tag → upload as the GitHub release asset
  (AGENTS.md:375-383).

## 6. Uncertain / anomalous components (findings, not changes)

| Item | Observation | Disposition in blueprint |
|---|---|---|
| `gui/constants.py` | Stale duplicate of `constants.py`; imported only by 5 test files | One constants package in Go; duplicate not ported |
| `prompts/01/02/03_*.txt` | No production consumer (GitHub prompt lives in `constants.py`) | Not ported; noted |
| `WebsiteStateDB.settle_urls` | Dead method (no callers) | Not ported |
| `vault_hygiene.collect_real_note_identities` | Dead (superseded by one-walk) | Not ported |
| `chat_with_timeout` (llm_client) | No production caller | Not ported |
| `checkpoints` table | Created, never used | Not ported; old DBs left untouched |
| Dormant cloud quartet | `cloudflare_manager/cloudflare_gui/gdrive_backup/gdrive_gui` (~2,584 LOC), kept by owner decision "fixed, not deleted" | Not ported in Phase 1; documented in `spec-cloud-dormant.md` with a revive-note |
| Worker secret `HMAC_SECRET` | Set at install time, never read by any Worker code | Not our concern (Worker untouched); documented |
| `dedupe_urls` (links) | Test-only in production | Ported only if tests require |
| Stale READMEs | Root/app README claim v0.26.0/837 tests | Blueprint cites VERSION (0.91.0) |
| `app/cloudflare-bot/dashboard/` | A second dashboard artifact besides the Worker's inline HTML | Out of scope (ADR-0005) |

## 7. Coverage accounting

- **123/123 production Python files** are inventoried above and mapped in
  `05-python-to-go-mapping.md`; each has a home in `module-specs/` (individually or in a justified
  group).
- All entry points: `python main.py` (GUI), `--cli`/`--headless` (CLI/headless),
  `python -m gitcurator.cli`, `tools/*.py` standalone CLIs, `.bat`/`.sh` launchers, the Worker's
  own deploy path (out of scope).
- Tests/assets/config/CI/build all inspected (sections 4–5).
- Non-Python artifacts (Worker, Next.js apps) are documented as **contracts** (ADR-0005) rather
  than ported.
