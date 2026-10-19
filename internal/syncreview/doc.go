// Package syncreview ports core/sync_review.py (v0.86.0 THE DEFERRED
// LEDGER — the Sync view's brain): the four-state cell model (Match /
// Behind / Reviewed / Unknown), the banner verdict (worst-first, with
// v0.91.0's THE NAMED VERDICT + the setup/measure split), the
// micro-copy caption law, the relative "added" clock, the chip
// grammar, and the DeferredLog itself — a local SQLite ledger that
// remembers every review so the same gap never cries twice. Plus the
// three read-only doors the gather composes: the published index GET
// (the one network door — GET-only, UA worn, never fails loudly), the
// Websites vault's public site-note walk, and the vault's git backup
// gap.
//
// The owner's law, verbatim: "A gap is no longer just 'bad.' It now
// has two outcomes: it gets fixed, or it gets reviewed and explained."
//
// Zero Qt, zero wails — hermetic, headless, CI-safe; the views
// (webui's Sync page) only paint and forward. THE SQLITE LAW
// (cache_db.py's own): every connection rides the store.Open pattern
// (busy_timeout 30000 + WAL + synchronous NORMAL via the DSN,
// MaxOpenConns(1)) and every statement runs under the ledger's own
// mutex; close is idempotent.
//
// Spec: docs/port-blueprint/module-specs/spec-core-sync-review.md;
// task T-75 (Phase 6, landed SESSION-038 — the placeholder seat T-71
// recorded finally filled).
package syncreview
