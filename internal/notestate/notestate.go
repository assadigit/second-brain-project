package notestate

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notebuilder"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/taxonomy"
)

// ---------------------------------------------------------------------------
// Configuration (note_state.py:56-78)
// ---------------------------------------------------------------------------

// RecallStart / RecallEnd delimit the app-managed block (Phase 6 writes
// it; the fingerprint ignores it).
const (
	RecallStart = "<!-- gitcurator:recall:start -->"
	RecallEnd   = "<!-- gitcurator:recall:end -->"
)

// The same special folders VaultIndex skips — plus _missing and .trash
// (note_state.py:60-67: a swept/deleted placeholder must never read as
// a dismissal of the repo itself). SUBSTRING match on the walked path.
var SkippedFolderMarks = []string{
	"_moc", "_inbox", "_missing", "attachments", ".obsidian", ".trash",
}

// Vault labels (identity = label, not path — paths move).
const (
	VaultGithub   = "github"
	VaultWebsites = "websites"
)

// folderToCategory — the reverse of CATEGORY_FOLDERS (folder -> key).
var folderToCategory = func() map[string]string {
	m := make(map[string]string, len(constants.CategoryFolders))
	for _, e := range constants.CategoryFolders {
		m[e.Folder] = e.Key
	}
	return m
}()

// pyS is Python's str-pattern \s class (Unicode whitespace):
// [\t\n\f\r ], \v, \x1c-\x1f, \x85 and the Z categories. Go's own \s is
// ASCII-only; every ported pattern that carried a Python \s uses this.
const pyS = `[\s\v\x1c-\x1f\x85\p{Z}]`

var (
	recallBlockRE = regexp.MustCompile(
		`(?s)` + regexp.QuoteMeta(RecallStart) + `.*?` + regexp.QuoteMeta(RecallEnd))

	sourceRE = regexp.MustCompile(`(?m)^source:` + pyS + `*(.+)$`)

	tagsRE = regexp.MustCompile(`^tags:` + pyS + `*\[(.*)\]` + pyS + `*$`)
)

// LogFn mirrors the reference's optional log callback: log(msg, level).
// nil means the no-op Python falls back to (lambda *a, **k: None).
type LogFn func(msg, level string)

func noopLog(msg, level string) {}

// ---------------------------------------------------------------------------
// Fingerprint (note_state.py:85-95)
// ---------------------------------------------------------------------------

// ComputeFingerprint mirrors compute_fingerprint: the app-managed recall
// block (if present) is stripped (non-greedy, DOTALL, each block), the
// tail is rstripped (Unicode whitespace), then sha256 hex of the UTF-8
// bytes. Pinned byte-exact by testdata/notestate_pure.golden.json
// (generated from the reference's own compute_fingerprint, invalid-UTF-8
// replacement semantics included).
func ComputeFingerprint(content string) string {
	stripped := helpers.Rstrip(recallBlockRE.ReplaceAllString(content, ""))
	sum := sha256.Sum256([]byte(stripped))
	return hex.EncodeToString(sum[:])
}

// stripRecallBlock is the strip half of the fingerprint, exposed to the
// characterization tests (the pure golden pins the intermediate itself).
func stripRecallBlock(content string) string {
	return helpers.Rstrip(recallBlockRE.ReplaceAllString(content, ""))
}

// ---------------------------------------------------------------------------
// Vault walking (note_state.py:102-175)
// ---------------------------------------------------------------------------

// skipDir — SUBSTRING match on the walked directory path (note_state.py:
// 102-103): a mark anywhere in the path (vault prefix included) skips
// the directory and everything under it.
func skipDir(root string) bool {
	for _, mark := range SkippedFolderMarks {
		if strings.Contains(root, mark) {
			return true
		}
	}
	return false
}

// Note is one scanned vault file (note_state.py:146-153). Category comes
// from the FOLDER, never the frontmatter (§4.4: the folder is the
// correction signal).
type Note struct {
	Path        string
	SourceURL   string
	Folder      string
	Category    *string
	Subcategory *string
	Unmanaged   bool
}

// Resolver maps a vault-relative folder to (category, subcategory)
// (note_state.py:106's resolver parameter; None = folder_category).
type Resolver func(folder string) (category, subcategory *string)

// ScanVault mirrors scan_vault (note_state.py:106-154): a read-only walk
// returning one Note per .md file — with a source: line in the first
// 800 characters (code points, utf-8 errors='replace'), plus every
// unmanaged file flagged as such. Unreadable files are skipped silently.
// The walk is lexical (documented divergence — see package doc).
func ScanVault(vaultPath string, resolver Resolver) []Note {
	if resolver == nil {
		resolver = FolderCategory
	}
	var notes []Note
	if vaultPath == "" {
		return notes
	}
	info, err := os.Stat(vaultPath)
	if err != nil || !info.IsDir() {
		return notes
	}
	_ = filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return nil // os.walk's silent skip (onerror=None)
		}
		if d.IsDir() {
			if skipDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		rel, relErr := filepath.Rel(vaultPath, path)
		if relErr != nil {
			return nil
		}
		rel = strings.ReplaceAll(rel, "\\", "/")
		folder := posixDirname(rel)
		category, subcategory := resolver(folder)
		head, ok := readHead800(path)
		if !ok {
			return nil
		}
		sourceURL := ""
		if m := sourceRE.FindStringSubmatch(head); m != nil {
			sourceURL = links.NormalizeURL(
				strings.Trim(helpers.Strip(m[1]), "\"'"))
		}
		notes = append(notes, Note{
			Path:        path,
			SourceURL:   sourceURL,
			Folder:      folder,
			Category:    category,
			Subcategory: subcategory,
			Unmanaged:   sourceURL == "",
		})
		return nil
	})
	return notes
}

// posixDirname is os.path.dirname over an already '/'-joined relpath:
// everything before the last '/', or "" when there is none.
func posixDirname(rel string) string {
	if i := strings.LastIndex(rel, "/"); i >= 0 {
		return rel[:i]
	}
	return ""
}

