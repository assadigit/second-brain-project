// settings_test.go — T-73's acceptance battery: the save-law
// (merge/in-place) tests + the six pages' workflow tests, all over
// FAKE deps (the hermetic law — every network/process seam faked at
// construction, nothing real touched).
package app

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// The harness: fake deps + a recording Write door.
// ---------------------------------------------------------------------------

// fakeSettingsDeps is the hermetic wiring: every seam recorded, none
// real.
type fakeSettingsDeps struct {
	rec *busRec

	written   []*storage.OrderedMap
	writeFail bool

	discovered []string

	userStatus string
	userPath   string
	openedPath string

	githubReplies map[string]connectioncheck.GitHubReply

	detectOllama func(base string, timeoutS float64) llmclient.DetectOllamaResult
	probeLlama   func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult
	detectLlama  func(apiKey string) *llmclient.ProbeResult

	spawnPID int
	spawnErr string

	soundDir string

	refreshCalls atomic.Int32
}

func (f *fakeSettingsDeps) deps() SettingsDeps {
	return SettingsDeps{
		Config: func() *storage.OrderedMap { return nil },
		Write: func(cfg *storage.OrderedMap) bool {
			if f.writeFail {
				return false
			}
			f.written = append(f.written, cfg)
			return true
		},
		DiscoverVaults:   func() []string { return f.discovered },
		UserConfigStatus: func() string { return f.userStatus },
		UserConfigPath:   func() string { return f.userPath },
		OpenFileLocation: func(p string) { f.openedPath = p },
		GitHubGet: func(path, token, apiBase string, timeoutS float64) connectioncheck.GitHubReply {
			if r, ok := f.githubReplies[token]; ok {
				return r
			}
			return connectioncheck.GitHubReply{Status: 0, Err: "connection refused"}
		},
		DetectOllama: func(base string, timeoutS float64) llmclient.DetectOllamaResult {
			if f.detectOllama != nil {
				return f.detectOllama(base, timeoutS)
			}
			return llmclient.DetectOllamaResult{Found: false, BaseURL: base, Models: []string{}, Detail: "URLError: refused"}
		},
		ProbeLlama: func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
			if f.probeLlama != nil {
				return f.probeLlama(baseURL, apiKey, timeoutS)
			}
			return &llmclient.ProbeResult{Found: false, BaseURL: baseURL, Models: []string{}}
		},
		DetectLlama: func(apiKey string) *llmclient.ProbeResult {
			if f.detectLlama != nil {
				return f.detectLlama(apiKey)
			}
			return nil
		},
		SpawnOllamaServe: func() (int, string) { return f.spawnPID, f.spawnErr },
		SoundDir:         func() string { return f.soundDir },
		SentinelRefresh:  func() { f.refreshCalls.Add(1) },
	}
}

// newSettingsTest builds the surface over a fresh live config + the
// recording bus.
func newSettingsTest(t *testing.T) (*Settings, *busRec, *fakeSettingsDeps) {
	t.Helper()
	bus := NewBus()
	rec := &busRec{}
	rec.subscribe(bus)
	f := &fakeSettingsDeps{githubReplies: map[string]connectioncheck.GitHubReply{}}
	cfg := storage.NewOrderedMap()
	s := NewSettings(bus, cfg, &OpLock{}, f.deps())
	return s, rec, f
}

// writeCfgFile seeds the live config from a JSON string (the storage
// decoder keeps the order + the Python number laws).
func writeCfgFile(t *testing.T, cfg *storage.OrderedMap, raw string) {
	t.Helper()
	parsed, err := storage.DecodeJSON([]byte(raw))
	if err != nil {
		t.Fatalf("bad test config JSON: %v", err)
	}
	for _, k := range parsed.Keys() {
		v, _ := parsed.Get(k)
		cfg.Set(k, v)
	}
}

func wantLog(t *testing.T, rec *busRec, part, level string) LogLine {
	t.Helper()
	l, ok := rec.findLog(part)
	if !ok {
		t.Fatalf("no log line contains %q (have %d lines)", part, len(rec.logs))
	}
	if l.Level != level {
		t.Fatalf("line %q has level %q, want %q", part, l.Level, level)
	}
	return l
}

func noLog(t *testing.T, rec *busRec, part string) {
	t.Helper()
	if l, ok := rec.findLog(part); ok {
		t.Fatalf("unexpected log line: %q (level %q)", l.Msg, l.Level)
	}
}

// ---------------------------------------------------------------------------
// THE SAVE LAW — merge (never whitelist-rebuild).
// ---------------------------------------------------------------------------

