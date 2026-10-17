// stubworker_test.go — THE LOOPBACK RIG (the port of the reference's
// _StubWorker + _StubWorker2): a real local HTTP server on an
// ephemeral 127.0.0.1 port speaking the Worker's directory protocol —
// the Bearer check, /api/directory/sync (publish), /api/commands
// (pull), /api/commands/<id>/ack — with an ORDERED event log (the
// golden-rule-5 order proof). Loopback only; no real network.
package directorysync

import (
	"io"
	"net"
	"net/http"
	"strconv"
	"sync"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

const stubToken = "stub-token-0123456789abcdef"

// stubEvent is one ordered wire event: ("pull"), ("sync", n), ("ack",
// id, status).
type stubEvent struct {
	kind    string
	syncN   int
	command string
	status  string
}

// stubWorker is the loopback Worker (the reference's _StubWorker2 —
// which subsumes _StubWorker's publish door).
type stubWorker struct {
	t        *testing.T
	token    string
	srv      *http.Server
	listener net.Listener

	mu             sync.Mutex
	received       []*storage.OrderedMap
	publishedIndex *storage.OrderedMap
	commands       []any
	acks           [][2]string // (command_id, json body)
	events         []stubEvent
}

func newStubWorker(t *testing.T, token string) *stubWorker {
	t.Helper()
	if token == "" {
		token = stubToken
	}
	sw := &stubWorker{t: t, token: token, commands: []any{}}
	mux := http.NewServeMux()
	mux.HandleFunc("/", sw.route)
	sw.srv = &http.Server{Handler: mux}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	sw.listener = listener
	go func() { _ = sw.srv.Serve(listener) }()
	t.Cleanup(func() { _ = sw.srv.Close() })
	return sw
}

func (sw *stubWorker) url() string {
	return "http://127.0.0.1:" + strconv.Itoa(sw.listener.Addr().(*net.TCPAddr).Port)
}

func (sw *stubWorker) setCommands(commands []any) {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	sw.commands = commands
}

func (sw *stubWorker) eventLog() []stubEvent {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	out := make([]stubEvent, len(sw.events))
	copy(out, sw.events)
	return out
}

func (sw *stubWorker) receivedIndexes() []*storage.OrderedMap {
	sw.mu.Lock()
	defer sw.mu.Unlock()
	out := make([]*storage.OrderedMap, len(sw.received))
	copy(out, sw.received)
	return out
}

func (sw *stubWorker) bearerOK(r *http.Request) bool {
	return r.Header.Get("Authorization") == "Bearer "+sw.token
}

func (sw *stubWorker) writeJSON(w http.ResponseWriter, code int, payload any) {
	data := []byte(storage.PyDumpsCompact(payload))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func (sw *stubWorker) route(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	// the read-back door first: GET /api/directory/index is the PUBLIC
	// site's index endpoint — the reference's fetch_published_index
	// wears only the UA + Accept headers, never the Bearer.
	if path == "/api/directory/index" && r.Method == http.MethodGet {
		sw.mu.Lock()
		published := sw.publishedIndex
		sw.mu.Unlock()
		if published == nil {
			// the Worker's own honest 404 — never-published is a STATE,
			// not an error
			sw.writeJSON(w, 404, map[string]any{
				"error": "index not published yet",
				"hint":  "the automation publishes it via POST /api/directory/sync"})
			return
		}
		sw.writeJSON(w, 200, published)
		return
	}
	if !sw.bearerOK(r) {
		sw.writeJSON(w, 401, map[string]any{"error": "unauthorized"})
		return
	}
	switch {
	case path == "/api/commands" && r.Method == http.MethodGet:
		sw.mu.Lock()
		sw.events = append(sw.events, stubEvent{kind: "pull"})
		commands := sw.commands
		sw.mu.Unlock()
		sw.writeJSON(w, 200, map[string]any{"commands": commands})
	case len(path) > len("/api/commands/") && r.URL.Path[len(path)-len("/ack"):] == "/ack" &&
		r.Method == http.MethodPost:
		commandID := path[len("/api/commands/") : len(path)-len("/ack")]
		body, _ := io.ReadAll(r.Body)
		payload, _ := storage.DecodeJSONValue(body)
		status := pyStrOrEmpty(dictDoorOf(payload).Get("status"))
		sw.mu.Lock()
		sw.acks = append(sw.acks, [2]string{commandID, string(body)})
		sw.events = append(sw.events, stubEvent{kind: "ack", command: commandID, status: status})
		sw.mu.Unlock()
		sw.writeJSON(w, 200, map[string]any{"success": true})
	case path == "/api/directory/sync" && r.Method == http.MethodPost:
		body, _ := io.ReadAll(r.Body)
		payload, err := storage.DecodeJSONValue(body)
		if err != nil || !dictDoorOf(payload).Has("schema") {
			sw.writeJSON(w, 400, map[string]any{"error": "invalid JSON"})
			return
		}
		sw.mu.Lock()
		if om, ok := payload.(*storage.OrderedMap); ok {
			sw.received = append(sw.received, om)
			sw.publishedIndex = om
		}
		n := 0
		if sites, ok := dictDoorOf(payload).Get("sites").([]any); ok {
			n = len(sites)
		}
		sw.events = append(sw.events, stubEvent{kind: "sync", syncN: n})
		sw.mu.Unlock()
		sw.writeJSON(w, 200, map[string]any{
			"success": true,
			"stats":   map[string]any{"received": n, "stored": n},
		})
	default:
		sw.writeJSON(w, 404, map[string]any{"error": "not found"})
	}
}
