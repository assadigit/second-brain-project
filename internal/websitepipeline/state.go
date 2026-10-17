// state.go — T-23's state layer (core/website_state.py): WebsiteStateDB,
// THE ONE SPELLING boundary, the 2×attempts backoff, both settlements,
// the v0.64.0 healing pass. The package laws and the full divergence
// record live in doc.go.
package websitepipeline

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/store"
)

// The fetch-retry policy constants (website_state.py:23-24) + the two
// settlement meta keys (v0.63.2 / v0.64.1).
const (
	// MaxFetchRetries — SPEC: retried automatically up to 3 times.
	MaxFetchRetries = 3
	// RetryBackoffDays — "over several days".
	RetryBackoffDays = 2
	// SettledMetaKey — v0.63.2 THE SETTLEMENT (one time): every website
	// the system already knew when the owner's law arrived is
	// "addressed and processed" — never machine-fetched again.
	SettledMetaKey = "websites_settled_at"
	// QueueHistorySettledMetaKey — v0.64.1 THE QUEUE-HISTORY SETTLEMENT
	// (the second one-time pass): every link present in the bot's
	// history at the queue door is settled.
	QueueHistorySettledMetaKey = "websites_queue_history_settled_at"
)

// StateKey mirrors _key (website_state.py:44-56) — THE ONE SPELLING at
// the state DB's own boundary: every URL that crosses into or out of a
// ledger method is normalized FIRST, so a row written as
// "https://site/" answers a probe for "https://site" (and every other
// spelling) — the tables hold one canonical key per link, whatever
// spelling the caller carried. Empty input folds to "" (Python's
// str(url or ”).strip() — the None case included); the reference's
// defensive except-path (broken input returned raw) is unreachable for
// the pure-string normalizer and thus not carried.
func StateKey(url string) string {
	u := strings.TrimSpace(url)
	if u == "" {
		return ""
	}
	return links.NormalizeWebsiteURL(u)
}

// The adopt-as-is DDL (website_state.py:94-132) — byte-identical to
// the reference (ADR-0003: adopt the existing database, no migrations).
const (
	ddlWebsitesProcessed = `CREATE TABLE IF NOT EXISTS websites_processed (
                    url TEXT PRIMARY KEY,
                    note_path TEXT,
                    category TEXT,
                    subcategory TEXT,
                    fetch_status TEXT,
                    processed_at TEXT
                )`
	ddlWebsiteRetryQueue = `CREATE TABLE IF NOT EXISTS website_retry_queue (
                    url TEXT PRIMARY KEY,
                    attempts INTEGER NOT NULL DEFAULT 0,
                    last_error TEXT,
                    next_attempt_at TEXT,
                    first_failed_at TEXT,
                    last_failed_at TEXT
                )`
	ddlDismissedURLs = `CREATE TABLE IF NOT EXISTS dismissed_urls (
                    url TEXT PRIMARY KEY,
                    reason TEXT,
                    dismissed_at TEXT
                )`
	ddlWebsiteStateMeta = `CREATE TABLE IF NOT EXISTS website_state_meta (
                    key TEXT PRIMARY KEY,
                    value TEXT
                )`
	ddlWebsitesSettled = `CREATE TABLE IF NOT EXISTS websites_settled (
                    url TEXT PRIMARY KEY,
                    settled_at TEXT
                )`
)

// WebsiteStateDB is the Websites pipeline's persistent state ledger:
// the shared SQLite plumbing (ONE connection, the embedded
// not-reentrant RWMutex) plus the five cache.db tables. Same file as
// CacheDB (the call sites' resolved path); own connection.
type WebsiteStateDB struct {
	*store.DB
}

// NewWebsiteStateDB mirrors WebsiteStateDB.__init__
// (website_state.py:86-133): the literal default "cache.db" is anchored
// to APP_DIR (never the current working dir), the five tables are
// created if absent, busy_timeout=30000 (plus the Go store law's
// WAL/NORMAL — the documented unified-pragmas divergence).
func NewWebsiteStateDB(dbPath string) (*WebsiteStateDB, error) {
	if dbPath == "cache.db" {
		dbPath = filepath.Join(constants.AppDir(), "cache.db")
	}
	base, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	d := &WebsiteStateDB{DB: base}
	if err := d.createTables(); err != nil {
		base.Close()
		return nil, err
	}
	return d, nil
}

func (d *WebsiteStateDB) createTables() error {
	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return err
	}
	for _, stmt := range []string{
		ddlWebsitesProcessed, ddlWebsiteRetryQueue, ddlDismissedURLs,
		ddlWebsiteStateMeta, ddlWebsitesSettled,
	} {
		if _, err := tx.Exec(stmt); err != nil {
			_ = tx.Rollback()
			return err
		}
	}
	return tx.Commit()
}

// nowSeconds is the write shape law: naive-local
// datetime.now().isoformat(timespec='seconds') — YYYY-MM-DDTHH:MM:SS
// (every website_state timestamp writes this shape). T-40's clock
// seam: reads the package clockNow (the golden runner freezes it the
// same way the reference generator freezes ws.datetime); production
// reads the wall clock.
func nowSeconds() string {
	return store.PyISOFormatSeconds(clockNow())
}

// -- row shapes (Python's dict returns; NULL ⇔ nil *string) -------------

