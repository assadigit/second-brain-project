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
- **Phase 3 CLOSED (SESSION-011–014; all five tasks)**: T-30
  `internal/llmclient` (`f2073b5`), T-31 `internal/webfetch`
  (`4d93877`), T-32 `internal/webextract` (`833b21b`), T-33
  `internal/connectioncheck` (`e5846bd`), T-34 `internal/workerclient`
  (`6ffc4e9`) — 872 tests race-clean at the close, the exit gate's
  five stones all laid.
- **Phase 4 OPENED (SESSION-015)**: **T-40 `internal/websitepipeline`
  landed at `8098ac4`** — the per-link state machine on top of T-23's
  state layer: the full pipeline (the note builders, the
  consolidation, the review scans, the master table's grammar +
  writers + scans, THE BANISHMENT + its confirmation gate + the
  tally, the settlement, the splits, the four drivers), the
  `decomtable` grammar floor, the handdelivery consumption surface
  (T-43 completes that module), the directorysync banish-gate
  helpers (T-49 completes that module), and THE DIFFERENTIAL GOLDEN
  (51 sections, frozen clocks, generated from the unmodified
  reference by `testdiff/goldens/dump_websitepipeline.py`).
  **910/910 tests race-clean; T-40's acceptance (state-machine table
  tests; banishment flows confirmed/declined/timeout; a differential
  golden-set case) — all three stones PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — T-40 was pushed the moment its battery went green.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…015 (**ten so far**) PLUS
  SESSION-010's mid-session mode-flip event. The restore recipes:
  start-of-session (discover the remote via the GitHub API with the
  owner's in-chat PAT passed ephemerally as the Basic
  `x-access-token:<PAT>` header — SESSION-013/014/015 all had to do
  this from a bare sandbox; clone `assadigit/
  second-brain-project` with `git -c http.extraheader="Authorization:
  Basic $(printf 'x-access-token:%s' "$PAT" | base64)"` — the Basic
  scheme; the token scheme gets 401; never store it), GitCurator @
  `98f4fc7` (verify against `testdiff/PINNED-GITCURATOR-REV`),
  go1.27.2 → `/home/z/.local/go` (go.dev tarball), venv `/home/z/
  venvs/gitcurator` (Python 3.12, stdlib-only), symlinks
  `/home/z/second-brain-project` and `/home/z/GitCurator`;
  mid-session mode-flip (`git checkout-index -f -a` restores modes
  from the index; `git fsck` to sanity-check). Verify the pins at
  every start: `git rev-parse HEAD:go.mod` — expect the
  v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do NOT bump
  tls-client past v1.16.0). **SESSION-015's debugging lesson, kept
  for the next sessions: this sandbox's output renderer EATS the
  two-character sequence `[h` (a console escape) — verify suspicious
  "corruption" with %q byte dumps before believing rendered text.**

## 2. Next session: SESSION-016 — Phase 4 continues: T-41 directorynote (M), then T-42 githubpipeline + batch (L) if capacity remains

- **T-41 `internal/websitepipeline/directorynote.go`** (M): the
  Website Directory regeneration — `sweep_banned_notes` (the v0.28.0
  disk sweep whose seam T-40 documented: app-owned notes on banned
  domains → `.trash/banned-domains/`; platform `_inbox` tables
  quarantined when all platform domains banned) + `build_website_
  directory` (the `000 📚 Website Directory.md` regeneration — the
  frontmatter, the info callout with counts, the Cat/Sub headings
  with counts, the entry lines with the ≤80 alias / ≤140 TL;DR caps,
  the `" — "` joins, the `(title.lower(), relpath.lower())` sort, the
  taxonomy order + leftovers, the `_review` todo callout, the closing
  regeneration line, the 512,000-byte read cap, the walk's skip
  set). Spec: `spec-core-website-directory.md` (READ THE SPEC BEFORE
  THE CODE). Acceptance: **output golden byte-compare** (generate by
  running the unmodified Python reference through the venv — the
  house generator pattern; no per-run facts inside).
- **T-42 `internal/githubpipeline` + `internal/batch`** (L) if
  capacity remains: the per-repo loop, the banner thread, the
  reports, the undo file. Spec: `spec-gui-processing.md`. Acceptance:
  batch summary goldens; 404-strike tests; rate-limit law tests.
