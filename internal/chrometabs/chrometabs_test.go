// chrometabs_test.go — T-44: the behavior battery, the ported
// selections of the reference's own tests/test_fifthdoor.py (§1 the
// frame codec, §2 the DevTools client, §2b the page verdict + the
// load watch, §3 the candidates, §4 the session helpers + the
// fake-exe Chrome session, §5 the fetch driver's honest misses, §6
// the delivery, §7b the consume-side re-verification — the auto-door
// discard battery SESSION-018 deferred to exactly this module). The
// pure-protocol shapes are additionally byte-pinned by golden_test.go;
// this file ports the reference suite's own assertions (Go's shape,
// the same laws).
//
// Unit law: temp vaults, patched seams, zero real sockets in the
// codec/watch sections, no browser ever launches except the fake-exe
// scripts (POSIX sh, no network — the reference's own design).
package chrometabs

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

const fifthURL = "https://walled.example.net/article"
const fifthErr = "HTTP 403 — bot defense (server: cloudflare; challenge)"

var realHTMLDoc = "<html><head><title>The Real Page</title></head><body>" +
	strings.Repeat("r", 500) + "</body></html>"

// ---------------------------------------------------------------------------
// §1. the WebSocket frame codec (pure)
// ---------------------------------------------------------------------------

func TestClientFrameIsMaskedSmall(t *testing.T) {
	f := EncodeClientFrame([]byte("hello"), 0x1)
	if f[0]&0x80 != 0x80 {
		t.Errorf("FIN not set: %d", f[0])
	}
	if f[0]&0x0F != 0x1 {
		t.Errorf("not a text frame: %d", f[0]&0x0F)
	}
	if f[1]&0x80 != 0x80 {
		t.Errorf("MASK not set: %d", f[1])
	}
	if f[1]&0x7F != 5 {
		t.Errorf("length bits: %d", f[1]&0x7F)
	}
	frames, rest, err := DecodeFrames(f, MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(frames) != 1 || !frames[0].Fin || frames[0].Opcode != 1 ||
		string(frames[0].Payload) != "hello" {
		t.Errorf("decoded frame: %+v", frames)
	}
	if len(rest) != 0 {
		t.Errorf("rest: %q", rest)
	}
}

func TestClientFrame16BitLength(t *testing.T) {
	payload := []byte(strings.Repeat("x", 300))
	f := EncodeClientFrame(payload, 0x1)
	if f[1]&0x7F != 126 {
		t.Errorf("16-bit marker missing: %d", f[1]&0x7F)
	}
	frames, _, err := DecodeFrames(f, MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(frames[0].Payload) != string(payload) {
		t.Errorf("payload mismatch")
	}
}

func TestClientFrame64BitLength(t *testing.T) {
	payload := []byte(strings.Repeat("x", 70000))
	raw := []byte{0x80 | 0x1, 127, 0, 0, 0, 0, 0, 0, 0, 0}
	n := uint64(len(payload))
	for shift := uint(56); ; shift -= 8 {
		raw[9-shift/8] = byte(n >> shift)
		if shift == 0 {
			break
		}
	}
	raw = append(raw, payload...)
	frames, rest, err := DecodeFrames(raw, MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if string(frames[0].Payload) != string(payload) {
		t.Errorf("payload mismatch")
	}
	if len(rest) != 0 {
		t.Errorf("rest: %q", rest)
	}
}

func TestSplitFeedYieldsNothingUntilComplete(t *testing.T) {
	f := EncodeClientFrame([]byte(`{"id": 1}`), 0x1)
	cut := len(f) / 2
	frames, rest, err := DecodeFrames(f[:cut], MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode half: %v", err)
	}
	if len(frames) != 0 {
		t.Errorf("half feed decoded frames: %+v", frames)
	}
	if string(rest) != string(f[:cut]) {
		t.Errorf("half rest: %q", rest)
	}
	frames, _, err = DecodeFrames(append(rest, f[cut:]...), MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode done: %v", err)
	}
	if string(frames[0].Payload) != `{"id": 1}` {
		t.Errorf("completed payload: %q", frames[0].Payload)
	}
}

func TestFragmentedMessageAndControlFrame(t *testing.T) {
	data := append([]byte{0x1, 3}, []byte("abc")...) // fragment 1
	data = append(data, 0x89, 2, 'h', 'i')           // ping between fragments
	data = append(data, 0x80, 3, 'd', 'e', 'f')      // fragment 2 (FIN)
	frames, rest, err := DecodeFrames(data, MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	want := []Frame{
		{Fin: false, Opcode: 1, Payload: []byte("abc")},
		{Fin: true, Opcode: 9, Payload: []byte("hi")},
		{Fin: true, Opcode: 0, Payload: []byte("def")},
	}
	if len(frames) != len(want) {
		t.Fatalf("frames: %+v", frames)
	}
	for i, w := range want {
		if frames[i].Fin != w.Fin || frames[i].Opcode != w.Opcode ||
			string(frames[i].Payload) != string(w.Payload) {
			t.Errorf("frame %d: %+v want %+v", i, frames[i], w)
		}
	}
	if len(rest) != 0 {
		t.Errorf("rest: %q", rest)
	}
}

func TestRunawayLengthRaises(t *testing.T) {
	raw := []byte{0x80 | 0x1, 127, 0, 0, 0, 0, 0, 0, 0, 0}
	n := uint64(1) << 40
	for shift := uint(56); ; shift -= 8 {
		raw[9-shift/8] = byte(n >> shift)
		if shift == 0 {
			break
		}
	}
	_, _, err := DecodeFrames(raw, 1024)
	if err == nil {
		t.Fatalf("no error from the runaway guard")
	}
	if !asCDP(err) {
		t.Errorf("the guard must raise CDPError, got %T", err)
	}
}

func TestCloseFrameDecodes(t *testing.T) {
	frames, _, err := DecodeFrames([]byte{0x88, 0}, MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if len(frames) != 1 || !frames[0].Fin || frames[0].Opcode != 8 ||
		len(frames[0].Payload) != 0 {
		t.Errorf("close frame: %+v", frames)
	}
}

// ---------------------------------------------------------------------------
// §2. the DevTools client (a scripted fake connection — no real socket)
// ---------------------------------------------------------------------------

func TestHandshakeRefusalRaises(t *testing.T) {
	conn := newFakeNetConn([]byte("HTTP/1.1 403 Forbidden\r\n\r\n"))
	_, err := newCDPSocketConn(conn, "127.0.0.1", 9222, "/devtools/page/1",
		DefaultSocketTimeoutS)
	if err == nil {
		t.Fatalf("no error from a refused handshake")
	}
	if !asCDP(err) {
		t.Errorf("a refused handshake must raise CDPError, got %T: %v", err, err)
	}
}

func TestHandshakeRequestShape(t *testing.T) {
	ws, conn := newGoldenSocket(t)
	ws.Close()
	req := string(conn.sent[0])
	for _, want := range []string{
		"GET /devtools/page/1 HTTP/1.1",
		"Upgrade: websocket",
		"Sec-WebSocket-Key:",
		"Sec-WebSocket-Version: 13",
	} {
		if !strings.Contains(req, want) {
			t.Errorf("handshake request missing %q:\n%s", want, req)
		}
	}
	// no Origin header — the raw-socket client is not a webpage
	if strings.Contains(req, "Origin:") {
		t.Errorf("handshake request must carry no Origin:\n%s", req)
	}
}

func TestCallRoundTripsAndSkipsEvents(t *testing.T) {
	resp := `{"id": 1, "result": {"result": {"value": "complete"}}}`
	event := `{"method": "Page.frameStartedLoading", "params": {}}`
	ws, conn := newGoldenSocket(t, srvJSON(event), srvJSON(resp))
	params := storage.NewOrderedMap()
	params.Set("expression", "1")
	out, err := ws.Call("Runtime.evaluate", params, 0)
	if err != nil {
		t.Fatalf("call: %v", err)
	}
	if v, ok := out.Get("result"); !ok {
		t.Fatalf("no result: %v", out)
	} else if om, ok := v.(*storage.OrderedMap); !ok {
		t.Fatalf("result not an object: %v", v)
	} else if vv, _ := om.Get("value"); fmt.Sprint(vv) != "complete" {
		t.Errorf("value: %v", vv)
	}
	// the command went out as a MASKED text frame carrying the id
	sent := conn.sent[1]
	if sent[0]&0x0F != 0x1 {
		t.Errorf("not a text frame: %d", sent[0]&0x0F)
	}
	if sent[1]&0x80 != 0x80 {
		t.Errorf("the client frame is not masked")
	}
	frames, _, err := DecodeFrames(sent, MaxMessageBytes)
	if err != nil {
		t.Fatalf("decode sent: %v", err)
	}
	if !strings.Contains(string(frames[0].Payload), `"id": 1`) {
		t.Errorf("the sent frame carries no id: %s", frames[0].Payload)
	}
}

func TestPingAnsweredWithMaskedPong(t *testing.T) {
	okResp := `{"id": 1, "result": {}}`
	ws, conn := newGoldenSocket(t,
		srvFrame([]byte("beat"), 0x9), srvJSON(okResp))
	params := storage.NewOrderedMap()
	params.Set("expression", "1")
	if _, err := ws.Call("Runtime.evaluate", params, 0); err != nil {
		t.Fatalf("call: %v", err)
	}
	pongs := 0
	for _, sent := range conn.sent[1:] {
		if sent[0]&0x0F == 0xA {
			pongs++
			if sent[1]&0x80 != 0x80 {
				t.Errorf("the pong is not masked (the client law)")
			}
			frames, _, err := DecodeFrames(sent, MaxMessageBytes)
			if err != nil {
				t.Fatalf("decode pong: %v", err)
			}
			if string(frames[0].Payload) != "beat" {
				t.Errorf("pong payload: %q", frames[0].Payload)
			}
		}
	}
	if pongs != 1 {
		t.Errorf("pongs sent: %d", pongs)
	}
}

func TestCloseFrameRaises(t *testing.T) {
	ws, _ := newGoldenSocket(t, srvFrame([]byte(""), 0x8))
	if _, err := ws.recvMessage(); err == nil {
		t.Fatalf("no error from a close frame")
	} else {
		var ce *CDPClosedError
		if !asCDP(err) {
			t.Errorf("close must raise the CDP family, got %T", err)
		}
		_ = ce
	}
}

func TestErrorResponseRaises(t *testing.T) {
	bad := `{"id": 1, "error": {"message": "no such target"}}`
	ws, _ := newGoldenSocket(t, srvJSON(bad))
	_, err := ws.Call("Runtime.evaluate", nil, 0)
	if err == nil {
		t.Fatalf("no error from an error response")
	}
	if !asCDP(err) {
		t.Errorf("error response must raise CDPError, got %T", err)
	}
	if !strings.Contains(err.Error(), "no such target") {
		t.Errorf("error message: %s", err.Error())
	}
}

func TestStalledSocketRaisesTimeout(t *testing.T) {
	ws, _ := newGoldenSocket(t) // the handshake answers, then nothing
	_, err := ws.Call("Runtime.evaluate", nil, 0.2)
	if err == nil {
		t.Fatalf("no error from a stalled socket")
	}
	var te *CDPTimeoutError
	if !asCDP(err) {
		t.Errorf("stall must raise the CDP family, got %T", err)
	}
	if e, ok := err.(*CDPTimeoutError); !ok {
		_ = te
		t.Errorf("stall must raise CDPTimeoutError, got %T", err)
	} else if !strings.Contains(e.Error(), "timed out") {
		t.Errorf("timeout message: %s", e.Error())
	}
}

func TestEvaluateReturnsValue(t *testing.T) {
	resp := `{"id": 1, "result": {"result": {"type": "string", "value": "full"}}}`
	ws, _ := newGoldenSocket(t, srvJSON(resp))
	v, err := ws.Evaluate("document.readyState", 0)
	if err != nil {
		t.Fatalf("evaluate: %v", err)
	}
	if v != "full" {
		t.Errorf("evaluate: %q", v)
	}
}

func TestFragmentedResponseReassembles(t *testing.T) {
	body := `{"id": 1, "result": {}}`
	ws, _ := newGoldenSocket(t,
		append([]byte{0x1, 2}, []byte(body[:2])...),
		append([]byte{0x80, byte(len(body) - 2)}, []byte(body[2:])...))
	if _, err := ws.Call("Runtime.evaluate", nil, 0); err != nil {
		t.Fatalf("fragmented call: %v", err)
	}
}

// --- the load-watch harness ---------------------------------------------

// fakeTabSession is the reference tests' _FakeSession: a fixed port
// and a record of the tabs the door closed.
type fakeTabSession struct {
	port   int
	closed []string
}

func (s *fakeTabSession) NewTab(rawURL string) (TabInfo, error) {
	return TabInfo{ID: "T1", WsURL: "ws://127.0.0.1:9222/devtools/page/T1"},
		nil
}
func (s *fakeTabSession) CloseTab(targetID string) {
	s.closed = append(s.closed, targetID)
}
func (s *fakeTabSession) Alive() bool { return true }
func (s *fakeTabSession) Port() int   { return s.port }
func (s *fakeTabSession) Close()      {}

// watchSeams patches the timing laws for one scripted fetch and
// restores them after (the reference's mock.patch.object shape).
func watchSeams(t *testing.T, settle, poll, reconnectPause float64,
	antiBot bool, ws cdpConn) {
	t.Helper()
	oldSettle, oldPoll, oldPause, oldAnti, oldSocket :=
		PageSettleS, pollCadenceS, reconnectPauseS, antiBotPatch, newTabSocketFn
	PageSettleS = settle
	pollCadenceS = poll
	reconnectPauseS = reconnectPause
	antiBotPatch = antiBot
	newTabSocketFn = func(host string, port int, path string) (cdpConn, error) {
		return ws, nil
	}
	t.Cleanup(func() {
		PageSettleS, pollCadenceS, reconnectPauseS = oldSettle, oldPoll, oldPause
		antiBotPatch, newTabSocketFn = oldAnti, oldSocket
	})
}

// probeFrame is the load watch's probe answer: 'state|href' at id.
func probeFrame(id int, state, href string) []byte {
	return srvJSON(fmt.Sprintf(
		`{"id": %d, "result": {"result": {"value": %s}}}`, id,
		pyJSONString(state+"|"+href)))
}

// domFrame is a DOM/title answer at id.
func domFrame(id int, value string) []byte {
	return srvJSON(fmt.Sprintf(
		`{"id": %d, "result": {"result": {"value": %s}}}`, id,
		pyJSONString(value)))
}

// fetchWatched drives one scripted fetch through fetchOpenedTab (the
// reference's _fetch): the socket is pre-built against the scripted
// connection (handshake consumed), the seams patched (the poll
// cadence stays 0.4 — the reference's own scripts are timed against
// it: the second probe arrives after the settle window), the tab
// fetched.
func fetchWatched(t *testing.T, timeout float64, settle float64,
	antiBot bool, chunks ...[]byte) (PageResult, *fakeTabSession) {
	t.Helper()
	conn := newFakeNetConn(append([][]byte{handshakeAnswer}, chunks...)...)
	ws, err := newCDPSocketConn(conn, "127.0.0.1", 9222, "/devtools/page/T1",
		DefaultSocketTimeoutS)
	if err != nil {
		t.Fatalf("CDPSocket: %v", err)
	}
	session := &fakeTabSession{port: 9222}
	tab := TabInfo{ID: "T1", WsURL: "ws://127.0.0.1:9222/devtools/page/T1"}
	watchSeams(t, settle, pollCadenceS, 0.01, antiBot, ws)
	page := fetchOpenedTab(session, fifthURL, tab, timeout, nil)
	return page, session
}

func TestMidNavigationContextDestructionIsLoadingNotFailure(t *testing.T) {
	// "Execution context was destroyed" mid-navigation is a page STILL
	// LOADING, not a failure: the poll keeps going, and a context swap
	// landing exactly on the final DOM read gets ONE retry.
	ctxErr := func(id int) []byte {
		return srvJSON(fmt.Sprintf(
			`{"id": %d, "error": {"message": %s}}`, id,
			pyJSONString("Execution context was destroyed")))
	}
	html := "<html><body>" + strings.Repeat("z", 300) + "</body></html>"
	// ids: 1=Page.enable(ate by the anti-bot's swallowed error)
	// 2=probe(ok) 3=probe(settle) 4=DOM(err) 5=DOM(ok) 6=title
	page, session := fetchWatched(t, 10.0, 0.05, true,
		ctxErr(1),
		probeFrame(2, "complete", fifthURL),
		probeFrame(3, "complete", fifthURL),
		ctxErr(4),
		domFrame(5, html),
		domFrame(6, "Retried"))
	if !page.OK {
		t.Fatalf("page: %+v", page)
	}
	if !strings.Contains(page.HTML, strings.Repeat("z", 100)) {
		t.Errorf("html: %q", page.HTML[:100])
	}
	if page.Title != "Retried" {
		t.Errorf("title: %q", page.Title)
	}
	if len(session.closed) != 1 || session.closed[0] != "T1" {
		t.Errorf("closed tabs: %v", session.closed)
	}
}

// ---------------------------------------------------------------------------
// §2b. the page verdict (the pure law)
// ---------------------------------------------------------------------------

func TestVerdictRealPagePasses(t *testing.T) {
	ok, reason := PageIsReal(realHTMLDoc, fifthURL, "The Real Page")
	if !ok {
		t.Errorf("a real page must pass: %s", reason)
	}
	if reason != "" {
		t.Errorf("reason must be empty: %q", reason)
	}
}

func TestVerdictChromeErrorURLIsNotTheSite(t *testing.T) {
	ok, reason := PageIsReal(
		"<html><body>"+strings.Repeat("e", 500)+"</body></html>",
		"chrome-error://chromewebdata/", "")
	if ok {
		t.Errorf("chrome-error must not be the site")
	}
	if !strings.Contains(reason, "Chrome's own page") {
		t.Errorf("reason: %q", reason)
	}
}

func TestVerdictAboutBlankIsNotTheSite(t *testing.T) {
	ok, reason := PageIsReal("<html></html>", "about:blank", "")
	if ok || !strings.Contains(reason, "Chrome's own page") {
		t.Errorf("about:blank: ok=%v reason=%q", ok, reason)
	}
}

func TestVerdictEmptyHrefIsNotTheSite(t *testing.T) {
	ok, reason := PageIsReal(realHTMLDoc, "", "")
	if ok || !strings.Contains(reason, "empty URL") {
		t.Errorf("empty href: ok=%v reason=%q", ok, reason)
	}
}

func TestVerdictErrCodeNamesTheErrorPage(t *testing.T) {
	html := "<html><head><title>walled.example.net</title></head>" +
		"<body><h1>This site can’t be reached</h1>" +
		"<p>walled.example.net unexpectedly refused… " +
		"ERR_CONNECTION_RESET</p></body></html>"
	ok, reason := PageIsReal(html, fifthURL, "walled.example.net")
	if ok {
		t.Errorf("an ERR_ page must not be real")
	}
	if !strings.Contains(reason, "ERR_CONNECTION_RESET") ||
		!strings.Contains(reason, "Chrome's error page") {
		t.Errorf("reason: %q", reason)
	}
}

func TestVerdictCrashPageIsACrash(t *testing.T) {
	html := "<html><head><title>Aw, Snap!</title></head><body>" +
		strings.Repeat("c", 500) + "</body></html>"
	ok, reason := PageIsReal(html, fifthURL, "Aw, Snap!")
	if ok || !strings.Contains(reason, "crashed") {
		t.Errorf("crash: ok=%v reason=%q", ok, reason)
	}
}

func TestVerdictChallengeTitleIsAChallenge(t *testing.T) {
	html := "<html><head><title>Just a moment...</title></head>" +
		"<body>" + strings.Repeat("j", 500) + "</body></html>"
	ok, reason := PageIsReal(html, fifthURL, "Just a moment...")
	if ok || !strings.Contains(reason, "challenge") {
		t.Errorf("challenge title: ok=%v reason=%q", ok, reason)
	}
}

func TestVerdictChallengeMachineryIsAChallenge(t *testing.T) {
	html := "<html><head><title>Some title</title></head><body>" +
		"<script src=\"/cdn-cgi/challenge-platform/h/b/or.js\">" +
		"</script>" + strings.Repeat("j", 400) + "</body></html>"
	ok, reason := PageIsReal(html, fifthURL, "Some title")
	if ok || !strings.Contains(reason, "challenge") {
		t.Errorf("challenge machinery: ok=%v reason=%q", ok, reason)
	}
}

// ---------------------------------------------------------------------------
// §2b. the load watch (scripted connections; the LAW is tested, not
// the clock — the settle is patched tiny)
// ---------------------------------------------------------------------------

func TestWatchWaitedRealPageIsDelivered(t *testing.T) {
	page, session := fetchWatched(t, 10.0, 0.05, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		domFrame(3, realHTMLDoc),
		domFrame(4, "The Real Page"))
	if !page.OK {
		t.Fatalf("page: %+v", page)
	}
	if !strings.Contains(page.HTML, strings.Repeat("r", 100)) {
		t.Errorf("html missing the body")
	}
	if page.Title != "The Real Page" {
		t.Errorf("title: %q", page.Title)
	}
	if len(session.closed) != 1 || session.closed[0] != "T1" {
		t.Errorf("closed tabs: %v", session.closed)
	}
}

func TestWatchChromeErrorPageIsNeverADelivery(t *testing.T) {
	// the EXACT false positive the owner reported: readyState
	// 'complete', bytes of DOM — and the DOM is Chrome's own error page
	errHTML := "<html><head><title>walled.example.net</title></head>" +
		"<body><h1>This site can’t be reached</h1>" +
		"<p>ERR_CONNECTION_RESET</p></body></html>"
	page, session := fetchWatched(t, 10.0, 0.05, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		domFrame(3, errHTML),
		domFrame(4, "walled.example.net"))
	if page.OK {
		t.Fatalf("an error page was delivered: %+v", page)
	}
	if !strings.Contains(page.Error, "ERR_CONNECTION_RESET") ||
		!strings.Contains(page.Error, "Chrome's error page") {
		t.Errorf("error: %q", page.Error)
	}
	if len(session.closed) != 1 || session.closed[0] != "T1" {
		t.Errorf("closed tabs: %v", session.closed)
	}
}

func TestWatchCrashPageIsNeverADelivery(t *testing.T) {
	crashHTML := "<html><head><title>Aw, Snap!</title></head>" +
		"<body><div id=\"crash\">" + strings.Repeat("c", 500) +
		"</div></body></html>"
	page, _ := fetchWatched(t, 10.0, 0.05, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		domFrame(3, crashHTML),
		domFrame(4, "Aw, Snap!"))
	if page.OK {
		t.Fatalf("a crash page was delivered: %+v", page)
	}
	if !strings.Contains(page.Error, "crashed") {
		t.Errorf("error: %q", page.Error)
	}
}

func TestWatchChallengeThatPassesIsWaitedFor(t *testing.T) {
	// the door's whole point: the tab loads a challenge, the owner's
	// REAL Chrome passes it (the redirect swaps the page) — the settle
	// restarts on the href change and the REAL page is taken
	challenge := "<html><head><title>Just a moment...</title>" +
		"</head><body>" + strings.Repeat("j", 500) + "</body></html>"
	page, session := fetchWatched(t, 10.0, 0.05, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		domFrame(3, challenge),
		domFrame(4, "Just a moment..."),
		probeFrame(5, "complete", fifthURL),
		probeFrame(6, "complete", fifthURL),
		domFrame(7, realHTMLDoc),
		domFrame(8, "The Real Page"))
	if !page.OK {
		t.Fatalf("page: %+v", page)
	}
	if !strings.Contains(page.HTML, strings.Repeat("r", 100)) {
		t.Errorf("html missing the real body")
	}
	if page.Title != "The Real Page" {
		t.Errorf("title: %q", page.Title)
	}
	if len(session.closed) != 1 || session.closed[0] != "T1" {
		t.Errorf("closed tabs: %v", session.closed)
	}
}

func TestWatchChallengeThatNeverPassesIsNamed(t *testing.T) {
	challenge := "<html><head><title>Just a moment...</title>" +
		"</head><body>" + strings.Repeat("j", 500) + "</body></html>"
	page, _ := fetchWatched(t, 1.5, 0.05, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		domFrame(3, challenge),
		domFrame(4, "Just a moment..."),
		probeFrame(5, "complete", fifthURL),
		probeFrame(6, "complete", fifthURL),
		domFrame(7, challenge),
		domFrame(8, "Just a moment..."))
	if page.OK {
		t.Fatalf("a challenge was delivered: %+v", page)
	}
	if !strings.Contains(page.Error, "challenge") {
		t.Errorf("error: %q", page.Error)
	}
	if strings.Contains(page.HTML, "r") {
		t.Errorf("html leaked: %q", page.HTML)
	}
}

func TestWatchPageNeverFinishesIsNamed(t *testing.T) {
	page, _ := fetchWatched(t, 1.5, 0.05, false,
		probeFrame(1, "loading", fifthURL),
		probeFrame(2, "loading", fifthURL),
		probeFrame(3, "loading", fifthURL),
		probeFrame(4, "loading", fifthURL))
	if page.OK {
		t.Fatalf("an unfinished page was delivered: %+v", page)
	}
	if !strings.Contains(page.Error, "never finished loading") {
		t.Errorf("error: %q", page.Error)
	}
}

func TestWatchTabStuckOnAboutBlankIsNamed(t *testing.T) {
	page, _ := fetchWatched(t, 1.5, 0.05, false,
		probeFrame(1, "complete", "about:blank"),
		probeFrame(2, "complete", "about:blank"),
		probeFrame(3, "complete", "about:blank"),
		probeFrame(4, "complete", "about:blank"))
	if page.OK {
		t.Fatalf("about:blank was delivered: %+v", page)
	}
	if !strings.Contains(page.Error, "Chrome's own page") {
		t.Errorf("error: %q", page.Error)
	}
}

func TestWatchSettleRestartsOnHrefChange(t *testing.T) {
	// a late redirect after 'complete': the first URL settles halfway,
	// the href swaps — the settle restarts and the page at the FINAL
	// url is what gets taken (settle 1.2s so the swap lands mid-settle,
	// exactly the real-world shape)
	finalURL := "https://walled.example.net/article?page=2"
	page, _ := fetchWatched(t, 10.0, 1.2, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		probeFrame(3, "complete", finalURL), // the swap (settle restarts)
		probeFrame(4, "complete", finalURL),
		probeFrame(5, "complete", finalURL),
		probeFrame(6, "complete", finalURL), // the settle met
		domFrame(7, realHTMLDoc),
		domFrame(8, "The Real Page"))
	if !page.OK {
		t.Fatalf("page: %+v", page)
	}
	if !strings.Contains(page.HTML, strings.Repeat("r", 100)) {
		t.Errorf("html missing the real body")
	}
}

func TestWatchEmptyDomIsNotADelivery(t *testing.T) {
	page, _ := fetchWatched(t, 10.0, 0.05, false,
		probeFrame(1, "complete", fifthURL),
		probeFrame(2, "complete", fifthURL),
		domFrame(3, "<html></html>"),
		domFrame(4, ""))
	if page.OK {
		t.Fatalf("an empty DOM was delivered: %+v", page)
	}
	if !strings.Contains(page.Error, "rendered empty") {
		t.Errorf("error: %q", page.Error)
	}
}

// ---------------------------------------------------------------------------
// §3. the candidates — which links earn the offer
// ---------------------------------------------------------------------------

func candidateRows() []websitepipeline.Result {
	return []websitepipeline.Result{
		{URL: fifthURL, Outcome: "review", FetchStatus: "failed",
			Error: "fetch failed: " + fifthErr},
	}
}

func TestCandidatesFetchFailedReviewRowsAreCandidates(t *testing.T) {
	out := CollectFailedFetchLinks(candidateRows(), nil)
	if len(out) != 1 || out[0].URL != fifthURL ||
		out[0].Error != "fetch failed: "+fifthErr {
		t.Errorf("out: %+v", out)
	}
}

func TestCandidatesHardFailuresAreCandidates(t *testing.T) {
	out := CollectFailedFetchLinks([]websitepipeline.Result{
		{URL: "https://x.example/a", Outcome: "failed", Error: "boom"},
	}, nil)
	if len(out) != 1 || out[0].Error != "boom" {
		t.Errorf("out: %+v", out)
	}
}

func TestCandidatesSuccessesAndPartialsAreNot(t *testing.T) {
	out := CollectFailedFetchLinks([]websitepipeline.Result{
		{URL: "https://ok.example/a", Outcome: "processed"},
		{URL: "https://low.example/b", Outcome: "review",
			FetchStatus: "full", Error: "low confidence"},
		{URL: "https://arch.example/c", Outcome: "review",
			FetchStatus: "partial", Error: "archived copy"},
		{URL: "https://skip.example/d", Outcome: "skipped"},
	}, nil)
	if len(out) != 0 {
		t.Errorf("out: %+v", out)
	}
}

func TestCandidatesDedupeByCanonicalAndLoopbackDropped(t *testing.T) {
	out := CollectFailedFetchLinks([]websitepipeline.Result{
		{URL: fifthURL, Outcome: "review", FetchStatus: "failed",
			Error: "e1"},
		{URL: fifthURL + "?utm_source=bot", Outcome: "review",
			FetchStatus: "failed", Error: "e2"},
		{URL: "http://localhost:8080/x", Outcome: "review",
			FetchStatus: "failed", Error: "e3"},
		{URL: "not-a-link", Outcome: "review", FetchStatus: "failed",
			Error: "e4"},
	}, nil)
	if len(out) != 1 {
		t.Fatalf("out: %+v", out)
	}
	if out[0].URL != fifthURL || out[0].Error != "e1" {
		t.Errorf("out[0]: %+v (first wins)", out[0])
	}
}

func TestCandidatesDismissedNeverNag(t *testing.T) {
	out := CollectFailedFetchLinks(candidateRows(),
		func(string) bool { return true })
	if len(out) != 0 {
		t.Errorf("a dismissed link must never nag: %+v", out)
	}
}

func TestCandidatesDismissedCheckFailureOffersThem(t *testing.T) {
	out := CollectFailedFetchLinks(candidateRows(),
		func(string) bool { panic("no db") })
	if len(out) != 1 {
		t.Errorf("an unaskable state question must offer the link: %+v", out)
	}
}

// ---------------------------------------------------------------------------
// §4. the session helpers + the Chrome process (a fake exe, POSIX only)
// ---------------------------------------------------------------------------

func TestLoopbackLaw(t *testing.T) {
	for _, u := range []string{
		"http://localhost:3000/x", "http://127.0.0.1:9/x",
		"https://machine.local/y",
	} {
		if !IsLoopbackURL(u) {
			t.Errorf("is_loopback_url(%q) must be true", u)
		}
	}
	if IsLoopbackURL(fifthURL) {
		t.Errorf("is_loopback_url(%q) must be false", fifthURL)
	}
}

func TestPickFreePort(t *testing.T) {
	p := PickFreePort()
	if p <= 0 || p >= 65536 {
		t.Errorf("port out of range: %d", p)
	}
}

// fakeChromeExe writes a shell script that prints the given stderr
// lines (then sleeps) — optionally refusing to start without the
// required flag (the ladder's proof). No network is ever made.
func fakeChromeExe(t *testing.T, lines []string, then, require string) string {
	t.Helper()
	if runtimeOSIsWindows() {
		t.Skip("the fake-exe session test needs POSIX sh")
	}
	dir := t.TempDir()
	path := filepath.Join(dir, "chrome")
	var b strings.Builder
	b.WriteString("#!/bin/sh\n")
	if require != "" {
		fmt.Fprintf(&b, "case \" $* \" in\n"+
			"  *\" %s \"*) ;;\n"+
			"  *) exit 7 ;;\n"+
			"esac\n", require)
	}
	for _, line := range lines {
		// the DevTools line rides STDERR (stdout is DEVNULL in the
		// session launcher — exactly like real Chrome)
		fmt.Fprintf(&b, "echo \"%s\" 1>&2\n", line)
	}
	b.WriteString(then + "\n")
	if err := os.WriteFile(path, []byte(b.String()), 0o755); err != nil {
		t.Fatalf("write exe: %v", err)
	}
	return path
}

func runtimeOSIsWindows() bool { return goosIsWindows }

func newFakeSessionLogs(t *testing.T) (LogFn, *[]string) {
	t.Helper()
	logs := &[]string{}
	return func(msg, level string) { *logs = append(*logs, msg) }, logs
}

func TestChromeSessionEndpointLineParsedAndProfileErased(t *testing.T) {
	exe := fakeChromeExe(t,
		[]string{"DevTools listening on ws://127.0.0.1:59999/devtools/" +
			"browser/abc-123"}, "sleep 30", "")
	log, logs := newFakeSessionLogs(t)
	session, err := NewChromeSession(exe, log, DefaultPageTimeoutS, "",
		false, nil, false)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	profile := session.profile
	defer session.Close()
	if session.Port() != 59999 {
		t.Errorf("port: %d", session.Port())
	}
	if !session.Alive() {
		t.Errorf("the fake session must be alive")
	}
	found := false
	for _, m := range *logs {
		if strings.Contains(m, "Chrome session live") {
			found = true
		}
	}
	if !found {
		t.Errorf("the live log missing: %v", *logs)
	}
	session.Close()
	if profile == "" {
		t.Fatalf("no profile")
	}
	if _, err := os.Stat(profile); !os.IsNotExist(err) {
		t.Errorf("the temp profile was not erased: %s", profile)
	}
	if session.Alive() {
		t.Errorf("the closed session must not be alive")
	}
}

func TestChromeSessionDeadExeRaisesHonestly(t *testing.T) {
	exe := fakeChromeExe(t, nil, "exit 0", "")
	_, err := NewChromeSession(exe, nil, DefaultPageTimeoutS, "",
		false, nil, false)
	if err == nil {
		t.Fatalf("a dead exe must raise")
	}
	if !asCDP(err) {
		t.Errorf("a dead exe must raise CDPError, got %T: %v", err, err)
	}
}

func TestChromeSessionSilentExeRaisesHonestly(t *testing.T) {
	exe := fakeChromeExe(t, []string{"nothing useful"}, "exit 1", "")
	_, err := NewChromeSession(exe, nil, DefaultPageTimeoutS, "",
		false, nil, false)
	if err == nil || !asCDP(err) {
		t.Errorf("a silent exe must raise CDPError, got %v", err)
	}
}

func TestChromeSessionContainerFlagLadderSecondAttemptWins(t *testing.T) {
	exe := fakeChromeExe(t,
		[]string{"DevTools listening on ws://127.0.0.1:59998/devtools/" +
			"browser/tier-2"}, "sleep 30", "--no-sandbox")
	log, logs := newFakeSessionLogs(t)
	session, err := NewChromeSession(exe, log, DefaultPageTimeoutS, "",
		false, nil, false)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	defer session.Close()
	if session.Port() != 59998 {
		t.Errorf("port: %d", session.Port())
	}
	joined := strings.Join(*logs, "\n")
	if !strings.Contains(joined, "container flags") {
		t.Errorf("the container log missing: %v", *logs)
	}
	if strings.Contains(joined, "HEADLESS") {
		t.Errorf("the headless log must not appear: %v", *logs)
	}
}

func TestChromeSessionHeadlessLadderThirdAttemptWins(t *testing.T) {
	exe := fakeChromeExe(t,
		[]string{"DevTools listening on ws://127.0.0.1:59997/devtools/" +
			"browser/tier-3"}, "sleep 30", "--headless=new")
	log, logs := newFakeSessionLogs(t)
	session, err := NewChromeSession(exe, log, DefaultPageTimeoutS, "",
		false, nil, false)
	if err != nil {
		t.Fatalf("session: %v", err)
	}
	defer session.Close()
	if session.Port() != 59997 {
		t.Errorf("port: %d", session.Port())
	}
	joined := strings.Join(*logs, "\n")
	if !strings.Contains(joined, "container flags") ||
		!strings.Contains(joined, "HEADLESS") {
		t.Errorf("the ladder logs missing: %v", *logs)
	}
}

// ---------------------------------------------------------------------------
// §5. the fetch driver's honest misses (no browser ever launches)
// ---------------------------------------------------------------------------

func withFindChrome(t *testing.T, chrome string) {
	t.Helper()
	old := findChromeFn
	findChromeFn = func() string { return chrome }
	t.Cleanup(func() { findChromeFn = old })
}

func TestDriverNoChromeIsAPerURLFailure(t *testing.T) {
	withFindChrome(t, "")
	out := FetchPagesViaChrome([]string{fifthURL}, FetchOptions{})
	if len(out) != 1 {
		t.Fatalf("out: %+v", out)
	}
	if out[0].OK {
		t.Errorf("a no-Chrome fetch must fail")
	}
	if !strings.Contains(strings.ToLower(out[0].Error), "not found") {
		t.Errorf("error: %q", out[0].Error)
	}
}

func TestDriverLoopbackAndNonHTTPNeverOpen(t *testing.T) {
	withFindChrome(t, "")
	out := FetchPagesViaChrome([]string{
		"http://localhost:1/x", "ftp://nope.example/f"}, FetchOptions{})
	if len(out) != 2 {
		t.Fatalf("out: %+v", out)
	}
	if !strings.Contains(out[0].Error, "loopback") {
		t.Errorf("loopback error: %q", out[0].Error)
	}
	if !strings.Contains(out[1].Error, "http(s)") {
		t.Errorf("non-http error: %q", out[1].Error)
	}
}

func TestDriverDryRunNeverLaunches(t *testing.T) {
	dryrun.Enable()
	defer dryrun.Disable()
	defer dryrun.Clear()
	old := chromeSessionCtor
	chromeSessionCtor = func(exe string, log LogFn, pageTimeoutS float64,
		profileDir string, persist bool, extraFlags []string,
		realProfile bool) (TabSession, error) {
		panic("launched")
	}
	defer func() { chromeSessionCtor = old }()
	out := FetchPagesViaChrome([]string{fifthURL}, FetchOptions{})
	if len(out) != 1 || out[0].OK {
		t.Fatalf("out: %+v", out)
	}
	if !strings.Contains(out[0].Error, "dry-run") {
		t.Errorf("error: %q", out[0].Error)
	}
}

// ---------------------------------------------------------------------------
// §6. the delivery — pages into the fourth door's folder, 'auto' stamped
// ---------------------------------------------------------------------------

// fifthCase is the reference's _FifthCase: a temp vault, cleaned up.
func fifthCase(t *testing.T) string {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	t.Cleanup(func() {
		dryrun.Disable()
		dryrun.Clear()
	})
	vault := filepath.Join(t.TempDir(), "websites")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatalf("vault: %v", err)
	}
	return vault
}

// deliverPage writes one landed page at its suggested filename.
func deliverPage(t *testing.T, vault, url string, body []byte) string {
	t.Helper()
	dir := handdelivery.HandDeliveryDir(vault)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("dir: %v", err)
	}
	path := filepath.Join(dir, handdelivery.SuggestedFilename(url))
	if err := os.WriteFile(path, body, 0o644); err != nil {
		t.Fatalf("page: %v", err)
	}
	return path
}

var fifthDefaultBody = []byte("<html><head><title>Took From Chrome</title>" +
	"</head><body>the live DOM after the render</body></html>")

func goodFetcher(urls []string, log LogFn) []PageResult {
	out := make([]PageResult, 0, len(urls))
	for _, u := range urls {
		out = append(out, PageResult{URL: u, OK: true, Title: "T",
			HTML: "<html><body>" + strings.Repeat("x", 400) +
				"</body></html>"})
	}
	return out
}

func noopLogFn(string, string) {}

func TestDeliveryWritesQueueAndPages(t *testing.T) {
	vault := fifthCase(t)
	report := DeliverPagesViaChrome(vault,
		[]Link{{URL: fifthURL, Error: fifthErr}}, nil, nil, "", nil,
		goodFetcher)
	if report.Delivered != 1 {
		t.Fatalf("delivered: %+v", report)
	}
	if len(report.URLs) != 1 || report.URLs[0] != fifthURL {
		t.Errorf("urls: %v", report.URLs)
	}
	// the queue row exists, stamped with the auto door + the wall
	queue := handdelivery.ReadQueue(vault)
	meta := queue.Meta(fifthURL)
	if meta == nil {
		t.Fatalf("no queue row for %s", fifthURL)
	}
	if door, _ := meta.Get("door"); fmt.Sprint(door) != "auto" {
		t.Errorf("door: %v", door)
	}
	if wall, _ := meta.Get("wall"); fmt.Sprint(wall) != fifthErr {
		t.Errorf("wall: %v", wall)
	}
	// the page landed at the suggested filename
	path := filepath.Join(report.Folder,
		handdelivery.SuggestedFilename(fifthURL))
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("page: %v", err)
	}
	if !strings.Contains(string(data), "<body>") {
		t.Errorf("page content: %q", data)
	}
}

func TestDeliveryCountsFailuresHonestly(t *testing.T) {
	vault := fifthCase(t)
	fetcher := func(urls []string, log LogFn) []PageResult {
		out := make([]PageResult, 0, len(urls))
		for _, u := range urls {
			out = append(out, PageResult{URL: u, OK: false,
				Error: "the page never finished loading in 45s"})
		}
		return out
	}
	report := DeliverPagesViaChrome(vault,
		[]Link{{URL: fifthURL, Error: fifthErr}}, nil, nil, "", nil, fetcher)
	if report.Delivered != 0 || report.Failed != 1 {
		t.Fatalf("report: %+v", report)
	}
	// the queue row was still written (the link is queued for the
	// fourth door — an owner may still deliver it by hand)
	if handdelivery.ReadQueue(vault).Meta(fifthURL) == nil {
		t.Errorf("the queue row must survive a failed delivery")
	}
}

func TestDeliveryFailedLinksDetailRidesTheReport(t *testing.T) {
	vault := fifthCase(t)
	fetcher := func(urls []string, log LogFn) []PageResult {
		return []PageResult{
			{URL: "https://a.example.net/1", OK: true, Title: "A",
				HTML: "<html><body>" + strings.Repeat("x", 400) +
					"</body></html>"},
			{URL: "https://b.example.net/2", OK: false,
				Error: "the tab showed Chrome's error page " +
					"(ERR_CONNECTION_RESET)"},
		}
	}
	report := DeliverPagesViaChrome(vault, []Link{
		{URL: "https://a.example.net/1"},
		{URL: "https://b.example.net/2"},
	}, nil, nil, "", nil, fetcher)
	if report.Delivered != 1 || report.Failed != 1 {
		t.Fatalf("report: %+v", report)
	}
	if len(report.FailedLinks) != 1 {
		t.Fatalf("failed_links: %+v", report.FailedLinks)
	}
	if report.FailedLinks[0].URL != "https://b.example.net/2" ||
		!strings.Contains(report.FailedLinks[0].Error,
			"ERR_CONNECTION_RESET") {
		t.Errorf("failed_links[0]: %+v", report.FailedLinks[0])
	}
}

func TestDeliveryErrorPageFromAnyFetchPathIsNeverWritten(t *testing.T) {
	// the verdict is the DELIVERY's law too: a fetch path that hands
	// back a Chrome error page is a named failure — no file, no
	// consumed row, nothing delivered
	vault := fifthCase(t)
	errPage := "<html><head><title>x.example</title></head><body>" +
		"<h1>This site can’t be reached</h1>" +
		"<p>ERR_TIMED_OUT</p></body></html>"
	fetcher := func(urls []string, log LogFn) []PageResult {
		return []PageResult{{URL: fifthURL, OK: true, Title: "x.example",
			HTML: errPage}}
	}
	report := DeliverPagesViaChrome(vault,
		[]Link{{URL: fifthURL, Error: fifthErr}}, nil, nil, "", nil, fetcher)
	if report.Delivered != 0 || report.Failed != 1 {
		t.Fatalf("report: %+v", report)
	}
	if len(report.FailedLinks) != 1 ||
		!strings.Contains(report.FailedLinks[0].Error, "ERR_TIMED_OUT") {
		t.Errorf("failed_links: %+v", report.FailedLinks)
	}
	if _, err := os.Stat(filepath.Join(report.Folder,
		handdelivery.SuggestedFilename(fifthURL))); !os.IsNotExist(err) {
		t.Errorf("an error page was written")
	}
}

func TestDeliveryCrashingFetcherNeverBreaksTheDelivery(t *testing.T) {
	vault := fifthCase(t)
	boom := func(urls []string, log LogFn) []PageResult {
		panic("chrome exploded")
	}
	report := DeliverPagesViaChrome(vault,
		[]Link{{URL: fifthURL, Error: fifthErr}}, nil, nil, "", nil, boom)
	if report.Delivered != 0 || report.Failed != 1 {
		t.Fatalf("report: %+v", report)
	}
	if !strings.Contains(report.FailedLinks[0].Error, "chrome exploded") {
		t.Errorf("failed_links[0]: %+v", report.FailedLinks[0])
	}
}

func TestDeliveryDryRunLaunchesNothingAndWritesNothing(t *testing.T) {
	vault := fifthCase(t)
	dryrun.Enable()
	defer dryrun.Disable()
	defer dryrun.Clear()
	called := false
	report := DeliverPagesViaChrome(vault,
		[]Link{{URL: fifthURL, Error: fifthErr}}, nil, nil, "", nil,
		func(urls []string, log LogFn) []PageResult {
			called = true
			return nil
		})
	if report.Delivered != 0 || called {
		t.Fatalf("the dry-run fetcher ran: report=%+v called=%v",
			report, called)
	}
	if _, err := os.Stat(handdelivery.QueuePath(vault)); !os.IsNotExist(err) {
		t.Errorf("the dry-run wrote a queue")
	}
	if _, err := os.Stat(handdelivery.HandDeliveryDir(vault)); !os.IsNotExist(err) {
		t.Errorf("the dry-run wrote the hand-delivered dir")
	}
}

func TestDeliveryEmptyLinksAndMissingVaultAreQuiet(t *testing.T) {
	vault := fifthCase(t)
	report := DeliverPagesViaChrome(vault, nil, nil, nil, "", nil, nil)
	if report.Delivered != 0 {
		t.Errorf("empty links: %+v", report)
	}
	report = DeliverPagesViaChrome("",
		[]Link{{URL: fifthURL}}, nil, nil, "", nil, nil)
	if report.Delivered != 0 {
		t.Errorf("missing vault: %+v", report)
	}
}

// ---------------------------------------------------------------------------
// §7b. the consume-side re-verification — the auto-door discard
// battery SESSION-018 deferred to this module (the PageIsRealFn seam
// is WIRED by this package's init; an auto-delivered error page is
// discarded, never consumed; the owner's saves are trusted as always)
// ---------------------------------------------------------------------------

var fifthErrPage = []byte("<html><head><title>walled.example.net</title>" +
	"</head><body><h1>This site can’t be reached</h1>" +
	"<p>ERR_CONNECTION_RESET</p></body></html>")

var fifthChallengePage = []byte(
	"<html><head><title>Just a moment...</title></head><body><script " +
		"src=\"/cdn-cgi/challenge-platform/h/b/or.js\"></script></body></html>")

func enqueueAuto(t *testing.T, vault string, urls ...string) {
	t.Helper()
	handdelivery.EnqueueHandDelivery(vault, urls,
		map[string]string{fifthURL: fifthErr}, nil, false, "auto")
}

func TestConsumeAutoErrorPageIsDiscardedNotConsumed(t *testing.T) {
	vault := fifthCase(t)
	enqueueAuto(t, vault, fifthURL)
	path := deliverPage(t, vault, fifthURL, fifthErrPage)
	var logs []string
	res := handdelivery.TakeHandDelivered(vault, fifthURL,
		func(msg, level string) { logs = append(logs, msg) }, false)
	if res != nil {
		t.Fatalf("an auto error page must not be consumed: %+v", res)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the app's own file was not discarded")
	}
	meta := handdelivery.ReadQueue(vault).Meta(fifthURL)
	if meta == nil {
		t.Fatalf("queue row gone")
	}
	if consumed, _ := meta.Get("consumed"); consumed != nil {
		t.Errorf("the discarded row must keep waiting: %v", consumed)
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "NOT the site") {
		t.Errorf("the discard log missing: %v", logs)
	}
}

func TestConsumeAutoChallengePageIsDiscardedNotConsumed(t *testing.T) {
	vault := fifthCase(t)
	enqueueAuto(t, vault, fifthURL)
	path := deliverPage(t, vault, fifthURL, fifthChallengePage)
	res := handdelivery.TakeHandDelivered(vault, fifthURL, nil, false)
	if res != nil {
		t.Fatalf("an auto challenge page must not be consumed: %+v", res)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Errorf("the app's own file was not discarded")
	}
}

func TestConsumeAutoRealPageIsConsumedAsBefore(t *testing.T) {
	vault := fifthCase(t)
	enqueueAuto(t, vault, fifthURL)
	path := deliverPage(t, vault, fifthURL, fifthDefaultBody)
	res := handdelivery.TakeHandDelivered(vault, fifthURL, nil, false)
	if res == nil {
		t.Fatalf("a real auto page must be consumed")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the record must stay: %v", err)
	}
}

func TestConsumeOwnerSavedErrorPageIsTrusted(t *testing.T) {
	// no door stamp — the owner's Ctrl+S is the owner's verdict: even
	// an error-page save is consumed, exactly as before
	vault := fifthCase(t)
	handdelivery.EnqueueHandDelivery(vault, []string{fifthURL},
		map[string]string{fifthURL: fifthErr}, nil, false, "")
	path := deliverPage(t, vault, fifthURL, fifthErrPage)
	res := handdelivery.TakeHandDelivered(vault, fifthURL, nil, false)
	if res == nil {
		t.Fatalf("the owner's save must be trusted")
	}
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the owner's file must stay: %v", err)
	}
	meta := handdelivery.ReadQueue(vault).Meta(fifthURL)
	if consumed, _ := meta.Get("consumed"); consumed == nil {
		t.Errorf("the row must be stamped consumed")
	}
}

func TestConsumeDeliveredDiscardsAutoBadPages(t *testing.T) {
	// the batch-level pass: the same law for the table consume
	vault := fifthCase(t)
	okURL := "https://ok.example.net/real"
	handdelivery.EnqueueHandDelivery(vault, []string{fifthURL, okURL},
		map[string]string{fifthURL: fifthErr, okURL: ""}, nil, false, "auto")
	deliverPage(t, vault, fifthURL, fifthErrPage)
	deliverPage(t, vault, okURL, fifthDefaultBody)
	out := handdelivery.ConsumeDelivered(vault, nil)
	if len(out) != 1 || out[0].URL != okURL {
		t.Fatalf("out: %+v", out)
	}
	if _, err := os.Stat(filepath.Join(
		handdelivery.HandDeliveryDir(vault),
		handdelivery.SuggestedFilename(fifthURL))); !os.IsNotExist(err) {
		t.Errorf("the auto error page was not discarded")
	}
}

func TestConsumePendingReasksAfterADiscard(t *testing.T) {
	// the discarded auto page leaves the queue row unconsumed and the
	// file gone — the fifth door's to-do list asks for it again
	vault := fifthCase(t)
	enqueueAuto(t, vault, fifthURL)
	deliverPage(t, vault, fifthURL, fifthErrPage)
	before := handdelivery.PendingHandLinks(vault)
	handdelivery.TakeHandDelivered(vault, fifthURL, nil, false)
	after := handdelivery.PendingHandLinks(vault)
	if len(after) != 1 || after[0].URL != fifthURL {
		t.Fatalf("after: %+v", after)
	}
	if len(before) != 0 {
		t.Errorf("the file \"landed\" until judged: %+v", before)
	}
}
