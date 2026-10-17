// Package mirror ports core/mirror.py + mirror_keys.py + linking.py +
// embeddings.py + recall.py — the read-only Library/ mirror of
// machine-library notes into the Manual Notes vault, the local
// embedding layer, the link-suggestion lifecycle (pending → approved /
// rejected, never re-suggested), the Related (auto) blocks written
// into mirror copies only, and the recall hooks.
//
// Spec: docs/port-blueprint/module-specs/spec-core-mirror.md (the
// SESSION-025 shipped record + Go-note amendments); task T-50
// (Phase 4) — LANDED: the whole merged group (mirror.py +
// mirror_keys.py + linking.py + embeddings.py + recall.py), 90 tests
// + the 39-section golden, 1826/1826 race-clean at the close. The
// Python mirror_keys.py leaf (the import-cycle dodge) folds into this
// one package — Go has no cycle to dodge.
//
// House laws this port obeys (the accumulated record):
//   - CPython \s is Unicode-aware — every ported pattern carrying a
//     Python \s uses the house pyS class; Go's own \s stays out.
//   - Python text-mode reads translate universal newlines BEFORE any
//     cap or parse; read(n) counts CHARACTERS (code points).
//   - str.splitlines splits on the Unicode boundary set (\v, \f,
//     \x1c-\x1e, \x85, U+2028/9) — pySplitLines is the twin.
//   - The clock seam: every time.strftime write rides clockNow (the
//     golden freezes it on both sides).
//   - Go map ranges are RANDOM — every order the reference derives
//     from dict order rides an explicit slice (VaultPath, VecEntry,
//     the LinkStore pair canonicalization's insertion order).
//   - The reference's per-note error strings render
//     f"{type(e).__name__}: {e}" (Python exception classes); the Go
//     port renders the error's own text — the pinned contract is the
//     SHAPE (the entry degrades, the run continues), per the T-40 law.
//   - str.strip('"\”) is a character-SET strip (StripChars, never a
//     prefix trim).
//   - TEMP VAULTS/TREES ONLY in tests — never the owner's real data;
//     the embed endpoints are exercised against loopback stubs only.
package mirror
