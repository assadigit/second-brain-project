package app

// decomdoors_test.go — T-79e: the decommission picker's hermetic
// battery, pinned against the reference's own flow
// (processing_control.py:807-998 + _scan_backlog_alive at 152-173,
// all at 98f4fc7). The whole door rides local files + the rig's own
// temp cache.db (no network anywhere near — no new seams); every box
// + line asserted by level + verbatim text; the graveyard table rides
// the real writers + the real consume (the one-ledger law end to
// end).

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// newDecomDoorRig is the decommission doors' harness (the Chrome
// doors' rig shape): the Websites pipeline on + a real temp Websites
// vault + the rig's own temp APP_DIR (cache.db lands there).
func newDecomDoorRig(t *testing.T) (*Settings, *busRec, string) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	s, rec, _ := newSettingsTest(t)
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	vault := t.TempDir()
	s.cfg.Set("website_vault_path", vault)
	return s, rec, vault
}

// writeBacklogPlaceholder writes one fetch-failed placeholder note
// into the vault's _review (the scan's own frontmatter contract:
// managed_by gitcurator + fetch_status failed + an http(s) source).
func writeBacklogPlaceholder(t *testing.T, vault, name, src string) string {
	t.Helper()
	dir := filepath.Join(vault, "_review")
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, name)
	note := fmt.Sprintf("---\nsource: %q\nfetch_status: \"failed\"\nmanaged_by: \"gitcurator\"\n---\n\n> [!info] Managed by GitCurator — machine-written note.\n\nbody\n", src)
	if err := os.WriteFile(path, []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// decomState opens the rig's cache.db (the temp APP_DIR's own).
func decomState(t *testing.T) *websitepipeline.WebsiteStateDB {
	t.Helper()
	state, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(nil))
	if err != nil {
		t.Fatal(err)
	}
	return state
}

// flipTableRowStatus rewrites the row carrying url with the given
// Status cell text (the hand-edit twin — the owner's Obsidian gesture).
func flipTableRowStatus(t *testing.T, vault, url, status string) {
	t.Helper()
	path := websitepipeline.DecommissionTablePath(vault)
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimRight(string(data), "\n"), "\n")
	found := false
	for i, line := range lines {
		if strings.Contains(line, url) && strings.HasPrefix(line, "| ") {
			parts := strings.Split(line, "|")
			if len(parts) >= 8 {
				parts[6] = " " + status + " "
				lines[i] = strings.Join(parts, "|")
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no table row carries %q", url)
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}

// TestDecommissionDoorGuards pins the two-guard ladder (the shared
// Websites ladder — the reference's own order; NO busy guard on this
// door, pinned by the absence assertions in the scan tests).
func TestDecommissionDoorGuards(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)

	// the pipeline off → the remedy box.
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", false)
		return p
	}())
	out := s.DecommissionScan()
	if out.Box == nil || out.Box.Title != "Websites pipeline is off" ||
		out.Box.Body != "Turn the Websites pipeline on first (Settings → 📁 Vault)." {
		t.Fatalf("the pipeline-off box: %+v", out.Box)
	}
	if out.Links != nil {
		t.Fatal("no links past the guard")
	}

	// the pipeline on but no vault → the vault box.
	s.cfg.Set("pipelines", func() *storage.OrderedMap {
		p := storage.NewOrderedMap()
		p.Set("github", true)
		p.Set("websites", true)
		return p
	}())
	s.cfg.Set("website_vault_path", "")
	out = s.DecommissionScan()
	if out.Box == nil || out.Box.Title != "No Websites vault" ||
		out.Box.Body != "Set the Websites vault first (Settings → 📁 Vault)." {
		t.Fatalf("the no-vault box: %+v", out.Box)
	}

	// a vault that is not a directory → the same box.
	s.cfg.Set("website_vault_path", filepath.Join(vault, "not-a-dir"))
	out = s.DecommissionScan()
	if out.Box == nil || out.Box.Title != "No Websites vault" {
		t.Fatalf("the non-dir vault box: %+v", out.Box)
	}
	if len(rec.logs) != 0 {
		t.Fatalf("the guards never speak on the desk: %d lines", len(rec.logs))
	}
}

