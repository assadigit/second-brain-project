# SESSION-022 — 2026-10-17 — Phase 4 continues: T-47 vaultseal landed at d3a5d27 — 1550 tests race-clean (+86) — THE SEAL COMPLETE (the divergence ladder, the creation guard, the belts, the resync, the drift seal's pure half, CPython's fnmatch, the belt floor in directorysync); the sixteenth full wipe repaired before work (the PAT-passed clones with the extraheader scrubbed, go1.27.2 + the venv rebuilt, the baseline re-verified 1464 green BEFORE work)

## Goal (from NEXT-SESSION.md at start)

SESSION-022 = Phase 4 continues: **T-47 `internal/vaultseal`** (L) per
spec-integrations-vaultseal.md + ADR-0009 — the git-CLI-subprocess
seal of an Obsidian vault into a PRIVATE GitHub mirror. Acceptance:
the ported selections of the reference's real-git integration suites
(test_driftseal.py + test_sealfix.py — plus, as the acceptance's belt
clause demands, the belt's own suite test_localonly.py and the resync
suite test_resyncfix.py) — temp repos only, no real network. T-48
goodrepos + reporadar if capacity remained; close cleanly at a
natural boundary otherwise.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was WIPED (the sixteenth full wipe)**: neither
  repo, no Go, no venv, no symlinks — only the base Next.js scaffold
  survived. Rebuilt per NEXT-SESSION v22 §1 before any work: both
  clones via the owner's in-chat PAT passed ephemerally as the Basic
  `x-access-token` header (`git -c http.extraheader=…`), **scrubbed
  from BOTH configs immediately after cloning** (`git config
  --unset-all http.extraheader` — verified CLEAN in both), GitCurator
  checked out at `98f4fc7` (matches PINNED-GITCURATOR-REV
  byte-for-byte), go1.27.2 installed at /home/z/.local/go (go.dev
  tarball), the venv rebuilt (/home/z/venvs/gitcurator, Python
  3.12.15, stdlib-only), the symlinks re-created. The go.mod blob
  verified UNCHANGED (`5c0d4e4a48d2…`) and the baseline re-verified
  BEFORE work: 1464/1464 race-clean over 25 packages, vet + gofmt +
  GOOS=windows build clean.

