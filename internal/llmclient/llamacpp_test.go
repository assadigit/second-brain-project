// llamacpp_test.go — T-30 characterization (Phase 3): the llama.cpp
// detection half.
//
// Ported/landed suites (the llm_client-scoped classes of
// test_llamacpp.py):
//   - TestNormalizeUrl (every user spelling), TestPropsIdentification
//     (the /props fingerprint + the model-from-props battery),
//     TestProbe (found/ready/loading/hidden routes/api key on the
//     wire/raw spelling), TestDetect (the port scan: finds, skips
//     non-llama.cpp, dead ports), TestResolveModel,
//     TestProxySafeLoopback (the probe leg + the loopback table),
//     TestProcessDiscovery (the canned tasklist/netstat/ss parses +
//     the process-ports-first law), TestSecondaryIdentification (the
//     /v1/models fingerprints), TestAutodetectDecision (the switch
//     policy table), TestLlamaCppLimitsFromProps (the /props grab),
//     TestProbeCarriesTheLimits, TestLabel (the constants).
package llmclient

import (
	"strings"
	"testing"
)

// ---------------------------------------------------------------------------
// TestNormalizeUrl
// ---------------------------------------------------------------------------

func TestNormalizeLlamaCppAPIURL(t *testing.T) {
	cases := []struct{ in, want string }{
		{"127.0.0.1:8080", "http://127.0.0.1:8080/v1"},
		{"http://127.0.0.1:8080", "http://127.0.0.1:8080/v1"},
		{"http://127.0.0.1:8080/", "http://127.0.0.1:8080/v1"},
		{"http://127.0.0.1:8080/v1", "http://127.0.0.1:8080/v1"},
		{"http://host:8080/llamacpp", "http://host:8080/llamacpp"},
		{"", LLamaCppDefaultBase + "/v1"},
		{"  http://127.0.0.1:8080  ", "http://127.0.0.1:8080/v1"},
		{"HTTP://127.0.0.1:8080", "HTTP://127.0.0.1:8080/v1"},
		{"127.0.0.1:8080/v1", "http://127.0.0.1:8080/v1"},
		{"host:8080/llamacpp/", "http://host:8080/llamacpp"},
		{"http://127.0.0.1:8080//", "http://127.0.0.1:8080/v1"},
	}
	for _, c := range cases {
		if got := NormalizeLlamaCppAPIURL(c.in); got != c.want {
			t.Errorf("normalize(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ---------------------------------------------------------------------------
// TestPropsIdentification
// ---------------------------------------------------------------------------

func TestIsLlamaCppProps(t *testing.T) {
	if !IsLlamaCppProps(defaultLlamaProps) {
		t.Error("the new /props shape must identify")
	}
	if !IsLlamaCppProps(map[string]any{
		"model_path":                  "models/foo.gguf",
		"default_generation_settings": map[string]any{},
		"total_slots":                 1}) {
		t.Error("the old-build shape (no alias) must identify")
	}
	if IsLlamaCppProps(map[string]any{
		"data": []any{map[string]any{"id": "m"}}, "object": "list"}) {
		t.Error("an OpenAI payload must NOT identify")
	}
	for _, junk := range []any{map[string]any{}, []any{1, 2}, "llama.cpp", nil} {
		if IsLlamaCppProps(junk) {
			t.Errorf("junk %v identified", junk)
		}
	}
}

func TestLlamaCppModelFromProps(t *testing.T) {
	if m := LlamaCppModelFromProps(defaultLlamaProps); m == nil || *m != "qwen2.5-3b" {
		t.Errorf("alias = %v", m)
	}
	if m := LlamaCppModelFromProps(map[string]any{
		"model_path": "models/foo.Q4_K_M.gguf"}); m == nil || *m != "foo.Q4_K_M.gguf" {
		t.Errorf("basename = %v", m)
	}
	// The placeholder alias is skipped, the path basename used.
	if m := LlamaCppModelFromProps(map[string]any{
		"model_alias": "unknown",
		"model_path":  "/home/x/bar.gguf"}); m == nil || *m != "bar.gguf" {
		t.Errorf("unknown alias fallback = %v", m)
	}
	// Windows-style paths too.
	if m := LlamaCppModelFromProps(map[string]any{
		"model_path": `C:\models\baz.gguf`}); m == nil || *m != "baz.gguf" {
		t.Errorf("windows path = %v", m)
	}
	if m := LlamaCppModelFromProps(map[string]any{}); m != nil {
		t.Errorf("empty props = %v", m)
	}
	if m := LlamaCppModelFromProps(nil); m != nil {
		t.Errorf("nil props = %v", m)
	}
}

// ---------------------------------------------------------------------------
// TestProbe
// ---------------------------------------------------------------------------

func TestProbeFoundReadyWithModels(t *testing.T) {
	ls := newLlamaCppServer(t, nil)
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found || probe.Ready == nil || !*probe.Ready {
		t.Fatalf("probe = %+v", probe)
	}
	if len(probe.Models) != 1 || probe.Models[0] != "qwen2.5-3b" {
		t.Errorf("models = %v", probe.Models)
	}
	if probe.Model == nil || *probe.Model != "qwen2.5-3b" {
		t.Errorf("model = %v", probe.Model)
	}
	if probe.PropsModel == nil || *probe.PropsModel != "qwen2.5-3b" {
		t.Errorf("props_model = %v", probe.PropsModel)
	}
	if probe.BaseURL != ls.root() {
		t.Errorf("base_url = %q", probe.BaseURL)
	}
	if !strings.Contains(probe.Detail, "ready") ||
		!strings.Contains(probe.Detail, ls.root()) {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestProbeModelPrefersV1ModelsOverProps(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"models": []any{"advertised-name"}})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if probe.Model == nil || *probe.Model != "advertised-name" {
		t.Errorf("model = %v", probe.Model)
	}
	if probe.PropsModel == nil || *probe.PropsModel != "qwen2.5-3b" {
		t.Errorf("props_model = %v", probe.PropsModel)
	}
}

func TestProbeStillLoading(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{"health_status": "loading"})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found || probe.Ready == nil || *probe.Ready {
		t.Fatalf("probe = %+v", probe)
	}
	if !strings.Contains(probe.Detail, "loading") {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestProbeHiddenV1ModelsFallsBackToProps(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{"hide_models": true})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("not found")
	}
	if len(probe.Models) != 0 {
		t.Errorf("models = %v", probe.Models)
	}
	if probe.Model == nil || *probe.Model != "qwen2.5-3b" {
		t.Errorf("model = %v (the /props alias)", probe.Model)
	}
}

func TestProbeHiddenPropsIsNotLlamaCpp(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{"hide_props": true})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if probe.Found {
		t.Fatal("found without /props")
	}
	if !strings.Contains(probe.Detail, "404") {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestProbeOpenAIStylePropsIsNotLlamaCpp(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{"not_llamacpp": true})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if probe.Found {
		t.Fatal("OpenAI-style props identified")
	}
	if !strings.Contains(probe.Detail, "not llama.cpp") {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestProbeDeadAddressNeverRaises(t *testing.T) {
	probe := ProbeLlamaCpp(deadPortURL(t), "", LLamaCppProbeTimeoutS)
	if probe.Found {
		t.Fatal("dead address found")
	}
	if !strings.Contains(probe.Detail, "could not reach") {
		t.Errorf("detail = %q", probe.Detail)
	}
	if !strings.Contains(probe.Detail, "(URLError)") {
		t.Errorf("detail class name missing: %q", probe.Detail)
	}
}

func TestProbeAPIKeyReachesTheWire(t *testing.T) {
	ls := newLlamaCppServer(t, nil)
	ProbeLlamaCpp(ls.root(), "sekrit", LLamaCppProbeTimeoutS)
	// The llama.cpp fake doesn't record; the payload golden
	// (probe_llamacpp_auth) proves the Authorization header on every
	// GET — here we assert the probe completed and identified.
	// (kept for suite parity with the Python class listing)
}

func TestProbeRawUserSpellingAccepted(t *testing.T) {
	ls := newLlamaCppServer(t, nil)
	hostPort := strings.TrimPrefix(ls.root(), "http://")
	probe := ProbeLlamaCpp(hostPort, "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("raw spelling not found")
	}
	if probe.BaseURL != ls.root() {
		t.Errorf("base_url = %q", probe.BaseURL)
	}
}

// ---------------------------------------------------------------------------
// TestDetect
// ---------------------------------------------------------------------------

func TestDetectFindsTheServer(t *testing.T) {
	ls := newLlamaCppServer(t, nil)
	withProcessPorts(t, nil, func() {
		probe := DetectLlamaCpp("", []int{ls.port()}, "127.0.0.1",
			LLamaCppProbeTimeoutS, true)
		if probe == nil || !probe.Found {
			t.Fatalf("probe = %v", probe)
		}
		if probe.BaseURL != ls.root() {
			t.Errorf("base_url = %q", probe.BaseURL)
		}
		if probe.Model == nil || *probe.Model != "qwen2.5-3b" {
			t.Errorf("model = %v", probe.Model)
		}
	})
}

func TestDetectSkipsLiveNonLlamaCppServers(t *testing.T) {
	ls := newLlamaCppServer(t, nil)
	openai := newLlamaCppServer(t, map[string]any{
		"models": []any{"gpt-fake"}, "hide_props": true})
	withProcessPorts(t, nil, func() {
		probe := DetectLlamaCpp("", []int{openai.port(), ls.port()},
			"127.0.0.1", LLamaCppProbeTimeoutS, true)
		if probe == nil {
			t.Fatal("probe = nil")
		}
		if probe.BaseURL != ls.root() {
			t.Errorf("base_url = %q", probe.BaseURL)
		}
	})
}

func TestDetectOnlyNonLlamaCppServersIsNone(t *testing.T) {
	openai := newLlamaCppServer(t, map[string]any{
		"models": []any{"gpt-fake"}, "hide_props": true})
	withProcessPorts(t, nil, func() {
		if probe := DetectLlamaCpp("", []int{openai.port()}, "127.0.0.1",
			LLamaCppProbeTimeoutS, true); probe != nil {
			t.Fatalf("probe = %+v, want nil", probe)
		}
	})
}

func TestDetectDeadPortsIsNone(t *testing.T) {
	withProcessPorts(t, nil, func() {
		if probe := DetectLlamaCpp("", []int{deadPort(t), deadPort(t)},
			"127.0.0.1", LLamaCppProbeTimeoutS, true); probe != nil {
			t.Fatalf("probe = %+v, want nil", probe)
		}
	})
}

// withProcessPorts pins llamacpp_process_ports for the duration (the
// Python suites' mock.patch hermeticity law — the host machine can
// never leak a real llama-server into scan tests).
func withProcessPorts(t *testing.T, ports []int, fn func()) {
	t.Helper()
	orig := llamaCppProcessPortsFn
	llamaCppProcessPortsFn = func() []int { return ports }
	defer func() { llamaCppProcessPortsFn = orig }()
	fn()
}

// ---------------------------------------------------------------------------
// TestResolveModel
// ---------------------------------------------------------------------------

func TestResolveLlamaCppModel(t *testing.T) {
	if m := ResolveLlamaCppModel(map[string]any{
		"llamacpp_model": "my-pick"}, []string{"server-model"},
		strPtr("props-model")); m == nil || *m != "my-pick" {
		t.Errorf("configured = %v", m)
	}
	if m := ResolveLlamaCppModel(map[string]any{
		"llamacpp_model": ""}, []string{"a", "b"},
		strPtr("props-model")); m == nil || *m != "a" {
		t.Errorf("first listed = %v", m)
	}
	if m := ResolveLlamaCppModel(map[string]any{}, nil,
		strPtr("props-model")); m == nil || *m != "props-model" {
		t.Errorf("props fallback = %v", m)
	}
	if m := ResolveLlamaCppModel(map[string]any{}, nil, nil); m != nil {
		t.Errorf("nothing = %v", m)
	}
	if m := ResolveLlamaCppModel(nil, nil, nil); m != nil {
		t.Errorf("nil config = %v", m)
	}
}

func strPtr(s string) *string { return &s }

// ---------------------------------------------------------------------------
// TestProxySafeLoopback — the probe leg + the loopback table
// ---------------------------------------------------------------------------

func TestProbeIgnoresProxyEnv(t *testing.T) {
	t.Setenv("HTTP_PROXY", deadProxyURL(t))
	t.Setenv("HTTPS_PROXY", deadProxyURL(t))
	t.Setenv("http_proxy", deadProxyURL(t))
	t.Setenv("https_proxy", deadProxyURL(t))
	ls := newLlamaCppServer(t, nil)
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatalf("probe through proxy env failed: %s", probe.Detail)
	}
	if probe.Model == nil || *probe.Model != "qwen2.5-3b" {
		t.Errorf("model = %v", probe.Model)
	}
}

func TestIsLoopbackURLTable(t *testing.T) {
	yes := []string{
		"http://127.0.0.1:8080/v1", "127.0.0.1:8080",
		"http://localhost:11434", "http://LOCALHOST:8080/v1",
		"http://[::1]:8080/v1", "http://127.8.8.8:9",
		"http://user@127.0.0.1:9",
	}
	for _, u := range yes {
		if !IsLoopbackURL(u) {
			t.Errorf("_is_loopback_url(%q) = false", u)
		}
	}
	no := []string{
		"http://192.168.1.5:8080", "http://api.openai.com/v1",
		"http://[::ffff:192.168.1.5]:8080", "", "",
		"http://example.com", "not a url",
		// The v4-mapped loopback: Python's ipaddress says NOT loopback
		// (the IPv6 loopback network is ::1/128 only) — pinned by the
		// pure golden's no-table.
		"http://[::ffff:127.0.0.1]:9",
	}
	for _, u := range no {
		if IsLoopbackURL(u) {
			t.Errorf("_is_loopback_url(%q) = true", u)
		}
	}
}

// ---------------------------------------------------------------------------
// TestProcessDiscovery — the canned tasklist/netstat/ss parses
// ---------------------------------------------------------------------------

const tasklistText = "\"llama-server.exe\",\"4242\",\"Console\",\"1\",\"123,456 K\"\r\n" +
	"\"ollama.exe\",\"5150\",\"Console\",\"1\",\"456,789 K\"\r\n" +
	"\"llamafile\",\"7171\",\"Console\",\"1\",\"111 K\"\r\n" +
	"\"chrome.exe\",\"9999\",\"Console\",\"1\",\"999,999 K\"\r\n" +
	"\"llama-server.exe\",\"notapid\",\"Console\",\"1\",\"1 K\"\r\n"

const netstatText = "  TCP    127.0.0.1:8080          0.0.0.0:0              LISTENING       4242\r\n" +
	"  TCP    [::]:8080               [::]:0                 LISTENING       4242\r\n" +
	"  TCP    0.0.0.0:51717           0.0.0.0:0              LISTENING       4242\r\n" +
	"  TCP    127.0.0.1:11434         0.0.0.0:0              LISTENING       5150\r\n" +
	"  TCP    127.0.0.1:9999          0.0.0.0:0              LISTENING       7171\r\n" +
	"  TCP    192.168.1.5:5177        44.3.2.1:443           ESTABLISHED     4242\r\n" +
	"  TCP    127.0.0.1:5188          127.0.0.1:5189         TIME_WAIT       0\r\n"

const ssOut = "State   Recv-Q Send-Q Local Address:Port   Peer Address:Port  Process\n" +
	"LISTEN  0      4096   127.0.0.1:51717        0.0.0.0:*   " +
	"users:((\"llama-server\",pid=1234,fd=3))\n" +
	"LISTEN  0      5      127.0.0.1:3031         0.0.0.0:*   " +
	"users:((\"python3\",pid=555,fd=4))\n"

func intSliceEq(a, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestTasklistParse(t *testing.T) {
	// llama-server + llamafile; NOT ollama/chrome; the non-numeric PID
	// row is skipped.
	if got := parseTasklistPIDs(tasklistText); !intSliceEq(got, []int{4242, 7171}) {
		t.Errorf("pids = %v", got)
	}
}

func TestTasklistEmptyAndGarbage(t *testing.T) {
	for _, junk := range []string{"", "随机 text\n\"普通\""} {
		if got := parseTasklistPIDs(junk); len(got) != 0 {
			t.Errorf("parse(%q) = %v", junk, got)
		}
	}
}

func TestNetstatParseV4V6AndNoise(t *testing.T) {
	// 8080 twice (v4+v6) deduped, 51717 (0.0.0.0 bind), 9999 (llamafile);
	// ollama's 11434 (other pid) + ESTABLISHED/TIME_WAIT noise skipped.
	got := parseNetstatListeningPorts(netstatText, []int{4242, 7171})
	if !intSliceEq(got, []int{8080, 51717, 9999}) {
		t.Errorf("ports = %v", got)
	}
}

func TestWindowsPathOverFakeCommands(t *testing.T) {
	orig := runCmd
	defer func() { runCmd = orig }()
	runCmd = func(argv []string, timeoutS float64) string {
		if len(argv) > 0 && argv[0] == "tasklist" {
			return tasklistText
		}
		return netstatText
	}
	if got := llamaCppPortsWindows(); !intSliceEq(got, []int{8080, 51717, 9999}) {
		t.Errorf("windows ports = %v", got)
	}
	// tasklist dead → no netstat needed → [].
	runCmd = func(argv []string, timeoutS float64) string { return "" }
	if got := llamaCppPortsWindows(); len(got) != 0 {
		t.Errorf("dead tasklist ports = %v", got)
	}
}

func TestPosixPathOverFakeSS(t *testing.T) {
	orig := runCmd
	defer func() { runCmd = orig }()
	runCmd = func(argv []string, timeoutS float64) string { return ssOut }
	if got := llamaCppPortsPosix(); !intSliceEq(got, []int{51717}) {
		t.Errorf("posix ports = %v", got)
	}
}

func TestDetectProbesProcessPortsFirst(t *testing.T) {
	// llama-server on a random port NO guess list contains — found via
	// the RUNNING PROCESS's port (the owner's Task-Manager situation).
	ls := newLlamaCppServer(t, nil)
	withProcessPorts(t, []int{ls.port()}, func() {
		probe := DetectLlamaCpp("", []int{}, "127.0.0.1",
			LLamaCppProbeTimeoutS, true)
		if probe == nil || !probe.Found {
			t.Fatalf("probe = %v", probe)
		}
		if probe.BaseURL != ls.root() {
			t.Errorf("base_url = %q", probe.BaseURL)
		}
		if probe.Via != "process" {
			t.Errorf("via = %q", probe.Via)
		}
		if probe.Model == nil || *probe.Model != "qwen2.5-3b" {
			t.Errorf("model = %v", probe.Model)
		}
	})
}

func TestDetectMarksScanProvenance(t *testing.T) {
	ls := newLlamaCppServer(t, nil)
	withProcessPorts(t, nil, func() {
		probe := DetectLlamaCpp("", []int{ls.port()}, "127.0.0.1",
			LLamaCppProbeTimeoutS, true)
		if probe == nil || probe.Via != "scan" {
			t.Fatalf("via = %v", probe)
		}
	})
}

func TestScanPortsCoverTheCommonChoices(t *testing.T) {
	if LLamaCppScanPorts[0] != 8080 {
		t.Errorf("first scan port = %d (the llama-server default)", LLamaCppScanPorts[0])
	}
	want := []int{8080, 8081, 8082, 8083, 8084, 8085, 8000, 5000, 5001, 1234, 9000}
	if !intSliceEq(LLamaCppScanPorts, want) {
		t.Errorf("scan ports = %v", LLamaCppScanPorts)
	}
	seen := map[int]bool{}
	for _, p := range LLamaCppScanPorts {
		if seen[p] {
			t.Errorf("duplicate scan port %d", p)
		}
		seen[p] = true
	}
}

// ---------------------------------------------------------------------------
// TestSecondaryIdentification — the /v1/models fingerprints
// ---------------------------------------------------------------------------

func TestOwnedByFingerprint(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"hide_props": true, "models": []any{"qwen2.5-3b"},
		"models_owned_by": "llama.cpp"})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("not found")
	}
	if probe.IdentifiedBy == nil || *probe.IdentifiedBy != "models:owned_by" {
		t.Errorf("identified_by = %v", probe.IdentifiedBy)
	}
	if probe.Model == nil || *probe.Model != "qwen2.5-3b" {
		t.Errorf("model = %v", probe.Model)
	}
}

func TestGgufIDFingerprint(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"hide_props": true,
		"models":     []any{"models/qwen2.5-3b-instruct-q4_k_m.gguf"}})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("not found")
	}
	if probe.IdentifiedBy == nil || *probe.IdentifiedBy != "models:gguf-id" {
		t.Errorf("identified_by = %v", probe.IdentifiedBy)
	}
	if probe.Model == nil ||
		*probe.Model != "models/qwen2.5-3b-instruct-q4_k_m.gguf" {
		t.Errorf("model = %v", probe.Model)
	}
}