// TestDecommissionScanEmptyBacklog pins the empty box in the
// reference's own words (the graveyard stays available for hand-rows;
// the ♻️ hint).
func TestDecommissionScanEmptyBacklog(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)
	out := s.DecommissionScan()
	table := websitepipeline.DecommissionTablePath(vault)
	if out.Box == nil || out.Box.Title != "No failed links waiting" {
		t.Fatalf("the empty box: %+v", out.Box)
	}
	want := "No app-owned fetch-failed placeholders are waiting in _review — nothing new to decommission.\n\n" +
		"The graveyard table stays available for hand-rows:\n" + table +
		"\n\n(Set a row's Status to 🪦 dead anytime — even a link that never got a placeholder. ♻️ revived brings one back.)"
	if out.Box.Body != want {
		t.Fatalf("the empty box body:\n got: %q\nwant: %q", out.Box.Body, want)
	}
	if out.Table != table {
		t.Fatalf("the table path: %q", out.Table)
	}
	if len(out.Links) != 0 {
		t.Fatalf("no links in the empty backlog: %v", out.Links)
	}
	// the empty box is a SUCCESS tone in the reference (the desk stays
	// quiet — no warning lines).
	for _, l := range rec.logs {
		if l.Level == "warning" {
			t.Fatalf("unexpected warning: %q", l.Msg)
		}
	}
}

