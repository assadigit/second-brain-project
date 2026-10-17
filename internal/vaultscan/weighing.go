package vaultscan

import (
	"fmt"
	"math/bits"
	"regexp"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// WeighedHome mirrors weigh_candidate_homes' rows:
// {'folder', 'weight', 'reasons': [str, ...]}.
type WeighedHome struct {
	Folder  string
	Weight  float64
	Reasons []string
}

var wordTokenRE = regexp.MustCompile(`[a-z0-9]+`)

// NoteWordTokens mirrors _note_word_tokens: a free-text's weighing
// identity — lowercase alphanumeric runs of 4+ characters, grammar
// words dropped. Pure; junk-tolerant.
func NoteWordTokens(text string) map[string]bool {
	out := map[string]bool{}
	for _, w := range wordTokenRE.FindAllString(helpers.PyLower(text), -1) {
		if len(w) >= 4 && !taxonomyStopwords[w] {
			out[w] = true
		}
	}
	return out
}

// FolderProfile mirrors _folder_profiles' per-folder census: the
// folded tags its settled notes carry (with counts, insertion order
// kept), the title-word tokens, the resident count, and the repos
// languages census.
type FolderProfile struct {
	Tags        map[string]int
	TagOrder    []string
	TitleTokens map[string]bool
	NoteCount   int
	Languages   map[string]int
	LangOrder   []string
}

// FolderProfiles mirrors _folder_profiles (v0.74.0 — THE WEIGHED
// EVIDENCE's census): what the EXISTING notes on every folder actually
// hold. The teaching law is the sibling pass's own: a candidate
// teaches nothing, hand notes and protected/core/locked/failed rows
// teach nothing, the parking lot and bare roots are not homes. Pure;
// never raises.
func FolderProfiles(inv *Inventory, excludePaths map[string]bool) map[string]*FolderProfile {
	profiles := map[string]*FolderProfile{}
	for _, n := range inv.Notes {
		if excludePaths[n.Path] {
			continue // a candidate teaches nothing
		}
		if !n.AppOwned || n.Protected || n.Core || n.Locked {
			continue
		}
		fs := n.FetchStatus
		if fs == "" {
			fs = "full"
		}
		if fs == "failed" {
			continue // a failed placeholder is not a resident
		}
		folder := n.Folder
		if !strings.Contains(folder, "/") {
			continue // the depth law: no bare roots teach
		}
		folded := foldPath(folder)
		if folded == "" || strings.HasPrefix(strings.ToLower(folded), "uncategorized/") {
			continue // the parking lot is nobody's home
		}
		per := profiles[folder]
		if per == nil {
			per = &FolderProfile{
				Tags:        map[string]int{},
				TitleTokens: map[string]bool{},
				Languages:   map[string]int{},
			}
			profiles[folder] = per
		}
		per.NoteCount++
		for _, t := range n.Tags {
			ft := foldTag(t)
			if ft == "" {
				continue
			}
			if _, seen := per.Tags[ft]; !seen {
				per.TagOrder = append(per.TagOrder, ft)
			}
			per.Tags[ft]++
		}
		for w := range NoteWordTokens(n.Title) {
			per.TitleTokens[w] = true
		}
		// v0.83.0 — the repos census: the folder's own languages
		if lang := helpers.Strip(n.PrimaryLang); lang != "" {
			if _, seen := per.Languages[lang]; !seen {
				per.LangOrder = append(per.LangOrder, lang)
			}
			per.Languages[lang]++
		}
	}
	return profiles
}

// WeighingFolderSet mirrors _weighing_folder_set: the folders THE
// WEIGHED EVIDENCE may name — every depth-legal folder the inventory
// saw, minus the parking lot. Sorted for determinism.
func WeighingFolderSet(inv *Inventory) []string {
	seen := map[string]bool{}
	var out []string
	for _, f := range inv.Folders {
		rel := f.Rel
		if !strings.Contains(rel, "/") {
			continue // the depth law: never a bare root
		}
		folded := foldPath(rel)
		if folded == "" || strings.HasPrefix(strings.ToLower(folded), "uncategorized/") {
			continue // the parking lot is nobody's home
		}
		if !seen[rel] {
			seen[rel] = true
			out = append(out, rel)
		}
	}
	sort.Strings(out)
	return out
}

// WeighOptions carries weigh_candidate_homes' optional parameter set
// (profiles, domain_homes, vocab, folder_counts) — the scale's shared
// teaching material, precomputed once per plan.
type WeighOptions struct {
	Profiles     map[string]*FolderProfile
	DomainHomes  *DomainHomes
	Vocab        map[string]map[string]TagHomeInfo
	FolderCounts map[string]int
}

// WeighCandidateHomes mirrors weigh_candidate_homes (v0.74.0 — THE
// WEIGHED EVIDENCE's scale, pure, deterministic, never raises): for
// ONE move candidate, the ranked shortlist of up to WeightTopK
// candidate homes, each with a 0.0-1.0 weight and the reasons that
// earned it — settled siblings of the domain voting their share (repos:
// the LANGUAGE agreement), tags shared with the folder's existing
// notes, the taxonomy's vocabulary, title words in common, the
// folder's own name, the declared home's echoes (repos: the declared
// category's own folder), the popularity prior (only when w>0), and
// the too-broad penalty. round(min(w,1.0), 2) — banker's rounding.
func WeighCandidateHomes(cand *NoteRow, inv *Inventory, opts *WeighOptions,
	tooBroad int, kind string) []WeighedHome {
	var out []WeighedHome
	func() {
		defer func() { recover() }() // the reference's outer try/except
		if cand == nil {
			return
		}
		if opts == nil {
			opts = &WeighOptions{}
		}
		profiles := opts.Profiles
		if profiles == nil {
			profiles = map[string]*FolderProfile{}
		}
		domainHomes := opts.DomainHomes
		folderCounts := opts.FolderCounts
		if folderCounts == nil {
			folderCounts = map[string]int{}
			for _, f := range inv.Folders {
				folderCounts[f.Rel] = f.NoteCount
			}
		}
		var tagList []string
		for _, t := range cand.Tags {
			if helpers.Strip(t) != "" {
				tagList = append(tagList, t)
			}
		}
		tagFolded := map[string]bool{}
		for _, t := range tagList {
			if f := foldTag(t); f != "" {
				tagFolded[f] = true
			}
		}
		titleWords := NoteWordTokens(cand.Title)
		noteWords := map[string]bool{}
		for w := range titleWords {
			noteWords[w] = true
		}
		for f := range tagFolded {
			noteWords[f] = true
		}
		domain := strings.ToLower(helpers.Strip(cand.Domain))
		cat := helpers.Strip(cand.Category)
		sub := helpers.Strip(cand.Subcategory)
		repos := isReposKind(kind)
		candLang := helpers.Strip(cand.PrimaryLang)
		catFolder := ""
		if repos && cat != "" {
			catFolder = reposCategoryFolders()[cat]
		}
		var domHomes map[string]int
		domTotal := 0
		if domain != "" && domainHomes != nil {
			if per, ok := domainHomes.Homes[domain]; ok {
				domHomes = per
				for _, n := range per {
					domTotal += n
				}
			}
		}
		type scored struct {
			folder string
			w      float64
		}
		var scores []scored
		reasons := map[string][]string{}
		for _, folder := range WeighingFolderSet(inv) {
			w := 0.0
			var why []string
			per := profiles[folder]
			var folderTags map[string]int
			if per != nil {
				folderTags = per.Tags
			}
			// 1) settled siblings of the domain vote their share (repos:
			//    the LANGUAGE agreement — the folder's residents voting
			//    on their tongue)
			if repos {
				var langs map[string]int
				langTotal := 0
				if per != nil {
					langs = per.Languages
					for _, n := range langs {
						langTotal += n
					}
				}
				if candLang != "" && langTotal > 0 {
					votes := langs[candLang]
					if votes > 0 {
						share := float64(votes) / float64(langTotal)
						w += 0.30 * share
						if share >= 0.999 {
							w += 0.10 // unanimous — the language's word
							why = append(why, fmt.Sprintf(
								"all %d settled repo(s) there are written in %s",
								votes, candLang))
						} else {
							why = append(why, fmt.Sprintf(
								"%d of %d settled repo(s) there are written in %s",
								votes, langTotal, candLang))
						}
					}
				}
			} else if len(domHomes) > 0 && domTotal > 0 {
				votes := domHomes[foldPath(folder)]
				if votes > 0 {
					share := float64(votes) / float64(domTotal)
					w += 0.30 * share
					if share >= 0.999 {
						w += 0.10 // unanimous — the domain's word
						why = append(why, fmt.Sprintf(
							"%d settled sibling(s) of %s live there (unanimous)",
							votes, domain))
					} else {
						why = append(why, fmt.Sprintf(
							"%d of %d settled sibling(s) of %s vote there",
							votes, domTotal, domain))
					}
				}
			}
			// 2) tags shared with the folder's existing notes
			var shared []string
			for t := range tagFolded {
				if _, ok := folderTags[t]; ok {
					shared = append(shared, t)
				}
			}
			sort.Strings(shared)
			if len(shared) > 0 {
				w += minFloat(0.06*float64(len(shared)), 0.18)
				shown := shared
				if len(shown) > 4 {
					shown = shown[:4]
				}
				why = append(why, "shared tags: "+strings.Join(shown, ", "))
			}
			// 3) the taxonomy's own vocabulary names this home
			if len(opts.Vocab) > 0 {
				var hit string
				var sortedTagFolded []string
				for t := range tagFolded {
					sortedTagFolded = append(sortedTagFolded, t)
				}
				sort.Strings(sortedTagFolded)
				for _, tf := range sortedTagFolded {
					named := TagHome(tf, opts.Vocab)
					if len(named) == 0 {
						continue
					}
					match := false
					for _, x := range named {
						if foldPath(folder) == foldPath(x) {
							match = true
							break
						}
					}
					if match {
						hit = tf
						break
					}
				}
				if hit != "" {
					w += 0.10
					why = append(why, "taxonomy word: "+hit)
				}
			}
			// 4) title words the folder's notes' titles speak too
			var overlap []string
			if per != nil {
				for w := range titleWords {
					if per.TitleTokens[w] {
						overlap = append(overlap, w)
					}
				}
			}
			sort.Strings(overlap)
			if len(overlap) > 0 {
				w += minFloat(0.05*float64(len(overlap)), 0.10)
				shown := overlap
				if len(shown) > 4 {
					shown = shown[:4]
				}
				why = append(why, "title words shared: "+strings.Join(shown, ", "))
			}
			// 5) the folder's own name matches the note's words
			leafSeg := lastSegment(folder)
			leaf := NoteWordTokens(leafSeg)
			leafHit := false
			for w := range leaf {
				if noteWords[w] {
					leafHit = true
					break
				}
			}
			if len(leaf) > 0 && leafHit {
				w += 0.06
				why = append(why, "folder name matches the note's words")
			}
			// 6) the declared home's partial echoes (repos: the
			//    declared category's OWN folder, mapped through
			//    CATEGORY_FOLDERS)
			if repos {
				if catFolder != "" && foldPath(folder) == foldPath(catFolder) {
					w += 0.25
					why = append(why, "its declared category's own folder ("+cat+")")
				}
			} else {
				if cat != "" {
					if foldSegment(firstSegment(folder)) == foldSegment(cat) {
						w += 0.12
						why = append(why, "under its declared category ("+cat+")")
					}
				}
				if sub != "" {
					if foldSegment(leafSeg) == foldSegment(sub) {
						w += 0.15
						why = append(why, "its declared subfolder ("+sub+")")
					}
				}
			}
			// the prior scales evidence — it is not evidence: a folder
			// with no affinity to the note never rides on popularity
			// alone.
			if w <= 0 {
				continue
			}
			// 7) popularity prior — mild, log-scaled
			count := folderCounts[folder]
			if per != nil && count == 0 {
				count = per.NoteCount
			}
			if count > 0 {
				w += minFloat(0.05,
					float64(bits.Len(uint(count+1)))/24.0)
			}
			// 8) the too-broad penalty
			if count > tooBroad {
				w -= 0.10
				why = append(why,
					"folder already TOO BROAD (a poor destination)")
			}
			if w > 0 {
				scores = append(scores, scored{folder: folder, w: w})
				reasons[folder] = why
			}
		}
		sort.SliceStable(scores, func(i, j int) bool {
			if scores[i].w != scores[j].w {
				return scores[i].w > scores[j].w
			}
			return scores[i].folder < scores[j].folder
		})
		top := scores
		if len(top) > WeightTopK {
			top = top[:WeightTopK]
		}
		for _, s := range top {
			out = append(out, WeighedHome{
				Folder:  s.folder,
				Weight:  helpers.RoundHalfEven(minFloat(s.w, 1.0), 2),
				Reasons: reasons[s.folder],
			})
		}
	}()
	return out
}

// AttachWeighedHomes mirrors attach_weighed_homes (v0.74.0 — THE
// WEIGHED EVIDENCE's wiring): decorate every LLM-bound candidate note
// IN PLACE with its weighed shortlist (Homes) and the CLOSE CALL
// verdict (HomesClose — the top two sit within WeightTieMargin).
// Returns how many candidates carry a shortlist. Pure; never raises.
func AttachWeighedHomes(candidates []*NoteRow, inv *Inventory,
	excludePaths map[string]bool, config map[string]any,
	tooBroad int, kind string) int {
	withHomes := 0
	if len(candidates) == 0 {
		return 0
	}
	exclude := map[string]bool{}
	for p := range excludePaths {
		exclude[p] = true
	}
	profiles := FolderProfiles(inv, exclude)
	var domainHomes *DomainHomes
	vocab := map[string]map[string]TagHomeInfo{}
	if isReposKind(kind) {
		domainHomes = &DomainHomes{
			Homes: map[string]map[string]int{}, Spelling: map[string]string{},
			PerOrder: map[string][]string{},
		}
	} else {
		domainHomes = SettledDomainHomes(inv, exclude)
		vocab = TaxonomyVocabulary(config)
	}
	folderCounts := map[string]int{}
	for _, f := range inv.Folders {
		folderCounts[f.Rel] = f.NoteCount
	}
	opts := &WeighOptions{
		Profiles:     profiles,
		DomainHomes:  domainHomes,
		Vocab:        vocab,
		FolderCounts: folderCounts,
	}
	for _, cand := range candidates {
		homes := WeighCandidateHomes(cand, inv, opts, tooBroad, kind)
		cand.Homes = homes
		cand.HomesClose = len(homes) >= 2 &&
			homes[0].Weight-homes[1].Weight <= WeightTieMargin
		if len(homes) > 0 {
			withHomes++
		}
	}
	return withHomes
}

// HomesLine mirrors _homes_line: the digest's weighed-evidence cell —
// the ranked shortlist in the LLM's own reading shape, reasons
// attached, the CLOSE CALL marker riding when the arithmetic tied.
// Pure; junk-tolerant (no homes — no cell).
func HomesLine(cand *NoteRow) string {
	if cand == nil || len(cand.Homes) == 0 {
		return ""
	}
	var bits []string
	for i, h := range cand.Homes {
		why := h.Reasons
		if len(why) > 3 {
			why = why[:3]
		}
		cell := fmt.Sprintf("%d) %s %.2f", i+1, h.Folder, h.Weight)
		if len(why) > 0 {
			cell += " (" + strings.Join(why, "; ") + ")"
		}
		bits = append(bits, cell)
	}
	line := "homes: " + strings.Join(bits, " · ")
	if cand.HomesClose {
		line += " — CLOSE CALL: the top two weigh nearly the same, " +
			"you break the tie on the note's own words"
	}
	return line
}

func minFloat(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func firstSegment(path string) string {
	if i := strings.IndexByte(path, '/'); i >= 0 {
		return path[:i]
	}
	return path
}

func lastSegment(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	return path
}