- **T-47, commit `d3a5d27` (23 files, +6977)** — `internal/vaultseal`
  complete + the belt floor in `internal/directorysync` + the shipped
  config embed in `assets/directorysync/`:
  - *vaultseal* (11 files): the construction + `_git` runner
    (vaultseal.go — env scrubbing per goodrepos' `_git_env`
    [GIT_TERMINAL_PROMPT=0, the GIT_DIR-family stripped], timeouts
    60 s / push+fetch 180 s, the 124/127 doors, HideWindow on Windows
    [the llmclient precedent], **THE REDACTION LAW** — the token is
    masked in ANY captured output, a Go hardening the spec's Go notes
    mandate over the reference; the `_api` REST seam with the four
    headers + the 0-status network door; `ownerLogin` with the 401
    fresh-token hint); the bootstrap family (bootstrap.go — the
    nested-repo guard → `git init -b main` with the plain-init +
    symbolic-ref fallback, the managed .gitignore block, THE
    LOCAL-ONLY BELT's second managed block REBUILT from the current
    list, the untrack law — ONE `git rm -r --cached
    --ignore-unmatch` call, NEVER a disk delete; `ensureRepo` — the
    404 → creation-guard error, the opt-in POST /user/repos payload,
    the 422 race re-check, the token-less `remote add origin`); the
    seal (seal.go — the belt → `add -A -- .` → the unchanged skip →
    the commit-message doors → the one-time token URL push); THE
    DIVERGENCE LADDER (push.go — plain push → fetch +
    FETCH_HEAD redaction → stale-rejection retry / rebase /
    unrelated-histories merge / the `seal-rescue/<ts>` branch — main
    is NEVER force-pushed); the resync (resync.go — `--no-renames`
    delta counts, the write-tree + commit-tree reconciliation with
    both parents, the empty-tree baseline `4b825dc…`, the one
    fetch-first retry); `Status` (status.go — pure introspection, the
    `%h%x00%s%x00%aI` parse); the four config bridges (bridges.go —
    incl. manual's deliberately-NO-belt law and
    ResyncFromConfig's insertion-ordered map); the standalone CLI
    (cli.go); THE DRIFT SEAL's pure half (drift.go — the plan's legs
    with the batch-end guards verbatim + the runner with THE QUIET
    LAW and the v0.90 manual door; the publish door rides
    DriftSealDeps.PublishAfterBatch — T-49's seam, with the honest
    not-wired warning until then); **CPython 3.12's fnmatch**
    (fnmatch.go — the full translate port with the RE2 convergences:
    atomic groups → the non-capturing twin, `(?!)` → the never flag,
    `\Z` → `\z`).
  - *the belt floor* (internal/directorysync/belt.go —
    LocalOnlyNotePatterns, PrivateExclusionPatterns [the
    safe_filename folder forms + the disk two-level scan matched
    folded + the note globs + the vault's own local_only, deduped],
    LoadDirectoryConfig [→ the EMBEDDED shipped default at
    assets/directorysync/website-directory-config.json,
    sha256-pinned → the floor dict], PrettifyFolderName,
    gitignoreEscape, beltFolds) — the T-45 repobanish-floor
    precedent: T-47's bridges consume them; T-49 completes the
    module.
  - *THE GOLDEN*: testdata/vaultseal.golden.json (16 sections)
    GENERATED by testdiff/goldens/dump_vaultseal.py from the
    unmodified reference @ 98f4fc7 — paths folded to `<vault>/…`, the
    CLOCK frozen (the message doors), the GIT DATES frozen
    (deterministic commit shas for status()).
  - *The behavior batteries* (86 new top-level tests): the ported
    test_sealfix selections (TestSealPushReconcile's nine legs
    against REAL local bare repos + TestCreationGuardV0830's five
    doors), the ported test_resyncfix selections (13), the ported
    test_localonly selections (the managed block + the untrack law +
    the bridge's belt facts; the belt floor's pure builders in
    belt_test.go), the ported test_driftseal selections (the plan +
    the runner, 13), the 16-section golden differential, and the
    safety-law battery (laws_test.go — the env scrub with the
    hijack-defeat proof, the token-never-in-.git/config law, the
    redaction law, the CLI exit codes).
  - *Spec + records amended in the same commit*:
    spec-integrations-vaultseal.md gained the SESSION-022 shipped
    record + the Go-note amendments; spec-cloud-directory-sync.md
    gained the belt-floor record; the D11 row stamped; README's
    status paragraph; both doc.go records.

- **Pushed**: `4f2a8e1..d3a5d27 main -> main` (the ephemeral PAT via
  the transient `-c http.extraheader` Basic form; the config verified
  CLEAN after — no persisted header; HEAD == origin/main == `d3a5d27`).

- **T-48 — NOT STARTED, deliberately** (the one-task law): T-47's
  scope — the 1,301-LOC module + the ladder's every branch + the belt
  floor + the golden + 86 tests + the spec amendments — filled the
  session's capacity. A (M)-sized two-module task (goodrepos +
  reporadar) rushed at the tail would violate "one task done
  properly with proof beats a rushed whole."

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-47's acceptance — PASS**: *the ported selections of the
  reference's real-git integration suites* ✓ — test_sealfix's
  TestSealPushReconcile (the divergence ladder's EVERY branch:
  fast-forward, up-to-date, shared-history rebase, unrelated
  merge-keeps-both, the unrelated-conflict rescue with the
  remote-main-untouched pin, the rebase-conflict rescue, the
  non-rejection pass-through, the FETCH_HEAD token scrub, the
  seal-result push/rescue wiring) + TestCreationGuardV0830 (the
  404-without-opt-in with the zero-POST proof, the POST-once private
  payload, the empty-name NOT-CONFIGURED with the zero-calls proof,
  the existing-repo, the bridge opt-in) — **temp repos only, never
  the owner's vault; NO real network** (the GitHub API surface is
  the instance-seam fake exactly like the reference's own tests; the
  "remote" is a local bare repo). The acceptance's named clauses all
  land: first seal ✓ (the smoke + the belt cycle), unchanged skip ✓,
  the ladder's every branch ✓, the belt untrack law (files stay on
  disk — asserted with os.Stat pins) ✓, resync delta counts ✓ (the
  ported test_resyncfix selections), token-redaction assertions ✓
  (FETCH_HEAD + any captured output + .git/config), the creation
  guard ✓. Plus THE GOLDEN the house pattern warranted: 16 sections
  generated from the unmodified reference, with the drift-seal
  pure-half selections (test_driftseal §1-2) byte-differentially
  verified against the reference's own pure half (four
  manual/automatic cases — the ⚡/🛡️/🌐/✅/⚠️ lines verbatim).
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 -p 16 ./...` → **27 packages ok, 1550/1550 top-level
  tests PASS, 0 failures, 0 skips, 0 data races** (was 1464 at
  session start; +86 — the arithmetic closes exactly; the package
  count went 25 → 27: vaultseal is new AND directorysync gains its
  own battery); go.mod/go.sum UNCHANGED (`git rev-parse HEAD:go.mod`
  = `5c0d4e4a48d2b3549ea24e251b8b419050f7ed17` — zero new
  dependencies; fnmatch/regexp/crypto-sha256 are stdlib).
- **The golden caught a real divergence live** (the house pattern
  working exactly as designed): the CRLF tail case — Python's
  read_text translates universal newlines (`\r\n` → `\n`) before
  `ensure_gitignore`'s `rstrip("\n")` quirk fires; the port read raw
  bytes and diverged until readGitignore rode
  `helpers.PyNewlines(helpers.DecodeUTF8Replace(data))` — the T-41
  law applied to this module, recorded in the spec's Go notes.
- **Honest notes**: (1) the reference's `_Boom`-config drift-seal
  case (a dict whose .get raises) maps to raise-door panics
  (driftDictDoor/driftStrDoor) + the plan's recover — the same
  silence, verified against the reference's pure half extracted and
  exec'd with stub modules (the reference's backup_seal.py cannot
  import headlessly — PyQt6 guards the module header; the pure half
  itself is Qt-free); (2) the reference's bridge test that stubs the
  VaultSeal class wholesale maps to the ctor seam
  (`vaultSealCtor`) + the builder split (sealerFromConfig /
  websitesSealerFromConfig / manualSealerFromConfig) — the same
  captured-kwargs facts; (3) the mock-specific broken-belt test
  (patching the belt builders to raise) has no Go analog — the
  static import cannot fail; documented in the spec's Go notes; (4)
  the drift seal's publish door speaks the honest not-wired warning
  until T-49 lands (never a silent skip).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-integrations-vaultseal.md` —
  the SESSION-022 shipped record + the Go-note amendments: the
  universal-newlines find, the fnmatch RE2 convergences, the
  mock-surface → instance seams, the drift-seal pure half's home +
  the publish seam, the belt's import-failure door, the
  never-raises wrappers, ~user expanduser, the non-JSON API body's
  parse door, the vaultRepr "None" quirk, RunSummary's *int shape,
  the embedded shipped config.
