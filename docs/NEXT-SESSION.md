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
  SESSION-060; the owner has explicitly NO HASTE). **SESSION-052
  (commit `44a8bdb`) closed the last autonomous Phase-7 performance
  item: the AC-17 benchmark (Go vault-scan 14.6 ms cold vs Python
  32.8 ms same-box — the ≤ 50% gate PASSES; the record:
  `docs/poc-artifacts/ac17-go-vaultscan-benchmark.md`).**
  **SESSION-053–057 (docs-only) each ran the decision tree's path (c)
  — the menu surfaced, nothing moved. SESSION-058: the owner
  DELEGATED the pick ("you choose; no haste to start the app; only
  tell me if you really need me to test the app") — the session took
  the deferred-debts path and FOUND + CLOSED A REAL PARITY GAP: the
  CLI `--chrome-retry` LIVE leg never existed (the records had
  claimed it landed at SESSION-030 — commit `6ccf184`'s message named
  `hands.go`, its stat never touched it; the honest "lands with T-62"
  boundary card + its pinning test stood for 28 sessions; the
  un-ported `test_noteredo` TestCLIRouting selections were exactly
  the tests that would have caught it). SESSION-058 landed the real
  leg (the preflight-before-`__config_path__` quirk, the
  `ChromeDeliverFn`/`RunBatchFn` nil-defaulted doors, THE ROUTING LAW,
  the first-three-errors [:90]-runes composition, the exit codes),
  with the four ported selections + the post-batch warning test, the
  golden's two new sections, and the mutation teeth; the CI incident
  (the mutation preamble's raw open-Chrome ladder on a Chrome-bearing
  runner) root-caused + fixed green the same session.**
  **SESSION-059 (path (c) under the standing delegation) found +
  closed a SECOND REAL PARITY GAP: the master scans' delivered law.
  The reference INLINES its own delivered probe in BOTH
  `scan_master_hand_rows` + `scan_master_redo_rows` (a queue entry
  marked consumed OR its suggested page file present — isfile, no
  size check); the Go port had reused the BATCH-side collector
  (`CollectDelivered` — skip-consumed + size>0), so a CONSUMED row
  was invisible to both Go scans: the redo scan missed THE OWNER'S
  EXACT CASE (consumed + a half-fetched `_review` note — the very
  false success the v0.57.0 feature exists to redo) and the hand
  scan re-asked a consumed row for a Chrome tab. The fix:
  `decommission.go`'s own `deliveredURLSet` (the reference's inlined
  law; MetaStr/MetaTruthy exported from handdelivery) + the redo
  scan's probe-panic-ladder (the reference's try/except twin — the
  closed-DB probe panics in Go by store.DB's Close law); the port
  (`internal/websitepipeline/noteredo_test.go` — 11 TestRedoScan
  selections + the iconcolumn §7 never-re-ask leg) failed 4/12 on
  its first run (the empirical proof) and finished 12/12 green
  post-fix; a webui flake repaired honestly along the way
  (`TestHeroUIFanfareChainGate`'s bare async-delivery read → the
  rig's own waiter). The battery: 2394 + 89 = 2483.**
  **SESSION-060 (path (c) a third time) landed the harvest's
  probe-panic-ladder cousin — the remainder SESSION-059 recorded,
  closed with the same proof chain: the reference's
  `harvest_hand_rows` wraps `is_dismissed` + `processed_row` in ONE
  inner try/except ("a broken probe harvests nothing"), the Go
  `HarvestHandRows` handled only the error shape — the closed-DB
  probe PANICKED (store.DB's Close law → `SQL()` nil → `QueryRow` on
  the nil `*sql.DB`), so every direct caller inherited a hard crash
  where the reference answers 0 and keeps the gesture; the ported
  selection (the reference's own
  `test_a_broken_state_probe_harvests_nothing` setup — the
  WOULD-harvest shape — over a closed DB) panicked on its first run
  with the exact predicted stack, the state=nil hermetic leg
  likewise; the fix is the same ladder twin the redo scan rides +
  the two pinning selections
  (`internal/websitepipeline/handharvest_test.go`), all green
  post-fix. The battery: 2396 + 89 = 2485.** THE HONEST-ZIP FIND
  (standing): the LIVE `v0.92.0-go.1` release carries the
  CLI-milestone zip (12 files, no GUI exe); the desktop two-exe zip
  (13 files) rides the NEXT milestone release through the proven
  workflow. **What remains in Phase 7 is the OWNER's: T-82's live
  checklist (no haste — their words) + the milestone release
  decision.**
- **Push state: CLEAN** — SESSION-058 pushed its close `e43e023` +
  the CI-fix `584120b` + the postscript `2f41a49` (all green);
  SESSION-059's close push (`a743917`) verified GREEN at
  SESSION-060's open. SESSION-060 pushes its close (the harvest
  probe-ladder + the two pinning selections + the records) — go-ci
  to verify green at SESSION-061's open (a red is a REAL failure
  now; watch the Actions tab on pushes as normal CI hygiene).
