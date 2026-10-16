// ladder_test.go — the ladder's seam-scripted batteries, ported from
// test_ladder.py (TestDnsVerdict, TestDeadPageRescue, TestRatePenalty's
// fetch-level legs, TestOriginDownCourtesy) and test_bothdoors.py
// (TestRefusalAlternates, TestResourceAnswers, TestOldLaws): hand-built
// FetchResults and patched seams, zero sockets.
package webfetch

import (
	"reflect"
	"strings"
	"testing"
	"time"
)

var testProxy = &Proxy{Type: "socks5", Host: "p.example.net", Port: 10808}

var ownerTLSLine = "proxy: connection: [SSL: SSLV3_ALERT_HANDSHAKE_FAILURE] " +
	"sslv3 alert handshake failure (_ssl.c:1010) | direct: connection: " +
	"[Errno 11001] getaddrinfo failed"

// ---------------------------------------------------------------------------
// the DNS verdict (patched DoH probe — zero sockets)
// ---------------------------------------------------------------------------

func armDoH(t *testing.T, verdict string) *[][2]any {
	t.Helper()
	calls := &[][2]any{}
	dohProbeOnceVar = func(host string, proxy *Proxy) string {
		*calls = append(*calls, [2]any{host, proxy})
		return verdict
	}
	return calls
}

func dnsFail(url string, timeoutS float64, maxBytes int, userAgent string, proxy *Proxy, started time.Time) *FetchResult {
	return res("connection: [Errno 11001] getaddrinfo failed", nil, "retry_later", nil)
}

func TestDNSVerdictNxdomainMarksDead(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	calls := armDoH(t, "nxdomain")
	fetchOnceVar = dnsFail
	r := FetchURL("http://gone.example.net/x", &Options{TimeoutS: 2})
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.Category != "dead" {
		t.Fatalf("category = %q", r.Category)
	}
	if !strings.Contains(r.Reason, "NXDOMAIN even via DNS-over-HTTPS") {
		t.Fatalf("reason = %q", r.Reason)
	}
	wantCall := [2]any{"gone.example.net", (*Proxy)(nil)}
	if len(*calls) != 1 || !reflect.DeepEqual((*calls)[0], wantCall) {
		t.Fatalf("calls = %v", *calls)
	}
}

func TestDNSVerdictExistsNamesPoisonedResolver(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	calls := armDoH(t, "exists")
	fetchOnceVar = dnsFail
	r := FetchURL("http://x.example.net/x", &Options{TimeoutS: 2})
	if r.OK() || r.Category != "retry_later" {
		t.Fatalf("status=%q category=%q", r.Status, r.Category)
	}
	if !strings.Contains(r.Reason, "local resolver is lying") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*calls) != 1 {
		t.Fatalf("calls = %v", *calls)
	}
}

func TestDNSVerdictUnknownNeverAccuses(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armDoH(t, "unknown")
	fetchOnceVar = dnsFail
	r := FetchURL("http://y.example.net/x", &Options{TimeoutS: 2})
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, "verdict unknown") || r.Category != "retry_later" {
		t.Fatalf("reason=%q category=%q", r.Reason, r.Category)
	}
}

func TestDNSVerdictIsCachedPerHost(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	calls := armDoH(t, "nxdomain")
	fetchOnceVar = dnsFail
	FetchURL("http://z.example.net/x", &Options{TimeoutS: 2})
	FetchURL("http://z.example.net/x", &Options{TimeoutS: 2})
	if len(*calls) != 1 {
		t.Fatalf("probe ran %d times (must be cached)", len(*calls))
	}
}

func TestDNSVerdictOptOutSkipsProbe(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	calls := armDoH(t, "nxdomain")
	fetchOnceVar = dnsFail
	r := FetchURL("http://w.example.net/x", &Options{TimeoutS: 2, NoDoHProbe: true})
	if r.OK() {
		t.Fatal("must fail")
	}
	if len(*calls) != 0 {
		t.Fatalf("probe ran: %v", *calls)
	}
	if strings.Contains(r.Reason, "dns:") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestDNSVerdictNonDNSFailureNeverProbes(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	calls := armDoH(t, "nxdomain")
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("timeout", nil, "retry_later", nil)
	}
	r := FetchURL("http://v.example.net/x", &Options{TimeoutS: 2})
	if len(*calls) != 0 {
		t.Fatalf("probe ran: %v", *calls)
	}
	if r.Reason != "timeout" {
		t.Fatalf("reason = %q", r.Reason)
	}
}

