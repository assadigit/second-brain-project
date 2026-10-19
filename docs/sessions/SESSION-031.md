# SESSION-031 — 2026-10-19 — Phase 5: T-61 COMPLETE (the close-out audit) + T-63 COMPLETE (THE LAUNCHERS + THE DISPATCH FLIP) — 1922 tests race-clean over 33 packages; the twenty-fifth full wipe repaired before work (the recovery took the full recipe: the repos, the sha256-verified toolchain, the venv, the reference at the pinned rev — the battery re-verified at the pushed HEAD before any code)

## Goal (from NEXT-SESSION.md at start)

SESSION-031 = **T-61's close-out audit + T-63 the launchers** (both S —
one session fits both) per spec-cli's Shims+launchers section + doc 10
§2-6: audit what T-61's row still demands (the Windows-console
verification legs; the clean-VM legs belong to Phase 7) and close the
row with evidence or record the exact honest remainder; land
cmd/gitcurator-cli (the arg wiring → Run, the exit pass-through), the
.bat set (ASCII+CRLF, the walk-up, the exe DIRECT, the reference's own
defaults, EXITCODE echo + pause), gitcurator-cli.sh, the
1-INSTALL-equivalent checks, the VERSION stamp + the build command
shape; extend the differential stone if the launchers have byte-level
surfaces the reference owns (the .bat set IS byte-anchored). The
capacity clause: one slice done properly with proof beats a rushed
whole.

## What was done (task IDs, files, commits — cite hashes)

- **The twenty-fifth full wipe repaired BEFORE work**: the sandbox had
  wiped both repos, the Go toolchain, and the venv. Recovery per the
  recipe: the work repo re-established by init+fetch (never
  `git clone -c` — the header is never persisted) at origin/main
  `580a2ae` (≥ the required `02d4ed8`; the tip carries the two additive
  NEXT-SESSION v31 polish commits), checked out clean; the reference
  re-cloned at `98f4fc71cfb8…` (byte-equal to
  testdiff/PINNED-GITCURATOR-REV); go1.27.2 downloaded from go.dev and
  hash-verified (`ecadb99091a3f46e31f5f934b068b1864eafa7995211b39
  eaddf76996045fe5` — byte-equal to go.dev's published sha256; the
  recorded prefix `ecadb99…` matches); the venv rebuilt (Python
  3.12.15, stdlib-only — PyQt6/requests absent); the symlinks; go.mod
  verified UNCHANGED (the authorized set exactly; zero new
  dependencies this session). **The battery re-verified at the pushed
  HEAD before any code: 1911/1911 top-level tests race-clean over 32
  packages** (exit 0, 0 FAIL / 0 SKIP) — SESSION-030's record
  reproduced exactly.

