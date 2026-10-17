// fifthpipeline_test.go — T-44: the fifth door's PIPELINE leg — the
// reference's test_fifthdoor §8 TestPipelineFifthDoor selection (the
// full circle: the run fails the link → the candidates earn the offer
// → the fifth door delivers the page (a scripted fetcher) → the re-run
// processes it as a REAL fetch with the auto-delivery story).
//
// This file is an EXTERNAL test package (websitepipeline_test)
// because it imports internal/chrometabs (chrometabs imports
// websitepipeline for the candidates — the cycle the external test
// package breaks). Linking chrometabs wires the PageIsRealFn seam for
// the whole test binary — exactly what any real binary that links the
// fifth door gets (the deliver.go init's law).
package websitepipeline_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/chrometabs"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/webfetch"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

const fifthFullURL = "https://walled.example.net/article"
const fifthWallErr = "HTTP 403 — bot defense (server: cloudflare; challenge)"

var fifthCanon = links.NormalizeWebsiteURL(fifthFullURL)

// fifthLLM mirrors the reference's _FakeLLM: the classify prompt gets
// a REAL taxonomy category, the subcategory prompt its answer, the
// analysis prompt the site card — a delivered page must be able to
// reach 'processed', not die in classification.
type fifthLLM struct{}

func (fifthLLM) call(messages []llmclient.Message, task string) string {
	text := messages[0].Content
	if strings.Contains(text, "filing a website into a personal library") {
		return `{"category": "Design", "confidence": "high",
                        "reason": "testing"}`
	}
	if strings.Contains(text, "was filed under") {
		return `{"subcategory": "Assets & Resources",
                        "confidence": "medium"}`
	}
	return `{"name": "Walled Site", "one_line": "A page behind a wall.",
                "core_offerings": ["One"],
                "best_used_for": "Use when you need to beat a bot wall.",
                "pricing": "free", "login_required": "no",
                "similar_tools": [], "tags": ["design"],
                "confidence": "high"}`
}

// fifthFetch mirrors the reference's _FakeFetch: one canned failure
// (the wall), counting its calls.
type fifthFetch struct {
	calls []string
}

func (f *fifthFetch) fetch(u string, _ *webfetch.Options) *webfetch.FetchResult {
	f.calls = append(f.calls, u)
	return &webfetch.FetchResult{URL: u, Status: "failed",
		Reason: fifthWallErr, Category: "blocked_bot"}
}

// fifthCase mirrors the reference's _PipelineCase: a temp vault + the
// state DB + the logs.
type fifthCase struct {
	t       *testing.T
	tmp     string
	vault   string
	db      *websitepipeline.WebsiteStateDB
	inVault map[string]bool
	logs    []string
}

