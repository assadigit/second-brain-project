//go:build !windows

// platform_posix.go — the POSIX half of the process control (the
// reference's os.name == 'posix' branches): the session's own process
// group (start_new_session), the tree kill (killpg SIGTERM→SIGKILL),
// pgrep over the usual browser names, and plain symlinks for the
// attach link.
package chrometabs

import (
	"os"
	"os/exec"
	"syscall"
	"time"
)

// applySessionSpawn mirrors start_new_session=True: the throwaway
// session gets its OWN process group, so close() can kill the whole
// tree in one signal — never the app's own group.
func applySessionSpawn(cmd *exec.Cmd) {
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
}

// goosIsWindows answers the build's OS (the tests' skip guard).
const goosIsWindows = false

// signalTreeOS mirrors _signal_tree's POSIX half: killpg SIGTERM,
// wait 5 s, escalate SIGKILL; a lone pid falls back to the cmd's own
// Signal. Best-effort by law.
func signalTreeOS(s *ChromeSession, proc *exec.Cmd) {
	for _, sig := range []syscall.Signal{syscall.SIGTERM, syscall.SIGKILL} {
		func() {
			defer func() { recover() }()
			if pgid, err := syscall.Getpgid(proc.Process.Pid); err == nil {
				_ = syscall.Kill(-pgid, sig)
			} else {
				_ = proc.Process.Signal(sig)
			}
		}()
		if waitProcess(s, 5*time.Second) {
			return
		}
	}
	waitProcess(s, 3*time.Second) // a zombie — the endpoint probe tells the truth
}

// runningBrowserNames mirrors _RUNNING_BROWSER_NAMES (pgrep -x).
var runningBrowserNames = []string{
	"chrome", "chromium", "chromium-browser",
	"google-chrome", "google-chrome-stable",
}

// chromeIsRunningOS mirrors chrome_is_running's POSIX half: pgrep -x
// over the usual names. Best-effort by law: an unaskable question
// answers False.
func chromeIsRunningOS() bool {
	defer func() {
		if r := recover(); r != nil {
			// never raises
		}
	}()
	if _, err := exec.LookPath("pgrep"); err != nil {
		return false
	}
	for _, name := range runningBrowserNames {
		cmd := exec.Command("pgrep", "-x", name)
		if err := cmd.Run(); err == nil {
			return true
		}
	}
	return false
}

// makeDirLink mirrors the POSIX branch of real_profile_link's link
// creation: a plain symlink.
func makeDirLink(link, target string) error {
	return os.Symlink(target, link)
}

// removeDirLink mirrors _remove_link's POSIX half: unlink the link
// (a real directory is never the door's to remove — the caller
// checks); True when nothing (or the link) is gone.
func removeDirLink(link string) bool {
	fi, err := os.Lstat(link)
	if err != nil {
		return true // not there — nothing to remove
	}
	if fi.IsDir() && fi.Mode()&os.ModeSymlink == 0 {
		return false // a real directory is never the door's to remove
	}
	return os.Remove(link) == nil
}