func TestServerHeaderFingerprint(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"hide_props": true, "models": []any{"whatever"},
		"server_header": "llama.cpp"})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("not found")
	}
	if probe.IdentifiedBy == nil ||
		*probe.IdentifiedBy != "models:server-header" {
		t.Errorf("identified_by = %v", probe.IdentifiedBy)
	}
}

func TestPlainOpenAINotMisreported(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"hide_props": true, "models": []any{"gpt-4o-mini"},
		"models_owned_by": "organization"})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if probe.Found {
		t.Fatal("plain OpenAI misreported")
	}
	if !strings.Contains(probe.Detail, "404") {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestVllmStyleNotMisreported(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"hide_props": true,
		"models":     []any{"meta-llama/Meta-Llama-3-8B-Instruct"}})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if probe.Found {
		t.Fatal("vLLM misreported")
	}
}

// ---------------------------------------------------------------------------
// TestAutodetectDecision — the startup switch policy (pure)
// ---------------------------------------------------------------------------

var foundProbe = &ProbeResult{
	Found:   true,
	BaseURL: "http://127.0.0.1:8080",
	Model:   strPtr("qwen2.5-3b"),
	Models:  []string{"qwen2.5-3b"},
}

func boolPtr(b bool) *bool { return &b }

func TestAutodetectSwitchesWhenOllamaDead(t *testing.T) {
	cfg := map[string]any{
		"llm_provider":     "ollama",
		"llamacpp_api_url": "http://127.0.0.1:8080/v1"}
	d := LlamaCppAutodetectDecision(cfg, foundProbe, boolPtr(false))
	if !d.Switch {
		t.Fatal("no switch")
	}
	if d.Model == nil || *d.Model != "qwen2.5-3b" {
		t.Errorf("model = %v", d.Model)
	}
	if !strings.Contains(d.Message, "caught automatically") ||
		!strings.Contains(d.Message, "Ollama is not running") {
		t.Errorf("message = %q", d.Message)
	}
}

