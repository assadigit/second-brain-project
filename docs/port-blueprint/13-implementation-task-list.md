# 13 — Implementation Task List

> Ordered, actionable, traceable. Effort: XS/S/M/L/XL (relative). Each task cites its spec.
> The next session executes these in order; phases map to `12-migration-roadmap.md`.

## Phase 0 — Foundations and PoCs

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-00 | Set up `go-desktop/` module skeleton (`go.mod`, `cmd/`, `internal/`, `assets/`, CI stub workflow); install Go; verify Python venv runs the existing suite | 04 §1; 10 §1 | S | `go build ./...` green; Python smoke (`tests.test_core`) green |
| T-01 | PoC A: gotd/td login + bot-chat iteration + read-acknowledge over SOCKS | ADR-0004 | M | 100-message fetch on the owner's account OR documented fallback decision recorded as ADR-0004 update |
| T-02 | PoC B: tls-client Chrome profile vs stdlib on walled pages; verify license + SOCKS support | ADR-0006; 06 §1 | S | verdict recorded; library pinned or utls fallback chosen |
| T-03 | Differential harness skeleton: case dirs, Python driver (runs unmodified Python app), normalizing comparator | 09 §3 | M | one end-to-end case (e.g. `--version`, import grammar) produces a signed-off diff report |
| T-04 | Performance baselines: startup, scan inventory, log flood, queue classification (Python, offscreen) | 07 §B.4 | S | numbers recorded in the Phase-0 report (no invention) |

## Phase 1 — Pure foundations (byte-parity)

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-10 | `internal/constants`: APP_DIR resolution, stamps, CONFIG_EXAMPLE defaults, CATEGORY_FOLDERS, law lists, EXPECTED_WORKER_VERSION | spec-core-constants | S | golden test vs Python dict dump |
| T-11 | `internal/links`: both normalizers, THE LAW lists, platform classify, secret scrub, import parser (bare domains, comments, dupes) | spec-core-links | L | ported `test_core` grammar cases + Unicode edge fixtures; differential import-file case |
| T-12 | `internal/storage` + `internal/dryrun`: atomic writes, safe_filename, unique_path, merge_config, write_config_file (no HTML escape, 2-space indent); dryrun recorder + shadow cache (incl. WAL sidecars) | spec-core-storage; spec-core-dryrun | M | atomicity + traversal tests; config round-trip byte-stable vs Python on a real config copy |
| T-13 | `internal/prompts` + embedded prompt files; slot refusal laws | spec-core-prompts | XS | ported slot tests; note dead 01/02/03 prompts not shipped |
| T-14 | `internal/taxonomy` parser + folder rules | spec-core-taxonomy | M | parser grammar tests incl. emoji/dash/`no subcategories` cases |
| T-15 | `internal/notebuilder`: sanitizers, yamlScalar, build_note, missing-repo note (byte-exact) | spec-core-note-builder | M | note-format goldens generated from Python for 6 sample repos |
| T-16 | `internal/config`: load/merge/save + user memory (fill/mirror/unset grammar incl. numeric-0) + Known-Folder resolution | spec-core-user-config; 08 §1 | M | ported `test_userconfig` cases; `Documents/.gitcurator` interop test |
| T-17 | Rune/rounding/ordering helper library (`cutRunes`, `rstripChars`, `roundHalfEven`, ordered-dedupe, `normcase`) + tests | 05 §3 | S | table tests incl. U+0130, U+2026, 🗑️ bytes |

