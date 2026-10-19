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
  (all 31 generators × two passes GREEN — now 33 with theme+icons).
- **Phase 6 OPEN (SESSION-033–035)** — T-70 COMPLETE (the entry gate,
  `86790fb` → `8939cc4` → `7f613dd`), T-71 COMPLETE (`internal/app`,
  SESSION-034, `5453214` → `00f4cdb` → `eb74171` → `74ae634`), and now
  **T-72 COMPLETE: THE FRONTEND FOUNDATION** (SESSION-035, commits
  `cc89cb0` + `2502e7a` (the recovery) → `88e5c6b` (the theme-token CSS
  golden + the icon-pack golden) → `16897b5` (the production shell)):
  the theme tables + the icon pack locked against TWO
  reference-generated goldens (three twins: Go tables, the CSS var
  dump, the JS pack); the production shell binds internal/app DIRECTLY
  (the typed Bus → the deskbus's 120 ms/≤250 coalescing → the
  cache-mirror desk; the startup passes' real lines; the
  single-instance lock with the new AppLock.Refused law); guibridge/
  RETIRED (deleted with its tests) and the T-70 gate legs retired with
  it. The dissolution probe ran FIRST per Am.4's law: wails into the
  MAIN module forces x/net v0.56.0 + x/text v0.39.0 + the 1.25.0
  directive — BLOCKED by the authorized-pins law, fully reverted; the
  boundary stays nested, the owner's call in **ADR-0002 Am.5**. The
  acceptance GREEN: the token golden + the log perf test (the gate
  leg: first frame 283 ms; the 10k flood through the FULL production
  chain 434,782 lines/s producer-side, 0 drops, 22 ms max handler, all
  10,001 rows rendered). **1994 main-battery tests race-clean over 35
  packages + 21 webui tests over 4 packages (12 + 6 uitheme + 3
  shell).**
