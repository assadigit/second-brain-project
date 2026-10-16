package connectioncheck

// fakes_test.go — the T-33 hermetic rig: a recording JSON-route
// server (the reference suites' stdlib http.server fake) that captures
// method, path, User-Agent, Authorization, Accept and body length for
// every request — the UA table and the read-only law are provable from
// it — plus the guaranteed-dead loopback URL helper. No network
// beyond 127.0.0.1 (the Phase-3 exit-gate law).

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// capturedReq is one recorded request (the UA table's evidence).
type capturedReq struct {
	Method string
	Path   string
	UA     string
	Auth   string
	Accept string
	BodyN  int
}

// fakeJSONServer is _start_json_server: GET routes from
// {path: (code, payload)}; unknown paths 404 {"message": "Not Found"};
// every request recorded.
type fakeJSONServer struct {
	srv      *http.Server
	listener net.Listener
	mu       sync.Mutex
	requests []capturedReq
	routes   map[string][2]any
}

// route builds a (code, payload) tuple for the routes map.
func route(code int, payload any) [2]any { return [2]any{code, payload} }

func newJSONServer(t *testing.T, routes map[string][2]any) *fakeJSONServer {
	t.Helper()
	fs := &fakeJSONServer{routes: routes}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		fs.mu.Lock()
		fs.requests = append(fs.requests, capturedReq{
			Method: r.Method,
			Path:   strings.Split(r.URL.Path, "?")[0],
			UA:     r.Header.Get("User-Agent"),
			Auth:   r.Header.Get("Authorization"),
			Accept: r.Header.Get("Accept"),
			BodyN:  len(body),
		})
		routes := fs.routes
		fs.mu.Unlock()
		path := strings.Split(r.URL.Path, "?")[0]
		rt, ok := routes[path]
		if !ok {
			writeFakeJSON(w, 404, map[string]any{"message": "Not Found"})
			return
		}
		writeFakeJSON(w, rt[0].(int), rt[1])
	})
	fs.srv = &http.Server{Handler: mux}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fs.listener = listener
	go func() { _ = fs.srv.Serve(listener) }()
	t.Cleanup(func() { _ = fs.srv.Close() })
	return fs
}

func writeFakeJSON(w http.ResponseWriter, code int, payload any) {
	data := []byte(storage.PyDumpsCompact(payload))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func (fs *fakeJSONServer) root() string {
	return "http://127.0.0.1:" + strconv.Itoa(fs.port())
}

func (fs *fakeJSONServer) port() int {
	return fs.listener.Addr().(*net.TCPAddr).Port
}

func (fs *fakeJSONServer) captured() []capturedReq {
	fs.mu.Lock()
	defer fs.mu.Unlock()
	out := make([]capturedReq, len(fs.requests))
	copy(out, fs.requests)
	return out
}

// deadURL is a loopback URL that is guaranteed to refuse connections
// (the reference's _dead_url).
func deadURL(t *testing.T) string {
	t.Helper()
	s, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	u := "http://127.0.0.1:" + strconv.Itoa(s.Addr().(*net.TCPAddr).Port)
	_ = s.Close()
	return u
}

// withSeam swaps a package seam for one test and restores it.
func withSeam[T any](t *testing.T, slot *T, fake T) {
	t.Helper()
	old := *slot
	*slot = fake
	t.Cleanup(func() { *slot = old })
}

// withProbes swaps the sentinel ladder's fixed endpoints (the Python
// suite monkeypatches urllib; the Go suite redirects the ladder at
// loopback fakes — same hermeticity, real HTTP on 127.0.0.1 only).
func withProbes(t *testing.T, github, generic string) {
	t.Helper()
	oldG, oldC := githubProbeURL, genericProbeURL
	githubProbeURL, genericProbeURL = github, generic
	t.Cleanup(func() { githubProbeURL, genericProbeURL = oldG, oldC })
}
