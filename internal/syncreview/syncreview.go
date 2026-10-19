package syncreview

// syncreview.go — the state model half of core/sync_review.py: the
// truth table (count + item forms), the banner verdict (worst-first,
// v0.91.0's named verdict + the setup/measure split), the micro-copy
// caption law, the relative "added" clock, the oldest-first sort, and
// the chip grammar's disabled-until law. PURE — no I/O anywhere in
// this file.

import (
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
)

// The four cell states, worst-first for precedence (banner_verdict).
const (
	StateBehind   = "behind"
	StateUnknown  = "unknown"
	StateReviewed = "reviewed"
	StateMatch    = "match"
)

// States mirrors STATES: the precedence order itself.
var States = []string{StateBehind, StateUnknown, StateReviewed, StateMatch}

// Reasons mirrors REASONS — the four review reasons (the spec's chips;
// Hick's Law keeps it at four: one tap per reason, the log stays
// useful without typing).
var Reasons = []string{"not_ready", "private", "duplicate", "other"}

// ReasonLabels mirrors REASON_LABELS.
var ReasonLabels = map[string]string{
	"not_ready": "Not ready",
	"private":   "Private",
	"duplicate": "Duplicate",
	"other":     "Other",
}

// ValidReason is valid_reason: True when reason is one of the four
// chips.
func ValidReason(reason string) bool {
	for _, r := range Reasons {
		if r == reason {
			return true
		}
	}
	return false
}

// intPtr / intOf are the *int helpers (nil = the Python None).
func intPtr(n int) *int { return &n }

// CellState is cell_state — THE truth table, count form (the paint
// side may only know counts):
//
//   - not set up (or a side unmeasured) → unknown
//   - gap == 0                          → match
//   - every gap item deferred           → reviewed
//   - any unexplained gap               → behind
func CellState(setUp bool, sourceCount, targetCount *int,
	deferredGapCount int) string {
	if !setUp || sourceCount == nil || targetCount == nil {
		return StateUnknown
	}
	gap := *sourceCount - *targetCount
	if gap < 0 {
		gap = 0
	}
	if gap == 0 {
		return StateMatch
	}
	if deferredGapCount >= gap {
		return StateReviewed
	}
	return StateBehind
}

// SplitGap is split_gap: split a stage's gap into (unexplained,
// deferredGap) — identity by the STABLE-ID law (note path or URL),
// never a position. Order is preserved (the caller sorts oldest-first).
func SplitGap(gapItems, deferredIDs []string) (unexplained, deferredGap []string) {
	deferred := make(map[string]bool, len(deferredIDs))
	for _, id := range deferredIDs {
		deferred[id] = true
	}
	for _, item := range gapItems {
		if deferred[item] {
			deferredGap = append(deferredGap, item)
		} else {
			unexplained = append(unexplained, item)
		}
	}
	return unexplained, deferredGap
}

// Cell is stage_cell's answer — the one object the view paints and the
// dialog diffs. A nil SourceCount/TargetCount means "that side could
// not be measured" (the honest verdict is unknown, never a guess).
// Stream/Stage/Reason/Needs are the identity fuel v0.91.0's named
// verdict reads (the gather fills them; bare cells keep the v0.86.0
// grammar byte-for-byte).
type Cell struct {
	State            string   `json:"state"`
	SourceCount      *int     `json:"source_count"`
	TargetCount      *int     `json:"target_count"`
	GapCount         int      `json:"gap_count"`
	UnexplainedIDs   []string `json:"unexplained_ids"`
	DeferredGapIDs   []string `json:"deferred_gap_ids"`
	UnexplainedCount int      `json:"unexplained_count"`
	DeferredGapCount int      `json:"deferred_gap_count"`

	// The identity (v0.91.0 THE NAMED VERDICT's fuel).
	Stream string `json:"stream,omitempty"`
	Stage  string `json:"stage,omitempty"`
	Reason string `json:"reason,omitempty"`
	Needs  string `json:"needs,omitempty"` // 'setup' | 'measure' | ""
}

