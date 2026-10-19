// process_posix.go — the POSIX halves of the process doors: the
// single-instance lock's PID liveness probe (gui/headless.py:
// _is_process_running's non-Windows branch: the plain signal-0 probe;
// the reference's Windows-specific image-name check is the windows
// twin's job) and the detached `ollama serve` spawn (the CLI's
// platform twin, start_new_session=True).
//
//go:build !windows

package app

import (
	"errors"
	"os"
	"strings"
	"syscall"
)

// isProcessRunning reports whether pid is a live process, mirroring
// _is_process_running's non-Windows branch (gui/headless.py:246-289):
// the plain signal-0 probe wrapped in the reference's blanket except —
// EVERY failure (no such process, not ours to signal, bad pid) reads
// as not-running. T-78's side-by-side audit caught the divergence this
// corrects: the port had treated EPERM (exists, another user's process)
// as alive, but the reference's `except (OSError, …): return False`
// swallows the PermissionError and reports DEAD — a lock file whose
// stale PID was recycled to a stranger's process is take-over-able in
// Python, so it must be in Go too (the same lock semantics, ADR-0008;
// the cross-binary differential pins it).
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	return killVerdict(syscall.Kill(pid, 0))
}

// killVerdict is the except-ladder as a pure function: nil (the signal
// was permitted — the process is ours and alive) ⇒ true; ANY error —
// ESRCH (no such process) and EPERM (exists but not ours to signal)
// both included — ⇒ false, exactly Python's blanket except.
func killVerdict(err error) bool {
	switch {
	case err == nil:
		return true
	case errors.Is(err, syscall.EPERM):
		return false // the reference's except swallows PermissionError
	default:
		return false
	}
}

// spawnOllamaServe mirrors start_ollama_server's Popen half on POSIX:
// start_new_session (setsid) so the server survives the app, stdout +
// stderr to DEVNULL. Returns the PID or the error class token
// ("FileNotFoundError" when the binary is not on PATH).
func spawnOllamaServe() (int, string) {
	return spawnDetached("ollama", "serve")
}

// spawnDetached runs argv detached in a new session.
func spawnDetached(argv ...string) (int, string) {
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return 0, err.Error()
	}
	defer devNull.Close()
	attr := &os.ProcAttr{
		Files: []*os.File{devNull, devNull, devNull},
		Sys:   &syscall.SysProcAttr{Setsid: true},
	}
	proc, err := os.StartProcess(argv[0], argv, attr)
	if err != nil {
		return 0, pySpawnErrClass(err)
	}
	return proc.Pid, ""
}

// pySpawnErrClass maps the spawn failure to the reference's except
// FileNotFoundError name (the platform's not-found spelling).
func pySpawnErrClass(err error) string {
	if err == nil {
		return ""
	}
	if strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "no such file or directory") {
		return "FileNotFoundError"
	}
	return err.Error()
}
