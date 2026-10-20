package app

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// pertests.go — T-79a: the More menu's five per-test doors
// (connection_tests.py:69/182/338/470 + llamacpp.py:377, the pinned
// reference at 98f4fc7). Each door is the reference's own flow — the
// opening line, the per-step verdicts streamed onto the log desk, and
// (for the cloud + llama.cpp doors) the reference's custom message
// box, returned as the (title, body) pair the shell's themed box
// renders. Every network-touching seam rides SettingsDeps — faked at
// construction per THE HERMETIC-SEAM law.
//
// THE OPTION-B RE-SCOPES (ADR-0004 Am.3 — the queue rides the Worker
// HTTP API only; the Telegram MTProto leg is the accepted loss):
//   · Test Telegram & GitHub  → the GitHub token half verbatim + the
//     WORKER QUEUE leg in the Telegram half's seat (the battery's own
//     vocabulary; the opening line re-worded to speak the truth).
//   · Test Proxy Connection   → the sentinel's ProxyReachable TCP
//     verdict (the reference's deep half — a Telegram round trip
//     THROUGH the proxy — is the same accepted loss). The enabled
//     check + its warning line stay verbatim; the verdict sentence is
//     the sentinel's own ("Proxy OK (host:port)").
// The guard difference: the reference's test_telegram_github and
// test_proxy hold the TELEGRAM worker's mutex (dropped with the
// worker); these doors are read-only probes and need no guard — the
// reference's own Ollama/cloud/llama.cpp doors never guarded either.

// TestTelegramGitHub is 🔑 Test Telegram & GitHub — the GitHub half
// (inline, fast — the token as the live config holds it under the
// port's save-on-change law) + the Worker queue leg (the re-scoped
// Telegram half). Returns no box (the reference's door is log-only).
func (s *Settings) TestTelegramGitHub() (string, string) {
	s.SpeakOnBus("🔑 Testing GitHub & the sync queue…", "info")
	token := strings.TrimSpace(pyStrCfg(s.cfg, "github_token"))
	if s.Deps.GitHubGet != nil {
		if token != "" {
			r := s.Deps.GitHubGet("/user", token, GitHubAPIBase, 15.0)
			if r.Status == 200 {
				s.SpeakOnBus(fmt.Sprintf("✅ GitHub token valid (user: %s)",
					githubLoginOf(r)), "success")
			} else {
				s.SpeakOnBus(fmt.Sprintf("❌ GitHub test failed: %s",
					githubPayloadString(r.Payload, r.Err)), "error")
			}
		} else {
			r := s.Deps.GitHubGet("/users/octocat", "", GitHubAPIBase, 15.0)
			if r.Status == 200 {
				s.SpeakOnBus("✅ GitHub public API is accessible.", "success")
			} else {
				s.SpeakOnBus(fmt.Sprintf("❌ GitHub test failed: %s",
					githubPayloadString(r.Payload, r.Err)), "error")
			}
		}
	}
	if s.Deps.WorkerCheck == nil {
		return "", ""
	}
	wc := s.Deps.WorkerCheck(s.cfg)
	if wc == nil {
		// The queue is optional — the honest line (the reference's
		// telegram half always ran; an unconfigured Worker is the
		// port's own "nothing to check" truth).
		s.SpeakOnBus("ℹ️ Worker queue not configured — the sync is optional (set the Worker's URL in Settings ▸ Bot).", "info")
		return "", ""
	}
	s.SpeakOnBus("☁️ Worker queue: "+wc.Detail,
		connectioncheck.GUILevels[wc.Level])
	return "", ""
}

// githubLoginOf pulls the login out of a GET /user payload ("?" when
// the payload refuses to speak one — the token test's own law).
func githubLoginOf(r connectioncheck.GitHubReply) string {
	login := "?"
	if om, ok := r.Payload.(*storage.OrderedMap); ok {
		if v, ok2 := om.Get("login"); ok2 {
			if s2, is := v.(string); is && s2 != "" {
				login = s2
			}
		}
	}
	return login
}

