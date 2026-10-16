// socks_test.go — the proxy pipes, ported from test_webproxy.py:
// TestSocksPipe (the REAL fake-SOCKS5 server: full fetch through the
// tunnel, THE RDNS PROOF — ATYP=DOMAIN, the dead proxy, loopback never
// proxied, the preflight), TestHttpProxyAndOpeners (the absolute-URI
// http-type proxy wire + the route-builder laws).
package webfetch

import (
	"net/http"
	"strconv"
	"strings"
	"testing"
	"time"
)

func TestSocksPipeFetchFullThroughSocks5(t *testing.T) {
	fs := newFakeSocks5Server(t, nil)
	r := FetchURL("http://example.test/page", &Options{TimeoutS: 8, Proxy: fs.proxy()})
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 200 {
		t.Fatalf("http=%v", r.HTTPStatus)
	}
	if !strings.Contains(string(r.Body), "Tunnel") {
		t.Fatalf("body = %q", string(r.Body))
	}
	seen := fs.records()
	if len(seen) != 1 {
		t.Fatalf("server saw %d connections", len(seen))
	}
	if seen[0].Host != "example.test" || seen[0].Port != 80 {
		t.Fatalf("asked for %s:%d", seen[0].Host, seen[0].Port)
	}
	if !strings.HasPrefix(seen[0].RequestLine, "GET /page ") {
		t.Fatalf("request line = %q", seen[0].RequestLine)
	}
}

func TestSocksPipeDNSResolvedAtTheProxy(t *testing.T) {
	// The rdns contract: the client must send the HOSTNAME (ATYP=3).
	// example.test does not exist in DNS — a local resolve would fail.
	fs := newFakeSocks5Server(t, nil)
	r := FetchURL("http://example.test/rdns", &Options{TimeoutS: 8, Proxy: fs.proxy()})
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	seen := fs.records()
	if len(seen) == 0 || !seen[0].DNSAtProxy {
		t.Fatalf("dns_at_proxy not proven: %+v", seen)
	}
}

func TestSocksPipeDeadProxyIsClearFailedResult(t *testing.T) {
	port := deadPort(t)
	r := FetchURL("http://example.test/x", &Options{TimeoutS: 4,
		Proxy: &Proxy{Type: "socks5", Host: "127.0.0.1", Port: port}})
	if r.Status != "failed" {
		t.Fatalf("status=%q", r.Status)
	}
	if !strings.Contains(strings.ToLower(r.Reason), "proxy") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if !strings.Contains(r.Reason, "127.0.0.1:"+itoaTest(port)) {
		t.Fatalf("reason = %q (must name host:port)", r.Reason)
	}
	if r.Category != "proxy_error" {
		t.Fatalf("category = %q", r.Category)
	}
}

func TestSocksPipeLoopbackNeverRidesTheProxy(t *testing.T) {
	// v0.15.1 rule: loopback targets go DIRECT even when a proxy is
	// configured — proven with a SEPARATE plain HTTP server on loopback:
	// the request must land there verbatim, and the fake SOCKS5 server
	// must see NOTHING new.
	fs := newFakeSocks5Server(t, nil)
	seenPlain := map[string]string{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		seenPlain["request_line"] = r.RequestURI
		seenPlain["method"] = r.Method
		body := []byte("<html><head><title>Local</title></head></html>")
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Content-Length", itoaTest(len(body)))
		w.WriteHeader(200)
		_, _ = w.Write(body)
	})
	plain := startServer(t, mux)
	socksSeenBefore := len(fs.records())
	r := FetchURL(plain.root()+"/loop", &Options{TimeoutS: 8, Proxy: fs.proxy()})
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if !strings.Contains(string(r.Body), "Local") {
		t.Fatalf("body = %q", string(r.Body))
	}
	if seenPlain["request_line"] != "/loop" || seenPlain["method"] != "GET" {
		t.Fatalf("plain server saw %v", seenPlain)
	}
	if len(fs.records()) != socksSeenBefore {
		t.Fatalf("the SOCKS server saw %d new connections",
			len(fs.records())-socksSeenBefore)
	}
}

