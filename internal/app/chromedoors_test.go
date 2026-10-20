package app

// chromedoors_test.go — T-79d: the Chrome doors' hermetic battery,
// pinned against the reference's own flows (processing_control.py
// 615/669/1000/1601 + the worker half at 1351-1519 + _on_delivery_
// failed at 1521-1566, all at 98f4fc7). Every CDP/DevTools-touching
// + browser-launching seam faked at construction; every box + line
// asserted by level + verbatim text; the walled/retry scans ride the
// rig's own temp cache.db; the hand rows ride the real table file.

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// newChromeDoorRig is the Chrome doors' harness: the websites
// pipeline on + a real temp Websites vault + the rig's own temp
// APP_DIR (the cache.db + chrome-attach.bat land there).
func newChromeDoorRig(t *testing.T) (*Settings, *busRec, *fakeSettingsDeps) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	s, rec, f := newSettingsTest(t)
	f.chromeAttachBatOK = true
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	s.cfg.Set("website_vault_path", t.TempDir())
	return s, rec, f
}

// chromeDoorState opens the rig's cache.db (the temp APP_DIR's own)
// and returns the state DB over it.
func chromeDoorState(t *testing.T) *websitepipeline.WebsiteStateDB {
	t.Helper()
	state, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(nil))
	if err != nil {
		t.Fatal(err)
	}
	return state
}

