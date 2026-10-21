// noteredo_test.go — SESSION-059's deferred-debts slice: the port of
// the reference's test_noteredo §3 TestRedoScan (the CORE-level
// redo-scan selections — the reference's own authority on the scan's
// shapes, per TestRedoPassWiring's docstring: "the scan itself is
// tested at the core level (TestRedoScan)") + test_iconcolumn §7's
// never-re-ask selection (the hand-scan side of the SAME delivered
// law). The app seam (T-79f's TestFanfareRedoHandNotesScan) covers
// the guards + ONE false-success shape; the CLI golden sections cover
// the routing; these selections drive ScanMasterRedoRows directly —
// every edge the reference pins: the owner's exact case (consumed +
// half-fetched _review note), the proper note, the no-state-row
// falsest success, the un-delivered Chrome pass's set, the routing
// bug's residue (page landed, consume never stamped), the verdict
// precedence, the dismissal, state=None + the broken probe (the
// hermetic law), the missing table, first-wins dedupe — and the
// consumed/landed rows the hand scan never re-asks.
package websitepipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/links"
)

// handDeliverConsumed is the reference rig's deliver(consume=True): the
// enqueue + the page + the queue row stamped consumed — the
// FALSE-SUCCESS shape: consumed in an earlier run, the note never
// became proper. The first consume date is the record; the reference's
// own rig stamps the same fixed date.
func (c *pipeCase) handDeliverConsumed(url string) string {
	c.t.Helper()
	path := c.handDeliver(url)
	q := handdelivery.ReadQueue(c.vault)
	if m := q.Meta(links.NormalizeWebsiteURL(url)); m != nil {
		m.Set("consumed", "2026-10-01 10:00")
		if err := q.WriteAtomic(c.vault); err != nil {
			c.t.Fatal(err)
		}
	}
	return path
}

// --- TestRedoScan: the core-level scan selections ------------------------

// test_the_owners_exact_case_is_the_redo_set: delivered + consumed +
// only a half-fetched _review note → "false success and must be redo".
func TestRedoScanTheOwnersExactCaseIsTheRedoSet(t *testing.T) {
	c := newPipeCase(t)
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	rows := ScanMasterRedoRows(c.vault, c.db, nil)
	if len(rows) != 1 {
		t.Fatalf("the redo rows: %+v", rows)
	}
	if rows[0].URL != handWalledCanon {
		t.Fatalf("the url: %+v", rows[0])
	}
	if !strings.Contains(rows[0].Reason, "_review") {
		t.Fatalf("the reason carries the _review story: %q", rows[0].Reason)
	}
}

// test_a_proper_note_is_not_the_redos.
func TestRedoScanAProperNoteIsNotTheRedos(t *testing.T) {
	c := newPipeCase(t)
	proper := c.writeProperNote(handWalledURL)
	c.db.MarkProcessed(handWalledCanon, proper, "Design", "", "full")
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if rows := ScanMasterRedoRows(c.vault, c.db, nil); len(rows) != 0 {
		t.Fatalf("a proper note is done, honestly: %+v", rows)
	}
}

// test_no_state_row_after_delivery_is_the_redo_set: consumed, but the
// note never even landed — the falsest of the false successes.
func TestRedoScanNoStateRowAfterDeliveryIsTheRedoSet(t *testing.T) {
	c := newPipeCase(t)
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	rows := ScanMasterRedoRows(c.vault, c.db, nil)
	if len(rows) != 1 {
		t.Fatalf("the redo rows: %+v", rows)
	}
	if !strings.Contains(rows[0].Reason, "never landed") {
		t.Fatalf("the reason carries the never-landed story: %q", rows[0].Reason)
	}
}

// test_an_undelivered_hand_row_is_the_chrome_passs: no queue row, no
// page file — that row wants a Chrome tab, never a redo.
func TestRedoScanAnUndeliveredHandRowIsTheChromePasss(t *testing.T) {
	c := newPipeCase(t)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if rows := ScanMasterRedoRows(c.vault, c.db, nil); len(rows) != 0 {
		t.Fatalf("the un-delivered row is the Chrome pass's: %+v", rows)
	}
}

// test_a_page_file_without_the_consume_stamp_is_delivered: the
// delivery wrote the file but the consume never happened (the routing
// bug's exact residue) — the redo takes it too.
func TestRedoScanAPageFileWithoutTheConsumeStampIsDelivered(t *testing.T) {
	c := newPipeCase(t)
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	c.handDeliver(handWalledURL) // no consume stamp
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	rows := ScanMasterRedoRows(c.vault, c.db, nil)
	if len(rows) != 1 {
		t.Fatalf("the landed page is a delivery: %+v", rows)
	}
}