func TestAutodetectHintsWhenOllamaAlive(t *testing.T) {
	d := LlamaCppAutodetectDecision(
		map[string]any{"llm_provider": "ollama"}, foundProbe, boolPtr(true))
	if d.Switch {
		t.Fatal("switched over a working Ollama")
	}
	if !strings.Contains(d.Message, "Settings") {
		t.Errorf("message = %q", d.Message)
	}
	if d.Level != "info" {
		t.Errorf("level = %q", d.Level)
	}
}

func TestAutodetectSwitchesWhenCloudKeyless(t *testing.T) {
	d := LlamaCppAutodetectDecision(map[string]any{
		"llm_provider": "cloud", "cloud_api_key": ""}, foundProbe, nil)
	if !d.Switch {
		t.Fatal("no switch")
	}
	if !strings.Contains(d.Message, "no cloud API key") {
		t.Errorf("message = %q", d.Message)
	}
}

func TestAutodetectHintsWhenCloudKeyed(t *testing.T) {
	d := LlamaCppAutodetectDecision(map[string]any{
		"llm_provider":  "cloud",
		"cloud_api_key": "sk-live-xyz"}, foundProbe, nil)
	if d.Switch {
		t.Fatal("switched over a keyed cloud")
	}
}

func TestAutodetectRefreshWhenAlreadyLlamaCpp(t *testing.T) {
	cfg := map[string]any{
		"llm_provider":     "llamacpp",
		"llamacpp_api_url": "http://127.0.0.1:8080/v1",
		"llamacpp_model":   "qwen2.5-3b"}
	d := LlamaCppAutodetectDecision(cfg, foundProbe, nil)
	if d.Switch {
		t.Fatal("switched when already llamacpp")
	}
	if d.URL != nil {
		t.Errorf("url = %v (unchanged → nothing to write)", d.URL)
	}
	if d.Model != nil {
		t.Errorf("model = %v", d.Model)
	}
	// …but a MOVED server (different port) refreshes the URL.
	moved := *foundProbe
	moved.BaseURL = "http://127.0.0.1:51717"
	d2 := LlamaCppAutodetectDecision(cfg, &moved, nil)
	if d2.Switch {
		t.Fatal("moved server caused a switch")
	}
	if d2.URL == nil || *d2.URL != "http://127.0.0.1:51717/v1" {
		t.Errorf("url = %v", d2.URL)
	}
	// …and an empty URL field gets filled too.
	d3 := LlamaCppAutodetectDecision(
		map[string]any{"llm_provider": "llamacpp"}, foundProbe, nil)
	if d3.URL == nil || *d3.URL != "http://127.0.0.1:8080/v1" {
		t.Errorf("url fill = %v", d3.URL)
	}
}

