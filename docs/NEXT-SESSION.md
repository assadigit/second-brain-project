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
- **Phase 6 OPEN (SESSION-033–036)** — T-70 COMPLETE (the entry gate),
  T-71 COMPLETE (`internal/app`, SESSION-034), T-72 COMPLETE (the
  frontend foundation, SESSION-035), and now **T-73 COMPLETE: THE
  SETTINGS PAGES A** (SESSION-036, commits `595a362` (the
  internal/app settings surface — THE LIVE-CONFIG LAW: config.json
  loads ONCE at construction through the GUI load path and the app
  holds the same map all session, saves mutate it IN PLACE; the ONE
  merge save with all four v30 laws; the six pages' engines over
  injectable SettingsDeps — the token test as-typed, the About Me
  wizard, the Ollama refresh/spawn, the llama.cpp detect/refresh +
  THE SERVER CONFIG GRAB, the Detect & Set fast lane, the sound
  gate) → `48157da` (the window on the shell: settingsui.go the
  thin-shell wrapper + the two native pickers; frontend/settings.js —
  the gear door, the 190px sidebar with the six rows + the TAB_INFO
  hints, the six pages rendered from the snapshot with every label
  verbatim; THE SAVE-ON-CHANGE law; the mini-dialog kit; the fill
  line)). **The acceptance GREEN: the page workflow tests + the
  save-law (merge/in-place) tests — 2020 main-battery tests race-clean
  over 35 packages (1994 + 26 settings) + 23 webui tests over 4
  packages (21 + 2).** The gate leg re-run GREEN (first frame 963 ms;
  the 10k flood 526,315 lines/s, 0 drops); the Settings screenshot
  VLM-verified (all six rows, the quiet-truth line VERBATIM). The
  commit-cadence dated note landed in spec-gui-pages-a (the port
  commits all six pages per edit; the reference saves creds/proxy/LLM
  only at closeEvent — every reference-persisted value persists).
- **Push state: CLEAN** — SESSION-036 pushed after every task; verify
  `git status` is in sync at the start (the tip at or beyond
  `48157da`).
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
  the .pc prefix rewrites, the dev-symlink relinking — SESSION-036
  again needed exactly 26 relinks to the system runtime copies), all
  per build/TOOLCHAIN.md Am.3; the MAIN go.mod exactly
  modernc.org/sqlite v1.36.1 + bogdanfinn/tls-client v1.16.0 +
  golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 + the go
  directive at 1.24.1 (never bump; ADR-0002 Am.2). The NESTED webui
  module rides go 1.25.0 (wails v2.16.0's floor) + the wails require +
  the transitive requires the direct internal/app import pulls
  (indirect via the parent replace — ADR-0002 Am.5; the MAIN go.mod
  stays byte-unchanged).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE ONE-SLICE-PROPERLY RECORD** (SESSION-036): the CLI-parity
    shell wiring was scoped for the T-73 binding but NOT rushed —
    T-73's own acceptance landed complete with proof, and the wiring
    rides the next binding as the recorded FIRST item. Capacity
    honesty beats a rushed whole; the mandate's own closing law.
  - **THE LIVE-CONFIG LAW** (SESSION-036): the GUI app loads
    config.json ONCE through config.LoadAppConfig and holds the same
    *storage.OrderedMap all session (the reference's self.config);
    every save mutates it IN PLACE (clear+update, never a rebind);
    a failed write leaves the live map untouched. New pages build on
    this — never re-read per command.
  - **THE GROUP-POINTER LAW** (SESSION-036): a page's save group
    carries ONLY its own keys (the hasattr pattern) — an absent group
    keeps every config value; the merge never whitelist-rebuilds.
  - **THE DISSOLUTION-PROBE RECORD** (SESSION-035, ADR-0002 Am.5):
    probe, record, never improvise past a blocker.
  - **THE GOLDEN-MARSHAL-ORDER TRAP** (SESSION-035): Go maps marshal
    keys SORTED; structs emit DECLARATION order.
  - **THE SUITE-REVERIFICATION DIVIDEND** (SESSION-035): re-verifying
    the suite after a wipe caught a REAL production race — the law
    pays in bugs, not just safety.
  - **THE WAILS-BUILD MODE FLIPS** (SESSION-035): every `wails build`
    re-flips the wailsjs bindings to 755; a stray `go build .` in
    webui/ leaves a 22 MB `webui` binary — check `git diff --raw`
    after every build, restore via chmod, never commit the binary.
  - **THE HERMETIC-SEAM LAW** (SESSION-034); **THE
    RESUMED-CONTEXT RE-VERIFICATION LAW** (SESSION-034); **THE
    MODE-FLIP DEGRADATION LAW** (SESSION-033); **THE
    MODULE-BOUNDARY PROBE LAW** (SESSION-033, ADR-0002 Am.4); **THE
    FRONTEND-MODULE LESSON** (SESSION-033: ES modules need
    `<script type="module">` + explicit `.js` extensions); **THE
    TIMER-CLAMP INSTRUMENT CAVEAT** (SESSION-033); **THE
    TOOLCHAIN-ROUND-2/3 PROOF** (SESSION-033/035); THE T-65
    ZIP-COMPOSITION RESOLUTION; THE CORPUS-REGENERATION LAW; THE
    DISPATCH-COMPLETION LAW; THE STATE-CHAIN + MUTATION-TEETH laws;
    THE CONTEXT-DEATH CONTINUATION recipe; THE DSN-EMPTY-PATH HAZARD;
    THE FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the fake-clock
    arithmetic; the exception class-name tokens; the
    stdout-capture-helper limit.
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd — the sandbox's shell resets cwd between calls; never commit
    outside the work repo; never checkout blindly); **THE
    RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F
    FILE); **THE WRITE-TOOL REDACTION lesson** (byte-level secret-scan
    before every commit); **THE BACKGROUND-PROCESS lesson** (long
    suites in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for
    the PAT the moment a push is due; never store it).

