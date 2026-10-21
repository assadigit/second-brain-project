package websitepipeline

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/webfetch"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// decommission.go — the graveyard table's WRITERS and SCANS
// (website_pipeline.py:721-3329): the candidates writer, the dead
// marker, the stored stamp, the prune, the waiting/hand/redo scans,
// note_is_properly_stored, and the hand harvest. The parse and the
// status predicates live in decomtable (the shared grammar floor).

// DecommissionTablePath mirrors decommission_table_path.
func DecommissionTablePath(vaultPath string) string {
	return decomtable.DecommissionTablePath(vaultPath)
}

// readTableLines reads the table file, split Python-splitlines-wise.
func readTableLines(path string) ([]string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	return pySplitLines(decodeUTF8Replace(string(data))), true
}

// ParseDecommissionRows is the package-level view of the grammar
// floor's parse.
func ParseDecommissionRows(lines []string) []decomtable.Row {
	return decomtable.ParseDecommissionRows(lines)
}

// GraveyardHeader mirrors _GRAVEYARD_HEADER (the table's legend —
// byte-stable except the {now} timestamp).
const GraveyardHeader = `# Review Master Table — decommission or approve

> The master note for EVERY link waiting in _review — fetch failures
> waiting for retries AND the human-eye classes (low classification
> confidence, analysis failures, archived rescues): one row per link,
> every link with a Status cell you can set. Edit ONLY the Status
> column — do not delete rows. The # column is yours too: an emoji in
> the FIRST cell of a row (🖐 / 💀 / ✅ / ♻️) reads exactly like the
> same emoji in the Status cell — set it wherever is faster.
> ✅ reviewed — you handled this link: never fetched, never retried
>   again, its _review placeholder is swept.
> 🪦 dead / ❌ dead / ☠️ dead / 💀 dead — the link is dead: same
>   never-fetch retirement, same placeholder sweep.
> 🗑️ banished / delete / auto_delete — the site outlived its welcome
>   (terms changed, no longer free, no longer useful): the STRONGEST
>   retirement — the note itself is REMOVED from the library
>   (recoverable in .trash/banished) AND the URL is blacklisted,
>   never fetched again. ♻️ revived undoes it. You can also set
>   this verdict INSIDE the note itself: open any note and add
>   the tag 🗑️ (or "delete" / "auto_delete") — the next run shows
>   you the count and ASKS first (the Telegram deletion review:
>   "N notes marked for deletion" — 🗑️ Delete / ✋ Keep); on your
>   confirm it removes the note, blacklists the URL, and writes
>   the record row here. Every run answers with the tally line:
>   how many were removed and never fetched again because you
>   confirmed them.
> ♻️ revived (in place of a dead or ✅ Status) — the link is fetched
>   like new again.
> 🖐 hand — the fourth door's gesture, the fifth door's engine: set
>   it (the # cell or the Status cell) and GitCurator itself opens
>   your real Chrome, one tab per hand row, takes the live pages, and
>   processes them as real fetches — no manual saving (you can still
>   save a page into _review/hand-delivered/ yourself; a delivered
>   page is consumed either way). NOT a retirement — the link keeps
>   waiting until the page is delivered. When the delivery STORES
>   the proper, categorized note, the row retires itself:
>   ✅ hand-delivered — fetched <date> (the half-fetched _review
>   item is swept — the note in the vault is the record).
> blank / unreviewed / " - " — still waiting: the caught-up check
>   retries these rows BEFORE "everything is up to date" is said
>   (the link is valid, its fetch failed — the machine tries it
>   again; what still fails keeps waiting, its Status decides).
> 📁 stored — the retry succeeded: the link's note is in the vault
>   (the row is the record, nothing more to do).
> Rows marked "auto" were retired by the fetcher's own verdict
> (dead / paywalled / refused) — set ♻️ revived to disagree.
> v0.60.2 — THE COMPACT TABLE: rows whose verdict is WRITTEN
>   (confirmed / banished / stored / hand-delivered / auto) LEAVE this
>   table after every batch — their records live in the state DB and
>   .trash/banished, and this table keeps only the links that still
>   need you. To REVIVE any retired link (bring it back to be fetched
>   like new), add a row with ♻️ and its URL — any shape works:
>
>     | ♻️ | | https://the-link.example/its-path | | | | |
>
>   (indented here so it stays an instruction, not a row) — the next
>   run reads your row and un-retires the link.
> Auto-updated after every batch. Last updated: %s

| # | Date | URL | Domain | Source | Status | Notes |
|---|------|-----|--------|--------|--------|-------|
`

