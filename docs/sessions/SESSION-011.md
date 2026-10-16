# SESSION-011 — 2026-10-16 — Phase 3 OPEN (T-30 llmclient landed) — the LLM client port with fake-server payload goldens: 556 tests race-clean (+131), the twelfth golden generator, the environment rebuilt from the sixth full wipe

## Goal (from NEXT-SESSION.md at start)

SESSION-011 = Phase 3 "Network and LLM" opens: T-30 first
(`internal/llmclient` — port `core/llm_client.py`: the three providers,
the JSON-mode memoized retry, extract_json, the detection probes, the
llama.cpp /props grab, the autodetect policy), per the roadmap's strict
order; T-31 (webfetch, XL) explicitly deferred to its own session; T-32
(webextract, S) the candidate second slice if capacity remained.

## What was done (task IDs, files, commits)

- **The environment was fully wiped again (the sixth full reset)** —
  verified with evidence at hash level before anything else: both repos
  gone, `/home/z/.local/go` gone, `/home/z/venvs` gone, no local trace
  of the remote URLs. Rebuilt per docs/NEXT-SESSION.md §1: the PAT from
  chat context was passed ephemerally via the Basic
  `x-access-token:<PAT>` header (never stored anywhere — it appears in
  no file, no env dump, no log); the GitHub API identified the account
  (`assadigit`) and both repos were re-cloned;
  `second-brain-project` clean at `031fe8c` = origin/main (T-25 =
  `901644a` underneath — the Phase-2 close verified), GitCurator
  checked out at `98f4fc71cfb862cee7225d864636c08062724348` (exact
  match with `testdiff/PINNED-GITCURATOR-REV`), symlinks recreated,
  go1.27.2 reinstalled to `/home/z/.local/go`, the stdlib-only Python
  3.12 venv recreated at `/home/z/venvs/gitcurator`. **The go.mod pin
  verified at blob level before any work**
  (`git rev-parse HEAD:go.mod` =
  `6721ebb8766573ec2209e6f536755822a905cee8`, the v1.36.1/go-1.22
  bytes) and **the 425-test baseline re-verified green (race-clean)
  on the rebuilt toolchain before any new work.**

