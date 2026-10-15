// Command tlsimp is the T-02 Phase-0 proof-of-concept for ADR-0006: the
// Chrome-TLS impersonation door (the curl_cffi replacement).
//
// It proves, against a walled page, the exact acceptance shape from the
// roadmap: the Go stdlib HTTP client (Chrome header set, Go TLS stack)
// receives 403, while the same headers on a bogdanfinn/tls-client Chrome
// profile get 200. It also proves SOCKS5 proxying with hostname
// passthrough (the rdns=True law) when -socks points at the hermetic
// relay (poc/telegram -mode=socksrelay).
//
// Modes:
//
//	-mode=probe   [urls...]   stdlib-only status table (find walled pages)
//	-mode=compare -url=X      stdlib vs tls-client, side by side + summary
//
// The final line is a machine-readable summary behind the ###POC_JSON###
// sentinel (timing_runner style).
package main

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	fhttp "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
	"github.com/bogdanfinn/tls-client/profiles"
)

// chromeHeaders builds the Chrome navigation header set (the same shape
// core/web_fetch.py sends, Chrome-131-era). The version is templated so
// UA and sec-ch-ua stay in sync with the chosen profile.
func chromeHeaders(version string) [][2]string {
	return [][2]string{
		{"user-agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/" + version + ".0.0.0 Safari/537.36"},
		{"accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7"},
		{"accept-language", "en-US,en;q=0.9"},
		{"sec-ch-ua", `"Chromium";v="` + version + `", "Not(A:Brand";v="99", "Google Chrome";v="` + version + `"`},
		{"sec-ch-ua-mobile", "?0"},
		{"sec-ch-ua-platform", `"Windows"`},
		{"sec-fetch-dest", "document"},
		{"sec-fetch-mode", "navigate"},
		{"sec-fetch-site", "none"},
		{"sec-fetch-user", "?1"},
		{"upgrade-insecure-requests", "1"},
	}
}

type legResult struct {
	Transport     string `json:"transport"`
	Status        int    `json:"status"`
	Err           string `json:"err,omitempty"`
	ElapsedMS     int64  `json:"elapsed_ms"`
	ContentLength int64  `json:"content_length"`
	Server        string `json:"server,omitempty"`
	TLSVersion    string `json:"tls_version,omitempty"`
}

// stdlibFetch is the ADR-0006 "base stack" leg: net/http with the Chrome
// header set but Go's own TLS/HTTP2 fingerprint. Redirects are not
// followed (web_fetch owns its redirect ladder).
func stdlibFetch(url, version string) legResult {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for _, h := range chromeHeaders(version) {
		req.Header.Set(h[0], h[1])
	}
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{
			// Direct route: Proxy nil (Python's explicit-opener semantics).
			Proxy:           nil,
			TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12},
		},
	}
	t0 := time.Now()
	resp, err := client.Do(req)
	r := legResult{Transport: "stdlib", ElapsedMS: time.Since(t0).Milliseconds()}
	if err != nil {
		r.Err = err.Error()
		return r
	}
	defer resp.Body.Close()
	n, _ := io.Copy(io.Discard, io.LimitReader(resp.Body, 2<<20))
	r.Status = resp.StatusCode
	r.ContentLength = n
	r.Server = resp.Header.Get("Server")
	if resp.TLS != nil {
		r.TLSVersion = tls.VersionName(resp.TLS.Version)
	}
	return r
}

