//go:build windows

// platform_windows.go — the Windows half of the terminal probes: the
// console TTY checks, the buffer-info window size, and _harden_stdio's
// SetConsoleOutputCP(65001) + SetConsoleCP(65001) (kernel32 via
// syscall.NewLazyDLL — no dependency added).
package cli

import (
	"os"
	"strings"
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

// spawnOllamaServe mirrors _autostart_ollama's Popen half on Windows:
// CREATE_NEW_PROCESS_GROUP | DETACHED_PROCESS so the server survives
// the app, stdout + stderr to DEVNULL. Returns the PID or the error
// class token ("FileNotFoundError" when the binary is not on PATH).
func spawnOllamaServe() (int, string) {
	return spawnDetached("ollama", "serve")
}

// spawnDetached runs argv detached in a new process group.
func spawnDetached(argv ...string) (int, string) {
	devNull, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		return 0, err.Error()
	}
	defer devNull.Close()
	attr := &os.ProcAttr{
		Files: []*os.File{devNull, devNull, devNull},
		Sys: &syscall.SysProcAttr{
			CreationFlags: syscall.CREATE_NEW_PROCESS_GROUP | 0x00000008, // DETACHED_PROCESS
		},
	}
	proc, err := os.StartProcess(argv[0], argv, attr)
	if err != nil {
		return 0, pySpawnErrClassWin(err)
	}
	return proc.Pid, ""
}

// pySpawnErrClassWin maps the spawn failure to the reference's except
// classes.
func pySpawnErrClassWin(err error) string {
	if err == nil {
		return ""
	}
	if strings.Contains(err.Error(), "executable file not found") ||
		strings.Contains(err.Error(), "The system cannot find the file specified") {
		return "FileNotFoundError"
	}
	return err.Error()
}
