// bridges.go — publish_from_config: build a GoodRepos from the app's
// config dict and publish once (the GUI + headless door).
package goodrepos

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// cfgGet reads a key out of a Python-dict-shaped config (nil /
// *storage.OrderedMap / map[string]any — the workerclient pattern).
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

// cfgValue is cfgGet without the ok.
func cfgValue(config any, key string) any {
	v, _ := cfgGet(config, key)
	return v
}

// dictOfValue answers the dict-shaped value at key (or nil).
func dictOfValue(config any, key string) any {
	switch t := cfgValue(config, key).(type) {
	case *storage.OrderedMap:
		return t
	case map[string]any:
		return t
	default:
		return nil
	}
}

// pyStrOf mirrors str(v or "").
func pyStrOf(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		if !pyTruthy(t) {
			return ""
		}
		return pyStrOrEmpty(t)
	}
}

// pyBoolOf mirrors bool(v) over the config shapes (Python truthiness).
func pyBoolOf(v any) bool { return pyTruthy(v) }

// PublishFromConfig mirrors publish_from_config: reads vault_path,
// github_token, and the “goodrepos“ section ({enabled, repo_name,
// auto_push, create_repo}); v0.83.0 — an empty repo_name means NOT
// configured: publish skips.
func PublishFromConfig(config any, runSummary *RunSummary, log LogFn) *PublishResult {
	cfg := dictOfValue(config, "goodrepos")
	publisher := NewGoodRepos(
		pyStrOf(cfgValue(config, "vault_path")),
		pyStrOf(cfgValue(config, "github_token")),
		pyStrOf(cfgValue(cfg, "repo_name")),
		cfgBoolDefault(cfg, "auto_push", true),
		cfgBoolDefault(cfg, "enabled", true),
		0, log, cfgBoolDefault(cfg, "create_repo", false),
	)
	return publisher.Publish(runSummary, false)
}

// cfgBoolDefault mirrors cfg.get(key, default) through bool().
func cfgBoolDefault(cfg any, key string, def bool) bool {
	switch t := cfg.(type) {
	case nil:
		return def
	case *storage.OrderedMap:
		v, ok := t.Get(key)
		if !ok || v == nil {
			return def
		}
		return pyTruthy(v)
	case map[string]any:
		v, ok := t[key]
		if !ok || v == nil {
			return def
		}
		return pyTruthy(v)
	default:
		return def
	}
}

// ensure bridges.go keeps the strings import honest for future
// config-grammar work.
var _ = strings.TrimSpace
