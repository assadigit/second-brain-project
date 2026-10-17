package vaultscan

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// AbsorbNearDuplicate mirrors _absorb_near_duplicate (v0.63.0 — THE
// FRESH-EYES LAW's lookup): which existing folder (if any) already
// covers the proposed one? Exact paths folded equal match first; then
// a near-identical SIBLING — same folded parent, leaf names that
// contain one another (both at least 5 folded chars). Sorted iteration
// keeps the pick deterministic. "" when the library genuinely lacks
// the category.
func AbsorbNearDuplicate(clean string, existing []string) string {
	sorted := append([]string{}, existing...)
	sort.Strings(sorted)
	for _, ex := range sorted {
		if foldPath(ex) == foldPath(clean) {
			return ex
		}
	}
	parts := strings.Split(clean, "/")
	parentFold := foldPath(strings.Join(parts[:len(parts)-1], "/"))
	leafFold := foldSegment(parts[len(parts)-1])
	if len(leafFold) < 5 {
		return "" // too short to match by containment
	}
	for _, ex := range sorted {
		exParts := strings.Split(ex, "/")
		if len(exParts) < 2 ||
			foldPath(strings.Join(exParts[:len(exParts)-1], "/")) != parentFold {
			continue // not a sibling under the same parent
		}
		exLeaf := foldSegment(exParts[len(exParts)-1])
		if len(exLeaf) >= 5 && (strings.Contains(exLeaf, leafFold) ||
			strings.Contains(leafFold, exLeaf)) {
			return ex
		}
	}
	return ""
}

// TreeRoots mirrors _tree_roots (v0.63.0): every folder the tree
// implies, ancestors included — a parent that holds no note of its own
// is still a real place to hang a new subfolder under.
func TreeRoots(existing map[string]bool) map[string]bool {
	roots := map[string]bool{}
	for rel := range existing {
		roots[rel] = true
	}
	for rel := range existing {
		parts := strings.Split(rel, "/")
		for i := 1; i < len(parts); i++ {
			roots[strings.Join(parts[:i], "/")] = true
		}
	}
	return roots
}

// BendPath mirrors _bend_path (v0.82.0 — THE LANDABLE MOVE's name
// bend): the house's own uniquify law (Name_v1.md, Name_v2.md, …)
// extended over the plan's OWN promised destinations, so two notes of
// one name filing into one folder bend past each other at plan time —
// never at apply time. Pure.
func BendPath(dst string, planned map[string]bool) string {
	if !fsExists(dst) && !planned[dst] {
		return dst
	}
	ext := filepath.Ext(dst)
	base := strings.TrimSuffix(dst, ext)
	for i := 1; ; i++ {
		cand := fmt.Sprintf("%s_v%d%s", base, i, ext)
		if !fsExists(cand) && !planned[cand] {
			return cand
		}
	}
}

// MoveRow mirrors the validated plan's move rows:
// {'note', 'path', 'title', 'from', 'to', 'as', 'reason',
// 'low_confidence'}.
type MoveRow struct {
	Note          string
	Path          string
	Title         string
	From          string
	To            string
	As            string
	Reason        string
	LowConfidence bool
}

