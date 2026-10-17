// repobanish_test.go — T-46: the ported selections of the reference's
// own test_repobanish suite — the gate's eyes (the mark grammar, the
// trash-folder move, the master-table gestures, the dismissed gate, the
// trailing-slash twin), the burial (every never-fetch gate, the
// handwritten-kept sacred law, the table-gesture doors, the leftover
// quarantine subsumption, the dry-run rehearsal), THE ♻️ DOOR BACK
// (revive clears dismissal + settlement, idempotent, websites rows
// ignored), the pass (the count spoken before the ask, the verdict
// gates, the config opt-out, the ungated revival, the broken channel /
// broken scan honesty), and the never-fetch promise (the batch loop's
// own loads).
//
// The CacheDB helper legs already live in internal/store's suites
// (processed_row_for found/normalized/missing, forget_processed_url
// true-then-false — repos_test.go); the no-op-settle locking pin ports
// here (the test_repobanish-specific regression). The source-contract
// legs (the GUI worker's wiring, the hermetic-law source pin,
// build_note's teaching line) are the Python repo's own CI facts —
// not portable (the T-44/T-45 precedents; the note-teaching line rides
// notebuilder's port).
package repobanish

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// --- the hermetic fixtures (the reference suite's shapes) -------------------

// writeRepoNote mirrors the reference's _repo_note: hand-write a repo
// note the way build_note shapes it (the frontmatter the banishment
// grammar reads: source + managed_by + tags).
func writeRepoNote(t *testing.T, vault, rel, url string, opts ...repoNoteOpt) string {
	t.Helper()
	n := repoNoteSpec{managed: "gitcurator", tags: "[cli]", body: ""}
	for _, o := range opts {
		o(&n)
	}
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	body := n.body
	if body == "" {
		body = "A curated repo note. " + strings.Repeat("x", 200)
	}
	name := url
	if i := strings.LastIndex(url, "/"); i >= 0 {
		name = url[i+1:]
	}
	text := fmt.Sprintf(`---
source: %s
aliases: []
tags: %s
category: "Agents"
date_processed: 2026-10-09
managed_by: "%s"
%s---
# %s

%s

---
*Source: [GitHub](%s)*
`, url, n.tags, n.managed, n.extra, name, body, url)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

type repoNoteOpt func(*repoNoteOpt2)

type repoNoteOpt2 struct {
	managed string
	tags    string
	body    string
	extra   string
}

func withManaged(m string) repoNoteOpt { return func(n *repoNoteOpt2) { n.managed = m } }
func withTags(t string) repoNoteOpt    { return func(n *repoNoteOpt2) { n.tags = t } }
func withBody(b string) repoNoteOpt    { return func(n *repoNoteOpt2) { n.body = b } }
func withExtraFM(e string) repoNoteOpt { return func(n *repoNoteOpt2) { n.extra = e } }

type repoNoteSpec = repoNoteOpt2

// markNote mirrors the reference's _mark: add the delete mark to a
// note's tags line.
func markNote(t *testing.T, path string, marker string) {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "tags: [") {
		text = strings.Replace(text, "tags: [",
			"tags: ["+marker+", ", 1)
	} else { // block style (Obsidian's property editor)
		text = strings.Replace(text, "tags:\n",
			"tags:\n  - "+marker+"\n", 1)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// writeTable mirrors the reference's _table: the vault's own
// DECOMMISSIONED.md with the given data rows.
func writeTable(t *testing.T, vault string, rows [][3]string) string {
	t.Helper()
	path := filepath.Join(vault, "_review", "DECOMMISSIONED.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	lines := []string{
		"# Review Master Table — decommission or approve", "",
		"| # | Date | URL | Domain | Source | Status | Notes |",
		"|---|------|-----|--------|--------|--------|-------|",
	}
	for _, r := range rows {
		icon, url, status := r[0], r[1], r[2]
		domain := url
		if i := strings.Index(url, "//"); i >= 0 {
			domain = url[i+2:]
			if j := strings.Index(domain, "/"); j >= 0 {
				domain = domain[:j]
			}
		}
		lines = append(lines, fmt.Sprintf(
			"| %s | 2026-10-09 | %s | %s | test | %s | |",
			icon, url, domain, status))
	}
	if err := os.WriteFile(path, []byte(strings.Join(lines, "\n")+"\n"),
		0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// recorder mirrors the reference's _Recorder log sink.
type recorder struct {
	lines [][2]string
}

func (r *recorder) log(msg, level string) {
	r.lines = append(r.lines, [2]string{level, msg})
}

func (r *recorder) joined() string {
	msgs := make([]string, len(r.lines))
	for i, e := range r.lines {
		msgs[i] = e[1]
	}
	return strings.Join(msgs, "\n")
}

// channel mirrors the reference's _Channel: the injected
// banish_confirm stand-in — records the items it was asked, answers the
// verdict the test scripts.
type channel struct {
	verdict string
	asks    [][]PendingRepoBanishment
}

func (ch *channel) fn(items []PendingRepoBanishment, log LogFn) *RepoBanishConfirmResult {
	cp := append([]PendingRepoBanishment(nil), items...)
	ch.asks = append(ch.asks, cp)
	return &RepoBanishConfirmResult{Verdict: ch.verdict}
}

// scanCase is the gate's-eyes plumbing.
type scanCase struct {
	t     *testing.T
	tmp   string
	vault string
	log   *recorder
}

func newScanCase(t *testing.T) *scanCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "GitHubVault")
	if err := os.MkdirAll(filepath.Join(vault, "AI-Domain"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &scanCase{t: t, tmp: tmp, vault: vault, log: &recorder{}}
}

func (sc *scanCase) scan() PendingRepoBanishmentsReport {
	return ScanPendingRepoBanishments(sc.vault, nil, sc.log.log)
}

// buryCase is the burial/revive/pass plumbing: the vault + the cache +
// the note state (two connections over one cache.db — the reference's
// own shape).
type buryCase struct {
	t     *testing.T
	tmp   string
	vault string
	cache *store.CacheDB
	ns    *notestate.DB
	log   *recorder
}

func newBuryCase(t *testing.T) *buryCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "GitHubVault")
	if err := os.MkdirAll(filepath.Join(vault, "AI-Domain"), 0o755); err != nil {
		t.Fatal(err)
	}
	dbPath := filepath.Join(tmp, "cache.db")
	cache, err := store.NewCacheDB(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	ns, err := notestate.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		dryrun.Disable()
		dryrun.Clear()
		cache.Close()
		ns.Close()
	})
	return &buryCase{t: t, tmp: tmp, vault: vault, cache: cache, ns: ns,
		log: &recorder{}}
}

func (bc *buryCase) bury(items []PendingRepoBanishment) BanishMarkedReposReport {
	return BanishMarkedRepos(bc.cache, bc.ns, bc.vault, bc.log.log, items)
}

// --- 1. the gate's eyes ------------------------------------------------------

func TestMarkedRepoNoteIsOnePendingItem(t *testing.T) {
	sc := newScanCase(t)
	p := writeRepoNote(t, sc.vault, "AI-Domain/agent.md",
		"https://github.com/owner/agent")
	markNote(t, p, "🗑️")
	out := sc.scan()
	if len(out.Items) != 1 {
		t.Fatalf("items %d; want 1", len(out.Items))
	}
	it := out.Items[0]
	if it.Canonical != "https://github.com/owner/agent" ||
		it.Door != "note tag" || it.Path != p {
		t.Fatalf("item %+v", it)
	}
	if out.KeptHandwritten != 0 {
		t.Fatalf("kept_handwritten %d", out.KeptHandwritten)
	}
}

func TestBodyHashtagAndFrontmatterKeyFire(t *testing.T) {
	sc := newScanCase(t)
	writeRepoNote(t, sc.vault, "AI-Domain/one.md",
		"https://github.com/owner/one", withBody("uses agents #auto-delete"))
	writeRepoNote(t, sc.vault, "AI-Domain/two.md",
		"https://github.com/owner/two", withExtraFM("decommission: true\n"))
	out := sc.scan()
	got := map[string]bool{}
	for _, it := range out.Items {
		got[it.Canonical] = true
	}
	if !got["https://github.com/owner/one"] || !got["https://github.com/owner/two"] {
		t.Fatalf("items %v", got)
	}
}

func TestBlockStyleTagsAndAutoDeleteWordFire(t *testing.T) {
	sc := newScanCase(t)
	p := writeRepoNote(t, sc.vault, "AI-Domain/block.md",
		"https://github.com/owner/block", withTags(""))
	data, err := os.ReadFile(p)
	if err != nil {
		t.Fatal(err)
	}
	text := strings.Replace(string(data), "tags:", "tags:\n  - auto_delete", 1)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	out := sc.scan()
	if len(out.Items) != 1 {
		t.Fatalf("items %d; want 1", len(out.Items))
	}
	if !strings.Contains(out.Items[0].Marker, "auto_delete") {
		t.Fatalf("marker %q", out.Items[0].Marker)
	}
}

func TestTrashFolderMoveIsTheVerdict(t *testing.T) {
	sc := newScanCase(t)
	writeRepoNote(t, sc.vault, "Trash/moved.md", "https://github.com/owner/moved")
	out := sc.scan()
	if len(out.Items) != 1 || out.Items[0].Door != "trash folder" {
		t.Fatalf("items %+v", out.Items)
	}
}

func TestHandwrittenMarkedNoteIsKeptNotAsked(t *testing.T) {
	sc := newScanCase(t)
	p := writeRepoNote(t, sc.vault, "AI-Domain/human.md",
		"https://github.com/owner/human", withManaged("human"))
	markNote(t, p, "🗑️")
	out := sc.scan()
	if len(out.Items) != 0 {
		t.Fatalf("items %v; want none", out.Items)
	}
	if out.KeptHandwritten != 1 {
		t.Fatalf("kept_handwritten %d; want 1", out.KeptHandwritten)
	}
}

func TestNonGithubSourceIsNotThisDoorsBusiness(t *testing.T) {
	sc := newScanCase(t)
	p := writeRepoNote(t, sc.vault, "AI-Domain/site.md", "https://cleanup.pictures/")
	markNote(t, p, "🗑️")
	out := sc.scan()
	// defensive: an app-owned non-repo source never fires and never
	// counts as kept
	if len(out.Items) != 0 || out.KeptHandwritten != 0 {
		t.Fatalf("items %v kept %d; want none/none", out.Items, out.KeptHandwritten)
	}
}

func TestTrailingSlashTwinIsOneAsk(t *testing.T) {
	sc := newScanCase(t)
	p1 := writeRepoNote(t, sc.vault, "AI-Domain/a.md",
		"https://github.com/owner/twin")
	p2 := writeRepoNote(t, sc.vault, "AI-Domain/b.md",
		"https://github.com/owner/twin/")
	markNote(t, p1, "🗑️")
	markNote(t, p2, "🗑️")
	out := sc.scan()
	if len(out.Items) != 1 {
		t.Fatalf("items %d; want 1 (the trailing-slash twin is ONE ask)",
			len(out.Items))
	}
	if out.Items[0].Canonical != "https://github.com/owner/twin" {
		t.Fatalf("canonical %q", out.Items[0].Canonical)
	}
}

func TestAlreadyDismissedIsNeverReAsked(t *testing.T) {
	sc := newScanCase(t)
	p := writeRepoNote(t, sc.vault, "AI-Domain/gone.md",
		"https://github.com/owner/gone")
	markNote(t, p, "🗑️")
	db, err := notestate.New(filepath.Join(sc.tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Dismiss(notestate.VaultGithub,
		"https://github.com/owner/gone", p); err != nil {
		t.Fatal(err)
	}
	out := ScanPendingRepoBanishments(sc.vault, db, sc.log.log)
	if len(out.Items) != 0 {
		t.Fatalf("items %v; want none (already dismissed)", out.Items)
	}
}

func TestNoteInTrashQuarantineNeverRefound(t *testing.T) {
	sc := newScanCase(t)
	p := writeRepoNote(t, sc.vault, ".trash/banished/resting.md",
		"https://github.com/owner/resting")
	markNote(t, p, "🗑️")
	out := sc.scan()
	if len(out.Items) != 0 {
		t.Fatalf("items %v; want none (the scan never enters .trash)", out.Items)
	}
}

func TestTableGestureIsAPendingAsk(t *testing.T) {
	sc := newScanCase(t)
	writeTable(t, sc.vault, [][3]string{
		{"🗑️", "https://github.com/owner/row", "delete"}})
	out := sc.scan()
	if len(out.Items) != 1 {
		t.Fatalf("items %d; want 1", len(out.Items))
	}
	if out.Items[0].Door != "master-table gesture" ||
		out.Items[0].Canonical != "https://github.com/owner/row" {
		t.Fatalf("item %+v", out.Items[0])
	}
}

func TestConfirmedRowIsHistoryNotAnAsk(t *testing.T) {
	sc := newScanCase(t)
	writeTable(t, sc.vault, [][3]string{
		{"🗑️", "https://github.com/owner/done",
			"🗑️ banished — confirmed 2026-10-09"}})
	out := sc.scan()
	if len(out.Items) != 0 {
		t.Fatalf("items %v; want none (history)", out.Items)
	}
}

func TestRevivedRowIsNotABanishmentAsk(t *testing.T) {
	sc := newScanCase(t)
	writeTable(t, sc.vault, [][3]string{
		{"♻️", "https://github.com/owner/back", "revived"}})
	out := sc.scan()
	if len(out.Items) != 0 {
		t.Fatalf("items %v; want none (a revival)", out.Items)
	}
}

func TestWebsitesRowInThisVaultTableIsIgnored(t *testing.T) {
	sc := newScanCase(t)
	writeTable(t, sc.vault, [][3]string{
		{"🗑️", "https://cleanup.pictures/", "delete"}})
	out := sc.scan()
	if len(out.Items) != 0 {
		t.Fatalf("items %v; want none (a websites row)", out.Items)
	}
}

func TestMissingVaultIsQuiet(t *testing.T) {
	out := ScanPendingRepoBanishments(
		filepath.Join(t.TempDir(), "nope"), nil, func(string, string) {})
	if len(out.Items) != 0 || out.KeptHandwritten != 0 {
		t.Fatalf("missing vault answered %+v", out)
	}
}

// --- 2. the burial -------------------------------------------------------------

func TestTheFullBurialEveryGateHolds(t *testing.T) {
	bc := newBuryCase(t)
	p := writeRepoNote(t, bc.vault, "AI-Domain/agent.md",
		"https://github.com/owner/agent")
	markNote(t, p, "🗑️")
	if err := bc.cache.AddProcessed(1, "https://github.com/owner/agent",
		"owner", "agent", p, "Agents"); err != nil {
		t.Fatal(err)
	}
	bc.cache.AddFailed("https://github.com/owner/agent", "boom")
	rep := bc.bury(nil)
	// the note left the library for .trash/banished
	if _, err := os.Stat(p); err == nil {
		t.Fatal("the note stayed")
	}
	if _, err := os.Stat(filepath.Join(bc.vault, ".trash", "banished",
		"agent.md")); err != nil {
		t.Fatal("the note is not in .trash/banished")
	}
	// DB first: dismissed + settled + retry resolved + row forgotten
	if ok, err := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/agent"); err != nil || !ok {
		t.Fatalf("dismissed = %v/%v", ok, err)
	}
	if !bc.cache.IsRepoSettled("https://github.com/owner/agent") {
		t.Fatal("not settled")
	}
	if got := bc.cache.GetFailedURLs(); len(got) != 0 {
		t.Fatalf("failed rows left: %v", got)
	}
	if row := bc.cache.ProcessedRowFor("https://github.com/owner/agent"); row != nil {
		t.Fatalf("processed row survived: %+v", row)
	}
	// the report + the record row
	if rep.Banished != 1 || rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	if len(rep.URLs) != 1 || rep.URLs[0] != "https://github.com/owner/agent" {
		t.Fatalf("urls %v", rep.URLs)
	}
	if rep.RowsWritten != 1 {
		t.Fatalf("rows_written %d", rep.RowsWritten)
	}
	tbl := websitepipeline.ScanDecommissionTable(bc.vault)
	if st, ok := tbl["https://github.com/owner/agent"]; !ok ||
		!strings.Contains(st, "confirmed") {
		t.Fatalf("table row %q ok=%v", st, ok)
	}
	// the burial's own summary line speaks the promise
	j := bc.log.joined()
	if !strings.Contains(j, "never fetched") || !strings.Contains(j, "row-forgotten") {
		t.Fatalf("the promise missing from the log:\n%s", j)
	}
}

func TestIdempotentSecondRunBanishesNothing(t *testing.T) {
	bc := newBuryCase(t)
	p := writeRepoNote(t, bc.vault, "AI-Domain/agent.md",
		"https://github.com/owner/agent")
	markNote(t, p, "🗑️")
	bc.bury(nil)
	rep2 := bc.bury(nil)
	if rep2.Banished != 0 || len(rep2.URLs) != 0 {
		t.Fatalf("second run banished %+v", rep2)
	}
}

func TestHandwrittenMarkedNoteIsKept(t *testing.T) {
	bc := newBuryCase(t)
	p := writeRepoNote(t, bc.vault, "AI-Domain/human.md",
		"https://github.com/owner/human", withManaged("human"))
	markNote(t, p, "🗑️")
	rep := bc.bury(nil)
	if rep.Banished != 0 {
		t.Fatalf("banished %d; want 0", rep.Banished)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("the handwritten note was touched")
	}
	if !strings.Contains(bc.log.joined(), "never deletes what it did not write") {
		t.Fatalf("the sacred law's line missing:\n%s", bc.log.joined())
	}
}

func TestTableGestureWithNoteRemovesTheNote(t *testing.T) {
	bc := newBuryCase(t)
	p := writeRepoNote(t, bc.vault, "AI-Domain/row.md",
		"https://github.com/owner/row")
	if err := bc.cache.AddProcessed(2, "https://github.com/owner/row",
		"owner", "row", p, "Agents"); err != nil {
		t.Fatal(err)
	}
	writeTable(t, bc.vault, [][3]string{
		{"🗑️", "https://github.com/owner/row", "delete"}})
	gate := ScanPendingRepoBanishments(bc.vault, bc.ns, bc.log.log)
	rep := bc.bury(gate.Items)
	if _, err := os.Stat(p); err == nil {
		t.Fatal("the row's note stayed")
	}
	if !bc.cache.IsRepoSettled("https://github.com/owner/row") {
		t.Fatal("not settled")
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/row"); !ok {
		t.Fatal("not dismissed")
	}
	if row := bc.cache.ProcessedRowFor("https://github.com/owner/row"); row != nil {
		t.Fatalf("row survived: %+v", row)
	}
	if rep.Banished != 1 {
		t.Fatalf("banished %d", rep.Banished)
	}
}

func TestTableGestureNoteAlreadyGoneIsDBOnly(t *testing.T) {
	bc := newBuryCase(t)
	// the owner deleted the file by hand and marked the row — his exact
	// workaround; the burial closes the loop DB-only
	writeTable(t, bc.vault, [][3]string{
		{"🗑️", "https://github.com/owner/gone", "delete"}})
	gate := ScanPendingRepoBanishments(bc.vault, bc.ns, bc.log.log)
	rep := bc.bury(gate.Items)
	if rep.Banished != 1 || rep.NotesMoved != 0 {
		t.Fatalf("report %+v; want banished 1, moved 0", rep)
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/gone"); !ok {
		t.Fatal("not dismissed")
	}
	if !bc.cache.IsRepoSettled("https://github.com/owner/gone") {
		t.Fatal("not settled")
	}
	if !strings.Contains(bc.log.joined(), "already gone") {
		t.Fatalf("the honest line missing:\n%s", bc.log.joined())
	}
}

func TestLeftoverQuarantineRowIsSubsumed(t *testing.T) {
	bc := newBuryCase(t)
	p := writeRepoNote(t, bc.vault, "AI-Domain/quar.md",
		"https://github.com/owner/quar")
	markNote(t, p, "🗑️")
	bc.cache.Decommission("https://github.com/owner/quar", "404")
	bc.bury(nil)
	// the banishment is the STRONGER verdict: the quarantine row is gone
	// (its ♻️ reset can never un-do this burial) and the repo stays
	// settled
	if bc.cache.IsDecommissioned("https://github.com/owner/quar") {
		t.Fatal("the quarantine row survived the burial")
	}
	if !bc.cache.IsRepoSettled("https://github.com/owner/quar") {
		t.Fatal("not settled")
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/quar"); !ok {
		t.Fatal("not dismissed")
	}
}

func TestDryRunMovesNothingAndRecordsNoDismissal(t *testing.T) {
	bc := newBuryCase(t)
	p := writeRepoNote(t, bc.vault, "AI-Domain/dry.md",
		"https://github.com/owner/dry")
	markNote(t, p, "🗑️")
	if err := bc.cache.AddProcessed(3, "https://github.com/owner/dry",
		"owner", "dry", p, "Agents"); err != nil {
		t.Fatal(err)
	}
	dryrun.Enable()
	rep := bc.bury(nil)
	dryrun.Disable()
	dryrun.Clear()
	if _, err := os.Stat(p); err != nil {
		t.Fatal("the note moved in a rehearsal")
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/dry"); ok {
		t.Fatal("note_state written in a rehearsal")
	}
	if rep.Banished != 1 {
		t.Fatalf("the rehearsal counted %d; want 1 (honest counts)", rep.Banished)
	}
	if !strings.Contains(bc.log.joined(), "rehearsal") {
		t.Fatalf("the rehearsal line missing:\n%s", bc.log.joined())
	}
}

func TestMissingVaultBurialIsQuiet(t *testing.T) {
	bc := newBuryCase(t)
	rep := BanishMarkedRepos(bc.cache, bc.ns,
		filepath.Join(bc.tmp, "nope"), bc.log.log, nil)
	if rep.Banished != 0 {
		t.Fatalf("missing vault banished %d", rep.Banished)
	}
}

// --- 3. THE ♻️ DOOR BACK --------------------------------------------------------

func TestReviveClearsDismissalAndSettlement(t *testing.T) {
	bc := newBuryCase(t)
	url := "https://github.com/owner/back"
	if err := bc.ns.Dismiss(notestate.VaultGithub, url, "note.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := bc.cache.SettleRepos([]string{url}); err != nil {
		t.Fatal(err)
	}
	writeTable(t, bc.vault, [][3]string{{"♻️", url, "revived"}})
	n := ReviveBanishedRepos(bc.cache, bc.ns, bc.vault, bc.log.log)
	if n != 1 {
		t.Fatalf("revived %d; want 1", n)
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub, url); ok {
		t.Fatal("the dismissal survived")
	}
	if bc.cache.IsRepoSettled(url) {
		t.Fatal("the settlement survived")
	}
	if !strings.Contains(bc.log.joined(), "fetched like new") {
		t.Fatalf("the revival line missing:\n%s", bc.log.joined())
	}
}

func TestReviveIsIdempotent(t *testing.T) {
	bc := newBuryCase(t)
	url := "https://github.com/owner/back"
	if err := bc.ns.Dismiss(notestate.VaultGithub, url, "note.md"); err != nil {
		t.Fatal(err)
	}
	if _, err := bc.cache.SettleRepos([]string{url}); err != nil {
		t.Fatal(err)
	}
	writeTable(t, bc.vault, [][3]string{{"♻️", url, "revived"}})
	ReviveBanishedRepos(bc.cache, bc.ns, bc.vault, bc.log.log)
	n2 := ReviveBanishedRepos(bc.cache, bc.ns, bc.vault, bc.log.log)
	if n2 != 0 {
		t.Fatalf("second revive = %d; want 0", n2)
	}
}

func TestWebsitesReviveRowIsIgnored(t *testing.T) {
	bc := newBuryCase(t)
	if _, err := bc.cache.SettleRepos(
		[]string{"https://cleanup.pictures/"}); err != nil {
		t.Fatal(err)
	}
	writeTable(t, bc.vault, [][3]string{
		{"♻️", "https://cleanup.pictures/", "revived"}})
	n := ReviveBanishedRepos(bc.cache, bc.ns, bc.vault, bc.log.log)
	if n != 0 {
		t.Fatalf("websites row revived %d; want 0", n)
	}
	if !bc.cache.IsRepoSettled("https://cleanup.pictures/") {
		t.Fatal("the websites settlement was touched")
	}
}

func TestNoTableReviveIsQuiet(t *testing.T) {
	bc := newBuryCase(t)
	n := ReviveBanishedRepos(bc.cache, bc.ns, bc.vault, bc.log.log)
	if n != 0 {
		t.Fatalf("no table revived %d", n)
	}
}

// --- 4. the pass (the gate's orchestration) -----------------------------------

func (bc *buryCase) runPass(factory RepoBanishConfirmFactory,
	config map[string]any) RepoBanishPassReport {
	return RunRepoBanishmentPass(bc.cache, bc.ns, bc.vault, bc.log.log,
		config, factory)
}

func (bc *buryCase) marked(url string) string {
	p := writeRepoNote(bc.t, bc.vault, "AI-Domain/agent.md", url)
	markNote(bc.t, p, "🗑️")
	return p
}

func TestCleanVaultIsAQuietNoop(t *testing.T) {
	bc := newBuryCase(t)
	built := 0
	factory := func() (RepoBanishChannelFn, error) {
		built++
		return nil, nil
	}
	out := bc.runPass(factory, nil)
	if out.Pending != 0 || out.Verdict != "auto" {
		t.Fatalf("out %+v", out)
	}
	if built != 0 {
		t.Fatalf("the channel was built %d times", built)
	}
	if len(bc.log.lines) != 0 {
		t.Fatalf("the clean run spoke:\n%s", bc.log.joined())
	}
}

func TestCountSpokenBeforeTheAsk(t *testing.T) {
	bc := newBuryCase(t)
	bc.marked("https://github.com/owner/agent")
	ch := &channel{verdict: "confirmed"}
	bc.runPass(func() (RepoBanishChannelFn, error) { return ch.fn, nil }, nil)
	j := bc.log.joined()
	if !strings.Contains(j, "1 repo note(s) marked for deletion") {
		t.Fatalf("the count line missing:\n%s", j)
	}
	if len(ch.asks) != 1 {
		t.Fatalf("asks %d; want 1", len(ch.asks))
	}
	if len(ch.asks[0]) != 1 ||
		ch.asks[0][0].URL != "https://github.com/owner/agent" ||
		ch.asks[0][0].Door != "note tag" {
		t.Fatalf("the ask carried %+v", ch.asks[0])
	}
}

func TestConfirmedBuriesAndTallies(t *testing.T) {
	bc := newBuryCase(t)
	p := bc.marked("https://github.com/owner/agent")
	out := bc.runPass(func() (RepoBanishChannelFn, error) {
		return (&channel{verdict: "confirmed"}).fn, nil
	}, nil)
	if out.Verdict != "confirmed" || out.Banished != 1 {
		t.Fatalf("out %+v", out)
	}
	if _, err := os.Stat(p); err == nil {
		t.Fatal("the note stayed")
	}
	if !strings.Contains(bc.log.joined(), "Repo run tally: 1 repo(s) removed") {
		t.Fatalf("the tally missing:\n%s", bc.log.joined())
	}
}

func TestDeclinedKeepsMarksAndNotes(t *testing.T) {
	bc := newBuryCase(t)
	p := bc.marked("https://github.com/owner/agent")
	out := bc.runPass(func() (RepoBanishChannelFn, error) {
		return (&channel{verdict: "declined"}).fn, nil
	}, nil)
	if out.Verdict != "declined" || out.Banished != 0 {
		t.Fatalf("out %+v", out)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("the note was touched after a decline")
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/agent"); ok {
		t.Fatal("dismissed after a decline")
	}
	if !strings.Contains(bc.log.joined(), "KEPT") {
		t.Fatalf("the KEPT line missing:\n%s", bc.log.joined())
	}
}

func TestTimeoutKeepsMarksAndNotes(t *testing.T) {
	bc := newBuryCase(t)
	p := bc.marked("https://github.com/owner/agent")
	out := bc.runPass(func() (RepoBanishChannelFn, error) {
		return (&channel{verdict: "timeout"}).fn, nil
	}, nil)
	if out.Verdict != "timeout" {
		t.Fatalf("verdict %q", out.Verdict)
	}
	if _, err := os.Stat(p); err != nil {
		t.Fatal("the note was touched after a timeout")
	}
}

func TestNoChannelDefersWithTheHonestWarning(t *testing.T) {
	bc := newBuryCase(t)
	bc.marked("https://github.com/owner/agent")
	out := bc.runPass(nil, nil)
	if out.Verdict != "defer" || out.Banished != 0 {
		t.Fatalf("out %+v", out)
	}
	if !strings.Contains(bc.log.joined(), "no confirmation channel") {
		t.Fatalf("the honest warning missing:\n%s", bc.log.joined())
	}
}

func TestConfigOptOutRestoresTheAutoReflex(t *testing.T) {
	bc := newBuryCase(t)
	p := bc.marked("https://github.com/owner/agent")
	out := bc.runPass(nil, map[string]any{"banish_confirm": false})
	if out.Verdict != "auto" || out.Banished != 1 {
		t.Fatalf("out %+v", out)
	}
	if _, err := os.Stat(p); err == nil {
		t.Fatal("the note stayed under the auto reflex")
	}
}

func TestRevivalRunsBeforeTheGateWhateverTheVerdict(t *testing.T) {
	bc := newBuryCase(t)
	url := "https://github.com/owner/back"
	if err := bc.ns.Dismiss(notestate.VaultGithub, url, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := bc.cache.SettleRepos([]string{url}); err != nil {
		t.Fatal(err)
	}
	writeTable(t, bc.vault, [][3]string{{"♻️", url, "revived"}})
	bc.marked("https://github.com/owner/agent")
	out := bc.runPass(func() (RepoBanishChannelFn, error) {
		return (&channel{verdict: "declined"}).fn, nil
	}, nil)
	// the revival fired even though the gate declined
	if out.Revived != 1 {
		t.Fatalf("revived %d; want 1 (the ungated ♻️ door)", out.Revived)
	}
	if bc.cache.IsRepoSettled(url) {
		t.Fatal("the revival's settlement survived")
	}
}

func TestBrokenChannelIsADeferNeverAPanic(t *testing.T) {
	bc := newBuryCase(t)
	bc.marked("https://github.com/owner/agent")
	out := bc.runPass(func() (RepoBanishChannelFn, error) {
		return nil, errors.New("boom")
	}, nil)
	if out.Verdict != "defer" {
		t.Fatalf("verdict %q", out.Verdict)
	}
	if !strings.Contains(bc.log.joined(), "channel unavailable") {
		t.Fatalf("the honest warning missing:\n%s", bc.log.joined())
	}
}

func TestBrokenScanNeverPanics(t *testing.T) {
	bc := newBuryCase(t)
	// a vault path that is a FILE, not a dir — the walk stays quiet
	fp := filepath.Join(bc.tmp, "notadir")
	if err := os.WriteFile(fp, []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	out := RunRepoBanishmentPass(bc.cache, bc.ns, fp, bc.log.log, nil, nil)
	if out.Pending != 0 {
		t.Fatalf("pending %d", out.Pending)
	}
}

// --- 5. the never-fetch promise (the loop, closed) ---------------------------

func TestBatchLoopGatesDismissedAndSettledBothHold(t *testing.T) {
	bc := newBuryCase(t)
	url := "https://github.com/owner/agent"
	p := writeRepoNote(t, bc.vault, "AI-Domain/agent.md", url)
	markNote(t, p, "🗑️")
	BanishMarkedRepos(bc.cache, bc.ns, bc.vault, bc.log.log, nil)
	// the batch loop's own loads (the exact shapes the worker reads)
	dismissed, err := bc.ns.DismissedSet(notestate.VaultGithub)
	if err != nil {
		t.Fatal(err)
	}
	settled := bc.cache.GetSettledRepoSet()
	norm := url
	if !dismissed[norm] {
		t.Fatal("the §4.4 skip does not fire")
	}
	if !settled[norm] {
		t.Fatal("the settled skip does not fire")
	}
	// the retry queue has nothing to cry about
	if got := bc.cache.GetFailedURLs(); len(got) != 0 {
		t.Fatalf("failed rows left: %v", got)
	}
}

// TestQueueDoorCountsABanishedRepoSettledNeverPending ports the
// reference's queue-door leg: the queue classification's exact ladder —
// in_vault (a fresh VaultIndex over the post-banishment vault) →
// decommissioned → settled → pending. The banished repo lands in
// SETTLED.
func TestQueueDoorCountsABanishedRepoSettledNeverPending(t *testing.T) {
	bc := newBuryCase(t)
	url := "https://github.com/owner/agent"
	p := writeRepoNote(t, bc.vault, "AI-Domain/agent.md", url)
	markNote(t, p, "🗑️")
	BanishMarkedRepos(bc.cache, bc.ns, bc.vault, bc.log.log, nil)
	vi := batch.NewVaultIndex(bc.vault, nil)
	vi.Rebuild(nil)
	if vi.HasURL(url) {
		t.Fatal("the banished note is still in the vault index")
	}
	if got := bc.cache.GetDeadURLSet(0); len(got) != 0 {
		t.Fatalf("dead rows left: %v", got)
	}
	if !bc.cache.GetSettledRepoSet()[links.NormalizeURL(url)] {
		t.Fatal("the banished repo is not settled")
	}
}

func TestReviveThenRebanishFullCycle(t *testing.T) {
	bc := newBuryCase(t)
	url := "https://github.com/owner/cycle"
	p := writeRepoNote(t, bc.vault, "AI-Domain/cycle.md", url)
	markNote(t, p, "🗑️")
	BanishMarkedRepos(bc.cache, bc.ns, bc.vault, bc.log.log, nil)
	if !bc.cache.IsRepoSettled(url) {
		t.Fatal("not settled after the burial")
	}
	// the owner revives: the note comes back (his hand), the ♻️ row
	// opens the gates
	restored := writeRepoNote(t, bc.vault, "AI-Domain/cycle.md", url)
	writeTable(t, bc.vault, [][3]string{{"♻️", url, "revived"}})
	ReviveBanishedRepos(bc.cache, bc.ns, bc.vault, bc.log.log)
	if bc.cache.IsRepoSettled(url) {
		t.Fatal("still settled after the revival")
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub, url); ok {
		t.Fatal("still dismissed after the revival")
	}
	if _, err := os.Stat(restored); err != nil {
		t.Fatal("the restored note was touched")
	}
}

// --- 6. the CacheDB helpers' regression pin ----------------------------------

// TestNoopSettleNeverLocksTheOtherConnection ports the test_repobanish
// regression pin: settle_repos commits UNCONDITIONALLY — an INSERT OR
// IGNORE that ignores (the idempotent no-op, the COMMON case) still
// opens a write transaction, and a skipped commit would leave cache.db
// write-locked for every OTHER connection (the note-state door). Two
// connections, the no-op settle, then the other connection must write.
func TestNoopSettleNeverLocksTheOtherConnection(t *testing.T) {
	bc := newBuryCase(t)
	if _, err := bc.cache.SettleRepos(
		[]string{"https://github.com/owner/agent"}); err != nil {
		t.Fatal(err)
	}
	// the no-op settle (nothing added)
	if n, err := bc.cache.SettleRepos(
		[]string{"https://github.com/owner/agent"}); err != nil || n != 0 {
		t.Fatalf("second settle = %d/%v; want 0", n, err)
	}
	// the OTHER connection writes cleanly — no "database is locked"
	if err := bc.ns.Dismiss(notestate.VaultGithub,
		"https://github.com/owner/agent", "n.md"); err != nil {
		t.Fatalf("the other connection could not write: %v", err)
	}
	if ok, _ := bc.ns.IsDismissed(notestate.VaultGithub,
		"https://github.com/owner/agent"); !ok {
		t.Fatal("the other connection's write was lost")
	}
}
