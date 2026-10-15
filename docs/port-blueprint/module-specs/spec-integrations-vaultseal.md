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