## 2. Next session: SESSION-037 — T-74 (L): Settings pages B — Bot, Dashboard, Sources, Backup (+ radar, fast sync)

- **Verify the push state first** (§1): confirm the tip is at or beyond
  `48157da`.
- **THE SESSION'S FIRST SLICE — the carried wiring debt**: the
  CLI-parity shell wiring on the app surface (the warmup family, the
  inputs fetchers, the websites phase, the Phase-5 CLEAR — engine.go's
  runImpl composition re-hosted on internal/app with its own tests).
  It gates T-77's hero batch composition; land it as its own committed
  slice before or alongside the pages, and close the debt T-73
  recorded honestly.
- **T-74 per doc 13's row, the pages matched BY NAME to their spec
  files** (the R-0 dated note in doc 13): Bot (bot_queue.py — spec
  lives in spec-gui-pages-a.md, the Option-B ripples: the Worker
  pending queue is the data source, no Telethon), Dashboard
  (dashboard.py — spec-gui-pages-a.md), Sources + Backup +
  backup_seal.py (spec-gui-pages-b.md — drift seal + fast sync), + the
  radar/fast-sync doors. Acceptance: queue report rendering from a
  canned job result; the fast-sync laws (the ported pure halves are
  already tested).
- The Settings window (settingsui.go + settings.js) is the established
  host: the new pages join SECTIONS in the sidebar at the reference's
  own order (Dashboard after Input; Sync is T-75's; Bot, Sources,
  Backup next; the rows use the pack's glyphs bar-chart-3 / bot /
  satellite-dish / archive).

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-036.md` — the most recent report (T-73
  COMPLETE; the live-config law; the carried wiring debt).
2. `spec-gui-pages-a.md` (bot_queue.py + dashboard.py sections + the
  Option-B dated notes) — the authority for two of the pages.
3. `spec-gui-pages-b.md` (backup_seal.py — the drift seal + fast
  sync).
4. `internal/app/settings.go` + `settings_test.go` — the established
  page-engine pattern (the live config, the group-pointer law, the
  hermetic seams) the new pages build on.
5. `webui/settingsui.go` + `frontend/settings.js` — the binding
  wrapper + the page-builder pattern (SECTIONS, the mini-dialog kit).
6. `internal/cli/engine.go` (runImpl: the warmup family, the inputs
  fetchers, the websites phase, the Phase-5 CLEAR) + `phase5.go` —
  the CLI composition the app-surface wiring re-hosts.
7. `docs/port-blueprint/13-implementation-task-list.md` — the T-74
  row + Phase 6's book.

## 4. Known constraints / owner actions for SESSION-037

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The carried wiring debt is the first slice** (§2) — do not let the
  pages crowd it out; if capacity runs short, the wiring slice done
  properly with proof beats both half-done.
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush; T-74
  proceeds on (b) either way.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- **The go-directive acknowledgement (STILL WAITING — twenty-seventh
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755) and never commit a stray `webui/webui`
  binary.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs + the launchers' clean-VM smoke + the
  bootstrapper leg (Phase 7 checklist, doc 10 §7); the sound WAV
  assets (not yet committed — the preview's missing-asset notice is
  the parity behavior until T-77/T-80 ship the files).
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

## 5. THE PROMPT TO COPY (SESSION-037)

```
You are starting SESSION-037 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (thirty-one wipe events across
   SESSION-005..036 — SESSION-036 absorbed the thirty-first full wipe BEFORE
   work: both repos, the toolchain, the venv, the module cache, AND the
   rootless webkit tree all gone, everything rebuilt from the committed recipes
   — the fourth time the webkit recipe rebuilt the tree, again with exactly 26
   dev-symlink relinks to the system runtime copies; the continuation recipe is
   PROVEN at all levels: git status + git log --oneline -5 + file timestamps
   tell what ACTUALLY survived; a preserved commit or in-flight draft is
   continued and AUDITED — never blindly trusted, never restarted from zero; a
   resumed context re-verifies from scratch (the full battery + the hygiene
   scans) before committing; a wipe event may flip tracked files' MODES
   644->755 while preserving content — diagnose via 'git diff --raw' and
   restore via chmod off the parsed list, NEVER git checkout; and the standing
   addendum: after every `wails build` the wailsjs bindings re-flip to 755,
   and a stray `go build .` in webui/ leaves a 22 MB binary named `webui` —
   never commit it). FIRST verify with evidence: work repo at
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
   `git log --oneline -3`: SESSION-036 landed T-73 COMPLETE (the Settings
   pages A: internal/app/settings.go — THE LIVE-CONFIG LAW (config.json loads
   ONCE at construction through the GUI load path; the session-held map; saves
   mutate IN PLACE) + the six pages' engines over injectable SettingsDeps (the
   ONE merge save with all four v30 laws; the ladder; the token test; the
   wizard; the Ollama/llama.cpp detect family; the quick-switch fast lane; the
   sound gate) — commit 595a362 — and the window on the production shell
   (webui/settingsui.go the thin-shell wrapper + frontend/settings.js — the
   gear door, the 190px sidebar with the six rows, the pages from the snapshot,
   THE SAVE-ON-CHANGE law, the mini-dialog kit) — commit 48157da) and PUSHED
   everything after every task — 2020 main-battery tests race-clean over 35
   packages + 23 webui tests over 4. Verify git status shows up to date with
   origin/main (the tip at or beyond 48157da). Use the PAT only ephemerally:
   git -c http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off — the law pays in
   bugs, not just safety.

