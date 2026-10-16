// openai_test.go — T-30 characterization (Phase 3): the
// OpenAI-compatible half.
//
// Ported/landed suites:
//   - test_phase4.py: TestOpenAIChatSuccess (plain/json-mode/empty
//     content/no auth header), TestOpenAIChatFallback (the
//     response_format 400 → one clean retry + memoization, the other
//     400s raise), TestOpenAIChatFailures (timeout → TimeoutError,
//     malformed body, error object, no choices, connection refused,
//     empty URL), TestModelsPreflight (listed/not-listed/hidden route/
//     unreachable), the num_predict leg (max_tokens).
//   - test_v0230.py: TestOutputTokens.test_openai_chat_sends_max_tokens
//   - test_llamacpp.py: TestProxySafeLoopback (list/chat legs).
package llmclient

import (
	"errors"
	"strings"
	"testing"
	"time"
)

var okChatReply = route(200, map[string]any{
	"choices": []any{map[string]any{
		"message": map[string]any{"content": "ok"}}}})

func newOpenAIServer(t *testing.T) *recordingServer {
	t.Helper()
	ResetJSONModeMemo()
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": okChatReply,
	})
	t.Cleanup(ResetJSONModeMemo)
	return rs
}

// ---------------------------------------------------------------------------
// TestOpenAIChatSuccess
// ---------------------------------------------------------------------------

func TestOpenAIChatSuccessPlain(t *testing.T) {
	rs := newOpenAIServer(t)
	out, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("openai_chat = %q", out)
	}
	req := rs.captured()[0]
	if req.Body != `{"model": "m", "messages": [{"role": "user", "content": "hi"}], "temperature": 0.7}` {
		t.Errorf("wire body = %s", req.Body)
	}
	if req.Headers["authorization"] != "Bearer k" {
		t.Errorf("auth = %q", req.Headers["authorization"])
	}
	if _, has := req.Headers["content-type"]; !has {
		t.Error("content-type missing")
	}
}

func TestOpenAIChatJSONModeSendsResponseFormat(t *testing.T) {
	rs := newOpenAIServer(t)
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&ChatOptions{JSONMode: true})
	if err != nil {
		t.Fatal(err)
	}
	body := rs.captured()[0].Body
	if !strings.Contains(body, `"response_format": {"type": "json_object"}`) {
		t.Errorf("response_format missing: %s", body)
	}
}

func TestOpenAIChatEmptyContentIsValidAnswer(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200, map[string]any{
			"choices": []any{map[string]any{
				"message": map[string]any{"content": ""}}}})})
	out, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "" {
		t.Fatalf("empty content = %q", out)
	}
}

func TestOpenAIChatEmptyKeySendsNoAuthHeader(t *testing.T) {
	rs := newOpenAIServer(t)
	_, err := OpenAIChat(rs.root()+"/v1", "", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if auth, has := rs.captured()[0].Headers["authorization"]; has {
		t.Errorf("authorization header present: %q", auth)
	}
}

// ---------------------------------------------------------------------------
// TestOpenAIChatFallback
// ---------------------------------------------------------------------------

func rejectResponseFormat(parsed map[string]any) (int, any) {
	if _, has := parsed["response_format"]; has {
		return 400, map[string]any{
			"error": map[string]any{"message": "response_format is not supported"}}
	}
	return 200, map[string]any{
		"choices": []any{map[string]any{
			"message": map[string]any{"content": "ok"}}}}
}

func newRejectServer(t *testing.T) *recordingServer {
	t.Helper()
	ResetJSONModeMemo()
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": rejectResponseFormat,
	})
	t.Cleanup(ResetJSONModeMemo)
	return rs
}

func TestOpenAIChatRejectionFallsBackCleanly(t *testing.T) {
	rs := newRejectServer(t)
	out, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&ChatOptions{JSONMode: true})
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("fallback reply = %q", out)
	}
	bodies := []string{}
	for _, r := range rs.captured() {
		bodies = append(bodies, r.Body)
	}
	if !strings.Contains(bodies[0], "response_format") {
		t.Errorf("first attempt without response_format: %s", bodies[0])
	}
	if strings.Contains(bodies[1], "response_format") {
		t.Errorf("retry kept response_format: %s", bodies[1])
	}
}

func TestOpenAIChatFallbackIsMemoized(t *testing.T) {
	rs := newRejectServer(t)
	msgs := []Message{{Role: "user", Content: "hi"}}
	base := rs.root() + "/v1"
	if _, err := OpenAIChat(base, "k", "m", msgs, 10, &ChatOptions{JSONMode: true}); err != nil {
		t.Fatal(err)
	}
	nAfterFirst := len(rs.captured())
	// Second call: NO doomed attempt — the memo skips response_format.
	if _, err := OpenAIChat(base, "k", "m", msgs, 10, &ChatOptions{JSONMode: true}); err != nil {
		t.Fatal(err)
	}
	if got := len(rs.captured()); got != nAfterFirst+1 {
		t.Fatalf("requests after second call = %d, want %d", got, nAfterFirst+1)
	}
	if strings.Contains(rs.captured()[len(rs.captured())-1].Body, "response_format") {
		t.Error("memoized call still sent response_format")
	}
	if !JSONModeRejected(base) {
		t.Error("json_mode_rejected(base) = false")
	}
}