// ScanDecommissionTable mirrors scan_decommission_table (v0.44.0):
// {url: status} for EVERY data row — first row wins on a hand-added
// duplicate URL. Pure file read; no state DB, no network.
func ScanDecommissionTable(vaultPath string) map[string]string {
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return map[string]string{}
	}
	if _, err := os.Stat(path); err != nil {
		return map[string]string{}
	}
	lines, ok := readTableLines(path)
	if !ok {
		return map[string]string{}
	}
	out := map[string]string{}
	for _, row := range ParseDecommissionRows(lines) {
		if _, exists := out[row.URL]; !exists {
			out[row.URL] = row.Status
		}
	}
	return out
}

// EnsureTableShape mirrors _ensure_table_shape (v0.52.0): restore the
// table's full grammar when the owner's hand has trimmed it — pad the
// header + separator back to the canonical seven columns and pad every
// short data row to the same width (the owner's cells are never
// touched, only empty cells are appended). Idempotent; atomic; never
// raises. Returns True when the file was rewritten.
func EnsureTableShape(path string, log LogFn) bool {
	if log == nil {
		log = noopLog
	}
	if path == "" {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	lines, ok := readTableLines(path)
	if !ok {
		log("⚠️ Table shape check skipped — the table could not be read", "warning")
		return false
	}
	changed := false
	// the header: the first pipe-row that is not a separator
	headerIdx := -1
	for i, line := range lines {
		if strings.HasPrefix(line, "|") && !decomtable.IsSeparatorLine(line) {
			headerIdx = i
			break
		}
	}
	if headerIdx >= 0 {
		headerCells := 0
		if strings.HasSuffix(strings.TrimSpace(lines[headerIdx]), "|") {
			headerCells = len(strings.Split(lines[headerIdx], "|")) - 2
		} else {
			headerCells = len(strings.Split(lines[headerIdx], "|")) - 1
		}
		if headerCells < decomtable.TableHeaderCells {
			lines[headerIdx] = decomtable.CanonicalHeader
			if headerIdx+1 < len(lines) && decomtable.IsSeparatorLine(lines[headerIdx+1]) {
				lines[headerIdx+1] = decomtable.CanonicalSeparator
			}
			changed = true
			log("📋 Master table: the trimmed header is restored to the full seven columns (the # column is still yours — the emoji there reads exactly like the Status cell)", "info")
		}
	}
	// the data rows: pad every short pipe-row that carries a link
	for i, line := range lines {
		if !(strings.HasPrefix(line, "| ") && strings.Contains(line, "http")) {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 2+decomtable.TableHeaderCells {
			for len(parts) < 2+decomtable.TableHeaderCells {
				parts = append(parts, "")
			}
			lines[i] = strings.Join(parts, "|")
			changed = true
		}
	}
	if !changed {
		return false
	}
	if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(lines, "\n"), "\n")+"\n"); err != nil {
		log(fmt.Sprintf("⚠️ Table shape could not be restored: %v", err), "warning")
		return false
	}
	return true
}

// dateNow mirrors time.strftime('%Y-%m-%d') (local time).
func dateNow() string { return clockNow().Format("2006-01-02") }

// stampNow mirrors time.strftime('%Y-%m-%d %H:%M') (local time).
func stampNow() string { return clockNow().Format("2006-01-02 15:04") }

