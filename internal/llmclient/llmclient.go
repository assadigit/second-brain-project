// llmclient.go — the core half of the llm_client.py port (T-30):
// timeouts, extract_json, the loopback law, the JSON-mode memo,
// estimate_tokens, resolve_task_model, the CloudLLMError family, and
// the Python-value helpers the package shares.
package llmclient

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// The module constants (llm_client.py lines 31-32, 236-261, 770-790).
// ---------------------------------------------------------------------------

const (
	// DefaultTimeoutS — chat calls; generous, models are slow to warm.
	DefaultTimeoutS = 300.0
	// DefaultListTimeoutS — list/connectivity; should be instant.
	DefaultListTimeoutS = 15.0

	// DefaultNumCtx — the explicit context window (tokens).
	DefaultNumCtx = 8192
	// DefaultMaxOutputTokens — 0 = leave the output cap to the server.
	DefaultMaxOutputTokens = 0
	// AnthropicFallbackMaxTokens — Anthropic REQUIRES max_tokens.
	AnthropicFallbackMaxTokens = 4096

	// CloudProviderLabel — the canonical cloud provider option label.
	CloudProviderLabel = "OpenAI-compatible endpoint"

	// AnthropicVersionHeader — the Messages-API version header.
	AnthropicVersionHeader = "2023-06-01"

	// LLamaCppProviderLabel — the llama.cpp provider option label.
	LLamaCppProviderLabel = "llama.cpp server (local)"
	// LLamaCppDefaultBase — llama-server's default address.
	LLamaCppDefaultBase = "http://127.0.0.1:8080"
	// LLamaCppProbeTimeoutS — per-port probe budget.
	LLamaCppProbeTimeoutS = 2.0

	// DefaultOllamaBase — Ollama's default address (detect_ollama /
	// ollama_reachable signatures).
	DefaultOllamaBase = "http://127.0.0.1:11434"
)

// AnthropicHostHints — substring hints that route an API URL onto the
// Messages wire.
var AnthropicHostHints = []string{"api.anthropic.com", "anthropic.com"}

// LLamaCppScanPorts — the ports the Detect button / auto-detect scan
// probes, in order (8080 is the llama-server default; the rest are
// common manual choices).
var LLamaCppScanPorts = []int{
	8080, 8081, 8082, 8083, 8084, 8085, 8000, 5000, 5001, 1234, 9000,
}

// ---------------------------------------------------------------------------
// The timeout wrapper (call_with_timeout — Go: contexts + a goroutine
// select for non-HTTP clients, per the spec's Go notes).
// ---------------------------------------------------------------------------

// TimeoutError is Python's TimeoutError with the app's message law:
// "LLM call did not respond within {n}s" (round-half-even, like
// Python's :.0f format).
type TimeoutError struct {
	Seconds float64
}

func (e *TimeoutError) Error() string {
	return fmt.Sprintf("LLM call did not respond within %.0fs", e.Seconds)
}

// callWithTimeoutResult runs fn under a wall-clock timeout. timeoutS
// <= 0 → the direct call (no deadline — mirrors Python's early return).
// On timeout the caller gets *TimeoutError; the fn's late result is
// discarded (the leaked-worker trade, Go-shaped: the goroutine runs to
// completion, nobody reads it).
func callWithTimeoutResult[T any](timeoutS float64,
	fn func(ctx context.Context) (T, error)) (T, error) {
	var zero T
	if timeoutS <= 0 {
		return fn(context.Background())
	}
	ctx, cancel := context.WithTimeout(
		context.Background(), secondsDur(timeoutS))
	defer cancel()
	type result struct {
		v   T
		err error
	}
	done := make(chan result, 1)
	go func() {
		v, err := fn(ctx)
		done <- result{v, err}
	}()
	select {
	case r := <-done:
		return r.v, r.err
	case <-ctx.Done():
		return zero, &TimeoutError{Seconds: timeoutS}
	}
}

