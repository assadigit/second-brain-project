# SESSION-PROTOCOL — how every session runs and hands off

> **The law of this project:** the owner never guesses what's next. Every session ends by
> writing the *next* session's starting prompt into `docs/NEXT-SESSION.md`. The owner's only
> job: open a fresh AI session and paste what that file (or the closing report) says.

## The loop (every session, no exceptions)

1. **Bootstrap** — clone this repo; read `README.md`, this file, `docs/NEXT-SESSION.md`, the
   two most recent reports in `docs/sessions/`, and `git log --oneline -15`.
2. **Orient** — the state card (`NEXT-SESSION.md`) names the phase, the session goal, the
   task IDs, the specs to read, and any known blockers. Read those specs *before* writing
   code. Specs live in `docs/port-blueprint/` (blueprint = binding; ADRs = binding unless
   their PoC verdict says otherwise).
3. **Work** — execute the phase per `docs/port-blueprint/12-migration-roadmap.md` and task
   list `13-implementation-task-list.md`, obeying the standing laws:
   - **Reality wins**: when source code contradicts a spec, verify in source and fix the spec
     in the same commit.
   - **Tests are never weakened, skipped, or deleted to get green.**
   - **No secrets, tokens, session files, or owner data (config.json / cache.db / vaults)
     ever enter git.**
   - **Stay inside the phase**; surface blockers with evidence instead of improvising past
     them.
   - Small, well-messaged commits during the session.
4. **Close** — in this order:
   a. Evaluate the phase exit gate **honestly** (evidence: test output, PoC artifacts, diffs).
   b. Commit and push **all** work (code + spec updates + reports); the tree must end clean.
   c. Write `docs/sessions/SESSION-NNN.md` (next free number; template below).
   d. **Rewrite `docs/NEXT-SESSION.md`** for the next session (template below) and push it.
   e. If the environment has a worklog (`/home/z/my-project/worklog.md`), append the session
      record there too (house `---`/Task ID format).
5. **Report to the owner in chat** — short plain-language report (done / blocked / evidence)
   **plus the literal next prompt** to copy.

## The eternal prompt (fallback — always works, even if chat context is lost)

> You are continuing the second-brain-project (the Go port of GitCurator). Clone
> https://github.com/assadigit/second-brain-project and follow `docs/SESSION-PROTOCOL.md`.
> Your state card is `docs/NEXT-SESSION.md` — read it first and execute exactly what it says.

## `docs/sessions/SESSION-NNN.md` template

```markdown
# SESSION-NNN — <date> — <phase> — <one-line result>

## Goal (from NEXT-SESSION.md at start)
## What was done (task IDs, files, commits — cite hashes)
## Gate evaluation (pass/fail/partial + EVIDENCE: test output, artifacts, numbers)
## Spec updates made (file → what changed, why)
## Blockers & decisions needed
## Owner actions requested (if any)
## Handoff
- Next session: SESSION-NNN+1 = <phase/goal>
- NEXT-SESSION.md rewritten: yes/no
- The next prompt: (verbatim block)
```

## `docs/NEXT-SESSION.md` template (rewritten at every close)

```markdown
# NEXT-SESSION — the state card (rewritten at every session close)

## 1. Where the project stands (phase, last session, gate status)
## 2. Next session: ID, phase, goal, task IDs, exit gate
## 3. Read-first list (exact spec/ADR files)
## 4. Known constraints / blockers / owner actions
## 5. THE PROMPT TO COPY (verbatim, self-sufficient)
```

## Hard rules

- **Formal releases** (tags, artifacts, GitHub releases) happen only at roadmap milestones
  (GUI parity, final release) — never as a per-session ritual. Per-session: commit + push +
  reports only.
- **Phase gates are hard.** Don't start the next phase in the same session unless the gate is
  green *and* capacity remains; prefer closing cleanly. One phase done properly beats two
  half-done.
- **If blocked:** write the blocker (evidence + the decision needed) into `NEXT-SESSION.md`;
  the next session or the owner resolves it *first*.
- **Owner decisions arriving mid-session** get recorded in the session report and reflected
  in `NEXT-SESSION.md`; consequential ones also amend the relevant ADR (append-only).
- **Session numbering** is monotonic; never reuse or renumber.
