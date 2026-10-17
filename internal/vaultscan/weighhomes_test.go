// weighhomes_test.go — T-45: the ported test_weighhomes selections
// (v0.74.0 THE WEIGHED EVIDENCE — the scale, the attachment, the
// alignment door, and the wiring). The prompt-contract and
// release-bookkeeping legs are the Python repo's own CI facts — not
// portable (the house precedent).
package vaultscan

import (
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/llmclient"
)

// --- TestTheWordTokens -------------------------------------------------

func TestGrammarAndFragmentsNeverVote(t *testing.T) {
	toks := NoteWordTokens("The AI and Color Palette Tool")
	if len(toks) != 3 || !toks["palette"] || !toks["color"] || !toks["tool"] {
		t.Fatalf("tokens %v", toks)
	}
}

func TestJunkNeverRaisesForTokens(t *testing.T) {
	if len(NoteWordTokens("")) != 0 {
		t.Fatal("empty voted")
	}
	if len(NoteWordTokens("!!! … —")) != 0 {
		t.Fatal("punctuation voted")
	}
}

// --- TestTheTeachingSet ------------------------------------------------

func weighInvRows(rows ...NoteRow) *Inventory {
	return &Inventory{Notes: rows}
}

func settledRow(path, folder, domain string, kw ...func(*NoteRow)) NoteRow {
	n := NoteRow{
		Path: path, AppOwned: true, Protected: false, Core: false,
		FetchStatus: "full", Folder: folder, Domain: domain,
	}
	for _, k := range kw {
		k(&n)
	}
	return n
}

func withRowTags(tags ...string) func(*NoteRow) {
	return func(n *NoteRow) { n.Tags = tags }
}

func withRowTitle(title string) func(*NoteRow) {
	return func(n *NoteRow) { n.Title = title }
}

func TestSettledNotesTeachTheirDomainHomes(t *testing.T) {
	inv := weighInvRows(
		settledRow("/v/a.md", "AI/Agents", "sketch.com"),
		settledRow("/v/b.md", "AI/Agents", "sketch.com"))
	homes := SettledDomainHomes(inv, nil)
	if len(homes.Homes) != 1 {
		t.Fatalf("homes %+v", homes.Homes)
	}
	per := homes.Homes["sketch.com"]
	if len(per) != 1 || per["ai/agents"] != 2 {
		t.Fatalf("per %v", per)
	}
	if homes.Spelling["sketch.com\x00ai/agents"] != "AI/Agents" {
		t.Fatalf("spelling %v", homes.Spelling)
	}
}

func TestACandidateTeachesNothing(t *testing.T) {
	inv := weighInvRows(settledRow("/v/a.md", "AI/Agents", "sketch.com"))
	homes := SettledDomainHomes(inv, map[string]bool{"/v/a.md": true})
	if len(homes.Homes) != 0 {
		t.Fatalf("homes %+v", homes.Homes)
	}
}

func TestTheUnfitNeverTeach(t *testing.T) {
	rows := []NoteRow{
		settledRow("/v/h.md", "AI/Agents", "hand.com",
			func(n *NoteRow) { n.AppOwned = false }),
		settledRow("/v/p.md", "_review/X", "prot.com",
			func(n *NoteRow) { n.Protected = true }),
		settledRow("/v/c.md", "AI/Agents", "core.com",
			func(n *NoteRow) { n.Core = true }),
		settledRow("/v/f.md", "AI/Agents", "fail.com",
			func(n *NoteRow) { n.FetchStatus = "failed" }),
		settledRow("/v/r.md", "AI", "root.com"),
		settledRow("/v/u.md", "Uncategorized/Sub", "park.com"),
		settledRow("/v/n.md", "AI/Agents", ""),
	}
	homes := SettledDomainHomes(weighInvRows(rows...), nil)
	if len(homes.Homes) != 0 || len(homes.Spelling) != 0 {
		t.Fatalf("homes %+v spelling %v", homes.Homes, homes.Spelling)
	}
}

// --- TestTheFolderCensus ----------------------------------------------

