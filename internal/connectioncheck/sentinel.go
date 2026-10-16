package connectioncheck

// sentinel.go — net_guard.py (v0.85.0): THE SENTINEL. One background
// job, three probes, zero GUI:
//
//   - ProbeInternet — can the app reach the internet AT ALL? The
//     configured Worker URL first (the app's own service), then
//     api.github.com (the sync backbone), then cloudflare-dns.com
//     (generic transport). ANY HTTP answer — even a 401/404 — proves
//     the transport worked; only transport failures (DNS, timeout,
//     connection refused) count as offline.
//   - AIStatus — which engine answers RIGHT NOW: on (cloud) / on
//     (local: Ollama or llama.cpp) / warn (up but misconfigured) /
//     off (unreachable) / none (nothing configured).
//   - ProxyReachable — the configured SOCKS/HTTP proxy port answers a
//     TCP connect (the v22 health monitor's probe, off the GUI thread).
//
// READ-ONLY — GETs only, no payload, no write verb anywhere. UA:
// GitCurator-sentinel (never the pipeline's UA).

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"syscall"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// sentinelUA is the probe's own identity (kept short; never the
// pipeline's UA).
const sentinelUA = "GitCurator-sentinel"

// The transport-fallback ladder: the app's own Worker service first
// (it is the thing SYNC actually talks to), then the sync backbone,
// then a generic anycast endpoint. ANY answer = transport works.
// (vars, not consts: the reference's module attributes are
// monkeypatchable by its suite; the Go suite redirects the ladder at
// loopback fakes for the same hermeticity.)
var (
	githubProbeURL  = "https://api.github.com"
	genericProbeURL = "https://cloudflare-dns.com"
)

// v22's proxy defaults, preserved (the widget defaults say the same).
const (
	proxyDefaultHost = "127.0.0.1"
	proxyDefaultPort = 10808
)

// probeInternetFn / aiStatusFn / proxyReachableFn are the sentinel's
// own seams (the reference's suites monkeypatch the net_guard module
// attributes; the seams keep the same testability).
var (
	probeInternetFn  = ProbeInternet
	aiStatusFn       = AIStatus
	proxyReachableFn = ProxyReachable
)

// probeEndpoints is the ordered endpoint ladder for ProbeInternet —
// the configured Worker URL (when set) first, then the fixed
// fallbacks.
func probeEndpoints(config any) []string {
	eps := []string{}
	w := trimPySpace(pyStr(pyOr(cfgValue(config, "cloudflare_worker_url"), "")))
	if w != "" {
		eps = append(eps, w)
	}
	return append(eps, githubProbeURL, genericProbeURL)
}

// transportVerdict is _http_transport_ok's result: ok is true for ANY
// HTTP answer (all prove transport), false only for transport
// failures. latencyMs < 0 means "none" (Python's None).
type transportVerdict struct {
	ok        bool
	status    int
	latencyMs int
	note      string
}

// httpTransportOk GETs url; ok is true for ANY HTTP answer
// (2xx/3xx/4xx/5xx — all prove transport) and false only for
// transport failures. Never raises. Drains ≤2048 bytes to keep the
// socket sane. Loopback never rides the environment proxy (the
// v0.15.1 rule).
func httpTransportOk(u string, timeoutS float64) transportVerdict {
	t0 := time.Now()
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return transportVerdict{false, 0, -1,
			truncateRunes(errClassColon(err), 120)}
	}
	req.Header.Set("User-Agent", sentinelUA)
	client := clientForURL(u, timeoutS)
	res, err := client.Do(req)
	if err != nil {
		return transportVerdict{false, 0, -1,
			truncateRunes(errClassColon(err), 120)}
	}
	status := res.StatusCode
	_, _ = io.CopyN(io.Discard, res.Body, 2048)
	_ = res.Body.Close()
	return transportVerdict{true, status,
		int(time.Since(t0).Milliseconds()), ""}
}

// truncateRunes is Python's s[:n] (rune-wise).
func truncateRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	return string([]rune(s)[:n])
}

