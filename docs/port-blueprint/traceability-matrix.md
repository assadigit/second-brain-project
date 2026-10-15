# Traceability Matrix

> Connects each important existing behavior → source evidence → module spec → planned Go
> component → test category → implementation task → acceptance criterion. Purpose: nothing
> disappears during migration. Format: **B**ehavior rows grouped by workflow (doc
> `03-behavioral-specification.md` W-numbers).

Legend — Evidence `app/…:line` at v0.91.0 · Spec `module-specs/…` · Go `internal/…` · Task
`13-implementation-task-list.md` · AC `15-acceptance-criteria.md`.

## A. Foundations

| # | Behavior | Evidence | Spec | Go | Tests | Task | AC |
|---|---|---|---|---|---|---|---|
| A1 | URL canonicalization (GitHub) incl. twitter→x substring, path-case preserve | `core/links.py:187-223` | spec-core-links | `links` | golden grammar + Unicode | T-11 | AC-13 |
| A2 | Website canonicalization (https-force, www-drop, tracking params, one-spelling) | `core/links.py:590-649` | spec-core-links | `links` | golden | T-11 | AC-13 |
| A3 | THE LAW blocked domains (41) + config-add-only | `core/links.py:298-435` | spec-core-links | `links` | law golden | T-11 | AC-2 |
| A4 | Secret query scrubbing (`…` U+2026) | `core/links.py:547-576` | spec-core-links | `links` | scrub corpus | T-11 | AC-13 |
| A5 | Atomic writes (same-dir temp, fsync, rename, cleanup) | `core/storage.py:37-99` | spec-core-storage | `storage` | torn-write probe | T-12 | AC-12 |
| A6 | safe_filename traversal-proofing | `core/storage.py:106-112` | spec-core-storage | `storage` | traversal corpus | T-12 | AC-13 |
| A7 | Config merge law (nested, unknown-keys survive, tuning defaults) | `core/storage.py:145-177`; `gui/main_window/vaults_config.py:260-281` | spec-core-storage; spec-gui-pages-a | `config` | merge cases + round-trip | T-12/T-16 | AC-10 |
| A8 | Dry-run recorder + shadow cache (WAL sidecars) | `core/dryrun.py:74-258` | spec-core-storage (dryrun §) | `dryrun` | no-write snapshot diff | T-12 | AC-12 |
| A9 | Prompt slot refusal laws | `core/prompts.py:48-80` | spec-core-prompts | `prompts` | refusal strings | T-13 | AC-13 |
| A10 | Taxonomy grammar (emoji/dash/no-subcats/judgment rules) | `core/taxonomy.py:218-325` | spec-core-taxonomy | `taxonomy` | parser suite | T-14 | AC-13 |
| A11 | GitHub note format + YAML sanitizers (security) | `core/note_builder.py:200-345, 48-193` | spec-core-note-builder | `notebuilder` | goldens + sanitizer battery | T-15 | AC-13 |
| A12 | User memory fill/mirror incl. numeric-0 unset | `core/user_config.py:155-460` | spec-core-user-config | `config` | 83-case port | T-16 | AC-9 |

## B. State engines

| # | Behavior | Evidence | Spec | Go | Tests | Task | AC |
|---|---|---|---|---|---|---|---|
| B1 | Moves-are-corrections state machine | `core/note_state.py:477-842` | spec-core-note-state | `notestate` | branch suite | T-21 | AC-2 |
| B2 | One-open note reads + parse-version re-read | `core/note_reads.py:297-592` | spec-core-note-reads | `notereads` | one-open law | T-22 | AC-2 |
| B3 | Websites retry/backoff (3 max, 2×attempts days) | `core/website_state.py:152-172` | spec-core-website-pipeline (state §) | `websitepipeline/state` | arithmetic tests | T-23 | AC-2 |
| B4 | cache.db CacheDB laws (record_404 fail-closed; settle idempotent) | `gui/cache_db.py:497-519, 739-810` | spec-gui-cache-db | `store` | behavior ports | T-24 | AC-9 |
| B5 | Link manifest full accounting + all-clear law | `gui/link_tracker.py:206-523` | spec-gui-cache-db | `batch/linktracker` | accounting suite | T-25 | AC-2 |
| B6 | VaultIndex gates (skip set, `_review` included, site map) | `gui/vault_index.py:74-193` | spec-gui-cache-db | `batch/vaultindex` | gate tests | T-25 | AC-2 |
| B7 | sync_review four-state model + DeferredLog never-delete | `core/sync_review.py:82-708` | spec-core-sync-review | `syncreview` | 143-case port | T-45 (pure) | AC-2 |

## C. Network & contracts

