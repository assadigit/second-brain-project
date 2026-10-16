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
  store` (the CAST read law), T-21 `internal/notestate`, T-22
  `internal/notereads` (THE ONE-OPEN LAW), T-23 `internal/
  websitepipeline/state.go`, T-24 `internal/store/repos.go`, T-25
  `internal/batch` — 425 tests race-clean at the close.
- **Phase 3 OPEN (SESSION-011)**: **T-30 `internal/llmclient` landed
  (commit `f2073b5`) — the three providers (Ollama HTTP wire,
  OpenAI-compatible, Anthropic Messages) on Python-json.dumps-identical
  WIRE BYTES (the payload goldens — the first stone of the Phase-3
  exit gate), the JSON-mode memoized 400-fallback, extract_json, the
  llama.cpp detection engine (the /props grab, the fingerprints, the
  process-port scan) and the autodetect policy. 556/556 tests
  race-clean; the TWELFTH golden generator
  (`testdiff/goldens/dump_llmclient.py` — pure corpus + payload
  goldens + behavior golden with `<ROOT>`/`<API>`/`<DEADURL>`
  normalization); `storage.PyDumpsCompact` + `DecodeJSONValue`
  promoted (the wire-bytes law's encoder).** The environment was
  fully wiped AGAIN at the session start (the sixth full reset) and
  rebuilt per the §1 recipes before work.
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — T-30 was pushed the moment its battery went green.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…011 (six so far) PLUS
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
  index; `git fsck` to sanity-check). Verify the pin at every start
  (`git rev-parse HEAD:go.mod` — expect blob
  `6721ebb8766573ec2209e6f536755822a905cee8`, the v1.36.1/1.22
  bytes; do NOT bump it).

## 2. Next session: SESSION-012 — Phase 3 continues: T-31 webfetch (XL — the whole session's task)

- **T-31 is the session**: `internal/webfetch` — port
  `core/web_fetch.py` (routes: proxy/direct/loopback; the redirect
  ladder; headers; the rate limiter; caps; decompression; failure
  categories; DoH; the wayback machine; the impersonation door). Spec:
  `spec-core-web-fetch.md` — read it BEFORE the code. Effort XL: the
  roadmap and SESSION-011's close both assign it the whole session;
  do NOT start it with leftover capacity and do not expect to land
  anything else after it unless it finishes early.
- **T-31's dependency**: tls-client is ALREADY authorized by
  ADR-0006 — **read `docs/port-blueprint/decisions/0006-network-
  stack.md` §2 for the exact pin BEFORE `go get`**; add it to go.mod
  pinned at that version (this is the ONE Phase-3 dependency add
  that needs no new owner decision). modernc.org/sqlite stays at
  v1.36.1 — do not `go get -u` anything.
- **T-32 (webextract, S) — WAITING ON AN OWNER DECISION** (recorded
  in SESSION-011 §Blockers): its module spec's Go notes name
  `golang.org/x/net/html`, a NEW dependency with no ADR. Options:
  (a) the owner approves an ADR adding x/net/html (pinned) — the
  pragmatic choice; or (b) T-32 hand-rolls the reference's
  stdlib `html.parser` behavior (the reference itself is
  stdlib-only). If the session prompt carries the verdict, honor it;
  if not, ask in chat and surface it again rather than improvising.
- Working examples to imitate: **`internal/llmclient` is now the
  Phase-3 pattern** — the fake-server rig on real 127.0.0.1
  net/http listeners (fakes_test.go), the recording golden generator
  (dump_llmclient.py), the wire-bytes law (storage.PyDumpsCompact
  for every request body the golden byte-pins), the per-run-fact
  normalization (`<ROOT>`/`<API>`/`<DEADURL>`), the tests-first
  file layout, and doc.go recording every divergence.
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; atomic saves; the dry-run seam at call sites; rows
  collected before nested statements; tests never weakened; reality
  wins over specs (fix the spec in the same commit); goldens are
  GENERATED by running the unmodified Python app, never hand-typed.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-011.md` — the most recent report (the T-30
   evidence, the sixth wipe + rebuild, the goldens-caught-bugs story,
   the T-32 blocker).
2. `docs/port-blueprint/module-specs/spec-core-web-fetch.md` —
   BEFORE T-31's code.
