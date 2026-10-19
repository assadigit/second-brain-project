// Package store, file repos.go — the port of gui/cache_db.py's CacheDB
// (the GitHub-side SQLite sidecar) plus gui/dead_links.py's threshold
// delegate. Task T-24 (Phase 2); spec: spec-gui-cache-db.md §cache_db.py
// + §Delegates. gui/link_helpers.py's three one-liners are pure
// re-exports of core.links (NormalizeURL / CleanURL / ExtractGitHubURLs
// — already ported at T-11) and carry no Go code of their own.
//
// THE LOCKING LAW (store/doc.go, applied here): CacheDB embeds the
// shared store plumbing — ONE connection (MaxOpenConns(1)) opened with
// the ADR-0003 DSN pragmas, an embedded RWMutex that is NOT reentrant
// (public entry points acquire, private helpers run free), rows
// collected before any nested statement. close() is idempotent
// (inherited from store.DB).
//
// THE CAST READ LAW (ADR-0003 Am.1) — GENUINELY LIVE in this store:
// processed_repos.processed_at/last_checked, failed_repos.failed_at,
// decommissioned_repos.decommissioned_at and repos_settled.settled_at
// are TIMESTAMP-DECLARED columns (pinned by the T-20 schema golden),
// which the modernc driver converts to time.Time on read (a mattn-
// compatible heuristic Python's sqlite3 never had). Every ported query
// selecting one of those columns wraps it in CAST(col AS TEXT):
// GetDeadURLs and GetQuarantineStats select
// CAST(decommissioned_at AS TEXT). No other CacheDB method selects a
// timestamp-declared column (ProcessedRowFor reads repo_id/url/owner/
// repo_name/note_path/category; GetFailedURLs reads url/error; the
// purge compares last_checked inside SQL where no conversion happens).
//
// THE BEST-EFFORT LAW, ported method-by-method as the reference has
// it: most CacheDB methods swallow DB errors and answer a safe
// fallback (False / empty / 0 / nil — "never crash the batch on a DB
// error"; Record404 FAILS CLOSED, answering the threshold so the
// batch moves on). Those Go methods return no error. The methods the
// reference lets raise (AddProcessed, UpdateLastChecked, PurgeOld,
// ClearCache, RemoveEntry, IsDuplicate, GetNotePath, IsNoteValid,
// SettleRepos, SettleExistingRepos, and the constructor) return an
// error. A closed CacheDB answers every best-effort probe with its
// fallback — exactly like the reference's post-close except path
// (pinned by the behavior golden's fail-closed probe).
//
// THE DIVERGENCES from the Python law, deliberate and documented:
//
//   - checkpoints is NOT ported (dead schema: created, never used —
//     ADR-0003 consequence 2). Opening an old DB leaves the table in
//     place untouched; the T-20 schema golden records its DDL.
//   - failed_rows_truth' legacy row shapes (dict rows, None rows) are
//     structurally absent: Go's []FailedRow is typed — the observable
//     buckets over the representable inputs are identical (the pure
//     golden pins them).
//   - threshold=0 means "the default" exactly like Python's
//     `threshold or DEAD_LINK_THRESHOLD` (0 is falsy).
//   - The open-time migrations run inside the createTables sequence
//     (fail_count column add; notfound_strikes merge + DROP) —
//     best-effort like the reference, whose except-paths leave the
//     old shapes working.
//
// The dry-run seam stays at call sites exactly like Python
// (db_path=_dryrun.shadow_cache_path(...)): NewCacheDB receives an
// already-shadowed path and needs no dryrun knowledge.

package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/links"
)

// DeadLinkThreshold mirrors gui/dead_links.py's DEAD_LINK_THRESHOLD —
// the v0.08 404-quarantine threshold (owner spec: "after 2-3 tries
// across the sessions, system ignore those links").
const DeadLinkThreshold = 3

// ReposSettledMetaKey mirrors cache_db.py's REPOS_SETTLED_META_KEY —
// THE SETTLED REPOS LEDGER's one-time guard (the twin of the websites'
// websites_settled_at).
const ReposSettledMetaKey = "repos_settled_at"

// DeadLinkThresholdFromConfig mirrors dead_links.dead_link_threshold:
// the threshold is CONFIGURABLE (notfound_strike_threshold — the GUI
// spinbox, CLI --strikes N, config.json), clamped to >= 2; any error
// falls back to DeadLinkThreshold. cfg is the config map (the JSON
// shape); a nil map reads as empty (the reference's None case reads
// config.json from disk — a call-site convenience the typed Go callers
// replace by passing their map).
func DeadLinkThresholdFromConfig(cfg map[string]any) int {
	raw := any(DeadLinkThreshold)
	if v, ok := cfg["notfound_strike_threshold"]; ok {
		raw = v
	}
	switch x := raw.(type) {
	case nil:
		return DeadLinkThreshold
	case int:
		if x < 2 {
			return 2
		}
		return x
	case int64:
		if x < 2 {
			return 2
		}
		return int(x)
	case float64:
		// Python int(raw) truncates floats.
		i := int(x)
		if i < 2 {
			return 2
		}
		return i
	case json.Number:
		// Python int(raw) accepts integral strings AND truncates floats.
		if i, err := x.Int64(); err == nil {
			if i < 2 {
				return 2
			}
			return int(i)
		}
		if f, err := x.Float64(); err == nil {
			i := int(f)
			if i < 2 {
				return 2
			}
			return i
		}
		return DeadLinkThreshold
	case string:
		var i int
		if _, err := fmt.Sscanf(x, "%d", &i); err == nil {
			if i < 2 {
				return 2
			}
			return i
		}
		return DeadLinkThreshold
	default:
		return DeadLinkThreshold
	}
}

