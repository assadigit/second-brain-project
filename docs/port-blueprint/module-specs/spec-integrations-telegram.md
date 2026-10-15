# Module Spec — `integrations/telegram_fetch_worker.py` + `telethon_fetcher.py` + `subprocess_runner.py`

> **DORMANT (2026-10-15, ADR-0004 Am.3 — owner stamp "OPTION B CONFIRMED"):** the Go app
> drops the direct Telegram (MTProto) leg entirely; the bot queue rides the Cloudflare
> Worker HTTP API only, and read-acknowledge is dropped. This spec is **not implemented**
> on the port's critical path — it is preserved as the contract for the **Phase-8 optional
> revival** (capability proven by the T-01 PoC on the owner's account: ADR-0004 Am.2;
> `poc/telegram` + the gotd v0.110.0 pin remain in the repo as the dormant proof). The
> spec below is retained verbatim for that future use.

- **LOC**: 741 + 432 + 456 · **Go home**: `internal/telegram` (gotd/td — ADR-0004) ·
  **Status**: **replace** (the *contract* below is preserved; the subprocess mechanism is not)
- **Responsibility**: the Telethon fetch engine (subprocess) + its parent supervisor. In Go, a
  gotd client package implements the same surface natively.

## The contract the Go client must implement (from `telegram_fetch_worker.py:`)

**Modes** (mutually exclusive, first-match): `keyword_search` (328-426) · `bot_queue` (433-600) ·
`preview_only` (602-631) · range/offset fetch (632-698).

- **bot_queue mode**: iterate ALL messages in the bot chat (no limit), **reversed to
  oldest-first** (:501); `min_id > 0` fetches only newer (:474-480); heartbeat every 500 messages
  keeps the parent's idle watchdog alive; `mark_read` → `send_read_acknowledge` (:572-577);
  result `{success, urls (github), non_github_urls, total_messages, bot_queue: true,
  max_message_id, raw_url_count, duplicates_removed, min_id_used}`; `max_message_id` falls back
  to `min_id` when nothing new.
- **Result shapes** (keyword): `{success, keyword_search: true, start_id, end_id,
  start_preview, end_preview, searched_count, match_count}`; (preview): `{success, preview:
  {first: […], last: […], total_count}}`; (range): `{success, urls, non_github_urls,
  total_messages}`; failure always `{success: false, error: "TypeName: message"}`.
- **Link extraction**: `core.links` grammar; GitHub links inside generic links promoted to
  canonical form; dedupe via sets; trailing `'.,);'` stripped.
- **Connection law**: SOCKS proxy (default 127.0.0.1:10808) with **pre-flight TCP probe (2 s)**
  BEFORE touching Telegram; a *disabled* proxy is warned, never silently bypassed (censorship
  law); 3 connect attempts (90 s each, 3 s/6 s sleeps); **DC rotation only for fresh sessions**
  (`_ALT_DCS = [(4,'149.154.167.91',443), (1,'149.154.175.53',443)]` — auth keys are DC-bound);
  retries deliberately low on the client (the supervisor owns deadlines); server-side session
  validity checked with `get_me()`; stale session deleted + fresh auth.
- **Auth**: `__NEED_CODE__` / `__NEED_PASSWORD__` markers answered by one stdin line each (in
  Go: the ask-code channel with the same 300 s window).

## The supervision law (from `subprocess_runner.py:`)

- Idle deadline **180 s** (any output resets; the child's heartbeat exists precisely for this);
  **auth grace 360 s** (extends idle while a human types); **hard cap 1,800 s**; kill reasons
  verbatim ("hard cap of Ns exceeded", "no worker output for Ns (network/proxy stall?)", "auth
  needed but no callback wired", "auth cancelled by user", "worker died during auth/2FA",
  "stderr closed but process did not exit"); kill + `wait(10)`; result carries `stderr_tail`
  (last 10 lines); registry + `kill_all_workers()` at shutdown.
- In Go: the same deadlines enforced by a supervisor goroutine around the gotd session; the
  **35-minute GUI watchdog** stays (window.py law).

## telethon_fetcher.py (legacy in-process path)

Only reachable via CLI `--single-id` (telegram_lazy). **Collapsed into the Go client** (one
engine; no dual path). Its forensic docstring (WinError-121 lessons) informs the gotd error
mapping.

## Go acceptance tests

Fake-peer queue fetch (100 messages, oldest-first, min_id respected); heartbeat cadence; kill
reasons on idle/hard-cap/auth-cancel; mark-read called only on success; error strings verbatim;
proxy pre-flight honored; result JSON shapes byte-equal to the Python subprocess contract (the
parent-side code reads these fields).