func TestTheExistingNotesOnAFolderAreItsCensus(t *testing.T) {
	inv := weighInvRows(
		settledRow("/v/a.md", "AI/Agents", "x.com",
			withRowTags("agents", "llm"), withRowTitle("Agent Chat Tools")),
		settledRow("/v/b.md", "AI/Agents", "x.com",
			withRowTags("agents"), withRowTitle("Autonomous Agents List")))
	prof := FolderProfiles(inv, nil)
	per := prof["AI/Agents"]
	if per == nil || per.NoteCount != 2 {
		t.Fatalf("profile %+v", per)
	}
	if per.Tags["agents"] != 2 || per.Tags["llm"] != 1 {
		t.Fatalf("tags %v", per.Tags)
	}
	for _, w := range []string{"agents", "autonomous", "tools"} {
		if !per.TitleTokens[w] {
			t.Fatalf("title token %q missing", w)
		}
	}
}

func TestALockedNoteNeverTeachesTheCensus(t *testing.T) {
	inv := weighInvRows(settledRow("/v/a.md", "AI/Agents", "x.com",
		withRowTags("agents"), withRowTitle("Agents"),
		func(n *NoteRow) { n.Locked = true }))
	if prof := FolderProfiles(inv, nil); len(prof) != 0 {
		t.Fatalf("profiles %v", prof)
	}
}

func TestBareRootsAndTheParkingLotHaveNoCensus(t *testing.T) {
	inv := weighInvRows(
		settledRow("/v/a.md", "AI", "x.com", withRowTags("agents"),
			withRowTitle("Agents")),
		settledRow("/v/b.md", "Uncategorized/X", "x.com",
			withRowTags("agents"), withRowTitle("Agents")))
	if prof := FolderProfiles(inv, nil); len(prof) != 0 {
		t.Fatalf("profiles %v", prof)
	}
}

// --- TestTheWeighingFolderSet -----------------------------------------

func TestDepthLawAndParkingLotOut(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI", NoteCount: 3},
		{Rel: "AI/Agents", NoteCount: 3},
		{Rel: "Design/Color_Tools", NoteCount: 2},
		{Rel: "Uncategorized/Sub", NoteCount: 1},
		{Rel: "", NoteCount: 1},
	}}
	got := WeighingFolderSet(inv)
	if len(got) != 2 || got[0] != "AI/Agents" ||
		got[1] != "Design/Color_Tools" {
		t.Fatalf("set %v", got)
	}
}

func TestSortedAndDeduped(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "B/X", NoteCount: 1},
		{Rel: "A/Z", NoteCount: 1},
		{Rel: "A/Z", NoteCount: 5},
	}}
	got := WeighingFolderSet(inv)
	if len(got) != 2 || got[0] != "A/Z" || got[1] != "B/X" {
		t.Fatalf("set %v", got)
	}
}

// --- TestTheScale ------------------------------------------------------

type scaleCase struct {
	inv     *Inventory
	profile map[string]*FolderProfile
	homes   *DomainHomes
}

func newScaleCase() *scaleCase {
	inv := &Inventory{
		Folders: []FolderCount{
			{Rel: "AI/Agents", NoteCount: 2},
			{Rel: "AI/Chatbots", NoteCount: 1},
			{Rel: "AI/Huge", NoteCount: 60},
			{Rel: "AI/Small", NoteCount: 2},
			{Rel: "Design/Color_Tools", NoteCount: 1},
			{Rel: "Design/Fonts", NoteCount: 0},
			{Rel: "AI", NoteCount: 3},
			{Rel: "Uncategorized/Sub", NoteCount: 1},
		},
		Notes: []NoteRow{
			settledRow("/v/1.md", "AI/Agents", "split.com",
				withRowTags("agents"), withRowTitle("Agents One")),
			settledRow("/v/2.md", "AI/Agents", "split.com",
				withRowTags("agents", "llm"), withRowTitle("Agents Two")),
			settledRow("/v/3.md", "AI/Chatbots", "split.com",
				withRowTitle("C3")),
			settledRow("/v/4.md", "AI/Agents", "one.com",
				withRowTitle("One")),
			settledRow("/v/5.md", "AI/Agents", "one.com",
				withRowTitle("Two2")),
			settledRow("/v/6.md", "Design/Color_Tools", "x.com",
				withRowTags("fonts"), withRowTitle("Color Palette Picker")),
			settledRow("/v/7.md", "AI/Huge", "x.com",
				withRowTags("agents"), withRowTitle("Agent Thing")),
			settledRow("/v/8.md", "AI/Small", "x.com",
				withRowTags("agents"), withRowTitle("Agent Other")),
		},
	}
	return &scaleCase{
		inv:     inv,
		profile: FolderProfiles(inv, nil),
		homes:   SettledDomainHomes(inv, nil),
	}
}

