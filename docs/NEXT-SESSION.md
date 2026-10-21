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
- **Phase 4 CLOSED (SESSION-015–026)**: T-40…T-50 + the Phase-4 closer
  (T-51 removed by the owner's Option-B stamp).
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity
  release `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip
  asset sha-verified + round-trip verified).
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70 gate
  through the T-79a…T-79f Path-B rows.
- **Phase 7 — THE IMPLEMENTATION SIDE IS DONE (SESSION-048–052)**:
  T-80 ✓ (the deterministic zip + the ritual), T-81 ✓ (the release
  workflow, dry-run green on a real runner + the cross-machine
  byte-identity proof + the refusal leg), T-83 ✓ (SESSION-051: the
  docs — `docs/PARITY-REPORT.md` (AC-25), the README front page +
  the QUICKSTART §0 (AC-26), commit `e58b7e4`), T-82 PARTIAL (the
  preparable surface landed at SESSION-050; items 1–8 are the owner's
  live Windows legs — the checklist slots were still empty at
  SESSION-054). **SESSION-052 (commit `44a8bdb`) closed the last
  autonomous Phase-7 item: the AC-17 benchmark — the Go vault-scan
  wall-time measured on the T-04 synthetic corpus (`testdiff/gobench`,
  the new rig; Go cold 14.6 ms vs Python 32.8 ms same-box = 44.5%,
  41.4% vs the recorded baseline 35.3 ms — the ≤ 50% gate PASSES;
  cache-warm 5.3 vs 16.9; 2.25×; same-input tree-hash-equal +
  same-output census-identical proven; the record:
  `docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`). The parity
  report now has NO open performance row.** **SESSION-053 (docs-only)
  and SESSION-054 (docs-only) each ran the decision tree's path (c):
  the owner brought no T-82 results and no "ship it" — the menu
  surfaced (this card §2/§4 + the session reports), the 48th and 49th
  wipes absorbed at their opens, the battery re-verified green both
  times (2377 + 89, the exact committed shape); nothing else
  moved.** THE HONEST-ZIP FIND (stated in all three T-83 docs): the
  LIVE `v0.92.0-go.1` release carries the CLI-milestone zip (12
  files, no GUI exe — cut at SESSION-032 before Phase 6); the desktop
  two-exe zip (13 files) rides the NEXT milestone release through the
  proven workflow. **What remains in Phase 7 is the OWNER's: T-82's
  live checklist + the milestone release decision.**
- **Push state: CLEAN** — SESSION-054 committed + pushed its docs-only
  close (the ephemeral GIT_CONFIG_COUNT PAT form; never stored).
  Verify `git status` in sync at the start.
- **The repo is PUBLIC** (the owner's SESSION-050 instruction): pushes
  still need the PAT (ephemeral, as always); Actions minutes are
  unlimited; the no-secrets law is public-facing. **go-ci is GREEN on
  the current tree** (runs `39517044092`, `39517366335`,
  `39522289819`, `39526755160` at `44a8bdb`, and `39529633876` at
  `35eb065` — all success). SESSION-054's docs-only close push is the
  latest push — verify its run green at the session's open (a red is
  a REAL failure now; watching the Actions tab on pushes is normal CI
  hygiene).
- **The environment recipe** (unchanged; the 49th wipe was absorbed at
  SESSION-054's open): work repo at
  /home/z/my-project/second-brain-project (**clone it ANONYMOUSLY — it
  is public; the credentials block's GitHub URL may point at the
  PRIVATE Python reference `assadigit/GitCurator` — do not confuse the
  two**; the /home/z/second-brain-project symlink is DOWN while
  symlink(2) is blocked — use the real path), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (must match
  testdiff/PINNED-GITCURATOR-REV; clone with the PAT then SCRUB it),
  toolchain go1.27.2 → /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` —
  64 hex chars; verify the locally-computed hash against the COMMITTED
  pin in build/TOOLCHAIN.md Am.2 — the pin is the authority; note
  go.dev/dl redirects — curl needs `-L`), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the plain
  `python3 -m venv` form when ensurepip is present, else
  `--without-pip --copies`), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — DOWN since the
  47th wipe; the EIGHTEENTH rebuild is pending: rebuild it FIRST only
  when the session runs a LINUX GUI build — the windows cross-compile
  legs are CGO-free, and the full battery (main + webui) runs WITHOUT
  it, re-proven at SESSION-054 — per webui/README.md +
  docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4:
  144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
  hardlink fallback expected). The MAIN go.mod: modernc.org/sqlite at
  v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0
  + golang.org/x/text v0.32.0 direct, and the go directive at 1.24.1
  (the tls-client-forced floor — ADR-0002 Am.2; do not bump any of
  these). The NESTED webui module rides go 1.25.0 (wails v2.16.0's
  floor) + the wails require + the transitive requires (ADR-0002
  Am.5; the MAIN go.mod stays byte-unchanged). **The downloads must
  run in FOREGROUND CHUNKS** — the sandbox kills backgrounded children
  when the call returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them): **THE 49TH-WIPE RECIPE (SESSION-054 —
  the standard rebuild held without correction again; the go.dev
  download needs `-L`)**; **THE AC-17 BENCHMARK RECORD
  (SESSION-052)**: performance rows close with same-input +
  same-output proofs (tree-hash the corpora, cross-scan the census)
  and same-box medians on both engines — never a stale-baseline-only
  comparison; **THE HONEST-ZIP RECORD (SESSION-051)**: the live
  release's contents must be checked via the API before any doc says
  "how to run it"; **THE 46TH-WIPE RECIPE (SESSION-051)**; **THE
  STALE-COMPARISON-TARGET LAW (SESSION-050)**; **THE PUBLIC-REPO
  RECORD (SESSION-050)**; **THE 45TH-WIPE RECIPE (SESSION-049)**;
  **THE TAG-STATE BUILD LAW (SESSION-049, re-proven on the runner at
  050)**; **THE ACTIONS-BILLING BLOCK RECORD (SESSION-049 — RESOLVED
  at 050 by the public flip)**; **THE -X-UNDER-TRIMPATH LAW
  (SESSION-048)**; **THE VERIFY-THE-VERIFIER LESSON
  (SESSION-048/049/050/052 — the renderer ate '[main 44a8bdb]' at
  SESSION-052's commit display too; the -F FILE message + the log line
  byte-verified fine)**; **THE DEFAULT-ICON FIND (SESSION-048)**;
  **THE 43RD-WIPE RECIPE (SESSION-047)**; **THE FLAG-CLEAR LAW
  (SESSION-047)**; **THE REFERENCE'S-OWN-ARITHMETIC LAW
  (SESSION-046)**; **THE 42ND-WIPE RECIPE**; **THE VLM MISREAD LEDGER
  (FOUR incidents: 042/043/044/045; 046–054 ran clean)**; **THE
  BINDING-IMPORT LAW (SESSION-044)**; **THE RACE-PASS DIVIDEND
  (SESSION-045)**; **THE EXTRACT-AND-EXEC SCOPE LAW (SESSION-041)**;
  **THE CROSS-BINARY DISAGREEMENT LAW**; **THE SENTINEL-SEAT RACE LAW
  (SESSION-040)**; **THE RETROACTIVE FIELD-TAG AUDIT (SESSION-040)**;
  **THE NIL-OPTIONAL DEREF (SESSION-040)**; **THE
  ONE-OWNER-PER-SURFACE law (SESSION-040)**; **THE FIELD-TAG LAW
  (SESSION-038)**; **THE DISPLAY-PIPELINE CORRUPTION LESSON
  (SESSION-038 — the Edit tool's SPACE-vs-TAB; gofmt -l after every
  edit batch)**; **THE SYMLINK-BLOCK FALLBACK (TOOLCHAIN.md Am.4)**;
  **THE RENDER-CONTEXT LAW (SESSION-037)**; **THE VLM RATE-LIMIT
  WINDOW (SESSION-039)**; **THE PINNED-VERDICT-ROW LESSON
  (SESSION-039)**; **THE ONE-SLICE-PROPERLY RECORD (036–054)**; and
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
    every commit — sha256 hashes are fine, tokens are not); **THE
    BACKGROUND-PROCESS lesson** (long suites + downloads in FOREGROUND
    CHUNKS); **PUSH AFTER EVERY TASK** (ask for the PAT the moment a
    push is due; never store it).

