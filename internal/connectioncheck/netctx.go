package connectioncheck

// netctx.go — netctx.py (v0.26.0): the outbound TLS switch behind one
// explicit config flag. verify_ssl false/absent (the DEFAULT — the
// historical censored-network setting) disables certificate
// verification on the non-websites outbound paths; true enforces it.
// THE SPLIT LAW: the websites fetcher (internal/webfetch) always
// verifies and is not affected by this flag; the LLM/cloud paths take
// their verifyTLS parameter (llmclient) threaded from this switch;
// the Worker-sync and GitHub-metadata callers consume OutboundTLSConfig.

import (
	"crypto/tls"
)

// VerifySSLEnabled reads the explicit verify_ssl flag (default false
// = the historical censored-network behavior).
func VerifySSLEnabled(config any) bool {
	return pyTruthy(dictDefault(config, "verify_ssl", false))
}

// OutboundTLSConfig is a TLS config honoring the verify_ssl flag:
// false/absent → certificate verification disabled (check_hostname
// False + CERT_NONE — Go's InsecureSkipVerify covers both), true →
// certificates verified; a bad cert fails the call.
func OutboundTLSConfig(config any) *tls.Config {
	if VerifySSLEnabled(config) {
		return &tls.Config{}
	}
	return &tls.Config{InsecureSkipVerify: true}
}
