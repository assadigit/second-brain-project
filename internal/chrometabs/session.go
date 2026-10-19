// session.go — the Chrome process (chrome_tabs.py:723-1328): the
// loopback laws and free-port pick, the DevTools HTTP endpoints, the
// anti-bot patch, and ChromeSession / AttachedChromeSession — the
// owner's real Chrome as a fetch engine, launched with a persistent
// dedicated profile (the anti-bot identity) or adopted when its
// endpoint answers.
package chrometabs

import (
	"bufio"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// LogFn is the logger shape the whole module rides (handdelivery's
// own — an alias so one type flows through both doors).
type LogFn = func(msg, level string)

// noopLog mirrors the reference's `lambda *a, **k: None`.
func noopLog(string, string) {}

// DefaultPageTimeoutS mirrors DEFAULT_PAGE_TIMEOUT_S.
const DefaultPageTimeoutS = 45.0

// DefaultWave mirrors DEFAULT_WAVE.
const DefaultWave = 8

// DefaultAttachPort mirrors DEFAULT_ATTACH_PORT.
const DefaultAttachPort = 9222

// chromeStartupTimeoutS mirrors CHROME_STARTUP_TIMEOUT_S.
const chromeStartupTimeoutS = 25.0

// devtoolsLine mirrors _DEVTOOLS_LINE: the stderr marker Chrome
// prints when its DevTools endpoint is live.
const devtoolsLine = "DevTools listening on "

// chromeFlags mirrors _CHROME_FLAGS (first-run noise off, the
// backgrounding disablers, the automation blink feature off).
var chromeFlags = []string{
	"--no-first-run",
	"--no-default-browser-check",
	"--disable-backgrounding-occluded-windows",
	"--disable-renderer-backgrounding",
	"--disable-background-timer-throttling",
	"--disable-blink-features=AutomationControlled",
}

// containerFlags mirrors _CONTAINER_FLAGS (the SUID-sandbox fallback
// for containers and hardened boxes).
var containerFlags = []string{
	"--no-sandbox",
	"--disable-gpu",
	"--disable-dev-shm-usage",
}

// headlessFlags mirrors _HEADLESS_FLAGS (a display-less machine gets
// the SAME real Chromium headless).
var headlessFlags = []string{"--headless=new"}

// antiBotPatch mirrors _ANTI_BOT_PATCH (patchable in tests so the
// scripted-socket scripts stay honest).
var antiBotPatch = true

// AntiBotScript mirrors _ANTI_BOT_SCRIPT: navigator.webdriver
// patched undefined on every new document.
const AntiBotScript = "Object.defineProperty(navigator, 'webdriver', " +
	"{get: function () { return undefined; }});"

// Timing seams (patchable in tests — the reference's own tests patch
// the module constants).
var (
	// PageSettleS mirrors PAGE_SETTLE_S: how long a page must stay
	// COMPLETE at ONE http(s) URL before the DOM is read.
	PageSettleS = 2.0
	// pollCadenceS mirrors _POLL_CADENCE_S.
	pollCadenceS = 0.4
	// reconnectPauseS mirrors _RECONNECT_PAUSE_S.
	reconnectPauseS = 0.5
)

// socketReconnects mirrors _SOCKET_RECONNECTS.
const socketReconnects = 2

// IsLoopbackURL mirrors is_loopback_url: True for localhost /
// 127.0.0.1 / ::1 / .local hosts — the v0.15.1 law: no door ever
// escalates off-machine.
func IsLoopbackURL(u string) bool {
	_, hostname, _ := pyURLParts(u)
	host := helpers.PyLower(hostname)
	return host == "localhost" || host == "127.0.0.1" || host == "::1" ||
		strings.HasSuffix(host, ".local")
}

// PickFreePort mirrors _pick_free_port: an unused TCP port on
// loopback (bind(0), read it, release).
func PickFreePort() int {
	s, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return 0
	}
	defer s.Close()
	return s.Addr().(*net.TCPAddr).Port
}

