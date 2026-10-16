package connectioncheck

// llmsection.go — section 2: the ACTIVE LLM provider (cloud API /
// Ollama / llama.cpp), with a down provider earning one hint line
// pointing at the alternatives. The probes ride llmclient (T-30) via
// the package seams (the reference monkeypatches the llm_client module
// attributes; the seams keep the same testability).

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/llmclient"
)

// probeLlamaCppFn / detectLlamaCppFn / preflightCloudFn are the
// llm_client seams (mock.patch.object(_llm, …) in the reference's
// suites). Production: the real llmclient functions.
var (
	probeLlamaCppFn  = llmclient.ProbeLlamaCpp
	detectLlamaCppFn = llmclient.DetectLlamaCpp
	preflightCloudFn = llmclient.PreflightCloud
)

// CheckOllama probes baseURL/api/tags: reachable + models + the
// configured-model check (not pulled → warn; no models → error).
func CheckOllama(baseURL any, configuredModel any) Result {
	base := trimPySpace(pyStr(pyOr(baseURL, "http://127.0.0.1:11434")))
	u := strings.TrimRight(base, "/") + "/api/tags"
	raw, err := httpGetJSON(u, 6.0)
	if err != nil {
		return result("Ollama", LevelError,
			"not reachable at "+base+" ("+errClass(err)+") — "+
				"start the Ollama app, then retry")
	}
	var models []string
	if list, ok := dictValue(raw, "models").([]any); ok {
		for _, m := range list {
			name := trimPySpace(pyStr(pyOr(dictValue(m, "name"),
				dictValue(m, "model"))))
			if name != "" {
				models = append(models, name)
			}
		}
	}
	if len(models) == 0 {
		return result("Ollama", LevelError,
			"up @ "+base+" but NO models installed — run "+
				"`ollama pull <model>` first")
	}
	detail := "up @ " + base + " · " +
		itoa(len(models)) + " model(s)"
	cfgModel := trimPySpace(pyStr(pyOr(configuredModel, "")))
	if cfgModel != "" {
		if containsStr(models, cfgModel) {
			return result("Ollama", LevelOK,
				detail+" · '"+cfgModel+"' ready")
		}
		return result("Ollama", LevelWarn,
			detail+" — configured '"+cfgModel+"' is NOT pulled "+
				"(the app will prompt at run time)")
	}
	return result("Ollama", LevelOK,
		detail+" · no model configured (picked at run time)")
}

// CheckLlamaCpp probes the configured (or detected) llama-server:
// probe → detect fallback → the model verdict with the /props limits
// grab (v0.70.0 — the launch limits the server published).
func CheckLlamaCpp(config any) Result {
	u := trimPySpace(pyStr(pyOr(cfgValue(config, "llamacpp_api_url"), "")))
	if u == "" {
		u = llmclient.LLamaCppDefaultBase
	}
	key := pyStr(pyOr(cfgValue(config, "llamacpp_api_key"), ""))
	probe := probeLlamaCppFn(u, key, llmclient.LLamaCppProbeTimeoutS)
	if !probe.Found {
		detected := detectLlamaCppFn(key, nil, "",
			llmclient.LLamaCppProbeTimeoutS, true)
		if detected != nil {
			probe = detected
		}
	}
	if !probe.Found {
		det := probe.Detail
		if det == "" {
			det = "no answer"
		}
		return result("llama.cpp", LevelError,
			"server not detected ("+det+") "+
				"— start it with: llama-server -m <model>.gguf --port 8080")
	}
	base := probe.BaseURL
	if base == "" {
		base = u
	}
	loading := ""
	if probe.Ready != nil && !*probe.Ready {
		loading = " · still loading"
	}
	var model *string
	if probe != nil {
		model = llmclient.ResolveLlamaCppModel(config, probe.Models,
			probe.PropsModel)
	}
	if model == nil {
		return result("llama.cpp", LevelError,
			"up @ "+base+" but no model loaded — start llama-server "+
				"with -m <model>.gguf")
	}
	// v0.70.0 — THE SERVER CONFIG GRAB: the Test-Connection line speaks
	// the launch limits the server published on /props.
	limits := llmclient.LlamaCppLimitsLine(probe.PropsLimits)
	limitsNote := ""
	if limits != "" {
		limitsNote = " · " + limits
	}
	return result("llama.cpp", LevelOK,
		"up @ "+base+" · model '"+*model+"'"+loading+limitsNote)
}

// CheckCloud preflights the configured cloud API (routes by URL:
// api.anthropic.com → the Claude /v1/models preflight, everything
// else → OpenAI-compatible).
func CheckCloud(config any) Result {
	u := trimPySpace(pyStr(pyOr(cfgValue(config, "cloud_api_url"), "")))
	key := pyStr(pyOr(cfgValue(config, "cloud_api_key"), ""))
	model := trimPySpace(pyStr(pyOr(cfgValue(config, "cloud_model"), "")))
	if u == "" {
		return result("Cloud API", LevelError,
			"no API URL configured — Settings ▸ LLM")
	}
	flavor := "OpenAI-compatible"
	if llmclient.IsAnthropicURL(u) {
		flavor = "Claude"
	}
	var (
		endpointOK  bool
		message     string
		modelListed *bool
		probeErr    error
	)
	func() {
		defer func() {
			if r := recover(); r != nil {
				probeErr = panicAsError(r)
			}
		}()
		endpointOK, message, modelListed = preflightCloudFn(u, key, model, 15)
	}()
	if probeErr != nil {
		// preflight is never-a-gate; belt & braces
		return result("Cloud API", LevelError,
			u+" — check failed ("+errClassColon(probeErr)+")")
	}
	if !endpointOK {
		return result("Cloud API", LevelError,
			u+" ("+flavor+") — "+message)
	}
	detail := u + " (" + flavor + ") — " + message
	if modelListed != nil && *modelListed {
		return result("Cloud API", LevelOK,
			detail+" · '"+model+"' listed")
	}
	if modelListed != nil && !*modelListed {
		return result("Cloud API", LevelWarn,
			detail+" — model '"+model+"' NOT in the list (trying anyway)")
	}
	return result("Cloud API", LevelOK,
		detail+" · model '"+model+"' (model list hidden — fine)")
}

// CheckLLM routes to the ACTIVE provider; a down provider earns one
// hint line pointing at the alternatives.
func CheckLLM(config any) []Result {
	provider := trimPySpace(pyStr(pyOr(cfgValue(config, "llm_provider"), "ollama")))
	if provider == "" {
		provider = "ollama"
	}
	var out []Result
	switch provider {
	case "llamacpp":
		out = []Result{CheckLlamaCpp(config)}
	case "cloud":
		out = []Result{CheckCloud(config)}
	default:
		oll := cfgValue(config, "ollama")
		out = []Result{CheckOllama(dictValueAny(oll, "base_url",
			"http://127.0.0.1:11434"), dictValue(oll, "model"))}
	}
	if len(out) > 0 && out[0].Level == LevelError {
		out = append(out, result("Hint", LevelInfo,
			"the active provider is down — start it, or switch "+
				"provider in Settings ▸ LLM (Ollama / cloud API / llama.cpp)"))
	}
	return out
}

// dictValueAny is dict.get(key, default) on a nested dict value.
func dictValueAny(m any, key string, def any) any {
	if v, ok := cfgGet(m, key); ok {
		return pyOr(v, def)
	}
	return def
}

func containsStr(haystack []string, needle string) bool {
	for _, s := range haystack {
		if s == needle {
			return true
		}
	}
	return false
}

func itoa(n int) string {
	return pyStr(n)
}