- **T-61's close-out audit — commit `d66cb2a`**: the row's four demands
  audited against the landed evidence (the paint box with the
  NO_COLOR/TTY/--no-color law + term_width's COLUMNS law +
  StatusLine/render_bar/spinner — terminal.go, SESSION-027; the
  spinner under a REAL loop — the engine's 100 ms heartbeat ticker,
  SESSION-030; the 65001 hardening — platform_windows.go's
  SetConsoleOutputCP/SetConsoleCP via syscall). New compile-level
  evidence taken this session: `GOOS=windows GOARCH=amd64
  CGO_ENABLED=0 go build ./...` + `GOOS=windows go vet ./...` both
  clean. The exact honest remainder recorded where it belongs: the
  RUNTIME Windows-console legs (an actual console on 65001 rendering
  the braille/box glyphs + the VT sequences live) are Phase 7's
  clean-machine checklist items (doc 10 §7; T-70's VM). Plus **ADR-0007
  Am.2** (append-only): Am.1 had cut the owner-facing CLI polish; the
  owner's SESSION-027..031 instruction line re-instated it — the
  amendment records the reversal so the ADR stops contradicting the
  practice (flagged for the owner's veto in this report).

- **T-63 slice 1 — commit `e89eaaa`**: `cmd/gitcurator-cli/main.go` (the
  main package: **the --cli strip law ported verbatim from main.py's
  routing** — `[a for a in sys.argv[1:] if a != "--cli"]` — so the
  launchers keep the reference's own `--cli --auto --yes` run lines
  and muscle memory keeps working; the **--version stamp door** per
  doc 10 §5 — `gitcurator-cli <version>`, exit 0, before the parser;
  a documented divergence: the reference's argparse has no
  --version; the `version` var stamped via `-ldflags "-X
  main.version=$(cat VERSION)"`, fallback "dev"; the exit code passes
  through from cli.Main); `VERSION` (repo root, `0.92.0-go.1` — the
  milestone T-65's tag will carry); **THE LAUNCHER SET** in
  cmd/gitcurator-cli/ (the owner's folder law; T-65's mkzip flattens
  them to the zip root): GitCurator-CLI.bat (`--cli --auto --yes`),
  GitCurator-SYNC.bat (`--cli --sync %*`), GitCurator-CLI-Setup.bat
  (`--cli --init`), Start-GitCurator-CLI.bat (`--cli --auto --yes %*`
  + the owner's EXITCODE-card normalization the reference's v0.09
  alias lacked), gitcurator-cli.sh (set -euo pipefail, the walk-up,
  the exe guard card, the same default + the pass-through),
  1-INSTALL.bat (the Go equivalent: `where git` + `git --version` +
  the next-steps card; NO registry writes, NO PATH mutations —
  nothing to install); `cmd/gitcurator-cli/launchers_test.go` (the
  byte laws: every .bat ASCII-ONLY + CRLF — no byte > 0x7F, every
  newline CRLF, the final CRLF; the .sh LF-only UTF-8 + executable;
  the grammar skeleton with the walk-up landmark VERSION; every
  python-discovery command line banned; chcp banned; the EXITCODE
  echo + pause + endlocal tail; the git-only checker with the
  no-mutation bans) + `main_test.go` (the version door + the stamp
  feeds it + the strip law).

- **THE DISPATCH FLIP — in `e89eaaa` (the session's reality-wins
  find)**: reading cli.go for T-63's arg wiring revealed that the
  T-62 drivers had LANDED but Run's dispatch still routed `--auto`,
  `--sync`, `--retry-failed`, `--import-file` to not-yet cards — the
  shipped binary would have refused its flagship mode (the unit
  battery called the cmd* methods directly, so the suite stayed
  green; TestDispatchOrder even pinned the stale card). Flipped:
  --auto → cmdAuto, --sync → cmdSyncLive, --retry-failed →
  cmdRetryFailed, --import-file → cmdManualBatch; the not-yet cards
  and the beat-1a cmdSync stub retired; the golden's
  sync_no_config/sync_bad_config sections byte-unchanged (cmdSyncLive
  opens with the same beat-1a check); TestBatchDispatchSmoke added
  (the four doors route to the REAL drivers — their own honest cards
  answer, never a not-yet card).

- **T-63 slice 2 — commit `abda6cf` (THE DIFFERENTIAL STONE)**:
  dump_cli.py's payload now reads the reference's OWN launcher bytes
  (the six files) and records them RAW — the full text with CRLF
  preserved inside the JSON strings + the byte-level facts —
  goldencli.golden.json at **101,777 bytes** (the growth line
  48,698 → 69,028 → 86,570 → 101,777), two-run byte-identical. The
  replay (TestGoldenCLILaunchers) reads the Go repo's launcher files
  and enforces the per-file CONTRACT: the SHARED lines byte-compared
  (the skeleton + the exit cards + the .sh's usage block), the
  DIVERGENT pairs asserted BOTH ways (each present in its own side
  and absent from the other: the walk-up landmark main.py → VERSION,
  the python invocation → the direct exe run, 1-INSTALL's
  Using/version lines, the sh's exec lines), the ref-only python-world
  lines banned from the Go side, the go-only exe-world lines present,
  the reference's own laws on its recorded facts + the Go side's
  recomputed from raw bytes. The mutation teeth
  (TestGoldenCLIMutationLaunchers): three surfaces (a mutated
  reference run line, a mutated shared card line, a flipped
  byte-level fact) each FAIL the real comparison while the unmutated
  golden still passes — verified biting (the comparator caught two
  contract-table errors on its first run: the reference's if-block
  indents on `set "PY=py -3"` and the sh's default exec — the teeth
  work in development too).

- **The docs**: the task-list rows (T-61 + T-63 COMPLETE with the
  evidence), the traceability E9 tail + F5 (the Windows launch
  contract), the roadmap's SESSION-031 progress entry, the README's
  status entry, spec-cli's two SESSION-031 shipped records.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**T-61 = PASS — the row closed with evidence** (its acceptance
"Windows console test" evaluated at the compile level here):
- ANSI/NO_COLOR paint: TestEnableColorsDoor + the golden's
  paint_shapes (colors ON and OFF, byte-compared) + the paint unit
  battery — green.
- Spinner/bar: TestStatusLineGrammar/BarWidth +
  TestRenderBarBankersRounding + engine_test.go's heartbeat proof
  (the ticker under a real loop) — green.
- UTF-8 hardening (65001): platform_windows.go's syscall doors +
  hardenStdio at Run's start + `GOOS=windows GOARCH=amd64
  CGO_ENABLED=0 go build ./...` + `GOOS=windows go vet ./...` clean
  (2026-10-19). The runtime console legs (an actual Windows console
  switching to 65001 and rendering the glyphs) are recorded as Phase
  7's clean-machine checklist items — per the kickoff's own scoping
  ("the clean-VM legs belong to Phase 7, not here").

**T-63 = PASS — the launchers complete per spec-cli's Shims+launchers
+ doc 10 §2-6**:
- cmd/gitcurator-cli: the arg wiring (the --cli strip law), the exit
  pass-through, stdin/stdout wired to the real TTY doors (NewCLI's
  binding), the --version stamp — main_test.go green.
- The .bat set: ASCII+CRLF BOTH asserted in tests (byte level), the
  walk-up kept (landmark VERSION — the Go app root's marker; doc 10
  §6's "walk-up cd removed" note superseded by spec-cli's
  "unnecessary but harmless — keep" + the kickoff's explicit shape;
  the roadmap note stands as history), the exe run DIRECT with the
  reference's own defaults, the EXITCODE echo + pause on all four
  runners (Start normalized per NEXT-SESSION v31's letter).
- gitcurator-cli.sh: set -euo pipefail, the walk-up, the exe, the
  default — green.
- The 1-INSTALL equivalent: git.exe on PATH checked, the next-steps
  card, NO registry writes, no PATH mutations — asserted in tests.
- The VERSION stamp + the build command shape: VERSION lands; the
  stamp mechanism proven by TestVersionStampFeedsTheDoor; the release
  packaging itself is T-65's.
- The differential: the golden's launchers section + the replay + the
  mutation teeth (three surfaces, verified biting).
- GOOS=windows build of ./cmd/... clean (and the whole module).
- The acceptance column's "clean-VM smoke" — the same Phase-7 scoping
  as T-61's runtime legs (doc 10 §7's checklist runs the launchers on
  a real Windows box; recorded, not improvised).

**Arithmetic**: 1911 (SESSION-030) − 1 (TestNotYetCards retired) +
10 (the cmd package's 9 + TestBatchDispatchSmoke) + 2 (the golden
launchers tests) = **1922** — closes exactly; **33 packages** (the
new cmd/gitcurator-cli test package); gofmt/vet/GOOS=windows clean;
zero new dependencies.

**Honest limits**:
- The GUI-side launchers (GitCurator.bat, GitCurator-DRY-RUN.bat at the
  reference's app root) are NOT in T-63's set (the kickoff lists the
  CLI set); they are recorded for T-65's zip (the DRY-RUN one runs
  `--cli --auto --yes --dry-run`, which the Go CLI supports).
- T-65's zip composition question (mkzip is Phase 7's T-80 per the
  traceability F4, but T-65's acceptance wants a GitHub release with
  an asset) is left for T-65's session to resolve per the specs —
  surfaced, not improvised.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/decisions/0007-cli-first-order.md` — **Am.2**
  (the re-instatement record; append-only).
- `docs/port-blueprint/module-specs/spec-cli.md` — the two SESSION-031
  shipped records (T-61's audit; T-63's inventory + the differential +
  the Option-B divergences) + this session's new GO NOTE (below).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-61 and
  T-63 rows COMPLETE with the evidence.
- `docs/port-blueprint/traceability-matrix.md` — E9's status tail +
  F5 (the Windows launch contract).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-031
  progress entry.
- `README.md` — the SESSION-031 status entry.

## New house law (this session's lesson, in spec-cli's GO NOTES)

**THE DISPATCH-COMPLETION LAW**: a command is only LANDED when Run's
dispatch routes the argv to it — unit tests that call the cmd* methods
directly keep the suite green while the user-facing door stays stale
(T-62's drivers landed in SESSION-030; T-63's smoke through Run caught
`--auto` still answering the retired not-yet card). The acceptance for
a CLI command drives `Run(argv)` — the smoke test is not optional
decoration, it is the completion proof.

## Blockers & decisions needed

- None technical. The ADR-0007 Am.2 record (the re-instatement) is the
  owner's decision written down by the single writer per the standing
  law — flagged here for the owner's veto if the intent was otherwise.

## Owner actions requested

1. None blocking. (The go-directive acknowledgement — ADR-0002 Am.2 —
   still waits, twenty-first session carrying it; informational.)
2. Optional veto: ADR-0007 Am.2's framing (see Blockers).

## Handoff

- Next session: SESSION-032 = **T-64 (the differential corpus full run
  + fix loop per doc 09 §3) then T-65 (the `v0.92.0-go.1` tag + zip) if
  capacity remains** — Phase 5's last two books.
- NEXT-SESSION.md rewritten: yes (v32).
- The next prompt: (verbatim block in NEXT-SESSION.md §5 — T-64 first,
  T-65 second; the push state CLEAN; PUSH AFTER EVERY TASK).
