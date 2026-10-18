// handdelivery_test.go — T-43: the ported behavior battery — the
// reference's own suites (tests/test_handdelivery.py +
// tests/test_handharvest.py + the test_iconcolumn pending legs),
// selections that belong to THIS package (the pipeline-level legs of
// the same suites live with websitepipeline's battery; the
// release-bookkeeping legs are the Python repo's own ritual).
//
// The unit law (the reference's own): temp vaults, patched seams,
// zero sockets, no browser ever launches.
package handdelivery

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

const (
	wallURL    = "https://walled.example.net/article"
	wallReason = "HTTP 403 — bot defense (server: cloudflare; challenge) | chrome-impersonated: HTTP 403"
)

// handCase mirrors _HandCase: temp vault, dry-run reset, cleanup.
type handCase struct {
	t     *testing.T
	tmp   string
	vault string
}

func newHandCase(t *testing.T) *handCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	t.Cleanup(func() { dryrun.Disable() })
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "websites")
	os.MkdirAll(vault, 0o755)
	return &handCase{t: t, tmp: tmp, vault: vault}
}

// queueFile drops a queue.json row for url (enqueue's write side).
func (c *handCase) queueFile(url string, wall string) {
	c.t.Helper()
	EnqueueHandDelivery(c.vault, []string{url}, map[string]string{url: wall}, nil, false, "")
}

// deliver writes the delivered page for url.
func (c *handCase) deliver(url string, body []byte) string {
	c.t.Helper()
	path := filepath.Join(HandDeliveryDir(c.vault), SuggestedFilename(url))
	os.MkdirAll(filepath.Dir(path), 0o755)
	if err := os.WriteFile(path, body, 0o644); err != nil {
		c.t.Fatal(err)
	}
	return path
}

func (c *handCase) readQueueJSON() map[string]any {
	c.t.Helper()
	data, err := os.ReadFile(QueuePath(c.vault))
	if err != nil {
		c.t.Fatal(err)
	}
	var q map[string]any
	if err := json.Unmarshal(data, &q); err != nil {
		c.t.Fatal(err)
	}
	return q
}

// --- 1. the wall predicate (test_handdelivery §1) ----------------------

func TestWallPredicateRefusalFamilyEarnsTheDoor(t *testing.T) {
	for _, reason := range []string{
		"HTTP 403", "proxy: HTTP 405 | direct: HTTP 405",
		"HTTP 429 — bot defense", "HTTP 451",
	} {
		if !IsWalledReason(reason) {
			t.Errorf("is_walled_reason(%q) = false, want true", reason)
		}
	}
}

func TestWallPredicateBotDefenseAndChallenge(t *testing.T) {
	for _, reason := range []string{
		"HTTP 403 — bot defense (server: cloudflare)",
		"the challenge needs a live browser",
	} {
		if !IsWalledReason(reason) {
			t.Errorf("is_walled_reason(%q) = false, want true", reason)
		}
	}
}

func TestWallPredicateTLSAndThirdDoorLines(t *testing.T) {
	for _, reason := range []string{
		"proxy: connection: [SSL: SSLV3_ALERT_HANDSHAKE_FAILURE]",
		"chrome-impersonated: HTTP 403",
		"a third door exists: pip install curl_cffi",
	} {
		if !IsWalledReason(reason) {
			t.Errorf("is_walled_reason(%q) = false, want true", reason)
		}
	}
}

func TestWallPredicateResourceTruthsNeverAsk(t *testing.T) {
	for _, reason := range []string{
		"HTTP 404 — the page is gone", "HTTP 402 — paywalled",
		"HTTP 523", "timeout", "connection: reset by peer",
		"not an http(s) URL",
	} {
		if IsWalledReason(reason) {
			t.Errorf("is_walled_reason(%q) = true, want false", reason)
		}
	}
}

func TestWallPredicateEmptyIsNotAWall(t *testing.T) {
	if IsWalledReason("") {
		t.Error("empty reason is not a wall")
	}
}

func TestWalledRetryRowsFiltersTheQueue(t *testing.T) {
	state := fakeRowsState{rows: []RetrySourceRow{
		{URL: wallURL, LastError: wallReason, Attempts: 2},
		{URL: "https://dead.example.org/x", LastError: "HTTP 404 — gone", Attempts: 1},
		{URL: "https://ok.example.org/y", LastError: "timeout", Attempts: 1},
	}}
	out := WalledRetryRows(state)
	if len(out) != 1 || out[0].URL != wallURL {
		t.Fatalf("walled_retry_rows: %+v", out)
	}
	if out[0].Wall != wallReason || out[0].Attempts != 2 {
		t.Errorf("row shape: %+v", out[0])
	}
}

