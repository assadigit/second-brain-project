package directorysync

// export.go — the tiny exported twins of the walk helpers the
// syncreview package's vault-side doors reuse (sync_review.py's own
// law: "mirroring build_index's own filter laws … reuse, never
// re-invented"). No behavior change — pure delegations.

// IterNotes is _iter_notes' exported twin: (relativePath, absPath) for
// every .md under the vault's Category/Subcategory tree — system
// folders and dot-dirs skipped, the directory note itself never
// indexed, sorted for determinism.
func IterNotes(vault string, visit func(rel, absPath string)) {
	iterNotes(vault, visit)
}

// ReadText is _read_text's exported twin: UTF-8 with errors='replace',
// universal newlines, capped at the index's own character limit.
func ReadText(path string) string { return readText(path) }

// NormHost is _norm_host's exported twin: the host out of a raw URL
// string, www. stripped.
func NormHost(rawURL string) string { return normHost(rawURL) }

// SharedHostsOf is the config read's exported twin (the banish wall's
// shared-host exceptions, lowercased — DefaultSharedHosts when unset).
func SharedHostsOf(config any) []string { return sharedHostsOf(config) }
