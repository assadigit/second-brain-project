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

## SESSION-023 shipped record (T-48 — both modules COMPLETE)

`internal/goodrepos` (goodrepos.go / notes.go / readme.go / publisher.go / bridges.go /
cli.go + the platform twins) and `internal/reporadar` (radar.go / evidence.go / detect.go)
landed at SESSION-023: 1658/1658 tests race-clean over 29 packages (+108). Evidence:

- **goodrepos**: the ported selections of the reference's whole test_goodrepos.py
  (TestPublish / TestAnchors / TestOrdering / TestNoteParsing / TestLifecycle /
  TestConfigBridge / TestVaultHygiene / TestHelpers / TestCreationGuardV0830) + the
  local-bare-repos publish-flow battery (the full cycle incl. the second-publish
  skip through the fetched alignment, the push-failure that keeps the commit with
  the token REDACTED, the FETCH_HEAD scrub through a real fetch) + the safety-law
  battery (the env scrub with the hijack-defeat proof, the 127 no-staging door,
  the redaction laws, the CLI flag surface + exit codes) + THE GOLDEN (25 sections,
  generated by testdiff/goldens/dump_goodrepos.py — paths folded to
  `<vault>`/`<staging>`, the CLOCK frozen, the GIT DATES frozen so the publish
  commit SHA is byte-deterministic: the Go port produces the reference's exact
  `a300991`).
- **reporadar**: the ported selections of the reference's whole test_reporadar.py
  (the pure evidence pieces incl. the real-git `hash-object` cross-check, the
  ladder's every branch, the three slots together incl. the never-raises doors,
  apply/report/pending, the read-only law, the constants contract, the owner's
  exact end-to-end scenario) + THE GOLDEN (12 sections, generated by
  testdiff/goldens/dump_reporadar.py — paths folded to `<root>`).

## Go-note amendments (found at implementation time, evidence at hash level)

- **atexit → CleanupStaging**: Go has no process-exit hook; the reference's
  `atexit.register(_cleanup_staging)` maps to the exported `CleanupStaging()` (the
  CLI defers it; the app wiring calls it at shutdown). Dry-run dirs are never
  registered — the reference's inspection door stands.
- **copy2 → mode-preserving content copy**: git only ever sees content + the exec
  bit; `shutil.copy2` converges to read + write-with-source-mode.
- **The instance seams**: `apiFn` is the reference's own mock surface (the
  creation-guard suite patches `_api`); `tokenURLFn` is the `_token_url` builder —
  the one door that lets the fetch/push ladder run against LOCAL BARE REPOS in the
  tests (the default builds the one-time https token URL; the token NEVER lands in
  .git/config). The T-47 instance-seam law.
- **The base-dict freeze quirk preserved**: `_publish` builds the result's
  repo_name/repo_url base BEFORE `_ensure_repo` runs — a pushed result still
  reports the RAW name on the first publish (owner-qualified only from the second
  publish of the same instance, when `_repo_full` is already cached). Ported
  verbatim (the `baseRepoName` snapshot).
- **The configparser read converges to a minimal INI scan**: `[remote "origin"]`
  → `url = …` with full-line #/; comments skipped; the reference's except-door
  (malformed file → None) maps to the lenient scan's "" (a malformed config never
  carries a findable origin in practice).
- **The sampler's walk order**: CPython's `os.walk` yields filenames in readdir
  (filesystem) order; Go's `ReadDir` sorts by name. The diversification law is
  order-insensitive (the round-robin takes one file per group per round either
  way); the goldens are generated from trees written in sorted name order so both
  sides agree.
- **The never-raises wrappers are unreachable in Go** (the T-46 convergence):
  `publish()`'s try/except and the radar's per-slot except doors have no
  panicking analog in the ported code — EXCEPT where the reference's own tests
  drive raise-doors (BoomAPI): those map to panicking fakes recovered per slot
  and SPOKEN in the detail (the T-47 raise-door law), verified against the
  golden.
- **Detection is a dict**: the Go `Detection` embeds `*storage.OrderedMap` (the
  house dict) with named accessors — the worker payload shape IS the dict; plain
  `map[string]any` results flow through `apply_detections`/`report_lines`
  exactly like the reference's plain-dict door.
- **The push-failure ok law**: an error set ⇒ `ok == false` even with the kept
  local commit (the CLI exits 1; `describe` speaks "(commit kept locally)").
  The unconfigured-name skip, by contrast, is ok=true (exit 0).
- **Text-mode reads**: `_parse_note`'s `read_text(errors="replace")` rides
  `helpers.PyNewlines(helpers.DecodeUTF8Replace(data))` before the splitlines
  parse — the T-41 universal-newlines law.
- **The Python-source contract tests have no Go analog** (the read-only-verb
  urllib inspection, the GUI wiring pins in test_reporadar §6): the read-only
  law is instead proven from the fake API's recorded call paths; the GUI wiring
  belongs to the deferred GUI worker shell task.
