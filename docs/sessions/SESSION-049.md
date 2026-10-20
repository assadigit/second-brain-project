# SESSION-049 — 2026-10-20 — Phase 7 (second row): T-81 LANDED with the dry-run's runner legs blocked on the account's Actions billing (the owner action); THE TAG-STATE BUILD LAW found; 2377 main-battery tests over 36 packages + 89 webui over 4

## Goal (from NEXT-SESSION.md at start)

T-81 (S): the release workflow (doc 10 §8): tag push → build both exes
→ mkzip → attach `GitCurator-v<version>-windows.zip` to the GitHub
release (the owner's law: a release without its zip is not done). The
row's acceptance: "dry-run on a test tag" — the workflow wraps
`build/release.sh`; the existing `go-ci.yml` untouched (a NEW workflow
file under a new path).

## What was done (task IDs, files, commits — cite hashes)

- **The forty-fifth full wipe repaired BEFORE work**: the ENTIRE
  environment was gone at boot (both repos, the toolchain, the venv,
  the wails CLI). Everything rebuilt from the committed recipes: both
  repos re-cloned (the ephemeral GIT_CONFIG_COUNT PAT form; `git
  remote -v` clean), the reference at `98f4fc71cfb862cee7225d864636c08062724348`
  = testdiff/PINNED-GITCURATOR-REV exactly, go1.27.2 (the local sha256
  matching TOOLCHAIN.md Am.2's committed 64-char pin byte-for-byte),
  the venv (Python 3.12.15, the plain ensurepip form), the wails CLI
  v2.16.0. No webkit rebuild needed (no linux GUI build this session).
  The FULL battery re-verified green pre-work: **2371 main + 89 webui,
  plain + race — the exact committed numbers.**
- **T-81 core (commit `b5c9a27`)**:
  - `.github/workflows/release.yml` — the NEW workflow (go-ci.yml
    untouched): the tag-only door (`on: push: tags: ['v*']` — no
    branch trigger, no pull_request), the toolchain pins (go-version
    '1.27.2' + `go install …wails@v2.16.0`), **the tag law**
    (one-version-truth for tags: a real release tag must equal
    `v$(cat VERSION)`; the reserved dry-run prefix `v0.0.0-` rides with
    the assertion waived → DRYRUN=1 → a draft+prerelease release titled
    `[DRY-RUN] <tag>`), the ritual wrap (`./build/release.sh` — the
    single owner of the build commands; ubuntu-latest; no npm — the
    frontend is committed vanilla), the AC-23 notes step (python:
    re-derives + asserts A==B==cold, renders the parity table + the
    zip's files/bytes/sha256 + the reproduce-from-tag recipe + the
    [DRY-RUN] banner), the attach step (stages
    `GitCurator-v${VERSION}-windows.zip` from the ritual's `.a`
    artifact; delete-first convergence; `gh release create`;
    `permissions: contents: write`; the job's own `github.token`; no
    secrets in the file).
  - `build/mkzip/workflow_test.go` — 6 law tests, all byte-pins over
    the committed workflow: TestReleaseWorkflowTrigger (the v* filter;
    NO branches/pull_request), …ToolchainPins (cross-checked against
    build/TOOLCHAIN.md), …RitualWrap (`./build/release.sh` present;
    NO `go build`/`wails build` in the workflow; ubuntu-latest),
    …TagLaw (the real-tag assertion + the v0.0.0- exemption), …AssetLaw
    (the exact zip name + the `.a` staging + the AC-23 table shape),
    …PermissionsAndDryRun (contents: write + `--draft --prerelease` +
    the [DRY-RUN] title).
- **T-81 rider (commit `0961ad9`)**: the notes' "bytes zipped"
  vocabulary aligned to mkzip's own (the zip FILE SIZE, not the sum of
  member compress_sizes — 27,225,660 vs 27,223,462 at the tagged
  state); **THE TAG-STATE BUILD LAW** recorded in release.sh's header +
  the workflow's ritual step (see the find below).
- **The local proofs (the strongest sandbox legs, all byte-level)**:
  - **The FULL ritual at the TAGGED state** (the runner's tree state —
    run as the ritual's steps in sequence): CLI exe `13808277…`
    (A=B=cold identical), GUI exe `99cc22e6…` (A=B=cold — T-80's own
    number), the zip `071e1968…` (A=B; 13 files, 56,908,881 bytes in,
    27,225,660 zipped).
  - The workflow's own notes step proven by **EXTRACT-AND-EXEC** from
    the workflow's own bytes (regex-pinned heredoc extraction; never
    hand-transcribed) — the full AC-23 body rendered over the leg set.
  - The asset-staging leg proven byte-equal (staged == `.a`).
  - The 45th-wipe toolchain reproduced T-80's evidence byte-for-byte:
    a CLI build at `20356d6` in a temp worktree → `82bfa4ec…` exactly.
- **The dry-run attempt (the acceptance leg)**: the test tag
  `v0.0.0-t81-dryrun` pushed (first at `b5c9a27`, then moved to
  `0961ad9` so the re-run rides the FIXED workflow). Both runs
  **failed at job setup — the account's Actions billing refused every
  runner job** (the evidence below). The blocker surfaced with
  evidence, not improvised past: no release was created, no tag
  deleted (the tag stays for the re-run), and the owner's fix + re-run
  recipe recorded everywhere it matters.
- **T-81 evidence + spec updates (commit `e6927b9`)**:
  `docs/poc-artifacts/t-81-release-workflow.md` (the laws, the local
  proofs, the find, the blocked leg + the re-run recipe); doc 10 §8's
  dated note; doc 13's T-81 row (the honest PARTIAL); doc 12's
  SESSION-049 entry; README's SESSION-049 block.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PARTIAL — honest.** T-81's acceptance = "dry-run on a test tag":

- **LANDED + locally proven (GREEN)**: the workflow file (validated
  YAML — registered `active` on the Actions API; 7 steps in order),
  the trigger (both tag pushes created runs at the right `head_sha`),
  the 6 law tests (the build/mkzip package now 16, race-clean 3×), and
  every command the workflow runs proven in this sandbox: the full
  ritual at the tagged state (the table above), the notes render, the
  asset staging. The close battery: **2377 main-battery tests
  race-clean over 36 packages (2371 + 6) + 89 webui over 4; gofmt/vet
  clean over both modules; the MAIN go.mod + go.sum byte-unchanged.**
- **BLOCKED (the runner legs)**: the dry-run's runs
  `39511458026` (at `b5c9a27`; attempt 1 + a re-run) and `39512685975`
  (at `0961ad9`, the current tag target) each died in ~2 s at setup —
  `runner_id: 0`, zero steps — with the annotation: *"The job was not
  started because recent account payments have failed or your spending
  limit needs to be increased. Please check the 'Billing & plans'
  section in your settings."* The account is on the **free plan**
  (private-repo Actions minutes exhausted / no spending headroom).
  **The block PREDATES T-81**: go-ci (the untouched main CI) has
  failed the same way on every recent main push — including
  SESSION-048's close `96ce15e` (run `39505614616`, the same
  annotation) — 8/8 recent runs; nobody was watching the Actions tab
  (the battery runs locally). The acceptance's runner legs (the
  end-to-end build + zip + attach + the cross-machine sha comparison
  vs `071e1968…`) complete at the re-run after the owner's billing
  fix.

