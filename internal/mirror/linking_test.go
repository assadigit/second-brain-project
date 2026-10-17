// linking_test.go — the ported test_phase6 linking selections: the
// candidate pairs, the confirm prompt/parsing, the link store's
// lifecycle (pending → approved/rejected, never re-suggested), the
// Suggestions note + decisions grammar, the Related (auto) blocks
// (mirror copies ONLY), and collect_note_fields.
package mirror

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// -- TestCandidates -------------------------------------------------------------

func TestCandidatesFindCandidatePairs(t *testing.T) {
	entries := []VecEntry{
		{Key: "a", Vec: []float64{1.0, 0.0}},
		{Key: "b", Vec: []float64{0.99, 0.1}}, // near-duplicate of a
		{Key: "c", Vec: []float64{0.0, 1.0}},  // orthogonal
		{Key: "d", Vec: []float64{0.0, 0.98}}, // near-duplicate of c
	}
	pairs := FindCandidatePairs(entries, 2, 0.5)
	if len(pairs) == 0 {
		t.Fatal("no pairs")
	}
	// each unordered pair appears once, sorted by score desc
	seen := map[[2]string]bool{}
	for _, p := range pairs {
		if p.A > p.B {
			t.Fatalf("pair not canonicalized: %+v", p)
		}
		key := [2]string{p.A, p.B}
		if seen[key] {
			t.Fatalf("pair %v listed twice", key)
		}
		seen[key] = true
	}
	if !seen[[2]string{"a", "b"}] {
		t.Fatal("a-b missing")
	}
	if !seen[[2]string{"c", "d"}] {
		t.Fatal("c-d missing")
	}
	if seen[[2]string{"a", "c"}] { // orthogonal < 0.5
		t.Fatal("orthogonal pair survived the floor")
	}
	for i := 1; i < len(pairs); i++ {
		if pairs[i-1].Score < pairs[i].Score {
			t.Fatalf("pairs not sorted desc: %+v", pairs)
		}
	}
}

func TestCandidatesMinScorePrunesEverything(t *testing.T) {
	// near-identical vectors survive a 0.99 floor…
	pairs := FindCandidatePairs([]VecEntry{
		{Key: "a", Vec: []float64{1.0}},
		{Key: "b", Vec: []float64{1.0}},
	}, 8, 0.99)
	if len(pairs) != 1 {
		t.Fatalf("pairs = %d, want 1", len(pairs))
	}
	// …but a 0.6 similarity is pruned at 0.999
	pairs = FindCandidatePairs([]VecEntry{
		{Key: "a", Vec: []float64{1.0, 0.1}},
		{Key: "b", Vec: []float64{0.6, 0.8}},
	}, 8, 0.999)
	if len(pairs) != 0 {
		t.Fatalf("pairs = %d, want 0", len(pairs))
	}
}

// -- TestConfirm ------------------------------------------------------------------

var (
	confirmFieldsA = RecallFields{
		Name: "Palette Gen", OneLine: "Palettes from images",
		Recall: "Use when you need to build a scheme.", Tags: "design"}
	confirmFieldsB = RecallFields{
		Name: "Contrast Check", OneLine: "A11y contrast",
		Recall: "Use when you need to check contrast.", Tags: "design"}
)

func TestConfirmPromptFilled(t *testing.T) {
	prompt, err := BuildConfirmPrompt(confirmFieldsA, confirmFieldsB)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "Palette Gen") ||
		!strings.Contains(prompt, "Contrast Check") {
		t.Fatalf("names missing:\n%s", prompt)
	}
	if strings.Contains(prompt, "{{") {
		t.Fatal("unfilled slot")
	}
}

func TestConfirmParseVariants(t *testing.T) {
	related, reason := ParseConfirmReply(
		`{"related": true, "reason": "both color work"}`)
	if related == nil || !*related || reason != "both color work" {
		t.Fatalf("got (%v, %q)", related, reason)
	}
	related, reason = ParseConfirmReply(
		`Sure! {"related": false, "reason": "no overlap"}`)
	if related == nil || *related || reason != "no overlap" {
		t.Fatalf("got (%v, %q)", related, reason)
	}
	related, reason = ParseConfirmReply(`{"related": "yes", "reason": "overlap"}`)
	if related == nil || !*related || reason != "overlap" {
		t.Fatalf("got (%v, %q)", related, reason)
	}
	related, reason = ParseConfirmReply("garbage")
	if related != nil || reason != "" {
		t.Fatalf("garbage got (%v, %q)", related, reason)
	}
}

