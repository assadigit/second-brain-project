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
- **Phase 7 — THE IMPLEMENTATION SIDE IS DONE, AND THE DEBT BOOKS ARE
  EMPTY (SESSION-048–063)**: T-80 ✓ (the deterministic zip + the
  ritual), T-81 ✓ (the release workflow, dry-run green on a real
  runner + the cross-machine byte-identity proof + the refusal leg),
  T-83 ✓ (SESSION-051: the docs — `docs/PARITY-REPORT.md` (AC-25),
  the README front page + the QUICKSTART §0 (AC-26), commit
  `e58b7e4`), AC-17 CLOSED (SESSION-052), T-82 PARTIAL (items 1–8 are
  the owner's live Windows legs — the owner has explicitly NO HASTE:
  their SESSION-058 words were "I dont have urge and haste to start
  the app ... only tell me if you really need me to test the app").
  The deferred-debts polish ran out of named slices across
  SESSION-058–062, each landing with proof (the chrome-retry LIVE
  leg; the master scans' delivered law; the harvest's
  probe-panic-ladder cousin; the note_state surgery's CRLF text-mode
  differential; the GUI-side hero/wiring equivalents — the §9
  test-debt inventory emptied at SESSION-062). **SESSION-063 (path
  (c) a sixth time) landed the LAST named test-adjacent debt: the
  1-INSTALL BOOTSTRAPPER LEG (doc 10 §4's [verify] on the WebView2
  row resolved)** — `1-INSTALL.bat` gained the WebView2 Evergreen
  Runtime check (a READ-ONLY registry probe of the Evergreen
  runtime's own key, per-user then per-machine, value `pv`), the
  found/absent cards (the honest pointing: the desktop app offers
  the one-click download at its first launch), the conditional final
  card, pinned BOTH sides (`TestInstallBatChecksGitAndWebview2` — a
  strict sharpening of the git-only test; reg query is the ONE
  allowed reg verb — and the golden differential's 1-INSTALL goOnly
  contract) with the mutation teeth proven on the corrupted-GUID
  shape. **The t-70 poc plan's recorded GUID was a CORRUPTION** (a
  scrambled last group, standing wrong since SESSION-033; a probe
  carrying it reads a nonexistent key and reports a FALSE absent) —
  settled at byte level against learn.microsoft.com's distribution
  page (4 hits correct / 0 corrupted), corrected with a dated note.
  The exe's first-launch strategy settled at source level (no ritual
  change): wails v2.16.0's default `download` — one confirm dialog,
  then the Evergreen fetch (`go.microsoft.com/fwlink/p/?
  LinkId=2124703`). The battery: 2397 + 95 = 2492 (unchanged — the
  slice extended one test + the differential contract, not the
  rows). **THE RECORDED DEBT BOOKS ARE NOW EMPTY** — what remains on
  the bootstrapper leg is the LIVE half only (the owner's T-82 item
  1; the launchers' clean-VM smoke rides T-82 the same way).**
  **THE HONEST-ZIP FIND (standing): the LIVE `v0.92.0-go.1` release
  carries the CLI-milestone zip (12 files, no GUI exe, the OLD
  git-only 1-INSTALL checker); the desktop two-exe zip (13 files)
  with the new bootstrapper leg rides the NEXT milestone release
  through the proven workflow. What remains in Phase 7 is the
  OWNER's: T-82's live checklist (no haste — their words) + the
  milestone release decision.**
- **Push state: CLEAN** — SESSION-061's close push (`505f1f2`, run
  39696190997) verified GREEN at SESSION-062's open; SESSION-062's
  close push (`d2fa36c`, run 39833368088) verified GREEN at
  SESSION-063's open. SESSION-063 pushes its close (the bootstrapper
  leg + the records) — go-ci to verify green at SESSION-064's open
  (a red is a REAL failure now; watch the Actions tab on pushes as
  normal CI hygiene).
