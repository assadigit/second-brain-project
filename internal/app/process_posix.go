// process_posix.go — the POSIX half of the single-instance lock's PID
// liveness probe (gui/headless.py:_is_process_running's non-Windows
// branch: the plain signal-0 probe; the reference's Windows-specific
// image-name check is the windows twin's job).
//
//go:build !windows

package app

import (
	"errors"
	"syscall"
)

// isProcessRunning reports whether pid is a live process. Signal 0
// performs the kernel's permission+existence check: EPERM means the
// process exists but belongs to someone else (alive); ESRCH means
// no such process.
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	err := syscall.Kill(pid, 0)
	switch {
	case err == nil:
		return true
	case errors.Is(err, syscall.EPERM):
		return true // exists, not ours to signal
	default:
		return false
	}
}