## Phase 2 — Stores and state

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-20 | `internal/store`: modernc DSN pragmas (WAL, busy_timeout, synchronous), per-store mutex, naive/UTC timestamp helpers | ADR-0003 | M | open a Python-created cache.db fixture; WAL mode asserted |
| T-21 | `internal/notestate`: scan_vault, classify/apply corrections, NoteStateDB | spec-core-note-state | L | moves-as-corrections suite ported; fingerprint parity (sha256 with recall block stripped) |
| T-22 | `internal/notereads`: RAM + SQLite layers, parse-version 4, banish grammar, stars-fix documented | spec-core-note-reads | M | one-open law tests; DB-hit rows now carry `stars` (documented fix) |
| T-23 | `internal/websitepipeline/state.go`: WebsiteStateDB + retry/backoff | spec-core-website-state | M | backoff arithmetic tests (2×attempts days); settle/heal passes |
| T-24 | `internal/store/repos.go`: CacheDB (processed/failed/decommissioned/settled/meta; skip dead `checkpoints`) | spec-gui-cache-db | M | behavior tests; `record_404` fail-closed test |
| T-25 | `internal/batch/linktracker.go` + `vaultindex.go`: manifest accounting, all-clear law, site map | spec-gui-cache-db §link_tracker/vault_index | L | accounting tests (X/Y accounted); site-note registration gates |

## Phase 3 — Network & LLM

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-30 | `internal/llmclient`: Ollama HTTP, OpenAI-compatible, Anthropic, JSON-mode memoized retry, `extract_json`, detection probes, llama.cpp /props grab, autodetect policy | spec-core-llm-client | L | fake-server payload goldens; `test_detectset`-equivalent; `test_llamacpp`-equivalent |
| T-31 | `internal/webfetch`: routes (proxy/direct/loopback), redirect ladder, headers, rate limiter, caps, decompression, failure categories, DoH, wayback, impersonation door | spec-core-web-fetch; ADR-0006 | XL | ladder tests incl. redirect/cookie/Referer parity; fake SOCKS server asserts domain passthrough; impersonation trigger matrix |
| T-32 | `internal/webextract`: parser + heuristics | spec-core-web-extract | S | 5 HTML fixtures byte-compared fields |
| T-33 | `internal/connectioncheck` + sentinel: battery shapes, UA table, offline verdict | spec-core-connection | M | result-shape tests; UA table asserted |
| T-34 | `internal/workerclient`: pairing, HMAC, all endpoints, 1010 UA, proxy bypass, honest-poll | spec-cloud-worker-client; spec-worker-contract | M | **loopback rig ported from `test_synclive`** (real 127.0.0.1 server verifying signatures) |

## Phase 4 — Pipelines & integrations

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-40 | `internal/websitepipeline`: per-link state machine, gates, consolidation, review notes, master table, banishment + gate, tally | spec-core-website-pipeline | XL | state-machine table tests; banishment flows (confirmed/declined/timeout); differential golden-set case |
| T-41 | `internal/websitepipeline/directorynote.go`: Website Directory regeneration | spec-core-website-directory | M | output golden byte-compare |
| T-42 | `internal/githubpipeline` + `internal/batch`: per-repo loop, banner thread, reports, undo file | spec-gui-processing | L | batch summary goldens; 404-strike tests; rate-limit law tests |
| T-43 | `internal/handdelivery`: queue/consume/harvest/stamps | spec-core-hand-delivery | M | ported `test_handdelivery`/`test_handharvest` cases |
| T-44 | `internal/chrometabs`: WS+CDP client, sessions, attach/launch, wave engine, page_is_real, delivery | spec-core-chrome-tabs | XL | protocol unit tests + manual rig vs local Chromium; wave/stop tests |
| T-45 | `internal/vaultscan`: radar passes, weighing, chunked LLM, validation, apply, census | spec-core-vault-scan | XL | ported `test_vaultscan`/`test_weighhomes`/`test_onespelling` selections; plan-validate table tests |
| T-46 | `internal/vaulthygiene` + `internal/repobanish` | spec-core-vault-hygiene; spec-core-repo-banish | M/L | fingerprint tests; repo banish gate tests |
| T-47 | `internal/vaultseal`: ladder, creation guard, belts, FETCH_HEAD redaction, resync | spec-integrations-vaultseal; ADR-0009 | L | real-git integration tests on temp repos (incl. divergence cases) |
| T-48 | `internal/goodrepos` + `reporadar` | spec-integrations-goodrepos | M | README/publish tests with fake GitHub; radar dominance tests |
| T-49 | `internal/directorysync`: build_index, publish, commands cycle, config trio | spec-cloud-directory-sync | L | index golden; command-cycle tests; `test_directorysync` ports |
| T-50 | `internal/mirror` (+ linking/embeddings/recall) | spec-core-mirror; spec-core-linking | M | block-surgery idempotence; containment guards; suggestion lifecycle |
| T-51 | `internal/telegram`: gotd client per PoC verdict | spec-integrations-telegram; ADR-0004 | L | queue-fetch contract tests (fake peers); login flow tests |

