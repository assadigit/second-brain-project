// settingsui_test.go — the Settings binding wrapper's wiring smoke:
// the wails-bound surface delegates to internal/app's Settings with
// the same shapes the generated bindings carry (the page workflow
// tests' engine laws live in internal/app/settings_test.go; this is
// the shell-side proof that the pages' door is real).
package main

import (
	"testing"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/webui/uitheme"
)

func TestSettingsUIWiringSmoke(t *testing.T) {
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	cfg.Set("github_token", "gt")
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{})
	ui := NewSettingsUI(settings)

	// The snapshot door: the six pages' payload builds through the
	// wrapper exactly as the frontend receives it.
	snap := ui.Snapshot()
	if snap == nil {
		t.Fatal("Snapshot returned nil")
	}
	if snap.GitHubToken != "gt" {
		t.Fatalf("github_token = %q, want gt", snap.GitHubToken)
	}
	if len(snap.BlockedLawDomains) == 0 {
		t.Fatal("the law list did not ride the snapshot")
	}

	// The ladder door (the keystroke refresh the Vault page drives).
	st := ui.VaultStatus("  ", "github")
	if st.Text != "● not set" || st.State != "warning" {
		t.Fatalf("ladder = %+v", st)
	}

	// The save door: the merge save lands through the wrapper (the
	// save-on-change law's transport leg).
	res := ui.SaveSettings(&app.SettingsValues{
		Credentials: &app.CredentialsValues{GitHubToken: "gt2"}})
	if !res.Saved {
		t.Fatal("the wrapper save failed (no Write seam = honest refusal)")
	}
	if v, _ := cfg.Get("github_token"); v != "gt2" {
		t.Fatalf("github_token = %v, want gt2", v)
	}

	// The import-file line door speaks on the bus.
	n := 0
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) { n++ }})
	ui.settings.EmitImportFileSelected("/tmp/links.txt")
	if n == 0 {
		t.Fatal("the import-file line never spoke")
	}
}

func TestSettingsUISectionGlyphsInPack(t *testing.T) {
	// The sidebar's section glyphs are the icon pack's own (the T-72
	// golden locked the pack; the Settings rows render from it).
	for _, g := range []string{"key-round", "globe", "folder", "brain",
		"inbox", "volume-2", "settings"} {
		if _, ok := uitheme.Icons[g]; !ok {
			t.Fatalf("the Settings UI needs glyph %q — missing from the pack", g)
		}
	}
}