// FailedRow is one get_failed_urls row (the Python (url, error) tuple).
type FailedRow struct {
	URL   string
	Error string
}

// FailedRowsTruth mirrors cache_db.failed_rows_truth (v0.63.1) — THE
// TRUTH PASS over the retry queue's rows (pure, no DB, no files — the
// caller supplies the probes). Re-asks the only question that matters:
// is the note actually missing? githubHas is the optional
// url -> bool probe over the GitHub vault's index (nil = no probe
// available). Returns (stale, missing, websites): stale — github urls
// whose notes ARE in the vault; missing — github urls whose notes are
// genuinely absent (or unprovable — a missing probe keeps the old,
// loud behavior); websites — non-github urls, never the repos retry
// queue's business.
func FailedRowsTruth(rows []FailedRow,
	githubHas func(string) bool) (stale, missing, websites []string) {
	for _, row := range rows {
		url := strings.TrimSpace(row.URL)
		if url == "" {
			continue // a null-url row is nobody's retry
		}
		if len(links.ExtractGitHubURLs(url)) > 0 {
			if githubHas != nil && githubHas(url) {
				stale = append(stale, url)
			} else {
				missing = append(missing, url)
			}
		} else {
			websites = append(websites, url)
		}
	}
	return stale, missing, websites
}

// The adopt-as-is DDL (cache_db.py:150-288) — byte-identical to the
// reference (ADR-0003: adopt the existing database, no migrations
// beyond the reference's own). checkpoints is NOT created (dead
// schema — see the file doc).
const (
	ddlProcessedRepos = `CREATE TABLE IF NOT EXISTS processed_repos (
                repo_id INTEGER PRIMARY KEY,
                url TEXT,
                owner TEXT,
                repo_name TEXT,
                processed_at TIMESTAMP,
                last_checked TIMESTAMP,
                note_path TEXT,
                category TEXT
            )`
	ddlIdxProcessedReposURL = `CREATE INDEX IF NOT EXISTS idx_processed_repos_url
            ON processed_repos(url)`
	ddlIdxProcessedReposNotePath = `CREATE INDEX IF NOT EXISTS idx_processed_repos_note_path
            ON processed_repos(note_path)`
	ddlFailedRepos = `CREATE TABLE IF NOT EXISTS failed_repos (
                id INTEGER PRIMARY KEY AUTOINCREMENT,
                url TEXT,
                error TEXT,
                failed_at TIMESTAMP,
                retry_count INTEGER DEFAULT 0,
                resolved BOOLEAN DEFAULT 0
            )`
	ddlIdxFailedReposURL = `CREATE INDEX IF NOT EXISTS idx_failed_repos_url
            ON failed_repos(url)`
	ddlDecommissionedRepos = `CREATE TABLE IF NOT EXISTS decommissioned_repos (
                url TEXT PRIMARY KEY,
                reason TEXT,
                decommissioned_at TIMESTAMP,
                fail_count INTEGER NOT NULL DEFAULT 3
            )`
	ddlReposSettled = `CREATE TABLE IF NOT EXISTS repos_settled (
                url TEXT PRIMARY KEY,
                settled_at TIMESTAMP
            )`
	ddlCacheMeta = `CREATE TABLE IF NOT EXISTS cache_meta (
                key TEXT PRIMARY KEY,
                value TEXT
            )`
)

// nowAuto is the write shape law for this store's legacy tables:
// datetime.now().isoformat() (timespec='auto') — the fraction is
// emitted IFF the microsecond part is nonzero (processed_at,
// last_checked, failed_at, decommissioned_at).
func nowAuto() string {
	return PyISOFormat(time.Now())
}

// CacheDB is the GitHub-side SQLite sidecar: the shared store plumbing
// (ONE connection, the embedded not-reentrant RWMutex) plus the
// processed / failed / decommissioned / settled / meta ledgers. Same
// file family as the other stores (the call sites' resolved path);
// own connection.
type CacheDB struct {
	*DB
}

// ProcessedRepoRow mirrors processed_row_for's dict.
type ProcessedRepoRow struct {
	RepoID   int64
	URL      *string
	Owner    *string
	RepoName *string
	NotePath *string
	Category *string
}

// DeadURLRow is the quarantine-viewer row shape (url, reason,
// fail_count, decommissioned_at).
type DeadURLRow struct {
	URL              string
	Reason           *string
	FailCount        int64
	DecommissionedAt *string // read through CAST (the live CAST law)
}

// SettleReposReport mirrors settle_existing_repos's dict:
// {'settled': n, 'retries_cleared': m, 'already': bool}.
type SettleReposReport struct {
	Settled        int
	RetriesCleared int
	Already        bool
}

