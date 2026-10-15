# Module Spec — Worker contract (app/cloudflare-bot — NOT migrated; client interop only)

- **JS LOC**: 8,243 (`src/`) · **Status**: contract-only (ADR-0005). The Worker stays live
  (0.40.0 expected by the desktop, `constants.py:313`).
- **Purpose of this spec**: everything the Go desktop client must honor. Storage model included
  for semantic understanding (the desktop never touches D1 directly).

## 1. Surface & routing

`/webhook` (POST, Telegram) · `/health` + `/` (GET) · `/dashboard` (HTML) · `/auth` (302) ·
`/api/*` (→ api.js) · else JSON 404. CORS `*` on JSON responses (allowed headers include the
three `X-Auth-*`).

## 2. Auth schemes

1. **Pairing** (no HMAC): `POST /api/pair {"code"}` — code from Telegram `/pair` (8-char
   uppercase, uuid prefix; KV `pair:<code>` TTL **600 s**, single-use). Success:
   `{"success":true,"install_id":"<uuid>","shared_secret":"<64-hex>"}`. Errors: 400 missing
   code; 401 invalid/expired.
2. **HMAC** (all desktop endpoints): headers `X-Auth-Install`, `X-Auth-Timestamp` (unix ms),
   `X-Auth-Signature` (hex). Message `METHOD\npath_with_query\nts\nsha256hex(body)`; **key =
   sha256hex(shared_secret) UTF-8 bytes**; ±5 min window; constant-time compare; updates
   `last_seen_at`. Errors 401: `Missing auth headers | Timestamp out of range | Unknown install
   | Invalid signature`.
3. **Bearer** (automation): `Authorization: Bearer DIRECTORY_SYNC_TOKEN` on `/api/directory/*`
   and `/api/commands*` (constant-time).
4. **1010 UA law**: Cloudflare's edge bans `Python-urllib/3.x` (and by extension any
   non-browser default) with 403 error 1010 **before the Worker runs**. Every desktop request
   wears `GitCurator/<version> (+https://github.com/assadigit/GitCurator)`.

## 3. Desktop endpoints (HMAC unless noted)

| Endpoint | Method | Request → Response |
|---|---|---|
| `/api/pending` | GET | → `{success, pending:[{url_normalized, url_original, url_type, github_owner, github_repo, github_metadata{stars,description,language,topics,readme_excerpt}\|null, first_seen_at, forward_count, ledger_id}], count}` — ledger LEFT JOIN vault_mirror, `forgotten=0`, mirror status NULL/'pending', not decommissioned, not blocked-domain dead letter, ORDER BY first_seen_at ASC, **LIMIT 100**; side effect sets `sync_state.desktop_last_poll` |
| `/api/verify` | GET | → `{success, stats{total_ledger,in_vault,pending,decommissioned,dead_letters,non_github}, ledger_hash, vault_mirror_hash}` (hashes = sha256 of `GROUP_CONCAT(url_normalized,'\|')`) |
| `/api/decommissions?since=` | GET | → `{success, decommissions, count}` — only `source='bot'` rows, LIMIT 200 |
| `/api/vault_mirror/push` | POST | `{install_id, sync_id, vault_index_hash, entries[{url_normalized, vault_path, category, title, status, first_processed_at, last_updated_at}], decommission_events[...]}` → `{success, synced_entries, synced_decomm, conflicts:[], vault_mirror_total, idempotent?}` — **idempotent via sync_state `sync_processed:<sync_id>`**; this is how items leave the pending queue |
| `/api/decommission` | POST | `{install_id, url_normalized, reason, details}` → `{success, decommissioned:true}` |
| `/api/errors` | POST | `{install_id, errors[{severity, error_code, message, details, occurred_at}]}` → `{success, received, dm_sent}` — CRITICAL → immediate Telegram DM |
| `/api/gdrive/backup_status` | POST | snapshot fields → `{success}` |
| `/api/backfill` | POST | `{install_id, entries[{url_normalized, url_original, url_type, github_owner, github_repo, first_seen_at}], chunk_num, total_chunks}` → `{success, inserted, skipped, progress}` — idempotent per URL |
| `/api/banish/propose` | POST | `{install_id, count, items[{url,title,marker,door}], timeout_s}` → `{success, id, message_id}` — sends the Telegram ask; state `banish_confirm:<id>`; supersedes previous pending ask; 502 if Telegram send fails |
| `/api/banish/status?id=` | GET | → `{success, id, status: pending\|confirmed\|declined\|timeout\|superseded\|done, count}` (404 unknown) |
| `/api/banish/result` | POST | `{install_id, id, outcome, deleted}` → `{success, status}` — edits the Telegram message into the closing line |
| `/api/scan/propose` `/status?id=` `/result` | POST/GET/POST | twin trio with `{deletions, moves, new_folders, items[{kind,title,detail,marker}], summary≤400, timeout_s}` / `{id}` / `{id, outcome: applied\|timeout, applied, moved, folders}` |
| `/health` | GET | `{status:'ok', service:'github-curator-bot', version, last_webhook_at, timestamp}` |

## 4. Directory/automation endpoints (Bearer)

