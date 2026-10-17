# SESSION-024 — 2026-10-17 — Phase 4 continues: T-49 directorysync landed at e2c941a — 1736 tests race-clean (+78) — THE DIRECTORY SYNC COMPLETE (the parser, the config trio, build_index + THE MIRROR LAW, the publish leg + the four truths, the command cycle with trash-not-delete, the CLI, THE DRIFT SEAL'S PUBLISH DOOR WIRED); the byte-exact publish body + the deterministic mirror sha in the 22-section golden; the eighteenth full wipe repaired before work; T-50 deliberately deferred (the one-task law)

## Goal (from NEXT-SESSION.md at start)

SESSION-024 = Phase 4 continues: **T-49 `internal/directorysync`**
(L) per spec-cloud-directory-sync.md — the REST of the module:
build_index, publish, the command cycle, the config trio; the drift
seal's publish door (vaultseal.DriftSealDeps.PublishAfterBatch)
wires here. Acceptance per the task list: the index golden; the
command-cycle tests; the test_directorysync ports. T-50 mirror (M)
if capacity remained; close cleanly at a natural boundary otherwise.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was WIPED (the eighteenth full wipe)**: neither
  repo, no Go, no venv, no symlinks. Rebuilt per NEXT-SESSION v24 §1
  before any work. Git evidence over assumptions: the credentials
  block's first repo ("Currena") turned out to be an UNRELATED
  sibling project (a deployed financial-insights site) — the work
  repo was located via the GitHub API's authenticated repo listing
  (`assadigit/second-brain-project`, private) and cloned with the
  PAT-passed ephemeral Basic header, **scrubbed from the config
  immediately** (verified by direct config read — 0 extraheader, 0
  token matches). GitCurator checked out at `98f4fc7` (matches
  PINNED-GITCURATOR-REV byte-for-byte), go1.27.2 installed at
  /home/z/.local/go (go.dev tarball), the venv rebuilt
  (/home/z/venvs/gitcurator, Python 3.12.15, stdlib-only), the
  symlinks re-created. go.mod verified UNCHANGED
  (`git rev-parse HEAD:go.mod` — the v1.36.1/tls-client/1.24.1 pins
  intact; a go.sum churn from `go mod download all` was REVERTED —
  zero dependency movement). The baseline re-verified BEFORE work:
  1658/1658 race-clean over 29 packages (the -list count matched the
  card exactly).

