// scandialogs_test.go — T-76's workflow tests + selection semantics
// (the spec's own list): Picker — the display shape, the status lines,
// the live arithmetic, the safe-cancel scope; Plan — the section
// grammars, the freeze guard, the low-confidence pre-uncheck, the
// master sweep, the Apply label + the arm-guard's enable law, the
// four-key selection; Scorecard — the K-of-N tones, the named
// failures, the heading law, the per-vault sections, the hygiene row;
// plus the story extension's own laws (the picked-counts replacement,
// the repos promotion, the GUI door's ask line).
package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/repobanish"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaulthygiene"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// ---------------------------------------------------------------------------
// THE PICKER.
// ---------------------------------------------------------------------------

func pickerRows(t *testing.T) []SyncVaultRow {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	wv := t.TempDir()
	rv := t.TempDir()
	n := 7
	writeNote(t, wv, "Agents/one.md", "https://github.com/octo/one")
	_ = n
	cfg := storage.NewOrderedMap()
	cfg.Set("website_vault_path", wv)
	cfg.Set("vault_path", rv)
	cfg.Set("manual_vault_path", filepath.Join(t.TempDir(), "personal"))
	return DetectScanVaults(cfg, true)
}

func TestPickDisplayModelTheThreeRowsAndStatusLines(t *testing.T) {
	rows := pickerRows(t)
	model := PickDisplayModelOf(rows)
	if model.Title != "Choose the vault to scan" {
		t.Fatalf("title = %q", model.Title)
	}
	if !strings.Contains(model.Header, "only the ticked library's plan opens for your review") {
		t.Fatalf("header = %q", model.Header)
	}
	if len(model.Rows) != 3 {
		t.Fatalf("rows = %d, want the three vaults", len(model.Rows))
	}
	// the order is the Settings page's own (websites, repos, personal).
	if model.Rows[0].Kind != "websites" || model.Rows[1].Kind != "repos" ||
		model.Rows[2].Kind != "personal" {
		t.Fatalf("order = %v %v %v", model.Rows[0].Kind, model.Rows[1].Kind, model.Rows[2].Kind)
	}
	// the two machine vaults are tickable; Personal is honestly read-only.
	if !model.Rows[0].Selectable || !model.Rows[1].Selectable {
		t.Fatal("a connected machine vault must be tickable")
	}
	if model.Rows[2].Scannable || model.Rows[2].Selectable {
		t.Fatal("the Personal vault must never be tickable")
	}
	if !strings.Contains(model.Rows[2].Why, "Owner-owned") {
		t.Fatalf("personal why = %q", model.Rows[2].Why)
	}
	// the status lines speak the connection truth in the owner's voice.
	if !strings.HasSuffix(model.Rows[0].StatusText, "note(s)") {
		t.Fatalf("connected status = %q, want the note count", model.Rows[0].StatusText)
	}
	if model.Rows[1].StatusText != "note(s)" && !strings.HasSuffix(model.Rows[1].StatusText, "note(s)") {
		t.Fatalf("repos status = %q", model.Rows[1].StatusText)
	}
	if model.AnySelectable != true || len(model.SelectableKinds) != 2 {
		t.Fatalf("selectable kinds = %v", model.SelectableKinds)
	}
	if model.Hint != "" {
		t.Fatalf("a connected desk must not speak the hint: %q", model.Hint)
	}
}

func TestPickDisplayModelNotSetAndHintLaws(t *testing.T) {
	cfg := storage.NewOrderedMap() // nothing set
	model := PickDisplayModelOf(DetectScanVaults(cfg, true))
	for _, row := range model.Rows {
		if row.StatusText != "Not set — Settings → 📁 Vault" {
			t.Fatalf("not-set status = %q", row.StatusText)
		}
		if row.Selectable {
			t.Fatalf("an unset vault must not be tickable: %+v", row)
		}
	}
	if model.AnySelectable {
		t.Fatal("nothing selectable")
	}
	if !strings.Contains(model.Hint, "No scan-able vault is connected yet") {
		t.Fatalf("hint = %q", model.Hint)
	}
	// a not-found vault speaks its own miss.
	wv := filepath.Join(t.TempDir(), "missing-vault")
	cfg2 := storage.NewOrderedMap()
	cfg2.Set("website_vault_path", wv)
	model2 := PickDisplayModelOf(DetectScanVaults(cfg2, false))
	if model2.Rows[0].StatusText != "Not found at this path" {
		t.Fatalf("not-found status = %q", model2.Rows[0].StatusText)
	}
	if !strings.Contains(model2.Rows[0].Why, "The folder is not found at this path") {
		t.Fatalf("not-found why = %q", model2.Rows[0].Why)
	}
}

