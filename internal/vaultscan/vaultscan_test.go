// vaultscan_test.go — T-45: the ported test_vaultscan, test_censusstamp
// and test_picklock selections (the reference suites' own shapes, the
// hermetic law — zero network, temp vaults, a fake LLM), plus the
// shared fixtures every vaultscan test file rides.
//
// The GUI-layer legs of the reference suites (the Scan channel's
// Telegram round trip, the ScanPlanDialog checkbox battery, the
// ask-gate shapes, the two-vault worker job, the GUI scorecard
// models) stay with the future GUI task — the decomtable layering law
// (the T-44 §9b precedent). The release-bookkeeping legs
// (VERSION/CHANGELOG/ci.yml/AGENTS.md pins) are the Python repo's own
// CI facts — not portable (the T-44 §10 precedent). The hygiene-pass
// leg (test_picklock's vault_hygiene test) rides T-46's own suite.
package vaultscan

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// scanLog is the test log sink.
type scanLog struct {
	mu    sync.Mutex
	lines []string
}

func (l *scanLog) log(msg, level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, msg)
}

func (l *scanLog) joined() string {
	l.mu.Lock()
	defer l.mu.Unlock()
	return strings.Join(l.lines, "\n")
}

func (l *scanLog) has(needle string) bool {
	return strings.Contains(l.joined(), needle)
}

// scanCase is the shared plumbing: temp vault + state db.
type scanCase struct {
	t     *testing.T
	tmp   string
	vault string
	db    *websitepipeline.WebsiteStateDB
	logs  *scanLog
}

func newScanCase(t *testing.T) *scanCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "websites")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	db, err := websitepipeline.NewWebsiteStateDB(
		filepath.Join(tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		dryrun.Disable()
		db.Close()
	})
	return &scanCase{
		t: t, tmp: tmp, vault: vault, db: db, logs: &scanLog{},
	}
}

// note mirrors the reference tests' _note fixture (the websites
// shape).
func (c *scanCase) note(rel, url string, opts ...noteOpt) string {
	c.t.Helper()
	n := noteSpec{managed: "gitcurator", fetchStatus: "full"}
	for _, o := range opts {
		o(&n)
	}
	return writeNote(c.t, c.vault, rel, url, n)
}

type noteOpt func(*noteSpec)

type noteSpec struct {
	managed      string
	tags         []string
	body         string
	category     string
	subcategory  string
	kind         string
	fetchStatus  string
	extra        string
	stars        string
	language     string
	org          string
	sourceQuoted bool
}

func withManaged(m string) noteOpt    { return func(n *noteSpec) { n.managed = m } }
func withTags(tags ...string) noteOpt { return func(n *noteSpec) { n.tags = tags } }
func withBody(b string) noteOpt       { return func(n *noteSpec) { n.body = b } }
func withCategory(cat string) noteOpt {
	return func(n *noteSpec) { n.category = cat }
}
func withSubcategory(sub string) noteOpt {
	return func(n *noteSpec) { n.subcategory = sub }
}
func withNoteKind(k string) noteOpt    { return func(n *noteSpec) { n.kind = k } }
func withFetchStatus(f string) noteOpt { return func(n *noteSpec) { n.fetchStatus = f } }
func withExtra(e string) noteOpt       { return func(n *noteSpec) { n.extra = e } }
func withStars(s string) noteOpt       { return func(n *noteSpec) { n.stars = s } }
func withLanguage(l string) noteOpt    { return func(n *noteSpec) { n.language = l } }
func withOrg(o string) noteOpt         { return func(n *noteSpec) { n.org = o } }

// writeNote writes a machine-vault note the reference fixture shape.
func writeNote(t *testing.T, vault, rel, url string, n noteSpec) string {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	tagLine := "tags: [" + strings.Join(n.tags, ", ") + "]"
	kindLine := ""
	if n.kind != "" {
		kindLine = fmt.Sprintf("kind: %q\n", n.kind)
	}
	text := fmt.Sprintf(`---
source: "%s"
%s
category: "%s"
subcategory: "%s"
%sfetch_status: "%s"
managed_by: "%s"
stars: "%s"
primary_language: "%s"
org: "%s"
%s
---

# A Note

%s
`, url, tagLine, n.category, n.subcategory, kindLine, n.fetchStatus,
		n.managed, n.stars, n.language, n.org, n.extra, n.body)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// fakeLLM mirrors the reference tests' _FakeLLM: a canned filing plan
// (or a raw string); captures the prompts it was fed. A nil planFn
// that panics mirrors the raising LLM.
type fakeLLM struct {
	plan        string
	raw         string
	calls       [][]llmclient.Message
	tasks       []string
	panics      bool
	panicsAfter int
	nCalls      int
}

func (f *fakeLLM) call(messages []llmclient.Message, task string) string {
	f.calls = append(f.calls, messages)
	f.tasks = append(f.tasks, task)
	f.nCalls++
	if f.panicsAfter > 0 && f.nCalls == f.panicsAfter {
		panic("server hiccup")
	}
	if f.panics {
		panic("the model is down")
	}
	if f.raw != "" {
		return f.raw
	}
	return f.plan
}

func newFakeLLM(plan map[string]any) *fakeLLM {
	b, _ := json.Marshal(plan)
	return &fakeLLM{plan: string(b)}
}

func newRawLLM(raw string) *fakeLLM { return &fakeLLM{raw: raw} }

// buildPlan runs BuildScanPlan with the case's log.
func (c *scanCase) buildPlan(llm LLMCallFn, opts ...planOpt) Plan {
	c.t.Helper()
	p := planCfg{}
	for _, o := range opts {
		o(&p)
	}
	cfg := map[string]any{}
	for k, v := range p.config {
		cfg[k] = v
	}
	return BuildScanPlan(c.vault, llm, c.logs.log, cfg, nil, nil,
		p.kind, nil)
}

type planCfg struct {
	config map[string]any
	kind   string
}

type planOpt func(*planCfg)

func withConfig(k string, v any) planOpt {
	return func(p *planCfg) {
		if p.config == nil {
			p.config = map[string]any{}
		}
		p.config[k] = v
	}
}

func withKind(k string) planOpt { return func(p *planCfg) { p.kind = k } }

// applyPlan runs ApplyScanPlan with the case's log (whole-plan law).
func (c *scanCase) applyPlan(plan *Plan, opts ...applyOpt) ApplyReport {
	c.t.Helper()
	a := applyCfg{}
	for _, o := range opts {
		o(&a)
	}
	return ApplyScanPlan(plan, c.vault, c.db, c.logs.log, a.sel,
		a.kind, nil, nil)
}

type applyCfg struct {
	sel  Selection
	kind string
}

type applyOpt func(*applyCfg)

func withSelection(s Selection) applyOpt { return func(a *applyCfg) { a.sel = s } }
func withApplyKind(k string) applyOpt    { return func(a *applyCfg) { a.kind = k } }

func mustBeDir(t *testing.T, path string) {
	t.Helper()
	if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
		t.Fatalf("expected dir %s", path)
	}
}

func mustBeFile(t *testing.T, path string) {
	t.Helper()
	if fi, err := os.Stat(path); err != nil || fi.IsDir() {
		t.Fatalf("expected file %s", path)
	}
}

func mustBeGone(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("expected %s gone", path)
	}
}

func mustExist(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected %s to exist", path)
	}
}

func readFileT(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// --- TestInventory (the reference suite's own class) -------------------

func TestTheWalkCountsFoldersAndNotes(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/A.md", "https://a.example/x", withCategory("Agents"))
	c.note("AI-Domain/Agents/B.md", "https://b.example/y", withCategory("Agents"))
	c.note("Tools/Scraping/C.md", "https://c.example/z")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	if inv.TotalNotes != 3 {
		t.Fatalf("total %d != 3", inv.TotalNotes)
	}
	byRel := map[string]int{}
	for _, f := range inv.Folders {
		byRel[f.Rel] = f.NoteCount
	}
	if byRel["AI-Domain/Agents"] != 2 || byRel["Tools/Scraping"] != 1 {
		t.Fatalf("folders %v", byRel)
	}
	var a NoteRow
	found := false
	for _, n := range inv.Notes {
		if n.File == "A.md" {
			a = n
			found = true
		}
	}
	if !found {
		t.Fatal("A.md missing")
	}
	if a.Domain != "a.example" || a.Category != "Agents" || !a.AppOwned {
		t.Fatalf("row %+v", a)
	}
}

func TestProtectedFoldersAreSeenNeverCandidates(t *testing.T) {
	c := newScanCase(t)
	c.note("_review/placeholder.md", "https://r.example/p")
	c.note("_moc/index.md", "https://r.example/i")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	if inv.TotalNotes != 2 {
		t.Fatalf("total %d != 2", inv.TotalNotes)
	}
	for _, n := range inv.Notes {
		if !n.Protected {
			t.Fatalf("note %s not protected", n.File)
		}
	}
	if cands := MoveCandidates(&inv, 40); len(cands) != 0 {
		t.Fatalf("candidates %v", cands)
	}
}

func TestHandNotesAreCountedNeverMoved(t *testing.T) {
	c := newScanCase(t)
	c.note("Uncategorized/mine.md", "", withManaged("a-human"))
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	if inv.HandNotes != 1 || inv.UncategorizedNotes != 1 {
		t.Fatalf("counts %+v", inv)
	}
	if cands := MoveCandidates(&inv, 40); len(cands) != 0 {
		t.Fatalf("candidates %v", cands)
	}
}

func TestDotFoldersAndInboxAreInvisible(t *testing.T) {
	c := newScanCase(t)
	c.note(".obsidian/x.md", "https://d.example/x")
	c.note("_inbox/y.md", "https://d.example/y")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	if inv.TotalNotes != 0 {
		t.Fatalf("total %d != 0", inv.TotalNotes)
	}
}

// --- TestMoveCandidates -----------------------------------------------

func TestRootAndUncategorizedAndBroad(t *testing.T) {
	c := newScanCase(t)
	c.note("orphan.md", "https://o.example/1")
	c.note("Uncategorized/parked.md", "https://u.example/2")
	for i := 0; i < 3; i++ {
		c.note(fmt.Sprintf("AI-Domain/LLM-Tools/n%d.md", i),
			fmt.Sprintf("https://t.example/%d", i))
	}
	c.note("Tools/Scraping/small.md", "https://s.example/3")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	want := map[string]bool{"orphan.md": true, "parked.md": true,
		"n0.md": true, "n1.md": true, "n2.md": true}
	got := map[string]bool{}
	for _, n := range MoveCandidates(&inv, 2) {
		got[n.File] = true
	}
	if len(got) != len(want) {
		t.Fatalf("got %v want %v", got, want)
	}
	for k := range want {
		if !got[k] {
			t.Fatalf("missing %s in %v", k, got)
		}
	}
	got3 := map[string]bool{}
	for _, n := range MoveCandidates(&inv, 3) {
		got3[n.File] = true
	}
	if len(got3) != 2 || !got3["orphan.md"] || !got3["parked.md"] {
		t.Fatalf("threshold 3: %v", got3)
	}
}

// --- TestTheGrammarReuse ----------------------------------------------

func TestDeletionsAreTheBanishGateItems(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/fm.md", "https://g.example/fm", withTags("auto_delete"))
	c.note("Design/body.md", "https://g.example/body",
		withBody("not needed #auto-delete"))
	c.note("Design/emoji.md", "https://g.example/emoji", withTags("🗑️"))
	p := c.note("Design/key.md", "https://g.example/key", withCategory("x"))
	text := readFileT(t, p)
	text = strings.Replace(text, `category: "x"`,
		"category: \"x\"\nblacklist: true", 1)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	c.note("Design/negative.md", "https://g.example/neg",
		withBody("#deleted-files not a mark"))
	table := filepath.Join(c.vault, "_review", "DECOMMISSIONED.md")
	if err := os.MkdirAll(filepath.Dir(table), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(table, []byte(
		"| # | Date | URL | Domain | Source | Status | Notes |\n"+
			"|---|------|-----|--------|--------|--------|-------|\n"+
			"| 🗑️ | 2026-10-09 | https://g.example/table | g.example"+
			" | test | 🗑️ banished | |"), 0o644); err != nil {
		t.Fatal(err)
	}
	gate := websitepipeline.ScanPendingBanishments(c.vault, c.logs.log, nil)
	plan := c.buildPlan(nil)
	if len(plan.Deletions) != len(gate.Items) {
		t.Fatalf("plan deletions %d != gate items %d",
			len(plan.Deletions), len(gate.Items))
	}
	if len(plan.Deletions) != 5 {
		t.Fatalf("deletions %d != 5 (the negative control never rides)",
			len(plan.Deletions))
	}
	markers := map[string]bool{}
	for _, d := range plan.Deletions {
		markers[d.Marker] = true
	}
	for _, m := range []string{"auto_delete", "#auto-delete", "🗑️",
		"blacklist: true"} {
		if !markers[m] {
			t.Fatalf("marker %q missing", m)
		}
	}
}

// --- TestTheLLMPass ---------------------------------------------------

func inventoryWithCandidates(c *scanCase) *Inventory {
	c.note("Uncategorized/Agent One.md", "https://a.example/one",
		withTags("agents"))
	c.note("Uncategorized/Agent Two.md", "https://t.example/two",
		withTags("agents"))
	c.note("AI-Domain/Agents/Existing.md", "https://b.example/existing")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	return &inv
}

func TestAValidProposalLands(t *testing.T) {
	c := newScanCase(t)
	inventoryWithCandidates(c)
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"AI-Domain/Agents/Agent-Frameworks"},
		"moves": []any{
			map[string]any{"note": "Agent One.md",
				"to":     "AI-Domain/Agents/Agent-Frameworks",
				"reason": "an agent framework"},
			map[string]any{"note": "Agent Two.md",
				"to":     "AI-Domain/Agents",
				"reason": "agent tooling"},
		},
		"summary": "Two agent notes filed.",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 2 {
		t.Fatalf("moves %d != 2", len(plan.Moves))
	}
	if len(plan.NewFolders) != 1 ||
		plan.NewFolders[0] != "AI-Domain/Agents/Agent-Frameworks" {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	if plan.Summary != "Two agent notes filed." {
		t.Fatalf("summary %q", plan.Summary)
	}
	if len(llm.calls) != 1 || llm.tasks[0] != "vaultscan" {
		t.Fatalf("calls %d tasks %v", len(llm.calls), llm.tasks)
	}
	body := llm.calls[0][1].Content
	for _, needle := range []string{"Uncategorized", "Agent One", "notes"} {
		if !strings.Contains(body, needle) {
			t.Fatalf("user message missing %q", needle)
		}
	}
}

func TestUnlistedNotesAndIllegalDestinationsDrop(t *testing.T) {
	c := newScanCase(t)
	inventoryWithCandidates(c)
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"../escape", "AI-Domain/Agents/New-Sub",
			"Tools/No-Parent"},
		"moves": []any{
			map[string]any{"note": "Existing.md",
				"to": "AI-Domain/Agents/New-Sub", "reason": "no"},
			map[string]any{"note": "Not Listed.md",
				"to": "AI-Domain/Agents", "reason": "no"},
			map[string]any{"note": "Agent One.md", "to": "_review",
				"reason": "never"},
			map[string]any{"note": "Agent One.md",
				"to": "AI-Domain/Agents/New-Sub", "reason": "yes"},
			map[string]any{"note": "Agent Two.md",
				"to": "AI-Domain/Agents", "reason": "yes"},
		},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	var notes []string
	for _, m := range plan.Moves {
		notes = append(notes, m.Note)
	}
	sort.Strings(notes)
	if len(notes) != 2 || notes[0] != "Agent One.md" ||
		notes[1] != "Agent Two.md" {
		t.Fatalf("moves %v", notes)
	}
	if len(plan.NewFolders) != 1 ||
		plan.NewFolders[0] != "AI-Domain/Agents/New-Sub" {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	if !c.logs.has("not a legal destination") {
		t.Fatal("missing 'not a legal destination' log")
	}
	if !c.logs.has("no parent in the tree") {
		t.Fatal("missing 'no parent in the tree' log")
	}
}

func TestABrokenAnswerIsAnEmptyPlanNeverACrash(t *testing.T) {
	c := newScanCase(t)
	inventoryWithCandidates(c)
	plan := c.buildPlan(newRawLLM("total garbage {").call)
	if len(plan.Moves) != 0 || len(plan.NewFolders) != 0 {
		t.Fatalf("plan %+v", plan)
	}
	if !c.logs.has("not JSON I could read") {
		t.Fatal("missing JSON warning")
	}
	// an LLM that RAISES is survived too:
	boom := &fakeLLM{panics: true}
	plan2 := c.buildPlan(boom.call)
	if len(plan2.Moves) != 0 {
		t.Fatalf("moves %v", plan2.Moves)
	}
	if !c.logs.has("LLM call failed") {
		t.Fatal("missing 'LLM call failed' log")
	}
}

func TestNoCandidatesMeanNoLLMCall(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Assets/Fine.md", "https://f.example/ok")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "Fine.md", "to": "X"}},
	})
	plan := c.buildPlan(llm.call)
	if len(llm.calls) != 0 {
		t.Fatalf("the LLM was asked %d times", len(llm.calls))
	}
	if len(plan.Moves) != 0 {
		t.Fatalf("moves %v", plan.Moves)
	}
}

