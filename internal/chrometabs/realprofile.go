// realprofile.go — the owner's own window's strongest launched rung
// (chrome_tabs.py:1331-1601): the real user-data dir probe, the
// attach link (junction/symlink — the one honest way past Chrome
// 136+'s refusal to debug the default user-data-dir path), the
// running-Chrome pre-check, the rung itself, and the identity ladder,
// plus chrome-attach.bat (pure ASCII + CRLF — the .bat law).
package chrometabs

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/assadigit/second-brain-project/internal/handdelivery"
)

// realDataDirsWindows mirrors _REAL_DATA_DIRS_WINDOWS (relative to
// %LOCALAPPDATA%).
var realDataDirsWindows = []string{`Google\Chrome\User Data`}

// realDataDirsPosix mirrors _REAL_DATA_DIRS_POSIX (relative to ~).
var realDataDirsPosix = []string{
	".config/google-chrome",
	".config/chromium",
	"Library/Application Support/Google/Chrome",
}

// chromeIsRunningFn is the pre-check seam (the reference's own tests
// patch ct.chrome_is_running).
var chromeIsRunningFn = ChromeIsRunning

// ChromeIsRunning mirrors chrome_is_running: is a Chrome (any of the
// usual binaries) running right now? Windows asks tasklist; POSIX
// asks pgrep -x over the usual names (the platform files carry each
// half). Best-effort by law: an unaskable question answers False.
func ChromeIsRunning() bool {
	return chromeIsRunningOS()
}

// realUserDataDirFn is the discovery seam (the reference's tests
// patch ct.real_user_data_dir).
var realUserDataDirFn = RealUserDataDir

// RealUserDataDir mirrors real_user_data_dir: the owner's REAL Chrome
// user-data directory, when one exists on this machine. "" when no
// real profile is found. Pure path probing; never raises.
func RealUserDataDir() string {
	defer func() {
		if r := recover(); r != nil {
			// never raises
		}
	}()
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	if runtime.GOOS == "windows" {
		local := os.Getenv("LOCALAPPDATA")
		if local == "" {
			local = filepath.Join(home, "AppData", "Local")
		}
		for _, rel := range realDataDirsWindows {
			cand := filepath.Join(local, rel)
			if isDir(cand) {
				return cand
			}
		}
		return ""
	}
	for _, rel := range realDataDirsPosix {
		cand := filepath.Join(home, rel)
		if isDir(cand) {
			return cand
		}
	}
	return ""
}

