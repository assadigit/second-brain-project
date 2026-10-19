// syncuib.go — T-75 slice (c): the Sync page's binding wrapper (the
// fourth wails struct, bound in main.go): the ONE Refresh door (one
// payload, both desks — the flow cards above + the gap matrix below),
// the differences modal's doors, the fix doors' queue-check leg, and
// the deferred-log reads. Every engine law lives in internal/app
// (syncdisplay.go + syncview.go); this file binds it to the window.
package main

import (
	"context"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/syncreview"
)

// SyncUI is the Sync page's bound surface (the wailsjs bindings land
// in frontend/wailsjs/go/main/SyncUI.js).
type SyncUI struct {
	settings *app.Settings
	commands *app.Commands
	ctx      context.Context
}

// NewSyncUI wires the wrapper over the app's Settings + Commands.
func NewSyncUI(settings *app.Settings, commands *app.Commands) *SyncUI {
	return &SyncUI{settings: settings, commands: commands}
}

// bindCtx captures the wails context at startup.
func (u *SyncUI) bindCtx(ctx context.Context) {
	u.ctx = ctx
}

// Refresh is the ONE CONSENT DOOR (v0.87.0's convergence): the flow's
// four elements AND the stage matrix, measured together — the local
// walks always run; the live doors (the bot queue, the published
// site) are asked only because the owner clicked. Read-only.
func (u *SyncUI) Refresh() *app.SyncViewPayload {
	return u.settings.RefreshSyncView()
}

// MarkResult is the modal's Mark door's payload (wails bindings carry
// one return value — the struct folds the count + the confirm line).
type MarkResult struct {
	Marked int    `json:"marked"`
	Line   string `json:"line"`
}

// MarkIntentional is the differences modal's Mark as intentional
// button: bulk marking with ONE shared reason (the disabled-until law
// is the frontend's gate; the door re-checks it — never a lie in).
func (u *SyncUI) MarkIntentional(stageKey string, itemIDs []string,
	reason, note string, displayNames map[string]string) *MarkResult {
	n, line := u.settings.SyncMarkIntentional(stageKey, itemIDs, reason,
		note, displayNames)
	return &MarkResult{Marked: n, Line: line}
}

// ReopenResult is the modal's Reopen door's payload.
type ReopenResult struct {
	Ok   bool   `json:"ok"`
	Line string `json:"line"`
}

// ReopenItem is the deferred log's Reopen button (Active rows only).
func (u *SyncUI) ReopenItem(stageKey, itemID string) *ReopenResult {
	ok, line := u.settings.SyncReopenItem(stageKey, itemID)
	return &ReopenResult{Ok: ok, Line: line}
}

// StageDifferences answers the differences modal's data for one stage
// (the record + the heading + the log rows — pure local, instant).
func (u *SyncUI) StageDifferences(stageKey string) *app.SyncDifferencesPayload {
	return u.settings.SyncDifferences(stageKey)
}

// DeferredRows answers the log tab's rows (search included).
func (u *SyncUI) DeferredRows(stage, query string) []syncreview.LedgerRow {
	return u.settings.SyncDeferredRows(stage, query)
}

// LedgerCount answers the ledger's own row count.
func (u *SyncUI) LedgerCount() int {
	return u.settings.SyncLedgerCount()
}

// StageTables answers the matrix's shape (the defs + the layout + the
// columns) — the frontend renders from the reference's own tables.
func (u *SyncUI) StageTables() *app.SyncStageTables {
	return app.TheSyncStageTables()
}

// SyncQueueCheckResult is the matrix's "Check the bot queue" fix
// door's payload (the queue stages' data lands with it — the
// reference's own hook).
type SyncQueueCheckResult struct {
	Refusal string `json:"refusal"`
	Error   string `json:"error"`
	Report  string `json:"report"`
}

// CheckBotQueue is the fix door behind the queue stages' cells: the
// same engine the Bot page's own button rides (the username guard +
// the honest poll + the classification + the render), with the queue
// state landed for the next gather. The username rides the LIVE
// config (the Sync page has no username field of its own).
func (u *SyncUI) CheckBotQueue() *SyncQueueCheckResult {
	username := u.settings.BotUsernameLive()
	if line := app.BotUsernameGuard(username); line != "" {
		u.settings.SpeakOnBus(line, "error")
		return &SyncQueueCheckResult{Refusal: line}
	}
	u.settings.SpeakOnBus("📬 Checking bot queue (@"+
		app.TrimAt(username)+")...", "info")
	rep := u.commands.SyncFetchQueue(context.Background())
	if rep.State != "ok" {
		u.settings.SpeakOnBus("❌ Bot queue check failed: "+rep.Error, "error")
		return &SyncQueueCheckResult{Error: rep.Error,
			Report: "Error: " + rep.Error}
	}
	for _, l := range app.QueueLogLines(rep) {
		u.settings.SpeakOnBus(l.Msg, l.Level)
	}
	// the queue stages' data is fresh — land it (the reference's
	// _refresh_sync_after_queue hook, folded into the door itself)
	u.settings.SetSyncQueueState(&app.SyncQueueState{
		Checked:         true,
		PendingRepos:    rep.PendingURLs,
		PendingWebsites: rep.PendingWebsites,
		InVault:         rep.InVault,
		WebsitesInVault: rep.WebsitesInVault,
	})
	return &SyncQueueCheckResult{
		Report: app.RenderQueueReport(u.settings.PlainConfig(), rep),
	}
}

// SealNow is the Vault → Backup stages' fix door (the Backup page's
// own engine; the re-measure arming is the frontend's 4 s timer).
func (u *SyncUI) SealNow(kind string) *app.SealResultRow {
	return u.commands.BackupSealVault(context.Background(), kind)
}

// PublishNow is the Vault → Website stage's fix door (the publish
// leg's own engine).
func (u *SyncUI) PublishNow() *app.SealResultRow {
	return u.commands.BackupPublish(context.Background())
}

// OpenNote is the modal's per-row "Open note" door (verify before
// excusing — that is the point of a review).
func (u *SyncUI) OpenNote(path string) bool {
	return u.settings.OpenNote(path)
}
