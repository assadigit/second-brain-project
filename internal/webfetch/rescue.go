// rescue.go — the dead-page rescue ladder (URL variants, then the
// Wayback Machine), the polite circuit breaker's price, the category
// combination, and the final verdict every failure shares (the DNS
// verdict, the rate penalty, the origin-down Wayback courtesy).
package webfetch

import (
	"net"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// archiveFn — the seam signature (the Python suites patch
// _archive_lookup).
type archiveFn func(url string, proxy *Proxy, timeoutS float64, started time.Time) (string, string)

// archiveLookupVar — the module-level seam.
var archiveLookupVar archiveFn = archiveLookup

// Variant — one honest variant of a dead URL (label + URL).
type Variant struct {
	Label string
	URL   string
}

// IsBareHostname — true for a plain DNS host ("example.org"): no
// userinfo, no port, not an IP, not localhost. Only these have a
// meaningful www variant.
func IsBareHostname(netloc string) bool {
	nl := strings.TrimSpace(netloc)
	if nl == "" || strings.Contains(nl, "@") || strings.Contains(nl, ":") {
		return false
	}
	if strings.EqualFold(nl, "localhost") || !strings.Contains(nl, ".") {
		return false
	}
	if ip := net.ParseIP(nl); ip != nil {
		return false // an IP has no www form
	}
	return true
}

// URLVariants — up to two honest variants of a dead URL: the
// trailing-slash toggle and the www toggle (the two moves that rescue
// the common "moved without a redirect" class). https stays https
// (never a scheme downgrade); an IP or localhost host has no www form.
func URLVariants(raw string) []Variant {
	u, err := url.Parse(strings.TrimSpace(raw))
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") {
		return nil
	}
	var out []Variant
	path := u.EscapedPath()
	var variantPath, label string
	if strings.HasSuffix(path, "/") && len(path) > 1 {
		variantPath = path[:len(path)-1]
		label = "no-trailing-slash"
	} else {
		variantPath = path + "/"
		label = "trailing-slash"
	}
	out = append(out, Variant{label, rebuild(u, variantPath)})
	nl := u.Host
	// Python's netloc keeps the userinfo; a userinfo'd URL has no www
	// variant (Go's u.Host strips it — check u.User instead).
	if u.User == nil {
		switch {
		case strings.HasPrefix(strings.ToLower(nl), "www."):
			out = append(out, Variant{"non-www", rebuildHost(u, nl[4:], path)})
		case IsBareHostname(nl):
			out = append(out, Variant{"www", rebuildHost(u, "www."+nl, path)})
		}
	}
	filtered := out[:0]
	for _, v := range out {
		if v.URL != "" && v.URL != raw {
			filtered = append(filtered, v)
		}
	}
	return filtered
}

// rebuild — urlunsplit semantics with a replaced path.
func rebuild(u *url.URL, path string) string {
	clone := *u
	clone.Path = path
	clone.RawPath = ""
	return clone.String()
}

// rebuildHost — urlunsplit semantics with a replaced netloc.
func rebuildHost(u *url.URL, host, path string) string {
	clone := *u
	clone.Host = host
	clone.Path = path
	clone.RawPath = ""
	return clone.String()
}

// DomainOfURL — the ladder's politeness key.
func waitPolite(limiter RateLimiter, urlOrDomain string) {
	if limiter == nil {
		return
	}
	d := DomainOf(urlOrDomain)
	if d == "" {
		d = urlOrDomain
	}
	limiter.Wait(d)
}

// RateLimitPenalty — the polite circuit breaker's price: when any leg
// of the final failure answered 429/503, the domain's next hit waits
// the site's Retry-After (or the 60 s default), capped at five
// minutes. nil when no rate-limit answer was seen.
func RateLimitPenalty(results ...*FetchResult) *float64 {
	for _, res := range results {
		if res != nil && res.HTTPStatus != nil &&
			(*res.HTTPStatus == 429 || *res.HTTPStatus == 503) {
			penalty := RatePenaltyDefaultS
			if res.RetryAfterS != nil {
				penalty = *res.RetryAfterS
			}
			if penalty > RatePenaltyCapS {
				penalty = RatePenaltyCapS
			}
			return &penalty
		}
	}
	return nil
}

// CombinedCategory — the most informative leg's failure class wins (a
// 404 truth beats the 403 wall that hid it; a paywall beats a timeout).
func CombinedCategory(results ...*FetchResult) string {
	cats := map[string]bool{}
	for _, res := range results {
		if res != nil && res.Category != "" {
			cats[res.Category] = true
		}
	}
	for _, cat := range CategoryPrecedence {
		if cats[cat] {
			return cat
		}
	}
	return ""
}