// CallWithTimeout is the tested general wrapper (call_with_timeout):
// fn runs with a context carrying the deadline; a timeout surfaces as
// *TimeoutError.
func CallWithTimeout(timeoutS float64, fn func(ctx context.Context) error) error {
	_, err := callWithTimeoutResult(timeoutS, func(
		ctx context.Context) (struct{}, error) {
		return struct{}{}, fn(ctx)
	})
	return err
}

func secondsDur(seconds float64) time.Duration {
	return time.Duration(seconds * float64(time.Second))
}

// isDeadline reports whether err (possibly wrapped by net/http) is a
// context deadline — the wall-clock timeout firing.
func isDeadline(err error) bool {
	return errors.Is(err, context.DeadlineExceeded)
}

// ---------------------------------------------------------------------------
// Robust JSON extraction (extract_json)
// ---------------------------------------------------------------------------

// ExtractJSON robustly extracts JSON from LLM output: direct parse →
// ```json fenced block → the first balanced {…} (ONE attempt, then the
// break — the reference's law). Top-level non-objects parse and return
// as-is (Python json.loads semantics). Failure → the exact
// "No valid JSON found in response" error.
func ExtractJSON(text string) (any, error) {
	text = strings.TrimSpace(text)

	// Direct parse first.
	if v, err := storage.DecodeJSONValue([]byte(text)); err == nil {
		return v, nil
	}

	// Strip markdown fences.
	if m := fenceRe.FindStringSubmatch(text); m != nil {
		if v, err := storage.DecodeJSONValue(
			[]byte(strings.TrimSpace(m[1]))); err == nil {
			return v, nil
		}
	}

	// First { ... } block (greedy enough to handle nested).
	start := strings.IndexByte(text, '{')
	if start != -1 {
		depth := 0
		for i, c := range text[start:] {
			if c == '{' {
				depth++
			} else if c == '}' {
				depth--
				if depth == 0 {
					// One attempt at the first balanced block, then break.
					end := start + i + utf8.RuneLen(c)
					if v, err := storage.DecodeJSONValue(
						[]byte(text[start:end])); err == nil {
						return v, nil
					}
					break
				}
			}
		}
	}

	return nil, errors.New("No valid JSON found in response")
}

// fenceRe mirrors re.compile(r'```(?:json)?\s*(.*?)```', re.DOTALL).
var fenceRe = regexp.MustCompile("(?s)```(?:json)?\\s*(.*?)```")

// ---------------------------------------------------------------------------
// The loopback law (v0.15.1) — loopback traffic must NEVER ride a proxy.
// ---------------------------------------------------------------------------

// IsLoopbackURL reports whether the URL points at THIS machine
// (127.x.x.x — the whole /8, ::1, localhost; bare host:port spellings
// too). Python-semantics: an IPv6-spelled host must equal ::1 exactly —
// a v4-mapped ::ffff:127.0.0.1 is NOT loopback (ipaddress's law,
// pinned by the golden's no-table).
func IsLoopbackURL(rawURL string) bool {
	raw := strings.TrimSpace(rawURL)
	if raw == "" {
		return false
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
		return false
	}
	if strings.EqualFold(host, "localhost") {
		return true
	}
	if strings.Contains(host, ":") {
		// IPv6 spelling: the loopback network is ::1/128 only.
		ip := net.ParseIP(host)
		return ip != nil && ip.Equal(net.IPv6loopback)
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.To4() != nil && ip.To4()[0] == 127
}

// httpClient builds the per-call client: direct (Proxy nil) for
// loopback targets — a proxy can only break them; the environment
// proxy for everything else (cloud endpoints may need it). TLS
// verification is per-call (verify_tls defaults FALSE in the
// reference: self-hosted endpoints often run self-signed certs).
//
// Divergence note (doc.go): Go's ProxyFromEnvironment never proxies
// loopback addresses by net/http law, so an anthropic-shaped loopback
// URL with a proxy env behaves DIRECT here where the reference would
// attempt the proxy — an unreachable corner (cloud URLs are never
// loopback; recorded for honesty).
func httpClient(verifyTLS, direct bool) *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: !verifyTLS,
		},
	}
	if !direct {
		tr.Proxy = http.ProxyFromEnvironment
	}
	return &http.Client{Transport: tr}
}