## Phase 5 — CLI parity

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-60 | `internal/cli`: full flag surface, wizard, status/test/detect/hand/chrome/retry commands | spec-cli | L | flag-surface tests; output snapshots |
| T-61 | Terminal paint: ANSI, UTF-8 console hardening (65001), NO_COLOR, spinner/bar | spec-cli §terminal | S | Windows console test |
| T-62 | `--sync` door + `--auto` engine + dry-run report writer | spec-cli; spec-gui-processing | M | differential `--sync --dry-run` case vs Python |
| T-63 | `cmd/gitcurator-cli` + launcher `.bat`s (ASCII/CRLF) + `1-INSTALL` equivalent checks | 10 §2-6 | S | clean-VM smoke passes |
| T-64 | Differential corpus full run + fix loop | 09 §3 | M | all cases green (normalized); report archived |
| T-65 | **Tag `v0.92.0-go.1` + zip artifact** | 10 §8 | XS | GitHub release with asset |

## Phase 6 — GUI

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-70 | PoC C: Wails themed window + log desk + one job | ADR-0001 | M | startup ≤1.5 s; 10k-line flood interactive; verdict recorded |
| T-71 | `internal/app`: commands + event bus + ask-gates + lifecycle/locks | 04 §2/§5 | L | command contract tests; lock/watchdog tests |
| T-72 | Frontend foundation: theme tokens CSS, icons, fonts, layout shell, log desk w/ coalescing | spec-gui-theme-icons; spec-gui-ui | L | token golden (CSS var dump); log perf test |
| T-73 | Settings pages A: Vault, Credentials, Proxy, LLM, Input, Sound | spec-gui-pages-a | L | page workflow tests; save-law (merge/in-place) tests |
| T-74 | Settings pages B: Bot, Dashboard, Sources, Backup (+ radar, fast sync) | spec-gui-pages-b | L | queue report rendering; fast-sync laws (ported pure halves already tested) |
| T-75 | Sync view: flow cards + matrix + differences modal | spec-gui-sync-views | L | four-state rendering tests; deferred-ledger UX tests |
| T-76 | Scan flow: picker → plan modal (masters, arm-guard) → scorecard | spec-gui-dialogs; spec-gui-pages-a(vault_scan_ui) | M | workflow tests; selection semantics |
| T-77 | Main window chrome: hero button machine, progress row, More menu, sentinel labels, dialogs polish, sounds | spec-gui-ui; spec-gui-window | L | hero state machine tests; startup cadence test |
| T-78 | Side-by-side verification vs Python app (lock mutual exclusion, same data) | ADR-0008 | S | lock test across both binaries |

## Phase 7 — Packaging & release

| ID | Task | Spec refs | Effort | Acceptance |
|---|---|---|---|---|
| T-80 | Deterministic zip builder + version stamping + manifest/icon | 10 §2/§5 | M | rebuild byte-identical (sha256 twice) |
| T-81 | Release workflow (tag → build → attach zip) | 10 §8 | S | dry-run on a test tag |
| T-82 | Clean-machine checklist execution + AV scan note | 10 §7/§9 | M | signed checklist in the release report |
| T-83 | Docs: Go README, QUICKSTART update, parity report | 15 | S | owner-reviewable summary |

## Cross-cutting standing orders (apply to every task)

1. Port the corresponding Python tests **first** (characterization), then the code.
2. Any deviation from a spec → record in the implementation report + update the spec file in the
   same commit (spec files live in-repo for exactly this reason).
3. No new dependencies beyond doc 06 §2 without a new ADR.
4. Never weaken/skip a test to get green (repo house rule).
5. Secrets: never in code, logs, reports, or the new repo tree.
