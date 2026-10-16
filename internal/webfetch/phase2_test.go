// phase2_test.go — the web_fetch legs of test_phase2.py's TestWebFetch:
// the local-server battery (full / redirect / 404 / timeout / size cap /
// pdf / latin1 / redirect loop / bad scheme / the browser header set /
// gzip / deflate / the gzip bomb / the 403 bot-defense line).
package webfetch

import (
	"strings"
	"testing"
	"time"
)

func TestPhase2FullFetch(t *testing.T) {
	srv := newPhase2Server(t)
	r := FetchURL(srv.root()+"/article", nil)
	if r.Status != "full" || !r.OK() {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 200 {
		t.Fatalf("http=%v", r.HTTPStatus)
	}
	if !strings.Contains(r.Text, "Practical Typography Guide") {
		t.Fatalf("text = %q", r.Text)
	}
}

func TestPhase2FollowsRedirect(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/redirect", nil)
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if r.URL == r.FinalURL || !strings.Contains(r.FinalURL, "/article") {
		t.Fatalf("url=%q final=%q", r.URL, r.FinalURL)
	}
}

func TestPhase2_404IsFailedNotPanic(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/nope", nil)
	if r.Status != "failed" || r.OK() {
		t.Fatalf("status=%q", r.Status)
	}
	if !strings.Contains(r.Reason, "404") {
		t.Fatalf("reason = %q", r.Reason)
	}
	// loopback never gets the rescue ladder (the v0.15.1 rule — the
	// Python suite asserts only the honest 404 verdict here)
	if strings.Contains(r.Reason, "variants tried") {
		t.Fatalf("reason = %q (loopback must not get the ladder)", r.Reason)
	}
}

func TestPhase2TimeoutIsFailed(t *testing.T) {
	srv := newPhase2Server(t)
	t0 := time.Now()
	r := FetchURL(srv.root()+"/slow", &Options{TimeoutS: 1})
	elapsed := time.Since(t0)
	if r.Status != "failed" {
		t.Fatalf("status=%q", r.Status)
	}
	if r.Reason != "timeout" {
		t.Fatalf("reason = %q (want timeout)", r.Reason)
	}
	if elapsed >= 5*time.Second {
		t.Fatalf("elapsed = %v", elapsed)
	}
}

func TestPhase2SizeCapTruncatesToPartial(t *testing.T) {
	srv := newPhase2Server(t)
	r := FetchURL(srv.root()+"/huge", &Options{MaxBytes: 50_000})
	if r.Status != "partial" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if !strings.Contains(r.Reason, "size cap") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if len(r.Body) > 50_000 {
		t.Fatalf("body = %d bytes", len(r.Body))
	}
}

func TestPhase2PDFIsPartial(t *testing.T) {
	srv := newPhase2Server(t)
	r := FetchURL(srv.root()+"/pdf", nil)
	if r.Status != "partial" || r.Reason != "pdf" || len(r.Body) != 0 {
		t.Fatalf("status=%q reason=%q body=%d", r.Status, r.Reason, len(r.Body))
	}
}

func TestPhase2NonUTF8CharsetDecodes(t *testing.T) {
	srv := newPhase2Server(t)
	r := FetchURL(srv.root()+"/latin1", nil)
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if !strings.Contains(r.Text, "Diseño") {
		t.Fatalf("text = %q", r.Text)
	}
}

func TestPhase2RedirectLoopIsFailed(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/redirect-loop", nil)
	if r.Status != "failed" {
		t.Fatalf("status=%q", r.Status)
	}
}

func TestPhase2BadSchemeIsFailed(t *testing.T) {
	r := FetchURL("ftp://example.com/x", &Options{TimeoutS: 5})
	if r.Status != "failed" {
		t.Fatalf("status=%q", r.Status)
	}
	if r.Reason != "not an http(s) URL" || r.Category != "bad_url" {
		t.Fatalf("reason=%q category=%q", r.Reason, r.Category)
	}
}

func TestPhase2BrowserHeadersSent(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/headers", nil)
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	low := strings.ToLower(r.Text)
	for _, want := range []string{
		"user-agent: mozilla/5.0", "chrome/131",
		"accept-encoding: gzip, deflate",
		"sec-fetch-dest: document", "sec-fetch-mode: navigate",
		"accept-language: en-us,en;q=0.9", "sec-ch-ua:",
	} {
		if !strings.Contains(low, want) {
			t.Fatalf("headers echo missing %q in:\n%s", want, r.Text)
		}
	}
}

func TestPhase2CustomUADropsChromeClientHints(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/headers", &Options{UserAgent: "GitCurator/0.41 test"})
	if r.Status != "full" {
		t.Fatalf("status=%q", r.Status)
	}
	if !strings.Contains(r.Text, "User-Agent: GitCurator/0.41 test") {
		t.Fatalf("UA missing:\n%s", r.Text)
	}
	if strings.Contains(strings.ToLower(r.Text), "sec-ch-ua") {
		t.Fatalf("client hints must drop:\n%s", r.Text)
	}
	if !strings.Contains(r.Text, "Accept-Encoding: gzip, deflate") {
		t.Fatalf("rest of browser set missing:\n%s", r.Text)
	}
}

func TestPhase2GzipBodyDecoded(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/gzip", nil)
	if r.Status != "full" || !r.OK() {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 200 {
		t.Fatalf("http=%v", r.HTTPStatus)
	}
	if !strings.Contains(r.Text, "Gzip Page") || !strings.Contains(r.Text, "design tools") {
		t.Fatalf("text = %q", r.Text)
	}
}

func TestPhase2DeflateBodyDecoded(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/deflate", nil)
	if r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if !strings.Contains(r.Text, "Deflate Page") {
		t.Fatalf("text = %q", r.Text)
	}
}

func TestPhase2GzipBombCappedAtDecompression(t *testing.T) {
	srv := newPhase2Server(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/gzip-bomb", &Options{MaxBytes: 50_000})
	if r.Status != "partial" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
	if !strings.Contains(r.Reason, "size cap") || len(r.Body) > 50_000 {
		t.Fatalf("reason=%q body=%d", r.Reason, len(r.Body))
	}
	if !strings.HasPrefix(string(r.Body), "bomb ") {
		t.Fatalf("body head = %q", string(r.Body[:10]))
	}
}

func TestPhase2_403NamesTheBotDefense(t *testing.T) {
	srv := newPhase2Server(t)
	r := FetchURL(srv.root()+"/blocked", nil)
	if r.Status != "failed" {
		t.Fatalf("status=%q", r.Status)
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 403 {
		t.Fatalf("http=%v", r.HTTPStatus)
	}
	if !strings.Contains(r.Reason, "HTTP 403") ||
		!strings.Contains(r.Reason, "bot defense") ||
		!strings.Contains(r.Reason, "cloudflare") {
		t.Fatalf("reason = %q", r.Reason)
	}
	if r.Category != "blocked_bot" {
		t.Fatalf("category = %q", r.Category)
	}
}
