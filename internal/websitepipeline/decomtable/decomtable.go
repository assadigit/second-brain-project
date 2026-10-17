// Package decomtable is the graveyard table's own grammar floor — the
// parse and the status predicates of “_review/DECOMMISSIONED.md“
// (website_pipeline.py's table layer, v0.44.0…v0.60.2).
//
// WHY A SUBPACKAGE (the layering record): in the reference,
// website_pipeline.py imports hand_delivery.py at module level while
// hand_delivery.py imports website_pipeline.py's table helpers
// (decommission_table_path, _parse_decommission_rows, the status
// predicates) INSIDE its stamp functions — Python's function-level
// import dissolves the cycle. Go has no deferred imports, so the table
// grammar the two packages share lives HERE, below both:
//
//	decomtable  ←  websitepipeline
//	decomtable  ←  handdelivery
//	links       ←  decomtable
//
// T-40 (SESSION-015) carved this floor; the table WRITERS
// (write_decommission_candidates, mark_urls_dead_in_table,
// stamp_stored_rows, prune, refresh, consume) stay in the parent
// websitepipeline package, where the reference keeps them.
package decomtable

import (
	"regexp"
	"strings"

	"github.com/assadigit/second-brain-project/internal/links"
)

// ReviewFolder mirrors REVIEW_FOLDER (website_pipeline.py:81). The
// literal lives here (not in the parent) so handdelivery's
// HandDeliveryDir can use it without an import cycle — the same
// deferred-import boundary the reference draws.
const ReviewFolder = "_review"

// DecommissionTable mirrors DECOMMISSION_TABLE (v0.44.0): one markdown
// table per vault at <vault>/_review/DECOMMISSIONED.md — the graveyard.
const DecommissionTable = "DECOMMISSIONED.md"

// TableHeaderCells mirrors _TABLE_HEADER_CELLS (v0.52.0): the canonical
// column set every stamping site writes into (parts index 6 = the
// Status cell; GFM drops cells beyond the header count, so a table
// whose header lost columns would hide the stamps).
const TableHeaderCells = 7

// CanonicalHeader / CanonicalSeparator mirror _CANONICAL_HEADER /
// _CANONICAL_SEPARATOR — the seven-column master-table grammar.
const (
	CanonicalHeader    = "| # | Date | URL | Domain | Source | Status | Notes |"
	CanonicalSeparator = "|---|------|-----|--------|--------|--------|-------|"
)

// DeadMarkers mirrors DEAD_MARKERS (v0.44.0): what reads as DEAD in a
// Status cell — the burial emojis plus the plain words (substring +
// case-insensitive). ✅ / unreviewed / blank never do.
var DeadMarkers = []string{"🪦", "❌", "☠️", "💀", "decommissioned", "retired", "dead"}

// ReviveMarkers mirrors REVIVE_MARKERS: what reads as REVIVED — the
// graveyard's other door (a link buried by mistake is fetched like new
// again).
var ReviveMarkers = []string{"♻️", "revived", "restored", "un-decommissioned"}

// ReviewedMarkers mirrors REVIEWED_MARKERS (v0.47.0): what reads as
// REVIEWED-AND-KEPT — the tick family. 'unreviewed' (the pre-filled
// default) NEVER counts via the word 'reviewed' (a substring of the
// default); a revived Status never counts either.
var ReviewedMarkers = []string{"✅", "✔", "☑", "reviewed", "kept", "done"}

// ReviewedTickMarkers is the strict-tick subset REVIEWED uses BEFORE
// the 'unreviewed' exclusion (the v0.52.0 icon-column law: an ✅ in the
// # column over an 'unreviewed' Status cell is a REVIEWED row).
var ReviewedTickMarkers = []string{"✅", "✔", "☑", "kept", "done"}

// BanishEmoji mirrors BANISH_EMOJI (v0.58.0 — THE BANISHMENT).
const BanishEmoji = "🗑️"

