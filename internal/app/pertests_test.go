package app

import (
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// pertests_test.go — T-79a: the five per-test doors' line grammar,
// pinned against the reference's own flows (connection_tests.py:69/
// 182/338/470 + llamacpp.py:377 at 98f4fc7). Every seam faked at
// construction; every line asserted by level + verbatim text.

func wantLogAt(t *testing.T, rec *busRec, part, level string, idx int) LogLine {
	t.Helper()
	l, ok := rec.findLog(part)
	if !ok {
		t.Fatalf("step %d: no log line contains %q (have %d lines)", idx, part, len(rec.logs))
	}
	if l.Level != level {
		t.Fatalf("step %d: line %q has level %q, want %q", idx, part, l.Level, level)
	}
	return l
}

// ---------------------------------------------------------------------------
// 🔑 Test Telegram & GitHub (the Option-B re-scope).
// ---------------------------------------------------------------------------

func TestPerTestTelegramGitHubTokenValid(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"github_token": "tok-alice"}`)
	f.githubReplies["tok-alice"] = connectioncheck.GitHubReply{
		Status: 200, Payload: mapLoginPayload("alice")}
	f.workerCheck = func(config any) *connectioncheck.Result {
		return &connectioncheck.Result{Name: "Bot Worker", Level: connectioncheck.LevelOK,
			Detail: "v0.40.0 — matches this app (expected v0.40.0)"}
	}
	title, _ := s.TestTelegramGitHub()
	wantLogAt(t, rec, "🔑 Testing GitHub & the sync queue…", "info", 1)
	wantLogAt(t, rec, "✅ GitHub token valid (user: alice)", "success", 2)
	l := wantLogAt(t, rec, "☁️ Worker queue: v0.40.0 — matches this app", "success", 3)
	if !strings.HasPrefix(l.Msg, "☁️ Worker queue: ") {
		t.Fatalf("worker line frame: %q", l.Msg)
	}
	if title != "" {
		t.Fatalf("door is log-only, title = %q", title)
	}
}

func TestPerTestTelegramGitHubAnonymous(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	f.githubReplies[""] = connectioncheck.GitHubReply{Status: 200}
	f.workerCheck = func(config any) *connectioncheck.Result { return nil }
	s.TestTelegramGitHub()
	wantLogAt(t, rec, "✅ GitHub public API is accessible.", "success", 1)
	wantLogAt(t, rec, "ℹ️ Worker queue not configured — the sync is optional", "info", 2)
}

func TestPerTestTelegramGitHubFailureCarriesPayloadMessage(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"github_token": "tok-bad"}`)
	f.githubReplies["tok-bad"] = connectioncheck.GitHubReply{
		Status: 401, Payload: mapLoginPayload("")}
	payloadMsg := storage.NewOrderedMap()
	payloadMsg.Set("message", "Bad credentials")
	f.githubReplies["tok-bad"] = connectioncheck.GitHubReply{Status: 401, Payload: payloadMsg}
	s.TestTelegramGitHub()
	wantLogAt(t, rec, "❌ GitHub test failed: Bad credentials", "error", 1)
}

func TestPerTestTelegramGitHubWorkerLevels(t *testing.T) {
	for _, tc := range []struct {
		level, wantGUI string
	}{
		{connectioncheck.LevelWarn, "warning"},
		{connectioncheck.LevelError, "error"},
		{connectioncheck.LevelInfo, "info"},
	} {
		s, rec, f := newSettingsTest(t)
		f.workerCheck = func(config any) *connectioncheck.Result {
			return &connectioncheck.Result{Name: "Bot Worker", Level: tc.level, Detail: "the detail"}
		}
		s.TestTelegramGitHub()
		wantLogAt(t, rec, "☁️ Worker queue: the detail", tc.wantGUI, 1)
	}
}

// mapLoginPayload builds a GET /user-shaped payload.
func mapLoginPayload(login string) *storage.OrderedMap {
	m := storage.NewOrderedMap()
	m.Set("login", login)
	return m
}

// ---------------------------------------------------------------------------
// 🌐 Test Proxy Connection.
// ---------------------------------------------------------------------------

