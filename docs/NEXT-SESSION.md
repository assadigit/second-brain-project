# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker
  HTTP API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity (constants, links, storage+dryrun,
  prompts, taxonomy, notebuilder, config, helpers).
- **Phase 2 CLOSED** (SESSION-006–010; all six tasks): T-20 `internal/
  store`, T-21 `internal/notestate`, T-22 `internal/notereads`, T-23
  `internal/websitepipeline/state.go`, T-24 `internal/store/repos.go`,
  T-25 `internal/batch` — 425 tests race-clean at the close.
- **Phase 3 OPEN (SESSION-011–012)**: **T-30 `internal/llmclient`
  landed (`f2073b5`; SESSION-011) and T-31 `internal/webfetch` landed
  (`4d93877`; SESSION-012) — the whole fetch ladder: the routes
  (socks5 rdns via x/net/proxy + hand-rolled socks4a, http-proxy
  absolute-URI, loopback/direct never proxied), the OWN redirect loop
  (hop cookies, 308, loop naming, the no-Location verdict, the
  Referer law), the Chrome-131 header set, the rate limiter with
  penalties, the caps + capped decompression, the failure taxonomy,
  DoH, the dead-page rescue + the Wayback courtesy, and the third
  door (tls-client v1.16.0 — ships armed). 676/676 tests race-clean;
  the THIRTEENTH golden generator (`testdiff/goldens/dump_webfetch.py`
  — pure + payload + behavior goldens with the
  `<ROOT>`/`<DOH>`/`<ARCHIVE>`/`<PROXY>`/`<DEADPORT>` normalization);
  the goldens caught four real bugs before landing.** The environment
  was fully wiped AGAIN at the session start (the seventh full reset)
  and rebuilt per the §1 recipes before work.
- **SESSION-012's same-commit reality correction**: the main module's
  `go` directive is now **1.24.1**, forced by the owner-acknowledged
  tls-client v1.16.0 pin (its own go.mod declares 1.24.1; the Go
  toolchain enforces the module-graph rule). Recorded in ADR-0002
  Am.2 + TOOLCHAIN.md Am.1 with the rollback path;
  `modernc.org/sqlite` stays exactly v1.36.1; x/net v0.48.0 and
  x/text v0.32.0 are direct requires (ADR-0006 §1 names x/net/proxy).
  **Owner acknowledgement requested** (see §4).
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — T-31 was pushed the moment its battery went green.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…012 (seven so far) PLUS
  SESSION-010's mid-session mode-flip event. The restore recipes:
  start-of-session (clone via the owner's in-chat PAT with `git -c
  http.extraheader="Authorization: Basic $(printf
  'x-access-token:%s' "$PAT" | base64)"` — the Basic scheme; the
  token scheme gets 401; never store it), GitCurator @ `98f4fc7`
  (verify against `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 →
  `/home/z/.local/go` (go.dev tarball), venv `/home/z/venvs/
  gitcurator` (Python 3.12, stdlib-only), symlinks
  `/home/z/second-brain-project` and `/home/z/GitCurator`; mid-session
  mode-flip (`git checkout-index -f -a` restores modes from the
  index; `git fsck` to sanity-check). Verify the pins at every start:
  `git rev-parse HEAD:go.mod` — expect the v1.36.1/tls-client/1.24.1
  bytes (do NOT bump sqlite; do NOT bump tls-client past v1.16.0).

## 2. Next session: SESSION-013 — Phase 3 continues: T-32 webextract (S) with the owner's verdict, then T-33 connectioncheck (M) if capacity remains

- **T-32 `internal/webextract`** (S): port `core/web_extract.py` per
  `spec-core-web-extract.md` — the parser + heuristics (title/meta/
  text extraction, the JS-shell and paywall detectors, the 4000-char
  text cap, the charset re-decode from the page's own `<meta
  charset>`). Acceptance: the 5 HTML fixtures byte-compared fields
  (the reference's `app/tests/fixtures/web/` set: article, landing,
  js_shell, paywall, latin1 — drive them through the generator for
  the goldens, the llmclient/webfetch pattern). **Read the spec
  BEFORE the code; check §4's dependency verdict first.**
- **T-33 `internal/connectioncheck` + sentinel** (M): per
  `spec-core-connection` — battery shapes, the UA table, the offline
  verdict. Acceptance: result-shape tests + the UA table asserted.
- **T-34 `internal/workerclient`** (M) is the gate's last stone (the
  HMAC loopback rig ported from test_synclive) — likely its own
  session or the second half of SESSION-013.
- Working examples to imitate: `internal/webfetch` (SESSION-012's
  port) is now the freshest Phase-3 pattern — the three-golden
  generator, the wire recorder, the seam save/restore rig, doc.go's
  divergence records; `internal/llmclient` for the LLM-side shapes.
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; atomic saves; the dry-run seam at call sites; rows
  collected before nested statements; tests never weakened; reality
  wins over specs (fix the spec in the same commit); goldens are
  GENERATED by running the unmodified Python app, never hand-typed;
  no per-run facts in goldens.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-012.md` — the most recent report (the T-31
   evidence, the seventh wipe + rebuild, the four goldens-caught-bugs
   story, the go-directive correction).
2. `docs/port-blueprint/module-specs/spec-core-web-extract.md` —
   BEFORE T-32's code (and its Go-notes dependency line — see §4).
3. The reference: `app/gitcurator/core/web_extract.py` + its test
   legs (`test_phase2.py`'s TestWebExtract + the fixtures under
   `app/tests/fixtures/web/`).
4. `docs/port-blueprint/module-specs/spec-core-connection.md` —
   before T-33 (grep its test legs: `test_connection.py` +
   `test_netguard.py`).
5. Working example: `internal/webfetch/` (the generator, the fakes
   rig, the golden drivers, doc.go).

## 4. Known constraints / owner actions for SESSION-013

- **The T-32 dependency verdict (STILL WAITING — decide FIRST)**:
  `spec-core-web-extract.md`'s Go notes name `golang.org/x/net/html`
  — a new dependency with no ADR. Options: (a) the owner approves an
  ADR adding x/net/html (pinned) — the pragmatic choice; or (b) T-32
  hand-rolls the reference's stdlib `html.parser` behavior (larger
  effort, zero dependencies). **NEW FACT since SESSION-011's
  surface: the x/net MODULE is now already in go.mod's graph
  (tls-client requires it; v0.48.0) — option (a) adds NO new module,
  only a new package import.** If the session prompt carries the
  verdict, honor it; if not, ask early.
