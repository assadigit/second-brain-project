// watch.go — the door KEEPS ITS GRIP and WAITS FOR THE PAGE
// (chrome_tabs.py:2138-2334, 1763-1830): the DOM read with its one
// context-swap retry, the v0.53 load watch (complete at an http(s)
// URL, still for the settle, real by the verdict), the per-tab fetch
// with its socket re-connects, and the wave engine every session and
// every proxy leg drive.
package chrometabs

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// livePage is _read_and_verdict's result shape: {ok, html, title,
// href, error}.
type livePage struct {
	OK    bool
	HTML  string
	Title string
	Href  string
	Error string
}

// readDOM mirrors _read_dom: the live DOM + the title, with the one
// context-swap retry the v0.50 door learned. (”, ”) when both
// reads fail — the caller's verdict names the page honestly. A
// non-CDP socket error is RETURNED (the reference's exception
// propagates to the fetch driver's reconnect path).
func readDOM(ws cdpConn) (string, string, error) {
	html := ""
	for retry := 0; retry < 2; retry++ {
		v, err := ws.Evaluate("document.documentElement.outerHTML", 10.0)
		if err == nil {
			html = v
			break
		}
		if !asCDP(err) {
			return "", "", err // a reset — the reconnect path
		}
		if retry == 1 {
			break
		}
		time.Sleep(time.Second)
	}
	title := ""
	if v, err := ws.Evaluate("document.title", 5.0); err == nil {
		title = v
	} else if !asCDP(err) {
		return "", "", err
	}
	return html, title, nil
}

// readAndVerdict mirrors _read_and_verdict: one DOM read judged by
// PageIsReal — the take step's own gate.
func readAndVerdict(ws cdpConn, href string) (livePage, error) {
	html, title, err := readDOM(ws)
	if err != nil {
		return livePage{}, err
	}
	ok, reason := PageIsReal(html, href, title)
	if !ok {
		return livePage{OK: false, Href: href, Error: reason}, nil
	}
	if html == "" || len([]rune(helpers.Strip(html))) < minRealHTMLChars {
		return livePage{OK: false, Title: title, Href: href,
			Error: "the page rendered empty"}, nil
	}
	return livePage{OK: true, HTML: html, Title: title, Href: href}, nil
}

// takeLivePage mirrors _take_live_page: the load watch. The door
// WAITS for the page: a page counts as loaded only when it is
// 'complete' AT an http(s) URL AND has stayed there for PageSettleS
// (every href change or readyState drop restarts the settle); only
// then is the DOM read — and the DOM must pass PageIsReal. A page
// that reads as a CHALLENGE keeps the whole remaining budget; a page
// that reads as a Chrome error page or a crash is a named failure
// immediately. Returns an error only for a GONE tab (CDPClosedError)
// or a raw socket error (the caller's reconnect path).
func takeLivePage(ws cdpConn, pageTimeoutS float64) (livePage, error) {
	deadline := time.Now().Add(time.Duration(pageTimeoutS * float64(time.Second)))
	probe := ""
	var completeAt *time.Time
	settledHref := ""
	lastReason := ""
	for time.Now().Before(deadline) {
		v, err := ws.Evaluate(
			"document.readyState + \"|\" + location.href", 5.0)
		if err != nil {
			var te *CDPTimeoutError
			var ce *CDPClosedError
			switch {
			case errors.As(err, &te):
				// a busy page — the last observation stands
			case errors.As(err, &ce):
				return livePage{}, err // the tab is GONE — fail fast
			case asCDP(err):
				probe = "" // mid-navigation context swap — loading
			default:
				return livePage{}, err // a reset — the reconnect path
			}
		} else {
			probe = v
		}
		state, href := splitProbe(probe)
		if state == "complete" && isSiteURL(href) {
			if completeAt == nil || href != settledHref {
				now := time.Now()
				completeAt = &now
				settledHref = href
			}
			if time.Since(*completeAt) >=
				time.Duration(PageSettleS*float64(time.Second)) {
				// the page has been STILL — read it and judge it
				page, rerr := readAndVerdict(ws, settledHref)
				if rerr != nil {
					return livePage{}, rerr
				}
				if page.OK || !strings.Contains(page.Error, "challenge") {
					return page, nil
				}
				// a challenge: the owner's Chrome may still pass it —
				// restart the settle and keep the remaining budget
				lastReason = page.Error
				completeAt = nil
				settledHref = ""
				probe = ""
			}
		} else {
			completeAt = nil
			settledHref = ""
		}
		time.Sleep(time.Duration(pollCadenceS * float64(time.Second)))
	}
	// the budget ran out — tell what the tab was LAST showing
	if lastReason != "" && strings.Contains(lastReason, "challenge") {
		return livePage{OK: false, Error: strings.ReplaceAll(lastReason,
			"stayed on a bot challenge",
			"never got past the bot challenge")}, nil
	}
	state, href := splitProbe(probe)
	if state != "complete" {
		return livePage{OK: false, Href: href, Error: fmt.Sprintf(
			"the page never finished loading in %.0fs", pageTimeoutS)}, nil
	}
	finalHref := href
	if finalHref == "" {
		finalHref = settledHref
	}
	page, rerr := readAndVerdict(ws, finalHref)
	if rerr != nil {
		return livePage{}, rerr
	}
	return page, nil
}

