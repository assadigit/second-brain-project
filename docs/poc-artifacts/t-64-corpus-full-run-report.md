# T-64 — THE DIFFERENTIAL CORPUS FULL RUN (the archived report)

- **Run**: 2026-10-19 (SESSION-032), both passes + all three legs green
- **Work repo**: `assadigit/second-brain-project` @ the T-64 fix-loop commit
  (`42739b3` + this report's landing commit)
- **Reference**: `assadigit/GitCurator` @ `98f4fc71cfb862cee7225d864636c08062724348`
  (byte-equal to `testdiff/PINNED-GITCURATOR-REV`; the rev guard ran on every
  generator — every `[goldens] reference rev` line matched)
- **Toolchain**: go1.27.2 linux/amd64 (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`),
  venv `/home/z/venvs/gitcurator` (Python 3.12.15, stdlib-only)
- **Driver**: `testdiff/run_corpus.py` (new this session — the corpus's own
  runner; usage + laws in its docstring, per-run artifacts gitignored under
  `testdiff/runs/corpus/`)

## Verdict

**ALL GREEN.** The full corpus (doc 09 §3's case inventory, realized as the
31 generators under `testdiff/goldens/` — every golden the project has ever
landed) ran twice end-to-end against the unmodified pinned reference and
proved the three corpus laws:

1. **Per-generator health** — all 31 exited 0 on BOTH passes, every one
   reporting the pinned rev, every one writing its goldens (mtime
   touched-set non-empty; the silent-generator guard found none).
2. **THE TWO-RUN BYTE-IDENTICAL PROOF** — pass 1 vs pass 2 full-tree sha
   snapshots: `{'only_in_1': [], 'only_in_2': [], 'sha_mismatch': []}`
   (beyond the four atomic pairs' documented timestamp churn — see the pair
   law below).
3. **COMMITTED == REGENERATED** — after both passes the tree is clean
   except the four atomic pairs and the driver's own source: every
   committed golden equals the reference's current bytes at the pinned rev.

**The atomic-pair law** (the four `.db`-fixture generators — `store`,
`notestate`, `notereads`, `websitestate`): their Python-created fixtures
embed the reference's own `now()` strings and *necessarily change per run*
(the generators' own documented law); the corpus proof for them is
threefold: (a) the **PAIR TIE** — the regenerated `.db`'s sha256 equals the
same run's golden `fixture_sha256` (all four, both passes: OK); (b) the
**two-run stability of the NORMALIZED content** (ISO timestamps masked,
`fixture_sha256` masked, scratch roots folded, `mtime_ns`/epoch-second
floats masked, the dismissed-ledger tie order compared as a sorted
multiset — all four: identical); (c) the **committed pairs stay canonical**
(restored after the run; the Go suite proves each committed pair's own tie
on every test run).

## The three legs

| Leg | What ran | Evidence |
|---|---|---|
| **Generators** | 31 generators × 2 passes (`testdiff/run_corpus.py --pass 1/2 --report`) | verdict GREEN; per-generator table below |
| **Go replays** | `go test -race -count=1 ./...` (the replay tests byte-compare the Go port against the goldens the generators produce) | **1922/1922 top-level tests PASS, 0 FAIL, 0 SKIP, 0 races, 33 packages ok** |
| **CLI-level harness** | `python3 testdiff/harness.py` (all cases) | `status-no-config`: **PASS** — python/go stdout sha256 `f71074cd…` identical (821 bytes), stderr identical (0 bytes), exit codes match; the run report is regenerated under `testdiff/runs/status-no-config/` |

## Per-generator results (both passes)

| generator | exit p1/p2 | two-run | touched | churned | pair |
|---|---|---|---|---|---|
| store | 0/0 | identical | 4 | 2* | tie OK; normalized identical |
| cacherepos | 0/0 | identical | 6 | 0 | — |
| chrometabs | 0/0 | identical | 7 | 0 | — |
| cli | 0/0 | identical | 4 | 0 | — |
| config | 0/0 | identical | 3 | 0 | — |
| constants | 0/0 | identical | 5 | 0 | — |
| directorynote | 0/0 | identical | 6 | 0 | — |
| directorysync | 0/0 | identical | 8 | 0 | — |
| githubpipeline | 0/0 | identical | 4 | 0 | — |
| goldenwebsites | 0/0 | identical | 4 | 0 | — |
| goodrepos | 0/0 | identical | 2 | 0 | — |
| handdelivery | 0/0 | identical | 3 | 0 | — |
| helpers | 0/0 | identical | 3 | 0 | — |
| links | 0/0 | identical | 5 | 0 | — |
| linktracker | 0/0 | identical | 7 | 0 | — |
| llmclient | 0/0 | identical | 9 | 0 | — |
| mirror | 0/0 | identical | 4 | 0 | — |
| notebuilder | 0/0 | identical | 5 | 0 | — |
| notereads | 0/0 | identical | 7 | 2* | tie OK; normalized identical |
| notestate | 0/0 | identical | 11 | 2* | tie OK; normalized identical |
| prompts | 0/0 | identical | 18 | 0 | — |
| reporadar | 0/0 | identical | 18 | 0 | — |
| storage | 0/0 | identical | 29 | 0 | — |
| taxonomy | 0/0 | identical | 27 | 0 | — |
| vaulthygiene | 0/0 | identical | 24 | 0 | — |
| vaultscan | 0/0 | identical | 17 | 0 | — |
| vaultseal | 0/0 | identical | 6 | 0 | — |
| webextract | 0/0 | identical | 7 | 0 | — |
| webfetch | 0/0 | identical | 9 | 0 | — |
| websitepipeline | 0/0 | identical | 4 | 0 | — |
| websitestate | 0/0 | identical | 9 | 2* | tie OK; normalized identical |

\* churn = the atomic pair's documented timestamp churn (the `.db` + its
content golden regenerate together; the committed pair restored after the
run). "touched" counts files the generator rewrote (byte-identical
rewrites included — the mtime evidence the generator actually wrote).

## The corpus ↔ the ten named stones (doc 09 §3's inventory)

| Stone (doc 09 §3 / SESSION-025 §gate) | Generators that carry it | Go replay |
|---|---|---|
| the 5 HTML fixtures | `webextract` (corpus.fixtures: article/js_shell/landing/latin1/paywall + the embedded tables) | `internal/webextract` |
| import-file grammar corpus | `links` (+ the CLI-level `--import-file` leg in `cli`/`goldenwebsites`) | `internal/links`, `internal/cli` |
| moves-as-corrections scenarios | `notestate`, `notereads` | `internal/notestate`, `internal/notereads` |
| banishment declined/confirmed/timeout | `websitepipeline` (the banish flows), `websitestate` | `internal/websitepipeline` |
| scan plan apply with picked subsets | `vaultscan` | `internal/vaultscan` |
| seal ladder on temp repos | `vaultseal`, `goodrepos`, `vaulthygiene` | `internal/vaultseal`, `internal/store` |
| retry backoff clock (faked `now`) | `websitestate` (due_retries at FIXED_NOW), `cacherepos` (threshold/backoff) | `internal/websitepipeline`, `internal/store` |
| dry-run report equality | `storage` (the dryrun seam), `websitepipeline`, `githubpipeline`, `linktracker`, `cli` | every module's golden pins its dry-run outputs |
| the 30-link golden set | `goldenwebsites` | `internal/websitepipeline/goldenrunner_test.go` |
| the offline-golden-runner equivalent | `goldenwebsites` (same corpus, the Go replay IS the runner equivalent) | `internal/websitepipeline` |
| (Phase 5) the CLI differential stone | `cli` (goldencli, 101,777 bytes incl. the launchers section) | `internal/cli` |

## The fix loop (what the full run caught — 5 finds, all fixed in-commit)

The corpus run's FIRST pass (before the fixes) caught four real
divergences — all generator-side determinism bugs, none a Go-port
divergence (the replays were already correct), plus one latent
normalization gap in the pair law itself:

1. **`dump_githubpipeline.py`** — the golden recorded the RAW mkdtemp root
   (`/tmp/ghgolden-XXXX`, 4 path fields): every regeneration produced
   different bytes (the two-run law was impossible; the committed golden
   carried one machine's random suffix). FIX: fold the scratch root to
   `<TMP>` at payload end. The Go replay's `ghgoldenRootRE` fold treats
   the folded prefix as a no-op — the replay stays byte-exact (re-verified
   green).
2. **`dump_websitepipeline.py`** — same bug (`/tmp/wpgolden-XXXX`, 12 path
   fields incl. DB-dumped `note_path` rows that arrive as TUPLES — the
   tuple branch was the fold's second iteration). Same fix; replay green.
3. **`dump_directorynote.py`** — same bug (`/tmp/dirgolden-XXXX`, 59 path
   fields). Same fix; replay green.
4. **`dump_mirror.py`** — the `collect_note_fields` duplicate fixture used
   two DIFFERENT stems (`p.md` + `p-dup.md`); the reference keeps the
   FIRST hit over RAW `os.walk` order (filesystem-dependent — doc 09 §5's
   pre-classified quirk), so the recorded winner's stem flipped between
   machines (the committed golden pinned one machine's accident; this
   sandbox's /tmp produced the other winner). FIX: the duplicate keeps
   the SAME STEM in a different folder (`Design/Misc/p.md`) — one stem,
   nothing to flip; the duplicate count (the contract) unchanged. The Go
   golden replay's fixture mirrors the same shape (NOT a weakening: the
   test still byte-compares the whole section; it now pins the
   order-independent contract; the lexical-walk mimic law itself is
   unchanged in `linking.go`).
5. **The pair-stability normalization gap** (the runner's own first
   iteration) — the four atomic pairs' content goldens embed MORE per-run
   facts than ISO timestamps: the notereads rows carry the scratch root,
   `mtime_ns` ints and `scanned_at` epoch floats; the websitestate
   dismissed ledger's row order among same-second batch dismissals is
   UNSPECIFIED ON BOTH SIDES (the reference's `purge_blocked_domains`
   iterates a Python SET under a tied seconds-precision `dismissed_at`;
   the Go port iterates a map — faithfully mirrored). The runner's
   stability digest now masks exactly those documented per-run facts
   (goldens stay RAW — the Go replays prove each pair's internal
   consistency through the exact frozen values; the masks apply only to
   the cross-run stability proof). Verified: 6 consecutive regenerations
   of each previously-flaky pair produce a single stable digest.

**No Go-port divergence was found** — the fix loop's five finds were all
about making the *provenance proof* (regeneration) deterministic; the
*parity proof* (the replays) was green throughout (1922 tests before the
run, 1922 after the fixes).

## Reproducing

```sh
python3 testdiff/run_corpus.py --pass 1   # ~35 s (llmclient 19 s, webfetch 12 s)
python3 testdiff/run_corpus.py --pass 2
python3 testdiff/run_corpus.py --report   # exit 0 = GREEN
go test -race -count=1 ./...              # the replays
python3 testdiff/harness.py               # the CLI-level cases
```
