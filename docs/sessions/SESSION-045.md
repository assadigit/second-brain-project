# SESSION-045 — 2026-10-20 — Phase 6 (Path B): T-79d COMPLETE (the four Chrome doors) + the 41st-wipe repair + one committed-battery flake root-caused and fixed BEFORE work; 2337 main-battery tests over 36 packages + 73 webui over 4 [corrected at SESSION-046 — an arithmetic slip in this close line; six shell tests landed]

## Goal (from NEXT-SESSION.md at start)

T-79d (L): the Chrome doors as More-menu doors — 🖐 Hand-deliver
walled links + 🖐 Scrape hand rows via Chrome + 🤖 Chrome tab-retry
failed links + 🪄 Attach to my Chrome, the four anytime triggers over
the ALREADY-LIVE engines (internal/chrometabs, T-44;
internal/handdelivery). THE COLLABORATOR LAW: Chrome is OPTIONAL —
the reference's optional-collaborator guards (hero.py) preserved:
when Chrome is absent or its DevTools port is closed, the doors
refuse honestly with the reference's own remedy lines, never a
crash, never a silent no-op.

## What was done (task IDs, files, commits — cite hashes)

- **The forty-first full wipe repaired BEFORE work**: the ENTIRE
  environment was gone again (both repos, the toolchain, the venv,
  the module cache, the wails CLI, AND the webkit tree). Everything
  rebuilt from the committed recipes: both repos re-cloned with the
  ephemeral GIT_CONFIG_COUNT PAT form (nothing persisted, both
  configs verified clean); the work repo at `6dd7537` =
  SESSION-044's close, up to date with origin/main; the reference
  at `98f4fc71cfb862cee7225d864636c08062724348` =
  testdiff/PINNED-GITCURATOR-REV exactly. go1.27.2 from go.dev; the
  local sha256 matches TOOLCHAIN.md Am.2's committed 64-char pin
  byte-for-byte. The venv (Python 3.12.15, the plain form — this
  boot HAS ensurepip). The wails CLI v2.16.0. The THIRTEENTH
  webkit-tree rebuild: 144 debs / 77.7 MB → dpkg -x → the 42 .pc
  prefix rewrites → the 26 dangling dev links as Am.4 hardlinks to
  the resolved system targets (symlink(2) BLOCKED this boot — the
  standalone probe refused again; zero unresolved dev links; the
  wails build's clean CGO link is the proof).
- **THE PRE-WORK FLAKE REPAIR (commit `4c0234c`, pushed before the
  task work)**: the full-battery re-verification caught
  `TestBotDoorExportOrderedMapItems` failing 2-in-7 on the fresh
  rebuild. Settled per the flake law (repeated runs + the byte-level
  root cause BEFORE touching anything): the reference's own
  `_item_url` (sync_display.py:118-127) walks the FIXED key tuple
  `('url_normalized', 'url_original', 'url')` BY KEY LOOKUP — dict
  insertion order never decides. The Go readers'
  `*storage.OrderedMap` branches read the row's FIRST INSERTED KEY
  instead: any Worker row whose first key was not url_normalized
  disagreed with the reference (the URL dropped, or url_original
  answered where Python answered url_normalized) — THE
  CROSS-BINARY DISAGREEMENT LAW class, and the Worker contract's
  rows carry forward_count/ledger_id/… alongside the URL keys, so
  any order is a legal row (the test rig's randomly-ordered
  two-key row hit it ~half the time). Fixed in both copies
  (`internal/app/queue.go` queueItemURL + `internal/cli/
  fetchqueue.go` itemURL — the OrderedMap branch now walks the same
  priority tuple with the isinstance fallthrough) + pinned
  deterministically by `TestQueueItemURLOrderedMapPriorityTuple` +
  `TestItemURLOrderedMapPriorityTuple` (the flake's own
  forward_count-first row shape; the priority beats insertion
  order; the non-string fallthrough; the junk row; the routed split
  end-to-end). Tests never weakened — the engine was wrong, the
  rig was right.
- **T-79d implemented — `internal/app/chromedoors.go`** (the
  engines, the pertests/linkdoors/botdoors pattern — Settings
  methods): the four doors over SIX new hermetic `SettingsDeps`
  seams, every one faked at construction in every test per THE
  HERMETIC-SEAM law (the production wiring lives in App.New):
  `ChromeFind` (handdelivery.FindChrome), `ChromeUserDataDir`
  (chrometabs.RealUserDataDir), `ChromeAttachBat`
  (chrometabs.WriteChromeAttachBat), `ChromeAttachRun` (the cmd /c
  detached launch), **`ChromeDeliver` — THE CDP/DevTools seam**
  (chrometabs.DeliverPagesViaChrome over the plain-map config), and
  `HandEnqueue` (handdelivery.EnqueueHandDelivery with its
  open-Chrome leg — the seam keeps the browser launch hermetic).
  **🖐 Hand-deliver walled links** (processing_control.py:1000): the
  guards' remedy boxes, the walled-rows scan (the honest
  read-failure warning + the corrupt-cache tolerance — never a
  crash), the queued-count variant of the empty box, the cancel
  line, the enqueue with the walls + the open leg, the
  Queued-for-hand-delivery box with both opened variants, and the
  picker's Open-the-folder reveal (`OpenHandDeliveryFolder` over
  the existing OpenFileLocation seam). **🖐 Scrape hand rows via
  Chrome** (615): the master-table hand-rows scan, the setdefault
  error law (the wall, else the gesture), the >10 confirm, the
  reference's own empty box. **🤖 Chrome tab-retry failed links**
  (1601): the retry-queue scan with the http(s)/loopback/dismissed
  filters + the session's offered set (never re-offered once
  picked). **The worker's half** (1351-1519):
  `ChromeDeliverLinks` (the whole-body crash guard answering the
  honest failed report — the reference's try/except) +
  `ChromeDeliveryFailedBox` (the `_on_delivery_failed` modal
  grammar, pure: the five-name cap, the `… and N more` tail, the
  `(a link)` fallback). **🪄 Attach to my Chrome** (669): the whole
  ladder — find Chrome (the remedy box), find the profile (the
  remedy box), write the .bat (the write-failure box), the platform
  split (the ONE explicit OK question on Windows / the honest POSIX
  ready box), the restart's own boxes, and the
  `web_browser_attach_port` coercion with the 9222 fallback over
  every junk shape (incl. the plain-int case a Go-side cfg.Set
  produces).
- **THE BUSY LADDERS carried in the reference's own order** (the
  cross-binary law — a pipeline-off config answers the pipeline box
  even while busy): `HandRowsScan(chromeBusy bool)` +
  `ChromeTabRetryScan(batchBusy bool)` take their flags as
  parameters (the live seats are the shell's, exactly like the
  reference's `_chrome_delivery_busy` / `_batch_running`); pinned
  by `TestChromeDoorBusyLadders` (guards → busy → scan).
- **The shell** — `webui/heroui.go`: seven More doors +
  `startChromeTabRetry` (`_start_chrome_tab_retry`'s whole law):
  ONE delivery at a time — a second request while a delivery runs
  MERGED into the backlog the finish handler drains (the busy line
  in the reference's own words); the delivery on a goroutine (the
  desk speaks its own lines through the bus); the failure story
  FIRST (the ⚠ line + the `chrome:failed` wails event carrying the
  modal — ONE listener per surface); the delivered pages re-run
  through the Websites pipeline as a fresh DIRECT batch over
  `NonGithub` (THE NOTE IS THE SUCCESS, the v0.57 routing law);
  `MoreHandRowsDeliver(confirmed)` is the two-phase confirm door
  (the held rows on the shell's own seat under mu — THE
  SENTINEL-SEAT RACE LAW; a stale Continue re-scans, never a blind
  run).
- **The frontend** — `frontend/heroflow.js`: the four rows LIVE
  with door tips + the dispatch; the two picker dialogs
  (`openHandDeliverDialog` + `openChromeRetryDialog` — the
  reference's own words, checkable rows with the wall/error under
  each URL, all ticked by default, the buttons pinned OUTSIDE the
  scrolling body); the attach confirm's default-No law (the No
  button focused); the `chrome:failed` failure modal's ONE
  listener; **THE BINDING-IMPORT LAW honored** — all seven doors
  joined the import statement, pinned by the new
  `TestHeroflowChromeDoorsLive` import pin (the 044 lesson).
  `frontend/style.css`: the chrome-pick family.
  `shell.go`'s probe comment extended; the probe's **'hand' click
  pick** (the fourth door's scan: the "Websites pipeline is off"
  guard box over the unconfigured fixture — the hermetic refusal
  class, no Chrome anywhere near).
- **THE FIELD-TAG LAW honored**: ChromeLinkRow / HandWallRow /
  HandDeliverScan / HandRowsScan / ChromeTabRetryScan /
  ChromeDeliveryReport / AttachChromeResult tagged + pinned by
  `TestChromeDoorPayloadFieldTags`.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-79d's acceptance = "the four doors live over the
chrometabs engine; the collaborator guards honest when Chrome is
absent" — **all GREEN**:

- **The doors live over the engine**: every door's box + line
  grammar pinned by the engine cases (13 new tests in
  `internal/app/chromedoors_test.go`): the guard ladders, the
  walled/hand/retry scans (incl. the read-failure warnings + the
  corrupt-cache tolerance), the queued box's two variants, the
  cancel lines, the offered-set filter + the concurrent hammer (the
  sentinel-seat law), the >10 boundaries (10/11), the delivery's
  crash guard + the seam contract (the live config rides), the
  failure box's grammar, the attach ladder + the port coercion
  table, the payload field-tags. The shell cases (6 new tests in
  `webui/heroui_test.go`): the two-phase flows, the held-seat
  clearing, the busy merge + the backlog drain, the offered
  marking, the delivered re-run batch's opts (the DIRECT NonGithub
  shape), the no-Chrome box.
