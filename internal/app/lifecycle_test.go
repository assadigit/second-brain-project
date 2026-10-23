// lifecycle_test.go — the lock/watchdog tests (T-71 slice (c)'s
// acceptance half): the single-instance lock's healing paths, the
// stuck-lock watchdog, the startup passes' cadences, and the shutdown
// grace windows.
package app

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"testing"
	"time"
)

// captureStdout swaps os.Stdout for a pipe around fn (the prints the
// reference's lock speaks — app.py's own console lines).
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w
	defer func() { os.Stdout = old }()
	fn()
	_ = w.Close()
	buf := make([]byte, 8192)
	n, _ := r.Read(buf)
	return string(buf[:n])
}

// deadPID spawns a child that exits and is reaped — a PID the
// liveness probe must read as dead.
func deadPID(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("true")
	if err := cmd.Run(); err != nil {
		t.Skipf("cannot spawn a dying child on this box: %v", err)
	}
	return cmd.ProcessState.Pid()
}

func TestAppLockFreshAcquire(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.lock")
	l := AcquireAppLock(path)
	if !l.Held || !l.exclusive {
		t.Fatalf("fresh acquire: held=%v exclusive=%v", l.Held, l.exclusive)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("the lock file never landed: %v", err)
	}
	// T-78: the BYTES law, exact — str(os.getpid()).encode() with no
	// newline, no padding; the cross-binary file is byte-comparable.
	if got := string(raw); got != strconv.Itoa(os.Getpid()) {
		t.Fatalf("lock bytes = %q (len %d), want the bare decimal PID %q",
			got, len(got), strconv.Itoa(os.Getpid()))
	}
	l.Release()
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("Release left the lock file: %v", err)
	}
}

func TestAppLockRefusesWhileInstanceAlive(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.lock")
	// OUR OWN PID is alive — the honest "another instance" refusal.
	if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		if l := AcquireAppLock(path); l.Held {
			t.Fatal("the lock was taken while its PID is alive")
		}
	})
	if !strings.Contains(out, "Another instance is already running") {
		t.Fatalf("the refusal never spoke: %q", out)
	}
	// The file is untouched (the other instance still owns it).
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("the refused acquire disturbed the lock: %v", err)
	}
}

func TestAppLockHealsStaleLock(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.lock")
	pid := deadPID(t)
	if isProcessRunning(pid) {
		t.Fatalf("the probe says dead PID %d is alive — the test fixture is broken", pid)
	}
	if err := os.WriteFile(path, []byte(strconv.Itoa(pid)), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		l := AcquireAppLock(path)
		if !l.Held || !l.exclusive {
			t.Fatalf("stale healing: held=%v exclusive=%v", l.Held, l.exclusive)
		}
	})
	if !strings.Contains(out, "Removing stale lock file") {
		t.Fatalf("the healing never spoke: %q", out)
	}
	if !strings.Contains(out, "no longer running") {
		t.Fatalf("the healing line lost its reason: %q", out)
	}
	// The healed lock carries OUR pid now — bare bytes (T-78's law).
	raw, _ := os.ReadFile(path)
	if got := string(raw); got != strconv.Itoa(os.Getpid()) {
		t.Fatalf("healed lock bytes = %q, want the bare decimal PID", got)
	}
}

func TestAppLockJunkContentReadsAsUnknownStale(t *testing.T) {
	path := filepath.Join(t.TempDir(), "app.lock")
	if err := os.WriteFile(path, []byte("not a pid at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := captureStdout(t, func() {
		l := AcquireAppLock(path)
		if !l.Held {
			t.Fatal("junk content blocked the lock (the reference treats unreadable as stale)")
		}
	})
	if !strings.Contains(out, "PID unknown") {
		t.Fatalf("the unknown-PID line never spoke: %q", out)
	}
}

func TestIsProcessRunningAliveAndDead(t *testing.T) {
	if !isProcessRunning(os.Getpid()) {
		t.Fatal("our own PID reads as dead")
	}
	if isProcessRunning(-1) || isProcessRunning(0) {
		t.Fatal("a non-positive PID reads as alive")
	}
	pid := deadPID(t)
	if isProcessRunning(pid) {
		t.Fatalf("the reaped child %d reads as alive", pid)
	}
}