// ---------------------------------------------------------------------------
// The JSON-mode memo (response_format rejections, process-lifetime)
// ---------------------------------------------------------------------------

var (
	jsonModeMu        sync.Mutex
	jsonModeRejected_ = map[string]struct{}{}
)

func jsonModeKey(apiURL string) string {
	return strings.ToLower(strings.TrimRight(apiURL, "/"))
}

// JSONModeRejected reports whether this base URL already rejected
// response_format (HTTP 400).
func JSONModeRejected(apiURL string) bool {
	jsonModeMu.Lock()
	defer jsonModeMu.Unlock()
	_, ok := jsonModeRejected_[jsonModeKey(apiURL)]
	return ok
}

func rememberJSONModeRejected(base string) {
	jsonModeMu.Lock()
	defer jsonModeMu.Unlock()
	jsonModeRejected_[strings.ToLower(base)] = struct{}{}
}

// ResetJSONModeMemo forgets every remembered rejection (tests only).
func ResetJSONModeMemo() {
	jsonModeMu.Lock()
	defer jsonModeMu.Unlock()
	jsonModeRejected_ = map[string]struct{}{}
}

// ---------------------------------------------------------------------------
// estimate_tokens / resolve_task_model
// ---------------------------------------------------------------------------

// Message is one chat turn ({role, content} — the Python dict shape).
type Message struct {
	Role    string
	Content string
}

// EstimateTokens is the rough prompt size (~4 chars per token plus a
// small per-message envelope). Warning-only, never a hard decision.
func EstimateTokens(messages []Message) int {
	total := 0
	for _, m := range messages {
		total += utf8.RuneCountInString(m.Content) + 16
	}
	return total / 4
}

// ResolveTaskModel honors per-task model overrides:
// config['models']['classify'/'analyze'] beat the default. Empty or
// missing overrides (or unknown task) keep defaultModel.
func ResolveTaskModel(config any, task, defaultModel string) string {
	if task == "" || !isMapLike(config) {
		return defaultModel
	}
	modelsV, _ := cfgGet(config, "models")
	if isMapLike(modelsV) {
		overrideV, _ := cfgGet(modelsV, task)
		override := strings.TrimSpace(helpers.PyStr(pyOr(overrideV, "")))
		if override != "" {
			return override
		}
	}
	return defaultModel
}

// ---------------------------------------------------------------------------
// The CloudLLMError family
// ---------------------------------------------------------------------------

// CloudLLMError is the base: a clear message, never a raw traceback.
type CloudLLMError struct {
	msg string
}

func (e *CloudLLMError) Error() string { return e.msg }

func newCloudLLMError(format string, args ...any) *CloudLLMError {
	return &CloudLLMError{msg: fmt.Sprintf(format, args...)}
}

// CloudLLMHTTPError — the endpoint answered with an HTTP error status.
type CloudLLMHTTPError struct {
	Code int
	Body string // the first 400 runes
	msg  string
}

// NewCloudLLMHTTPError builds the error with the reference's message
// law: "endpoint answered HTTP {code}: {body[:200]}" (rune-truncated).
func NewCloudLLMHTTPError(code int, body string) *CloudLLMHTTPError {
	b := helpers.CutRunes(body, 400)
	return &CloudLLMHTTPError{
		Code: code,
		Body: b,
		msg: fmt.Sprintf("endpoint answered HTTP %d: %s",
			code, helpers.CutRunes(b, 200)),
	}
}

func (e *CloudLLMHTTPError) Error() string { return e.msg }

// CloudLLMBadResponse — HTTP 200 but the body is not a usable chat
// completion (malformed JSON, an embedded error object, no choices).
type CloudLLMBadResponse struct {
	msg string
}

func (e *CloudLLMBadResponse) Error() string { return e.msg }

func newBadResponse(format string, args ...any) *CloudLLMBadResponse {
	return &CloudLLMBadResponse{msg: fmt.Sprintf(format, args...)}
}

