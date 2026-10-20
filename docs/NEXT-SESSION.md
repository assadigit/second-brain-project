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
  (`internal/app/pertests.go` over five new hermetic `SettingsDeps`
  seams). **T-79b COMPLETE (SESSION-043)** — the five linking-tools
  doors LIVE (`internal/app/linkdoors.go` over three new hermetic
  `SettingsDeps` seams — `LinkOllamaChat`/`LinkCloudChat`/
  `LinkEmbedCall`; the TestWorker law: the busy guard → the captured
  run → the last-40 tail → the finish verdict; the dry-run +
  one-write SAFE-DEFAULT laws pinned; the recategorize two-door flow
  with the table modal; the guard widened to all five doors per the
  row's naming — the recorded divergence; the evidence at
  `docs/poc-artifacts/t-79b-linking-doors.md` + the two screenshots +
  the gate-results JSON). **The remaining tail: T-79c (the bot-queue
  tool dialogs: verify-all/export-all/retry-failed), T-79d (the
  Chrome doors: hand-deliver/hand-rows-scrape/tab-retry/attach),
  T-79e (the decommission picker)** — then Phase 6 closes and Phase 7
  (T-80: the deterministic zip builder …) opens. The export zip rides
  T-80's packaging composition per the standing assignment.
- **Push state: CLEAN** — SESSION-043 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
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
  stdlib-only — SESSION-043's boots HAD ensurepip, so the plain
  `python3 -m venv` form worked twice; if a future boot lacks it, the
  `--without-pip --copies` deviation is the fallback), the wails CLI
  v2.16.0 (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — **GONE at
  SESSION-043's close: the 39th wipe took it and the close itself
  needed no `wails build`; SESSION-044 IS GUI-touching (the T-79c
  doors need a wails build for the regenerated bindings) — rebuild it
  FIRST per webui/README.md + docs/poc-artifacts/t-70-gate.md +
  build/TOOLCHAIN.md Am.3 + Am.4 [144 debs → dpkg -x → the 42 .pc
  prefix rewrites → the 26 dev links; expect the hardlink fallback
  while symlink(2) is blocked] — the ELEVENTH rebuild)**, the MAIN
  go.mod exactly modernc.org/sqlite v1.36.1 +
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
  - **THE CONTEXT-DEATH CONTINUATION RECIPE, PROVEN AT THE CLOSEST
    CALL YET (SESSION-043)**: the first context died AFTER the full
    implementation + evidence but BEFORE any commit; the resumed
    context found the repos preserved WITH the in-flight draft (plus
    a fresh wipe of the toolchain) — `git status` + `git log` + file
    timestamps told what survived; the draft was AUDITED
    line-by-line against the reference's own bytes (never blindly
    trusted, never restarted from zero); the from-scratch
    re-verification (the full battery + race ×3 + gofmt/vet + the
    go.mod byte check) ran BEFORE the commit; the one audit gap
    (webui/README.md's map entries) was fixed in the same commit.
  - **THE VLM-FIRST-PASS MISREAD LESSON** (SESSION-042); **THE
    EXTRACT-AND-EXEC SCOPE LAW** (SESSION-041); **THE
    CROSS-BINARY DISAGREEMENT LAW** (SESSION-041); **THE
    SENTINEL-SEAT RACE LAW** (SESSION-040); **THE RETROACTIVE
    FIELD-TAG AUDIT** (SESSION-040); **THE NIL-OPTIONAL DEREF**
    (SESSION-040); **THE ONE-OWNER-PER-SURFACE law** (SESSION-040);
    **THE FIELD-TAG LAW** (SESSION-038); **THE DISPLAY-PIPELINE
    CORRUPTION LESSON** (SESSION-038 — the Edit tool strikes
    SPACE-vs-TAB on Go source: gofmt -l after every edit batch,
    gofmt -w restores since HEAD is gofmt-clean; patch with python
    reading the real bytes); **THE SYMLINK-BLOCK FALLBACK**
    (SESSION-037, TOOLCHAIN.md Am.4 — `ln -s` refuses AND the
    refusing call kills the whole bash invocation — never put `ln -s`
    in a compound command); **THE RENDER-CONTEXT LAW** (SESSION-037);
    **THE VLM RATE-LIMIT WINDOW** (SESSION-039 — pause between the
    screenshot verifications); **THE PINNED-VERDICT-ROW LESSON**
    (SESSION-039); **THE ONE-SLICE-PROPERLY RECORD** (036–043); **THE
    LIVE-CONFIG LAW**; **THE GROUP-POINTER LAW**; **THE
    DISSOLUTION-PROBE RECORD** (ADR-0002 Am.5); **THE
    GOLDEN-MARSHAL-ORDER TRAP**; **THE SUITE-REVERIFICATION
    DIVIDEND**; **THE WAILS-BUILD MODE FLIPS** (the wailsjs committed
    baseline IS 755; only a CONTENT diff should remain after a build;
    never commit a stray binary — build with `-o` into gitignored
    paths); **THE HERMETIC-SEAM LAW**; **THE RESUMED-CONTEXT
    RE-VERIFICATION LAW**; **THE MODE-FLIP DEGRADATION LAW** (a wipe
    may flip tracked modes 644→755 while preserving content —
    diagnose `git diff --raw`, restore via chmod off the parsed list,
    NEVER git checkout); **THE MODULE-BOUNDARY PROBE LAW**; **THE
    FRONTEND-MODULE LESSON**; **THE TIMER-CLAMP INSTRUMENT CAVEAT**;
    **THE TOOLCHAIN-ROUND-2/3 PROOF**; THE T-65 ZIP-COMPOSITION
    RESOLUTION; THE CORPUS-REGENERATION LAW; THE
    DISPATCH-COMPLETION LAW; THE STATE-CHAIN + MUTATION-TEETH laws;
    THE CONTEXT-DEATH CONTINUATION recipe; THE DSN-EMPTY-PATH HAZARD;
    THE FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the
    fake-clock arithmetic; the exception class-name tokens; the
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

## 2. Next session: SESSION-044 — T-79c: the bot-queue tool dialogs (the More menu's third door family)

- **SESSION-044 = T-79c (M)**: the bot-queue tool dialogs as
  More-menu doors — ✅ **Verify All Processed** + 📋 **Export All
  Links** + 🔄 **Retry Failed**, the reference's own flows over the
  ported queue engines.
- The reference's homes: `bot_queue.py:929` (export_all_bot_links) +
  `bot_queue.py:1026` (verify_all_bot_links) +
  `dashboard.py:770` (retry_failed_repos — the SQLite-cache failed-URL
  fetch + the reprocess batch; note the retry-banner clear + the
  `_bot_queue_urls` clearing).
- The ported engines to read FIRST: the Worker/bot queue surfaces
  (internal/workerclient, internal/app's queue/batch doors — T-77's
  MoreRetryBacklog/MoreRetryMaster are the already-live batch retry
  modes; T-79c's doors are the bot-queue TOOL dialogs), internal/store
  (the failed-queue reads), spec-gui-worker-jobs. THE GUARD LAW:
  these doors run worker jobs — carry the reference's own guards
  (the telegram lock where the reference takes it, the busy guard
  where a batch runs) exactly as the reference does.
- Follow T-79a/T-79b's shape: the engines as `Settings`/app methods
  over hermetic seams, the `HeroUI` More-door wrappers, the
  heroflow.js flip + dispatch, the line grammar pinned hermetically,
  the evidence probe's click leg, THE FIELD-TAG law honored (tag +
  pin every genuinely new payload).
- If capacity runs short: one door family done properly with proof
  beats two half-done; close cleanly at a natural boundary.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-043.md` — the most recent report (T-79b's
   landing; the two-wipe session + the continuation recipe's closest
   call).
2. `docs/port-blueprint/module-specs/spec-gui-worker-jobs.md` — the
   bot-queue section (the three dialogs' own spec).
3. `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the
   More-menu section + the dated notes (T-79a/T-79b's records are the
   patterns to follow).
4. The reference at 98f4fc7:
   `gui/main_window/bot_queue.py` (929 + 1026 — read the FULL bodies
   incl. the workers they spawn), `gui/main_window/dashboard.py:770`
   (retry_failed_repos' full body), and the engines behind them.
5. The Go engines' surfaces: `internal/workerclient/`,
   `internal/app/` (the queue/batch doors + MoreRetryBacklog/
   MoreRetryMaster's already-live shapes), `internal/store/repos.go`
   (the failed-queue reads: AddFailed/IsRepoSettled/the failed-URL
   fetch), `internal/cli`'s tool commands.
6. `docs/port-blueprint/13-implementation-task-list.md` — the T-79c
   row (the addendum's own book).
7. `docs/port-blueprint/15-acceptance-criteria.md` — AC-8 (the
   every-control coverage the doors serve).
8. As needed: `docs/poc-artifacts/t-79b-linking-doors.md` (the
   evidence pattern + the divergences' recording style).

## 4. Known constraints / owner actions for SESSION-044

- **THE WEBKIT TREE IS GONE** (the 39th wipe; SESSION-043's close
  needed no wails build) — SESSION-044 IS GUI-touching: rebuild FIRST
  per §1's recipe (the ELEVENTH rebuild; 144 debs → dpkg -x → the 42
  .pc rewrites → the 26 dev links, the hardlink fallback expected).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — thirty-fourth
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

## 5. THE PROMPT TO COPY (SESSION-044)

```
You are starting SESSION-044 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 6 is OPEN under the owner's Path-B stamp
   (SESSION-042: "the slow, more reliable and more rigorous method" — the
   More menu's honestly-disabled tail became the named rows T-79a…T-79e on
   doc 13; Phase 7 opens only when the tail is empty). T-79 + T-79a (the
   five per-test doors) + T-79b (the five linking-tools doors) are
   COMPLETE. SESSION-044 = T-79c (M): the bot-queue tool dialogs as
   More-menu doors — ✅ Verify All Processed + 📋 Export All Links + 🔄
   Retry Failed, the reference's own flows over the ported queue engines.
   The reference's homes: bot_queue.py:929 (export_all_bot_links) +
   bot_queue.py:1026 (verify_all_bot_links) + dashboard.py:770
   (retry_failed_repos). THE GUARD LAW: these doors run worker jobs —
   carry the reference's own guards (the telegram lock where the reference
   takes it, the busy guard where a batch runs) exactly as the reference
   does. Follow T-79a/T-79b's landed shape (internal/app/pertests.go +
   internal/app/linkdoors.go's pattern: the engines over hermetic seams +
   the HeroUI wrappers + the heroflow.js flip + the line grammar pinned
   hermetically + the evidence probe's click leg + THE FIELD-TAG law on
   every genuinely new payload). If capacity runs short: one door family
   done properly with proof beats two half-done.

1. The environment is REBUILDABLE, not persistent (thirty-nine wipe events
   across SESSION-005..043 — SESSION-043 absorbed TWO: the thirty-eighth
   before work [both repos, the toolchain, the venv, the module cache, the
   wails CLI, AND the webkit tree — all rebuilt from the committed
   recipes] and the thirty-ninth at the context resumption [the repos
   survived WITH the in-flight draft + tracked-file MODE FLIPS 644→755,
   content byte-intact — repaired via chmod off the parsed
   'git diff --raw' list, NEVER git checkout; the toolchain rebuilt; the
   draft audited line-by-line against the reference's own bytes, never
   blindly trusted]). NOTE the standing lessons: downloads and long suites
   must run in FOREGROUND CHUNKS (the sandbox kills backgrounded children
   when the call returns); this sandbox's tool-output rendering
   occasionally drops characters — settle every suspect reading with
   byte-level counts/hex before "fixing" anything (the Edit tool's
   space-vs-tab corruption struck AGAIN in 042 writing Go source — gofmt
   -l after every edit batch, gofmt -w restores because HEAD is
   gofmt-clean; patch with python reading the real bytes); the VLM CLI
   rate-limits (429) under rapid calls — pause between the screenshot
   verifications — AND the VLM can misread enabled-vs-disabled states on
   emoji-led labels (042's first pass called two LIVE doors "grayed";
   043's pass misread the lower region the same way — the byte-level
   source check + a focused cropped re-pass settled both; never "fix"
   code off a suspect VLM verdict); 'ln -s' at this boot REFUSES and
   kills the whole compound bash call — never put it in a compound command
   (the venv needed --without-pip --copies on ensurepip-less boots; the
   webkit tree's 26 dev links needed hardlinks); the continuation recipe
   is PROVEN at the closest call yet (043: the first context died after
   the full implementation + evidence but before any commit; git status +
   git log --oneline -5 + file timestamps told what ACTUALLY survived; a
   preserved commit or in-flight draft is continued and AUDITED — never
   blindly trusted, never restarted from zero; a resumed context
   re-verifies from scratch [the full battery + the race ritual + the
   hygiene scans] before committing). FIRST verify with evidence: work
   repo at /home/z/my-project/second-brain-project (the
   /home/z/second-brain-project symlink is DOWN while symlink(2) is
   blocked — use the real path), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator —
   same caveat; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64
   hex chars, verified against go.dev's published JSON; BEWARE the
   transcription corruption — the committed pin in build/TOOLCHAIN.md
   Am.2 is the authority), venv /home/z/venvs/gitcurator (Python 3.12,
   stdlib-only; the plain 'python3 -m venv' form when ensurepip is
   present, else --without-pip --copies), the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install
   github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only), and
   the rootless webkit tree (~/.local/lib-wails — GONE at 043's close;
   SESSION-044 IS GUI-touching: rebuild it FIRST per webui/README.md +
   docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4:
   144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
   hardlink fallback expected — the ELEVENTH rebuild). The MAIN go.mod:
   modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the
   go directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2;
   do not bump any of these). The NESTED webui module rides go 1.25.0
   (wails v2.16.0's floor) + the wails require + the transitive requires
   the direct internal/app import pulls (indirect via the parent replace
   — ADR-0002 Am.5); the MAIN go.mod stays byte-unchanged. Then check
   'git status' and 'git log --oneline -3': SESSION-043 landed T-79b in
   its close commits and PUSHED everything — 2312 main-battery tests
   race-clean over 36 packages (2280 + 32) + 63 webui tests over 4
   (59 + 4). Verify git status shows up to date with origin/main. Use
   the PAT only ephemerally: git -c http.extraheader="Authorization:
   Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and never store
   it — AND SCRUB IT if you ever CLONE with it (the GIT_CONFIG_COUNT
   env-var form avoids the persistence entirely). Re-verify the suite
   before work if anything looks off — the law pays in bugs, not just
   safety (paid three times now: 040's nil-deref + race, 041's EPERM
   divergence).

2. Read, in order: the repo README.md (SESSION-043's status block —
   T-79b) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records SESSION-043's
   close) → docs/sessions/SESSION-043.md (the most recent report —
   the two-wipe session + the continuation recipe's closest call) →
   docs/port-blueprint/module-specs/spec-gui-worker-jobs.md (the
   bot-queue section) → spec-gui-pages-a.md (the More-menu section +
   the T-79a/T-79b dated notes — the pattern to follow) → the
   reference's bot_queue.py:929/1026 + dashboard.py:770 (full bodies
   incl. the workers they spawn) → the Go engines' surfaces
   (internal/workerclient, internal/app's queue/batch doors —
   MoreRetryBacklog/MoreRetryMaster already live, internal/store's
   failed-queue reads, internal/cli's tool commands) → doc 13's T-79c
   row → doc 15's AC-8 row.

3. THIS SESSION = T-79c: the bot-queue tool dialogs as More-menu doors
   (the reference's own flows over the ported queue engines; the
   guard laws exactly as the reference takes them). One slice done
   properly with proof beats two half-done; close cleanly at a natural
   boundary when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git
   (byte-level secret-scan before every commit; never write
   token-shaped strings even as fakes); no new dependencies without an
   ADR (the authorized runtime set is exactly modernc.org/sqlite
   v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; the MAIN go.mod stays byte-unchanged; the wails require +
   the indirect transitive requires stay in the NESTED webui module
   per ADR-0002 Am.5); surface blockers with evidence (hash-level, not
   display) instead of improvising past them; probe before claiming any
   module-boundary rule (ADR-0002 Am.4); fake every network-touching
   seam in tests at construction (THE HERMETIC-SEAM law); when the
   reference's own code sits behind an import wall, drive the
   reference's own BYTES by pinned line-range extraction — structural
   pins + the rev guard, never hand-transcription, never a stubbed
   wall (THE EXTRACT-AND-EXEC SCOPE LAW); "same semantics" means the
   two binaries agree on EVERY input to the shared protocol, edge
   branches included — audit the except ladders (THE CROSS-BINARY
   DISAGREEMENT LAW); the payload structs' JSON tags mirror the Python
   dict keys and the frontend reads the tagged names (THE FIELD-TAG
   LAW — extend the payload field-tags test with every new payload);
   a transition seat read+written by concurrent passes holds it under
   a mutex as ONE unit (THE SENTINEL-SEAT RACE LAW — run the battery
   ≥3× before trusting a new concurrent seam; every live-map mutation
   holds the Settings mutex, the SaveSettings discipline); every
   pointer/optional field read is settled with a nil check (THE
   NIL-OPTIONAL DEREF); one owner per frontend surface (duplicate
   EventsOn listeners double-paint); a modal's action buttons live
   OUTSIDE the scrollable body; after every wails build, check 'git
   diff --raw' for mode flips (the wailsjs committed baseline IS 755)
   and never commit a stray binary; gofmt -l after every Go edit batch
   (the Edit tool's space-vs-tab corruption — gofmt -w restores);
   every git/go command carries its own cd AND its own
   PATH=/home/z/.local/go/bin export (the sandbox shell resets both
   between calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation
   (the chosen path's own acceptance), commit + push everything (ask
   me for the GitHub PAT the moment a push is due — never store it
   anywhere; never end the session push-pending), write
   docs/sessions/SESSION-044.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language
   report plus the prompt I should copy to start SESSION-045.

One slice done properly with proof beats two half-done; close cleanly
at a natural boundary when capacity runs out.
```
