// Package config ports the configuration doors: the user memory
// (core/user_config.py — Documents/.gitcurator: THE FILL and THE MIRROR,
// the unset grammar incl. numeric-0-as-unset, the path law) plus the
// config.json load/save layer ON TOP of the T-12 storage primitives
// (storage.OrderedMap/DecodeJSON/WriteConfigFile and
// constants.ExampleTree — no second tree representation).
//
// Spec: docs/port-blueprint/module-specs/spec-core-user-config.md +
// doc 08 §1; task T-16 (Phase 1). Parity is pinned by the generated
// golden in testdata/ (fill/mirror matrices, the owner round trip, the
// ordered example bytes) and the ported test_userconfig suite.
//
// Documented port decisions: the Windows Known-Folder call uses the raw
// syscall (mirroring the reference's own ctypes approach — Phase 1 stays
// stdlib-pure; x/sys remains the doc-06 option); the fill's change
// detection compares with Python == semantics (numeric cross-type,
// order-insensitive maps); a non-object config.json recovers to the
// fresh door where Python would carry the garbage (typed-tree
// constraint, strictly better user outcome — see config.go).
package config
