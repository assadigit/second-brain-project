package app

// decomdoors.go — T-79e: the decommission picker's engines (the More
// menu's fifth family — THE LAST tail row), pinned against the
// reference's own flow (processing_control.py:807-998 at 98f4fc7):
// decommission_dead_links_now (the guards → the backlog-alive scan →
// the candidates pre-fill → the picker → the marking → the immediate
// enforcement → the report) + _scan_backlog_alive (152-173) + the
// picker's "Open the table" reveal. The engine's own writers and the
// enforcement (write_decommission_candidates /
// mark_urls_dead_in_table / consume_decommission_table /
// scan_review_backlog) are LIVE in internal/websitepipeline; this
// door is their anytime trigger, the Settings-method shape of
// T-79a/b/c/d.
//
// The one-ledger law (the task row's own words): the picker writes
// the SAME Status cells the owner would edit by hand in Obsidian —
// the table stays the ONE ledger; the door then enforces the burials
// immediately (the reference's own comment: the same state DB a batch
// uses; the dry-run law does not apply — this is the OWNER's explicit
// hand, and the next batch would enforce it anyway). ♻️ revived on a
// row brings a link back to life (the same consume pass enforces it).
//
// The recorded divergences (the T-79c/d style — evidence doc
// docs/poc-artifacts/t-79e-decommission-doors.md):
//   - The reference's dry-run shadow-cache branch (WebsiteStateDB over
//     _dryrun.shadow_cache_path, both in _scan_backlog_alive and the
//     enforcement) is dead in its own GUI (the switch is CLI-only;
//     the GUI never flips it) — the port's door reads the live cache,
//     the same behavior in every reachable GUI state (the T-79d class).
//   - The writers' own expected-failure warnings (⚠️ Could not
//     create/read/write the graveyard table) speak through the passed
//     log exactly like the reference's; the reference's outer
//     try/except belts (⚠️ Graveyard candidates write skipped /
//     marking skipped) fire only on UNEXPECTED exceptions — the Go
//     doors carry the same belts as recover() (never a crash).
//   - The picker dialog is the frontend's (the two-door shape of the
//     hand-deliver flow); the picked URLs ride back as the dialog's
//     own data (the reference's picked-from-candidates law).
//   - The "Open the table" reveal rides OpenDecommissionTable over
//     the existing OpenFileLocation seam (the reference's
//     QDesktopServices.openUrl twin; best-effort, never raises — the
//     reference's ⚠️ Could-not-open line is unreachable through the
//     seam's never-raise contract). The _review dir is created when
//     missing — the OpenHandDeliveryFolder sibling's own law.
//   - NO busy guard: the reference's door carries none (unlike the
//     hand/Chrome doors) — the enforcement rides the same SQLite
//     state DB a batch uses and the reference accepts the overlap by
//     design; pinned by test (the guard-ladder order: guards → scan).

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// ---------------------------------------------------------------------------
// The payloads (THE FIELD-TAG LAW — the tags mirror the keys the
// frontend reads; pinned by TestDecommissionDoorPayloadFieldTags).
// ---------------------------------------------------------------------------

// DecommissionScan is 🪦 Decommission dead links' scan verdict: the
// guards'/empty box, or the picker's own data — the sorted, deduped
// waiting URLs plus the graveyard table path (the picker's ledger
// label + the reveal door's target).
type DecommissionScan struct {
	Box   *HeroBox `json:"box"`
	Links []string `json:"links"`
	Table string   `json:"table"`
}

// ---------------------------------------------------------------------------
// _scan_backlog_alive (processing_control.py:152-173).
// ---------------------------------------------------------------------------

// scanBacklogAlive is the backlog scan minus the RETIRED links: the
// state DB's is_dismissed probe filters the fetcher's own auto-verdict
// retirements; a broken DB falls back to the pure file read (a broken
// probe never hides a waiting link from the owner).
func (s *Settings) scanBacklogAlive(vault string) []websitepipeline.BacklogItem {
	if state, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(s.cfg)); err == nil {
		items := websitepipeline.ScanReviewBacklog(vault, nil, func(u string) bool {
			hit, _ := state.IsDismissed(u)
			return hit
		})
		state.Close()
		return items
	}
	return websitepipeline.ScanReviewBacklog(vault, nil, nil)
}

// ---------------------------------------------------------------------------
// 🪦 Decommission dead links (processing_control.py:807).
// ---------------------------------------------------------------------------

