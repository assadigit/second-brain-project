//go:build windows

// platform_windows.go — the Windows half of the mirror leg's git
// spawn (ADR-0009's quirk list): HideWindow + CREATE_NO_WINDOW so
// push_mirror_repo's git calls never flash a console popup.
package directorysync

import (
	"os/exec"
	"syscall"
)

// createNoWindow is CREATE_NO_WINDOW (0x08000000) — the stdlib
// syscall package does not export it on every toolchain; the literal
// keeps the dependency law (stdlib only).
const createNoWindow = 0x08000000

// hideWindow sets the windowless spawn attributes on cmd.
func hideWindow(cmd *exec.Cmd) {
	if cmd.SysProcAttr == nil {
		cmd.SysProcAttr = &syscall.SysProcAttr{}
	}
	cmd.SysProcAttr.HideWindow = true
	cmd.SysProcAttr.CreationFlags |= createNoWindow
}