// TestProxy is 🌐 Test Proxy Connection — the enabled check verbatim,
// then the sentinel's own TCP verdict (the re-scoped deep half).
// Returns no box (the reference's door is log-only).
func (s *Settings) TestProxy() (string, string) {
	s.SpeakOnBus("🌐 Testing proxy connection...", "info")
	if s.Deps.ProxyCheck == nil {
		return "", ""
	}
	pr := s.Deps.ProxyCheck(s.cfg, 2.0)
	switch pr.State {
	case "disabled":
		s.SpeakOnBus("⚠️ Proxy is not enabled. Enable it in the Proxy tab.", "warning")
	case "connected":
		s.SpeakOnBus("✅ "+pr.Detail, "success")
	default:
		s.SpeakOnBus("❌ Proxy test failed: "+pr.Detail, "error")
	}
	return "", ""
}

// The Ollama test prompt — the reference's own constant
// (connection_tests.py:369).
const ollamaTestPrompt = "What is 2 + 3? Answer with just the number."

// TestOllama is 🧠 Test Ollama — the connection + model list, the
// configured-model check, then the real inference ping. The empty
// reply retries once on the port's options vocabulary (num_ctx from
// the config + num_predict 50 — the reference's retry also carried
// temperature 0.7, a key the port's golden-pinned Ollama wire shape
// does not carry; recorded in the spec's dated note). Returns no box
// (the reference's door is log-only).
func (s *Settings) TestOllama() (string, string) {
	s.SpeakOnBus("🧠 Testing Ollama...", "info")
	url := strings.TrimSpace(cfgNestedStr(s.cfg, "ollama", "base_url", ""))
	if url == "" {
		url = "http://127.0.0.1:11434"
	}
	model := strings.TrimSpace(cfgNestedStr(s.cfg, "ollama", "model", ""))
	if s.Deps.DetectOllama == nil || s.Deps.OllamaChatPing == nil {
		return "", ""
	}
	// Step 1: connection + the model list (the reference's own shape).
	r := s.Deps.DetectOllama(url, 15.0)
	if !r.Found {
		s.SpeakOnBus(fmt.Sprintf(
			"❌ Ollama test failed: %s. Click '🚀 Start Ollama Server' if it isn't running.",
			r.Detail), "error")
		return "", ""
	}
	s.SpeakOnBus(fmt.Sprintf("✅ Ollama is running at %s. Available models: %s",
		r.BaseURL, strings.Join(r.Models, ", ")), "success")
	if model == "" || !containsString(r.Models, model) {
		s.SpeakOnBus(fmt.Sprintf(
			"⚠️ Model '%s' not found in Ollama. Pull it with: ollama pull %s",
			model, model), "warning")
		return "", ""
	}
	// Step 2: the real inference ping.
	s.SpeakOnBus(fmt.Sprintf("💬 Sending test prompt to '%s'...", model), "info")
	reply, errText := s.Deps.OllamaChatPing(url, model, ollamaTestPrompt,
		&llmclient.OllamaChatOptions{JSONModeOff: true}, 120.0)
	if errText != "" {
		s.SpeakOnBus(fmt.Sprintf(
			"❌ Model test failed (model exists but can't generate): %s",
			errText), "error")
		return "", ""
	}
	if reply != "" {
		s.SpeakOnBus(fmt.Sprintf("✅ Model '%s' responded: \"%s\"",
			model, clipRunes(reply, 100)), "success")
		s.SpeakOnBus("✅ Ollama is fully functional (connection + model + inference).", "success")
		return "", ""
	}
	// The empty-reply retry (the reference's options retry, on the
	// port's options vocabulary).
	s.SpeakOnBus("   (empty response, retrying with options...)", "info")
	numCtx := pyIntOr(cfgVal(s.cfg, "llm_num_ctx"), llmclient.DefaultNumCtx)
	reply, errText = s.Deps.OllamaChatPing(url, model, ollamaTestPrompt,
		&llmclient.OllamaChatOptions{JSONModeOff: true, NumCtx: numCtx, NumPredict: 50}, 120.0)
	if errText != "" {
		s.SpeakOnBus(fmt.Sprintf(
			"❌ Model test failed (model exists but can't generate): %s",
			errText), "error")
		return "", ""
	}
	if reply != "" {
		s.SpeakOnBus(fmt.Sprintf("✅ Model '%s' responded: \"%s\"",
			model, clipRunes(reply, 100)), "success")
		s.SpeakOnBus("✅ Ollama is fully functional (connection + model + inference).", "success")
		return "", ""
	}
	s.SpeakOnBus(fmt.Sprintf(
		"⚠️ Model '%s' returned an empty response even after retry.", model), "warning")
	s.SpeakOnBus("   This model may not work with this prompt format. Try selecting a different model from the dropdown (e.g. Qwen3.5-9B-Uncensored).", "warning")
	// The reference's third line speaks the Python response type — the
	// port's parsed wire shape has no Python type name; recorded in
	// the spec's dated note.
	return "", ""
}

