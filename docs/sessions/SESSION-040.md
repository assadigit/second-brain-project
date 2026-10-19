# SESSION-040 — 2026-10-20 — Phase 6: T-77 COMPLETE (the main window chrome — the hero button machine, the progress row, the More menu, the sentinel labels, the battery modal, the sounds) — 2252 main-battery tests over 36 packages + 56 webui tests over 4; the thirty-fifth full wipe repaired BEFORE work (the 8th webkit rebuild); TWO reality fixes the re-read caught (the nil-latency panic + the sentinel data race)

## Goal (from NEXT-SESSION.md at start)

SESSION-040 = **T-77 (L): THE MAIN WINDOW CHROME** per doc 13's row —
the hero button machine (the sync → fetching → process → running
state machine on ONE button — hero.py), the progress row's live
formats, the More menu, the sentinel labels (Net/AI/Proxy), the
dialogs polish (the ConnectionTestDialog's battery modal — the 4 rows
+ Start Syncing gated on all-ok), the sounds (BatchSound — the WAV
assets an OPEN item on the books). ACCEPTANCE: hero state machine
tests; startup cadence test.

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-fifth full wipe repaired BEFORE work** (the work repo,
  the reference, the toolchain, the venv, the module cache, and the
  webkit tree all gone at boot; everything rebuilt from the committed
  recipes): both repos re-cloned with the ephemeral `GIT_CONFIG_COUNT`
  form (no PAT persisted anywhere; the work repo's tip verified at
  `12f35ae` = SESSION-039's close — one commit BEYOND the prompt's
  `4626a39` expectation, the report+state-card commit; the tree clean;
  the reference checked out at `98f4fc7` = PINNED-GITCURATOR-REV).
  go1.27.2 downloaded from go.dev and sha-verified against BOTH the
  committed 64-char pin (TOOLCHAIN.md Am.2) AND go.dev's published
  `?mode=json` listing. The venv (Python 3.12.15, stdlib-only), the
  wails CLI v2.16.0 (GOBIN install), and the rootless webkit tree
  (144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dangling
  dev links as HARDLINKS to the resolved system targets — **Am.4's
  fallback again, the 8th webkit rebuild: symlink(2) refused at BOOT
  ("Creating symbolic links is not allowed")**; the wails build + the
  gate leg the proof). The re-verification law executed before any
  new work: the FULL main battery + the webui battery green pre-work
  — **exactly the recorded state: 2230 over 36 + 40 over 4,
  race-clean**. (The BACKGROUND-PROCESS lesson paid once: the
  backgrounded downloads died when the calls returned — everything
  re-run in foreground chunks.)