func TestSaveLawMergeNeverDropsUnknownKeys(t *testing.T) {
	s, _, f := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"hmac_secret": "keep-me", "cloudflare_install_id": "abc",
                "my_own_key": [1, 2, 3], "timeout_per_repo": 120, "llm_provider": "ollama"}`)
	cfg := s.srvCfg(t)

	res := s.SaveSettings(&SettingsValues{
		Credentials: &CredentialsValues{GitHubToken: "ghp_new"},
	})
	if !res.Saved {
		t.Fatal("save failed")
	}
	for _, key := range []string{"hmac_secret", "cloudflare_install_id", "my_own_key"} {
		if _, ok := cfg.Get(key); !ok {
			t.Fatalf("merge dropped the unknown key %q", key)
		}
	}
	if v, _ := cfg.Get("my_own_key"); len(v.([]any)) != 3 {
		t.Fatalf("merge mangled my_own_key: %v", v)
	}
	// The tuning key the reference stomped pre-v30: a USER value must
	// survive (only defaulted when absent).
	if v, _ := cfg.Get("timeout_per_repo"); fmt.Sprint(v) != "120" {
		t.Fatalf("timeout_per_repo = %v, want the user's 120 (never stomped)", v)
	}
	// The saved keys land.
	if v, _ := cfg.Get("github_token"); v != "ghp_new" {
		t.Fatalf("github_token = %v", v)
	}
	// The written file carries everything too.
	if len(f.written) != 1 {
		t.Fatalf("write called %d times", len(f.written))
	}
	for _, key := range []string{"hmac_secret", "my_own_key", "github_token"} {
		if _, ok := f.written[0].Get(key); !ok {
			t.Fatalf("the written config dropped %q", key)
		}
	}
}

// srvCfg is the test-local accessor for the live config (the struct
// field is package-private; the tests are in-package).
func (s *Settings) srvCfg(t *testing.T) *storage.OrderedMap {
	t.Helper()
	if s.cfg == nil {
		t.Fatal("the settings surface has no live config")
	}
	return s.cfg
}

func TestSaveLawInPlaceMutationSeenByRunningWorker(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"github_token": "old", "ollama": {"base_url": "http://x", "model": "m1"}}`)
	cfg := s.srvCfg(t)

	// THE RUNNING WORKER FAKE: it holds the same map the app holds
	// (exactly how ProcessingWorker held self.config).
	workerView := cfg

	res := s.SaveSettings(&SettingsValues{
		LLM: &LLMValues{Provider: "ollama", OllamaURL: "http://y", OllamaModel: "m2"},
	})
	if !res.Saved {
		t.Fatal("save failed")
	}

	// The worker SEES the save through the same instance — the v30 law
	// (never a rebind; clear+update in place).
	if v, _ := workerView.Get("github_token"); v != "old" {
		t.Fatalf("the worker's github_token changed: %v", v)
	}
	if oll, ok := workerView.Get("ollama"); !ok {
		t.Fatal("ollama key vanished from the worker's view")
	} else {
		om := oll.(*storage.OrderedMap)
		if v, _ := om.Get("model"); v != "m2" {
			t.Fatalf("the worker does not see the saved model: %v", v)
		}
		if v, _ := om.Get("base_url"); v != "http://y" {
			t.Fatalf("the worker does not see the saved base_url: %v", v)
		}
	}
	// And the top-level identity is unchanged (the same map, not a swap).
	if s.srvCfg(t) != workerView {
		t.Fatal("the live config was REBOUND — the in-place law is broken")
	}
}

func TestSaveLawAbsentGroupsKeepConfigValues(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"bot_token": "keep", "github_token": "gt",
                "backup_folder": "/my/backups", "notfound_strike_threshold": 7,
                "llm_num_ctx": 4096, "llm_max_output_tokens": 512}`)
	cfg := s.srvCfg(t)

	// A Sound-page save only: every OTHER page's keys keep their values
	// (the defensive hasattr pattern — the T-74 pages do not exist yet,
	// their keys must never be wiped).
	res := s.SaveSettings(&SettingsValues{Sound: &SoundValues{Enabled: true, Volume: 40}})
	if !res.Saved {
		t.Fatal("save failed")
	}
	for _, kv := range [][2]any{{"bot_token", "keep"}, {"github_token", "gt"},
		{"backup_folder", "/my/backups"}, {"notfound_strike_threshold", 7},
		{"llm_num_ctx", 4096}, {"llm_max_output_tokens", 512}} {
		v, _ := cfg.Get(kv[0].(string))
		if fmt.Sprint(v) != fmt.Sprint(kv[1]) {
			t.Fatalf("%s = %v, want %v (an absent group must keep the config value)",
				kv[0], v, kv[1])
		}
	}
	if v, _ := cfg.Get("sound_volume"); v != 0.4 {
		t.Fatalf("sound_volume = %v, want 0.4", v)
	}
}

func TestSaveLawLogLevelQuirkAndLenientReaders(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"log_level": "DEBUG", "llm_num_ctx": 2048,
                "llm_max_output_tokens": 128}`)
	cfg := s.srvCfg(t)

	// The one hardcoded quirk: EVERY save writes log_level "INFO" — a
	// hand-edited DEBUG resets (ui_values verbatim).
	res := s.SaveSettings(&SettingsValues{
		LLM: &LLMValues{NumCtx: "garbage!", MaxOutputTokens: ""}})
	if !res.Saved {
		t.Fatal("save failed")
	}
	if v, _ := cfg.Get("log_level"); v != "INFO" {
		t.Fatalf("log_level = %v, want the hardcoded INFO quirk", v)
	}
	// The lenient readers: garbage/empty fields keep the previous
	// config values — a bad field can never break a save.
	if v, _ := cfg.Get("llm_num_ctx"); v != 2048 {
		t.Fatalf("llm_num_ctx = %v, want the kept 2048", v)
	}
	if v, _ := cfg.Get("llm_max_output_tokens"); v != 128 {
		t.Fatalf("llm_max_output_tokens = %v, want the kept 128", v)
	}
	// Digits ≥ 0 are taken as-is.
	s.SaveSettings(&SettingsValues{LLM: &LLMValues{NumCtx: "16384", MaxOutputTokens: "99"}})
	if v, _ := cfg.Get("llm_num_ctx"); v != 16384 {
		t.Fatalf("llm_num_ctx = %v, want 16384", v)
	}
	if v, _ := cfg.Get("llm_max_output_tokens"); v != 99 {
		t.Fatalf("llm_max_output_tokens = %v, want 99", v)
	}
}

