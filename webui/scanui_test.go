// scanui_test.go — T-76's Scan flow binding wrapper wiring smoke: the
// picker's data door, the CTA door's guards (the batch-busy refusal +
// the one-scan law + the picker's safe-cancel quiet line), the REAL
// flow end-to-end over a fixture vault (the job → the bus gate → the
// apply → the finished routing with the scorecard), the held-ask
// answer door (first answer wins), and the frontend's own laws read
// where they render (the fresh-click door's beat, the four-key
// selection, the verbatim texts, the field-tag reads).
package main

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
)

// newScanUI builds a hermetic wrapper: a Settings + Commands pair over
// a fixture repos vault with ONE marked note (the plan the gate will
// ask about), plus a log-capturing bus subscriber.
func newScanUI(t *testing.T) (*ScanUI, *app.Commands, *[]string) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	writeMarkedScanNote(t, vault, "Agents/doomed.md", "https://github.com/octo/doomed")
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{})
	var spoken []string
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		spoken = append(spoken, l.Msg)
	}})
	commands := app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
		Scan: func(ctx context.Context, cfg *storage.OrderedMap, vaults []string,
			ev app.ScanEvents) *app.ScanResult {
			return app.RunVaultScan(cfg, vaults, ev)
		},
		ApplyPlan: func(cfg *storage.OrderedMap, kind string, plan *vaultscan.Plan,
			selection map[string][]string, log app.LogFn) *app.ApplyResult {
			rep := vaultscan.ApplyScanPlan(plan, cfg2VaultPath(cfg), nil, log,
				vaultscan.NormalizeSelection(selection), kind, nil, nil)
			return &app.ApplyResult{Success: true, Report: &rep}
		},
	})
	return NewScanUI(settings, commands), commands, &spoken
}

func cfg2VaultPath(cfg *storage.OrderedMap) string {
	v, _ := cfg.Get("vault_path")
	s, _ := v.(string)
	return s
}

