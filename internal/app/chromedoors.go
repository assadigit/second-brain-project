package app

// chromedoors.go — T-79d: the Chrome doors' engines (the More menu's
// fourth family), pinned against the reference's own flows
// (processing_control.py:615/669/1000/1601 + the shared
// _start_chrome_tab_retry worker half at 1351 + _on_delivery_failed
// at 1521, all at 98f4fc7). The engines live in internal/chrometabs
// (T-44) + internal/handdelivery; these doors are the anytime
// triggers over them, the Settings-method shape of T-79a/b/c.
//
// THE COLLABORATOR LAW (hero.py's optional-collaborator guards): the
// fifth door is OPTIONAL — Chrome absent, its DevTools port closed, a
// profile missing, a scan hiccup: every door refuses honestly with
// the reference's own remedy lines, never a crash (every body rides
// the reference's whole-try/except shape), never a silent no-op (the
// box or the desk line always speaks).
//
// The recorded divergences (the T-79c style — evidence doc
// docs/poc-artifacts/t-79d-chrome-doors.md):
//   - The reference's dry-run shadow-cache branch (WebsiteStateDB over
//     _dryrun.shadow_cache_path) is dead in its own GUI (the switch is
//     CLI-only; the GUI never flips it) — the port's doors read the
//     live cache, the same behavior in every reachable GUI state.
//   - The reference's _chrome_delivery_pending counter feeds the
//     batch-scorecard stash (v0.54's held scorecard — the batch
//     finish surface); the port's manual doors re-run the delivered
//     pages as their own batch whose own finish tells the whole
//     story, so no stash rides here.
//   - The picker dialogs are the frontend's (the two-door shape of
//     the recategorize flow); the picked rows ride back as the
//     dialog's own data (the reference's picked-from-candidates law).
//   - The "Open the folder" reveal rides the OpenHandDeliveryFolder
//     door (the OS file manager over the existing seam).

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/chrometabs"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// ---------------------------------------------------------------------------
// The payloads (THE FIELD-TAG LAW — the tags mirror the keys the
// frontend reads; pinned by TestChromeDoorPayloadFieldTags).
// ---------------------------------------------------------------------------

// ChromeLinkRow is the fifth door's link shape: {'url', 'error'}
// (the candidates' item + the delivery worker's row + the failure
// modal's failed-link row).
type ChromeLinkRow struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

// HandWallRow is the fourth door's picker row: {'url', 'wall',
// 'attempts'} (walled_retry_rows' own output shape).
type HandWallRow struct {
	URL      string `json:"url"`
	Wall     string `json:"wall"`
	Attempts int    `json:"attempts"`
}

// HandDeliverScan is 🖐 Hand-deliver walled links' scan verdict: the
// guards'/empty box (Box set, nothing else), or the picker's own data
// (the walled rows + the hand-delivered folder + the already-queued
// count the empty box speaks).
type HandDeliverScan struct {
	Box    *HeroBox      `json:"box"`
	Links  []HandWallRow `json:"links"`
	Folder string        `json:"folder"`
	Queued int           `json:"queued"`
}

// HandRowsScan is 🖐 Scrape hand rows via Chrome's scan verdict: the
// guards'/empty box, the >10 confirm question, and the rows in the
// delivery worker's shape (the error defaults already applied).
type HandRowsScan struct {
	Box     *HeroBox        `json:"box"`
	Confirm *HeroBox        `json:"confirm"`
	Rows    []ChromeLinkRow `json:"rows"`
}

// ChromeTabRetryScan is 🤖 Chrome tab-retry failed links' scan
// verdict: the guards'/empty box, or the offer's rows ({'url',
// 'error'} with each link's last error).
type ChromeTabRetryScan struct {
	Box   *HeroBox        `json:"box"`
	Links []ChromeLinkRow `json:"links"`
}

// ChromeDeliveryReport is the fifth door's delivery report:
// {'delivered', 'failed', 'urls', 'failed_links', 'folder'} (the
// worker's kept report — the failure modal + the delivered re-run
// read it).
type ChromeDeliveryReport struct {
	Delivered   int             `json:"delivered"`
	Failed      int             `json:"failed"`
	URLs        []string        `json:"urls"`
	FailedLinks []ChromeLinkRow `json:"failed_links"`
	Folder      string          `json:"folder"`
}

