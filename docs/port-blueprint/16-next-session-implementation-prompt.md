# 16 — Next-Session Implementation Prompt

> Copy everything below the line into the next AI coding-agent session's first message
> (adjust only the working-directory line if the repo is cloned elsewhere).

---

You are continuing the GitCurator Python→Go migration. Work in the repository root
(`/home/z/GitCurator` or your local clone of https://github.com/assadigit/GitCurator, branch
`main`).

## Your briefing

1. **Read first, in this order:**
   - `port-blueprint/README.md` (the reading guide)
   - `port-blueprint/00-executive-summary.md`
   - `port-blueprint/12-migration-roadmap.md` and `port-blueprint/13-implementation-task-list.md`
   - `port-blueprint/decisions/` (all 10 ADRs — they are binding unless their PoC verdicts say
     otherwise)
   - The module specs for whatever you implement that day
     (`port-blueprint/module-specs/spec-*.md`).
2. **The blueprint is evidence-based but not gospel.** Every claim cites `path:line` at v0.91.0
   (HEAD 8fd4179). If the source code contradicts a spec, **reality wins**: verify in the source,
   fix the spec file in the same commit, and note the deviation in your report. Never implement
   from the spec against the code.
3. **Scope law**: you implement the Go app under `go-desktop/` ONLY. Do not modify the Python
   application, its tests, configuration, CI, or release tooling. The Cloudflare Worker and the
   Next.js apps are out of scope entirely (ADR-0005). No visual redesign (ADR-0001 §"out of
   scope").
4. **Start with Phase 0** (roadmap): environment, the three PoCs (gotd, tls-client, Wails), the
   differential harness skeleton, and the Python performance baselines. Do NOT skip the PoCs —
   they gate ADRs 0001/0004/0006.
5. **Test law** (the repo's own culture, now yours):
   - Port the corresponding Python tests *before* porting each module (characterization first).
   - Tests are never weakened, skipped, or deleted to get green.
   - New behavior-preserving helpers get table tests including the Unicode edge fixtures.
   - Keep `go vet` and `gofmt` clean.
6. **Behavior preservation defaults**: preserve observable behavior including quirks; the only
   pre-approved fixes are the five items in `port-blueprint/09-testing-and-parity.md` §5. Anything
   else you believe is a bug gets recorded, not fixed (unless it blocks compilation).
7. **Secrets law**: never print, log, commit, or embed credentials, tokens, session files, or the
   owner's vault content. The git-token one-time-URL law and FETCH_HEAD redaction are hard
   requirements (doc 11 §2).
8. **Data law**: the Go app opens the owner's real `cache.db`/`config.json` in place; same DDL,
   same merge laws, same lock file (ADR-0008). Dry-run must write nothing (shadow-cache law).
9. **Report cadence**: work in the roadmap's phase order; at each phase exit gate, produce a
   short report (completed / pending / blocked + evidence: test output, diff reports, PoC
   artifacts). Stop and surface (do not improvise past) any blocker in
   `port-blueprint/14-risks-assumptions-open-questions.md` §3.
10. **Update the worklog**: append your session record to `/home/z/my-project/worklog.md` (or the
    repo's equivalent if working elsewhere) using the established `---`/Task ID format, and update
    the affected blueprint files when reality differs.

## Your first concrete steps (Phase 0)

1. `T-00`: create `go-desktop/` skeleton (module `github.com/assadigit/GitCurator/go-desktop`,
   `go 1.22+`), empty packages per `04-target-go-architecture.md` §1, `go build ./...` green.
2. `T-01`: gotd PoC (login + bot-chat iteration + read-acknowledge via SOCKS). Record the verdict
   in `decisions/0004-telegram-gotd.md` (status → accepted/updated).
3. `T-02`: tls-client PoC + license check. Record in `decisions/0006-network-stack.md`.
4. `T-03`: differential harness skeleton per `09-testing-and-parity.md` §3 (case dirs, Python
   driver that *runs* the unmodified Python app, normalizing comparator). One case green.
5. `T-04`: Python performance baselines per `07-ui-and-performance-assessment.md` §B.4; record
   the measured numbers in your report (never estimate them).

## Definitions of success for your session

- Phase 0 exit gate met (PoC verdicts recorded, harness skeleton demonstrated, baselines
  measured), or — if blocked — a precise blocker report citing the failing evidence and the
  decision you need.
- No modifications outside `go-desktop/` (plus blueprint/ADR updates inside
  `port-blueprint/`).
- All claims of "works" backed by test output or recorded artifacts.

Remember the owner: a designer on Windows 10 who launches with `.bat` files, reads plain-language
reports, and values "no link left behind" over elegance. Deliver complete, verified, working
increments.

---

## Addendum (2026-10, post-blueprint owner decisions — read before executing anything above)

These owner decisions supersede the *where* and *milestone* details above; the *what* and *laws*
are unchanged:

1. **Work in the private repo `assadigit/second-brain-project`** (already created and seeded with
   this blueprint at `docs/port-blueprint/`), NOT a `go-desktop/` directory inside GitCurator.
   Module path: `github.com/assadigit/second-brain-project` (ADR-0002 Am.1). Keep a local clone
   of `assadigit/GitCurator` (pinned rev) as the unmodified Python differential reference.
2. **The owner-facing CLI milestone is CUT** — CLI is internal proving ground / test driver /
   emergency fallback only; no terminal-paint polish, no `.bat` launcher parity, no tagged CLI
   release. The **GUI (Phase 6) is the first owner-visible release**, carrying the tag + dual
   artifact (setup.exe + portable zip) ritual (ADR-0007 Am.1, roadmap Am.1, ADR-0010 Am.1).
3. **Wails is owner-confirmed**; WebView2 is *assumed absent* on the owner's machine (EOL Win10,
   Edge deleted) — the Evergreen bootstrapper is the expected path and **T-70 must prove it
   end-to-end on a WebView2-less Windows VM** (ADR-0001 Am.1).
4. **Icon**: use `assets/icons/gitcurator.ico` (repo's brand mark, ready-made per owner; swap
   file later if the owner provides a custom one) — ADR-0010 Am.2.
5. Telegram: **one-time re-login accepted by the owner** (still gated by PoC T-01). Data: **open
   the owner's real cache.db/config.json in place** (backup ritual + cross-binary lock,
   ADR-0008). Windows-only v1. All quirks preserved (safest). All owner answers: doc 14 §6.
6. In step 1 of "Your first concrete steps", create the skeleton at the **repo root** of
   `second-brain-project` (not `go-desktop/`), module path as above; the package tree from
   `04-target-go-architecture.md` §1 is unchanged otherwise.