// ---------------------------------------------------------------------------
// The watchdog.
// ---------------------------------------------------------------------------

func TestWatchdogForceReleasesStuckLock(t *testing.T) {
	b := NewBus()
	tinyTimeouts(b)
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log})
	l := &OpLock{}
	lc := NewLifecycle(b, l, StartupPasses{})
	lc.WatchdogInterval = 5 * time.Millisecond
	lc.WatchdogMaxAge = 5 * time.Millisecond
	ctx := lc.Start()
	defer lc.Shutdown(nil)

	l.Acquire("batch")
	deadline := time.After(3 * time.Second)
	// The watchdog releases the lock FIRST and speaks the loud
	// line SECOND (lifecycle.go's own order, the reference's twin) —
	// both on its goroutine, so the release can be observable
	// while the line is still in flight. Wait for BOTH under the
	// same deadline: the bare post-release read once lost the coin
	// flip on a loaded CI runner (run 39878157076).
	released := false
	var got LogLine
	var ok bool
	for {
		if !l.Busy() {
			released = true
			got, ok = rec.find("Watchdog")
			if ok {
				break // the release AND the loud line both landed
			}
		}
		select {
		case <-deadline:
			if !released {
				t.Fatal("the watchdog never released the stuck lock")
			}
			t.Fatalf("the watchdog's release line never landed: %v", rec.lines)
		case <-time.After(2 * time.Millisecond):
		}
	}
	if !strings.Contains(got.Msg, "'batch'") || !strings.Contains(got.Msg, "force-released") {
		t.Fatalf("the release line = %q", got.Msg)
	}
	if got.Level != "warning" {
		t.Fatalf("the release line level = %q, want warning", got.Level)
	}
	_ = ctx
}

func TestWatchdogLeavesFreshLocksAlone(t *testing.T) {
	b := NewBus()
	l := &OpLock{}
	lc := NewLifecycle(b, l, StartupPasses{})
	lc.WatchdogInterval = 5 * time.Millisecond
	lc.WatchdogMaxAge = time.Hour // fresh locks are never stuck
	lc.Start()
	defer lc.Shutdown(nil)
	l.Acquire("batch")
	time.Sleep(40 * time.Millisecond)
	if !l.Busy() || l.Owner() != "batch" {
		t.Fatal("the watchdog released a fresh lock")
	}
}

// ---------------------------------------------------------------------------
// The startup passes.
// ---------------------------------------------------------------------------

func TestStartupPassesRunAtTheirCadences(t *testing.T) {
	b := NewBus()
	var mu sync.Mutex
	seen := map[string]int{}
	pass := func(name string) func() {
		return func() {
			mu.Lock()
			seen[name]++
			mu.Unlock()
		}
	}
	lc := NewLifecycle(b, &OpLock{}, StartupPasses{
		Sentinel:     pass("sentinel"),
		LlamaDetect:  pass("llama"),
		Heal:         pass("heal"),
		BacklogCheck: pass("backlog"),
		Radar:        pass("radar"),
	})
	// Tiny delays, same ORDER law: sentinel(1.2) < llama(1.5) <
	// heal(2.0) < backlog(2.5) < radar(3.5) — scaled down together.
	scale := time.Millisecond / 10 // 1200ms→120ms … 3500ms→350ms
	lc.SentinelDelay = SentinelDelay / 10 * (scale / (time.Millisecond / 10))
	lc.SentinelDelay = 120 * time.Millisecond
	lc.LlamaDetectDelay = 150 * time.Millisecond
	lc.HealDelay = 200 * time.Millisecond
	lc.BacklogDelay = 250 * time.Millisecond
	lc.RadarDelay = 350 * time.Millisecond
	lc.SentinelInterval = 100 * time.Millisecond
	lc.WatchdogInterval = time.Hour // quiet in this test

	start := time.Now()
	lc.Start()
	// Wait past the radar delay (the last pass).
	time.Sleep(500 * time.Millisecond)
	lc.Shutdown(nil)

	mu.Lock()
	defer mu.Unlock()
	for _, name := range []string{"sentinel", "llama", "heal", "backlog", "radar"} {
		if seen[name] == 0 {
			t.Fatalf("the %s pass never ran (seen=%v, elapsed=%v)", name, seen, time.Since(start))
		}
	}
	if seen["sentinel"] < 2 {
		t.Fatalf("the sentinel never recurred: %d runs (the 60 s cadence's first beat)", seen["sentinel"])
	}
	if seen["llama"] != 1 || seen["heal"] != 1 || seen["backlog"] != 1 || seen["radar"] != 1 {
		t.Fatalf("a once-per-session pass recurred: %v", seen)
	}
}