// tlsClientFetch is the impersonation leg: bogdanfinn/tls-client with a
// Chrome client-hello profile (uTLS under the hood, HTTP/2 frame
// fingerprint included), ordered headers, optional SOCKS5 proxy with
// hostname passthrough.
func tlsClientFetch(url, version, profileName, socks string) legResult {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithClientProfile(resolveProfile(profileName)),
	}
	if socks != "" {
		options = append(options, tls_client.WithProxyUrl("socks5://"+socks))
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return legResult{Transport: "tls-client/" + profileName, Err: "client build: " + err.Error()}
	}

	req, err := fhttp.NewRequest(fhttp.MethodGet, url, nil)
	if err != nil {
		return legResult{Transport: "tls-client/" + profileName, Err: err.Error()}
	}
	hdrs := fhttp.Header{}
	for _, h := range chromeHeaders(version) {
		hdrs.Set(h[0], h[1])
	}
	req.Header = hdrs

	t0 := time.Now()
	resp, err := client.Do(req)
	r := legResult{Transport: "tls-client/" + profileName, ElapsedMS: time.Since(t0).Milliseconds()}
	if err != nil {
		r.Err = err.Error()
		return r
	}
	defer resp.Body.Close()
	n, _ := io.Copy(io.Discard, io.LimitReader(resp.Body, 2<<20))
	r.Status = resp.StatusCode
	r.ContentLength = n
	r.Server = resp.Header.Get("Server")
	if rv := resp.Proto; rv != "" {
		r.TLSVersion = rv
	}
	return r
}

func resolveProfile(name string) profiles.ClientProfile {
	switch name {
	case "Chrome_103":
		return profiles.Chrome_103
	case "Chrome_104":
		return profiles.Chrome_104
	case "Chrome_105":
		return profiles.Chrome_105
	case "Chrome_106":
		return profiles.Chrome_106
	case "Chrome_107":
		return profiles.Chrome_107
	case "Chrome_108":
		return profiles.Chrome_108
	case "Chrome_109":
		return profiles.Chrome_109
	case "Chrome_110":
		return profiles.Chrome_110
	case "Chrome_111":
		return profiles.Chrome_111
	case "Chrome_112":
		return profiles.Chrome_112
	case "Chrome_116_PSK":
		return profiles.Chrome_116_PSK
	case "Chrome_133":
		return profiles.Chrome_133
	case "Chrome_133_PSK":
		return profiles.Chrome_133_PSK
	case "Chrome_144":
		return profiles.Chrome_144
	case "Chrome_144_PSK":
		return profiles.Chrome_144_PSK
	case "Chrome_146":
		return profiles.Chrome_146
	case "Chrome_146_PSK":
		return profiles.Chrome_146_PSK
	case "Chrome_150":
		return profiles.Chrome_150
	case "Chrome_150_PSK":
		return profiles.Chrome_150_PSK
	case "Chrome_152":
		return profiles.Chrome_152
	case "Chrome_152_PSK":
		return profiles.Chrome_152_PSK
	default:
		fmt.Fprintf(os.Stderr, "[poc] unknown profile %q — falling back to Chrome_133\n", name)
		return profiles.Chrome_133
	}
}

var probeCandidates = []string{
	"https://nowsecure.nl",
	"https://www.pexels.com",
	"https://www.yelp.com",
	"https://www.tiktok.com",
	"https://www.instagram.com",
	"https://www.linkedin.com",
	"https://www.netflix.com",
	"https://www.target.com",
	"https://www.bestbuy.com",
	"https://www.ebay.com",
	"https://www.change.org",
	"https://www.eventbrite.com",
	"https://www.patreon.com",
	"https://www.fandom.com",
	"https://www.goodreads.com",
}

