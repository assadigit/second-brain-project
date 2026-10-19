// lifecycle.go — T-71 slice (c): the lifecycle — the single-instance
// lock (app.lock + PID liveness healing, gui/app.py:393-440), the
// stuck-lock watchdog (the 60 s recurring check + the 35-minute
// threshold — window.py:135 + telegram_ui.py:_tg_lock_watchdog), the
// startup passes as sequential goroutines with the reference's delays
// (window.py:371-451: sentinel first pass 1.2 s, llama autodetect
// 1.5 s, heal 2.0 s, backlog 2.5 s, radar 3.5 s), and the shutdown
// grace windows (closeEvent, lifecycle.py:173-243: the batch 5 s,
// the test workers 2 s each).
package app

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// The single-instance lock (gui/app.py:393-440, verbatim shape).
// ---------------------------------------------------------------------------

// The reference's own sentences (app.py's prints — the Go port routes
// them through the caller; the lock itself stays wordless).
const (
	anotherInstanceLine = "Another instance is already running (PID %d). Exiting."
	staleLockLine       = "Removing stale lock file (PID %s is no longer running)..."
	lockCreateErrorLine = "Cannot create lock file (%s): "
)

// AppLock is the single-instance lock's state: the app.lock file, the
// PID inside it, and whether THIS process holds it.
type AppLock struct {
	Path string
	// Held is true when this process created (or legitimately
	// overwrote) the lock. A lock we could not take AT ALL is Held
	// false — and the reference's law is CONTINUE WITHOUT IT (never
	// exit on a lock failure).
	Held bool
	// Refused is true ONLY on the alive-PID path — another instance is
	// running right now. That is the ONE not-held case whose reference
	// behavior is sys.exit(1) (app.py: "Another instance is already
	// running. Exiting."); every other failure CONTINUES WITHOUT the
	// lock. Exposed so the shells can honor the distinction (added at
	// T-72; the callers before it had no way to tell the paths apart).
	Refused bool
	// exclusive is false on the last-resort overwrite path (the
	// reference: "we don't have an exclusive lock, but that's OK").
	exclusive bool
}

// AcquireAppLock mirrors app.py:393-440 exactly:
//
//	O_CREAT|O_EXCL create + our PID           → held
//	already exists, PID alive                  → "Another instance…" (not held)
//	already exists, PID dead                   → remove the stale lock,
//	                                             retry create
//	  remove failed (zombie handle)            → sleep 0.5 s, retry,
//	                                             then last resort:
//	                                             overwrite the content
//	retry create lost a race                   → overwrite the content
//	other OS error on the first create         → "Cannot create lock
//	                                             file" — CONTINUE
//	                                             WITHOUT the lock
func AcquireAppLock(path string) *AppLock {
	l := &AppLock{Path: path}
	pid := strconv.Itoa(os.Getpid())

	fd, err := os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = fd.WriteString(pid)
		_ = fd.Close()
		l.Held, l.exclusive = true, true
		return l
	}
	if !os.IsExist(err) {
		// The reference's belt: any other failure is spoken and
		// SURVIVED — the app continues without the lock.
		fmt.Printf(lockCreateErrorLine+"\n", err)
		return l
	}

	// The lock exists — is its PID still alive?
	oldPID, readable := readLockPID(path)
	if readable && oldPID > 0 && isProcessRunning(oldPID) {
		fmt.Printf(anotherInstanceLine+"\n", oldPID)
		l.Refused = true
		return l
	}

	// Stale (or unreadable): force remove and retry.
	who := "unknown"
	if readable && oldPID > 0 {
		who = strconv.Itoa(oldPID)
	}
	fmt.Printf(staleLockLine+"\n", who)
	if err := os.Remove(path); err != nil {
		// The zombie-handle ladder: wait, retry, then overwrite.
		time.Sleep(500 * time.Millisecond)
		if err2 := os.Remove(path); err2 != nil {
			// Last resort: overwrite the content without O_EXCL.
			if werr := os.WriteFile(path, []byte(pid), 0o644); werr == nil {
				l.Held = true
			}
			return l
		}
	}
	fd, err = os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0o644)
	if err == nil {
		_, _ = fd.WriteString(pid)
		_ = fd.Close()
		l.Held, l.exclusive = true, true
		return l
	}
	// O_EXCL still failing (a racing writer won) — overwrite the
	// content; not exclusive, but held.
	if werr := os.WriteFile(path, []byte(pid), 0o644); werr == nil {
		l.Held = true
	}
	return l
}

