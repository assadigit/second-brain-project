# Module Spec — `integrations/vaultseal.py`

- **LOC**: 1,301 · **Go home**: `internal/vaultseal` · **Status**: port (ADR-0009: git CLI
  subprocess)
- **Responsibility**: seal an Obsidian vault into a **private** GitHub mirror after every run —
  `git add -A → commit → push` with the reconciliation ladder; never raises; never force-pushes
  main; token never touches `.git/config`.

## Construction (183-216)

`vault_path` (expanduser), `token`, `repo_name` (**empty = NOT configured** — hard error before
any network), `auto_push = auto_push AND token`, `enabled`, timeout 60 s, `allow_create`
(the v0.83 creation guard — 404 is an ERROR spoken fast; opt-in creates **private**), `local_only`
patterns (the belt).

## `status()` (973-1009) — pure introspection, no network

`{vault_path, exists, is_git_repo (rev-parse --show-toplevel realpath == vault realpath),
has_commits (rev-parse --verify -q HEAD), dirty_count (status --porcelain lines), last_commit
{sha, subject, date} (log -1 --format=%h%x00%s%x00%aI), remote_url, repo_name}`.

## `seal()` algorithm (528-602 + ensure_repo 425-516)

1. disabled → skip; vault missing → error.
2. `ensure_repo`:
   - **Nested-repo guard**: `rev-parse --show-toplevel` mismatch ⇒ `git init -b main` (fallback
     `git init` + `symbolic-ref HEAD refs/heads/main`) — `add -A` is repo-wide in git ≥2.0.
   - `ensure_gitignore` (294-318): merge the managed block
     `# --- VaultSeal (machine-specific state — not backed up) ---` + DEFAULT_IGNORE_LINES
     (`.obsidian/workspace.json`, `workspace-mobile.json`, `.bak`, `cache`, `.trash/`,
     `__pycache__/`, `*.pyc`, `.DS_Store`, `Thumbs.db`, `desktop.ini`, `*.tmp`, `*.bak`,
     `~$*`); idempotent; atomic write via `.gitignore.tmp` + `os.replace`.
   - `_ensure_identity` (731-738): local-only `config user.name "GitCurator VaultSeal"` /
     `user.email "vaultseal@users.noreply.github.com"` — only when unset.
   - Remote: no push possible → done; empty repo_name → the creation-guard error; `GET /user`
     (401 → fresh-token hint); `GET /repos/{owner}/{repo}`; 404 + allow_create ⇒
     `POST /user/repos {name, private: true, description: "VaultSeal — automatic Obsidian vault
     backup (GitCurator)", has_issues: false, has_wiki: false, has_projects: false,
     auto_init: false}` (422 race re-check); 404 without opt-in ⇒ guard error; `remote get-url
     origin` failing ⇒ `remote add origin https://github.com/{owner}/{repo}.git` (**token-less**).
3. `_apply_local_only` (320-328): `ensure_local_only_gitignore` (330-376 — a **SECOND** managed
   block `# --- VaultSeal (local-only notes — kept on this machine, never mirrored) ---`,
   REBUILT from the current list so a removed pattern disappears) then `_untrack_local_only`
   (378-423): `git ls-files` → fnmatch (case-folded; folder patterns match prefixes, stem
   patterns match basenames) → one `git rm -r --cached --ignore-unmatch -- <victims>` (**NEVER
   a disk delete** — the files never leave the machine).
4. `git add -A -- .` (pathspec-limited to the vault subtree).
5. clean + has HEAD ⇒ skip "vault unchanged since the last seal".
6. `git commit -m "seal: {processed} note(s) curated ({processed}/{total} repos) — {YYYY-MM-DD
   HH:MM UTC}"` (or "seal: vault snapshot — …"); `rev-parse --short=7 HEAD`.