func TestWalledRetryRowsStateFailureMeansEmptyPicker(t *testing.T) {
	if got := WalledRetryRows(errRowsState{}); got != nil {
		t.Errorf("erroring state: %+v", got)
	}
	if got := WalledRetryRows(panicRowsState{}); got != nil {
		t.Errorf("panicking state: %+v", got)
	}
	if got := WalledRetryRows(nil); got != nil {
		t.Errorf("nil state: %+v", got)
	}
}

type errRowsState struct{}

func (errRowsState) AllRetryRows() ([]RetrySourceRow, error) {
	return nil, errors.New("db locked")
}

type panicRowsState struct{}

func (panicRowsState) AllRetryRows() ([]RetrySourceRow, error) {
	panic("db locked")
}

// --- 2. the suggested filename (test_handdelivery §2) ------------------

func TestSuggestedFilenameDeterministic(t *testing.T) {
	a := SuggestedFilename(wallURL)
	b := SuggestedFilename(wallURL)
	if a != b {
		t.Errorf("suggested_filename not deterministic: %q vs %q", a, b)
	}
	if !strings.HasPrefix(a, "hand-") || !strings.HasSuffix(a, ".html") {
		t.Errorf("shape: %q", a)
	}
}

func TestSuggestedFilenameTwoPagesNeverCollide(t *testing.T) {
	a := SuggestedFilename("https://site.example.net/a")
	b := SuggestedFilename("https://site.example.net/b")
	if a == b {
		t.Errorf("collision on one domain: %q", a)
	}
	if SuggestedFilename("") == SuggestedFilename("x") {
		t.Error("empty and x collide")
	}
}

// --- 3. the queue, write side (test_handdelivery §3) -------------------

func TestEnqueueWritesQueueAndReadme(t *testing.T) {
	c := newHandCase(t)
	report := EnqueueHandDelivery(c.vault, []string{wallURL},
		map[string]string{wallURL: wallReason}, nil, false, "")
	if report.Added != 1 || report.QueuedTotal != 1 {
		t.Fatalf("report: %+v", report)
	}
	q := c.readQueueJSON()
	links := q["links"].(map[string]any)
	meta := links[wallURL].(map[string]any)
	if meta["wall"] != wallReason {
		t.Errorf("wall: %v", meta["wall"])
	}
	if meta["suggested"] != SuggestedFilename(wallURL) {
		t.Errorf("suggested: %v", meta["suggested"])
	}
	readme, err := os.ReadFile(ReadmePath(c.vault))
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{wallURL, SuggestedFilename(wallURL), wallReason[:40]} {
		if !strings.Contains(string(readme), want) {
			t.Errorf("README lacks %q:\n%s", want, readme)
		}
	}
}

func TestEnqueueReenqueueIsANoopMerge(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	report := EnqueueHandDelivery(c.vault, []string{wallURL}, nil, nil, false, "")
	if report.Added != 0 || report.QueuedTotal != 1 {
		t.Errorf("report: %+v", report)
	}
}

func TestEnqueueReenqueueOfConsumedResetsIt(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	q := ReadQueue(c.vault)
	if m := q.Meta(wallURL); m != nil {
		m.Set("consumed", "2026-10-08 10:00")
	}
	if err := q.WriteAtomic(c.vault); err != nil {
		t.Fatal(err)
	}
	report := EnqueueHandDelivery(c.vault, []string{wallURL}, nil, nil, false, "")
	if report.Added != 1 {
		t.Fatalf("report: %+v", report)
	}
	entry := c.readQueueJSON()["links"].(map[string]any)[wallURL].(map[string]any)
	// reset: the consumed stamp is cleared (the owner may deliver
	// again) and the requeue is stamped.
	if v, ok := entry["consumed"]; ok && v != false {
		t.Errorf("consumed not reset: %v", v)
	}
	if v, ok := entry["requeued"].(string); !ok || v == "" {
		t.Errorf("requeued not stamped: %v", entry["requeued"])
	}
}

func TestEnqueueNonHTTPURLsAreNeverQueued(t *testing.T) {
	c := newHandCase(t)
	report := EnqueueHandDelivery(c.vault,
		[]string{"not-a-url", "ftp://x.example.net/a"}, nil, nil, false, "")
	if report.Added != 0 {
		t.Errorf("report: %+v", report)
	}
}

