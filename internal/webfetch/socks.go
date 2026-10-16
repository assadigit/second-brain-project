// socks.go — the SOCKS dialers for the stdlib door: SOCKS5 via
// golang.org/x/net/proxy (the dialer sends the HOSTNAME — ATYP=DOMAIN,
// never resolving locally: Python PySocks' rdns=True law) and a
// hand-rolled SOCKS4/4a CONNECT (PySocks' socks4 shape: a real IP
// speaks plain SOCKS4, a hostname rides the 4a extension — the same
// DNS-at-the-exit protection). Errors surface as WebProxyError so a
// dead proxy is a clear failed FetchResult, never a broken batch.
package webfetch

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"strconv"
	"time"

	"golang.org/x/net/proxy"
)

// socksDialer builds the DialContext for a socks5/socks4 proxy: every
// connection rides the proxy with the target hostname unresolved.
func socksDialer(p *Proxy) (func(ctx context.Context, network, addr string) (net.Conn, error), error) {
	proxyAddr := net.JoinHostPort(p.Host, strconv.Itoa(p.Port))
	switch p.Type {
	case "socks5":
		var auth *proxy.Auth
		if p.Username != "" || p.Password != "" {
			auth = &proxy.Auth{User: p.Username, Password: p.Password}
		}
		dialer, err := proxy.SOCKS5("tcp", proxyAddr, auth, proxy.Direct)
		if err != nil {
			return nil, &WebProxyError{msg: fmt.Sprintf(
				"proxy %s:%d unusable: %v", p.Host, p.Port, err)}
		}
		// rdns=True: x/net's SOCKS5 dialer passes the HOSTNAME through
		// (ATYP=DOMAIN) — the fake-SOCKS5 test proves it on the wire.
		return func(ctx context.Context, network, addr string) (net.Conn, error) {
			type dialerWithDeadlines interface {
				DialContext(ctx context.Context, network, addr string) (net.Conn, error)
			}
			if d, ok := dialer.(dialerWithDeadlines); ok {
				conn, err := d.DialContext(ctx, network, addr)
				return conn, wrapSocksErr(p, err)
			}
			conn, err := dialer.Dial(network, addr)
			return conn, wrapSocksErr(p, err)
		}, nil
	case "socks4":
		return func(ctx context.Context, network, addr string) (net.Conn, error) {
			host, portStr, err := net.SplitHostPort(addr)
			if err != nil {
				return nil, err
			}
			port, err := strconv.Atoi(portStr)
			if err != nil {
				return nil, err
			}
			conn, derr := socks4aConnect(proxyAddr, p.Username, host, port)
			return conn, wrapSocksErr(p, derr)
		}, nil
	}
	return nil, &WebProxyError{msg: fmt.Sprintf(
		"proxy %s:%d has an unsupported type %q", p.Host, p.Port, p.Type)}
}

// wrapSocksErr — a proxy-specific failure becomes WebProxyError (the
// "proxy host:port connection failed: …" reason shape).
func wrapSocksErr(p *Proxy, err error) error {
	if err == nil {
		return nil
	}
	if _, ok := err.(*WebProxyError); ok {
		return err
	}
	return &WebProxyError{msg: fmt.Sprintf(
		"proxy %s:%d connection failed: %v", p.Host, p.Port, err)}
}

// socks4aConnect opens a socket to addr THROUGH a SOCKS4/4a proxy,
// mirroring PySocks with rdns=True: an IP literal speaks plain SOCKS4;
// a hostname rides the SOCKS4a extension (0.0.0.1 + the name — DNS at
// the exit).
func socks4aConnect(proxyAddr, username, host string, port int) (net.Conn, error) {
	conn, err := net.DialTimeout("tcp", proxyAddr, 15*time.Second)
	if err != nil {
		return nil, err
	}
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))
	var req []byte
	req = append(req, 0x04, 0x01)
	var portBytes [2]byte
	binary.BigEndian.PutUint16(portBytes[:], uint16(port))
	req = append(req, portBytes[:]...)
	if ip := net.ParseIP(host); ip != nil && ip.To4() != nil {
		req = append(req, ip.To4()...) // plain SOCKS4: the resolved IP
	} else {
		req = append(req, 0, 0, 0, 1) // SOCKS4a: the non-zeros placeholder
	}
	req = append(req, username...)
	req = append(req, 0)
	if ip := net.ParseIP(host); !(ip != nil && ip.To4() != nil) {
		req = append(req, host...)
		req = append(req, 0)
	}
	if _, err := conn.Write(req); err != nil {
		_ = conn.Close()
		return nil, err
	}
	var resp [8]byte
	if _, err := readFull(conn, resp[:]); err != nil {
		_ = conn.Close()
		return nil, err
	}
	if resp[1] != 90 { // 0x5A = request granted
		_ = conn.Close()
		return nil, fmt.Errorf("socks4 proxy refused the connection (code %d)", resp[1])
	}
	_ = conn.SetDeadline(time.Time{})
	return conn, nil
}

// readFull — io.ReadFull without the io import cycle in this file's
// comment voice (plain helper).
func readFull(conn net.Conn, buf []byte) (int, error) {
	total := 0
	for total < len(buf) {
		n, err := conn.Read(buf[total:])
		total += n
		if err != nil {
			return total, err
		}
	}
	return total, nil
}