// readHead800 mirrors open(..., encoding='utf-8', errors='replace')
// .read(800): the first 800 CODE POINTS of the replaced-decoded text —
// INCLUDING text mode's universal-newlines translation (\r\n and lone
// \r become \n; SESSION-008 fidelity fix: CRLF notes are the norm on
// the owner's Windows vault, and the reference's fingerprint hashes the
// TRANSLATED head — pinned by the pure golden's crlf_file_read case).
func readHead800(path string) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	return helpers.CutRunes(helpers.PyNewlines(decodeUTF8Replace(data)), 800), true
}

// decodeUTF8Replace mirrors Python's utf-8 errors='replace' decode with
// the maximal-subpart rule the reference interpreter applies: each
// maximal invalid subpart (the longest prefix that is still a prefix of
// some valid sequence) collapses into ONE U+FFFD. Pinned byte-exact by
// the pure golden's invalid-UTF-8 battery:
//
//	b"\xff"→1×U+FFFD; b"\xc2A"→U+FFFD+"A"; b"a\xe2\x82"→"a"+1×U+FFFD
//	(the truncated 3-byte attempt is ONE subpart); b"\xc3\xc3A"→2×
//	U+FFFD+"A" (each C3 is its own failed attempt).
func decodeUTF8Replace(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	var b strings.Builder
	b.Grow(len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r != utf8.RuneError || size != 1 {
			// Valid sequence — keep the original bytes (a literal
			// U+FFFD in the input stays exactly the bytes it was).
			b.Write(data[:size])
			data = data[size:]
			continue
		}
		b.WriteRune(utf8.RuneError)
		data = data[subpartLen(data):]
	}
	return b.String()
}

// subpartLen returns the length of the maximal invalid subpart starting
// at data[0]: the lead byte plus however many following bytes are legal
// continuations of a still-possibly-valid sequence (lead-specific range
// for the first continuation, 0x80-0xBF for the rest). C0/C1 (overlong
// 2-byte leads), 0x80-0xBF (bare continuations) and 0xF5-0xFF are
// invalid leads on their own.
func subpartLen(data []byte) int {
	lead := data[0]
	var expected int
	switch {
	case lead < 0xC2: // bare continuation, C0, C1
		return 1
	case lead < 0xE0: // C2-DF: one continuation
		expected = 1
	case lead < 0xF0: // E0-EF: two
		expected = 2
	case lead < 0xF5: // F0-F4: three
		expected = 3
	default: // F5-FF
		return 1
	}
	n := 1
	for i := 1; i <= expected && i < len(data); i++ {
		c := data[i]
		if c < 0x80 || c > 0xBF {
			break
		}
		// Lead-specific first-continuation ranges (anti-overlong and
		// anti-surrogate): E0 needs A0-BF, ED needs 80-9F, F0 needs
		// 90-BF, F4 needs 80-8F.
		if i == 1 {
			switch {
			case lead == 0xE0 && c < 0xA0:
				return n
			case lead == 0xED && c > 0x9F:
				return n
			case lead == 0xF0 && c < 0x90:
				return n
			case lead == 0xF4 && c > 0x8F:
				return n
			}
		}
		n++
	}
	return n
}

// FolderCategory mirrors folder_category (note_state.py:157-175): exact
// CATEGORY_FOLDERS reverse hit wins; the vault root and _review are
// (nil, nil); a registered prefix maps the rest into the subcategory;
// anything else keeps the folder name raw with subcategory nil
// (classify_changes reports those as 'unmapped').
func FolderCategory(folder string) (*string, *string) {
	folder = strings.ReplaceAll(strings.Trim(folder, "/"), "\\", "/")
	if folder == "" || folder == "_review" {
		return nil, nil
	}
	if key, ok := folderToCategory[folder]; ok {
		return &key, nil
	}
	parts := strings.Split(folder, "/")
	for depth := len(parts) - 1; depth >= 1; depth-- {
		prefix := strings.Join(parts[:depth], "/")
		if key, ok := folderToCategory[prefix]; ok {
			sub := strings.Join(parts[depth:], "/")
			return &key, &sub
		}
	}
	return &folder, nil
}

// TaxonomyResolver mirrors _taxonomy_resolver (note_state.py:585-599):
// the websites-vault resolver — folders are taxonomy NAMES
// (Category or Category/Subcategory); special folders behave exactly
// like the GitHub resolver (root / _review -> (nil, nil)).
func TaxonomyResolver(t *taxonomy.Taxonomy) Resolver {
	return func(folder string) (*string, *string) {
		folder = strings.ReplaceAll(strings.Trim(folder, "/"), "\\", "/")
		if folder == "" || folder == "_review" {
			return nil, nil
		}
		var parts []string
		for _, p := range strings.Split(folder, "/") {
			if p != "" {
				parts = append(parts, p)
			}
		}
		if len(parts) == 0 {
			return nil, nil
		}
		category := parts[0]
		if len(parts) == 1 {
			return &category, nil
		}
		sub := strings.Join(parts[1:], "/")
		return &category, &sub
	}
}

// ---------------------------------------------------------------------------
// The database (note_state.py:182-436)
// ---------------------------------------------------------------------------