// --- TestTheFold ------------------------------------------------------

func TestCaseAndSeparatorsFoldAway(t *testing.T) {
	if foldPath("Design/Color_Tools") != foldPath("design/Color Tools") {
		t.Fatal("fold mismatch")
	}
	if foldPath("AI-Domain/Agents") != foldPath("ai domain/AGENTS") {
		t.Fatal("fold mismatch")
	}
	if foldSegment("Color_Tools") != "colortools" ||
		foldSegment("color tools") != "colortools" {
		t.Fatal("segment mismatch")
	}
}

func TestSegmentsNeverCollideAcrossLevels(t *testing.T) {
	if foldPath("A/B") == foldPath("A_B") {
		t.Fatal("A/B collides with A_B")
	}
}

// --- TestTheThreeLaws -------------------------------------------------

func TestTheDirectoryIsCoreAndNeverACandidate(t *testing.T) {
	c := newScanCase(t)
	c.note("000 📚 Website Directory.md", "", withNoteKind("directory"))
	c.note("orphan.md", "https://o.example/1")
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	if inv.CoreNotes != 1 {
		t.Fatalf("core %d", inv.CoreNotes)
	}
	var row NoteRow
	for _, n := range inv.Notes {
		if n.File == "000 📚 Website Directory.md" {
			row = n
		}
	}
	if !row.Core || !row.AppOwned {
		t.Fatalf("row %+v", row)
	}
	var files []string
	for _, n := range MoveCandidates(&inv, 40) {
		files = append(files, n.File)
	}
	if len(files) != 1 || files[0] != "orphan.md" {
		t.Fatalf("candidates %v", files)
	}
	if !c.logs.has("never re-filed") {
		t.Fatal("missing core log")
	}
}

func TestACoreNoteByKindIsImmuneWhereverItSits(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/catalog.md", "https://c.example/x", withNoteKind("directory"))
	c.note("Design/normal.md", "https://c.example/y")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	byFile := map[string]NoteRow{}
	for _, n := range inv.Notes {
		byFile[n.File] = n
	}
	if !byFile["catalog.md"].Core || byFile["normal.md"].Core {
		t.Fatalf("core flags wrong")
	}
	var files []string
	for _, n := range MoveCandidates(&inv, 40) {
		files = append(files, n.File)
	}
	if len(files) != 1 || files[0] != "normal.md" {
		t.Fatalf("candidates %v", files)
	}
}

