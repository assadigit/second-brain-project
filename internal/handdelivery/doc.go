// Package handdelivery ports core/hand_delivery.py — THE FOURTH DOOR:
// the owner's own Chrome (v0.48.0): walled links queued for hand
// delivery (JSON queue + README in the vault, optional open-in-Chrome)
// and the consume side turning delivered .html files into real
// FetchResults (with re-verification of app-delivered pages and
// master-table gesture stamping).
//
// Spec: docs/port-blueprint/module-specs/spec-core-hand-delivery.md;
// task T-43 (Phase 4) — T-40 floored the consumption surface, T-43
// completed the module (the Chrome launch family, the pickers, the
// batch consume, the queue stamp, the module battery + the generated
// golden).
package handdelivery