func TestOpenAIChatOther400WithoutJSONModeRaises(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(400, map[string]any{
			"error": map[string]any{"message": "bad request"}})})
	ResetJSONModeMemo()
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var httpErr *CloudLLMHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error type %T, want CloudLLMHTTPError", err)
	}
	if httpErr.Code != 400 {
		t.Errorf("code = %d", httpErr.Code)
	}
}

func TestOpenAIChat400WithJSONModeThatKeepsFailingRaises(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(400, map[string]any{
			"error": map[string]any{"message": "bad request"}})})
	ResetJSONModeMemo()
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&ChatOptions{JSONMode: true})
	var httpErr *CloudLLMHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error type %T, want CloudLLMHTTPError", err)
	}
	// Both attempts happened (with, then without).
	if got := len(rs.captured()); got != 2 {
		t.Fatalf("attempts = %d, want 2", got)
	}
}

// ---------------------------------------------------------------------------
// TestOpenAIChatFailures
// ---------------------------------------------------------------------------

func TestOpenAIChatTimeoutRaisesTimeoutError(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": slowReply})
	t0 := time.Now()
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 1, nil)
	var te *TimeoutError
	if !errors.As(err, &te) {
		t.Fatalf("error type %T (%v), want TimeoutError", err, err)
	}
	if elapsed := time.Since(t0); elapsed >= 4*time.Second {
		t.Fatalf("wall clock not honored: %v", elapsed)
	}
}

func TestOpenAIChatMalformedBody(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200,
			[]byte(`{"choices": [{"mes`))})
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var bad *CloudLLMBadResponse
	if !errors.As(err, &bad) {
		t.Fatalf("error type %T, want CloudLLMBadResponse", err)
	}
	if !strings.Contains(err.Error(), "non-JSON") {
		t.Errorf("message = %q", err.Error())
	}
}

func TestOpenAIChatErrorObjectBody(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200, map[string]any{
			"error": map[string]any{"message": "boom from server"}})})
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var bad *CloudLLMBadResponse
	if !errors.As(err, &bad) {
		t.Fatalf("error type %T, want CloudLLMBadResponse", err)
	}
	if !strings.Contains(err.Error(), "boom from server") {
		t.Errorf("message = %q", err.Error())
	}
}

func TestOpenAIChatNoChoicesBody(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200, map[string]any{
			"choices": []any{}})})
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var bad *CloudLLMBadResponse
	if !errors.As(err, &bad) {
		t.Fatalf("error type %T, want CloudLLMBadResponse", err)
	}
	if err.Error() != "endpoint returned no choices" {
		t.Errorf("message = %q", err.Error())
	}
}

func TestOpenAIChatConnectionRefused(t *testing.T) {
	_, err := OpenAIChat("http://127.0.0.1:9/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 5, nil)
	var cloud *CloudLLMError
	if !errors.As(err, &cloud) {
		t.Fatalf("error type %T, want CloudLLMError", err)
	}
	if !strings.HasPrefix(err.Error(), "could not reach the OpenAI-compatible endpoint http://127.0.0.1:9/v1: ") {
		t.Errorf("message = %q", err.Error())
	}
}

func TestOpenAIChatEmptyURL(t *testing.T) {
	_, err := OpenAIChat("", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 5, nil)
	var cloud *CloudLLMError
	if !errors.As(err, &cloud) {
		t.Fatalf("error type %T, want CloudLLMError", err)
	}
	if err.Error() != "no API URL configured for the OpenAI-compatible endpoint" {
		t.Errorf("message = %q", err.Error())
	}
}

func TestOpenAIChatHTTPErrorMessageLaw(t *testing.T) {
	// CloudLLMHTTPError(code, body): message truncates to 200 chars of
	// the 400-char body — the Python law.
	e := NewCloudLLMHTTPError(500, strings.Repeat("x", 500))
	if len(e.Body) != 400 {
		t.Errorf("body kept %d chars, want 400", len(e.Body))
	}
	want := "endpoint answered HTTP 500: " + strings.Repeat("x", 200)
	if e.Error() != want {
		t.Errorf("message law broken: len=%d", len(e.Error()))
	}
}

func TestOpenAIChatOverBudgetWarning(t *testing.T) {
	rs := newOpenAIServer(t)
	warnings := []string{}
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Content: strings.Repeat("x", 40000)}}, 10,
		&ChatOptions{NumCtx: 8192, OnWarn: func(m string) {
			warnings = append(warnings, m)
		}})
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 {
		t.Fatalf("warnings = %v", warnings)
	}
	if !strings.HasPrefix(warnings[0], "prompt ≈10004 tokens may exceed the server's context window (llm_num_ctx=8192)") {
		t.Errorf("warning = %q", warnings[0])
	}
}

