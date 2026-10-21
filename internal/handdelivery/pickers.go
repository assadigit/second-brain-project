// pickers.go — the picker's data (walled_retry_rows,
// hand_delivery.py:177-192) and the fifth door's to-do list
// (pending_hand_links, hand_delivery.py:746-773).
package handdelivery

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// RetrySourceRow mirrors one all_retry_rows() row as the picker reads
// it: {'url', 'last_error', 'attempts'}.
type RetrySourceRow struct {
	URL       string
	LastError string
	Attempts  int
}

// RetryRowsState is the walled_retry_rows state seam: the reference
// takes the pipeline's WebsiteStateDB duck-typed and calls
// all_retry_rows() on it. websitepipeline imports this package (the
// reference dodges the cycle with function-level imports; Go has no
// deferred imports — the decomtable law), so the concrete StateDB
// adapter lives with the call sites (the GUI picker / the CLI hand
// leg); the battery and the golden replay the same row shape through
// fakes (the reference's own _State pattern in test_handdelivery).
type RetryRowsState interface {
	// AllRetryRows mirrors all_retry_rows(): every waiting
	// retry-queue row, oldest first.
	AllRetryRows() ([]RetrySourceRow, error)
}

// WalledRetryRow mirrors one walled_retry_rows() output row:
// {'url', 'wall', 'attempts'}.
type WalledRetryRow struct {
	URL      string
	Wall     string
	Attempts int
}

// WalledRetryRows mirrors walled_retry_rows: the picker's data —
// retry-queue rows whose last error names a wall (the three machine
// doors tried and failed). Never raises — a state DB hiccup means an
// empty picker, not a crash.
func WalledRetryRows(state RetryRowsState) []WalledRetryRow {
	if state == nil {
		return nil
	}
	rows, err := func() (rows []RetrySourceRow, err error) {
		defer func() {
			if r := recover(); r != nil {
				rows, err = nil, fmt.Errorf("%v", r)
			}
		}()
		return state.AllRetryRows()
	}()
	if err != nil {
		return nil
	}
	var out []WalledRetryRow
	for _, row := range rows {
		u := strings.TrimSpace(row.URL)
		if u == "" {
			continue
		}
		if !IsWalledReason(row.LastError) {
			continue
		}
		out = append(out, WalledRetryRow{
			URL:      u,
			Wall:     row.LastError,
			Attempts: row.Attempts,
		})
	}
	return out
}

// PendingHandLink mirrors one pending_hand_links() row: {'url',
// 'wall'} — a queued link whose page has NOT landed yet.
type PendingHandLink struct {
	URL  string
	Wall string
}

// PendingHandLinks mirrors pending_hand_links (v0.52.0): the queued
// hand links whose page has NOT landed yet — every queue.json entry
// that is neither consumed nor sitting in the hand-delivered folder
// as its suggested file. This is the fifth door's to-do list (the
// Chrome scraping pass reads it to know which walled links still owe
// the owner's browser a tab). Pure file reads; never raises (a broken
// queue reads as empty — the table's own scan,
// websitepipeline.ScanMasterHandRows, is the fuller source that also
// catches never-enqueued rows).
func PendingHandLinks(vaultPath string) []PendingHandLink {
	var out []PendingHandLink
	queue := ReadQueue(vaultPath)
	folder := HandDeliveryDir(vaultPath)
	for _, u := range queue.Links.Keys() {
		m := queue.Meta(u)
		if m == nil || MetaTruthy(m, "consumed") {
			continue
		}
		suggested := MetaStr(m, "suggested")
		if suggested == "" {
			suggested = SuggestedFilename(u)
		}
		if st, err := os.Stat(filepath.Join(folder, suggested)); err == nil && !st.IsDir() {
			continue // the page landed — it is delivered
		}
		out = append(out, PendingHandLink{URL: u, Wall: MetaStr(m, "wall")})
	}
	return out
}
