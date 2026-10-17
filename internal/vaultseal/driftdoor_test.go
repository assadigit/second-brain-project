// driftdoor_test.go — T-49's wiring proof: the drift seal's publish
// door (DriftSealDeps.PublishAfterBatch) is wired to the REAL
// directorysync.PublishAfterBatch in DefaultDriftSealDeps — the
// honest not-wired warning retires. The manual fast-sync path runs
// the full golden-rule-5 cycle against the loopback stub (loopback
// only, no real network).
package vaultseal

import (
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// stubMux is the directory-sync loopback stub for the drift door's
// proof: the Bearer-checked publish + the commands pull/ack doors
// (loopback only; no real network).
type stubMux struct {
	url       string
	published *[]*storage.OrderedMap
}

func newStubMux(t *testing.T, token string, published *[]*storage.OrderedMap) *stubMux {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/api/commands", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			writeStubJSON(w, 401, map[string]any{"error": "unauthorized"})
			return
		}
		writeStubJSON(w, 200, map[string]any{"commands": []any{}})
	})
	mux.HandleFunc("/api/directory/sync", func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer "+token {
			writeStubJSON(w, 401, map[string]any{"error": "unauthorized"})
			return
		}
		body, _ := io.ReadAll(r.Body)
		if om, err := storage.DecodeJSON(body); err == nil {
			*published = append(*published, om)
		}
		writeStubJSON(w, 200, map[string]any{"success": true})
	})
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	srv := &http.Server{Handler: mux}
	go func() { _ = srv.Serve(listener) }()
	t.Cleanup(func() { _ = srv.Close() })
	return &stubMux{url: "http://127.0.0.1:" + strconv.Itoa(listener.Addr().(*net.TCPAddr).Port)}
}

func writeStubJSON(w http.ResponseWriter, code int, payload any) {
	data := []byte(storage.PyDumpsCompact(payload))
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	w.WriteHeader(code)
	_, _ = w.Write(data)
}

func TestDriftSealDoor_TheDefaultWiringBindsTheRealPublish(t *testing.T) {
	deps := DefaultDriftSealDeps()
	if deps.PublishAfterBatch == nil {
		t.Fatal("the default publish door is not wired (the T-49 door)")
	}
}

func TestDriftSealDoor_TheManualPathRunsTheDirectoryCycle(t *testing.T) {
	// manual=true + a websites-enabled config: the fast-sync door asks
	// for the site's leg — the real directorysync publish rides the
	// stub Worker (loopback), the full command cycle runs, and the
	// verdict speaks the published line.
	vault := t.TempDir()
	if err := os.MkdirAll(filepath.Join(vault, "Design", "Assets & Resources"), 0o755); err != nil {
		t.Fatal(err)
	}
	note := "---\nsource: https://coolors.co/\naliases: []\ntags: [color-tool]\n" +
		"category: \"Design\"\nsubcategory: \"Assets & Resources\"\npricing: \"freemium\"\n" +
		"login_required: \"no\"\n---\n\n# Coolors\n\n> **TL;DR:** t\n"
	if err := os.WriteFile(filepath.Join(vault, "Design", "Assets & Resources", "Coolors.md"),
		[]byte(note), 0o644); err != nil {
		t.Fatal(err)
	}

	var published []*storage.OrderedMap
	token := "stub-token-0123456789abcdef"
	mux := newStubMux(t, token, &published)

	config := map[string]any{
		"pipelines":            map[string]any{"websites": true},
		"website_vault_path":   vault,
		"directory_worker_url": mux.url,
		"directory_sync_token": token,
	}
	var logs []string
	line, level := RunDriftSeal(config, nil, func(msg, lvl string) { logs = append(logs, msg) }, true)
	if level != "success" && level != "warning" {
		t.Fatalf("level: %q (%s)", level, line)
	}
	if len(published) != 1 {
		t.Fatalf("the directory cycle did not publish: %d (line: %s)", len(published), line)
	}
	if !containsSub(line, "directory published") {
		t.Fatalf("the verdict does not speak the publish: %s", line)
	}
}

func containsSub(haystack, needle string) bool {
	for i := 0; i+len(needle) <= len(haystack); i++ {
		if haystack[i:i+len(needle)] == needle {
			return true
		}
	}
	return false
}
