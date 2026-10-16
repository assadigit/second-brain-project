// anthropic_test.go — T-30 characterization (Phase 3): the Anthropic
// Messages-API half + the cloud router.
//
// Ported/landed suites:
//   - test_v0230.py: TestAnthropicChat (is_anthropic_url, headers+body,
//     max_tokens fallback, the json-mode system instruction, text-block
//     joining, the error object, HTTP errors, unreachable), TestAnthropic
//     ListModels (headers + HTTP error), TestCloudChatRouter (routes by
//     URL — observed on the wire through the /anthropic.com path trick —
//     and forwards the budget kwargs), TestPreflightCloud.
package llmclient

import (
	"errors"
	"strings"
	"testing"
)

var anthropicOKReply = route(200, map[string]any{
	"content": []any{map[string]any{
		"type": "text", "text": `{"summary": "ok"}`}}})

// ---------------------------------------------------------------------------
// TestAnthropicChat
// ---------------------------------------------------------------------------

func TestIsAnthropicURL(t *testing.T) {
	yes := []string{
		"https://api.anthropic.com/v1", "https://api.anthropic.com",
		"https://prefix.anthropic.com/v1"}
	for _, u := range yes {
		if !IsAnthropicURL(u) {
			t.Errorf("is_anthropic_url(%q) = false", u)
		}
	}
	no := []string{"https://api.openai.com/v1", "http://127.0.0.1:8080/v1", ""}
	for _, u := range no {
		if IsAnthropicURL(u) {
			t.Errorf("is_anthropic_url(%q) = true", u)
		}
	}
}

