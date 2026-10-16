// fakes_test.go — the T-30 fake-server rig: local net/http listeners
// that mirror the Python suites' stdlib http.server fakes (the recording
// routes server, the llama.cpp behavior server) plus the shared
// capture/normalization helpers the golden drivers use. No network
// beyond 127.0.0.1 — the Phase-3 exit-gate law.
package llmclient

import (
	"encoding/json"

	"github.com/assadigit/second-brain-project/internal/storage"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
)

// ---------------------------------------------------------------------------
// The recording server (routes: {path: (code, payload)} or a callable
// body → (code, payload)) — the generator's RecordingHandler shape.
// ---------------------------------------------------------------------------

type reqRecord struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
	Body    string            `json:"body"`
}

type recordingServer struct {
	srv      *http.Server
	listener net.Listener
	mu       sync.Mutex
	requests []reqRecord
	routes   map[string]any // route(code, payload) tuples or func(map[string]any) (int, any)
}

// appHeaderMap keeps only the app-chosen headers (lowercased) — the
// runtime-default User-Agent is dropped unless the app set its own
// (mirrors the generator's capture filter).
func appHeaderMap(h http.Header) map[string]string {
	keep := map[string]string{}
	for k, v := range h {
		if len(v) == 0 {
			continue
		}
		switch lk := strings.ToLower(k); lk {
		case "content-type", "authorization", "x-api-key", "anthropic-version":
			keep[lk] = v[0]
		case "user-agent":
			if v[0] != "Go-http-client/1.1" {
				keep[lk] = v[0]
			}
		}
	}
	return keep
}

// route builds a (code, payload) tuple for the routes maps — a
// readable one-liner instead of nested composite literals.
func route(code int, payload any) [2]any { return [2]any{code, payload} }

func writeJSONReply(w http.ResponseWriter, code int, payload any) {
	var data []byte
	if raw, ok := payload.([]byte); ok {
		data = raw
	} else {
		// The Python fakes reply with json.dumps(payload) — the
		// ", "/": " separator bytes are pinned by the goldens.
		data = []byte(storage.PyDumpsCompact(payload))
	}
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func newRecordingServer(t *testing.T, routes map[string]any) *recordingServer {
	t.Helper()
	rs := &recordingServer{routes: routes}
	rs.srv = &http.Server{Handler: recordingMux(rs)}
	serveLocal(t, rs.srv, &rs.listener)
	t.Cleanup(func() { _ = rs.srv.Close() })
	return rs
}

func recordingMux(rs *recordingServer) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		rs.mu.Lock()
		rs.requests = append(rs.requests, reqRecord{
			Method:  r.Method,
			Path:    strings.Split(r.URL.Path, "?")[0],
			Headers: appHeaderMap(r.Header),
			Body:    string(body),
		})
		routes := rs.routes
		rs.mu.Unlock()
		path := strings.Split(r.URL.Path, "?")[0]
		route, ok := routes[path]
		if !ok {
			writeJSONReply(w, 404, map[string]any{"error": "no such route"})
			return
		}
		var code int
		var payload any
		if fn, isFn := route.(func(map[string]any) (int, any)); isFn {
			parsed := map[string]any{}
			if len(body) > 0 {
				_ = json.Unmarshal(body, &parsed)
			}
			code, payload = fn(parsed)
		} else {
			tup := route.([2]any)
			code = tup[0].(int)
			payload = tup[1]
		}
		writeJSONReply(w, code, payload)
	})
	return mux
}

func serveLocal(t *testing.T, srv *http.Server, ln *net.Listener) {
	t.Helper()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	*ln = listener
	go func() { _ = srv.Serve(listener) }()
}

func (rs *recordingServer) root() string {
	return "http://127.0.0.1:" + strconv.Itoa(rs.port())
}

func (rs *recordingServer) port() int {
	return rs.listener.Addr().(*net.TCPAddr).Port
}

func (rs *recordingServer) captured() []reqRecord {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	out := make([]reqRecord, len(rs.requests))
	copy(out, rs.requests)
	return out
}

func (rs *recordingServer) reset() {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	rs.requests = nil
}

// ---------------------------------------------------------------------------
// The fake llama.cpp server — the test_llamacpp.py _FakeLlamaCppHandler
// shapes: behavior switches for props/health/models.
// ---------------------------------------------------------------------------

var defaultLlamaProps = map[string]any{
	"model_path":  "models/qwen2.5-3b-instruct-q4_k_m.gguf",
	"model_alias": "qwen2.5-3b",
	"total_slots": 1,
	"default_generation_settings": map[string]any{
		"temperature": 0.8, "n_predict": -1},
}

