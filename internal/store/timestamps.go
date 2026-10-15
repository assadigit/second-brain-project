package store

import (
	"fmt"
	"time"
)

// The naive/UTC timestamp helpers (ADR-0003 decision 4): Go writes the
// same string shapes Python wrote — naive-local strings without offset
// for the legacy tables, UTC-aware ISO strings for sync_deferrals.
// Grammar pinned byte-exact by testdata/timestamp_grammar.golden.json
// (generated from the reference interpreter's own datetime — see
// testdiff/goldens/dump_store.py).

// PyISOFormat mirrors Python datetime.isoformat() (timespec='auto'):
// YYYY-MM-DDTHH:MM:SS with a .ffffff fraction IFF the microsecond part
// is nonzero — the quirk the reference's now() writes rely on (cache_db
// processed_at / failed_at / decommissioned_at, link_tracker stamps).
// It formats t's own wall clock (callers pass time.Now() for the local
// naive law); nanoseconds truncate to Python's microsecond resolution.
func PyISOFormat(t time.Time) string {
	base := t.Format("2006-01-02T15:04:05")
	if us := t.Nanosecond() / 1000; us != 0 {
		return base + fmt.Sprintf(".%06d", us)
	}
	return base
}

// PyISOFormatSeconds mirrors datetime.isoformat(timespec='seconds'):
// YYYY-MM-DDTHH:MM:SS, never a fraction (website_state's
// websites_settled_at / note_state's corrections / cache_db's
// settle_repos and repos_settled_at).
func PyISOFormatSeconds(t time.Time) string {
	return t.Format("2006-01-02T15:04:05")
}

// PyISOUTCSeconds mirrors
// datetime.now(timezone.utc).isoformat(timespec='seconds'):
// YYYY-MM-DDTHH:MM:SS+00:00 — the UTC-aware shape sync_review's
// sync_deferrals write. t is CONVERTED to UTC first (the same instant,
// expressed in UTC — the astimezone(timezone.utc) semantics).
func PyISOUTCSeconds(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05") + "+00:00"
}