- **The repo is PUBLIC** (the owner's SESSION-050 instruction): pushes
  still need the PAT (ephemeral, as always); Actions minutes are
  unlimited; the no-secrets law is public-facing. **go-ci: GREEN on
  the tree through `a743917`; SESSION-060's close push is the
  latest — verify its run green at the open.**
- **The environment recipe** (the 55th wipe was absorbed at
  SESSION-060's open — the standard recipe held, including
  SESSION-057's archive-listing correction): work repo at
  /home/z/my-project/second-brain-project (**clone it ANONYMOUSLY — it
  is public; the credentials block's GitHub URL may point at the
  PRIVATE Python reference `assadigit/GitCurator` — do not confuse the
  two**; the /home/z/second-brain-project symlink is DOWN while
  symlink(2) is blocked — use the real path), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (must match
  testdiff/PINNED-GITCURATOR-REV; clone with the PAT then SCRUB it —
  the GIT_CONFIG_COUNT env-var form avoids persistence entirely),
  toolchain go1.27.2 → /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` —
  64 hex chars; verify the locally-computed hash against the COMMITTED
  pin in build/TOOLCHAIN.md Am.2 — the pin is the authority; note
  go.dev/dl redirects — curl needs `-L`; **the SECOND verify side is
  the ARCHIVE listing `https://go.dev/dl/?mode=json&include=all` —
  SESSION-057's correction: the stable `?mode=json` listing no longer
  carries go1.27.2 now that go1.27.3 has shipped; the archive form
  does, and the MATCH ×2 held**), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the plain
  'python3 -m venv' form when ensurepip is present, else
  `--without-pip --copies`), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only;
  reinstalled at SESSION-060's rebuild), and the rootless
  webkit tree (~/.local/lib-wails — **DOWN; last rebuilt at
  SESSION-047 (the FIFTEENTH rebuild); never rebuilt since
  (SESSION-048–060 all ran no Linux GUI build — the windows legs are
  CGO-free and the full battery [main + webui] runs without it,
  re-proven at SESSION-060); the SIXTEENTH rebuild is pending:
  rebuild it FIRST only when the session runs a LINUX GUI build** —
  per webui/README.md + docs/poc-artifacts/t-70-gate.md +
  build/TOOLCHAIN.md Am.3 + Am.4: 144 debs → dpkg -x → the 42 .pc
  prefix rewrites → the 26 dev links, hardlink fallback expected.
  [Ordinal corrected at SESSION-055: the true count is fifteen done,
  sixteenth pending.]) The MAIN go.mod: modernc.org/sqlite at
  v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net
  v0.48.0 + golang.org/x/text v0.32.0 direct, and the go directive at
  1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not bump
  any of these). The NESTED webui module rides go 1.25.0 (wails
  v2.16.0's floor) + the wails require + the transitive requires
  (ADR-0002 Am.5; the MAIN go.mod stays byte-unchanged). **The
  downloads must run in FOREGROUND CHUNKS** — the sandbox kills
  backgrounded children when the call returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them): **THE HARVEST-COUSIN RECORD
  (SESSION-060 — the probe-panic-ladder now rides in BOTH the redo
  scan AND `HarvestHandRows`; the state=nil hermetic shape is the
  same ladder class; the caller-side never-fires belts (the
  reference logs "Hand-harvest pass skipped" if one ever fires, the
  Go master seam swallows silently) are unreachable by construction
  on both sides — recorded, not changed)**; **THE DELIVERED-LAW
  RECORD (SESSION-059 — the master scans' delivered probe is
  consumed-OR-landed, INLINED in the reference in both scans; NEVER
  reuse the batch-side collector `CollectDelivered` for it; the redo
  scan's probe calls ride a panic-ladder — the reference's
  try/except twin — and since SESSION-060 the harvest rides the same
  twin)**; **THE MULTIEDIT-TAB CORRUPTION, THIRD STRIKE (SESSION-060
  — the Edit tool re-expanded a whole Go file's tabs to spaces AND
  the Write tool wrote a NEW Go file with spaces-for-tabs; both
  caught by gofmt -l after the edit batch, both repaired by
  gofmt -w [HEAD is gofmt-clean], the post-repair git diff verified
  to be EXACTLY the edit region; patch Go source with PYTHON on the
  real bytes; sed is safe for identifier renames; also: the eat
  struck a COMMITTED card once — card v59's "battery ain + webui]"
  for "battery [main + webui]", fixed in v60)**; **THE ASYNC-READ
  FLAKE LESSON (SESSION-059 — an async delivery count read bare is a
  race under multi-package CPU contention; the rig's own
  waitDeliverCalls before the exact count; the repair keeps the
  teeth)**; **THE LIVE-LEG RECORD (SESSION-058 — the deferred-debt
  list can point at a REAL parity gap, not just a test gap: settle
  every "already landed" claim against the commit's own stat before
  trusting it)**; **THE 55TH-WIPE RECIPE (SESSION-060 — held without
  correction)**; **THE WEBKIT-ORDINAL CORRECTION (SESSION-055 — the
  pending rebuild is the SIXTEENTH, last done at SESSION-047)**;
  **THE AC-17 BENCHMARK RECORD (SESSION-052)**: performance rows
  close with same-input + same-output proofs and same-box medians on
  both engines; **THE HONEST-ZIP RECORD (SESSION-051)**; **THE
  STALE-COMPARISON-TARGET LAW (SESSION-050)**; **THE PUBLIC-REPO
  RECORD (SESSION-050)**; **THE TAG-STATE BUILD LAW (SESSION-049,
  re-proven on the runner at 050)**; **THE -X-UNDER-TRIMPATH LAW
  (SESSION-048)**; **THE VERIFY-THE-VERIFIER LESSON
  (SESSION-048/049/050/052/059/060 — the renderer eats characters
  on displays AND the tool echoes; settle every suspect reading with
  byte-level counts)**; **THE DEFAULT-ICON FIND (SESSION-048)**;
  **THE FLAG-CLEAR LAW (SESSION-047)**; **THE REFERENCE'S-OWN-
  ARITHMETIC LAW (SESSION-046/048)**; **THE VLM MISREAD LEDGER (FOUR
  incidents: 042/043/044/045; 046–060 ran clean)**; **THE
  BINDING-IMPORT LAW (SESSION-044)**; **THE RACE-PASS DIVIDEND
  (SESSION-045)**; **THE EXTRACT-AND-EXEC SCOPE LAW (SESSION-041)**;
  **THE CROSS-BINARY DISAGREEMENT LAW**; **THE SENTINEL-SEAT RACE LAW
  (SESSION-040)**; **THE RETROACTIVE FIELD-TAG AUDIT (SESSION-040)**;
  **THE NIL-OPTIONAL DEREF (SESSION-040)**; **THE
  ONE-OWNER-PER-SURFACE law (SESSION-040)**; **THE FIELD-TAG LAW
  (SESSION-038)**; **THE DISPLAY-PIPELINE CORRUPTION LESSON
  (SESSION-038 — the Edit tool's SPACE-vs-TAB; gofmt -l after every
  edit batch; python-patch Go source when the tool rendering is
  suspect)**; **THE SYMLINK-BLOCK FALLBACK (TOOLCHAIN.md Am.4)**;
  **THE RENDER-CONTEXT LAW (SESSION-037)**; **THE VLM RATE-LIMIT
  WINDOW (SESSION-039)**; **THE PINNED-VERDICT-ROW LESSON
  (SESSION-039)**; **THE ONE-SLICE-PROPERLY RECORD (036–060)**; and
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

