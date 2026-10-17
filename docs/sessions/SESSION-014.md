# SESSION-014 — 2026-10-16 — Phase 3 CLOSED (T-33 connectioncheck + T-34 workerclient both landed) — the ninth full environment wipe rebuilt before work; 872 tests race-clean (+159); the golden signature vectors + the two differentials; Phase 4 (the engine room) owns SESSION-015

## Goal (from NEXT-SESSION.md at start)

SESSION-014 = Phase 3 continues: **T-33 `internal/connectioncheck` +
the sentinel** (M) per spec-core-connection.md (the RE-SCOPED battery
per ADR-0004 Am.3/R-0), then **T-34 `internal/workerclient`** (M) if
capacity remained — the Phase 3 exit gate's last stone.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was fully wiped AGAIN (the ninth full reset)** —
  verified with evidence before anything else: both repos gone, the
  toolchain gone, the venv gone, the my-project dir back to the fresh
  Next.js scaffold. Rebuilt per docs/NEXT-SESSION.md §1's recipes: the
  remote discovered through the GitHub API with the in-chat PAT passed
  ephemerally as the Basic `x-access-token:<PAT>` header (never
  stored — no file, no env dump, no git config),
  `second-brain-project` re-cloned clean at `51a84df` = origin/main
  (T-32 = `833b21b` underneath), GitCurator checked out at
  `98f4fc71cfb862cee7225d864636c08062724348` (exact match with
  `testdiff/PINNED-GITCURATOR-REV`), symlinks recreated, go1.27.2
  reinstalled to `/home/z/.local/go` (the go.dev tarball), the
  stdlib-only Python 3.12.15 venv recreated. **The go.mod pin
  verified at blob level before work** (`git rev-parse HEAD:go.mod` =
  working tree, the v1.36.1/tls-client-v1.16.0/x-net-v0.48.0/
  x-text-v0.32.0/1.24.1 bytes) and **the 713-test baseline
  re-verified green (race-clean)** on the rebuilt toolchain.

