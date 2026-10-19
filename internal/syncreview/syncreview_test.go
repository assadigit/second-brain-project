package syncreview

// syncreview_test.go — the ported battery from tests/test_syncreview.py
// (the spec's list): the truth table, the banner precedence incl. the
// setup/measure split, the micro-copy captions, the ledger lifecycle +
// concurrency + never-delete + stable-ids + search, the chip grammar,
// the relative clock, the oldest-first sort, the vault/git/published
// helpers. THE DEFERRED-LEDGER UX TESTS (T-75's acceptance half).

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

func ip(n int) *int { return &n }

// ---------------------------------------------------------------------------
// 1. cell_state — the count form of THE truth table
// ---------------------------------------------------------------------------

func TestCellStateNotSetUpIsUnknown(t *testing.T) {
	if got := CellState(false, ip(5), ip(5), 0); got != StateUnknown {
		t.Fatalf("not set up = %q, want unknown", got)
	}
}

func TestCellStateUnmeasuredSideIsUnknown(t *testing.T) {
	for _, tc := range []struct{ src, tgt *int }{
		{nil, ip(5)}, {ip(5), nil}, {nil, nil},
	} {
		if got := CellState(true, tc.src, tc.tgt, 0); got != StateUnknown {
			t.Fatalf("unmeasured side (%v,%v) = %q, want unknown", tc.src, tc.tgt, got)
		}
	}
}

func TestCellStateZeroGapIsMatch(t *testing.T) {
	for _, tc := range []struct{ src, tgt int }{{5, 5}, {0, 0}} {
		if got := CellState(true, ip(tc.src), ip(tc.tgt), 0); got != StateMatch {
			t.Fatalf("(%d,%d) = %q, want match", tc.src, tc.tgt, got)
		}
	}
}

func TestCellStateTargetAheadIsMatchNeverNegative(t *testing.T) {
	// the target carrying MORE than the source is not a gap this view
	// judges (the publish leg's own story)
	if got := CellState(true, ip(5), ip(9), 0); got != StateMatch {
		t.Fatalf("target ahead = %q, want match", got)
	}
}

func TestCellStateUnreviewedGapIsBehind(t *testing.T) {
	if got := CellState(true, ip(10), ip(5), 0); got != StateBehind {
		t.Fatalf("unreviewed gap = %q, want behind", got)
	}
}

func TestCellStateFullyDeferredGapIsReviewed(t *testing.T) {
	if got := CellState(true, ip(10), ip(5), 5); got != StateReviewed {
		t.Fatalf("fully deferred = %q, want reviewed", got)
	}
}

func TestCellStatePartiallyDeferredGapIsBehind(t *testing.T) {
	if got := CellState(true, ip(10), ip(5), 4); got != StateBehind {
		t.Fatalf("partially deferred = %q, want behind", got)
	}
}

func TestCellStateDeferredBeyondGapIsReviewedNotWeird(t *testing.T) {
	// stale deferrals can exceed the current gap — Reviewed is the
	// honest ceiling (the log keeps its rows; the gap only shrinks)
	if got := CellState(true, ip(10), ip(9), 5); got != StateReviewed {
		t.Fatalf("deferred beyond gap = %q, want reviewed", got)
	}
}

// ---------------------------------------------------------------------------
// 2. stage_cell — the item form: the one object the view paints
// ---------------------------------------------------------------------------

func TestStageCellUnknownWhenNotSetUp(t *testing.T) {
	c := StageCell(false, []string{"a"}, []string{}, nil)
	if c.State != StateUnknown || c.SourceCount != nil || c.TargetCount != nil ||
		c.GapCount != 0 || len(c.UnexplainedIDs) != 0 || len(c.DeferredGapIDs) != 0 {
		t.Fatalf("not set up cell = %+v", c)
	}
}

func TestStageCellUnknownWhenASideIsUnmeasured(t *testing.T) {
	if got := StageCell(true, nil, []string{}, nil).State; got != StateUnknown {
		t.Fatalf("nil source = %q", got)
	}
	if got := StageCell(true, []string{"a"}, nil, nil).State; got != StateUnknown {
		t.Fatalf("nil target = %q", got)
	}
}

func TestStageCellMatchAndCounts(t *testing.T) {
	c := StageCell(true, []string{"a", "b"}, []string{"a", "b"}, nil)
	if c.State != StateMatch || *c.SourceCount != 2 || *c.TargetCount != 2 ||
		c.GapCount != 0 {
		t.Fatalf("match cell = %+v", c)
	}
}

func TestStageCellBehindWithTheSplit(t *testing.T) {
	c := StageCell(true, []string{"a", "b", "c"}, []string{"a"},
		[]string{"b"})
	if c.State != StateBehind || c.GapCount != 2 ||
		!reflect.DeepEqual(c.UnexplainedIDs, []string{"c"}) ||
		!reflect.DeepEqual(c.DeferredGapIDs, []string{"b"}) ||
		c.UnexplainedCount != 1 || c.DeferredGapCount != 1 {
		t.Fatalf("behind split = %+v", c)
	}
}

func TestStageCellReviewedWhenEveryGapDeferred(t *testing.T) {
	c := StageCell(true, []string{"a", "b"}, []string{},
		[]string{"a", "b"})
	if c.State != StateReviewed || c.DeferredGapCount != 2 || c.UnexplainedCount != 0 {
		t.Fatalf("reviewed cell = %+v", c)
	}
}

func TestStageCellSourceOrderPreservedAndDeduped(t *testing.T) {
	c := StageCell(true, []string{"c", "a", "c", "b"}, []string{"b"}, nil)
	if !reflect.DeepEqual(c.UnexplainedIDs, []string{"c", "a"}) {
		t.Fatalf("first-seen order + dedupe = %v", c.UnexplainedIDs)
	}
}

func TestStageCellTargetOnlyItemsNeverCount(t *testing.T) {
	c := StageCell(true, []string{"a"}, []string{"a", "z"}, nil)
	if c.State != StateMatch || c.GapCount != 0 {
		t.Fatalf("target-only items must never count: %+v", c)
	}
}

func TestSplitGapByIdentity(t *testing.T) {
	un, df := SplitGap([]string{"a", "b", "c"}, []string{"b", "z"})
	if !reflect.DeepEqual(un, []string{"a", "c"}) ||
		!reflect.DeepEqual(df, []string{"b"}) {
		t.Fatalf("split = %v / %v", un, df)
	}
}

