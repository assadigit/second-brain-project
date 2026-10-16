// detectset_test.go — T-30 characterization (Phase 3): the Detect & Set
// fast lane's llm_client half.
//
// Ported/landed suites:
//   - test_detectset.py: TestDetectOllama (found + models by name key,
//     the model-key fallback + dedup keeping order, up-but-empty, dead
//     port never raises, non-dict JSON is not found, the missing /api/
//     tags route is not found, URL normalization (no scheme / trailing
//     slash), hostile inputs never raise, the default-URL law).
package llmclient

import (
	"strings"
	"testing"
)

func TestDetectOllamaFoundWithModelsNameKey(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, map[string]any{
			"models": []any{
				map[string]any{"name": "llama3:latest"},
				map[string]any{"name": "qwen2.5:7b"}}})})
	r := DetectOllama(rs.root(), 4.0)
	if !r.Found {
		t.Fatalf("found = false (detail %q)", r.Detail)
	}
	if r.BaseURL != rs.root() {
		t.Errorf("base_url = %q", r.BaseURL)
	}
	if len(r.Models) != 2 || r.Models[0] != "llama3:latest" ||
		r.Models[1] != "qwen2.5:7b" {
		t.Errorf("models = %v", r.Models)
	}
	if !strings.Contains(r.Detail, "2 model(s)") {
		t.Errorf("detail = %q", r.Detail)
	}
}

func TestDetectOllamaModelKeyFallbackAndDedupKeepsOrder(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, map[string]any{
			"models": []any{
				map[string]any{"model": "gemma2:2b"},
				map[string]any{"name": "gemma2:2b"}, // duplicate, other key
				map[string]any{"name": ""},
				map[string]any{"model": nil},
				map[string]any{"name": "mistral:latest"}}})})
	r := DetectOllama(rs.root(), 4.0)
	if !r.Found {
		t.Fatal("found = false")
	}
	if len(r.Models) != 2 || r.Models[0] != "gemma2:2b" ||
		r.Models[1] != "mistral:latest" {
		t.Errorf("models = %v", r.Models)
	}
}

func TestDetectOllamaUpButNoModels(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, map[string]any{
			"models": []any{}})})
	r := DetectOllama(rs.root(), 4.0)
	if !r.Found {
		t.Fatal("found = false")
	}
	if len(r.Models) != 0 {
		t.Errorf("models = %v", r.Models)
	}
	if !strings.Contains(r.Detail, "NO models") {
		t.Errorf("detail = %q", r.Detail)
	}
}

func TestDetectOllamaDeadPortNeverRaises(t *testing.T) {
	r := DetectOllama(deadPortURL(t), 4.0)
	if r.Found {
		t.Fatal("dead port found")
	}
	if len(r.Models) != 0 {
		t.Errorf("models = %v", r.Models)
	}
	if r.Detail == "" {
		t.Error("empty detail")
	}
}

func TestDetectOllamaNonDictJSONIsNotFound(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, []any{"not", "a", "dict"})})
	r := DetectOllama(rs.root(), 4.0)
	if r.Found {
		t.Fatal("non-dict JSON found")
	}
	if r.Detail != "ValueError: response is not a JSON object" {
		t.Errorf("detail = %q", r.Detail)
	}
}

func TestDetectOllamaMissingTagsRouteIsNotFound(t *testing.T) {
	// An HTTP server that 404s /api/tags (i.e. NOT an Ollama server).
	rs := newRecordingServer(t, map[string]any{
		"/": route(200, map[string]any{})})
	r := DetectOllama(rs.root(), 4.0)
	if r.Found {
		t.Fatal("missing route found")
	}
	if !strings.HasPrefix(r.Detail, "HTTPError") {
		t.Errorf("detail = %q", r.Detail)
	}
}

func TestDetectOllamaURLNormalizationSchemeAndSlash(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, map[string]any{
			"models": []any{map[string]any{"name": "m1"}}})})
	root := rs.root()
	for _, spelling := range []string{root, root + "/",
		strings.TrimPrefix(root, "http://")} {
		r := DetectOllama(spelling, 4.0)
		if !r.Found {
			t.Errorf("spelling %q not found (detail %q)", spelling, r.Detail)
			continue
		}
		if r.BaseURL != root {
			t.Errorf("spelling %q → base_url %q", spelling, r.BaseURL)
		}
	}
}

func TestDetectOllamaHostileInputsNeverRaise(t *testing.T) {
	for _, bad := range []string{"", "   ", "::::", "http://", "123"} {
		r := DetectOllama(bad, 0.5)
		if r.Found {
			t.Errorf("hostile input %q found something", bad)
		}
	}
}

func TestDetectOllamaDefaultURLIsTheOllamaPort(t *testing.T) {
	// The default must stay the Ollama default (11434), never invented.
	if DefaultOllamaBase != "http://127.0.0.1:11434" {
		t.Errorf("default base = %q", DefaultOllamaBase)
	}
	// An empty input probes the default (nothing listens there in the
	// sandbox → not found, never a crash).
	r := DetectOllama("", 0.5)
	if r.Found || r.BaseURL != DefaultOllamaBase {
		t.Errorf("default probe = %+v", r)
	}
}