func TestPerTestProxyDisabled(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	f.proxyCheck = func(config any, timeoutS float64) connectioncheck.ProxyResult {
		return connectioncheck.ProxyResult{State: "disabled", Detail: "Proxy disabled"}
	}
	title, _ := s.TestProxy()
	wantLogAt(t, rec, "🌐 Testing proxy connection...", "info", 1)
	wantLogAt(t, rec, "⚠️ Proxy is not enabled. Enable it in the Proxy tab.", "warning", 2)
	if title != "" {
		t.Fatalf("door is log-only, title = %q", title)
	}
}

func TestPerTestProxyConnectedSpeaksTheSentinelSentence(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	f.proxyCheck = func(config any, timeoutS float64) connectioncheck.ProxyResult {
		return connectioncheck.ProxyResult{State: "connected", Host: "127.0.0.1", Port: 10808,
			Detail: "Proxy OK (127.0.0.1:10808)"}
	}
	s.TestProxy()
	wantLogAt(t, rec, "✅ Proxy OK (127.0.0.1:10808)", "success", 1)
}

func TestPerTestProxyUnreachable(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	f.proxyCheck = func(config any, timeoutS float64) connectioncheck.ProxyResult {
		return connectioncheck.ProxyResult{State: "unreachable", Host: "127.0.0.1", Port: 10808,
			Detail: "Proxy unreachable (127.0.0.1:10808) — errno 111"}
	}
	s.TestProxy()
	wantLogAt(t, rec, "❌ Proxy test failed: Proxy unreachable (127.0.0.1:10808) — errno 111", "error", 1)
}

// ---------------------------------------------------------------------------
// 🧠 Test Ollama.
// ---------------------------------------------------------------------------

func TestPerTestOllamaNotRunning(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"}}`)
	// the rig's default DetectOllama fake: not found, URLError detail
	title, _ := s.TestOllama()
	wantLogAt(t, rec, "🧠 Testing Ollama...", "info", 1)
	wantLogAt(t, rec, "❌ Ollama test failed: URLError: refused. Click '🚀 Start Ollama Server' if it isn't running.", "error", 2)
	if title != "" {
		t.Fatalf("door is log-only, title = %q", title)
	}
}

func TestPerTestOllamaModelMissing(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"}}`)
	f.detectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"qwen", "mistral"}, Detail: "ok"}
	}
	s.TestOllama()
	wantLogAt(t, rec, "✅ Ollama is running at http://127.0.0.1:11434. Available models: qwen, mistral", "success", 1)
	wantLogAt(t, rec, "⚠️ Model 'llama3' not found in Ollama. Pull it with: ollama pull llama3", "warning", 2)
	if _, ok := rec.findLog("💬 Sending test prompt"); ok {
		t.Fatalf("no prompt is sent when the model is missing")
	}
}

func TestPerTestOllamaInferenceSucceeds(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"}}`)
	f.detectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3"}, Detail: "ok"}
	}
	var gotPrompt string
	var gotJSONOff bool
	f.ollamaChatPing = func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (string, string) {
		gotPrompt = prompt
		gotJSONOff = opts != nil && opts.JSONModeOff
		return "5", ""
	}
	s.TestOllama()
	if gotPrompt != "What is 2 + 3? Answer with just the number." {
		t.Fatalf("prompt = %q", gotPrompt)
	}
	if !gotJSONOff {
		t.Fatalf("the test ping is a PLAIN chat (JSON mode off)")
	}
	wantLogAt(t, rec, "💬 Sending test prompt to 'llama3'...", "info", 1)
	wantLogAt(t, rec, `✅ Model 'llama3' responded: "5"`, "success", 2)
	wantLogAt(t, rec, "✅ Ollama is fully functional (connection + model + inference).", "success", 3)
}

func TestPerTestOllamaEmptyThenRetrySucceeds(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"}}`)
	f.detectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3"}, Detail: "ok"}
	}
	calls := 0
	var secondOpts *llmclient.OllamaChatOptions
	f.ollamaChatPing = func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (string, string) {
		calls++
		if calls == 2 {
			secondOpts = opts
			return "5", ""
		}
		return "", ""
	}
	s.TestOllama()
	if calls != 2 {
		t.Fatalf("calls = %d, want 2 (the options retry)", calls)
	}
	if secondOpts == nil || secondOpts.NumPredict != 50 || secondOpts.NumCtx != llmclient.DefaultNumCtx {
		t.Fatalf("retry options = %+v (want num_predict 50 + the config's num_ctx)", secondOpts)
	}
	wantLogAt(t, rec, "   (empty response, retrying with options...)", "info", 1)
	wantLogAt(t, rec, `✅ Model 'llama3' responded: "5"`, "success", 2)
	wantLogAt(t, rec, "✅ Ollama is fully functional (connection + model + inference).", "success", 3)
}