// StageCell is stage_cell — THE truth table, item form: the gap in
// first-seen order, de-duplicated (stable IDs, no dupes), minus the
// target set; the deferred split over that gap.
func StageCell(setUp bool, sourceItems, targetItems []string,
	deferredIDs []string) *Cell {
	if !setUp || sourceItems == nil || targetItems == nil {
		return &Cell{State: StateUnknown, GapCount: 0,
			UnexplainedIDs: []string{}, DeferredGapIDs: []string{}}
	}
	sourceSet := map[string]bool{}
	for _, id := range sourceItems {
		sourceSet[id] = true
	}
	targetSet := map[string]bool{}
	for _, id := range targetItems {
		targetSet[id] = true
	}
	var gap []string
	seen := map[string]bool{}
	for _, id := range sourceItems {
		if seen[id] || targetSet[id] {
			continue
		}
		seen[id] = true // dict.fromkeys: first-seen order, deduped
		gap = append(gap, id)
	}
	unexplained, deferredGap := SplitGap(gap, deferredIDs)
	srcN, tgtN := len(sourceSet), len(targetSet)
	return &Cell{
		State:            CellState(true, intPtr(srcN), intPtr(tgtN), len(deferredGap)),
		SourceCount:      intPtr(srcN),
		TargetCount:      intPtr(tgtN),
		GapCount:         len(gap),
		UnexplainedIDs:   unexplained,
		DeferredGapIDs:   deferredGap,
		UnexplainedCount: len(unexplained),
		DeferredGapCount: len(deferredGap),
	}
}

// Verdict is banner_verdict's answer: the tone (warning | muted |
// success) + the text.
type Verdict struct {
	Tone string `json:"tone"`
	Text string `json:"text"`
}

func plural(n int, one, many string) string {
	if n != 1 {
		return many
	}
	return one
}

// BannerVerdict is banner_verdict — the banner's verdict over every
// cell, WORST-FIRST (the spec: "Conflicts → show the worst"):
//
//   - any Behind    → warning, "N items waiting" (N = unexplained
//     total); named cells append every problem stage and its count
//   - else any Unknown → muted, "N stages not set up"; named cells
//     split "not set up" (needs='setup' — a Settings door) vs
//     "not measured" (needs='measure' — a check that has not run)
//   - else healthy  → "In sync · N deferred" when N > 0 (the calm
//     count never disappears), plain "In sync" otherwise
//
// Bare cells (no identity) keep the v0.86.0 grammar byte-for-byte.
func BannerVerdict(cells []*Cell) Verdict {
	var behind []*Cell
	for _, c := range cells {
		if c != nil && c.State == StateBehind {
			behind = append(behind, c)
		}
	}
	if len(behind) > 0 {
		total := 0
		var named []*Cell
		for _, c := range behind {
			total += c.UnexplainedCount
			if c.Stream != "" && c.Stage != "" {
				named = append(named, c)
			}
		}
		if len(named) > 0 {
			parts := make([]string, 0, len(named))
			for _, c := range named {
				parts = append(parts, c.Stream+" · "+c.Stage+": "+
					strconv.Itoa(c.UnexplainedCount)+" waiting")
			}
			return Verdict{"warning",
				strconv.Itoa(total) + " item" + plural(total, "", "s") +
					" waiting — " + strings.Join(parts, "; ")}
		}
		return Verdict{"warning",
			strconv.Itoa(total) + " item" + plural(total, "", "s") + " waiting"}
	}
	var unknown []*Cell
	for _, c := range cells {
		if c != nil && c.State == StateUnknown {
			unknown = append(unknown, c)
		}
	}
	if len(unknown) > 0 {
		var named []*Cell
		for _, c := range unknown {
			if c.Stream != "" && c.Stage != "" {
				named = append(named, c)
			}
		}
		if len(named) == 0 {
			n := len(unknown)
			return Verdict{"muted",
				strconv.Itoa(n) + " stage" + plural(n, "", "s") + " not set up"}
		}
		// THE SPLIT: "not set up" (config missing — a Settings door)
		// vs "not measured" (a check that has not run this session).
		var setup, measure []*Cell
		for _, c := range named {
			needs := c.Needs
			if needs == "" {
				needs = "setup"
			}
			if needs == "setup" {
				setup = append(setup, c)
			} else if needs == "measure" {
				measure = append(measure, c)
			}
		}
		var parts []string
		if len(setup) > 0 {
			who := make([]string, 0, len(setup))
			for _, c := range setup {
				reason := c.Reason
				if reason == "" {
					reason = "not set up"
				}
				who = append(who, c.Stream+" · "+c.Stage+" — "+reason)
			}
			parts = append(parts, strconv.Itoa(len(setup))+" stage"+
				plural(len(setup), "", "s")+" not set up: "+
				strings.Join(who, "; "))
		}
		if len(measure) > 0 {
			who := make([]string, 0, len(measure))
			for _, c := range measure {
				reason := c.Reason
				if reason == "" {
					reason = "not measured yet"
				}
				who = append(who, c.Stream+" · "+c.Stage+" — "+reason)
			}
			parts = append(parts, strconv.Itoa(len(measure))+" stage"+
				plural(len(measure), "", "s")+" not measured: "+
				strings.Join(who, "; ")+" (click the cell to run its check)")
		}
		return Verdict{"muted", strings.Join(parts, " · ")}
	}
	deferred := 0
	for _, c := range cells {
		if c != nil {
			deferred += c.DeferredGapCount
		}
	}
	if deferred > 0 {
		return Verdict{"success", "In sync · " + strconv.Itoa(deferred) + " deferred"}
	}
	return Verdict{"success", "In sync"}
}

