package connectioncheck

// sentinel_test.go — THE SENTINEL selections: probe_internet (the
// any-HTTP-answer law, the endpoint ladder, the offline verdict, the
// GET-only read-only law), ai_status (every provider verdict),
// proxy_reachable, sentinel_snapshot, offline_verdict's warn-once law,
// sentinel_line, and THE SENTINEL UA. Ported from the reference's
// test_netguard.py selections; the Python suite mocks urllib/sockets —
// the Go suite redirects the ladder at loopback fakes and swaps the
// llmclient seams (same hermeticity, no network beyond 127.0.0.1).

import (
	"net"
	"net/http"
	"strconv"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/llmclient"
)

// ---------------------------------------------------------------------------
// 1) ProbeInternet
// ---------------------------------------------------------------------------

func TestProbeInternetOnlineViaWorkerURL(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/health": route(200, "ok"),
	})
	out := ProbeInternet(map[string]any{
		"cloudflare_worker_url": srv.root() + "/health"}, 5.0)
	if !out.Online {
		t.Fatalf("offline: %+v", out)
	}
	if out.Endpoint != srv.root()+"/health" {
		t.Errorf("endpoint = %q", out.Endpoint)
	}
	if out.Status != 200 {
		t.Errorf("status = %d", out.Status)
	}
	if out.LatencyMs == nil {
		t.Error("latency_ms is nil")
	}
	if !strings.Contains(out.Detail, "answered HTTP 200") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestProbeInternetHTTPErrorStillOnline(t *testing.T) {
	// The law: a 401 proves DNS+TCP+TLS worked — transport is up.
	srv := newJSONServer(t, map[string][2]any{
		"/": route(401, map[string]any{"message": "no"}),
	})
	out := ProbeInternet(map[string]any{
		"cloudflare_worker_url": srv.root() + "/"}, 5.0)
	if !out.Online {
		t.Fatalf("offline: %+v", out)
	}
	if out.Status != 401 {
		t.Errorf("status = %d", out.Status)
	}
}

func TestProbeInternetWorkerDeadGitHubAnswers(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/": route(200, map[string]any{}),
	})
	withProbes(t, srv.root()+"/", deadURL(t))
	out := ProbeInternet(map[string]any{
		"cloudflare_worker_url": deadURL(t) + "/"}, 5.0)
	if !out.Online {
		t.Fatalf("offline: %+v", out)
	}
	if out.Endpoint != srv.root()+"/" {
		t.Errorf("endpoint = %q (the github fallback answered)",
			out.Endpoint)
	}
}

func TestProbeInternetGenericFallbackAnswersLast(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/": route(503, map[string]any{}),
	})
	withProbes(t, deadURL(t), srv.root())
	out := ProbeInternet(map[string]any{}, 5.0)
	if !out.Online {
		t.Fatalf("offline: %+v", out)
	}
	if out.Endpoint != srv.root() {
		t.Errorf("endpoint = %q (the generic fallback answered)",
			out.Endpoint)
	}
	if out.Status != 503 {
		t.Errorf("status = %d", out.Status)
	}
}

func TestProbeInternetAllDeadIsOffline(t *testing.T) {
	withProbes(t, deadURL(t), deadURL(t))
	out := ProbeInternet(map[string]any{}, 2.0)
	if out.Online {
		t.Fatal("online with everything dead")
	}
	if !strings.Contains(out.Detail, "no endpoint answered") {
		t.Errorf("detail: %q", out.Detail)
	}
	if out.LatencyMs != nil {
		t.Error("latency present while offline")
	}
}

func TestProbeInternetLadderOrderWorkerFirst(t *testing.T) {
	// the ladder: the configured Worker URL probes FIRST — when it
	// answers, the fallbacks are never touched
	worker := newJSONServer(t, map[string][2]any{
		"/health": route(200, map[string]any{}),
	})
	github := newJSONServer(t, map[string][2]any{
		"/": route(200, map[string]any{}),
	})
	withProbes(t, github.root()+"/", deadURL(t))
	out := ProbeInternet(map[string]any{
		"cloudflare_worker_url": worker.root() + "/health"}, 5.0)
	if !out.Online || out.Endpoint != worker.root()+"/health" {
		t.Fatalf("worker did not win the ladder: %+v", out)
	}
	if len(github.captured()) != 0 {
		t.Error("github fallback was probed although the worker answered")
	}
}

