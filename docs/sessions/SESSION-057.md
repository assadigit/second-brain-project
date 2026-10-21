# SESSION-057 — 2026-10-21 — Phase 7: the Phase-7 finish, owner-input driven — the decision tree ran path (c) a fifth time (no T-82 results, no "ship it"): the menu surfaced; the 52nd wipe absorbed; the battery re-verified green (2377 + 89, the exact committed shape); the go.dev-listing recipe correction (the toolchain double-verify's archive side)

## Goal (from NEXT-SESSION.md at start)

SESSION-057 = the Phase-7 finish, OWNER-INPUT DRIVEN — the decision
tree, in priority order: (a) the owner brings T-82 results → the
intake into the checklist scaffold + doc 13's flip → COMPLETE + the
Phase-7 exit evaluation; (b) the owner says "ship it" → the milestone
release session (VERSION bump, the tag law, the proven workflow, the
AC-23 release report); (c) otherwise → surface the menu honestly and
let the owner pick (Phase-8 scoping from doc 12's register / small
polish from the deferred-debts list). One slice done properly with
proof beats two half-done; close cleanly at a natural boundary when
capacity runs out.

## What was done (task IDs, files, commits — cite hashes)

- **The 52nd wipe absorbed before work** (the environment fully down
  again: both repos, the toolchain, the venv, the wails CLI, the
  module caches — /home/z held only the sandbox's Next.js scaffold).
  Rebuilt per the committed recipes: the work repo re-cloned
  **anonymously** (`assadigit/second-brain-project`, public; HEAD
  `3ae2443` — the SESSION-056 close, tree clean, `git log --oneline`
  showing `3ae2443` + `5014d0b` + `8d05c58`, exactly the expected
  state; the anonymous clone carries no credentials by construction),
  the reference re-cloned via the **ephemeral GIT_CONFIG_COUNT
  env-var PAT form** (zero persistence by construction — the remote
  URL stays the clean anonymous form) then **byte-scanned**
  (`assadigit/GitCurator` checked out at
  `98f4fc71cfb862cee7225d864636c08062724348` — byte-equal to
  `testdiff/PINNED-GITCURATOR-REV`, the full 40-char pinned SHA; the
  post-clone scan for the token over the whole `.git` tree AND the
  worktree: zero hits; the extraheader pattern: zero hits), go1.27.2
  downloaded from go.dev (the `-L` redirect follow; 70,590,635 bytes —
  not the 75-byte stub) + sha-verified **twice** — against
  TOOLCHAIN.md Am.2's committed pin AND against go.dev's published
  listing (programmatic compares, 64 hex chars on all three sides,
  MATCH ×2), the venv `/home/z/venvs/gitcurator` (plain form;
  ensurepip present, Python 3.12.15), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install …wails@v2.16.0`; `wails
  version` → v2.16.0). **No webkit rebuild** (no Linux GUI build this
  session — the full battery runs without it, re-proven again; the
  pending rebuild stays the SIXTEENTH, last actual rebuild the
  fifteenth at SESSION-047 — the corrected ordinal held).
- **THE SESSION'S ONE SLICE (docs-only, the reality-wins class): the
  go.dev-listing recipe correction.** The toolchain sha256's SECOND
  verify side — go.dev's published listing — no longer carries
  go1.27.2 on the stable `?mode=json` endpoint: **go1.27.3 has
  shipped and is the current stable**, so 1.27.2 dropped off the
  two-line stable listing. The correction, settled byte-level before
  anything was "fixed": the archive form
  `https://go.dev/dl/?mode=json&include=all` DOES carry go1.27.2
  linux/amd64 with the published sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
  — programmatic compare against BOTH the locally-computed hash AND
  the committed pin (parsed from TOOLCHAIN.md's own bytes, backtick
  strip): all three sides agree, 64 hex, MATCH ×2. **The committed
  pin stays the authority** (it matched first, before the listing was
  consulted); the listing is the second side, now via the archive
  form. Recorded here + in the state card v58's recipe text + doc
  12's SESSION-057 entry + the README's entry — the same commit
  (reality wins; fix the spec in the same commit).
- **go-ci verified GREEN at the open**: run `39583012792` on
  `3ae2443` (SESSION-056's docs-only close push — the latest push) —
  completed / success. The runs below it on the ledger
  (`39580153947` at `5014d0b`, `39573753456` at `8d05c58`,
  `39529633876` at `35eb065`, `39527338331` at `28958bd`,
  `39526755160` at `44a8bdb`, `39522756047` at `5325623`,
  `39522289819` at `e58b7e4`) all green. CI hygiene holds.
- **The battery re-verified green pre-work — the exact committed
  shape**: main plain **2377 top-level PASS over 36 ok packages, 0
  FAILs, 0 skips** (`go test -count=1 -v ./...`, exit 0; counts from
  the verbose log's `--- PASS`/`--- FAIL`/`--- SKIP`/`ok ` lines);
  main race **36 ok packages, 0 FAILs, 0 data races** (`go test -race
  -count=1 ./...`, exit 0); webui **89 top-level PASS over 4 ok
  packages, 0 FAILs** (its own cd); `gofmt -l` empty in both modules;
  `go vet ./...` + `go build ./...` clean in both modules. The tree
  stayed clean through all legs (no stray artifacts).
- **The decision tree ran path (c) a fifth time**: the owner brought
  no T-82 results and no "ship it" (the checklist slots in
  `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1 verified still
  empty — all nine rows carry the `____` placeholder in RESULT/DATE/
  NOTES, parsed cell-by-cell; the live release list checked via the
  API: only `v0.92.0-go.1`, published, with the single CLI-milestone
  zip asset `GitCurator-v0.92.0-go.1-windows.zip` [15,988,399 bytes] —
  the honest-zip state holds; no release command arrived. The
  anonymous listing API rate-limited at that moment, so the one
  release-list GET rode the ephemeral PAT header — a read-only call,
  nothing persisted, the token never touched a file). The menu is
  surfaced in this report (§Owner actions), the state card v58, and
  the chat report. No autonomous Phase-7 item remains (the AC-17
  benchmark — the last one — closed at SESSION-052), and Phase 8 is
  post-parity, out of blueprint scope (doc 12's Phase-8 register) —
  entering it is the owner's strategic pick, not a session's
  unilateral move; the same holds for the deferred-debts polish list
  (parity report §9 — evidence-recorded, owner-prioritized). The
  session therefore closed cleanly at its natural boundary: the
  honest menu + a verified-clean handoff, with zero implementation
  rows moved (none exists to move autonomously).
- **A second docs-only micro-correction, the same arithmetic-slip
  class as SESSION-055's**: the state card's go-directive
  acknowledgement had drifted its "Nth session carrying it" counter
  (the v57 card said "forty-seventh"; a recount of the records shows
  the acknowledgement has ridden every card since the ADR-0002 Am.2
  record landed at SESSION-012 — 39 session reports carry the phrase
  outright, SESSION-012 through SESSION-050). Settled the fragile
  counter by REPLACING it with the anchored form in the v58 card
  ("carried in every card since the ADR-0002 Am.2 record landed at
  SESSION-012") — countable from the records, immune to the drift;
  the substance is unchanged (the acknowledgement is still the
  owner's standing call).
- **The pinned states re-verified for the record** (the "re-verify
  the suite before work if anything looks off" law, applied
  routinely): `VERSION` = `0.92.0-go.1` (byte-level, 12 bytes + LF —
  untouched); the MAIN `go.mod` exactly the authorized runtime set
  (`tls-client v1.16.0`, `x/net v0.48.0`, `x/text v0.32.0`,
  `modernc.org/sqlite v1.36.1`; the `go 1.24.1` directive) —
  byte-unchanged (the clean fresh clone is the proof); the NESTED
  webui module rides `go 1.25.0` + `require wails/v2 v2.16.0` + the
  parent replace (ADR-0002 Am.5).
- **The docs (this close commit)**:
  `docs/sessions/SESSION-057.md` (this report),
  `docs/NEXT-SESSION.md` (the state card v58 + the SESSION-058
  prompt), doc 12's SESSION-057 tail entry, the README's SESSION-057
  status entry. No Go, workflow, or asset bytes touched (docs-only
  session — the diff proves it).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**The chosen path's own acceptance (path (c): "surface the menu
honestly and let the owner pick"): PASS.** The menu is surfaced with
every option's current state and its evidence pointers (T-82's
checklist doc — slots verified still empty, all nine §1 rows; the
proven release workflow `t-81-release-workflow.md` §5; doc 12's
Phase-8 register; the parity report §9's deferred debts). Nothing
was picked on the owner's behalf; no Phase-8 or polish work was
started unilaterally.

**The re-verification gate (the open-of-session law after a wipe):
PASS** — the battery numbers above, exactly the committed shape
(2377 + 89, the same numbers go-ci proves on every push; the counts
taken from the verbose logs with exit codes recorded: 2377/0/0 over
36 ok; race 36 ok / 0 races; webui 89/0/0 over 4 ok).

**The recipe-correction slice: PASS** — the go.dev-listing change
was settled at byte level (three-side programmatic compare, MATCH
×2) before any doc was touched, and the correction rides this
session's own close commit (the reality-wins law). The toolchain
itself is exactly the pinned one (the pin matched first).

**Phase 7's state is unchanged by design**: T-80 ✓, T-81 ✓, T-83 ✓,
T-82 PARTIAL (items 1–8 the owner's live Windows legs; the checklist
slots still empty — no intake this session, recorded honestly; item 9
automated + green since SESSION-050). The parity report has NO open
implementation row and NO open performance row. What remains in
Phase 7 is the OWNER's: T-82's live checklist + the milestone
release decision.

## Spec updates made (file → what changed, why)

- `docs/sessions/SESSION-057.md` — NEW (this report; the protocol's
  required close record).
- `docs/NEXT-SESSION.md` — REWRITTEN (the state card v58; the
  SESSION-058 prompt; the wipe count 51 → 52; the go.dev-listing
  recipe correction in the environment block — the toolchain
  double-verify's second side is now the ARCHIVE listing
  `?mode=json&include=all`, because go1.27.2 left the stable listing
  when go1.27.3 shipped; this session's re-verification + the open
  CI verdict recorded; the webkit ordinal unchanged from the
  corrected form — the SIXTEENTH is the pending one, last done the
  fifteenth at SESSION-047).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-057
  entry appended at the tail (the house pattern).
- `README.md` — the SESSION-057 status entry appended (the house
  pattern).
- One spec-vs-reality deviation found and fixed in this commit (the
  go.dev stable-listing assumption in the rebuild recipe — see the
  slice above); the battery itself matched the committed shape
  exactly, and the environment recipes otherwise held on the 52nd
  rebuild without correction.

## Blockers & decisions needed

- **NONE blocking.**
- Carried, unchanged: the milestone release decision (the owner's);
  T-82's live execution (the owner's); the dissolution call
  (ADR-0002 Am.5); the vanilla-vs-vite call; the go-directive
  acknowledgement (the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client pin); x/net/html keep-or-swap; the
  answering-window belt + the T-71 merged-verdict question; the
  batch-scorecard modal question; the v0.91.0 zip-parity reading
  (minor); the deferred test debt (parity report §9).

## Owner actions requested (the menu, in priority order)

1. **T-82's live execution, when convenient** (path (a) next session):
   run `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1's nine items
   (the CLI legs against the live `v0.92.0-go.1` release zip, or the
   full two-exe set against the next release) and fill §2 — the signed
   checklist closes T-82 and unlocks the Phase-7 exit evaluation.
2. **"Ship it" — the milestone release** (path (b)): the desktop app
   is complete, proven, and one `git tag` away (the workflow is
   proven end-to-end + cross-machine). Say the word + pick the number
   (the desktop milestone suggests `v0.92.0-go.2` or a `v1.0.0-go`
   shape — the owner's call) and the next session cuts it: VERSION →
   the tag `v$(cat VERSION)` (the tag law + the tag-state build law) →
   the proven workflow → the release report per AC-23's scaffold
   (T-82's checklist then runs against the real release).
3. **Otherwise — the Phase-8 pick** (path (c)): doc 12's future-work
   register — the dormant cloud quartet revival, the gotd/td Telegram
   leg revival (T-51 back; capability PoC-proven on the owner's
   account, dormant since the Option-B stamp), Linux/macOS builds,
   localization, the visual redesign (separate spec; doc 07 §D
   boundary); or small polish from the deferred-debts list (parity
   report §9: `test_noteredo`'s remaining un-ported selections; the
   note_state SURGERY path's CRLF round-trip differential; the
   standing owner calls).

## Handoff

- **Next session: SESSION-058 = the Phase-7 finish, owner-input
  driven** — the same decision tree, unchanged: (a) T-82 intake → the
  flip + the Phase-7 exit evaluation; (b) "ship it" → the release
  session; (c) the menu again.
- NEXT-SESSION.md rewritten: yes (the state card v58, the SESSION-058
  prompt inside).
- Push state: CLEAN — this close commit (docs-only) is the session's
  single push; the ephemeral GIT_CONFIG_COUNT PAT form; never stored.
  The push's go-ci run is watched per the CI-hygiene law (the chat
  report carries the verdict; go-ci was green at this session's open
  — run `39583012792` at `3ae2443` — and all battery legs were green
  locally pre-push).