// WriteDecommissionCandidates mirrors write_decommission_candidates
// (v0.44.0): pre-fill the graveyard with the failed links waiting in
// _review so the owner only has to SET THE EMOJI. Rows are APPENDED
// after the last data row; URLs already present (any Status) are never
// duplicated, owner-edited rows are never touched. Creates the table
// with its header when the file does not exist. Atomic + dry-run
// aware. Returns how many new rows were written.
func WriteDecommissionCandidates(vaultPath string, urls []string,
	source string, log LogFn, notes map[string]string, status string) int {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" {
		return 0
	}
	path := DecommissionTablePath(vaultPath)
	reviewDir := filepath.Join(vaultPath, ReviewFolder)
	if err := dryrun.MkdirAll(reviewDir, true); err != nil {
		log(fmt.Sprintf("⚠️ Could not create %s: %v", reviewDir, err), "warning")
		return 0
	}
	var rows []decomtable.Row
	if lines, ok := readTableLines(path); ok {
		rows = ParseDecommissionRows(lines)
	}
	existing := map[string]bool{}
	for _, r := range rows {
		existing[r.URL] = true
	}
	dateStr := dateNow()
	if notes == nil {
		notes = map[string]string{}
	}
	statusCell := decomtable.TableCell(status, 48)
	if statusCell == "" {
		statusCell = "unreviewed"
	}
	var newRows []string
	for _, u := range urls {
		if u == "" || existing[u] {
			continue
		}
		existing[u] = true
		domain := "unknown"
		if parsed, err := url.Parse(u); err == nil && parsed.Host != "" {
			domain = parsed.Host
		}
		note := decomtable.TableCell(notes[u], 96)
		row := fmt.Sprintf("| - | %s | %s | %s | %s | %s",
			dateStr, links.ScrubURLToken(u),
			decomtable.TableCell(domain, 64),
			decomtable.TableCell(source, 32), statusCell)
		if note != "" {
			row += fmt.Sprintf(" | %s |", note)
		} else {
			row += " | |"
		}
		newRows = append(newRows, row)
	}
	if len(newRows) == 0 {
		return 0
	}
	now := stampNow()
	if _, err := os.Stat(path); err == nil {
		lines, ok := readTableLines(path)
		if !ok {
			log("⚠️ Could not read the graveyard table", "warning")
			return 0
		}
		lastData := 0
		for i, line := range lines {
			if strings.HasPrefix(line, "| ") && strings.Contains(line, "http") {
				lastData = i
			}
		}
		lines = append(lines[:lastData+1], append(append([]string{}, newRows...), lines[lastData+1:]...)...)
		for i, line := range lines {
			if strings.HasPrefix(line, "> Last updated:") {
				lines[i] = "> Last updated: " + now
			}
		}
		if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(lines, "\n"), "\n")+"\n"); err != nil {
			log(fmt.Sprintf("⚠️ Could not write the graveyard table: %v", err), "warning")
			return 0
		}
	} else {
		content := fmt.Sprintf(GraveyardHeader, now) + strings.Join(newRows, "\n") + "\n"
		if err := storage.AtomicWriteText(path, content); err != nil {
			log(fmt.Sprintf("⚠️ Could not write the graveyard table: %v", err), "warning")
			return 0
		}
	}
	log(fmt.Sprintf("📋 %d row(s) written to %s (the master table — set Status to 🪦 dead to bury or ✅ reviewed to keep; ♻️ revived brings one back)", len(newRows), DecommissionTable), "info")
	return len(newRows)
}

// MarkURLsDeadInTable mirrors mark_urls_dead_in_table (v0.44.0): set
// the burial marker on chosen rows (the in-app picker's hand — it
// writes the SAME Status cell the owner would edit by hand in
// Obsidian, so the table stays the one ledger). Rows whose URL is not
// in the table are appended as dead first. Returns how many rows were
// marked.
func MarkURLsDeadInTable(vaultPath string, urls []string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" {
		return 0
	}
	path := DecommissionTablePath(vaultPath)
	wanted := dedupeKeepOrder(urls)
	if len(wanted) == 0 {
		return 0
	}
	// Append the missing ones as fresh rows, then flip every Status.
	WriteDecommissionCandidates(vaultPath, wanted, "picker", nil, nil, "unreviewed")
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	lines, ok := readTableLines(path)
	if !ok {
		return 0
	}
	rows := ParseDecommissionRows(lines)
	if len(rows) == 0 {
		return 0
	}
	dateStr := dateNow()
	wantedSet := map[string]bool{}
	for _, u := range wanted {
		wantedSet[u] = true
	}
	marked := 0
	for _, row := range rows {
		if !wantedSet[row.URL] {
			continue
		}
		parts := strings.Split(row.Raw, "|")
		if len(parts) < 8 {
			continue
		}
		parts[6] = fmt.Sprintf(" 🪦 dead — decommissioned %s ", dateStr)
		if row.Line < len(lines) {
			lines[row.Line] = strings.Join(parts, "|")
			marked++
		}
	}
	now := stampNow()
	for i, line := range lines {
		if strings.HasPrefix(line, "> Last updated:") {
			lines[i] = "> Last updated: " + now
		}
	}
	if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(lines, "\n"), "\n")+"\n"); err != nil {
		log(fmt.Sprintf("⚠️ Could not write the graveyard table: %v", err), "warning")
		return 0
	}
	if marked > 0 {
		log(fmt.Sprintf("🪦 %d link(s) marked dead in %s", marked, DecommissionTable), "info")
	}
	return marked
}

