package store

import (
	"database/sql"
	"sync"

	// Pure-Go SQLite driver (ADR-0003: no CGO anywhere; doc 06 §2 pins
	// this dependency). Registered as "sqlite".
	_ "modernc.org/sqlite"
)

// dsnPragmas — the ADR-0003 DSN parameters, applied by the modernc
// driver to every connection it opens (with MaxOpenConns(1): exactly the
// one connection the Python law describes).
const dsnPragmas = "_pragma=busy_timeout(30000)" +
	"&_pragma=journal_mode(WAL)" +
	"&_pragma=synchronous(NORMAL)"

// DSN returns the modernc.org/sqlite data-source name for a database
// path: the path plus the ADR-0003 pragmas. Exposed so tests (and, if a
// future ADR needs it, special openers) can assert the exact shape.
func DSN(path string) string {
	return path + "?" + dsnPragmas
}

// DB is one SQLite store: the single connection (as *sql.DB with
// MaxOpenConns(1)) plus the per-store guard. The zero value is not
// usable; Open constructs it.
type DB struct {
	// sync.RWMutex is THE per-store guard (doc 04 §3). Embedded so a
	// ported store writes d.Lock()/d.RLock() directly — the Go shape of
	// Python's `with self._lock:`. NOT reentrant (see the package law in
	// doc.go): public entry points acquire, private helpers run free.
	sync.RWMutex

	path string
	db   *sql.DB
}

// Open opens (creating if absent — Python's sqlite3.connect semantics)
// one store on path with the ADR-0003 pragmas. The connection is opened
// eagerly so a bad path fails here, exactly like Python's connect.
func Open(path string) (*DB, error) {
	db, err := sql.Open("sqlite", DSN(path))
	if err != nil {
		return nil, err
	}
	// The locking law's connection shape: exactly one, kept forever
	// (Python's persistent single connection).
	db.SetMaxOpenConns(1)
	db.SetMaxIdleConns(1)
	db.SetConnMaxLifetime(0)
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, err
	}
	return &DB{path: path, db: db}, nil
}

// Close is idempotent (the v30 law; ported test: TestCloseIdempotent).
// The underlying connection pool is released; SQL() returns nil after.
func (d *DB) Close() error {
	if d == nil || d.db == nil {
		return nil
	}
	err := d.db.Close()
	d.db = nil
	return err
}

// SQL returns the database handle for statements. Callers hold the guard
// for multi-statement sequences and collect rows before nesting (the
// package law in doc.go). Nil after Close.
func (d *DB) SQL() *sql.DB { return d.db }

// Path returns the path this store was opened with (the dry-run seam's
// call sites already resolved any shadowing before Open).
func (d *DB) Path() string { return d.path }
