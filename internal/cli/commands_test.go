package cli

// commands_test.go — the T-60 remainder's unit battery: the --init
// wizard's flow laws (the YOUR_ placeholder, the isdir retry loop, the
// api-id digit law, the provider branches with the probe seam, the
// threshold clamp, the radar's wires, the saved config), the
// --test-connection card over the connectioncheck seams, --detect-llm
// (the menu beats, set + save), the model-menu half, and the
// hand-delivery / chrome-retry wirings. The differential bytes are the
// golden replay's job (goldencli_test.go); these pin the behavioral
// laws.

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/reporadar"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// wizardCLI builds a shell with the wizard's seams stubbed
// deterministically.
func wizardCLI(out *bytes.Buffer, in string) *CLI {
	c := testCLI(out, in)
	c.IsDir = func(p string) bool {
		fi, err := os.Stat(p)
		return err == nil && fi.IsDir()
	}
	c.RadarAll = func(cfg any) map[string]any {
		return map[string]any{
			"github": reporadar.NewDetection(
				"kind", "github", "vault_path", "/vaults/main",
				"repo_name", "main-vault-backup",
				"confidence", "high", "method", "remote",
				"detail", "3 of 4 sampled files match"),
			"websites": reporadar.NewDetection(
				"kind", "websites", "vault_path", "/vaults/web",
				"repo_name", "", "detail", "no dominant match — 2 candidates too close"),
			"personal": reporadar.NewDetection(
				"kind", "personal", "vault_path", "", "repo_name", "",
				"detail", "no vault path configured"),
		}
	}
	c.ProbeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		model := "qwen2.5-coder:7b"
		return &llmclient.ProbeResult{
			Found: true, BaseURL: baseURL, Detail: "up @ " + baseURL +
				" · model '" + model + "'", Model: &model, Models: []string{model},
		}
	}
	return c
}

// TestWizardFlow drives the full wizard over a crafted starting config
// with scripted stdin: the YOUR_ placeholder counts as unset, the
// isdir retry loop re-asks, the radar's report + applied note print,
// the llamacpp probe seam offers the detected model, the threshold
// clamps, the save lands, the closing card carries NO login line.
func TestWizardFlow(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	vaults := t.TempDir()
	vaultDir := filepath.Join(vaults, "vault")
	if err := os.MkdirAll(vaultDir, 0o755); err != nil {
		t.Fatal(err)
	}
	startCfg := storage.NewOrderedMap()
	startCfg.Set("vault_path", "YOUR_OBSIDIAN_VAULT_HERE")
	startCfg.Set("telegram_api_id", int64(12345))
	startCfg.Set("telegram_api_hash", "abcdef1234567890abcdef12")
	startCfg.Set("telegram_phone", "+989123456789")
	startCfg.Set("bot_username", "curator_bot")
	startCfg.Set("github_token", "fake-github-token-0123456789")
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, startCfg); err != nil {
		t.Fatal(err)
	}

	// the script: vault (bad → retry → good), Enter×5 (api id,
	// hash, phone, bot, token keep their values), the other two
	// vaults, provider 3, Enter×3 (url+key+model), y + Enter×3
	// (proxy), 15 (the clamp)
	script := "/no/such/dir\n" + vaultDir + "\n\n\n\n\n\n" +
		filepath.Join(vaults, "webvault") + "\n" +
		filepath.Join(vaults, "manual") + "\n" +
		"3\n\n\n\ny\n\n\n\n15\n"
	buf := &bytes.Buffer{}
	c := wizardCLI(buf, script)
	rc := c.Run([]string{"--init", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("wizard rc = %d, want 0\noutput:\n%s", rc, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"First-run setup",
		"That folder does not exist — try again (or leave empty to skip).",
		"📡 Repo radar — looking for the vaults' backup repos",
		"🔗 GitHub vault: main-vault-backup — detected via its own git remote",
		"⚪ Websites vault: no backup repo found — no dominant match — 2 candidates too close",
		"⚪ Personal vault: not configured (no vault path)",
		"  → 1 backup repo name(s) set automatically — nothing to type, nothing created",
		"llama.cpp detected: up @ http://127.0.0.1:8080/v1 · model 'qwen2.5-coder:7b'",
		"Setup complete — no Telegram login is needed in the Go port",
		"gitcurator-cli --auto",
		"┌─ saved ─────",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("wizard output missing %q", want)
		}
	}
	for _, banned := range []string{"Log in before the first run", "--login"} {
		if strings.Contains(out, banned) {
			t.Errorf("wizard closing card still teaches the dropped login line %q", banned)
		}
	}

	// the saved config: the radar's slot + the wizard's answers.
	raw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := storage.DecodeJSON(raw)
	if err != nil {
		t.Fatal(err)
	}
	checks := []struct{ key, want string }{
		{"vault_path", vaultDir},
		{"telegram_api_id", "12345"},
		{"github_token", "fake-github-token-0123456789"},
		{"llm_provider", "llamacpp"},
		{"llamacpp_model", "qwen2.5-coder:7b"},
		{"notfound_strike_threshold", "10"},
	}
	for _, ck := range checks {
		v, _ := saved.Get(ck.key)
		if got := llmPyStr(v); got != ck.want {
			t.Errorf("saved %s = %q, want %q", ck.key, got, ck.want)
		}
	}
	seal := cfgGetOr(saved, "vaultseal")
	if got := llmPyStr(seal["repo_name"]); got != "main-vault-backup" {
		t.Errorf("the radar's slot = %q, want main-vault-backup", got)
	}
	px := cfgGetOr(saved, "proxy")
	if px["enabled"] != true {
		t.Errorf("proxy enabled = %v, want true", px["enabled"])
	}
	if got := llmPyStr(px["port"]); got != "10808" {
		t.Errorf("proxy port = %q, want 10808", got)
	}
}

