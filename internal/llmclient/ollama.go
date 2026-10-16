// ollama.go — the Ollama half (llm_client.py ollama_chat /
// chat_with_timeout / list_models_with_timeout / response_content /
// ollama_reachable). The Python side injects the ollama-py SDK client
// object; the port implements the /api/chat + /api/tags wire directly
// (the SDK is a Python package — no new Go dependency without an ADR,
// and the SDK's own tests pin the same wire: options/format/model).
package llmclient

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// The client seam (Python: the injected ollama client object)
// ---------------------------------------------------------------------------

// OllamaAPI is the injected client seam — callers hand in the HTTP
// client (or a fake in tests).
type OllamaAPI interface {
	Chat(ctx context.Context, req *OllamaChatRequest) (*OllamaChatResponse, error)
	List(ctx context.Context) ([]string, error)
}

// OllamaOptions is the options dict ({num_ctx, num_predict}).
type OllamaOptions struct {
	NumCtx     int
	NumPredict int
}

// OllamaChatRequest is the /api/chat request body: options and format
// are ABSENT when nil/empty (the kwargs law).
type OllamaChatRequest struct {
	Model    string
	Messages []Message
	Options  *OllamaOptions
	Format   string
}

// OllamaChatMessage is the wire message ({role, content}).
type OllamaChatMessage struct {
	Role    string
	Content string
}

// OllamaChatResponse is the /api/chat answer (both ollama-py response
// shapes are ONE wire shape — the documented divergence).
type OllamaChatResponse struct {
	Model   string
	Message *OllamaChatMessage
}

// OllamaClient speaks the Ollama HTTP API directly.
type OllamaClient struct {
	BaseURL string
}

// NewOllamaClient builds a client for the given host base URL.
func NewOllamaClient(host string) *OllamaClient {
	return &OllamaClient{BaseURL: strings.TrimRight(host, "/")}
}

// Chat POSTs {base}/api/chat and parses the answer.
func (c *OllamaClient) Chat(ctx context.Context,
	req *OllamaChatRequest) (*OllamaChatResponse, error) {
	body := storage.NewOrderedMap()
	body.Set("model", req.Model)
	body.Set("messages", messagesToAny(req.Messages))
	if req.Options != nil {
		opts := storage.NewOrderedMap()
		opts.Set("num_ctx", req.Options.NumCtx)
		if req.Options.NumPredict > 0 {
			opts.Set("num_predict", req.Options.NumPredict)
		}
		body.Set("options", opts)
	}
	if req.Format != "" {
		body.Set("format", req.Format)
	}
	data := []byte(storage.PyDumpsCompact(body))
	httpReq, err := http.NewRequestWithContext(ctx, "POST",
		c.BaseURL+"/api/chat", strings.NewReader(string(data)))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")
	resp, err := httpClient(false, false).Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &httpStatusError{code: resp.StatusCode, body: string(raw)}
	}
	var payload struct {
		Model   string `json:"model"`
		Message *struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, err
	}
	out := &OllamaChatResponse{Model: payload.Model}
	if payload.Message != nil {
		out.Message = &OllamaChatMessage{
			Role: payload.Message.Role, Content: payload.Message.Content}
	}
	return out, nil
}

// List GETs {base}/api/tags and extracts the model names (the
// name-or-model wire rule — the same rule the app's own detect_ollama
// speaks).
func (c *OllamaClient) List(ctx context.Context) ([]string, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET",
		c.BaseURL+"/api/tags", nil)
	if err != nil {
		return nil, err
	}
	resp, err := httpClient(false, false).Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, &httpStatusError{code: resp.StatusCode, body: string(raw)}
	}
	payload, err := storage.DecodeJSONValue(raw)
	if err != nil {
		return nil, err
	}
	names := []string{}
	if om, ok := payload.(*storage.OrderedMap); ok {
		if modelsV, p := om.Get("models"); p {
			if list, ok := modelsV.([]any); ok {
				names = extractTagNames(list)
			}
		}
	}
	return names, nil
}