// readLockPID parses the lock's content (junk reads as unreadable —
// the reference's ValueError/IOError pass).
func readLockPID(path string) (int, bool) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return 0, false
	}
	pid, err := strconv.Atoi(strings.TrimSpace(string(raw)))
	if err != nil || pid <= 0 {
		return 0, false
	}
	return pid, true
}

// Release removes the lock file (best-effort — a crashed shutdown
// never worsens).
func (l *AppLock) Release() {
	if l == nil || l.Path == "" || !l.Held {
		return
	}
	_ = os.Remove(l.Path)
	l.Held = false
}

// ---------------------------------------------------------------------------
// The lifecycle: startup passes + watchdog + shutdown graces.
// ---------------------------------------------------------------------------

// The startup delays (window.py's singleShot values — the cadence is
// behavior; never bump without an owner decision).
const (
	SentinelDelay    = 1200 * time.Millisecond
	LlamaDetectDelay = 1500 * time.Millisecond
	HealDelay        = 2000 * time.Millisecond
	BacklogDelay     = 2500 * time.Millisecond
	RadarDelay       = 3500 * time.Millisecond
	// SentinelInterval is the recurring pass's cadence (the ONE 60 s
	// background pass feeding Net · AI · Proxy).
	SentinelInterval = 60 * time.Second
	// WatchdogInterval is the stuck-lock check's cadence (window.py:
	// the 60 s timer).
	WatchdogInterval = 60 * time.Second
	// The shutdown grace windows (closeEvent: worker.wait(5000), the
	// test workers' w.wait(2000)).
	BatchGrace  = 5 * time.Second
	WorkerGrace = 2 * time.Second
)

// StartupPasses is the injected pass set — the DI root wires the real
// engines (the sentinel's connection checks, the llamacpp autodetect,
// the hygiene heal, the review-backlog notice, the quiet radar); nil
// passes are skipped (the cadence still runs — a missing engine is
// not a crash).
type StartupPasses struct {
	Sentinel     func() // the recurring pass's each run
	LlamaDetect  func() // once per session
	Heal         func() // once (vault_hygiene_on_start is the engine's own gate)
	BacklogCheck func() // once
	Radar        func() // once (radar_pending is the engine's own gate)
}

// Lifecycle owns the startup cadences, the watchdog, and the
// shutdown drain.
type Lifecycle struct {
	Bus    *Bus
	Lock   *OpLock
	Passes StartupPasses

	// The instrument fields (production defaults = the reference's
	// numbers; tests shrink them — the fake-clock law's sibling).
	SentinelDelay    time.Duration
	LlamaDetectDelay time.Duration
	HealDelay        time.Duration
	BacklogDelay     time.Duration
	RadarDelay       time.Duration
	SentinelInterval time.Duration
	WatchdogInterval time.Duration
	WatchdogMaxAge   time.Duration
	BatchGrace       time.Duration
	WorkerGrace      time.Duration

	cancel   context.CancelFunc
	stopOnce sync.Once

	mu      sync.Mutex
	workers []chan struct{} // in-flight worker dones (the test workers' Go twins)
}

// NewLifecycle builds the lifecycle with the reference's cadences.
func NewLifecycle(bus *Bus, lock *OpLock, passes StartupPasses) *Lifecycle {
	return &Lifecycle{
		Bus:              bus,
		Lock:             lock,
		Passes:           passes,
		SentinelDelay:    SentinelDelay,
		LlamaDetectDelay: LlamaDetectDelay,
		HealDelay:        HealDelay,
		BacklogDelay:     BacklogDelay,
		RadarDelay:       RadarDelay,
		SentinelInterval: SentinelInterval,
		WatchdogInterval: WatchdogInterval,
		WatchdogMaxAge:   StuckSeconds,
		BatchGrace:       BatchGrace,
		WorkerGrace:      WorkerGrace,
	}
}

// RunSentinelPass asks for ONE extra sentinel pass NOW (the save law's
// 250 ms re-arm — vaults_config.py's QTimer.singleShot(250,
// _run_sentinel_pass)): provider / proxy / URL edits reach the top-bar
// cues (Net · AI · Proxy) within a second, not at the next 60 s tick.
// The pass itself is all background — this only asks for it; it never
// raises and never blocks the caller.
func (l *Lifecycle) RunSentinelPass() {
	go runPass("sentinel-refresh", l.Passes.Sentinel)
}