// The DDL the reference's own __init__ executes, byte-exact — WITH the
// IF NOT EXISTS guards (sqlite_master stores the NORMALIZED form — the
// IF NOT EXISTS clause is dropped — which is what the golden carries and
// what the schema-identity test compares). Copied from the reference's
// source, never hand-typed: regenerate with dump_notestate.py. Every
// column is TEXT/INTEGER-declared: no DATE/DATETIME/TIMESTAMP column
// exists here, so the CAST read law is satisfied vacuously (package
// doc).
const (
	ddlNoteState = `CREATE TABLE IF NOT EXISTS note_state (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    vault TEXT NOT NULL,
                    source_url TEXT NOT NULL,
                    path TEXT NOT NULL,
                    fingerprint TEXT NOT NULL,
                    category TEXT,
                    subcategory TEXT,
                    locked INTEGER NOT NULL DEFAULT 0,
                    first_seen TEXT NOT NULL,
                    updated_at TEXT NOT NULL,
                    UNIQUE(vault, source_url)
                )`
	ddlIdxNoteStateVault = `CREATE INDEX IF NOT EXISTS idx_note_state_vault ON note_state (vault)`
	ddlCorrectionsLog    = `CREATE TABLE IF NOT EXISTS corrections_log (
                    id INTEGER PRIMARY KEY AUTOINCREMENT,
                    ts TEXT NOT NULL,
                    vault TEXT NOT NULL,
                    source_url TEXT NOT NULL,
                    from_category TEXT,
                    to_category TEXT,
                    from_path TEXT,
                    to_path TEXT,
                    note TEXT
                )`
	ddlIdxCorrectionsVault = `CREATE INDEX IF NOT EXISTS idx_corrections_vault ON corrections_log (vault)`
	ddlNoteStateDismissed  = `CREATE TABLE IF NOT EXISTS note_state_dismissed (
                    vault TEXT NOT NULL,
                    source_url TEXT NOT NULL,
                    last_path TEXT,
                    dismissed_at TEXT NOT NULL,
                    PRIMARY KEY (vault, source_url)
                )`
)

// DB is the note-state store: the shared SQLite plumbing (ONE connection,
// the embedded not-reentrant RWMutex) plus the three §4.4 tables. Same
// file as CacheDB (the call sites' resolved path); own connection.
type DB struct {
	*store.DB
}

// New mirrors NoteStateDB.__init__ (note_state.py:187-242): the literal
// default "cache.db" is anchored to APP_DIR (the v0.09.4 fix — never the
// current working dir), the three tables + two indexes are created if
// absent, busy_timeout=30000 (plus the Go store law's WAL/NORMAL — the
// documented unified-pragmas divergence).
func New(dbPath string) (*DB, error) {
	if dbPath == "cache.db" {
		dbPath = filepath.Join(constants.AppDir(), "cache.db")
	}
	base, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	d := &DB{DB: base}
	if err := d.createTables(); err != nil {
		base.Close()
		return nil, err
	}
	return d, nil
}

func (d *DB) createTables() error {
	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return err
	}
	for _, stmt := range []string{
		ddlNoteState, ddlIdxNoteStateVault, ddlCorrectionsLog,
		ddlIdxCorrectionsVault, ddlNoteStateDismissed,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// nullable converts a *string into a driver value (nil ⇔ SQL NULL).
func nullable(s *string) any {
	if s == nil {
		return nil
	}
	return *s
}

// stringPtr converts a scanned nullable column back into a *string.
func stringPtr(ns sql.NullString) *string {
	if !ns.Valid {
		return nil
	}
	v := ns.String
	return &v
}

// nowSeconds is the write shape law: naive-local
// datetime.now().isoformat(timespec='seconds') — YYYY-MM-DDTHH:MM:SS
// (every note_state timestamp writes this shape).
func nowSeconds() string {
	return store.PyISOFormatSeconds(time.Now())
}

// -- writes -----------------------------------------------------------

// Upsert mirrors upsert (note_state.py:246-265): insert or refresh one
// note's row (identity = vault + source_url) as ONE transaction
// (UPDATE then INSERT OR IGNORE — the UPDATE-then-insert shape keeps
// first_seen on refresh, Python's execute-execute-commit under the lock).
func (d *DB) Upsert(vault, sourceURL, path, fingerprint string,
	category, subcategory *string, locked bool) error {
	now := nowSeconds()
	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return err
	}
	if _, err := tx.Exec(
		"UPDATE note_state SET path=?, fingerprint=?, category=?, "+
			"subcategory=?, locked=?, updated_at=? "+
			"WHERE vault=? AND source_url=?",
		path, fingerprint, nullable(category), nullable(subcategory),
		boolToInt(locked), now, vault, sourceURL); err != nil {
		_ = tx.Rollback()
		return err
	}
	if _, err := tx.Exec(
		"INSERT OR IGNORE INTO note_state "+
			"(vault, source_url, path, fingerprint, category, subcategory,"+
			" locked, first_seen, updated_at) VALUES (?,?,?,?,?,?,?,?,?)",
		vault, sourceURL, path, fingerprint, nullable(category),
		nullable(subcategory), boolToInt(locked), now, now); err != nil {
		_ = tx.Rollback()
		return err
	}
	return tx.Commit()
}

// RecordNote mirrors record_note (note_state.py:267-287): record one note
// the app just wrote. content nil ⇒ the file is read (utf-8 replace);
// normalizer nil ⇒ links.NormalizeURL; a normalizer returning "" falls
// back to the raw source_url.
func (d *DB) RecordNote(vault, sourceURL, path string, content *string,
	category, subcategory *string, normalizer func(string) string) error {
	text := ""
	if content != nil {
		text = *content
	} else if data, err := os.ReadFile(path); err == nil {
		text = decodeUTF8Replace(data)
	}
	keyFn := normalizer
	if keyFn == nil {
		keyFn = links.NormalizeURL
	}
	key := keyFn(sourceURL)
	if key == "" {
		key = sourceURL
	}
	return d.Upsert(vault, key, path, ComputeFingerprint(text),
		category, subcategory, false)
}

// SetLocked mirrors set_locked (note_state.py:289-296): locked flag +
// updated_at refresh only.
func (d *DB) SetLocked(vault, sourceURL string, locked bool) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec(
		"UPDATE note_state SET locked=?, updated_at=? "+
			"WHERE vault=? AND source_url=?",
		boolToInt(locked), nowSeconds(), vault, sourceURL)
	return err
}