// GapItem is one stage's item row (the differences list's shape): the
// STABLE ID (the note path or the URL, never a position), the display
// name, the added-at stamp, and the note-path door. AddedDaysAgo is
// the sort key (nil = undated).
type GapItem struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	URL          string `json:"url,omitempty"`
	AddedAt      string `json:"added_at"`
	NotePath     string `json:"note_path"`
	AddedDaysAgo *int   `json:"added_days_ago"`
}

// StageRecord is the gather's stage shape — the cell's verdict plus
// the identity + the items; the caption, the matrix paint and the
// differences modal all read this one record.
type StageRecord struct {
	Key       string `json:"key"`    // 'repos:queue_to_vault' | …
	Stream    string `json:"stream"` // Repos | Websites | Personal
	Stage     string `json:"stage"`  // Queue → Vault | Vault → Website | Vault → Backup
	State     string `json:"state"`
	SetUp     bool   `json:"set_up"`
	SetUpHint string `json:"set_up_hint"`
	Reason    string `json:"reason"`
	Needs     string `json:"needs"` // 'setup' | 'measure' | ""

	SourceLabel string `json:"source_label"`
	TargetLabel string `json:"target_label"`
	FixLabel    string `json:"fix_label"`
	FixAction   string `json:"fix_action"`

	DisplayNumber *int `json:"display_number"`

	SourceCount      *int     `json:"source_count"`
	TargetCount      *int     `json:"target_count"`
	GapCount         int      `json:"gap_count"`
	UnexplainedCount int      `json:"unexplained_count"`
	DeferredGapCount int      `json:"deferred_gap_count"`
	UnexplainedIDs   []string `json:"unexplained_ids"`
	DeferredGapIDs   []string `json:"deferred_gap_ids"`

	Items []GapItem `json:"items"`
}

// Caption is cell_caption's answer: the BIG NUMBER the eye compares
// plus the sentence that names both sides.
type Caption struct {
	Num string `json:"num"`
	Cap string `json:"cap"`
}

// CellCaption is cell_caption — THE MICRO-COPY LAW (v0.91.0): every
// matrix cell's number is NAMED, because a bare number teaches
// nothing. Branches on the stage's own key suffix; unknown writes its
// own reason on itself; junk degrades to dashes, never a guess, never
// a raise.
func CellCaption(stage *StageRecord) Caption {
	if stage == nil {
		stage = &StageRecord{}
	}
	state := stage.State
	if state == "" {
		state = StateUnknown
	}
	gap := stage.GapCount
	deferred := stage.DeferredGapCount
	if state == StateUnknown {
		reason := strings.TrimSpace(stage.Reason)
		if reason == "" {
			reason = "not set up or not measured"
		}
		return Caption{"—", reason}
	}
	num := "—"
	inVault, sealed := 0, 0
	if stage.DisplayNumber != nil {
		num = strconv.Itoa(*stage.DisplayNumber)
		inVault = *stage.DisplayNumber
		sealed = *stage.DisplayNumber
	}
	var v, p string
	if stage.SourceCount != nil {
		v = strconv.Itoa(*stage.SourceCount)
	} else {
		v = "—"
	}
	if stage.TargetCount != nil {
		p = strconv.Itoa(*stage.TargetCount)
	} else {
		p = "—"
	}
	var cap string
	switch {
	case strings.HasSuffix(stage.Key, ":queue_to_vault"):
		switch state {
		case StateBehind:
			cap = strconv.Itoa(gap) + " waiting in the queue · " +
				strconv.Itoa(inVault) + " already in the vault"
		case StateReviewed:
			cap = strconv.Itoa(deferred) + " deferred (explained) · " +
				strconv.Itoa(inVault) + " in the vault"
		default:
			cap = strconv.Itoa(inVault) + " in the vault · " +
				strconv.Itoa(gap) + " waiting in the queue"
		}
	case strings.HasSuffix(stage.Key, ":vault_to_site"):
		switch state {
		case StateBehind:
			cap = strconv.Itoa(gap) + " to publish · " + v +
				" in the vault · " + p + " on the site"
		case StateReviewed:
			cap = strconv.Itoa(deferred) + " deferred (explained) · " +
				v + " in the vault · " + p + " on the site"
		default:
			cap = v + " in the vault = " + p + " on the site"
		}
	default: // *:vault_to_backup
		total := sealed + gap
		switch state {
		case StateBehind:
			cap = strconv.Itoa(gap) + " unsealed · " +
				strconv.Itoa(total) + " in the vault · " +
				strconv.Itoa(sealed) + " sealed in the mirror"
		case StateReviewed:
			cap = strconv.Itoa(deferred) + " deferred (explained) · " +
				strconv.Itoa(total) + " in the vault · " +
				strconv.Itoa(sealed) + " sealed in the mirror"
		default:
			cap = strconv.Itoa(total) + " in the vault = " +
				strconv.Itoa(sealed) + " in the mirror"
		}
	}
	return Caption{num, cap}
}

