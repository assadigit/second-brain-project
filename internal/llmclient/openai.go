// openai.go — the OpenAI-compatible wire (llm_client.py openai_chat /
// openai_list_models / preflight_openai): POST {base}/chat/completions,
// JSON mode with the memoized one-shot 400 fallback, the wall-clock
// timeout as a context deadline, and the exact error strings.
package llmclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ChatOptions carries the keyword half of the chat calls
// (json_mode, temperature, verify_tls, num_ctx, max_output_tokens,
// on_warn). The zero value means the Python defaults: JSON mode off,
// temperature 0.7 (a literal 0.0 is not representable — no caller
// sends one; documented in doc.go), TLS verification OFF (the
// reference's default for self-signed local endpoints).
type ChatOptions struct {
	JSONMode        bool
	Temperature     float64
	VerifyTLS       bool
	NumCtx          int
	MaxOutputTokens int
	OnWarn          func(string)
}

func (o *ChatOptions) temperature() float64 {
	if o == nil || o.Temperature == 0 {
		return 0.7
	}
	return o.Temperature
}

// OpenAIChat POSTs {base}/chat/completions. JSON mode sends
// response_format unless the memo remembers a rejection; a 400 to the
// parameter gets ONE clean retry without it and the rejection is
// memoized. Returns the assistant content (” on a well-formed but
// empty answer). Errors: *TimeoutError (wall clock), *CloudLLMHTTPError
// (HTTP status), *CloudLLMBadResponse (malformed body), *CloudLLMError
// (connection failure / no URL).
func OpenAIChat(apiURL, apiKey, model string, messages []Message,
	timeoutS float64, opts *ChatOptions) (string, error) {
	base := strings.TrimRight(apiURL, "/")
	if base == "" {
		return "", newCloudLLMError(
			"no API URL configured for the OpenAI-compatible endpoint")
	}

	// The over-budget warning (the "never truncate silently" half).
	if opts != nil && opts.NumCtx > 0 && opts.OnWarn != nil {
		if est := EstimateTokens(messages); est > opts.NumCtx {
			opts.OnWarn(fmt.Sprintf(
				"prompt ≈%d tokens may exceed the server's context window "+
					"(llm_num_ctx=%d) — launch llama.cpp with -c (or vLLM with "+
					"--max-model-len) at least that large if answers degrade.",
				est, opts.NumCtx))
		}
	}

	sendJSON := opts != nil && opts.JSONMode && !JSONModeRejected(base)
	res, err := openaiPost(base, apiKey, model, messages, timeoutS, opts, sendJSON)
	var httpErr *httpStatusError
	if errors.As(err, &httpErr) {
		if sendJSON && httpErr.code == 400 {
			// The server may be rejecting response_format itself — fall
			// back cleanly: ONE retry without the parameter.
			if opts != nil && opts.OnWarn != nil {
				opts.OnWarn("endpoint rejected response_format — " +
					"retrying without JSON mode (memoized)")
			}
			res2, err2 := openaiPost(base, apiKey, model, messages,
				timeoutS, opts, false)
			var httpErr2 *httpStatusError
			if errors.As(err2, &httpErr2) {
				return "", NewCloudLLMHTTPError(httpErr2.code, httpErr2.body)
			}
			if err2 != nil {
				if isDeadline(err2) {
					return "", &TimeoutError{Seconds: timeoutS}
				}
				return "", newCloudLLMError(
					"could not reach the OpenAI-compatible endpoint %s: %s",
					base, pyURLOpenError(err2))
			}
			rememberJSONModeRejected(base)
			payload, perr := parseChatBody(res2.body)
			if perr != nil {
				return "", perr
			}
			return openaiContent(payload)
		}
		return "", NewCloudLLMHTTPError(httpErr.code, httpErr.body)
	}
	if err != nil {
		if isDeadline(err) {
			return "", &TimeoutError{Seconds: timeoutS}
		}
		return "", newCloudLLMError(
			"could not reach the OpenAI-compatible endpoint %s: %s",
			base, pyURLOpenError(err))
	}
	payload, perr := parseChatBody(res.body)
	if perr != nil {
		return "", perr
	}
	return openaiContent(payload)
}

