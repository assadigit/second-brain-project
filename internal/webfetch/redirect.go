// redirect.go — v0.46.0 THE LADDER's redirect truth: redirects are
// followed by OUR loop, not the client's (CheckRedirect
// ErrUseLastResponse + manual hops). Hop cookies ride along (the
// "lost a cookie during it" login-loop class), 308 is followed, loops
// are named, a redirect with no Location is an honest broken-redirect
// failure — never a bare "HTTP 307".
package webfetch

import (
	"fmt"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strings"
	"sync"
)

// RedirectBrokenError — a redirect answer that cannot be followed (no
// Location header, a non-http target, a loop, or a chain past the
// budget). Carried to the verdict as category "redirect_broken" with
// the precise story.
type RedirectBrokenError struct {
	Code   int
	Detail string
}

func (e *RedirectBrokenError) Error() string {
	return fmt.Sprintf("HTTP %d — broken redirect: %s", e.Code, e.Detail)
}

// The persistent cookie jar (the browser-session law: a clearance
// cookie earned on one domain keeps working on the next visit, and
// cookies SET ON A REDIRECT HOP reach the next hop). Process-lifetime,
// domain-scoped by the jar itself, guarded by a lock.
var (
	cookieJar  = mustJar()
	cookieLock sync.Mutex
)

func mustJar() *cookiejar.Jar {
	jar, err := cookiejar.New(nil)
	if err != nil {
		panic(err)
	}
	return jar
}

// ResetCookieJar — the test seam (the suites clear the jar per probe).
func ResetCookieJar() {
	cookieLock.Lock()
	defer cookieLock.Unlock()
	cookieJar = mustJar()
}

// absorbCookies extracts Set-Cookie from one response into the shared
// jar. Bookkeeping never kills a fetch.
func absorbCookies(u *url.URL, resp *http.Response) {
	if u == nil || resp == nil {
		return
	}
	defer func() { _ = recover() }()
	cookieLock.Lock()
	defer cookieLock.Unlock()
	cookieJar.SetCookies(u, resp.Cookies())
}

// attachCookies adds the jar's Cookie header for this request's domain.
func attachCookies(req *http.Request, u *url.URL) {
	if req == nil || u == nil {
		return
	}
	defer func() { _ = recover() }()
	cookieLock.Lock()
	defer cookieLock.Unlock()
	for _, c := range cookieJar.Cookies(u) {
		req.AddCookie(c)
	}
}

// openFollowing opens url following redirect hops OURSELVES. Carries
// the cookie jar across hops, follows 301/302/303/307/308 (308 needs
// it on urllib < 3.11 — here it is just ours), caps hops at
// MaxRedirects, detects loops, and returns a precise
// *RedirectBrokenError when the redirect answer is broken.
//
// Returns (resp, finalURL, err) — the FINAL non-redirect response (the
// caller owns closing it). Non-redirect HTTP error statuses come back
// as responses (the caller's verdict table names them); everything
// else errors exactly as the client did.
func openFollowing(client *http.Client, rawURL string, headers []HeaderPair) (*http.Response, string, error) {
	current := rawURL
	seen := map[string]bool{rawURL: true}
	hops := 0
	for {
		u, err := url.Parse(current)
		if err != nil {
			return nil, "", err
		}
		req, err := http.NewRequest(http.MethodGet, current, nil)
		if err != nil {
			return nil, "", err
		}
		for _, p := range headers {
			req.Header.Set(p.Key, p.Value)
		}
		attachCookies(req, u)
		resp, err := client.Do(req)
		if err != nil {
			// Cookies from error responses are absorbed by the client
			// error path too (urllib's HTTPError carries headers; Go's
			// Do-error may carry a response with Set-Cookie).
			if resp != nil {
				absorbCookies(u, resp)
				_ = resp.Body.Close()
			}
			return nil, "", err
		}
		absorbCookies(u, resp)
		if isRedirectStatus(resp.StatusCode) {
			target := RedirectTarget(resp.Header)
			if target == "" {
				_ = resp.Body.Close()
				return nil, "", &RedirectBrokenError{
					Code: resp.StatusCode,
					Detail: "no Location header in the answer — the " +
						"server's redirect is broken (site-side)",
				}
			}
			nextURL := resolveRef(current, target)
			scheme := strings.ToLower(uScheme(nextURL))
			if scheme != "http" && scheme != "https" {
				_ = resp.Body.Close()
				detail := fmt.Sprintf("redirect to a non-http target "+
					"(%s: %s)", orDefault(scheme, "no scheme"), clip(target, 80))
				return nil, "", &RedirectBrokenError{Code: resp.StatusCode, Detail: detail}
			}
			if seen[nextURL] {
				_ = resp.Body.Close()
				detail := fmt.Sprintf("redirect loop (the chain revisits "+
					"%s after %d hop(s) — cookies are carried, so this is "+
					"the server's own loop)", clip(nextURL, 120), hops)
				return nil, "", &RedirectBrokenError{Code: resp.StatusCode, Detail: detail}
			}
			if hops >= MaxRedirects {
				_ = resp.Body.Close()
				return nil, "", &RedirectBrokenError{
					Code: resp.StatusCode,
					Detail: fmt.Sprintf("more than %d redirects "+
						"(a tracker chute or a wandering chain)", MaxRedirects),
				}
			}
			hops++
			seen[nextURL] = true
			headers = HopHeaders(headers, current, nextURL)
			_ = resp.Body.Close()
			current = nextURL
			continue
		}
		return resp, current, nil
	}
}

// resolveRef — urljoin semantics (an absolute target wins).
func resolveRef(base, target string) string {
	b, err := url.Parse(base)
	if err != nil {
		return target
	}
	t, err := url.Parse(target)
	if err != nil {
		return target
	}
	return b.ResolveReference(t).String()
}

// uScheme — the scheme of a URL string ("" when absent).
func uScheme(raw string) string {
	if i := strings.Index(raw, "://"); i >= 0 {
		return raw[:i]
	}
	if i := strings.Index(raw, ":"); i >= 0 && !strings.ContainsAny(raw[:i], "/?#") {
		return raw[:i]
	}
	return ""
}

func orDefault(s, def string) string {
	if s == "" {
		return def
	}
	return s
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
