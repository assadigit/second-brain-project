package storage

// MergeConfig merges UI-derived updates into existing WITHOUT dropping
// keys (storage.py merge_config, 145-177) — the v30 fix for the
// save_config key-wipe. Rules (all observable behavior, pinned by the
// generated merge goldens):
//
//   - unknown keys in existing are PRESERVED
//   - nested dicts (e.g. 'ollama', 'proxy') are MERGED key-by-key (the
//     nested dict's own order: existing keys keep their position, new
//     keys append), never replaced wholesale
//   - keys in updates win over existing (in-place for existing keys)
//   - tuning keys (timeout_per_repo, max_retries,
//     delay_between_api_calls) are only defaulted when ABSENT — never
//     overwritten; the setdefault calls append them in this fixed order
//     when missing
func MergeConfig(existing, updates *OrderedMap) *OrderedMap {
	merged := existing.Clone() // dict(existing): shallow, order preserved
	for _, key := range updates.Keys() {
		value, _ := updates.Get(key)
		if cur, ok := merged.Get(key); ok {
			curMap, curIsDict := cur.(*OrderedMap)
			vMap, vIsDict := value.(*OrderedMap)
			if curIsDict && vIsDict {
				nested := curMap.Clone() // dict(merged[key])
				for _, nk := range vMap.Keys() {
					nv, _ := vMap.Get(nk)
					nested.Set(nk, nv) // .update(value)
				}
				merged.Set(key, nested)
				continue
			}
		}
		merged.Set(key, value)
	}
	// Tuning keys: default only, never stomp.
	defaultIfAbsent(merged, "timeout_per_repo", 60)
	defaultIfAbsent(merged, "max_retries", 3)
	defaultIfAbsent(merged, "delay_between_api_calls", 0.5)
	return merged
}

func defaultIfAbsent(m *OrderedMap, key string, value any) {
	if _, ok := m.Get(key); !ok {
		m.Set(key, value)
	}
}

// WriteConfigFile pretty-prints + atomically writes a config dict
// (storage.py write_config_file, 180-182): UTF-8, no BOM, no trailing
// newline, no HTML escaping, 2-space indent, key order preserved —
// byte-identical to Python's
// json.dumps(config, indent=2, ensure_ascii=False). Gated by dry-run
// like every canonical write.
func WriteConfigFile(path string, config *OrderedMap) error {
	return AtomicWriteText(path, PyDumps(config))
}
