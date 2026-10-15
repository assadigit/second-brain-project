# Module Spec — `integrations/goodrepos.py` + `integrations/repo_radar.py`

- **LOC**: 1,125 + 661 · **Go home**: `internal/goodrepos` (+ `radar.go`) · **Status**: port
  (merged group: both are the GitHub-publishing pair)
- **Responsibility**: publish the curated vault as the public "good repos" directory (staging
  tree + emoji README + git push); the read-only radar that detects which existing repo is each
  vault's mirror.

## goodrepos.py (`integrations/goodrepos.py:`)

- `IGNORED_DIRS {.obsidian,.trash,_inbox,_review,_moc,__pycache__,.git}` (86-88).
- Note parsing: tiny frontmatter subset (`_parse_frontmatter` 211-236) + `_parse_note`
  (307-367): title from `# `, TL;DR from `> **TL;DR:**` (fallback `## What is it?`).
- `CATEGORY_EMOJI` map (92-117) + `github_anchor` / `display_name` / `emoji_for` /
  `format_stars` helpers (130-173; `format_stars` renders 1.2k/15.4k style groupings).
- `_build_readme` (744-815): emoji-rich categorized README with counts and clickable lines.
- **GitHub REST via `_api` (537-563)**: base `https://api.github.com`; headers `Authorization:
  Bearer`, `Accept: application/vnd.github+json`, `X-GitHub-Api-Version: 2022-11-28`,
  `User-Agent: GitCurator-GoodRepos`; timeouts API 30 / git 60 / push-fetch 180 (72-74);
  `GET /user`; `GET /repos/{owner}/{repo}`; opt-in `POST /user/repos {name, private: true,
  description, has_issues: true, has_wiki: false, auto_init: false}` (630-637).
- **Git sequence** (staging dir via `tempfile.mkdtemp`, registered for atexit rmtree; env
  scrubbing incl. `GIT_TERMINAL_PROMPT=0`, 461-468): `git init -b main` (fallback symbolic-ref)
  → identity `GoodRepos Publisher <goodrepos@users.noreply.github.com>` → `fetch <token-url>
  main` + `reset --mixed FETCH_HEAD` (history alignment; FETCH_HEAD sanitized 676-689) →
  `add -A -- .` → `status --porcelain` + `rev-parse --verify -q HEAD` (skip when unchanged) →
  `commit -m "publish: N repos across M categories — <UTC ts>"` (`-c commit.gpgsign=false`) →
  `rev-parse --short=7 HEAD` → `remote remove origin` (defensive, before AND after push) →
  `push <token-url> HEAD:refs/heads/main` (986-989). Token law: one-time URL
  `https://x-access-token:<tok>@github.com/<o>/<r>.git`; `_redact` on all captured output.
- Behavior: skip when unchanged; empty `repo_name` ⇒ skip with reason (creation guard —
  `DEFAULT_REPO_NAME "good-repos"` kept only for import compat); dry-run builds the tree, keeps
  the staging dir, runs NO git; push failure keeps the local commit ("commit kept locally").
- `publish_from_config` + standalone CLI `_main` (1071-1121).

## repo_radar.py (`integrations/repo_radar.py:`)

- `RadarAPI` (97-176): read-only; UA `GitCurator-RepoRadar`, same GitHub headers, timeout 15 s;
  `me()`, `list_repos()` (`GET /user/repos?per_page=100&sort=pushed&page={1,2,3}`),
  `repo_exists` (404→None), `tree_blob_shas` (`GET /repos/{full}/git/trees/{ref}?recursive=1`).
- `parse_remote_repo` / `remote_origin_repo` (191-225): parse `github.com[/:]owner/repo(.git)`
  from the vault's `.git/config` `remote "origin"` (configparser).
- `blob_sha(data)` (232-238): git blob SHA-1 = `sha1(b"blob <len>\x00" + data)` — path-independent
  (matches GitHub tree SHAs).
- `sample_vault_files` (241-283): sorted walk (skip `.git/.trash/.DS_Store`, files ≤512 KiB),
  grouped by top-level folder, **round-robin**, max 40.
- `detect_backup_repo` (437-531) — the evidence ladder: (1) REMOTE (vault's own origin, verified
  with the token) = certain; (2) NAME (repo named after the vault folder exists) = certain;
  (3) FINGERPRINT: candidates ranked (non-archived/non-fork, vault-ish names
  `vault/obsidian/directory/library/notes/repos/websites` first, cap `MAX_TREE_FETCHES=24`);
  winner needs ≥`MIN_DOMINANT_MATCHES=3` matched blobs AND ≥2× the runner-up
  (`DOMINANCE_FACTOR`), else ambiguous/weak = no guess.
- `radar_all_vaults` / `apply_detections` (fills only EMPTY slots — typed names never
  overwritten) / `report_lines` / `radar_pending` (quiet-startup gate: token + ≥1 vault path +
  ≥1 empty slot).
- Repo API `size` explicitly NOT an emptiness signal (KB rounded-down — the live lesson,
  300-310).
- `Detection` dict crosses thread boundaries (kind/vault_path/repo_name/confidence/method/
  matched/sampled/detail).

## Tests

Fake-GitHub publish flow (unchanged skip, creation guard, private-create payload, redaction);
README/publish tree goldens; radar: blob_sha correctness vs real git hashes, ladder branches
(remote/name/fingerprint/dominance/ambiguous), apply-only-empty law, quiet-startup gate.
