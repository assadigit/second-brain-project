// shell_test.go — the production transport's headless proof: the REAL
// internal/app.Bus (the fan-out, exactly as the engines drive it) →
// this shell's subscriber shape → the deskbus (the 120 ms coalescing
// beat, the ≤250-row batches, the cap-1024 backpressure, the drain
// law). The wails EventsEmit hop is the only piece this test cannot
// exercise headless — the emitter is injected there (the same seam the
// production shell uses at the window); the gate leg (build-support/
// run-gate.sh) is the end-to-end evidence for that hop.
package main

import (
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/webui/deskbus"
)

// subscribeDesk wires the production subscriber shape over a real Bus:
// LogLine → the TS-formatted deskbus push (the shell's own law).
func subscribeDesk(bus *app.Bus, desk *deskbus.LogBus) (cancel func()) {
	return bus.Subscribe(&app.Handlers{
		LogLine: func(ev app.LogLine) {
			desk.Push(deskbus.LogLine{
				TS:    time.Now().Format("15:04:05"),
				Level: ev.Level,
				Msg:   ev.Msg,
			})
		},
	})
}

func TestProductionTransportFlood(t *testing.T) {
	// THE LOG PERF TEST's headless half: 10,000 lines through the FULL
	// production chain — Bus.EmitLogLine (the fan-out) → the subscriber
	// → the deskbus — all delivered in order, zero drops (backpressure,
	// never silent loss), every batch ≤ 250 rows, the accept rate far
	// past the ≥2,000 rows/s gate.
	const n = 10000
	var batches [][]deskbus.LogLine
	desk := deskbus.New(func(rows []deskbus.LogLine) {
		cp := make([]deskbus.LogLine, len(rows))
		copy(cp, rows)
		batches = append(batches, cp)
	})
	desk.Start()
	defer desk.Stop()

	bus := app.NewBus()
	cancel := subscribeDesk(bus, desk)
	defer cancel()

	t0 := time.Now()
	for i := 0; i < n; i++ {
		bus.EmitLogLine(fmt.Sprintf("flood line %d — the production transport under load", i), "info")
	}
	pushed := time.Since(t0)

	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		if desk.Ingested()-desk.Dropped() >= n {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	desk.Stop() // drain + land the tail

	var got []deskbus.LogLine
	for _, b := range batches {
		if len(b) > deskbus.BatchRows {
			t.Fatalf("batch of %d rows exceeds the %d cap", len(b), deskbus.BatchRows)
		}
		got = append(got, b...)
	}
	if len(got) != n {
		t.Fatalf("delivered %d of %d lines through the production transport", len(got), n)
	}
	for i := range got {
		if !strings.HasPrefix(got[i].Msg, fmt.Sprintf("flood line %d —", i)) {
			t.Fatalf("order broken at %d: %q", i, got[i].Msg)
		}
		if got[i].TS == "" {
			t.Fatalf("row %d has no [ts] (the subscriber must format it)", i)
		}
	}
	if desk.Dropped() != 0 {
		t.Fatalf("dropped = %d, want 0 (backpressure never drops)", desk.Dropped())
	}
	if desk.Ingested() != n {
		t.Fatalf("ingested = %d, want %d", desk.Ingested(), n)
	}
	linesPerS := float64(n) / pushed.Seconds()
	if linesPerS < 2000 {
		t.Fatalf("the production transport accepted only %.0f lines/s (gate ≥ 2,000)", linesPerS)
	}
	t.Logf("production transport: %d lines in %v (≥ %.0f lines/s), %d batches, 0 drops",
		n, pushed.Round(time.Millisecond), linesPerS, len(batches))
}

func TestProductionTransportLevelsAndTS(t *testing.T) {
	// the typed levels ride the same channel; the [ts] is built at the
	// subscriber (the reference's log_message law). The collector is
	// mutex-guarded: the deskbus's loop goroutine appends while the
	// test body polls (the recorder law — deskbus_test.go's own shape).
	var mu sync.Mutex
	var got []deskbus.LogLine
	desk := deskbus.New(func(rows []deskbus.LogLine) {
		mu.Lock()
		got = append(got, rows...)
		mu.Unlock()
	})
	desk.Start()
	defer desk.Stop()

	bus := app.NewBus()
	cancel := subscribeDesk(bus, desk)
	defer cancel()

	for _, lvl := range []string{"info", "success", "warning", "error"} {
		bus.EmitLogLine("one "+lvl+" line", lvl)
	}
	deadline := time.Now().Add(3 * deskbus.FlushInterval)
	for time.Now().Before(deadline) {
		mu.Lock()
		n := len(got)
		mu.Unlock()
		if n >= 4 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(got) != 4 {
		t.Fatalf("got %d rows, want 4", len(got))
	}
	for i, lvl := range []string{"info", "success", "warning", "error"} {
		if got[i].Level != lvl {
			t.Fatalf("row %d level = %q, want %q", i, got[i].Level, lvl)
		}
		if got[i].TS == "" {
			t.Fatalf("row %d has no [ts]", i)
		}
	}
}

func TestShellLockRefusalField(t *testing.T) {
	// the additive field's law: Refused is true ONLY on the alive-PID
	// path — the fresh AcquireAppLock on a free path must be Held, not
	// Refused (the refusal itself is lifecycle_test's battery; this is
	// the shell-side contract that the distinction EXISTS).
	l := app.AcquireAppLock(t.TempDir() + "/app.lock")
	if !l.Held {
		t.Fatalf("a fresh lock must be held")
	}
	if l.Refused {
		t.Fatalf("a fresh lock must not be refused")
	}
	l.Release()
}
