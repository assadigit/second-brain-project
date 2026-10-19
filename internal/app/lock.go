// lock.go — T-71: the one-operation lock — the port of
// gitcurator/gui/telegram_lock.py's TelegramLockManager (the v0.05
// forever-bug's structural fix), repurposed for the Go port's
// one-batch-at-a-time law (doc 04 §2: "an app-level mutex + running
// flag mirrors the TelegramLockManager law (owner-scoped
// acquire/release, 35 min watchdog)").
//
// The laws that made the reference's lock correct (each fixes a real
// defect class — the docstring's own three):
//
//   - OWNER TRACKING: every acquire names its owner; release is
//     owner-scoped — a finished batch can never release a lock held
//     by a different operation (the v0.05 over-release bug).
//   - NO NESTED RE-ACQUIRE: the current owner re-acquiring is refused
//     too (the old code let a nested acquire silently succeed and
//     then double-release).
//   - AGE + DESCRIPTION: the busy state says WHAT holds the lock and
//     for HOW LONG ('batch (running 84s)') — an actionable
//     diagnostic, never a bare "busy".
//   - WATCHDOG SUPPORT: isStuck(maxAge) lets the lifecycle watchdog
//     detect and force-release a lock whose holder died without
//     cleanup.
//
// All methods are idempotent and never raise (a lock helper that
// throws would itself be a reliability bug).
package app

import (
	"fmt"
	"sync"
	"time"
)

// StuckSeconds is the watchdog threshold — the reference's
// _TG_STUCK_SECONDS (window.py:135): 35 min, longer than any real
// operation's hard cap, so a lock still held past it means the
// finishing chain itself died.
const StuckSeconds = 35 * 60

// The watchdog's release line — shaped on the reference's own
// (telegram_ui.py:_tg_lock_watchdog) but honest to the Go port's
// lock (the ONE-BATCH law, not a Telegram session guard — Option B
// retired the MTProto client): the owner, the age, the advice.
const watchdogReleaseLine = "🔓 Watchdog: the operation lock held by '%s' for %ds looks stuck — force-released. If the batch misbehaves, restart the app."

// OpLock is the thread-safe single-slot lock with owner tracking and
// age reporting.
type OpLock struct {
	mu         sync.Mutex
	owner      string
	acquiredAt time.Time
}

// Acquire tries to take the lock for owner. Returns false when
// already held (by anyone, including owner itself — no nested
// re-acquire).
func (l *OpLock) Acquire(owner string) bool {
	if owner == "" {
		owner = "?"
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owner != "" {
		return false
	}
	l.owner = owner
	l.acquiredAt = time.Now()
	return true
}

// Release frees the lock when owner still holds it ("" releases
// unconditionally — the shutdown path). Returns whether this call
// actually released.
func (l *OpLock) Release(owner string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owner == "" {
		return false
	}
	if owner != "" && l.owner != owner {
		return false
	}
	l.owner = ""
	l.acquiredAt = time.Time{}
	return true
}

// ForceRelease unconditionally clears the lock; returns the evicted
// owner ("" when it was already free). The watchdog (and shutdown)
// use this.
func (l *OpLock) ForceRelease(reason string) string {
	_ = reason // kept for logging symmetry at call sites (the reference's own shape)
	l.mu.Lock()
	defer l.mu.Unlock()
	evicted := l.owner
	l.owner = ""
	l.acquiredAt = time.Time{}
	return evicted
}

// Busy reports whether the lock is held.
func (l *OpLock) Busy() bool {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.owner != ""
}

// Owner names the current holder ("" when free).
func (l *OpLock) Owner() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return l.owner
}

// AgeSeconds reports how long the current holder has held the lock
// (0 when free).
func (l *OpLock) AgeSeconds() float64 {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owner == "" {
		return 0
	}
	d := time.Since(l.acquiredAt).Seconds()
	if d < 0 {
		return 0
	}
	return d
}

// Describe renders the busy state for log lines:
// "batch (running 84s)" or "free".
func (l *OpLock) Describe() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.owner == "" {
		return "free"
	}
	return fmt.Sprintf("%s (running %ds)", l.owner, int(time.Since(l.acquiredAt).Seconds()))
}

// IsStuck reports whether the lock has been held longer than maxAge
// (the watchdog's probe).
func (l *OpLock) IsStuck(maxAge time.Duration) bool {
	return l.Busy() && l.AgeSeconds() > maxAge.Seconds()
}
