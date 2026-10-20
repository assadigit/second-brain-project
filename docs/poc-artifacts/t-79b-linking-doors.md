# T-79b — the linking tools as More-menu doors — evidence

Recorded 2026-10-20, SESSION-043. Environment: the thirty-eighth full
wipe repaired BEFORE work (both repos re-cloned with the ephemeral
`GIT_CONFIG_COUNT` PAT form, go1.27.2 sha-verified against
TOOLCHAIN.md Am.2, the venv rebuilt — this boot HAS ensurepip, the
plain form — the wails CLI v2.16.0, the module caches, and the TENTH
webkit-tree rebuild: 144 debs (77.7 MB), the 42 .pc rewrites, the 26
dev links as Am.4 hardlinks, zero missing). The pre-work
re-verification: exactly the recorded state — 2280 over 36 + 59 over
4, race-clean.

## What landed (the task row's own terms)

The reference's SAFE defaults driven from the GUI over the ported
engines — the five doors:

- **🪝 Recall hooks (dry-run)** — `add_recall_hooks.main(argv=[])`
  through the TestWorker law (linking_tools.py:107): the busy guard,
  the "⏳ Recall hooks (dry-run) running… (the GUI stays usable)"
  opening, the captured stdout+stderr (nothing streams while the tool
  runs), the LAST 40 LINES on the desk at completion, the finish
  verdict ("✅ Recall hooks (dry-run) finished." / "❌ … finished with
  exit code {rc} (see the lines above — the command-line tools print
  the exact next step)."). DRY-RUN BY DEFAULT: apply=false, no
  sample, no force — the diff lands on the desk +
  APP_DIR/reports/recall/recall-dryrun.md (the only write; the note
  files stay byte-unchanged — pinned by test).
- **🔗 Build link suggestions** — `build_links.main(argv=[])`'s
  `_suggest` half (linking_tools.py:116): the manual-vault +
  Library/ guards, the recall-fields collection over both machine
  vaults, the stale-only embeddings refresh through the configured
  provider, the cosine neighbors (top_k 8, min_score 0.30), the
  never-resuggest filter, the per-pair LLM confirm (max 50/run), the
  rejected-pairs recording, and the ONE WRITE — the Suggestions note
  under the manual vault's Library/ (`mirror.SuggestionsPath`), with
  the pending/approved/rejected counts line (cap 7 links/note) and
  the tick/strike/--collect closing line.
- **✅ Validate Vault** — `connection_tests.py:219 test_vault`
  verbatim: the five-line grammar (validating / no vault selected /
  path does not exist / valid (contains .obsidian folder) / the
  no-.obsidian warning), log-only.
- **🔍 Verify Vault** — `dashboard.py:302-498 verify_vault`: the
  manifest read (the no-manifest box + the ℹ️ line), the tracker's
  full-accounting Verify over the vault ground truth (the websites
  vault's `_inbox` in the extra dirs), the "🔍 Verify: {n} processed,
  {m} failed" line, the report (the reference's dashboard-text
  rendering — the box carries it in the port), and the enqueue legs:
  github-only failures, the settled-repo filter (the owner's law),
  the 🤝/ℹ️/📥 lines, the retry queue writes.
- **📁 Recategorize Notes** — `dashboard.py:500-660` as the TWO-DOOR
  flow: the open door (the vault check → the "Error" box; the walk
  with the locked-notes skip + the 🔒 line; the "No notes found in
  vault." box; the table payload — notes + the category list) and the
  apply door (per changed row: the EVERY-occurrence frontmatter
  rewrite, the category-folder move, the old-file removal, the cache
  row move with the ⚠️ warning line, the "📁 Recategorized {n} notes"
  finish). The frontend renders the reference's table dialog (Note |
  Current Category | New Category, per-row selects pre-selected at
  the current category, Apply/Cancel right-aligned — the buttons
  pinned OUTSIDE the scrolling body).

## The engines + seams

- `internal/app/linkdoors.go` — the five doors as `Settings` methods
  (the pertests.go pattern): the runner (`runLinkingTool` — the
  guard, the opening, the crash-safe captured run, the tail-40
  streaming, the finish verdict), the LLM router
  (`buildLinkLLM` — `_build_llm_call`'s twin: ollama/llamacpp/cloud,
  the probe→detect model resolve, the "  LLM: …" line with the
  num_ctx suffix, the provider aborts), the two tool compositions,
  and the recategorize pair.
- Three new hermetic `SettingsDeps` seams: `LinkOllamaChat` +
  `LinkCloudChat` (message-list-shaped — the tools' `llm(messages)`
  closures, JSON mode on, num_ctx only when the config speaks one)
  and `LinkEmbedCall` (the embeddings router — the seam returns the
  BUILT call, so the production wiring IS `mirror.BuildEmbeddingCall`
  and the tests inject deterministic vectors). Production-wired in
  `App.New`; faked at construction in every test (THE HERMETIC-SEAM
  law).
- `internal/store/repos.go` — `RecategorizeNote` (the cache leg:
  find by the old note_path, move to the new path + category).
- `internal/mirror/linking.go` — `KnownPairsSet` (the composition's
  never-resuggest filter; the unexported pair key crosses as sorted
  [2]string tuples).
- The shell: `webui/heroui.go`'s six More doors + `heroflow.js` (the
  five MORE_ACTIONS items flipped LIVE with their door tips + the
  dispatch + the recategorize modal + the long-body box law
  (`themedBox`'s scroll option) + the generalized probe click leg)
  + `style.css` (the `linkdoor-box` family + the recategorize
  table).

## THE GUARD LAW (the recorded widening)

The reference guards only the two `_run_phase6_tool` runners
(linking_tools.py:62-66: "⏳ A batch is running — wait for it to
finish before running the linking tools."). The task row names all
five doors "the linking tools" — the port carries the SAME guard line
on all five (validate/verify/recategorize read/write the vaults a
running batch may be mid-writing; the owner's Path-B stamp prefers
the safer family). Recorded divergence: the reference's three vault
doors run unguarded.

## The divergences (all recorded, none improvised)

1. **The guard widening** (above).
2. **The provider abort's exit code**: the reference's tools raise
   `SystemExit("message")` — exit 1 from the CLI; the GUI worker's
   `int(exc.code or 0)` CRASHES on the string code (ValueError noise,
   the message lost — a reference display bug). The port speaks the
   message into the captured stderr and exits 2 (the tool docstring's
   own "2 provider error" vocabulary). The message survives; the
   crash noise does not.
3. **The 💥 crash line's type name**: the reference renders
   `f"{type(exc).__name__}: {exc}"` (Python class names); the port
   renders the recovered panic's own text (the T-40 house law — the
   pinned contract is the shape, never the class names).
4. **Verify Vault's report surface**: the reference renders into the
   main window's dashboard text area + a second custom box ("Failed
   Links Found"); the port's More door renders ONE themed box with
   the report (the long-body scroll law) — the port's Dashboard is a
   Settings page, and the enqueue legs speak their own 📥/🤝/ℹ️ lines
   on the desk (the second box's information rides the report body +
   the log lines).
5. **The recategorize cache connection**: once per apply pass (the
   reference opens per note — same end state; per-note ⚠️ warnings
   preserved).
6. **update_dashboard() after moves**: the port's Dashboard renders on
   open — no main-window page to refresh.
7. **The vectors' iteration order** (link suggestions): the reference
   rides the SQLite row order; the port pins by_key (collect) order +
   sorted stale keys — deterministic (the tie-stability law's own
   demand).
8. **The recategorize walk order**: lexical (filepath.WalkDir) vs the
   reference's os.walk filesystem order — same note set, sorted rows.
9. **Invalid-UTF-8 notes in the recategorize walk**: the reference's
   per-file `except Exception: pass` skips them (the open raises);
   the port scans them with replacement semantics (the read itself
   never fails). Noise-level edge, recorded for the except-ladder
   audit.

## The proof

- **The hermetic tests** (the grammar pins): 32 new engine cases in
  `internal/app/linkdoors_test.go` — every door's line grammar +
  level + every early-return branch (the five validate verdicts, the
  no-vault/no-manifest/report/failed-enqueue/websites/settled verify
  paths, the tool runners' full grammar incl. the provider abort +
  the crash line + the last-40 law end-to-end over a 45-note vault,
  the suggest flow's every line + THE ONE-WRITE LAW + the
  rejected-recorded + fewer-than-2 + embed-error paths, the
  recategorize walk/apply/locked/no-notes/missing-file paths, the
  five busy guards) + THE FIELD-TAG LAW pinned on the four new
  payloads (RecategorizeRow/Model/Change/Result — the tags mirror the
  frontend's reads: path/name/category, boxTitle/boxBody/notes/
  categories, path/oldCategory/category, moved). The rig extension:
  the three seams faked in `settings_test.go`; the wrapper tests:
  4 new cases in `webui/heroui_test.go` (the doors over faked seams +
  the two-door recategorize flow + the five-guard sweep) + the
  frontend-law updates (the five live keys + the binding names + the
  landed notes GONE + the remaining tail T-79c..e's notes still
  present + the modal laws).
- **The battery at the close: 2312 main-battery tests race-clean over
  36 packages (2280 + 32) + 63 webui tests over 4 (59 + 4)**; the
  internal/app package 3× race-clean post-change; gofmt/vet clean
  over both modules; the MAIN go.mod byte-unchanged all session.
- **The `wails build`** links clean over the tenth-rebuilt webkit
  tree (24.6 s); **no mode flips** (`git diff --raw`: the wailsjs
  files stay at their 755 baseline; the content-only diffs = the
  regenerated HeroUI bindings + models.ts carrying the six new doors
  + the four payload classes); no stray binaries (the shell lands in
  gitignored `build/bin/`).
- **The gate legs GREEN**: first frame 726 ms / 269 ms (≤ the 1.5 s
  gate); the 10k flood through the production transport with 0 drops
  (41 batches, 28 ms max handler, 744 ms total).
- **The screenshots** (`t-79b-more-menu.png` +
  `t-79b-recall-click.png` + `t-79b-gate-results.json`): the VLM
  read the menu's five linking doors LIVE (normal dark text) in the
  first pass; the lower-region verdicts misread (three LIVE doors
  called grayed, one disabled door called enabled — the 042 lesson
  verbatim) and were settled per the law: the byte-level source check
  (the five live keys, the uniform rendering path — the same branch
  the per-test doors ride) + two focused cropped passes (the
  emoji-led Validate/Verify labels flip-flopped between passes — the
  character-level noise the lesson records; the hermetic pins own the
  exact bytes). **The click leg** (`SB_POC_OPEN_MORE=recall`): the
  probe clicks 🪝 Recall hooks (dry-run) LIVE and the desk speaks the
  door's own runner grammar end-to-end — "⏳ Recall hooks (dry-run)
  running… (the GUI stays usable)" + the honest "GitHub vault not
  set or not found (config vault_path or --vault)." refusal + "❌
  Recall hooks (dry-run) finished with exit code 1 (see the lines
  above — the command-line tools print the exact next step)."
  (the VLM's glyph-level ✖/… noise noted; the URLs/structure match
  the engine's format exactly).