func TestEnqueueCorruptQueueJSONIsTolerated(t *testing.T) {
	c := newHandCase(t)
	os.MkdirAll(HandDeliveryDir(c.vault), 0o755)
	os.WriteFile(QueuePath(c.vault), []byte("{ not json !!"), 0o644)
	report := EnqueueHandDelivery(c.vault, []string{wallURL}, nil, nil, false, "")
	if report.Added != 1 {
		t.Errorf("the corrupt file is replaced: %+v", report)
	}
}

func TestEnqueueDryRunRecordsWithoutWriting(t *testing.T) {
	c := newHandCase(t)
	dryrun.Enable()
	t.Cleanup(dryrun.Disable)
	report := EnqueueHandDelivery(c.vault, []string{wallURL}, nil, nil, false, "")
	if report.Added != 1 {
		t.Fatalf("report: %+v", report)
	}
	if _, err := os.Stat(QueuePath(c.vault)); err == nil {
		t.Error("dry-run wrote the queue")
	}
	if dryrun.EntryCount() < 2 {
		t.Errorf("dry-run records: %d", dryrun.EntryCount())
	}
}

func TestEnqueueEmptyVaultIsASilentNoop(t *testing.T) {
	report := EnqueueHandDelivery("", []string{wallURL}, nil, nil, false, "")
	if report.Added != 0 {
		t.Errorf("report: %+v", report)
	}
}

func TestEnqueueOpenChromeLegCountsTheOpened(t *testing.T) {
	c := newHandCase(t)
	patchChromeSeams(t,
		func() string { return "/fake/chrome" },
		func(argv []string) error { return nil },
		nil)
	report := EnqueueHandDelivery(c.vault, []string{wallURL}, nil, nil, true, "")
	if report.Added != 1 || report.Opened != 1 {
		t.Errorf("report: %+v", report)
	}
}

// --- 4. the real Chrome (test_handdelivery §4) -------------------------

// patchChromeSeams patches the open ladder's seams (FindChromeFn /
// SubprocessPopenFn / WebbrowserOpenFn — the reference patches
// hd.find_chrome, hd.subprocess.Popen and hd.webbrowser.open) and
// restores them at cleanup.
func patchChromeSeams(t *testing.T,
	find func() string,
	popen func(argv []string) error,
	wopen func(u string) (bool, error),
) {
	t.Helper()
	origFind, origPopen, origWopen := FindChromeFn, SubprocessPopenFn, WebbrowserOpenFn
	FindChromeFn = find
	if popen != nil {
		SubprocessPopenFn = popen
	}
	if wopen != nil {
		WebbrowserOpenFn = wopen
	}
	t.Cleanup(func() {
		FindChromeFn, SubprocessPopenFn, WebbrowserOpenFn = origFind, origPopen, origWopen
	})
}

func TestOpenInChromeLoopbackNeverOpensABrowser(t *testing.T) {
	popenCalls, wopenCalls := 0, 0
	patchChromeSeams(t,
		func() string { return "/fake/chrome" },
		func(argv []string) error { popenCalls++; return nil },
		func(u string) (bool, error) { wopenCalls++; return true, nil })
	out := OpenInChrome("http://127.0.0.1:3000/x", nil)
	if out != "" {
		t.Errorf("loopback: %q", out)
	}
	if popenCalls != 0 || wopenCalls != 0 {
		t.Errorf("a browser opened: popen=%d wopen=%d", popenCalls, wopenCalls)
	}
	// the documented law's other members (localhost / ::1 / .local)
	for _, u := range []string{"http://localhost:8080/", "http://[::1]:9090/", "https://machine.local/"} {
		patchChromeSeams(t,
			func() string { return "/fake/chrome" },
			func(argv []string) error { popenCalls++; return nil },
			func(u string) (bool, error) { wopenCalls++; return true, nil })
		if out := OpenInChrome(u, nil); out != "" {
			t.Errorf("loopback-family %q: %q", u, out)
		}
	}
	if popenCalls != 0 || wopenCalls != 0 {
		t.Errorf("a browser opened: popen=%d wopen=%d", popenCalls, wopenCalls)
	}
}

func TestOpenInChromeFoundChromeLaunchesDetached(t *testing.T) {
	var argvs [][]string
	patchChromeSeams(t,
		func() string { return "/fake/chrome" },
		func(argv []string) error { argvs = append(argvs, argv); return nil },
		nil)
	out := OpenInChrome(wallURL, nil)
	if out != "chrome" {
		t.Errorf("out: %q", out)
	}
	if len(argvs) != 1 || len(argvs[0]) != 2 ||
		argvs[0][0] != "/fake/chrome" || argvs[0][1] != wallURL {
		t.Errorf("popen argv: %v", argvs)
	}
}

