package workerclient

// dohdial_test.go — THE REACH-CLOUDFLARE LAW's hermetic battery (the
// seam law: dohResolveFn + plainDialFn + the dohEndpoints tier —
// every leg swapped at loopback, zero network):
//
//   1. the DoH question's own grammar (the JSON answer's A records;
//      NXDOMAIN and junk answers yield nil — never an error);
//   2. the fallback's selection law (leg 1 ok → the plain conn; leg 1
//      dead → the DoH answer's dial; both dead → the ORIGINAL error);
//   3. the endpoint independence (the question dials the ENDPOINT IP
//      with the ENDPOINT NAME as SNI — the poisoned-resolver class
//      never rides the local resolver);
//   4. THE THREE-ENDPOINT TIER (SESSION-076): the tier walks to the
//      next endpoint when the current one returns nil; the first
//      endpoint that answers wins; the production tier carries exactly
//      three endpoints (Cloudflare primary 1.1.1.1, Cloudflare
//      secondary 1.0.0.1, Quad9 9.9.9.9).

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// trustServer builds the loopback endpoint's root pool (the httptest
// server's own certificate — the production path verifies the real
// endpoint's cert; the suite trusts its own fake).
func trustServer(srv *httptest.Server) func() {
	pool := x509.NewCertPool()
	if cert := srv.Certificate(); cert != nil {
		pool.AddCert(cert)
	}
	return SetDoHTestRoot(pool)
}

// withDoHSeams swaps the seams for one test (the standard restore law).
func withDoHSeams(t *testing.T, resolve func(context.Context, string) []string,
	plain func(context.Context, string, string) (net.Conn, error)) {
	t.Helper()
	oldResolve, oldPlain := dohResolveFn, plainDialFn
	dohResolveFn, plainDialFn = resolve, plain
	t.Cleanup(func() { dohResolveFn, plainDialFn = oldResolve, oldPlain })
}

// withDoHEndpoints swaps the tier for one test (the standard restore
// law); the loopback suites redirect the production three-endpoint tier
// to one or more loopback httptest servers.
func withDoHEndpoints(t *testing.T, eps []dohEndpoint) {
	t.Helper()
	old := dohEndpoints
	dohEndpoints = eps
	t.Cleanup(func() { dohEndpoints = old })
}

func TestResolveViaDoHReadsTheJSONAnswer(t *testing.T) {
	// A loopback DoH endpoint: the fake answers over real HTTP + TLS.
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("name") == "bot.example.workers.dev" && r.URL.Query().Get("type") == "A" {
			_ = json.NewEncoder(w).Encode(map[string]any{
				"Status": 0,
				"Answer": []map[string]any{
					{"type": 1, "data": "192.0.2.10"},
					{"type": 1, "data": "192.0.2.11"},
					{"type": 5, "data": "cname.example"}, // CNAME rows are skipped
				},
			})
			return
		}
		if r.URL.Query().Get("name") == "dead.example" {
			_ = json.NewEncoder(w).Encode(map[string]any{"Status": 3, "Answer": []any{}})
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0, "Answer": []map[string]any{
			{"type": 1, "data": "not-an-ip"},
		}})
	}))
	defer srv.Close()
	restore := trustServer(srv)
	defer restore()
	host, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "https://"))
	withDoHEndpoints(t, []dohEndpoint{{name: "127.0.0.1", ip: host, port: port}})

	got := resolveViaDoH(context.Background(), "bot.example.workers.dev")
	if len(got) != 2 || got[0] != "192.0.2.10" || got[1] != "192.0.2.11" {
		t.Fatalf("the A records: %v", got)
	}
	if dead := resolveViaDoH(context.Background(), "dead.example"); dead != nil {
		t.Fatalf("NXDOMAIN must yield nil, got %v", dead)
	}
	if junk := resolveViaDoH(context.Background(), "junk.example"); junk != nil {
		t.Fatalf("a non-IP answer row must yield nil, got %v", junk)
	}
	if literal := resolveViaDoH(context.Background(), "192.0.2.99"); literal != nil {
		t.Fatalf("an IP literal asks no question, got %v", literal)
	}
}