func TestProbeInternetLoopbackNeverRidesTheProxy(t *testing.T) {
	// the v0.15.1 rule: loopback targets get the direct client (no
	// proxy hook); everything else the environment-proxy client
	direct := clientForURL("http://127.0.0.1:8787/health", 1.0)
	dtr, ok := direct.Transport.(*http.Transport)
	if !ok || dtr.Proxy != nil {
		t.Error("loopback URL did not route through the direct client")
	}
	plain := clientForURL("https://api.github.com", 1.0)
	ptr, ok := plain.Transport.(*http.Transport)
	if !ok || ptr.Proxy == nil {
		t.Error("non-loopback URL lost the environment proxy")
	}
}

func TestProbeInternetGetOnlyReadOnlyLaw(t *testing.T) {
	// the radar's read-only law, applied to the sentinel: every probe
	// is a bare GET with no payload
	srv := newJSONServer(t, map[string][2]any{
		"/": route(200, map[string]any{}),
	})
	withProbes(t, srv.root()+"/", deadURL(t))
	_ = ProbeInternet(map[string]any{}, 5.0)
	captured := srv.captured()
	if len(captured) != 1 {
		t.Fatalf("requests = %d, want 1", len(captured))
	}
	if captured[0].Method != "GET" {
		t.Errorf("method = %s, want GET", captured[0].Method)
	}
	if captured[0].BodyN != 0 {
		t.Errorf("GET carried a body: %d bytes", captured[0].BodyN)
	}
}

func TestProbeInternetGarbageConfigNeverPanics(t *testing.T) {
	withProbes(t, deadURL(t), deadURL(t))
	out := ProbeInternet(map[string]any{
		"cloudflare_worker_url": "not a url at all"}, 2.0)
	if out.Online {
		t.Error("garbage URL counted as online")
	}
}

func TestProbeInternetLatencyPresentWhenOnline(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/": route(200, map[string]any{}),
	})
	withProbes(t, srv.root()+"/", deadURL(t))
	out := ProbeInternet(map[string]any{}, 5.0)
	if !out.Online {
		t.Fatal("offline")
	}
	if out.LatencyMs == nil || *out.LatencyMs < 0 {
		t.Errorf("latency = %+v", out.LatencyMs)
	}
	if !strings.Contains(out.Detail, " in ") ||
		!strings.Contains(out.Detail, " ms") {
		t.Errorf("detail lacks the latency: %q", out.Detail)
	}
}

// THE SENTINEL UA: the probes speak "GitCurator-sentinel" — never the
// pipeline's UA (pinned from the wire).
func TestProbeInternetSentinelUA(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/": route(200, map[string]any{}),
	})
	withProbes(t, srv.root()+"/", deadURL(t))
	_ = ProbeInternet(map[string]any{}, 5.0)
	captured := srv.captured()
	if len(captured) != 1 {
		t.Fatalf("requests = %d", len(captured))
	}
	if captured[0].UA != "GitCurator-sentinel" {
		t.Errorf("UA = %q, want GitCurator-sentinel", captured[0].UA)
	}
}

// ---------------------------------------------------------------------------
// 2) AIStatus — Ollama
// ---------------------------------------------------------------------------

func TestAIStatusOllamaOnWithConfiguredModel(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "qwen3:latest"},
			map[string]any{"name": "llama3:8b"},
		}}),
	})
	out := AIStatus(map[string]any{
		"llm_provider": "ollama",
		"ollama": map[string]any{
			"base_url": srv.root(), "model": "qwen3:latest"}}, 6.0)
	if out.State != "on" || out.Kind != "local" || out.Engine != "Ollama" {
		t.Fatalf("row = %+v", out)
	}
	if !strings.Contains(out.Detail, "qwen3:latest") {
		t.Errorf("detail: %q", out.Detail)
	}
	if out.Short != "on" {
		t.Errorf("short: %q", out.Short)
	}
}

