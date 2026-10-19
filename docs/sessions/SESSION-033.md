# SESSION-033 — 2026-10-19 — Phase 6's entry: T-70 COMPLETE (the PoC C GUI-shell gate re-run with ONE REAL JOB — the verdict GREEN) — 1932 + 12 tests race-clean; the twenty-seventh full wipe repaired BEFORE work + one mid-session degradation (the second full toolchain rebuild) survived without losing work

## Goal (from NEXT-SESSION.md at start)

SESSION-033 = **Phase 6's entry: T-70 PoC C THE WAILS GUI SHELL GATE** (per
ADR-0001's PoC C terms + the roadmap §Phase 6 entry + doc 07 §B.4): the
themed window (the spec-gui-theme-icons tokens), the log desk, ONE REAL job
wired through — startup ≤ 1.5 s and a 10k-line log flood staying
interactive, the verdict recorded per ADR-0001's gate discipline. The Wails
toolchain record (ADR-0002 Am.3-style) surfaced BEFORE the first wails
invocation. Then T-71 only if capacity remains; one slice done properly
beats a rushed whole.

## What was done (task IDs, files, commits — cite hashes)

- **The twenty-seventh full wipe repaired BEFORE work** — commit `5a688a2`:
  both repos re-cloned with the ephemeral `GIT_CONFIG_COUNT` header form
  (no PAT persisted in either clone's config — verified by direct config
  read); the work repo at origin/main `1cb6d28` (a39cb3d an ancestor); the
  reference at `98f4fc71cfb8` (byte-equal to PINNED-GITCURATOR-REV);
  go1.27.2 sha256-verified against go.dev's published JSON; the Python 3.12
  stdlib-only venv; the symlinks; go.mod verified unchanged (the authorized
  four). The recovery battery surfaced ONE flake the earlier boxes never
  hit: `TestSocksConnectSendsHostname` — the fake SOCKS5 server published
  its parsed record AFTER writing the HTTP response (a TOCTOU: `client.Get`
  returns on the headers; a loaded 2-CPU box let `records()` read the
  still-empty slot). Characterized first (20/20 green in isolation), then
  fixed by moving the publication point BEFORE the response write — every
  assertion unchanged; 2×25 concurrent stress green; the battery closed at
  exactly 1927/1927 over 34 packages.

- **T-70 preflight — commit `86790fb`**: THE WAILS BUILD-TIME TOOLCHAIN
  RECORD (ADR-0002 Am.3 + TOOLCHAIN.md Am.3) landed in git BEFORE the first
  `wails` invocation, per the mandate: the CLI pinned at v2.16.0
  (ADR-0001 Am.2's confirmed version; v2.17.0 exists and is deliberately
  not taken), GOBIN into the toolchain tree, BUILD-TIME ONLY; the runtime
  story (the nested webui/ module: require wails v2.16.0 + replace to the
  parent; `go 1.25.0` — the pin's own floor, verified from the proxy's
  published .mod; the main module's set stays the authorized four
  byte-level); the frontend step (NONE for the PoC — vanilla assets, the
  wails.json install/build commands empty; the production vanilla-vs-vite
  question OPEN for the owner); the CGO story (Linux dev legs = CGO + the
  rootless webkit tree; the Windows exe = CGO-free cross-compile).

- **T-70 slice 1 — commit `937797a`**: `guibridge/` — the interim real-job
  seam: `RunNoteStateBaseline` drives the REAL engine composition of
  record_baseline_if_empty (notestate.ScanVault → managed filter → per-note
  read + helpers.DecodeUTF8Replace + ComputeFingerprint + DB.Upsert; the
  reference's own final log line byte-identical; the one-time law; the
  DSN-empty-path hazard refused at the door) + 5 main-battery contract
  tests. 1932 tests / 35 packages.

- **T-70 slice 2 — commit `8939cc4`**: THE GUI SHELL LANDED at `webui/`:
  `main.go` (the themed window — the LIGHT tokens, 1000×750, cream
  BackgroundColour), `deskbus/` (THE COALESCED LOG's Go half — the 120 ms
  beat, ≤250-row batches, cap-1024 backpressure, the ack'd synchronous
  clear door; 8 headless tests), `fixture/` (the deterministic 800-note
  synthetic vault; 4 tests), `frontend/` (vanilla: the LIGHT token table as
  CSS vars, the btn_kind grammar, the SegmentBar twin, the chips/search/
  clear doors, the CACHE-MIRROR desk — cap 1000, filter re-render from
  cache, per-row tooltips, ONE fragment+scroll per batch — the four
  lucide-static v0.544.0 glyphs at verbatim geometry, the flood probe),
  `app.go`/`job.go` (the cold-start clock, THE REAL JOB, the flood leg with
  honest accounting, the Stats JSON), `build-support/` (the LD_PRELOAD
  webkit path shim — grown to intercept posix_spawn, glib's spawn path — +
  run-gate.sh, both COMMITTED this time). **ADR-0002 Am.4** (reality wins,
  in the same commit): Am.3's claim that the internal-package rule forces
  the seam is FALSE — the boundary probe (a direct `internal/notestate`
  import from webui/: `go build` exit 0, `go vet` exit 0) disproved it;
  guibridge's doc restated to the honest rationale (doc 04 §2's event-seam
  shape, main-battery tested; the T-72 retirement law stands); webui/
  README.md rewritten to the current truth.

