// Package dryrun ports app/gitcurator/core/dryrun.py — the Phase 0
// dry-run mechanism (v0.09.5): a global rehearsal switch, the
// withheld-operation log, the filesystem stand-ins that record instead of
// touch, the shadow cache copy (WAL sidecars included), and the Markdown
// report renderer.
//
// Spec: docs/port-blueprint/module-specs/spec-core-storage.md §dryrun.py
// (the dryrun contract lives in that file — the "spec-core-dryrun.md"
// pointer in the older task list was stale and reality-fixed with this
// task); task T-12 (Phase 1). Parity is pinned by the generated goldens in
// testdata/ (see dryrun_test.go).
//
// Preserved quirks (verified against the reference @ 98f4fc7 and pinned
// by TestGoldenDryrunStandin; NOT pre-classified fixes in doc 09 §5, so
// they are ported literally, never silently "fixed"):
//
//   - disable() zeroes the shadow dir BEFORE the cleanup call reads it —
//     the cleanup is dead code and the temp dir SURVIVES disable().
//   - enable() resets the shadow cache path but NOT the shadow dir — a
//     second enable() reuses the same temp dir.
//
// Thread model (as in Python): one batch at a time; the log is
// mutex-guarded because the banner thread writes markers concurrently.
package dryrun

import (
	"github.com/assadigit/second-brain-project/internal/helpers"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// PreviewChars mirrors PREVIEW_CHARS: how many characters of withheld
// content to keep in the log preview (dryrun.py:57).
const PreviewChars = 160

// Entry is one withheld operation (dryrun.py record()'s entry dict).
// Size is nil when no content was recorded (Python None).
type Entry struct {
	Time    string
	Op      string
	Path    string
	Size    *int
	Preview string
	Note    string
}

// OpCount is one summary() pair; Summary preserves first-seen order
// (Python dict insertion order — observable via the golden envelope).
type OpCount struct {
	Op    string
	Count int
}

var d struct {
	mu          sync.Mutex
	enabled     bool
	entries     []Entry
	shadowDir   string // temp dir holding the throwaway cache copy
	shadowCache string // resolved path of the shadow cache ("" == unset)
}

// ---------------------------------------------------------------------------
// Switch (dryrun.py:74-95)
// ---------------------------------------------------------------------------

// Enable turns dry-run ON. Starts a fresh log (previous entries are
// dropped). NOTE (preserved quirk): the shadow DIR is intentionally NOT
// reset — only the cached path is.
func Enable() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.enabled = true
	d.entries = nil
	d.shadowCache = ""
}

// Disable turns dry-run OFF. Keeps the log so the caller can still print
// the report. NOTE (preserved quirk, ported literally): the shadow dir
// field is zeroed BEFORE the cleanup call runs, so the cleanup never
// removes anything (dead code in the reference; the temp dir survives).
func Disable() {
	d.mu.Lock()
	d.enabled = false
	d.shadowDir = ""
	d.shadowCache = ""
	d.mu.Unlock()
	cleanupShadowDir()
}

// IsEnabled reports whether a dry-run batch is active.
func IsEnabled() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.enabled
}

// ---------------------------------------------------------------------------
// Log (dryrun.py:102-157)
// ---------------------------------------------------------------------------

// Record one withheld operation. Called by every helper while enabled.
// Content may be string, []byte, or nil; only a short preview is kept.
// Never panics — a logging failure must not break a batch (Python wraps
// everything in try/except).
func Record(op, path string, content any, note string) {
	defer func() { _ = recover() }() // "never raises"
	var size *int
	preview := ""
	switch c := content.(type) {
	case []byte:
		n := len(c)
		cut := c
		if len(cut) > PreviewChars {
			cut = cut[:PreviewChars] // Python slices BYTES first, then decodes
		}
		preview = pyUTF8Replace(cut)
		size = &n
	case string:
		n := len(c) // UTF-8 byte length (Python len(content.encode('utf-8')))
		r := []rune(c)
		if len(r) > PreviewChars {
			r = r[:PreviewChars] // Python slices str by CODE POINTS
		}
		preview = string(r)
		size = &n
	default:
		// nil or unknown types: size stays None, preview stays "".
	}
	preview = strings.ReplaceAll(preview, "\r", "")
	preview = strings.ReplaceAll(preview, "\n", "⏎")
	// len() on a Python str counts code points — so does this.
	if len([]rune(preview)) >= PreviewChars {
		preview = string([]rune(preview)[:PreviewChars]) + "…"
	}
	d.mu.Lock()
	d.entries = append(d.entries, Entry{
		Time:    time.Now().Format("15:04:05"), // %H:%M:%S local
		Op:      op,
		Path:    path,
		Size:    size,
		Preview: preview,
		Note:    note,
	})
	d.mu.Unlock()
}

