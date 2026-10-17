// handpipeline_test.go — T-43: the PIPELINE legs of the reference's
// own hand suites (test_handdelivery §7 TestPipelineFourthDoor +
// test_handharvest §4 TestHarvestDelivery + §5 TestProbes) — the
// fourth door's answers through ProcessLink/Run, the owner's exact
// burned-out→gesture→delivery flow, the half-fetched exemption, the
// controls, and the cached table probes. The handdelivery package's
// own battery covers the module surface; these pin the WIRING (door
// 0 answers before any machine door, the reborn counter, the harvest
// hooks, the run's uninvited pull).
package websitepipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/links"
)

const handWalledURL = "https://walled.example.net/article"

var handWalledCanon = links.NormalizeWebsiteURL(handWalledURL)

const handErr = "HTTP 403 — bot defense (server: cloudflare)"

// --- shared helpers (the _HarvestCase pattern) -------------------------

// writeReviewNote drops an app-owned _review note (the reference's
// write_review_note: BuildReviewNote + a plain write).
func (c *pipeCase) writeReviewNote(url, name, fetchStatus, reason string) string {
	c.t.Helper()
	reviewDir := filepath.Join(c.vault, "_review")
	os.MkdirAll(reviewDir, 0o755)
	path := filepath.Join(reviewDir, name)
	note := BuildReviewNote(url, fetchStatus, reason, "", nil)
	if err := os.WriteFile(path, []byte(note), 0o644); err != nil {
		c.t.Fatal(err)
	}
	return path
}

// writeHalfNote drops a LOW-CONFIDENCE half-fetched item: a real
// app-owned _review note whose fetch SUCCEEDED but whose
// classification parked it for human eyes (fetch_status 'full').
func (c *pipeCase) writeHalfNote(url, name string) string {
	c.t.Helper()
	return c.writeReviewNote(url, name, "full",
		"classification confidence was low")
}

// writeProperNote drops a REAL proper, categorized note ON DISK (the
// strict harvest's own bar: the file exists, app-owned, outside
// _review).
func (c *pipeCase) writeProperNote(url string) string {
	c.t.Helper()
	folder := filepath.Join(c.vault, "Design")
	os.MkdirAll(folder, 0o755)
	path := filepath.Join(folder, "proper-note.md")
	note := BuildWebsiteNote(url, map[string]any{
		"name": "Proper Note", "one_line": "A proper note.",
		"what_it_does":  []any{"It works."},
		"best_used_for": "Use it when testing.",
		"pricing":       "free", "login_required": "no",
	}, "Design", "", "full", nil)
	if err := os.WriteFile(path, []byte(note), 0o644); err != nil {
		c.t.Fatal(err)
	}
	return path
}

// handDeliver enqueues the link for hand delivery and drops the
// delivered page into the folder (the owner's save).
func (c *pipeCase) handDeliver(url string) string {
	c.t.Helper()
	handdelivery.EnqueueHandDelivery(c.vault, []string{url},
		map[string]string{url: handErr}, nil, false, "")
	path := filepath.Join(handdelivery.HandDeliveryDir(c.vault),
		handdelivery.SuggestedFilename(url))
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, []byte("<html><head><title>Hand-saved</title></head><body>saved by the owner</body></html>"), 0o644); err != nil {
		c.t.Fatal(err)
	}
	return path
}

// tableDataRows returns the table's data rows (the house writeTable
// writes "| - |" rows; the graveyard header's 🖐/✅ documentation
// lives in " > " blockquotes, not rows — the reference's minimal test
// tables carry no header docs, so its assertNotIn('🖐', table) maps
// to no-ROWS-contain here).
func tableDataRows(table string) []string {
	var rows []string
	for _, line := range strings.Split(table, "\n") {
		if strings.HasPrefix(line, "| - |") {
			rows = append(rows, line)
		}
	}
	return rows
}

// rowContains reports whether any data row contains part.
func rowsContain(table, part string) bool {
	for _, row := range tableDataRows(table) {
		if strings.Contains(row, part) {
			return true
		}
	}
	return false
}

// --- 1. the fourth door answers first (test_handdelivery §7) ----------