func TestOpenInChromeSystemBrowserIsTheHonestFallback(t *testing.T) {
	popenCalls, wopenCalls := 0, 0
	patchChromeSeams(t,
		func() string { return "" },
		func(argv []string) error { popenCalls++; return nil },
		func(u string) (bool, error) { wopenCalls++; return true, nil })
	out := OpenInChrome(wallURL, nil)
	if out != "browser" {
		t.Errorf("out: %q", out)
	}
	if popenCalls != 0 {
		t.Error("the chrome ladder ran without a chrome")
	}
	if wopenCalls != 1 {
		t.Errorf("webbrowser calls: %d", wopenCalls)
	}
}

func TestOpenInChromeALaunchFailureNeverRaises(t *testing.T) {
	patchChromeSeams(t,
		func() string { return "/fake/chrome" },
		func(argv []string) error { return errors.New("nope") },
		nil)
	if out := OpenInChrome(wallURL, nil); out != "" {
		t.Errorf("out: %q", out)
	}
}

func TestOpenInChromeNoBrowserAtAll(t *testing.T) {
	patchChromeSeams(t,
		func() string { return "" },
		nil,
		func(u string) (bool, error) { return false, nil })
	if out := OpenInChrome(wallURL, nil); out != "" {
		t.Errorf("out: %q", out)
	}
	// webbrowser raising lands in the same honest warning as a failed
	// Chrome launch (the reference's one except path)
	patchChromeSeams(t,
		func() string { return "" },
		nil,
		func(u string) (bool, error) { return false, errors.New("no browser") })
	if out := OpenInChrome(wallURL, nil); out != "" {
		t.Errorf("out: %q", out)
	}
}

func TestOpenInChromeNonHTTPNeverOpens(t *testing.T) {
	popenCalls := 0
	patchChromeSeams(t,
		func() string { return "/fake/chrome" },
		func(argv []string) error { popenCalls++; return nil },
		nil)
	if out := OpenInChrome("ftp://x.example.net/a", nil); out != "" {
		t.Errorf("out: %q", out)
	}
	if popenCalls != 0 {
		t.Error("a non-http url launched a browser")
	}
}

func TestFindChromeReadsTheWindowsCandidates(t *testing.T) {
	origOS, origIsFile, origWhich := chromeOSName, chromeIsFile, chromeWhich
	chromeOSName = func() string { return "windows" }
	chromeIsFile = func(p string) bool {
		return strings.HasSuffix(p, "chrome.exe")
	}
	chromeWhich = func(name string) string { return "" }
	t.Cleanup(func() { chromeOSName, chromeIsFile, chromeWhich = origOS, origIsFile, origWhich })
	found := FindChrome()
	if found == "" || !strings.HasSuffix(found, "chrome.exe") {
		t.Errorf("found: %q", found)
	}
}

func TestFindChromePosixWhichLadder(t *testing.T) {
	origOS, origIsFile, origWhich := chromeOSName, chromeIsFile, chromeWhich
	chromeOSName = func() string { return "linux" }
	chromeIsFile = func(p string) bool { return false }
	chromeWhich = func(name string) string {
		if name == "google-chrome" {
			return "/usr/bin/google-chrome"
		}
		return ""
	}
	t.Cleanup(func() { chromeOSName, chromeIsFile, chromeWhich = origOS, origIsFile, origWhich })
	if found := FindChrome(); found != "/usr/bin/google-chrome" {
		t.Errorf("found: %q", found)
	}
	// nothing found → empty (the honest fallback's door)
	chromeWhich = func(name string) string { return "" }
	if found := FindChrome(); found != "" {
		t.Errorf("found: %q", found)
	}
}

// --- 5. the consume side (test_handdelivery §5) ------------------------