- **T-70 COMPLETE — commit `7f613dd`**: THE VERDICT RECORDED —
  `docs/poc-artifacts/t-70-gate.md` (the full evidence document), ADR-0001
  Am.3 (the PASS verdict on the gate's own ADR), the task-list T-70 row
  COMPLETE, the roadmap's Phase-6 entry GREEN note, the README status.

- **THE MEASURED EVIDENCE** (two runs — see the gate artifact for the full
  table): first frame **266 ms / 679 ms** (the gate ≤ 1.5 s; run 2 after the
  mid-session degradation + rebuild); the 10k flood **458,695 / 454,133
  lines/s, 0 drops**, the flood's DOM-append cost to the main thread
  **23–25 ms max per batch** (638–741 ms total — the loop free ~93% of the
  window); THE REAL JOB **800 notes in 94–101 ms** (6 stubs never adopted,
  rows DB-verified); the Windows exe **14.95 MB CGO-free**; the desk's cap
  law **VLM-verified** off the screenshot ("1000 rows / cache 1000", the
  themed window, the filled hero, the 800/800 progress row, the
  glyph-colored mono rows — every number cross-checked against the results
  JSON).

- **The mid-session degradation survived**: between the two gate runs, the
  wipe event took the Go toolchain, the module cache, the webkit tree, /tmp,
  and the venv — the repos and ALL in-flight work survived; 584 tracked
  files' modes flipped 644→755 (content intact — diagnosed via
  `git diff --raw` mode-only rows, restored surgically via chmod from the
  parsed list, never checkout). Everything rebuilt from the committed
  recipe; the gate re-run GREEN as the reproducibility leg. One anomalous
  battery run in the immediate settling window (3 packages, names not
  captured) was followed by two consecutive fully-clean runs (1932/1932
  over 35 packages) — recorded for honesty.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**T-70 = PASS — the Phase-6 entry gate is GREEN.**
- The gate's terms (ADR-0001 §Decision 4 + the roadmap entry + the row):
  the themed window ✓ (the spec-gui-theme-icons LIGHT tokens — VLM-verified
  screenshot); the log desk ✓ (the full coalesced-log contract, both
  halves); ONE REAL job wired through ✓ (the note-state baseline — REAL
  engine composition, not a synthetic loop; guibridge contract tests +
  the in-shell 800-note run); startup ≤ 1.5 s ✓ (266/679 ms); the 10k
  flood stays interactive ✓ (0 drops; 23–25 ms max handler time per batch;
  the desk rendered + the probe chain alive); the verdict recorded ✓
  (ADR-0001 Am.3 + t-70-gate.md + the row + the roadmap + the README).
- Honest caveats, recorded in the artifact: the WM-less-Xvfb timer clamp
  makes the task-gap instrument read throttling (the handler-time
  instrument + the rendered desk are the interactivity evidence); the
  Windows runtime + bootstrapper legs remain the owner's Phase-7 items;
  all numbers from a 2-CPU box.
- **T-71: NOT STARTED** — capacity ran short after the two environment
  rebuilds; per the mandate's own law, T-70 done properly with the verdict
  recorded beats a rushed T-71. It is SESSION-034's primary task.

**Arithmetic**: 1927 (start) + 5 (guibridge) = **1932 main-battery tests
race-clean over 35 packages**; + **12 webui tests** (deskbus 8 + fixture 4,
run from the nested module). go.mod unchanged (the authorized four); zero
unauthorized dependencies (wails v2.16.0 lives only in the nested module's
go.mod per ADR-0002 Am.3).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/decisions/0002-module-name-and-layout.md` — **Am.3**
  (the Wails build-time toolchain record — pinned CLI, nested module,
  frontend step, CGO story; surfaced BEFORE use) + **Am.4** (the
  internal-boundary claim disproved by the boundary probe — reality wins,
  fixed in the same commit as the disproving code).
- `build/TOOLCHAIN.md` — **Am.3** (the wails CLI install record + the
  runtime/CGO/frontend facts) + the Am.3 correction note (pointing at
  Am.4).
- `docs/port-blueprint/decisions/0001-gui-framework.md` — **Am.3** (the
  Phase-6 entry gate verdict: PASS, with the evidence table and the
  standing caveats).
- `docs/poc-artifacts/t-70-gate.md` — NEW (the gate's evidence document);
  + `t-70-gate-results.json`, `t-70-gate-results-rerun.json`,
  `t-70-gate-screenshot.png` (the artifacts).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-70 row
  COMPLETE with the evidence.
- `docs/port-blueprint/12-migration-roadmap.md` — the Phase-6 entry GREEN
  note.
- `README.md` — the SESSION-033 status entry.
- `webui/README.md` — rewritten from T-00's placeholder to the current
  truth (the shell, the module laws, the recipe, the retirement law).
- `internal/webfetch/fakes_test.go` — the TOCTOU fix (the publication point
  moved before the response write; assertions unchanged).

## Blockers & decisions needed

- None technical. The Windows runtime legs (the exe on a real box; the
  silent Evergreen bootstrapper path on a WebView2-less VM) remain the
  owner's Phase-7 checklist items — unchanged from ADR-0001 Am.1/Am.2's
  register.
- The production frontend tooling question (vanilla vs vite for T-72+) is
  OPEN for the owner — both work under wails v2; the PoC rides vanilla.

## Owner actions requested

1. Optional: the go-directive acknowledgement (ADR-0002 Am.2) still waits —
   twenty-fourth session carrying it; informational only.
2. Optional: pre-decide the T-72+ frontend tooling (vanilla vs vite) —
   SESSION-034 can proceed either way; the question only affects the
   production frontend's build step.

## Handoff

- Next session: SESSION-034 = **T-71 (L): internal/app** — the commands +
  the typed event bus + the ask-gates + lifecycle/locks, per doc 04 §2/§5
  (the command contract tests + the lock/watchdog tests) — with T-72 only
  if capacity remains after a proper T-71.
- NEXT-SESSION.md rewritten: yes (v34).
- The next prompt: (verbatim block in NEXT-SESSION.md §5).