func TestAIStatusOllamaWarnWhenModelNotPulled(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "qwen3:latest"},
		}}),
	})
	out := AIStatus(map[string]any{
		"llm_provider": "ollama",
		"ollama": map[string]any{
			"base_url": srv.root(), "model": "missing:latest"}}, 6.0)
	if out.State != "warn" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "NOT pulled") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusOllamaWarnWhenNoModels(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{}}),
	})
	out := AIStatus(map[string]any{
		"llm_provider": "ollama",
		"ollama":       map[string]any{"base_url": srv.root()}}, 6.0)
	if out.State != "warn" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "NO models") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusOllamaOffWhenUnreachable(t *testing.T) {
	out := AIStatus(map[string]any{
		"llm_provider": "ollama",
		"ollama":       map[string]any{"base_url": deadURL(t)}}, 2.0)
	if out.State != "off" || out.Kind != "local" {
		t.Fatalf("row = %+v", out)
	}
	if !strings.Contains(out.Detail, "start the Ollama app") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusDefaultProviderIsOllama(t *testing.T) {
	// No llm_provider key at all → the ollama path (the app default)
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "m"},
		}}),
	})
	out := AIStatus(map[string]any{
		"ollama": map[string]any{"base_url": srv.root(), "model": "m"}}, 6.0)
	if out.State != "on" || out.Engine != "Ollama" {
		t.Fatalf("row = %+v", out)
	}
}

func TestAIStatusOllamaSentinelUA(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "m"},
		}}),
	})
	_ = AIStatus(map[string]any{
		"ollama": map[string]any{"base_url": srv.root()}}, 6.0)
	captured := srv.captured()
	if len(captured) != 1 {
		t.Fatalf("requests = %d", len(captured))
	}
	if captured[0].UA != "GitCurator-sentinel" {
		t.Errorf("UA = %q, want GitCurator-sentinel", captured[0].UA)
	}
	if captured[0].Method != "GET" || captured[0].BodyN != 0 {
		t.Errorf("read-only law broken: %+v", captured[0])
	}
}

// ---------------------------------------------------------------------------
// 2) AIStatus — llama.cpp
// ---------------------------------------------------------------------------

func fakeProbe(found bool, detail string) *llmclient.ProbeResult {
	return &llmclient.ProbeResult{
		Found: found, Models: []string{}, Detail: detail,
	}
}

func TestAIStatusLlamaCppOnWithModel(t *testing.T) {
	ready := true
	model := "qwen3.gguf"
	withSeam(t, &probeLlamaCppFn,
		func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
			return &llmclient.ProbeResult{
				Found: true, BaseURL: "http://127.0.0.1:8080",
				Models: []string{"qwen3.gguf"}, PropsModel: &model,
				Ready: &ready, Detail: "llama.cpp @ 8080"}
		})
	out := AIStatus(map[string]any{
		"llm_provider": "llamacpp", "llamacpp_model": "qwen3.gguf"}, 6.0)
	if out.State != "on" || out.Kind != "local" || out.Engine != "llama.cpp" {
		t.Fatalf("row = %+v", out)
	}
	if !strings.Contains(out.Detail, "qwen3.gguf") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusLlamaCppWarnWhenNoModelLoaded(t *testing.T) {
	ready := false
	withSeam(t, &probeLlamaCppFn,
		func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
			return &llmclient.ProbeResult{
				Found: true, BaseURL: "http://127.0.0.1:8080",
				Models: []string{}, Ready: &ready, Detail: "loading"}
		})
	out := AIStatus(map[string]any{"llm_provider": "llamacpp"}, 6.0)
	if out.State != "warn" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "no model") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusLlamaCppOffWhenNotFound(t *testing.T) {
	withSeam(t, &probeLlamaCppFn,
		func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
			return fakeProbe(false, "no answer")
		})
	out := AIStatus(map[string]any{"llm_provider": "llamacpp"}, 6.0)
	if out.State != "off" || out.Engine != "llama.cpp" {
		t.Fatalf("row = %+v", out)
	}
	if !strings.Contains(out.Detail, "llama-server -m") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusLlamaCppStillLoadingNoted(t *testing.T) {
	ready := false
	model := "m.gguf"
	withSeam(t, &probeLlamaCppFn,
		func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
			return &llmclient.ProbeResult{
				Found: true, BaseURL: "http://127.0.0.1:8080",
				Models: []string{"m.gguf"}, PropsModel: &model,
				Ready: &ready}
		})
	out := AIStatus(map[string]any{"llm_provider": "llamacpp"}, 6.0)
	if out.State != "on" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "still loading") {
		t.Errorf("detail: %q", out.Detail)
	}
}

// ---------------------------------------------------------------------------
// 2) AIStatus — cloud
// ---------------------------------------------------------------------------