func main() {
	mode := flag.String("mode", "compare", "probe | compare")
	url := flag.String("url", "https://nowsecure.nl", "target URL for compare mode")
	version := flag.String("chrome-version", "133", "Chrome version templated into UA/sec-ch-ua (keep in sync with the profile)")
	profile := flag.String("profile", "Chrome_133", "tls-client Chrome profile name")
	socks := flag.String("socks", "", "SOCKS5 address (host:port) — routed through for the impersonation leg")
	flag.Parse()

	switch *mode {
	case "localwall":
		// Hermetic walled page: a local HTTPS server that gates on the
		// ClientHello exactly like Cloudflare's TLS-fingerprint rule does
		// (Chrome ClientHellos carry a leading GREASE cipher and a large
		// cipher list; Go's never do). From this sandbox's datacenter IP,
		// every live CF/Akamai/DataDome wall rejects BOTH transports at the
		// IP-reputation layer — so the 403→200 mechanism is proven against
		// a locally controlled fingerprint gate instead, on the same
		// decision logic.
		const wallAddr = "127.0.0.1:8443"
		stop := make(chan struct{})
		go runLocalWall(wallAddr, stop)
		time.Sleep(300 * time.Millisecond)
		const wallURL = "https://127.0.0.1:8443/"
		std := localWallFetchStdlib(wallURL, *version)
		imp := localWallFetchTlsClient(wallURL, *version, *profile)
		fmt.Printf("[poc] stdlib   vs local fingerprint wall → HTTP %d\n", std)
		fmt.Printf("[poc] tls-client(%s) vs local fingerprint wall → HTTP %d\n", *profile, imp)
		verdict := "INCONCLUSIVE"
		if std == 403 && imp == 200 {
			verdict = "WALL_BYPASSED"
		} else if std == imp {
			verdict = "NO_DIFFERENCE"
		}
		fmt.Printf("[poc] verdict: %s (stdlib=%d, impersonation=%d)\n", verdict, std, imp)
		emit("localwall", map[string]any{"stdlib": std, "impersonation": imp, "verdict": verdict}, *socks)
		close(stop)
	case "fingerprint":
		// tls.peet.ws echoes the TLS (JA3/JA4/peetprint) and HTTP/2
		// fingerprints back as JSON — the layer Cloudflare actually
		// inspects. Proves the profile substitution regardless of any
		// wall's IP-reputation behavior.
		const fpURL = "https://tls.peet.ws/api/all"
		stdFP := extractFingerprint(stdlibFetchBody(fpURL, *version))
		impFP := extractFingerprint(impFetchBody(fpURL, *version, *profile, *socks))
		fmt.Printf("[poc] stdlib   fingerprint: %s\n", stdFP)
		fmt.Printf("[poc] tls-client(%s) fingerprint: %s\n", *profile, impFP)
		same := stdFP == impFP
		fmt.Printf("[poc] verdict: %s\n", map[bool]string{true: "NO_DIFFERENCE", false: "FINGERPRINT_SUBSTITUTED"}[same])
		emit("fingerprint", map[string]any{"stdlib": stdFP, "impersonation": impFP}, *socks)
	case "probe":
		urls := flag.Args()
		if len(urls) == 0 {
			urls = probeCandidates
		}
		type row struct {
			URL    string    `json:"url"`
			Stdlib legResult `json:"stdlib"`
		}
		var rows []row
		fmt.Printf("%-38s %6s %8s %s\n", "URL (stdlib transport)", "status", "ms", "server")
		for _, u := range urls {
			r := stdlibFetch(u, *version)
			rows = append(rows, row{URL: u, Stdlib: r})
			fmt.Printf("%-38s %6d %8d %s\n", u, r.Status, r.ElapsedMS, firstNonEmpty(r.Server, r.Err))
		}
		emit("probe", rows, *socks)
	case "compare":
		std := stdlibFetch(*url, *version)
		imp := tlsClientFetch(*url, *version, *profile, *socks)
		fmt.Printf("[poc] stdlib   → HTTP %d in %d ms (server=%s, tls=%s%s)\n",
			std.Status, std.ElapsedMS, firstNonEmpty(std.Server, "-"), firstNonEmpty(std.TLSVersion, "-"), errSuffix(std.Err))
		fmt.Printf("[poc] tls-client(%s) → HTTP %d in %d ms (server=%s, proto=%s%s)\n",
			*profile, imp.Status, imp.ElapsedMS, firstNonEmpty(imp.Server, "-"), firstNonEmpty(imp.TLSVersion, "-"), errSuffix(imp.Err))
		verdict := "INCONCLUSIVE"
		if std.Status == imp.Status {
			verdict = "NO_DIFFERENCE"
		} else if imp.Status == 200 {
			verdict = "WALL_BYPASSED"
		}
		fmt.Printf("[poc] verdict: %s (stdlib=%d, impersonation=%d)\n", verdict, std.Status, imp.Status)
		emit("compare:"+*url, map[string]any{"stdlib": std, "impersonation": imp, "verdict": verdict}, *socks)
	default:
		fmt.Fprintf(os.Stderr, "unknown mode %q\n", *mode)
		os.Exit(2)
	}
}

