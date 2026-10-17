// banishgate_test.go — T-40's THE CONFIRMATION GATE battery (the
// reference's TestTheConfirmationGate selections): the run opens with
// the vault scan, the count is SPOKEN and asked over the injected
// channel, and NOTHING is removed until the verdict — confirmed
// deletes (both doors + the twins sweep), declined/timeout keep
// everything and name the reason, no channel defers, config opts back
// into auto, and the tally never claims removals that did not happen.
package websitepipeline

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/handdelivery"
)

// gateChannel records the ask and answers a scripted verdict.
type gateChannel struct {
	verdict  string
	reportFn func(n int)
	asked    []int // the pending counts it was asked about
	panics   bool
}

func (g *gateChannel) confirm(items []PendingBanishment, log LogFn) *BanishConfirmResult {
	g.asked = append(g.asked, len(items))
	if g.panics {
		panic("channel is down")
	}
	res := &BanishConfirmResult{Verdict: g.verdict}
	if g.reportFn != nil {
		res.Report = g.reportFn
	}
	return res
}

func (c *pipeCase) storeMarkedNote(u string) string {
	c.t.Helper()
	_, note := c.storeProperNote(u)
	c.tagNote(note, "🗑️")
	return note
}

func (c *pipeCase) tableText() string {
	path := DecommissionTablePath(c.vault)
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	return readFileT(c.t, path)
}

func (c *pipeCase) trashNames() []string {
	dir := filepath.Join(c.vault, ".trash", "banished")
	if _, err := os.Stat(dir); err != nil {
		return nil
	}
	return listDirT(c.t, dir)
}

func TestBanishGateConfirmedDeletes(t *testing.T) {
	c := newPipeCase(t)
	reported := 0
	ch := &gateChannel{verdict: "confirmed", reportFn: func(n int) { reported = n }}
	c.gateFn = ch.confirm
	note := c.storeMarkedNote("https://marked.example/x")
	p := c.makePipeline()
	if len(ch.asked) != 1 || ch.asked[0] != 1 {
		t.Fatalf("the ask: %v", ch.asked)
	}
	if !c.hasLog("info", "🗑️ Vault scan: 1 note(s) marked for deletion (🗑️ / delete / auto_delete) — asking you to confirm before anything is removed") {
		t.Errorf("the spoken count missing:\n%s", c.allLogs())
	}
	// the note left the library for .trash/banished
	if _, err := os.Stat(note); !os.IsNotExist(err) {
		t.Error("the marked note still sits in the library")
	}
	if len(c.trashNames()) != 1 {
		t.Errorf("trash: %v", c.trashNames())
	}
	dismissed, _ := c.db.IsDismissed("https://marked.example/x")
	if !dismissed {
		t.Error("the URL was not blacklisted")
	}
	if len(p.BanishedURLs()) != 1 || p.BanishedURLs()[0] != "https://marked.example/x" {
		t.Errorf("banished_urls: %v", p.BanishedURLs())
	}
	if reported != 1 {
		t.Errorf("the report answered %d", reported)
	}
	if !c.hasLogContaining("🗑️ Run tally: 1 website(s) removed this run") {
		t.Errorf("the tally line missing:\n%s", c.allLogs())
	}
	// THE LOOP, CLOSED: a future paste of the same URL is skipped
	callsBefore := len(c.fake.calls)
	r := p.ProcessLink("https://marked.example/x")
	if r.Outcome != "skipped" || !stringsContains(r.Error, "banished by owner") {
		t.Fatalf("re-paste result %+v", r)
	}
	if len(c.fake.calls) != callsBefore {
		t.Error("a banished URL was fetched")
	}
}

func TestBanishGateDeclinedKeepsEverything(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "declined"}).confirm
	note := c.storeMarkedNote("https://kept.example/x")
	c.makePipeline()
	if _, err := os.Stat(note); err != nil {
		t.Error("the declined run deleted the note")
	}
	if len(c.trashNames()) != 0 {
		t.Errorf("trash: %v", c.trashNames())
	}
	dismissed, _ := c.db.IsDismissed("https://kept.example/x")
	if dismissed {
		t.Error("the declined run blacklisted the URL")
	}
	if !c.hasLogContaining("🗑️ Run tally: 0 websites removed this run — 1 marked note(s) KEPT (you said keep") {
		t.Errorf("the kept tally missing:\n%s", c.allLogs())
	}
	if !c.hasLog("info", "👌 You kept them — the 1 marked note(s) stay in the vault; the marks stay too, I'll ask again next run") {
		t.Error("the declined line missing")
	}
}