// ValidateMoves mirrors _validate_moves: the guest's proposal meets
// the vault's laws — only listed candidates move; destinations are
// existing folders (or the deterministic passes' disk-verified
// extras) or sanitized new ones; the same note never moves twice; a
// move to the note's own folder is dropped; the three laws (fresh
// eyes, depth, core notes) bite HERE; the low-confidence flag rides
// the row (the guest's own word, or the weighed evidence's override
// check); and a colliding landing name is pre-bent at plan time (the
// row carries 'as'). Returns (moves, new_folders).
func ValidateMoves(data LLMProposal, candidates []*NoteRow,
	inv *Inventory, log LogFn, extraFolders []string,
	vaultPath, kind string) ([]MoveRow, []string) {
	if log == nil {
		log = noopLog
	}
	existing := map[string]bool{}
	for _, f := range inv.Folders {
		existing[f.Rel] = true
	}
	roots := TreeRoots(existing) // ancestors are parents too
	// the deterministic passes' destinations: real folders the
	// note-count walk never saw.
	for _, extra := range extraFolders {
		clean := SanitizeFolder(extra)
		if clean != "" && !existing[clean] {
			existing[clean] = true
			roots[clean] = true
		}
	}
	byName := map[string]*NoteRow{}
	for _, cand := range candidates {
		if _, ok := byName[cand.File]; !ok {
			byName[cand.File] = cand
		}
		if _, ok := byName[cand.Rel]; !ok {
			byName[cand.Rel] = cand
		}
	}
	var newFolders []string
	alias := map[string]string{} // proposed path → the covering folder
	seenNew := map[string]bool{}
	seenFold := map[string]string{}
	for _, nf := range data.NewFolders {
		clean := SanitizeFolder(nf)
		if clean == "" || seenNew[clean] {
			continue
		}
		if !strings.Contains(clean, "/") {
			log(fmt.Sprintf("⚠️ Scan plan: new folder “%s” is a bare "+
				"root cell — new folders are SUBFOLDERS under a category "+
				"(the depth law) — dropped", nf), "warning")
			continue
		}
		parent := strings.Join(strings.Split(clean, "/")[:len(strings.Split(clean, "/"))-1], "/")
		if parent != "" && !roots[parent] && !inList(parent, newFolders) {
			log(fmt.Sprintf("⚠️ Scan plan: new folder “%s” has no parent "+
				"in the tree — dropped", nf), "warning")
			continue
		}
		existingList := make([]string, 0, len(existing))
		for e := range existing {
			existingList = append(existingList, e)
		}
		if hit := AbsorbNearDuplicate(clean, existingList); hit != "" {
			alias[clean] = hit
			seenNew[clean] = true
			log(fmt.Sprintf("♻️ Scan plan: “%s” already exists as “%s” "+
				"(the same or a near-identical category) — filing into "+
				"the existing folder instead, nothing duplicate created",
				clean, hit), "info")
			continue
		}
		if planHit, ok := seenFold[foldPath(clean)]; ok && planHit != clean {
			alias[clean] = planHit
			seenNew[clean] = true
			log(fmt.Sprintf("♻️ Scan plan: “%s” duplicates “%s” from the "+
				"same plan — filing into the one folder instead",
				clean, planHit), "info")
			continue
		}
		seenNew[clean] = true
		seenFold[foldPath(clean)] = clean
		newFolders = append(newFolders, clean)
	}
	legal := map[string]bool{}
	for f := range existing {
		legal[f] = true
	}
	for _, nf := range newFolders {
		legal[nf] = true
	}
	var moves []MoveRow
	taken := map[string]bool{}
	planned := map[string]bool{} // the promised destination filenames
	for _, m := range data.Moves {
		noteRef := helpers.Strip(m.Note)
		dest := SanitizeFolder(helpers.Strip(m.To))
		if dest != "" {
			if a, ok := alias[dest]; ok {
				dest = a // the fresh-eyes re-route
			}
		}
		note := byName[noteRef]
		if note == nil || taken[noteRef] {
			continue // not a listed candidate — never our call
		}
		if dest == "" || !legal[dest] {
			log(fmt.Sprintf("⚠️ Scan plan: “%s” → “%s” is not a legal "+
				"destination — dropped", noteRef, m.To), "warning")
			continue
		}
		if !strings.Contains(dest, "/") {
			what := "website"
			if isReposKind(kind) {
				what = "repo note"
			}
			log(fmt.Sprintf("⚠️ Scan plan: “%s” → “%s” files a note "+
				"into a bare category root — every %s lives under a "+
				"subfolder (Category/Subfolder) — dropped", noteRef, dest,
				what), "warning")
			continue
		}
		if foldPath(dest) == foldPath(note.Folder) {
			continue // already there (folded identity)
		}
		taken[noteRef] = true
		// v0.82.0 — THE LANDABLE MOVE: the plan never proposes a move
		// the apply must refuse. A destination that already holds a
		// note of the same name is resolved HERE — the move rides under
		// the house's own bend and the row carries its landing name as
		// 'as'; a destination that IS the note itself is already home.
		asName := ""
		if vaultPath != "" {
			npath := note.Path
			base := ""
			if npath != "" {
				base = filepath.Base(npath)
			}
			if base != "" {
				dst := filepath.Join(append([]string{vaultPath},
					append(strings.Split(dest, "/"), base)...)...)
				if fsExists(dst) || planned[dst] {
					same := false
					if fsExists(dst) && npath != "" && fsIsFile(npath) {
						if fi1, err1 := os.Stat(npath); err1 == nil {
							if fi2, err2 := os.Stat(dst); err2 == nil {
								same = os.SameFile(fi1, fi2)
							}
						}
					}
					if same {
						continue // the destination IS this note — already
						// home, never a move
					}
					bent := BendPath(dst, planned)
					asName = filepath.Base(bent)
					if asName != base {
						log(fmt.Sprintf("♻️ Scan plan: a note named %s "+
							"already lives in %s — the move rides as %s "+
							"(nothing is ever overwritten; the same _v "+
							"bend a new note takes at creation)", base,
							dest, asName), "info")
					}
					planned[bent] = true
				} else {
					planned[dst] = true
				}
			}
		}
		// v0.73.0 — THE LLM'S OWN CONFIDENCE: the guest's own word
		// rides the plan (a "low" verdict flags the row; silence or any
		// other word reads as high).
		low := strings.ToLower(helpers.Strip(m.Confidence)) == "low"
		// v0.74.0 — THE WEIGHED EVIDENCE's alignment check: when the
		// evidence weighed a STRONG top home and the guest filed the
		// note somewhere clearly lighter (or somewhere unlisted), the
		// move rides flagged low_confidence too. Folded comparison.
		if len(note.Homes) > 0 {
			top := note.Homes[0]
			if top.Weight >= StrongEvidenceWeight &&
				foldPath(dest) != foldPath(top.Folder) {
				chosenW := -1.0
				for _, h := range note.Homes {
					if foldPath(dest) == foldPath(h.Folder) {
						chosenW = h.Weight
						break
					}
				}
				if chosenW < 0 || top.Weight-chosenW >= OverrideMargin {
					low = true
					log(fmt.Sprintf("⚖️ Scan plan: “%s” → “%s” rides "+
						"against the weighed evidence (the top home was "+
						"“%s” at %.2f) — shown un-ticked; the owner "+
						"decides", noteRef, dest, top.Folder, top.Weight),
						"info")
				}
			}
		}
		title := note.Title
		if title == "" {
			title = noteRef
		}
		from := note.Folder
		if from == "" {
			from = "(vault root)"
		}
		moves = append(moves, MoveRow{
			Note: noteRef, Path: note.Path,
			Title:         title,
			From:          from,
			To:            dest,
			As:            asName,
			Reason:        cutRunesStr(m.Reason, 200),
			LowConfidence: low,
		})
	}
	return moves, newFolders
}
