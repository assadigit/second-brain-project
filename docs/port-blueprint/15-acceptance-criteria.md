# 15 — Acceptance Criteria

> How the implementation team demonstrates the port is complete. Every criterion is verifiable by
> an artifact (test run, report, zip, screenshot/recording) — no "looks done" claims.

## 1. Functionality parity

- **AC-1** The Go CLI implements the full flag surface (`spec-cli.md` §1) and passes the ported
  flag-surface tests.
- **AC-2** The websites pipeline reproduces the per-link state machine (gates, consolidation,
  review, retries, banishment + confirmation gate, tally) — the ported state-machine table tests
  green.
- **AC-3** The GitHub pipeline reproduces the per-repo loop incl. 404 quarantine, anonymous
  fallback, rate-limit law, banner handling, `_review` routing.
- **AC-4** Moves-as-corrections, one-open note reads, scan radar/weighing/validation/apply,
  hygiene, repo banish, sync-review model: ported suites green.
- **AC-5** Seals & publish: VaultSeal ladder + creation guard + local-only belt; GoodRepos;
  directory publish (index schema + private belt + tombstones); repo radar — integration tests on
  real git + fake GitHub green.
- **AC-6** Worker client: HMAC signature byte-accepted by the loopback rig; 1010-UA rejection
  honored; pairing flow works end-to-end against the rig.
- **AC-7** Telegram: queue fetch with `min_id`, mark-read, login+2FA — PoC-verified via gotd (or
  the fallback ADR documents the accepted change).
  *Amended 2026-10-15 (ADR-0004 Am.3, owner stamp "OPTION B CONFIRMED"): the desktop's
  direct Telegram leg is dropped — the queue rides the Worker HTTP API only (covered by
  the workerclient acceptance in T-34) and read-acknowledge is dropped. The gotd
  capability proof above stands for the Phase-8 optional revival; no MTProto acceptance
  criterion applies to the Go app v1.*
- **AC-8** The GUI exposes every page, dialog, and control enumerated in the GUI module specs,
  with the 10 honesty laws demonstrated (workflow test recordings + checklist).

## 2. Data compatibility

- **AC-9** The Go app opens a Python-created `cache.db` + `config.json` + `sync_review.db`
  (sanitized copies) and passes read/write probes without migration (fixture test).
- **AC-10** Config round-trip is byte-stable (merge law, no HTML escaping) — differential test.
- **AC-11** Side-by-side lock: Go and Python instances mutually exclude via `app.lock` —
  cross-binary test.
- **AC-12** Dry-run writes nothing (snapshot-diff test on a real vault copy).

## 3. Correctness (differential)

- **AC-13** The differential corpus (golden 30-link set, HTML fixtures, import grammar,
  corrections scenarios, banishment verdicts, scan-apply subset, dry-run reports) matches Python
  outputs under the documented normalization rules — final report archived in-repo.
- **AC-14** All pre-classified behavior-preservation items (doc 09 §5) have their dispositions
  implemented and tested (incl. the two documented fixes with regression tests).

## 4. Responsiveness (performance)

- **AC-15** Cold start (GUI) ≤ 1.5 s to first frame on the baseline machine; CLI `--status`
  ≤ 0.5 s (measured against T-04 baselines — report both numbers).
- **AC-16** Log desk sustains a 10,000-line flood with the UI staying interactive (coalescing
  parity: ≤120 ms flush, bounded batches).
- **AC-17** Full vault inventory scan of the owner-scale vault (≈1,300 notes) completes ≤ 50% of
  the measured Python baseline.
- **AC-18** No network or LLM operation blocks the UI thread (verified by the sentinel/batch
  running while interacting).

## 5. Testing

- **AC-19** Unit/state/contract/integration suites green in CI (`go vet` clean; `gofmt` clean);
  coverage of the ported characterization suites ≥ 90% of the enumerated case lists.
- **AC-20** The Python CI gate remains green and untouched through the migration (the Python app
  is not modified).

## 6. Packaging & release

- **AC-21** `GitCurator.exe` launches with no console window; `GitCurator-cli.exe` runs in
  terminals; both from the portable zip on a clean Windows 10 VM (checklist doc 10 §7 recorded).
- **AC-22** The deterministic zip rebuilds byte-identical (sha256 match on two builds from the
  same commit).
- **AC-23** A GitHub release carries `GitCurator-v<version>-windows.zip` + sha256 + the parity
  report.
- **AC-24** Runtime requirements documented (WebView2, git) with the check/bootstrap path.

## 7. Documentation

- **AC-25** Implementation report: per-task status (completed/pending/blocked), deviations from
  specs with reasons, updated ADRs (PoC verdicts), updated spec files where reality differed.
- **AC-26** Go README + updated WINDOWS-QUICKSTART covering the Go launch path; the parity report
  is owner-readable (plain language, per SPEC.md §0).

## 8. Sign-off

The migration is complete when AC-1…AC-26 hold and the owner has run the Go app on real data
(dry-run then live) alongside the Python app without data divergence. The Python app may then be
retired at the owner's discretion (a final backup + the last Python zip retained).
