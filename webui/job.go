// job.go — THE ONE REAL JOB of the T-70 gate, wired through the guibridge
// interim seam: the note-state baseline (the record_baseline_if_empty
// composition, note_state.py:443-474 — the same beat the batch engine's
// setup family runs) over a generated fixture vault, with the engine's own
// log lines and per-note progress flowing to the desk and the progress row.
// Plus the 10k-line flood leg (the desk's throughput gate).
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/assadigit/second-brain-project/guibridge"
	"github.com/assadigit/second-brain-project/webui/deskbus"
	"github.com/assadigit/second-brain-project/webui/fixture"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// fixtureNotes is the PoC-scale vault (the reference's real vault is 1,309
// notes — doc 07 §B.4; 800 is the honest PoC-scale twin).
const fixtureNotes = 800

// RunBaselineJob is THE real job: generate the fixture vault, run the
// engine's baseline composition over it (a REAL vault walk + per-note
// read + sha256 fingerprint + SQLite upserts — not a synthetic loop),
// streaming the engine's log lines to the desk and per-note progress to
// the progress row. Returns the job's summary JSON (also folded into
// Stats()).
func (a *App) RunBaselineJob() string {
	if !a.jobRunning.CompareAndSwap(false, true) {
		a.logf("warning", "job already running — one job at a time (the batch law)")
		return `{"ok":false,"error":"already running"}`
	}
	defer a.jobRunning.Store(false)

	vaultRoot, err := os.MkdirTemp("", "gc-t70-vault-")
	if err != nil {
		return a.jobFailed("fixture vault: " + err.Error())
	}
	defer os.RemoveAll(vaultRoot)
	dbDir, err := os.MkdirTemp("", "gc-t70-db-")
	if err != nil {
		return a.jobFailed("db dir: " + err.Error())
	}
	defer os.RemoveAll(dbDir)
	dbPath := filepath.Join(dbDir, "cache.db")

	v, err := fixture.WriteVault(vaultRoot, fixtureNotes)
	if err != nil {
		return a.jobFailed("fixture: " + err.Error())
	}
	a.logf("info", "baseline job started: scanning the fixture vault (%d notes + %d stubs)",
		v.Notes, v.Unmanaged)

	t := time.Now()
	res, err := guibridge.RunNoteStateBaseline("github", vaultRoot, dbPath,
		func(msg, level string) { a.logf(level, "%s", msg) },
		func(pos, total int) {
			runtime.EventsEmit(a.ctx, "job:progress", map[string]int{"pos": pos, "total": total})
		})
	d := time.Since(t)
	if err != nil {
		return a.jobFailed("baseline: " + err.Error())
	}

	summary := map[string]any{
		"ok":          true,
		"notes":       res.Recorded,
		"unmanaged":   res.Unmanaged,
		"existing":    res.Existing,
		"ms":          d.Milliseconds(),
		"ms_per_note": float64(d.Microseconds()) / 1000.0 / float64(fixtureNotes),
	}
	b, _ := json.Marshal(summary)
	a.jobSummary = string(b)
	a.logf("success", "baseline job finished in %s (%.2f ms/note, %d notes recorded, %d unmanaged never adopted)",
		d.Round(time.Millisecond), summary["ms_per_note"], res.Recorded, res.Unmanaged)
	return a.jobSummary
}

func (a *App) jobFailed(reason string) string {
	a.logf("error", "baseline job failed: %s", reason)
	summary := map[string]any{"ok": false, "error": reason}
	b, _ := json.Marshal(summary)
	a.jobSummary = string(b)
	return a.jobSummary
}

// FloodLines pushes n lines through the coalescing bus as fast as the
// producer can (backpressure blocks, never drops) and measures the accept
// rate — the 10k-line flood gate's Go half. The frontend's event-loop
// probe (ReportFloodUI) measures the other half: the UI staying
// interactive while the batches land.
func (a *App) FloodLines(n int) string {
	if n <= 0 {
		n = 10000
	}
	a.floodIngested.Store(0)
	a.floodDropped.Store(0)
	base := a.bus.Ingested()
	runtime.EventsEmit(a.ctx, "flood:start", map[string]int{"lines": n})
	t := time.Now()
	for i := 1; i <= n; i++ {
		ok := a.bus.Push(deskLogLine(i, n))
		if !ok {
			break // the bus stopped (shutdown) — count what landed
		}
	}
	accept := time.Since(t)
	dropped := a.bus.Dropped()
	accepted := a.bus.Ingested() - base - dropped
	a.floodIngested.Store(a.bus.Ingested())
	a.floodDropped.Store(dropped)
	// let the desk drain what it owes (the 120 ms beats land the tail;
	// three beats cover the last partial batch), THEN close the probe
	// window so the handler-time evidence covers the whole drain.
	time.Sleep(3 * deskbus.FlushInterval)
	runtime.EventsEmit(a.ctx, "flood:done", nil)
	perS := 0.0
	if accept.Milliseconds() > 0 {
		perS = float64(accepted) / (float64(accept.Microseconds()) / 1e6)
	}
	stats := map[string]any{
		"ok":                 dropped == 0,
		"lines":              n,
		"ingested":           accepted,
		"dropped":            dropped,
		"accept_ms":          accept.Milliseconds(),
		"accept_lines_per_s": perS,
	}
	b, _ := json.Marshal(stats)
	a.floodStats = string(b)
	a.logf("success", "flood: %d lines accepted in %dms (%.0f lines/s, drops: %d) — the desk must stay interactive",
		accepted, accept.Milliseconds(), perS, dropped)
	return a.floodStats
}

// deskLogLine builds flood line i's desk row (level mixed in tranches so
// the chips + the filter re-render have something true to filter).
func deskLogLine(i, n int) deskbus.LogLine {
	level := "info"
	switch {
	case i%97 == 0:
		level = "warning"
	case i%251 == 0:
		level = "error"
	case i%53 == 0:
		level = "success"
	}
	msg := fmt.Sprintf("flood line %d of %d — the desk must stay interactive", i, n)
	return deskbus.LogLine{TS: time.Now().Format("15:04:05"), Level: level, Msg: msg}
}