func TestSaveLawFailedWriteLeavesLiveMapUntouched(t *testing.T) {
	s, _, f := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"github_token": "original"}`)
	f.writeFail = true
	res := s.SaveSettings(&SettingsValues{Credentials: &CredentialsValues{GitHubToken: "changed"}})
	if res.Saved {
		t.Fatal("a failed write must report unsaved")
	}
	if v, _ := s.srvCfg(t).Get("github_token"); v != "original" {
		t.Fatalf("a failed write mutated the live map: %v", v)
	}
}

func TestSaveLawSentinelReArm(t *testing.T) {
	s, _, f := newSettingsTest(t)
	s.SaveSettings(&SettingsValues{Sound: &SoundValues{Enabled: true}})
	// The 250 ms singleShot: the refresh fires after the beat.
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if f.refreshCalls.Load() >= 1 {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	t.Fatal("the sentinel re-arm never fired after the save")
}

func TestSaveLawProxyPortDigitsAndDomains(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	res := s.SaveSettings(&SettingsValues{
		Proxy: &ProxyValues{Enabled: true, Type: "http", Host: " proxy.example.com ",
			Port: "garbage", UseForWeb: true},
		Vault: &VaultValues{WebBlocked: " x.com , EXTRA.com ,", WebSelf: ""},
	})
	if !res.Saved {
		t.Fatal("save failed")
	}
	cfg := s.srvCfg(t)
	if px, ok := cfg.Get("proxy"); !ok {
		t.Fatal("no proxy key")
	} else {
		om := px.(*storage.OrderedMap)
		if v, _ := om.Get("port"); v != 10808 {
			t.Fatalf("garbage port = %v, want the 10808 default", v)
		}
		if v, _ := om.Get("host"); v != "proxy.example.com" {
			t.Fatalf("host = %v (must be stripped)", v)
		}
	}
	if v, _ := cfg.Get("web_blocked_domains"); fmt.Sprint(v) != "[x.com extra.com]" {
		t.Fatalf("web_blocked_domains = %v, want [x.com extra.com] (trimmed + lowered)", v)
	}
	if v, _ := cfg.Get("web_self_domains"); len(v.([]any)) != 0 {
		t.Fatalf("an empty self field must save as the deliberate empty list, got %v", v)
	}
}

// ---------------------------------------------------------------------------
// THE VAULT LADDER.
// ---------------------------------------------------------------------------

func TestVaultPathStatusLadder(t *testing.T) {
	dir := t.TempDir()
	existing := filepath.Join(dir, "vault")
	if err := os.MkdirAll(existing, 0o755); err != nil {
		t.Fatal(err)
	}
	cases := []struct {
		path, kind, text, state string
	}{
		{"", "github", "● not set", "warning"},
		{"  ", "websites", "● not set", "warning"},
		{existing, "github", "● found", "success"},
		{existing, "manual", "● found", "success"},
		{filepath.Join(dir, "nope"), "websites", "● will be created (folder does not exist yet)", "warning"},
		{filepath.Join(dir, "nope"), "github", "● missing on disk", "error"},
		{filepath.Join(dir, "nope"), "manual", "● not found — you create this vault yourself", "warning"},
	}
	for _, c := range cases {
		got := VaultPathStatus(c.path, c.kind)
		if got.Text != c.text || got.State != c.state {
			t.Fatalf("ladder(%q, %q) = %q/%q, want %q/%q",
				c.path, c.kind, got.Text, got.State, c.text, c.state)
		}
	}
}

func TestRemoveVaultHistoryLaw(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"vault_path": "/v/a", "vaults_history": ["/v/a", "/v/b", "/v/c"]}`)
	cfg := s.srvCfg(t)

	res := s.RemoveVault("/v/a")
	if !res.Saved {
		t.Fatal("remove failed")
	}
	if v, _ := cfg.Get("vault_path"); v != "" {
		t.Fatalf("removing the CURRENT vault must clear vault_path, got %q", v)
	}
	if v, _ := cfg.Get("vaults_history"); fmt.Sprint(v) != "[/v/b /v/c]" {
		t.Fatalf("history = %v, want [/v/b /v/c]", v)
	}

	// Removing a NON-current vault keeps vault_path.
	writeCfgFile(t, cfg, `{"vault_path": "/v/b", "vaults_history": ["/v/b", "/v/c"]}`)
	s.RemoveVault("/v/c")
	if v, _ := cfg.Get("vault_path"); v != "/v/b" {
		t.Fatalf("vault_path must survive a non-current removal, got %q", v)
	}
	if v, _ := cfg.Get("vaults_history"); fmt.Sprint(v) != "[/v/b]" {
		t.Fatalf("history = %v, want [/v/b]", v)
	}
}

// ---------------------------------------------------------------------------
// The SNAPSHOT (populate_vaults' dedupe + the pages' defaults).
// ---------------------------------------------------------------------------