func TestPickScanButtonLiveArithmetic(t *testing.T) {
	if got := PickScanButtonLabel(nil); got != "🔍 Scan" {
		t.Fatalf("none = %q", got)
	}
	if got := PickScanButtonLabel([]string{"repos"}); got != "🔍 Scan this vault" {
		t.Fatalf("one = %q", got)
	}
	if got := PickScanButtonLabel([]string{"websites"}); got != "🔍 Scan this vault" {
		t.Fatalf("one websites = %q", got)
	}
	// the count is order-tolerant and case-tolerant — the box order is
	// not the display order.
	if got := PickScanButtonLabel([]string{"repos", "websites"}); got != "🔍 Scan both vaults" {
		t.Fatalf("both = %q", got)
	}
	if got := PickScanButtonLabel([]string{"Websites", "REPOS"}); got != "🔍 Scan both vaults" {
		t.Fatalf("both (case) = %q", got)
	}
}

func TestPickedScanKindsSafeCancelAndScopeLaw(t *testing.T) {
	// the safe cancel: no choice, no scan (the caller never even builds
	// a worker).
	if got := PickedScanKinds(nil); got != nil {
		t.Fatalf("nil picked = %v, want the safe cancel", got)
	}
	if got := PickedScanKinds([]string{}); got != nil {
		t.Fatalf("empty picked = %v, want the safe cancel", got)
	}
	// junk falls away; the display order wins.
	if got := PickedScanKinds([]string{"repos", "personal", "junk"}); len(got) != 1 || got[0] != "repos" {
		t.Fatalf("junk scope = %v", got)
	}
	if got := PickedScanKinds([]string{"Repos", " websites "}); len(got) != 2 ||
		got[0] != "websites" || got[1] != "repos" {
		t.Fatalf("scope = %v, want websites first", got)
	}
}

// ---------------------------------------------------------------------------
// THE PLAN MODAL — the display model.
// ---------------------------------------------------------------------------

func samplePlan() *vaultscan.Plan {
	return &vaultscan.Plan{
		VaultKind:  "websites",
		VaultLabel: "Websites vault",
		Deletions:  []websitepipeline.PendingBanishment{{URL: "https://example.com/a", Canonical: "https://example.com/a", Title: "A note", Marker: "frontmatter: auto-delete", Door: "note tag"}},
		NewFolders: []string{"Tools/Scraping"},
		Moves: []vaultscan.MoveRow{
			{Note: "B.md", Path: "B.md", From: "", To: "Tools/Scraping", Reason: "the filing eyes", LowConfidence: false},
			{Note: "C.md", Path: "C.md", From: "Agents", To: "Tools/Scraping", As: "C_v1.md", Reason: "a colliding landing name", LowConfidence: true},
		},
		Consolidations: []vaultscan.Consolidation{
			{Site: "example.com", Keep: vaultscan.KeepRow{File: "Example.md"},
				Merge: []vaultscan.MergeRow{{File: "Example 2.md", Path: "Example 2.md", Source: "https://example.com/x"}}},
		},
		Summary:         "the LLM's own summary",
		KeptHandwritten: 2,
		Inventory: vaultscan.PlanSummary{TotalNotes: 12, FolderCount: 3,
			RootNotes: 2, UncategorizedNotes: 4, TrashNotes: 1, LockedNotes: 5},
	}
}