- **`POST /api/directory/sync`**: full index (schema `gitcurator-directory/1`) ≤ **20 MB**
  (`MAX_SYNC_BYTES`, 413 over) and ≤ **5,000 sites**; all-or-nothing validation + filtering
  (private categories, active tombstones, banished hosts, invalid sites, dedup) BEFORE the
  single-row store → `{success, published_at, stats{received, stored, filtered_private,
  filtered_tombstone, filtered_banished, filtered_invalid}}`.
- **`GET /api/commands?status=pending`** (Bearer): `{schema:'gitcurator-commands/1',
  commands[{command_id, type, site_id, slug, host, payload, created_at}]}` — oldest-first;
  **held commands (inside the 10 s undo window) invisible**.
- **`POST /api/commands/<urlquoted-id>/ack`** (Bearer): `{status:'applied'\|'failed',
  error?≤500}` (≤64 KB body); idempotent re-ack; tombstone cleared only on applied
  delete/banish.
- **`GET /api/directory/index`** (public): stored index + pending-command overlay;
  `note_path` stripped; `ETag "<content_hash>-<commandsFingerprint>"`, `Cache-Control: public,
  max-age=120`, CORS `*`; 304 on If-None-Match; **`404 {"error":"index not published yet"}` when
  never published** (the desktop's read-back door).

## 5. Public directory endpoints (read-back only)

`GET /api/directory/ratings` (aggregates; ETag/304) · `POST /api/directory/ratings {slug,
stars 1-5, voter 8-64 [A-Za-z0-9_-]}` (upsert + aggregate; unknown slug 404; **soft per-IP
rate limit 40 writes/10 min** via KV) · `GET /api/directory/link-status` (suspect/dead rows;
max-age 30).

## 6. Storage model (D1 `curator-bot`; KV `CACHE`; Queues)

16 tables (full DDL `app/cloudflare-bot/schema.sql`): `ever_seen_ledger` (immutable intake
ledger + GitHub enrichment), `vault_mirror` (desktop-owned status; PK url_normalized),
`decommission_events` (append-only), `dead_letters` (reason invalid_format/persistent_fetch_fail/
blocked_domain/dlq_exhausted), `desktop_errors`, `sync_state` (kv: desktop_last_sync/poll,
vault_index_hash, backfill_progress, `sync_processed:<id>`, gate states, admin lockout),
`gdrive_snapshots`, `dashboard_sessions`, `desktop_installs` (install_id PK, shared_secret_hash
= sha256 hex, paired_at/last_seen_at, user_agent), `activity_log`, `directory_index` (single row
`id=1` CHECK, payload, published_at, source, content_hash, site/category counts),
`directory_tombstones` (no-resurrection guard; `cleared_at NULL` = active),
`directory_ratings`, `admin_commands` (queue with status/attempts/held_until),
`admin_sessions` (token_hash + csrf), `link_checks`.

KV families: `dedup:<url>` (TTL 3600), `pair:<code>` (TTL 600), ratings rate-limit keys.
Queues: `curator-ingest` (batch 10 / timeout 5 / retries 5 → DLQ) and `curator-ingest-dlq`
(retries 1; DLQ consumer records `dlq_exhausted` dead letters — which do NOT block re-sends;
only `blocked_domain` blocks intake).

Cron (active): `*/20 * * * *` — directory GitHub fallback rebuild (from
`DIRECTORY_GITHUB_REPO`) + dead-link sweep (10 stalest). Secrets: BOT_TOKEN, GITHUB_PAT,
ALLOWED_USER_IDS, DIRECTORY_SYNC_TOKEN, WEBHOOK_SECRET (optional), ADMIN_PASSWORD_HASH.
`HMAC_SECRET` is vestigial (never read by code).

## 7. Queue semantics (intake path)

webhook → allowlist → `extractUrls` (http(s) anywhere + whole-line bare addresses) → ONE queue
message (`{type:'urls', urls, chat_id, user_id, message_id, text, received_at}`) → consumer:
canonicalize (github.io mapped) → self-domain skipped (never stored) → **banned-domain LAW**
(403 reply for GitHub-likes; dead letter `blocked_domain` for websites) → decommissioned check
→ vault_mirror `in_vault` → "already in vault" reply → ledger exists → bump forward_count
("already pending") → else INSERT ledger + KV dedup + activity → categorized reply (single-URL
detailed with 🗑️ inline decommission button; multi-URL summary). New GitHub URLs → `enrich`
message → GitHub API with GITHUB_PAT (404 → decommission_events '404'; 200 → stars/desc/lang/
topics/README-500-chars + reply edit). Webhook always returns 200 (Telegram contract).

## 8. Go client obligations (summary)

1. Implement pairing + HMAC exactly (§2) incl. the hex-string-as-key detail and ms timestamps.
2. Wear the 1010 UA on every leg; bypass proxies for workers.dev.
3. `/api/pending` LIMIT 100 — pagination-free by design; treat count as the truth.
4. Mirror-push is the ack (idempotent by sync_id).
5. Never treat a down Worker as an empty queue (None ≠ []).
6. Publish door: Bearer; index ≤20 MB/≤5k sites; ack commands only after publish.
7. Read-back: public index 404 = never-published (honest state, not an error).