func TestSnapshotBuildsAllSixPages(t *testing.T) {
	s, _, f := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"github_token": "gt", "telegram_api_id": 123,
                "proxy": {"enabled": true, "type": "http", "host": "h", "port": 1234, "use_for_web": false},
                "vault_path": "/v/current", "vaults_history": ["/v/current", "/v/old"],
                "llm_provider": "cloud", "llm_num_ctx": 555,
                "ollama": {"base_url": "http://oll", "model": "m"},
                "sound_volume": 0.5, "sound_enabled": false,
                "pipelines": {"github": false, "websites": true}}`)
	f.discovered = []string{"/v/current", "/v/found1", "/v/old", "/v/found2"}
	f.userStatus = "the status line"
	f.userPath = "/home/u/.gitcurator/config.json"

	snap := s.Snapshot()

	// Credentials.
	if snap.GitHubToken != "gt" || snap.TelegramAPIID != "123" {
		t.Fatalf("credentials fields: %+v", snap)
	}
	if snap.UserConfigStatus != "the status line" || !snap.UserConfigShowable {
		t.Fatalf("user-config truth: %q / %v", snap.UserConfigStatus, snap.UserConfigShowable)
	}
	// Proxy.
	if !snap.ProxyEnabled || snap.ProxyType != "http" || snap.ProxyHost != "h" ||
		snap.ProxyPort != "1234" || snap.ProxyUseForWeb {
		t.Fatalf("proxy fields: %+v", snap)
	}
	// Vault combo: current FIRST, history deduped, discovery appended.
	want := []string{"/v/current", "/v/old", "/v/found1", "/v/found2"}
	if strings.Join(snap.VaultChoices, "|") != strings.Join(want, "|") {
		t.Fatalf("vault choices = %v, want %v", snap.VaultChoices, want)
	}
	// LLM.
	if snap.LLMProvider != "cloud" || snap.LLMNumCtx != "555" ||
		snap.OllamaURL != "http://oll" || snap.OllamaModel != "m" {
		t.Fatalf("llm fields: %+v", snap)
	}
	// Sound: the slider's int(round(0.5*100)).
	if snap.SoundVolume != 50 || snap.SoundEnabled {
		t.Fatalf("sound fields: %+v", snap)
	}
	// Pipelines.
	if snap.PipelineGithub || !snap.PipelineWebsites {
		t.Fatalf("pipelines: %+v", snap)
	}
	// The blocked field shows the LAW+extras union (ui.py's read).
	if !strings.Contains(snap.WebBlocked, "x.com") || !strings.Contains(snap.WebBlocked, "t.co") {
		t.Fatalf("the blocked field must show the law list, got %q", snap.WebBlocked)
	}
	// The law list rides along for the '?' help.
	if len(snap.BlockedLawDomains) == 0 || snap.BlockedLawDomains[0] != "x.com" {
		t.Fatalf("the law list: %v", snap.BlockedLawDomains)
	}
}

func TestSnapshotDefaultsAndSoundHeal(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	// A garbage/NaN volume heals to the 0.8 default (the widget-init
	// law) — garbage here is a non-number, which the loader keeps as a
	// string; the read heals to 0.8.
	writeCfgFile(t, s.srvCfg(t), `{"sound_volume": "garbage", "sound_enabled": true}`)
	snap := s.Snapshot()
	if snap.SoundVolume != 80 {
		t.Fatalf("a garbage volume must heal to 80, got %d", snap.SoundVolume)
	}
	if snap.LLMProvider != "ollama" || snap.OllamaURL != "http://localhost:11434" ||
		snap.OllamaModel != "qwythos-9b" || snap.LLMNumCtx != "8192" ||
		snap.LlamaCppURL != llmclient.LLamaCppDefaultBase+"/v1" ||
		snap.CloudURL != "https://api.openai.com/v1" || snap.CloudModel != "gpt-4o-mini" {
		t.Fatalf("the empty config's page defaults: %+v", snap)
	}
	// No pre-selection: an empty vault_path leaves the choices empty.
	if snap.VaultPath != "" || len(snap.VaultChoices) != 0 {
		t.Fatalf("no pre-selection law: %q %v", snap.VaultPath, snap.VaultChoices)
	}
}

// ---------------------------------------------------------------------------
// Credentials page: the token test + the About Me wizard + the Show door.
// ---------------------------------------------------------------------------

func TestTestGitHubTokenFlows(t *testing.T) {
	s, rec, f := newSettingsTest(t)

	// Empty → the anonymous warning.
	r := s.TestGitHubToken("  ")
	if r.OK || r.Level != "warning" {
		t.Fatalf("empty token: %+v", r)
	}
	wantLog(t, rec, "No GitHub token entered", "warning")

	// 200 → the account login.
	loginOM := storage.NewOrderedMap()
	loginOM.Set("login", "the-owner")
	f.githubReplies["ghp_good"] = connectioncheck.GitHubReply{Status: 200, Payload: loginOM}
	r = s.TestGitHubToken("ghp_good")
	if !r.OK || r.Login != "the-owner" {
		t.Fatalf("good token: %+v", r)
	}
	wantLog(t, rec, "✅ GitHub token valid — account: the-owner (5000 requests/hour enabled)", "success")

	// 401 → the actionable line.
	f.githubReplies["ghp_bad"] = connectioncheck.GitHubReply{Status: 401}
	r = s.TestGitHubToken("ghp_bad")
	if r.OK {
		t.Fatalf("401 must not be ok: %+v", r)
	}
	wantLog(t, rec, "REJECTED (401 Bad credentials)", "error")

	// 403 → the forbidden line with the payload's message.
	forbidden := storage.NewOrderedMap()
	forbidden.Set("message", "rate limited")
	f.githubReplies["ghp_limited"] = connectioncheck.GitHubReply{Status: 403, Payload: forbidden}
	r = s.TestGitHubToken("ghp_limited")
	if r.OK {
		t.Fatalf("403 must not be ok: %+v", r)
	}
	wantLog(t, rec, "forbidden (403): rate limited", "error")

	// Transport failure → the network line.
	f.githubReplies["ghp_dead"] = connectioncheck.GitHubReply{Status: 0, Err: "dial tcp: refused"}
	r = s.TestGitHubToken("ghp_dead")
	if r.OK {
		t.Fatalf("network failure must not be ok: %+v", r)
	}
	wantLog(t, rec, "GitHub token test failed (network)", "error")
}

func TestBuildAboutMePureTemplate(t *testing.T) {
	// The full form.
	got := BuildAboutMe(AboutMeFields{Name: "Software Engineer", Role: "AI/ML",
		Interests: "agents", Objectives: "learn", Tech: "Go", VaultUse: "recall"})
	for _, part := range []string{"# About Me", "## Who I am\nI am a Software Engineer.",
		"## My role\nAI/ML", "## My interests\nagents", "## My objectives\nlearn",
		"## Technologies I use\nGo", "## How I use this knowledge base\nrecall"} {
		if !strings.Contains(got, part) {
			t.Fatalf("missing %q in:\n%s", part, got)
		}
	}
	// The all-empty default (the reference's two default sections).
	got = BuildAboutMe(AboutMeFields{})
	if !strings.Contains(got, "## Who I am\nI am a software engineer who curates GitHub projects.") ||
		!strings.Contains(got, "## My objectives\nDiscover useful tools and frameworks for my development workflow.") {
		t.Fatalf("the empty wizard must write the default sections, got:\n%s", got)
	}
	// Partial: only the filled answers land.
	got = BuildAboutMe(AboutMeFields{Tech: "  Rust  "})
	if strings.Contains(got, "Who I am") || !strings.Contains(got, "## Technologies I use\nRust") {
		t.Fatalf("partial answers must skip empties + trim, got:\n%s", got)
	}
}

func TestGenerateAboutMeWritesTheFile(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	dir := t.TempDir()
	oldWd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(oldWd) }()

	r := s.GenerateAboutMe(AboutMeFields{Name: "Engineer"})
	if !r.OK || !r.Written {
		t.Fatalf("wizard result: %+v", r)
	}
	data, err := os.ReadFile(filepath.Join(dir, "about_me.md"))
	if err != nil {
		t.Fatalf("about_me.md was not written to the CWD: %v", err)
	}
	if r.Chars != len(data) || !strings.HasPrefix(string(data), "# About Me") {
		t.Fatalf("the file's bytes disagree with the result: %d vs %d", r.Chars, len(data))
	}
	wantLog(t, rec, fmt.Sprintf("📝 Generated about_me.md (%d chars)", r.Chars), "success")
	wantLog(t, rec, "The LLM will now personalize analyses based on your profile.", "info")
}

func TestShowUserConfigFileDoor(t *testing.T) {
	s, _, f := newSettingsTest(t)
	f.userPath = filepath.Join(t.TempDir(), "home", ".gitcurator", "config.json")
	if !s.ShowUserConfigFile() {
		t.Fatal("the door must open")
	}
	// The folder is created first (the reference's makedirs — the file
	// must be findable, not theoretical).
	if fi, err := os.Stat(filepath.Dir(f.userPath)); err != nil || !fi.IsDir() {
		t.Fatalf("the user config folder was not created: %v", err)
	}
	if f.openedPath == "" {
		t.Fatal("the opener never fired")
	}
	// An empty path = the disabled door (the reference's setEnabled law).
	f.userPath = ""
	if s.ShowUserConfigFile() {
		t.Fatal("an empty path must not open")
	}
}

// ---------------------------------------------------------------------------
// LLM page: Ollama Refresh + Start Server.
// ---------------------------------------------------------------------------

func TestRefreshOllamaModelsFlows(t *testing.T) {
	s, rec, f := newSettingsTest(t)

	// Empty URL.
	if s.RefreshOllamaModels("  ", "").OK {
		t.Fatal("an empty URL must fail")
	}
	wantLog(t, rec, "❌ Ollama URL is empty.", "error")

	// Server down.
	r := s.RefreshOllamaModels("http://down", "")
	if r.OK {
		t.Fatal("a down server must fail")
	}
	wantLog(t, rec, "Could not list models. Is Ollama running?", "error")

	// Up, no models.
	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base, Models: []string{}}
	}
	r = s.RefreshOllamaModels("http://up", "")
	if !r.OK || len(r.Models) != 0 {
		t.Fatalf("no-models result: %+v", r)
	}
	wantLog(t, rec, "Ollama is running but no models are installed", "warning")

	// Models + a custom typed name survives at the top (the combo's
	// law) + the in-place ollama.model sync (the v30 fix).
	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"m1", "m2"}}
	}
	r = s.RefreshOllamaModels("http://up", "my-custom")
	if !r.OK || r.Model != "my-custom" || len(r.Models) != 3 {
		t.Fatalf("custom-name result: %+v", r)
	}
	if r.Models[0] != "my-custom" {
		t.Fatalf("the custom name must ride at the top: %v", r.Models)
	}
	oll, _ := s.srvCfg(t).Get("ollama")
	if v, _ := oll.(*storage.OrderedMap).Get("model"); v != "my-custom" {
		t.Fatalf("the in-place ollama.model sync failed: %v", v)
	}
	wantLog(t, rec, "✅ Found 2 model(s): m1, m2", "success")

	// No current: the first listed model selects.
	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"m1", "m2"}}
	}
	r = s.RefreshOllamaModels("http://up", "")
	if r.Model != "m1" {
		t.Fatalf("the first model must select: %+v", r)
	}
}

func TestStartOllamaServerFlows(t *testing.T) {
	s, rec, f := newSettingsTest(t)

	// Not on PATH.
	f.spawnErr = "FileNotFoundError"
	if s.StartOllamaServer().OK {
		t.Fatal("a missing binary must fail")
	}
	wantLog(t, rec, "'ollama' not found on PATH", "error")

	// Spawned.
	f.spawnErr = ""
	f.spawnPID = 4242
	r := s.StartOllamaServer()
	if !r.OK || r.PID != 4242 {
		t.Fatalf("spawn result: %+v", r)
	}
	wantLog(t, rec, "✅ Ollama server starting in background (PID 4242)", "success")
}

// ---------------------------------------------------------------------------
// LLM page: llama.cpp Detect / Refresh + THE SERVER CONFIG GRAB.
// ---------------------------------------------------------------------------

func llamaProbe(base string, found bool, models []string, limits *llmclient.LlamaCppLimits) *llmclient.ProbeResult {
	detail := "llama.cpp server at " + base
	return &llmclient.ProbeResult{Found: found, BaseURL: base, Models: models, Detail: detail, PropsLimits: limits}
}

func TestDetectLlamaCppFlows(t *testing.T) {
	s, rec, f := newSettingsTest(t)

	// Nothing anywhere.
	f.detectLlama = func(string) *llmclient.ProbeResult { return nil }
	r := s.DetectLlamaCpp("")
	if r.OK && r.Found {
		t.Fatal("nothing found must not report found")
	}
	wantLog(t, rec, "No llama.cpp server found", "error")

	// Found via the scan + the props grab rides the probe.
	nCtx, nPred := 8192, 512
	f.detectLlama = func(string) *llmclient.ProbeResult {
		return llamaProbe("http://127.0.0.1:8080", true,
			[]string{"llama-model"}, &llmclient.LlamaCppLimits{NCtx: &nCtx, NPredict: &nPred})
	}
	r = s.DetectLlamaCpp("")
	if !r.Found || r.URL != "http://127.0.0.1:8080/v1" || r.Model != "llama-model" {
		t.Fatalf("detect result: %+v", r)
	}
	// The grab filled the limits IN PLACE + saved.
	cfg := s.srvCfg(t)
	if v, _ := cfg.Get("llm_num_ctx"); v != 8192 {
		t.Fatalf("the grab did not fill llm_num_ctx: %v", v)
	}
	if v, _ := cfg.Get("llm_max_output_tokens"); v != 512 {
		t.Fatalf("the grab did not fill llm_max_output_tokens: %v", v)
	}
	if v, _ := cfg.Get("llamacpp_api_url"); v != "http://127.0.0.1:8080/v1" {
		t.Fatalf("the URL did not sync in place: %v", v)
	}
	wantLog(t, rec, "🦙 llama.cpp server config grabbed: context window 8192 tokens · output cap 512 tokens", "success")

	// No model loaded anywhere → the warning (a fresh surface: the
	// current-model law would otherwise keep a previously synced pick).
	s2, rec2, f2 := newSettingsTest(t)
	f2.detectLlama = func(string) *llmclient.ProbeResult {
		return llamaProbe("http://127.0.0.1:8080", true, nil, nil)
	}
	r = s2.DetectLlamaCpp("")
	if r.Model != "" {
		t.Fatalf("a model-less probe must not pick: %+v", r)
	}
	wantLog(t, rec2, "Server detected but no model is loaded", "warning")
	// The current-model law: with a PREVIOUSLY synced pick, the same
	// probe keeps it (the reference's _fill_llamacpp_models law).
	writeCfgFile(t, s.srvCfg(t), `{"llamacpp_model": "kept"}`)
	r = s.DetectLlamaCpp("")
	if r.Model != "kept" {
		t.Fatalf("the current model must survive a model-less probe: %+v", r)
	}
}

func TestRefreshLlamaCppModelsFlows(t *testing.T) {
	s, rec, f := newSettingsTest(t)

	// Empty URL.
	if s.RefreshLlamaCppModels("", "").OK {
		t.Fatal("an empty URL must fail")
	}
	wantLog(t, rec, "❌ llama.cpp Server URL is empty.", "error")

	// The URL answers nothing.
	r := s.RefreshLlamaCppModels("http://127.0.0.1:9999", "")
	if r.OK && r.Found {
		t.Fatal("a dead URL must not report found")
	}
	wantLog(t, rec, "No llama.cpp server at http://127.0.0.1:9999", "error")

	// Found + a model list fills + the current model law (the current
	// wins over the first listed).
	model := "kept-model"
	f.probeLlama = func(base, key string, t float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: "http://127.0.0.1:8080",
			Models: []string{"first", "second"}, Model: &model,
			Detail: "llama.cpp identified by props"}
	}
	r = s.RefreshLlamaCppModels("http://127.0.0.1:8080/v1", "")
	if !r.Found || r.Model != "kept-model" {
		t.Fatalf("refresh result: %+v", r)
	}
	wantLog(t, rec, "✅ llama.cpp identified by props", "success")
}

// ---------------------------------------------------------------------------
// The quick-switch fast lane (Detect & Set).
// ---------------------------------------------------------------------------

func TestQuickDetectBatchRunningGuard(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	s.lock.Acquire("a-batch")
	defer s.lock.Release("a-batch")
	r := s.QuickDetect("ollama", "", "")
	if r.OK {
		t.Fatal("a running batch must refuse the switch")
	}
	wantLog(t, rec, "⏳ A batch is running — wait for it to finish before switching the LLM engine.", "warning")
}

func TestQuickDetectOllamaSingleModelAutoApplies(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"llm_provider": "llamacpp", "ollama": {"base_url": "http://old", "model": "old"}}`)
	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: "http://fresh:11434",
			Models: []string{"the-one"}}
	}
	r := s.QuickDetect("ollama", "", "")
	if !r.OK || !r.Found || !r.Applied || r.Choice != "the-one" {
		t.Fatalf("quick result: %+v", r)
	}
	cfg := s.srvCfg(t)
	if v, _ := cfg.Get("llm_provider"); v != "ollama" {
		t.Fatalf("llm_provider = %v", v)
	}
	oll, _ := cfg.Get("ollama")
	om := oll.(*storage.OrderedMap)
	if v, _ := om.Get("base_url"); v != "http://fresh:11434" {
		t.Fatalf("base_url = %v", v)
	}
	if v, _ := om.Get("model"); v != "the-one" {
		t.Fatalf("model = %v", v)
	}
	wantLog(t, rec, "✅ 🧠 Ollama detected at http://fresh:11434 · model 'the-one'", "success")
	wantLog(t, rec, "✅ 🧠 LLM provider SET to Ollama — http://fresh:11434 · model 'the-one' — saved", "success")
}

