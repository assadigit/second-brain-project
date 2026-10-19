//go:build windows

// platform_windows.go — the Windows half of the process control (the
// reference's os.name == 'nt' branches): taskkill /T for the tree
// (a plain terminate kills only the launcher and leaves zombie
// Chrome holding the profile), tasklist for the running pre-check,
// and mklink /J junctions for the attach link (no admin rights).
package chrometabs

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// applySessionSpawn — Windows has no start_new_session equivalent in
// the reference (the guard is os.name == 'posix'); nothing to do.
func applySessionSpawn(cmd *exec.Cmd) {}

// goosIsWindows answers the build's OS (the tests' skip guard).
const goosIsWindows = true

// signalTreeOS mirrors _signal_tree's Windows half: taskkill /PID
// {pid} /T, then /T /F — the tree kill is the law.
func signalTreeOS(s *ChromeSession, proc *exec.Cmd) {
	for _, args := range [][]string{
		{"taskkill", "/PID", itoa(proc.Process.Pid), "/T"},
		{"taskkill", "/PID", itoa(proc.Process.Pid), "/T", "/F"},
	} {
		cmd := exec.Command(args[0], args[1:]...)
		cmd.Stdout = nil
		cmd.Stderr = nil
		done := make(chan error, 1)
		go func() { done <- cmd.Run() }()
		select {
		case <-done:
		case <-time.After(10 * time.Second):
		}
		if waitProcess(s, 5*time.Second) {
			return
		}
	}
	waitProcess(s, 3*time.Second) // a zombie — the endpoint probe tells the truth
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

// chromeIsRunningOS mirrors chrome_is_running's Windows half:
// tasklist asks for chrome.exe. Best-effort by law.
func chromeIsRunningOS() bool {
	defer func() {
		if r := recover(); r != nil {
			// never raises
		}
	}()
	cmd := exec.Command("tasklist", "/FI", "IMAGENAME eq chrome.exe",
		"/FO", "CSV", "/NH")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.Contains(strings.ToLower(string(out)), "chrome.exe")
}

// makeDirLink mirrors the Windows branch: mklink /J — a directory
// junction needs no admin rights (the reference's 15 s timeout).
func makeDirLink(link, target string) error {
	cmd := exec.Command("cmd", "/c", "mklink", "/J", link, target)
	cmd.Stdout = nil
	cmd.Stderr = nil
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()
	select {
	case err := <-done:
		if err != nil {
			return err
		}
	case <-time.After(15 * time.Second):
		return fmt.Errorf("mklink /J timed out after 15s")
	}
	if _, err := os.Stat(link); err != nil {
		return err
	}
	return nil
}

// removeDirLink mirrors _remove_link's Windows half: rmdir on the
// reparse point (never touches the target; the reference's 10 s
// timeout).
func removeDirLink(link string) bool {
	fi, err := os.Lstat(link)
	if err != nil {
		return true // not there — nothing to remove
	}
	_ = fi
	cmd := exec.Command("cmd", "/c", "rmdir", link)
	cmd.Stdout = nil
	cmd.Stderr = nil
	done := make(chan error, 1)
	go func() { done <- cmd.Run() }()
	select {
	case <-done:
	case <-time.After(10 * time.Second):
		return false
	}
	_, err = os.Stat(link)
	return err != nil
}