- `docs/port-blueprint/module-specs/spec-cloud-directory-sync.md` —
  the belt-floor shipped-early record (T-47; T-49 completes).
- `docs/port-blueprint/traceability-matrix.md` — the D11 row stamped
  (T-47 ✓).
- `README.md` — the status paragraph (T-47 complete, 1550
  race-clean, next: T-48).
- `internal/vaultseal/doc.go` + `internal/directorysync/doc.go` —
  the package records.

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — eleventh
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2
  has the record + the rollback path). A one-line acknowledgement
  (or a rollback instruction) closes it.
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- **The drift-seal publish door**: wired to
  DriftSealDeps.PublishAfterBatch (nil until T-49's directorysync
  publish lands; the runner speaks the honest not-wired warning) —
  the GUI worker task wires the real one after T-49.

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session
   (it worked exactly as designed this session: the sixteenth-wipe
   rebuild + the push both rode the ephemeral form).

## Handoff

- Next session: **SESSION-023 = Phase 4 continues — T-48
  `internal/goodrepos` + `reporadar`** (M) per
  spec-integrations-goodrepos.md (README/publish tests with fake
  GitHub; radar dominance tests — ADR-0009's second and third
  CLI-subprocess consumers), then T-49 directorysync (L) if capacity
  remains.
- NEXT-SESSION.md rewritten: yes (v23).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
