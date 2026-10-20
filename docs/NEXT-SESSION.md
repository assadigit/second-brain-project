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
  full run green.
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70 gate
  through the T-79a…T-79f Path-B rows.
- **Phase 7 OPEN — T-80 COMPLETE (SESSION-048); T-81 COMPLETE
  (SESSION-049 landed + SESSION-050 completed the dry-run); T-82's
  PREPARABLE legs landed (SESSION-050), the live execution is the
  owner's; T-83 (the docs) is the next slice.** T-81's completion
  record: the owner made the repo PUBLIC ("Make it public, so you can
  have unlimited CI/CD" — free unlimited Actions minutes; a
  full-history secret scan ran before the flip: 1,415 blobs, zero
  token-shaped strings); the re-run of `39512685975` went green
  end-to-end on a real runner (all 8 steps; the ritual's own PASS line
  on the runner); the `[DRY-RUN]` draft+prerelease release carried the
  zip asset + the AC-23 notes; **the cross-machine proof: the runner's
  numbers == a full sandbox rebuild at the same tagged state
  (`0961ad9`+tag) — CLI `f2ef8362…`, GUI `99cc22e6…`, zip
  `6409352a…`, byte-identical across machines**; the refusal leg
  (`v9.9.9-t81-refusal`, run `39515838232`) failed EXACTLY at the
  tag-law step and created NO release; the test release + both test
  tags deleted after the proof (only the real `v0.92.0-go.1` remains).
  THE STALE-TARGET FIND (recorded in the evidence doc §2.1's
  correction): the §2.1 table was measured at `b5c9a27`+tag — the tag
  moved before the run; the CLI exe embeds `vcs.revision` per commit —
  NOT a reproducibility failure. T-82's preparable surface:
  `docs/poc-artifacts/t-82-clean-vm-checklist.md` (the nine §7 items
  as the owner's fill-in record + AC-23's release-report scaffold +
  the AV note with the `-w -s` test-before-adopt rule).