// LogCorrection mirrors log_correction (note_state.py:300-313): one
// append-only corrections_log row.
func (d *DB) LogCorrection(vault, sourceURL string,
	fromCategory, toCategory *string, fromPath, toPath, note string) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec(
		"INSERT INTO corrections_log "+
			"(ts, vault, source_url, from_category, to_category, "+
			" from_path, to_path, note) VALUES (?,?,?,?,?,?,?,?)",
		nowSeconds(), vault, sourceURL, nullable(fromCategory),
		nullable(toCategory), fromPath, toPath, note)
	return err
}

// Correction is one recent_corrections row (note_state.py:322-331).
type Correction struct {
	TS           string
	SourceURL    string
	FromCategory *string
	ToCategory   *string
	FromPath     string
	ToPath       string
	Note         string
}

// CorrectionHistory is one corrections_for row (note_state.py:333-343).
type CorrectionHistory struct {
	TS           string
	FromCategory *string
	ToCategory   *string
	Note         string
}

// CorrectionCount mirrors correction_count (note_state.py:315-320).
func (d *DB) CorrectionCount(vault string) (int, error) {
	d.RLock()
	defer d.RUnlock()
	var n int
	err := d.SQL().QueryRow(
		"SELECT COUNT(*) FROM corrections_log WHERE vault=?", vault).Scan(&n)
	return n, err
}

// RecentCorrections mirrors recent_corrections (note_state.py:322-331):
// newest first, limited.
func (d *DB) RecentCorrections(vault string, limit int) ([]Correction, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT ts, source_url, from_category, to_category, "+
			"from_path, to_path, note FROM corrections_log "+
			"WHERE vault=? ORDER BY id DESC LIMIT ?", vault, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Correction
	for rows.Next() {
		var c Correction
		var from, to sql.NullString
		if err := rows.Scan(&c.TS, &c.SourceURL, &from, &to, &c.FromPath,
			&c.ToPath, &c.Note); err != nil {
			return nil, err
		}
		c.FromCategory = stringPtr(from)
		c.ToCategory = stringPtr(to)
		out = append(out, c)
	}
	return out, rows.Err()
}

// CorrectionsFor mirrors corrections_for (note_state.py:333-343): the
// full history for ONE note, oldest first.
func (d *DB) CorrectionsFor(vault, sourceURL string) ([]CorrectionHistory, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT ts, from_category, to_category, note "+
			"FROM corrections_log WHERE vault=? AND source_url=? "+
			"ORDER BY id ASC", vault, sourceURL)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []CorrectionHistory
	for rows.Next() {
		var c CorrectionHistory
		var from, to sql.NullString
		if err := rows.Scan(&c.TS, &from, &to, &c.Note); err != nil {
			return nil, err
		}
		c.FromCategory = stringPtr(from)
		c.ToCategory = stringPtr(to)
		out = append(out, c)
	}
	return out, rows.Err()
}

// Dismiss mirrors dismiss (note_state.py:347-356): record a deleted
// note's URL — §4.4: never re-add it (INSERT OR REPLACE).
func (d *DB) Dismiss(vault, sourceURL, lastPath string) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec(
		"INSERT OR REPLACE INTO note_state_dismissed "+
			"(vault, source_url, last_path, dismissed_at) VALUES (?,?,?,?)",
		vault, sourceURL, lastPath, nowSeconds())
	return err
}

// IsDismissed mirrors is_dismissed (note_state.py:358-363).
func (d *DB) IsDismissed(vault, sourceURL string) (bool, error) {
	d.RLock()
	defer d.RUnlock()
	var one int
	err := d.SQL().QueryRow(
		"SELECT 1 FROM note_state_dismissed "+
			"WHERE vault=? AND source_url=?", vault, sourceURL).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

// DismissedCount mirrors dismissed_count (note_state.py:365-370).
func (d *DB) DismissedCount(vault string) (int, error) {
	d.RLock()
	defer d.RUnlock()
	var n int
	err := d.SQL().QueryRow(
		"SELECT COUNT(*) FROM note_state_dismissed WHERE vault=?",
		vault).Scan(&n)
	return n, err
}

// DismissedSet mirrors dismissed_set (note_state.py:372-379): all
// dismissed source URLs for one vault (the processing loops check this
// before doing any work).
func (d *DB) DismissedSet(vault string) (map[string]bool, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT source_url FROM note_state_dismissed WHERE vault=?", vault)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var u string
		if err := rows.Scan(&u); err != nil {
			return nil, err
		}
		out[u] = true
	}
	return out, rows.Err()
}

// ClearDismissed mirrors clear_dismissed (note_state.py:381-389):
// un-dismiss one URL; true when a row was removed.
func (d *DB) ClearDismissed(vault, sourceURL string) (bool, error) {
	d.Lock()
	defer d.Unlock()
	res, err := d.SQL().Exec(
		"DELETE FROM note_state_dismissed WHERE vault=? AND source_url=?",
		vault, sourceURL)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// -- reads ------------------------------------------------------------

// Row is one note_state row (note_state.py:399-429). Category and
// Subcategory are nil ⇔ SQL NULL (no category).
type Row struct {
	SourceURL   string
	Path        string
	Fingerprint string
	Category    *string
	Subcategory *string
	Locked      bool
	FirstSeen   string
	UpdatedAt   string
}

// Count mirrors count (note_state.py:393-397).
func (d *DB) Count(vault string) (int, error) {
	d.RLock()
	defer d.RUnlock()
	var n int
	err := d.SQL().QueryRow(
		"SELECT COUNT(*) FROM note_state WHERE vault=?", vault).Scan(&n)
	return n, err
}

// AllRows mirrors all_rows (note_state.py:399-411): source_url -> row
// for one vault (lookup semantics; the iteration order Python dicts carry
// lives in rowsOrdered, which classify uses).
func (d *DB) AllRows(vault string) (map[string]Row, error) {
	rows, err := d.rowsOrdered(vault)
	if err != nil {
		return nil, err
	}
	out := make(map[string]Row, len(rows))
	for _, r := range rows {
		out[r.SourceURL] = r
	}
	return out, nil
}

// rowsOrdered is the SELECT-order half of all_rows (no ORDER BY — rowid
// order, exactly Python's dict insertion order). Private: classify
// depends on the order; the public map API does not carry it.
func (d *DB) rowsOrdered(vault string) ([]Row, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT source_url, path, fingerprint, category, subcategory,"+
			" locked, first_seen, updated_at FROM note_state WHERE vault=?",
		vault)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	for rows.Next() {
		var r Row
		var category, subcategory sql.NullString
		var locked int64
		if err := rows.Scan(&r.SourceURL, &r.Path, &r.Fingerprint, &category,
			&subcategory, &locked, &r.FirstSeen, &r.UpdatedAt); err != nil {
			return nil, err
		}
		r.Category = stringPtr(category)
		r.Subcategory = stringPtr(subcategory)
		r.Locked = locked != 0
		out = append(out, r)
	}
	return out, rows.Err()
}

