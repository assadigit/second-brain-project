# T-79e — the decommission picker (🪦 Decommission dead links) — evidence

Recorded 2026-10-20, SESSION-046. Environment: Debian 13 trixie sandbox
(the FORTY-SECOND full wipe absorbed before work — both repos, the
toolchain, the venv, the module cache, the wails CLI, AND the webkit
tree all rebuilt from the committed recipes; the FOURTEENTH webkit
rebuild, the 26 dev links as Am.4 hardlinks [symlink(2) still
blocked]; go1.27.2 sha-verified against TOOLCHAIN.md Am.2's committed
pin), go1.27.2, wails CLI v2.16.0, webkit2gtk-4.1 2.54.0 via the
rootless extracted tree, Xvfb 1100×750.

The task row's acceptance: **"the picker's writes land in the
graveyard table + the quarantine view refreshes"** — the picker's
writes land in the graveyard table ✓ (the one-ledger law: the picker
writes the SAME Status cells a hand-edit writes, then enforces the
burials immediately); the quarantine-side refresh rides the SAME
consume pass the batch runs (the state DB's dismissals ARE the
quarantine's ledger — the v0.63.3 "the owner's verdict settles" law,
already landed at T-79c's manual-resolve doors and enforced here
through the same engine).

## What landed

- **`internal/app/decomdoors.go`** — the engines (Settings methods,
  the T-79a/b/c/d shape), pinned against the reference's own flow
  (processing_control.py:807-998 + `_scan_backlog_alive` at
  152-173, all at 98f4fc7):
  - `DecommissionScan` — the door's first half: the two-guard ladder
    (the shared Websites ladder — NO busy guard, the reference's own
    shape, pinned by test), the backlog-alive scan (the state DB's
    is_dismissed probe filters the fetcher's own auto-verdict
    retirements; a broken DB falls back to the pure file read — a
    broken probe never hides a waiting link), the `sorted({…})`
    dedup law, the empty box in the reference's own words (the
    graveyard stays available for hand-rows + the ♻️ hint), and the
    candidates pre-fill (`WriteDecommissionCandidates` — the belt
    only ever speaks on the unexpected; the writer's own ⚠/📋 lines
    ride the passed log).
  - `DecommissionBury` — the picker's answer: the picked URLs marked
    dead (`MarkURLsDeadInTable` — the SAME Status cell the owner
    would edit by hand in Obsidian, the table stays the ONE
    ledger), then the burials enforced immediately
    (`ConsumeDecommissionTable` over the same state DB a batch uses
    — dismiss + retry-row drop + placeholder sweep + the confirmed
    stamp), the exact success line (the buried/swept/remaining
    arithmetic), and the exact box (the marked/enforced/waiting/
    ledger/♻️ story). An empty list is the dialog's Cancel — the
    reference's cancel line on the desk. An unopenable state DB
    defers honestly (the ⚠ line; "the next batch enforces the
    burials").
  - `OpenDecommissionTable` — the picker's "Open the table" reveal
    (the OS file manager over the existing `OpenFileLocation` seam;
    the _review dir created when missing — the sibling door's law).
  - THE FIELD-TAG LAW: `DecommissionScan` tagged (`box`/`links`/
    `table`) + pinned by `TestDecommissionDoorPayloadFieldTags`.
- **`webui/heroui.go`** — the THREE More doors (the sixth struct's
  fifth family): `MoreDecommissionScan` (the picker's data door) +
  `MoreDecommissionBury` (the answer door — the MoreDoorResult
  shape) + `MoreOpenDecommissionTable` (the reveal door).
- **`frontend/heroflow.js`** — the menu row LIVE (the tail's LAST
  row flipped) + the dispatch + `openDecommissionDialog` (the
  reference's own words verbatim: the GENUINELY-dead message, the
  ledger label, the three buttons — Open the table / Cancel /
  **🪦 Decommission selected (danger)** — pinned OUTSIDE the
  scrolling body; every row starts UNCHECKED — the reference's
  `setCheckState(Unchecked)` law, a burial is never a default) +
  THE BINDING-IMPORT LAW honored (the three doors joined the import
  statement, pinned by `TestHeroflowDecommissionDoorsLive`) + the
  probe's **'grave' click pick**.
- **`shell.go`** — the probe comment extended (the 'grave' leg); the
  probe's SB_POC_OPEN_MORE door values now cover all five families.

## The hermetic battery

- **11 engine tests** (`internal/app/decomdoors_test.go`): the guard
  ladder (+ the no-busy-guard order), the empty box, the picker's
  data laws (the dedup, the human/non-failed/dismissed filters, the
  candidates pre-fill), the candidates write-failure belt, the
  cancel line, the happy path (the marking + the enforcement + the
  sweep + the state DB + the confirmed stamp + the EXACT lines +
  box), the cleared variants, the enforcement-deferred law (the
  reference's own arithmetic — the marked row already filters the
  backlog scan, so the box says clear), the ♻️ revive enforcement
  (the same consume pass un-dismisses), the reveal door, the payload
  field-tags.
- **4 shell tests** (`webui/heroui_test.go`): the scan door (the
  guard box + the picker's data + the pre-fill through the door),
  the bury door (the exact box + the sweep + the state DB + the
  cancel line), the reveal door, the heroflow live-pin (the row
  LIVE + the dispatch + the picker's own words + the UNCHECKED law +
  THE BINDING-IMPORT pin + the probe pick).
- The two STALE tail-pins updated honestly
  (`TestHeroflowPerTestDoorsLive` + `TestHeroflowBotQueueDoorsLive`):
  the decommission row's landing note is GONE (the door is live) —
  the only remaining honestly-disabled row is the 📤 export zip
  (T-80's standing assignment), pinned as such.

## The gate legs (the T-70 law)

- first frame **869 ms** (≤ the 1.5 s gate); the 10k flood through
  the production transport: **526,315 lines/s accepted, 0 drops**
  (handler max 30 ms/batch) — `t-79e-gate-results.json`.
- The `wails build` over the 14th-rebuilt webkit tree: links clean;
  `git diff --raw`: NO mode flips (the wailsjs files stay at their
  755 baseline; content-only diffs — the regenerated bindings carry
  the three doors + the `DecommissionScan` payload class).

## The screenshots (VLM-verified, the full-image passes)

- `t-79e-more-menu.png`: the menu with **🪦 Decommission dead links
  LIVE** (clickable, solid, no disabled wash) — the VLM read the row
  as "a normal, fully active (clickable) menu item … no visual
  indicators of being disabled".
- `t-79e-grave-click.png`: the click leg end-to-end (SB_POC_OPEN_
  MORE=grave — the click → the binding → the engine → the guard
  box): the **"Websites pipeline is off"** box with the reference's
  own remedy body ("Turn the Websites pipeline on first (Settings →
  📁 Vault).") + its OK button, over the gate's unconfigured
  fixture — a pure config read, no network anywhere near.

## The recorded divergences (the honest boundary)

1. **The dead dry-run shadow-cache branch** (both in
   `_scan_backlog_alive` and the enforcement): the reference's
   `WebsiteStateDB(db_path=_dryrun.shadow_cache_path(...))` is dead
   in its own GUI (the switch is CLI-only; the GUI never flips it) —
   the port's door reads the live cache, the same behavior in every
   reachable GUI state (the T-79c/d class — recorded there too).
2. **The belt-vs-writer line split**: the reference's outer
   try/except belts (⚠️ Graveyard candidates write skipped /
   marking skipped) fire only on UNEXPECTED exceptions; the Go
   doors carry the same belts as `recover()` — the writers' own
   expected-failure warnings (⚠️ Could not create/read/write the
   graveyard table) speak through the passed log exactly like the
   reference's (pinned by the write-failure test).
3. **The picker dialog is the frontend's** (the two-door shape of
   the hand-deliver flow): the picked URLs ride back as the
   dialog's own data; the UNCHECKED default + the danger bury
   button + the buttons-outside-the-body laws are the port's
   established dialog kit over the reference's own words.
4. **The "Open the table" reveal rides the OpenFileLocation seam**
   (the reference's QDesktopServices.openUrl twin; best-effort,
   never raises — the reference's ⚠️ Could-not-open line is
   unreachable through the seam's never-raise contract). The
   _review dir is created when missing — the OpenHandDeliveryFolder
   sibling's own law.
5. **NO busy guard** (the reference's door carries none — unlike
   the hand/Chrome doors): the enforcement rides the same SQLite
   state DB a batch uses and the reference accepts the overlap by
   design; pinned by test (the guard-ladder order: guards → scan).
6. **The deferral case's "backlog is clear" arithmetic**: the
   reference's own computation — after `mark_urls_dead_in_table`
   succeeds, the placeholder's URL is dead in the table, so
   `_scan_backlog_alive` filters it → `remaining = 0` → the box
   says "The _review backlog is clear — no more notices." even
   while the placeholder file waits for the next batch's sweep (the
   table is the truth). My first test expectation was WRONG (I
   expected "still wait"); the engine matched the reference —
   reality wins, the test was fixed to pin the reference's own
   arithmetic.

## The scoping note settled (the session's own decision)

The end-of-run Chrome offer (`_maybe_offer_chrome_tab_retry`),
hero.py's caught-up hand scrape, and `_maybe_redo_hand_notes` are
named as **T-79f — the Phase-6 closure row** on doc 13 (they are
the batch-finish/hero-machine surfaces riding the landed
`startChromeTabRetry` runner — a follow-up wiring with its own
pinned tests, NOT a fold-in to this slice; "one slice done properly
with proof beats two half-done"). Phase 6 closes when T-79f lands.
The owner can override this naming in the next session.

## The reproduce recipe (committed shape)

```sh
cd webui && \
PKG_CONFIG_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu/pkgconfig \
LD_LIBRARY_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu \
wails build -tags webkit2_41
# the menu shot:
SB_POC_OPEN_MORE=menu SB_POC_RESULTS=/tmp/t79e-menu.json \
SB_SCREENSHOT=/tmp/t79e-more-menu.png sh build-support/run-gate.sh 5
# the click leg (the decommission guard box):
SB_POC_OPEN_MORE=grave SB_POC_RESULTS=/tmp/t79e-click.json \
SB_SCREENSHOT=/tmp/t79e-grave-click.png sh build-support/run-gate.sh 7
```