func TestPreflightAliveAndDead(t *testing.T) {
	fs := newFakeSocks5Server(t, nil)
	ok, why := WebProxyPreflight(fs.proxy(), 2)
	if !ok || why != "" {
		t.Fatalf("alive: %v %q", ok, why)
	}
	port := deadPort(t)
	ok, why = WebProxyPreflight(&Proxy{Type: "socks5", Host: "127.0.0.1", Port: port}, 1)
	if ok {
		t.Fatal("dead port must fail")
	}
	if !strings.Contains(why, "127.0.0.1:"+itoaTest(port)) || !strings.Contains(why, "unreachable") {
		t.Fatalf("why = %q", why)
	}
}

func TestPreflightNone(t *testing.T) {
	ok, why := WebProxyPreflight(nil, 2)
	if ok || why != "no proxy configured" {
		t.Fatalf("none: %v %q", ok, why)
	}
}

// ---------------------------------------------------------------------------
// the http-type proxy wire (absolute URI) + the route builders
// ---------------------------------------------------------------------------

func TestHTTPProxyAbsoluteURI(t *testing.T) {
	// The http-type proxy rides the transport's ProxyURL: the request
	// line carries the FULL absolute URI (proxy-side DNS — the same
	// protection as rdns=True).
	seen := map[string]string{}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		seen["request_uri"] = r.RequestURI
		body := []byte("<html><head><title>Via HTTP proxy</title></head></html>")
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.Header().Set("Content-Length", itoaTest(len(body)))
		w.WriteHeader(200)
		_, _ = w.Write(body)
	})
	srv := startServer(t, mux)
	ResetCookieJar()
	r := FetchURL("http://example.test/via-proxy", &Options{TimeoutS: 8,
		Proxy: &Proxy{Type: "http", Host: "127.0.0.1", Port: srv.port()}})
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if !strings.Contains(string(r.Body), "Via HTTP proxy") {
		t.Fatalf("body = %q", string(r.Body))
	}
	if seen["request_uri"] != "http://example.test/via-proxy" {
		t.Fatalf("request URI = %q (must be absolute)", seen["request_uri"])
	}
}

func TestBuildRouteVariants(t *testing.T) {
	// The route-builder law (the Go shape of _build_opener's contract):
	// direct/loopback routes never proxy (not even env), the http-type
	// route rides ProxyURL, the socks route rides the SOCKS dialer.
	direct, err := buildRoute(nil, false)
	if err != nil {
		t.Fatal(err)
	}
	if direct.transport.Proxy != nil {
		t.Fatal("the direct route must never proxy")
	}
	loop, err := buildRoute(&Proxy{Type: "socks5", Host: "127.0.0.1", Port: 1}, true)
	if err != nil {
		t.Fatal(err)
	}
	if loop.transport.Proxy != nil || loop.transport.DialContext != nil {
		t.Fatal("the forced-direct route must have no proxy and no socks dialer")
	}
	httpP, err := buildRoute(&Proxy{Type: "http", Host: "127.0.0.1", Port: 3128}, false)
	if err != nil {
		t.Fatal(err)
	}
	if httpP.transport.Proxy == nil {
		t.Fatal("the http-type route must ride ProxyURL")
	}
	socks, err := buildRoute(&Proxy{Type: "socks5", Host: "127.0.0.1", Port: 10808}, false)
	if err != nil {
		t.Fatal(err)
	}
	if socks.transport.Proxy != nil || socks.transport.DialContext == nil {
		t.Fatal("the socks route must ride the dialer, not ProxyURL")
	}
}

func TestSocksConnectSendsHostname(t *testing.T) {
	// The ATYP=DOMAIN proof lives in TestSocksPipeDNSResolvedAtTheProxy
	// (the full pipe); this leg pins the dialer's behavior through a
	// fresh socks5 route against the recording server.
	fs := newFakeSocks5Server(t, nil)
	route, err := buildRoute(fs.proxy(), false)
	if err != nil {
		t.Fatal(err)
	}
	client := &http.Client{
		Transport: route.transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: 8 * time.Second,
	}
	resp, err := client.Get("http://example.test/dialer")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	seen := fs.records()
	if len(seen) != 1 || !seen[0].DNSAtProxy || seen[0].Host != "example.test" {
		t.Fatalf("dialer records = %+v", seen)
	}
}

func itoaTest(i int) string { return strconv.Itoa(i) }
