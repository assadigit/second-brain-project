# SESSION-074 — 2026-10-24 — Phase 7: the T-82 INTAKE (path (a) of card v73's tree — the prompt CARRIED the owner's live results, the first live Windows run ever, and they are NEGATIVE) + THE FIX RELEASE v0.92.0-go.3 CUT UNDER THE OWNER'S DIRECTIVE ("Please make sure the next version both CLI and GUI WORKS!" — portable, without install): the go.2 failure recorded verbatim-honest (items 1/3 FAIL, no error text) + the diagnosis (the shipped bytes proven correct three ways; the root cause on the owner's machine UNDETERMINED — SmartScreen/AV/extract-first/pre-Win10 candidates, a Windows-only runtime bug not excludable) + the portable-first reframe shipped (0-START-HERE.txt FIRST in the 14-file zip + the QUICKSTART reframe + the release-notes First-run section) + THE PHASE-7 EXIT EVALUATION: NOT GREEN (the live legs failed; no owner sign-off) — the fix cycle continues under the owner's directive; the battery 2398 + 95 = 2493 green at the close; go-ci + the release run GREEN; the release verified THREE ways

## Goal (from NEXT-SESSION.md at start)

SESSION-074 = THE POST-RELEASE SESSION — the decision tree in
priority order: (a) if the prompt carried the owner's T-82 results
(the nine clean-VM checklist items run against the live
`v0.92.0-go.2` zip): the intake (the checklist doc §1 + the release
report's results table + the AV section + the doc-13 flip, then the
Phase-7 exit evaluation honestly against doc 12's gates); (b)
otherwise the standing delegation. **The prompt carried the owner's
live results — path (a), the first live intake ever — AND their
directive: the results were NEGATIVE ("I have downloaded the latest
github release. none of them works, neither cli or GUI. Couldn't you
just compile it fully and make it working without install? like a
portable app. ... Please make sure the next version both CLI and
GUI WORKS!") — the intake was done AND the fix cycle executed under
the explicit directive (a real owner decision, recorded here + in
the card v75).**

## What was done (task IDs, files, commits — cite hashes)

- **The open's state check** (the conversation carried SESSION-073's
  close in the same sandbox): HEAD `0015860`, go-ci run 40003430900
  verified success at 073's close. A PARTIAL wipe had hit the
  environment between sessions (the go toolchain + the venv gone;
  the repo + the reference survived): absorbed (go1.27.2 re-extracted
  from the sha-verified surviving tarball; the venv recreated —
  Python 3.12.15; the wails CLI v2.16.0 reinstalled) + **a
  filesystem-restore artifact absorbed: all 794 tracked files arrived
  mode-flipped 100644→100755** (content unchanged — `git diff` proved
  0 insertions/deletions) — repaired by restoring every tracked
  file's mode FROM THE INDEX (`git ls-files -s`), tree clean after.
- **THE INTAKE (path (a))** — the owner's results recorded
  verbatim-honest, marking what is NOT known (no error text was
  provided): `t-82-clean-vm-checklist.md` §1 (item 1 FAIL — the GUI
  does not launch; item 2 NOT REACHED; item 3 FAIL — "neither cli or
  GUI"; items 4–7 NOT RUN; item 8 UNKNOWN — no scanner verdict, AV
  interference a live hypothesis; item 9 PASS automation-half) + the
  go.3 re-run slots + **the SESSION-074 diagnosis record** + the
  dated §4 note; `release-v0.92.0-go.2.md`'s results table + the AV
  section filled the same shape; doc 13's T-82 row flipped
  PARTIAL→OPEN. Commit `97ad2fa`.
- **THE DIAGNOSIS (evidence-based, the honest ceiling)** — the live
  go.2 asset re-downloaded fresh: sha256
  `78990ccb…d6c3cd` **byte-equal** to the recorded number; both exes
  extracted **byte-equal** to the parity table (`9d6c2c00…`,
  `a6539a1d…`); PE-inspected (python, byte-level): both PE32+ x86-64
  **AMD64**, `GitCurator.exe` subsystem **WINDOWS_GUI**,
  `GitCurator-cli.exe` subsystem **WINDOWS_CUI**, both stamped
  Windows 10+; the CLI's buildinfo carries `go1.27.2` + `mod …
  v0.92.0-go.2` — **the shipped bytes are exactly the verified,
  reproducible release artifacts; the failure is at the owner's
  machine gate.** wine absent (no Windows execution possible in this
  sandbox — standing since SESSION-001; the exes had NEVER been
  executed on real Windows before this report). The candidates,
  recorded unranked: (a) SmartScreen blocking the unsigned exes;
  (b) an antivirus quarantining/blocking; (c) running from inside
  the zip window; (d) pre-Windows-10 or 32-bit Windows (Go 1.21+
  requires 64-bit Win10+); (e) a Windows-only runtime bug — not
  excludable without the owner's exact error text. THE ASK rides
  §Owner actions + the chat report.
- **THE FIX UNDER THE DIRECTIVE — the portable-first reframe**
  (commit `8deaf29`, after one absorbed corruption strike): NEW
  `0-START-HERE.txt` (pure ASCII + CRLF, the Notepad laws) ships
  FIRST in the zip (mkzip `rootAssets` + the composition test's
  must-list + the new START-HERE determinism block: extract-ALL
  first, the SmartScreen "More info → Run anyway" walk-through, the
  AV restore/allow guidance with the sha256 proof, the 64-bit
  Win10/11 requirement, the GUI start door + the WebView2
  first-launch offer, the CLI terminal invocation card, the
  ask-the-exact-error ask); `WINDOWS-QUICKSTART.md` reframed (no
  "Install (once)" step — the guide leads with extract-first + the
  SmartScreen box + start-the-app; `1-INSTALL.bat` repositioned as
  the OPTIONAL checker it always was; the WebView2/git notes moved
  to the start section; the file table gains both rows); the
  release workflow's notes template gains the First-run section
  (additive; all six workflow law tests green). The launchers, the
  exes' source, and the engine UNTOUCHED (the parity laws hold).
- **THE MULTIEDIT-TAB STRIKE, SIXTH (absorbed within the session)**:
  the Edit tool rewrote BOTH mkzip Go files' tab indentation to
  spaces (637 + 699 changed lines — the whole files; caught by the
  battery's `gofmt -l` leg, which the pre-commit run executes by
  law) — repaired per the standing law: both files restored from
  the committed bytes, the changes re-applied with PYTHON on the
  real bytes with ANCHOR ASSERTIONS (the first patch's own wrong
  tab-depth anchor failed loud, as designed; the second attempt
  initially broke the brace balance — caught by `go vet`, re-anchored
  surgically on the unique next-function comment, verified clean) —
  `gofmt -l` empty, vet OK, the mkzip package green after. **The
  law's next step (recorded for the card): NEVER touch a Go file
  with the Edit tool — python-on-bytes only, anchor-asserted.**
- **THE BATTERY** — re-verified green AFTER the repaired hardening
  landed: main plain **2398 top-level PASS over 36 ok packages, 0
  FAILs, 0 skips** (exit 0); main race **36 ok, 0 data races**
  (exit 0); webui **95 top-level PASS over 4 ok packages** (exit 0);
  `gofmt -l` empty + `go vet` + `go build` clean in both modules;
  the tree clean through every leg — **2493, the exact committed
  number** (the count unchanged: the new assertions extended
  existing test functions).
- **THE RELEASE v0.92.0-go.3** (the proven ritual, verbatim): the
  bump commit `5e0c09d` (VERSION + wails.json productVersion →
  `0.92.0-go.3`, the one-version-truth) carrying the tag → pushed
  (main + tag, the ephemeral inline PAT, nothing persisted) → **go-ci
  run 40022070901 completed/success** + **release run 40022070760
  completed/success** (both watched to completion) → **the release
  422298475 LIVE** (published 2026-10-23T20:51:23Z, not a draft, not
  a prerelease) with `GitCurator-v0.92.0-go.3-windows.zip` [27,237,300
  bytes, uploaded] → **verified THREE ways**: the fresh download's
  sha256 `65eb64b4…` == the notes' number; the contents (14 members
  with `0-START-HERE.txt` FIRST [3,338 bytes], both exes' sha256 ==
  the notes' table, the VERSION member `0.92.0-go.3\n` od-verified,
  the CLI buildinfo `mod … v0.92.0-go.3`); and the cross-machine
  byte-identity — the full sandbox rebuild at the tagged checkout
  reproduced every number (exit 0, "PASS — byte-identical rebuilds";
  14 files / 56,935,578 in / 27,237,300 zipped). The release report:
  `docs/poc-artifacts/release-v0.92.0-go.3.md` (the go.3 numbers +
  the re-run slots + the AV section carrying the rung-4
  code-signing call surfaced to the owner).
- **The close records (the remainder — this report, the card v75,
  doc 12's tail entry, the README's front page + status entry, the
  worklog append)** — the front page now points at `v0.92.0-go.3`
  with the portable framing; the status entry + doc 12's tail entry
  follow the house pattern.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**THE PHASE-7 EXIT EVALUATION, honestly against doc 12's gate
("acceptance criteria demonstrated on a clean Windows VM; release
attached; owner sign-off"): NOT GREEN.** The evidence, item by item:

- T-80 ✓ (the deterministic zip + the ritual — unchanged, and
  re-proven on this release: the double-build + cold-cache legs all
  byte-identical, exit 0).
- T-81 ✓ (the release workflow — re-proven live: run 40022070760
  green end-to-end, the second REAL release through it).
- T-83 ✓ (the docs — the QUICKSTART reframed this session; the
  README front page current).
- AC-17 ✓ (the benchmark record — unchanged).
- The release legs: **PASS mechanically** (attached, sha-verified,
  three-ways reproducible — both go.2 and go.3) — **but the
  demonstration on Windows FAILED on the owner's machine** (T-82
  items 1/3 FAIL live; items 4–8 not run/unknown) **and the owner
  sign-off is NOT given — the owner's message is the opposite of a
  sign-off.** The gate's clean-VM leg and the sign-off leg are both
  unmet.
- **VERDICT: PHASE 7 DOES NOT EXIT GREEN. The fix cycle reopened
  under the owner's explicit directive** ("make the next version
  both CLI and GUI WORK — portable, without install"): go.3 is the
  first iteration (the portable-first reframe); what would close the
  gate remains the owner's live legs passing (the go.3 re-run slots)
  + their sign-off — with the root cause on their machine still
  undetermined pending the exact error text. **The honest boundary,
  restated: this sandbox has no Windows box (standing since
  SESSION-001); the binaries are byte-proven (the three-way
  verification above) but have never been EXECUTED on any real
  Windows machine by the project; the owner is the first live
  Windows user; without their error text, the failure's root cause
  is a candidate list, not a settled fact.**

**The session's own acceptance (the intake + the diagnosis + the
fix release, per the directive): PASS** — the evidence above: the
intake recorded verbatim-honest (commit `97ad2fa`); the diagnosis
byte-settled to the honest ceiling; the hardening shipped + tested
(`8deaf29`, battery 2493 green); the release cut + verified three
ways (`5e0c09d` → tag → runs green → release 422298475 live).

## Spec updates made (file → what changed, why)

- `0-START-HERE.txt` — NEW (the owner's first file in every release
  zip from go.3 on; the SmartScreen/AV/extract-first/Win10-64-bit
  card).
- `build/mkzip/main.go` + `build/mkzip/main_test.go` — the
  composition gains the START-HERE member (rootAssets + the
  must-list + the Notepad ASCII+CRLF determinism block).
- `WINDOWS-QUICKSTART.md` — the portable-first reframe (no install
  step; the SmartScreen section; 1-INSTALL as the optional checker;
  version header go.3).
- `.github/workflows/release.yml` — the notes template's First-run
  section (additive).
- `docs/poc-artifacts/t-82-clean-vm-checklist.md` — §1 intake + the
  go.3 re-run slots + the diagnosis record + the §4 dated note.
- `docs/poc-artifacts/release-v0.92.0-go.2.md` — the results table +
  the AV section filled (the owner's go.2 results, honest-coarse).
- `docs/poc-artifacts/release-v0.92.0-go.3.md` — NEW (the AC-23
  scaffold, the go.3 numbers, the re-run slots, the AV section with
  the rung-4 call).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-82
  row flipped PARTIAL→OPEN (the live run failed; the go.3 re-run +
  the error details are the path).
- `VERSION` + `webui/wails.json` — the bump to `0.92.0-go.3`
  (commit `5e0c09d`, the tag's ride).
- `README.md` — the front page points at go.3 (the portable
  framing) + the honest What's-left updated + the SESSION-074
  status entry.
- Reality-vs-spec: the PARTIAL-WIPE MODE-FLIP class is NEW (a
  filesystem restore arriving with every tracked file +x) — recorded
  in the card's recipe (the index-mode restore command) + the
  worklog; the MULTIEDIT-TAB law gained its SIXTH strike + the
  "python-only for Go files" hardening (the card carries it).

## Blockers & decisions needed

- **THE BLOCKER (the honest one): the root cause on the owner's
  machine is UNDETERMINED.** What is needed from the owner (30
  seconds, item 1 of §Owner actions): the exact error text /
  screenshot of whatever they see when an exe "doesn't work" + their
  Windows version (`winver`) + whether any antivirus beyond
  Defender runs. Without it, the candidate list stands
  (SmartScreen/AV/extract-first/pre-Win10-64/Windows-only-bug) and
  further fixes are guided shots.
- **The code-signing decision (the owner's)**: an EV/OV cert is the
  only rung that removes the SmartScreen class at the source (doc 10
  §9's rung 4 — surfaced with cost/benefit in the go.3 report's AV
  section).
- **The fully-portable-shape decision (the owner's, if go.3's
  reframe still isn't "portable enough")**: bundling the WebView2
  Fixed-Version runtime (~100 MB more) + a portable git (MinGit,
  ~40–50 MB more) inside the zip would make the GUI truly
  xcopy-portable on ANY 64-bit Win10+ machine with zero downloads —
  at ~5× the zip size. A surfaced option, not a default.
- Carried, unchanged: the Phase-8 register (the menu stands from
  SESSION-064); the standing small calls (the dissolution ADR-0002
  Am.5; vanilla-vs-vite; the batch-scorecard modal; the v0.91.0
  zip-parity reading; the answering-window belt + T-71
  merged-verdict).

## Owner actions requested

1. **THE 30-SECOND ASK (the blocker's key)**: when you try the new
   `v0.92.0-go.3` and if ANYTHING still fails — screenshot or copy
   the EXACT error text you see (the blue SmartScreen box, an
   antivirus message, or the console window behind the app), plus
   your Windows version (`winver`) and whether an antivirus beyond
   Defender runs. That one picture pinpoints the cause immediately —
   the shipped bytes are proven correct three ways, so the failure
   is a machine-gate question only your screen can answer.
2. **The re-run against go.3** (the checklist's re-run slots):
   download the new zip, **Extract All**, open
   **`0-START-HERE.txt`** (it sorts first — 30 seconds), then
   double-click `GitCurator.bat`. Paste the results (any form) into
   the next session's prompt — it does the intake + closes Phase 7
   if green.
3. **The code-signing call** (optional, the AV ladder's rung 4): an
   EV/OV certificate would remove the SmartScreen blue box for
   every future user (annual fee + validation; the only real fix
   for the unsigned-app warning class) — your call, no rush.
4. **The fully-portable-shape call** (optional): I can bundle the
   WebView2 runtime + a portable git inside the zip (~150 MB total)
   so NOTHING is ever downloaded and no component is ever missing —
   your call if go.3's shape still isn't enough.
5. The Phase-8 menu stands (the full menu from SESSION-064; the
   register's eight options) — unchanged, whenever wanted.

## Handoff

- **Next session: SESSION-075 = the post-fix session — the decision
  tree, updated:** (a) if the prompt carries the owner's go.3
  re-run results (or their error text): the intake (the checklist's
  re-run slots + the go.3 report's table) +, if green, the Phase-7
  exit evaluation re-run (the gate: the live legs passing + the
  owner sign-off); if the error text names a machine gate
  (SmartScreen/AV/pre-Win10), record it + walk the owner through
  the gate or take the corresponding decision (the cert, the
  portable shape); if it names a Windows-only bug, it becomes the
  next fix cycle's target with the error as the anchor; (b)
  otherwise the standing delegation (re-verify + re-surface + close
  cleanly). The owner's directive stands until the app WORKS on
  their machine.
- NEXT-SESSION.md rewritten: yes (the state card v75, the
  SESSION-075 prompt inside).
- Push state: CLEAN — `8deaf29` (hardening) + `97ad2fa` (intake) +
  `5e0c09d` (the bump, the tag's ride) pushed (go-ci run 40022070901
  + release run 40022070760 both green, watched); the close commit
  (docs-only) is the final push, its go-ci run watched to completion
  per the standing hygiene.

## The next prompt (verbatim, for the owner to copy)

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
