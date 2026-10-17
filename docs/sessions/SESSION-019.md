# SESSION-019 — 2026-10-17 — Phase 4 continues: T-44 chrometabs landed at 71e5e34 — 1132 tests race-clean (+74) — THE FIFTH DOOR COMPLETE (the hand-rolled CDP WebSocket + the golden + the ported test_fifthdoor selections + the deferred auto-door discard battery + the full-circle pipeline leg); the fourteenth environment degradation repaired mid-session (the toolchain wiped, the in-progress draft preserved)

## Goal (from NEXT-SESSION.md at start)

SESSION-019 = Phase 4 continues: **T-44 `internal/chrometabs`** (L)
per spec-core-chrome-tabs.md — the fifth door:
core/chrome_tabs.py:484-2563 — the CDP WebSocket frame codec, the
DevTools socket laws, page_is_real (wired into the
handdelivery.PageIsRealFn seam SESSION-018 staged) with the deferred
auto-door discard battery, the load watch, the candidates, the
session helpers, the Chrome session ladders, the fetch driver
honesty. Acceptance: the ported test_fifthdoor selections (the house
generator pattern if a golden was warranted). T-45 vaultscan (L) if
capacity remained — with the standing instruction to close cleanly at
a natural boundary.

## What was done (task IDs, files, commits — cite hashes)

- **The environment had degraded AGAIN (the fourteenth event, this
  one MID-draft)**: the repos survived at the right revisions
  (work repo at `c2fea40` = the SESSION-018 close = origin/main;
  GitCurator @ `98f4fc7` matching PINNED-GITCURATOR-REV; both
  .git/configs clean of the extraheader), and — the lucky shape —
  this session's own IN-PROGRESS T-44 DRAFT survived (14 files,
  written 06:52–07:02 by the pre-context-loss part of this same
  session). But: a mode-flip event had touched 411 files
  (644→755), the Go toolchain and the venv were GONE, and the
  symlinks too. Repaired per the recipes: the draft's doc.go
  amendment preserved aside, `git fsck` (clean) +
  `git checkout-index -f -a` (modes restored from the index), the
  draft restored, go1.27.2 reinstalled (go.dev tarball →
  /home/z/.local/go), the stdlib-only Python 3.12.15 venv recreated,
  symlinks recreated, the go.mod pins re-verified (go 1.24.1 / sqlite
  v1.36.1 / tls-client v1.16.0 / x-net v0.48.0 / x-text v0.32.0 —
  `git rev-parse HEAD:go.mod` = 5c0d4e4), and the 20-package baseline
  re-verified green before work (the 21st package was the draft
  itself, mid-build).