// ProbeInternetResult is probe_internet's plain dict shape.
type ProbeInternetResult struct {
	Online    bool   `json:"online"`
	Endpoint  string `json:"endpoint"`
	Status    int    `json:"status"`
	LatencyMs *int   `json:"latency_ms"`
	Detail    string `json:"detail"`
}

// ProbeInternet answers: is the internet reachable at all? The
// any-HTTP-answer law (see the package doc). Never raises.
func ProbeInternet(config any, timeoutS float64) ProbeInternetResult {
	out := ProbeInternetResult{Online: false, Status: 0}
	failures := []string{}
	for _, ep := range probeEndpoints(config) {
		v := httpTransportOk(ep, timeoutS)
		if v.ok {
			out.Online = true
			out.Endpoint = ep
			out.Status = v.status
			latency := v.latencyMs
			out.LatencyMs = &latency
			detail := fmt.Sprintf("%s answered HTTP %d", ep, v.status)
			if v.latencyMs >= 0 {
				detail += fmt.Sprintf(" in %d ms", v.latencyMs)
			}
			out.Detail = detail
			return out
		}
		note := v.note
		if note == "" {
			note = "no answer"
		}
		failures = append(failures, ep+": "+note)
	}
	out.Detail = truncateRunes(
		"no endpoint answered — "+strings.Join(failures, "; "), 300)
	return out
}

// AIStatusResult is ai_status's plain dict shape.
type AIStatusResult struct {
	State  string `json:"state"`
	Kind   string `json:"kind"`
	Engine string `json:"engine"`
	Detail string `json:"detail"`
	Short  string `json:"short"`
}

func aiRow(state, kind, engine, detail, short string) AIStatusResult {
	return AIStatusResult{state, kind, engine,
		pyStr(pyOr(detail, "")), pyStr(pyOr(short, ""))}
}

func aiOllama(config any, timeoutS float64) AIStatusResult {
	cfg := pyOr(cfgValue(config, "ollama"), map[string]any{})
	base := trimPySpace(pyStr(pyOr(dictDefault(cfg, "base_url", ""),
		"http://127.0.0.1:11434")))
	base = orDefault(strings.TrimRight(base, "/"),
		"http://127.0.0.1:11434")
	u := base + "/api/tags"
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return aiRow("off", "local", "Ollama",
			"Ollama not reachable at "+base+" ("+errClass(err)+") — "+
				"start the Ollama app", "off — unreachable")
	}
	req.Header.Set("User-Agent", sentinelUA)
	client := clientForURL(u, timeoutS)
	res, err := client.Do(req)
	if err != nil {
		return aiRow("off", "local", "Ollama",
			"Ollama not reachable at "+base+" ("+errClass(err)+") — "+
				"start the Ollama app", "off — unreachable")
	}
	raw, rerr := io.ReadAll(res.Body)
	_ = res.Body.Close()
	var body any
	if rerr == nil && utf8.Valid(raw) {
		s := string(raw)
		if s == "" {
			s = "{}"
		}
		if v, perr := storage.DecodeJSONValue([]byte(s)); perr == nil {
			body = v
		}
	}
	if body == nil {
		return aiRow("off", "local", "Ollama",
			"Ollama not reachable at "+base+" (ValueError) — "+
				"start the Ollama app", "off — unreachable")
	}
	var models []string
	if list, ok := dictValue(body, "models").([]any); ok {
		for _, m := range list {
			name := trimPySpace(pyStr(pyOr(dictValue(m, "name"),
				dictValue(m, "model"))))
			if name != "" {
				models = append(models, name)
			}
		}
	}
	model := trimPySpace(pyStr(pyOr(dictValue(cfg, "model"), "")))
	if len(models) == 0 {
		return aiRow("warn", "local", "Ollama",
			"Ollama up @ "+base+" but NO models installed — "+
				"run `ollama pull <model>` first", "warn — no models")
	}
	if model != "" && !containsStr(models, model) {
		return aiRow("warn", "local", "Ollama",
			"Ollama up @ "+base+" · "+itoa(len(models))+
				" model(s) — configured '"+model+"' is NOT pulled",
			"warn — '"+model+"' not pulled")
	}
	detail := "Ollama up @ " + base + " · " + itoa(len(models)) + " model(s)"
	if model != "" {
		return aiRow("on", "local", "Ollama",
			detail+" · '"+model+"' ready", "on")
	}
	return aiRow("on", "local", "Ollama",
		detail+" · no model configured (picked at run time)", "on")
}

