# Module Spec — `cloud/directory_sync.py`

- **LOC**: 1,826 · **Go home**: `internal/directorysync` · **Status**: port
- **Responsibility**: the Website Directory publish leg + the Phase-2 admin-command cycle —
  the one-way content flow Obsidian → GitHub → Website.

## Index build (`build_index` 820-1008) — THE VAULT MIRROR LAW (v0.88.0)

- The published category tree **IS the vault's folder tree**: the `_system/directory-config.json`
  is a metadata registry only (ghost entries never publish; undeclared folders publish with
  prettified names); folder-name translation `safe_filename(name) ↔ display name` is
  case-insensitive; the private belt is folded
  (`prettify_folder_name(safe_filename(name)).lower()`).
- Filters: banished hosts (host-level, with the shared-host exception list github.io,
  notion.site, vercel.app, blogspot.com, wordpress.com, netlify.app, pages.dev, gitlab.io,
  webflow.io, medium.com, substack.com — URL-level bans only on shared hosts); private
  categories; invalid sites; dedup.
- Site record: `{id: sha256(normalized)[:16], slug, name, url, host, category, subcategory,
  summary, tags, featured, needs_review, added_at, updated_at, what_it_does, best_used_for,
  practical_details, watch_out, similar_tools, extra_sections, signup_required, pricing,
  detail_markdown}`.
- Index: `{schema: "gitcurator-directory/1", generated_at, counts, categories, sites,
  popular_tags (top 12), tag_aliases, synonyms, banished_hosts, new_days: 14}`.

## Note parser (`parse_note` 216-383)

Frontmatter + drawer-section parser (TL;DR / What it does / Best used for / Practical details /
Watch out / Similar tools + extra sections); `signup_required` chip; pricing enum
(free/freemium/paid/unknown); `featured`; `needs_review` (v0.39.0); banish-tags stripped.

## Config trio + belts

- `_system/{directory-config.json, banished-sites.md, search-synonyms.md}` in the vault;
  `ensure_system_files` creates when missing, never overwrites; shipped fallbacks at
  `app/taxonomy/`.
- `local_only_note_patterns` (478-495; the six tombstoned stems as globs) +
  `private_exclusion_patterns(vault)` (498-567; private-category folders **discovered on disk**
  derived from belt names (folded), the gitignore escape law, the vault's own
  `_system/directory-config.json` `local_only` list as the owner's extension point, never
  raising) — feeds **both** the publish filter and VaultSeal's second gitignore block.
- `is_banished(url, banished, shared_hosts)` (619-637).

## Publish + read-back

- `publish_index(index, worker_url, token, timeout=30)` (1031-1065): `POST
  {worker}/api/directory/sync`; body `json.dumps(index, ensure_ascii=False)` UTF-8; headers
  `Content-Type: application/json`, `Authorization: Bearer <token>`, 1010 UA.
- `fetch_published_index(config)` (1068-1135): `GET /api/directory/index` → four truths:
  `ok (sites count + generated_at)`, `never_published` (404), `not_configured`, `error`.
- `run_sync` / `publish_after_batch` (the end-of-batch hook = full command cycle).
- Env overrides `DIRECTORY_WORKER_URL/TOKEN/VAULT`; config `directory_worker_url` (wins over
  `cloudflare_worker_url`), `directory_sync_token`, `directory_poll_interval` 300 s.

## Admin command cycle (Phase 2 — golden rule 5 order)

`fetch_pending_commands` (`GET /api/commands?status=pending`, Bearer) → `apply_admin_commands`
(types `delete, banish, feature, unfeature, review, move, edit, unban` — each idempotent,
failure-isolated; vault appliers `_edit_note_fields`, `_move_note_file` (safe_filename folders
+ numeric suffix on collision), `_trash_note` (→ `<vault>/.trash/<stem>-<ts>.<ext>` — never a
permanent delete), `_banish_add/_banish_remove`) → `push_mirror_repo` (plain
`status→add -A→commit "admin commands applied: <ts>"→push` when the vault is a repo with a
remote) → publish → `ack_command` (`POST /api/commands/<urlquoted-id>/ack {status: applied|
failed, error ≤500}`) — **ack only after publish (no resurrection window)**; held commands
(10 s undo window) invisible server-side; a down Worker returns None (commands stay un-acked).

## CLI (`main` 1228-1292)

`python -m gitcurator.cloud.directory_sync` with `--vault`, `--config` (default `./config.json`),
`--sync-only`, `--publish`, `--print-index`, `--out FILE`, `--make-admin-hash` (PBKDF2 helper).

## Tests

Index golden (tree-from-disk law, ghost/undeclared cases, folded belt); parser field matrix;
publish payload shape (Bearer + non-ASCII body); read-back four truths; command cycle (order,
idempotent applies, ack-after-publish, held-window invisibility, trash-not-delete); trio
creation law; belts builders (port `test_localonly` + `test_directorysync` selections).

## Belt floor shipped early — T-47 (SESSION-022)

`internal/directorysync/belt.go` carries the v0.89 LOCAL-ONLY BELT's
pattern builders (local_only_note_patterns, private_exclusion_
patterns, load_directory_config + the pure helpers) because T-47's
vaultseal config bridges consume them — the T-45 repobanish-floor
precedent. The shipped directory-config default is EMBEDDED at
`assets/directorysync/website-directory-config.json` (sha256-pinned
by vaultseal's golden) where the reference reads app/taxonomy/ from
disk — the doc 04 §1 assets layout. T-49 completes the module
(build_index, publish, the command cycle, the config trio).
