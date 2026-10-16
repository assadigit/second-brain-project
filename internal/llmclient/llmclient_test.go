// llmclient_test.go — T-30 characterization (Phase 3): the core half.
//
// Ported/landed suites:
//   - test_core.py: TestLLMClient (extract_json plain/fenced/embedded/
//     invalid, call_with_timeout fast/times-out/zero-is-direct,
//     response shapes, chat_with_timeout) — the object-shape legs ride
//     the wire shape (the documented ollama-py duality collapse).
//   - the pure golden corpus (llmclient_pure.golden.json): extract_json
//     battery, estimate_tokens, resolve_task_model, the constants.
package llmclient

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// test_core.py: TestLLMClient
// ---------------------------------------------------------------------------

func TestExtractJSONPlain(t *testing.T) {
	v, err := ExtractJSON(`{"a": 1}`)
	if err != nil {
		t.Fatal(err)
	}
	om := v.(*storage.OrderedMap)
	got, _ := om.Get("a")
	if got != int64(1) {
		t.Fatalf("extract_json plain = %v", got)
	}
}

func TestExtractJSONFenced(t *testing.T) {
	text := "Sure! Here is the JSON:\n```json\n{\"a\": 2}\n```\nDone."
	v, err := ExtractJSON(text)
	if err != nil {
		t.Fatal(err)
	}
	om := v.(*storage.OrderedMap)
	got, _ := om.Get("a")
	if got != int64(2) {
		t.Fatalf("extract_json fenced = %v", got)
	}
}

func TestExtractJSONEmbeddedInProse(t *testing.T) {
	text := "The result is {\"a\": {\"b\": 3}} as requested."
	v, err := ExtractJSON(text)
	if err != nil {
		t.Fatal(err)
	}
	om := v.(*storage.OrderedMap)
	inner, _ := om.Get("a")
	innerOm := inner.(*storage.OrderedMap)
	got, _ := innerOm.Get("b")
	if got != int64(3) {
		t.Fatalf("extract_json embedded = %v", got)
	}
}

func TestExtractJSONInvalidRaises(t *testing.T) {
	for _, text := range []string{"no json here at all", ""} {
		_, err := ExtractJSON(text)
		if err == nil {
			t.Fatalf("extract_json(%q) did not raise", text)
		}
		if err.Error() != "No valid JSON found in response" {
			t.Errorf("error = %q", err.Error())
		}
	}
}

