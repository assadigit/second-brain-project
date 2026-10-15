package storage

import (
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
)

// AtomicWriteText atomically writes content to path (storage.py:37-70):
// a temp file in the SAME directory (prefix ".tmp_", suffix = the
// target's extension or ".tmp"), fsync, then rename over the target —
// readers only ever see the old complete file or the new complete file.
// While a dry-run batch is active the write is recorded instead
// (op "atomic-write").
//
// Port notes (parity, even where suboptimal — spec compat reqs 1–2):
// no chmod (Python's mkstemp 0600 == os.CreateTemp), no directory fsync;
// Python's newline=” means no newline translation, so the bytes written
// are exactly the UTF-8 encoding of content (the Python encoding
// parameter is always utf-8 at every call site).
func AtomicWriteText(path, content string) error {
	if dryrun.IsEnabled() {
		dryrun.Record("atomic-write", path, content, "")
		return nil
	}
	return atomicWrite(path, []byte(content))
}

// AtomicWriteBytes atomically writes binary data to path (banners,
// images; storage.py:73-99). Dry-run records the bytes content.
func AtomicWriteBytes(path string, data []byte) error {
	if dryrun.IsEnabled() {
		dryrun.Record("atomic-write", path, data, "")
		return nil
	}
	return atomicWrite(path, data)
}

func atomicWrite(path string, data []byte) error {
	directory := filepath.Dir(path)
	if err := os.MkdirAll(directory, 0o777); err != nil {
		return err
	}
	// Python: tempfile.mkstemp(prefix='.tmp_', suffix=ext-or-'.tmp',
	// dir=directory) — the suffix rule is os.path.splitext on the
	// BASENAME: a leading dot is NOT an extension (".hidden" → ".tmp").
	_, ext := pySplitExt(filepath.Base(path))
	if ext == "" {
		ext = ".tmp"
	}
	tmp, err := os.CreateTemp(directory, ".tmp_*"+ext)
	if err != nil {
		return err
	}
	tmpName := tmp.Name()
	cleanup := true
	defer func() {
		if cleanup {
			// BaseException parity: best-effort unlink, then re-raise.
			tmp.Close()
			os.Remove(tmpName)
		}
	}()
	if _, err := tmp.Write(data); err != nil {
		return err
	}
	if err := tmp.Sync(); err != nil {
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	// os.replace: atomic on both NTFS and POSIX.
	if err := os.Rename(tmpName, path); err != nil {
		return err
	}
	cleanup = false
	return nil
}

// ---------------------------------------------------------------------------
// Path helpers (storage.py:106-138)
// ---------------------------------------------------------------------------

// safeFilenameRe strips everything outside [a-zA-Z0-9\-_] (one underscore
// per run) — this ALSO defeats path traversal (no '..', '/' or '\' can
// survive). Unicode collapses to '_' per run.
var safeFilenameRe = regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)

// SafeFilename makes a string safe for use as a filename component
// (storage.py safe_filename): every run of characters outside
// [a-zA-Z0-9\-_] becomes ONE underscore.
func SafeFilename(name string) string {
	return safeFilenameRe.ReplaceAllString(name, "_")
}

// UniquePath returns a non-colliding path: appends _v1, _v2, ... if path
// exists (storage.py unique_path — the extension split follows Python's
// os.path.splitext, so ".hidden" → ".hidden_v1", not "_v1.hidden").
// Not race-free; callers re-check.
func UniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	base, ext := pySplitExt(path)
	for counter := 1; ; counter++ {
		candidate := base + "_v" + strconv.Itoa(counter) + ext
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

// BuildNoteFilename is the canonical note filename:
// <repo>_<category>_<first-tag>.md (storage.py build_note_filename). An
// empty first tag yields an empty component ("" → ""), and no tags at
// all falls back to "misc" — both preserved literally.
func BuildNoteFilename(repoName, categoryKey string, tags []string) string {
	safeName := SafeFilename(repoName)
	secondary := "misc"
	if len(tags) > 0 {
		secondary = SafeFilename(tags[0])
	}
	category := SafeFilename(strings.ReplaceAll(categoryKey, "/", "_"))
	return safeName + "_" + category + "_" + secondary + ".md"
}

// pySplitExt ports Python's os.path.splitext (posixpath semantics: '/' is
// the only separator — the reference's goldens were generated under
// posixpath; ntpath additionally treats '\\' as a separator, a
// Windows-port caveat recorded in SESSION-004). The extension is
// everything from the LAST dot to the end, PROVIDED at least one non-dot
// character precedes that dot in the final path component; otherwise
// there is no extension (".hidden" → (".hidden", ""), "..dots" →
// ("..dots", ""), "a." → ("a", ".")).
func pySplitExt(path string) (string, string) {
	dot := strings.LastIndexByte(path, '.')
	if dot < 0 {
		return path, ""
	}
	sep := strings.LastIndexByte(path, '/')
	if dot < sep {
		// the last dot belongs to a parent component ("a.b/c")
		return path, ""
	}
	// skip leading dots of the final component: the dot must follow at
	// least one non-dot character (posixpath _splitext).
	for i := sep + 1; i < dot; i++ {
		if path[i] != '.' {
			return path[:dot], path[dot:]
		}
	}
	return path, ""
}
