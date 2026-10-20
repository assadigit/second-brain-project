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
- **Phase 4 CLOSED (SESSION-015–026)**: T-40…T-50 + the Phase-4 closer.
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity
  release `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip
  asset sha-verified + round-trip verified); the differential corpus
  full run (all generators × two passes GREEN).
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70 gate
  through the T-79a…T-79f Path-B rows.
- **Phase 7 OPEN — T-80 COMPLETE (SESSION-048)**; **T-81 (the release
  workflow) LANDED at SESSION-049 with the dry-run's RUNNER LEGS
  BLOCKED on the account's Actions billing (the owner action)**: the
  workflow (`.github/workflows/release.yml`; go-ci.yml untouched) +
  6 law tests landed + every command it runs proven locally at the
  TAGGED state (CLI `13808277…` A=B=cold; GUI `99cc22e6…` A=B=cold;
  the zip `071e1968…` A=B — the runner-comparison target); the notes
  step proven by extract-and-exec; **THE TAG-STATE BUILD LAW found**
  (go1.24+ embeds the VCS-derived module version — an untagged and a
  tagged build of the SAME commit differ; the release numbers live at
  the tagged checkout; the wails GUI exe is exempt). The blocked leg:
  runs `39511458026`/`39512685975` died at job setup — the free-plan
  account's Actions minutes/spending limit refused every runner job
  (and go-ci has been failing the SAME way on every recent main push,
  8/8, incl. SESSION-048's close — nobody watched the Actions tab).
  **The owner fix (Billing & plans) + the re-run recipe:
  `docs/poc-artifacts/t-81-release-workflow.md` §3.** T-82 (the
  clean-VM checklist) + T-83 (the docs) follow.
- **Push state: CLEAN** — SESSION-049 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (the /home/z/second-brain-project
  symlink is DOWN while symlink(2) is blocked — use the real path),
  read-only reference /home/z/my-project/GitCurator @ `98f4fc7`
  (must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars; NOTE: go.dev's ?mode=json has ROTATED past 1.27.2: verify
  the locally-computed hash against the COMMITTED pin in build/
  TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the transcription
  corruption), venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only —
  the plain `python3 -m venv` form when ensurepip is present, else the
  `--without-pip --copies` fallback), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — DOWN after the
  45th wipe; the SEVENTEENTH rebuild is pending: rebuild it FIRST only
  when the session runs a LINUX GUI build (the windows cross-compile
  legs are CGO-free — SESSION-048/049 needed no webkit) per
  webui/README.md + docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md
  Am.3 + Am.4: 144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26
  dev links, hardlink fallback expected), the MAIN go.mod exactly
  modernc.org/sqlite v1.36.1 + bogdanfinn/tls-client v1.16.0 +
  golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 + the go
  directive at 1.24.1 (never bump; ADR-0002 Am.2). The NESTED webui
  module rides go 1.25.0 + the wails require + the transitive requires
  (indirect via the parent replace — ADR-0002 Am.5; the MAIN go.mod
  stays byte-unchanged). **The downloads must run in FOREGROUND
  CHUNKS** — the sandbox kills backgrounded children when the call
  returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE 45TH-WIPE RECIPE (SESSION-049)**: same as the 41st–44th —
    clone both with the ephemeral GIT_CONFIG_COUNT PAT form, verify
    the rev pin, rebuild per the committed recipes, and re-verify the
    FULL battery (plain + race) before touching code (this session:
    2371 + 89 pre-work, 2377 + 89 at the close — 6 new).
  - **THE TAG-STATE BUILD LAW (SESSION-049)**: go1.24+ embeds the
    VCS-derived module version in buildinfo — the SAME COMMIT built
    untagged vs tagged produces DIFFERENT CLI-exe bytes (pseudo-version
    vs the tag name); the release numbers live at the TAGGED checkout;
    never compare a runner/tagged zip against an untagged local build;
    the wails GUI exe is exempt ("(devel)" — byte-stable).
  - **THE ACTIONS-BILLING BLOCK RECORD (SESSION-049)**: the free-plan
    account's Actions minutes/spending limit refused every runner job
    (release runs 39511458026 + 39512685975 died at setup, runner_id 0,
    zero steps; go-ci 8/8 recent pushes the same way — SILENTLY red
    since before SESSION-048's close). Until the owner fixes Billing &
    plans, NO workflow can run; watch the Actions tab after the fix.
  - **THE -X-UNDER-TRIMPATH LAW (SESSION-048)**; **THE
    VERIFY-THE-VERIFIER LESSON (SESSION-048 — paid again at 049: a
    zip-member extraction bug produced a "mystery third sha" — the
    filter matched the SETUP BAT, not the exe; settled byte-level)**;
    **THE DEFAULT-ICON FIND (SESSION-048)**; **THE 43RD-WIPE RECIPE
    (SESSION-047)**; **THE FLAG-CLEAR LAW (SESSION-047)**; **THE
    REFERENCE'S-OWN-ARITHMETIC LAW (SESSION-046)**; **THE 42ND-WIPE
    RECIPE**; **THE VLM MISREAD LEDGER (FOUR incidents: 042/043/044/
    045; 046 AND 047 ran clean)**; **THE BINDING-IMPORT LAW
    (SESSION-044)**; **THE RACE-PASS DIVIDEND (SESSION-045)**; **THE
    EXTRACT-AND-EXEC SCOPE LAW (SESSION-041 — used again at 049: the
    workflow's own notes python proven from the workflow's bytes)**;
    **THE CROSS-BINARY DISAGREEMENT LAW**; **THE SENTINEL-SEAT RACE
    LAW (SESSION-040)**; **THE RETROACTIVE FIELD-TAG AUDIT
    (SESSION-040)**; **THE NIL-OPTIONAL DEREF (SESSION-040)**; **THE
    ONE-OWNER-PER-SURFACE law (SESSION-040)**; **THE FIELD-TAG LAW
    (SESSION-038)**; **THE DISPLAY-PIPELINE CORRUPTION LESSON
    (SESSION-038 — the Edit tool's SPACE-vs-TAB; gofmt -l after every
    edit batch)**; **THE SYMLINK-BLOCK FALLBACK (TOOLCHAIN.md Am.4)**;
    **THE RENDER-CONTEXT LAW (SESSION-037)**; **THE VLM RATE-LIMIT
    WINDOW (SESSION-039)**; **THE PINNED-VERDICT-ROW LESSON
    (SESSION-039)**; **THE ONE-SLICE-PROPERLY RECORD (036–049)**; and
    the standing set (THE LIVE-CONFIG LAW; THE GROUP-POINTER LAW; THE
    DISSOLUTION-PROBE RECORD; THE GOLDEN-MARSHAL-ORDER TRAP; THE
    SUITE-REVERIFICATION DIVIDEND; THE WAILS-BUILD MODE FLIPS; THE
    HERMETIC-SEAM LAW; THE RESUMED-CONTEXT RE-VERIFICATION LAW; THE
    MODE-FLIP DEGRADATION LAW; THE MODULE-BOUNDARY PROBE LAW; THE
    FRONTEND-MODULE LESSON; the T-65 zip-composition resolution; the
    corpus-regeneration law; the dispatch-completion law; the
    state-chain + mutation-teeth laws; the context-death continuation
    recipe; the DSN-empty-path hazard; the fixed-length roots; the
    shallow-copy snapshot; the fake-clock arithmetic; the exception
    class-name tokens; the stdout-capture-helper limit).
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd AND its own PATH=/home/z/.local/go/bin export); **THE
    RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F FILE);
    **THE WRITE-TOOL REDACTION lesson** (byte-level secret-scan before
    every commit); **THE BACKGROUND-PROCESS lesson** (long suites +
    downloads in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for
    the PAT the moment a push is due; never store it).

## 2. Next session: SESSION-050 — complete T-81's dry-run (if the billing fix landed) + T-82's preparable legs

- **FIRST (if the owner fixed Billing & plans)**: complete T-81's
  acceptance — re-run `39512685975` (Actions → the `release` workflow;
  the test tag `v0.0.0-t81-dryrun` still points at `0961ad9`), verify
  the green run end-to-end (the ritual's logs + the `[DRY-RUN]` draft
  release + the zip asset `GitCurator-v0.92.0-go.1-windows.zip`), run
  the cross-machine sha256 comparison (the asset vs the tagged-state
  table: CLI `13808277…` / GUI `99cc22e6…` / zip `071e1968…` —
  identical = cross-machine reproducibility proven; a difference =
  recorded honestly as the divergence), optionally the refusal leg
  (`v9.9.9-t81-refusal` must fail at the tag-law step + create NO
  release), then DELETE the test release + the test tag and flip doc
  13's T-81 row to COMPLETE. If billing is NOT fixed: re-surface the
  blocker (the evidence doc §3 stands) and proceed.
- **THEN T-82 (the clean-VM checklist + the AV scan note — doc 10
  §7/§9)**: the session's preparable legs = the checklist document
  (the nine §7 items as the owner's fill-in execution record) + the
  release-report scaffold (AC-23's sha256 + parity report, ready to
  carry the owner's signed results) + the AV note (doc 10 §9's
  mitigation ladder + the `-w -s` verdict placeholder). The live
  Windows execution is the owner's (this sandbox has no Windows box —
  the standing constraint since SESSION-001).
- T-83 (the docs: Go README, QUICKSTART update, parity report) follows
  — one slice done properly with proof beats two half-done; close
  cleanly at a natural boundary when capacity runs short.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-049.md` — the most recent report (T-81's
   landing + the find + the blocked leg).
