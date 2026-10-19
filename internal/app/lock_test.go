// lock_test.go — the OpLock contract tests (T-71): the reference
// telegram_lock.py's own defect classes, each pinned.
package app

import (
	"strings"
	"testing"
	"time"
)

func TestOpLockAcquireReleaseLaws(t *testing.T) {
	l := &OpLock{}
	if !l.Acquire("batch") {
		t.Fatal("first acquire refused")
	}
	if l.Busy() != true || l.Owner() != "batch" {
		t.Fatalf("state = busy=%v owner=%q", l.Busy(), l.Owner())
	}
	// No nested re-acquire (the double-release bug's door).
	if l.Acquire("batch") {
		t.Fatal("nested re-acquire by the same owner succeeded — the v0.05 law forbids it")
	}
	if l.Acquire("other") {
		t.Fatal("second acquire while held succeeded")
	}
	// Owner-scoped release: a different owner's release is refused.
	if l.Release("other") {
		t.Fatal("a different owner released the lock — the over-release bug")
	}
	if !l.Busy() {
		t.Fatal("the wrong-owner release actually freed it")
	}
	if !l.Release("batch") {
		t.Fatal("the owner's own release refused")
	}
	if l.Busy() || l.Owner() != "" {
		t.Fatalf("after release: busy=%v owner=%q", l.Busy(), l.Owner())
	}
	// Release when free: refused, no-op.
	if l.Release("batch") {
		t.Fatal("release-when-free returned true")
	}
}

func TestOpLockForceReleaseAndDescribe(t *testing.T) {
	l := &OpLock{}
	if ev := l.ForceRelease("watchdog"); ev != "" {
		t.Fatalf("force on a free lock evicted %q", ev)
	}
	l.Acquire("batch")
	if ev := l.ForceRelease("watchdog"); ev != "batch" {
		t.Fatalf("force evicted %q, want batch", ev)
	}
	if l.Busy() {
		t.Fatal("force left the lock held")
	}
	// Describe: free vs held-with-age.
	if got := l.Describe(); got != "free" {
		t.Fatalf("Describe(free) = %q", got)
	}
	l.Acquire("scan")
	d := l.Describe()
	if !strings.HasPrefix(d, "scan (running ") || !strings.HasSuffix(d, "s)") {
		t.Fatalf("Describe(held) = %q, want 'scan (running Ns)'", d)
	}
}

func TestOpLockAgeAndStuck(t *testing.T) {
	l := &OpLock{}
	if l.AgeSeconds() != 0 {
		t.Fatal("age of a free lock must be 0")
	}
	if l.IsStuck(time.Hour) {
		t.Fatal("free lock reported stuck")
	}
	l.Acquire("batch")
	if l.AgeSeconds() > 5 {
		t.Fatalf("fresh lock age = %v", l.AgeSeconds())
	}
	if l.IsStuck(time.Hour) {
		t.Fatal("fresh lock reported stuck")
	}
	if !l.IsStuck(-time.Second) {
		t.Fatal("a negative max age must report stuck (age > max)")
	}
	_ = l.ForceRelease("test")
}

func TestStuckSecondsIsTheReferenceWatchdog(t *testing.T) {
	if StuckSeconds != 35*60 {
		t.Fatalf("StuckSeconds = %v, want the reference's 35 min", StuckSeconds)
	}
}