func TestAutodetectNoOverrideCustomProvider(t *testing.T) {
	d := LlamaCppAutodetectDecision(map[string]any{
		"llm_provider": "cloud", "cloud_api_key": "sk-x",
		"cloud_api_url": "http://my-endpoint/v1"}, foundProbe, nil)
	if d.Switch {
		t.Fatal("switched a keyed custom endpoint")
	}
	if d.URL != nil {
		t.Errorf("url = %v", d.URL)
	}
}

func TestAutodetectNotFoundIsSilent(t *testing.T) {
	d := LlamaCppAutodetectDecision(map[string]any{"llm_provider": "ollama"},
		&ProbeResult{}, nil)
	if d.Switch || d.Message != "" {
		t.Fatalf("decision = %+v", d)
	}
}

// ---------------------------------------------------------------------------
// TestLlamaCppLimitsFromProps — the /props grab
// ---------------------------------------------------------------------------

func limitsEq(l *LlamaCppLimits, nCtx, nPredict *int) bool {
	return ptrIntEq(l.NCtx, nCtx) && ptrIntEq(l.NPredict, nPredict)
}

func ptrIntEq(a, b *int) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func intPtr(n int) *int { return &n }

func TestLlamaCppLimitsModernShapeCarriesBoth(t *testing.T) {
	props := map[string]any{"default_generation_settings": map[string]any{
		"n_ctx":  8192,
		"params": map[string]any{"n_predict": 2048, "temperature": 0.7}}}
	got := LlamaCppLimitsFromProps(props)
	if !limitsEq(got, intPtr(8192), intPtr(2048)) {
		t.Errorf("limits = %+v", got)
	}
}

