# SESSION-042 — 2026-10-20 — Phase 6 (Path B): T-79 the GUI doors addendum + T-79a COMPLETE (the five per-test doors) — the owner's Path-B stamp ("the slow, more reliable and more rigorous method"); 2280 main-battery tests over 36 packages + 59 webui over 4; the thirty-seventh full wipe repaired BEFORE work + the ninth webkit rebuild

## Goal (from NEXT-SESSION.md at start)

SESSION-042 = the owner's phase-exit call FIRST (the fork: A = declare
Phase 6 CLOSED → T-80; B = the tasks addendum → T-79), then the chosen
path's work. The call arrived in the owner's first reply: **"Do the
Slowr, more reliable and more rigorous method, I dont mind how long it
takes."** — read as **Path B** (the slower, more thorough route: name
every honestly-disabled door as a task, then build them; Phase 6 stays
OPEN until the tail is empty). SESSION-042 = T-79: the addendum rows on
doc 13 + the first implementation slice (T-79a, the five per-test
dialogs).

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-seventh full wipe repaired BEFORE work.** At boot BOTH
  repos, the toolchain, the venv, the module cache, and the wails CLI
  were gone (the survey: only the sandbox scaffold + `.gitconfig`
  survived). Recovery, all from the committed recipes, all in
  foreground chunks: both repos re-cloned with the ephemeral
  `GIT_CONFIG_COUNT` PAT form (nothing persisted; both `.git/config`s
  verified clean); the work repo tip verified at `b8cffd7` =
  SESSION-041's close (T-78 COMPLETE), in sync with origin/main, the
  tree clean; the reference pinned at `98f4fc71cfb8…` =
  `testdiff/PINNED-GITCURATOR-REV` byte-match; go1.27.2 sha-verified
  THREE ways (the downloaded file = TOOLCHAIN.md Am.2's committed
  64-char pin = go.dev's published JSON); the venv rebuilt at Python
  3.12.15 — **one deviation recorded**: `--without-pip --copies`
  (this boot's system python lacks ensurepip; symlink(2) is blocked at
  boot — confirmed live, `ln -s` refuses and kills the call); the
  wails CLI v2.16.0 (GOBIN install); the module caches for both
  modules; the MAIN go.mod verified byte-unchanged (the authorized
  four @ go 1.24.1). The re-verification law executed before any new
  work: the FULL main battery + the webui battery green pre-work —
  **exactly the recorded state: 2253 over 36 + 56 over 4, race-clean**.

- **The owner's Path-B stamp — commit `d8f255d`** (pushed): the
  bookkeeping slice. The More menu's honestly-disabled tail
  (SESSION-040's record: the 19 doors) became the NAMED rows
  **T-79a…T-79e** on doc 13's Phase 6 table + the amendment record at
  the doc's end (the owner's call verbatim; the export zip rides
  T-80's packaging composition per the standing assignment — no
  duplicate row; Phase 7 opens only when the tail is empty). Every
  row cites its spec refs + the pinned reference's file:line homes
  (connection_tests.py:69/182/338/470, llamacpp.py:377,
  linking_tools.py:107/116, dashboard.py:313/500/770,
  bot_queue.py:929/1026, processing_control.py:615/669/807/1000/1601).

- **The ninth webkit-tree rebuild** (GUI-touching session → the tree
  first, per the state card): 144 debs downloaded (77.7 MB — matching
  SESSION-037's count), `dpkg -x` into `~/.local/lib-wails`, the 42
  .pc files' prefixes rewritten, and the 26 dangling dev links landed
  as **hardlinks to the RESOLVED system targets** (Am.4's fallback —
  symlink(2) blocked at boot; zero missing). The `wails build` CGO
  link + the gate leg below are the proof the tree is sound.

- **T-79a COMPLETE — commit `0a2ae76`** (pushed): the five per-test
  doors LIVE.
  - **The engines** (`internal/app/pertests.go`): five `Settings`
    methods streaming the reference's own line grammar onto the desk —
    🔑 Test Telegram & GitHub (the Option-B re-scope: the GitHub half
    verbatim over `GitHubGet` — the token from the LIVE config under
    the save-on-change law, the anonymous octocat fallback, the
    payload-message failure line; + the Worker queue leg over
    `WorkerCheck` = `CheckWorkerVersion` with the `GUILevels` mapping
    and the port's own Worker-not-configured info line; the opening
    re-worded "🔑 Testing GitHub & the sync queue…" — the battery's
    own vocabulary); 🌐 Test Proxy (the re-scope: the enabled-check
    warning verbatim + the sentinel's `ProxyReachable` TCP verdict
    with its own sentences); 🧠 Test Ollama (DetectOllama's list + the
    configured-model check + the REAL "What is 2 + 3?" prompt as a
    PLAIN chat (JSON mode off) + the empty-reply options retry on the
    port's vocabulary); 🔌 Test Cloud API (the flavor line by URL +
    the required-field refusals + the empty-key warning + the
    warn-never-block /models pre-flight grammar + the "Say hello" ping
    + the THREE verdict boxes with the reference's own box texts); 🦙
    Test llama.cpp (ProbeLlama + the DetectLlama fallback + the
    `_fill_llamacpp_models` pick order preserved + the model-seat
    write HOLDING THE SETTINGS MUTEX + the ping at base+"/v1" + the
    THREE verdict boxes).
  - **The seams**: five new `SettingsDeps` fields
    (OllamaChatPing/CloudChatPing/PreflightCloud/WorkerCheck/
    ProxyCheck), production-wired in `App.New`, faked at construction
    in every test (THE HERMETIC-SEAM law).
  - **The shell**: the five `HeroUI` More doors (the `MoreDoorResult`
    shape reused — title "" = the log-only doors) + `heroflow.js` (the
    five MORE_ACTIONS items flipped LIVE with their door tips + the
    `runMoreAction` dispatch + the title-only themedBox rendering) +
    `shell.go`'s `SB_POC_OPEN_MORE` probe (opens the menu for the
    screenshot; a door value CLICKS one door end-to-end — the T-70
    law applied to the menu).
  - **The guard difference, recorded**: the reference's
    telegram-lock doors guarded the TELEGRAM worker (dropped under
    Option B); these doors are read-only probes — no busy guard,
    exactly like the reference's own Ollama/cloud/llama.cpp doors.
  - **The tests**: 27 engine cases (`internal/app/pertests_test.go`:
    every line's grammar + level + every early-return branch + the box
    texts + the model-pick order + the options-retry contract + the
    seat sync) + 3 shell tests (`webui/heroui_test.go`: the wrappers
    over faked seams + the frontend laws read where they render + the
    REMAINING tail's honest-disabled notes still present) + the
    fake-seam rig extended in `settings_test.go`.
  - **THE FIELD-TAG LAW**: no new crossed payload (`MoreDoorResult`
    reused; `res.title`/`res.body` join the read-list test).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-79a's acceptance = "every door live: the opening line, the
per-step verdicts, the closing verdict; the two verdict boxes (Cloud +
llama.cpp); hermetic seams faked at construction; THE FIELD-TAG LAW
extended with every new payload" — **all GREEN**:

- Every door's line grammar pinned by the hermetic tests (27 cases);
  the two box doors' three box shapes each, byte-checked against the
  reference's own texts.
- The battery at the close: **2280 main-battery tests race-clean over
  36 packages (2253 + 27) + 59 webui tests over 4 (56 + 3)**; the
  internal/app package 3× race-clean post-change (the ritual; the one
  new mutation point — the llama.cpp model seat — holds the Settings
  mutex, the SaveSettings discipline); gofmt/vet clean over both
  modules; the MAIN go.mod byte-unchanged all session (verified at
  every step).
- The `wails build` links clean over the 9th-rebuilt webkit tree
  (33.9 s); **no mode flips** (`git diff --raw`: the wailsjs files
  stay at their 755 baseline; the content-only diffs = the five
  regenerated bindings); no stray binaries (the shell lands in
  gitignored `build/bin/`).
- The gate leg GREEN: first frame 779 ms / 284 ms (≤ the 1.5 s gate);
  the 10k flood through the production transport with 0 drops (41
  batches, 28 ms max handler).
- The screenshots VLM-verified (3 passes, rate-limit pauses between):
  the More menu open with the reference's own verbatim action list;
  **the five per-test doors rendered LIVE (the same dark text as the
  live battery door) beside the remaining tail's honest-disabled
  wash** — the first pass's misread of the two emoji-led labels
  (Ollama/llama.cpp "grayed") was settled per the display-pipeline
  lesson: the byte-level source check (the five live keys, the uniform
  rendering path) + a focused cropped pass (all five SAME). The
  end-to-end click leg: the probe clicks 🧠 Test Ollama LIVE and the
  desk speaks the door's own lines ("🧠 Testing Ollama…" + the honest
  URLError refusal with the reference's own remedy line — the
  character-level VLM noise on the emoji noted; the URL/errno details
  match the engine's format exactly).
- The evidence archived at `docs/poc-artifacts/t-79a-per-test-doors.md`
  (+ the two screenshots + the gate-results JSON).
- The bookkeeping gate: the rows T-79a…T-79e + the amendment on doc 13
  (commit `d8f255d`); the T-79/T-79a rows COMPLETE (`0a2ae76`).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/13-implementation-task-list.md` — the T-79
  umbrella row + the T-79a…T-79e addendum rows + the amendment record
  (the owner's Path-B stamp); the T-79 + T-79a rows COMPLETE with the
  full records.
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-042
  entry.
- `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the dated
  note on the connection_tests section: T-79a's landing + the recorded
  divergences (the options-retry vocabulary, the raw-response-type
  line skipped, the model-seat mutex, the Worker-not-configured line,
  the re-scoped doors' opening lines).
- `README.md` — the SESSION-042 status block.
- `webui/README.md` — heroui.go's + heroflow.js's map entries (the
  per-test doors + the probe).
- `docs/poc-artifacts/t-79a-per-test-doors.md` + the two screenshots +
  the results JSON — the archived evidence.

## Blockers & decisions needed

- **No new blockers.** The phase-exit fork was RESOLVED by the owner's
  Path-B stamp; Phase 6 stays OPEN with the named tail.
- Carried, unchanged: the dissolution call (ADR-0002 Am.5); the
  vanilla-vs-vite call; the go-directive acknowledgement (thirty-third
  session carrying it); x/net/html keep-or-swap; the answering-window
  belt + the T-71 merged-verdict question (SESSION-039's record).
- **The environment deviation, recorded**: the venv's
  `--without-pip --copies` shape (this boot's ensurepip absence +
  the symlink block) — harmless for the stdlib-only harness; noted in
  the session report + the row; TOOLCHAIN.md Am.4's fallback record
  covers the symlink half.
- The deferred test debt on the books is unchanged (test_noteredo's
  un-ported selections; the note_state surgery-path CRLF round-trip
  differential).

## Owner actions requested

- None urgent. When convenient (no rush): the standing calls —
  dissolution, vanilla-vs-vite.

## Handoff

- **SESSION-043 = T-79b: the linking tools** (🪝 Recall hooks dry-run +
  🔗 Build link suggestions + ✅ Validate Vault + 🔍 Verify Vault +
  📁 Recategorize Notes — the reference's SAFE defaults over the
  ported engines; spec-tools' build_links/add_recall_hooks +
  linking_tools.py:107/116 + dashboard.py:313/500 +
  connection_tests.py:219). The engines live (T-50's
  internal/mirror linking; T-45/T-46's vaultscan/vaulthygiene).
- The remaining tail after that: T-79c (the bot-queue tool dialogs),
  T-79d (the Chrome doors), T-79e (the decommission picker) — then
  Phase 6 closes and Phase 7 (T-80…) opens.
- NEXT-SESSION.md rewritten: yes.
- Push state: all work committed + pushed (the ephemeral PAT; never
  stored, never written) — `d8f255d` (the bookkeeping) + `0a2ae76`
  (T-79a).