- **The go-directive acknowledgement**: the main module rides
  `go 1.24.1` (tls-client v1.16.0's own floor — ADR-0002 Am.2 has
  the record + the rollback path). A one-line acknowledgement (or a
  rollback instruction) closes it.
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it; goldens are GENERATED by
  running its unmodified code (`PYTHONDONTWRITEBYTECODE=1` keeps it
  pristine).
- **Dependency law**: `modernc.org/sqlite` pinned at v1.36.1;
  `bogdanfinn/tls-client` pinned at v1.16.0; `golang.org/x/net`
  v0.48.0 + `golang.org/x/text` v0.32.0 (direct, per ADR-0006 §1's
  named dialer + the charset ladder); ANY other add needs a new ADR —
  surface, don't improvise.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the note_state SURGERY
  path's CRLF round-trip differential (Phase 3+); the Windows live
  legs (Phase 7 clean-machine checklist); a Windows VM without
  WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-013)

```
You are starting SESSION-013 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (full wipes at the starts of
   SESSION-005 through SESSION-012, plus a MID-session degradation in
   SESSION-010). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-project),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 → /home/z/.local/go, venv /home/z/venvs/gitcurator (Python 3.12,
   stdlib-only) per build/TOOLCHAIN.md, and go.mod with modernc.org/sqlite at
   v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 +
   golang.org/x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these). Then
   check `git status` and `git log --oneline -5`: SESSION-012 landed T-31
   (commit 4d93877; the tree must be clean and origin/main at the SESSION-012
   close). If anything is missing, rebuild per docs/NEXT-SESSION.md §1 (clone
   via the PAT I give you when you ask — pass it ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it; a mid-session mode-flip is
   repaired with `git checkout-index -f -a` after `git fsck`) and re-verify the
   suite (676 green) before work. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 3 OPEN
   state with T-30 + T-31 done at 676 tests race-clean, the seventh full wipe +
   rebuild story, the go-directive correction, and the environment recipes) →
   docs/sessions/SESSION-012.md (the most recent report).

3. THIS SESSION IS T-32 + (if capacity remains) T-33: first T-32
   internal/webextract (S) — port core/web_extract.py per
   docs/port-blueprint/module-specs/spec-core-web-extract.md (READ THE SPEC
   BEFORE THE CODE): the parser + heuristics, the JS-shell and paywall
   detectors, the text cap, the charset re-decode. Acceptance: the 5 HTML
   fixtures byte-compared fields — generate the goldens from the reference's
   own fixtures via a fourteenth generator (the dump_llmclient/dump_webfetch
   pattern; the fixtures live in the reference's app/tests/fixtures/web/).
   CHECK THE DEPENDENCY VERDICT FIRST (NEXT-SESSION §4): if the owner's
   verdict on golang.org/x/net/html is in the prompt, honor it; if not, ask
   before adding anything (note: the x/net module is already in go.mod's
   graph — an ADR would add no new module, only an import; a hand-rolled
   stdlib port is the zero-dependency alternative). Then, if capacity
   remains, T-33 internal/connectioncheck + the sentinel (M) per
   spec-core-connection.md: the battery shapes, the UA table, the offline
   verdict — result-shape tests + the UA table asserted. Standing port laws
   carry: the CAST read law where store-adjacent columns appear,
   store.PyISOFormat* timestamp shapes, atomic saves, the dry-run seam at call
   sites, rows collected before nested statements, no per-run facts in
   goldens. internal/webfetch (SESSION-012) is the working example.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do
   not bump any); surface blockers with evidence (hash-level, not display)
   instead of improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase
   3's exit gate: fake-server contract tests green — T-30's payload goldens
   and T-31's ladder tests are laid; T-32's fixture parity and T-34's HMAC
   loopback rig remain), commit + push everything (ask me for the GitHub PAT
   if it is not already available — never store it anywhere), write
   docs/sessions/SESSION-013.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-014.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