func TestStartupPassOrderObserved(t *testing.T) {
	b := NewBus()
	var mu sync.Mutex
	var order []string
	pass := func(name string) func() {
		return func() {
			mu.Lock()
			order = append(order, name)
			mu.Unlock()
		}
	}
	lc := NewLifecycle(b, &OpLock{}, StartupPasses{
		Sentinel:     pass("sentinel"),
		LlamaDetect:  pass("llama"),
		Heal:         pass("heal"),
		BacklogCheck: pass("backlog"),
		Radar:        pass("radar"),
	})
	lc.SentinelDelay = 12 * time.Millisecond
	lc.LlamaDetectDelay = 15 * time.Millisecond
	lc.HealDelay = 20 * time.Millisecond
	lc.BacklogDelay = 25 * time.Millisecond
	lc.RadarDelay = 35 * time.Millisecond
	lc.SentinelInterval = time.Hour
	lc.WatchdogInterval = time.Hour
	lc.Start()
	time.Sleep(80 * time.Millisecond)
	lc.Shutdown(nil)

	mu.Lock()
	defer mu.Unlock()
	want := []string{"sentinel", "llama", "heal", "backlog", "radar"}
	if len(order) < len(want) {
		t.Fatalf("only %v ran (want all five)", order)
	}
	// The first five entries must be exactly the delay order (the
	// sentinel's recurrence is excluded from the order check).
	for i, w := range want {
		if order[i] != w {
			t.Fatalf("pass order = %v, want the reference's delay order %v", order, want)
		}
	}
}

func TestStartupPassCrashIsolated(t *testing.T) {
	b := NewBus()
	done := make(chan struct{})
	lc := NewLifecycle(b, &OpLock{}, StartupPasses{
		Heal:  func() { panic("heal exploded") },
		Radar: func() { close(done) },
	})
	lc.SentinelDelay = time.Hour
	lc.LlamaDetectDelay = time.Hour
	lc.HealDelay = 5 * time.Millisecond
	lc.BacklogDelay = time.Hour
	lc.RadarDelay = 10 * time.Millisecond
	lc.SentinelInterval = time.Hour
	lc.WatchdogInterval = time.Hour
	lc.Start()
	defer lc.Shutdown(nil)
	select {
	case <-done:
		// the radar pass still ran after the heal crashed — the
		// isolation held.
	case <-time.After(3 * time.Second):
		t.Fatal("a crashing startup pass killed the cadence")
	}
}

func TestStartupCancelledBeforeDelay(t *testing.T) {
	b := NewBus()
	ran := make(chan struct{})
	lc := NewLifecycle(b, &OpLock{}, StartupPasses{
		Radar: func() { close(ran) },
	})
	lc.SentinelDelay = time.Hour
	lc.LlamaDetectDelay = time.Hour
	lc.HealDelay = time.Hour
	lc.BacklogDelay = time.Hour
	lc.RadarDelay = 200 * time.Millisecond
	lc.SentinelInterval = time.Hour
	lc.WatchdogInterval = time.Hour
	lc.Start()
	time.Sleep(20 * time.Millisecond)
	lc.Shutdown(nil) // cancels before the radar delay
	select {
	case <-ran:
		t.Fatal("the radar pass ran after Shutdown")
	case <-time.After(300 * time.Millisecond):
		// clean: the pass never fired.
	}
}

// ---------------------------------------------------------------------------
// The shutdown grace windows.
// ---------------------------------------------------------------------------

