// Config door tests — T-16 characterization (Phase 1): the CLI load/save
// doors (ported from app/tests/test_userconfig.py TestCliDoor) and the
// app load path's pure halves (from TestGuiUserConfig — fresh detection,
// example base, the fill, the v0.07.1 credential strip heal), plus the
// ordered-example golden (write_config_file(CONFIG_EXAMPLE) bytes from
// the reference — content AND key order, recursively).
package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// Ported: TestCliDoor (app/tests/test_userconfig.py).
// ---------------------------------------------------------------------------

func writeLocalConfig(t *testing.T, s *memSandbox, cfg string) string {
	t.Helper()
	path := filepath.Join(s.tmp, "cli-config.json")
	if err := os.WriteFile(path, []byte(cfg), 0o644); err != nil {
		t.Fatalf("write local config: %v", err)
	}
	return path
}

// Port of test_cli_load_fills_blank_keys.
func TestCLILoadFillsBlankKeys(t *testing.T) {
	s := newMemSandbox(t)
	s.remember(`{"github_token": "ghp_memory", "cloud_api_key": "sk-memory"}`)
	path := writeLocalConfig(t, s,
		`{"github_token": "ghp_LOCAL", "cloud_api_key": ""}`)
	cfg, err := Load(path)
	if err != nil {
		t.Fatalf("load_config: %v", err)
	}
	if v, _ := cfg.Get("github_token"); v != "ghp_LOCAL" {
		t.Errorf("github_token = %v (local wins)", v)
	}
	if v, _ := cfg.Get("cloud_api_key"); v != "sk-memory" {
		t.Errorf("cloud_api_key = %v (blank filled)", v)
	}
}

// Port of test_cli_load_missing_file_is_none.
func TestCLILoadMissingFileIsNone(t *testing.T) {
	newMemSandbox(t)
	cfg, err := Load(filepath.Join(t.TempDir(), "nope.json"))
	if cfg != nil || err == nil {
		t.Fatalf("load_config(missing) = (%v, %v), want (nil, err)", cfg, err)
	}
}

// Port of test_cli_save_mirrors.
func TestCLISaveMirrors(t *testing.T) {
	s := newMemSandbox(t)
	path := writeLocalConfig(t, s, `{}`)
	if !Save(path, treeFromJSON(t, `{"github_token": "ghp_cli", "vault_path": "/nowhere"}`)) {
		t.Fatal("save_config returned false")
	}
	mem := treeFromJSON(t, string(s.memoryJSON()))
	if v, _ := mem.Get("github_token"); v != "ghp_cli" {
		t.Errorf("github_token = %v", v)
	}
	if v, _ := mem.Get("vault_path"); v != "/nowhere" {
		t.Errorf("vault_path = %v (remembered like a credential since v0.83.0)", v)
	}
}

// ---------------------------------------------------------------------------
// The app load path (the GUI load_config pure halves; the Qt halves are
// Phase 6's rehost).
// ---------------------------------------------------------------------------

// The fresh-version door: config.json missing → the example base + the
// fresh fill (every remembered key restored over the placeholders).
func TestLoadAppConfigFreshRestores(t *testing.T) {
	s := newMemSandbox(t)
	s.remember(`{
          "github_token": "ghp_real", "telegram_api_id": 24611234,
          "telegram_api_hash": "hash", "vault_path": "G:/vault"
        }`)
	cfgPath := filepath.Join(s.tmp, "config.json")
	res := LoadAppConfig(cfgPath)
	if !res.Fresh {
		t.Fatal("fresh not detected for a missing config.json")
	}
	if v, _ := res.Config.Get("github_token"); v != "ghp_real" {
		t.Errorf("github_token = %v", v)
	}
	if v, _ := res.Config.Get("vault_path"); v != "G:/vault" {
		t.Errorf("vault_path = %v", v)
	}
	if v, _ := res.Config.Get("llm_provider"); v != "llamacpp" {
		t.Errorf("llm_provider = %v (the example's default)", v)
	}
	if res.Filled < 3 {
		t.Errorf("filled = %d", res.Filled)
	}
}

