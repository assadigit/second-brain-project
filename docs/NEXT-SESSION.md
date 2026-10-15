# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- Blueprint complete, owner-confirmed, mirrored here at `docs/port-blueprint/` (canonical).
  Language: **Go**. GUI: **Wails** (Fyne = PoC-failure fallback only). Repo: private,
  codename `second-brain-project` (rebrand later). All owner decisions recorded in
  `docs/port-blueprint/14-risks-assumptions-open-questions.md` §6 (Rounds 1–4).
- Audit currency verified 2026-10-15: the blueprint matches the released Python v0.91.0
  (GitCurator `main` @ `98f4fc7`; zero app-code drift vs the audited commit).
- **Nothing has been ported yet.** SESSION-000 created this protocol; SESSION-001 is next.

## 2. Next session: SESSION-001 — Phase 0 "Proofs & Baselines"

- **Goal:** retire the three risky library decisions with working PoCs, stand up the
  differential harness, capture Python performance baselines.
- **Exit gate:** PoC verdicts recorded as ADR updates (`decisions/0004`, `0006`, `0001`);
  harness runs one trivial differential case end-to-end; baseline table filled with *measured*
  numbers (never estimated).
- **Tasks:** T-00 repo skeleton (module `github.com/assadigit/second-brain-project` at repo
  root) · T-01 gotd/td Telegram PoC · T-02 Chrome-TLS impersonation PoC · T-03 differential
  harness skeleton · T-04 Python performance baselines · T-70 Wails + silent WebView2
  bootstrapper PoC.

## 3. Read-first list (in order)

1. `docs/port-blueprint/16-next-session-implementation-prompt.md` — the full briefing **incl.
   the 2026-10 Addendum** (new-repo location, CLI re-scope, Wails confirmation, icon).
2. `docs/port-blueprint/12-migration-roadmap.md` — Phase 0 + Amendment 1.
3. `docs/port-blueprint/13-implementation-task-list.md` — T-00…T-04, T-70 details.
4. `docs/port-blueprint/decisions/0001-gui-framework.md`, `0004-telegram-gotd.md`,
   `0006-network-stack.md` — the ADRs your PoCs will confirm or overturn.

## 4. Known constraints / owner actions for SESSION-001

- **T-01 needs the owner reachable** (Telegram login code arrives on their phone) and needs
  the bot token + api_id/api_hash — pasted into chat only, **never committed**.
- **T-70's bootstrapper leg needs a real Windows VM/machine without WebView2** (the owner's
  scenario). If unavailable in the environment: record as blocked-with-plan, continue the
  other tasks, and list it as the owner's action.
- Python reference: `https://github.com/assadigit/GitCurator` @ `98f4fc7` — **read-only**,
  clone alongside; the differential harness drives its unmodified code.
- Sandbox note: PyQt6 may be unavailable; GUI-behavior baselines use the repo's own tests /
  `QT_QPA_PLATFORM=offscreen` where possible, and honest "not measurable here" entries
  otherwise.

## 5. THE PROMPT TO COPY (SESSION-001)

```
You are starting SESSION-001 of the second-brain-project — the multi-session Go port of
the GitCurator Python app (a personal second-brain library builder; full specs exist).

1. Clone https://github.com/assadigit/second-brain-project (private; I will authorize
   access if your environment needs it). The read-only Python reference repo is
   https://github.com/assadigit/GitCurator (v0.91.0, main 98f4fc7) — clone it alongside;
   never modify it.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working law of this
   project) → docs/NEXT-SESSION.md (your state card — it fully describes this session).

3. Execute SESSION-001 = Phase 0 "Proofs & Baselines" of
   docs/port-blueprint/12-migration-roadmap.md:
   - T-00: repo skeleton (Go module github.com/assadigit/second-brain-project, repo root,
     empty package tree per docs/port-blueprint/04-target-go-architecture.md §1).
   - T-01: gotd/td Telegram PoC — login + bot-chat iteration + read-acknowledge. I will be
     present to receive the login code; I will paste credentials in chat when asked. Never
     write credentials or session files into the repo.
   - T-02: Chrome-TLS impersonation PoC (stdlib UA gets 403 on a walled page; impersonation
     profile succeeds) + library license check.
   - T-70: Wails PoC — themed window + log desk + one worker job, and the silent WebView2
     Evergreen-bootstrapper path on a Windows machine/VM WITHOUT WebView2 (if no Windows
     environment is available, record it as blocked-with-plan and continue).
   - T-03: differential harness skeleton per docs/port-blueprint/09-testing-and-parity.md §3,
     one trivial case green end-to-end (e.g. --version output parity).
   - T-04: Python performance baselines per
     docs/port-blueprint/07-ui-and-performance-assessment.md §B.4 — measured numbers only.

4. Work strictly inside Phase 0. Standing laws: reality wins over specs (fix the spec in the
   same commit); tests are never weakened; no secrets or owner data in git; surface blockers
   with evidence instead of improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation, commit + push
   everything, write docs/sessions/SESSION-001.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language report plus the prompt I
   should copy to start SESSION-002.

Do not attempt to finish the whole port — one phase, done properly, with proof.
```
