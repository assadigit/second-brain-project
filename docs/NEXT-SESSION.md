# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker
  HTTP API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity.
- **Phase 2 CLOSED** (SESSION-006–010): T-20…T-25.
- **Phase 3 CLOSED** (SESSION-011–014): T-30…T-34.
- **Phase 4 CLOSED (SESSION-015–026)**: T-40…T-50 + the Phase-4 closer
  (T-51 removed by the owner's Option-B stamp).
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity
  release `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip
  asset sha-verified + round-trip verified).
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70
  gate through the T-79a…T-79f Path-B rows.
- **Phase 7 — THE DESKTOP MILESTONE RELEASE IS LIVE
  (SESSION-048–070)**: T-80 ✓ (the deterministic zip + the ritual),
  T-81 ✓ (the release workflow, dry-run green on a real runner + the
  cross-machine byte-identity proof + the refusal leg), T-83 ✓
  (SESSION-051: the docs), AC-17 ✓ (SESSION-052), **THE MILESTONE
  RELEASE CUT ✓ (SESSION-070, 2026-10-23 — the owner's post-close
  directive "The work is continued until GUI is ready" executed)**:
  the bump commit `a74b112` (VERSION + wails.json productVersion →
  `0.92.0-go.2`, go-ci run 39951244699 GREEN) → the tag
  `v0.92.0-go.2` riding it (THE TAG-STATE BUILD LAW) → the release
  workflow run **39951431147 ALL 8 STEPS GREEN** → the release
  **421804097 LIVE** ("GitCurator v0.92.0-go.2 — the Windows
  release") with the two-exe zip
  `GitCurator-v0.92.0-go.2-windows.zip` [13 files, 56,931,140 bytes
  in, 27,235,329 zipped, sha256 `78990ccb…`] — verified THREE ways
  (the downloaded asset's sha256 == the notes' number; the contents:
  both exes' sha256 == the notes' table + the buildinfo tag-stamp +
  the byte-exact VERSION member; **and the cross-machine
  byte-identity proof** — a full sandbox rebuild at the tagged
  checkout reproduced every number). The release report (AC-23's
  scaffold filled): `docs/poc-artifacts/release-v0.92.0-go.2.md`.
  **THE HONEST-ZIP FIND IS RESOLVED: the LIVE release list now
  carries the desktop two-exe zip with the new git+WebView2
  bootstrapper leg** (the older `v0.92.0-go.1` CLI-milestone zip
  stays listed alongside). T-82 PARTIAL (item 9's automation half
  green + SESSION-070's two legs against the real release; items
  1–8 + AV the owner's live Windows legs — the owner has explicitly
  NO HASTE: their SESSION-058 words were "I dont have urge and haste
  to start the app ... only tell me if you really need me to test
  the app"). **SESSION-071 (2026-10-23) re-verified the whole state
  under the standing delegation** (path (b): go-ci run 39953390474
  GREEN at the open on SESSION-070's close push `c241bc8`; the
  battery re-verified green in the exact committed shape 2398 + 95
  = 2493, race-clean, gofmt/vet/build clean in both modules; the
  live release + its two-exe zip re-verified; the T-82 §1 slots
  verified still empty — no intake; the menu re-surfaced compactly;
  the clean close).** What remains in Phase 7: the owner's T-82 run
  + the Phase-7 exit evaluation (the gate: the signed checklist
  closing T-82 — or the owner's explicit deferral call). The debt
  books are EMPTY (the §9 inventory since SESSION-062). **The
  Phase-8 menu stands surfaced in full from SESSION-064** (doc 12's
  register; re-surfaced compactly at SESSION-065…069 + 071); the
  owner's pick is the only mover — no autonomous implementation
  slice exists.
- **Push state: CLEAN** — the bump `a74b112` → run 39951244699
  GREEN; the tag → the release run 39951431147 GREEN with the
  release live; SESSION-070's close push `c241bc8` → run 39953390474
  GREEN (verified at SESSION-071's open); **SESSION-071's close push
  (docs-only: the report + this card + doc 12's entry + the README
  entry) is the latest — verify its go-ci run green at SESSION-072's
  open (a red is a REAL failure; watch the Actions tab on pushes as
  normal CI hygiene)**.
- **The repo is PUBLIC** (the owner's SESSION-050 instruction): pushes
  still need the PAT (ephemeral, as always — re-inline it per call;
  the sandbox shell does not carry exports between tool calls);
  Actions minutes are unlimited; the no-secrets law is public-facing.
  **go-ci: GREEN on the tree through SESSION-071** (the ledger: the
  release run 39951431147 + `a74b112` → 39951244699 + `c241bc8` →
  39953390474; the close push is the latest: verify at the open).
- **The environment recipe** (the 66th wipe was absorbed at
  SESSION-071's open — the standard recipe held, zero corrections):
  work repo at /home/z/my-project/second-brain-project (**clone it
  ANONYMOUSLY — it is public; the credentials block's GitHub URL may
  point at the PRIVATE Python reference `assadigit/GitCurator` — do
  not confuse the two**; scrub any persisted token from the remote
  URL — the ephemeral `-c http.extraheader` clone form avoids
  persistence entirely), read-only reference /home/z/my-project/GitCurator
  @ `98f4fc7` (must match testdiff/PINNED-GITCURATOR-REV; the
  GIT_CONFIG_COUNT / `-c` env-var forms avoid persistence entirely),
  toolchain go1.27.2 → /home/z/.local/go (go.dev tarball — extract
  WITHOUT --strip-components so the tree lands at .local/go; sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
  — 64 hex chars; verify the locally-computed hash against the
  COMMITTED pin in build/TOOLCHAIN.md Am.2 — the pin is the authority;
  note go.dev/dl redirects — curl needs `-L`; the SECOND verify side
  is the ARCHIVE listing `https://go.dev/dl/?mode=json&include=all`),
  venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the plain
  'python3 -m venv' form when ensurepip is present, else
  `--without-pip --copies`), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — **DOWN; last
  rebuilt at SESSION-047 (the FIFTEENTH rebuild); never rebuilt since
  (SESSION-048–071 all ran no Linux GUI build — the windows legs are
  CGO-free and the full battery runs without it, re-proven at
  SESSION-070/071; the SIXTEENTH rebuild is pending: rebuild it FIRST
  only when the session runs a LINUX GUI build** — per webui/README.md
  + docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4:
  144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
  hardlink fallback expected). The MAIN go.mod: modernc.org/sqlite at
  v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0
  + golang.org/x/text v0.32.0 direct, and the go directive at
  1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not bump
  any of these). The NESTED webui module rides go 1.25.0 (wails
  v2.16.0's floor) + the wails require + the transitive requires
  (ADR-0002 Am.5; the MAIN go.mod stays byte-unchanged). **The
  downloads must run in FOREGROUND CHUNKS** — the sandbox kills
  backgrounded children when the call returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them): **THE RELEASE-RECORD (SESSION-070 —
  the milestone release cut per the owner's directive: the bump →
  the tag on the bump commit → the workflow green → the release
  verified THREE ways incl. the cross-machine byte-identity proof;
  the reproduce-from-tag recipe is the release report's own section;
  every future release reuses the proven ritual verbatim)**; **THE
  MENU-SURFACED RECORD (SESSION-064, exercised at SESSION-065…069 +
  071 — the Phase-8 menu is surfaced in full [SESSION-064's report +
  the card §4]; with the debt books empty, the owner's pick is the
  only mover; a standing-delegation session re-verifies the state +
  re-surfaces the menu compactly + closes cleanly — never starts
  Phase-8 work unilaterally)**; **THE ENVIRONMENT-OWNED-TEXT LAW
  (SESSION-063)**; **THE BOOTSTRAPPER-LEG RECORD (SESSION-063 — the
  1-INSTALL WebView2 check is READ-ONLY; the Evergreen GUID is
  `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`; the leg now rides the
  LIVE zip)**; **THE DEBT-BOOKS-EMPTY RECORD (SESSION-063)**; **THE
  TEXT-MODE-READ LAW (SESSION-061)**; **THE MULTIEDIT-TAB
  CORRUPTION, FIFTH STRIKE (SESSION-063 — patch Go source with
  PYTHON on the real bytes and ANCHOR-ASSERT every patch; a wrong
  anchor guess must FAIL LOUD — SESSION-070's README move hit the
  non-unique-anchor variant in MARKDOWN: the edit landed after the
  wrong entry; settled with a byte-safe python move + grep
  verification — anchor assertions are for prose too)**; **THE
  HARVEST-COUSIN RECORD (SESSION-060)**; **THE DELIVERED-LAW RECORD
  (SESSION-059)**; **THE MULTIEDIT-TAB CORRUPTION, THIRD STRIKE
  (SESSION-060)**; **THE ASYNC-READ FLAKE LESSON (SESSION-059,
  SECOND STRIKE SESSION-068)**; **THE LIVE-LEG RECORD (SESSION-058)**;
  **THE 66TH-WIPE RECIPE (SESSION-071 — held without correction; the
  go tarball extracts to .local/go WITHOUT --strip-components)**;
  **THE WEBKIT-ORDINAL CORRECTION (SESSION-055 — the pending rebuild
  is the SIXTEENTH, last done at SESSION-047)**; **THE AC-17
  BENCHMARK RECORD (SESSION-052)**; **THE HONEST-ZIP RECORD
  (SESSION-051 — RESOLVED at SESSION-070: the live list carries the
  two-exe zip)**; **THE STALE-COMPARISON-TARGET LAW (SESSION-050)**;
  **THE PUBLIC-REPO RECORD (SESSION-050)**; **THE TAG-STATE BUILD
  LAW (SESSION-049, re-proven on the runner at 050 and on the REAL
  release at 070 — the CLI exe's buildinfo carries `mod …
  v0.92.0-go.2`)**; **THE -X-UNDER-TRIMPATH LAW (SESSION-048)**;
  **THE VERIFY-THE-VERIFIER LESSON (SESSION-048–071 — the renderer
  eats characters on displays AND the tool echoes; settle every
  suspect reading with byte-level counts; SESSION-070's README-tail
  assertion failed on a `the\nenvironment` line-wrap — the od dump
  settled it)**; **THE DEFAULT-ICON FIND (SESSION-048)**; **THE
  FLAG-CLEAR LAW (SESSION-047)**; **THE REFERENCE'S-OWN-ARITHMETIC
  LAW (SESSION-046/048)**; **THE VLM MISREAD LEDGER (FOUR incidents:
  042/043/044/045; 046–071 ran clean)**; **THE BINDING-IMPORT LAW
  (SESSION-044)**; **THE RACE-PASS DIVIDEND (SESSION-045)**; **THE
  EXTRACT-AND-EXEC SCOPE LAW (SESSION-041)**; **THE CROSS-BINARY
  DISAGREEMENT LAW**; **THE SENTINEL-SEAT RACE LAW (SESSION-040)**;
  **THE RETROACTIVE FIELD-TAG AUDIT (SESSION-040)**; **THE
  NIL-OPTIONAL DEREF (SESSION-040)**; **THE ONE-OWNER-PER-SURFACE
  law (SESSION-040)**; **THE FIELD-TAG LAW (SESSION-038)**; **THE
  DISPLAY-PIPELINE CORRUPTION LESSON (SESSION-038)**; **THE
  SYMLINK-BLOCK FALLBACK (TOOLCHAIN.md Am.4)**; **THE
  RENDER-CONTEXT LAW (SESSION-037)**; **THE VLM RATE-LIMIT WINDOW
  (SESSION-039)**; **THE PINNED-VERDICT-ROW LESSON (SESSION-039)**;
  **THE ONE-SLICE-PROPERLY RECORD (036–071)**; and the standing set
  (THE LIVE-CONFIG LAW; THE GROUP-POINTER LAW; THE
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
    every commit — sha256 hashes are fine, tokens are not; the
    repo-wide grep's hits must be INSPECTED before declaring them
    false positives — SESSION-070 proved its hits were the reference's
    own test-fixture placeholders); **THE BACKGROUND-PROCESS lesson**
    (long suites + downloads in FOREGROUND CHUNKS); **PUSH AFTER
    EVERY TASK** (ask for the PAT the moment a push is due; never
    store).

## 2. Next session: SESSION-072 — the post-release session (the decision tree, in priority order)

1. **(a) T-82 INTAKE** — if the owner's prompt carries their live
   Windows results (the nine checklist items run against the live
   `v0.92.0-go.2` zip): fill
   `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1's slots +
   `docs/poc-artifacts/release-v0.92.0-go.2.md`'s results table + the
   AV section with the owner's signed results; flip doc 13's T-82
   row; then **the Phase-7 exit evaluation, honestly** against doc
   12's gates (T-80…T-83 + AC-17 + the release: what is proven, what
   the honest boundary is) and record the verdict; then the Phase-8
   question is the owner's.