// An existing config.json loads as-is (no example merge) and only unset
// keys fill.
func TestLoadAppConfigExistingOnlyFillsBlanks(t *testing.T) {
	s := newMemSandbox(t)
	s.remember(`{"github_token": "ghp_memory", "cloud_api_key": "sk"}`)
	cfgPath := filepath.Join(s.tmp, "config.json")
	body := `{"github_token": "ghp_LOCAL", "llm_provider": "cloud"}`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	res := LoadAppConfig(cfgPath)
	if res.Fresh {
		t.Fatal("fresh misdetected")
	}
	if v, _ := res.Config.Get("github_token"); v != "ghp_LOCAL" {
		t.Errorf("github_token = %v", v)
	}
	if v, _ := res.Config.Get("llm_provider"); v != "cloud" {
		t.Errorf("llm_provider = %v (unknown keys kept, no example merge)", v)
	}
	if v, _ := res.Config.Get("cloud_api_key"); v != "sk" {
		t.Errorf("cloud_api_key = %v (blank filled from the memory)", v)
	}
}

// A broken config.json → the fresh door (the whole local config is
// example placeholders, so every remembered key restores).
func TestLoadAppConfigBrokenIsFresh(t *testing.T) {
	s := newMemSandbox(t)
	s.remember(`{"github_token": "ghp_real"}`)
	cfgPath := filepath.Join(s.tmp, "config.json")
	if err := os.WriteFile(cfgPath, []byte("{ not json"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := LoadAppConfig(cfgPath)
	if !res.Fresh {
		t.Fatal("broken config not treated as fresh")
	}
	if v, _ := res.Config.Get("github_token"); v != "ghp_real" {
		t.Errorf("github_token = %v", v)
	}
	if v, _ := res.Config.Get("timeout_per_repo"); fmt.Sprint(v) != "60" {
		t.Errorf("timeout_per_repo = %v (the example's default)", v)
	}
}

// The v0.07.1 heal: credentials pasted with surrounding whitespace are
// stripped at load (a trailing newline must never reach the file or the
// API headers again) — INCLUDING the remembered values (the heal runs
// after the fill, under the same law).
func TestLoadAppConfigStripsCredentialWhitespace(t *testing.T) {
	s := newMemSandbox(t)
	s.remember(`{"github_token": "ghp_real \n"}`)
	cfgPath := filepath.Join(s.tmp, "config.json")
	body := `{"telegram_api_hash": " hash\n", "github_token": " local \t", "cloudflare_worker_url": " https://bot/x "}`
	if err := os.WriteFile(cfgPath, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	res := LoadAppConfig(cfgPath)
	for _, k := range []string{"telegram_api_hash", "github_token", "cloudflare_worker_url"} {
		v, _ := res.Config.Get(k)
		if s, ok := v.(string); !ok || s != strings.TrimSpace(s) {
			t.Errorf("%s not stripped: %q", k, v)
		}
	}
	// the memory's value was filled first, then stripped
	// github_token was REAL locally (" local \t") → local wins, then strip
	if v, _ := res.Config.Get("github_token"); v != "local" {
		t.Errorf("github_token = %q, want %q", v, "local")
	}
	// proxy.host strips too
	proxied := `{"proxy": {"host": " 127.0.0.1 "}}`
	if err := os.WriteFile(cfgPath, []byte(proxied), 0o644); err != nil {
		t.Fatal(err)
	}
	res = LoadAppConfig(cfgPath)
	px, _ := res.Config.Get("proxy")
	if v, _ := px.(*storage.OrderedMap).Get("host"); v != "127.0.0.1" {
		t.Errorf("proxy.host = %q, want stripped", v)
	}
}

// A config.json holding literal null behaves like Python: the example
// base with the NON-fresh fill (documented quirk).
func TestLoadAppConfigNullFile(t *testing.T) {
	s := newMemSandbox(t)
	s.remember(`{"github_token": "ghp_real"}`)
	cfgPath := filepath.Join(s.tmp, "config.json")
	if err := os.WriteFile(cfgPath, []byte("null"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := LoadAppConfig(cfgPath)
	if res.Fresh {
		t.Fatal("null file: fresh must stay false (Python json.load returns None)")
	}
	if v, _ := res.Config.Get("llm_provider"); v != "llamacpp" {
		t.Fatalf("null file did not fall back to the example: %v", v)
	}
	if v, _ := res.Config.Get("github_token"); v != "ghp_real" {
		t.Errorf("github_token = %v (the fill still ran)", v)
	}
}

// ---------------------------------------------------------------------------
// The ordered-example golden (T-16): PyDumps(ExampleTree()) must equal
// the reference's write_config_file(CONFIG_EXAMPLE) bytes — content AND
// key order, recursively.
// ---------------------------------------------------------------------------

func TestGoldenExampleOrdered(t *testing.T) {
	want := string(goldenBytesConfig(t, "config_example_ordered.golden.json"))
	got := storage.PyDumps(constants.ExampleTree())
	if got != want {
		t.Fatalf("PyDumps(ExampleTree()) drifted from the reference's\n"+
			"write_config_file(CONFIG_EXAMPLE) bytes.\n--- got ---\n%s\n--- want ---\n%s", got, want)
	}
}

func TestGoldenExampleKeys(t *testing.T) {
	doc := loadCorpusConfig(t)
	var want []string
	for _, k := range corpusSectionConfig(t, doc, "example_keys") {
		want = append(want, corpusStr(t, k))
	}
	if got := constants.ExampleTree().Keys(); fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("example keys = %v, want %v", got, want)
	}
}

// ExampleTree must return a fresh deep copy (the aliasing law's other
// half — the Python suite's TestConfigExampleDeepCopy for the ordered
// view).
func TestExampleTreeDeepCopy(t *testing.T) {
	a := constants.ExampleTree()
	px, _ := a.Get("proxy")
	px.(*storage.OrderedMap).Set("enabled", "poisoned")
	a.Set("github_token", "poisoned")
	b := constants.ExampleTree()
	if v, _ := b.Get("github_token"); v != "" {
		t.Fatalf("flat mutation leaked: %v", v)
	}
	pxb, _ := b.Get("proxy")
	if v, _ := pxb.(*storage.OrderedMap).Get("enabled"); v != false {
		t.Fatalf("nested mutation leaked: %v", v)
	}
}

// TestApplyD1CredsDropin pins THE PRE-CONFIG DROP-IN LAW (SESSION-081):
// d1creds.json (next to config.json) merges the 3 Cloudflare-API creds
// into the live config when present + non-empty. Absent → no-op.
// Partial → only the present creds merge. Unparseable → no-op.
func TestApplyD1CredsDropin(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.json")
	write := func(name, body string) {
		if err := os.WriteFile(filepath.Join(dir, name), []byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	read := func() *storage.OrderedMap {
		res := LoadAppConfig(cfgPath)
		return res.Config
	}
	cred := func(cfg *storage.OrderedMap, key string) string {
		v, _ := cfg.Get(key)
		s, _ := v.(string)
		return s
	}

	// Base config: no D1 creds.
	write("config.json", `{"cloudflare_worker_url": "https://bot.example"}`)

	// 1. No drop-in → creds stay empty (the common case).
	c := read()
	for _, k := range d1CredsDropinKeys {
		if got := cred(c, k); got != "" {
			t.Fatalf("absent drop-in: %s = %q; want empty", k, got)
		}
	}

	// 2. Full drop-in → all 3 merge.
	write("d1creds.json", `{
                "cloudflare_api_token": "tok-xyz",
                "cloudflare_account_id": "acct-xyz",
                "cloudflare_d1_database_id": "db-xyz"
        }`)
	c = read()
	for _, k := range d1CredsDropinKeys {
		want := map[string]string{
			"cloudflare_api_token":      "tok-xyz",
			"cloudflare_account_id":     "acct-xyz",
			"cloudflare_d1_database_id": "db-xyz",
		}[k]
		if got := cred(c, k); got != want {
			t.Fatalf("full drop-in: %s = %q; want %q", k, got, want)
		}
	}

	// 3. The OTHER config keys are untouched (the drop-in MERGES, not overwrites).
	if got, _ := c.Get("cloudflare_worker_url"); got != "https://bot.example" {
		t.Fatalf("drop-in clobbered cloudflare_worker_url: %v", got)
	}

	// 4. Partial drop-in → only the present cred merges; the others stay as-is.
	write("d1creds.json", `{"cloudflare_api_token": "tok-only"}`)
	c = read()
	if got := cred(c, "cloudflare_api_token"); got != "tok-only" {
		t.Fatalf("partial drop-in: api_token = %q; want tok-only", got)
	}
	// account_id + database_id stay empty (the partial drop-in doesn't wipe them).
	if got := cred(c, "cloudflare_account_id"); got != "" {
		t.Fatalf("partial drop-in: account_id = %q; want empty (not in drop-in)", got)
	}

	// 5. Empty-string drop-in value → does NOT overwrite (only non-empty merges).
	write("d1creds.json", `{"cloudflare_api_token": ""}`)
	c = read()
	if got := cred(c, "cloudflare_api_token"); got != "" {
		t.Fatalf("empty-value drop-in: api_token = %q; want empty (empty values don't merge)", got)
	}

	// 6. Unparseable drop-in → no-op (a broken drop-in is ignored).
	write("d1creds.json", `{ not json`)
	c = read()
	if got := cred(c, "cloudflare_api_token"); got != "" {
		t.Fatalf("unparseable drop-in: api_token = %q; want empty", got)
	}
}
