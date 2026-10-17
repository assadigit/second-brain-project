# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker
  HTTP API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity (constants, links, storage+dryrun,
  prompts, taxonomy, notebuilder, config, helpers).
- **Phase 2 CLOSED (SESSION-006–010; all six tasks)**: T-20 `internal/
  store`, T-21 `internal/notestate`, T-22 `internal/notereads`, T-23
  `internal/websitepipeline/state.go`, T-24 `internal/store/repos.go`,
  T-25 `internal/batch` — 425 tests race-clean at the close.
- **Phase 3 CLOSED (SESSION-011–014; all five tasks)**: T-30 `internal/
  llmclient` (`f2073b5`), T-31 `internal/webfetch` (`4d93877`), T-32
  `internal/webextract` (`833b21b`), T-33 `internal/connectioncheck`
  (`e5846bd`), T-34 `internal/workerclient` (`6ffc4e9`) — 872 tests
  race-clean at the close.
- **Phase 4 OPEN (SESSION-015–024)**: T-40 `internal/websitepipeline`
  at `8098ac4`; T-41 `internal/websitepipeline/directorynote.go` at
  `f528c22`; T-42 `internal/githubpipeline` + `internal/batch/
  summary.go` at `e92e4e7`; T-43 `internal/handdelivery` at `be688bd`;
  T-44 `internal/chrometabs` at `71e5e34` (THE FIFTH DOOR COMPLETE);
  T-45 `internal/vaultscan` at `5c0508e` (THE VAULT SCAN COMPLETE +
  the repos-banishment floor); T-46 `internal/vaulthygiene` + the
  COMPLETION of `internal/repobanish` at `992a3d1` (THE STRAY-NOTES
  DOOR + THE REPOS BANISHMENT COMPLETE); T-47 `internal/vaultseal`
  at `d3a5d27` (SESSION-022 — THE SEAL COMPLETE, + the
  directorysync belt floor); T-48 `internal/goodrepos` +
  `internal/reporadar` at `7179a6c` (SESSION-023 — THE PUBLISHER +
  THE RADAR COMPLETE); **T-49 `internal/directorysync` landed at
  `e2c941a` (SESSION-024) — THE DIRECTORY SYNC COMPLETE**: the whole
  of cloud/directory_sync.py (the note-section parser, the config
  trio with the embedded shipped synonyms, build_index + THE VAULT
  MIRROR LAW, the publish leg with the byte-exact
  json.dumps(ensure_ascii=False) wire body, the read-back door's
  four truths, the admin command cycle with trash-not-delete, the
  CLI with the hand-rolled PBKDF2, THE DRIFT SEAL'S PUBLISH DOOR
  WIRED — vaultseal.DefaultDriftSealDeps binds directorysync.
  PublishAfterBatch, the not-wired warning retired) — with the
  ported test_directorysync selections (the reference's whole
  suite) + the four-truths battery + the loopback stub with the
  publish-before-ack order proof + the local-bare mirror-push
  battery + THE GOLDEN (22 sections — the byte-exact publish body
  10,514 + the deterministic mirror commit sha c28cc921;
  dump_directorysync.py imports the reference's own fixtures via
  importlib). **1736/1736 tests race-clean over 29 packages; T-49's
  acceptance — PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-024 pushed T-49 the moment it went green. The
  recovery order for ANY mid-task death stays: (1) `git status` +
  `git log --oneline -5` + file timestamps to learn what ACTUALLY
  survived — never trust a handoff summary's claim over `git log`;
  (2) a preserved commit or draft is continued and audited (the full
  battery re-run), never blindly trusted, never restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **eighteen full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth, SESSION-024's
  eighteenth), plus SESSION-010's mid-session mode-flip, plus
  SESSION-021's CONTEXT death with the sandbox intact. The restore
  recipes: start-of-session (clone `assadigit/second-brain-project`
  with the PAT passed ephemerally as the Basic `x-access-token:<PAT>`
  header — **SESSION-024's find: the credentials block's first-listed
  repo (Currena) is an UNRELATED sibling project — locate the work
  repo via the GitHub API's authenticated /user/repos listing if the
  clone looks wrong**; never store the PAT — AND SCRUB IT: `git
  clone -c http.extraheader=…` PERSISTS the PAT into the new repo's
  .git/config, so run `git config --unset-all http.extraheader` in
  BOTH clones right after cloning, then push with the ephemeral
  `-c` form only); GitCurator @ `98f4fc7` (verify against
  `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/.local/go`
  (go.dev tarball), venv `/home/z/venvs/gitcurator` (Python 3.12,
  stdlib-only), symlinks `/home/z/second-brain-project` and
  `/home/z/GitCurator`; mid-session mode-flip (`git checkout-index
  -f -a` restores modes from the index; `git fsck` to sanity-check).
  Verify the pins at every start: `git rev-parse HEAD:go.mod` —
  expect the v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do
  NOT bump tls-client past v1.16.0). **Do NOT run `go mod download
  all`** (SESSION-024: it churns go.sum with extra verification
  hashes — revert the churn; `go build`/`go test` work without it).
  **The renderer-escape + cd-discipline laws stay: the two-character
  pairs [h and [m are eaten on the OUTPUT path, the INPUT path, AND
  the BASH-ARG path — write commit messages to a FILE and commit
  with -F, then verify the committed bytes with 'git log
  --format=%B'; note the DISPLAY of test diffs eats the pairs too
  (SESSION-024 chased a phantom "corrupted golden" that was only
  the tool display — grep the FILE bytes before believing a
  corruption claim). AND the cd-discipline law: the shell RESETS its
  working directory between tool calls — EVERY git command carries
  its own cd; never commit outside
  /home/z/my-project/second-brain-project.**

## 2. Next session: SESSION-025 — Phase 4 continues: T-50 mirror (M), then the Phase-4 close-out evaluation

- **T-50 `internal/mirror` (+ linking/embeddings/recall)** (M) per
  **`docs/port-blueprint/module-specs/spec-core-mirror.md` +
  `spec-core-links.md`** (NOTE: the task list's "spec-core-linking"
  is named spec-core-links.md on disk) — READ THE SPECS BEFORE THE
  CODE. The reference: `app/gitcurator/core/mirror.py` (701 LOC) +
  the linking/embeddings/recall surface. Acceptance per the task
  list: block-surgery idempotence; containment guards; the
  suggestion lifecycle. `internal/mirror/` currently holds only the
  doc.go placeholder.
- After T-50: **the Phase-4 close-out evaluation** (T-40..T-49 all
  landed — evaluate the phase honestly against the roadmap's Phase-4
  exit: the full-suite numbers + the spec sweep; surface anything
  unfinished with evidence rather than improvising).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; the dry-run seam at call sites; rows collected before
  nested statements; tests never weakened; reality wins over specs
  (fix the spec in the same commit); goldens are GENERATED by
  running the unmodified Python app, never hand-typed; no per-run
  facts in goldens; the frozen-clock seams (`clockNow` per package —
  directorysync's is the newest). T-41's: text-mode reads translate
  universal newlines before any cap or parse; invalid UTF-8 decodes
  with the maximal-subpart U+FFFD law; vault walks are os.walk's
  files-before-subdirs order. T-42's: the house JSON decoder's
  shapes; floats render through helpers.PyFloatRepr. T-43's:
  `pyBytesBlank` is the blank-body law; Go MAP ranges are RANDOM.
  T-44's: Go's `url.Parse` is NOT CPython's urlsplit; `cmd.Wait()`
  closes the pipe; scripted timing tests keep the reference's
  UNPATCHED cadences. T-45's: THE POINTER-THREADING LAW
  ([]*NoteRow); keep BOTH frontmatter parsers; Python text-mode
  read(n) counts CHARACTERS; the _windows_eyes law rides a stat
  seam. T-46's: the WALK-ORDER FOLD; bounded reads count runes;
  None doors type-narrow to the zero value; a revive-style table
  iteration rides sorted keys; dry-run asymmetries preserved. T-47's:
  THE REDACTION LAW; the token NEVER touches .git/config; FETCH_HEAD
  rewritten after every token-URL fetch; env scrubbing + HideWindow
  where the REFERENCE's own module scrubs; frozen GIT DATES make
  commit shas deterministic; the reference's mock-surface maps to
  instance seams. T-48's: the apiFn/tokenURLFn instance seams; the
  ReadDir-order convergence; atexit maps to an exported Cleanup;
  the base-dict freeze quirk; the frozen-dates sha trick.
  **T-49's additions: the []-vs-null byte law (a Go nil slice dumps
  as null — initialize `[]any{}` wherever Python builds `[]`; the
  byte-exact wire golden catches what tree comparison cannot see);
  `storage.PyDumpsUTF8` is json.dumps(ensure_ascii=False) — the
  publish wire shape; dictDoor (present-aware) for `'key' in
  payload` doors — and NEVER double-wrap a door (pass the RAW value
  into the wrapping function); the reference's unknown-command-type
  failure is SILENT (no log, bare error) while except-path failures
  log str(e) and carry the f'{type}: {e}' prefix; the on_disk_cats
  SET iteration converges to walk order (the stable sort settles
  it); push_mirror_repo's runner inherits env and never redacts
  (its module has no _git_env — reality wins) and the first-ever
  plain `git push` needs upstream tracking; CPython re \s/\b are
  Unicode-aware vs Go regexp ASCII (hand-roll the boundary check
  where it matters); lstrip('www.') is a CHARACTER-SET strip
  (TrimLeft, not TrimPrefix); banished.json round-trips ride the
  RAW tree so the owner's extra keys survive; the generator can
  IMPORT the reference's own test fixtures (importlib) — zero
  transcription risk.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-024.md` — the most recent report (the
   T-49 evidence, the 22-section golden + the 78 tests, the
   eighteenth-wipe rebuild, the Currena/unrelated-repo find, the
   go.sum-churn note).
2. `docs/port-blueprint/module-specs/spec-core-mirror.md` +
   `docs/port-blueprint/module-specs/spec-core-links.md` — BEFORE
   T-50's code (the binding specs; note the task list's name
   discrepancy).
3. The reference: `app/gitcurator/core/mirror.py` (701 LOC — all of
   it) + the linking/embeddings/recall modules the specs name + the
   relevant test suites (read the selections before porting).
4. Working examples: `internal/directorysync` (T-49 — the newest
   full-module port: the golden generator that imports the
   reference's fixtures, the loopback stub, the local-bare mirror
   push, the byte-exact wire body), `internal/goodrepos` (T-48 —
   the apiFn/tokenURLFn seams), and
   `testdiff/goldens/dump_directorysync.py` (the newest generator).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-025

- **The go-directive acknowledgement (STILL WAITING — thirteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2
  has the record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers
  x/net/html, approve an ADR and a future session swaps
  `internal/webextract/parser.go`'s scan internals).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB the persisted
  extraheader from .git/config after the clone — see §1); goldens are
  GENERATED by running its unmodified code (`PYTHONDONTWRITEBYTECODE=1`
  keeps it pristine).
- **Dependency law**: `modernc.org/sqlite` pinned at v1.36.1;
  `bogdanfinn/tls-client` pinned at v1.16.0; `golang.org/x/net`
  v0.48.0 + `golang.org/x/text` v0.32.0 (direct); ANY other add needs
  a new ADR — surface, don't improvise. (T-49 added zero;
  PyDumpsUTF8 and the PBKDF2 are stdlib.)
- **T-50-specific safety laws**: temp vaults/trees only — never the
  owner's real data; no real network in tests; the redaction +
  token laws carry where any remote talk exists; tests never
  weakened.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the GUI worker shell
  task (the mode wiring, the Ollama warmup, the inputs fetchers, the
  interactive doors' GUI halves — incl. the drift seal's QThread
  wrapper [its publish wiring is DONE — T-49 wired the engine half],
  test_fifthdoor §9b's end-of-delivery modal legs, test_vaultscan's
  GUI legs, the two-vault worker job, the GUI scorecard models, the
  Python-source contract tests of test_reporadar §6); the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs (Phase 7 clean-machine checklist); a Windows
  VM without WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-025)

```
You are starting SESSION-025 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (eighteen full wipes across
   SESSION-005..024 plus mid-session degradations; SESSION-021 added the
   CONTEXT-death class — the working context ended AFTER the T-46 commit
   existed but BEFORE the push+close; the continuation recovered by trusting
   git evidence over the lossy handoff summary, auditing the preserved commit
   with the full battery, then pushing). FIRST verify with evidence: work repo
   at /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-
   project), read-only reference /home/z/my-project/GitCurator @ 98f4fc7
   (symlinked /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV),
   toolchain go1.27.2 → /home/z/.local/go, venv /home/z/venvs/gitcurator
   (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, and go.mod with
   modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the go
   directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not
   bump any of these). Then check `git status` and `git log --oneline -3`:
   SESSION-024 landed T-49 directorysync at e2c941a (the tree must be clean
   and origin/main at the SESSION-024 close; the baseline is 1736 tests
   race-clean over 29 packages — re-verify the suite before work if anything
   looks off). If anything is missing, rebuild per docs/NEXT-SESSION.md §1
   (clone via the PAT I give you when you ask — pass it ephemerally with git
   -c http.extraheader="Authorization: Basic $(printf 'x-access-token:%s'
   "$PAT" | base64)", the Basic scheme; never store it — AND SCRUB IT: `git
   clone -c http.extraheader=…` PERSISTS the header into the new repo's
   .git/config, so run `git config --unset-all http.extraheader` in BOTH
   clones right after cloning; a mid-session mode-flip is repaired with `git
   checkout-index -f -a` after `git fsck`; and if the sandbox or the context
   dies MID-TASK, learn what actually survived from git evidence — `git
   status` + `git log --oneline -5` + file timestamps — before restarting
   from zero: a preserved commit or draft is continued and audited, never
   blindly trusted, and a handoff summary's claim never outranks `git
   log`'s). SESSION-024's operational finds: the credentials block's
   FIRST-listed repo (Currena) is an UNRELATED sibling project — locate the
   work repo (assadigit/second-brain-project, private) via the GitHub API's
   authenticated /user/repos listing when in doubt; do NOT run `go mod
   download all` (it churns go.sum — plain build/test works without it);
   the renderer eats the escape pairs on the DISPLAY of test diffs too (a
   phantom "corrupted golden" was chased — grep the FILE bytes before
   believing a corruption claim). The older laws stand: (a) THE
   CD-DISCIPLINE LAW — the shell resets its working directory between tool
   calls, so every git/go command carries its own cd (or go's -C flag);
   never commit outside /home/z/my-project/second-brain-project; (b) the
   renderer-eats-escape-pairs lesson — the two-character pairs [h and [m
   are eaten on the OUTPUT path, the INPUT path, AND THE BASH-ARG path:
   avoid those spellings, or write commit messages to a FILE and commit
   with -F, then verify the committed bytes with 'git log --format=%B'
   before pushing. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4
   state with T-40 at 8098ac4, T-41 at f528c22, T-42 at e92e4e7, T-43 at
   be688bd, T-44 at 71e5e34, T-45 at 5c0508e, T-46 at 992a3d1, T-47 at
   d3a5d27, T-48 at 7179a6c, and T-49 at e2c941a, 1736 tests race-clean
   over 29 packages, the environment recipes, the clone-config PAT-scrub
   hazard, the renderer + cd-discipline lessons, and the standing port laws
   incl. T-49's additions — the []-vs-null byte law, PyDumpsUTF8, the
   dictDoor presence checks + the double-wrapping trap, the silent
   unknown-type failure, the set-iteration convergence, the push_mirror
   env-inheritance law, the CPython re vs Go regexp convergence, the
   lstrip character-set quirk, the raw-tree banished round-trip, the
   importlib fixture-import generator trick) → docs/sessions/SESSION-024.md
   (the most recent report — T-49's 22-section golden + the 78 tests, the
   eighteenth-wipe rebuild, the operational finds).

3. THIS SESSION IS T-50 (M): Phase 4 continues — internal/mirror (+ the
   linking/embeddings/recall surface) per docs/port-blueprint/module-specs/
   spec-core-mirror.md AND spec-core-links.md (NOTE: the task list calls it
   "spec-core-linking" — the file on disk is spec-core-links.md) — READ THE
   SPECS BEFORE THE CODE. Acceptance per the task list: block-surgery
   idempotence; containment guards; the suggestion lifecycle. The reference
   is app/gitcurator/core/mirror.py (701 LOC) plus the linking/embeddings/
   recall modules the specs name — read the relevant reference test suites
   before porting. internal/mirror currently holds only the doc.go
   placeholder. The working precedents: T-49's internal/directorysync (the
   newest full-module port — the generator that imports the reference's
   own fixtures, the loopback stub, the local-bare mirror push, the
   byte-exact wire golden), T-48's internal/goodrepos, T-47's
   internal/vaultseal. TEMP VAULTS/TREES ONLY, NO real network. After T-50
   goes green + pushes: the Phase-4 close-out evaluation (T-40..T-49 all
   landed — evaluate the phase honestly against the roadmap's Phase-4
   exit; surface anything unfinished with evidence) — but close cleanly at
   a natural boundary if capacity runs out; one task done properly with
   proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
   Phase-4 close-out evaluation if T-50 lands, else T-50's own acceptance
   honestly), commit + push everything (ask me for the GitHub PAT if it is
   not already available — never store it anywhere), write
   docs/sessions/SESSION-025.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus
   the prompt I should copy to start SESSION-026.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
