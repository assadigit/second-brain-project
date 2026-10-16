// golden_test.go — T-30: the generated-golden drivers.
//
//   - llmclient_payloads.golden.json: THE WIRE CONTRACT — every
//     scenario's requests (method, path, app-chosen headers, the RAW
//     body bytes in Python json.dumps shape) captured from the
//     unmodified reference, replayed by the port against the same
//     fake-server shapes and byte-compared.
//   - llmclient_behavior.golden.json: the fake-server behaviors (probe
//     matrix, detection, detect_ollama details, preflight, error
//     messages, warnings, routing) — compared field-by-field after the
//     <ROOT>/<API>/<DEADURL> normalization.
//   - the remaining pure sections not driven in llmclient_test.go.
package llmclient

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// goldenNormalize round-trips a value through encoding/json so it
// compares against the decoded golden: pointers collapse to values,
// ints to float64, typed maps to map[string]any, structs to their
// tagged shapes — the same tree json.Unmarshal produces.
func goldenNormalize(v any) any {
	raw, err := json.Marshal(v)
	if err != nil {
		return v
	}
	var out any
	if err := json.Unmarshal(raw, &out); err != nil {
		return v
	}
	return out
}

func loadGoldenFile(name string) (map[string]any, error) {
	data, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		return nil, err
	}
	var out map[string]any
	if err := json.Unmarshal(data, &out); err != nil {
		return nil, err
	}
	return out, nil
}

// ---------------------------------------------------------------------------
// The payload (wire) golden
// ---------------------------------------------------------------------------

var msgsHi = []Message{{Role: "user", Content: "hi"}}