func TestTheLLMCannotMoveTheDirectoryEvenWhenAsked(t *testing.T) {
	c := newScanCase(t)
	c.note("000 📚 Website Directory.md", "", withNoteKind("directory"))
	c.note("orphan.md", "https://o.example/1")
	c.note("Knowledge_Research_Reference/ref.md", "https://k.example/ref")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Knowledge_Research_Reference/Indexes"},
		"moves": []any{
			map[string]any{"note": "000 📚 Website Directory.md",
				"to":     "Knowledge_Research_Reference/Indexes",
				"reason": "it is an index"},
			map[string]any{"note": "orphan.md",
				"to":     "Knowledge_Research_Reference/Indexes",
				"reason": "research"},
		},
		"summary": "file both",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 1 || plan.Moves[0].Note != "orphan.md" {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestStraysAtACategoryRootAreCandidates(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/stray.md", "https://s.example/1")
	c.note("Design/Color_Tools/filed.md", "https://s.example/2")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	cands := MoveCandidates(&inv, 40)
	if len(cands) != 1 || cands[0].File != "stray.md" {
		t.Fatalf("candidates %+v", cands)
	}
}

func TestABareCategoryDestinationIsDropped(t *testing.T) {
	c := newScanCase(t)
	c.note("orphan.md", "https://o.example/1")
	c.note("Knowledge_Research_Reference/ref.md", "https://k.example/ref")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{map[string]any{"note": "orphan.md",
			"to": "Knowledge_Research_Reference", "reason": "research"}},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 0 {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if !c.logs.has("bare category root") {
		t.Fatal("missing bare-root log")
	}
}

func TestALoneNewFolderIsDropped(t *testing.T) {
	c := newScanCase(t)
	c.note("orphan.md", "https://o.example/1")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Knowledge_Research_Reference"},
		"moves":       []any{}, "summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.NewFolders) != 0 {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	if !c.logs.has("bare root cell") {
		t.Fatal("missing bare-root-cell log")
	}
}

func TestNearDuplicatesAreAbsorbedIntoTheExisting(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/ColorTools/Coolors.md", "https://c.example/1")
	c.note("orphan.md", "https://o.example/1")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Design/Color_Tools"},
		"moves": []any{map[string]any{"note": "orphan.md",
			"to": "Design/Color_Tools", "reason": "color tooling"}},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.NewFolders) != 0 {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	if len(plan.Moves) != 1 || plan.Moves[0].To != "Design/ColorTools" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if !c.logs.has("already exists as") {
		t.Fatal("missing absorption log")
	}
}

func TestANearIdenticalSiblingIsAbsorbed(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Color_Palettes/Hunt.md", "https://p.example/1")
	c.note("orphan.md", "https://o.example/1")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Design/Palettes"},
		"moves": []any{map[string]any{"note": "orphan.md",
			"to": "Design/Palettes", "reason": "palettes"}},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.NewFolders) != 0 {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	if len(plan.Moves) != 1 || plan.Moves[0].To != "Design/Color_Palettes" {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestAnExactDuplicateOfAnExistingFolderIsAbsorbed(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Color_Tools/Hunt.md", "https://p.example/1")
	c.note("orphan.md", "https://o.example/1")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Design/Color_Tools"},
		"moves": []any{map[string]any{"note": "orphan.md",
			"to": "Design/Color_Tools", "reason": "color tooling"}},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.NewFolders) != 0 {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	if len(plan.Moves) != 1 || plan.Moves[0].To != "Design/Color_Tools" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if !c.logs.has("already exists as") {
		t.Fatal("missing absorption log")
	}
}

func TestAPlanMayNotDuplicateItself(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Somewhere/a.md", "https://d.example/1")
	c.note("orphan.md", "https://o.example/1")
	c.note("second.md", "https://s.example/2")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Design/Color_Tools", "Design/color tools"},
		"moves": []any{
			map[string]any{"note": "orphan.md", "to": "Design/Color_Tools",
				"reason": ""},
			map[string]any{"note": "second.md", "to": "Design/color tools",
				"reason": ""},
		},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.NewFolders) != 1 || plan.NewFolders[0] != "Design/Color_Tools" {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
	var tos []string
	for _, m := range plan.Moves {
		tos = append(tos, m.To)
	}
	sort.Strings(tos)
	if len(tos) != 2 || tos[0] != "Design/Color_Tools" ||
		tos[1] != "Design/Color_Tools" {
		t.Fatalf("tos %v", tos)
	}
	if !c.logs.has("duplicates") {
		t.Fatal("missing duplicates log")
	}
}

func TestShortLeavesNeverFalseHitByContainment(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Web_Workers/a.md", "https://w.example/1")
	c.note("orphan.md", "https://o.example/1")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Design/Web"},
		"moves": []any{map[string]any{"note": "orphan.md",
			"to": "Design/Web", "reason": ""}},
		"summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.NewFolders) != 1 || plan.NewFolders[0] != "Design/Web" {
		t.Fatalf("new folders %v (Web must be kept)", plan.NewFolders)
	}
}

func TestTheLLMHearsTheLawsAtThePointOfDecision(t *testing.T) {
	c := newScanCase(t)
	c.note("orphan.md", "https://o.example/1")
	llm := newFakeLLM(map[string]any{"new_folders": []any{},
		"moves": []any{}, "summary": "nothing"})
	c.buildPlan(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
	body := llm.calls[0][1].Content
	for _, needle := range []string{"THE LATEST STATE",
		"Category/Subfolder", "Website Directory"} {
		if !strings.Contains(body, needle) {
			t.Fatalf("user message missing %q", needle)
		}
	}
}

// --- TestTheApplyPass -------------------------------------------------

func applyCasePlan(c *scanCase) *Plan {
	c.note("Uncategorized/Mover.md", "https://m.example/one",
		withCategory("Uncategorized"))
	return &Plan{
		NewFolders: []string{"AI-Domain/Agents"},
		Moves: []MoveRow{{
			Note: "Mover.md",
			Path: filepath.Join(c.vault, "Uncategorized", "Mover.md"),
			From: "Uncategorized", To: "AI-Domain/Agents",
			Reason: "agents",
		}},
	}
}

func TestMovesAreByteIdenticalAndStateRepointed(t *testing.T) {
	c := newScanCase(t)
	plan := applyCasePlan(c)
	src := filepath.Join(c.vault, "Uncategorized", "Mover.md")
	before := []byte("---\nsource: \"https://m.example/one\"\n---\nBYTES ⚙️\n")
	if err := os.WriteFile(src, before, 0o644); err != nil {
		t.Fatal(err)
	}
	rep := c.applyPlan(plan)
	if rep.FoldersCreated != 1 || rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	dst := filepath.Join(c.vault, "AI-Domain", "Agents", "Mover.md")
	mustBeFile(t, dst)
	mustBeGone(t, src)
	after := readFileT(t, dst)
	// v0.78.0 — the move is byte-identical; the stamp is the ONE
	// sanctioned frontmatter edit AFTER the move.
	if !strings.HasPrefix(after, "---\nsource: \"https://m.example/one\"\n") {
		t.Fatalf("head changed: %q", after)
	}
	if !strings.Contains(after, "\ncurated: true\n") {
		t.Fatalf("no stamp: %q", after)
	}
	stripped := strings.Replace(after, "\ncurated: true\n", "\n", 1)
	if stripped != string(before) {
		t.Fatalf("body changed:\n%q\n%q", stripped, string(before))
	}
	if rep.Stamped != 1 {
		t.Fatalf("stamped %d", rep.Stamped)
	}
	if !c.logs.has("stamped curated at its new home") {
		t.Fatal("missing stamp log")
	}
	row, err := c.db.ProcessedRow(
		links.NormalizeWebsiteURL("https://m.example/one"))
	if err != nil || row == nil {
		t.Fatalf("row %v %v", row, err)
	}
	if row.NotePath == nil || *row.NotePath != dst ||
		row.Category == nil || *row.Category != "AI-Domain" {
		t.Fatalf("row %+v", row)
	}
	if !c.logs.has("moved → AI-Domain/Agents") {
		t.Fatal("missing move log")
	}
}

func TestNothingIsEverOverwritten(t *testing.T) {
	c := newScanCase(t)
	plan := applyCasePlan(c)
	dst := filepath.Join(c.vault, "AI-Domain", "Agents", "Mover.md")
	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dst, []byte("THE OWNER'S OWN NOTE"), 0o644); err != nil {
		t.Fatal(err)
	}
	rep := c.applyPlan(plan)
	if rep.NotesMoved != 0 || rep.FailedMoves != 1 {
		t.Fatalf("report %+v", rep)
	}
	if readFileT(t, dst) != "THE OWNER'S OWN NOTE" {
		t.Fatal("the resident was touched")
	}
	mustExist(t, filepath.Join(c.vault, "Uncategorized", "Mover.md"))
	if !c.logs.has("nothing is overwritten") {
		t.Fatal("missing overwrite log")
	}
}

func TestDryRunRehearsesMoves(t *testing.T) {
	c := newScanCase(t)
	plan := applyCasePlan(c)
	dryrun.Enable()
	rep := c.applyPlan(plan)
	dryrun.Disable()
	if rep.NotesMoved != 1 || rep.FoldersCreated != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustExist(t, filepath.Join(c.vault, "Uncategorized", "Mover.md"))
	mustBeGone(t, filepath.Join(c.vault, "AI-Domain", "Agents", "Mover.md"))
	if !c.logs.has("would move") {
		t.Fatal("missing rehearsal log")
	}
}

func TestTheDeletionsRideTheBanishMachinery(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Marked.md", "https://d.example/marked",
		withTags("auto_delete"))
	plan := &Plan{}
	rep := c.applyPlan(plan)
	if rep.Banished != 1 {
		t.Fatalf("banished %d", rep.Banished)
	}
	mustBeGone(t, filepath.Join(c.vault, "Design", "Marked.md"))
	dismissed, err := c.db.IsDismissed(
		links.NormalizeWebsiteURL("https://d.example/marked"))
	if err != nil || !dismissed {
		t.Fatalf("dismissed %v %v", dismissed, err)
	}
}

// --- TestTheLandableMove ----------------------------------------------

func TestTheCollisionLoopIsDead(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Palettes/Same.md", "https://settled.example/x")
	c.note("Same.md", "https://orphan.example/y",
		withCategory("Design"), withSubcategory("Palettes"))
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	mv := plan.Moves[0]
	if mv.To != "Design/Palettes" {
		t.Fatalf("to %q", mv.To)
	}
	if mv.As != "Same_v1.md" {
		t.Fatalf("as %q", mv.As)
	}
	if !c.logs.has("the move rides as Same_v1.md") {
		t.Fatal("missing bend log")
	}
	rep := c.applyPlan(&plan)
	if rep.NotesMoved != 1 || rep.FailedMoves != 0 || rep.Stamped != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, filepath.Join(c.vault, "Design", "Palettes", "Same_v1.md"))
	mustBeGone(t, filepath.Join(c.vault, "Same.md"))
	if !c.logs.has("(as Same_v1.md)") {
		t.Fatal("missing as-log")
	}
	mustBeFile(t, filepath.Join(c.vault, "Design", "Palettes", "Same.md"))
	// THE LOOP: the second scan proposes nothing for the note.
	plan2 := c.buildPlan(nil)
	if len(plan2.Moves) != 0 {
		t.Fatalf("second-scan moves %+v", plan2.Moves)
	}
}

// windowsEyes mirrors the reference suite's _windows_eyes: while the
// eyes stand, the stat seam answers as a case-insensitive Windows
// disk would (every exact hit delegates to the real stat untouched).
type windowsEyes struct{}

func (windowsEyes) resolve(p string) string {
	if _, err := os.Stat(p); err == nil {
		return p
	}
	var segs []string
	for _, s := range strings.Split(strings.ReplaceAll(p, "\\", "/"), "/") {
		if s != "" {
			segs = append(segs, s)
		}
	}
	cur := "/"
	for _, s := range segs {
		if s == "." || s == ".." {
			return ""
		}
		entries, err := os.ReadDir(cur)
		if err != nil {
			return ""
		}
		hit := ""
		for _, e := range entries {
			if e.Name() == s {
				hit = e.Name()
				break
			}
		}
		if hit == "" {
			for _, e := range entries {
				if strings.EqualFold(e.Name(), s) {
					hit = e.Name()
					break
				}
			}
		}
		if hit == "" {
			return ""
		}
		cur = filepath.Join(cur, hit)
	}
	return cur
}

func withWindowsEyes(t *testing.T) {
	t.Helper()
	origExists, origIsDir, origIsFile := fsExists, fsIsDir, fsIsFile
	eyes := windowsEyes{}
	fsExists = func(p string) bool {
		r := eyes.resolve(p)
		return r != "" && origExists(r)
	}
	fsIsDir = func(p string) bool {
		r := eyes.resolve(p)
		return r != "" && origIsDir(r)
	}
	fsIsFile = func(p string) bool {
		r := eyes.resolve(p)
		return r != "" && origIsFile(r)
	}
	t.Cleanup(func() {
		fsExists, fsIsDir, fsIsFile = origExists, origIsDir, origIsFile
	})
}

func TestASelfMoveBySpellingDriftIsNeverProposed(t *testing.T) {
	c := newScanCase(t)
	for i := 0; i < 41; i++ {
		c.note(fmt.Sprintf("AI-Domain/Agents/filler-%02d.md", i),
			fmt.Sprintf("https://f%02d.example/%d", i, i))
	}
	c.note("AI-Domain/Agents/Self.md", "https://self.example/z",
		withCategory("AI-Domain"), withSubcategory("agents"))
	withWindowsEyes(t)
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 0 {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestTheDiskSpellingWins(t *testing.T) {
	c := newScanCase(t)
	if err := os.MkdirAll(filepath.Join(c.vault, "AI-Domain", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	memo := map[string]*string{}
	withWindowsEyes(t)
	hit := ResolveDiskFolder(c.vault, "AI-Domain/agents", memo)
	if hit == nil || *hit != "AI-Domain/Agents" {
		t.Fatalf("hit %v", hit)
	}
	if m := memo[foldPath("AI-Domain/agents")]; m == nil ||
		*m != "AI-Domain/Agents" {
		t.Fatalf("memo %v", memo)
	}
}

func TestADeclaredHomeLandsInTheDiskSpelling(t *testing.T) {
	c := newScanCase(t)
	if err := os.MkdirAll(filepath.Join(c.vault, "AI-Domain", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	c.note("Drift.md", "https://d.example/1",
		withCategory("AI-Domain"), withSubcategory("agents"))
	withWindowsEyes(t)
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 1 || plan.Moves[0].To != "AI-Domain/Agents" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	rep := c.applyPlan(&plan)
	if rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, filepath.Join(c.vault, "AI-Domain", "Agents", "Drift.md"))
}

func TestTwoNotesOfOneNameBendPastEachOther(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Palettes/Name.md", "https://settled.example/x")
	c.note("Name.md", "https://one.example/1",
		withCategory("Design"), withSubcategory("Palettes"))
	c.note("orphan/Name_v1.md", "https://two.example/2",
		withCategory("Design"), withSubcategory("Palettes"))
	plan := c.buildPlan(nil)
	var as []string
	for _, m := range plan.Moves {
		as = append(as, m.As)
	}
	if len(as) != 2 || as[0] != "Name_v1.md" || as[1] != "Name_v1_v1.md" {
		t.Fatalf("as %v", as)
	}
	rep := c.applyPlan(&plan)
	if rep.NotesMoved != 2 || rep.FailedMoves != 0 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, filepath.Join(c.vault, "Design", "Palettes", "Name_v1.md"))
	mustBeFile(t, filepath.Join(c.vault, "Design", "Palettes", "Name_v1_v1.md"))
}

func TestTheValidatorFoldsAlreadyHome(t *testing.T) {
	c := newScanCase(t)
	p := c.note("AI-Domain/Agents/Home.md", "https://h.example/1")
	cands := []*NoteRow{{
		File: "Home.md", Rel: "AI-Domain/Agents/Home.md",
		Path: p, Folder: "AI-Domain/Agents",
	}}
	data := LLMProposal{
		NewFolders: []string{},
		Moves: []RawMove{{Note: "Home.md", To: "AI-Domain/agents",
			Reason: "drifted"}},
	}
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI-Domain/Agents", NoteCount: 1}}}
	moves, _ := ValidateMoves(data, cands, inv, c.logs.log,
		[]string{"AI-Domain/agents"}, c.vault, VaultKindWebsites)
	if len(moves) != 0 {
		t.Fatalf("moves %+v", moves)
	}
}

func TestADestinationThatIsTheNoteItselfIsAlreadyHome(t *testing.T) {
	c := newScanCase(t)
	p := c.note("AI-Domain/Agents/Home.md", "https://h.example/1")
	if err := os.MkdirAll(filepath.Join(c.vault, "Design", "Palettes"),
		0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.Symlink(p, filepath.Join(c.vault, "Design", "Palettes",
		"Home.md")); err != nil {
		t.Skipf("symlink unavailable: %v", err)
	}
	cands := []*NoteRow{{
		File: "Home.md", Rel: "AI-Domain/Agents/Home.md",
		Path: p, Folder: "AI-Domain/Agents",
	}}
	data := LLMProposal{
		NewFolders: []string{},
		Moves: []RawMove{{Note: "Home.md", To: "Design/Palettes",
			Reason: "alias"}},
	}
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI-Domain/Agents", NoteCount: 1},
		{Rel: "Design/Palettes", NoteCount: 1}}}
	moves, _ := ValidateMoves(data, cands, inv, c.logs.log, nil,
		c.vault, VaultKindWebsites)
	if len(moves) != 0 {
		t.Fatalf("moves %+v", moves)
	}
}

func TestAVanishedSourceIsSpokenByName(t *testing.T) {
	c := newScanCase(t)
	plan := &Plan{
		Moves: []MoveRow{{
			Note: "Ghost.md",
			Path: filepath.Join(c.vault, "Ghost.md"),
			From: "(vault root)", To: "AI-Domain/Agents", Reason: "x",
		}},
	}
	rep := c.applyPlan(plan)
	if rep.FailedMoves != 1 {
		t.Fatalf("failed %d", rep.FailedMoves)
	}
	if !c.logs.has("Ghost.md is gone") {
		t.Fatal("missing gone log")
	}
	if len(rep.Failed) != 1 || rep.Failed[0].Note != "Ghost.md" ||
		rep.Failed[0].To != "AI-Domain/Agents" ||
		rep.Failed[0].Why != "gone" {
		t.Fatalf("failed ledger %+v", rep.Failed)
	}
}

func TestAMoveThatDidNotLandIsSpoken(t *testing.T) {
	c := newScanCase(t)
	c.note("Mover.md", "https://m.example/1")
	plan := &Plan{
		Moves: []MoveRow{{
			Note: "Mover.md",
			Path: filepath.Join(c.vault, "Mover.md"),
			From: "(vault root)", To: "AI-Domain/Agents", Reason: "x",
		}},
	}
	// a move that silently no-ops: patch the dry-run seam's move —
	// the Go shape for the reference's mock.patch of _dryrun.move.
	origMove := dryrunMoveSeam
	dryrunMoveSeam = func(src, dst string) error { return nil }
	defer func() { dryrunMoveSeam = origMove }()
	rep := c.applyPlan(plan)
	dryrunMoveSeam = origMove
	if rep.FailedMoves != 1 {
		t.Fatalf("failed %d", rep.FailedMoves)
	}
	if !c.logs.has("Could not verify the move of Mover.md") {
		t.Fatal("missing did-not-land log")
	}
	if len(rep.Failed) != 1 || rep.Failed[0].Note != "Mover.md" ||
		rep.Failed[0].Why != "did not land" {
		t.Fatalf("failed ledger %+v", rep.Failed)
	}
	mustExist(t, filepath.Join(c.vault, "Mover.md"))
}

// --- TestTheRegexBrain (the consolidation magnet) ----------------------

func dupeVault(c *scanCase) {
	c.note("Design/Assets/Cleanup Pictures.md", "https://cleanup.pictures",
		withCategory("Design"), withSubcategory("Assets"))
	c.note("Design/Assets/Cleanup Pictures_v1.md",
		"https://cleanup.pictures/",
		withCategory("Design"), withSubcategory("Assets"))
	c.note("Design/Assets/X Tools.md", "https://xtools.example.net",
		withCategory("Design"), withSubcategory("Assets"))
	c.note("Design/Assets/X Tools_v1.md",
		"https://xtools.example.net/gradients",
		withCategory("Design"), withSubcategory("Assets"))
}

func TestTheOwnersPairBecomesOneProposal(t *testing.T) {
	c := newScanCase(t)
	dupeVault(c)
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	props := ScanSiteDuplicates(&inv, c.logs.log, VaultKindWebsites)
	if len(props) != 2 {
		t.Fatalf("props %d", len(props))
	}
	bySite := map[string]Consolidation{}
	for _, p := range props {
		bySite[p.Site] = p
	}
	cp, ok := bySite["cleanup.pictures"]
	if !ok {
		t.Fatalf("sites %v", bySite)
	}
	if cp.Keep.File != "Cleanup Pictures.md" {
		t.Fatalf("keep %q", cp.Keep.File)
	}
	if len(cp.Merge) != 1 || cp.Merge[0].File != "Cleanup Pictures_v1.md" {
		t.Fatalf("merge %+v", cp.Merge)
	}
	xt := bySite["xtools.example.net"]
	if xt.Keep.File != "X Tools.md" || len(xt.Merge) != 1 {
		t.Fatalf("xt %+v", xt)
	}
	if !strings.Contains(c.logs.joined(), "🧲") {
		t.Fatal("missing magnet log")
	}
}

func TestHandNotesNeverRideConsolidations(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Assets/App.md", "https://s.example", withCategory("Design"))
	c.note("Design/Assets/Mine.md", "https://s.example/other",
		withManaged("human"), withCategory("Design"))
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	if props := ScanSiteDuplicates(&inv, nil, VaultKindWebsites); len(props) != 0 {
		t.Fatalf("props %+v", props)
	}
}

func TestCoreNotesNeverRideConsolidations(t *testing.T) {
	c := newScanCase(t)
	c.note("Website Directory.md", "https://dir.example", withNoteKind("directory"))
	c.note("Design/Assets/App.md", "https://dir.example/page",
		withCategory("Design"))
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	if props := ScanSiteDuplicates(&inv, nil, VaultKindWebsites); len(props) != 0 {
		t.Fatalf("props %+v", props)
	}
}

func TestPlaceholdersAloneProduceNothing(t *testing.T) {
	c := newScanCase(t)
	c.note("_review/a.md", "https://w.example/one", withFetchStatus("failed"))
	c.note("_review/b.md", "https://w.example/two", withFetchStatus("failed"))
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	if props := ScanSiteDuplicates(&inv, nil, VaultKindWebsites); len(props) != 0 {
		t.Fatalf("props %+v", props)
	}
}

func TestAPlaceholderMergesIntoTheRealNote(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Assets/Real.md", "https://w.example", withCategory("Design"))
	c.note("_review/w.md", "https://w.example/walled", withFetchStatus("failed"))
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	props := ScanSiteDuplicates(&inv, nil, VaultKindWebsites)
	if len(props) != 1 {
		t.Fatalf("props %d", len(props))
	}
	if props[0].Keep.File != "Real.md" {
		t.Fatalf("keep %q", props[0].Keep.File)
	}
	if len(props[0].Merge) != 1 || props[0].Merge[0].File != "w.md" {
		t.Fatalf("merge %+v", props[0].Merge)
	}
}

func TestTheLLMIsNeverAskedForTheRegexHalf(t *testing.T) {
	c := newScanCase(t)
	dupeVault(c)
	calls := 0
	llm := func([]llmclient.Message, string) string {
		calls++
		return "{}"
	}
	plan := c.buildPlan(llm)
	if len(plan.Consolidations) != 2 {
		t.Fatalf("consolidations %d", len(plan.Consolidations))
	}
	if calls != 0 {
		t.Fatalf("the LLM was asked %d times", calls)
	}
	if len(plan.Moves) != 0 || len(plan.NewFolders) != 0 {
		t.Fatalf("plan %+v", plan)
	}
}

// --- TestTheConsolidationApply ----------------------------------------

func consolidationPlan(c *scanCase) (*Plan, string, string) {
	keep := c.note("Design/Assets/X Tools.md", "https://xtools.example.net",
		withCategory("Design"), withSubcategory("Assets"))
	dupe := c.note("Design/Assets/X Tools_v1.md",
		"https://xtools.example.net/gradients",
		withCategory("Design"), withSubcategory("Assets"))
	return &Plan{
		Consolidations: []Consolidation{{
			Site: "xtools.example.net",
			Keep: KeepRow{
				File: "X Tools.md", Path: keep,
				Source: "https://xtools.example.net",
				Title:  "X Tools", Category: "Design",
				Subcategory: "Assets & Resources",
			},
			Merge: []MergeRow{{
				File: "X Tools_v1.md", Path: dupe,
				Source: "https://xtools.example.net/gradients",
				Title:  "X Tools",
			}},
			URLs: []string{"https://xtools.example.net/gradients"},
		}},
	}, keep, dupe
}

func TestApplyMergesAndRetires(t *testing.T) {
	c := newScanCase(t)
	plan, keep, dupe := consolidationPlan(c)
	if err := c.db.MarkProcessed(
		"https://xtools.example.net/gradients", dupe,
		"Design", "Assets", "failed"); err != nil {
		t.Fatal(err)
	}
	c.db.EnqueueRetry("https://xtools.example.net/gradients", "wall")
	rep := c.applyPlan(plan)
	if rep.Consolidated != 1 {
		t.Fatalf("consolidated %d", rep.Consolidated)
	}
	mustBeGone(t, dupe)
	mustBeFile(t, filepath.Join(c.vault, ".trash", "consolidated",
		"X Tools_v1.md"))
	mustBeFile(t, keep)
	content := readFileT(t, keep)
	if !strings.Contains(content,
		"site_links: [https://xtools.example.net/gradients]") {
		t.Fatalf("no link: %q", content)
	}
	if !strings.Contains(content, websitepipeline.SiteLinksHeading) {
		t.Fatal("no heading")
	}
	row, err := c.db.ProcessedRow("https://xtools.example.net/gradients")
	if err != nil || row == nil {
		t.Fatalf("row %v", row)
	}
	if row.NotePath == nil || *row.NotePath != keep {
		t.Fatalf("row path %v", row.NotePath)
	}
	if retry, err := c.db.RetryRow(
		"https://xtools.example.net/gradients"); err != nil || retry != nil {
		t.Fatalf("retry %v %v", retry, err)
	}
	dismissed, err := c.db.IsDismissed(
		"https://xtools.example.net/gradients")
	if err != nil || dismissed {
		t.Fatal("the site was dismissed — it is wanted")
	}
	if !strings.Contains(c.logs.joined(), "🧲") {
		t.Fatal("missing magnet log")
	}
}

func TestTheOwnersSlashPairAddsNoRedundantLink(t *testing.T) {
	c := newScanCase(t)
	keep := c.note("Design/Assets/Cleanup Pictures.md",
		"https://cleanup.pictures", withCategory("Design"))
	dupe := c.note("Design/Assets/Cleanup Pictures_v1.md",
		"https://cleanup.pictures/", withCategory("Design"))
	plan := &Plan{
		Consolidations: []Consolidation{{
			Site: "cleanup.pictures",
			Keep: KeepRow{
				File: "Cleanup Pictures.md", Path: keep,
				Source: "https://cleanup.pictures",
				Title:  "Cleanup", Category: "Design",
			},
			Merge: []MergeRow{{
				File: "Cleanup Pictures_v1.md", Path: dupe,
				Source: "https://cleanup.pictures/",
				Title:  "Cleanup",
			}},
			URLs: []string{"https://cleanup.pictures/"},
		}},
	}
	rep := c.applyPlan(plan)
	if rep.Consolidated != 1 {
		t.Fatalf("consolidated %d", rep.Consolidated)
	}
	mustBeGone(t, dupe)
	content := readFileT(t, keep)
	if strings.Contains(content, "site_links") {
		t.Fatal("self-referential site_links")
	}
	if strings.Contains(content, websitepipeline.SiteLinksHeading) {
		t.Fatal("heading written")
	}
	row, err := c.db.ProcessedRow("https://cleanup.pictures/")
	if err != nil || row == nil {
		t.Fatalf("row %v %v", row, err)
	}
	if row.NotePath == nil || *row.NotePath != keep {
		t.Fatalf("row path %v", row.NotePath)
	}
}

func TestDryRunIsARehearsalForConsolidations(t *testing.T) {
	c := newScanCase(t)
	plan, keep, dupe := consolidationPlan(c)
	dryrun.Enable()
	rep := c.applyPlan(plan)
	dryrun.Disable()
	if rep.Consolidated < 1 {
		t.Fatalf("report %+v", rep)
	}
	mustExist(t, dupe)
	if strings.Contains(readFileT(t, keep), "site_links") {
		t.Fatal("the rehearsal wrote")
	}
}

func TestAMissingKeeperIsAQuietSkip(t *testing.T) {
	c := newScanCase(t)
	plan, keep, dupe := consolidationPlan(c)
	if err := os.Remove(keep); err != nil {
		t.Fatal(err)
	}
	rep := c.applyPlan(plan)
	if rep.Consolidated != 0 {
		t.Fatalf("consolidated %d", rep.Consolidated)
	}
	mustExist(t, dupe)
	if !c.logs.has("gone") {
		t.Fatal("missing gone log")
	}
}

// --- TestTheDeclaredHomePass ------------------------------------------

func TestADeclaredHomeFilesWithoutAskingTheLLM(t *testing.T) {
	c := newScanCase(t)
	c.note("stray-root.md", "https://a.example/declared",
		withCategory("AI-Domain"), withSubcategory("Agents"))
	if err := os.MkdirAll(filepath.Join(c.vault, "AI-Domain", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	boom := func([]llmclient.Message, string) string {
		panic("the LLM was asked about a note that declares its own home")
	}
	plan := c.buildPlan(boom)
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	mv := plan.Moves[0]
	if mv.Note != "stray-root.md" || mv.To != "AI-Domain/Agents" {
		t.Fatalf("move %+v", mv)
	}
	if !strings.Contains(mv.Reason, "frontmatter") {
		t.Fatalf("reason %q", mv.Reason)
	}
	if !c.logs.has("regex brain") {
		t.Fatal("missing regex-brain log")
	}
}

func TestADeclaredHomeMissingOnDiskRidesTheLLM(t *testing.T) {
	c := newScanCase(t)
	c.note("rooted.md", "https://a.example/rooted",
		withCategory("AI-Domain"), withSubcategory("Nope"))
	c.note("AI-Domain/Agents/Existing.md", "https://b.example/existing")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "rooted.md",
			"to": "AI-Domain/Agents", "reason": "closest existing home"}},
		"new_folders": []any{}, "summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 1 || plan.Moves[0].To != "AI-Domain/Agents" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
}

func TestACategoryWithoutASubcategoryRidesTheLLM(t *testing.T) {
	c := newScanCase(t)
	c.note("half-declared.md", "https://a.example/half",
		withCategory("AI-Domain"))
	c.note("AI-Domain/Agents/Existing.md", "https://b.example/existing")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "half-declared.md",
			"to": "AI-Domain/Agents", "reason": "thinking"}},
		"new_folders": []any{}, "summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 1 || len(llm.calls) != 1 {
		t.Fatalf("moves %+v calls %d", plan.Moves, len(llm.calls))
	}
}

func TestTheDeclaredMoveSurvivesValidationAndApplies(t *testing.T) {
	c := newScanCase(t)
	c.note("applied.md", "https://a.example/applied",
		withCategory("AI-Domain"), withSubcategory("Agents"))
	if err := os.MkdirAll(filepath.Join(c.vault, "AI-Domain", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	rep := c.applyPlan(&plan)
	if rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, filepath.Join(c.vault, "AI-Domain", "Agents", "applied.md"))
	mustBeGone(t, filepath.Join(c.vault, "applied.md"))
}

func TestADeclaredHomeWithCaseFoldMeetsTheDiskFolder(t *testing.T) {
	c := newScanCase(t)
	c.note("cased.md", "https://a.example/cased",
		withCategory("AI-Domain"), withSubcategory("agents"))
	if err := os.MkdirAll(filepath.Join(c.vault, "AI-Domain", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 1 || plan.Moves[0].To != "AI-Domain/Agents" {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestAlreadyHomeDeclaredNotesNeverMove(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/home.md", "https://a.example/home",
		withCategory("AI-Domain"), withSubcategory("Agents"))
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 0 {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

// --- TestTheChunkedFiling ---------------------------------------------

func manyCandidates(c *scanCase, count int) {
	c.t.Helper()
	for i := 0; i < count; i++ {
		c.note(fmt.Sprintf("orphan-%03d.md", i),
			fmt.Sprintf("https://o%03d.example/%d", i, i))
	}
}

func TestTheLLMIsAskedInChunksNotOneGiantPrompt(t *testing.T) {
	c := newScanCase(t)
	manyCandidates(c, 130)
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "orphan-000.md",
			"to": "AI-Domain/Agents", "reason": "x"}},
		"new_folders": []any{}, "summary": "",
	})
	c.note("AI-Domain/Agents/Existing.md", "https://b.example/existing")
	c.buildPlan(llm.call, withConfig("scan_llm_batch", 60))
	if len(llm.calls) != 3 { // 60 + 60 + 10
		t.Fatalf("calls %d != 3", len(llm.calls))
	}
	for _, msgs := range llm.calls {
		body := msgs[1].Content
		if !strings.Contains(body, "THE LATEST STATE") ||
			!strings.Contains(body, "three laws") {
			t.Fatal("a chunk missed the laws")
		}
	}
	if !c.logs.has("Filing pass 1 of 3") || !c.logs.has("Filing pass 3 of 3") {
		t.Fatal("missing per-chunk story")
	}
}

func TestOneFailedChunkNeverKillsTheRest(t *testing.T) {
	c := newScanCase(t)
	manyCandidates(c, 130)
	c.note("AI-Domain/Agents/Existing.md", "https://b.example/existing")
	flaky := &fakeLLM{
		plan:        `{"moves": [{"note": "orphan-000.md", "to": "AI-Domain/Agents", "reason": "x"}], "new_folders": [], "summary": ""}`,
		panicsAfter: 2,
	}
	plan := c.buildPlan(flaky.call, withConfig("scan_llm_batch", 60))
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	if !c.logs.has("landed 1 of 3") {
		t.Fatal("missing landed log")
	}
}

func TestTheDefaultBatchIsTheHouseNumber(t *testing.T) {
	if DefaultLLMBatch != 60 {
		t.Fatalf("batch %d", DefaultLLMBatch)
	}
	c := newScanCase(t)
	manyCandidates(c, 10)
	llm := newFakeLLM(map[string]any{"moves": []any{},
		"new_folders": []any{}, "summary": ""})
	c.buildPlan(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
	if c.logs.has("Filing pass") {
		t.Fatal("chunk noise for a small pile")
	}
}

func TestTheProgressCallableHearsTheWalkAndThePasses(t *testing.T) {
	c := newScanCase(t)
	manyCandidates(c, 130)
	c.note("AI-Domain/Agents/Existing.md", "https://b.example/existing")
	var seen [][3]any // label, value, max — max mixed int types
	_ = seen
	type progRow struct {
		label string
		value int
		max   int
	}
	var rows []progRow
	progress := func(label string, value, max int) {
		rows = append(rows, progRow{label, value, max})
	}
	plan := BuildScanPlan(c.vault,
		func([]llmclient.Message, string) string { return "{}" },
		c.logs.log, map[string]any{"scan_llm_batch": 60}, progress,
		nil, VaultKindWebsites, nil)
	_ = plan
	var filing []progRow
	reading := false
	for _, r := range rows {
		if strings.Contains(r.label, "reading") {
			reading = true
		}
		if strings.Contains(r.label, "LLM filing pass") {
			filing = append(filing, r)
		}
	}
	if !reading {
		t.Fatal("the walk's progress never spoke")
	}
	if len(filing) == 0 {
		t.Fatal("the filing pass never spoke")
	}
	last := filing[len(filing)-1]
	if last.value != 3 || last.max != 3 {
		t.Fatalf("last filing progress %+v", last)
	}
}

func TestTheFrontmatterReadIsTheCheapHead(t *testing.T) {
	if FrontmatterReadRunes != 16000 {
		t.Fatalf("head %d", FrontmatterReadRunes)
	}
	c := newScanCase(t)
	path := filepath.Join(c.vault, "big.md")
	big := "---\nsource: \"https://a.example/big\"\n" +
		"tags: [x]\ncategory: \"\"\nsubcategory: \"\"\n" +
		"managed_by: \"gitcurator\"\nfetch_status: \"full\"\n" +
		"---\n\n# Big\n\n" + strings.Repeat("body text\n", 6000)
	if err := os.WriteFile(path, []byte(big), 0o644); err != nil {
		t.Fatal(err)
	}
	fm := parseNoteFrontmatter(path)
	if fm.Source != "https://a.example/big" {
		t.Fatalf("source %q", fm.Source)
	}
	if fm.ManagedBy != "gitcurator" {
		t.Fatalf("managed_by %q", fm.ManagedBy)
	}
}

// --- TestTheSiblingDomainPass -----------------------------------------

func TestADomainsAgreedHomeFilesWithoutAskingTheLLM(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/Settled One.md",
		"https://figma.example/library/one")
	c.note("AI-Domain/Agents/Settled Two.md",
		"https://figma.example/library/two")
	c.note("stray.md", "https://figma.example")
	boom := func([]llmclient.Message, string) string {
		panic("the LLM was asked about a note whose domain has an agreed home")
	}
	plan := c.buildPlan(boom)
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	mv := plan.Moves[0]
	if mv.Note != "stray.md" || mv.To != "AI-Domain/Agents" {
		t.Fatalf("move %+v", mv)
	}
	if !strings.Contains(mv.Reason, "settled sibling") {
		t.Fatalf("reason %q", mv.Reason)
	}
	if !c.logs.has("filed beside their siblings") {
		t.Fatal("missing siblings log")
	}
}

func TestSplitSiblingsNeverTeach(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/Home A.md", "https://figma.example/library/a")
	c.note("Design/Assets/Home B.md", "https://figma.example/library/b")
	c.note("stray.md", "https://figma.example")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "stray.md",
			"to": "AI-Domain/Agents", "reason": "thinking"}},
		"new_folders": []any{}, "summary": "",
	})
	plan := c.buildPlan(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
	if len(plan.Moves) != 1 || plan.Moves[0].Reason != "thinking" {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestNoSettledSiblingsRidesTheLLM(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/Other.md", "https://b.example/x")
	c.note("stray.md", "https://figma.example/stray")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "stray.md",
			"to": "AI-Domain/Agents", "reason": "thinking"}},
		"new_folders": []any{}, "summary": "",
	})
	c.buildPlan(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
}

