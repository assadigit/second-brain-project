// thirddoor_test.go — the third door (the handshake), ported from
// test_thirddoor.py (TestTriggerLaw, TestVerdicts, TestFetchUrlWiring,
// the REAL-library class) and test_autopip.py's trigger/once-per-process
// laws: the fake-session seam (one level down from the injected-fetch
// pattern — the factory is package-level for exactly this), the
// availability seam, and the REAL tls-client transport against a local
// listener (the Go door ships armed — ADR-0006 §4).
package webfetch

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"
)

var ownerLine = "proxy: HTTP 403 — bot defense (Cloudflare: challenge) | " +
	"direct: HTTP 403 — bot defense (Cloudflare: challenge)"

func walledRes(reason string, httpStatus *int) *FetchResult {
	r := res(reason, httpStatus, "", nil)
	return r
}

func askThirdDoor(t *testing.T, walled *FetchResult, routes []*Proxy, enabled bool) *FetchResult {
	t.Helper()
	return MaybeImpersonate("http://walled.example.net/page", 5, 2_000_000,
		routes, enabled, walled, time.Now(), false)
}

func TestThirdDoorOwnersLineOpensDoor(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	r := askThirdDoor(t, walledRes(ownerLine, intPtr(403)), []*Proxy{nil}, true)
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "full" {
		t.Fatalf("status = %q", r.Status)
	}
	if !strings.Contains(r.Reason, "via impersonated Chrome") || !strings.Contains(r.Reason, ownerLine) {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*sessions) != 1 || len((*sessions)[0].calls) != 1 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
	call := (*sessions)[0].calls[0]
	if call.URL != "http://walled.example.net/page" || call.ProxyURL != "" {
		t.Fatalf("call = %+v (one dial, direct)", call)
	}
}

func TestThirdDoorEveryRefusalStatusEarnsDoor(t *testing.T) {
	for _, code := range []int{403, 405, 429, 451} {
		state := saveDoorState()
		sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
		r := askThirdDoor(t, walledRes("HTTP "+itoa(code), &code), []*Proxy{nil}, true)
		if !r.OK() {
			t.Fatalf("%d: not ok: %s", code, r.Reason)
		}
		if len(*sessions) != 1 {
			t.Fatalf("%d: sessions = %d", code, len(*sessions))
		}
		state.restore()
	}
}