func TestSplitGapEmptyDeferred(t *testing.T) {
	un, df := SplitGap([]string{"a"}, nil)
	if len(un) != 1 || len(df) != 0 {
		t.Fatalf("empty deferred = %v / %v", un, df)
	}
}

// ---------------------------------------------------------------------------
// 2a. banner_verdict — worst-first precedence
// ---------------------------------------------------------------------------

func TestBannerBehindIsTheWorst(t *testing.T) {
	v := BannerVerdict([]*Cell{
		{State: StateBehind, UnexplainedCount: 5},
		{State: StateUnknown},
	})
	if v.Tone != "warning" || v.Text != "5 items waiting" {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerSingularWaiting(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateBehind, UnexplainedCount: 1}})
	if v.Text != "1 item waiting" {
		t.Fatalf("text = %q", v.Text)
	}
}

func TestBannerUnknownWhenNoBehind(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateMatch}, {State: StateUnknown},
		{State: StateUnknown}})
	if v.Tone != "muted" || v.Text != "2 stages not set up" {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerSingularStage(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateUnknown}})
	if v.Text != "1 stage not set up" {
		t.Fatalf("text = %q", v.Text)
	}
}

func TestBannerHealthyWithDeferred(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateMatch},
		{State: StateReviewed, DeferredGapCount: 37}})
	if v.Tone != "success" || v.Text != "In sync · 37 deferred" {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerHealthyClean(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateMatch}})
	if v.Tone != "success" || v.Text != "In sync" {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerDeferredSumsAcrossCells(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateReviewed, DeferredGapCount: 3},
		{State: StateMatch}, {State: StateReviewed, DeferredGapCount: 4}})
	if v.Text != "In sync · 7 deferred" {
		t.Fatalf("text = %q", v.Text)
	}
}

func TestBannerConflictsBehindOverUnknown(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateUnknown},
		{State: StateBehind, UnexplainedCount: 2}})
	if v.Tone != "warning" {
		t.Fatalf("tone = %q", v.Tone)
	}
}

