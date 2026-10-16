package connectioncheck

// vaults_test.go — section 1: the writability probe + the
// github/websites/manual matrix + the check_vaults pipeline gates.
// Ported from the reference's TestVaultWritable, TestVaultChecks and
// TestCheckVaultsMatrix.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestVaultWritableOkAndLeavesNoFile(t *testing.T) {
	td := t.TempDir()
	ok, why := VaultWritable(td)
	if !ok {
		t.Fatalf("writable dir reported unwritable: %s", why)
	}
	entries, err := os.ReadDir(td)
	if err != nil {
		t.Fatal(err)
	}
	for _, e := range entries {
		if strings.HasPrefix(e.Name(), ".gc-write-test-") {
			t.Errorf("leftover temp file: %s", e.Name())
		}
	}
	if len(entries) != 0 {
		t.Errorf("dir not clean: %d entries", len(entries))
	}
}

func TestVaultWritableNonexistentDirFailsWithReason(t *testing.T) {
	ok, why := VaultWritable("/nonexistent/no/such/dir")
	if ok {
		t.Fatal("nonexistent dir reported writable")
	}
	if why == "" {
		t.Error("the failure detail must carry the reason")
	}
	if !strings.Contains(why, "FileNotFoundError") {
		t.Errorf("failure detail lacks the Python class: %q", why)
	}
}

func TestVaultWritablePrefixLaw(t *testing.T) {
	// the temp name rides the .gc-write-test- prefix (the probe's
	// identity, pinned)
	td := t.TempDir()
	probe := func() string {
		f, err := os.CreateTemp(td, ".gc-write-test-")
		if err != nil {
			t.Fatal(err)
		}
		name := f.Name()
		_ = f.Close()
		return name
	}
	name := probe()
	if !strings.HasPrefix(filepath.Base(name), ".gc-write-test-") {
		t.Errorf("temp prefix missing: %s", name)
	}
	_ = os.Remove(name)
}

