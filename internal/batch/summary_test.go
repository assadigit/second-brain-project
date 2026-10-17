// summary_test.go — T-42: the ported legs of tests/test_batchfinish
// .py::TestBuildBatchSummary (the headless __new__ harness — attributes
// pre-set, no batch machinery) plus the pure-shape laws.
package batch

import "testing"

func TestBuildBatchSummaryCleanGitHubBatch(t *testing.T) {
	s := BuildBatchSummary(false, 18, 20, 18, nil, nil,
		"/v/_processing_report_x.md", "/v/_summary.txt")
	if s.Stopped {
		t.Fatalf("stopped on a clean batch")
	}
	if s.GithubProcessed != 18 || s.GithubTotal != 20 || s.NewNotes != 18 {
		t.Fatalf("counts: %+v", s)
	}
	if s.Websites != nil {
		t.Fatalf("websites should be nil when the phase never ran")
	}
	if s.FailedLinks != 0 {
		t.Fatalf("failed links = %d", s.FailedLinks)
	}
	if s.ReportPath != "/v/_processing_report_x.md" ||
		s.SummaryPath != "/v/_summary.txt" {
		t.Fatalf("paths: %+v", s)
	}
}

func TestBuildBatchSummaryStoppedBatchFlagged(t *testing.T) {
	s := BuildBatchSummary(true, 3, 620, 0, nil, nil, "", "")
	if !s.Stopped {
		t.Fatalf("the GUI withholds the fanfare for a stopped batch")
	}
}

func TestBuildBatchSummaryWebsitesCounters(t *testing.T) {
	ws := &WebsitesCounters{Processed: 7, Review: 2, Skipped: 1, Failed: 0}
	s := BuildBatchSummary(false, 0, 0, 0, ws, nil, "", "")
	if s.Websites == nil || s.Websites.Processed != 7 ||
		s.Websites.Review != 2 || s.Websites.Skipped != 1 ||
		s.Websites.Failed != 0 {
		t.Fatalf("websites counters: %+v", s.Websites)
	}
}

func TestBuildBatchSummaryFailedLinksCountedFromManifest(t *testing.T) {
	dir := t.TempDir()
	tracker := NewLinkTracker(dir)
	tracker.Manifest.Links = []*ManifestLink{
		{URL: "https://x/0", Normalized: "x/0", Type: "github",
			Status: "processed"},
		{URL: "https://x/1", Normalized: "x/1", Type: "github",
			Status: "processed"},
		{URL: "https://x/2", Normalized: "x/2", Type: "github",
			Status: "failed"},
		{URL: "https://x/3", Normalized: "x/3", Type: "github",
			Status: "skipped"},
		{URL: "https://x/4", Normalized: "x/4", Type: "github",
			Status: "failed"},
	}
	s := BuildBatchSummary(false, 0, 0, 0, nil, tracker, "", "")
	if s.FailedLinks != 2 {
		t.Fatalf("failed links = %d, want 2", s.FailedLinks)
	}
}

func TestBuildBatchSummaryNoTrackerMeansZeroFailed(t *testing.T) {
	s := BuildBatchSummary(false, 0, 0, 0, nil, nil, "", "")
	if s.FailedLinks != 0 {
		t.Fatalf("no tracker → 0 failed, got %d", s.FailedLinks)
	}
}

func TestBuildBatchSummaryBrokenManifestIsZeroFailed(t *testing.T) {
	// The defensive-read law: a tracker whose manifest is nil reads
	// as zero failed (the Python except swallows the AttributeError).
	tracker := &LinkTracker{Manifest: nil}
	s := BuildBatchSummary(false, 0, 0, 0, nil, tracker, "", "")
	if s.FailedLinks != 0 {
		t.Fatalf("broken manifest → 0 failed, got %d", s.FailedLinks)
	}
}
