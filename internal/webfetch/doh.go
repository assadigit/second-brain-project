// doh.go — the DNS-over-HTTPS verdict (the owner's "getaddrinfo
// failed" class): when the DIRECT line could not resolve the name, one
// Cloudflare DoH question names whose fault it is — a dead domain
// (NXDOMAIN), a lying local resolver (the name exists out there), or
// no verdict at all. Cached per host, opt-out honored.
package webfetch

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"
)

// IsTLSHandshakeFailure — true when a connection-class reason names a
// TLS HANDSHAKE refusal (the fingerprint verdict). Certificate
// failures are deliberately excluded: a bad cert is the site's
// problem, not a wall to talk down. The Python marker list is checked
// case-insensitively; the Go runtime's own dialects are recognized
// alongside (the documented divergence — the marker law's intent
// transfers to the dialect the stdlib actually speaks).
func IsTLSHandshakeFailure(reason string) bool {
	low := strings.ToLower(reason)
	for _, marker := range TLSHandshakeMarkers {
		if strings.Contains(low, strings.ToLower(marker)) {
			return true
		}
	}
	for _, marker := range goTLSMarkerExtensions {
		if strings.Contains(low, marker) {
			return true
		}
	}
	return false
}

// IsDNSFailure — true when a connection-class reason says the NAME
// would not resolve.
func IsDNSFailure(reason string) bool {
	for _, marker := range DNSFailMarkers {
		if strings.Contains(reason, marker) {
			return true
		}
	}
	return false
}

// dohFn — the seam signature (the Python suites patch _doh_probe_once).
type dohFn func(host string, proxy *Proxy) string

// dohProbeOnceVar — the module-level seam.
var dohProbeOnceVar dohFn = dohProbeOnce

// The DoH verdict cache: a domain's existence rarely flips
// mid-process; one question per host, forever cached.
var (
	dohCache   = map[string]string{}
	dohMu      sync.Mutex
	dohVerdict = "unknown"
)

// ResetDoHCache — the test seam (the suites clear the cache per probe).
func ResetDoHCache() {
	dohMu.Lock()
	defer dohMu.Unlock()
	dohCache = map[string]string{}
}

// DoHVerdict — the cached verdict for a host.
func DoHVerdict(host string, proxy *Proxy) string {
	key := strings.ToLower(host)
	if key == "" {
		return "unknown"
	}
	dohMu.Lock()
	if v, ok := dohCache[key]; ok {
		dohMu.Unlock()
		return v
	}
	dohMu.Unlock()
	verdict := dohProbeOnceVar(key, proxy)
	dohMu.Lock()
	dohCache[key] = verdict
	dohMu.Unlock()
	return verdict
}

// dohProbeOnce — ONE DNS-over-HTTPS question to Cloudflare's resolver
// (JSON API): "exists" / "nxdomain" / "unknown". Rides the primary
// route (the proxy when one is configured — the local resolver is the
// suspect when direct DNS failed). Never errors.
func dohProbeOnce(host string, proxy *Proxy) string {
	client, err := buildClient(proxy, false, DoHTimeoutS)
	if err != nil {
		return "unknown"
	}
	req, err := http.NewRequest(http.MethodGet,
		DoHEndpoint+"?name="+url.QueryEscape(host)+"&type=A", nil)
	if err != nil {
		return "unknown"
	}
	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", UserAgent)
	// urllib's wire default is identity; Go's transport would add
	// "gzip" (and then transparently decode) — pin the reference's
	// wire shape.
	req.Header.Set("Accept-Encoding", "identity")
	resp, err := client.Do(req)
	if err != nil {
		return "unknown"
	}
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, ArchiveMaxBytes))
	if err != nil {
		return "unknown"
	}
	var data struct {
		Status float64 `json:"Status"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return "unknown"
	}
	switch data.Status {
	case 0:
		return "exists"
	case 3:
		return "nxdomain"
	}
	return "unknown"
}

// dnsVerdictSuffix — when the DIRECT line could not resolve the name,
// ask DNS-over-HTTPS whose fault it is: (reason suffix, isDead).
func dnsVerdictSuffix(directReason, rawURL string, proxy *Proxy, dohProbe bool) (string, bool) {
	if !dohProbe || !IsDNSFailure(directReason) {
		return "", false
	}
	u, err := url.Parse(rawURL)
	if err != nil {
		return "", false
	}
	host := strings.TrimSpace(u.Hostname())
	if host == "" {
		return "", false
	}
	switch DoHVerdict(host, proxy) {
	case "nxdomain":
		return " | dns: NXDOMAIN even via DNS-over-HTTPS — the domain " +
			"itself is dead (the local resolver was right)", true
	case "exists":
		return " | dns: the name resolves via DNS-over-HTTPS — the " +
			"local resolver is lying (poisoned DNS; the proxy route " +
			"is the door that can still reach it)", false
	}
	return " | dns: DNS-over-HTTPS itself unreachable — verdict unknown", false
}

// jsonUnmarshal — the archive JSON parse (kept tiny for the seam).
func jsonUnmarshal(body string, v any) error {
	return json.Unmarshal([]byte(body), v)
}

var _ = time.Now