7. No auto_push ⇒ commit kept locally (warning).
8. **Push with the one-time token URL** `https://x-access-token:<token>@github.com/<o>/<r>.git`
   via `_push()` (643-713) — THE DIVERGENCE LADDER:
   - `git push <url> HEAD:refs/heads/main` (180 s) → rc 0 done.
   - Error NOT containing "fetch first"/"non-fast-forward" ⇒ reported verbatim (auth/network).
   - `git fetch <url> refs/heads/main` (180 s) → `_sanitize_fetch_head` (606-623: rewrite
     `.git/FETCH_HEAD` replacing the token with `***`).
   - `merge-base --is-ancestor FETCH_HEAD HEAD` == 0 ⇒ stale rejection → plain push retry.
   - `merge-base HEAD FETCH_HEAD` == 0 ⇒ `git rebase FETCH_HEAD`; fail → `rebase --abort` →
     `_rescue()`.
   - else ⇒ `git merge FETCH_HEAD --allow-unrelated-histories -m "seal: reconcile mirror
     histories (v0.21.0)"`; fail → `merge --abort` → `_rescue()`.
   - `_rescue` (625-641): `git push <url> HEAD:refs/heads/seal-rescue/<YYYYmmdd-HHMMSS>`; the
     error names the branch ("Compare & pull request …"); **main is NEVER force-pushed**.
9. After success the sha is re-read (reconciliation may have rebased).
10. `seal()` wraps all in try/except — a backup failure must never fail the curation run.

## `resync()` (820-969)

The owner's "mirror follows the vault EXACTLY" action: stage `add -A`; fetch; mirror empty or
ancestor → ordinary commit with `write-tree`-based delta counts
(`resync: mirror follows the vault — {deleted} file(s) deleted, {added} added/updated (<ts>)`)
then push; else the reconciliation commit `git write-tree` → `git commit-tree <tree> -p
<local> -p <remote> -m <msg>` → `git update-ref HEAD <sha>` → push; one "fetch first" retry;
never force — falls to `_rescue`. Baseline for an empty mirror = the empty tree SHA
`4b825dc642cb6eb9a060e54bf8d69288fbee4904` (729).

## Config bridges

`seal_from_config` (github vault, note-stem belt) · `websites_seal_from_config` (pipeline-gated,
FULL belt = directory's `private_exclusion_patterns`) · `manual_seal_from_config` (personal
vault — **deliberately NO belt**: "a backup that filters the owner's own files is not a backup")
· `resync_from_config` (all three). Standalone CLI (1247-1297): `--vault --repo-name
--create-repo --token[=$GITHUB_TOKEN] --no-push --resync --json --status`.

## Go notes

`exec.Command("git", …)` sequences with the exact arguments, env scrubbing
(`GIT_TERMINAL_PROMPT=0`, strip `GIT_DIR`/`GIT_WORK_TREE`/`GIT_INDEX_FILE`/`GIT_CONFIG`/
`GIT_CONFIG_NOSYSTEM`), `HideWindow` on Windows, redaction of token in any captured output,
FETCH_HEAD rewrite. Timeouts: API 30 s, git 60 s, push/fetch 180 s.

## Tests

Real-git integration tests on temp repos: first seal (init/identity/ignore blocks), unchanged
skip, divergence ladder each branch (stale-rejection, rebase path, unrelated-histories path,
rescue), belt untrack (files stay on disk), resync delta counts, token redaction (assert
`.git/config` and logs never contain the token), creation guard (404 → error; opt-in → private
repo payload).

## Shipped record — T-47, SESSION-022 (2026-10-17)

