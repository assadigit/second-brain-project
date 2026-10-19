// fakes_test.go — the T-31 fake-server rig: the local net/http listeners
// that mirror the Python suites' stdlib http.server fakes (the ladder
// server from test_ladder.py, the phase-2 server from test_phase2.py,
// the wire recorder, the plain HTTP proxy, and the REAL SOCKS5 server
// from test_webproxy.py that proves the rdns law), plus the fake
// limiter, the fake impersonation session, the case-insensitive header
// rig, and the shared seam save/restore helpers. No network beyond
// 127.0.0.1 — the Phase-3 exit-gate law.
package webfetch

import (
	"bytes"
	"compress/gzip"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// ---------------------------------------------------------------------------
// the _Headers rig (test_ladder.py): a plain map with case-insensitive
// Get — satisfies the package's headerGetter structurally.
// ---------------------------------------------------------------------------

// mapHeaders is the _Headers rig: a plain map with case-insensitive Get.
type mapHeaders map[string]string

func (m mapHeaders) Get(key string) string {
	for k, v := range m {
		if strings.EqualFold(k, key) {
			return v
		}
	}
	return ""
}

func hdr(pairs ...string) mapHeaders {
	out := mapHeaders{}
	for i := 0; i+1 < len(pairs); i += 2 {
		out[pairs[i]] = pairs[i+1]
	}
	return out
}

// ---------------------------------------------------------------------------
// The ladder server — test_ladder.py's _LadderHandler.
// ---------------------------------------------------------------------------

type ladderServer struct {
	srv      *http.Server
	listener net.Listener
}

func newLadderServer(t *testing.T) *ladderServer {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/article":
			body := []byte("<html><head><title>Practical Typography Guide" +
				"</title></head><body>the article</body></html>")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeader(200)
			_, _ = w.Write(body)
		case r.URL.Path == "/hop-cookie":
			w.Header().Set("Location", "/hop-cookie-target")
			w.Header().Add("Set-Cookie", "gate=open; Path=/")
			w.WriteHeader(307)
		case r.URL.Path == "/hop-cookie-target":
			if strings.Contains(r.Header.Get("Cookie"), "gate=open") {
				body := []byte("<html><head><title>Behind the gate</title></head></html>")
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.Header().Set("Content-Length", strconv.Itoa(len(body)))
				w.WriteHeader(200)
				_, _ = w.Write(body)
			} else {
				w.Header().Set("Content-Length", "0")
				w.WriteHeader(403)
			}
		case r.URL.Path == "/no-location":
			// A redirect answer with no Location header at all.
			w.WriteHeader(307)
		case r.URL.Path == "/loop":
			w.Header().Set("Location", "/loop")
			w.WriteHeader(302)
		case strings.HasPrefix(r.URL.Path, "/chain/"):
			n, _ := strconv.Atoi(r.URL.Path[strings.LastIndex(r.URL.Path, "/")+1:])
			w.Header().Set("Location", fmt.Sprintf("/chain/%d", n+1))
			w.WriteHeader(302)
		case r.URL.Path == "/moved-308":
			w.Header().Set("Location", "/article")
			w.WriteHeader(308)
		default:
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(404)
		}
	})
	return startServer(t, mux)
}

func startServer(t *testing.T, handler http.Handler) *ladderServer {
	t.Helper()
	ls := &ladderServer{}
	ls.srv = &http.Server{Handler: handler}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ls.listener = listener
	go func() { _ = ls.srv.Serve(listener) }()
	t.Cleanup(func() { _ = ls.srv.Close() })
	return ls
}

func (ls *ladderServer) root() string {
	return "http://127.0.0.1:" + strconv.Itoa(ls.port())
}

func (ls *ladderServer) port() int {
	return ls.listener.Addr().(*net.TCPAddr).Port
}

// ---------------------------------------------------------------------------
// The phase-2 server — test_phase2.py's _Handler (the web_fetch legs).
// ---------------------------------------------------------------------------

