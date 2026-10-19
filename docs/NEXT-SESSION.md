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
- **Phase 6 OPEN (SESSION-033–034)** — **THE ENTRY GATE IS GREEN: T-70
  COMPLETE** (SESSION-033, `86790fb` → `8939cc4` → `7f613dd`: the themed
  Wails window + the coalesced log desk + ONE REAL JOB — the note-state
  baseline over an 800-note fixture vault; first frame 266/679 ms over two
  runs, the 10k flood 458,695/454,133 lines/s with 0 drops; the verdict at
  ADR-0001 Am.3 + `docs/poc-artifacts/t-70-gate.md`), and now **T-71
  COMPLETE: `internal/app` IS REAL** (SESSION-034, commits `5453214` →
  `00f4cdb` → `eb74171` → `74ae634`): the composition layer in four
  slices — (a) the typed fan-out Bus + the three ask-gates with the
  reference's IDENTICAL timeouts (300/300/600 s, the headless + no-reader
  fail-fasts); (b) the command surface over injected seams (doc 04 §5's
  table) with the crash-safety wrapper + the OpLock (the TelegramLockManager
  port) + the one-batch law; (c) the single-instance app.lock with PID
  healing, the watchdog, the startup passes at the reference's delays, the
  shutdown graces; (d) the DI root wiring every production seam (the batch
  composition with LLMDecision → the 600 s gate, the queue classification
  with the settlement at the door, the scan job with BOTH confirm doors, the
  backup zip's laws). doc 04 §2's landed-shape note records every delta the
  port's laws forced. **T-72 was NOT started** (closed at T-71's natural
  boundary — the mandate's own law). **1999 main-battery tests race-clean
  over 36 packages (67 in internal/app) + 12 webui tests** (deskbus +
  fixture, run from the nested module) at the close.
- **Push state: CLEAN** — SESSION-034 pushed after every task (`5453214` →
  `00f4cdb` → `eb74171` → `74ae634` + the session close); verify `git
  status` is in sync at the start (the tip at or beyond `74ae634`).
- **The environment recipe** (GUI edition): work repo at
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
  `webui/README.md` + `docs/poc-artifacts/t-70-gate.md`; **DEFERRED at
  SESSION-034's rebuild — T-71 was pure Go; T-72's build legs MUST rebuild
  it first**), all per build/TOOLCHAIN.md Am.3; go.mod exactly
  modernc.org/sqlite v1.36.1 + bogdanfinn/tls-client v1.16.0 +
  golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 + the go directive
  at 1.24.1 (the tls-client floor, ADR-0002 Am.2; never bump). The webui
  module's own go.mod rides `go 1.25.0` (wails v2.16.0's floor —
  ADR-0002 Am.3).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE HERMETIC-SEAM LAW** (SESSION-034): every network-touching seam in
    a test environment must be FAKED AT CONSTRUCTION — a real fetch inside
    a shared test helper (the engine tests' real DownloadBanner) can
    outlive its own test via an escape hatch and record a late write into
    a LATER test's window (the dry-run golden's intermittent 13th entry);
    the runner's own doc always said "the tests inject the fake banner".
  - **THE RESUMED-CONTEXT RE-VERIFICATION LAW** (SESSION-034): the
    context-death continuation recipe extends to CHAT-context deaths —
    diagnose the tree from evidence (git status + log + file mtimes), then
    re-verify preserved drafts FROM SCRATCH (the full battery + the hygiene
    scans) before committing; a preserved draft is continued and audited,
    never blindly trusted.
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

## 2. Next session: SESSION-035 — T-72 (L): the frontend foundation — the theme-token CSS golden + the production log desk

- **Verify the push state first** (§1): SESSION-034 pushed everything; the
  tree ends clean; confirm the tip is at or beyond `74ae634`.
- **FIRST the webkit tree rebuild** (deferred at SESSION-034 — T-72's build
  legs need it; the committed recipe: `webui/README.md` +
  `docs/poc-artifacts/t-70-gate.md`; budget for it BEFORE planning slices).
- **ASK THE OWNER the vanilla-vs-vite question EARLY** (both work under
  wails v2; the PoC rides vanilla; the choice needs an ADR either way —
  the theme-token golden itself is toolchain-independent, so land it first
  regardless).