// writeHandTable writes the master table with one row per (url,
// status, notes) triple — the same shape websitepipeline's own tests
// write; the notes column is the Wall the hand rows carry.
func writeHandTable(t *testing.T, vault string, rows [][3]string) {
	t.Helper()
	path := websitepipeline.DecommissionTablePath(vault)
	var lines []string
	for _, row := range rows {
		domain := "unknown"
		if i := strings.Index(row[0], "://"); i >= 0 {
			rest := row[0][i+3:]
			if j := strings.IndexAny(rest, "/?#"); j >= 0 {
				rest = rest[:j]
			}
			if rest != "" {
				domain = rest
			}
		}
		lines = append(lines, fmt.Sprintf("| - | 2026-10-09 | %s | %s | test | %s | %s |",
			row[0], domain, row[1], row[2]))
	}
	content := fmt.Sprintf(websitepipeline.GraveyardHeader, "2026-10-09 12:00") +
		strings.Join(lines, "\n") + "\n"
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// The shared guards (the reference's own remedy boxes, verbatim).
// ---------------------------------------------------------------------------

func TestChromeDoorsWebsitesGuards(t *testing.T) {
	s, rec, f := newChromeDoorRig(t)
	_ = f
	// the pipeline off — every scan door refuses with the same box.
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	for name, box := range map[string]*HeroBox{
		"hand-deliver": s.HandDeliverScan().Box,
		"hand-rows":    s.HandRowsScan(false).Box,
		"chrome-retry": s.ChromeTabRetryScan(false).Box,
	} {
		if box == nil || box.Title != "Websites pipeline is off" ||
			box.Body != "Turn the Websites pipeline on first (Settings → 📁 Vault)." {
			t.Fatalf("%s: the pipeline guard = %+v", name, box)
		}
	}
	// the vault guard — empty, then a missing dir.
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	s.cfg.Set("website_vault_path", "")
	for name, box := range map[string]*HeroBox{
		"hand-deliver": s.HandDeliverScan().Box,
		"hand-rows":    s.HandRowsScan(false).Box,
		"chrome-retry": s.ChromeTabRetryScan(false).Box,
	} {
		if box == nil || box.Title != "No Websites vault" ||
			box.Body != "Set the Websites vault first (Settings → 📁 Vault)." {
			t.Fatalf("%s: the empty-vault guard = %+v", name, box)
		}
	}
	s.cfg.Set("website_vault_path", filepath.Join(t.TempDir(), "missing"))
	for name, box := range map[string]*HeroBox{
		"hand-deliver": s.HandDeliverScan().Box,
		"hand-rows":    s.HandRowsScan(false).Box,
		"chrome-retry": s.ChromeTabRetryScan(false).Box,
	} {
		if box == nil || box.Title != "No Websites vault" {
			t.Fatalf("%s: the missing-dir guard = %+v", name, box)
		}
	}
	if len(rec.logs) != 0 {
		t.Fatalf("a guard box is silent on the desk (have %d lines)", len(rec.logs))
	}
}

// ---------------------------------------------------------------------------
// 🖐 Hand-deliver walled links (processing_control.py:1000).
// ---------------------------------------------------------------------------

func TestHandDeliverScanGrammar(t *testing.T) {
	s, rec, f := newChromeDoorRig(t)
	vault, _ := s.cfg.Get("website_vault_path")
	_ = f

	// the empty picker: the box + the hint half.
	out := s.HandDeliverScan()
	if out.Box == nil || out.Box.Title != "No walled links waiting" ||
		!strings.Contains(out.Box.Body, "No retry-queued link currently ends in a wall the machine doors could not open (403 / bot defense / TLS fingerprint).\n\nWhen one appears, its _review note and master-table row will say so (the fourth door hint).") {
		t.Fatalf("the empty box: %+v", out.Box)
	}
	if out.Folder != filepath.Join(vault.(string), "_review", "hand-delivered") {
		t.Fatalf("the folder: %q", out.Folder)
	}
	if out.Queued != 0 || out.Links != nil {
		t.Fatalf("the empty picker's data: %+v", out)
	}

	// the queued variant: a queue row already waits — the count names
	// the folder.
	if q := os.MkdirAll(filepath.Join(vault.(string), "_review", "hand-delivered"), 0o755); q != nil {
		t.Fatal(q)
	}
	if err := os.WriteFile(
		filepath.Join(vault.(string), "_review", "hand-delivered", "queue.json"),
		[]byte(`{"links":{"https://example.com/waiting":{"wall":"x","queued":"2026-10-09 12:00","suggested":"a.html"}}}`),
		0o644); err != nil {
		t.Fatal(err)
	}
	out = s.HandDeliverScan()
	if out.Box == nil || !strings.Contains(out.Box.Body,
		"1 link(s) are already queued for hand delivery — save their pages into:\n"+out.Folder) {
		t.Fatalf("the queued box: %+v", out.Box)
	}

	// the walled rows: the picker's data in the reference's row shape.
	state := chromeDoorState(t)
	if err := state.EnqueueRetry("https://example.com/walled", "HTTP 403 — bot defense"); err != nil {
		t.Fatal(err)
	}
	if err := state.EnqueueRetry("https://example.com/plain", "HTTP 500"); err != nil {
		t.Fatal(err)
	}
	state.Close()
	out = s.HandDeliverScan()
	if out.Box != nil {
		t.Fatalf("the walled picker answers data, not a box: %+v", out.Box)
	}
	if len(out.Links) != 1 {
		t.Fatalf("the walled rows: %+v", out.Links)
	}
	if out.Links[0].URL != "https://example.com/walled" ||
		out.Links[0].Wall != "HTTP 403 — bot defense" || out.Links[0].Attempts != 1 {
		t.Fatalf("the walled row: %+v", out.Links[0])
	}
	if out.Queued != 1 {
		t.Fatalf("the queued count rides along: %d", out.Queued)
	}

	// the read failure: a corrupt cache.db answers the warning line +
	// the empty picker (the reference's except — never a crash).
	if err := os.Remove(cacheDBPathOf(s.cfg)); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(cacheDBPathOf(s.cfg), []byte("not a database"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec.logs = nil
	out = s.HandDeliverScan()
	if l, ok := rec.findLog("⚠️ Walled-links scan skipped: "); !ok || l.Level != "warning" {
		t.Fatalf(" the scan-failure line: %+v", l)
	}
	if out.Box == nil || out.Box.Title != "No walled links waiting" {
		t.Fatalf("the corrupt-cache box: %+v", out.Box)
	}
}

func TestHandDeliverQueueGrammar(t *testing.T) {
	s, rec, f := newChromeDoorRig(t)

	// the cancel: the dialog's empty answer speaks the reference's
	// own line, nothing is queued.
	title, body := s.HandDeliverQueue(nil)
	if title != "" || body != "" {
		t.Fatalf("the cancel is log-only: %q / %q", title, body)
	}
	wantLogAt(t, rec, "⏭️ Hand-delivery cancelled — nothing was queued. More ▸ 🖐 Hand-deliver walled links anytime (the master table's 🖐 hand Status queues a single link too).", "info", 1)
	if f.handEnqueueCalls != 0 {
		t.Fatalf("the cancel never enqueues (%d calls)", f.handEnqueueCalls)
	}

	// the queue: the picked rows (the dialog's own data echoed back),
	// the enqueue with the walls + the open leg, and the box.
	f.handEnqueueFn = func(vault string, urls []string,
		walls map[string]string, log LogFn, openChrome bool) (int, int) {
		if len(urls) != 2 || !openChrome {
			t.Fatalf("the enqueue call: %v open=%v", urls, openChrome)
		}
		if walls["https://example.com/one"] != "HTTP 403" {
			t.Fatalf("the walls ride the rows: %v", walls)
		}
		log("🖐 Hand-delivery: 2 link(s) queued — save each page (Ctrl+S, 'Webpage, HTML Only') as its suggested name into the folder; the next batch consumes them", "info")
		return 2, 2
	}
	vault, _ := s.cfg.Get("website_vault_path")
	title, body = s.HandDeliverQueue([]HandWallRow{
		{URL: "https://example.com/one", Wall: "HTTP 403"},
		{URL: "https://example.com/two", Wall: ""},
	})
	if title != "Queued for hand-delivery" {
		t.Fatalf("the queued box title: %q", title)
	}
	wantBody := "2 link(s) queued, 2 opened in your Chrome.\n\n" +
		"For each link: in Chrome, Ctrl+S → format 'Webpage, HTML Only' → filename = the suggested name (the README lists them) → folder:\n" +
		filepath.Join(vault.(string), "_review", "hand-delivered") + "\n\n" +
		"The next SYNC batch consumes every delivered page as a REAL fetch — the note is written, the retry clears."
	if body != wantBody {
		t.Fatalf("the queued box body:\n%q\nwant:\n%q", body, wantBody)
	}
	if f.handEnqueueCalls != 1 {
		t.Fatalf("one enqueue call (have %d)", f.handEnqueueCalls)
	}

	// the no-browser variant: 0 opened — the README hint half.
	f.handEnqueueFn = func(vault string, urls []string,
		walls map[string]string, log LogFn, openChrome bool) (int, int) {
		return 3, 0
	}
	title, body = s.HandDeliverQueue([]HandWallRow{{URL: "https://example.com/x"}})
	if !strings.Contains(body, "3 link(s) queued — open them from the README (no browser could be launched here).") {
		t.Fatalf("the no-browser half: %q", body)
	}
	if title != "Queued for hand-delivery" {
		t.Fatalf("the title: %q", title)
	}
}

func TestHandDeliverOpenFolder(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)
	opened := ""
	s.Deps.OpenFileLocation = func(p string) { opened = p }
	if !s.OpenHandDeliveryFolder() {
		t.Fatal("the reveal door answers true")
	}
	vault, _ := s.cfg.Get("website_vault_path")
	if opened != filepath.Join(vault.(string), "_review", "hand-delivered") {
		t.Fatalf("the revealed folder: %q", opened)
	}
	if fi, err := os.Stat(opened); err != nil || !fi.IsDir() {
		t.Fatalf("the folder was created: %v", err)
	}
	// no vault — the quiet refusal.
	s.cfg.Set("website_vault_path", "")
	if s.OpenHandDeliveryFolder() {
		t.Fatal("no vault, no reveal")
	}
}

// ---------------------------------------------------------------------------
// 🖐 Scrape hand rows via Chrome (processing_control.py:615).
// ---------------------------------------------------------------------------

func TestHandRowsScanGrammar(t *testing.T) {
	s, _, f := newChromeDoorRig(t)
	vault, _ := s.cfg.Get("website_vault_path")
	_ = f

	// the empty table: the reference's own box.
	out := s.HandRowsScan(false)
	if out.Box == nil || out.Box.Title != "No hand rows waiting" ||
		out.Box.Body != "The master table (_review/DECOMMISSIONED.md) has no 🖐 hand row waiting for a Chrome scrape (a delivered row is consumed — set 🖐 anywhere in a row: the # cell or the Status cell)." {
		t.Fatalf("the empty box: %+v", out.Box)
	}

	// the hand rows: the wall rides as the error, the gesture answers
	// when no wall speaks (the setdefault law).
	writeHandTable(t, vault.(string), [][3]string{
		{"https://example.com/walled-hand", "🖐 hand", "HTTP 403 — bot defense"},
		{"https://example.com/gesture", "🖐", ""},
		{"https://example.com/reviewed", "✅ reviewed", ""},
	})
	out = s.HandRowsScan(false)
	if out.Box != nil || out.Confirm != nil {
		t.Fatalf("the hand rows answer data: %+v / %+v", out.Box, out.Confirm)
	}
	if len(out.Rows) != 2 {
		t.Fatalf("the hand rows (the reviewed row never asks): %+v", out.Rows)
	}
	if out.Rows[0].URL != "https://example.com/walled-hand" ||
		out.Rows[0].Error != "HTTP 403 — bot defense" {
		t.Fatalf("the wall row: %+v", out.Rows[0])
	}
	if out.Rows[1].URL != "https://example.com/gesture" ||
		out.Rows[1].Error != "the 🖐 hand gesture in the master table" {
		t.Fatalf("the gesture row: %+v", out.Rows[1])
	}

	// the >10 gate (v31.1): 11 rows ask, 10 fall straight through.
	rows := make([][3]string, 11)
	for i := range rows {
		rows[i] = [3]string{fmt.Sprintf("https://example.com/h-%02d", i), "🖐", ""}
	}
	writeHandTable(t, vault.(string), rows)
	out = s.HandRowsScan(false)
	if out.Confirm == nil || out.Confirm.Title != "Confirm Large Batch" ||
		!strings.Contains(out.Confirm.Body, "This will process 11 items (the table's 🖐 hand rows).") {
		t.Fatalf("the >10 gate: %+v", out.Confirm)
	}
	rows = rows[:10]
	writeHandTable(t, vault.(string), rows)
	out = s.HandRowsScan(false)
	if out.Confirm != nil {
		t.Fatalf("10 rows fall straight through: %+v", out.Confirm)
	}
	if len(out.Rows) != 10 {
		t.Fatalf("the rows: %d", len(out.Rows))
	}
}

// ---------------------------------------------------------------------------
// 🤖 Chrome tab-retry failed links (processing_control.py:1601).
// ---------------------------------------------------------------------------

func TestChromeTabRetryScanGrammar(t *testing.T) {
	s, rec, _ := newChromeDoorRig(t)

	// the empty queue: the box in the reference's own words.
	out := s.ChromeTabRetryScan(false)
	if out.Box == nil || out.Box.Title != "No failed links waiting" ||
		!strings.Contains(out.Box.Body, "The websites retry queue has no failed link waiting for the fifth door (loopback and auto-retired links never ask). When a fetch fails, the end-of-run modal offers the retry automatically.") {
		t.Fatalf("the empty box: %+v", out.Box)
	}

	// the routing filters: http(s) only, loopback never asks, the
	// ladder's own dismissed verdicts never ask.
	state := chromeDoorState(t)
	for _, row := range [][2]string{
		{"https://example.com/live-one", "HTTP 403 — bot defense"},
		{"http://example.com/live-two", "connection reset"},
		{"https://127.0.0.1:8080/local", "loopback"},
	} {
		if err := state.EnqueueRetry(row[0], row[1]); err != nil {
			t.Fatal(err)
		}
	}
	if err := state.Dismiss("https://example.com/live-two", "auto-verdict"); err != nil {
		t.Fatal(err)
	}
	state.Close()
	out = s.ChromeTabRetryScan(false)
	if out.Box != nil || len(out.Links) != 1 {
		t.Fatalf("the filters: %+v", out)
	}
	if out.Links[0].URL != "https://example.com/live-one" ||
		out.Links[0].Error != "HTTP 403 — bot defense" {
		t.Fatalf("the row: %+v", out.Links[0])
	}

	// the offered set: a picked link is never re-offered this session.
	s.MarkChromeTabOffered([]ChromeLinkRow{{URL: "https://example.com/live-one"}})
	out = s.ChromeTabRetryScan(false)
	if out.Box == nil || out.Box.Title != "No failed links waiting" {
		t.Fatalf("the offered filter: %+v", out.Box)
	}
	// an empty url never marks (junk tolerance).
	s.MarkChromeTabOffered([]ChromeLinkRow{{URL: "  "}})
	if l, ok := rec.findLog("⚠️ Retry-queue scan skipped: "); ok {
		t.Fatalf("no scan ran since: %q", l.Msg)
	}
}

func TestChromeDoorBusyLadders(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)
	// the busy guards sit BETWEEN the config guards and the scan (the
	// reference's own ladder order — a pipeline-off config answers the
	// pipeline box even while busy, the cross-binary law).
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	if out := s.HandRowsScan(true); out.Box == nil || out.Box.Title != "Websites pipeline is off" {
		t.Fatalf("the guards speak before the busy flag: %+v", out.Box)
	}
	if out := s.ChromeTabRetryScan(true); out.Box == nil || out.Box.Title != "Websites pipeline is off" {
		t.Fatalf("the guards speak before the busy flag: %+v", out.Box)
	}
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	// the busy boxes in the reference's own words.
	if out := s.HandRowsScan(true); out.Box == nil || out.Box.Title != "A Chrome delivery is running" ||
		out.Box.Body != "A delivery is already open in your Chrome — the hand rows queued behind it open when the current tabs finish." {
		t.Fatalf("the hand-rows busy box: %+v", out.Box)
	}
	if out := s.ChromeTabRetryScan(true); out.Box == nil || out.Box.Title != "Batch already running" ||
		out.Box.Body != "A batch is already running — finish or stop it first." {
		t.Fatalf("the chrome-retry busy box: %+v", out.Box)
	}
	// unbusy: both scan (the empty boxes answer — the scans ran).
	if out := s.HandRowsScan(false); out.Box == nil || out.Box.Title != "No hand rows waiting" {
		t.Fatalf("the unbusy hand-rows scan: %+v", out.Box)
	}
	if out := s.ChromeTabRetryScan(false); out.Box == nil || out.Box.Title != "No failed links waiting" {
		t.Fatalf("the unbusy chrome-retry scan: %+v", out.Box)
	}
}

func TestChromeTabRetryScanReadFailure(t *testing.T) {
	s, rec, _ := newChromeDoorRig(t)
	_ = s
	// a corrupt cache.db: the warning line + the empty picker.
	if err := os.WriteFile(cacheDBPathOf(s.cfg), []byte("junk"), 0o644); err != nil {
		t.Fatal(err)
	}
	rec.logs = nil
	out := s.ChromeTabRetryScan(false)
	wantLogAt(t, rec, "⚠️ Retry-queue scan skipped: ", "warning", 1)
	if out.Box == nil || out.Box.Title != "No failed links waiting" {
		t.Fatalf("the read-failure box: %+v", out.Box)
	}
}

func TestMarkChromeTabOfferedConcurrent(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)
	// THE SENTINEL-SEAT RACE LAW: the offered set is read+written by
	// concurrent passes — the mutex keeps it ONE unit.
	var wg sync.WaitGroup
	for i := 0; i < 16; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			for j := 0; j < 32; j++ {
				s.MarkChromeTabOffered([]ChromeLinkRow{
					{URL: fmt.Sprintf("https://example.com/%d-%d", n, j)}})
				_ = s.ChromeTabRetryScan(false)
			}
		}(i)
	}
	wg.Wait()
	s.mu.Lock()
	count := len(s.chromeTabOffered)
	s.mu.Unlock()
	if count != 16*32 {
		t.Fatalf("the offered set: %d marks", count)
	}
}