// RowFor mirrors row_for (note_state.py:413-429): the row for ONE note
// (identity = vault + source URL), or nil.
func (d *DB) RowFor(vault, sourceURL string) (*Row, error) {
	d.RLock()
	defer d.RUnlock()
	var r Row
	var category, subcategory sql.NullString
	var locked int64
	err := d.SQL().QueryRow(
		"SELECT source_url, path, fingerprint, category, subcategory,"+
			" locked, first_seen, updated_at FROM note_state"+
			" WHERE vault=? AND source_url=?", vault, sourceURL).
		Scan(&r.SourceURL, &r.Path, &r.Fingerprint, &category, &subcategory,
			&locked, &r.FirstSeen, &r.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Category = stringPtr(category)
	r.Subcategory = stringPtr(subcategory)
	r.Locked = locked != 0
	return &r, nil
}

// ---------------------------------------------------------------------------
// Baseline + change detection (note_state.py:443-578)
// ---------------------------------------------------------------------------

// RecordBaselineIfEmpty mirrors record_baseline_if_empty
// (note_state.py:443-474): the ONE-TIME silent baseline — records every
// managed note found by ScanVault ONLY when the vault has no rows yet.
// Returns the number recorded, or nil when a baseline already existed.
// Unmanaged notes are never adopted. NOTE (quirk, ported faithfully):
// like the reference, this walks with the DEFAULT (GitHub) resolver —
// the websites baseline also categorizes through CATEGORY_FOLDERS.
func RecordBaselineIfEmpty(vault, vaultPath, dbPath string, log LogFn) (*int, error) {
	db, err := New(dbPath)
	if err != nil {
		return nil, err
	}
	defer db.Close()
	n, err := db.Count(vault)
	if err != nil {
		return nil, err
	}
	if n > 0 {
		return nil, nil
	}
	var managed []Note
	for _, nt := range ScanVault(vaultPath, nil) {
		if !nt.Unmanaged {
			managed = append(managed, nt)
		}
	}
	for _, nt := range managed {
		content := ""
		if data, err := os.ReadFile(nt.Path); err == nil {
			content = decodeUTF8Replace(data)
		}
		if err := db.Upsert(vault, nt.SourceURL, nt.Path,
			ComputeFingerprint(content), nt.Category, nt.Subcategory,
			false); err != nil {
			return nil, err
		}
	}
	if log != nil && len(managed) > 0 {
		log(fmt.Sprintf("🗂️ Note-state baseline recorded: %d notes (%s vault)",
			len(managed), vault), "info")
	}
	count := len(managed)
	return &count, nil
}

// The §4.4 classification table (note_state.py:477-578).

type (
	// Moved — same source, different path (folder correction).
	Moved struct {
		SourceURL    string
		FromPath     string
		ToPath       string
		FromCategory *string
		ToCategory   *string
		ToFolder     string
	}
	// Edited — same path, content fingerprint changed (human edit).
	Edited struct {
		SourceURL string
		Path      string
	}
	// Deleted — in state, file gone (→ dismissed list).
	Deleted struct {
		SourceURL string
		LastPath  string
	}
	// Duplicate — two files on disk share one source (flag, touch none).
	Duplicate struct {
		SourceURL string
		Paths     []string
	}
	// UnmanagedNote — file without a source: line.
	UnmanagedNote struct {
		Path string
	}
	// UnmappedNote — note in a folder no category maps to.
	UnmappedNote struct {
		SourceURL string
		Path      string
		Folder    string
	}
	// Unknown — file with a source the app never recorded.
	Unknown struct {
		SourceURL string
		Path      string
	}
)

// Changes is the §4.4 table classify_changes returns (note_state.py:
// 574-578). Field order = Python's dict order (the dry-run log line and
// every report iterate it).
type Changes struct {
	Moved      []Moved
	Edited     []Edited
	Deleted    []Deleted
	Duplicates []Duplicate
	Unmanaged  []UnmanagedNote
	Unmapped   []UnmappedNote
	Unknown    []Unknown
}

// ClassifyChanges mirrors classify_changes (note_state.py:477-578): pure
// DETECTION — compares a vault on disk with the stored state and sorts
// every difference into the §4.4 table. tax non-nil switches the
// folder -> (category, subcategory) resolver to the taxonomy's own
// names (the websites vault); 'unmapped' then means "folder the taxonomy
// does not define".
func ClassifyChanges(vault, vaultPath, dbPath string,
	tax *taxonomy.Taxonomy) (Changes, error) {
	var resolver Resolver
	if tax != nil {
		resolver = TaxonomyResolver(tax)
	} else {
		resolver = FolderCategory
	}

	db, err := New(dbPath)
	if err != nil {
		return Changes{}, err
	}
	state, err := db.rowsOrdered(vault)
	closeErr := db.Close()
	if err != nil {
		return Changes{}, err
	}
	if closeErr != nil {
		return Changes{}, closeErr
	}

	disk := ScanVault(vaultPath, resolver)

	// by_source in scan (walk) order — the insertion order Python's
	// dict carries; classify iterates it.
	var order []string
	bySource := map[string][]Note{}
	for _, n := range disk {
		if n.Unmanaged {
			continue
		}
		if _, ok := bySource[n.SourceURL]; !ok {
			order = append(order, n.SourceURL)
		}
		bySource[n.SourceURL] = append(bySource[n.SourceURL], n)
	}
	stateMap := make(map[string]Row, len(state))
	for _, r := range state {
		stateMap[r.SourceURL] = r
	}

	var ch Changes
	seen := make(map[string]bool, len(order))

	for _, source := range order {
		files := bySource[source]
		seen[source] = true
		if len(files) > 1 {
			dup := Duplicate{SourceURL: source}
			for _, f := range files {
				dup.Paths = append(dup.Paths, f.Path)
			}
			ch.Duplicates = append(ch.Duplicates, dup)
			continue // §4.4: flag both, touch neither
		}
		note := files[0]
		row, hasRow := stateMap[source]
		switch {
		case !hasRow:
			ch.Unknown = append(ch.Unknown,
				Unknown{SourceURL: source, Path: note.Path})
		case helpers.Normcase(note.Path) != helpers.Normcase(row.Path):
			ch.Moved = append(ch.Moved, Moved{
				SourceURL:    source,
				FromPath:     row.Path,
				ToPath:       note.Path,
				FromCategory: row.Category,
				ToCategory:   note.Category,
				ToFolder:     note.Folder,
			})
		default:
			current := ""
			if data, err := os.ReadFile(note.Path); err == nil {
				current = ComputeFingerprint(decodeUTF8Replace(data))
			}
			if current != row.Fingerprint {
				ch.Edited = append(ch.Edited,
					Edited{SourceURL: source, Path: note.Path})
			}
		}
	}

	for _, r := range state { // SELECT (rowid) order — Python's dict order
		if !seen[r.SourceURL] {
			ch.Deleted = append(ch.Deleted,
				Deleted{SourceURL: r.SourceURL, LastPath: r.Path})
		}
	}

	for _, n := range disk {
		if n.Unmanaged {
			ch.Unmanaged = append(ch.Unmanaged, UnmanagedNote{Path: n.Path})
		}
	}

	// unmapped: a KNOWN note whose current folder matches no category and
	// is not a special one (root/_review). Duplicates/unmanaged already
	// listed are not double-reported.
	for _, source := range order {
		if len(bySource[source]) > 1 {
			continue
		}
		note := bySource[source][0]
		if note.Folder == "" || note.Folder == "_review" {
			continue
		}
		if tax != nil {
			category, sub := resolver(note.Folder)
			categoryStr, subStr := "", ""
			if category != nil {
				categoryStr = *category
			}
			if sub != nil {
				subStr = *sub
			}
			valid := tax.IsCategory(categoryStr) &&
				(subStr == "" || tax.IsSubcategoryOf(categoryStr, subStr))
			if !valid {
				ch.Unmapped = append(ch.Unmapped, UnmappedNote{
					SourceURL: source, Path: note.Path, Folder: note.Folder})
			}
		} else if _, ok := folderToCategory[note.Folder]; !ok {
			if note.Category != nil && *note.Category == note.Folder {
				ch.Unmapped = append(ch.Unmapped, UnmappedNote{
					SourceURL: source, Path: note.Path, Folder: note.Folder})
			}
		}
	}

	return ch, nil
}

// ---------------------------------------------------------------------------
// Phase 3 — acting on the §4.4 table (note_state.py:604-842)
// ---------------------------------------------------------------------------

// fmBounds mirrors _fm_bounds (note_state.py:604-611): (start, end) of
// the leading front-matter block — the interior line range [start, end)
// between the opening and closing '---' fences — or ok=false. An
// unterminated block is NEVER touched.
func fmBounds(lines []string) (start, end int, ok bool) {
	if len(lines) == 0 || helpers.Strip(lines[0]) != "---" {
		return 0, 0, false
	}
	for i := 1; i < len(lines); i++ {
		if helpers.Strip(lines[i]) == "---" {
			return 1, i, true
		}
	}
	return 0, 0, false
}

// FmReplaceField mirrors fm_replace_field (note_state.py:614-631):
// replace a `field:` line inside the front-matter (or insert the line
// just before the closing `---`). Returns (content, changed). A file
// without a well-formed front-matter block is returned UNCHANGED. The
// field match requires the line to START with the field name (a colon
// follows, optional whitespace between) — 'categories:' never matches
// a 'category' replacement.
func FmReplaceField(content, field, newLine string) (string, bool) {
	lines := strings.Split(content, "\n")
	start, end, ok := fmBounds(lines)
	if !ok {
		return content, false
	}
	pat := regexp.MustCompile(`^` + regexp.QuoteMeta(field) + pyS + "*:")
	for i := start; i < end; i++ {
		if pat.MatchString(lines[i]) {
			if lines[i] == newLine {
				return content, false
			}
			lines[i] = newLine
			return strings.Join(lines, "\n"), true
		}
	}
	// insert just before the closing fence (Python's lines.insert(end,
	// new_line))
	lines = append(lines, "")
	copy(lines[end+1:], lines[end:])
	lines[end] = newLine
	return strings.Join(lines, "\n"), true
}

// FmSwapTag mirrors fm_swap_tag (note_state.py:634-662): swap one token
// inside the `tags: [ ... ]` front-matter line — remove stripped, add
// PREPENDED, every other tag exactly as-is; a missing tags: line leaves
// the content untouched (no insertion).
func FmSwapTag(content, remove, add string) (string, bool) {
	remove = helpers.Strip(remove)
	add = helpers.Strip(add)
	if remove == "" && add == "" {
		return content, false
	}
	lines := strings.Split(content, "\n")
	start, end, ok := fmBounds(lines)
	if !ok {
		return content, false
	}
	for i := start; i < end; i++ {
		m := tagsRE.FindStringSubmatch(lines[i])
		if m == nil {
			continue
		}
		var items []string
		for _, t := range strings.Split(m[1], ",") {
			if helpers.Strip(t) == "" {
				continue
			}
			items = append(items, strings.Trim(helpers.Strip(t), "\"'"))
		}
		if remove != "" {
			kept := make([]string, 0, len(items))
			for _, it := range items {
				if it != remove {
					kept = append(kept, it)
				}
			}
			items = kept
		}
		hasAdd := false
		for _, it := range items {
			if it == add {
				hasAdd = true
			}
		}
		if add != "" && !hasAdd {
			items = append([]string{add}, items...)
		}
		newLine := "tags: []"
		if len(items) > 0 {
			newLine = "tags: [" + strings.Join(items, ", ") + "]"
		}
		if newLine == lines[i] {
			return content, false
		}
		lines[i] = newLine
		return strings.Join(lines, "\n"), true
	}
	return content, false
}

// movePairOrQ is the summary-line's None/”-to-'?' fold.
func movePairOrQ(s *string) string {
	if s == nil || *s == "" {
		return "?"
	}
	return *s
}

// MoveSummaryLines mirrors move_summary_lines (note_state.py:665-675):
// repeated moves grouped — "9 notes moved from X to Y" — most-common
// first (count descending; ties in first-encountered order, Python's
// stable-sort-over-dict semantics).
func MoveSummaryLines(corrections []CorrectionRecord) []string {
	type pair struct{ from, to string }
	var order []pair
	counts := map[pair]int{}
	for _, c := range corrections {
		k := pair{movePairOrQ(c.FromCategory), movePairOrQ(&c.ToCategory)}
		if _, ok := counts[k]; !ok {
			order = append(order, k)
		}
		counts[k]++
	}
	type stat struct {
		from, to string
		n        int
	}
	stats := make([]stat, 0, len(order))
	for _, k := range order {
		stats = append(stats, stat{from: k.from, to: k.to, n: counts[k]})
	}
	sort.SliceStable(stats, func(i, j int) bool { return stats[i].n > stats[j].n })
	lines := make([]string, 0, len(stats))
	for _, s := range stats {
		plural := "notes"
		if s.n == 1 {
			plural = "note"
		}
		lines = append(lines, fmt.Sprintf(
			"%d %s moved from %s to %s", s.n, plural, s.from, s.to))
	}
	return lines
}

// CorrectionRecord is one summary['corrections'] entry
// (note_state.py:779-781).
type CorrectionRecord struct {
	SourceURL    string
	FromCategory *string
	ToCategory   string
	ToPath       string
}

// Summary is apply_corrections' run report (note_state.py:699-700).
// MovedUnmapped is always present (0 when unused — Python's dict omits
// the key on the dry-run early return; unobservable here). Skipped is
// the run_start_check no-DB marker ('no note-state connection').
type Summary struct {
	MovedApplied  int
	MovedFailed   int
	MovedUnmapped int
	Dismissed     int
	Corrections   []CorrectionRecord
	Errors        []string
	DryRun        bool
	Skipped       string
}

func derefOrEmpty(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

// ApplyCorrections mirrors apply_corrections (note_state.py:678-796):
// ACT on classify_changes' results per the §4.4 table.
//
// Writes (all atomic / gated by core.dryrun): moved → the note's
// `category:` (+ `subcategory:` on websites) and `tags:` lines updated
// to match the new folder, a `category_locked: true` line added, the
// note_state row refreshed (new path + locked), one corrections_log
// row. Moving OUT of _review counts as a correction the same way.
// deleted → the URL goes on the dismissed list (never re-added).
// Everything else (edited/duplicates/unmanaged/unmapped/unknown) is
// REPORT-ONLY — the app never touches those files.
//
// A dry-run records no bookkeeping and edits nothing (the Phase 1
// contract: a rehearsal records nothing — detection results are logged
// by the CALLER).
func ApplyCorrections(vault, vaultPath string, changes Changes, db *DB,
	log LogFn, tax *taxonomy.Taxonomy) (Summary, error) {
	_ = vaultPath // scanned already (classify); Python keeps the param
	var summary Summary
	if log == nil {
		log = noopLog
	}
	if dryrun.IsEnabled() {
		summary.DryRun = true
		return summary, nil
	}

	// A move into a folder no taxonomy/category maps to is NOT a
	// correction (SPEC §4.4 "unmapped": keep it there, report it, never
	// invent a category) — those source URLs are skipped here.
	unmappedSources := make(map[string]bool, len(changes.Unmapped))
	for _, u := range changes.Unmapped {
		unmappedSources[u.SourceURL] = true
	}

	for _, mv := range changes.Moved {
		source := mv.SourceURL
		if unmappedSources[source] {
			summary.MovedUnmapped++
			log(fmt.Sprintf(
				"📍 Note moved into an unmapped folder — kept as-is, reported (%s)",
				source), "warning")
			continue
		}
		toPath := mv.ToPath
		fromCat := mv.FromCategory
		folder := mv.ToFolder
		var newCat, newSub string
		if tax != nil {
			var parts []string
			for _, p := range strings.Split(
				strings.ReplaceAll(strings.Trim(folder, "/"), "\\", "/"), "/") {
				if p != "" {
					parts = append(parts, p)
				}
			}
			if len(parts) > 0 {
				newCat = parts[0]
			}
			if len(parts) > 1 {
				newSub = strings.Join(parts[1:], "/")
			}
		} else {
			category, sub := FolderCategory(folder)
			newCat = derefOrEmpty(category)
			newSub = derefOrEmpty(sub)
		}
		data, err := os.ReadFile(toPath)
		if err != nil {
			summary.MovedFailed++
			summary.Errors = append(summary.Errors, fmt.Sprintf("%s: %s",
				toPath, helpers.PyOSErrorText(err, toPath)))
			continue
		}
		content := decodeUTF8Replace(data)

		changed := false
		if newCat != "" {
			var c bool
			content, c = FmReplaceField(content, "category", fmt.Sprintf(
				"category: %s", notebuilder.YamlScalar(newCat, notebuilder.MaxScalar)))
			changed = changed || c
			content, c = FmSwapTag(content, derefOrEmpty(fromCat), newCat)
			changed = changed || c
		}
		if vault == VaultWebsites || newSub != "" {
			// Website notes always carry the line (build_website_note);
			// GitHub notes only when a subcategory exists (nested folder).
			subLine := `""`
			if newSub != "" {
				subLine = notebuilder.YamlScalar(newSub, notebuilder.MaxScalar)
			}
			var c bool
			content, c = FmReplaceField(content, "subcategory",
				"subcategory: "+subLine)
			changed = changed || c
		}
		var c bool
		content, c = FmReplaceField(content, "category_locked",
			"category_locked: true")
		changed = changed || c

		var catPtr, subPtr *string
		if newCat != "" {
			catPtr = &newCat
		}
		if newSub != "" {
			subPtr = &newSub
		}
		if changed {
			// atomic_write_text is gated by dryrun: in a dry-run the edit
			// is recorded, not written (unreachable here — the early
			// return above — but the gate stays at the storage layer).
			if err := storage.AtomicWriteText(toPath, content); err != nil {
				summary.MovedFailed++
				summary.Errors = append(summary.Errors, fmt.Sprintf("%s: %s",
					toPath, helpers.PyOSErrorText(err, toPath)))
				continue
			}
		}
		if err := db.Upsert(vault, source, toPath, ComputeFingerprint(content),
			catPtr, subPtr, true); err != nil {
			return summary, err
		}
		toCat := newCat
		if err := db.LogCorrection(vault, source, fromCat, &toCat,
			mv.FromPath, toPath, ""); err != nil {
			return summary, err
		}

		summary.MovedApplied++
		summary.Corrections = append(summary.Corrections, CorrectionRecord{
			SourceURL:    source,
			FromCategory: fromCat,
			ToCategory:   newCat,
			ToPath:       toPath,
		})
		msg := fmt.Sprintf("📌 Correction: note moved to %s", newCat)
		if newSub != "" {
			msg += fmt.Sprintf(" / %s", newSub)
		}
		msg += fmt.Sprintf(
			" — front-matter updated, category locked (%s)", source)
		log(msg, "success")
	}

	for _, dl := range changes.Deleted {
		source := dl.SourceURL
		if source == "" {
			continue
		}
		dismissed, err := db.IsDismissed(vault, source)
		if err != nil {
			return summary, err
		}
		if !dismissed {
			if err := db.Dismiss(vault, source, dl.LastPath); err != nil {
				return summary, err
			}
			summary.Dismissed++
			log(fmt.Sprintf(
				"🗑️ Note deleted by you — URL dismissed, will not be re-added (%s)",
				source), "info")
		}
	}

	return summary, nil
}

// RunResult is run_start_check's return (note_state.py:805-812):
// baseline nil ⇔ a baseline already existed; Applied carries either the
// apply summary, the dry-run marker, or the no-DB skip marker.
type RunResult struct {
	Baseline *int
	Changes  Changes
	Applied  Summary
}

// RunStartCheck mirrors run_start_check (note_state.py:799-842): the
// start-of-run §4.4 pass for ONE vault — silent baseline when the vault
// has no rows yet, then detect + apply. Returns nil when the vault path
// is missing. tax is only for the websites vault (nil = GitHub
// CATEGORY_FOLDERS semantics). db is the caller's open connection (real
// runs); dbPath is the fallback used when db is nil (dry-runs open no
// connection — a rehearsal records nothing, it only detects and logs).
func RunStartCheck(vault, vaultPath string, db *DB, log LogFn,
	tax *taxonomy.Taxonomy, dbPath string) (*RunResult, error) {
	if vaultPath == "" {
		return nil, nil
	}
	if info, err := os.Stat(vaultPath); err != nil || !info.IsDir() {
		return nil, nil
	}
	if log == nil {
		log = noopLog
	}
	path := dbPath
	if db != nil {
		path = db.Path()
	}

	if dryrun.IsEnabled() {
		changes, err := ClassifyChanges(vault, vaultPath, path, tax)
		if err != nil {
			return nil, err
		}
		type bucket struct {
			name string
			n    int
		}
		var flagged []string
		for _, b := range []bucket{
			{"moved", len(changes.Moved)},
			{"edited", len(changes.Edited)},
			{"deleted", len(changes.Deleted)},
			{"duplicates", len(changes.Duplicates)},
			{"unmanaged", len(changes.Unmanaged)},
			{"unmapped", len(changes.Unmapped)},
			{"unknown", len(changes.Unknown)},
		} {
			if b.n > 0 {
				flagged = append(flagged, fmt.Sprintf("%s=%d", b.name, b.n))
			}
		}
		if len(flagged) > 0 {
			log(fmt.Sprintf(
				"🔍 [dry-run] note state (%s vault): %s — nothing applied, nothing recorded",
				vault, strings.Join(flagged, ", ")), "info")
		}
		return &RunResult{Changes: changes, Applied: Summary{DryRun: true}},
			nil
	}

	baseline, err := RecordBaselineIfEmpty(vault, vaultPath, path, log)
	if err != nil {
		return nil, err
	}
	changes, err := ClassifyChanges(vault, vaultPath, path, tax)
	if err != nil {
		return nil, err
	}
	if db == nil {
		return &RunResult{Baseline: baseline, Changes: changes,
			Applied: Summary{Skipped: "no note-state connection"}}, nil
	}
	applied, err := ApplyCorrections(vault, vaultPath, changes, db, log, tax)
	if err != nil {
		return nil, err
	}
	return &RunResult{Baseline: baseline, Changes: changes, Applied: applied},
		nil
}
