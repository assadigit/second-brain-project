// Package guibridge is the INTERIM SEAM (T-70, Phase 6's entry) between the
// nested webui/ Wails module and this module's engine packages.
//
// Why it exists (the honest record, corrected 2026-10-19 at T-70's landing):
// the nested webui module exists because the main module's runtime
// dependency set must stay exactly the authorized four (ADR-0002 Am.3) —
// and the PoC's ONE REAL JOB must run REAL engine code inside the shell.
// ADR-0002 Am.3 originally claimed Go's internal-package rule FORCED a
// non-internal seam; the landing's boundary probe disproved that (a direct
// internal/ import from the nested module builds and vets clean — see
// Am.4). The seam is kept for its architectural value: it is doc 04 §2's
// event-seam shape (LogFn/ProgressFn callbacks + one command method) at PoC
// scale, its composition + contract tests live in the MAIN battery, and it
// evolves into internal/app's real bindings surface at T-71.
//
// THE RETIREMENT LAW: this package RETIRES at T-72, when webui/ dissolves
// into the main module (the wails require enters the main go.mod via its own
// ADR-0002 amendment) and webui/bindings.go imports internal/app directly.
// Nothing else may grow dependencies on it.
package guibridge

import (
	"fmt"
	"os"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/notestate"
)

// LogFn mirrors the engine's optional log callback (notestate.LogFn):
// log(msg, level). nil means no-op.
type LogFn func(msg, level string)

// ProgressFn reports job progress: pos of total items done. nil means no-op.
type ProgressFn func(pos, total int)

// BaselineResult is the note-state baseline job's report.
type BaselineResult struct {
	// Existing is true when the vault already had note-state rows — the
	// one-time baseline law (record_baseline_if_empty's no-op path).
	Existing bool
	// Managed/Unmanaged are the scan walk's counts; Recorded is the number
	// written (== Managed on the recording path, 0 on the existing path).
	Managed   int
	Unmanaged int
	Recorded  int
}

// progressLogEvery is the desk's per-note log cadence — shell-side
// instrumentation only: the reference's own record_baseline_if_empty logs
// just the final sentence (note_state.py:443-474).
const progressLogEvery = 50

// RunNoteStateBaseline drives the REAL engine composition of
// record_baseline_if_empty (note_state.py:443-474 — the same beat the batch
// engine's setup family runs): the ScanVault walk, the managed-note filter,
// then per-note read + ComputeFingerprint + DB upsert, recording ONLY when
// the vault has no note-state rows yet (the one-time law).
//
// The WORK is the engine's own exported functions, composed exactly as
// notestate.RecordBaselineIfEmpty composes them; the progress/log callbacks
// are the GUI-side instrumentation this seam exists to provide. The final
// log line is the reference's own sentence, byte-identical.
//
// dbPath must be a real path — "" is refused (THE DSN-EMPTY-PATH HAZARD,
// SESSION-030's law; anchor the path on the caller's side).
func RunNoteStateBaseline(vault, vaultPath, dbPath string, log LogFn, progress ProgressFn) (*BaselineResult, error) {
	if dbPath == "" {
		return nil, fmt.Errorf("guibridge: dbPath must not be empty (the DSN-empty-path hazard; anchor a real path on the caller's side)")
	}
	if log == nil {
		log = func(string, string) {}
	}
	if progress == nil {
		progress = func(int, int) {}
	}
	db, err := notestate.New(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	n, err := db.Count(vault)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		return &BaselineResult{Existing: true}, nil
	}
	notes := notestate.ScanVault(vaultPath, nil)
	managed := make([]notestate.Note, 0, len(notes))
	unmanaged := 0
	for _, nt := range notes {
		if nt.Unmanaged {
			unmanaged++
			continue
		}
		managed = append(managed, nt)
	}
	log(fmt.Sprintf("scan walk complete: %d managed notes, %d unmanaged", len(managed), unmanaged), "info")
	for i, nt := range managed {
		content := ""
		if data, err := os.ReadFile(nt.Path); err == nil {
			content = helpers.DecodeUTF8Replace(data)
		}
		if err := db.Upsert(vault, nt.SourceURL, nt.Path,
			notestate.ComputeFingerprint(content), nt.Category, nt.Subcategory,
			false); err != nil {
			return nil, err
		}
		progress(i+1, len(managed))
		if (i+1)%progressLogEvery == 0 || i+1 == len(managed) {
			log(fmt.Sprintf("baseline %d/%d notes recorded", i+1, len(managed)), "info")
		}
	}
	if len(managed) > 0 {
		log(fmt.Sprintf("🗂️ Note-state baseline recorded: %d notes (%s vault)", len(managed), vault), "info")
	}
	return &BaselineResult{Managed: len(managed), Unmanaged: unmanaged, Recorded: len(managed)}, nil
}