func TestBanishGateTimeoutKeepsAndNamesTheReason(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "timeout"}).confirm
	note := c.storeMarkedNote("https://waited.example/x")
	c.makePipeline()
	if _, err := os.Stat(note); err != nil {
		t.Error("the timeout run deleted the note")
	}
	if !c.hasLog("info", "⌛ No answer in time — the 1 marked note(s) stay; the marks stay, I'll ask again next run") {
		t.Error("the timeout line missing")
	}
	if !c.hasLogContaining("no answer in time") {
		t.Error("the tally's reason missing")
	}
}

func TestBanishGateNoChannelDefersByDefault(t *testing.T) {
	c := newPipeCase(t)
	note := c.storeMarkedNote("https://deferred.example/x")
	p := c.makePipeline()
	if _, err := os.Stat(note); err != nil {
		t.Error("the deferred run deleted the note")
	}
	if p.Gate().Verdict != "defer" {
		t.Errorf("verdict: %v", p.Gate())
	}
	if !c.hasLogContaining("no confirmation channel is reachable") {
		t.Errorf("the defer warning missing:\n%s", c.allLogs())
	}
	if !c.hasLogContaining("🗑️ Run tally: 0 websites removed this run") {
		t.Error("the zero tally missing")
	}
}

func TestBanishGateConfigOptOutRestoresAuto(t *testing.T) {
	c := newPipeCase(t)
	c.extraCfg = map[string]any{"banish_confirm": false}
	note := c.storeMarkedNote("https://auto.example/x")
	p := c.makePipeline()
	if _, err := os.Stat(note); !os.IsNotExist(err) {
		t.Error("the auto run kept the note")
	}
	if p.Gate().Verdict != "auto" {
		t.Errorf("verdict: %v", p.Gate())
	}
	if len(p.BanishedURLs()) != 1 {
		t.Errorf("banished_urls: %v", p.BanishedURLs())
	}
}

func TestBanishGateBrokenChannelDefersSafely(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "confirmed", panics: true}).confirm
	note := c.storeMarkedNote("https://broken.example/x")
	p := c.makePipeline()
	if _, err := os.Stat(note); err != nil {
		t.Error("the broken channel deleted the note")
	}
	if p.Gate().Verdict != "defer" {
		t.Errorf("verdict: %v", p.Gate())
	}
	if !c.hasLogContaining("Banish confirmation failed") {
		t.Errorf("the failure line missing:\n%s", c.allLogs())
	}
}

func TestBanishGateBareStringVerdict(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = func(items []PendingBanishment, log LogFn) *BanishConfirmResult {
		return &BanishConfirmResult{Verdict: "declined"}
	}
	note := c.storeMarkedNote("https://bare.example/x")
	c.makePipeline()
	if _, err := os.Stat(note); err != nil {
		t.Error("a bare-string declined verdict deleted the note")
	}
}

func TestBanishGateTwinsLeaveTheOtherVaultOnConfirm(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "confirmed"}).confirm
	c.storeMarkedNote("https://twin.example/x")
	// the OTHER vault (config vault_path) carries a twin + a human note
	other := filepath.Join(c.tmp, "github-projects")
	os.MkdirAll(filepath.Join(other, "Design"), 0o755)
	twin := gwriteNote(t, other, "Design/twin.md", "https://twin.example/x", "[design]", "gitcurator", "full", "")
	human := gwriteNote(t, other, "Design/human.md", "https://twin.example/x", "[design]", "human", "full", "")
	c.extraCfg = map[string]any{"vault_path": other}
	c.makePipeline()
	if _, err := os.Stat(twin); !os.IsNotExist(err) {
		t.Error("the twin stayed in the other vault")
	}
	if _, err := os.Stat(human); err != nil {
		t.Error("a hand-written twin was touched (the sacred law)")
	}
	if !c.hasLogContaining("Both vaults: 1 twin note(s) removed from the other vault") {
		t.Errorf("the twins line missing:\n%s", c.allLogs())
	}
}