// DefaultProfileDir mirrors default_profile_dir: <APP_DIR>/
// chrome-profile (the app's own folder, next to cache.db — the
// profile travels with the app). The reference's import-failure
// fallback cannot happen in Go — constants.AppDir always resolves.
func DefaultProfileDir() string {
	return filepath.Join(constants.AppDir(), "chrome-profile")
}

// applyAntiBot mirrors _apply_anti_bot: Page.enable +
// Page.addScriptToEvaluateOnNewDocument with the navigator.webdriver
// erasure. Best-effort by law (a page that refuses the patch is still
// fetched — the verdict, not the patch, is the door's law); skipped
// when antiBotPatch is off. A CDP error is swallowed; a non-CDP
// socket error is PROPAGATED (the reference's except catches only
// CDPError — a reset rides the reconnect path).
func applyAntiBot(ws cdpConn) error {
	if !antiBotPatch {
		return nil
	}
	if _, err := ws.Call("Page.enable", nil, 5.0); err != nil {
		if asCDP(err) {
			return nil // a page that refuses the patch is still fetchable
		}
		return err
	}
	params := storage.NewOrderedMap()
	params.Set("source", AntiBotScript)
	if _, err := ws.Call("Page.addScriptToEvaluateOnNewDocument", params, 5.0); err != nil {
		if asCDP(err) {
			return nil
		}
		return err
	}
	return nil
}

// probeVersionEndpoint mirrors _probe_version_endpoint: the one
// honest liveness test — does the DevTools HTTP endpoint on port
// answer GET /json/version with 200? Pure stdlib, loopback only.
func probeVersionEndpoint(port int, timeoutS float64) bool {
	if timeoutS <= 0 {
		timeoutS = 3.0
	}
	client := &http.Client{
		Timeout: time.Duration(timeoutS * float64(time.Second)),
	}
	resp, err := client.Get("http://127.0.0.1:" + strconv.Itoa(port) +
		"/json/version")
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)
	return resp.StatusCode == 200
}

// ProbeVersionEndpoint is the exported liveness probe (the battery
// rides it directly).
func ProbeVersionEndpoint(port int) bool { return probeVersionEndpoint(port, 3.0) }

// httpOn mirrors _http_on: one HTTP round trip on a DevTools endpoint
// (loopback only). Raises on any miss — the callers name it honestly.
func httpOn(port int, method, path string, timeoutS float64) (int, []byte, error) {
	if timeoutS <= 0 {
		timeoutS = 15.0
	}
	client := &http.Client{
		Timeout: time.Duration(timeoutS * float64(time.Second)),
	}
	req, err := http.NewRequest(method,
		"http://127.0.0.1:"+strconv.Itoa(port)+path, nil)
	if err != nil {
		return 0, nil, err
	}
	resp, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, err
	}
	return resp.StatusCode, body, nil
}

// pyQuoteSafeEmpty mirrors urllib.parse.quote(url, safe=”): percent-
// encode everything except the unreserved set (letters, digits,
// _.-~), uppercase hex.
func pyQuoteSafeEmpty(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z',
			c >= '0' && c <= '9', c == '_', c == '.', c == '-', c == '~':
			b.WriteByte(c)
		default:
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0x0f])
		}
	}
	return b.String()
}

// TabInfo is the /json/new target dict's two fields the door reads:
// the target id and its webSocketDebuggerUrl.
type TabInfo struct {
	ID    string
	WsURL string
}

// openTabHTTP mirrors _open_tab_http: open one tab at url on the
// endpoint at port (PUT — the Chrome 111+ law — with a GET fallback
// for older builds). CDPError on any miss.
func openTabHTTP(port int, rawURL string) (TabInfo, error) {
	q := pyQuoteSafeEmpty(rawURL)
	last := "no HTTP answer"
	for _, method := range []string{"PUT", "GET"} {
		status, body, err := httpOn(port, method, "/json/new?"+q, 15.0)
		if err != nil {
			last = err.Error()
			continue
		}
		if status == 200 {
			om, derr := storage.DecodeJSON(body)
			if derr != nil {
				return TabInfo{}, newCDPError(
					"the new-tab answer was unreadable: %s", derr.Error())
			}
			tab := TabInfo{}
			if v, ok := om.Get("id"); ok && v != nil {
				tab.ID = helpers.PyStr(v)
			}
			if v, ok := om.Get("webSocketDebuggerUrl"); ok && v != nil {
				tab.WsURL = helpers.PyStr(v)
			}
			return tab, nil
		}
		last = "HTTP " + strconv.Itoa(status)
	}
	return TabInfo{}, newCDPError("Chrome would not open a tab (%s)", last)
}