2. `docs/poc-artifacts/t-81-release-workflow.md` — the evidence +
   the re-run recipe (§3) + the tagged-state comparison table.
3. `.github/workflows/release.yml` + `build/mkzip/workflow_test.go` +
   `build/release.sh`'s header — the landed automation + its laws.
4. `docs/port-blueprint/10-build-release-packaging.md` §7/§8/§9 —
   T-82's checklist + the AV notes + §8's dated note.
5. `docs/port-blueprint/13-implementation-task-list.md` — the T-81
   row (the PARTIAL record) + the T-82/T-83 rows.
6. `docs/port-blueprint/15-acceptance-criteria.md` §6 — AC-23/AC-24.

## 4. Known constraints / owner actions for SESSION-050

- **OWNER ACTION #1 (urgent for T-81's completion): the Actions
  billing fix** — GitHub → Settings → Billing & plans (payment method /
  spending limit; the private repo's Actions minutes). Until then NO
  workflow runs (go-ci silently red too — it goes green on the next
  push after the fix).
- **The test tag `v0.0.0-t81-dryrun` lives at `0961ad9`** (pointing at
  the FIXED workflow) awaiting the re-run; delete it + the `[DRY-RUN]`
  release after the proof. NEVER push a real `v<VERSION>` tag casually
  — formal releases happen only at roadmap milestones (the tag law:
  a real tag must equal `v$(cat VERSION)`; the workflow enforces it).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (the 45th wipe took it; the SEVENTEENTH
  rebuild is pending) — rebuild it FIRST only if the session runs a
  LINUX GUI build (the windows legs are CGO-free; SESSION-048/049
  worked without it).
- The dissolution call (the owner's, ADR-0002 Am.5); the
  vanilla-vs-vite call (the owner's); the batch-scorecard modal
  question (the owner's); the v0.91.0 zip-parity reading (the owner's,
  minor); the answering-window belt + the T-71 merged-verdict question
  (SESSION-039's record).
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — fortieth session
  carrying it): the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the wailsjs baseline IS 755) and never commit a stray binary (nor a
  generated notes.md).
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke + the bootstrapper leg
  (Phase 7 checklist, doc 10 §7).
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

## 5. THE PROMPT TO COPY (SESSION-050)

```
You are starting SESSION-050 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 7 is OPEN — T-80 (the deterministic zip builder +
version stamping + manifest/icon) COMPLETE at SESSION-048; T-81 (the release
workflow) LANDED at SESSION-049 with the dry-run's RUNNER LEGS BLOCKED on the
account's Actions billing: the workflow (.github/workflows/release.yml — the
NEW file; go-ci.yml untouched) + 6 law tests (build/mkzip now 16) + every
command it runs proven locally at the TAGGED state (CLI 13808277… A=B=cold;
GUI 99cc22e6… A=B=cold; the zip 071e1968… A=B — the runner-comparison
target), the notes step proven by extract-and-exec from the workflow's own
bytes, and THE TAG-STATE BUILD LAW found (go1.24+ embeds the VCS-derived
module version — an untagged and a tagged build of the SAME commit differ;
the release numbers live at the tagged checkout; the wails GUI exe is exempt
— "(devel)"). The blocked leg: runs 39511458026/39512685975 died at job
setup — the free-plan account's Actions minutes/spending limit refused every
runner job (and go-ci has been failing the SAME way on every recent main
push, 8/8 — nobody watched the Actions tab). SESSION-050 = FIRST complete
T-81's dry-run IF the owner fixed Billing & plans (re-run 39512685975 — the
test tag v0.0.0-t81-dryrun still points at 0961ad9 — verify the green run +
the [DRY-RUN] draft release + the zip asset, run the cross-machine sha256
comparison vs 071e1968…, optionally the refusal leg v9.9.9-t81-refusal, then
DELETE the test release + tag and flip doc 13's row to COMPLETE; if billing
is NOT fixed, re-surface the blocker and proceed), THEN T-82 (the clean-VM
checklist + the AV scan note, doc 10 §7/§9): the checklist document + the
release-report scaffold (AC-23) + the AV note — the owner's live Windows
execution records land into them. T-83 (the docs) follows — one slice done
properly with proof beats two half-done; close cleanly at a natural boundary
when capacity runs short.

1. The environment is REBUILDABLE, not persistent (forty-five wipe events
   across SESSION-005..049 — SESSION-049 opened on the forty-fifth: the
   WHOLE environment was gone again, both repos included, all rebuilt from
   the committed recipes). NOTE the standing lessons: downloads and long
   suites must run in FOREGROUND CHUNKS (the sandbox kills backgrounded
   children when the call returns); this sandbox's tool-output rendering
   occasionally drops characters — settle every suspect reading with
   byte-level counts/hex before "fixing" anything (the Edit tool's
   space-vs-tab corruption struck 042/044/046/047/048 writing Go source —
   gofmt -l after every edit batch, gofmt -w restores because HEAD is
   gofmt-clean; patch with python reading the real bytes; AND 048's
   sharper edge: the renderer ate '[main]' into 'ain]' ON THE
   VERIFICATION ITSELF; AND 049's cousin: a zip-member extraction filter
   matched the SETUP BAT instead of the exe and nearly manufactured a
   "mystery third sha" — byte-settled; distrust the extraction AND the
   rendering of the check too); the VLM CLI rate-limits (429) under rapid
   calls AND misreads enabled-vs-disabled states (FOUR incidents — never
   "fix" code off a suspect VLM verdict); 'ln -s' refuses AND kills the
   whole compound bash call at this boot — never put it in a compound
   command (the venv needs --without-pip --copies on ensurepip-less
   boots; the webkit tree's 26 dev links need Am.4 hardlinks); THE
   TAG-STATE BUILD LAW (049); THE -X-UNDER-TRIMPATH LAW (048); THE
   DEFAULT-ICON FIND (048); THE BINDING-IMPORT LAW (044); THE
   FLAG-CLEAR LAW (047); THE REFERENCE'S-OWN-ARITHMETIC LAW (046/048).
   FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (the
   /home/z/second-brain-project symlink is DOWN while symlink(2) is
   blocked — use the real path), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (must match
   testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 ->
   /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64
   hex chars; go.dev's ?mode=json has ROTATED past 1.27.2: verify the
   locally-computed hash against the COMMITTED pin in build/
   TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the
   transcription corruption), venv /home/z/venvs/gitcurator (Python 3.12,
   stdlib-only; the plain 'python3 -m venv' form when ensurepip is
   present, else --without-pip --copies), the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install
   github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only), and
   the rootless webkit tree (~/.local/lib-wails — DOWN after the 45th
   wipe; the SEVENTEENTH rebuild is pending: rebuild it FIRST only when
   the session runs a LINUX GUI build — the windows cross-compile legs
   are CGO-free, SESSION-048/049 needed no webkit — per webui/README.md +
   docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4:
   144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
   hardlink fallback expected). The MAIN go.mod: modernc.org/sqlite at
   v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 +
   golang.org/x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these).
   The NESTED webui module rides go 1.25.0 (wails v2.16.0's floor) + the
   wails require + the transitive requires the direct internal/app import
   pulls (indirect via the parent replace — ADR-0002 Am.5); the MAIN
   go.mod stays byte-unchanged. Then check 'git status' and 'git log
   --oneline -3': SESSION-049 landed T-81 in three commits (b5c9a27 core,
   0961ad9 rider, e6927b9 evidence) and PUSHED everything — 2377
   main-battery tests race-clean over 36 packages (2371 + 6) + 89 webui
   tests over 4. Verify git status shows up to date with origin/main.
   Use the PAT only ephemerally: the GIT_CONFIG_COUNT env-var form avoids
   persistence entirely — and SCRUB IT if you ever CLONE with it.
   Re-verify the suite before work if anything looks off — the law pays
   in bugs, not just safety.

2. Read, in order: the repo README.md (SESSION-049's status block) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md
   (your state card — §1 records SESSION-049's close) →
   docs/sessions/SESSION-049.md (the most recent report — the landing +
   the find + the blocked leg) → docs/poc-artifacts/
   t-81-release-workflow.md (the evidence + the re-run recipe §3 +
   the tagged-state comparison table) → .github/workflows/release.yml +
   build/mkzip/workflow_test.go + build/release.sh's header (the landed
   automation + its laws) → doc 10 §7/§8/§9 (T-82's checklist + the AV
   notes + §8's dated note) → doc 13's T-81/T-82/T-83 rows → doc 15 §6
   (AC-23/AC-24).

3. THIS SESSION = the T-81 dry-run completion (billing permitting) +
   T-82's preparable legs. Follow the landed shape: the engines
   main-battery-tested over hermetic seams + the pinned reference
   citations + the byte-level evidence. The dry-run rides the EXISTING
   test tag v0.0.0-t81-dryrun (already pointing at the fixed workflow
   at 0961ad9) — never a real version bump; the refusal leg (if run)
   rides v9.9.9-t81-refusal, deleted after. One slice done properly
   with proof beats two half-done; close cleanly at a natural boundary
   when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git
   (byte-level secret-scan before every commit; never write
   token-shaped strings even as fakes); no new dependencies without an
   ADR (the authorized runtime set is exactly modernc.org/sqlite
   v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; the MAIN go.mod stays byte-unchanged; the wails require +
   the indirect transitive requires stay in the NESTED webui module per
   ADR-0002 Am.5); surface blockers with evidence (hash-level, not
   display) instead of improvising past them; probe before claiming any
   module-boundary rule (ADR-0002 Am.4); fake every network-touching
   seam in tests at construction (THE HERMETIC-SEAM law); when the
   reference's own code sits behind an import wall, drive the
   reference's own BYTES by pinned line-range extraction — structural
   pins + the rev guard, never hand-transcription, never a stubbed wall
   (THE EXTRACT-AND-EXEC SCOPE LAW); "same semantics" means the two
   binaries agree on EVERY input to the shared protocol, edge branches
   included — audit the except ladders (THE CROSS-BINARY DISAGREEMENT
   LAW); the payload structs' JSON tags mirror the Python dict keys and
   the frontend reads the tagged names (THE FIELD-TAG LAW — extend the
   payload field-tags test with every new payload); a transition seat
   read+written by concurrent passes holds it under a mutex as ONE unit
   (THE SENTINEL-SEAT RACE LAW — run the battery ≥3× before trusting a
   new concurrent seam; every live-map mutation holds the Settings
   mutex, the SaveSettings discipline); every pointer/optional field
   read is settled with a nil check (THE NIL-OPTIONAL DEREF); one owner
   per frontend surface (duplicate EventsOn listeners double-paint); a
   modal's action buttons live OUTSIDE the scrollable body; every new
   wails door JOINS the owning module's binding IMPORT statement (THE
   BINDING-IMPORT LAW — pin it by test); after every wails build, check
   'git diff --raw' for mode flips (the wailsjs committed baseline IS
   755) and never commit a stray binary; gofmt -l after every Go edit
   batch (the Edit tool's space-vs-tab corruption — gofmt -w restores);
   every git/go command carries its own cd AND its own
   PATH=/home/z/.local/go/bin export (the sandbox shell resets both
   between calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation
   (the chosen path's own acceptance), commit + push everything (ask
   me for the GitHub PAT the moment a push is due — never store it
   anywhere; never end the session push-pending), write
   docs/sessions/SESSION-050.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language
   report plus the prompt I should copy to start SESSION-051.

One slice done properly with proof beats two half-done; close cleanly
at a natural boundary when capacity runs out.
```
