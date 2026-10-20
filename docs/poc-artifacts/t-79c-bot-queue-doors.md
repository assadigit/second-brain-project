# T-79c — the bot-queue tool doors (the More menu's third family)

> SESSION-044, 2026-10-20. The three More-menu rows the reference's
> Bot tab owns — ✅ **Verify All Processed** + 📋 **Export All Links** +
> 🔄 **Retry Failed** — landed LIVE over the ported queue engines:
> `internal/app/botdoors.go` (the engines, T-79a/T-79b's
> Settings-method pattern) + `webui/heroui.go`'s seven wrappers +
> `frontend/heroflow.js`'s dispatch + the manual-resolve dialog. The
> pinned reference at `98f4fc7`: `bot_queue.py:929` (export_all_bot_links),
> `bot_queue.py:1026` (verify_all_bot_links, the per-URL ladder + the
> report + the resolve dialog at :1368-1496 + _process_missing_urls at
> :1355), `dashboard.py:770` (retry_failed_repos).

## What landed

- **✅ Verify All Processed** (`Settings.VerifyAllBotLinks` →
  `BotVerifyResult{title, body, missing}`): the username + vault
  guards, the Worker pending poll (the new `BotQueueFetch` seam —
  the SAME `a.queueFetch` composition the queue command rides), the
  vault-index rebuild ("📚 Vault index: N notes indexed"), the
  ONE-connection cache block (the dead set + the processed set +
  THE SETTLEMENT with its 🤝 line + the settled set), the per-URL
  ladder in the reference's exact order (in-vault → decommissioned →
  cache-processed with its "✅ Found in cache (processed before)"
  line → fuzzy with its "✅ Fuzzy match: {url} → {note}" line →
  settled → missing), the "⏳ N missing GitHub link(s):" numbered
  lines, the VERIFICATION REPORT (byte-parity text — the count lines'
  exact spacing pinned), the permanent
  `_verification_report_{stamp}.md` under the vault (the summary +
  detailed tables, byte-parity), the summary line (the missing
  variant / the accounted variant), and the manual-resolve payload.
  Never takes the app down (the reference's whole-body try/except —
  the v26 Fix 6 law).
- **📋 Export All Links** (`Settings.ExportAllBotLinks` → title+body):
  the guards, "📋 Exporting ALL links from bot (no limit)…", the
  poll, the routed lists + **Σ forward_count as the message count**
  (the Worker ledger's own count of the bot's link-bearing
  deliveries), the `_bot_links_export_{stamp}.md` file (byte-parity
  content), "📋 Export saved: {path}" + the "📊 Total messages: N |
  GitHub: G | Non-GitHub: NG | Duplicates: D" line, and the 📋 Links
  Exported box (byte-parity body).
- **🔄 Retry Failed** (`Settings.RetryFailedScan` → `BotRetryScan{box,
  confirm, urls}` + the shell's two-phase `MoreRetryFailed(confirmed)`):
  the vault-error box, the retry-queue read (the honest
  read-failure line), the "✓ No failed repos to retry." + Retry Queue
  Empty box, the settled-skip layer (the 🤝 left-the-queue line — the
  ledger's second guard), the >10 confirm (ConfirmLargeBatch over
  "the retry queue"), then the shell's post-confirm leg: the "🔄
  Retrying N previously-failed repos…" line, the pending-lists
  CLEARING (the reference's `_bot_queue_urls` /
  `_bot_queue_pending_websites` law — the exact twins, so PROCESS can
  never pick a retry batch up), and the DIRECT batch
  (bot_source=False). A stale Continue with no held list re-scans —
  never a blind run.
- **The manual-resolve engines** (`BotResolveMarkProcessed` /
  `BotResolveDecommission` / `BotResolveMarkAll` → `BotResolveAction`):
  the actions take the dialog's CURRENT list (the reference removes
  the row before the empty check), every write settles the ledger
  too (v0.63.3's "the owner's verdict settles" law), and the
  all-resolved arithmetic speaks "✅ All missing links resolved!" at
  the reference's own moment. `MoreResolveProcess` is Process Selected
  (_process_missing_urls): the "🚀 Processing N missing URL(s)…" line
  + the bot-source batch + the pending seat's assignment.
- **The frontend** (`frontend/heroflow.js` + `frontend/style.css`):
  the three rows LIVE with door tips; the verify report rides the
  long-body box with a **Resolve Missing Links** primary when
  missing (the one-overlay-at-a-time law — the resolve dialog opens
  from the box's own button); the retry door's two-phase box
  (Continue → the confirmed pass); `openBotResolveDialog` — the
  reference's own words (Manual Resolve — Missing Links / the ⏳
  header / Process Selected / Mark as Processed / Decommission /
  Mark ALL as Processed / Close; ONE filled primary + outlined
  secondaries + ONE filled danger, the buttons pinned OUTSIDE the
  scrolling body; click-to-select rows; the actions no-op with
  nothing selected, the reference's silent return); the probe's
  **'export' click pick**. The T-79c section in `style.css` also
  closes an old gap: `.btn-primary` / `.btn-danger` (the btn_kind
  law's filled variants — the kinds were in use since T-73/T-77
  without their fills; every dialog's primary/danger buttons gain
  the reference's look).

## THE GUARD LAW (the reference's own, carried exactly)

The reference guards export/verify with the **TELEGRAM lock only**
(bot_queue.py:932/:1043) — no batch-busy guard; they run happily
alongside a batch. Under Option B the Telegram lock is the accepted
loss (the same law that dropped it from the Check Queue door), so the
port's doors carry **no busy guard** — pinned by
`TestBotDoorNoBusyGuard` (all three doors run under a held batch
lock; no guard line ever speaks). The retry door's batch carries
`SyncRunBatch`'s own one-at-a-time refusal (the reference's
`_start_worker` had NO busy check for 'direct' mode — the port's
honest refusal is SESSION-035's already-recorded improvement).

## The recorded divergences

1. **The data source is the Worker pending poll** (ADR-0004 Am.3):
   the Telethon ALL-messages fetch is gone. The export's "Total
   messages in bot" renders **Σ forward_count** over the pending
   rows — the Worker's own count of the bot's link-bearing
   deliveries (the Telethon total counted every chat message;
   text-only messages never enter the URL-keyed ledger).
2. **The save-before-verify rides the save-on-change law**: the
   reference's `self.save_config()` before the fetch has no port leg
   (the config IS the live field; the Bot page's group save would
   also rewrite the token).
3. **The all-accounted verify's "mark all as read?" question is
   DROPPED** — read-acknowledge is Option B's accepted loss (the
   same law that dropped Mark All as Read from the Bot page; the
   mirror push is the ack). The success summary line is the whole
   verdict.
4. **The report renders in the More door's themed box** (the
   long-body scroll law) — the reference's queue_display is a Bot-tab
   page surface the More door does not own; the fetching placeholder
   ("⏳ Fetching ALL links from bot…") rides the box's completion
   rendering.
5. **The resolve dialog opens through the report box's Resolve
   button** (the one-overlay-at-a-time law) — the reference opens it
   automatically over the page-surface report.
6. **The cache blocks open ONE connection per door** (the reference
   opens back-to-back connections; the T-79b recategorize precedent).
7. **The fuzzy match iterates MirrorIndex's sorted map** (the
   reference rides its dict's insertion order — the recorded
   sorted-iteration divergence class; the first matching pair wins
   either way, and a pair collision under two spellings of the same
   repo is not an observable the reference pins).