// BanishWords mirrors BANISH_WORDS: the plain words that read as
// delete-me (auto_delete joined v0.59.0 — the word Obsidian
// autocompletes the owner's typed "delete" to; the hyphen twin rides
// along).
var BanishWords = []string{"delete", "banish", "blacklist", "purge", "auto_delete", "auto-delete"}

// BanishKeys mirrors BANISH_KEYS: the boolean frontmatter keys that
// carry the same verdict.
var BanishKeys = []string{"decommission", "banish", "blacklist"}

// Row mirrors one _parse_decommission_rows row:
// {'url', 'status', 'icon', 'notes', 'line', 'raw'}.
type Row struct {
	URL    string
	Status string // the COMBINED gesture text (icon + Status cell)
	Icon   string // the # cell
	Notes  string
	Line   int // the row's 0-based line index in the file
	Raw    string
}

// DecommissionTablePath mirrors decommission_table_path: the graveyard
// file for one vault.
func DecommissionTablePath(vaultPath string) string {
	if vaultPath == "" {
		return ReviewFolder + "/" + DecommissionTable
	}
	return strings.TrimSuffix(vaultPath, "/") + "/" + ReviewFolder + "/" + DecommissionTable
}

// IsSeparatorLine mirrors _is_separator_line: a markdown table's
// |---|---| row (dashes, colons, spaces, pipes — nothing else).
func IsSeparatorLine(line string) bool {
	s := strings.TrimSpace(line)
	if !strings.HasPrefix(s, "|") {
		return false
	}
	for _, c := range s {
		switch c {
		case '-', ':', '|', ' ', '\t':
		default:
			return false
		}
	}
	return true
}

// startsHTTP reports whether s starts with http:// or https://
// (Python's str.lower().startswith(('http://', 'https://'))).
func startsHTTP(s string) bool {
	l := strings.ToLower(s)
	return strings.HasPrefix(l, "http://") || strings.HasPrefix(l, "https://")
}

// ParseDecommissionRows mirrors _parse_decommission_rows (v0.44.0 +
// v0.52.0's icon-column law): data rows of the graveyard table.
// Malformed rows are skipped, never raised — a hand-edited table must
// never crash a batch. A row needs only FOUR cells to exist; the
// gesture is the COMBINED icon + Status text, so every verdict
// predicate reads the emoji wherever the owner put it. The # column is
// the owner's column now.
//
// Python reads with f.read().splitlines() (the full boundary set); the
// caller passes already-split lines (see readPySplitLines).
func ParseDecommissionRows(lines []string) []Row {
	var rows []Row
	for idx, line := range lines {
		if !(strings.HasPrefix(line, "| ") && strings.Contains(line, "http")) {
			continue
		}
		parts := strings.Split(line, "|")
		if len(parts) < 5 { // | # | Date | URL | Domain | (the 4-col law)
			continue
		}
		url := strings.TrimSpace(parts[3])
		if !startsHTTP(url) {
			// a hand-made shape — the URL may sit in any early
			// cell (Python's parts[1:6] clamps to len(parts))
			url = ""
			probe := parts[1:]
			if len(parts) >= 6 {
				probe = parts[1:6]
			}
			for _, cell := range probe {
				c := strings.TrimSpace(cell)
				if startsHTTP(c) {
					url = c
					break
				}
			}
			if url == "" {
				continue
			}
		}
		icon := strings.TrimSpace(parts[1])
		if icon == "-" || icon == "—" || icon == "" {
			icon = "" // the app's own # cell — no gesture
		}
		statusCell := ""
		if len(parts) >= 8 {
			statusCell = strings.TrimSpace(parts[6])
		}
		gestureParts := make([]string, 0, 2)
		if icon != "" {
			gestureParts = append(gestureParts, icon)
		}
		if statusCell != "" {
			gestureParts = append(gestureParts, statusCell)
		}
		gesture := strings.TrimSpace(strings.Join(gestureParts, " "))
		notes := ""
		if len(parts) > 7 {
			notes = strings.TrimSpace(parts[7])
		}
		rows = append(rows, Row{
			URL: url, Status: gesture, Icon: icon, Notes: notes,
			Line: idx, Raw: line,
		})
	}
	return rows
}