// StateDB is the scans' state seam — the concrete WebsiteStateDB
// (tests ride real temp-file DBs, the reference suites' own pattern;
// the reference's Optional[state] becomes a nilable pointer).
type StateDB = WebsiteStateDB

// StampStoredRows mirrors stamp_stored_rows (v0.51.0 — the truth
// stamp): a master-table row whose Status still reads as waiting while
// its link is ALREADY STORED (a processed row with a real note outside
// _review) is rewritten to "📁 stored — <date>". Owner-set verdicts
// are NEVER touched; a broken state probe stamps nothing (never
// crashes a batch). Returns how many rows were stamped.
func StampStoredRows(state *StateDB, vaultPath string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return 0
	}
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	lines, ok := readTableLines(path)
	if !ok {
		return 0
	}
	rows := ParseDecommissionRows(lines)
	if len(rows) == 0 {
		return 0
	}
	dateStr := dateNow()
	stamped := 0
	var outLines []string
	for _, row := range rows {
		if !StatusIsWaiting(row.Status) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			continue
		}
		dismissed, err := state.IsDismissed(canonical)
		if err != nil {
			continue // a broken probe stamps nothing
		}
		if dismissed {
			continue // retired rows are the verdicts' business
		}
		prior, err := state.ProcessedRow(canonical)
		if err != nil || prior == nil {
			continue
		}
		if prior.FetchStatus != nil && *prior.FetchStatus == "failed" {
			continue // the placeholder story — still waiting
		}
		if strings.Contains(strings.ReplaceAll(ptrStr(prior.NotePath), "\\", "/"), "/_review/") {
			continue // a _review note waits for human eyes
		}
		parts := strings.Split(row.Raw, "|")
		if len(parts) < 8 {
			continue
		}
		parts[6] = fmt.Sprintf(" 📁 stored — fetched %s ", dateStr)
		if outLines == nil {
			outLines = lines
		}
		if row.Line < len(outLines) {
			outLines[row.Line] = strings.Join(parts, "|")
			stamped++
		}
	}
	if stamped > 0 {
		now := stampNow()
		for i, line := range outLines {
			if strings.HasPrefix(line, "> Last updated:") {
				outLines[i] = "> Last updated: " + now
			}
		}
		if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(outLines, "\n"), "\n")+"\n"); err != nil {
			log(fmt.Sprintf("⚠️ Could not write the master table: %v", err), "warning")
			return 0
		}
		log(fmt.Sprintf("📁 %d master-table row(s) stamped 'stored' — their retries succeeded, the notes are in the vault", stamped), "info")
	}
	return stamped
}

// PruneReport mirrors prune_decommission_table's return:
// {'pruned', 'kept', 'pruned_urls'}.
type PruneReport struct {
	Pruned     int
	Kept       int
	PrunedURLs []string
}