func TestAIStatusCloudOnWhenEndpointOKAndModelListed(t *testing.T) {
	listed := true
	withSeam(t, &preflightCloudFn,
		func(apiURL, apiKey, model string,
			timeoutS float64) (bool, string, *bool) {
			return true, "12 model(s) listed", &listed
		})
	out := AIStatus(map[string]any{
		"llm_provider":  "cloud",
		"cloud_api_url": "https://api.openai.com/v1",
		"cloud_api_key": "sk-x", "cloud_model": "gpt-4o-mini"}, 6.0)
	if out.State != "on" || out.Kind != "cloud" {
		t.Fatalf("row = %+v", out)
	}
	if !strings.Contains(out.Detail, "gpt-4o-mini") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusCloudWarnWhenModelNotListed(t *testing.T) {
	listed := false
	withSeam(t, &preflightCloudFn,
		func(apiURL, apiKey, model string,
			timeoutS float64) (bool, string, *bool) {
			return true, "12 model(s) listed", &listed
		})
	out := AIStatus(map[string]any{
		"llm_provider":  "cloud",
		"cloud_api_url": "https://api.openai.com/v1",
		"cloud_api_key": "sk-x", "cloud_model": "nope"}, 6.0)
	if out.State != "warn" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "NOT in the list") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestAIStatusCloudOffWhenEndpointRejected(t *testing.T) {
	withSeam(t, &preflightCloudFn,
		func(apiURL, apiKey, model string,
			timeoutS float64) (bool, string, *bool) {
			return false, "answered HTTP 401", nil
		})
	out := AIStatus(map[string]any{
		"llm_provider":  "cloud",
		"cloud_api_url": "https://api.openai.com/v1",
		"cloud_api_key": "bad", "cloud_model": "x"}, 6.0)
	if out.State != "off" || out.Kind != "cloud" {
		t.Fatalf("row = %+v", out)
	}
}

func TestAIStatusCloudNoneWhenNoURL(t *testing.T) {
	out := AIStatus(map[string]any{"llm_provider": "cloud"}, 6.0)
	if out.State != "none" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if out.Short != "not configured" {
		t.Errorf("short: %q", out.Short)
	}
}

func TestAIStatusCloudPreflightRaisingIsOffNotCrash(t *testing.T) {
	withSeam(t, &preflightCloudFn,
		func(apiURL, apiKey, model string,
			timeoutS float64) (bool, string, *bool) {
			panic("boom")
		})
	out := AIStatus(map[string]any{
		"llm_provider": "cloud", "cloud_api_url": "https://x.test"}, 6.0)
	if out.State != "off" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "check failed") {
		t.Errorf("detail: %q", out.Detail)
	}
}

// ---------------------------------------------------------------------------
// 3) ProxyReachable
// ---------------------------------------------------------------------------

func TestProxyReachableDisabled(t *testing.T) {
	out := ProxyReachable(map[string]any{
		"proxy": map[string]any{"enabled": false}}, 2.0)
	if out.State != "disabled" {
		t.Fatalf("state = %s", out.State)
	}
	if out.Detail != "Proxy disabled" {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestProxyReachableNoProxyKey(t *testing.T) {
	out := ProxyReachable(map[string]any{}, 2.0)
	if out.State != "disabled" {
		t.Fatalf("state = %s", out.State)
	}
}

func TestProxyReachableConnected(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	port := listener.Addr().(*net.TCPAddr).Port
	out := ProxyReachable(map[string]any{
		"proxy": map[string]any{"enabled": true, "host": "127.0.0.1",
			"port": port}}, 2.0)
	if out.State != "connected" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, strconv.Itoa(port)) {
		t.Errorf("detail: %q", out.Detail)
	}
	if !strings.Contains(out.Detail, "Proxy OK (127.0.0.1:"+strconv.Itoa(port)+")") {
		t.Errorf("detail: %q", out.Detail)
	}
}

func TestProxyReachableUnreachableErrnoDetail(t *testing.T) {
	// a dead loopback port → ECONNREFUSED (errno 111 on Linux — the
	// reference's connect_ex report)
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	_ = listener.Close()
	out := ProxyReachable(map[string]any{
		"proxy": map[string]any{"enabled": true, "host": "127.0.0.1",
			"port": port}}, 2.0)
	if out.State != "unreachable" {
		t.Fatalf("state = %s (%s)", out.State, out.Detail)
	}
	if !strings.Contains(out.Detail, "errno ") {
		t.Errorf("detail lacks the errno: %q", out.Detail)
	}
	if !strings.Contains(out.Detail, "errno 111") {
		t.Errorf("detail lacks ECONNREFUSED: %q", out.Detail)
	}
}

func TestProxyReachableGarbagePortDefaultsNeverPanic(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = listener.Close() })
	port := listener.Addr().(*net.TCPAddr).Port
	out := ProxyReachable(map[string]any{
		"proxy": map[string]any{"enabled": true, "host": "127.0.0.1",
			"port": "not-a-number"}}, 2.0)
	// the garbage port falls back to the v22 default 10808 — dead in
	// the sandbox → unreachable, never a crash
	if out.State == "disabled" {
		t.Fatalf("state = %s", out.State)
	}
	if out.Port != 10808 {
		t.Errorf("port = %d, want the 10808 default", out.Port)
	}
	_ = port
}

