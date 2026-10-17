# Module Spec — `app/gitcurator/core/hand_delivery.py`

- **LOC**: 919 · **Go home**: `internal/handdelivery` · **Status**: port
- **Responsibility**: the fourth door — walled links queued for the owner's own Chrome: JSON
  queue + README in the vault, optional open-in-Chrome, and the consume side turning delivered
  `.html` files into real `FetchResult`s (with re-verification of app-delivered pages and
  master-table gesture stamping).

## Constants & files (`hand_delivery.py:`)

- Folder `<vault>/_review/hand-delivered/`; `QUEUE_FILENAME queue.json` (shape
  `{"links": {url: {wall, queued, suggested, door, consumed, requeued}}}`; corrupt → empty);
  `README_FILENAME README.txt` (regenerated per enqueue).
- `HAND_MARKERS ("🖐","✋","🫱","hand-deliver","hand deliver","hand","chrome")`;
  `WALL_MARKERS ('http 403','http 405','http 429','http 451','bot defense','challenge',
  'blocked_bot','refused','ssl','handshake','chrome-impersonated','curl_cffi','fingerprint')`.
- `is_walled_reason` (164-174): case-insensitive substring over WALL_MARKERS.

## Symbols

| Symbol | Lines | Contract |
|---|---|---|
| `walled_retry_rows(state)` | 177-192 | retry rows whose `last_error` is walled |
| `suggested_filename(url)` | 200-222 | `hand-{domain-slug ≤40}-{md5(canonical)[:8]}.html`; slug strips ` /\\?%*:|"<>.~#&=+,@` and non-alnum → `-`; hash over `normalize_website_url` |
| `enqueue_hand_delivery(vault, urls, walls, log, open_chrome, door)` | 289-390 | URL-keyed merge (re-enqueue no-op; canonical re-keying heal; consumed re-enqueue resets); writes queue + README via dryrun stand-ins; optional Chrome open; returns `{'added','queued_total','opened','folder'}` |
| `find_chrome()` / `open_in_chrome(url)` | 398-459 | Windows paths → POSIX which; detached Popen or `webbrowser.open`; loopback never opens |
| `collect_delivered(vault)` | 467-492 | queued & unconsumed links whose suggested file exists non-empty; orphans ignored |
| `consume_delivered(vault, log)` | 520-583 | batch consume: stamps `consumed` `'%Y-%m-%d %H:%M'` **but keeps the file** (the record); dry-run reads but doesn't stamp |
| `hand_fetch_result(url, body, wall, door)` | 598-623 | builds `FetchResult(status='full', http_status=200, content_type='text/html', charset='utf-8', reason="auto-delivered via the owner's own Chrome (the fifth door's tab retry…)" or "hand-delivered via the owner's real Chrome — the machine doors were walled: {wall}")` |
| `take_hand_delivered(vault, canonical, log, allow_consumed=False)` | 626-720 | the per-link pre-fetch probe the pipeline calls BEFORE any machine door; legacy-spelling heal; `allow_consumed` = redo-consume (🖐 gesture re-reads without re-stamping) |
| `_status_is_hand(status)` | 728-743 | revived (♻/revived/restored) beats hand; any HAND_MARKER wins otherwise |
| `pending_hand_links(vault)` | 746-773 | queue rows with no delivered file (the fifth door's to-do list) |
| `stamp_hand_rows` / `stamp_hand_delivered_rows` | 776-907 | master-table column-7 rewrites (≥8 pipe-parts); ` 🖐 hand — queued {date} `; ` ✅ hand-delivered — fetched {date} `; updates `> Last updated:`; atomic writes; dead/reviewed/revived verdicts own the cell first |

## Integration points

- Pipeline: `website_pipeline.py:4604-4624` calls `take_hand_delivered` (door 0) and the redo
  path (`allow_consumed`) via the 🖐 gesture; harvest on success stamps `✅ hand-delivered —
  fetched <date>` + sweeps `_review` leftovers.
- chrome_tabs delivers pages here with `door='auto'`; consume re-verifies auto pages via
  `page_is_real` and discards failures.

## Tests

Port `test_handdelivery` + `test_handharvest` + `test_fifthdoor` selections: enqueue merge/heal,
suggested filename determinism, consume stamping (file kept), redo-consume, gesture status
grammar, master-table stamp precedence, queue corruption tolerance, dry-run laws.

## SESSION-018 shipped record (T-43, 2026-10-17)

- **The module is complete.** T-40 (SESSION-015) floored the consumption
  surface (`handdelivery.go`: the predicates, the paths, the queue +
  README writers, `SuggestedFilename`, `EnqueueHandDelivery`,
  `CollectDelivered`, `HandFetchResult`, `TakeHandDelivered`,
  `StampHandDeliveredRows`, the `Clock` seam); T-43 added:
  - `chrome.go` — `FindChrome` (the Windows chrome.exe ladder with
    LOCALAPPDATA expanded at call time / POSIX `exec.LookPath` over the
    same name ladder) + `OpenInChrome` (the honest open ladder: the
    found Chrome detached via `exec.Start` — the app never waits — else
    the system browser, loopback NEVER opens, the same one except-path
    message for a failed Chrome launch and a raising fallback). The
    seams: `findChromeFn` / `subprocessPopen` / `webbrowserOpen` /
    `chromeOSName` / `chromeIsFile` / `chromeWhich` (the reference's
    tests patch `hd.find_chrome`, `hd.subprocess.Popen`,
    `hd.webbrowser.open`, `os.name`, `os.path.isfile`, `shutil.which` —
    the port's battery patches these).
  - `pickers.go` — `WalledRetryRows` over the `RetryRowsState` seam
    (the picker's data; websitepipeline imports this package, so the
    concrete StateDB adapter lives with the future GUI/CLI call sites —
    the decomtable layering law) + `PendingHandLinks` (the fifth door's
    to-do list; a landed file counts regardless of content — `isfile`
    is the only gate).
  - `consume.go` — `ConsumeDelivered` (the batch consume: read, stamp
    `consumed` on the queue, keep the file; the auto-door re-verify;
    dry-run takes the pages WITHOUT stamping) + the `PageIsRealFn`
    seam (chrome_tabs.page_is_real — **unset until T-44 lands**; the
    honest degrade answers `(true, "")` exactly as the reference's
    except-path defines: "a verdict that cannot be asked is not a
    wall"; the ERROR-page discard battery lands with T-44) +
    `discardAutoPage`.
  - `stamp.go` — `StampHandRows` (the master-table queue stamp: RAW-URL
    `wanted` set — no canonical folding, no status-grammar checks: the
    stronger verdicts are the CALLER's first checks, exactly the
    reference's own split from `stamp_hand_delivered_rows`).
  - `EnqueueHandDelivery` gained the `openChrome` leg (the reference's
    fifth parameter — counts `report.Opened` through OpenInChrome's
    ladder).
- **THE GOLDEN**: `testdata/handdelivery.golden.json` (5 sections:
  `walled_retry_rows` — the real reference WebsiteStateDB's rows as
  input + the filtered output; `consume_delivered` — the five-link
  batch flow with the taken rows + the queue.json bytes + the log
  lines; `stamp_hand_rows` — the stamped count + the table bytes incl.
  the Last-updated refresh; `pending_hand_links`; `hand_fetch_result`
  — both wordings) generated by `testdiff/goldens/dump_handdelivery.py`
  from the unmodified reference @ 98f4fc7 with `hd.time` frozen at
  2026-10-17 12:00. hand_delivery is core/ and stdlib-only — no
  GUI-layer stubs needed (the plain generator pattern). The auto-door
  page in the consume flow is a REAL page: the reference's own
  page_is_real verdict and the port's degrade agree. The T-40 sections
  (enqueue / README / take / redo / heal) stay pinned by
  websitepipeline's golden — extend, never duplicate.
- **The behavior battery** (`handdelivery_test.go`, 50 tests): the
  wall predicate (incl. `WalledRetryRows` with erroring/panicking
  fakes — the empty picker), the suggested filename, the enqueue laws
  (write/noop/reset/non-http/corrupt/dry-run/empty vault/the
  openChrome counting), the Chrome ladder (loopback family, detached
  argv, the fallback, launch failure, the Windows candidates, the
  POSIX which ladder), the consume laws (file kept + queue stamped +
  second collect empty, empty page, orphans, dry-run unstamped), the
  blank-body law, the gesture grammar, `StampHandRows`' own laws, the
  pending legs (test_iconcolumn §6), the harvest-stamp battery
  (test_handharvest §1), the queue wire-format round-trip.
- **The pipeline legs** (`internal/websitepipeline/handpipeline_test.go`,
  20 tests): the reference suites' own §7/§4/§5 selections — the
  delivered page answers BEFORE any machine door (the fetcher is
  never called), the run's uninvited pull, the wall hint everywhere
  (error + log + note + the retry row's last_error), the opt-out, the
  owner's exact burned-out→gesture→delivery flow (retire + sweep +
  the harvest/reborn logs), the fresh-retry shape, the half-fetched
  exemption + its control, the stored-link control, the
  delivered-page-alone reborn, the machine-success harvest, the
  run-ends-with-refresh leg, and the three cached probes.
- **Reality-wins fixes this task forced**:
  1. **The blank-body law** (`pyBytesBlank`): Python's
     `not body.strip()` on BYTES strips ASCII whitespace only — a
     NBSP-only body is NOT blank for the reference (it gets consumed).
     The T-40 floor used `strings.TrimSpace` (Unicode-aware — wrongly
     blanked the NBSP body); fixed in both `TakeHandDelivered` and
     `ConsumeDelivered`, pinned by `TestTakeHandDeliveredBlankLaw`.
  2. **A latent T-42 flake root-caused** (githubpipeline, found while
     verifying this session): the final report's "Repos by Category"
     table iterated a Go MAP then stable-sorted — ties flipped
     randomly per run (Python dicts keep insertion order; Go map
     ranges are randomized). Fixed to ride the tracked
     `categoryOrder` slice; verified with 20 package runs + 6
     full-suite `-p 16` runs.
- **Deferred to T-44** (internal/chrometabs): wiring `PageIsRealFn`
  (the auto-door ERROR-page discard legs of test_fifthdoor §6: the
  challenge/crash/ERR-page verdicts, the consume-level discard, the
  pending-reask) + `find_chrome`'s chrome_tabs callers.
