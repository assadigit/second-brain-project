// Package workerclient ports cloud/cloudflare_sync.py (T-34, Phase 3)
// — the desktop↔Worker REST client: pairing, HMAC-signed polling and
// pushes, the two confirmation-gate round-trips (banish + scan), the
// backfill, health, and the URL normalizer that must byte-match the
// Worker's own.
//
// Spec: docs/port-blueprint/module-specs/spec-cloud-worker-client.md
// + spec-worker-contract.md (the Worker stays JS — ADR-0005; this is
// the client interop only).
//
// THE WIRE LAWS (every leg):
//   - THE 1010 LAW: Cloudflare's edge answers banned signatures
//     (Python-urllib/3.x, Go-http-client/2.0) with 403 error 1010
//     BEFORE the Worker runs — every request wears
//     "GitCurator/0.91.0 (+https://github.com/assadigit/GitCurator)"
//     (kept in lockstep with the reference's VERSION at release).
//   - PROXY BYPASS: every request builds a no-proxy opener
//     (ProxyHandler({}) in the reference; Transport.Proxy == nil here)
//     — workers.dev was intercepted by local v2rayN into 403s.
//   - Timeouts: 30 s all legs; health 10 s.
//   - MakeRequest NEVER raises: network failure = status 0 +
//     {'error': ...}; the success law is status==200 AND
//     data['success'].
//
// THE HMAC LAW (v30's fix): the Worker stores sha256hex(shared_secret)
// and verifies with hmacSha256(secret_hash, message) — so the HMAC KEY
// is the sha256 HEX STRING's UTF-8 bytes, not the raw secret:
//
//	timestamp  = str(int(time.time() * 1000))   // ms epoch decimal
//	body_hash  = sha256hex(body_utf8)           // '{}' when no body
//	message    = METHOD\npath_with_query\ntimestamp\nbody_hash
//	key        = sha256hex(shared_secret) as UTF-8 bytes
//	signature  = HMAC-SHA256(key, message), lowercase hex
//
// Worker side: ±5-min window; recomputed over pathname + search;
// constant-time compare; 401 reasons "Missing auth headers" |
// "Timestamp out of range" | "Unknown install" | "Invalid signature".
//
// THE HONEST POLL (v0.86.0): get_pending_result distinguishes the
// three truths a display must never merge — ('ok', items) the live
// queue, ('disabled', []) not configured, ('error', []) attempted and
// failed. Never a silent empty list pretending to be a calm queue.
//
// DOCUMENTED DIVERGENCES (each per the port law):
//   - The reference prints diagnostics to stdout
//     ("[CloudflareSync] ... failed: ..."); the port routes them
//     through the DebugLog hook (default: silent) — the wire behavior
//     is the contract, not the console.
//   - should_poll's interval read degrades on garbage values (the
//     tolerant int with the 300 default) where Python's float>=str
//     comparison would raise TypeError — a hostile-config corner no
//     reference test pins.
//   - normalize_url's twitter.com→x.com replaces EVERY occurrence
//     (Python str.replace) where the Worker's JS .replace(string)
//     replaces only the first — a pre-existing reference↔Worker
//     divergence on path-embedded twitter.com spellings, preserved
//     byte-for-byte from the reference (the port's source of truth).
//   - build_decommission_events: the reference calls .get() on the
//     (url, reason) TUPLE rows get_all_decommissioned returns — an
//     AttributeError on real data its caller's try/except swallows
//     (a latent reference bug). The Go port cannot crash here; it
//     maps the settled []FailedRow{URL, Error} shape to the event
//     body directly (the clear intent), documented per reality-wins.
//   - The uuid4 for sync_id is hand-rolled (crypto/rand + the v4
//     version/variant bits) — no dependency change (the authorized
//     set is closed).
//   - push_backup_status merges an OrderedMap's keys in order (the
//     **backup_data law); a Go map[string]any merges in sorted key
//     order (Go maps are unordered — deterministic wire bytes).
package workerclient
