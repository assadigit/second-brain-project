// llamacpp.go — the llama.cpp detection engine (llm_client.py's
// llama.cpp block): URL normalization, the /props fingerprint, the
// /props limits grab, the /v1/models fingerprints, the probe, the port
// scan, the model resolution, and the startup auto-detect policy.
package llmclient

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// asOM adapts a Python-dict-shaped value (the decoded *storage.
// OrderedMap from the wire, or a plain map[string]any from tests and
// golden drivers) to the OrderedMap view the props functions read.
func asOM(v any) (*storage.OrderedMap, bool) {
	switch t := v.(type) {
	case *storage.OrderedMap:
		return t, true
	case map[string]any:
		om := storage.NewOrderedMap()
		for k, val := range t {
			om.Set(k, val)
		}
		return om, true
	}
	return nil, false
}

// llamacppPropsMarkers are the /props keys that fingerprint llama.cpp
// across server builds.
var llamacppPropsMarkers = []string{
	"model_path", "model_alias", "default_generation_settings", "total_slots",
}

// NormalizeLlamaCppAPIURL turns user input into the OpenAI-compatible
// base URL for a llama.cpp server: every shape people actually type —
// bare host:port, scheme without path, trailing slash, a full /v1 base
// — becomes the …/v1 base; a deeper path (reverse proxies) is kept; a
// URL that already carries a path beyond the bare host is the user's
// choice. Empty input yields the llama-server default.
func NormalizeLlamaCppAPIURL(rawURL string) string {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		raw = LLamaCppDefaultBase
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	base := strings.TrimRight(raw, "/")
	parts, err := url.Parse(base)
	if err != nil {
		return LLamaCppDefaultBase + "/v1"
	}
	if parts.Path == "" || parts.Path == "/" {
		return base + "/v1"
	}
	return base
}

// IsLlamaCppProps reports whether a JSON body has the llama.cpp /props
// shape (the marker keys guard against a reverse-proxied other service
// that happens to expose /props).
func IsLlamaCppProps(props any) bool {
	om, ok := asOM(props)
	if !ok {
		return false
	}
	for _, marker := range llamacppPropsMarkers {
		if _, p := om.Get(marker); p {
			return true
		}
	}
	return false
}

// LlamaCppModelFromProps is the best model name out of a /props
// payload: model_alias (skipping the 'unknown' placeholder), else the
// basename of model_path (backslashes normalized). nil when the
// payload names nothing.
func LlamaCppModelFromProps(props any) *string {
	om, ok := asOM(props)
	if !ok {
		return nil
	}
	alias := ""
	if av, p := om.Get("model_alias"); p {
		alias = strings.TrimSpace(helpers.PyStr(pyOr(av, "")))
	}
	if alias != "" && !strings.EqualFold(alias, "unknown") {
		return &alias
	}
	path := ""
	if pv, p := om.Get("model_path"); p {
		path = strings.TrimSpace(helpers.PyStr(pyOr(pv, "")))
	}
	if path != "" {
		p := strings.ReplaceAll(path, "\\", "/")
		p = strings.TrimRight(p, "/")
		if idx := strings.LastIndex(p, "/"); idx >= 0 {
			p = p[idx+1:]
		}
		return &p
	}
	return nil
}

// LlamaCppLimits is the /props launch-limits grab
// ({n_ctx, n_predict} — nil means "the server did not say").
type LlamaCppLimits struct {
	NCtx     *int `json:"n_ctx"`
	NPredict *int `json:"n_predict"`
}

