package syncreview

// ledger.go — the DeferredLog: the permanent record of every review.
// SQLite (the store.Open law: busy_timeout 30000 + WAL + synchronous
// NORMAL via the DSN, exactly one connection), every statement under
// the ledger's own mutex. NEVER deletes a row: status changes only.
// close() is idempotent.

import (
	"database/sql"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/store"
)

// LedgerFilename mirrors LEDGER_FILENAME: the ledger's default
// filename (the app's data folder, beside cache.db — APP_DIR).
const LedgerFilename = "sync_review.db"

// Row status vocabulary: ACTIVE (the gap is explained), RESOLVED (the
// item later appeared in the target — kept forever), REOPENED (the
// owner withdrew the explanation — the item is back in Differences).
const (
	StatusActive   = "active"
	StatusResolved = "resolved"
	StatusReopened = "reopened"
)

// LedgerDefaultPath mirrors ledger_default_path: the ledger's home is
// the app's data folder (APP_DIR), the same folder cache.db lives in.
func LedgerDefaultPath(appDir string) string {
	base := appDir
	if base == "" {
		base = constants.AppDir()
	}
	return filepath.Join(base, LedgerFilename)
}

// nowISOSeconds mirrors datetime.now(timezone.utc).isoformat(
// timespec='seconds') — the ledger's own stamp shape.
func nowISOSeconds() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05+00:00")
}

// LedgerRow is one ledger record (the log tab's row shape).
type LedgerRow struct {
	Stage        string `json:"stage"`
	ItemID       string `json:"item_id"`
	DisplayName  string `json:"display_name"`
	Reason       string `json:"reason"`
	Note         string `json:"note"`
	Status       string `json:"status"`
	DateReviewed string `json:"date_reviewed"`
	DateResolved string `json:"date_resolved"`
	ReopenCount  int    `json:"reopen_count"`
}

// DeferredLog is the local SQLite ledger that remembers every review
// so the same gap never cries twice. The zero value is not usable;
// OpenLedger constructs it.
type DeferredLog struct {
	mu   sync.Mutex
	db   *store.DB
	path string
}

