// backupjob.go — T-71 slice (d): the BackupWorker port
// (gui/main_window/backup_seal.py:548-641 — _backup_now's background
// zip job): the vault zip with the reference's exact exclude rules,
// the FIFO rotation over backup_max, and THE DEFERRED LEDGER rider
// (v0.86.0 — the sync-review log rides every vault zip so the review
// memory survives a machine loss). The result verdict speaks the
// reference's own shape: "{file} ({count} files, {size})"; a backup
// never fails on the rider (best-effort).
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
)

// The reference's exclude laws (BackupWorker.run): never the config,
// never the git machinery, never the caches; no .lock files; no
// Obsidian workspace state.
var backupExclude = map[string]bool{
	"config.json": true, ".git": true, "__pycache__": true,
}

var backupExcludeSuffix = []string{".lock"}
var backupExcludePrefix = []string{".obsidian/workspace", ".obsidian/app.json"}

// RunBackupNow is the production BackupNow seam: build
// vault-<timestamp>.zip in the configured backup folder (created when
// missing), rotate the FIFO pile down to backup_max (default 10),
// and return the done-signal's payload.
func RunBackupNow(cfg *storage.OrderedMap, log LogFn) *SealResultRow {
	vaultPath := pyStrCfg(cfg, "vault_path")
	backupFolder := pyStrCfg(cfg, "backup_folder")
	if vaultPath == "" {
		return &SealResultRow{Success: false,
			Message: "Set a vault path first (📁 Vault tab).", Level: "error"}
	}
	if backupFolder == "" {
		return &SealResultRow{Success: false,
			Message: "Select a backup folder first.", Level: "error"}
	}
	if !fsIsDirPath(backupFolder) {
		if err := os.MkdirAll(backupFolder, 0o755); err != nil {
			return &SealResultRow{Success: false,
				Message: "Cannot create folder: " + pyErrString(err), Level: "error"}
		}
	}
	maxBackups := 10
	if v, ok := cfg.Get("backup_max"); ok {
		switch t := v.(type) {
		case float64:
			if t > 0 {
				maxBackups = int(t)
			}
		case int:
			if t > 0 {
				maxBackups = t
			}
		}
	}

	timestamp := time.Now().UTC().Format("2006-01-02T150405Z")
	fileName := "vault-" + timestamp + ".zip"
	zipPath := filepath.Join(backupFolder, fileName)

	fileCount := 0
	zf, err := os.Create(zipPath)
	if err != nil {
		return &SealResultRow{Success: false, Message: pyErrString(err), Level: "error"}
	}
	zw := zip.NewWriter(zf)
	// The walk: the vault's parent is the arc root (the reference's
	// relative_to(vault_path_obj.parent) — the zip carries the vault
	// FOLDER, not its contents loose).
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
	// v0.86.0 — THE DEFERRED LEDGER rider (best-effort — a backup
	// never fails on it): the sync-review log rides beside the vault
	// folder inside the zip. (T-71 had recorded the rider's absence
	// against the then-placeholder syncreview seat; the ledger is
	// real since the Phase-3/4 landings — the rider gains its
	// bytes at T-80, reality wins.)
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
		return &SealResultRow{Success: false, Message: pyErrString(walkErr), Level: "error"}
	}

	// FIFO rotation: keep the newest backup_max, drop the rest
	// (best-effort removals).
	pile, _ := filepath.Glob(filepath.Join(backupFolder, "vault-*.zip"))
	if len(pile) > maxBackups {
		sort.Strings(pile)
		for _, old := range pile[:len(pile)-maxBackups] {
			_ = os.Remove(old)
		}
	}

	size := int64(0)
	if fi, serr := os.Stat(zipPath); serr == nil {
		size = fi.Size()
	}
	sizeStr := fmt.Sprintf("%.1f MB", float64(size)/1024/1024)
	if size < 1024*1024 {
		sizeStr = fmt.Sprintf("%.1f KB", float64(size)/1024)
	}
	return &SealResultRow{Success: true,
		Message: fmt.Sprintf("💾 Backup created: %s (%d files, %s)",
			fileName, fileCount, sizeStr),
		Level: "success"}
}