func TestQuickDetectSeveralModelsReturnsTheMenu(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"m1", "m2", "m3"}}
	}
	r := s.QuickDetect("ollama", "", "")
	if !r.Found || r.Applied || len(r.Models) != 3 {
		t.Fatalf("the several-models case must return the menu list: %+v", r)
	}
	wantLog(t, rec, "3 models — pick one:", "success")

	// The menu's Cancel → the honest line, nothing changed.
	writeCfgFile(t, s.srvCfg(t), `{"llm_provider": "ollama"}`)
	res := s.ApplyQuickDetect("ollama", "http://x", []string{"m1", "m2"}, "", nil)
	if res.Applied {
		t.Fatal("a cancelled menu must not apply")
	}
	wantLog(t, rec, "⏭️ Cancelled — the LLM provider was NOT changed (still ollama).", "info")

	// The menu's pick applies + saves.
	res = s.ApplyQuickDetect("ollama", "http://x:11434", []string{"m1", "m2"}, "m2", nil)
	if !res.Applied {
		t.Fatal("the chosen model must apply")
	}
	oll, _ := s.srvCfg(t).Get("ollama")
	if v, _ := oll.(*storage.OrderedMap).Get("model"); v != "m2" {
		t.Fatalf("model = %v, want m2", v)
	}
	wantLog(t, rec, "✅ 🧠 LLM provider SET to Ollama — http://x:11434 · model 'm2' — saved", "success")
}

