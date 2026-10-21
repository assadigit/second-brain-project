// handharvest_test.go — SESSION-060's deferred-debts slice: the port of
// the reference's test_handharvest §4 catch-up selections that pin the
// probe's own failure laws (the harvest's probe-panic-ladder cousin —
// the recorded remainder of the SESSION-059 delivered-law fix, where
// the redo scan got the ladder and the harvest was left recorded, not
// fixed). The reference's _Broken probe raises on BOTH is_dismissed +
// processed_row and harvest_hand_rows' ONE inner try/except answers
// "a broken probe harvests nothing" — the Go twin of a raising probe
// is the closed *StateDB (SQL() is nil after Close by the store's own
// law, so every probe call panics rather than erroring) plus the
// state=nil hermetic shape (the docstring's "the state probe is
// optional"). The happy paths (the backlog row retiring, the sweep,
// the verdicts, the waiting rows) are pinned by the golden corpus
// (flow E's harvest sections, byte-compared against the reference's
// own runs) and the app seam — these selections drive HarvestHandRows
// directly, exactly the shapes the corpus cannot fabricate.
package websitepipeline

import (
	"os"
	"testing"
)

// test_a_broken_state_probe_harvests_nothing: the setup is the
// harvest's OWN happy shape (a proper note on disk + a processed row +
// the 🖐 gesture — the probe working would retire the row); the probe
// breaking alone must harvest nothing, and the gesture stays (the
// reference's _Broken raises on both probe calls; the Go twin is the
// closed DB — every IsDismissed/ProcessedRow call panics).
func TestHarvestABrokenStateProbeHarvestsNothing(t *testing.T) {
	c := newPipeCase(t)
	proper := c.writeProperNote(handWalledURL)
	c.db.MarkProcessed(handWalledCanon, proper, "Design", "", "full")
	tablePath := c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	c.db.Close() // the broken probe
	if got := HarvestHandRows(c.db, c.vault, nil); got != 0 {
		t.Fatalf("a broken probe harvests nothing: %d", got)
	}
	raw, err := os.ReadFile(tablePath)
	if err != nil {
		t.Fatal(err)
	}
	table := string(raw)
	if !rowsContain(table, "🖐") {
		t.Fatalf("the gesture stays: %q", table)
	}
	if rowsContain(table, "✅ hand-delivered") {
		t.Fatalf("the green verdict never landed: %q", table)
	}
}

// test_no_state_degrades_to_nothing: the hermetic law — the state
// probe is optional (the reference's docstring); state=None raises
// AttributeError inside the inner try/except and harvests nothing,
// the Go twin is state=nil (the nil receiver panics inside the probe
// — the ladder answers the same broken verdict).
func TestHarvestNoStateDegradesToNothing(t *testing.T) {
	c := newPipeCase(t)
	proper := c.writeProperNote(handWalledURL)
	c.db.MarkProcessed(handWalledCanon, proper, "Design", "", "full")
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if got := HarvestHandRows(nil, c.vault, nil); got != 0 {
		t.Fatalf("state=None harvests nothing: %d", got)
	}
}