func TestPipelineFourthDoorDeliveredPageAnswersBeforeAnyMachineDoor(t *testing.T) {
	c := newPipeCase(t)
	c.db.EnqueueRetry(handWalledURL, handErr)
	c.handDeliver(handWalledURL)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	// the machine fetcher was NEVER asked…
	if len(c.fake.calls) != 0 {
		t.Fatalf("the machine door was asked: %v", c.fake.calls)
	}
	// …the note is a real processed note…
	if r.Outcome != "processed" || r.FetchStatus != "full" {
		t.Fatalf("result: %+v", r)
	}
	if _, err := os.Stat(r.NotePath); err != nil {
		t.Fatalf("note missing: %v", err)
	}
	// …the story is told…
	if !c.hasLogContaining("hand-delivered") {
		t.Errorf("the story is missing:\n%s", c.allLogs())
	}
	// …and the retry row resolved (a successful fetch)
	rr, err := c.db.RetryRow(handWalledCanon)
	if err != nil {
		t.Fatal(err)
	}
	if rr != nil {
		t.Errorf("retry row stayed: %+v", rr)
	}
}

func TestPipelineFourthDoorRunPullsDeliveredPagesIntoTheBatch(t *testing.T) {
	c := newPipeCase(t)
	other := "https://fresh.example.net/page"
	c.handDeliver(handWalledURL)
	results := c.makePipeline().Run([]string{other}, nil, nil)
	urls := map[string]bool{}
	for _, r := range results {
		urls[r.URL] = true
	}
	if !urls[handWalledURL] {
		t.Fatalf("the delivered page did not join the batch: %v", urls)
	}
	walled := map[string]Result{}
	for _, r := range results {
		walled[r.URL] = r
	}
	if walled[handWalledURL].Outcome != "processed" {
		t.Errorf("walled result: %+v", walled[handWalledURL])
	}
	if !c.hasLogContaining("join this batch") {
		t.Errorf("the join log is missing:\n%s", c.allLogs())
	}
}

func TestPipelineFourthDoorWalledFailureGainsTheHintEverywhere(t *testing.T) {
	c := newPipeCase(t)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "review" {
		t.Fatalf("result: %+v", r)
	}
	if !strings.Contains(r.Error, "fourth door") {
		t.Errorf("error lacks the hint: %q", r.Error)
	}
	if !c.hasLogContaining("fourth door") {
		t.Errorf("the log lacks the hint:\n%s", c.allLogs())
	}
	note, err := os.ReadFile(r.NotePath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(note), "fourth door") {
		t.Errorf("the note lacks the hint:\n%s", note)
	}
	// the retry row's last_error carries it too (the picker's data)
	rr, err := c.db.RetryRow(handWalledCanon)
	if err != nil || rr == nil {
		t.Fatal("no retry row")
	}
	if rr.LastError == nil || !strings.Contains(*rr.LastError, "fourth door") {
		t.Errorf("last_error lacks the hint: %v", rr.LastError)
	}
}

func TestPipelineFourthDoorResourceTruthsGetNoHint(t *testing.T) {
	c := newPipeCase(t)
	c.fake.fail("/x", "HTTP 404 — the page is gone", "dead")
	r := c.makePipeline().ProcessLink("https://gone.example.org/x")
	if strings.Contains(r.Error, "fourth door") {
		t.Errorf("a 404 got the door hint: %q", r.Error)
	}
}

func TestPipelineFourthDoorOptOutKillsTheWholeDoor(t *testing.T) {
	c := newPipeCase(t)
	c.extraCfg = map[string]any{"web_hand_delivery": false}
	c.db.EnqueueRetry(handWalledURL, handErr)
	c.handDeliver(handWalledURL)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	// the machine door was asked (the walled answer)…
	if len(c.fake.calls) != 1 || c.fake.calls[0] != handWalledURL {
		t.Fatalf("the machine door was not asked: %v", c.fake.calls)
	}
	if r.Outcome != "review" {
		t.Fatalf("result: %+v", r)
	}
	// …no hint…
	if strings.Contains(r.Error, "fourth door") {
		t.Errorf("error: %q", r.Error)
	}
	// …and the queue row is untouched
	q := handdelivery.ReadQueue(c.vault)
	if m := q.Meta(handWalledCanon); m != nil {
		if v, ok := m.Get("consumed"); ok && v != nil && v != false {
			t.Errorf("the queue row was consumed: %v", v)
		}
	}
}

