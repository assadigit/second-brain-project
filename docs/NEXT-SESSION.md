# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker HTTP
  API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity.
- **Phase 2 CLOSED** (SESSION-006–010): T-20…T-25.
- **Phase 3 CLOSED** (SESSION-011–014): T-30…T-34.
- **Phase 4 CLOSED** (SESSION-015–026): T-40…T-50 + the Phase-4 closer.
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity release
  `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip asset
  sha-verified + round-trip verified); the differential corpus full run
  (all 31 generators × two passes GREEN).
- **Phase 6 OPEN (SESSION-033)** — **THE ENTRY GATE IS GREEN: T-70
  COMPLETE** (commits `86790fb` → `937797a` → `8939cc4` → `7f613dd`): the
  PoC C GUI-shell gate re-run with ONE REAL JOB wired through — the
  note-state baseline (record_baseline_if_empty's composition) over a
  deterministic 800-note fixture vault via the `guibridge/` interim seam.
  The shell at `webui/` (a NESTED Go module per ADR-0002 Am.3: require
  wails v2.16.0 + replace to the parent — the main module's runtime set
  stays the authorized four): the themed window (the spec-gui-theme-icons
  LIGHT tokens), THE COALESCED LOG full contract (the deskbus Go half +
  the cache-mirror frontend half), the vanilla frontend (NO npm step), the
  committed rootless-webkit recipe (webui/build-support/). THE MEASURED
  VERDICT over two runs: first frame 266/679 ms (gate ≤ 1.5 s), the 10k
  flood 458,695/454,133 lines/s with 0 drops and 23–25 ms max handler time
  per batch, the real job 800 notes in 94–101 ms, the Windows exe CGO-free
  at 14.95 MB; the desk's cap law VLM-verified off the screenshot.
  Evidence: `docs/poc-artifacts/t-70-gate.md`; the verdict: ADR-0001 Am.3.
  **T-71 was NOT started** (capacity ran short after two environment
  rebuilds — the mandate's own law: one slice done properly beats a rushed
  whole). **1932 main-battery tests race-clean over 35 packages (+5
  guibridge) + 12 webui tests** (deskbus + fixture, run from the nested
  module) at the close.
- **Push state: CLEAN** — SESSION-033 pushed after every task
  (`5a688a2` → `86790fb` → `937797a` → `8939cc4` → `7f613dd`); verify
  `git status` is in sync at the start (the tip at or beyond `7f613dd`).
- **The environment recipe** (GUI edition — T-70 grew it): work repo at
  /home/z/my-project/second-brain-project (symlinked
  /home/z/second-brain-project), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (symlinked /home/z/GitCurator;
  must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` —
  64 hex chars, verified against go.dev's published JSON), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only), **the wails CLI
  v2.16.0** (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only), and
  **the rootless webkit tree** (~/.local/lib-wails: apt --download-only +
  dpkg -x of libwebkit2gtk-4.1-dev + libgtk-3-dev, the .pc prefix
  rewrites, the dev-symlink relinking — the full recipe is committed in
  `webui/README.md` + `docs/poc-artifacts/t-70-gate.md`), all per
  build/TOOLCHAIN.md Am.3; go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (the
  tls-client floor, ADR-0002 Am.2; never bump). The webui module's own
  go.mod rides `go 1.25.0` (wails v2.16.0's floor — ADR-0002 Am.3).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE MODE-FLIP DEGRADATION LAW** (SESSION-033): a wipe event may flip
    tracked files' modes (644→755) while preserving CONTENT — diagnose
    with `git diff --raw` (mode-only rows: `:100644 100755 … M\t<path>`),
    restore with chmod 644 off the parsed list; NEVER `git checkout` (the
    CD-DISCIPLINE law's sibling — restore only what you mean to restore).
  - **THE MODULE-BOUNDARY PROBE LAW** (SESSION-033, ADR-0002 Am.4): any
    internal-rule claim at a nested-module boundary must be PROBED (the
    ten-line boundary_probe.go: build + vet exit codes) before it lands in
    a record — the rule's edge at local-replace nested modules bites in
    the OTHER direction than intuition says.
  - **THE FRONTEND-MODULE LESSON** (SESSION-033): an ES-module frontend
    file needs `<script type="module">` AND explicit `.js` extensions in
    import specifiers (wails' asset server does not resolve
    extensionless paths); a classic script tag silently fails on `import`
    syntax — the page renders static but never wires up.
  - **THE TIMER-CLAMP INSTRUMENT CAVEAT** (SESSION-033): under a WM-less
    Xvfb, WebKit clamps background-page timers to ~1 Hz — task-gap probes
    read the clamp, not event-loop blockage; interactivity evidence must
    come from handler-time instruments (the event's own work) + the
    rendered output.
  - **THE TOOLCHAIN-ROUND-2 PROOF** (SESSION-033): the committed recipe
    (TOOLCHAIN.md Am.3 + webui/README.md + t-70-gate.md) rebuilds the
    whole GUI toolchain from a clean tree — proven by the gate's
    post-degradation run 2 landing GREEN.
  - **THE T-65 ZIP-COMPOSITION RESOLUTION** (SESSION-032): the minimal
    deterministic mkzip slice landed at T-65; T-80/Phase 7 keeps the FULL
    builder per F4 — do not re-litigate.
  - **THE CORPUS-REGENERATION LAW** (SESSION-032, T-64): a golden that
    embeds per-run bytes can never regenerate byte-identical — fold the
    roots at payload end, design order-independent fixtures, mask
    documented per-run facts ONLY in the corpus driver's digest.
  - **THE DISPATCH-COMPLETION LAW** (SESSION-031): a command is only
    LANDED when Run's dispatch routes the argv to it.
  - **THE STATE-CHAIN LAW** (SESSION-030) + **THE MUTATION-TEETH LAW**
    (SESSION-030) + **THE CONTEXT-DEATH CONTINUATION recipe** (a full tool
    outage may just need waiting out — probe periodically, continue where
    you stopped).
  - **THE DSN-EMPTY-PATH HAZARD** (never "" as a db path); **THE
    FIXED-LENGTH ROOTS** law; **THE SHALLOW-COPY SNAPSHOT** law; the
    fake-clock arithmetic; exception class-name tokens; the
    stdout-capture-helper limit (a section capturing a RETURN VALUE cannot
    use it).
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own cd;
    never commit outside the work repo; never checkout blindly); **THE
    RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F FILE,
    verify with `git log --format=%B`); **THE WRITE-TOOL REDACTION
    lesson** (craft safe-shaped test tokens; byte-level secret-scan before
    every commit); **THE BACKGROUND-PROCESS lesson** (long suites in
    FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for the PAT the
    moment a push is due; never store it).

## 2. Next session: SESSION-034 — T-71 (L): `internal/app` — the commands + the typed event bus + the ask-gates + lifecycle/locks

- **Verify the push state first** (§1): SESSION-033 pushed everything; the
  tree ends clean; confirm the tip is at or beyond `7f613dd`.
- **T-71 per doc 04 §2/§5** (read them FIRST — the authority): the
  `internal/app` composition layer — the App struct (config, stores, the
  event bus, the injected clients), the typed event bus (`LogLine`,
  `Progress`, `Status`, `BatchFinished`, the ask-gates `AskCode`/
  `AskScanConfirm`/`LLMFailed` as channels with the reference's identical
  timeouts — 300 s code, 300 s scan, 600 s LLM), the command surface (doc
  04 §5's table: Sync.FetchQueue, Sync.RunBatch, Sync.FastSync,
  Sync.DriftSeal, Scan.Run, Scan.ApplyPlan, Backup.*, Radar.Detect), the
  one-batch-at-a-time mutex + running flag (the TelegramLockManager law),
  the single-instance lock (`app.lock` + PID liveness healing,
  gui/app.py:393-440), the startup passes' sequential goroutines with the
  reference's delays, and the shutdown grace windows. **Acceptance: the
  command contract tests + the lock/watchdog tests.** The `guibridge/`
  seam (doc 04 §2's event-seam shape at PoC scale) is the starting
  pattern — internal/app is its production evolution (and guibridge
  RETIRES at T-72).
- **Then, only if capacity remains after a PROPER T-71**, T-72 per the
  roadmap (the frontend foundation: the theme-token CSS golden + the
  production log desk) — but prefer closing cleanly at T-71's natural
  boundary; one slice done properly with proof beats a rushed whole.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-033.md` — the most recent report (T-70 COMPLETE;
  the gate's evidence; the two environment rebuilds).
2. `docs/port-blueprint/04-target-go-architecture.md` §2 (the app layer +
  the event bus + the ask-gates) and §5 (the command surface table) —
  T-71's authority.
3. `docs/port-blueprint/module-specs/spec-gui-processing.md` +
  `spec-gui-processing-control.md` + `spec-gui-worker-jobs.md` — the
  Python origins of every command (the 8 queued signals, the Event
  round-trips, the timeouts).
4. `guibridge/bridge.go` + `webui/deskbus/deskbus.go` — the landed seam
  pattern + the coalescing bus (T-71's event bus grows from these shapes).
5. `docs/port-blueprint/13-implementation-task-list.md` — the T-71 row +
  Phase 6's full book.

## 4. Known constraints / owner actions for SESSION-034

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for the
  PAT the moment a push is due (never stored, never written).
- **The authorized runtime set is UNCHANGED** (the four; ADR-0002 Am.3's
  nested-module strategy holds — anything new needs an ADR).
- **The T-72+ production frontend tooling question (vanilla vs vite) is
  OPEN for the owner** — both work under wails v2; the PoC rides vanilla;
  SESSION-034 does not need the answer (T-71 is pure Go).
- **The go-directive acknowledgement (STILL WAITING — twenty-fourth
  session carrying it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker**.
- **The release feedback loop**: the owner may report v0.92.0-go.1 or the
  T-70 shell experiences — record them in the session report; runtime
  findings route to Phase 7's checklist or immediate fixes per severity.
- **Deferred test debt on the books (evidence-recorded, not blockers)**:
  test_noteredo's remaining un-ported selections; the note_state SURGERY
  path's CRLF round-trip differential (Phase 4+); the Windows live legs +
  the launchers' clean-VM smoke + the bootstrapper leg (Phase 7 checklist,
  doc 10 §7).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB any persisted
  extraheader from .git/config after the clone — the
  GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit); tests never weakened; reality wins over specs
  (fix the spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-034)

```
You are starting SESSION-034 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-eight wipe events
   across SESSION-005..033 — SESSION-033 alone absorbed the twenty-seventh
   full wipe BEFORE work plus a mid-session degradation that took the toolchain,
   the module cache, the webkit tree, and /tmp while BOTH repos and all
   in-flight work survived; the continuation recipe is PROVEN: git status +
   git log --oneline -5 + file timestamps tell what ACTUALLY survived; a
   preserved commit or in-flight draft is continued and AUDITED — never
   blindly trusted, never restarted from zero; a wipe event may flip tracked
   files' MODES 644->755 while preserving content — diagnose via
   'git diff --raw' mode-only rows and restore via chmod 644 off the parsed
   list, NEVER git checkout). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-project),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON), venv /home/z/venvs/gitcurator
   (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
   — build-time only) + the rootless webkit tree (~/.local/lib-wails; the full
   committed recipe is in webui/README.md + docs/poc-artifacts/t-70-gate.md)
   per build/TOOLCHAIN.md Am.3, and go.mod with modernc.org/sqlite at v1.36.1,
   bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text
   v0.32.0 direct, and the go directive at 1.24.1 (the tls-client-forced floor —
   ADR-0002 Am.2; do not bump any of these; the webui module's own go.mod rides
   go 1.25.0 — wails v2.16.0's floor — and holds the wails require, per
   ADR-0002 Am.3). Then check `git status` and `git log --oneline -3`:
   SESSION-033 landed T-70 COMPLETE (the Phase-6 entry gate GREEN: the themed
   Wails window + the coalesced log desk + ONE REAL JOB — the note-state
   baseline over an 800-note fixture vault via the guibridge/ seam; first frame
   266/679 ms over two runs, the 10k flood 458,695/454,133 lines/s with 0 drops
   and 23-25 ms max handler time per batch, the Windows exe CGO-free at
   14.95 MB; the verdict at ADR-0001 Am.3 + docs/poc-artifacts/t-70-gate.md;
   the toolchain record at ADR-0002 Am.3 + Am.4) and PUSHED everything after
   every task — 1932 main-battery tests race-clean over 35 packages + 12 webui
   tests. Verify git status shows up to date with origin/main (the tip at or
   beyond 7f613dd). Use the PAT only ephemerally: git -c http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off.