func aiLlamaCpp(config any, timeoutS float64) AIStatusResult {
	cfg := config
	u := trimPySpace(pyStr(pyOr(cfgValue(cfg, "llamacpp_api_url"), "")))
	if u == "" {
		u = llmclient.LLamaCppDefaultBase
	}
	key := pyStr(pyOr(cfgValue(cfg, "llamacpp_api_key"), ""))
	var probe *llmclient.ProbeResult
	var probePanic any
	func() {
		defer func() { probePanic = recover() }()
		probe = probeLlamaCppFn(u, key, timeoutS)
	}()
	if probePanic != nil {
		probe = &llmclient.ProbeResult{Detail: panicClass(probePanic) + ": " +
			panicAsError(probePanic).Error()}
	}
	if probe == nil {
		probe = &llmclient.ProbeResult{}
	}
	if !probe.Found {
		det := probe.Detail
		if det == "" {
			det = "no answer"
		}
		return aiRow("off", "local", "llama.cpp",
			"llama-server not detected at "+u+" ("+det+") — "+
				"start it with: llama-server -m <model>.gguf",
			"off — unreachable")
	}
	base := probe.BaseURL
	if base == "" {
		base = u
	}
	var model *string
	model = llmclient.ResolveLlamaCppModel(cfg, probe.Models,
		probe.PropsModel)
	loading := ""
	if probe.Ready != nil && !*probe.Ready {
		loading = " · still loading"
	}
	if model == nil {
		return aiRow("warn", "local", "llama.cpp",
			"llama-server up @ "+base+" but no model loaded — "+
				"start it with -m <model>.gguf", "warn — no model")
	}
	return aiRow("on", "local", "llama.cpp",
		"llama-server up @ "+base+" · model '"+*model+"'"+loading, "on")
}

func aiCloud(config any, timeoutS float64) AIStatusResult {
	cfg := config
	u := trimPySpace(pyStr(pyOr(cfgValue(cfg, "cloud_api_url"), "")))
	key := pyStr(pyOr(cfgValue(cfg, "cloud_api_key"), ""))
	model := trimPySpace(pyStr(pyOr(cfgValue(cfg, "cloud_model"), "")))
	if u == "" {
		return aiRow("none", "cloud", "cloud",
			"no cloud API URL configured — Settings ▸ LLM",
			"not configured")
	}
	flavor := "OpenAI-compatible"
	if llmclient.IsAnthropicURL(u) {
		flavor = "Claude"
	}
	var (
		endpointOK  bool
		message     string
		modelListed *bool
	)
	if err := func() (err error) {
		defer func() {
			if r := recover(); r != nil {
				err = panicAsError(r)
			}
		}()
		endpointOK, message, modelListed = preflightCloudFn(u, key, model, timeoutS)
		return nil
	}(); err != nil {
		return aiRow("off", "cloud", "cloud",
			u+" — check failed ("+errClassColon(err)+")",
			"off — check failed")
	}
	if !endpointOK {
		return aiRow("off", "cloud", "cloud",
			u+" ("+flavor+") — "+message, "off — endpoint rejected")
	}
	if modelListed != nil && !*modelListed {
		return aiRow("warn", "cloud", "cloud",
			u+" ("+flavor+") — "+message+" · model '"+model+
				"' NOT in the list (trying anyway)",
			"warn — '"+model+"' not listed")
	}
	note := u + " (" + flavor + ") — " + message
	if model != "" {
		note += " · '" + model + "'"
	}
	return aiRow("on", "cloud", "cloud", note, "on")
}

