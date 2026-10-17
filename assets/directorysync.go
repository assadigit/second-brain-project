package assets

import _ "embed"

// WebsiteDirectoryConfig is the SHIPPED Websites directory default
// config (app/taxonomy/website-directory-config.json in the
// reference), copied verbatim by testdiff/goldens/dump_vaultseal.py —
// never hand-edited; sha256-pinned by the vaultseal golden
// (belt_floor section).
//
// The reference's load_directory_config reads this file from disk at
// runtime (app/taxonomy/); the Go binary embeds the same bytes as the
// shipped default (doc 04 §1's assets layout) — a documented
// convergence recorded in spec-integrations-vaultseal.md's Go notes.
// T-49's directorysync port reuses this embed for the config trio.
//
//go:embed directorysync/website-directory-config.json
var WebsiteDirectoryConfig []byte

// SearchSynonyms is the SHIPPED search-synonyms default
// (app/taxonomy/search-synonyms.json in the reference), copied
// verbatim — never hand-edited; sha256-pinned by the directorysync
// golden (system_files section). The same embed convergence as
// WebsiteDirectoryConfig above (T-49's config trio).
//
//go:embed directorysync/search-synonyms.json
var SearchSynonyms []byte