- **T-30, commit `f2073b5` (131 new tests → 556 total, race-clean)** —
  `internal/llmclient`, tests written first, the implementation made to
  pass:

  - *The generator (tests-first law's golden half)*:
    `testdiff/goldens/dump_llmclient.py` — the TWELFTH generator, run
    against the unmodified Python reference at the pinned rev. It
    drives the reference's own `llm_client` functions against
    recording stdlib http.servers on 127.0.0.1 and dumps three
    goldens into `internal/llmclient/testdata/`:
    **llmclient_payloads.golden.json** (15 scenarios — THE WIRE
    CONTRACT: the exact requests the reference puts on the wire:
    method, path, the app-chosen headers (lowercased; the
    runtime-default User-Agent dropped — only the app's own
    `GitCurator-detect` UA is pinned), and the RAW body bytes in
    Python json.dumps shape), **llmclient_behavior.golden.json**
    (46 probes: the probe matrix, detection, detect_ollama details,
    preflight, error-family messages, warnings, routing — per-run
    server ports normalized to `<ROOT>`/`<API>`/`<DEADURL>`
    placeholders), and **llmclient_pure.golden.json** (18 sections:
    the pure-logic corpus incl. the quirk batteries). Generator
    determinism proven: two consecutive runs produce byte-identical
    sha256s for all three files.

  - *The storage promotion*: `internal/storage/pyjson.go` gained
    **`PyDumpsCompact`** (Python `json.dumps(v)` DEFAULT settings —
    the WIRE shape: ensure_ascii=True with `\uXXXX` lowercase-hex
    escapes, astral surrogate pairs, ", "/": " separators, no indent)
    and **`DecodeJSONValue`** (value-level `json.loads` — any
    top-level shape, the same int64/float64/OrderedMap rules as
    DecodeJSON), each with its own unit tests
    (`pyjson_compact_test.go`). These are load-bearing: the LLM wire
    bodies are byte-compared against the payload goldens, so the
    encoder must reproduce Python's bytes exactly (the golden pins
    `{"model": "m", "messages": [{"role": "user", "content": "hi"}],
    "temperature": 0.7}` — separators and all).

  - *The port* (`internal/llmclient/`): `llmclient.go` (the timeout
    wrapper as context deadlines — the spec's Go note; `extract_json`
    with the direct → fenced → first-balanced-brace ladder incl. the
    ONE-attempt-then-break law; the loopback law; the JSON-mode memo;
    `estimate_tokens`; `resolve_task_model`; the CloudLLMError family
    with the body[:400]/message[:200] rune-truncation law; the
    Python-value helpers `pyTruthy`/`pyOr`/`pyInt`), `openai.go`
    (openai_chat incl. the 400 → one-clean-retry → memoize law,
    openai_list_models, preflight_openai), `anthropic.go`
    (anthropic_chat with the system split + REQUIRED max_tokens +
    content-block joining + the json-mode-appends-only-to-EXISTING-
    system quirk, anthropic_list_models, cloud_chat, preflight_cloud),
    `ollama.go` (the OllamaAPI seam + the /api/chat + /api/tags wire
    client, ollama_chat's kwargs law — num_predict rides ONLY inside
    the num_ctx branch, the reference's nested-branch quirk; JSON
    mode defaults ON via `JSONModeOff` inversion; ollama_reachable),
    `llamacpp.go` (URL normalization, the /props fingerprint, the
    /props LIMITS GRAB incl. the Python int() quirks (bool→1,
    string-number, float truncation; -1/0 = unlimited→nil), the
    /v1/models fingerprints (Server header / owned_by / .gguf id),
    probe_llamacpp (never raises; the detail lines byte-exact incl.
    the `·` separators), detect_llamacpp (process ports first, scan
    provenance), resolve_llamacpp_model, the autodetect decision
    table with the exact emoji messages), `process.go` +
    `process_windows.go`/`process_other.go` (the tasklist/netstat/ss
    subprocess parses — kept for parity per the spec's Go notes, with
    CREATE_NO_WINDOW on Windows), `detect.go` (detect_ollama — the
    GitCurator-detect UA, the name-or-model + dedup-keep-order rule,
    the failure details with Python exception-class prefixes,
    found=true even with ZERO models — the Python law the port
    initially got wrong and the golden caught).

  - *The test files* (ported first, then made to pass):
    `llmclient_test.go` (test_core's TestLLMClient + the pure-golden
    drivers for extract_json/estimate_tokens/resolve_task_model/
    constants), `openai_test.go` (test_phase4's
    TestOpenAIChatSuccess/Fallback/Failures + TestModelsPreflight +
    the max_tokens leg + the proxy-safe loopback legs),
    `anthropic_test.go` (test_v0230's TestAnthropicChat/
    ListModels/CloudChatRouter/PreflightCloud — the router tests
    prove routing ON THE WIRE via the `/anthropic.com` path trick,
    no mocks), `ollama_test.go` (TestOllamaChat + TestOllamaRealHTTP
    — the port's own HTTP client plays the SDK's role),
    `llamacpp_test.go` (test_llamacpp's llm_client classes:
    TestNormalizeUrl, TestPropsIdentification, TestProbe, TestDetect,
    TestResolveModel, TestProxySafeLoopback, TestProcessDiscovery,
    TestSecondaryIdentification, TestAutodetectDecision,
    TestLlamaCppLimitsFromProps, TestProbeCarriesTheLimits,
    TestLabel), `detectset_test.go` (test_detectset's
    TestDetectOllama), `golden_test.go` (the payload/behavior/pure
    golden drivers), `fakes_test.go` (the fake-server rig — real
    local net/http listeners mirroring the Python suites' stdlib
    http.server fakes; the replies speak PyDumpsCompact bytes because
    the golden pins them).

  - *The goldens caught four real port bugs before landing* — the
    honest evidence the golden discipline exists for: (1)
    DetectOllama returned found=false on an up-but-empty server
    (Python's law: found=true, detail "up but NO models installed");
    (2) the nested `default_generation_settings`/`params` reads
    bypassed the props grab (the OrderedMap/map duality needed an
    `asOM` adapter); (3) the fallback-memoization golden exposed the
    memo key spelling (`base` vs `base + '/v1'`) — a generator bug
    fixed on the Python side, regenerating the golden (the reference
    run, not a hand edit); (4) `ChatOptions.Temperature 0` collapsed
    to the 0.7 default — fine for every real caller, recorded as a
    divergence in doc.go.

  - *Spec-vs-reality check*: spec-core-llm-client.md held up against
    the source throughout — no blueprint module-spec needed amending.
    The one spec INaccuracy found was in its Tests section wording
    ("fake-server payload goldens (3 providers)") which the
    implementation satisfies exactly. `internal/llmclient/doc.go`
    records every documented divergence (the SDK-object duality
    collapse, JSONModeOff, the temperature-zero corner, the int()
    quirks not replicated, the Go ProxyFromEnvironment loopback
    corner, the process-parse parity).

- **T-31 (webfetch, XL)**: NOT started — the roadmap and the session
  prompt both say it deserves its own session, and this session's
  capacity went to the environment rebuild + T-30.

- **T-32 (webextract, S)**: NOT started, with a **blocker surfaced
  for the owner instead of improvised past** (see Blockers): the
  module spec's Go notes name `golang.org/x/net/html` — a NEW
  dependency, forbidden by the standing law without an ADR, and
  Phase 3's task-list row carries no ADR reference (unlike T-31's
  tls-client, which ADR-0006 pre-authorizes). The two lawful options
  (an ADR adding the dependency, or a hand-rolled port of the
  reference's stdlib `html.parser` behavior — the reference itself is
  stdlib-only) are an owner call; it goes to NEXT-SESSION.md.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-30's task-list acceptance — PASS**:
  1. *Fake-server payload goldens* —
     `internal/llmclient/testdata/llmclient_payloads.golden.json`
     (15 scenarios, all three providers, byte-compared by
     `TestGoldenPayloads`), generated by the twelfth generator
     against the unmodified reference; determinism proven
     (regeneration → identical sha256s:
     `0b8003fe…`/`3eddb574…`/`f19acb8f…` after the final content).
  2. *test_detectset-equivalent* — `detectset_test.go`'s
     TestDetectOllama battery (9 tests) + the detect_ollama legs of
     the behavior golden; every llm_client-scoped case of the Python
     class is ported (the GUI/CLI halves belong to later phases and
     stay out per the package-scope law).
  3. *test_llamacpp-equivalent* — `llamacpp_test.go` ports every
     llm_client-scoped class of test_llamacpp.py (12 classes, 60+
     assertions incl. the process-parse batteries, the fingerprint
     matrix, the autodetect decision table, the limits grab); the
     GUI-only classes (TestApplyAutodetect, TestStartupWiring,
     TestTheGuiGrab…) are Phase-6 scope, recorded in the file header.