// ---------------------------------------------------------------------------
// the dead-page rescue ladder (patched fetchOnce + archiveLookup)
// ---------------------------------------------------------------------------

func armArchive(t *testing.T, verdict, timestamp string) *[]string {
	t.Helper()
	calls := &[]string{}
	archiveLookupVar = func(url string, proxy *Proxy, timeoutS float64, started time.Time) (string, string) {
		*calls = append(*calls, url)
		return verdict, timestamp
	}
	return calls
}

func TestDeadPageRescueWWWVariantRescues(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	pages := map[string]*FetchResult{
		"https://moved.example.org/page":  res("HTTP 404 — the page is gone", intPtr(404), "dead", nil),
		"https://moved.example.org/page/": res("HTTP 404 — the page is gone", intPtr(404), "dead", nil),
		"https://www.moved.example.org/page": okRes("https://www.moved.example.org/page",
			[]byte("<html><body>the page</body></html>")),
	}
	archive := armArchive(t, "no-snapshot", "")
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return pages[url]
	}
	r := FetchURL("https://moved.example.org/page", &Options{TimeoutS: 2})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "full" {
		t.Fatalf("status = %q", r.Status)
	}
	if !strings.Contains(r.Reason, "www variant answered; rescued") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if r.FinalURL != "https://www.moved.example.org/page" {
		t.Fatalf("final = %q", r.FinalURL)
	}
	if len(*archive) != 0 {
		t.Fatalf("the archive door was never asked: %v", *archive)
	}
}

func TestDeadPageRescueNothingRescuedIsDeadWithStory(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armArchive(t, "no-snapshot", "")
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 404 — the page is gone", intPtr(404), "dead", nil)
	}
	r := FetchURL("https://gone.example.org/dead", &Options{TimeoutS: 2})
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.Category != "dead" {
		t.Fatalf("category = %q", r.Category)
	}
	if !strings.Contains(r.Reason, "variants tried: trailing-slash, www") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if !strings.Contains(r.Reason, "no archived copy exists") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestDeadPageRescueArchivedSnapshotIsPartialWithStory(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	snap := "https://web.archive.org/web/20150601/https://past.example.org/a"
	armArchive(t, snap, "20150601")
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		if strings.Contains(url, "web.archive.org") {
			return okRes(url, []byte("<html><head><title>The Past</title></head></html>"))
		}
		return res("HTTP 410 — the page was removed", intPtr(410), "dead", nil)
	}
	r := FetchURL("https://past.example.org/a", &Options{TimeoutS: 2})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "partial" || r.Category != "archived" {
		t.Fatalf("status=%q category=%q", r.Status, r.Category)
	}
	for _, want := range []string{"archived copy", "20150601", "HTTP 410"} {
		if !strings.Contains(r.Reason, want) {
			t.Fatalf("reason = %q (missing %q)", r.Reason, want)
		}
	}
	if !strings.Contains(r.Text, "The Past") {
		t.Fatalf("text = %q", r.Text)
	}
}

func TestDeadPageRescueArchiveOptOut(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	archive := armArchive(t, "no-snapshot", "")
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 404 — the page is gone", intPtr(404), "dead", nil)
	}
	r := FetchURL("https://opt.example.org/x", &Options{TimeoutS: 2, NoArchive: true})
	if r.OK() || r.Category != "dead" {
		t.Fatalf("status=%q category=%q", r.Status, r.Category)
	}
	if len(*archive) != 0 {
		t.Fatalf("archive asked: %v", *archive)
	}
}

func TestDeadPageRescueLoopbackNeverGetsLadder(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	calls := &[]string{}
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		*calls = append(*calls, url)
		return res("HTTP 404 — the page is gone", intPtr(404), "dead", nil)
	}
	r := FetchURL("http://127.0.0.1:9/x", &Options{TimeoutS: 2})
	if r.OK() || r.Category != "dead" {
		t.Fatalf("status=%q category=%q", r.Status, r.Category)
	}
	if !reflect.DeepEqual(*calls, []string{"http://127.0.0.1:9/x"}) {
		t.Fatalf("calls = %v", *calls)
	}
}

// ---------------------------------------------------------------------------
// the rate penalties at the fetch level (the circuit breaker)
// ---------------------------------------------------------------------------

