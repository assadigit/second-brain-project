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
  (all generators × two passes GREEN).
- **Phase 6 OPEN — THE OWNER'S PATH-B STAMP (SESSION-042)**: "the slow,
  more reliable and more rigorous method" — Phase 6 stays OPEN past
  T-78; the More menu's honestly-disabled tail (the 19 doors,
  SESSION-040's record) became the NAMED addendum rows on doc 13:
  **T-79a COMPLETE (SESSION-042)** — the five per-test doors LIVE
  (🔑 Test Telegram & GitHub [the Option-B re-scope: the GitHub half +
  the Worker queue leg], 🌐 Test Proxy [the sentinel's TCP verdict],
  🧠 Test Ollama [models + the real "2 + 3" ping + the options retry],
  🔌 Test Cloud API [the /models pre-flight + the "Say hello" ping +
  the three verdict boxes], 🦙 Test llama.cpp [the probe + the pick +
  the ping + the three boxes]) — `internal/app/pertests.go` over five
  new hermetic `SettingsDeps` seams + the `HeroUI` wrappers + the
  `heroflow.js` wiring + the `SB_POC_OPEN_MORE` evidence probe; the
  evidence at `docs/poc-artifacts/t-79a-per-test-doors.md`. **The
  remaining tail: T-79b (the linking tools: recall-hooks dry-run +
  link suggestions + validate + verify + recategorize), T-79c (the
  bot-queue tool dialogs: verify-all/export-all/retry-failed), T-79d
  (the Chrome doors: hand-deliver/hand-rows-scrape/tab-retry/attach),
  T-79e (the decommission picker)** — then Phase 6 closes and Phase 7
  (T-80: the deterministic zip builder …) opens. The export zip rides
  T-80's packaging composition per the standing assignment.
- **Push state: CLEAN** — SESSION-042 committed + pushed everything at
  the close (`d8f255d` the bookkeeping + `0a2ae76` T-79a; the
  ephemeral PAT; never stored, never written). Verify `git status` in
  sync at the start.
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (the /home/z/second-brain-project
  symlink is DOWN while symlink(2) is blocked — use the real path),
  read-only reference /home/z/my-project/GitCurator @ `98f4fc7`
  (symlinked /home/z/GitCurator — same caveat; must match
  testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/go
  (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars, verified against go.dev's published JSON — BEWARE: the
  chat-prompt copies of this string keep losing a character, the same
  transcription corruption TOOLCHAIN.md Am.2 pinned; the committed pin
  is the authority), venv /home/z/venvs/gitcurator (Python 3.12,
  stdlib-only — **SESSION-042's deviation: built as `--without-pip
  --copies` because this boot's system python lacks ensurepip AND
  symlink(2) is blocked; if a future boot has ensurepip, the plain
  `python3.12 -m venv` form is the recipe**), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — **BUILT at
  SESSION-042's close (the NINTH rebuild, Am.4's hardlink fallback:
  the 26 dev links as hardlinks to the resolved system targets);
  if a wipe took it, rebuild per webui/README.md +
  docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4
  (144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev
  links; expect the hardlink fallback while symlink(2) is blocked)**),
  the MAIN go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (never bump;
  ADR-0002 Am.2). The NESTED webui module rides go 1.25.0 (wails
  v2.16.0's floor) + the wails require + the transitive requires the
  direct internal/app import pulls (indirect via the parent replace —
  ADR-0002 Am.5; the MAIN go.mod stays byte-unchanged). **The downloads
  must run in FOREGROUND CHUNKS** — the sandbox kills backgrounded
  children when the call returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE VLM-FIRST-PASS MISREAD LESSON** (SESSION-042): the VLM can
    misread enabled-vs-disabled states on emoji-led labels (it called
    two LIVE per-test doors "grayed out") — settle EVERY suspect
    display verdict with the byte-level source check + a focused
    cropped re-pass before "fixing" anything (the display-pipeline
    lesson's VLM twin); the character-level transcription noise on
    emoji/lines is real (🧠 read as 🧪) — trust the hermetic tests'
    pinned grammar + the format details (URLs, errnos), not the VLM's
    exact glyphs.
  - **THE EXTRACT-AND-EXEC SCOPE LAW** (SESSION-041); **THE
    CROSS-BINARY DISAGREEMENT LAW** (SESSION-041); **THE
    SENTINEL-SEAT RACE LAW** (SESSION-040 — the T-79a application:
    the llama.cpp model-seat write holds the Settings mutex, the
    SaveSettings discipline for every live-map mutation); **THE
    RETROACTIVE FIELD-TAG AUDIT** (SESSION-040); **THE NIL-OPTIONAL
    DEREF** (SESSION-040); **THE ONE-OWNER-PER-SURFACE law**
    (SESSION-040); **THE FIELD-TAG LAW** (SESSION-038); **THE
    DISPLAY-PIPELINE CORRUPTION LESSON** (SESSION-038 — the Edit tool
    strikes SPACE-vs-TAB on Go source: gofmt -l after every edit
    batch, gofmt -w restores since HEAD is gofmt-clean; patch with
    python reading the real bytes); **THE SYMLINK-BLOCK FALLBACK**
    (SESSION-037, TOOLCHAIN.md Am.4 — STILL blocked at SESSION-042's
    boot: `ln -s` refuses AND the refusing call kills the whole bash
    invocation — never put `ln -s` in a compound command); **THE
    RENDER-CONTEXT LAW** (SESSION-037); **THE VLM RATE-LIMIT WINDOW**
    (SESSION-039 — pause between the screenshot verifications); **THE
    PINNED-VERDICT-ROW LESSON** (SESSION-039); **THE ONE-SLICE-
    PROPERLY RECORD** (036–042); **THE LIVE-CONFIG LAW**; **THE
    GROUP-POINTER LAW**; **THE DISSOLUTION-PROBE RECORD** (ADR-0002
    Am.5); **THE GOLDEN-MARSHAL-ORDER TRAP**; **THE
    SUITE-REVERIFICATION DIVIDEND**; **THE WAILS-BUILD MODE FLIPS**
    (the wailsjs committed baseline IS 755; only a CONTENT diff should
    remain after a build; never commit a stray binary — build with
    `-o` into gitignored paths); **THE HERMETIC-SEAM LAW**; **THE
    RESUMED-CONTEXT RE-VERIFICATION LAW**; **THE MODE-FLIP
    DEGRADATION LAW**; **THE MODULE-BOUNDARY PROBE LAW**; **THE
    FRONTEND-MODULE LESSON**; **THE TIMER-CLAMP INSTRUMENT CAVEAT**;
    **THE TOOLCHAIN-ROUND-2/3 PROOF**; THE T-65 ZIP-COMPOSITION
    RESOLUTION; THE CORPUS-REGENERATION LAW; THE
    DISPATCH-COMPLETION LAW; THE STATE-CHAIN + MUTATION-TEETH laws;
    THE CONTEXT-DEATH CONTINUATION recipe; THE DSN-EMPTY-PATH HAZARD;
    THE FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the fake-clock
    arithmetic; the exception class-name tokens; the
    stdout-capture-helper limit.
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd — the sandbox's shell resets cwd AND PATH between calls;
    export PATH=/home/z/.local/go/bin:$PATH in the same command; the
    NESTED webui module needs its own cd; never commit outside the
    work repo; never checkout blindly); **THE RENDERER-EATS-ESCAPE-PAIRS
    lesson** (commit messages via -F FILE); **THE WRITE-TOOL
    REDACTION lesson** (byte-level secret-scan before every commit);
    **THE BACKGROUND-PROCESS lesson** (long suites + downloads in
    FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for the PAT the
    moment a push is due; never store it).

## 2. Next session: SESSION-043 — T-79b: the linking tools (the More menu's second door family)

- **SESSION-043 = T-79b (L)**: the linking tools as More-menu doors,
  the reference's SAFE defaults driven from the GUI over the ported
  engines: 🪝 **Recall hooks (dry-run — diff only)** +
  🔗 **Build link suggestions (suggest mode — the one write is the
  Suggestions note under the manual vault's Library/)** + ✅ **Validate
  Vault** + 🔍 **Verify Vault** + 📁 **Recategorize Notes**.
- The reference's homes: `linking_tools.py:107/116` (the two tool
  runners inside TestWorker with stdout capture, the last 40 lines
  streamed; the guard: no tool while a batch runs),
  `dashboard.py:313/500` (verify_vault + recategorize_notes),
  `connection_tests.py:219` (test_vault — the .obsidian check).
- The ported engines: T-50's `internal/mirror` (+ the linking/
  embeddings/recall doors), T-45/T-46's vaultscan/vaulthygiene —
  READ the engines' actual surfaces before designing the doors (the
  spec's "same flags, same outputs" law; the GUI runs the SAFE
  defaults only).
- THE GUARD LAW (the reference's own): no tool while a batch runs —
  the linking-tools doors DO carry the busy guard (unlike T-79a's
  read-only probes; these tools read/write the vaults).
- Follow T-79a's shape: the engines as `Settings`/app methods over
  hermetic seams, the `HeroUI` More-door wrappers, the heroflow.js
  flip + dispatch, the line grammar pinned hermetically, the evidence
  probe extended (a door click leg), the FIELD-TAG law honored (no
  new crossed payloads unless a payload is genuinely new — then tag
  it + pin it).
- If capacity runs short: one door family done properly with proof
  beats two half-done; close cleanly at a natural boundary.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-042.md` — the most recent report (the
   Path-B stamp; T-79a's landing; the 37th wipe + the 9th webkit
   rebuild).
2. `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the
   linking_tools section + the dated notes (T-79a's record is the
   pattern to follow).
3. `docs/port-blueprint/module-specs/spec-tools.md` — build_links.py +
   add_recall_hooks.py (the SAFE defaults the GUI drives).
4. The reference at 98f4fc7: `gui/main_window/linking_tools.py` (124
   lines — read it ALL), `gui/main_window/dashboard.py:313/500`,
   `gui/main_window/connection_tests.py:219`, and the tools'
   engines (`gitcurator/tools/build_links.py`,
   `add_recall_hooks.py`).
5. The Go engines' surfaces: `internal/mirror/` (the linking doors),
   `internal/vaultscan/`, `internal/vaulthygiene/` — what exists,
   what the CLI already drives (`internal/cli`'s tool commands).
6. `docs/port-blueprint/13-implementation-task-list.md` — the T-79b
   row (the addendum's own book).
7. `docs/port-blueprint/15-acceptance-criteria.md` — AC-8 (the
   every-control coverage the doors serve).
8. As needed: `docs/poc-artifacts/t-79a-per-test-doors.md` (the
   evidence pattern + the divergences' recording style).

## 4. Known constraints / owner actions for SESSION-043

- **The webkit tree is BUILT at SESSION-042's close** — verify at boot
  (`ls ~/.local/lib-wails/usr/lib/x86_64-linux-gnu/pkgconfig`); if a
  wipe took it, rebuild FIRST per §1's recipe (144 debs → dpkg -x →
  the .pc rewrites → the 26 dev links with the hardlink fallback).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — thirty-third
  session carrying it): the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755 — the committed baseline IS 755 for the
  wailsjs files; only a CONTENT diff should remain) and never commit a
  stray binary (webui/webui, or a repo-root package binary — build
  with `-o` into gitignored paths).
- The answering-window belt + the T-71 merged-verdict question
  (SESSION-039's report): two recorded divergences awaiting the owner;
  neither gates the tail.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke + the bootstrapper leg
  (Phase 7 checklist, doc 10 §7); the Backup page's Restore button
  (honestly disabled — rides T-80's packaging composition with the
  export zip).
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

## 5. THE PROMPT TO COPY (SESSION-043)

```
You are starting SESSION-043 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 6 is OPEN under the owner's Path-B stamp
   (SESSION-042: "the slow, more reliable and more rigorous method" — the
   More menu's honestly-disabled tail became the named rows T-79a…T-79e on
   doc 13; Phase 7 opens only when the tail is empty). T-79 (the addendum)
   + T-79a (the five per-test doors) are COMPLETE. SESSION-043 = T-79b (L):
   the linking tools as More-menu doors — the reference's SAFE defaults
   driven from the GUI over the ported engines: 🪝 Recall hooks (dry-run —
   diff only) + 🔗 Build link suggestions (suggest mode — the one write is
   the Suggestions note under the manual vault's Library/) + ✅ Validate
   Vault + 🔍 Verify Vault + 📁 Recategorize Notes. The reference's homes:
   linking_tools.py:107/116 (the TestWorker runners, stdout capture, the
   last 40 lines streamed, the busy guard), dashboard.py:313/500,
   connection_tests.py:219. THE GUARD LAW: these doors DO carry the
   batch-busy guard (they read/write the vaults, unlike T-79a's read-only
   probes). Follow T-79a's landed shape (internal/app/pertests.go's pattern:
   the engines over hermetic seams + the HeroUI wrappers + the heroflow.js
   flip + the line grammar pinned hermetically + the evidence probe's click
   leg). If capacity runs short: one door family done properly with proof
   beats two half-done.

1. The environment is REBUILDABLE, not persistent (thirty-seven wipe events
   across SESSION-005..042 — SESSION-042 absorbed the thirty-seventh full
   wipe BEFORE work: both repos, the toolchain, the venv, the module cache,
   AND the wails CLI gone; everything rebuilt from the committed recipes,
   INCLUDING the webkit tree [the NINTH rebuild — a GUI-touching session
   needed it], Am.4's hardlink fallback again); NOTE the standing lessons:
   downloads and long suites must run in FOREGROUND CHUNKS (the sandbox
   kills backgrounded children when the call returns); this sandbox's
   tool-output rendering occasionally drops characters — settle every
   suspect reading with byte-level counts/hex before "fixing" anything
   (the Edit tool's space-vs-tab corruption struck AGAIN in 042 writing Go
   source — gofmt -l after every edit batch, gofmt -w restores because
   HEAD is gofmt-clean; patch with python reading the real bytes); the
   VLM CLI rate-limits (429) under rapid calls — pause between the
   screenshot verifications — AND the VLM can misread enabled-vs-disabled
   states on emoji-led labels (042's first pass called two LIVE doors
   "grayed"; the byte-level source check + a focused cropped re-pass
   settled it — never "fix" code off a suspect VLM verdict); `ln -s` at
   this boot REFUSES and kills the whole compound bash call — never put
   it in a compound command (the venv needed --without-pip --copies; the
   webkit tree's 26 dev links needed hardlinks); the continuation recipe
   is PROVEN at all levels: git status + git log --oneline -5 + file
   timestamps tell what ACTUALLY survived; a preserved commit or in-flight
   draft is continued and AUDITED — never blindly trusted, never restarted
   from zero; a resumed context re-verifies from scratch (the full battery
   + the hygiene scans) before committing; a wipe event may flip tracked
   files' MODES 644->755 while preserving content — diagnose via 'git
   diff --raw' and restore via chmod off the parsed list, NEVER git
   checkout; and the standing addendum: after every `wails build` the
   wailsjs bindings re-flip to 755 (the repo's committed baseline for the
   wailsjs files IS 755 — only a CONTENT diff should remain after a
   build), a stray `go build .` in webui/ leaves a 22 MB binary named
   `webui`, and a stray `go build ./pkg/` at the repo root leaves the
   package's binary there — never commit either; build with -o into
   gitignored paths). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (the /home/z/second-brain-project
   symlink is DOWN while symlink(2) is blocked — use the real path),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator — same caveat; must match testdiff/PINNED-GITCURATOR-REV),
   toolchain go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON; BEWARE the transcription
   corruption — the committed pin in build/TOOLCHAIN.md Am.2 is the authority),
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only; 042's deviation:
   --without-pip --copies if ensurepip is absent again), the wails CLI
   v2.16.0 (GOBIN=/home/z/.local/go/bin go install
   github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only), and
   the rootless webkit tree (~/.local/lib-wails — BUILT at 042's close; if
   a wipe took it, rebuild per webui/README.md + docs/poc-artifacts/
   t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4: 144 debs → dpkg -x →
   the 42 .pc prefix rewrites → the 26 dev links, hardlink fallback
   expected). The MAIN go.mod: modernc.org/sqlite at v1.36.1,
   bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 +
   golang.org/x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these).
   The NESTED webui module rides go 1.25.0 (wails v2.16.0's floor) + the
   wails require + the transitive requires the direct internal/app import
   pulls (indirect via the parent replace — ADR-0002 Am.5); the MAIN
   go.mod stays byte-unchanged. Then check `git status` and `git log
   --oneline -3`: SESSION-042 landed T-79 (the addendum rows) + T-79a
   (the five per-test doors) in two commits — d8f255d the bookkeeping +
   0a2ae76 T-79a — and PUSHED everything at the close — 2280 main-battery
   tests race-clean over 36 packages (2253 + 27) + 59 webui tests over 4
   (56 + 3). Verify git status shows up to date with origin/main (the tip
   at or beyond 0a2ae76). Use the PAT only ephemerally: git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s'
   "$PAT" | base64)" and never store it — AND SCRUB IT if you ever CLONE
   with it (the GIT_CONFIG_COUNT env-var form avoids the persistence
   entirely). Re-verify the suite before work if anything looks off — the
   law pays in bugs, not just safety (paid three times now: 040's
   nil-deref + race, 041's EPERM divergence).

2. Read, in order: the repo README.md (SESSION-042's status block — the
   Path-B stamp + T-79a) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records SESSION-042's close)
   → docs/sessions/SESSION-042.md (the most recent report) →
   docs/port-blueprint/module-specs/spec-gui-pages-a.md (the linking_tools
   section + T-79a's dated note — the pattern to follow) → spec-tools.md
   (build_links/add_recall_hooks — the SAFE defaults) → the reference's
   linking_tools.py (ALL 124 lines) + dashboard.py:313/500 +
   connection_tests.py:219 → the Go engines' surfaces (internal/mirror's
   linking doors, internal/vaultscan, internal/vaulthygiene, internal/cli's
   tool commands) → doc 13's T-79b row → doc 15's AC-8 row.

3. THIS SESSION = T-79b: the linking tools as More-menu doors (the
   reference's SAFE defaults over the ported engines; the busy guard law;
   the TestWorker-stdout-capture→last-40-lines-streamed law; the dry-run
   diff-only law for recall hooks; the suggest-mode one-write law for link
   suggestions). One slice done properly with proof beats two half-done;
   close cleanly at a natural boundary when capacity runs out.

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
   network-touching seam in tests at construction (THE HERMETIC-SEAM law);
   when the reference's own code sits behind an import wall, drive the
   reference's own BYTES by pinned line-range extraction — structural pins
   + the rev guard, never hand-transcription, never a stubbed wall (THE
   EXTRACT-AND-EXEC SCOPE LAW); "same semantics" means the two binaries
   agree on EVERY input to the shared protocol, edge branches included —
   audit the except ladders (THE CROSS-BINARY DISAGREEMENT LAW); the payload
   structs' JSON tags mirror the Python dict keys and the frontend reads the
   tagged names (THE FIELD-TAG LAW — extend TestCrossedPayloadFieldTags with
   every new payload); a transition seat read+written by concurrent passes
   holds it under a mutex as ONE unit (THE SENTINEL-SEAT RACE LAW — run the
   battery ≥3× before trusting a new concurrent seam; every live-map
   mutation holds the Settings mutex, the SaveSettings discipline); every
   *T field read is settled with a nil check (THE NIL-OPTIONAL DEREF); one
   owner per frontend surface (duplicate EventsOn listeners double-paint);
   a modal's action buttons live OUTSIDE the scrollable body; after every
   wails build, check `git diff --raw` for mode flips (the wailsjs committed
   baseline IS 755) and never commit a stray binary; gofmt -l after every
   Go edit batch (the Edit tool's space-vs-tab corruption — gofmt -w
   restores); every git/go command carries its own cd AND its own
   PATH=/home/z/.local/go/bin export (the sandbox shell resets both between
   calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
   chosen path's own acceptance), commit + push everything (ask me for the
   GitHub PAT the moment a push is due — never store it anywhere; never end
   the session push-pending), write docs/sessions/SESSION-043.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-044.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