func (s *scaleCase) weigh(cand NoteRow,
	vocab map[string]map[string]TagHomeInfo) []WeighedHome {
	opts := &WeighOptions{Profiles: s.profile, DomainHomes: s.homes,
		Vocab: vocab}
	return WeighCandidateHomes(&cand, s.inv, opts, 40,
		VaultKindWebsites)
}

func TestNoEvidenceNoShortlist(t *testing.T) {
	s := newScaleCase()
	out := s.weigh(NoteRow{Title: "Unrelated", Domain: "else.com"}, nil)
	if len(out) != 0 {
		t.Fatalf("out %+v", out)
	}
}

func TestTagsSharedWithTheExistingNotesOutweighNoise(t *testing.T) {
	s := newScaleCase()
	out := s.weigh(NoteRow{Title: "Some Agent Site",
		Tags: []string{"agents", "llm"}, Domain: "new.com"}, nil)
	if len(out) == 0 || out[0].Folder != "AI/Agents" {
		t.Fatalf("out %+v", out)
	}
	joined := strings.Join(out[0].Reasons, " ")
	if !strings.Contains(joined, "shared tags: agents, llm") {
		t.Fatalf("reasons %q", joined)
	}
	for _, h := range out {
		if h.Folder == "Design/Color_Tools" {
			t.Fatal("the signal-free folder rode — popularity is not evidence")
		}
	}
}

func TestSettledSiblingsOfTheDomainVoteTheirShare(t *testing.T) {
	s := newScaleCase()
	out := s.weigh(NoteRow{Title: "Split Domain Note",
		Domain: "split.com"}, nil)
	if len(out) == 0 || out[0].Folder != "AI/Agents" {
		t.Fatalf("out %+v", out)
	}
	w := map[string]float64{}
	for _, h := range out {
		w[h.Folder] = h.Weight
	}
	if !(w["AI/Agents"] > w["AI/Chatbots"]) {
		t.Fatalf("weights %v", w)
	}
	if !(w["AI/Agents"] < 0.30) {
		t.Fatalf("AI/Agents %v not under 0.30", w["AI/Agents"])
	}
	reason := strings.Join(out[0].Reasons, " ")
	if !strings.Contains(reason, "2 of 3 settled sibling(s)") {
		t.Fatalf("reasons %q", reason)
	}
}

func TestAUnanimousDomainIsTheDomainsOwnWord(t *testing.T) {
	s := newScaleCase()
	out := s.weigh(NoteRow{Title: "New Note Of The Domain",
		Domain: "one.com"}, nil)
	if len(out) == 0 {
		t.Fatal("no shortlist")
	}
	if !(out[0].Weight >= 0.40) {
		t.Fatalf("weight %v < 0.40", out[0].Weight)
	}
	if !strings.Contains(strings.Join(out[0].Reasons, " "), "unanimous") {
		t.Fatalf("reasons %+v", out[0].Reasons)
	}
}

func TestTheTaxonomysOwnVocabularyEarnsItsWeight(t *testing.T) {
	s := newScaleCase()
	vocab := map[string]map[string]TagHomeInfo{
		"fonts": {"Design/Fonts": {CatName: "Design",
			SubName: "Fonts"}}}
	out := s.weigh(NoteRow{Title: "A Font Library",
		Tags: []string{"fonts"}, Domain: "f.com"}, vocab)
	if len(out) == 0 || out[0].Folder != "Design/Fonts" {
		t.Fatalf("out %+v", out)
	}
	if !strings.Contains(strings.Join(out[0].Reasons, " "),
		"taxonomy word: fonts") {
		t.Fatalf("reasons %+v", out[0].Reasons)
	}
}