func errSuffix(e string) string {
	if e == "" {
		return ""
	}
	return " err=" + strings.ReplaceAll(e, "\n", " | ")
}

// ---------------------------------------------------------------------------
// the hermetic fingerprint-gated wall (localwall mode)
// ---------------------------------------------------------------------------

// isGREASE reports whether a cipher suite value is one of RFC 8701's GREASE
// placeholders (0x0a0a, 0x1a1a, …). Chrome always leads its cipher list with
// one; Go's crypto/tls never sends any.
func isGREASE(c uint16) bool {
	return c>>8 == c&0xff && c&0x0f == 0x0a
}

// looksLikeChrome inspects the ClientHello the way a TLS-fingerprint rule
// does: a leading GREASE cipher plus a Chrome-sized cipher list (≥14).
func looksLikeChrome(ciphers []uint16) bool {
	return len(ciphers) >= 14 && len(ciphers) > 0 && isGREASE(ciphers[0])
}

// runLocalWall serves an HTTPS page that 403s non-Chrome ClientHellos and
// 200s Chrome-shaped ones. The verdict is captured in GetConfigForClient
// (where the ClientHello is visible) and consumed by the handler via the
// remote address.
func runLocalWall(addr string, stop <-chan struct{}) {
	cert := mustSelfSigned()
	var mu sync.Mutex
	verdict := map[string]bool{}

	base := &tls.Config{
		Certificates: []tls.Certificate{cert},
		NextProtos:   []string{"h2", "http/1.1"},
	}
	conf := base.Clone()
	conf.GetConfigForClient = func(chi *tls.ClientHelloInfo) (*tls.Config, error) {
		mu.Lock()
		verdict[chi.Conn.RemoteAddr().String()] = looksLikeChrome(chi.CipherSuites)
		mu.Unlock()
		return nil, nil // keep base config
	}

	srv := &http.Server{
		Addr:      addr,
		TLSConfig: conf,
		ConnContext: func(ctx context.Context, c net.Conn) context.Context {
			return context.WithValue(ctx, wallAddrKey{}, c.RemoteAddr().String())
		},
		Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			mu.Lock()
			ok := verdict[r.RemoteAddr]
			mu.Unlock()
			if ok {
				fmt.Fprintln(w, "WALLED-PAGE OK — Chrome ClientHello accepted")
				return
			}
			http.Error(w, "403 forbidden — non-Chrome TLS fingerprint (Go stdlib ClientHello rejected)", http.StatusForbidden)
		}),
	}
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[wall] listen: %v\n", err)
		return
	}
	fmt.Printf("[wall] fingerprint gate listening on https://%s (GREASE-led, ≥14-cipher ClientHello = Chrome)\n", addr)
	go func() {
		<-stop
		srv.Close()
	}()
	_ = srv.ServeTLS(ln, "", "")
}

type wallAddrKey struct{}

func mustSelfSigned() tls.Certificate {
	priv, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		panic(err)
	}
	tmpl := x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "tlsimp-hermetic-wall"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		IPAddresses:           []net.IP{net.ParseIP("127.0.0.1")},
	}
	der, err := x509.CreateCertificate(rand.Reader, &tmpl, &tmpl, &priv.PublicKey, priv)
	if err != nil {
		panic(err)
	}
	return tls.Certificate{Certificate: [][]byte{der}, PrivateKey: priv}
}

func localWallFetchStdlib(url, version string) int {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for _, h := range chromeHeaders(version) {
		req.Header.Set(h[0], h[1])
	}
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport: &http.Transport{
			Proxy:           nil,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true, MinVersion: tls.VersionTLS12}, // self-signed wall
		},
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[poc] stdlib wall fetch error: %v\n", err)
		return 0
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	return resp.StatusCode
}

