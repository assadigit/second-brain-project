# T-79d — the Chrome doors as More-menu doors — evidence

Recorded 2026-10-20, SESSION-045. Environment: the 41st full wipe
repaired BEFORE work (both repos re-cloned with the ephemeral
GIT_CONFIG_COUNT PAT form — nothing persisted; go1.27.2's sha256
verified against TOOLCHAIN.md Am.2's committed pin; the venv rebuilt;
the wails CLI v2.16.0; the THIRTEENTH webkit-tree rebuild: 144 debs →
dpkg -x → the 42 .pc prefix rewrites → the 26 dev links as Am.4
hardlinks — symlink(2) still blocked at this boot). The pre-work
re-verification from scratch caught ONE REAL FLAKE in the committed
battery — `TestBotDoorExportOrderedMapItems` — root-caused at the
byte level against the reference's own `_item_url` law
(sync_display.py:118-127: the FIXED key tuple
`('url_normalized','url_original','url')` speaks by priority — dict
insertion order never decides) and repaired in both Go copies
(`internal/app/queue.go` + `internal/cli/fetchqueue.go` read the
row's FIRST inserted key instead — the cross-binary disagreement
class; the Worker row carries forward_count/ledger_id/… alongside
the URL keys, so any order is a legal row). Commit `4c0234c`, pushed
before the task work; pinned by the new
`TestQueueItemURLOrderedMapPriorityTuple` +
`TestItemURLOrderedMapPriorityTuple` (the flake's own row shape).

## What landed

The four anytime triggers over the ALREADY-LIVE engines
(internal/chrometabs, T-44; internal/handdelivery), the reference's
homes processing_control.py:615/669/1000/1601 + the shared
`_start_chrome_tab_retry` (1351-1519) + `_on_delivery_failed`
(1521-1566):

- **`internal/app/chromedoors.go`** (the engines, Settings methods
  over hermetic seams): `HandDeliverScan`/`HandDeliverQueue`
  (1000-1093 — the walled-rows scan with the queued-count variant of
  the empty box, the cancel line, the enqueue with the walls + the
  open-Chrome leg over the `HandEnqueue` seam, the Queued-for-
  hand-delivery box with both opened variants) + `OpenHandDelivery-
  Folder` (the picker's reveal) + `HandRowsScan` (615-665 — the
  master-table hand-rows scan, the setdefault error law, the >10
  confirm) + `ChromeTabRetryScan`/`MarkChromeTabOffered` (1601-1671
  — the retry-queue scan with the http(s)/loopback/dismissed filters
  + the session's offered set) + `ChromeDeliverLinks` (the worker's
  run() half over THE `ChromeDeliver` CDP/DevTools seam, with the
  reference's whole-body crash guard answering the honest failed
  report) + `ChromeDeliveryFailedBox` (the `_on_delivery_failed`
  modal grammar, pure) + `AttachMyChrome` (669-780 — the whole
  ladder: find Chrome → find the profile → write the .bat → the
  platform split → the ONE explicit OK on Windows / the honest
  POSIX box; the `web_browser_attach_port` coercion with the 9222
  fallback over every junk shape).
- **THE COLLABORATOR LAW preserved**: Chrome is optional — no Chrome
  (the `No Chrome found` remedy box), no profile (the `No Chrome
  profile found` box), the DevTools port closed (the engine's own
  per-page failures through the delivery report + the failure
  modal), a crashed seam (the ⚠ line + the honest failed report) —
  never a crash, never a silent no-op. Every body rides the
  reference's whole-try/except shape.
- **The shell** (`webui/heroui.go`): seven More doors —
  `MoreHandDeliverWalled` + `MoreHandDeliverQueue` +
  `MoreOpenHandFolder` (the picker family), `MoreHandRowsDeliver(con-
  firmed)` (the two-phase confirm door; the held rows on the shell's
  own seat under mu — THE SENTINEL-SEAT RACE LAW), `MoreChromeTabRe-
  try` + `MoreChromeTabRetryRun` (the offer + its answer), `MoreAtta-
  chMyChrome(confirmed)` + `startChromeTabRetry` — `_start_chrome_
  tab_retry`'s whole law: ONE delivery at a time (the second request
  MERGED into the backlog the finish handler drains), the delivery
  on a goroutine (the desk speaks its own lines through the bus),
  the failure story FIRST (the `chrome:failed` wails event carries
  the modal — ONE listener per surface), then the delivered pages
  re-processed through the Websites pipeline as a fresh DIRECT
  batch over `NonGithub` (THE NOTE IS THE SUCCESS, the v0.57 routing
  law).
