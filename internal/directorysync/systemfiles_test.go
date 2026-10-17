// systemfiles_test.go — T-49: the ported selections of the reference's
// TestSystemFiles (test_directorysync.py) — the trio creation law
// (create once, never overwrite) and the shipped config's private
// visibility marks.
package directorysync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestSystemFiles_CreatedOnceNeverOverwritten(t *testing.T) {
	vault := t.TempDir()
	created := EnsureSystemFiles(vault, nil)
	if len(created) != 3 {
		t.Fatalf("created: %v", created)
	}
	// The owner edits the vault copy by hand…
	banishedPath := filepath.Join(vault, "_system", "banished.json")
	if err := os.WriteFile(banishedPath, []byte(`[{"host": "edited.example"}]`), 0o644); err != nil {
		t.Fatal(err)
	}
	again := EnsureSystemFiles(vault, nil)
	if len(again) != 0 {
		t.Fatalf("created again: %v", again)
	}
	data, err := os.ReadFile(banishedPath)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), "edited.example") {
		t.Fatalf("the owner's edit was overwritten: %s", data)
	}
}

func TestSystemFiles_ShippedConfigLoadsAndMarksPrivate(t *testing.T) {
	vault := t.TempDir()
	EnsureSystemFiles(vault, nil)
	cfgGet, _ := dictOf(LoadDirectoryConfig(vault))
	names := map[string]bool{}
	for _, cat := range dictList(cfgGet("categories")) {
		names[pyStrOrEmpty(cat("name"))] = true
	}
	if !names["Design"] || !names["Download Resources"] {
		t.Fatalf("category names: %v", names)
	}
	var shadowVisibility string
	for _, cat := range dictList(cfgGet("categories")) {
		if pyStrOrEmpty(cat("name")) != "Download Resources" {
			continue
		}
		for _, sub := range dictList(cat("subcategories")) {
			if pyStrOrEmpty(sub("name")) == "Torrent & Shadow Libraries" {
				shadowVisibility = pyStrOrEmpty(sub("visibility"))
			}
		}
	}
	if shadowVisibility != "private" {
		t.Fatalf("Torrent & Shadow Libraries visibility: %q", shadowVisibility)
	}
}
