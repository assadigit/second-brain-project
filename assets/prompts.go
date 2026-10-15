package assets

import "embed"

// PromptsFS holds the seven live prompt templates, copied verbatim from
// the Python reference's app/prompts/ by testdiff/goldens/dump_prompts.py
// (never hand-edited; sha256-pinned in
// internal/prompts/testdata/prompts_corpus.golden.json).
//
// The dead 01/02/03 prompt files are deliberately NOT embedded: they have
// no consumer (the GitHub prompt lives in constants) — see
// docs/port-blueprint/module-specs/spec-core-prompts.md §Prompt inventory.
//
// Paths inside the FS are rooted at "prompts/<name>.txt" — the virtual
// equivalent of the reference's PROMPTS_DIR (APP_DIR/prompts), which the
// error messages use after dump-time normalization.
//
//go:embed prompts/*.txt
var PromptsFS embed.FS
