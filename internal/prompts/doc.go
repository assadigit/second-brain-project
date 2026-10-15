// Package prompts ports app/gitcurator/core/prompts.py — the {{SLOT}}
// template loader with the refusal laws: no unfilled slots, no None/empty
// values, no slot markers smuggled inside values (untrusted-input guard).
//
// The seven live prompt templates ship embedded in assets/prompts
// (generated copies; the dead 01/02/03 prompts are not shipped). See
// prompts.go for the two structural Go decisions (embed vs. PROMPTS_DIR;
// ntpath filename grammar).
//
// Spec: docs/port-blueprint/module-specs/spec-core-prompts.md;
// task T-13 (Phase 1).
package prompts