// ProcessedRow mirrors processed_row's dict.
type ProcessedRow struct {
	URL         string
	NotePath    *string
	Category    *string
	Subcategory *string
	FetchStatus *string
	ProcessedAt *string
}

// RetryRow mirrors retry_row's dict.
type RetryRow struct {
	URL           string
	Attempts      int64
	LastError     *string
	NextAttemptAt *string
	FirstFailedAt *string
	LastFailedAt  *string
}

// DismissedRow mirrors dismissed_row's dict.
type DismissedRow struct {
	URL         string
	Reason      *string
	DismissedAt *string
}

// SettleExistingReport mirrors settle_existing's dict:
// {'settled': n, 'retries_cleared': m, 'already': bool}.
type SettleExistingReport struct {
	Settled        int
	RetriesCleared int
	Already        bool
}

// SettleQueueHistoryReport mirrors settle_queue_history's dict:
// {'settled': n, 'already': bool}.
type SettleQueueHistoryReport struct {
	Settled int
	Already bool
}

// NormalizeLedgerReport mirrors normalize_ledger_keys's dict:
// {'rekeyed': n, 'merged': m}.
type NormalizeLedgerReport struct {
	Rekeyed int
	Merged  int
}

// PurgePlaceholder is one (url, note_path) pair from
// purge_blocked_domains' report.
type PurgePlaceholder struct {
	URL      string
	NotePath string
}

// PurgeReport mirrors purge_blocked_domains' dict:
// {'retries': n, 'placeholders': [(url, path)], 'dismissed': n}.
type PurgeReport struct {
	Retries      int
	Placeholders []PurgePlaceholder
	Dismissed    int
}

// -- writes -----------------------------------------------------------

// MarkProcessed mirrors mark_processed (website_state.py:137-150):
// record the dedupe row for a processed URL. Does NOT touch the retry
// queue — a fetch-failed link keeps its _review note AND its pending
// retries; ResolveRetry clears the queue only on full success.
func (d *WebsiteStateDB) MarkProcessed(url, notePath, category,
	subcategory, fetchStatus string) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec(
		"INSERT OR REPLACE INTO websites_processed "+
			"(url, note_path, category, subcategory, fetch_status,"+
			" processed_at) VALUES (?,?,?,?,?,?)",
		StateKey(url), notePath, category, subcategory, fetchStatus,
		nowSeconds())
	return err
}

// EnqueueRetry mirrors enqueue_retry (website_state.py:152-172):
// record/refresh a fetch failure. next_attempt_at backs off
// RetryBackoffDays per attempt (2×attempts days — timedelta's exact
// hours, not calendar days) so retries spread over several days. The
// ORIGINAL first_failed_at survives refreshes; last_error is
// rune-truncated to 500.
func (d *WebsiteStateDB) EnqueueRetry(url, errMsg string) error {
	now := time.Now()
	nowISO := store.PyISOFormatSeconds(now)
	key := StateKey(url)

	d.Lock()
	defer d.Unlock()
	var attempts int64
	var firstFailed sql.NullString
	err := d.SQL().QueryRow(
		"SELECT attempts, first_failed_at FROM website_retry_queue"+
			" WHERE url=?", key).Scan(&attempts, &firstFailed)
	if err != nil && err != sql.ErrNoRows {
		return err
	}
	if err == sql.ErrNoRows {
		attempts = 1
	} else {
		attempts++
	}
	if err == sql.ErrNoRows || !firstFailed.Valid || firstFailed.String == "" {
		firstFailed = sql.NullString{String: nowISO, Valid: true}
	}
	// timedelta(days=RETRY_BACKOFF_DAYS * attempts) — exact duration.
	nextAt := now.Add(time.Duration(RetryBackoffDays*attempts) * 24 *
		time.Hour)
	_, err = d.SQL().Exec(
		"INSERT OR REPLACE INTO website_retry_queue "+
			"(url, attempts, last_error, next_attempt_at, first_failed_at,"+
			" last_failed_at) VALUES (?,?,?,?,?,?)",
		key, attempts, helpers.CutRunes(errMsg, 500),
		store.PyISOFormatSeconds(nextAt), firstFailed.String, nowISO)
	return err
}

// ResolveRetry mirrors resolve_retry (website_state.py:174-182): drop a
// link from the retry queue — called when processing fully succeeded
// (note written with real content). Kept SEPARATE from MarkProcessed on
// purpose: a fetch-failed link writes a _review note AND stays queued
// for its automatic retries.
func (d *WebsiteStateDB) ResolveRetry(url string) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec("DELETE FROM website_retry_queue WHERE url=?",
		StateKey(url))
	return err
}

// Dismiss mirrors dismiss (website_state.py:184-191); the reason is
// rune-truncated to 200.
func (d *WebsiteStateDB) Dismiss(url, reason string) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec(
		"INSERT OR REPLACE INTO dismissed_urls (url, reason,"+
			" dismissed_at) VALUES (?,?,?)",
		StateKey(url), helpers.CutRunes(reason, 200), nowSeconds())
	return err
}