func llmPyStr(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	default:
		b, _ := json.Marshal(v)
		return strings.Trim(string(b), `"`)
	}
}

func cfgGetOr(om *storage.OrderedMap, key string) map[string]any {
	out := map[string]any{}
	if om == nil {
		return out
	}
	if v, ok := om.Get(key); ok {
		if m, ok2 := v.(*storage.OrderedMap); ok2 {
			for _, k := range m.Keys() {
				vv, _ := m.Get(k)
				out[k] = vv
			}
		}
	}
	return out
}

// TestWizardAPIDigitLaw — int(s) if s.isdigit() else 0: a non-digit
// answer stores 0; the YOUR_ placeholder leaves the vault default
// empty.
func TestWizardAPIDigitLaw(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "config.json")
	script := "\nabc\n\n\n\n\n\n\n1\n\n\nn\n\n3\n"
	buf := &bytes.Buffer{}
	c := wizardCLI(buf, script)
	rc := c.Run([]string{"--init", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("wizard rc = %d, want 0", rc)
	}
	savedRaw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := storage.DecodeJSON(savedRaw)
	if err != nil {
		t.Fatal(err)
	}
	v, _ := saved.Get("telegram_api_id")
	if got := llmPyStr(v); got != "0" {
		t.Errorf("telegram_api_id = %q, want 0 (the non-digit law)", got)
	}
	if _, ok := saved.Get("vault_path"); ok {
		v, _ := saved.Get("vault_path")
		if llmPyStr(v) != "" {
			t.Errorf("vault_path = %q, want empty (the empty-start law)", llmPyStr(v))
		}
	}
}