// archiveLookup — the Wayback Machine's availability answer for a dead
// page: ("no-snapshot", "") when nothing is archived, ("unreachable",
// "") when the archive itself could not be asked, or (snapshotURL,
// timestamp) when a capture exists. One polite probe, primary route,
// 64 KB cap.
func archiveLookup(rawURL string, proxy *Proxy, timeoutS float64, started time.Time) (string, string) {
	probeURL := ArchiveAPI + "?url=" + pyQuote(rawURL)
	probeTimeout := timeoutS
	if probeTimeout > 15.0 {
		probeTimeout = 15.0
	}
	res, err := func() (*FetchResult, error) {
		defer func() {
			if r := recover(); r != nil {
				// politeness never breaks the fetch
			}
		}()
		return fetchOnceVar(probeURL, probeTimeout, ArchiveMaxBytes,
			UserAgent, proxy, started), nil
	}()
	if err != nil || res == nil || !res.OK() {
		return "unreachable", ""
	}
	type closestShape struct {
		Available bool   `json:"available"`
		URL       string `json:"url"`
		Timestamp string `json:"timestamp"`
	}
	var data struct {
		ArchivedSnapshots struct {
			Closest *closestShape `json:"closest"`
		} `json:"archived_snapshots"`
	}
	body := strings.TrimSpace(res.Text)
	if body == "" {
		body = "{}"
	}
	if err := jsonUnmarshal(body, &data); err != nil {
		return "no-snapshot", ""
	}
	if c := data.ArchivedSnapshots.Closest; c != nil && c.Available && c.URL != "" {
		return c.URL, c.Timestamp
	}
	return "no-snapshot", ""
}

// deadPageRescue — a 404/410 verdict climbs its rescue ladder before
// it is final: URL variants (trailing slash, www — the "moved without
// a redirect" class), then the Wayback Machine's archived copy. A
// rescue is a REAL result with the story in the reason. Loopback
// NEVER gets the ladder (no door escalates off-machine).
func deadPageRescue(rawURL string, dead *FetchResult, timeoutS float64,
	maxBytes int, userAgent string, proxy *Proxy, limiter RateLimiter,
	archiveFallback bool, started time.Time) (*FetchResult, []string, string) {
	if IsLoopbackURL(rawURL) {
		return nil, nil, "skipped"
	}
	var tried []string
	for _, variant := range URLVariants(rawURL) {
		waitPolite(limiter, variant.URL)
		res := fetchOnceVar(variant.URL, timeoutS, maxBytes, userAgent, proxy, started)
		tried = append(tried, variant.Label)
		if res.OK() {
			res.Reason = "the original URL answered HTTP " +
				strconv.Itoa(deref(dead.HTTPStatus)) + " — the " +
				variant.Label + " variant answered; rescued"
			return res, tried, "skipped"
		}
	}
	archiveVerdict := "skipped"
	if archiveFallback {
		waitPolite(limiter, "archive.org")
		snapshotURL, ts := archiveLookupVar(rawURL, proxy, timeoutS, started)
		if snapshotURL != "no-snapshot" && snapshotURL != "unreachable" {
			archiveVerdict = snapshotURL
			waitPolite(limiter, "web.archive.org")
			page := fetchOnceVar(snapshotURL, timeoutS, maxBytes, userAgent, proxy, started)
			if page.OK() {
				// An archived copy is real content, but the LIVE page
				// is gone: partial + category "archived" — the pipeline
				// files it in _review with the story.
				page.Status = "partial"
				page.Category = "archived"
				story := "archived copy — the live page answered HTTP " +
					strconv.Itoa(deref(dead.HTTPStatus)) +
					"; the Wayback Machine's capture"
				if ts != "" {
					story += " (" + ts + ")"
				}
				story += " served this"
				page.Reason = story
				return page, tried, "served"
			}
		} else {
			archiveVerdict = snapshotURL
		}
	}
	return nil, tried, archiveVerdict
}

// finalVerdict — the rungs every final failure answer shares: the DNS
// verdict (when the DIRECT line could not resolve the name,
// DNS-over-HTTPS names whose fault it is), the polite circuit
// breaker's price, and the origin-down courtesy (a 521-524 leg names
// the Wayback's last good copy). Mutates and returns result.
func finalVerdict(result *FetchResult, directReason, rawURL string,
	proxy *Proxy, limiter RateLimiter, domain string, dohProbe bool,
	legs []*FetchResult, archiveFallback bool, started time.Time) *FetchResult {
	suffix, dead := dnsVerdictSuffix(directReason, rawURL, proxy, dohProbe)
	if suffix != "" {
		result.Reason += suffix
		if dead {
			result.Category = "dead"
		}
	}
	if penalty := RateLimitPenalty(legs...); penalty != nil && limiter != nil {
		limiter.Penalize(domain, *penalty)
	}
	if archiveFallback && !IsLoopbackURL(rawURL) {
		for _, leg := range legs {
			if leg == nil || leg.HTTPStatus == nil {
				continue
			}
			if _, ok := CloudflareOriginStatuses[*leg.HTTPStatus]; ok {
				waitPolite(limiter, "archive.org")
				snapshot, ts := archiveLookupVar(rawURL, proxy, OriginArchiveTimeoutS, started)
				if snapshot != "no-snapshot" && snapshot != "unreachable" {
					courtesy := " | wayback: an archived copy exists — " + snapshot
					if ts != "" {
						courtesy += " (captured " + ts + ")"
					}
					result.Reason += courtesy
				}
				break
			}
		}
	}
	return result
}

func deref(p *int) int {
	if p == nil {
		return 0
	}
	return *p
}
