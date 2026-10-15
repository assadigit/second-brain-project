package dryrun

import (
	"os"
	"runtime"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// FormatComma delegates to the canonical helpers home — the T-17
// promotion law: formatComma MOVED to internal/helpers (promote, never
// duplicate); this alias keeps dryrun's historical callers compiling.
func FormatComma(n int64) string { return helpers.FormatComma(n) }

// translateNewlines mirrors Python text-mode writes: on Windows open('w')
// translates '\n' to os.linesep (and, like Python's naive translation,
// an existing "\r\n" in content becomes "\r\r\n" — faithful, not fixed).
// On every other platform the content passes through unchanged.
func translateNewlines(content string) string {
	if runtime.GOOS != "windows" {
		return content
	}
	return strings.ReplaceAll(content, "\n", "\r\n")
}

// errExist mirrors Python's FileExistsError from os.makedirs with
// exist_ok=False (or exist_ok=True over an existing FILE).
func errExist(path string) error {
	return &os.PathError{Op: "mkdir", Path: path, Err: os.ErrExist}
}
