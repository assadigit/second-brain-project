// impersonate.go — the third door (v0.45–v0.47): the handshake. When
// every stdlib door is walled with a refusal-family answer or a
// TLS-handshake refusal, the wall is aimed at the REQUESTER's
// fingerprint — the one thing the stdlib can never present. The door
// re-asks the URL with Chrome's own TLS/HTTP2 handshake via
// tls-client (ADR-0006's curl_cffi replacement, pinned v1.16.0), one
// attempt per allowed route, same politeness.
//
// THE GO LAW (ADR-0006 §4): the door ships armed — tls-client is
// compiled in, so impersonatorAvailable defaults TRUE and the install
// hint never appears in production. The availability/installer seams
// keep the reference's decision structure testable.
package webfetch

import (
	"io"
	"strings"
	"sync"
	"time"

	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// impResponse — the response shape the door's caller consumes (the
// curl_cffi Response's observable surface).
type impResponse struct {
	StatusCode int
	Header     headerGetter
	URL        string
	Charset    string
	Body       io.ReadCloser
}

// impSessionIface — the third door's dialer (the fake-session
// pattern one level down: the factory is package-level for exactly
// this). Get dials url with timeoutS, through proxyURL when non-empty
// (socks5h://… — DNS at the proxy), direct otherwise.
type impSessionIface interface {
	Get(url string, timeoutS float64, proxyURL string) (impResponse, error)
}

// The third door's state — every seam the reference's suites flip.
var (
	impMu sync.Mutex
	// impersonatorAvailable — curl_cffi_available()'s seam. The Go
	// door ships armed (true); tests flip it to replay the
	// missing-library scenarios.
	impersonatorAvailable = func() bool { return true }
	// pipInstallImpersonator — _pip_install_curl_cffi()'s seam. The Go
	// no-op (nothing to install — the library is compiled in).
	pipInstallImpersonator = func() bool { return false }
	// newImpersonationSession — the session factory (tests swap in
	// recording fakes).
	newImpersonationSession = func() impSessionIface {
		return newTLSSession()
	}
	// impSessionCached — the ONE cached session (the cookie-
	// persistence law: a clearance cookie earned on one domain keeps
	// working on the next visit).
	impSessionCached impSessionIface
	// the once-per-process autopip state.
	autopipTried bool
	autopipOK    bool
)

// ResetImpersonationSession — the test seam (a cached session must
// not outlive a re-arming).
func ResetImpersonationSession() {
	impMu.Lock()
	defer impMu.Unlock()
	impSessionCached = nil
}

// ensureImpersonatorInstalled — install the impersonator ONCE per
// process when the door is wanted but the library is missing, then
// re-probe. In Go the default installer is a no-op (the door ships
// armed — ADR-0006 §4); the seam keeps the once-per-process law
// testable.
func ensureImpersonatorInstalled() bool {
	impMu.Lock()
	if autopipTried {
		ok := autopipOK
		impMu.Unlock()
		return ok
	}
	autopipTried = true
	impMu.Unlock()
	// The installer runs WITHOUT the lock — it may flip the
	// availability seam (Python's probe-reset dance), and like the
	// reference's tried-then-install order, a concurrent caller sees
	// the stale not-yet-installed verdict.
	ok := pipInstallImpersonator()
	if ok {
		ok = impersonatorAvailable()
	}
	impMu.Lock()
	autopipOK = ok
	impMu.Unlock()
	return ok
}

// tlsImpSession — the REAL dialer: one session object sharing one
// cookie jar; tls-client takes its options per client, so each Get
// builds a client for (timeout, proxy) riding the shared jar —
// observably curl_cffi's per-request proxies over one session.
type tlsImpSession struct {
	jar tls_client.CookieJar
}

func newTLSSession() *tlsImpSession {
	return &tlsImpSession{jar: tls_client.NewCookieJar()}
}

// Get dials with Chrome's own handshake. No custom headers ride — the
// profile's own Chrome header set speaks (the same law as the
// reference's headerless session.get).
func (s *tlsImpSession) Get(rawURL string, timeoutS float64, proxyURL string) (impResponse, error) {
	secs := 1
	if timeoutS > 1 {
		secs = int(timeoutS)
	}
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(secs),
		// The PoC-verified Chrome profile (ADR-0006 Am.1: the
		// fingerprint-substitution proof) — the closest pinned
		// equivalent of curl_cffi's impersonate="chrome".
		tls_client.WithClientProfile(profiles.Chrome_133),
		tls_client.WithCookieJar(s.jar),
	}
	if proxyURL != "" {
		options = append(options, tls_client.WithProxyUrl(proxyURL))
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return impResponse{}, err
	}
	resp, err := client.Get(rawURL)
	if err != nil {
		return impResponse{}, err
	}
	finalURL := rawURL
	if resp.Request != nil && resp.Request.URL != nil {
		finalURL = resp.Request.URL.String()
	}
	return impResponse{
		StatusCode: resp.StatusCode,
		Header:     resp.Header,
		URL:        finalURL,
		Body:       resp.Body,
	}, nil
}

