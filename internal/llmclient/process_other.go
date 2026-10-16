//go:build !windows

package llmclient

import (
	"context"
	"syscall"
	"time"
)

func noWindowAttr() *syscall.SysProcAttr { return nil }

func timeoutCtxPair(ctx context.Context, timeoutS float64) (context.Context, context.CancelFunc) {
	return context.WithTimeout(ctx, time.Duration(timeoutS*float64(time.Second)))
}