// AIStatus is the ACTIVE LLM provider's live verdict. Never raises.
func AIStatus(config any, timeoutS float64) (out AIStatusResult) {
	provider := strings.ToLower(trimPySpace(
		pyStr(pyOr(cfgValue(config, "llm_provider"), "ollama"))))
	if provider == "" {
		provider = "ollama"
	}
	defer func() {
		if r := recover(); r != nil {
			kind := "local"
			if provider == "cloud" {
				kind = "cloud"
			}
			engine := provider
			if engine == "" {
				engine = "ollama"
			}
			out = aiRow("off", kind, engine,
				"probe failed ("+panicClass(r)+": "+
					panicAsError(r).Error()+")",
				"off — probe failed")
		}
	}()
	switch provider {
	case "cloud":
		return aiCloud(config, timeoutS)
	case "llamacpp":
		return aiLlamaCpp(config, timeoutS)
	default:
		return aiOllama(config, timeoutS)
	}
}

// ProxyResult is proxy_reachable's plain dict shape.
type ProxyResult struct {
	State  string `json:"state"`
	Host   string `json:"host"`
	Port   int    `json:"port"`
	Detail string `json:"detail"`
}

// ProxyReachable is the configured proxy's TCP verdict — the exact
// probe the v22 health monitor ran on the GUI thread, now safe to
// call from any thread. Never raises. The unreachable detail carries
// the real syscall errno ("— errno {n}").
func ProxyReachable(config any, timeoutS float64) ProxyResult {
	proxy := pyOr(cfgValue(config, "proxy"), map[string]any{})
	if !pyTruthy(dictValue(proxy, "enabled")) {
		return ProxyResult{"disabled", "", 0, "Proxy disabled"}
	}
	host := trimPySpace(pyStr(pyOr(dictValue(proxy, "host"), "")))
	if host == "" {
		host = proxyDefaultHost
	}
	port := pyIntDefault(dictValue(proxy, "port"), proxyDefaultPort)
	timeout := secondsDur(maxFloat(0.5, timeoutS))
	conn, err := (&net.Dialer{Timeout: timeout}).
		Dial("tcp4", net.JoinHostPort(host, strconv.Itoa(port)))
	if err == nil {
		_ = conn.Close()
		return ProxyResult{"connected", host, port,
			"Proxy OK (" + host + ":" + strconv.Itoa(port) + ")"}
	}
	var opErr *net.OpError
	if errors.As(err, &opErr) {
		var errno syscall.Errno
		if errors.As(opErr.Err, &errno) {
			return ProxyResult{"unreachable", host, port,
				"Proxy unreachable (" + host + ":" + strconv.Itoa(port) +
					") — errno " + strconv.Itoa(int(errno))}
		}
	}
	return ProxyResult{"unreachable", host, port,
		"Proxy check failed (" + errClass(err) + ")"}
}

// pyIntDefault mirrors int(value) with the (TypeError, ValueError)
// fallback: bool → 1/0, float truncates toward zero, decimal strings
// parse, anything else → def.
func pyIntDefault(v any, def int) int {
	switch t := v.(type) {
	case nil:
		return def
	case bool:
		if t {
			return 1
		}
		return 0
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil {
			return n
		}
		return def
	default:
		return def
	}
}

func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

func orDefault(s, def string) string {
	if s != "" {
		return s
	}
	return def
}

// Snapshot is sentinel_snapshot's plain dict shape: {'ts', 'net',
// 'ai', 'proxy'}.
type Snapshot struct {
	Ts    float64             `json:"ts"`
	Net   ProbeInternetResult `json:"net"`
	AI    AIStatusResult      `json:"ai"`
	Proxy ProxyResult         `json:"proxy"`
}

// EmptySnapshot is the all-unknown snapshot (used when the job itself
// fails).
func EmptySnapshot() Snapshot {
	return Snapshot{
		Ts: pyNow(),
		Net: ProbeInternetResult{
			Online: false, Endpoint: "", Status: 0,
			Detail: "sentinel did not run",
		},
		AI: AIStatusResult{
			State: "off", Kind: "local", Engine: "",
			Detail: "sentinel did not run", Short: "",
		},
		Proxy: ProxyResult{State: "disabled", Host: "", Port: 0, Detail: ""},
	}
}

