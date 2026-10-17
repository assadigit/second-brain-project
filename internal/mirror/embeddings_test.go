// embeddings_test.go — the ported test_phase6 embedding selections:
// the cosine math, the embed TEXT composition, the Ollama /
// OpenAI-compatible endpoints against a loopback stub (the reference's
// _FakeEmbedHandler, ported), the provider routing, the SQLite store,
// and the stale-only refresh.
package mirror

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"
)

// fakeEmbedBehavior is the loopback stub's configuration (the
// reference's server.behavior dict).
type fakeEmbedBehavior struct {
	mode     string // 'openai' | 'ollama' | 'ollama-legacy'
	dim      int    // vector dimension (default 4)
	badCount bool   // return one vector short (error path)
}

// startFakeEmbed mirrors _start_embed: one loopback server, configured
// through the behavior struct. NO real network — 127.0.0.1 only.
func startFakeEmbed(t *testing.T, behavior fakeEmbedBehavior) *httptest.Server {
	t.Helper()
	if behavior.dim == 0 {
		behavior.dim = 4
	}
	if behavior.mode == "" {
		behavior.mode = "openai"
	}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Input  []string `json:"input"`
			Prompt string   `json:"prompt"`
			Model  string   `json:"model"`
		}
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		texts := body.Input
		if body.Prompt != "" {
			texts = []string{body.Prompt}
		}
		if behavior.badCount && len(texts) > 1 {
			texts = texts[:len(texts)-1] // one vector short
		}
		writeJSON := func(payload any) {
			w.Header().Set("Content-Type", "application/json")
			_ = json.NewEncoder(w).Encode(payload)
		}
		switch behavior.mode {
		case "openai":
			// entries are REVERSED on purpose, marked with index — the
			// client must honor the index, not the list order
			var data []map[string]any
			for i := range texts {
				vec := make([]float64, behavior.dim)
				for j := range vec {
					vec[j] = float64(i+1) + 0.1*float64(j)
				}
				data = append(data, map[string]any{
					"object":    "embedding",
					"index":     len(texts) - 1 - i,
					"embedding": vec,
				})
			}
			writeJSON(map[string]any{"object": "list", "data": data})
		case "ollama":
			var vectors [][]float64
			for i := range texts {
				vec := make([]float64, behavior.dim)
				for j := range vec {
					vec[j] = float64(i+1) + 0.1*float64(j)
				}
				vectors = append(vectors, vec)
			}
			writeJSON(map[string]any{"model": body.Model, "embeddings": vectors})
		default:
			// legacy /api/embeddings — one prompt per call
			i := 0
			if body.Prompt == "second" {
				i = 1
			}
			vec := make([]float64, behavior.dim)
			for j := range vec {
				vec[j] = float64(i+1) + 0.1*float64(j)
			}
			writeJSON(map[string]any{"embedding": vec})
		}
	}))
	t.Cleanup(server.Close)
	return server
}

// startLegacyFallback mirrors the ollama-fallback variant: 404 on
// /api/embed (old server) → the client falls back to per-text
// /api/embeddings.
func startLegacyFallback(t *testing.T) *httptest.Server {
	t.Helper()
	var calls int32
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(strings.TrimRight(r.URL.Path, "/"), "/api/embed") {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(404)
			_, _ = w.Write([]byte(`{"error": "not found"}`))
			return
		}
		atomic.AddInt32(&calls, 1)
		var body struct {
			Prompt string `json:"prompt"`
		}
		data, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(data, &body)
		i := 0
		if body.Prompt == "second" {
			i = 1
		}
		vec := make([]float64, 4)
		for j := range vec {
			vec[j] = float64(i+1) + 0.1*float64(j)
		}
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{"embedding": vec})
	}))
	t.Cleanup(server.Close)
	return server
}

// -- TestEmbedMath ------------------------------------------------------------

func TestEmbedMathCosine(t *testing.T) {
	approx := func(a, b float64) bool {
		d := a - b
		return d < 1e-9 && d > -1e-9
	}
	if !approx(Cosine([]float64{1, 0}, []float64{1, 0}), 1.0) {
		t.Fatal("cos(a,a) != 1")
	}
	if !approx(Cosine([]float64{1, 0}, []float64{0, 1}), 0.0) {
		t.Fatal("orthogonal != 0")
	}
	if !approx(Cosine([]float64{1, 0}, []float64{-1, 0}), -1.0) {
		t.Fatal("opposite != -1")
	}
	if Cosine(nil, []float64{1}) != 0.0 {
		t.Fatal("empty != 0")
	}
	if Cosine([]float64{1}, []float64{1, 2}) != 0.0 {
		t.Fatal("mismatched lengths != 0")
	}
	if Cosine([]float64{0, 0}, []float64{0, 0}) != 0.0 {
		t.Fatal("zero vectors != 0")
	}
}

