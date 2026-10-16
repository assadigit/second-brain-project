// ollama_test.go — T-30 characterization (Phase 3): the Ollama half.
//
// Ported/landed suites:
//   - test_phase4.py: TestOllamaChat (options.num_ctx + format json on
//     every call, no options when num_ctx off, json_mode off, the
//     over-budget warning, the response shapes) + TestOllamaRealHTTP
//     (the REAL client against a fake local server — options.num_ctx
//     and format actually reach the wire; the port's own HTTP client
//     plays the SDK's role — the documented divergence), TestEstimate
//     Tokens scales-with-content (in llmclient_test.go).
//   - test_v0230.py: TestOutputTokens.test_ollama_chat_sends_num_predict
//     (num_predict only rides options WHEN num_ctx is on — the
//     reference's nested-branch quirk).
package llmclient

import (
	"context"
	"strings"
	"testing"
)

// fakeOllamaClient records chat requests; answers the wire shape (the
// Python suites' _FakeOllamaClient).
type fakeOllamaClient struct {
	calls []*OllamaChatRequest
}

func (f *fakeOllamaClient) Chat(_ context.Context, req *OllamaChatRequest) (*OllamaChatResponse, error) {
	f.calls = append(f.calls, req)
	return &OllamaChatResponse{
		Message: &OllamaChatMessage{Content: `{"category": "Design"}`}}, nil
}

func (f *fakeOllamaClient) List(_ context.Context) ([]string, error) {
	return nil, nil
}

func TestOllamaChatNumCtxAndFormatOnEveryCall(t *testing.T) {
	client := &fakeOllamaClient{}
	if _, err := OllamaChat(client, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&OllamaChatOptions{NumCtx: 8192}); err != nil {
		t.Fatal(err)
	}
	call := client.calls[0]
	if call.Options == nil || call.Options.NumCtx != 8192 {
		t.Fatalf("options = %+v", call.Options)
	}
	if call.Format != "json" {
		t.Errorf("format = %q", call.Format)
	}
}

func TestOllamaChatNoOptionsWhenNumCtxOff(t *testing.T) {
	client := &fakeOllamaClient{}
	for _, numCtx := range []int{0} {
		if _, err := OllamaChat(client, "m",
			[]Message{{Role: "user", Content: "hi"}}, 10,
			&OllamaChatOptions{NumCtx: numCtx}); err != nil {
			t.Fatal(err)
		}
	}
	// num_ctx nil → no options at all.
	if _, err := OllamaChat(client, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil); err != nil {
		t.Fatal(err)
	}
	for i, call := range client.calls {
		if call.Options != nil {
			t.Errorf("call %d carried options: %+v", i, call.Options)
		}
	}
}

func TestOllamaChatJSONModeOff(t *testing.T) {
	client := &fakeOllamaClient{}
	if _, err := OllamaChat(client, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&OllamaChatOptions{JSONModeOff: true}); err != nil {
		t.Fatal(err)
	}
	if client.calls[0].Format != "" {
		t.Errorf("format = %q, want empty", client.calls[0].Format)
	}
}

func TestOllamaChatJSONModeDefaultsOn(t *testing.T) {
	// Python's ollama_chat(json_mode=True) default: opts nil → JSON on.
	client := &fakeOllamaClient{}
	if _, err := OllamaChat(client, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil); err != nil {
		t.Fatal(err)
	}
	if client.calls[0].Format != "json" {
		t.Errorf("format = %q, want json (the default)", client.calls[0].Format)
	}
}

func TestOllamaChatOverBudgetWarningFires(t *testing.T) {
	client := &fakeOllamaClient{}
	warnings := []string{}
	if _, err := OllamaChat(client, "m",
		[]Message{{Content: strings.Repeat("x", 40000)}}, 10,
		&OllamaChatOptions{NumCtx: 512, OnWarn: func(m string) {
			warnings = append(warnings, m)
		}}); err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v", warnings)
	}
	if !strings.HasPrefix(warnings[0], "prompt ≈10004 tokens exceeds the context window (num_ctx=512)") {
		t.Errorf("warning = %q", warnings[0])
	}
}

