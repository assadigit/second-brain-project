# SESSION-020 — 2026-10-17 — Phase 4 continues: T-45 vaultscan landed at 5c0508e — 1347 tests race-clean (+215) — THE VAULT SCAN COMPLETE (the radar, ONE NOTE PER SITE, the regex brain, THE WEIGHED EVIDENCE, the chunked LLM filing, the validation door with the landable move, the picking door + the lock law, the census stamp, the apply pass; the golden + the ported selections of all five core suites; the repos banishment floor landed with it); the FIFTEENTH environment degradation repaired before work (a full wipe — both repos, the toolchain, the venv, the symlinks)

## Goal (from NEXT-SESSION.md at start)

SESSION-020 = Phase 4 continues: **T-45 `internal/vaultscan`** (L)
per spec-core-vault-scan.md — the vault scan: the radar, the
weighing, the validation, the apply pass (core/vault_scan.py:548-
3353). Acceptance: the ported vault-scan suites (the reference's
own suite; the house generator pattern if a golden was warranted).
T-46 vaulthygiene + repobanish if capacity remained — with the
standing instruction to close cleanly at a natural boundary.

## What was done (task IDs, files, commits — cite hashes)

- **The environment had degraded AGAIN (the FIFTEENTH event, a full
  wipe this time)**: BOTH repos gone (work repo + GitCurator), the
  Go toolchain gone, the venv gone, the symlinks gone — only the
  sandbox's Next.js shell survived. Repaired per the §1 recipes:
  discovered the remote via the GitHub API with the in-chat PAT
  passed ephemerally (the Basic `x-access-token:<PAT>` header form),
  cloned `assadigit/second-brain-project` (clean tree at `05e67e0` =
  the SESSION-019 close = origin/main; T-44 at `71e5e34` verified)
  and `assadigit/GitCurator` (`98f4fc71…` matching
  PINNED-GITCURATOR-REV byte-for-byte), **scrubbed the persisted
  extraheader from BOTH .git/configs right after the clones**
  (verified: no `ghp_`/`x-access-token` traces in either config),
  reinstalled go1.27.2 (the go.dev tarball → /home/z/.local/go,
  sha256 ecbadb99…), recreated the stdlib-only Python 3.12.15 venv,
  recreated the symlinks, verified the go.mod pins
  (`git rev-parse HEAD:go.mod` = 5c0d4e4a — the same bytes as the
  SESSION-019 close), and re-verified the baseline suite green
  (22 packages ok, 1132/1132 race-clean) before any work.

