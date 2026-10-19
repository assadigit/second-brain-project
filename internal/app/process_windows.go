// process_windows.go — the Windows half of the single-instance lock's
// PID liveness probe (gui/headless.py:_is_process_running's Windows
// branch, ported to the llmclient process pattern: Windows recycles
// PIDs aggressively, so the probe verifies the IMAGE NAME too — only
// OUR OWN executable counts as "the app still running" (the reference
// checked for python.exe; the Go port checks the packaged exe).
//
// The Windows live legs remain the owner's Phase-7 checklist items
// (the standing register); this port follows the reference's logic
// via `tasklist` (the llmclient/process.go precedent — no new
// dependencies, stdlib only).
//
//go:build windows

package app

import (
	"os"
	"os/exec"
	"strings"
)

// isProcessRunning reports whether pid is a live process whose image
// is our own executable (the PID-recycling guard).
func isProcessRunning(pid int) bool {
	if pid <= 0 {
		return false
	}
	out, err := exec.Command("tasklist", "/FI", "PID eq "+itoa(pid),
		"/FO", "CSV", "/NH").Output()
	if err != nil {
		return false // tasklist itself failed — treat as not running
	}
	row := strings.TrimSpace(string(out))
	if row == "" || strings.Contains(row, "INFO:") {
		return false // no matching task
	}
	// CSV row: "image","PID","Session","Mem". The image must be ours.
	fields := strings.Split(row, ",")
	if len(fields) < 2 {
		return false
	}
	image := strings.ToLower(strings.Trim(fields[0], "\""))
	own := strings.ToLower(strings.TrimSuffix(os.Args[0], ".exe"))
	ownBase := own
	if i := strings.LastIndexAny(own, `/\`); i >= 0 {
		ownBase = own[i+1:]
	}
	return image == ownBase+".exe" || image == ownBase
}

// itoa is the tiny strconv twin (kept local for parity with the
// reference's str() conversions).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}
