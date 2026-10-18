package cli

// cfgutil.go — the Python-dict access laws over the house OrderedMap
// (the or-door, the truthiness door, the nested get) shared by the CLI
// commands. Every accessor mirrors the exact reference spelling it
// serves, so the falsy doors (0 → "(not set)", "" → default, None →
// "") behave byte-for-byte.

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// getOr mirrors dict.get(key, default): the stored value even when
// nil, the default when absent.
func getOr(cfg *storage.OrderedMap, key string, def any) any {
	if cfg == nil {
		return def
	}
	if v, ok := cfg.Get(key); ok {
		return v
	}
	return def
}

// getMap mirrors (cfg.get(key) or {}): the nested dict, or an empty
// one for nil/missing/non-dict values.
func getMap(cfg *storage.OrderedMap, key string) *storage.OrderedMap {
	v := getOr(cfg, key, nil)
	if om, ok := v.(*storage.OrderedMap); ok {
		return om
	}
	return storage.NewOrderedMap()
}

// pyTruthy mirrors Python truthiness for the JSON-decoded shapes:
// nil/false/0/0.0/""/empty dict are falsy.
func pyTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case int64:
		return t != 0
	case float64:
		return t != 0
	case string:
		return t != ""
	case *storage.OrderedMap:
		return t != nil && t.Len() > 0
	default:
		return true
	}
}

// orStr mirrors str(value or ""): the falsy door then str().
func orStr(v any) string {
	if !pyTruthy(v) {
		return ""
	}
	return helpers.PyStr(v)
}

// cfgOrStr mirrors str(cfg.get(key, "") or "").
func cfgOrStr(cfg *storage.OrderedMap, key string) string {
	return orStr(getOr(cfg, key, ""))
}

// mapOrStr mirrors str(oll.get(key, "") or "") over a nested dict.
func mapOrStr(m *storage.OrderedMap, key string) string {
	if m == nil {
		return ""
	}
	return orStr(getOr(m, key, ""))
}

// cfgProvider mirrors str(cfg.get("llm_provider", "ollama") or
// "ollama").strip().lower().
func cfgProvider(cfg *storage.OrderedMap) string {
	v := getOr(cfg, "llm_provider", "ollama")
	if !pyTruthy(v) {
		v = "ollama"
	}
	return strings.ToLower(pyStrip(helpers.PyStr(v)))
}

// lstripAt mirrors .lstrip("@") — strips ALL leading '@' runes.
func lstripAt(s string) string { return strings.TrimLeft(s, "@") }
