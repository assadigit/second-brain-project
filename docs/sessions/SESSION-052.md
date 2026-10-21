# SESSION-052 — 2026-10-21 — Phase 7: the AC-17 benchmark COMPLETE — the Go vault-scan wall-time measured (14.6 ms vs Python's 32.8 ms, 44.5%); the parity report's one open performance row closed; battery 2377 + 89 green

## Goal (from NEXT-SESSION.md at start)

SESSION-052 = the Phase-7 finish, OWNER-INPUT DRIVEN — the decision
tree: (a) T-82 results → the intake + the flip + the Phase-7 exit
evaluation; (b) "ship it" → the milestone release session; (c)
otherwise surface the menu and let the owner pick (Phase-8 scoping /
the AC-17 benchmark / small polish). One slice done properly with
proof beats two half-done.

## What was done (task IDs, files, commits — cite hashes)

- **The 47th wipe absorbed before work** (the environment was fully
  down: both repos, the toolchain, the venv, the wails CLI, the
  module caches). Rebuilt per the committed recipes: the work repo
  re-cloned **anonymously** (`assadigit/second-brain-project`, public;
  HEAD `5325623` — the exact expected SESSION-051 close, tree clean),
  the reference re-cloned with the ephemeral PAT then **scrubbed +
  byte-scanned** (`assadigit/GitCurator` @ `98f4fc71cfb8…` —
  byte-equal to `testdiff/PINNED-GITCURATOR-REV`; the post-scrub
  scan of both `.git` trees: zero hits), go1.27.2 downloaded +
  sha-verified against TOOLCHAIN.md Am.2's committed pin
  (`ecbadb99…`, 64 hex chars, programmatic compare — MATCH), the
  venv (plain form; ensurepip present, Python 3.12.15), the wails CLI
  v2.16.0. **No webkit rebuild** (no Linux GUI build this session —
  the benchmark is engine-level; the windows legs are CGO-free).
- **The decision tree ran path (c)**: the owner brought no T-82
  results and no "ship it". The menu is surfaced in this report +
  the state card v53; the executed slice was the tree's own (c)
  benchmark option — **the AC-17 vault-scan wall-time**, the parity
  report's one open performance row (a Phase-7 loose end that needs
  no owner input and preempts no owner call).
- **The rig — `testdiff/gobench` (NEW)**, the goside pattern (test
  instrumentation, not a shipped command): one mode verb, `scan`,
  mirroring `testdiff/python_baselines.py`'s mode_scan exactly —
  `_build_vault`'s byte-mirror as the corpus generator (1,309 notes
  per doc 07 §B.4 + 5 root, six categories, the same frontmatter
  template), the same four timed passes through the production door
  `vaultscan.ScanVaultInventory` (nil log/progress exactly like
  `internal/syncreview`'s production call; `notereads.New(<tmp>/
  reads.db)` as the one-open reads cache for passes 3–4), the same
  JSON field names + the census counts, the same
  `###BASELINE_JSON###` sentinel, `--vault DIR` to keep the corpus
  for the cross-side checks. gofmt clean, vet clean, build clean; the
  package carries no tests — the battery shape is unchanged.
- **The measurements** (3 fresh runs per side, same box, 2026-10-21;
  the Python side is the COMMITTED T-04 rig re-run fresh against the
  pinned reference):
  - Go: cold 15.1/14.6/14.2 → **median 14.6 ms**; re-read 13.1;
    cache-fill 73.5; warm 5.3; ~90.5k notes/s cold.
  - Python: cold 32.8/33.1/32.7 → **median 32.8 ms**; re-read 30.6;
    cache-fill 249.0; warm 16.9; ~40.2k notes/s cold (T-04's recorded
    baseline 35.3/32.1/215.7/16.7 reproduced within ~7% on cold, ~1%
    on warm — the same sandbox class, as assumed).
- **The proofs**: (1) same-input — the Python corpus (built by the
  COMMITTED generator through a one-shot driver) and the Go corpus
  tree-hash EQUAL: sha256 `1091ed894d33…` over 1,314 files each;
  (2) same-output — the reference's own `scan_vault_inventory` over
  the GO-built vault returns the IDENTICAL census (all ten counts +
  `notes_seen_by_scan` 1319 — the same arithmetic the Python rig's
  own JSON reports).
- **The docs**: `docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`
  (NEW — the full record: method, environment, both proofs, all six
  runs' numbers, the verdict table, the honest caveats, the rerun
  recipes + the driver's bytes); `docs/PARITY-REPORT.md` (§5's row
  measured + closed — the gate cell now `≤ 50% of Python ✅`; the §5
  note; §9's open bullet retired; §10's double-check line; the dated
  tail note); doc 15 §4's AC-17 dated completion note; doc 12's
  SESSION-052 entry; the README's SESSION-052 status entry.
