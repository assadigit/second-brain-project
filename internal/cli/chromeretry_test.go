package cli

// chromeretry_test.go — SESSION-058: the chrome-retry LIVE leg's tests
// (the port of test_noteredo's TestCLIRouting selections — the
// reference's own three, plus the post-batch ⚠️ leg the first_errs
// composition rides). All hermetic at construction (THE HERMETIC-SEAM
// law): the whole delivery rides the ChromeDeliverFn seam (the
// reference patches deliver_pages_via_chrome at the module level), the
// batch engine rides RunBatchFn (mock.patch.object's twin), the GitHub
// preflight leg rides PreflightGithub, and the cloud provider's LLM
// leg needs no seam at all (the reference's own _CliCase config
// shape: llm_provider=cloud, a token present).

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/chrometabs"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// chromeRetryRig — the shared plumbing: a temp home + the webvault +
// the passing-preflight config + the walled cache.db (the two retry
// rows) + the GitHub preflight fake.
type chromeRetryRig struct {
	c       *CLI
	out     *bytes.Buffer
	vault   string
	cfgPath string
}

func newChromeRetryRig(t *testing.T) *chromeRetryRig {
	t.Helper()
	home := t.TempDir()
	t.Setenv("GITCURATOR_HOME", home)
	t.Setenv("GITCURATOR_USER_CONFIG", filepath.Join(home, "user-config.json"))
	vault := filepath.Join(home, "webvault")
	cfg := dictOf("website_vault_path", vault)
	pipes := dictOf("websites", true)
	pipes.Set("github", false)
	cfg.Set("pipelines", pipes)
	cfg.Set("llm_provider", "cloud")
	cfg.Set("cloud_api_url", "http://x")
	cfg.Set("cloud_model", "m")
	cfg.Set("github_token", "fake-github-token-0123456789")
	cfgPath := filepath.Join(home, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	craftWalledDB(t, filepath.Join(home, "cache.db"))
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	c.PreflightGithub = func(string) PreflightGHReply {
		return PreflightGHReply{Status: 200, Login: "octo-owner"}
	}
	return &chromeRetryRig{c: c, out: buf, vault: vault, cfgPath: cfgPath}
}

// writeHandRow crafts the master table with one row (the app rig's
// grammar: the GraveyardHeader + the row line; the domain cell derived
// from the URL).
func writeHandRow(t *testing.T, vault, url, status, notes string) {
	t.Helper()
	domain := "unknown"
	if i := strings.Index(url, "://"); i >= 0 {
		rest := url[i+3:]
		if j := strings.IndexAny(rest, "/?#"); j >= 0 {
			rest = rest[:j]
		}
		if rest != "" {
			domain = rest
		}
	}
	path := websitepipeline.DecommissionTablePath(vault)
	content := fmt.Sprintf(websitepipeline.GraveyardHeader, "2026-10-09 12:00") +
		fmt.Sprintf("| - | 2026-10-09 | %s | %s | test | %s | %s |\n",
			url, domain, status, notes)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeDeliveredPage enqueues the URL + writes its page file (the file
// is the record: an unconsumed queue row with a page on disk counts as
// delivered — the reference's own law).
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

// TestChromeRetryLiveRoutingLaw — the reference's
// test_delivered_pages_ride_as_website_links: a delivered page rides
// the re-run as a WEBSITE link (non_github), NEVER as a GitHub url —
// the v0.57.0 routing fix (the owner's "the LLM does not actually
// work on that website" was exactly the old call shape feeding website
// URLs to the GitHub loop).
func TestChromeRetryLiveRoutingLaw(t *testing.T) {
	r := newChromeRetryRig(t)
	r.c.ChromeDeliverFn = func(_ string, _ []chrometabs.Link,
		_ chrometabs.LogFn, _ map[string]any) chrometabs.DeliverReport {
		return chrometabs.DeliverReport{
			Delivered: 1,
			URLs:      []string{"https://example.com/walled"},
		}
	}
	var gotOpts *RunBatchOptions
	r.c.RunBatchFn = func(_ *storage.OrderedMap, opts RunBatchOptions) int {
		o := opts
		gotOpts = &o
		fmt.Fprintf(r.c.Out, "[BATCH] mode=%s urls=%v non_github=%v\n",
			opts.Mode, opts.URLs, opts.NonGithub)
		return 0
	}
	rc := r.c.Run([]string{"--chrome-retry", "--config", r.cfgPath})
	if rc != 0 {
		t.Fatalf("rc = %d, want 0\n%s", rc, r.out.String())
	}
	if gotOpts == nil {
		t.Fatal("the batch never ran")
	}
	if len(gotOpts.URLs) != 0 {
		t.Fatalf("THE ROUTING LAW: website URLs must NEVER ride the GitHub "+
			"urls: %v", gotOpts.URLs)
	}
	if len(gotOpts.NonGithub) != 1 ||
		gotOpts.NonGithub[0] != "https://example.com/walled" {
		t.Fatalf("the delivered page must ride non_github: %v", gotOpts.NonGithub)
	}
	out := r.out.String()
	if !strings.Contains(out, "🤖 1 page(s) delivered from your real "+
		"Chrome — processing them now through the Websites pipeline as "+
		"real fetches…") {
		t.Errorf("missing the delivered card:\n%s", out)
	}
	if strings.Contains(out, "⚠️") {
		t.Errorf("no tab failed — the ⚠️ card must stay quiet:\n%s", out)
	}
}

// TestChromeRetryLiveRedoJoinsRerun — the reference's
// test_the_redo_set_joins_the_rerun: a hand link delivered long ago
// whose note never became proper rides the re-run ALONE (no new tab
// was opened — the delivered page is the record, the re-run re-reads
// it).
func TestChromeRetryLiveRedoJoinsRerun(t *testing.T) {
	r := newChromeRetryRig(t)
	const handURL = "https://example.com/false-success"
	writeHandRow(t, r.vault, handURL, "🖐", "HTTP 403")
	writeDeliveredPage(t, r.vault, handURL, "HTTP 403")
	// nothing NEW was delivered (the page landed long ago)
	r.c.ChromeDeliverFn = func(_ string, _ []chrometabs.Link,
		_ chrometabs.LogFn, _ map[string]any) chrometabs.DeliverReport {
		return chrometabs.DeliverReport{}
	}
	var gotOpts *RunBatchOptions
	r.c.RunBatchFn = func(_ *storage.OrderedMap, opts RunBatchOptions) int {
		o := opts
		gotOpts = &o
		return 0
	}
	rc := r.c.Run([]string{"--chrome-retry", "--config", r.cfgPath})
	if rc != 0 {
		t.Fatalf("rc = %d, want 0\n%s", rc, r.out.String())
	}
	if gotOpts == nil {
		t.Fatal("the batch never ran")
	}
	if len(gotOpts.NonGithub) != 1 || gotOpts.NonGithub[0] != handURL {
		t.Fatalf("the redo set must ride non_github alone: %v",
			gotOpts.NonGithub)
	}
	out := r.out.String()
	if !strings.Contains(out, "🔁 1 hand note(s) not properly stored — "+
		"re-generating them from the delivered pages (the LLM writes "+
		"the proper, categorized notes)") {
		t.Errorf("missing the redo card:\n%s", out)
	}
	if strings.Contains(out, "page(s) delivered from your real Chrome") {
		t.Errorf("nothing NEW was delivered — the delivered card must stay "+
			"quiet:\n%s", out)
	}
}

// TestChromeRetryLiveNothingDelivered — the reference's
// test_no_delivered_and_no_redo_is_the_honest_one: an un-delivered
// hand row (the Chrome pass's set, never the redo's) + a failed
// delivery → the honest card (with the first error riding it) + exit
// 1, and the batch never runs.
func TestChromeRetryLiveNothingDelivered(t *testing.T) {
	r := newChromeRetryRig(t)
	writeHandRow(t, r.vault, "https://example.com/false-success", "🖐", "HTTP 403")
	r.c.ChromeDeliverFn = func(_ string, _ []chrometabs.Link,
		_ chrometabs.LogFn, _ map[string]any) chrometabs.DeliverReport {
		return chrometabs.DeliverReport{
			Failed: 1,
			FailedLinks: []chrometabs.Link{{
				URL: "https://example.com/walled", Error: "chrome refused"}},
		}
	}
	batchRan := false
	r.c.RunBatchFn = func(*storage.OrderedMap, RunBatchOptions) int {
		batchRan = true
		return 0
	}
	rc := r.c.Run([]string{"--chrome-retry", "--config", r.cfgPath})
	if rc != 1 {
		t.Fatalf("rc = %d, want 1\n%s", rc, r.out.String())
	}
	if batchRan {
		t.Error("nothing delivered, nothing to redo — the batch must not run")
	}
	if !strings.Contains(r.out.String(), "🤖 The fifth door delivered "+
		"nothing — 1 tab(s) failed (chrome refused); the links keep "+
		"waiting.") {
		t.Errorf("missing the honest nothing card:\n%s", r.out.String())
	}
}

// TestChromeRetryLiveFailedTabsWarning — the post-batch ⚠️ card: a
// partial delivery (one delivered, one failed with a LONG error)
// prints the warning AFTER the batch, the error sliced to 90 RUNES
// (the reference's [:90] over Python str).
func TestChromeRetryLiveFailedTabsWarning(t *testing.T) {
	r := newChromeRetryRig(t)
	longErr := strings.Repeat("e", 120)
	r.c.ChromeDeliverFn = func(_ string, _ []chrometabs.Link,
		_ chrometabs.LogFn, _ map[string]any) chrometabs.DeliverReport {
		return chrometabs.DeliverReport{
			Delivered: 1,
			URLs:      []string{"https://example.com/walled"},
			Failed:    1,
			FailedLinks: []chrometabs.Link{{
				URL: "https://example.com/plain", Error: longErr}},
		}
	}
	r.c.RunBatchFn = func(_ *storage.OrderedMap, opts RunBatchOptions) int {
		fmt.Fprintf(r.c.Out, "[BATCH] mode=%s\n", opts.Mode)
		return 0
	}
	rc := r.c.Run([]string{"--chrome-retry", "--config", r.cfgPath})
	if rc != 0 {
		t.Fatalf("rc = %d, want 0 (the batch's own verdict)\n%s", rc,
			r.out.String())
	}
	out := r.out.String()
	batchAt := strings.Index(out, "[BATCH] mode=direct")
	warnAt := strings.Index(out, "⚠️ 1 tab(s) failed — those links keep "+
		"waiting in the retry queue. ("+strings.Repeat("e", 90)+")")
	if batchAt < 0 || warnAt < 0 {
		t.Fatalf("the batch marker + the sliced-error warning must both "+
			"print:\n%s", out)
	}
	if batchAt > warnAt {
		t.Errorf("the ⚠️ card must come AFTER the batch run")
	}
}