func TestPerTestOllamaEmptyEvenAfterRetry(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"}}`)
	f.detectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3"}, Detail: "ok"}
	}
	f.ollamaChatPing = func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (string, string) {
		return "", ""
	}
	s.TestOllama()
	wantLogAt(t, rec, "⚠️ Model 'llama3' returned an empty response even after retry.", "warning", 1)
	wantLogAt(t, rec, "   This model may not work with this prompt format. Try selecting a different model from the dropdown (e.g. Qwen3.5-9B-Uncensored).", "warning", 2)
}

func TestPerTestOllamaPingError(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"}}`)
	f.detectOllama = func(base string, timeoutS float64) llmclient.DetectOllamaResult {
		return llmclient.DetectOllamaResult{Found: true, BaseURL: base,
			Models: []string{"llama3"}, Detail: "ok"}
	}
	f.ollamaChatPing = func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (string, string) {
		return "", "boom"
	}
	s.TestOllama()
	wantLogAt(t, rec, "❌ Model test failed (model exists but can't generate): boom", "error", 1)
}

// ---------------------------------------------------------------------------
// 🔌 Test Cloud API.
// ---------------------------------------------------------------------------

func TestPerTestCloudRequiresURLAndModel(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	title, _ := s.TestCloudLLM()
	wantLogAt(t, rec, "❌ Endpoint test failed: API URL is required.", "error", 1)
	if title != "" {
		t.Fatalf("no box on the refusal, title = %q", title)
	}
	s2, rec2, _ := newSettingsTest(t)
	writeCfgFile(t, s2.cfg, `{"cloud_api_url": "https://api.openai.com/v1"}`)
	s2.TestCloudLLM()
	wantLogAt(t, rec2, "❌ Endpoint test failed: Model name is required.", "error", 1)
}

func TestPerTestCloudFlavorFollowsTheURL(t *testing.T) {
	for _, tc := range []struct{ url, flavor string }{
		{"https://api.anthropic.com", "Claude"},
		{"https://api.openai.com/v1", "OpenAI-compatible"},
	} {
		s, rec, f := newSettingsTest(t)
		writeCfgFile(t, s.cfg, `{"cloud_api_url": "`+tc.url+`", "cloud_api_key": "k", "cloud_model": "m"}`)
		f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
			return "", "down"
		}
		s.TestCloudLLM()
		wantLogAt(t, rec, "🔌 Testing Cloud API ("+tc.flavor+")...", "info", 1)
	}
}

func TestPerTestCloudEmptyKeyWarns(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"cloud_api_url": "https://api.openai.com/v1", "cloud_model": "m"}`)
	f.preflightCloud = func(apiURL, apiKey, model string, timeoutS float64) (bool, string, *bool) {
		return false, "404", nil
	}
	f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		return "", "down"
	}
	s.TestCloudLLM()
	wantLogAt(t, rec, "⚠️ API Key is empty — proceeding anyway (only works for self-hosted servers without auth).", "warning", 1)
}

func TestPerTestCloudPreflightGrammar(t *testing.T) {
	listed := true
	notListed := false
	for _, tc := range []struct {
		name, wantLine, wantLevel string
		ok                        bool
		listed                    *bool
	}{
		{"listed", "📋 /models check: 12 models — 'm' is listed.", "success", true, &listed},
		{"not-listed", "📋 /models check: 12 models — 'm' is NOT listed (check the name or load it on the server).", "warning", true, &notListed},
		{"unavailable", "📋 /models check unavailable: 12 models — continuing with the chat ping.", "info", false, nil},
	} {
		s, rec, f := newSettingsTest(t)
		writeCfgFile(t, s.cfg, `{"cloud_api_url": "https://api.openai.com/v1", "cloud_api_key": "k", "cloud_model": "m"}`)
		f.preflightCloud = func(apiURL, apiKey, model string, timeoutS float64) (bool, string, *bool) {
			return tc.ok, "12 models", tc.listed
		}
		f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
			return "", "down"
		}
		s.TestCloudLLM()
		wantLogAt(t, rec, tc.wantLine, tc.wantLevel, 1)
	}
}

func TestPerTestCloudConnectedBox(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"cloud_api_url": "https://api.openai.com/v1", "cloud_api_key": "k", "cloud_model": "gpt-x"}`)
	f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		if prompt != "Say hello" {
			t.Fatalf("prompt = %q", prompt)
		}
		return "Hello there!", ""
	}
	title, body := s.TestCloudLLM()
	wantLogAt(t, rec, `✅ Cloud LLM responded: "Hello there!"`, "success", 1)
	wantLogAt(t, rec, "✅ Cloud LLM is fully functional.", "success", 2)
	if title != "Cloud LLM Connected" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{
		"Model 'gpt-x' responded:", "\"Hello there!\"", "Cloud LLM is fully functional.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q: %q", want, body)
		}
	}
}