func TestRatePenaltyBothRoutesPayRetryAfter(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false) // the third door never dials here
	lim := &fakeLimiter{}
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 429 — rate limited (Retry-After: 120s — the site asked for a pause)",
			intPtr(429), "retry_later", floatPtr(120))
	}
	r := FetchURL("http://limited.example.net/x", &Options{TimeoutS: 2, Proxy: testProxy, RateLimiter: lim})
	if r.OK() {
		t.Fatal("must fail")
	}
	if !strings.Contains(r.Reason, "proxy: HTTP 429") || !strings.Contains(r.Reason, "direct: HTTP 429") {
		t.Fatalf("reason = %q", r.Reason)
	}
	want := []penaltyRecord{{"limited.example.net", 120}}
	if !reflect.DeepEqual(lim.penalties, want) {
		t.Fatalf("penalties = %v", lim.penalties)
	}
}

func TestRatePenaltyBare429PaysCappedDefault(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	lim := &fakeLimiter{}
	fetchOnceVar = func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 429 — rate limited", intPtr(429), "retry_later", nil)
	}
	r := FetchURL("http://bare.example.net/x", &Options{TimeoutS: 2, RateLimiter: lim})
	if r.OK() {
		t.Fatal("must fail")
	}
	want := []penaltyRecord{{"bare.example.net", RatePenaltyDefaultS}}
	if !reflect.DeepEqual(lim.penalties, want) {
		t.Fatalf("penalties = %v", lim.penalties)
	}
}

// ---------------------------------------------------------------------------
// the origin-down Wayback courtesy (v0.47.0)
// ---------------------------------------------------------------------------

func verdict52x(legs ...*FetchResult) *FetchResult {
	result := &FetchResult{
		URL:      "https://origin.example.net/x",
		FinalURL: "https://origin.example.net/x",
		Status:   "failed",
		Reason: "proxy: HTTP 403 — forbidden | direct: HTTP 523 — " +
			"Cloudflare: the site's own server origin is " +
			"unreachable (site-side, retry later — not the " +
			"fetcher's fault)",
		Category: "blocked_bot",
	}
	return finalVerdict(result, "", "https://origin.example.net/x", nil, nil,
		"origin.example.net", false, legs, true, time.Now())
}

func TestOriginDownCourtesy52xNamesSnapshot(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	snap := "https://web.archive.org/web/20230101120000/https://origin.example.net/x"
	archive := armArchive(t, snap, "20230101120000")
	r := verdict52x(res("HTTP 403", intPtr(403), "", nil), res("HTTP 523", intPtr(523), "", nil))
	want := "wayback: an archived copy exists — " + snap
	if !strings.Contains(r.Reason, want) {
		t.Fatalf("reason = %q (missing %q)", r.Reason, want)
	}
	if !strings.Contains(r.Reason, "(captured 20230101120000)") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if !reflect.DeepEqual(*archive, []string{"https://origin.example.net/x"}) {
		t.Fatalf("archive calls = %v", *archive)
	}
	if r.Category != "blocked_bot" {
		t.Fatalf("the verdict class is untouched: %q", r.Category)
	}
}

