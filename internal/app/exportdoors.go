// exportdoors.go — T-80's standing assignment (spec-gui-pages-b's dated
// note): the export-zip engine (backup_seal.py:1274-1344 —
// _backup_export_zip's ExportWorker data half) + the safety-net restore
// family (backup_seal.py:1356-1441 — _backup_restore's list /
// extract-to-new-folder / copy doors) + THE DEFERRED LEDGER RIDER
// (v0.86.0 — the sync-review log rides EVERY vault zip so the review
// memory survives a machine loss). The rider landed here on BOTH zips:
// T-71's backupjob.go had recorded its absence against the then-
// placeholder syncreview seat; the ledger is real since the Phase-3/4
// landings, so the rider gains its bytes now (reality wins — the stale
// note retires in this slice).
//
// The engine's laws (the reference's own, byte-shaped where pinned):
//   - the export's exclude set is the backup's own (config.json, .git,
//     __pycache__ pruned as dirs AND skipped as names; *.lock skipped;
//     .obsidian/workspace + .obsidian/app.json skipped by the path
//     substring);
//   - the arc root is the vault's PARENT (the zip carries the vault
//     FOLDER, not its contents loose);
//   - the ledger rides as GitCurator-data/sync_review.db, best-effort
//     (an export/backup NEVER fails on the rider) and counts as a file;
//   - the done message: "{zip_path} ({count} files, {size})" with the
//     1-decimal MB/KB size string;
//   - the restore NEVER touches the current vault: the pick extracts
//     into vault-restored-<ts> NEXT TO it ("Review and merge manually").
package app

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/syncreview"
)

// exportLedgerPathFn is the ledger-path seam (THE HERMETIC-SEAM law:
// tests inject a temp ledger; the production default is the app dir's
// sync_review.db — LedgerDefaultPath's twin of ledger_default_path).
var exportLedgerPathFn = func() string {
	return syncreview.LedgerDefaultPath("")
}

// ExportZipGuards is the export door's honest refusal
// (_backup_export_zip 1276-1279): "" title = the door opens.
func ExportZipGuards(cfg *storage.OrderedMap) (title, body string) {
	if strings.TrimSpace(pyStrCfg(cfg, "vault_path")) == "" {
		return "No Vault", "Set a vault path first (📁 Vault tab)."
	}
	return "", ""
}

// pyExportStamp is the export's default filename stamp
// (datetime.now().strftime('%Y-%m-%d_%H-%M-%S') — local time, the
// reference's own shape).
func pyExportStamp(t time.Time) string {
	return t.Format("2006-01-02_15-04-05")
}

// ExportDefaultName answers the save dialog's default filename.
func ExportDefaultName(now time.Time) string {
	return "vault-export-" + pyExportStamp(now) + ".zip"
}

// sizeStrMBKB mirrors the reference's size string: 1 decimal, MB at
// >= 1 MiB, KB below.
func sizeStrMBKB(size int64) string {
	if size >= 1024*1024 {
		return fmt.Sprintf("%.1f MB", float64(size)/1024/1024)
	}
	return fmt.Sprintf("%.1f KB", float64(size)/1024)
}

// RunVaultExport is the ExportWorker's data half: walk the vault under
// the exclude laws, ride the ledger, answer the done message. It never
// speaks (the shell door does — the reference's worker only emits the
// done signal); title "" + message "" = the caller cancelled the pick.
func RunVaultExport(vaultPath, zipPath string, now time.Time) (message string, err error) {
	fileCount := 0
	zf, err := os.Create(zipPath)
	if err != nil {
		return "", err
	}
	zw := zip.NewWriter(zf)
	// The arc root: the vault's parent (relative_to(vault_path_obj.parent)).
	root := filepath.Dir(strings.TrimRight(vaultPath, string(os.PathSeparator)))
	var walkErr error
	filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // best-effort walk (the reference's os.walk swallows)
		}
		if info.IsDir() {
			if backupExclude[info.Name()] {
				return filepath.SkipDir
			}
			return nil
		}
		if backupExclude[info.Name()] {
			return nil
		}
		for _, suf := range backupExcludeSuffix {
			if strings.HasSuffix(info.Name(), suf) {
				return nil
			}
		}
		for _, pre := range backupExcludePrefix {
			if strings.Contains(path, pre) {
				return nil
			}
		}
		arc, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		src, oerr := os.Open(path)
		if oerr != nil {
			return nil
		}
		w, werr := zw.Create(filepath.ToSlash(arc))
		if werr == nil {
			if _, cerr := io.Copy(w, src); cerr != nil && walkErr == nil {
				walkErr = cerr
			}
			fileCount++
		}
		_ = src.Close()
		return nil
	})
	// v0.86.0 — THE DEFERRED LEDGER rider (best-effort — an export never
	// fails on it): the sync-review log rides beside the vault folder.
	if ledger := exportLedgerPathFn(); ledger != "" {
		if data, lerr := os.ReadFile(ledger); lerr == nil {
			if w, werr := zw.Create("GitCurator-data/sync_review.db"); werr == nil {
				_, _ = w.Write(data)
				fileCount++
			}
		}
	}
	_ = zw.Close()
	_ = zf.Close()
	if walkErr != nil {
		return "", walkErr
	}
	size := int64(0)
	if fi, serr := os.Stat(zipPath); serr == nil {
		size = fi.Size()
	}
	return fmt.Sprintf("%s (%d files, %s)", zipPath, fileCount,
		sizeStrMBKB(size)), nil
}