// TestDecommissionScanPickerData pins the scan's own laws: the
// sorted-dedup URL set (two placeholders of one URL = ONE row), the
// non-failed + human notes filtered, the dismissed link retired (the
// state DB's probe — _scan_backlog_alive), the candidates pre-fill
// (the table written + the 📋 line), and NO busy guard (the scan runs
// with the lock held by nobody — the flag is not even read).
func TestDecommissionScanPickerData(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)
	p1 := writeBacklogPlaceholder(t, vault, "a-placeholder.md", "https://one.example/lost")
	writeBacklogPlaceholder(t, vault, "b-placeholder.md", "https://one.example/lost") // the same URL twice
	writeBacklogPlaceholder(t, vault, "c-placeholder.md", "https://two.example/wall")
	// a non-failed note (in _review for human eyes) + a human's note —
	// both never the picker's.
	writeBacklogPlaceholder(t, vault, "d-ok.md", "https://ok.example/")
	if err := os.WriteFile(filepath.Join(vault, "_review", "d-ok.md"),
		[]byte("---\nsource: \"https://ok.example/\"\nfetch_status: \"ok\"\nmanaged_by: \"gitcurator\"\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vault, "_review", "e-human.md"),
		[]byte("---\nsource: \"https://human.example/\"\nfetch_status: \"failed\"\nmanaged_by: \"a-human\"\n---\nbody\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// a link RETIRED by the fetcher's own auto-verdict: dismissed in
	// the state DB → not a waiting wall.
	state := decomState(t)
	if err := state.Dismiss("https://two.example/wall", "auto-verdict: dead"); err != nil {
		t.Fatal(err)
	}
	state.Close()

	out := s.DecommissionScan()
	if out.Box != nil {
		t.Fatalf("no box over a live backlog: %+v", out.Box)
	}
	if len(out.Links) != 1 || out.Links[0] != "https://one.example/lost" {
		t.Fatalf("the picker's links (dedup + retired filter): %v", out.Links)
	}
	if out.Table != websitepipeline.DecommissionTablePath(vault) {
		t.Fatalf("the table path: %q", out.Table)
	}
	// the candidates pre-fill: the row IS in the table now (the
	// one-ledger law starts here — the picker pre-fills the SAME table).
	rows := websitepipeline.ScanDecommissionTable(vault)
	if st, ok := rows["https://one.example/lost"]; !ok || !strings.Contains(strings.ToLower(st), "unreviewed") {
		t.Fatalf("the candidates row missing/unreviewed: %q", st)
	}
	wantLog(t, rec, "📋 1 row(s) written to", "info")
	if _, err := os.Stat(p1); err != nil {
		t.Fatal("the scan never sweeps — the placeholder stays")
	}
}

// TestDecommissionScanCandidatesWriteFailure pins the belt law: the
// table unreadable (a directory squatting on the table's path) → the
// writer's own ⚠ line speaks; the picker still answers its links
// (never a crash, never a silent no-op).
func TestDecommissionScanCandidatesWriteFailure(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)
	writeBacklogPlaceholder(t, vault, "a-placeholder.md", "https://one.example/lost")
	if err := os.MkdirAll(filepath.Join(vault, "_review", "DECOMMISSIONED.md"), 0o755); err != nil {
		t.Fatal(err)
	}
	out := s.DecommissionScan()
	if out.Box != nil {
		t.Fatalf("the write failure never blocks the picker: %+v", out.Box)
	}
	if len(out.Links) != 1 || out.Links[0] != "https://one.example/lost" {
		t.Fatalf("the links: %v", out.Links)
	}
	wantLog(t, rec, "⚠️ Could not read the graveyard table", "warning")
}

// TestDecommissionBuryCancel pins the cancel line (an empty picked
// list is the dialog's Cancel — the reference's if-not-picked log).
func TestDecommissionBuryCancel(t *testing.T) {
	s, rec, _ := newDecomDoorRig(t)
	title, body := s.DecommissionBury(nil)
	if title != "" || body != "" {
		t.Fatalf("the cancel answers no box: %q / %q", title, body)
	}
	wantLog(t, rec, "⏭️ Decommission cancelled — nothing was buried. The candidate rows wait in the table (set Status to 🪦 dead by hand in Obsidian; the next batch enforces it).", "info")
}

// TestDecommissionBuryHappyPath pins the whole post-dialog story over
// the real writers + the real consume: the marking (the SAME Status
// cell the owner would edit by hand), the immediate enforcement
// (dismiss + sweep + confirm — the placeholder file GONE, the state
// DB carries the dismissal), the remaining count, the exact success
// line, and the exact box.
func TestDecommissionBuryHappyPath(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)
	p1 := writeBacklogPlaceholder(t, vault, "a-placeholder.md", "https://one.example/lost")
	p2 := writeBacklogPlaceholder(t, vault, "b-placeholder.md", "https://two.example/wall")
	if out := s.DecommissionScan(); out.Box != nil {
		t.Fatalf("the scan: %+v", out.Box)
	}

	title, body := s.DecommissionBury([]string{"https://one.example/lost"})
	if title != "Decommissioned" {
		t.Fatalf("the box title: %q", title)
	}
	table := websitepipeline.DecommissionTablePath(vault)
	want := "1 link(s) marked dead — 1 enforced now, 1 placeholder(s) swept.\n\n" +
		"1 link(s) still wait in _review.\n\n" +
		"The ledger: " + table + "\n(♻️ revived on a row brings a link back to life.)"
	if body != want {
		t.Fatalf("the box body:\n got: %q\nwant: %q", body, want)
	}
	// the marking's own line + the door's report line.
	wantLog(t, rec, "🪦 1 link(s) marked dead in DECOMMISSIONED.md", "info")
	wantLog(t, rec, "🪦 Graveyard: 1 link(s) decommissioned, 1 placeholder(s) swept — 1 still waiting in _review", "success")
	// the enforcement's own evidence: the placeholder swept, the other
	// kept, the dismissal in the state DB, the row confirmed dead.
	if _, err := os.Stat(p1); !os.IsNotExist(err) {
		t.Fatal("the buried link's placeholder was swept")
	}
	if _, err := os.Stat(p2); err != nil {
		t.Fatal(" the waiting link's placeholder stays")
	}
	state := decomState(t)
	if hit, _ := state.IsDismissed("https://one.example/lost"); !hit {
		t.Fatal("the burial's dismissal lives in the state DB")
	}
	if hit, _ := state.IsDismissed("https://two.example/wall"); hit {
		t.Fatal("the waiting link was never dismissed")
	}
	state.Close()
	rows := websitepipeline.ScanDecommissionTable(vault)
	if st := rows["https://one.example/lost"]; !decomtable.StatusIsDead(st) || !strings.Contains(strings.ToLower(st), "confirmed") {
		t.Fatalf("the row's dead-confirm stamp: %q", st)
	}
	if st := rows["https://two.example/wall"]; !strings.Contains(strings.ToLower(st), "unreviewed") {
		t.Fatalf("the waiting row keeps its candidate status: %q", st)
	}
}