// TestCloudLLM is 🔌 Test Cloud API — the flavor line (the URL decides
// the wire format, exactly like CloudChat), the required-field
// refusals, the warn-never-block /models pre-flight, then the
// "Say hello" chat ping — with the reference's custom message box on
// all three chat outcomes (connected / empty / failed).
func (s *Settings) TestCloudLLM() (string, string) {
	apiURL := strings.TrimSpace(pyStrCfg(s.cfg, "cloud_api_url"))
	flavor := "OpenAI-compatible"
	if llmclient.IsAnthropicURL(apiURL) {
		flavor = "Claude"
	}
	s.SpeakOnBus(fmt.Sprintf("🔌 Testing Cloud API (%s)...", flavor), "info")
	apiKey := strings.TrimSpace(pyStrCfg(s.cfg, "cloud_api_key"))
	model := strings.TrimSpace(pyStrCfg(s.cfg, "cloud_model"))
	if apiURL == "" {
		s.SpeakOnBus("❌ Endpoint test failed: API URL is required.", "error")
		return "", ""
	}
	if model == "" {
		s.SpeakOnBus("❌ Endpoint test failed: Model name is required.", "error")
		return "", ""
	}
	if apiKey == "" {
		s.SpeakOnBus("⚠️ API Key is empty — proceeding anyway (only works for self-hosted servers without auth).", "warning")
	}
	// The /models pre-flight (warn-never-block — the reference's own
	// law; a server that hides /models reports and the ping runs
	// anyway).
	if s.Deps.PreflightCloud != nil {
		ok, message, listed := s.Deps.PreflightCloud(apiURL, apiKey, model, 10.0)
		switch {
		case ok:
			if listed != nil && *listed {
				s.SpeakOnBus(fmt.Sprintf(
					"📋 /models check: %s — '%s' is listed.", message, model), "success")
			} else {
				s.SpeakOnBus(fmt.Sprintf(
					"📋 /models check: %s — '%s' is NOT listed (check the name or load it on the server).",
					message, model), "warning")
			}
		default:
			s.SpeakOnBus(fmt.Sprintf(
				"📋 /models check unavailable: %s — continuing with the chat ping.",
				message), "info")
		}
	}
	// The chat ping.
	s.SpeakOnBus(fmt.Sprintf("💬 Sending test prompt to '%s' at %s...",
		model, apiURL), "info")
	if s.Deps.CloudChatPing == nil {
		return "", ""
	}
	reply, errText := s.Deps.CloudChatPing(apiURL, apiKey, model, "Say hello", 120.0)
	switch {
	case errText != "":
		s.SpeakOnBus(fmt.Sprintf("❌ Cloud LLM test failed: %s", errText), "error")
		return "Cloud LLM — Connection Failed",
			fmt.Sprintf("Could not connect to the Cloud LLM API.\n\n"+
				"Error: %s\n\n"+
				"Check:\n"+
				"  • API URL is correct (e.g. https://api.openai.com/v1)\n"+
				"  • API key is valid\n"+
				"  • Model name is spelled correctly\n"+
				"  • Network/proxy allows the connection", errText)
	case reply != "":
		s.SpeakOnBus(fmt.Sprintf("✅ Cloud LLM responded: \"%s\"",
			clipRunes(reply, 100)), "success")
		s.SpeakOnBus("✅ Cloud LLM is fully functional.", "success")
		return "Cloud LLM Connected",
			fmt.Sprintf("Model '%s' responded:\n\n  \"%s\"\n\n"+
				"Cloud LLM is fully functional.", model, clipRunes(reply, 200))
	default:
		s.SpeakOnBus("⚠️ Cloud LLM returned an empty response. Check the model name and API key.", "warning")
		return "Cloud LLM — Empty Response",
			"The API returned 200 OK but the response content was empty.\n\n" +
				"Check that the model name is correct and that your API key " +
				"has access to it."
	}
}