// Undismiss mirrors undismiss (website_state.py:193-202) — the
// graveyard's other door: remove a dismissal so a ♻️-revived link is
// fetched like new again. Returns True when a dismissal was actually
// removed.
func (d *WebsiteStateDB) Undismiss(url string) (bool, error) {
	d.Lock()
	defer d.Unlock()
	res, err := d.SQL().Exec(
		"DELETE FROM dismissed_urls WHERE url=?", StateKey(url))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// ForgetFailedRow mirrors forget_failed_row (website_state.py:204-225):
// delete a link's processed row ONLY when it is a failed _review
// placeholder record (fetch_status 'failed' + note_path inside
// _review) — a real note's row is NEVER touched. Returns True when a
// row was removed.
func (d *WebsiteStateDB) ForgetFailedRow(url string) (bool, error) {
	key := StateKey(url)
	d.Lock()
	defer d.Unlock()
	var path sql.NullString
	var status sql.NullString
	err := d.SQL().QueryRow(
		"SELECT note_path, fetch_status FROM websites_processed"+
			" WHERE url=?", key).Scan(&path, &status)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !status.Valid || status.String != "failed" ||
		!strings.Contains(strings.ReplaceAll(path.String, "\\", "/"),
			"_review") {
		return false, nil
	}
	res, err := d.SQL().Exec(
		"DELETE FROM websites_processed WHERE url=?", key)
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// ForgetRow mirrors forget_row (website_state.py:227-245) — THE
// BANISHMENT's ledger half: remove a link's processed row
// UNCONDITIONALLY (the caller just removed the note FILE on purpose).
// Returns True when a row was removed.
func (d *WebsiteStateDB) ForgetRow(url string) (bool, error) {
	d.Lock()
	defer d.Unlock()
	res, err := d.SQL().Exec(
		"DELETE FROM websites_processed WHERE url=?", StateKey(url))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// -- v0.63.2 THE SETTLED LEDGER ----------------------------------------

// IsSettled mirrors is_settled (website_state.py:249-256).
func (d *WebsiteStateDB) IsSettled(url string) (bool, error) {
	d.RLock()
	defer d.RUnlock()
	var one int
	err := d.SQL().QueryRow(
		"SELECT 1 FROM websites_settled WHERE url=?",
		StateKey(url)).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// Unsettle mirrors unsettle (website_state.py:279-289) — the ♻️ door
// back: remove one URL from the settled ledger so a revived link is
// fetched like new again. Returns True when a settled row was actually
// removed.
func (d *WebsiteStateDB) Unsettle(url string) (bool, error) {
	d.Lock()
	defer d.Unlock()
	res, err := d.SQL().Exec(
		"DELETE FROM websites_settled WHERE url=?", StateKey(url))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// SettledCount mirrors settled_count (website_state.py:291-296) — the
// ledger's size (the honest settlement line).
func (d *WebsiteStateDB) SettledCount() (int, error) {
	d.RLock()
	defer d.RUnlock()
	var n int
	err := d.SQL().QueryRow(
		"SELECT COUNT(*) FROM websites_settled").Scan(&n)
	return n, err
}

// SettleExisting mirrors settle_existing (website_state.py:298-361) —
// THE SETTLEMENT, one time, on the owner's word. Seeds the settled
// ledger from every URL the state knows (processed + retry queue +
// dismissed) plus the extraURLs the vault-level caller adds; the retry
// queue (whose every row predates the law) is bulk-cleared. Guarded by
// the SettledMetaKey: a second call is a no-op. NOTE the honest-count
// quirk: Settled counts the unique URLs this call considered — NOT the
// rows actually inserted (INSERT OR IGNORE may skip existing settled
// rows) — exactly like the reference.
func (d *WebsiteStateDB) SettleExisting(extraURLs []string) (
	SettleExistingReport, error) {
	// The meta guard (read under the read lock).
	d.RLock()
	var guard sql.NullString
	err := d.SQL().QueryRow(
		"SELECT value FROM website_state_meta WHERE key=?",
		SettledMetaKey).Scan(&guard)
	d.RUnlock()
	if err != nil && err != sql.ErrNoRows {
		return SettleExistingReport{}, err
	}
	if err == nil {
		return SettleExistingReport{Settled: 0, RetriesCleared: 0,
			Already: true}, nil
	}

	nowISO := nowSeconds()
	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return SettleExistingReport{}, err
	}
	var urls []string
	for _, q := range []string{
		"SELECT url FROM websites_processed",
		"SELECT url FROM website_retry_queue",
		"SELECT url FROM dismissed_urls",
	} {
		rows, err := tx.Query(q)
		if err != nil {
			_ = tx.Rollback()
			return SettleExistingReport{}, err
		}
		for rows.Next() {
			var u sql.NullString
			if err := rows.Scan(&u); err != nil {
				rows.Close()
				_ = tx.Rollback()
				return SettleExistingReport{}, err
			}
			urls = append(urls, u.String)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			_ = tx.Rollback()
			return SettleExistingReport{}, err
		}
	}
	urls = append(urls, extraURLs...)
	// THE ONE SPELLING: whatever spelling a row (or a caller's extra)
	// carries, the ledger settles the CANONICAL key.
	seen := make(map[string]bool)
	var unique []string
	for _, u := range urls {
		k := links.NormalizeWebsiteURL(u)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		unique = append(unique, k)
	}
	for _, u := range unique {
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO websites_settled (url, settled_at)"+
				" VALUES (?,?)", u, nowISO); err != nil {
			_ = tx.Rollback()
			return SettleExistingReport{}, err
		}
	}
	res, err := tx.Exec("DELETE FROM website_retry_queue")
	if err != nil {
		_ = tx.Rollback()
		return SettleExistingReport{}, err
	}
	cleared, _ := res.RowsAffected()
	if _, err := tx.Exec(
		"INSERT OR REPLACE INTO website_state_meta (key, value)"+
			" VALUES (?,?)", SettledMetaKey, nowISO); err != nil {
		_ = tx.Rollback()
		return SettleExistingReport{}, err
	}
	if err := tx.Commit(); err != nil {
		return SettleExistingReport{}, err
	}
	return SettleExistingReport{Settled: len(unique),
		RetriesCleared: int(cleared), Already: false}, nil
}

// SettleQueueHistory mirrors settle_queue_history
// (website_state.py:365-429) — THE QUEUE-HISTORY SETTLEMENT, one time,
// on the owner's word: every URL the queue door hands it (the FULL bot
// history) is settled; those URLs' retry rows resolve, but the queue is
// NOT bulk-cleared (post-law failures keep their honest 3-retry
// lifecycle). Guarded by the QueueHistorySettledMetaKey. Settled counts
// only the rows this pass actually INSERTED (the before/after COUNT
// difference — the honesty the v0.64.1 report demanded).
func (d *WebsiteStateDB) SettleQueueHistory(extraURLs []string) (
	SettleQueueHistoryReport, error) {
	d.RLock()
	var guard sql.NullString
	err := d.SQL().QueryRow(
		"SELECT value FROM website_state_meta WHERE key=?",
		QueueHistorySettledMetaKey).Scan(&guard)
	d.RUnlock()
	if err != nil && err != sql.ErrNoRows {
		return SettleQueueHistoryReport{}, err
	}
	if err == nil {
		return SettleQueueHistoryReport{Settled: 0, Already: true}, nil
	}

	nowISO := nowSeconds()
	// THE ONE SPELLING: whatever spelling the history carries, the
	// ledger settles the CANONICAL key.
	seen := make(map[string]bool)
	var unique []string
	for _, u := range extraURLs {
		if u == "" {
			continue
		}
		k := links.NormalizeWebsiteURL(u)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		unique = append(unique, k)
	}

	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return SettleQueueHistoryReport{}, err
	}
	var before int
	if err := tx.QueryRow(
		"SELECT COUNT(*) FROM websites_settled").Scan(&before); err != nil {
		_ = tx.Rollback()
		return SettleQueueHistoryReport{}, err
	}
	for _, u := range unique {
		if _, err := tx.Exec(
			"INSERT OR IGNORE INTO websites_settled (url, settled_at)"+
				" VALUES (?,?)", u, nowISO); err != nil {
			_ = tx.Rollback()
			return SettleQueueHistoryReport{}, err
		}
	}
	var after int
	if err := tx.QueryRow(
		"SELECT COUNT(*) FROM websites_settled").Scan(&after); err != nil {
		_ = tx.Rollback()
		return SettleQueueHistoryReport{}, err
	}
	// the settled URLs' retry rows resolve — a settled link's retries
	// never fire again (the queue itself is NOT bulk-cleared).
	for _, u := range unique {
		if _, err := tx.Exec(
			"DELETE FROM website_retry_queue WHERE url=?", u); err != nil {
			_ = tx.Rollback()
			return SettleQueueHistoryReport{}, err
		}
	}
	if _, err := tx.Exec(
		"INSERT OR REPLACE INTO website_state_meta (key, value)"+
			" VALUES (?,?)", QueueHistorySettledMetaKey, nowISO); err != nil {
		_ = tx.Rollback()
		return SettleQueueHistoryReport{}, err
	}
	if err := tx.Commit(); err != nil {
		return SettleQueueHistoryReport{}, err
	}
	return SettleQueueHistoryReport{Settled: after - before,
		Already: false}, nil
}

// -- v0.64.0 THE ONE SPELLING — the ledgers' healing pass ------------

// healProcRow / healRetryRow — the collected healing-pass rows.
type healProcRow struct {
	url, notePath, category, subcategory, fetchStatus, processedAt string
}

type healRetryRow struct {
	url                                string
	attempts                           int64
	lastError, nextAt, firstAt, lastAt string
}

// NormalizeLedgerKeys mirrors normalize_ledger_keys
// (website_state.py:433-545): re-key every ledger row under the fixed
// normalizer. When both spellings of one link already exist, the rows
// MERGE under a deterministic preference (a processed row whose note
// exists on disk beats one whose does not; a retry row with more
// attempts beats one with fewer; the dismissed and settled tables keep
// the first row by key order). Idempotent and cheap, so every door may
// call it freely. Like the reference's blanket try/except, a failed
// statement rolls the pass back and returns the counts so far — never
// an error on the caller's head.
func (d *WebsiteStateDB) NormalizeLedgerKeys() (NormalizeLedgerReport, error) {
	rep := NormalizeLedgerReport{}
	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return rep, nil // the Python swallow: never raises
	}
	// websites_processed: prefer a live note on disk.
	rows, err := tx.Query(
		"SELECT url, note_path, category, subcategory," +
			" fetch_status, processed_at FROM websites_processed" +
			" ORDER BY url")
	if err != nil {
		_ = tx.Rollback()
		return rep, nil
	}
	var procRows []healProcRow
	for rows.Next() {
		var r healProcRow
		var np, cat, sub, fs, pa sql.NullString
		if err := rows.Scan(&r.url, &np, &cat, &sub, &fs, &pa); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return rep, nil
		}
		r.notePath, r.category, r.subcategory = np.String, cat.String,
			sub.String
		r.fetchStatus, r.processedAt = fs.String, pa.String
		procRows = append(procRows, r)
	}
	rows.Close()
	byKey := make(map[string]healProcRow)
	var order []string
	for _, r := range procRows {
		key := ""
		if r.url != "" {
			key = links.NormalizeWebsiteURL(r.url)
		}
		if key == "" {
			continue
		}
		if key == r.url {
			if _, ok := byKey[key]; !ok {
				byKey[key] = r
				order = append(order, key)
			}
			continue
		}
		prev, ok := byKey[key]
		if !ok || (noteAlive(r.notePath) && !noteAlive(prev.notePath)) {
			if !ok {
				order = append(order, key)
			}
			byKey[key] = r // the better row wins
		}
		if ok {
			rep.Merged++
		}
		rep.Rekeyed++
	}
	if anyNonCanonical(procRows) {
		if _, err := tx.Exec("DELETE FROM websites_processed"); err != nil {
			_ = tx.Rollback()
			return rep, nil
		}
		for _, k := range order {
			v := byKey[k]
			if _, err := tx.Exec(
				"INSERT OR REPLACE INTO websites_processed"+
					" (url, note_path, category, subcategory,"+
					" fetch_status, processed_at) VALUES (?,?,?,?,?,?)",
				k, v.notePath, v.category, v.subcategory, v.fetchStatus,
				v.processedAt); err != nil {
				_ = tx.Rollback()
				return rep, nil
			}
		}
	}
	// website_retry_queue: more attempts wins.
	rows, err = tx.Query(
		"SELECT url, attempts, last_error, next_attempt_at," +
			" first_failed_at, last_failed_at FROM" +
			" website_retry_queue ORDER BY url")
	if err != nil {
		_ = tx.Rollback()
		return rep, nil
	}
	var retryRows []healRetryRow
	for rows.Next() {
		var r healRetryRow
		var le, na, fa, la sql.NullString
		if err := rows.Scan(&r.url, &r.attempts, &le, &na, &fa, &la); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return rep, nil
		}
		r.lastError, r.nextAt, r.firstAt, r.lastAt = le.String, na.String,
			fa.String, la.String
		retryRows = append(retryRows, r)
	}
	rows.Close()
	byRetry := make(map[string]healRetryRow)
	var rOrder []string
	for _, r := range retryRows {
		key := ""
		if r.url != "" {
			key = links.NormalizeWebsiteURL(r.url)
		}
		if key == "" {
			continue
		}
		if key == r.url {
			if _, ok := byRetry[key]; !ok {
				byRetry[key] = r
				rOrder = append(rOrder, key)
			}
			continue
		}
		prev, ok := byRetry[key]
		if !ok || r.attempts > prev.attempts {
			if !ok {
				rOrder = append(rOrder, key)
			}
			byRetry[key] = r
		}
		if ok {
			rep.Merged++
		}
		rep.Rekeyed++
	}
	if anyNonCanonicalRetry(retryRows) {
		if _, err := tx.Exec("DELETE FROM website_retry_queue"); err != nil {
			_ = tx.Rollback()
			return rep, nil
		}
		for _, k := range rOrder {
			v := byRetry[k]
			if _, err := tx.Exec(
				"INSERT OR REPLACE INTO website_retry_queue"+
					" (url, attempts, last_error, next_attempt_at,"+
					" first_failed_at, last_failed_at)"+
					" VALUES (?,?,?,?,?,?)",
				k, v.attempts, v.lastError, v.nextAt, v.firstAt,
				v.lastAt); err != nil {
				_ = tx.Rollback()
				return rep, nil
			}
		}
	}
	// dismissed_urls + websites_settled: first wins (ORDER BY url).
	for _, table := range []string{"dismissed_urls", "websites_settled"} {
		rows, err = tx.Query("SELECT * FROM " + table + " ORDER BY url")
		if err != nil {
			_ = tx.Rollback()
			return rep, nil
		}
		var rawRows [][]sql.NullString
		cols, _ := rows.Columns()
		for rows.Next() {
			vals := make([]sql.NullString, len(cols))
			ptrs := make([]any, len(cols))
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				rows.Close()
				_ = tx.Rollback()
				return rep, nil
			}
			rawRows = append(rawRows, vals)
		}
		rows.Close()
		changed := false
		// kept: first-wins by canonical key, insertion order preserved.
		type keptRow struct {
			key  string
			vals []sql.NullString
		}
		var kept []keptRow
		keptIdx := make(map[string]int)
		for _, r := range rawRows {
			rawURL := ""
			if len(r) > 0 && r[0].Valid {
				rawURL = r[0].String
			}
			key := ""
			if rawURL != "" {
				key = links.NormalizeWebsiteURL(rawURL)
			}
			if key == "" {
				// a row whose URL normalizes to nothing stays as-is,
				// keyed by its raw spelling (first wins).
				if _, ok := keptIdx[rawURL]; !ok {
					keptIdx[rawURL] = len(kept)
					kept = append(kept, keptRow{rawURL, r})
				}
				continue
			}
			if key == rawURL {
				if _, ok := keptIdx[key]; !ok {
					keptIdx[key] = len(kept)
					kept = append(kept, keptRow{key, r})
				}
				continue
			}
			changed = true
			rep.Rekeyed++
			if idx, ok := keptIdx[key]; ok {
				rep.Merged++
				_ = idx
			} else {
				nr := append([]sql.NullString{
					{String: key, Valid: true}}, r[1:]...)
				keptIdx[key] = len(kept)
				kept = append(kept, keptRow{key, nr})
			}
		}
		if changed {
			if _, err := tx.Exec("DELETE FROM " + table); err != nil {
				_ = tx.Rollback()
				return rep, nil
			}
			// the reference writes the full column list from _TABLE_COLS.
			var colList string
			var n int
			if table == "dismissed_urls" {
				colList, n = "url, reason, dismissed_at", 3
			} else {
				colList, n = "url, settled_at", 2
			}
			ph := strings.Repeat("?,", n)
			ph = strings.TrimSuffix(ph, ",")
			for _, kr := range kept {
				vals := make([]any, len(kr.vals))
				for i, v := range kr.vals {
					if v.Valid {
						vals[i] = v.String
					} else {
						vals[i] = nil
					}
				}
				if _, err := tx.Exec(
					"INSERT OR REPLACE INTO "+table+" ("+colList+
						") VALUES ("+ph+")", vals...); err != nil {
					_ = tx.Rollback()
					return rep, nil
				}
			}
		}
	}
	if err := tx.Commit(); err != nil {
		return rep, nil
	}
	return rep, nil
}

