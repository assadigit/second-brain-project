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
│                                   # rot-proof default-now probe)
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