- **T-49, commit `e2c941a` (30 files, +~11,700)** — the module
  complete:
  - *noteparse.go*: parse_note (the drawer's grammar: the
    frontmatter scan, the banner callout strip, the footer strip,
    the stray-artifact cleanup, the H1/TL;DR ladder with the
    search-fallback, the sections walk, the sign-up chip with the
    hand-rolled Unicode \b check, the pricing enum, the banish-tag
    filter, the — → empty summary swap AFTER the structured check) +
    ParsedNote.ToMap() (the reference's exact dict shape in the
    reference's key order).
  - *systemfiles.go*: EnsureSystemFiles (create-once,
    never-overwrite; the embedded shipped synonyms at
    assets/directorysync/search-synonyms.json — copied verbatim by
    the generator, sha256-pinned) + LoadSynonyms.
  - *index.go*: BuildIndex — THE VAULT MIRROR LAW (the tree IS the
    folder tree; config ghosts drop; undeclared folders publish
    prettified; the case-insensitive folder-name translation over
    UNIQUE sanitized matches only; the FOLDED private belt before
    the on-disk pair joins; the banished-host skip; the banish wall
    with the shared-host exceptions; the popular tags; the (cat,
    sub) tuple pairs that kill the phantom category; the
    insertion-ordered index tree) + slugify (NFKD + the Persian
    block) + iterNotes (the os.walk files-before-subdirs law) +
    readText (decode → newlines → 512k RUNES) +
    publicCategoryRecords + banishedHostWall.
  - *publish.go*: PublishIndex (POST, Bearer, 1010 UA, the
    json.dumps(ensure_ascii=False) wire bytes through the NEW
    storage.PyDumpsUTF8) + FetchPublishedIndex (the four truths —
    ok/never_published/not_configured/error) + effectiveConfig (the
    env override matrix) + RunSync + PublishAfterBatch +
    pyReprValue (the dict-repr log lines).
  - *commands.go*: FetchPendingCommands + AckCommand (the
    urlquoted-id door, PyDumpsCompact ack bodies) + the present-aware
    dictDoor ('key in payload' — the belt floor's reader could not
    answer absent-vs-nil) + the vault appliers (editNoteFields with
    the whole-body edit + the verbatim insert(1) quirk, moveNoteFile
    with the collision suffix, trashNote → .trash/ with the frozen
    local stamp, banishAdd's upsert over the RAW tree [the owner's
    extra keys ride verbatim], banishRemove) + ApplyAdminCommands
    (idempotent, failure-isolated; the unknown type is a SILENT
    failure — no log, no ValueError prefix) + PushMirrorRepo (the
    plain runner: env inherited, never redacted — the reference's own
    module has no _git_env; HideWindow per ADR-0009's Windows quirk
    list) + RunCommandCycle (apply → mirror → publish → ack; the
    down-Worker never looks like a drained queue).
  - *cli.go*: Main (--vault/--config/--sync-only/--publish/
    --print-index/--out/--make-admin-hash) + the PBKDF2 helper
    hand-rolled over stdlib crypto/hmac (RFC 2898; the x/crypto door
    stays shut) + the getpass fallback convergence (plain stdin).
  - *platform twins* (hideWindow), the assets embed
    (assets.SearchSynonyms), **the drift seal's publish door WIRED**
    (vaultseal.DefaultDriftSealDeps binds directorysync.
    PublishAfterBatch — the not-wired warning retires; the log
    adapter tags plain lines "info").
  - *storage*: PyDumpsUTF8 (json.dumps ensure_ascii=False — the
    publish wire shape) + the latent \b/\f escape fidelity fix in
    pyEncodeString (the full suite proves no golden depended on the
    old \u0008/\u000c forms).
  - *THE GOLDEN* (22 sections, 87,503 bytes,
    testdata/directorysync.golden.json): GENERATED by the new
    testdiff/goldens/dump_directorysync.py from the unmodified
    reference @ 98f4fc7 — **the generator imports the reference's own
    test fixtures via importlib (zero transcription risk)**; run
    twice, byte-identical. Sections: the parser matrix (the
    reference suite's seven note shapes + loose/banish/featured),
    the pure helpers (slugify/section_key/prettify/norm_host/
    is_banished incl. the lstrip character-set quirk/banished_wall),
    the trio's creation law + the shipped bytes + the synonyms
    sha256, build_index on the all-laws fixture (10 sites, 5
    categories, the banished skip, the private fold, the ghost drop,
    the case-drift translation, the straggler, the slug -2 laws),
    **the publish body's EXACT wire bytes (10,514 — key order +
    separators + raw Farsi, byte-compared)**, the walk's id map, the
    appliers' exact file surgery (12 cases with frozen trash/banish
    stamps), the frontmatter_set matrix, the effective_config
    override matrix (7 env cases), the CLI shapes, and **the mirror
    push's deterministic commit (the local bare remote + frozen
    clock + frozen git dates → the reference's exact sha
    c28cc921a561…)**.
  - *The batteries* (78 new top-level tests): the ported selections
    of the reference's whole test_directorysync.py (every test
    function of TestParser/TestBanish/TestSystemFiles/
    TestBuildIndex/TestPublish/TestCli/TestCommandAppliers/
    TestBanishWall/TestCommandCycle) + the four-truths read-back
    battery + the non-int-site-count honest error + the loopback
    stub worker (the _StubWorker/_StubWorker2 ports with the ORDERED
    event log — the publish-before-ack proof) + the local-bare
    mirror-push battery (the full ladder, the clean-tree quiet true,
    the no-remote skip) + the wiring proofs (the default door bound;
    the manual fast-sync path runs the real cycle against the
    loopback stub) + the golden differential.
  - *Spec + records amended in the same commit*: the trio-filename
    reality-wins fix (banished.json/search-synonyms.json — the spec
    had stale .md names), the SESSION-024 shipped record + the
    Go-note amendments; D13 stamped; README's status paragraph; the
    package doc.go records the T-49 completion.

- **Pushed**: `c0a8ca4..e2c941a main -> main` (the ephemeral PAT via
  the transient `-c http.extraheader` Basic form; the config verified
  CLEAN after — 0 extraheader; HEAD == origin/main == `e2c941a`).