2. Read, in order: the repo README.md (Phase 6 OPEN — T-73 at 48157da) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-036's close: T-73 COMPLETE, the newest
   laws — the ONE-SLICE-PROPERLY record, the LIVE-CONFIG law, the
   GROUP-POINTER law) → docs/sessions/SESSION-036.md (the most recent
   report) → the T-74 specs (match the pages BY NAME: bot_queue.py +
   dashboard.py live in spec-gui-pages-a.md; backup_seal.py (drift seal +
   fast sync) in spec-gui-pages-b.md — doc 13's R-0 dated note) →
   internal/app/settings.go + webui/settingsui.go + frontend/settings.js
   (the established page pattern) → internal/cli/engine.go + phase5.go
   (the CLI composition the wiring re-hosts).

3. THIS SESSION IS T-74 (L): SETTINGS PAGES B per doc 13's row — Bot,
   Dashboard, Sources, Backup (+ radar, fast sync) — the pages matched
   BY NAME to their spec files (the R-0 note: bot_queue's spec lives in
   spec-gui-pages-a.md despite the A/B label; backup_seal's in
   spec-gui-pages-b.md). **THE SESSION'S FIRST SLICE — the carried
   wiring debt**: the CLI-parity shell wiring on the app surface (the
   warmup family, the inputs fetchers, the websites phase, the Phase-5
   CLEAR — engine.go's runImpl composition re-hosted on internal/app
   with its own tests, landed as its own committed slice). It gates
   T-77's hero batch; the pages' engine laws ride the established
   settings pattern (the live config, the group-pointer law, the
   hermetic seams; new pages join the sidebar at the reference's order
   with the pack's glyphs). ACCEPTANCE: the wiring slice's own tests +
   queue report rendering from a canned job result + the fast-sync laws
   (the ported pure halves are already tested). If capacity runs
   short: the wiring slice done properly with proof beats both
   half-done; close cleanly at a natural boundary.

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
   commit a stray webui/webui binary; every git/go command carries its own cd
   (the sandbox shell resets cwd between calls).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-74's
   own acceptance: the wiring slice's tests + the queue report rendering
   tests + the fast-sync law tests), commit + push everything (ask me for
   the GitHub PAT the moment a push is due — never store it anywhere;
   never end the session push-pending), write docs/sessions/SESSION-037.md,
   rewrite docs/NEXT-SESSION.md with the next session's prompt, and end your
   reply with a plain-language report plus the prompt I should copy to start
   SESSION-038.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
