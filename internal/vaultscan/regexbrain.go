package vaultscan

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/taxonomy"
)

// RawMove mirrors the deterministic passes' move rows (and the guest's
// own): {'note', 'to', 'reason'} (+ the guest's optional confidence
// word, carried by the JSON door).
type RawMove struct {
	Note       string
	To         string
	Reason     string
	Confidence string // the guest's own word; "" reads as high
}

// DeclaredHomeMoves mirrors _declared_home_moves (v0.70.0 — THE
// DECLARED-HOME PASS): candidates whose own frontmatter declares
// category+subcategory (websites) or a CATEGORY_FOLDERS key (repos)
// file into the EXISTING folder deterministically — no LLM asked.
// Returns the moves plus the verified destination folders (they exist
// on disk — the caller injects them into the validator's legal set).
func DeclaredHomeMoves(candidates []*NoteRow, vaultPath string, log LogFn,
	kind string) ([]RawMove, []string) {
	var moves []RawMove
	var extraFolders []string
	func() {
		defer func() { recover() }() // the reference's outer try/except
		if vaultPath == "" {
			return
		}
		repos := isReposKind(kind)
		catMap := map[string]string{}
		if repos {
			catMap = reposCategoryFolders()
		}
		foldedKnown := map[string]*string{}
		for _, cand := range candidates {
			cat := helpers.Strip(cand.Category)
			sub := helpers.Strip(cand.Subcategory)
			var dest, reason string
			if repos {
				if cat == "" {
					continue // no declared key — the LLM's call
				}
				destRaw := catMap[cat]
				if destRaw == "" {
					continue // a stale/unmapped key — the LLM's call
				}
				dest = SanitizeFolder(destRaw)
				reason = "the note's own frontmatter declares category " +
					cat + " → " + destRaw
			} else {
				if cat == "" || sub == "" {
					continue // not a declared home — the LLM's call
				}
				dest = SanitizeFolder(cat + "/" + sub)
				reason = "the note's own frontmatter declares " + cat +
					"/" + sub
			}
			if dest == "" || !strings.Contains(dest, "/") {
				continue // illegal cells — never our call
			}
			// only EXISTING folders, checked on the LIVE tree; the
			// disk's own spelling wins.
			resolved := ResolveDiskFolder(vaultPath, dest, foldedKnown)
			if resolved == nil {
				continue // the declared home is not on disk — LLM
			}
			dest = *resolved
			if foldPath(dest) == foldPath(cand.Folder) {
				continue // already there — FOLDED (v0.82.0)
			}
			note := cand.File
			if note == "" {
				note = cand.Rel
			}
			moves = append(moves, RawMove{Note: note, To: dest, Reason: reason})
			if !inList(dest, extraFolders) {
				extraFolders = append(extraFolders, dest)
			}
		}
		if len(moves) > 0 {
			if repos {
				log(fmt.Sprintf("🗂️ %d candidate repo note(s) carry "+
					"their home in their own frontmatter (the "+
					"category key the intake wrote) — filed by the "+
					"regex brain, no LLM asked (the scan's speed law: "+
					"the LLM only does the thinking part)", len(moves)), "info")
			} else {
				log(fmt.Sprintf("🗂️ %d candidate note(s) carry their home in "+
					"their own frontmatter (category/subcategory) — filed "+
					"by the regex brain, no LLM asked (the scan's speed "+
					"law: the LLM only does the thinking part)", len(moves)), "info")
			}
		}
	}()
	return moves, extraFolders
}

// DomainHomes mirrors _settled_domain_homes' homes map:
// {domain: {folded_home: count}} with the per-domain insertion order
// kept for determinism.
type DomainHomes struct {
	Homes    map[string]map[string]int
	Order    []string            // domains, first-seen order
	PerOrder map[string][]string // domain → folded homes, first-seen
	Spelling map[string]string   // "domain\x00folded" → one disk spelling
}

// SettledDomainHomes mirrors _settled_domain_homes (v0.74.0 — the
// weighed evidence's shared teaching set): for every domain, the homes
// its SETTLED notes hold — app-owned, not a candidate, not protected
// or core, not a failed placeholder, filed at least Cat/Sub deep, not
// in the parking lot. Pure; never raises.
func SettledDomainHomes(inv *Inventory, excludePaths map[string]bool) *DomainHomes {
	out := &DomainHomes{
		Homes:    map[string]map[string]int{},
		Order:    []string{},
		PerOrder: map[string][]string{},
		Spelling: map[string]string{},
	}
	for _, n := range inv.Notes {
		if excludePaths[n.Path] {
			continue // a candidate teaches nothing
		}
		if !n.AppOwned || n.Protected || n.Core {
			continue
		}
		fs := n.FetchStatus
		if fs == "" {
			fs = "full"
		}
		if fs == "failed" {
			continue // a failed placeholder's spot is not a home
		}
		folder := n.Folder
		if !strings.Contains(folder, "/") {
			continue // the depth law: a bare category root is not a home
		}
		domain := strings.ToLower(helpers.Strip(n.Domain))
		if domain == "" {
			continue
		}
		folded := foldPath(folder)
		if folded == "" || strings.HasPrefix(strings.ToLower(folded), "uncategorized/") {
			continue // the parking lot is nobody's home
		}
		if _, ok := out.Homes[domain]; !ok {
			out.Homes[domain] = map[string]int{}
			out.Order = append(out.Order, domain)
		}
		if _, seen := out.Homes[domain][folded]; !seen {
			out.PerOrder[domain] = append(out.PerOrder[domain], folded)
		}
		out.Homes[domain][folded]++
		key := domain + "\x00" + folded
		if _, ok := out.Spelling[key]; !ok {
			out.Spelling[key] = folder
		}
	}
	return out
}