// TestTestConnectionCard drives the battery over the connectioncheck
// seams (the ollama + github net doors): the four sections render, the
// Worker section is the Go battery's own (no Telegram leg), the
// verdict + the exit-code law hold.
func TestTestConnectionCard(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	vaults := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", filepath.Join(vaults, "vault"))
	cfg.Set("llm_provider", "ollama")
	oll := storage.NewOrderedMap()
	oll.Set("base_url", "http://127.0.0.1:11434")
	oll.Set("model", "mistral")
	cfg.Set("ollama", oll)
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(vaults, "vault"), 0o755); err != nil {
		t.Fatal(err)
	}

	prevGet := connectioncheck.HTTPGetJSONFn
	prevGH := connectioncheck.GitHubGetFn
	defer func() {
		connectioncheck.HTTPGetJSONFn = prevGet
		connectioncheck.GitHubGetFn = prevGH
	}()
	connectioncheck.HTTPGetJSONFn = func(u string, timeoutS float64) (any, error) {
		if strings.HasSuffix(u, "/api/tags") {
			om := storage.NewOrderedMap()
			models := []any{
				dictOf("name", "llama3"), dictOf("name", "qwen2.5-coder:7b"),
			}
			om.Set("models", models)
			return om, nil
		}
		return nil, &errFake{"URLError", "refused"}
	}
	connectioncheck.GitHubGetFn = func(path, token, apiBase string, timeoutS float64) connectioncheck.GitHubReply {
		if path == "/user" {
			om := storage.NewOrderedMap()
			om.Set("login", "octo-owner")
			return connectioncheck.GitHubReply{Status: 200, Payload: om}
		}
		return connectioncheck.GitHubReply{Status: 404}
	}

	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run([]string{"--test-connection", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("rc = %d, want 0 (warnings don't fail)\n%s", rc, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"Test Connection — is everything up and ready?",
		"vaults · LLM · GitHub · the sync queue (the Worker)",
		"[1/4] Vaults",
		"[2/4] LLM",
		"[3/4] GitHub",
		"[4/4] Worker",
		"GitHub vault",
		"Ollama",
		"configured 'mistral' is NOT pulled",
		"GitHub token",
		"VERDICT",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("battery card missing %q", want)
		}
	}
	if strings.Contains(out, "[4/4] Telegram") {
		t.Errorf("the Go battery rendered the reference's Telegram section title")
	}
	if strings.Contains(out, "ALL SYSTEMS READY") {
		t.Errorf("the warn-level env must not answer the all-clear")
	}
}

func dictOf(kv ...any) *storage.OrderedMap {
	om := storage.NewOrderedMap()
	for i := 0; i+1 < len(kv); i += 2 {
		om.Set(kv[i].(string), kv[i+1])
	}
	return om
}

type errFake struct{ class, msg string }

func (e *errFake) Error() string { return e.class + ": " + e.msg }

// TestTestConnectionNoConfig — the warning card + exit 1.
func TestTestConnectionNoConfig(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run([]string{"--test-connection", "--config", filepath.Join(home, "missing.json")})
	if rc != 1 {
		t.Fatalf("rc = %d, want 1", rc)
	}
	if !strings.Contains(buf.String(), "No config at") {
		t.Errorf("missing the no-config card")
	}
}

// TestDetectLLMMenu — found with several models: the numbered menu,
// the pick, set + save; the saved config carries the provider switch.
func TestDetectLLMMenu(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfg := storage.NewOrderedMap()
	oll := storage.NewOrderedMap()
	oll.Set("base_url", "http://127.0.0.1:11434")
	oll.Set("model", "llama3")
	cfg.Set("llm_provider", "ollama")
	cfg.Set("ollama", oll)
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	c := testCLI(buf, "2\n")
	c.DetectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3", "qwen2.5-coder:7b"},
			Detail: "up · 2 model(s)"}
	}
	rc := c.Run([]string{"--detect-llm", "ollama", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("rc = %d, want 0\n%s", rc, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"🧠 Detect & Set Ollama — probing the local engine…",
		"🧠 Detecting Ollama at http://127.0.0.1:11434…",
		"✅ Ollama detected at http://127.0.0.1:11434",
		"2 models detected — pick one:",
		"1. llama3",
		"2. qwen2.5-coder:7b",
		"← current",
		"✅ LLM provider SET to Ollama — http://127.0.0.1:11434 · model 'qwen2.5-coder:7b'",
		"saved to",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("detect card missing %q", want)
		}
	}
	savedRaw, err := os.ReadFile(cfgPath)
	if err != nil {
		t.Fatal(err)
	}
	saved, err := storage.DecodeJSON(savedRaw)
	if err != nil {
		t.Fatal(err)
	}
	savedOll := cfgGetOr(saved, "ollama")
	if got := llmPyStr(savedOll["model"]); got != "qwen2.5-coder:7b" {
		t.Errorf("saved ollama.model = %q, want qwen2.5-coder:7b", got)
	}
}

