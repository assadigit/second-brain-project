// settingsuib_test.go — T-74's pages B binding wrapper wiring smoke:
// the bound surface delegates to internal/app's engines + Commands
// with the same shapes the generated bindings carry (the engine laws
// live in internal/app's batteries; this is the shell-side proof that
// the pages' doors are real).
package main

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/webui/uitheme"
)

// newPagesBUI builds a hermetic wrapper over fakes (the queue poll +
// the batch seam answer canned data — the hermetic law at the shell).
func newPagesBUI(t *testing.T) (*PagesBUI, *storage.OrderedMap, *[]string) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	cfg.Set("website_vault_path", t.TempDir())
	pipes := storage.NewOrderedMap()
	pipes.Set("github", true)
	pipes.Set("websites", true)
	cfg.Set("pipelines", pipes)
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{})
	var spoken []string
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		spoken = append(spoken, l.Msg)
	}})
	// Prime the settlement (the one-time law: a FRESH cache would
	// settle the whole queue history — the golden harness's own step).
	{
		cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
		if err != nil {
			t.Fatalf("prime store: %v", err)
		}
		if _, err := cache.SettleExistingRepos(nil); err != nil {
			t.Fatal(err)
		}
		_ = cache.Close()
	}

	commands := app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) {
			return "ok", []any{
				map[string]any{"url": "https://github.com/octo/pending"},
				map[string]any{"url": "https://example.com/page"},
			}
		},
		QueueClassify: app.ClassifyQueueLinks,
		Batch: func(ctx context.Context, cfg *storage.OrderedMap, opts app.BatchOptions, ev app.BatchEvents) *app.BatchResult {
			return &app.BatchResult{Success: true, Message: "the canned batch ran",
				Processed: len(opts.URLs), Total: len(opts.URLs) + len(opts.NonGithub)}
		},
	})
	return NewPagesBUI(settings, commands), cfg, &spoken
}

func TestPagesBUIQueueCheck(t *testing.T) {
	ui, _, spoken := newPagesBUI(t)

	// The username guard refuses honestly.
	res := ui.CheckBotQueue("  ")
	if res.Refusal != "❌ Please enter the bot username first." {
		t.Fatalf("refusal = %q", res.Refusal)
	}

	// The happy path: the render + the badge + the payload.
	res = ui.CheckBotQueue("@mybot")
	if res.Refusal != "" || res.Error != "" {
		t.Fatalf("res = %+v", res)
	}
	if res.Badge == nil || res.Badge.Text != "⏳ 1 pending" || !res.Badge.Pending {
		t.Fatalf("badge = %+v", res.Badge)
	}
	if len(res.PendingURLs) != 1 || res.PendingURLs[0] != "https://github.com/octo/pending" {
		t.Fatalf("pending = %v", res.PendingURLs)
	}
	// The queue-history settlement (v0.64.1) settles the website link
	// at the queue's own door — the report's honest line carries it.
	if !contains(res.Report, "🤝 Websites settled:        1") {
		t.Fatalf("the settled websites line never rode:\n%s", res.Report)
	}
	for _, want := range []string{
		"📬 Checking bot queue (@mybot)...",
		"📬 Queue: 1 new repo(s) + 0 website link(s) pending",
	} {
		found := false
		for _, l := range *spoken {
			if len(l) >= len(want) && l[:len(want)] == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("missing line %q in %v", want, *spoken)
		}
	}
	// The rendered report carries the reference's line names.
	for _, want := range []string{
		"📬 Bot Queue Results",
		"Total GitHub URLs in bot:  1",
		"⏳ Pending (not in vault):  1",
		"🔗 Non-GitHub links:        1",
	} {
		if !contains(res.Report, want) {
			t.Fatalf("the report is missing %q:\n%s", want, res.Report)
		}
	}
	// The username saved through the door (the reference's pre-fetch save).
	if v, _ := ui.settings.PlainConfig().Get("bot_username"); v != "mybot" {
		t.Fatalf("bot_username = %v", v)
	}
}

func TestPagesBUIProcessQueue(t *testing.T) {
	ui, _, _ := newPagesBUI(t)
	res := ui.ProcessBotQueue([]string{"https://github.com/octo/pending"},
		[]string{"https://example.com/page"}, 0, 2)
	if !res.Success || res.Processed != 1 || res.Total != 2 {
		t.Fatalf("res = %+v", res)
	}
}

