// shell.go — T-72's production shell: the wails window's Go half,
// binding internal/app DIRECTLY (the Am.4-probed direct import — the
// nested webui module may import the parent's internal/ packages; the
// guibridge/ interim seam RETIRES with this landing).
//
// THE PRODUCTION TRANSPORT (doc 04 §2/§4): the app's typed Bus fans the
// engines' events out; this shell is ONE subscriber — the desk's own
// law (ui.py:65-86): LogLine lands in the deskbus (the 120 ms coalescing
// beat, ≤250-row batches), the progress/status/batch events cross to
// the frontend as wails events. The reference's log_message built the
// timestamp at enqueue time; the twin formats it here, at the subscriber.
//
// THE LIFECYCLE: the single-instance app.lock (the CONTINUE-WITHOUT law
// — a lock failure never kills the app), the startup passes (sentinel,
// llama autodetect, the quiet heal, the backlog notice, the radar) —
// each pass's engine speaks its real lines onto the desk through this
// same transport, and the shutdown drains the deskbus (its buffered
// rows still land) before the graces.
//
// THE PERF DOOR (T-72's acceptance: the log perf test): FloodLines
// pushes its lines through Bus.EmitLogLine — the FULL production chain
// (the fan-out hop included), never a side channel — so the measurement
// is the desk the owner will actually use.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/webui/deskbus"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// Shell is the production window surface (the bound methods are
// mirrored in frontend/wailsjs/go/main/Shell.js by the wails build).
type Shell struct {
	ctx        context.Context
	t0         time.Time // process start — the cold-start clock
	startupAt  time.Time
	firstFrame time.Time

	frameSet atomic.Bool

	theApp       *app.App
	settingsUI   *SettingsUI
	pagesBUI     *PagesBUI
	syncUI       *SyncUI
	bus          *deskbus.LogBus
	cancelBusSub func()

	lock *app.AppLock

	// the flood leg's honest counters (the perf door).
	floodPushed  atomic.Int64
	floodDropped atomic.Int64

	uiProbeMu sync.Mutex
	uiProbe   map[string]any

	// lastFlood is the last flood leg's JSON (the honest record).
	lastFlood string
}

// NewShell builds the production composition around the process-start
// clock: the app over the exe-dir config (constants.ConfigFile — the
// reference's APP_DIR/config.json). headless=false: a gate nobody
// answers fails fast (the no-reader law) — T-72's shell renders no ask
// dialogs yet; the ask-gates' modals land with T-76.
//
// The single-instance lock is taken HERE, before the window exists
// (the reference's own order): on the alive-PID refusal the loud line
// is already printed by AcquireAppLock and this process exits 1 — the
// second window never opens. Every other lock failure CONTINUES
// WITHOUT the lock (the reference's belt).
func NewShell(t0 time.Time) *Shell {
	s := &Shell{t0: t0, uiProbe: map[string]any{}}
	s.lock = app.AcquireAppLock(filepath.Join(constants.AppDir(), "app.lock"))
	if s.lock.Refused {
		os.Exit(1)
	}
	s.theApp = app.New(constants.ConfigFile(), false)
	s.settingsUI = NewSettingsUI(s.theApp.Settings)
	s.pagesBUI = NewPagesBUI(s.theApp.Settings, s.theApp.Commands)
	s.syncUI = NewSyncUI(s.theApp.Settings, s.theApp.Commands)
	return s
}