// DecommissionScan is decommission_dead_links_now's first half: the
// two guards (the shared Websites ladder), the backlog-alive scan, the
// empty box in the reference's own words (the graveyard stays
// available for hand-rows), and the candidates pre-fill (the same
// try/except belt — the write never blocks the picker).
func (s *Settings) DecommissionScan() *DecommissionScan {
	out := &DecommissionScan{}
	if box := s.chromeWebsitesGuard(); box != nil {
		out.Box = box
		return out
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	out.Table = websitepipeline.DecommissionTablePath(vault)
	items := s.scanBacklogAlive(vault)
	if len(items) == 0 {
		out.Box = &HeroBox{Title: "No failed links waiting",
			Body: "No app-owned fetch-failed placeholders are waiting in _review — nothing new to decommission.\n\nThe graveyard table stays available for hand-rows:\n" +
				out.Table + "\n\n(Set a row's Status to 🪦 dead anytime — even a link that never got a placeholder. ♻️ revived brings one back.)"}
		return out
	}
	// urls = sorted({i['url'] for i in items}) — the dedup + sort law.
	seen := map[string]bool{}
	for _, it := range items {
		seen[it.URL] = true
	}
	urls := make([]string, 0, len(seen))
	for u := range seen {
		urls = append(urls, u)
	}
	sort.Strings(urls)
	// the candidates pre-fill (the reference's own order: write, then
	// the picker; the belt only ever speaks on the unexpected).
	func() {
		defer func() {
			if r := recover(); r != nil {
				s.SpeakOnBus(fmt.Sprintf("⚠️ Graveyard candidates write skipped: %v", r), "warning")
			}
		}()
		websitepipeline.WriteDecommissionCandidates(vault, urls, "retry backlog",
			func(msg, level string) { s.SpeakOnBus(msg, level) }, nil, "unreviewed")
	}()
	out.Links = urls
	return out
}

// DecommissionBury is the picker's answer (the reference's
// post-dialog half): the picked URLs get their graveyard-table Status
// cells written (the SAME cells the owner would edit by hand — the
// table stays the one ledger), then the burials are enforced
// immediately (dismiss + sweep through the same state DB a batch
// uses). An empty list is the dialog's Cancel — the reference's
// cancel line speaks on the desk. Title "" = cancelled (no box).
func (s *Settings) DecommissionBury(picked []string) (title, body string) {
	if len(picked) == 0 {
		s.SpeakOnBus("⏭️ Decommission cancelled — nothing was buried. The candidate rows wait in the table (set Status to 🪦 dead by hand in Obsidian; the next batch enforces it).", "info")
		return "", ""
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	table := websitepipeline.DecommissionTablePath(vault)
	// the marking (the belt only ever speaks on the unexpected; the
	// writer's own ⚠/🪦 lines ride the passed log).
	func() {
		defer func() {
			if r := recover(); r != nil {
				s.SpeakOnBus(fmt.Sprintf("⚠️ Graveyard marking skipped: %v", r), "warning")
			}
		}()
		websitepipeline.MarkURLsDeadInTable(vault, picked,
			func(msg, level string) { s.SpeakOnBus(msg, level) })
	}()
	// Enforce now (the same state DB a batch uses; the dry-run law
	// does not apply — this is the OWNER's explicit hand, and the next
	// batch would enforce it anyway). A failure defers honestly to the
	// next batch — never a crash, the box says so.
	var report *websitepipeline.ConsumeReport
	if state, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(s.cfg)); err == nil {
		rep := websitepipeline.ConsumeDecommissionTable(state, vault,
			func(msg, level string) { s.SpeakOnBus(msg, level) }, true, nil)
		report = &rep
		state.Close()
	} else {
		s.SpeakOnBus(fmt.Sprintf("⚠️ Graveyard enforcement deferred to the next batch: %v", err), "warning")
	}
	buried, swept := 0, 0
	if report != nil {
		buried, swept = report.Dead, report.PlaceholdersSwept
	}
	remaining := len(s.scanBacklogAlive(vault))
	line := fmt.Sprintf("🪦 Graveyard: %d link(s) decommissioned", buried)
	if swept > 0 {
		line += fmt.Sprintf(", %d placeholder(s) swept", swept)
	}
	if remaining > 0 {
		line += fmt.Sprintf(" — %d still waiting in _review", remaining)
	} else {
		line += " — the _review backlog is clear"
	}
	s.SpeakOnBus(line, "success")
	out := fmt.Sprintf("%d link(s) marked dead", len(picked))
	if report != nil {
		out += fmt.Sprintf(" — %d enforced now, %d placeholder(s) swept.", buried, swept)
	} else {
		out += " — the next batch enforces the burials."
	}
	if remaining > 0 {
		out += fmt.Sprintf("\n\n%d link(s) still wait in _review.", remaining)
	} else {
		out += "\n\nThe _review backlog is clear — no more notices."
	}
	out += "\n\nThe ledger: " + table + "\n(♻️ revived on a row brings a link back to life.)"
	return "Decommissioned", out
}

// OpenDecommissionTable is the picker's "Open the table" reveal (the
// reference's _open_table): the graveyard table revealed in the OS
// file manager for hand-editing — best-effort over the existing
// OpenFileLocation seam (the folder is created when missing, the
// OpenHandDeliveryFolder sibling's own law).
func (s *Settings) OpenDecommissionTable() bool {
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))
	if vault == "" {
		return false
	}
	table := websitepipeline.DecommissionTablePath(vault)
	_ = os.MkdirAll(filepath.Dir(table), 0o755)
	if s.Deps.OpenFileLocation == nil {
		return false
	}
	s.Deps.OpenFileLocation(table)
	return true
}
