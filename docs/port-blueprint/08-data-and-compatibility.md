# 08 — Data, Configuration, and Compatibility

> Every format, file, protocol, and persistence contract the Go app must preserve. Sources: module
> specs + direct reads. This is the "do not break the owner's data" document.

## 1. Configuration

### 1.1 `app/config.json`

- 109 keys; defaults = `CONFIG_EXAMPLE` (`gitcurator/constants.py:122-305`, template
  `app/config.example.json`). Full key list with types/defaults: `spec-core-constants.md` §2.
- **Load**: deep-merge defaults under the user file (missing keys defaulted; unknown keys kept).
- **Save**: merge-never-rebuild — nested dicts merged key-by-key, unknown keys survive, tuning keys
  (`timeout_per_repo`/`max_retries`/`delay_between_api_calls`) only defaulted when absent
  (`gui/main_window/vaults_config.py:260-281`); written atomically as UTF-8 JSON, 2-space indent,
  **non-ASCII unescaped**.
- Go: `map[string]any` tree + typed accessors; `json.Encoder` with `SetEscapeHTML(false)` and
  `SetIndent("", "  ")`.

### 1.2 `Documents/.gitcurator` (user memory)

- JSON; marker `"_gitcurator_user_config": 1`; chmod 600 best-effort.
- Key families: 15 flat credential keys + 3 nested (`proxy`, `ollama`, `models`) + credential
  prefixes (`cloudflare_`, `gdrive_`, `hmac_`) + 6 flat vault keys + 2 nested (`vaultseal`,
  `goodrepos`) — exact lists in `spec-core-user-config.md`.
- **Unset grammar** (fill side): None/blank/`YOUR…_HERE`/**numeric 0**/empty list ⇒ unset;
  booleans never unset.
- Location: env `GITCURATOR_USER_CONFIG` (parent created) → Windows Known-Folder Documents
  (GUID `{FDD39AD0-238F-46AF-ADB4-6C85480369C7}`, OneDrive-aware) → `~/Documents` only if it exists
  → feature sleeps. Port with `golang.org/x/sys/windows`.

### 1.3 Environment variables actually read

| Var | Where | Effect |
|---|---|---|
| `GITCURATOR_USER_CONFIG` | user_config | user-memory path override |
| `HTTP_PROXY`/`HTTPS_PROXY` (implicit) | urllib default openers | honored on non-loopback remote legs (llm_client preflight, GitHub checks) — Go must **opt out** where Python opted out (loopback/Worker legs) |
| `DIRECTORY_WORKER_URL`, `DIRECTORY_SYNC_TOKEN`, `DIRECTORY_VAULT` | `cloud/directory_sync.py:1015-1028` | env overrides for the publish leg (tests scrub them) |
| `NO_COLOR` | `cli/cli_terminal.py` | CLI color law |
| `GITHUB_TOKEN` | vaultseal standalone CLI default | tools only |

## 2. Local data files

| File/Dir | Owner | Format notes |
|---|---|---|
| `app/cache.db` | 7 SQLite stores (doc 02 §5) | WAL; exact DDL in specs; naive-local text dates except `sync_deferrals` (UTC) |
| `app/sync_review.db` | DeferredLog | own file, WAL, never-delete rows |
| `app/error_outbox.db` | ErrorReporter | severity routing table |
| `app/session.session` | Telethon | **accepted break** — re-login (ADR-0004) |
| `app/chrome-profile/` | dedicated Chrome profile | persists; kept |
| `app/chrome-proxy-memory.json` | route memory `{domain: direct|proxy}` | atomic writes, indent 2, non-ASCII raw |
| `app/app.lock` | single-instance | PID file; Go implements same name+semantics |
| `app/reports/` | run reports, dry-run reports, dist zips, recall reports | naming patterns in specs |
| `app/logs/processing_<ts>.md/.txt` | rotating logs 5 MB×5 | keep format |
| `<vault>/links_manifest.json` | LinkTracker | atomic; dry-run recorded not written |
| `<vault>/_review/hand-delivered/queue.json` | hand-delivery | shape in spec; corrupt → empty |
| `<vault>/_review/DECOMMISSIONED.md` | master table | exact header/row grammar (7 cells) |
| `<vault>/_inbox/<platform>_links.md` | platform intake tables | exact table grammar |
| `<vault>/000 📚 Website Directory.md` | directory note | exact template |
| `<vault>/_system/{directory-config.json, banished-sites.md, search-synonyms.md}` | directory sync config trio | created if missing, never overwritten |
| `<vault>/.gitignore` managed blocks | VaultSeal | two managed blocks (machine state + local-only), rebuilt not appended |
| `Documents/.gitcurator` | see §1.2 | — |
| `<vault>/Library/` mirror | mirror.py | `mirror_of:` markers, read-only banners |