## 2. Next session: SESSION-061 — the Phase-7 finish, owner-input driven (the owner's standing delegation recorded)

The implementation side of Phase 7 is DONE (including the AC-17
benchmark at SESSION-052, the chrome-retry live leg at SESSION-058,
the master-scans delivered-law fix at SESSION-059, and the harvest's
probe-panic-ladder cousin at SESSION-060). **The owner's standing
input (SESSION-058, verbatim): "Continue your work please. do based
on our documentation. you choose. I dont have urge and haste to
start the app. you can keep working. only tell me if you really need
me to test the app."** — a delegation of the pick, no haste on T-82
or the release, and nothing currently needs the owner's testing.
SESSION-061 follows the decision tree (in priority order):

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
- **(c) Otherwise (the standing delegation)**: continue the
  deferred-debts work autonomously, ONE slice done properly with
  proof — the sharpened §9 inventory (now TWO named slices): the
  GUI-side hero/wiring equivalents (test_noteredo's
  TestHeroRedoRouting + TestRedoPassWiring at the Go app seam, where
  the app seam already covers the guards), or the note_state SURGERY
  path's CRLF round-trip differential. Surface the Phase-8 menu only
  if the owner asks; close cleanly at a natural boundary when
  capacity runs out.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-060.md` — the most recent report (the
   harvest cousin's find + fix + the pinning selections; the battery
   2396 + 89).
2. `docs/PARITY-REPORT.md` — the owner-reviewable state (§9's
   sharpened not-done list is the standing menu — two slices left).
3. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — the owner's
   fill-in surface (the intake target if results arrive).
4. `docs/poc-artifacts/t-81-release-workflow.md` §5 — the proven
   release path (the recipe if the owner says ship it).
5. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + the Phase-8 future-work register (the tail).
6. `README.md` (the front page + the SESSION-060 status entry).

## 4. Known constraints / owner actions for SESSION-061

- **OWNER ACTION #1 (no haste — the owner's own words): T-82's live
  execution** — the owner's Windows machine runs the nine checklist
  items against the current `v0.92.0-go.1` zip (the CLI legs) or the
  next release (the full two-exe set) and fills §2; the signed
  checklist closes T-82. Nothing blocks on it.
- **OWNER ACTION #2: the milestone release decision** — say the word
  and the session cuts it (see §2b).
- **OWNER ACTION #3: the Phase-8 pick** — only if the owner wants it
  (doc 12's register); otherwise the standing delegation continues
  the deferred-debts polish.
- **The repo is PUBLIC** (SESSION-050) — never commit secrets
  (public-facing guarantee); pushes still need the ephemeral PAT.
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (last rebuilt at SESSION-047 — the
  fifteenth; the SIXTEENTH rebuild is pending) — rebuild it FIRST
  only if the session runs a LINUX GUI build (the windows legs are
  CGO-free; the full battery runs without it — re-proven at
  SESSION-060).
- The dissolution call (the owner's, ADR-0002 Am.5); the
  vanilla-vs-vite call (the owner's); the batch-scorecard modal
  question (the owner's); the v0.91.0 zip-parity reading (the
  owner's, minor); the answering-window belt + the T-71
  merged-verdict question (SESSION-039's record).
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — carried in every
  card since the ADR-0002 Am.2 record landed at SESSION-012; the
  anchored form): the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the wailsjs baseline IS 755) and never commit a stray binary (nor a
  generated notes.md).
- **Deferred test debt on the books** (the sharpened §9 inventory —
  now TWO slices, evidence-recorded, not blockers): the noteredo
  remainder's GUI-side hero/wiring equivalents (TestHeroRedoRouting +
  TestRedoPassWiring — the app seam already covers the guards); the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke (T-82's checklist, the
  owner's — no haste) + the bootstrapper leg (Phase 7 checklist,
  doc 10 §7).
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

## 5. THE PROMPT TO COPY (SESSION-061)

```
You are starting SESSION-061 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 7's IMPLEMENTATION SIDE IS DONE, WITH NO OPEN ROW
LEFT: T-80 COMPLETE; T-81 COMPLETE (the dry-run green on a real runner + the
cross-machine byte-identity proof); T-83 COMPLETE (the docs); T-82 PARTIAL
(items 1–8 are the owner's live Windows legs — the owner has explicitly NO
HASTE: their SESSION-058 words were "I dont have urge and haste to start the
app ... only tell me if you really need me to test the app"); AC-17 CLOSED
(SESSION-052). SESSION-053–057 (docs-only) each ran path (c) — the menu
surfaced, nothing moved. SESSION-058: the owner DELEGATED the pick — the
session took the deferred-debts path and FOUND + CLOSED A REAL PARITY GAP (the
CLI --chrome-retry LIVE leg never existed; landed with the four ported
selections + the golden's two sections + the mutation teeth; the CI incident
root-caused + fixed green the same session). SESSION-059 (path (c) under the
standing delegation) FOUND + CLOSED A SECOND REAL PARITY GAP: the master
scans' delivered law — the reference INLINES its own delivered probe in BOTH
scan_master_hand_rows + scan_master_redo_rows (consumed OR the page file
present, isfile, no size check), but the Go port had reused the BATCH-side
collector (skip-consumed + size>0), so a CONSUMED row was invisible to both
Go scans: the redo scan missed THE OWNER'S EXACT CASE (consumed + a
half-fetched _review note) and the hand scan re-asked a consumed row. The
fix: decommission.go's own deliveredURLSet (MetaStr/MetaTruthy exported from
handdelivery) + the redo scan's probe-panic-ladder (the reference's try/except
twin — the closed-DB probe PANICS in Go by store.DB's own Close law); the
port (internal/websitepipeline/noteredo_test.go — 11 TestRedoScan selections
+ the iconcolumn §7 never-re-ask leg) failed 4/12 on its first run (the
empirical proof) and finished 12/12 green post-fix; a webui flake repaired
honestly along the way (TestHeroUIFanfareChainGate's bare async-delivery read
→ the rig's own waiter). SESSION-060 (path (c) a third time) landed the
harvest's probe-panic-ladder cousin — the remainder SESSION-059 recorded,
closed with the same proof chain: the reference's harvest_hand_rows wraps
is_dismissed + processed_row in ONE inner try/except ("a broken probe
harvests nothing"), the Go HarvestHandRows handled only the error shape, the
closed-DB probe PANICKED (store.DB's Close law → SQL() nil → QueryRow on the
nil *sql.DB), the ported selection (the reference's own broken-probe setup —
the WOULD-harvest shape — over a closed DB) panicked on its first run with
the exact predicted stack, the state=nil hermetic leg likewise; the fix is
the same ladder twin the redo scan rides + the two pinning selections
(internal/websitepipeline/handharvest_test.go), all green post-fix. The
battery: 2396 main + 89 webui = 2485. The records fixed in the same commit
(spec-core-website-pipeline's dated correction, the parity report §4.2 + §9
[the harvest cousin CLOSED; the §9 remainder is now TWO slices: the GUI-side
hero/wiring equivalents + the CRLF differential], doc 12's tail entry, the
README front-page count + the SESSION-060 entry). go-ci: GREEN through
a743917; SESSION-060's close push is the latest push — verify its run green
at the open (a red is a REAL failure now). What remains in Phase 7 is the
OWNER's: T-82's live checklist (no haste) + the milestone release decision.

1. The environment is REBUILDABLE, not persistent (fifty-five wipe events
across SESSION-005..060 — SESSION-060 absorbed the 55th at its open; if the
sandbox wiped again, rebuild per the committed recipes). NOTE the standing
lessons: downloads and long suites must run in FOREGROUND CHUNKS (the sandbox
kills backgrounded children when the call returns); this sandbox's
tool-output rendering occasionally drops characters — settle every suspect
reading with byte-level counts/hex before "fixing" anything (THE MULTIEDIT-TAB
CORRUPTION struck a THIRD time at SESSION-060: the Edit tool re-expanded a
whole Go file's tabs to spaces AND the Write tool wrote a NEW Go file with
spaces-for-tabs — gofmt -l after every edit batch catches both, gofmt -w
restores because HEAD is gofmt-clean, and the post-repair git diff must be
verified to be EXACTLY the edit region; patch with python reading the real
bytes; sed is safe for identifier renames; ALSO: the eat struck a COMMITTED
card once — card v59's "battery [main + webui]" shipped as "battery ain +
webui]", fixed in v60); the go.dev tarball download needs curl -L (the /dl/
URL redirects; a 75-byte stub means you forgot it); the toolchain sha256's
SECOND verify side is the ARCHIVE listing (curl
'https://go.dev/dl/?mode=json&include=all' — the stable ?mode=json listing no
longer carries go1.27.2 since go1.27.3 shipped; SESSION-057's correction; the
committed pin in build/TOOLCHAIN.md Am.2 stays the authority); the VLM CLI
rate-limits (429) under rapid calls AND misreads enabled-vs-disabled states
(FOUR incidents — never "fix" code off a suspect VLM verdict); 'ln -s'
refuses AND kills the whole compound bash call at this boot — never put it in
a compound command (the venv needs --without-pip --copies on ensurepip-less
boots; the webkit tree's 26 dev links need Am.4 hardlinks); THE
WEBKIT-ORDINAL CORRECTION (055 — the pending webkit rebuild is the SIXTEENTH;
count actual rebuilds from the session records, never the card's own
ordinal); THE LIVE-LEG RECORD (058 — settle every "already landed" claim
against the commit's own stat before trusting it; the deferred-debt list can
point at a REAL parity gap, not just a test gap — 059 PROVED IT A SECOND
TIME); THE HARVEST-COUSIN RECORD (060 — the probe-panic-ladder now rides in
BOTH the redo scan AND HarvestHandRows; state=nil is the same ladder class;
the caller-side never-fires belts are unreachable by construction on both
sides — recorded, not changed); THE DELIVERED-LAW RECORD (059 — the master
scans' delivered probe is consumed-OR-landed, INLINED in both reference
scans; never reuse the batch-side collector; the redo scan's probe calls
ride the panic-ladder); THE ASYNC-READ FLAKE LESSON (059 — an async delivery
count read bare is a race under multi-package CPU contention; wait for it
before the exact count); THE LADDER-STUB LESSON (058 — stub the open-Chrome
ladder everywhere a hand-delivery command runs: the CI runner SHIPS Chrome,
the sandbox does not); THE STALE-COMPARISON-TARGET LAW (050); THE TAG-STATE
BUILD LAW (049, re-proven on the runner at 050); THE -X-UNDER-TRIMPATH LAW
(048); THE DEFAULT-ICON FIND (048); THE BINDING-IMPORT LAW (044); THE
FLAG-CLEAR LAW (047); THE REFERENCE'S-OWN-ARITHMETIC LAW (046/048); THE
AC-17 BENCHMARK RECORD (052 — performance rows close with same-input +
same-output proofs and same-box medians on both engines). FIRST verify with
evidence: work repo at /home/z/my-project/second-brain-project (CLONE
ANONYMOUSLY — it is public; beware: a credentials block's GitHub URL may
point at the PRIVATE Python reference assadigit/GitCurator — do not confuse
the two; the work repo is assadigit/second-brain-project), read-only
reference /home/z/my-project/GitCurator @ 98f4fc7 (must match
testdiff/PINNED-GITCURATOR-REV; clone with the PAT then SCRUB it — the
GIT_CONFIG_COUNT env-var form avoids persistence entirely), toolchain
go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
chars; verify the locally-computed hash against the COMMITTED pin in
build/TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the transcription
corruption), venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the
plain 'python3 -m venv' form when ensurepip is present, else
--without-pip --copies), the wails CLI v2.16.0
(GOBIN=/home/z/.local/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
— build-time only), and the rootless webkit tree (~/.local/lib-wails — DOWN;
last rebuilt at SESSION-047 [the fifteenth]; the SIXTEENTH rebuild is
pending: rebuild it FIRST only when the session runs a LINUX GUI build — the
windows cross-compile legs are CGO-free, and the full battery [main + webui]
runs WITHOUT it, re-proven at SESSION-060 — per webui/README.md +
docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4: 144 debs →
dpkg -x → the 42 .pc prefix rewrites → the 26 dev links, hardlink fallback
expected). The MAIN go.mod: modernc.org/sqlite at v1.36.1,
bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text
v0.32.0 direct, and the go directive at 1.24.1 (the tls-client-forced floor
— ADR-0002 Am.2; do not bump any of these). The NESTED webui module rides go
1.25.0 (wails v2.16.0's floor) + the wails require + the transitive requires
(indirect via the parent replace — ADR-0002 Am.5); the MAIN go.mod stays
byte-unchanged. THE REPO IS PUBLIC: pushes still need the ephemeral PAT
(GIT_CONFIG_COUNT env-var form — never persisted; SCRUB if you ever CLONE
with it); Actions minutes are unlimited; the no-secrets law is public-facing.
Then check 'git status' and 'git log --oneline -3': SESSION-060's close is
the HEAD (the harvest probe-ladder + the two pinning selections + the
records — then a743917 SESSION-059's close, then 2f41a49 SESSION-058's
postscript). Re-verify the suite before work if anything looks off — the law
pays in bugs, not just safety.

2. Read, in order: docs/sessions/SESSION-060.md (the most recent report) →
docs/PARITY-REPORT.md (no open row left; §9's sharpened not-done list — TWO
slices — is the standing menu) → docs/poc-artifacts/t-82-clean-vm-checklist.md
(the owner's fill-in surface) → docs/poc-artifacts/t-81-release-workflow.md §5
(the proven release path) → docs/port-blueprint/12-migration-roadmap.md (the
Phase-7 records + the Phase-8 register at the tail) → README.md (the front
page + the SESSION-060 status entry).

3. THIS SESSION = the Phase-7 finish, OWNER-INPUT DRIVEN — the decision tree
in priority order, with the owner's standing delegation on record ("you
choose; no haste to start the app; only tell me if you really need me to test
the app"): (a) IF the owner brings T-82 results (the filled checklist):
record them into the scaffold + the AC-23 release report, flip doc 13's T-82
row → COMPLETE, run the Phase-7 exit evaluation (every acceptance with its
evidence, honest); (b) IF the owner says "ship it": the milestone release
session — bump VERSION (the owner picks the number), commit, push the tag
v$(cat VERSION) (THE TAG LAW + THE TAG-STATE BUILD LAW: the release numbers
live at the tagged checkout), watch the release workflow build + attach the
two-exe zip (proven end-to-end + cross-machine), verify the release notes'
table, write the release report per AC-23's scaffold (T-82's checklist then
runs against the real release); (c) OTHERWISE (the standing delegation):
continue the deferred-debts polish autonomously, ONE slice done properly
with proof — the sharpened §9 inventory's TWO remaining slices: the GUI-side
hero/wiring equivalents (test_noteredo's TestHeroRedoRouting +
TestRedoPassWiring at the Go app seam, where the app seam already covers the
guards) or the note_state SURGERY path's CRLF round-trip differential. One
slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.

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
(THE HERMETIC-SEAM law — including the open-Chrome ladder on any
hand-delivery run: the CI runner SHIPS Chrome, the sandbox does not); when
the reference's own code sits behind an import wall, drive the reference's
own BYTES by pinned line-range extraction — structural pins + the rev guard,
never hand-transcription, never a stubbed wall (THE EXTRACT-AND-EXEC SCOPE
LAW); "same semantics" means the two binaries agree on EVERY input to the
shared protocol, edge branches included — audit the except ladders (THE
CROSS-BINARY DISAGREEMENT LAW — SESSION-059's probe-panic-ladder and
SESSION-060's harvest cousin were exactly this class); the payload structs'
JSON tags mirror the Python dict keys and the frontend reads the tagged names
(THE FIELD-TAG LAW); a transition seat read+written by concurrent passes
holds it under a mutex as ONE unit (THE SENTINEL-SEAT RACE LAW — run the
battery ≥3× before trusting a new concurrent seam); every pointer/optional
field read is settled with a nil check (THE NIL-OPTIONAL DEREF); one owner
per frontend surface; a modal's action buttons live OUTSIDE the scrollable
body; every new wails door JOINS the owning module's binding IMPORT
statement (THE BINDING-IMPORT LAW — pin it by test); after every wails build,
check 'git diff --raw' for mode flips (the wailsjs committed baseline IS 755)
and never commit a stray binary; gofmt -l after every Go edit batch (the
Edit AND Write tools both corrupt tabs — the third strike at SESSION-060);
every git/go command carries its own cd AND its own
PATH=/home/z/.local/go/bin export (the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
chosen path's own acceptance), commit + push everything (ask me for the
GitHub PAT the moment a push is due — never store it anywhere; never end the
session push-pending), write docs/sessions/SESSION-061.md, rewrite
docs/NEXT-SESSION.md with the next session's prompt, and end your reply with
a plain-language report plus the prompt I should copy to start SESSION-062.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