2. Read, in order: the repo README.md (Phase 6 OPEN — T-70 at 7f613dd) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-033's close: the gate GREEN, the newest
   laws — the MODE-FLIP DEGRADATION law, the MODULE-BOUNDARY PROBE law, the
   FRONTEND-MODULE lesson, the TIMER-CLAMP instrument caveat, the
   TOOLCHAIN-ROUND-2 proof) → docs/sessions/SESSION-033.md (the most recent
   report).

3. THIS SESSION IS T-71 (L): internal/app per doc 04 §2/§5 (READ THEM FIRST —
   the authority): the composition layer — the App struct (config, stores,
   the typed event bus, the injected clients), the event bus (LogLine,
   Progress, Status, BatchFinished, the ask-gates AskCode/AskScanConfirm/
   LLMFailed as channels with the reference's identical timeouts — 300 s
   code, 300 s scan, 600 s LLM; the 8 queued Qt signals' replacement), the
   command surface (doc 04 §5's table — Sync.FetchQueue, Sync.RunBatch,
   Sync.FastSync, Sync.DriftSeal, Scan.Run, Scan.ApplyPlan, Backup.SealVault/
   Resync/Now/Publish, Radar.Detect), one-batch-at-a-time (the mutex +
   running flag, the TelegramLockManager law), the single-instance lock
   (app.lock + PID liveness healing, gui/app.py:393-440), the startup
   passes' sequential goroutines with the reference's delays, the shutdown
   grace windows (batch 5 s, test workers 2 s). ACCEPTANCE: the command
   contract tests + the lock/watchdog tests. The landed guibridge/ seam
   (doc 04 §2's event-seam shape at PoC scale) + webui/deskbus (the
   coalescing bus) are the starting patterns — internal/app is their
   production evolution; guibridge RETIRES at T-72, not before. Slices in
   order, each committed+pushed: (a) events.go (the bus + the ask-gates)
   with its tests; (b) commands.go (the contract tests over fake deps per
   the hermetic law); (c) lifecycle.go (the lock + the watchdog + the
   startup passes) with the lock/watchdog tests; (d) app.go (the DI root)
   + the doc updates. If capacity remains after a PROPER T-71: T-72's
   first slice per the roadmap (the theme-token CSS golden) — but prefer
   closing cleanly at T-71's natural boundary. One slice done properly
   with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git (byte-level
   secret-scan before every commit; never write token-shaped strings even as
   fakes); no new dependencies without an ADR (the authorized runtime set is
   exactly modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0,
   x/text v0.32.0 — do not bump any; the wails require stays in the nested
   webui module per ADR-0002 Am.3 until T-72's own amendment; x/net/html
   stays OUT unless the owner approves an ADR); surface blockers with
   evidence (hash-level, not display) instead of improvising past them;
   probe before claiming a module-boundary rule (ADR-0002 Am.4).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-71's
   own acceptance: the command contract tests + the lock/watchdog tests —
   and T-72's slice's if reached), commit + push everything (ask me for the
   GitHub PAT the moment a push is due — never store it anywhere; never end
   the session push-pending), write docs/sessions/SESSION-034.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-035.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
