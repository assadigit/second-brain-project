// webfetch.go — the constants, FetchResult, the rate limiter, and the
// small pure helpers (web_fetch.py's module head).
package webfetch

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

// The browser-grade UA (v0.41.0): the pre-v0.41 honest string was
// 403-walled by CDN bot defenses on working sites. Overridable per
// install via config "web_user_agent".
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) " +
	"AppleWebKit/537.36 (KHTML, like Gecko) " +
	"Chrome/131.0.0.0 Safari/537.36"

const (
	DefaultProxyType       = "socks5" // v2rayN's default listener
	DefaultProxyHost       = "127.0.0.1"
	DefaultProxyPort       = 10808
	ProxyPreflightTimeoutS = 2.5 // proxy TCP reachability probe
	DefaultTimeoutS        = 20.0
	DefaultMaxBytes        = 2_000_000 // 2 MB body cap
	DefaultDomainDelayS    = 2.0       // min seconds between hits on one domain
	MaxRedirects           = 5
	RatePenaltyDefaultS    = 60.0
	RatePenaltyCapS        = 300.0
	DoHTimeoutS            = 5.0
	ArchiveMaxBytes        = 65_536
	OriginArchiveTimeoutS  = 15.0
	ImpersonateTarget      = "chrome"
	AutopipTimeoutS        = 240.0
	ImpInstallHint         = " — a third door exists: pip install curl_cffi (the browser-TLS handshake)"
	ChallengeWonHint       = " — the challenge needs a live browser; the fetcher cannot solve it"
)

// DoHEndpoint / ArchiveAPI are vars (the module attrs the reference's
// own tests patch; the golden generator patches them to local
// recorders — the same seam, one level up from the function seams).
var (
	DoHEndpoint = "https://cloudflare-dns.com/dns-query"
	ArchiveAPI  = "https://archive.org/wayback/available"
)

// RedirectStatuses — v0.46.0 THE LADDER: the codes our own loop follows.
var RedirectStatuses = []int{301, 302, 303, 307, 308}

// CloudflareOriginStatuses — "the site's own server is broken" answers
// (521-524): site-side truths, never the fetcher's fault.
var CloudflareOriginStatuses = map[int]string{
	521: "web server is down",
	522: "connection to the origin timed out",
	523: "origin is unreachable",
	524: "timed out waiting for the origin",
}

// TLSHandshakeMarkers — connection-class reasons that mean "the TLS
// handshake itself was refused" (a fingerprint verdict, the third
// door's trigger). Certificate failures are deliberately NOT here.
var TLSHandshakeMarkers = []string{
	"SSLV3_ALERT_HANDSHAKE_FAILURE", "alert handshake failure",
	"handshake operation timed out", "handshake timed out",
	"UNEXPECTED_EOF_WHILE_READING", "EOF occurred in violation of protocol",
	"WRONG_VERSION_NUMBER", "TLSV1_ALERT", "sslv3 alert",
}

// goTLSMarkerExtensions — the Go runtime's own dialects of the same
// handshake-refusal verdicts (documented divergence: the marker law's
// intent transfers to the dialect the stdlib actually speaks).
var goTLSMarkerExtensions = []string{
	"tls: handshake failure", "tls handshake timeout",
	"first record does not look like a tls handshake",
	"server gave http response to https client",
	"tls: wrong version number",
}

// DNSFailMarkers — connection-class reasons that mean "the NAME would
// not resolve" (the DNS verdict's trigger).
var DNSFailMarkers = []string{
	"getaddrinfo", "name or service not known", "no such host",
	"temporary failure in name resolution", "nodename nor servname",
	"WSAHOST_NOT_FOUND", "Name or service not known",
}

// CategoryPrecedence / NoHealCategories — what the PIPELINE reads to
// retry only what time can heal.
var (
	CategoryPrecedence = []string{"paywalled", "dead", "refused",
		"redirect_broken", "blocked_bot", "proxy_error", "retry_later"}
	NoHealCategories = []string{"dead", "paywalled", "refused"}
)

// RouteRefusalStatuses — HTTP answers aimed at the ROUTE, not the page
// (the WAF wall, the method-block, the per-IP rate limit, the
// geo-block). The OTHER route still gets its one attempt.
var RouteRefusalStatuses = []int{403, 405, 429, 451}

// BotDefenseServers — Server values that, on a 403/429, mean "a
// bot-defense answered".
var BotDefenseServers = []string{"cloudflare", "akamai", "cloudfront",
	"fastly", "imperva", "incapsula", "sucuri", "edgecast"}

// HTMLTypes — content types HTML-ish enough to parse.
var HTMLTypes = []string{"text/html", "application/xhtml", "text/plain"}

// ProxyTypes — the legal proxy config types.
var ProxyTypes = []string{"socks5", "socks4", "http"}

func isRedirectStatus(code int) bool {
	for _, c := range RedirectStatuses {
		if c == code {
			return true
		}
	}
	return false
}