func TestProxyReachableHostDefaultsWhenEmpty(t *testing.T) {
	out := ProxyReachable(map[string]any{
		"proxy": map[string]any{"enabled": true, "host": "",
			"port": "not-a-number"}}, 2.0)
	if out.Host != "127.0.0.1" {
		t.Errorf("host = %q, want the 127.0.0.1 default", out.Host)
	}
	if out.Port != 10808 {
		t.Errorf("port = %d, want 10808", out.Port)
	}
}

func TestProxyReachableUnresolvableHostIsCheckFailed(t *testing.T) {
	// DNS failure is not an errno — the "Proxy check failed ({class})"
	// shape (Python: gaierror)
	out := ProxyReachable(map[string]any{
		"proxy": map[string]any{"enabled": true, "host": "no.such.host.invalid",
			"port": 1}}, 2.0)
	if out.State != "unreachable" {
		t.Fatalf("state = %s", out.State)
	}
	if !strings.Contains(out.Detail, "Proxy check failed (") {
		t.Errorf("detail: %q", out.Detail)
	}
}

// ---------------------------------------------------------------------------
// 4) SentinelSnapshot
// ---------------------------------------------------------------------------

func TestSentinelSnapshotShapeAndAggregation(t *testing.T) {
	withSeam(t, &probeInternetFn,
		func(config any, timeoutS float64) ProbeInternetResult {
			return ProbeInternetResult{Online: true}
		})
	withSeam(t, &aiStatusFn,
		func(config any, timeoutS float64) AIStatusResult {
			return AIStatusResult{State: "on"}
		})
	withSeam(t, &proxyReachableFn,
		func(config any, timeoutS float64) ProxyResult {
			return ProxyResult{State: "connected"}
		})
	snap := SentinelSnapshot(map[string]any{}, 6.0)
	if snap.Ts == 0 {
		t.Error("ts missing")
	}
	if !snap.Net.Online {
		t.Error("net not aggregated")
	}
	if snap.AI.State != "on" {
		t.Error("ai not aggregated")
	}
	if snap.Proxy.State != "connected" {
		t.Error("proxy not aggregated")
	}
}

func TestSentinelSnapshotNeverPanicsWhenProbesPanic(t *testing.T) {
	withSeam(t, &probeInternetFn,
		func(config any, timeoutS float64) ProbeInternetResult {
			panic("x")
		})
	withSeam(t, &aiStatusFn,
		func(config any, timeoutS float64) AIStatusResult { panic("y") })
	withSeam(t, &proxyReachableFn,
		func(config any, timeoutS float64) ProxyResult { panic("z") })
	snap := SentinelSnapshot(map[string]any{}, 6.0)
	if snap.Net.Online {
		t.Error("panicked probe reported online")
	}
	if !strings.Contains(snap.Net.Detail, "probe failed") {
		t.Errorf("net detail: %q", snap.Net.Detail)
	}
	if snap.AI.State != "off" {
		t.Errorf("ai state = %s", snap.AI.State)
	}
	if !strings.Contains(snap.AI.Detail, "probe failed") {
		t.Errorf("ai detail: %q", snap.AI.Detail)
	}
	if snap.Proxy.State != "disabled" {
		t.Errorf("proxy state = %s", snap.Proxy.State)
	}
	if !strings.Contains(snap.Proxy.Detail, "probe failed") {
		t.Errorf("proxy detail: %q", snap.Proxy.Detail)
	}
}

func TestEmptySnapshotShape(t *testing.T) {
	snap := EmptySnapshot()
	if snap.Net.Online {
		t.Error("empty snapshot online")
	}
	if !strings.Contains(snap.Net.Detail, "sentinel did not run") {
		t.Errorf("net detail: %q", snap.Net.Detail)
	}
	if snap.AI.State != "off" || snap.Proxy.State != "disabled" {
		t.Errorf("shape: %+v", snap)
	}
}

// ---------------------------------------------------------------------------
// 5) OfflineVerdict — the warn-once law
// ---------------------------------------------------------------------------

