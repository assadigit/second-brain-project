package connectioncheck

// llmsection_test.go — section 2: the ACTIVE provider routes against
// local recording servers (ollama / cloud / llama.cpp shapes) + the
// UA table's _http_get_json entry. Ported from the reference's
// TestCheckOllama, TestCheckCloud, TestCheckLlamacpp.

import (
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/llmclient"
)

func TestCheckOllamaUpWithConfiguredModel(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "llama3:latest"},
			map[string]any{"name": "qwen:7b"},
		}}),
	})
	r := CheckOllama(srv.root(), "qwen:7b")
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "2 model(s)") {
		t.Errorf("detail: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, "'qwen:7b' ready") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckOllamaConfiguredModelNotPulledIsWarn(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "llama3:latest"},
		}}),
	})
	r := CheckOllama(srv.root(), "mistral")
	if r.Level != LevelWarn {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "NOT pulled") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckOllamaNoModelsIsError(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{}}),
	})
	r := CheckOllama(srv.root(), "")
	if r.Level != LevelError {
		t.Fatalf("level = %s", r.Level)
	}
	if !strings.Contains(r.Detail, "NO models") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckOllamaDeadAddressIsError(t *testing.T) {
	r := CheckOllama(deadURL(t), "")
	if r.Level != LevelError {
		t.Fatalf("level = %s", r.Level)
	}
	if !strings.Contains(r.Detail, "not reachable") {
		t.Errorf("detail: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, "URLError") {
		t.Errorf("detail lacks the Python transport class: %q", r.Detail)
	}
}

func TestCheckOllamaModelFallsBackToModelKey(t *testing.T) {
	// the name-or-model law: entries may carry either key
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"model": "gemma:2b"},
		}}),
	})
	r := CheckOllama(srv.root(), "gemma:2b")
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "'gemma:2b' ready") {
		t.Errorf("detail: %q", r.Detail)
	}
}

// THE UA TABLE (1/3): _http_get_json speaks
// "GitCurator-connection-check" — pinned from the wire.
func TestCheckOllamaUATableEntry(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "llama3:latest"},
		}}),
	})
	_ = CheckOllama(srv.root(), "")
	captured := srv.captured()
	if len(captured) != 1 {
		t.Fatalf("requests = %d, want 1", len(captured))
	}
	req := captured[0]
	if req.UA != "GitCurator-connection-check" {
		t.Errorf("UA = %q, want GitCurator-connection-check", req.UA)
	}
	if req.Method != "GET" {
		t.Errorf("method = %s, want GET", req.Method)
	}
	if req.BodyN != 0 {
		t.Errorf("GET carried a body: %d bytes", req.BodyN)
	}
}

