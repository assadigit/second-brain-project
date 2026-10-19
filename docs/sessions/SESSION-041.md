# SESSION-041 — 2026-10-20 — Phase 6: T-78 COMPLETE (the side-by-side verification vs the Python app — the lock mutual exclusion across both binaries, the same-data config laws) — 21/21 side-by-side checks GREEN; ONE real divergence found and fixed (the EPERM liveness branch); 2253 main-battery tests over 36 packages + 56 webui over 4; the thirty-sixth full wipe repaired BEFORE work (the webkit tree deliberately deferred)

## Goal (from NEXT-SESSION.md at start)

SESSION-041 = **T-78 (S): THE SIDE-BY-SIDE VERIFICATION VS THE PYTHON
APP** per doc 13's row (ADR-0008): the Go app and the Python app against
the SAME data (the same app/ folder: config.json, cache.db, the vaults),
proving the compatibility laws live — the app.lock mutual exclusion
ACROSS THE TWO BINARIES (the same file, the same PID-liveness law, the
alive-refusal), the merge-never-rebuild config law on the reference's
own config shape, the byte-comparable JSON (SetEscapeHTML(false)), the
user-config fill/mirror. ACCEPTANCE: lock test across both binaries.
Scope honestly (the PyQt6 question); surface the owner's bookkeeping
question; T-78 is the LAST Phase 6 task — the phase-exit question is
the owner's to see, not to improvise.

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-sixth full wipe repaired BEFORE work.** At boot BOTH
  repos, the toolchain, the venv, the module cache, the worklog, AND
  the rootless webkit tree were gone (verified by directory survey +
  `git status` evidence: both `cd`s failed). Recovery, all from the
  committed recipes, all in foreground chunks:
  - Both repos re-cloned with the ephemeral `GIT_CONFIG_COUNT` PAT form
    (nothing persisted anywhere): the work repo tip verified at
    `e602cd6` (SESSION-040's docs close — beyond the prompt's `c14fb8f`
    expectation; the tree clean; the remote URL clean); the reference
    checked out at `98f4fc7` = PINNED-GITCURATOR-REV (byte-match).
  - go1.27.2 from go.dev, sha256
    `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
    (byte-equal to TOOLCHAIN.md Am.2's committed 64-char pin); the main
    go.mod verified unchanged (the authorized four + the 1.24.1
    directive); the venv (Python 3.12.15, stdlib-only); the wails CLI
    v2.16.0 (GOBIN install).
  - **The webkit tree deliberately NOT rebuilt** — T-78's legs are
    data-side (no `wails build` was needed); the next GUI-touching
    session rebuilds it per the committed recipe (recorded here + in
    the roadmap entry + the state card).
  - The re-verification law executed before any new work: the FULL
    main battery + the webui battery green pre-work — **exactly the
    recorded state: 2252 over 36 + 56 over 4, race-clean**.

- **The PyQt6 scope question, probed and answered with hash-level
  evidence** (the task's own mandate):
  - The reference's lock lives INLINE in `gui/app.py`'s `main()` (the
    dance at lines 393-440) with `_is_process_running` imported from
    `gui/headless.py` (246-282) — not a callable module surface.
  - The import chain is PyQt6-gated: `gui/headless.py:34` →
    `gui/icons.py:45` (`from PyQt6.QtCore import QByteArray, QSize, Qt`)
    → `ModuleNotFoundError: No module named 'PyQt6'` under the
    stdlib-only venv — the raw failure fires BEFORE headless's own
    graceful gate (its try/except PyQt6 star-import that sys.exit(1)s
    is never reached). `gui/dialogs.py` gates the same way from
    app.py's own imports.
  - `--cli` (dispatched in main.py BEFORE the GUI import) and
    `--headless` (the first branch of main()) BOTH exit BEFORE the lock
    dance — **the GUI entry is the reference's ONLY lock-taking
    binary**, so no CLI-shaped Python binary can stand in for the lock
    leg.
  - The scoped leg, per the mandate's own words ("check whether a
    minimal harness can drive the reference's own lock code"):
    **extract-and-exec** — `testdiff/side_by_side.py` slices the dance
    + the liveness probe VERBATIM from the pinned read-only checkout by
    line range, with structural pins (first-line + last-marker
    assertions) + the rev guard that fail loudly on drift, and records
    both slices' sha256 in the report (gui/app.py:393-440 →
    `8025651fcb52eb5c…`, gui/headless.py:246-282 → `c96105db365077cb…`).
    The Python half of every lock scenario IS the reference's own bytes.

- **The machine (committed)**: `testdiff/side_by_side.py` (the
  orchestrator + report writer) + `testdiff/goside/` (the Go driver — a
  main-module package of test instrumentation, NOT a shipped command:
  lock-hold/lock-try over `internal/app.AcquireAppLock` — the exact
  door the wails shell takes before its window — and config-save/fill/
  mirror over `internal/config`'s app-door sequence + the user-memory
  doors, the user-memory path steered through the env override exactly
  like the shells). The harness builds goside into `testdiff/runs/bin/`
  (gitignored); no stray binary in the tree.

- **LEG A — the acceptance, 9/9 GREEN:**
  - A0 the fresh-take BYTES law on both sides: the lock file is the
    bare decimal PID, no newline, no padding (hex-pinned evidence).
  - A1 **Go holds → the reference's own dance REFUSES**: exit 1, the
    exact sentence "Another instance is already running (PID <go-pid>).
    Exiting.", the lock file untouched (still the Go holder's bytes).
  - A2 **Python holds → Go REFUSES**: exit 1, the same sentence
    (AcquireAppLock's own print), GOSIDE REFUSED, the file untouched.
  - A3 the Go holder kill-9'd → the reference heals the stale: the
    exact stale sentence + the takeover.
  - A4/A4b the Python holder kill-9'd → Go heals the stale; the healed
    file's bare bytes pinned.
  - A6 **the EPERM probe — the divergence the audit caught and FIXED**
    (below): both binaries read the live not-ours PID (PID 1 on this
    box) as DEAD and take the lock over identically, live.

- **THE REAL DIVERGENCE FOUND AND FIXED — `internal/app/process_posix.go`:**
  the port's POSIX liveness probe read EPERM (kill(pid, 0) → the
  process exists but belongs to another user) as ALIVE; the reference's
  blanket `except (OSError, ProcessLookupError, ValueError, Exception):
  return False` swallows the PermissionError and reads DEAD. A lock
  whose stale PID was recycled to a stranger's process was therefore
  take-over-able in Python but REFUSED by Go — the two binaries could
  DISAGREE on the same file (the mutual-exclusion law broken in the
  edge). Fixed: the verdict factored as the pure `killVerdict(err)`
  (nil ⇒ alive; EPERM ⇒ dead — the reference's except; anything else ⇒
  dead), pinned hermetically by the new
  `TestKillVerdictMirrorsTheBlanketExcept` (POSIX-tagged test file —
  the production target is Windows, the tag keeps the battery
  cross-platform). The pre-fix divergence was LIVE-reproducible (A6's
  probe against PID 1); post-fix both sides agree. Also tightened, not
  weakened: `TestAppLockFreshAcquire` + `TestAppLockHealsStaleLock` now
  pin the RAW bytes (no TrimSpace leniency — the bare-decimal law).

- **LEG B — the same-data config laws, 12/12 GREEN** (the reference's
  own pure modules — `gitcurator.core.storage`, `core.user_config`,
  `constants` — imported in-process as the oracle, all PyQt-free,
  probed):
  - B1 both apps write **BYTE-IDENTICAL config.json** from the same
    input + the same updates through their real doors (Go:
    LoadAppConfig + MergeConfig + WriteConfigFile — the app door;
    Python: json.load + merge_config + write_config_file — the GUI's
    save engine): sha256 `e15c880650e0f07f…` equal on both sides.
  - B1a-d the observables, pinned byte-level on the Go-written file:
    unknown keys survive (both the fixture's and the update's), tuning
    keys unstomped (99/7/2.5 kept — never the 60/3/0.5 defaults),
    nested dicts merged key-by-key (proxy.host edited + ollama.model
    added, the rest kept), the non-ASCII + HTML-ish value's raw UTF-8
    byte-present (ensure_ascii=False + SetEscapeHTML(false) both live).
  - B1e/B1f **the cross re-save stability**: Python's re-save of Go's
    file and Go's re-save of Python's file are both byte-stable — all
    FOUR sha256 equal (the side-by-side install's core promise: neither
    app disturbs the other's file).
  - B2 the fill: same bytes + same counts in both modes (kept: 2 fills
    + the local-wins law pinned; fresh: 8 restores), the memory fixture
    written by the reference's OWN mirror door.
  - B3 the mirror: identical `.gitcurator` bytes (sha256 equal), the
    never-loses law (a memory-only remembered key survives BOTH
    mirrors), the 0o600 permission law on both sides.

- **The docs**: doc 13's T-78 row COMPLETE (the full record); doc 12's
  SESSION-041 entry; ADR-0008 **Amendment 1** (the verification record
  + the EPERM divergence, append-only); README's SESSION-041 status;
  testdiff/README.md's map (side_by_side.py + goside/ + the usage
  line); the evidence archived at
  `docs/poc-artifacts/t-78-side-by-side.md` (the signed-off report).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-78's acceptance = "lock test across both binaries" —
**GREEN with live proof**:

- The full side-by-side run: **21/21 checks GREEN**
  (`testdiff/side_by_side.py`, both legs; the report at
  `testdiff/runs/side-by-side/report.md`, the signed-off copy archived
  at `docs/poc-artifacts/t-78-side-by-side.md` with the provenance:
  the pinned rev, the extraction line-ranges + sha256s, every
  scenario's exit codes + spoken sentences + file bytes).
- AC-11 (side-by-side lock: Go and Python instances mutually exclude
  via app.lock — cross-binary test) is now DEMONSTRATED live, in both
  directions, plus both stale-healing paths + the bytes law.
- AC-10's differential substance (config round-trip byte-stable) is
  demonstrated live cross-app (B1/B1e/B1f — all four sha256 equal).
- The battery at the close: **2253 main-module tests race-clean over 36
  packages (2252 + 1 new) + 56 webui tests over 4**; the internal/app
  package 3× clean post-change (the race-law ritual for the touched
  package; no new concurrent seam was introduced — the EPERM fix is a
  pure branch flip); gofmt/vet clean over both modules; the MAIN
  go.mod byte-unchanged all session (verified at every step); no mode
  flips; no stray binaries (the goside build lands in gitignored
  runs/bin/; the one stray repo-root binary from a direct `go build`
  was caught + removed immediately).
- **Phase 6's task list (doc 13) is now EMPTY** — T-70…T-78 all
  COMPLETE. The phase-exit question is surfaced to the owner below
  (NOT improvised).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-78 row
  COMPLETE (the full record incl. the 36th-wipe repair, the PyQt6
  evidence, the EPERM fix, both legs' proof).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-041
  entry (incl. the webkit-tree deferral decision).
- `docs/port-blueprint/decisions/0008-config-data-compat.md` —
  Amendment 1: the T-78 verification record + the EPERM divergence
  found and fixed (append-only; reality recorded).
- `README.md` — the SESSION-041 status entry.
- `testdiff/README.md` — side_by_side.py + goside/ in the tree map +
  the usage block.
- `docs/poc-artifacts/t-78-side-by-side.md` — the archived evidence.

## Blockers & decisions needed

- **No new blockers.** The PyQt6 wall was probed, evidence-pinned, and
  scoped past WITHOUT improvising (extract-and-exec drives the
  reference's own bytes; the technique + its justification are recorded
  in the harness docstring + the row + ADR-0008 Am.1).
- **THE PHASE-EXIT QUESTION (the owner's call — T-78 was the LAST
  Phase 6 task)**: Phase 6's task list is empty and its development
  gates are green (the T-70 entry gate + every task acceptance since).
  What remains before the GUI-parity RELEASE is demonstrated lives in
  doc 15 (AC-8's every-control coverage, AC-9's formal probe set,
  AC-15/16/18's owner-machine measurements, AC-21's clean-VM leg) —
  the standing reading (T-70's gate doc + ADR-0001 Am.2) assigns those
  to Phase 7's clean-VM checklist. The fork:
  - **(A) Declare Phase 6 CLOSED → open Phase 7** (T-80: the
    deterministic zip builder + version stamping — doc 13's Phase-7
    book).
  - **(B) Add a tasks addendum first** (a T-79: the More menu's
    honestly-disabled tail — the per-test dialogs, the linking tools,
    the bot-queue tool dialogs, the Chrome doors, the decommission
    picker, the export zip — as named tasks) before closing.
  - This is the SAME fork as the bookkeeping question below — one
    decision resolves both.
- **The bookkeeping question (carried from the v41 card, NOW DUE)**:
  the More menu's honestly-disabled tail has no named task on doc 13 —
  the owner's call: a tasks addendum (T-79) vs folding into T-83's
  parity report (Phase 7).
- Carried, unchanged: the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (thirty-second
  session carrying it); x/net/html keep-or-swap; the answering-window
  belt + the T-71 merged-verdict question (SESSION-039's record).
- **Environment note**: the webkit tree is ABSENT (deliberately not
  rebuilt this session — no wails build was needed); the next
  GUI-touching session rebuilds it per build/TOOLCHAIN.md Am.3 + Am.4
  (expect the hardlink fallback: symlink(2) was blocked at boot again —
  the wipe survey found the convenience symlinks down).

## Owner actions requested

- **The phase-exit call (A vs B above)** — one word closes the fork;
  the next session's prompt (NEXT-SESSION §5) carries both paths and
  asks for the call first.
- When convenient (no rush): the standing calls — dissolution,
  vanilla-vs-vite.

## Handoff

- **SESSION-042 = the owner's phase-exit call resolved first, then
  either T-80 (Path A: Phase 7 — the deterministic zip builder +
  version stamping + manifest/icon) or T-79 (Path B: the GUI doors
  addendum).** The state card v42 carries both paths + the prompt.
- NEXT-SESSION.md rewritten: yes.
- Push state: all work committed + pushed (the ephemeral PAT; never
  stored, never written).