// extractTagNames is the /api/tags models extraction: dicts → name or
// model (truthy), skipping the rest. The old/new ollama-py SDK shape
// duality collapses onto this one wire rule (documented divergence).
func extractTagNames(list []any) []string {
	names := []string{}
	for _, m := range list {
		om, ok := asOM(m)
		if !ok {
			continue
		}
		name := ""
		if nv, p := om.Get("name"); p && pyTruthy(nv) {
			name = strings.TrimSpace(pyStrOf(nv))
		}
		if name == "" {
			if mv, p := om.Get("model"); p && pyTruthy(mv) {
				name = strings.TrimSpace(pyStrOf(mv))
			}
		}
		if name != "" {
			names = append(names, name)
		}
	}
	return names
}

func pyStrOf(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// ---------------------------------------------------------------------------
// The timeout-wrapped helpers
// ---------------------------------------------------------------------------

// OllamaChatOptions carries ollama_chat's keyword half. JSON mode
// defaults ON (the reference's json_mode=True): the zero value keeps
// it on; set JSONModeOff to disable (the explicit json_mode=False
// calls). opts nil → JSON on.
type OllamaChatOptions struct {
	JSONModeOff bool
	NumCtx      int
	NumPredict  int
	OnWarn      func(string)
}

// OllamaChat builds the request per the kwargs law — options only when
// num_ctx is on (num_predict rides ONLY inside that branch — the
// reference's nested-branch quirk), format json when JSON mode is on —
// and calls client.Chat under the wall-clock timeout. Returns the
// assistant content string.
func OllamaChat(client OllamaAPI, model string, messages []Message,
	timeoutS float64, opts *OllamaChatOptions) (string, error) {
	req := &OllamaChatRequest{Model: model, Messages: messages}
	jsonMode := true
	if opts != nil && opts.JSONModeOff {
		jsonMode = false
	}
	if opts != nil {
		if opts.NumCtx > 0 {
			o := &OllamaOptions{NumCtx: opts.NumCtx}
			if opts.NumPredict > 0 {
				o.NumPredict = opts.NumPredict
			}
			req.Options = o
			if est := EstimateTokens(messages); est > opts.NumCtx &&
				opts.OnWarn != nil {
				opts.OnWarn(fmt.Sprintf(
					"prompt ≈%d tokens exceeds the context window "+
						"(num_ctx=%d) — the model may not see the start of "+
						"it. Raise llm_num_ctx in Settings if answers degrade.",
					est, opts.NumCtx))
			}
		}
	}
	if jsonMode {
		req.Format = "json"
	}
	return ChatWithTimeout(client, timeoutS, req)
}

// ChatWithTimeout is chat_with_timeout: client.Chat under the wall
// clock, then ResponseContent.
func ChatWithTimeout(client OllamaAPI, timeoutS float64,
	req *OllamaChatRequest) (string, error) {
	resp, err := callWithTimeoutResult(timeoutS, func(
		ctx context.Context) (*OllamaChatResponse, error) {
		return client.Chat(ctx, req)
	})
	if err != nil {
		return "", err
	}
	return ResponseContent(resp), nil
}

// ListModelsWithTimeout is list_models_with_timeout.
func ListModelsWithTimeout(client OllamaAPI,
	timeoutS float64) ([]string, error) {
	return callWithTimeoutResult(timeoutS, client.List)
}

// ResponseContent extracts the assistant text from the response shape:
// nil response or nil message → ”; the content as-is otherwise.
func ResponseContent(resp *OllamaChatResponse) string {
	if resp == nil || resp.Message == nil {
		return ""
	}
	return resp.Message.Content
}

// OllamaReachable is the quick socket-level check: is an Ollama server
// listening at baseURL? No HTTP, never raises.
func OllamaReachable(baseURL string, timeoutS float64) bool {
	raw := strings.TrimSpace(baseURL)
	if raw == "" {
		raw = DefaultOllamaBase
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	host := u.Hostname()
	if host == "" {
		host = "127.0.0.1"
	}
	port := u.Port()
	if port == "" {
		port = "11434"
	}
	conn, err := net.DialTimeout("tcp", net.JoinHostPort(host, port),
		secondsDur(timeoutS))
	if err != nil {
		return false
	}
	_ = conn.Close()
	return true
}

var _ = time.Second
