// settingsui.go — the Settings window's binding wrapper: the thin
// shell over internal/app's Settings surface (doc 04 §3's law — "cli,
// ui: thin shells over app.Commands"; the pages drive the app's
// command surface, this file only binds it to the wails window) plus
// the two native pickers the pages own (the import file picker + the
// vault folder pickers — wails runtime dialogs, the QFileDialog twins).
package main

import (
	"context"

	"github.com/assadigit/second-brain-project/internal/app"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// SettingsUI is the Settings window's bound surface (the wailsjs
// bindings land in frontend/wailsjs/go/main/SettingsUI.js).
type SettingsUI struct {
	settings *app.Settings
	ctx      context.Context
}

// NewSettingsUI wires the wrapper over the app's Settings surface.
func NewSettingsUI(settings *app.Settings) *SettingsUI {
	return &SettingsUI{settings: settings}
}

// bindCtx captures the wails context at startup (the dialogs' owner).
func (u *SettingsUI) bindCtx(ctx context.Context) {
	u.ctx = ctx
}

// ---------------------------------------------------------------------------
// The snapshot + the ladder + the save door.
// ---------------------------------------------------------------------------

// Snapshot builds every page's initial state (one call, the reference's
// initUI shape).
func (u *SettingsUI) Snapshot() *app.SettingsSnapshot {
	return u.settings.Snapshot()
}

// VaultStatus is one picker's live status row (the keystroke ladder).
func (u *SettingsUI) VaultStatus(path, kind string) *app.VaultStatus {
	st := app.VaultPathStatus(path, kind)
	return &st
}

// SaveSettings is THE ONE merge-based save (the save-on-change law's
// Go half: the page savers call this with their own group).
func (u *SettingsUI) SaveSettings(vals *app.SettingsValues) *app.SaveResult {
	return u.settings.SaveSettings(vals)
}

// RemoveVault takes a path off the vault combo (the confirm dialog is
// the frontend's; this door only edits config).
func (u *SettingsUI) RemoveVault(path string) *app.SaveResult {
	return u.settings.RemoveVault(path)
}

// ---------------------------------------------------------------------------
// Credentials page.
// ---------------------------------------------------------------------------

// TestGitHubToken validates the token AS TYPED (before any save) and
// speaks the verdict lines on the desk.
func (u *SettingsUI) TestGitHubToken(token string) *app.TokenTestResult {
	return u.settings.TestGitHubToken(token)
}

// GenerateAboutMe writes about_me.md from the wizard's answers.
func (u *SettingsUI) GenerateAboutMe(f app.AboutMeFields) *app.AboutMeResult {
	return u.settings.GenerateAboutMe(f)
}

// UserConfigStatusLine repaints the quiet-truth line (after saves).
func (u *SettingsUI) UserConfigStatusLine() string {
	return u.settings.UserConfigStatusLine()
}

// ShowUserConfigFile opens the OS file manager at the credentials file.
func (u *SettingsUI) ShowUserConfigFile() bool {
	return u.settings.ShowUserConfigFile()
}

// ---------------------------------------------------------------------------
// LLM page.
// ---------------------------------------------------------------------------

// RefreshOllamaModels pulls the model list from the Ollama server at
// url (current = the combo's live text, preserved when not listed).
func (u *SettingsUI) RefreshOllamaModels(url, current string) *app.OllamaModelsResult {
	return u.settings.RefreshOllamaModels(url, current)
}

// StartOllamaServer spawns `ollama serve` detached; the 3000 ms
// auto-refresh is the frontend's setTimeout (the singleShot twin).
func (u *SettingsUI) StartOllamaServer() *app.StartServerResult {
	return u.settings.StartOllamaServer()
}

// DetectLlamaCpp scans for a running llama-server (process ports first,
// then the common ports) and fills URL + model.
func (u *SettingsUI) DetectLlamaCpp(apiKey string) *app.LlamaDetectResult {
	return u.settings.DetectLlamaCpp(apiKey)
}

// RefreshLlamaCppModels probes the server at the field's URL.
func (u *SettingsUI) RefreshLlamaCppModels(url, apiKey string) *app.LlamaDetectResult {
	return u.settings.RefreshLlamaCppModels(url, apiKey)
}

// QuickDetect is the fast lane's probe half (the batch-running guard
// included); several models → the frontend's menu, then ApplyQuickDetect.
func (u *SettingsUI) QuickDetect(provider, liveURL, liveKey string) *app.QuickDetectResult {
	return u.settings.QuickDetect(provider, liveURL, liveKey)
}

// ApplyQuickDetect is the SET half: provider + URL + model land in place
// + merge-save. choice "" = the menu's Cancel. live = the frontend's
// current LLM field state (the snapshot law: unsaved edits count).
func (u *SettingsUI) ApplyQuickDetect(provider, base string, models []string,
	choice string, live *app.LLMValues) *app.ApplyQuickDetectResult {
	return u.settings.ApplyQuickDetect(provider, base, models, choice, live)
}

// ---------------------------------------------------------------------------
// Sound page.
// ---------------------------------------------------------------------------

// PreviewSound is the exact batch gate's decision half: the verdict +
// the wav payload for the <audio> twin (or the honest 🔇/🔕 lines).
func (u *SettingsUI) PreviewSound(needsRetry bool) *app.SoundPreviewResult {
	return u.settings.PreviewSound(needsRetry)
}

// ---------------------------------------------------------------------------
// The native pickers (the pages' QFileDialog twins).
// ---------------------------------------------------------------------------

// SelectImportFile is the Input page's picker (.txt AND .md — one URL
// per line). An empty answer (cancel) changes nothing; a pick sets the
// field + the 📄 line (select_import_file).
func (u *SettingsUI) SelectImportFile() string {
	path, _ := runtime.OpenFileDialog(u.ctx, runtime.OpenDialogOptions{
		Title: "Import txt file",
		Filters: []runtime.FileFilter{
			{DisplayName: "Text & Markdown (*.txt *.md)", Pattern: "*.txt;*.md"},
			{DisplayName: "Text Files (*.txt)", Pattern: "*.txt"},
			{DisplayName: "Markdown (*.md)", Pattern: "*.md"},
			{DisplayName: "All files (*)", Pattern: "*"},
		},
	})
	if path != "" {
		u.settings.EmitImportFileSelected(path)
	}
	return path
}

// BrowseVault is the folder picker; kind picks the reference's own
// dialog titles: github = "Select Obsidian Vault", websites = "Select
// Websites Vault (a folder that does not exist yet is fine)", manual =
// "Select Manual Notes Vault", backup = "Select Backup Folder"
// (T-74's Backup page).
func (u *SettingsUI) BrowseVault(kind string) string {
	title := "Select Obsidian Vault"
	switch kind {
	case "websites":
		title = "Select Websites Vault (a folder that does not exist yet is fine)"
	case "manual":
		title = "Select Manual Notes Vault"
	case "backup":
		title = "Select Backup Folder"
	}
	path, _ := runtime.OpenDirectoryDialog(u.ctx, runtime.OpenDialogOptions{
		Title: title,
	})
	return path
}
