// app.go — the T-70 shell surface: the desk bus wired to the wails event
// bridge, the cold-start clock, THE ONE REAL JOB (RunBaselineJob — the
// guibridge note-state baseline over a generated fixture vault), the 10k
// flood leg, the UI probe report, and the honest measurement file
// (autorun mode — the same evidence discipline as SESSION-001's PoC).
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/assadigit/second-brain-project/webui/deskbus"
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// App is the bound surface (frontend/wailsjs/go/main/App.js mirrors it).
type App struct {
	ctx           context.Context
	t0            time.Time // process start
	startupAt     time.Time // OnStartup callback
	firstFrame    time.Time // MarkFirstFrame() from the frontend
	firstFrameSet bool
	bus           *deskbus.LogBus

	jobRunning atomic.Bool
	jobSummary string

	floodStats    string
	floodIngested atomic.Int64
	floodDropped  atomic.Int64

	uiProbeMu sync.Mutex
	uiProbe   map[string]any
}

// NewApp builds the shell around the process-start clock.
func NewApp(t0 time.Time) *App {
	return &App{t0: t0, uiProbe: map[string]any{}}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startupAt = time.Now()
	a.bus = deskbus.New(func(rows []deskbus.LogLine) {
		runtime.EventsEmit(ctx, "log:batch", rows)
	})
	a.bus.Start()
	a.logf("info", "wails startup callback at %s after process start", a.startupAt.Sub(a.t0))
	if os.Getenv("SB_POC_AUTORUN") != "" {
		go a.autorun()
	}
}

func (a *App) shutdown(ctx context.Context) {
	// the drain law: buffered lines still land, then the loop exits.
	if a.bus != nil {
		a.bus.Stop()
	}
	_ = ctx
}

// autorun drives the whole gate without human interaction (headless
// evidence mode): first frame → THE REAL JOB → the 10k flood → the
// results file. Enabled by SB_POC_AUTORUN=1; results land in
// SB_POC_RESULTS (default /tmp/t70-gate-results.json — outside the repo).
func (a *App) autorun() {
	time.Sleep(3 * time.Second) // let the first frame land
	a.logf("info", "autorun: starting the real-job leg (the note-state baseline)")
	a.RunBaselineJob()
	time.Sleep(500 * time.Millisecond)
	a.logf("info", "autorun: starting the 10k flood leg")
	a.FloodLines(10000)
	// let the coalesced batches drain into the desk (120 ms cadence)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if a.bus.Ingested()-a.bus.Dropped() >= 10000 && time.Since(a.startupAt) > 12*time.Second {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(2 * time.Second)
	a.logf("success", "autorun complete — see the results file for measurements")
	_ = os.WriteFile(envOr("SB_POC_RESULTS", "/tmp/t70-gate-results.json"), []byte(a.Stats()), 0o600)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (a *App) logf(level, format string, args ...any) {
	_ = a.bus.Push(deskbus.LogLine{
		TS:    time.Now().Format("15:04:05"),
		Level: level,
		Msg:   fmt.Sprintf(format, args...),
	})
}

// MarkFirstFrame is called by the frontend on DOMContentLoaded — the
// cold-start measurement point (PoC gate: ≤ 1.5 s).
func (a *App) MarkFirstFrame() string {
	if !a.firstFrameSet {
		a.firstFrame = time.Now()
		a.firstFrameSet = true
		d := a.firstFrame.Sub(a.t0)
		a.logf("success", "first frame (DOMContentLoaded → MarkFirstFrame) at %s — gate ≤ 1.5 s: %s",
			d.Round(time.Millisecond), passFail(d <= 1500*time.Millisecond))
		return d.String()
	}
	return a.firstFrame.Sub(a.t0).String()
}

func passFail(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}

// ReportFloodUI receives the frontend's event-loop probe for the flood
// window — the "UI stays interactive" evidence, measured (not asserted):
// max task-queue gap + gaps over 100 ms, sampled by a setTimeout chain
// that would starve if the loop blocked.
func (a *App) ReportFloodUI(probe map[string]any) {
	a.uiProbeMu.Lock()
	a.uiProbe = probe
	a.uiProbeMu.Unlock()
}

// ClearDesk is the clear door's Go half: the bus wipes its pending queue
// and buffered lines ("clear wipes cache+view+pending" — the frontend
// wipes its own cache+view around this call).
func (a *App) ClearDesk() string {
	if a.bus != nil {
		a.bus.Clear()
	}
	return "cleared"
}

// Stats is the honest measurement file: every number the gate's verdict
// cites, in one JSON blob (the same discipline as SESSION-001's PoC).
func (a *App) Stats() string {
	a.uiProbeMu.Lock()
	probe := a.uiProbe
	a.uiProbeMu.Unlock()
	out := map[string]any{
		"startup_callback_ms":       a.startupAt.Sub(a.t0).Milliseconds(),
		"startup_to_first_frame_ms": -1,
		"last_job":                  json.RawMessage("null"),
		"last_flood":                json.RawMessage("null"),
		"bus_ingested":              a.bus.Ingested(),
		"bus_dropped":               a.bus.Dropped(),
		"flood_ui_probe":            probe,
	}
	if a.firstFrameSet {
		out["startup_to_first_frame_ms"] = a.firstFrame.Sub(a.t0).Milliseconds()
	}
	if a.jobSummary != "" {
		out["last_job"] = json.RawMessage(a.jobSummary)
	}
	if a.floodStats != "" {
		out["last_flood"] = json.RawMessage(a.floodStats)
	}
	b, err := json.MarshalIndent(out, "", " ")
	if err != nil {
		return `{"ok": false, "error": "marshal: ` + err.Error() + `"}`
	}
	return string(b)
}
