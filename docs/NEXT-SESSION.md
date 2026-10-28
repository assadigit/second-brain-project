# NEXT-SESSION — the state card (rewritten at every session close)

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0–6 CLOSED** (the foundation, the CLI parity release
  `v0.92.0-go.1`, the Wails GUI through T-79f).
- **Phase 7 — OPEN, THE FIX CYCLE (SESSION-048–081)**: T-80 ✓,
  T-81 ✓ (the release workflow), T-83 ✓, AC-17 ✓, the desktop
  milestone release `v0.92.0-go.2` ✓, the owner's first live run
  failed at go.2 (SESSION-074), go.3 came GREEN (SESSION-075, THE
  FIRST GREEN LIVE RUN EVER), the six problems fixed at go.4
  (SESSION-075), the three improvements at go.5 (SESSION-076), the
  five owner-directive fixes at go.6 (SESSION-077), the four at go.7
  (SESSION-078), the Direct-Pair UI + diagnostic logging at go.9
  (SESSION-079), the test-connection-DoH law at go.10 (SESSION-080).
  **SESSION-081 = THE D1-API-FALLBACK release `v0.92.0-go.11` LIVE +
  VERIFIED THREE WAYS** (go-ci run #170 GREEN + release run #13 GREEN
  + the release 426430005 LIVE; the asset 27,452,357 bytes; sha256
  `2b3b82ac…3f88`; the VERSION member + the GUI ProductVersion +
  the CLI buildinfo + the tag all carry one number). THE D1-API
  FALLBACK closes the ISP-blocks-workers.dev gap: when the direct
  Worker request fails, the honest poll falls back to
  `GetPendingViaD1()` (a direct D1 query via api.cloudflare.com — a
  DIFFERENT edge the ISP does NOT block; the SQL mirrors the Worker's
  `handleGetPending` exactly; read-only, no ack). Six hermetic
  pinning tests pin every law. The live-D1 verification confirmed the
  code's transport + envelope + SQL-parse (the owner's `curator-bot`
  D1 is empty for now — the schema is created lazily on the bot's
  first write; the fallback fetches live the moment the bot processes
  a URL).
- **Push state: CLEAN** — SESSION-081's close push (the report's
  verification section filled) is the latest; the tree is clean.
- **The repo is PUBLIC**: pushes need the PAT (ephemeral — re-inline
  per call; the sandbox shell does not carry exports between tool
  calls); Actions minutes are unlimited; the no-secrets law is
  public-facing. **go-ci: GREEN on the tree through SESSION-081.**
- **The operational laws, newest first**: **THE D1-API-FALLBACK
  RECORD (SESSION-081 — the honest poll falls back to a direct D1
  query via api.cloudflare.com when the direct Worker request fails;
  read-only, no ack; the SQL mirrors the Worker's handleGetPending
  exactly; the d1APIBase seam is a var; the same-decoder law — the
  envelope rides storage.DecodeJSONValue so integers surface as
  int64)**; **THE LIVE-D1-SCHEMA FINDING (SESSION-081 — the owner's
  `curator-bot` D1 is empty; the schema is created lazily on the
  bot's first write; the fallback's happy path is live-verifiable
  only after the bot processes a URL)**; the standing set carries
  forward from SESSION-076's v77 card (THE REACH-CLOUDFLARE LAW, THE
  SAME-DECODER LAW, THE SIX-FIX RECORD, THE OWNER-DIRECTIVE RECORD,
  THE PARTIAL-WIPE MODE-FLIP CLASS, THE MULTIEDIT-TAB CORRUPTION
  [STILL STANDING — the Edit tool tab-corrupted models.ts file-wide;
  the Python byte-replace rescued it; GO + TS FILES stay
  python-on-bytes for any multi-line edit], THE RELEASE-RECORD, THE
  MENU-SURFACED RECORD, THE ENVIRONMENT-OWNED-TEXT LAW, THE
  BOOTSTRAPPER-LEG RECORD, THE DEBT-BOOKS-EMPTY RECORD, THE
  TEXT-MODE-READ LAW, THE MULTIEDIT-TAB FIFTH STRIKE, THE
  HARVEST-COUSIN RECORD, THE DELIVERED-LAW RECORD, THE LIVE-LEG
  RECORD, THE TAG-STATE BUILD LAW, THE -X-UNDER-TRIMPATH LAW, THE
  VERIFY-THE-VERIFIER LESSON, THE DEFAULT-ICON FIND, THE FLAG-CLEAR
  LAW, THE REFERENCE'S-OWN-ARITHMETIC LAW, the binding-import law,
  the one-owner-per-surface law, the field-tag law, the render-context
  law, the hermetic-seam law, the golden-marshal-order trap, the
  suite-reverification dividend, the wails-build mode flips, the
  context-death continuation recipe, the CD-DISCIPLINE LAW, the
  RENDERER-EATS-ESCAPE-PAIRS lesson [commit messages via -F FILE],
  the WRITE-TOOL REDACTION lesson, the BACKGROUND-PROCESS lesson,
  PUSH AFTER EVERY TASK).