func newPhase2Server(t *testing.T) *ladderServer {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		path := strings.Split(r.URL.Path, "?")[0]
		switch path {
		case "/slow":
			time.Sleep(8 * time.Second)
			w.WriteHeader(200)
		case "/headers":
			var lines []string
			for k, v := range r.Header {
				lines = append(lines, fmt.Sprintf("%s: %s", k, strings.Join(v, ", ")))
			}
			body := []byte(strings.Join(lines, "\n"))
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeader(200)
			_, _ = w.Write(body)
		case "/gzip":
			writeCompressed(t, w, "gzip", "<html><head><title>Gzip Page</title></head><body>"+
				"<p>gzipped content about design tools and assets</p></body></html>")
		case "/deflate":
			writeCompressed(t, w, "deflate", "<html><head><title>Deflate Page</title></head><body>"+
				"<p>deflated content for the reading list</p></body></html>")
		case "/gzip-bomb":
			writeCompressed(t, w, "gzip", strings.Repeat("bomb ", 100_000))
		case "/blocked":
			w.Header().Set("Server", "cloudflare")
			w.Header().Set("Date", "Tue, 07 Oct 2026 00:00:00 GMT")
			w.WriteHeader(403)
		case "/redirect":
			w.Header().Set("Location", "/article")
			w.WriteHeader(301)
		case "/redirect-loop":
			w.Header().Set("Location", "/redirect-loop")
			w.WriteHeader(302)
		case "/pdf":
			body := []byte("%PDF-1.4 fake pdf bytes for the fixture")
			w.Header().Set("Content-Type", "application/pdf")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeader(200)
			_, _ = w.Write(body)
		case "/latin1":
			body := []byte("<html><head><title>Caf\xe9 Dise\xf1o</title></head>" +
				"<body><p>P\xe1gina sobre dise\xf1o gr\xe1fico y " +
				"tipograf\xeda editorial para el proyecto.</p></body></html>")
			w.Header().Set("Content-Type", "text/html; charset=windows-1252")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeader(200)
			_, _ = w.Write(body)
		case "/huge":
			chunk := strings.Repeat("<p>"+strings.Repeat("word ", 40)+"</p>\n", 2000)
			body := []byte("<html><head><title>Huge</title></head><body>" +
				chunk + "</body></html>")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeader(200)
			_, _ = w.Write(body)
		case "/article":
			body := []byte("<html><head><title>Practical Typography Guide" +
				"</title></head><body>the article</body></html>")
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Header().Set("Content-Length", strconv.Itoa(len(body)))
			w.WriteHeader(200)
			_, _ = w.Write(body)
		default:
			w.Header().Set("Content-Length", "0")
			w.WriteHeader(404)
		}
	})
	return startServer(t, mux)
}

// ---------------------------------------------------------------------------
// The wire recorder — the generator's WireRecorder: ordered header pairs.
// ---------------------------------------------------------------------------

type wireRecord struct {
	Method  string
	Path    string
	Headers [][2]string
}

type wireServer struct {
	srv      *http.Server
	listener net.Listener
	mu       sync.Mutex
	requests []wireRecord
	routes   map[string]wireRoute
}

type wireRoute struct {
	code    int
	payload []byte
	headers [][2]string
}

func newWireServer(t *testing.T, routes map[string]wireRoute) *wireServer {
	t.Helper()
	ws := &wireServer{routes: routes}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		var pairs [][2]string
		for k, v := range r.Header {
			pairs = append(pairs, [2]string{k, strings.Join(v, ", ")})
		}
		// http.Header is a map — preserve a deterministic order for the
		// recording: the ARRIVAL order is not recoverable from net/http,
		// so the recorder sorts (the ORDER pin lives in the pure golden's
		// RequestHeaders pairs, which are the app's own construction
		// order; the wire driver compares sets).
		sortWirePairs(pairs)
		ws.mu.Lock()
		ws.requests = append(ws.requests, wireRecord{
			Method:  r.Method,
			Path:    r.RequestURI, // the raw target: absolute-form on proxy requests
			Headers: pairs,
		})
		routes := ws.routes
		ws.mu.Unlock()
		path := strings.Split(r.URL.Path, "?")[0]
		route, ok := routes[path]
		if !ok {
			w.WriteHeader(404)
			return
		}
		for _, h := range route.headers {
			w.Header().Set(h[0], h[1])
		}
		if route.payload != nil {
			w.Header().Set("Content-Length", strconv.Itoa(len(route.payload)))
		}
		w.WriteHeader(route.code)
		if route.payload != nil {
			_, _ = w.Write(route.payload)
		}
	})
	ws.srv = &http.Server{Handler: mux}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	ws.listener = listener
	go func() { _ = ws.srv.Serve(listener) }()
	t.Cleanup(func() { _ = ws.srv.Close() })
	return ws
}