func TestConfirmPairThroughLLM(t *testing.T) {
	frozenClock(t, frozenTime())
	llm := func(messages []llmMessage) (string, error) {
		return `{"related": true, "reason": "both color tools"}`, nil
	}
	related, reason := ConfirmPair(confirmFieldsA, confirmFieldsB, llm)
	if related == nil || !*related {
		t.Fatalf("related = %v", related)
	}
	if reason != "both color tools" {
		t.Fatalf("reason = %q", reason)
	}
}

// -- TestLinkStore ------------------------------------------------------------------

func newTestLinkStore(t *testing.T) *LinkStore {
	t.Helper()
	s, err := NewLinkStore(filepath.Join(t.TempDir(), "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

func TestLinkStoreLifecycleAndNeverResuggest(t *testing.T) {
	frozenClock(t, frozenTime())
	store := newTestLinkStore(t)
	if status, err := store.Suggest("a", "b", 0.9, "why"); err != nil ||
		status != StatusPending {
		t.Fatalf("suggest = (%q, %v)", status, err)
	}
	// same pair again → existing status, no duplicate row
	if status, _ := store.Suggest("a", "b", 0.9, "why"); status != StatusPending {
		t.Fatalf("re-suggest = %q", status)
	}
	counts, _ := store.Counts()
	if counts[StatusPending] != 1 {
		t.Fatalf("pending = %d", counts[StatusPending])
	}
	// order-independent identity
	if status, _ := store.Suggest("b", "a", 0.9, ""); status != StatusPending {
		t.Fatalf("mirrored suggest = %q", status)
	}
	changed, err := store.Decide("a", "b", StatusApproved)
	if err != nil || !changed {
		t.Fatalf("decide = (%v, %v)", changed, err)
	}
	if changed, _ := store.Decide("a", "b", StatusApproved); changed {
		t.Fatal("already-decided pair decided twice")
	}
	// a decided pair is NEVER suggested again
	if status, _ := store.Suggest("a", "b", 0.95, "re-ask"); status != StatusApproved {
		t.Fatalf("re-ask = %q", status)
	}
	if status, _ := store.Suggest("b", "a", 0.95, "re-ask"); status != StatusApproved {
		t.Fatalf("mirrored re-ask = %q", status)
	}
	// rejected likewise
	if _, err := store.Suggest("c", "d", 0.8, ""); err != nil {
		t.Fatal(err)
	}
	if _, err := store.Decide("c", "d", StatusRejected); err != nil {
		t.Fatal(err)
	}
	if status, _ := store.Suggest("c", "d", 0.9, ""); status != StatusRejected {
		t.Fatalf("rejected re-ask = %q", status)
	}
	counts, _ = store.Counts()
	if counts[StatusPending] != 0 || counts[StatusApproved] != 1 ||
		counts[StatusRejected] != 1 {
		t.Fatalf("counts = %v", counts)
	}
	// invalid status is the reference's ValueError
	if _, err := store.Decide("a", "b", "maybe"); err == nil {
		t.Fatal("invalid status accepted")
	}
}

func TestLinkStoreApprovedLinksCapped(t *testing.T) {
	frozenClock(t, frozenTime())
	store := newTestLinkStore(t)
	for i := 0; i < 10; i++ {
		key := "https://x/" + strings.Repeat("0", 2-len(itoa(i))) + itoa(i)
		if _, err := store.Suggest("https://center", key,
			1.0-float64(i)*0.01, ""); err != nil {
			t.Fatal(err)
		}
		if _, err := store.Decide("https://center", key, StatusApproved); err != nil {
			t.Fatal(err)
		}
	}
	links, err := store.ApprovedLinks()
	if err != nil {
		t.Fatal(err)
	}
	if len(links["https://center"]) != LinksCap {
		t.Fatalf("cap = %d links, want %d", len(links["https://center"]), LinksCap)
	}
	if LinksCap != 7 {
		t.Fatalf("LINKS_CAP = %d, want 7", LinksCap)
	}
	// strongest first
	if links["https://center"][0].Other != "https://x/00" {
		t.Fatalf("first = %q", links["https://center"][0].Other)
	}
}

// itoa is the small local int formatter.
func itoa(i int) string {
	if i == 0 {
		return "0"
	}
	var digits []byte
	for i > 0 {
		digits = append([]byte{byte('0' + i%10)}, digits...)
		i /= 10
	}
	return string(digits)
}

// -- TestSuggestionsNote --------------------------------------------------------------

func suggestionsFixture(t *testing.T) (*LinkStore, map[string]*FieldsEntry) {
	t.Helper()
	frozenClock(t, frozenTime())
	store := newTestLinkStore(t)
	mustSuggest(t, store, "https://a", "https://b", 0.9, "both color tools")
	mustSuggest(t, store, "https://c", "https://d", 0.8, "similar purpose")
	if _, err := store.Decide("https://c", "https://d", StatusApproved); err != nil {
		t.Fatal(err)
	}
	mustSuggest(t, store, "https://e", "https://f", 0.7, "weak")
	if _, err := store.Decide("https://e", "https://f", StatusRejected); err != nil {
		t.Fatal(err)
	}
	fields := map[string]*FieldsEntry{
		"https://a": {Fields: RecallFields{Name: "Alpha"}},
		"https://b": {Fields: RecallFields{Name: "Beta"}},
		"https://c": {Fields: RecallFields{Name: "Gamma"}},
		"https://d": {Fields: RecallFields{Name: "Delta"}},
		"https://e": {Fields: RecallFields{Name: "Epsilon"}},
		"https://f": {Fields: RecallFields{Name: "Zeta"}},
	}
	return store, fields
}

func mustSuggest(t *testing.T, s *LinkStore, a, b string, score float64, reason string) {
	t.Helper()
	if _, err := s.Suggest(a, b, score, reason); err != nil {
		t.Fatal(err)
	}
}

func TestSuggestionsNoteStructureAndRoundtrip(t *testing.T) {
	store, fields := suggestionsFixture(t)
	note, err := BuildSuggestionsNote(store, fields)
	if err != nil {
		t.Fatal(err)
	}
	for _, needle := range []string{
		"Link suggestions", "- [ ] [[Alpha]] ↔ [[Beta]]",
		"`https://a` / `https://b`", "## Approved", "## Rejected",
	} {
		if !strings.Contains(note, needle) {
			t.Fatalf("needle %q missing:\n%s", needle, note)
		}
	}
	// roundtrip: nothing decided yet
	path := filepath.Join(t.TempDir(), "Suggestions.md")
	if err := storage.AtomicWriteText(path, note); err != nil {
		t.Fatal(err)
	}
	if decisions := CollectDecisions(path); len(decisions) != 0 {
		t.Fatalf("decisions = %+v", decisions)
	}
	// tick the pending one → approved
	if err := storage.AtomicWriteText(path,
		strings.Replace(note, "- [ ] [[Alpha]]", "- [x] [[Alpha]]", 1)); err != nil {
		t.Fatal(err)
	}
	decisions := CollectDecisions(path)
	if len(decisions) != 1 || decisions[0].KeyA != "https://a" ||
		decisions[0].KeyB != "https://b" || decisions[0].Status != StatusApproved {
		t.Fatalf("decisions = %+v", decisions)
	}
	// strike it instead → rejected
	if err := storage.AtomicWriteText(path, strings.Replace(note,
		"- [ ] [[Alpha]] ↔ [[Beta]]",
		"- [ ] ~~[[Alpha]] ↔ [[Beta]]~~", 1)); err != nil {
		t.Fatal(err)
	}
	decisions = CollectDecisions(path)
	if len(decisions) != 1 || decisions[0].Status != StatusRejected {
		t.Fatalf("decisions = %+v", decisions)
	}
	// free text without URLs is never misparsed as a decision
	ticked := strings.Replace(note, "- [ ] [[Alpha]]", "- [x] [[Alpha]]", 1)
	if err := storage.AtomicWriteText(path, ticked+"\n- [x] random line\n"); err != nil {
		t.Fatal(err)
	}
	if got := len(CollectDecisions(path)); got != 1 {
		t.Fatalf("decisions = %d", got)
	}
}

func TestSuggestionsNoteApplyDecisions(t *testing.T) {
	store, _ := suggestionsFixture(t)
	approved, rejected, ignored, err := ApplyDecisions(store, []Decision{
		{KeyA: "https://a", KeyB: "https://b", Status: StatusApproved},
		{KeyA: "https://zz", KeyB: "https://yy", Status: StatusApproved}, // unknown pair
	})
	if err != nil {
		t.Fatal(err)
	}
	if approved != 1 || rejected != 0 || ignored != 1 {
		t.Fatalf("counts = (%d, %d, %d)", approved, rejected, ignored)
	}
	counts, _ := store.Counts()
	if counts[StatusApproved] != 2 {
		t.Fatalf("approved = %d", counts[StatusApproved])
	}
}

// -- TestRelatedBlocks -------------------------------------------------------------------

func TestRelatedBlocksRenderInsertExtract(t *testing.T) {
	block := RenderRelatedBlock(
		[]ApprovedLink{{Other: "https://x", Reason: "same problem", Score: 0.9}},
		func(url string) string { return "Other Note" })
	if !strings.Contains(block, RelatedStart) {
		t.Fatal("start marker missing")
	}
	if !strings.Contains(block, "[[Other Note]] — same problem") {
		t.Fatalf("block = %q", block)
	}
	note := InsertRelatedBlock("# Title\n\nBody\n", block)
	if !HasRelatedBlock(note) {
		t.Fatal("block missing")
	}
	if ExtractRelatedBlock(note) != block {
		t.Fatal("extract diverges")
	}
	// replace in place
	block2 := RenderRelatedBlock(nil, func(u string) string { return "" })
	note2 := InsertRelatedBlock(note, block2)
	if ExtractRelatedBlock(note2) != block2 {
		t.Fatal("replace diverges")
	}
	if strings.Count(note, RelatedStart) != 1 {
		t.Fatal("more than one block")
	}
}

func TestRelatedBlocksApplyOnlyTouchesMirrorCopies(t *testing.T) {
	tmp := t.TempDir()
	library := makeVault(t, filepath.Join(tmp, "Library"), [][2]string{
		{"GitHub Projects/Dev Tools/mir.md", mirrorNote6("https://github.com/a/a", "Mirror note", "")},
		{"GitHub Projects/Dev Tools/owner.md", "# My own note, no mirror_of\n"},
		{"Websites/Design/web.md", mirrorNote6("https://palette.example.com", "Web note", "")},
	})
	wrote := map[string]string{}
	writer := func(path, content string) {
		if err := storage.AtomicWriteText(path, content); err != nil {
			t.Fatal(err)
		}
		wrote[filepath.Base(path)] = content
	}
	approved := map[string][]ApprovedLink{
		"https://github.com/a/a": {
			{Other: "https://palette.example.com",
				Reason: "both design-adjacent", Score: 0.9}},
	}
	fields := map[string]*FieldsEntry{
		"https://palette.example.com": {Fields: RecallFields{Name: "Palette Gen"}},
	}
	updated := ApplyRelatedBlocks(library, approved, fields, writer)
	var sortedUpdated []string
	sortedUpdated = append(sortedUpdated, updated...)
	sort.Strings(sortedUpdated)
	want := []string{"GitHub Projects/Dev Tools/mir.md"}
	if strings.Join(sortedUpdated, "|") != strings.Join(want, "|") {
		t.Fatalf("updated = %v", sortedUpdated)
	}
	if !strings.Contains(wrote["mir.md"], "[[Palette Gen]] — both design-adjacent") {
		t.Fatalf("mir.md = %q", wrote["mir.md"])
	}
	if _, touched := wrote["owner.md"]; touched { // unmarked: untouched
		t.Fatal("the owner's unmarked file was touched")
	}
	// removing the approval REMOVES the block on the next pass
	updated = ApplyRelatedBlocks(library, map[string][]ApprovedLink{}, fields, writer)
	found := false
	for _, u := range updated {
		if u == "GitHub Projects/Dev Tools/mir.md" {
			found = true
		}
	}
	if !found {
		t.Fatalf("removal pass did not update the mirror copy: %v", updated)
	}
	data, _ := os.ReadFile(filepath.Join(library, "GitHub Projects", "Dev Tools", "mir.md"))
	if HasRelatedBlock(string(data)) {
		t.Fatal("the block survived the removal pass")
	}
}

func TestRelatedBlocksPreserveRelatedBlock(t *testing.T) {
	block := RenderRelatedBlock(
		[]ApprovedLink{{Other: "https://x", Reason: "r", Score: 0.5}},
		func(u string) string { return "X" })
	old := InsertRelatedBlock("# Old\n\nBody\n", block)
	new := "# New body\n\nRebuilt.\n"
	out := PreserveRelatedBlock(old, new)
	if ExtractRelatedBlock(out) != block {
		t.Fatalf("block lost: %q", out)
	}
	if !strings.HasPrefix(out, "# New body") {
		t.Fatalf("out = %q", out)
	}
	if got := PreserveRelatedBlock("# No block\n", "# New\n"); got != "# New\n" {
		t.Fatalf("no-block carry = %q", got)
	}
}

func TestRelatedBlocksMirrorResyncKeepsTheBlock(t *testing.T) {
	// THE Phase 5 ↔ 6 integration acceptance: an approved Related block
	// survives a full mirror re-apply.
	tmp := t.TempDir()
	machine := makeVault(t, filepath.Join(tmp, "machine"), [][2]string{
		{"Dev Tools/demo.md",
			githubNote6("demo-tool", "https://github.com/a/demo", "t", "x.")},
	})
	manual := filepath.Join(tmp, "manual")
	if err := os.MkdirAll(manual, 0o755); err != nil { // the mirror fills, never creates
		t.Fatal(err)
	}
	if _, err := RunMirror(manual, machine, "", true, nil); err != nil {
		t.Fatal(err)
	}
	mirPath := filepath.Join(manual, "Library", GitHubTree, "Dev Tools", "demo.md")
	first, _ := os.ReadFile(mirPath)
	if HasRelatedBlock(string(first)) {
		t.Fatal("unexpected block on the first mirror")
	}
	block := RenderRelatedBlock(
		[]ApprovedLink{{Other: "https://palette.example.com",
			Reason: "design adjacent", Score: 0.9}},
		func(u string) string { return "Palette Gen" })
	if err := storage.AtomicWriteText(mirPath,
		InsertRelatedBlock(string(first), block)); err != nil {
		t.Fatal(err)
	}
	// machine note changes → the mirror rebuilds — the approved block
	// must carry over
	writeFixture(t, filepath.Join(machine, "Dev Tools", "demo.md"),
		githubNote6("demo-tool", "https://github.com/a/demo", "t", "x.")+"\nEdited.\n")
	if _, err := RunMirror(manual, machine, "", true, nil); err != nil {
		t.Fatal(err)
	}
	second, _ := os.ReadFile(mirPath)
	if !strings.Contains(string(second), "Edited.") {
		t.Fatal("the edit did not reach the mirror")
	}
	if ExtractRelatedBlock(string(second)) != block {
		t.Fatalf("the approved block was not carried over:\n%s", second)
	}
	// and an UNCHANGED note (block already there) is a plain keep
	plan, err := RunMirror(manual, machine, "", false, nil)
	if err != nil {
		t.Fatal(err)
	}
	tree := plan.Trees[GitHubTree]
	if tree.Keeps != 1 || tree.Changes() != 0 {
		t.Fatalf("keeps = %d, changes = %d", tree.Keeps, tree.Changes())
	}
}

// -- TestCollectFields -----------------------------------------------------------------------

func TestCollectFieldsBothVaultsAndDuplicates(t *testing.T) {
	tmp := t.TempDir()
	github := makeVault(t, filepath.Join(tmp, "gh"), [][2]string{
		{"Dev Tools/a.md", InsertRecallBlock(
			githubNote6("a", "https://github.com/a/a", "t", "x."),
			"Use when you need to a.")},
		{"Dev Tools/unmanaged.md", "# nothing\n"},
	})
	websites := makeVault(t, filepath.Join(tmp, "web"), [][2]string{
		{"Design/Color Tools/p.md", websiteNote6(
			"Palette Gen", "https://palette.example.com",
			"Use when you need to build a color scheme fast.")},
		{"Design/Color Tools/p-dup.md", websiteNote6(
			"Palette Gen", "https://palette.example.com",
			"Use when you need to build a color scheme fast.")},
	})
	byKey := CollectNoteFields([]VaultPath{
		{Vault: "github", Path: github},
		{Vault: "websites", Path: websites},
	})
	if byKey.Duplicates != 1 {
		t.Fatalf("duplicates = %d, want 1", byKey.Duplicates)
	}
	if len(byKey.BySource) != 2 {
		t.Fatalf("sources = %v", byKey.Order)
	}
	if byKey.BySource["https://palette.example.com"].Fields.Name != "Palette Gen" {
		t.Fatalf("web name = %q",
			byKey.BySource["https://palette.example.com"].Fields.Name)
	}
	if byKey.BySource["https://github.com/a/a"].Fields.Recall != "Use when you need to a." {
		t.Fatalf("github recall = %q",
			byKey.BySource["https://github.com/a/a"].Fields.Recall)
	}
}