## 2. Next session: SESSION-055 — the Phase-7 finish, owner-input driven

The implementation side of Phase 7 is DONE (including the AC-17
benchmark at SESSION-052 — the parity report has no open row left;
SESSION-053 and SESSION-054 surfaced the menu with no owner input
arriving). SESSION-055 follows the owner's input; the decision tree
(in priority order):

- **(a) If the owner brings T-82 results** (the filled checklist
  `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1/§2): record the
  results into the doc's scaffold + the release report per AC-23,
  flip doc 13's T-82 row → COMPLETE, and run the **Phase-7 exit
  evaluation** (the honest gate: every Phase-7 acceptance with its
  evidence).
- **(b) If the owner says "ship it"** (the milestone release): the
  release session — bump `VERSION` (the owner picks the number; the
  desktop milestone suggests `v0.92.0-go.2` or a `v1.0.0-go` shape —
  the owner's call), commit, push the tag `v$(cat VERSION)` (the tag
  law), watch the release workflow build + attach the two-exe zip
  (the ritual is proven end-to-end + cross-machine), verify the
  release notes' table, and write the release report per AC-23's
  scaffold (T-82's checklist then runs against the real release).
  THE TAG-STATE BUILD LAW applies: the release numbers live at the
  tagged checkout.
- **(c) Otherwise**: surface the menu honestly and let the owner
  pick — Phase-8 scoping from doc 12's future-work register (the
  dormant cloud quartet revival / the gotd revival / Linux/macOS
  builds / localization — the owner's pick), or small polish from the
  deferred-debts list (the recorded test debts are in the parity
  report §9). One slice done properly with proof beats two
  half-done; close cleanly at a natural boundary when capacity runs
  out.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-054.md` — the most recent report (path (c)
   again: the menu surfaced; the 49th wipe; the battery re-verified).
2. `docs/PARITY-REPORT.md` — the owner-reviewable state (§9's
   not-done list is the standing menu).
3. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — the owner's
   fill-in surface (the intake target if results arrive).
4. `docs/poc-artifacts/t-81-release-workflow.md` §5 — the proven
   release path (the recipe if the owner says ship it).
5. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + the Phase-8 future-work register (the tail).
6. `README.md` (the front page + the SESSION-054 status entry).

## 4. Known constraints / owner actions for SESSION-055

- **OWNER ACTION #1: T-82's live execution** — the owner's Windows
  machine runs the nine checklist items
  (`docs/poc-artifacts/t-82-clean-vm-checklist.md` §1) against the
  current `v0.92.0-go.1` zip (the CLI legs) or the next release (the
  full two-exe set) and fills §2; the signed checklist closes T-82.
- **OWNER ACTION #2: the milestone release decision** — say the word
  and SESSION-055 cuts it (see §2b).
- **OWNER ACTION #3 (if neither): the Phase-8 pick** — doc 12's
  future-work register (the dormant cloud quartet / gotd / Linux/macOS
  builds / localization) or the polish list.
- **The repo is PUBLIC** (SESSION-050) — never commit secrets
  (public-facing guarantee); pushes still need the ephemeral PAT.
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (down since the 47th wipe; the
  EIGHTEENTH rebuild is pending) — rebuild it FIRST only if the
  session runs a LINUX GUI build (the windows legs are CGO-free; the
  full battery runs without it — re-proven at SESSION-054).
- The dissolution call (the owner's, ADR-0002 Am.5); the
  vanilla-vs-vite call (the owner's); the batch-scorecard modal
  question (the owner's); the v0.91.0 zip-parity reading (the
  owner's, minor); the answering-window belt + the T-71
  merged-verdict question (SESSION-039's record).
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — forty-fifth
  session carrying it): the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the wailsjs baseline IS 755) and never commit a stray binary (nor a
  generated notes.md).
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke (T-82's checklist, the
  owner's) + the bootstrapper leg (Phase 7 checklist, doc 10 §7).
  (AC-17's Go-side wall-time is CLOSED — SESSION-052.)
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

## 5. THE PROMPT TO COPY (SESSION-055)

```
You are starting SESSION-055 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 7's IMPLEMENTATION SIDE IS DONE, WITH NO OPEN ROW
LEFT (SESSION-048–052): T-80 COMPLETE (the deterministic zip + the double-build
ritual); T-81 COMPLETE (SESSION-049 landed + SESSION-050 completed the dry-run:
green on a real runner, the [DRY-RUN] release with the zip asset, the
CROSS-MACHINE BYTE-IDENTITY PROOF, the refusal leg, the cleanup); T-83 COMPLETE
(SESSION-051, commit e58b7e4: the docs — docs/PARITY-REPORT.md (the
plain-language implementation report, AC-25), the README front page, the
WINDOWS-QUICKSTART §0 launch path, doc 13's T-83 row, doc 15 §7's dated note,
doc 12's entry); T-82 PARTIAL (the preparable surface landed at SESSION-050;
items 1–8 are the owner's live Windows legs — the checklist slots were still
empty at SESSION-054); AC-17 CLOSED (SESSION-052, commit 44a8bdb: the AC-17
benchmark — the Go vault-scan wall-time measured on the T-04 synthetic corpus
via the new testdiff/gobench rig; Go cold 14.6 ms vs Python 32.8 ms same-box
(44.5%) and 35.3 ms recorded baseline (41.4%) — the ≤ 50% gate PASSES;
cache-warm 5.3 vs 16.9 ms; ~90.5k vs ~40.2k notes/s (2.25×); same-input
(tree-hash equal over 1,314 files) + same-output (the census identical,
notes_seen 1319) proven; the record: docs/poc-artifacts/ac17-go-vaultscan-benchmark.md;
the parity report §5/§9 updated — NO open performance row remains). SESSION-053
and SESSION-054 (both docs-only) each ran the decision tree's path (c): the
owner brought no T-82 results and no "ship it" — the menu surfaced both times
(T-82's live legs / the milestone release / Phase-8 scoping / small polish from
the deferred-debts list), the 48th and 49th wipes absorbed at their opens, the
battery re-verified green both times (2377 + 89, the exact committed shape);
nothing else moved. THE HONEST-ZIP FIND (stated in all three T-83 docs): the
LIVE v0.92.0-go.1 release carries the CLI-milestone zip (12 files, NO GUI exe
— cut at SESSION-032 before Phase 6); the desktop two-exe zip (13 files) rides
the NEXT milestone release through the proven workflow. go-ci: GREEN on the
current tree (runs 39517044092, 39517366335, 39522289819, 39526755160 at
44a8bdb, and 39529633876 at 35eb065 — all success; SESSION-054's docs-only
close push is the latest push — verify its run green at the open; a red is a
REAL failure now; watch the Actions tab on pushes as normal CI hygiene). What
remains in Phase 7 is the OWNER's: T-82's live checklist + the milestone
release decision.

1. The environment is REBUILDABLE, not persistent (forty-nine wipe events
across SESSION-005..054 — SESSION-054 absorbed the 49th at its open; if the
sandbox wiped again, rebuild per the committed recipes). NOTE the standing
lessons: downloads and long suites must run in FOREGROUND CHUNKS (the sandbox
kills backgrounded children when the call returns); this sandbox's tool-output
rendering occasionally drops characters — settle every suspect reading with
byte-level counts/hex before "fixing" anything (the Edit tool's space-vs-tab
corruption struck 042/044/046/047/048 writing Go source — gofmt -l after every
edit batch, gofmt -w restores because HEAD is gofmt-clean; patch with python
reading the real bytes; 048's sharper edge: the renderer ate '[main]' into
'ain]' ON THE VERIFICATION ITSELF; 051/052's cousins: the same eat on the
commit display — the commits were byte-verified fine via -F FILE + the log
line); the go.dev tarball download needs curl -L (the /dl/ URL redirects; a
75-byte stub means you forgot it); the VLM CLI rate-limits (429) under rapid
calls AND misreads enabled-vs-disabled states (FOUR incidents — never "fix"
code off a suspect VLM verdict); 'ln -s' refuses AND kills the whole compound
bash call at this boot — never put it in a compound command (the venv needs
--without-pip --copies on ensurepip-less boots; the webkit tree's 26 dev links
need Am.4 hardlinks); THE STALE-COMPARISON-TARGET LAW (050); THE TAG-STATE
BUILD LAW (049, re-proven on the runner at 050); THE -X-UNDER-TRIMPATH LAW
(048); THE DEFAULT-ICON FIND (048); THE BINDING-IMPORT LAW (044); THE
FLAG-CLEAR LAW (047); THE REFERENCE'S-OWN-ARITHMETIC LAW (046/048); THE AC-17
BENCHMARK RECORD (052 — performance rows close with same-input + same-output
proofs and same-box medians on both engines). FIRST verify with evidence: work
repo at /home/z/my-project/second-brain-project (CLONE ANONYMOUSLY — it is
public; beware: a credentials block's GitHub URL may point at the PRIVATE
Python reference assadigit/GitCurator — do not confuse the two; the
/home/z/second-brain-project symlink is DOWN while symlink(2) is blocked — use
the real path), read-only reference /home/z/my-project/GitCurator @ 98f4fc7
(must match testdiff/PINNED-GITCURATOR-REV; clone with the PAT then SCRUB it),
toolchain go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
chars; verify the locally-computed hash against the COMMITTED pin in
build/TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the transcription
corruption), venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the
plain 'python3 -m venv' form when ensurepip is present, else
--without-pip --copies), the wails CLI v2.16.0 (GOBIN=/home/z/.local/go/bin
go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only),
and the rootless webkit tree (~/.local/lib-wails — DOWN since the 47th wipe;
the EIGHTEENTH rebuild is pending: rebuild it FIRST only when the session runs
a LINUX GUI build — the windows cross-compile legs are CGO-free, and the full
battery [main + webui] runs WITHOUT it, re-proven at SESSION-054 — per
webui/README.md + docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 +
Am.4: 144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
hardlink fallback expected). The MAIN go.mod: modernc.org/sqlite at v1.36.1,
bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text
v0.32.0 direct, and the go directive at 1.24.1 (the tls-client-forced floor —
ADR-0002 Am.2; do not bump any of these). The NESTED webui module rides go
1.25.0 (wails v2.16.0's floor) + the wails require + the transitive requires
(indirect via the parent replace — ADR-0002 Am.5); the MAIN go.mod stays
byte-unchanged. THE REPO IS PUBLIC: pushes still need the ephemeral PAT
(GIT_CONFIG_COUNT env-var form — never persisted; SCRUB if you ever CLONE with
it); Actions minutes are unlimited; the no-secrets law is public-facing. Then
check 'git status' and 'git log --oneline -3': SESSION-054's close is the HEAD
(docs-only: the SESSION-054 report + the state card v55 + doc 12's entry + the
README's entry; below it 35eb065 SESSION-053 close + 28958bd SESSION-052
close). Re-verify the suite before work if anything looks off — the law
pays in bugs, not just safety.

2. Read, in order: docs/sessions/SESSION-054.md (the most recent report) →
docs/PARITY-REPORT.md (no open row left; §9's not-done list is the standing
menu) → docs/poc-artifacts/t-82-clean-vm-checklist.md (the owner's fill-in
surface) → docs/poc-artifacts/t-81-release-workflow.md §5 (the proven release
path) → docs/port-blueprint/12-migration-roadmap.md (the Phase-7 records + the
Phase-8 register at the tail) → README.md (the front page + the SESSION-054
status entry).

3. THIS SESSION = the Phase-7 finish, OWNER-INPUT DRIVEN — follow the
decision tree in priority order: (a) IF the owner brings T-82 results (the
filled checklist): record them into the scaffold + the AC-23 release report,
flip doc 13's T-82 row → COMPLETE, run the Phase-7 exit evaluation (every
acceptance with its evidence, honest); (b) IF the owner says "ship it": the
milestone release session — bump VERSION (the owner picks the number), commit,
push the tag v$(cat VERSION) (THE TAG LAW + THE TAG-STATE BUILD LAW: the
release numbers live at the tagged checkout), watch the release workflow build
+ attach the two-exe zip (proven end-to-end + cross-machine), verify the
release notes' table, write the release report per AC-23's scaffold (T-82's
checklist then runs against the real release); (c) OTHERWISE: surface the menu
honestly and let the owner pick — Phase-8 scoping from doc 12's register (the
dormant cloud quartet / the gotd revival / Linux builds / localization), or
small polish from the deferred-debts list. One slice done properly with proof
beats two half-done; close cleanly at a natural boundary when capacity runs
out.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
tests are never weakened; no secrets or owner data in git (byte-level
secret-scan before every commit; never write token-shaped strings even as
fakes — the repo is PUBLIC; sha256 hashes are fine); no new dependencies
without an ADR (the authorized runtime set is exactly modernc.org/sqlite
v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not bump any;
the MAIN go.mod stays byte-unchanged; the wails require + the indirect
transitive requires stay in the NESTED webui module per ADR-0002 Am.5);
surface blockers with evidence (hash-level, not display) instead of
improvising past them; probe before claiming any module-boundary rule
(ADR-0002 Am.4); fake every network-touching seam in tests at construction
(THE HERMETIC-SEAM law); when the reference's own code sits behind an import
wall, drive the reference's own BYTES by pinned line-range extraction —
structural pins + the rev guard, never hand-transcription, never a stubbed
wall (THE EXTRACT-AND-EXEC SCOPE LAW); "same semantics" means the two binaries
agree on EVERY input to the shared protocol, edge branches included — audit
the except ladders (THE CROSS-BINARY DISAGREEMENT LAW); the payload structs'
JSON tags mirror the Python dict keys and the frontend reads the tagged names
(THE FIELD-TAG LAW); a transition seat read+written by concurrent passes
holds it under a mutex as ONE unit (THE SENTINEL-SEAT RACE LAW — run the
battery ≥3× before trusting a new concurrent seam); every pointer/optional
field read is settled with a nil check (THE NIL-OPTIONAL DEREF); one owner per
frontend surface; a modal's action buttons live OUTSIDE the scrollable body;
every new wails door JOINS the owning module's binding IMPORT statement (THE
BINDING-IMPORT LAW — pin it by test); after every wails build, check 'git diff
--raw' for mode flips (the wailsjs committed baseline IS 755) and never commit
a stray binary; gofmt -l after every Go edit batch; every git/go command
carries its own cd AND its own PATH=/home/z/.local/go/bin export (the NESTED
webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
chosen path's own acceptance), commit + push everything (ask me for the GitHub
PAT the moment a push is due — never store it anywhere; never end the session
push-pending), write docs/sessions/SESSION-055.md, rewrite
docs/NEXT-SESSION.md with the next session's prompt, and end your reply with
a plain-language report plus the prompt I should copy to start SESSION-056.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
