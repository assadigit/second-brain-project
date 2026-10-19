# SESSION-032 — 2026-10-19 — Phase 5 CLOSED: T-64 COMPLETE (the differential corpus full run, all 31 generators × two passes GREEN) + T-65 COMPLETE (the v0.92.0-go.1 tag + zip + the GUI-side launchers + build/mkzip) — 1927 tests race-clean over 34 packages; the twenty-sixth full wipe repaired before work (plus a ~40-minute mid-session tool outage waited out)

## Goal (from NEXT-SESSION.md at start)

SESSION-032 = **T-64 the differential corpus full run + fix loop** (M)
per doc 09 §3 (the case inventory — the corpus's authority) + the
task-list row, then **T-65 the tag + zip** (XS) per doc 10 §2/§6/§8 if
capacity remains: resolve the zip-composition question per the specs
FIRST (surface, don't improvise), write the GUI-side launcher
equivalents, land the v0.92.0-go.1 tag + the GitHub release with the
zip asset.

## What was done (task IDs, files, commits — cite hashes)

- **The twenty-sixth full wipe repaired BEFORE work**: both repos, the
  Go toolchain, and the venv were gone. Rebuilt per the recipe: both
  repos cloned with the ephemeral PAT header (via `GIT_CONFIG_COUNT`
  env — no header persisted into either clone's config; verified by
  direct config read) — the work repo at origin/main `4d2f4d2`
  (SESSION-031's close commit, in sync), the reference at
  `98f4fc71cfb8…` (byte-equal to testdiff/PINNED-GITCURATOR-REV);
  go1.27.2 downloaded from go.dev and **sha256-verified
  (`ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`,
  byte-equal to go.dev's published JSON)**; the venv rebuilt (Python
  3.12.15, stdlib-only); the symlinks; go.mod verified UNCHANGED (the
  authorized set exactly; zero new dependencies this session). The
  battery re-verified at the pushed HEAD before any code: **1922/1922
  race-clean over 33 packages**. A ~40-minute tool outage mid-rebuild
  (every tool call failing) was waited out with periodic probes; the
  work continued exactly where it stopped.

- **T-64 — commits `42739b3` (the fix loop slice) + `9dc79db` (the
  close)**:
  - **The corpus driver** `testdiff/run_corpus.py` (new): drives ALL
    31 generators under testdiff/goldens/ in TWO full passes against
    the pinned reference and proves the three corpus laws —
    per-generator health (exit 0 + the pinned-rev line + the mtime
    touched-set guard against silent generators), THE TWO-RUN
    BYTE-IDENTICAL PROOF (pass1-vs-pass2 full-tree sha snapshots),
    and COMMITTED == REGENERATED (the tree clean after pass 2) — plus
    the four atomic pairs' own law (PAIR TIE + normalized-content
    stability + the committed pairs stay canonical). Resumable
    (merge-mode pass JSONs, `--only` subsets), `--report` emits the
    machine verdict.
  - **The fix loop's five finds — ALL generator-side determinism
    bugs, ZERO Go-port divergences**:
    1. `dump_githubpipeline.py` recorded the RAW mkdtemp root
       (/tmp/ghgolden-XXXX, 4 path fields) — every regeneration
       produced different bytes (the two-run law was impossible).
       FIX: fold the scratch root to `<TMP>` at payload end; the Go
       replay's ghgoldenRootRE fold makes it a no-op (replay green).
    2. `dump_websitepipeline.py` — same bug (12 path fields incl.
       DB-dumped note_path rows arriving as TUPLES — the tuple branch
       was the fold's second iteration); same fix.
    3. `dump_directorynote.py` — same bug (59 path fields); same fix.
    4. `dump_mirror.py` — the collect_note_fields duplicate fixture
       had two DIFFERENT stems; the reference keeps the FIRST hit
       over RAW os.walk order (doc 09 §5's quirk), so the golden's
       recorded winner flipped between machines. FIX: the duplicate
       keeps the SAME STEM in a different folder; the Go golden
       replay's fixture mirrors it (still a full byte-comparison —
       now pinning the order-independent contract).
    5. The runner's own pair-stability digest under-masked the
       notereads/websitestate pairs' per-run facts (the notereads
       rows carry the scratch root + mtime_ns ints + scanned_at
       epoch floats; the websitestate dismissed-ledger tie order is
       UNSPECIFIED ON BOTH SIDES — the reference's
       purge_blocked_domains iterates a Python SET under a tied
       seconds-precision dismissed_at; the Go port faithfully mirrors
       with a map). The digest now masks exactly those documented
       per-run facts (the goldens stay RAW — the Go replays prove
       each pair's internal consistency through the exact frozen
       values); proven stable over 6 consecutive regenerations of
       each previously-flaky pair.
  - **The three legs GREEN**: the generators (31 × 2 passes, verdict
    GREEN, rev guard OK on every run), the Go replays
    (`go test -race -count=1 ./...` = 1922/1922, 0 FAIL/SKIP/races,
    33 packages), the CLI-level harness (status-no-config PASS —
    python/go stdout sha256 identical, 821 bytes).
  - **The report archived** at
    `docs/poc-artifacts/t-64-corpus-full-run-report.md` (the verdict,
    the three legs' evidence, the per-generator table, the
    corpus↔stone attribution, the fix-loop record, the reproduce
    commands).
  - **A reality-wins doc fix in the same commit**: the go1.27.2
    tarball sha256 recorded in NEXT-SESSION §1 + SESSION-031's report
    was a 63-CHARACTER transcription (one dropped character); the
    true 64-char value verified against go.dev's published JSON and
    recorded in NEXT-SESSION §1, build/TOOLCHAIN.md Am.2, and
    SESSION-031's post-close correction note.

- **T-65 — commit `a39cb3d`** (three slices + the release):
  - **The zip-composition question RESOLVED per the specs** (surfaced
    with citations, not improvised): the MINIMAL deterministic mkzip
    slice landed NOW — doc 10 §2's laws are fully specified, F5's
    letter assigns "the zip flattening" to T-65's packaging, T-65's
    own acceptance demands the asset, and the owner's law says a
    release without its zip is not done — while T-80/Phase 7 keeps
    the FULL builder per traceability F4 (the double-build
    reproducibility ritual as CI + the two-exe GUI composition + the
    release automation).
  - **Slice A — the GUI-side launcher pair at the repo root**:
    GitCurator.bat (the desktop door; the reference's grammar with
    the python ladder replaced by an HONEST exe-missing card —
    GitCurator.exe ships with the GUI milestone; the card points at
    the CLI) + GitCurator-DRY-RUN.bat (the SAFE first run, fully
    functional today: `GitCurator-cli.exe --cli --auto --yes
    --dry-run` + the reference's own finished card). Both ASCII+CRLF
    at byte level (1373 + 1197 bytes).
  - **Slice B — the differential stone grew**: dump_cli.py records
    the reference's OWN two app-root .bats raw (the same recording
    law) — the golden at 105,092 payload bytes (was 101,777),
    two-run byte-identical; the replay grew the per-contract goPath
    (the GUI-side pair reads the repo ROOT) + the two contracts
    (shared grammar byte-compared, every divergence asserted both
    ways: the python ladder vs the exe guards, the run lines, the
    gitcurator\cli\ path lines vs the root paths, the codepage
    comment, the Settings REM); a FOURTH mutation surface (the
    DRY-RUN run line).
  - **Slice C — build/mkzip + the assets**: `build/mkzip/main.go`
    (the deterministic builder: fixed timestamps 2026-09-29 00:00:00
    UTC, sorted arcnames, 0o644 attrs, DEFLATE-9, tracked-files-only,
    the banned-secrets guard, .bat ASCII-assert + idempotent CRLF
    normalization; the CLI composition: the exe + the flattened
    launchers + the GUI-side pair + the assets + the stamped VERSION)
    + `main_test.go` (5 law tests: composition, determinism laws,
    byte-identity, the banned guard over a scratch git repo, the .bat
    normalization battery) + config.example.json (the reference's own
    bytes, sha256 a9c98238…) + WINDOWS-QUICKSTART.md (the Go
    first-run guide — the reference's structure, honest about the CLI
    milestone).
  - **THE RELEASE**: the exe built twice byte-identical
    (`760942e4a69c3d032901cd7d4843ea61e62ac679aacb921b2e49b25cb7e1d6f6`);
    the zip built twice byte-identical —
    **GitCurator-v0.92.0-go.1-windows.zip, 12 files, 15,988,399
    bytes, sha256
    `d2679a1de9d600e11d2e590016ed5359ca3fe0f841802289506f4cdd43fecebb`**;
    the tag `v0.92.0-go.1` at `a39cb3d` pushed; GitHub release
    **416838956** published with the zip asset and the sha256 in the
    notes (the reproduce-from-tag recipe included); **the asset
    round-trip download verified byte-identical**.

- **The docs**: the task-list T-64 + T-65 rows COMPLETE with the
  evidence; the traceability F4 (the minimal-slice record + T-80's
  full-builder ownership) + F5 (the GUI-side pair + the zip
  flattening landed); the roadmap's SESSION-032 entries + the
  Phase-5 exit-gate evaluation; the README's status entries.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**T-64 = PASS** — the task-list acceptance "all cases green
(normalized); report archived":
- All 31 generators green on both passes (exit 0, the pinned-rev
  line, non-empty write evidence); the two-run byte-identical proof
  held (empty full-tree snapshot diff beyond the four atomic pairs'
  documented churn); committed == regenerated (the tree clean after
  pass 2 beyond the pairs + the driver itself); the four atomic
  pairs: tie OK + normalized-content stable both passes. The "fix
  what diverges" clause: five divergences found, ALL fixed in-commit
  (reality wins), none a Go-port divergence.
- The Go replays green: 1922/1922 race-clean over 33 packages.
- The harness case green: status-no-config PASS (sha-level outputs).
- The report archived: docs/poc-artifacts/t-64-corpus-full-run-report.md.

**T-65 = PASS** — the task-list acceptance "GitHub release with
asset":
- The zip-composition decision surfaced with the spec citations (the
  minimal mkzip slice now, T-80 keeps the full builder) — recorded in
  the task-list row + the traceability F4 + this report.
- The GUI-side launchers written with the byte laws + the golden's
  launchers section extended (the reference's own app-root pair
  recorded, divergences asserted both ways, a fourth mutation
  surface) — the differential green first-run.
- The release LIVE: tag v0.92.0-go.1 at a39cb3d; release 416838956
  with GitCurator-v0.92.0-go.1-windows.zip (15,988,399 bytes, sha256
  d2679a1d… in the notes); the double-build proof (exe ×2 identical,
  zip ×2 identical) and the round-trip download verification.

**The Phase-5 exit gate = GREEN — PHASE 5 CLOSED**: the owner can run
the Go CLI alongside the Python app on real data (dry-run first — the
DRY-RUN launcher ships in the zip); the differential harness is green
on the full case list (the corpus report); the CLI parity release
0.92.0-go.1 is tagged with the zip artifact.

**Arithmetic**: 1922 (start) + 5 (the mkzip package's law tests) =
**1927** — closes exactly; **34 packages** (build/mkzip); gofmt/vet
clean; zero new dependencies.

**Honest limits**:
- The v0.92.0-go.1 zip is the CLI-milestone composition (12 files —
  no GUI exe; GitCurator.bat ships with an honest card pointing at
  the CLI until Phase 6 lands the desktop exe). The full doc-10-§6
  layout (GitCurator.exe + the two-exe zip) is the GUI-parity
  release's.
- The clean-VM runtime legs (the launchers + the exe on a real
  Windows box) remain Phase 7's checklist (doc 10 §7) — unchanged
  from T-63's record.
- The corpus driver's mtime touched-sets can be inflated by
  concurrent filesystem activity (informational only — the sha-level
  proofs carry the law); the silent-generator guard (empty touched
  set) is unaffected.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-64 and
  T-65 rows COMPLETE with the evidence.
- `docs/port-blueprint/traceability-matrix.md` — F4 (the minimal
  mkzip slice landed; T-80's full-builder ownership preserved) + F5
  (the GUI-side pair + the zip flattening landed).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-032
  progress entries + the Phase-5 exit-gate evaluation (GREEN, the
  books empty).
- `README.md` — the SESSION-032 status entries (T-64 + T-65 + PHASE
  5 CLOSED).
- `build/TOOLCHAIN.md` — **Am.2**: the exact verified go1.27.2 tarball
  sha256 (the 63-char transcription corruption fixed — reality wins).
- `docs/NEXT-SESSION.md` — §1's recipe sha corrected (rewritten in
  full at this close anyway).
- `docs/sessions/SESSION-031.md` — the post-close correction note
  (the sha transcription; the verification itself was sound).
- `docs/poc-artifacts/t-64-corpus-full-run-report.md` — NEW: the
  archived corpus report (T-64's acceptance artifact).

## Blockers & decisions needed

- None technical. The zip-composition decision (the minimal mkzip
  slice at T-65, the full builder at T-80) is recorded with its spec
  citations for the owner's review — the release is live either way.

## Owner actions requested

1. Optional: try the release — download
   `GitCurator-v0.92.0-go.1-windows.zip` from the release page
   (verify the sha256 against the notes), unzip, 1-INSTALL → Setup →
   DRY-RUN alongside the Python app.
2. The go-directive acknowledgement (ADR-0002 Am.2) still waits,
   twenty-second session carrying it; informational.

## Handoff

- Next session: SESSION-033 = **Phase 6's entry — the GUI shell
  (PoC C: the Wails themed window + log desk + one job, T-70)** per
  the roadmap's Phase-6 entry conditions.
- NEXT-SESSION.md rewritten: yes (v33).
- The next prompt: (verbatim block in NEXT-SESSION.md §5).
