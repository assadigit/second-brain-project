package workerclient

// dohdial.go — THE REACH-CLOUDFLARE LAW (the owner's SESSION-075
// directive + the SESSION-076 resilience expansion): "the app must be
// able to fetch from telegram bot, regardless of the windows is proxy
// on or off, because telegram bot is connected to cloudflare and
// system can try to reach cloudflare instead."
//
// The Worker channel was ALREADY proxy-independent (directClient's
// Proxy == nil — the v2rayN-interception fix; both proxy states ride
// the same direct sockets). The gap this file closes is the DNS class:
// with the system proxy OFF, a censored local resolver can poison the
// workers.dev name (NXDOMAIN or a black-hole address) — and the queue
// poll dies before a single byte crosses. THE FIX: when the plain dial
// fails, ask DoH for the name, then dial the ANSWER's address directly;
// the TLS layer still handshakes with the ORIGINAL hostname (SNI +
// certificate verification ride the addr the transport passed in — the
// answer's IP only carries the packets).
//
// THE THREE-ENDPOINT RESILIENCE LAW (SESSION-076): the DoH question
// rides a TIER of three endpoints, each dialed BY IP with its own name
// as SNI (the poisoned local resolver is never consulted on any leg):
//   1. Cloudflare primary — 1.1.1.1 / cloudflare-dns.com (the original
//      SESSION-075 endpoint).
//   2. Cloudflare secondary — 1.0.0.1 / cloudflare-dns.com (the same
//      anycast service on the companion IP; resilience if 1.1.1.1 is
//      specifically blocked at the network edge).
//   3. Quad9 — 9.9.9.9 / dns.quad9.net (a DIFFERENT provider for
//      cross-provider redundancy; if Cloudflare's anycast is down or
//      hijacked, Quad9's answer still rides).
// The first endpoint that returns a non-empty A-record list wins; an
// endpoint that times out or returns NXDOMAIN/junk yields nil and the
// tier walks on. The whole tier is bounded by dohTimeout.
//
// Read-only law: the DoH question is a GET; no write verb anywhere.
// Seams: dohResolveFn + plainDialFn + dohEndpoints (the hermetic-suite
// law — the loopback fakes swap every leg without any network).

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// dohEndpoint is one DoH resolver's own coordinates (the SNI/TLS name
// + the IP to dial + the port). Dialing the IP with the name's SNI
// keeps the question independent of the local resolver entirely.
type dohEndpoint struct {
	name string
	ip   string
	port string
}

// dohEndpoints is the TIER (vars, not consts — the loopback fakes
// redirect the slice; the first endpoint that answers wins). Order
// matters: Cloudflare primary first (the original law's endpoint),
// Cloudflare secondary second (the companion anycast IP), Quad9 third
// (the cross-provider redundancy leg).
var dohEndpoints = []dohEndpoint{
	{name: "cloudflare-dns.com", ip: "1.1.1.1", port: "443"},
	{name: "cloudflare-dns.com", ip: "1.0.0.1", port: "443"},
	{name: "dns.quad9.net", ip: "9.9.9.9", port: "443"},
}

// dohResolveFn is the resolution seam (the suites inject canned
// answers); plainDialFn is the plain-dial seam; dohTLSConf is the
// endpoint TLS seam (production verifies the endpoint's certificate —
// the loopback suites substitute the test server's own root).
var (
	dohResolveFn = resolveViaDoH
	plainDialFn  = defaultPlainDial
	dohTLSConf   = defaultDoHTLSConf
)

// defaultDoHTLSConf — the endpoint's name-verified TLS.
func defaultDoHTLSConf(name string) *tls.Config {
	return &tls.Config{ServerName: name}
}

// SetDoHTestRoot swaps the endpoint's root pool (the loopback suite's
// own door; never called outside tests).
func SetDoHTestRoot(pool *x509.CertPool) func() {
	old := dohTLSConf
	dohTLSConf = func(name string) *tls.Config {
		return &tls.Config{ServerName: name, RootCAs: pool}
	}
	return func() { dohTLSConf = old }
}

