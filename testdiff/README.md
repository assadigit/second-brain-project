# testdiff/ — the differential (Python-vs-Go) harness

The parity evidence machine (blueprint `09-testing-and-parity.md` §3):
the **unmodified** Python app (`assadigit/GitCurator`, pinned rev in
`PINNED-GITCURATOR-REV`) and the Go port run the same commands on the same
inputs in the same scratch cwd; a normalizing comparator diffs
stdout/stderr/exit code (and, in later phases, note files, DB tables and
manifests) and writes a per-case report with sha256s of the raw outputs.

```
testdiff/
├── harness.py                     # the runner (drivers both sides + comparator)
├── PINNED-GITCURATOR-REV          # the reference commit the corpus is pinned to
├── cases/<case>/case.json         # case definitions (args per side + env)
├── goldens/                       # GENERATORS that dump the unmodified Python
│                                   # app into Go testdata/ goldens (Phase 1 law:
│                                   # goldens are generated, never hand-typed;
│                                   # e.g. dump_constants.py → internal/constants,
│                                   # dump_links.py → internal/links,
│                                   # dump_storage.py → internal/storage +
│                                   # internal/dryrun, T-12;
│                                   # dump_prompts.py → internal/prompts +
│                                   # assets/prompts, T-13;
│                                   # dump_taxonomy.py → internal/taxonomy +
│                                   # assets/taxonomy, T-14;
│                                   # dump_notebuilder.py → internal/notebuilder,
│                                   # T-15;
│                                   # dump_config.py → internal/config +
│                                   # internal/constants (ordered example),
│                                   # T-16;
│                                   # dump_helpers.py → internal/helpers,
│                                   # T-17;
│                                   # dump_store.py → internal/store,
│                                   # T-20 — incl. the Python-created
│                                   # cache.db fixture, generated as an
│                                   # ATOMIC PAIR with its content
│                                   # golden via fixture_sha256: both
│                                   # regenerate together, the sha ties
│                                   # them; the other store goldens
│                                   # (timestamp grammar, schema) are
│                                   # byte-stable);
│                                   # dump_notestate.py →
│                                   # internal/notestate, T-21 — the
│                                   # pure corpus (fingerprint incl.
│                                   # the invalid-UTF-8 replacement
│                                   # battery + folder_category
│                                   # matrix, byte-stable) + the
│                                   # Python-created notestate DB
│                                   # fixture as an ATOMIC PAIR with its
│                                   # content golden + the schema golden
│                                   # (the CAST-law guard: all-TEXT
│                                   # declared columns);
│                                   # dump_notereads.py →
│                                   # internal/notereads, T-22 — the
│                                   # pure corpus (the unified parse +
│                                   # the body-tag grammar + the truthy
│                                   # fold + the text-mode read law
│                                   # incl. universal newlines and the
│                                   # maximal-subpart replacement +
│                                   # the shape adapters + the row-JSON
│                                   # reader WITH the stars-quirk
│                                   # evidence, byte-stable) + the
│                                   # Python-created note_reads DB
│                                   # fixture as an ATOMIC PAIR with
│                                   # its content golden (the
│                                   # frontmatter_json bytes are the
│                                   # Go JSON writer's byte-parity
│                                   # pin) + the schema golden (the
│                                   # CAST-law guard: all
│                                   # TEXT/INTEGER/REAL columns) + the
│                                   # behavior golden (the reference's
│                                   # own probe counts: second-scan
│                                   # opens 4→0 / db_hits 6, stale
│                                   # rows, prune, broken db, the
│                                   # character-budget baseline the
│                                   # Go byte-budget fix diverges
│                                   # from);
│                                   # dump_linktracker.py →
│                                   # internal/batch, T-25 — the
│                                   # pure corpus (manifest_bytes:
│                                   # the reference's own atomic
│                                   # writer output over a FIXED
│                                   # manifest — the Go writer's
│                                   # json.dump(indent=2) byte-parity
│                                   # pin, null fields + unicode
│                                   # escapes included; the
│                                   # SafeMocName battery) + the
│                                   # env-steered obsidian discovery
│                                   # golden + the behavior golden
│                                   # (the VaultIndex skip sets,
│                                   # quote stripping, the 4000-char
│                                   # read boundary, the site_links
│                                   # grammar, the site-map
│                                   # registration gates, add_url's
│                                   # incremental learn);
│                                   # dump_cacherepos.py →
│                                   # internal/store/repos.go, T-24 —
│                                   # the pure corpus (the
│                                   # failed_rows_truth classifier
│                                   # battery + the threshold clamp
│                                   # battery, byte-stable) + the
│                                   # behavior golden (the open-time
│                                   # migrations with the legacy DDL
│                                   # recorded verbatim for byte-exact
│                                   # replay, THE SETTLED REPOS
│                                   # LEDGER, the ♻️ revive door,
│                                   # record_404's FAIL-CLOSED, the
│                                   # retry queue's laws, threshold
│                                   # steering, purge_old, the legacy
│                                   # decommission across a reopen);
│                                   # REUSES T-20's Python-created
│                                   # cache fixture for the fixture
│                                   # acceptance;
│                                   # dump_websitestate.py →
│                                   # internal/websitepipeline, T-23 —
│                                   # the pure corpus (the _key
│                                   # one-spelling boundary battery +
│                                   # the 2×attempts backoff table +
│                                   # the rune-truncation battery,
│                                   # byte-stable) + the Python-created
│                                   # website-state DB fixture as an
│                                   # ATOMIC PAIR with its content
│                                   # golden + the schema golden (the
│                                   # CAST-law guard: all TEXT/INTEGER
│                                   # columns) + the behavior golden
│                                   # (the healing matrix, both
│                                   # settlements, and the due_retries
│                                   # battery at an injected fixed now
│                                   # — incl. the aware-datetime
│                                   # TypeError→due quirk and the
│                                   # rot-proof default-now probe);
│                                   # dump_llmclient.py →
│                                   # internal/llmclient, T-30 — the
│                                   # pure corpus (extract_json incl.
│                                   # top-level non-objects,
│                                   # estimate_tokens, the URL
│                                   # normalizer, the loopback table
│                                   # incl. the v4-mapped corner, the
│                                   # /props batteries incl. the
│                                   # Python int() quirks, the
│                                   # autodetect decision table with
│                                   # the exact emoji messages, the
│                                   # canned tasklist/netstat/ss
│                                   # parses, the constants) + THE
│                                   # PAYLOAD GOLDENS (the exact wire
│                                   # requests — method, path,
│                                   # app-chosen headers with the
│                                   # GitCurator-detect UA, the RAW
│                                   # bodies in Python json.dumps
│                                   # compact shape — for all three
│                                   # providers, the response_format
│                                   # 400-fallback + memoization, and
│                                   # the ollama request-building law
│                                   # via a recording fake client) +
│                                   # the behavior golden (the probe
│                                   # matrix, detect_llamacpp
│                                   # provenance, detect_ollama
│                                   # details, preflight, the
│                                   # error-family messages, the
│                                   # over-budget warnings — per-run
│                                   # ports normalized to
│                                   # <ROOT>/<API>/<DEADURL>)
│                                   #
│                                   # dump_webfetch.py →
│                                   # internal/webfetch, T-31 — the
│                                   # THIRTEENTH generator: the pure
│                                   # corpus (constants, the proxy
│                                   # config contract, the Chrome-131
│                                   # header set WITH ORDER, the
│                                   # verdict table, Retry-After, the
│                                   # URL variants, decompression,
│                                   # decode, penalties, the
│                                   # TLS/DNS marker matrices), THE
│                                   # PAYLOAD GOLDENS (the exact wire
│                                   # requests: the plain fetch's
│                                   # header set, the redirect chain
│                                   # with hop cookies + the Referer
│                                   # law LIVE, the custom UA, the DoH
│                                   # probe, the Wayback availability
│                                   # probe, the http-proxy absolute
│                                   # URI), and the behavior golden
│                                   # (the redirect truth, the phase-2
│                                   # battery, DoH verdicts, the
│                                   # rescue ladder, rate penalties,
│                                   # the origin-down courtesy, the
│                                   # third door's trigger matrix +
│                                   # verdicts + the autopip laws, the
│                                   # both-doors table, the preflight
│                                   # facts — per-run ports
│                                   # normalized to
│                                   # <ROOT>/<ROOT2>/<DOH>/<ARCHIVE>/
│                                   # <PROXY>/<DEADPORT>)
│                                   #
│                                   # dump_webextract.py →
│                                   # internal/webextract, T-32 — the
│                                   # FOURTEENTH generator: the pure
│                                   # corpus (the constants, the 5
│                                   # HTML fixtures byte-compared on
│                                   # every Extracted field, the
│                                   # 86-case extract() battery —
│                                   # title/meta precedence incl.
│                                   # first-wins og/name, the meta
│                                   # elif chain, duplicate attrs,
│                                   # charset shapes, raw-mode
│                                   # xmp/iframe, RCDATA title,
│                                   # comments/PIs/CDATA/marked
│                                   # sections, the stray-endtag
│                                   # newline law, the incomplete-tag
│                                   # discard laws, block-tag
│                                   # newline collapse, nbsp/CR
│                                   # survival, the 20k cap, the
│                                   # 120-word js-shell boundary, the
│                                   # 4000-char paywall window, the
│                                   # <script\b> word-boundary law —
│                                   # the parser-level battery, the
│                                   # 95-case unescape battery (the
│                                   # eight spots where Go's stdlib
│                                   # html.UnescapeString diverges
│                                   # from Python's html.unescape are
│                                   # pinned), the utf-8
│                                   # errors='replace' maximal-subpart
│                                   # law, the 74-name charset ladder
│                                   # (resolution, the strict-decode
│                                   # law, per-byte holes, full-256
│                                   # decodes), the extract_from_bytes
│                                   # battery (header/sniff/replace
│                                   # flows, BOM, the windows-1252
│                                   # hole fallback, the charset_used
│                                   # backfill law) — PLUS the three
│                                   # EMBEDDED DATA TABLES it emits:
│                                   # entities_html5.json (the HTML5
│                                   # named-reference table from the
│                                   # reference runtime's own
│                                   # html.entities.html5), charsets_
│                                   # table.json + cp949_table.json
│                                   # (single-byte and UHC-Korean
│                                   # decode tables from the reference
│                                   # runtime's own codecs), and
│                                   # dbcs_accept.json (the DBCS
│                                   # acceptance bitmaps that gate
│                                   # x/text — Python's codecs accept
│                                   # strictly less than x/text on
│                                   # Big5/Shift-JIS/gb2312); each
│                                   # table pinned at sha256 level)
└── runs/                          # per-run artifacts — gitignored, copy the
                                    # signed-off report into docs/poc-artifacts/
```