func TestAFailedPlaceholderNeverTeaches(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/Wall.md",
		"https://figma.example/library/wall", withFetchStatus("failed"))
	c.note("stray.md", "https://figma.example")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "stray.md",
			"to": "AI-Domain/Agents", "reason": "thinking"}},
		"new_folders": []any{}, "summary": "",
	})
	c.buildPlan(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
}

func TestACandidateNeverTeaches(t *testing.T) {
	c := newScanCase(t)
	c.note("stray-a.md", "https://figma.example")
	c.note("stray-b.md", "https://figma.example/twin")
	llm := newFakeLLM(map[string]any{
		"moves": []any{}, "new_folders": []any{}, "summary": ""})
	c.buildPlan(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
}

func TestTheSiblingMoveSurvivesValidationAndApplies(t *testing.T) {
	c := newScanCase(t)
	c.note("AI-Domain/Agents/Settled.md",
		"https://figma.example/library/settled")
	c.note("stray.md", "https://figma.example")
	plan := c.buildPlan(nil)
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	rep := c.applyPlan(&plan)
	if rep.NotesMoved != 1 || rep.Consolidated != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, filepath.Join(c.vault, "AI-Domain", "Agents", "stray.md"))
	mustBeGone(t, filepath.Join(c.vault, "stray.md"))
	mustBeGone(t, filepath.Join(c.vault, "AI-Domain", "Agents", "Settled.md"))
}