func TestConsumeDeliveredPageIsCollectedAndConsumed(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	path := c.deliver(wallURL, []byte("<html><head><title>Hand-saved</title></head><body>saved by the owner</body></html>"))
	collected := CollectDelivered(c.vault)
	if len(collected) != 1 || collected[0].URL != wallURL {
		t.Fatalf("collect: %+v", collected)
	}
	taken := ConsumeDelivered(c.vault, nil)
	if len(taken) != 1 || taken[0].URL != wallURL {
		t.Fatalf("taken: %+v", taken)
	}
	if !strings.Contains(string(taken[0].Body), "Hand-saved") {
		t.Errorf("body: %q", taken[0].Body)
	}
	// the FILE stays — it is the record…
	if _, err := os.Stat(path); err != nil {
		t.Errorf("the record was removed: %v", err)
	}
	// …and the queue is stamped consumed
	entry := c.readQueueJSON()["links"].(map[string]any)[wallURL].(map[string]any)
	if v, ok := entry["consumed"].(string); !ok || v == "" {
		t.Errorf("consumed stamp: %v", entry["consumed"])
	}
	// a second collect sees nothing (already consumed)
	if got := CollectDelivered(c.vault); len(got) != 0 {
		t.Errorf("second collect: %+v", got)
	}
}

func TestConsumeMissingFileIsNotDelivered(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	if got := CollectDelivered(c.vault); len(got) != 0 {
		t.Errorf("collect: %+v", got)
	}
}

func TestConsumeEmptyPageKeepsWaiting(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	c.deliver(wallURL, []byte("   "))
	if got := ConsumeDelivered(c.vault, nil); len(got) != 0 {
		t.Fatalf("taken: %+v", got)
	}
	entry := c.readQueueJSON()["links"].(map[string]any)[wallURL].(map[string]any)
	if v, ok := entry["consumed"]; ok && v != false {
		t.Errorf("consumed stamped: %v", v)
	}
}

func TestConsumeOrphanFilesAreIgnored(t *testing.T) {
	c := newHandCase(t)
	os.MkdirAll(HandDeliveryDir(c.vault), 0o755)
	os.WriteFile(filepath.Join(HandDeliveryDir(c.vault), "random.html"),
		[]byte("<html></html>"), 0o644)
	if got := CollectDelivered(c.vault); len(got) != 0 {
		t.Errorf("collect: %+v", got)
	}
	if got := ConsumeDelivered(c.vault, nil); len(got) != 0 {
		t.Errorf("taken: %+v", got)
	}
}

func TestConsumeDryRunTakesThePagesWithoutStamping(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	c.deliver(wallURL, []byte("<html><body>real</body></html>"))
	dryrun.Enable()
	t.Cleanup(dryrun.Disable)
	taken := ConsumeDelivered(c.vault, nil)
	if len(taken) != 1 {
		t.Fatalf("taken: %+v", taken)
	}
	entry := c.readQueueJSON()["links"].(map[string]any)[wallURL].(map[string]any)
	if v, ok := entry["consumed"]; ok && v != false {
		t.Errorf("dry-run stamped: %v", v)
	}
}

func TestConsumeHandFetchResultIsARealFullResultWithTheStory(t *testing.T) {
	res := HandFetchResult(wallURL, []byte("<html></html>"), wallReason, "")
	if !res.OK() || res.Status != "full" {
		t.Fatalf("res: %+v", res)
	}
	if !strings.Contains(res.Reason, "hand-delivered via the owner's real Chrome") {
		t.Errorf("reason: %q", res.Reason)
	}
	if !strings.Contains(res.Reason, wallReason) {
		t.Errorf("reason lacks the wall: %q", res.Reason)
	}
}

// TestTakeHandDeliveredBlankLaw pins the pyBytesBlank reality-wins
// fix: Python's bytes.strip() is ASCII-only — a NBSP-only body is NOT
// blank (the reference consumes it), a spaces-only body IS.
func TestTakeHandDeliveredBlankLaw(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	c.deliver(wallURL, []byte("\xc2\xa0")) // NBSP-only — not blank
	res := TakeHandDelivered(c.vault, wallURL, nil, false)
	if res == nil {
		t.Fatal("a NBSP-only page is a delivered page for the reference")
	}
	c2 := newHandCase(t)
	c2.queueFile(wallURL, wallReason)
	c2.deliver(wallURL, []byte(" \t\n\r\x0b\x0c "))
	if res := TakeHandDelivered(c2.vault, wallURL, nil, false); res != nil {
		t.Error("an ASCII-whitespace-only page is blank")
	}
}

// --- 6. the master-table gesture grammar (test_handdelivery §6) --------

func TestStatusIsHandGrammar(t *testing.T) {
	for _, s := range []string{"🖐 hand", "Hand-deliver this one"} {
		if !StatusIsHand(s) {
			t.Errorf("_status_is_hand(%q) = false", s)
		}
	}
	for _, s := range []string{"unreviewed", "", "♻️ revived", "✅ reviewed"} {
		if StatusIsHand(s) {
			t.Errorf("_status_is_hand(%q) = true", s)
		}
	}
}