// wireReply is the 2xx answer: the decoded body text.
type wireReply struct{ body string }

// openaiPost performs one POST with the exact wire shape (the body is
// an ordered dict serialized in Python json.dumps compact form — the
// payload golden pins the bytes).
func openaiPost(base, apiKey, model string, messages []Message,
	timeoutS float64, opts *ChatOptions, useJSONMode bool) (*wireReply, error) {
	body := storage.NewOrderedMap()
	body.Set("model", model)
	body.Set("messages", messagesToAny(messages))
	body.Set("temperature", opts.temperature())
	if opts != nil && opts.MaxOutputTokens > 0 {
		body.Set("max_tokens", opts.MaxOutputTokens)
	}
	if useJSONMode {
		rf := storage.NewOrderedMap()
		rf.Set("type", "json_object")
		body.Set("response_format", rf)
	}
	data := []byte(storage.PyDumpsCompact(body))

	req, err := http.NewRequest("POST", base+"/chat/completions",
		bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	return doWireRequest(req, timeoutS, opts != nil && opts.VerifyTLS,
		IsLoopbackURL(base))
}

// doWireRequest executes req under the timeout law: timeoutS > 0 → a
// context deadline (the deterministic authority — Python's sock timeout
// sat 5s ABOVE it and is subsumed); timeoutS <= 0 → the 120s socket
// cap the reference falls back to. Non-2xx → *httpStatusError with the
// body readable.
func doWireRequest(req *http.Request, timeoutS float64,
	verifyTLS, direct bool) (*wireReply, error) {
	client := httpClient(verifyTLS, direct)
	if timeoutS > 0 {
		ctx, cancel := context.WithTimeout(
			req.Context(), secondsDur(timeoutS))
		defer cancel()
		req = req.WithContext(ctx)
	} else {
		client.Timeout = 120 * time.Second
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &httpStatusError{code: resp.StatusCode, body: string(raw)}
	}
	return &wireReply{body: string(raw)}, nil
}

// parseChatBody decodes the answer body with the exact bad-response
// message (the 160-rune preview, '(empty)' when blank).
func parseChatBody(raw string) (any, error) {
	v, err := storage.DecodeJSONValue([]byte(raw))
	if err != nil {
		preview := helpers.CutRunes(raw, 160)
		if preview == "" {
			preview = "(empty)"
		}
		return nil, newBadResponse(
			"endpoint returned a non-JSON body: %s", preview)
	}
	return v, nil
}

// openaiContent extracts choices[0].message.content with the reference
// quirks: an embedded truthy error object raises, no choices raises,
// a non-string content yields ”.
func openaiContent(payload any) (string, error) {
	om, ok := asOM(payload)
	if !ok {
		return "", newBadResponse("endpoint returned a non-object body")
	}
	if errV, present := om.Get("error"); present && pyTruthy(errV) {
		msg := ""
		if m, isMap := asOM(errV); isMap {
			if mv, p := m.Get("message"); p {
				msg = helpers.PyStr(mv)
			}
		} else {
			msg = helpers.PyStr(errV)
		}
		return "", newBadResponse("endpoint error object: %s", msg)
	}
	choicesV, _ := om.Get("choices")
	choices, isList := choicesV.([]any)
	if !isList || len(choices) == 0 {
		return "", newBadResponse("endpoint returned no choices")
	}
	// choices[0] or {} — a non-dict first choice degrades to empty.
	if first, ok := asOM(choices[0]); ok {
		if msgV, p := first.Get("message"); p {
			if msg, ok := asOM(msgV); ok {
				if cv, p := msg.Get("content"); p {
					if cs, ok := cv.(string); ok {
						return cs, nil
					}
				}
			}
		}
	}
	return "", nil
}

// OpenAIListModels GETs {base}/models and returns the id/name list.
// Raises *CloudLLMError on anything but HTTP 200 (404 counts: llama.cpp
// with no models route, misconfigured proxies…).
func OpenAIListModels(apiURL, apiKey string, timeoutS float64) ([]string, error) {
	base := strings.TrimRight(apiURL, "/")
	if base == "" {
		return nil, newCloudLLMError(
			"no API URL configured for the OpenAI-compatible endpoint")
	}
	req, err := http.NewRequest("GET", base+"/models", nil)
	if err != nil {
		return nil, newCloudLLMError(
			"could not reach %s/models: %s", base, pyURLOpenError(err))
	}
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	res, err := doListRequest(req, timeoutS, IsLoopbackURL(base))
	var httpErr *httpStatusError
	if errors.As(err, &httpErr) {
		return nil, newCloudLLMError(
			"GET %s/models answered HTTP %d", base, httpErr.code)
	}
	if err != nil {
		return nil, newCloudLLMError(
			"could not reach %s/models: %s", base, pyURLOpenError(err))
	}
	payload, perr := storage.DecodeJSONValue([]byte(res.body))
	if perr != nil {
		return nil, newCloudLLMError(
			"GET %s/models returned a non-JSON body", base)
	}
	return modelIDsFromModelsPayload(payload), nil
}

// doListRequest is doWireRequest without the wall-clock wrapper (the
// reference gives list calls a plain socket timeout, so no
// *TimeoutError — a timeout surfaces as a URLError-class failure).
func doListRequest(req *http.Request, timeoutS float64,
	direct bool) (*wireReply, error) {
	client := httpClient(false, direct)
	if timeoutS > 0 {
		ctx, cancel := context.WithTimeout(
			req.Context(), secondsDur(timeoutS))
		defer cancel()
		req = req.WithContext(ctx)
	}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &httpStatusError{code: resp.StatusCode, body: string(raw)}
	}
	return &wireReply{body: string(raw)}, nil
}

