package app

// fanfaredoors_test.go — T-79f's engine tests: the three automatic
// triggers' data halves, pinned against the reference's own laws at
// 98f4fc7 (processing_control.py:475/1197 + hero.py:280 +
// _scan_master_redo's hermetic guards).

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// writeDeliveredPage seeds the fourth door's delivered set: the queue
// row (collect_delivered reads the QUEUE, not loose page files) + the
// page file at the queue's own suggested filename.
func writeDeliveredPage(t *testing.T, vault, url, wall string) {
	t.Helper()
	handdelivery.EnqueueHandDelivery(vault, []string{url},
		map[string]string{url: wall}, nil, false, "")
	dir := handdelivery.HandDeliveryDir(vault)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	page := filepath.Join(dir, handdelivery.SuggestedFilename(url))
	if err := os.WriteFile(page,
		[]byte("<html><head><title>Hand</title></head><body>delivered</body></html>"),
		0o644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// (a1) HandQueuePass — the end-of-run hand pass.
// ---------------------------------------------------------------------------

func TestFanfareHandQueuePass(t *testing.T) {
	s, rec, _ := newChromeDoorRig(t)
	_ = rec

	// the guards are QUIET: the knob off, the pipeline off, the vault
	// missing — nil every time, never a box, never a line (the pass
	// never costs the owner his scorecard).
	s.cfg.Set("web_browser_retry", false)
	if links := s.HandQueuePass(); links != nil {
		t.Fatalf("the knob opts the automatic pass out: %+v", links)
	}
	s.cfg.Set("web_browser_retry", true)
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	if links := s.HandQueuePass(); links != nil {
		t.Fatalf("the pipeline guard: %+v", links)
	}
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	s.cfg.Set("website_vault_path", "/not/a/vault")
	if links := s.HandQueuePass(); links != nil {
		t.Fatalf("the vault guard: %+v", links)
	}
	vault := t.TempDir()
	s.cfg.Set("website_vault_path", vault)
	if links := s.HandQueuePass(); links != nil {
		t.Fatalf("an empty table is a quiet no-op: %+v", links)
	}

	// two 🖐 rows: the links carry the wall (or the gesture) + the
	// tried set marks them — the SECOND pass answers nothing for
	// them, but a NEW row still earns its pass.
	writeHandTable(t, vault, [][3]string{
		{"https://example.com/hand-one", "🖐", ""},
		{"https://example.com/hand-two", "🖐", "HTTP 403"},
	})
	links := s.HandQueuePass()
	if len(links) != 2 {
		t.Fatalf("the rows: %+v", links)
	}
	if links[0].URL != "https://example.com/hand-one" ||
		links[0].Error != "the 🖐 hand gesture in the master table" {
		t.Fatalf("the gesture default: %+v", links[0])
	}
	if links[1].URL != "https://example.com/hand-two" ||
		links[1].Error != "HTTP 403" {
		t.Fatalf("the wall rides: %+v", links[1])
	}
	// once per session per link — the second pass is quiet.
	if again := s.HandQueuePass(); again != nil {
		t.Fatalf("the tried belt: %+v", again)
	}
	// a NEW row still passes.
	writeHandTable(t, vault, [][3]string{
		{"https://example.com/hand-one", "🖐", ""},
		{"https://example.com/hand-two", "🖐", "HTTP 403"},
		{"https://example.com/hand-three", "🖐", "HTTP 429"},
	})
	links = s.HandQueuePass()
	if len(links) != 1 || links[0].URL != "https://example.com/hand-three" {
		t.Fatalf("only the new row passes: %+v", links)
	}
}

// ---------------------------------------------------------------------------
// (a2) CaughtUpHandRows — hero.py's caught-up scrape (NO tried-set).
// ---------------------------------------------------------------------------

func TestFanfareCaughtUpHandRows(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)
	v, _ := s.cfg.Get("website_vault_path")
	vault := v.(string)
	writeHandTable(t, vault, [][3]string{
		{"https://example.com/hand-one", "🖐", "HTTP 403"},
	})
	links := s.CaughtUpHandRows(false)
	if len(links) != 1 || links[0].URL != "https://example.com/hand-one" ||
		links[0].Error != "HTTP 403" {
		t.Fatalf("the caught-up rows: %+v", links)
	}
	// THE CAUGHT-UP LAW: no tried-set — a second sync scrapes the same
	// rows again (hero.py:280-281 filters nothing).
	if again := s.CaughtUpHandRows(false); len(again) != 1 {
		t.Fatalf("the caught-up pass has NO tried belt: %+v", again)
	}
	// a running batch owns the table's truth (the scan's own guard).
	if busy := s.CaughtUpHandRows(true); busy != nil {
		t.Fatalf("the batch guard: %+v", busy)
	}
	// the pipeline + vault guards stay quiet.
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	if links := s.CaughtUpHandRows(false); links != nil {
		t.Fatalf("the pipeline guard: %+v", links)
	}
}

// ---------------------------------------------------------------------------
// (b) RedoHandNotesScan — the redo pass (THE NOTE IS THE SUCCESS).
// ---------------------------------------------------------------------------

func TestFanfareRedoHandNotesScan(t *testing.T) {
	s, rec, _ := newChromeDoorRig(t)
	v, _ := s.cfg.Get("website_vault_path")
	vault := v.(string)

	// the quiet guards: a gathering delivery + a running batch both
	// defer (the delivery's/running batch's own re-run owns the links).
	writeHandTable(t, vault, [][3]string{
		{"https://example.com/false-success", "🖐", "HTTP 403"},
	})
	writeDeliveredPage(t, vault, "https://example.com/false-success", "HTTP 403")
	if rows := s.RedoHandNotesScan(true, false); rows != nil {
		t.Fatalf("a gathering delivery defers the redo: %+v", rows)
	}
	if rows := s.RedoHandNotesScan(false, true); rows != nil {
		t.Fatalf("a running batch defers the redo: %+v", rows)
	}

	// the false success: the page was delivered but NO processed row
	// ever landed — the redo row + the 🔁 line with the first reason.
	rows := s.RedoHandNotesScan(false, false)
	if len(rows) != 1 || rows[0].URL != "https://example.com/false-success" {
		t.Fatalf("the redo rows: %+v", rows)
	}
	wantLog(t, rec, "🔁 1 hand note(s) NOT properly stored — refetching and "+
		"re-generating them now (the delivered page is re-read, the LLM writes "+
		"the proper, categorized note; a ✅ is only earned by the note itself; "+
		"first reason: no processed row — the note never landed)", "info")

	// once per session per link — the second scan is quiet (the LLM's
	// answer will not change on an immediate second ask).
	rec.logs = nil
	if rows := s.RedoHandNotesScan(false, false); rows != nil {
		t.Fatalf("the redo tried belt: %+v", rows)
	}
	if l, ok := rec.findLog("🔁"); ok {
		t.Fatalf("the tried belt is quiet: %+v", l)
	}

	// the guards stay quiet everywhere: the pipeline off + the vault
	// missing.
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	if rows := s.RedoHandNotesScan(false, false); rows != nil {
		t.Fatalf("the pipeline guard: %+v", rows)
	}
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	s.cfg.Set("website_vault_path", "/not/a/vault")
	if rows := s.RedoHandNotesScan(false, false); rows != nil {
		t.Fatalf("the vault guard: %+v", rows)
	}
	s.cfg.Set("website_vault_path", vault)
}

// ---------------------------------------------------------------------------
// (c) ChromeOfferFromResults — the end-of-run offer's data.
// ---------------------------------------------------------------------------

func TestFanfareChromeOfferFromResults(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)

	entries := []githubpipeline.WebsiteResultEntry{
		{URL: "https://example.com/failed", Outcome: "failed",
			Error: "connection reset"},
		{URL: "https://review.example/half", Outcome: "review",
			FetchStatus: "failed", Error: "HTTP 403 — bot defense"},
		{URL: "https://example.com/processed", Outcome: "processed"},
		{URL: "https://example.com/failed", Outcome: "failed",
			Error: "the duplicate drops"},
		{URL: "http://127.0.0.1:8080/local", Outcome: "failed",
			Error: "loopback never asks"},
		{URL: "ftp://not-http.example/x", Outcome: "failed"},
	}

	// the quiet guards first: the knob, the pipeline, the vault.
	s.cfg.Set("web_browser_retry", false)
	if links := s.ChromeOfferFromResults(entries); links != nil {
		t.Fatalf("the knob guard: %+v", links)
	}
	s.cfg.Set("web_browser_retry", true)
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	if links := s.ChromeOfferFromResults(entries); links != nil {
		t.Fatalf("the pipeline guard: %+v", links)
	}
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	s.cfg.Set("website_vault_path", "/not/a/vault")
	if links := s.ChromeOfferFromResults(entries); links != nil {
		t.Fatalf("the vault guard: %+v", links)
	}
	vault := t.TempDir()
	s.cfg.Set("website_vault_path", vault)

	// the collect: the hard miss + the half-fetched review row both
	// ask; deduped; loopback + non-http drop; processed never asks.
	links := s.ChromeOfferFromResults(entries)
	if len(links) != 2 {
		t.Fatalf("the offer rows: %+v", links)
	}
	if links[0].URL != "https://example.com/failed" ||
		links[0].Error != "connection reset" {
		t.Fatalf("the first row (run order, first wins): %+v", links[0])
	}
	if links[1].URL != "https://review.example/half" ||
		links[1].Error != "HTTP 403 — bot defense" {
		t.Fatalf("the review row: %+v", links[1])
	}

	// the fetcher's own retired verdict never nags: the dismissed
	// filter drops its row (the is_dismissed probe).
	state := chromeDoorState(t)
	if err := state.Dismiss("https://example.com/failed",
		"auto-verdict: dead — HTTP 404"); err != nil {
		t.Fatal(err)
	}
	state.Close()
	links = s.ChromeOfferFromResults(entries)
	if len(links) != 1 || links[0].URL != "https://review.example/half" {
		t.Fatalf("the dismissed filter: %+v", links)
	}

	// the session's offered set: a link the owner picked is never
	// re-offered; a DECLINED link was never marked (the answer door
	// marks only the picked).
	s.MarkChromeTabOffered([]ChromeLinkRow{{URL: "https://review.example/half"}})
	if links := s.ChromeOfferFromResults(entries); links != nil {
		t.Fatalf("the offered filter: %+v", links)
	}
}

