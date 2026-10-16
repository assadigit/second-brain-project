package connectioncheck

// worker_test.go — the Worker leg (PROMOTED to queue-reachability per
// ADR-0004 Am.3/R-0): the version compare, the deploy-staleness
// ladder, the enabled-gated reachability failure, and THE WORKER UA
// LAW (GitCurator-connection-check + proxy bypassed). Ported from the
// reference's TestCheckWorkerVersion.

import (
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
)

func TestCheckWorkerVersionNoURLConfiguredReturnsNil(t *testing.T) {
	if r := CheckWorkerVersion(map[string]any{}, nil); r != nil {
		t.Errorf("empty config: %+v", r)
	}
	if r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": ""}, nil); r != nil {
		t.Errorf("empty URL: %+v", r)
	}
}

func TestCheckWorkerVersionMatchingVersionIsOk(t *testing.T) {
	r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": "https://bot.example"},
		func(u string, timeoutS float64) (any, error) {
			return map[string]any{"version": constants.ExpectedWorkerVersion}, nil
		})
	if r == nil || r.Level != "ok" {
		t.Fatalf("result = %+v", r)
	}
	if !strings.Contains(r.Detail, "matches") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckWorkerVersionStaleWarnsWithTheRedeployRemedy(t *testing.T) {
	r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": "https://bot.example"},
		func(u string, timeoutS float64) (any, error) {
			return map[string]any{"version": "0.22.0"}, nil
		})
	if r == nil || r.Level != "warn" {
		t.Fatalf("result = %+v", r)
	}
	if !strings.Contains(r.Detail, "0.22.0") {
		t.Errorf("detail: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, constants.ExpectedWorkerVersion) {
		t.Errorf("detail: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, "deploy-latest.sh") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckWorkerVersionNewerDeployedWorkerIsInfoNotAWarning(t *testing.T) {
	parts := strings.Split(constants.ExpectedWorkerVersion, ".")
	// minor + 1 (the reference's own computation)
	if len(parts) < 2 {
		t.Fatalf("unexpected EXPECTED_WORKER_VERSION %q",
			constants.ExpectedWorkerVersion)
	}
	newer := parts[0] + "." + itoa(pyIntDefault(parts[1], 0)+1)
	if len(parts) > 2 {
		newer += "." + parts[2]
	}
	r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": "https://bot.example"},
		func(u string, timeoutS float64) (any, error) {
			return map[string]any{"version": newer}, nil
		})
	if r == nil || r.Level != "info" {
		t.Fatalf("result = %+v", r)
	}
	if !strings.Contains(r.Detail, "NEWER") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckWorkerVersionNoVersionReportedIsWarn(t *testing.T) {
	r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": "https://bot.example"},
		func(u string, timeoutS float64) (any, error) {
			return map[string]any{}, nil
		})
	if r == nil || r.Level != "warn" {
		t.Fatalf("result = %+v", r)
	}
	if !strings.Contains(r.Detail, "no version reported by /health") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestVersionCmpDirections(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"0.25.0", "0.25.0", 0},
		{"0.22.0", "0.25.0", -1},
		{"0.26.1", "0.25.0", 1},
		{"v0.25.0", "0.25.0", 0},
		{"0.25", "0.25.0", 0},
		{"1.0.0", "0.99.99", 1},
		// the digits-extraction law: "0-rc1" keeps its digits "01" → 1
		// (Python's filter does exactly this — a suffix is NOT zero)
		{"0.25.0-rc1", "0.25.0", 1},
		// truly non-numeric segments compare as 0
		{"x.y.z", "0.0.0", 0},
	}
	for _, c := range cases {
		if got := VersionCmp(c.a, c.b); got != c.want {
			t.Errorf("VersionCmp(%q, %q) = %d, want %d", c.a, c.b, got, c.want)
		}
	}
}

func TestCheckWorkerVersionUnreachableWarnsOrErrorsByEnabled(t *testing.T) {
	boom := func(u string, timeoutS float64) (any, error) {
		return nil, errors.New("refused")
	}
	r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": "https://bot.example"}, boom)
	if r == nil || r.Level != "warn" {
		t.Fatalf("result = %+v", r)
	}
	r2 := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": "https://bot.example",
			"cloudflare_enabled": true}, boom)
	if r2 == nil || r2.Level != "error" {
		t.Fatalf("result = %+v", r2)
	}
	if !strings.Contains(r2.Detail, "https://bot.example/health") {
		t.Errorf("detail: %q", r2.Detail)
	}
}

// THE UA TABLE (3/3) — THE WORKER UA LAW: the production /health fetch
// speaks "GitCurator-connection-check", is a bare GET, and BYPASSES
// the proxy (the sync client's own law — an intercepting proxy must
// never fake a ⚠️/❌ for a Worker that works).
func TestWorkerDirectGetUALaw(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/health": route(200, map[string]any{
			"version": constants.ExpectedWorkerVersion}),
	})
	data, err := workerDirectGet(srv.root()+"/health", 8.0)
	if err != nil {
		t.Fatal(err)
	}
	if got := pyStr(dictValue(data, "version")); got != constants.ExpectedWorkerVersion {
		t.Errorf("version = %q", got)
	}
	captured := srv.captured()
	if len(captured) != 1 {
		t.Fatalf("requests = %d, want 1", len(captured))
	}
	if captured[0].UA != "GitCurator-connection-check" {
		t.Errorf("UA = %q", captured[0].UA)
	}
	if captured[0].Method != "GET" {
		t.Errorf("method = %s", captured[0].Method)
	}
	if captured[0].BodyN != 0 {
		t.Errorf("GET carried a body: %d bytes", captured[0].BodyN)
	}
}

func TestWorkerFetchProxyBypassStructuralLaw(t *testing.T) {
	// the transport IS the law: the direct client never consults a
	// proxy (ProxyHandler({}) in the reference; Proxy == nil here)
	tr, ok := directClient(1.0).Transport.(*http.Transport)
	if !ok {
		t.Fatal("direct client transport is not *http.Transport")
	}
	if tr.Proxy != nil {
		t.Error("direct client has a proxy hook — the bypass law is broken")
	}
	// and the plain client (the non-worker HTTP path) honors the env
	ptr, ok := plainClient(1.0).Transport.(*http.Transport)
	if !ok {
		t.Fatal("plain client transport is not *http.Transport")
	}
	if ptr.Proxy == nil {
		t.Error("plain client lost the environment proxy")
	}
}

func TestCheckWorkerVersionLiveAgainstFakeWorker(t *testing.T) {
	// the production fetch against a fake Worker — the promoted
	// queue-reachability leg end-to-end on loopback
	srv := newJSONServer(t, map[string][2]any{
		"/health": route(200, map[string]any{
			"version": constants.ExpectedWorkerVersion}),
	})
	r := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": srv.root()}, nil)
	if r == nil || r.Level != LevelOK {
		t.Fatalf("result = %+v", r)
	}
	if !strings.Contains(r.Detail, "matches this app") {
		t.Errorf("detail: %q", r.Detail)
	}
	// and the dead Worker → the reachable-check-failed line
	r2 := CheckWorkerVersion(
		map[string]any{"cloudflare_worker_url": deadURL(t)}, nil)
	if r2 == nil || r2.Level != LevelWarn {
		t.Fatalf("result = %+v", r2)
	}
	if !strings.Contains(r2.Detail, "reachable check failed") {
		t.Errorf("detail: %q", r2.Detail)
	}
}