- **T-45, commit `5c0508e` (pushed the moment the suite went
  green)** — `internal/vaultscan` complete (8 source files + the
  golden generator + the golden + three behavior batteries) plus
  the repos-banishment floor in `internal/repobanish`:
  - *The radar* (`scan.go`): the constants, the vault-kind door,
    the truthy/lock/curated predicates, the direct frontmatter
    parser with vault_scan's OWN grammar preserved beside
    note_reads' unified parse exactly as the reference keeps both
    (flow tags REPLACE; the strip chain ends at the quote-strip),
    the CPython urlparse netloc/path shapes hand-rolled (the
    \t\r\n removal, the case-preserved netloc, the bracket law's
    raise reading as "no parse", `_splitparams`, the 'z' except
    sentinel — probed against the venv's actual CPython before
    coding), the core-note law, the pruned inventory walk (the
    house walk law) with the one-open cache door, and the move
    candidates with the lock/curated radar.
  - *ONE NOTE PER SITE* (`duplicates.go`): the consolidation
    magnet (locked notes invisible to it — never merged, never
    keeper, never even a group member), the keeper rank, the
    locked-canonicals deletion door, the tree/digest lines, the
    fold laws, the sanitizer, the disk spelling + the shared
    resolver (THE DISK SPELLING WINS, memoized).
  - *The regex brain* (`regexbrain.go`): the declared-home pass
    (websites Cat/Sub; repos the CATEGORY_FOLDERS key), the
    sibling-domain pass (the domain's settled notes agree on ONE
    home), the taxonomy pass (every matched tag agrees,
    plural-tolerant both ways) — each returning the
    declared-home contract for the validator's legal set.
  - *THE WEIGHED EVIDENCE* (`weighing.go`): the word-token
    identity, the folder census (with the repos languages), the
    weighing folder set, the scale itself (domain votes their
    share / the repos language agreement, shared tags, the
    taxonomy vocabulary, title words, the folder name, the
    declared echoes / the repos category's own folder at +0.25,
    the popularity prior capped at 0.05 and riding ONLY when
    w>0, the too-broad penalty, `round(min(w,1.0), 2)` with
    banker's rounding via helpers.RoundHalfEven, top-K by
    (-weight, folder)), the attachment with the CLOSE CALL
    verdict, and the digest's homes cell.
  - *The chunked LLM filing* (`llmpass.go`): the shared tree, the
    exact user message byte-shape (the three laws at the point of
    decision, the confidence ask with the per-kind parentheticals,
    the homes paragraph with the per-kind evidence lists),
    per-chunk failure dropping only that chunk, a panicking guest
    survived exactly like the reference's raising one (recover),
    and the honest review list (the census law — only LANDED
    chunks review).
  - *The validation door* (`validate.go`): the fresh-eyes
    absorption (folded-exact then the ≥5-char sibling
    containment) + intra-plan folded dedupe, the depth law, the
    taken-set precedence for deterministic moves, THE LANDABLE
    MOVE (the `_v` bend pre-computed at plan time and carried in
    the row as `as`; the os.SameFile self-move guard), the
    low-confidence flag (the guest's own word) + the
    weighed-evidence override check (top ≥ 0.55 and the chosen
    unlisted or ≥ 0.15 lighter → flagged).
  - *The plan* (`plan.go`) and *the apply* (`apply.go`): the
    gate's own deletions with the lock filter; consolidations
    before any filing; the retiring-never-ride law; the
    three-pass residue chain (the repos chain honestly skipping
    the domain/taxonomy passes — the poisoned well); the weighed
    evidence; the review list. Apply: the consolidations (links
    appended through AddSiteLinksToNote, files retired to
    .trash/consolidated, the slash-pair no-redundant-link law,
    ledgers re-pointed both kinds), the folders, the moves (the
    lock re-read at apply, the `as` bend honored, verified
    landings, the SPOKEN failure ledger), the census stamps
    (moved at the new home / declined in place when a selection
    spoke / reviewed), the deletions through
    ConsumeDecommissionTable + BanishMarkedNotes with
    skip_canonicals (websites) or BanishMarkedRepos with the
    picked items (repos), and the repos ledger re-point with the
    CATEGORY reverse map.
  - *The repos banishment floor* (`internal/repobanish/
    repobanish.go`): `ScanPendingRepoBanishments` (both doors, one
    list, the canonical-repo key, the dismissed-set DB gate, the
    master-table rows) + `BanishMarkedRepos` (the never-fetch
    gates — dismissed + settled + retry-resolved + row-forgotten —
    the .trash/banished move, the confirmed record rows, the
    handwritten-kept sacred law) + the note-finder. **Ported WITH
    T-45** because vaultscan's kind door consumes them and the
    test_reposcan selections exercise them end to end; T-46 keeps
    revive + the orchestration + test_repobanish (the split is
    recorded in repobanish/doc.go, the spec's shipped record, and
    the D8 row).
  - *THE GOLDEN*: `testdata/vaultscan.golden.json` — 21 sections
    (kind_doors, predicates, core_note, frontmatter, pyurl,
    fold_laws, sanitize, candidate_reason, digest_lines,
    homes_line, inventory, settled_homes, folder_profiles,
    weighing_set, weigh, attach, absorb, bend, disk_spell,
    selection, validate) GENERATED by
    `testdiff/goldens/dump_vaultscan.py` from the unmodified
    reference @ 98f4fc7 (the plain core/ generator pattern; every
    temp-vault path folded to `<vault>/…` — no per-run facts; the
    None-keys and raise-shapes pinned as the reference produces
    them).
  - *The behavior batteries*: `vaultscan_test.go` (the ported
    test_vaultscan + test_censusstamp + test_picklock core
    selections — incl. the `_windows_eyes` case-insensitive-disk
    replay through the new `fsExists`/`fsIsDir`/`fsIsFile` stat
    seam, the two-notes-loop shapes, the one-open law with the
    persistent cache), `weighhomes_test.go` (the test_weighhomes
    selections — the scale signal-by-signal, the alignment door,
    the wiring), `reposcan_test.go` (the test_reposcan selections
    — the inventory/plan/apply kind doors, ONE NOTE PER REPO, the
    repos digest + prompt, the repos brain + evidence, the round
    trip, the repos dry-run).
  - *Spec amended in the same commit*: the SESSION-020 shipped
    record + the Go-note amendments (below); the D8 traceability
    row widened (T-45 ✓); README's status paragraph; both
    package doc.go records.

- **REALITY-WINS finds the task forced** (both evidence-pinned,
  both fixed in the same commit):
  1. **THE POINTER-THREADING LAW** — the reference threads Python
     dict REFERENCES from the candidates list through
     `attach_weighed_homes` into `_validate_moves`; the validator
     reads the homes the scale attached. The first Go draft
     threaded `[]NoteRow` (struct copies) and the aliasing broke
     SILENTLY: the weighed-evidence override flag never fired (the
     failing test proved it — the homes were attached to copies
     the validator never saw). Fixed by threading `[]*NoteRow`
     through the whole chain (MoveCandidates → the passes → the
     attachment → the validation), and documented as the spec's
     first Go-note amendment.
  2. **The repoint unmapped-folder fallback** — the reference's
     `_repoint_repo_ledger` falls back to the ROW's own category
     when the folder reverse-map misses; the draft's early return
     dropped the re-point entirely (the reference's own test leg
     caught it).

- **T-46 — NOT STARTED, deliberately.** T-45's own scope (the
  ~2,800-line module, the golden generator, the 215-test battery,
  the repobanish floor, the environment repair, two reality-wins
  finds) filled the session. The one-task law applied: one task
  done properly with proof beats two half-done; closed cleanly at
  the natural boundary.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-45's acceptance — PASS**: *the ported vault-scan suites* ✓ —
  the reference suites' core-layer selections all landed:
  test_vaultscan (the inventory, the candidates, the grammar
  reuse, the LLM pass incl. the chunked filing and the broken
  answers, the three laws, the apply pass, THE LANDABLE MOVE
  battery, the regex brain, the consolidation apply, the
  declared-home/sibling/taxonomy passes, the one-open scan),
  test_censusstamp (the grammar, the radar, the writer, the apply
  census, the review list, the second scan), test_picklock (the
  lock predicate/parse/radar/apply, the skip door, the picking
  door's pure + apply layers, the confidence law, the selection
  law), test_weighhomes (the word tokens, the teaching set, the
  census, the weighing folder set, the scale signal-by-signal,
  the attachment, the homes cell, the alignment door, the
  wiring), test_reposcan (the inventory/plan/apply kind doors,
  ONE NOTE PER REPO, the repos digest + prompt, the repos brain +
  weighed evidence, the round trip, the dry run). Plus the golden
  the house pattern warranted: 21 sections generated from the
  unmodified reference, byte-compared through the canonical
  Python dumps. The GUI legs stay with the future GUI task and
  the release-bookkeeping legs are the Python repo's own CI facts
  (the T-44 precedents); the hygiene leg rides T-46's suite;
  test_onespelling's vaultscan-relevant selections were verified
  to already live in T-23's state port.
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt
  -l` clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 -p 16 ./...` → **23 packages ok, 1347/1347 tests PASS,
  race detector clean, 0 failures** (was 1132 at session start;
  +215 — the arithmetic closes exactly); go.mod/go.sum UNCHANGED
  (`git rev-parse HEAD:go.mod` = 5c0d4e4a, the same bytes as the
  SESSION-019 close — zero new dependencies).