| # | Behavior | Evidence | Spec | Go | Tests | Task | AC |
|---|---|---|---|---|---|---|---|
| C1 | LLM wire formats (Ollama/OpenAI/Anthropic) + JSON-mode retry | `core/llm_client.py:337-745` | spec-core-llm-client | `llmclient` | payload goldens | T-30 | AC-1 |
| C2 | llama.cpp detection + /props config grab | `core/llm_client.py:770-1459` | spec-core-llm-client | `llmclient` | detect table | T-30 | AC-1 |
| C3 | Web-fetch ladder + failure categories + rate limiter | `core/web_fetch.py:1559-1719` | spec-core-web-fetch | `webfetch` | fake-server ladder | T-31 | AC-2 |
| C4 | Chrome-TLS impersonation trigger matrix | `core/web_fetch.py:1457-1511` | spec-core-web-fetch; ADR-0006 | `webfetch` | trigger tests + PoC | T-02/T-31 | AC-2 |
| C5 | HTML extraction + partial heuristics | `core/web_extract.py:58-227` | spec-core-web-extract | `webextract` | 5 fixtures | T-32 | AC-13 |
| C6 | Worker HMAC + 1010 UA + proxy bypass + honest poll | `cloud/cloudflare_sync.py:59-304` | spec-cloud-worker-client; spec-worker-contract | `workerclient` | **loopback rig** | T-34 | AC-6 |
| C7 | Sentinel + offline warn-once + any-answer-is-online | `core/net_guard.py:94-419` | spec-core-connection | `connectioncheck` | transition tests | T-33 | AC-8 |
| C8 | Test-connection battery shapes + UA table | `core/connection_check.py:47-789` | spec-core-connection | `connectioncheck` | shape tests | T-33 | AC-1 |

## D. Pipelines & integrations

| # | Behavior | Evidence | Spec | Go | Tests | Task | AC |
|---|---|---|---|---|---|---|---|
| D1 | Websites per-link state machine (gates→fetch→classify→analyze→write) | `core/website_pipeline.py:4370-4875` | spec-core-website-pipeline | `websitepipeline` | table tests + golden set | T-40 | AC-2 |
| D2 | Banishment grammar + Telegram gate + tally | `core/website_pipeline.py:2243-3154` | spec-core-website-pipeline | `websitepipeline` | verdict flows | T-40 | AC-2 |
| D3 | One-note-per-site consolidation + site_links rewrite | `core/website_pipeline.py:294-418, 4490-4523` | spec-core-website-pipeline | `websitepipeline` | idempotence | T-40 | AC-2 |
| D4 | Website Directory note regeneration | `core/website_directory.py:308-432` | spec-core-website-pipeline | `websitepipeline/directorynote` | output golden | T-41 | AC-13 |
| D5 | GitHub per-repo loop + rate-limit law + 404 quarantine | `gui/processing_worker.py:947-1719` | spec-gui-processing | `githubpipeline` | batch goldens | T-42 | AC-3 |
| D6 | Hand-delivery door 4 (queue/consume/harvest/stamps) | `core/hand_delivery.py:88-907` | spec-core-hand-delivery | `handdelivery` | ported suites | T-43 | AC-2 |
| D7 | Chrome CDP door 5 (attach/launch/wave/page_is_real) | `core/chrome_tabs.py:484-2563` | spec-core-chrome-tabs | `chrometabs` | protocol + rig | T-44 | AC-2 |
| D8 | Vault scan radar/weighing/validation/apply | `core/vault_scan.py:548-3353` | spec-core-vault-scan | `vaultscan` | ported suites | T-45 | AC-2 |
| D9 | Hygiene fingerprints (shape/name-twin/linked-ghost/lock) | `core/vault_hygiene.py:141-796` | spec-core-vault-hygiene | `vaulthygiene` | fingerprint tests | T-46 | AC-2 |
| D10 | Repo banish (DB-first ordering, gate, revive) | `core/repo_banish.py:110-659` | spec-core-repo-banish | `repobanish` | ported suite | T-46 | AC-2 |
| D11 | VaultSeal ladder + creation guard + belts + token law | `integrations/vaultseal.py:183-1240` | spec-integrations-vaultseal | `vaultseal` | real-git integration | T-47 | AC-5 |
| D12 | GoodRepos publish + repo radar dominance | `integrations/goodrepos.py`; `repo_radar.py` | spec-integrations-goodrepos | `goodrepos` | fake-GitHub tests | T-48 | AC-5 |
| D13 | Directory publish (tree law, belts, tombstones) + command cycle | `cloud/directory_sync.py:820-1822` | spec-cloud-directory-sync | `directorysync` | index golden + cycle | T-49 | AC-5 |
| D14 | Mirror + linking + embeddings + recall blocks | `core/mirror.py` + Phase-6 group | spec-core-mirror | `mirror` | block surgery | T-50 | AC-2 |
| D15 | Telegram queue fetch + mark-read + login (gotd) | `integrations/telegram_fetch_worker.py` | spec-integrations-telegram | `telegram` | PoC + contract tests | T-01/T-51 | AC-7 |

