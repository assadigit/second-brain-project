// Package app is the composition layer (doc 04 §2/§5 — the Qt
// replacement seam): the App struct (config, stores, the typed event
// bus, the injected clients), the command surface both the GUI and
// the CLI drive, the ask-gates, and the lifecycle (single-instance
// lock, the one-batch-at-a-time op lock + watchdog, startup passes,
// shutdown graces).
//
// It is the production evolution of the guibridge/ seam (T-70, the
// Phase-6 entry gate's real-job wiring) and webui/deskbus (the
// coalescing log desk's Go half); guibridge RETIRES at T-72, when
// webui/ dissolves into the main module and binds internal/app
// directly.
//
// Spec: docs/port-blueprint/04-target-go-architecture.md §2/§5;
// task T-71 (Phase 6). Slice (a) — the typed event bus + the
// ask-gates (events.go) — is landed; the command surface, the
// lifecycle/locks, and the DI root land in this task's later slices.
package app