func sortWirePairs(pairs [][2]string) {
	for i := 1; i < len(pairs); i++ {
		for j := i; j > 0 && pairs[j][0] < pairs[j-1][0]; j-- {
			pairs[j], pairs[j-1] = pairs[j-1], pairs[j]
		}
	}
}

func (ws *wireServer) root() string {
	return "http://127.0.0.1:" + strconv.Itoa(ws.port())
}

func (ws *wireServer) port() int {
	return ws.listener.Addr().(*net.TCPAddr).Port
}

func (ws *wireServer) captured() []wireRecord {
	ws.mu.Lock()
	defer ws.mu.Unlock()
	out := make([]wireRecord, len(ws.requests))
	copy(out, ws.requests)
	return out
}

// ---------------------------------------------------------------------------
// The REAL SOCKS5 server — test_webproxy.py's _FakeSocks5Server: no-auth
// greeting, CONNECT with the DOMAINNAME address type, then a canned HTTP
// response through the tunnel. Records what the client asked it to reach
// — the rdns proof: a correct client sends the HOSTNAME (never resolves
// locally).
// ---------------------------------------------------------------------------

type socksSeen struct {
	Host        string
	Port        int
	DNSAtProxy  bool
	RequestLine string
	Err         string
}

type fakeSocks5Server struct {
	body   []byte
	mu     sync.Mutex
	seen   []socksSeen
	srv    net.Listener
	port   int
	stopCh chan struct{}
	wg     sync.WaitGroup
}

func newFakeSocks5Server(t *testing.T, body []byte) *fakeSocks5Server {
	t.Helper()
	if body == nil {
		body = []byte("<html><head><title>Tunnel</title></head><body>ok</body></html>")
	}
	fs := &fakeSocks5Server{body: body, stopCh: make(chan struct{})}
	srv, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fs.srv = srv
	fs.port = srv.Addr().(*net.TCPAddr).Port
	fs.wg.Add(1)
	go fs.loop()
	t.Cleanup(fs.stop)
	return fs
}

func (fs *fakeSocks5Server) loop() {
	defer fs.wg.Done()
	for {
		conn, err := fs.srv.Accept()
		if err != nil {
			return
		}
		fs.wg.Add(1)
		go fs.handle(conn)
	}
}

func (fs *fakeSocks5Server) handle(conn net.Conn) {
	defer fs.wg.Done()
	defer func() { _ = conn.Close() }()
	rec := socksSeen{}
	fs.mu.Lock()
	fs.seen = append(fs.seen, rec)
	idx := len(fs.seen) - 1
	fs.mu.Unlock()
	fail := func(msg string) {
		fs.mu.Lock()
		rec.Err = msg
		fs.seen[idx] = rec
		fs.mu.Unlock()
	}
	_ = conn.SetDeadline(time.Now().Add(10 * time.Second))
	greeting := make([]byte, 64)
	n, err := conn.Read(greeting)
	if err != nil || n == 0 || greeting[0] != 5 {
		fail(fmt.Sprintf("bad greeting %q", greeting[:n]))
		return
	}
	_, _ = conn.Write([]byte{5, 0}) // no auth required
	req := make([]byte, 512)
	n, err = conn.Read(req)
	if err != nil || n < 8 || req[0] != 5 || req[1] != 1 {
		fail(fmt.Sprintf("not a CONNECT request %q", req[:n]))
		return
	}
	atyp := req[3]
	switch atyp {
	case 3: // domainname — the rdns proof
		ln := int(req[4])
		rec.Host = string(req[5 : 5+ln])
		rec.Port = int(req[5+ln])<<8 | int(req[6+ln])
		rec.DNSAtProxy = true
	case 1: // client resolved locally
		rec.Host = net.IP(req[4:8]).String()
		rec.Port = int(req[8])<<8 | int(req[9])
		rec.DNSAtProxy = false
	default:
		fail(fmt.Sprintf("unexpected ATYP %d", atyp))
		return
	}
	_, _ = conn.Write([]byte{5, 0, 0, 1, 0, 0, 0, 0, 0, 0}) // success
	httpReq := make([]byte, 4096)
	n, err = conn.Read(httpReq)
	if err == nil && n > 0 {
		rec.RequestLine = strings.SplitN(string(httpReq[:n]), "\r\n", 2)[0]
	}
	// Publish the record BEFORE the response write: the client can
	// observe the response (client.Get returns on the headers alone)
	// and call records() the instant the write lands. Publishing after
	// the write left a TOCTOU that a loaded 2-CPU box surfaced at
	// SESSION-033's recovery battery — the assertions are unchanged.
	fs.mu.Lock()
	fs.seen[idx] = rec
	fs.mu.Unlock()
	resp := []byte("HTTP/1.1 200 OK\r\n" +
		"Content-Type: text/html; charset=utf-8\r\n" +
		"Content-Length: " + strconv.Itoa(len(fs.body)) + "\r\n\r\n")
	_, _ = conn.Write(append(resp, fs.body...))
}