// TestLlamaCpp is 🦙 Test llama.cpp — the /props probe (with the
// process/port detection fallback), the model resolve (the reference's
// _fill_llamacpp_models pick: the probe's model, else the config's
// current, else the first listed — names carry the props model), then
// the real one-token chat ping through the SAME OpenAI-compatible
// path the batch uses — with the reference's custom message box on the
// not-detected / connected / chat-failed outcomes.
func (s *Settings) TestLlamaCpp() (string, string) {
	s.SpeakOnBus("🦙 Testing llama.cpp…", "info")
	url := strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_api_url"))
	key := strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_api_key"))
	if url == "" {
		s.SpeakOnBus("❌ llama.cpp test failed: Server URL is empty.", "error")
		return "", ""
	}
	if s.Deps.ProbeLlama == nil {
		return "", ""
	}
	probe := s.Deps.ProbeLlama(url, key, llmclient.LLamaCppProbeTimeoutS)
	if probe == nil || !probe.Found {
		if s.Deps.DetectLlama != nil {
			if det := s.Deps.DetectLlama(key); det != nil && det.Found {
				probe = det
			}
		}
	}
	if probe == nil || !probe.Found {
		detail := ""
		if probe != nil {
			detail = probe.Detail
		}
		s.SpeakOnBus(fmt.Sprintf(
			"❌ llama.cpp test failed: %s. Start it with: llama-server -m <model>.gguf --port 8080",
			detail), "error")
		return "llama.cpp — Not Detected",
			"No llama.cpp server was found.\n\n" +
				"Start it with:\n" +
				"  llama-server -m <model>.gguf --port 8080\n\n" +
				"then click '🔍 Detect' in Settings → 🧠 LLM."
	}
	s.SpeakOnBus("✅ "+probe.Detail, "success")
	// The model resolve — the reference's _fill_llamacpp_models pick
	// (probe.Model → the config's current → the first listed; the
	// names carry the props model when it is not already listed).
	names := make([]string, 0, len(probe.Models)+1)
	for _, m := range probe.Models {
		if strings.TrimSpace(m) != "" {
			names = append(names, m)
		}
	}
	if probe.PropsModel != nil && *probe.PropsModel != "" &&
		!containsString(names, *probe.PropsModel) {
		names = append(names, *probe.PropsModel)
	}
	choice := ""
	if probe.Model != nil && *probe.Model != "" {
		choice = *probe.Model
	}
	if choice == "" {
		choice = strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_model"))
	}
	if choice == "" && len(names) > 0 {
		choice = names[0]
	}
	// The reference's fill syncs the live config's model seat in place
	// (the v30 rule) — mirrored, no save (the reference's own door
	// saves nothing; only the limits grab does, and the door never
	// grabs). The write holds the Settings mutex — the same discipline
	// SaveSettings owns for every live-map mutation (the OrderedMap
	// itself is unsynchronized; s.mu is the writers' seat).
	if choice != "" {
		s.mu.Lock()
		s.cfg.Set("llamacpp_model", choice)
		s.mu.Unlock()
	}
	if choice == "" {
		s.SpeakOnBus("❌ Server detected but no model is loaded — start it with -m <model>.gguf.", "error")
		return "", ""
	}
	base := probe.BaseURL
	if base == "" {
		base = url
	}
	// The chat ping — through the same OpenAI-compatible path the
	// batch uses (base + /v1, the CloudChat router's own lane).
	s.SpeakOnBus(fmt.Sprintf("💬 Sending test prompt to '%s' at %s…",
		choice, base), "info")
	if s.Deps.CloudChatPing == nil {
		return "", ""
	}
	reply, errText := s.Deps.CloudChatPing(base+"/v1", key, choice, "Say hello", 120.0)
	switch {
	case errText != "":
		s.SpeakOnBus(fmt.Sprintf("❌ llama.cpp test failed: %s", errText), "error")
		return "llama.cpp — Chat Failed",
			fmt.Sprintf("The server was detected but the chat ping failed.\n\n"+
				"Error: %s\n\n"+
				"Check that the model is fully loaded (⏳ loading state) "+
				"and that the context window (-c) is large enough.", errText)
	case reply != "":
		s.SpeakOnBus(fmt.Sprintf("✅ llama.cpp responded: \"%s\"",
			clipRunes(reply, 100)), "success")
		return "llama.cpp Connected",
			fmt.Sprintf("llama.cpp server at %s answered with model '%s'.\n\n"+
				"Reply: %s", base, choice, clipRunes(reply, 200))
	default:
		s.SpeakOnBus("⚠️ llama.cpp returned an empty response (the model may still be loading).", "warning")
		return "", ""
	}
}