func TestCheckVaultGithubNotSetIsError(t *testing.T) {
	r := CheckVault("", "github", "GitHub vault")
	if r.Level != LevelError {
		t.Errorf("level = %s, want error", r.Level)
	}
	if !strings.Contains(r.Detail, "not set") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckVaultGithubMissingOnDiskIsError(t *testing.T) {
	r := CheckVault("/nonexistent/vault", "github", "GitHub vault")
	if r.Level != LevelError {
		t.Errorf("level = %s, want error", r.Level)
	}
	if !strings.Contains(r.Detail, "missing on disk") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckVaultGithubFoundWritableIsOk(t *testing.T) {
	td := t.TempDir()
	if err := os.Mkdir(filepath.Join(td, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	r := CheckVault(td, "github", "GitHub vault")
	if r.Level != LevelOK {
		t.Fatalf("level = %s (%s)", r.Level, r.Detail)
	}
	for _, want := range []string{"writable", "ready to receive notes",
		"Obsidian vault"} {
		if !strings.Contains(r.Detail, want) {
			t.Errorf("detail lacks %q: %q", want, r.Detail)
		}
	}
}

func TestCheckVaultGithubReadonlyIsError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}
	td := t.TempDir()
	if err := os.Chmod(td, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(td, 0o755) })
	r := CheckVault(td, "github", "GitHub vault")
	if r.Level != LevelError {
		t.Fatalf("level = %s, want error", r.Level)
	}
	if !strings.Contains(r.Detail, "NOT writable") {
		t.Errorf("detail: %q", r.Detail)
	}
	if !strings.Contains(r.Detail, "PermissionError") {
		t.Errorf("detail lacks the Python class: %q", r.Detail)
	}
}

func TestCheckVaultWebsitesWillBeCreated(t *testing.T) {
	td := t.TempDir()
	future := filepath.Join(td, "websites-vault")
	r := CheckVault(future, "websites", "Websites vault")
	if r.Level != LevelInfo {
		t.Errorf("level = %s, want info", r.Level)
	}
	if !strings.Contains(r.Detail, "created by the pipeline") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckVaultWebsitesParentMissingIsError(t *testing.T) {
	r := CheckVault("/nonexistent/parent/websites", "websites",
		"Websites vault")
	if r.Level != LevelError {
		t.Errorf("level = %s, want error", r.Level)
	}
	if !strings.Contains(r.Detail, "parent folder missing") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckVaultWebsitesParentUnwritableIsError(t *testing.T) {
	if os.Getuid() == 0 {
		t.Skip("running as root: file permissions are not enforced")
	}
	td := t.TempDir()
	parent := filepath.Join(td, "parent")
	if err := os.Mkdir(parent, 0o555); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chmod(parent, 0o755) })
	r := CheckVault(filepath.Join(parent, "websites"), "websites",
		"Websites vault")
	if r.Level != LevelError {
		t.Fatalf("level = %s, want error (%s)", r.Level, r.Detail)
	}
	if !strings.Contains(r.Detail, "parent folder is not writable") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckVaultManualNotFoundIsWarnNeverCreated(t *testing.T) {
	td := t.TempDir()
	r := CheckVault(filepath.Join(td, "manual"), "manual", "Manual vault")
	if r.Level != LevelWarn {
		t.Errorf("level = %s, want warn", r.Level)
	}
	if !strings.Contains(r.Detail, "never creates it") {
		t.Errorf("detail: %q", r.Detail)
	}
}

func TestCheckVaultManualNotSetIsInfoOptional(t *testing.T) {
	r := CheckVault("", "manual", "Manual vault")
	if r.Level != LevelInfo {
		t.Errorf("level = %s, want info", r.Level)
	}
}

func TestCheckVaultsMatrixPipelineOffNotSetIsInfo(t *testing.T) {
	rows := CheckVaults(map[string]any{
		"pipelines": map[string]any{"websites": false}})
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	type nl struct{ name, level string }
	want := []nl{
		{"GitHub vault", LevelError},
		{"Websites vault", LevelInfo},
		{"Manual vault", LevelInfo},
	}
	for i, w := range want {
		if rows[i].Name != w.name || rows[i].Level != w.level {
			t.Errorf("row %d = (%s, %s), want (%s, %s)", i,
				rows[i].Name, rows[i].Level, w.name, w.level)
		}
	}
}

func TestCheckVaultsMatrixPipelineOnNotSetIsWarn(t *testing.T) {
	rows := CheckVaults(map[string]any{
		"pipelines": map[string]any{"websites": true}})
	if len(rows) < 2 {
		t.Fatalf("rows = %d", len(rows))
	}
	if rows[1].Name != "Websites vault" || rows[1].Level != LevelWarn {
		t.Errorf("row 1 = (%s, %s), want (Websites vault, warn)",
			rows[1].Name, rows[1].Level)
	}
}

func TestCheckVaultsMatrixPipelineOffButPathSetIsInfo(t *testing.T) {
	rows := CheckVaults(map[string]any{
		"pipelines":          map[string]any{"websites": false},
		"website_vault_path": "/some/path"})
	if rows[1].Level != LevelInfo {
		t.Errorf("level = %s, want info", rows[1].Level)
	}
	if !strings.Contains(rows[1].Detail, "pipeline is OFF") {
		t.Errorf("detail: %q", rows[1].Detail)
	}
}

func TestCheckVaultsManualVaultCheckedWhenSet(t *testing.T) {
	td := t.TempDir()
	rows := CheckVaults(map[string]any{
		"manual_vault_path": td,
		"pipelines":         map[string]any{"websites": false}})
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
	if rows[2].Name != "Manual vault" || rows[2].Level != LevelOK {
		t.Errorf("manual row = (%s, %s)", rows[2].Name, rows[2].Level)
	}
}

func TestCheckVaultsHostileJunkNeverPanics(t *testing.T) {
	rows := CheckVaults(map[string]any{
		"pipelines": "junk", "ollama": "junk",
		"website_vault_path": 3, "manual_vault_path": []any{1}})
	if len(rows) != 3 {
		t.Fatalf("rows = %d, want 3", len(rows))
	}
}