- **Phase 3 exit gate — ON TRACK (first stone laid)**: "fake-server
  contract tests green (Ollama/OpenAI/Anthropic payload goldens)" —
  the payload goldens exist and are green; the HMAC loopback rig
  (T-34) and redirect/UA tables (T-31/T-33) remain for later
  sessions.
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -count=1
  -race ./internal/...` → **15 packages ok, 556/556 tests PASS,
  race detector clean, 0 failures** (was 425 at session start);
  differential case `status-no-config` **PASS**
  (`[testdiff] status-no-config: PASS`); the goldens regenerate
  byte-identically.
- **Honest notes**: (1) T-31/T-32 not started — capacity went to the
  sixth environment rebuild + T-30's size (the prompt's own "one task
  done properly with proof beats two half-done" rule applied).
  (2) The fake-server rig in the Go tests mirrors the Python fakes'
  behavior, but the SERVER-SIDE reply bytes are the Go rig's own
  PyDumpsCompact output — what the goldens pin is the CLIENT's wire
  behavior (requests), which is the contract that matters; the
  reference's own replies were captured behaviorally (returns,
  errors, warnings) rather than byte-pinned (runtime JSON separators
  don't exist on the reference's READ side). (3) The mid-session
  repair loop on the test files (a scripted literal-rewrite mishap
  cost several iterations of brace surgery) touched only test files,
  never goldens or implementation; the final `gofmt -l` + vet +
  full battery prove the end state clean.

## Spec updates made (file → what changed, why)

- `testdiff/README.md` — the generator inventory grew by one
  (`dump_llmclient.py` — the twelfth), with its three golden shapes
  described (pure corpus + THE PAYLOAD GOLDENS + the behavior golden
  with the `<ROOT>`/`<API>`/`<DEADURL>` normalization law).
- `internal/llmclient/doc.go` — the package law doc: the wire-bytes
  law, the timeout law, the loopback law, the memo law, and every
  documented divergence with its reason.
- `internal/storage/doc.go` — NOT amended (the additions live in
  pyjson.go's own doc comments; the package doc's scope statement
  already covers "the JSON encoders").
- No blueprint module-spec needed amending (spec-core-llm-client.md
  matched the pinned reference throughout).

## Blockers & decisions needed

- **T-32 dependency question (owner decision requested, NOT
  blocking T-31)**: `spec-core-web-extract.md`'s Go notes name
  `golang.org/x/net/html`, but the standing law forbids new
  dependencies without an ADR and no ADR authorizes it. Options:
  (a) the owner approves a new ADR adding golang.org/x/net/html
  (pinned) for T-32 — the pragmatic choice, x/net/html is the
  de-facto standard HTML parser; or (b) T-32 hand-rolls the
  reference's `html.parser` behavior on the stdlib (the reference
  itself is stdlib-only, so a faithful port is feasible — larger
  effort, zero dependencies). T-31 (webfetch) is unaffected either
  way (its tls-client pin is already authorized by ADR-0006 — check
  doc 06 §2 for the exact pin before `go get`).
- The standing sandbox-reliability observation carries (sixth full
  wipe; push-after-every-task remains the load-bearing mitigation —
  T-30 was pushed the moment its battery went green, commit
  `f2073b5`).

## Owner actions requested

1. The T-32 dependency decision above (an ADR for x/net/html, or the
   hand-rolled stdlib port) — a one-line verdict in the next session
   prompt or the overseer chat is enough.
2. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-012 = Phase 3 continues — T-31 webfetch
  (XL), its own session per the roadmap**; T-32 (with the owner's
  dependency verdict) follows.
- `NEXT-SESSION.md` rewritten: yes (v12 — the Phase-3 continuation
  handoff).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