func (fs *fakeSocks5Server) proxy() *Proxy {
	return &Proxy{Type: "socks5", Host: "127.0.0.1", Port: fs.port}
}

func (fs *fakeSocks5Server) records() []socksSeen {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	out := make([]socksSeen, len(fs.seen))
	copy(out, fs.seen)
	return out
}

func (fs *fakeSocks5Server) stop() {
	close(fs.stopCh)
	_ = fs.srv.Close()
	fs.wg.Wait()
}

// deadPort yields a port that NOTHING listens on (the _DeadPort rig).
func deadPort(t *testing.T) int {
	t.Helper()
	s, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := s.Addr().(*net.TCPAddr).Port
	_ = s.Close()
	return port
}

// alivePort yields a port that IS listening (TCP only — the _AliveSocket
// rig).
func alivePort(t *testing.T) int {
	t.Helper()
	s, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s.Addr().(*net.TCPAddr).Port
}

// ---------------------------------------------------------------------------
// The fake limiter — records politeness asks and penalties; never sleeps.
// ---------------------------------------------------------------------------

type fakeLimiter struct {
	waits     []string
	penalties []penaltyRecord
}

type penaltyRecord struct {
	domain  string
	seconds float64
}

func (f *fakeLimiter) Wait(domain string) { f.waits = append(f.waits, domain) }
func (f *fakeLimiter) Penalize(domain string, s float64) {
	f.penalties = append(f.penalties, penaltyRecord{domain, s})
}

// ---------------------------------------------------------------------------
// The fake impersonation session — the fake-session pattern one level
// down: the factory is package-level for exactly this (test_thirddoor).
// ---------------------------------------------------------------------------

type fakeImpResponse struct {
	StatusCode int
	Body       []byte
	Header     mapHeaders
	URL        string
	Charset    string
}

func (r *fakeImpResponse) Get(key string) string { return r.Header.Get(key) }

type fakeImpSession struct {
	responses []*fakeImpResponse
	calls     []impCall
}

type impCall struct {
	URL      string
	ProxyURL string
	TimeoutS float64
}

func (s *fakeImpSession) Get(url string, timeoutS float64, proxyURL string) (impResponse, error) {
	s.calls = append(s.calls, impCall{URL: url, ProxyURL: proxyURL, TimeoutS: timeoutS})
	// The Python FakeImpResponse defaults: body "<html><body>ok</body>
	// </html>", Content-Type text/html; charset=utf-8, charset utf-8.
	body := []byte("<html><body>ok</body></html>")
	header := hdr("Content-Type", "text/html; charset=utf-8")
	charset := "utf-8"
	status := 403
	if len(s.responses) > 0 {
		r := s.responses[0]
		s.responses = s.responses[1:]
		status = r.StatusCode
		if r.Body != nil {
			body = r.Body
		}
		if r.Header != nil {
			header = r.Header
		}
		if r.Charset != "" {
			charset = r.Charset
		}
		return impResponse{
			StatusCode: status,
			Header:     header,
			URL:        r.URL,
			Charset:    charset,
			Body:       io.NopCloser(bytes.NewReader(body)),
		}, nil
	}
	return impResponse{
		StatusCode: status,
		Header:     header,
		Body:       io.NopCloser(bytes.NewReader(nil)),
	}, nil
}

// ---------------------------------------------------------------------------
// The door-state save/restore — every seam the Python suites flip, in
// one place (the setUp/tearDown contract of _ThirdDoorCase et al.).
// ---------------------------------------------------------------------------

type doorState struct {
	fetchOnce    fetchOnceFn
	archive      archiveFn
	doh          dohFn
	available    func() bool
	pipInstall   func() bool
	factory      func() impSessionIface
	session      impSessionIface
	autopipTried bool
	autopipOK    bool
}