// PruneDecommissionTable mirrors prune_decommission_table (v0.60.2):
// the table keeps only what still needs the owner — every terminal
// row (verdicts already written and enforced) LEAVES; the pending ones
// stay. The header and its legend are never touched. Pure file
// rewrite, atomic + dry-run aware; never raises.
func PruneDecommissionTable(vaultPath string, log LogFn) PruneReport {
	if log == nil {
		log = noopLog
	}
	out := PruneReport{}
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return out
	}
	if _, err := os.Stat(path); err != nil {
		return out
	}
	lines, ok := readTableLines(path)
	if !ok {
		log("⚠️ Master-table prune skipped — the table could not be read", "warning")
		return out
	}
	rows := ParseDecommissionRows(lines)
	if len(rows) == 0 {
		return out // nothing data-shaped — the legend stands
	}
	dropLines := map[int]bool{}
	for _, row := range rows {
		if decomtable.StatusIsTerminal(row.Status) {
			dropLines[row.Line] = true
			out.Pruned++
			cu := links.NormalizeWebsiteURL(row.URL)
			if cu == "" {
				cu = row.URL
			}
			out.PrunedURLs = append(out.PrunedURLs, cu)
		} else {
			out.Kept++
		}
	}
	if out.Pruned == 0 {
		return out // a pending-only table is already the law
	}
	var keptLines []string
	for i, ln := range lines {
		if !dropLines[i] {
			keptLines = append(keptLines, ln)
		}
	}
	now := stampNow()
	for i, ln := range keptLines {
		if strings.HasPrefix(ln, "> Auto-updated after every batch.") {
			keptLines[i] = "> Auto-updated after every batch. Last updated: " + now
		}
	}
	if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(keptLines, "\n"), "\n")+"\n"); err != nil {
		log(fmt.Sprintf("⚠️ Master-table prune could not be written: %v", err), "warning")
		return PruneReport{Kept: out.Kept + out.Pruned}
	}
	if out.Pruned > 0 {
		log(fmt.Sprintf("📋 Master table pruned: %d terminal row(s) left the table (%d pending row(s) stay) — their verdicts live in the state DB and .trash/banished; to revive any of them, add a row with ♻️ and the URL (the legend above shows the shape)", out.Pruned, out.Kept), "info")
	}
	return out
}

// StatusIsWaiting mirrors _status_is_waiting (v0.51.0): a Status cell
// that carries NO decision at all — blank, ' - ', '—', 'unreviewed',
// or any text that is neither dead nor reviewed nor revived nor the
// fourth door's queue nor one of the app's own stamps.
func StatusIsWaiting(status string) bool {
	s := strings.TrimSpace(status)
	if s == "" {
		return true // blank — still waiting
	}
	if decomtable.StatusIsDead(s) || decomtable.StatusIsReviewed(s) || decomtable.StatusIsRevived(s) {
		return false
	}
	if decomtable.StatusIsBanished(s) {
		return false // v0.58.0 — 🗑️ owns the row
	}
	if handStatusIsHand(s) {
		return false // the fourth door owns it
	}
	low := strings.ToLower(s)
	if strings.Contains(low, "queued") || strings.Contains(low, "confirmed") || strings.Contains(low, "auto") {
		return false // a stamp the app itself wrote
	}
	if strings.Contains(low, "stored") {
		return false // v0.51.0 — the retry succeeded
	}
	return true
}

// handStatusIsHand carries the reference's
// _hand_delivery._status_is_hand call site.
func handStatusIsHand(s string) bool { return handdelivery.StatusIsHand(s) }

// WaitingRow mirrors scan_master_waiting_rows' items:
// {'url', 'status', 'notes', 'path', 'kind'}.
type WaitingRow struct {
	URL    string
	Status string
	Notes  string
	Path   string
	Kind   string // 'fetch' | 'eyes'
}