## 3. Vault content contracts (markdown)

- **Note identity**: normalized `source:` URL (GitHub: `normalize_url`; websites:
  `normalize_website_url`) — the ONE SPELLING laws.
- **Frontmatter**: GitHub notes per `spec-core-note-builder.md` §4 (key order, quoting, caps);
  websites notes per `spec-core-website-pipeline.md` §note-format (incl. `site_links:` line,
  `managed_by: gitcurator`, `schema_version: "1"`, `prompt_version: web-v2|gh-v1`).
- **Ownership markers**: `managed_by`, OWNERSHIP_BANNER callout; `_missing/` placeholders (delete =
  re-check trigger); locked notes (`locked`/`category_locked`/lock tags); curated stamp.
- **App-managed blocks**: recall block `<!-- gitcurator:recall:start/end -->`; related block
  `<!-- gitcurator:related:start/end -->` — fingerprint-invisible, idempotent surgery.
- The Go writer must be byte-compatible: same key order, same quoting, same section headings, same
  footers (delete-hint block, source footer), same CRLF preservation rules (dominant-line-ending
  detection).

## 4. SQLite schemas (compatibility-critical DDL)

Pinned verbatim in the module specs (source of truth): `spec-gui-cache-db.md`,
`spec-core-note-state.md`, `spec-core-website-state.md`, `spec-core-note-reads.md`,
`spec-core-linking.md`, `spec-core-sync-review.md`, `spec-integrations-misc.md` (error_outbox).
Rules for the Go port:

- Same table/column names and constraints; new indexes allowed only if behavior-neutral.
- Timestamp formats preserved per store (naive-local vs UTC — see mapping §3 item 5).
- `note_reads.frontmatter_json` carries the parse-version marker `v` (currently **4**) and the
  exact key set — bump semantics must be replicated (old rows re-read once).
- `INSERT OR REPLACE` / `ON CONFLICT … DO UPDATE` semantics per store as documented.
- Migration: none. The Go app opens the existing file. A side-by-installation runs the same lock
  file so both apps never run simultaneously.

## 5. External API contracts

### 5.1 Cloudflare Worker (desktop↔bot)

- Base: `config.cloudflare_worker_url`; **every** request wears
  `User-Agent: GitCurator/<version> (+https://github.com/assadigit/GitCurator)` (1010 law) and
  bypasses proxies.
- **Pairing**: `POST /api/pair {code}` → `{install_id, shared_secret}` (code from Telegram `/pair`,
  10-min TTL, single-use).
- **HMAC** on all `/api/*` desktop endpoints: headers `X-Auth-Install`, `X-Auth-Timestamp` (ms),
  `X-Auth-Signature`; message = `METHOD\npath_with_query\nts\nsha256hex(body)`; **key =
  sha256hex(shared_secret) as UTF-8 bytes**; ±5 min window; constant-time compare server-side.
- Endpoints: `/api/pending` (GET), `/api/vault_mirror/push` (POST — how items leave the queue;
  idempotent by `sync_id`), `/api/verify`, `/api/decommission`, `/api/decommissions?since=`,
  `/api/errors`, `/api/gdrive/backup_status`, `/api/backfill`, `/api/banish/{propose,status,result}`,
  `/api/scan/{propose,status,result}`, `/health`. Full request/response shapes:
  `module-specs/spec-worker-contract.md`.
- **Directory (Bearer)**: `POST /api/directory/sync` (full index ≤20 MB/≤5,000 sites, all-or-nothing
  validation), `GET /api/commands?status=pending`, `POST /api/commands/<id>/ack`,
  public `GET /api/directory/index` (404 = never published).
- Error law: 200 + `success:true` ⇒ success; anything else is the honest error string; network
  failure = status 0; **never raises**.

### 5.2 GitHub REST