// closeTabHTTP mirrors _close_tab_http: best-effort tab close on the
// endpoint (GET first; a miss is cosmetic).
func closeTabHTTP(port int, targetID string) {
	for _, method := range []string{"GET", "PUT"} {
		status, _, err := httpOn(port, method, "/json/close/"+targetID, 8.0)
		if err != nil {
			continue
		}
		if status == 200 || status == 204 {
			return
		}
	}
}

// TabSession is the wave engine's session contract (the reference
// duck-types new_tab / close_tab / alive / port; the port rides the
// ws-URL fallback).
type TabSession interface {
	NewTab(rawURL string) (TabInfo, error)
	CloseTab(targetID string)
	Alive() bool
	Port() int
	Close()
}

// ChromeSession mirrors ChromeSession: the owner's real Chrome as a
// fetch engine — launched with a persistent dedicated profile (the
// anti-bot identity), ADOPTED through DevToolsActivePort when a
// Chrome already holds the profile, climbing ONE honest launch ladder
// (container flags, then headless). exe may be injected (tests pass a
// script that prints the endpoint line).
type ChromeSession struct {
	log                LogFn
	pageTimeoutS       float64
	extraFlags         []string
	realProfile        bool
	proc               *exec.Cmd
	procDone           chan struct{}
	stderrPipe         io.ReadCloser
	stderrReader       *bufio.Reader
	drainStarted       bool
	profile            string
	port               int
	browserWSPath      string // for the graceful Browser.close
	profileIsThrowaway bool
}

// chromeSessionCtor is the constructor seam (the reference's tests
// patch ct.ChromeSession itself).
var chromeSessionCtor = func(exe string, log LogFn, pageTimeoutS float64,
	profileDir string, persist bool, extraFlags []string,
	realProfile bool) (TabSession, error) {
	return NewChromeSession(exe, log, pageTimeoutS, profileDir, persist,
		extraFlags, realProfile)
}

