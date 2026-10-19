// settingsuib.go — T-74: the Settings window's pages B binding wrapper
// (the thin shell over internal/app's pages B engines + the Commands
// surface — doc 04 §3's law: "ui — thin shells over app"). Every
// engine law lives in internal/app; this file binds it to the wails
// window: the Bot page's Check Queue + Process All, the Dashboard
// page's refresh + quarantine + undo, the Sources page's fetch +
// process, and the Backup page's saves + ladders + job buttons.
package main

import (
	"context"

	"github.com/assadigit/second-brain-project/internal/app"
)

// PagesBUI is the pages B's bound surface (the wailsjs bindings land
// in frontend/wailsjs/go/main/PagesBUI.js).
type PagesBUI struct {
	settings *app.Settings
	commands *app.Commands
	ctx      context.Context
}

// NewPagesBUI wires the wrapper over the app's Settings + Commands
// surfaces.
func NewPagesBUI(settings *app.Settings, commands *app.Commands) *PagesBUI {
	return &PagesBUI{settings: settings, commands: commands}
}

// bindCtx captures the wails context at startup.
func (u *PagesBUI) bindCtx(ctx context.Context) {
	u.ctx = ctx
}

// ---------------------------------------------------------------------------
// The Bot page (bot_queue.py — the Worker pending source, Option B).
// ---------------------------------------------------------------------------

// SaveBotGroup commits the username + token fields (the save-on-change
// law's Bot group).
func (u *PagesBUI) SaveBotGroup(vals *app.BotValues) *app.SaveResult {
	return u.settings.SaveBotGroup(vals)
}

// QueueCheckResult is the Check Queue door's payload: the rendered
// report + the badge + the classification the Process door consumes.
type QueueCheckResult struct {
	// Refusal is the page's own guard line ("" = the check ran).
	Refusal string `json:"refusal"`
	// Error carries the fetch's honest failure (state error/disabled).
	Error string `json:"error"`
	// Report is the rendered queue display (the canned-result law).
	Report string `json:"report"`
	// Badge is the ⏳/✅ badge (hidden while Refusal/Error speak).
	Badge *app.QueueBadgeState `json:"badge"`
	// PendingURLs / PendingWebsites / counts feed the Process door.
	PendingURLs     []string `json:"pendingUrls"`
	PendingWebsites []string `json:"pendingWebsites"`
	Duplicates      int      `json:"duplicates"`
	RawCount        int      `json:"rawCount"`
}

// CheckBotQueue is the Check Queue button (the Worker pending poll +
// the classification + the render + the badge + the desk lines —
// ui.py's flow minus the Telethon legs, per ADR-0004 Am.3).
func (u *PagesBUI) CheckBotQueue(username string) *QueueCheckResult {
	if line := app.BotUsernameGuard(username); line != "" {
		u.settings.SpeakOnBus(line, "error")
		return &QueueCheckResult{Refusal: line}
	}
	// The reference saves the config before the fetch (the field's
	// value persists even when the fetch fails).
	_ = u.settings.SaveBotGroup(&app.BotValues{Username: username})
	u.settings.SpeakOnBus("📬 Checking bot queue (@"+
		app.TrimAt(username)+")...", "info")

	rep := u.commands.SyncFetchQueue(context.Background())
	if rep.State != "ok" {
		u.settings.SpeakOnBus("❌ Bot queue check failed: "+rep.Error, "error")
		return &QueueCheckResult{Error: rep.Error,
			Report: "Error: " + rep.Error}
	}
	for _, l := range app.QueueLogLines(rep) {
		u.settings.SpeakOnBus(l.Msg, l.Level)
	}
	badge := app.QueueBadge(rep)
	return &QueueCheckResult{
		Report:          app.RenderQueueReport(u.settings.PlainConfig(), rep),
		Badge:           &badge,
		PendingURLs:     rep.PendingURLs,
		PendingWebsites: rep.PendingWebsites,
		Duplicates:      rep.DuplicatesRemoved,
		RawCount:        rep.RawCount,
	}
}

