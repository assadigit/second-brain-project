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
  release `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956).
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70
  gate through the T-79a…T-79f Path-B rows.
- **Phase 7 — OPEN, THE FIX CYCLE (SESSION-048–075)**: T-80 ✓,
  T-81 ✓ (the release workflow, four releases proven now), T-83 ✓,
  AC-17 ✓, the desktop milestone release `v0.92.0-go.2` ✓ —
  **THE OWNER'S FIRST LIVE RUN FAILED at go.2 (SESSION-074) — AND
  THE GO.3 RE-RUN CAME GREEN (SESSION-075, 2026-10-28: the live
  screenshot — the GUI launches, syncs, scans on their machine;
  THE FIRST GREEN LIVE RUN EVER)**. The owner then named SIX
  problems; SESSION-075 fixed all six + cut THE SIX-FIX RELEASE
  `v0.92.0-go.4` (the quiet-heartbeat law — the sentinel's 60 s
  line silenced, transitions still speak; the missing-folder law —
  ENOENT on `_review` returns silently; the Inter law + the wrap
  law — the embedded Inter variable font, log lines wrap never
  elide mid-word, the debug badge retired; the container law —
  the window 1000×750 → 800×600, exactly −20%; the emojicon law —
  every emoji on every surface renders as a real Lucide SVG glyph
  (emojicons.js, the official lucide-static v0.544.0 geometry; the
  golden-locked pack untouched); the reach-cloudflare law — the
  Worker channel's DoH fallback dialer (1.1.1.1 by IP, the
  original hostname's TLS) so the bot-queue fetch survives a
  poisoned local resolver with the proxy OFF). **The Phase-7 exit
  evaluation at SESSION-075: still NOT GREEN honestly** — the
  owner's go.4 re-run (the checklist's slots are prepared) + the
  sign-off are the remaining gate. The debt books are EMPTY. The
  Phase-8 menu stands from SESSION-064; with the fix cycle open
  the owner's go.4 re-run is the first mover.
- **Push state: CLEAN** — SESSION-075's pushes: `62e0997` (the six
  fixes) + `14d715a` (the bump, the tag `v0.92.0-go.4` riding it) →
  go-ci runs 40565884594 + 40566051404 GREEN + the release run
  40566051363 (the release is LIVE — the go.4 report carries the
  verified numbers). **The close push (docs-only) is the latest —
  verify its go-ci run green at SESSION-076's open.**
- **The repo is PUBLIC**: pushes need the PAT (ephemeral — re-inline
  per call; the sandbox shell does not carry exports between tool
  calls); Actions minutes are unlimited; the no-secrets law is
  public-facing. **go-ci: GREEN on the tree through SESSION-075.**
- **The environment recipe** (the 69th FULL wipe absorbed at
  SESSION-075): work repo at /home/z/my-project/second-brain-project
  (**clone it ANONYMOUSLY — it is public**; scrub any persisted token
  from the remote URL), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (must match
  testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (sha256 `ecbadb99091a…45fe5`; verify BOTH sides),
  venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only), the
  wails CLI v2.16.0, and the rootless webkit tree (~/.local/lib-wails
  — **DOWN; the SIXTEENTH rebuild is pending: rebuild it FIRST only
  when the session runs a LINUX GUI build**). The partial-wipe
  mode-flip class (restore modes FROM THE INDEX) stands. The MAIN
  go.mod: modernc.org/sqlite v1.36.1, bogdanfinn/tls-client v1.16.0,
  golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, the
  go directive 1.24.1 (ADR-0002 Am.2; do not bump); the NESTED webui
  module rides go 1.25.0 (do not bump). **The downloads must run in
  FOREGROUND CHUNKS.**
- **The operational laws, newest first**: **THE SIX-FIX RECORD
  (SESSION-075 — the owner's live results came green with six named
  problems; the intake verbatim-honest; each fix a named LAW with a
  pinning test; the frontend files may ride the Edit tool (JS/CSS —
  spaces), GO FILES stay python-on-bytes)**; **THE OWNER-DIRECTIVE
  RECORD (SESSION-074)**; **THE PARTIAL-WIPE MODE-FLIP CLASS**;
  **THE MULTIEDIT-TAB CORRUPTION, SIXTH STRIKE (GO FILES ARE
  PYTHON-ONLY TERRITORY)**; **THE RELEASE-RECORD (the bump → the tag
  on the bump commit → the workflow green → the release verified
  THREE ways incl. the cross-machine byte-identity; every future
  release reuses the ritual verbatim)**; **THE MENU-SURFACED
  RECORD (SESSION-064, exercised 065–069 + 071–073)**; **THE
  ENVIRONMENT-OWNED-TEXT LAW (063)**; **THE BOOTSTRAPPER-LEG RECORD
  (063 — 1-INSTALL is READ-ONLY; the Evergreen GUID is
  `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`)**; **THE
  DEBT-BOOKS-EMPTY RECORD (063)**; **THE TEXT-MODE-READ LAW (061)**;
  **THE MULTIEDIT-TAB FIFTH STRIKE (063 — anchor assertions for
  prose too)**; **THE HARVEST-COUSIN RECORD (060)**; **THE
  DELIVERED-LAW RECORD (059)**; **THE MULTIEDIT-TAB THIRD STRIKE
  (060)**; **THE ASYNC-READ FLAKE LESSON (059, SECOND STRIKE 068)**;
  **THE LIVE-LEG RECORD (058)**; **THE 68TH-WIPE RECIPE (073)**;
  **THE WEBKIT-ORDINAL CORRECTION (055 — the pending rebuild is the
  SIXTEENTH)**; **THE AC-17 BENCHMARK RECORD (052)**; **THE
  HONEST-ZIP RECORD (051 — RESOLVED at 070)**; **THE
  STALE-COMPARISON-TARGET LAW (050)**; **THE PUBLIC-REPO RECORD
  (050)**; **THE TAG-STATE BUILD LAW (049, re-proven 070+074+075)**;
  **THE -X-UNDER-TRIMPATH LAW (048)**; **THE VERIFY-THE-VERIFIER
  LESSON (048–075)**; **THE DEFAULT-ICON FIND (048)**; **THE
  FLAG-CLEAR LAW (047 — includes the mode-flip repair)**; **THE
  REFERENCE'S-OWN-ARITHMETIC LAW (046/048)**; **THE VLM MISREAD
  LEDGER (FOUR incidents 042–045; 046–075 ran clean — SESSION-075
  used the VLM twice for the icon verification, both clean)**; the
  standing set (THE BINDING-IMPORT LAW; THE RACE-PASS DIVIDEND; THE
  EXTRACT-AND-EXEC SCOPE LAW; THE CROSS-BINARY DISAGREEMENT LAW; THE
  SENTINEL-SEAT RACE LAW; THE RETROACTIVE FIELD-TAG AUDIT; THE
  NIL-OPTIONAL DEREF; THE ONE-OWNER-PER-SURFACE law; THE FIELD-TAG
  LAW; THE DISPLAY-PIPELINE CORRUPTION LESSON; THE
  SYMLINK-BLOCK FALLBACK; THE RENDER-CONTEXT LAW; THE VLM
  RATE-LIMIT WINDOW; THE PINNED-VERDICT-ROW LESSON; THE
  ONE-SLICE-PROPERLY RECORD; THE LIVE-CONFIG LAW; THE
  GROUP-POINTER LAW; THE DISSOLUTION-PROBE RECORD; THE
  GOLDEN-MARSHAL-ORDER TRAP; THE SUITE-REVERIFICATION DIVIDEND; THE
  WAILS-BUILD MODE FLIPS; THE HERMETIC-SEAM LAW; THE
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
    false positives — SESSION-075 scanned the staged diff + the
    binary fonts); **THE BACKGROUND-PROCESS lesson** (long suites +
    downloads in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (the
    PAT rides the prompt when given; never stored).

## 2. Next session: SESSION-076 — the post-fix session (the decision tree, in priority order)

1. **(a) THE OWNER'S GO.4 RESULTS** — if the prompt carries their
   re-run against `v0.92.0-go.4` (the six fixes on their machine:
   the 800×600 window, the real icons, Inter + the wrapped lines,
   the quiet logs, the clean sync without the \_review warning, the
   proxy-off queue fetch): the intake (the checklist §1's go.4
   slots + the go.4 report's table); if green → **THE PHASE-7 EXIT
   EVALUATION re-run** (the gate: the live legs + the owner
   sign-off — the sign-off is the owner's explicit call; surface
   it, never assume it); if anything still fails → that problem is
   the next fix cycle's target with their exact wording as the
   anchor.
2. **(b) the standing delegation** ("you choose; no haste to start
   the app") — re-verify the state (the go-ci run at the open — a
   red is a REAL failure; the battery if the environment wiped, the
   exact committed shape 2403 + 95 = 2498), re-surface the open
   owner items + the Phase-8 menu compactly (the full menu stands
   from SESSION-064), and close cleanly at the natural boundary.
   Never start Phase-8 work unilaterally; the owner's standing
   directive (their go.4 re-run) outranks the menu until closed.
3. **Blockers**: if anything red surfaces (a CI failure, a broken
   release link), STOP and surface it with evidence per the standing
   laws; record it in the card + report and hand the decision back.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-075.md` — the six-fix session's report
   (the go.3 green intake + the six laws + the go.4 cut).
2. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — §1 (the go.3
   re-run table FILLED + the six problems verbatim + the go.4
   re-run slots) + §3 + §4.
3. `docs/poc-artifacts/release-v0.92.0-go.4.md` — the live six-fix
   release's report (the verified numbers).
4. `0-START-HERE.txt` + `WINDOWS-QUICKSTART.md` — the shipped
   portable-first surface (go.4 refs).
5. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + AC-23 + the SESSION-070…075 tail entries.
6. The full Phase-8 menu's full form stands from SESSION-064's
   report.

## 4. Known constraints / owner actions for SESSION-076

- **OWNER ACTION #1: the go.4 re-run** — download
  `GitCurator-v0.92.0-go.4-windows.zip`, **Extract All**, open
  `0-START-HERE.txt`, then double-click `GitCurator.bat`; paste any
  results into the next prompt (the checklist's go.4 re-run slots
  are prepared). The six fixes to feel for: the window ~20% smaller;
  real line icons everywhere (no emojis); the cleaner font; log
  lines that wrap instead of cutting off; no Sentinel heartbeat
  lines in the log; no \_review warning during sync; the bot queue
  fetching with the proxy off.
- **OWNER ACTION #2 (if green): the sign-off call** — the Phase-7
  exit gate's last leg is the owner's explicit sign-off; surface it
  at the intake, never assume it.