func TestShutdownWaitsBatchThenReleases(t *testing.T) {
	b := NewBus()
	l := &OpLock{}
	lc := NewLifecycle(b, l, StartupPasses{})
	lc.SentinelDelay = time.Hour
	lc.LlamaDetectDelay = time.Hour
	lc.HealDelay = time.Hour
	lc.BacklogDelay = time.Hour
	lc.RadarDelay = time.Hour
	lc.SentinelInterval = time.Hour
	lc.WatchdogInterval = time.Hour
	lc.BatchGrace = 50 * time.Millisecond
	lc.WorkerGrace = 20 * time.Millisecond
	lc.Start()

	l.Acquire("batch")
	batchDone := make(chan struct{})
	go func() {
		time.Sleep(20 * time.Millisecond) // finishes inside the grace
		close(batchDone)
	}()
	start := time.Now()
	lc.Shutdown(batchDone)
	if d := time.Since(start); d > 200*time.Millisecond {
		t.Fatalf("Shutdown took %v — it did not honor the early finish", d)
	}
	if l.Busy() {
		t.Fatal("Shutdown left the op lock held (the force-release law)")
	}
}

func TestShutdownBatchGraceExpires(t *testing.T) {
	b := NewBus()
	l := &OpLock{}
	lc := NewLifecycle(b, l, StartupPasses{})
	lc.SentinelDelay = time.Hour
	lc.LlamaDetectDelay = time.Hour
	lc.HealDelay = time.Hour
	lc.BacklogDelay = time.Hour
	lc.RadarDelay = time.Hour
	lc.SentinelInterval = time.Hour
	lc.WatchdogInterval = time.Hour
	lc.BatchGrace = 40 * time.Millisecond
	lc.WorkerGrace = 20 * time.Millisecond
	lc.Start()

	l.Acquire("batch")
	batchDone := make(chan struct{}) // never closes — the hung batch
	defer close(batchDone)
	start := time.Now()
	lc.Shutdown(batchDone)
	if d := time.Since(start); d < 35*time.Millisecond {
		t.Fatalf("Shutdown returned in %v — the grace window was not waited", d)
	}
	if d := time.Since(start); d > 2*time.Second {
		t.Fatalf("Shutdown took %v — the grace window expired into a hang", d)
	}
	if l.Busy() {
		t.Fatal("the hung batch's lock outlived the grace")
	}
}

func TestShutdownDrainsWorkersSequentially(t *testing.T) {
	b := NewBus()
	lc := NewLifecycle(b, &OpLock{}, StartupPasses{})
	lc.SentinelDelay = time.Hour
	lc.LlamaDetectDelay = time.Hour
	lc.HealDelay = time.Hour
	lc.BacklogDelay = time.Hour
	lc.RadarDelay = time.Hour
	lc.SentinelInterval = time.Hour
	lc.WatchdogInterval = time.Hour
	lc.BatchGrace = 20 * time.Millisecond
	lc.WorkerGrace = 30 * time.Millisecond
	lc.Start()

	// Two workers: one finishes, one hangs past its grace.
	w1 := make(chan struct{})
	w2 := make(chan struct{})
	defer close(w2)
	un1 := lc.TrackWorker(w1)
	go func() { time.Sleep(5 * time.Millisecond); close(w1) }()
	lc.TrackWorker(w2)
	un1()

	start := time.Now()
	lc.Shutdown(nil)
	d := time.Since(start)
	if d < 25*time.Millisecond {
		t.Fatalf("Shutdown = %v — the hung worker's grace was not waited", d)
	}
	if d > 2*time.Second {
		t.Fatalf("Shutdown = %v — a hung worker hung the shutdown", d)
	}
	// Double Shutdown is a no-op (the once law).
	lc.Shutdown(nil)
}

func TestLifecycleTrackWorkerUntrack(t *testing.T) {
	lc := NewLifecycle(NewBus(), &OpLock{}, StartupPasses{})
	done := make(chan struct{})
	un := lc.TrackWorker(done)
	un()
	un() // idempotent
	lc.mu.Lock()
	n := len(lc.workers)
	lc.mu.Unlock()
	if n != 0 {
		t.Fatalf("untrack left %d workers registered", n)
	}
	_ = context.Background
}