// NewCacheDB mirrors CacheDB.__init__ (cache_db.py:116-147): the
// literal default "cache.db" is anchored to APP_DIR (the v0.09.4
// GUI+CLI shared-cache fix — never the current working dir), the
// tables + indexes are created if absent, busy_timeout=30000, WAL +
// synchronous=NORMAL (CacheDB itself sets these — the unified Go DSN
// matches the reference exactly here), and the reference's own
// open-time migrations run (fail_count column add; notfound_strikes
// merge + DROP).
func NewCacheDB(dbPath string) (*CacheDB, error) {
	if dbPath == "cache.db" {
		dbPath = filepath.Join(constants.AppDir(), "cache.db")
	}
	base, err := Open(dbPath)
	if err != nil {
		return nil, err
	}
	c := &CacheDB{DB: base}
	if err := c.createTables(); err != nil {
		base.Close()
		return nil, err
	}
	return c, nil
}

func (c *CacheDB) createTables() error {
	c.Lock()
	defer c.Unlock()
	tx, err := c.SQL().Begin()
	if err != nil {
		return err
	}
	for _, stmt := range []string{
		ddlProcessedRepos, ddlIdxProcessedReposURL,
		ddlIdxProcessedReposNotePath, ddlFailedRepos, ddlIdxFailedReposURL,
		ddlDecommissionedRepos, ddlReposSettled, ddlCacheMeta,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	// The reference's best-effort migrations (cache_db.py:226-266) —
	// an except-path leaves the old shapes working.
	c.migrateFailCount()
	c.migrateNotfoundStrikes()
	return nil
}

// migrateFailCount mirrors the v0.08 migration: pre-v0.08 databases
// have no fail_count column; rows that already existed were 404'd in a
// previous session, so the column add's DEFAULT 3 treats them as fully
// confirmed dead.
func (c *CacheDB) migrateFailCount() {
	rows, err := c.SQL().Query(
		"PRAGMA table_info(decommissioned_repos)")
	if err != nil {
		return
	}
	var cols []string
	for rows.Next() {
		var cid int64
		var name, typ string
		var notnull int64
		var dflt any
		var pk int64
		if err := rows.Scan(&cid, &name, &typ, &notnull, &dflt, &pk); err != nil {
			rows.Close()
			return
		}
		cols = append(cols, name)
	}
	rows.Close()
	for _, col := range cols {
		if col == "fail_count" {
			return
		}
	}
	_, _ = c.SQL().Exec("ALTER TABLE decommissioned_repos " +
		"ADD COLUMN fail_count INTEGER NOT NULL DEFAULT 3")
}

// migrateNotfoundStrikes mirrors the v0.09 lineage merge: the v0.07
// notfound_strikes table migrates into decommissioned_repos.fail_count
// (keeping the HIGHER count when a URL exists in both), then the old
// table is dropped. Rows are collected before the nested writes (the
// package law).
func (c *CacheDB) migrateNotfoundStrikes() {
	var one int
	err := c.SQL().QueryRow(
		"SELECT COUNT(*) FROM sqlite_master " +
			"WHERE type='table' AND name='notfound_strikes'").Scan(&one)
	if err != nil || one == 0 {
		return
	}
	type strikeRow struct {
		url     string
		strikes sql.NullInt64
	}
	var strikes []strikeRow
	rows, err := c.SQL().Query(
		"SELECT url, strikes FROM notfound_strikes")
	if err != nil {
		return
	}
	for rows.Next() {
		var s strikeRow
		var u sql.NullString
		if err := rows.Scan(&u, &s.strikes); err != nil {
			rows.Close()
			return
		}
		s.url = u.String
		strikes = append(strikes, s)
	}
	rows.Close()
	now := nowAuto()
	for _, s := range strikes {
		if s.url == "" {
			continue
		}
		n := int64(1)
		if s.strikes.Valid && s.strikes.Int64 > 0 {
			n = s.strikes.Int64
		}
		_, _ = c.SQL().Exec(
			"INSERT INTO decommissioned_repos"+
				" (url, reason, decommissioned_at, fail_count)"+
				" VALUES (?, ?, ?, ?)"+
				" ON CONFLICT(url) DO UPDATE SET"+
				" fail_count = MAX(fail_count, excluded.fail_count)",
			links.NormalizeURL(s.url),
			"404 Not Found (migrated v0.07 strike counter)", now, n)
	}
	_, _ = c.SQL().Exec("DROP TABLE notfound_strikes")
}

// -- processed_repos ---------------------------------------------------

// IsDuplicate mirrors is_duplicate (raises in the reference).
func (c *CacheDB) IsDuplicate(repoID int64) (bool, error) {
	c.RLock()
	defer c.RUnlock()
	var one int
	err := c.SQL().QueryRow(
		"SELECT repo_id FROM processed_repos WHERE repo_id = ?",
		repoID).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// GetNotePath mirrors get_note_path (raises in the reference): the
// stored note_path for a repo, or nil if not cached.
func (c *CacheDB) GetNotePath(repoID int64) (*string, error) {
	c.RLock()
	defer c.RUnlock()
	var v sql.NullString
	err := c.SQL().QueryRow(
		"SELECT note_path FROM processed_repos WHERE repo_id = ?",
		repoID).Scan(&v)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return nullStr(v), nil
}

// RemoveEntry mirrors remove_entry (raises in the reference): remove a
// repo from the cache (used when its note file was deleted).
func (c *CacheDB) RemoveEntry(repoID int64) error {
	c.Lock()
	defer c.Unlock()
	_, err := c.SQL().Exec(
		"DELETE FROM processed_repos WHERE repo_id = ?", repoID)
	return err
}

// ProcessedRowFor mirrors processed_row_for (best-effort: nil on any
// error or when the URL was never processed — never raises): the
// processed_repos row for one URL, URL-keyed and normalized through
// the same link grammar every repos door uses.
func (c *CacheDB) ProcessedRowFor(url string) *ProcessedRepoRow {
	c.RLock()
	defer c.RUnlock()
	var r ProcessedRepoRow
	var u, owner, name, note, cat sql.NullString
	err := c.SQL().QueryRow(
		"SELECT repo_id, url, owner, repo_name, note_path, "+
			"category FROM processed_repos WHERE url = ?",
		links.NormalizeURL(url)).
		Scan(&r.RepoID, &u, &owner, &name, &note, &cat)
	if err != nil {
		return nil
	}
	r.URL, r.Owner = nullStr(u), nullStr(owner)
	r.RepoName, r.NotePath, r.Category = nullStr(name), nullStr(note),
		nullStr(cat)
	return &r
}

// ForgetProcessedNotePath deletes processed_repos rows by note_path
// (undo_last_batch's cache half — dashboard.py's
// "DELETE FROM processed_repos WHERE note_path = ?"). Returns the
// removed row count.
func (c *CacheDB) ForgetProcessedNotePath(notePath string) int {
	res, err := c.SQL().Exec(
		"DELETE FROM processed_repos WHERE note_path = ?", notePath)
	if err != nil {
		return 0
	}
	n, _ := res.RowsAffected()
	return int(n)
}

// ForgetProcessedURL mirrors forget_processed_url (best-effort): THE
// REPOS BANISHMENT's ledger forget — the processed_repos row for one
// URL is deleted. Returns True when a row was actually removed.
func (c *CacheDB) ForgetProcessedURL(url string) bool {
	c.Lock()
	defer c.Unlock()
	res, err := c.SQL().Exec(
		"DELETE FROM processed_repos WHERE url = ?",
		links.NormalizeURL(url))
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

// IsNoteValid mirrors is_note_valid (raises in the reference): does
// the cached repo's note file still exist on disk?
func (c *CacheDB) IsNoteValid(repoID int64) (bool, error) {
	notePath, err := c.GetNotePath(repoID)
	if err != nil {
		return false, err
	}
	if notePath == nil || *notePath == "" {
		return false, nil
	}
	fi, err := os.Stat(*notePath)
	return err == nil && !fi.IsDir(), nil
}

// AddProcessed mirrors add_processed (raises in the reference).
func (c *CacheDB) AddProcessed(repoID int64, url, owner, repoName,
	notePath, category string) error {
	now := nowAuto()
	c.Lock()
	defer c.Unlock()
	_, err := c.SQL().Exec(`
                INSERT OR REPLACE INTO processed_repos
                (repo_id, url, owner, repo_name, processed_at, last_checked, note_path, category)
                VALUES (?, ?, ?, ?, ?, ?, ?, ?)
            `, repoID, url, owner, repoName, now, now, notePath, category)
	return err
}

// UpdateLastChecked mirrors update_last_checked (raises).
func (c *CacheDB) UpdateLastChecked(repoID int64) error {
	now := nowAuto()
	c.Lock()
	defer c.Unlock()
	_, err := c.SQL().Exec(
		"UPDATE processed_repos SET last_checked = ? WHERE repo_id = ?",
		now, repoID)
	return err
}

// PurgeOld mirrors purge_old (raises): delete processed rows whose
// last_checked is older than months*30 days.
func (c *CacheDB) PurgeOld(months int) error {
	cutoff := PyISOFormat(
		time.Now().Add(-time.Duration(30*months) * 24 * time.Hour))
	c.Lock()
	defer c.Unlock()
	_, err := c.SQL().Exec(
		"DELETE FROM processed_repos WHERE last_checked < ?", cutoff)
	return err
}

// ClearCache mirrors clear_cache (raises).
func (c *CacheDB) ClearCache() error {
	c.Lock()
	defer c.Unlock()
	_, err := c.SQL().Exec("DELETE FROM processed_repos")
	return err
}

// -- failed_repos (the retry queue) -------------------------------------

// AddFailed mirrors add_failed (best-effort): record a failed repo so
// it can be retried later. An unresolved row for the URL is UPDATED in
// place (the fresher error wins — never stacks); resolved history is
// never touched; the row is keyed by the normalized URL so a later
// resolve under a trailing-slash / case-drifted spelling matches.
func (c *CacheDB) AddFailed(url, errMsg string) {
	now := nowAuto()
	norm := links.NormalizeURL(url)
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return
	}
	res, err := c.SQL().Exec(
		"UPDATE failed_repos SET error = ?, failed_at = ? "+
			"WHERE url = ? AND resolved = 0", errMsg, now, norm)
	if err != nil {
		return
	}
	if n, _ := res.RowsAffected(); n > 0 {
		return
	}
	_, _ = c.SQL().Exec(
		"INSERT INTO failed_repos "+
			"(url, error, failed_at, retry_count) "+
			"VALUES (?, ?, ?, 0)", norm, errMsg, now)
}

// GetFailedURLs mirrors get_failed_urls (best-effort): all unresolved
// (url, error) rows, most-recent first.
func (c *CacheDB) GetFailedURLs() []FailedRow {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return nil
	}
	rows, err := c.SQL().Query(
		"SELECT url, error FROM failed_repos WHERE resolved = 0" +
			" ORDER BY failed_at DESC")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []FailedRow
	for rows.Next() {
		var u, e sql.NullString
		if err := rows.Scan(&u, &e); err != nil {
			return nil
		}
		out = append(out, FailedRow{URL: u.String, Error: e.String})
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return out
}

// MarkFailedResolved mirrors mark_failed_resolved (best-effort): mark
// a previously-failed URL as resolved — matching BOTH spellings (the
// normalized key AND the raw URL, so legacy rows clear too).
func (c *CacheDB) MarkFailedResolved(url string) {
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return
	}
	_, _ = c.SQL().Exec(
		"UPDATE failed_repos SET resolved = 1 "+
			"WHERE url = ? OR url = ?", url, links.NormalizeURL(url))
}

// ResolveFailedURLs mirrors resolve_failed_urls (best-effort): bulk
// MarkFailedResolved, each URL under both spellings. Returns how many
// rows actually flipped (0 when nothing matched — the caller may stay
// quiet).
func (c *CacheDB) ResolveFailedURLs(urls []string) int {
	flipped := 0
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return 0
	}
	for _, url := range urls {
		res, err := c.SQL().Exec(
			"UPDATE failed_repos SET resolved = 1 "+
				"WHERE (url = ? OR url = ?) AND resolved = 0",
			url, links.NormalizeURL(url))
		if err != nil {
			continue
		}
		if n, err := res.RowsAffected(); err == nil {
			flipped += int(n)
		}
	}
	return flipped
}

