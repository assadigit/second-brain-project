# SESSION-039 — 2026-10-20 — Phase 6: T-76 COMPLETE (the Scan flow — the vault picker → the plan modal → the applied scorecard) — 2230 main-battery tests over 36 packages + 40 webui tests over 4; the thirty-fourth full wipe repaired BEFORE work (the 7th webkit rebuild)

## Goal (from NEXT-SESSION.md at start)

SESSION-039 = **T-76 (M): THE SCAN FLOW** per doc 13's row — the
picker → the plan modal (the masters + the arm-guard) → the scorecard
(spec-gui-dialogs + spec-gui-pages-a's vault_scan_ui): the three UI
legs over the engine that already lives (scanjob.go's confirm doors +
the scope picker + the picking arithmetic + the merged verdict). The
picker's first pure half (DetectScanVaults) already ported at T-75 —
build on it, never re-invent. ACCEPTANCE: workflow tests; selection
semantics. If capacity runs short: one slice done properly with proof
beats two half-done.

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-fourth full wipe repaired BEFORE work** (both repos,
  the toolchain, the venv, the module cache, AND the webkit tree all
  gone at boot; everything rebuilt from the committed recipes): both
  repos re-cloned with the ephemeral `GIT_CONFIG_COUNT` form (no PAT
  persisted anywhere; the work repo's tip verified at `d60e0c9` =
  SESSION-038's close, the tree clean; the reference checked out at
  `98f4fc7` = PINNED-GITCURATOR-REV, byte-verified). go1.27.2
  downloaded from go.dev and sha-verified against BOTH the committed
  64-char pin (TOOLCHAIN.md Am.2) AND go.dev's published `?mode=json`
  listing — `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
  (this session prompt's own sha copy arrived INTACT — 64 chars,
  byte-checked; the corruption family stays a live caution). The venv
  (Python 3.12.15, stdlib-only), the wails CLI v2.16.0 (GOBIN
  install), and the rootless webkit tree (144 debs → dpkg -x → the 42
  .pc prefix rewrites → the 26 dangling dev links as HARDLINKS to the
  resolved system targets — **Am.4's fallback again, the 7th webkit
  rebuild: symlink(2) blocked at BOOT this time, the plain `ln -s`
  refused with "Creating symbolic links is not allowed"**; the wails
  build + the gate leg the proof). The re-verification law executed
  before any new work: MAIN go.mod byte-verified (the authorized four
  @ go 1.24.1; the nested webui @ 1.25.0 + wails + the parent
  replace), the full main battery + the webui battery green pre-work
  — **exactly the recorded state: 2211 over 36 + 33 over 4,
  race-clean**.

- **T-76 slice (a) — commit `f745b18`: internal/app/scandialogs.go +
  the scanjob.go story extension — the three dialogs' pure engines.**
  - The story extension: `ScanStory` (the reference's per-vault story
    dict) — the planned counts **spoken in PICKED units once a
    confirmed selection rides** (worker_jobs.py:1130-1136 — the
    scorecard's 'K of N' line never reads an unticked row as a
    failure), the landed counts from the ApplyReport, the failed
    moves BY NAME; `ScanResult` grew the out-dict shape (Story = the
    Websites vault's story or the PROMOTED repos story when repos
    rides alone; Repos = the sub-story; VaultsRan; Hygiene = the
    per-vault heal counts; the vault/kind/label seats) — the
    never-read `AppliedCounts` field (dead since T-71) replaced by
    the real shape. The GUI confirm door's ask line landed
    byte-for-byte (worker_jobs.py:950-954) over the configured
    `scan_confirm_timeout_s`; `Commands.BatchBusy` the scan guard's
    honest reader.
  - `PickDisplayModelOf` (pick_display_model: the status lines in
    the owner's voice, the Personal read-only why, the hint law) +
    `PickScanButtonLabel` (the live arithmetic) + `PickedScanKinds`
    (the safe-cancel scope law: nil/empty → nil — nothing scanned).
  - `PlanDisplayModelOf` (plan_display_model): the four sections in
    dialog order with the reference's own washes + tones +
    headings-with-totals; every row carries its stable id (the
    selection key), its one-line grammar, and its full tooltip story
    — the deletion door words (your mark / you moved it to Trash /
    the raw door), the consolidation's one-note-per-site grammar with
    the recoverable-by-hand tooltip, **THE LANDABLE MOVE's
    "(as C_v1.md)"**, the inventory block + the trash/locked/kept/
    low-confidence lines, the vault-label law incl. the repos
    KeptHandwrittenR seat, the MAX_ROWS 500 freeze guard with the
    counted-not-rendered tail.
  - **`PlanPickState`** — the picking door's selection semantics as a
    pure state machine (main-battery-testable, the house law): born
    checked except the LLM's own low-confidence calls, the master
    sweep (any unticked → all ticked; all ticked → none), the
    three-state master mirror, `PickedTotals`, the live Apply label
    "🗂️ Apply plan (12 of 17)", **THE ARM-GUARD'S ENABLE LAW**
    (armed AND (picked>0 OR nothing to pick); Keep gated by the arm
    alone), the four-key `Selection`.
  - `ResultDisplayModelOf` (result_display_model): the LANDED counts
    never the plan's promises, the honest 'K of N' warning tones +
    the shortfall tips, the named failed moves (cap 3 + the …), the
    aggregate 🧹 row with the linked-ghosts/app-owned detail, the
    three-way heading law (healed-only / nothing-landed / applied),
    the nothing-line + the recoverability footnote, the per-vault
    sections incl. the repos sub-story's own section + the
    github-only "projects vault" heal section. **The reference's
    repos-only double-section quirk PRESERVED** — the out dict
    carries `out['repos']` unconditionally and the finished routing
    hands the raw dict to the scorecard, so a repos-only scan
    renders the promoted section AND the sub-story section BOTH:
    verified by RUNNING the reference's `result_display_model` over
    the repos-only shape at 98f4fc7 (the venv's own purpose), pinned
    by the port's test with the probe citation.
  - **19 tests** (scandialogs_test.go — the spec's own workflow
    list). Riding along: the gofmt hygiene fix on syncview.go's
    struct alignment (a T-75 leftover).

- **T-76 slice (b) — commit `98823ad`: the shell.**
  - webui/scanui.go — the FIFTH bound struct: `PickVaults` (the
    picker's data door); **`StartScan` — THE CTA's whole flow**
    (vault_scan_ui.py's `_start_vault_scan` + `_vault_scan_finished`,
    one door): the guards verbatim (the batch-busy check over
    BatchBusy + the one-scan flag — both refusal lines byte-for-byte),
    the picker's safe-cancel quiet line, the start line with the
    chosen-vault names, the blocking `ScanRun` over the chosen scope
    (the desk events flow on their own transport while the JS promise
    waits — the TestWorker law's transport twin), the finished
    routing's lines verbatim (clean+healed ⇒ the scorecard, confirmed
    ⇒ the complete line + the repos sub-story line + the scorecard,
    error ⇒ 💥) + the scan:finished event; `ScanAnswer` the held
    gate's answer door (first answer wins); the two evidence doors
    (`ProbeAskSample` / `ProbeScorecardSample` — canned fixture
    shapes, no owner data).
  - webui/shell.go — the AskScanConfirm wiring (the plan becomes the
    display model + crosses as the scan:ask event with the answering
    window; the bus's own 300 s window the belt under the modal's
    config countdown); main.go binds the fifth struct.
  - frontend/scanflow.js — the three modals' renderers: the picker
    (nothing pre-checked; the read-only Personal row; the
    live-arithmetic Scan button; Cancel/Esc = the safe cancel through
    the one door); **THE PLAN MODAL** (the four wash panels over the
    locked plan-* tokens, the master boxes with the indeterminate
    partial state, the low-confidence pre-unchecked rows, **THE
    FRESH-CLICK DOOR** — the 350 ms capture-phase input swallow +
    both verdict buttons born disabled + the armed-refusal spine, the
    char-budget elideMiddle twin with the full story on the tooltip,
    the m:ss countdown, the four-key selection riding ONLY confirmed,
    **THE PINNED VERDICT ROW** — the body scrolls, the clock + the
    doors never leave the screen); **THE SCORECARD** (the per-vault
    sections with the changed counts, the K-of-N warning values, the
    tooltip whys, the footnote, the 12 s self-close ticking line,
    OK); **the Scan CTA's own machine** (the reference's scan_btn
    tooltip verbatim — ui.py wires it in vault_scan_ui.py, NOT
    hero.py; the SYNC/Fast/Test buttons stay T-77's hero machine).
  - **7 shell tests** (scanui_test.go): the picker data door, the
    quiet cancel + the one-scan law over a held gate, the batch-busy
    refusal, **THE REAL FLOW end-to-end over the bus gate** (the
    marked note → the ask → the answer → the apply → the scorecard +
    every desk line), the held-ask first-answer-wins, the probe
    samples' shapes, the frontend laws read where they render.

- **The docs — commit `4626a39`**: doc 13's T-76 row COMPLETE; doc
  12's SESSION-039 entry; spec-gui-dialogs's dated landing note (the
  six port-level decisions); spec-gui-pages-a's vault_scan_ui dated
  note; README's SESSION-039 status; webui/README.md + internal/app/
  doc.go maps.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-76's acceptance = workflow tests + selection semantics —
both green with proof:

- **The workflow tests**: the 19-test engine battery (the picker's
  display shape + status lines + arithmetic + safe-cancel scope; the
  plan's section grammars + door words + freeze guard; the scorecard's
  K-of-N + named failures + headings + per-vault sections + hygiene
  row) + the story laws (the picked-counts replacement — an unticked
  deletion row never counts as a failure and never touches the note;
  the repos-only promotion; the GUI door's ask line with the
  configured window; the timeout verdict) + the 7 shell tests incl.
  THE REAL FLOW end-to-end over the bus gate (StartScan → the guards
  → the start line → the job → the ask held → the answer → the apply
  → the finished routing → the scorecard — every desk line asserted).
- **The selection semantics**: PlanPickState's own battery (the born
  law + the low-confidence pre-uncheck, the master sweep's both
  directions + the three-state mirror, the arm-guard's enable law —
  an unarmed Apply stays dead whatever the arithmetic; nothing-picked
  disables; an empty plan stays armed-eligible; Keep gated by the arm
  alone, the four-key selection with the empty-list law) + the
  frontend twin's laws pinned in scanui_test.go + the bus gate's own
  shape law (only a confirmed verdict carries a selection — the
  pre-existing events battery).
- **The end-to-end evidence**: the wails build links clean over the
  7th-rebuilt webkit tree; the gate legs GREEN (first frame
  **339/283/280 ms ≤ the 1.5 s gate** across the runs; the 10k flood
  through the FULL production transport, 0 drops); the screenshots
  **VLM-verified**: the plan modal (the four wash sections + the
  masters + **the low-confidence row PRE-UNCHECKED** + the landable
  "(as C_v1.md)" row + the pinned verdict row speaking **"Apply plan
  (4 of 5)"** — the live arithmetic — + the ticking "Answering
  window: 4:55" line), the picker (the three rows, NOTHING
  pre-checked, the Scan button honestly disabled with nothing ticked,
  the no-scan-able-vault hint over the unconfigured desk), the
  scorecard (both vault sections with their change counts, the "3 of
  5" + "2 of 4" warning values, the recoverability footnote, the
  self-close line ticking at 8s, the OK button).
- **The battery at the close**: **2230 main-module tests race-clean
  over 36 packages** (2211 + 19 — arithmetic closed) + **40 webui
  tests over 4 packages** (33 + 7); gofmt/vet clean over both
  modules; the MAIN go.mod byte-unchanged all session (verified at
  every commit); no mode flips committed (models.ts's flip matched
  HEAD's committed 755 baseline; the new ScanUI bindings ride the
  repo's established 755 convention); no stray webui/webui binary;
  the tree clean at the close.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-76 row
  COMPLETE (the full record incl. the 34th-wipe repair + the two
  slices' hashes).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-039
  entry.
- `docs/port-blueprint/module-specs/spec-gui-dialogs.md` — the dated
  landing note: the six port-level decisions (the elide-middle
  char-budget twin; the repos-only double-section quirk preserved
  with the probe evidence; the story extension + the AppliedCounts
  replacement; the Scan CTA's own machine; the answering windows'
  belt note — a configured window LONGER than the bus's 300 s would
  be cut by the belt, recorded as a known divergence for the owner;
  the evidence doors) + the ConnectionTestDialog row staying T-77's.
- `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the
  vault_scan_ui dated landing note (the mixin's flow as ONE binding
  door; the quiet line's owner; the scorecard-skip law riding T-77's
  window lifecycle).
- `README.md` — the SESSION-039 status entry.
- `webui/README.md` + `internal/app/doc.go` — the new files in the
  maps.

## Blockers & decisions needed

- No new blockers. The authorized runtime set is UNCHANGED (the four;
  the MAIN go.mod byte-unchanged all session — verified at every
  commit). No new dependencies.
- Carried (no rush): the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (thirtieth
  session carrying it); x/net/html keep-or-swap.
- **The answering-window belt divergence, recorded** (spec-gui-dialogs
  note #5): the modal's countdown ticks the configured
  `scan_confirm_timeout_s`, but the bus's ask gate pins the reference
  constant 300 s — a configured window LONGER than 300 s would be cut
  by the bus's belt (the reference's worker waits the same config
  value). The owner's call if he ever configures >300 s; a per-ask
  window on the bus would be the fix (an events.go change, out of
  T-76's scope).
- **The T-71 merged-verdict question, surfaced with evidence**
  (NOT acted on — out of T-76's scope, recorded for the owner/next
  session): the reference's merged verdict (worker_jobs.py:1262-1274)
  is "a confirmed vault anywhere → confirmed; else clean when every
  scanned vault is clean; else the top-level verdict" — the Go
  `mergeScanVerdicts` (T-71) is worst-word-wins (declined+confirmed →
  declined, where the reference answers confirmed). Only bites on a
  BOTH-vault scan where one confirms and the other declines; pinned
  by T-71's own tests as the port's current law. Hash-level evidence
  recorded; needs an owner decision (fix the merge to the reference's
  law, or record the divergence as accepted).
- **The sandbox's symlink block persists** (TOOLCHAIN.md Am.4) — the
  7th webkit rebuild needed the hardlink fallback again, now blocked
  at BOOT (not mid-session); recorded with the evidence.
- The deferred test debt on the books is unchanged (the T-74 list).

## Owner actions requested

- None urgent. The standing calls (dissolution, vanilla-vs-vite)
  remain open; T-77 proceeds on the nested boundary either way.
- Two recorded divergences await the owner's eye when convenient: the
  answering-window belt (only if a >300 s window is ever configured)
  and the T-71 merged-verdict question above.

## Handoff

- **SESSION-040 = T-77 (L): the main window chrome** per doc 13's row
  — the hero button machine (the sync → fetching → process → running
  state machine on ONE button), the progress row's live wiring, the
  More menu, the sentinel labels, the dialogs polish, the sounds
  (spec-gui-ui + spec-gui-window). ACCEPTANCE: hero state machine
  tests; startup cadence test.
- The state card v40 (`docs/NEXT-SESSION.md`) carries the full
  read-first list + the SESSION-040 prompt.
- Push state: slices (a) + (b) + the docs committed and pushed (the
  ephemeral PAT at each push; never stored, never written).