func saveDoorState() doorState {
	impMu.Lock()
	defer impMu.Unlock()
	dohMu.Lock()
	defer dohMu.Unlock()
	return doorState{
		fetchOnce:    fetchOnceVar,
		archive:      archiveLookupVar,
		doh:          dohProbeOnceVar,
		available:    impersonatorAvailable,
		pipInstall:   pipInstallImpersonator,
		factory:      newImpersonationSession,
		session:      impSessionCached,
		autopipTried: autopipTried,
		autopipOK:    autopipOK,
	}
}

func (d doorState) restore() {
	impMu.Lock()
	defer impMu.Unlock()
	dohMu.Lock()
	defer dohMu.Unlock()
	fetchOnceVar = d.fetchOnce
	archiveLookupVar = d.archive
	dohProbeOnceVar = d.doh
	impersonatorAvailable = d.available
	pipInstallImpersonator = d.pipInstall
	newImpersonationSession = d.factory
	impSessionCached = d.session
	autopipTried = d.autopipTried
	autopipOK = d.autopipOK
	dohCache = map[string]string{} // inline: dohMu is already held
	ResetCookieJar()
}

// armFakeSession swaps the factory for a recording fake; the returned
// pointer-to-slice mirrors the Python rig's self.sessions (a pointer:
// the factory's appends must be visible to the caller).
func armFakeSession(responses ...*fakeImpResponse) *[]*fakeImpSession {
	sessions := &[]*fakeImpSession{}
	impMu.Lock()
	defer impMu.Unlock()
	impSessionCached = nil
	newImpersonationSession = func() impSessionIface {
		s := &fakeImpSession{responses: responses}
		*sessions = append(*sessions, s)
		return s
	}
	return sessions
}

func armAvailable(ok bool) {
	impMu.Lock()
	defer impMu.Unlock()
	impersonatorAvailable = func() bool { return ok }
}

// ---------------------------------------------------------------------------
// Shared tiny helpers for the batteries.
// ---------------------------------------------------------------------------

func intPtr(i int) *int { return &i }

func floatPtr(f float64) *float64 { return &f }

// res builds a failed FetchResult (the _res rig).
func res(reason string, httpStatus *int, category string, retryAfter *float64) *FetchResult {
	return &FetchResult{
		URL:         "http://walled.example.net/page",
		FinalURL:    "http://walled.example.net/page",
		Status:      "failed",
		Reason:      reason,
		HTTPStatus:  httpStatus,
		Category:    category,
		RetryAfterS: retryAfter,
	}
}

// okRes builds a full FetchResult (the _ok rig).
func okRes(url string, body []byte) *FetchResult {
	return &FetchResult{
		URL:         url,
		FinalURL:    url,
		Status:      "full",
		HTTPStatus:  intPtr(200),
		ContentType: "text/html",
		Charset:     "utf-8",
		Body:        body,
		Text:        string(body),
	}
}

// scriptFetchOnce — the scripted _fetch_once recorder: the Python
// suites' route scripting, one closure per shape. installs a fake fetchOnce that records the route
// (nil = direct) per call.
func scriptFetchOnce(fake fetchOnceFn) *[]*Proxy {
	calls := &[]*Proxy{}
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, userAgent string, proxy *Proxy, started time.Time) *FetchResult {
		*calls = append(*calls, proxy)
		return fake(url, timeoutS, maxBytes, userAgent, proxy, started)
	}
	return calls
}

// jsonBytes — compact JSON for the fake replies.
func jsonBytes(v any) []byte {
	data, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return data
}

// writeCompressed — the phase-2 fake's compressed-body routes (the
// Python _Handler's /gzip /deflate /gzip-bomb legs).
func writeCompressed(t *testing.T, w http.ResponseWriter, encoding, payload string) {
	t.Helper()
	var data []byte
	switch encoding {
	case "gzip":
		var buf bytes.Buffer
		zw := gzip.NewWriter(&buf)
		if _, err := zw.Write([]byte(payload)); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		data = buf.Bytes()
	case "deflate":
		var buf bytes.Buffer
		zw := zlib.NewWriter(&buf)
		if _, err := zw.Write([]byte(payload)); err != nil {
			t.Fatal(err)
		}
		if err := zw.Close(); err != nil {
			t.Fatal(err)
		}
		data = buf.Bytes()
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Content-Encoding", encoding)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(200)
	_, _ = w.Write(data)
}