// ---------------------------------------------------------------------------
// The relative "added" time — "3 d ago", locale-friendly
// ---------------------------------------------------------------------------

// dtRe mirrors _DT_RE: YYYY-MM-DD, optional HH:MM[:SS], 'T' or space.
var dtRe = regexp.MustCompile(
	`^(\d{4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2})(?::(\d{2}))?)?`)

// parseAdded mirrors _parse_added: None when unparseable (a bad date
// like 2026-13-45 is a ValueError → None, never a crash).
func parseAdded(addedAt string) (time.Time, bool) {
	raw := strings.TrimSpace(addedAt)
	if raw == "" {
		return time.Time{}, false
	}
	m := dtRe.FindStringSubmatch(raw)
	if m == nil {
		return time.Time{}, false
	}
	atoi := func(s string, def int) int {
		if s == "" {
			return def
		}
		n, err := strconv.Atoi(s)
		if err != nil {
			return def
		}
		return n
	}
	y, mo, d := atoi(m[1], 0), atoi(m[2], 0), atoi(m[3], 0)
	hh, mm, ss := atoi(m[4], 0), atoi(m[5], 0), atoi(m[6], 0)
	then := time.Date(y, time.Month(mo), d, hh, mm, ss, 0, time.UTC)
	// Python's datetime(y, mo, d, …) raises on an invalid calendar
	// date; time.Date normalizes instead — reject what differs.
	if then.Year() != y || int(then.Month()) != mo || then.Day() != d ||
		then.Hour() != hh || then.Minute() != mm || then.Second() != ss {
		return time.Time{}, false
	}
	return then, true
}

func utcNowOr(now time.Time) time.Time {
	if now.IsZero() {
		return time.Now().UTC()
	}
	if now.Location() == nil {
		return now.UTC() // a naive now is tolerated (assumed UTC)
	}
	return now
}

// FormatAddedAgo is format_added_ago: "just now" · "5 m ago" ·
// "3 h ago" · "3 d ago" · "2 mo ago" · "1 y ago". Unknown/missing
// dates answer "—" (honest silence — the queue items carry no date).
func FormatAddedAgo(addedAt string, now time.Time) string {
	then, ok := parseAdded(addedAt)
	if !ok {
		return "—"
	}
	now = utcNowOr(now)
	seconds := now.Sub(then).Seconds()
	if seconds < 0 {
		seconds = 0 // a clock skew never says "in the future"
	}
	if seconds < 60 {
		return "just now"
	}
	minutes := int(seconds / 60)
	if minutes < 60 {
		return strconv.Itoa(minutes) + " m ago"
	}
	hours := minutes / 60
	if hours < 24 {
		return strconv.Itoa(hours) + " h ago"
	}
	days := hours / 24
	if days < 30 {
		return strconv.Itoa(days) + " d ago"
	}
	months := days / 30
	if months < 12 {
		return strconv.Itoa(months) + " mo ago"
	}
	return strconv.Itoa(months/12) + " y ago"
}

// AddedDaysAgo is added_days_ago: the numeric age in days (nil when
// unknown) — the OLDEST-FIRST sort key for the Differences list.
func AddedDaysAgo(addedAt string, now time.Time) *int {
	then, ok := parseAdded(addedAt)
	if !ok {
		return nil
	}
	now = utcNowOr(now)
	days := int(now.Sub(then).Seconds() / 86400)
	if days < 0 {
		days = 0
	}
	return &days
}

// SortOldestFirst is sort_oldest_first: by days descending (oldest
// first), items without a date after them, stable within equal keys.
func SortOldestFirst(items []GapItem) []GapItem {
	out := append([]GapItem(nil), items...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].AddedDaysAgo, out[j].AddedDaysAgo
		switch {
		case a != nil && b != nil:
			return *a > *b // (-days): oldest first
		case a != nil:
			return true // dated before undated
		default:
			return false
		}
	})
	return out
}

// MarkAllowed is mark_allowed — THE disabled-until law: "Mark as
// intentional" stays disabled until ≥1 item is checked AND a chip is
// chosen. (An empty reason is NO choice; only a valid chip arms the
// button.)
func MarkAllowed(checkedCount int, reason string) bool {
	return checkedCount >= 1 && ValidReason(reason)
}