func TestCallWithTimeoutFastFn(t *testing.T) {
	err := CallWithTimeout(5, func(ctx context.Context) error {
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}

func TestCallWithTimeoutTimesOut(t *testing.T) {
	start := time.Now()
	err := CallWithTimeout(0.2, func(ctx context.Context) error {
		time.Sleep(2 * time.Second)
		return errors.New("late")
	})
	if err == nil {
		t.Fatal("no timeout raised")
	}
	var te *TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("error type %T, want *TimeoutError", err)
	}
	if elapsed := time.Since(start); elapsed > 1500*time.Millisecond {
		t.Fatalf("timeout returned late: %v", elapsed)
	}
}

func TestCallWithTimeoutZeroIsDirect(t *testing.T) {
	called := false
	err := CallWithTimeout(0, func(ctx context.Context) error {
		called = true
		return nil
	})
	if err != nil || !called {
		t.Fatalf("zero timeout: err=%v called=%v", err, called)
	}
}

func TestTimeoutErrorMessage(t *testing.T) {
	// Python: f"LLM call did not respond within {timeout_s:.0f}s"
	cases := []struct {
		secs float64
		want string
	}{
		{300, "LLM call did not respond within 300s"},
		{0.2, "LLM call did not respond within 0s"},
		{1.5, "LLM call did not respond within 2s"},
	}
	for _, c := range cases {
		e := &TimeoutError{Seconds: c.secs}
		if e.Error() != c.want {
			t.Errorf("TimeoutError(%.1f) = %q, want %q", c.secs, e.Error(), c.want)
		}
	}
}

func TestResponseContentShapes(t *testing.T) {
	// The representable wire shapes (the ollama-py object/dict duality
	// collapses onto one wire shape — the documented divergence).
	if got := ResponseContent(&OllamaChatResponse{
		Message: &OllamaChatMessage{Content: "obj-content"}}); got != "obj-content" {
		t.Errorf("content = %q", got)
	}
	if got := ResponseContent(&OllamaChatResponse{
		Message: &OllamaChatMessage{Content: ""}}); got != "" {
		t.Errorf("empty content = %q", got)
	}
	if got := ResponseContent(&OllamaChatResponse{}); got != "" {
		t.Errorf("no message = %q", got)
	}
	if got := ResponseContent(nil); got != "" {
		t.Errorf("nil response = %q", got)
	}
}

func TestChatWithTimeout(t *testing.T) {
	// test_core.test_chat_with_timeout: kwargs echo through the client.
	fake := &fakeOllamaClient{}
	got, err := ChatWithTimeout(fake, 5, &OllamaChatRequest{
		Model: "m1", Messages: []Message{}})
	if err != nil {
		t.Fatal(err)
	}
	if got != `{"category": "Design"}` {
		t.Fatalf("chat_with_timeout = %q", got)
	}
	if fake.calls[0].Model != "m1" {
		t.Errorf("model = %q", fake.calls[0].Model)
	}
}

// ---------------------------------------------------------------------------
// The pure golden corpus
// ---------------------------------------------------------------------------

func loadPureGolden(t *testing.T) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata",
		"llmclient_pure.golden.json"))
	if err != nil {
		t.Fatal(err)
	}
	v, err := storage.DecodeJSONValue(raw)
	if err != nil {
		t.Fatal(err)
	}
	doc, _ := v.(*storage.OrderedMap)
	corpusV, _ := doc.Get("corpus")
	corpus, _ := corpusV.(*storage.OrderedMap)
	// the OrderedMap view keeps Python int-ness (42 ≠ 42.0) — the
	// golden drivers depend on it
	out := map[string]any{}
	for _, k := range corpus.Keys() {
		val, _ := corpus.Get(k)
		out[k] = val
	}
	return out
}

// omToMap converts a decoded *OrderedMap tree into plain Go maps so the
// canonical PyDumpsCompact (sorted keys) equals Python's sort_keys dump.
func omToMap(v any) any {
	switch tt := v.(type) {
	case *storage.OrderedMap:
		out := map[string]any{}
		for _, k := range tt.Keys() {
			val, _ := tt.Get(k)
			out[k] = omToMap(val)
		}
		return out
	case []any:
		out := make([]any, len(tt))
		for i, vv := range tt {
			out[i] = omToMap(vv)
		}
		return out
	default:
		return v
	}
}

func TestGoldenExtractJSONCorpus(t *testing.T) {
	corpus := loadPureGolden(t)["extract_json"].([]any)
	for _, entryAny := range corpus {
		entry := entryAny.(*storage.OrderedMap)
		text := goldenStr(mapGet(entry, "text"))
		wantOK, _ := mapGet(entry, "ok").(bool)
		v, err := ExtractJSON(text)
		if !wantOK {
			if err == nil {
				t.Errorf("extract_json(%q): expected failure", text)
			}
			continue
		}
		if err != nil {
			t.Errorf("extract_json(%q): %v", text, err)
			continue
		}
		want := goldenStr(mapGet(entry, "result"))
		if got := storage.PyDumpsCompact(omToMap(v)); got != want {
			t.Errorf("extract_json(%q) = %s, want %s", text, got, want)
		}
	}
}

func TestGoldenEstimateTokens(t *testing.T) {
	entries := loadPureGolden(t)["estimate_tokens"].([]any)
	for _, entryAny := range entries {
		entry := entryAny.(*storage.OrderedMap)
		want := pyInt(mapGet(entry, "want"))
		var messages []Message
		if raw, ok := mapGet(entry, "messages").([]any); ok {
			for _, m := range raw {
				messages = append(messages, goldenMessage(m))
			}
		}
		if got := EstimateTokens(messages); got != want {
			t.Errorf("estimate_tokens(%v) = %d, want %d", mapGet(entry, "messages"), got, want)
		}
	}
}