// TestDetectLLMSingleAndAbort — the only-one model sets directly; an
// invalid menu choice aborts with nothing changed.
func TestDetectLLMSingleAndAbort(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, dictOf("llm_provider", "llamacpp")); err != nil {
		t.Fatal(err)
	}
	// single model: no menu
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	c.DetectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3"}, Detail: "up · 1 model(s)"}
	}
	rc := c.Run([]string{"--detect-llm", "ollama", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("single-model rc = %d, want 0", rc)
	}
	if !strings.Contains(buf.String(), "(the only one installed)") {
		t.Errorf("missing the only-one line")
	}
	// the abort: an invalid choice changes nothing (a multi-model probe
	// so the menu actually asks)
	before, _ := os.ReadFile(cfgPath)
	buf2 := &bytes.Buffer{}
	c2 := testCLI(buf2, "x\n")
	c2.DetectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3", "qwen2.5-coder:7b"},
			Detail: "up · 2 model(s)"}
	}
	rc2 := c2.Run([]string{"--detect-llm", "ollama", "--config", cfgPath})
	if rc2 != 1 {
		t.Fatalf("abort rc = %d, want 1", rc2)
	}
	if !strings.Contains(buf2.String(), "'x' is not a valid choice — aborted, nothing changed.") {
		t.Errorf("missing the abort card")
	}
	after, _ := os.ReadFile(cfgPath)
	if string(before) != string(after) {
		t.Errorf("the aborted choice changed config.json")
	}
}

// TestDetectLLMNotFound — the honest not-running card + exit 1.
func TestDetectLLMNotFound(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, dictOf()); err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	c.DetectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: false, BaseURL: base,
			Detail: "URLError: connection refused"}
	}
	rc := c.Run([]string{"--detect-llm", "ollama", "--config", cfgPath})
	if rc != 1 {
		t.Fatalf("rc = %d, want 1", rc)
	}
	for _, want := range []string{
		"Ollama is not running at http://127.0.0.1:11434 (URLError: connection refused).",
		"Start the Ollama app (or `ollama serve`), then run this again.",
	} {
		if !strings.Contains(buf.String(), want) {
			t.Errorf("not-found card missing %q", want)
		}
	}
}

