//go:build !windows

// platform_posix.go — the POSIX half of the terminal probes: stdout's
// TTY check and the window size (shutil.get_terminal_size's
// os.get_terminal_size twin). The console-codepage half of _harden_stdio
// is a no-op off Windows (the reference guards os.name == "nt").
package cli

import (
	"os"
	"syscall"
	"unsafe"
)

// isATTYStdout mirrors sys.stdout.isatty() for the color door.
func isATTYStdout() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// isATTYStdin mirrors sys.stdin.isatty() for the prompt doors.
func isATTYStdin() bool {
	fi, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}

// stdoutWidth mirrors os.get_terminal_size(sys.__stdout__.fileno()):
// the real console's columns, or 0 when stdout is not a terminal.
func stdoutWidth() int {
	var ws struct{ Row, Col, Xpixel, Ypixel uint16 }
	_, _, errno := syscall.Syscall(syscall.SYS_IOCTL, os.Stdout.Fd(),
		uintptr(syscall.TIOCGWINSZ), uintptr(unsafe.Pointer(&ws)))
	if errno != 0 {
		return 0
	}
	return int(ws.Col)
}

// setConsoleUTF8 is the Windows-only SetConsoleOutputCP(65001) twin —
// a no-op on POSIX.
func setConsoleUTF8() {}