// PreflightOpenAI is the /v1/models pre-flight check. NEVER a gate:
// (endpoint_ok, message, model_listed) where model_listed is nil when
// the list could not be read — a server that hides /models is still
// perfectly usable.
func PreflightOpenAI(apiURL, apiKey, model string,
	timeoutS float64) (bool, string, *bool) {
	names, err := OpenAIListModels(apiURL, apiKey, timeoutS)
	if err != nil {
		return false, err.Error(), nil
	}
	var listed *bool
	if len(names) > 0 {
		lowered := map[string]bool{}
		for _, n := range names {
			lowered[strings.ToLower(n)] = true
		}
		v := lowered[strings.ToLower(strings.TrimSpace(
			helpers.PyStr(pyOr(model, ""))))]
		listed = &v
	}
	return true, fmt.Sprintf("%d model(s) listed", len(names)), listed
}

// modelIDsFromModelsPayload extracts the id/name list from a
// /v1/models payload (dicts → id or name — the truthy check; non-dict
// truthy entries → their str()).
func modelIDsFromModelsPayload(payload any) []string {
	names := []string{}
	om, ok := asOM(payload)
	if !ok {
		return names
	}
	dataV, _ := om.Get("data")
	list, ok := dataV.([]any)
	if !ok {
		return names
	}
	for _, m := range list {
		if m == nil {
			continue
		}
		if mom, ok := m.(*storage.OrderedMap); ok {
			var name any
			if idV, p := mom.Get("id"); p && pyTruthy(idV) {
				name = idV
			} else if nV, p := mom.Get("name"); p && pyTruthy(nV) {
				name = nV
			}
			if name != nil {
				names = append(names, helpers.PyStr(name))
			}
		} else if pyTruthy(m) {
			names = append(names, helpers.PyStr(m))
		}
	}
	return names
}

var _ = strconv.Itoa