- **T-44, commit `71e5e34` (pushed the moment the suite went
  green)** — `internal/chrometabs` completed (14 files: the draft
  finished, not restarted — 3 compile errors and 5 fidelity bugs
  found and fixed against the reference) + the golden + the battery
  + the pipeline leg:
  - *The frame codec* (`frame.go`): masked client frames (all three
    length encodings), server frames unmasked AND masked,
    fragmentation + control frames, the 64 MB runaway guard
    (`CDPError 'a DevTools frame claimed {n} bytes…'`).
  - *The DevTools socket* (`socket.go`): the RFC 6455 handshake
    (GET {path} HTTP/1.1, Sec-WebSocket-Key base64(urandom 16),
    64 KB cap, 101 required, no Origin), the call/event round-trip
    (id increment, `json.dumps(ensure_ascii=False)` wire bytes via
    cdpWire, stale-id and event skipping, error →
    `CDPError '{method}: {message}'`, deadline →
    `CDPTimeoutError '{method} timed out after {n}s'`),
    ping→pong, close→`CDPClosedError`, Runtime.evaluate.
  - *The Chrome process* (`session.go` + both platform files): the
    launch ladder (base → +container → +headless `--headless=new`),
    `_launch_and_announce` (the stderr DevTools line within 25 s, the
    dead/silent exe honesty, the drain goroutine), adoption through
    `DevToolsActivePort` (only when the port ANSWERS), the graceful
    `Browser.close` (5 s calls, 8 s wait) then the tree kill
    (POSIX killpg SIGTERM→SIGKILL / Windows taskkill /T /T /F),
    the throwaway-profile erase law, the HTTP endpoints (PUT
    /json/new with the GET fallback — the Chrome 111+ law), the
    anti-bot patch, `AttachedChromeSession` (never launches, never
    kills, closes only its own tabs).
  - *The real-profile rung* (`realprofile.go`): the attach link
    (Windows mklink /J with the 15 s timeout / POSIX symlink), the
    running pre-check, the identity ladder,
    `write_chrome_attach_bat` (pure ASCII + CRLF).
  - *The proxy legs* (`proxy.go`): the Chrome argv, the per-site
    route memory (`<APP_DIR>/chrome-proxy-memory.json`, atomic,
    indent 2), `plan_proxy_legs` (the proxy leg ALWAYS planned as
    the escalation's landing pad), `remember_route` only for
    delivered pages.
  - *The load watch* (`watch.go`): complete at an http(s) URL, STILL
    for `PAGE_SETTLE_S` (href change or readyState drop restarts),
    the DOM read with the one context-swap retry, a challenge
    restarts the settle and keeps the budget, the honest deadline
    sentences, the v0.54 socket-reconnect grip (up to 2 re-connects,
    the anti-bot re-applied), the wave engine.
  - *The fetch driver* (`driver.go`): ATTACH first (the owner's own
    window when its port answers), then the legs with escalation of
    direct-leg failures onto the proxy leg, per-URL
    `{url, ok, html, title, error}`, both-leg failures joined
    `' | '`, no-Chrome/loopback/non-http/dry-run honesty.
  - *page_is_real* (`verdict.go`): the full grammar — the tab-URL
    law, the ERR_* code, the crash markers, the strong/weak title
    laws (≤60 chars, ≤4 words), the HTML markers only on a thin page
    (`_visible_text_len < 500`), the exact reason strings.
    `_visible_text_len`'s backreferencing regex
    `(?is)<(script|style|noscript|template)\b[^>]*>.*?</\1\s*>` has
    no RE2 equivalent — the port scans the same grammar by hand.
  - *The candidates* (`candidates.go`) and *the delivery*
    (`deliver.go` — queue stamped door 'auto' FIRST, the
    re-verification before every write, the partial-delivery
    first-five log, the crashing-fetcher tolerance).
  - **THE WIRING**: `deliver.go`'s `init` sets
    `handdelivery.PageIsRealFn` — the seam SESSION-018 staged. Any
    binary that links the fifth door gets the real verdict at consume
    time; handdelivery itself never imports chrometabs (the cycle the
    seam breaks — proven by the external `websitepipeline_test`
    package that links both).
  - *pyurl.go*: CPython 3.12's lenient `urlsplit` hand-rolled (Go's
    `url.Parse` rejects what Python accepts): the lstrip of C0+space,
    the WHATWG `\t\r\n` removal, the ASCII-letter-first scheme rule,
    the lowercase hostname, **THE BRACKET LAW** (unbalanced brackets
    or a non-IPv6 bracketed host make urlsplit RAISE — every
    caller's except reads "no parse"; verified against the venv's
    actual CPython source, `_check_bracketed_netloc` and all).
  - *THE GOLDEN*: `testdata/chrometabs.golden.json` (9 sections —
    frame_codec, cdp_wire, page_verdict, visible_text, url_shapes,
    loopback_law, split_probe, proxy_laws, candidates) GENERATED by
    `testdiff/goldens/dump_chrometabs.py` from the unmodified
    reference @ 98f4fc7 (the plain core/ generator pattern —
    chrome_tabs is stdlib-only). NO PER-RUN FACTS: the client-frame
    codec's random mask bytes and the handshake's random
    Sec-WebSocket-Key are folded out (only deterministic header
    bits, lengths and DECODED payloads ride the golden).
  - *The behavior battery* (`chrometabs_test.go`, 63 tests): the
    ported test_fifthdoor selections — §1 the codec (7), §2 the
    DevTools client (10, incl. the mid-navigation context-swap retry
    whose anti-bot Page.enable eats the first destroyed-error frame —
    the ids line up exactly like the reference's), §2b the verdict
    (8) + the load watch (9 legs through scripted fake-net
    connections), §3 the candidates (6), §4 the loopback/free-port
    laws + the fake-exe session ladder (7: endpoint parsed + profile
    erased, dead/silent exe honesty, container/headless retries),
    §5 the driver's honest misses (3), §6 the delivery (7), **§7b
    THE DEFERRED AUTO-DOOR DISCARD BATTERY (6)** — an auto-delivered
    error/challenge page is discarded (the app's own file), the link
    keeps waiting (PendingHandLinks re-asks), the owner's saves stay
    trusted.
  - *The pipeline leg* (`websitepipeline/fifthpipeline_test.go`, the
    EXTERNAL test package — the import-cycle breaker, 2 tests): the
    FULL CIRCLE (fail → candidates → deliver → re-run as real
    fetches; the fetcher called only for the failed attempt; the
    retry row resolved) + the owner-saved manual story.
  - *Spec amended in the same commit*: the SESSION-019 Go-note
    supersession (the CDP WebSocket is hand-rolled stdlib — the old
    gorilla/websocket note is dead; zero new dependencies, the
    reference itself hand-rolls it) + the shipped record; the D7
    traceability row widened; README's status paragraph.

- **REALITY-WINS fixes the task forced** (both evidence-pinned, both
  in the same commit):
  1. **Go's `cmd.Wait()` CLOSES the stderr pipe** — a reaped
     Chrome's reads answer "file already closed", not EOF. The
     draft's `launchAndAnnounce` loop required `rerr == io.EOF`
     specifically and would have spun to the 25 s deadline on every
     dead exe (found as a real 120 s test hang, root-caused with a
     minimal repro). Fixed to the reference's own semantics: ANY
     read error + process-exited = "Chrome exited before its DevTools
     endpoint came up".
  2. **The load-watch scripts ride the UNPATCHED 0.4 s poll
     cadence** — the reference's test scripts are timed against it
     (the second probe arrives after the settle window); patching it
     smaller breaks the probe/DOM chunk sequencing. The draft's
     harness had patched it; fixed to leave the package default.

- **T-45 `internal/vaultscan` (L) — NOT STARTED, deliberately.**
  T-44's own scope (the draft completion, the golden, the 74-test
  battery, the pipeline leg, two reality-wins fixes, the environment
  repair) filled the session, and T-45 is a full L task
  (vault_scan.py:548-3353 — ~2,800 lines, bigger than the fifth
  door). The one-task law applied: one task done properly with proof
  beats two half-done; closed cleanly at the natural boundary.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-44's task-list acceptance — PASS**: *the ported test_fifthdoor
  selections* ✓ — the reference suite's own selections all landed:
  frame codec (round-trips, 16/64-bit lengths, split feed,
  fragmentation + control, oversize guard, close frame); CDP socket
  (handshake refusal + shape, call round-trip + event skipping,
  masked pong, close raises, error raises, stall timeouts, evaluate,
  fragmented reassembly, the mid-navigation context destruction); the
  page verdict (all 8 grammar laws); the load watch (all 9 legs: the
  waited real page, the error/crash never-delivered, the challenge
  that passes / never passes, never-finishes, about:blank, the
  href-change settle restart, the empty DOM); the candidates (all
  6); the session helpers + the fake-exe ladders (all 7); the
  driver's honest misses (all 3); the delivery (all 7); the consume
  verify (all 6 — the deferred §7b battery, the seam now wired); the
  §8 full circle. Plus the golden the house pattern warranted: 9
  sections generated from the unmodified reference, byte-compared
  through the canonical Python dumps.
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 -p 16 ./...` → **22 packages ok, 1132/1132 tests PASS,
  race detector clean, 0 failures** (was 1058 at session start; +74
  — 63 behavior + 9 golden + 2 pipeline legs; the arithmetic closes
  exactly); go.mod/go.sum UNCHANGED (zero new dependencies — the
  authorized set untouched; `git rev-parse HEAD:go.mod` =
  5c0d4e4a, the same bytes as SESSION-018's close).
- **Honest notes**: (1) the golden was generated ONCE and never
  hand-edited; (2) one deliberate documented micro-divergence: a
  malformed webSocketDebuggerUrl with unbalanced brackets — the
  reference's outermost except answers `unexpected: Invalid IPv6
  URL`, the port's portOfNetloc degrades to the session's own port
  (Chrome never emits such a URL; recorded in the spec's shipped
  record); (3) the §9b end-of-delivery GUI modal legs stay with the
  future GUI task (the decomtable layering law), and the reference's
  §10 release-bookkeeping tests are the Python repo's own CI facts —
  not portable; (4) the T-44 draft from the pre-context-loss part of
  this session was CONTINUED, not restarted — every one of its 14
  files was re-audited line-by-line against the reference before the
  battery was written on top.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-core-chrome-tabs.md` — the
  SESSION-019 amendment (the Go-note supersession: the hand-rolled
  stdlib WebSocket, per the dependency law and the reference's own
  design) + the dated shipped record (the files, the wiring, the
  deferred battery landed, the three reality-wins finds, the
  micro-divergence).
- `docs/port-blueprint/traceability-matrix.md` — the D7 row widened
  (the complete door, the golden + the ported selections; T-44 ✓).
- `README.md` — the status paragraph (T-44 complete, the amendment,
  1132 race-clean, next: T-45).
- `internal/chrometabs/doc.go` — the package record updated from the
  T-40 placeholder to the complete fifth-door surface (the draft's
  own amendment, preserved through the environment repair).

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — eighth session
  carrying it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path). A one-line acknowledgement (or a
  rollback instruction) closes it.
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- The standing sandbox-reliability observation carries (the
  FOURTEENTH degradation event — this one mid-draft; push-after-every
  task and the draft-preserving repair both proved load-bearing
  again).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-020 = Phase 4 continues — T-45
  `internal/vaultscan`** (L) per spec-core-vault-scan.md, then T-46
  vaulthygiene + repobanish if capacity remains.
- NEXT-SESSION.md rewritten: yes (v20).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