func TestOfflineVerdictFirstObservationOfflineWarnsNow(t *testing.T) {
	// THE owner's ask: opening offline → the warning, exactly once.
	if got := OfflineVerdict(nil, false); got != "warn-now" {
		t.Errorf("got %q, want warn-now", got)
	}
}

func TestOfflineVerdictFirstObservationOnline(t *testing.T) {
	if got := OfflineVerdict(nil, true); got != "first-online" {
		t.Errorf("got %q, want first-online", got)
	}
}

func TestOfflineVerdictFellOffline(t *testing.T) {
	prev := true
	if got := OfflineVerdict(&prev, false); got != "fell-offline" {
		t.Errorf("got %q, want fell-offline", got)
	}
}

func TestOfflineVerdictBackOnline(t *testing.T) {
	prev := false
	if got := OfflineVerdict(&prev, true); got != "back-online" {
		t.Errorf("got %q, want back-online", got)
	}
}

func TestOfflineVerdictSteadyStates(t *testing.T) {
	on, off := true, false
	if got := OfflineVerdict(&on, true); got != "online" {
		t.Errorf("got %q, want online", got)
	}
	if got := OfflineVerdict(&off, false); got != "offline" {
		t.Errorf("got %q, want offline", got)
	}
}

func TestOfflineVerdictTheWarnOnceFlow(t *testing.T) {
	// open offline → warn-now (the modal); recover → back-online;
	// fall again → fell-offline (a line, never a second modal)
	if got := OfflineVerdict(nil, false); got != "warn-now" {
		t.Errorf("got %q", got)
	}
	off := false
	if got := OfflineVerdict(&off, true); got != "back-online" {
		t.Errorf("got %q", got)
	}
	on := true
	if got := OfflineVerdict(&on, false); got != "fell-offline" {
		t.Errorf("got %q", got)
	}
	off2 := false
	if got := OfflineVerdict(&off2, false); got != "offline" {
		t.Errorf("got %q", got)
	}
}

// ---------------------------------------------------------------------------
// 6) SentinelLine
// ---------------------------------------------------------------------------

func TestSentinelLineFullLine(t *testing.T) {
	line := SentinelLine(Snapshot{
		Net:   ProbeInternetResult{Online: true, Endpoint: "https://api.github.com"},
		AI:    AIStatusResult{State: "on", Kind: "local", Engine: "Ollama"},
		Proxy: ProxyResult{State: "connected"},
	})
	if !strings.Contains(line, "Net: online (api.github.com)") {
		t.Errorf("line: %q", line)
	}
	if !strings.Contains(line, "AI: on — local (Ollama)") {
		t.Errorf("line: %q", line)
	}
	if !strings.Contains(line, "Proxy: connected") {
		t.Errorf("line: %q", line)
	}
}

func TestSentinelLineOfflineCloudAI(t *testing.T) {
	line := SentinelLine(Snapshot{
		Net:   ProbeInternetResult{Online: false},
		AI:    AIStatusResult{State: "on", Kind: "cloud"},
		Proxy: ProxyResult{State: "disabled"},
	})
	if !strings.Contains(line, "Net: offline") {
		t.Errorf("line: %q", line)
	}
	if !strings.Contains(line, "AI: on — cloud") {
		t.Errorf("line: %q", line)
	}
	if !strings.Contains(line, "Proxy: off") {
		t.Errorf("line: %q", line)
	}
}

func TestSentinelLineHostileSnapshotNeverPanics(t *testing.T) {
	line := SentinelLine(Snapshot{})
	if line == "" {
		t.Error("empty line")
	}
	// the zero snapshot: offline net, unknown ai, unknown proxy
	want := "Net: offline · AI: unknown · Proxy: unknown"
	if line != want {
		t.Errorf("line = %q, want %q", line, want)
	}
}

func TestSentinelLineShortFallsBackToState(t *testing.T) {
	line := SentinelLine(Snapshot{
		AI: AIStatusResult{State: "warn", Short: ""}})
	if !strings.Contains(line, "AI: warn") {
		t.Errorf("line: %q", line)
	}
}

func TestSentinelLineEndpointTruncatedTo32(t *testing.T) {
	long := "https://" + strings.Repeat("x", 80) + ".example"
	line := SentinelLine(Snapshot{
		Net: ProbeInternetResult{Online: true, Endpoint: long}})
	// host part truncated to 32 runes
	if !strings.Contains(line, "("+strings.Repeat("x", 32)+")") {
		t.Errorf("line: %q", line)
	}
}