// IsAnthropicURL — True when the base URL carries an Anthropic host
// hint (substring on the lowered URL; enterprise gateways keep the
// host name).
func IsAnthropicURL(apiURL string) bool {
	base := strings.ToLower(strings.TrimSpace(apiURL))
	for _, hint := range AnthropicHostHints {
		if strings.Contains(base, hint) {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// The Python-value helpers shared by this package
// ---------------------------------------------------------------------------

// pyTruthy mirrors Python truthiness for the JSON-decoded shapes.
func pyTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case int64:
		return t != 0
	case int:
		return t != 0
	case float64:
		return t != 0
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return f != 0
		}
		return true
	case []any:
		return len(t) > 0
	case *storage.OrderedMap:
		return t.Len() > 0
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}

// pyOr mirrors `x or fallback` (falsy x → fallback).
func pyOr(v, fallback any) any {
	if pyTruthy(v) {
		return v
	}
	return fallback
}

// pyInt mirrors Python int() for the guarded /props parses: bool → 1/0,
// float truncation toward zero, decimal strings (whitespace-tolerant),
// anything else → 0 (the callers' except TypeError/ValueError → 0).
// Python's int('5_0') underscore quirk is not replicated (documented).
func pyInt(v any) int {
	switch t := v.(type) {
	case nil:
		return 0
	case bool:
		if t {
			return 1
		}
		return 0
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case json.Number:
		if i, err := strconv.ParseInt(t.String(), 10, 64); err == nil {
			return int(i)
		}
		if f, err := t.Float64(); err == nil {
			return int(f)
		}
		return pyIntString(t.String())
	case string:
		return pyIntString(t)
	default:
		return 0
	}
}

func pyIntString(s string) int {
	i, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return i
}

// cfgGet reads a key from a config dict (either representation).
func cfgGet(config any, key string) (any, bool) {
	switch t := config.(type) {
	case nil:
		return nil, false
	case *storage.OrderedMap:
		return t.Get(key)
	case map[string]any:
		v, ok := t[key]
		return v, ok
	default:
		return nil, false
	}
}

// isMapLike reports whether v is a Python-dict-shaped value.
func isMapLike(v any) bool {
	switch v.(type) {
	case *storage.OrderedMap, map[string]any:
		return true
	}
	return false
}

// messagesToAny renders []Message as the ordered-dict list the wire
// writer expects (role first — the reference's dict literal order).
func messagesToAny(messages []Message) []any {
	out := make([]any, 0, len(messages))
	for _, m := range messages {
		om := storage.NewOrderedMap()
		om.Set("role", m.Role)
		om.Set("content", m.Content)
		out = append(out, om)
	}
	return out
}

// httpStatusError is urllib's HTTPError: a non-2xx status with the
// body readable (the probe/health paths consume the body).
type httpStatusError struct {
	code int
	body string
}

func (e *httpStatusError) Error() string {
	return fmt.Sprintf("HTTP Error %d", e.code)
}

// pyURLOpenError renders the URLError suffix shape — the runtime text
// inside differs (Go dial errors vs Python errno), the shape matches.
func pyURLOpenError(err error) string {
	return fmt.Sprintf("<urlopen error %s>", err)
}

// timeoutCtx applies the package's seconds convention.
func timeoutCtx(ctx context.Context, timeoutS float64) context.Context {
	c, cancel := timeoutCtxPair(ctx, timeoutS)
	// The caller cannot cancel mid-request (the reference's urlopen has
	// no cancel path either); the deadline rides the request context.
	_ = cancel //nolint:lostcancel // deferred-cancel pattern is unneeded here
	return c
}

// readAllBody drains the body (io.ReadAll with the package's error
// swallow — a truncated read still yields what arrived).
func readAllBody(resp *http.Response) (string, error) {
	b, err := io.ReadAll(resp.Body)
	return string(b), err
}

// reasonPhrase extracts the reason text from the status line
// ("404 Not Found" → "Not Found").
func reasonPhrase(resp *http.Response) string {
	status := resp.Status
	if i := strings.IndexByte(status, ' '); i >= 0 {
		return status[i+1:]
	}
	return ""
}