// test_the_verdicts_win: dead / reviewed / revived outrank the hand
// gesture — none of them is ever a redo.
func TestRedoScanTheVerdictsWin(t *testing.T) {
	for _, s := range []string{"🪦 dead", "✅ reviewed", "♻️ revived"} {
		c := newPipeCase(t)
		half := c.writeHalfNote(handWalledURL, "a.md")
		c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
		c.handDeliverConsumed(handWalledURL)
		c.writeTable([][2]string{{handWalledURL, s}})
		if rows := ScanMasterRedoRows(c.vault, c.db, nil); len(rows) != 0 {
			t.Fatalf("the verdict %q owns the cell: %+v", s, rows)
		}
	}
}

// test_a_dismissed_link_is_never_redone: retired rows are the
// verdicts' business.
func TestRedoScanADismissedLinkIsNeverRedone(t *testing.T) {
	c := newPipeCase(t)
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	if err := c.db.Dismiss(handWalledCanon, "owner retired it"); err != nil {
		t.Fatal(err)
	}
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if rows := ScanMasterRedoRows(c.vault, c.db, nil); len(rows) != 0 {
		t.Fatalf("a dismissed link is never redone: %+v", rows)
	}
}

// test_no_state_degrades_to_nothing: the hermetic law — nothing can
// be proven false without the ledger.
func TestRedoScanNoStateDegradesToNothing(t *testing.T) {
	c := newPipeCase(t)
	c.writeHalfNote(handWalledURL, "a.md")
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	if rows := ScanMasterRedoRows(c.vault, nil, nil); len(rows) != 0 {
		t.Fatalf("state=None redoes nothing: %+v", rows)
	}
}

// test_a_broken_probe_redoes_nothing: a probe whose every query errors
// answers like no probe at all (the reference's _Broken raises on
// both is_dismissed and processed_row; the Go twin of a raising
// probe is a closed *StateDB — every IsDismissed/ProcessedRow call
// errors).
func TestRedoScanABrokenProbeRedoesNothing(t *testing.T) {
	c := newPipeCase(t)
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{{handWalledURL, "🖐 hand"}})
	c.db.Close() // the broken probe
	if rows := ScanMasterRedoRows(c.vault, c.db, nil); len(rows) != 0 {
		t.Fatalf("a broken probe redoes nothing: %+v", rows)
	}
}

// test_a_missing_table_reads_empty.
func TestRedoScanAMissingTableReadsEmpty(t *testing.T) {
	c := newPipeCase(t)
	if rows := ScanMasterRedoRows(c.vault, c.db, nil); len(rows) != 0 {
		t.Fatalf("no table, no redo: %+v", rows)
	}
}

// test_duplicate_rows_first_wins: the first row carries the verdict
// (a hand row above a dead duplicate is still the redo's).
func TestRedoScanDuplicateRowsFirstWins(t *testing.T) {
	c := newPipeCase(t)
	half := c.writeHalfNote(handWalledURL, "a.md")
	c.db.MarkProcessed(handWalledCanon, half, "", "", "full")
	c.handDeliverConsumed(handWalledURL)
	c.writeTable([][2]string{
		{handWalledURL, "🖐 hand"},
		{handWalledURL, "🪦 dead"},
	})
	rows := ScanMasterRedoRows(c.vault, c.db, nil)
	if len(rows) != 1 {
		t.Fatalf("the first row wins: %+v", rows)
	}
}

// --- test_iconcolumn §7's delivered law at the hand scan -----------

// test_delivered_and_consumed_rows_never_re_ask: _URL consumed (the
// record says the page was taken), _URL2 landed but no batch consumed
// it yet — the hand scan owes NEITHER a Chrome tab (the master
// scans' delivered probe: consumed OR landed).
func TestHandScanDeliveredAndConsumedRowsNeverReAsk(t *testing.T) {
	c := newPipeCase(t)
	u1 := "https://done.example/one"
	u2 := "https://landed.example/two"
	handdelivery.EnqueueHandDelivery(c.vault, []string{u1, u2},
		nil, nil, false, "")
	q := handdelivery.ReadQueue(c.vault)
	if m := q.Meta(links.NormalizeWebsiteURL(u1)); m != nil {
		m.Set("consumed", "2026-10-07 20:00")
		if err := q.WriteAtomic(c.vault); err != nil {
			t.Fatal(err)
		}
	}
	page := filepath.Join(handdelivery.HandDeliveryDir(c.vault),
		handdelivery.SuggestedFilename(u2))
	os.MkdirAll(filepath.Dir(page), 0o755)
	if err := os.WriteFile(page, []byte("<html>landed</html>"), 0o644); err != nil {
		t.Fatal(err)
	}
	c.writeTable([][2]string{{u1, "🖐 hand"}, {u2, "🖐 hand"}})
	if rows := ScanMasterHandRows(c.vault, nil); len(rows) != 0 {
		t.Fatalf("delivered and consumed rows never re-ask: %+v", rows)
	}
}
