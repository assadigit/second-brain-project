// anthropic.go — the Anthropic Claude Messages wire (llm_client.py
// anthropic_chat / anthropic_list_models / cloud_chat /
// preflight_cloud): x-api-key + anthropic-version headers, the
// top-level system parameter, REQUIRED max_tokens, content-block
// joining — and the ONE cloud router (the URL decides the wire).
package llmclient

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// AnthropicChat POSTs {base}/messages — the Claude wire format. Same
// contract as OpenAIChat (wall-clock timeout, CloudLLM* errors, ” on
// a well-formed empty answer) with the Messages-API differences: the
// system message rides the top-level `system` parameter (user/assistant
// turns only stay in messages), max_tokens is REQUIRED (the user's cap
// when set, else the fallback), and the answer joins the text content
// blocks. response_format does not exist on this API: JSON mode only
// appends the instruction to EXISTING system text (the reference's
// and-system_text guard — no system message → no instruction).
func AnthropicChat(apiURL, apiKey, model string, messages []Message,
	timeoutS float64, opts *ChatOptions) (string, error) {
	base := strings.TrimRight(apiURL, "/")
	if base == "" {
		return "", newCloudLLMError(
			"no API URL configured for the Claude endpoint")
	}

	if opts != nil && opts.NumCtx > 0 && opts.OnWarn != nil {
		if est := EstimateTokens(messages); est > opts.NumCtx {
			opts.OnWarn(fmt.Sprintf(
				"prompt ≈%d tokens may exceed the model's context window "+
					"(llm_num_ctx=%d) — answers may degrade if the model "+
					"can't see the whole prompt.", est, opts.NumCtx))
		}
	}

	// Split the system messages out (Messages API: top-level system).
	systemParts := []string{}
	chatMessages := make([]Message, 0, len(messages))
	for _, m := range messages {
		if m.Role == "system" {
			if m.Content != "" {
				systemParts = append(systemParts, m.Content)
			}
			continue
		}
		role := m.Role
		if role != "user" && role != "assistant" {
			role = "user"
		}
		chatMessages = append(chatMessages, Message{
			Role: role, Content: m.Content})
	}
	systemText := strings.Join(systemParts, "\n\n")
	if opts != nil && opts.JSONMode && systemText != "" {
		systemText = systemText +
			"\n\nReply with ONLY a valid JSON object — no prose, " +
			"no markdown fences."
	}

	body := storage.NewOrderedMap()
	body.Set("model", model)
	body.Set("messages", messagesToAny(chatMessages))
	maxTokens := AnthropicFallbackMaxTokens
	if opts != nil && opts.MaxOutputTokens > 0 {
		maxTokens = opts.MaxOutputTokens
	}
	body.Set("max_tokens", maxTokens)
	body.Set("temperature", opts.temperature())
	if systemText != "" {
		body.Set("system", systemText)
	}
	data := []byte(storage.PyDumpsCompact(body))

	req, err := http.NewRequest("POST", base+"/messages",
		bytes.NewReader(data))
	if err != nil {
		return "", newCloudLLMError(
			"could not reach the Claude endpoint %s: %s",
			base, pyURLOpenError(err))
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Anthropic-Version", AnthropicVersionHeader)
	if apiKey != "" {
		req.Header.Set("X-Api-Key", apiKey)
	}

	// The reference's quirk: the Anthropic path always rides the
	// DEFAULT urlopen behavior (env proxy) — no loopback branch here.
	res, err := doWireRequest(req, timeoutS, opts != nil && opts.VerifyTLS, false)
	var httpErr *httpStatusError
	if errors.As(err, &httpErr) {
		return "", NewCloudLLMHTTPError(httpErr.code, httpErr.body)
	}
	if err != nil {
		if isDeadline(err) {
			return "", &TimeoutError{Seconds: timeoutS}
		}
		return "", newCloudLLMError(
			"could not reach the Claude endpoint %s: %s",
			base, pyURLOpenError(err))
	}
	payload, perr := parseChatBody(res.body)
	if perr != nil {
		return "", perr
	}
	return anthropicContent(payload)
}

// anthropicContent joins the text blocks with the reference quirks: an
// embedded truthy error object raises, missing/empty content blocks
// raise, non-text blocks are skipped, empty texts drop out of the join.
func anthropicContent(payload any) (string, error) {
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
	blocksV, _ := om.Get("content")
	blocks, isList := blocksV.([]any)
	if !isList || len(blocks) == 0 {
		return "", newBadResponse("endpoint returned no content blocks")
	}
	texts := []string{}
	for _, b := range blocks {
		bom, ok := asOM(b)
		if !ok {
			continue
		}
		if tv, p := bom.Get("type"); !p || tv != "text" {
			continue
		}
		text := ""
		if xv, p := bom.Get("text"); p {
			text = helpers.PyStr(pyOr(xv, ""))
		}
		if text != "" {
			texts = append(texts, text)
		}
	}
	return strings.Join(texts, "\n"), nil
}

// AnthropicListModels GETs {base}/models with the Anthropic headers —
// same contract as OpenAIListModels.
func AnthropicListModels(apiURL, apiKey string,
	timeoutS float64) ([]string, error) {
	base := strings.TrimRight(apiURL, "/")
	if base == "" {
		return nil, newCloudLLMError(
			"no API URL configured for the Claude endpoint")
	}
	req, err := http.NewRequest("GET", base+"/models", nil)
	if err != nil {
		return nil, newCloudLLMError(
			"could not reach %s/models: %s", base, pyURLOpenError(err))
	}
	req.Header.Set("X-Api-Key", apiKey)
	req.Header.Set("Anthropic-Version", AnthropicVersionHeader)
	// No loopback branch (the reference's plain urlopen).
	res, err := doListRequest(req, timeoutS, false)
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
	names := []string{}
	if om, ok := asOM(payload); ok {
		if dataV, p := om.Get("data"); p {
			if list, ok := dataV.([]any); ok {
				for _, m := range list {
					mom, ok := asOM(m)
					if !ok {
						continue
					}
					var mid any
					if idV, p := mom.Get("id"); p && pyTruthy(idV) {
						mid = idV
					} else if nV, p := mom.Get("name"); p && pyTruthy(nV) {
						mid = nV
					}
					if mid != nil {
						names = append(names, helpers.PyStr(mid))
					}
				}
			}
		}
	}
	return names, nil
}

// CloudChat is the ONE cloud router: the URL decides the wire format.
// api.anthropic.com → AnthropicChat (Claude), everything else →
// OpenAIChat (llama.cpp, vLLM, LM Studio, OpenAI, OpenRouter, Together…).
func CloudChat(apiURL, apiKey, model string, messages []Message,
	timeoutS float64, opts *ChatOptions) (string, error) {
	if IsAnthropicURL(apiURL) {
		return AnthropicChat(apiURL, apiKey, model, messages, timeoutS, opts)
	}
	return OpenAIChat(apiURL, apiKey, model, messages, timeoutS, opts)
}

// PreflightCloud is the model-list pre-flight for EITHER cloud wire
// format (the URL decides, exactly like CloudChat). Same contract as
// PreflightOpenAI — never a gate.
func PreflightCloud(apiURL, apiKey, model string,
	timeoutS float64) (bool, string, *bool) {
	var (
		names []string
		err   error
	)
	if IsAnthropicURL(apiURL) {
		names, err = AnthropicListModels(apiURL, apiKey, timeoutS)
	} else {
		names, err = OpenAIListModels(apiURL, apiKey, timeoutS)
	}
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

var (
	_ = context.Background
	_ = io.EOF
)