// ProcessBotQueue is the Process All button: the settled-exclusion
// second layer (the owner's law — settled links never enter a batch),
// then the batch itself (the confirm gate is the frontend's —
// >10 items asks first). The result is the batch's own story.
func (u *PagesBUI) ProcessBotQueue(pending []string, websites []string,
	duplicates, rawCount int) *app.BatchResult {
	return u.commands.SyncRunBatch(context.Background(), app.BatchOptions{
		URLs:             pending,
		NonGithub:        websites,
		Source:           "bot",
		IntakeDuplicates: duplicates,
		RawURLCount:      rawCount,
	})
}

// ProcessSourcesUrls is the Sources page's Process Fetched URLs
// button (the confirm gate is the frontend's).
func (u *PagesBUI) ProcessSourcesUrls(urls []string) *app.BatchResult {
	return u.commands.SyncRunBatch(context.Background(), app.BatchOptions{
		URLs:   urls,
		Source: "direct",
	})
}

// ---------------------------------------------------------------------------
// The Dashboard page (dashboard.py).
// ---------------------------------------------------------------------------

// DashboardViewResult is the Refresh Dashboard door's payload.
type DashboardViewResult struct {
	// NoVault mirrors the reference's "No vault selected." panel text.
	NoVault    bool   `json:"noVault"`
	Text       string `json:"text"`
	Quarantine string `json:"quarantine"`
}

// RefreshDashboard scans the vault + renders the stats + the quarantine
// table (update_dashboard's flow, the desk lines included).
func (u *PagesBUI) RefreshDashboard() *DashboardViewResult {
	vault := u.settings.LiveVaultPath()
	if vault == "" {
		return &DashboardViewResult{NoVault: true, Text: "No vault selected.",
			Quarantine: app.QuarantineView(u.settings.PlainConfig())}
	}
	u.settings.SpeakOnBus("📊 Refreshing dashboard...", "info")
	stats := app.ScanDashboardStats(vault)
	u.settings.SpeakOnBus("✅ Dashboard updated: "+
		itoa(stats.TotalNotes)+" notes, "+itoa(len(stats.Categories))+
		" categories", "success")
	return &DashboardViewResult{
		Text:       app.RenderDashboardText(vault, stats),
		Quarantine: app.QuarantineView(u.settings.PlainConfig()),
	}
}

// ClearQuarantine resets every 404 counter (the themed confirm is the
// frontend's) + returns the refreshed view.
func (u *PagesBUI) ClearQuarantine() *DashboardViewResult {
	removed, ok := app.ClearQuarantine(u.settings.PlainConfig())
	if !ok {
		u.settings.SpeakOnBus("⚠️ Could not reset the 404 quarantine", "warning")
	} else {
		u.settings.SpeakOnBus(app.ResetQuarantineLine(removed), "success")
	}
	return &DashboardViewResult{Quarantine: app.QuarantineView(u.settings.PlainConfig())}
}

// SaveQuarantineGroup persists the threshold spinbox (the
// valueChanged save).
func (u *PagesBUI) SaveQuarantineGroup(vals *app.QuarantineValues) *app.SaveResult {
	return u.settings.SaveQuarantineGroup(vals)
}

// ReadUndoLastBatch loads the undo facts (the themed dialogs').
func (u *PagesBUI) ReadUndoLastBatch() *app.UndoBatchResult {
	return app.ReadUndoList(u.settings.LiveVaultPath())
}

// ApplyUndoLastBatch performs the undo (the confirm already answered).
func (u *PagesBUI) ApplyUndoLastBatch() *app.UndoBatchResult {
	res := app.ApplyUndoLastBatch(u.settings.LiveVaultPath())
	if res.Stage == "ok" {
		if line, level := app.UndoLogLine(res.Deleted, res.Errors); line != "" {
			u.settings.SpeakOnBus(line, level)
		}
	}
	return res
}

