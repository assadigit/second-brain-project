// redowiring_test.go — the GUI-side hero/wiring equivalents of the
// reference's test_noteredo TestHeroRedoRouting + TestRedoPassWiring
// (app/tests/test_noteredo.py:795-973), ported at the Go app seam —
// the LAST §9 deferred-test-debt slice (the parity report §9's one
// remaining entry, closed here).
//
// The reference's two classes, both ported:
//   - TestHeroRedoRouting — the caught-up claim may only be said
//     after the redo check (a hand link whose note is not properly
//     stored is work, not up-to-date);
//   - TestRedoPassWiring — the redo pass itself at the GUI seam: the
//     once-per-session bound, the mini-batch's ROUTING LAW, and the
//     busy/delivery/pipeline gates.
//
// Already covered at this seam when this file landed (the selections
// cited, never duplicated):
//   - test_redo_work_holds_back_the_caught_up_claim →
//     TestHeroUIFanfareCaughtUpRedoPass (heroui_test.go);
//   - test_no_redo_work_keeps_the_old_law +
//     test_no_redo_rows_is_a_quiet_noop →
//     TestHeroUIFanfareCaughtUpPlainClaimIntact (the claim + zero
//     batches);
//   - the guards' data halves (the delivery/batch/pipeline/vault
//     ladders + the tried belt itself) → TestFanfareRedoHandNotesScan
//     (internal/app/fanfaredoors_test.go).
//
// TWO RECORDS ride with this port:
//   - THE START-FAILURE BELTS (the reference's
//     test_a_start_failure_is_honest_not_fatal): the Go startRedoRerun
//     carries the reference's try/except twin as TWO belts — the
//     recover belt (unreachable by construction: SyncRunBatch's
//     guardCommand owns every panic from the engine seam, the crashed
//     verdict + the lock-leak laws pinned at the commands seam) and
//     the "already running" refusal line (the race-window belt: the
//     redo scan and the start gate on the SAME shared batch lock, so
//     the refusal needs another holder between the scan's read and
//     the goroutine's acquire — microseconds, never deterministic).
//     The never-fires-belt class of SESSION-060: recorded, not
//     changed — the honest refusal (⏳ A batch is already running) is
//     SyncRunBatch's own story and the crashed-run shape is pinned at
//     the commands seam (commands_test.go's crash rows).
//   - THE PENDING-COUNT MODEL (the reference's
//     test_a_pending_delivery_holds_the_pass, second half): the
//     reference keeps a separate _chrome_delivery_pending count (the
//     scorecard-stash hold); the Go model folds the whole pending
//     window into chromeDeliveryBusy (the backlog lives only while a
//     delivery runs — T-79d's recorded divergence; the scorecard
//     modal is the owner's open call). The busy flag covers every
//     reachable state the reference's two-term guard reads; the
//     unreachable pending-without-busy shape has no Go counterpart
//     to pin.
package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// seedRedoSubject writes the redo pass's subject at the GUI seam: the
// 🖐 table row + the delivered page + the queue entry, with NO
// processed row (the note never landed — the false success).
func seedRedoSubject(t *testing.T, rig *heroRig, url string) {
	t.Helper()
	vault := rigConfigVault(t, rig)
	seedFalseSuccess(t, vault, url, "HTTP 403")
	writeHandRowsTable(t, vault, [][3]string{
		{url, "\U0001f590", "HTTP 403"},
	})
}