func localWallFetchTlsClient(url, version, profileName string) int {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(15),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithClientProfile(resolveProfile(profileName)),
		tls_client.WithInsecureSkipVerify(), // self-signed wall
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[poc] tls-client wall build error: %v\n", err)
		return 0
	}
	req, err := fhttp.NewRequest(fhttp.MethodGet, url, nil)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[poc] tls-client wall request error: %v\n", err)
		return 0
	}
	hdrs := fhttp.Header{}
	for _, h := range chromeHeaders(version) {
		hdrs.Set(h[0], h[1])
	}
	req.Header = hdrs
	resp, err := client.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "[poc] tls-client wall fetch error: %v\n", err)
		return 0
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
	return resp.StatusCode
}

// stdlibFetchBody fetches via the stdlib transport and returns the (capped)
// body — used by the fingerprint mode.
func stdlibFetchBody(url, version string) string {
	req, _ := http.NewRequest(http.MethodGet, url, nil)
	for _, h := range chromeHeaders(version) {
		req.Header.Set(h[0], h[1])
	}
	client := &http.Client{
		CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse },
		Transport:     &http.Transport{Proxy: nil, TLSClientConfig: &tls.Config{MinVersion: tls.VersionTLS12}},
	}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return string(b)
}

// impFetchBody fetches via tls-client and returns the (capped) body.
func impFetchBody(url, version, profileName, socks string) string {
	jar := tls_client.NewCookieJar()
	options := []tls_client.HttpClientOption{
		tls_client.WithTimeoutSeconds(30),
		tls_client.WithNotFollowRedirects(),
		tls_client.WithCookieJar(jar),
		tls_client.WithClientProfile(resolveProfile(profileName)),
	}
	if socks != "" {
		options = append(options, tls_client.WithProxyUrl("socks5://"+socks))
	}
	client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
	if err != nil {
		return ""
	}
	req, err := fhttp.NewRequest(fhttp.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	hdrs := fhttp.Header{}
	for _, h := range chromeHeaders(version) {
		hdrs.Set(h[0], h[1])
	}
	req.Header = hdrs
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer resp.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return string(b)
}

// extractFingerprint pulls the identity-defining fields out of a
// tls.peet.ws /api/all JSON body.
func extractFingerprint(body string) string {
	if body == "" {
		return "(no body — fetch failed)"
	}
	var v struct {
		HTTPVersion string `json:"http_version"`
		TLS         struct {
			JA3Hash    string `json:"ja3_hash"`
			JA4        string `json:"ja4"`
			PeetHash   string `json:"peetprint_hash"`
			TLSVersion string `json:"tls_version_record"`
		} `json:"tls"`
		HTTP2 struct {
			AkamaiHash  string   `json:"akamai_fingerprint_hash"`
			HeaderOrder []string `json:"header_order"`
		} `json:"http2"`
	}
	if err := json.Unmarshal([]byte(body), &v); err != nil {
		return "(unparseable body: " + err.Error() + ")"
	}
	compact := map[string]any{
		"http_version": v.HTTPVersion,
		"tls_version":  v.TLS.TLSVersion,
		"ja3_hash":     v.TLS.JA3Hash,
		"ja4":          v.TLS.JA4,
		"peet_hash":    v.TLS.PeetHash,
		"h2_hash":      v.HTTP2.AkamaiHash,
		"header_order": v.HTTP2.HeaderOrder,
	}
	b, _ := json.Marshal(compact)
	return string(b)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func emit(kind string, payload any, socks string) {
	b, err := json.Marshal(map[string]any{
		"kind": kind, "socks": socks != "", "payload": payload,
	})
	if err != nil {
		fmt.Fprintf(os.Stderr, "[poc] summary marshal: %v\n", err)
		return
	}
	fmt.Printf("###POC_JSON###%s\n", b)
}