// GetFailedCount mirrors get_failed_count (best-effort): count of
// unresolved failed URLs.
func (c *CacheDB) GetFailedCount() int {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return 0
	}
	var n int
	if err := c.SQL().QueryRow(
		"SELECT COUNT(*) FROM failed_repos WHERE resolved = 0").
		Scan(&n); err != nil {
		return 0
	}
	return n
}

// -- decommissioned_repos (the 404 quarantine) --------------------------

// Decommission mirrors decommission (best-effort; the legacy API —
// cloud sync / older call sites): mark a URL as decommissioned
// (permanently skipped), fully confirmed.
func (c *CacheDB) Decommission(url, reason string) {
	now := nowAuto()
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return
	}
	_, _ = c.SQL().Exec(
		"INSERT OR REPLACE INTO decommissioned_repos"+
			" (url, reason, decommissioned_at, fail_count)"+
			" VALUES (?, ?, ?, ?)",
		links.NormalizeURL(url), reason, now, DeadLinkThreshold)
}

// Record404 mirrors record_404 (best-effort, FAILS CLOSED): count one
// 404 attempt for a URL (across sessions). Returns the NEW consecutive
// failure count (1 on first sighting). When the count reaches the
// threshold the link is CONFIRMED DEAD and every input path skips it
// silently. On a DB error the answer is the THRESHOLD (fail closed:
// treat as confirmed so the batch still moves on).
func (c *CacheDB) Record404(url, reason string) int {
	now := nowAuto()
	norm := links.NormalizeURL(url)
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return DeadLinkThreshold // fail closed
	}
	var fc sql.NullInt64
	err := c.SQL().QueryRow(
		"SELECT fail_count FROM decommissioned_repos WHERE url = ?",
		norm).Scan(&fc)
	if err != nil && err != sql.ErrNoRows {
		return DeadLinkThreshold // fail closed
	}
	count := int64(1)
	if err == nil && fc.Valid && fc.Int64 > 0 {
		count = fc.Int64 + 1
	}
	if _, err := c.SQL().Exec(
		"INSERT OR REPLACE INTO decommissioned_repos"+
			" (url, reason, decommissioned_at, fail_count)"+
			" VALUES (?, ?, ?, ?)", norm, reason, now, count); err != nil {
		return DeadLinkThreshold // fail closed
	}
	return int(count)
}