func TestOriginDownCourtesyNoSnapshotNoSuffix(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	archive := armArchive(t, "no-snapshot", "")
	r := verdict52x(res("HTTP 403", intPtr(403), "", nil), res("HTTP 523", intPtr(523), "", nil))
	if len(*archive) != 1 {
		t.Fatalf("archive calls = %v", *archive)
	}
	if strings.Contains(r.Reason, "wayback") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestOriginDownCourtesyUnreachableNoSuffix(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armArchive(t, "unreachable", "")
	r := verdict52x(res("HTTP 403", intPtr(403), "", nil), res("HTTP 523", intPtr(523), "", nil))
	if strings.Contains(r.Reason, "wayback") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestOriginDownCourtesyOptOutNeverProbes(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	archive := armArchive(t, "no-snapshot", "")
	result := &FetchResult{URL: "https://origin.example.net/x", Status: "failed",
		Reason: "HTTP 523", Category: "retry_later"}
	finalVerdict(result, "", "https://origin.example.net/x", nil, nil,
		"origin.example.net", false, []*FetchResult{res("HTTP 523", intPtr(523), "", nil)},
		false, time.Now())
	if len(*archive) != 0 {
		t.Fatalf("archive calls = %v", *archive)
	}
}

func TestOriginDownCourtesyNo52xNeverProbes(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	archive := armArchive(t, "no-snapshot", "")
	r := verdict52x(res("HTTP 403", intPtr(403), "", nil), res("HTTP 429", intPtr(429), "", nil))
	if len(*archive) != 0 {
		t.Fatalf("archive calls = %v", *archive)
	}
	if strings.Contains(r.Reason, "wayback") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestOriginDownCourtesyLoopbackNeverProbes(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	archive := armArchive(t, "no-snapshot", "")
	url := "http://127.0.0.1:9/x"
	result := &FetchResult{URL: url, Status: "failed", Reason: "HTTP 523", Category: "retry_later"}
	finalVerdict(result, "", url, nil, nil, "127.0.0.1:9", false,
		[]*FetchResult{resAt("HTTP 523", intPtr(523), url)}, true, time.Now())
	if len(*archive) != 0 {
		t.Fatalf("archive calls = %v", *archive)
	}
}

func TestOriginDownCourtesySingle52xLeg(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	snap := "https://web.archive.org/web/2023/https://solo.example.org"
	archive := armArchive(t, snap, "")
	url := "https://solo.example.org/y"
	result := &FetchResult{URL: url, Status: "failed", Reason: "HTTP 523 — Cloudflare", Category: "retry_later"}
	r := finalVerdict(result, "", url, nil, nil, "solo.example.org", false,
		[]*FetchResult{resAt("HTTP 523", intPtr(523), url)}, true, time.Now())
	if len(*archive) != 1 {
		t.Fatalf("archive calls = %v", *archive)
	}
	if !strings.Contains(r.Reason, snap) {
		t.Fatalf("reason = %q", r.Reason)
	}
}

// ---------------------------------------------------------------------------
// the both-doors alternation table (test_bothdoors.py)
// ---------------------------------------------------------------------------

var bothProxy = &Proxy{Type: "socks5", Host: "127.0.0.1", Port: 1}

func fullRoute(url string) *FetchResult {
	return &FetchResult{URL: url, FinalURL: url, Status: "full",
		HTTPStatus: intPtr(200), Body: []byte("<html>walled out</html>"),
		Text: "<html>walled out</html>"}
}

func TestBothDoorsRefusalFamilyAlternates(t *testing.T) {
	cases := []struct {
		name   string
		reason string
		code   int
	}{
		{"403", "HTTP 403 — bot defense (server: cloudflare)", 403},
		{"405", "HTTP 405", 405},
		{"429", "HTTP 429", 429},
		{"451", "HTTP 451", 451},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			state := saveDoorState()
			defer state.restore()
			armAvailable(false)
			calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
				if p != nil {
					return res(tc.reason, &tc.code, "", nil)
				}
				return fullRoute("https://w.test/" + tc.name)
			})
			r := FetchURL("https://w.test/"+tc.name, &Options{TimeoutS: 2, Proxy: bothProxy})
			if !r.OK() {
				t.Fatalf("not ok: %s", r.Reason)
			}
			if r.Status != "full" || r.HTTPStatus == nil || *r.HTTPStatus != 200 {
				t.Fatalf("status=%q http=%v", r.Status, r.HTTPStatus)
			}
			if !strings.Contains(r.Reason, "direct fallback") || !strings.Contains(r.Reason, tc.reason) {
				t.Fatalf("reason = %q", r.Reason)
			}
			if len(*calls) != 2 || (*calls)[0] == nil || (*calls)[1] != nil {
				t.Fatalf("routes = %v (proxy then direct)", *calls)
			}
		})
	}
}

func TestBothDoorsTimeoutStillAlternates(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		if p != nil {
			return res("timeout", nil, "retry_later", nil)
		}
		return fullRoute("https://w.test/f")
	})
	r := FetchURL("https://w.test/f", &Options{TimeoutS: 2, Proxy: bothProxy})
	if !r.OK() || !strings.Contains(r.Reason, "timeout") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*calls) != 2 {
		t.Fatalf("routes = %v", *calls)
	}
}

func TestBothDoorsWalledOnBothReportsBoth(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		if p != nil {
			return res("HTTP 403 — bot defense (server: cloudflare)", intPtr(403), "", nil)
		}
		return res("HTTP 405", intPtr(405), "", nil)
	})
	r := FetchURL("https://w.test/e", &Options{TimeoutS: 2, Proxy: bothProxy})
	if r.OK() {
		t.Fatal("must fail")
	}
	if len(*calls) != 2 {
		t.Fatalf("routes = %v", *calls)
	}
	for _, want := range []string{"proxy: HTTP 403", "bot defense", "direct: HTTP 405"} {
		if !strings.Contains(r.Reason, want) {
			t.Fatalf("reason = %q (missing %q)", r.Reason, want)
		}
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 405 {
		t.Fatalf("http = %v (the second door's answer)", r.HTTPStatus)
	}
}