func TestTitleWordsInCommonEarnTheirWeight(t *testing.T) {
	s := newScaleCase()
	out := s.weigh(NoteRow{Title: "Palette Generator", Domain: "p.com"},
		nil)
	if len(out) == 0 || out[0].Folder != "Design/Color_Tools" {
		t.Fatalf("out %+v", out)
	}
	if !strings.Contains(strings.Join(out[0].Reasons, " "),
		"title words shared: palette") {
		t.Fatalf("reasons %+v", out[0].Reasons)
	}
}

func TestTheFoldersOwnNameMatchesTheNotesWords(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI/Chatbots", NoteCount: 0}}}
	out := WeighCandidateHomes(&NoteRow{Title: "Chatbots Hub",
		Domain: "c.com"}, inv, nil, 40, VaultKindWebsites)
	if len(out) != 1 || out[0].Folder != "AI/Chatbots" ||
		out[0].Weight != 0.06 ||
		len(out[0].Reasons) != 1 ||
		out[0].Reasons[0] != "folder name matches the note's words" {
		t.Fatalf("out %+v", out)
	}
}

func TestTheDeclaredHomeEchoesEarnTheirWeights(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI/Chatbots", NoteCount: 0},
		{Rel: "Design/Chatbots", NoteCount: 0},
		{Rel: "AI/Agents", NoteCount: 0}}}
	out := WeighCandidateHomes(&NoteRow{Title: "Something",
		Domain: "s.com", Category: "AI", Subcategory: "Chatbots"},
		inv, nil, 40, VaultKindWebsites)
	w := map[string]float64{}
	for _, h := range out {
		w[h.Folder] = h.Weight
	}
	if w["AI/Chatbots"] != 0.27 || w["Design/Chatbots"] != 0.15 ||
		w["AI/Agents"] != 0.12 {
		t.Fatalf("weights %v", w)
	}
}

func TestTheTooBroadPenaltyBites(t *testing.T) {
	inv := &Inventory{
		Folders: []FolderCount{
			{Rel: "AI/Huge", NoteCount: 60},
			{Rel: "AI/Small", NoteCount: 2}},
		Notes: []NoteRow{
			settledRow("/v/7.md", "AI/Huge", "x.com",
				withRowTags("agents"), withRowTitle("Agent Thing")),
			settledRow("/v/8.md", "AI/Small", "x.com",
				withRowTags("agents"), withRowTitle("Agent Other"))}}
	opts := &WeighOptions{Profiles: FolderProfiles(inv, nil),
		DomainHomes: SettledDomainHomes(inv, nil)}
	out := WeighCandidateHomes(&NoteRow{Title: "Agent Site",
		Tags: []string{"agents"}, Domain: "a.com"}, inv, opts, 40,
		VaultKindWebsites)
	if len(out) < 2 || out[0].Folder != "AI/Small" {
		t.Fatalf("out %+v", out)
	}
	if !strings.Contains(strings.Join(out[1].Reasons, " "), "TOO BROAD") {
		t.Fatalf("reasons %+v", out[1].Reasons)
	}
}

func TestTheTopKCapRides(t *testing.T) {
	var folders []FolderCount
	for i := 0; i < 6; i++ {
		folders = append(folders, FolderCount{
			Rel: "AI/Folder00"[:9] + string(rune('0'+i)), NoteCount: 0})
	}
	for i := 0; i < 6; i++ {
		folders[i] = FolderCount{
			Rel: "AI/Folder0" + string(rune('0'+i)), NoteCount: 0}
	}
	inv := &Inventory{Folders: folders}
	cand := NoteRow{
		Title:  "Folder00 Folder01 Folder02 Folder03 Folder04 Folder05",
		Domain: "x.com"}
	out := WeighCandidateHomes(&cand, inv, nil, 40, VaultKindWebsites)
	if len(out) != WeightTopK {
		t.Fatalf("len %d != %d", len(out), WeightTopK)
	}
	if out[0].Folder != "AI/Folder00" || out[1].Folder != "AI/Folder01" ||
		out[2].Folder != "AI/Folder02" {
		t.Fatalf("folders %+v", out)
	}
}