// SiblingDomainMoves mirrors _sibling_domain_moves (v0.71.0 — THE
// SIBLING-DOMAIN PASS): a candidate whose domain already has SETTLED
// notes under ONE agreed home joins them by pure Python. A domain
// split across homes is an AMBIGUOUS signal — the LLM's plate.
func SiblingDomainMoves(candidates []*NoteRow, inv *Inventory,
	vaultPath string, log LogFn) ([]RawMove, []string) {
	var moves []RawMove
	var extraFolders []string
	func() {
		defer func() { recover() }() // the reference's outer try/except
		if vaultPath == "" || len(candidates) == 0 {
			return
		}
		candidatePaths := map[string]bool{}
		for _, n := range candidates {
			candidatePaths[n.Path] = true
		}
		homes := SettledDomainHomes(inv, candidatePaths)
		if len(homes.Homes) == 0 {
			return
		}
		foldedKnown := map[string]*string{}
		// homeSpelling is a dict in the reference — iteration order
		// must be deterministic: sorted keys (the reference finds the
		// FIRST key whose domain prefix + folded spelling match; the
		// keys are unique per (domain, folded) so order only matters
		// for identical prefixes, which cannot collide).
		for _, cand := range candidates {
			domain := strings.ToLower(helpers.Strip(cand.Domain))
			if domain == "" {
				continue
			}
			per, ok := homes.Homes[domain]
			if !ok {
				continue // no settled siblings — the LLM's call
			}
			if len(per) != 1 {
				continue // the domain is split — a DECISION
			}
			folded := homes.PerOrder[domain][0]
			settled := per[folded]
			// the on-disk spelling of that agreed home (folded lookup —
			// the sibling's own folder, verified live). The spelling
			// map's key IS "domain\x00folded", so the reference's
			// prefix scan selects exactly this key.
			dest := (*string)(nil)
			if spelling, ok := homes.Spelling[domain+"\x00"+folded]; ok {
				if foldPath(spelling) == folded {
					if r := ResolveDiskFolder(vaultPath, spelling, foldedKnown); r != nil {
						dest = r
					}
				}
			}
			if dest == nil {
				continue // the home vanished mid-scan — LLM
			}
			if foldPath(*dest) == foldPath(cand.Folder) {
				continue // already there — FOLDED (v0.82.0)
			}
			note := cand.File
			if note == "" {
				note = cand.Rel
			}
			moves = append(moves, RawMove{
				Note: note, To: *dest,
				Reason: fmt.Sprintf("%d settled sibling note(s) of %s "+
					"live in %s — the domain's own established home",
					settled, domain, *dest),
			})
			if !inList(*dest, extraFolders) {
				extraFolders = append(extraFolders, *dest)
			}
		}
		if len(moves) > 0 {
			log(fmt.Sprintf("🧭 %d candidate note(s) carry a domain whose "+
				"settled notes already agree on one home — filed beside "+
				"their siblings by the regex brain, no LLM asked",
				len(moves)), "info")
		}
	}()
	return moves, extraFolders
}

// TagHomeInfo mirrors the vocabulary's value triple:
// (cat_name, sub_name, display tags).
type TagHomeInfo struct {
	CatName string
	SubName string
}

// TaxonomyVocabulary mirrors _taxonomy_vocabulary (v0.71.0 — THE
// TAXONOMY PASS's map): folded tag → the subcategory homes whose own
// vocabulary contains it (the subcategory's NAME words + its
// definition's words, folded, stopwords and sub-3-letter fragments
// dropped). Deterministic per-home insertion order.
func TaxonomyVocabulary(config map[string]any) map[string]map[string]TagHomeInfo {
	vocab := map[string]map[string]TagHomeInfo{}
	tax, err := taxonomy.LoadTaxonomyFromConfig(config)
	if err != nil {
		return vocab
	}
	wordRE := regexp.MustCompile(`[a-z0-9]+`)
	for _, cat := range tax.Categories() {
		catFolder := storage.SafeFilename(cat.Name)
		if catFolder == "" {
			continue
		}
		for _, sub := range cat.Subcategories {
			subFolder := storage.SafeFilename(sub)
			if subFolder == "" {
				continue
			}
			home := catFolder + "/" + subFolder
			var words []string
			words = append(words, wordRE.FindAllString(
				helpers.PyLower(sub), -1)...)
			def := cat.SubcategoryDefinitions[sub]
			if def == "" {
				def = ""
			}
			words = append(words, wordRE.FindAllString(
				helpers.PyLower(def), -1)...)
			seenWords := map[string]bool{}
			for _, w := range words {
				if len(w) < 3 || taxonomyStopwords[w] {
					continue
				}
				if seenWords[w] {
					continue
				}
				seenWords[w] = true
				if vocab[w] == nil {
					vocab[w] = map[string]TagHomeInfo{}
				}
				if _, exists := vocab[w][home]; !exists {
					vocab[w][home] = TagHomeInfo{CatName: cat.Name, SubName: sub}
				}
			}
		}
	}
	return vocab
}

