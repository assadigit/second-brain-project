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

## 3. User login + bot-chat iteration + read-acknowledge (owner-assisted leg)

Pending the owner's phone number and login code (pasted in chat, never
committed). The PoC implements the full flow: `__NEED_PHONE__` →
`__NEED_CODE__` → optional `__NEED_PASSWORD__` markers on stderr (the
Python worker's protocol, plus a phone prelude), then
`contacts.resolveUsername(githubfetcherbot)` → `messages.getHistory`
pages of 100 (min_id support, everything fetched, oldest-first) →
`messages.readHistory(max_id)` read-acknowledge → timing summary.
Session file: outside the repo (`~/.cache/second-brain/tg-poc.session.json`).

This leg runs the moment the owner provides the phone number; completion
is the first item of SESSION-002 if the code does not arrive during this
session.