// TestDecommissionBuryClearBacklog pins the cleared variants (the
// line's tail + the box's tail when nothing waits anymore).
func TestDecommissionBuryClearBacklog(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)
	writeBacklogPlaceholder(t, vault, "a-placeholder.md", "https://one.example/lost")
	writeBacklogPlaceholder(t, vault, "b-placeholder.md", "https://two.example/wall")
	if out := s.DecommissionScan(); out.Box != nil {
		t.Fatalf("the scan: %+v", out.Box)
	}
	title, body := s.DecommissionBury([]string{
		"https://one.example/lost", "https://two.example/wall"})
	if title != "Decommissioned" {
		t.Fatalf("the box title: %q", title)
	}
	table := websitepipeline.DecommissionTablePath(vault)
	want := "2 link(s) marked dead — 2 enforced now, 2 placeholder(s) swept.\n\n" +
		"The _review backlog is clear — no more notices.\n\n" +
		"The ledger: " + table + "\n(♻️ revived on a row brings a link back to life.)"
	if body != want {
		t.Fatalf("the clear-backlog box body:\n got: %q\nwant: %q", body, want)
	}
	wantLog(t, rec, "🪦 Graveyard: 2 link(s) decommissioned, 2 placeholder(s) swept — the _review backlog is clear", "success")
}

// TestDecommissionBuryEnforcementDeferred pins the deferral law: the
// state DB unopenable (a directory squatting on cache.db) → the ⚠ line
// (warning), the table STILL marked dead (the marking is its own
// step), and the box honestly says the next batch enforces the
// burials.
func TestDecommissionBuryEnforcementDeferred(t *testing.T) {
	s, rec, vault := newDecomDoorRig(t)
	writeBacklogPlaceholder(t, vault, "a-placeholder.md", "https://one.example/lost")
	if out := s.DecommissionScan(); out.Box != nil {
		t.Fatalf("the scan: %+v", out.Box)
	}
	// squat a directory on the cache.db path AFTER the scan (the scan's
	// own probe already ran clean — the file it created goes first).
	if err := os.Remove(cacheDBPathOf(nil)); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(constants.AppDir(), "cache.db"), 0o755); err != nil {
		t.Fatal(err)
	}
	title, body := s.DecommissionBury([]string{"https://one.example/lost"})
	if title != "Decommissioned" {
		t.Fatalf("the box title: %q", title)
	}
	table := websitepipeline.DecommissionTablePath(vault)
	// the reference's own arithmetic: the marking already buried the
	// row in the table, so the backlog-alive scan filters it — the box
	// says clear even while the placeholder file waits for the next
	// batch's sweep (the table is the truth).
	want := "1 link(s) marked dead — the next batch enforces the burials.\n\n" +
		"The _review backlog is clear — no more notices.\n\n" +
		"The ledger: " + table + "\n(♻️ revived on a row brings a link back to life.)"
	if body != want {
		t.Fatalf("the deferred box body:\n got: %q\nwant: %q", body, want)
	}
	wantLog(t, rec, "⚠️ Graveyard enforcement deferred to the next batch:", "warning")
	wantLog(t, rec, "🪦 1 link(s) marked dead in DECOMMISSIONED.md", "info")
	// the deferred report's own line (0 enforced — the table's truth).
	wantLog(t, rec, "🪦 Graveyard: 0 link(s) decommissioned — the _review backlog is clear", "success")
	// the table still carries the burial (the next batch enforces it).
	rows := websitepipeline.ScanDecommissionTable(vault)
	if st := rows["https://one.example/lost"]; !decomtable.StatusIsDead(st) {
		t.Fatalf("the marking survived the enforcement failure: %q", st)
	}
}

