# Module Spec — `cloud/cloudflare_sync.py` (the Worker REST client)

- **LOC**: 735 · **Go home**: `internal/workerclient` · **Status**: port
- **Responsibility**: the desktop↔Worker REST client: pairing, HMAC-signed polling/pushes, the
  two confirmation-gate round-trips (delegated to banish/scan confirm), backfill, health,
  vault-mirror pushes.

## Wire law (all legs)

- **UA (1010 law)**: `HTTP_USER_AGENT = "GitCurator/0.91.0 (+https://github.com/assadigit/
  GitCurator)"` (`cloudflare_sync.py:59`) — kept in lockstep with VERSION at release. Cloudflare's
  edge answers `Python-urllib/3.x` with **403 error 1010** before the Worker runs. Go's default
  UA (`Go-http-client/2.0`) is equally banned — the UA is set explicitly on every client.
- **Proxy bypass**: every request builds an opener with `ProxyHandler({})` (workers.dev was
  intercepted by local v2rayN → 403s) (`:142-143, 250-253, 636-637`).
- Timeouts: 30 s all legs; health 10 s. `is_enabled()` = master switch + URL + install_id +
  secret; `is_paired()`.
- `_make_request` never raises: network failure = status 0 `{'error': str(e)}`; success law =
  `status==200 and data.get('success')`.

## HMAC (189-231; Worker side `api.js:249-296`)

```
timestamp  = str(int(time.time() * 1000))          # ms epoch decimal string
body_hash  = sha256hex(body_utf8)                  # '{}' when no body
message    = f"{METHOD}\n{path_with_query}\n{timestamp}\n{body_hash}"
key        = sha256hex(shared_secret).encode('utf-8')   # hex STRING's bytes!
signature  = HMAC-SHA256(key, message) lowercase hex
headers    = X-Auth-Install, X-Auth-Timestamp, X-Auth-Signature,
             Content-Type: application/json, User-Agent: <1010 UA>
```

Worker: ±5-min window; recomputes over `pathname + search`; constant-time compare; errors 401
`Missing auth headers | Timestamp out of range | Unknown install | Invalid signature`.

## Endpoint methods (client side)

`pair(code)` (118-177; 403 → redeploy/wrong-URL advice; 401 → "run /pair again … within 10
minutes") · `get_pending_result()` (271-296; the honest poll: `('ok', items)` /
`('disabled', [])` / `('error', [])`) · `get_pending()` (byte-compat delegate) ·
`should_poll()` (300 s interval) · `push_vault_mirror(entries, decommission_events,
vault_index_hash)` (315-350; sync_id idempotency) · `push_decommission` ·
`propose_banish/banish_status/report_banish_result` · `propose_scan/scan_status/
report_scan_result` · `push_errors` · `push_backup_status` · `verify` ·
`get_decommissions(since)` · `backfill_chunk` · `health_check` · `normalize_url` (648-680;
must byte-match the Worker's `normalizeUrl`: strip trailing `.,);`, twitter→x, drop
fragment/query, drop trailing `/`, lowercase domain only).

## Config keys (66-72)

`cloudflare_worker_url`, `cloudflare_install_id` (UUID), `cloudflare_shared_secret` (64-hex),
`cloudflare_enabled`, `cloudflare_poll_interval` (300 s). Pairing writes all three + enabled.

## Concurrency

One `threading.Lock` serializing polls/pushes; `_last_poll` for should_poll.

## Tests

**The loopback rig** (ported from `tests/test_synclive.py`): a real local HTTP server
implementing `/api/pending` HMAC verification byte-for-byte (incl. 1010-UA rejection and the
ms-window), `/api/directory/index` honest 404, `/api/commands`, and the Bearer publish door.
Client tests: signature vectors (golden HMAC fixtures), pair error mapping, honest-poll three
truths, mirror push idempotency, proxy-bypass assertion (fake env proxy ignored).