// NewChromeSession mirrors ChromeSession.__init__ (exe, log,
// page_timeout_s, profile_dir, persist, extra_flags, real_profile).
// The launch ladder: base flags → +container → +headless.
func NewChromeSession(exe string, log LogFn, pageTimeoutS float64,
	profileDir string, persist bool, extraFlags []string,
	realProfile bool) (*ChromeSession, error) {
	if log == nil {
		log = noopLog
	}
	s := &ChromeSession{
		log:          log,
		pageTimeoutS: pageTimeoutS,
		realProfile:  realProfile,
	}
	for _, f := range extraFlags {
		if f != "" {
			s.extraFlags = append(s.extraFlags, f)
		}
	}
	s.profileIsThrowaway = !persist && !realProfile
	if persist && !realProfile {
		dir := strings.TrimSpace(profileDir)
		if dir == "" {
			dir = DefaultProfileDir()
		}
		s.profile = absPath(dir)
		if err := os.MkdirAll(s.profile, 0o755); err != nil {
			log(fmt.Sprintf("🤖 The dedicated Chrome profile could not be "+
				"created (%s) — falling back to a throwaway session (the "+
				"cookies will not persist this time)",
				helpers.PyOSErrorText(err, s.profile)), "warning")
			s.profile = ""
			s.profileIsThrowaway = true
		} else {
			if s.tryAdopt() {
				log(fmt.Sprintf("🤖 Adopted the still-running Chrome session "+
					"of the dedicated profile (its cookies stay warm; DevTools "+
					"on 127.0.0.1:%d) — no new window was opened", s.port), "info")
				return s, nil
			}
		}
	} else if realProfile {
		// the owner's REAL profile through the attach link: the caller
		// resolved the link and pre-checked that his Chrome is not
		// running; never adopted, never erased.
		dir := strings.TrimSpace(profileDir)
		if dir == "" {
			return nil, newCDPError("the real-profile rung needs the attach " +
				"link path")
		}
		s.profile = absPath(dir)
		if err := os.MkdirAll(s.profile, 0o755); err != nil {
			return nil, err
		}
	}
	// else: the throwaway law — a fresh temp profile (persist off).
	attempts := [][]string{
		append([]string{}, chromeFlags...),
		append(append([]string{}, chromeFlags...), containerFlags...),
		append(append(append([]string{}, chromeFlags...), containerFlags...),
			headlessFlags...),
	}
	var lastErr error
	launched := false
	for attempt, extra := range attempts {
		if err := s.launchAndAnnounce(exe, extra); err != nil {
			lastErr = err
			s.Close() // this attempt's process + throwaway profile
			if attempt == 0 {
				s.log("🤖 Chrome died before its DevTools endpoint — "+
					"retrying with the container flags (--no-sandbox "+
					"--disable-gpu — the SUID-sandbox fallback for "+
					"containers and hardened boxes)", "warning")
			} else if attempt == 1 {
				s.log("🤖 Chrome still cannot open its window (no display?) "+
					"— last resort: the same real Chrome HEADLESS (the tabs "+
					"are not visible, but the engine and the live DOM are "+
					"identical)", "warning")
			}
			continue
		}
		launched = true
		break
	}
	if !launched {
		if lastErr != nil {
			return nil, lastErr
		}
		return nil, newCDPError("Chrome never announced its DevTools endpoint")
	}
	s.drainStderr()
	if s.realProfile {
		s.log(fmt.Sprintf("🤖 Chrome session live — YOUR REAL PROFILE "+
			"through the attach link (your cookies, your logins, your "+
			"clearances; DevTools on 127.0.0.1:%d) — the door closes only "+
			"its own tabs when it is done", s.port), "info")
	} else if s.profileIsThrowaway {
		s.log(fmt.Sprintf("🤖 Chrome session live (your own Chrome, a fresh "+
			"throwaway profile — your running window is untouched; DevTools "+
			"on 127.0.0.1:%d)", s.port), "info")
	} else {
		s.log(fmt.Sprintf("🤖 Chrome session live (your own Chrome, the "+
			"door's DEDICATED profile — a returning visitor the sites "+
			"remember, your own window and profile untouched; DevTools on "+
			"127.0.0.1:%d)", s.port), "info")
	}
	return s, nil
}

func absPath(p string) string {
	if a, err := filepath.Abs(p); err == nil {
		return a
	}
	return p
}

// tryAdopt mirrors _try_adopt: a Chrome still running with the
// dedicated profile is the door's own leftover — Chrome writes
// DevToolsActivePort into the user-data-dir the moment its endpoint
// is live. Adoption succeeds only when the port ANSWERS
// /json/version.
func (s *ChromeSession) tryAdopt() bool {
	portFile := filepath.Join(s.profile, "DevToolsActivePort")
	data, err := os.ReadFile(portFile)
	if err != nil {
		return false
	}
	var lines []string
	for _, l := range helpers.SplitLines(string(data)) {
		if t := strings.TrimSpace(l); t != "" {
			lines = append(lines, t)
		}
	}
	if len(lines) == 0 {
		return false
	}
	port, err := strconv.Atoi(lines[0])
	if err != nil {
		return false
	}
	wsPath := ""
	if len(lines) > 1 {
		wsPath = lines[1]
	}
	if !probeVersionEndpoint(port, 3.0) {
		return false
	}
	s.port = port
	if wsPath == "" {
		wsPath = s.lookupBrowserWS()
	}
	s.browserWSPath = wsPath
	return true
}

