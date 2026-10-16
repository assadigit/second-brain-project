package connectioncheck

// battery_test.go — run_local_checks (the RE-SCOPED four sections:
// Vaults, LLM, GitHub, Worker) + summarize (the re-worded headline
// pinned exactly) + the mutability law + the callback contracts.
// Ported from the reference's TestRunLocalChecks, TestSummarize and
// the two battery/worker-line tests.

import (
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
)

func TestRunLocalChecksFourSectionsInOrderWithCallbacks(t *testing.T) {
	var sectionsSeen []struct {
		title string
		idx   int
		total int
	}
	var resultsSeen []Result
	sections := RunLocalChecks(map[string]any{},
		func(title string, idx, total int) {
			sectionsSeen = append(sectionsSeen,
				struct {
					title string
					idx   int
					total int
				}{title, idx, total})
		},
		func(r Result) { resultsSeen = append(resultsSeen, r) })
	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	// THE RE-SCOPED SECTION LIST (ADR-0004 Am.3/R-0): the Telegram
	// section is GONE; the Worker leg is promoted to queue-reachability.
	wantTitles := []string{"Vaults", "LLM", "GitHub", "Worker"}
	if strings.Join(titles, "|") != strings.Join(wantTitles, "|") {
		t.Fatalf("titles = %v, want %v", titles, wantTitles)
	}
	if len(sectionsSeen) == 0 ||
		sectionsSeen[0].title != "Vaults" || sectionsSeen[0].idx != 1 ||
		sectionsSeen[0].total != 4 {
		t.Errorf("first callback = %+v", sectionsSeen[0])
	}
	last := sectionsSeen[len(sectionsSeen)-1]
	if last.title != "Worker" || last.idx != 4 || last.total != 4 {
		t.Errorf("last callback = %+v", last)
	}
	if len(resultsSeen) == 0 {
		t.Fatal("no results reached the callback")
	}
	valid := map[string]bool{
		LevelOK: true, LevelWarn: true, LevelError: true, LevelInfo: true}
	for _, r := range resultsSeen {
		if !valid[r.Level] {
			t.Errorf("result level %q outside the vocabulary", r.Level)
		}
	}
}

func TestRunLocalChecksHostileConfigNeverPanics(t *testing.T) {
	sections := RunLocalChecks(map[string]any{
		"pipelines": "junk", "ollama": "junk",
		"vaultseal": 3, "proxy": "junk"}, nil, nil)
	if len(sections) != 4 {
		t.Fatalf("sections = %d, want 4", len(sections))
	}
}

func TestRunLocalChecksCrashedSectionDegradesToErrorRow(t *testing.T) {
	// a check that panics must surface as the section's crash row,
	// never an empty-sections "ALL SYSTEMS READY" verdict — the crash
	// law itself (guardedResults is the battery's guard for every
	// section)
	rows := guardedResults("Vaults", func() []Result { panic("boom") })
	if len(rows) != 1 {
		t.Fatalf("crash section rows = %d, want 1", len(rows))
	}
	crash := rows[0]
	if crash.Name != "Vaults" || crash.Level != LevelError {
		t.Errorf("crash row = (%s, %s)", crash.Name, crash.Level)
	}
	if !strings.Contains(crash.Detail, "check crashed") {
		t.Errorf("detail: %q", crash.Detail)
	}
	// and the guard never leaks the panic
	rows2 := guardedResults("LLM", func() []Result {
		return []Result{okRow()}
	})
	if len(rows2) != 1 || rows2[0].Level != LevelOK {
		t.Errorf("normal path broken: %+v", rows2)
	}
}

func TestRunLocalChecksCallbackPanicsAreSwallowed(t *testing.T) {
	// the reference swallows callback exceptions — the Go law: a
	// panicking callback never kills the battery
	sections := RunLocalChecks(map[string]any{},
		func(title string, idx, total int) { panic("section cb") },
		func(r Result) { panic("result cb") })
	if len(sections) != 4 {
		t.Fatalf("sections = %d, want 4", len(sections))
	}
}

func TestRunLocalChecksSectionsAreMutableForTheLiveLeg(t *testing.T) {
	sections := RunLocalChecks(map[string]any{}, nil, nil)
	sections[len(sections)-1].Results = append(
		sections[len(sections)-1].Results,
		Result{Name: "Live connection", Level: LevelOK, Detail: "x"})
	last := sections[len(sections)-1].Results
	if last[len(last)-1].Name != "Live connection" {
		t.Errorf("append failed: %+v", last)
	}
}

func TestRunLocalChecksAppendsTheWorkerLineToTheWorkerSection(t *testing.T) {
	// the promoted leg joins the battery as the Worker section's line
	// (the reference appended it to the Telegram section — ADR-0004
	// Am.3 re-homes it)
	withSeam(t, &workerDirectGet,
		func(u string, timeoutS float64) (any, error) {
			return map[string]any{"version": constants.ExpectedWorkerVersion}, nil
		})
	sections := RunLocalChecks(map[string]any{
		"cloudflare_worker_url": "https://bot.example"}, nil, nil)
	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	want := []string{"Vaults", "LLM", "GitHub", "Worker"}
	if strings.Join(titles, "|") != strings.Join(want, "|") {
		t.Fatalf("titles = %v", titles)
	}
	var names []string
	for _, r := range sections[3].Results {
		names = append(names, r.Name)
	}
	if !containsStr(names, "Bot Worker") {
		t.Errorf("Worker section rows = %v", names)
	}
}

