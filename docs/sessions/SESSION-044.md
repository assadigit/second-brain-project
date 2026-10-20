# SESSION-044 — 2026-10-20 — Phase 6 (Path B): T-79c COMPLETE (the three bot-queue tool doors) — the reference's own guards carried exactly (no busy guard where the reference took only the TELEGRAM lock); 2322 main-battery tests over 36 packages + 67 webui over 4; the fortieth full wipe repaired BEFORE work

## Goal (from NEXT-SESSION.md at start)

T-79c (M): the bot-queue tool dialogs as More-menu doors — ✅
Verify All Processed + 📋 Export All Links + 🔄 Retry Failed, the
reference's own flows (bot_queue.py:929/1026 +
dashboard.py:770) over the ported queue engines. THE GUARD LAW:
carry the reference's own guards (the telegram lock where the
reference takes it, the busy guard where a batch runs) exactly as
the reference does. Follow T-79a/T-79b's landed shape (the engines
over hermetic seams + the HeroUI wrappers + the heroflow.js flip +
the line grammar pinned hermetically + the evidence probe's click
leg + THE FIELD-TAG law on every genuinely new payload).

## What was done (task IDs, files, commits — cite hashes)

- **The fortieth full wipe repaired BEFORE work**: the ENTIRE
  environment was gone (both repos, the toolchain, the venv, the
  module cache, the wails CLI, AND the webkit tree — the working
  directory held only the stock sandbox scaffold). Everything
  rebuilt from the committed recipes: both repos re-cloned with the
  ephemeral GIT_CONFIG_COUNT PAT form (nothing persisted, both
  configs verified clean); the work repo at `d400a5f` =
  SESSION-043's close commit, up to date with origin/main; the
  reference at `98f4fc71cfb862cee7225d864636c08062724348` =
  testdiff/PINNED-GITCURATOR-REV exactly. go1.27.2 downloaded from
  go.dev over HTTPS; the local sha256 matches TOOLCHAIN.md Am.2's
  committed 64-char pin byte-for-byte (go.dev's ?mode=json has
  rotated to 1.27.3/1.26.10 — 1.27.2 no longer listed; the
  committed pin is the authority and it matches). The venv
  (Python 3.12.15, the plain form — this boot HAS ensurepip). The
  wails CLI v2.16.0. The TWELFTH webkit-tree rebuild: 144 debs /
  77.7 MB → dpkg -x → the 42 .pc prefix rewrites → the 26 dangling
  dev links as Am.4 hardlinks to the resolved system targets
  (symlink(2) BLOCKED this boot — the standalone `ln -s` test
  refused with "Creating symbolic links is not allowed", so the
  Am.4 fallback applied; zero dangling links remain; the wails
  build's clean CGO link over the tree is the proof). The pre-work
  re-verification from scratch: **exactly the recorded state — 2312
  main-battery tests over 36 packages PLAIN (0 fail; the -v count
  = 2312) + the full -race pass green + 63 webui over 4 (plain +
  race) + gofmt/vet clean over both modules + the tree clean vs
  d400a5f + the MAIN go.mod at the authorized pins.**
- **T-79c implemented** — `internal/app/botdoors.go` (the engines,
  the pertests/linkdoors pattern — Settings methods): the three
  doors over ONE new hermetic `SettingsDeps` seam
  (`BotQueueFetch` — the same Worker /api/pending honest poll
  composition the queue command rides; App.New wires BOTH to
  `a.queueFetch`, production-wired, faked at construction in every
  test). **Verify All Processed**: the username + vault guards, the
  poll, the vault-index rebuild ("📚 Vault index: N notes
  indexed"), the ONE-connection cache block (the dead set + the
  processed set + THE SETTLEMENT with its 🤝 line + the settled
  set), the per-URL ladder in the reference's exact order (in-vault
  → decommissioned → cache-processed ["✅ Found in cache
  (processed before): {url}"] → fuzzy [the reference's own regex,
  the .git-suffix strip, the lowercase pair compare over
  VaultIndex.MirrorIndex] → settled → missing), the "⏳ N missing
  GitHub link(s):" numbered lines, the byte-parity VERIFICATION
  REPORT (the count lines' exact spacing byte-verified against the
  reference's own source before writing), the permanent
  `_verification_report_{stamp}.md` (the summary + detailed tables),
  the missing/accounted summary line, the resolve payload, and the
  whole-body crash guard (the reference's v26 Fix 6 law — "never let
  Verify All take down the app"). **Export All Links**: the guards,
  "📋 Exporting ALL links from bot (no limit)…", the poll, the
  routed lists + **Σ forward_count as the message count** (the
  Worker ledger's own count of the bot's link-bearing deliveries),
  the byte-parity `_bot_links_export_{stamp}.md`, the two verdict
  lines, the 📋 Links Exported box. **Retry Failed**: the scan half
  (the vault-error box, the honest read-failure line, the "✓ No
  failed repos to retry." + Retry Queue Empty box, the settled-skip
  layer with its 🤝 line, the re-empty verdict, the >10 confirm via
  ConfirmLargeBatch over "the retry queue") + the resolve engines
  (Mark as Processed / Decommission / Mark ALL — every write settles
  the ledger too, v0.63.3's law; the all-resolved arithmetic over
  the dialog's CURRENT list).
- **THE GUARD LAW (the reference's own, carried exactly)**:
  export/verify take **NO busy guard** — the reference's TELEGRAM
  lock was their only guard (bot_queue.py:932/:1043), and it is
  Option B's accepted loss (the same law that dropped it from the
  Check Queue door); the doors run alongside a batch, the
  reference's own concurrency semantics — pinned by
  `TestBotDoorNoBusyGuard`. The retry door's batch carries
  SyncRunBatch's own one-at-a-time refusal (the reference's
  _start_worker had no busy check for 'direct' mode — the port's
  honest refusal is SESSION-035's already-recorded improvement; the
  scan itself reads freely, exactly the reference).
- **The shell** — `webui/heroui.go`: seven More doors.
  `MoreRetryFailed(confirmed bool)` is the two-phase confirm (the
  shell holds the scanned urls on its own `retryFailedURLs` seat,
  read+written under `u.mu` as ONE unit — THE SENTINEL-SEAT RACE
  LAW; a ≤10 list falls straight through to the batch, the
  reference's mid-flow _confirm_batch shape; a stale Continue with
  no held list re-scans, never a blind run); the post-confirm leg
  speaks "🔄 Retrying N previously-failed repos…", **CLEARS
  u.pendingRepos/u.pendingWebsites** (the reference's
  `_bot_queue_urls`/`_bot_queue_pending_websites` law — the exact
  twins, so PROCESS can never pick a retry batch up), mirrors the
  hero machine, and runs the DIRECT batch (bot_source=False).
  `MoreResolveProcess` is _process_missing_urls: the 🚀 line + the
  bot-source batch + the pending-seat assignment.
  `frontend/heroflow.js`: the three MORE_ACTIONS rows LIVE with
  door tips + the dispatch (the verify report's long-body box with a
  **Resolve Missing Links** primary when missing — the resolve
  dialog opens from the box's own button, the one-overlay-at-a-time
  law; the retry two-phase box; `openBotResolveDialog` — the
  reference's own words + button hierarchy, ONE filled primary +
  outlined secondaries + ONE filled danger, the buttons pinned
  OUTSIDE the scrolling body, click-to-select rows, the silent
  no-selection return) + the probe's **'export' click pick**.
  `frontend/style.css`: the resolve-list family + the
  `.btn-primary`/`.btn-danger` fills (an old gap closed: the kinds
  were in use since T-73/T-77 without their design-system fills —
  every dialog's primary/danger buttons gain the reference's look in
  this commit, recorded in the evidence doc). `shell.go`'s probe
  comment extended.
- **THE FIELD-TAG LAW honored**: BotVerifyResult / BotRetryScan /
  BotResolveAction tagged + pinned by
  `TestBotDoorPayloadFieldTags`; the shell adds no new payload
  shapes (the doors return the app types).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-79c's acceptance = "the three dialogs' verdicts stream +
the queue state refreshes" — **all GREEN**:

- **The verdicts stream**: every door's line grammar + level pinned
  by the engine cases (10 new tests in
  `internal/app/botdoors_test.go`): the username/vault refusals,
  both fetch-error states, the export save-failure, the export file
  + the report file **byte-checked against the reference's own
  formats** (the exact spacings byte-verified against the
  reference's source first), the classification ladder (in-vault /
  cache-processed / the .git+case-drift fuzzy case / decommissioned
  / settled / missing), THE SETTLEMENT's first-run line + the
  primed-guard silence, the >10 confirm's exact box (the
  10-boundary included), the resolve engines (incl. the repo_id-0
  replaces pin — the reference's own schema, the cross-binary law),
  the no-busy-guard pin.
- **The queue state refreshes**: the shell cases (4 new tests)
  pin the retry door's pending-lists clearing + the /direct source
  + the held-seat law + the two-phase flow + the resolve wrappers
  (the /bot source + the pending-seat assignment).
- **The battery at the close: 2322 main-battery tests race-clean
  over 36 packages (2312 + 10) + 67 webui tests over 4 (63 + 4)**;
  the touched packages (internal/app + webui) 3× race-clean
  post-change (the ritual); gofmt/vet clean over both modules; the
  MAIN go.mod + go.sum + the nested module's files byte-unchanged
  (0 diffs, verified).
- **The `wails build`** over the 12th-rebuilt webkit tree: links
  clean (29.5 s); no mode flips (`git diff --raw`: the wailsjs
  files stay at their 755 baseline; the content-only diffs = the
  regenerated bindings carrying the seven doors + the three payload
  classes); no stray binaries (build/bin stays gitignored).
- **The gate legs GREEN**: first frame 892 / 307 / 269 ms (≤ the
  1.5 s gate); the 10k flood through the production transport with
  0 drops (555k lines/s accepted).
- **The screenshots VLM-verified** (`t-79c-more-menu.png` +
  `t-79c-export-click.png`): the three doors LIVE in the menu. The
  first pass called Retry Failed "grayed" — the 042/043 emoji-led
  misread twin, settled per the law: the byte-level source check
  (the live keys + the uniform render path, pinned by
  `TestHeroflowBotQueueDoorsLive`) + a focused cropped re-pass
  (all three doors dark/LIVE, the bottom tail washed). **The click
  leg** (`SB_POC_OPEN_MORE=export`): the probe clicks 📋 Export All
  Links LIVE and the desk speaks the door's own refusal line
  end-to-end — "❌ Please enter the bot username first." **The
  first click-leg run caught a REAL gap** (the seven doors missing
  from heroflow.js's binding IMPORT list — a silent ReferenceError,
  the desk never spoke; the settled-by-the-crop investigation found
  it): fixed (the import list) + **pinned by a new law test** (the
  import-statement pin in TestHeroflowBotQueueDoorsLive) — the
  evidence loop paid again.
- The evidence archived at `docs/poc-artifacts/t-79c-bot-queue-doors.md`
  (+ the two screenshots + the results JSON), with the TEN recorded
  divergences (Σ forward_count as the message count; the
  save-before-verify rides the save-on-change law; the
  mark-all-read question dropped [read-acknowledge = Option B's
  accepted loss]; the report rides the More door's themed box; the
  resolve dialog opens through the box's own button; the
  one-connection cache blocks; the sorted fuzzy iteration; the
  plain-open file writes [the atomic wrapper's MkdirAll would paper
  over the bad-vault-path failure the reference exposes]; repo_id 0
  replaces; the btn-primary/btn-danger fills).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-79c
  row COMPLETE with the full record.
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-044
  entry.
- `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — TWO
  dated notes: the bot-queue section's (T-79c's landing + the guard
  law + the seams + the divergences' pointer) + the dashboard
  section's completion half (the failed-repos retry launcher landed;
  the reconciliation-links retry remains the Sync view's door).
- `README.md` — the SESSION-044 status block.
- `webui/README.md` — the heroflow.js + heroui.go map entries
  extended with T-79c's doors.
- `docs/poc-artifacts/t-79c-bot-queue-doors.md` + the two
  screenshots + `t-79c-gate-results.json` — the archived evidence.
- `docs/NEXT-SESSION.md` — rewritten for SESSION-045 (the state
  card v45).

## Blockers & decisions needed

- **No new blockers.** T-79c closed clean.
- Carried, unchanged: the dissolution call (ADR-0002 Am.5);
  the vanilla-vs-vite call; the go-directive acknowledgement (now
  the thirty-fifth session carrying it); x/net/html keep-or-swap;
  the answering-window belt + the T-71 merged-verdict question
  (SESSION-039's record); the deferred test debt
  (test_noteredo's un-ported selections; the note_state
  surgery-path CRLF round-trip differential).
- **The go.dev listing rotation** (a small environmental note for
  the next sessions): `?mode=json` now lists only go1.27.3 +
  go1.26.10 — the Am.2 verification procedure now compares the
  locally-computed hash against the COMMITTED pin (which itself was
  verified against go.dev's published listing at SESSION-032); the
  download stays HTTPS from go.dev.

## Owner actions requested

- None urgent. When convenient (no rush): the standing calls —
  dissolution, vanilla-vs-vite.

## Handoff

- **SESSION-045 = T-79d: the Chrome doors** (🖐 Hand-deliver walled
  links + 🖐 Scrape hand rows via Chrome + 🤖 Chrome tab-retry
  failed links + 🪄 Attach to my Chrome — the engines already live
  in internal/chrometabs; the reference's homes: the More-menu
  action list + handdelivery/chrometabs flows). Then the last tail
  row: T-79e (the decommission picker) — then Phase 6 closes and
  Phase 7 (T-80: the deterministic zip builder …) opens.
- NEXT-SESSION.md rewritten: yes (the state card v45, the
  SESSION-045 prompt inside).
- Push state: all work committed + pushed at the close (the
  ephemeral PAT; never stored, never written).