// lookupBrowserWS mirrors _lookup_browser_ws: /json/version → the
// browser's webSocketDebuggerUrl path (for the graceful close).
// Best-effort; ” on any miss.
func (s *ChromeSession) lookupBrowserWS() string {
	status, body, err := httpOn(s.port, "GET", "/json/version", 3.0)
	if err != nil || status != 200 {
		return ""
	}
	om, err := storage.DecodeJSON(body)
	if err != nil {
		return ""
	}
	ws := ""
	if v, ok := om.Get("webSocketDebuggerUrl"); ok && v != nil {
		ws = helpers.PyStr(v)
	}
	return pathOfURL(ws)
}

// launchAndAnnounce mirrors _launch_and_announce: start the process
// (the dedicated profile when persist is on, else a fresh throwaway
// temp dir; its own process group on POSIX), then read the DevTools
// endpoint off its stderr. CDPError on any miss — the constructor
// decides whether to retry. The session's own extraFlags ride EVERY
// attempt.
func (s *ChromeSession) launchAndAnnounce(exe string, extra []string) error {
	s.port = PickFreePort()
	s.browserWSPath = ""
	if s.profile == "" {
		dir, err := os.MkdirTemp("", "gitcurator-chrome-")
		if err != nil {
			return newCDPError("Chrome could not be launched (%s)", err.Error())
		}
		s.profile = dir
	} else {
		if err := os.MkdirAll(s.profile, 0o755); err != nil {
			// exist_ok=True still raises on a DANGLING symlink (the
			// makedirs quirk) — a launch failure, honestly named.
			return newCDPError("the profile directory could not be "+
				"prepared (%s)", err.Error())
		}
	}
	s.drainStarted = false
	cmd := exec.Command(exe)
	cmd.Args = append(cmd.Args,
		fmt.Sprintf("--remote-debugging-port=%d", s.port),
		fmt.Sprintf("--user-data-dir=%s", s.profile))
	cmd.Args = append(cmd.Args, extra...)
	cmd.Args = append(cmd.Args, s.extraFlags...)
	cmd.Args = append(cmd.Args, "about:blank")
	// stdout DEVNULL, stdin DEVNULL (nil in Go IS /dev/null); the
	// stderr pipe is taken below.
	applySessionSpawn(cmd) // POSIX: the session's own process group
	pipe, err := cmd.StderrPipe()
	if err != nil {
		return newCDPError("Chrome could not be launched (%s)", err.Error())
	}
	if err := cmd.Start(); err != nil {
		return newCDPError("Chrome could not be launched (%s)", err.Error())
	}
	s.proc = cmd
	s.stderrPipe = pipe
	s.stderrReader = bufio.NewReader(pipe)
	done := make(chan struct{})
	s.procDone = done
	go func() {
		_ = cmd.Wait()
		close(done)
	}()
	// Chrome announces "DevTools listening on ws://127.0.0.1:<port>/…"
	// on stderr — read it here; the drain takes over after.
	deadline := time.Now().Add(time.Duration(chromeStartupTimeoutS * float64(time.Second)))
	endpoint := ""
	for time.Now().Before(deadline) {
		line, rerr := s.stderrReader.ReadString('\n')
		if len(line) > 0 {
			text := decodeUTF8Replace([]byte(line))
			if strings.Contains(text, devtoolsLine) {
				endpoint = strings.TrimSpace(
					strings.SplitN(text, devtoolsLine, 2)[1])
				break
			}
		}
		if rerr != nil {
			if s.procExited() {
				return newCDPError("Chrome exited before its DevTools " +
					"endpoint came up")
			}
			// EOF on a pipe the OS has not reaped yet — never a busy
			// spin; the deadline is the honest end. (Any read error
			// counts: Go's cmd.Wait() CLOSES the pipe, so a reaped
			// process's reads answer "file already closed", not EOF.)
			time.Sleep(50 * time.Millisecond)
			continue
		}
	}
	if endpoint == "" {
		return newCDPError("Chrome never announced its DevTools endpoint")
	}
	// ws_url = endpoint.split()[0]; the port rides `parsed.port or
	// self.port` (a 0/absent port keeps the picked one); the browser's
	// own WebSocket path rides the same announcement.
	fields := strings.Fields(endpoint)
	if len(fields) > 0 {
		wsFirst := fields[0]
		if portStr := portOfNetloc(wsFirst); portStr != "" {
			if n, err := strconv.Atoi(portStr); err == nil && n != 0 {
				s.port = n
			}
		}
		s.browserWSPath = pathOfURL(wsFirst)
	}
	return nil
}