func TestBanishGateSameVaultNeverSweptAsOther(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "confirmed"}).confirm
	note := c.storeMarkedNote("https://same.example/x")
	c.extraCfg = map[string]any{"vault_path": c.vault}
	c.makePipeline()
	// the note left exactly once (to this vault's own .trash)
	if len(c.trashNames()) != 1 {
		t.Errorf("trash: %v", c.trashNames())
	}
	_ = note
}

func TestBanishGateHoldsTheTableDoorNotTheDead(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "declined"}).confirm
	// a 🗑 table row (the destructive door — held) + a 🪦 dead row (the
	// owner's explicit hand — runs unchanged)
	c.writeTable([][2]string{
		{"https://held.example/h", "🗑️ delete"},
		{"https://deadc.example/d", "🪦 dead"},
	})
	c.makePipeline()
	dismissedHeld, _ := c.db.IsDismissed("https://held.example/h")
	if dismissedHeld {
		t.Error("the held 🗑 row was consumed")
	}
	dismissedDead, _ := c.db.IsDismissed("https://deadc.example/d")
	if !dismissedDead {
		t.Error("the dead row's verdict did not run (the gate holds ONLY the destructive door)")
	}
	if !stringsContains(c.tableText(), "🗑️ delete") {
		t.Error("the held row's mark was rewritten")
	}
}

func TestBanishGateConfirmedTableRowsAreNotReAsked(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "confirmed"}).confirm
	c.writeTable([][2]string{
		{"https://history.example/h", "🗑️ banished — confirmed 2026-10-01"},
	})
	p := c.makePipeline()
	if p.Gate().Pending != 0 {
		t.Errorf("a confirmed history row was re-asked: %d", p.Gate().Pending)
	}
}

func TestBanishTallyDedupesAcrossTheDoors(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "confirmed"}).confirm
	// the SAME link marked on its note AND in the table = ONE deletion
	note := c.storeMarkedNote("https://both.example/x")
	c.writeTable([][2]string{{"https://both.example/x", "🗑️ delete"}})
	p := c.makePipeline()
	if len(p.BanishedURLs()) != 1 {
		t.Errorf("banished_urls: %v", p.BanishedURLs())
	}
	if len(c.trashNames()) != 1 {
		t.Errorf("trash: %v", c.trashNames())
	}
	_ = note
}

func TestBanishHeldGateNeverClaimsRemovals(t *testing.T) {
	c := newPipeCase(t)
	c.gateFn = (&gateChannel{verdict: "timeout"}).confirm
	note := c.storeMarkedNote("https://heldtally.example/x")
	c.writeTable([][2]string{{"https://heldtally.example/x", "🗑️ delete"}})
	p := c.makePipeline()
	if len(p.BanishedURLs()) != 0 {
		t.Errorf("a held run claimed removals: %v", p.BanishedURLs())
	}
	if !c.hasLogContaining("🗑️ Run tally: 0 websites removed this run") {
		t.Error("the zero tally missing")
	}
	if _, err := os.Stat(note); err != nil {
		t.Error("the held run removed the note")
	}
}

func TestHandDeliveryQueueWireFormat(t *testing.T) {
	// the queue.json wire shape through the Go writer (the golden pins
	// the reference's bytes; this pins the structural law)
	c := newPipeCase(t)
	handdelivery.EnqueueHandDelivery(c.vault, []string{"https://hand.example/x"},
		map[string]string{"https://hand.example/x": "HTTP 403 — refused"}, nil, "")
	q := handdelivery.ReadQueue(c.vault)
	if len(q.Links.Keys()) != 1 {
		t.Fatalf("queue: %v", q.Links.Keys())
	}
	dump := q.Dumps()
	if !stringsContains(dump, `"wall": "HTTP 403 — refused"`) ||
		!stringsContains(dump, `"suggested": "hand-hand-example-`) {
		t.Errorf("queue json:\n%s", dump)
	}
	// re-enqueue is a no-op
	handdelivery.EnqueueHandDelivery(c.vault, []string{"https://hand.example/x"}, nil, nil, "")
	q2 := handdelivery.ReadQueue(c.vault)
	if q2.Links.Len() != 1 {
		t.Errorf("duplicate queue row: %d", q2.Links.Len())
	}
}

func stringsContains(s, part string) bool { return strings.Contains(s, part) }
