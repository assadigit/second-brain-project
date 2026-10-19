# SESSION-038 — 2026-10-20 — Phase 6: T-75 COMPLETE (the Sync view — the flow cards + the matrix + the differences modal; the syncreview placeholder seat finally filled) — 2211 main-battery tests over 36 packages + 33 webui tests over 4; the thirty-third full wipe repaired BEFORE work (the 6th webkit rebuild)

## Goal (from NEXT-SESSION.md at start)

SESSION-038 = **T-75 (L): SYNC VIEW** per doc 13's row — the flow cards
+ the matrix + the differences modal (spec-gui-sync-views): THE VISUAL
SYNC DISPLAY (🤖 bot → 📓 vault → ☁️ GitHub → 🌐 the site — read-only,
refreshed on demand, the click the consent for every network leg)
above THE DEFERRED LEDGER (the stage-by-stage gap matrix) + the
differences modal. The page joins the Settings window's SECTIONS at
the reference's own order (between Dashboard and Bot). ACCEPTANCE:
the four-state rendering tests + the deferred-ledger UX tests. If
capacity runs short: one slice done properly with proof beats two
half-done.

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-third full wipe repaired BEFORE work** (both repos,
  the toolchain, the venv, the module cache, AND the webkit tree all
  gone at boot; everything rebuilt from the committed recipes): both
  repos re-cloned with the ephemeral `GIT_CONFIG_COUNT` form (no PAT
  persisted anywhere; the work repo's tip verified at `3878a2a` =
  SESSION-037's close, the tree clean; the reference checked out at
  `98f4fc7` = PINNED-GITCURATOR-REV, byte-verified). go1.27.2
  downloaded from go.dev and sha-verified against BOTH the committed
  64-char pin (TOOLCHAIN.md Am.2) AND go.dev's published `?mode=json`
  listing — `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
  (the session prompt's own sha string arrived one character short —
  the SAME transcription corruption Am.2 documented; the committed
  pin is the authority). The venv (Python 3.12.15, stdlib-only), the
  wails CLI v2.16.0 (GOBIN install), and the rootless webkit tree
  (144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dangling
  dev links as HARDLINKS to the resolved system targets — **Am.4's
  fallback again: symlink(2) is STILL blocked**; the relink command
  itself had to resolve the link's readlink target, not the link
  name, on the first pass). The re-verification law executed before
  any new work: MAIN go.mod byte-verified (the authorized four @ go
  1.24.1; the nested webui @ 1.25.0 + wails + the parent replace),
  the full main battery + the webui battery + vet/gofmt green
  pre-work (2060 / 35 + 28 / 4).
  - **The display-pipeline corruption observed and compensated** (a
    session-level lesson, recorded here): this sandbox's tool-output
    rendering occasionally DROPS the two characters `[h` — the sha
    string's lost character in the session prompt, and
    `.settings-view[hidden]` rendering as `.settings-viewidden]` are
    the same family. Byte-level counts/hex dumps settled every
    suspect reading (the CSS was CORRECT — no fix needed; the
    sha pin was correct — the prompt's copy was the corrupted one).

- **T-75 slice (a) — commit `6eea74d`: internal/syncreview —
  core/sync_review.py PORTED, the placeholder seat T-71 recorded
  finally filled.** The pure-decision half: the four-state truth
  table (`CellState` count form + `StageCell` item form with the
  first-seen dedupe + `SplitGap`'s stable-ID split); `BannerVerdict`
  worst-first incl. v0.91.0 THE NAMED VERDICT (the stream · stage
  naming, the setup/measure split, the bare-cell v0.86.0 grammar
  byte-for-byte); `CellCaption` — THE MICRO-COPY LAW (the owner's
  own 720/710 example, every branch verbatim, junk degrading to
  dashes); the relative clock (`FormatAddedAgo`/`AddedDaysAgo` +
  `SortOldestFirst`); `MarkAllowed` (THE disabled-until law); the
  **DeferredLog over store.Open** (THE SQLITE LAW: busy_timeout
  30000 + WAL + synchronous NORMAL via the DSN, one connection, the
  per-ledger mutex; mark/reopen/resolve_seen/resolve_gone/
  reviewed_ids/rows/row_count; NEVER a delete — status changes
  only, pinned by a source scan); the three read-only doors:
  `FetchPublishedIndex` (the ONE network door — GET-only, the 1010
  UA worn, never raises, the hermetic request-level `httpDo` seam) +
  `PublishedSiteUrls`; `VaultSiteItems` (build_index's own filter
  laws via the new `internal/directorysync/export.go` wrappers —
  IterNotes/ReadText/NormHost/SharedHostsOf, reuse never
  re-invented); `VaultBackupGap` (the read-only git status
  --porcelain walk, the never-sealed every-note law, renames take
  the NEW path, .obsidian/.trash noise out — the read-only law
  pinned by a source scan); `VaultNoteCount` (the inventory's own
  totals). **101 tests** race-clean (the ported battery from
  tests/test_syncreview.py: the truth table, the banner precedence +
  the named verdict + the setup/measure split, the micro-copy
  captions, the ledger lifecycle/never-delete/stable-ids/search/
  two-writers-same-db concurrency, the chip grammar, the clock
  family, the oldest-first sort, the vault walks over fixture vaults
  + fixture git repos, the network door over the seam).

- **T-75 slice (b) — commit `04768e7`: internal/app/syncdisplay.go +
  syncview.go — the display model + the view half.**
  - syncdisplay.go: `ClassifyPending` (the router's own split —
    repos + websites + unrouted == total, gists are websites, junk
    spoken); `DetectScanVaults` (vault_pick_dialog's walk, finally
    ported — the three rows' connection truth + note counts + the
    Personal vault's owner-owned read-only law); `CollectLocalFacts`
    (the per-CONFIGURED-backup-repo VaultSeal.status rows — silence
    when unconfigured, the creation guard's law — + the Websites
    vault's public-ready count over BuildIndex); `FriendlyWhen`;
    `SyncDisplayModelOf` — the four cards in the owner's own flow
    order (bot → vault → GitHub → website), every line/hero/stat/
    hint/foot verbatim, THE §5 ARITHMETIC's every edge (V>P needs
    publishing, V<P "−N since the last publish" with U+2212,
    never-published is never, unreachable degrades to unknown,
    no-worker-URL speaks its door), the calm line only when
    waiting==0 AND any_measured, total junk tolerance.
  - syncview.go: the matrix tables (the six defs + the layout + the
    columns); `GatherSyncSnapshot` over injectable walks (the
    HERMETIC-SEAM law) — the flagship stage's setup/measure reasons
    + resolve_seen, THE QUEUE STAGES RIDE THE BOT PAGE'S OWN
    CLASSIFIED CHECK (the raw poll never lies into them — the
    reference's own law, kept), the backup stages' sealing-off/
    vault-not-set setup law + display numbers + resolve_gone, the
    items capped at 400 with the oldest-first sort;
    `SyncRecomputeFromLedger` (the instant repaint after modal
    changes); `FetchPublishedFacts` (the code-aware read-back door's
    four truths: ok / the Worker's 404 = never_published /
    not_configured / error spoken); the Settings doors:
    **`RefreshSyncView` — THE ONE CONSENT DOOR** (one payload, both
    desks: the facts + the model + the snapshot + the banner — the
    shared single gather the reference records as its v2 follow-up,
    taken now per the spec's own "one Refresh for both"),
    `SyncDifferences` + `SyncDiffHeading`, `SyncMarkIntentional`
    (the disabled-until re-check + the ONE inline confirmation line
    + the recompute), `SyncReopenItem`, `SyncDeferredRows` +
    `SyncLedgerCount`, `OpenNote`.
  - **50 tests** (syncdisplay_test.go's four-state rendering battery
    + syncview_test.go's gather/modal UX battery); the app package
    at 183 tests race-clean.

- **T-75 slice (c) — commit `478101c`: the shell.**
  - webui/syncuib.go — the FOURTH bound struct (SyncUI): Refresh
    (the one door), MarkIntentional + ReopenItem (one-struct returns
    for the bindings), StageDifferences + DeferredRows +
    LedgerCount, StageTables (the matrix's shape), CheckBotQueue
    (the queue stages' fix door — the same engine the Bot page
    rides, the queue state landed with it), SealNow/PublishNow,
    OpenNote. webui/settingsuib.go — the Bot page's own Check Queue
    now lands the session's queue state too (the reference's
    _refresh_sync_after_queue hook folded into the door).
  - frontend/settings.js — the Sync page at the reference's own
    sidebar order (between Dashboard and Bot, the 'refresh' glyph
    from the locked pack, the TAB_INFO hint verbatim): the teaching
    header, the Refresh row + the measured-at note, THE 2×2 COMPACT
    GRID (the four cards, the arrows with their meaning tooltips,
    the hero + stat rows + foot + hint per card, the calm line, the
    legend), THE DEFERRED LEDGER half (the banner with the tone
    washes, the 3×3 matrix with the four cell states — match ✓ /
    behind ⚠ / reviewed ⓘ / unknown dashed with the reason written
    on the cell, the legend, the last-measured line with the ledger
    count), THE DIFFERENCES MODAL (the tabs with counts, Select all,
    the 44px rows with the ago column + the deferred chip + Open
    note, the four exclusive reason chips + the note input + Mark as
    intentional with THE disabled-until law, the ONE inline
    confirmation line, the log tab with search + Reopen on Active
    rows + the nothing-is-ever-deleted note), the first-show door
    (ONE automatic refresh), the fix doors with the 4s seal
    re-measure arming. frontend/style.css — the sync_* roles over
    the locked semantic tokens.
  - **The first-render field-law bug, caught by the evidence leg and
    fixed in the same commit**: the JS read camelCase
    (`payload.measuredAt`, `model.dotTone`) against the Go tags'
    snake_case (`measured_at`, `dot_tone` — the port's own law: the
    JSON tags mirror the Python dict keys). The VLM read the stuck
    "Not measured yet" note; the fix landed; the re-run's VLM read
    "measured at 18:49:32 — the numbers are as fresh as this line".
  - syncuib_test.go — 5 shell tests: the one-payload-two-desks
    refresh over hermetic fakes, the queue-check fix door landing
    the state, the mark/reopen round-trip (the stage turns Reviewed
    behind the modal), the stage tables' shape, and the sidebar laws
    (the row order + the glyph + the verbatim texts).

- **Spec updates riding the landing (reality wins, same commits +
  the docs commit)**: doc 13's T-75 row COMPLETE (the full
  three-slice record); doc 12's SESSION-038 entry; README's
  SESSION-038 status; spec-gui-sync-views's dated note (the three
  port-level decisions: the shared single gather taken now, the
  queue-stages law, the snake_case field law); spec-core-sync-
  review's dated note (store.Open, the request-level seam, the
  export.go reuse); webui/README.md's map; internal/app/doc.go's
  file map; internal/syncreview/doc.go rewritten from the
  placeholder.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-75's acceptance = the four-state rendering tests + the
deferred-ledger UX tests — both green with proof:

- **The four-state rendering tests**: syncdisplay_test.go's battery —
  the display model's four states (ok / not_set / error / never)
  across all four cards, the §5 arithmetic's every edge (the owner's
  own example, the shrink case with U+2212, never-published,
  unreachable, not-configured), the hero laws, the calm line law,
  the flow order, the junk tolerance, the picker's walk, the local
  facts, the published door's four truths; PLUS the matrix cells'
  four states through the gather battery (behind/match/reviewed/
  unknown with the exact captions).
- **The deferred-ledger UX tests**: syncreview's 101-test battery
  (the truth table, the banner precedence + the named verdict + the
  setup/measure split, the micro-copy captions, the ledger
  lifecycle + never-delete + stable-ids + search + the
  two-writers-same-db concurrency, the chip grammar, the clock
  family, the vault/git/published helpers) + the app-level modal
  flows (the mark's confirm line + the recompute turning the cell
  Reviewed, the reopen's back-to-differences, the search, the
  heading's split).
- **The end-to-end evidence**: the wails build links clean over the
  6th-rebuilt webkit tree (29.7 s); the gate leg re-run GREEN —
  first frame **292/730 ms ≤ the 1.5 s gate** (two runs), the 10k
  flood through the FULL production transport at **588,235 lines/s,
  0 drops, 168 ms drain, 28 ms max handler**; the screenshots
  **VLM-verified**: the Sync row between Dashboard and Bot
  (selected on the accent fill), the four flow cards with the
  honest not-set grammar ("queue not measured" / "no vault
  connected" / "no backup repos" / "vault not set" — never a
  fabricated count), the live measured-at line, the dashed unknown
  matrix, the desk living underneath.
- **The battery at the close**: **2211 main-module tests race-clean
  over 36 packages** (2060 + 101 + 50 — arithmetic closed) + **33
  webui tests over 4 packages** (28 + 5); gofmt/vet clean over both
  modules; the MAIN go.mod byte-unchanged all session (verified at
  every commit); the wailsjs mode flips restored via chmod after
  the builds; no stray webui/webui binary (build/bin is ignored;
  the tree clean at the close).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-75 row
  COMPLETE (the full record incl. the 33rd-wipe repair + the three
  slices' hashes).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-038
  entry.
- `README.md` — the SESSION-038 status entry.
- `docs/port-blueprint/module-specs/spec-gui-sync-views.md` — the
  dated landing note (the three port-level decisions).
- `docs/port-blueprint/module-specs/spec-core-sync-review.md` — the
  dated landing note (the placeholder filled; the port decisions).
- `webui/README.md` — syncuib.go in the map.
- `internal/app/doc.go` — the file map (syncdisplay.go +
  syncview.go).
- `internal/syncreview/doc.go` — rewritten from the 5-line
  placeholder to the real package record.

## Blockers & decisions needed

- No new blockers. The authorized runtime set is UNCHANGED (the
  four; the MAIN go.mod byte-unchanged all session — verified at
  every commit). No new dependencies (syncreview rides store's
  modernc + the existing packages only).
- Carried (no rush): the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement
  (twenty-ninth session carrying it); x/net/html keep-or-swap.
- **The sandbox's symlink block persists** (TOOLCHAIN.md Am.4) —
  the 6th webkit rebuild needed the hardlink fallback again;
  recorded with the evidence.
- The deferred test debt on the books is unchanged (the T-74 list).

## Owner actions requested

- None urgent. The two standing calls (dissolution, vanilla-vs-vite)
  remain open; T-76 proceeds on the nested boundary either way.

## Handoff

- **SESSION-039 = T-76 (M): THE SCAN FLOW** per doc 13's row — the
  picker → the plan modal (the masters + the arm-guard) → the
  scorecard (spec-gui-dialogs + spec-gui-pages-a's vault_scan_ui;
  the _vault_scan_job ENGINE already lives in scanjob.go from T-71 —
  the UI legs are the work). ACCEPTANCE: workflow tests; selection
  semantics.
- The state card v39 (`docs/NEXT-SESSION.md`) carries the full
  read-first list + the SESSION-039 prompt.
- Push state: slices (a) + (b) + (c) + the docs committed and pushed
  (the ephemeral PAT at each push; never stored, never written).
