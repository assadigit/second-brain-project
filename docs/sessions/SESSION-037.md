# SESSION-037 — 2026-10-20 — Phase 6: T-74 COMPLETE (Settings pages B + THE CLI-PARITY SHELL WIRING — the carried debt closed; the queue report rendering + the fast-sync laws + the wiring's own battery) — 2060 main-battery tests over 35 packages + 28 webui tests over 4; the thirty-second full wipe repaired BEFORE work

## Goal (from NEXT-SESSION.md at start)

SESSION-037 = **T-74 (L): SETTINGS PAGES B** per doc 13's row — Bot,
Dashboard, Sources, Backup (+ radar, fast sync) — the pages matched
BY NAME to their spec files (the R-0 note: bot_queue's spec lives in
spec-gui-pages-a.md despite the A/B label; backup_seal's in
spec-gui-pages-b.md). **THE SESSION'S FIRST SLICE — the carried
wiring debt**: the CLI-parity shell wiring on the app surface (the
warmup family, the inputs fetchers, the websites phase, the Phase-5
CLEAR — engine.go's runImpl composition re-hosted on internal/app
with its own tests, landed as its own committed slice). ACCEPTANCE:
the wiring slice's own tests + queue report rendering from a canned
job result + the fast-sync laws (the ported pure halves already
tested). If capacity runs short: the wiring slice done properly with
proof beats both half-done.

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-second full wipe repaired BEFORE work** (both repos,
  the toolchain, the venv, the module cache, AND the webkit tree all
  gone at boot; everything rebuilt from the committed recipes): both
  repos re-cloned with the ephemeral `GIT_CONFIG_COUNT` form (no PAT
  persisted anywhere; the work repo's tip verified at `01d6646` =
  SESSION-036's close, the tree clean; the reference checked out at
  `98f4fc7` = PINNED-GITCURATOR-REV, byte-verified); go1.27.2
  sha-verified THREE ways (local = go.dev's JSON = the committed
  64-char pin); the venv (Python 3.12.15); the wails CLI v2.16.0; the
  rootless webkit tree (144 debs → dpkg -x → the .pc prefix rewrites →
  the 26 dangling dev links — **see the deviation below**). The
  re-verification law executed BEFORE any new work: MAIN go.mod
  byte-verified (the authorized four @ go 1.24.1; the nested webui @
  1.25.0 + wails + the parent replace), the full main battery (2020
  tests / 35 packages race-clean) + the webui battery (23 / 4) +
  vet/gofmt + the clean-tree check all green pre-work.
  - **THE SYMLINK-SYSCALL DEGRADATION (recorded in build/TOOLCHAIN.md
    Am.4)**: mid-rebuild the sandbox began refusing `symlink(2)`
    outright (`ln -s` AND Python `os.symlink`, after the repo
    convenience links had worked earlier in the session). The 26 dev
    `.so` links therefore landed as **hardlinks to the RESOLVED system
    targets** (readlink -f first — the system `.so.0` files are
    themselves versioned symlinks): same bytes, same link-time
    resolution, the `wails build` CGO link + the headless gate leg the
    proof the tree is sound. The committed recipe stays the symlink
    form; Am.4 is the fallback record for any future session that
    hits the same block.