// waitBatchIdle polls the Commands' batch lock until no batch holds it
// (the once-per-session bound's settle: the first redo batch must be
// fully home before the second click's scan runs — the selection pins
// the tried belt, not the busy gate).
func waitBatchIdle(t *testing.T, rig *heroRig) {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if !rig.ui.commands.BatchBusy() {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatal("the batch never released the lock")
}

// assertRedoRouting pins THE ROUTING LAW at the GUI seam (the
// reference's own words: website links, never GitHub urls — the
// v0.57.0 fix; the owner's report was exactly the old call shape
// feeding the website URLs to the GitHub loop, which marked each one
// "non-GitHub URL — skipped" while the Websites pipeline never saw
// them).
func assertRedoRouting(t *testing.T, rig *heroRig, site, url string) {
	t.Helper()
	rig.mu.Lock()
	opts := append([]app.BatchOptions(nil), rig.batchOpts...)
	rig.mu.Unlock()
	if len(opts) != 1 {
		t.Fatalf("%s: exactly one redo batch (have %d)", site, len(opts))
	}
	if len(opts[0].URLs) != 0 {
		t.Fatalf("%s: THE ROUTING LAW — the GitHub list stays empty: %v",
			site, opts[0].URLs)
	}
	if len(opts[0].NonGithub) != 1 || opts[0].NonGithub[0] != url {
		t.Fatalf("%s: the redo rows ride as the website list: %v",
			site, opts[0].NonGithub)
	}
	if opts[0].Source != "direct" {
		t.Fatalf("%s: the intake source = %q", site, opts[0].Source)
	}
}

// TestRedoRoutingDegradedProbeKeepsTheClaim is the reference's
// test_a_bare_hero_without_the_collaborator_is_unchanged, at the Go
// seam: the reference's bare hero carries no processing-control
// collaborator (the getattr-missing shape) and still claims caught
// up; the Go composition wires the collaborator statically, so the
// honest twin is its QUIET DEGRADATION — a state ledger that cannot
// be read (a garbage cache.db) answers nothing (the hermetic law:
// nothing can be proven false), and the caught-up claim still lands,
// nothing breaks, nothing starts.
func TestRedoRoutingDegradedProbeKeepsTheClaim(t *testing.T) {
	rig := newFanfareUIRig(t)
	seedRedoSubject(t, rig, "https://example.com/false-success")
	if err := os.WriteFile(filepath.Join(constants.AppDir(), "cache.db"),
		[]byte("this is not a sqlite database at all"), 0o644); err != nil {
		t.Fatal(err)
	}
	res := rig.ui.HeroClick("")
	if res.State != app.HeroSync || res.Format != app.FormatReady {
		t.Fatalf("the degraded probe keeps the claim = %q / %q",
			res.State, res.Format)
	}
	if !strings.Contains(rig.logs(), "All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the claim line missing:\n%s", rig.logs())
	}
	if strings.Contains(rig.logs(), "\U0001f501") {
		t.Fatalf("the degraded probe never redoes:\n%s", rig.logs())
	}
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches != 0 {
		t.Fatalf("the degraded probe starts nothing (have %d)", batches)
	}
}

// TestRedoWiringRoutingLaw is the reference's
// test_the_pass_starts_the_website_mini_batch, at BOTH Go call sites
// (the caught-up click — hero.py's caught-up tail; the fanfare
// crossing — processing_finished's tail): the pass starts ONE batch,
// the GitHub list EMPTY, the redo rows riding as the website list,
// the intake source "direct" — and the 🔁 line speaks (the reference
// asserts 'NOT properly stored' in the log; the port's own words).
func TestRedoWiringRoutingLaw(t *testing.T) {
	const redoURL = "https://example.com/false-success"

	// call site 1 — the caught-up click.
	rig := newFanfareUIRig(t)
	seedRedoSubject(t, rig, redoURL)
	res := rig.ui.HeroClick("")
	if !strings.Contains(rig.logs(), "\U0001f501 1 hand note(s) NOT properly stored") {
		t.Fatalf("the caught-up redo line missing:\n%s", rig.logs())
	}
	if res.Format != "" {
		t.Fatalf("the redo case keeps the bar: %q", res.Format)
	}
	waitBatchCalls(t, rig, 1)
	assertRedoRouting(t, rig, "the caught-up click", redoURL)

	// call site 2 — the fanfare crossing.
	rig = newFanfareUIRig(t)
	seedRedoSubject(t, rig, redoURL)
	rig.ui.onBatchFanfare(app.BatchFinished{OK: true, Natural: true})
	waitBatchCalls(t, rig, 1)
	assertRedoRouting(t, rig, "the fanfare crossing", redoURL)
}

// TestRedoWiringOncePerSession is the reference's
// test_once_per_session_is_the_bound: the first caught-up click redoes
// (one batch, no claim); the second click in the SAME session — after
// the batch is fully home — starts NOTHING (the LLM's answer will not
// change on an immediate second ask; the tried belt) and the claim
// lands, with the 🔁 line spoken exactly once.
func TestRedoWiringOncePerSession(t *testing.T) {
	rig := newFanfareUIRig(t)
	seedRedoSubject(t, rig, "https://example.com/false-success")

	// the first click: work started — no claim, the redo batch.
	rig.ui.HeroClick("")
	if strings.Contains(rig.logs(), "All caught up") {
		t.Fatalf("work started — no caught-up claim:\n%s", rig.logs())
	}
	if !strings.Contains(rig.logs(), "\U0001f501 1 hand note(s) NOT properly stored") {
		t.Fatalf("the redo line missing:\n%s", rig.logs())
	}
	waitBatchCalls(t, rig, 1)
	waitBatchIdle(t, rig)

	// the second click in the same session: the bound holds.
	rig.ui.HeroClick("")
	if !strings.Contains(rig.logs(), "All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the second click claims caught up:\n%s", rig.logs())
	}
	if got := strings.Count(rig.logs(), "\U0001f501 1 hand note(s)"); got != 1 {
		t.Fatalf("the redo line spoke %d times (once per session is the bound):\n%s",
			got, rig.logs())
	}
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches != 1 {
		t.Fatalf("once per session is the bound (have %d batches)", batches)
	}
}