### THE FIND (byte-settled): THE TAG-STATE BUILD LAW

The same commit built untagged vs tagged produces **different CLI-exe
bytes**. Isolation chain: (1) an A-leg build at `b5c9a27` (untagged) →
`6fc30b51…`; after creating the tag, a rebuild at the same commit →
`13808277…` (stable, `.a`==`.b`). (2) The full `go version -m`
buildinfo diff (t1's exe recovered byte-exact from the zip member):
the ONLY difference — the main module's version line:
`v0.92.0-go.1.0.20261020225135-b5c9a27efb95` (a pseudo-version derived
from the ancestor tag) vs `v0.0.0-t81-dryrun` (the tag). Go 1.24+
embeds the VCS-derived module version. (3) The exe byte-diff: one
~24-byte shift re-locates the layout (10.4M differing bytes, ~29.8k
ranges — the build ID + every relocated operand). Consequences: the
identity claims are per-tree-state; the release numbers live at the
tagged checkout (the runner's state; three zip states measured:
`61b4cf2a…` → `a92b53d8…` → `071e1968…`); the wails-built GUI exe is
exempt (`(devel)` + a `v0.0.0` parent dep — byte-stable across commits
AND tag states, re-proven); the tag law makes the versions agree on a
real release (the tag becomes the buildinfo version == the -X stamp).
Also settled en route: my own zip-member extraction bug (the filter
`'cli' in name` matched `GitCurator-CLI-Setup.bat`) — the "third sha"
was the bat, not a mystery; the zip's member proven byte-equal to the
A-leg exe after the fix. (And the standing verify-the-verifier law
held: every suspect number was re-derived from bytes, not display.)

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/10-build-release-packaging.md` — §8's dated
  note: the landed shape + the blocked dry-run + the tag-state law.
- `docs/port-blueprint/13-implementation-task-list.md` — the T-81 row:
  LANDED + locally proven; the runner legs blocked on the owner's
  billing fix (the honest PARTIAL record with the re-run recipe).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-049
  entry.
- `build/release.sh` — the header's TAG-STATE BUILD LAW note (reality
  wins: the next releaser must not compare a tagged build against an
  untagged one).
- `README.md` — the SESSION-049 status block.
- `docs/poc-artifacts/t-81-release-workflow.md` — the evidence doc.
- `docs/NEXT-SESSION.md` — rewritten for SESSION-050 (the state card
  v50).
- `docs/sessions/SESSION-049.md` — this report.

## Blockers & decisions needed

- **THE ACTIONS BILLING BLOCK (the owner's, URGENT for T-81's
  completion)**: GitHub → Settings → Billing & plans — add/fix the
  payment method or raise the spending limit (the private repo's
  Actions minutes). Until then NO workflow can run (go-ci included —
  it has been silently failing the same way on every main push). The
  re-run recipe: re-run `39512685975` (the tag `v0.0.0-t81-dryrun`
  still points at `0961ad9`) → verify the green run + the `[DRY-RUN]`
  draft release + the zip asset (cross-machine sha256 vs
  `071e1968…`) → optionally the refusal leg (`v9.9.9-t81-refusal`
  must be refused at the tag-law step) → delete the test release +
  tag. Full recipe: the evidence doc §3.
- Carried, unchanged: the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (the
  fortieth session carrying it); x/net/html keep-or-swap; the
  answering-window belt + the T-71 merged-verdict question; the
  deferred test debt; the batch-scorecard modal question; the v0.91.0
  zip-parity reading.
- NEW, minor: whether the owner ALSO wants go-ci watched (it has been
  red-on-every-push for a while due to the same billing block; once
  billing is fixed it goes green on the next push by itself).

## Owner actions requested

1. **The billing fix** (above) — the only urgent one; it unblocks
   T-81's dry-run AND the silently-red go-ci.
2. When convenient: the standing calls (dissolution, vanilla-vs-vite,
   the scorecard-modal question).

## Handoff

- **SESSION-050 = complete T-81's dry-run (if billing fixed: the
  re-run + the verification + the test-tag deletion, ~15 minutes) then
  T-82's preparable legs (the clean-VM checklist document + the AV
  scan note + the release-report scaffold — the owner's live Windows
  execution records land into it as they arrive).** T-83 (the docs)
  follows.
- NEXT-SESSION.md rewritten: yes (the state card v50, the SESSION-050
  prompt inside).
- Push state: all work committed + pushed at the close (the ephemeral
  PAT; never stored, never written).
