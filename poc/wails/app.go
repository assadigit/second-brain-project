package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sync/atomic"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ---------------------------------------------------------------------------
// The coalescing log bus — the ui.py:65-86 law, ported shape-for-shape:
// producer-side channel (cap 1024, doc 04 §6), consumer-side 120 ms ticker
// that flushes pending rows in insert batches of ≤ 250 (the documented
// batch cap), each batch delivered to the frontend as one event.
// ---------------------------------------------------------------------------

const (
	flushInterval = 120 * time.Millisecond
	batchRows     = 250
	busCap        = 1024
)

// LogLine is one desk-log row (mirrors the LogView row shape: timestamp,
// level, message).
type LogLine struct {
	TS    string `json:"ts"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

type LogBus struct {
	ch       chan LogLine
	ingested atomic.Int64
	dropped  atomic.Int64
	started  atomic.Bool
}

func newLogBus() *LogBus {
	return &LogBus{ch: make(chan LogLine, busCap)}
}

func (b *LogBus) start(ctx context.Context) {
	if !b.started.CompareAndSwap(false, true) {
		return
	}
	go func() {
		ticker := time.NewTicker(flushInterval)
		defer ticker.Stop()
		pending := make([]LogLine, 0, batchRows)
		for {
			select {
			case <-ctx.Done():
				return
			case line := <-b.ch:
				pending = append(pending, line)
				// batch cap → flush early (the 250-row law)
				if len(pending) >= batchRows {
					emitBatch(ctx, pending)
					pending = pending[:0]
				}
			case <-ticker.C:
				if len(pending) == 0 {
					continue
				}
				emitBatch(ctx, pending)
				pending = pending[:0]
			}
		}
	}()
}

func emitBatch(ctx context.Context, rows []LogLine) {
	if len(rows) == 0 {
		return
	}
	runtime.EventsEmit(ctx, "log:batch", rows)
}

// push accepts one line, blocking (with ctx escape) when the bus is
// full — backpressure instead of silent drops: the Python desk never
// lost lines (unbounded queued signals), so the PoC keeps that honesty;
// the drop counter stays for the ctx-cancelled case.
func (b *LogBus) push(ctx context.Context, line LogLine) bool {
	b.ingested.Add(1)
	select {
	case b.ch <- line:
		return true
	case <-ctx.Done():
		b.dropped.Add(1)
		return false
	}
}

// ---------------------------------------------------------------------------
// App — the T-70 surface: themed window (tokens in the CSS), log desk fed
// by the bus, ONE worker job, the flood button, and honest measurements.
// ---------------------------------------------------------------------------

type App struct {
	ctx           context.Context
	t0            time.Time // process start
	startupAt     time.Time // OnStartup callback
	firstFrame    time.Time // MarkFirstFrame() from the frontend
	firstFrameSet bool
	bus           *LogBus

	jobRunning atomic.Bool
	jobResult  string
	floodStats string
}

func NewApp(t0 time.Time) *App {
	return &App{t0: t0, bus: newLogBus()}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	a.startupAt = time.Now()
	a.bus.start(ctx)
	a.logf("info", "wails startup callback at %s after process start", a.startupAt.Sub(a.t0))
	if os.Getenv("SB_POC_AUTORUN") != "" {
		go a.autorun()
	}
}

// autorun drives the whole PoC without human interaction (headless
// evidence mode): first frame has landed → the one worker job → the 10k
// flood → results file. Enabled by SB_POC_AUTORUN=1; results land in
// SB_POC_RESULTS (default /tmp/t70-results.json — outside the repo).
func (a *App) autorun() {
	time.Sleep(3 * time.Second) // let the first frame land
	a.logf("info", "autorun: starting the worker job leg")
	a.RunJob(50)
	time.Sleep(500 * time.Millisecond)
	a.logf("info", "autorun: starting the 10k flood leg")
	a.FloodLines(10000)
	// let the coalesced batches drain into the desk (120 ms cadence)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		if a.bus.ingested.Load()-a.bus.dropped.Load() >= 10000 && time.Since(a.startupAt) > 12*time.Second {
			break
		}
		time.Sleep(500 * time.Millisecond)
	}
	time.Sleep(2 * time.Second)
	a.logf("success", "autorun complete — see the results file for measurements")
	_ = os.WriteFile(envOr("SB_POC_RESULTS", "/tmp/t70-results.json"), []byte(a.Stats()), 0o600)
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func (a *App) logf(level, format string, args ...any) {
	_ = a.bus.push(a.ctx, LogLine{
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

// RunJob is THE one worker job of the PoC: a deliberately sequential loop
// (the batch-sequential law, doc 04 §6) with per-item work, progress
// events and desk-log lines.
func (a *App) RunJob(items int) string {
	if items <= 0 {
		items = 50
	}
	if !a.jobRunning.CompareAndSwap(false, true) {
		a.logf("warning", "job already running — one job at a time (the batch law)")
		return `{"ok":false,"error":"already running"}`
	}
	defer a.jobRunning.Store(false)

	t := time.Now()
	a.logf("info", "worker job started: %d items, sequential, 20 ms/item", items)
	for i := 1; i <= items; i++ {
		// the "work" — paced so progress is observable
		time.Sleep(20 * time.Millisecond)
		if i%10 == 0 || i == items {
			a.logf("info", "item %d/%d processed", i, items)
		}
		runtime.EventsEmit(a.ctx, "job:progress", map[string]int{"pos": i, "total": items})
	}
	d := time.Since(t)
	summary := map[string]any{
		"ok": true, "items": items, "ms": d.Milliseconds(),
		"ms_per_item": float64(d.Milliseconds()) / float64(items),
	}
	b, _ := json.Marshal(summary)
	a.jobResult = string(b)
	a.logf("success", "worker job finished in %s (%.1f ms/item)", d.Round(time.Millisecond), float64(d.Milliseconds())/float64(items))
	return a.jobResult
}

// FloodLines pushes n lines through the coalescing bus as fast as the
// producer can and measures the accept rate + drop count (the 10k-line
// flood gate: the desk must stay interactive).
func (a *App) FloodLines(n int) string {
	if n <= 0 {
		n = 10000
	}
	t := time.Now()
	for i := 1; i <= n; i++ {
		a.bus.push(a.ctx, LogLine{
			TS:    time.Now().Format("15:04:05"),
			Level: "info",
			Msg:   fmt.Sprintf("flood line %d — the desk must stay interactive", i),
		})
	}
	accepted := time.Since(t)
	b, _ := json.Marshal(map[string]any{
		"ok":                 true,
		"lines":              n,
		"accept_ms":          accepted.Milliseconds(),
		"ingested":           a.bus.ingested.Load(),
		"dropped":            a.bus.dropped.Load(),
		"accept_lines_per_s": float64(n) / (float64(accepted.Milliseconds()) / 1000.0),
	})
	a.floodStats = string(b)
	a.logf("success", "flood: %d lines accepted in %s (drops so far: %d) — flushing at the 120 ms cadence",
		n, accepted.Round(time.Millisecond), a.bus.dropped.Load())
	return a.floodStats
}

// Stats returns the PoC measurements as one JSON blob (the evidence line).
func (a *App) Stats() string {
	b, _ := json.Marshal(map[string]any{
		"startup_to_first_frame_ms": a.firstFrame.Sub(a.t0).Milliseconds(),
		"startup_callback_ms":       a.startupAt.Sub(a.t0).Milliseconds(),
		"last_job":                  json.RawMessage(orNull(a.jobResult)),
		"last_flood":                json.RawMessage(orNull(a.floodStats)),
	})
	return string(b)
}

func orNull(s string) string {
	if s == "" {
		return "null"
	}
	return s
}

func passFail(ok bool) string {
	if ok {
		return "PASS"
	}
	return "FAIL"
}
