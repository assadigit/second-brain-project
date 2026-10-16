package connectioncheck

// config.go — the Python-dict config access layer (the llmclient
// pattern: configs arrive as *storage.OrderedMap or map[string]any and
// every read must survive hostile shapes exactly like Python's
// dict.get on junk).

import (
	"encoding/json"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// cfgGet reads a key from a config dict (either representation).
func cfgGet(config any, key string) (any, bool) {
	switch t := config.(type) {
	case nil:
		return nil, false
	case *storage.OrderedMap:
		return t.Get(key)
	case map[string]any:
		v, ok := t[key]
		return v, ok
	default:
		return nil, false
	}
}

// cfgValue is dict.get(key) — nil when absent.
func cfgValue(config any, key string) any {
	v, _ := cfgGet(config, key)
	return v
}

// pyTruthy mirrors Python truthiness for the config shapes that reach
// this package (None/False/""/0/[]/{}/{} falsy; everything else true).
func pyTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return f != 0
		}
		return true
	case []any:
		return len(t) > 0
	case *storage.OrderedMap:
		return t.Len() > 0
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}

// pyOr mirrors `x or fallback` (falsy x → fallback).
func pyOr(v, fallback any) any {
	if pyTruthy(v) {
		return v
	}
	return fallback
}

// pyStr is helpers.PyStr (Python str() for the config value shapes).
func pyStr(v any) string {
	return helpers.PyStr(v)
}

// trimPySpace is str.strip() with Python's Unicode whitespace set.
func trimPySpace(s string) string {
	return helpers.Strip(s)
}

// cfgStr is str(config.get(key, "") or "").strip() — the battery's
// dominant string read.
func cfgStr(config any, key string) string {
	return trimPySpace(pyStr(pyOr(cfgValue(config, key), "")))
}

// rstripSlashBackslash strips trailing "/" and "\" runes (the
// websites-vault parent law's p.rstrip("/\\")).
func rstripSlashBackslash(s string) string {
	return strings.TrimRight(s, "/\\")
}