// TestFanfareChromeOfferNilResults pins the no-summary shape: a batch
// whose websites phase never ran (nil results) offers nothing — the
// reference's `if not summary: return`.
func TestFanfareChromeOfferNilResults(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)
	if links := s.ChromeOfferFromResults(nil); links != nil {
		t.Fatalf("nil results offer nothing: %+v", links)
	}
}

// TestFanfareOfferFromStateKeepsUnprovableLinks pins the degraded
// probe's law: a state question that cannot be asked still OFFERS the
// links (the reference's `except Exception: pass`).
func TestFanfareOfferFromStateKeepsUnprovableLinks(t *testing.T) {
	s, _, _ := newChromeDoorRig(t)
	entries := []githubpipeline.WebsiteResultEntry{
		{URL: "https://provable.example/one", Outcome: "failed", Error: "e1"},
		{URL: "https://unprovable.example/two", Outcome: "failed", Error: "e2"},
	}
	// mark the FIRST link dismissed in the REAL state DB; then point
	// the scan at an unopenable DB path so the whole probe degrades —
	// BOTH links must be offered (the broken probe never hides a
	// link), the dismiss record notwithstanding.
	state := chromeDoorState(t)
	if err := state.Dismiss("https://provable.example/one",
		"auto-verdict: dead"); err != nil {
		t.Fatal(err)
	}
	state.Close()
	// break the state DB itself: a DIRECTORY at cache.db makes every
	// open fail — the degraded probe's law: offer them anyway.
	dbPath := filepath.Join(constants.AppDir(), "cache.db")
	_ = state.Close()
	_ = os.Remove(dbPath)
	if err := os.Mkdir(dbPath, 0o755); err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(dbPath)
	links := s.ChromeOfferFromResults(entries)
	if len(links) != 2 {
		t.Fatalf("the degraded probe offers everything: %+v", links)
	}
}