- **Push state: CLEAN** — SESSION-035 pushed after every task; verify
  `git status` is in sync at the start (the tip at or beyond
  `16897b5`).
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (symlinked
  /home/z/second-brain-project), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (symlinked /home/z/GitCurator;
  must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars, verified against go.dev's published JSON), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only), the wails CLI
  v2.16.0 (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails: apt
  --download-only + dpkg -x of libwebkit2gtk-4.1-dev + libgtk-3-dev,
  the .pc prefix rewrites, the dev-symlink relinking — this session
  needed 26 relinks to the system runtime copies because the fresh
  base image ships the gtk/atk runtimes system-wide; the full recipe
  in `webui/README.md` + `docs/poc-artifacts/t-70-gate.md`), all per
  build/TOOLCHAIN.md Am.3; the MAIN go.mod exactly modernc.org/sqlite
  v1.36.1 + bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0
  + golang.org/x/text v0.32.0 + the go directive at 1.24.1 (never
  bump; ADR-0002 Am.2). The NESTED webui module rides go 1.25.0 (wails
  v2.16.0's floor) and now ALSO carries the transitive requires the
  direct internal/app import pulls (indirect via the parent replace —
  ADR-0002 Am.5; the MAIN go.mod stays byte-unchanged).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE DISSOLUTION-PROBE RECORD** (SESSION-035, ADR-0002 Am.5): the
    mandate's module-boundary move was probed BEFORE any claim; the
    probe BLOCKED it with hash-level evidence (wails' own go.mod) and
    the session landed the rest of the mandate on the nested shape —
    probe, record, never improvise past a blocker.
  - **THE GOLDEN-MARSHAL-ORDER TRAP** (SESSION-035): Go maps marshal
    keys SORTED; structs emit DECLARATION order — a golden that
    byte-diffs must marshal a MAP (or the golden's own order), or the
    diff fails with identical data and no per-field summary.
  - **THE SUITE-REVERIFICATION DIVIDEND** (SESSION-035): re-verifying
    the suite after a wipe caught a REAL production race (the
    post-Stop push coin flip) that six prior green runs had ridden at
    P≈1.5% — the re-verification law pays in bugs, not just safety.
  - **THE WAILS-BUILD MODE FLIPS** (SESSION-035): every `wails build`
    re-flips the wailsjs bindings' modes to 755 (and a stray `go
    build .` in webui/ leaves a 22 MB binary named `webui` — never
    commit it); check `git diff --raw` after every build, restore via
    chmod off the parsed list, never git checkout.
  - **THE HERMETIC-SEAM LAW** (SESSION-034): every network-touching
    seam in a test environment must be FAKED AT CONSTRUCTION — a real
    fetch inside a shared test helper can outlive its test via an
    escape hatch and record a late write into a LATER test's window.
  - **THE RESUMED-CONTEXT RE-VERIFICATION LAW** (SESSION-034): a
    resumed context re-verifies preserved drafts FROM SCRATCH (the
    full battery + the hygiene scans) before committing.
  - **THE MODE-FLIP DEGRADATION LAW** (SESSION-033): a wipe event may
    flip tracked files' modes (644→755) while preserving CONTENT —
    diagnose with `git diff --raw`, restore with chmod off the parsed
    list; NEVER `git checkout`.
  - **THE MODULE-BOUNDARY PROBE LAW** (SESSION-033, ADR-0002 Am.4):
    any internal-rule claim at a nested-module boundary must be PROBED
    before it lands in a record.
  - **THE FRONTEND-MODULE LESSON** (SESSION-033): an ES-module frontend
    needs `<script type="module">` AND explicit `.js` extensions in
    import specifiers.
  - **THE TIMER-CLAMP INSTRUMENT CAVEAT** (SESSION-033): under a
    WM-less Xvfb, WebKit clamps background-page timers to ~1 Hz —
    interactivity evidence must come from handler-time instruments +
    the rendered output.
  - **THE TOOLCHAIN-ROUND-2 PROOF** (SESSION-033) + the ROUND-3 proof
    (SESSION-035 — the webkit tree rebuilt from the committed recipe a
    third time, incl. the 26 dev-symlink relinks).
  - THE T-65 ZIP-COMPOSITION RESOLUTION; THE CORPUS-REGENERATION LAW;
    THE DISPATCH-COMPLETION LAW; THE STATE-CHAIN + MUTATION-TEETH
    laws; THE CONTEXT-DEATH CONTINUATION recipe; THE DSN-EMPTY-PATH
    HAZARD; THE FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the
    fake-clock arithmetic; the exception class-name tokens; the
    stdout-capture-helper limit.
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd; never commit outside the work repo; never checkout blindly);
    **THE RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F
    FILE); **THE WRITE-TOOL REDACTION lesson** (byte-level secret-scan
    before every commit); **THE BACKGROUND-PROCESS lesson** (long
    suites in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for
    the PAT the moment a push is due; never store it).

## 2. Next session: SESSION-036 — T-73 (L): Settings pages A — Vault, Credentials, Proxy, LLM, Input, Sound

- **Verify the push state first** (§1): confirm the tip is at or beyond
  `16897b5`.
- **T-73 per doc 13's row + spec-gui-pages-a** (READ IT FIRST — the
  authority): the six Settings pages on the production shell
  (webui/), driving internal/app's command surface; the save-on-change
  law (editingFinished/toggled → the page savers); the golden-locked
  theme tables + the 28-glyph icon pack are ready (the Settings
  sidebar's section glyphs are in the pack). Acceptance: page workflow
  tests + the save-law (merge/in-place) tests.
- **The shell's wiring debt** (honest, recorded): the CTA buttons are
  disabled until T-77's hero machine; the CLI-parity shell wiring
  (warmup family, inputs fetchers, websites phase, Phase-5 CLEAR)
  lands on the app surface at the T-73+ binding, exactly as doc 04
  §2's landed-shape note says.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-035.md` — the most recent report (T-72
  COMPLETE; the dissolution probe; the race fix).
2. `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — T-73's
  authority (the six pages, the save laws).
3. `webui/` — the production shell (README.md's map; shell.go's Bus
  subscriptions are the transport the pages ride; frontend/ the
  vanilla twin: theme.css + icons.js + the desk).
4. `internal/app/` — commands.go's Deps (the seams the pages drive)
  + app.go's DI root.
5. `docs/port-blueprint/04-target-go-architecture.md` §2 (the
  landed-shape note) and §5.
6. `docs/port-blueprint/13-implementation-task-list.md` — the T-73
  row + Phase 6's book.

## 4. Known constraints / owner actions for SESSION-036

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The dissolution call (the owner's, ADR-0002 Am.5)**: (a) amend the
  authorized set vs (b) keep the nested boundary — no rush; T-73
  proceeds on (b) either way.
- **The vanilla-vs-vite call (the owner's)**: vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- **The go-directive acknowledgement (STILL WAITING — twenty-sixth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker**.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755) and never commit a stray `webui/webui`
  binary (`go build .` in webui/ leaves it).
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs + the launchers' clean-VM smoke + the
  bootstrapper leg (Phase 7 checklist, doc 10 §7).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB any persisted
  extraheader — the GIT_CONFIG_COUNT env-var form avoids it entirely).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit); tests never weakened; reality wins over specs
  (fix the spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-036)

```
You are starting SESSION-036 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (thirty wipe events across
   SESSION-005..035 — SESSION-035 absorbed the thirtieth full wipe BEFORE work:
   both repos, the toolchain, the venv, the module cache, AND the rootless
   webkit tree all gone, everything rebuilt from the committed recipes — the
   third time the webkit recipe rebuilt the tree, this round including 26
   dev-symlink relinks to the system runtime copies because the fresh base
   image ships the gtk/atk runtimes system-wide; the continuation recipe is
   PROVEN at all levels: git status + git log --oneline -5 + file timestamps
   tell what ACTUALLY survived; a preserved commit or in-flight draft is
   continued and AUDITED — never blindly trusted, never restarted from zero; a
   resumed context re-verifies from scratch (the full battery + the hygiene
   scans) before committing; a wipe event may flip tracked files' MODES
   644->755 while preserving content — diagnose via 'git diff --raw' and
   restore via chmod off the parsed list, NEVER git checkout; and the
   SESSION-035 addendum: after every `wails build` the wailsjs bindings
   re-flip to 755, and a stray `go build .` in webui/ leaves a 22 MB binary
   named `webui` — never commit it). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-project),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON), venv /home/z/venvs/gitcurator
   (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
   — build-time only), and the rootless webkit tree (~/.local/lib-wails per
   webui/README.md + docs/poc-artifacts/t-70-gate.md) per build/TOOLCHAIN.md
   Am.3. The MAIN go.mod: modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client
   at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and
   the go directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do
   not bump any of these). The NESTED webui module rides go 1.25.0 (wails
   v2.16.0's floor) + the wails require + the transitive requires the direct
   internal/app import pulls (indirect via the parent replace — ADR-0002 Am.5);
   the MAIN go.mod stays byte-unchanged. Then check `git status` and
   `git log --oneline -3`: SESSION-035 landed T-72 COMPLETE (the frontend
   foundation: the theme-token CSS golden + the icon-pack golden — three twins
   locked against two reference-generated goldens; the production shell binding
   internal/app DIRECTLY — the typed Bus -> the deskbus's 120ms/<=250
   coalescing -> the cache-mirror desk; the startup passes' real lines; the
   single-instance lock with the new AppLock.Refused law; guibridge/ RETIRED
   with its tests) and PUSHED everything after every task — 1994 main-battery
   tests race-clean over 35 packages + 21 webui tests over 4. Verify git
   status shows up to date with origin/main (the tip at or beyond 16897b5).
   Use the PAT only ephemerally: git -c http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off — the law pays in
   bugs, not just safety (SESSION-035's re-verification caught a REAL
   production race: the deskbus post-Stop push coin flip).

2. Read, in order: the repo README.md (Phase 6 OPEN — T-72 at 16897b5) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-035's close: T-72 COMPLETE, the newest
   laws — the DISSOLUTION-PROBE record, the GOLDEN-MARSHAL-ORDER trap, the
   SUITE-REVERIFICATION dividend, the WAILS-BUILD MODE FLIPS) →
   docs/sessions/SESSION-035.md (the most recent report) → the T-73 spec
   (spec-gui-pages-a.md) → webui/ (the production shell: README.md's map +
   shell.go's Bus subscriptions) + internal/app/commands.go (the Deps seams
   the pages drive).

3. THIS SESSION IS T-73 (L): SETTINGS PAGES A per doc 13's row +
   spec-gui-pages-a (READ IT FIRST — the authority): Vault, Credentials,
   Proxy, LLM, Input, Sound — the six pages on the production shell (webui/),
   driving internal/app's command surface; the save-on-change law
   (editingFinished/toggled → the page savers); the golden-locked theme
   tables + the 28-glyph icon pack are ready (the Settings sidebar's section
   glyphs are in the pack). ACCEPTANCE: page workflow tests + the save-law
   (merge/in-place) tests. The CLI-parity shell wiring (warmup family,
   inputs fetchers, websites phase, Phase-5 CLEAR) lands on the app surface
   at this binding. If capacity runs short: prefer closing cleanly at a
   natural slice boundary; one slice done properly with proof beats a
   rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git (byte-level
   secret-scan before every commit; never write token-shaped strings even as
   fakes); no new dependencies without an ADR (the authorized runtime set is
   exactly modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0,
   x/text v0.32.0 — do not bump any; the MAIN go.mod stays byte-unchanged;
   the wails require + the indirect transitive requires stay in the NESTED
   webui module per ADR-0002 Am.5); surface blockers with evidence
   (hash-level, not display) instead of improvising past them; probe before
   claiming any module-boundary rule (ADR-0002 Am.4); fake every
   network-touching seam in tests at construction (the HERMETIC-SEAM law);
   after every wails build, check `git diff --raw` for mode flips and never
   commit a stray webui/webui binary.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-73's
   own acceptance: the page workflow tests + the save-law tests), commit +
   push everything (ask me for the GitHub PAT the moment a push is due —
   never store it anywhere; never end the session push-pending), write
   docs/sessions/SESSION-036.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus
   the prompt I should copy to start SESSION-037.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