// --- TestTheTaxonomyPass ----------------------------------------------

const tinyTaxonomy = `# The test taxonomy

## Categories

### Design
- **Assets & Resources** — icons, mockups, fonts, illustrations
- **UI/UX & Product Design** — design systems, wireframing

### Developer Tools
*(no subcategories)*
`

type taxCase struct {
	*scanCase
	cfg map[string]any
}

func newTaxCase(t *testing.T) *taxCase {
	c := newScanCase(t)
	tax := filepath.Join(c.tmp, "taxonomy.md")
	if err := os.WriteFile(tax, []byte(tinyTaxonomy), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c.vault, "Design",
		"Assets & Resources"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(filepath.Join(c.vault, "Design",
		"UI_UX_Product_Design"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &taxCase{scanCase: c,
		cfg: map[string]any{"taxonomy_path": tax}}
}

func (tc *taxCase) build(llm LLMCallFn) Plan {
	return BuildScanPlan(tc.vault, llm, tc.logs.log, tc.cfg, nil, nil,
		VaultKindWebsites, nil)
}

func TestTagsThatNameOneHomeFileByTheTaxonomy(t *testing.T) {
	tc := newTaxCase(t)
	tc.note("stray.md", "https://a.example/stray",
		withTags("icons", "mockup"))
	boom := func([]llmclient.Message, string) string {
		panic("the LLM was asked about a note whose tags name one taxonomy home")
	}
	plan := tc.build(boom)
	if len(plan.Moves) != 1 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	mv := plan.Moves[0]
	if mv.Note != "stray.md" {
		t.Fatalf("note %q", mv.Note)
	}
	if mv.To != "Design/Assets & Resources" {
		t.Fatalf("to %q (the DISK spelling wins)", mv.To)
	}
	if !strings.Contains(mv.Reason, "taxonomy") {
		t.Fatalf("reason %q", mv.Reason)
	}
	if !tc.logs.has("the owner's own category map") {
		t.Fatal("missing taxonomy log")
	}
}

func TestConflictingTagsRideTheLLM(t *testing.T) {
	tc := newTaxCase(t)
	tc.note("Design/Assets & Resources/Existing.md", "https://b.example/x")
	tc.note("stray.md", "https://a.example/stray",
		withTags("icons", "wireframing"))
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "stray.md",
			"to": "Design/Assets & Resources", "reason": "thinking"}},
		"new_folders": []any{}, "summary": "",
	})
	plan := tc.build(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
	if plan.Moves[0].Reason != "thinking" {
		t.Fatalf("reason %q", plan.Moves[0].Reason)
	}
}

func TestAHomeMissingOnDiskRidesTheLLM(t *testing.T) {
	tc := newTaxCase(t)
	if err := os.Remove(filepath.Join(tc.vault, "Design",
		"UI_UX_Product_Design")); err != nil {
		t.Fatal(err)
	}
	tc.note("Design/Assets & Resources/Existing.md", "https://b.example/x")
	tc.note("stray.md", "https://a.example/stray",
		withTags("wireframing"))
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"Design/UI_UX_Product_Design"},
		"moves": []any{map[string]any{"note": "stray.md",
			"to": "Design/UI_UX_Product_Design", "reason": "thinking"}},
		"summary": "",
	})
	plan := tc.build(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d", len(llm.calls))
	}
	if len(plan.NewFolders) != 1 ||
		plan.NewFolders[0] != "Design/UI_UX_Product_Design" {
		t.Fatalf("new folders %v", plan.NewFolders)
	}
}