// ProxyURLForImpersonator — the config proxy as the door's proxy URL.
// SOCKS becomes socks5h:// — the SAME DNS-at-the-proxy law the stdlib
// door's rdns=True enforces.
func ProxyURLForImpersonator(p *Proxy) string {
	if p == nil {
		return ""
	}
	var scheme string
	switch strings.ToLower(strings.TrimSpace(p.Type)) {
	case "socks5":
		scheme = "socks5h"
	case "socks4":
		scheme = "socks4"
	case "http":
		scheme = "http"
	default:
		return ""
	}
	auth := ""
	if strings.TrimSpace(p.Username) != "" {
		auth = pyQuote(p.Username) + ":" + pyQuote(p.Password) + "@"
	}
	return scheme + "://" + auth + p.Host + ":" + itoa(p.Port)
}

func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	neg := i < 0
	if neg {
		i = -i
	}
	var buf [20]byte
	pos := len(buf)
	for i > 0 {
		pos--
		buf[pos] = byte('0' + i%10)
		i /= 10
	}
	if neg {
		pos--
		buf[pos] = '-'
	}
	return string(buf[pos:])
}

// maybeImpersonateDirect — the seam FetchURL consults (the Python
// suites spy on _maybe_impersonate itself).
var maybeImpersonateDirect = MaybeImpersonate

// MaybeImpersonate — the third door's own law. routes is the door
// sequence the CALLER allows (FetchURL computes it) — the third door
// never widens a route the owner opted out of. Fires when the stdlib
// door(s) are walled with a refusal-family answer (the wall aimed at
// the REQUESTER) or a TLS-handshake refusal; a plain timeout or DNS
// failure stays a network truth; resource truths never reach here;
// loopback never gets ANY door's escalation. One impersonated attempt
// per allowed route. Success via the third door is a REAL success
// with the wall named; a challenge that survives the Chrome handshake
// gets the honest three-leg line plus the "needs a live browser"
// verdict.
func MaybeImpersonate(rawURL string, timeoutS float64, maxBytes int,
	routes []*Proxy, enabled bool, walled *FetchResult,
	started time.Time, autopip bool) *FetchResult {
	if !enabled || IsLoopbackURL(rawURL) {
		return walled
	}
	if walled.HTTPStatus != nil {
		if !isRouteRefusal(*walled.HTTPStatus) {
			return walled // the site answered about the RESOURCE
		}
	} else if !IsTLSHandshakeFailure(walled.Reason) {
		return walled // connection-class truth — not a fingerprint verdict
	}
	if !impersonatorAvailable() {
		if autopip && ensureImpersonatorInstalled() {
			// the door armed itself; dial below
		} else {
			walled.Reason = (walled.Reason) + ImpInstallHint
			return walled
		}
	}
	var tried []string
	for _, route := range routes {
		res := impersonatedFetchOnce(rawURL, timeoutS, maxBytes, route, started)
		if res.OK() {
			res.Reason = "via impersonated Chrome (curl_cffi) — the " +
				"stdlib doors were walled: " + walled.Reason
			return res
		}
		if res.Reason != "" {
			tried = append(tried, res.Reason)
		}
	}
	combined := strings.Join(dedupKeepOrder(tried), " / ")
	walled.Reason = walled.Reason + " | chrome-impersonated: " + combined
	if strings.Contains(strings.ToLower(combined), "challenge") {
		walled.Reason += ChallengeWonHint
	}
	return walled
}