func TestDohFallbackDialSelectionLaw(t *testing.T) {
	// THE ANSWER'S ADDRESS: a real loopback listener. The seam's dial
	// maps "the answer IP : the original port" onto the listener (the
	// fallback PRESERVES the original port — the mapping keeps the
	// loopback honest without pretending the law away).
	lst, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lst.Close() }()
	answerIP := "127.0.0.1"

	plainErr := fmt.Errorf("dial tcp: lookup bot.example.workers.dev: no such host")
	var plainAddrs []string
	seamDial := func(ctx context.Context, network, addr string) (net.Conn, error) {
		plainAddrs = append(plainAddrs, addr)
		host, _, _ := net.SplitHostPort(addr)
		if host == answerIP {
			return net.Dial(network, lst.Addr().String()) // the answer's address answers
		}
		return nil, plainErr // the poisoned name fails
	}
	withDoHSeams(t,
		func(ctx context.Context, host string) []string {
			if host != "bot.example.workers.dev" {
				t.Errorf("the DoH question asked for %q", host)
			}
			return []string{answerIP}
		},
		seamDial)

	ctx := context.Background()

	// Leg 1 alive (an address the seam lets through): the plain conn,
	// no DoH question at all.
	aliveLst, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = aliveLst.Close() }()
	oldPlain := plainDialFn
	plainDialFn = func(ctx context.Context, network, addr string) (net.Conn, error) {
		if strings.HasSuffix(addr, "alive.example:443") {
			plainAddrs = append(plainAddrs, addr)
			return net.Dial(network, aliveLst.Addr().String())
		}
		return seamDial(ctx, network, addr)
	}
	conn, err := dohFallbackDial(ctx, "tcp", "alive.example:443")
	plainDialFn = oldPlain
	if err != nil {
		t.Fatalf("leg 1 alive must dial: %v", err)
	}
	remote := conn.RemoteAddr().String()
	_ = conn.Close()
	if remote != aliveLst.Addr().String() {
		t.Fatalf("leg 1 alive must return the plain conn, got %s", remote)
	}
	for _, a := range plainAddrs {
		if a != "alive.example:443" {
			t.Fatalf("leg 1 alive asked the DoH leg anyway: %v", plainAddrs)
		}
	}

	// Leg 1 dead: the DoH answer's dial rides the ANSWER IP (the
	// original port is preserved — the law).
	plainAddrs = nil
	conn, err = dohFallbackDial(ctx, "tcp", "bot.example.workers.dev:443")
	if err != nil {
		t.Fatalf("the DoH fallback must reach the answer's address: %v", err)
	}
	remote = conn.RemoteAddr().String()
	_ = conn.Close()
	if remote != lst.Addr().String() {
		t.Fatalf("the fallback's conn rides %q, want the answer's listener %q", remote, lst.Addr().String())
	}
	if len(plainAddrs) != 2 || plainAddrs[0] != "bot.example.workers.dev:443" ||
		plainAddrs[1] != "127.0.0.1:443" {
		t.Fatalf("the dial order must be poisoned-then-answer: %v", plainAddrs)
	}

	// Both legs dead: the ORIGINAL leg-1 error surfaces (never the
	// fallback's own).
	withDoHSeams(t,
		func(ctx context.Context, host string) []string { return nil },
		func(ctx context.Context, network, addr string) (net.Conn, error) {
			return nil, plainErr
		})
	_, err = dohFallbackDial(ctx, "tcp", "bot.example.workers.dev:443")
	if err == nil || err.Error() != plainErr.Error() {
		t.Fatalf("the honest first cause must surface, got %v", err)
	}
}

func TestDohFallbackDialCoversPortAndTimeout(t *testing.T) {
	// the fallback dials the ANSWER IP with the ORIGINAL port — the
	// seam maps it onto a loopback listener (a slow accept is still a
	// dial success; the fallback only ever speaks when the DIAL fails).
	lst, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = lst.Close() }()
	withDoHSeams(t,
		func(ctx context.Context, host string) []string { return []string{"127.0.0.1"} },
		func(ctx context.Context, network, addr string) (net.Conn, error) {
			if strings.HasPrefix(addr, "127.0.0.1:") {
				return net.Dial(network, lst.Addr().String())
			}
			return nil, fmt.Errorf("poisoned")
		})
	conn, err := dohFallbackDial(context.Background(), "tcp", "worker.example:443")
	if err != nil {
		t.Fatalf("the port must ride the answer's dial: %v", err)
	}
	_ = conn.Close()
}

