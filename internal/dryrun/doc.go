// Package dryrun ports app/gitcurator/core/dryrun.py — the global
// dry-run switch, the withheld-operation recorder, the filesystem
// stand-ins, the shadow cache copy (WAL sidecars included), and the
// Markdown report renderer.
//
// Spec: docs/port-blueprint/module-specs/spec-core-storage.md §dryrun.py
// (the dryrun contract lives inside that file — the older
// "spec-core-dryrun.md" name was a stale pointer, reality-fixed with
// T-12); task T-12 (Phase 1).
package dryrun
