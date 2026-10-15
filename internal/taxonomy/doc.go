// Package taxonomy ports app/gitcurator/core/taxonomy.py — parses the
// owner-editable Websites category file into the Taxonomy model and
// validates model answers against parsed names before any path is touched.
//
// Spec: docs/port-blueprint/module-specs/spec-core-taxonomy.md;
// task T-14 (Phase 1). See taxonomy.go's package-notes block for the
// documented port decisions (splitlines boundaries, Unicode strip,
// platform-separator join, BOM-sig read, UnicodeDecodeError escape).
package taxonomy