- **The repo is PUBLIC** (the owner's SESSION-050 instruction): pushes
  still need the PAT (ephemeral, as always); Actions minutes are
  unlimited; the no-secrets law is public-facing. **go-ci: GREEN on
  the tree through `d2fa36c`; SESSION-063's close push is the
  latest — verify its run green at the open.**
- **The environment recipe** (the 58th wipe was absorbed at
  SESSION-063's open — the standard recipe held): work repo at
  /home/z/my-project/second-brain-project (**clone it ANONYMOUSLY — it
  is public; the credentials block's GitHub URL may point at the
  PRIVATE Python reference `assadigit/GitCurator` — do not confuse the
  two**; scrub any persisted token from the remote URL — the clone
  with PAT then set-url anonymous form works), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (must match
  testdiff/PINNED-GITCURATOR-REV; the GIT_CONFIG_COUNT env-var form
  avoids persistence entirely), toolchain go1.27.2 → /home/z/.local/go
  (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
  — 64 hex chars; verify the locally-computed hash against the
  COMMITTED pin in build/TOOLCHAIN.md Am.2 — the pin is the authority;
  note go.dev/dl redirects — curl needs `-L`; the SECOND verify side
  is the ARCHIVE listing `https://go.dev/dl/?mode=json&include=all`
  — the stable `?mode=json` listing no longer carries go1.27.2 now
  that go1.27.3 has shipped; the archive form does), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only; the plain
  'python3 -m venv' form when ensurepip is present, else
  `--without-pip --copies`), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only;
  reinstalled at SESSION-063's rebuild), and the rootless
  webkit tree (~/.local/lib-wails — **DOWN; last rebuilt at
  SESSION-047 (the FIFTEENTH rebuild); never rebuilt since
  (SESSION-048–063 all ran no Linux GUI build — the windows legs are
  CGO-free and the full battery [main + webui] runs without it,
  re-proven at SESSION-063); the SIXTEENTH rebuild is pending:
  rebuild it FIRST only when the session runs a LINUX GUI build** —
  per webui/README.md + docs/poc-artifacts/t-70-gate.md +
  build/TOOLCHAIN.md Am.3 + Am.4: 144 debs → dpkg -x → the 42 .pc
  prefix rewrites → the 26 dev links, hardlink fallback expected.)
  The MAIN go.mod: modernc.org/sqlite at v1.36.1,
  bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 direct, and the go directive at
  1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not bump
  any of these). The NESTED webui module rides go 1.25.0 (wails
  v2.16.0's floor) + the wails require + the transitive requires
  (ADR-0002 Am.5; the MAIN go.mod stays byte-unchanged). **The
  downloads must run in FOREGROUND CHUNKS** — the sandbox kills
  backgrounded children when the call returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them): **THE BOOTSTRAPPER-LEG RECORD
  (SESSION-063 — the 1-INSTALL WebView2 check is READ-ONLY (reg
  query is the ONE allowed reg verb; the mutating verbs stay
  banned); the Evergreen GUID is
  `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}` (the t-70 plan's recorded
  form was a corruption — a probe carrying a wrong GUID reads a
  nonexistent key and reports a FALSE absent); the two halves of the
  promise are separate: the INSTALLER half checks + points, the
  RUNTIME half rides the exe's own first-launch path (wails
  v2.16.0's default `download` strategy: one confirm dialog + the
  Evergreen fetch — never promise "silent" for the portable zip's
  path; fully-silent chains were the NSIS artifact's `/silent
  /install`, which never shipped); the new leg rides the NEXT
  release's zip (the honest-zip law)**; **THE DEBT-BOOKS-EMPTY
  RECORD (SESSION-063 — the §9 inventory is empty of autonomous
  items; the next path-(c) work is the Phase-8 menu surfacing or
  whatever the owner's T-82/release inputs trigger)**; **THE
  TEXT-MODE-READ LAW (SESSION-061)**; **THE MULTIEDIT-TAB
  CORRUPTION, FIFTH STRIKE (SESSION-063 — both Go test files in one
  session, each re-expanded whole-file to spaces; gofmt -l after
  EVERY tool interaction caught both, gofmt -w repaired both, the
  post-repair diffs verified to be exactly the edit regions; one
  anchor-asserted patch refused on a wrong indent guess — nothing
  was patched, exactly the belt's purpose; patch Go source with
  PYTHON on the real bytes and ANCHOR-ASSERT every patch: a wrong
  anchor guess must FAIL LOUD)**; **THE HARVEST-COUSIN RECORD
  (SESSION-060)**; **THE DELIVERED-LAW RECORD (SESSION-059)**; **THE
  MULTIEDIT-TAB CORRUPTION, THIRD STRIKE (SESSION-060 — also: the
  eat struck a COMMITTED card once — card v59's "battery ain +
  webui]" for "battery ain + webui]", fixed in v60)**; **THE
  ASYNC-READ FLAKE LESSON (SESSION-059)**; **THE LIVE-LEG RECORD
  (SESSION-058 — settle every "already landed" claim against the
  commit's own stat before trusting it)**; **THE 58TH-WIPE RECIPE
  (SESSION-063 — held without correction)**; **THE WEBKIT-ORDINAL
  CORRECTION (SESSION-055 — the pending rebuild is the SIXTEENTH,
  last done at SESSION-047)**; **THE AC-17 BENCHMARK RECORD
  (SESSION-052)**: performance rows close with same-input +
  same-output proofs and same-box medians on both engines; **THE
  HONEST-ZIP RECORD (SESSION-051)**; **THE
  STALE-COMPARISON-TARGET LAW (SESSION-050)**; **THE PUBLIC-REPO
  RECORD (SESSION-050)**; **THE TAG-STATE BUILD LAW (SESSION-049,
  re-proven on the runner at 050)**; **THE -X-UNDER-TRIMPATH LAW
  (SESSION-048)**; **THE VERIFY-THE-VERIFIER LESSON
  (SESSION-048/049/050/052/059/060/061/062/063 — the renderer eats
  characters on displays AND the tool echoes; settle every suspect
  reading with byte-level counts)**; **THE DEFAULT-ICON FIND
  (SESSION-048)**; **THE FLAG-CLEAR LAW (SESSION-047)**; **THE
  REFERENCE'S-OWN-ARITHMETIC LAW (SESSION-046/048)**; **THE VLM
  MISREAD LEDGER (FOUR incidents: 042/043/044/045; 046–063 ran
  clean)**; **THE BINDING-IMPORT LAW (SESSION-044)**; **THE
  RACE-PASS DIVIDEND (SESSION-045)**; **THE EXTRACT-AND-EXEC
  SCOPE LAW (SESSION-041)**; **THE CROSS-BINARY DISAGREEMENT
  LAW**; **THE SENTINEL-SEAT RACE LAW (SESSION-040)**; **THE
  RETROACTIVE FIELD-TAG AUDIT (SESSION-040)**; **THE NIL-OPTIONAL
  DEREF (SESSION-040)**; **THE ONE-OWNER-PER-SURFACE law
  (SESSION-040)**; **THE FIELD-TAG LAW (SESSION-038)**; **THE
  DISPLAY-PIPELINE CORRUPTION LESSON (SESSION-038 — the Edit
  tool's SPACE-vs-TAB; gofmt -l after every edit batch;
  python-patch Go source when the tool rendering is suspect)**;
  **THE SYMLINK-BLOCK FALLBACK (TOOLCHAIN.md Am.4)**; **THE
  RENDER-CONTEXT LAW (SESSION-037)**; **THE VLM RATE-LIMIT WINDOW
  (SESSION-039)**; **THE PINNED-VERDICT-ROW LESSON (SESSION-039)**;
  **THE ONE-SLICE-PROPERLY RECORD (036–063)**; and the standing set
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
    every commit — sha256 hashes are fine, tokens are not); **THE
    BACKGROUND-PROCESS lesson** (long suites + downloads in FOREGROUND
    CHUNKS); **PUSH AFTER EVERY TASK** (ask for the PAT the moment a
    push is due; never store it).

## 2. Next session: SESSION-064 — the Phase-7 finish, owner-input driven (the owner's standing delegation recorded)

The implementation side of Phase 7 is DONE and **the debt books are
EMPTY** (the deferred-debts polish completed through SESSION-062; the
last named test-adjacent debt — the 1-INSTALL bootstrapper leg —
landed at SESSION-063 with the t-70 GUID corruption corrected and
teeth on both pins). **The owner's standing input (SESSION-058,
verbatim): "Continue your work please. do based on our documentation.
you choose. I dont have urge and haste to start the app. you can keep
working. only tell me if you really need me to test the app."** — a
delegation of the pick, no haste on T-82 or the release, and nothing
currently needs the owner's testing. SESSION-064 follows the decision
tree (in priority order):

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
- **(c) Otherwise (the standing delegation)**: with the debt books
  EMPTY there is no named slice left — **the honest move is to
  SURFACE THE PHASE-8 MENU** (doc 12's future-work register — the
  owner's pick when they want it; the menu surfacing itself is the
  session's one slice: present the register's options plainly in the
  chat + the report, and close cleanly at a natural boundary when
  capacity runs out). If the owner's T-82/release inputs arrive
  first, they take priority.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-063.md` — the most recent report (the
   bootstrapper leg landed with proof; the t-70 GUID corruption
   corrected; the debt books empty; the battery 2397 + 95 unchanged).
2. `docs/PARITY-REPORT.md` — the owner-reviewable state (§9's
   not-done list is now only the owner's LIVE legs).
3. `docs/poc-artifacts/t-82-clean-vm-checklist.md` — the owner's
   fill-in surface (the intake target if results arrive; §4's dated
   note on the new leg's honest-zip boundary).
4. `docs/poc-artifacts/t-81-release-workflow.md` §5 — the proven
   release path (the recipe if the owner says ship it).
5. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + the Phase-8 future-work register (the tail).
6. `README.md` (the front page + the SESSION-063 status entry).

## 4. Known constraints / owner actions for SESSION-064

- **OWNER ACTION #1 (no haste — the owner's own words): T-82's live
  execution** — the owner's Windows machine runs the nine checklist
  items against the NEXT release's two-exe zip (item 1 now carries
  the full installer check: git + the WebView2 probe + the exe's
  first-launch one-click offer; the current live zip carries the old
  git-only checker — §4's dated note in the checklist doc) and fills
  §2; the signed checklist closes T-82. Nothing blocks on it.
- **OWNER ACTION #2: the milestone release decision** — say the word
  and the session cuts it (see §2b).
- **OWNER ACTION #3: the Phase-8 pick** — only if the owner wants it
  (doc 12's register); otherwise the standing delegation surfaces
  the menu and closes cleanly.
- **The repo is PUBLIC** (SESSION-050) — never commit secrets
  (public-facing guarantee); pushes still need the ephemeral PAT.
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (last rebuilt at SESSION-047 — the
  fifteenth; the SIXTEENTH rebuild is pending) — rebuild it FIRST
  only if the session runs a LINUX GUI build (the windows legs are
  CGO-free; the full battery runs without it — re-proven at
  SESSION-063).
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
- **Deferred test debt on the books**: NONE (the §9 inventory was
  emptied at SESSION-062; the last named test-adjacent debt — the
  bootstrapper leg — landed at SESSION-063). Only the owner's LIVE
  legs remain (T-82 item 1 for the bootstrapper leg; the launchers'
  clean-VM smoke rides T-82 the same way).
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

## 5. THE PROMPT TO COPY (SESSION-064)

```
You are starting SESSION-064 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 7's IMPLEMENTATION SIDE IS DONE, WITH THE DEBT BOOKS
EMPTY: T-80 COMPLETE; T-81 COMPLETE (the dry-run green on a real runner + the
cross-machine byte-identity proof); T-83 COMPLETE (the docs); T-82 PARTIAL
(items 1–8 are the owner's live Windows legs — the owner has explicitly NO
HASTE: their SESSION-058 words were "I dont have urge and haste to start the
app ... only tell me if you really need me to test the app"); AC-17 CLOSED
(SESSION-052). The deferred-debts polish completed across SESSION-058–062
(the chrome-retry LIVE leg; the master scans' delivered law; the harvest's
probe-panic-ladder cousin; the note_state surgery's CRLF text-mode
differential; the GUI-side hero/wiring equivalents — the §9 test-debt
inventory emptied). SESSION-063 (path (c) a sixth time) landed the LAST named
test-adjacent debt: the 1-INSTALL BOOTSTRAPPER LEG — 1-INSTALL.bat's WebView2
Evergreen Runtime check (a READ-ONLY registry probe of the Evergreen runtime's
own key, per-user then per-machine, value pv), the found/absent cards, the
conditional final card, pinned BOTH sides (TestInstallBatChecksGitAndWebview2
— a strict sharpening; reg query is the ONE allowed reg verb — and the golden
differential's 1-INSTALL goOnly contract), the mutation teeth proven on the
corrupted-GUID shape (the t-70 poc plan's recorded GUID was a CORRUPTION since
SESSION-033 — settled against learn.microsoft.com, corrected with a dated
note), and the exe's first-launch strategy settled at source level (wails
v2.16.0's default download: one confirm dialog + the Evergreen fetch). The
battery: 2397 main + 95 webui = 2492 (unchanged — the slice extended one test
+ the differential contract, not the rows). THE RECORDED DEBT BOOKS ARE NOW
EMPTY — what remains on the leg is the LIVE half only (the owner's T-82 item
1; the new checker rides the NEXT release's zip — the current live zip
carries the old git-only one). … go-ci: GREEN through d2fa36c (run
39833368088); SESSION-063's close push is the latest — verify its run green
at the open (a red is a REAL failure now). What remains in Phase 7 is the
OWNER's: T-82's live checklist (no haste) + the milestone release decision.

Then follow docs/SESSION-PROTOCOL.md exactly: bootstrap (clone
https://github.com/assadigit/second-brain-project — the work repo, public;
read README.md, docs/SESSION-PROTOCOL.md, docs/NEXT-SESSION.md, the two most
recent session reports, git log --oneline -15), orient (the state card §2's
decision tree: (a) T-82 intake if the results arrive; (b) the release session
if "ship it" + the number arrive; (c) otherwise — with the debt books EMPTY,
the honest move is to SURFACE THE PHASE-8 MENU [doc 12's register — the
owner's pick] and close cleanly at a natural boundary), work (the standing
laws: reality wins, tests never weakened, no secrets ever, stay inside the
phase, push after every task), close (the honest gate evaluation, the commit
+ push, the SESSION-064 report, the NEXT-SESSION.md rewrite, the worklog
append), and report in chat with the literal next prompt.
```