func TestPlanDisplayModelSectionsAndGrammars(t *testing.T) {
	model := PlanDisplayModelOf(samplePlan())
	if model.Title != "Vault scan review — Websites vault" {
		t.Fatalf("title = %q", model.Title)
	}
	if !strings.Contains(model.Header, "The whole plan, before anything is touched. Untick anything you deny") {
		t.Fatalf("header = %q", model.Header)
	}
	if !strings.Contains(model.LibraryLine, "🏛️ The library (Websites vault): 12 note(s) in 3 folder(s) — 2 orphaned, 4 uncategorized") {
		t.Fatalf("library line = %q", model.LibraryLine)
	}
	if !strings.Contains(model.TrashLine, "1 note(s) rest in the Trash folder — they ride the deletions below") {
		t.Fatalf("trash line = %q", model.TrashLine)
	}
	if !strings.Contains(model.LockedLine, "5 locked note(s) are out of the radar") {
		t.Fatalf("locked line = %q", model.LockedLine)
	}
	if !strings.Contains(model.KeptLine, "2 hand-written note(s) carry marks — kept") {
		t.Fatalf("kept line = %q", model.KeptLine)
	}
	// the four sections, in the dialog's own order.
	if len(model.Sections) != 4 {
		t.Fatalf("sections = %d", len(model.Sections))
	}
	wantKeys := []string{"deletions", "consolidations", "new_folders", "moves"}
	for i, want := range wantKeys {
		if model.Sections[i].Key != want {
			t.Fatalf("section %d = %s, want %s", i, model.Sections[i].Key, want)
		}
	}
	del := model.Sections[0]
	if del.Heading != "🗑️ Marked for deletion — 1" || del.Card != "plan_delete_card" || del.Tone != "danger" {
		t.Fatalf("deletion section = %+v", del)
	}
	if del.Rows[0].Text != "🗑️ A note — via your mark" {
		t.Fatalf("deletion line = %q", del.Rows[0].Text)
	}
	if !strings.Contains(del.Rows[0].Full, "marker: frontmatter: auto-delete · door: note tag") {
		t.Fatalf("deletion full = %q", del.Rows[0].Full)
	}
	if del.Rows[0].ID != "https://example.com/a" {
		t.Fatalf("deletion id = %q (the canonical URL is the selection key)", del.Rows[0].ID)
	}
	cons := model.Sections[1]
	if cons.Heading != "🧲 Duplicate notes to merge (one note per site) — 1" {
		t.Fatalf("consolidation heading = %q", cons.Heading)
	}
	if cons.Rows[0].Text != "🧲 Example 2.md → merges into Example.md (site: example.com)" {
		t.Fatalf("consolidation line = %q", cons.Rows[0].Text)
	}
	if !strings.Contains(cons.Rows[0].Full, "(the duplicate rests in .trash/consolidated — recoverable by hand; the site is wanted, nothing is blacklisted)") {
		t.Fatalf("consolidation full = %q", cons.Rows[0].Full)
	}
	if cons.Rows[0].ID != "Example 2.md" {
		t.Fatalf("consolidation id = %q (the merged note's path is the key)", cons.Rows[0].ID)
	}
	folders := model.Sections[2]
	if folders.Rows[0].Text != "📁 Tools/Scraping" || folders.Rows[0].ID != "Tools/Scraping" {
		t.Fatalf("folder row = %+v", folders.Rows[0])
	}
	moves := model.Sections[3]
	// THE LANDABLE MOVE: the bent name is SPOKEN in the row itself.
	if moves.Rows[0].Text != "🚚 B.md: (vault root) → Tools/Scraping" {
		t.Fatalf("move line = %q", moves.Rows[0].Text)
	}
	if moves.Rows[1].Text != "🚚 C.md: Agents → Tools/Scraping (as C_v1.md)" {
		t.Fatalf("landable move line = %q", moves.Rows[1].Text)
	}
	if !strings.Contains(moves.Rows[1].Full, "why: a colliding landing name") {
		t.Fatalf("move full = %q", moves.Rows[1].Full)
	}
	// the has-flags + the low-confidence teaching line.
	if !model.HasDeletions || !model.HasFiling || !model.HasConsolidations || !model.HasLowConfidence {
		t.Fatalf("flags = %+v", model)
	}
	if !strings.Contains(model.LowConfidenceLine, "the LLM's own low-confidence calls") {
		t.Fatalf("low-confidence line = %q", model.LowConfidenceLine)
	}
	if model.Summary != "the LLM's own summary" {
		t.Fatalf("summary = %q", model.Summary)
	}
}

func TestPlanDisplayModelDoorWordsAndVaultLabelLaw(t *testing.T) {
	plan := samplePlan()
	plan.Deletions[0].Door = "trash folder"
	model := PlanDisplayModelOf(plan)
	if model.Sections[0].Rows[0].Text != "🗑️ A note — via you moved it to Trash" {
		t.Fatalf("trash-door line = %q", model.Sections[0].Rows[0].Text)
	}
	plan2 := samplePlan()
	plan2.Deletions[0].Door = "master table"
	model2 := PlanDisplayModelOf(plan2)
	if model2.Sections[0].Rows[0].Text != "🗑️ A note — via master table" {
		t.Fatalf("master-door line = %q", model2.Sections[0].Rows[0].Text)
	}
	// the vault-kind door: an unlabeled repos plan names its library.
	plan3 := samplePlan()
	plan3.VaultKind = "repos"
	plan3.VaultLabel = ""
	plan3.DeletionsRepos = []repobanish.PendingRepoBanishment{
		{URL: "https://github.com/octo/a", Canonical: "https://github.com/octo/a",
			Title: "A note", Marker: "frontmatter: auto-delete", Door: "note tag"}}
	plan3.Deletions = nil
	plan3.KeptHandwrittenR = 3
	plan3.KeptHandwritten = 0
	model3 := PlanDisplayModelOf(plan3)
	if model3.Title != "Vault scan review — GitHub Repos vault" {
		t.Fatalf("repos title = %q", model3.Title)
	}
	if !strings.Contains(model3.KeptLine, "3 hand-written note(s)") {
		t.Fatalf("repos kept line = %q (KeptHandwrittenR is the repos seat)", model3.KeptLine)
	}
	if len(model3.Sections[0].Rows) != 1 {
		t.Fatalf("repos deletions = %d", len(model3.Sections[0].Rows))
	}
}

