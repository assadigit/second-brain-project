# T-04 — Python performance baselines (doc 07 §B.4) — measured numbers

Recorded 2026-10-15, SESSION-001, by `testdiff/python_baselines.py all`
(median-of-3 where noted). Environment: Debian 13 sandbox, Intel Xeon,
4 GB RAM, Python 3.12.15, PyQt6 6.11.0 **offscreen** (rootless libEGL),
reference repo `assadigit/GitCurator` @ `98f4fc7` (pristine after the runs —
`git status` clean). All numbers measured, none estimated. **This is a fast
datacenter CPU, not the owner's Windows 10** — the doc's "est. 2–5 s" for
real-hardware cold start stands as context; the numbers below are the
sandbox's honest baseline and the rig is rerunnable on the owner's machine.

## The doc 07 §B.4 table, filled

| Metric | Python baseline (measured) | Go target (context) |
|---|---|---|
| Cold start → first frame (GUI) | **≈ 805 ms** offscreen (driver t0→frame **794.5 ms** + interpreter boot 10.4 ms). Decomposition: PyQt6 import 22.7 ms · `gitcurator.gui.app` import **607.0 ms** · MainWindow construct 146.4 ms · show→first paint 17.4 ms | ≤ 1.5 s — Wails PoC measured **213–301 ms** (T-70) |
| CLI `--status` wall time | **49.3 ms** (missing-config path) / **893.6 ms** (config-card path, median of 3 — the card pulls the heavier imports) | ≤ 0.5 s |
| Log line ingest (desk) | **9.29 µs/line producer-side** (10,000 `log_message` calls in 92.9 ms — matches the in-repo 8.7 µs claim, re-measured); the 10k flood **drains in 1,960.9 ms** (~196 µs/line of document work at the 120 ms/250-row cadence) | ≥ same; UI interactive at 1,000-line bursts — Wails PoC: 10k accepted in 16 ms, 0 drops |
| Full vault scan inventory | **35.3 ms cold** for a synthetic 1,314-note vault (≈ 37.4k notes/s); re-read 32.1 ms; one-open cache fill 215.7 ms; **cache-warm 16.7 ms = 0.473× cold** | ≤ 50% of Python |
| Batch of websites (fake fetch+LLM) | the in-repo offline golden runner: **30/30 links in 205.1 ms** through the real pipeline (fake fetch + fake LLM; 0 invalid categories) ≈ 6.8 ms/link — a 10-website batch is sub-second | wall time within +10% |
| Sentinel pass (3 probes) | off-thread already (doc 07 §B.4 — no Python measurement required) | off-thread, < 250 ms each |
| Memory (idle GUI) | **RSS 121.6 MB** after first frame · **123.9 MB** after 3 s idle · **135.6 MB** after the 10k flood (VmHWM 135.6 MB) | ≤ 200 MB |

## The procedure's extra timed cases

| Case | Measured |
|---|---|
| Queue classification (fake links) | **1,142 links in 1.42 ms median** (805k links/s) — the `telegram_fetch_worker` extraction/dedup loop over a mixed corpus (714 GitHub + 286 non-GitHub + duplicates). The pure loop is NOT the historical freeze; the freezes came from running fetch+classify on the GUI thread (investigation 1-d) |
| note_reads one-open cache | visible in the scan row: warm cache halves the scan (**0.473×**); the cache-fill pass costs 215.7 ms (it writes the reads DB) |

## Caveats (honest)

- The scan corpus is synthetic (1,314 notes with realistic frontmatter
  shapes, 12 category folders, 5 root notes) — the owner's real vault may
  parse heavier; the rig is the deliverable, rerunnable anywhere.
- Offscreen Qt skips real-window compositing; on-screen Windows numbers
  will be higher (the doc's own estimate for the owner's machine: 2–5 s).
- Raw per-mode JSON: `runs/` is gitignored; the combined JSON is embedded
  below (trimmed to the numbers).

```json
{"cli":{"cli_status_missing_config_ms":49.3,"cli_status_empty_config_ms":893.6,"python_interpreter_boot_ms":10.4},
 "queue":{"messages":1142,"raw_links_classified":1142,"github_urls":714,"non_github_urls":286,"median_ms":1.42},
 "scan":{"synthetic_notes_created":1314,"notes_seen_by_scan":1319,"scan_cold_nocache_ms":35.3,"scan_reread_nocache_ms":32.1,"scan_cache_fill_ms":215.7,"scan_cache_warm_ms":16.7,"cache_warm_vs_cold_ratio":0.473},
 "gui":{"qt_import_ms":22.7,"gui_app_import_ms":607.0,"mainwindow_construct_ms":146.4,"show_to_first_frame_ms":17.4,"driver_t0_to_first_frame_ms":794.5,"log_flood_producer_ms":92.9,"log_flood_producer_us_per_line":9.29,"log_flood_drain_ms":1960.9,"log_flood_total_ms":2053.8,"rss_mb_after_first_frame":121.6,"rss_mb_idle_3s":123.9,"rss_mb_after_10k_flood":135.6,"vmhwm_mb":135.6},
 "golden":{"cmd":"run_golden_websites.py --offline","wall_ms":205.1,"ok":true}}
```