// ScanMasterWaitingRows mirrors scan_master_waiting_rows (v0.51.0):
// the " - " rows of the master table, classified — kind 'fetch' (the
// machine's to retry) vs 'eyes' (the owner's to decide). state is
// OPTIONAL (the hermetic law): without it the scan is a pure file read
// and the queue/stored probes are skipped. First row wins on a
// hand-added duplicate URL, sorted by the table's own row order.
// Never raises on a hand-edited table.
func ScanMasterWaitingRows(vaultPath string, state *StateDB, log LogFn) []WaitingRow {
	if log == nil {
		log = noopLog
	}
	var out []WaitingRow
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return out
	}
	if _, err := os.Stat(path); err != nil {
		return out
	}
	// one pass over _review pairs every app-owned note with its source
	notes := map[string]ReviewNoteItem{}
	for _, it := range ScanReviewNotes(vaultPath, nil) {
		cu := links.NormalizeWebsiteURL(it.URL)
		if cu != "" {
			if _, exists := notes[cu]; !exists {
				notes[cu] = it
			}
		}
	}
	lines, ok := readTableLines(path)
	if !ok {
		return out
	}
	seen := map[string]bool{}
	for _, row := range ParseDecommissionRows(lines) {
		if !StatusIsWaiting(row.Status) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" || seen[canonical] {
			continue
		}
		note, hasNote := notes[canonical]
		kind := "fetch"
		placeholder := ""
		if hasNote {
			placeholder = note.Path
			if !strings.EqualFold(strings.TrimSpace(note.FetchStatus), "failed") {
				kind = "eyes" // a human-eye note — not ours
			}
		}
		if state != nil && kind == "fetch" {
			dismissed, err := state.IsDismissed(canonical)
			if err != nil {
				log(fmt.Sprintf("⚠️ Master-table state probe skipped for %s: %v", row.URL, err), "warning")
			} else if dismissed {
				continue // retired — ♻️ is its door back
			} else {
				settled, err := state.IsSettled(canonical)
				if err == nil && settled {
					continue // v0.63.2 — settled, never re-fetched
				}
				prior, err := state.ProcessedRow(canonical)
				if err == nil && prior != nil && prior.FetchStatus != nil &&
					*prior.FetchStatus != "failed" &&
					!strings.Contains(strings.ReplaceAll(ptrStr(prior.NotePath), "\\", "/"), "_review") {
					continue // already stored — not waiting
				}
			}
		}
		seen[canonical] = true
		out = append(out, WaitingRow{
			URL: row.URL, Status: row.Status, Notes: row.Notes,
			Path: placeholder, Kind: kind,
		})
	}
	return out
}

// HandRow mirrors scan_master_hand_rows' items:
// {'url', 'wall', 'status'}.
type HandRow struct {
	URL    string
	Wall   string
	Status string
}

// ScanMasterHandRows mirrors scan_master_hand_rows (v0.52.0): the
// master table's 🖐 hand rows, the ones the fifth door owes a Chrome
// scrape to. Pure file reads (the table + the fourth door's queue) —
// no state DB, no network; never raises on a hand-edited table.
func ScanMasterHandRows(vaultPath string, log LogFn) []HandRow {
	var out []HandRow
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return out
	}
	if _, err := os.Stat(path); err != nil {
		return out
	}
	// the delivered probe: a queue entry marked consumed, or whose
	// suggested page file already sits in the hand-delivered folder.
	delivered := deliveredURLSet(vaultPath)
	seen := map[string]bool{}
	lines, ok := readTableLines(path)
	if !ok {
		return out
	}
	for _, row := range ParseDecommissionRows(lines) {
		s := row.Status
		if decomtable.StatusIsDead(s) || decomtable.StatusIsReviewed(s) || decomtable.StatusIsRevived(s) {
			continue
		}
		if !handStatusIsHand(s) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" || seen[canonical] {
			continue
		}
		if delivered[canonical] {
			continue // the page already landed — it is consumed
		}
		if webfetch.IsLoopbackURL(canonical) {
			continue // never opens a browser (v0.15.1)
		}
		seen[canonical] = true
		out = append(out, HandRow{URL: row.URL, Wall: row.Notes, Status: s})
	}
	return out
}

// RedoRow mirrors scan_master_redo_rows' items: {'url', 'reason'}.
type RedoRow struct {
	URL    string
	Reason string
}