## E. Workflows & UX laws

| # | Behavior | Evidence | Spec | Go | Tests | Task | AC |
|---|---|---|---|---|---|---|---|
| E1 | SYNC hero machine + offline guard + caught-up/drift seal | `gui/main_window/hero.py:107-489` | spec-gui-pages-a | `ui` + `app` | workflow tests | T-77 | AC-8 |
| E2 | Fast Sync laws (always-publish, always-speak, guards) | `gui/main_window/backup_seal.py:64-198, 892-938` | spec-gui-pages-b | `app` (pure) + `ui` | ported drift/fastsync | T-74 | AC-2 |
| E3 | Completion chain order + Phase-5 CLEAR gating | `gui/main_window/processing_control.py:1870-2230` | spec-gui-processing-control | `app` | chain-order test | T-71 | AC-3 |
| E4 | Coalesced log (120 ms/250 rows/1000 cap) | `gui/main_window/ui.py:65-86, 2374-2794` | spec-gui-ui | `ui` + bus | flood perf test | T-72 | AC-16 |
| E5 | Sync view (flow cards + matrix + differences modal + honesty states) | `gui/main_window/sync_flow.py`; `sync_view.py`; `gui/sync_display.py` | spec-gui-sync-views | `ui/sync` + `app` | render tests | T-75 | AC-8 |
| E6 | Scan flow (picker→plan arm-guard→scorecard landed counts) | `gui/scan_plan_dialog.py`; `scan_result_dialog.py`; `vault_pick_dialog.py` | spec-gui-dialogs | `ui/dialogs` | workflow tests | T-76 | AC-8 |
| E7 | Settings pages inventory + save-on-change + status ladders | `gui/main_window/ui.py:179-2100`; `vaults_config.py` | spec-gui-ui; spec-gui-pages-a | `ui/pages` | page snapshot | T-73/T-74 | AC-8 |
| E8 | Theme tokens + icons + sounds + one-primary law | `gui/theme.py:58-237`; `icons.py`; `sound.py` | spec-gui-theme-icons | `ui/theme` | CSS var golden | T-72 | AC-8 |
| E9 | CLI surface + `--sync` door + dry-run report | `cli/*.py` | spec-cli | `cli` | flag snapshot + differential | T-60/62/64 | AC-1 |
| E10 | Startup cadences + quiet-open + shutdown chain | `gui/main_window/window.py:115-451`; `lifecycle.py:173-281` | spec-gui-window | `app/lifecycle` | cadence test | T-71 | AC-8/18 |
| E11 | Single-instance lock (cross-binary) | `gui/app.py:393-440` | spec-gui-app-shell | `app/lifecycle` | cross-binary test | T-78 | AC-11 |
| E12 | Platform intake tables + telegram lock owner-scope | `gui/platform_intake.py`; `telegram_lock.py` | spec-gui-intake | `batch/intake`; `app/tglock` | grammar + lock tests | T-25/T-71 | AC-2 |

## F. Data & packaging

| # | Behavior | Evidence | Spec | Go | Tests | Task | AC |
|---|---|---|---|---|---|---|---|
| F1 | cache.db/sync_review.db open-in-place (same DDL) | all stores (doc 02 §5) | store specs | `store` | fixture DB probes | T-20 | AC-9 |
| F2 | Naive-local vs UTC timestamp shapes | store specs | store specs | `store` | format goldens | T-20 | AC-13 |
| F3 | Backup zip laws (excludes, FIFO, sync_review rider) | `gui/main_window/backup_seal.py:548-640` | spec-gui-pages-b | `app` | zip content tests | T-74 | AC-5 |
| F4 | Deterministic release zip | `tools/build_zip.py:126-217` | spec-tools; doc 10 | `build/mkzip` | double-build sha256 | T-80 | AC-22 |
| F5 | Windows launch contract (ASCII .bat, no console for GUI) | `app/*.bat`; ADR-0010 | spec-cli; doc 10 | `build` | clean-VM checklist | T-63/T-82 | AC-21 |
| F6 | Token redaction everywhere (git URLs, FETCH_HEAD, logs) | `vaultseal.py:606-623`; `goodrepos.py:531-535` | spec-integrations-vaultseal | seals | redaction asserts | T-47 | AC-5 |

## Coverage assertion

Every production module in `01-repository-inventory.md` §2 maps to at least one row above via
its spec; every row cites evidence, a Go home, a test category, a task, and an acceptance
criterion. The implementation session must keep this matrix updated as suites land (doc 09 §6).