// socketFailureSentence mirrors _socket_failure_sentence: a closed
// socket names the TAB loss; anything else names the socket's own
// error.
func socketFailureSentence(err error) string {
	text := err.Error()
	var ce *CDPClosedError
	if errors.As(err, &ce) || strings.Contains(strings.ToLower(text), "closed") {
		return fmt.Sprintf("the DevTools socket closed and the tab was "+
			"gone (%s)", text)
	}
	return fmt.Sprintf("the DevTools socket failed: %s", text)
}

// newTabSocketFn is the tab-socket seam (the reference's tests patch
// ct.CDPSocket for exactly this site).
var newTabSocketFn = func(host string, port int, path string) (cdpConn, error) {
	return NewCDPSocket(host, port, path)
}

// PageResult is the fetch driver's per-URL result: {url, ok, html,
// title, error}.
type PageResult struct {
	URL   string
	OK    bool
	HTML  string
	Title string
	Error string
}

// fetchOpenedTab mirrors _fetch_opened_tab: take one page through a
// tab that is ALREADY open — connect its DevTools socket, WAIT for
// the page (the load watch), read the live DOM, close the tab. Same
// result shape, never raises.
//
// v0.54 — the door KEEPS ITS GRIP: a DevTools socket that drops
// mid-watch is a BLIP, not a verdict: the socket is RE-CONNECTED (up
// to socketReconnects times) and the watch RESUMES with the remaining
// budget. A tab that is really GONE refuses its own WebSocket on the
// re-connect. The anti-bot patch is (re-)applied on every connection.
func fetchOpenedTab(session TabSession, rawURL string, tab TabInfo,
	pageTimeoutS float64, log LogFn) (result PageResult) {
	if log == nil {
		log = noopLog
	}
	targetID := tab.ID
	wsURL := tab.WsURL
	defer func() {
		if r := recover(); r != nil {
			// the reference's outermost except: 'unexpected: {e}'
			result = PageResult{URL: rawURL, OK: false,
				Error: fmt.Sprintf("unexpected: %v", r)}
		}
		if targetID != "" {
			session.CloseTab(targetID)
		}
	}()
	if wsURL == "" {
		return PageResult{URL: rawURL, OK: false,
			Error: "the tab reported no DevTools socket"}
	}
	_, host, _ := pyURLParts(wsURL)
	if host == "" {
		host = "127.0.0.1"
	}
	port := session.Port()
	if p := portOfNetloc(wsURL); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n != 0 {
			port = n
		}
	}
	path := pathOfURL(wsURL)
	if path == "" {
		path = "/"
	}
	deadline := time.Now().Add(time.Duration(pageTimeoutS * float64(time.Second)))
	reconnects := socketReconnects
	var page *livePage
	reason := ""
	for {
		remaining := time.Until(deadline)
		if remaining <= 0 {
			reason = fmt.Sprintf("the page never finished loading in %.0fs",
				pageTimeoutS)
			break
		}
		ws, err := newTabSocketFn(host, port, path)
		if err != nil {
			// the tab's own WebSocket refused — the TARGET is gone
			reason = fmt.Sprintf("the tab closed itself (%s)", err.Error())
			break
		}
		var perr error
		var p livePage
		if aerr := applyAntiBot(ws); aerr != nil {
			perr = aerr
		} else {
			p, perr = takeLivePage(ws, remaining.Seconds())
			if perr == nil {
				page = &p
			}
		}
		if perr == nil {
			ws.Close()
			break
		}
		ws.Close()
		if reconnects > 0 {
			reconnects--
			log(fmt.Sprintf("🤖 %s: the DevTools socket dropped mid-watch "+
				"(%s) — re-connecting to the still-open tab (the watch "+
				"resumes; the page budget keeps running)", rawURL,
				perr.Error()), "info")
			time.Sleep(time.Duration(reconnectPauseS * float64(time.Second)))
			continue
		}
		reason = socketFailureSentence(perr)
		break
	}
	if page != nil {
		if !page.OK {
			errStr := page.Error
			if errStr == "" {
				errStr = "the page never loaded"
			}
			return PageResult{URL: rawURL, OK: false, Title: page.Title,
				Error: errStr}
		}
		return PageResult{URL: rawURL, OK: true, HTML: page.HTML,
			Title: page.Title}
	}
	if reason == "" {
		reason = "the page never loaded"
	}
	return PageResult{URL: rawURL, OK: false, Error: reason}
}