// ScanMasterRedoRows mirrors scan_master_redo_rows (v0.57.0 — THE
// NOTE IS THE SUCCESS, the redo scan): the 🖐 hand rows whose delivery
// was a FALSE SUCCESS — the page was delivered but the note never
// became a proper, categorized note. state=None degrades honestly:
// nothing can be proven false, so nothing is redone. Pure file +
// state reads; never raises on a hand-edited table.
func ScanMasterRedoRows(vaultPath string, state *StateDB, log LogFn) []RedoRow {
	if log == nil {
		log = noopLog
	}
	var out []RedoRow
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return out
	}
	if _, err := os.Stat(path); err != nil {
		return out
	}
	if state == nil {
		return out // nothing can be proven false — the hermetic law
	}
	delivered := deliveredURLSet(vaultPath)
	seen := map[string]bool{}
	lines, ok := readTableLines(path)
	if !ok {
		return out
	}
	for _, row := range ParseDecommissionRows(lines) {
		s := row.Status
		if decomtable.StatusIsDead(s) || decomtable.StatusIsReviewed(s) || decomtable.StatusIsRevived(s) {
			continue
		}
		if !handStatusIsHand(s) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" || seen[canonical] {
			continue
		}
		if !delivered[canonical] {
			continue // no delivery yet — the Chrome pass's set
		}
		// the probe ladder (the reference's try/except around
		// is_dismissed + processed_row): a broken probe — a query
		// error OR a panicked handle (the closed-DB shape; SQL() is nil
		// after Close by the store's own law, and the reference's
		// _Broken raises the Python-class twin) — redoes nothing.
		prior, probeBroken := func() (prior *ProcessedRow, broken bool) {
			defer func() {
				if recover() != nil {
					broken = true
				}
			}()
			dismissed, err := state.IsDismissed(canonical)
			if err != nil || dismissed {
				return nil, true // retired rows are the verdicts' business
			}
			prior, err = state.ProcessedRow(canonical)
			if err != nil {
				return nil, true
			}
			return prior, false
		}()
		if probeBroken {
			continue
		}
		proper, why := NoteIsProperlyStored(prior)
		if proper {
			continue // the note IS properly stored — done, honestly
		}
		if why == "" {
			why = "the note is not properly stored"
		}
		seen[canonical] = true
		out = append(out, RedoRow{URL: canonical, Reason: why})
	}
	return out
}

// deliveredURLSet mirrors the delivered probe the reference INLINES in
// BOTH master scans (scan_master_hand_rows + scan_master_redo_rows —
// NOT collect_delivered, the batch side's own skip-consumed collector:
// the SESSION-059 find, proven by the ported TestRedoScan selections):
// a queue entry marked consumed, OR whose suggested page file already
// sits in the hand-delivered folder (os.path.isfile — NO size check:
// a 0-byte save still landed; collect_delivered's getsize>0 is the
// BATCH side's law). (meta or {}): a nil-meta row still checks its
// suggested page file. A broken queue degrades to the empty set behind
// the never-raise law (a broken queue never hides a hand row).
func deliveredURLSet(vaultPath string) map[string]bool {
	delivered := map[string]bool{}
	func() {
		defer func() {
			if recover() != nil {
				delivered = map[string]bool{}
			}
		}()
		queue := handdelivery.ReadQueue(vaultPath)
		folder := handdelivery.HandDeliveryDir(vaultPath)
		for _, u := range queue.Links.Keys() {
			m := queue.Meta(u)
			suggested := handdelivery.MetaStr(m, "suggested")
			if suggested == "" {
				suggested = handdelivery.SuggestedFilename(u)
			}
			pageLanded := false
			if st, err := os.Stat(filepath.Join(folder, suggested)); err == nil && !st.IsDir() {
				pageLanded = true
			}
			if handdelivery.MetaTruthy(m, "consumed") || pageLanded {
				delivered[links.NormalizeWebsiteURL(u)] = true
			}
		}
	}()
	return delivered
}

