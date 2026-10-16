// redirect_test.go — the redirect truth (test_ladder.py's
// TestRedirectTruth): REAL local listeners, because the redirect law is
// the one law that needs a live wire to prove (hop cookies ride the
// redirect, a redirect with no Location is an honest broken-redirect
// failure, loops are named, the hop budget holds, 308 is followed).
package webfetch

import (
	"strings"
	"testing"
)

func TestRedirectTruthHopCookiesRide(t *testing.T) {
	// The 'lost a cookie during it' class: the hop sets the cookie, the
	// target serves ONLY with it — the loop must carry it across.
	srv := newLadderServer(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/hop-cookie", &Options{TimeoutS: 10})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "full" {
		t.Fatalf("status = %q", r.Status)
	}
	if !strings.Contains(r.Text, "Behind the gate") {
		t.Fatalf("text = %q", r.Text)
	}
	if r.FinalURL != srv.root()+"/hop-cookie-target" {
		t.Fatalf("final_url = %q", r.FinalURL)
	}
}

func TestRedirectTruthWithoutLocationIsHonest(t *testing.T) {
	srv := newLadderServer(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/no-location", &Options{TimeoutS: 10})
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.HTTPStatus == nil || *r.HTTPStatus != 307 {
		t.Fatalf("http_status = %v", r.HTTPStatus)
	}
	if r.Category != "redirect_broken" {
		t.Fatalf("category = %q", r.Category)
	}
	if !strings.Contains(r.Reason, "no Location header") ||
		!strings.Contains(r.Reason, "broken redirect") {
		t.Fatalf("reason = %q", r.Reason)
	}
}

func TestRedirectTruthLoopIsNamed(t *testing.T) {
	srv := newLadderServer(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/loop", &Options{TimeoutS: 10})
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.Category != "redirect_broken" || !strings.Contains(r.Reason, "redirect loop") {
		t.Fatalf("category=%q reason=%q", r.Category, r.Reason)
	}
}

func TestRedirectTruthHopBudgetHolds(t *testing.T) {
	// A wandering chain (a tracker chute) stops at MaxRedirects.
	srv := newLadderServer(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/chain/1", &Options{TimeoutS: 10})
	if r.OK() {
		t.Fatal("must fail")
	}
	if r.Category != "redirect_broken" {
		t.Fatalf("category = %q", r.Category)
	}
	want := "more than 5 redirects"
	if !strings.Contains(r.Reason, want) {
		t.Fatalf("reason = %q (want %q)", r.Reason, want)
	}
}

func TestRedirectTruth308IsFollowed(t *testing.T) {
	// urllib < 3.11 refuses 308; the house loop follows it.
	srv := newLadderServer(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/moved-308", &Options{TimeoutS: 10})
	if !r.OK() {
		t.Fatalf("not ok: %s", r.Reason)
	}
	if r.Status != "full" || !strings.Contains(r.FinalURL, "/article") {
		t.Fatalf("status=%q final=%q", r.Status, r.FinalURL)
	}
}

func TestRedirectTruthPlainFetchStillFull(t *testing.T) {
	srv := newLadderServer(t)
	ResetCookieJar()
	r := FetchURL(srv.root()+"/article", &Options{TimeoutS: 10})
	if !r.OK() || r.Status != "full" {
		t.Fatalf("status=%q reason=%q", r.Status, r.Reason)
	}
}
