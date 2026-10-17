// recall_test.go — the ported test_phase6 selections for the recall
// block + pipeline: insert/get roundtrip, idempotence, in-place
// replacement, CRLF preservation, fingerprint invisibility, the
// sanitizers, field extraction, the neutral prompt, the plan/run doors.
package mirror

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// -- TestRecallBlock ---------------------------------------------------------

func TestRecallBlockInsertGetRoundtrip(t *testing.T) {
	note := githubNote6("demo-tool", "https://github.com/acme/demo-tool",
		"python, cli", "A fast demo tool for testing.")
	out := InsertRecallBlock(note, "Use when you need to test.")
	if !HasRecallBlock(out) {
		t.Fatal("block missing")
	}
	if got := GetRecallSentence(out); got != "Use when you need to test." {
		t.Fatalf("sentence = %q", got)
	}
	if !strings.Contains(out, notestate.RecallStart) ||
		!strings.Contains(out, notestate.RecallEnd) {
		t.Fatal("markers missing")
	}
}

func TestRecallBlockIdempotentSameSentence(t *testing.T) {
	note := InsertRecallBlock(
		githubNote6("demo-tool", "https://github.com/acme/demo-tool",
			"python, cli", "A fast demo tool for testing."),
		"Use when you need to test.")
	again := InsertRecallBlock(note, "Use when you need to test.")
	if note != again {
		t.Fatal("re-insert is not byte-identical")
	}
}

func TestRecallBlockReplaceInPlace(t *testing.T) {
	note := InsertRecallBlock(
		githubNote6("demo-tool", "https://github.com/acme/demo-tool",
			"python, cli", "A fast demo tool for testing."),
		"Use when you need to test.")
	note2 := InsertRecallBlock(note, "Use when you need to demo.")
	if got := GetRecallSentence(note2); got != "Use when you need to demo." {
		t.Fatalf("sentence = %q", got)
	}
	// exactly ONE block, and everything outside it is untouched
	if strings.Count(note2, notestate.RecallStart) != 1 {
		t.Fatalf("start marker count = %d", strings.Count(note2, notestate.RecallStart))
	}
	if StripRecallBlock(note) != StripRecallBlock(note2) {
		t.Fatal("the strip view changed")
	}
}

func TestRecallBlockCrlfNotesStayCrlf(t *testing.T) {
	note := strings.ReplaceAll(
		githubNote6("demo-tool", "https://github.com/acme/demo-tool",
			"python, cli", "A fast demo tool for testing."), "\n", "\r\n")
	out := InsertRecallBlock(note, "Use when you need to test.")
	if !strings.Contains(out, "\r\n") {
		t.Fatal("CRLF lost")
	}
	if strings.Contains(out, "\r\r") {
		t.Fatal("double carriage returns introduced")
	}
	if !HasRecallBlock(out) {
		t.Fatal("block missing")
	}
}

func TestRecallBlockFingerprintIgnoresTheBlock(t *testing.T) {
	note := githubNote6("demo-tool", "https://github.com/acme/demo-tool",
		"python, cli", "A fast demo tool for testing.")
	before := notestate.ComputeFingerprint(note)
	hooked := InsertRecallBlock(note, "Use when you need to test.")
	if before != notestate.ComputeFingerprint(hooked) {
		t.Fatal("the recall hook reads as a human edit")
	}
	// and a HUMAN edit after the hook still reads as changed
	if before == notestate.ComputeFingerprint(hooked+"\nhand-written line\n") {
		t.Fatal("a human edit is invisible")
	}
}

func TestRecallBlockSanitizeAcceptsValid(t *testing.T) {
	ok := "Use when you need to generate fixture data for tests."
	if got := SanitizeRecallSentence(ok); got != ok {
		t.Fatalf("got %q", got)
	}
	if got := SanitizeRecallSentence("  " + ok + "  "); got != ok {
		t.Fatalf("padded got %q", got)
	}
}

func TestRecallBlockSanitizeRejectsAndFallsBack(t *testing.T) {
	bad := []string{
		"Try this amazing tool!!", "", "multiple\nlines\nhere",
		"", "Use when you need to\nsecond line",
	}
	for _, raw := range bad {
		if got := SanitizeRecallSentence(raw); got != RecallPlaceholder {
			t.Fatalf("raw %q sanitized to %q", raw, got)
		}
	}
}

