# SESSION-035 — 2026-10-19 — Phase 6's frontend foundation: T-72 COMPLETE (the theme-token golden + the icon golden + the production log desk + the production shell binding internal/app directly; the dissolution probe recorded, the boundary stays nested) — 1994 main-battery tests over 35 packages + 21 webui tests over 4; the thirtieth full wipe repaired BEFORE work; a REAL production race found and fixed by the re-verification law

## Goal (from NEXT-SESSION.md at start)

SESSION-035 = **T-72 (L): THE FRONTEND FOUNDATION** per doc 13's row +
spec-gui-theme-icons + spec-gui-ui (read first — the authority). Order of
work: (i) the webkit tree rebuild FIRST; (ii) ask the owner the
vanilla-vs-vite question EARLY; (iii) the slices in order, each
committed+pushed: (a) the theme-token CSS golden; (b) the production log
desk with coalescing + the layout shell + icons + fonts. Acceptance: the
token golden + the log perf test. guibridge/ retires at T-72; webui/
dissolves into the main module and binds internal/app DIRECTLY; the
module-boundary moves need ADR-0002's next amendment (probe first per
Am.4's law). If capacity runs short: close cleanly at a natural slice
boundary.

## What was done (task IDs, files, commits — cite hashes)

- **The thirtieth full wipe repaired BEFORE work** (both repos, the
  toolchain, the venv, the module cache, AND the webkit tree all gone at
  boot; everything rebuilt from the committed recipes): both repos
  re-cloned with the ephemeral `GIT_CONFIG_COUNT` form (no PAT
  persisted); go1.27.2 sha-verified against go.dev's published JSON
  (`ecbadb99091a…`, byte-equal to the pinned string); the venv
  (3.12.15); the wails CLI v2.16.0; the rootless webkit tree (144 debs
  → dpkg -x → the .pc prefix rewrites → **26 dangling dev symlinks
  relinked to the system runtime copies** — this fresh base image ships
  the gtk/atk runtimes system-wide, so apt skipped those downloads; the
  committed recipe's "in-tree or system" law handled it). The rebuild
  PROVEN at both levels: the CGO build links in 2.2s and the headless
  gate leg re-ran GREEN (first frame 742 ms; the 10k flood 478k
  lines/s, 0 drops, 24 ms max handler; the real job 800 notes/95 ms).

- **The suite re-verification (the law: re-verify before work when
  anything looks off — a full wipe is exactly that) caught a REAL
  production race — commit `cc89cb0`**: deskbus's `Push` after `Stop`
  was a ~50/50 select coin flip (a below-cap buffered send and a closed
  stop are BOTH ready; Go's select picks uniformly at random) — a
  post-Stop push could succeed and the line would sit in a channel
  nobody drains: silently lost, drop counter and all, violating the
  package's own law. Fixed deterministically (the stop-check BEFORE
  the send); the test untouched; `-count=50 -race` green after, the
  full webui module green under race. Prior sessions had ridden the
  coin flip (~6+ suite runs at P≈1.5% all-green — recorded honestly).

- **Repo hygiene — commit `2502e7a`**: 52 wipe-era mode flips restored
  to 644 (the two genuinely-executed launcher scripts keep +x); the
  orphan 5.7 MB wails CLI ELF binary at poc/wails/wails (committed by
  accident at SESSION-001, zero references repo-wide) removed. Both
  pushed `d81ced5..2502e7a`.

- **THE DISSOLUTION PROBE (Am.4's law, run FIRST — fully reverted)**:
  adding wails v2.16.0 to the MAIN module's go.mod forced the directive
  1.24.1 → 1.25.0 AND bumped the authorized pins (x/net v0.48.0 →
  v0.56.0, x/text v0.32.0 → v0.39.0 — wails' own go.mod demands them).
  The authorized-pins law forbids the bumps → BLOCKED; the probe was
  fully reverted (git restore go.mod go.sum — the tree ended
  byte-identical; `go build ./...` + `go vet ./...` re-verified). The
  boundary STAYS nested; the record + the owner's pending call:
  **ADR-0002 Am.5**. The direct binding rode Am.4's probe instead (the
  nested module may import the parent's internal/ packages) — nothing
  was lost.

- **T-72 slice (a) — commit `88e5c6b`**: THE THEME-TOKEN GOLDEN + THE
  ICON-PACK GOLDEN. Two reference-driven generators
  (testdiff/goldens/dump_theme.py + dump_icons.py — harness-side PyQt6
  import stubs because the venv is stdlib-only; the reference repo
  stays UNMODIFIED; both registered in run_corpus.py's CORPUS, 33
  generators) → two goldens locking THREE twins: the Go tables
  (webui/uitheme/theme.go — 76 LIGHT + 76 DARK tokens, python-keyed;
  icons.go — the 28-glyph lucide-static v0.544.0 pack, '{color}'
  markers verbatim), the CSS var dump (webui/frontend/theme.css — the
  FULL tables as custom properties, `:root` LIGHT +
  `[data-theme="dark"]`, kebab-mapped), and the JS pack
  (webui/frontend/icons.js — currentColor tint; the reference's
  _LOG_LEVEL_ICONS law: info=dot, success=check, warning=triangle-alert,
  error=circle-x). style.css trimmed to the type-scale + spacing laws
  (one definition per token; the invented --hover-bg/--hover-fg names
  renamed to the table's canonical --ghost-hover-*). A
  struct-marshal-order trap caught on the first test run (Go maps
  marshal keys sorted; structs emit declaration order — the golden's
  own order is sorted; recorded in the test's comment). The spec's
  "30 glyphs" corrected to the real 28 (reality wins, same commit).

- **T-72 slice (b) — commit `16897b5`**: THE PRODUCTION SHELL.
  webui/shell.go binds internal/app DIRECTLY (the typed Bus → the
  shell's ONE subscriber → the deskbus's 120 ms/≤250 coalescing → the
  wails bridge; the progress/status/batch/job-progress events cross as
  bus:* events); the startup passes speak their real lines through the
  same transport; the single-instance app.lock taken BEFORE the window
  — AppLock gained the additive `Refused` field (the alive-PID refusal
  is the ONE not-held case whose reference behavior is exit(1); every
  other failure CONTINUES WITHOUT); the sync_card CTA band (the 56px
  SYNC hero + Fast Sync + Scan + Test Connectivity, honestly disabled
  until T-77's hero machine); the fonts = the type-scale CSS laws (the
  reference ships no font files — the graceful-degradation stack IS
  the parity); the perf door (FloodLines) pushes through the FULL
  production chain, flood:done only after the drain. RETIRED:
  guibridge/ (deleted with its 5 tests, per the retirement law),
  webui/app.go + job.go (the T-70 gate legs), the stale App.js
  bindings. webui/shell_test.go: the transport's headless proof (the
  REAL internal/app.Bus). The nested go.mod gained the transitive
  requires as indirect (the MAIN go.mod byte-unchanged — verified).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-72's acceptance = the token golden + the log perf test
(+ the webkit rebuild's proof).

- **The webkit rebuild's proof**: the CGO link (2.2s) + the headless
  gate leg on the unmodified T-70 shell (742 ms first frame; 478,262
  lines/s, 0 drops, 24 ms max handler; 800 notes/95 ms) — then the
  production shell rebuilt + re-run against the same tree.
- **The token golden**: 6 uitheme tests green — the Go tables, the CSS
  var dump, the Go icon pack, and the JS pack ALL byte-diffed against
  the two reference-generated goldens (76+76 tokens; 28 glyphs; the
  RGBA/SVG laws; the level-map coverage).
- **The log perf test**: TWO legs. Headless (webui/shell_test.go): the
  REAL internal/app.Bus, 10,000 lines in 9 ms — ≥1.07M lines/s, 0
  drops, every batch ≤250, the [ts] law, order preserved. End-to-end
  (the gate leg, Xvfb, autorun): first frame **283 ms** (gate ≤1.5 s);
  the 10k flood through the FULL production chain (Bus → deskbus →
  desk) at **434,782 lines/s producer-side** (23 ms push / 173 ms
  drain), **0 drops**, 41 batches, **22 ms max handler time** (676 ms
  total — the main thread free ~93% of the window), **all 10,001 rows
  rendered** (the probe sees every batch — flood:done fires only after
  the drain). The task-gap sampler's numbers (689 ms max gap, the
  known WM-less-Xvfb timer clamp) recorded with the standing caveat:
  the handler-time instrument is the interactivity evidence.
- **The full battery at the close**: 1994 main-module tests race-clean
  over 35 packages (1999 − 5 guibridge — arithmetic closed) + 21 webui
  tests race-clean over 4 packages (12 prior + 6 uitheme + 3 shell).
  gofmt/vet clean over both modules; the secret scan clean on every
  commit; no mode flips in the landed tree (only the two launcher
  scripts at +x, by design).
- The dissolution was NOT landed — BLOCKED with evidence, not skipped:
  the probe record + the owner's pending call live in ADR-0002 Am.5.
  T-72's own acceptance did not require it (the direct binding rode
  Am.4's probe instead).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-gui-theme-icons.md` §4 — "30
  glyphs" → the real 28 (the golden counted them; reality wins, fixed
  in the same commit as the golden that proved it) + the landing
  record appended.
- `docs/port-blueprint/decisions/0002-module-name-and-layout.md` —
  **Am.5**: the dissolution probe record (hash-level evidence), the
  blocker (the authorized-pins law), the two owner options, and what
  landed on (b) anyway.
- `docs/port-blueprint/04-target-go-architecture.md` — no §2 change
  needed (the landed-shape note's "until T-72's binding" clause
  resolved by Am.5: the binding landed; the CLI-parity wiring moves to
  the next binding).
- `internal/app/doc.go` — the retirement note updated to the landed
  shape (guibridge RETIRED at T-72; the dissolution blocked per Am.5).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-72 row
  COMPLETE (the full record).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-035
  entry.
- `README.md` — the SESSION-035 status entry.
- `webui/README.md` — rewritten for the production shell (the module
  laws now Am.3+Am.4+Am.5).

## Blockers & decisions needed

- **The dissolution question (the owner's, recorded in ADR-0002 Am.5)**:
  (a) amend the authorized runtime set (accept wails' forced pins in
  the MAIN module — the dissolution can then proceed as mandated), or
  (b) keep the nested boundary (this session's working shape — zero
  cost, the direct binding already rides Am.4's probe). **T-73+ does
  not need this call**; it becomes urgent only if the owner wants the
  single-module layout for its own sake.
- **The vanilla-vs-vite call (the owner's, carried from SESSION-034)**:
  T-72 landed VANILLA (the PoC's proven path — zero new toolchain
  deps; the golden + the desk contract are toolchain-independent). An
  owner flip to vite later needs only an ADR + the build step change.
- Carried: the go-directive acknowledgement (twenty-sixth session
  carrying it); x/net/html keep-or-swap; the CLI-parity shell wiring
  (warmup family, inputs fetchers, websites phase, Phase-5 CLEAR)
  still lands on the app surface at the T-73+ binding.

## Owner actions requested

- The dissolution call (a) vs (b) per ADR-0002 Am.5 — no rush; T-73
  proceeds on (b) either way.
- The vanilla-vs-vite call — no rush; vanilla is landed and proven.
- Nothing else urgent: no new blockers, no new runtime dependencies
  (the authorized four unchanged; the MAIN go.mod byte-unchanged all
  session — verified at every commit).

## Handoff

- **SESSION-036 = T-73 (L): Settings pages A** per doc 13's row (Vault,
  Credentials, Proxy, LLM, Input, Sound — spec-gui-pages-a): the
  production shell (webui/) is the foundation it builds on; the pages
  drive internal/app's command surface; the theme tables + the icon
  pack are golden-locked and ready (the 28-glyph pack covers the
  Settings sidebar's section glyphs).
- The state card v36 (`docs/NEXT-SESSION.md`) carries the full
  read-first list + the SESSION-036 prompt.
- Push state: slices (a) and (b) + this close committed and pushed
  (the ephemeral PAT at each push; never stored, never written).