func TestThirdDoorResourceTruthNeverAsks(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	r := askThirdDoor(t, walledRes("HTTP 404", intPtr(404)), []*Proxy{nil}, true)
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.Reason != "HTTP 404" {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestThirdDoorConnectionTruthNeverAsks(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	r := askThirdDoor(t, walledRes("proxy: timeout | direct: connection: timed out", nil), []*Proxy{nil}, true)
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, "timed out") || strings.Contains(r.Reason, "chrome-impersonated") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestThirdDoorOwnersTLSLineOpensDoor(t *testing.T) {
	// v0.46.0 — a TLS-handshake-class wall opens the door too; the DNS
	// verdict lands on the direct leg's getaddrinfo.
	state := saveDoorState()
	defer state.restore()
	armDoH(t, "unknown")
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	calls := &[]*Proxy{}
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		*calls = append(*calls, p)
		if p != nil {
			return res("connection: [SSL: SSLV3_ALERT_HANDSHAKE_FAILURE] sslv3 alert handshake failure (_ssl.c:1010)", nil, "", nil)
		}
		return res("connection: [Errno 11001] getaddrinfo failed", nil, "", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, Proxy: testProxy})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "full" {
		t.Fatalf("status = %q", r.Status)
	}
	if !strings.Contains(r.Reason, "via impersonated Chrome") ||
		!strings.Contains(r.Reason, "SSLV3_ALERT_HANDSHAKE_FAILURE") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if !strings.Contains(r.Reason, "verdict unknown") {
		t.Fatalf("reason = %q (the DNS verdict must land)", r.Reason)
	}
	if len(*calls) != 2 {
		t.Fatalf("both stdlib routes must run first: %v", *calls)
	}
	if len(*sessions) != 1 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestThirdDoorPlainTimeoutNeverAsksChrome(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armDoH(t, "unknown")
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("timeout", nil, "retry_later", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, Proxy: testProxy})
	if r.OK() || r.Category != "retry_later" {
		t.Fatalf("status=%q category=%q", r.Status, r.Category)
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestThirdDoorSingleDoorTLSOpensDirect(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("connection: [SSL: UNEXPECTED_EOF_WHILE_READING] EOF occurred in violation of protocol (_ssl.c:1010)", nil, "", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if len(*sessions) != 1 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
	call := (*sessions)[0].calls[0]
	if call.ProxyURL != "" {
		t.Fatalf("call = %+v (direct, no proxy)", call)
	}
}

func TestThirdDoorCertFailuresDoNotOpenDoor(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("connection: [SSL: CERTIFICATE_VERIFY_FAILED] certificate verify failed", nil, "", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2})
	if r.OK() {
		t.Fatal("must fail")
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestThirdDoorOptOutFlagHonored(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	r := askThirdDoor(t, walledRes(ownerLine, intPtr(403)), []*Proxy{nil}, false)
	if r.OK() || r.Reason != ownerLine {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestThirdDoorLoopbackNeverEscalates(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	r := MaybeImpersonate("http://127.0.0.1:8080/x", 5, 2_000_000,
		[]*Proxy{nil}, true, walledRes(ownerLine, intPtr(403)), time.Now(), false)
	if r.OK() || r.Reason != ownerLine {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

// ---------------------------------------------------------------------------
// the missing-library + autopip laws (the seams; the Go door ships
// armed — ADR-0006 §4's documented divergence)
// ---------------------------------------------------------------------------

func TestThirdDoorAlwaysArmedInGo(t *testing.T) {
	// The Go reality: tls-client is compiled in — the availability
	// default is TRUE and the install hint NEVER appears in production.
	state := saveDoorState()
	defer state.restore()
	if !impersonatorAvailable() {
		t.Fatal("the Go door ships armed (ADR-0006 §4)")
	}
}

func TestThirdDoorMissingLibraryGivesHint(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	r := askThirdDoor(t, walledRes(ownerLine, intPtr(403)), []*Proxy{nil}, true)
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, ownerLine) || !strings.Contains(r.Reason, "pip install curl_cffi") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*sessions) != 0 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestAutopipInstallsAndAnswers(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	pipCalls := 0
	armAvailable(false)
	impMu.Lock()
	autopipTried = false
	autopipOK = false
	pipInstallImpersonator = func() bool {
		pipCalls++
		armAvailable(true) // the probe re-arms after a success
		return true
	}
	impMu.Unlock()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, ImpersonateAutopip: true})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if pipCalls != 1 {
		t.Fatalf("pip calls = %d", pipCalls)
	}
	if len(*sessions) != 1 {
		t.Fatalf("sessions = %d", len(*sessions))
	}
}

func TestAutopipOffKeepsHint(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	pipCalls := 0
	armAvailable(false)
	impMu.Lock()
	pipInstallImpersonator = func() bool {
		pipCalls++
		return true
	}
	impMu.Unlock()
	armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2})
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, "pip install curl_cffi") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if pipCalls != 0 {
		t.Fatalf("pip calls = %d", pipCalls)
	}
}

func TestAutopipFailedInstallKeepsHintNeverPanics(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	pipCalls := 0
	armAvailable(false)
	impMu.Lock()
	autopipTried = false
	autopipOK = false
	pipInstallImpersonator = func() bool {
		pipCalls++
		return false
	}
	impMu.Unlock()
	armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	}
	r := FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, ImpersonateAutopip: true})
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, "pip install curl_cffi") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if pipCalls != 1 {
		t.Fatalf("pip calls = %d", pipCalls)
	}
}

func TestAutopipOncePerProcess(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	pipCalls := 0
	armAvailable(false)
	impMu.Lock()
	pipInstallImpersonator = func() bool {
		pipCalls++
		return false // failed installs are remembered too
	}
	autopipTried = false
	autopipOK = false
	impMu.Unlock()
	armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	}
	FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, ImpersonateAutopip: true})
	FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, ImpersonateAutopip: true})
	if pipCalls != 1 {
		t.Fatalf("pip calls = %d (once per process)", pipCalls)
	}
}

func TestAutopipResourceTruthNeverInstalls(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	pipCalls := 0
	armAvailable(false)
	impMu.Lock()
	pipInstallImpersonator = func() bool {
		pipCalls++
		return true
	}
	impMu.Unlock()
	armFakeSession(&fakeImpResponse{StatusCode: 200})
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 404", intPtr(404), "dead", nil)
	}
	armArchive(t, "no-snapshot", "")
	FetchURL("http://walled.example.net/page", &Options{TimeoutS: 2, ImpersonateAutopip: true})
	if pipCalls != 0 {
		t.Fatalf("pip calls = %d", pipCalls)
	}
}

// ---------------------------------------------------------------------------
// the verdicts (fake sessions)
// ---------------------------------------------------------------------------