func TestRecallBlockSanitizeStripsQuotes(t *testing.T) {
	got := SanitizeRecallSentence(`"Use when you need to quote things."`)
	if got != "Use when you need to quote things." {
		t.Fatalf("got %q", got)
	}
}

func TestRecallBlockSanitizeTrimsRambling(t *testing.T) {
	longOK := "Use when you need to " + strings.Repeat("very ", 60) + "long"
	got := SanitizeRecallSentence(longOK)
	if !strings.HasPrefix(got, "Use when you need to") {
		t.Fatalf("got %q", got)
	}
	// len() in code points (the Python len law)
	if runeLen(got) > recallMaxLen+1 {
		t.Fatalf("rambling survived: %d runes", runeLen(got))
	}
}

// runeLen is len(s) over code points.
func runeLen(s string) int {
	n := 0
	for range s {
		n++
	}
	return n
}

// -- TestRecallPipeline --------------------------------------------------------

func TestRecallPipelineExtractGithubFields(t *testing.T) {
	fields := ExtractGithubFields(
		githubNote6("demo-tool", "https://github.com/acme/demo-tool",
			"python, cli", "A fast demo tool for testing."))
	if fields.RepoName != "demo-tool" {
		t.Fatalf("REPO_NAME = %q", fields.RepoName)
	}
	if fields.TLDR != "A fast demo tool for testing." {
		t.Fatalf("TLDR = %q", fields.TLDR)
	}
	if !strings.Contains(fields.Summary, "automates the boring parts") {
		t.Fatalf("SUMMARY = %q", fields.Summary)
	}
	if !strings.Contains(fields.Features, "fixture generation") {
		t.Fatalf("FEATURES = %q", fields.Features)
	}
	if fields.Tags != "python, cli" {
		t.Fatalf("TAGS = %q", fields.Tags)
	}
}

func TestRecallPipelineExtractWebsiteFields(t *testing.T) {
	fields := ExtractWebsiteFields(websiteNote6(
		"Palette Gen", "https://palette.example.com",
		"Use when you need to build a color scheme fast."))
	if fields.Name != "Palette Gen" {
		t.Fatalf("name = %q", fields.Name)
	}
	if fields.BestUsedFor != "Use when you need to build a color scheme fast." {
		t.Fatalf("best_used_for = %q", fields.BestUsedFor)
	}
	if fields.Tags != "design, colors" {
		t.Fatalf("tags = %q", fields.Tags)
	}
}

func TestRecallPipelineRecallFieldsForBothVaults(t *testing.T) {
	github := RecallFieldsFor(
		InsertRecallBlock(
			githubNote6("demo-tool", "https://github.com/acme/demo-tool",
				"python, cli", "A fast demo tool for testing."),
			"Use when you need to test."),
		"github")
	if github.Recall != "Use when you need to test." {
		t.Fatalf("github recall = %q", github.Recall)
	}
	web := RecallFieldsFor(websiteNote6(
		"Palette Gen", "https://palette.example.com",
		"Use when you need to build a color scheme fast."), "websites")
	if web.Recall != "Use when you need to build a color scheme fast." {
		t.Fatalf("web recall = %q", web.Recall)
	}
}

func TestRecallPipelinePromptFilledAndNeutral(t *testing.T) {
	prompt, err := BuildRecallPrompt(ExtractGithubFields(
		githubNote6("demo-tool", "https://github.com/acme/demo-tool",
			"python, cli", "A fast demo tool for testing.")))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(prompt, "demo-tool") {
		t.Fatal("repo name missing from the prompt")
	}
	if strings.Contains(prompt, "{{") { // no unfilled slot
		t.Fatal("unfilled slot")
	}
	if !strings.Contains(prompt, "Use when you need to") {
		t.Fatal("the recall instruction missing")
	}
}

