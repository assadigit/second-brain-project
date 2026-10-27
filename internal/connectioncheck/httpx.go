package connectioncheck

// httpx.go — the battery's small network helpers: _http_get_json
// (UA GitCurator-connection-check, loopback-direct), _github_get
// (Bearer + the GitHub Accept header), and the Worker's proxy-bypassed
// /health fetch. TLS verification is ON for all of them (Python's
// plain urlopen — the verify_ssl switch belongs to netctx's callers,
// never to these probes).

import (
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// connectionUA is the battery's own identity on every probe.
const connectionUA = "GitCurator-connection-check"

// GitHubAPIBase is the GitHub REST root (api.github.com is NOT
// loopback, so the environment proxy is honored exactly like the
// pipeline's own calls).
const GitHubAPIBase = "https://api.github.com"

// valueError mirrors Python's ValueError (raised by _http_get_json on
// non-object JSON; errClass renders the class name).
type valueError struct{ msg string }

func (e *valueError) Error() string { return e.msg }

// unicodeDecodeError mirrors Python's UnicodeDecodeError (invalid
// UTF-8 response bodies).
type unicodeDecodeError struct{ msg string }

func (e *unicodeDecodeError) Error() string { return e.msg }

// errClass renders the exception-class token for detail strings:
// Python's urllib taxonomy where a mapping exists, Go's %T base name
// otherwise (doc.go's divergence record).
func errClass(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		if ue.Timeout() {
			return "TimeoutError"
		}
		return "URLError"
	}
	var ve *valueError
	if errors.As(err, &ve) {
		return "ValueError"
	}
	var de *unicodeDecodeError
	if errors.As(err, &de) {
		return "UnicodeDecodeError"
	}
	s := fmt.Sprintf("%T", err)
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimPrefix(s, "*")
}

// plainClient is Python's plain urllib.request.urlopen: the
// environment proxy is honored, certificates verified, per-request
// timeout. (Go's ProxyFromEnvironment never proxies loopback — the
// reference reaches the same corner through its own law.)
func plainClient(timeoutS float64) *http.Client {
	return &http.Client{
		Timeout: secondsDur(timeoutS),
		Transport: &http.Transport{
			Proxy:           http.ProxyFromEnvironment,
			TLSClientConfig: &tls.Config{},
		},
	}
}

// directClient is _urlopen_direct: NO proxy, ever (loopback targets
// and the Worker's /health — a proxy client can only break them).
func directClient(timeoutS float64) *http.Client {
	return &http.Client{
		Timeout: secondsDur(timeoutS),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}
}

// clientForURL is the loopback law as one routing decision: loopback
// targets get the direct client (the v0.15.1 rule — a proxy can only
// break them), everything else the environment-proxy client.
func clientForURL(u string, timeoutS float64) *http.Client {
	if llmclient.IsLoopbackURL(u) {
		return directClient(timeoutS)
	}
	return plainClient(timeoutS)
}

