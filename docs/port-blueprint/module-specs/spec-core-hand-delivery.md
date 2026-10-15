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