// ConfirmDead mirrors confirm_dead (best-effort): immediately mark a
// URL CONFIRMED dead (fail_count set to at least the threshold — never
// lowered). threshold 0 means the default.
func (c *CacheDB) ConfirmDead(url, reason string, threshold int) {
	thr := threshold
	if thr == 0 {
		thr = DeadLinkThreshold
	}
	now := nowAuto()
	norm := links.NormalizeURL(url)
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return
	}
	var fc sql.NullInt64
	err := c.SQL().QueryRow(
		"SELECT fail_count FROM decommissioned_repos WHERE url = ?",
		norm).Scan(&fc)
	if err != nil && err != sql.ErrNoRows {
		return
	}
	count := int64(thr)
	if err == nil && fc.Valid && fc.Int64 > count {
		count = fc.Int64
	}
	_, _ = c.SQL().Exec(
		"INSERT OR REPLACE INTO decommissioned_repos"+
			" (url, reason, decommissioned_at, fail_count)"+
			" VALUES (?, ?, ?, ?)", norm, reason, now, count)
}

// GetUnconfirmed404s mirrors get_unconfirmed_404s (best-effort):
// [(url, fail_count)] for rows that 404'd at least once but are not
// yet confirmed (1 <= fail_count < threshold), most-strikes first.

