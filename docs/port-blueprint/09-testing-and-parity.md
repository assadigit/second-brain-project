# 09 — Testing and Behavioral Parity

## 1. What the Python suite already proves (the porting backbone)

- **3,192 hermetic test methods** across 83 modules, zero real network (loopback-only exception:
  `test_synclive`'s real local Worker). They pin: URL grammar, sanitizers, note formats,
  moves-as-corrections, retry/backoff, banishment grammar and gates, scan validation/apply,
  sync-review truth table, log coalescing, fast-sync/drift-seal laws, local-only belt, creation
  guards, one-spelling, hygiene fingerprints, UA/HMAC protocol.
- **Golden corpus**: 30-link websites golden set run offline through the real pipeline
  (`run_golden_websites.py --offline`, part of CI); HTML fixtures (article/landing/latin1/
  js_shell/paywall).
- **Live protocol rig**: `tests/test_synclive.py` — a real HTTP server on 127.0.0.1 speaking the
  Worker protocol byte-for-byte (HMAC incl. 1010-UA rejection, honest-404 index, Bearer publish),
  real vaults on disk, real git seals, outage/restart cases. **This rig is the single most
  valuable test to port early.**

The Go test plan translates this backbone rather than inventing a new one.

## 2. Go test pyramid

| Layer | What | Tooling |
|---|---|---|
| **Unit (golden)** | links grammar, storage atomicity, sanitizers, taxonomy parser, prompt loader, timestamp helpers, casing/folding helpers | `go test` + `testdata/` goldens (ported fixtures) |
| **State machines** | websites per-link lifecycle, note_state classification, retry queue, sync_review cells/ledger, scan plan validation, hygiene fingerprints | table-driven tests mirroring the Python case lists |
| **Store tests** | cache.db schemas + behaviors (opening Python-created DB files) | `modernc` in-memory/file DBs + fixture DB dumps |
| **Contract tests (fakes)** | Ollama/OpenAI/Anthropic/llama.cpp payload shapes; GitHub REST; Worker HMAC endpoints; redirect chains; SOCKS dialer | `httptest.Server` fakes ported from the Python fakes |
| **Integration (real binaries)** | vaultseal ladder + goodrepos + radar with real `git` in temp dirs; chrome_tabs against a local Chromium if available | `go test -tags integration` |
| **End-to-end batch** | fake fetch + fake LLM → real vault temp dir → compare notes/DB/manifests with Python outputs on the same inputs | the **differential harness** (§5) |
| **GUI workflow** | Wails-driven page/dialog flows (scan picker → plan → apply; sync view refresh; log flood) | headless webview driver; keep minimal, logic already covered below the UI |
| **Performance** | startup, scan inventory, log flood, memory | benchmarks + CI-reported numbers |
| **Packaging smoke** | run the built exe on a clean Windows VM: cold start, dry-run batch, zip contents | manual checklist + scripted assertions |

## 3. Differential (Python-vs-Go) testing

**Harness design** (to be built in Phase 1, task T-03):

1. A `testdata/cases/` directory of input sets: URLs + fake fetch bodies + fake LLM answers + a
   config.json + an empty vault tree.
2. A Python driver (kept inside the *existing* repo's test area? **No** — the harness lives in
   `go-desktop/testdiff/` and shells out to `python main.py --cli --auto --yes --dry-run`-style
   runs; Python stays unmodified per the safety rules — the driver only *runs* the app, it does
   not change it).
3. A Go driver running the same commands on the same inputs.
4. A normalizing comparator:
   - **must match exactly**: note file bytes (modulo the timestamp fields below), DB table
     contents (modulo timestamps), manifest structure, log verdict lines, exit codes.
   - **normalized**: `date_processed`/`processed_at`/`generated_at`/`settled_at`-style fields
     (regex → `<TS>`), temp paths (→ `<TMP>`), durations, version strings in UA-bearing logs.
   - **documented acceptable differences**: OS-walk ordering effects where the Python behavior was
     itself unsorted (see mapping §3 item 6) — pinned per case with the winner noted.
5. The comparator's diff report is the parity evidence for acceptance (doc 15).

**Scope of differential cases** (minimum): 30-link golden set; the 5 HTML fixtures; import-file
grammar corpus; moves-as-corrections scenarios (move/edit/delete/duplicate); banishment with
declined/confirmed/timeout verdicts; scan plan apply with picked subsets; retry backoff clock
(faked `now`); dry-run report equality.

## 4. Characterization tests to write FIRST (before porting each engine)

For each core package, port the corresponding Python test module's *case list* as Go table tests
(mechanical translation, no reinterpretation). Mapping of the highest-value suites:

| Python suite (cases) | Go package under test |
|---|---|
| `test_core` (34) | links, storage, note_builder sanitizers |
| `test_vaultscan` (91) | vaultscan radar/weighing/validation |
| `test_banishment` (70) | websitepipeline banish machinery |
| `test_syncreview` (143) | syncreview truth table + ledger |
| `test_syncdisplay` (65) + `test_fastsync` (27) + `test_driftseal` (24) | app commands (pure halves) |
| `test_synclive` (30) | workerclient + directorysync loopback rig |
| `test_directorysync` (63) | directorysync build_index/publish |
| `test_localonly` (27) | vaultseal belt builders |
| `test_sealfix` (43) + `test_resyncfix` (17) | vaultseal ladder (real git) |
| `test_notereads` (16) + `test_noteredo` (48) | notereads cache laws |
| `test_userconfig` (83) | config/user memory |
| `test_llamacpp` (93) + `test_detectset` (47) | llmclient detection |
| `test_webproxy` (45) + `test_ladder` (66) + `test_thirddoor` (26) + `test_fifthdoor` (83) | webfetch ladder + chrometabs |
| `test_phase0-6` (318) | the historical feature suites (select per-package cases) |

(Counts from CI/AGENTS.md; case lists read in investigation 1-g.)

## 5. Bug policy (preserve vs fix)

Default: **preserve observable behavior**, including quirks, unless listed below. Suspected
defects found during porting are recorded in the implementation report, not silently fixed.

Pre-classified quirks with explicit dispositions:

| Quirk | Evidence | Disposition |
|---|---|---|
| `note_reads._unified_from_json` drops `stars` on DB hits | investigation 1-a | **Fix in Go** (restore `stars`), regression test added, CHANGELOG note — it is a latent bug with no user reliance (fresh parses return stars) |
| RAM budget counted in characters not bytes | `note_reads.py:434` | **Fix** (byte budget) — internal only, no observable contract |
| `os.walk` unsorted ordering | multiple | **Preserve** for outputs that depend on it (duplicate winners) by mimicking: use lexical walk + document; differential cases pin the winner per corpus |
| `telegram_api_id: 0` counts as unset | `user_config.py:175` | **Preserve** (0 is never a real api_id) |
| `_status_is_reviewed` truthy set includes `on` inconsistently | investigation 1-b | **Preserve** (grammar rarely hit; note in spec) |
| Python default openers honor proxy env for some remote legs | llm_client preflight etc. | **Preserve semantics**: Go remote legs honor env proxies except where Python explicitly bypassed |
| Dead prompt files, dead tables, dormant modules | inventory §6 | Not ported (not quirks — removals) |

Any *new* fix during implementation requires: rationale + expected new behavior + regression test
+ entry in the implementation report (mirrors SPEC.md's bug-recording law).

## 6. Traceability

`traceability-matrix.md` connects requirement → source evidence → module spec → Go component →
test category → task → acceptance criterion. Maintained by the implementation session as suites
land.

## 7. CI shape for the Go app (target)

- `go vet`, `gofmt -l`, `go test ./...` (unit+state+contract), `go test -tags integration ./...`
  (git/CDP), differential harness job (needs Python available — run on the repo's existing
  ubuntu runner), packaging job (build exe + zip + sha256), benchmarks reported on a schedule.
- The **Python CI gate stays untouched** until the cutover release (the Python app remains the
  product during migration).