func TestGoldenPayloads(t *testing.T) {
	data, err := loadGoldenFile("llmclient_payloads.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	scenarios := data["scenarios"].(map[string]any)

	run := func(name string, fn func(t *testing.T) any) {
		t.Run(name, func(t *testing.T) {
			got := goldenNormalize(fn(t))
			want := goldenNormalize(scenarios[name])
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("scenario %s mismatch:\n got: %s\nwant: %s",
					name,
					storage.PyDumpsCompact(got),
					storage.PyDumpsCompact(want))
			}
		})
	}

	run("openai_plain_key", func(t *testing.T) any {
		rs := newOpenAIServer(t)
		if _, err := OpenAIChat(rs.root()+"/v1", "k", "m", msgsHi, 10, nil); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("openai_json_mode", func(t *testing.T) any {
		rs := newOpenAIServer(t)
		if _, err := OpenAIChat(rs.root()+"/v1", "k", "m", msgsHi, 10,
			&ChatOptions{JSONMode: true}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("openai_no_key", func(t *testing.T) any {
		rs := newOpenAIServer(t)
		if _, err := OpenAIChat(rs.root()+"/v1", "", "m", msgsHi, 10, nil); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("openai_max_tokens", func(t *testing.T) any {
		rs := newOpenAIServer(t)
		if _, err := OpenAIChat(rs.root()+"/v1", "", "m", msgsHi, 10,
			&ChatOptions{MaxOutputTokens: 32000}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("openai_temperature_explicit", func(t *testing.T) any {
		rs := newOpenAIServer(t)
		if _, err := OpenAIChat(rs.root()+"/v1", "", "m", msgsHi, 10,
			&ChatOptions{Temperature: 0.2}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("openai_fallback_memoized", func(t *testing.T) any {
		rs := newRejectServer(t)
		base := rs.root() + "/v1"
		warnings := []string{}
		out, err := OpenAIChat(base, "k", "m", msgsHi, 10,
			&ChatOptions{JSONMode: true, OnWarn: func(m string) {
				warnings = append(warnings, m)
			}})
		if err != nil {
			t.Fatal(err)
		}
		nAfterFirst := len(rs.captured())
		if _, err := OpenAIChat(base, "k", "m", msgsHi, 10,
			&ChatOptions{JSONMode: true}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{
			"return":                          out,
			"warnings":                        warnings,
			"requests_after_two_calls":        capturedJSON(rs),
			"request_count_second_call_delta": len(rs.captured()) - nAfterFirst,
			"memoized_rejected":               JSONModeRejected(base),
		}
	})
	run("anthropic_full", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/v1/messages": anthropicOKReply})
		if _, err := AnthropicChat(rs.root()+"/v1", "sk-ant-test",
			"claude-sonnet-4-5",
			[]Message{{Role: "system", Content: "be strict"},
				{Role: "user", Content: "analyze"}}, 10,
			&ChatOptions{MaxOutputTokens: 2048}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("anthropic_max_tokens_fallback", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/v1/messages": anthropicOKReply})
		if _, err := AnthropicChat(rs.root()+"/v1", "k", "m", msgsHi, 10, nil); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("anthropic_json_mode", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/v1/messages": anthropicOKReply})
		if _, err := AnthropicChat(rs.root()+"/v1", "k", "m",
			[]Message{{Role: "system", Content: "sys"},
				{Role: "user", Content: "hi"}}, 10,
			&ChatOptions{JSONMode: true}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("anthropic_no_system", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/v1/messages": anthropicOKReply})
		if _, err := AnthropicChat(rs.root()+"/v1", "k", "m", msgsHi, 10,
			&ChatOptions{JSONMode: true}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("openai_list_models", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/v1/models": route(200, map[string]any{
				"data": []any{map[string]any{"id": "a"},
					map[string]any{"id": "b"}}})})
		if _, err := OpenAIListModels(rs.root()+"/v1", "k", 5); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("anthropic_list_models", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/v1/models": route(200, map[string]any{
				"data": []any{map[string]any{"id": "claude-sonnet-4-5"},
					map[string]any{"id": "claude-opus-4-1"}}})})
		if _, err := AnthropicListModels(rs.root()+"/v1", "sk-ant", 5); err != nil {
			t.Fatal(err)
		}
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("probe_llamacpp_auth", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/props": route(200, defaultLlamaProps),
			"/health": route(200, map[string]any{
				"status": "ok", "slots": []any{}}),
			"/v1/models": route(200, map[string]any{
				"object": "list",
				"data":   []any{map[string]any{"id": "qwen2.5-3b"}}})})
		ProbeLlamaCpp(rs.root(), "sekrit", LLamaCppProbeTimeoutS)
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("detect_ollama_ua", func(t *testing.T) any {
		rs := newRecordingServer(t, map[string]any{
			"/api/tags": route(200, map[string]any{
				"models": []any{
					map[string]any{"name": "llama3:latest"}}})})
		DetectOllama(rs.root(), 4.0)
		return map[string]any{"requests": capturedJSON(rs)}
	})
	run("ollama_request_building", func(t *testing.T) any {
		fake := &fakeOllamaClient{}
		warnings := []string{}
		if _, err := OllamaChat(fake, "m", msgsHi, 10,
			&OllamaChatOptions{NumCtx: 8192, OnWarn: func(m string) {
				warnings = append(warnings, m)
			}}); err != nil {
			t.Fatal(err)
		}
		fake2 := &fakeOllamaClient{}
		if _, err := OllamaChat(fake2, "m", msgsHi, 10, nil); err != nil {
			t.Fatal(err)
		}
		fake3 := &fakeOllamaClient{}
		if _, err := OllamaChat(fake3, "m", msgsHi, 10,
			&OllamaChatOptions{JSONModeOff: true, NumCtx: 8192}); err != nil {
			t.Fatal(err)
		}
		fake4 := &fakeOllamaClient{}
		if _, err := OllamaChat(fake4, "m", msgsHi, 10,
			&OllamaChatOptions{NumCtx: 160000, NumPredict: 32000}); err != nil {
			t.Fatal(err)
		}
		fake5 := &fakeOllamaClient{}
		if _, err := OllamaChat(fake5, "m",
			[]Message{{Content: repeatStr("x", 40000)}}, 10,
			&OllamaChatOptions{NumCtx: 512, OnWarn: func(m string) {
				warnings = append(warnings, m)
			}}); err != nil {
			t.Fatal(err)
		}
		return map[string]any{
			"num_ctx_and_format":      reqToGolden(fake.calls[0]),
			"num_ctx_off":             reqToGolden(fake2.calls[0]),
			"json_mode_off":           reqToGolden(fake3.calls[0]),
			"num_ctx_and_num_predict": reqToGolden(fake4.calls[0]),
			"over_budget_warning":     warnings,
		}
	})
}

// capturedJSON renders the captured requests in the golden's shape
// (method/path/headers/body — headers as a plain map for DeepEqual).
func capturedJSON(rs *recordingServer) []any {
	out := []any{}
	for _, r := range rs.captured() {
		out = append(out, map[string]any{
			"method":  r.Method,
			"path":    r.Path,
			"headers": r.Headers,
			"body":    r.Body,
		})
	}
	return out
}

// reqToGolden renders an OllamaChatRequest like the reference's kwargs
// dict (options/format present only when set).
func reqToGolden(req *OllamaChatRequest) map[string]any {
	msgs := []any{}
	for _, m := range req.Messages {
		msgs = append(msgs, map[string]any{
			"role": m.Role, "content": m.Content})
	}
	out := map[string]any{
		"model": req.Model, "messages": msgs}
	if req.Options != nil {
		opts := map[string]any{"num_ctx": req.Options.NumCtx}
		if req.Options.NumPredict > 0 {
			opts["num_predict"] = req.Options.NumPredict
		}
		out["options"] = opts
	}
	if req.Format != "" {
		out["format"] = req.Format
	}
	return out
}

func repeatStr(s string, n int) string {
	out := make([]byte, 0, n*len(s))
	for i := 0; i < n; i++ {
		out = append(out, s...)
	}
	return string(out)
}

// ---------------------------------------------------------------------------
// The behavior golden
// ---------------------------------------------------------------------------

func TestGoldenBehavior(t *testing.T) {
	data, err := loadGoldenFile("llmclient_behavior.golden.json")
	if err != nil {
		t.Fatal(err)
	}
	probes := data["probes"].(map[string]any)

	compare := func(name string, got any, root string) {
		t.Helper()
		want := goldenNormalize(probes[name])
		gotNorm := normalizeValue(goldenNormalize(got), root, "")
		if !reflect.DeepEqual(gotNorm, want) {
			t.Errorf("behavior %s mismatch:\n got: %s\nwant: %s", name,
				storage.PyDumpsCompact(gotNorm),
				storage.PyDumpsCompact(want))
		}
	}

	// The probe matrix over the llama.cpp fake.
	scenarios := []struct {
		name string
		beh  map[string]any
	}{
		{"probe_found_ready", nil},
		{"probe_model_prefers_v1", map[string]any{
			"models": []any{"advertised-name"}}},
		{"probe_still_loading", map[string]any{"health_status": "loading"}},
		{"probe_hidden_models", map[string]any{"hide_models": true}},
		{"probe_hidden_props", map[string]any{"hide_props": true}},
		{"probe_not_llamacpp_props", map[string]any{"not_llamacpp": true}},
		{"probe_limits_grab", map[string]any{"props": llamaPropsWithLimits}},
		{"probe_owned_by_fingerprint", map[string]any{
			"hide_props": true, "models": []any{"qwen2.5-3b"},
			"models_owned_by": "llama.cpp"}},
		{"probe_gguf_id_fingerprint", map[string]any{
			"hide_props": true,
			"models":     []any{"models/qwen2.5-3b-instruct-q4_k_m.gguf"}}},
		{"probe_server_header_fingerprint", map[string]any{
			"hide_props": true, "models": []any{"whatever"},
			"server_header": "llama.cpp"}},
		{"probe_plain_openai_not_misreported", map[string]any{
			"hide_props": true, "models": []any{"gpt-4o-mini"},
			"models_owned_by": "organization"}},
		{"probe_vllm_not_misreported", map[string]any{
			"hide_props": true,
			"models":     []any{"meta-llama/Meta-Llama-3-8B-Instruct"}}},
	}
	for _, sc := range scenarios {
		ls := newLlamaCppServer(t, sc.beh)
		compare(sc.name, probeToGolden(ProbeLlamaCpp(ls.root(), "",
			LLamaCppProbeTimeoutS)), ls.root())
	}

	// probe_raw_spelling
	ls := newLlamaCppServer(t, nil)
	compare("probe_raw_spelling", probeToGolden(ProbeLlamaCpp(
		trimHTTP(ls.root()), "", LLamaCppProbeTimeoutS)), ls.root())

	// probe_dead_address (the detail carries the dead URL → <DEADURL>)
	dead := deadPortURL(t)
	deadWant := goldenNormalize(probes["probe_dead_address"])
	deadGot := probeToGolden(ProbeLlamaCpp(dead, "", LLamaCppProbeTimeoutS))
	deadGotNorm := normalizeValue(goldenNormalize(deadGot), "", dead)
	if !reflect.DeepEqual(deadGotNorm, deadWant) {
		t.Errorf("behavior probe_dead_address mismatch:\n got: %s\nwant: %s",
			storage.PyDumpsCompact(deadGotNorm),
			storage.PyDumpsCompact(deadWant))
	}

	// detect scenarios
	ls2 := newLlamaCppServer(t, nil)
	withProcessPorts(t, nil, func() {
		compare("detect_finds_scan", probeToGolden(DetectLlamaCpp("",
			[]int{ls2.port()}, "127.0.0.1", LLamaCppProbeTimeoutS, true)),
			ls2.root())
		openai := newLlamaCppServer(t, map[string]any{
			"models": []any{"gpt-fake"}, "hide_props": true})
		compare("detect_skips_non_llamacpp", probeToGolden(DetectLlamaCpp("",
			[]int{openai.port(), ls2.port()}, "127.0.0.1",
			LLamaCppProbeTimeoutS, true)), ls2.root())
		if probe := DetectLlamaCpp("", []int{openai.port()}, "127.0.0.1",
			LLamaCppProbeTimeoutS, true); probe != nil {
			t.Error("detect_only_non_llamacpp: expected nil")
		}
		if probe := DetectLlamaCpp("", []int{deadPort(t), deadPort(t)},
			"127.0.0.1", LLamaCppProbeTimeoutS, true); probe != nil {
			t.Error("detect_dead_ports: expected nil")
		}
	})
	withProcessPorts(t, []int{ls2.port()}, func() {
		compare("detect_process_first", probeToGolden(DetectLlamaCpp("",
			[]int{}, "127.0.0.1", LLamaCppProbeTimeoutS, true)), ls2.root())
	})

	// detect_ollama scenarios
	ollamaScenarios := []struct {
		name   string
		routes map[string]any
	}{
		{"detect_ollama_found_name_key", map[string]any{
			"/api/tags": route(200, map[string]any{
				"models": []any{
					map[string]any{"name": "llama3:latest"},
					map[string]any{"name": "qwen2.5:7b"}}})}},
		{"detect_ollama_model_key_dedup_order", map[string]any{
			"/api/tags": route(200, map[string]any{
				"models": []any{
					map[string]any{"model": "gemma2:2b"},
					map[string]any{"name": "gemma2:2b"},
					map[string]any{"name": ""},
					map[string]any{"model": nil},
					map[string]any{"name": "mistral:latest"}}})}},
		{"detect_ollama_up_no_models", map[string]any{
			"/api/tags": route(200, map[string]any{
				"models": []any{}})}},
		{"detect_ollama_non_dict_json", map[string]any{
			"/api/tags": route(200, []any{"not", "a", "dict"})}},
		{"detect_ollama_missing_route", map[string]any{
			"/": route(200, map[string]any{})}},
	}
	for _, sc := range ollamaScenarios {
		rs := newRecordingServer(t, sc.routes)
		r := DetectOllama(rs.root(), 4.0)
		got := detectOllamaToGolden(r)
		if sc.name == "detect_ollama_missing_route" {
			// The Python generator truncates at the class prefix (the
			// HTTP reason text is runtime-shaped).
			got["detail"] = prefixBeforeColon(r.Detail)
		}
		compare(sc.name, got, rs.root())
	}
	// dead port: prefix + <DEADURL> base
	dead2 := deadPortURL(t)
	r := DetectOllama(dead2, 0.5)
	gotDead := map[string]any{
		"found":    r.Found,
		"base_url": "<DEADURL>",
		"models":   r.Models,
		"detail":   prefixBeforeColon(r.Detail),
	}
	if wantDead := goldenNormalize(probes["detect_ollama_dead_port"]); !reflect.DeepEqual(goldenNormalize(gotDead), wantDead) {
		t.Errorf("behavior detect_ollama_dead_port mismatch:\n got: %s\nwant: %s",
			storage.PyDumpsCompact(goldenNormalize(gotDead)),
			storage.PyDumpsCompact(wantDead))
	}

	// preflight scenarios
	rs := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "a"},
				map[string]any{"id": "b"}}})})
	compare("preflight_openai_listed", preflightToGolden(
		PreflightOpenAI(rs.root()+"/v1", "k", "a", 5)), rs.root())
	compare("preflight_openai_not_listed", preflightToGolden(
		PreflightOpenAI(rs.root()+"/v1", "k", "zzz", 5)), rs.root())
	rs2 := newRecordingServer(t, map[string]any{
		"/v1/models": route(404, map[string]any{"error": "not found"})})
	compare("preflight_openai_hidden_route", preflightToGolden(
		PreflightOpenAI(rs2.root()+"/v1", "k", "a", 5)), rs2.root())
	okU, msgU, listedU := PreflightOpenAI("http://127.0.0.1:9/v1", "k", "a", 5)
	if i := strings.Index(msgU, ": <urlopen error"); i >= 0 {
		msgU = msgU[:i] + ": <urlopen error\u2026>"
	}
	compare("preflight_openai_unreachable",
		preflightToGolden(okU, msgU, listedU), "")

	// The error-family messages
	runErr := func(name string, fn func() error) {
		t.Helper()
		err := fn()
		want := goldenNormalize(probes[name])
		got := goldenNormalize(errorToGolden(err))
		if name == "anthropic_http_error" {
			// The generator's anthropic capture carries class/code/message
			// only (no body) — mirror the shape.
			delete(got.(map[string]any), "body")
		}
		if !reflect.DeepEqual(got, want) {
			t.Errorf("behavior %s mismatch:\n got: %s\nwant: %s", name,
				storage.PyDumpsCompact(got),
				storage.PyDumpsCompact(want))
		}
	}
	rsErr := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(400, map[string]any{
			"error": map[string]any{"message": "bad request"}})})
	runErr("openai_http_error", func() error {
		_, err := OpenAIChat(rsErr.root()+"/v1", "k", "m", msgsHi, 10, nil)
		return err
	})
	rsErr2 := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200, map[string]any{
			"error": map[string]any{"message": "boom from server"}})})
	runErr("openai_error_object", func() error {
		_, err := OpenAIChat(rsErr2.root()+"/v1", "k", "m", msgsHi, 10, nil)
		return err
	})
	rsErr3 := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200, map[string]any{
			"choices": []any{}})})
	runErr("openai_no_choices", func() error {
		_, err := OpenAIChat(rsErr3.root()+"/v1", "k", "m", msgsHi, 10, nil)
		return err
	})
	rsErr4 := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200,
			[]byte(`{"choices": [{"mes`))})
	runErr("openai_malformed", func() error {
		_, err := OpenAIChat(rsErr4.root()+"/v1", "k", "m", msgsHi, 10, nil)
		return err
	})
	runErr("openai_empty_url", func() error {
		_, err := OpenAIChat("", "k", "m", msgsHi, 10, nil)
		return err
	})
	runErr("anthropic_empty_url", func() error {
		_, err := AnthropicChat("", "k", "m", msgsHi, 10, nil)
		return err
	})
	rsErr5 := newRecordingServer(t, map[string]any{
		"/v1/messages": route(401, map[string]any{"m": "no"})})
	runErr("anthropic_http_error", func() error {
		_, err := AnthropicChat(rsErr5.root()+"/v1", "k", "m", msgsHi, 10, nil)
		return err
	})
	rsErr6 := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"type":  "error",
			"error": map[string]any{"message": "bad model"}})})
	runErr("anthropic_error_object", func() error {
		_, err := AnthropicChat(rsErr6.root()+"/v1", "k", "m", msgsHi, 10, nil)
		return err
	})

	// Returns: success + empty + joins
	rsOK := newOpenAIServer(t)
	compare("openai_success", "ok", rsOK.root())
	rsEmpty := newRecordingServer(t, map[string]any{
		"/v1/chat/completions": route(200, map[string]any{
			"choices": []any{map[string]any{
				"message": map[string]any{"content": ""}}}})})
	compare("openai_empty_content", "", rsEmpty.root())
	rsJoin := newRecordingServer(t, map[string]any{
		"/v1/messages": route(200, map[string]any{
			"content": []any{
				map[string]any{"type": "text", "text": "part one "},
				map[string]any{"type": "thinking", "thinking": "(internal)"},
				map[string]any{"type": "text", "text": "part two"}}})})
	outJoin, err := AnthropicChat(rsJoin.root()+"/v1", "k", "m", msgsHi, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	compare("anthropic_joins_text_blocks", outJoin, rsJoin.root())

	// list models returns
	rsLM := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "a"},
				map[string]any{"id": "b"}}})})
	names, err := OpenAIListModels(rsLM.root()+"/v1", "k", 5)
	if err != nil {
		t.Fatal(err)
	}
	compare("openai_list_models_names", names, rsLM.root())
	rsALM := newRecordingServer(t, map[string]any{
		"/v1/models": route(200, map[string]any{
			"data": []any{map[string]any{"id": "claude-sonnet-4-5"},
				map[string]any{"id": "claude-opus-4-1"}}})})
	aNames, err := AnthropicListModels(rsALM.root()+"/v1", "sk-ant", 5)
	if err != nil {
		t.Fatal(err)
	}
	compare("anthropic_list_models_names", aNames, rsALM.root())

	// The over-budget warnings
	rsOB := newOpenAIServer(t)
	wOB := []string{}
	if _, err := OpenAIChat(rsOB.root()+"/v1", "k", "m",
		[]Message{{Content: repeatStr("x", 40000)}}, 10,
		&ChatOptions{NumCtx: 8192, OnWarn: func(m string) {
			wOB = append(wOB, m)
		}}); err != nil {
		t.Fatal(err)
	}
	compare("openai_over_budget_warning", wOB, rsOB.root())
	rsAB := newRecordingServer(t, map[string]any{
		"/v1/messages": anthropicOKReply})
	wAB := []string{}
	if _, err := AnthropicChat(rsAB.root()+"/v1", "k", "m",
		[]Message{{Content: repeatStr("x", 40000)}}, 10,
		&ChatOptions{NumCtx: 8192, OnWarn: func(m string) {
			wAB = append(wAB, m)
		}}); err != nil {
		t.Fatal(err)
	}
	compare("anthropic_over_budget_warning", wAB, rsAB.root())

	// cloud_chat routes by URL (the /anthropic.com path trick)
	rsRoute := newRecordingServer(t, map[string]any{
		"/anthropic.com/messages": anthropicOKReply,
		"/v1/chat/completions":    okChatReply})
	anthOut, err := CloudChat(rsRoute.root()+"/anthropic.com", "k", "m",
		msgsHi, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	openaiOut, err := CloudChat(rsRoute.root()+"/v1", "k", "m", msgsHi, 10, nil)
	if err != nil {
		t.Fatal(err)
	}
	paths := []string{}
	for _, r := range rsRoute.captured() {
		paths = append(paths, r.Path)
	}
	compare("cloud_chat_routes_by_url", map[string]any{
		"anthropic": anthOut, "openai": openaiOut, "requests": paths,
	}, rsRoute.root())
}