- The Phase 4 queue behind them: T-43 handdelivery (completes the
  module T-40 floored), T-44 chrometabs, T-45 vaultscan, T-46
  vaulthygiene + repobanish, T-47 vaultseal, T-48 goodrepos +
  reporadar, T-49 directorysync (REUSES T-34's rig directory doors;
  completes the module T-40 floored), T-50 mirror, T-51 telegram
  (dormant per ADR-0004 Am.3 — the register's Phase-8 door).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for any
  store-adjacent column; `store.PyISOFormat*` for timestamp writes;
  atomic saves; the dry-run seam at call sites; rows collected before
  nested statements; tests never weakened; reality wins over specs
  (fix the spec in the same commit); goldens are GENERATED by running
  the unmodified Python app, never hand-typed; no per-run facts in
  goldens; the frozen-clock seam (`clockNow`/`handdelivery.Clock`)
  exists for exactly this.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-015.md` — the most recent report (the T-40
   evidence, the tenth wipe + rebuild, the layering record, the
   `[h` renderer lesson).
2. `docs/port-blueprint/module-specs/spec-core-website-directory.md`
   — BEFORE T-41's code (the directory-note laws, the sweep, the
   caps, the walk).
3. The reference: `app/gitcurator/core/website_directory.py` (432
   LOC — the whole module; T-41 is its first port) + the test legs
   (`test_directorynote*`/the directory selections of test_phase*).
4. Working examples: `internal/websitepipeline` itself (T-40 — the
   golden-generator pattern with frozen clocks, the walkVault
   helper, the gnote/gwriteNote fixture factories) and
   `internal/batch` (T-25's VaultIndex — the site-map/walk laws the
   directory regen shares).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-016

- **The go-directive acknowledgement (STILL WAITING — one line closes
  it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers x/net/html,
  approve an ADR and a future session swaps
  `internal/webextract/parser.go`'s scan internals).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it; goldens are GENERATED by
  running its unmodified code (`PYTHONDONTWRITEBYTECODE=1` keeps it
  pristine).
- **Dependency law**: `modernc.org/sqlite` pinned at v1.36.1;
  `bogdanfinn/tls-client` pinned at v1.16.0; `golang.org/x/net`
  v0.48.0 + `golang.org/x/text` v0.32.0 (direct); ANY other add needs
  a new ADR — surface, don't improvise.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the note_state SURGERY
  path's CRLF round-trip differential (Phase 4+); the Windows live
  legs (Phase 7 clean-machine checklist); a Windows VM without
  WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-016)

```
You are starting SESSION-016 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (full wipes at the starts of
   SESSION-005 through SESSION-015 — ten so far — plus a MID-session
   degradation in SESSION-010; SESSION-013/014/015 all started from a BARE
   sandbox and had to discover the remote via the GitHub API with the PAT
   before cloning). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked
   /home/z/second-brain-project), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator; must
   match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/go,
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
   build/TOOLCHAIN.md, and go.mod with modernc.org/sqlite at v1.36.1,
   bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/
   x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these). Then
   check `git status` and `git log --oneline -3`: SESSION-015 opened Phase 4
   (T-40 websitepipeline at 8098ac4; the tree must be clean and origin/main at
   the SESSION-015 close). If anything is missing, rebuild per
   docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass it
   ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it; a mid-session mode-flip is
   repaired with `git checkout-index -f -a` after `git fsck`) and re-verify the
   suite (910 green) before work. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4
   OPENED state with T-40 done at 910 tests race-clean, the tenth full wipe +
   bare-sandbox rebuild story, the environment recipes, and the `[h` output-
   renderer lesson) → docs/sessions/SESSION-015.md (the most recent report).

3. THIS SESSION IS T-41 (M): Phase 4 continues —
   internal/websitepipeline/directorynote.go, the Website Directory
   regeneration. READ docs/port-blueprint/module-specs/spec-core-website-
   directory.md BEFORE THE CODE, then port core/website_directory.py (432
   LOC): sweep_banned_notes (the v0.28.0 disk sweep whose seam T-40 documented
   in the pipeline — app-owned notes on banned domains → .trash/banned-
   domains/, platform _inbox tables quarantined when all platform domains are
   banned) + build_website_directory (regenerates `000 📚 Website Directory.md`
   — the frontmatter managed_by/kind: directory/generated, the # 📚 heading,
   the info callout with counts, the ## Cat (n) / ### Sub (n) headings, the
   entries `- [[path|alias ≤80]] — TL;DR ≤140 — pricing — [↗](url)` joined
   " — ", sorted (title.lower(), relpath.lower()), taxonomy order then
   leftovers, the _review todo callout, the closing regeneration line, the
   512,000-byte read cap, the walk skipping _inbox/_review/_missing/_moc/
   attachments/.obsidian/.trash/.git). Acceptance: an OUTPUT GOLDEN
   BYTE-COMPARE — generate it by running the unmodified Python reference
   through the venv (the house generator pattern: testdiff/goldens/
   dump_websitepipeline.py is T-40's working example, frozen clocks via the
   clockNow seam, no per-run facts inside). If capacity remains after T-41
   goes green + pushes: T-42 internal/githubpipeline + batch (L) per
   spec-gui-processing.md — but close cleanly at a natural boundary if
   capacity runs out; one task done properly with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do
   not bump any; x/net/html stays OUT unless the owner approves an ADR);
   surface blockers with evidence (hash-level, not display) instead of
   improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-41's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-016.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-017.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```
