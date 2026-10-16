package connectioncheck

// netctx_test.go — the verify_ssl split law (netctx.py's own
// contract): default OFF (the historical censored-network setting),
// switchable ON; the websites fetcher is unaffected by this flag
// (webfetch's own always-verify law — see internal/webfetch/doc.go).

import (
	"crypto/tls"
	"testing"
)

func TestVerifySSLEnabledDefaultFalse(t *testing.T) {
	if VerifySSLEnabled(map[string]any{}) {
		t.Error("absent verify_ssl must default to false")
	}
	if VerifySSLEnabled(nil) {
		t.Error("nil config must default to false")
	}
	if VerifySSLEnabled(map[string]any{"verify_ssl": false}) {
		t.Error("explicit false must stay false")
	}
}

func TestVerifySSLEnabledTrueWhenSet(t *testing.T) {
	cfg := map[string]any{"verify_ssl": true}
	if !VerifySSLEnabled(cfg) {
		t.Error("true must be honored")
	}
	// truthiness law: any truthy value counts (Python bool())
	if !VerifySSLEnabled(map[string]any{"verify_ssl": 1}) {
		t.Error("truthy 1 must count")
	}
	if VerifySSLEnabled(map[string]any{"verify_ssl": ""}) {
		t.Error("falsy '' must not count")
	}
}

func TestOutboundTLSConfigDisabledByDefault(t *testing.T) {
	// verify_ssl false/absent → verification disabled (the pre-v0.26.0
	// behavior, preserved by default: check_hostname=False +
	// CERT_NONE; Go's InsecureSkipVerify covers both)
	ctx := OutboundTLSConfig(map[string]any{})
	if !ctx.InsecureSkipVerify {
		t.Error("default config verifies certificates — the historical behavior is broken")
	}
	ctx2 := OutboundTLSConfig(nil)
	if !ctx2.InsecureSkipVerify {
		t.Error("nil config must disable verification")
	}
}

func TestOutboundTLSConfigVerifiesWhenEnabled(t *testing.T) {
	ctx := OutboundTLSConfig(map[string]any{"verify_ssl": true})
	if ctx.InsecureSkipVerify {
		t.Error("verify_ssl=true must verify certificates")
	}
	// a zero tls.Config verifies (chain + hostname) — the Go law the
	// reference's create_default_context mirrors
	if !(&tls.Config{}).InsecureSkipVerify == false {
		// tautological sanity: zero config verifies
		_ = ctx
	}
}