- **T-50 — NOT STARTED, deliberately** (the one-task law): the
  session's capacity went to the eighteenth-wipe rebuild + the
  largest single-module port so far (1,826 LOC of reference, 78
  tests, the golden + its generator, the spec amendments). T-50 is a
  fresh (M)-sized port (mirror.py 701 LOC + the linking/embeddings/
  recall scope per spec-core-mirror.md + spec-core-links.md) — a
  rushed tail would violate "one task done properly with proof beats
  a rushed whole." NOTE for the card: the task list names
  "spec-core-linking" but the file on disk is spec-core-links.md.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-49's acceptance — PASS**: *the index golden* ✓ —
  testdata/directorysync.golden.json, GENERATED from the unmodified
  reference @ 98f4fc7 (the generator verifies the pinned rev before
  running; the fixture bytes come from the reference's own test
  module via importlib), run twice byte-identical; the golden pins
  the full index tree, the exact publish-body wire bytes (10,514),
  and the deterministic mirror-commit sha. *the command-cycle tests*
  ✓ — the ported TestCommandCycle selections (the full-cycle order
  proof with the event log: pull → sync BEFORE ack; the
  down-Worker-commands-wait law; the wrong-token pull; the
  dead-worker ack; the plain-folder mirror no-op) + the local-bare
  mirror-push battery. *the test_directorysync ports* ✓ — every test
  function of the reference's suite ported (TestParser 10,
  TestBanish 4, TestSystemFiles 2, TestBuildIndex 14, TestPublish 5,
  TestCli 2, TestCommandAppliers 16, TestBanishWall 3,
  TestCommandCycle 5) + the honest additions (the four truths, the
  non-int count, the non-ASCII raw-UTF8 wire leg, the drift-door
  wiring proofs). Honest notes: (1) the reference suite has NO
  fetch_published_index tests — the four-truths battery is a
  spec-driven Go-side addition (the read-back door's states asserted
  against the loopback stub + the honest dead-port error); (2) the
  held-window invisibility is a Worker-side concept (the client
  never sees held commands) — nothing to port, noted in the spec;
  (3) the reference suite never exercises push_mirror_repo against a
  real git remote — the local-bare battery covers that leg with the
  T-48 technique (the golden pins the deterministic sha). *TEMP
  VAULTS/TREES ONLY, NO real network* ✓ — every git operation is
  local (t.TempDir trees, local bare repos); the Worker surface is
  the loopback stub (127.0.0.1 ephemeral ports); a secrets sweep
  over the committed files found nothing (the pre-existing fake
  fixtures untouched).
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -count=1
  -race -p 2 ./...` → **29 packages ok, 1736/1736 top-level tests
  PASS, 0 failures, 0 skips, 0 data races** (was 1658 at session
  start; +78 — the arithmetic closes exactly; the package count
  stays 29 — directorysync already counted as the T-47 floor);
  go.mod UNCHANGED (`git rev-parse HEAD:go.mod` identical — zero new
  dependencies; the accidental go.sum churn from `go mod download
  all` was reverted and the suite re-verified green WITHOUT it).
- **The golden is an honest differential**: generated from the
  unmodified reference; the two-run byte-identical proof; it caught
  TWO live port bugs before they could land (the nil-slice
  null-vs-[] wire divergence in the subcategories; the silent
  unknown-type failure's log/prefix shape) — the golden doing
  exactly its job.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-cloud-directory-sync.md` —
  (1) the reality-wins trio-filename fix (banished.json /
  search-synonyms.json — the code's constants over the spec's stale
  .md names, same commit); (2) the SESSION-024 shipped record; (3)
  the Go-note amendments (PyDumpsUTF8 + the \b/\f fix, the
  []-vs-null byte law, the dictDoor presence checks + the
  double-wrapping trap, the silent unknown-type failure, the
  set-iteration convergence, the push_mirror env-inheritance law +
  the upstream-tracking first-push note, the CPython re vs Go
  regexp \s/\b convergence, the lstrip character-set quirk fix, the
  getpass fallback door, the raw-tree banished round-trip, the
  drift-seal log adapter).
- `docs/port-blueprint/traceability-matrix.md` — the D13 row stamped
  (T-49 ✓).
- `README.md` — the status paragraph (T-49 complete, 1736
  race-clean, next: T-50 or the Phase-4 close-out).
- `internal/directorysync/doc.go` — the T-49 completion record.

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — thirteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2
  has the record + the rollback path).
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- The drift seal's publish door is now WIRED; the deferred GUI
  halves (the QThread wrapper, the drift seal's GUI wiring) stay on
  the GUI worker shell task's books as before.

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session
   (it worked exactly as designed this session: the
   eighteenth-wipe rebuild + the T-49 push both rode the ephemeral
   form; the config verified clean after every use).

## Handoff

- Next session: **SESSION-025 = Phase 4 continues — T-50
  `internal/mirror` (+ linking/embeddings/recall)** (M) per
  spec-core-mirror.md + spec-core-links.md (NOTE: the task list
  says "spec-core-linking" — the file on disk is spec-core-links.md)
  — block-surgery idempotence, containment guards, the suggestion
  lifecycle; then the Phase-4 close-out evaluation (T-40..T-49 all
  landed; the phase's exit is the honest full-suite + spec sweep).
- NEXT-SESSION.md rewritten: yes (v25).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