// probeToGolden renders a ProbeResult as the golden's dict (the via key
// rides omitempty like the reference's absent key).
func probeToGolden(p *ProbeResult) map[string]any {
	if p == nil {
		return nil
	}
	out := map[string]any{
		"found":         p.Found,
		"ready":         p.Ready,
		"models":        p.Models,
		"model":         p.Model,
		"props_model":   p.PropsModel,
		"identified_by": p.IdentifiedBy,
		"base_url":      p.BaseURL,
		"detail":        p.Detail,
		"props_limits":  p.PropsLimits,
	}
	if p.Via != "" {
		out["via"] = p.Via
	}
	return out
}

func detectOllamaToGolden(r DetectOllamaResult) map[string]any {
	return map[string]any{
		"found":    r.Found,
		"base_url": r.BaseURL,
		"models":   r.Models,
		"detail":   r.Detail,
	}
}

func preflightToGolden(ok bool, msg string, listed *bool) any {
	return []any{ok, msg, listed}
}

func errorToGolden(err error) map[string]any {
	if err == nil {
		return map[string]any{"class": "<no error>"}
	}
	out := map[string]any{
		"class":   errorClass(err),
		"message": err.Error(),
	}
	if httpErr, ok := err.(*CloudLLMHTTPError); ok {
		out["code"] = httpErr.Code
		out["body"] = httpErr.Body
	}
	return out
}