- **The frontend** (`frontend/heroflow.js`): the four rows LIVE with
  door tips, the two picker dialogs (`openHandDeliverDialog` +
  `openChromeRetryDialog` — the reference's own words, checkable
  rows with the wall/error under each URL, all ticked by default,
  the buttons pinned OUTSIDE the scrolling body), the attach confirm
  with the reference's default-No law (the No button focused), the
  `chrome:failed` failure modal listener, AND THE BINDING-IMPORT
  LAW — all seven doors joined the import statement (pinned by the
  new `TestHeroflowChromeDoorsLive` import pin, the 044 lesson).
  `frontend/style.css`: the chrome-pick family (the checkable rows
  + the header glyph row). `shell.go`'s probe comment extended.
- **THE FIELD-TAG LAW**: ChromeLinkRow / HandWallRow / HandDeliverS-
  can / HandRowsScan / ChromeTabRetryScan / ChromeDeliveryReport /
  AttachChromeResult all tagged + pinned by
  `TestChromeDoorPayloadFieldTags`.
- **The seams** (`SettingsDeps`, production-wired in App.New):
  ChromeFind (handdelivery.FindChrome), ChromeUserDataDir
  (chrometabs.RealUserDataDir), ChromeAttachBat
  (chrometabs.WriteChromeAttachBat), ChromeAttachRun (the cmd /c
  detached launch), ChromeDeliver (chrometabs.DeliverPagesViaChrome
  over the plain-map config — THE CDP/DevTools seam), HandEnqueue
  (handdelivery.EnqueueHandDelivery with the open-Chrome leg). Every
  one faked at construction in every test per THE HERMETIC-SEAM law.
- **The busy ladders carried exactly** (the cross-binary law): the
  guards → the busy flag → the scan, in the reference's own order —
  `HandRowsScan(chromeBusy)` + `ChromeTabRetryScan(batchBusy)` take
  their flags as parameters (the live seats are the shell's, exactly
  like the reference's `_chrome_delivery_busy` / `_batch_running`);
  pinned by `TestChromeDoorBusyLadders`.

## The recorded divergences

1. **The dead dry-run shadow branch not carried**: the reference's
   doors read `_dryrun.shadow_cache_path(...)` when the global
   dry-run switch is on — dead in its own GUI (the switch is
   CLI-only; the GUI never flips it). The port's doors read the live
   cache — identical behavior in every reachable GUI state.
2. **The `_chrome_delivery_pending` counter not carried**: it feeds
   v0.54's batch-scorecard stash (the held scorecard the batch
   finish flushes). The port's manual doors re-run the delivered
   pages as their own batch whose own finish tells the whole story;
   the stash interplay belongs to the end-of-run offer surface
   (below).
3. **The picker dialogs are the frontend's** (the two-door shape of
   the recategorize flow): the picked rows ride back as the dialog's
   own data (the reference's picked-from-candidates law).
4. **The end-of-run offer + the caught-up hand pass + the redo pass
   remain unlanded** (the honest boundary): `_maybe_offer_chrome_
   tab_retry` (the end-of-batch modal), hero.py's caught-up hand
   scrape (280-305), and `_maybe_redo_hand_notes` are the batch-
   finish/hero-machine surfaces — NOT the More-menu doors this row
   names. They ride the same `startChromeTabRetry` runner the shell
   now owns, so their landing is a follow-up wiring, not new
   engines.
5. **The "Open the folder" reveal rides a new door**
   (`MoreOpenHandFolder` → Settings.OpenHandDeliveryFolder over the
   existing OpenFileLocation seam — the reference's
   QDesktopServices twin; best-effort by law).
6. **The attach restart legs are exercised per platform**: the
   Windows confirm/run legs are compile-checked everywhere
   (`GOOS=windows` build clean) and exercised only on a Windows box
   (the sandbox has none — the same constraint SESSION-001 recorded);
   the POSIX legs are pinned by this battery.

## Gate evaluation (T-79d's own acceptance: "the four doors live
over the chrometabs engine; the collaborator guards honest when
Chrome is absent") — **PASS**:

- **The doors live over the engine**: the engine cases (13 new tests
  in `internal/app/chromedoors_test.go`) pin every box + line by
  level + verbatim text — the guard ladders, the walled/hand/retry
  scans (incl. the read-failure warnings + the corrupt-cache
  tolerance), the queued box's two variants, the cancel lines, the
  offered-set filter, the >10 boundaries (10/11), the delivery's
  crash guard + the seam contract (the live config rides), the
  failure box's grammar (the five-name cap, the `… and N more` tail,
  the `(a link)` fallback), the attach ladder (every remedy box +
  the port coercion table), the concurrent offered-set hammer
  (THE SENTINEL-SEAT RACE LAW), and the payload field-tags.
