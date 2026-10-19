// scandialogs.go — T-76: the Scan flow's three UI legs' pure engines
// (spec-gui-dialogs — THE AUTHORITY). The scan job itself (the engine
// under all three) already lives in scanjob.go (T-71's _vault_scan_job
// port); this file ports the dialog models:
//
//   - PickDisplayModelOf — vault_pick_dialog.py's pick_display_model
//     (detect_scan_vaults's walk is already DetectScanVaults at T-75):
//     the three rows' display shape, the status lines, the hint when no
//     vault can be scanned, and the Scan button's live arithmetic;
//   - PlanDisplayModelOf — scan_plan_dialog.py's plan_display_model:
//     the whole plan as display rows (the four sections in dialog
//     order, every row carrying its stable id + its one-line grammar +
//     its full tooltip story, the MAX_ROWS freeze guard, the inventory
//     block, the landable move's "(as …)" name);
//   - PlanPickState — the picking door's selection semantics as a pure
//     state machine: checked-by-default rows, low-confidence
//     PRE-UNCHECKED, the master sweep (one gesture, one recount), the
//     Apply label's live count, the arm-guard's enable law, and the
//     four-key selection only a confirmed answer ever rides;
//   - ResultDisplayModelOf — scan_result_dialog.py's
//     result_display_model: the LANDED counts (never the plan's
//     promises), the honest 'K of N' warning tones, the failed moves
//     BY NAME, the per-vault sections, the nothing-line + the
//     recoverability footnote.
//
// The renderers live on the shell (frontend/settings.js over the
// wails bindings); every string here is the reference's own, verbatim.
package app

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaulthygiene"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
)

// ---------------------------------------------------------------------------
// The shared laws (the spec's own numbers — the clocks are the law both
// sides read; the frontend's timers tick them).
// ---------------------------------------------------------------------------

// PlanMaxRows is the freeze guard (MAX_ROWS = 500): a list past this
// size is counted instead of rendered — a never-freeze law, far beyond
// any real plan.
const PlanMaxRows = 500

// PlanArmMs is THE FRESH-CLICK DOOR's arming beat (350 ms): the
// frontend swallows every input aimed at the plan modal for this long
// after open, and both verdict buttons are born disabled — a click from
// before the window was live is not consent, it is debris.
const PlanArmMs = 350

// ResultAutoCloseS is the scorecard's self-close window (BRIEF is the
// owner's own word): 12 s, the ticking line says so, OK and Esc
// dismiss sooner, nothing is ever blocked on the modal.
const ResultAutoCloseS = 12

// The picking door's section keys (the selection's own vocabulary —
// the plan's four proposal lists, in dialog order).
const (
	ScanSectionDeletions      = "deletions"
	ScanSectionConsolidations = "consolidations"
	ScanSectionNewFolders     = "new_folders"
	ScanSectionMoves          = "moves"
)

// scanSectionKeys is the dialog's own section order.
var scanSectionKeys = []string{
	ScanSectionDeletions, ScanSectionConsolidations,
	ScanSectionNewFolders, ScanSectionMoves,
}

// The scorecard's row tones.
const (
	ToneSuccess = "success"
	ToneWarning = "warning"
)

// ScanConfirmTimeoutS reads scan_confirm_timeout_s (the reference's
// own lenient reader: junk degrades to the 300 s default). The shell's
// ask payload carries it — the plan modal's countdown ticks it.
func ScanConfirmTimeoutS(cfg *storage.OrderedMap) float64 {
	return scanConfirmTimeoutSOf(cfg)
}

// ---------------------------------------------------------------------------
// THE VAULT PICKER — pick_display_model (vault_pick_dialog.py:199-251).
// ---------------------------------------------------------------------------