// portOfNetloc extracts the port text from a URL's netloc.
func portOfNetloc(u string) string {
	rest := u
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	}
	end := len(rest)
	for j := 0; j < len(rest); j++ {
		if rest[j] == '/' || rest[j] == '?' || rest[j] == '#' {
			end = j
			break
		}
	}
	netloc := rest[:end]
	if at := strings.LastIndexByte(netloc, '@'); at >= 0 {
		netloc = netloc[at+1:]
	}
	if strings.HasPrefix(netloc, "[") {
		if rb := strings.IndexByte(netloc, ']'); rb >= 0 {
			netloc = netloc[rb+1:]
			if strings.HasPrefix(netloc, ":") {
				return netloc[1:]
			}
			return ""
		}
	}
	if c := strings.IndexByte(netloc, ':'); c >= 0 {
		return netloc[c+1:]
	}
	return ""
}

// pathOfURL extracts the path of a URL (urlparse().path — ” when
// absent).
func pathOfURL(u string) string {
	rest := u
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
	} else if i := strings.Index(rest, ":"); i >= 0 && !strings.HasPrefix(rest[i:], "://") {
		// scheme:path (no netloc) — urlparse's path is everything
		// after the scheme.
		return rest[i+1:]
	}
	if c := strings.IndexByte(rest, '/'); c >= 0 {
		end := len(rest)
		for j := c; j < len(rest); j++ {
			if rest[j] == '?' || rest[j] == '#' {
				end = j
				break
			}
		}
		return rest[c:end]
	}
	return ""
}

// procExited answers whether the launched process has been reaped
// (Python's proc.poll() is not None).
func (s *ChromeSession) procExited() bool {
	if s.procDone == nil {
		return false
	}
	select {
	case <-s.procDone:
		return true
	default:
		return false
	}
}

// waitProcess waits for the reaper goroutine's signal (the Wait
// goroutine closes procDone) — both platform halves ride it.
func waitProcess(s *ChromeSession, d time.Duration) bool {
	if s.procDone == nil {
		return false
	}
	select {
	case <-s.procDone:
		return true
	case <-time.After(d):
		return false
	}
}

// drainStderr mirrors _drain_stderr: own the stderr pipe for the
// session's life — keep reading so a full pipe can never block the
// browser, and CLOSE it on EOF.
func (s *ChromeSession) drainStderr() {
	if s.proc == nil || s.stderrReader == nil {
		return
	}
	reader := s.stderrReader
	pipe := s.stderrPipe
	s.drainStarted = true
	go func() {
		defer func() { recover() }()
		for {
			if _, err := reader.ReadString('\n'); err != nil {
				_ = pipe.Close()
				return
			}
		}
	}()
}

// NewTab mirrors new_tab: open one tab at url (PUT — the Chrome 111+
// law — with a GET fallback).
func (s *ChromeSession) NewTab(rawURL string) (TabInfo, error) {
	return openTabHTTP(s.port, rawURL)
}

// CloseTab mirrors close_tab: best-effort tab close.
func (s *ChromeSession) CloseTab(targetID string) {
	closeTabHTTP(s.port, targetID)
}

// Port answers the session's DevTools port.
func (s *ChromeSession) Port() int { return s.port }

// Alive mirrors alive: the honest liveness — the launcher process
// alive OR the DevTools endpoint ANSWERING (the launcher exiting is
// NOT session death — Windows chrome.exe delegates).
func (s *ChromeSession) Alive() bool {
	if s.proc != nil && !s.procExited() {
		return true
	}
	if s.port != 0 {
		return probeVersionEndpoint(s.port, 3.0)
	}
	return false
}

// newBrowserSocketFn is the graceful-close socket seam (the
// reference's tests patch ct.CDPSocket for the same site).
var newBrowserSocketFn = func(host string, port int, path string, timeoutS float64) (cdpConn, error) {
	ws, err := NewCDPSocket(host, port, path)
	if err != nil {
		return nil, err
	}
	ws.timeout = timeoutS
	return ws, nil
}