- **T-77 slice (a) — commit `9cd7586`: the chrome's pure engines +
  the retroactive fixes + the sounds.**
  - `internal/app/herostate.go` — **HeroMachine**, hero.py's GUI-only
    state machine as a pure engine: the four states (sync → fetching →
    process → running), the click routing with the guards as data (the
    vault box, the offline advisory's ask + the owner's-word-wins
    continuation, the nothing-to-sync split **re-worded for Option B's
    Worker queue** per ADR-0004 Am.3 — "Set the Worker's URL (or pair
    the app) in Settings → Bot"), the fetch routing (_after_sync_fetch:
    the both-pipelines arithmetic (v0.24.1) + the 🟢 line's parts law,
    the import-ready path, the caught-up law (the closing-line choice
    by the waiting count + the waiting-rows report's grammar + the
    deferred drift seal), the error restore), the mirror
    (_sync_run_button's defensive law — the running flag wins over
    ANY state, fetching included), the stopping render (v0.32's
    no-double-click), Render's exact labels/glyphs/fills/tooltips
    (the long FULL SYNC tooltip verbatim), ConfirmLargeBatch (>10
    states the exact count; the "the bot queue" source word verbatim
    from bot_queue.py:721), the progress row's live formats (the
    SegmentBar vocabulary), and the sentinel labels' paint laws
    (Net/AI/Proxy — the words + dots + semantics + tooltips, the
    proxy vocabulary off/connected/unreachable verbatim) — **21
    tests**.
  - **THE RETROACTIVE FIELD-TAG FIX** (SESSION-038's law applied to
    T-71's tagless shapes — a REAL display bug family): the bus
    payloads (Progress, Status, BatchFinished, JobProgress, DiskFull,
    ModelChanged) crossed to the frontend WITHOUT JSON tags — Go
    marshaled "OK"/"Pos"/"Text" while app.js read ok/pos/text — **so
    the progress bar NEVER filled during a real batch and every
    finish rendered "❌ Failed"**. Every crossed payload re-tagged,
    pinned by `TestCrossedPayloadFieldTags`; `BatchFinished` gains
    `needs_retry` (the retry chime's selector, v0.40.0).
  - **The cooperative stop door**: `batchEngine.requestStop` (the
    runner's between-items flag covers the GitHub loop, the rate-limit
    waits, and the websites phase — `shouldContinue` reads
    `runner.RunningNow()`; the `stopRequested` seat keeps a warmup
    arrival from being lost; `buildRunner` closes the lost-stop
    window) + the App's `batchStopHandle` seat (the arm token — a
    late clear from a finished engine never disarms a newer batch) +
    `Commands.StopBatch` (the honest no-op when no batch runs);
    `BatchResult.FailedLinks` reads the tracker's failed count at
    finish.
  - **The sentinel pass's full emission**: `SentinelSnapshot` (the
    complete net+ai+proxy probe set) + `OfflineVerdict`'s transitions
    (warn-now/fell-offline/back-online lines — the reference's exact
    sentences; the ONE launch modal rides the verdict word) + the new
    `Sentinel` bus event (the chrome's paint source; the probe
    injectable per the HERMETIC-SEAM law). **TWO reality fixes the
    re-read caught**: (1) the old 🛰️ line dereferenced `net.LatencyMs`
    unconditionally — ProbeInternet leaves it nil on the offline path,
    so an offline pass PANICKED into the crash-isolation and the
    labels never painted (the latency is nil-safe "n/a" now);
    (2) the `sentinelOnline` seat RACED — the lifecycle's recurring
    tick and the save law's 250 ms re-arm can pass concurrently
    (found by the race detector on a re-run; the read-transition-write
    is ONE unit under `sentinelMu`, mirroring the reference's
    GUI-thread serialization; eight clean -race runs after the fix,
    and the full battery stable across three consecutive runs).
  - **The sounds SHIPPED** (the deferred-debt note "until T-77/T-80
    ship the files" closes): success.wav + retry.wav + the README
    from the reference (sha-identical copies; synthesized,
    license-free), embedded in the assets package with the
    **disk-first law** (an owner's APP_DIR/assets/sounds copy always
    wins — the swap-the-chime law; the embed backs up). The gate
    split: the quiet `soundFor` (the exact batch gate — the
    reference's `_play_batch_sound` speaks nothing on a switched-off
    chime) + `PreviewSound`'s own 🔇 lines + `Settings.BatchChime`
    (the shell's door). The T-73 sound test updated to pin the NEW
    laws (disk-wins, embed-fallback, the poisoned-session
    one-notice) — tighter, not weaker.
  - Settings grows `WorkerQueueEnabled` (the hero's queue-configured
    guard — workerclient.IsEnabled over the live config) +
    `SaveDarkMode` (the More menu's toggle). The MAIN go.mod
    byte-unchanged.

- **T-77 slice (b) — commit `c14fb8f`: the shell.**
  - `webui/heroui.go` — the SIXTH bound struct: **HeroClick THE one
    click door** (the guards → the boxes; the offline ask →
    `HeroAnswerOffline`; the >10 confirm → `HeroRunQueue`; the fetch
    leg over SyncFetchQueue with the classification lines + the Sync
    view's freshness hook + the caught-up extras (the waiting-rows
    scan over websitepipeline.ScanMasterWaitingRows + the drift seal
    + the closing line); the process leg (the fetched queue, both
    pipelines); the import leg (the Input mode); the stop leg
    ("⏹️ Stopping..." + the winding-down render)); the **200 ms mirror
    loop** (the reference's _run_mirror_timer — a batch another door
    started flips this button too); **TestConnection** — the battery
    door over connectioncheck.RunLocalChecks (the guards, the live
    cc:section/cc:result events, the log's full render lines, the
    degenerate-sections guard, the settled verdicts + the all-ok
    gate); the More menu's live doors (MoreOpenDashboard over the
    Settings URL law, MoreDarkMode, MoreRetryBacklog/MoreRetryMaster
    over the retry batch modes, ChromeBoot — the persisted dark_mode
    boot paint) — **16 shell tests** incl. the frontend laws read
    where they render.
  - `shell.go` — the Sentinel crossing (the snapshot paints the
    labels; the hero machine's offline advisory reads the same
    verdict) + the batch-finished enrichment (the hero's SYNC restore
    + the chime payload through Settings.BatchChime — the embedded
    struct's tags ride through) + the mirror's shutdown stop;
    `main.go` binds the sixth struct.
  - `frontend/heroflow.js` — the chrome's frontend half: the
    hero:state renderings (the label, the lucide glyph with the spin,
    the disabled flag, the danger fill, the tooltips), the click door
    + the guard boxes + the offline question (the reference's own
    text) + the confirm gate, the progress row's live formats + the
    status texts + the terminal marks + **THE BATCH CHIME** (<audio>
    — the QSoundEffect twin, never an error), ⚡ Fast Sync over the
    PagesB door with a busy render, the sentinel labels' paint (the
    dot + the word + the tooltip + the semantic class; the warn-now
    box's own text), **THE BATTERY MODAL** (the four rows —
    Vaults/LLM/GitHub/the Worker, the braille spinner, the compact
    one-line items with the full text on the tooltip, Start Syncing
    born disabled + green only on all-ok + the deferred hero
    handoff), **THE MORE MENU** (the reference's own action list
    VERBATIM — the live doors wired, the not-yet-landed engines
    HONESTLY DISABLED with their landing notes, the Settings submenu
    carrying Dark Mode). `app.js` hands the bar + the chime to
    heroflow (one owner per surface; the log rows stay);
    `index.html` — the live CTA band + the sentinel dot/word
    structure; `style.css` — the danger fill, the sentinel dots, the
    battery rows, the menu.

- **The docs — the third commit** (this one): doc 13's T-77 row
  COMPLETE; doc 12's SESSION-040 entry; spec-gui-ui's dated landing
  note (the seven port-level decisions incl. the nothing-box
  re-wording + the caught-up Chrome-door deferral); spec-gui-window's
  dated note (the snapshot emission + the two reality fixes + the
  cadence acceptance); spec-gui-pages-a's Sound note (the assets
  ship, disk-first); README's SESSION-040 status; webui/README.md +
  internal/app/doc.go maps.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-77's acceptance = hero state machine tests + startup
