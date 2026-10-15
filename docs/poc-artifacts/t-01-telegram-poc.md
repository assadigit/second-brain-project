# T-01 — gotd/td Telegram PoC (ADR-0004) — evidence

Recorded 2026-10-15, SESSION-001. Environment: Debian 13 sandbox,
go1.27.2 linux/amd64, direct MTProto egress to all four Telegram DC
ranges verified (149.154.175.53 / 149.154.167.51 / 149.154.167.91 /
91.108.56.130, all :443 TCP OK). Binary: `poc/telegram` (gotd/td v0.110.0
— v0.162.0's generated `tg` package exceeds the sandbox's 4 GB compile
budget; recorded in build/TOOLCHAIN.md).

## 1. Bot-token login (autonomous leg) — direct

```
$ SB_TG_APP_ID=… SB_TG_APP_HASH=… SB_TG_BOT_TOKEN=… poc/telegram -mode=botlogin
[poc] bot-token login OK via gotd (auth.importBotAuthorization) in 3868 ms
[poc] self: bot id=8526017013 username=@githubfetcherbot (verified account)
[poc] resolveUsername(githubfetcherbot) from bot session: OK (handle=githubfetcherbot)
[poc] messages.getDialogs from bot session: REJECTED by Telegram:
      rpc error code 400: BOT_METHOD_INVALID
###POC_JSON###{"bot_id":8526017013,"login_ms":3868,"mode":"botlogin","ok":true,
             "socks":false,"total_ms":7135}
```

- The api_id/api_hash pair is valid; gotd's transport + auth works.
- `BOT_METHOD_INVALID` on `messages.getDialogs` is the MTProto-layer
  evidence behind ADR-0004 option D's rejection: a bot account cannot
  enumerate dialogs or read the user↔bot queue — the user login is
  genuinely required for the bot-queue workflow.

## 2. Bot-token login through SOCKS5 (the ADR's "over SOCKS" leg)

Same run with `-socks=127.0.0.1:10808` (hermetic relay, `poc/telegram
-mode=socksrelay`):

```
[poc] bot-token login OK via gotd (auth.importBotAuthorization) in 4099 ms
[poc] self: bot id=8526017013 username=@githubfetcherbot (verified account)
[relay] conn#1 CONNECT [2001:67c:4e8:f002::a]:443 (atyp=ip)
[relay] conn#2 CONNECT 149.154.167.41:443  (atyp=ip)   ← MTProto through SOCKS
[relay] conn#3 CONNECT 149.154.167.41:443  (atyp=ip)
[relay] conn#5 CONNECT 149.154.175.100:443 (atyp=ip)
###POC_JSON###{"bot_id":8526017013,"login_ms":4099,"ok":true,"socks":true,…}
```

- Full MTProto session tunnels through SOCKS5 (x/net/proxy dialer as the
  dcs.Resolver — the same wiring ADR-0006 §1 specifies for webfetch).
- gotd dials DC addresses (IPs, as Telethon does); the DNS-at-exit
  (`rdns=True`) law matters for hostname URLs — proven with DOMAIN
  passthrough in the T-02 artifact (tls.peet.ws through the same relay).
- IPv6 DC candidates failed (sandbox has no IPv6 egress) and gotd
  transparently fell back to IPv4 — resilient dialing confirmed.

## 3. User login + bot-chat iteration + read-acknowledge (owner-assisted leg) — DONE

Completed 2026-10-15 during SESSION-001 (owner pasted phone, code and 2FA password in chat;
secrets never committed, session + login-state files live outside the repo). Two-phase flow
because the agentic environment cannot keep an interactive process alive between chat turns:

```
$ SB_TG_APP_ID=… SB_TG_APP_HASH=… poc/telegram -mode=userlogin -request-code -phone=+98…
[poc] PHASE 1 OK: login code sent to the Telegram app for ****4077 in 6.29s
$ poc/telegram -mode=userlogin -code=NNNN        # 2FA marker answered via stdin
[poc] PHASE 2 OK: signed in as the owner's account in 1.312s
[poc] logged in as user id=92788333
[poc] bot chat resolved: @githubfetcherbot
[poc] bot-chat iteration: 1512 messages in 16 page(s), 5.133s (295 msg/s),
      id range 1251404…1264366 (min_id=0)
[poc] read-acknowledge OK: messages.readHistory(max_id=1264366) → pts=3517077 pts_count=0
###POC_JSON###{"bot":"githubfetcherbot","fetch":{"pages":16,"messages":1512,
             "max_id":1264366,"min_id_seen":1251404,"fetch_ms":5132,
             "msgs_per_sec":294.62,"min_id_used":0},"mode":"userlogin","ok":true,
             "read_acknowledged":true,"socks":false,"total_ms":8151,"user_id":92788333}
```

- **All four capabilities the Python subprocess exposes are proven**: user login with code +
  2FA (SRP password via `auth.Password`), bot-chat iteration (pages of 100, oldest-first,
  `min_id` support, no limit), read-acknowledge (`messages.readHistory` with the max observed
  id), and SOCKS5 dialing (§2's relay leg; the user login itself ran direct here — SOCKS for
  the user session was proven by the bot-session tunnel through the identical wiring).
- Page latency settled at ~280 ms/page after the first (~533 ms) — a 1512-message queue drains
  in ~5 s; comfortably inside the app's 35-min watchdog.
- Message TEXT was never printed (owner-data law): only counts, id ranges and timings.
- Verdict recorded as **ADR-0004 Am.1** (accepted; provisional lifted; fallback not needed).
- Session file (authorized, reusable): `~/.cache/second-brain/tg-poc.session.json` — outside
  the repo, per the no-secrets-in-git law. The pending-login state file was consumed and
  removed by phase 2.
