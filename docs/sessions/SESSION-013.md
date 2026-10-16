# SESSION-013 — 2026-10-16 — Phase 3 continues (T-32 webextract landed) — the hand-rolled HTML tokenizer: 713 tests race-clean (+37), the fourteenth golden generator, the zero-dependency alternative taken (no owner verdict in hand), and the eighth full environment wipe rebuilt before work

## Goal (from NEXT-SESSION.md at start)

SESSION-013 = Phase 3 continues: **T-32 `internal/webextract`** (S) —
port `core/web_extract.py` per spec-core-web-extract.md with the 5 HTML
fixtures byte-compared on all Extracted fields via a fourteenth golden
generator; **check the dependency verdict FIRST** (NEXT-SESSION §4: the
owner's verdict on golang.org/x/net/html). Then, if capacity remained,
T-33 `internal/connectioncheck` + the sentinel.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was fully wiped AGAIN (the eighth full reset)** —
  verified with evidence before anything else: both repos gone, the
  toolchain gone, the venv gone, no Go on PATH, the my-project dir back
  to the fresh Next.js scaffold. Rebuilt per docs/NEXT-SESSION.md §1:
  the PAT from chat context passed ephemerally via the Basic
  `x-access-token:<PAT>` header (never stored — no file, no env dump,
  no git config; the remote URL discovered through the GitHub API with
  the same ephemeral header), `second-brain-project` re-cloned clean at
  `5cee3fb` = origin/main (T-31 = `4d93877` underneath), GitCurator
  checked out at `98f4fc71cfb862cee7225d864636c08062724348` (exact match
  with `testdiff/PINNED-GITCURATOR-REV`), symlinks recreated,
  go1.27.2 reinstalled to `/home/z/.local/go` (the go.dev/dl.google.com
  tarball, sha256 `ecbadb99…`), the stdlib-only Python 3.12.15 venv
  recreated. **The go.mod pin verified at blob level before work**
  (`git rev-parse HEAD:go.mod` = working tree, the
  v1.36.1/tls-client-v1.16.0/x-net-v0.48.0/x-text-v0.32.0/1.24.1 bytes)
  and **the 676-test baseline re-verified green (race-clean)** on the
  rebuilt toolchain before any new work.

- **THE DEPENDENCY VERDICT DID NOT ARRIVE — option (b) taken, with the
  decision recorded everywhere it matters.** The session prompt carried
  no owner verdict on golang.org/x/net/html (it restates NEXT-SESSION
  §4's two options and names the hand-rolled stdlib port as "the
  zero-dependency alternative"). Per the prompt's own instruction
  ("ask before adding anything"), NO import was added and NO ADR was
  written: T-32 hand-rolls the reference's own stdlib parser behavior
  instead. The verdict question stays surfaced for the owner (this
  report + NEXT-SESSION §4): an x/net/html ADR remains one decision
  away, and the swap is confined to `parser.go`'s scan internals — the
  handler laws, the goldens, and the tests stay the referee either way.

- **T-32, commit `833b21b` (37 new tests → 713 total, race-clean;
  pushed the moment the battery went green)** — `internal/webextract`:

  - *The generator (the tests-first law's golden half)*:
    `testdiff/goldens/dump_webextract.py` — the FOURTEENTH generator,
    run against the unmodified Python reference at the pinned rev. One
    golden: `internal/webextract/testdata/webextract_pure.golden.json`
    (constants; the 5 fixtures with every Extracted field + the fixture
    bytes as hex so the Go driver is self-contained; the 86-case
    extract() battery — title/meta precedence incl. first-wins
    og/name, the meta elif chain, duplicate attrs (last wins), charset
    detection shapes, raw-mode xmp/iframe, RCDATA title/textarea
    entities, comments/PIs/CDATA/marked sections/bogus comments, the
    stray-endtag newline law, the incomplete-tag discard laws, the
    attribute-names-may-contain-quotes law, block-tag newline collapse,
    nbsp/CR survival, the 20k cap, the 120-word js-shell boundary, the
    4000-char paywall window, the `<script\b` Unicode word-boundary
    law; the parser-level battery via the reference's own _PageParser;
    the 95-case unescape battery; the utf-8 errors='replace'
    maximal-subpart law; the 74-name charset ladder — resolution,
    strict-decode verdicts, per-byte holes, full-256 decodes; the
    22-case extract_from_bytes battery — header/sniff/replace flows,
    BOM, the windows-1252 hole fallback, the Shift_JIS
    case-preservation, the charset_used backfill law). Generator
    determinism proven: consecutive runs produce byte-identical files
    (sha256-stable).

  - *The generated data tables (embedded, sha-pinned by the golden)*:
    `entities_html5.json` (2231 entries — the HTML5
    named-character-reference table from the reference runtime's own
    `html.entities.html5`), `charsets_table.json` (13 single-byte
    charsets × 256 codepoints from the reference runtime's own codecs
    — tis-620 and cp874 as the DISTINCT codecs Python says they are),
    `cp949_table.json` (the UHC-extended Korean codec — x/text has no
    equivalent), `dbcs_accept.json` (singles/pairs/triples acceptance
    bitmaps for big5/shift_jis/gbk/gb2312/gb18030/euc-jp/euc-kr —
    which byte sequences Python's codecs actually accept, generated by
    the len==1 disambiguation). All four ride `go:embed` and are
    pinned at sha256 level over the generated file bytes.

  - *The port* (`internal/webextract/`, 6 implementation files):
    `parser.go` (the hand-rolled CPython 3.12.15 html.parser port —
    goahead's end=1 single-pass semantics proven equivalent to the
    reference's feed-whole+close; tagfind/attrfind/locatetagend
    semantics exactly, incl. the lookbehind that lets an attribute
    start after a closing quote and the '/'-not-before-'>' guards; raw
    mode for script/style/xmp/iframe/noembed/noframes/plaintext with
    the set_cdata_mode lookahead; RCDATA for title/textarea; the
    reference's _PageParser handlers — first-wins description, the
    elif chain, invisible-depth data gating, block-tag newlines),
    `entities.go` (pyUnescape — Python's html.unescape algorithm
    exactly, table and all), `charsets.go` (the decode ladder: strict
    utf-8/ascii, generated single-byte tables, hand-rolled
    utf-16/utf-32 with Python's BOM rules, x/text behind the generated
    acceptance bitmaps for the DBCS family, cp949 via its own table,
    gb18030's 4-byte grammar + the FFFD scan for the piecewise gaps,
    the meta-charset sniff, pyUTF8Replace with the maximal-subpart
    law), `pystr.go` (the Python string-law helpers —
    strip/split/lower with U+0130's two-rune lowercase and the
    \x1c-\x1f whitespace row, regex \s, rune slicing, the
    `<script\b` boundary), `extract.go` (Extracted, Extract,
    ExtractFromBytes, the constants), `doc.go` (the package laws and
    every documented divergence).

  - *The test files*: `golden_test.go` (the drivers for every golden
    section — fixtures, extract/parser/unescape/decode-law batteries,
    the 74-name ladder with probes and hole facts, the bytes battery,
    the four table sha256 pins + spot checks),
    `extract_test.go` (the reference's own TestWebExtract battery
    ported verbatim — article/landing/js_shell/paywall/latin1/empty/
    huge-cap — plus the package's law tests: has_content thresholds,
    the paywall window, the js-shell boundary, the script word
    boundary incl. the `<scripté` Unicode case, the backfill law, a
    never-panics hostile sweep), `tokenizer_test.go` (the focused
    event-level laws so a regression names its rule).

  - *The goldens caught five real divergences before landing* — the
    honest evidence the discipline exists for: (1) Go's stdlib
    `html.UnescapeString` diverges from Python's `html.unescape` on
    eight observable shapes (numeric refs to invalid codepoints —
    Python DROPS them to '', Go keeps the rune; `&#x;` stays literal
    in Python and becomes U+FFFD in Go) — the port carries Python's
    algorithm itself, table and all; (2) `strings.ToValidUTF8`
    collapses whole invalid runs into a single U+FFFD where Python's
    errors='replace' emits one per maximal subpart — pyUTF8Replace
    implements the exact law; (3) x/text's Big5 accepts the WHATWG
    extension leads Python rejects, and its Shift-JIS maps the 0x80
    Python errors on — the acceptance bitmaps now gate every DBCS
    decode; (4) cp949's UHC extension has no x/text equivalent — its
    table is generated and embedded; (5) even Go string literals
    betrayed the C1 row (`"\x81"` is a raw byte, not U+0081) — the
    battery caught it in the unescape spot checks. Plus the
    reality-check pass caught four wrong assumptions in my own
    understanding of the reference before a line of Go was written
    (two titles ACCUMULATE; `b` is not a block tag; attribute names
    may contain quotes; `<p>\r</p>` collapses to '').

- **T-33 (connectioncheck + sentinel, M)**: NOT started — capacity
  ran out at the natural T-32 boundary (the session law: one task
  done properly with proof beats two half-done). T-33 is a full
  battery+sentinel port (1282 LOC of reference across
  connection_check/net_guard/netctx, ~129 test selections, the
  re-scoped battery per ADR-0004 Am.3 with the Telegram section
  dropped and the Worker leg promoted) — it owns SESSION-014.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-32's task-list acceptance — PASS**:
  1. *The 5 HTML fixtures byte-compared on all Extracted fields* —
     TestGoldenFixtures drives article/landing/js_shell/paywall/
     latin1.html through ExtractFromBytes and compares title,
     meta_description, text, is_js_shell, is_paywall, charset_used,
     has_content byte-for-byte against the reference's own output
     (the fixture bytes travel inside the golden, so the driver is
     self-contained).
  2. *The parser + heuristics, the JS-shell and paywall detectors, the
     text cap, the charset re-decode* — all pinned by the 86-case
     extract battery + the 22-case bytes battery + the 11-case parser
     battery + the reference's own test battery ported verbatim.
  3. *The fourteenth generator (the dump_llmclient/dump_webfetch
     pattern)* — `testdiff/goldens/dump_webextract.py`, run against
     the unmodified reference at the pinned rev; deterministic
     (byte-identical regeneration).
- **Phase 3 exit gate — ON TRACK (the third stone laid)**:
  "fake-server contract tests green" — T-30's payload goldens, T-31's
  ladder tests, and now T-32's fixture parity are all in place; the
  HMAC loopback rig (T-34) remains.
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -count=1
  -race ./internal/...` → **16 packages ok, 713/713 tests PASS, race
  detector clean, 0 failures** (was 676 at session start); the golden
  regenerates byte-identically; go.mod/go.sum UNCHANGED (zero new
  dependencies — the authorized set is untouched).
- **Honest notes**: (1) The DBCS mapping tables (beyond the pinned
  probes) ride x/text behind the generated acceptance bitmaps — a
  mapping-table difference on an accepted-but-rare code would be a
  documented unpinned divergence (doc.go); (2) gb18030's legitimate
  U+FFFD encoding is misjudged as a gap by the FFFD scan (one
  pathological code, documented); (3) utf-16/utf-32 without a BOM
  decode little-endian (Python uses the platform's native order —
  every pinned platform is LE); (4) the tokenizer is a one-pass
  implementation of feed-whole+close — proven equivalent for that
  usage, and the only usage the reference makes.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-core-web-extract.md` —
  the Go notes amended (reality wins): the x/net/html recommendation
  now carries the SESSION-013 record — no verdict arrived, option (b)
  (the hand-rolled tokenizer) was taken, the swap path is documented,
  and the generated-tables approach for the charset ladder is
  recorded.
- `testdiff/README.md` — the generator inventory grew by one
  (`dump_webextract.py` — the fourteenth), with its golden sections
  and the four embedded data tables described.
- `internal/webextract/doc.go` — the package law doc: the hand-rolled
  tokenizer decision (the dependency-verdict record), the
  generated-tables law, the Python string-semantics laws, and every
  documented divergence.
- `README.md` — the status paragraph: T-32 landed, T-33/T-34 next.

## Blockers & decisions needed

- **The x/net/html verdict (SURFACED AGAIN, now with the full cost
  record)**: T-32 shipped WITHOUT it (option b — the zero-dependency
  alternative the session prompt itself names). The port is green and
  golden-pinned as-is; the owner may now either (a) leave it (the
  hand-rolled tokenizer is permanent) or (b) approve an ADR for
  x/net/html and a future session swaps parser.go's scan internals.
  Both paths are documented; nothing blocks on it.
- **The go-directive acknowledgement (STILL WAITING, carried from
  SESSION-012)**: the main module rides `go 1.24.1` (tls-client
  v1.16.0's own floor — ADR-0002 Am.2 has the record + the rollback
  path). A one-line acknowledgement (or a rollback instruction)
  closes it.
- The standing sandbox-reliability observation carries (eighth full
  wipe; push-after-every-task remains the load-bearing mitigation —
  T-32 was pushed the moment its battery went green, commit
  `833b21b`).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. The x/net/html verdict — now a "keep or swap" decision rather than
   a blocker (T-32 shipped on option b).
3. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-014 = Phase 3 continues — T-33
  `internal/connectioncheck` + the sentinel (M)** per
  spec-core-connection.md (the re-scoped battery: vaults, LLM,
  GitHub, the Worker leg promoted to queue-reachability; Telegram
  dropped per ADR-0004 Am.3; the re-worded summarize headline pinned
  by T-33's own result-shape tests), then T-34 `internal/workerclient`
  (the gate's last stone) if capacity remains.
- NEXT-SESSION.md rewritten: yes (v14).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.