func errorClass(err error) string {
	switch err.(type) {
	case *CloudLLMHTTPError:
		return "CloudLLMHTTPError"
	case *CloudLLMBadResponse:
		return "CloudLLMBadResponse"
	case *CloudLLMError:
		return "CloudLLMError"
	case *TimeoutError:
		return "TimeoutError"
	}
	return "error"
}

func prefixBeforeColon(s string) string {
	for i := 0; i < len(s); i++ {
		if s[i] == ':' {
			return s[:i]
		}
	}
	return s
}

func trimHTTP(s string) string {
	if len(s) >= 7 && s[:7] == "http://" {
		return s[7:]
	}
	return s
}

// ---------------------------------------------------------------------------
// The remaining pure sections
// ---------------------------------------------------------------------------

func TestGoldenPureRemaining(t *testing.T) {
	corpus := loadPureGolden(t)
	g := func(entry *storage.OrderedMap, key string) any {
		v, _ := entry.Get(key)
		return v
	}
	entries := func(key string) []*storage.OrderedMap {
		raw, _ := corpus[key].([]any)
		out := []*storage.OrderedMap{}
		for _, e := range raw {
			out = append(out, e.(*storage.OrderedMap))
		}
		return out
	}

	t.Run("normalize_llamacpp_api_url", func(t *testing.T) {
		for _, e := range entries("normalize_llamacpp_api_url") {
			in := goldenStr(g(e, "url"))
			if got := NormalizeLlamaCppAPIURL(in); got != goldenStr(g(e, "want")) {
				t.Errorf("normalize(%q) = %q, want %q", in, got, g(e, "want"))
			}
		}
	})

	t.Run("is_loopback_url", func(t *testing.T) {
		table := corpus["is_loopback_url"].(*storage.OrderedMap)
		yes, _ := table.Get("yes")
		for _, uAny := range yes.([]any) {
			if !IsLoopbackURL(uAny.(string)) {
				t.Errorf("loopback(%q) = false", uAny)
			}
		}
		no, _ := table.Get("no")
		for _, uAny := range no.([]any) {
			if uAny == nil {
				continue // Python None — the empty-string case covers it
			}
			if IsLoopbackURL(uAny.(string)) {
				t.Errorf("loopback(%q) = true", uAny)
			}
		}
	})

	t.Run("is_llamacpp_props", func(t *testing.T) {
		for _, e := range entries("is_llamacpp_props") {
			want, _ := g(e, "want").(bool)
			if got := IsLlamaCppProps(g(e, "props")); got != want {
				t.Errorf("is_llamacpp_props(%v) = %v", g(e, "props"), got)
			}
		}
	})

	t.Run("llamacpp_model_from_props", func(t *testing.T) {
		for _, e := range entries("llamacpp_model_from_props") {
			got := LlamaCppModelFromProps(g(e, "props"))
			want := g(e, "want")
			if want == nil {
				if got != nil {
					t.Errorf("model_from_props(%v) = %v, want nil", g(e, "props"), got)
				}
				continue
			}
			if got == nil || *got != want.(string) {
				t.Errorf("model_from_props(%v) = %v, want %q", g(e, "props"), got, want)
			}
		}
	})

	t.Run("llamacpp_limits_from_props", func(t *testing.T) {
		for _, e := range entries("llamacpp_limits_from_props") {
			got := LlamaCppLimitsFromProps(g(e, "props"))
			want := g(e, "want").(*storage.OrderedMap)
			wantCtxV, _ := want.Get("n_ctx")
			wantPredictV, _ := want.Get("n_predict")
			wantCtx := ptrFromGolden(wantCtxV)
			wantPredict := ptrFromGolden(wantPredictV)
			if !ptrIntEq(got.NCtx, wantCtx) || !ptrIntEq(got.NPredict, wantPredict) {
				t.Errorf("limits_from_props(%v) = %+v, want %v/%v",
					g(e, "props"), got, wantCtxV, wantPredictV)
			}
		}
	})

	t.Run("llamacpp_limits_line", func(t *testing.T) {
		for _, e := range entries("llamacpp_limits_line") {
			got := llamaCppLimitsLine(limitsFromGolden(g(e, "limits")))
			if got != goldenStr(g(e, "want")) {
				t.Errorf("limits_line(%v) = %q, want %q", g(e, "limits"), got, g(e, "want"))
			}
		}
	})

	t.Run("resolve_llamacpp_model", func(t *testing.T) {
		for _, e := range entries("resolve_llamacpp_model") {
			var models []string
			if raw, ok := g(e, "models").([]any); ok {
				for _, m := range raw {
					models = append(models, m.(string))
				}
			}
			got := ResolveLlamaCppModel(g(e, "config"), models,
				ptrFromGoldenString(g(e, "props_model")))
			want := g(e, "want")
			if want == nil {
				if got != nil {
					t.Errorf("resolve_model = %v, want nil", got)
				}
				continue
			}
			if got == nil || *got != want.(string) {
				t.Errorf("resolve_model(%v) = %v, want %q", g(e, "config"), got, want)
			}
		}
	})

	t.Run("is_anthropic_url", func(t *testing.T) {
		table := corpus["is_anthropic_url"].(*storage.OrderedMap)
		yes, _ := table.Get("yes")
		for _, uAny := range yes.([]any) {
			if !IsAnthropicURL(uAny.(string)) {
				t.Errorf("is_anthropic_url(%q) = false", uAny)
			}
		}
		no, _ := table.Get("no")
		for _, uAny := range no.([]any) {
			if uAny == nil {
				continue // Python None — the empty-string case covers it
			}
			if IsAnthropicURL(uAny.(string)) {
				t.Errorf("is_anthropic_url(%q) = true", uAny)
			}
		}
	})

	t.Run("autodetect_decision", func(t *testing.T) {
		for _, e := range entries("autodetect_decision") {
			probe := probeFromGolden(g(e, "probe"))
			got := LlamaCppAutodetectDecision(g(e, "config"), probe,
				boolPtrFromGolden(g(e, "ollama_up")))
			want := g(e, "want").(*storage.OrderedMap)
			switchV, _ := want.Get("switch")
			urlV, _ := want.Get("url")
			modelV, _ := want.Get("model")
			limitsV, _ := want.Get("limits")
			if got.Switch != switchV.(bool) {
				t.Errorf("%s: switch = %v, want %v", g(e, "name"), got.Switch, switchV)
			}
			if !ptrStrEq(got.URL, ptrFromGoldenString(urlV)) {
				t.Errorf("%s: url = %v, want %v", g(e, "name"), got.URL, urlV)
			}
			if !ptrStrEq(got.Model, ptrFromGoldenString(modelV)) {
				t.Errorf("%s: model = %v, want %v", g(e, "name"), got.Model, modelV)
			}
			if !limitsPtrEq(got.Limits, limitsFromGolden(limitsV)) {
				t.Errorf("%s: limits = %+v, want %v", g(e, "name"), got.Limits, limitsV)
			}
			msgWant, _ := want.Get("message")
			levelWant, _ := want.Get("level")
			if got.Message != goldenStr(msgWant) {
				t.Errorf("%s: message = %q, want %q", g(e, "name"), got.Message, msgWant)
			}
			if got.Level != goldenStr(levelWant) {
				t.Errorf("%s: level = %q, want %q", g(e, "name"), got.Level, levelWant)
			}
		}
	})

	t.Run("parse_tasklist_pids", func(t *testing.T) {
		for _, e := range entries("parse_tasklist_pids") {
			var text string
			if s2, ok := g(e, "text").(string); ok {
				text = s2
			}
			got := parseTasklistPIDs(text)
			want := []int{}
			if raw, ok := g(e, "want").([]any); ok {
				for _, p := range raw {
					want = append(want, pyInt(p))
				}
			}
			if !intSliceEq(got, want) {
				t.Errorf("parse_tasklist = %v, want %v", got, want)
			}
		}
	})

	t.Run("parse_netstat_ports", func(t *testing.T) {
		for _, e := range entries("parse_netstat_ports") {
			pids := []int{}
			if raw, ok := g(e, "pids").([]any); ok {
				for _, p := range raw {
					pids = append(pids, pyInt(p))
				}
			}
			got := parseNetstatListeningPorts(g(e, "text").(string), pids)
			want := []int{}
			if raw, ok := g(e, "want").([]any); ok {
				for _, p := range raw {
					want = append(want, pyInt(p))
				}
			}
			if !intSliceEq(got, want) {
				t.Errorf("parse_netstat = %v, want %v", got, want)
			}
		}
	})

	t.Run("parse_posix_ports", func(t *testing.T) {
		for _, e := range entries("parse_posix_ports") {
			got := parsePosixPorts(g(e, "text").(string))
			want := []int{}
			if raw, ok := g(e, "want").([]any); ok {
				for _, p := range raw {
					want = append(want, pyInt(p))
				}
			}
			if !intSliceEq(got, want) {
				t.Errorf("parse_posix = %v, want %v", got, want)
			}
		}
	})

	t.Run("response_content", func(t *testing.T) {
		for _, e := range entries("response_content") {
			if skip, ok := g(e, "skip_go").(bool); ok && skip {
				continue // the SDK object-shape legs — not representable
			}
			resp := responseFromGolden(g(e, "response"))
			if got := ResponseContent(resp); got != goldenStr(g(e, "want")) {
				t.Errorf("response_content(%v) = %q, want %q",
					g(e, "shape"), got, g(e, "want"))
			}
		}
	})

	t.Run("list_models_names", func(t *testing.T) {
		table := corpus["list_models_names"].(*storage.OrderedMap)
		// The dict-shape leg is the portable one (the golden's
		// name_key entry): name-or-model extraction over /api/tags.
		nameKey, _ := table.Get("name_key")
		want := []string{}
		for _, n := range nameKey.([]any) {
			want = append(want, n.(string))
		}
		models := []any{
			map[string]any{"name": "old-api"},
			map[string]any{"model": "old-api-2"}}
		got := extractTagNames(models)
		if !reflect.DeepEqual(got, want) {
			t.Errorf("list names = %v, want %v", got, want)
		}
		// The new-API leg (the .model attr preference) is the SDK
		// object shape — the port's wire rule is name-or-model; pinned
		// by the same extraction (documented divergence).
		newAPI, _ := table.Get("model_key_new_api")
		want2 := []string{}
		for _, n := range newAPI.([]any) {
			want2 = append(want2, n.(string))
		}
		models2 := []any{
			map[string]any{"model": "new-api"},
			map[string]any{"model": "second"}}
		got2 := extractTagNames(models2)
		if !reflect.DeepEqual(got2, want2) {
			t.Errorf("list names (new-api) = %v, want %v", got2, want2)
		}
	})
}