// AttachChromeResult is 🪄 Attach to my Chrome's verdict: a remedy
// box, or the one explicit OK question (Windows — his Chrome is
// about to be closed and reopened), or the restart's own box.
type AttachChromeResult struct {
	Box     *HeroBox `json:"box"`
	Confirm *HeroBox `json:"confirm"`
}

// ---------------------------------------------------------------------------
// The shared guards + helpers.
// ---------------------------------------------------------------------------

// chromeWebsitesGuard is the two Chrome-doors' shared guard ladder
// (processing_control.py:622-635 / 1010-1023 / 1608-1621): the
// Websites pipeline must be on and the Websites vault must be a
// directory. Non-nil = the remedy box.
func (s *Settings) chromeWebsitesGuard() *HeroBox {
	if !cfgNestedBool(s.cfg, "pipelines", "websites") {
		return &HeroBox{Title: "Websites pipeline is off",
			Body: "Turn the Websites pipeline on first (Settings → 📁 Vault)."}
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		return &HeroBox{Title: "No Websites vault",
			Body: "Set the Websites vault first (Settings → 📁 Vault)."}
	}
	return nil
}

// appRowsAdapter adapts websitepipeline.WebsiteStateDB onto
// handdelivery.RetryRowsState (the reference duck-types the
// pipeline's state DB into the picker; the concrete adapter lives at
// each call site — the decomtable law; the CLI hand leg carries its
// own twin).
type appRowsAdapter struct {
	db *websitepipeline.WebsiteStateDB
}

// AllRetryRows mirrors all_retry_rows(): every waiting retry-queue
// row, oldest first, in the picker's row shape.
func (a *appRowsAdapter) AllRetryRows() ([]handdelivery.RetrySourceRow, error) {
	rows, err := a.db.AllRetryRows()
	if err != nil {
		return nil, err
	}
	out := make([]handdelivery.RetrySourceRow, 0, len(rows))
	for _, r := range rows {
		le := ""
		if r.LastError != nil {
			le = *r.LastError
		}
		out = append(out, handdelivery.RetrySourceRow{
			URL: r.URL, LastError: le, Attempts: int(r.Attempts),
		})
	}
	return out, nil
}

// chromeHandGesture is the hand rows' error default (the reference's
// setdefault law, processing_control.py:658-660): the wall the row
// names, else the gesture itself.
const chromeHandGesture = "the 🖐 hand gesture in the master table"

// ---------------------------------------------------------------------------
// 🖐 Hand-deliver walled links (processing_control.py:1000).
// ---------------------------------------------------------------------------

// HandDeliverScan is hand_deliver_walled_links_now's scan half: the
// guards, the walled-rows scan (an honest read-failure line + the
// empty picker, never a crash), the already-queued count, and the
// empty box in the reference's own words (the queued variant names
// the folder).
func (s *Settings) HandDeliverScan() *HandDeliverScan {
	out := &HandDeliverScan{}
	if box := s.chromeWebsitesGuard(); box != nil {
		out.Box = box
		return out
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	folder := handdelivery.HandDeliveryDir(vault)
	out.Folder = folder

	var candidates []handdelivery.WalledRetryRow
	if state, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(s.cfg)); err == nil {
		candidates = handdelivery.WalledRetryRows(&appRowsAdapter{state})
		state.Close()
	} else {
		// the reference's except: the scan is skipped with a warning,
		// the empty picker follows (the box speaks the no-walled
		// story; the warning line tells the why).
		s.SpeakOnBus(fmt.Sprintf("⚠️ Walled-links scan skipped: %v", err), "warning")
	}
	// links already queued are kept out of the "new" count (the
	// reference's law — a re-open is harmless, the queue merge is a
	// no-op).
	if q := handdelivery.ReadQueue(vault); q != nil && q.Links != nil {
		out.Queued = q.Links.Len()
	}
	if len(candidates) == 0 {
		extra := "When one appears, its _review note and master-table row will say so (the fourth door hint)."
		if out.Queued > 0 {
			extra = fmt.Sprintf("%d link(s) are already queued for hand delivery — save their pages into:\n%s",
				out.Queued, folder)
		}
		out.Box = &HeroBox{Title: "No walled links waiting",
			Body: "No retry-queued link currently ends in a wall the machine doors could not open (403 / bot defense / TLS fingerprint).\n\n" + extra}
		return out
	}
	for _, c := range candidates {
		out.Links = append(out.Links, HandWallRow{
			URL: c.URL, Wall: c.Wall, Attempts: c.Attempts,
		})
	}
	return out
}