// --- 2. the harvest delivery (test_handharvest §4) --------------------

func TestHarvestDeliveryTheOwnersExactFlowBurnedOutThenDelivered(t *testing.T) {
	c := newPipeCase(t)
	// walled → 3 retries burned out → the owner sets 🖐 → the page
	// is delivered → the re-run completes it all
	for i := 0; i < 3; i++ {
		c.db.EnqueueRetry(handWalledCanon, handErr)
	}
	rr, _ := c.db.RetryRow(handWalledCanon)
	if rr == nil || rr.Attempts != 3 {
		t.Fatalf("setup: attempts %v", rr)
	}
	placeholder := c.writeReviewNote(handWalledURL, "a.md", "failed", "Fetch failed: "+handErr)
	c.inVault[handWalledCanon] = true
	c.writeTable([][2]string{{handWalledURL, "🖐 hand — queued 2026-10-01"}})
	c.handDeliver(handWalledURL)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	// the machine door was never asked — the delivered page answered
	if len(c.fake.calls) != 0 {
		t.Fatalf("the machine door was asked: %v", c.fake.calls)
	}
	if r.Outcome != "processed" {
		t.Fatalf("result: %+v", r)
	}
	if _, err := os.Stat(r.NotePath); err != nil {
		t.Fatalf("note missing: %v", err)
	}
	// THE HARVEST: the row retired, the placeholder swept
	table := c.tableText()
	if !rowsContain(table, "✅ hand-delivered — fetched") {
		t.Errorf("the row did not retire:\n%s", table)
	}
	if rowsContain(table, "🖐") {
		t.Errorf("the gesture stayed:\n%s", table)
	}
	if _, err := os.Stat(placeholder); err == nil {
		t.Error("the placeholder was not swept")
	}
	if !c.hasLogContaining("the hand's harvest") {
		t.Errorf("the harvest log is missing:\n%s", c.allLogs())
	}
	if !c.hasLogContaining("reborn") {
		t.Errorf("the reborn log is missing:\n%s", c.allLogs())
	}
	// the link is DONE: the retired row is final for every scan
	if got := ScanMasterHandRows(c.vault, nil); len(got) != 0 {
		t.Errorf("hand rows: %+v", got)
	}
	for _, w := range ScanMasterWaitingRows(c.vault, nil, nil) {
		if w.URL == handWalledURL || w.URL == handWalledCanon {
			t.Errorf("the retired row still waits: %+v", w)
		}
	}
}

func TestHarvestDeliveryTheHandAnswerWithAFreshRetry(t *testing.T) {
	c := newPipeCase(t)
	// the simpler shape: first failure, the gesture, the delivery
	c.db.EnqueueRetry(handWalledCanon, handErr)
	placeholder := c.writeReviewNote(handWalledURL, "a.md", "failed", "Fetch failed: "+handErr)
	c.inVault[handWalledCanon] = true
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	c.handDeliver(handWalledURL)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "processed" {
		t.Fatalf("result: %+v", r)
	}
	if !rowsContain(c.tableText(), "✅ hand-delivered") {
		t.Errorf("table:\n%s", c.tableText())
	}
	if _, err := os.Stat(placeholder); err == nil {
		t.Error("the placeholder was not swept")
	}
}

func TestHarvestDeliveryTheHalfFetchedExemptionCompletesAParkedLink(t *testing.T) {
	c := newPipeCase(t)
	// a 🖐 link parked in _review with a NON-failed note (low
	// confidence) was skipped "already in the websites vault" forever
	// — the gesture now says FINISH THIS ONE
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.inVault[handWalledCanon] = true
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	c.handDeliver(handWalledURL)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "processed" {
		t.Fatalf("result: %+v (err %q)", r, r.Error)
	}
	if !rowsContain(c.tableText(), "✅ hand-delivered") {
		t.Errorf("table:\n%s", c.tableText())
	}
	if _, err := os.Stat(half); err == nil {
		t.Error("the half note was not swept")
	}
	if !c.hasLogContaining("the hand's harvest") {
		t.Errorf("the harvest log is missing:\n%s", c.allLogs())
	}
}