// noteAlive mirrors _row_note_alive (website_state.py:35-42): does a
// processed row's note still exist on disk? (the healing pass's
// preference: a live note beats a remembered one).
func noteAlive(notePath string) bool {
	if notePath == "" {
		return false
	}
	fi, err := os.Stat(notePath)
	return err == nil && !fi.IsDir()
}

func anyNonCanonical(rows []healProcRow) bool {
	for _, r := range rows {
		if r.url != "" && links.NormalizeWebsiteURL(r.url) != r.url {
			return true
		}
	}
	return false
}

func anyNonCanonicalRetry(rows []healRetryRow) bool {
	for _, r := range rows {
		if r.url != "" && links.NormalizeWebsiteURL(r.url) != r.url {
			return true
		}
	}
	return false
}

// -- v0.19.0 proxy-epoch meta + retry re-arm --------------------------

// GetMeta mirrors get_meta (website_state.py:549-555): one row from
// website_state_meta (nil when absent).
func (d *WebsiteStateDB) GetMeta(key string) (*string, error) {
	d.RLock()
	defer d.RUnlock()
	var v sql.NullString
	err := d.SQL().QueryRow(
		"SELECT value FROM website_state_meta WHERE key=?",
		key).Scan(&v)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	s := v.String
	return &s, nil
}