- **The collaborator guards honest**: the no-Chrome box is THE
  evidence — `TestAttachMyChromeLadder` +
  `TestHeroUIChromeAttachDoor` pin the reference's own words; the
  click leg below proves it end-to-end on this Chrome-less box.
- **The battery at the close: 2337 main-battery tests race-clean
  over 36 packages (2324 + 13) + 71 webui tests over 4 (67 + 4)**;
  the touched packages (internal/app, internal/cli, webui) 3×
  race-clean post-change (the ritual); gofmt/vet clean over both
  modules; the MAIN go.mod + go.sum + the nested module's files
  byte-unchanged (0 diffs, verified). One real data race in MY OWN
  new shell test (`rig.batchBlock` read+write across the delivery
  goroutine) caught by the race pass and fixed (the fake batch's
  field reads now hold rig.mu — the law pays in bugs, not just
  safety).
- **The `wails build`** over the 13th-rebuilt tree: links clean
  (29.7 s); `git diff --raw`: NO mode flips (the wailsjs files stay
  at their 755 baseline; only content diffs — the regenerated
  bindings carrying the seven doors + the seven payload classes; 35
  exported door functions).
- **The gate legs GREEN** (`run-gate.sh` under Xvfb): first frame
  889 ms / 284 ms (≤ the 1.5 s gate); the 10k flood through the
  production transport with 0 drops (208k lines/s accepted; handler
  max 50 ms/batch).
- **The screenshots VLM-verified** (`t-79d-more-menu.png` +
  `t-79d-hand-click.png`): the four doors LIVE in the menu (dark
  readable, clickable) + the tail honestly disabled. THE VLM
  MISREAD PATTERN STRUCK A FOURTH TIME on the focused crop: a
  mid-menu re-pass read five LIVE doors (Export All Links, Retry
  Failed, View 404 Quarantine, both Retry rows) as "faded/disabled"
  and hallucinated "checked" states on plain rows — settled per the
  law with the byte-level source check (the live keys + the uniform
  render path `more-item` + only `!live` washes, pinned by
  TestHeroflowChromeDoorsLive + TestHeroflowBotQueueDoorsLive +
  TestHeroflowPerTestDoorsLive) + the full-image first pass's
  verdict (the four Chrome doors enabled). No code was touched on
  the VLM's word.
- **The click leg** (`SB_POC_OPEN_MORE=hand`): the probe clicks 🖐
  Hand-deliver walled links LIVE and the door answers end-to-end —
  the "Websites pipeline is off" guard box over the unconfigured
  fixture (the reference's own remedy line, pure config read, no
  Chrome anywhere near) with its OK button. The same
  hermetic-refusal class as T-79c's export pick.

## The reproduce recipe (committed shape)

```sh
cd webui && \
PKG_CONFIG_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu/pkgconfig \
LD_LIBRARY_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu \
wails build -tags webkit2_41
# the menu shot:
SB_POC_OPEN_MORE=menu SB_POC_RESULTS=/tmp/t79d-menu.json \
SB_SCREENSHOT=/tmp/t79d-more-menu.png sh build-support/run-gate.sh 5
# the click leg (the hand-deliver refusal):
SB_POC_OPEN_MORE=hand SB_POC_RESULTS=/tmp/t79d-click.json \
SB_SCREENSHOT=/tmp/t79d-hand-click.png sh build-support/run-gate.sh 6
```