Run:

```sh
python3 testdiff/harness.py                     # all cases
python3 testdiff/harness.py cases/status-no-config
```

Overrides: `GC_PYTHON` (venv python), `GC_APP_DIR` (reference app dir),
`GO_BIN` (prebuilt Go CLI), `GO` (toolchain). The harness builds
`cmd/gitcurator-cli` itself by default.

## Normalization rules (doc 09 §3, implemented in harness.py)

| rule | shape | replaced by |
|---|---|---|
| timestamp | `HH:MM:SS` | `<TS>` |
| timestamp-iso | ISO 8601 shapes | `<TS>` |
| temp-path | the run's scratch dir | `<TMP>` |
| duration | `123ms`, `1.2 s`, `5µs` | `<DUR>` |
| ua-version | `GitCurator/v…` in UA-bearing logs | `<VER>` |

Everything else must match byte-exactly. Documented acceptable differences
(unsorted-walk ordering effects etc.) get pinned per case once they appear,
with the winner noted — never silently normalized.

## Case inventory

| case | phase | status |
|---|---|---|
| `status-no-config` | 0 (T-03 skeleton proof) | **green** — report archived at `docs/poc-artifacts/t-03-status-no-config-report.md` |
| (golden) `import-file grammar` | 1 (T-11) | **green** — realized in its spec-defined form (`spec-core-links` §Go mapping: Python `parse_import_text` vs Go on the same corpus, JSON-compared): `internal/links` `TestGoldenImportGrammar` over the generated `testdata/import_corpus.txt` + `link_grammar.golden.json`. A CLI-level `--import-file` case joins when the flag surface exists (Phase 5, T-62). |

Planned (per doc 09 §3 scope): 30-link golden websites set, moves-as-corrections
scenarios, banishment verdicts, scan plan apply, retry backoff clock, dry-run report
equality — cases grow with
Phases 1–5; each case dir gains its input set (URLs, fake fetch bodies,
fake LLM answers, config.json, empty vault tree) as those inputs exist.