func TestPagesBUIDashboardDoors(t *testing.T) {
	ui, cfg, _ := newPagesBUI(t)

	// No vault → the honest panel text.
	old := ui.settings
	_ = old
	cfg.Set("vault_path", "")
	res := ui.RefreshDashboard()
	if !res.NoVault || res.Text != "No vault selected." {
		t.Fatalf("res = %+v", res)
	}

	// A vault answers the scan + the render (the quarantine view rides
	// along — the empty answer over a fresh cache).
	cfg.Set("vault_path", t.TempDir())
	res = ui.RefreshDashboard()
	if res.NoVault || !contains(res.Text, "📊 VAULT DASHBOARD") ||
		!contains(res.Text, "Total notes: 0") {
		t.Fatalf("res = %+v (%s)", res, res.Text)
	}
	if !contains(res.Quarantine, "✅ No 404 attempts on record.") {
		t.Fatalf("quarantine = %q", res.Quarantine)
	}

	// The threshold save door.
	if r := ui.SaveQuarantineGroup(&app.QuarantineValues{StrikeThreshold: 5}); !r.Saved {
		t.Fatal("the threshold save failed")
	}
	if v, _ := cfg.Get("notfound_strike_threshold"); v != 5 {
		t.Fatalf("threshold = %v", v)
	}
}

func TestPagesBUIBackupDoors(t *testing.T) {
	ui, cfg, _ := newPagesBUI(t)

	// The ladders door: the disabled rungs over the live config.
	ladders := ui.RefreshBackupLadders("", false, true, true, true, true)
	if ladders.Backup.Text != "● Disabled" {
		t.Fatalf("backup dot = %+v", ladders.Backup)
	}
	if ladders.VaultSeal.Text != "● Local-only (no GitHub token — commits, no push)" {
		t.Fatalf("seal dot = %+v", ladders.VaultSeal)
	}
	cfg.Set("github_token", "gt")
	ladders = ui.RefreshBackupLadders(t.TempDir(), true, true, true, true, true)
	if ladders.Backup.Text != "● Ready" || ladders.VaultSeal.Text != "● Ready — auto-seal after every run" {
		t.Fatalf("ladders = %+v %+v", ladders.Backup, ladders.VaultSeal)
	}

	// The guards door (the themed dialog's facts).
	title, body := ui.BackupNowGuards()
	if title != "No Backup Folder" || body != "Select a backup folder first." {
		t.Fatalf("guards = %q %q", title, body)
	}
	cfg.Set("backup_folder", t.TempDir())
	if title, _ = ui.BackupNowGuards(); title != "" {
		t.Fatalf("guards should open: %q", title)
	}

	// The save group door (one commit = one merge).
	if r := ui.SaveBackupGroup(&app.BackupValues{BackupFolder: "/b",
		BackupMax: "7", WorkerURL: "https://w.dev/",
		VaultSealRepo: "mirror"}); !r.Saved {
		t.Fatal("the backup save failed")
	}
	if v, _ := cfg.Get("backup_max"); v != 7 {
		t.Fatalf("backup_max = %v", v)
	}

	// The undo door's early answer (no undo file in a fresh vault).
	if res := ui.ReadUndoLastBatch(); res.Stage != "no-file" {
		t.Fatalf("undo stage = %q", res.Stage)
	}
}

func TestPagesBUISectionGlyphsInPack(t *testing.T) {
	// The pages B's sidebar glyphs are the icon pack's own (the T-72
	// golden locked the pack; the new rows render from it).
	for _, g := range []string{"bar-chart-3", "bot", "satellite-dish",
		"archive", "refresh"} {
		if _, ok := uitheme.Icons[g]; !ok {
			t.Fatalf("the pages B need glyph %q — missing from the pack", g)
		}
	}
}

func contains(hay, needle string) bool {
	return len(hay) >= len(needle) && (func() bool {
		for i := 0; i+len(needle) <= len(hay); i++ {
			if hay[i:i+len(needle)] == needle {
				return true
			}
		}
		return false
	})()
}