func TestWeightsAreBoundedAndDeterministic(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI/Agents", NoteCount: 500}}}
	var notes []NoteRow
	for i := 0; i < 5; i++ {
		notes = append(notes, settledRow(
			fmt.Sprintf("/v/%d.md", i), "AI/Agents", "k.com",
			withRowTags("agents", "llm", "ai", "bots"),
			withRowTitle("Agent LLM AI Bots")))
	}
	inv.Notes = notes
	opts := &WeighOptions{Profiles: FolderProfiles(inv, nil),
		DomainHomes: SettledDomainHomes(inv, nil)}
	cand := NoteRow{Title: "Agent LLM AI Bots Kitchen Sink",
		Tags:   []string{"agents", "llm", "ai", "bots"},
		Domain: "k.com", Category: "AI", Subcategory: "Agents"}
	first := WeighCandidateHomes(&cand, inv, opts, 40, VaultKindWebsites)
	again := WeighCandidateHomes(&cand, inv, opts, 40, VaultKindWebsites)
	if len(first) != len(again) {
		t.Fatal("nondeterministic")
	}
	for i := range first {
		if first[i].Folder != again[i].Folder ||
			first[i].Weight != again[i].Weight {
			t.Fatal("nondeterministic")
		}
		if first[i].Weight < 0.0 || first[i].Weight > 1.0 {
			t.Fatalf("weight %v out of bounds", first[i].Weight)
		}
	}
	if first[0].Weight > 1.0 {
		t.Fatal("unbounded")
	}
}

// --- TestTheAttachment -------------------------------------------------

func TestDecoratesInPlaceAndCountsHonestly(t *testing.T) {
	inv := &Inventory{
		Folders: []FolderCount{{Rel: "AI/Chatbots", NoteCount: 1}},
		Notes: []NoteRow{
			settledRow("/v/AI/Chatbots/c.md", "AI/Chatbots",
				"x.com", withRowTags("chatbots"),
				withRowTitle("Chatbot Central"))}}
	cands := []*NoteRow{
		{File: "orphan.md", Path: "/v/orphan.md",
			Title: "A Chatbot Site", Tags: []string{"chatbots"},
			Domain: "new.com"},
		{File: "alien.md", Path: "/v/alien.md",
			Title: "Alien Topic", Domain: "alien.com"},
	}
	n := AttachWeighedHomes(cands, inv,
		map[string]bool{"/v/orphan.md": true, "/v/alien.md": true},
		nil, 40, VaultKindWebsites)
	if n != 1 {
		t.Fatalf("n %d", n)
	}
	if len(cands[0].Homes) == 0 ||
		cands[0].Homes[0].Folder != "AI/Chatbots" {
		t.Fatalf("homes %+v", cands[0].Homes)
	}
	if len(cands[1].Homes) != 0 || cands[1].HomesClose {
		t.Fatalf("alien %+v", cands[1])
	}
}

func TestTheCloseCallVerdict(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI/Agents", NoteCount: 0},
		{Rel: "AI/Chatbots", NoteCount: 0}}}
	cands := []*NoteRow{{Title: "Agents Chatbots", Domain: "x.com"}}
	AttachWeighedHomes(cands, inv, nil, nil, 40, VaultKindWebsites)
	if !cands[0].HomesClose {
		t.Fatal("not a close call")
	}
	if len(cands[0].Homes) != 2 {
		t.Fatalf("homes %+v", cands[0].Homes)
	}
}

func TestALandslideIsNotACloseCall(t *testing.T) {
	inv := &Inventory{Folders: []FolderCount{
		{Rel: "AI/Agents", NoteCount: 0},
		{Rel: "AI/Chatbots", NoteCount: 0}}}
	cands := []*NoteRow{{Title: "Chatbots Agents", Domain: "x.com",
		Category: "AI", Subcategory: "Chatbots"}}
	AttachWeighedHomes(cands, inv, nil, nil, 40, VaultKindWebsites)
	w := map[string]float64{}
	for _, h := range cands[0].Homes {
		w[h.Folder] = h.Weight
	}
	if !(w["AI/Chatbots"]-w["AI/Agents"] > WeightTieMargin) {
		t.Fatalf("weights %v", w)
	}
	if cands[0].HomesClose {
		t.Fatal("a landslide read as close")
	}
}

// --- TestTheHomesCell --------------------------------------------------

