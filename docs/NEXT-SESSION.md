# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 "Proofs & Baselines" executed and owner-confirmed (2026-10-15).** SESSION-001
  working commits `8de1d7d`…`4068bb0` are on `main`; full evidence in
  `docs/sessions/SESSION-001.md` and `docs/poc-artifacts/`.
- PoC verdicts: **Wails accepted** (ADR-0001 Am.2 — all gates PASS: 213–301 ms startup,
  625k lines/s flood, 9.1/11.8 MB exe) · **tls-client accepted** (ADR-0006 Am.1 + Am.2 —
  licence reality BSD-4-Clause acknowledged by the owner; utls fallback stays recorded) ·
  **gotd/td provisional** (ADR-0004 Am.1 — every autonomous leg proven; the
  owner-assisted user-login leg is THIS session's first order of business).
- Differential harness: `testdiff/` green on its first case (`status-no-config`).
- Python baselines: measured and archived (t-04 artifact; the rig is rerunnable).
- **Phase 1 has not started.** Go skeleton: 33 packages as placeholders + CLI parity seed.
- Outstanding owner-dependent legs: **T-01 user login** (phone + code) and **T-70
  Windows bootstrapper** (Windows VM without WebView2 — non-blocking for Phases 1–5).

## 2. Next session: SESSION-002 — T-01 completion + Phase 1 "Pure-foundation packages"

- **Goal:** (a) complete T-01's owner-assisted leg — the demand the owner attached on
  2026-10-15 — then (b) start Phase 1 (byte-parity foundations) with the remaining
  capacity.
- **T-01 completion, exact shape:** drive the staged two-phase login in `poc/telegram`
  (`-mode=userlogin -request-code` with the owner's phone → owner reads the code in their
  Telegram app and pastes it in chat → `-code …`, optional `-password`), then run the
  **100-message bot-chat fetch + read-acknowledge** on the owner's account
  (`messages.getHistory` pages of 100, `min_id`, oldest-first → `messages.readHistory`),
  record measured timings, and write **ADR-0004's final verdict** — or, if the leg fails,
  the documented Option-B (Worker-only) fallback decision as a superseding ADR.
- **Phase 1 work** (roadmap §Phase 1; task list T-10…T-17, in order): `constants` →
  `links` → `storage`+`dryrun` → `prompts` → `taxonomy` → `notebuilder` → `config` →
  rune/rounding/ordering helpers. Port the Python characterization tests FIRST, then the
  code. Phase 1 will likely span more than one session — take tasks in order and close
  cleanly when capacity runs out.
- **Exit gate:** T-01 leg done (verdict in ADR-0004) + as much of Phase 1 as capacity
  allows, with its gate honestly evaluated: characterization tests green (ported
  `test_core`, `test_userconfig`, taxonomy tests); differential case "import-file
  grammar" matches; config round-trip byte-stable against a real `config.json` copy.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-001.md` — what SESSION-001 proved and left you.
2. `docs/poc-artifacts/t-01-telegram-poc.md` — the staged login procedure you will drive.
3. `docs/port-blueprint/decisions/0004-telegram-gotd.md` (incl. Am.1) — the verdict you
   will finalize.
4. `docs/port-blueprint/12-migration-roadmap.md` — Phase 1 (+ Amendment 1 context).
5. `docs/port-blueprint/13-implementation-task-list.md` — T-10…T-17 details + the
   cross-cutting standing orders (tests first; spec fixes in-commit; no new deps without
   an ADR).
6. `docs/port-blueprint/05-python-to-go-mapping.md` §3 — semantics items 1–13 (Phase 1's
   key risks).
7. The module spec for whichever package is being ported (`spec-core-constants`,
   `spec-core-links`, `spec-core-storage`, `spec-core-prompts`, `spec-core-taxonomy`,
   `spec-core-note-builder`, `spec-core-user-config`) — read each before its task.

## 4. Known constraints / owner actions for SESSION-002

- **THE ATTACHED DEMAND (owner, 2026-10-15): complete the T-01 owner-assisted leg.** The
  owner pastes their Telegram phone number in international format at session start;
  Telegram delivers the login code to their app; the owner pastes the code in chat (and
  the 2FA password if prompted). Credentials travel through chat/environment only —
  **never committed**; the session file and pending-login state live outside the repo
  (`~/.cache/second-brain/`).
- api_id/api_hash + bot token also needed for the PoC legs — pasted in chat when asked,
  never committed (unchanged law from SESSION-001).
- **T-70 Windows bootstrapper leg** stays an owner action (a Windows VM/machine without
  WebView2); it does not block Phases 1–5 and can be folded into any later session.
- gotd is pinned at v0.110.0 (v0.162's `tg` package exceeds the sandbox's 4 GB compile
  budget — `build/TOOLCHAIN.md`); revisit the pin when a Windows environment exists.
- Python reference: `https://github.com/assadigit/GitCurator` @ `98f4fc7` — **read-only**,
  clone alongside; the differential harness drives its unmodified code.

## 5. THE PROMPT TO COPY (SESSION-002)

```
You are starting SESSION-002 of the second-brain-project — the multi-session Go port of
the GitCurator Python app (a personal second-brain library builder; full specs exist).

1. Clone https://github.com/assadigit/second-brain-project (private; I will authorize
   access if your environment needs it). The read-only Python reference repo is
   https://github.com/assadigit/GitCurator (v0.91.0, main 98f4fc7) — clone it alongside;
   never modify it.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working law of this
   project) → docs/NEXT-SESSION.md (your state card) → docs/sessions/SESSION-001.md (what
   SESSION-001 proved and left you).

3. FIRST ORDER OF BUSINESS — complete T-01's owner-assisted leg (the demand I attached to
   this session on 2026-10-15): I will paste my Telegram phone number in international
   format (e.g. +98XXXXXXXXXX) now, and my api_id/api_hash + bot token when asked (chat
   only — never committed). Drive the staged two-phase login in poc/telegram
   (-mode=userlogin -request-code → I read the code in my Telegram app and paste it here
   → -code …, optional -password), then run the 100-message bot-chat fetch +
   read-acknowledge on my account, record the measured timings, and write ADR-0004's
   FINAL verdict — or, if the leg fails, the documented Option-B (Worker-only) fallback
   decision as a superseding ADR. Session and login-state files stay outside the repo.

4. THEN start Phase 1 "Pure-foundation packages (byte-parity layer)" per
   docs/port-blueprint/12-migration-roadmap.md: tasks T-10…T-17 from
   docs/port-blueprint/13-implementation-task-list.md, in order — port the Python
   characterization tests first, then the code, with the spec for each package read
   before its task. Phase 1 will likely span more than one session; when capacity runs
   out, close cleanly rather than half-done.

5. Work under the standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git; surface blockers
   with evidence instead of improvising past them.

6. Close the session per SESSION-PROTOCOL.md: honest gate evaluation, commit + push
   everything, write docs/sessions/SESSION-002.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language report plus the prompt
   I should copy to start SESSION-003.

Do not attempt to finish the whole port — the T-01 leg plus a clean slice of Phase 1,
done properly, with proof.
```