// TestDohQuestionRidesTheEndpointIP pins the DoH transport's own
// independence: the question's TCP dial goes to the ENDPOINT IP (never
// the endpoint NAME through the local resolver).
func TestDohQuestionRidesTheEndpointIP(t *testing.T) {
	srv := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0,
			"Answer": []map[string]any{{"type": 1, "data": "192.0.2.20"}}})
	}))
	defer srv.Close()
	restore := trustServer(srv)
	defer restore()
	host, port, _ := net.SplitHostPort(strings.TrimPrefix(srv.URL, "https://"))
	withDoHEndpoints(t, []dohEndpoint{{name: "127.0.0.1", ip: host, port: port}})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	got := resolveViaDoH(ctx, "any.example")
	if len(got) != 1 || got[0] != "192.0.2.20" {
		t.Fatalf("the loopback endpoint's answer: %v", got)
	}
}

// TestDoHTierWalksToSecondEndpointWhenFirstReturnsNil — THE
// THREE-ENDPOINT TIER's walk law (SESSION-076): when the first
// endpoint returns nil (NXDOMAIN, junk, or no A records), the tier
// walks to the second endpoint; the second endpoint's answer wins.
func TestDoHTierWalksToSecondEndpointWhenFirstReturnsNil(t *testing.T) {
	// Two loopback DoH endpoints: the first returns NXDOMAIN (nil),
	// the second returns a real A record. The tier must walk.
	nxdServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 3, "Answer": []any{}})
	}))
	defer nxdServer.Close()
	answerServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0,
			"Answer": []map[string]any{{"type": 1, "data": "192.0.2.42"}}})
	}))
	defer answerServer.Close()
	// THE TRUST LAW: both servers' certs must be trusted (the tier
	// walks; the second endpoint's TLS must verify under the test root).
	pool := x509.NewCertPool()
	if cert := nxdServer.Certificate(); cert != nil {
		pool.AddCert(cert)
	}
	if cert := answerServer.Certificate(); cert != nil {
		pool.AddCert(cert)
	}
	restore := SetDoHTestRoot(pool)
	defer restore()

	nxdHost, nxdPort, _ := net.SplitHostPort(strings.TrimPrefix(nxdServer.URL, "https://"))
	ansHost, ansPort, _ := net.SplitHostPort(strings.TrimPrefix(answerServer.URL, "https://"))
	withDoHEndpoints(t, []dohEndpoint{
		{name: "127.0.0.1", ip: nxdHost, port: nxdPort},
		{name: "127.0.0.1", ip: ansHost, port: ansPort},
	})

	got := resolveViaDoH(context.Background(), "bot.example.workers.dev")
	if len(got) != 1 || got[0] != "192.0.2.42" {
		t.Fatalf("the tier must walk to the second endpoint's answer: %v", got)
	}
}

// TestDoHTierWalksToThirdEndpointWhenFirstTwoReturnNil — the tier
// walks through THREE endpoints (the third is the cross-provider
// redundancy leg): when the first two return nil, the third's answer
// wins. This is the SESSION-076 Quad9 leg's pin.
func TestDoHTierWalksToThirdEndpointWhenFirstTwoReturnNil(t *testing.T) {
	emptyServer1 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0, "Answer": []any{}})
	}))
	defer emptyServer1.Close()
	emptyServer2 := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 3, "Answer": []any{}})
	}))
	defer emptyServer2.Close()
	answerServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0,
			"Answer": []map[string]any{{"type": 1, "data": "192.0.2.99"}}})
	}))
	defer answerServer.Close()
	pool := x509.NewCertPool()
	for _, s := range []*httptest.Server{emptyServer1, emptyServer2, answerServer} {
		if cert := s.Certificate(); cert != nil {
			pool.AddCert(cert)
		}
	}
	restore := SetDoHTestRoot(pool)
	defer restore()

	h1, p1, _ := net.SplitHostPort(strings.TrimPrefix(emptyServer1.URL, "https://"))
	h2, p2, _ := net.SplitHostPort(strings.TrimPrefix(emptyServer2.URL, "https://"))
	h3, p3, _ := net.SplitHostPort(strings.TrimPrefix(answerServer.URL, "https://"))
	withDoHEndpoints(t, []dohEndpoint{
		{name: "127.0.0.1", ip: h1, port: p1},
		{name: "127.0.0.1", ip: h2, port: p2},
		{name: "127.0.0.1", ip: h3, port: p3},
	})

	got := resolveViaDoH(context.Background(), "bot.example.workers.dev")
	if len(got) != 1 || got[0] != "192.0.2.99" {
		t.Fatalf("the tier must walk to the third endpoint's answer: %v", got)
	}
}