func isDir(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// realProfileLinkFn is the link seam (the reference's tests patch
// ct.real_profile_link).
var realProfileLinkFn = RealProfileLink

// RealProfileLink mirrors real_profile_link: the ATTACH LINK — a
// junction (Windows) or symlink (POSIX) at <app>/chrome-real-link
// pointing at the owner's real user-data dir. A link that already
// points at dataDir is reused; one that is DANGLING or points
// elsewhere is re-pointed. Best-effort: "" on failure. Never raises,
// never erases the TARGET.
func RealProfileLink(dataDir string, linkDir string) string {
	defer func() {
		if r := recover(); r != nil {
			// best-effort by law
		}
	}()
	if dataDir == "" {
		dataDir = realUserDataDirFn()
	}
	if dataDir == "" || !isDir(dataDir) {
		return ""
	}
	if linkDir == "" {
		linkDir = filepath.Dir(DefaultProfileDir())
	}
	link := filepath.Join(linkDir, "chrome-real-link")
	wantsRecreate := false
	if _, err := os.Lstat(link); err == nil {
		if sameRealpath(link, dataDir) {
			return link // a link that already points there
		}
		wantsRecreate = true // stale or dangling — re-point
	} else {
		_ = os.MkdirAll(linkDir, 0o755)
	}
	if wantsRecreate {
		if !removeDirLink(link) {
			return "" // a wrong link we cannot fix — skip
		}
	}
	if err := makeDirLink(link, dataDir); err != nil {
		return ""
	}
	return link
}

// sameRealpath answers whether link's resolved target is dataDir
// (os.path.realpath comparison; a dangling link resolves differently
// than a live one).
func sameRealpath(link, dataDir string) bool {
	a, aerr := filepath.EvalSymlinks(link)
	b, berr := filepath.EvalSymlinks(dataDir)
	if aerr != nil || berr != nil {
		return false
	}
	return a == b
}

// tryRealProfileSession mirrors _try_real_profile_session: the
// real-profile rung — launch the owner's REAL profile (through the
// attach link) with the DevTools port open. Honest gates, in order:
// no real profile found → nil; a Chrome already running → nil (never
// fought); no attach link → nil; a launch that fails → nil.
func tryRealProfileSession(exe string, log LogFn, pageTimeoutS float64,
	extraFlags []string) (TabSession, error) {
	dataDir := realUserDataDirFn()
	if dataDir == "" {
		return nil, nil // no real profile on this machine
	}
	if chromeIsRunningFn() {
		log("🤖 Your Chrome is already running without its debug port — "+
			"the door will not fight or restart it on its own. More ▸ "+
			"🪄 Attach to my Chrome restarts it WITH the port (your tabs "+
			"come back); this delivery uses the door's own identity.", "info")
		return nil, nil
	}
	link := realProfileLinkFn(dataDir, "")
	if link == "" {
		log("🤖 The attach link to your real Chrome profile could not be "+
			"created — the door uses its own identity this time", "warning")
		return nil, nil
	}
	session, err := chromeSessionCtor(exe, log, pageTimeoutS, link, true,
		extraFlags, true)
	if err != nil {
		log(fmt.Sprintf("🤖 Your real profile could not be launched with "+
			"the debug port (%s) — the door uses its own identity this "+
			"time (Chrome 136+ guards the default profile tightly; More ▸ "+
			"🪄 Attach to my Chrome is the reliable route)", err.Error()),
			"warning")
		return nil, nil
	}
	return session, nil
}

// launchIdentitySessionFn is the ladder seam (the reference's tests
// patch ct._launch_identity_session).
var launchIdentitySessionFn = launchIdentitySession

// launchIdentitySession mirrors _launch_identity_session: the door's
// identity LADDER for one proxy leg — the real-profile rung first
// (skipped when the caller pinned an explicit profile dir, which IS
// an identity choice), then the dedicated persistent profile. The
// factory seam bypasses the ladder entirely (the scripted session IS
// the chosen identity). Never raises; nil session without error is
// the factory's own failure path.
func launchIdentitySession(exe string, log LogFn, pageTimeoutS float64,
	profileDir string, persist, realProfile bool, extraFlags []string,
	factory func(exe string, log LogFn) (TabSession, error)) (TabSession, error) {
	if factory != nil {
		return factory(exe, log)
	}
	if realProfile && strings.TrimSpace(profileDir) == "" {
		rung, err := tryRealProfileSession(exe, log, pageTimeoutS, extraFlags)
		if err != nil {
			return nil, err
		}
		if rung != nil {
			return rung, nil
		}
	}
	return chromeSessionCtor(exe, log, pageTimeoutS, profileDir, persist,
		extraFlags, false)
}

// WriteChromeAttachBat mirrors write_chrome_attach_bat: the ONE-TIME
// setup that teaches the owner's own Chrome to answer the door.
// Pure ASCII + CRLF (the .bat law); best-effort, never raises;
// returns the path or "".
func WriteChromeAttachBat(path, chromeExe string, port int, dataDir,
	linkDir string) string {
	defer func() {
		if r := recover(); r != nil {
			// best-effort by law
		}
	}()
	if chromeExe == "" {
		chromeExe = handdelivery.FindChrome()
	}
	if dataDir == "" {
		dataDir = RealUserDataDir()
	}
	if chromeExe == "" || dataDir == "" {
		return ""
	}
	link := RealProfileLink(dataDir, linkDir)
	if link == "" {
		return ""
	}
	if port == 0 {
		port = DefaultAttachPort
	}
	lines := []string{
		"@echo off",
		"REM GitCurator v0.57.0 - attach the fifth door to YOUR Chrome",
		"REM (restarts Chrome with its DevTools port open; your tabs",
		"REM  come back; GitCurator closes only ITS OWN tabs; to undo,",
		"REM  just relaunch Chrome normally)",
		"taskkill /IM chrome.exe /T /F >nul 2>&1",
		"timeout /t 2 /nobreak >nul",
		fmt.Sprintf("if not exist \"%s\" mklink /J \"%s\" \"%s\"", link, link, dataDir),
		fmt.Sprintf("start \"\" \"%s\" ", chromeExe) +
			fmt.Sprintf("--remote-debugging-port=%d ", port) +
			fmt.Sprintf("--user-data-dir=\"%s\" --restore-last-session", link),
		"",
	}
	text := strings.Join(lines, "\r\n")
	// the .bat law: pure ASCII (the reference encodes to verify).
	for _, r := range text {
		if r > 0x7f {
			return ""
		}
	}
	dir := filepath.Dir(absPath(path))
	if dir != "" && dir != "." {
		_ = os.MkdirAll(dir, 0o755)
	}
	fh, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o644)
	if err != nil {
		return ""
	}
	defer fh.Close()
	if _, err := fh.WriteString(text); err != nil {
		return ""
	}
	return path
}