// LlamaCppLimitsFromProps is THE SERVER CONFIG GRAB: llama-server
// publishes its live launch config on /props —
// default_generation_settings.n_ctx is the context window (-c), and
// n_predict (top-level inside the settings, or in params on newer
// builds) is the output cap. -1/0 read as nil (llama.cpp's spelling of
// "no cap" — a server without a cap never writes a bogus 1-token limit
// into Settings). Pure; never raises; junk in, Nones out.
func LlamaCppLimitsFromProps(props any) *LlamaCppLimits {
	out := &LlamaCppLimits{}
	om, ok := asOM(props)
	if !ok {
		return out
	}
	settingsV, _ := om.Get("default_generation_settings")
	settings, ok := asOM(settingsV)
	if !ok {
		return out
	}
	if nCtx := pyInt(pyOr(mapGet(settings, "n_ctx"), 0)); nCtx > 0 {
		out.NCtx = &nCtx
	}
	// The output cap: newer builds put n_predict inside params; some
	// builds expose it at the settings' top level.
	rawPredict, present := settings.Get("n_predict")
	if !present || rawPredict == nil {
		if paramsV, p := settings.Get("params"); p {
			if params, ok := asOM(paramsV); ok {
				rawPredict, present = params.Get("n_predict")
			}
		}
	}
	if present && rawPredict != nil {
		if n := pyInt(rawPredict); n > 0 {
			out.NPredict = &n
		}
	}
	return out
}

func mapGet(om *storage.OrderedMap, key string) any {
	v, _ := om.Get(key)
	return v
}

// llamaCppLimitsLine is one human line for the grabbed server limits
// (” when the server said nothing worth saying).
func llamaCppLimitsLine(limits *LlamaCppLimits) string {
	if limits == nil {
		return ""
	}
	bits := []string{}
	if limits.NCtx != nil && *limits.NCtx > 0 {
		bits = append(bits, fmt.Sprintf("context window %d tokens", *limits.NCtx))
	}
	if limits.NPredict != nil && *limits.NPredict > 0 {
		bits = append(bits, fmt.Sprintf("output cap %d tokens", *limits.NPredict))
	}
	return strings.Join(bits, " · ")
}

// LlamaCppLimitsLine is the exported seam for the battery's
// Test-Connection line (connection_check.check_llamacpp's
// _llm._llamacpp_limits_line call — T-33). Same law: "" when the
// server said nothing worth saying.
func LlamaCppLimitsLine(limits *LlamaCppLimits) string {
	return llamaCppLimitsLine(limits)
}

// modelsFingerprint checks the llama.cpp fingerprints on a /v1/models
// response (used ONLY when /props did not identify the server): the
// Server header ('server-header'), owned_by ('owned_by'), or a .gguf
// id ('gguf-id'). Every Server header occurrence is checked.
func modelsFingerprint(payload any, serverHeaders []string) string {
	for _, v := range serverHeaders {
		if strings.Contains(strings.ToLower(v), "llama.cpp") {
			return "server-header"
		}
	}
	om, ok := asOM(payload)
	if !ok {
		return ""
	}
	dataV, _ := om.Get("data")
	list, ok := dataV.([]any)
	if !ok {
		return ""
	}
	for _, m := range list {
		mom, ok := m.(*storage.OrderedMap)
		if !ok {
			continue
		}
		if ov, p := mom.Get("owned_by"); p {
			if strings.ToLower(helpers.PyStr(pyOr(ov, ""))) == "llama.cpp" {
				return "owned_by"
			}
		}
		if iv, p := mom.Get("id"); p {
			if strings.HasSuffix(strings.ToLower(
				helpers.PyStr(pyOr(iv, ""))), ".gguf") {
				return "gguf-id"
			}
		}
	}
	return ""
}

// ---------------------------------------------------------------------------
// probe_llamacpp
// ---------------------------------------------------------------------------

// ProbeResult is the probe dict (JSON-tagged for the behavior golden;
// Via rides omitempty — the plain probe has no via key, detect sets
// it).
type ProbeResult struct {
	Found        bool            `json:"found"`
	Ready        *bool           `json:"ready"`
	Models       []string        `json:"models"`
	Model        *string         `json:"model"`
	PropsModel   *string         `json:"props_model"`
	IdentifiedBy *string         `json:"identified_by"`
	BaseURL      string          `json:"base_url"`
	Detail       string          `json:"detail"`
	PropsLimits  *LlamaCppLimits `json:"props_limits"`
	Via          string          `json:"via,omitempty"`
}