// RestoreBackupRow is one row of the restore dialog's list (the
// reference's "{name}    [ {size} ]" line split into its two fields;
// THE FIELD-TAG LAW: the frontend reads the tagged names).
type RestoreBackupRow struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

// ListRestoreBackups is the restore dialog's list law: the vault-*.zip
// pile, newest first (sorted reverse — the reference's glob + sorted
// reverse=True), each with the size string. An empty answer renders the
// "No Backups" box.
func ListRestoreBackups(backupFolder string) []RestoreBackupRow {
	pile, _ := filepath.Glob(filepath.Join(backupFolder, "vault-*.zip"))
	sort.Sort(sort.Reverse(sort.StringSlice(pile)))
	rows := []RestoreBackupRow{}
	for _, p := range pile {
		size := int64(0)
		if fi, err := os.Stat(p); err == nil {
			size = fi.Size()
		}
		rows = append(rows, RestoreBackupRow{
			Name: filepath.Base(p),
			Size: sizeStrMBKB(size),
		})
	}
	return rows
}

// RestoreGuards is the restore door's honest refusals
// (_backup_restore 1356-1368): "" title = the dialog opens; the second
// pair is the empty-pile box ("No Backups" — success-shaped in the
// reference, success=True).
func RestoreGuards(cfg *storage.OrderedMap) (title, body string) {
	if strings.TrimSpace(pyStrCfg(cfg, "backup_folder")) == "" {
		return "No Backup Folder", "Select a backup folder first."
	}
	return "", ""
}

// RestoreEmptyBox is the no-backups box (the reference's own words).
func RestoreEmptyBox() (title, body string) {
	return "No Backups", "No backups found yet.\n\nClick 'Backup Now' first."
}

// pyRestoreStamp is the restore folder's stamp
// (datetime.now().strftime('%Y%m%d-%H%M%S') — local time).
func pyRestoreStamp(t time.Time) string {
	return t.Format("20060102-150405")
}

// RestoreBackupToNewFolder is the safety net itself: extract the picked
// backup into vault-restored-<ts> NEXT TO the current vault (mkdir
// parents; the current vault is never touched — the reference's
// extractall over the fresh folder).
func RestoreBackupToNewFolder(backupPath, vaultPath string, now time.Time) (string, error) {
	newFolder := filepath.Join(filepath.Dir(strings.TrimRight(vaultPath,
		string(os.PathSeparator))), "vault-restored-"+pyRestoreStamp(now))
	if err := os.MkdirAll(newFolder, 0o755); err != nil {
		return "", err
	}
	zr, err := zip.OpenReader(backupPath)
	if err != nil {
		return newFolder, err
	}
	defer zr.Close()
	for _, f := range zr.File {
		dest := filepath.Join(newFolder, filepath.FromSlash(f.Name))
		// The safety belt the reference gets from extractall's own path
		// handling: refuse a member that escapes the fresh folder.
		if !strings.HasPrefix(dest, newFolder+string(os.PathSeparator)) {
			continue
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(dest, 0o755); err != nil {
				return newFolder, err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(dest), 0o755); err != nil {
			return newFolder, err
		}
		rc, oerr := f.Open()
		if oerr != nil {
			return newFolder, oerr
		}
		out, cerr := os.Create(dest)
		if cerr != nil {
			_ = rc.Close()
			return newFolder, cerr
		}
		_, _ = io.Copy(out, rc)
		_ = out.Close()
		_ = rc.Close()
	}
	return newFolder, nil
}

// CopyBackupFile is do_copy's shutil.copy2 twin: the bytes + the mtime
// (best-effort metadata) to the picked destination.
func CopyBackupFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err = io.Copy(out, in); err != nil {
		_ = out.Close()
		return err
	}
	if err = out.Close(); err != nil {
		return err
	}
	if fi, serr := os.Stat(src); serr == nil {
		_ = os.Chtimes(dst, fi.ModTime(), fi.ModTime())
	}
	return nil
}
