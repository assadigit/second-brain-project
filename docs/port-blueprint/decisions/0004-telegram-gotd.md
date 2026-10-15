# ADR-0004 — Telegram access in Go: replace Telethon with gotd/td

- **Status**: accepted (primary), provisional on a login-flow PoC
- **Date**: blueprint session (2026-10)

## Context

Telegram access in the Python app is **Telethon (MTProto user client)**, run as an **isolated
subprocess** (`app/gitcurator/integrations/telegram_fetch_worker.py`), supervised by
`subprocess_runner.py`:

- Modes: bot-queue iteration (all messages in the bot chat, oldest-first, `min_id` support,
  `mark_read` acknowledge via `send_read_acknowledge`), Saved-Messages ranges/offsets, keyword
  search, single-id, preview. JSON-line protocol on stdin/stdout with `__NEED_CODE__` /
  `__NEED_PASSWORD__` auth markers on stderr (`telegram_fetch_worker.py:43-156, 704-741`).
- SOCKS proxy (default `127.0.0.1:10808`) with a pre-flight TCP probe; **never silent direct
  fallback** (censorship law — a direct connection leaks Telegram usage to the ISP);
  DC rotation only for fresh sessions (`telegram_fetch_worker.py:92-95`).
- Session file `app/session.session` (Telethon SQLite format), auth flow with code + 2FA password.
- The Worker HTTP API (`/api/pending`, `/api/vault_mirror/push`) is the *other*, newer queue path;
  both are live today (queue check rides Telethon; the Sync display's `get_pending_result` and the
  banish/scan confirms ride the Worker).

## Problem

Pick the Go replacement for Telethon's MTProto client functionality. There is no drop-in "Telethon
for Go".

## Options considered

| Option | Assessment |
|---|---|
| **A. `github.com/gotd/td`** — pure-Go MTProto client, actively maintained, full auth/sessions/messages API | Primary candidate. Covers: sign-in with code + 2FA, message iteration, read acknowledge, SOCKS5 dialing, session storage pluggable |
| **B. Worker-API-only architecture** — drop MTProto entirely; fetch the queue only from `/api/pending`, mark-read via a new Worker endpoint | Cleanest design, but changes observable behavior (Saved-Messages ranges, keyword search, single-id CLI modes and the live Telegram leg of Test Connection all need MTProto), and requires Worker changes — violating "preserve workflows" in the first migration phase |
| **C. Keep the Python Telethon subprocess bundled with the Go exe** | Reintroduces a Python runtime; defeats the packaging goal; rejected |
| **D. Bot API only** | Cannot read *the owner's* bot chat history or Saved Messages with a user account; cannot do `send_read_acknowledge` for the user; rejected |

## Decision

1. **Primary: `gotd/td`** inside a dedicated `telegram` Go package implementing the same surface the
   Python subprocess exposed (queue fetch with `min_id`, mark-read, ranges, keyword search, single-id,
   preview, login flow with code/2FA callbacks, SOCKS proxy with pre-flight, honest errors).
2. **Keep the subprocess isolation *pattern*** — run the MTProto client on a dedicated goroutine with
   a context; the app-level contract (never two concurrent fetches — the TelegramLockManager law,
   `gui/telegram_lock.py`) is preserved by the same single-slot lock in Go.
3. **Session migration**: Telethon `.session` files are NOT portable to gotd. The owner performs a
   **one-time re-login** on first Go launch (config carries the same api_id/api_hash/phone). The
   Python app remains functional during transition. Documented as an accepted, visible change
   (once-per-migration, not per-run).
4. **Fallback / contingency**: if the gotd PoC cannot complete interactive login or bot-chat
   iteration on the owner's account within the PoC timebox, fall back to **Option B** (Worker-only
   queue + a small Worker addition for read-acknowledge) as a documented behavior change — decision
   then recorded as a new ADR superseding this one.
5. **Worker HTTP paths are ported regardless** (`cloudflare_sync` client: HMAC, 1010 UA law, pairing,
   pending, mirror push, banish/scan confirms) — they are independent of MTProto.

## Consequences

- One new dependency (gotd/td + deps). Pure Go; no CGO.
- The auth UX stays the same shape (modal asking for the code / 2FA password), timeout windows kept
  (300 s).
- `resend`, retry, DC-rotation subtleties are gotd's responsibility; the code must preserve the
  *honest error strings* users see today (see module spec `spec-integrations-telegram.md`).

## Evidence

- Subprocess protocol and laws: `integrations/telegram_fetch_worker.py`, `integrations/subprocess_runner.py`
  (verified in investigation report 1-f)
- Single-slot lock: `gui/telegram_lock.py:69-99`; watchdog 35 min (`gui/main_window/window.py:131-135`)
- Worker API independence: `cloud/cloudflare_sync.py:271-304` (get_pending/get_pending_result)

## Verification required in implementation

- PoC: login (code + optional 2FA) over SOCKS5 on Windows; iterate the bot chat with `min_id`;
  `send_read_acknowledge` equivalent (gotd `messages.readHistory` / `ReadHistory` on the peer);
  measure fetch time for a 1000-message chat.

## Amendment 1 — Interim PoC verdict (T-01, SESSION-001, 2026-10-15)

**Status: provisional — every autonomous leg proven; the owner-assisted leg is deferred to
SESSION-002 by owner decision (owner confirmed the SESSION-001 status report on 2026-10-15
and attached this leg as the next session's first order of business).**

Evidence (full detail: `docs/poc-artifacts/t-01-telegram-poc.md`):

1. **Bot-token login proven direct and through SOCKS5** (`auth.importBotAuthorization`):
   3.9 s direct, 4.1 s via the SOCKS5 relay (x/net/proxy dialer as the `dcs.Resolver` —
   the same wiring ADR-0006 §1 specifies for webfetch). IPv6 DC candidates failed in the
   IPv6-less sandbox and gotd transparently fell back to IPv4 — resilient dialing confirmed.
2. **`messages.getDialogs` from a bot session → `BOT_METHOD_INVALID`** — MTProto-layer
   confirmation of Option D's rejection: the user↔bot queue is unreadable from the bot
   account, so the user login is structurally irreplaceable for the queue workflow.
3. **gotd pinned at v0.110.0** — v0.162.0's generated `tg` package exceeds the sandbox's
   4 GB compile budget (recorded in `build/TOOLCHAIN.md`); revisit the pin on the owner's
   Windows machine.
4. **The full owner-assisted flow is implemented and staged** (commit `4068bb0`): two-phase
   `-request-code` / `-code` login for agentic environments (pending-login state outside
   the repo, chmod 600), the Python worker's `__NEED_PHONE__` / `__NEED_CODE__` /
   `__NEED_PASSWORD__` stderr protocol, `contacts.resolveUsername` → `messages.getHistory`
   pages of 100 (`min_id`, oldest-first) → `messages.readHistory` read-acknowledge, with
   phone masked to last-4 in logs. It runs the moment the owner provides the phone number.

**Final verdict deferred**: the task-list acceptance ("100-message fetch on the owner's
account OR documented fallback decision") lands in SESSION-002; Option B (Worker-only
queue + a small Worker addition for read-acknowledge) remains the recorded contingency
if that leg fails — its decision would then be recorded as a superseding ADR.