func TestThirdDoorSuccessIsRealSuccess(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armFakeSession(&fakeImpResponse{
		StatusCode: 200,
		Body:       []byte("<html><head><title>The Real Page</title></head><body>fix live</body></html>"),
	})
	r := askThirdDoor(t, walledRes(ownerLine, intPtr(403)), []*Proxy{nil}, true)
	if !r.OK() || r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 200 {
		t.Fatalf("http = %v", r.HTTPStatus)
	}
	if !strings.Contains(string(r.Body), "fix live") {
		t.Fatalf("body = %q", string(r.Body))
	}
	if !strings.Contains(r.Reason, "bot defense (Cloudflare: challenge)") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestThirdDoorChallengeSurvivesThreeLegLine(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	challenged := &fakeImpResponse{
		StatusCode: 403,
		Header:     hdr("Server", "cloudflare", "cf-mitigated", "challenge"),
	}
	sessions := armFakeSession(challenged)
	r := askThirdDoor(t, walledRes(ownerLine, intPtr(403)), []*Proxy{nil}, true)
	if r.OK() {
		t.Fatal("must fail")
	}
	for _, want := range []string{ownerLine, "chrome-impersonated:",
		"HTTP 403 — bot defense (Cloudflare: challenge)", "needs a live browser"} {
		if !strings.Contains(r.Reason, want) {
			t.Fatalf("reason = %q (missing %q)", r.Reason, want)
		}
	}
	_ = sessions
}

func TestThirdDoorTwoRoutesTwoAsksProxyFirst(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	challenged := &fakeImpResponse{
		StatusCode: 403,
		Header:     hdr("Server", "cloudflare", "cf-mitigated", "challenge"),
	}
	sessions := armFakeSession(challenged, challenged)
	r := askThirdDoor(t, walledRes(ownerLine, intPtr(403)), []*Proxy{testProxy, nil}, true)
	if r.OK() {
		t.Fatal("must fail")
	}
	calls := (*sessions)[0].calls
	if len(calls) != 2 {
		t.Fatalf("calls = %d", len(calls))
	}
	// first ask through the proxy (socks5h — DNS at the proxy):
	if calls[0].ProxyURL != "socks5h://p.example.net:10808" {
		t.Fatalf("first call proxy = %q", calls[0].ProxyURL)
	}
	// second ask direct:
	if calls[1].ProxyURL != "" {
		t.Fatalf("second call proxy = %q", calls[1].ProxyURL)
	}
	// one honest line, both legs:
	if strings.Count(r.Reason, "chrome-impersonated:") != 1 {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestThirdDoorPlain403ReportsCleanly(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 403})
	r := askThirdDoor(t, walledRes("proxy: HTTP 403 | direct: HTTP 403", intPtr(403)), []*Proxy{nil}, true)
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, "| chrome-impersonated: HTTP 403") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if strings.Contains(r.Reason, "needs a live browser") {
		t.Fatalf("reason = %q (no challenge marker → no hint)", r.Reason)
	}
	_ = sessions
}

// ---------------------------------------------------------------------------
// fetch_url's wiring — the routes the CALLER allows are never widened
// (test_thirddoor.py's TestFetchUrlWiring, mock-level)
// ---------------------------------------------------------------------------

func TestThirdDoorDirectOptOutConfinesToProxy(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	deadProxy := &Proxy{Type: "socks5", Host: "127.0.0.1", Port: 1}
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200})
	defer func() { _ = sessions }()
	var gotRoutes []*Proxy
	realMaybe := maybeImpersonateDirect
	maybeImpersonateDirect = func(url string, timeoutS float64, maxBytes int, routes []*Proxy, enabled bool, walled *FetchResult, started time.Time, autopip bool) *FetchResult {
		gotRoutes = routes
		return realMaybe(url, timeoutS, maxBytes, routes, enabled, walled, started, autopip)
	}
	defer func() { maybeImpersonateDirect = realMaybe }()
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	}
	FetchURL("http://walled.example.net/page", &Options{TimeoutS: 8, Proxy: deadProxy, NoDirectFallback: true})
	if len(gotRoutes) != 1 || gotRoutes[0] != deadProxy {
		t.Fatalf("routes = %v (proxy-only — an opt-out is never widened)", gotRoutes)
	}
	_ = sessions
}