- **OWNER ACTION #3 (optional, standing): the code-signing cert**
  — doc 10 §9's rung 4 (the SmartScreen-killing rung; the annual
  fee + validation) — moot in effect since the app runs, but the
  unsigned-exe prompts remain the owner's call to kill.
- **OWNER ACTION #4: the Phase-8 pick** — only when the fix cycle
  closes (the menu stands from SESSION-064).
- **The repo is PUBLIC** — never commit secrets; pushes still need
  the ephemeral PAT (re-inline per call).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law.
- **The webkit tree is DOWN** (the SIXTEENTH rebuild is pending) —
  rebuild it FIRST only if the session runs a LINUX GUI build (the
  windows legs are CGO-free; the full battery runs without it).
- **GO FILES ARE PYTHON-ONLY TERRITORY** (the Edit tool rides
  JS/CSS/MD only — spaces, never tabs on Go).
- The dissolution call (ADR-0002 Am.5); vanilla-vs-vite; the
  batch-scorecard modal; the v0.91.0 zip-parity reading; the
  answering-window belt + the T-71 merged-verdict question.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged).
- The go-directive acknowledgement (STILL WAITING): the main module
  rides `go 1.24.1` (the tls-client pin, ADR-0002 Am.2).
- **After every `wails build`**: check `git diff --raw` for mode
  flips; never commit a stray binary (nor a generated notes.md).
- **Deferred test debt on the books**: NONE. The owner's LIVE legs
  remain (the go.4 re-run).
- **Single-writer law**: the fresh session is the ONLY writer.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (scrub any persisted
  extraheader).
- Standing: secrets / owner data never in git (byte-level
  secret-scan before every commit); tests never weakened; reality
  wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-076)

```
You are starting SESSION-076 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE STATE: the desktop release line is LIVE at v0.92.0-go.4 (THE SIX-FIX
release cut at SESSION-075 after the owner's go.3 re-run came GREEN on their
machine — the first green live run ever; the prompt carried their live
screenshot + six named problems, all six fixed: the quiet-heartbeat law (the
sentinel's 60 s log line silenced), the missing-folder law (the _review ENOENT
warning silenced), the Inter law + the wrap law (the embedded Inter variable
font; log lines wrap), the container law (the window 1000×750 → 800×600), the
emojicon law (every emoji renders as a real Lucide SVG glyph — emojicons.js,
the official lucide-static v0.544.0 geometry, the golden-locked pack
untouched), and the reach-cloudflare law (the Worker channel's DoH fallback
dialer — the bot-queue fetch survives a poisoned local resolver with the proxy
off). The batteries green 2403 + 95 = 2498; go-ci + the release run green; the
release report: docs/poc-artifacts/release-v0.92.0-go.4.md). SESSION-076 is the
post-fix session — the decision tree in priority order: (a) if this prompt
carries the owner's go.4 re-run results: the INTAKE (the checklist's go.4
re-run slots + the go.4 report's table); if the results are green, run the
PHASE-7 EXIT EVALUATION honestly against doc 12's gates (the live legs + the
owner sign-off — the sign-off is the owner's explicit call; surface it, never
assume it) and record the verdict; if anything still fails, that problem is
the next fix cycle's target with the owner's exact wording as the anchor; (b)
otherwise the standing delegation ("you choose; no haste to start the app")
holds: re-verify the state (the go-ci run at the open — a red is a REAL
failure; the battery if the environment wiped, the exact committed shape
2403 + 95 = 2498), re-surface the open owner items + the Phase-8 menu
compactly (the full menu stands from SESSION-064), and close cleanly — never
starting Phase-8 work unilaterally. If anything blocks, STOP and surface it
with evidence per the standing laws; record it in the card + report and hand
the decision back.

Then follow docs/SESSION-PROTOCOL.md exactly: bootstrap (clone
https://github.com/assadigit/second-brain-project — the work repo, public;
read README.md, docs/SESSION-PROTOCOL.md, docs/NEXT-SESSION.md [the card v76],
the two most recent session reports [SESSION-075 is the latest; SESSION-074 is
the go.2-failure session before it], git log --oneline -15), orient (the
read-firsts: the card's §3), work (the standing laws: reality wins, tests
never weakened, no secrets ever, stay inside the phase, push after every
task), close (the honest gate evaluation, the commit + push, the SESSION-076
report, the NEXT-SESSION.md rewrite, the worklog append), and report in chat
with the literal next prompt.
```