func TestCheckCloudNoURLIsError(t *testing.T) {
	r := CheckCloud(map[string]any{})
	if r.Level != LevelError {
		t.Fatalf("level = %s", r.Level)
	}
	if !strings.Contains(r.Detail, "no API URL") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckCloudModelListedIsOk(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/v1/models": route(200, map[string]any{
			"object": "list",
			"data":   []any{map[string]any{"id": "zai-glm"}},
		}),
	})
	r := CheckCloud(map[string]any{
		"cloud_api_url": srv.root() + "/v1",
		"cloud_api_key": "", "cloud_model": "zai-glm"})
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "'zai-glm' listed") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckCloudModelNotListedIsWarn(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/v1/models": route(200, map[string]any{
			"object": "list",
			"data":   []any{map[string]any{"id": "other"}},
		}),
	})
	r := CheckCloud(map[string]any{
		"cloud_api_url": srv.root() + "/v1",
		"cloud_api_key": "", "cloud_model": "zai-glm"})
	if r.Level != LevelWarn {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "NOT in the list") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckCloudModelListHiddenIsOk(t *testing.T) {
	// a 200 with an EMPTY model list → model_listed None → the
	// hidden-list OK note (endpoint answered; nothing to compare)
	srv := newJSONServer(t, map[string][2]any{
		"/v1/models": route(200, map[string]any{
			"object": "list", "data": []any{}}),
	})
	r := CheckCloud(map[string]any{
		"cloud_api_url": srv.root() + "/v1",
		"cloud_api_key": "", "cloud_model": "m"})
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "model list hidden") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckCloudListEndpoint404IsError(t *testing.T) {
	// a 404 on /v1/models is a failed preflight, NOT the hidden-list
	// note (the never-a-gate preflight still reports endpoint failure)
	srv := newJSONServer(t, map[string][2]any{
		"/v1/models": route(404, map[string]any{"message": "Not Found"}),
	})
	r := CheckCloud(map[string]any{
		"cloud_api_url": srv.root() + "/v1",
		"cloud_api_key": "", "cloud_model": "m"})
	if r.Level != LevelError {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
}

func TestCheckCloudDeadEndpointIsError(t *testing.T) {
	r := CheckCloud(map[string]any{
		"cloud_api_url": deadURL(t) + "/v1",
		"cloud_api_key": "", "cloud_model": "m"})
	if r.Level != LevelError {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
}

func llamaCppFake(t *testing.T, health [2]any) *fakeJSONServer {
	t.Helper()
	return newJSONServer(t, map[string][2]any{
		"/props": route(200, map[string]any{
			"model_path":  "models/qwen2.5-3b-instruct-q4_k_m.gguf",
			"model_alias": "qwen2.5-3b",
			"total_slots": 1,
			"default_generation_settings": map[string]any{
				"temperature": 0.8, "n_predict": -1},
		}),
		"/health": health,
		"/v1/models": route(200, map[string]any{
			"object": "list",
			"data": []any{map[string]any{
				"id": "qwen2.5-3b", "object": "model",
				"owned_by": "llama.cpp"}},
		}),
	})
}

func TestCheckLlamaCppUpWithModelIsOk(t *testing.T) {
	srv := llamaCppFake(t, route(200, map[string]any{"status": "ok"}))
	r := CheckLlamaCpp(map[string]any{
		"llamacpp_api_url": srv.root()})
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "model 'qwen2.5-3b'") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckLlamaCppStillLoadingNoted(t *testing.T) {
	srv := llamaCppFake(t, route(503, map[string]any{"error": "loading model"}))
	r := CheckLlamaCpp(map[string]any{
		"llamacpp_api_url": srv.root()})
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "still loading") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckLlamaCppLimitsGrabRidesTheLine(t *testing.T) {
	// v0.70.0 — the server config grab: /props limits speak in the
	// Test-Connection line
	srv := newJSONServer(t, map[string][2]any{
		"/props": route(200, map[string]any{
			"model_path":  "models/qwen2.5-3b-instruct-q4_k_m.gguf",
			"model_alias": "qwen2.5-3b",
			"total_slots": 1,
			"default_generation_settings": map[string]any{
				"n_ctx": 8192, "n_predict": 512, "temperature": 0.8},
		}),
		"/health": route(200, map[string]any{"status": "ok"}),
		"/v1/models": route(200, map[string]any{
			"object": "list",
			"data": []any{map[string]any{
				"id": "qwen2.5-3b", "object": "model",
				"owned_by": "llama.cpp"}},
		}),
	})
	r := CheckLlamaCpp(map[string]any{
		"llamacpp_api_url": srv.root()})
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	for _, want := range []string{"context window 8192 tokens",
		"output cap 512 tokens"} {
		if !strings.Contains(r.Detail, want) {
			t.Errorf("detail lacks %q: %q", want, r.Detail)
		}
	}
}

func TestCheckLlamaCppNothingRunningIsErrorWithCommandHint(t *testing.T) {
	// the port scan is llmclient's business (79 tests there); here we
	// pin it to "nothing" so the check reports the clear error.
	withSeam(t, &detectLlamaCppFn,
		func(string, []int, string, float64, bool) *llmclient.ProbeResult {
			return nil
		})
	r := CheckLlamaCpp(map[string]any{
		"llamacpp_api_url": deadURL(t) + "/v1"})
	if r.Level != LevelError {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "llama-server -m") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckLLMHintLineWhenProviderDown(t *testing.T) {
	withSeam(t, &detectLlamaCppFn,
		func(string, []int, string, float64, bool) *llmclient.ProbeResult {
			return nil
		})
	out := CheckLLM(map[string]any{
		"llm_provider":     "llamacpp",
		"llamacpp_api_url": deadURL(t) + "/v1"})
	if len(out) != 2 {
		t.Fatalf("rows = %d, want 2 (the error + the alternatives hint)", len(out))
	}
	if out[1].Name != "Hint" || out[1].Level != LevelInfo {
		t.Errorf("hint row = (%s, %s)", out[1].Name, out[1].Level)
	}
}

func TestCheckLLMProviderDefaultIsOllama(t *testing.T) {
	srv := newJSONServer(t, map[string][2]any{
		"/api/tags": route(200, map[string]any{"models": []any{
			map[string]any{"name": "llama3:latest"},
		}}),
	})
	out := CheckLLM(map[string]any{
		"ollama": map[string]any{
			"base_url": srv.root(), "model": "llama3:latest"}})
	if len(out) != 1 {
		t.Fatalf("rows = %d, want 1", len(out))
	}
	if out[0].Name != "Ollama" || out[0].Level != LevelOK {
		t.Errorf("row = (%s, %s) (%s)", out[0].Name, out[0].Level,
			out[0].Detail)
	}
}