func TestBothDoorsConnectionDeadLineUnchanged(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		if p != nil {
			return res("timeout", nil, "", nil)
		}
		return res("connection: timed out", nil, "", nil)
	})
	r := FetchURL("https://w.test/g", &Options{TimeoutS: 2, Proxy: bothProxy})
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.Reason != "proxy: timeout | direct: connection: timed out" {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestBothDoorsFirstDoorSuccessNeverDialsSecond(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return fullRoute("https://w.test/h")
	})
	r := FetchURL("https://w.test/h", &Options{TimeoutS: 2, Proxy: bothProxy})
	if !r.OK() || r.Reason != "" {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(*calls) != 1 {
		t.Fatalf("routes = %v", *calls)
	}
}

func TestBothDoors404TruthSingleRouteWithLadder(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	armArchive(t, "no-snapshot", "")
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 404", intPtr(404), "", nil)
	})
	r := FetchURL("https://gone.test/404", &Options{TimeoutS: 2, Proxy: bothProxy})
	if r.OK() {
		t.Fatal("must fail")
	}
	// every ask rode the PRIMARY route: the original + two variants
	if len(*calls) != 3 {
		t.Fatalf("routes = %v (original + 2 variants)", *calls)
	}
	for _, c := range *calls {
		if c != bothProxy {
			t.Fatalf("unexpected route: %v", c)
		}
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 404 || r.Category != "dead" {
		t.Fatalf("http=%v category=%q", r.HTTPStatus, r.Category)
	}
	if !strings.Contains(r.Reason, "variants tried") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestBothDoorsResourceTruthsSingleAttempt(t *testing.T) {
	for _, tc := range []struct {
		code   int
		reason string
	}{
		{401, "HTTP 401"}, {500, "HTTP 500"},
	} {
		t.Run(tc.reason, func(t *testing.T) {
			state := saveDoorState()
			defer state.restore()
			armAvailable(false)
			calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
				return res(tc.reason, &tc.code, "", nil)
			})
			r := FetchURL("https://gone.test/x", &Options{TimeoutS: 2, Proxy: bothProxy})
			if r.OK() {
				t.Fatal("must fail")
			}
			if len(*calls) != 1 {
				t.Fatalf("routes = %v (the site ANSWERED — truth)", *calls)
			}
			if r.HTTPStatus == nil || *r.HTTPStatus != tc.code || r.Reason != tc.reason {
				t.Fatalf("http=%v reason=%q", r.HTTPStatus, r.Reason)
			}
		})
	}
}

func TestBothDoorsNoProxySingleAttempt(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	})
	r := FetchURL("https://walled.test/x", &Options{TimeoutS: 2})
	if r.OK() {
		t.Fatal("must fail")
	}
	if len(*calls) != 1 || (*calls)[0] != nil {
		t.Fatalf("routes = %v", *calls)
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 403 {
		t.Fatalf("http = %v", r.HTTPStatus)
	}
}

func TestBothDoorsDirectOptOutSingleAttempt(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	})
	r := FetchURL("https://walled.test/y", &Options{TimeoutS: 2, Proxy: bothProxy, NoDirectFallback: true})
	if r.OK() {
		t.Fatal("must fail")
	}
	if len(*calls) != 1 || (*calls)[0] == nil {
		t.Fatalf("routes = %v", *calls)
	}
}

func TestBothDoorsLoopbackNeverAlternates(t *testing.T) {
	state := saveDoorState()
	defer state.restore()
	armAvailable(false)
	calls := scriptFetchOnce(func(url string, timeoutS float64, maxBytes int, ua string, p *Proxy, s time.Time) *FetchResult {
		return res("HTTP 403", intPtr(403), "", nil)
	})
	r := FetchURL("http://127.0.0.1:9/z", &Options{TimeoutS: 2, Proxy: bothProxy})
	if r.OK() {
		t.Fatal("must fail")
	}
	if len(*calls) != 1 {
		t.Fatalf("routes = %v (one attempt, forced direct)", *calls)
	}
}

func TestBothDoorsRefusalFamilyIsExactlyFour(t *testing.T) {
	if !reflect.DeepEqual(RouteRefusalStatuses, []int{403, 405, 429, 451}) {
		t.Fatalf("ROUTE_REFUSAL_STATUSES = %v", RouteRefusalStatuses)
	}
}

// resAt builds a failed result at a specific URL.
func resAt(reason string, httpStatus *int, url string) *FetchResult {
	r := res(reason, httpStatus, "", nil)
	r.URL = url
	r.FinalURL = url
	return r
}