func TestPerTestCloudEmptyResponseBox(t *testing.T) {
	s, _, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"cloud_api_url": "https://api.openai.com/v1", "cloud_api_key": "k", "cloud_model": "m"}`)
	f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		return "", ""
	}
	title, body := s.TestCloudLLM()
	if title != "Cloud LLM — Empty Response" {
		t.Fatalf("title = %q", title)
	}
	if !strings.Contains(body, "The API returned 200 OK but the response content was empty.") {
		t.Fatalf("body: %q", body)
	}
}

func TestPerTestCloudFailedBoxCarriesTheChecklist(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"cloud_api_url": "https://api.openai.com/v1", "cloud_api_key": "k", "cloud_model": "m"}`)
	f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		return "", "dial timeout"
	}
	title, body := s.TestCloudLLM()
	wantLogAt(t, rec, "❌ Cloud LLM test failed: dial timeout", "error", 1)
	if title != "Cloud LLM — Connection Failed" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{
		"Error: dial timeout", "API URL is correct (e.g. https://api.openai.com/v1)",
		"API key is valid", "Model name is spelled correctly",
		"Network/proxy allows the connection",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q: %q", want, body)
		}
	}
}

// ---------------------------------------------------------------------------
// 🦙 Test llama.cpp.
// ---------------------------------------------------------------------------

func TestPerTestLlamaCppEmptyURL(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	title, _ := s.TestLlamaCpp()
	wantLogAt(t, rec, "❌ llama.cpp test failed: Server URL is empty.", "error", 1)
	if title != "" {
		t.Fatalf("no box on the refusal, title = %q", title)
	}
}

func TestPerTestLlamaCppNotDetectedBox(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080"}`)
	f.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: false, BaseURL: baseURL,
			Models: []string{}, Detail: "URLError: refused"}
	}
	f.detectLlama = func(apiKey string) *llmclient.ProbeResult { return nil }
	title, body := s.TestLlamaCpp()
	wantLogAt(t, rec, "❌ llama.cpp test failed: URLError: refused. Start it with: llama-server -m <model>.gguf --port 8080", "error", 1)
	if title != "llama.cpp — Not Detected" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{
		"No llama.cpp server was found.", "llama-server -m <model>.gguf --port 8080",
		"then click '🔍 Detect' in Settings → 🧠 LLM.",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q: %q", want, body)
		}
	}
}

func TestPerTestLlamaCppDetectFallbackFindsIt(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080"}`)
	f.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: false, BaseURL: baseURL, Models: []string{}}
	}
	model := "qwen3"
	f.detectLlama = func(apiKey string) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: "http://127.0.0.1:8090",
			Models: []string{model}, Model: &model,
			Detail: "llama.cpp server at 127.0.0.1:8090 (process port)"}
	}
	f.cloudChatPing = func(apiURL, apiKey, model2, prompt string, timeoutS float64) (string, string) {
		if apiURL != "http://127.0.0.1:8090/v1" {
			t.Fatalf("chat apiURL = %q (the OpenAI-compatible path over the DETECTED base)", apiURL)
		}
		return "hi", ""
	}
	title, _ := s.TestLlamaCpp()
	wantLogAt(t, rec, "✅ llama.cpp server at 127.0.0.1:8090 (process port)", "success", 1)
	wantLogAt(t, rec, "💬 Sending test prompt to 'qwen3' at http://127.0.0.1:8090…", "info", 2)
	if title != "llama.cpp Connected" {
		t.Fatalf("title = %q", title)
	}
	// the live config's model seat syncs in place (the v30 rule)
	if v, ok := s.cfg.Get("llamacpp_model"); !ok || v != "qwen3" {
		t.Fatalf("llamacpp_model seat = %v, ok=%v", v, ok)
	}
}