// ---------------------------------------------------------------------------
// The delivery worker's half (processing_control.py:1351-1519).
// ---------------------------------------------------------------------------

func TestChromeDeliverLinks(t *testing.T) {
	s, rec, f := newChromeDoorRig(t)

	// the early returns: no vault / no links — the empty report.
	s.cfg.Set("website_vault_path", "")
	if rep := s.ChromeDeliverLinks([]ChromeLinkRow{{URL: "https://example.com/x"}}); rep.Delivered != 0 || rep.Failed != 0 {
		t.Fatalf("the no-vault early return: %+v", rep)
	}
	s.cfg.Set("website_vault_path", t.TempDir())
	if rep := s.ChromeDeliverLinks(nil); rep.Delivered != 0 {
		t.Fatalf("the no-links early return: %+v", rep)
	}

	// the happy delivery: the seam's report passes through, the log
	// rides SpeakOnBus.
	f.chromeDeliverFn = func(vault string, links []ChromeLinkRow,
		log LogFn, config *storage.OrderedMap) ChromeDeliveryReport {
		if len(links) != 1 || links[0].Error != "the wall" {
			t.Fatalf("the seam call: %+v", links)
		}
		if _, ok := config.Get("website_vault_path"); !ok {
			t.Fatal("the live config rides the delivery")
		}
		log("🤖 Fifth door: 1 page(s) delivered from your real Chrome", "success")
		return ChromeDeliveryReport{Delivered: 1, URLs: []string{"https://example.com/x"}}
	}
	rep := s.ChromeDeliverLinks([]ChromeLinkRow{{URL: "https://example.com/x", Error: "the wall"}})
	if rep.Delivered != 1 || len(rep.URLs) != 1 || rep.Failed != 0 {
		t.Fatalf("the report: %+v", rep)
	}
	wantLogAt(t, rec, "🤖 Fifth door: 1 page(s) delivered from your real Chrome", "success", 1)

	// the crash guard: a panicking seam answers the honest failed
	// report + the ⚠ line (the reference's try/except — the door
	// never breaks anything).
	f.chromeDeliverFn = func(vault string, links []ChromeLinkRow,
		log LogFn, config *storage.OrderedMap) ChromeDeliveryReport {
		panic("the CDP socket exploded")
	}
	rep = s.ChromeDeliverLinks([]ChromeLinkRow{{URL: "https://example.com/x"}})
	if rep.Delivered != 0 || rep.Failed != 1 || len(rep.FailedLinks) != 1 ||
		rep.FailedLinks[0].Error != "Chrome tab retry failed: the CDP socket exploded" {
		t.Fatalf("the crash report: %+v", rep)
	}
	wantLogAt(t, rec, "⚠️ Chrome tab retry failed: the CDP socket exploded", "warning", 2)
}

