// Package notebuilder ports app/gitcurator/core/note_builder.py — the
// GitHub-pipeline Obsidian note (YAML frontmatter + body) built from
// LLM/GitHub data with every YAML-bound value treated as hostile input,
// plus the _missing/ 404 placeholder note.
//
// Spec: docs/port-blueprint/module-specs/spec-core-note-builder.md;
// task T-15 (Phase 1). See notebuilder.go's package-notes block for the
// documented port decisions (int numeric fields, any-typed list-or-scalar
// fields, Python's UNICODE \s collapse, FormatComma/PyStr promotions,
// local dates).
package notebuilder