func TestHarvestDeliveryTheSameParkedLinkWithoutTheGestureStillSkips(t *testing.T) {
	c := newPipeCase(t)
	// the old law holds for every link without the gesture — the
	// exemption is the hand's, not a blanket change
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.inVault[handWalledCanon] = true
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	c.writeTable([][2]string{{handWalledURL, " - "}})
	c.handDeliver(handWalledURL)
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "skipped" || !strings.Contains(r.Error, "already in the websites vault") {
		t.Fatalf("result: %+v (err %q)", r, r.Error)
	}
	if _, err := os.Stat(half); err != nil {
		t.Errorf("the half note was swept: %v", err)
	}
}

func TestHarvestDeliveryAStoredLinkWithoutTheGestureIsTheControl(t *testing.T) {
	c := newPipeCase(t)
	proper := c.writeProperNote(handWalledURL)
	c.inVault[handWalledCanon] = true
	c.db.MarkProcessed(handWalledCanon, proper, "Design", "", "full")
	c.writeTable([][2]string{{handWalledURL, " - "}})
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "skipped" {
		t.Fatalf("result: %+v (err %q)", r, r.Error)
	}
}

func TestHarvestDeliveryBurnedOutWithoutGestureOrPageKeepsTheOldLaw(t *testing.T) {
	c := newPipeCase(t)
	for i := 0; i < 3; i++ {
		c.db.EnqueueRetry(handWalledCanon, handErr)
	}
	placeholder := c.writeReviewNote(handWalledURL, "a.md", "failed", "Fetch failed: "+handErr)
	c.inVault[handWalledCanon] = true
	c.db.MarkProcessed(handWalledCanon, placeholder, "", "", "failed")
	c.writeTable([][2]string{{handWalledURL, " - "}})
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "skipped" || !strings.Contains(r.Error, "no more retries") {
		t.Fatalf("result: %+v (err %q)", r, r.Error)
	}
}

func TestHarvestDeliveryADeliveredPageAloneRebornsTheCounter(t *testing.T) {
	c := newPipeCase(t)
	// the picker's queue (a delivered page waiting) reborns even
	// without a hand-marked row — the page IS the ask
	for i := 0; i < 3; i++ {
		c.db.EnqueueRetry(handWalledCanon, handErr)
	}
	placeholder := c.writeReviewNote(handWalledURL, "a.md", "failed", "Fetch failed: "+handErr)
	c.inVault[handWalledCanon] = true
	c.db.MarkProcessed(handWalledCanon, placeholder, "", "", "failed")
	c.handDeliver(handWalledURL)
	c.fake.fail("/article", handErr, "blocked_bot")
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "processed" {
		t.Fatalf("result: %+v (err %q)", r, r.Error)
	}
	if len(c.fake.calls) != 0 {
		t.Fatalf("the machine door was asked: %v", c.fake.calls)
	}
	if !c.hasLogContaining("reborn") {
		t.Errorf("the reborn log is missing:\n%s", c.allLogs())
	}
}

func TestHarvestDeliveryAMachineSuccessOnAHandRowIsHarvestedToo(t *testing.T) {
	c := newPipeCase(t)
	// the ladder got through on its own — the law is about the ROW
	// (the owner's gesture), not the door that answered
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	r := c.makePipeline().ProcessLink(handWalledURL)
	if r.Outcome != "processed" {
		t.Fatalf("result: %+v (err %q)", r, r.Error)
	}
	if !rowsContain(c.tableText(), "✅ hand-delivered") {
		t.Errorf("table:\n%s", c.tableText())
	}
	if !c.hasLogContaining("the hand's harvest") {
		t.Errorf("the harvest log is missing:\n%s", c.allLogs())
	}
}