- **T-74 slice (a) — commit `a4619fe`: internal/app/batchwire.go, THE
  CLI-PARITY SHELL WIRING — the debt T-73 carried, CLOSED.**
  engine.go's runImpl composition re-hosted on the app surface
  (~1000 lines over the App's six new injectable seams —
  ProbeLlama/DetectLlama/AutostartServe/CloudListModels/
  WorkerMirrorPush/WriteLiveConfig, all wired to the production doors
  in New(); BatchOptions gained Mode + ImportFile):
  - the dry-run switch flip/restore + the crash wrapper (the 💥 line
    + the honest finish verdict — exactly-once) + the pipeline-gates
    + the mode dispatch (direct | import | review_retry |
    master_retry — the Sync view's retry doors now have their engine;
    the Telethon mode is Option B's accepted loss);
  - THE INTAKE with the banned/self drops + the _inbox tables (websites
    OFF) or the queued line (ON); the manifest line;
  - THE WARMUP FAMILY: the Ollama list/timeout + the autostart ladder
    (spawn + 16×1.5 s polls + the FileNotFoundError refusal) + the
    warmup chat + the model adaptation ladder (the single-model
    auto-switch, the model-prompt door over the shell's seam, the
    best-stand-in auto-pick) + the llama.cpp re-probe/detect + URL
    switch + model resolve + the cloud /models pre-flight with the
    CloudLLMError shape's two excepts;
  - the setup family: the dry-run shadow cache, the note-state, the
    CONFIRMED-dead set, THE REPOS BANISHMENT pass, THE SETTLED REPOS
    LEDGER, the vault index + SetVaultIndex, the missing-repo
    backfill, the §4.4 start-of-run passes for BOTH vaults, the
    dismissed set, the pre-batch snapshot;
  - THE WEBSITES PHASE: the full websitepipeline composition (the
    banned/self drops with the manifest marking, the settlement, the
    websites vault's own index, the LLM router honoring the per-task
    overrides, the banish gate round-trip, the pre-drop, the due
    retries + the backlog retry modes, the Website Directory rebuild,
    the _inbox prune, the summary + the done line);
  - THE FINISH CHAIN: the seals in the reference's exact order
    (VaultSeal → Good Repos → Websites → Personal) + THE PHASE-5
    CLEAR (the vault-mirror push is the ack — every batch,
    best-effort, the disabled Worker silent; the all-clear verdict
    gated by success AND bot_source) + the dry-run report path +
    the model-changed law (the live map + config.Save).
  - **batchwire_test.go — 19 tests**: the pipeline gates; the warmup
    family's doors (the dead Ollama + the FileNotFoundError refusal +
    the autostart-comes-up ladder + the single-model auto-switch +
    the model-prompt door + the declined auto-pick + the llama.cpp
    abort + the detect-and-switch save + the cloud pre-flight's three
    verdicts over the CloudListModels seam with the class-name token
    law); the import fetcher (the one-link grammar + the _inbox
    tables + the counters + the missing file); the dry-run law (the
    banner + the entry count + NO vault writes + NO seals + the
    report under APP_DIR); the finish seal order; THE PHASE-5 CLEAR
    (the mirror-push legs incl. the disabled-Worker silence + the
    verdict shapes over crafted trackers + the end-to-end bot batch
    with the push seam); the websites phase's wiring (ON with a vault
    runs the phase; ON without a vault keeps the batch with the honest
    warning).

