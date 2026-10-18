# SESSION-026 — 2026-10-18 — Phase 4 CLOSED: the closer landed — the 30-link offline golden runner equivalent green on the first run (the reference's own run_golden_websites.py --offline replayed byte-for-byte in Go) — 1828 tests race-clean; the twentieth full wipe repaired before work

## Goal (from NEXT-SESSION.md at start)

SESSION-026 = **the Phase-4 closer (M)** — the exit gate's last named
artifact: the 30-link offline golden runner equivalent. A generator
driving the UNMODIFIED reference's `tools/run_golden_websites.py
--offline` over its own corpus into a throwaway vault, dumping the
notes + the report; a Go replay byte-compared. Then Phase 5 OPENS
(T-60 CLI parity) if capacity remained.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was WIPED (the twentieth full wipe)**: neither
  repo, no Go, no venv, no symlinks. Rebuilt per NEXT-SESSION v26 §1
  before any work: the PAT verified against the GitHub API's
  authenticated /user/repos listing; both repos cloned with the
  ephemeral Basic `x-access-token` header and **scrubbed immediately**
  (verified by direct config read — 0 extraheader, 0 token bytes in
  both configs, 0 token strings anywhere in the clones);
  GitCurator checked out at `98f4fc7` (matches
  testdiff/PINNED-GITCURATOR-REV byte-for-byte); go1.27.2 installed at
  /home/z/.local/go (go.dev tarball, sha256 `ecbadb99…` — the same
  digest SESSION-025 recorded); the venv rebuilt
  (/home/z/venvs/gitcurator, Python 3.12.15, stdlib-only); the
  symlinks re-created. go.mod verified UNCHANGED (v1.36.1 /
  tls-client v1.16.0 / x/net v0.48.0 / x/text v0.32.0 / go 1.24.1 —
  the authorized set exactly). The baseline re-verified BEFORE work:
  **1826/1826 race-clean over 30 packages** (run in foreground chunks
  after the first two detached background attempts were reaped by the
  sandbox mid-module-download — a new operational find, recorded in
  §Handoff; the top-level test count matched the card exactly via
  `go test -v | rg '^=== RUN' | rg -v '/' | wc -l` → 1826).

- **Read-first, per the card**: the reference's
  `gitcurator/tools/run_golden_websites.py` (the --offline mode: the
  fake LLM keyed by canonical URL answering from the golden file, the
  offline_fetch canned pages, render_report + verdict_of, the
  exit-code law) + `tests/golden/websites.json` (the 30-entry corpus)
  + the working precedents (dump_mirror.py's fold/frozen-clock/
  importlib patterns, T-40's golden_test.go harness).

- **Probes before code** (the evidence-first habit): the runner driven
  manually (`--offline --out /tmp/…` → 30/30 processed, exit 0, the
  report + the taxonomy-summary path find); the pipeline driven
  directly with capturing seams (the throwaway vault's tree = 30
  `Golden_test_site*.md` notes across category folders with
  unique_path v-suffixes; the results' 8-key shape with absolute
  note_path; the console's 31 log lines + 3 stat lines, zero tmpdir
  leakage) — and the CHOREOGRAPHY find: **the per-link LLM call order
  is classify → classify → analyze (90 calls total), and the analysis
  prompt (w03) carries no `Website: ` line** — so the fake's URL scan
  misses and every analysis answer is the DEFAULT dict: the quirk all
  30 golden notes are named "Golden test site" by. Also confirmed:
  Python `str.capitalize` lowercases the REST ("AI illustration
  maker — …" → "Ai illustration maker — …") — the canned title quirk.

- **THE CLOSER, commit `7aea86c` (8 files, +2,011/−5)**:
  - *`testdiff/goldens/dump_goldenwebsites.py`* — the generator. Drives
    the reference's OWN `main(['--offline', '--out', …])` (the tool
    module imported via importlib — UNMODIFIED) with the house seams:
    the frozen clock (2026-10-17 12:00:00 local — wp/ws/nb/runner
    `datetime` rebound), a tempfile recorder + an rmtree-recording
    shutil stub (the throwaway vault survives the dump; the generator
    cleans up with the REAL shutil after), and a recording
    WebsitePipeline subclass (captures the LLM calls + run()'s
    results; delegates everything). Folds by REPLACE: both temp roots
    → `<tmp>`, the reference app dir → `<app>` (the report EMBEDS the
    taxonomy path — the SESSION-025 law applied). The outer harness
    runs the payload TWICE and refuses to write unless byte-identical
    — the determinism proof is part of every regeneration.
  - *`internal/websitepipeline/testdata/goldenwebsites.golden.json`*
    (94,979 bytes, two-run byte-identical) — 11 sections: corpus (the
    30 entries), canned_fetch (3 representative pages — the simple,
    the em-dash/arrow, the no-subcategory), llm_calls (the full
    90-call choreography: url/task/reply), prompt_samples (the first
    three calls' exact message bytes), results (the 30 dicts,
    note_path folded), notes (the whole throwaway vault tree — 30
    files), report, console, exit_code, invalid, processed_count.
  - *`internal/websitepipeline/goldenrunner_test.go`* — the Go replay.
    Rebuilds the runner's own logic as the Go twins (gwOfflineFetch +
    gwCapitalize, gwOfflineLLM over OrderedMap+PyDumpsCompact — the
    Python json.dumps byte law, gwVerdictOf + gwRenderReport ported
    byte-exactly), runs the ported pipeline over the SAME corpus with
    the same config dict (taxonomy_path → the repo's byte-identical
    assets copy, sha256-verified equal to the reference's) and the
    same frozen clock, then byte-compares all seven surfaces via
    t.Run subtests: canned pages, results, notes (the vault walk),
    report (the `<repo>/assets` → `<app>` fold), console (the log
    lines + the stats prints), exit-code law, LLM choreography +
    prompt samples.
  - **The result: GREEN ON THE FIRST RUN** — T-40's port needed ZERO
    fixes to replay the reference's own end-to-end corpus run. The
    mutation test proves the differential has teeth: one byte in the
    golden's report section ("30 links" → "31 links") fails the report
    subtest with a precise diff; restored, green again.
  - *The stamps*: the roadmap's Phase-4 exit gate paragraph carries
    **GATE CLOSED (SESSION-026)** citing SESSION-025.md's stone table
    (the table's last two rows are the SAME artifact; with the closer
    landed, all 10 stones are green); README declares **Phase 4
    CLOSED (SESSION-015–026)** with the closer's record + the new
    leading status; traceability D1's evidence cell carries the
    closer; internal/websitepipeline/doc.go + spec-core-website-
    pipeline.md carry the SESSION-026 shipped records (the two quirks
    written down as Go notes: the analysis-prompt URL-scan miss, the
    capitalize rest-lowercasing).

- **T-60 (Phase 5's first task) deliberately deferred after a scoped
  capacity assessment**: the reference CLI surface is 3,502 LOC
  (cli.py 1,777 + cli_run.py 719 + cli_sync.py 381 + cli_settings.py
  211 + cli_terminal.py 321 + shims) + the spec's full test battery —
  a session-sized task on its own. The one-task law applied: the
  closer done properly with proof beats a rushed half-CLI. Phase 5
  OPENS with SESSION-027 = T-60.

- **Pushed**: `7301f19..7aea86c main -> main` (the ephemeral PAT via
  the transient `-c http.extraheader` Basic form; the config verified
  CLEAN after — 0 extraheader, 0 token bytes; the commit message
  written to a FILE and committed with -F, the committed bytes
  verified with `git log --format=%B` — only git's own trailing
  newline differs, the content byte-identical; no `[h`/`[m` pairs in
  the message).

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **The Phase-4 exit gate: CLOSED.** The closer — the gate's last
  named artifact — landed and pushed (`7aea86c`). SESSION-025's
  stone-by-stone table recorded 8 of 10 stones green with ONE
  artifact outstanding; the table's last two rows (the 30-link golden
  set + the offline-golden-runner equivalent) are the SAME artifact,
  now landed: the generator drives the unmodified reference's own
  `--offline` runner over its own 30-entry corpus (94,979-byte golden,
  two-run byte-identical — the proof runs on every regeneration), and
  the Go replay byte-compares seven surfaces — **all green on the
  first run, zero port fixes needed**. Evidence: `go test -race
  -count=1 -run TestGoldenWebsites ./internal/websitepipeline/` →
  PASS (both tests, all 7 subtests); the mutation test (golden byte
  changed) → precise FAIL → restore → PASS.
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean (after fixing one tab-mangling my own edit tool introduced in
  doc.go's ASCII diagram — caught by the battery, restored by gofmt,
  the diff vs HEAD verified to be ONLY the intended paragraph);
  `GOOS=windows go build ./...` clean; `go test -race -count=1 -p 2
  ./...` → **30 packages ok, 1828/1828 top-level tests PASS, 0
  failures, 0 skips, 0 data races** (was 1826 at session start; +2 —
  TestGoldenWebsitesCannedFetch + TestGoldenWebsitesOfflineRunner —
  the arithmetic closes exactly); go.mod/go.sum UNCHANGED (zero new
  dependencies).
- **The golden is an honest differential**: generated from the
  unmodified reference at the pinned rev (the pinned-rev check runs
  before the payload); the runner's own main() produced every byte;
  no per-run facts survive the folds (proven by the two-run
  byte-identity built into the generator).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/12-migration-roadmap.md` — the Phase-4 exit
  gate's GATE CLOSED stamp with the full evidence citation.
- `README.md` — Phase 4 CLOSED (SESSION-015–026) in the leading
  status + the closer's record paragraph; "Next: Phase 5 — the engine
  parity milestone (the CLI surface, T-60)".
- `docs/port-blueprint/traceability-matrix.md` — D1's evidence cell
  carries the closer.
- `docs/port-blueprint/module-specs/spec-core-website-pipeline.md` —
  the SESSION-026 shipped record (the closer's construction + the two
  quirks as Go notes + the fold law for the taxonomy path).
- `internal/websitepipeline/doc.go` — the closer's completion record
  in the package doc.
- No reality-wins corrections were needed: the port matched the
  reference's end-to-end run exactly (the first golden in the project
  to land with zero forced fixes — T-40's engine room held).

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — sixteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2
  has the record + the rollback path).
- **The x/net/html keep-or-swap** (standing, not a blocker).
- **Deferred test debt (recorded with evidence, not blockers)**:
  test_noteredo's remaining un-ported selections — the redo-scan
  battery (ScanMasterRedoRows has zero dedicated Go tests) is the top
  item; then the redo-consume laws, the proper-note laws, the
  pipeline-redo legs, the CLI routing legs (T-60's own spec lists the
  CLI-side battery — those land WITH T-60). The GUI wiring legs stay
  the GUI worker shell task's books.
- **Phase 5's T-62 open item stands**: the all-clear gate's Phase-5
  re-spec (ADR-0004 Am.3 bounded; the spec-cli.md dated note records
  it) — T-62's book, never improvised before then.

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session
   (it carried the twentieth-wipe rebuild + the closer's push; the
   configs verified clean after every use).

## Handoff

- Next session: **SESSION-027 = Phase 5 OPENS — T-60 `internal/cli`
  (the engine parity milestone per roadmap Amendment 1): the full
  flag surface, the wizard, the status/test/detect/hand/chrome/retry
  commands, the `--sync` door, dry-run reports, `cmd/gitcurator-cli`
  packaging — per spec-cli.md (READ THE SPEC BEFORE THE CODE; the
  Option-B re-scope notes and the T-62 bounded open item are in it).
  The mirror/tool legs (mirror_manual, add_recall_hooks, build_links,
  the status row) are T-60's books per the T-50 session's record.**
- NEXT-SESSION.md rewritten: yes (v27).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