// ---------------------------------------------------------------------------
// The Sources page (bot_queue.py 72-140).
// ---------------------------------------------------------------------------

// FetchFromSources pulls the feed + extracts the GitHub URLs.
func (u *PagesBUI) FetchFromSources(url string) *app.SourcesResult {
	return u.settings.FetchFromSources(url)
}

// ---------------------------------------------------------------------------
// The Backup page (backup_seal.py).
// ---------------------------------------------------------------------------

// SaveBackupGroup commits the page's fields (one group = one merge).
func (u *PagesBUI) SaveBackupGroup(vals *app.BackupValues) *app.SaveResult {
	return u.settings.SaveBackupGroup(vals)
}

// BackupLadders answers the three groups' status dots off the LIVE
// values the frontend holds (the keystroke-refresh law).
type BackupLadders struct {
	Backup    app.StatusDot `json:"backup"`
	VaultSeal app.StatusDot `json:"vaultSeal"`
	GoodRepos app.StatusDot `json:"goodRepos"`
}

// RefreshBackupLadders computes the dots (the pure rungs) — the
// frontend passes its live field state (unsaved edits count).
func (u *PagesBUI) RefreshBackupLadders(folder string, backupEnabled,
	sealEnabled, sealPush, goodEnabled, goodPush bool) *BackupLadders {
	_ = sealPush
	_ = goodPush
	hasVault := u.settings.LiveVaultPath() != ""
	hasToken := u.settings.HasGitHubToken()
	return &BackupLadders{
		Backup:    app.BackupFolderStatus(folder, backupEnabled),
		VaultSeal: app.VaultSealStatus(sealEnabled, hasVault, hasToken),
		GoodRepos: app.GoodReposStatus(goodEnabled, hasVault, hasToken),
	}
}

// BackupNowGuards answers the Backup Now door's refusals ("" title =
// the door opens).
func (u *PagesBUI) BackupNowGuards() (title, body string) {
	return app.BackupNowGuards(u.settings.PlainConfig())
}

// BackupNow runs the vault zip (the BackupWorker port).
func (u *PagesBUI) BackupNow() *app.SealResultRow {
	return u.commands.BackupNow(context.Background())
}

// SealNow is the VaultSeal group's Seal Now button (kind repos |
// websites | personal).
func (u *PagesBUI) SealNow(kind string) *app.SealResultRow {
	return u.commands.BackupSealVault(context.Background(), kind)
}

// ResyncMirror is the warned Resync button's door (the confirm is the
// frontend's).
func (u *PagesBUI) ResyncMirror() *app.SealResultRow {
	return u.commands.BackupResync(context.Background())
}

// PublishNow is the GoodRepos group's Publish Now button.
func (u *PagesBUI) PublishNow() *app.SealResultRow {
	return u.commands.BackupPublish(context.Background())
}

// DetectRepos is THE REPO RADAR's manual trigger (the Backup page's
// Detect Repos button — read-only, the typed names never overwritten).
func (u *PagesBUI) DetectRepos() *app.RadarResult {
	return u.commands.RadarDetect(context.Background())
}

// FastSync is the ⚡ Fast Sync door (the triad's own mirror — no
// Telegram fetch, no LLM): the manual drift-seal with the
// ALWAYS-SPEAK + ALWAYS-PUBLISH laws. T-77's hero machine wires the
// main-window button here; the door itself is the app surface's own.
func (u *PagesBUI) FastSync() *app.SealResultRow {
	return u.commands.SyncFastSync(context.Background())
}

// OpenDashboard is the Dashboard group's Open Dashboard button: the
// URL law + the browser door. A refused title renders the themed
// dialog; "" = the browser opened.
func (u *PagesBUI) OpenDashboard(fieldURL string) (refusedTitle, refusedBody string) {
	return u.settings.OpenDashboardBrowser(fieldURL)
}

// itoa is the tiny int render (fmt without the import weight at the
// binding surface).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
