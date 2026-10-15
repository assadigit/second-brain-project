# SESSION-001 — 2026-10-15 — Phase 0 "Proofs & Baselines" — all autonomous legs proven; owner-assisted T-01 leg deferred with owner confirmation

> Close-out note: the session's working commits (`8de1d7d`…`4068bb0`) landed during
> execution; the session's context window ended before the protocol close (report +
> state-card rewrite) could be written. The owner confirmed the in-chat status report on
> 2026-10-15 and instructed that its pending items be attached to the next session's
> demands; this report, the ADR-0004 interim verdict, ADR-0006 Am.2, and
> `NEXT-SESSION.md` v2 complete that close.

## Goal (from NEXT-SESSION.md at start)

Retire the three risky library decisions with working PoCs (ADR-0004 gotd/td, ADR-0006
tls-client impersonation, ADR-0001 Wails), stand up the differential harness, capture
Python performance baselines. Tasks: T-00, T-01, T-02, T-03, T-04, T-70.

## What was done (task IDs, files, commits)

| Task | Commit | Result |
|---|---|---|
| T-00 repo skeleton | `8de1d7d` | Module `github.com/assadigit/second-brain-project` at repo root (ADR-0002 Am.1 layout: `cmd/`, `internal/`, `assets/`, `webui/`, `build/`); 33 internal packages as `doc.go` placeholders, each citing Python origin + module spec + roadmap task; `internal/cli` Phase-0 seed byte-exact with Python `--status` no-config output; `.github/workflows/go-ci.yml` stub; `.gitignore` hardened per doc 11 §2; `assets/embed.go` icon embed; `build/TOOLCHAIN.md` records go1.27.2. Spec fix in-commit: doc 04 §1 updated to the repo-root layout. Python smoke: `tests.test_core` 34/34 OK. |
| T-01 gotd/td PoC | `1349ae7`, `4068bb0` | Bot-token login proven direct (3.9 s) and through SOCKS5 (4.1 s); `messages.getDialogs` from a bot session → `BOT_METHOD_INVALID` (hard proof the user login is irreplaceable); hermetic SOCKS5 relay logs the DC CONNECT targets (IPv6→IPv4 fallback survived). `4068bb0` adds the two-phase user login (`-request-code` / `-code`, optional `-password`) for agentic environments — pending-login state outside the repo, phone masked to last-4. Evidence: `docs/poc-artifacts/t-01-telegram-poc.md`. gotd pinned v0.110.0 (v0.162 `tg` package exceeds the sandbox's 4 GB compile budget). |
| T-02 Chrome-TLS PoC | `c1c5401` | Live fingerprint substitution on tls.peet.ws (Go stdlib JA4 → Chrome_133's exact JA4 + HTTP/2 fp); 14 real walled pages reproduce the 403/429 trigger matrix for stdlib; live 403→200 blocked at the IP-reputation layer for both transports from the datacenter IP (statuses recorded) — mechanism proven hermetically: fingerprint-gated local wall, stdlib 403 / Chrome profile 200 (`WALL_BYPASSED`); SOCKS5 passes the domain unresolved (`ATYP=DOMAIN` — the rdns law). Licence reality fix: tls-client is **BSD-4-Clause**, not Apache-2.0. Evidence: `docs/poc-artifacts/t-02-tlsimp-poc.md`; ADR-0006 Am.1 appended. |
| T-70 Wails PoC | `8e07893` | Themed window (real cream/plum tokens from `gui/theme.py`) + coalescing log desk (1024-cap channel, 120 ms ticker, ≤250-row batches) + one sequential worker job. Built and run on Linux (rootless webkit2gtk-4.1 + Xvfb + LD_PRELOAD shim; screenshot committed and VLM-verified). Gates: startup→first frame **213–301 ms** (≤1.5 s PASS); flood **625k lines/s, 0 drops** (≥2k rows/s PASS); exe **9.1 MB Linux / 11.8 MB Windows** cross-compile (<20 MB PASS). Windows WebView2-less bootstrapper leg = **blocked-with-plan** (no Windows in the sandbox) — owner action, does not block Phases 1–5. Evidence: `docs/poc-artifacts/t-70-wails-poc.md` + screenshot; ADR-0001 Am.2 appended. |
| T-03 differential harness | `0087126` | `testdiff/`: `harness.py` drives the unmodified Python app and the Go CLI on the same scratch cwd + env, normalizes per doc 09 §3 (timestamps, ISO stamps, temp paths, durations, UA version strings), diffs stdout/stderr/exit code, writes sha256-stamped reports. First case `status-no-config`: **PASS** — byte-identical after normalization (raw stdout sha256s matched byte-for-byte this run). Report archived: `docs/poc-artifacts/t-03-status-no-config-report.md`. Reference repo stays pristine. |
| T-04 Python baselines | `80cb8ca` | `testdiff/python_baselines.py` — the doc 07 §B.4 procedure as a rerunnable rig (fresh subprocesses for honest cold runs, offscreen PyQt6, temp CONFIG_FILE). Measured (sandbox Xeon — environment context; the owner's Windows will differ; the rig is the deliverable): GUI cold start→frame ~805 ms offscreen; CLI `--status` 49.3 ms (no config) / 893.6 ms (config card); log ingest 9.29 µs/line; scan 1,314-note synthetic vault 35.3 ms cold / 16.7 ms warm; golden offline batch 30/30 links in 205 ms; queue classification 1,142 links in 1.42 ms; idle-GUI RSS 121.6–135.6 MB. Evidence: `docs/poc-artifacts/t-04-python-baselines.md`. |

## Gate evaluation (pass/fail/partial + EVIDENCE)

Phase-0 exit gate (roadmap §Phase 0): *"PoC verdicts recorded as ADR updates; harness
runs one trivial case end-to-end; baseline table filled."*

- **ADR-0006 (T-02)** — verdict recorded as Am.1: tls-client v1.16.0 accepted with the
  licence reality correction; utls fallback recorded. ✅
- **ADR-0001 (T-70)** — verdict recorded as Am.2: Wails v2.16.0 accepted, all three
  acceptance gates PASS with measured numbers; bootstrapper leg blocked-with-plan. ✅
- **ADR-0004 (T-01)** — interim verdict recorded as Am.1 (at this close): every
  autonomous leg proven (bot login direct + SOCKS5; BOT_METHOD_INVALID); the
  owner-assisted leg could not run without the owner's phone number. **Honest partial**:
  the user-login leg is structurally impossible to fake — `BOT_METHOD_INVALID` proves the
  bot account cannot read the user↔bot queue — so it must run with the owner. Deferred to
  SESSION-002 by owner decision; final verdict (or the Option-B fallback) lands there. ⚠️→SESSION-002
- **Harness** — one case green end-to-end (`status-no-config`, PASS, report archived). ✅
- **Baselines** — table filled with measured numbers only (t-04 artifact). ✅

**Verdict: Phase 0 gate green modulo one owner-dependent leg** — not a failure, an
explicit deferral with the owner's confirmation (2026-10-15).

## Spec updates made

- `docs/port-blueprint/04-target-go-architecture.md` §1 → repo-root layout per ADR-0002
  Am.1 (commit `8de1d7d` — it still described the superseded `go-desktop/` plan).
- `docs/port-blueprint/decisions/0006-network-stack.md` Am.1 — T-02 verdict + licence
  reality (commit `c1c5401`).
- `docs/port-blueprint/decisions/0001-gui-framework.md` Am.2 — T-70 verdict (commit
  `8e07893`).
- `build/TOOLCHAIN.md` — gotd v0.110.0 pin + compile-budget record.
- At close: ADR-0004 Am.1 (interim T-01 verdict), ADR-0006 Am.2 (owner licence
  acknowledgement), `README.md` status line, `docs/NEXT-SESSION.md` v2.

## Blockers & decisions needed

- **T-01 user-login leg**: needs the owner's phone number + login code (chat only, never
  committed). The owner confirmed the report and attached this leg as SESSION-002's first
  order of business — recorded here and in `NEXT-SESSION.md` v2.
- **T-70 Windows bootstrapper leg**: no Windows environment in the sandbox —
  blocked-with-plan (concrete NSIS/bootstrapper plan in the t-70 artifact); owner action;
  does not block Phases 1–5.
- **gotd pin**: v0.110.0 (v0.162 compile budget) — revisit on the owner's Windows machine;
  recorded, not blocking.
- **Open owner decision (stamp pending)**: in the post-report chat the owner proposed
  dropping the desktop's direct Telegram leg entirely (country-level Telegram block; the
  queue would flow only via the Cloudflare Worker) — i.e., adopting ADR-0004's Option B
  as the primary architecture, with gotd/td demoted to an optional Phase-8 enhancement.
  The feasibility analysis (environment worklog, Task ID 10) recommended adoption — zero
  Worker changes for the core flow, no proxy for the queue leg, no one-time re-login,
  less censorship exposure; losses: read-acknowledge badge-clearing, the
  Saved-Messages/keyword/single-id modes, webhook-miss recovery via chat re-read
  (mitigation: re-forward), the Telegram leg of Test Connection. The owner has **not yet
  stamped the decision**; the 2026-10-15 confirmation covered the SESSION-001 status
  report, so the attached T-01 demand stands as written until this fork is resolved.
  Carried in `NEXT-SESSION.md` v2 §4.

## Owner actions requested

1. **At SESSION-002 start**: paste the Telegram phone number in international format
   (e.g. `+98XXXXXXXXXX`); be reachable to read the login code in the Telegram app (and
   the 2FA password if prompted); have api_id/api_hash + bot token ready to paste in chat
   when asked. Never committed; session files live outside the repo.
2. **Optional, any session**: a Windows VM/machine without WebView2 for the T-70
   bootstrapper leg.
3. ~~Licence acknowledgement (BSD-4-Clause tls-client + fhttp)~~ — **resolved
   2026-10-15**: the owner confirmed the session report; recorded as ADR-0006 Am.2. If
   the owner ever intends otherwise, the utls fallback path stays ready.

## Owner decisions recorded (2026-10-15, this close)

- The owner confirmed the SESSION-001 status report **in full** — including the
  ADR-0006 licence-reality acknowledgement — and instructed that its pending items be
  attached to the next session's demands. Reflected in: this report, ADR-0004 Am.1,
  ADR-0006 Am.2, and `NEXT-SESSION.md` v2 (§2, §4, §5).
- Explicitly **not** stamped by that confirmation: the Option-B adoption proposal from
  the post-report chat (see Blockers above) — it remains the owner's open call, recorded
  as an open decision in `NEXT-SESSION.md` v2 §4.

## Handoff

- Next session: **SESSION-002 = T-01 owner-assisted leg completion → ADR-0004 final
  verdict; then Phase 1 "Pure-foundation packages" (T-10…T-17) per roadmap order** —
  Phase 1 will likely span multiple sessions; take tasks in order and close cleanly.
- `NEXT-SESSION.md` rewritten: yes (v2).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies it verbatim.
