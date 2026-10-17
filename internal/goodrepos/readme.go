// readme.go — _build_readme: the emoji-rich categorized README with
// counts and clickable anchor lines (the directory's front door).
package goodrepos

import (
	"sort"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// buildReadme mirrors _build_readme. cats is the scan result; the
// today stamp rides the clock seam.
func (s *GoodRepos) buildReadme(cats []*CategoryNotes) string {
	today := clockNow().UTC().Format("2006-01-02")
	entries := 0
	for _, c := range cats {
		entries += len(c.Notes)
	}

	// group category dirs: top component -> sub path ("") -> notes
	tops := map[string]map[string][]*Note{}
	var rels []string
	for _, c := range cats {
		rels = append(rels, c.Rel)
	}
	sort.Strings(rels)
	for _, rel := range rels {
		var notes []*Note
		for _, c := range cats {
			if c.Rel == rel {
				notes = c.Notes
				break
			}
		}
		parts := strings.Split(rel, "/")
		top := parts[0]
		sub := strings.Join(parts[1:], "/")
		if tops[top] == nil {
			tops[top] = map[string][]*Note{}
		}
		tops[top][sub] = append(tops[top][sub], notes...)
	}
	topOrder := sortedTops(tops)

	subDisplay := func(sub string) string {
		var comps []string
		for _, c := range strings.Split(sub, "/") {
			if c != "" {
				comps = append(comps, DisplayName(c))
			}
		}
		return strings.Join(comps, " / ")
	}
	subOrder := func(top string) []string {
		var subs []string
		for s := range tops[top] {
			if s != "" {
				subs = append(subs, s)
			}
		}
		sort.SliceStable(subs, func(i, j int) bool {
			ki, kj := subDisplay(subs[i]), subDisplay(subs[j])
			if ki != kj {
				return ki < kj
			}
			return subs[i] < subs[j]
		})
		return subs
	}

	sortKeyLess := func(a, b *Note) bool { // (-stars, display.lower(), rel_path)
		sa, sb := -1, -1
		if a.Stars != nil {
			sa = *a.Stars
		}
		if b.Stars != nil {
			sb = *b.Stars
		}
		if sa != sb {
			return sa > sb
		}
		da, db := helpers.PyLower(a.Display), helpers.PyLower(b.Display)
		if da != db {
			return da < db
		}
		return a.RelPath < b.RelPath
	}
	sortNotes := func(notes []*Note) {
		sort.SliceStable(notes, func(i, j int) bool { return sortKeyLess(notes[i], notes[j]) })
	}

	plural := func(n int, one, many string) string {
		if n == 1 {
			return one
		}
		return many
	}

	out := []string{
		"# 🗂️ Good Repos",
		"",
		"> ✨ A curated directory of quality GitHub repositories — auto-maintained by",
		"> [GitCurator](https://github.com/assadigit/GitCurator).",
		"",
		"📊 **" + strconv.Itoa(entries) + " " + plural(entries, "repo", "repos") + "** · " +
			"🗂️ **" + strconv.Itoa(len(cats)) + " " + plural(len(cats), "category", "categories") + "** · " +
			"🕰️ Updated " + today,
		"",
		"## 📇 Contents",
		"",
	}
	for _, top := range topOrder {
		heading := topHeading(top)
		out = append(out, "- ["+heading+"](#"+GitHubAnchor(heading)+")")
		for _, sub := range subOrder(top) {
			count := len(tops[top][sub])
			sheading := subHeading(sub, count)
			out = append(out, "  - ["+EmojiFor(strings.Split(sub, "/")...)+" "+subDisplay(sub)+"]"+
				"(#"+GitHubAnchor(sheading)+")")
		}
	}
	out = append(out, "")

	for _, top := range topOrder {
		heading := topHeading(top)
		out = append(out, "## "+heading, "",
			"_Curated repositories for the "+introName(DisplayName(top))+"._", "")
		direct := tops[top][""]
		if len(direct) > 0 {
			sorted := append([]*Note(nil), direct...)
			sortNotes(sorted)
			for _, n := range sorted {
				out = append(out, entryLine(n))
			}
			out = append(out, "")
		}
		for _, sub := range subOrder(top) {
			notes := tops[top][sub]
			out = append(out, "### "+subHeading(sub, len(notes)), "")
			sorted := append([]*Note(nil), notes...)
			sortNotes(sorted)
			for _, n := range sorted {
				out = append(out, entryLine(n))
			}
			out = append(out, "")
		}
	}

	out = append(out,
		"---",
		"",
		"_🤖 This directory is regenerated automatically after every GitCurator "+
			"curation run. Browsing the category folders gives you the full curated "+
			"notes._",
		"")
	return strings.Join(out, "\n")
}

// topHeading mirrors _top_heading.
func topHeading(top string) string {
	return EmojiFor(top) + " " + DisplayName(top)
}

// subHeading mirrors _sub_heading.
func subHeading(sub string, count int) string {
	var components []string
	for _, c := range strings.Split(sub, "/") {
		if c != "" {
			components = append(components, c)
		}
	}
	var disps []string
	for _, c := range components {
		disps = append(disps, DisplayName(c))
	}
	return EmojiFor(components...) + " " + strings.Join(disps, " / ") + " (" + strconv.Itoa(count) + ")"
}

// entryLine mirrors _entry_line.
func entryLine(note *Note) string {
	var line string
	if note.URL != "" {
		line = "- 📦 [" + note.Display + "](" + note.URL + ")"
	} else {
		line = "- 📦 " + note.Display
	}
	if note.Tldr != "" {
		line += " — " + note.Tldr
	}
	if note.Stars != nil {
		line += " · ⭐ " + FormatStars(note.Stars)
	}
	if note.Language != "" {
		line += " · 🔧 " + note.Language
	}
	if len(note.Tags) > 0 {
		var tagged []string
		for _, t := range note.Tags {
			tagged = append(tagged, "`"+t+"`")
		}
		line += " · " + strings.Join(tagged, " ")
	}
	return line
}

// sortedTops mirrors sorted(tops, key=lambda t: (display_name(t).lower(), t))
// — the top categories alphabetical by DISPLAY name.
func sortedTops(tops map[string]map[string][]*Note) []string {
	keys := make([]string, 0, len(tops))
	for k := range tops {
		keys = append(keys, k)
	}
	sort.SliceStable(keys, func(i, j int) bool {
		ki, kj := helpers.PyLower(DisplayName(keys[i])), helpers.PyLower(DisplayName(keys[j]))
		if ki != kj {
			return ki < kj
		}
		return keys[i] < keys[j]
	})
	return keys
}
