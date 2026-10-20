# SESSION-050 — 2026-10-20 — Phase 7: T-81's dry-run COMPLETED (the owner's public-repo fix; the acceptance GREEN on a real runner; cross-machine reproducibility PROVEN) + T-82's preparable legs; battery 2377 + 89 re-verified green

## Goal (from NEXT-SESSION.md at start)

SESSION-050 = FIRST complete T-81's dry-run IF the owner fixed the
Actions billing (re-run `39512685975`, verify the green run + the
`[DRY-RUN]` release + the zip asset, the cross-machine sha256
comparison, optionally the refusal leg, then DELETE the test release +
tag and flip doc 13's row to COMPLETE), THEN T-82's preparable legs
(the clean-VM checklist document + the release-report scaffold + the
AV note — doc 10 §7/§9).

## What was done (task IDs, files, commits — cite hashes)

- **The owner's fix arrived mid-session as an instruction + credentials:
  "Make it public, so you can have unlimited CI/CD."** The billing fix
  was never a billing change — the owner chose to flip
  `assadigit/second-brain-project` to PUBLIC (public repos get free
  unlimited Actions minutes; no billing needed).
- **The pre-flip full-history secret scan** (the responsible gate
  before making 50 sessions of history world-readable): every object in
  the DB — 1,415 blobs via `git cat-file --batch-all-objects` — scanned
  for 17 token-shaped patterns (GitHub PATs, Cloudflare `cfut_`,
  Telegram bot tokens, the owner's actual secrets shared in chat, AWS
  keys, password/secret/token assignments). **Zero real hits**: the only
  matches were test fixtures (`secret=bbbb…`/`token=…` parser vectors in
  the links-grammar goldens, `token=0123456789abcdef…` in a
  platformintake test) + the `KindPassword = "PASSWORD"` event constant;
  the author identities are GitHub-noreply + sandbox locals. The flip
  landed via the API (`PATCH …/repos/assadigit/second-brain-project`,
  `private: false` → `visibility: public`, admin scope, HTTP-verified).
- **T-81's dry-run — every acceptance leg, all green**:
  1. **The re-run** (`POST …/actions/runs/39512685975/rerun`, attempt
     2): a REAL runner picked it up (`runner_id: 1000004107` vs the
     block's `runner_id: 0`), 23:26:18 → 23:30:24 UTC, **all 8 steps
     green** — checkout → setup-go 1.27.2 → the wails CLI v2.16.0 → the
     tag law (DRY-RUN waiver logged) → THE FULL DOUBLE-BUILD RITUAL
     (the ritual's own line on the runner: "PASS — byte-identical
     rebuilds: CLI exe + GUI exe + zip (warm x2 + cold x1 per exe)") →
     the AC-23 notes (A==B==cold re-derived + asserted) → the zip
     attach.
  2. **The `[DRY-RUN] v0.0.0-t81-dryrun` release** (draft + prerelease,
     id 418982158): created with the asset
     `GitCurator-v0.92.0-go.1-windows.zip` (27,225,658 bytes; downloads
     at sha256 `6409352a…` exactly) + the full AC-23 notes body.
  3. **THE CROSS-MACHINE PROOF (the session's strongest leg)**: the
     asset's numbers vs a full sandbox rebuild at the runner's exact
     tree state (a clean worktree at `0961ad9` + the tag, the ritual's
     8 legs re-run in sequence): CLI `f2ef8362…` (A=B=cold), GUI
     `99cc22e6…` (A=B=cold — T-80's own number), zip `6409352a…`
     (A=B; 13 files, 56,908,881 in, 27,225,658 zipped) — **byte-ident
     ical across machines**. The extracted CLI exe's buildinfo on both
     machines: `mod …second-brain-project v0.0.0-t81-dryrun` +
     `vcs.revision=0961ad90…` + `vcs.modified=false`.
  4. **THE STALE-TARGET FIND (recorded honestly)**: the evidence doc
     §2.1's comparison table (CLI `13808277…`, zip `071e1968…`) was
     measured at `b5c9a27`+tag — the tag MOVED to `0961ad9` (the rider)
     before the run, and the CLI exe embeds `vcs.revision`, so the
     runner's numbers were never going to equal that table. NOT a
     reproducibility failure — a stale comparison target; §2.1 now
     carries the correction + §5 the true runner-state table. (The GUI
     row was exempt all along: `99cc22e6…` held at T-80, §2.1, the
     runner, and the §5 rebuild.)
  5. **The refusal leg**: the wrong-shape tag `v9.9.9-t81-refusal`
     (pushed at `2115298`; release.yml byte-identical there — verified
     by `git diff 0961ad9 2115298 -- .github/workflows/release.yml`,
     empty): run `39515838232` **failed EXACTLY at the tag-law step**
     (`::error::release tag v9.9.9-t81-refusal != v$VERSION
     (v0.92.0-go.1) — the one-version-truth law…`), the ritual/notes/
     attach steps SKIPPED, **NO release created** — the guard's teeth
     proven on a real runner.
  6. **The cleanup**: the `[DRY-RUN]` release + both test tags
     (`v0.0.0-t81-dryrun`, `v9.9.9-t81-refusal`) deleted via the API
     (204s; local tags too); only the real milestone release
     `v0.92.0-go.1` remains; the VERSION file never touched.
- **T-82's preparable legs** (doc 10 §7/§9; the live Windows execution
  is the owner's — no Windows box in this sandbox, standing since
  SESSION-001): `docs/poc-artifacts/t-82-clean-vm-checklist.md` — the
  nine §7 items as the owner's fill-in execution record (the do / the
  verify / RESULT+DATE+NOTES each) + the item-9 spot-check recipe +
  AC-23's release-report scaffold (the per-release parity table + the
  nine-item results table + the AV section + the sign-off) + the AV
  note (doc 10 §9's mitigation ladder, the `-w -s`
  test-before-adopt rule — a stripped build is a DIFFERENT artifact,
  never the release one — and the verdict slots).
- **go-ci's state, settled with evidence**: the two red go-ci runs on
  record (`39512680594` at `0961ad9` 23:05:35Z, `39513359096` at
  `2115298` 23:11:35Z) both PREDATE the flip — billing refusals, no
  runner, the same signature as the release runs. The first post-flip
  go-ci run rides this session's close push (watched at the close; see
  the gate below).
- **Spec updates** (the commit at the close): the evidence doc's
  completion sections; doc 10 §8's second dated note; doc 13's T-81
  row → COMPLETE + the T-82 row's preparable record; doc 12's
  SESSION-050 entry; README's SESSION-050 block; this report +
  NEXT-SESSION.md (the state card v51).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**T-81: PASS — the acceptance "dry-run on a test tag" green
end-to-end.** Evidence: run `39512685975` attempt 2 (success, all 8
steps, real runner 1000004107); the `[DRY-RUN]` draft+prerelease
release with the byte-verified asset; the cross-machine table (runner
== sandbox rebuild at `0961ad9`+tag, all three artifacts byte-identical
— a STRONGER result than the recipe's original comparison); the
refusal leg (the exact-step failure + no release); the clean cleanup
(only `v0.92.0-go.1` remains). SESSION-049's PARTIAL is complete.

**T-82: PARTIAL BY DESIGN** — the preparable surface landed (the
checklist doc + the scaffold + the AV note); the live execution (items
1–8) is the owner's, per the standing no-Windows constraint.

**The close battery** (a docs-only session — no Go source, no workflow
bytes touched; `git diff` proves it): **36 main packages ok + 4 webui
packages ok, zero FAIL lines, gofmt clean both modules** — the exact
committed shape (2377 + 89 tests). Plus the ritual itself re-proven at
the tagged state (8/8 legs byte-identical, §the cross-machine proof).

**go-ci**: this session's close push is the first post-flip go-ci
trigger — its result is recorded in NEXT-SESSION.md §4 (watched before
the session ended; if red, it is a REAL failure now — the billing
excuse is gone with the private repo).

## Spec updates made (file → what changed, why)

- `docs/poc-artifacts/t-81-release-workflow.md` — the header + §2.1's
  stale-target correction + §3's RESOLVED banner + §5 (the green run,
  the cross-machine table, the refusal leg, the cleanup) — reality
  wins: the record now matches what actually happened.
- `docs/port-blueprint/10-build-release-packaging.md` — §8's second
  dated note (the dry-run completed; the public-repo fix; the
  cross-machine proof; the refusal leg).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-81 row
  → COMPLETE (the full record); the T-82 row → the preparable legs
  landed (the honest PARTIAL).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-050
  entry.
- `README.md` — the SESSION-050 status block.
- `docs/poc-artifacts/t-82-clean-vm-checklist.md` — NEW (T-82's
  preparable legs).
- `docs/NEXT-SESSION.md` — rewritten for SESSION-051 (the state card
  v51).
- `docs/sessions/SESSION-050.md` — this report.

## Blockers & decisions needed

- **NONE blocking.** T-81 is complete; T-82's remaining half + T-83
  are the next slices.
- The repo is now PUBLIC (the owner's instruction) — recorded
  everywhere it matters. Consequences carried forward: pushes still
  need the PAT (ephemeral, as always); Actions minutes are unlimited;
  the full history is world-readable (the pre-flip scan proved it
  secret-free; the standing no-secrets law keeps it that way).
- Carried, unchanged: the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement; x/net/html
  keep-or-swap; the answering-window belt + the T-71 merged-verdict
  question; the deferred test debt; the batch-scorecard modal
  question; the v0.91.0 zip-parity reading.

## Owner actions requested

1. **T-82's live execution, when convenient**: the owner's Windows
   machine runs `docs/poc-artifacts/t-82-clean-vm-checklist.md` §1's
   nine items against the NEXT real release (or the current
   `v0.92.0-go.1` zip) and fills the scaffold (§2) — the signed
   checklist closes T-82's acceptance.
2. When convenient: the standing calls (dissolution, vanilla-vs-vite,
   the scorecard-modal question).

## Handoff

- **Next session: SESSION-051 = T-83 (the docs: the Go README, the
  QUICKSTART update, the parity report — doc 15's AC-25/AC-26 shape)
  + the go-ci watch (the record in NEXT-SESSION.md §4) + anything
  T-82's owner results bring back.**
- NEXT-SESSION.md rewritten: yes (the state card v51, the SESSION-051
  prompt inside).
- Push state: all work committed + pushed at the close (the ephemeral
  PAT; never stored, never written).