// NoteIsProperlyStored mirrors note_is_properly_stored (v0.57.0 — THE
// NOTE IS THE SUCCESS, the test): does a state row's note amount to a
// PROPER, CATEGORIZED note in the vault? Each half-fetched shape fails
// with its own one-line reason. Pure; never raises (anything
// unreadable is simply not proper).
func NoteIsProperlyStored(prior *ProcessedRow) (bool, string) {
	if prior == nil {
		return false, "no processed row — the note never landed"
	}
	if prior.FetchStatus != nil && *prior.FetchStatus == "failed" {
		return false, "the row records a failed fetch"
	}
	path := strings.ReplaceAll(ptrStr(prior.NotePath), "\\", "/")
	if path == "" {
		return false, "no note path recorded"
	}
	if strings.Contains("/"+strings.Trim(path, "/")+"/", "/_review/") {
		return false, "the note is half-fetched in _review"
	}
	if _, err := os.Stat(path); err != nil {
		return false, "the note file is missing on disk"
	}
	fm := ParseReviewFrontmatter(path)
	if fm == nil || !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
		return false, "the note is not app-owned"
	}
	source := strings.TrimSpace(fm.Source)
	if source == "" {
		return false, "the note carries no source link"
	}
	own := ""
	if prior.URL != "" {
		own = links.NormalizeWebsiteURL(prior.URL)
	}
	if own != "" && links.NormalizeWebsiteURL(source) != own {
		// v0.64.0 — ONE NOTE PER SITE: a note whose site_links list
		// carries this link is a PROPER delivery too — the link
		// consolidated into its site's note. The site key must agree.
		if !(own != "" && links.SiteKeyOf(own) != "" &&
			links.SiteKeyOf(own) == links.SiteKeyOf(source) &&
			NoteListsSiteLink(path, own)) {
			return false, "the note belongs to another link"
		}
	}
	return true, ""
}

// HarvestHandRows mirrors harvest_hand_rows (v0.56.0 — THE HAND'S
// HARVEST, the catch-up pass): a 🖐 row whose link is ALREADY STORED
// (note_is_properly_stored — v0.57.0's law) is retired to
// "✅ hand-delivered" and its half-fetched _review items are swept.
// The state probe is optional and a broken probe harvests nothing.
// Returns rows harvested.
func HarvestHandRows(state *StateDB, vaultPath string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return 0
	}
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	var retired []string
	lines, ok := readTableLines(path)
	if !ok {
		return 0
	}
	for _, row := range ParseDecommissionRows(lines) {
		s := row.Status
		if decomtable.StatusIsDead(s) || decomtable.StatusIsReviewed(s) || decomtable.StatusIsRevived(s) {
			continue // a stronger verdict owns the cell
		}
		if !handStatusIsHand(s) {
			continue // only the gesture retires here
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			continue
		}
		already := false
		for _, r := range retired {
			if r == canonical {
				already = true
				break
			}
		}
		if already {
			continue
		}
		// the probe ladder (the reference's ONE try/except around
		// is_dismissed + processed_row — "a broken probe harvests
		// nothing"; the same twin the redo scan rides): a broken
		// probe — a query error OR a panicked handle (the closed-DB
		// shape: SQL() is nil after Close by the store's own law, and
		// the reference's _Broken raises the Python-class twin) —
		// harvests nothing. state=nil is the same class (the hermetic
		// law: the probe is optional — the nil receiver panics inside
		// the probe, and the ladder answers the same broken verdict).
		prior, probeBroken := func() (prior *ProcessedRow, broken bool) {
			defer func() {
				if recover() != nil {
					broken = true
				}
			}()
			dismissed, err := state.IsDismissed(canonical)
			if err != nil || dismissed {
				return nil, true // retired rows are the verdicts' business
			}
			prior, err = state.ProcessedRow(canonical)
			if err != nil {
				return nil, true
			}
			return prior, false
		}()
		if probeBroken {
			continue
		}
		proper, _ := NoteIsProperlyStored(prior)
		if prior == nil || !proper {
			continue // no proper, categorized note — no harvest
		}
		retired = append(retired, canonical)
	}
	harvested := 0
	if len(retired) > 0 {
		harvested = handdelivery.StampHandDeliveredRows(vaultPath, retired, log)
		for _, canonical := range retired {
			SweepReviewLeftovers(vaultPath, canonical, "", log)
		}
	}
	return harvested
}

// dedupeKeepOrder mirrors list(dict.fromkeys(urls)).
func dedupeKeepOrder(urls []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, u := range urls {
		if !seen[u] {
			seen[u] = true
			out = append(out, u)
		}
	}
	return out
}

// ptrStr safely dereferences a *string (NULL → "").
func ptrStr(p *string) string {
	if p == nil {
		return ""
	}
	return *p
}
