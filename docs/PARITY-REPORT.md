# The parity report — where the Go port stands (T-83, SESSION-051, 2026-10-21)

> **Who this is for:** the owner. Plain language, every claim with its
> evidence. This is the implementation report doc 15 §7 asked for (AC-25),
> written to be read without touching code (AC-26, per the reference's own
> SPEC.md §0 law). Numbers here are either measured (the evidence doc is
> named) or marked honestly as not-yet-measured.

## 1. The verdict in plain words

The Go port of GitCurator is **functionally complete and proven
byte-equal to the Python app** on everything the two share: same config
file, same cache database, same vaults, same notes out, same command
line, same desktop window, same words on screen. The proof is not "it
looks the same" — it is a test system that **runs the unmodified Python
app and the Go app on the same inputs and compares the outputs byte for
byte** (the differential corpus, §4). The port is also **faster**
(§5) and ships as **one folder with no Python installed** (§2).

What is genuinely still open is small and listed in §9: the owner's
live Windows checklist (T-82), the milestone release cut that carries
the desktop app (the owner's call), and a handful of recorded debts.

## 2. What shipped, and how to get it

**Two faces, one engine:**

- **GitCurator-cli.exe** — the full command line (`--status`,
  `--init`, `--test-connection`, `--detect-llm`, `--dry-run`, `--auto`,
  `--sync`, `--retry-failed`, `--import-file`, `--hand-delivery`,
  `--chrome-retry`, `--list-dead`, `--reset-dead`).
- **GitCurator.exe** — the desktop app (the themed window: the SYNC
  desk with the log and the batch machine, the scan views, every
  Settings page, the More tools — every page, dialog and control the
  Python app has).

**What you can download today:** the live GitHub release
[`v0.92.0-go.1`](https://github.com/assadigit/second-brain-project/releases/tag/v0.92.0-go.1)
— the **CLI milestone** zip (12 files; the GUI exe was not in it yet
when it was cut). **The desktop app is complete and build-proven**
(§4.4) and rides the **next milestone release** — the two-exe zip
(13 files, both exes + launchers + assets + VERSION). Cutting that
release is one `git tag` away through the proven release workflow;
the decision is the owner's (formal releases happen only at roadmap
milestones, per the session protocol).

First-run guide: **`WINDOWS-QUICKSTART.md`** (updated this session
with the download + verify steps).

## 3. Task-by-task status (the whole task list, one line each)

62 rows in doc 13. 61 actionable (T-51 was removed by the owner's
Option-B decision before implementation). Status today: **59 COMPLETE,
1 PARTIAL (the owner's leg), 1 COMPLETE this session (T-83)**.

| Phase | Tasks | Status |
|---|---|---|
| **0 — proofs & harness** (SESSION-001/002) | T-00 module skeleton · T-01 Telegram PoC (gotd proven on the owner's account) · T-02 TLS impersonation PoC · T-03 differential harness · T-04 Python baselines | **ALL COMPLETE** — the three PoC verdicts final; the baselines measured (`t-04-python-baselines.md`) |
| **1 — foundation** (SESSION-003–005) | T-10 constants · T-11 links · T-12 storage+dryrun · T-13 prompts · T-14 taxonomy · T-15 notebuilder · T-16 config · T-17 helpers | **ALL COMPLETE** — generated-golden parity |
| **2 — data layer** (SESSION-006–010) | T-20 store · T-21 notestate · T-22 notereads · T-23 websitestate · T-24 cacherepos · T-25 batch/linktracker | **ALL COMPLETE** — opens Python-created DBs without migration (AC-9) |
| **3 — network & LLM** (SESSION-011–014) | T-30 llmclient · T-31 webfetch · T-32 webextract · T-33 connectioncheck · T-34 workerclient | **ALL COMPLETE** — payload bytes identical to Python's |
| **4 — pipelines** (SESSION-015–026) | T-40 websitepipeline · T-41 directorynote · T-42 githubpipeline · T-43 handdelivery · T-44 chrometabs · T-45 vaultscan · T-46 hygiene+banish · T-47 vaultseal · T-48 goodrepos+radar · T-49 directorysync · T-50 mirror/linking | **ALL COMPLETE** — T-51 (direct Telegram) **removed by the owner's Option-B stamp** (ADR-0004 Am.3; §6.1 below) |
| **5 — CLI parity milestone** (SESSION-027–032) | T-60 cli · T-61 terminal paint · T-62 sync/auto/dryrun · T-63 launchers · T-64 corpus full run · T-65 the release | **ALL COMPLETE** — `v0.92.0-go.1` tagged + released with the verified zip |
| **6 — GUI** (SESSION-033–047) | T-70 Wails PoC/gate · T-71 app commands/bus · T-72 frontend foundation · T-73 Settings A · T-74 Settings B · T-75 Sync view · T-76 Scan flow · T-77 window chrome · T-78 side-by-side · T-79 doors addendum · T-79a per-test dialogs · T-79b linking tools · T-79c bot-queue tools · T-79d Chrome doors · T-79e decommission picker · T-79f fanfare triggers | **ALL COMPLETE** — the More menu has **no disabled row left**; every divergence recorded in its evidence doc |
| **7 — packaging, release, docs** (SESSION-048–051) | T-80 deterministic zip + ritual · T-81 release workflow · T-82 clean-VM checklist · T-83 docs (this) | T-80 **COMPLETE** · T-81 **COMPLETE** (dry-run green on a real runner + cross-machine proof) · T-82 **PARTIAL** (items 1–8 are the owner's Windows legs — the prepared checklist is `docs/poc-artifacts/t-82-clean-vm-checklist.md`; item 9's automation half is green) · T-83 **COMPLETE** (this report + the README front page + the QUICKSTART update) |

## 4. How we know it matches Python (the proof system)

Four independent kinds of evidence, all archived in this repo:

### 4.1 The differential corpus — the strongest proof

33 golden generators (`testdiff/goldens/`) each **run the unmodified
Python app** (pinned at `98f4fc7`, the exact rev in
`testdiff/PINNED-GITCURATOR-REV`) and freeze its output as the golden;
the Go suite must reproduce those bytes. The full-corpus run
(`t-64-corpus-full-run-report.md`): **all generators green twice in a
row, byte-identical between runs, and the committed goldens equal a
fresh regeneration** — the corpus cannot drift. Every new module lands
its generator with it (the corpus law), so the count grew from 31 at
T-64 to 33 today.

### 4.2 The test battery

**2,397 main-module tests + 89 webui tests = 2,486, over 40 packages —
zero failures, race-detector clean, `gofmt` clean** (re-verified at
SESSION-061's close; 2,377 + 89 until SESSION-058, which added the
five chrome-retry live-leg tests, then SESSION-059 added the twelve
noteredo core-level selections — and closed the REAL delivered-law
parity gap they exposed, §9 — then SESSION-060 added the two harvest
probe-ladder selections that closed the recorded cousin of that gap
(the closed-DB probe panicked where the reference answers 0; the
record is in spec-core-website-pipeline.md's Tests section), and
SESSION-061 added the CRLF round-trip differential that closed the
surgery path's text-mode remainder — the Go read CRLF notes raw where
the reference's text mode translates, so a moved CRLF note came out
with mixed line endings and every stored fingerprint diverged from
Python's; the fix + the record are in spec-core-note-state.md's Tests
section; the same numbers CI proves on every push).

### 4.3 Data compatibility (your existing data works as-is)

- The Go app opens a **Python-created `cache.db`, `config.json` and
  `sync_review.db`** without migration (fixture tests; AC-9).
- The config round-trip is **byte-stable** (saving from Go rewrites
  the file exactly the way Python would; AC-10).
- Go and Python instances **mutually exclude** through the same
  `app.lock` — they can sit side by side on the same data (T-78's
  side-by-side verification; AC-11).
- Dry-run writes **nothing** — snapshot-diff proven (AC-12).

### 4.4 The release-build proof (Phase 7)

- The **double-build ritual**: both exes + the zip are built twice
  plus one cold-cache leg each; every sha256 pair identical (AC-22).
- The **release workflow** (tag → build → attach zip) ran green
  end-to-end on a real GitHub runner, produced a draft `[DRY-RUN]`
  release with the zip, and — the strongest result — **the runner's
  three artifacts were byte-identical to a full sandbox rebuild at the
  same tagged state** (cross-machine reproducibility; the full table
  is `t-81-release-workflow.md` §5). The refusal leg proved the
  guard: a wrong-shaped tag fails exactly at the version check and
  creates no release.

## 5. Speed — the measured numbers

Same sandbox class for both columns (the Python column is T-04's
baseline rig + a fresh same-box re-run at SESSION-052; the Go column
is the gate-leg measurements + the SESSION-051 CLI numbers + the
SESSION-052 scan benchmark — the scan row's two sides measured on the
same box):

| What | Python (measured) | Go (measured) | Target |
|---|---|---|---|
| Cold start → first frame (GUI) | ≈ 805 ms | 213–892 ms across the gate legs | ≤ 1.5 s ✅ |
| CLI `--status`, no config yet | 49.3 ms | **4.8 ms** (measured 2026-10-21, median of 3) | ≤ 0.5 s ✅ |
| CLI `--status`, full config card | 893.6 ms | **≈ 134 ms** (measured 2026-10-21, median of 3) | ≤ 0.5 s ✅ |
| Log desk, 10,000-line flood | drains in 1,960.9 ms | ≥ 454,000 lines/s, **0 dropped lines**, handler ≤ 30 ms/batch | UI stays interactive ✅ |
| GUI exe size | (Python needs an interpreter + packages) | 14.95 MB, single file, no installer | < 20 MB ✅ |
| Full vault scan (owner-scale ≈ 1,300 notes) | 35.3 ms cold / 16.7 ms warm (synthetic vault; same-box re-run 2026-10-21: 32.8 / 16.9) | **14.6 ms cold / 5.3 ms warm** (synthetic vault, measured 2026-10-21, median of 3 — 44.5% of Python same-box, 41.4% of the recorded baseline; same-input + same-output proven: [`ac17-go-vaultscan-benchmark.md`](poc-artifacts/ac17-go-vaultscan-benchmark.md)) | ≤ 50% of Python ✅ |

Both `--status` rows and the flood row beat Python by 6–10×; the vault
scan is 2.25× faster (measured 2026-10-21 on the same synthetic corpus
Python's baseline used, same-input and same-output proven — the record
is `docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`). The sandbox
numbers are not your Windows numbers — T-82's checklist items 2/3
measure them on your machine (that is its purpose), and the scan row's
real-vault number lands at your first live run.

## 6. Every deviation from the original plan (and why)

The blueprint is binding, but reality won a few times. Every deviation
is recorded, most with an ADR amendment. The complete list:

| # | Deviation | Why | Where recorded |
|---|---|---|---|
| 1 | **The desktop's direct Telegram leg is dropped** (T-51 removed). The bot queue rides the Cloudflare Worker's HTTP API only; read-acknowledge is gone. | The owner's own stamp ("OPTION B CONFIRMED", 2026-10-15): one queue owner (the Worker), no MTProto session on the desktop. The gotd capability was PoC-proven first and sits in the Phase-8 register if you ever want it back. | ADR-0004 Am.3; doc 13's amendment |
| 2 | **The GUI lives in its own nested Go module** (`webui/`), not the main one. | Wails (the GUI framework) drags its own dependency tree. Keeping it nested means the main module carries **exactly four runtime dependencies** (§7) and the GUI tooling never pollutes the engine. The T-72 dissolution probe tested merging and found it worse; the final merge/dissolve call stays yours. | ADR-0002 Am.3 + Am.5 |
| 3 | **The main module's Go directive sits at 1.24.1**, not newer. | Forced by tls-client v1.16.0's own floor (the impersonation library). Owner-acknowledged. | ADR-0002 Am.2 |
| 4 | **The HTML parser is hand-rolled** (byte-compatible with CPython's `html.parser`), not a library. | The reference's extraction behavior is pinned to CPython's exact tokenizer quirks; a third-party parser would break byte-parity. Zero new dependencies; four generated data tables; differential-proven. | T-32's record (doc 13) |
| 5 | **Two pre-classified bugs fixed in Go** (the only behavior fixes allowed): `stars` restored on DB hits in note-reads; the RAM budget counted in bytes, not characters. | Both were latent Python bugs with no user reliance, pre-classified as "fix" by the blueprint's bug policy before porting started. Regression tests + CHANGELOG notes ride both. | doc 09 §5; T-22's record |
| 6 | **The Sync view shares one Refresh payload for both desks** (the reference defers this to its own v2). | The reference's own stated intent, taken now — one gather instead of two divergent ones. | SESSION-038's record (doc 13, T-75) |
| 7 | **The zip grew from 12 files (CLI milestone) to 13** (both exes + assets), and the v0.86.0 deferred-ledger rider now lands on both zips. | The desktop app shipped after the CLI milestone; the rider fix corrects where the backup ledger lands inside export zips (the reference's own arcname). | T-80's record; `t-80-reproducible-packaging.md` |
| 8 | **Every GUI door's small divergences** (6 at T-79e, 6 at T-79f, etc.). | Each is a Wails-vs-PyQt mechanical difference (e.g. a dialog's exact close behavior), recorded and pinned by test — not silent. | `t-79a…t-79f` evidence docs |
| 9 | **The exe icon was the Wails template default** for the T-70…T-79 era builds; fixed at T-80 (the brand icon is now embedded, byte-proven inside the exe). | A build-config find, not a behavior change. Any exe you got before SESSION-048 shows the generic icon. | T-80's record |

## 7. The decisions log (ADR amendments, one line each)

Ten architecture decision records; the amendments are the project's
decision history:

- **ADR-0001 GUI framework** — Wails v2 chosen. Am.1 owner Q&A ·
  Am.2 the PoC verdict (T-70 gate green) · Am.3 the Phase-6 entry
  gate re-run.
- **ADR-0002 module/layout** — Am.1 the move to this repo · Am.2 the
  go-directive floor · Am.3 the nested webui module · Am.4 the
  internal-boundary reality correction · Am.5 the T-72 dissolution
  probe (stays nested).
- **ADR-0003 SQLite via modernc.org/sqlite** — Am.1 the CAST read
  law (Python's sqlite3 returns strings where Go's driver returns
  ints; the port normalizes to Python's shapes).
- **ADR-0004 Telegram** — Am.1 interim PoC · Am.2 final PoC
  (ACCEPTED — gotd works) · Am.3 Option B (the owner's stamp; §6.1).
- **ADR-0005 the Worker stays JavaScript** — no amendments (holding).
- **ADR-0006 network stack** — Am.1 the tls-client PoC verdict ·
  Am.2 the owner acknowledgement (stdlib + SOCKS + Chrome-profile
  impersonation through tls-client v1.16.0).
- **ADR-0007 CLI-first order** — Am.1 the owner re-scope · Am.2 the
  CLI legs re-instated.
- **ADR-0008 config/data compat** — Am.1 the T-78 side-by-side
  record (the lock law proven across binaries).
- **ADR-0009 git stays a subprocess** — no amendments (holding).
- **ADR-0010 packaging** — Am.1 the owner Q&A (dual artifact,
  bootstrapper chain, icon, signing) · Am.2 the app icon resolved.

**The authorized runtime dependency set (the whole list):**
`modernc.org/sqlite` v1.36.1 · `bogdanfinn/tls-client` v1.16.0 ·
`golang.org/x/net` v0.48.0 · `golang.org/x/text` v0.32.0 — plus the
Wails build-time toolchain in the nested GUI module. Nothing else;
nothing bumped.

## 8. Specs corrected where reality differed

The standing law is "reality wins — fix the spec in the same
commit." The notable corrections: the T-12 spec-name fix
(`spec-core-dryrun.md` never existed; the row was reality-fixed);
ADR-0002 Am.4 (the internal-boundary reality); doc 10's `[verify]`
markers settled with dated notes at T-80 (the GUI build command's
reality; the hardcoded UA truth); the T-49/T-50 spec corrections
(trio filenames; the test-suite mapping); the stale-target correction
in the T-81 evidence doc §2.1. Each rode the commit that found it.

## 9. What is honestly not done yet

- **T-82 items 1–8 — the owner's live Windows checklist** (fresh-VM
  install, cold start on your hardware, the AV scan, the batch
  smoke). The prepared fill-in document is
  `docs/poc-artifacts/t-82-clean-vm-checklist.md`; your signed
  results close it. Item 9 (sha256 + reproduce spot-check) is already
  automated and green.
- **The desktop milestone release** — cutting the next tag ships the
  two-exe zip through the proven workflow. The owner's call.
- **Recorded test debts** (evidence-recorded, not blockers):
  `test_noteredo`'s remaining un-ported selections — the CLI
  routing-law class (TestCLIRouting) landed at SESSION-058 with the
  chrome-retry live leg it was guarding, the CORE-LEVEL
  TestRedoScan shapes + the iconcolumn §7 never-re-ask leg landed at
  SESSION-059 (the port closed a REAL parity gap: both master scans
  had reused the batch-side collector as their delivered probe, so a
  CONSUMED row was invisible to the redo scan — THE OWNER'S EXACT
  CASE — and re-asked by the hand scan; the record is in
  spec-core-website-pipeline.md's Tests section), the harvest's
  probe-panic-ladder cousin landed at SESSION-060 (the closed-DB
  probe panicked where the reference's try/except answers "a broken
  probe harvests nothing"; the two pinning selections are
  `internal/websitepipeline/handharvest_test.go`), and the note_state
  SURGERY path's CRLF round-trip differential landed at SESSION-061
  (a REAL parity gap of the same cross-binary class: the reference
  reads every note in TEXT MODE — universal newlines on read, raw
  `newline=''` write — so a CRLF note through the moves surgery comes
  out ALL-LF and every stored fingerprint is of the translated
  content, but the Go read raw at four sites: mixed-endings output +
  raw fingerprints a Python-baselined row would disagree with forever;
  the golden + the fix + the record are
  spec-core-note-state.md's Tests section); what remains is the
  GUI-side hero/wiring equivalents (TestHeroRedoRouting +
  TestRedoPassWiring) where the app seam already covers the guards;
  the launchers' clean-VM smoke (rides T-82); the bootstrapper leg
  (doc 10 §7).
- **Standing owner calls carried**: the webui-module dissolution
  (ADR-0002 Am.5); vanilla-vs-vite for the frontend; the
  batch-scorecard modal question; the v0.91.0 zip-parity reading
  (minor); the answering-window belt + the T-71 merged-verdict
  question.

## 10. How to double-check any claim here

- The corpus: `testdiff/README.md` + the archived full-run report.
- The battery: the `go-ci` workflow on any recent push (green), or
  `go test ./...` in both modules from a clean checkout.
- The release proof: `docs/poc-artifacts/t-81-release-workflow.md`
  (the runner table + the refusal leg + the cleanup).
- The scan row: `docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`
  (both rigs + the cross-check driver rerunnable; the recipes inside).
- Any task's detail: its row in
  `docs/port-blueprint/13-implementation-task-list.md` names its
  evidence doc; every session's work is in `docs/sessions/`.
- The reproduce-from-tag recipe (any machine with go 1.27.2):
  `git clone --branch <tag> … && ./build/release.sh` — the ritual's
  table must equal the release notes' table.

---

*Report assembled at SESSION-051 (T-83); §5's scan row measured and
closed at SESSION-052 (2026-10-21, the AC-17 benchmark —
`docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`). The per-session
engineering history lives in `docs/sessions/`; the task list with every
row's full record is `docs/port-blueprint/13-implementation-task-list.md`.*