var llamaPropsWithLimits = map[string]any{
	"model_path":  "models/qwen2.5-3b-instruct-q4_k_m.gguf",
	"model_alias": "qwen2.5-3b",
	"total_slots": 1,
	"default_generation_settings": map[string]any{
		"n_ctx": 8192, "n_predict": 512, "temperature": 0.8},
}

type llamaCppServer struct {
	srv      *http.Server
	listener net.Listener
	mu       sync.Mutex
	behavior map[string]any
}

func newLlamaCppServer(t *testing.T, behavior map[string]any) *llamaCppServer {
	t.Helper()
	ls := &llamaCppServer{behavior: map[string]any{}}
	for k, v := range behavior {
		ls.behavior[k] = v
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		ls.mu.Lock()
		b := map[string]any{}
		for k, v := range ls.behavior {
			b[k] = v
		}
		ls.mu.Unlock()
		path := strings.TrimRight(r.URL.Path, "/")
		switch {
		case strings.HasSuffix(path, "/props"):
			if truthyAny(b["hide_props"]) {
				writeJSONReply(w, 404, map[string]any{"error": "not found"})
				return
			}
			props := defaultLlamaProps
			if p, ok := b["props"].(map[string]any); ok {
				props = p
			}
			if truthyAny(b["not_llamacpp"]) {
				props = map[string]any{
					"data":   []any{map[string]any{"id": "some-model"}},
					"object": "list"}
			}
			writeJSONReply(w, 200, props)
		case strings.HasSuffix(path, "/health"):
			status := "ok"
			if s, ok := b["health_status"].(string); ok {
				status = s
			}
			code := 200
			if c, ok := b["health_code"].(int); ok {
				code = c
			} else if status == "loading" {
				code = 503
			}
			writeJSONReply(w, code, map[string]any{
				"status": status, "slots": []any{}})
		case strings.HasSuffix(path, "/models"):
			if truthyAny(b["hide_models"]) {
				writeJSONReply(w, 404, map[string]any{"error": "not found"})
				return
			}
			names := []any{"qwen2.5-3b"}
			if m, ok := b["models"].([]any); ok {
				names = m
			}
			owned, hasOwned := b["models_owned_by"]
			data := []any{}
			for _, n := range names {
				entry := map[string]any{"id": n, "object": "model"}
				if hasOwned {
					entry["owned_by"] = owned
				}
				data = append(data, entry)
			}
			if sh, ok := b["server_header"].(string); ok {
				w.Header().Set("Server", sh)
			}
			writeJSONReply(w, 200, map[string]any{
				"object": "list", "data": data})
		default:
			writeJSONReply(w, 404, map[string]any{"error": "no such route"})
		}
	})
	ls.srv = &http.Server{Handler: mux}
	serveLocal(t, ls.srv, &ls.listener)
	t.Cleanup(func() { _ = ls.srv.Close() })
	return ls
}

func (ls *llamaCppServer) root() string {
	return "http://127.0.0.1:" + strconv.Itoa(ls.port())
}

func (ls *llamaCppServer) port() int {
	return ls.listener.Addr().(*net.TCPAddr).Port
}

func (ls *llamaCppServer) setBehavior(key string, value any) {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	ls.behavior[key] = value
}

func truthyAny(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	}
	return v != nil
}

// deadPortURL returns a loopback URL that is guaranteed to refuse
// connections (the suites' _dead_url/_dead_port helpers).
func deadPortURL(t *testing.T) string {
	t.Helper()
	return "http://127.0.0.1:" + strconv.Itoa(deadPort(t))
}

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

// normString applies the golden normalization: <API> before <ROOT>
// (the API contains the ROOT), <DEADURL> for dead addresses.
func normString(s, root, dead string) string {
	if root != "" {
		s = strings.ReplaceAll(s, root+"/v1", "<API>")
		s = strings.ReplaceAll(s, root, "<ROOT>")
	}
	if dead != "" {
		s = strings.ReplaceAll(s, dead, "<DEADURL>")
	}
	return s
}

// normalizeValue walks a decoded-JSON value replacing server URLs with
// the golden placeholders (the generator's norm()).
func normalizeValue(v any, root, dead string) any {
	switch t := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, vv := range t {
			out[k] = normalizeValue(vv, root, dead)
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, vv := range t {
			out[i] = normalizeValue(vv, root, dead)
		}
		return out
	case string:
		return normString(t, root, dead)
	default:
		return v
	}
}
