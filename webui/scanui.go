// scanui.go — T-76: the Scan flow's binding wrapper (the FIFTH bound
// struct, bound in main.go): the vault picker's data door, the CTA
// flow's one door (the guards → the start line → the job over the
// chosen scope → the finished routing), and the plan modal's answer
// door over the bus's held ask. Every engine law lives in internal/app
// (scandialogs.go + scanjob.go); this file binds the flow to the
// window: the plan display model crosses to the frontend as the
// scan:ask event (the shell's bus subscription hands it here), the
// frontend answers through ScanAnswer, and the finished story crosses
// as the scan:finished event (the desk lines ride the bus as they
// always have).
package main

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// ScanUI is the Scan flow's bound surface (the wailsjs bindings land
// in frontend/wailsjs/go/main/ScanUI.js).
type ScanUI struct {
	settings *app.Settings
	commands *app.Commands
	ctx      context.Context

	mu      sync.Mutex
	pending *app.AskScanConfirmEvent // the held ask (first answer wins)
	running bool                     // the one-scan-at-a-time law
}

// NewScanUI wires the wrapper over the app's Settings + Commands.
func NewScanUI(settings *app.Settings, commands *app.Commands) *ScanUI {
	return &ScanUI{settings: settings, commands: commands}
}

// bindCtx captures the wails context at startup.
func (u *ScanUI) bindCtx(ctx context.Context) {
	u.ctx = ctx
}

// ScanAskPayload is the scan:ask event's payload: the plan's display
// model + the answering window (the modal's countdown ticks it).
type ScanAskPayload struct {
	Model    *app.PlanDisplayModel `json:"model"`
	TimeoutS float64               `json:"timeout_s"`
}

// onAskScanConfirm is the shell's bus-wired handler (AskScanConfirm):
// the plan becomes the display model (pure), the model + the window
// cross to the frontend, and the event is held for the answer door.
// The bus's own 300 s window is the belt under the modal's countdown.
func (u *ScanUI) onAskScanConfirm(ev *app.AskScanConfirmEvent) {
	plan, _ := ev.Plan.(*vaultscan.Plan)
	payload := &ScanAskPayload{
		Model:    app.PlanDisplayModelOf(plan),
		TimeoutS: app.ScanConfirmTimeoutS(u.liveConfig()),
	}
	u.mu.Lock()
	u.pending = ev
	u.mu.Unlock()
	if u.ctx != nil {
		runtime.EventsEmit(u.ctx, "scan:ask", payload)
	}
}

func (u *ScanUI) liveConfig() *storage.OrderedMap {
	if u.settings != nil {
		if cfg := u.settings.PlainConfig(); cfg != nil {
			return cfg
		}
	}
	return storage.NewOrderedMap()
}

// ScanAnswer is the plan modal's answer door: the verdict (and, only
// on a confirmed verdict, the four section key sets — the shape law
// the bus itself enforces). The first answer wins; a late or orphan
// answer is dropped honestly (false).
func (u *ScanUI) ScanAnswer(verdict string, selection map[string][]string) bool {
	u.mu.Lock()
	ev := u.pending
	u.pending = nil
	u.mu.Unlock()
	if ev == nil {
		return false
	}
	ev.Answer(verdict, selection)
	return true
}

// ScanFlowResult is the CTA door's one answer: the guard's refusal,
// the finished story, and the scorecard model when the glance opens.
type ScanFlowResult struct {
	Started   bool                    `json:"started"`
	Refusal   string                  `json:"refusal"`
	Result    *app.ScanResult         `json:"result"`
	Scorecard *app.ResultDisplayModel `json:"scorecard"`
}

// scanVaultNames maps the picked kinds onto the desk's own names.
func scanVaultNames(kinds []string) string {
	names := map[string]string{
		"websites": "Websites vault",
		"repos":    "GitHub Repos vault",
	}
	parts := []string{}
	for _, k := range kinds {
		if n, ok := names[k]; ok {
			parts = append(parts, n)
		} else {
			parts = append(parts, k)
		}
	}
	if len(parts) == 0 {
		return "the chosen vault"
	}
	return strings.Join(parts, " + ")
}