// PickRow is one vault's display row (the picker's own shape).
type PickRow struct {
	Kind       string `json:"kind"`
	Label      string `json:"label"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	StatusText string `json:"status_text"`
	Scannable  bool   `json:"scannable"`
	Selectable bool   `json:"selectable"`
	Why        string `json:"why"`
}

// PickDisplayModel is the picker's display shape (the dialog only
// renders it): nothing pre-checked (the choice is the owner's, every
// click), the hint speaks when no vault can be scanned at all.
type PickDisplayModel struct {
	Title           string    `json:"title"`
	Header          string    `json:"header"`
	Rows            []PickRow `json:"rows"`
	SelectableKinds []string  `json:"selectable_kinds"`
	AnySelectable   bool      `json:"any_selectable"`
	Hint            string    `json:"hint"`
}

// pickStatusText is _status_text: the one line under each vault's name
// — the connection truth in the owner's own voice.
func pickStatusText(row SyncVaultRow) string {
	switch row.Status {
	case VaultStatusConnected:
		if row.NoteCount != nil && *row.NoteCount >= 0 {
			return fmt.Sprintf("%d note(s)", *row.NoteCount)
		}
		return "Connected"
	case VaultStatusNotFound:
		return "Not found at this path"
	default:
		return "Not set — Settings → 📁 Vault"
	}
}

// PickDisplayModelOf is pick_display_model — PURE (the dialog only
// renders it). The rows arrive from DetectScanVaults (the pure walk
// already ported at T-75); junk rows degrade honestly (an unknown kind
// renders as an untickable row, never a crash).
func PickDisplayModelOf(rows []SyncVaultRow) *PickDisplayModel {
	out := &PickDisplayModel{
		Title: "Choose the vault to scan",
		Header: "Which library should this scan pass read? Tick the " +
			"vault(s) to scan — only the ticked library's plan " +
			"opens for your review; the other stays untouched.",
		Rows:            []PickRow{},
		SelectableKinds: []string{},
	}
	for _, row := range rows {
		out.Rows = append(out.Rows, PickRow{
			Kind:       row.Kind,
			Label:      strOr(row.Label, "Vault"),
			Icon:       strOr(row.Icon, "🗂️"),
			Path:       row.Path,
			Status:     strOr(row.Status, VaultStatusNotSet),
			StatusText: pickStatusText(row),
			Scannable:  row.Scannable,
			Selectable: row.Selectable,
			Why:        row.Why,
		})
	}
	for _, row := range out.Rows {
		if row.Selectable {
			out.SelectableKinds = append(out.SelectableKinds, row.Kind)
		}
	}
	out.AnySelectable = len(out.SelectableKinds) > 0
	if !out.AnySelectable {
		out.Hint = "No scan-able vault is connected yet — set the Websites " +
			"vault or the GitHub Repos vault first (Settings → 📁 " +
			"Vault), then Scan again."
	}
	return out
}

// PickScanButtonLabel is the Scan button's live arithmetic (the
// picker's own law): disabled-shaped with nothing ticked, the singular
// with one, the both-vaults word with two. The picked kinds' ORDER is
// the display order (websites before repos), whatever the caller's
// box order.
func PickScanButtonLabel(picked []string) string {
	n := 0
	for _, kind := range []string{VaultKindWebsites, VaultKindRepos} {
		for _, p := range picked {
			if strings.EqualFold(strings.TrimSpace(p), kind) {
				n++
			}
		}
	}
	switch n {
	case 0:
		return "🔍 Scan"
	case 1:
		return "🔍 Scan this vault"
	default:
		return "🔍 Scan both vaults"
	}
}

// PickedScanKinds normalizes the picker's answer into the job's scope
// vocabulary (the display order; junk falls away — the scope law
// scanjob.go itself re-checks).
func PickedScanKinds(picked []string) []string {
	out := []string{}
	for _, kind := range []string{VaultKindWebsites, VaultKindRepos} {
		for _, p := range picked {
			if strings.EqualFold(strings.TrimSpace(p), kind) {
				out = append(out, kind)
				break
			}
		}
	}
	if len(out) == 0 {
		return nil // the safe cancel: nothing scanned
	}
	return out
}

// ---------------------------------------------------------------------------
// THE PLAN MODAL — plan_display_model (scan_plan_dialog.py:160-268).
// ---------------------------------------------------------------------------

// PlanRow is one proposal row: the stable id (the selection's own key
// — the canonical URL / the note path / the folder path), the one-line
// grammar (the renderer elides middle), and the full tooltip story.
type PlanRow struct {
	ID            string `json:"id"`
	Text          string `json:"text"`
	Full          string `json:"full"`
	LowConfidence bool   `json:"low_confidence"`
}

// PlanSection is one of the four section cards: its wash role, its
// tone, its heading (with the live total), its rows (capped by the
// freeze guard), and the counted-not-rendered tail.
type PlanSection struct {
	Key     string    `json:"key"`
	Card    string    `json:"card"`
	Tone    string    `json:"tone"`
	Heading string    `json:"heading"`
	Rows    []PlanRow `json:"rows"`
	Total   int       `json:"total"`
	More    int       `json:"more"`
}

// PlanInventory is the library line's own numbers (the inventory
// block the story opens with).
type PlanInventory struct {
	TotalNotes         int `json:"total_notes"`
	FolderCount        int `json:"folder_count"`
	RootNotes          int `json:"root_notes"`
	UncategorizedNotes int `json:"uncategorized_notes"`
	TrashNotes         int `json:"trash_notes"`
	LockedNotes        int `json:"locked_notes"`
}

// PlanDisplayModel is the plan modal's whole display shape (the
// renderer only paints it; the freeze guard has already bitten).
type PlanDisplayModel struct {
	Title             string        `json:"title"`
	Header            string        `json:"header"`
	LibraryLine       string        `json:"library_line"`
	TrashLine         string        `json:"trash_line"`
	LockedLine        string        `json:"locked_line"`
	KeptLine          string        `json:"kept_line"`
	LowConfidenceLine string        `json:"low_confidence_line"`
	Sections          []PlanSection `json:"sections"`
	Summary           string        `json:"summary"`
	VaultLabel        string        `json:"vault_label"`
	Inventory         PlanInventory `json:"inventory"`
	HasDeletions      bool          `json:"has_deletions"`
	HasFiling         bool          `json:"has_filing"`
	HasConsolidations bool          `json:"has_consolidations"`
	HasLowConfidence  bool          `json:"has_low_confidence"`
}

// planDeletionRow shapes one deletion (the _deletion_line grammar):
// the door's own words ride both the line and the full story.
func planDeletionRow(id, title, detail, marker, door string) PlanRow {
	if door == "" {
		door = "note tag"
	}
	via := door
	switch door {
	case "note tag":
		via = "your mark"
	case "trash folder":
		via = "you moved it to Trash"
	}
	return PlanRow{
		ID:   id,
		Text: fmt.Sprintf("🗑️ %s — via %s", title, via),
		Full: fmt.Sprintf("%s\n%s\nmarker: %s · door: %s",
			title, detail, marker, door),
	}
}

// planFolderRow shapes one new folder (the _folder_line grammar).
func planFolderRow(folder string) PlanRow {
	return PlanRow{
		ID:   folder,
		Text: "📁 " + folder,
		Full: "New folder: " + folder,
	}
}

// planConsolidationRow shapes one merged note (the _consolidation_line
// grammar — ONE NOTE PER SITE: each merge note is its own row).
func planConsolidationRow(id, site, keep, note, source string) PlanRow {
	return PlanRow{
		ID: id,
		Text: fmt.Sprintf("🧲 %s → merges into %s (site: %s)",
			note, keep, site),
		Full: fmt.Sprintf("%s\nmerges into %s\nsite: %s\nsource: %s\n"+
			"(the duplicate rests in .trash/consolidated — "+
			"recoverable by hand; the site is wanted, nothing is "+
			"blacklisted)", note, keep, site, source),
	}
}

// planMoveRow shapes one move (the _move_line grammar — THE LANDABLE
// MOVE: a bent landing name is SPOKEN in the row itself).
func planMoveRow(id, note, from, to, as, reason string, low bool) PlanRow {
	if from == "" {
		from = "(vault root)"
	}
	text := fmt.Sprintf("🚚 %s: %s → %s", note, from, to)
	asName := strings.TrimSpace(as)
	if asName != "" && asName != note {
		text += fmt.Sprintf(" (as %s)", asName)
	}
	full := text
	if reason != "" {
		full += "\nwhy: " + reason
	}
	return PlanRow{ID: id, Text: text, Full: full, LowConfidence: low}
}

// clipRunes clips a string to at most n runes (the model's own
// truncation laws: marker[:80], reason[:120], summary[:400]).
func clipRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

// PlanDisplayModelOf is plan_display_model — PURE (no files, never
// raises): the vaultscan plan in, the modal's display rows out. The
// whole list rides the model (the scrollable-panel law); only a list
// past PlanMaxRows is counted instead of rendered.
func PlanDisplayModelOf(plan *vaultscan.Plan) *PlanDisplayModel {
	if plan == nil {
		plan = &vaultscan.Plan{}
	}
	label := plan.VaultLabel
	if label == "" {
		label = "Websites vault"
		if plan.VaultKind == vaultscan.VaultKindRepos {
			label = "GitHub Repos vault"
		}
	}
	inv := plan.Inventory

	out := &PlanDisplayModel{
		Title:  "Vault scan review — " + label,
		Header: fmt.Sprintf("🗂️ Vault scan review — %s. The whole plan, before anything is touched. Untick anything you deny — only the ticked rows ride; nothing is deleted or moved until you answer.", label),
		LibraryLine: fmt.Sprintf("🏛️ The library (%s): %d note(s) in %d folder(s) — %d orphaned, %d uncategorized",
			label, inv.TotalNotes, inv.FolderCount, inv.RootNotes,
			inv.UncategorizedNotes),
		Summary:    clipRunes(plan.Summary, 400),
		VaultLabel: label,
		Inventory: PlanInventory{
			TotalNotes:         inv.TotalNotes,
			FolderCount:        inv.FolderCount,
			RootNotes:          inv.RootNotes,
			UncategorizedNotes: inv.UncategorizedNotes,
			TrashNotes:         inv.TrashNotes,
			LockedNotes:        inv.LockedNotes,
		},
	}
	if inv.TrashNotes > 0 {
		out.TrashLine = fmt.Sprintf("🗑️ %d note(s) rest in the Trash folder — they ride the deletions below",
			inv.TrashNotes)
	}
	if inv.LockedNotes > 0 {
		out.LockedLine = fmt.Sprintf("🔒 %d locked note(s) are out of the radar — never proposed (locked: true / category_locked: true / the lock tag in the frontmatter; remove the lock by hand to let a note ride again)",
			inv.LockedNotes)
	}
	kept := plan.KeptHandwritten
	if plan.VaultKind == vaultscan.VaultKindRepos {
		kept = plan.KeptHandwrittenR
	}
	if kept > 0 {
		out.KeptLine = fmt.Sprintf("✍️ %d hand-written note(s) carry marks — kept (yours; the app never deletes what it did not write)",
			kept)
	}

	// the four sections, in the dialog's own order. The deletions
	// list rides the vault's OWN kind (the repos vault rides
	// DeletionsRepos — the repobanish gate's items; scanjob.go's
	// own twin law).
	del := []PlanRow{}
	if plan.VaultKind == vaultscan.VaultKindRepos || len(plan.DeletionsRepos) > 0 {
		for _, d := range plan.DeletionsRepos {
			del = append(del, planDeletionRow(
				strOr(d.Canonical, d.URL), strOr(d.Title, d.URL),
				d.URL, clipRunes(d.Marker, 80), d.Door))
		}
	} else {
		for _, d := range plan.Deletions {
			del = append(del, planDeletionRow(
				strOr(d.Canonical, d.URL), strOr(d.Title, d.URL),
				d.URL, clipRunes(d.Marker, 80), d.Door))
		}
	}
	cons := []PlanRow{}
	consTotal := 0
	for _, c := range plan.Consolidations {
		consTotal += len(c.Merge)
		for _, m := range c.Merge {
			cons = append(cons, planConsolidationRow(
				m.Path, c.Site, c.Keep.File, m.File, m.Source))
		}
	}
	folders := []string{}
	for _, f := range plan.NewFolders {
		folders = append(folders, f)
	}
	moves := []PlanRow{}
	for _, m := range plan.Moves {
		moves = append(moves, planMoveRow(
			strOr(m.Path, m.Note), m.Note, m.From, m.To, m.As,
			clipRunes(m.Reason, 120), m.LowConfidence))
	}

	addSection := func(key, card, tone, heading string, rows []PlanRow, total int) {
		shown := rows
		if len(shown) > PlanMaxRows {
			shown = shown[:PlanMaxRows]
		}
		more := total - len(shown)
		if more < 0 {
			more = 0
		}
		head := fmt.Sprintf("%s — %d", heading, total)
		out.Sections = append(out.Sections, PlanSection{
			Key: key, Card: card, Tone: tone, Heading: head,
			Rows: shown, Total: total, More: more,
		})
	}
	addSection(ScanSectionDeletions, "plan_delete_card", "danger",
		"🗑️ Marked for deletion", del, len(del))
	addSection(ScanSectionConsolidations, "plan_consolidate_card", "magnet",
		"🧲 Duplicate notes to merge (one note per site)", cons, consTotal)
	addSection(ScanSectionNewFolders, "plan_create_card", "grow",
		"🌱 New folders to create",
		mapRows(folders, planFolderRow), len(folders))
	addSection(ScanSectionMoves, "plan_move_card", "neutral",
		"🚚 Notes to re-file", moves, len(moves))

	// the has-flags (the reference's own arithmetic over the
	// TOTALS, not the capped rows).
	out.HasDeletions = len(del) > 0
	out.HasConsolidations = consTotal > 0
	out.HasFiling = len(moves)+len(folders) > 0
	for _, m := range moves {
		if m.LowConfidence {
			out.HasLowConfidence = true
			break
		}
	}
	if out.HasLowConfidence {
		out.LowConfidenceLine = "🤔 Unticked move(s) below are the LLM's own " +
			"low-confidence calls — it asked for your eye; " +
			"tick one back if you disagree with its doubt"
	}
	return out
}

// mapRows applies shaper over the folder list (the section builder's
// own helper).
func mapRows(folders []string, shaper func(string) PlanRow) []PlanRow {
	out := []PlanRow{}
	for _, f := range folders {
		out = append(out, shaper(f))
	}
	return out
}

// ---------------------------------------------------------------------------
// THE PICKING DOOR — the selection semantics (scan_plan_dialog.py's
// _boxes/_masters/_refresh_apply/_toggle_section/selection, as a pure
// state machine the shell renders and the tests pin).
// ---------------------------------------------------------------------------

// The master box's three states (the panel's mirror).
const (
	MasterChecked   = "checked"
	MasterUnchecked = "unchecked"
	MasterPartial   = "partial"
)

// PlanPickState is the picking door's own state: every section's row
// ids in display order + the checked map + the arm-guard's word. The
// reference's blockSignals sweep is a Qt mechanic — the Go twin is the
// ONE-recount law instead: ToggleSection mutates silently and the
// caller recomputes the label once.
type PlanPickState struct {
	order   map[string][]string
	checked map[string]map[string]bool
	armed   bool
}

// NewPlanPickState builds the door's born state over a display model:
// every row CHECKED except the LLM's own low-confidence calls (the
// guest's self-doubt is a default, never a surprise). The arm-guard
// starts UNARMED (both verdict buttons are born disabled — the
// fresh-click door); Arm() is the settle beat's end.
func NewPlanPickState(model *PlanDisplayModel) *PlanPickState {
	s := &PlanPickState{
		order:   map[string][]string{},
		checked: map[string]map[string]bool{},
	}
	if model == nil {
		return s
	}
	for _, sec := range model.Sections {
		s.order[sec.Key] = []string{}
		s.checked[sec.Key] = map[string]bool{}
		for _, row := range sec.Rows {
			s.order[sec.Key] = append(s.order[sec.Key], row.ID)
			s.checked[sec.Key][row.ID] = !row.LowConfidence
		}
	}
	return s
}

// IsChecked answers one row's tick.
func (s *PlanPickState) IsChecked(section, id string) bool {
	return s.checked[section] != nil && s.checked[section][id]
}

// SetChecked answers one row's tick change (the row's own click; the
// label's recount is the caller's — the reference's toggled →
// _refresh_apply).
func (s *PlanPickState) SetChecked(section, id string, on bool) {
	if s.checked[section] == nil {
		return
	}
	s.checked[section][id] = on
}

// ToggleSection is the master box's click (the ONE gesture every list
// UI already taught): any unticked row ticks the whole panel; an
// all-ticked panel unticks entirely. The sweep is silent (the
// blockSignals law) — the caller recounts ONCE at the end.
func (s *PlanPickState) ToggleSection(section string) {
	ids := s.order[section]
	if len(ids) == 0 {
		return
	}
	target := !s.allChecked(ids, section)
	for _, id := range ids {
		s.checked[section][id] = target
	}
}

func (s *PlanPickState) allChecked(ids []string, section string) bool {
	for _, id := range ids {
		if !s.checked[section][id] {
			return false
		}
	}
	return true
}

// MasterState is the master box's mirror of its panel: checked when
// all the rows ride, unchecked when none, PARTIAL when they disagree
// (displayed, never user-cycled — a click always means "make them all
// agree with the majority").
func (s *PlanPickState) MasterState(section string) string {
	ids := s.order[section]
	if len(ids) == 0 {
		return MasterUnchecked
	}
	all, any := true, false
	for _, id := range ids {
		if s.checked[section][id] {
			any = true
		} else {
			all = false
		}
	}
	switch {
	case all:
		return MasterChecked
	case any:
		return MasterPartial
	default:
		return MasterUnchecked
	}
}

// PickedTotals is (picked, total) across every panel — the Apply
// button's own arithmetic.
func (s *PlanPickState) PickedTotals() (int, int) {
	total, picked := 0, 0
	for _, section := range scanSectionKeys {
		for _, id := range s.order[section] {
			total++
			if s.checked[section][id] {
				picked++
			}
		}
	}
	return picked, total
}

// ApplyLabel is the Apply button's live word: the picked count over
// the whole plan.
func (s *PlanPickState) ApplyLabel() string {
	picked, total := s.PickedTotals()
	if total > 0 {
		return fmt.Sprintf("🗂️ Apply plan (%d of %d)", picked, total)
	}
	return "🗂️ Apply plan"
}

// ApplyEnabled is the enable law: armed AND (something picked, or
// nothing to pick). An UNARMED Apply stays dead whatever the
// arithmetic says — a fossil click is not consent.
func (s *PlanPickState) ApplyEnabled() bool {
	picked, total := s.PickedTotals()
	return s.armed && (picked > 0 || total == 0)
}

// KeepEnabled is the Keep door's own law: the arm-guard alone gates it
// (Keep is the safe default — nothing-picked never disables it).
func (s *PlanPickState) KeepEnabled() bool {
	return s.armed
}

// Arm ends the settle beat (the doors come alive; only LIVE hands can
// answer the plan from here).
func (s *PlanPickState) Arm() { s.armed = true }

// Armed answers the guard's word.
func (s *PlanPickState) Armed() bool { return s.armed }

// Selection is the picking door's answer: every section's CHECKED row
// ids, all four keys always riding (an empty list means the owner
// unticked every row of that panel). Only a CONFIRMED verdict ever
// carries it — the shape law (AskScanConfirmEvent.Answer enforces the
// pairing at the bus).
func (s *PlanPickState) Selection() map[string][]string {
	out := map[string][]string{
		ScanSectionDeletions:      {},
		ScanSectionConsolidations: {},
		ScanSectionNewFolders:     {},
		ScanSectionMoves:          {},
	}
	for _, section := range scanSectionKeys {
		ids := []string{}
		for _, id := range s.order[section] {
			if s.checked[section][id] {
				ids = append(ids, id)
			}
		}
		out[section] = ids
	}
	return out
}

// ---------------------------------------------------------------------------
// THE SCORECARD — result_display_model (scan_result_dialog.py:129-413).
// ---------------------------------------------------------------------------

// ScoreRow is one scorecard row: the plain count when everything
// landed, the honest 'K of N' (warning tone) when the machinery found
// less than the plan proposed.
type ScoreRow struct {
	Shown bool   `json:"shown"`
	Name  string `json:"name"`
	Value string `json:"value"`
	Tone  string `json:"tone"`
	Tip   string `json:"tip"`
}

// VaultSection is one vault's group (v0.70.0 — the per-vault
// scorecard): the label + the changed count + the rows that landed
// there. A section appears ONLY when something actually changed.
type VaultSection struct {
	Label   string     `json:"label"`
	Changed int        `json:"changed"`
	Tip     string     `json:"tip"`
	Rows    []ScoreRow `json:"rows"`
}

// ResultDisplayModel is the applied scorecard's whole display shape:
// the LANDED counts, never the plan's promises.
type ResultDisplayModel struct {
	Heading       string         `json:"heading"`
	Rows          []ScoreRow     `json:"rows"`
	VaultSections []VaultSection `json:"vault_sections"`
	LandedTotal   int            `json:"landed_total"`
	NothingLanded bool           `json:"nothing_landed"`
	FailedMoves   int            `json:"failed_moves"`
	NothingLine   string         `json:"nothing_line"`
	Footnote      string         `json:"footnote"`
}

// scoreRow is the reference's _row: hidden when neither side has a
// number, the honest 'K of N' + the shortfall tip when the plan
// promised more than the machinery landed, the plain count + the
// honest tip otherwise.
func scoreRow(name string, landed, planned int, honest, shortfall string) ScoreRow {
	if landed <= 0 && planned <= 0 {
		return ScoreRow{Shown: false}
	}
	if planned > landed && planned > 0 {
		return ScoreRow{Shown: true, Name: name,
			Value: fmt.Sprintf("%d of %d", landed, planned),
			Tone:  ToneWarning, Tip: shortfall}
	}
	return ScoreRow{Shown: true, Name: name, Value: fmt.Sprintf("%d", landed),
		Tone: ToneSuccess, Tip: honest}
}

// namedFailures builds the failed moves' NAMED tail (cap 3 + the …).
func namedFailures(failed []vaultscan.FailedMove) string {
	names := []string{}
	for _, f := range failed {
		if n := strings.TrimSpace(f.Note); n != "" {
			names = append(names, n)
		}
	}
	if len(names) == 0 {
		return ""
	}
	cap := 3
	if len(names) < cap {
		cap = len(names)
	}
	named := " — " + strings.Join(names[:cap], ", ")
	if len(names) > cap {
		named += "…"
	}
	return named
}

// hygieneHealOf is one vault's heal number (removed_index +
// retired_stubs + retired_reports).
func hygieneHealOf(counts vaulthygiene.HygieneCounts) int {
	return counts.RemovedIndex + counts.RetiredStubs + counts.RetiredReports
}

// hygieneRow is one vault's 🧹 row (None when that vault healed
// nothing): the detail tooltip carries the new fingerprints' own
// counts.
func hygieneRow(counts vaulthygiene.HygieneCounts, vaultPath string) *ScoreRow {
	total := hygieneHealOf(counts)
	if total <= 0 {
		return nil
	}
	detail := ""
	if counts.RetiredLinked > 0 {
		detail += fmt.Sprintf(" %d of them linked ghosts (empty notes other notes pointed at).",
			counts.RetiredLinked)
	}
	if counts.RetiredAppOwned > 0 {
		detail += fmt.Sprintf(" %d of them app-owned bodyless notes (failed writes that never got their content).",
			counts.RetiredAppOwned)
	}
	tip := fmt.Sprintf("%d stray file(s) retired to .trash in this vault (recoverable by hand) — the retired index scaffold, empty ghost-link notes, legacy root reports.%s The graph unwelds; deletions stop warning.",
		total, detail)
	if vaultPath != "" {
		tip += "\nVault: " + vaultPath
	}
	return &ScoreRow{Shown: true, Name: "🧹 Strays healed",
		Value: fmt.Sprintf("%d", total), Tone: ToneSuccess, Tip: tip}
}

// storyRows builds one story's four landed rows (the shared grammar
// both the flat list and the vault sections speak; the repos flavor
// names the vault in the two honest tips).
func storyRows(st *ScanStory, repos bool) []ScoreRow {
	banishedTip := fmt.Sprintf("%d note(s) resting in .trash/banished — never to be fetched again (♻️ on the record row revives)",
		st.Applied)
	mergedTip := fmt.Sprintf("%d duplicate note(s) pulled into their site's note — resting in .trash/consolidated, nothing blacklisted",
		st.NotesConsolidated)
	if repos {
		banishedTip = fmt.Sprintf("%d repo note(s) resting in .trash/banished — never to be fetched again (♻️ on the record row revives)",
			st.Applied)
		mergedTip = fmt.Sprintf("%d duplicate repo note(s) pulled into their repo's note — resting in .trash/consolidated, nothing blacklisted",
			st.NotesConsolidated)
	}
	rows := []ScoreRow{}
	rows = append(rows, scoreRow("🗑️ Removed (banished)",
		st.Applied, st.Deletions, banishedTip,
		"the rest were already gone or kept — the log holds each why"))
	rows = append(rows, scoreRow("🧲 Duplicates merged",
		st.NotesConsolidated, st.ConsolidationNotes, mergedTip,
		"the rest were already gone or kept — the log holds each why"))
	moveRow := scoreRow("🚚 Notes re-filed",
		st.Moved, st.Moves,
		"moved byte-identically — never a byte rewritten",
		"the moves that failed are in the log (the files stayed put)")
	if st.FailedMoves > 0 {
		// a move that failed is the machinery's own shortfall —
		// v0.82.0: the failed files are NAMED.
		moveRow.Tip = fmt.Sprintf("%d move(s) failed%s — the files stayed put; the log says why",
			st.FailedMoves, namedFailures(st.Failed))
		moveRow.Tone = ToneWarning
	}
	rows = append(rows, moveRow)
	rows = append(rows, scoreRow("🌱 New folders created",
		st.FoldersCreated, st.NewFolders,
		"created empty and waiting for their notes",
		"the rest already existed — nothing created twice"))
	return rows
}

// ResultDisplayModelOf is result_display_model — PURE: the scan
// job's story in, the scorecard's rows out (never raises; a nil story
// degrades to the nothing-landed glance).
func ResultDisplayModelOf(res *ScanResult) *ResultDisplayModel {
	out := &ResultDisplayModel{
		NothingLine: "Every mark had already healed by hand, or was a " +
			"hand-written note (kept — always), or its file was " +
			"already gone. The vault is exactly as the plan found it; " +
			"the full story is in the log.",
		Footnote: "♻️ Every retired note rests in .trash (banished or " +
			"consolidated) — recoverable by hand; a banished site " +
			"never fetches again, a merged site was never " +
			"blacklisted.",
	}
	if res == nil {
		out.Heading = "The plan is applied — nothing needed the machinery"
		out.NothingLanded = true
		return out
	}
	story := res.Story
	if story == nil {
		story = &ScanStory{}
	}
	applied, consolidated := story.Applied, story.NotesConsolidated
	moved, folders := story.Moved, story.FoldersCreated
	failedMoves := story.FailedMoves

	// the hygiene heal (both machine vaults).
	healed := res.HygieneTotal
	var websitesCounts, githubCounts vaulthygiene.HygieneCounts
	if res.Hygiene != nil {
		websitesCounts = res.Hygiene["websites"]
		githubCounts = res.Hygiene["github"]
	}
	websitesHygiene := hygieneHealOf(websitesCounts)
	githubChanged := hygieneHealOf(githubCounts)

	rows := storyRows(story, false)
	if healed > 0 {
		linked := websitesCounts.RetiredLinked + githubCounts.RetiredLinked
		appOwned := websitesCounts.RetiredAppOwned + githubCounts.RetiredAppOwned
		detail := ""
		if linked > 0 {
			detail += fmt.Sprintf(" %d of them linked ghosts (empty notes other notes pointed at).", linked)
		}
		if appOwned > 0 {
			detail += fmt.Sprintf(" %d of them app-owned bodyless notes (failed writes that never got their content).", appOwned)
		}
		rows = append(rows, ScoreRow{Shown: true,
			Name: "🧹 Strays healed", Value: fmt.Sprintf("%d", healed),
			Tone: ToneSuccess,
			Tip: fmt.Sprintf("%d stray file(s) — the retired index scaffold, empty ghost-link notes, legacy root reports — retired to .trash across your machine vaults (recoverable by hand).%s The graph unwelds; deletions stop warning.",
				healed, detail)})
	}
	shown := []ScoreRow{}
	for _, r := range rows {
		if r.Shown {
			shown = append(shown, r)
		}
	}
	out.Rows = shown

	landedTotal := applied + consolidated + moved + folders
	nothingLanded := landedTotal <= 0 && failedMoves == 0 && healed <= 0
	out.LandedTotal = landedTotal
	out.NothingLanded = nothingLanded
	out.FailedMoves = failedMoves
	switch {
	case healed > 0 && landedTotal <= 0 && failedMoves == 0:
		out.Heading = "The vault is healed — the plan had nothing to do"
	case nothingLanded:
		out.Heading = "The plan is applied — nothing needed the machinery"
	default:
		out.Heading = "The plan is applied"
	}

	// ---- v0.70/0.83 — the per-vault sections -----------------------
	kind := strings.ToLower(strings.TrimSpace(res.VaultKind))
	if kind == "" {
		kind = "websites"
	}
	label := res.VaultLabel
	if label == "" {
		label = "Websites vault"
		if kind == "repos" {
			label = "GitHub Repos vault"
		}
	}
	planRows := []ScoreRow{}
	for _, r := range shown {
		if r.Name != "🧹 Strays healed" {
			planRows = append(planRows, r)
		}
	}
	sections := []VaultSection{}
	if len(planRows) > 0 || websitesHygiene > 0 {
		mainHygiene := hygieneRow(websitesCounts, res.Vault)
		if kind == "repos" {
			mainHygiene = hygieneRow(githubCounts, res.GitHubVault)
		}
		secRows := planRows
		if mainHygiene != nil {
			secRows = append(append([]ScoreRow{}, planRows...), *mainHygiene)
		}
		changed := landedTotal + websitesHygiene
		tip := fmt.Sprintf("The plan applied here: %d change(s) landed in this vault.", changed)
		if res.Vault != "" {
			tip += "\nPath: " + res.Vault
		}
		sections = append(sections, VaultSection{
			Label: "📁 " + label, Changed: changed, Tip: tip, Rows: secRows})
	}
	if res.Repos != nil {
		rRows := storyRows(res.Repos, true)
		rShown := []ScoreRow{}
		for _, r := range rRows {
			if r.Shown {
				rShown = append(rShown, r)
			}
		}
		if kind != "repos" {
			if gh := hygieneRow(githubCounts, res.GitHubVault); gh != nil {
				rShown = append(rShown, *gh)
			}
		}
		rChanged := res.Repos.Applied + res.Repos.Moved +
			res.Repos.FoldersCreated + res.Repos.NotesConsolidated
		if kind != "repos" {
			rChanged += githubChanged
		}
		if len(rShown) > 0 {
			tip := fmt.Sprintf("The repos plan applied here: %d change(s) landed in this vault.", rChanged)
			if res.GitHubVault != "" {
				tip += "\nPath: " + res.GitHubVault
			}
			sections = append(sections, VaultSection{
				Label: "📁 GitHub Repos vault", Changed: rChanged,
				Tip: tip, Rows: rShown})
		}
	} else if githubChanged > 0 {
		if gh := hygieneRow(githubCounts, res.GitHubVault); gh != nil {
			sections = append(sections, VaultSection{
				Label: "📁 GitHub projects vault", Changed: githubChanged,
				Tip: fmt.Sprintf("%d stray file(s) healed in this vault by the scan's hygiene pass — recoverable in .trash.",
					githubChanged),
				Rows: []ScoreRow{*gh}})
		}
	}
	out.VaultSections = sections
	return out
}

// strOr is the model's own junk tolerance (an empty string reads as
// the fallback).
func strOr(s, def string) string {
	if strings.TrimSpace(s) != "" {
		return s
	}
	return def
}
