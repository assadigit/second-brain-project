# SESSION-012 — 2026-10-16 — Phase 3 continues (T-31 webfetch landed) — the web-fetch ladder port with three goldens: 676 tests race-clean (+120), the thirteenth golden generator, the environment rebuilt from the seventh full wipe, and the go-directive reality recorded

## Goal (from NEXT-SESSION.md at start)

SESSION-012 = Phase 3 continues: **T-31 `internal/webfetch`** (XL — the
whole session's task) — port `core/web_fetch.py` per
spec-core-web-fetch.md: the route ladder (proxy/direct/loopback), the
redirect ladder, the header set, the rate limiter, the size/time caps,
decompression, the failure categories, DoH, the wayback machine, and
the impersonation door. T-32 (webextract, S) only if T-31 finished
early AND the owner's dependency verdict was in hand (it was not — the
verdict stays surfaced, not improvised).

## What was done (task IDs, files, commits — cite hashes)

- **The environment was fully wiped AGAIN (the seventh full reset)** —
  verified with evidence before anything else: both repos gone, the
  toolchain gone, the venv gone. Rebuilt per docs/NEXT-SESSION.md §1:
  the PAT from chat context passed ephemerally via the Basic
  `x-access-token:<PAT>` header (never stored — no file, no env dump,
  no log); `second-brain-project` re-cloned clean at `a3466e7` =
  origin/main (T-30 = `f2073b5` underneath), GitCurator checked out at
  `98f4fc71cfb862cee7225d864636c08062724348` (exact match with
  `testdiff/PINNED-GITCURATOR-REV`), symlinks recreated, go1.27.2
  reinstalled to `/home/z/.local/go`, the stdlib-only Python 3.12 venv
  recreated. **The go.mod pin verified at blob level before work**
  (`6721ebb8766573ec2209e6f536755822a905cee8`, the v1.36.1/1.22
  bytes) and **the 556-test baseline re-verified green (race-clean)
  plus the differential `status-no-config` PASS** on the rebuilt
  toolchain before any new work.