func TestPlanDisplayModelFreezeGuard(t *testing.T) {
	plan := samplePlan()
	plan.NewFolders = nil
	for i := 0; i < PlanMaxRows+100; i++ {
		plan.NewFolders = append(plan.NewFolders, "F/folder-"+string(rune('a'+i%26))+string(rune('a'+i/26)))
	}
	model := PlanDisplayModelOf(plan)
	folders := model.Sections[2]
	if len(folders.Rows) != PlanMaxRows {
		t.Fatalf("rows = %d, want the %d cap", len(folders.Rows), PlanMaxRows)
	}
	if folders.Total != PlanMaxRows+100 || folders.More != 100 {
		t.Fatalf("total/more = %d/%d", folders.Total, folders.More)
	}
	if !strings.Contains(folders.Heading, "— 600") {
		t.Fatalf("heading = %q (the total speaks the WHOLE pile)", folders.Heading)
	}
}

// ---------------------------------------------------------------------------
// THE PICKING DOOR — the selection semantics.
// ---------------------------------------------------------------------------

func TestPlanPickStateBornLawAndLowConfidencePreUncheck(t *testing.T) {
	state := NewPlanPickState(PlanDisplayModelOf(samplePlan()))
	// checked by default…
	if !state.IsChecked(ScanSectionDeletions, "https://example.com/a") {
		t.Fatal("a normal row must be born checked")
	}
	if !state.IsChecked(ScanSectionMoves, "B.md") {
		t.Fatal("a normal move must be born checked")
	}
	// …except the LLM's own low-confidence call (pre-unchecked).
	if state.IsChecked(ScanSectionMoves, "C.md") {
		t.Fatal("a low-confidence move must be born UNCHECKED")
	}
	// the plan holds 5 rows (1 deletion + 1 merge note + 1 folder + 2
	// moves); the low-confidence row is denied at birth.
	picked, total := state.PickedTotals()
	if picked != 4 || total != 5 {
		t.Fatalf("picked/total = %d/%d, want 4 of 5 (the low-confidence row denied)", picked, total)
	}
	if state.ApplyLabel() != "🗂️ Apply plan (4 of 5)" {
		t.Fatalf("label = %q", state.ApplyLabel())
	}
}

func TestPlanPickStateMasterSweepAndMirrors(t *testing.T) {
	state := NewPlanPickState(PlanDisplayModelOf(samplePlan()))
	// the moves panel disagrees (2 rows, 1 unchecked) — partial.
	if state.MasterState(ScanSectionMoves) != MasterPartial {
		t.Fatalf("moves master = %q, want partial", state.MasterState(ScanSectionMoves))
	}
	// the sweep: any unticked row ticks the whole panel (the moves
	// hold 2 rows, one denied at birth).
	state.ToggleSection(ScanSectionMoves)
	if state.MasterState(ScanSectionMoves) != MasterChecked {
		t.Fatalf("after sweep = %q", state.MasterState(ScanSectionMoves))
	}
	picked, _ := state.PickedTotals()
	if picked != 5 {
		t.Fatalf("picked = %d, want the whole plan (5 rows)", picked)
	}
	// an all-ticked panel unticks entirely (the one gesture).
	state.ToggleSection(ScanSectionMoves)
	if state.MasterState(ScanSectionMoves) != MasterUnchecked {
		t.Fatalf("after untick sweep = %q", state.MasterState(ScanSectionMoves))
	}
	picked, _ = state.PickedTotals()
	if picked != 3 {
		t.Fatalf("picked = %d, want the moves gone", picked)
	}
	// the per-row tick still works (the owner ticks one back).
	state.SetChecked(ScanSectionMoves, "C.md", true)
	if !state.IsChecked(ScanSectionMoves, "C.md") {
		t.Fatal("the tick never landed")
	}
}

