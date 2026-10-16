// proxy.go — the web proxy (Settings → Proxy rides the Websites
// pipeline): the config contract, the preflight, and the route builder
// (the Go shape of _build_opener: socks5/socks4 via the dialers, the
// http-type via ProxyURL, direct/loopback with NO proxy of any kind).
package webfetch

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Proxy — the web-fetch proxy (proxy_from_config's shape).
type Proxy struct {
	Type     string `json:"type"`
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Username string `json:"username"`
	Password string `json:"password"`
}

// WebProxyError — a proxy-specific failure with a user-facing remedy
// (unreachable proxy). Never escapes FetchURL — it becomes a
// FetchResult reason so one bad proxy never kills a batch.
type WebProxyError struct{ msg string }

func (e *WebProxyError) Error() string { return e.msg }

// ProxyFromConfig extracts the web-fetch proxy from the app config
// (JSON-shaped map). Never errors — an unusable config means NO proxy:
//   - config["proxy"] missing/not a map, or enabled falsy → nil
//   - use_for_web explicitly false → nil; MISSING → ON (v0.19.0's fix
//     works on the existing config.json without touching Settings)
//   - type must be socks5/socks4/http; host non-empty; port 1..65535
//     (defaults 127.0.0.1:10808 — the v2rayN listener).
func ProxyFromConfig(config map[string]any) *Proxy {
	blockAny, ok := config["proxy"]
	if !ok || blockAny == nil {
		return nil
	}
	block, ok := blockAny.(map[string]any)
	if !ok || !pyTruthy(block["enabled"]) {
		return nil
	}
	if v, present := block["use_for_web"]; present && v == false {
		return nil
	}
	ptype := strings.ToLower(strings.TrimSpace(pyStr(block["type"], DefaultProxyType)))
	legal := false
	for _, t := range ProxyTypes {
		if ptype == t {
			legal = true
			break
		}
	}
	if !legal {
		return nil
	}
	host := strings.TrimSpace(pyStr(block["host"], DefaultProxyHost))
	if host == "" {
		return nil
	}
	port := DefaultProxyPort
	switch raw := block["port"].(type) {
	case nil:
	case string:
		if strings.TrimSpace(raw) == "" {
			break
		}
		n, err := strconv.Atoi(strings.TrimSpace(raw))
		if err != nil {
			return nil
		}
		port = n
	case float64:
		port = int(raw)
	case int:
		port = raw
	default:
		return nil
	}
	if port < 1 || port > 65535 {
		return nil
	}
	return &Proxy{
		Type:     ptype,
		Host:     host,
		Port:     port,
		Username: pyStr(block["username"], ""),
		Password: pyStr(block["password"], ""),
	}
}

// pyStr — Python's str(x or default): non-empty value wins.
func pyStr(v any, def string) string {
	switch x := v.(type) {
	case nil:
		return def
	case string:
		if x == "" {
			return def
		}
		return x
	case float64:
		return strconv.FormatFloat(x, 'f', -1, 64)
	case bool:
		if x {
			return "True"
		}
		return "False"
	default:
		return fmt.Sprintf("%v", x)
	}
}

// ProxyLabel — "SOCKS5 127.0.0.1:10808" for log lines.
func ProxyLabel(p *Proxy) string {
	if p == nil {
		return "none"
	}
	return fmt.Sprintf("%s %s:%d", strings.ToUpper(p.Type), p.Host, p.Port)
}

// WebProxyPreflight — is the proxy usable RIGHT NOW? (TCP reachable.)
// Pure stdlib, never errors, fast — called once per Websites batch so
// the whole run knows before link #1.
func WebProxyPreflight(p *Proxy, timeoutS float64) (bool, string) {
	if p == nil {
		return false, "no proxy configured"
	}
	if timeoutS <= 0 {
		timeoutS = ProxyPreflightTimeoutS
	}
	conn, err := net.DialTimeout("tcp",
		net.JoinHostPort(p.Host, strconv.Itoa(p.Port)),
		time.Duration(timeoutS*float64(time.Second)))
	if err != nil {
		return false, fmt.Sprintf("%s:%d unreachable (%s) — is the "+
			"proxy client (v2rayN) running?", p.Host, p.Port, dialErrText(err))
	}
	_ = conn.Close()
	return true, ""
}

// dialErrText — the OS error's text (Python's strerror-or-e).
func dialErrText(err error) string {
	var oe *net.OpError
	if e, ok := err.(*net.OpError); ok {
		oe = e
	}
	if oe != nil && oe.Err != nil {
		return oe.Err.Error()
	}
	return err.Error()
}

// ---------------------------------------------------------------------------
// The route builder — the Go shape of _build_opener.
// ---------------------------------------------------------------------------

// route is one door's transport bundle: proxy/direct/loopback. The
// impersonation door builds its own (tls-client).
type route struct {
	transport *http.Transport
}

// buildRoute builds the transport for one fetch route:
//   - forceDirect (loopback, the v0.15.1 rule): NO proxy of any kind,
//     not even the environment (Transport.Proxy nil).
//   - proxy set: ALL traffic forced through it (socks5/socks4 via the
//     dialers with DNS at the exit; the http-type via ProxyURL —
//     proxy-side DNS, the same protection as rdns=True).
//   - neither: today's direct behavior, env proxies deliberately NOT
//     honored (ADR-0006 §1 — Python's explicit-opener semantics).
func buildRoute(p *Proxy, forceDirect bool) (*route, error) {
	if p == nil || forceDirect {
		return &route{transport: &http.Transport{
			Proxy:               nil,
			ForceAttemptHTTP2:   false,
			MaxIdleConns:        4,
			IdleConnTimeout:     30 * time.Second,
			TLSHandshakeTimeout: 30 * time.Second,
		}}, nil
	}
	switch p.Type {
	case "http":
		auth := ""
		if p.Username != "" {
			auth = pyQuote(p.Username) + ":" + pyQuote(p.Password) + "@"
		}
		raw := fmt.Sprintf("http://%s%s:%d", auth, p.Host, p.Port)
		u, err := url.Parse(raw)
		if err != nil {
			return nil, &WebProxyError{msg: fmt.Sprintf(
				"proxy %s:%d unusable: %v", p.Host, p.Port, err)}
		}
		return &route{transport: &http.Transport{
			Proxy: http.ProxyURL(u),
		}}, nil
	case "socks5", "socks4":
		dial, err := socksDialer(p)
		if err != nil {
			return nil, err
		}
		return &route{transport: &http.Transport{
			DialContext: dial,
		}}, nil
	}
	return nil, &WebProxyError{msg: fmt.Sprintf(
		"proxy %s:%d has an unsupported type %q", p.Host, p.Port, p.Type)}
}

// buildClient — the fetch client for one route: the transport plus
// the house CheckRedirect (ErrUseLastResponse — the redirect ladder is
// OUR loop) and the wall-clock timeout.
func buildClient(p *Proxy, forceDirect bool, timeoutS float64) (*http.Client, error) {
	r, err := buildRoute(p, forceDirect)
	if err != nil {
		return nil, err
	}
	return &http.Client{
		Transport: r.transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
		Timeout: time.Duration(timeoutS * float64(time.Second)),
	}, nil
}

// contextDialer adapts a DialContext into the transport shape.
type contextDialer func(ctx context.Context, network, addr string) (net.Conn, error)

var _ = contextDialer(nil)