// TrackWorker registers an in-flight worker's done channel (the
// _active_test_workers list's Go twin) so Shutdown can drain it
// inside its grace window.
func (l *Lifecycle) TrackWorker(done chan struct{}) (untrack func()) {
	l.mu.Lock()
	l.workers = append(l.workers, done)
	l.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			l.mu.Lock()
			for i, d := range l.workers {
				if d == done {
					l.workers = append(l.workers[:i], l.workers[i+1:]...)
					break
				}
			}
			l.mu.Unlock()
		})
	}
}

// runPass runs one pass recovered — a startup crash never takes the
// app down (the TestWorker contract's own law).
func runPass(name string, fn func()) {
	if fn == nil {
		return
	}
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[app-lifecycle] startup pass '%s' crashed (isolated): %v\n", name, r)
		}
	}()
	fn()
}

// delayed runs fn after d (cancelled by ctx — the cooperative Stop
// semantics: the chunked sleeps become the select law, doc 04 §4).
func delayed(ctx context.Context, d time.Duration, fn func()) bool {
	select {
	case <-ctx.Done():
		return false
	case <-time.After(d):
		return true
	}
}

// Start launches the startup passes as sequential goroutines with the
// reference's delays (each pass independent; the sentinel recurs on
// its 60 s cadence; the watchdog ticks on its own) + the watchdog.
// The returned context is the lifecycle's root — Shutdown cancels it.
func (l *Lifecycle) Start() context.Context {
	ctx, cancel := context.WithCancel(context.Background())
	l.cancel = cancel

	// The sentinel: first pass at the delay, then the recurring
	// cadence (doc 04 §6: "independent goroutines with the same
	// cadences").
	go func() {
		if delayed(ctx, l.SentinelDelay, func() {}) {
			runPass("sentinel", l.Passes.Sentinel)
		}
		t := time.NewTicker(l.SentinelInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				runPass("sentinel", l.Passes.Sentinel)
			}
		}
	}()
	// The once-per-session passes, each at its own delay.
	go func() {
		if delayed(ctx, l.LlamaDetectDelay, func() {}) {
			runPass("llamacpp-autodetect", l.Passes.LlamaDetect)
		}
	}()
	go func() {
		if delayed(ctx, l.HealDelay, func() {}) {
			runPass("vault-heal", l.Passes.Heal)
		}
	}()
	go func() {
		if delayed(ctx, l.BacklogDelay, func() {}) {
			runPass("review-backlog", l.Passes.BacklogCheck)
		}
	}()
	go func() {
		if delayed(ctx, l.RadarDelay, func() {}) {
			runPass("repo-radar", l.Passes.Radar)
		}
	}()

	// The watchdog: every WatchdogInterval, a lock held longer than
	// WatchdogMaxAge is force-released with the loud line (the v0.05
	// forever-bug's safety net).
	go func() {
		t := time.NewTicker(l.WatchdogInterval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				l.runWatchdogTick()
			}
		}
	}()
	return ctx
}

// runWatchdogTick is one watchdog check (also exported for the
// lifecycle tests + the shell's manual door).
func (l *Lifecycle) runWatchdogTick() {
	defer func() { recover() }() // best-effort — never crash the timer
	if l.Lock == nil || !l.Lock.IsStuck(l.WatchdogMaxAge) {
		return
	}
	heldFor := int(l.Lock.AgeSeconds())
	evicted := l.Lock.ForceRelease("watchdog")
	l.Bus.EmitLogLine(fmt.Sprintf(watchdogReleaseLine, evicted, heldFor), "warning")
}

// Shutdown mirrors closeEvent's drain order: stop the timers (the
// cancel), wait the in-flight batch inside its 5 s grace, wait each
// worker inside its 2 s grace, then force-release the op lock (the
// "harmless at exit, correct if the app is reused" law).
func (l *Lifecycle) Shutdown(batchDone <-chan struct{}) {
	l.stopOnce.Do(func() {
		if l.cancel != nil {
			l.cancel() // the _closing flag + every timer stopped
		}
		// The batch's grace window.
		if batchDone != nil {
			select {
			case <-batchDone:
			case <-time.After(l.BatchGrace):
			}
		}
		// The workers' grace windows — sequential, each its own 2 s
		// cap (the reference's for-loop over _active_test_workers).
		l.mu.Lock()
		workers := append([]chan struct{}(nil), l.workers...)
		l.mu.Unlock()
		for _, done := range workers {
			select {
			case <-done:
			case <-time.After(l.WorkerGrace):
			}
		}
		// The lock never stays held past shutdown.
		if l.Lock != nil {
			l.Lock.ForceRelease("shutdown")
		}
	})
}