func TestPlanPickStateArmGuardEnableLaw(t *testing.T) {
	model := PlanDisplayModelOf(samplePlan())
	state := NewPlanPickState(model)
	// THE FRESH-CLICK DOOR: born unarmed — both verdict buttons dead
	// whatever the arithmetic says.
	if state.Armed() {
		t.Fatal("the door must be born unarmed")
	}
	if state.ApplyEnabled() {
		t.Fatal("an unarmed Apply must stay dead")
	}
	if state.KeepEnabled() {
		t.Fatal("an unarmed Keep must stay dead")
	}
	// the arm beat ends: the doors come alive through the same
	// arithmetic.
	state.Arm()
	if !state.ApplyEnabled() || !state.KeepEnabled() {
		t.Fatal("an armed door with picks must be alive")
	}
	// nothing-picked disables Apply (Keep is the only honest answer).
	// the moves panel is PARTIAL at birth — the first sweep ticks it
	// whole, the second unticks it entirely.
	state.ToggleSection(ScanSectionMoves)
	for _, key := range []string{ScanSectionDeletions,
		ScanSectionConsolidations, ScanSectionNewFolders, ScanSectionMoves} {
		state.ToggleSection(key)
	}
	if picked, total := state.PickedTotals(); picked != 0 || total != 5 {
		t.Fatalf("picked/total = %d/%d, want nothing picked of 5", picked, total)
	}
	if state.ApplyEnabled() {
		t.Fatal("nothing picked must disable Apply")
	}
	if !state.KeepEnabled() {
		t.Fatal("Keep never depends on the picks")
	}
	// an empty plan: nothing to pick, Apply stays armed-eligible.
	empty := NewPlanPickState(PlanDisplayModelOf(&vaultscan.Plan{}))
	empty.Arm()
	if !empty.ApplyEnabled() {
		t.Fatal("an empty plan's Apply is armed-eligible (nothing to pick)")
	}
	if empty.ApplyLabel() != "🗂️ Apply plan" {
		t.Fatalf("empty label = %q", empty.ApplyLabel())
	}
}

func TestPlanPickStateSelectionAllFourKeys(t *testing.T) {
	state := NewPlanPickState(PlanDisplayModelOf(samplePlan()))
	state.SetChecked(ScanSectionMoves, "C.md", true) // ticked back
	sel := state.Selection()
	if len(sel) != 4 {
		t.Fatalf("selection keys = %v, want all four always riding", sel)
	}
	if len(sel[ScanSectionDeletions]) != 1 || sel[ScanSectionDeletions][0] != "https://example.com/a" {
		t.Fatalf("deletions sel = %v", sel[ScanSectionDeletions])
	}
	if len(sel[ScanSectionConsolidations]) != 1 {
		t.Fatalf("consolidations sel = %v", sel[ScanSectionConsolidations])
	}
	if len(sel[ScanSectionNewFolders]) != 1 {
		t.Fatalf("folders sel = %v", sel[ScanSectionNewFolders])
	}
	if len(sel[ScanSectionMoves]) != 2 {
		t.Fatalf("moves sel = %v", sel[ScanSectionMoves])
	}
	// an unticked panel answers an EMPTY list (never a missing key).
	state.ToggleSection(ScanSectionMoves)
	if len(state.Selection()[ScanSectionMoves]) != 0 {
		t.Fatalf("unticked moves sel = %v", state.Selection()[ScanSectionMoves])
	}
}

// ---------------------------------------------------------------------------
// THE SCORECARD.
// ---------------------------------------------------------------------------

func TestResultDisplayModelKOfNLawAndTones(t *testing.T) {
	res := &ScanResult{Verdict: "confirmed",
		Story: &ScanStory{Deletions: 5, Applied: 3, ConsolidationNotes: 2,
			NotesConsolidated: 2, Moves: 4, Moved: 4, NewFolders: 1, FoldersCreated: 1}}
	model := ResultDisplayModelOf(res)
	if model.Heading != "The plan is applied" {
		t.Fatalf("heading = %q", model.Heading)
	}
	if model.LandedTotal != 10 || model.NothingLanded {
		t.Fatalf("landed = %d nothing=%v", model.LandedTotal, model.NothingLanded)
	}
	byName := map[string]ScoreRow{}
	for _, r := range model.Rows {
		byName[r.Name] = r
	}
	// the short row speaks the honest 'K of N' in the warning tone.
	if r := byName["🗑️ Removed (banished)"]; r.Value != "3 of 5" || r.Tone != ToneWarning {
		t.Fatalf("banished row = %+v", r)
	}
	if !strings.Contains(byName["🗑️ Removed (banished)"].Tip, "the log holds each why") {
		t.Fatalf("shortfall tip = %q", byName["🗑️ Removed (banished)"].Tip)
	}
	// a full landing speaks the plain count in the success tone.
	if r := byName["🚚 Notes re-filed"]; r.Value != "4" || r.Tone != ToneSuccess {
		t.Fatalf("moved row = %+v", r)
	}
	if !strings.Contains(byName["🚚 Notes re-filed"].Tip, "moved byte-identically") {
		t.Fatalf("moved tip = %q", byName["🚚 Notes re-filed"].Tip)
	}
	// the footnote teaches recoverability.
	if !strings.Contains(model.Footnote, "a banished site never fetches again") {
		t.Fatalf("footnote = %q", model.Footnote)
	}
	// a vault section groups the plan's rows under the story vault.
	if len(model.VaultSections) != 1 || model.VaultSections[0].Label != "📁 Websites vault" {
		t.Fatalf("sections = %+v", model.VaultSections)
	}
	if model.VaultSections[0].Changed != 10 {
		t.Fatalf("changed = %d", model.VaultSections[0].Changed)
	}
}