- **T-72 per doc 13's row + spec-gui-theme-icons + spec-gui-ui** (read them
  FIRST — the authority): theme tokens CSS (the token table → the CSS var
  dump), icons, fonts, the layout shell, and the log desk with coalescing
  at PRODUCTION scale (deskbus is the Go half's starting pattern; the
  cache-mirror desk the frontend half's). **Acceptance: the token golden
  (CSS var dump) + the log perf test.**
- **`guibridge/` RETIRES at T-72** (delete it + its tests when the
  production shell replaces its wiring — not before); `webui/` dissolves
  into the main module and binds `internal/app` DIRECTLY per doc 04 §2's
  landed-shape note — the batch's CLI-parity shell wiring (warmup family,
  inputs fetchers, websites phase, Phase-5 CLEAR) lands on the app surface
  at that binding. The module-boundary moves need ADR-0002's next
  amendment — probe before claiming any boundary rule (Am.4's law).

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-034.md` — the most recent report (T-71 COMPLETE;
  the four slices; the context-death continuation; the hermetic banner
  stub's root cause).
2. `docs/port-blueprint/module-specs/spec-gui-theme-icons.md` +
   `spec-gui-ui.md` — T-72's authority (the tokens, the desk contract).
3. `internal/app/` — the landed composition layer (doc.go's file map is
  the tour; commands.go's Deps are the seams the frontend binds).
4. `webui/deskbus/deskbus.go` + `webui/frontend/` — the coalescing bus +
  the cache-mirror desk (the production log desk grows from these shapes).
5. `docs/port-blueprint/04-target-go-architecture.md` §2 (the
  landed-shape note — what T-72's binding owes) and §5.
6. `docs/port-blueprint/13-implementation-task-list.md` — the T-72 row +
  Phase 6's full book.

## 4. Known constraints / owner actions for SESSION-035

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for the
  PAT the moment a push is due (never stored, never written).
- **The vanilla-vs-vite call is the owner's** — ask early; the ADR records
  the choice either way.
- **The authorized runtime set is UNCHANGED** (the four; the wails require
  stays in the nested webui module per ADR-0002 Am.3 until T-72's own
  amendment dissolves the boundary — anything new needs an ADR).
- **The go-directive acknowledgement (STILL WAITING — twenty-fifth session
  carrying it)**: the main module rides `go 1.24.1`, forced by the
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

## 5. THE PROMPT TO COPY (SESSION-035)

```
You are starting SESSION-035 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-nine wipe events
   across SESSION-005..034 — SESSION-034 alone absorbed the twenty-ninth
   full wipe BEFORE work: both repos, the toolchain, the venv, and the module
   cache all gone, everything rebuilt from the committed recipes; SESSION-034
   also survived a mid-session CHAT-CONTEXT death without losing work — the
   continuation recipe is PROVEN at both levels: git status + git log
   --oneline -5 + file timestamps tell what ACTUALLY survived; a preserved
   commit or in-flight draft is continued and AUDITED — never blindly
   trusted, never restarted from zero; a resumed context re-verifies from
   scratch (the full battery) before committing; a wipe event may flip
   tracked files' MODES 644->755 while preserving content — diagnose via
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
   — build-time only), and THE ROOTLESS WEBKIT TREE (~/.local/lib-wails —
   DEFERRED at SESSION-034's rebuild, so it is likely ABSENT: T-72's FIRST
   act is rebuilding it per the committed recipe in webui/README.md +
   docs/poc-artifacts/t-70-gate.md) per build/TOOLCHAIN.md Am.3, and go.mod
   with modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the go
   directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not
   bump any of these; the webui module's own go.mod rides go 1.25.0 — wails
   v2.16.0's floor — and holds the wails require, per ADR-0002 Am.3). Then
   check `git status` and `git log --oneline -3`: SESSION-034 landed T-71
   COMPLETE (internal/app is real — the typed bus + the ask-gates with the
   reference's identical 300/300/600 s timeouts, the command surface with
   the crash-safety wrapper + the one-batch law, the single-instance lock +
   the watchdog + the startup passes + the shutdown graces, the DI root
   wiring every production seam; doc 04 §2's landed-shape note records the
   deltas) and PUSHED everything after every task — 1999 main-battery tests
   race-clean over 36 packages (67 in internal/app) + 12 webui tests.
   Verify git status shows up to date with origin/main (the tip at or
   beyond 74ae634). Use the PAT only ephemerally: git -c http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off.

2. Read, in order: the repo README.md (Phase 6 OPEN — T-71 at 74ae634) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-034's close: T-71 COMPLETE, the newest
   laws — the HERMETIC-SEAM law, the RESUMED-CONTEXT RE-VERIFICATION law) →
   docs/sessions/SESSION-034.md (the most recent report) → the T-72 specs
   (spec-gui-theme-icons.md + spec-gui-ui.md) → internal/app/doc.go (the
   file map) + commands.go (the Deps seams the frontend binds).

3. THIS SESSION IS T-72 (L): THE FRONTEND FOUNDATION per doc 13's row +
   spec-gui-theme-icons + spec-gui-ui (READ THEM FIRST — the authority).
   Order of work: (i) the webkit tree rebuild FIRST (the committed recipe;
   budget for it before planning slices); (ii) ASK THE OWNER the
   vanilla-vs-vite question EARLY (both work under wails v2; the PoC rides
   vanilla; an ADR records the choice either way — the theme-token golden
   is toolchain-independent, land it first regardless); (iii) the slices in
   order, each committed+pushed: (a) the theme-token CSS golden (the token
   table → the CSS var dump, golden-tested); (b) the production log desk
   with coalescing (deskbus the Go half's starting pattern, the cache-mirror
   desk the frontend half's) + the layout shell + icons + fonts. ACCEPTANCE:
   the token golden (CSS var dump) + the log perf test. guibridge/ RETIRES
   at T-72 (delete it + its tests when the production shell replaces its
   wiring — not before); webui/ dissolves into the main module and binds
   internal/app DIRECTLY per doc 04 §2's landed-shape note — the batch's
   CLI-parity shell wiring (warmup family, inputs fetchers, websites phase,
   Phase-5 CLEAR) lands on the app surface at that binding; the
   module-boundary moves need ADR-0002's next amendment (probe before
   claiming any boundary rule — Am.4's law). If capacity runs short: prefer
   closing cleanly at a natural slice boundary; one slice done properly
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
   probe before claiming a module-boundary rule (ADR-0002 Am.4); fake every
   network-touching seam in tests at construction (the HERMETIC-SEAM law —
   a real fetch can outlive its test and write into a later test's window).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-72's
   own acceptance: the token golden + the log perf test — and the webkit
   rebuild's proof), commit + push everything (ask me for the GitHub PAT the
   moment a push is due — never store it anywhere; never end the session
   push-pending), write docs/sessions/SESSION-035.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-036.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