// startup is the wails OnStartup door.
func (s *Shell) startup(ctx context.Context) {
	s.ctx = ctx
	s.startupAt = time.Now()
	s.settingsUI.bindCtx(ctx)
	s.pagesBUI.bindCtx(ctx)
	s.syncUI.bindCtx(ctx)

	// The desk's Go half: the 120 ms coalescing beat, wired to the
	// bridge (one flushed batch = one log:batch event).
	s.bus = deskbus.New(func(rows []deskbus.LogLine) {
		runtime.EventsEmit(ctx, "log:batch", rows)
	})
	s.bus.Start()

	// THE PRODUCTION TRANSPORT: this shell is one Bus subscriber.
	// The reference's log_message built the [ts] at enqueue time; the
	// twin formats it here. The rest of the typed events cross to the
	// frontend as wails events (the queued-connection law: this handler
	// never blocks — EmitLogLine fans out, the deskbus buffers).
	s.cancelBusSub = s.theApp.Bus.Subscribe(&app.Handlers{
		LogLine: func(ev app.LogLine) {
			s.bus.Push(deskbus.LogLine{
				TS:    time.Now().Format("15:04:05"),
				Level: ev.Level,
				Msg:   ev.Msg,
			})
		},
		Progress: func(ev app.Progress) {
			runtime.EventsEmit(ctx, "bus:progress", ev)
		},
		Status: func(ev app.Status) {
			runtime.EventsEmit(ctx, "bus:status", ev)
		},
		BatchFinished: func(ev app.BatchFinished) {
			runtime.EventsEmit(ctx, "bus:batch-finished", ev)
		},
		JobProgress: func(ev app.JobProgress) {
			runtime.EventsEmit(ctx, "bus:job-progress", ev)
		},
		DiskFull: func(ev app.DiskFull) {
			runtime.EventsEmit(ctx, "bus:disk-full", ev)
		},
		ModelChanged: func(ev app.ModelChanged) {
			runtime.EventsEmit(ctx, "bus:model-changed", ev)
		},
		// The ask-gates: T-72's shell renders no modals — a gate with
		// no reader fails fast per the bus's own law (the headless
		// twin). The modals land with T-76's dialogs; until then the
		// gates decline honestly, nothing hangs.
	})

	// The startup passes — each engine speaks its real lines onto the
	// desk through the transport above (the sentinel's probes, the
	// llama autodetect, the quiet heal, the backlog notice, the radar;
	// all crash-isolated, all at the reference's delays).
	s.theApp.Lifecycle.Start()

	// v0.69.0's one quiet line (window.py:155-173): if THIS launch
	// restored credentials from the user memory, say so — the magic
	// is never silent.
	if line := s.theApp.TheFillLine(); line != "" {
		s.logf("success", "%s", line)
	}

	s.logf("info", "wails startup callback at %s after process start", s.startupAt.Sub(s.t0))
	if os.Getenv("SB_POC_AUTORUN") != "" {
		go s.autorun()
	}
}

// shutdown is the wails OnShutdown door: the deskbus drains first (its
// buffered rows still land), then the lifecycle's graces, then the
// lock goes home.
func (s *Shell) shutdown(ctx context.Context) {
	if s.bus != nil {
		s.bus.Stop()
	}
	if s.cancelBusSub != nil {
		s.cancelBusSub()
	}
	if s.theApp != nil && s.theApp.Lifecycle != nil {
		s.theApp.Lifecycle.Shutdown(nil)
	}
	if s.lock != nil {
		s.lock.Release()
	}
	_ = ctx
}

// logf is the shell's own voice on the desk (the startup line above;
// the engines' lines ride the Bus, never this door).
func (s *Shell) logf(level, format string, args ...any) {
	if s.bus == nil {
		return
	}
	_ = s.bus.Push(deskbus.LogLine{
		TS:    time.Now().Format("15:04:05"),
		Level: level,
		Msg:   fmt.Sprintf(format, args...),
	})
}

// MarkFirstFrame is the frontend's DOMContentLoaded call — the
// cold-start measurement point (the gate law: ≤ 1.5 s).
func (s *Shell) MarkFirstFrame() string {
	if s.frameSet.CompareAndSwap(false, true) {
		s.firstFrame = time.Now()
		d := s.firstFrame.Sub(s.t0)
		s.logf("success", "first frame (DOMContentLoaded → MarkFirstFrame) at %s — gate ≤ 1.5 s: %s",
			d.Round(time.Millisecond), passFail(d <= 1500*time.Millisecond))
		return d.String()
	}
	return s.firstFrame.Sub(s.t0).String()
}

// ClearDesk is the clear door's Go half: the bus wipes its pending
// queue and buffered lines (the frontend wipes its own cache+view
// around this call).
func (s *Shell) ClearDesk() string {
	if s.bus != nil {
		s.bus.Clear()
	}
	return "cleared"
}

// ReportFloodUI receives the frontend's event-loop probe for the flood
// window (the honest measurement discipline — the handler-time
// instrument is the interactivity evidence; the task-gap sampler reads
// the WM-less-Xvfb timer clamp, recorded as such).
func (s *Shell) ReportFloodUI(probe map[string]any) {
	s.uiProbeMu.Lock()
	s.uiProbe = probe
	s.uiProbeMu.Unlock()
}