// OpenLedger opens (creating if absent) the ledger at path. An empty
// path rides LedgerDefaultPath(""). Never raises: a bad path answers
// an error the callers treat as the honest zero (the reference wraps
// every door in sqlite3.Error guards).
func OpenLedger(path string) (*DeferredLog, error) {
	if path == "" {
		path = LedgerDefaultPath("")
	}
	db, err := store.Open(path)
	if err != nil {
		return nil, err
	}
	l := &DeferredLog{db: db, path: path}
	if err := l.createTables(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return l, nil
}

// createTables mirrors _create_tables (the schema + the stable-ID
// law's index).
func (l *DeferredLog) createTables() error {
	_, err := l.db.SQL().Exec(`
                CREATE TABLE IF NOT EXISTS sync_deferrals (
                        id INTEGER PRIMARY KEY AUTOINCREMENT,
                        stage TEXT NOT NULL,
                        item_id TEXT NOT NULL,
                        display_name TEXT DEFAULT '',
                        reason TEXT NOT NULL,
                        note TEXT DEFAULT '',
                        status TEXT NOT NULL DEFAULT 'active',
                        date_reviewed TEXT NOT NULL,
                        date_resolved TEXT,
                        reopen_count INTEGER NOT NULL DEFAULT 0,
                        UNIQUE(stage, item_id)
                )`)
	if err != nil {
		return err
	}
	_, err = l.db.SQL().Exec(`
                CREATE INDEX IF NOT EXISTS idx_sync_deferrals_stage_item
                ON sync_deferrals(stage, item_id)`)
	return err
}

// Path answers the ledger's own path.
func (l *DeferredLog) Path() string { return l.path }

// Mark mirrors mark: every (stage, item) row becomes ACTIVE with this
// reason/note. A re-mark updates the SAME row in place (the item
// stays reviewed even if its note is edited later — the row's
// date_reviewed refreshes). Never deletes. Returns the number of
// items marked. Never raises.
func (l *DeferredLog) Mark(stage string, itemIDs []string, reason,
	note string, displayNames map[string]string, now string) int {
	if !ValidReason(reason) {
		return 0
	}
	if displayNames == nil {
		displayNames = map[string]string{}
	}
	if now == "" {
		now = nowISOSeconds()
	}
	count := 0
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, item := range itemIDs {
		if item == "" {
			continue
		}
		name := displayNames[item]
		_, err := l.db.SQL().Exec(`
                        INSERT INTO sync_deferrals
                                (stage, item_id, display_name, reason, note,
                                 status, date_reviewed, reopen_count)
                        VALUES (?, ?, ?, ?, ?, 'active', ?, 0)
                        ON CONFLICT(stage, item_id) DO UPDATE SET
                                display_name = excluded.display_name,
                                reason = excluded.reason,
                                note = excluded.note,
                                status = 'active',
                                date_reviewed = excluded.date_reviewed,
                                date_resolved = NULL`,
			stage, item, name, reason, note, now)
		if err != nil {
			continue // sqlite3.Error: continue — never raises
		}
		count++
	}
	return count
}

// Reopen mirrors reopen: the row's status becomes REOPENED (it
// returns to the Differences tab; the reopen is logged in the row's
// reopen_count). Only an ACTIVE row reopens. True when it happened.
// Never deletes, never raises.
func (l *DeferredLog) Reopen(stage, itemID, now string) bool {
	if now == "" {
		now = nowISOSeconds()
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	res, err := l.db.SQL().Exec(`
                UPDATE sync_deferrals
                SET status = 'reopened', reopen_count = reopen_count + 1
                WHERE stage = ? AND item_id = ? AND status = 'active'`,
		stage, itemID)
	if err != nil {
		return false
	}
	n, _ := res.RowsAffected()
	return n > 0
}

// ResolveSeen mirrors resolve_seen (target semantics): every ACTIVE
// row whose item now appears in the target stage becomes RESOLVED (row
// kept forever). A RESOLVED row whose item left the target again
// flips back to ACTIVE — the deferral still stands. Returns the
// number resolved. Never raises.
func (l *DeferredLog) ResolveSeen(stage string, targetIDs []string,
	now string) int {
	return l.resolve(stage, targetIDs, true, now)
}

// ResolveGone mirrors resolve_gone (gap semantics — the queue and
// backup stages): every ACTIVE row whose item is NO LONGER in the
// stage's current gap list becomes RESOLVED. A RESOLVED row whose
// item re-enters the gap flips back to ACTIVE. Returns the number
// resolved. Never raises.
func (l *DeferredLog) ResolveGone(stage string, presentIDs []string,
	now string) int {
	return l.resolve(stage, presentIDs, false, now)
}

func (l *DeferredLog) resolve(stage string, ids []string,
	presentResolves bool, now string) int {
	if now == "" {
		now = nowISOSeconds()
	}
	present := map[string]bool{}
	for _, id := range ids {
		if id != "" {
			present[id] = true
		}
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	rows, err := l.db.SQL().Query(
		`SELECT item_id, status FROM sync_deferrals WHERE stage = ?`, stage)
	if err != nil {
		return 0
	}
	type rowState struct{ itemID, status string }
	var all []rowState
	for rows.Next() {
		var r rowState
		if err := rows.Scan(&r.itemID, &r.status); err != nil {
			continue
		}
		all = append(all, r)
	}
	rows.Close()
	resolved := 0
	for _, r := range all {
		inPresent := present[r.itemID]
		var resolves, returns bool
		if presentResolves {
			resolves, returns = inPresent, !inPresent
		} else {
			resolves, returns = !inPresent, inPresent
		}
		if r.status == StatusActive && resolves {
			if _, err := l.db.SQL().Exec(`
                                UPDATE sync_deferrals
                                SET status = 'resolved', date_resolved = ?
                                WHERE stage = ? AND item_id = ? AND status = 'active'`,
				now, stage, r.itemID); err == nil {
				resolved++
			}
		} else if r.status == StatusResolved && returns {
			_, _ = l.db.SQL().Exec(`
                                UPDATE sync_deferrals
                                SET status = 'active', date_resolved = NULL
                                WHERE stage = ? AND item_id = ? AND status = 'resolved'`,
				stage, r.itemID)
		}
	}
	return resolved
}

// ReviewedIDs mirrors reviewed_ids: the stage's reviewed set — every
// item whose deferral STANDS (active or resolved — anything but
// reopened). This is what stops repeat alerts. Never raises.
func (l *DeferredLog) ReviewedIDs(stage string) map[string]bool {
	out := map[string]bool{}
	l.mu.Lock()
	defer l.mu.Unlock()
	rows, err := l.db.SQL().Query(`
                SELECT item_id FROM sync_deferrals
                WHERE stage = ? AND status != 'reopened'`, stage)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var itemID string
		if err := rows.Scan(&itemID); err == nil {
			out[itemID] = true
		}
	}
	return out
}

// RowCount mirrors row_count: total ledger rows (the log's own count
// — never shrinks). An empty stage counts every row.
func (l *DeferredLog) RowCount(stage string) int {
	l.mu.Lock()
	defer l.mu.Unlock()
	var n int
	var row *sql.Row
	if stage == "" {
		row = l.db.SQL().QueryRow(`SELECT COUNT(*) FROM sync_deferrals`)
	} else {
		row = l.db.SQL().QueryRow(
			`SELECT COUNT(*) FROM sync_deferrals WHERE stage = ?`, stage)
	}
	if err := row.Scan(&n); err != nil {
		return 0
	}
	return n
}

// Rows mirrors rows: the Deferred log tab's rows — newest-reviewed
// first (the log's own story order). Optional stage filter + a SEARCH
// over item, display name, reason and note (case-folded; the reason's
// own label rides the haystack). Never raises.
func (l *DeferredLog) Rows(stage, query string) []LedgerRow {
	q := strings.ToLower(strings.TrimSpace(query))
	l.mu.Lock()
	defer l.mu.Unlock()
	var (
		rows *sql.Rows
		err  error
	)
	if stage == "" {
		rows, err = l.db.SQL().Query(`
                        SELECT stage, item_id, display_name, reason, note,
                               status, date_reviewed, date_resolved, reopen_count
                        FROM sync_deferrals
                        ORDER BY date_reviewed DESC, id DESC`)
	} else {
		rows, err = l.db.SQL().Query(`
                        SELECT stage, item_id, display_name, reason, note,
                               status, date_reviewed, date_resolved, reopen_count
                        FROM sync_deferrals
                        WHERE stage = ?
                        ORDER BY date_reviewed DESC, id DESC`, stage)
	}
	if err != nil {
		return nil
	}
	defer rows.Close()
	var out []LedgerRow
	for rows.Next() {
		var r LedgerRow
		var resolved sql.NullString
		if err := rows.Scan(&r.Stage, &r.ItemID, &r.DisplayName, &r.Reason,
			&r.Note, &r.Status, &r.DateReviewed, &resolved,
			&r.ReopenCount); err != nil {
			continue
		}
		r.DateResolved = resolved.String
		if q != "" {
			label := ReasonLabels[r.Reason]
			haystack := strings.ToLower(r.ItemID + " " + r.DisplayName +
				" " + r.Reason + " " + r.Note + " " + label)
			if !strings.Contains(haystack, q) {
				continue
			}
		}
		out = append(out, r)
	}
	return out
}

// Close is idempotent (the cache_db law).
func (l *DeferredLog) Close() error {
	if l == nil || l.db == nil {
		return nil
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	err := l.db.Close()
	l.db = nil
	return err
}