- **The collaborator guards honest**: the no-Chrome remedy box IS
  the click-leg evidence (`t-79d-hand-click.png` — VLM-verified:
  the "Websites pipeline is off" guard box, the reference's own
  words, with its OK button; a pure config read, no Chrome
  anywhere near).
- **The battery at the close: 2337 main-battery tests race-clean
  over 36 packages (2324 + 13) + 73 webui tests over 4 (67 + 6)**; [corrected at SESSION-046 — the arithmetic slip; the report body's own "6 new shell tests" was the truth]
  the touched packages (internal/app, internal/cli, webui) 3×
  race-clean post-change (the ritual). The race pass caught ONE
  REAL DATA RACE in my own new shell test (`rig.batchBlock`
  read+write across the delivery goroutine) — fixed (the fake
  batch's field reads now hold rig.mu); the law pays in bugs, not
  just safety. gofmt/vet clean over both modules; the MAIN go.mod
  + go.sum + the nested module's files byte-unchanged (0 diffs,
  verified).
- **The `wails build`** over the 13th-rebuilt webkit tree: links
  clean (29.7 s); `git diff --raw`: NO mode flips (the wailsjs
  files stay at their 755 baseline; only content diffs — the
  regenerated bindings carrying the seven doors + the seven payload
  classes; 35 exported door functions).
- **The gate legs GREEN**: first frame 889 ms / 284 ms (≤ the 1.5 s
  gate); the 10k flood through the production transport with 0
  drops (208k lines/s accepted; handler max 50 ms/batch).
- **The screenshots VLM-verified** (`t-79d-more-menu.png` +
  `t-79d-hand-click.png`): the four doors LIVE in the menu (dark
  readable, clickable) + the tail honestly disabled. **THE VLM
  MISREAD PATTERN STRUCK A FOURTH TIME** on the focused crop: a
  mid-menu re-pass read five LIVE doors (Export All Links, Retry
  Failed, View 404 Quarantine, both Retry rows) as "faded/disabled"
  and hallucinated "checked" states on plain rows — settled per the
  law with the byte-level source check (the live keys + the uniform
  render path — only `!live` washes — pinned by
  TestHeroflowChromeDoorsLive + TestHeroflowBotQueueDoorsLive +
  TestHeroflowPerTestDoorsLive) + the full-image first pass's
  verdict. No code was touched on the VLM's word.
- The evidence archived at `docs/poc-artifacts/t-79d-chrome-doors.md`
  (+ the two screenshots + the results JSON), with the SIX recorded
  divergences (the dead dry-run shadow branch; the
  `_chrome_delivery_pending` stash counter; the frontend pickers;
  the end-of-run offer + the caught-up hand pass + the redo pass as
  the honest boundary — they ride the same `startChromeTabRetry`
  runner, so their landing is a follow-up wiring, not new engines;
  the reveal door; the per-platform attach legs).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-79d
  row COMPLETE with the full record.
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-045
  entry.
- `docs/port-blueprint/module-specs/spec-gui-processing-control.md`
  — the dated note (the four doors + the runner + the seams + the
  boundary's pointer).
- `README.md` — the SESSION-045 status block.
- `webui/README.md` — the heroui.go + heroflow.js map entries
  extended with T-79d's doors + the probe's 'hand' pick.
- `docs/poc-artifacts/t-79d-chrome-doors.md` + the two screenshots
  + `t-79d-gate-results.json` — the archived evidence.
- `docs/NEXT-SESSION.md` — rewritten for SESSION-046 (the state
  card v46).
- `docs/sessions/SESSION-045.md` — this report.

## Blockers & decisions needed

- **No new blockers.** T-79d closed clean.
- Carried, unchanged: the dissolution call (ADR-0002 Am.5);
  the vanilla-vs-vite call; the go-directive acknowledgement (now
  the thirty-sixth session carrying it); x/net/html keep-or-swap;
  the answering-window belt + the T-71 merged-verdict question
  (SESSION-039's record); the deferred test debt
  (test_noteredo's un-ported selections; the note_state
  surgery-path CRLF round-trip differential).
- **A small scoping note for the tail**: the end-of-run Chrome
  offer (`_maybe_offer_chrome_tab_retry`), hero.py's caught-up hand
  scrape, and `_maybe_redo_hand_notes` remain unlanded BY DESIGN
  this session (they are the batch-finish/hero-machine surfaces,
  not More-menu doors — the T-79d row's own naming). They ride the
  already-landed `startChromeTabRetry` runner, so their landing is
  a follow-up wiring; the NEXT session should decide whether they
  join T-79e's slice or ride a named Phase-6-closure row.

## Owner actions requested

- None urgent. When convenient (no rush): the standing calls —
  dissolution, vanilla-vs-vite; + the tail-scoping note above
  (where the end-of-run offer lands).

## Handoff

- **SESSION-046 = T-79e (M): the decommission picker** (🪦
  Decommission dead links — processing_control.py:807; the
  graveyard procedure's in-app picker writing the same Status
  cells; the confirmed-only quarantine law) — the LAST tail row;
  then Phase 6 closes and Phase 7 (T-80: the deterministic zip
  builder …) opens.
- NEXT-SESSION.md rewritten: yes (the state card v46, the
  SESSION-046 prompt inside).
- Push state: all work committed + pushed at the close (the
  ephemeral PAT; never stored, never written).