func TestResultDisplayModelNamedFailures(t *testing.T) {
	res := &ScanResult{Verdict: "confirmed",
		Story: &ScanStory{Moves: 5, Moved: 2, FailedMoves: 3,
			Failed: []vaultscan.FailedMove{
				{Note: "One.md", To: "Tools", Why: "busy"},
				{Note: "Two.md", To: "Tools", Why: "busy"},
				{Note: "Three.md", To: "Tools", Why: "busy"},
				{Note: "Four.md", To: "Tools", Why: "busy"},
			}}}
	model := ResultDisplayModelOf(res)
	var moveRow ScoreRow
	for _, r := range model.Rows {
		if r.Name == "🚚 Notes re-filed" {
			moveRow = r
		}
	}
	if moveRow.Tone != ToneWarning {
		t.Fatalf("a failed move must warn: %+v", moveRow)
	}
	// the failed files are NAMED (cap 3 + the …).
	if !strings.Contains(moveRow.Tip, "3 move(s) failed — One.md, Two.md, Three.md…") {
		t.Fatalf("named tip = %q", moveRow.Tip)
	}
	if !strings.Contains(moveRow.Tip, "the files stayed put; the log says why") {
		t.Fatalf("named tip = %q", moveRow.Tip)
	}
}

func TestResultDisplayModelHeadingsAndNothingLine(t *testing.T) {
	// nothing landed at all — the teaching line + the honest heading.
	res := &ScanResult{Verdict: "confirmed", Story: &ScanStory{Deletions: 1, Applied: 0}}
	model := ResultDisplayModelOf(res)
	if !model.NothingLanded {
		t.Fatal("nothing landed")
	}
	if model.Heading != "The plan is applied — nothing needed the machinery" {
		t.Fatalf("heading = %q", model.Heading)
	}
	if len(model.Rows) != 1 {
		t.Fatalf("rows = %d (only the shortfall row speaks)", len(model.Rows))
	}
	if !strings.Contains(model.NothingLine, "Every mark had already healed by hand") {
		t.Fatalf("nothing line = %q", model.NothingLine)
	}
	// the healed-only story — the hygiene pass did the whole work.
	res2 := &ScanResult{Verdict: "clean", HygieneTotal: 4,
		Hygiene: map[string]vaulthygiene.HygieneCounts{
			"websites": {RemovedIndex: 2, RetiredStubs: 1, RetiredReports: 1, RetiredLinked: 1},
		},
		Story: &ScanStory{Verdict: "clean"}}
	model2 := ResultDisplayModelOf(res2)
	if model2.Heading != "The vault is healed — the plan had nothing to do" {
		t.Fatalf("healed heading = %q", model2.Heading)
	}
	// the 🧹 row speaks the aggregate with the linked-ghost detail.
	var hygiene ScoreRow
	for _, r := range model2.Rows {
		if r.Name == "🧹 Strays healed" {
			hygiene = r
		}
	}
	if hygiene.Value != "4" || hygiene.Tone != ToneSuccess {
		t.Fatalf("hygiene row = %+v", hygiene)
	}
	if !strings.Contains(hygiene.Tip, "1 of them linked ghosts (empty notes other notes pointed at).") {
		t.Fatalf("hygiene tip = %q", hygiene.Tip)
	}
	// a nil story degrades to the nothing-landed glance (never a crash).
	m3 := ResultDisplayModelOf(nil)
	if !m3.NothingLanded {
		t.Fatal("a nil result must degrade to nothing-landed")
	}
}