- **Commit `44a8bdb`** (6 files, 596 insertions, 11 deletions; the
  commit message byte-verified identical to its -F FILE source after
  the known renderer display-eat `[main 44a8bdb]` → `ain 44a8bdb]`
  struck again — the log line carried the true hash). Pushed with the
  ephemeral GIT_CONFIG_COUNT PAT form; the post-push scrub check
  clean; **go-ci run `39526755160` at `44a8bdb`: completed success**
  (watched to completion — normal CI hygiene held).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**The session slice (AC-17): PASS.** The acceptance (doc 15 §4): "Full
vault inventory scan of the owner-scale vault (≈1,300 notes) completes
≤ 50% of the measured Python baseline." Evidence: Go cold **14.6 ms**
(median of 3) vs Python **32.8 ms** same-box = **44.5%**, and vs
T-04's recorded baseline 35.3 ms = **41.4%** — under the 50% gate on
both comparisons; cache-warm 5.3 vs 16.9 (31.4%); cache-fill 73.5 vs
249.0 (29.5%); ~90.5k vs ~40.2k notes/s (2.25×). Same-input proven
(tree-hash equal, 1,314 files); same-output proven (the census
identical, `notes_seen_by_scan` 1319 both sides); the full record:
`docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`.

**The close battery — the exact committed shape**: main plain
**2377 top-level PASS over 36 ok packages, 0 FAILs, 0 skips**; main
**race: 36 ok, 0 FAILs, 0 data races**; webui **89 top-level PASS
over 4 ok packages, 0 FAILs**; gofmt clean both modules; vet + build
clean. **go-ci green on the push** (run `39526755160`). The honest
boundary: the corpus is synthetic (T-04's own caveat carried — the
owner's real vault may parse heavier; the rigs rerunnable anywhere
incl. the owner's Windows box), and the real-vault number still lands
at the owner's first live run (stated in the artifact §6, doc 15's
note, and the parity report §5).

**Phase 7's state is unchanged by design**: T-80 ✓, T-81 ✓, T-83 ✓,
T-82 PARTIAL (the owner's live legs). What remains in Phase 7 is the
OWNER's: T-82's live checklist + the milestone release decision. The
AC-17 row was the last autonomous Phase-7 item; with it closed, the
parity report has no open implementation row left.

## Spec updates made (file → what changed, why)

- `testdiff/gobench/main.go` — NEW (the AC-17 rig; mirrors the
  committed T-04 rig's scan mode; no runtime/engine bytes touched).
- `docs/poc-artifacts/ac17-go-vaultscan-benchmark.md` — NEW (the
  evidence artifact).
- `docs/PARITY-REPORT.md` — §5's row closed with the measured
  numbers (reality wins: the row said "not yet measured"; it is now
  measured); §9's open bullet retired; §10's line; the dated tail.
- `docs/port-blueprint/15-acceptance-criteria.md` — AC-17's dated
  completion note (the AC-25/AC-26 pattern).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-052
  entry.
- `README.md` — the SESSION-052 status entry.

## Blockers & decisions needed

- **NONE blocking.** The AC-17 slice is complete.
- Carried, unchanged: the milestone release decision (the owner's);
  T-82's live execution (the owner's); the dissolution call
  (ADR-0002 Am.5); the vanilla-vs-vite call; the go-directive
  acknowledgement; x/net/html keep-or-swap; the answering-window belt
  + the T-71 merged-verdict question; the batch-scorecard modal
  question; the v0.91.0 zip-parity reading; the deferred test debt.

## Owner actions requested

1. **T-82's live execution, when convenient**: run
   `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1's nine items
   (against the current `v0.92.0-go.1` zip for the CLI legs, or the
   next release for the full two-exe set) and fill §2 — the signed
   checklist closes T-82.
2. **The milestone release decision**: the desktop app is complete,
   proven, and one `git tag` away from shipping as the next release
   (the workflow is proven end-to-end + cross-machine). Say the word
   and the next session cuts it (VERSION → tag → the proven workflow
   → the release report per AC-23's scaffold).
3. **Otherwise — pick from the menu for Phase 8** (the owner's
   strategic call): the dormant cloud quartet revival, the gotd/td
   revival, Linux/macOS builds, localization — doc 12's future-work
   register has the details; or small polish from the deferred-debts
   list. The AC-17 benchmark option is now closed (done this session).

## Handoff

- **Next session: SESSION-053 = the Phase-7 finish, owner-input
  driven** — the same decision tree, minus the benchmark option:
  (a) the owner brings T-82 results → the intake + the T-82 flip +
  the Phase-7 exit evaluation; (b) the owner says "ship it" → the
  milestone release session; (c) otherwise → Phase-8 scoping from
  doc 12's register or small polish from the deferred-debts list —
  surface the menu and let the owner pick.
- NEXT-SESSION.md rewritten: yes (the state card v53, the
  SESSION-053 prompt inside).
- Push state: CLEAN — the work pushed at `44a8bdb` (go-ci
  `39526755160` green); the close (this report + the state card)
  pushes with this session's final commit.