3. `docs/port-blueprint/decisions/0006-network-stack.md` — §2's
   tls-client pin (T-31's only pre-authorized dependency add).
4. The reference: `app/gitcurator/core/web_fetch.py` + its tests
   (`test_netguard.py`, `test_webproxy.py`, and any web_fetch legs
   in test_phase*/test_e2e — grep first).
5. Working example: `internal/llmclient/` (fakes_test.go's rig,
   golden_test.go's drivers, dump_llmclient.py's generator shape).

## 4. Known constraints / owner actions for SESSION-012

- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- **The T-32 verdict** (§2) — an owner one-liner in the prompt
  resolves it; otherwise ask early so the decision doesn't block a
  later session.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it; goldens are GENERATED by
  running its unmodified code (`PYTHONDONTWRITEBYTECODE=1` keeps it
  pristine).
- **Dependency law**: modernc.org/sqlite pinned at v1.36.1 (the go
  1.22 floor); tls-client at ADR-0006 §2's pin (T-31 only); ANY
  other add needs a new ADR — surface, don't improvise.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display — only blob hashes
  and `od -c` are truth).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the note_state SURGERY
  path's CRLF round-trip differential (Phase 3+); the Windows live
  legs (Phase 7 clean-machine checklist); a Windows VM without
  WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-012)

```
You are starting SESSION-012 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (full wipes at the starts of
   SESSION-005 through SESSION-011, plus a MID-session degradation in
   SESSION-010). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-project),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 → /home/z/.local/go, venv /home/z/venvs/gitcurator (Python 3.12,
   stdlib-only) per build/TOOLCHAIN.md, and modernc.org/sqlite pinned at v1.36.1
   in go.mod with the go 1.22 floor (verify the go.mod CONTENT — `git rev-parse
   HEAD:go.mod`; at SESSION-011's close the blob was
   6721ebb8766573ec2209e6f536755822a905cee8, the
   v1.36.1/1.22 bytes; do not upgrade it). Then check `git status` and
   `git log --oneline -5`: SESSION-011 opened Phase 3 (T-30 = commit f2073b5;
   the tree must be clean and origin/main at f2073b5). If anything is missing,
   rebuild per docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you
   ask — pass it ephemerally with git -c http.extraheader="Authorization: Basic
   $(printf 'x-access-token:%s' "$PAT" | base64)", the Basic scheme; never
   store it; a mid-session mode-flip is repaired with `git checkout-index -f
   -a` after `git fsck`) and re-verify the suite (556 green) before work.
   PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 3 OPEN
   state with T-30 done at 556 tests race-clean, the sixth full wipe + rebuild
   story, and the environment recipes) → docs/sessions/SESSION-011.md (the most
   recent report).

3. THIS SESSION IS T-31: internal/webfetch (XL) — port core/web_fetch.py per
   docs/port-blueprint/module-specs/spec-core-web-fetch.md (READ THE SPEC
   BEFORE THE CODE): the route ladder (proxy/direct/loopback), the redirect
   ladder, the header set, the rate limiter, the size/time caps, decompression,
   the failure categories, DoH, the wayback machine, and the impersonation
   door. tls-client is ALREADY authorized by ADR-0006 — read
   docs/port-blueprint/decisions/0006-network-stack.md §2 for the exact pin
   BEFORE `go get`, pin it at that version, and touch NOTHING else in go.mod
   (sqlite stays v1.36.1). Port the Python characterization tests FIRST, then
   the code; generate goldens by running the unmodified Python app
   (testdiff/goldens/ — twelve generators exist; dump_llmclient.py is the
   newest pattern: recording fake servers, per-run-fact hygiene with
   <ROOT>/<API>/<DEADURL> normalization, wire bytes pinned via
   storage.PyDumpsCompact). internal/llmclient (SESSION-011's port) is the
   working example to imitate — its fakes_test.go rig, its golden drivers,
   its doc.go divergence records. T-31 is XL: it owns the whole session; if it
   finishes early with capacity left, T-32 (webextract, S) is next ONLY IF the
   owner's dependency verdict is in hand (spec-core-web-extract names
   golang.org/x/net/html — a new dependency needing an ADR or a hand-rolled
   stdlib port; surface the question, do not improvise). Standing port laws
   carry: the CAST read law where store-adjacent columns appear,
   store.PyISOFormat* timestamp shapes, atomic saves, the dry-run seam at call
   sites, rows collected before nested statements, no per-run facts in goldens.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (modernc.org/sqlite v1.36.1 and T-31's
   tls-client pin are the already-authorized set — do not bump either);
   surface blockers with evidence (hash-level, not display) instead of
   improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase
   3's exit gate: fake-server contract tests green — T-30's payload goldens
   are its first stone; T-31's ladder/redirect/impersonation tests are its
   next), commit + push everything (ask me for the GitHub PAT if it is not
   already available — never store it anywhere), write
   docs/sessions/SESSION-012.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-013.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
