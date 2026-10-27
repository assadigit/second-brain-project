package workerclient

// config.go — the Sync handle (CloudflareSync): the live config object
// (pairing WRITES into it), the enabled/paired gates, and the package's
// Python-dict access helpers (the house pattern).

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ConfigKeys mirrors CONFIG_KEYS (stored in config.json).
var ConfigKeys = map[string]any{
	"cloudflare_worker_url":    "",
	"cloudflare_install_id":    "",
	"cloudflare_shared_secret": "",
	"cloudflare_enabled":       false,
	"cloudflare_poll_interval": 300,
}

// Sync handles all communication with the Cloudflare Worker
// (the CloudflareSync class). config is the LIVE config object —
// Pair writes the pairing keys into it (either *storage.OrderedMap or
// map[string]any).
type Sync struct {
	config   any
	mu       chanMutex
	lastPoll float64
}

// chanMutex serializes polls/pushes exactly where the reference takes
// its threading.Lock (a plain sync.Mutex would be identical; the named
// type documents the law).
type chanMutex struct {
	ch chan struct{}
}

func newChanMutex() chanMutex {
	ch := make(chan struct{}, 1)
	ch <- struct{}{}
	return chanMutex{ch: ch}
}

func (m *chanMutex) lock()   { <-m.ch }
func (m *chanMutex) unlock() { m.ch <- struct{}{} }

// NewSync builds the client over the live config object.
func NewSync(config any) *Sync {
	return &Sync{config: config, mu: newChanMutex()}
}

// WorkerURL is the configured Worker root (trailing "/" trimmed).
func (s *Sync) WorkerURL() string {
	return trimRightSlash(pyStr(pyOr(cfgValue(s.config,
		"cloudflare_worker_url"), "")))
}

// InstallID is the pairing's install UUID.
func (s *Sync) InstallID() string {
	return pyStr(pyOr(cfgValue(s.config, "cloudflare_install_id"), ""))
}

// SharedSecret is the pairing's 64-hex secret.
func (s *Sync) SharedSecret() string {
	return pyStr(pyOr(cfgValue(s.config, "cloudflare_shared_secret"), ""))
}

// IsEnabled: the master switch AND the URL AND the install_id AND the
// secret — all four, or the channel stays down.
func (s *Sync) IsEnabled() bool {
	return pyTruthy(cfgValue(s.config, "cloudflare_enabled")) &&
		s.WorkerURL() != "" &&
		s.InstallID() != "" &&
		s.SharedSecret() != ""
}

// disabledReason returns a human-readable string naming WHICH key is
// missing when IsEnabled returns false (the diagnostic law — the owner
// sees exactly what to fix, not just "not configured").
func (s *Sync) DisabledReason() string {
	var missing []string
	if !pyTruthy(cfgValue(s.config, "cloudflare_enabled")) {
		missing = append(missing, "cloudflare_enabled is OFF")
	}
	if s.WorkerURL() == "" {
		missing = append(missing, "cloudflare_worker_url is empty")
	}
	if s.InstallID() == "" {
		missing = append(missing, "cloudflare_install_id is empty (not paired)")
	}
	if s.SharedSecret() == "" {
		missing = append(missing, "cloudflare_shared_secret is empty (not paired)")
	}
	if len(missing) == 0 {
		return "all keys set but IsEnabled still false (internal error)"
	}
	return strings.Join(missing, "; ")
}

// IsPaired: install_id AND shared_secret (regardless of the switch).
func (s *Sync) IsPaired() bool {
	return s.InstallID() != "" && s.SharedSecret() != ""
}

// ---------------------------------------------------------------------------
// The Python-dict access helpers (the house pattern).
// ---------------------------------------------------------------------------

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

func cfgValue(config any, key string) any {
	v, _ := cfgGet(config, key)
	return v
}

// cfgSet writes into the live config object (Pair's writes).
func cfgSet(config any, key string, value any) {
	switch t := config.(type) {
	case *storage.OrderedMap:
		t.Set(key, value)
	case map[string]any:
		t[key] = value
	}
}

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

func pyOr(v, fallback any) any {
	if pyTruthy(v) {
		return v
	}
	return fallback
}

// pyStr mirrors str() for the config value shapes.
func pyStr(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "True"
		}
		return "False"
	default:
		return helpers.PyStr(v)
	}
}

// isDictLike reports whether v is a Python-dict-shaped value.
func isDictLike(v any) bool {
	switch v.(type) {
	case *storage.OrderedMap, map[string]any:
		return true
	}
	return false
}

func trimRightSlash(s string) string {
	for len(s) > 0 && s[len(s)-1] == '/' {
		s = s[:len(s)-1]
	}
	return s
}

// dictValue reads a key from a decoded JSON dict value.
func dictValue(m any, key string) any {
	v, _ := cfgGet(m, key)
	return v
}