// UnconfirmedRow is one get_unconfirmed_404s row (url, fail_count).
type UnconfirmedRow struct {
	URL       string
	FailCount int64
}

func (c *CacheDB) GetUnconfirmed404s(threshold int) []UnconfirmedRow {
	thr := threshold
	if thr == 0 {
		thr = DeadLinkThreshold
	}
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return nil
	}
	rows, err := c.SQL().Query(
		"SELECT url, fail_count FROM decommissioned_repos "+
			"WHERE fail_count >= 1 AND fail_count < ? "+
			"ORDER BY fail_count DESC, url", thr)
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []UnconfirmedRow
	for rows.Next() {
		var u sql.NullString
		var n int64
		if err := rows.Scan(&u, &n); err != nil {
			return nil
		}
		out = append(out, UnconfirmedRow{URL: u.String, FailCount: n})
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return out
}

// IsDeadLink mirrors is_dead_link (best-effort): True when the URL is
// CONFIRMED dead (fail_count >= threshold). Unconfirmed entries (1-2
// attempts) return False so they get their remaining attempts.
func (c *CacheDB) IsDeadLink(url string, threshold int) bool {
	thr := threshold
	if thr == 0 {
		thr = DeadLinkThreshold
	}
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return false
	}
	var fc sql.NullInt64
	err := c.SQL().QueryRow(
		"SELECT fail_count FROM decommissioned_repos WHERE url = ?",
		links.NormalizeURL(url)).Scan(&fc)
	return err == nil && fc.Valid && fc.Int64 >= int64(thr)
}

// GetDeadURLSet mirrors get_dead_url_set (best-effort): set of
// CONFIRMED-dead normalized URLs. Loaded ONCE per batch/queue-check
// and tested with `NormalizeURL(url) in dead` — no per-URL queries.
func (c *CacheDB) GetDeadURLSet(threshold int) map[string]bool {
	thr := threshold
	if thr == 0 {
		thr = DeadLinkThreshold
	}
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return map[string]bool{}
	}
	rows, err := c.SQL().Query(
		"SELECT url FROM decommissioned_repos WHERE fail_count >= ?",
		thr)
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var u sql.NullString
		if err := rows.Scan(&u); err != nil {
			return map[string]bool{}
		}
		out[u.String] = true
	}
	return out
}

// GetDeadURLs mirrors get_dead_urls (best-effort): confirmed-dead
// rows as (url, reason, fail_count, decommissioned_at), oldest first —
// the quarantine viewer / CLI listing. decommissioned_at is read
// through CAST(col AS TEXT): THE CAST READ LAW is live here (the
// TIMESTAMP-declared column would otherwise come back as a re-serialized
// time.Time, not Python's raw str).
func (c *CacheDB) GetDeadURLs(threshold int) []DeadURLRow {
	thr := threshold
	if thr == 0 {
		thr = DeadLinkThreshold
	}
	return c.queryQuarantine(
		"WHERE fail_count >= ?", thr, "ORDER BY decommissioned_at ASC")
}

// GetQuarantineStats mirrors get_quarantine_stats (best-effort): EVERY
// quarantine row (confirmed AND in-progress attempts), most-strikes
// first. decommissioned_at is read through CAST (the live CAST law).
func (c *CacheDB) GetQuarantineStats() []DeadURLRow {
	return c.queryQuarantine("", 0,
		"ORDER BY fail_count DESC, decommissioned_at DESC")
}

func (c *CacheDB) queryQuarantine(where string, thr int,
	orderBy string) []DeadURLRow {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return nil
	}
	q := "SELECT url, reason, fail_count, " +
		"CAST(decommissioned_at AS TEXT) FROM decommissioned_repos "
	if where != "" {
		q += where + " "
	}
	q += orderBy
	var rows *sql.Rows
	var err error
	if where != "" {
		rows, err = c.SQL().Query(q, thr)
	} else {
		rows, err = c.SQL().Query(q)
	}
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []DeadURLRow
	for rows.Next() {
		var r DeadURLRow
		var reason, at sql.NullString
		if err := rows.Scan(&r.URL, &reason, &r.FailCount, &at); err != nil {
			return nil
		}
		r.Reason = nullStr(reason)
		r.DecommissionedAt = nullStr(at)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil
	}
	return out
}