func TestTheCellNamesTheHomesTheirWeightsAndReasons(t *testing.T) {
	line := HomesLine(&NoteRow{
		Homes: []WeighedHome{
			{Folder: "AI/Chatbots", Weight: 0.62,
				Reasons: []string{"shared tags: chat", "name match"}},
			{Folder: "AI/Agents", Weight: 0.41,
				Reasons: []string{"shared tags: chat"}},
		},
		HomesClose: false})
	if !strings.HasPrefix(line, "homes: 1) AI/Chatbots 0.62") {
		t.Fatalf("line %q", line)
	}
	if !strings.Contains(line, "(shared tags: chat; name match)") {
		t.Fatalf("line %q", line)
	}
	if !strings.Contains(line, "2) AI/Agents 0.41") {
		t.Fatalf("line %q", line)
	}
	if strings.Contains(line, "CLOSE CALL") {
		t.Fatal("false CLOSE CALL")
	}
}

func TestTheCloseCallMarkerRides(t *testing.T) {
	line := HomesLine(&NoteRow{
		Homes: []WeighedHome{
			{Folder: "AI/Agents", Weight: 0.06},
			{Folder: "AI/Chatbots", Weight: 0.06},
		},
		HomesClose: true})
	if !strings.Contains(line, "CLOSE CALL") ||
		!strings.Contains(line, "you break the tie") {
		t.Fatalf("line %q", line)
	}
}

func TestNoHomesNoCell(t *testing.T) {
	if HomesLine(&NoteRow{}) != "" {
		t.Fatal("a cell rode without homes")
	}
	if HomesLine(nil) != "" {
		t.Fatal("nil rode")
	}
}

func TestTheDigestLineIsUnchangedWithoutHomes(t *testing.T) {
	lines := DigestLines([]NoteRow{{
		Title: "Note", Domain: "n.com", Folder: "", InRoot: true,
		Tags: []string{"tool"},
	}}, VaultKindWebsites)
	if len(lines) != 1 {
		t.Fatalf("lines %d", len(lines))
	}
	if strings.Contains(lines[0], "homes:") {
		t.Fatalf("line %q", lines[0])
	}
	if !strings.HasSuffix(lines[0], "tags: tool") {
		t.Fatalf("line %q", lines[0])
	}
}

func TestTheDigestLineCarriesTheShortlist(t *testing.T) {
	lines := DigestLines([]NoteRow{{
		Title: "Note", Domain: "n.com", Folder: "", InRoot: true,
		Tags: []string{"tool"},
		Homes: []WeighedHome{{Folder: "AI/Agents", Weight: 0.31,
			Reasons: []string{"shared tags: tool"}}},
	}}, VaultKindWebsites)
	if !strings.Contains(lines[0], "homes: 1) AI/Agents 0.31") {
		t.Fatalf("line %q", lines[0])
	}
	if !strings.Contains(lines[0], "shared tags: tool") {
		t.Fatalf("line %q", lines[0])
	}
}

// --- TestTheAlignmentDoor ----------------------------------------------

func vaultWithEvidence(c *scanCase) {
	c.note("AI/Chatbots/one.md", "https://split.example/1",
		withTags("chatbots", "llm", "ai"))
	c.note("AI/Chatbots/two.md", "https://split.example/2",
		withTags("chatbots", "llm", "ai"))
	c.note("AI/Agents/three.md", "https://split.example/3",
		withTags("agents"))
	c.note("wondering.md", "https://split.example",
		withTags("chatbots", "llm", "ai"),
		withCategory("AI Extra"), withSubcategory("Chatbots"))
	c.note("alien.md", "https://alien.example/1")
}

func TestFilingIntoTheTopHomeRidesUnflagged(t *testing.T) {
	c := newScanCase(t)
	vaultWithEvidence(c)
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{map[string]any{"note": "wondering.md",
			"to":         "AI/Chatbots",
			"reason":     "settled siblings and shared tags",
			"confidence": "high"}},
		"summary": "ok"})
	plan := c.buildPlan(llm.call)
	for _, m := range plan.Moves {
		if m.Note == "wondering.md" && m.LowConfidence {
			t.Fatal("the top-home move was flagged")
		}
	}
	if c.logs.has("rides against the weighed evidence") {
		t.Fatal("a false override log")
	}
}