func TestAnthropicChatHeadersAndBody(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": anthropicOKReply})
	out, err := AnthropicChat(rs.root()+"/v1", "sk-ant-test",
		"claude-sonnet-4-5",
		[]Message{
			{Role: "system", Content: "be strict"},
			{Role: "user", Content: "analyze"}}, 10,
		&ChatOptions{MaxOutputTokens: 2048})
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"summary": "ok"}` {
		t.Fatalf("anthropic_chat = %q", out)
	}
	req := rs.captured()[0]
	if req.Path != "/v1/messages" {
		t.Errorf("path = %q", req.Path)
	}
	// The Anthropic auth header set — NOT a Bearer.
	if req.Headers["x-api-key"] != "sk-ant-test" {
		t.Errorf("x-api-key = %q", req.Headers["x-api-key"])
	}
	if req.Headers["anthropic-version"] != AnthropicVersionHeader {
		t.Errorf("anthropic-version = %q", req.Headers["anthropic-version"])
	}
	if _, has := req.Headers["authorization"]; has {
		t.Error("Authorization header sent on the Anthropic path")
	}
	if req.Body != `{"model": "claude-sonnet-4-5", "messages": [{"role": "user", "content": "analyze"}], "max_tokens": 2048, "temperature": 0.7, "system": "be strict"}` {
		t.Errorf("wire body = %s", req.Body)
	}
}

func TestAnthropicMaxTokensFallbackWhenUnset(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"content": []any{map[string]any{
				"type": "text", "text": "x"}}})})
	if _, err := AnthropicChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rs.captured()[0].Body, `"max_tokens": 4096`) {
		t.Errorf("fallback max_tokens missing: %s", rs.captured()[0].Body)
	}
}

func TestAnthropicJSONModeAddsTheSystemInstruction(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"content": []any{map[string]any{
				"type": "text", "text": "{}"}}})})
	if _, err := AnthropicChat(rs.root()+"/v1", "k", "m",
		[]Message{
			{Role: "system", Content: "sys"},
			{Role: "user", Content: "hi"}}, 10,
		&ChatOptions{JSONMode: true}); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(rs.captured()[0].Body, "ONLY a valid JSON object") {
		t.Errorf("system instruction missing: %s", rs.captured()[0].Body)
	}
}

func TestAnthropicJSONModeWithoutSystemAddsNothing(t *testing.T) {
	// The reference quirk: json_mode only appends to EXISTING system
	// text — no system message → no instruction at all.
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"content": []any{map[string]any{
				"type": "text", "text": "{}"}}})})
	if _, err := AnthropicChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10,
		&ChatOptions{JSONMode: true}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(rs.captured()[0].Body, "ONLY a valid JSON object") {
		t.Errorf("instruction added without system: %s", rs.captured()[0].Body)
	}
	if strings.Contains(rs.captured()[0].Body, `"system"`) {
		t.Errorf("system key sent with no system message: %s", rs.captured()[0].Body)
	}
}

func TestAnthropicJoinsTextBlocks(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "part one "},
				map[string]any{"type": "thinking", "thinking": "(internal)"},
				map[string]any{"type": "text", "text": "part two"}}})})
	out, err := AnthropicChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "part one \npart two" {
		t.Fatalf("joined = %q", out)
	}
}

func TestAnthropicErrorObjectRaises(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"type":  "error",
			"error": map[string]any{"message": "bad model"}})})
	_, err := AnthropicChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var bad *CloudLLMBadResponse
	if !errors.As(err, &bad) {
		t.Fatalf("error type %T, want CloudLLMBadResponse", err)
	}
	if !strings.Contains(err.Error(), "bad model") {
		t.Errorf("message = %q", err.Error())
	}
}

func TestAnthropicHTTPErrorRaises(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": route(401, map[string]any{"m": "no"})})
	_, err := AnthropicChat(rs.root()+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var httpErr *CloudLLMHTTPError
	if !errors.As(err, &httpErr) {
		t.Fatalf("error type %T, want CloudLLMHTTPError", err)
	}
	if httpErr.Code != 401 {
		t.Errorf("code = %d", httpErr.Code)
	}
}

func TestAnthropicUnreachableRaisesCloudLLMError(t *testing.T) {
	_, err := AnthropicChat(deadPortURL(t)+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 3, nil)
	var cloud *CloudLLMError
	if !errors.As(err, &cloud) {
		t.Fatalf("error type %T, want CloudLLMError", err)
	}
	if !strings.HasPrefix(err.Error(), "could not reach the Claude endpoint ") {
		t.Errorf("message = %q", err.Error())
	}
}

func TestAnthropicChatEmptyURL(t *testing.T) {
	_, err := AnthropicChat("", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	var cloud *CloudLLMError
	if !errors.As(err, &cloud) {
		t.Fatalf("error type %T", err)
	}
	if err.Error() != "no API URL configured for the Claude endpoint" {
		t.Errorf("message = %q", err.Error())
	}
}

func TestAnthropicChatOverBudgetWarning(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/messages": anthropicOKReply})
	warnings := []string{}
	_, err := AnthropicChat(rs.root()+"/v1", "k", "m",
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
	if !strings.HasPrefix(warnings[0], "prompt ≈10004 tokens may exceed the model's context window (llm_num_ctx=8192)") {
		t.Errorf("warning = %q", warnings[0])
	}
}

// ---------------------------------------------------------------------------
// TestAnthropicListModels
// ---------------------------------------------------------------------------

func TestAnthropicListModelsHeaders(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "claude-sonnet-4-5"},
				map[string]any{"id": "claude-opus-4-1"}}})})
	names, err := AnthropicListModels(rs.root()+"/v1", "sk-ant", 5)
	if err != nil {
		t.Fatal(err)
	}
	if len(names) != 2 || names[0] != "claude-sonnet-4-5" ||
		names[1] != "claude-opus-4-1" {
		t.Fatalf("names = %v", names)
	}
	req := rs.captured()[0]
	if req.Headers["x-api-key"] != "sk-ant" {
		t.Errorf("x-api-key = %q", req.Headers["x-api-key"])
	}
	if req.Headers["anthropic-version"] != AnthropicVersionHeader {
		t.Errorf("anthropic-version = %q", req.Headers["anthropic-version"])
	}
}

func TestAnthropicListModelsHTTPErrorRaises(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(401, map[string]any{"m": "no"})})
	_, err := AnthropicListModels(rs.root()+"/v1", "sk-ant", 5)
	var cloud *CloudLLMError
	if !errors.As(err, &cloud) {
		t.Fatalf("error type %T, want CloudLLMError", err)
	}
}

// ---------------------------------------------------------------------------
// TestCloudChatRouter + TestPreflightCloud (observed on the wire)
// ---------------------------------------------------------------------------

func TestCloudChatRoutesByURL(t *testing.T) {
	// The /anthropic.com path trick: the URL carries the anthropic.com
	// substring → the Messages wire; everything else → OpenAI wire.
	// Observable end-to-end, no mocks (the Python suite mocks the
	// providers; the port proves the routing on the wire).
	rs := newRecordingServer(t, map[string]any{
		"/anthropic.com/messages": anthropicOKReply,
		"/v1/chat/completions":    okChatReply})
	base := rs.root()
	out, err := CloudChat(base+"/anthropic.com", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != `{"summary": "ok"}` {
		t.Fatalf("anthropic route = %q", out)
	}
	out, err = CloudChat(base+"/v1", "k", "m",
		[]Message{{Role: "user", Content: "hi"}}, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	if out != "ok" {
		t.Fatalf("openai route = %q", out)
	}
	paths := []string{}
	for _, r := range rs.captured() {
		paths = append(paths, r.Path)
	}
	if len(paths) != 2 || paths[0] != "/anthropic.com/messages" ||
		paths[1] != "/v1/chat/completions" {
		t.Fatalf("paths = %v", paths)
	}
}

func TestCloudChatRouterForwardsBudgetKwargs(t *testing.T) {
	// num_ctx forwarded to the anthropic path → the over-budget warn
	// fires (the Python suite proves the kwargs via mocks; the port
	// proves the observable warning).
	rs := newRecordingServer(t, map[string]any{
		"/anthropic.com/messages": anthropicOKReply})
	warnings := []string{}
	_, err := CloudChat(rs.root()+"/anthropic.com", "k", "m",
		[]Message{{Content: strings.Repeat("x", 40000)}}, 10,
		&ChatOptions{JSONMode: true, NumCtx: 8192,
			MaxOutputTokens: 32000, OnWarn: func(m string) {
				warnings = append(warnings, m)
			}})
	if err != nil {
		t.Fatal(err)
	}
	if len(warnings) != 1 ||
		!strings.HasPrefix(warnings[0], "prompt ≈10004 tokens may exceed the model's context window (llm_num_ctx=8192)") {
		t.Fatalf("warnings = %v", warnings)
	}
	// max_tokens forwarded (the wire proves it).
	if !strings.Contains(rs.captured()[0].Body, `"max_tokens": 32000`) {
		t.Errorf("max_tokens not forwarded: %s", rs.captured()[0].Body)
	}
}

func TestPreflightCloudRoutesByURL(t *testing.T) {
	rs := newRecordingServer(t, map[string]any{
		"/anthropic.com/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "claude-x"}}}),
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "gpt-x"}}})})
	base := rs.root()
	ok, _, listed := PreflightCloud(base+"/anthropic.com", "k", "claude-x", 5)
	if !ok || listed == nil || !*listed {
		t.Fatalf("anthropic preflight = %v %v", ok, listed)
	}
	ok, _, listed = PreflightCloud(base+"/v1", "k", "gpt-x", 5)
	if !ok || listed == nil || !*listed {
		t.Fatalf("openai preflight = %v %v", ok, listed)
	}
	// The anthropic route hit the anthropic-shaped path (the wire is
	// the proof — exactly one request per route, each on its own path).
	paths := []string{}
	for _, r := range rs.captured() {
		paths = append(paths, r.Path)
	}
	if len(paths) != 2 || paths[0] != "/anthropic.com/models" ||
		paths[1] != "/v1/models" {
		t.Fatalf("paths = %v", paths)
	}
}
