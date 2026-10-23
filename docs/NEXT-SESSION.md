# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker
  HTTP API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity.
- **Phase 2 CLOSED** (SESSION-006–010): T-20…T-25.
- **Phase 3 CLOSED** (SESSION-011–014): T-30…T-34.
- **Phase 4 CLOSED (SESSION-015–026)**: T-40…T-50 + the Phase-4 closer
  (T-51 removed by the owner's Option-B stamp).
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity
  release `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip
  asset sha-verified + round-trip verified).
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70
  gate through the T-79a…T-79f Path-B rows.
- **Phase 7 — OPEN, THE FIX CYCLE UNDER THE OWNER'S DIRECTIVE
  (SESSION-048–074)**: T-80 ✓ (the deterministic zip + the ritual),
  T-81 ✓ (the release workflow, twice proven on real releases now),
  T-83 ✓ (the docs; the QUICKSTART reframed at SESSION-074), AC-17 ✓,
  **THE DESKTOP MILESTONE RELEASE `v0.92.0-go.2` CUT ✓ (SESSION-070)
  — AND THE OWNER'S FIRST LIVE RUN FAILED (SESSION-074, 2026-10-24:
  "I have downloaded the latest github release. none of them works,
  neither cli or GUI" — their machine, no error text; the intake +
  the diagnosis record: the checklist §1 + the go.2 report's table;
  THE SHIPPED BYTES PROVEN CORRECT three ways — the fresh download
  byte-equal, the exes byte-equal, PE-inspected; the root cause on
  their machine UNDETERMINED: SmartScreen/AV/extract-first/
  pre-Win10-64/Windows-only-bug, the last not excludable without
  their error text)**. **THE FIX RELEASE `v0.92.0-go.3` CUT ✓
  (SESSION-074) under the owner's directive ("Please make sure the
  next version both CLI and GUI WORKS!" — portable, no install): the
  bump `5e0c09d` → the tag → the release run 40022070760 green →
  release 422298475 LIVE (14 files, 56,935,578 in / 27,237,300
  zipped, sha256 `65eb64b4…`, verified three ways incl. the
  cross-machine byte-identity; `0-START-HERE.txt` FIRST in the zip +
  the QUICKSTART reframe + the release-notes First-run section; the
  go.3 report: `docs/poc-artifacts/release-v0.92.0-go.3.md`)**.
  **THE PHASE-7 EXIT EVALUATION RUN AT SESSION-074: NOT GREEN** (the
  clean-VM demonstration failed on the owner's machine + no owner
  sign-off; the gate needs their live legs passing + the sign-off —
  the go.3 re-run slots are prepared in the checklist §1). What
  remains in Phase 7: the owner's go.3 re-run (+ their exact error
  text if anything still fails — THE ASK) and, if green, the exit
  evaluation re-run; if a machine gate is named, the corresponding
  owner decision (the code-signing cert — doc 10 §9 rung 4, the only
  SmartScreen-killing rung; the fully-portable bundle — WebView2
  fixed + MinGit inside the zip, ~150 MB). The debt books are EMPTY.
  **The Phase-8 menu stands surfaced in full from SESSION-064** (doc
  12's register); with the fix cycle open, the owner's go.3 re-run is
  the first mover; no autonomous implementation slice exists.
- **Push state: CLEAN** — SESSION-074's pushes: `8deaf29` (the
  hardening) + `97ad2fa` (the intake) + `5e0c09d` (the bump, the tag
  `v0.92.0-go.3` riding it) → go-ci run 40022070901 + release run
  40022070760 both GREEN (watched to completion); **SESSION-074's
  close push (docs-only: the reports + this card + doc 12's entry +
  the README entries + the worklog) is the latest — verify its go-ci
  run green at SESSION-075's open (a red is a REAL failure; watch
  the Actions tab on pushes as normal CI hygiene)**.
- **The repo is PUBLIC** (the owner's SESSION-050 instruction): pushes
  still need the PAT (ephemeral, as always — re-inline it per call;
  the sandbox shell does not carry exports between tool calls);
  Actions minutes are unlimited; the no-secrets law is public-facing.
  **go-ci: GREEN on the tree through SESSION-074** (the ledger: the
  go.3 release run 40022070760 + the bump → 40022070901 + `0e59425`
  → 40000283140 + `d00ea95` → 39980891673 + `c241bc8` → 39953390474;
  the close push is the latest: verify at the open).
- **The environment recipe** (the 68th wipe absorbed at SESSION-073;
  SESSION-074 absorbed a PARTIAL wipe — the toolchain/venv gone, the
  repo + the reference surviving, go re-extracted from the surviving
  sha-verified tarball, the venv + the wails CLI rebuilt): work repo
  at /home/z/my-project/second-brain-project (**clone it
  ANONYMOUSLY — it is public; the credentials block's GitHub URL may
  point at the PRIVATE Python reference `assadigit/GitCurator` — do
  not confuse the two**; scrub any persisted token from the remote
  URL — the ephemeral `-c http.extraheader` clone form avoids
  persistence entirely), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (must match
  testdiff/PINNED-GITCURATOR-REV; the GIT_CONFIG_COUNT / `-c`
  env-var forms avoid persistence entirely), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball — extract WITHOUT
  --strip-components; sha256 `ecbadb99091a…45fe5`; verify BOTH sides —
  the COMMITTED pin in build/TOOLCHAIN.md Am.2 + the archive listing
  `?mode=json&include=all`; curl needs `-L`), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only), the wails CLI
  v2.16.0 (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0`), and the rootless
  webkit tree (~/.local/lib-wails — **DOWN; last rebuilt at
  SESSION-047; the SIXTEENTH rebuild is pending: rebuild it FIRST
  only when the session runs a LINUX GUI build**). **NEW (the
  SESSION-074 partial-wipe class): a filesystem restore may arrive
  with EVERY tracked file mode-flipped 100644→100755 (content
  unchanged — `git diff` shows 0 insertions/deletions); the repair:
  restore every tracked file's mode FROM THE INDEX —
  `git ls-files -s | awk '$1 == "100644" {print $4}' | tr '\n' '\0' |
  xargs -0 chmod 644` and the 755 twin — then verify `git status`
  clean.** The MAIN go.mod: modernc.org/sqlite v1.36.1,
  bogdanfinn/tls-client v1.16.0, golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 direct, the go directive 1.24.1
  (ADR-0002 Am.2; do not bump). The NESTED webui module rides go
  1.25.0 (wails v2.16.0's floor; ADR-0002 Am.5; the MAIN go.mod stays
  byte-unchanged). **The downloads must run in FOREGROUND CHUNKS.**
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them): **THE OWNER-DIRECTIVE RECORD
  (SESSION-074 — the go.2 live failure + the fix directive: the
  intake is recorded verbatim-honest even when the results are
  negative; the diagnosis runs to the honest CEILING (bytes proven,
  machine-gate candidates listed, the not-excludable named); the fix
  releases reuse the proven ritual; THE ASK (the exact error text +
  winver + the AV name) is the blocker's key)**; **THE
  PARTIAL-WIPE MODE-FLIP CLASS (SESSION-074 — a restored tree may
  arrive all-755: repair from the INDEX, not by hand)**; **THE
  MULTIEDIT-TAB CORRUPTION, SIXTH STRIKE (SESSION-074 — the Edit
  tool rewrote BOTH mkzip files' tabs to spaces (the whole files,
  637+699 lines); caught by the battery's gofmt leg; repaired
  python-on-bytes with anchor assertions — one anchor failed loud as
  designed, one brace-balance slip caught by vet before commit; **GO
  FILES ARE PYTHON-ONLY TERRITORY — never the Edit tool, even for a
  one-line insert)**; **THE RELEASE-RECORD (SESSION-070, re-proven
  at 074 — the bump → the tag on the bump commit → the workflow
  green → the release verified THREE ways incl. the cross-machine
  byte-identity; every future release reuses the proven ritual
  verbatim)**; **THE MENU-SURFACED RECORD (SESSION-064, exercised
  at SESSION-065…069 + 071–073)**; **THE ENVIRONMENT-OWNED-TEXT LAW
  (SESSION-063)**; **THE BOOTSTRAPPER-LEG RECORD (SESSION-063 — the
  1-INSTALL WebView2 check is READ-ONLY; the Evergreen GUID is
  `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`; from go.3 the zip also
  carries 0-START-HERE.txt first)**; **THE DEBT-BOOKS-EMPTY RECORD
  (SESSION-063)**; **THE TEXT-MODE-READ LAW (SESSION-061)**; **THE
  MULTIEDIT-TAB CORRUPTION, FIFTH STRIKE (SESSION-063 — anchor
  assertions are for prose too)**; **THE HARVEST-COUSIN RECORD
  (SESSION-060)**; **THE DELIVERED-LAW RECORD (SESSION-059)**; **THE
  MULTIEDIT-TAB CORRUPTION, THIRD STRIKE (SESSION-060)**; **THE
  ASYNC-READ FLAKE LESSON (SESSION-059, SECOND STRIKE
  SESSION-068)**; **THE LIVE-LEG RECORD (SESSION-058)**; **THE
  68TH-WIPE RECIPE (SESSION-073)**; **THE WEBKIT-ORDINAL CORRECTION
  (SESSION-055 — the pending rebuild is the SIXTEENTH)**; **THE
  AC-17 BENCHMARK RECORD (SESSION-052)**; **THE HONEST-ZIP RECORD
  (SESSION-051 — RESOLVED at SESSION-070)**; **THE
  STALE-COMPARISON-TARGET LAW (SESSION-050)**; **THE PUBLIC-REPO
  RECORD (SESSION-050)**; **THE TAG-STATE BUILD LAW (SESSION-049,
  re-proven at 070 + 074 — the CLI exe's buildinfo carries `mod …
  v0.92.0-go.3`)**; **THE -X-UNDER-TRIMPATH LAW (SESSION-048)**;
  **THE VERIFY-THE-VERIFIER LESSON (SESSION-048–074)**; **THE
  DEFAULT-ICON FIND (SESSION-048)**; **THE FLAG-CLEAR LAW
  (SESSION-047 — now includes the partial-wipe mode-flip repair)**;
  **THE REFERENCE'S-OWN-ARITHMETIC LAW (SESSION-046/048)**; **THE
  VLM MISREAD LEDGER (FOUR incidents: 042/043/044/045; 046–074 ran
  clean)**; **THE BINDING-IMPORT LAW (SESSION-044)**; **THE
  RACE-PASS DIVIDEND (SESSION-045)**; **THE EXTRACT-AND-EXEC SCOPE
  LAW (SESSION-041)**; **THE CROSS-BINARY DISAGREEMENT LAW**; **THE
  SENTINEL-SEAT RACE LAW (SESSION-040)**; **THE RETROACTIVE
  FIELD-TAG AUDIT (SESSION-040)**; **THE NIL-OPTIONAL DEREF
  (SESSION-040)**; **THE ONE-OWNER-PER-SURFACE law (SESSION-040)**;
  **THE FIELD-TAG LAW (SESSION-038)**; **THE DISPLAY-PIPELINE
  CORRUPTION LESSON (SESSION-038)**; **THE SYMLINK-BLOCK FALLBACK
  (TOOLCHAIN.md Am.4)**; **THE RENDER-CONTEXT LAW (SESSION-037)**;
  **THE VLM RATE-LIMIT WINDOW (SESSION-039)**; **THE
  PINNED-VERDICT-ROW LESSON (SESSION-039)**; **THE
  ONE-SLICE-PROPERLY RECORD (036–074)**; and the standing set (THE
  LIVE-CONFIG LAW; THE GROUP-POINTER LAW; THE DISSOLUTION-PROBE
  RECORD; THE GOLDEN-MARSHAL-ORDER TRAP; THE SUITE-REVERIFICATION
  DIVIDEND; THE WAILS-BUILD MODE FLIPS; THE HERMETIC-SEAM LAW; THE
  RESUMED-CONTEXT RE-VERIFICATION LAW; THE MODE-FLIP DEGRADATION
  LAW; THE MODULE-BOUNDARY PROBE LAW; THE FRONTEND-MODULE LESSON;
  the T-65 zip-composition resolution; the corpus-regeneration law;
  the dispatch-completion law; the state-chain + mutation-teeth
  laws; the context-death continuation recipe; the DSN-empty-path
  hazard; the fixed-length roots; the shallow-copy snapshot; the
  fake-clock arithmetic; the exception class-name tokens; the
  stdout-capture-helper limit).
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd AND its own PATH=/home/z/.local/go/bin export); **THE
    RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F
    FILE); **THE WRITE-TOOL REDACTION lesson** (byte-level
    secret-scan before every commit; inspect hits before declaring
    false positives); **THE BACKGROUND-PROCESS lesson** (long suites
    + downloads in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK**
    (ask for the PAT the moment a push is due; never store).

## 2. Next session: SESSION-075 — the post-fix session (the decision tree, in priority order)

1. **(a) THE OWNER'S GO.3 RESULTS / ERROR TEXT** — if the prompt
   carries their re-run results or their exact error text: the
   intake (the checklist §1's go.3 re-run slots + the go.3 report's
   table); if green → the Phase-7 exit evaluation re-run (the gate:
   the live legs passing + the owner sign-off); if the error names
   a machine gate (SmartScreen/AV/pre-Win10/32-bit) → record it +
   walk the owner through the gate or surface the decision (the
   code-signing cert; the fully-portable bundle); if it names a
   Windows-only bug → that bug is the next fix cycle's target with
   the error text as the anchor.
2. **(b) the standing delegation** ("you choose; no haste to start
   the app") — re-verify the state (the go-ci run at the open — a
   red is a REAL failure; the battery if the environment wiped, the
   exact committed shape 2398 + 95 = 2493), re-surface the open owner
   items + the Phase-8 menu compactly (the full menu stands from
   SESSION-064), and close cleanly at the natural boundary. Never
   start Phase-8 work unilaterally; the owner's standing directive
   (make it work on their machine) outranks the menu until closed.
3. **Blockers**: if anything red surfaces (a CI failure, a broken
   release link), STOP and surface it with evidence per the standing
   laws; record it in the card + report and hand the decision back.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-074.md` — the intake + fix-release
   session's report (the go.2 failure intake + the diagnosis + the
   go.3 cut + the Phase-7 NOT-GREEN verdict + THE ASK).
2. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — §1 (the go.2
   results + the diagnosis record + the go.3 re-run slots) + §3 +
   §4 (the dated notes).
3. `docs/poc-artifacts/release-v0.92.0-go.3.md` — the live release's
   report (the verified numbers + the re-run slots + the AV section
   with the rung-4 code-signing call).
4. `docs/poc-artifacts/release-v0.92.0-go.2.md` — the go.2 report
   (the intake table + the verdict block; the failure's context).
5. `0-START-HERE.txt` + `WINDOWS-QUICKSTART.md` — the shipped
   portable-first surface (what the owner now sees).
6. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + AC-23 + the SESSION-070…074 tail entries.
7. `docs/sessions/SESSION-070.md` — the release ritual's full
   record (only if the ritual's details are needed). The full
   Phase-8 menu's full form stands from SESSION-064's report.

## 4. Known constraints / owner actions for SESSION-074

- **OWNER ACTION #1 (THE 30-SECOND ASK — the blocker's key): the
  exact error text** — whatever the owner sees when an exe "doesn't
  work" (the blue SmartScreen box's wording, an AV message, the
  console behind the app) + their Windows version (`winver`) + the
  AV's name (if any beyond Defender). The shipped bytes are proven
  correct three ways; the failure is a machine-gate question only
  their screen can answer.
- **OWNER ACTION #2: the go.3 re-run** — download
  `GitCurator-v0.92.0-go.3-windows.zip`, **Extract All**, open
  `0-START-HERE.txt` (it sorts first), then double-click
  `GitCurator.bat`; paste any results into the next prompt (the
  checklist's go.3 re-run slots are prepared). Item 1's full check
  (git + the WebView2 probe + the first-launch one-click offer)
  rides it as before.
- **OWNER ACTION #3 (optional): the code-signing cert** — doc 10
  §9's rung 4, the only rung that kills the SmartScreen class at
  the source (annual fee + validation; the go.3 report's AV
  section carries the cost/benefit).
- **OWNER ACTION #4 (optional): the fully-portable bundle** —
  WebView2 fixed-version + MinGit inside the zip (~150 MB total)
  so nothing is ever downloaded and no component can be missing;
  the owner's call if go.3's shape still isn't enough.
- **OWNER ACTION #5: the Phase-8 pick** — only when the fix cycle
  closes (the menu stands from SESSION-064 + compactly at
  SESSION-065–069 + 071–073).
- **The repo is PUBLIC** (SESSION-050) — never commit secrets;
  pushes still need the ephemeral PAT (re-inline per call).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (last rebuilt at SESSION-047; the
  SIXTEENTH rebuild is pending) — rebuild it FIRST only if the
  session runs a LINUX GUI build (the windows legs are CGO-free;
  the full battery runs without it — re-proven at
  SESSION-070–074).
- **GO FILES ARE PYTHON-ONLY TERRITORY** (the sixth strike's law —
  never the Edit tool on a .go file; anchor-asserted python on the
  real bytes; the battery's gofmt leg is the tripwire).
- The dissolution call (ADR-0002 Am.5); vanilla-vs-vite; the
  batch-scorecard modal; the v0.91.0 zip-parity reading; the
  answering-window belt + the T-71 merged-verdict question.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged).
- The go-directive acknowledgement (STILL WAITING): the main
  module rides `go 1.24.1` (the tls-client pin, ADR-0002 Am.2).
- **After every `wails build`**: check `git diff --raw` for mode
  flips; never commit a stray binary (nor a generated notes.md).
- **Deferred test debt on the books**: NONE. The owner's LIVE legs
  remain (the go.3 re-run; the launchers' clean-VM smoke rides it).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (scrub any persisted
  extraheader).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit); tests never weakened; reality wins over specs
  (fix the spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-075)

```
You are starting SESSION-075 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE STATE: the desktop release line is LIVE at v0.92.0-go.3 (the FIX release
cut at SESSION-074 under the owner's directive after v0.92.0-go.2 failed on
their machine — "none of them works, neither cli or GUI", no error text; the
intake + the diagnosis record: docs/poc-artifacts/t-82-clean-vm-checklist.md
§1; the go.3 release report: docs/poc-artifacts/release-v0.92.0-go.3.md — 14
files with 0-START-HERE.txt first, verified three ways incl. the cross-machine
byte-identity; the batteries green 2398 + 95 = 2493; go-ci + the release run
green). The owner's standing directive: "Please make sure the next version both
CLI and GUI WORKS!" — it stands until the app runs on their machine. SESSION-075
is the post-fix session — the decision tree in priority order: (a) if this
prompt carries the owner's go.3 re-run results or their EXACT ERROR TEXT (a
screenshot's words, the blue box's title, an AV message, the console behind the
app): do the INTAKE (the checklist's go.3 re-run slots + the go.3 report's
table); if the results are green, run the PHASE-7 EXIT EVALUATION honestly
against doc 12's gates (T-80…T-83 + AC-17 + the release + the live legs + the
owner sign-off) and record the verdict; if the error text names a machine gate
(SmartScreen / AV / pre-Win10 / 32-bit), record it and walk the owner through
the gate or surface the corresponding decision (the code-signing cert, the
fully-portable bundle); if it names a Windows-only runtime bug, that bug is the
next fix cycle's target with the error text as the anchor; (b) otherwise the
standing delegation ("you choose; no haste to start the app") holds: re-verify
the state (the go-ci run at the open — a red is a REAL failure; the battery if
the environment wiped), re-surface the open owner items + the Phase-8 menu
compactly (the full menu stands from SESSION-064), and close cleanly — never
starting Phase-8 work unilaterally. If anything blocks, STOP and surface it
with evidence per the standing laws; record it in the card + report and hand
the decision back.

Then follow docs/SESSION-PROTOCOL.md exactly: bootstrap (clone
https://github.com/assadigit/second-brain-project — the work repo, public;
read README.md, docs/SESSION-PROTOCOL.md, docs/NEXT-SESSION.md [the card v75],
the two most recent session reports [SESSION-074 is the latest; SESSION-073 is
the standing-delegation session before it], git log --oneline -15), orient (the
read-firsts: the card's §3), work (the standing laws: reality wins, tests
never weakened, no secrets ever, stay inside the phase, push after every
task), close (the honest gate evaluation, the commit + push, the SESSION-075
report, the NEXT-SESSION.md rewrite, the worklog append), and report in chat
with the literal next prompt.
```
