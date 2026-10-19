// deskbus_test.go — the coalescing bus's own laws, headless (the emit
// function is injected; no wails, no window): the 250-row early flush, the
// 120 ms partial-batch beat, flood order + zero drops under backpressure,
// the stop-drain, and the post-stop refusal.
package deskbus

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

// recorder collects emitted batches concurrently-safely.
type recorder struct {
	mu      sync.Mutex
	batches [][]LogLine
}

func (r *recorder) emit(rows []LogLine) {
	cp := make([]LogLine, len(rows))
	copy(cp, rows)
	r.mu.Lock()
	r.batches = append(r.batches, cp)
	r.mu.Unlock()
}

func (r *recorder) all() []LogLine {
	r.mu.Lock()
	defer r.mu.Unlock()
	var out []LogLine
	for _, b := range r.batches {
		out = append(out, b...)
	}
	return out
}

func (r *recorder) batchSizes() []int {
	r.mu.Lock()
	defer r.mu.Unlock()
	sizes := make([]int, len(r.batches))
	for i, b := range r.batches {
		sizes[i] = len(b)
	}
	return sizes
}

func line(i int) LogLine {
	return LogLine{TS: "00:00:00", Level: "info", Msg: fmt.Sprintf("line %d", i)}
}

func TestBatchCapFlushesEarly(t *testing.T) {
	// 250 rows: the batch cap flushes WITHOUT waiting for the 120 ms tick
	// — the flush must land well inside one coalesce interval. (The LAW
	// asserted is delivery + the ≤250 batch cap: a scheduler-noise split
	// into two legal partial batches stays within the contract; the
	// latency claim is the poll deadline itself.)
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	defer bus.Stop()
	for i := 0; i < 250; i++ {
		bus.Push(line(i))
	}
	deadline := time.Now().Add(80 * time.Millisecond) // < FlushInterval
	for time.Now().Before(deadline) {
		if got := rec.all(); len(got) == 250 {
			for _, size := range rec.batchSizes() {
				if size > BatchRows {
					t.Fatalf("batch of %d rows exceeds the %d cap", size, BatchRows)
				}
			}
			if got[0].Msg != "line 0" || got[249].Msg != "line 249" {
				t.Fatalf("order broken: %q … %q", got[0].Msg, got[249].Msg)
			}
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatalf("250 rows not flushed within 80ms (the early-flush latency): batches=%v", rec.batchSizes())
}

func TestTickerFlushesPartialBatches(t *testing.T) {
	// 3 rows: below the cap, so only the 120 ms beat can land them — all
	// three delivered within a few beats, in order, every batch ≤ 250.
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	defer bus.Stop()
	for i := 0; i < 3; i++ {
		bus.Push(line(i))
	}
	deadline := time.Now().Add(3 * FlushInterval)
	for time.Now().Before(deadline) {
		if got := rec.all(); len(got) == 3 {
			if got[0].Msg != "line 0" || got[1].Msg != "line 1" || got[2].Msg != "line 2" {
				t.Fatalf("order broken: %v", got)
			}
			for _, size := range rec.batchSizes() {
				if size > BatchRows {
					t.Fatalf("batch of %d rows exceeds the %d cap", size, BatchRows)
				}
			}
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("3 rows never flushed within 3 beats: %v", rec.batchSizes())
}

func TestFloodOrderAndZeroDrops(t *testing.T) {
	// THE 10k FLOOD GATE, bus half: 10,000 lines pushed as fast as the
	// producer can — all 10,000 delivered IN ORDER, zero drops (backpressure
	// blocks, never drops), every batch ≤ 250 rows.
	const n = 10000
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < n; i++ {
			if !bus.Push(line(i)) {
				t.Errorf("push %d refused mid-flood", i)
				return
			}
		}
	}()
	<-done
	bus.Stop() // drain + land the tail
	got := rec.all()
	if len(got) != n {
		t.Fatalf("delivered %d of %d lines", len(got), n)
	}
	for i := range got {
		if got[i].Msg != fmt.Sprintf("line %d", i) {
			t.Fatalf("order broken at %d: %q", i, got[i].Msg)
		}
	}
	if bus.Dropped() != 0 {
		t.Fatalf("dropped = %d, want 0 (backpressure never drops)", bus.Dropped())
	}
	if bus.Ingested() != n {
		t.Fatalf("ingested = %d, want %d", bus.Ingested(), n)
	}
	for _, size := range rec.batchSizes() {
		if size > BatchRows {
			t.Fatalf("batch of %d rows exceeds the %d cap", size, BatchRows)
		}
	}
}

func TestStopDrainsPending(t *testing.T) {
	// lines buffered at Stop time are still delivered (the drain law).
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	for i := 0; i < 100; i++ {
		bus.Push(line(i))
	}
	bus.Stop()
	if got := rec.all(); len(got) != 100 {
		t.Fatalf("drained %d of 100 buffered lines", len(got))
	}
}

func TestPushAfterStopIsRefused(t *testing.T) {
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	bus.Stop()
	if bus.Push(line(0)) {
		t.Fatal("push after Stop must be refused")
	}
	if bus.Dropped() != 1 {
		t.Fatalf("dropped = %d, want 1", bus.Dropped())
	}
}

func TestStartIsIdempotent(t *testing.T) {
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	bus.Start() // must not double-launch (one loop, one drain)
	for i := 0; i < 10; i++ {
		bus.Push(line(i))
	}
	bus.Stop()
	if got := rec.all(); len(got) != 10 {
		t.Fatalf("delivered %d of 10 (double Start must not duplicate or lose)", len(got))
	}
}

func TestClearWipesPending(t *testing.T) {
	// the clear door's law: pending + buffered drop; lines pushed AFTER
	// Clear() returns land normally (the ack makes the ordering
	// deterministic — no select-randomness flake).
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Start()
	defer bus.Stop()
	for i := 0; i < 50; i++ {
		bus.Push(line(i))
	}
	bus.Clear()
	for i := 100; i < 110; i++ {
		bus.Push(line(i))
	}
	bus.Stop()
	got := rec.all()
	if len(got) != 10 {
		t.Fatalf("delivered %d lines after Clear, want only the 10 post-clear ones: %v", len(got), got)
	}
	for i, l := range got {
		if l.Msg != fmt.Sprintf("line %d", 100+i) {
			t.Fatalf("post-clear order broken at %d: %q", i, l.Msg)
		}
	}
}

func TestClearAndStopBeforeStartAreNoops(t *testing.T) {
	rec := &recorder{}
	bus := New(rec.emit)
	bus.Clear() // must not deadlock (never started)
	bus.Stop()  // must not deadlock (never started)
	if got := rec.all(); len(got) != 0 {
		t.Fatalf("never-started bus delivered %d lines", len(got))
	}
}