func TestOllamaChatNoWarningWhenWithinBudget(t *testing.T) {
	client := &fakeOllamaClient{}
	warnings := []string{}
	if _, err := OllamaChat(client, "m",
		[]Message{{Content: strings.Repeat("x", 100)}}, 10,
		&OllamaChatOptions{NumCtx: 8192, OnWarn: func(m string) {
			warnings = append(warnings, m)
		}}); err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v", warnings)
	}
}

func TestOllamaChatBothResponseShapes(t *testing.T) {
	// The ollama-py old/new object duality is ONE wire shape (the
	// documented divergence) — the fake client answers the dict shape.
	client := &fakeOllamaClient{}
	out, err := OllamaChat(client, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"category": "Design"}` {
		t.Fatalf("out = %q", out)
	}
}

func TestOllamaChatSendsNumPredictWhenSet(t *testing.T) {
	client := &fakeOllamaClient{}
	if _, err := OllamaChat(client, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&OllamaChatOptions{NumCtx: 160000, NumPredict: 32000}); err != nil {
		t.Fatal(err)
	}
	o := client.calls[0].Options
	if o == nil || o.NumCtx != 160000 || o.NumPredict != 32000 {
		t.Fatalf("options = %+v", o)
	}
	// The reference quirk: num_predict rides ONLY inside the num_ctx
	// branch — num_predict without num_ctx is NOT sent.
	client2 := &fakeOllamaClient{}
	if _, err := OllamaChat(client2, "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&OllamaChatOptions{NumPredict: 32000}); err != nil {
		t.Fatal(err)
	}
	if client2.calls[0].Options != nil {
		t.Errorf("options without num_ctx = %+v", client2.calls[0].Options)
	}
}

// ---------------------------------------------------------------------------
// TestOllamaRealHTTP — the port's own client against a fake local
// Ollama server: options.num_ctx and format actually reach the wire.
// ---------------------------------------------------------------------------

func TestOllamaRealClientSendsNumCtx(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/chat": route(200, map[string]any{
			"model": "llama3",
			"message": map[string]any{
				"role": "assistant", "content": `{"category": "Design"}`}})})
	client := NewOllamaClient(rs.root())
	out, err := OllamaChat(client, "llama3",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&OllamaChatOptions{NumCtx: 4096})
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"category": "Design"}` {
		t.Fatalf("out = %q", out)
	}
	body := rs.captured()[0].Body
	if !strings.Contains(body, `"options": {"num_ctx": 4096}`) {
		t.Errorf("options on wire = %s", body)
	}
	if !strings.Contains(body, `"format": "json"`) {
		t.Errorf("format on wire = %s", body)
	}
	if !strings.Contains(body, `"model": "llama3"`) {
		t.Errorf("model on wire = %s", body)
	}
	if !strings.Contains(body, `"messages": [{"role": "user", "content": "hi"}]`) {
		t.Errorf("messages on wire = %s", body)
	}
}

func TestOllamaClientListNames(t *testing.T) {
	// list_models_with_timeout against the /api/tags wire shape: the
	// name key, the model key, both, neither — the name-or-model rule.
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, map[string]any{
			"models": []any{
				map[string]any{"name": "old-api"},
				map[string]any{"model": "old-api-2"}}})})
	client := NewOllamaClient(rs.root())
	names, err := ListModelsWithTimeout(client, 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "old-api" || names[1] != "old-api-2" {
		t.Fatalf("names = %v", names)
	}
}

func TestOllamaReachable(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/api/tags": route(200, map[string]any{"models": []any{}})})
	if !OllamaReachable(rs.root(), 1.5) {
		t.Error("reachable = false for a live server")
	}
	if OllamaReachable(deadPortURL(t), 0.5) {
		t.Error("reachable = true for a dead port")
	}
	// The default URL law: empty input → the Ollama default port.
	if OllamaReachable("", 0.2) {
		t.Error("reachable = true at the default URL (nothing listens)")
	}
}