func TestChromeDeliveryFailedBoxGrammar(t *testing.T) {
	// nil when nothing failed (nothing to say).
	if box := ChromeDeliveryFailedBox(nil); box != nil {
		t.Fatalf("the empty box: %+v", box)
	}
	// the reference's own modal grammar, five rows by name.
	links := []ChromeLinkRow{
		{URL: "https://example.com/one", Error: "challenge page"},
		{URL: "https://example.com/two", Error: ""},
		{URL: "", Error: "the tab crashed"},
	}
	box := ChromeDeliveryFailedBox(links)
	if box.Title != "Chrome tab retry — pages not taken" {
		t.Fatalf("the title: %q", box.Title)
	}
	want := "3 link(s) didn't generate content from your Chrome (the tabs crashed, showed an error page, or never finished loading).\n\n" +
		"They keep waiting in the retry queue — the next run offers them again (More ▸ 🤖 Chrome tab-retry anytime).\n\n" +
		"• https://example.com/one — challenge page\n" +
		"• https://example.com/two\n" +
		"• (a link) — the tab crashed"
	if box.Body != want {
		t.Fatalf("the body:\n%q\nwant:\n%q", box.Body, want)
	}
	// the >5 tail: the first five + the … line.
	seven := make([]ChromeLinkRow, 7)
	for i := range seven {
		seven[i] = ChromeLinkRow{URL: fmt.Sprintf("https://example.com/%d", i)}
	}
	box = ChromeDeliveryFailedBox(seven)
	if !strings.HasSuffix(box.Body, "… and 2 more") ||
		strings.Count(box.Body, "• https://") != 5 {
		t.Fatalf("the >5 tail: %q", box.Body)
	}
}

