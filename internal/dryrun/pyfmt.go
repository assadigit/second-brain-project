package dryrun

import (
	"os"
	"runtime"
	"strconv"
	"strings"
)

// formatComma mirrors Python's f"{n:,}" — thousands separators in groups
// of three (doc 05 §3 item 12). Sizes are non-negative here, but the
// negative form is handled for completeness. Candidate for promotion
// into the T-17 helpers package when that lands.
func formatComma(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	out := strings.Join(parts, ",")
	if neg {
		return "-" + out
	}
	return out
}

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