cadence test — both green with proof:

- **The hero state machine tests**: the 21-test pure battery (the
  born law + the SYNC render's verbatim tooltip; the four renders;
  the click routing through every guard incl. the offline
  ask/continuation/refusal; the fetch routing's every branch with the
  exact lines; the mirror's three laws incl. the running-flag-wins
  defensive law; the stopping render + the finish restore; the >10
  gate's exact text; the paint laws' full vocabulary; the format
  constants) + the 16 shell tests (the vault guard; the
  nothing-to-sync split + the import batch; the offline advisory's
  full continuation incl. the caught-up closing line; the fetch flow
  end-to-end over the faked queue seams (the 🟢 line, the PROCESS (2)
  render, the direct batch, the SYNC restore); the error restore; the
  >10 confirm gate (an unanswered confirm never runs; Continue runs
  it); the battery door's four rows + the all-ok gate + the busy
  refusal over a REAL held batch lock; the sentinel seat feeding the
  advisory; the crossing payloads' field tags; the stop leg's
  winding-down render + the mirror restore over a blocking batch; the
  frontend laws read where they render — the field-tag reads, the
  verbatim More-menu label list, the battery row words, the sentinel
  paint laws) + the two app-level tests (the sentinel pass's own
  emission law over the hermetic probe seam — the warn-now/first/
  offline/back-online/fell-offline transitions with the exact lines
  and the nil-latency regression; the stop door's arm/clear wiring).
- **The startup cadence test**: the lifecycle's own battery
  (TestStartupPassesRunAtTheirCadences, TestStartupPassOrderObserved,
  TestStartupPassCrashIsolated, TestStartupCancelledBeforeDelay +
  the shutdown graces) re-verified green in the full run, plus the
  sentinel pass's emission test proving the labels' data door fires
  from the pass.
- **The end-to-end evidence**: the wails build links clean over the
  8th-rebuilt webkit tree (29.6 s); the gate leg GREEN (first frame
  **908 ms ≤ the 1.5 s gate** on the post-rebuild cold run; the 10k
  flood through the FULL production transport **555,555 lines/s,
  0 drops**, 168 ms drain, 41 batches, **26 ms max handler**; the UI
  probe's rows_appended 10,001); the screenshot **VLM-verified**: the
  sentinel labels painting LIVE from the real snapshot — **"Net:
  online" (green dot), "AI: off — unreachable" (red dot), "Proxy:
  off" (gray dot)** — the SYNC hero enabled with its accent fill, ⚡
  Fast Sync + 🔍 Scan + 🔌 Test Connectivity + More ▾ all present, NO
  disabled placeholders left in the CTA band, the progress row
  reading "Ready", the desk filled.
- **The battery at the close**: **2252 main-module tests race-clean
  over 36 packages** (2230 + 22 — arithmetic closed) + **56 webui
  tests over 4 packages** (40 + 16); **stable across three
  consecutive full -race runs** (the flake hunt: the sentinelOnline
  data race caught on the second re-run, fixed, then eight clean
  package runs + three clean full runs); gofmt/vet clean over both
  modules; the MAIN go.mod byte-unchanged all session; no mode flips
  committed (the regenerated wailsjs bindings ride the committed 755
  baseline — only content diffs); no stray webui/webui binary; the
  tree clean at the close.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-77 row
  COMPLETE (the full record incl. the 35th-wipe repair + the two
  slices' hashes + the two reality fixes).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-040
  entry.
- `docs/port-blueprint/module-specs/spec-gui-ui.md` — the dated
  landing note: the seven port-level decisions (the pure machine; the
  nothing-box re-wording for Option B; THE RETROACTIVE FIELD-TAG FIX;
  the More menu's honest-disabled law + the per-test dialogs NOT
  T-77's; the caught-up path's Chrome-door legs deferred (the
  reference's own optional collaborators); the paint in Go; the
  embedded disk-first sounds).
- `docs/port-blueprint/module-specs/spec-gui-window.md` — the dated
  note: the snapshot emission, the two reality fixes, the cadence
  acceptance, the shutdown chain unchanged.
- `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the Sound
  dated note (the assets ship, disk-first, the quiet gate split).
- `README.md` — the SESSION-040 status entry.
- `webui/README.md` + `internal/app/doc.go` — the new files in the
  maps.

## Blockers & decisions needed

- No new blockers. The authorized runtime set is UNCHANGED (the four;
  the MAIN go.mod byte-unchanged all session — verified at every
  commit). No new dependencies (the WAVs are embedded data, not a
  dependency).
- Carried (no rush): the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (thirty-first
  session carrying it); x/net/html keep-or-swap; the answering-window
  belt + the T-71 merged-verdict question (SESSION-039's record).
- **The deferred GUI doors, honestly recorded** (the More menu's
  disabled tail): the per-test dialogs (proxy/ollama/cloud/llamacpp/
  telegram-and-github), the linking tools (recall hooks, link
  suggestions, validate/verify/recategorize), the bot-queue tools
  (verify-all/export-all/retry-failed as dialogs), the Chrome doors
  (hand-deliver, hand-rows scrape, tab-retry, attach), decommission,
  the export backup zip (the packaging phase). These ride their own
  tasks — none is on doc 13's remaining rows as a named task, so the
  NEXT session should surface the bookkeeping question for the owner
  (a small tasks addendum vs folding into T-78/Phase 7's parity
  report).
- **The caught-up path's Chrome-door legs deferred** (spec-gui-ui's
  note): the hand-rows scrape + the hand-note redo pass ride the
  Chrome-doors GUI task (the engines live; the reference itself guards
  them as optional collaborators in hero.py).
- The sandbox's symlink block persists (TOOLCHAIN.md Am.4) — the 8th
  webkit rebuild needed the hardlink fallback again, still at BOOT.
- The deferred test debt on the books is unchanged (the T-74 list
  MINUS the sound WAV assets — that item closes here).

## Owner actions requested

- None urgent. The standing calls (dissolution, vanilla-vs-vite)
  remain open.
- When convenient: the GUI-doors bookkeeping question above (the
  More-menu disabled tail needs a home on the task list — a tasks
  addendum or the Phase-7 parity report's scope).

## Handoff

- **SESSION-041 = T-78 (S): the side-by-side verification vs the
  Python app** per doc 13's row — the lock mutual exclusion across
  both binaries (ADR-0008's same-data law: the app.lock, the merge
  laws, the byte-comparable config). The LAST Phase 6 task; Phase 7
  (T-80..T-83) follows.
- The state card v41 (`docs/NEXT-SESSION.md`) carries the full
  read-first list + the SESSION-041 prompt.
- Push state: slices (a) + (b) committed and pushed (the ephemeral
  PAT at each push; never stored, never written); the docs commit
  rides this close.