// ResetDeadLinks mirrors reset_dead_links (best-effort): clear the 404
// quarantine — the whole table when url is empty, or one URL. For
// false positives (a repo that went PRIVATE reads as 404 to an
// unauthorized token). v0.63.3 — the ♻️ door back through THE SETTLED
// REPOS LEDGER: removing a quarantine row WITHOUT un-settling it would
// be a silent no-op under the owner's law, so the reset removes BOTH
// marks (and ONLY for the URLs whose quarantine row is actually going).
// Returns how many entries were removed.
func (c *CacheDB) ResetDeadLinks(url string) int {
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return 0
	}
	if url != "" {
		norm := links.NormalizeURL(url)
		res, err := c.SQL().Exec(
			"DELETE FROM decommissioned_repos WHERE url = ?", norm)
		if err != nil {
			return 0
		}
		removed, _ := res.RowsAffected()
		if removed > 0 {
			_, _ = c.SQL().Exec(
				"DELETE FROM repos_settled WHERE url = ?", norm)
		}
		return int(removed)
	}
	// whole table — collect the revived first (the package law).
	var revived []string
	rows, err := c.SQL().Query(
		"SELECT url FROM decommissioned_repos")
	if err != nil {
		return 0
	}
	for rows.Next() {
		var u sql.NullString
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			return 0
		}
		revived = append(revived, u.String)
	}
	rows.Close()
	res, err := c.SQL().Exec("DELETE FROM decommissioned_repos")
	if err != nil {
		return 0
	}
	removed, _ := res.RowsAffected()
	if len(revived) > 0 {
		for _, u := range revived {
			_, _ = c.SQL().Exec(
				"DELETE FROM repos_settled WHERE url = ?", u)
		}
	}
	return int(removed)
}

// IsDecommissioned mirrors is_decommissioned (best-effort): any entry,
// confirmed or not (legacy callers). Prefer IsDeadLink /
// GetDeadURLSet for batch filtering.
func (c *CacheDB) IsDecommissioned(url string) bool {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return false
	}
	var one int
	err := c.SQL().QueryRow(
		"SELECT 1 FROM decommissioned_repos WHERE url = ?",
		links.NormalizeURL(url)).Scan(&one)
	return err == nil
}

// GetAllDecommissioned mirrors get_all_decommissioned (best-effort):
// all decommissioned (url, reason) rows — the legacy tuple shape.
func (c *CacheDB) GetAllDecommissioned() []FailedRow {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return nil
	}
	rows, err := c.SQL().Query(
		"SELECT url, reason FROM decommissioned_repos")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []FailedRow
	for rows.Next() {
		var u, r sql.NullString
		if err := rows.Scan(&u, &r); err != nil {
			return nil
		}
		out = append(out, FailedRow{URL: u.String, Error: r.String})
	}
	return out
}

// GetAllProcessedURLs mirrors get_all_processed_urls (best-effort).
func (c *CacheDB) GetAllProcessedURLs() []string {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return nil
	}
	rows, err := c.SQL().Query("SELECT url FROM processed_repos")
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []string
	for rows.Next() {
		var u sql.NullString
		if err := rows.Scan(&u); err != nil {
			return nil
		}
		out = append(out, u.String)
	}
	return out
}

// -- v0.63.3 THE SETTLED REPOS LEDGER -----------------------------------

// GetMeta mirrors get_meta (best-effort): cache_meta read.
func (c *CacheDB) GetMeta(key string) *string {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return nil
	}
	var v sql.NullString
	err := c.SQL().QueryRow(
		"SELECT value FROM cache_meta WHERE key=?", key).Scan(&v)
	if err != nil {
		return nil
	}
	s := v.String
	return &s
}

// SetMeta mirrors set_meta (best-effort: a meta write never breaks a
// caller).
func (c *CacheDB) SetMeta(key, value string) {
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return
	}
	_, _ = c.SQL().Exec(
		"INSERT OR REPLACE INTO cache_meta (key, value) VALUES (?,?)",
		key, value)
}

// IsRepoSettled mirrors is_repo_settled (best-effort).
func (c *CacheDB) IsRepoSettled(url string) bool {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return false
	}
	var one int
	err := c.SQL().QueryRow(
		"SELECT 1 FROM repos_settled WHERE url=?",
		links.NormalizeURL(url)).Scan(&one)
	return err == nil
}