8. **The export/report files write with the plain single-open**
   (`os.WriteFile` — the reference's `open('w')` twin): the atomic
   wrapper's MkdirAll would paper over the bad-vault-path failure
   the reference exposes at this door ("❌ Failed to save export: …").
9. **repo_id 0 replaces** in processed_repos (the reference's own
   schema, `cache_db.py:151`): two manual marks leave only the LAST
   processed row — pinned as the cross-binary law's own truth; the
   SETTLEMENT is the durable record.
10. **The btn-primary/btn-danger fills** (style.css) — the T-73-era
    kinds gain their design-system fills at T-79c's dialog (a
    cosmetic catch-up, recorded here because the earlier boxes'
    primary buttons change look in the same commit).

## The hermetic pins

- `internal/app/botdoors_test.go` — 10 engine cases: every door's
  line grammar + level + every early return (the username/vault
  refusals, both fetch-error states, the export save-failure), the
  export file + the report file byte-checked against the
  reference's own formats, the classification ladder (in-vault /
  cache-processed / fuzzy-with-.git-and-case-drift / decommissioned /
  settled / missing), THE SETTLEMENT's first-run line + the
  primed-guard silence, the >10 confirm's exact box (the 10-boundary
  included), the resolve engines' rows + settles + all-resolved
  arithmetic, the no-busy-guard pin, and THE FIELD-TAG LAW on the
  three payloads (`TestBotDoorPayloadFieldTags`).
- `webui/heroui_test.go` — 4 shell cases: the export + verify
  wrappers over the faked `BotQueueFetch` seam (the file lands, the
  box, the accounted summary), the retry two-phase (the empty box,
  the stale-Continue re-scan, the ≤10 immediate run with the
  pending-lists clearing + the `/direct` source + the held-seat
  clearing, the >10 ask + the confirmed run over the HELD urls), the
  resolve doors (the `/bot` source + the pending-seat assignment),
  and the heroflow law (the live keys, the dialog's own words, the
  import statement pin, the remaining tail T-79d/e's landing notes,
  the probe pick).

## The evidence

- **The `wails build`** over the 12th-rebuilt webkit tree: links
  clean (29.5 s); no mode flips (`git diff --raw`: the wailsjs files
  stay at their 755 baseline, content-only diffs = the regenerated
  bindings carrying the seven doors + the three payload classes); no
  stray binaries.
- **The gate legs**: first frame 892 ms / 307 ms / 269 ms (≤ the 1.5 s
  gate); the 10k flood through the production transport with 0 drops
  (555k lines/s accepted).
- **The screenshots** (VLM-verified, `t-79c-more-menu.png` +
  `t-79c-export-click.png`): the three bot-queue doors LIVE in the
  menu. The first pass called Retry Failed "grayed" — the 042/043
  emoji-led misread twin: settled per the law with the byte-level
  source check (the live keys + the uniform render path, pinned by
  `TestHeroflowBotQueueDoorsLive`) + a focused cropped re-pass (all
  three doors dark/LIVE, the bottom tail washed). **The click leg**
  (`SB_POC_OPEN_MORE=export`): the probe clicks 📋 Export All Links
  LIVE and the desk speaks the door's own refusal line end-to-end —
  "❌ Please enter the bot username first." The first click-leg run
  caught a real gap (the doors missing from heroflow.js's binding
  IMPORT list — a silent ReferenceError; now pinned by the import
  law test).
- **The battery at the close**: 2322 main-battery tests over 36
  packages plain + full -race green (+10 over SESSION-043's 2312 =
  the engine cases) + 67 webui tests over 4 (+4 = the shell cases);
  the touched packages 3× race-clean post-change; gofmt/vet clean
  over both modules; the MAIN go.mod + go.sum + the nested module's
  files byte-unchanged (0 diffs).