func pyNow() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// SentinelSnapshot is THE ONE JOB the GUI's 60s timer runs: all three
// probes, one snapshot, never raises — a failing probe degrades to
// its own empty row, not a crash.
func SentinelSnapshot(config any, timeoutS float64) Snapshot {
	snap := EmptySnapshot()
	snap.Net = guardedNet(func() ProbeInternetResult {
		return probeInternetFn(config, timeoutS)
	}, snap.Net)
	snap.AI = guardedAI(func() AIStatusResult {
		return aiStatusFn(config, timeoutS)
	}, snap.AI)
	snap.Proxy = guardedProxy(func() ProxyResult {
		return proxyReachableFn(config, 2.0)
	}, snap.Proxy)
	return snap
}

// guardedNet/guardedAI/guardedProxy run one probe with the never-crash
// law: a panicked probe yields the fallback row carrying
// "probe failed ({class}: {msg})" (net_guard's _guarded).
func guardedNet(fn func() ProbeInternetResult,
	fallback ProbeInternetResult) (out ProbeInternetResult) {
	out = fallback
	defer func() {
		if r := recover(); r != nil {
			out = fallback
			out.Detail = "probe failed (" + panicClass(r) + ": " +
				panicAsError(r).Error() + ")"
		}
	}()
	return fn()
}

func guardedAI(fn func() AIStatusResult,
	fallback AIStatusResult) (out AIStatusResult) {
	out = fallback
	defer func() {
		if r := recover(); r != nil {
			out = fallback
			out.Detail = "probe failed (" + panicClass(r) + ": " +
				panicAsError(r).Error() + ")"
		}
	}()
	return fn()
}

func guardedProxy(fn func() ProxyResult,
	fallback ProxyResult) (out ProxyResult) {
	out = fallback
	defer func() {
		if r := recover(); r != nil {
			out = fallback
			out.Detail = "probe failed (" + panicClass(r) + ": " +
				panicAsError(r).Error() + ")"
		}
	}()
	return fn()
}

// OfflineVerdict is the warn-once law as PURE logic:
//
//   - previous == nil (first observation) → "warn-now" when offline —
//     THE offline-launch modal, exactly once per session — or
//     "first-online".
//   - online → offline later → "fell-offline" (a log line; the red
//     cue carries the rest — never a modal mid-session).
//   - offline → online → "back-online" (the cue turns green).
//   - unchanged → "online" / "offline".
func OfflineVerdict(previous *bool, online bool) string {
	if previous == nil {
		if !online {
			return "warn-now"
		}
		return "first-online"
	}
	if *previous && !online {
		return "fell-offline"
	}
	if !*previous && online {
		return "back-online"
	}
	if online {
		return "online"
	}
	return "offline"
}

// SentinelLine is one human line for logs/CLI:
// "Net: online ({host}) · AI: on — local (Ollama) · Proxy: connected".
// Never raises.
func SentinelLine(snap Snapshot) (line string) {
	defer func() {
		// a log line never crashes (the reference's except → the
		// all-unknown line)
		if recover() != nil {
			line = "Net: unknown · AI: unknown · Proxy: unknown"
		}
	}()
	netTxt := "offline"
	if snap.Net.Online {
		netTxt = "online"
	}
	ep := snap.Net.Endpoint
	if ep != "" {
		// ep.split("//", 1)[-1][:32] — everything after the first
		// "//" (the whole string when absent).
		if idx := strings.Index(ep, "//"); idx >= 0 {
			ep = ep[idx+2:]
		}
		ep = truncateRunes(ep, 32)
		netTxt += " (" + ep + ")"
	}
	aiTxt := "unknown"
	switch snap.AI.State {
	case "on":
		if snap.AI.Kind == "cloud" {
			aiTxt = "on — cloud"
		} else {
			engine := snap.AI.Engine
			if engine == "" {
				engine = "local"
			}
			aiTxt = "on — local (" + engine + ")"
		}
	case "none":
		aiTxt = "not configured"
	case "":
		aiTxt = "unknown"
	default:
		aiTxt = orDefault(snap.AI.Short, snap.AI.State)
	}
	proxyTxt := "unknown"
	switch snap.Proxy.State {
	case "connected":
		proxyTxt = "connected"
	case "unreachable":
		proxyTxt = "unreachable"
	case "disabled":
		proxyTxt = "off"
	}
	return "Net: " + netTxt + " · AI: " + aiTxt + " · Proxy: " + proxyTxt
}