// SetMeta mirrors set_meta (website_state.py:557-562).
func (d *WebsiteStateDB) SetMeta(key, value string) error {
	d.Lock()
	defer d.Unlock()
	_, err := d.SQL().Exec(
		"INSERT OR REPLACE INTO website_state_meta (key, value)"+
			" VALUES (?,?)", key, value)
	return err
}

// ResetRetryAttempts mirrors reset_retry_attempts
// (website_state.py:564-581) — ONE link's retry counter reborn:
// attempts=0, due NOW. Returns True when a row was actually reset.
func (d *WebsiteStateDB) ResetRetryAttempts(url string) (bool, error) {
	d.Lock()
	defer d.Unlock()
	res, err := d.SQL().Exec(
		"UPDATE website_retry_queue"+
			" SET attempts=0, next_attempt_at=? WHERE url=?",
		nowSeconds(), StateKey(url))
	if err != nil {
		return false, err
	}
	n, err := res.RowsAffected()
	return n > 0, err
}

// RearmRetries mirrors rearm_retries (website_state.py:583-596) —
// reset the ENTIRE retry queue (attempts=0, due NOW). Returns how many
// rows were re-armed.
func (d *WebsiteStateDB) RearmRetries() (int, error) {
	d.Lock()
	defer d.Unlock()
	res, err := d.SQL().Exec(
		"UPDATE website_retry_queue SET attempts=0, next_attempt_at=?",
		nowSeconds())
	if err != nil {
		return 0, err
	}
	n, err := res.RowsAffected()
	return int(n), err
}