- **Push state: CLEAN** — SESSION-050 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
- **The repo is PUBLIC now (the owner's instruction at SESSION-050).**
  Consequences for every future session: pushes still need the PAT
  (ephemeral, as always); Actions minutes are unlimited (the billing
  block is structurally cured); the full history is world-readable —
  the standing no-secrets law is now a public-facing guarantee (the
  pre-flip scan: zero token-shaped strings in all 1,415 blobs; keep it
  that way — the byte-level secret-scan before every commit).
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
  legs are CGO-free — SESSION-048/049/050 needed no webkit) per
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
  - **THE STALE-COMPARISON-TARGET LAW (SESSION-050)**: a
    reproducibility table is pinned to its exact tree state (commit +
    tag position) — if the tag moves between the measurement and the
    comparison, the comparison target is STALE, not the build (the CLI
    exe embeds `vcs.revision`; the same-length revision shifts the
    bytes and even the zip's compressed size by a couple bytes while
    the file count and bytes-in stay identical). Always rebuild the
    comparison target at the runner's exact state before recording a
    divergence.
  - **THE PUBLIC-REPO RECORD (SESSION-050)**: the repo is public —
    Actions minutes unlimited (the billing block cannot recur);
    pushes still need the ephemeral PAT; the no-secrets law is now
    public-facing (the pre-flip full-history scan: 1,415 blobs, zero
    token-shaped strings; scan BEFORE any future visibility or
    artifact-publishing change).
  - **THE 45TH-WIPE RECIPE (SESSION-049)**; **THE TAG-STATE BUILD LAW
    (SESSION-049 — re-proven ON THE RUNNER at SESSION-050: the CLI
    exe's buildinfo carries mod version `v0.0.0-t81-dryrun` +
    `vcs.revision=0961ad90…` on both machines)**; **THE
    ACTIONS-BILLING BLOCK RECORD (SESSION-049 — RESOLVED at 050 by the
    public flip)**; **THE -X-UNDER-TRIMPATH LAW (SESSION-048)**; **THE
    VERIFY-THE-VERIFIER LESSON (SESSION-048/049 — held again at 050:
    the runner's first zip sha differed from the table, and the
    divergence was settled byte-level — the extracted exe's buildinfo —
    before any conclusion)**; **THE DEFAULT-ICON FIND (SESSION-048)**;
    **THE 43RD-WIPE RECIPE (SESSION-047)**; **THE FLAG-CLEAR LAW
    (SESSION-047)**; **THE REFERENCE'S-OWN-ARITHMETIC LAW
    (SESSION-046)**; **THE 42ND-WIPE RECIPE**; **THE VLM MISREAD
    LEDGER (FOUR incidents: 042/043/044/045; 046–050 ran clean)**;
    **THE BINDING-IMPORT LAW (SESSION-044)**; **THE RACE-PASS DIVIDEND
    (SESSION-045)**; **THE EXTRACT-AND-EXEC SCOPE LAW
    (SESSION-041)**; **THE CROSS-BINARY DISAGREEMENT LAW**; **THE
    SENTINEL-SEAT RACE LAW (SESSION-040)**; **THE RETROACTIVE
    FIELD-TAG AUDIT (SESSION-040)**; **THE NIL-OPTIONAL DEREF
    (SESSION-040)**; **THE ONE-OWNER-PER-SURFACE law (SESSION-040)**;
    **THE FIELD-TAG LAW (SESSION-038)**; **THE DISPLAY-PIPELINE
    CORRUPTION LESSON (SESSION-038 — the Edit tool's SPACE-vs-TAB;
    gofmt -l after every edit batch)**; **THE SYMLINK-BLOCK FALLBACK
    (TOOLCHAIN.md Am.4)**; **THE RENDER-CONTEXT LAW (SESSION-037)**;
    **THE VLM RATE-LIMIT WINDOW (SESSION-039)**; **THE
    PINNED-VERDICT-ROW LESSON (SESSION-039)**; **THE
    ONE-SLICE-PROPERLY RECORD (036–050)**; and the standing set (THE
    LIVE-CONFIG LAW; THE GROUP-POINTER LAW; THE DISSOLUTION-PROBE
    RECORD; THE GOLDEN-MARSHAL-ORDER TRAP; THE SUITE-REVERIFICATION
    DIVIDEND; THE WAILS-BUILD MODE FLIPS; THE HERMETIC-SEAM LAW; THE
    RESUMED-CONTEXT RE-VERIFICATION LAW; THE MODE-FLIP DEGRADATION
    LAW; THE MODULE-BOUNDARY PROBE LAW; THE FRONTEND-MODULE LESSON;
    the T-65 zip-composition resolution; the corpus-regeneration law;
    the dispatch-completion law; the state-chain + mutation-teeth
    laws; the context-death continuation recipe; the DSN-empty-path
    hazard; the fixed-length roots; the shallow-copy snapshot; the
    fake-clock arithmetic; the exception class-name tokens; the
    stdout-capture-helper limit).
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd AND its own PATH=/home/z/.local/go/bin export); **THE
    RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F FILE);
    **THE WRITE-TOOL REDACTION lesson** (byte-level secret-scan before
    every commit); **THE BACKGROUND-PROCESS lesson** (long suites +
    downloads in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for
    the PAT the moment a push is due; never store it).

## 2. Next session: SESSION-051 — T-83 (the docs) + the go-ci watch + T-82's owner-results intake

- **T-83 (doc 15 §7 — AC-25/AC-26)**: the Go README (the owner-readable
  front page: what shipped, how to run it, the parity state), the
  WINDOWS-QUICKSTART update (the Go launch path), and the parity
  report (per-task status, deviations with reasons, the ADR/spec
  updates where reality differed — plain language per SPEC.md §0).
  Most of the raw material already lives in doc 13's rows + the
  evidence docs + the session reports; T-83 assembles the
  owner-reviewable summary.