// TagHome mirrors _tag_home: the set of homes a folded tag names in
// the vocabulary — exact match, plural-tolerant both ways. nil when
// the tag names nothing (an unknown word is no signal at all, never a
// veto). The homes ride in the vocabulary's insertion order.
func TagHome(tagFolded string,
	vocab map[string]map[string]TagHomeInfo) []string {
	if tagFolded == "" {
		return nil
	}
	if entry, ok := vocab[tagFolded]; ok {
		return sortedVocabHomes(entry)
	}
	plural := ""
	if !strings.HasSuffix(tagFolded, "s") {
		plural = tagFolded + "s"
	} else {
		plural = tagFolded[:len(tagFolded)-1]
	}
	if plural != "" {
		if entry, ok := vocab[plural]; ok {
			return sortedVocabHomes(entry)
		}
	}
	return nil
}

// sortedVocabHomes reads a vocabulary entry's home keys — sorted for
// determinism (the reference returns a set; the pass intersects sets
// and finally requires len==1, so order never changes the verdict).
func sortedVocabHomes(entry map[string]TagHomeInfo) []string {
	homes := make([]string, 0, len(entry))
	for h := range entry {
		homes = append(homes, h)
	}
	sort.Strings(homes)
	return homes
}

// TaxonomyTagMoves mirrors _taxonomy_tag_moves (v0.71.0 — THE
// TAXONOMY PASS): a candidate whose tags hit the taxonomy's own
// subcategory vocabulary files into that subcategory — when EVERY
// matched tag agrees on ONE home and that home exists on the LIVE
// tree. Conflicts stay on the LLM's plate.
func TaxonomyTagMoves(candidates []*NoteRow, vaultPath string, log LogFn,
	config map[string]any) ([]RawMove, []string) {
	var moves []RawMove
	var extraFolders []string
	func() {
		defer func() { recover() }() // the reference's outer try/except
		if vaultPath == "" || len(candidates) == 0 {
			return
		}
		vocab := TaxonomyVocabulary(config)
		if len(vocab) == 0 {
			return
		}
		foldedKnown := map[string]*string{}
		for _, cand := range candidates {
			var tags []string
			for _, t := range cand.Tags {
				if helpers.Strip(t) != "" {
					tags = append(tags, t)
				}
			}
			if len(tags) == 0 {
				continue // an untagged note is the LLM's to think
			}
			var agreed map[string]bool
			var matched []string
			for _, tag := range tags {
				homes := TagHome(foldTag(tag), vocab)
				if homes == nil {
					continue // names nothing — no signal, no veto
				}
				matched = append(matched, tag)
				if agreed == nil {
					agreed = map[string]bool{}
					for _, h := range homes {
						agreed[h] = true
					}
				} else {
					next := map[string]bool{}
					for _, h := range homes {
						if agreed[h] {
							next[h] = true
						}
					}
					agreed = next
				}
				if len(agreed) == 0 {
					break // already conflicting — stop reading
				}
			}
			if len(matched) == 0 || len(agreed) != 1 {
				continue // no signal, or a conflict — the LLM's call
			}
			home := ""
			for h := range agreed {
				home = h
			}
			// the taxonomy's own folder spelling, verified live
			dest := ResolveDiskFolder(vaultPath, home, foldedKnown)
			if dest == nil {
				continue // the library lacks the folder — a decision
			}
			if foldPath(*dest) == foldPath(cand.Folder) {
				continue // already there — FOLDED (v0.82.0)
			}
			note := cand.File
			if note == "" {
				note = cand.Rel
			}
			shown := matched
			if len(shown) > 4 {
				shown = shown[:4]
			}
			ellipsis := ""
			if len(matched) > 4 {
				ellipsis = ", …"
			}
			moves = append(moves, RawMove{
				Note: note, To: *dest,
				Reason: "the note's tags (" + strings.Join(shown, ", ") +
					ellipsis + ") match the taxonomy's own vocabulary for " +
					*dest,
			})
			if !inList(*dest, extraFolders) {
				extraFolders = append(extraFolders, *dest)
			}
		}
		if len(moves) > 0 {
			log(fmt.Sprintf("📐 %d candidate note(s) carry tags that name "+
				"one taxonomy subcategory — filed by the regex brain "+
				"from the owner's own category map, no LLM asked",
				len(moves)), "info")
		}
	}()
	return moves, extraFolders
}