// gatherThroughSession mirrors _gather_through_session: the wave
// engine — one tab per link opened in waves so tabs load together,
// each page WAITED FOR (the load watch) and taken from the live DOM.
// Returns one result per url (never raises; a dead session is a
// per-URL honest failure).
func gatherThroughSession(session TabSession, work []string,
	pageTimeoutS float64, wave int, shouldContinue func() bool,
	log LogFn) []PageResult {
	var results []PageResult
	pending := append([]string{}, work...)
	for len(pending) > 0 {
		if shouldContinue != nil && !shouldContinue() {
			log("⏹️ Chrome tab retry stopped — the remaining links "+
				"keep waiting in the retry queue", "warning")
			for _, u := range pending {
				results = append(results, PageResult{URL: u, OK: false,
					Error: "stopped by the user"})
			}
			break
		}
		batch := pending
		if len(batch) > wave {
			batch = batch[:wave]
		}
		pending = pending[len(batch):]
		log(fmt.Sprintf("🤖 Chrome wave: opening %d tab(s) in your real "+
			"Chrome…", len(batch)), "info")
		// open every tab of the wave first (parallel loading), then
		// take each page in turn
		type openedTab struct {
			url string
			tab TabInfo
		}
		var opened []openedTab
		for _, u := range batch {
			// the reference's except around session.new_tab: a raising
			// fake is an honest per-URL failure, never a broken batch
			tab, err := func() (t TabInfo, e error) {
				defer func() {
					if r := recover(); r != nil {
						e = newCDPError("%v", r)
					}
				}()
				return session.NewTab(u)
			}()
			if err != nil {
				results = append(results, PageResult{URL: u, OK: false,
					Error: err.Error()})
				continue
			}
			opened = append(opened, openedTab{url: u, tab: tab})
			log(fmt.Sprintf("🤖 Tab opened: %s", u), "info")
		}
		for _, ot := range opened {
			if !session.Alive() {
				results = append(results, PageResult{URL: ot.url, OK: false,
					Error: "the Chrome session died mid-run (its DevTools " +
						"endpoint stopped answering)"})
				continue
			}
			// fetch this page through its already-open tab
			page := fetchOpenedTab(session, ot.url, ot.tab, pageTimeoutS, log)
			results = append(results, page)
			if page.OK {
				log(fmt.Sprintf("✅ %s: took the page from your Chrome "+
					"(\"%s\", %d KB of live DOM)", ot.url,
					helpers.CutRunes(page.Title, 60),
					len([]rune(page.HTML))/1024), "success")
			} else {
				log(fmt.Sprintf("⚠️ %s: %s", ot.url, page.Error), "warning")
			}
		}
	}
	return results
}
