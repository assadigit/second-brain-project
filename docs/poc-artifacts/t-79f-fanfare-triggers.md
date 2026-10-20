# T-79f evidence — the runner's automatic triggers (the Phase-6 closure row)

**Task (doc 13)**: the end-of-run Chrome offer
(`_maybe_offer_chrome_tab_retry`) + hero.py's caught-up hand scrape +
`_maybe_redo_hand_notes` — the batch-finish/hero-machine surfaces riding
the ALREADY-LANDED `startChromeTabRetry` runner (T-79d), a follow-up
wiring with its own pinned tests, not new engines.

**Pinned reference** (98f4fc7):
`gui/main_window/processing_control.py` 475-520
(`_maybe_deliver_hand_queue` + `_scan_master_hand`), 524-613
(`_scan_master_redo` + `_maybe_redo_hand_notes`), 1197-1264
(`_maybe_offer_chrome_tab_retry` + the dialog), 2060-2224
(`processing_finished`'s tail — the chain's order + the gate);
`gui/main_window/hero.py` 241-344 (the caught-up story: the waiting
report → the hand scrape → the redo check → the work-started return →
the closing line + the drift seal).

## What landed (commit `b3e8264` + this session's close commits)

- **The engines — `internal/app/fanfaredoors.go`** (the
  Settings-method shape of T-79a…T-79e, over the ALREADY-LANDED
  scans — `ScanMasterHandRows` / `ScanMasterRedoRows` /
  `CollectFailedFetchLinks`; no new engines):
  `HandQueuePass` (the end-of-run hand pass's data half: the knob →
  pipeline → vault quiet ladder, the wall/gesture mapping, the
  once-per-session `_hand_delivery_tried` belt), `CaughtUpHandRows`
  (hero.py's caught-up scrape — NO tried-set, the reference's own law
  at that call site), `RedoHandNotesScan` (the redo pass: the quiet
  guards incl. the gathering-delivery + running-batch deferrals, the
  false-success scan over the state DB, the once-per-session
  `_hand_redoTried`→`handRedoTried` belt, the 🔁 line spoken IN the
  engine — the reference's shared-method placement, both call sites
  hear it, with the first row's reason formatted exactly as the
  reference formats it), `ChromeOfferFromResults` (the offer's data
  half: the knob/pipeline/vault ladder, the collect over THIS run's
  results, the degraded dismissed probe OFFERS anyway, the session's
  offered set — only PICKED links ever marked).
- **The plumbing**: `BatchResult` carries the reference's
  batch_summary facts — `Natural` (the end-of-batch presence marker:
  batch_summary is set ONLY at the end-of-batch stage; the early
  returns leave it absent), `Stopped` (`summary['stopped']` — the
  user's cooperative stop), `WebsiteResults`
  (`_website_summary['results']`) — and `WebsiteResultEntry` grew the
  full Result-dict mirror (canonical/note_path/fetch_status) + THE
  FIELD-TAG LAW's tags (the pipeline Result dict's own keys at
  98f4fc7: url/canonical/outcome/note_path/category/subcategory/
  fetch_status/error). **THE FLAG-CLEAR LAW** in `SyncRunBatch`: the
  running flag + the lock go home BEFORE the finished signal crosses
  (the reference's `processing_finished` opens with
  `self._batch_running = False` — the chain's own re-run batch must
  never see a ghost batch), pinned by
  `TestSyncRunBatchFlagClearLawAndFanfareFacts`.
  `EmitBatchFinished(res)` carries the whole finished state; the CLI
  twin's summary agrees on every field (THE CROSS-BINARY
  DISAGREEMENT LAW).