// TestPromptModelMenu — the recommendation is entry 1, the numbered
// pick, the substring pick, q declines, five bad tries give up, and a
// non-TTY stdin auto-picks the recommendation.
func TestPromptModelMenu(t *testing.T) {
	available := []string{"llama3", "qwen2.5-coder:7b", "nomic-embed-text"}
	// the smart pick: the same family as the configured llama3 → llama3
	// first, the embedder demoted by its tag.
	buf := &bytes.Buffer{}
	c := testCLI(buf, "2\n")
	c.IsATTYIn = func() bool { return true }
	got := c.PromptModelMenu("llama3", available, nil, true)
	if got != "qwen2.5-coder:7b" {
		t.Errorf("pick 2 = %q, want qwen2.5-coder:7b", got)
	}
	out := buf.String()
	for _, want := range []string{
		"Configured model 'llama3' is not installed.",
		"Models available in Ollama:",
		"← recommended",
		"(embedding — cannot analyze)",
		"q = abort the run",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("menu missing %q", want)
		}
	}
	// the substring hit
	buf2 := &bytes.Buffer{}
	c2 := testCLI(buf2, "coder\n")
	c2.IsATTYIn = func() bool { return true }
	if got := c2.PromptModelMenu("llama3", available, nil, true); got != "qwen2.5-coder:7b" {
		t.Errorf("substring pick = %q, want qwen2.5-coder:7b", got)
	}
	// q declines
	buf3 := &bytes.Buffer{}
	c3 := testCLI(buf3, "q\n")
	c3.IsATTYIn = func() bool { return true }
	if got := c3.PromptModelMenu("llama3", available, nil, true); got != "" {
		t.Errorf("q = %q, want the decline", got)
	}
	// five bad answers give up ("zz" hits nothing — not even a
	// substring; "x" would hit nomic-embed-te**x**-t, the reference's
	// own substring law)
	buf4 := &bytes.Buffer{}
	c4 := testCLI(buf4, "zz\nzz\nzz\nzz\nzz\n")
	c4.IsATTYIn = func() bool { return true }
	if got := c4.PromptModelMenu("llama3", available, nil, true); got != "" {
		t.Errorf("five bad tries = %q, want the give-up", got)
	}
	if got := strings.Count(buf4.String(), "✖ Not a number or unique name"); got != 5 {
		t.Errorf("retry lines = %d, want 5", got)
	}
	// non-TTY auto-picks the recommendation (the smart pick's law:
	// the same family wins — llama3; without a family match the
	// biggest parameter count wins — qwen2.5-coder:7b)
	buf5 := &bytes.Buffer{}
	c5 := testCLI(buf5, "")
	c5.IsATTYIn = func() bool { return false }
	if got := c5.PromptModelMenu("llama3", available, nil, true); got != "llama3" {
		t.Errorf("non-TTY same-family = %q, want llama3", got)
	}
	if got := c5.PromptModelMenu("mistral", available, nil, true); got != "qwen2.5-coder:7b" {
		t.Errorf("non-TTY biggest-size = %q, want qwen2.5-coder:7b", got)
	}
	if strings.Contains(buf5.String(), "Models available") {
		t.Errorf("non-TTY menu printed anyway")
	}
	// empty available → "" without asking
	if got := (&CLI{}).PromptModelMenu("x", nil, nil, true); got != "" {
		t.Errorf("empty available = %q, want \"\"", got)
	}
}

// TestApplyModelChoice — the provider-aware writes.
func TestApplyModelChoice(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "config.json")
	cfg := dictOf("llm_provider", "ollama", "ollama", dictOf("model", "old"))
	c := testCLI(&bytes.Buffer{}, "")
	c.ApplyModelChoice(cfg, "new-model", cfgPath)
	if got := llmPyStr(cfgGetOr(cfg, "ollama")["model"]); got != "new-model" {
		t.Errorf("ollama write = %q", got)
	}
	cfg2 := dictOf("llm_provider", "cloud", "cloud_model", "old")
	c.ApplyModelChoice(cfg2, "gpt-4o-mini", "")
	if got := llmPyStr(cfgGet(cfg2, "cloud_model")); got != "gpt-4o-mini" {
		t.Errorf("cloud write = %q", got)
	}
	cfg3 := dictOf("llm_provider", "llamacpp")
	c.ApplyModelChoice(cfg3, "qwen.gguf", "")
	if got := llmPyStr(cfgGet(cfg3, "llamacpp_model")); got != "qwen.gguf" {
		t.Errorf("llamacpp write = %q", got)
	}
}

// cfgGet reads a scalar config value as any.
func cfgGet(om *storage.OrderedMap, key string) any {
	if om == nil {
		return nil
	}
	v, _ := om.Get(key)
	return v
}

// craftWalledDB builds a cache.db whose retry queue carries one walled
// row and one plain-failed row.
func craftWalledDB(t *testing.T, dbPath string) {
	t.Helper()
	ws, err := websitepipeline.NewWebsiteStateDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer ws.Close()
	if err := ws.EnqueueRetry("https://example.com/walled", "HTTP 403 — bot defense"); err != nil {
		t.Fatal(err)
	}
	if err := ws.EnqueueRetry("https://example.com/plain", "HTTP 500"); err != nil {
		t.Fatal(err)
	}
}

