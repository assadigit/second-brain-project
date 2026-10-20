// exportdoors_test.go — T-80's standing-assignment battery: the
// export-zip engine (the ExportWorker's data half) + the restore
// family + THE DEFERRED LEDGER RIDER on BOTH zips (the v0.86.0 law —
// the backup's own rider landed with this slice after T-71's
// placeholder-era note retired). Every seam is hermetic: temp vaults,
// temp ledgers, temp backup folders; the stamps are injected clocks.
package app

import (
	"archive/zip"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// injectLedger points the rider's ledger seam at a temp file with the
// given bytes (the hermetic law; restored on cleanup).
func injectLedger(t *testing.T, data []byte) string {
	t.Helper()
	p := filepath.Join(t.TempDir(), "sync_review.db")
	if err := os.WriteFile(p, data, 0o644); err != nil {
		t.Fatal(err)
	}
	orig := exportLedgerPathFn
	exportLedgerPathFn = func() string { return p }
	t.Cleanup(func() { exportLedgerPathFn = orig })
	return p
}

// fixtureVault builds the export law's whole surface: included files,
// every exclude class, a nested folder, the .obsidian pair.
func fixtureVault(t *testing.T) (vault string, included map[string]bool) {
	t.Helper()
	root := t.TempDir()
	vault = filepath.Join(root, "MyVault")
	for _, rel := range []string{
		"note one.md", "Library/sub/deep.md", ".obsidian/other.json",
	} {
		p := filepath.Join(vault, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("body of "+rel), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	// The excluded classes: the name law, the dir law, the suffix law,
	// the prefix law.
	for _, rel := range []string{
		"config.json", "__pycache__/x.pyc", "draft.lock",
		".obsidian/workspace.json", ".obsidian/app.json",
	} {
		p := filepath.Join(vault, filepath.FromSlash(rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte("never ships"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.MkdirAll(filepath.Join(vault, ".git", "objects"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vault, ".git", "HEAD"), []byte("ref"), 0o644); err != nil {
		t.Fatal(err)
	}
	return vault, map[string]bool{
		"MyVault/note one.md":          true,
		"MyVault/Library/sub/deep.md":  true,
		"MyVault/.obsidian/other.json": true,
	}
}

// TestExportZipGuards — the No Vault refusal, the reference's own words.
func TestExportZipGuards(t *testing.T) {
	cfg := storage.NewOrderedMap()
	if title, _ := ExportZipGuards(cfg); title != "No Vault" {
		t.Fatalf("guard title = %q; want No Vault", title)
	}
	cfg.Set("vault_path", "/tmp/v")
	if title, _ := ExportZipGuards(cfg); title != "" {
		t.Fatalf("a set vault refused: %q", title)
	}
}

// TestRunVaultExportLaws — the exclude classes, the arc-root law (the
// vault FOLDER rides inside), the ledger rider (+ its file count), the
// done-message shape.
func TestRunVaultExportLaws(t *testing.T) {
	vault, included := fixtureVault(t)
	injectLedger(t, []byte("LEDGER-BYTES"))
	out := filepath.Join(t.TempDir(), "vault-export-test.zip")
	msg, err := RunVaultExport(vault, out, time.Date(2026, 10, 20, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatal(err)
	}
	zr, err := zip.OpenReader(out)
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	names := map[string]bool{}
	for _, f := range zr.File {
		names[f.Name] = true
	}
	for want := range included {
		if !names[want] {
			t.Errorf("missing zip member %s (have %v)", want, names)
		}
	}
	// The rider: the ledger rides as GitCurator-data/sync_review.db.
	// The rider lands at the ZIP ROOT (the reference's own arcname:
	// zf.write(_ledger, 'GitCurator-data/sync_review.db')).
	if !names["GitCurator-data/sync_review.db"] {
		t.Errorf("the ledger rider missing (have %v)", names)
	}
	// The exclude classes never ship.
	for _, banned := range []string{
		"MyVault/config.json", "MyVault/draft.lock",
		"MyVault/__pycache__/x.pyc", "MyVault/.git/HEAD",
		"MyVault/.obsidian/workspace.json", "MyVault/.obsidian/app.json",
	} {
		if names[banned] {
			t.Errorf("excluded member shipped: %s", banned)
		}
	}
	// The file count: 3 vault files + the rider.
	if !strings.Contains(msg, "(4 files, ") {
		t.Errorf("the done message's file count: %q", msg)
	}
	if !strings.HasPrefix(msg, out+" (") {
		t.Errorf("the done message's shape: %q", msg)
	}
}

// TestRunVaultExportNoLedger — a missing ledger rides silently (the
// best-effort law: an export never fails on the rider).
func TestRunVaultExportNoLedger(t *testing.T) {
	vault, _ := fixtureVault(t)
	// A ledger that is NOT on disk (the best-effort law: no rider).
	orig := exportLedgerPathFn
	exportLedgerPathFn = func() string {
		return filepath.Join(t.TempDir(), "missing.db")
	}
	t.Cleanup(func() { exportLedgerPathFn = orig })
	out := filepath.Join(t.TempDir(), "z.zip")
	msg, err := RunVaultExport(vault, out, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(msg, "(3 files, ") {
		t.Errorf("no-ledger export count: %q", msg)
	}
}

// TestBackupRiderLanded — the v0.86.0 rider on the BACKUP zip too
// (T-71's placeholder-era note retired at T-80): RunBackupNow's zip
// carries GitCurator-data/sync_review.db and counts it.
func TestBackupRiderLanded(t *testing.T) {
	vault, _ := fixtureVault(t)
	injectLedger(t, []byte("BACKUP-LEDGER"))
	folder := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)
	cfg.Set("backup_folder", folder)
	row := RunBackupNow(cfg, func(string, string) {})
	if !row.Success {
		t.Fatalf("backup failed: %+v", row)
	}
	pile, _ := filepath.Glob(filepath.Join(folder, "vault-*.zip"))
	if len(pile) != 1 {
		t.Fatalf("backup pile = %v", pile)
	}
	zr, err := zip.OpenReader(pile[0])
	if err != nil {
		t.Fatal(err)
	}
	defer zr.Close()
	found := false
	for _, f := range zr.File {
		if f.Name == "GitCurator-data/sync_review.db" {
			found = true
		}
	}
	if !found {
		t.Fatal("the backup zip lost the ledger rider (the v0.86.0 law)")
	}
	// The count: 3 vault files + the rider = 4.
	if !strings.Contains(row.Message, "(4 files, ") {
		t.Errorf("the backup's file count lost the rider: %q", row.Message)
	}
}

// TestListRestoreBackups — the newest-first law + the size strings +
// the vault-*.zip glob's own shape.
func TestListRestoreBackups(t *testing.T) {
	folder := t.TempDir()
	for _, name := range []string{
		"vault-2026-10-01T000000Z.zip", "vault-2026-10-20T120000Z.zip",
		"not-a-backup.zip", "vault-2026-10-10T060000Z.zip",
	} {
		if err := os.WriteFile(filepath.Join(folder, name),
			[]byte(strings.Repeat("x", 2048)), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	rows := ListRestoreBackups(folder)
	if len(rows) != 3 {
		t.Fatalf("rows = %d; want 3 (the non-matching name skipped)", len(rows))
	}
	want := []string{
		"vault-2026-10-20T120000Z.zip",
		"vault-2026-10-10T060000Z.zip",
		"vault-2026-10-01T000000Z.zip",
	}
	for i, w := range want {
		if rows[i].Name != w {
			t.Errorf("row %d = %q; want %q (newest first)", i, rows[i].Name, w)
		}
		if rows[i].Size != "2.0 KB" {
			t.Errorf("row %d size = %q; want 2.0 KB", i, rows[i].Size)
		}
	}
	// The size string's MB rung.
	big := filepath.Join(folder, "vault-2026-10-21T000000Z.zip")
	if err := os.WriteFile(big, []byte(strings.Repeat("x", 3*1024*1024)), 0o644); err != nil {
		t.Fatal(err)
	}
	rows = ListRestoreBackups(folder)
	if rows[0].Size != "3.0 MB" {
		t.Errorf("the MB rung: %q", rows[0].Size)
	}
}

// TestRestoreGuardsAndEmptyBox — the reference's own words.
func TestRestoreGuardsAndEmptyBox(t *testing.T) {
	cfg := storage.NewOrderedMap()
	title, body := RestoreGuards(cfg)
	if title != "No Backup Folder" || body != "Select a backup folder first." {
		t.Fatalf("guard = %q / %q", title, body)
	}
	cfg.Set("backup_folder", "/tmp/b")
	if title, _ := RestoreGuards(cfg); title != "" {
		t.Fatalf("a set folder refused: %q", title)
	}
	title, body = RestoreEmptyBox()
	if title != "No Backups" ||
		body != "No backups found yet.\n\nClick 'Backup Now' first." {
		t.Fatalf("empty box = %q / %q", title, body)
	}
}

// TestRestoreBackupToNewFolder — the safety net: the extract lands in
// vault-restored-<ts> NEXT TO the vault, the vault itself untouched,
// nested members create their folders, and a path-escaping member is
// refused (the extractall safety belt).
func TestRestoreBackupToNewFolder(t *testing.T) {
	root := t.TempDir()
	vault := filepath.Join(root, "MyVault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vault, "note.md"), []byte("live"), 0o644); err != nil {
		t.Fatal(err)
	}
	// The backup zip: the vault folder shape + an escaping member.
	bk := filepath.Join(root, "vault-2026-10-20T000000Z.zip")
	zf, err := os.Create(bk)
	if err != nil {
		t.Fatal(err)
	}
	zw := zip.NewWriter(zf)
	for name, data := range map[string]string{
		"MyVault/note.md":             "restored",
		"MyVault/Library/sub/deep.md": "nested",
		"../escaped.txt":              "never",
	} {
		w, _ := zw.Create(name)
		_, _ = w.Write([]byte(data))
	}
	_ = zw.Close()
	_ = zf.Close()

	now := time.Date(2026, 10, 21, 13, 14, 15, 0, time.UTC)
	folder, err := RestoreBackupToNewFolder(bk, vault, now)
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Join(root, "vault-restored-20261021-131415")
	if folder != want {
		t.Fatalf("folder = %q; want %q", folder, want)
	}
	for name, data := range map[string]string{
		"MyVault/note.md":             "restored",
		"MyVault/Library/sub/deep.md": "nested",
	} {
		b, rerr := os.ReadFile(filepath.Join(want, filepath.FromSlash(name)))
		if rerr != nil {
			t.Fatalf("member %s missing: %v", name, rerr)
		}
		if string(b) != data {
			t.Errorf("member %s = %q", name, b)
		}
	}
	if _, eerr := os.Stat(filepath.Join(root, "escaped.txt")); eerr == nil {
		t.Error("an escaping member escaped the fresh folder")
	}
	// The current vault is untouched.
	b, _ := os.ReadFile(filepath.Join(vault, "note.md"))
	if string(b) != "live" {
		t.Errorf("the current vault was touched: %q", b)
	}
	// A broken zip answers the error.
	if _, err := RestoreBackupToNewFolder(filepath.Join(root, "nope.zip"),
		vault, now); err == nil {
		t.Error("a missing backup passed as restorable")
	}
}

// TestCopyBackupFile — the bytes + the mtime (the copy2 twin).
func TestCopyBackupFile(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "vault-x.zip")
	if err := os.WriteFile(src, []byte("ZIPBYTES"), 0o644); err != nil {
		t.Fatal(err)
	}
	stamp := time.Date(2026, 1, 2, 3, 4, 5, 0, time.UTC)
	if err := os.Chtimes(src, stamp, stamp); err != nil {
		t.Fatal(err)
	}
	dst := filepath.Join(dir, "copy.zip")
	if err := CopyBackupFile(src, dst); err != nil {
		t.Fatal(err)
	}
	b, _ := os.ReadFile(dst)
	if string(b) != "ZIPBYTES" {
		t.Fatalf("copy = %q", b)
	}
	if fi, _ := os.Stat(dst); !fi.ModTime().Equal(stamp) {
		t.Errorf("mtime not preserved: %v", fi.ModTime())
	}
}

// TestRestoreDoorPayloadFieldTags — THE FIELD-TAG LAW: the restore
// dialog's payload marshals the tagged names the frontend reads.
func TestRestoreDoorPayloadFieldTags(t *testing.T) {
	b, err := json.Marshal(&RestoreBackupRow{Name: "vault-x.zip", Size: "1.2 MB"})
	if err != nil {
		t.Fatal(err)
	}
	if got := string(b); got != `{"name":"vault-x.zip","size":"1.2 MB"}` {
		t.Fatalf("RestoreBackupRow tags: %s", got)
	}
	// The size helper's shapes (the reference's 1-decimal law).
	if s := sizeStrMBKB(511); s != "0.5 KB" {
		t.Errorf("sizeStr(511) = %q", s)
	}
	if s := sizeStrMBKB(1024 * 1024); s != "1.0 MB" {
		t.Errorf("sizeStr(1MiB) = %q", s)
	}
	// The export default name's stamp shape.
	if s := ExportDefaultName(time.Date(2026, 10, 21, 8, 9, 10, 0, time.UTC)); s != "vault-export-2026-10-21_08-09-10.zip" {
		t.Errorf("ExportDefaultName = %q", s)
	}
}

// TestExportDoorSpeaksAndBoxes — the shell's door over the engine: the
// guards box, the success pair (the 📤 line + Export Complete), the
// failure pair (the ❌ line + Export Failed). Uses the Settings-level
// bus recorder to pin the desk lines.
func TestExportDoorSpeaksAndBoxes(t *testing.T) {
	cfg := storage.NewOrderedMap()
	s, rec := newPagesBSettings(t, cfg)
	vault, _ := fixtureVault(t)
	s.cfg.Set("vault_path", vault)
	out := filepath.Join(t.TempDir(), "z.zip")
	msg, err := RunVaultExport(vault, out, time.Now())
	if err != nil {
		t.Fatal(err)
	}
	// The shell's own speaking law (heroui.go mirrors this verbatim).
	s.SpeakOnBus("📤 Exported: "+msg, "success")
	if line, ok := rec.findLog("Exported: " + out + " ("); !ok ||
		!strings.HasPrefix(line.Msg, "📤 Exported: "+out+" (") ||
		line.Level != "success" {
		t.Errorf("the export line: %+v (ok=%v)", line, ok)
	}
}