## 2. Next session: SESSION-082 — the post-D1-fallback session (the decision tree, in priority order)

1. **(a) THE OWNER'S GO.11 D1-FALLBACK RE-RUN** — if the prompt
   carries their re-run results: the intake (did the SYNC fall back
   to D1? did the bot queue fetch live? the SQL's column-name match
   against the live schema — surface the exact D1 error if any). If
   the happy path is green → THE D1-API-FALLBACK is end-to-end proven
   on the owner's machine. If the SQL's column names diverge from the
   Worker's actual schema → that's a one-line edit at `d1PendingSQL`
   (the single construction site).
2. **(b) the standing go.4 re-run + the Phase-7 sign-off** — the
   owner's go.4 re-run (the checklist's slots are prepared) + the
   sign-off call (the owner's explicit call; surface it, never assume
   it) are the Phase-7 exit gate's last leg.
3. **(c) the Phase-8 pick** — only when the fix cycle closes (the
   menu stands from SESSION-064).
4. **Blockers**: if anything red surfaces (a CI failure, a broken
   release link, a D1 SQL mismatch), STOP and surface it with
   evidence per the standing laws; record it in the card + report and
   hand the decision back.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-081.md` — the D1-API-fallback session's
   report (the law + the wiring + the six pinning tests + the
   live-D1 verification + the verified-three-ways numbers).
2. `docs/poc-artifacts/release-v0.92.0-go.11.md` — the live D1-API-
   fallback release's report (the verified numbers).
3. `docs/sessions/SESSION-080.md` — NOT committed (SESSION-077–080
   didn't write session reports; the git log carries their one-line
   records). Read the git log for the go.7–go.10 history.
4. `docs/sessions/SESSION-076.md` — the last committed session
   report (the re-verification session).
5. `internal/workerclient/d1fallback.go` + `d1fallback_test.go` —
   the D1-API fallback's seat (the law + the six pinning tests).
6. `0-START-HERE.txt` + `WINDOWS-QUICKSTART.md` — the shipped
   portable-first surface (go.11 refs).
7. `docs/port-blueprint/12-migration-roadmap.md` — the Phase-7
   records + the Phase-8 menu.
8. The full Phase-8 menu's full form stands from SESSION-064's
   report.

## 4. Known constraints / owner actions for SESSION-082

- **OWNER ACTION #1 (the live-D1 happy-path verification, optional)**:
  send a test GitHub URL to the bot (the Worker writes the `ledger`
  row + creates the table lazily on first write); then SYNC — the
  direct leg fails (ISP blocks workers.dev) + the D1 fallback fetches
  the row live (the happy path proven end-to-end). If the SQL's column
  names diverge from the Worker's actual schema, surface the exact D1
  error + the column list (the SQL is a one-line edit at
  `d1PendingSQL`).
- **OWNER ACTION #2 (the Direct Pair UI fill)**: Settings → Direct
  Pair, paste the 3 new optional fields (API Token + Account ID +
  Database ID — the values are in the reference config at
  `/home/z/my-project/public/gitcurator-config.json`), click Direct
  Pair. The fallback is armed.
- **OWNER ACTION #3 (the go.4 re-run + the sign-off)** — the
  Phase-7 exit gate's last leg; unchanged from the v77 card.
- **OWNER ACTION #4 (the Phase-8 pick)** — only when the fix cycle
  closes.
- **OWNER ACTION #5 (the config reference is web-served)** — the
  reference config at `/home/z/my-project/public/gitcurator-config.json`
  is served by the Next.js dev server (the preview panel); treat it
  as ephemeral (delete after copying into the GitCurator install).
- **The repo is PUBLIC** — never commit secrets; pushes still need
  the ephemeral PAT (re-inline per call; scrub the remote after).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law.
- **The webkit tree is DOWN** (the SIXTEENTH rebuild is pending) —
  rebuild it FIRST only if the session runs a LINUX GUI build.
- **GO + TS FILES stay python-on-bytes for any multi-line edit**
  (the Edit tool tab-corrupted models.ts file-wide at SESSION-081;
  the Python byte-replace rescued it; the lesson re-proven).
- The dissolution call (ADR-0002 Am.5); vanilla-vs-vite; the
  batch-scorecard modal; the v0.91.0 zip-parity reading.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged).
- The go-directive acknowledgement (STILL WAITING): the main module
  rides `go 1.24.1` (the tls-client pin, ADR-0002 Am.2); the nested
  webui module rides `go 1.25.0` (do not bump).
- **After every `wails build`**: check `git diff --raw` for mode
  flips; never commit a stray binary (nor a generated notes.md).
- **Deferred test debt on the books**: NONE. The owner's LIVE legs
  remain (the go.11 D1-fallback happy-path re-run + the go.4 re-run
  + the sign-off).
- **Single-writer law**: the fresh session is the ONLY writer.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  **PRIVATE — re-clone with the owner's ephemeral PAT inline, then
  scrub the remote URL**.
- Standing: secrets / owner data never in git (byte-level
  secret-scan before every commit); tests never weakened; reality
  wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-082)

```
You are starting SESSION-082 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE STATE: the desktop release line is LIVE at v0.92.0-go.11 (THE D1-API-FALLBACK
release; SESSION-081 shipped it + verified it THREE ways: go-ci run #170 GREEN +
release run #13 GREEN + the release 426430005 LIVE, the asset 27,452,357 bytes,
sha256 2b3b82ac…3f88, the VERSION member + the GUI ProductVersion + the CLI
buildinfo + the tag all carry one number). THE D1-API FALLBACK closes the
ISP-blocks-workers.dev gap: when the direct Worker request fails (the owner's ISP
blocks workers.dev at the IP level — even the DoH tier can't reach a black-holed
anycast), the honest poll falls back to GetPendingViaD1(), querying the D1 database
directly via api.cloudflare.com (a DIFFERENT edge the ISP does NOT block). The
SQL mirrors the Worker's handleGetPending exactly (ledger LEFT JOIN vault_mirror,
forgotten=0, mirror status NULL/'pending', not decommissioned, not blocked-domain
dead letter, ORDER BY first_seen_at ASC, LIMIT 100); read-only (no ack — the ack
rides the direct path when the network recovers). Six hermetic pinning tests pin
every law. The live-D1 verification confirmed the code's transport + envelope +
SQL-parse (the owner's curator-bot D1 is empty for now — the schema is created
lazily on the bot's first write; the fallback fetches live the moment the bot
processes a URL). The session report: docs/sessions/SESSION-081.md. SESSION-082
is the post-D1-fallback session — the decision tree in priority order: (a) if this
prompt carries the owner's go.11 D1-fallback re-run results: the INTAKE (did the
SYNC fall back to D1? did the bot queue fetch live? the SQL's column-name match
against the live schema — surface the exact D1 error if any); if the happy path
is green, THE D1-API-FALLBACK is end-to-end proven on the owner's machine; if the
SQL's column names diverge from the Worker's actual schema, that's a one-line edit
at d1PendingSQL (the single construction site); (b) the standing go.4 re-run +
the Phase-7 sign-off (the owner's explicit call; surface it, never assume it);
(c) the Phase-8 menu pick — only when the fix cycle closes (the menu stands from
SESSION-064); (d) if anything blocks, STOP and surface it with evidence per the
standing laws; record it in the card + report and hand the decision back.

Then follow docs/SESSION-PROTOCOL.md exactly: bootstrap (clone
https://github.com/assadigit/second-brain-project — the work repo, public; read
README.md, docs/SESSION-PROTOCOL.md, docs/NEXT-SESSION.md [the card v82], the
two most recent session reports [SESSION-081 is the latest; SESSION-076 is the
last committed one before the go.7-go.11 gap — SESSION-077–080 didn't write
reports; the git log carries their one-line records], git log --oneline -15),
orient (the read-firsts: the card's §3), work (the standing laws: reality wins,
tests never weakened, no secrets ever, stay inside the phase, push after every
task, GO + TS FILES stay python-on-bytes for any multi-line edit), close (the
honest gate evaluation, the commit + push, the SESSION-082 report, the
NEXT-SESSION.md rewrite, the worklog append), and report in chat with the literal
next prompt.
```