func TestRecallPipelinePlanSkipsHookedAndUnmanaged(t *testing.T) {
	tmp := t.TempDir()
	vault := makeVault(t, tmp, [][2]string{
		{"Dev Tools/a.md", githubNote6("a", "https://github.com/a/a", "t", "x.")},
		{"Dev Tools/b.md", InsertRecallBlock(
			githubNote6("b", "https://github.com/b/b", "t", "x."),
			"Use when you need to b.")},
		{"Dev Tools/unmanaged.md", "# no source here\n"},
	})
	plan := PlanRecallHooks(vault, false)
	var names []string
	for _, p := range plan {
		names = append(names, filepath.Base(p.Path))
	}
	if len(names) != 1 || names[0] != "a.md" {
		t.Fatalf("plan = %v", names)
	}
	forced := PlanRecallHooks(vault, true)
	if len(forced) != 2 {
		t.Fatalf("forced plan = %d entries", len(forced))
	}
}

func TestRecallPipelineRunDryRunThenApply(t *testing.T) {
	frozenClock(t, frozenTime())
	var wrote [][2]string
	writer := func(path, content string) {
		if err := storage.AtomicWriteText(path, content); err != nil {
			t.Fatal(err)
		}
		wrote = append(wrote, [2]string{path, content})
	}
	llm := func(messages []llmMessage) (string, error) {
		return `{"best_used_for": "Use when you need to test things."}`, nil
	}
	tmp := t.TempDir()
	vault := makeVault(t, tmp, [][2]string{
		{"Dev Tools/a.md", githubNote6("a", "https://github.com/a/a", "t", "x.")},
	})
	plan := PlanRecallHooks(vault, false)
	results := RunRecallHooks(plan, llm, false, 0, nil, writer)
	if len(wrote) != 0 {
		t.Fatal("dry-run wrote") // dry-run writes nothing
	}
	if results[0].Sentence != "Use when you need to test things." {
		t.Fatalf("sentence = %q", results[0].Sentence)
	}
	results = RunRecallHooks(plan, llm, true, 0, nil, writer)
	if len(wrote) != 1 {
		t.Fatalf("wrote %d files", len(wrote))
	}
	onDisk, _ := os.ReadFile(wrote[0][0])
	if !HasRecallBlock(string(onDisk)) {
		t.Fatal("hook not on disk")
	}
	// re-run with the SAME sentence → unchanged, no rewrite
	results = RunRecallHooks(PlanRecallHooks(vault, true), llm, true, 0, nil, writer)
	if !results[0].Unchanged {
		t.Fatal("re-run not unchanged")
	}
	if len(wrote) != 1 {
		t.Fatalf("re-run rewrote: %d writes", len(wrote))
	}
}

func TestRecallPipelineRunGarbageLlmGetsPlaceholder(t *testing.T) {
	frozenClock(t, frozenTime())
	llm := func(messages []llmMessage) (string, error) {
		return "I cannot answer that.", nil
	}
	tmp := t.TempDir()
	vault := makeVault(t, tmp, [][2]string{
		{"Dev Tools/a.md", githubNote6("a", "https://github.com/a/a", "t", "x.")},
	})
	results := RunRecallHooks(PlanRecallHooks(vault, false), llm, false, 0, nil, nil)
	if results[0].Sentence != RecallPlaceholder {
		t.Fatalf("sentence = %q", results[0].Sentence)
	}
}

func TestRecallPipelineRunSingleFailureDoesNotAbort(t *testing.T) {
	frozenClock(t, frozenTime())
	llm := func(messages []llmMessage) (string, error) {
		return "", errBoom
	}
	tmp := t.TempDir()
	var notes [][2]string
	for i := 0; i < 3; i++ {
		notes = append(notes, [2]string{
			filepath.Join("Dev Tools", string(rune('n'+rune(i)))+".md"),
			githubNote6("n"+string(rune('0'+rune(i))),
				"https://github.com/x/n"+string(rune('0'+rune(i))), "t", "x."),
		})
	}
	vault := makeVault(t, tmp, notes)
	results := RunRecallHooks(PlanRecallHooks(vault, false), llm, false, 0, nil, nil)
	if len(results) != 3 {
		t.Fatalf("results = %d", len(results))
	}
	for i, r := range results {
		if r.Error == "" {
			t.Fatalf("result %d has no error", i)
		}
	}
	report := RecallReport(results, vault, false)
	if !strings.Contains(report, "ERROR") {
		t.Fatal("report does not carry the errors")
	}
}