// SettleRepos mirrors settle_repos (raises in the reference — the one
// CacheDB write method without a try): add repos to the settled ledger
// (idempotent; the owner's verdict doors use this). Returns how many
// rows were actually inserted. The commit is UNCONDITIONAL (the v0.65
// write-transaction law: a skipped commit left cache.db write-locked
// for every OTHER connection).
func (c *CacheDB) SettleRepos(urls []string) (int, error) {
	nowISO := PyISOFormatSeconds(time.Now())
	added := 0
	c.Lock()
	defer c.Unlock()
	tx, err := c.SQL().Begin()
	if err != nil {
		return 0, err
	}
	for _, u := range urls {
		if u == "" {
			continue
		}
		res, err := tx.Exec(
			"INSERT OR IGNORE INTO repos_settled (url, settled_at)"+
				" VALUES (?,?)", links.NormalizeURL(u), nowISO)
		if err != nil {
			_ = tx.Rollback()
			return 0, err
		}
		if n, err := res.RowsAffected(); err == nil {
			added += int(n)
		}
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return added, nil
}

// UnsettleRepo mirrors unsettle_repo (best-effort): the ♻️ door back —
// remove one repo from the settled ledger. Returns True when a settled
// row was actually removed.
func (c *CacheDB) UnsettleRepo(url string) bool {
	c.Lock()
	defer c.Unlock()
	if c.SQL() == nil {
		return false
	}
	res, err := c.SQL().Exec(
		"DELETE FROM repos_settled WHERE url=?",
		links.NormalizeURL(url))
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

// SettledReposCount mirrors settled_repos_count (best-effort): the
// ledger's size (the honest settlement line).
func (c *CacheDB) SettledReposCount() int {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return 0
	}
	var n int
	if err := c.SQL().QueryRow(
		"SELECT COUNT(*) FROM repos_settled").Scan(&n); err != nil {
		return 0
	}
	return n
}

// GetSettledRepoSet mirrors get_settled_repo_set (best-effort): the
// settled set, loaded ONCE per queue-check/batch and tested with
// `NormalizeURL(url) in settled` — no per-URL queries.
func (c *CacheDB) GetSettledRepoSet() map[string]bool {
	c.RLock()
	defer c.RUnlock()
	if c.SQL() == nil {
		return map[string]bool{}
	}
	rows, err := c.SQL().Query("SELECT url FROM repos_settled")
	if err != nil {
		return map[string]bool{}
	}
	defer rows.Close()
	out := map[string]bool{}
	for rows.Next() {
		var u sql.NullString
		if err := rows.Scan(&u); err != nil {
			return map[string]bool{}
		}
		out[u.String] = true
	}
	return out
}

// ReposLedgerSettled mirrors repos_ledger_settled (best-effort): has
// THE SETTLEMENT run on this machine yet?
func (c *CacheDB) ReposLedgerSettled() bool {
	return c.GetMeta(ReposSettledMetaKey) != nil
}

// SettleExistingRepos mirrors settle_existing_repos (raises in the
// reference): THE SETTLEMENT, one time, on the owner's word. Seeds the
// ledger from every URL the cache knows (processed + failed +
// decommissioned) plus the extraURLs the queue-door caller adds;
// the retry queue is bulk-cleared. Guarded by the ReposSettledMetaKey:
// a second call is a no-op. NOTE the honest-count quirk (the same one
// as the websites twin): Settled counts the unique normalized URLs
// this call considered — NOT the rows actually inserted.
func (c *CacheDB) SettleExistingRepos(extraURLs []string) (
	SettleReposReport, error) {
	c.RLock()
	var guard sql.NullString
	err := c.SQL().QueryRow(
		"SELECT value FROM cache_meta WHERE key=?",
		ReposSettledMetaKey).Scan(&guard)
	c.RUnlock()
	if err != nil && err != sql.ErrNoRows {
		return SettleReposReport{}, err
	}
	if err == nil {
		return SettleReposReport{Settled: 0, RetriesCleared: 0,
			Already: true}, nil
	}

	nowISO := PyISOFormatSeconds(time.Now())
	c.Lock()
	defer c.Unlock()
	tx, err := c.SQL().Begin()
	if err != nil {
		return SettleReposReport{}, err
	}
	var urls []string
	for _, q := range []string{
		"SELECT url FROM processed_repos",
		"SELECT url FROM failed_repos",
		"SELECT url FROM decommissioned_repos",
	} {
		rows, err := tx.Query(q)
		if err != nil {
			_ = tx.Rollback()
			return SettleReposReport{}, err
		}
		for rows.Next() {
			var u sql.NullString
			if err := rows.Scan(&u); err != nil {
				rows.Close()
				_ = tx.Rollback()
				return SettleReposReport{}, err
			}
			urls = append(urls, u.String)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			_ = tx.Rollback()
			return SettleReposReport{}, err
		}
	}
	urls = append(urls, extraURLs...)
	seen := map[string]bool{}
	var unique []string
	for _, u := range urls {
		norm := ""
		if u != "" {
			norm = links.NormalizeURL(u)
		}
		if norm == "" || seen[norm] {
			continue
		}
		seen[norm] = true
		unique = append(unique, norm)
	}
	for _, u := range unique {
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO repos_settled (url, settled_at)"+
				" VALUES (?,?)", u, nowISO); err != nil {
			_ = tx.Rollback()
			return SettleReposReport{}, err
		}
	}
	res, err := tx.Exec("DELETE FROM failed_repos")
	if err != nil {
		_ = tx.Rollback()
		return SettleReposReport{}, err
	}
	cleared, _ := res.RowsAffected()
	if _, err := tx.Exec(
		"INSERT OR REPLACE INTO cache_meta (key, value) VALUES (?,?)",
		ReposSettledMetaKey, nowISO); err != nil {
		_ = tx.Rollback()
		return SettleReposReport{}, err
	}
	if err := tx.Commit(); err != nil {
		return SettleReposReport{}, err
	}
	return SettleReposReport{Settled: len(unique),
		RetriesCleared: int(cleared), Already: false}, nil
}

// nullStr converts a scanned nullable column into *string (nil ⇔ NULL).
func nullStr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}