func TestPluralToleranceBothWays(t *testing.T) {
	tc := newTaxCase(t)
	tc.note("stray.md", "https://a.example/stray", withTags("font"))
	plan := tc.build(nil)
	if len(plan.Moves) != 1 || plan.Moves[0].To != "Design/Assets & Resources" {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestUntaggedAndUnknownTagsRideTheLLM(t *testing.T) {
	tc := newTaxCase(t)
	tc.note("AI-Domain/Agents/Existing.md", "https://b.example/x")
	tc.note("no-tags.md", "https://a.example/one")
	tc.note("unknown.md", "https://u.example/two",
		withTags("zebra-unicorn"))
	llm := newFakeLLM(map[string]any{"moves": []any{},
		"new_folders": []any{}, "summary": ""})
	tc.build(llm.call)
	if len(llm.calls) != 1 {
		t.Fatalf("calls %d (both must ride the LLM)", len(llm.calls))
	}
}

func TestTheHonestSplitLineNamesEveryPass(t *testing.T) {
	tc := newTaxCase(t)
	tc.note("Design/Assets & Resources/Settled.md",
		"https://figma.example/library/settled")
	tc.note("declared.md", "https://d.example/declared",
		withCategory("Design"), withSubcategory("Assets & Resources"))
	tc.note("sibling.md", "https://figma.example")
	tc.note("tagged.md", "https://t.example/tagged", withTags("icons"))
	plan := tc.build(nil)
	if len(plan.Moves) != 3 {
		t.Fatalf("moves %d", len(plan.Moves))
	}
	if plan.RegexFiled != 3 || plan.LLMCandidates != 0 {
		t.Fatalf("filed %d candidates %d", plan.RegexFiled,
			plan.LLMCandidates)
	}
	if !tc.logs.has("The regex brain filed 3 of 3") {
		t.Fatal("missing split line")
	}
	if !tc.logs.has("declared home 1") ||
		!tc.logs.has("siblings 1") || !tc.logs.has("taxonomy 1") {
		t.Fatal("missing per-pass counts")
	}
}

// --- TestTheOneOpenScan -----------------------------------------------

func TestTheWalksShareOneOpenPerNote(t *testing.T) {
	c := newScanCase(t)
	for i := 0; i < 4; i++ {
		c.note(fmt.Sprintf("note-%d.md", i),
			fmt.Sprintf("https://a.example/%d", i))
	}
	c.note("marked.md", "https://a.example/marked",
		withBody("not needed #auto-delete"))
	cache := notereads.New(filepath.Join(c.tmp, "reads.db"))
	defer cache.Close(false)
	plan := BuildScanPlan(c.vault, nil, c.logs.log, nil, nil, cache,
		VaultKindWebsites, nil)
	if cache.Stats().Opens != 5 {
		t.Fatalf("opens %d != 5", cache.Stats().Opens)
	}
	if len(plan.Deletions) != 1 {
		t.Fatalf("deletions %d", len(plan.Deletions))
	}
	if plan.Deletions[0].Marker != "#auto-delete" {
		t.Fatalf("marker %q", plan.Deletions[0].Marker)
	}
}

func TestTheNextScanOfTheUnchangedVaultOpensNothing(t *testing.T) {
	c := newScanCase(t)
	for i := 0; i < 4; i++ {
		c.note(fmt.Sprintf("note-%d.md", i),
			fmt.Sprintf("https://a.example/%d", i))
	}
	c.note("marked.md", "https://a.example/marked",
		withBody("not needed #auto-delete"))
	db := filepath.Join(c.tmp, "reads.db")
	cache1 := notereads.New(db)
	plan1 := BuildScanPlan(c.vault, nil, c.logs.log, nil, nil, cache1,
		VaultKindWebsites, nil)
	if cache1.Stats().Opens != 5 {
		t.Fatalf("opens %d != 5", cache1.Stats().Opens)
	}
	cache1.Close(false)
	cache2 := notereads.New(db)
	defer cache2.Close(false)
	plan2 := BuildScanPlan(c.vault, nil, c.logs.log, nil, nil, cache2,
		VaultKindWebsites, nil)
	stats := cache2.Stats()
	if stats.Opens != 0 {
		t.Fatalf("opens %d != 0 (the near-instant law)", stats.Opens)
	}
	if stats.DbHits < 5 {
		t.Fatalf("db_hits %d < 5", stats.DbHits)
	}
	if len(plan2.Deletions) != len(plan1.Deletions) {
		t.Fatalf("deletions diverged")
	}
	if len(plan2.Moves) != len(plan1.Moves) {
		t.Fatalf("moves diverged")
	}
}

// --- TestTheCensusStamp (test_censusstamp selections) -------------------

func TestTheKeyDoorReadsTruthyValues(t *testing.T) {
	for _, val := range []string{"true", "True", "TRUE", "yes", "1",
		`"true"`} {
		if !noteCurated(&notereads.Inventory{Curated: truthy(val)}) {
			t.Fatalf("curated %q should stamp", val)
		}
	}
}

func TestFalsyAndJunkValuesAreNotStamps(t *testing.T) {
	for _, val := range []string{"", "false", "no", "0", "maybe"} {
		if noteCurated(&notereads.Inventory{Curated: truthy(val)}) {
			t.Fatalf("curated %q should not stamp", val)
		}
	}
}

func TestTheTagDoorFoldsPastCaseAndSeparators(t *testing.T) {
	for _, tag := range []string{"curated", "Curated", "CURATED"} {
		if !noteCurated(&notereads.Inventory{Tags: []string{tag}}) {
			t.Fatalf("tag %q should stamp", tag)
		}
	}
}

func TestJunkFrontmatterNeverStamps(t *testing.T) {
	if noteCurated(nil) || noteCurated(&notereads.Inventory{}) {
		t.Fatal("junk stamped")
	}
	if noteCurated(&notereads.Inventory{Tags: []string{"curated-ish"}}) {
		t.Fatal("curated-ish stamped")
	}
	if noteCurated(&notereads.Inventory{
		Tags: []string{"lock"}}) {
		t.Fatal("lock stamped as curated")
	}
}

func TestTheLockAndTheStampAreIndependentWords(t *testing.T) {
	if noteLocked(&notereads.Inventory{Curated: true}) {
		t.Fatal("a stamp locked")
	}
	if noteCurated(&notereads.Inventory{Locked: true}) {
		t.Fatal("a lock stamped")
	}
}

func TestTheDirectParserReadsTheCuratedKey(t *testing.T) {
	c := newScanCase(t)
	p := filepath.Join(c.vault, "n.md")
	if err := os.WriteFile(p, []byte(
		"---\ncurated: true\nsource: \"https://x/1\"\n---\n\n# n\n"),
		0o644); err != nil {
		t.Fatal(err)
	}
	if fm := parseNoteFrontmatter(p); !fm.Curated {
		t.Fatalf("fm %+v", fm)
	}
}

func TestAStampedRootOrphanIsNeverACandidate(t *testing.T) {
	c := newScanCase(t)
	c.note("N.md", "https://n.example/1", withExtra("curated: true"))
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	if inv.CuratedNotes != 1 {
		t.Fatalf("curated %d", inv.CuratedNotes)
	}
	if cands := MoveCandidates(&inv, 40); len(cands) != 0 {
		t.Fatalf("candidates %v", cands)
	}
}

func TestAStampedNoteInABroadFolderIsNeverACandidate(t *testing.T) {
	c := newScanCase(t)
	for i := 0; i < 50; i++ {
		extra := ""
		if i == 0 {
			extra = "curated: true\n"
		}
		c.note(fmt.Sprintf("Big/Sub/n%d.md", i),
			fmt.Sprintf("https://b.example/%d", i), withExtra(extra))
	}
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	cands := MoveCandidates(&inv, 40)
	if len(cands) != 49 {
		t.Fatalf("candidates %d != 49", len(cands))
	}
	for _, cand := range cands {
		if cand.File == "n0.md" {
			t.Fatal("the stamped note rode")
		}
	}
}

func TestTheInventoryStillCountsAndCensusesTheStamped(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Palettes/a.md", "https://a.example/1",
		withExtra("curated: true"))
	c.note("Design/Palettes/b.md", "https://b.example/1")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	if inv.TotalNotes != 2 || inv.CuratedNotes != 1 {
		t.Fatalf("counts %+v", inv)
	}
	folders := map[string]int{}
	for _, f := range inv.Folders {
		folders[f.Rel] = f.NoteCount
	}
	if folders["Design/Palettes"] != 2 {
		t.Fatalf("census %v", folders)
	}
	rows := map[string]NoteRow{}
	for _, n := range inv.Notes {
		rows[n.File] = n
	}
	if !rows["a.md"].Curated || rows["b.md"].Curated {
		t.Fatal("curated flags wrong")
	}
}

func TestTheWalkTeachesTheGrammarAndTheEscapeHatch(t *testing.T) {
	c := newScanCase(t)
	c.note("n.md", "https://n.example/1", withExtra("curated: true"))
	ScanVaultInventory(c.vault, c.logs.log, nil, nil, VaultKindWebsites)
	if !c.logs.has("1 curated note") {
		t.Fatal("missing curated count")
	}
	if !c.logs.has("remove the line by hand") {
		t.Fatal("missing escape hatch")
	}
}

func TestTheStampWriterWritesExactlyOneLine(t *testing.T) {
	c := newScanCase(t)
	p := c.note("n.md", "https://n.example/1")
	before := readFileT(t, p)
	if !StampCurated(p, c.logs.log) {
		t.Fatal("stamp failed")
	}
	after := readFileT(t, p)
	if strings.Replace(after, "\ncurated: true\n", "\n", 1) != before {
		t.Fatal("more than the one line changed")
	}
	if strings.Count(after, "curated: true") != 1 {
		t.Fatal("not exactly one stamp")
	}
}

func TestTheStampWriterIsIdempotent(t *testing.T) {
	c := newScanCase(t)
	p := c.note("n.md", "https://n.example/1")
	if !StampCurated(p, c.logs.log) {
		t.Fatal("first stamp failed")
	}
	if StampCurated(p, c.logs.log) {
		t.Fatal("second stamp changed something")
	}
	if strings.Count(readFileT(t, p), "curated: true") != 1 {
		t.Fatal("not exactly one stamp")
	}
}

func TestANoteWithoutFrontmatterIsNeverGivenOne(t *testing.T) {
	c := newScanCase(t)
	p := filepath.Join(c.vault, "plain.md")
	if err := os.WriteFile(p, []byte("# the owner's own plain note\n"),
		0o644); err != nil {
		t.Fatal(err)
	}
	if StampCurated(p, c.logs.log) {
		t.Fatal("a block was invented")
	}
	if readFileT(t, p) != "# the owner's own plain note\n" {
		t.Fatal("the note changed")
	}
}

func TestAMissingPathIsAQuietNo(t *testing.T) {
	c := newScanCase(t)
	if StampCurated(filepath.Join(c.vault, "nope.md"), c.logs.log) {
		t.Fatal("a ghost was stamped")
	}
}

func TestARehearsalRecordsAndNeverTouchesTheDisk(t *testing.T) {
	c := newScanCase(t)
	p := c.note("n.md", "https://n.example/1")
	before := readFileT(t, p)
	dryrun.Enable()
	ok := StampCurated(p, c.logs.log)
	dryrun.Disable()
	if !ok {
		t.Fatal("the rehearsal did not record")
	}
	if readFileT(t, p) != before {
		t.Fatal("the rehearsal touched the disk")
	}
	if dryrun.EntryCount() < 1 {
		t.Fatal("nothing recorded")
	}
}

func movedPlan(c *scanCase) *Plan {
	c.note("Design/Palettes/resident.md", "https://r.example/1")
	c.note("Mover.md", "https://m.example/1")
	return &Plan{
		Moves: []MoveRow{{
			Note: "Mover.md",
			Path: filepath.Join(c.vault, "Mover.md"),
			From: "(vault root)", To: "Design/Palettes", Reason: "fits",
		}},
	}
}

func TestAMovedNoteIsStampedAtItsNewHome(t *testing.T) {
	c := newScanCase(t)
	plan := movedPlan(c)
	rep := c.applyPlan(plan)
	if rep.NotesMoved != 1 || rep.Stamped != 1 {
		t.Fatalf("report %+v", rep)
	}
	dst := filepath.Join(c.vault, "Design", "Palettes", "Mover.md")
	if !strings.Contains(readFileT(t, dst), "curated: true") {
		t.Fatal("no stamp at the new home")
	}
	if !c.logs.has("stamped curated at its new home") {
		t.Fatal("missing log")
	}
}

func TestADeclinedRowIsStampedInPlace(t *testing.T) {
	c := newScanCase(t)
	plan := movedPlan(c)
	sel := Selection{Moves: map[string]bool{}} // ALL unchecked
	rep := c.applyPlan(plan, withSelection(sel))
	if rep.NotesMoved != 0 || rep.Stamped != 1 {
		t.Fatalf("report %+v", rep)
	}
	src := filepath.Join(c.vault, "Mover.md")
	if !strings.Contains(readFileT(t, src), "curated: true") {
		t.Fatal("no stamp in place")
	}
	if !c.logs.has("stamped curated — the declined move is remembered") {
		t.Fatal("missing decline log")
	}
}

func TestTheTelegramDoorNeverDeclinesAnything(t *testing.T) {
	c := newScanCase(t)
	plan := movedPlan(c)
	rep := c.applyPlan(plan) // selection == zero value → whole plan
	if rep.NotesMoved != 1 || rep.Stamped != 1 {
		t.Fatalf("report %+v", rep)
	}
}

func TestAReviewedNoteIsStampedInPlace(t *testing.T) {
	c := newScanCase(t)
	p := c.note("Stay.md", "https://s.example/1")
	plan := &Plan{Reviewed: []string{p}}
	rep := c.applyPlan(plan)
	if rep.Stamped != 1 {
		t.Fatalf("stamped %d", rep.Stamped)
	}
	if !strings.Contains(readFileT(t, p), "curated: true") {
		t.Fatal("no stamp")
	}
	if !c.logs.has("reviewed it and proposed nothing") {
		t.Fatal("missing review log")
	}
}

func TestALockedNoteIsNeverStamped(t *testing.T) {
	c := newScanCase(t)
	p := c.note("Locked.md", "https://l.example/1",
		withExtra("locked: true"))
	plan := &Plan{Reviewed: []string{p}}
	rep := c.applyPlan(plan)
	if rep.Stamped != 0 {
		t.Fatalf("stamped %d", rep.Stamped)
	}
	if strings.Contains(readFileT(t, p), "curated: true") {
		t.Fatal("a locked note was stamped")
	}
}

func TestAFailedMoveIsNeverStamped(t *testing.T) {
	c := newScanCase(t)
	c.note("Ghost.md", "https://g.example/1")
	plan := &Plan{
		Moves: []MoveRow{{
			Note: "Ghost.md",
			Path: filepath.Join(c.vault, "gone.md"),
			From: "(vault root)", To: "Design/Palettes", Reason: "x",
		}},
	}
	rep := c.applyPlan(plan)
	if rep.FailedMoves != 1 || rep.Stamped != 0 {
		t.Fatalf("report %+v", rep)
	}
}

func TestARehearsalRehearsesTheStamps(t *testing.T) {
	c := newScanCase(t)
	plan := movedPlan(c)
	dryrun.Enable()
	rep := c.applyPlan(plan)
	dryrun.Disable()
	if rep.NotesMoved != 1 || rep.Stamped != 1 {
		t.Fatalf("report %+v", rep)
	}
	if !c.logs.has("would be stamped curated") {
		t.Fatal("missing rehearsal log")
	}
	mustBeFile(t, filepath.Join(c.vault, "Mover.md"))
	if strings.Contains(readFileT(t, filepath.Join(c.vault, "Mover.md")),
		"curated: true") {
		t.Fatal("the rehearsal wrote the stamp")
	}
}

func TestTheClosingLineSpeaksTheLaw(t *testing.T) {
	c := newScanCase(t)
	p := c.note("Stay.md", "https://s.example/1")
	plan := &Plan{Reviewed: []string{p}}
	c.applyPlan(plan)
	if !c.logs.has("the scan remembers its own work") {
		t.Fatal("missing closing line")
	}
}

// --- TestTheReviewedList (test_censusstamp) ---------------------------

func TestExaminedWithNoProposalIsReviewed(t *testing.T) {
	c := newScanCase(t)
	c.note("n0.md", "https://n0.example/1")
	c.note("n1.md", "https://n1.example/1")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "n0.md",
			"to": "Nowhere/Sub", "reason": "x"}}})
	plan := c.buildPlan(llm.call)
	want := []string{filepath.Join(c.vault, "n1.md")}
	if len(plan.Reviewed) != 1 || plan.Reviewed[0] != want[0] {
		t.Fatalf("reviewed %v want %v", plan.Reviewed, want)
	}
}

func TestAProposedNoteIsNeverReviewedEvenWhenDropped(t *testing.T) {
	c := newScanCase(t)
	c.note("n0.md", "https://n0.example/1")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "n0.md",
			"to": "Bare_Root", "reason": "x"}}})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 0 || len(plan.Reviewed) != 0 {
		t.Fatalf("moves %v reviewed %v", plan.Moves, plan.Reviewed)
	}
}

func TestAFailedCallReviewsNothing(t *testing.T) {
	c := newScanCase(t)
	c.note("n0.md", "https://n0.example/1")
	c.note("n1.md", "https://n1.example/1")
	plan := c.buildPlan(newRawLLM("not json at all").call)
	if len(plan.Reviewed) != 0 {
		t.Fatalf("reviewed %v", plan.Reviewed)
	}
}