func TestBannerConflictsUnknownOverHealthy(t *testing.T) {
	v := BannerVerdict([]*Cell{{State: StateUnknown},
		{State: StateReviewed, DeferredGapCount: 9}})
	if v.Tone != "muted" || !strings.Contains(v.Text, "not set up") {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerEmptyMatrixIsClean(t *testing.T) {
	if v := BannerVerdict(nil); v.Tone != "success" || v.Text != "In sync" {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerHostileCellsNeverCrash(t *testing.T) {
	v := BannerVerdict([]*Cell{nil, {}, {State: "bogus"}, {State: StateBehind}})
	if v.Tone != "warning" {
		t.Fatalf("v = %+v", v)
	}
}

// ---------------------------------------------------------------------------
// 2b. v0.91.0 — THE NAMED VERDICT (the banner names its stages)
// ---------------------------------------------------------------------------

func TestBannerTheOwnersOwnTrap(t *testing.T) {
	// no Worker URL: no LLM, scan or fetch ever sets it — the banner
	// must NAME the stage and its reason, not count nameless mysteries.
	v := BannerVerdict([]*Cell{
		{State: StateMatch, Stream: "Repos", Stage: "Vault → Backup"},
		{State: StateUnknown, Stream: "Websites", Stage: "Vault → Website",
			Reason: "no Worker URL set", Needs: "setup"},
	})
	if v.Tone != "muted" ||
		!strings.Contains(v.Text, "1 stage not set up") ||
		!strings.Contains(v.Text, "Websites · Vault → Website") ||
		!strings.Contains(v.Text, "no Worker URL set") {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerTheSetupMeasureSplit(t *testing.T) {
	v := BannerVerdict([]*Cell{
		{State: StateUnknown, Stream: "Repos", Stage: "Queue → Vault",
			Reason: "bot queue not checked this session", Needs: "measure"},
		{State: StateUnknown, Stream: "Websites", Stage: "Queue → Vault",
			Reason: "bot queue not checked this session", Needs: "measure"},
		{State: StateUnknown, Stream: "Websites", Stage: "Vault → Website",
			Reason: "no Worker URL set", Needs: "setup"},
	})
	if v.Tone != "muted" ||
		!strings.Contains(v.Text, "1 stage not set up") ||
		!strings.Contains(v.Text, "2 stages not measured") ||
		!strings.Contains(v.Text, "click the cell to run its check") {
		t.Fatalf("v = %+v", v)
	}
}

func TestBannerBehindNamesItsStagesToo(t *testing.T) {
	v := BannerVerdict([]*Cell{
		{State: StateBehind, Stream: "Websites", Stage: "Vault → Website",
			UnexplainedCount: 10},
	})
	want := "10 items waiting — Websites · Vault → Website: 10 waiting"
	if v.Text != want {
		t.Fatalf("text = %q, want %q", v.Text, want)
	}
}

func TestBannerAReasonlessNamedCellStillSpeaks(t *testing.T) {
	v := BannerVerdict([]*Cell{
		{State: StateUnknown, Stream: "Repos", Stage: "Vault → Backup"},
	})
	if !strings.Contains(v.Text, "not set up") ||
		!strings.Contains(v.Text, "Repos · Vault → Backup") {
		t.Fatalf("v = %+v", v)
	}
}

// ---------------------------------------------------------------------------
// 2c. cell_caption — THE MICRO-COPY LAW
// ---------------------------------------------------------------------------

func TestCaptionTheOwnersOwnExample(t *testing.T) {
	stage := &StageRecord{Key: "websites:vault_to_site", State: StateBehind,
		SourceCount: ip(720), TargetCount: ip(710), DisplayNumber: ip(710),
		GapCount: 10, UnexplainedCount: 10}
	cap := CellCaption(stage)
	if cap.Num != "710" ||
		!strings.Contains(cap.Cap, "10 to publish") ||
		!strings.Contains(cap.Cap, "720 in the vault") ||
		!strings.Contains(cap.Cap, "710 on the site") {
		t.Fatalf("cap = %+v", cap)
	}
}

func TestCaptionMatchSpeaksBothSidesWithEquals(t *testing.T) {
	stage := &StageRecord{Key: "websites:vault_to_site", State: StateMatch,
		SourceCount: ip(720), TargetCount: ip(720), DisplayNumber: ip(720)}
	cap := CellCaption(stage)
	if cap.Num != "720" || cap.Cap != "720 in the vault = 720 on the site" {
		t.Fatalf("cap = %+v", cap)
	}
}

func TestCaptionBackupSpeaksTheSealedPair(t *testing.T) {
	stage := &StageRecord{Key: "repos:vault_to_backup", State: StateMatch,
		SourceCount: ip(0), TargetCount: ip(0), DisplayNumber: ip(678)}
	cap := CellCaption(stage)
	if cap.Num != "678" || cap.Cap != "678 in the vault = 678 in the mirror" {
		t.Fatalf("cap = %+v", cap)
	}
	stageB := &StageRecord{Key: "repos:vault_to_backup", State: StateBehind,
		SourceCount: ip(0), TargetCount: ip(0), DisplayNumber: ip(674),
		GapCount: 4, UnexplainedCount: 4}
	capB := CellCaption(stageB)
	if capB.Num != "674" ||
		!strings.Contains(capB.Cap, "4 unsealed") ||
		!strings.Contains(capB.Cap, "678 in the vault") ||
		!strings.Contains(capB.Cap, "674 sealed in the mirror") {
		t.Fatalf("capB = %+v", capB)
	}
}

func TestCaptionQueueSpeaksWaitingAndInVault(t *testing.T) {
	stage := &StageRecord{Key: "repos:queue_to_vault", State: StateBehind,
		SourceCount: ip(3), TargetCount: ip(0), DisplayNumber: ip(631),
		GapCount: 3, UnexplainedCount: 3}
	cap := CellCaption(stage)
	if cap.Num != "631" ||
		!strings.Contains(cap.Cap, "3 waiting in the queue") ||
		!strings.Contains(cap.Cap, "631 already in the vault") {
		t.Fatalf("cap = %+v", cap)
	}
	stageR := &StageRecord{Key: "repos:queue_to_vault", State: StateReviewed,
		SourceCount: ip(3), TargetCount: ip(0), DisplayNumber: ip(631),
		GapCount: 3, DeferredGapCount: 3}
	if capR := CellCaption(stageR); !strings.Contains(capR.Cap, "3 deferred (explained)") {
		t.Fatalf("capR = %+v", capR)
	}
}

func TestCaptionUnknownWritesItsReasonOnItself(t *testing.T) {
	stage := &StageRecord{Key: "websites:vault_to_site",
		State: StateUnknown, Reason: "no Worker URL set"}
	cap := CellCaption(stage)
	if cap.Num != "—" || cap.Cap != "no Worker URL set" {
		t.Fatalf("cap = %+v", cap)
	}
	stage2 := &StageRecord{Key: "websites:vault_to_site", State: StateUnknown}
	if cap2 := CellCaption(stage2); cap2.Cap != "not set up or not measured" {
		t.Fatalf("cap2 = %+v", cap2)
	}
}

func TestCaptionJunkNeverRaises(t *testing.T) {
	if cap := CellCaption(nil); cap.Num != "—" {
		t.Fatalf("nil caption = %+v", cap)
	}
	if cap := CellCaption(&StageRecord{State: StateMatch}); cap.Num != "—" {
		t.Fatalf("displayless match caption = %+v", cap)
	}
}

// ---------------------------------------------------------------------------
// 3. the deferred ledger's lifecycle
// ---------------------------------------------------------------------------

func tmpLog(t *testing.T) *DeferredLog {
	t.Helper()
	dir := t.TempDir()
	log, err := OpenLedger(filepath.Join(dir, "sync_review.db"))
	if err != nil {
		t.Fatalf("open ledger: %v", err)
	}
	t.Cleanup(func() { _ = log.Close() })
	return log
}

func TestLedgerMarkMakesActiveRows(t *testing.T) {
	log := tmpLog(t)
	if n := log.Mark("s", []string{"a", "b"}, "private", "", nil, ""); n != 2 {
		t.Fatalf("marked %d", n)
	}
	rows := log.Rows("s", "")
	if len(rows) != 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	for _, r := range rows {
		if r.Status != StatusActive || r.Reason != "private" {
			t.Fatalf("row = %+v", r)
		}
	}
}

func TestLedgerMarkStoresDisplayNames(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"https://a.com"}, "other", "n",
		map[string]string{"https://a.com": "A Site"}, "")
	rows := log.Rows("s", "")
	if len(rows) != 1 || rows[0].DisplayName != "A Site" ||
		rows[0].Note != "n" {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestLedgerRemarkUpdatesInPlaceNeverGrows(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "private", "", nil, "")
	log.Mark("s", []string{"a"}, "duplicate", "second", nil, "")
	if got := log.RowCount("s"); got != 1 {
		t.Fatalf("row count = %d, want 1 (a re-mark updates in place)", got)
	}
	rows := log.Rows("s", "")
	if rows[0].Reason != "duplicate" || rows[0].Note != "second" {
		t.Fatalf("row = %+v", rows)
	}
}

func TestLedgerInvalidReasonIsRefused(t *testing.T) {
	log := tmpLog(t)
	if n := log.Mark("s", []string{"a"}, "bogus", "", nil, ""); n != 0 {
		t.Fatalf("marked %d with a bogus reason", n)
	}
	if n := log.Mark("s", []string{"a"}, "", "", nil, ""); n != 0 {
		t.Fatalf("marked %d with no reason", n)
	}
	if got := log.RowCount("s"); got != 0 {
		t.Fatalf("row count = %d", got)
	}
}

func TestLedgerEmptyItemsMarkNothing(t *testing.T) {
	log := tmpLog(t)
	if n := log.Mark("s", nil, "other", "", nil, ""); n != 0 {
		t.Fatalf("marked %d of no items", n)
	}
	if n := log.Mark("s", []string{}, "other", "", nil, ""); n != 0 {
		t.Fatalf("marked %d of an empty list", n)
	}
	if n := log.Mark("s", []string{""}, "other", "", nil, ""); n != 0 {
		t.Fatalf("marked %d of one empty-string item", n)
	}
}

func TestLedgerBulkMarkOneSharedReason(t *testing.T) {
	log := tmpLog(t)
	ids := []string{"a", "b", "c"}
	if n := log.Mark("s", ids, "not_ready", "", nil, ""); n != 3 {
		t.Fatalf("marked %d", n)
	}
	for _, r := range log.Rows("s", "") {
		if r.Reason != "not_ready" {
			t.Fatalf("row = %+v", r)
		}
	}
}

func TestLedgerStagesAreIndependent(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s1", []string{"a"}, "other", "", nil, "")
	log.Mark("s2", []string{"a"}, "private", "", nil, "")
	if !reflect.DeepEqual(log.ReviewedIDs("s1"), map[string]bool{"a": true}) {
		t.Fatalf("s1 reviewed = %v", log.ReviewedIDs("s1"))
	}
	if !reflect.DeepEqual(log.ReviewedIDs("s2"), map[string]bool{"a": true}) {
		t.Fatalf("s2 reviewed = %v", log.ReviewedIDs("s2"))
	}
}

func TestLedgerReopenActiveReturnsToDifferences(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "")
	if !log.Reopen("s", "a", "") {
		t.Fatal("reopen of an active row must succeed")
	}
	if len(log.ReviewedIDs("s")) != 0 {
		t.Fatalf("reviewed after reopen = %v", log.ReviewedIDs("s"))
	}
}

func TestLedgerReopenIsLoggedNotDeleted(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "")
	log.Reopen("s", "a", "")
	rows := log.Rows("s", "")
	if len(rows) != 1 || rows[0].Status != StatusReopened || rows[0].ReopenCount != 1 {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestLedgerOnlyActiveRowsReopen(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "")
	log.Reopen("s", "a", "")
	if log.Reopen("s", "a", "") {
		t.Fatal("a reopened row must not reopen again")
	}
	if log.Reopen("s", "never-marked", "") {
		t.Fatal("an absent row must not reopen")
	}
}

func TestLedgerReopenCountsRepeats(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "")
	log.Reopen("s", "a", "")
	log.Mark("s", []string{"a"}, "private", "", nil, "")
	log.Reopen("s", "a", "")
	rows := log.Rows("s", "")
	if len(rows) != 1 || rows[0].ReopenCount != 2 {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestLedgerResolveSeenWhenItemAppears(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a", "b"}, "other", "", nil, "")
	if n := log.ResolveSeen("s", []string{"a"}, ""); n != 1 {
		t.Fatalf("resolved %d", n)
	}
	rows := log.Rows("s", "")
	byID := map[string]LedgerRow{}
	for _, r := range rows {
		byID[r.ItemID] = r
	}
	if byID["a"].Status != StatusResolved || byID["b"].Status != StatusActive {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestLedgerResolvedRowFlipsBackWhenGapReturns(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "")
	log.ResolveSeen("s", []string{"a"}, "")
	log.ResolveSeen("s", nil, "") // the item left the target again
	rows := log.Rows("s", "")
	if rows[0].Status != StatusActive {
		t.Fatalf("status = %q, want active (the deferral stands)", rows[0].Status)
	}
}

func TestLedgerResolveGoneQueueSemantics(t *testing.T) {
	log := tmpLog(t)
	log.Mark("q", []string{"a", "b"}, "other", "", nil, "")
	// 'a' left the gap (processed) → resolved; 'b' still pending
	if n := log.ResolveGone("q", []string{"b"}, ""); n != 1 {
		t.Fatalf("resolved %d", n)
	}
	rows := log.Rows("q", "")
	byID := map[string]LedgerRow{}
	for _, r := range rows {
		byID[r.ItemID] = r
	}
	if byID["a"].Status != StatusResolved || byID["b"].Status != StatusActive {
		t.Fatalf("rows = %+v", rows)
	}
}

func TestLedgerResolveGoneFlipBackOnReappearance(t *testing.T) {
	log := tmpLog(t)
	log.Mark("q", []string{"a"}, "other", "", nil, "")
	log.ResolveGone("q", nil, "")           // processed → resolved
	log.ResolveGone("q", []string{"a"}, "") // back in the gap
	if rows := log.Rows("q", ""); rows[0].Status != StatusActive {
		t.Fatalf("status = %q", rows[0].Status)
	}
}

func TestLedgerResolvedNeverLeavesReviewed(t *testing.T) {
	// reviewed_ids = every item whose deferral STANDS (active or
	// resolved — anything but reopened)
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "")
	log.ResolveSeen("s", []string{"a"}, "")
	if !log.ReviewedIDs("s")["a"] {
		t.Fatal("a resolved row still counts as reviewed")
	}
}

func TestLedgerFullLifecycleNeverShrinks(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a", "b"}, "other", "", nil, "")
	log.Reopen("s", "a", "")
	log.ResolveSeen("s", []string{"b"}, "")
	log.Mark("s", []string{"a"}, "private", "", nil, "")
	log.ResolveGone("s", []string{"a"}, "")
	if got := log.RowCount(""); got != 2 {
		t.Fatalf("row count = %d, want 2 (nothing is ever deleted)", got)
	}
}

func TestLedgerNoDeleteVerbInTheCore(t *testing.T) {
	// the never-delete law, at the source level
	src, err := os.ReadFile("ledger.go")
	if err != nil {
		t.Fatalf("read ledger.go: %v", err)
	}
	for _, banned := range []string{"DELETE FROM", "DROP TABLE", "TRUNCATE"} {
		if strings.Contains(string(src), banned) {
			t.Fatalf("the ledger must never carry %q", banned)
		}
	}
}

func TestLedgerIdsSurviveReSorts(t *testing.T) {
	log := tmpLog(t)
	ids := []string{"https://site-0.com", "https://site-1.com",
		"https://site-2.com", "https://site-3.com", "https://site-4.com"}
	log.Mark("s", ids[:2], "private", "", nil, "")
	reviewed := log.ReviewedIDs("s")
	shuffled := []string{ids[4], ids[3], ids[2], ids[1], ids[0]}
	c := StageCell(true, shuffled, []string{}, keysOf(reviewed))
	if c.DeferredGapCount != 2 || c.UnexplainedCount != 3 {
		t.Fatalf("cell = %+v (a re-sort must keep every deferral)", c)
	}
	// and by NOTE PATH too
	paths := []string{"AI/Skills/One.md", "AI/Skills/Two.md"}
	log.Mark("b", []string{paths[0]}, "not_ready", "", nil, "")
	c2 := StageCell(true, []string{paths[1], paths[0]}, []string{},
		keysOf(log.ReviewedIDs("b")))
	if !reflect.DeepEqual(c2.DeferredGapIDs, []string{"AI/Skills/One.md"}) {
		t.Fatalf("c2 = %+v", c2)
	}
}

func keysOf(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func TestLedgerSearchFiltersItemReasonNote(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"https://alpha.com"}, "private", "secret", nil, "")
	log.Mark("s", []string{"https://beta.com"}, "duplicate", "dupe", nil, "")
	for q, want := range map[string]int{
		"alpha": 1, "beta": 1, "private": 1, "secret": 1,
		"nothing-matches": 0, "Private": 1,
	} {
		if got := len(log.Rows("s", q)); got != want {
			t.Fatalf("rows(%q) = %d, want %d", q, got, want)
		}
	}
}

func TestLedgerRowsNewestReviewedFirst(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s", []string{"a"}, "other", "", nil, "2026-10-14T10:00:00+00:00")
	log.Mark("s", []string{"b"}, "other", "", nil, "2026-10-14T12:00:00+00:00")
	got := log.Rows("s", "")
	if len(got) != 2 || got[0].ItemID != "b" || got[1].ItemID != "a" {
		t.Fatalf("rows = %+v (newest-reviewed first)", got)
	}
}

func TestLedgerRowCountAllVsStage(t *testing.T) {
	log := tmpLog(t)
	log.Mark("s1", []string{"a"}, "other", "", nil, "")
	log.Mark("s2", []string{"b"}, "other", "", nil, "")
	if log.RowCount("") != 2 || log.RowCount("s1") != 1 {
		t.Fatalf("counts = %d / %d", log.RowCount(""), log.RowCount("s1"))
	}
}

func TestLedgerCloseIsIdempotent(t *testing.T) {
	dir := t.TempDir()
	log, err := OpenLedger(filepath.Join(dir, "x.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatal(err)
	}
	if err := log.Close(); err != nil {
		t.Fatalf("second close: %v", err)
	}
}

func TestLedgerTwoWritersMarkTheSameDB(t *testing.T) {
	// THE SQLITE LAW: busy_timeout + WAL + the per-ledger mutex — a
	// second writer on the same file gets a clean wait, never a crash.
	dir := t.TempDir()
	path := filepath.Join(dir, "shared.db")
	main, err := OpenLedger(path)
	if err != nil {
		t.Fatal(err)
	}
	defer main.Close()
	var wg sync.WaitGroup
	errs := make(chan error, 4)
	for n := 0; n < 4; n++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			other, err := OpenLedger(path)
			if err != nil {
				errs <- err
				return
			}
			for i := 0; i < 20; i++ {
				if m := other.Mark("s",
					[]string{"t" + itoa(n) + "-" + itoa(i)}, "other",
					"", nil, ""); m != 1 {
					errs <- os.ErrInvalid
					_ = other.Close()
					return
				}
			}
			_ = other.Close()
		}(n)
	}
	wg.Wait()
	close(errs)
	for err := range errs {
		t.Fatalf("concurrent writer: %v", err)
	}
	if got := main.RowCount("s"); got != 80 {
		t.Fatalf("row count = %d, want 80", got)
	}
}

func TestLedgerLockingLawInSource(t *testing.T) {
	// the busy_timeout + WAL + mutex law, at the source level (the
	// Python test scans for timeout=30 / PRAGMA busy_timeout / RLock)
	src, err := os.ReadFile("ledger.go")
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"sync.Mutex", "store.Open"} {
		if !strings.Contains(string(src), want) {
			t.Fatalf("ledger.go must carry %q", want)
		}
	}
	dsn := store.DSN(LedgerFilename)
	if !strings.Contains(dsn, "busy_timeout(30000)") ||
		!strings.Contains(dsn, "journal_mode(WAL)") ||
		!strings.Contains(dsn, "synchronous(NORMAL)") {
		t.Fatalf("the DSN must carry the pragma law: %q", dsn)
	}
}

// ---------------------------------------------------------------------------
// 4. the chip grammar + the disabled-until law
// ---------------------------------------------------------------------------

func TestChipGrammarExactlyFourReasons(t *testing.T) {
	if len(Reasons) != 4 {
		t.Fatalf("reasons = %v", Reasons)
	}
	want := []string{"not_ready", "private", "duplicate", "other"}
	if !reflect.DeepEqual(Reasons, want) {
		t.Fatalf("reasons = %v", Reasons)
	}
	wantLabels := map[string]string{
		"not_ready": "Not ready", "private": "Private",
		"duplicate": "Duplicate", "other": "Other",
	}
	if !reflect.DeepEqual(ReasonLabels, wantLabels) {
		t.Fatalf("labels = %v", ReasonLabels)
	}
}

func TestChipGrammarValidReason(t *testing.T) {
	for _, r := range Reasons {
		if !ValidReason(r) {
			t.Fatalf("%q must be valid", r)
		}
	}
	for _, r := range []string{"", "bogus", "PRIVATE"} {
		if ValidReason(r) {
			t.Fatalf("%q must be invalid", r)
		}
	}
}

func TestChipGrammarDisabledUntilChipAndItem(t *testing.T) {
	if MarkAllowed(0, "private") {
		t.Fatal("no items checked must stay disabled")
	}
	if MarkAllowed(1, "") {
		t.Fatal("no chip chosen must stay disabled")
	}
	if !MarkAllowed(1, "private") {
		t.Fatal("≥1 item AND a chip must arm the button")
	}
	if !MarkAllowed(3, "other") {
		t.Fatal("many items + a chip must arm the button")
	}
}

func TestChipGrammarInvalidChipNeverArms(t *testing.T) {
	if MarkAllowed(2, "bogus") || MarkAllowed(2, "  ") {
		t.Fatal("an invalid chip must never arm the button")
	}
}

// ---------------------------------------------------------------------------
// 5. the relative clock
// ---------------------------------------------------------------------------

func agoNow(t *testing.T) time.Time {
	t.Helper()
	now, err := time.Parse(time.RFC3339, "2026-10-14T12:00:00Z")
	if err != nil {
		t.Fatal(err)
	}
	return now
}

func TestFormatAddedAgoDays(t *testing.T) {
	now := agoNow(t)
	for days, want := range map[int]string{
		1: "1 d ago", 3: "3 d ago", 29: "29 d ago",
	} {
		stamp := now.AddDate(0, 0, -days).Format("2006-01-02")
		if got := FormatAddedAgo(stamp, now); got != want {
			t.Fatalf("days=%d: %q, want %q", days, got, want)
		}
	}
}

func TestFormatAddedAgoHoursAndMinutes(t *testing.T) {
	now := agoNow(t)
	cases := []struct {
		delta time.Duration
		want  string
	}{
		{-30 * time.Second, "just now"},
		{-5 * time.Minute, "5 m ago"},
		{-59 * time.Minute, "59 m ago"},
		{-3 * time.Hour, "3 h ago"},
		{-23 * time.Hour, "23 h ago"},
	}
	for _, tc := range cases {
		stamp := now.Add(tc.delta).UTC().Format(time.RFC3339)
		if got := FormatAddedAgo(stamp, now); got != tc.want {
			t.Fatalf("delta=%v: %q, want %q", tc.delta, got, tc.want)
		}
	}
}

func TestFormatAddedAgoJustNow(t *testing.T) {
	now := agoNow(t)
	if got := FormatAddedAgo(now.UTC().Format(time.RFC3339), now); got != "just now" {
		t.Fatalf("got %q", got)
	}
}

func TestFormatAddedAgoMonthsAndYears(t *testing.T) {
	now := agoNow(t)
	cases := []struct {
		stamp string
		want  string
	}{
		{"2026-08-01", "2 mo ago"}, // the reference's own pair
		{"2024-10-14", "2 y ago"},  // 730 d = 24 mo = 2 y
	}
	for _, tc := range cases {
		if got := FormatAddedAgo(tc.stamp, now); got != tc.want {
			t.Fatalf("%s: %q, want %q", tc.stamp, got, tc.want)
		}
	}
}

func TestFormatAddedAgoMissingIsHonestDash(t *testing.T) {
	now := agoNow(t)
	for _, stamp := range []string{"", "   ", "not-a-date"} {
		if got := FormatAddedAgo(stamp, now); got != "—" {
			t.Fatalf("%q: %q, want the honest dash", stamp, got)
		}
	}
}

func TestFormatAddedAgoGarbageIsHonestDash(t *testing.T) {
	if got := FormatAddedAgo("2026-13-45", agoNow(t)); got != "—" {
		t.Fatalf("an invalid calendar date must answer —, got %q", got)
	}
}

func TestFormatAddedAgoClockSkewNeverFuture(t *testing.T) {
	now := agoNow(t)
	future := now.Add(5 * time.Minute).UTC().Format(time.RFC3339)
	if got := FormatAddedAgo(future, now); got != "just now" {
		t.Fatalf("clock skew: %q, want just now (never the future)", got)
	}
}

func TestFormatAddedAgoNaiveNowIsTolerated(t *testing.T) {
	// a zero now means the real clock — the stamp rides 3 d back, the
	// answer is the same relative verdict (never a crash, never a
	// future word)
	stamp := time.Now().UTC().AddDate(0, 0, -3).Format("2006-01-02")
	if got := FormatAddedAgo(stamp, time.Time{}); got != "3 d ago" {
		t.Fatalf("naive now: %q, want 3 d ago", got)
	}
}

func TestAddedDaysAgoAndSortOldestFirst(t *testing.T) {
	now := agoNow(t)
	d := func(days int) *int { return AddedDaysAgo(now.AddDate(0, 0, -days).Format("2006-01-02"), now) }
	if *d(3) != 3 {
		t.Fatalf("added_days_ago = %d", *d(3))
	}
	if AddedDaysAagoMissing(now) != nil {
		t.Fatal("a missing date must answer nil")
	}
	items := []GapItem{
		{ID: "undated"}, {ID: "old", AddedDaysAgo: d(9)},
		{ID: "newest", AddedDaysAgo: d(1)}, {ID: "mid", AddedDaysAgo: d(9)},
	}
	sorted := SortOldestFirst(items)
	want := []string{"old", "mid", "newest", "undated"}
	got := make([]string, 4)
	for i, it := range sorted {
		got[i] = it.ID
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("sort = %v, want %v (oldest first, undated last, stable)", got, want)
	}
}

func AddedDaysAagoMissing(now time.Time) *int {
	return AddedDaysAgo("", now)
}

// ---------------------------------------------------------------------------
// 6. the vault-side walk (fixture vault, the build_index filter laws)
// ---------------------------------------------------------------------------

func writeSiteNote(t *testing.T, vault, rel, source, name, added string) {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := "---\nsource: " + source + "\n"
	if name != "" {
		body += "name: " + name + "\n"
	}
	if added != "" {
		body += "date_processed: " + added + "\n"
	}
	body += "---\n\n# " + name + "\n"
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
}

func TestVaultSiteItemsCarryUrlNameDateNotePath(t *testing.T) {
	vault := t.TempDir()
	writeSiteNote(t, vault, "AI/Tools/a.md", "https://a.com", "A Site",
		"2026-10-01")
	items, ok := VaultSiteItems(vault, nil, 0)
	if !ok || len(items) != 1 {
		t.Fatalf("items = %+v ok=%v", items, ok)
	}
	it := items[0]
	if it.ID != "https://a.com" || it.Name != "A Site" ||
		it.AddedAt != "2026-10-01" || it.NotePath == "" ||
		!strings.HasSuffix(filepath.ToSlash(it.NotePath), "AI/Tools/a.md") {
		t.Fatalf("item = %+v", it)
	}
}

func TestVaultSiteItemsNoSourceNoteIsNotASite(t *testing.T) {
	vault := t.TempDir()
	// a plain note without the app's frontmatter shape
	if err := os.WriteFile(filepath.Join(vault, "plain.md"),
		[]byte("# no frontmatter\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items, ok := VaultSiteItems(vault, nil, 0)
	if !ok || len(items) != 0 {
		t.Fatalf("items = %+v ok=%v", items, ok)
	}
}

func TestVaultSiteItemsPrivateCategoryFiltered(t *testing.T) {
	vault := t.TempDir()
	// the private default name (belt's DefaultPrivateNames)
	writeSiteNote(t, vault, "Torrent & Shadow Libraries/x.md",
		"https://secret.com", "S", "")
	writeSiteNote(t, vault, "AI/Tools/ok.md", "https://ok.com", "O", "")
	items, ok := VaultSiteItems(vault, nil, 0)
	if !ok || len(items) != 1 || items[0].ID != "https://ok.com" {
		t.Fatalf("items = %+v ok=%v", items, ok)
	}
}

func TestVaultSiteItemsDuplicateUrlsDedupe(t *testing.T) {
	vault := t.TempDir()
	writeSiteNote(t, vault, "AI/Tools/a.md", "https://a.com", "A", "")
	writeSiteNote(t, vault, "AI/Other/b.md", "https://a.com", "B", "")
	items, _ := VaultSiteItems(vault, nil, 0)
	if len(items) != 1 {
		t.Fatalf("items = %+v (dedupe by URL)", items)
	}
}

func TestVaultSiteItemsMissingVaultIsNotSetUp(t *testing.T) {
	if _, ok := VaultSiteItems("", nil, 0); ok {
		t.Fatal("an empty vault is not set up")
	}
	if _, ok := VaultSiteItems("/no/such/vault", nil, 0); ok {
		t.Fatal("a missing vault is not set up")
	}
}

func TestVaultSiteItemsCapIsHonored(t *testing.T) {
	vault := t.TempDir()
	for i := 0; i < 5; i++ {
		writeSiteNote(t, vault, "AI/Tools/n"+itoa(i)+".md",
			"https://n"+itoa(i)+".com", "N", "")
	}
	var lines []string
	items, _ := VaultSiteItems(vault, func(m string) { lines = append(lines, m) }, 3)
	if len(items) != 3 {
		t.Fatalf("items = %d, want the cap 3", len(items))
	}
	if len(lines) == 0 || !strings.Contains(lines[0], "capped at 3") {
		t.Fatalf("the cap must be spoken: %v", lines)
	}
}

// ---------------------------------------------------------------------------
// 7. the backup gap (fixture git repos — the read-only git status walk)
// ---------------------------------------------------------------------------

func backupVault(t *testing.T, n int) string {
	t.Helper()
	vault := filepath.Join(t.TempDir(), "V")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < n; i++ {
		body := "---\n---\n# Note " + itoa(i) + "\n"
		if err := os.WriteFile(filepath.Join(vault, "Note"+itoa(i)+".md"),
			[]byte(body), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return vault
}

func git(t *testing.T, vault string, args ...string) {
	t.Helper()
	cmd := exec.Command("git", append([]string{"-C", vault}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		t.Fatalf("git %v: %v\n%s", args, err, out)
	}
}

func gitSeal(t *testing.T, vault string) {
	t.Helper()
	git(t, vault, "init", "-q")
	git(t, vault, "add", "-A")
	git(t, vault, "-c", "user.email=t@t", "-c", "user.name=t",
		"commit", "-q", "-m", "seal")
}

func TestVaultBackupGapNoVaultIsNotSetUp(t *testing.T) {
	if _, ok := VaultBackupGap("", nil, 0); ok {
		t.Fatal("an empty vault is not set up")
	}
	if _, ok := VaultBackupGap("/no/such/vault", nil, 0); ok {
		t.Fatal("a missing vault is not set up")
	}
}

func TestVaultBackupGapNeverSealedEveryNoteIsPending(t *testing.T) {
	vault := backupVault(t, 3)
	items, ok := VaultBackupGap(vault, nil, 0)
	if !ok || len(items) != 3 {
		t.Fatalf("items = %+v ok=%v", items, ok)
	}
	for _, g := range items {
		if !strings.HasSuffix(g.ID, ".md") {
			t.Fatalf("id = %q", g.ID)
		}
	}
}

func TestVaultBackupGapSealedVaultHasNoGap(t *testing.T) {
	vault := backupVault(t, 2)
	gitSeal(t, vault)
	items, ok := VaultBackupGap(vault, nil, 0)
	if !ok || len(items) != 0 {
		t.Fatalf("items = %+v ok=%v", items, ok)
	}
}

func TestVaultBackupGapNewUntrackedNoteIsTheGap(t *testing.T) {
	vault := backupVault(t, 2)
	gitSeal(t, vault)
	if err := os.WriteFile(filepath.Join(vault, "New.md"),
		[]byte("# new\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	items, _ := VaultBackupGap(vault, nil, 0)
	if len(items) != 1 || items[0].ID != "New.md" {
		t.Fatalf("items = %+v", items)
	}
}

func TestVaultBackupGapModifiedNoteIsTheGap(t *testing.T) {
	vault := backupVault(t, 1)
	gitSeal(t, vault)
	f, err := os.OpenFile(filepath.Join(vault, "Note0.md"),
		os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		t.Fatal(err)
	}
	f.WriteString("edited\n")
	f.Close()
	items, _ := VaultBackupGap(vault, nil, 0)
	if len(items) != 1 || items[0].ID != "Note0.md" {
		t.Fatalf("items = %+v", items)
	}
}

func TestVaultBackupGapNonMdAndObsidianNoiseNeverCount(t *testing.T) {
	vault := backupVault(t, 1)
	git(t, vault, "init", "-q")
	if err := os.MkdirAll(filepath.Join(vault, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(vault, ".obsidian", "workspace.json"),
		[]byte("{}"), 0o644)
	_ = os.WriteFile(filepath.Join(vault, "config.json"), []byte("{}"), 0o644)
	items, _ := VaultBackupGap(vault, nil, 0)
	if len(items) != 1 || items[0].ID != "Note0.md" {
		t.Fatalf("items = %+v", items)
	}
}

func TestVaultBackupGapReadOnlyLaw(t *testing.T) {
	// the repo_radar law, at the source level: git status only — no
	// write verb anywhere in the walk
	src, err := os.ReadFile("vaultwalk.go")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(src), `"status", "--porcelain"`) {
		t.Fatal("the backup gap must ride git status --porcelain")
	}
	for _, banned := range []string{`"commit"`, `"push"`, `"add"`, `"reset"`,
		`"checkout"`} {
		if strings.Contains(string(src), banned) {
			t.Fatalf("the walk must never carry the write verb %s", banned)
		}
	}
}

// ---------------------------------------------------------------------------
// 8. the note count
// ---------------------------------------------------------------------------

func TestVaultNoteCountWalksTheInventoryLaw(t *testing.T) {
	vault := t.TempDir()
	writeSiteNote(t, vault, "AI/Tools/a.md", "https://a.com", "A", "")
	writeSiteNote(t, vault, "AI/Tools/b.md", "https://b.com", "B", "")
	_ = os.WriteFile(filepath.Join(vault, "plain.md"),
		[]byte("# plain\n"), 0o644) // counted by the inventory too
	n := VaultNoteCount(vault)
	if n == nil || *n != 3 {
		t.Fatalf("count = %v, want 3", n)
	}
}

func TestVaultNoteCountMissingVaultIsNil(t *testing.T) {
	if VaultNoteCount("") != nil {
		t.Fatal("an empty vault answers nil")
	}
	if VaultNoteCount("/no/such/vault") != nil {
		t.Fatal("a missing vault answers nil")
	}
}

// ---------------------------------------------------------------------------
// 9. the one network door (the hermetic seam — canned responses)
// ---------------------------------------------------------------------------

func fakeHTTP(status int, body string) func(*http.Request, float64) (*http.Response, error) {
	return func(req *http.Request, timeoutS float64) (*http.Response, error) {
		return &http.Response{
			StatusCode: status,
			Body:       io.NopCloser(bytes.NewReader([]byte(body))),
			Header:     http.Header{},
			Request:    req,
		}, nil
	}
}

func TestFetchPublishedIndexGetOnlyAndUALaw(t *testing.T) {
	orig := httpDo
	defer func() { httpDo = orig }()
	var seen *http.Request
	httpDo = func(req *http.Request, timeoutS float64) (*http.Response, error) {
		seen = req
		return fakeHTTP(200, `{"sites": []}`)(req, timeoutS)
	}
	if FetchPublishedIndex("https://w.example", 0) == nil {
		t.Fatal("a live index must answer")
	}
	if seen == nil {
		t.Fatal("the seam never ran")
	}
	if seen.Method != "GET" {
		t.Fatalf("method = %q, want GET (the net_guard law)", seen.Method)
	}
	if got := seen.Header.Get("User-Agent"); got != workerclient.HTTPUserAgent {
		t.Fatalf("UA = %q, want the 1010 law's own %q", got, workerclient.HTTPUserAgent)
	}
	if got := seen.URL.String(); got != "https://w.example/api/directory/index" {
		t.Fatalf("url = %q", got)
	}
}

func TestFetchPublishedIndexSuccessParsesSites(t *testing.T) {
	orig := httpDo
	defer func() { httpDo = orig }()
	httpDo = fakeHTTP(200,
		`{"sites": [{"url": "https://a.com"}, {"url": "https://b.com"}],
		  "counts": {"sites": 2}}`)
	idx := FetchPublishedIndex("https://w.example", 0)
	if idx == nil {
		t.Fatal("idx = nil")
	}
	urls, ok := PublishedSiteUrls(idx)
	if !ok || !reflect.DeepEqual(urls, []string{"https://a.com", "https://b.com"}) {
		t.Fatalf("urls = %v ok=%v", urls, ok)
	}
}

func TestFetchPublishedIndexAnyFailureAnswersNil(t *testing.T) {
	orig := httpDo
	defer func() { httpDo = orig }()
	httpDo = func(req *http.Request, timeoutS float64) (*http.Response, error) {
		return nil, os.ErrDeadlineExceeded
	}
	if FetchPublishedIndex("https://w.example", 0) != nil {
		t.Fatal("a network failure must answer nil")
	}
	httpDo = fakeHTTP(200, "not json")
	if FetchPublishedIndex("https://w.example", 0) != nil {
		t.Fatal("garbage JSON must answer nil")
	}
}

func TestFetchPublishedIndexBadShapeAnswersNil(t *testing.T) {
	orig := httpDo
	defer func() { httpDo = orig }()
	httpDo = fakeHTTP(200, `{"error": "not published"}`)
	if FetchPublishedIndex("https://w.example", 0) != nil {
		t.Fatal("a dict without a sites list must answer nil")
	}
	httpDo = fakeHTTP(200, `{"sites": {"not": "a list"}}`)
	if FetchPublishedIndex("https://w.example", 0) != nil {
		t.Fatal("sites-not-a-list must answer nil")
	}
}

func TestFetchPublishedIndexEmptyURLAnswersNil(t *testing.T) {
	if FetchPublishedIndex("", 0) != nil {
		t.Fatal("an empty URL answers nil")
	}
	if FetchPublishedIndex("not-a-url", 0) != nil {
		t.Fatal("a non-http URL answers nil")
	}
}

func TestPublishedUrlsNoneIndexIsNotMeasured(t *testing.T) {
	if _, ok := PublishedSiteUrls(nil); ok {
		t.Fatal("a nil index is unmeasured (ok=false)")
	}
}

func TestPublishedUrlsEmptyIndexIsTheHonestEmpty(t *testing.T) {
	idx := storage.NewOrderedMap()
	idx.Set("sites", []any{})
	urls, ok := PublishedSiteUrls(idx)
	if !ok || len(urls) != 0 {
		t.Fatalf("urls = %v ok=%v", urls, ok)
	}
}

func TestPublishedUrlsNeverRaisesOnHostilePayloads(t *testing.T) {
	// hostile payloads never crash the door — decoded through the
	// storage decoder the production door itself rides
	idx, err := storage.DecodeJSONValue([]byte(
		`{"sites": [null, 5, {"url": null}, {"url": "https://ok.com"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	om, _ := idx.(*storage.OrderedMap)
	urls, ok := PublishedSiteUrls(om)
	if !ok || !reflect.DeepEqual(urls, []string{"https://ok.com"}) {
		t.Fatalf("urls = %v ok=%v", urls, ok)
	}
	// a plain map shape is tolerated too (junk-tolerance, never a crash)
	plain := storage.NewOrderedMap()
	plain.Set("sites", []any{map[string]any{"url": "https://plain.com"}})
	urls2, ok2 := PublishedSiteUrls(plain)
	if !ok2 || !reflect.DeepEqual(urls2, []string{"https://plain.com"}) {
		t.Fatalf("urls2 = %v ok2=%v", urls2, ok2)
	}
}

// ---------------------------------------------------------------------------
// 10. the default home + the marshal shape (the snapshot's own laws)
// ---------------------------------------------------------------------------

func TestLedgersDefaultHomeIsTheAppDataFolder(t *testing.T) {
	p := LedgerDefaultPath("")
	if !strings.HasSuffix(filepath.ToSlash(p), "/sync_review.db") {
		t.Fatalf("path = %q", p)
	}
	if filepath.Base(p) != LedgerFilename {
		t.Fatalf("base = %q", filepath.Base(p))
	}
	// an explicit app_dir wins
	if got := LedgerDefaultPath("/x"); got != "/x"+string(filepath.Separator)+LedgerFilename {
		t.Fatalf("explicit dir = %q", got)
	}
}

func TestCellMarshalKeepsThePaintShape(t *testing.T) {
	// the JSON the shell transports must carry the paint's own keys
	// (structs emit DECLARATION order — the golden-marshal-order trap's
	// own law, asserted once here for the Sync shapes)
	c := StageCell(true, []string{"a"}, []string{}, nil)
	c.Stream, c.Stage = "Websites", "Vault → Website"
	m, err := json.Marshal(c)
	if err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{`"state"`, `"source_count"`, `"target_count"`,
		`"gap_count"`, `"unexplained_ids"`, `"deferred_gap_ids"`,
		`"unexplained_count"`, `"deferred_gap_count"`, `"stream"`, `"stage"`} {
		if !strings.Contains(string(m), key) {
			t.Fatalf("marshal missing %s: %s", key, m)
		}
	}
}