**Landed** (`internal/vaultseal`, 11 files + the golden generator +
the golden + the behavior battery; + the belt floor in
`internal/directorysync/belt.go` + the shipped-config embed at
`assets/directorysync/`): the construction (`vaultseal.go` — the
constants, `SealResult` with ok/describe/to_dict [the None doors
re-opened as null through `pyNoneIfEmpty`], the constructor's doors
[strip, expanduser, the autoPush AND token degradation, timeout 0 →
60], `Slugify`, the `_git` subprocess runner [env scrubbing per
goodrepos' `_git_env`, timeouts 60 s / push+fetch 180 s, the 124/127
timeout/missing-git doors, **THE REDACTION LAW — the token is masked
in ANY captured output, a Go hardening the spec's Go notes mandate**],
the `_api` REST seam [Bearer, the four headers, 30 s, the 0-status
network door], `ownerLogin` [the cached-on-success retry-on-failure
law, the 401 fresh-token hint]), the bootstrap family
(`bootstrap.go` — `pyRealpath`, the managed `.gitignore` block
`EnsureGitignore`, `EnsureLocalOnlyGitignore` [the REBUILT second
managed block], `untrackLocalOnly` [the fnmatch victim ladder: folder
patterns prefix-or-wrapped, stem patterns basename-or-wrapped-or-
exact, case-folded; ONE `git rm -r --cached --ignore-unmatch --` call
— NEVER a disk delete], `ensureIdentity`, `EnsureRepo` [the
nested-repo guard → `git init -b main` with the plain-init +
symbolic-ref fallback, the 404 → guard error, the opt-in
`POST /user/repos` payload, the 422 race re-check, the token-less
`remote add origin`]), the seal (`seal.go` — `commitMessage` [the
processed/total ints door vs the snapshot], `sealInternal` [the belt
→ `add -A -- .` → the unchanged skip → the commit → the one-time
token URL push with the sha re-read after reconciliation]), THE
DIVERGENCE LADDER (`push.go` — plain push → the fetch-first/
non-fast-forward gate → fetch + `sanitizeFetchHead` → the
stale-rejection retry / rebase-then-push [abort → rescue on
conflict] / merge-allow-unrelated-histories [abort → rescue on
conflict] → `rescue` [the `seal-rescue/<ts>` branch, main NEVER
force-pushed]), the resync (`resync.go` — `deltaCounts` with
`--no-renames`, `resyncPlain` [the write-tree delta counts + the
empty-tree baseline], the reconciliation path [write-tree →
commit-tree with both parents → update-ref → push → the one
fetch-first retry on a fresh tip] → rescue), `Status` (`status.go` —
pure introspection, the `%h%x00%s%x00%aI` parse with the 3-tuple
pad), the four config bridges + `_local_only_for` (`bridges.go` —
`SealFromConfig`/`WebsitesSealFromConfig`/`ManualSealFromConfig`
[the deliberately-NO-belt law]/`ResyncFromConfig` [the
insertion-ordered {github, websites, manual} map]), the standalone
CLI (`cli.go` — the full flag surface, exit 0/1 by result.ok, exit 2
on flag errors), THE DRIFT SEAL's pure half (`drift.go` — the three
legs with the batch-end guards verbatim, `DriftSealPlan` [the
raise-door mapped to the recover silence], `RunDriftSeal` [THE QUIET
LAW, the publish-rides-only-the-websites-leg law, the v0.90 manual
door with the always-speaks + always-publishes laws, the
raising-leg recover]), and CPython 3.12's fnmatch (`fnmatch.go` —
`FnMatch` + the full `translate` port [the class grammar's chunk
splitting, empty-range handling, set-op escapes]).

**The belt floor** (`internal/directorysync/belt.go`): the v0.89
pattern builders — `LocalOnlyNotePatterns`, `PrivateExclusionPatterns`
(the safe_filename folder forms + the disk two-level scan matched
folded + the note globs + the vault's own local_only list, deduped,
folder-forms-first), `LoadDirectoryConfig` (the vault's
`_system/directory-config.json` → the EMBEDDED shipped default at
`assets/directorysync/website-directory-config.json` [sha256-pinned
by the golden] → the floor dict), `PrettifyFolderName`,
`gitignoreEscape`, `beltFolds` — landed with T-47 exactly the way
T-45 landed the repobanish floor (the vaultseal bridges consume
them; T-49 completes the module).