func TestARaisingCallReviewsNothing(t *testing.T) {
	c := newScanCase(t)
	c.note("n0.md", "https://n0.example/1")
	c.note("n1.md", "https://n1.example/1")
	boom := &fakeLLM{panics: true}
	plan := c.buildPlan(boom.call)
	if len(plan.Reviewed) != 0 {
		t.Fatalf("reviewed %v", plan.Reviewed)
	}
}

func TestAValidatedMoveNeverRidesTheReviewList(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Palettes/r.md", "https://r.example/1")
	c.note("n0.md", "https://n0.example/1")
	c.note("n1.md", "https://n1.example/1")
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "n0.md",
			"to": "Design/Palettes", "reason": "fits"}}})
	plan := c.buildPlan(llm.call)
	if len(plan.Moves) != 1 || plan.Moves[0].Note != "n0.md" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if len(plan.Reviewed) != 1 ||
		plan.Reviewed[0] != filepath.Join(c.vault, "n1.md") {
		t.Fatalf("reviewed %v", plan.Reviewed)
	}
}

// --- TestTheSecondScan (test_censusstamp) ------------------------------

func settledVault(c *scanCase) {
	c.note("Design/Palettes/r.md", "https://r.example/1")
	c.note("A.md", "https://a.example/1")
	c.note("B.md", "https://b.example/1")
	c.note("C.md", "https://c.example/1")
}

func TestTheSecondScanSeesOnlyTheNewNotes(t *testing.T) {
	c := newScanCase(t)
	settledVault(c)
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "A.md",
			"to": "Design/Palettes", "reason": "fits"}}})
	plan := c.buildPlan(llm.call)
	rep := c.applyPlan(&plan)
	if rep.NotesMoved != 1 || rep.Stamped != 3 {
		t.Fatalf("report %+v (A moved; B, C reviewed)", rep)
	}
	c.note("NEW.md", "https://new.example/1")
	llm2 := newFakeLLM(map[string]any{"moves": []any{},
		"new_folders": []any{}})
	plan2 := c.buildPlan(llm2.call)
	if plan2.LLMCandidates != 1 {
		t.Fatalf("llm candidates %d", plan2.LLMCandidates)
	}
	if plan2.Inventory.CuratedNotes != 3 {
		t.Fatalf("curated %d", plan2.Inventory.CuratedNotes)
	}
	if len(plan2.Moves) != 0 {
		t.Fatalf("moves %v", plan2.Moves)
	}
	if len(plan2.Reviewed) != 1 ||
		plan2.Reviewed[0] != filepath.Join(c.vault, "NEW.md") {
		t.Fatalf("reviewed %v", plan2.Reviewed)
	}
}

func TestTheDeclinedVerdictHoldsAcrossScans(t *testing.T) {
	c := newScanCase(t)
	settledVault(c)
	llm := newFakeLLM(map[string]any{
		"moves": []any{
			map[string]any{"note": "A.md", "to": "Design/Palettes",
				"reason": "fits"},
			map[string]any{"note": "B.md", "to": "Design/Palettes",
				"reason": "fits"},
			map[string]any{"note": "C.md", "to": "Design/Palettes",
				"reason": "fits"},
		}})
	plan := c.buildPlan(llm.call)
	sel := Selection{Moves: map[string]bool{
		filepath.Join(c.vault, "A.md"): true}}
	c.applyPlan(&plan, withSelection(sel))
	llm2 := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "B.md",
			"to": "Design/Palettes", "reason": "again"}}})
	plan2 := c.buildPlan(llm2.call)
	if plan2.LLMCandidates != 0 || len(plan2.Moves) != 0 {
		t.Fatalf("plan2 %+v", plan2)
	}
	if plan2.Inventory.CuratedNotes != 3 {
		t.Fatalf("curated %d", plan2.Inventory.CuratedNotes)
	}
}

func TestAStaleLLMAnswerCannotRemoveAStampedNote(t *testing.T) {
	c := newScanCase(t)
	settledVault(c)
	llm := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "A.md",
			"to": "Design/Palettes", "reason": "fits"}}})
	plan := c.buildPlan(llm.call)
	c.applyPlan(&plan)
	llm2 := newFakeLLM(map[string]any{
		"moves": []any{map[string]any{"note": "A.md",
			"to": "Design/Palettes", "reason": "again"}}})
	plan2 := c.buildPlan(llm2.call)
	if len(plan2.Moves) != 0 {
		t.Fatalf("moves %+v", plan2.Moves)
	}
}

// --- TestTheLockPredicate (test_picklock) ------------------------------

func TestTheMetadataDoorReadsTruthyForLocks(t *testing.T) {
	for _, val := range []string{"true", "True", `"true"`, "yes", "1"} {
		if !noteLocked(&notereads.Inventory{Locked: truthy(val)}) {
			t.Fatalf("locked %q", val)
		}
		if !noteLocked(&notereads.Inventory{CategoryLocked: truthy(val)}) {
			t.Fatalf("category_locked %q", val)
		}
	}
}

func TestTheMetadataDoorReadsFalsyAsUnlocked(t *testing.T) {
	for _, val := range []string{"", "false", "no", "0"} {
		if noteLocked(&notereads.Inventory{Locked: truthy(val)}) {
			t.Fatalf("locked %q", val)
		}
		if noteLocked(&notereads.Inventory{CategoryLocked: truthy(val)}) {
			t.Fatalf("category_locked %q", val)
		}
	}
}

func TestTheTagsDoorFoldsCaseAndSpellingForLocks(t *testing.T) {
	for _, tag := range []string{"lock", "Lock", "LOCK", "locked",
		"Locked"} {
		if !noteLocked(&notereads.Inventory{Tags: []string{tag, "design"}}) {
			t.Fatalf("tag %q", tag)
		}
	}
	if noteLocked(&notereads.Inventory{Tags: []string{"notelock"}}) {
		t.Fatal("notelock locked")
	}
	if noteLocked(&notereads.Inventory{Tags: []string{"design", "tools"}}) {
		t.Fatal("design locked")
	}
}

func TestAnyDoorWinsForLocks(t *testing.T) {
	if !noteLocked(&notereads.Inventory{Locked: false,
		CategoryLocked: true}) {
		t.Fatal("category_locked did not win")
	}
	if !noteLocked(&notereads.Inventory{Locked: false,
		Tags: []string{"lock"}}) {
		t.Fatal("the tag did not win")
	}
}

func TestTheDirectParserReadsBothLockKeys(t *testing.T) {
	c := newScanCase(t)
	p := c.note("a.md", "https://a.example/1", withExtra("locked: true"))
	fm := parseNoteFrontmatter(p)
	if !fm.Locked || fm.CategoryLocked {
		t.Fatalf("fm %+v", fm)
	}
	p2 := c.note("b.md", "https://b.example/2",
		withExtra("category_locked: true"))
	fm2 := parseNoteFrontmatter(p2)
	if !fm2.CategoryLocked || fm2.Locked {
		t.Fatalf("fm2 %+v", fm2)
	}
	p3 := c.note("c.md", "https://c.example/3")
	fm3 := parseNoteFrontmatter(p3)
	if fm3.Locked || fm3.CategoryLocked {
		t.Fatalf("fm3 %+v", fm3)
	}
}

func TestTheCacheServesTheSameLockVerdict(t *testing.T) {
	c := newScanCase(t)
	p := c.note("a.md", "https://a.example/1", withTags("lock"),
		withExtra("category_locked: true"))
	cache := notereads.New("")
	defer cache.Close(false)
	direct := parseNoteFrontmatter(p)
	cached := cache.Frontmatter(p)
	if cached.Locked != direct.Locked ||
		cached.CategoryLocked != direct.CategoryLocked {
		t.Fatalf("direct %+v cached %+v", direct, cached)
	}
	if !noteLocked(&cached) {
		t.Fatal("the cached parse did not lock")
	}
}

// --- TestTheLocksRadar (test_picklock) ---------------------------------

func TestTheInventoryCountsLocksAndTheLogTeaches(t *testing.T) {
	c := newScanCase(t)
	c.note("orphan.md", "https://o.example/1", withExtra("locked: true"))
	c.note("orphan2.md", "https://o.example/2")
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	if inv.LockedNotes != 1 {
		t.Fatalf("locked %d", inv.LockedNotes)
	}
	byFile := map[string]NoteRow{}
	for _, n := range inv.Notes {
		byFile[n.File] = n
	}
	if !byFile["orphan.md"].Locked || byFile["orphan2.md"].Locked {
		t.Fatal("lock flags wrong")
	}
	if !c.logs.has("out of the scan") {
		t.Fatal("missing lock log")
	}
}

func TestALockedNoteIsNeverAMoveCandidate(t *testing.T) {
	c := newScanCase(t)
	c.note("locked-root.md", "https://r.example/1",
		withExtra("locked: true"))
	c.note("Uncategorized/parked.md", "https://u.example/2",
		withTags("lock"))
	c.note("Design/stray.md", "https://s.example/3", withCategory("Design"),
		withExtra("category_locked: true"))
	c.note("plain-root.md", "https://p.example/4")
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	cands := MoveCandidates(&inv, 40)
	if len(cands) != 1 || cands[0].File != "plain-root.md" {
		t.Fatalf("candidates %+v", cands)
	}
}

func TestALockedNoteNeverRidesTheLLMAsk(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Agents/settled.md", "https://settled.example/1")
	c.note("locked.md", "https://l.example/1",
		withTags("design", "agents"), withExtra("locked: true"))
	c.note("free.md", "https://f.example/2")
	llm := func(messages []llmclient.Message, task string) string {
		digest := messages[1].Content
		if !strings.Contains(digest, " - free — f.example") {
			t.Fatal("the free note's digest line missing")
		}
		if strings.Contains(digest, " - locked — l.example") ||
			strings.Contains(digest, "locked.md") {
			t.Fatal("the locked note rode the digest")
		}
		return `{"new_folders": [], "moves": [{"note": "free.md",` +
			` "to": "Design/Agents", "reason": "its tags say so",` +
			` "confidence": "high"}], "summary": "file the free one"}`
	}
	plan := c.buildPlan(llm)
	if len(plan.Moves) != 1 || plan.Moves[0].Note != "free.md" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if plan.LockedNotes != 1 || plan.Inventory.LockedNotes != 1 {
		t.Fatalf("plan %+v", plan)
	}
}