// httpGetJSON GETs url → the parsed JSON dict. Loopback targets
// bypass the environment proxy (a VPN client can only break them);
// everything else honors it. Raises (returns an error) on any failure
// — callers always catch. UA: GitCurator-connection-check.
func httpGetJSON(rawURL string, timeoutS float64) (any, error) {
	req, err := http.NewRequest("GET", rawURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", connectionUA)
	client := clientForURL(rawURL, timeoutS)
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(body) {
		return nil, &unicodeDecodeError{"invalid UTF-8 in response body"}
	}
	raw := string(body)
	if raw == "" {
		raw = "{}"
	}
	payload, perr := storage.DecodeJSONValue([]byte(raw))
	if perr != nil {
		return nil, &valueError{perr.Error()}
	}
	if !isDictLike(payload) {
		return nil, &valueError{"response is not a JSON object"}
	}
	return payload, nil
}

// isDictLike reports whether v is a Python-dict-shaped value.
func isDictLike(v any) bool {
	switch v.(type) {
	case *storage.OrderedMap, map[string]any:
		return true
	}
	return false
}

// githubReply is _github_get's (status, payload, err) triple —
// status 0 means transport failure.
type githubReply struct {
	status  int
	payload any
	err     string
}

// githubGet GETs apiBase+path with the Bearer token. Never raises;
// status 0 means transport failure. Headers: Authorization Bearer,
// Accept application/vnd.github+json, UA GitCurator-connection-check.
// The environment proxy is honored (never a loopback branch — the
// reference has none here; Go's own loopback law covers the corner).
func githubGet(path, token, apiBase string, timeoutS float64) githubReply {
	u := strings.TrimRight(apiBase, "/") + path
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return githubReply{0, nil, errClassColon(err)}
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", connectionUA)
	res, err := plainClient(timeoutS).Do(req)
	if err != nil {
		return githubReply{0, nil, errClassColon(err)}
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	if !utf8.Valid(body) {
		return githubReply{res.StatusCode, nil, ""}
	}
	raw := string(body)
	if raw == "" {
		raw = "{}"
	}
	payload, perr := storage.DecodeJSONValue([]byte(raw))
	if perr != nil {
		if res.StatusCode >= 200 && res.StatusCode < 300 {
			return githubReply{res.StatusCode, nil, "bad JSON from server"}
		}
		return githubReply{res.StatusCode, nil, ""}
	}
	return githubReply{res.StatusCode, payload, ""}
}

// errClassColon renders "{class}: {msg}" (the transport-failure
// detail shape).
func errClassColon(err error) string {
	return errClass(err) + ": " + err.Error()
}

// WorkerHTTPGet is the injectable Worker /health fetch
// ((url, timeout_s) → the parsed dict, or an error).
type WorkerHTTPGet func(u string, timeoutS float64) (any, error)

// workerDirectGet is the production Worker fetch — the /health GET
// with the connection UA, proxy BYPASSED exactly like the sync client
// it mirrors (cloudflare_sync): owners behind an intercepting proxy
// would otherwise see a false ⚠️/❌ for a Worker that works. A package
// seam so the battery tests can inject (the reference monkeypatches).
var workerDirectGet WorkerHTTPGet = func(u string, timeoutS float64) (any, error) {
	// THE TEST-CONNECTION-DOH LAW (SESSION-080): the Test Connection's
	// Worker /health check rides the SAME DoH fallback dialer the sync
	// channel uses (workerclient.DirectClient) — so "Worker not
	// connected" doesn't lie when the only thing failing is the ISP's
	// DNS poisoning. The plain directClient used the system resolver;
	// the DoH client bypasses it (Cloudflare's 1.1.1.1 → resolve → dial
	// the IP directly with the original name's TLS).
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", connectionUA)
	res, err := workerclient.DirectClient(timeoutS).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}
	if !utf8.Valid(body) {
		return nil, &unicodeDecodeError{"invalid UTF-8 in response body"}
	}
	raw := string(body)
	if raw == "" {
		raw = "{}"
	}
	payload, perr := storage.DecodeJSONValue([]byte(raw))
	if perr != nil {
		return nil, &valueError{perr.Error()}
	}
	if !isDictLike(payload) {
		return nil, &valueError{"non-dict /health payload"}
	}
	return payload, nil
}

// HTTPGetJSONFn is the battery's _http_get_json seam: the reference's
// own suites monkeypatch the module attribute (mock.patch.object(cc,
// "_http_get_json", …)), and the CLI golden replay injects the same
// way — the exported var is that mock surface (production: the real
// probe above; nothing else changes).
var HTTPGetJSONFn = httpGetJSON

// GitHubReply is _github_get's return triple as the seam's shape:
// Status 0 means transport failure.
type GitHubReply struct {
	Status  int
	Payload any
	Err     string
}

// GitHubGetFn is the _github_get seam: (path, token, apiBase,
// timeoutS) → the reply triple (the reference's tests monkeypatch the
// module attribute the same way; the CLI golden replay injects the
// same way — production: the real probe above).
var GitHubGetFn = func(path, token, apiBase string, timeoutS float64) GitHubReply {
	r := githubGet(path, token, apiBase, timeoutS)
	return GitHubReply{Status: r.status, Payload: r.payload, Err: r.err}
}

// secondsDur mirrors llmclient's seconds convention (kept local: a
// tiny, obvious helper — the package never imports unexported state).
func secondsDur(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}