func TestHarvestDeliveryRunEndsWithTheRefreshHarvest(t *testing.T) {
	c := newPipeCase(t)
	// the belt-and-braces half: the batch's last move retires the row
	// even if the step-7 hook was somehow skipped — and the prune then
	// carries the terminal verdict away (the table keeps only what
	// still needs the owner)
	proper := c.writeProperNote(handWalledURL)
	c.db.MarkProcessed(handWalledCanon, proper, "Design", "", "full")
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	other := "https://fresh.example.net/page"
	c.makePipeline().Run([]string{other}, nil, nil)
	if !c.hasLogContaining("retired to the green checkbox") {
		t.Errorf("the retirement log is missing:\n%s", c.allLogs())
	}
	if rowsContain(c.tableText(), handWalledURL) {
		t.Errorf("the terminal row stayed:\n%s", c.tableText())
	}
}

// --- 3. the probes (test_handharvest §5) -------------------------------

func TestProbesRowReadsHandByTheTablesGrammar(t *testing.T) {
	c := newPipeCase(t)
	p := c.makePipeline()
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if !p.rowReadsHand(handWalledCanon) {
		t.Error("a hand row does not read as hand")
	}
	for _, s := range []string{"🪦 dead", "✅ reviewed", "♻️ revived", " - ", "unreviewed"} {
		c := newPipeCase(t)
		p := c.makePipeline()
		c.writeTable([][2]string{{handWalledURL, s}})
		if p.rowReadsHand(handWalledCanon) {
			t.Errorf("status %q reads as hand", s)
		}
	}
}

func TestProbesRowReadsHandNoTableNoRow(t *testing.T) {
	c := newPipeCase(t)
	p := c.makePipeline()
	if p.rowReadsHand(handWalledCanon) {
		t.Error("no table reads as hand")
	}
	c.writeTable([][2]string{{"https://other.example.org/x", "🖐 hand"}})
	if p.rowReadsHand(handWalledCanon) {
		t.Error("another link's row reads as this link's hand")
	}
}

func TestProbesTheCacheRefreshesAfterAMidRunStamp(t *testing.T) {
	c := newPipeCase(t)
	p := c.makePipeline()
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if !p.rowReadsHand(handWalledCanon) {
		t.Fatal("setup: the row does not read as hand")
	}
	handdelivery.StampHandDeliveredRows(c.vault, []string{handWalledCanon}, nil)
	// the mtime-keyed cache re-reads — a retired row is never
	// re-harvested
	if p.rowReadsHand(handWalledCanon) {
		t.Error("the cache did not refresh after the stamp")
	}
}

func TestProbesHasAppReviewNoteAnyStatus(t *testing.T) {
	c := newPipeCase(t)
	p := c.makePipeline()
	c.writeReviewNote(handWalledURL, "a.md", "failed", "Fetch failed: "+handErr)
	if !p.hasAppReviewNote(handWalledCanon) {
		t.Error("a failed placeholder is not an app review note")
	}
	c2 := newPipeCase(t)
	p2 := c2.makePipeline()
	c2.writeHalfNote(handWalledURL, "a.md")
	if !p2.hasAppReviewNote(handWalledCanon) {
		t.Error("a half-fetched note is not an app review note")
	}
}

func TestProbesHasAppReviewNoteHandWrittenIsNo(t *testing.T) {
	c := newPipeCase(t)
	p := c.makePipeline()
	path := filepath.Join(c.vault, "_review", "owners.md")
	note := fmt.Sprintf("---\nsource: %s\n---\n\nby hand\n", handWalledURL)
	if err := os.WriteFile(path, []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	if p.hasAppReviewNote(handWalledCanon) {
		t.Error("a hand-written note read as app-owned")
	}
}

func TestProbesHasAppReviewNoteMissingFolderIsNo(t *testing.T) {
	c := newPipeCase(t)
	p := c.makePipeline()
	os.RemoveAll(filepath.Join(c.vault, "_review"))
	if p.hasAppReviewNote(handWalledCanon) {
		t.Error("a missing folder read as app-owned")
	}
}

// the fakeFetch's options shape rides pipeline_test's own webfetch
// import; nothing else needed here.