**Tests** (86 new top-level tests, 1464 → 1550 race-clean over 27
packages): the ported test_sealfix selections
(`vaultseal_test.go` — TestSealPushReconcile's nine legs [the
fast-forward, the up-to-date, the shared-history rebase, the
unrelated-histories merge, the two conflict-rescue branches with the
remote-main-untouched pins, the non-rejection pass-through, the
FETCH_HEAD scrub, the seal-result push/rescue wiring] +
TestCreationGuardV0830's five doors [the 404-without-opt-in with the
zero-POST proof, the POST-once payload, the empty-name
not-configured with the zero-calls proof, the existing-repo, the
bridge opt-in]); the ported test_resyncfix selections
(`resync_test.go` — the deletions-reach-the-mirror owner report, the
idempotence, the fresh-vault-never-resurrects, the
mirror-ahead-vault-wins, additions+deletions, the empty mirror,
describe/to_dict, the guards, `resync_from_config`'s
explicit-action semantics [the ctor-seam capture]); the ported
test_localonly selections (the managed block's merge/rebuild/
coexistence, the untrack law [files stay on disk] + the _review
twin + the seal-applies-the-belt cycle, the config bridge's belt
facts, the belt floor's pure builders in
`internal/directorysync/belt_test.go`); the ported test_driftseal
selections (`drift_test.go` — the plan's legs + the
runner's laws [the quiet, the counts, the publish door, the error
masking, the raising leg, the manual door]); the safety-law battery
(`laws_test.go` — the env scrub with the hijack-defeat proof, the
token-never-in-.git/config law, the redaction law, the CLI's exit
codes); and THE GOLDEN (16 sections — slugify, describe, to_dict,
messages, owner_login, creation_guard, gitignore_merge,
local_only_gitignore, untrack_belt, untrack_review_twin,
delta_counts, fnmatch, status, local_only_for, fetch_head_scrub +
the shipped-config sha pin) GENERATED by
`testdiff/goldens/dump_vaultseal.py` from the unmodified reference
@ 98f4fc7 (the plain integrations/ generator pattern; paths folded
to `<vault>/…`, the CLOCK frozen [the message doors], the GIT DATES
frozen [deterministic commit shas for status()]).

**Go-note amendments (found at implementation time):**
- **THE UNIVERSAL-NEWLINES FIND (the golden caught it live)**:
  `read_text` on the .gitignore translates `\r\n` → `\n` before the
  `rstrip("\n")` quirk fires — the port reads through
  `helpers.PyNewlines(helpers.DecodeUTF8Replace(data))` (the T-41
  law applied to this module; the golden's CRLF tail case diverged
  until it landed).
- CPython 3.12's fnmatch.translate emits `(?>.*?fixed)` atomic
  groups and `(?!)` never-match lookaheads — Go's RE2 has neither:
  the atomic group becomes the non-capturing twin (the same match
  language for whole-string boolean matching), the never-match
  short-circuits FnMatch to false, and `(?s:…)\Z` becomes `…\z`.
- The reference's mock surface (instance-level `_api`, patched
  `ensure_repo`/`_push`, the class-patch in resync_from_config's
  tests) maps to instance seams: `apiFn`/`pushFn`/`ensureFn` fields
  bound in `NewVaultSeal`, and the package-level `vaultSealCtor` for
  the resync bridge's construction capture.
- The drift-seal pure half (backup_seal.py's `_drift_seal_plan` +
  `_run_drift_seal`) lands IN vaultseal (drift.go): the reference
  keeps it in the GUI package but pure ("no Qt, no GUI thread"); the
  QThread wrapper is the future GUI task's. The publish door
  (`publish_after_batch`) rides `DriftSealDeps.PublishAfterBatch` —
  T-49's directorysync surface; until it is wired the runner speaks
  the honest not-wired warning on the websites-sealed path (never a
  silent skip). The reference's `_Boom`-config silence maps to
  raise-doors (`driftDictDoor`/`driftStrDoor` panic on
  truthy-but-wrong-typed values) + the plan's recover.
- The belt's import-failure door (`except Exception: return []`)
  has no Go analog (the static import cannot fail; the builders
  never error) — a documented convergence.
- The never-raises wrappers (seal/resync) are unreachable in Go
  (the T-46 law); the honest result is the answer.
- `Path.expanduser`'s `~user` form is not ported (the config never
  carries it — the house usermemory precedent).
- A non-JSON (or non-object) API body answers the parse door
  (status 0 with the parse error as the message) — Python's
  json.loads-inside-try equivalence.
- `f"vault directory not found: {self.vault}"` renders None for an
  unconfigured vault — `vaultRepr` keeps the quirk ("None").
- The run_summary isinstance-int door maps to `RunSummary{Processed,
  Total *int}` (nil = the key was absent).
- The shipped directory-config is EMBEDDED (assets/directorysync,
  sha-pinned) where the reference reads app/taxonomy/ from disk —
  the doc 04 §1 assets layout.