// goldenMessage converts a golden messages entry (dict or junk) into
// the typed Message — junk entries map to the empty message (the
// Python math only depends on the content string; nil content → "").
func goldenMessage(m any) Message {
	om, ok := asOM(m)
	if !ok {
		return Message{}
	}
	msg := Message{}
	if r, p := om.Get("role"); p {
		msg.Role = helpers.PyStr(r)
	}
	if c, p := om.Get("content"); p {
		if cs, ok := c.(string); ok {
			msg.Content = cs
		} else {
			msg.Content = helpers.PyStr(c)
		}
	}
	return msg
}

func TestGoldenResolveTaskModel(t *testing.T) {
	entries := loadPureGolden(t)["resolve_task_model"].([]any)
	for _, entryAny := range entries {
		entry := entryAny.(*storage.OrderedMap)
		want := goldenStr(mapGet(entry, "want"))
		got := ResolveTaskModel(goldenAny(mapGet(entry, "config")),
			goldenStr(mapGet(entry, "task")),
			goldenStr(mapGet(entry, "default")))
		if got != want {
			t.Errorf("resolve_task_model(%v, %v, %v) = %q, want %q",
				mapGet(entry, "config"), mapGet(entry, "task"),
				mapGet(entry, "default"), got, want)
		}
	}
}

func TestGoldenConstants(t *testing.T) {
	c := loadPureGolden(t)["constants"].(*storage.OrderedMap)
	checks := []struct {
		key string
		got any
	}{
		{"DEFAULT_TIMEOUT_S", DefaultTimeoutS},
		{"DEFAULT_LIST_TIMEOUT_S", DefaultListTimeoutS},
		{"DEFAULT_NUM_CTX", DefaultNumCtx},
		{"DEFAULT_MAX_OUTPUT_TOKENS", DefaultMaxOutputTokens},
		{"ANTHROPIC_FALLBACK_MAX_TOKENS", AnthropicFallbackMaxTokens},
		{"CLOUD_PROVIDER_LABEL", CloudProviderLabel},
		{"ANTHROPIC_VERSION_HEADER", AnthropicVersionHeader},
		{"LLAMACPP_PROVIDER_LABEL", LLamaCppProviderLabel},
		{"LLAMACPP_DEFAULT_BASE", LLamaCppDefaultBase},
		{"LLAMACPP_PROBE_TIMEOUT_S", LLamaCppProbeTimeoutS},
	}
	for _, c2 := range checks {
		want, _ := c.Get(c2.key)
		if !numericEqual(c2.got, want) {
			t.Errorf("constant %s = %v, want %v", c2.key, c2.got, want)
		}
	}
	// The hint/scan tuples.
	hintsV, _ := c.Get("ANTHROPIC_HOST_HINTS")
	hints, _ := hintsV.([]any)
	if len(hints) != len(AnthropicHostHints) {
		t.Fatalf("host hints len = %d", len(hints))
	}
	for i, h := range hints {
		if h.(string) != AnthropicHostHints[i] {
			t.Errorf("host hint %d = %q", i, h)
		}
	}
	portsV, _ := c.Get("LLAMACPP_SCAN_PORTS")
	ports, _ := portsV.([]any)
	if len(ports) != len(LLamaCppScanPorts) {
		t.Fatalf("scan ports len = %d", len(ports))
	}
	for i, p := range ports {
		if pyInt(p) != LLamaCppScanPorts[i] {
			t.Errorf("scan port %d = %v", i, p)
		}
	}
}

// numericEqual compares a Go constant against the decoded golden value
// (int64/float64) with numeric coercion.
func numericEqual(got, want any) bool {
	switch g := got.(type) {
	case int:
		switch w := want.(type) {
		case int64:
			return int64(g) == w
		case float64:
			return float64(g) == w
		}
	case float64:
		switch w := want.(type) {
		case int64:
			return g == float64(w)
		case float64:
			return g == w
		}
	case string:
		w, ok := want.(string)
		return ok && g == w
	}
	return reflect.DeepEqual(got, want)
}

func goldenStr(v any) string {
	if v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return helpers.PyStr(v)
}

// goldenAny keeps a decoded golden value as-is for `any`-typed params.
func goldenAny(v any) any { return v }

var _ = json.Marshal
var _ = strings.TrimSpace