// ---------------------------------------------------------------------------
// 🪄 Attach to my Chrome (processing_control.py:669).
// ---------------------------------------------------------------------------

func TestAttachMyChromeLadder(t *testing.T) {
	s, rec, f := newChromeDoorRig(t)

	// no Chrome on this machine: the remedy box, nothing else.
	out := s.AttachMyChrome(false)
	if out.Box == nil || out.Box.Title != "No Chrome found" ||
		out.Box.Body != "GitCurator could not find Google Chrome on this machine — the fifth door needs the real browser." {
		t.Fatalf("the no-chrome box: %+v", out.Box)
	}

	// a Chrome but no profile: the second remedy box.
	f.chromeFind = "/usr/bin/google-chrome"
	out = s.AttachMyChrome(false)
	if out.Box == nil || out.Box.Title != "No Chrome profile found" ||
		out.Box.Body != "GitCurator could not find your Chrome profile (the usual location is empty) — nothing to attach to yet." {
		t.Fatalf("the no-profile box: %+v", out.Box)
	}

	// the bat could not be written: the write-failure box.
	f.chromeDataDir = t.TempDir()
	f.chromeAttachBatOK = false
	out = s.AttachMyChrome(false)
	if out.Box == nil || out.Box.Title != "The attach script could not be written" ||
		!strings.HasPrefix(out.Box.Body, "GitCurator could not write ") ||
		!strings.HasSuffix(out.Box.Body, " — check the folder's permissions and try again.") {
		t.Fatalf("the write-failure box: %+v", out.Box)
	}

	// the platform split (the reference's os.name law): POSIX answers
	// the ready line + box without asking; Windows asks first.
	f.chromeAttachBatOK = true
	out = s.AttachMyChrome(false)
	batPath := filepath.Join(appDirOf(t), "chrome-attach.bat")
	if isWindowsRuntime() {
		if out.Confirm == nil || out.Confirm.Title != "Attach the fifth door to MY Chrome" ||
			!strings.Contains(out.Confirm.Body, "GitCurator will now close and restart your Chrome with its DevTools port open") {
			t.Fatalf("the windows confirm: %+v", out.Confirm)
		}
	} else {
		if out.Box == nil || out.Box.Title != "The attach script is ready" ||
			!strings.Contains(out.Box.Body, "A Windows .bat (this machine is not Windows) — start your Chrome with --remote-debugging-port=9222 and every delivery attaches to it automatically.") {
			t.Fatalf("the posix box: %+v", out.Box)
		}
		wantLogAt(t, rec, "🪄 The attach script is ready: "+batPath+" — a Windows .bat (on this machine, start your Chrome yourself with --remote-debugging-port=9222 and the door attaches automatically)", "info", 1)
	}

	// the port coercion: every junk shape reads the default (9222),
	// a speaking value rides the box's line.
	for _, tc := range []struct {
		raw  any
		want int
	}{
		{nil, 9222}, {false, 9222}, {int64(0), 9222}, {"", 9222},
		{int64(9300), 9300}, {float64(9300), 9300}, {"9300", 9300},
		{"junk", 9222}, {float64(9.5), 9},
	} {
		if got := attachPortCoerce(tc.raw); got != tc.want {
			t.Fatalf("attachPortCoerce(%v) = %d, want %d", tc.raw, got, tc.want)
		}
	}
	s.cfg.Set("web_browser_attach_port", 9300)
	f.chromeAttachBatOK = true
	_ = s.AttachMyChrome(false)
	if !isWindowsRuntime() {
		if l, ok := rec.findLog("--remote-debugging-port=9300"); !ok {
			t.Fatalf("the configured port rides the line: %q", l.Msg)
		}
	}

	// the windows legs (compile-checked everywhere; exercised on a
	// windows box — this battery runs the POSIX half).
	if !isWindowsRuntime() {
		return
	}
	out = s.AttachMyChrome(true)
	if out.Box == nil || out.Box.Title != "Your Chrome is being restarted with the door attached" {
		t.Fatalf("the restart box: %+v", out.Box)
	}
	f.chromeAttachRunErr = "the launch failed"
	out = s.AttachMyChrome(true)
	if out.Box == nil || out.Box.Title != "Run it by hand" ||
		!strings.Contains(out.Box.Body, "The script is ready but could not be launched (the launch failed).") {
		t.Fatalf("the run-failure box: %+v", out.Box)
	}
}