// --- 7. stamp_hand_rows (the queue stamp's own laws) -------------------

func writeHandTable(t *testing.T, vault string, rows [][2]string, lastUpdated string) string {
	t.Helper()
	path := decomtable.DecommissionTablePath(vault)
	os.MkdirAll(filepath.Dir(path), 0o755)
	lines := []string{
		"# Review Master Table — decommission or approve",
		"",
	}
	if lastUpdated != "" {
		lines = append(lines, "> Last updated: "+lastUpdated, "")
	}
	lines = append(lines,
		"| # | Date | URL | Domain | Source | Status | Notes |",
		"|---|------|-----|--------|--------|--------|-------|")
	for _, row := range rows {
		lines = append(lines, fmt.Sprintf("| - | 2026-10-01 | %s | x.example | test | %s | |", row[0], row[1]))
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestStampHandRowsStampsTheWantedRows(t *testing.T) {
	c := newHandCase(t)
	path := writeHandTable(t, c.vault, [][2]string{
		{"https://a.example/1", "unreviewed"},
		{"https://b.example/2", "hand"},
		{"https://c.example/3", "unreviewed"},
	}, "2001-01-01 00:00")
	stamped := StampHandRows(c.vault, []string{"https://a.example/1", "https://b.example/2"}, nil)
	if stamped != 2 {
		t.Fatalf("stamped: %d", stamped)
	}
	body, _ := os.ReadFile(path)
	text := string(body)
	if !strings.Contains(text, "| 🖐 hand — queued ") {
		t.Errorf("table:\n%s", text)
	}
	// exactly the two wanted rows stamped; the third untouched
	if got := strings.Count(text, "🖐 hand — queued"); got != 2 {
		t.Errorf("stamped rows: %d\n%s", got, text)
	}
	if !strings.Contains(text, "unreviewed") {
		t.Errorf("the unw anted row was touched:\n%s", text)
	}
	if strings.Contains(text, "2001-01-01") {
		t.Error("the Last-updated line did not refresh")
	}
	if !strings.Contains(text, "> Last updated: ") {
		t.Error("the Last-updated line missing")
	}
}

func TestStampHandRowsMissingTableAndEmptyURLsReadZero(t *testing.T) {
	c := newHandCase(t)
	if got := StampHandRows(c.vault, []string{"https://a.example/1"}, nil); got != 0 {
		t.Errorf("missing table: %d", got)
	}
	writeHandTable(t, c.vault, [][2]string{{"https://a.example/1", "hand"}}, "")
	if got := StampHandRows(c.vault, nil, nil); got != 0 {
		t.Errorf("empty urls: %d", got)
	}
}

func TestStampHandRowsShortRowsAreSkipped(t *testing.T) {
	c := newHandCase(t)
	path := decomtable.DecommissionTablePath(c.vault)
	os.MkdirAll(filepath.Dir(path), 0o755)
	// a 3-cell row — parts < 8, the stamp's guard skips it
	os.WriteFile(path, []byte("| - | 2026-10-01 | https://a.example/1 |\n"), 0o644)
	if got := StampHandRows(c.vault, []string{"https://a.example/1"}, nil); got != 0 {
		t.Errorf("short row stamped: %d", got)
	}
}

// --- 8. pending_hand_links (test_iconcolumn §6) ------------------------

func TestPendingQueuedUndeliveredLinks(t *testing.T) {
	c := newHandCase(t)
	u1 := "https://pending.example/one"
	u2 := "https://pending.example/two"
	EnqueueHandDelivery(c.vault, []string{u1, u2},
		map[string]string{u1: wallReason}, nil, false, "")
	pending := PendingHandLinks(c.vault)
	if len(pending) != 2 {
		t.Fatalf("pending: %+v", pending)
	}
	byURL := map[string]string{}
	for _, p := range pending {
		byURL[p.URL] = p.Wall
	}
	if byURL[u1] != wallReason {
		t.Errorf("wall for %s: %q", u1, byURL[u1])
	}
	if _, ok := byURL[u2]; !ok {
		t.Errorf("missing %s: %+v", u2, pending)
	}
}

func TestPendingConsumedAndLandedLinksAreNotPending(t *testing.T) {
	c := newHandCase(t)
	u1 := "https://done.example/one"
	u2 := "https://landed.example/two"
	EnqueueHandDelivery(c.vault, []string{u1, u2}, nil, nil, false, "")
	q := ReadQueue(c.vault)
	if m := q.Meta(u1); m != nil {
		m.Set("consumed", "2026-10-07 20:00")
	}
	if err := q.WriteAtomic(c.vault); err != nil {
		t.Fatal(err)
	}
	c.deliver(u2, []byte("<html>landed</html>"))
	if got := PendingHandLinks(c.vault); len(got) != 0 {
		t.Errorf("pending: %+v", got)
	}
}

func TestPendingABrokenQueueReadsEmpty(t *testing.T) {
	c := newHandCase(t)
	os.MkdirAll(HandDeliveryDir(c.vault), 0o755)
	os.WriteFile(QueuePath(c.vault), []byte("{not json"), 0o644)
	if got := PendingHandLinks(c.vault); len(got) != 0 {
		t.Errorf("pending: %+v", got)
	}
}

func TestPendingAnEmptyFileCountsAsLanded(t *testing.T) {
	c := newHandCase(t)
	u := "https://blank.example/x"
	EnqueueHandDelivery(c.vault, []string{u}, nil, nil, false, "")
	c.deliver(u, []byte("   "))
	// the file landed (isfile is the only gate) — the link is not the
	// fifth door's to-do anymore
	if got := PendingHandLinks(c.vault); len(got) != 0 {
		t.Errorf("pending: %+v", got)
	}
}

// --- 9. the harvest stamp (test_handharvest §1) ------------------------

func TestHarvestStampTheHandRowRetiresToTheGreenCheckbox(t *testing.T) {
	c := newHandCase(t)
	writeHandTable(t, c.vault, [][2]string{
		{wallURL, "🖐 hand — queued 2026-10-01"},
	}, "")
	stamped := StampHandDeliveredRows(c.vault, []string{wallURL}, nil)
	if stamped != 1 {
		t.Fatalf("stamped: %d", stamped)
	}
	body, _ := os.ReadFile(decomtable.DecommissionTablePath(c.vault))
	text := string(body)
	if !strings.Contains(text, "✅ hand-delivered — fetched") {
		t.Errorf("table:\n%s", text)
	}
	if strings.Contains(text, "🖐") {
		t.Errorf("the gesture stayed:\n%s", text)
	}
	// the retired cell is the table's own SUCCESS verdict — the
	// grammar reads it as reviewed (final, never re-queued); the
	// reviewed verdict itself implies not-waiting (the waiting
	// predicate's first retirement check)
	for _, line := range strings.Split(text, "\n") {
		if strings.HasPrefix(line, "| - |") {
			status := strings.Split(line, "|")[6]
			if !decomtable.StatusIsReviewed(status) {
				t.Errorf("status %q is not reviewed", status)
			}
			trimmed := strings.TrimSpace(status)
			if trimmed == "" || trimmed == "-" || trimmed == " - " {
				t.Errorf("status %q reads as waiting", status)
			}
		}
	}
}

func TestHarvestStampTheGestureFormsAllRetire(t *testing.T) {
	for _, s := range []string{"🖐", "✋ hand", "hand", "🖐 hand — queued 2026-10-01"} {
		c := newHandCase(t)
		writeHandTable(t, c.vault, [][2]string{{wallURL, s}}, "")
		if got := StampHandDeliveredRows(c.vault, []string{wallURL}, nil); got != 1 {
			t.Errorf("gesture %q: stamped %d", s, got)
			continue
		}
		body, _ := os.ReadFile(decomtable.DecommissionTablePath(c.vault))
		if !strings.Contains(string(body), "✅ hand-delivered") {
			t.Errorf("gesture %q: table:\n%s", s, body)
		}
	}
}

func TestHarvestStampIdempotentOnceRetired(t *testing.T) {
	c := newHandCase(t)
	writeHandTable(t, c.vault, [][2]string{{wallURL, "🖐 hand"}}, "")
	if got := StampHandDeliveredRows(c.vault, []string{wallURL}, nil); got != 1 {
		t.Fatalf("first: %d", got)
	}
	if got := StampHandDeliveredRows(c.vault, []string{wallURL}, nil); got != 0 {
		t.Errorf("second: %d", got)
	}
}

func TestHarvestStampAWaitingRowIsNeverTheHarvests(t *testing.T) {
	c := newHandCase(t)
	writeHandTable(t, c.vault, [][2]string{{wallURL, " - "}}, "")
	if got := StampHandDeliveredRows(c.vault, []string{wallURL}, nil); got != 0 {
		t.Errorf("stamped: %d", got)
	}
	body, _ := os.ReadFile(decomtable.DecommissionTablePath(c.vault))
	if !strings.Contains(string(body), " - ") {
		t.Errorf("the waiting cell was touched:\n%s", body)
	}
}

func TestHarvestStampDeathReviewedRevivedWin(t *testing.T) {
	for _, s := range []string{"🪦 dead", "✅ reviewed", "♻️ revived"} {
		c := newHandCase(t)
		writeHandTable(t, c.vault, [][2]string{{wallURL, s}}, "")
		if got := StampHandDeliveredRows(c.vault, []string{wallURL}, nil); got != 0 {
			t.Errorf("status %q: stamped %d", s, got)
			continue
		}
		body, _ := os.ReadFile(decomtable.DecommissionTablePath(c.vault))
		if !strings.Contains(string(body), s) {
			t.Errorf("status %q was touched:\n%s", s, body)
		}
	}
}

func TestHarvestStampURLsMatchByCanonicalForm(t *testing.T) {
	c := newHandCase(t)
	rowURL := "https://www.walled.example.net/article"
	writeHandTable(t, c.vault, [][2]string{{rowURL, "🖐 hand"}}, "")
	// called with the bare form — the canonical law bridges them
	if got := StampHandDeliveredRows(c.vault, []string{"https://walled.example.net/article"}, nil); got != 1 {
		t.Errorf("stamped: %d", got)
	}
}

func TestHarvestStampMissingTableAndEmptyURLsReadZero(t *testing.T) {
	c := newHandCase(t)
	if got := StampHandDeliveredRows(c.vault, []string{wallURL}, nil); got != 0 {
		t.Errorf("missing table: %d", got)
	}
	writeHandTable(t, c.vault, [][2]string{{wallURL, "🖐 hand"}}, "")
	if got := StampHandDeliveredRows(c.vault, nil, nil); got != 0 {
		t.Errorf("empty urls: %d", got)
	}
}

func TestHarvestStampTheLastUpdatedLineRefreshes(t *testing.T) {
	c := newHandCase(t)
	writeHandTable(t, c.vault, [][2]string{{wallURL, "🖐 hand"}}, "2001-01-01 00:00")
	StampHandDeliveredRows(c.vault, []string{wallURL}, nil)
	body, _ := os.ReadFile(decomtable.DecommissionTablePath(c.vault))
	text := string(body)
	if !strings.Contains(text, "Last updated: 20") {
		t.Errorf("no refreshed Last-updated:\n%s", text)
	}
	if strings.Contains(text, "2001-01-01") {
		t.Errorf("the stale date stayed:\n%s", text)
	}
}

func TestHarvestStampOtherRowsAreUntouched(t *testing.T) {
	c := newHandCase(t)
	other := "https://other.example.org/x"
	writeHandTable(t, c.vault, [][2]string{
		{other, " - "},
		{wallURL, "🖐 hand"},
	}, "")
	StampHandDeliveredRows(c.vault, []string{wallURL}, nil)
	body, _ := os.ReadFile(decomtable.DecommissionTablePath(c.vault))
	text := string(body)
	if !strings.Contains(text, " - ") {
		t.Errorf("the waiting row was touched:\n%s", text)
	}
	if !strings.Contains(text, "✅ hand-delivered") {
		t.Errorf("the hand row did not retire:\n%s", text)
	}
}

// --- 10. the queue's wire format round-trip (the OrderedMap law) -------

func TestQueueDumpsIsPyJSON(t *testing.T) {
	c := newHandCase(t)
	c.queueFile(wallURL, wallReason)
	raw, err := os.ReadFile(QueuePath(c.vault))
	if err != nil {
		t.Fatal(err)
	}
	q := ReadQueue(c.vault)
	if q.Dumps() != string(raw) {
		t.Errorf("queue round-trip diverges:\n--- dumps ---\n%s\n--- file ---\n%s", q.Dumps(), raw)
	}
	// the stamp's own write path rides the same encoder
	m := q.Meta(wallURL)
	if m == nil {
		t.Fatal("meta missing")
	}
	m.Set("consumed", "2026-10-17 12:00")
	if err := storage.AtomicWriteText(QueuePath(c.vault), q.Dumps()); err != nil {
		t.Fatal(err)
	}
	raw2, _ := os.ReadFile(QueuePath(c.vault))
	if q.Dumps() != string(raw2) {
		t.Errorf("stamped round-trip diverges:\n%s", raw2)
	}
}
