// noteparse_test.go — T-49: the ported selections of the reference's
// TestParser (test_directorysync.py) — the five real-note variants
// from UI spec §13.2 plus the loose-heading, banish-tag,
// no-frontmatter, and featured doors.
package directorysync

import (
	"strings"
	"testing"
)

func TestParser_LongNoteParsesEverySection(t *testing.T) {
	p := ParseNote(longNote)
	if p == nil {
		t.Fatal("parse_note returned nil")
	}
	if p.Name != "Coolors" {
		t.Fatalf("name: %q", p.Name)
	}
	if p.Summary != "The super-fast color scheme generator for designers." {
		t.Fatalf("summary: %q", p.Summary)
	}
	if len(p.Sections.WhatItDoes) != 4 {
		t.Fatalf("what_it_does: %d", len(p.Sections.WhatItDoes))
	}
	if !strings.HasPrefix(p.Sections.BestUsedFor, "Use it when") {
		t.Fatalf("best_used_for: %q", p.Sections.BestUsedFor)
	}
	if len(p.Sections.PracticalDetails) != 3 {
		t.Fatalf("practical_details: %d (the sign-up line became a chip, not a bullet)",
			len(p.Sections.PracticalDetails))
	}
	var export bool
	for _, d := range p.Sections.PracticalDetails {
		if strings.HasPrefix(d, "Export formats") {
			export = true
		}
	}
	if !export {
		t.Fatalf("practical_details: %v", p.Sections.PracticalDetails)
	}
	if p.SignupRequired != "no" {
		t.Fatalf("signup_required: %q", p.SignupRequired)
	}
	if p.Pricing != "freemium" {
		t.Fatalf("pricing: %q", p.Pricing)
	}
	if len(p.Sections.WatchOut) != 1 {
		t.Fatalf("watch_out: %d", len(p.Sections.WatchOut))
	}
	if p.Sections.SimilarTools != "Adobe Color, Khroma, Paletton" {
		t.Fatalf("similar_tools: %q", p.Sections.SimilarTools)
	}
	if len(p.Tags) != 2 || p.Tags[0] != "color-tool" || p.Tags[1] != "palette" {
		t.Fatalf("tags: %v", p.Tags)
	}
	if !p.Structured {
		t.Fatal("structured")
	}
	if p.DetailMarkdown != "" {
		t.Fatalf("detail_markdown: %q", p.DetailMarkdown)
	}
	if p.AddedAt != "2026-09-28" {
		t.Fatalf("added_at: %q", p.AddedAt)
	}
}

func TestParser_ShortNoteParses(t *testing.T) {
	p := ParseNote(shortNote)
	if p.Name != "Fast.com" {
		t.Fatalf("name: %q", p.Name)
	}
	if p.Summary != "Netflix's simple internet speed test." {
		t.Fatalf("summary: %q", p.Summary)
	}
	if len(p.Sections.WhatItDoes) != 0 {
		t.Fatalf("what_it_does: %v", p.Sections.WhatItDoes)
	}
	if p.Sections.BestUsedFor != "When you want one number, right now." {
		t.Fatalf("best_used_for: %q", p.Sections.BestUsedFor)
	}
	if !p.Structured {
		t.Fatal("structured")
	}
}

func TestParser_MissingSectionsHiddenAndEnumsKept(t *testing.T) {
	p := ParseNote(missingNote)
	if len(p.Sections.WhatItDoes) != 0 || len(p.Sections.WatchOut) != 0 {
		t.Fatalf("sections: %+v", p.Sections)
	}
	if p.SignupRequired != "yes" {
		t.Fatalf("login_required frontmatter fills the chip: %q", p.SignupRequired)
	}
	if p.Pricing != "freemium" {
		t.Fatalf("pricing: %q", p.Pricing)
	}
	if p.Sections.SimilarTools != "—" {
		t.Fatalf("similar_tools: %q", p.Sections.SimilarTools)
	}
}

func TestParser_EmptyBulletDropped(t *testing.T) {
	p := ParseNote(emptyBulletNote)
	if len(p.Sections.WatchOut) != 1 || p.Sections.WatchOut[0] != "The free tier is limited" {
		t.Fatalf("watch_out: %v", p.Sections.WatchOut)
	}
}

func TestParser_StrayRuleLineNeverAppears(t *testing.T) {
	p := ParseNote(strayNote)
	if p.Name != "Icons8" {
		t.Fatalf("name: %q", p.Name)
	}
	if p.Summary != "Icons, illustrations and photos with consistent styles." {
		t.Fatalf("summary: %q", p.Summary)
	}
	if len(p.Sections.WhatItDoes) != 2 {
		t.Fatalf("what_it_does: %v", p.Sections.WhatItDoes)
	}
	blob := pyReprValue(p.ToMap())
	if strings.Contains(blob, "Descript") {
		t.Fatalf("the stray artifact survived: %s", blob)
	}
}

func TestParser_UnstructuredNoteKeepsRawBody(t *testing.T) {
	p := ParseNote(unstructuredNote)
	if p.Structured {
		t.Fatal("structured")
	}
	if !strings.Contains(p.DetailMarkdown, "This is just prose") {
		t.Fatalf("detail_markdown: %q", p.DetailMarkdown)
	}
	if strings.Contains(p.DetailMarkdown, "[!info]") {
		t.Fatalf("the banner survived: %q", p.DetailMarkdown)
	}
	if strings.Contains(p.DetailMarkdown, "Source:") {
		t.Fatalf("the footer survived: %q", p.DetailMarkdown)
	}
}

func TestParser_LooseHeadingsMatch(t *testing.T) {
	loose := refNote(fmFields(map[string]string{"source": "https://loose.example/"}),
		"# Loose\n\n"+
			"> **TL;DR:** Loose labels.\n\n"+
			"## **What it does**:\n"+
			"- one bullet\n\n"+
			"## BEST USED FOR\n"+
			"Loose paragraph.\n"+
			refFooter("https://loose.example/"))
	p := ParseNote(loose)
	if len(p.Sections.WhatItDoes) != 1 || p.Sections.WhatItDoes[0] != "one bullet" {
		t.Fatalf("what_it_does: %v", p.Sections.WhatItDoes)
	}
	if p.Sections.BestUsedFor != "Loose paragraph." {
		t.Fatalf("best_used_for: %q", p.Sections.BestUsedFor)
	}
}

func TestParser_BanishTagsNeverRidePublic(t *testing.T) {
	tagged := refNote(fmFields(map[string]string{
		"tags":   "[ui, delete]",
		"source": "https://t.example/",
	}), "# T\n\n> **TL;DR:** x\n"+refFooter("https://t.example/"))
	p := ParseNote(tagged)
	if len(p.Tags) != 1 || p.Tags[0] != "ui" {
		t.Fatalf("tags: %v", p.Tags)
	}
}

func TestParser_NoFrontmatterRejected(t *testing.T) {
	if ParseNote("just some prose") != nil {
		t.Fatal("prose parsed")
	}
	if ParseNote("") != nil {
		t.Fatal("empty parsed")
	}
}

func TestParser_FeaturedFromFrontmatter(t *testing.T) {
	p := ParseNote(refNote(
		fmFields(map[string]string{"featured": "true", "source": "https://f.example/"}),
		"# F\n\n> **TL;DR:** pick me\n"+refFooter("https://f.example/")))
	if !p.Featured {
		t.Fatal("featured")
	}
}