func TestPerTestLlamaCppModelPickOrder(t *testing.T) {
	// probe.Model wins; else the config's current; else the first
	// listed (the names carry the props model).
	probeModel := "from-probe"
	s, _, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080", "llamacpp_model": "from-config"}`)
	f.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
			Models: []string{"listed-a"}, Model: &probeModel,
			Detail: "found"}
	}
	var got string
	f.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		got = model
		return "hi", ""
	}
	s.TestLlamaCpp()
	if got != "from-probe" {
		t.Fatalf("pick = %q, want the probe's own model", got)
	}

	s2, _, f2 := newSettingsTest(t)
	writeCfgFile(t, s2.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080", "llamacpp_model": "from-config"}`)
	f2.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
			Models: []string{"listed-a"}, Detail: "found"}
	}
	f2.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		if model != "from-config" {
			t.Fatalf("pick = %q, want the config's current", model)
		}
		return "hi", ""
	}
	s2.TestLlamaCpp()

	s3, _, f3 := newSettingsTest(t)
	writeCfgFile(t, s3.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080"}`)
	propsModel := "from-props"
	f3.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
			Models: []string{}, PropsModel: &propsModel, Detail: "found"}
	}
	f3.cloudChatPing = func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
		if model != "from-props" {
			t.Fatalf("pick = %q, want the props model (appended to the names)", model)
		}
		return "hi", ""
	}
	s3.TestLlamaCpp()
}

func TestPerTestLlamaCppNoModelLoaded(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080"}`)
	f.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
			Models: []string{}, Detail: "found"}
	}
	title, _ := s.TestLlamaCpp()
	wantLogAt(t, rec, "❌ Server detected but no model is loaded — start it with -m <model>.gguf.", "error", 1)
	if title != "" {
		t.Fatalf("no box, title = %q", title)
	}
}

func TestPerTestLlamaCppChatFailedBox(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080"}`)
	model := "qwen3"
	f.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
			Models: []string{model}, Model: &model, Detail: "found"}
	}
	f.cloudChatPing = func(apiURL, apiKey, model2, prompt string, timeoutS float64) (string, string) {
		if apiURL != "http://127.0.0.1:8080/v1" {
			t.Fatalf("chat apiURL = %q", apiURL)
		}
		if prompt != "Say hello" {
			t.Fatalf("prompt = %q", prompt)
		}
		return "", "context too small"
	}
	title, body := s.TestLlamaCpp()
	wantLogAt(t, rec, "❌ llama.cpp test failed: context too small", "error", 1)
	if title != "llama.cpp — Chat Failed" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{
		"The server was detected but the chat ping failed.", "Error: context too small",
		"Check that the model is fully loaded (⏳ loading state)",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("body missing %q: %q", want, body)
		}
	}
}

func TestPerTestLlamaCppEmptyReplyWarns(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	writeCfgFile(t, s.cfg, `{"llamacpp_api_url": "http://127.0.0.1:8080"}`)
	model := "qwen3"
	f.probeLlama = func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
		return &llmclient.ProbeResult{Found: true, BaseURL: baseURL,
			Models: []string{model}, Model: &model, Detail: "found"}
	}
	f.cloudChatPing = func(apiURL, apiKey, model2, prompt string, timeoutS float64) (string, string) {
		return "", ""
	}
	title, _ := s.TestLlamaCpp()
	wantLogAt(t, rec, "⚠️ llama.cpp returned an empty response (the model may still be loading).", "warning", 1)
	if title != "" {
		t.Fatalf("no box on the empty reply, title = %q", title)
	}
}