- Desktop paths: PyGithub `get_repo` (metadata/README/topics/languages), `GET /user`,
  `GET /repos/{owner}/{repo}`, opt-in `POST /user/repos {private:true}`;
  banner images from `https://opengraph.githubassets.com/1/{owner}/{repo}`.
- Headers: `Authorization: Bearer`, `Accept: application/vnd.github+json`,
  `X-GitHub-Api-Version: 2022-11-28` (goodrepos/radar), UA per context
  (`GitCurator-GoodRepos`, `GitCurator-RepoRadar`, `GitCurator-connection-check`).
- Rate-limit law: pause under 50 remaining until exact reset; 403-rate → 3,605 s wait; 401 →
  anonymous fallback (GitHub pipeline only).

### 5.3 LLM endpoints

- Ollama: `POST /api/chat` (SDK-shaped; `format:'json'`, `options.num_ctx/num_predict`),
  `GET /api/tags`; UA `GitCurator-detect` on raw probes.
- OpenAI-compatible: `POST {base}/chat/completions`, `GET {base}/models` (base must include
  `/v1`); JSON-mode with one memoized clean retry on 400.
- Anthropic: `POST {base}/messages` + `GET {base}/models`; headers `anthropic-version:
  2023-06-01`, `x-api-key`; `max_tokens` required (4096 fallback); system-prompt JSON instruction.
- llama.cpp: `GET /props`, `/health`, `/v1/models` (loopback, always direct, optional Bearer).

### 5.4 Telegram MTProto (via gotd in Go)

Bot-chat iteration (all messages oldest-first), `min_id` support, read-acknowledge on the bot
peer; Saved-Messages ranges/offset/keyword/single-id; code+2FA auth; SOCKS with pre-flight;
honest errors. Session: new (accepted break).

### 5.5 Auxiliary web endpoints

`cloudflare-dns.com/dns-query` (DoH JSON), `archive.org/wayback/available` +
`web.archive.org` (rescue), Chrome DevTools on `127.0.0.1:<port>` (HTTP + WebSocket).

## 6. User-Agent inventory (exact strings — a compatibility table)

| Context | UA | Source |
|---|---|---|
| Worker (all legs) | `GitCurator/<version> (+https://github.com/assadigit/GitCurator)` | `cloudflare_sync.py:59` |
| Web fetch | Chrome 131 full string (config `web_user_agent` override) | `web_fetch.py:136-138` |
| GitHub (pipeline) | PyGithub default `PyGithub/2.x` — **[verify]** whether to keep or set `GitCurator` (behavior-neutral; recommend explicit `GitCurator/<version>`) | — |
| goodrepos / radar | `GitCurator-GoodRepos` / `GitCurator-RepoRadar` | `goodrepos.py`, `repo_radar.py` |
| Connection battery | `GitCurator-connection-check` | `connection_check.py:167,417,655` |
| Sentinel | `GitCurator-sentinel` | `net_guard.py:62` |
| Ollama detect | `GitCurator-detect` | `llm_client.py:1334` |

## 7. Import/export formats

- **Import**: `.txt`/`.md` link lists — grammar in `spec-core-links.md` §7 (markdown links, bare
  domains incl. `127.0.0.1:port`, comment lines, duplicates counted, unparsed listed).
- **Export**: `export_all_bot_links` (all-links report), verification reports (markdown tables),
  backup zips (exclusions + FIFO + `GitCurator-data/sync_review.db` inside every zip — v0.86.0),
  GoodRepos README + staging tree, `directory.json` (schema `gitcurator-directory/1`).
- **Undo**: `<vault>/_undo_last_batch.txt` diff list consumed by `undo_last_batch`.

## 8. Migration and backward-compatibility strategy

1. **No data migration** — same files, same schemas (ADR-0008). The Go app is a *reader/writer in
   place*.
2. **Side-by-side operation** guarded by the same `app.lock` (mutual exclusion between Python and
   Go instances).
3. **Dry-run parity**: the dry-run recorder + shadow cache behave identically, so the owner can
   rehearse the Go app against real vaults with zero writes before trusting it.
4. **Rollback**: the Python app remains installed and functional; the Go app is additive until
   acceptance (doc 15) is signed off.
5. **Testing with old fixtures**: golden files (`tests/golden/websites.json`, fixtures/web/*.html,
   a sanitized copy of a real `cache.db` and `config.json`) are the differential corpus (doc 09
   §5).