- **The go-ci watch**: SESSION-050's close push was the FIRST post-flip
  go-ci trigger (the two red runs on record predate the flip — billing
  refusals). Check the Actions tab at the start (the run for
  SESSION-050's close commit): green closes the silently-red record;
  a REAL red is now a real failure (the billing excuse is gone) —
  surface it with evidence, fix per the logs.
- **T-82's owner-results intake (if the owner ran the checklist)**:
  the owner's Windows execution fills
  `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1/§2 — record the
  results into the doc's scaffold + flip the T-82 row when the signed
  checklist lands.
- T-82/T-83 close Phase 7 (the milestone release decision is the
  owner's — formal releases happen only at roadmap milestones).

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-050.md` — the most recent report (T-81's
   completion + T-82's preparable legs).
2. `docs/poc-artifacts/t-81-release-workflow.md` §5 (+ §2.1's
   correction) — the green-run record + the cross-machine table.
3. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — the prepared
   surface for the owner's execution.
4. `docs/port-blueprint/13-implementation-task-list.md` — the T-82
   (PARTIAL) + T-83 rows.
5. `docs/port-blueprint/15-acceptance-criteria.md` §7 — AC-25/AC-26
   (T-83's acceptance).
6. The repo README's head (T-83 updates it) + `WINDOWS-QUICKSTART.md`.

## 4. Known constraints / owner actions for SESSION-051

- **OWNER ACTION #1 (when convenient): T-82's live execution** — the
  owner's Windows machine runs the nine checklist items
  (`docs/poc-artifacts/t-82-clean-vm-checklist.md` §1) against the
  current `v0.92.0-go.1` zip (or the next release) and fills the
  scaffold (§2); the signed checklist closes T-82.
- **The repo is PUBLIC** (SESSION-050) — see §1's record. Never commit
  secrets (now a public-facing guarantee); pushes still need the
  ephemeral PAT.
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (the 45th wipe took it; the SEVENTEENTH
  rebuild is pending) — rebuild it FIRST only if the session runs a
  LINUX GUI build (the windows legs are CGO-free; SESSION-048/049/050
  worked without it).
- The dissolution call (the owner's, ADR-0002 Am.5); the
  vanilla-vs-vite call (the owner's); the batch-scorecard modal
  question (the owner's); the v0.91.0 zip-parity reading (the owner's,
  minor); the answering-window belt + the T-71 merged-verdict question
  (SESSION-039's record).
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — forty-first session
  carrying it): the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the wailsjs baseline IS 755) and never commit a stray binary (nor a
  generated notes.md).
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke (T-82's checklist, the
  owner's) + the bootstrapper leg (Phase 7 checklist, doc 10 §7).
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

## 5. THE PROMPT TO COPY (SESSION-051)

```
You are starting SESSION-051 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 7 is OPEN — T-80 COMPLETE (SESSION-048); T-81
(the release workflow) COMPLETE (SESSION-049 landed + SESSION-050 completed
the dry-run: the owner made the repo PUBLIC — "Make it public, so you can
have unlimited CI/CD" — and the re-run of 39512685975 went green end-to-end
on a real runner: all 8 steps, the [DRY-RUN] draft+prerelease release with
the zip asset, the CROSS-MACHINE PROOF (the runner's numbers == a full
sandbox rebuild at the same tagged state 0961ad9+tag: CLI f2ef8362…, GUI
99cc22e6…, zip 6409352a… — byte-identical across machines), the refusal leg
(v9.9.9-t81-refusal failed EXACTLY at the tag-law step, NO release created),
the test release + both test tags deleted); THE STALE-TARGET FIND recorded
(the evidence doc §2.1's table was measured at b5c9a27+tag — the tag moved
before the run; the CLI exe embeds vcs.revision per commit — NOT a
reproducibility failure). T-82's PREPARABLE legs landed (SESSION-050:
docs/poc-artifacts/t-82-clean-vm-checklist.md — the nine §7 items as the
owner's fill-in record + AC-23's release-report scaffold + the AV note with
the -w -s test-before-adopt rule); the LIVE execution is the owner's (no
Windows box in the sandbox — standing since SESSION-001). SESSION-051 =
T-83 (the docs: the Go README, the WINDOWS-QUICKSTART update, the parity
report — doc 15 §7's AC-25/AC-26, plain language per SPEC.md §0) + the
go-ci watch (SESSION-050's close push was the FIRST post-flip go-ci
trigger; the two red runs on record predate the flip — billing refusals;
check the Actions tab at the start: green closes the silently-red record, a
real red is now a real failure) + T-82's owner-results intake if the owner
ran the checklist.

1. The environment is REBUILDABLE, not persistent (forty-five wipe events
   across SESSION-005..049 — SESSION-050 ran on the intact post-049
   environment; if the sandbox wiped it again, rebuild per the committed
   recipes). NOTE the standing lessons: downloads and long suites must run
   in FOREGROUND CHUNKS (the sandbox kills backgrounded children when the
   call returns); this sandbox's tool-output rendering occasionally drops
   characters — settle every suspect reading with byte-level counts/hex
   before "fixing" anything (the Edit tool's space-vs-tab corruption struck
   042/044/046/047/048 writing Go source — gofmt -l after every edit batch,
   gofmt -w restores because HEAD is gofmt-clean; patch with python reading
   the real bytes; 048's sharper edge: the renderer ate '[main]' into
   'ain]' ON THE VERIFICATION ITSELF; 049's cousin: a zip-member extraction
   filter matched the SETUP BAT instead of the exe; 050's cousin: the
   renderer ate '[main]' again in a config grep — the line was
   'branches: [main]'); the VLM CLI rate-limits (429) under rapid calls AND
   misreads enabled-vs-disabled states (FOUR incidents — never "fix" code
   off a suspect VLM verdict); 'ln -s' refuses AND kills the whole compound
   bash call at this boot — never put it in a compound command (the venv
   needs --without-pip --copies on ensurepip-less boots; the webkit tree's
   26 dev links need Am.4 hardlinks); THE STALE-COMPARISON-TARGET LAW (050:
   a reproducibility table is pinned to its exact tree state — commit + tag
   position; if the tag moved, the TARGET is stale, not the build); THE
   TAG-STATE BUILD LAW (049, re-proven on the runner at 050); THE
   -X-UNDER-TRIMPATH LAW (048); THE DEFAULT-ICON FIND (048); THE
   BINDING-IMPORT LAW (044); THE FLAG-CLEAR LAW (047); THE
   REFERENCE'S-OWN-ARITHMETIC LAW (046/048). FIRST verify with evidence:
   work repo at /home/z/my-project/second-brain-project (the
   /home/z/second-brain-project symlink is DOWN while symlink(2) is blocked
   — use the real path), read-only reference /home/z/my-project/GitCurator
   @ 98f4fc7 (must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64
   hex chars; go.dev's ?mode=json has ROTATED past 1.27.2: verify the
   locally-computed hash against the COMMITTED pin in build/TOOLCHAIN.md
   Am.2 — the pin is the authority; BEWARE the transcription corruption),
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the plain
   'python3 -m venv' form when ensurepip is present, else
   --without-pip --copies), the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install
   github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only), and
   the rootless webkit tree (~/.local/lib-wails — DOWN after the 45th
   wipe; the SEVENTEENTH rebuild is pending: rebuild it FIRST only when
   the session runs a LINUX GUI build — the windows cross-compile legs are
   CGO-free, SESSION-048/049/050 needed no webkit — per webui/README.md +
   docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4: 144
   debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
   hardlink fallback expected). The MAIN go.mod: modernc.org/sqlite at
   v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 +
   golang.org/x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these). The
   NESTED webui module rides go 1.25.0 (wails v2.16.0's floor) + the
   wails require + the transitive requires the direct internal/app import
   pulls (indirect via the parent replace — ADR-0002 Am.5); the MAIN
   go.mod stays byte-unchanged. THE REPO IS PUBLIC (the owner's
   SESSION-050 instruction): pushes still need the ephemeral PAT; Actions
   minutes are unlimited; the no-secrets law is now public-facing. Then
   check 'git status' and 'git log --oneline -3': SESSION-050 landed the
   T-81 completion + T-82 preparable docs (docs-only session; the close
   battery 36 + 4 packages ok, zero FAILs — the committed 2377 + 89
   shape). Use the PAT only ephemerally: the GIT_CONFIG_COUNT env-var
   form avoids persistence entirely — and SCRUB IT if you ever CLONE with
   it. Re-verify the suite before work if anything looks off — the law
   pays in bugs, not just safety.

2. Read, in order: the repo README.md (SESSION-050's status block) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-050's close) → docs/sessions/
   SESSION-050.md (the most recent report) → docs/poc-artifacts/
   t-81-release-workflow.md §5 + §2.1's correction (the green-run record
   + the cross-machine table) → docs/poc-artifacts/
   t-82-clean-vm-checklist.md (the prepared surface for the owner's
   execution) → doc 13's T-82/T-83 rows → doc 15 §7 (AC-25/AC-26) →
   WINDOWS-QUICKSTART.md (T-83 updates it).

3. THIS SESSION = T-83 (the docs: the Go README's owner-readable front
   page — what shipped, how to run it, the parity state; the
   WINDOWS-QUICKSTART update for the Go launch path; the parity report —
   per-task status, deviations with reasons, the ADR/spec updates where
   reality differed, plain language per SPEC.md §0) + the go-ci watch
   (check the Actions tab FIRST: the run for SESSION-050's close commit
   is the first post-flip go-ci run; green closes the record, a real red
   is a real failure now) + T-82's owner-results intake if the owner ran
   the checklist. The raw material lives in doc 13's rows + the evidence
   docs + the session reports — T-83 assembles the owner-reviewable
   summary. One slice done properly with proof beats two half-done;
   close cleanly at a natural boundary when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git
   (byte-level secret-scan before every commit; never write
   token-shaped strings even as fakes — the repo is PUBLIC now); no new
   dependencies without an ADR (the authorized runtime set is exactly
   modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0,
   x/text v0.32.0 — do not bump any; the MAIN go.mod stays
   byte-unchanged; the wails require + the indirect transitive requires
   stay in the NESTED webui module per ADR-0002 Am.5); surface blockers
   with evidence (hash-level, not display) instead of improvising past
   them; probe before claiming any module-boundary rule (ADR-0002 Am.4);
   fake every network-touching seam in tests at construction (THE
   HERMETIC-SEAM law); when the reference's own code sits behind an
   import wall, drive the reference's own BYTES by pinned line-range
   extraction — structural pins + the rev guard, never hand-transcription,
   never a stubbed wall (THE EXTRACT-AND-EXEC SCOPE LAW); "same
   semantics" means the two binaries agree on EVERY input to the shared
   protocol, edge branches included — audit the except ladders (THE
   CROSS-BINARY DISAGREEMENT LAW); the payload structs' JSON tags mirror
   the Python dict keys and the frontend reads the tagged names (THE
   FIELD-TAG LAW — extend the payload field-tags test with every new
   payload); a transition seat read+written by concurrent passes holds it
   under a mutex as ONE unit (THE SENTINEL-SEAT RACE LAW — run the
   battery ≥3× before trusting a new concurrent seam; every live-map
   mutation holds the Settings mutex, the SaveSettings discipline); every
   pointer/optional field read is settled with a nil check (THE
   NIL-OPTIONAL DEREF); one owner per frontend surface (duplicate
   EventsOn listeners double-paint); a modal's action buttons live
   OUTSIDE the scrollable body; every new wails door JOINS the owning
   module's binding IMPORT statement (THE BINDING-IMPORT LAW — pin it by
   test); after every wails build, check 'git diff --raw' for mode flips
   (the wailsjs committed baseline IS 755) and never commit a stray
   binary; gofmt -l after every Go edit batch (the Edit tool's
   space-vs-tab corruption — gofmt -w restores); every git/go command
   carries its own cd AND its own PATH=/home/z/.local/go/bin export (the
   sandbox shell resets both between calls; the NESTED webui module needs
   its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation
   (the chosen path's own acceptance), commit + push everything (ask
   me for the GitHub PAT the moment a push is due — never store it
   anywhere; never end the session push-pending), write
   docs/sessions/SESSION-051.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language
   report plus the prompt I should copy to start SESSION-052.

One slice done properly with proof beats two half-done; close cleanly
at a natural boundary when capacity runs out.
```
