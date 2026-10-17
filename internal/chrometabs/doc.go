// Package chrometabs ports core/chrome_tabs.py — THE FIFTH DOOR
// (chrome_tabs.py:1-2583, the T-44 range 484-2563): the owner's real
// Chrome fetches walled pages.
//
// The ladder: ATTACH (his running Chrome's DevTools port — his
// profile, his cookies; the door never closes his browser) → HIS
// PROFILE through the attach link when his Chrome is closed → THE
// DOOR'S OWN identities on two legs (direct first, then the owner's
// proxy, per-site route memory) — every page WAITED FOR (complete at
// an http(s) URL, still for the settle, real by the page_is_real
// verdict) and delivered into the fourth door's folder.
//
// Hand-rolled everywhere the reference is: the RFC 6455 WebSocket
// frame codec and the DevTools socket ride raw net + crypto/rand +
// encoding/base64 (the reference hand-rolls them on Python's socket
// module — no websocket dependency, per the dependency law; the spec's
// old "gorilla/websocket" Go-note is superseded, see the spec's
// SESSION-019 amendment).
//
// Spec: docs/port-blueprint/module-specs/spec-core-chrome-tabs.md;
// task T-44 (Phase 4).
package chrometabs
