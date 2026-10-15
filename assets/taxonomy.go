package assets

import _ "embed"

// TaxonomyDefault is the SHIPPED Websites category taxonomy
// (app/taxonomy/website-library-categories.md in the reference), copied
// verbatim by testdiff/goldens/dump_taxonomy.py — never hand-edited;
// sha256-pinned in internal/taxonomy/testdata/taxonomy_corpus.golden.json
// (shipped_snapshot).
//
// Per doc 04 §1's assets layout the default ships embedded; the PARITY
// loaders stay disk-based exactly like Python (constants.DefaultTaxonomyPath
// + taxonomy.LoadTaxonomyFromConfig) — the app-composition wiring that
// consumes these bytes (first-run extraction or direct parse) joins in a
// later phase; until then this embed is the blueprint-mandated home of the
// generated copy.
//
//go:embed taxonomy/website-library-categories.md
var TaxonomyDefault []byte
