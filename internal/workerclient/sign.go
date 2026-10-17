package workerclient

// sign.go — the HMAC request signing law (v30's fix: the KEY is the
// sha256 HEX STRING's UTF-8 bytes — the same sha256-hex-of-secret the
// Worker stores at pairing) + _make_request (never raises: network
// failure = status 0 + {'error': ...}).

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/tls"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// HTTPUserAgent is the channel's own name — THE 1010 LAW: the edge
// bans urllib's and Go's default signatures with a 403 that never
// reaches the Worker's code; the app's honest identity passes. Kept in
// lockstep with the reference's VERSION at release time.
const HTTPUserAgent = "GitCurator/0.91.0 (+https://github.com/assadigit/GitCurator)"

// nowMS is the clock seam (str(int(time.time() * 1000))): the ms
// epoch decimal string's source — swappable for the signature vectors
// and the window tests.
var nowMS = func() int64 {
	return time.Now().UnixMilli()
}

// DebugLog is the diagnostics hook (the reference prints
// "[CloudflareSync] ... failed: ..." to stdout; the port stays silent
// by default — the wire behavior is the contract, not the console).
var DebugLog func(format string, args ...any)

func debugf(format string, args ...any) {
	if DebugLog != nil {
		DebugLog(format, args...)
	}
}

// SignHeaders generates the HMAC signature headers for a request —
// the _sign_request port. body is the EXACT wire body string ('{}'
// when no body); path carries its query when present.
func (s *Sync) SignHeaders(method, path, body string) *storage.OrderedMap {
	timestamp := fmt.Sprintf("%d", nowMS())
	bodyHash := sha256Hex([]byte(body))
	message := method + "\n" + path + "\n" + timestamp + "\n" + bodyHash

	// Key = sha256(raw shared secret) as lowercase hex — matches the
	// Worker's stored shared_secret_hash exactly; hmacSha256 uses the
	// HEX STRING's UTF-8 bytes as the key.
	keyHex := sha256Hex([]byte(s.SharedSecret()))

	mac := hmac.New(sha256.New, []byte(keyHex))
	mac.Write([]byte(message))
	signature := hex.EncodeToString(mac.Sum(nil))

	om := storage.NewOrderedMap()
	om.Set("X-Auth-Install", s.InstallID())
	om.Set("X-Auth-Timestamp", timestamp)
	om.Set("X-Auth-Signature", signature)
	om.Set("Content-Type", "application/json")
	om.Set("User-Agent", HTTPUserAgent)
	return om
}

func sha256Hex(b []byte) string {
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// reply is _make_request's (status_code, response_json): status 0 =
// transport failure (the error rides data['error']).
type reply struct {
	status int
	data   any
}

// directClient is the proxy-BYPASSED opener (ProxyHandler({})):
// workers.dev was intercepted by local v2rayN into 403s — no proxy,
// ever, on any leg of this channel.
func directClient(timeoutS float64) *http.Client {
	return &http.Client{
		Timeout: secondsDur(timeoutS),
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{},
		},
	}
}

func secondsDur(seconds float64) time.Duration {
	if seconds <= 0 {
		return 0
	}
	return time.Duration(seconds * float64(time.Second))
}

// MakeRequest makes an HMAC-signed request to the Worker — never
// raises: transport failure = status 0 + {'error': str(e)}. The body
// (always present, '{}' for the GETs — the reference sends
// data=body_str on every leg, GET included) is signed byte-for-byte.
func (s *Sync) MakeRequest(method, path string, body *storage.OrderedMap) reply {
	bodyStr := "{}"
	if body != nil && body.Len() > 0 {
		bodyStr = storage.PyDumpsCompact(body)
	}
	headers := s.SignHeaders(method, path, bodyStr)
	u := s.WorkerURL() + path
	req, err := http.NewRequest(method, u, strings.NewReader(bodyStr))
	if err != nil {
		return reply{0, errMap("could not build request: " + err.Error())}
	}
	for _, k := range headers.Keys() {
		v, _ := headers.Get(k)
		req.Header.Set(k, pyStr(v))
	}
	res, err := directClient(30.0).Do(req)
	if err != nil {
		return reply{0, errMap(err.Error())}
	}
	defer res.Body.Close()
	raw, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return reply{res.StatusCode, errMap(rerr.Error())}
	}
	data, perr := storage.DecodeJSONValue(raw)
	if perr != nil {
		// the reference's except → {'error': str(e)} with the status
		return reply{res.StatusCode, errMap(perr.Error())}
	}
	return reply{res.StatusCode, data}
}

func errMap(msg string) *storage.OrderedMap {
	om := storage.NewOrderedMap()
	om.Set("error", msg)
	return om
}

// successLaw: status==200 AND data['success'] truthy.
func successLaw(r reply) bool {
	return r.status == 200 && pyTruthy(dictValue(r.data, "success"))
}

// errClass renders the exception-class token for the pair error
// mapping (JSONDecodeError for parse failures, the %T base otherwise).
func errClass(err error) string {
	var jd *jsonDecodeError
	if errors.As(err, &jd) {
		return "JSONDecodeError"
	}
	s := fmt.Sprintf("%T", err)
	if i := strings.LastIndexByte(s, '.'); i >= 0 {
		s = s[i+1:]
	}
	return strings.TrimPrefix(s, "*")
}

type jsonDecodeError struct{ msg string }

func (e *jsonDecodeError) Error() string { return e.msg }

// urlErrorReason renders URLError's e.reason (the inner cause text).
func urlErrorReason(err error) string {
	var ue *url.Error
	if errors.As(err, &ue) {
		return ue.Err.Error()
	}
	return err.Error()
}
