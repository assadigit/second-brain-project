package workerclient

// fakes_test.go — THE LOOPBACK RIG (ported from the reference's
// tests/test_synclive.py _FakeWorker): a real local HTTP server on an
// ephemeral 127.0.0.1 port speaking the Cloudflare Worker's own
// protocol byte-for-byte — the same GET /api/pending HMAC verification
// cloudflare-bot's verifyHmac performs (headers present, the 5-minute
// timestamp window, the known install, the signature recomputed over
// METHOD\nPATH\nTIMESTAMP\nsha256hex(body) keyed by
// sha256hex(shared_secret), constant-time), the 1010 law enforced
// SERVER-SIDE on every leg (a banned UA is rejected, so any green
// client test proves the desktop wore its own name on the wire), the
// pair door, the mirror-push idempotency (sync_processed:<sync_id>),
// the confirmation-gate trios, and the directory/automation doors (the
// honest 404, the Bearer publish, /api/commands) T-49 will reuse.

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
)

const (
	rigInstall = "11111111-2222-3333-4444-555555555555"
	rigSecret  = "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"
	rigToken   = "the-directory-sync-token"
)

// rigReq is one recorded request (the wire evidence).
type rigReq struct {
	Method string
	Path   string
	UA     string
	Body   string
	Authed bool
}

// rigRejection is one refused leg (the 401/403 story).
type rigRejection struct {
	Path string
	Why  string
}

// fakeWorker is the loopback Worker.
type fakeWorker struct {
	t        *testing.T
	install  string
	secret   string
	token    string
	srv      *http.Server
	listener net.Listener

	mu          sync.Mutex
	queue       []any
	published   *storage.OrderedMap
	requests    []rigReq
	rejections  []rigRejection
	syncSeen    map[string]int // sync_id -> pushes
	pairCode    string
	banishAsks  map[string]string
	scanAsks    map[string]string
	decommSeen  []any
	errorsSeen  []any
	backupSeen  []*storage.OrderedMap
	backfillN   int
	askCounter  int
	commandsOut []any
}

func newFakeWorker(t *testing.T) *fakeWorker {
	t.Helper()
	fw := &fakeWorker{
		t:          t,
		install:    rigInstall,
		secret:     rigSecret,
		token:      rigToken,
		syncSeen:   map[string]int{},
		banishAsks: map[string]string{},
		scanAsks:   map[string]string{},
		pairCode:   "ABC12345",
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/", fw.route)
	fw.srv = &http.Server{Handler: mux}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	fw.listener = listener
	go func() { _ = fw.srv.Serve(listener) }()
	t.Cleanup(func() { _ = fw.srv.Close() })
	return fw
}

func (fw *fakeWorker) url() string {
	return "http://127.0.0.1:" + strconv.Itoa(fw.port())
}

func (fw *fakeWorker) port() int {
	return fw.listener.Addr().(*net.TCPAddr).Port
}

func (fw *fakeWorker) addPending(rows ...any) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.queue = append(fw.queue, rows...)
}

func (fw *fakeWorker) clearPending() {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	fw.queue = nil
}

func (fw *fakeWorker) captured() []rigReq {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	out := make([]rigReq, len(fw.requests))
	copy(out, fw.requests)
	return out
}

func (fw *fakeWorker) rejected() []rigRejection {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	out := make([]rigRejection, len(fw.rejections))
	copy(out, fw.rejections)
	return out
}