// PurgeBlockedDomains mirrors purge_blocked_domains
// (website_state.py:598-642) — v0.20.0: remove blocked-domain links
// from the retry queue and the failed _review placeholder records,
// marking every purged URL dismissed so it can never re-enter.
// isBlocked(url) -> bool is injected (links.domain_is_blocked partial).
// Only app-owned _review placeholders are deleted (a real note for a
// blocked domain is the owner's — kept). File deletion is the CALLER's
// job (dry-run aware); this method only touches the DB. Idempotent —
// a second run finds nothing.
func (d *WebsiteStateDB) PurgeBlockedDomains(
	isBlocked func(string) bool) (PurgeReport, error) {
	rep := PurgeReport{}
	d.Lock()
	defer d.Unlock()
	tx, err := d.SQL().Begin()
	if err != nil {
		return rep, err
	}
	// Rows collected before any nested statement (the package law).
	var retryURLs []string
	rows, err := tx.Query("SELECT url FROM website_retry_queue")
	if err != nil {
		_ = tx.Rollback()
		return rep, err
	}
	for rows.Next() {
		var u sql.NullString
		if err := rows.Scan(&u); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return rep, err
		}
		retryURLs = append(retryURLs, u.String)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return rep, err
	}
	type ppRow struct{ url, path, status string }
	var processedRows []ppRow
	rows, err = tx.Query(
		"SELECT url, note_path, fetch_status FROM websites_processed")
	if err != nil {
		_ = tx.Rollback()
		return rep, err
	}
	for rows.Next() {
		var u, p, s sql.NullString
		if err := rows.Scan(&u, &p, &s); err != nil {
			rows.Close()
			_ = tx.Rollback()
			return rep, err
		}
		processedRows = append(processedRows,
			ppRow{u.String, p.String, s.String})
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		_ = tx.Rollback()
		return rep, err
	}
	for _, r := range processedRows {
		if r.status != "failed" || r.path == "" {
			continue
		}
		if !isBlocked(r.url) {
			continue
		}
		// ONLY app-owned _review placeholders — never a real note.
		if !strings.Contains(strings.ReplaceAll(r.path, "\\", "/"),
			"_review") {
			continue
		}
		rep.Placeholders = append(rep.Placeholders,
			PurgePlaceholder{URL: r.url, NotePath: r.path})
	}
	var retryGone []string
	for _, u := range retryURLs {
		if isBlocked(u) {
			retryGone = append(retryGone, u)
		}
	}
	nowISO := nowSeconds()
	purged := map[string]bool{}
	for _, u := range retryGone {
		purged[u] = true
	}
	for _, p := range rep.Placeholders {
		purged[p.URL] = true
	}
	for _, u := range retryGone {
		if _, err := tx.Exec(
			"DELETE FROM website_retry_queue WHERE url=?", u); err != nil {
			_ = tx.Rollback()
			return rep, err
		}
	}
	for _, p := range rep.Placeholders {
		if _, err := tx.Exec(
			"DELETE FROM websites_processed WHERE url=?", p.URL); err != nil {
			_ = tx.Rollback()
			return rep, err
		}
	}
	for u := range purged {
		if _, err := tx.Exec(
			"INSERT OR REPLACE INTO dismissed_urls (url, reason,"+
				" dismissed_at) VALUES (?,?,?)",
			u, "blocked domain (v0.20.0 purge)", nowISO); err != nil {
			_ = tx.Rollback()
			return rep, err
		}
	}
	if err := tx.Commit(); err != nil {
		return rep, err
	}
	rep.Retries = len(retryGone)
	rep.Dismissed = len(purged)
	return rep, nil
}