func TestTheDeclaredHomePassSkipsLockedNotes(t *testing.T) {
	c := newScanCase(t)
	if err := os.MkdirAll(filepath.Join(c.vault, "Design", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	c.note("Design/locked-stray.md", "https://d.example/1",
		withCategory("Design"), withSubcategory("Agents"),
		withExtra("locked: true"))
	c.note("Design/free-stray.md", "https://d.example/2",
		withCategory("Design"), withSubcategory("Agents"))
	plan := c.buildPlan(nil)
	var moved []string
	for _, m := range plan.Moves {
		moved = append(moved, m.Note)
	}
	sort.Strings(moved)
	if len(moved) != 1 || moved[0] != "free-stray.md" {
		t.Fatalf("moved %v", moved)
	}
}

func TestLockedDuplicatesNeverMergeAndNeverKeep(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Assets/keep.md", "https://dup.example/",
		withExtra("locked: true"))
	c.note("Design/Assets/d1.md", "https://dup.example/tools")
	c.note("Design/Assets/d2.md", "https://dup.example/privacy")
	plan := c.buildPlan(nil)
	if len(plan.Consolidations) != 1 {
		t.Fatalf("consolidations %d", len(plan.Consolidations))
	}
	cons := plan.Consolidations[0]
	if cons.Keep.File != "d1.md" {
		t.Fatalf("keep %q (the free one)", cons.Keep.File)
	}
	for _, m := range cons.Merge {
		if m.File == "keep.md" {
			t.Fatal("the locked twin rode the merge")
		}
	}
	// a lone locked duplicate beside a free keeper: nothing proposes
	// touching the locked twin
	tmp2 := t.TempDir()
	vault2 := filepath.Join(tmp2, "w")
	if err := os.MkdirAll(vault2, 0o755); err != nil {
		t.Fatal(err)
	}
	writeNote(t, vault2, "free.md", "https://solo.example/", noteSpec{})
	writeNote(t, vault2, "locked-twin.md", "https://solo.example/x",
		noteSpec{tags: []string{"locked"}})
	inv2 := ScanVaultInventory(vault2, nil, nil, nil, VaultKindWebsites)
	if props := ScanSiteDuplicates(&inv2, nil, VaultKindWebsites); len(props) != 0 {
		t.Fatalf("props %+v", props)
	}
}

func TestALockedNoteNeverRidesTheDeletionList(t *testing.T) {
	c := newScanCase(t)
	c.note("doomed.md", "https://doom.example/1", withTags("auto_delete"))
	c.note("locked-doomed.md", "https://doom.example/2",
		withTags("auto_delete", "locked"))
	plan := c.buildPlan(nil)
	if len(plan.Deletions) != 1 {
		t.Fatalf("deletions %d", len(plan.Deletions))
	}
	if plan.Deletions[0].URL != "https://doom.example/1" {
		t.Fatalf("deletion %+v", plan.Deletions[0])
	}
	if !c.logs.has("delete mark waits") {
		t.Fatal("missing lock log")
	}
}

// --- TestTheLockAtApply (test_picklock) --------------------------------

func planWithMove(path, to string) *Plan {
	return &Plan{
		Moves: []MoveRow{{
			Note: filepath.Base(path), Path: path,
			Title: strings.TrimSuffix(filepath.Base(path), ".md"),
			From:  "(vault root)", To: to, Reason: "test",
		}},
	}
}

func TestAMoveLockedMidplanIsDropped(t *testing.T) {
	c := newScanCase(t)
	p := c.note("later.md", "https://l.example/1")
	plan := planWithMove(p, "Design/Agents")
	if err := os.MkdirAll(filepath.Join(c.vault, "Design", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	// the owner locks it while the dialog stands open
	text := readFileT(t, p)
	text = strings.Replace(text, "managed_by:", "locked: true\nmanaged_by:", 1)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	rep := c.applyPlan(plan)
	if rep.NotesMoved != 0 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, p)
	if !c.logs.has("the move is dropped") {
		t.Fatal("missing drop log")
	}
}

func TestADeletionLockedMidplanWaits(t *testing.T) {
	c := newScanCase(t)
	p := c.note("doomed.md", "https://doom.example/1",
		withTags("auto_delete"))
	plan := c.buildPlan(nil)
	if len(plan.Deletions) != 1 {
		t.Fatalf("deletions %d", len(plan.Deletions))
	}
	text := readFileT(t, p)
	text = strings.Replace(text, "managed_by:", "locked: true\nmanaged_by:", 1)
	if err := os.WriteFile(p, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	rep := c.applyPlan(&plan)
	if rep.Banished != 0 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, p)
	if !c.logs.has("deletion waits") {
		t.Fatal("missing waits log")
	}
}

func TestALockedKeeperNeverReceivesTheMerge(t *testing.T) {
	c := newScanCase(t)
	if err := os.MkdirAll(filepath.Join(c.vault, "Design"), 0o755); err != nil {
		t.Fatal(err)
	}
	keep := c.note("Design/keep.md", "https://dup.example/")
	twin := c.note("Design/twin.md", "https://dup.example/x")
	plan := &Plan{
		Consolidations: []Consolidation{{
			Site: "dup.example",
			Keep: KeepRow{File: "keep.md", Path: keep,
				Source: "https://dup.example/"},
			Merge: []MergeRow{{File: "twin.md", Path: twin,
				Source: "https://dup.example/x"}},
			URLs: []string{"https://dup.example/x"},
		}},
	}
	text := readFileT(t, keep)
	text = strings.Replace(text, "managed_by:", "locked: true\nmanaged_by:", 1)
	if err := os.WriteFile(keep, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	rep := c.applyPlan(plan)
	if rep.Consolidated != 0 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, twin)
	if !c.logs.has("the consolidation waits") {
		t.Fatal("missing waits log")
	}
}

// --- TestThePickedApply (test_picklock) --------------------------------

func pickedVault(c *scanCase) (string, string, string, string) {
	c.t.Helper()
	if err := os.MkdirAll(filepath.Join(c.vault, "Design", "Agents"),
		0o755); err != nil {
		c.t.Fatal(err)
	}
	m1 := c.note("m1.md", "https://one.example/1", withTags("agents"))
	m2 := c.note("m2.md", "https://two.example/2", withTags("color"))
	c.note("d1.md", "https://dead1.example/1", withTags("auto_delete"))
	c.note("d2.md", "https://dead2.example/2", withTags("auto_delete"))
	return m1, m2, "", ""
}

func TestOnlyThePickedMovesRide(t *testing.T) {
	c := newScanCase(t)
	m1, m2, _, _ := pickedVault(c)
	plan := &Plan{
		Moves: []MoveRow{
			{Note: "m1.md", Path: m1, Title: "m1",
				From: "(vault root)", To: "Design/Agents", Reason: ""},
			{Note: "m2.md", Path: m2, Title: "m2",
				From: "(vault root)", To: "Design/Agents", Reason: ""},
		},
	}
	sel := Selection{
		Moves:      map[string]bool{m1: true},
		NewFolders: map[string]bool{},
	}
	rep := c.applyPlan(plan, withSelection(sel))
	if rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeGone(t, m1)
	mustBeFile(t, filepath.Join(c.vault, "Design", "Agents", "m1.md"))
	mustBeFile(t, m2)
}

func TestAnUnpickedFolderIsNeverBorn(t *testing.T) {
	c := newScanCase(t)
	plan := &Plan{NewFolders: []string{"A/B", "C/D"}}
	sel := Selection{NewFolders: map[string]bool{"C/D": true},
		Moves: map[string]bool{}}
	rep := c.applyPlan(plan, withSelection(sel))
	if rep.FoldersCreated != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeGone(t, filepath.Join(c.vault, "A"))
	mustBeDir(t, filepath.Join(c.vault, "C", "D"))
}

func TestOnlyThePickedDuplicatesMergeAndTheURLsFollow(t *testing.T) {
	c := newScanCase(t)
	keep := c.note("Design/keep.md", "https://dup.example/")
	t1 := c.note("Design/t1.md", "https://dup.example/a")
	t2 := c.note("Design/t2.md", "https://dup.example/b")
	plan := &Plan{
		Consolidations: []Consolidation{{
			Site: "dup.example",
			Keep: KeepRow{File: "keep.md", Path: keep,
				Source: "https://dup.example/", Category: "Design"},
			Merge: []MergeRow{
				{File: "t1.md", Path: t1, Source: "https://dup.example/a"},
				{File: "t2.md", Path: t2, Source: "https://dup.example/b"},
			},
			URLs: []string{"https://dup.example/a",
				"https://dup.example/b"},
		}},
	}
	sel := Selection{Consolidations: map[string]bool{t1: true},
		Moves: map[string]bool{}}
	rep := c.applyPlan(plan, withSelection(sel))
	if rep.Consolidated != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeGone(t, t1)
	mustBeFile(t, t2)
	kept := readFileT(t, keep)
	if !strings.Contains(kept, "https://dup.example/a") {
		t.Fatal("the picked twin's URL missing")
	}
	if strings.Contains(kept, "https://dup.example/b") {
		t.Fatal("the denied twin's URL rode")
	}
}

func TestTheUnpickedDeletionsWait(t *testing.T) {
	c := newScanCase(t)
	d1 := c.note("d1.md", "https://dead1.example/1", withTags("auto_delete"))
	d2 := c.note("d2.md", "https://dead2.example/2", withTags("auto_delete"))
	plan := c.buildPlan(nil)
	if len(plan.Deletions) != 2 {
		t.Fatalf("deletions %d", len(plan.Deletions))
	}
	del := map[string]bool{}
	for _, d := range plan.Deletions {
		if strings.Contains(d.URL, "/2") {
			del[d.Canonical] = true
		}
	}
	sel := Selection{Deletions: del, Moves: map[string]bool{}}
	rep := c.applyPlan(&plan, withSelection(sel))
	if rep.Banished != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, d1)
	mustBeGone(t, d2)
}

func TestNoneSelectionKeepsTheWholePlanLaw(t *testing.T) {
	c := newScanCase(t)
	c.note("d1.md", "https://dead1.example/1", withTags("auto_delete"))
	c.note("d2.md", "https://dead2.example/2", withTags("auto_delete"))
	c.note("Design/keep.md", "https://dup.example/")
	c.note("Design/t1.md", "https://dup.example/a")
	c.note("Design/t2.md", "https://dup.example/b")
	plan := c.buildPlan(nil)
	if len(plan.Deletions) != 2 || len(plan.Consolidations) != 1 {
		t.Fatalf("plan %+v", plan)
	}
	rep := c.applyPlan(&plan)
	if rep.Banished != 2 || rep.Consolidated != 2 {
		t.Fatalf("report %+v", rep)
	}
}

// --- TestTheConfidenceLaw (test_picklock) ------------------------------

func TestALowConfidenceWordFlagsTheMove(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/Agents/settled.md", "https://settled.example/1")
	c.note("unsure.md", "https://u.example/1")
	c.note("sure.md", "https://s.example/2")
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{
			map[string]any{"note": "unsure.md", "to": "Design/Agents",
				"reason": "thin signals", "confidence": "low"},
			map[string]any{"note": "sure.md", "to": "Design/Agents",
				"reason": "clear", "confidence": "high"},
			map[string]any{"note": "silent.md", "to": "Design/Agents",
				"reason": "no word"},
		},
		"summary": "mixed",
	})
	plan := c.buildPlan(llm.call)
	flags := map[string]bool{}
	for _, m := range plan.Moves {
		flags[m.Note] = m.LowConfidence
	}
	// 'silent.md' was never a candidate — only the two real notes
	if len(flags) != 2 || flags["unsure.md"] != true ||
		flags["sure.md"] != false {
		t.Fatalf("flags %v", flags)
	}
}

func TestTheDigestTeachesTheLLMItsOwnMaterial(t *testing.T) {
	c := newScanCase(t)
	c.note("Design/stray.md", "https://s.example/1", withCategory("Design"),
		withSubcategory("Color Tools"), withTags("palette"))
	inv := ScanVaultInventory(c.vault, nil, nil, nil, VaultKindWebsites)
	cands := MoveCandidates(&inv, 40)
	lines := DigestLines(cands, VaultKindWebsites)
	if len(lines) != 1 {
		t.Fatalf("lines %d", len(lines))
	}
	for _, needle := range []string{"why: a stray at its category root",
		"declares: Design / Color Tools", "tags: palette"} {
		if !strings.Contains(lines[0], needle) {
			t.Fatalf("line %q missing %q", lines[0], needle)
		}
	}
	if got := CandidateReason(&NoteRow{InRoot: true}); got !=
		"orphaned at the vault root" {
		t.Fatalf("reason %q", got)
	}
	if got := CandidateReason(&NoteRow{Folder: "Uncategorized"}); got !=
		"parked in Uncategorized" {
		t.Fatalf("reason %q", got)
	}
	if got := CandidateReason(&NoteRow{Folder: "Design"}); got !=
		"a stray at its category root (no subfolder of its own)" {
		t.Fatalf("reason %q", got)
	}
	if got := CandidateReason(&NoteRow{Folder: "Design/Color_Tools"}); got !=
		"filed in a too-broad folder (Design/Color_Tools)" {
		t.Fatalf("reason %q", got)
	}
}

// --- TestTheSelectionLaw (test_picklock) -------------------------------

var pickPlan = Plan{
	Deletions: []websitepipeline.PendingBanishment{
		{Canonical: "https://a.example/1", URL: "https://a.example/1"},
		{Canonical: "https://b.example/2", URL: "https://b.example/2"},
	},
	NewFolders: []string{"X/Y", "Z/W"},
	Moves: []MoveRow{
		{Path: "/v/m1.md", Note: "m1.md"},
		{Path: "/v/m2.md", Note: "m2.md"},
	},
	Consolidations: []Consolidation{
		{Site: "s.example",
			Merge: []MergeRow{{Path: "/v/k1.md"}, {Path: "/v/k2.md"}}},
		{Site: "t.example",
			Merge: []MergeRow{{Path: "/v/j1.md"}}},
	},
}

func TestNoneSelectionRidesWhole(t *testing.T) {
	sel := NormalizeSelection(nil)
	if sel.Deletions != nil || sel.Consolidations != nil ||
		sel.NewFolders != nil || sel.Moves != nil {
		t.Fatalf("sel %+v", sel)
	}
	counts := PickedCounts(&pickPlan, sel)
	if counts["deletions"] != 2 || counts["new_folders"] != 2 ||
		counts["moves"] != 2 || counts["consolidation_notes"] != 3 ||
		counts["consolidations"] != 2 {
		t.Fatalf("counts %v", counts)
	}
}

func TestAMissingKeyReadsAsTheWholeSection(t *testing.T) {
	sel := NormalizeSelection(map[string][]string{
		"moves": {"/v/m1.md"}})
	if sel.Deletions != nil {
		t.Fatal("deletions should ride whole")
	}
	counts := PickedCounts(&pickPlan, sel)
	if counts["deletions"] != 2 || counts["moves"] != 1 ||
		counts["consolidation_notes"] != 3 {
		t.Fatalf("counts %v", counts)
	}
}

func TestAnExplicitEmptyListRidesNothing(t *testing.T) {
	sel := NormalizeSelection(map[string][]string{
		"deletions": {}, "moves": {}, "new_folders": {},
		"consolidations": {}})
	counts := PickedCounts(&pickPlan, sel)
	if counts["deletions"] != 0 || counts["new_folders"] != 0 ||
		counts["moves"] != 0 || counts["consolidation_notes"] != 0 ||
		counts["consolidations"] != 0 {
		t.Fatalf("counts %v", counts)
	}
}

func TestThePickedArithmeticCountsEachShape(t *testing.T) {
	sel := NormalizeSelection(map[string][]string{
		"deletions":      {"https://b.example/2"},
		"new_folders":    {"Z/W"},
		"moves":          {"/v/m2.md"},
		"consolidations": {"/v/k2.md"},
	})
	counts := PickedCounts(&pickPlan, sel)
	if counts["deletions"] != 1 || counts["new_folders"] != 1 ||
		counts["moves"] != 1 || counts["consolidation_notes"] != 1 ||
		counts["consolidations"] != 1 {
		t.Fatalf("counts %v", counts)
	}
}

func TestAPartiallyPickedSiteStillCountsItsSite(t *testing.T) {
	sel := NormalizeSelection(map[string][]string{
		"consolidations": {"/v/k1.md", "/v/k2.md", "/v/j1.md"}})
	counts := PickedCounts(&pickPlan, sel)
	if counts["consolidation_notes"] != 3 || counts["consolidations"] != 2 {
		t.Fatalf("counts %v", counts)
	}
}

var _ = storage.PyDumps
var _ = os.WriteFile