func TestResultDisplayModelPerVaultSections(t *testing.T) {
	// both vaults ran: the story vault's section + the repos
	// sub-story's own section.
	res := &ScanResult{Verdict: "confirmed",
		Vault: "/wv", VaultKind: "websites", VaultLabel: "Websites vault",
		GitHubVault: "/rv",
		Story:       &ScanStory{Deletions: 2, Applied: 2},
		Repos:       &ScanStory{Deletions: 1, Applied: 1, Moves: 1, Moved: 1}}
	model := ResultDisplayModelOf(res)
	if len(model.VaultSections) != 2 {
		t.Fatalf("sections = %d", len(model.VaultSections))
	}
	if model.VaultSections[0].Label != "📁 Websites vault" || model.VaultSections[0].Changed != 2 {
		t.Fatalf("websites section = %+v", model.VaultSections[0])
	}
	if !strings.Contains(model.VaultSections[0].Tip, "\nPath: /wv") {
		t.Fatalf("websites tip = %q", model.VaultSections[0].Tip)
	}
	if model.VaultSections[1].Label != "📁 GitHub Repos vault" || model.VaultSections[1].Changed != 2 {
		t.Fatalf("repos section = %+v", model.VaultSections[1])
	}
	if !strings.Contains(model.VaultSections[1].Tip, "The repos plan applied here: 2 change(s) landed in this vault.") {
		t.Fatalf("repos tip = %q", model.VaultSections[1].Tip)
	}
	// the repos section's own honest tips name the repo flavor.
	for _, r := range model.VaultSections[1].Rows {
		if r.Name == "🧲 Duplicates merged" {
			t.Fatal("a hidden row must never ride a section")
		}
	}

	// a repos-only scan: the promoted story IS the story — and the
	// reference's own model renders the promoted section AND the
	// repos sub-story's section BOTH (the out dict carries
	// out['repos'] unconditionally; the finished routing hands the
	// raw dict to the scorecard — the double section is the
	// reference's observable behavior, verified by running
	// result_display_model over the repos-only shape at 98f4fc7).
	res2 := &ScanResult{Verdict: "confirmed", Vault: "/rv", VaultKind: "repos",
		VaultLabel: "GitHub Repos vault", GitHubVault: "/rv",
		Story: &ScanStory{Deletions: 1, Applied: 1},
		Repos: &ScanStory{Deletions: 1, Applied: 1}}
	model2 := ResultDisplayModelOf(res2)
	if len(model2.VaultSections) != 2 {
		t.Fatalf("repos-only sections = %+v (the quirk preserved: the promoted story + the sub-story both render)", model2.VaultSections)
	}
	for _, sec := range model2.VaultSections {
		if sec.Label != "📁 GitHub Repos vault" {
			t.Fatalf("label = %q", sec.Label)
		}
	}

	// the github hygiene heal alone: its own "projects vault" section.
	res3 := &ScanResult{Verdict: "clean",
		GitHubVault: "/rv",
		Hygiene: map[string]vaulthygiene.HygieneCounts{
			"github": {RetiredStubs: 2}},
		HygieneTotal: 2,
		Story:        &ScanStory{Verdict: "clean"}}
	model3 := ResultDisplayModelOf(res3)
	if len(model3.VaultSections) != 1 || model3.VaultSections[0].Label != "📁 GitHub projects vault" {
		t.Fatalf("github-only sections = %+v", model3.VaultSections)
	}
	if !strings.Contains(model3.VaultSections[0].Tip, "2 stray file(s) healed in this vault") {
		t.Fatalf("github tip = %q", model3.VaultSections[0].Tip)
	}
}

// ---------------------------------------------------------------------------
// The story extension — the scan job's own laws (worker_jobs.py's
// story dict, T-76's slice over the T-71 engine).
// ---------------------------------------------------------------------------

func TestScanStoryRidesTheConfirmedJob(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeMarkedNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(string, string) {},
		Progress: func(string, int, int) {},
		Confirm:  func(any) ScanAnswer { return ScanAnswer{Verdict: ScanConfirmed} },
	})
	if res.Verdict != "confirmed" {
		t.Fatalf("verdict = %q", res.Verdict)
	}
	// the repos story IS the story (the repos-only promotion).
	if res.Story == nil || res.Repos == nil {
		t.Fatal("the story seats must ride a confirmed job")
	}
	if res.VaultKind != "repos" || res.VaultLabel != "GitHub Repos vault" || res.Vault != vault {
		t.Fatalf("promotion = %q %q %q", res.VaultKind, res.VaultLabel, res.Vault)
	}
	if res.Story.Deletions != 1 || res.Story.Applied != 1 {
		t.Fatalf("story = %+v (planned 1, landed 1)", res.Story)
	}
	if res.Repos.Applied != 1 {
		t.Fatalf("repos sub-story = %+v", res.Repos)
	}
	if len(res.VaultsRan) != 1 || res.VaultsRan[0] != "repos" {
		t.Fatalf("vaults ran = %v", res.VaultsRan)
	}
	// the scorecard speaks it end-to-end (the promoted story + the
	// sub-story both render — the reference's own repos-only quirk,
	// probed at 98f4fc7).
	model := ResultDisplayModelOf(res)
	if model.Heading != "The plan is applied" || model.LandedTotal != 1 {
		t.Fatalf("scorecard = %+v", model)
	}
	if len(model.VaultSections) != 2 || model.VaultSections[0].Label != "📁 GitHub Repos vault" {
		t.Fatalf("sections = %+v", model.VaultSections)
	}
}