func TestQuickDetectLlamaCppViaScanAndLiveOverlay(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"llm_provider": "ollama", "llamacpp_api_url": "http://stale:8080/v1",
                "llamacpp_model": "", "cloud_api_key": "cloud-keep"}`)

	// The configured URL is dead; the SCAN finds the server.
	found := llamaProbe("http://127.0.0.1:8080", true, []string{"llama-x"}, nil)
	found.Via = "process"
	f.probeLlama = func(base, key string, t float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: false, BaseURL: base, Models: []string{}}
	}
	f.detectLlama = func(string) *llmclient.ProbeResult { return found }

	r := s.QuickDetect("llamacpp", "", "")
	if !r.Found || !r.Applied || r.Choice != "llama-x" {
		t.Fatalf("llamacpp quick result: %+v", r)
	}
	cfg := s.srvCfg(t)
	if v, _ := cfg.Get("llm_provider"); v != "llamacpp" {
		t.Fatalf("llm_provider = %v", v)
	}
	if v, _ := cfg.Get("llamacpp_api_url"); v != "http://127.0.0.1:8080/v1" {
		t.Fatalf("llamacpp_api_url = %v", v)
	}
	// The LIVE overlay: the other engines' current truth is preserved
	// through the save (cloud_api_key kept — never clobbered).
	if v, _ := cfg.Get("cloud_api_key"); v != "cloud-keep" {
		t.Fatalf("the untouched engines must keep their values: %v", v)
	}
	wantLog(t, rec, "✅ 🦙 llama.cpp detected at http://127.0.0.1:8080 · model 'llama-x'", "success")
}

func TestQuickDetectNotFoundLines(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: false, BaseURL: base, Models: []string{}, Detail: "URLError: refused"}
	}
	s.QuickDetect("ollama", "", "")
	wantLog(t, rec, "Ollama is not running at", "error")

	f.detectOllama = func(base string, t float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base, Models: []string{}}
	}
	s.QuickDetect("ollama", "", "")
	wantLog(t, rec, "NO models are installed", "warning")
}

// ---------------------------------------------------------------------------
// The Sound page: the preview gate.
// ---------------------------------------------------------------------------

func TestPreviewSoundGate(t *testing.T) {
	s, rec, f := newSettingsTest(t)

	// Switched off → the 🔇 line.
	writeCfgFile(t, s.srvCfg(t), `{"sound_enabled": false, "sound_volume": 0.8}`)
	r := s.PreviewSound(false)
	if r.Play {
		t.Fatal("a switched-off chime must not play")
	}
	wantLog(t, rec, "🔇 Preview silent — the chime is switched off (the switch above).", "info")

	// On, but the asset is missing → BatchSound's ONE notice, then
	// silence for the session.
	writeCfgFile(t, s.srvCfg(t), `{"sound_enabled": true, "sound_volume": 0.8}`)
	f.soundDir = t.TempDir()
	r = s.PreviewSound(false)
	if r.Play {
		t.Fatal("a missing asset must not play")
	}
	wantLog(t, rec, "🔕 Success chime disabled — success.wav chime missing", "warning")
	// The one-notice law: the second preview says NOTHING.
	before := len(rec.logs)
	r = s.PreviewSound(true)
	if r.Play {
		t.Fatal("the session-poisoned player must stay silent")
	}
	if len(rec.logs) != before {
		t.Fatalf("the one-notice law broken: %d extra lines", len(rec.logs)-before)
	}

	// A fresh surface + a real wav → the play payload (volume clamped
	// from the config).
	s2, _, f2 := newSettingsTest(t)
	dir := t.TempDir()
	wav := make([]byte, 64)
	for i := range wav {
		wav[i] = byte(i)
	}
	if err := os.WriteFile(filepath.Join(dir, "retry.wav"), wav, 0o644); err != nil {
		t.Fatal(err)
	}
	f2.soundDir = dir
	writeCfgFile(t, s2.srvCfg(t), `{"sound_enabled": true, "sound_volume": 0.25}`)
	r = s2.PreviewSound(true)
	if !r.Play || r.Volume != 0.25 {
		t.Fatalf("the play payload: %+v", r)
	}
	want, _ := base64.StdEncoding.DecodeString(strings.TrimPrefix(r.DataURL, "data:audio/wav;base64,"))
	if len(want) != 64 || want[7] != 7 {
		t.Fatalf("the wav bytes did not round-trip: %d bytes", len(want))
	}
}

// ---------------------------------------------------------------------------
// The pages' workflow wiring: the save door's group laws (the frontend
// calls one group per commit — the editingFinished/toggled law).
// ---------------------------------------------------------------------------

func TestPageCommitSaversCarryOnlyTheirGroup(t *testing.T) {
	s, _, f := newSettingsTest(t)
	writeCfgFile(t, s.srvCfg(t), `{"github_token": "gt", "proxy": {"enabled": false, "type": "socks5", "host": "h", "port": 1080, "use_for_web": true},
                "website_vault_path": "/v/web", "manual_vault_path": "/v/manual"}`)

	// A credentials commit: the proxy group keeps its values.
	s.SaveSettings(&SettingsValues{Credentials: &CredentialsValues{
		APIID: "12345", APIHash: "hash", Phone: "+98", GitHubToken: "gt2"}})
	cfg := s.srvCfg(t)
	px, _ := cfg.Get("proxy")
	om := px.(*storage.OrderedMap)
	if v, _ := om.Get("port"); fmt.Sprint(v) != "1080" {
		t.Fatalf("the proxy group must be untouched by a credentials commit: %v", v)
	}
	if v, _ := cfg.Get("website_vault_path"); v != "/v/web" {
		t.Fatalf("the vault group must be untouched: %v", v)
	}
	if v, _ := cfg.Get("telegram_api_id"); v != 12345 {
		t.Fatalf("telegram_api_id = %v (digits → int)", v)
	}

	// A vault commit with the pipeline toggles.
	github := false
	websites := true
	s.SaveSettings(&SettingsValues{Vault: &VaultValues{
		WebsiteVaultPath: " /v/web2 ", PipelineGithub: &github, PipelineWebsites: &websites}})
	if v, _ := cfg.Get("website_vault_path"); v != "/v/web2" {
		t.Fatalf("website_vault_path = %v (must be trimmed)", v)
	}
	pipes, _ := cfg.Get("pipelines")
	pm := pipes.(*storage.OrderedMap)
	if v, _ := pm.Get("github"); v != false {
		t.Fatalf("pipelines.github = %v", v)
	}
	if v, _ := pm.Get("websites"); v != true {
		t.Fatalf("pipelines.websites = %v", v)
	}
	_ = f
}