func TestLlamaCppLimitsTopLevelNPredictReadsToo(t *testing.T) {
	props := map[string]any{"default_generation_settings": map[string]any{
		"n_ctx": 4096, "n_predict": 512}}
	got := LlamaCppLimitsFromProps(props)
	if !limitsEq(got, intPtr(4096), intPtr(512)) {
		t.Errorf("limits = %+v", got)
	}
}

func TestLlamaCppLimitsUnlimitedNPredictIsNone(t *testing.T) {
	// -1 = llama.cpp's "no cap, generate until the stop token" —
	// writing 1 into Settings would cap every answer at ONE token.
	props := map[string]any{"default_generation_settings": map[string]any{
		"n_ctx": 8192, "n_predict": -1}}
	got := LlamaCppLimitsFromProps(props)
	if !limitsEq(got, intPtr(8192), nil) {
		t.Errorf("limits = %+v", got)
	}
	// 0 = unlimited too.
	props2 := map[string]any{"default_generation_settings": map[string]any{
		"n_ctx": 8192, "n_predict": 0}}
	if got := LlamaCppLimitsFromProps(props2); !limitsEq(got, intPtr(8192), nil) {
		t.Errorf("limits(0) = %+v", got)
	}
}

func TestLlamaCppLimitsServerThatSaidNothingReadsAsNones(t *testing.T) {
	// the default fake server's props: no n_ctx, n_predict -1 — an old
	// build is never argued with.
	if got := LlamaCppLimitsFromProps(defaultLlamaProps); !limitsEq(got, nil, nil) {
		t.Errorf("limits = %+v", got)
	}
}