// Entries returns a copy of the withheld-operation log.
func Entries() []Entry {
	d.mu.Lock()
	defer d.mu.Unlock()
	out := make([]Entry, len(d.entries))
	copy(out, d.entries)
	return out
}

// EntryCount returns the number of recorded entries.
func EntryCount() int {
	d.mu.Lock()
	defer d.mu.Unlock()
	return len(d.entries)
}

// Clear drops the log (used by tests).
func Clear() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.entries = nil
}

// Summary returns counts per operation type, first-seen order.
func Summary() []OpCount {
	d.mu.Lock()
	defer d.mu.Unlock()
	var out []OpCount
	index := map[string]int{}
	for _, e := range d.entries {
		if i, ok := index[e.Op]; ok {
			out[i].Count++
			continue
		}
		index[e.Op] = len(out)
		out = append(out, OpCount{Op: e.Op, Count: 1})
	}
	return out
}

// ---------------------------------------------------------------------------
// Filesystem stand-ins (dryrun.py:165-208) — identical to the raw calls
// they replace when OFF, recorded (not performed) when ON.
// ---------------------------------------------------------------------------

// MkdirAll is the os.makedirs stand-in. Dry-run: records, does not create.
// The note mirrors dryrun.py: 'create folder' vs 'folder exists'.
func MkdirAll(path string, existOk bool) error {
	if IsEnabled() {
		note := "create folder"
		if _, err := os.Stat(path); err == nil {
			note = "folder exists"
		}
		Record("makedirs", path, nil, note)
		return nil
	}
	// os.makedirs(path, exist_ok=exist_ok): EEXIST is suppressed only when
	// the existing path is a directory.
	st, err := os.Stat(path)
	if err == nil {
		if st.IsDir() && existOk {
			return nil
		}
		return errExist(path)
	}
	return os.MkdirAll(path, 0o777)
}

// WriteText is the plain open(path,'w') stand-in. Dry-run: records, does
// not write. On Windows the reference's text-mode write translates '\n'
// to os.linesep; this port performs the same translation only when built
// for windows (documented in SESSION-004).
func WriteText(path, content string) error {
	if IsEnabled() {
		Record("write", path, content, "")
		return nil
	}
	return os.WriteFile(path, []byte(translateNewlines(content)), 0o666)
}

// AppendText is the plain open(path,'a') stand-in. Dry-run: records, does
// not append.
func AppendText(path, text string) error {
	if IsEnabled() {
		Record("append", path, text, "")
		return nil
	}
	f, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o666)
	if err != nil {
		return err
	}
	if _, err := f.WriteString(translateNewlines(text)); err != nil {
		f.Close()
		return err
	}
	return f.Close()
}

// Remove is the os.remove stand-in. Dry-run: records, does not delete.
// When disabled it returns exactly os.Remove's error — the caller's
// existing error handling keeps working unchanged.
func Remove(path string) error {
	if IsEnabled() {
		Record("remove", path, nil, "delete file")
		return nil
	}
	return os.Remove(path)
}

// Move is the shutil.move stand-in. Dry-run: records on DST with
// note "from SRC", does not move.
func Move(src, dst string) error {
	if IsEnabled() {
		Record("move", dst, nil, "from "+src)
		return nil
	}
	// shutil.move: an existing directory destination receives the file
	// (dst/basename(src)).
	if st, err := os.Stat(dst); err == nil && st.IsDir() {
		dst = filepath.Join(dst, filepath.Base(src))
	}
	renameErr := os.Rename(src, dst)
	if renameErr == nil {
		return nil
	}
	// Cross-device fallback (shutil.move: copy2 + remove). If the fallback
	// also fails, the ORIGINAL rename error surfaces (shutil.move surfaces
	// the underlying OS error).
	data, rerr := os.ReadFile(src)
	if rerr != nil {
		return renameErr
	}
	if werr := os.WriteFile(dst, data, 0o666); werr != nil {
		return renameErr
	}
	if st, serr := os.Stat(src); serr == nil {
		_ = os.Chtimes(dst, st.ModTime(), st.ModTime()) // copy2 parity
	}
	return os.Remove(src)
}

// ---------------------------------------------------------------------------
// Shadow cache (dryrun.py:215-266)
// ---------------------------------------------------------------------------