func isRouteRefusal(code int) bool {
	for _, c := range RouteRefusalStatuses {
		if c == code {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// FetchResult
// ---------------------------------------------------------------------------

// FetchResult — one fetch attempt. Body is the raw (capped) bytes;
// Text is the best-effort decode (charset from headers; web_extract
// re-decodes with the page's own <meta charset> when the header lies).
// HTTPStatus nil = no HTTP answer; RetryAfterS nil = none given.
type FetchResult struct {
	URL         string
	FinalURL    string
	Status      string // full | partial | failed
	Reason      string
	HTTPStatus  *int
	ContentType string
	Charset     string
	Body        []byte
	Text        string
	ElapsedS    float64
	Category    string
	RetryAfterS *float64
}

// OK — true when SOMETHING was fetched (full or partial).
func (r *FetchResult) OK() bool { return r.Status != "failed" }

// ---------------------------------------------------------------------------
// The rate limiter
// ---------------------------------------------------------------------------

// RateLimiter is the politeness seam (the pipeline injects its own;
// the Python suites inject recording fakes).
type RateLimiter interface {
	Wait(domain string)
	Penalize(domain string, seconds float64)
}

// DomainRateLimiter — minimum delay between requests to the same
// domain. One instance is shared by a whole batch; thread-safe.
type DomainRateLimiter struct {
	DelayS       float64
	mu           sync.Mutex
	lastHit      map[string]float64
	penaltyUntil map[string]float64
}

// NewDomainRateLimiter builds a limiter with the given per-domain
// delay (<= 0 disables the plain delay; penalties still apply).
func NewDomainRateLimiter(delayS float64) *DomainRateLimiter {
	if delayS < 0 {
		delayS = 0
	}
	return &DomainRateLimiter{
		DelayS:       delayS,
		lastHit:      map[string]float64{},
		penaltyUntil: map[string]float64{},
	}
}

// Wait blocks (sleeps in <=1 s slices) until hitting domain is polite
// again — a domain carrying a PAID penalty waits until it expires too.
// Sleeps in slices so a big penalty never locks the thread
// unresponsively.
func (l *DomainRateLimiter) Wait(domain string) {
	hasPenalty := false
	l.mu.Lock()
	_, hasPenalty = l.penaltyUntil[domain]
	l.mu.Unlock()
	if domain == "" || (l.DelayS <= 0 && !hasPenalty) {
		return
	}
	for {
		l.mu.Lock()
		now := monotonic()
		last := l.lastHit[domain]
		due := last + l.DelayS
		if p, ok := l.penaltyUntil[domain]; ok && p > due {
			due = p
		}
		if now >= due {
			l.lastHit[domain] = now
			l.mu.Unlock()
			return
		}
		sleepFor := due - now
		l.mu.Unlock()
		if sleepFor > 1.0 {
			sleepFor = 1.0
		}
		time.Sleep(time.Duration(sleepFor * float64(time.Second)))
	}
}

// Penalize — the site answered 429/503 (maybe with Retry-After): the
// domain's NEXT hit waits this much longer. Never panics.
func (l *DomainRateLimiter) Penalize(domain string, seconds float64) {
	if domain == "" || seconds <= 0 {
		return
	}
	until := monotonic() + seconds
	l.mu.Lock()
	defer l.mu.Unlock()
	if until > l.penaltyUntil[domain] {
		l.penaltyUntil[domain] = until
	}
}

func monotonic() float64 { return float64(time.Now().UnixNano()) / 1e9 }

// ---------------------------------------------------------------------------
// The small pure helpers
// ---------------------------------------------------------------------------

// IsLoopbackURL — true when the URL points at THIS machine (127.x /
// ::1 / localhost). Loopback traffic must never ride a proxy.
func IsLoopbackURL(raw string) bool {
	raw = strings.TrimSpace(raw)
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
	host := strings.ToLower(u.Hostname())
	if host == "" {
		return false
	}
	if host == "localhost" {
		return true
	}
	if ip := net.ParseIP(host); ip != nil {
		return ip.IsLoopback()
	}
	return false
}

// DomainOf — the netloc (host:port), lowercased; "" when unparseable.
func DomainOf(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return ""
	}
	return strings.ToLower(u.Host)
}

// DecodeBestEffort — the header charset, else UTF-8 with replacement
// (a windows-1252 page decoded as UTF-8 keeps its text with a few
// replacement chars; web_extract re-decodes from Body when the page's
// own <meta charset> disagrees with the header).
func DecodeBestEffort(body []byte, charset string) string {
	for _, cs := range []string{charset, "utf-8"} {
		if cs == "" {
			continue
		}
		if dec := decodeCharset(body, cs); dec != nil {
			return *dec
		}
	}
	return string(bytesReplaceInvalid(body))
}

// pyTruthy — Python truthiness for config-read values (bools, strings,
// numbers; the reference's dict semantics).
func pyTruthy(v any) bool {
	switch x := v.(type) {
	case nil:
		return false
	case bool:
		return x
	case string:
		return x != ""
	case int:
		return x != 0
	case int64:
		return x != 0
	case float64:
		return x != 0
	case []any:
		return len(x) > 0
	case map[string]any:
		return len(x) > 0
	default:
		return true
	}
}

// pyQuote — Python's urllib.parse.quote(s, safe=”): percent-encode
// everything but the unreserved set [A-Za-z0-9-_.~].
func pyQuote(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '-' || c == '_' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0F])
		}
	}
	return b.String()
}

// humanizeInt — Python's '{n:,}' comma formatting (the size-cap
// reason's byte count).
func humanizeInt(n int) string {
	s := strconv.Itoa(n)
	if len(s) <= 3 {
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	return strings.Join(parts, ",")
}