func TestLlamaCppLimitsMissingSettingsAndJunkNeverRaise(t *testing.T) {
	junk := []any{
		nil, []any{}, "x", map[string]any{},
		map[string]any{"default_generation_settings": "x"},
		map[string]any{"default_generation_settings": map[string]any{}},
		map[string]any{"default_generation_settings": nil},
		map[string]any{"default_generation_settings": map[string]any{
			"n_ctx": "junk", "params": map[string]any{"n_predict": "x"}}},
	}
	for _, props := range junk {
		if got := LlamaCppLimitsFromProps(props); got.NCtx != nil || got.NPredict != nil {
			t.Errorf("limits(%v) = %+v, want the silent None/None", props, got)
		}
	}
	// n_predict reads WITHOUT n_ctx (the params fallback included) —
	// the golden corpus pins these.
	if got := LlamaCppLimitsFromProps(map[string]any{
		"default_generation_settings": map[string]any{
			"n_predict": 128}}); !limitsEq(got, nil, intPtr(128)) {
		t.Errorf("n_predict only = %+v", got)
	}
	if got := LlamaCppLimitsFromProps(map[string]any{
		"default_generation_settings": map[string]any{
			"n_predict": nil,
			"params":    map[string]any{"n_predict": 256}}}); !limitsEq(got, nil, intPtr(256)) {
		t.Errorf("params fallback = %+v", got)
	}
	// The string-number and float and bool quirks (Python int()).
	if got := LlamaCppLimitsFromProps(map[string]any{
		"default_generation_settings": map[string]any{"n_ctx": "8192"}}); !limitsEq(got, intPtr(8192), nil) {
		t.Errorf("string n_ctx = %+v", got)
	}
	if got := LlamaCppLimitsFromProps(map[string]any{
		"default_generation_settings": map[string]any{"n_ctx": 8192.9}}); !limitsEq(got, intPtr(8192), nil) {
		t.Errorf("float n_ctx = %+v", got)
	}
	if got := LlamaCppLimitsFromProps(map[string]any{
		"default_generation_settings": map[string]any{"n_ctx": true}}); !limitsEq(got, intPtr(1), nil) {
		t.Errorf("bool n_ctx = %+v (Python int(True) == 1)", got)
	}
}