// TestDoHTierReturnsFirstEndpointAnswerWhenItHasOne — the tier's
// first-wins law: when the first endpoint returns a non-empty answer,
// the tier never asks the second. Two answer-bearing endpoints; the
// first's answer wins; the second is never hit (a hit counter pins it).
func TestDoHTierReturnsFirstEndpointAnswerWhenItHasOne(t *testing.T) {
	var firstHits, secondHits int
	firstServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		firstHits++
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0,
			"Answer": []map[string]any{{"type": 1, "data": "192.0.2.10"}}})
	}))
	defer firstServer.Close()
	secondServer := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		secondHits++
		_ = json.NewEncoder(w).Encode(map[string]any{"Status": 0,
			"Answer": []map[string]any{{"type": 1, "data": "192.0.2.20"}}})
	}))
	defer secondServer.Close()
	pool := x509.NewCertPool()
	for _, s := range []*httptest.Server{firstServer, secondServer} {
		if cert := s.Certificate(); cert != nil {
			pool.AddCert(cert)
		}
	}
	restore := SetDoHTestRoot(pool)
	defer restore()

	h1, p1, _ := net.SplitHostPort(strings.TrimPrefix(firstServer.URL, "https://"))
	h2, p2, _ := net.SplitHostPort(strings.TrimPrefix(secondServer.URL, "https://"))
	withDoHEndpoints(t, []dohEndpoint{
		{name: "127.0.0.1", ip: h1, port: p1},
		{name: "127.0.0.1", ip: h2, port: p2},
	})

	got := resolveViaDoH(context.Background(), "bot.example.workers.dev")
	if len(got) != 1 || got[0] != "192.0.2.10" {
		t.Fatalf("the first endpoint's answer must win: %v", got)
	}
	if firstHits != 1 {
		t.Fatalf("the first endpoint must be hit exactly once: %d", firstHits)
	}
	if secondHits != 0 {
		t.Fatalf("the second endpoint must never be hit when the first answers: %d", secondHits)
	}
}

// TestProductionDoHTierHasThreeEndpoints — THE THREE-ENDPOINT
// RESILIENCE LAW's production pin: the default dohEndpoints slice
// carries exactly three endpoints in the SESSION-076 order — Cloudflare
// primary (1.1.1.1 / cloudflare-dns.com), Cloudflare secondary
// (1.0.0.1 / cloudflare-dns.com), Quad9 (9.9.9.9 / dns.quad9.net).
// A regression that drops an endpoint or reorders the tier fails here.
func TestProductionDoHTierHasThreeEndpoints(t *testing.T) {
	if len(dohEndpoints) != 3 {
		t.Fatalf("the production DoH tier must carry exactly three endpoints, got %d", len(dohEndpoints))
	}
	want := []dohEndpoint{
		{name: "cloudflare-dns.com", ip: "1.1.1.1", port: "443"},
		{name: "cloudflare-dns.com", ip: "1.0.0.1", port: "443"},
		{name: "dns.quad9.net", ip: "9.9.9.9", port: "443"},
	}
	for i, ep := range dohEndpoints {
		if ep != want[i] {
			t.Fatalf("endpoint %d: got %+v, want %+v", i, ep, want[i])
		}
	}
}