func newFifthCase(t *testing.T) *fifthCase {
	t.Helper()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "websites")
	os.MkdirAll(filepath.Join(vault, "_review"), 0o755)
	db, err := websitepipeline.NewWebsiteStateDB(
		filepath.Join(tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return &fifthCase{t: t, tmp: tmp, vault: vault, db: db,
		inVault: map[string]bool{}}
}

func (c *fifthCase) logFn(msg, level string) { c.logs = append(c.logs, msg) }

func (c *fifthCase) allLogs() string { return strings.Join(c.logs, "\n") }

func (c *fifthCase) makePipeline(fetch *fifthFetch) *websitepipeline.Pipeline {
	cfg := map[string]any{
		"website_vault_path": c.vault,
		"web_domain_delay_s": 0,
		"taxonomy_path":      "../taxonomy/testdata/website-library-categories.md",
	}
	p, err := websitepipeline.NewPipeline(cfg, fifthLLM{}.call,
		func(u string) any {
			if c.inVault[u] {
				return true
			}
			return nil
		},
		c.db, c.logFn, nil, nil, nil, fetch.fetch, nil, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	return p
}

// page writes one landed page at its suggested filename (the
// reference's _FifthCase.page).
func (c *fifthCase) page(url string, body []byte) string {
	c.t.Helper()
	dir := handdelivery.HandDeliveryDir(c.vault)
	os.MkdirAll(dir, 0o755)
	path := filepath.Join(dir, handdelivery.SuggestedFilename(url))
	if err := os.WriteFile(path, body, 0o644); err != nil {
		c.t.Fatal(err)
	}
	return path
}

// TestFifthDoorFullCircleDeliverThenRerun mirrors
// test_the_full_circle_deliver_then_rerun: 1. the run fails the link
// (the machine doors walled); 2. the candidates earn the offer; 3.
// the fifth door delivers the page (a scripted fetcher — the real one
// launches the owner's Chrome; never here); 4. the re-run processes
// the delivered page as a REAL fetch.
func TestFifthDoorFullCircleDeliverThenRerun(t *testing.T) {
	c := newFifthCase(t)

	// 1. the run fails the link (the machine doors walled)
	fetch := &fifthFetch{}
	pipe := c.makePipeline(fetch)
	res := pipe.ProcessLink(fifthFullURL)
	if res.Outcome != "review" {
		t.Fatalf("outcome: %+v", res)
	}
	if res.FetchStatus != "failed" {
		t.Fatalf("fetch_status: %+v", res)
	}

	// 2. the candidates earn the offer
	candidates := chrometabs.CollectFailedFetchLinks(
		pipe.LastResults(), func(string) bool { return false })
	if len(candidates) != 1 || candidates[0].URL != fifthCanon {
		t.Fatalf("candidates: %+v (want %s)", candidates, fifthCanon)
	}

	// 3. the fifth door delivers the page (a scripted fetcher)
	report := chrometabs.DeliverPagesViaChrome(c.vault, candidates,
		nil, nil, "", nil,
		func(urls []string, log chrometabs.LogFn) []chrometabs.PageResult {
			out := make([]chrometabs.PageResult, 0, len(urls))
			for _, u := range urls {
				out = append(out, chrometabs.PageResult{URL: u, OK: true,
					Title: "The Real Page",
					HTML: "<html><head><title>The Real Page</title>" +
						"</head><body>" + strings.Repeat("y", 500) +
						"</body></html>"})
			}
			return out
		})
	if report.Delivered != 1 {
		t.Fatalf("delivered: %+v", report)
	}

	// 4. the re-run processes the delivered page as a REAL fetch
	c.logs = nil
	results := pipe.Run(report.URLs, nil, nil)
	if len(results) != 1 || results[0].Outcome != "processed" {
		t.Fatalf("results: %+v", results)
	}
	if len(fetch.calls) != 1 || fetch.calls[0] != fifthFullURL {
		t.Errorf("fetch calls: %v (only the failed attempt)", fetch.calls)
	}
	if !strings.Contains(c.allLogs(), "auto-delivered") {
		t.Errorf("the auto-delivery story missing from the logs: %v", c.logs)
	}
	row, err := c.db.RetryRow(fifthCanon)
	if err != nil {
		t.Fatal(err)
	}
	if row != nil {
		t.Errorf("the retry row must be resolved: %+v", row)
	}
}

// TestFifthDoorAutoPageStillConsumedWithDoorDisabledByHand mirrors
// test_auto_page_still_consumed_with_door_disabled_by_hand: an
// owner-saved page (no door marker) keeps the manual story.
func TestFifthDoorAutoPageStillConsumedWithDoorDisabledByHand(t *testing.T) {
	c := newFifthCase(t)
	c.db.EnqueueRetry(fifthFullURL, fifthWallErr)
	handdelivery.EnqueueHandDelivery(c.vault, []string{fifthFullURL},
		map[string]string{fifthFullURL: fifthWallErr}, nil, false, "")
	c.page(fifthFullURL, []byte("<html><head><title>By Hand</title>"+
		"</head><body>owner saved me</body></html>"))
	fetch := &fifthFetch{}
	pipe := c.makePipeline(fetch)
	res := pipe.ProcessLink(fifthFullURL)
	if res.Outcome != "processed" {
		t.Fatalf("outcome: %+v", res)
	}
	if !strings.Contains(c.allLogs(), "hand-delivered") {
		t.Errorf("the hand-delivery story missing: %v", c.logs)
	}
}