- **T-33, commit `e5846bd` (132 new tests; pushed the moment the
  battery went green)** — `internal/connectioncheck`:
  - *The battery (connection_check.py)*: the Result line objects
    ({name, level ok|warn|error|info, detail}) with RenderLine and
    ShortLine — the v0.37.0 compact one-line law ported as a
    hand-rolled scanner of Python's `r"\s*\(([^()]*)\)"` sub with the
    UNICODE \s class (the whitespace-run consumption, the all-pathy
    guard, the rune-wise mid-ellipsis budget with the 12/8 floors) —
    **pinned by a 22-case DIFFERENTIAL against the reference's own
    `_short_detail` outputs** (generated at the pinned rev through the
    venv; committed as static facts, including the floor-dominated
    over-budget corner at budget 10 and the all-pathy-keeps-original
    guard). `_http_get_json` with the GitCurator-connection-check UA
    and the loopback-direct law. The sections: check_vaults (the
    github/websites/manual ladder, the first writability probe with
    the `.gc-write-test-` prefix, the parent-writability door),
    check_llm (ollama/llamacpp/cloud on llmclient's T-30 probes via
    the monkeypatch seams, the /props limits grab riding the line, the
    hint row when the active provider is down), check_github (the
    Bearer + vnd.github+json ladder: /user then per-repo 404 → the
    v0.83 never-creates info, public → warn, the
    seal/websites/manual gates), and check_worker_version PROMOTED to
    the queue-reachability leg (GET {worker}/health, proxy BYPASSED,
    the EXPECTED_WORKER_VERSION compare, the enabled-gated failure).
    **check_telegram_local and telegram_live_result are DROPPED**
    (no MTProto client ships; the Telegram config keys stay read by
    the config layer — the compat law). RunLocalChecks runs FOUR
    sections — Vaults, LLM, GitHub, Worker — each crash-guarded; the
    callbacks' panics swallowed. **Summarize's headline RE-WORDED and
    pinned exactly: "ALL SYSTEMS READY — vaults, LLM, GitHub and the
    sync queue are up"** (the spec's banner delegated the wording to
    T-33's own tests; the sync queue rides the Worker).
  - *The sentinel (net_guard.py)*: ProbeInternet's endpoint ladder
    (worker → api.github.com → cloudflare-dns.com; the fixed endpoints
    as swappable vars — the Python suite monkeypatches urllib, the Go
    suite redirects the ladder at loopback fakes) with the
    ANY-HTTP-ANSWER-IS-ONLINE law and the GitCurator-sentinel UA, the
    2048-byte drain, the 120/300-rune truncations; AIStatus's
    on/warn/off/none verdicts for all three providers; ProxyReachable
    (tcp4 dial with the 0.5 s floor, the REAL syscall errno in the
    detail — ECONNREFUSED 111 pinned on Linux, the garbage-port 10808
    default); SentinelSnapshot with the never-crash guarded rows;
    OfflineVerdict's warn-once modal law; SentinelLine.
  - *netctx (netctx.py)*: VerifySSLEnabled + OutboundTLSConfig with
    the split law (websites always verify — webfetch's own T-31 law;
    LLM/cloud/probes honor the switch).
  - *The tests*: the reference's test_connection + test_netguard
    selections (the vaults matrix, the LLM routes incl. the fake
    llama-server, the GitHub ladder, the worker version ladder, the
    battery order + mutability + hostile configs, the summarize
    transitions, the sentinel trio, the offline-verdict transitions,
    the verify_ssl split) + **THE UA TABLE pinned from the wire**
    (GitCurator-connection-check on the battery/GitHub/Worker fetches,
    GitCurator-sentinel on the probes, the Worker proxy-bypass
    structural law) + the GET-only read-only law from the wire + the
    differential. llmclient gained ONE exported seam
    (LlamaCppLimitsLine — the battery's limits-line call).

- **T-34, commit `6ffc4e9` (27 new tests; pushed immediately)** —
  `internal/workerclient` + **THE LOOPBACK RIG**:
  - *The client (cloudflare_sync.py)*: Pair (the code door, the
    403-redeploy/401-re-pair advice, the live-config writes), the HMAC
    signing law (v30's fix: the KEY is sha256hex(shared_secret)'s
    UTF-8 bytes; message = METHOD\npath+query\nms-timestamp\n
    sha256hex(body); the nowMS clock seam), MakeRequest's never-raise
    law (status 0 + {'error': ...}; success = 200 AND data['success']),
    **THE 1010 LAW** (HTTPUserAgent
    "GitCurator/0.91.0 (+https://github.com/assadigit/GitCurator)" on
    every leg), **the PROXY BYPASS on every leg**, the honest poll's
    three truths, ShouldPoll (the wall-clock seam), PushVaultMirror
    (sync_id idempotency; uuid4 HAND-ROLLED on crypto/rand — zero new
    dependencies), PushDecommission, the banish + scan trios
    (propose/status/result, the 400-rune summary cap), PushErrors,
    PushBackupStatus (install_id first, the snapshot keys in order),
    Verify, GetDecommissions, BackfillChunk, HealthCheck, NormalizeURL
    (the ladder incl. the twitter→x replace-EVERYTHING law and the
    bare-"scheme:" // regainer), and the entry builders (the wire key
    order; BuildDecommissionEvents mapped to the settled FailedRow
    shape — the reference's tuple/.get latent bug documented per
    reality-wins). **The body bytes are json.dumps-identical**
    (storage.PyDumpsCompact over OrderedMaps in the reference's
    literal key order) — the HMAC signs the exact Python wire.
  - *The rig (ported from test_synclive.py's _FakeWorker)*: a real
    127.0.0.1 HTTP server speaking the Worker's protocol byte-for-byte
    — verifyHmac exactly (missing headers / the ±5-minute ms window /
    unknown install / the 1010 UA law / the sha256hex-keyed signature
    with constant-time compare), the pair door, /api/pending, the
    mirror door with sync_processed idempotency, the
    decommission/errors/backup/verify/decommissions/backfill doors,
    the banish + scan trios, /health, and the directory/automation
    doors **T-49 reuses** (the honest never-published 404, the
    constant-time Bearer publish, /api/commands).
  - *The tests*: **the five GOLDEN SIGNATURE VECTORS** (generated by
    running the unmodified reference's `_sign_request` with a
    monkeypatched clock — byte-compared), the honest poll's three
    truths, the pair error mapping incl. the config writes,
    every-leg-HMAC-valid + UA-worn from the wire, **the body-parity
    pin** (the GETs' "{}" + the pushes' json.dumps bytes), the mirror
    idempotency, the four 401 reasons (window / install / signature /
    missing headers — the last from a bare raw request), the
    proxy-bypass law (a bogus env proxy ignored, behaviorally +
    structurally), the trios, the round-trips, the directory doors,
    the 17-case normalize_url differential (the reference's own
    outputs), the builders, the uuid4 shape law.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-33's task-list acceptance — PASS**: result-shape tests ✓ (the
  reference's own selections ported + the re-worded headline pinned
  exactly); UA table asserted ✓ (from the wire, all three identities:
  GitCurator-connection-check ×3 doors, GitCurator-sentinel, the
  Worker UA + proxy-bypass law).
- **T-34's task-list acceptance — PASS**: the loopback rig ported
  from test_synclive ✓ (a real 127.0.0.1 server verifying signatures
  byte-for-byte — the five golden vectors prove the client's HMAC is
  the Worker's own); pairing/HMAC/all endpoints/1010 UA/proxy
  bypass/honest-poll ✓.
- **PHASE 3 EXIT GATE — GREEN (all five stones laid)**:
  "fake-server contract tests green" — T-30's payload goldens
  (`f2073b5`), T-31's ladder + redirect-chain parity tests
  (`4d93877`), T-32's fixture parity (`833b21b`), T-33's result-shape
  tests + UA table (`e5846bd`), T-34's HMAC loopback rig (`6ffc4e9`).
  **Phase 3 is CLOSED.**
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -count=1
  -race ./internal/...` → **19 packages ok, 872/872 tests PASS, race
  detector clean, 0 failures** (was 713 at session start; +132 T-33,
  +27 T-34 — the arithmetic closes exactly); go.mod/go.sum UNCHANGED
  (zero new dependencies — the authorized set untouched; the uuid4 is
  hand-rolled precisely to keep it that way).
- **Honest notes**: (1) the reference's exception-class names render
  through a taxonomy map (URLError/TimeoutError/ValueError/
  UnicodeDecodeError/FileNotFoundError/PermissionError) with Go's %T
  base as the fallback — the pinned contract is the detail SHAPE (the
  reference's own tests assert substrings only); (2) the
  short_detail differential caught one real scanner bug before landing
  (the no-closing-paren branch emitted '(' before its pending
  whitespace — the differential's " ( (x)" case exposed it); (3) the
  rig's directory doors are exercised by rig-law tests only for now —
  their real client (T-49 directorysync) reuses them in Phase 4; (4)
  should_poll's interval read degrades tolerantly where Python would
  raise TypeError on garbage (documented, unpinned corner).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-core-connection.md` — the
  Tests section carries the dated SESSION-014 note: the selections
  actually ported (the Telegram test classes dropped with the section;
  the GUI-wiring and CLI classes re-homed to their phases; the landed
  battery described; the read-only laws pinned behaviorally from the
  wire rather than by source grep). Reality-wins in the same commit as
  the port.
- `internal/connectioncheck/doc.go` + `internal/workerclient/doc.go`
  — the package laws and every documented divergence (the re-scope
  record, the exception taxonomy, the twitter replace-law, the tuple
  bug, the uuid4 hand-roll, the stdout→DebugLog hook).
- `README.md` — the status paragraph (below).
- No ADR changes: the port introduced no dependency decisions (the
  authorized set is byte-identical).

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — third session
  carrying it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path). A one-line acknowledgement (or a
  rollback instruction) closes it.
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- The standing sandbox-reliability observation carries (ninth full
  wipe; push-after-every-task remained the load-bearing mitigation —
  T-33 and T-34 were each pushed the moment their batteries went
  green).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-015 = Phase 4 OPENS — T-40
  `internal/websitepipeline`** (XL — the per-link state machine on top
  of T-23's state layer) per spec-core-website-pipeline.md.
- NEXT-SESSION.md rewritten: yes (v15).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