// TestFanfareWebsiteResultsMirrorKeys pins THE FIELD-TAG LAW on the
// summary's result rows (the tags mirror the pipeline Result dict's
// own keys at 98f4fc7 — url/canonical/outcome/note_path/category/
// subcategory/fetch_status/error) + the BatchFinished crossing's own
// new tags (the marshal-compare shape of TestCrossedPayloadFieldTags).
func TestFanfareWebsiteResultsMirrorKeys(t *testing.T) {
	entry := githubpipeline.WebsiteResultEntry{
		URL: "https://example.com/x", Canonical: "https://example.com/x",
		Outcome: "review", NotePath: "/v/_review/half.md",
		Category: "Design", Subcategory: "Assets & Resources",
		FetchStatus: "failed", Error: "HTTP 403"}
	raw, err := json.Marshal(entry)
	if err != nil {
		t.Fatal(err)
	}
	want := `{"url":"https://example.com/x","canonical":"https://example.com/x",` +
		`"outcome":"review","note_path":"/v/_review/half.md",` +
		`"category":"Design","subcategory":"Assets \u0026 Resources",` +
		`"fetch_status":"failed","error":"HTTP 403"}`
	if string(raw) != want {
		t.Fatalf("WebsiteResultEntry = %s, want %s", raw, want)
	}
	// the crossing's own new tags.
	raw, err = json.Marshal(BatchFinished{OK: true, Summary: "done",
		NeedsRetry: true, Natural: true,
		WebsiteResults: []githubpipeline.WebsiteResultEntry{entry}})
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{
		`"natural":true`, `"stopped":false`, `"website_results":[{`,
		`"fetch_status":"failed"`,
	} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("BatchFinished missing %s: %s", key, raw)
		}
	}
}