func TestThirdDoorBothDoorsWalledGetsBothRoutes(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	deadProxy := &Proxy{Type: "socks5", Host: "127.0.0.1", Port: 1}
	sessions := armFakeSession(&fakeImpResponse{StatusCode: 200}, &fakeImpResponse{StatusCode: 200})
	var gotRoutes []*Proxy
	realMaybe := maybeImpersonateDirect
	maybeImpersonateDirect = func(url string, timeoutS float64, maxBytes int, routes []*Proxy, enabled bool, walled *FetchResult, started time.Time, autopip bool) *FetchResult {
		gotRoutes = routes
		return realMaybe(url, timeoutS, maxBytes, routes, enabled, walled, started, autopip)
	}
	defer func() { maybeImpersonateDirect = realMaybe }()
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		if p != nil {
			return res("connection: proxy 127.0.0.1:1 connection failed: refused", nil, "proxy_error", nil)
		}
		return res("HTTP 403 — bot defense (server: cloudflare)", intPtr(403), "blocked_bot", nil)
	}
	FetchURL("http://walled.example.net/page", &Options{TimeoutS: 8, Proxy: deadProxy})
	if len(gotRoutes) != 2 || gotRoutes[0] != deadProxy || gotRoutes[1] != nil {
		t.Fatalf("routes = %v (both, proxy first)", gotRoutes)
	}
	_ = sessions
}

// ---------------------------------------------------------------------------
// the REAL library — the Go door ships armed, so the REAL transport is
// the default one: a local listener proves the full FetchResult shape.
// (test_thirddoor.py's TestRealCurlCffi — CI installs curl_cffi there;
// here the library is compiled in.)
// ---------------------------------------------------------------------------

func TestRealImpersonatedFetchShapesResult(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	body := []byte("<html><head><title>Real Imp Page</title></head><body>handshake ok</body></html>")
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", itoa(len(body)))
		w.WriteHeader(200)
		_, _ = w.Write(body)
	})
	srv := startServer(t, mux)
	r := impersonatedFetchOnce(srv.root()+"/p", 5, 2_000_000, nil, time.Now())
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "full" || r.HTTPStatus == nil || *r.HTTPStatus != 200 {
		t.Fatalf("status=%q http=%v reason=%q", r.Status, r.HTTPStatus, r.Reason)
	}
	if !strings.Contains(string(r.Body), "handshake ok") {
		t.Fatalf("body = %q", string(r.Body))
	}
	if !strings.Contains(r.Text, "Real Imp Page") {
		t.Fatalf("text = %q", r.Text)
	}
	if r.Charset != "utf-8" {
		t.Fatalf("charset = %q", r.Charset)
	}
	if r.ElapsedS < 0 {
		t.Fatalf("elapsed = %v", r.ElapsedS)
	}
}

func TestRealImpersonatedFetchGzipAutoDecoded(t *testing.T) {
	// libcurl decodes what it asked for: the REAL session's body arrives
	// decompressed, so the cap bounds DEcompressed bytes — the same
	// budget law as the stdlib door.
	state := saveDoorState()
	defer state.restore()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		writeCompressed(t, w, "gzip", "<html><head><title>Gz Imp</title></head><body>"+strings.Repeat("payload ", 100)+"</body></html>")
	})
	srv := startServer(t, mux)
	r := impersonatedFetchOnce(srv.root()+"/gz", 5, 2_000_000, nil, time.Now())
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if !strings.Contains(r.Text, "Gz Imp") || !strings.Contains(r.Text, "payload") {
		t.Fatalf("text = %q (the gzip body must arrive decoded)", r.Text)
	}
}

func TestRealImpersonatedFetchCapped(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body := []byte(strings.Repeat("word ", 40000)) // 200 KB
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Content-Length", itoa(len(body)))
		w.WriteHeader(200)
		_, _ = w.Write(body)
	})
	srv := startServer(t, mux)
	r := impersonatedFetchOnce(srv.root()+"/big", 5, 50_000, nil, time.Now())
	if r.Status != "partial" || !strings.Contains(r.Reason, "size cap") {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if len(r.Body) != 50_000 {
		t.Fatalf("body = %d bytes", len(r.Body))
	}
}

func TestRealImpersonatedFetchHTTPErrors(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Server", "cloudflare")
		w.Header().Set("cf-mitigated", "challenge")
		w.WriteHeader(403)
	})
	srv := startServer(t, mux)
	r := impersonatedFetchOnce(srv.root()+"/walled", 5, 2_000_000, nil, time.Now())
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 403 {
		t.Fatalf("http = %v", r.HTTPStatus)
	}
	if r.Reason != "HTTP 403 — bot defense (Cloudflare: challenge)" {
		t.Fatalf("reason = %q", r.Reason)
	}
}

var _ = io.Discard
