//go:build windows

// platform_windows.go — the Windows half of the terminal probes: the
// console TTY checks, the buffer-info window size, and _harden_stdio's
// SetConsoleOutputCP(65001) + SetConsoleCP(65001) (kernel32 via
// syscall.NewLazyDLL — no dependency added).
package cli

import (
	"os"
	"syscall"
	"unsafe"
)

var kernel32DLL = syscall.NewLazyDLL("kernel32.dll")

// isATTYStdout mirrors sys.stdout.isatty(): a character device (the
// console) attached to stdout.
func isATTYStdout() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// isATTYStdin mirrors sys.stdin.isatty().
func isATTYStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

type coordStruct struct {
	X, Y int16
}

type smallRectStruct struct {
	Left, Top, Right, Bottom int16
}

type consoleScreenBufferInfoStruct struct {
	Size              coordStruct
	CursorPosition    coordStruct
	Attributes        uint16
	Window            smallRectStruct
	MaximumWindowSize coordStruct
}

// stdoutWidth mirrors os.get_terminal_size: the console's columns via
// GetConsoleScreenBufferInfo, or 0 when stdout is not a console.
func stdoutWidth() int {
	proc := kernel32DLL.NewProc("GetConsoleScreenBufferInfo")
	var info consoleScreenBufferInfoStruct
	r, _, _ := proc.Call(os.Stdout.Fd(), uintptr(unsafe.Pointer(&info)))
	if r == 0 {
		return 0
	}
	return int(info.Window.Right - info.Window.Left + 1)
}

// setConsoleUTF8 mirrors _harden_stdio's Windows half:
// SetConsoleOutputCP(65001) + SetConsoleCP(65001) — best effort, the
// errors="replace" half already makes every print safe.
func setConsoleUTF8() {
	out := kernel32DLL.NewProc("SetConsoleOutputCP")
	in := kernel32DLL.NewProc("SetConsoleCP")
	_, _, _ = out.Call(65001)
	_, _, _ = in.Call(65001)
}