- **Honest notes**: (1) the golden was generated ONCE from the
  unmodified reference and never hand-edited (three pure-shape
  divergences during the bring-up were all Go-side test-shape
  bugs — None-key rendering, int-map normalization, a stale test
  entry — not port divergences; the port itself needed only the
  two reality-wins fixes above); (2) the `_windows_eyes` replay
  rides the new stat seam (fsExists/fsIsDir/fsIsFile) rather than
  patching os — the reference's mock.patch equivalent, documented
  in the spec; (3) the selection's junk-tolerance shapes are
  type-impossible at the Go door — every unrepresentable shape
  reads as its section's nil, the same verdict the reference's
  junk law gives (documented in the golden test and the spec).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-core-vault-scan.md` — the
  SESSION-020 shipped record (the files, the wiring, the
  repobanish-floor split, the test accounting) + the Go-note
  amendments: the pointer-threading law, the stat seam, the
  urlparse shapes, the Python-truthiness JSON door, the
  16,000-CHARACTER head read (the reference's constant is named
  BYTES but the text-mode read counts chars; the Go constant is
  named FrontmatterReadRunes), and the junk-tolerance
  type-impossibility.
- `docs/port-blueprint/traceability-matrix.md` — the D8 row
  widened (the complete scan, the golden + the five ported
  selections, the repobanish floor split; T-45 ✓).
- `README.md` — the status paragraph (T-45 complete, the
  amendments, 1347 race-clean, next: T-46).
- `internal/vaultscan/doc.go` — the complete package record (the
  file map, the pointer-threading law, the stat seam).
- `internal/repobanish/doc.go` — the landed-vs-remaining split
  (the gate's eyes + the burial landed with T-45; revive + the
  orchestration + test_repobanish stay T-46's).
- `internal/chrometabs/pyurl.go` — `ValidPyNetloc` exported (the
  bracket law shared with vaultscan's urlparse netloc mirror; the
  netloc laws' docstring kept).

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — ninth
  session carrying it)**: the main module rides `go 1.24.1`,
  forced by the owner-acknowledged tls-client v1.16.0 pin
  (ADR-0002 Am.2 has the record + the rollback path). A one-line
  acknowledgement (or a rollback instruction) closes it.
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- The standing sandbox-reliability observation carries (the
  FIFTEENTH degradation event, a full wipe; the §1 recipes +
  push-after-every-task both proved load-bearing again).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-021 = Phase 4 continues — T-46
  `internal/vaulthygiene` + the completion of `internal/
  repobanish`** (L) per spec-core-vault-hygiene.md +
  spec-core-repo-banish.md, then T-47 vaultseal if capacity
  remains.
- NEXT-SESSION.md rewritten: yes (v21).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