// HandDeliverQueue is the picker's answer (the reference's
// post-dialog half): the picked rows (echoed back from the dialog's
// own data — the same truth it rendered), the enqueue (queue.json +
// README + the open-in-Chrome leg), and the Queued for hand-delivery
// box. Title "" = the cancel line spoke on the desk (an empty picked
// list is the dialog's cancel — the reference's if-not-picked log).
// The guards ran at the scan; the enqueue itself is tolerant (the
// reference's try/except ⚠ line), so a vault that moved underneath
// the dialog answers the write's own honest warning.
func (s *Settings) HandDeliverQueue(links []HandWallRow) (title, body string) {
	if len(links) == 0 {
		s.SpeakOnBus("⏭️ Hand-delivery cancelled — nothing was queued. More ▸ "+
			"🖐 Hand-deliver walled links anytime (the master table's 🖐 hand "+
			"Status queues a single link too).", "info")
		return "", ""
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	folder := handdelivery.HandDeliveryDir(vault)
	urls := make([]string, 0, len(links))
	walls := map[string]string{}
	for _, l := range links {
		u := strings.TrimSpace(l.URL)
		if u == "" {
			continue
		}
		urls = append(urls, u)
		walls[u] = l.Wall
	}
	added, opened := 0, 0
	if s.Deps.HandEnqueue != nil {
		added, opened = s.Deps.HandEnqueue(vault, urls, walls,
			func(msg, level string) { s.SpeakOnBus(msg, level) }, true)
	} else {
		s.SpeakOnBus("⚠️ Hand-delivery queue write skipped: "+notWired("hand enqueue"), "warning")
	}
	openedPart := " — open them from the README (no browser could be launched here)."
	if opened > 0 {
		openedPart = fmt.Sprintf(", %d opened in your Chrome.", opened)
	}
	return "Queued for hand-delivery",
		fmt.Sprintf("%d link(s) queued%s\n\nFor each link: in Chrome, Ctrl+S → format 'Webpage, HTML Only' → filename = the suggested name (the README lists them) → folder:\n%s\n\nThe next SYNC batch consumes every delivered page as a REAL fetch — the note is written, the retry clears.",
			added, openedPart, folder)
}

// OpenHandDeliveryFolder is the picker's "Open the folder" reveal
// (the reference's _open_folder): the hand-delivered folder, created
// when missing, revealed in the OS file manager — best-effort.
func (s *Settings) OpenHandDeliveryFolder() bool {
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	if vault == "" {
		return false
	}
	folder := handdelivery.HandDeliveryDir(vault)
	_ = os.MkdirAll(folder, 0o755)
	if s.Deps.OpenFileLocation == nil {
		return false
	}
	s.Deps.OpenFileLocation(folder)
	return true
}

// ---------------------------------------------------------------------------
// 🖐 Scrape hand rows via Chrome (processing_control.py:615).
// ---------------------------------------------------------------------------

// HandRowsScan is hand_rows_deliver_now's everything-but-the-delivery
// half: the guards, the busy guard (a Chrome delivery already running
// — the box in the reference's own words; the flag is the shell's
// live seat, exactly like the reference's _chrome_delivery_busy),
// the master-table hand-rows scan (the pure file read), the empty box
// (a delivered row is consumed — the reference's own words), and the
// >10 confirm question. Rows carries the delivery shape (the error
// default already applied — the reference's setdefault law).
func (s *Settings) HandRowsScan(chromeBusy bool) *HandRowsScan {
	out := &HandRowsScan{}
	if box := s.chromeWebsitesGuard(); box != nil {
		out.Box = box
		return out
	}
	if chromeBusy {
		out.Box = &HeroBox{Title: "A Chrome delivery is running",
			Body: "A delivery is already open in your Chrome — the hand rows queued behind it open when the current tabs finish."}
		return out
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	rows := websitepipeline.ScanMasterHandRows(vault, nil)
	if len(rows) == 0 {
		out.Box = &HeroBox{Title: "No hand rows waiting",
			Body: "The master table (_review/DECOMMISSIONED.md) has no 🖐 hand row waiting for a Chrome scrape (a delivered row is consumed — set 🖐 anywhere in a row: the # cell or the Status cell)."}
		return out
	}
	for _, r := range rows {
		errText := r.Wall
		if errText == "" {
			errText = chromeHandGesture
		}
		out.Rows = append(out.Rows, ChromeLinkRow{URL: r.URL, Error: errText})
	}
	// v31.1 safety gate: confirm before large batches (>10 items).
	if ask, title, body := ConfirmLargeBatch(len(out.Rows), "the table's 🖐 hand rows"); ask {
		out.Confirm = &HeroBox{Title: title, Body: body, Question: true}
	}
	return out
}

// ---------------------------------------------------------------------------
// 🤖 Chrome tab-retry failed links (processing_control.py:1601).
// ---------------------------------------------------------------------------

// ChromeTabRetryScan is chrome_tab_retry_now's scan half: the
// guards, the batch-busy guard (_batch_running — the flag is the
// Commands surface's live lock, passed by the shell), the
// retry-queue read (the honest read-failure line), the routing
// filters (http(s) only, loopback never asks, the ladder's own
// dismissed verdicts never ask), and the session's offered filter
// (the end-of-run offer and this door share one set — a link declined
// once this session is not re-offered).
func (s *Settings) ChromeTabRetryScan(batchBusy bool) *ChromeTabRetryScan {
	out := &ChromeTabRetryScan{}
	if box := s.chromeWebsitesGuard(); box != nil {
		out.Box = box
		return out
	}
	if batchBusy {
		out.Box = &HeroBox{Title: "Batch already running",
			Body: "A batch is already running — finish or stop it first."}
		return out
	}
	var candidates []ChromeLinkRow
	if state, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(s.cfg)); err == nil {
		rows, rerr := state.AllRetryRows()
		if rerr != nil {
			s.SpeakOnBus(fmt.Sprintf("⚠️ Retry-queue scan skipped: %v", rerr), "warning")
		} else {
			for _, row := range rows {
				u := strings.TrimSpace(row.URL)
				lu := strings.ToLower(u)
				if !strings.HasPrefix(lu, "http://") && !strings.HasPrefix(lu, "https://") {
					continue
				}
				if chrometabs.IsLoopbackURL(u) {
					continue
				}
				if dismissed, _ := state.IsDismissed(u); dismissed {
					continue // the ladder's own verdict answered it
				}
				errTxt := ""
				if row.LastError != nil {
					errTxt = *row.LastError
				}
				candidates = append(candidates, ChromeLinkRow{URL: u, Error: errTxt})
			}
		}
		state.Close()
	} else {
		s.SpeakOnBus(fmt.Sprintf("⚠️ Retry-queue scan skipped: %v", err), "warning")
	}
	s.mu.Lock()
	for _, c := range candidates {
		if s.chromeTabOffered[c.URL] {
			continue
		}
		out.Links = append(out.Links, c)
	}
	s.mu.Unlock()
	if len(out.Links) == 0 {
		out.Box = &HeroBox{Title: "No failed links waiting",
			Body: "The websites retry queue has no failed link waiting for the fifth door (loopback and auto-retired links never ask). When a fetch fails, the end-of-run modal offers the retry automatically."}
		return out
	}
	return out
}

// MarkChromeTabOffered is the offered set's update (the reference's
// offered.update(p['url'] for p in picked) — a link the owner picked
// is never re-offered this session).
func (s *Settings) MarkChromeTabOffered(links []ChromeLinkRow) {
	s.mu.Lock()
	if s.chromeTabOffered == nil {
		s.chromeTabOffered = map[string]bool{}
	}
	for _, l := range links {
		if strings.TrimSpace(l.URL) != "" {
			s.chromeTabOffered[strings.TrimSpace(l.URL)] = true
		}
	}
	s.mu.Unlock()
}

// ---------------------------------------------------------------------------
// The delivery worker's half (_start_chrome_tab_retry's run() +
// _on_delivery_failed's grammar — the shell owns the busy/backlog
// seat, the re-run batch, and the modal events).
// ---------------------------------------------------------------------------

// ChromeDeliverLinks is the ChromeTabRetryWorker's run() half: the
// fifth door's delivery over the ChromeDeliver seam (production:
// chrometabs.DeliverPagesViaChrome — THE CDP/DevTools-touching seam,
// faked at construction per THE HERMETIC-SEAM law). The whole-body
// guard is the reference's own try/except (the door never breaks
// anything: a crash answers the honest failed report with the ⚠ line
// on the desk).
func (s *Settings) ChromeDeliverLinks(links []ChromeLinkRow) (rep ChromeDeliveryReport) {
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	rep.Folder = handdelivery.HandDeliveryDir(vault)
	if vault == "" || len(links) == 0 {
		return rep
	}
	if s.Deps.ChromeDeliver == nil {
		s.SpeakOnBus("⚠️ Chrome tab retry failed: "+notWired("chrome deliver"), "warning")
		rep.Failed = 1
		rep.FailedLinks = []ChromeLinkRow{{URL: "",
			Error: "Chrome tab retry failed: " + notWired("chrome deliver")}}
		return rep
	}
	func() {
		defer func() {
			if r := recover(); r != nil {
				// the reference's except: the worker logs the crash and
				// emits the honest failed report (never a crash).
				s.SpeakOnBus(fmt.Sprintf("⚠️ Chrome tab retry failed: %v", r), "warning")
				rep = ChromeDeliveryReport{
					Failed: 1,
					FailedLinks: []ChromeLinkRow{{URL: "",
						Error: fmt.Sprintf("Chrome tab retry failed: %v", r)}},
				}
			}
		}()
		rep = s.Deps.ChromeDeliver(vault, links,
			func(msg, level string) { s.SpeakOnBus(msg, level) }, s.cfg)
	}()
	return rep
}

// ChromeDeliveryFailedBox is _on_delivery_failed's modal grammar
// (processing_control.py:1521-1566): the owner's words — the count,
// the they-keep-waiting story, the first five errors by name. Nil =
// no failed links (nothing to say).
func ChromeDeliveryFailedBox(failed []ChromeLinkRow) *HeroBox {
	if len(failed) == 0 {
		return nil
	}
	var names []string
	shown := failed
	if len(shown) > 5 {
		shown = shown[:5]
	}
	for _, f := range shown {
		url := strings.TrimSpace(f.URL)
		err := strings.TrimSpace(f.Error)
		line := "• " + url
		if url == "" {
			line = "• (a link)"
		}
		if err != "" {
			line += " — " + err
		}
		names = append(names, line)
	}
	if len(failed) > 5 {
		names = append(names, fmt.Sprintf("… and %d more", len(failed)-5))
	}
	return &HeroBox{Title: "Chrome tab retry — pages not taken",
		Body: fmt.Sprintf("%d link(s) didn't generate content from your Chrome (the tabs crashed, showed an error page, or never finished loading).\n\nThey keep waiting in the retry queue — the next run offers them again (More ▸ 🤖 Chrome tab-retry anytime).\n\n%s",
			len(failed), strings.Join(names, "\n"))}
}

// ---------------------------------------------------------------------------
// 🪄 Attach to my Chrome (processing_control.py:669).
// ---------------------------------------------------------------------------

// attachPortCoerce is the reference's port read: int(cfg.get(...) or
// 9222) with the TypeError/ValueError fallback — junk of every shape
// reads the default.
func attachPortCoerce(raw any) int {
	switch t := raw.(type) {
	case nil:
		return 9222
	case bool:
		if !t {
			return 9222
		}
		return 1 // int(True)
	case int:
		if t == 0 {
			return 9222
		}
		return t
	case int64:
		if t == 0 {
			return 9222
		}
		return int(t)
	case float64:
		if t == 0 {
			return 9222
		}
		return int(t)
	case string:
		if t == "" {
			return 9222
		}
		if n, err := strconv.Atoi(t); err == nil {
			if n == 0 {
				return 9222
			}
			return n
		}
	}
	return 9222
}

// AttachMyChrome is attach_my_chrome_now's whole ladder: find the
// real Chrome (the remedy box when absent), find the real profile
// (the remedy box when empty), write chrome-attach.bat (the write
// failure box), then the platform split — on Windows the ONE explicit
// OK question before the restart (the confirm gates the run; the
// confirmed pass runs the .bat and answers the restart's box, the
// launch failure answering the run-it-by-hand box), on POSIX the
// honest the-script-is-ready box (start Chrome yourself with the
// port; the door attaches automatically). The whole body rides the
// reference's own outer try/except — never a crash.
func (s *Settings) AttachMyChrome(confirmed bool) (res *AttachChromeResult) {
	res = &AttachChromeResult{}
	defer func() {
		if r := recover(); r != nil {
			s.SpeakOnBus(fmt.Sprintf("⚠️ The Chrome attach setup failed (%v)", r), "warning")
			res = &AttachChromeResult{}
		}
	}()
	chrome := ""
	if s.Deps.ChromeFind != nil {
		chrome = s.Deps.ChromeFind()
	}
	if chrome == "" {
		res.Box = &HeroBox{Title: "No Chrome found",
			Body: "GitCurator could not find Google Chrome on this machine — the fifth door needs the real browser."}
		return res
	}
	dataDir := ""
	if s.Deps.ChromeUserDataDir != nil {
		dataDir = s.Deps.ChromeUserDataDir()
	}
	if dataDir == "" {
		res.Box = &HeroBox{Title: "No Chrome profile found",
			Body: "GitCurator could not find your Chrome profile (the usual location is empty) — nothing to attach to yet."}
		return res
	}
	port := 9222
	if raw, ok := s.cfg.Get("web_browser_attach_port"); ok {
		port = attachPortCoerce(raw)
	} else {
		port = attachPortCoerce(nil)
	}
	batPath := filepath.Join(constants.AppDir(), "chrome-attach.bat")
	written := false
	if s.Deps.ChromeAttachBat != nil {
		written = s.Deps.ChromeAttachBat(batPath, chrome, port, dataDir, constants.AppDir())
	}
	if !written {
		res.Box = &HeroBox{Title: "The attach script could not be written",
			Body: fmt.Sprintf("GitCurator could not write %s — check the folder's permissions and try again.", batPath)}
		return res
	}
	if runtime.GOOS != "windows" {
		s.SpeakOnBus(fmt.Sprintf("🪄 The attach script is ready: %s — a Windows .bat (on this machine, start your Chrome yourself with --remote-debugging-port=%d and the door attaches automatically)", batPath, port), "info")
		res.Box = &HeroBox{Title: "The attach script is ready",
			Body: fmt.Sprintf("%s\n\nA Windows .bat (this machine is not Windows) — start your Chrome with --remote-debugging-port=%d and every delivery attaches to it automatically.", batPath, port)}
		return res
	}
	if !confirmed {
		res.Confirm = &HeroBox{Title: "Attach the fifth door to MY Chrome",
			Body:     "GitCurator will now close and restart your Chrome with its DevTools port open — your profile, your cookies, your logins; your tabs come back (--restore-last-session).\n\nThe fifth door then opens its tabs in YOUR window (it never closes your browser — only its own tabs).\n\nTo undo: relaunch Chrome normally.",
			Question: true}
		return res
	}
	s.SpeakOnBus(fmt.Sprintf("🪄 Restarting your Chrome with the door attached (DevTools on 127.0.0.1:%d) — your tabs come back; the next delivery opens ITS tabs in YOUR window", port), "info")
	errText := ""
	if s.Deps.ChromeAttachRun != nil {
		errText = s.Deps.ChromeAttachRun(batPath)
	}
	if errText != "" {
		s.SpeakOnBus(fmt.Sprintf("⚠️ The attach script could not be run (%s) — run it by hand: %s", errText, batPath), "warning")
		res.Box = &HeroBox{Title: "Run it by hand",
			Body: fmt.Sprintf("The script is ready but could not be launched (%s).\n\nRun it by hand:\n%s", errText, batPath)}
		return res
	}
	res.Box = &HeroBox{Title: "Your Chrome is being restarted with the door attached",
		Body: "Chrome is closing and reopening with its DevTools port open (--restore-last-session brings your tabs back).\n\nFrom now on every Chrome delivery (and More ▸ 🖐 Scrape hand rows via Chrome) opens its tabs in YOUR window — your profile, your cookies, your logins — and GitCurator never closes your browser (only its own tabs).\n\nTo undo: just relaunch Chrome normally."}
	return res
}