func TestOpenAIChatNoWarningWhenWithinBudget(t *testing.T) {
	rs := newOpenAIServer(t)
	warnings := []string{}
	_, err := OpenAIChat(rs.root()+"/v1", "k", "m",
		[]Message{{Content: strings.Repeat("x", 100)}}, 10,
		&ChatOptions{NumCtx: 8192, OnWarn: func(m string) {
			warnings = append(warnings, m)
		}})
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 0 {
		t.Errorf("warnings = %v", warnings)
	}
}

// ---------------------------------------------------------------------------
// TestModelsPreflight
// ---------------------------------------------------------------------------

func TestPreflightOpenAIModelListed(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "a"},
				map[string]any{"id": "b"}}})})
	ok, msg, listed := PreflightOpenAI(rs.root()+"/v1", "k", "a", 5)
	if !ok || listed == nil || !*listed {
		t.Fatalf("preflight = %v %v %v", ok, msg, listed)
	}
	if msg != "2 model(s) listed" {
		t.Errorf("msg = %q", msg)
	}
}

func TestPreflightOpenAIModelNotListed(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "a"},
				map[string]any{"id": "b"}}})})
	ok, _, listed := PreflightOpenAI(rs.root()+"/v1", "k", "zzz", 5)
	if !ok || listed == nil || *listed {
		t.Fatalf("preflight = %v %v", ok, listed)
	}
}

func TestPreflightOpenAIHiddenRouteIsNotAnError(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(404, map[string]any{"error": "not found"})})
	ok, msg, listed := PreflightOpenAI(rs.root()+"/v1", "k", "a", 5)
	if ok {
		t.Fatal("hidden route reported ok")
	}
	if listed != nil {
		t.Errorf("listed = %v, want nil", listed)
	}
	if msg != "GET "+rs.root()+"/v1/models answered HTTP 404" {
		t.Errorf("msg = %q", msg)
	}
}

func TestPreflightOpenAIUnreachableEndpoint(t *testing.T) {
	ok, _, listed := PreflightOpenAI("http://127.0.0.1:9/v1", "k", "a", 5)
	if ok || listed != nil {
		t.Fatalf("preflight = %v %v", ok, listed)
	}
}

func TestOpenAIListModelsNames(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "a"},
				map[string]any{"id": "b"}}})})
	names, err := OpenAIListModels(rs.root()+"/v1", "k", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "a" || names[1] != "b" {
		t.Fatalf("names = %v", names)
	}
}

// ---------------------------------------------------------------------------
// test_v0230.py: TestOutputTokens (the openai leg)
// ---------------------------------------------------------------------------

func TestOpenAIChatSendsMaxTokensWhenSet(t *testing.T) {
	rs := newOpenAIServer(t)
	msgs := []Message{{Role: "user", Content: "hi"}}
	if _, err := OpenAIChat(rs.root()+"/v1", "", "m", msgs, 10,
		&ChatOptions{MaxOutputTokens: 32000}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rs.captured()[0].Body, `"max_tokens": 32000`) {
		t.Errorf("max_tokens missing: %s", rs.captured()[0].Body)
	}
	// Unset → NOT sent (the server default stands).
	if _, err := OpenAIChat(rs.root()+"/v1", "", "m", msgs, 10, nil); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(rs.captured()[1].Body, "max_tokens") {
		t.Errorf("max_tokens sent when unset: %s", rs.captured()[1].Body)
	}
}

func TestOpenAIChatExplicitTemperature(t *testing.T) {
	rs := newOpenAIServer(t)
	if _, err := OpenAIChat(rs.root()+"/v1", "", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&ChatOptions{Temperature: 0.2}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rs.captured()[0].Body, `"temperature": 0.2`) {
		t.Errorf("temperature = %s", rs.captured()[0].Body)
	}
}

// ---------------------------------------------------------------------------
// test_llamacpp.py: TestProxySafeLoopback (the list/chat legs)
// ---------------------------------------------------------------------------

func TestOpenAIPathsIgnoreProxyEnv(t *testing.T) {
	t.Setenv("HTTP_PROXY", deadProxyURL(t))
	t.Setenv("HTTPS_PROXY", deadProxyURL(t))
	t.Setenv("http_proxy", deadProxyURL(t))
	t.Setenv("https_proxy", deadProxyURL(t))
	// list models against a LOOPBACK server must bypass the proxy
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "qwen2.5-3b"}}})})
	names, err := OpenAIListModels(rs.root()+"/v1", "", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 1 || names[0] != "qwen2.5-3b" {
		t.Fatalf("names = %v", names)
	}
	// chat against a LOOPBACK server must bypass the proxy too
	rs2 := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": okChatReply})
	out, err := OpenAIChat(rs2.root()+"/v1", "", "qwen2.5-3b",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("chat = %q", out)
	}
}

func deadProxyURL(t *testing.T) string {
	return deadPortURL(t)
}

// slowReply stalls past the wall-clock timeout (the sleep_s behavior).
func slowReply(_ map[string]any) (int, any) {
	time.Sleep(5 * time.Second)
	return 200, map[string]any{
		"choices": []any{map[string]any{
			"message": map[string]any{"content": "ok"}}}}
}