// ProbeLlamaCpp probes ONE address for a llama.cpp server. NEVER
// raises — the result dict carries everything: /props (the primary
// positive identification, carrying the limits grab), the /v1/models
// fingerprints for props-less builds, /health for ready-vs-loading,
// and the model list. Every request bypasses the proxy entirely
// (llama-server is a local service).
func ProbeLlamaCpp(baseURL, apiKey string, timeoutS float64) *ProbeResult {
	api := NormalizeLlamaCppAPIURL(baseURL)
	root := api
	if strings.HasSuffix(api, "/v1") {
		root = api[:len(api)-3]
	}

	result := &ProbeResult{Found: false, Models: []string{}, BaseURL: root}

	// 1) /props — the primary positive llama.cpp identification.
	var propsVal any
	propsErr := ""
	code, raw, _, err := probeGetJSON(root+"/props", apiKey, timeoutS)
	switch e := err.(type) {
	case nil:
		if code == 200 {
			if v, perr := storage.DecodeJSONValue([]byte(raw)); perr == nil {
				propsVal = v
			}
		}
	case *httpStatusError:
		propsErr = fmt.Sprintf("/props answered HTTP %d", e.code)
	default:
		propsErr = fmt.Sprintf("could not reach %s (URLError)", root)
	}
	if propsVal != nil && IsLlamaCppProps(propsVal) {
		result.Found = true
		id := "props"
		result.IdentifiedBy = &id
		result.PropsModel = LlamaCppModelFromProps(propsVal)
		// THE SERVER CONFIG GRAB: the launch limits ride the probe.
		result.PropsLimits = LlamaCppLimitsFromProps(propsVal)
	} else if propsVal != nil {
		propsErr = "/props answered but is not llama.cpp"
	}

	// 1b) Secondary identification: props-less builds still fingerprint
	// on /v1/models. The payload fetched here is reused for the model
	// list in step 3.
	var modelsPayload any
	var modelsHeaders []string
	if !result.Found {
		code, raw, hdrs, err := probeGetJSON(api+"/models", apiKey, timeoutS)
		if err == nil && code == 200 {
			if v, perr := storage.DecodeJSONValue([]byte(raw)); perr == nil {
				modelsPayload = v
				modelsHeaders = hdrs
			}
		}
		if modelsPayload != nil {
			if fp := modelsFingerprint(modelsPayload, modelsHeaders); fp != "" {
				result.Found = true
				id := "models:" + fp
				result.IdentifiedBy = &id
			}
		}
		if !result.Found && propsErr == "" {
			propsErr = "no llama.cpp /props at this address"
		}
	}

	if !result.Found {
		detail := propsErr
		if detail == "" {
			detail = "no llama.cpp server at this address"
		}
		result.Detail = detail
		return result
	}

	// 2) /health — ready vs still-loading. A non-2xx answer IS the
	// "loading" answer; an unreachable route leaves ready unknown.
	healthCode := -1
	healthRaw := ""
	hc, hraw, _, herr := probeGetJSON(root+"/health", apiKey, timeoutS)
	switch e := herr.(type) {
	case nil:
		healthCode, healthRaw = hc, hraw
	case *httpStatusError:
		healthCode, healthRaw = e.code, e.body
	default:
		healthCode = -1
	}
	if healthCode >= 0 {
		ready := false
		if healthCode == 200 {
			status := "ok"
			if v, err := storage.DecodeJSONValue([]byte(healthRaw)); err == nil {
				if om, ok := v.(*storage.OrderedMap); ok {
					if sv, p := om.Get("status"); p {
						status = helpers.PyStr(sv)
					}
				}
			}
			ready = strings.ToLower(status) != "loading"
		}
		result.Ready = &ready
	}

	// 3) /v1/models — the model list (a missing route is tolerated).
	if modelsPayload != nil {
		result.Models = modelIDsFromModelsPayload(modelsPayload)
	} else {
		if names, err := OpenAIListModels(api, apiKey, float64(int(timeoutS))); err == nil {
			result.Models = names
		} else {
			result.Models = []string{}
		}
	}

	// Best name: what the server ADVERTISES wins; /props is the
	// fallback for builds that hide the route.
	if len(result.Models) > 0 {
		m := result.Models[0]
		result.Model = &m
	} else {
		result.Model = result.PropsModel
	}

	state := "state unknown"
	if result.Ready != nil {
		if *result.Ready {
			state = "ready"
		} else {
			state = "still loading"
		}
	}
	detail := fmt.Sprintf("llama.cpp server at %s (%s)", root, state)
	if result.Model != nil && *result.Model != "" {
		detail += fmt.Sprintf(" · model '%s'", *result.Model)
	}
	if line := llamaCppLimitsLine(result.PropsLimits); line != "" {
		detail += " · " + line
	}
	result.Detail = detail
	return result
}