// TestHandDeliveryWired — the listing shows the walled row only, the
// enqueue writes queue.json, the folder card prints.
func TestHandDeliveryWired(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	vault := filepath.Join(home, "webvault")
	cfg := dictOf("website_vault_path", vault)
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	craftWalledDB(t, filepath.Join(home, "cache.db"))

	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run([]string{"--hand-delivery", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("rc = %d, want 0\n%s", rc, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"Walled links (the fourth door — your real Chrome)",
		"https://example.com/walled",
		"attempt 1 · the wall: HTTP 403 — bot defense",
		"queued for hand-delivery",
		"Next: in Chrome, Ctrl+S → 'Webpage, HTML Only'",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("hand-delivery card missing %q", want)
		}
	}
	if strings.Contains(out, "https://example.com/plain") {
		t.Errorf("the plain-failed row must not be walled-listed")
	}
	if _, err := os.Stat(filepath.Join(vault, "_review", "hand-delivered", "queue.json")); err != nil {
		t.Errorf("queue.json not written: %v", err)
	}
}

// TestHandDeliveryNoVault — the warning + exit 1.
func TestHandDeliveryNoVault(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, dictOf()); err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run([]string{"--hand-delivery", "--config", cfgPath})
	if rc != 1 {
		t.Fatalf("rc = %d, want 1", rc)
	}
	if !strings.Contains(buf.String(), "website_vault_path is not set — run --cli --init first.") {
		t.Errorf("missing the no-vault card")
	}
}

// TestChromeRetryDryRunAndBoundary — the dry-run rehearsal answers the
// honest card (Chrome never launched, nothing written); the live leg
// answers the T-62 boundary card.
func TestChromeRetryDryRunAndBoundary(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	vault := filepath.Join(home, "webvault")
	cfg := dictOf("website_vault_path", vault)
	pipes := dictOf("websites", true)
	cfg.Set("pipelines", pipes)
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	craftWalledDB(t, filepath.Join(home, "cache.db"))

	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run([]string{"--chrome-retry", "--dry-run", "--config", cfgPath})
	if rc != 0 {
		t.Fatalf("dry-run rc = %d, want 0\n%s", rc, buf.String())
	}
	out := buf.String()
	for _, want := range []string{
		"Failed links waiting for the fifth door",
		"https://example.com/walled",
		"https://example.com/plain",
		"🤖 Dry-run — Chrome is never launched; the fifth door rehearsed 2 link(s), delivered 0.",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("chrome-retry dry-run card missing %q", want)
		}
	}
	// nothing was written by the rehearsal
	if _, err := os.Stat(filepath.Join(vault, "_review", "hand-delivered", "queue.json")); err == nil {
		t.Errorf("the dry-run rehearsal wrote the queue — dry-run writes nothing")
	}

	// the live leg: the honest T-62 boundary card
	buf2 := &bytes.Buffer{}
	c2 := testCLI(buf2, "")
	rc2 := c2.Run([]string{"--chrome-retry", "--config", cfgPath})
	if rc2 != 1 {
		t.Fatalf("live rc = %d, want 1", rc2)
	}
	if !strings.Contains(buf2.String(), "lands with T-62 (Phase 5)") {
		t.Errorf("missing the T-62 boundary card")
	}
}

// TestChromeRetryPipelineOff — the websites-pipeline gate.
func TestChromeRetryPipelineOff(t *testing.T) {
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath,
		dictOf("website_vault_path", filepath.Join(home, "v"))); err != nil {
		t.Fatal(err)
	}
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	rc := c.Run([]string{"--chrome-retry", "--config", cfgPath})
	if rc != 1 {
		t.Fatalf("rc = %d, want 1", rc)
	}
	if !strings.Contains(buf.String(), "The websites pipeline is off") {
		t.Errorf("missing the pipeline-off card")
	}
}