func TestScanStoryPickedCountsReplaceThePlanned(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeMarkedNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	writeMarkedNote(t, vault, "Agents/second.md", "https://github.com/octo/second")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	// the owner unticks BOTH deletion rows: the picking door answers a
	// selection that carries no deletions.
	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(string, string) {},
		Progress: func(string, int, int) {},
		Confirm: func(plan any) ScanAnswer {
			return ScanAnswer{Verdict: ScanConfirmed,
				Selection: map[string][]string{
					ScanSectionDeletions:      {},
					ScanSectionConsolidations: {},
					ScanSectionNewFolders:     {},
					ScanSectionMoves:          {},
				}}
		},
	})
	if res.Verdict != "confirmed" {
		t.Fatalf("verdict = %q", res.Verdict)
	}
	// the story's planned counts speak the PICKED units — the unticked
	// rows are never a failure (worker_jobs.py:1130-1136).
	if res.Story.Deletions != 0 {
		t.Fatalf("planned deletions = %d, want the picked 0", res.Story.Deletions)
	}
	if res.Story.Applied != 0 {
		t.Fatalf("applied = %d, want nothing banished", res.Story.Applied)
	}
	// the notes rest untouched.
	if _, err := os.Stat(filepath.Join(vault, "Agents", "doomed.md")); err != nil {
		t.Fatalf("an unticked deletion still touched the note: %v", err)
	}
}

func TestScanStoryGuiDoorAskLine(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeMarkedNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	cfg.Set("scan_confirm_timeout_s", 90.0)

	var logs []string
	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(msg, level string) { logs = append(logs, msg) },
		Progress: func(string, int, int) {},
		Confirm:  func(any) ScanAnswer { return ScanAnswer{Verdict: ScanDeclined} },
	})
	if res.Verdict != "declined" {
		t.Fatalf("verdict = %q", res.Verdict)
	}
	joined := strings.Join(logs, "\n")
	// the GUI door's own opening line, byte-for-byte, with the
	// configured window spoken.
	if !strings.Contains(joined, "🗂️ Vault scan review (GitHub Repos vault) — the plan is on your screen now (tick what you want, untick what you deny; the vault stays untouched until you answer; up to 90s)") {
		t.Fatalf("the ask line never spoke: %v", logs)
	}
	// the declined story carries the planned counts it never landed.
	if res.Story.Deletions != 1 || res.Story.Applied != 0 {
		t.Fatalf("declined story = %+v", res.Story)
	}
}

func TestScanStoryTimeoutVerdictAndCleanLibrary(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeMarkedNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	var logs []string
	res := RunVaultScan(cfg, []string{"repos"}, ScanEvents{
		Log:      func(msg, level string) { logs = append(logs, msg) },
		Progress: func(string, int, int) {},
		Confirm:  func(any) ScanAnswer { return ScanAnswer{Verdict: ScanTimeout} },
	})
	if res.Verdict != "timeout" {
		t.Fatalf("verdict = %q, want the safe defer word", res.Verdict)
	}
	if !strings.Contains(strings.Join(logs, "\n"), "(no answer in time)") {
		t.Fatalf("the timeout why never spoke: %v", logs)
	}
	// a clean library carries a clean story with zero rows.
	vault2 := t.TempDir()
	writeNote(t, vault2, "Agents/clean.md", "https://github.com/octo/clean")
	cfg2 := storage.NewOrderedMap()
	cfg2.Set("vault_path", vault2)
	res2 := RunVaultScan(cfg2, []string{"repos"}, ScanEvents{
		Log:      func(string, string) {},
		Progress: func(string, int, int) {},
		Confirm:  func(any) ScanAnswer { return ScanAnswer{Verdict: ScanTimeout} },
	})
	if res2.Verdict != "clean" || res2.Story.Verdict != "clean" {
		t.Fatalf("clean story = %+v", res2.Story)
	}
	model := ResultDisplayModelOf(res2)
	if !model.NothingLanded {
		t.Fatalf("a clean scan's scorecard = %+v", model)
	}
}