func TestOverrulingStrongEvidenceRidesFlagged(t *testing.T) {
	c := newScanCase(t)
	vaultWithEvidence(c)
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{map[string]any{"note": "wondering.md",
			"to":         "AI/Agents",
			"reason":     "I see agent signals the numbers cannot",
			"confidence": "high"}},
		"summary": "ok"})
	plan := c.buildPlan(llm.call)
	var move *MoveRow
	for i := range plan.Moves {
		if plan.Moves[i].Note == "wondering.md" {
			move = &plan.Moves[i]
		}
	}
	if move == nil || !move.LowConfidence {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if !c.logs.has("rides against the weighed evidence") {
		t.Fatal("missing override log")
	}
	if !c.logs.has("AI/Chatbots") {
		t.Fatal("the top home's name missing from the log")
	}
}

func TestFilingIntoAnUnlistedHomeRidesFlagged(t *testing.T) {
	c := newScanCase(t)
	vaultWithEvidence(c)
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{"AI/Wonder-Land"},
		"moves": []any{map[string]any{"note": "wondering.md",
			"to": "AI/Wonder-Land", "reason": "a new home",
			"confidence": "high"}},
		"summary": "ok"})
	plan := c.buildPlan(llm.call)
	var move *MoveRow
	for i := range plan.Moves {
		if plan.Moves[i].Note == "wondering.md" {
			move = &plan.Moves[i]
		}
	}
	if move == nil || !move.LowConfidence {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestANearTieOverruledIsAJudgmentCall(t *testing.T) {
	c := newScanCase(t)
	c.note("AI/Agents/a.md", "https://tie.example/1",
		withTags("agents", "tie"))
	c.note("AI/Chatbots/b.md", "https://tie.example/2",
		withTags("chatbots", "tie"))
	c.note("wondering.md", "https://tie.example", withTags("tie"))
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{map[string]any{"note": "wondering.md",
			"to":         "AI/Chatbots",
			"reason":     "the tie broke on the note's own words",
			"confidence": "high"}},
		"summary": "ok"})
	plan := c.buildPlan(llm.call)
	var move *MoveRow
	for i := range plan.Moves {
		if plan.Moves[i].Note == "wondering.md" {
			move = &plan.Moves[i]
		}
	}
	if move == nil || move.LowConfidence {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

func TestTheGuestsOwnLowWordStillFlagsAlone(t *testing.T) {
	c := newScanCase(t)
	vaultWithEvidence(c)
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{map[string]any{"note": "alien.md",
			"to": "AI/Agents", "reason": "thin signals",
			"confidence": "low"}},
		"summary": "ok"})
	plan := c.buildPlan(llm.call)
	var move *MoveRow
	for i := range plan.Moves {
		if plan.Moves[i].Note == "alien.md" {
			move = &plan.Moves[i]
		}
	}
	if move == nil || !move.LowConfidence {
		t.Fatalf("moves %+v", plan.Moves)
	}
}

// --- TestTheWiring -----------------------------------------------------

func TestThePlanAndThePromptCarryTheEvidence(t *testing.T) {
	c := newScanCase(t)
	c.note("AI/Chatbots/one.md", "https://settled.example/1",
		withTags("chatbots"))
	c.note("AI/Chatbots/two.md", "https://settled.example/2",
		withTags("chatbots"))
	c.note("wondering.md", "https://wonder.example/9",
		withTags("chatbots"), withCategory("AI"),
		withSubcategory("Chatbots Extra"))
	llm := newFakeLLM(map[string]any{"new_folders": []any{},
		"moves": []any{}, "summary": "nothing"})
	plan := c.buildPlan(llm.call)
	if plan.LLMCandidates != 1 || plan.WeighedHomes != 1 {
		t.Fatalf("plan %+v", plan)
	}
	body := llm.calls[0][1].Content
	for _, needle := range []string{"homes: 1) AI/Chatbots",
		"0.0-1.0 weight", "ADVICE, not law", "CLOSE CALL"} {
		if !strings.Contains(body, needle) {
			t.Fatalf("user message missing %q", needle)
		}
	}
	if !c.logs.has("weighed shortlist") {
		t.Fatal("missing evidence log")
	}
}

func TestNoCandidatesStillAZeroCount(t *testing.T) {
	c := newScanCase(t)
	c.note("AI/Chatbots/one.md", "https://settled.example/1",
		withTags("chat"), withCategory("AI"),
		withSubcategory("Chatbots"))
	plan := c.buildPlan(nil)
	if plan.WeighedHomes != 0 || plan.LLMCandidates != 0 {
		t.Fatalf("plan %+v", plan)
	}
}