func TestEmbedMathEmbedTextComposition(t *testing.T) {
	fields := RecallFields{
		Name:    "Palette Gen",
		OneLine: "Generates color palettes.",
		Recall:  "Use when you need to build a scheme.",
		Tags:    "design, colors",
	}
	text := EmbedTextFor(fields)
	for _, needle := range []string{
		"Palette Gen", "Generates color palettes.",
		"Use when you need to build a scheme.", "tags: design, colors",
	} {
		if !strings.Contains(text, needle) {
			t.Fatalf("needle %q missing from %q", needle, text)
		}
	}
	// NEVER full text: the whole embedding input stays tiny
	if runeLen(text) > 1200 {
		t.Fatalf("embed text too long: %d runes", runeLen(text))
	}
	if EmbedTextFor(RecallFields{}) != "" {
		t.Fatal("empty fields embed to non-empty text")
	}
}

func TestEmbedMathHashChangesWithText(t *testing.T) {
	if HashEmbedText("a") == HashEmbedText("b") {
		t.Fatal("hashes collide")
	}
	if HashEmbedText("a") != HashEmbedText("a") {
		t.Fatal("hash not stable")
	}
}

// -- TestEmbedEndpoints ---------------------------------------------------------

func TestEmbedEndpointsOpenAIEmbedHonorsIndex(t *testing.T) {
	server := startFakeEmbed(t, fakeEmbedBehavior{mode: "openai"})
	vectors, err := OpenAIEmbed(server.URL, "m", []string{"first", "second"}, "", 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 2 {
		t.Fatalf("vectors = %d", len(vectors))
	}
	// the fake REVERSES entries and marks them with index — the client
	// must restore INPUT order: 'first' is index 0, whose entry carries
	// the [2.0…] vector
	want0 := []float64{2.0, 2.1, 2.2, 2.3}
	want1 := []float64{1.0, 1.1, 1.2, 1.3}
	for i := range want0 {
		if vectors[0][i] != want0[i] {
			t.Fatalf("vectors[0] = %v, want %v", vectors[0], want0)
		}
		if vectors[1][i] != want1[i] {
			t.Fatalf("vectors[1] = %v, want %v", vectors[1], want1)
		}
	}
}

func TestEmbedEndpointsOpenAIEmbedRejectsWrongCount(t *testing.T) {
	server := startFakeEmbed(t, fakeEmbedBehavior{mode: "openai", badCount: true})
	_, err := OpenAIEmbed(server.URL, "m", []string{"first", "second"}, "", 30)
	var embErr *EmbeddingError
	if err == nil || !errors.As(err, &embErr) {
		t.Fatalf("want EmbeddingError, got %v", err)
	}
}

func TestEmbedEndpointsOllamaEmbedBatched(t *testing.T) {
	server := startFakeEmbed(t, fakeEmbedBehavior{mode: "ollama"})
	vectors, err := OllamaEmbed(server.URL, "nomic-embed-text",
		[]string{"first", "second"}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 2 {
		t.Fatalf("vectors = %d", len(vectors))
	}
	if vectors[1][0] != 2.0 {
		t.Fatalf("vectors[1][0] = %v", vectors[1][0])
	}
}

func TestEmbedEndpointsOllamaLegacyFallback(t *testing.T) {
	// /api/embed 404s (old server) → per-text /api/embeddings
	server := startLegacyFallback(t)
	vectors, err := OllamaEmbed(server.URL, "nomic-embed-text",
		[]string{"first", "second"}, 30)
	if err != nil {
		t.Fatal(err)
	}
	if len(vectors) != 2 {
		t.Fatalf("vectors = %d", len(vectors))
	}
	if vectors[0][0] != 1.0 {
		t.Fatalf("vectors[0][0] = %v", vectors[0][0])
	}
	if vectors[1][0] != 2.0 {
		t.Fatalf("vectors[1][0] = %v", vectors[1][0])
	}
}

func TestEmbedEndpointsBuildEmbeddingCallRoutes(t *testing.T) {
	frozenClock(t, frozenTime())
	// ollama default (nested config) — the chat model as fallback
	embed, label, err := BuildEmbeddingCall(map[string]any{
		"llm_provider": "ollama",
		"ollama": map[string]any{
			"base_url": "http://127.0.0.1:11434",
			"model":    "llama3",
		},
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if embed == nil {
		t.Fatal("nil embed fn")
	}
	if !strings.Contains(label, "Ollama") || !strings.Contains(label, "llama3") {
		t.Fatalf("label = %q", label)
	}
	// embedding_model wins over the chat model
	_, label, err = BuildEmbeddingCall(map[string]any{
		"llm_provider":    "ollama",
		"embedding_model": "nomic-embed",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(label, "nomic-embed") {
		t.Fatalf("label = %q", label)
	}
	// llama.cpp normalization
	_, label, err = BuildEmbeddingCall(map[string]any{
		"llm_provider":     "llamacpp",
		"llamacpp_api_url": "127.0.0.1:8080",
		"llamacpp_model":   "qwen2.5-3b",
	}, "")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(label, "llama.cpp") {
		t.Fatalf("label = %q", label)
	}
	// cloud without a model is the honest error
	_, _, err = BuildEmbeddingCall(map[string]any{
		"llm_provider":  "cloud",
		"cloud_api_url": "https://x/v1",
		"cloud_api_key": "",
		"cloud_model":   "",
	}, "")
	var embErr *EmbeddingError
	if err == nil || !errors.As(err, &embErr) {
		t.Fatalf("want EmbeddingError, got %v", err)
	}
}

// -- TestEmbeddingStore ---------------------------------------------------------

func TestEmbeddingStoreRoundtripAndStaleness(t *testing.T) {
	frozenClock(t, frozenTime())
	db := filepath.Join(t.TempDir(), "cache.db")
	s, err := NewEmbeddingStore(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Put("linking", "https://a", "m", []float64{0.1, 0.2}, "h1"); err != nil {
		t.Fatal(err)
	}
	if err := s.Put("linking", "https://b", "m", []float64{0.3, 0.4}, "h2"); err != nil {
		t.Fatal(err)
	}
	n, _ := s.Count(strPtr("linking"))
	if n != 2 {
		t.Fatalf("count = %d", n)
	}
	loaded, err := s.Load("linking", "m")
	if err != nil {
		t.Fatal(err)
	}
	if len(loaded["https://a"].Vec) != 2 || loaded["https://a"].Vec[0] != 0.1 ||
		loaded["https://a"].TextHash != "h1" {
		t.Fatalf("loaded a = %+v", loaded["https://a"])
	}
	// upsert
	if err := s.Put("linking", "https://a", "m", []float64{0.9}, "h3"); err != nil {
		t.Fatal(err)
	}
	loaded, _ = s.Load("linking", "m")
	if len(loaded["https://a"].Vec) != 1 || loaded["https://a"].Vec[0] != 0.9 ||
		loaded["https://a"].TextHash != "h3" {
		t.Fatalf("upsert a = %+v", loaded["https://a"])
	}
	dropped, err := s.DropModel("linking", "m")
	if err != nil {
		t.Fatal(err)
	}
	if dropped != 2 {
		t.Fatalf("dropped = %d", dropped)
	}
	n, _ = s.Count(strPtr("linking"))
	if n != 0 {
		t.Fatalf("count after drop = %d", n)
	}
}

func TestEmbeddingStoreRefreshEmbeddingsStaleOnly(t *testing.T) {
	frozenClock(t, frozenTime())
	db := filepath.Join(t.TempDir(), "cache.db")
	s, err := NewEmbeddingStore(db)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	entries := []RefreshEntry{
		{SourceURL: "https://a", Fields: RecallFields{
			Name: "A", OneLine: "x", Recall: "Use when you need to a.", Tags: "t"}},
		{SourceURL: "https://b", Fields: RecallFields{
			Name: "B", OneLine: "y", Recall: "Use when you need to b.", Tags: "t"}},
	}
	embed := func(texts []string) ([][]float64, error) {
		vectors := make([][]float64, 0, len(texts))
		for range texts {
			vectors = append(vectors, []float64{1.0})
		}
		return vectors, nil
	}
	stats, err := RefreshEmbeddings(entries, embed, s, "linking", "m", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Embedded != 2 || stats.SkippedFresh != 0 || stats.Failed != 0 {
		t.Fatalf("stats = %+v", stats)
	}
	stats, _ = RefreshEmbeddings(entries, embed, s, "linking", "m", false, nil)
	if stats.SkippedFresh != 2 {
		t.Fatalf("skipped = %d", stats.SkippedFresh)
	}
	stats, _ = RefreshEmbeddings(entries, embed, s, "linking", "m", true, nil)
	if stats.Embedded != 2 {
		t.Fatalf("forced embedded = %d", stats.Embedded)
	}

	badEmbed := func(texts []string) ([][]float64, error) {
		return nil, &EmbeddingError{Msg: "down"}
	}
	entries2 := append(entries, RefreshEntry{
		SourceURL: "https://c", Fields: RecallFields{
			Name: "C", OneLine: "z", Recall: "Use when you need to c.", Tags: "t"}})
	stats, err = RefreshEmbeddings(entries2, badEmbed, s, "linking", "m", true, nil)
	if err != nil {
		t.Fatal(err)
	}
	if stats.Failed != 3 { // one batch, all fail
		t.Fatalf("failed = %d", stats.Failed)
	}
	// the OLD vectors survive a failed refresh
	loaded, _ := s.Load("linking", "m")
	if len(loaded) != 2 {
		t.Fatalf("old vectors lost: %d", len(loaded))
	}
}

// strPtr is the *string vault door's helper.
func strPtr(s string) *string { return &s }