// -- reads ------------------------------------------------------------

// IsProcessed mirrors is_processed (website_state.py:646-651).
func (d *WebsiteStateDB) IsProcessed(url string) (bool, error) {
	d.RLock()
	defer d.RUnlock()
	var one int
	err := d.SQL().QueryRow(
		"SELECT 1 FROM websites_processed WHERE url=?",
		StateKey(url)).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ProcessedRow mirrors processed_row (website_state.py:653-663).
func (d *WebsiteStateDB) ProcessedRow(url string) (*ProcessedRow, error) {
	d.RLock()
	defer d.RUnlock()
	var r ProcessedRow
	var np, cat, sub, fs, pa sql.NullString
	err := d.SQL().QueryRow(
		"SELECT url, note_path, category, subcategory, fetch_status,"+
			" processed_at FROM websites_processed WHERE url=?",
		StateKey(url)).Scan(&r.URL, &np, &cat, &sub, &fs, &pa)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.NotePath, r.Category = nullStrPtr(np), nullStrPtr(cat)
	r.Subcategory, r.FetchStatus = nullStrPtr(sub), nullStrPtr(fs)
	r.ProcessedAt = nullStrPtr(pa)
	return &r, nil
}

// IsDismissed mirrors is_dismissed (website_state.py:665-670).
func (d *WebsiteStateDB) IsDismissed(url string) (bool, error) {
	d.RLock()
	defer d.RUnlock()
	var one int
	err := d.SQL().QueryRow(
		"SELECT 1 FROM dismissed_urls WHERE url=?",
		StateKey(url)).Scan(&one)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// DismissedRow mirrors dismissed_row (website_state.py:672-683) — the
// dismissal's stored reason (raw: NULL stays nil).
func (d *WebsiteStateDB) DismissedRow(url string) (*DismissedRow, error) {
	d.RLock()
	defer d.RUnlock()
	var r DismissedRow
	var reason, at sql.NullString
	err := d.SQL().QueryRow(
		"SELECT url, reason, dismissed_at FROM dismissed_urls"+
			" WHERE url=?", StateKey(url)).Scan(&r.URL, &reason, &at)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.Reason, r.DismissedAt = nullStrPtr(reason), nullStrPtr(at)
	return &r, nil
}

// RetryRow mirrors retry_row (website_state.py:685-695).
func (d *WebsiteStateDB) RetryRow(url string) (*RetryRow, error) {
	d.RLock()
	defer d.RUnlock()
	var r RetryRow
	var le, na, fa, la sql.NullString
	err := d.SQL().QueryRow(
		"SELECT url, attempts, last_error, next_attempt_at,"+
			" first_failed_at, last_failed_at FROM website_retry_queue"+
			" WHERE url=?", StateKey(url)).
		Scan(&r.URL, &r.Attempts, &le, &na, &fa, &la)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	r.LastError, r.NextAttemptAt = nullStrPtr(le), nullStrPtr(na)
	r.FirstFailedAt, r.LastFailedAt = nullStrPtr(fa), nullStrPtr(la)
	return &r, nil
}

// AllRetryRows mirrors all_retry_rows (website_state.py:697-710) —
// EVERY waiting retry-queue row, oldest first (ORDER BY
// first_failed_at).
func (d *WebsiteStateDB) AllRetryRows() ([]RetryRow, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT url, attempts, last_error, next_attempt_at," +
			" first_failed_at, last_failed_at FROM" +
			" website_retry_queue ORDER BY first_failed_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []RetryRow
	for rows.Next() {
		var r RetryRow
		var le, na, fa, la sql.NullString
		if err := rows.Scan(&r.URL, &r.Attempts, &le, &na, &fa, &la); err != nil {
			return nil, err
		}
		r.LastError, r.NextAttemptAt = nullStrPtr(le), nullStrPtr(na)
		r.FirstFailedAt, r.LastFailedAt = nullStrPtr(fa), nullStrPtr(la)
		out = append(out, r)
	}
	return out, rows.Err()
}

// DismissedRows mirrors dismissed_rows (website_state.py:712-730) —
// the dismissals (optionally filtered by reason prefix), oldest first.
// A NULL reason reads as "" (Python's `reason or ”`).
func (d *WebsiteStateDB) DismissedRows(reasonPrefix string) (
	[]DismissedRow, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT url, reason, dismissed_at FROM dismissed_urls" +
			" ORDER BY dismissed_at")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DismissedRow
	for rows.Next() {
		var r DismissedRow
		var reason, at sql.NullString
		if err := rows.Scan(&r.URL, &reason, &at); err != nil {
			return nil, err
		}
		s := reason.String // NULL → "" (the `reason or ''` law)
		r.Reason = &s
		r.DismissedAt = nullStrPtr(at)
		if reasonPrefix != "" && !strings.HasPrefix(reason.String,
			reasonPrefix) {
			continue
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// DueRetries mirrors due_retries(now=None): retry-queue URLs whose
// backoff has elapsed (attempts < cap), judged against time.Now().
func (d *WebsiteStateDB) DueRetries() ([]string, error) {
	return d.DueRetriesAt(time.Now())
}

// DueRetriesAt is due_retries with an injected now (the Python
// signature's optional argument). Unparseable and offset-carrying
// next_attempt_at values count as due (the reference's own
// ValueError/TypeError catch paths — see the package doc).
func (d *WebsiteStateDB) DueRetriesAt(now time.Time) ([]string, error) {
	d.RLock()
	defer d.RUnlock()
	rows, err := d.SQL().Query(
		"SELECT url, attempts, next_attempt_at FROM website_retry_queue")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	type dueRow struct {
		url  string
		atmp int64
		next sql.NullString
	}
	var collected []dueRow
	for rows.Next() {
		var r dueRow
		if err := rows.Scan(&r.url, &r.atmp, &r.next); err != nil {
			return nil, err
		}
		collected = append(collected, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	var out []string
	for _, r := range collected {
		if r.atmp >= MaxFetchRetries {
			continue
		}
		if !r.next.Valid {
			out = append(out, r.url) // NULL → TypeError → due
			continue
		}
		t, aware, ok := parsePyFromISO(r.next.String)
		if !ok || aware {
			// unparseable → ValueError → due; offset-carrying → the
			// naive-vs-aware comparison raises TypeError → due.
			out = append(out, r.url)
			continue
		}
		if !t.After(now) { // <= now
			out = append(out, r.url)
		}
	}
	return out, nil
}

// parsePyFromISO approximates datetime.fromisoformat over the shapes
// the reference family writes. It returns (t, aware, ok): aware marks a
// value that carries a UTC offset (which due_retries' naive-vs-aware
// comparison turns into TypeError→due); ok=false marks unparseable
// (ValueError→due). For the DUE/NOT-DUE outcome every offset-carrying
// string is due either way (Python parses it aware → TypeError → due,
// or fails outright → ValueError → due), so a '+'/'-' inside the time
// portion or a trailing Z marks aware without further parsing.
func parsePyFromISO(s string) (t time.Time, aware bool, ok bool) {
	if len(s) > 10 {
		for i := 10; i < len(s); i++ {
			if s[i] == '+' || s[i] == '-' {
				return time.Time{}, true, true
			}
		}
		if c := s[len(s)-1]; c == 'Z' || c == 'z' {
			return time.Time{}, true, true
		}
	}
	layouts := []string{
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05.999999999",
		"2006-01-02 15:04:05.999999999",
		"2006-01-02T15:04",
		"2006-01-02 15:04",
		"2006-01-02",
	}
	for _, l := range layouts {
		if tt, err := time.ParseInLocation(l, s, time.Local); err == nil {
			return tt, false, true
		}
	}
	return time.Time{}, false, false
}

func nullStrPtr(v sql.NullString) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}