// FloodLines is THE PERF DOOR (T-72's acceptance: the log perf test):
// n lines through the FULL production chain — Bus.EmitLogLine (the
// fan-out) → this shell's subscriber → the deskbus (the 120 ms beat,
// ≤250-row batches) → the frontend's cache-mirror desk. Never a side
// channel: the measurement is the desk the owner uses. flood:start /
// flood:done bracket the leg for the frontend's probe instruments;
// flood:done fires only after the drain, so the probe sees every
// batch. The accept rate is the PRODUCER-side number (the push loop
// — the backpressure gate the T-70 gate measured); the per-batch
// handler times are the frontend's own probe.
func (s *Shell) FloodLines(n int) string {
	runtime.EventsEmit(s.ctx, "flood:start", nil)
	t0 := time.Now()
	for i := 0; i < n; i++ {
		s.floodPushed.Add(1)
		s.theApp.Bus.EmitLogLine(fmt.Sprintf("flood line %d — the production transport under load", i), "info")
	}
	pushMs := time.Since(t0).Milliseconds()

	// the drain: every buffered row lands before the probe closes; the
	// final flush beat gets one extra tick so the last batch is emitted.
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if s.bus.Ingested() >= int64(n) && s.bus.Dropped() == 0 {
			time.Sleep(deskbus.FlushInterval + 30*time.Millisecond)
			break
		}
		time.Sleep(25 * time.Millisecond)
	}
	drainMs := time.Since(t0).Milliseconds()
	runtime.EventsEmit(s.ctx, "flood:done", nil)

	dropped := s.bus.Dropped()
	s.floodDropped.Store(dropped)
	ingested := s.bus.Ingested()
	linesPerS := 0.0
	if pushMs > 0 {
		linesPerS = float64(n) / (float64(pushMs) / 1000.0)
	}
	out := map[string]any{
		"lines":              n,
		"push_ms":            pushMs,
		"ingested":           ingested,
		"dropped":            dropped,
		"drain_ms":           drainMs,
		"accept_lines_per_s": linesPerS,
		"ok":                 dropped == 0,
	}
	b, err := json.Marshal(out)
	if err != nil {
		return fmt.Sprintf(`{"ok": false, "error": %q}`, err.Error())
	}
	s.lastFlood = string(b)
	return string(b)
}

// Stats is the honest measurement file: every number the perf
// evidence cites, in one JSON blob.
func (s *Shell) Stats() string {
	s.uiProbeMu.Lock()
	probe := s.uiProbe
	s.uiProbeMu.Unlock()
	out := map[string]any{
		"startup_callback_ms":       s.startupAt.Sub(s.t0).Milliseconds(),
		"startup_to_first_frame_ms": -1,
		"bus_ingested":              int64(0),
		"bus_dropped":               int64(0),
		"last_flood":                json.RawMessage("null"),
		"flood_ui_probe":            probe,
	}
	if s.frameSet.Load() {
		out["startup_to_first_frame_ms"] = s.firstFrame.Sub(s.t0).Milliseconds()
	}
	if s.bus != nil {
		out["bus_ingested"] = s.bus.Ingested()
		out["bus_dropped"] = s.bus.Dropped()
	}
	if s.lastFlood != "" {
		out["last_flood"] = json.RawMessage(s.lastFlood)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return `{"ok": false, "error": "marshal: ` + err.Error() + `"}`
	}
	return string(b)
}

// autorun drives the perf leg headless (the evidence mode): first
// frame → the 10k flood through the production transport → the
// results file. The T-70 gate's real-job leg (the note-state baseline
// over the fixture via guibridge) RETIRED with guibridge at this
// landing — T-72's acceptance is the token golden + the log perf test;
// the real jobs return through the app's command surface when the
// pages land (T-73+).
func (s *Shell) autorun() {
	time.Sleep(3 * time.Second) // let the first frame + the startup passes' first lines land
	s.logf("info", "autorun: the 10k flood leg through the production transport (Bus → deskbus → desk)")
	stats := s.FloodLines(10000)
	s.logf("info", "flood: %s", stats)
	time.Sleep(2 * time.Second)
	s.logf("success", "autorun complete — see the results file for measurements")
	// The Settings-pages probe (the T-73 gate leg's visual evidence):
	// open the Settings window for the screenshot, then write the
	// results. The frontend renders the pages off this event; the
	// probe's VALUE picks the section (T-74: the pages B's own
	// evidence door — "backup", "bot", "dashboard", "sources").
	if v := os.Getenv("SB_POC_OPEN_SETTINGS"); v != "" {
		runtime.EventsEmit(s.ctx, "autorun:open-settings", v)
		time.Sleep(1500 * time.Millisecond)
	}
	_ = os.WriteFile(envOr("SB_POC_RESULTS", "/tmp/t70-gate-results.json"), []byte(s.Stats()), 0o600)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func passFail(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