// StatusIsDead mirrors _status_is_dead (substring + case-insensitive).
func StatusIsDead(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	return containsAny(s, DeadMarkers)
}

// StatusIsRevived mirrors _status_is_revived.
func StatusIsRevived(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	return containsAny(s, ReviveMarkers)
}

// StatusIsBanished mirrors _status_is_banished (v0.58.0): does a
// master-table gesture read as BANISHED? Precedence: ♻️ revived wins
// (the undo door of every burial); beyond that the banishment is the
// STRONGEST retirement in the grammar — when a cell says both dead and
// delete, the owner's removal intent wins.
func StatusIsBanished(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	if StatusIsRevived(s) {
		return false // the undo door outranks every burial
	}
	if strings.Contains(s, BanishEmoji) {
		return true
	}
	return containsAny(s, BanishWords)
}

// StatusIsReviewed mirrors _status_is_reviewed (v0.47.0 + v0.52.0):
// True when a Status cell reads as reviewed-and-kept. The tick markers
// are checked BEFORE the 'unreviewed' exclusion (the icon column's law:
// an ✅ over an 'unreviewed' cell is a REVIEWED row); the word
// 'reviewed' alone still needs the exclusion (it is a substring of the
// default).
func StatusIsReviewed(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	if containsAny(s, ReviveMarkers) {
		return false
	}
	if containsAny(s, ReviewedTickMarkers) {
		return true // a tick marker beats the other cell's default
	}
	if strings.Contains(s, "unreviewed") {
		return false
	}
	return strings.Contains(s, "reviewed")
}

// StatusIsTerminal mirrors _status_is_terminal (v0.60.2): a
// master-table row whose story is WRITTEN — every CONFIRMED stamp, the
// harvest's final green verdict, the stored stamp, and the fetcher's
// own auto-verdict rows. A PENDING gesture never matches; a ♻️ revived
// row stays while its revival is in flight.
func StatusIsTerminal(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	if StatusIsRevived(s) {
		return false // the undo door stays visible, in flight
	}
	if strings.Contains(s, "confirmed") {
		return true // every consume/harvest confirm stamp
	}
	if strings.Contains(s, "hand-delivered") {
		return true // the harvest's final green verdict
	}
	if strings.Contains(s, "stored") {
		return true // 📁 stored — the note is in the vault
	}
	if strings.Contains(s, "auto —") {
		return true // the fetcher's own verdict rows
	}
	return false
}

func containsAny(s string, markers []string) bool {
	for _, m := range markers {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}

// TableCell mirrors _table_cell (v0.47.0): one markdown-table-safe
// cell — pipes and newlines become '·' / spaces; anything longer than
// limit chars is truncated with an ellipsis. The whitespace collapse
// is Python's ' '.join(s.split()) (the full Unicode \s class).
func TableCell(text string, limit int) string {
	s := strings.ReplaceAll(text, "|", "·")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = collapseWhitespace(s)
	// Python len() counts codepoints; the truncation is rune-wise.
	runes := []rune(s)
	if len(runes) > limit {
		cut := strings.TrimRight(string(runes[:limit-1]), " \t\n\v\f\r")
		return cut + "…"
	}
	return s
}

var pyWhitespace = regexp.MustCompile(`[\s\v\x1c-\x1f\x85\p{Z}]+`)

// collapseWhitespace mirrors ' '.join(s.split()) — Python's str.split()
// with no argument splits on the Unicode whitespace class (which
// includes \x1c-\x1f and \x85 — Go's \s does not) and drops empties.
func collapseWhitespace(s string) string {
	return strings.TrimSpace(pyWhitespace.ReplaceAllString(s, " "))
}

// StatusKeyOfURL normalizes a table row's URL through the one-spelling
// normalizer (links.NormalizeWebsiteURL) — the shared helper the
// scans and writers use.
func StatusKeyOfURL(url string) string {
	return links.NormalizeWebsiteURL(url)
}