// appDirOf answers the test's temp APP_DIR (constants.AppDir over
// the GITCURATOR_HOME the rig set).
func appDirOf(t *testing.T) string {
	t.Helper()
	return constants.AppDir()
}

// isWindowsRuntime answers the platform split (the reference's
// os.name law — the ladder's legs are exercised per platform; the
// other half is compile-checked by the GOOS=windows build).
func isWindowsRuntime() bool { return runtime.GOOS == "windows" }

// ---------------------------------------------------------------------------
// THE FIELD-TAG LAW — the payload tags mirror the keys the frontend
// reads (pinned by marshalling every shape).
// ---------------------------------------------------------------------------

func TestChromeDoorPayloadFieldTags(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"ChromeLinkRow", ChromeLinkRow{URL: "u", Error: "e"},
			`{"url":"u","error":"e"}`},
		{"HandWallRow", HandWallRow{URL: "u", Wall: "w", Attempts: 3},
			`{"url":"u","wall":"w","attempts":3}`},
		{"HandDeliverScan", HandDeliverScan{Box: &HeroBox{Title: "T"},
			Links: []HandWallRow{{URL: "u"}}, Folder: "f", Queued: 2},
			`{"box":{"title":"T","body":""},"links":[{"url":"u","wall":"","attempts":0}],"folder":"f","queued":2}`},
		{"HandRowsScan", HandRowsScan{Confirm: &HeroBox{Title: "C", Question: true},
			Rows: []ChromeLinkRow{{URL: "u"}}},
			`{"box":null,"confirm":{"title":"C","body":"","question":true},"rows":[{"url":"u","error":""}]}`},
		{"ChromeTabRetryScan", ChromeTabRetryScan{Links: []ChromeLinkRow{{URL: "u"}}},
			`{"box":null,"links":[{"url":"u","error":""}]}`},
		{"ChromeDeliveryReport", ChromeDeliveryReport{Delivered: 1, Failed: 2,
			URLs: []string{"u"}, FailedLinks: []ChromeLinkRow{{URL: "u", Error: "e"}}, Folder: "f"},
			`{"delivered":1,"failed":2,"urls":["u"],"failed_links":[{"url":"u","error":"e"}],"folder":"f"}`},
		{"AttachChromeResult", AttachChromeResult{Box: &HeroBox{Title: "B"}},
			`{"box":{"title":"B","body":""},"confirm":null}`},
	}
	for _, tc := range cases {
		raw, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatalf("%s marshal: %v", tc.name, err)
		}
		if string(raw) != tc.want {
			t.Fatalf("%s = %s, want %s", tc.name, raw, tc.want)
		}
	}
}