func TestRunLocalChecksWithoutWorkerURLKeepsPlainEmptyWorkerSection(t *testing.T) {
	sections := RunLocalChecks(map[string]any{}, nil, nil)
	if sections[3].Title != "Worker" {
		t.Fatalf("section 3 = %s", sections[3].Title)
	}
	if len(sections[3].Results) != 0 {
		t.Errorf("Worker section rows = %+v, want none", sections[3].Results)
	}
	for _, s := range sections {
		for _, r := range s.Results {
			if r.Name == "Bot Worker" {
				t.Errorf("unexpected Bot Worker row: %+v", r)
			}
		}
	}
}

func okRow() Result { return result("x", LevelOK, "") }

func TestSummarizeAllOk(t *testing.T) {
	s := Summarize([]Section{
		{Title: "A", Results: []Result{okRow()}},
		{Title: "B", Results: []Result{okRow()}},
	})
	if s.Level != LevelOK {
		t.Errorf("level = %s", s.Level)
	}
	if s.Ready != 2 {
		t.Errorf("ready = %d", s.Ready)
	}
	if !strings.Contains(s.Headline, "ALL SYSTEMS READY") {
		t.Errorf("headline: %q", s.Headline)
	}
}

// THE RE-WORDED HEADLINE (pinned exactly per the spec's banner: the
// Python headline's "Telegram" mention goes — the sync queue rides the
// Worker; T-33's own result-shape tests pin the exact wording).
func TestSummarizeTheReWordedHeadline(t *testing.T) {
	s := Summarize([]Section{
		{Title: "Vaults", Results: []Result{okRow()}},
		{Title: "LLM", Results: []Result{okRow()}},
		{Title: "GitHub", Results: []Result{okRow()}},
		{Title: "Worker", Results: []Result{okRow()}},
	})
	want := "ALL SYSTEMS READY — vaults, LLM, GitHub and the sync queue are up"
	if s.Headline != want {
		t.Errorf("headline = %q, want %q", s.Headline, want)
	}
}

func TestSummarizeEmptyWorkerSectionIsReady(t *testing.T) {
	// the no-URL Worker section (zero rows) counts as ready — the
	// reference's empty-section law
	s := Summarize([]Section{
		{Title: "Vaults", Results: []Result{okRow()}},
		{Title: "Worker"},
	})
	if s.Level != LevelOK || s.Ready != 2 || s.Total != 2 {
		t.Errorf("summary = %+v", s)
	}
}

func TestSummarizeWarnBeatsOk(t *testing.T) {
	s := Summarize([]Section{
		{Title: "A", Results: []Result{okRow()}},
		{Title: "B", Results: []Result{result("x", LevelWarn, "")}},
	})
	if s.Level != LevelWarn {
		t.Errorf("level = %s", s.Level)
	}
	if s.Ready != 1 {
		t.Errorf("ready = %d", s.Ready)
	}
	if !strings.Contains(s.Headline, "1 warning(s)") {
		t.Errorf("headline: %q", s.Headline)
	}
}

func TestSummarizeErrorBeatsWarn(t *testing.T) {
	s := Summarize([]Section{
		{Title: "A", Results: []Result{result("x", LevelWarn, "")}},
		{Title: "B", Results: []Result{result("x", LevelError, "")}},
	})
	if s.Level != LevelError {
		t.Errorf("level = %s", s.Level)
	}
	if !strings.Contains(s.Headline, "1 error(s)") {
		t.Errorf("headline: %q", s.Headline)
	}
	if !strings.Contains(s.Headline, "1 warning(s)") {
		t.Errorf("headline lacks the warning bit: %q", s.Headline)
	}
	if !strings.Contains(s.Headline, "0/2 subsystems ready") {
		t.Errorf("headline: %q", s.Headline)
	}
}

func TestSummarizeEmptySectionsSafe(t *testing.T) {
	s := Summarize(nil)
	if s.Level != LevelOK {
		t.Errorf("level = %s", s.Level)
	}
	s2 := Summarize([]Section{})
	if s2.Level != LevelOK {
		t.Errorf("level = %s", s2.Level)
	}
}

func TestGUILevelsMapping(t *testing.T) {
	// the level mapping reaches the GUI vocabulary (identical for the
	// CLI)
	want := map[string]string{
		LevelOK: "success", LevelWarn: "warning",
		LevelError: "error", LevelInfo: "info",
	}
	for lv, gui := range want {
		if GUILevels[lv] != gui {
			t.Errorf("GUILevels[%s] = %s, want %s", lv, GUILevels[lv], gui)
		}
	}
}