- **The shell — `webui/heroui.go` + `shell.go`**: `onBatchFanfare`
  (the three triggers in the reference's order — hand pass → redo
  pass → offer — each leg TOLERATED with the reference's own warning
  lines, gated on `ev.Natural && !ev.Stopped`), wired into the
  BatchFinished crossing BEFORE the chime; the caught-up block's hand
  scrape + redo check + the work-started holds (the FETCHING render,
  no caught-up claim, no drift seal — hero.py:329-330's return-before-
  `_set_hero_state`); `startRedoRerun` (the reference's worker-THREAD
  shape — asynchronous, so the offer's moment is preserved);
  `emitChromeOffer` + `ChromeOfferAnswer` (the DECLINED line's door —
  the reference's 1258-1261 vs the manual door's CANCELLED one at
  1674-1677; the offered set marks only the picked links).
- **The frontend — `heroflow.js`**: the ONE `chrome:offer` listener
  (the ONE-OWNER-PER-SURFACE law) over the SAME checkable dialog the
  manual door uses (parameterized answer doors); `ChromeOfferAnswer`
  joined the HeroUI binding import (THE BINDING-IMPORT LAW, pinned by
  `TestHeroflowFanfareWiring`); the regenerated bindings carry the
  door + the `BatchResult` fields (content-only, the 755 baseline
  intact).
- **The probe leg**: `SB_POC_FANFARE=offer` (shell.go's autorun) —
  a CANNED `chrome:offer` payload (fixture links, no owner data — the
  same canned-shape law the scan probe's scorecard rides) for the
  screenshot evidence.

## The pinned tests

- **Engine (main battery, 7 new)**:
  `TestFanfareHandQueuePass` (the quiet guards, the wall/gesture
  mapping, the tried belt: once per session per link, a NEW row still
  passes), `TestFanfareCaughtUpHandRows` (NO tried-set — a second
  sync scrapes the same rows again; the batch-busy belt),
  `TestFanfareRedoHandNotesScan` (both deferral guards, the
  false-success row over the delivered-queue seed, the exact 🔁 line
  with the first reason, the quiet tried belt),
  `TestFanfareChromeOfferFromResults` (the guards, the collect:
  review+fetch-failed + hard-miss both ask, dedup, loopback +
  non-http drop, the dismissed filter, the offered filter),
  `TestFanfareChromeOfferNilResults` (a batch whose websites phase
  never ran offers nothing), `TestFanfareOfferFromStateKeeps
  UnprovableLinks` (a DIRECTORY at cache.db breaks every open — the
  degraded probe offers everything),
  `TestFanfareWebsiteResultsMirrorKeys` (THE FIELD-TAG LAW:
  marshal-pinned tags on the entry + the crossing's new fields).
- **The command pin**: `TestSyncRunBatchFlagClearLawAndFanfareFacts`
  (THE FLAG-CLEAR LAW: `BatchBusy()` false INSIDE the BatchFinished
  handler; the facts ride the event; the existing crossed-payload
  pins extended to the new tags — tests ADAPTED to the grown
  contract, never weakened).
- **Shell (webui, 7 new)**: `TestHeroUIFanfareChainGate` (the gate:
  not-natural + stopped keep the chain silent; the natural crossing
  fires the hand pass with the reference's exact line; the tried belt
  holds one delivery per link per session),
  `TestHeroUIFanfareChainRedoLeg` (the redo leg end-to-end through the
  REAL Commands — the 🔁 line + the asynchronous direct re-run batch
  + the recursion terminates through the tried belt),
  `TestHeroUIFanfareCaughtUpHandScrape` (the caught-up hand scrape:
  the exact 🖐 line, no caught-up claim, the disabled FETCHING
  render, the delivery's own re-run batch),
  `TestHeroUIFanfareCaughtUpRedoPass` (the caught-up redo: the 🔁
  line, the direct batch, the quiet bar format),
  `TestHeroUIFanfareCaughtUpPlainClaimIntact` (the regression pin:
  the plain caught-up claim + drift-seal path stays exactly as T-77
  landed it), `TestHeroUIChromeOfferAnswer` (the DECLINED line, the
  declined link stays offerable, the picked link marked + delivered +
  never re-offered), `TestHeroflowFanfareWiring` (the listener + the
  dialog's answer parameterization + THE BINDING-IMPORT pin).

**The battery at the close: 2356 main-battery tests race-clean over
36 packages (2348 + 8) + 84 webui tests over 4 (77 + 7)**; the
touched packages (internal/app, internal/cli, internal/githubpipeline,
webui) 3× race-clean post-change; gofmt/vet clean over both modules;
the MAIN go.mod + go.sum + the nested module's files byte-unchanged.

## Gate legs GREEN (the re-run over the 15th-rebuilt webkit tree)

- First frame **802 ms** (the 1.5 s gate; the second leg's fresh boot
  281 ms); the 10k flood through the production transport: **0
  drops**, handler max **24 ms/batch**, 41 batches —
  `t79f-gate-results.json`.
- The screenshots (VLM-verified, BOTH passes agreeing with the
  byte-level sources + the pinned tests — no misread incident this
  session): `t79f-offer-modal.png` — the end-of-run offer modal over
  the canned fixture links: the heading "2 link(s) couldn't generate
  content", BOTH rows ticked (the reference's all-ticked default for
  THIS dialog), the two enabled buttons "Not now" / "🤖 Retry in
  Chrome now" over the themed desk; `t79f-gate-desktop.png` — the
  themed window fully rendered (the SYNC chrome, the desk's live
  rows).

## The recorded divergences

1. **The `_closing`/`isVisible` guards are not ported** (the
   reference's hidden-window law): the Wails window is the
   always-visible surface — the same treatment the T-79d/e doors
   gave them.
2. **The `_delivery_rerun_pending` flag + the stashed-scorecard story
   are not ported**: they feed v0.54's batch-scorecard STASH (the
   held scorecard the delivery flushes + merges), which rides the
   Batch Complete MODAL (`_show_batch_success_modal`) — no T-79 row
   names that modal (the port's fanfare is the bar marks + the
   chime, landed at T-77). The re-run batches' own finishes tell
   their own stories through the same crossing. If/when the
   scorecard modal lands (a future row / the owner's call), the
   stash + the merge + the flag land WITH it — T-79d's divergence #2
   record anticipated exactly this surface.
3. **The chime's moment**: the reference runs the chain BEFORE
   `_celebrate_batch` (the chime plays when the SCORECARD opens —
   after the offer's blocking modal is answered); the port's
   two-phase shape crosses the chime with the batch-finished event
   while the offer modal may sit open. The audible difference is
   sub-second; the trigger moments themselves are preserved (the
   chain rides the crossing in the reference's order).
4. **The offer's modal swap**: the reference's blocking QDialog means
   a second offer cannot fire while the first sits; the port's
   one-overlay law REPLACES the modal if a second offer event crosses
   before the first is answered (the offered-set marks only at the
   answer, so a re-offer re-shows the same links — fresh data, same
   question).
5. **The caught-up hand-delivery hold**: hero.py:329-330's
   `return`-before-`_set_hero_state` leaves the button at the
   disabled FETCHING render while the delivery gathers — INCLUDING
   the all-failed-delivery case (no batch starts, the render holds
   until the next sync/mirror transition). Ported faithfully, quirk
   included (THE CROSS-BINARY DISAGREEMENT LAW: the edge branch is
   the reference's own).
6. **The redo pass's re-run is asynchronous** (`startRedoRerun`'s
   goroutine): the reference's `_start_worker_with_urls` starts a
   worker THREAD — the caught-up flow never waits on it. The
   port's shell doors otherwise run their own batches synchronously
   (T-77's blocking-door law), but the CHAIN's re-run follows the
   reference's thread shape so the offer's moment (leg 3) is not
   delayed by leg 2's whole batch.

## Acceptance (doc 13's row)

"the three triggers fire at their own moments over the landed runner;
the offered-set + hand-pass + redo-pass laws pinned" — **PASS**: the
chain fires at the batch-finished crossing (gated on the natural
completion + not-stopped verdict, in the reference's order), the
caught-up triggers fire inside the fetch leg's caught-up block with
the work-started holds, and every seat law (the two tried belts, the
offered set's only-picked marking, the declined-stays-offerable law)
is pinned by the tests above.