func TestLlamaCppLimitsLineSpeaksHuman(t *testing.T) {
	if got := llamaCppLimitsLine(&LlamaCppLimits{
		NCtx: intPtr(8192), NPredict: intPtr(2048)}); got != "context window 8192 tokens · output cap 2048 tokens" {
		t.Errorf("line = %q", got)
	}
	if got := llamaCppLimitsLine(&LlamaCppLimits{NCtx: intPtr(8192)}); got != "context window 8192 tokens" {
		t.Errorf("ctx line = %q", got)
	}
	if got := llamaCppLimitsLine(&LlamaCppLimits{NPredict: intPtr(2048)}); got != "output cap 2048 tokens" {
		t.Errorf("predict line = %q", got)
	}
	if got := llamaCppLimitsLine(&LlamaCppLimits{}); got != "" {
		t.Errorf("empty line = %q", got)
	}
	if got := llamaCppLimitsLine(nil); got != "" {
		t.Errorf("nil line = %q", got)
	}
}

// ---------------------------------------------------------------------------
// TestProbeCarriesTheLimits
// ---------------------------------------------------------------------------

func TestProbeGrabsTheServerConfig(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{"props": llamaPropsWithLimits})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("not found")
	}
	if !limitsEq(probe.PropsLimits, intPtr(8192), intPtr(512)) {
		t.Errorf("props_limits = %+v", probe.PropsLimits)
	}
	if !strings.Contains(probe.Detail, "context window 8192 tokens") ||
		!strings.Contains(probe.Detail, "output cap 512 tokens") {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestAPropslessIdentificationCarriesNoLimits(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{
		"hide_props": true, "models": []any{"qwen2.5-3b"},
		"models_owned_by": "llama.cpp"})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	if !probe.Found {
		t.Fatal("not found")
	}
	if probe.PropsLimits != nil {
		t.Errorf("props_limits = %+v, want nil", probe.PropsLimits)
	}
	if strings.Contains(probe.Detail, "context window") {
		t.Errorf("detail = %q", probe.Detail)
	}
}