// ShadowCachePath returns the path of a throwaway COPY of sourceDB for a
// dry-run batch. The copy is made once per enable() session; -wal/-shm
// sidecars are copied when present (a live SQLite database may keep recent
// commits in them). A missing source is normal (fresh install) and is
// recorded as an info entry; a failed copy is recorded as a warning and
// the dry-run proceeds with an empty shadow. The copy itself happens
// OUTSIDE the lock — Record takes the same lock (the Phase 0 deadlock
// lesson, caught live in the end-to-end demo).
func ShadowCachePath(sourceDB string) string {
	d.mu.Lock()
	if d.shadowCache != "" {
		defer d.mu.Unlock()
		return d.shadowCache
	}
	if d.shadowDir == "" {
		dir, err := os.MkdirTemp("", "gitcurator-dryrun-")
		if err != nil {
			d.mu.Unlock()
			Record("warning", "gitcurator-dryrun", nil,
				"could not create the shadow dir ("+err.Error()+"); "+
					"the dry-run sees an empty cache")
			return ""
		}
		d.shadowDir = dir
	}
	shadow := filepath.Join(d.shadowDir, "cache.db")
	d.shadowCache = shadow
	d.mu.Unlock()

	if _, err := os.Stat(sourceDB); err == nil {
		if err := copyFile(sourceDB, shadow); err != nil {
			Record("warning", shadow, nil,
				"could not copy the real cache ("+err.Error()+"); "+
					"the dry-run sees an empty cache")
		} else {
			for _, suffix := range []string{"-wal", "-shm"} {
				side := sourceDB + suffix
				if _, err := os.Stat(side); err == nil {
					if err := copyFile(side, shadow+suffix); err != nil {
						Record("warning", shadow+suffix, nil,
							"could not copy the real cache ("+err.Error()+"); "+
								"the dry-run sees an empty cache")
					}
				}
			}
		}
	} else {
		Record("info", shadow, nil,
			"no real cache.db yet — the dry-run starts from an empty shadow cache")
	}
	return shadow
}

// cleanupShadowDir is the best-effort removal called by Disable. PRESERVED
// QUIRK: Disable zeroes the dir field BEFORE this runs, so it never
// removes anything (dead code in the reference, ported literally — the
// observable behavior "temp dir survives disable" is pinned by the
// golden).
func cleanupShadowDir() {
	d.mu.Lock()
	dir := d.shadowDir
	d.mu.Unlock()
	if dir != "" {
		if _, err := os.Stat(dir); err == nil {
			_ = os.RemoveAll(dir)
		}
	}
}

// copyFile mirrors shutil.copy2 for files: content + mtime.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o666); err != nil {
		return err
	}
	if st, err := os.Stat(src); err == nil {
		return os.Chtimes(dst, st.ModTime(), st.ModTime())
	}
	return nil
}

// ---------------------------------------------------------------------------
// Reporting (dryrun.py:273-304)
// ---------------------------------------------------------------------------

// RenderReportMarkdown renders the withheld-operation log as a readable
// Markdown document.
func RenderReportMarkdown(title string) string {
	es := Entries()
	counts := Summary()
	var lines []string
	lines = append(lines, "# "+title, "")
	lines = append(lines, "*Generated: "+time.Now().Format("2006-01-02 15:04:05")+"*", "")
	lines = append(lines, "**Nothing below was written.** Every operation was recorded instead of performed.", "")
	if len(es) == 0 {
		lines = append(lines, "No vault writes were attempted during this dry-run.", "")
		return strings.Join(lines, "\n")
	}
	lines = append(lines, "**"+strconv.Itoa(len(es))+" operation(s) would have been performed:**", "")

	// sorted(counts.items()) — alphabetical by op
	sorted := make([]OpCount, len(counts))
	copy(sorted, counts)
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].Op < sorted[j].Op })
	for _, c := range sorted {
		lines = append(lines, "- "+c.Op+": "+strconv.Itoa(c.Count))
	}
	lines = append(lines, "")
	lines = append(lines, "| # | Time | Op | Target | Size | Preview / note |")
	lines = append(lines, "|---|------|----|--------|------|----------------|")
	for i, e := range es {
		size := ""
		if e.Size != nil {
			size = helpers.FormatComma(int64(*e.Size)) + " B" // f"{n:,} B"
		}
		extra := e.Preview
		if e.Preview == "" {
			extra = e.Note
		} else if e.Note != "" {
			extra = e.Preview + " (" + e.Note + ")"
		}
		extra = strings.ReplaceAll(extra, "|", `\|`)
		lines = append(lines, "| "+strconv.Itoa(i+1)+" | "+e.Time+" | "+e.Op+" | `"+e.Path+"` | "+size+" | "+extra+" |")
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n")
}