func TestACandidateNeverTeachesTheScale(t *testing.T) {
	c := newScanCase(t)
	for i := 0; i < 41; i++ {
		c.note(fmt.Sprintf("AI/Chatbots/bulk%02d.md", i),
			fmt.Sprintf("https://bulk.example/%d", i),
			withTags("wonder"))
	}
	c.note("AI/Agents/settled.md", "https://settled.example/1",
		withTags("wonder"))
	c.note("wondering.md", "https://wonder.example/9", withTags("wonder"))
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	// the scale decorates the SAME rows the plan threads — take the
	// candidates' rows by path from the inventory (the pointer law).
	moveRows := MoveCandidates(&inv, 40)
	byPath := map[string]*NoteRow{}
	for i := range inv.Notes {
		byPath[inv.Notes[i].Path] = &inv.Notes[i]
	}
	ptrs := make([]*NoteRow, 0, len(moveRows))
	excl := map[string]bool{}
	for _, m := range moveRows {
		ptrs = append(ptrs, byPath[m.Path])
		excl[m.Path] = true
	}
	AttachWeighedHomes(ptrs, &inv, excl, nil, 40, VaultKindWebsites)
	var wonder *NoteRow
	for _, cand := range ptrs {
		if cand.File == "wondering.md" {
			wonder = cand
		}
	}
	folders := map[string]bool{}
	for _, h := range wonder.Homes {
		folders[h.Folder] = true
	}
	if !folders["AI/Agents"] {
		t.Fatalf("the settled teacher missing: %v", folders)
	}
	if folders["AI/Chatbots"] {
		t.Fatal("the too-broad folder's candidates taught")
	}
}

func TestALockedNoteNeverTeachesTheScale(t *testing.T) {
	c := newScanCase(t)
	c.note("AI/Chatbots/locked.md", "https://locked.example/1",
		withTags("chatbots"), withExtra("locked: true"))
	c.note("AI/Agents/open.md", "https://open.example/1",
		withTags("agents"))
	c.note("wondering.md", "https://wonder.example/9",
		withTags("chatbots"))
	c.note("wondering2.md", "https://wonder2.example/9",
		withTags("agents"))
	inv := ScanVaultInventory(c.vault, c.logs.log, nil, nil,
		VaultKindWebsites)
	moveRows := MoveCandidates(&inv, 40)
	byPath := map[string]*NoteRow{}
	for i := range inv.Notes {
		byPath[inv.Notes[i].Path] = &inv.Notes[i]
	}
	ptrs := make([]*NoteRow, 0, len(moveRows))
	exclude := map[string]bool{}
	for _, m := range moveRows {
		ptrs = append(ptrs, byPath[m.Path])
		exclude[m.Path] = true
	}
	AttachWeighedHomes(ptrs, &inv, exclude, nil, 40, VaultKindWebsites)
	byFile := map[string]*NoteRow{}
	for _, cand := range ptrs {
		byFile[cand.File] = cand
	}
	// the open resident teaches: its shared tag is named
	agents := byFile["wondering2.md"].Homes
	openTaught := false
	for _, h := range agents {
		for _, r := range h.Reasons {
			if strings.Contains(r, "shared tags: agents") {
				openTaught = true
			}
		}
	}
	if !openTaught {
		t.Fatalf("the open resident never taught: %+v", agents)
	}
	// the locked resident never teaches
	chat := byFile["wondering.md"].Homes
	for _, h := range chat {
		for _, r := range h.Reasons {
			if strings.Contains(r, "shared tags") {
				t.Fatalf("the locked resident taught: %q", r)
			}
		}
	}
}

// --- TestTheModulePinsTheLawsConstants ---------------------------------

func TestTheModulePinsTheLawsConstants(t *testing.T) {
	if WeightTopK != 3 {
		t.Fatal("top K")
	}
	if WeightTieMargin != 0.08 {
		t.Fatal("tie margin")
	}
	if StrongEvidenceWeight != 0.55 {
		t.Fatal("strong weight")
	}
	if OverrideMargin != 0.15 {
		t.Fatal("override margin")
	}
}

var _ = sort.Strings
var _ = llmclient.Message{}