// dohTimeout bounds the whole DoH question (a slow question must not
// hold the worker leg hostage — the plain dial's own timeout governs
// the data path). The tier walks within this one budget; a single
// blocked endpoint cannot exhaust it (the per-endpoint TCP dial +
// TLS handshake + read ride the http.Transport's own default
// timeouts; a stalled endpoint returns nil well within the budget).
const dohTimeout = 4 * time.Second

// dialTimeout is each TCP dial's own budget (both legs).
const dialTimeout = 10 * time.Second

// askDoHEndpoint asks ONE endpoint for host's A records. Never errors
// loudly: no answer yields nil (the caller walks to the next endpoint
// in the tier). The question rides a TLS conn dialed BY IP with the
// endpoint's name as SNI — the poisoned local resolver is never
// consulted on any leg.
func askDoHEndpoint(ctx context.Context, ep dohEndpoint, host string) []string {
	addr := net.JoinHostPort(ep.ip, ep.port)
	transport := &http.Transport{
		DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
			var d net.Dialer
			return d.DialContext(ctx, "tcp", addr)
		},
		TLSClientConfig: dohTLSConf(ep.name),
	}
	defer transport.CloseIdleConnections()
	u := "https://" + ep.name + "/dns-query?name=" +
		url.QueryEscape(host) + "&type=A"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return nil
	}
	req.Header.Set("Accept", "application/dns-json")
	req.Header.Set("User-Agent", HTTPUserAgent)
	res, err := transport.RoundTrip(req)
	if err != nil {
		return nil
	}
	defer func() { _ = res.Body.Close() }()
	raw, err := io.ReadAll(io.LimitReader(res.Body, 65536))
	if err != nil {
		return nil
	}
	var data struct {
		Status int `json:"Status"`
		Answer []struct {
			Type int    `json:"type"`
			Data string `json:"data"`
		} `json:"Answer"`
	}
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil
	}
	if data.Status != 0 {
		return nil // NXDOMAIN (3) and friends: no answer to ride
	}
	var out []string
	for _, a := range data.Answer {
		if a.Type == 1 && net.ParseIP(a.Data) != nil {
			out = append(out, a.Data)
		}
	}
	return out
}

// resolveViaDoH asks the DoH tier for host's A records: each endpoint
// in turn until one returns a non-empty answer. Never errors loudly:
// no answer from any endpoint yields nil (the caller keeps the plain
// leg's own error as the honest cause). THE THREE-ENDPOINT LAW: the
// first endpoint that answers wins; an endpoint that returns nil
// (NXDOMAIN, junk, timeout, TLS failure) walks on to the next.
func resolveViaDoH(ctx context.Context, host string) []string {
	host = strings.TrimSpace(host)
	if host == "" {
		return nil
	}
	if net.ParseIP(host) != nil {
		return nil // an IP literal needs no question
	}
	qctx, cancel := context.WithTimeout(ctx, dohTimeout)
	defer cancel()
	for _, ep := range dohEndpoints {
		if ips := askDoHEndpoint(qctx, ep, host); len(ips) > 0 {
			return ips
		}
	}
	return nil
}

// defaultPlainDial is the first leg: the ordinary direct TCP dial
// (no proxy, ever — the channel's own law).
func defaultPlainDial(ctx context.Context, network, addr string) (net.Conn, error) {
	var d net.Dialer
	return d.DialContext(ctx, network, addr)
}

// dohFallbackDial is THE TRANSPORT'S DIALER: leg 1 — the plain direct
// dial (the system resolver, exactly as before); on failure — the DoH
// leg: the tier's answer for the name, dialed directly. The ORIGINAL
// leg-1 error is what surfaces when both legs fail (the honest first
// cause, never the fallback's).
func dohFallbackDial(ctx context.Context, network, addr string) (net.Conn, error) {
	conn, err := plainDialFn(ctx, network, addr)
	if err == nil {
		return conn, nil
	}
	host, port, splitErr := net.SplitHostPort(addr)
	if splitErr != nil || host == "" {
		return nil, err
	}
	ips := dohResolveFn(ctx, host)
	for _, ip := range ips {
		c, dialErr := plainDialFn(ctx, network, net.JoinHostPort(ip, port))
		if dialErr == nil {
			return c, nil
		}
	}
	return nil, err
}