// TestRedoWiringHeldByChromeDelivery is the reference's
// test_a_pending_delivery_holds_the_pass (the Go model's half — see
// the pending-count record in the file header): a gathering Chrome
// delivery holds the redo pass (its own re-run owns those links now),
// the caught-up claim still lands — and the RELEASED pass redoes on
// the next click: the deferral never consumes the tried belt (the
// guard answers BEFORE the belt marks).
func TestRedoWiringHeldByChromeDelivery(t *testing.T) {
	rig := newFanfareUIRig(t)
	seedRedoSubject(t, rig, "https://example.com/false-success")

	rig.ui.mu.Lock()
	rig.ui.chromeDeliveryBusy = true
	rig.ui.mu.Unlock()
	res := rig.ui.HeroClick("")
	if strings.Contains(rig.logs(), "\U0001f501") {
		t.Fatalf("a gathering delivery holds the redo:\n%s", rig.logs())
	}
	if !strings.Contains(rig.logs(), "All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the held pass still claims caught up:\n%s", rig.logs())
	}
	if res.State != app.HeroSync || res.Format != app.FormatReady {
		t.Fatalf("the held pass = %q / %q", res.State, res.Format)
	}
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches != 0 {
		t.Fatalf("the held pass starts nothing (have %d)", batches)
	}

	// released: the next click redoes — the deferral was never a try.
	rig.ui.mu.Lock()
	rig.ui.chromeDeliveryBusy = false
	rig.ui.mu.Unlock()
	rig.ui.HeroClick("")
	waitBatchCalls(t, rig, 1)
	if !strings.Contains(rig.logs(), "\U0001f501 1 hand note(s) NOT properly stored") {
		t.Fatalf("the released pass redoes:\n%s", rig.logs())
	}
}

// TestRedoWiringHeldByRunningBatch is the reference's
// test_a_running_batch_holds_the_pass: a running batch owns the
// table's truth — the redo pass defers to the running batch's own
// end-of-run fanfare, the caught-up claim still lands, and the
// running batch stays the only one.
func TestRedoWiringHeldByRunningBatch(t *testing.T) {
	rig := newFanfareUIRig(t)
	seedRedoSubject(t, rig, "https://example.com/false-success")

	rig.batchBlock = make(chan struct{})
	done := make(chan *app.BatchResult, 1)
	go func() {
		done <- rig.commands.SyncRunBatch(context.Background(),
			app.BatchOptions{Source: "bot"})
	}()
	waitBatchBusy(t, rig)
	res := rig.ui.HeroClick("")
	if strings.Contains(rig.logs(), "\U0001f501") {
		t.Fatalf("a running batch holds the redo:\n%s", rig.logs())
	}
	if !strings.Contains(rig.logs(), "All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the held pass still claims caught up:\n%s", rig.logs())
	}
	if res.State != app.HeroSync || res.Format != app.FormatReady {
		t.Fatalf("the held pass = %q / %q", res.State, res.Format)
	}
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches != 1 {
		t.Fatalf("the running batch is the only one (have %d)", batches)
	}
	close(rig.batchBlock)
	<-done
}

// TestRedoWiringWebsitesPipelineGate is the reference's
// test_the_websites_pipeline_gate_holds: the Websites pipeline off —
// the redo pass is a quiet no-op (the false success is not the
// machine's to redo when its pipeline is dark), the caught-up claim
// still lands, nothing starts.
func TestRedoWiringWebsitesPipelineGate(t *testing.T) {
	rig := newFanfareUIRig(t)
	seedRedoSubject(t, rig, "https://example.com/false-success")

	rig.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	res := rig.ui.HeroClick("")
	if strings.Contains(rig.logs(), "\U0001f501") {
		t.Fatalf("the pipeline gate holds the redo:\n%s", rig.logs())
	}
	if !strings.Contains(rig.logs(), "All caught up — nothing undone in the bot queue.") {
		t.Fatalf("the gated pass still claims caught up:\n%s", rig.logs())
	}
	if res.State != app.HeroSync || res.Format != app.FormatReady {
		t.Fatalf("the gated pass = %q / %q", res.State, res.Format)
	}
	rig.mu.Lock()
	batches := rig.batchCalls
	rig.mu.Unlock()
	if batches != 0 {
		t.Fatalf("the gated pass starts nothing (have %d)", batches)
	}
}