// writeMarkedScanNote writes one app-owned fixture note carrying the
// owner's delete mark (the app_test.go twin — local to this file so
// the webui module's test needs no main-module test helpers).
func writeMarkedScanNote(t *testing.T, vault, rel, url string) string {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	name := url
	if i := strings.LastIndex(url, "/"); i >= 0 {
		name = url[i+1:]
	}
	text := "---\nsource: " + url + "\naliases: []\ntags: [cli]\n" +
		"category: \"Agents\"\ndate_processed: 2026-10-09\nmanaged_by: \"gitcurator\"\n---\n" +
		"# " + name + "\n\nA curated fixture note whose body carries the owner's " +
		"delete mark #auto-delete and enough words for every gate.\n\n---\n" +
		"*Source: [GitHub](" + url + ")*\n"
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestScanUIPickVaultsDataDoor(t *testing.T) {
	ui, _, _ := newScanUI(t)
	model := ui.PickVaults()
	if model == nil || len(model.Rows) != 3 {
		t.Fatalf("model = %+v, want the three rows", model)
	}
	// the repos vault is connected (the fixture); the picker's first
	// pure half already proved the walk — this pins the BINDING's
	// shape + the field-tag law the frontend reads.
	if model.Rows[1].Kind != "repos" || !model.Rows[1].Selectable {
		t.Fatalf("repos row = %+v (connected + tickable)", model.Rows[1])
	}
	if model.Rows[1].StatusText == "" {
		t.Fatal("status_text must ride (the field-tag law)")
	}
	if model.Rows[2].Kind != "personal" || model.Rows[2].Selectable {
		t.Fatalf("personal row = %+v (read-only)", model.Rows[2])
	}
}

func TestScanUIQuietCancelAndGuards(t *testing.T) {
	ui, _, spoken := newScanUI(t)
	// the picker's safe cancel: the quiet line, nothing scanned.
	res := ui.StartScan(nil)
	if res.Started || res.Refusal != "" || res.Result != nil {
		t.Fatalf("quiet cancel = %+v", res)
	}
	if !contains(joinScanMsgs(spoken), "Vault scan cancelled at the picker") {
		t.Fatalf("the quiet line never spoke: %v", *spoken)
	}
	// the one-scan law: a gate that holds makes the second click refuse.
	asked := make(chan struct{}, 1)
	release := make(chan struct{})
	done := make(chan *ScanFlowResult, 1)
	ui.commands.Bus.Subscribe(&app.Handlers{AskScanConfirm: func(e *app.AskScanConfirmEvent) {
		asked <- struct{}{}
		<-release
		e.Answer("declined", nil)
	}})
	go func() { done <- ui.StartScan([]string{"repos"}) }()
	select {
	case <-asked:
	case <-time.After(10 * time.Second):
		t.Fatal("the gate never asked")
	}
	second := ui.StartScan([]string{"repos"})
	if second.Started || !strings.Contains(second.Refusal, "A vault scan is already running") {
		t.Fatalf("second = %+v, want the one-scan refusal", second)
	}
	close(release)
	select {
	case res := <-done:
		if !res.Started || res.Result == nil || res.Result.Verdict != "declined" {
			t.Fatalf("held scan = %+v", res)
		}
	case <-time.After(30 * time.Second):
		t.Fatal("the held scan never finished")
	}
}

func TestScanUIBatchBusyRefusal(t *testing.T) {
	ui, commands, spoken := newScanUI(t)
	_ = ui
	// a batch that holds the one-batch lock (the hermetic seam blocks
	// until released — the busy law is the Commands' own; the ScanUI
	// reads the SAME instance's lock).
	block := make(chan struct{})
	finished := make(chan struct{})
	busyCommands := app.NewCommands(commands.Bus, app.Deps{
		Config: commands.Deps.Config,
		Batch: func(ctx context.Context, cfg *storage.OrderedMap, opts app.BatchOptions, ev app.BatchEvents) *app.BatchResult {
			<-block
			close(finished)
			return &app.BatchResult{Success: true}
		},
	})
	busyUI := NewScanUI(ui.settings, busyCommands)
	go func() {
		_ = busyCommands.SyncRunBatch(context.Background(), app.BatchOptions{URLs: []string{"https://github.com/octo/x"}})
	}()
	// wait for the lock (the acquire is immediate on entry).
	time.Sleep(150 * time.Millisecond)
	res := busyUI.StartScan([]string{"repos"})
	if res.Started || !strings.Contains(res.Refusal, "A batch is running — Scan is available when it finishes.") {
		t.Fatalf("batch-busy = %+v", res)
	}
	if !contains(joinScanMsgs(spoken), "A batch is running") {
		t.Fatalf("the refusal never spoke: %v", *spoken)
	}
	close(block)
	select {
	case <-finished:
	case <-time.After(10 * time.Second):
		t.Fatal("the busy batch never finished")
	}
}

func TestScanUIFlowEndToEndOverTheBusGate(t *testing.T) {
	ui, _, spoken := newScanUI(t)
	// the shell's own wiring: the ask crosses through the ScanUI's
	// handler (the pending held + the payload built), and the answer
	// rides back through the answer door — the production round-trip.
	ui.commands.Bus.Subscribe(&app.Handlers{AskScanConfirm: ui.onAskScanConfirm})
	go func() {
		deadline := time.Now().Add(10 * time.Second)
		for time.Now().Before(deadline) {
			if ui.ScanAnswer("confirmed", map[string][]string{
				app.ScanSectionDeletions:      {"https://github.com/octo/doomed"},
				app.ScanSectionConsolidations: {},
				app.ScanSectionNewFolders:     {},
				app.ScanSectionMoves:          {},
			}) {
				return
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	res := ui.StartScan([]string{"repos"})
	if !res.Started || res.Result == nil || res.Result.Verdict != "confirmed" {
		t.Fatalf("flow = %+v", res)
	}
	if res.Result.Story == nil || res.Result.Story.Applied != 1 {
		t.Fatalf("story = %+v (the marked note applied)", res.Result.Story)
	}
	if res.Scorecard == nil || res.Scorecard.Heading != "The plan is applied" {
		t.Fatalf("scorecard = %+v", res.Scorecard)
	}
	joined := joinScanMsgs(spoken)
	for _, law := range []string{
		"🔍 Vault scan started (GitHub Repos vault)",
		"no answer in 300s = nothing done",
		"the plan is on your screen now",
		"✅ Vault scan complete — 1 note(s) removed",
		"✅ Vault scan (GitHub Repos vault) — 1 repo note(s) removed",
	} {
		if !contains(joined, law) {
			t.Fatalf("the desk missed %q: %v", law, *spoken)
		}
	}
	// the note went through the banishment machinery.
	if res.Result.Story.Applied != 1 {
		t.Fatalf("applied = %d", res.Result.Story.Applied)
	}
}

func TestScanUIHeldAskFirstAnswerWins(t *testing.T) {
	ui, _, _ := newScanUI(t)
	ui.commands.Bus.Subscribe(&app.Handlers{AskScanConfirm: ui.onAskScanConfirm})
	plan := &vaultscan.Plan{VaultKind: "repos", VaultLabel: "GitHub Repos vault"}
	ans := make(chan app.ScanAnswer, 1)
	go func() { ans <- ui.commands.Bus.AskScanConfirm(plan) }()
	// the pending ask appears; the FIRST answer wins, the late one is
	// dropped honestly.
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		if ui.ScanAnswer("declined", nil) {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	if ui.ScanAnswer("confirmed", nil) {
		t.Fatal("a second answer must be dropped (first wins)")
	}
	select {
	case a := <-ans:
		if a.Verdict != app.ScanDeclined {
			t.Fatalf("verdict = %q", a.Verdict)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the gate never returned")
	}
}

func TestScanUIProbeSamples(t *testing.T) {
	ui, _, _ := newScanUI(t)
	// the evidence doors' canned shapes (no owner data): the ask model
	// carries the landable move + the low-confidence flag; the
	// scorecard carries the K-of-N warning + the named failure + the
	// repos section.
	payload := ui.ProbeAskSample()
	if payload.Model == nil || len(payload.Model.Sections) != 4 {
		t.Fatalf("ask payload = %+v", payload)
	}
	foundLandable, foundLow := false, false
	for _, sec := range payload.Model.Sections {
		for _, row := range sec.Rows {
			if strings.Contains(row.Text, "(as C_v1.md)") {
				foundLandable = true
			}
			if row.LowConfidence {
				foundLow = true
			}
		}
	}
	if !foundLandable || !foundLow {
		t.Fatalf("the sample ask lost the landable/low rows: %+v", payload.Model.Sections)
	}
	if payload.TimeoutS != 300 {
		t.Fatalf("timeout = %v", payload.TimeoutS)
	}
	score := ui.ProbeScorecardSample()
	if score == nil || score.Heading != "The plan is applied" {
		t.Fatalf("scorecard = %+v", score)
	}
	if len(score.VaultSections) < 1 {
		t.Fatalf("sections = %+v", score.VaultSections)
	}
	var warned, named bool
	for _, sec := range score.VaultSections {
		for _, row := range sec.Rows {
			if row.Tone == app.ToneWarning {
				warned = true
			}
			if strings.Contains(row.Tip, "Stubborn.md") {
				named = true
			}
		}
	}
	if !warned || !named {
		t.Fatalf("the sample scorecard lost the warning/named rows: %+v", score)
	}
}

// TestScanFlowFrontendLaws reads scanflow.js + index.html where the
// laws render (the T-75 sidebar-law pattern): the fresh-click door's
// beat, the born-disabled verdicts, the four-key selection, the
// countdown + self-close texts, the live arithmetic, the Scan CTA's
// own machine, and the field-tag reads.
func TestScanFlowFrontendLaws(t *testing.T) {
	js, err := readFrontend("scanflow.js")
	if err != nil {
		t.Fatal(err)
	}
	// THE FRESH-CLICK DOOR: the 350 ms beat + the born-dead verdicts +
	// the guard's spine (an unarmed Apply refuses).
	for _, law := range []string{
		"const ARM_MS = 350",
		"keepBtn.disabled = true",
		"applyBtn.disabled = true",
		"if (!planState.armed) return;",
		"setTimeout(() => armPlanDoors(), ARM_MS)",
	} {
		if !contains(js, law) {
			t.Fatalf("scanflow.js missing %q", law)
		}
	}
	// the picking door: the low-confidence pre-uncheck, the master
	// sweep, the four-key selection, the nothing-picked disable.
	for _, law := range []string{
		"box2.checked = !row.low_confidence",
		"const out = { deletions: [], consolidations: [], new_folders: [], moves: [] };",
		"'🗂️ Apply plan (' + picked + ' of ' + total + ')'",
		"planState.armed && (picked > 0 || total === 0)",
		"section.master.indeterminate = any && !all",
	} {
		if !contains(js, law) {
			t.Fatalf("scanflow.js missing %q", law)
		}
	}
	// the reference's own texts, verbatim (the model-side texts —
	// the title, the header, the hint — are pinned by the Go model
	// battery; the JS laws here are the renderer's own).
	for _, law := range []string{
		"'🔍 Scan this vault'",
		"'🔍 Scan both vaults'",
		"at zero nothing is done (the same safe defer as a silent Telegram ask)",
		"This scorecard closes itself in",
		"more (counted, not listed)",
		"the LLM itself called this move low-confidence",
		"Nothing is scanned and nothing is touched — the Scan button stays ready for the next click.",
		"The safe default — nothing is deleted, nothing moves, no folder is created",
	} {
		if !contains(js, law) {
			t.Fatalf("scanflow.js missing the verbatim law %q", law)
		}
	}
	// THE FIELD-TAG LAW: the frontend reads the TAGGED names.
	for _, law := range []string{
		"row.status_text",
		"payload.timeout_s",
		"row.low_confidence",
		"model.vault_sections",
		"model.nothing_landed",
		"sec.changed + ' change(s)'",
	} {
		if !contains(js, law) {
			t.Fatalf("scanflow.js missing the field-tag read %q", law)
		}
	}
	// the Scan CTA's own machine + the scorecard's self-close.
	for _, law := range []string{
		"const AUTO_CLOSE_S = 12",
		"getElementById('btn-scan')",
		"openScanPicker",
		"'Scanning…'",
		"StartScan(picked || null)",
	} {
		if !contains(js, law) {
			t.Fatalf("scanflow.js missing %q", law)
		}
	}
	// index.html wires the module + the live Scan button.
	html, err := readFrontend("index.html")
	if err != nil {
		t.Fatal(err)
	}
	if !contains(html, `src="scanflow.js"`) {
		t.Fatal("index.html missing the scanflow.js module")
	}
	if !contains(html, `<button id="btn-scan" class="btn-secondary" aria-label="Scan the vault">🔍 Scan</button>`) {
		t.Fatal("the Scan CTA must ride its own live machine (no disabled placeholder)")
	}
}

func joinScanMsgs(msgs *[]string) string {
	if msgs == nil {
		return ""
	}
	return strings.Join(*msgs, "\n")
}
