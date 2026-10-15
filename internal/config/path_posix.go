//go:build !windows

package config

import (
	"os"
	"path/filepath"
	"runtime"
)

// documentsDir mirrors _documents_dir: the Windows Known-Folder first
// (path_windows.go), else ~/Documents ONLY when it exists; "" when the
// feature must sleep. The Documents folder itself is never auto-created.
func documentsDir() string {
	home, err := os.UserHomeDir()
	if err != nil || home == "" {
		return ""
	}
	cand := filepath.Join(home, "Documents")
	if fi, err := os.Stat(cand); err == nil && fi.IsDir() {
		return cand
	}
	return ""
}

// keep runtime imported for the build symmetry with the windows file
var _ = runtime.GOOS