// dedupKeepOrder — dict.fromkeys semantics.
func dedupKeepOrder(items []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, it := range items {
		if !seen[it] {
			seen[it] = true
			out = append(out, it)
		}
	}
	return out
}

// impersonatedFetchOnce — one fetch attempt through the impersonated
// Chrome handshake: the browser's own TLS/HTTP2 fingerprint, the same
// FetchResult contract as fetchOnce (streamed, size-capped,
// charset-aware, pdf-aware). The body arrives already decompressed
// (the library decodes what it asked for), so the cap bounds the
// DEcompressed bytes — the same budget law as the stdlib door.
func impersonatedFetchOnce(rawURL string, timeoutS float64, maxBytes int,
	proxy *Proxy, started time.Time) *FetchResult {
	proxyURL := ProxyURLForImpersonator(proxy)
	impMu.Lock()
	sess := impSessionCached
	if sess == nil {
		sess = newImpersonationSession()
		impSessionCached = sess
	}
	impMu.Unlock()
	resp, err := sess.Get(rawURL, timeoutS, proxyURL)
	if err != nil {
		// The library raises its own errors (timeouts, connect
		// failures, bad TLS) — one honest reason, never a panic.
		reason := "connection: " + err.Error()
		if strings.Contains(strings.ToLower(err.Error()), "timeout") {
			reason = "timeout"
		}
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: reason, ElapsedS: elapsedSince(started),
		}
	}
	httpStatus := 0
	if resp.StatusCode != 0 {
		httpStatus = resp.StatusCode
	}
	headers := resp.Header
	contentType := ""
	if headers != nil {
		contentType = strings.ToLower(headers.Get("Content-Type"))
	}
	if httpStatus >= 400 {
		reason := "HTTP " + itoa(httpStatus)
		if httpStatus == 403 || httpStatus == 429 {
			if headers != nil {
				if hint := BotDefenseHint(headers); hint != "" {
					reason = "HTTP " + itoa(httpStatus) + " — " + hint
				}
			}
		}
		st := httpStatus
		return &FetchResult{
			URL: rawURL, FinalURL: rawURL, Status: "failed",
			Reason: reason, HTTPStatus: &st,
			ElapsedS: elapsedSince(started),
		}
	}
	charset := strings.TrimSpace(resp.Charset)
	if charset == "" {
		charset = charsetFromContentType(contentType)
	}
	finalURL := resp.URL
	if finalURL == "" {
		finalURL = rawURL
	}
	// Stream with the size cap — same 64 KB chunk law as fetchOnce.
	var chunks [][]byte
	total := 0
	truncated := false
	if resp.Body != nil {
		defer func() { _ = resp.Body.Close() }()
		buf := make([]byte, 64*1024)
	readLoop:
		for {
			n, rerr := resp.Body.Read(buf)
			if n > 0 {
				total += n
				if total > maxBytes {
					keep := n - (total - maxBytes)
					chunks = append(chunks, append([]byte(nil), buf[:keep]...))
					truncated = true
					break readLoop
				}
				chunks = append(chunks, append([]byte(nil), buf[:n]...))
			}
			if rerr != nil {
				break readLoop
			}
		}
	}
	body := joinChunks(chunks)
	elapsedS := elapsedSince(started)
	st := httpStatus
	if truncated {
		return &FetchResult{
			URL: rawURL, FinalURL: finalURL, Status: "partial",
			Reason:     "size cap (" + humanizeInt(maxBytes) + " bytes) reached",
			HTTPStatus: &st, ContentType: contentType, Charset: charset,
			Body: body, Text: DecodeBestEffort(body, charset),
			ElapsedS: elapsedS,
		}
	}
	if LooksLikePDF(contentType, rawURL, body) {
		return &FetchResult{
			URL: rawURL, FinalURL: finalURL, Status: "partial",
			Reason: "pdf", HTTPStatus: &st, ContentType: contentType,
			Charset: charset, ElapsedS: elapsedS,
		}
	}
	return &FetchResult{
		URL: rawURL, FinalURL: finalURL, Status: "full",
		Reason: "", HTTPStatus: &st, ContentType: contentType,
		Charset: charset, Body: body,
		Text:     DecodeBestEffort(body, charset),
		ElapsedS: elapsedS,
	}
}

func elapsedSince(started time.Time) float64 {
	return time.Since(started).Seconds()
}
