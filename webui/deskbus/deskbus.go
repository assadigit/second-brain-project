// Package deskbus is the coalesced log desk's Go half — THE COALESCED LOG
// contract (spec-gui-ui, ui.py:65-86 + 2374-2794) ported shape-for-shape:
//
//   - log_message NEVER touches the document: the producer pushes into a
//     channel (cap 1024) and returns — the caller's turn costs microseconds
//     whatever the flood;
//   - ONE 120 ms flush beat lands ≤ 250 rows as ONE batch event (the single
//     cursor-edit-block + one-scroll law, become one frontend event);
//   - backpressure instead of silent drops: the Python desk never lost
//     lines; a full channel BLOCKS the producer (the drop counter only
//     counts pushes refused after Stop);
//   - the clear door: "clear wipes cache+view+pending" — Clear SYNCHRONOUSLY
//     drops the pending queue AND the buffered channel content (an ack from
//     the loop makes the wipe ordering deterministic).
//
// Pure Go — no wails imports: the emit function is injected, so the bus is
// testable headless (deskbus_test.go). The wails wiring lives in app.go.
package deskbus

import (
	"sync/atomic"
	"time"
)

// The contract's three numbers (ui.py:76-86): the coalesce interval, the
// flush batch cap, and the producer channel depth.
const (
	FlushInterval = 120 * time.Millisecond
	BatchRows     = 250
	BusCap        = 1024
)

// LogLine is one desk-log row: the [ts] prefix, the level glyph's level,
// and the message (the _log_row_html row shape — glyph · timestamp ·
// message; the glyph is rendered frontend-side from Level).
type LogLine struct {
	TS    string `json:"ts"`
	Level string `json:"level"`
	Msg   string `json:"msg"`
}

// EmitFn delivers ONE flushed batch (1..BatchRows rows, in stream order).
// It must not block (the wails EventsEmit is async; tests record).
type EmitFn func(rows []LogLine)

type clearReq struct{ ack chan struct{} }

// LogBus is the coalescing producer→flush pipeline.
type LogBus struct {
	ch       chan LogLine
	clear    chan clearReq
	emit     EmitFn
	ingested atomic.Int64
	dropped  atomic.Int64
	started  atomic.Bool
	stop     chan struct{}
	done     chan struct{}
}

// New builds a bus that delivers flushed batches to emit.
func New(emit EmitFn) *LogBus {
	return &LogBus{
		ch:    make(chan LogLine, BusCap),
		clear: make(chan clearReq),
		emit:  emit,
		stop:  make(chan struct{}),
		done:  make(chan struct{}),
	}
}

// Start launches the flush loop (idempotent). The loop owns the pending
// slice; the 120 ms ticker flushes partial batches, a full pending slice
// (BatchRows) flushes early, Clear wipes pending+buffered synchronously,
// and Stop drains what is buffered then exits.
func (b *LogBus) Start() {
	if !b.started.CompareAndSwap(false, true) {
		return
	}
	go func() {
		defer close(b.done)
		ticker := time.NewTicker(FlushInterval)
		defer ticker.Stop()
		pending := make([]LogLine, 0, BatchRows)
		flush := func() {
			if len(pending) == 0 {
				return
			}
			b.emit(pending)
			pending = pending[:0]
		}
		drainBuffered := func() {
			// empty the channel WITHOUT emitting (the clear door).
			for {
				select {
				case <-b.ch:
				default:
					return
				}
			}
		}
		for {
			select {
			case line := <-b.ch:
				pending = append(pending, line)
				if len(pending) >= BatchRows {
					// the batch cap → flush early (the 250-row law);
					// rows remaining behind the cap re-arm immediately
					// (the single-shot timer's own re-arm behavior).
					flush()
				}
			case req := <-b.clear:
				// "clear wipes pending": drop the pending queue and the
				// buffered channel content in one motion, then ack — the
				// ack makes the wipe's ordering deterministic (lines
				// pushed after Clear() returns can never be wiped by it).
				pending = pending[:0]
				drainBuffered()
				close(req.ack)
			case <-ticker.C:
				flush()
			case <-b.stop:
				// drain what is already buffered, land it, exit.
				for {
					select {
					case line := <-b.ch:
						pending = append(pending, line)
						if len(pending) >= BatchRows {
							flush()
						}
					default:
						flush()
						return
					}
				}
			}
		}
	}()
}

// Push accepts one line. It BLOCKS while the channel is full — backpressure
// instead of silent drops — and returns false (counting the drop) only when
// the bus has stopped.
func (b *LogBus) Push(line LogLine) bool {
	b.ingested.Add(1)
	select {
	case b.ch <- line:
		return true
	case <-b.stop:
		b.dropped.Add(1)
		return false
	}
}

// Clear drops everything pending and buffered (the reference's
// "clear wipes cache+view+pending" — the Go half; the frontend wipes its
// own cache+view around this call). Lines pushed AFTER Clear returns land
// normally. Synchronous: it returns only after the loop has actioned the
// wipe. After Stop it is a no-op.
func (b *LogBus) Clear() {
	if !b.started.Load() {
		return // never started: nothing ever flowed
	}
	req := clearReq{ack: make(chan struct{})}
	select {
	case b.clear <- req:
		<-req.ack
	case <-b.stop:
		// the loop is gone: nothing left to clear
	case <-b.done:
		// the loop exited (defensive twin of the stop case)
	}
}

// Stop drains the buffered lines (they are still delivered), ends the loop,
// and waits for it. Idempotent.
func (b *LogBus) Stop() {
	if !b.started.Load() {
		return // never started: no loop to stop (and done never closes)
	}
	select {
	case <-b.stop:
	default:
		close(b.stop)
	}
	<-b.done
}

// Ingested returns the total lines accepted-or-refused (the producer-side
// tally; FloodLines' accept rate divides over this).
func (b *LogBus) Ingested() int64 { return b.ingested.Load() }

// Dropped returns the lines refused after Stop (0 under backpressure — the
// desk never silently loses lines while running).
func (b *LogBus) Dropped() int64 { return b.dropped.Load() }