- **T-74 slice (b) — commit `6d479fb`: internal/app/settingsb.go, the
  pages B's engines** (the specs matched BY NAME — bot_queue +
  dashboard per spec-gui-pages-a, backup_seal per spec-gui-pages-b):
  - **THE QUEUE REPORT RENDERING (the acceptance item)**:
    `RenderQueueReport(cfg, rep)` builds the reference's exact
    labeled-count text from a canned QueueReport — every line name
    verbatim (Total GitHub URLs in bot / ✅ Already in vault / 🗑️
    Decommissioned / 🤝 Repos settled / 🚫 Blocked / 🔒 Self / ⏳
    Pending / 🔗 Non-GitHub), the websites sub-table with its own
    counts + the 📚 vault-notes line, the OFF and ON-no-vault
    advisories, the pending lists with the 50-cap + "… and N more",
    the caught-up family (settled vs plain + the confirm hint), the
    error/disabled shapes; `QueueBadge` (the zinc ⏳ / mint ✅
    two-state law, the count covering BOTH pipelines); QueueLogLines
    (the desk family incl. the two settled lines + the actionable
    hints); the username guard (Option B drops only the Telethon
    guards — recorded in the spec's dated note);
  - the DASHBOARD engines (dashboard.py): ScanDashboardStats (the
    first-1000-chars frontmatter walk) + RenderDashboardText (the
    exact lines) + QuarantineView (the top-30 table + the threshold
    footline) + ClearQuarantine + ResetQuarantineLine + the threshold
    group's clamped merge save + the UNDO ENGINE (ReadUndoList's
    no-file/empty/gone/no-vault doors + the top-10 confirm body +
    ApplyUndoLastBatch's best-effort deletes + the processed_repos row
    removal via the store's new ForgetProcessedNotePath + the once
    law) + UndoLogLine;
  - the SOURCES engines over the injectable SourcesFetch seam (the
    no-url guard, the error path, the RSS regex + the unique-set law,
    the Reddit .json walk, the empty answer — every desk line rides
    the bus);
  - the BACKUP page: the three status-dot ladders' every rung
    (BackupFolderStatus / VaultSealStatus / GoodReposStatus — the
    rungs verbatim) + the ONE BackupValues save group (the int-law
    heal, the URL strip, the nested groups through the merge) + the
    Bot + Quarantine groups + BackupNowDialogGuards + the Open
    Dashboard URL law + OpenDashboardBrowser;
  - the snapshot embeds the pages B fields (one payload builds every
    page); QueueReport gained the additive DuplicatesRemoved +
    WebsitesVaultIndex fields (the classification half records the 📚
    count).
  - **settingsb_test.go — 21 tests** (the canned-report battery incl.
    the cap + the caught-up family + the error shapes; the badge; the
    log family; the guard; the dashboard scan over a fixture vault +
    the exact text + the quarantine round-trip + the threshold clamp;
    the undo engine's every door + the top-10 law + the row removal +
    the once law; the sources engines over the fake seam; the ladders'
    every rung + the save groups' laws + the guards + the URL law; the
    @-strip; the snapshot's pages B fields).

- **T-74 slice (c) — commit `3091d76`: the pages on the production
  shell.**
  - webui/settingsuib.go — the pages B's binding wrapper (the THIRD
    wails struct, bound in main.go + shell.go): CheckBotQueue (the
    username guard + the pre-fetch config save + SyncFetchQueue +
    the render + the badge + the desk lines), ProcessBotQueue /
    ProcessSourcesUrls (the >10 confirm gates are the frontend's),
    RefreshDashboard / ClearQuarantine / SaveQuarantineGroup /
    Read+ApplyUndoLastBatch, FetchFromSources, SaveBackupGroup +
    RefreshBackupLadders (the keystroke law reads the live field
    state) + BackupNowGuards + the job buttons (Backup Now / Seal
    Now / Resync Mirror / Publish Now / 📡 Detect Repos) + **FastSync**
    (the bound door T-77's hero machine wires — the laws themselves
    are the already-tested pure halves) + OpenDashboard; the generated
    bindings land in frontend/wailsjs/go/main/PagesBUI.js + models.
  - frontend/settings.js — the four pages' HTML twins at the
    reference's own sidebar order (dialogs.py's TAB_INFO + ui.py's
    append order: **Dashboard after Input, then Bot, Sources, Backup,
    Sound** — the glyphs bar-chart-3 / bot / satellite-dish / archive
    from the locked pack, the hints verbatim): the Dashboard page
    (Refresh primary + Undo Last Batch danger + the results panel +
    the 404 Quarantine group with the spinbox's valueChanged save +
    the themed Reset confirm), the Bot page (the info callout, the
    username/token rows with the @-aware save-on-change, Check Queue +
    the zinc ⏳/mint ✅ badge + Process All + the queue display), the
    Sources page (the URL row + Fetch + the results + Process Fetched
    URLs), and the Backup page (the four compacted groups — every
    label, tooltip and '?' help text verbatim; the live status dots;
    the radar's full tooltip; the Resync confirm; **the Restore button
    honestly disabled with its landing note** — the safety-net restore
    engine rides the packaging phase, recorded in spec-gui-pages-b's
    dated note). The zinc/mint badge CSS rides the locked semantic
    tokens (the reference's zinc/mint QSS has no token-table twin —
    recorded). **The evidence door**: SB_POC_OPEN_SETTINGS now carries
    the section name (the autorun probe opens ANY page). The
    pre-existing T-73 CSS typo `.pageidden]` fixed (`.page[hidden]` —
    the UA default had masked it; reality wins).
  - settingsuib_test.go — the wiring smoke (5 tests): the queue
    check's guard + render + badge + payload over the fake queue poll
    (the queue-history settlement's honest line asserted), the process
    door's source law, the dashboard doors, the backup doors, and the
    pages B glyphs-in-pack law.

- **Spec updates riding the landing (reality wins, same commits)**:
  doc 13's T-74 row COMPLETE (the full three-slice record); doc 12's
  SESSION-037 entry; README's SESSION-037 status; webui/README.md's
  map (settingsuib.go + the pages B in settings.js); internal/app/
  doc.go's file map (settingsb.go + batchwire.go); spec-gui-pages-a's
  bot_queue + dashboard dated notes (the Option-B drops: Process New
  + Mark All as Read never render; the verify/recategorize/dead-links
  families stay T-76/T-77's); spec-gui-pages-b's T-74 landing note
  (the Restore button's honest disabled state + the badge token
  note); build/TOOLCHAIN.md Am.4 (the symlink-syscall deviation).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-74's acceptance = the wiring slice's own tests + the queue
report rendering tests + the fast-sync laws — all three green with
proof:

- **The wiring slice's own tests**: 19 batchwire tests over the App's
  fakes — the warmup family's every door (the dead-server abort with
  the FileNotFoundError refusal, the autostart ladder's recovery, the
  single-model auto-switch with the live-config save observed, the
  model-prompt door + the declined auto-pick, the llama.cpp abort +
  the detect-and-switch save with the URL + model landing in the live
  map, the cloud pre-flight's three verdicts), the import fetcher's
  grammar + counters + the _inbox table landing, the dry-run law (NO
  vault writes, NO seals, the report under APP_DIR/reports/dry-runs),
  the finish seal ORDER (VaultSeal → Good Repos — the reference's
  exact order), THE PHASE-5 CLEAR's legs (the mirror-push seam's
  three outcomes + the disabled silence + the ✓/⏸️/no-manifest
  verdict shapes + the end-to-end bot batch), the websites phase's
  wiring.
- **The queue report rendering tests**: the canned-result battery —
  the full text's every labeled-count line verbatim, the settled
  lines, the websites sub-table + both advisories, the 50-cap + "…
  and 10 more", the caught-up family, the error/disabled shapes, the
  badge's two states, the log-line family, the username guard.
- **The fast-sync laws**: the ported pure halves were already tested
  (vaultseal's own battery, test_driftseal + test_fastsync re-ported
  at T-66); the FastSync DOOR is bound on the shell for T-77's hero
  machine — no law weakened, no test touched.
- **The end-to-end evidence**: the wails build links clean over the
  rebuilt webkit tree (3.6 s); the gate leg re-run GREEN — first
  frame **283 ms / 942 ms ≤ the 1.5 s gate** (two runs, the second
  over the settled box), the 10k flood through the FULL production
  transport at **476,190 lines/s producer-side, 0 drops, 172 ms
  drain, 26 ms max handler**; the screenshots **VLM-verified**: all
  TEN sidebar rows in the reference's order (Credentials selected on
  the accent fill, the desk living underneath at "1000 rows / cache
  1000"), and the Backup page's groups — Vault Backup / VaultSeal —
  GitHub Mirror / Good Repos — Public Directory — with every button,
  checkbox, placeholder and status dot named back verbatim (the
  Dashboard group rides the page's own scroll, the reference's v32.2
  shape).
- **The battery at the close**: **2060 main-module tests race-clean
  over 35 packages** (2020 + 19 + 21 — arithmetic closed) + **28
  webui tests over 4 packages** (23 + 5 — arithmetic closed);
  gofmt/vet clean over both modules; the MAIN go.mod byte-unchanged
  all session; the known wailsjs mode flips restored via chmod after
  every wails build; no stray webui/webui binary committed (one
  appeared after a `go build .` in webui/ — removed before staging);
  the byte-level secret scan clean on every commit.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-74 row
  COMPLETE (the full record incl. the 32nd-wipe repair + the
  symlink-syscall deviation + the three slices' hashes).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-037
  entry.
- `README.md` — the SESSION-037 status entry.
- `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — two dated
  notes: the port's Check Queue door (the username guard kept, the
  Telethon guards dropped, Process New + Mark All as Read never
  render — Option B's accepted losses) + the dashboard engines'
  landing (the More-menu families stay T-76/T-77's).
- `docs/port-blueprint/module-specs/spec-gui-pages-b.md` — the T-74
  landing note (the page live, the Restore button honestly disabled
  with its landing note, the badge's token approximation).
- `webui/README.md` — settingsuib.go + the pages B in the map.
- `internal/app/doc.go` — the file map (settingsb.go + batchwire.go).
- `build/TOOLCHAIN.md` Am.4 — the symlink-syscall degradation record
  (the hardlink fallback; the recipe stays the symlink form).

## Blockers & decisions needed

- No new blockers. The authorized runtime set is UNCHANGED (the four;
  the MAIN go.mod byte-unchanged all session — verified at every
  commit). No new dependencies.
- Carried (no rush): the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement
  (twenty-eighth session carrying it); x/net/html keep-or-swap.
- **The sandbox's symlink block** (TOOLCHAIN.md Am.4) — a future wipe
  rebuild may need the hardlink fallback again; recorded with the
  evidence.
- **The Restore button's engine** (the safety-net restore + the
  export-zip door) — honestly disabled on the page; it rides the
  packaging phase (T-80's checklist) or a later binding; recorded in
  spec-gui-pages-b's dated note.

## Owner actions requested

- None urgent. The two standing calls (dissolution, vanilla-vs-vite)
  remain open; T-75 proceeds on the nested boundary either way.

## Handoff

- **SESSION-038 = T-75 (L): SYNC VIEW** per doc 13's row — the flow
  cards + the matrix + the differences modal (spec-gui-sync-views;
  the four-state rendering tests + the deferred-ledger UX tests).
  The v0.87.0 GRAND CONVERGENCE page (one Sync page, two desks: the
  four-element flow above the stage-by-stage gap matrix); the retry
  doors' engine half (review_retry / master_retry) is ALREADY LIVE in
  batchwire.go — the Sync page's UI wires to it.
- The state card v38 (`docs/NEXT-SESSION.md`) carries the full
  read-first list + the SESSION-038 prompt.
- Push state: slices (a) + (b) + (c) + the close committed and pushed
  (the ephemeral PAT at each push; never stored, never written).