- **T-31, commit `4d93877` (120 new tests → 676 total, race-clean;
  pushed as `a3466e7..4d93877`)** — `internal/webfetch`, the
  characterization tests ported FIRST, the implementation made to pass:

  - *The generator (the tests-first law's golden half)*:
    `testdiff/goldens/dump_webfetch.py` — the THIRTEENTH generator,
    run against the unmodified Python reference at the pinned rev.
    Three goldens into `internal/webfetch/testdata/`:
    **webfetch_pure.golden.json** (24 sections: the constants byte-
    exact, the whole proxy_from_config contract, the Chrome-131
    request header set WITH ORDER as pairs, the hop-header Referer law,
    the full HTTP-error verdict table, Retry-After parsing incl. the
    HTTP-date ladder, the URL variants, the loopback/bare-hostname
    tables, the capped decompression battery, the PDF sniff, the
    decode ladder, the rate penalties, the combined categories, the
    impersonation proxy-URL table, the TLS/DNS marker matrices),
    **webfetch_payloads.golden.json** (7 scenarios — THE WIRE
    CONTRACT: the plain fetch's header set, the redirect chain with
    hop cookies + the Referer law LIVE on the wire, the cross-host
    origin-only Referer, the custom UA with dropped client hints, the
    DoH probe (patched endpoint), the Wayback availability probe
    (patched API, the url= percent-encoding), and the http-type
    proxy's absolute-URI request line), and **webfetch_behavior.
    golden.json** (25 probes: the redirect truth, the phase-2 local-
    server battery, the DoH verdicts, the dead-page rescue ladder,
    the rate penalties, the origin-down Wayback courtesy, the third
    door's trigger matrix + verdicts + the autopip laws, the
    both-doors alternation table, the preflight facts — per-run ports
    normalized to `<ROOT>`/`<ROOT2>`/`<DOH>`/`<ARCHIVE>`/`<PROXY>`/
    `<DEADPORT>`). Generator determinism proven: consecutive runs
    produce byte-identical sha256s (`304907e3…` pure /
    `e5f15261…` payloads / `8a4e1599…` behavior after the final
    content).

  - *The port* (`internal/webfetch/`, 10 implementation files):
    `webfetch.go` (the constants, FetchResult, the sliced-sleep
    DomainRateLimiter with penalties, the loopback/domain helpers,
    pyTruthy/pyQuote/humanizeInt), `headers.go` (RequestHeaders in
    Chrome's order, BotDefenseHint, RedirectTarget with space
    encoding, the HopHeaders Referer insertion law, ParseRetryAfter,
    the HTTPErrorVerdict table, DecompressCapped with the zlib→raw
    deflate fallback, LooksLikePDF, the x/text charset ladder),
    `proxy.go` (the config contract, the preflight, buildRoute:
    http-type via ProxyURL, socks via the dialers, direct/loopback
    with Proxy nil — never env), `socks.go` (SOCKS5 via
    golang.org/x/net/proxy with the hostname UNRESOLVED — the
    rdns=True law — plus a hand-rolled SOCKS4/4a CONNECT mirroring
    PySocks' shape; failures surface as WebProxyError), `redirect.go`
    (the shared cookie jar, RedirectBrokenError, openFollowing — OUR
    redirect loop with hop cookies, 308, loop naming, the budget, the
    no-Location honest verdict), `fetch.go` (fetchOnce: the streamed
    64 KB capped read, the compressed-cap early partial, the
    error-before-body verdict law, the timeout/connection/proxy error
    classification; FetchURL: the whole ladder — the resource-truth
    branch with the dead-page rescue, the both-doors alternation, the
    single-door third-door path), `rescue.go` (URLVariants,
    IsBareHostname, the archive lookup seam, deadPageRescue, the rate
    penalty, the combined category, finalVerdict with the origin-down
    courtesy), `doh.go` (the marker checks incl. the Go dialects, the
    DoH probe with the identity Accept-Encoding wire pin, the verdict
    cache, dnsVerdictSuffix), `impersonate.go` (the third door:
    tls-client v1.16.0 with the PoC-verified Chrome_133 profile, one
    cached session with a shared cookie jar and per-route clients,
    ProxyURLForImpersonator with socks5h, MaybeImpersonate with the
    full trigger law, the availability/installer seams, the autopip
    once-per-process law), `doc.go` (every law and divergence).

  - *The test files* (ported first): `fakes_test.go` (the rig — the
    ladder server, the phase-2 server, the wire recorder with raw
    RequestURI capture, the REAL fake-SOCKS5 server that proves
    ATYP=DOMAIN, the fake limiter, the fake impersonation session
    with the Python fakes' default body/headers, the door-state
    save/restore), `webfetch_test.go` (the pure batteries: the config
    contract, the header law, the verdicts, Retry-After, the
    variants, the limiter laws, the proxy-URL table, the markers),
    `redirect_test.go` (the redirect truth on live listeners),
    `phase2_test.go` (the phase-2 battery), `ladder_test.go` (the
    DoH verdicts, the rescue ladder, the penalties, the origin-down
    courtesy, the both-doors table), `socks_test.go` (the SOCKS pipe
    with the rdns proof, the absolute-URI proxy wire, the preflight,
    the route-builder laws), `thirddoor_test.go` (the trigger matrix,
    the verdicts, the missing-library/autopip laws, the wiring spies,
    and the REAL tls-client legs: a local 200, gzip auto-decode, the
    cap, the walled 403), `golden_test.go` (the three drivers).

  - *The goldens caught four real bugs before landing* — the honest
    evidence the discipline exists for: (1) the DoH probe's wire —
    Go's transport transparently adds `Accept-Encoding: gzip` where
    urllib sends `identity` (pinned identity; the payload golden
    caught it); (2) userinfo URLs wrongly earned a www variant
    (Python's netloc keeps the userinfo; the port now checks u.User);
    (3) Go's `string()` conversion preserves invalid UTF-8 — the
    errors="replace" law now uses ToValidUTF8; (4) the GENERATOR's
    own `arm_archive` tuple-call bug fed the reference a tuple where
    a string verdict belongs — the golden initially captured the
    broken behavior; fixed on the Python side and REGENERATED (the
    reference run, not a hand edit). Plus the bring-up fixes the
    record owes: two seam deadlocks (the dohMu self-lock in the
    test-rig restore; the autopip installer called under impMu) and
    the armFakeSession slice-capture trap — all in test rigging, the
    goldens untouched.

- **The go-directive reality (a same-commit spec correction, evidence
  at hash level)**: importing tls-client v1.16.0 into the MAIN module
  forces the `go` directive to ≥1.24.x — the pin's own published
  go.mod (module proxy `.../v1.16.0.mod`) declares `go 1.24.1`, and
  go1.27.2's build/tidy enforce the module-graph rule (a hand-set
  `go 1.22` is refused/rewritten; the diff tidy insists on is exactly
  the directive line). The PoC had dodged this in the separate poc
  module; T-31 imports the door into the main binary (ADR-0006 §4).
  Resolution per the ADR-0006-Am.1 precedent: keep the
  owner-acknowledged v1.16.0 pin, ride the directive to 1.24.1,
  record the correction in the SAME commit — **ADR-0002 Am.2** (the
  decision record + the rollback path) and **TOOLCHAIN.md Am.1** (the
  table row + the reality note). ADR-0002's own letter already reads
  "Pin `go 1.22` (or newer)". `modernc.org/sqlite` stays exactly
  v1.36.1; x/net v0.48.0 + x/text v0.32.0 are direct requires
  (ADR-0006 §1 names x/net/proxy; x/text carries the charset ladder);
  x/sys v0.30.0→v0.39.0 is graph-forced.

- **T-32 (webextract, S)**: NOT started — the owner's dependency
  verdict (an ADR for golang.org/x/net/html, or the hand-rolled
  stdlib port) is still not in hand, and T-31 owned the session per
  the roadmap. The verdict stays surfaced (NEXT-SESSION §4).

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-31's task-list acceptance — PASS**:
  1. *Ladder tests incl. redirect/cookie/Referer parity* —
     `redirect_test.go` (6 tests on live listeners: hop cookies ride
     the redirect, no-Location is an honest redirect_broken, loops
     are named, the budget holds, 308 follows, plain fetch stays
     full) + the payload golden's redirect_chain/cross_host_redirect
     scenarios pinning the Referer law and the cookie carry LIVE on
     the wire, byte-compared against the reference.
  2. *Fake SOCKS server asserts domain passthrough* —
     `socks_test.go`'s TestSocksPipeDNSResolvedAtTheProxy: the
     recording SOCKS5 server sees ATYP=3 with the hostname
     (`example.test`, a TLD that cannot resolve locally) — the
     rdns=True law proven on the wire; plus the full tunnel fetch,
     the dead proxy's proxy_error verdict, loopback-never-proxied,
     and the preflight pair.
  3. *Impersonation trigger matrix* — `thirddoor_test.go` (the
     owner's exact TLS line opens the door, every refusal status
     earns it, resource/connection truths never do, the opt-out and
     loopback laws, the challenge three-leg line + hint, two routes
     two asks proxy-first) + the behavior golden's third_door_trigger
     and third_door_verdicts probes.
- **Phase 3 exit gate — ON TRACK (the second stone laid)**:
  "fake-server contract tests green" — T-30's payload goldens and
  T-31's ladder/redirect/impersonation tests are now both in place;
  the HMAC loopback rig (T-34) remains.
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -count=1
  -race ./internal/...` → **16 packages ok, 676/676 tests PASS,
  race detector clean, 0 failures** (was 556 at session start);
  differential `status-no-config` **PASS**; the three goldens
  regenerate byte-identically.
- **Honest notes**: (1) The SOCKS wire is NOT golden-pinned from the
  reference — the golden venv is stdlib-only by law and PySocks is
  absent, so the reference cannot run its own SOCKS path there; the
  Go port's own fake-SOCKS5 test carries the proof (recorded in the
  generator's header). Same for curl_cffi (the third door's REAL
  wire): the decision logic is golden-pinned via the seams, and the
  Go REAL-transport tests prove the live shape. (2) The go-directive
  bump is an owner-acknowledgement item, not a silent change —
  ADR-0002 Am.2 records it with the rollback path. (3) The
  stdlib door's header ORDER is not wire-reproducible (Go sorts
  outgoing header names); the ORDER is pinned as data in the pure
  golden, and the fingerprint-order walls are the third door's
  jurisdiction — documented in doc.go.

## Spec updates made (file → what changed, why)

- `build/TOOLCHAIN.md` — Am.1: the T-31 row + the dependency-forced
  go-directive reality (tls-client v1.16.0's own go.mod declares
  1.24.1; the toolchain enforces the graph rule).
- `docs/port-blueprint/decisions/0002-module-name-and-layout.md` —
  Am.2: the full decision record for the directive bump (applied
  with an owner-acknowledgement request, the ADR-0006-Am.1
  precedent), the unchanged sqlite pin, the x/net/x/text direct
  requires, and the rollback path.
- `testdiff/README.md` — the generator inventory grew by one
  (`dump_webfetch.py` — the thirteenth), with its three golden
  shapes and the normalization placeholder set described.
- `internal/webfetch/doc.go` — the package law doc: the package laws
  (the wire contract, the loopback law, the rdns law, the redirect
  ladder, the penalty law) and every documented divergence (the
  sorted-header-order reality, the env-proxy decision, the
  ships-armed door, the timeout collapse, the runtime error-text
  class, the marker dialects, the Python class prefixes).
- `docs/port-blueprint/module-specs/spec-core-web-fetch.md` — held
  up against the source throughout; no amendment needed.

## Blockers & decisions needed

- **The go-directive acknowledgement (owner decision requested, NOT
  blocking anything else)**: the main module's `go` directive now
  reads 1.24.1, forced by the owner-acknowledged tls-client v1.16.0
  pin (ADR-0006 Am.1+2). ADR-0002's letter permits "or newer" and
  the correction is recorded (Am.2), but the owner should
  acknowledge it — or invoke the rollback (an older tls-client pin
  needs a re-PoC of the fingerprint substitution). One line
  resolves it.
- **T-32 dependency question (STILL WAITING, the SESSION-011
  surface)**: `spec-core-web-extract.md`'s Go notes name
  `golang.org/x/net/html` — a new dependency with no ADR. Options:
  (a) an ADR adding x/net/html (pinned) — the pragmatic choice; or
  (b) T-32 hand-rolls the reference's stdlib `html.parser` behavior.
  NOTE for the owner: x/net (the MODULE) is now already in the graph
  (tls-client requires it; v0.48.0) — option (a) adds NO new module
  to go.mod, only a new package import. That fact may tip the
  decision.
- The standing sandbox-reliability observation carries (seventh full
  wipe; push-after-every-task remains the load-bearing mitigation —
  T-31 was pushed the moment its battery went green, commit
  `4d93877`).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. The T-32 dependency verdict (a or b above — with the new fact
   that x/net is already in the module graph).
3. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-013 = Phase 3 continues — T-32 webextract
  (S) with the owner's dependency verdict, then T-33 the pipeline
  glue per the roadmap's order**; the gate's remaining stone is the
  HMAC loopback rig (T-34).
- `NEXT-SESSION.md` rewritten: yes (v13).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