// StartScan is the Scan CTA's whole flow (vault_scan_ui.py's
// _start_vault_scan + _vault_scan_finished, one door): the guards (a
// batch running / a scan already running), the picker's answer (nil =
// the safe cancel — the quiet line, nothing scanned), the start line,
// the job over the CHOSEN scope (the plan modal opens mid-flight over
// the bus gate), and the finished routing (the desk lines + the
// scorecard when the vault changed). The binding call resolves when
// the story ends — the desk keeps living underneath (the events flow
// on their own transport).
func (u *ScanUI) StartScan(picked []string) *ScanFlowResult {
	speak := func(msg, level string) {
		if u.settings != nil {
			u.settings.SpeakOnBus(msg, level)
		}
	}
	// the reference's own guards, verbatim.
	if u.commands != nil && u.commands.BatchBusy() {
		line := "⏳ A batch is running — Scan is available when it finishes."
		speak(line, "warning")
		return &ScanFlowResult{Refusal: line}
	}
	u.mu.Lock()
	if u.running {
		u.mu.Unlock()
		line := "⏳ A vault scan is already running — its Telegram ask is waiting (or its plan is being applied)."
		speak(line, "warning")
		return &ScanFlowResult{Refusal: line}
	}
	u.running = true
	u.mu.Unlock()
	defer func() {
		u.mu.Lock()
		u.running = false
		u.mu.Unlock()
	}()

	// v0.84.0 — the picker's answer: nil/empty is the safe cancel (the
	// quiet line, the button stays armed, nothing is scanned).
	kinds := app.PickedScanKinds(picked)
	if kinds == nil {
		speak("👌 Vault scan cancelled at the picker — nothing was scanned, nothing touched (the button stays ready)", "info")
		return &ScanFlowResult{}
	}
	speak(fmt.Sprintf("🔍 Vault scan started (%s) — reading the vault, then the plan appears here for your confirmation BEFORE anything is deleted or moved (no answer in 300s = nothing done).",
		scanVaultNames(kinds)), "info")

	res := u.commands.ScanRun(context.Background(), kinds)

	// ---- the finished routing (_vault_scan_finished, verbatim) -----
	out := &ScanFlowResult{Started: true, Result: res}
	switch res.Verdict {
	case "clean":
		if res.HygieneTotal > 0 {
			speak(fmt.Sprintf("✅ Vault scan: the library is clean — and the hygiene pass healed %d stray file(s) (recoverable in .trash).",
				res.HygieneTotal), "success")
			out.Scorecard = app.ResultDisplayModelOf(res)
		} else {
			speak("✅ Vault scan: the library is clean — nothing to ask, nothing to do.", "success")
		}
	case "confirmed":
		st := res.Story
		if st == nil {
			st = &app.ScanStory{}
		}
		speak(fmt.Sprintf("✅ Vault scan complete — %d note(s) removed, %d re-filed, %d new folder(s).",
			st.Applied, st.Moved, st.FoldersCreated), "success")
		// v0.83.0 — the repos story speaks its own line.
		if res.Repos != nil {
			speak(fmt.Sprintf("✅ Vault scan (GitHub Repos vault) — %d repo note(s) removed, %d re-filed, %d new folder(s).",
				res.Repos.Applied, res.Repos.Moved, res.Repos.FoldersCreated), "success")
		}
		out.Scorecard = app.ResultDisplayModelOf(res)
	case "error":
		why := res.Error
		if why == "" {
			why = "?"
		}
		speak("💥 Vault scan failed: "+why, "error")
	}
	// declined / timeout / defer already spoke their own lines in the
	// job — the button is the only state left to restore (the
	// frontend's await resolves here).
	if u.ctx != nil {
		runtime.EventsEmit(u.ctx, "scan:finished", out)
	}
	return out
}

// PickVaults is the picker's data door: the three vaults' detection
// rows become the display model (pure, instant — the picker modal
// renders it; nothing is scanned).
func (u *ScanUI) PickVaults() *app.PickDisplayModel {
	return app.PickDisplayModelOf(app.DetectScanVaults(u.liveConfig(), true))
}

// ProbeAskSample is the evidence leg's canned plan ask (no owner
// data — the deterministic sample shape): emits the scan:ask event
// over a fixture-shaped plan display model so the plan modal's
// rendering (the masters, the arm-guard, the landable move) is
// screenshot-provable without a live vault. The held gate is NOT
// touched — the modal's answer finds no pending ask and is dropped
// honestly (the probe's modal is a rendering, never a live gate).
func (u *ScanUI) ProbeAskSample() *ScanAskPayload {
	plan := &vaultscan.Plan{
		VaultKind:  "websites",
		VaultLabel: "Websites vault",
		Deletions: []websitepipeline.PendingBanishment{
			{URL: "https://example.com/a", Canonical: "https://example.com/a",
				Title: "A note", Marker: "frontmatter: auto-delete", Door: "note tag"},
		},
		NewFolders: []string{"Tools/Scraping"},
		Moves: []vaultscan.MoveRow{
			{Note: "B.md", Path: "B.md", To: "Tools/Scraping", Reason: "the filing eyes"},
			{Note: "C.md", Path: "C.md", From: "Agents", To: "Tools/Scraping",
				As: "C_v1.md", Reason: "a colliding landing name", LowConfidence: true},
		},
		Consolidations: []vaultscan.Consolidation{
			{Site: "example.com", Keep: vaultscan.KeepRow{File: "Example.md"},
				Merge: []vaultscan.MergeRow{{File: "Example 2.md", Path: "Example 2.md",
					Source: "https://example.com/x"}}},
		},
		Summary: "The LLM's own sample summary — the fixture shape, never a live vault.",
		Inventory: vaultscan.PlanSummary{TotalNotes: 12, FolderCount: 3,
			RootNotes: 2, UncategorizedNotes: 4, TrashNotes: 1, LockedNotes: 5},
	}
	payload := &ScanAskPayload{
		Model:    app.PlanDisplayModelOf(plan),
		TimeoutS: 300.0,
	}
	if u.ctx != nil {
		runtime.EventsEmit(u.ctx, "scan:ask", payload)
	}
	return payload
}

// ProbeScorecardSample is the evidence leg's canned scorecard (no
// owner data): the confirmed shape with the honest 'K of N' warning,
// the named failed move, and the repos sub-story's own section — the
// screenshot's richest honest sample.
func (u *ScanUI) ProbeScorecardSample() *app.ResultDisplayModel {
	res := &app.ScanResult{Verdict: "confirmed",
		Vault: "/vaults/websites", VaultKind: "websites",
		VaultLabel: "Websites vault", GitHubVault: "/vaults/repos",
		Story: &app.ScanStory{Deletions: 5, Applied: 3, Moves: 4, Moved: 2,
			FailedMoves: 1, Failed: []vaultscan.FailedMove{
				{Note: "Stubborn.md", To: "Tools/Scraping", Why: "the file was locked"},
			}},
		Repos: &app.ScanStory{Deletions: 1, Applied: 1, Moves: 1, Moved: 1}}
	return app.ResultDisplayModelOf(res)
}