// Close mirrors close: the GRACEFUL way first — Browser.close over
// the browser's own WebSocket (Chrome flushes its cookie store) —
// then any survivor is tree-killed; the throwaway profile is erased,
// the DEDICATED profile NEVER. Never raises; safe to call twice.
func (s *ChromeSession) Close() {
	defer func() { recover() }()
	s.gracefulBrowserClose()
	proc := s.proc
	s.proc = nil
	if proc != nil && !s.procExited() {
		s.signalTree(proc)
	}
	if proc != nil && !s.drainStarted && s.stderrPipe != nil {
		// the launch failed before the drain ever started — the pipe
		// is ours to close (nothing else ever read it)
		_ = s.stderrPipe.Close()
	}
	s.cleanup()
}

// gracefulBrowserClose mirrors _graceful_browser_close:
// Browser.close over the browser's own WebSocket, then an ~8s wait.
// Best-effort by law: any miss falls through to the tree kill.
func (s *ChromeSession) gracefulBrowserClose() {
	defer func() { recover() }()
	wsPath := s.browserWSPath
	if wsPath == "" && s.port != 0 {
		wsPath = s.lookupBrowserWS()
	}
	if wsPath == "" || s.port == 0 {
		return
	}
	ws, err := newBrowserSocketFn("127.0.0.1", s.port, wsPath, 5.0)
	if err != nil {
		return
	}
	_, _ = ws.Call("Browser.close", nil, 5.0)
	ws.Close()
	if s.proc != nil && !s.procExited() {
		select {
		case <-s.procDone:
		case <-time.After(8 * time.Second):
		}
	}
}

// signalTree mirrors _signal_tree (the platform file carries the
// OS-specific half).
func (s *ChromeSession) signalTree(proc *exec.Cmd) { signalTreeOS(s, proc) }

// cleanup mirrors _cleanup: a throwaway profile is erased; the
// DEDICATED profile is NEVER erased — its cookies are the door's
// identity.
func (s *ChromeSession) cleanup() {
	profile := s.profile
	s.profile = ""
	if s.profileIsThrowaway && profile != "" {
		if st, err := os.Stat(profile); err == nil && st.IsDir() {
			_ = os.RemoveAll(profile)
		}
	}
}

// AttachedChromeSession mirrors AttachedChromeSession: the owner's
// OWN running Chrome, driven through its DevTools port — never
// launches, never kills; closes only its own tabs.
type AttachedChromeSession struct {
	port   int
	opened []string
}

// NewAttachedChromeSession mirrors AttachedChromeSession(port, log).
func NewAttachedChromeSession(port int, _ LogFn) *AttachedChromeSession {
	return &AttachedChromeSession{port: port}
}

// NewTab opens one tab in the owner's own Chrome; the tab id is
// remembered so Close stays gentle.
func (a *AttachedChromeSession) NewTab(rawURL string) (TabInfo, error) {
	tab, err := openTabHTTP(a.port, rawURL)
	if err != nil {
		return TabInfo{}, err
	}
	if tab.ID != "" {
		a.opened = append(a.opened, tab.ID)
	}
	return tab, nil
}

// CloseTab closes ONE tab — best-effort.
func (a *AttachedChromeSession) CloseTab(targetID string) {
	closeTabHTTP(a.port, targetID)
}

// Port answers the owner's Chrome DevTools port.
func (a *AttachedChromeSession) Port() int { return a.port }

// Alive: the endpoint's answer — the owner's Chrome lives exactly as
// long as it lives.
func (a *AttachedChromeSession) Alive() bool {
	return probeVersionEndpoint(a.port, 3.0)
}

// Close is GENTLE: never Browser.close, never a kill — only the tabs
// the door opened are closed (best-effort, idempotent, never raises).
func (a *AttachedChromeSession) Close() {
	for _, tid := range a.opened {
		func() {
			defer func() { recover() }()
			a.CloseTab(tid)
		}()
	}
	a.opened = nil
}
