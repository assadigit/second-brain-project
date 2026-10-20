# SESSION-046 — 2026-10-20 — Phase 6 (Path B): T-79e COMPLETE (the decommission picker — the LAST tail row) + the 42nd-wipe repair + the scoping note settled (T-79f named); 2348 main-battery tests over 36 packages + 77 webui over 4

## Goal (from NEXT-SESSION.md at start)

T-79e (M): the decommission picker — 🪦 Decommission dead links, the
graveyard procedure's in-app picker (processing_control.py:807): the
picker writes the SAME Status cells into _review/DECOMMISSIONED.md,
then enforces the burials immediately (dismissed + retry-queue rows
dropped + placeholders swept); ♻️ revived brings a link back. THE
LAST tail row — Phase 6 closes when it lands; also settle the
scoping note (the end-of-run Chrome offer + the caught-up hand pass
+ the redo pass).

## What was done (task IDs, files, commits — cite hashes)

- **The forty-second full wipe repaired BEFORE work**: the ENTIRE
  environment was gone again at boot (both repos, the toolchain, the
  venv, the module cache, the wails CLI, AND the webkit tree —
  /home/z/my-project held only the sandbox's base scaffold).
  Everything rebuilt from the committed recipes: both repos
  re-cloned with the ephemeral GIT_CONFIG_COUNT PAT form (nothing
  persisted — `git remote -v` verified clean, the extraheader form
  never touches disk); the work repo at `fd053ba` = SESSION-045's
  close, up to date with origin/main; the reference at
  `98f4fc71cfb862cee7225d864636c08062724348` =
  testdiff/PINNED-GITCURATOR-REV exactly. go1.27.2 from go.dev; the
  local sha256 matches TOOLCHAIN.md Am.2's committed 64-char pin
  byte-for-byte (`ecbadb…45fe5`). The venv (Python 3.12.15, the
  plain form — this boot HAS ensurepip). The wails CLI v2.16.0
  (GOBIN=/home/z/.local/go/bin). The FOURTEENTH webkit-tree rebuild:
  144 debs / 77.7 MB → dpkg -x → the 42 .pc prefix rewrites → the 26
  dangling dev links as Am.4 hardlinks to the resolved system
  targets (symlink(2) BLOCKED this boot — the standalone probe
  refused again and killed its compound call, the standing lesson
  held; zero unresolved dev links; the baseline `wails build`'s
  clean 30.1 s CGO link the proof the tree is sound).
- **THE PRE-WORK ARITHMETIC-SLIP REPAIR (the five records)**: the
  full-battery re-verification (2337 main + race-clean, 0 flakes —
  plain AND race, exactly the committed numbers on the main side)
  counted **73** webui tests where SESSION-045's close docs record
  "71 webui over 4 (67 + 4)". Settled at the byte level before
  touching anything: commit `fd053ba` landed **SIX** new shell tests
  in `webui/heroui_test.go` (TestHeroflowChromeDoorsLive +
  TestHeroUIChromeHandDeliverWalled + TestHeroUIChromeHandRowsTwoPhase
  + TestHeroUIChromeTabRetryFlow + TestHeroUIChromeDeliveryBacklogMerge
  + TestHeroUIChromeAttachDoor — `git show fd053ba -- webui/
  heroui_test.go | grep -c '^+func Test'` = 6), and SESSION-045's
  own report body says "6 new tests in webui/heroui_test.go" — the
  close-line arithmetic was the slip (67 + 6 = 73, not 67 + 4 = 71).
  The five records corrected in this session's close commits
  (README, the t-79d evidence doc, the roadmap, the SESSION-045
  report title + body; NEXT-SESSION.md was rewritten wholesale).
  Per-package counts pinned: 55 + 8 + 4 + 6 = 73.
- **T-79e implemented — `internal/app/decomdoors.go`** (the
  Settings-method engines, the T-79a/b/c/d shape), pinned against
  the reference's own flow (processing_control.py:807-998 +
  `_scan_backlog_alive` at 152-173, all at 98f4fc7):
  - `DecommissionScan` — the door's first half: the two-guard ladder
    (the shared Websites ladder via `chromeWebsitesGuard` — **NO
    busy guard**, the reference's own shape, pinned by test), the
    backlog-alive scan (the state DB's `IsDismissed` probe filters
    the fetcher's own auto-verdict retirements; a broken DB falls
    back to the pure file read — a broken probe never hides a
    waiting link), the `sorted({i['url'] for i in items})` dedup
    law, the empty box in the reference's own words (the graveyard
    stays available for hand-rows + the ♻️ hint), and the candidates
    pre-fill (`WriteDecommissionCandidates` over the live writers —
    the belt only ever speaks on the unexpected, as `recover()`).
  - `DecommissionBury` — the picker's answer: the picked URLs marked
    dead (`MarkURLsDeadInTable` — the SAME Status cell the owner
    would edit by hand in Obsidian, the table stays the ONE
    ledger), then the burials enforced immediately
    (`ConsumeDecommissionTable` over the same state DB a batch uses:
    dismiss + retry-row drop + placeholder sweep + the confirmed
    stamp — the reference's own comment: the dry-run law does not
    apply, this is the OWNER's explicit hand), the exact success
    line (the buried/swept/remaining arithmetic), and the exact box
    (the marked/enforced/waiting/ledger/♻️ story). An empty list is
    the dialog's Cancel — the reference's cancel line on the desk.
    An unopenable state DB defers honestly (the ⚠ line; "the next
    batch enforces the burials" — never a crash).
  - `OpenDecommissionTable` — the picker's "Open the table" reveal
    (the OS file manager over the existing `OpenFileLocation` seam;
    the _review dir created when missing — the sibling door's law).
  - THE FIELD-TAG LAW: `DecommissionScan` tagged (`box`/`links`/
    `table`) + pinned by `TestDecommissionDoorPayloadFieldTags`; the
    regenerated models.ts carries the payload class.
- **The shell — `webui/heroui.go`**: THREE More doors
  (`MoreDecommissionScan` + `MoreDecommissionBury` over the
  MoreDoorResult shape + `MoreOpenDecommissionTable`) — the sixth
  bound struct's fifth door family; the bindings regenerated
  (content-only, no mode flips).
- **The frontend — `frontend/heroflow.js`**: the menu row LIVE (THE
  TAIL'S LAST ROW FLIPPED — the honestly-disabled tail is now EMPTY)
  with its door tip; the dispatch entry; `openDecommissionDialog`
  (the reference's own words verbatim: the GENUINELY-dead message,
  the ledger label, the three buttons — Open the table / Cancel /
  **🪦 Decommission selected (danger)** — pinned OUTSIDE the
  scrolling body; **every row starts UNCHECKED** — the reference's
  `setCheckState(Unchecked)` law, a burial is never a default); THE
  BINDING-IMPORT LAW honored — the three doors joined the import
  statement, pinned by `TestHeroflowDecommissionDoorsLive` (the 044
  lesson); the probe's **'grave' click pick** (the fifth family's
  evidence leg). `shell.go`'s probe comment extended; the head
  comments updated to the honest tail-empty state.
- **The stale tail-pins updated honestly**:
  `TestHeroflowPerTestDoorsLive` + `TestHeroflowBotQueueDoorsLive`
  carried the decommission row's old landing note as a REQUIRED
  string — the door is live now, so the pins were updated to REQUIRE
  the note's absence + the export-zip row as the ONLY remaining
  honestly-disabled row (T-80's standing assignment). Tests never
  weakened — the pins track the honest state, exactly as the
  T-79b/c/d sessions did when their doors flipped.
- **The scoping note settled (the session's own decision, recorded
  everywhere)**: the end-of-run Chrome offer
  (`_maybe_offer_chrome_tab_retry`), hero.py's caught-up hand
  scrape, and `_maybe_redo_hand_notes` are named **T-79f — the
  Phase-6 closure row** on doc 13: they are the batch-finish/
  hero-machine surfaces riding the ALREADY-LANDED
  `startChromeTabRetry` runner (a follow-up wiring with its own
  pinned tests — the offered-set + the held-rows seats + the hero
  machine's caught-up story), NOT a fold-in to T-79e ("one slice
  done properly with proof beats two half-done"). Phase 6 closes
  when T-79f lands. The owner can override this naming.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-79e's acceptance = "the picker's writes land in the
graveyard table + the quarantine view refreshes" — **all GREEN**:

- **The picker's writes land in the graveyard table ✓ (the
  one-ledger law end to end)**: the engine cases (11 new tests in
  `internal/app/decomdoors_test.go`) pin the whole story over the
  REAL writers + the REAL consume: the scan's guards/empty box/
  picker data (dedup, the human + non-failed + dismissed filters,
  the candidates pre-fill + the 📋 line), the write-failure belt,
  the cancel line, the happy path (the marking's 🪦 line + the exact
  success line + the exact box; the placeholder file SWEPT; the
  other placeholder kept; the dismissal in the state DB; the
  dead-confirm stamp in the table; the waiting row untouched), the
  cleared variants, the enforcement-deferred law (the reference's
  own arithmetic — see the divergence below), the ♻️ revive
  enforcement (the same consume pass un-dismisses a hand-flipped
  revived row), the reveal door, the payload field-tags.
- **The quarantine side refreshes ✓** through the same enforcement
  the batch runs (the state DB's dismissals ARE the quarantine's
  ledger — the v0.63.3 "the owner's verdict settles" law; the same
  engine already live at T-79c's manual-resolve doors).
- **The battery at the close: 2348 main-battery tests race-clean
  over 36 packages (2337 + 11) + 77 webui tests over 4 (73 + 4)**;
  the touched packages (internal/app, webui) 3× race-clean
  post-change (the ritual). gofmt/vet clean over both modules; the
  MAIN go.mod + go.sum + the nested module's files byte-unchanged
  (0 diffs, verified).
- **The `wails build`** over the 14th-rebuilt webkit tree: links
  clean; `git diff --raw`: NO mode flips (the wailsjs files stay at
  their 755 baseline; content-only diffs — the regenerated bindings
  carrying the three doors + the `DecommissionScan` payload class).
- **The gate legs GREEN**: first frame 869 ms (≤ the 1.5 s gate);
  the 10k flood through the production transport: 526,315 lines/s
  accepted, 0 drops (handler max 30 ms/batch) —
  `t-79e-gate-results.json`.
- **The screenshots VLM-verified — BOTH passes agreeing, no misread
  incident this session** (`t-79e-more-menu.png`: the row LIVE —
  "a normal, fully active (clickable) menu item … no visual
  indicators of being disabled"; `t-79e-grave-click.png`: the click
  leg end-to-end — the "Websites pipeline is off" box with the
  reference's own remedy body + its OK button over the gate's
  unconfigured fixture, a pure config read, no network anywhere
  near).
- The evidence archived at `docs/poc-artifacts/
  t-79e-decommission-doors.md` (+ the two screenshots + the results
  JSON), with the SIX recorded divergences — the sixth is the
  honest record of my own first test expectation being WRONG (the
  deferral case's "backlog is clear" arithmetic: after the marking
  succeeds, the backlog-alive scan filters the dead row → remaining
  = 0 → the box says clear even while the placeholder file waits —
  the REFERENCE's own computation; the engine was right, the test
  was fixed to pin it).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-79e
  row COMPLETE with the full record + the NEW T-79f row (the
  Phase-6 closure row — the scoping decision).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-046
  entry (+ the 045 arithmetic-slip correction).
- `docs/port-blueprint/module-specs/spec-gui-processing-control.md`
  — the dated note (T-79e's landing + the T-79f naming).
- `README.md` — the SESSION-046 status block (+ the 045 count
  correction).
- `webui/README.md` — the heroui.go + heroflow.js map entries
  extended with T-79e's doors + the probe's 'grave' pick + the
  tail-empty state.
- `docs/poc-artifacts/t-79e-decommission-doors.md` + the two
  screenshots + `t-79e-gate-results.json` — the archived evidence.
- `README.md` / `docs/poc-artifacts/t-79d-chrome-doors.md` /
  `docs/port-blueprint/12-migration-roadmap.md` /
  `docs/sessions/SESSION-045.md` — the 71→73 webui-count
  corrections (the pre-work re-verification's catch).
- `docs/NEXT-SESSION.md` — rewritten for SESSION-047 (the state
  card v47).
- `docs/sessions/SESSION-046.md` — this report.

## Blockers & decisions needed

- **No new blockers.** T-79e closed clean.
- **The scoping decision made this session (the owner can
  override)**: T-79f = the runner's automatic triggers (the
  end-of-run Chrome offer + the caught-up hand pass + the redo
  pass) — the Phase-6 closure row; Phase 6 closes when it lands.
- Carried, unchanged: the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (now the
  thirty-seventh session carrying it); x/net/html keep-or-swap; the
  answering-window belt + the T-71 merged-verdict question
  (SESSION-039's record); the deferred test debt
  (test_noteredo's un-ported selections; the note_state
  surgery-path CRLF round-trip differential).

## Owner actions requested

- None urgent. When convenient (no rush): the standing calls —
  dissolution, vanilla-vs-vite; + confirm (or override) the T-79f
  naming for the three runner triggers.

## Handoff

- **SESSION-047 = T-79f: the Phase-6 closure row** (the runner's
  automatic triggers — the end-of-run Chrome offer + hero.py's
  caught-up hand scrape + the redo pass, over the landed
  `startChromeTabRetry` runner) — then Phase 6 CLOSES and Phase 7
  (T-80: the deterministic zip builder …) opens.
- NEXT-SESSION.md rewritten: yes (the state card v47, the
  SESSION-047 prompt inside).
- Push state: all work committed + pushed at the close (the
  ephemeral PAT; never stored, never written).
