// config.go — the config doors: the load/save/user-memory layer ON TOP of
// the storage primitives (storage.DecodeJSON / storage.WriteConfigFile /
// constants.ExampleTree — the T-12 representation, not a fork).
//
// Two doors, mirroring the Python surfaces:
//   - Load/Save: the CLI door (cli/cli_settings.py load_config /
//     save_config — one source of truth with the GUI's config.json).
//   - LoadAppConfig: the app door's pure half (gui/main_window/
//     vaults_config.py load_config minus the Qt): fresh detection, the
//     example base, THE FILL, the v0.07.1 credential strip heal. The
//     save side's pure half is MergeConfig + WriteConfigFile (T-12) +
//     MirrorCredentialsToGlobal — wired by the app/CLI shells in their
//     phases.
package config

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// stripHealKeys mirrors the v0.07.1 fix's flat list: credentials pasted
// with surrounding whitespace are stripped at load so a saved or
// hand-edited config can never poison the app (the heal runs AFTER the
// fill, so remembered values are stripped under the same law).
var stripHealKeys = []string{
	"telegram_api_id", "telegram_api_hash", "telegram_phone",
	"github_token", "bot_token", "cloud_api_key",
	"cloudflare_worker_url",
}

// LoadAppResult carries what the app load path produces.
type LoadAppResult struct {
	Config *storage.OrderedMap
	Fresh  bool // config.json was missing/unreadable — example placeholders
	Filled int  // the user-memory fill count (the one quiet log line's number)
}

// LoadAppConfig mirrors the GUI load_config pure path: fresh detection
// (missing file, or read/decode failure), the example base
// (constants.ExampleTree — a deep copy, so the shared tree is never
// corrupted), THE FILL, and the credential strip heal.
//
// Quirks preserved (pinned by tests):
//   - a config.json holding literal null loads the example base with
//     Fresh=false (Python's json.load returns None, which falls into the
//     "config is None" branch without the fresh flag).
//   - a non-object config.json (a hand-mangled file) is treated as
//     unreadable → the fresh door. Python would carry the garbage value
//     onward and crash later; the port's typed tree recovers instead
//     (documented divergence — the user-visible outcome is strictly
//     better, recorded in the T-16 session report).
func LoadAppConfig(configPath string) LoadAppResult {
	fresh := !fileExistsAt(configPath)
	var config *storage.OrderedMap
	if !fresh {
		raw, err := os.ReadFile(configPath)
		if err != nil {
			fresh = true
		} else if isJSONNull(raw) {
			// json.load("null") → None → the example base, fresh stays
			config = nil
		} else {
			config, err = storage.DecodeJSON(raw)
			if err != nil {
				fresh = true
				config = nil
			}
		}
	}
	if config == nil {
		config = constants.ExampleTree() // deep copy of the defaults
	}
	filled := ApplyGlobalCredentials(config, "", fresh)

	// THE PRE-CONFIG DROP-IN LAW (SESSION-081): the OPTIONAL d1creds.json
	// file (next to config.json) carries the 3 Cloudflare-API creds the
	// D1-API fallback needs. The owner downloads ONE small file + drops it
	// in the app folder; the app reads it on startup + arms the fallback
	// without manual Direct Pair field entry. No-op when absent (the
	// common case). Only non-empty values merge (a broken drop-in is
	// ignored, never poisons the config). The no-secrets law holds: the
	// creds ride the drop-in file the owner provides, never the release.
	applyD1CredsDropin(config, configPath)

	// v0.07.1 — heal hand-edited configs (after the fill, under the same
	// law).
	for _, k := range stripHealKeys {
		if v, ok := config.Get(k); ok {
			if s, isStr := v.(string); isStr {
				config.Set(k, strings.TrimSpace(s))
			}
		}
	}
	if px, ok := config.Get("proxy"); ok {
		if pxm, isMap := px.(*storage.OrderedMap); isMap {
			if h, ok2 := pxm.Get("host"); ok2 {
				if s, isStr := h.(string); isStr {
					pxm.Set("host", strings.TrimSpace(s))
				}
			}
		}
	}
	return LoadAppResult{Config: config, Fresh: fresh, Filled: filled}
}

// isJSONNull reports whether the file's content is a bare null literal
// (whitespace-tolerant, exactly Python's json.load null case).
func isJSONNull(raw []byte) bool {
	return bytes.Equal(bytes.TrimSpace(raw), []byte("null"))
}

// Load mirrors cli/cli_settings.py load_config: read + decode config.json
// (nil, err when missing or unreadable — the CLI shell prints the
// "Could not read config" line), then THE FILL (best-effort; only
// still-blank keys — a real value in config.json is always the local
// word). A non-object config.json is an error here too (Python's load
// would return the garbage; see LoadAppConfig's note).
func Load(path string) (*storage.OrderedMap, error) {
	if !fileExistsAt(path) {
		return nil, os.ErrNotExist
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if isJSONNull(raw) {
		// Python: json.load returns None; isinstance(cfg, dict) is False
		// → no fill → the caller gets None-equivalent (nil, no error).
		return nil, nil
	}
	cfg, err := storage.DecodeJSON(raw)
	if err != nil {
		return nil, err
	}
	ApplyGlobalCredentials(cfg, "", false)
	return cfg, nil
}

// Save mirrors cli/cli_settings.py save_config: the canonical atomic
// write (storage.WriteConfigFile — dry-run aware), then THE MIRROR (the
// credentials that just landed in config.json also land in the
// user-level memory, best-effort). False when the write fails.
func Save(path string, config *storage.OrderedMap) bool {
	if err := storage.WriteConfigFile(path, config); err != nil {
		return false
	}
	MirrorCredentialsToGlobal(config, "")
	return true
}

// d1CredsDropinKeys are the 3 Cloudflare-API creds the D1-API fallback
// reads (the documented /api/pending contract's api.cloudflare.com
// access set).
var d1CredsDropinKeys = []string{
	"cloudflare_api_token",
	"cloudflare_account_id",
	"cloudflare_d1_database_id",
}

// applyD1CredsDropin reads the OPTIONAL d1creds.json drop-in (next to
// config.json) + merges the 3 Cloudflare-API creds into the live
// config when present + non-empty. THE PRE-CONFIG DROP-IN LAW
// (SESSION-081): the owner downloads ONE small file (just the 3 creds)
// + drops it in the app folder; the app reads it on startup + arms the
// D1-API fallback without manual Direct Pair field entry. No-op when
// the file is absent (the common case). Only non-empty trimmed values
// merge (a broken or partial drop-in is ignored, never poisons the
// config). The no-secrets law holds: the creds ride the drop-in file
// the owner provides, never the public release.
func applyD1CredsDropin(config *storage.OrderedMap, configPath string) {
	dropinPath := filepath.Join(filepath.Dir(configPath), "d1creds.json")
	raw, err := os.ReadFile(dropinPath)
	if err != nil {
		return // absent — the common case; no-op
	}
	dropin, derr := storage.DecodeJSON(raw)
	if derr != nil {
		return // unparseable — no-op (a broken drop-in is ignored)
	}
	for _, k := range d1CredsDropinKeys {
		if v, ok := dropin.Get(k); ok {
			if s, isStr := v.(string); isStr {
				if trimmed := strings.TrimSpace(s); trimmed != "" {
					config.Set(k, trimmed)
				}
			}
		}
	}
}
