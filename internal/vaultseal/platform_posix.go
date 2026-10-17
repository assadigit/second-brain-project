//go:build !windows

// platform_posix.go — the POSIX half of the git subprocess spawn:
// no window attributes exist; the child inherits the terminal law of
// its parent (git -C <vault> … is a plain pipe-driven child).
package vaultseal

import "os/exec"

// hideWindow is the POSIX no-op (the windows twin sets
// HideWindow + CREATE_NO_WINDOW).
func hideWindow(*exec.Cmd) {}