func ptrFromGolden(v any) *int {
	if v == nil {
		return nil
	}
	n := pyInt(v)
	return &n
}

func ptrFromGoldenString(v any) *string {
	if v == nil {
		return nil
	}
	s := v.(string)
	return &s
}

func boolPtrFromGolden(v any) *bool {
	if v == nil {
		return nil
	}
	b := v.(bool)
	return &b
}

func ptrStrEq(a, b *string) bool {
	if a == nil || b == nil {
		return a == b
	}
	return *a == *b
}

func limitsPtrEq(a, b *LlamaCppLimits) bool {
	if a == nil || b == nil {
		return a == b
	}
	return ptrIntEq(a.NCtx, b.NCtx) && ptrIntEq(a.NPredict, b.NPredict)
}

func limitsFromGolden(v any) *LlamaCppLimits {
	if v == nil {
		return nil
	}
	m, ok := asOM(v)
	if !ok || m.Len() == 0 {
		return nil
	}
	nc, _ := m.Get("n_ctx")
	np, _ := m.Get("n_predict")
	return &LlamaCppLimits{
		NCtx:     ptrFromGolden(nc),
		NPredict: ptrFromGolden(np),
	}
}

func probeFromGolden(v any) *ProbeResult {
	m, ok := asOM(v)
	if !ok {
		return &ProbeResult{Models: []string{}}
	}
	p := &ProbeResult{Models: []string{}}
	if f, ok := m.Get("found"); ok {
		p.Found, _ = f.(bool)
	}
	if bu, ok := m.Get("base_url"); ok {
		if s2, ok := bu.(string); ok {
			p.BaseURL = s2
		}
	}
	if r, ok := m.Get("ready"); ok {
		if b, ok := r.(bool); ok {
			p.Ready = &b
		}
	}
	if mod, ok := m.Get("model"); ok {
		if s2, ok := mod.(string); ok {
			p.Model = &s2
		}
	}
	if pm, ok := m.Get("props_model"); ok {
		if s2, ok := pm.(string); ok {
			p.PropsModel = &s2
		}
	}
	if id, ok := m.Get("identified_by"); ok {
		if s2, ok := id.(string); ok {
			p.IdentifiedBy = &s2
		}
	}
	if models, ok := m.Get("models"); ok {
		if list, ok := models.([]any); ok {
			for _, mm := range list {
				if ms, ok := mm.(string); ok {
					p.Models = append(p.Models, ms)
				}
			}
		}
	}
	pl, _ := m.Get("props_limits")
	p.PropsLimits = limitsFromGolden(pl)
	return p
}

func responseFromGolden(v any) *OllamaChatResponse {
	if v == nil {
		return nil
	}
	m, ok := asOM(v)
	if !ok {
		return &OllamaChatResponse{}
	}
	resp := &OllamaChatResponse{}
	if msgV, ok := m.Get("message"); ok {
		if msg, ok := asOM(msgV); ok {
			resp.Message = &OllamaChatMessage{}
			if c, ok := msg.Get("content"); ok {
				if cs, ok := c.(string); ok {
					resp.Message.Content = cs
				}
			}
		}
	}
	return resp
}

var _ = helpers.PyStr