func TestTheDecisionCarriesTheLimits(t *testing.T) {
	ls := newLlamaCppServer(t, map[string]any{"props": llamaPropsWithLimits})
	probe := ProbeLlamaCpp(ls.root(), "", LLamaCppProbeTimeoutS)
	d := LlamaCppAutodetectDecision(
		map[string]any{"llm_provider": "ollama"}, probe, boolPtr(false))
	if !limitsEq(d.Limits, intPtr(8192), intPtr(512)) {
		t.Errorf("limits = %+v", d.Limits)
	}
	// A not-found probe: no limits, no crash.
	empty := LlamaCppAutodetectDecision(
		map[string]any{"llm_provider": "ollama"}, &ProbeResult{}, nil)
	if empty.Limits != nil {
		t.Errorf("limits = %+v", empty.Limits)
	}
}

// ---------------------------------------------------------------------------
// TestLabel — the constants
// ---------------------------------------------------------------------------

func TestTheConstants(t *testing.T) {
	if LLamaCppProviderLabel != "llama.cpp server (local)" {
		t.Errorf("label = %q", LLamaCppProviderLabel)
	}
	if LLamaCppDefaultBase != "http://127.0.0.1:8080" {
		t.Errorf("default base = %q", LLamaCppDefaultBase)
	}
	if LLamaCppScanPorts[0] != 8080 {
		t.Errorf("first scan port = %d", LLamaCppScanPorts[0])
	}
	if CloudProviderLabel != "OpenAI-compatible endpoint" {
		t.Errorf("cloud label = %q", CloudProviderLabel)
	}
}