// TestDecommissionBuryReviveEnforced pins the one-ledger law's other
// half: a ♻️ revived row (the owner's hand-edit in the table) is
// enforced by the SAME consume pass the door runs — the burial's
// dismissal is gone, the link is fetched like new.
func TestDecommissionBuryReviveEnforced(t *testing.T) {
	s, _, vault := newDecomDoorRig(t)
	writeBacklogPlaceholder(t, vault, "a-placeholder.md", "https://one.example/lost")
	writeBacklogPlaceholder(t, vault, "b-placeholder.md", "https://two.example/wall")
	if out := s.DecommissionScan(); out.Box != nil {
		t.Fatalf("the scan: %+v", out.Box)
	}
	if title, _ := s.DecommissionBury([]string{"https://one.example/lost"}); title != "Decommissioned" {
		t.Fatal("the first burial")
	}
	state := decomState(t)
	if hit, _ := state.IsDismissed("https://one.example/lost"); !hit {
		t.Fatal("the burial landed")
	}
	state.Close()
	// the owner's hand-edit: the row flips to ♻️ revived in the table.
	flipTableRowStatus(t, vault, "https://one.example/lost", "♻️ revived — back to life")
	// any later enforcement (the next bury's own consume) enforces it.
	if title, _ := s.DecommissionBury([]string{"https://two.example/wall"}); title != "Decommissioned" {
		t.Fatal("the second burial")
	}
	state = decomState(t)
	if hit, _ := state.IsDismissed("https://one.example/lost"); hit {
		t.Fatal("the revival was enforced — the dismissal is gone")
	}
	if hit, _ := state.IsDismissed("https://two.example/wall"); !hit {
		t.Fatal("the second burial landed")
	}
	state.Close()
}

// TestDecommissionOpenTable pins the reveal door: the table revealed
// through the OpenFileLocation seam (the _review dir created when
// missing — the sibling door's law), the nil seam answering false.
func TestDecommissionOpenTable(t *testing.T) {
	s, _, vault := newDecomDoorRig(t)
	opened := ""
	s.Deps.OpenFileLocation = func(p string) { opened = p }
	if !s.OpenDecommissionTable() {
		t.Fatal("the reveal answers true")
	}
	if want := websitepipeline.DecommissionTablePath(vault); opened != want {
		t.Fatalf("the reveal target: %q (want %q)", opened, want)
	}
	if _, err := os.Stat(filepath.Join(vault, "_review")); err != nil {
		t.Fatal("the _review dir was created for the reveal")
	}
	s.Deps.OpenFileLocation = nil
	if s.OpenDecommissionTable() {
		t.Fatal("the nil seam answers false")
	}
	// an empty vault path answers false too (the unconfigured desk).
	s.cfg.Set("website_vault_path", "")
	if s.OpenDecommissionTable() {
		t.Fatal("no reveal without a vault")
	}
}

// TestDecommissionDoorPayloadFieldTags pins THE FIELD-TAG LAW: the
// payload's JSON tags mirror the keys the frontend reads.
func TestDecommissionDoorPayloadFieldTags(t *testing.T) {
	b, err := json.Marshal(&DecommissionScan{
		Box:   &HeroBox{Title: "t", Body: "b", Question: true},
		Links: []string{"https://one.example/"},
		Table: "/tmp/T.md",
	})
	if err != nil {
		t.Fatal(err)
	}
	got := string(b)
	for _, want := range []string{
		`"box":{"title":"t","body":"b","question":true}`,
		`"links":["https://one.example/"]`,
		`"table":"/tmp/T.md"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("the field tags: %s missing %s", got, want)
		}
	}
	// the empty scan marshals the keys the frontend branches on.
	b, err = json.Marshal(&DecommissionScan{})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{`"box":null`, `"links":null`, `"table":""`} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("the empty payload: %s missing %s", string(b), want)
		}
	}
}