func (fw *fakeWorker) writeJSON(w http.ResponseWriter, code int, payload any) {
	data := []byte(storage.PyDumpsCompact(payload))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

// verifyHmac is the Worker's own verifyHmac (api.js), byte-for-byte —
// the reference rig's _verify_hmac: ” when the leg is authentic, the
// reason otherwise.
func (fw *fakeWorker) verifyHmac(r *http.Request, method, path string,
	body []byte) string {
	install := r.Header.Get("X-Auth-Install")
	ts := r.Header.Get("X-Auth-Timestamp")
	sig := r.Header.Get("X-Auth-Signature")
	if install == "" || ts == "" || sig == "" {
		return "Missing auth headers"
	}
	tsMS, err := strconv.ParseInt(ts, 10, 64)
	if err != nil {
		return "Timestamp out of range"
	}
	now := time.Now().UnixMilli()
	diff := now - tsMS
	if diff < 0 {
		diff = -diff
	}
	if diff > 5*60*1000 {
		return "Timestamp out of range"
	}
	if install != fw.install {
		return "Unknown install"
	}
	// the 1010 law — the edge's own grammar: the app's name, never a
	// banned signature.
	ua := r.Header.Get("User-Agent")
	if !strings.HasPrefix(ua, "GitCurator/") || strings.Contains(ua, "Python-urllib") {
		return "banned user agent (the 1010 law)"
	}
	// the signature the desktop computes and the Worker recomputes:
	// key = sha256hex(secret), message = METHOD\nPATH\nTS\nsha256hex(body)
	keyHex := sha256Hex([]byte(fw.secret))
	message := method + "\n" + path + "\n" + ts + "\n" + sha256Hex(body)
	mac := hmac.New(sha256.New, []byte(keyHex))
	mac.Write([]byte(message))
	expected := hex.EncodeToString(mac.Sum(nil))
	if subtle.ConstantTimeCompare([]byte(sig), []byte(expected)) != 1 {
		return "Invalid signature"
	}
	return ""
}

// bearerToken extracts + constant-time-compares the Bearer token (the
// publish door's law).
func (fw *fakeWorker) bearerOK(r *http.Request) bool {
	token := strings.TrimSpace(r.Header.Get("Authorization"))
	if !strings.HasPrefix(strings.ToLower(token), "bearer ") {
		return false
	}
	token = strings.TrimSpace(token[len("Bearer "):])
	return subtle.ConstantTimeCompare([]byte(token), []byte(fw.token)) == 1
}

// ua1010OK: the public doors' UA law (GitCurator/ prefix).
func ua1010OK(r *http.Request) bool {
	return strings.HasPrefix(r.Header.Get("User-Agent"), "GitCurator/")
}

func (fw *fakeWorker) route(w http.ResponseWriter, r *http.Request) {
	method := r.Method
	path := strings.Split(r.URL.RequestURI(), "?")[0]
	// the full RequestURI (path + query) — what the HMAC signs
	pathWithQuery := r.URL.RequestURI()
	body, _ := io.ReadAll(r.Body)
	rec := rigReq{Method: method, Path: pathWithQuery,
		UA: r.Header.Get("User-Agent"), Body: string(body)}
	fw.mu.Lock()
	fw.requests = append(fw.requests, rec)
	fw.mu.Unlock()

	reject := func(why string) {
		fw.mu.Lock()
		fw.rejections = append(fw.rejections, rigRejection{pathWithQuery, why})
		fw.mu.Unlock()
		fw.writeJSON(w, 401, map[string]any{"error": why})
	}

	switch {
	case path == "/api/pair" && method == "POST":
		if !ua1010OK(r) {
			fw.mu.Lock()
			fw.rejections = append(fw.rejections,
				rigRejection{path, "banned user agent"})
			fw.mu.Unlock()
			fw.writeJSON(w, 403, map[string]any{"error": "banned signature"})
			return
		}
		data, err := storage.DecodeJSONValue(body)
		code, _ := "", ""
		if err != nil || !isDictLike(data) {
			fw.writeJSON(w, 400, map[string]any{"error": "missing code"})
			return
		}
		code = pyStr(dictValue(data, "code"))
		if code == "" {
			fw.writeJSON(w, 400, map[string]any{"error": "missing code"})
			return
		}
		if code != fw.pairCode {
			fw.writeJSON(w, 401, map[string]any{"error": "invalid or expired code"})
			return
		}
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "install_id": fw.install,
			"shared_secret": fw.secret})

	case path == "/api/pending" && method == "GET":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		queue := make([]any, len(fw.queue))
		copy(queue, fw.queue)
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "pending": queue, "count": len(queue)})

	case path == "/api/vault_mirror/push" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		data, err := storage.DecodeJSONValue(body)
		if err != nil || !isDictLike(data) {
			fw.writeJSON(w, 400, map[string]any{"error": "bad body"})
			return
		}
		syncID := pyStr(dictValue(data, "sync_id"))
		fw.mu.Lock()
		fw.syncSeen[syncID]++
		nth := fw.syncSeen[syncID]
		total := len(fw.syncSeen)
		fw.mu.Unlock()
		entries, _ := dictValue(data, "entries").([]any)
		decomm, _ := dictValue(data, "decommission_events").([]any)
		fw.writeJSON(w, 200, map[string]any{
			"success":        true,
			"synced_entries": len(entries), "synced_decomm": len(decomm),
			"conflicts": []any{}, "vault_mirror_total": total,
			"idempotent": nth > 1})

	case path == "/api/decommission" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		fw.decommSeen = append(fw.decommSeen, mustDecode(body))
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "decommissioned": true})

	case path == "/api/errors" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		fw.errorsSeen = append(fw.errorsSeen, mustDecode(body))
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "received": true, "dm_sent": false})

	case path == "/api/gdrive/backup_status" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		fw.backupSeen = append(fw.backupSeen, mustOM(body))
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{"success": true})

	case path == "/api/banish/propose" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		fw.askCounter++
		id := fmt.Sprintf("ban-%d", fw.askCounter)
		fw.banishAsks[id] = "pending"
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "id": id, "message_id": 1000 + fw.askCounter})

	case path == "/api/banish/status" && method == "GET":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		id := r.URL.Query().Get("id")
		fw.mu.Lock()
		status, ok := fw.banishAsks[id]
		fw.mu.Unlock()
		if !ok {
			fw.writeJSON(w, 404, map[string]any{"error": "unknown id"})
			return
		}
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "id": id, "status": status, "count": 1})

	case path == "/api/banish/result" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.writeJSON(w, 200, map[string]any{"success": true, "status": "done"})

	case path == "/api/scan/propose" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		fw.askCounter++
		id := fmt.Sprintf("scan-%d", fw.askCounter)
		fw.scanAsks[id] = "pending"
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{"success": true, "id": id})

	case path == "/api/scan/status" && method == "GET":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		id := r.URL.Query().Get("id")
		fw.mu.Lock()
		status, ok := fw.scanAsks[id]
		fw.mu.Unlock()
		if !ok {
			fw.writeJSON(w, 404, map[string]any{"error": "unknown id"})
			return
		}
		fw.writeJSON(w, 200, map[string]any{"success": true, "id": id, "status": status})

	case path == "/api/scan/result" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.writeJSON(w, 200, map[string]any{"success": true, "status": "done"})

	case path == "/api/verify" && method == "GET":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.writeJSON(w, 200, map[string]any{
			"success": true,
			"stats": map[string]any{
				"total_ledger": 3, "in_vault": 2, "pending": 1,
				"decommissioned": 0, "dead_letters": 0, "non_github": 1},
			"ledger_hash": "abc", "vault_mirror_hash": "def"})

	case path == "/api/decommissions" && method == "GET":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		fw.mu.Lock()
		decomms := fw.decommSeen
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "decommissions": decomms, "count": len(decomms)})

	case path == "/api/backfill" && method == "POST":
		if why := fw.verifyHmac(r, method, pathWithQuery, body); why != "" {
			reject(why)
			return
		}
		fw.markAuthed(pathWithQuery)
		data := mustDecode(body)
		entries, _ := dictValue(data, "entries").([]any)
		fw.mu.Lock()
		fw.backfillN += len(entries)
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "inserted": len(entries), "skipped": 0,
			"progress": "1/1"})

	case path == "/health" && method == "GET":
		if !ua1010OK(r) {
			fw.mu.Lock()
			fw.rejections = append(fw.rejections,
				rigRejection{path, "banned user agent"})
			fw.mu.Unlock()
			fw.writeJSON(w, 403, map[string]any{"error": "banned signature"})
			return
		}
		fw.writeJSON(w, 200, map[string]any{
			"status": "ok", "service": "github-curator-bot",
			"version": "0.40.0", "timestamp": nowISO()})

	// ---- the directory/automation doors (Bearer; T-49 reuses) ----

	case path == "/api/directory/index" && method == "GET":
		if !ua1010OK(r) {
			fw.mu.Lock()
			fw.rejections = append(fw.rejections,
				rigRejection{path, "banned user agent"})
			fw.mu.Unlock()
			fw.writeJSON(w, 403, map[string]any{"error": "banned signature"})
			return
		}
		fw.mu.Lock()
		published := fw.published
		fw.mu.Unlock()
		if published == nil {
			// the Worker's own honest 404 — never-published is a
			// STATE, not an error
			fw.writeJSON(w, 404, map[string]any{
				"error": "index not published yet",
				"hint": "the automation publishes it via POST " +
					"/api/directory/sync"})
			return
		}
		fw.writeJSON(w, 200, published)

	case path == "/api/directory/sync" && method == "POST":
		if !ua1010OK(r) || !fw.bearerOK(r) {
			fw.mu.Lock()
			fw.rejections = append(fw.rejections,
				rigRejection{path, "unauthorized publish"})
			fw.mu.Unlock()
			fw.writeJSON(w, 401, map[string]any{"error": "unauthorized"})
			return
		}
		data, err := storage.DecodeJSONValue(body)
		if err != nil || !isDictLike(data) {
			fw.writeJSON(w, 400, map[string]any{"error": "invalid JSON"})
			return
		}
		sites := 0
		if counts, ok := dictValue(data, "counts").(*storage.OrderedMap); ok {
			v, _ := counts.Get("sites")
			sites = pyIntDefault(v, 0)
		}
		if sites <= 0 {
			fw.writeJSON(w, 400, map[string]any{"error": "no site count"})
			return
		}
		fw.mu.Lock()
		fw.published = mustOM(body)
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"success": true, "published_at": nowISO(),
			"stats": map[string]any{"sites": sites}})

	case path == "/api/commands" && method == "GET":
		if !fw.bearerOK(r) {
			fw.mu.Lock()
			fw.rejections = append(fw.rejections,
				rigRejection{path, "unauthorized commands"})
			fw.mu.Unlock()
			fw.writeJSON(w, 401, map[string]any{"error": "unauthorized"})
			return
		}
		fw.mu.Lock()
		commands := fw.commandsOut
		fw.mu.Unlock()
		fw.writeJSON(w, 200, map[string]any{
			"schema": "gitcurator-commands/1", "commands": commands})

	default:
		fw.writeJSON(w, 404, map[string]any{"error": "not found"})
	}
}

func (fw *fakeWorker) markAuthed(pathWithQuery string) {
	fw.mu.Lock()
	defer fw.mu.Unlock()
	if n := len(fw.requests); n > 0 {
		fw.requests[n-1].Authed = true
	}
}

func mustDecode(body []byte) any {
	v, err := storage.DecodeJSONValue(body)
	if err != nil {
		return storage.NewOrderedMap()
	}
	return v
}

func mustOM(body []byte) *storage.OrderedMap {
	if om, ok := mustDecode(body).(*storage.OrderedMap); ok {
		return om
	}
	return storage.NewOrderedMap()
}

// rigConfig builds the enabled client config pointed at the rig.
func rigConfig(fw *fakeWorker) map[string]any {
	return map[string]any{
		"cloudflare_worker_url":    fw.url(),
		"cloudflare_install_id":    rigInstall,
		"cloudflare_shared_secret": rigSecret,
		"cloudflare_enabled":       true,
	}
}

// withSeam swaps a package seam for one test and restores it.
func withSeam[T any](t *testing.T, slot *T, fake T) {
	t.Helper()
	old := *slot
	*slot = fake
	t.Cleanup(func() { *slot = old })
}