2. **(b) the standing delegation** ("you choose; no haste to start
   the app") — re-verify the state (the go-ci run at the open — a
   red is a REAL failure; the battery if the environment wiped, the
   exact committed shape 2398 + 95 = 2493), re-surface the Phase-8
   menu compactly (the full menu stands from SESSION-064), and close
   cleanly at the natural boundary. Never start Phase-8 work
   unilaterally.
3. **Blockers**: if anything red surfaces (a CI failure, a broken
   release link), STOP and surface it with evidence per the standing
   laws; record it in the card + report and hand the decision back.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-071.md` — the post-release session's
   report (the state re-verified + the menu re-surfaced compactly;
   the compact menu rides its §Owner actions).
2. `docs/poc-artifacts/release-v0.92.0-go.2.md` — the AC-23 release
   report (the live release's verified numbers; the owner's T-82
   fill-in slots live here + the checklist doc §1).
3. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — §1 (the nine
   items' slots) + §2 (the scaffold, now instantiated) + §4 (the
   dated note — resolved by the live release).
4. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + the release acceptances (AC-23) + the SESSION-070/071
   tail entries.
5. `README.md` (the front page — pointing at the live desktop
   milestone).
6. `docs/sessions/SESSION-070.md` — the release session's report
   (the ritual's every step + the verification evidence; only if the
   release ritual's details are needed). The full Phase-8 menu's
   full form stands from SESSION-064's report.

## 4. Known constraints / owner actions for SESSION-072

- **OWNER ACTION #1 (no haste — the owner's own words): T-82's live
  execution** — the owner's Windows machine runs the nine checklist
  items against the LIVE `v0.92.0-go.2` two-exe zip (item 1 carries
  the full installer check: git + the WebView2 probe + the exe's
  first-launch one-click offer) and fills §1 + the release report's
  table + the AV section; the signed checklist closes T-82 and
  unlocks the Phase-7 exit evaluation. Item 9's automation half is
  already proven (SESSION-070's download-sha256 + cross-machine
  reproduction legs); the owner's own spot-check is optional.
  Nothing blocks on it.
- **OWNER ACTION #2: the Phase-7 exit evaluation timing** — rides the
  T-82 intake (or the owner's explicit deferral call).
- **OWNER ACTION #3: the Phase-8 pick** — only if the owner wants it
  (the menu is surfaced in full at SESSION-064 + compactly at
  SESSION-065–069 + 071; the register's options + the standing small
  calls are listed there); otherwise the standing delegation
  re-surfaces the menu compactly and closes cleanly.
- **The repo is PUBLIC** (SESSION-050) — never commit secrets
  (public-facing guarantee); pushes still need the ephemeral PAT
  (re-inline per call — the shell does not carry exports between
  tool calls).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (last rebuilt at SESSION-047 — the
  fifteenth; the SIXTEENTH rebuild is pending) — rebuild it FIRST
  only if the session runs a LINUX GUI build (the windows legs are
  CGO-free; the full battery runs without it — re-proven at
  SESSION-070/071).
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
- **Deferred test debt on the books**: NONE. Only the owner's LIVE
  legs remain (T-82 items 1–8; the launchers' clean-VM smoke rides
  T-82 the same way).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB any persisted
  extraheader — the ephemeral `-c`/GIT_CONFIG_COUNT forms avoid it
  entirely).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit); tests never weakened; reality wins over specs
  (fix the spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-072)

```
You are starting SESSION-072 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE STATE: the desktop milestone release v0.92.0-go.2 is LIVE (cut at
SESSION-070; re-verified live at SESSION-071 — the two-exe zip GitCurator.exe +
GitCurator-cli.exe, 13 files, sha256-verified three ways incl. the cross-machine
byte-identity proof; the release report: docs/poc-artifacts/release-v0.92.0-go.2.md;
the T-82 §1 slots verified still empty at SESSION-071 — no intake yet).
SESSION-072 is the post-release session — the decision tree in priority order:
(a) if this prompt carries the owner's T-82 results (the nine clean-VM checklist
items run against the v0.92.0-go.2 zip), do the INTAKE: fill
docs/poc-artifacts/t-82-clean-vm-checklist.md §1 + the release report's results
table + the AV section with the owner's signed results, flip doc 13's T-82 row,
then run the PHASE-7 EXIT EVALUATION honestly against doc 12's gates (T-80…T-83
+ AC-17 + the release: what is proven, what the honest boundary is) and record
the verdict; (b) otherwise the standing delegation ("you choose; no haste to
start the app") holds: re-verify the state (the go-ci run at the open — a red is
a REAL failure; the battery if the environment wiped), re-surface the Phase-8
menu compactly (the full menu stands from SESSION-064), and close cleanly —
never starting Phase-8 work unilaterally. If anything blocks, STOP and surface
it with evidence per the standing laws; record it in the card + report and hand
the decision back.

Then follow docs/SESSION-PROTOCOL.md exactly: bootstrap (clone
https://github.com/assadigit/second-brain-project — the work repo, public;
read README.md, docs/SESSION-PROTOCOL.md, docs/NEXT-SESSION.md [the card v72],
the two most recent session reports [SESSION-071 is the latest; SESSION-070 is
the release report-carrying session], git log --oneline -15), orient (the
read-firsts: the card's §3), work (the standing laws: reality wins, tests never
weakened, no secrets ever, stay inside the phase, push after every task), close
(the honest gate evaluation, the commit + push, the SESSION-072 report, the
NEXT-SESSION.md rewrite, the worklog append), and report in chat with the
literal next prompt.
```