// TestSettingsRestoreButtonLive — T-80's standing assignment landed:
// the Backup page's Restore button is a live door (the honestly-
// disabled note gone), the three binding imports present (THE
// BINDING-IMPORT LAW), and the restore dialog's own laws: the save-
// first order, the guard boxes, the selectable list, the three buttons
// with the reference's own labels, the ⚠️ confirm's exact body, and
// the extract result's words.
func TestSettingsRestoreButtonLive(t *testing.T) {
	js, err := readFrontend("settings.js")
	if err != nil {
		t.Fatal(err)
	}
	// The binding imports.
	for _, want := range []string{
		"RestoreBackupsList, RestoreBackupRun, CopyBackupTo",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("settings.js missing the restore imports (%q)", want)
		}
	}
	// The button is live: the disabled shape + the landing note gone.
	for _, gone := range []string{
		"restore.disabled = true;",
		"lands with the packaging phase",
	} {
		if strings.Contains(js, gone) {
			t.Fatalf("the Restore button's honest-disabled note still present (%q)", gone)
		}
	}
	// The flow's laws.
	for _, want := range []string{
		"() => restoreFlow())",
		"saveBackupFields().then(() => {",
		"RestoreBackupsList().then((m) => {",
		"if (m.boxTitle) {",
		"openRestoreDialog(m.backups)",
		"function openRestoreDialog(backups) {",
		"'Found ' + backups.length + ' backup(s):'",
		"b.name + '    [ ' + b.size + ' ]'",
		"miniCustom('Restore from Backup', body,",
		"['Restore to New Folder', 'primary'], ['Copy ZIP to...', 'secondary'], ['Close', 'secondary']",
		"if (selected < 0) return;",
		"miniDialog('⚠️ Restore',",
		"'This will extract to a NEW folder next to your current vault\\n' +",
		"'(current vault is NOT touched). Continue?'",
		"RestoreBackupRun(pick).then((res) => {",
		"CopyBackupTo(pick).then((res) => {",
	} {
		if !strings.Contains(js, want) {
			t.Fatalf("the restore dialog's law missing %q", want)
		}
	}
	// The scrolling list lives INSIDE the dialog body; the three
	// action buttons ride miniCustom's button row OUTSIDE it (the
	// modal law).
	if !strings.Contains(js, "list.style.maxHeight = '260px';") {
		t.Fatal("the restore list's scroll cap missing")
	}
}

// TestSettingsRestoreDoors — the shell doors' own shapes: the list
// payload (the guards, the empty box, the rows), the run doors' guards
// (the path-safety belt), and the door results' box titles.
func TestSettingsRestoreDoors(t *testing.T) {
	src := readFileOrFatal(t, "settingsuib.go")
	for _, want := range []string{
		"func (u *PagesBUI) RestoreBackupsList() *RestoreDialogModel",
		"func (u *PagesBUI) RestoreBackupRun(name string) *RestoreRunResult",
		"func (u *PagesBUI) CopyBackupTo(name string) *RestoreRunResult",
		`app.RestoreGuards(u.settings.PlainConfig())`,
		`app.ListRestoreBackups(pyCfgStr(u.settings.PlainConfig(), "backup_folder"))`,
		`strings.ContainsAny(name, "/\\")`,
		`app.RestoreBackupToNewFolder(`,
		`"✅ Restored to: "+newFolder`,
		`Title: "Restore Complete"`,
		`"Restored to:\n" + newFolder + "\n\nReview and merge manually."`,
		`runtime.SaveFileDialog(u.ctx, runtime.SaveDialogOptions{`,
		`Title:           "Copy Backup"`,
		`app.CopyBackupFile(filepath.Join(folder, name), dst)`,
	} {
		if !strings.Contains(src, want) {
			t.Fatalf("the restore doors' law missing %q", want)
		}
	}
}

// TestRestoreDialogModelFieldTags — THE FIELD-TAG LAW: the dialog's
// payload marshals the tagged names the frontend reads.
func TestRestoreDialogModelFieldTags(t *testing.T) {
	b, err := json.Marshal(&RestoreDialogModel{
		BoxTitle: "No Backups",
		BoxBody:  "b",
		Backups:  []app.RestoreBackupRow{{Name: "vault-x.zip", Size: "1.0 KB"}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		`"boxTitle":"No Backups"`,
		`"boxBody":"b"`,
		`"backups":[{"name":"vault-x.zip","size":"1.0 KB"}]`,
	} {
		if !strings.Contains(string(b), want) {
			t.Fatalf("RestoreDialogModel tags: %s missing %s", b, want)
		}
	}
}
