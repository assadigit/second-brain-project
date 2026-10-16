//go:build windows

package llmclient

import (
	"context"
	"syscall"
	"time"
)

// createNoWindow is CREATE_NO_WINDOW (0x08000000) — the stdlib
// syscall package does not export it on every toolchain; the literal
// keeps the dependency law (stdlib only).
const createNoWindow = 0x08000000

// noWindowAttr keeps the packaged GUI windowless — a flashing cmd
// popup on every Detect would be a regression (CREATE_NO_WINDOW).
func noWindowAttr() *syscall.SysProcAttr {
	return &syscall.SysProcAttr{
		HideWindow:    true,
		CreationFlags: createNoWindow,
	}
}

func timeoutCtxPair(ctx context.Context, timeoutS float64) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(timeoutS*float64(time.Second)))
}
