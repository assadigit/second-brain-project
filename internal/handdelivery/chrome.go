// chrome.go — the Chrome launch family (find_chrome / open_in_chrome,
// hand_delivery.py:398-459): the real Chrome's executable and the
// honest open ladder (the found Chrome, else the system default
// browser, loopback never opens).
package handdelivery

import (
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
)

// windowsChromePaths mirrors _WINDOWS_CHROME_PATHS (the first two are
// literal; the reference's third entry expands %LOCALAPPDATA% through
// os.path.expandvars — the port expands LOCALAPPDATA at call time on
// the Windows leg, the same evaluation the reference performs).
var windowsChromePaths = []string{
	`C:\Program Files\Google\Chrome\Application\chrome.exe`,
	`C:\Program Files (x86)\Google\Chrome\Application\chrome.exe`,
}

// posixChromeNames mirrors _POSIX_CHROME_NAMES (shutil.which order,
// first hit wins — Chromium and Edge count: the door is 'a real
// browser the owner drives', and they render the same challenge
// pages).
var posixChromeNames = []string{
	"google-chrome", "google-chrome-stable", "chrome",
	"chromium", "chromium-browser", "microsoft-edge", "msedge",
}

// The FindChrome seams — the reference's own tests patch os.name,
// os.path.isfile and shutil.which; the port's battery patches these
// (the same contract, Go's shape).
var (
	// chromeOSName answers the platform (runtime.GOOS; "windows" is
	// the reference's os.name == 'nt').
	chromeOSName = func() string { return runtime.GOOS }
	// chromeIsFile mirrors os.path.isfile.
	chromeIsFile = func(p string) bool {
		st, err := os.Stat(p)
		return err == nil && !st.IsDir()
	}
	// chromeWhich mirrors shutil.which.
	chromeWhich = func(name string) string {
		found, err := exec.LookPath(name)
		if err != nil {
			return ""
		}
		return found
	}
)

// FindChromeFn is the open ladder's finder seam (the reference's
// own tests patch hd.find_chrome directly; the port's battery patches
// this — the finder itself stays testable through the three seams
// below).
var FindChromeFn = FindChrome

// FindChrome mirrors find_chrome: the real Chrome's executable, when
// one can be found — well-known chrome.exe paths on Windows,
// shutil.which over the usual suspects on POSIX. Empty when nothing
// is found (the caller falls back to the system default browser,
// honestly named).
func FindChrome() string {
	if chromeOSName() == "windows" {
		candidates := append([]string{}, windowsChromePaths...)
		if localAppData := os.Getenv("LOCALAPPDATA"); localAppData != "" {
			candidates = append(candidates,
				filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"))
		}
		for _, p := range candidates {
			if p != "" && chromeIsFile(p) {
				return p
			}
		}
		return ""
	}
	for _, name := range posixChromeNames {
		if found := chromeWhich(name); found != "" {
			return found
		}
	}
	return ""
}

// SubprocessPopenFn is the detached-launch seam (subprocess.Popen in
// the reference). argv[0] is the executable; the real implementation
// starts the browser and never waits (stdout/stderr/stdin ride the
// null device, the home dir is the cwd, close_fds is Go's default —
// the detached child keeps running after the app exits).
var SubprocessPopenFn = func(argv []string) error {
	cmd := exec.Command(argv[0], argv[1:]...)
	cmd.Stdout, cmd.Stderr, cmd.Stdin = nil, nil, nil
	if home, err := os.UserHomeDir(); err == nil {
		cmd.Dir = home
	}
	return cmd.Start()
}

// WebbrowserOpenFn is the system-browser fallback seam
// (webbrowser.open). The real implementation asks the OS's default
// browser door; an error mirrors Python's webbrowser raising (the
// reference's except path answers the same warning as a Chrome
// launch failure).
var WebbrowserOpenFn = func(u string) (bool, error) {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("rundll32.exe", "url.dll,FileProtocolHandler", u)
	case "darwin":
		cmd = exec.Command("open", u)
	default:
		cmd = exec.Command("xdg-open", u)
	}
	if err := cmd.Start(); err != nil {
		return false, err
	}
	return true, nil
}

// OpenInChrome mirrors open_in_chrome: open url in the owner's REAL
// Chrome — the found Chrome (detached — the app never waits for the
// browser to close), else the system default browser (named as the
// fallback in the log — the owner deserves to know which door
// opened). Loopback NEVER opens (the v0.15.1 rule). Returns
// "chrome" / "browser" / "" — never raises.
func OpenInChrome(rawURL string, log LogFn) string {
	if log == nil {
		log = noopLog
	}
	u := strings.TrimSpace(rawURL)
	lu := strings.ToLower(u)
	if u == "" || !(strings.HasPrefix(lu, "http://") || strings.HasPrefix(lu, "https://")) {
		return ""
	}
	// the loopback law; a parse that cannot be asked is not a block
	// (the reference's except: pass).
	if parsed, err := url.Parse(u); err == nil {
		host := strings.ToLower(parsed.Hostname())
		if host == "localhost" || host == "127.0.0.1" || host == "::1" ||
			strings.HasSuffix(host, ".local") {
			log(fmt.Sprintf("🖐 Hand-delivery: loopback URL never opens a browser (%s)", u), "info")
			return ""
		}
	}
	exe := FindChromeFn()
	if exe != "" {
		if err := SubprocessPopenFn([]string{exe, u}); err != nil {
			// the reference's except path — the same honest warning
			// for a failed launch and a failed fallback.
			log(fmt.Sprintf("⚠️ Chrome could not be launched (%v) — open %s by hand", err, u), "warning")
			return ""
		}
		log(fmt.Sprintf("🖐 Opened in your real Chrome: %s", u), "info")
		return "chrome"
	}
	opened, err := WebbrowserOpenFn(u)
	if err != nil {
		log(fmt.Sprintf("⚠️ Chrome could not be launched (%v) — open %s by hand", err, u), "warning")
		return ""
	}
	if opened {
		log(fmt.Sprintf("🖐 Opened in your default browser (no Chrome found — the honest fallback): %s", u), "info")
		return "browser"
	}
	log(fmt.Sprintf("⚠️ No browser could be opened for %s — open it by hand", u), "warning")
	return ""
}
