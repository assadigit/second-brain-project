# T-79a — the five per-test doors (the More menu's per-test dialogs) — evidence

Recorded 2026-10-20, SESSION-042 (the owner's Path-B stamp: the slow,
rigorous route — the GUI doors addendum before Phase 6 closes). The
environment: the 37th full wipe repaired before work (both repos
re-cloned with the ephemeral `GIT_CONFIG_COUNT` PAT form; go1.27.2
sha-verified three ways; the venv rebuilt as `--without-pip --copies`
— this boot's system python lacks ensurepip and symlink(2) is blocked;
the wails CLI v2.16.0; the webkit tree rebuilt — the NINTH rebuild,
Am.4's hardlink fallback again: the 26 dev links as hardlinks to the
resolved system targets, zero missing; the wails build + the gate leg
below the proof).

## What landed

The More menu's five per-test doors, LIVE (was: honestly disabled with
their landing notes):

| Door | The reference's flow | The port's flow |
|---|---|---|
| 🔑 Test Telegram & GitHub | connection_tests.py:69 — the Telegram worker (background) + the GitHub token (inline) | **Option-B re-scope**: the GitHub half verbatim (the token from the LIVE config under the port's save-on-change law; the anonymous octocat fallback) + the WORKER QUEUE leg in the Telegram half's seat (CheckWorkerVersion's verdict, the battery's own vocabulary). The opening line re-worded: "🔑 Testing GitHub & the sync queue…" (the battery's own re-wording precedent). The unconfigured queue speaks its own honest info line. |
| 🌐 Test Proxy Connection | connection_tests.py:182 — a Telegram MTProto round trip THROUGH the proxy | **Re-scope**: the sentinel's own `ProxyReachable` TCP verdict (the state machine + the exact sentences the status bar paints). The enabled-check warning line verbatim. The deep half (a real round trip through the proxy to a blocked destination) is Option B's accepted loss — every real fetch already rides the proxy when enabled (T-31's routes). |
| 🧠 Test Ollama | connection_tests.py:338 — list models → the configured-model check → the real "What is 2 + 3?" prompt | Verbatim line grammar over DetectOllama + the new OllamaChatPing seam (a PLAIN chat — JSON mode off). |
| 🔌 Test Cloud API | connection_tests.py:470 — the /models pre-flight (warn-never-block) + the "Say hello" ping + the custom message box | Verbatim line grammar + the three verdict boxes (Connected / Empty Response / Connection Failed — the reference's own box texts) over PreflightCloud + CloudChatPing. |
| 🦙 Test llama.cpp | llamacpp.py:377 — the /props probe (+ the detect fallback) → the model pick → the "Say hello" ping through the SAME OpenAI-compatible path | Verbatim line grammar + the three verdict boxes (Connected / Not Detected / Chat Failed) over ProbeLlama/DetectLlama + CloudChatPing at base+"/v1". The model pick preserves the reference's `_fill_llamacpp_models` order (probe.Model → the config's current → the first listed; the names carry the props model). |

The guard difference, recorded: the reference's test_telegram_github
and test_proxy hold the TELEGRAM worker's mutex (dropped with the
worker under Option B); these doors are read-only probes and carry no
busy guard — the reference's own Ollama/cloud/llama.cpp doors never
guarded either.

## The seams (THE HERMETIC-SEAM law)

Five new `SettingsDeps` seams, production-wired in `App.New`, faked at
construction in every test: `OllamaChatPing`, `CloudChatPing`,
`PreflightCloud`, `WorkerCheck`, `ProxyCheck`. The engines live in
`internal/app/pertests.go` (the five `Settings` methods); the wrappers
in `webui/heroui.go` (the five `HeroUI` More doors returning
`MoreDoorResult` — the same themed-box shape the dashboard door rides:
title "" = the log-only doors); the frontend wiring in
`frontend/heroflow.js` (the five MORE_ACTIONS items flipped live +
the `runMoreAction` dispatch + the verdict-box rendering through the
one `themedBox`).

## The recorded divergences (all verified against the reference at 98f4fc7)

1. **The Ollama options-retry vocabulary**: the reference's empty-reply
   retry sends `{"temperature": 0.7, "num_predict": 50}`; the port's
   golden-pinned Ollama wire shape carries num_ctx/num_predict only —
   the retry sends the config's llm_num_ctx + num_predict 50 (the
   behavior class preserved: one retry WITH options; the temperature
   key is the recorded loss).
2. **The raw-response-type line skipped**: the reference's empty
   branch speaks `Raw response type: {type(response).__name__}` — a
   Python type name with no Go meaning (the port's parsed wire shape
   is ONE shape, the documented T-30 divergence); the two advice lines
   carry.
3. **The model-seat write holds s.mu**: the reference's fill mutates
   `self.config['llamacpp_model']` from the GUI thread (the GIL
   serializes); the port's write holds the Settings mutex — the same
   discipline SaveSettings owns for every live-map mutation.
4. **The Worker-not-configured line is the port's own**: the
   reference's telegram half always runs (its worker exists); an
   unconfigured Worker is the port's own "nothing to check" truth and
   speaks its own info line.
5. **The opening lines of the two re-scoped doors** speak the port's
   truth (the battery's own established vocabulary), never naming the
   dropped Telegram leg.

## The proof

- **The engines**: 27 test cases in `internal/app/pertests_test.go`
  (the line grammar per door — every opening line, every per-step
  verdict, every early-return branch, the box texts, the level
  mapping, the model-pick order, the options-retry contract, the
  config-seat sync) + the fake-seam rig extended in
  `settings_test.go`.
- **The shell**: 3 new tests in `webui/heroui_test.go`
  (TestHeroUIPerTestDoors — the wrappers over faked seams, the
  title/body passthrough, the desk lines; TestHeroUIPerTestLlamaCppDoor
  — the probe→pick→ping chain; TestHeroflowPerTestDoorsLive — the
  frontend laws read where they render: the five live keys, the five
  bindings, the title-only verdict-box law, the REMAINING tail's
  honest-disabled notes still present).
- **THE FIELD-TAG LAW**: no new crossed payload — `MoreDoorResult`'s
  tagged names are reused; `res.title`/`res.body` join
  TestHeroflowReadsTheTaggedNames's read list.
- **The battery at the close**: **2280 main-battery tests race-clean
  over 36 packages (2253 + 27) + 59 webui tests over 4 (56 + 3)**;
  the internal/app package 3× race-clean post-change (the ritual);
  gofmt/vet clean over both modules; the MAIN go.mod byte-unchanged
  all session.
- **The wails build**: links clean over the 9th-rebuilt webkit tree
  (33.9 s); no mode flips (`git diff --raw` — the wailsjs files stay
  at their 755 baseline, content-only diffs = the five regenerated
  bindings); no stray binaries (the shell lands in gitignored
  build/bin/).
- **The gate leg + the screenshots** (Xvfb, autorun, the probe
  `SB_POC_OPEN_MORE`): first frame 779 ms / 284 ms (≤ the 1.5 s
  gate); the 10k flood through the production transport with 0 drops
  (41 batches, 28 ms max handler); the menu screenshot
  (`t-79a-more-menu-screenshot.png`) VLM-verified: the More menu open
  with the reference's own verbatim action list, **the five per-test
  doors rendered LIVE (the same dark text as the live battery door)**
  while the remaining tail (the linking tools, the bot-queue tools,
  the Chrome doors, the decommission, the export zip) renders in the
  honest-disabled wash. The focused re-read (a cropped pass) settled
  the first pass's misread of the two emoji-led labels — the
  byte-level source check (the five live keys, the uniform rendering
  path) is the settling evidence, per the display-pipeline lesson.
- **The end-to-end click leg** (`t-79a-ollama-click-screenshot.png`):
  the probe clicks 🧠 Test Ollama LIVE — the click → the regenerated
  binding → the engine → the bus → the desk speaks the door's own
  lines: "🧠 Testing Ollama..." (info) then "❌ Ollama test failed:
  URLError: … dial tcp [::1]:11434 … Click '🚀 Start Ollama Server'
  if it isn't running." (error) — the reference's own honest-failure
  path on a box without an Ollama server (the VLM's transcription
  carries character-level noise on the emoji; the URL + errno
  details match the engine's format exactly, and the line grammar is
  pinned by the hermetic tests).