// probeGetJSON GETs url DIRECT (never a proxy) with the optional
// Bearer; non-2xx answers come back as *httpStatusError with the body
// readable (urllib's HTTPError), transport failures as the raw error
// (the URLError class).
func probeGetJSON(target, apiKey string,
	timeoutS float64) (int, string, []string, error) {
	req, err := http.NewRequest("GET", target, nil)
	if err != nil {
		return 0, "", nil, err
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := httpClient(false, true) // always direct
	if timeoutS > 0 {
		req = req.WithContext(timeoutCtx(req.Context(), timeoutS))
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, "", nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return 0, "", nil, &httpStatusError{
			code: resp.StatusCode, body: string(raw)}
	}
	return resp.StatusCode, string(raw), resp.Header.Values("Server"), nil
}

// ---------------------------------------------------------------------------
// detect_llamacpp
// ---------------------------------------------------------------------------

// llamaCppProcessPortsFn is the detect seam (the tests pin it — the
// host machine can never leak a real llama-server into scan tests).
var llamaCppProcessPortsFn = LlamaCppProcessPorts

// DetectLlamaCpp finds a llama.cpp server on this machine: the
// LISTENING ports of every running llama-server/llamafile process
// FIRST (whatever --port it was given), then the common ports. Returns
// the first probe that positively identifies (via names which
// candidate list found it), nil when nothing matches. Dead local
// ports refuse instantly; live non-llama.cpp servers are skipped,
// never misreported.
func DetectLlamaCpp(apiKey string, ports []int, host string,
	timeoutS float64, includeProcessPorts bool) *ProbeResult {
	if host == "" {
		host = "127.0.0.1"
	}
	procPorts := []int{}
	if includeProcessPorts {
		procPorts = llamaCppProcessPortsFn()
	}
	scanPorts := LLamaCppScanPorts
	if ports != nil {
		scanPorts = ports
	}
	candidates := []int{}
	for _, port := range append(append([]int{}, procPorts...), scanPorts...) {
		if !containsInt(candidates, port) {
			candidates = append(candidates, port)
		}
	}
	for _, port := range candidates {
		probe := ProbeLlamaCpp(
			fmt.Sprintf("http://%s:%d", host, port), apiKey, timeoutS)
		if probe.Found {
			if containsInt(procPorts, port) {
				probe.Via = "process"
			} else {
				probe.Via = "scan"
			}
			return probe
		}
	}
	return nil
}

func containsInt(haystack []int, needle int) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// ResolveLlamaCppModel is "its model detected automatically": the
// configured llamacpp_model when set (a deliberate choice always
// wins), else the first advertised /v1/models entry, else the /props
// alias/basename, else nil.
func ResolveLlamaCppModel(config any, models []string,
	propsModel *string) *string {
	configured := ""
	if config != nil {
		if v, p := cfgGet(config, "llamacpp_model"); p {
			configured = strings.TrimSpace(helpers.PyStr(pyOr(v, "")))
		}
	}
	if configured != "" {
		return &configured
	}
	if len(models) > 0 {
		m := models[0]
		return &m
	}
	return propsModel
}

// ---------------------------------------------------------------------------
// llamacpp_autodetect_decision — the startup switch policy (pure)
// ---------------------------------------------------------------------------

// AutodetectDecision is the policy result: Switch (apply it), URL (the
// /v1 URL to persist — set only when it differs), Model (the detected
// name — nil = nothing to write, "" possible on the switch branch when
// nothing was detected), Limits (the /props grab), Message + Level.
type AutodetectDecision struct {
	Switch  bool            `json:"switch"`
	URL     *string         `json:"url"`
	Model   *string         `json:"model"`
	Limits  *LlamaCppLimits `json:"limits"`
	Message string          `json:"message"`
	Level   string          `json:"level"`
}

// LlamaCppAutodetectDecision — never override a provider choice that
// WORKS: provider already llamacpp → refresh URL/model (no switch);
// provider ollama + Ollama down → SWITCH; provider cloud + no API key
// → SWITCH; anything else → no switch, one hint line only.
func LlamaCppAutodetectDecision(config any, probe *ProbeResult,
	ollamaUp *bool) AutodetectDecision {
	if probe == nil || !probe.Found {
		return AutodetectDecision{Level: "info"}
	}
	root := strings.TrimRight(probe.BaseURL, "")
	if root == "" {
		root = LLamaCppDefaultBase
	}
	root = strings.TrimRight(root, "/")
	apiURL := root
	if !strings.HasSuffix(root, "/v1") {
		apiURL = root + "/v1"
	}
	model := probe.Model
	limits := probe.PropsLimits
	limitsNote := llamaCppLimitsLine(limits)

	provider := strings.ToLower(helpers.PyStr(pyOr(cfgValue(config,
		"llm_provider"), "ollama")))
	if provider == "" {
		provider = "ollama"
	}
	configuredURL := strings.TrimSpace(helpers.PyStr(
		pyOr(cfgValue(config, "llamacpp_api_url"), "")))
	configuredModel := strings.TrimSpace(helpers.PyStr(
		pyOr(cfgValue(config, "llamacpp_model"), "")))
	urlChanged := configuredURL != apiURL
	modelChanged := model != nil && *model != configuredModel

	where := "🦙 llama.cpp detected at " + root
	if model != nil && *model != "" {
		where += fmt.Sprintf(" · model '%s'", *model)
	}
	if limitsNote != "" {
		where += " · " + limitsNote
	}

	if provider == "llamacpp" {
		note := where
		if urlChanged && configuredURL != "" {
			note += fmt.Sprintf(" (was %s)", configuredURL)
		}
		d := AutodetectDecision{
			Switch: false, Limits: limits, Message: note, Level: "success",
		}
		if urlChanged {
			u := apiURL
			d.URL = &u
		}
		if modelChanged {
			d.Model = model
		}
		return d
	}

	switchNow := false
	why := ""
	if provider == "ollama" && ollamaUp != nil && !*ollamaUp {
		switchNow = true
		why = " — Ollama is not running, switching the LLM provider to llama.cpp"
	} else if provider == "cloud" {
		key := strings.TrimSpace(helpers.PyStr(
			pyOr(cfgValue(config, "cloud_api_key"), "")))
		if key == "" {
			switchNow = true
			why = " — no cloud API key is configured, switching the LLM " +
				"provider to llama.cpp"
		}
	}

	if switchNow {
		message := "🦙 llama.cpp caught automatically at " + root
		if model != nil && *model != "" {
			message += fmt.Sprintf(" · model '%s'", *model)
		}
		if limitsNote != "" {
			message += " · " + limitsNote
		}
		message += why + "."
		d := AutodetectDecision{
			Switch: true, Limits: limits, Message: message, Level: "success",
		}
		if urlChanged {
			u := apiURL
			d.URL = &u
		}
		m := ""
		if model != nil {
			m = *model
		}
		d.Model = &m
		return d
	}

	return AutodetectDecision{
		Switch: false, Limits: limits,
		Message: where + " — Settings → 🧠 LLM → 🦙 llama.cpp to use it.",
		Level:   "info",
	}
}

func cfgValue(config any, key string) any {
	v, _ := cfgGet(config, key)
	return v
}
