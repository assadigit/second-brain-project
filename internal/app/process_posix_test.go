// process_posix_test.go — the POSIX liveness verdict's pure law,
// pinned hermetically (T-78's side-by-side audit): killVerdict mirrors
// _is_process_running's blanket except — every failure, EPERM included,
// reads as not-running. The live cross-binary differential (a real
// other-user PID probed by both binaries) rides testdiff/side_by_side.py.
//
//go:build !windows

package app

import (
	"errors"
	"syscall"
	"testing"
)

func TestKillVerdictMirrorsTheBlanketExcept(t *testing.T) {
	// The signal was permitted — the process is ours and alive.
	if !killVerdict(nil) {
		t.Fatal("a permitted signal-0 probe must read alive")
	}
	// No such process — dead.
	if killVerdict(syscall.ESRCH) {
		t.Fatal("ESRCH must read dead")
	}
	// Exists but not ours to signal — the reference's blanket except
	// (PermissionError is an OSError subclass) swallows it: DEAD. The
	// pre-T-78 port read this branch as alive — the divergence the
	// side-by-side differential caught; Python takes over a lock whose
	// stale PID was recycled to a stranger's process, so Go must too.
	if killVerdict(syscall.EPERM) {
		t.Fatal("EPERM must read dead (the reference's except swallows PermissionError)")
	}
	// Any other failure — dead (the blanket except catches Exception).
	if killVerdict(errors.New("something else went wrong")) {
		t.Fatal("an unknown probe failure must read dead")
	}
}
