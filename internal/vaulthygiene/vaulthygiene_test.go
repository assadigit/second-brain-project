// vaulthygiene_test.go — T-46: the ported selections of the reference's
// own suites — test_strayheal (the identity harvest through the one
// walk, the stray retirement, THE LINKED-GHOST DOOR, heal_machine_
// vaults, the pure classifiers), test_vaulthygiene (the hygiene pass's
// laws through RunVaultHygiene directly — the reference drives them
// through the GUI worker's _vault_hygiene_pass, whose Go half is the
// future GUI task's) and test_picklock's locked-bodyless law (the lock
// guard over the app-owned bodyless door).
//
// The GUI legs (the scan door's both-vault heal, the startup heal, the
// scorecard rows) stay with the future GUI task; the release-bookkeeping
// legs are the Python repo's own CI facts; the resolvable-link law legs
// are core/linking.py's own suite (the Phase 1 links port). The
// collect_real_note_identities legs ride the ONE WALK (the reference's
// standalone collector is dead upstream — the identities are observable
// through the twins they vouch for). The reference's None/junk legs are
// type-narrowed where Go's string door carries the same verdict.
package vaulthygiene

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
)

// --- the shared fixtures (the reference suites' shapes) -------------------

func writeT(t *testing.T, path, text string) string {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

func readT(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// logSink collects (level, msg) — the reference suites' _Log/_Recorder.
type logSink struct {
	entries [][2]string
}

func (l *logSink) log(msg, level string) {
	l.entries = append(l.entries, [2]string{level, msg})
}

func (l *logSink) joined() string {
	msgs := make([]string, len(l.entries))
	for i, e := range l.entries {
		msgs[i] = e[1]
	}
	return strings.Join(msgs, "\n")
}

const realNoteText = `---
source: "https://github.com/someone/subtitle-translator"
aliases:
  - subtitle-translator
  - someone/subtitle-translator
tags: [typescript, video]
category: "LLM-Tools"
stars: 120
managed_by: gitcurator
---

# subtitle-translator

**TL;DR:** translates subtitles with LLMs.

## What is it?
A real note with real body text.
`

// strayVault mirrors StrayVaultBase: the owner's distilled GitHub
// vault shape.
type strayVault struct {
	t     *testing.T
	tmp   string
	vault string
	real  string
	logs  *logSink
}

func newStrayVault(t *testing.T) *strayVault {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "github-vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	s := &strayVault{t: t, tmp: tmp, vault: vault, logs: &logSink{}}
	s.real = writeT(t,
		filepath.Join(vault, "AI-Domain", "LLM-Tools",
			"subtitle-translator_LLM-Tools_typescript.md"), realNoteText)
	// the retired master-index scaffold (still in the wild)
	writeT(t, filepath.Join(vault, "_index.md"),
		"---\ntype: master-index\n---\n\n# Index\n\n"+
			"- [[subtitle-translator_LLM-Tools_typescript]]\n")
	writeT(t, filepath.Join(vault, "_moc", "LLM-Tools.md"),
		"---\ntype: moc\n---\n\n- "+
			"[[subtitle-translator_LLM-Tools_typescript]]\n")
	// THE STRAYS at the vault root (the owner's 3-4)
	writeT(t, filepath.Join(vault, "subtitle-translator.md"), "")
	writeT(t, filepath.Join(vault, "Subtitle Translator.md"),
		"# Subtitle Translator\n")
	writeT(t, filepath.Join(vault, "another-repo_LLM-Tools_automation.md"),
		"# another-repo_LLM-Tools_automation\n")
	// the owner's own — MUST SURVIVE
	writeT(t, filepath.Join(vault, "my-thoughts.md"), "# my-thoughts\n")
	// the legacy root report
	writeT(t, filepath.Join(vault, "_processing_report_20260101.md"),
		"# report\n")
	return s
}

func (s *strayVault) run() HygieneCounts {
	return RunVaultHygiene(s.vault, s.logs.log, nil, nil)
}

// --- TestNormalizeIdentity --------------------------------------------------

func TestTheThreeSpellingFamiliesMeet(t *testing.T) {
	a := NormalizeIdentity("subtitle-translator")
	if a != NormalizeIdentity("Subtitle Translator") ||
		a != NormalizeIdentity("subtitle_translator") || a != "subtitletranslator" {
		t.Fatalf("the families did not meet: %q", a)
	}
}

func TestJunkAndEmptyNeverDiverge(t *testing.T) {
	for _, c := range []struct{ in, want string }{
		{"", ""}, {"!!??__--", ""}, {"  A  B  ", "ab"},
	} {
		if got := NormalizeIdentity(c.in); got != c.want {
			t.Fatalf("NormalizeIdentity(%q) = %q; want %q", c.in, got, c.want)
		}
	}
}

// --- the identity harvest through the one walk ------------------------------
// (the reference's TestIdentityHarvest legs — the standalone collector
// is dead upstream; the identities are observable through the twins
// they vouch for)

func TestStemH1AndAliasesAllVouchForTheirTwins(t *testing.T) {
	v := t.TempDir()
	writeT(t, filepath.Join(v, "AI-Domain", "LLM-Tools",
		"subtitle-translator_LLM-Tools_typescript.md"), realNoteText)
	// three stubs, each a twin of one identity the real note answers to:
	// its H1/alias, its second alias, and its full stem (in another
	// folder — the same stem cannot live twice in one folder).
	writeT(t, filepath.Join(v, "subtitle-translator.md"), "")
	writeT(t, filepath.Join(v, "someone-subtitle-translator.md"), "")
	writeT(t, filepath.Join(v, "Elsewhere",
		"subtitle-translator_LLM-Tools_typescript.md"), "")
	res := RunVaultHygiene(v, nil, nil, nil)
	if res.RetiredStubs != 3 {
		t.Fatalf("retired %d; want 3 (the stem/H1/alias twins)", res.RetiredStubs)
	}
}

func TestInlineAliasFlowListVouches(t *testing.T) {
	v := t.TempDir()
	writeT(t, filepath.Join(v, "note.md"),
		"---\naliases: [Alpha Tool, beta]\n---\n\n# Heading\n\nBody text here.\n")
	for _, stray := range []string{"alpha-tool.md", "beta.md", "heading.md"} {
		writeT(t, filepath.Join(v, stray), "")
	}
	res := RunVaultHygiene(v, nil, nil, nil)
	if res.RetiredStubs != 3 {
		t.Fatalf("retired %d; want 3 (the inline aliases + the H1)", res.RetiredStubs)
	}
}

func TestTheAutoHeadingTrapNeverVouches(t *testing.T) {
	// A lone stray with an Obsidian auto-heading must NOT be retired —
	// its own heading can never vouch for it.
	v := t.TempDir()
	writeT(t, filepath.Join(v, "subtitle-translator.md"),
		"# subtitle-translator\n")
	res := RunVaultHygiene(v, nil, nil, nil)
	if res.RetiredStubs != 0 {
		t.Fatalf("retired %d; want 0 (the self-twin trap)", res.RetiredStubs)
	}
	if _, err := os.Stat(filepath.Join(v, "subtitle-translator.md")); err != nil {
		t.Fatal("the stray was touched")
	}
}

func TestABigFileContributesItsStemWithoutAHeadRead(t *testing.T) {
	v := t.TempDir()
	// bigger than the walk's read cap: real writing — the stem identity
	// rides the size gate, the head is never read (the H1 CANNOT vouch).
	writeT(t, filepath.Join(v, "AI-Domain", "LLM-Tools",
		"big-repo_LLM-Tools_go.md"),
		"# big-repo H1 words\n\n"+strings.Repeat("x", 1_100_000))
	writeT(t, filepath.Join(v, "big-repo_LLM-Tools_go.md"), "") // the stem twin
	writeT(t, filepath.Join(v, "big-repo-h1-words.md"), "")     // the H1 non-twin
	res := RunVaultHygiene(v, nil, nil, nil)
	if res.RetiredStubs != 1 {
		t.Fatalf("retired %d; want 1 (the stem twin only — the unread H1 never vouched)",
			res.RetiredStubs)
	}
	if _, err := os.Stat(filepath.Join(v, "big-repo-h1-words.md")); err != nil {
		t.Fatal("the H1 non-twin was retired — the head was read past the cap")
	}
}

func TestMissingVaultIsTheQuietZero(t *testing.T) {
	res := RunVaultHygiene(filepath.Join(t.TempDir(), "nope"), nil, nil, nil)
	if res.RetiredStubs != 0 || res.RemovedIndex != 0 || res.RetiredReports != 0 {
		t.Fatalf("missing vault answered %+v", res)
	}
}

// --- TestStrayRetirement -------------------------------------------------------

func TestTheOwnersExactScenarioHealsCompletely(t *testing.T) {
	s := newStrayVault(t)
	res := s.run()
	if res.RemovedIndex != 2 {
		t.Fatalf("removed_index %d; want 2 (_index + _moc)", res.RemovedIndex)
	}
	if res.RetiredStubs != 3 {
		t.Fatalf("retired_stubs %d; want 3 (the strays)", res.RetiredStubs)
	}
	if res.RetiredReports != 1 {
		t.Fatalf("retired_reports %d; want 1", res.RetiredReports)
	}
	for _, gone := range []string{
		"subtitle-translator.md", "Subtitle Translator.md",
		"another-repo_LLM-Tools_automation.md", "_index.md",
	} {
		if _, err := os.Stat(filepath.Join(s.vault, gone)); err == nil {
			t.Fatalf("%s still present", gone)
		}
	}
	// no .md left at the vault root except the owner's own
	rootMD := rootMDNames(t, s.vault)
	if len(rootMD) != 1 || rootMD[0] != "my-thoughts.md" {
		t.Fatalf("root .md after = %v; want [my-thoughts.md]", rootMD)
	}
}

func TestEveryRetirementIsRecoverable(t *testing.T) {
	s := newStrayVault(t)
	s.run()
	trash := filepath.Join(s.vault, ".trash", "empty-stubs")
	retired := listNames(t, trash)
	want := []string{"Subtitle Translator.md",
		"another-repo_LLM-Tools_automation.md", "subtitle-translator.md"}
	if strings.Join(retired, "|") != strings.Join(want, "|") {
		t.Fatalf("retired set = %v; want %v", retired, want)
	}
	for _, fname := range retired {
		if _, err := os.Stat(filepath.Join(trash, fname)); err != nil {
			t.Fatalf("%s not recoverable", fname)
		}
	}
	if body := readT(t, filepath.Join(trash, "subtitle-translator.md")); body != "" {
		t.Fatalf("bytes not preserved: %q", body)
	}
}

func TestTheRealNoteAndTheOwnersWritingSurvive(t *testing.T) {
	s := newStrayVault(t)
	s.run()
	if _, err := os.Stat(s.real); err != nil {
		t.Fatal("the real note was touched")
	}
	if body := readT(t, s.real); !strings.Contains(body, "# subtitle-translator") {
		t.Fatal("the real note's bytes changed")
	}
	if _, err := os.Stat(filepath.Join(s.vault, "my-thoughts.md")); err != nil {
		t.Fatal("the owner's note was touched")
	}
}

func TestIdempotentSecondRunRetiresNothing(t *testing.T) {
	s := newStrayVault(t)
	s.run()
	again := RunVaultHygiene(s.vault, nil, nil, nil)
	if again.RemovedIndex != 0 || again.RetiredStubs != 0 || again.RetiredReports != 0 {
		t.Fatalf("second run retired %+v", again)
	}
}

func TestTheStoryNamesTheStraysInTheLog(t *testing.T) {
	s := newStrayVault(t)
	s.run()
	j := s.logs.joined()
	for _, want := range []string{
		"subtitle-translator", "stray empty note", "master index",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("the story misses %q:\n%s", want, j)
		}
	}
}

func TestDryRunRehearsesAndTouchesNothing(t *testing.T) {
	s := newStrayVault(t)
	before := listNames(t, s.vault)
	dryrun.Enable()
	res := RunVaultHygiene(s.vault, nil, nil, nil)
	dryrun.Disable()
	dryrun.Clear()
	if res.RetiredStubs != 3 {
		t.Fatalf("rehearsed retired %d; want 3 (honest counts)", res.RetiredStubs)
	}
	if strings.Join(listNames(t, s.vault), "|") != strings.Join(before, "|") {
		t.Fatal("the dry run touched the vault")
	}
}

func TestAFrontmatterdStrayIsNeverTouched(t *testing.T) {
	s := newStrayVault(t)
	stray := writeT(t, filepath.Join(s.vault, "subtitle-translator-2.md"),
		"---\ntitle: mine\n---\n\n# mine\n")
	// give it a real twin: a real note whose alias is the stem
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools",
		"twin_LLM-Tools_x.md"),
		"---\naliases:\n  - subtitle-translator-2\n---\n\n# twin\n\nbody\n")
	RunVaultHygiene(s.vault, nil, nil, nil)
	if _, err := os.Stat(stray); err != nil {
		t.Fatal("a frontmatter'd stray was touched")
	}
}

func TestAStrayWithBodyTextIsNeverTouched(t *testing.T) {
	s := newStrayVault(t)
	stray := writeT(t, filepath.Join(s.vault, "subtitle-translator.md"),
		"# subtitle-translator\n\nreal body text here\n")
	RunVaultHygiene(s.vault, nil, nil, nil)
	if _, err := os.Stat(stray); err != nil {
		t.Fatal("a stray with body text was touched")
	}
}

// --- TestHealMachineVaults ------------------------------------------------------

func TestBothMachineVaultsHealUnderOneConfig(t *testing.T) {
	tmp := t.TempDir()
	github := filepath.Join(tmp, "github")
	websites := filepath.Join(tmp, "websites")
	for _, d := range []string{github, websites} {
		if err := os.MkdirAll(d, 0o755); err != nil {
			t.Fatal(err)
		}
	}
	writeT(t, filepath.Join(github, "stray_LLM-Tools_x.md"), "")
	writeT(t, filepath.Join(websites, "_index.md"),
		"---\ntype: master-index\n---\n\n# i\n")
	writeT(t, filepath.Join(websites, "AI-Domain", "LLM-Tools",
		"stray_LLM-Tools_x.md"), realNoteText)
	healed := HealMachineVaults(map[string]any{
		"vault_path": github, "website_vault_path": websites}, nil, nil, nil)
	if _, ok := healed["github"]; !ok {
		t.Fatal("github missing from healed")
	}
	if _, ok := healed["websites"]; !ok {
		t.Fatal("websites missing from healed")
	}
	if healed["github"].RetiredStubs != 1 {
		t.Fatalf("github retired %d; want 1", healed["github"].RetiredStubs)
	}
	if healed["websites"].RemovedIndex != 1 {
		t.Fatalf("websites removed_index %d; want 1", healed["websites"].RemovedIndex)
	}
	if got := HygieneTotal(healed); got != 2 {
		t.Fatalf("hygiene_total %d; want 2", got)
	}
}

func TestMissingAndJunkVaultsAreSkippedQuietly(t *testing.T) {
	tmp := t.TempDir()
	healed := HealMachineVaults(map[string]any{
		"vault_path": filepath.Join(tmp, "nope")}, nil, nil, nil)
	if len(healed) != 0 || HygieneTotal(healed) != 0 {
		t.Fatalf("missing vault healed %+v", healed)
	}
	// junk configs never raise — a nil config is the reference's None.
	if healed := HealMachineVaults(nil, nil, nil, nil); len(healed) != 0 {
		t.Fatalf("nil config healed %+v", healed)
	}
}

func TestHygieneTotalCountsTheThreeKinds(t *testing.T) {
	if got := HygieneTotal(nil); got != 0 {
		t.Fatalf("total(nil) = %d", got)
	}
	healed := map[string]HygieneCounts{
		"a": {RemovedIndex: 1, RetiredStubs: 2, RetiredReports: 3},
		"b": {},
	}
	if got := HygieneTotal(healed); got != 6 {
		t.Fatalf("total = %d; want 6", got)
	}
}

// --- TestHarvestLinkTargets --------------------------------------------------------

func TestEveryLinkShapeIsHarvested(t *testing.T) {
	content := "See [[subtitle-translator]] and " +
		"[[Subtitle Translator|the translator]] plus " +
		"[[ghost#anchor]] and [[folder/note name]].\n" +
		"![[an-embed]] too"
	targets := HarvestLinkTargets(content)
	for _, want := range []string{
		"subtitletranslator", "ghost", "notename", "anembed",
	} {
		if !targets[want] {
			t.Fatalf("target %q missing from %v", want, targets)
		}
	}
}

func TestJunkBodiesYieldTheEmptySet(t *testing.T) {
	for _, junk := range []string{"", "no links here", "[[", "[[ ]]"} {
		if got := HarvestLinkTargets(junk); len(got) != 0 {
			t.Fatalf("junk %q yielded %v", junk, got)
		}
	}
}

// --- TestIsAppOwnedBodyless ---------------------------------------------------------

func TestTheFailedWriteShapeIsTrue(t *testing.T) {
	content := "---\n" +
		"title: Some repo\n" +
		"source: https://github.com/owner/repo\n" +
		"managed_by: GitCurator\n" +
		"tags: [LLM-Tools]\n" +
		"---\n\n" +
		"# Some repo\n"
	if !IsAppOwnedBodyless(content) {
		t.Fatal("the failed-write shape was not recognized")
	}
}

func TestABodyWithRealTextIsFalse(t *testing.T) {
	content := "---\nmanaged_by: GitCurator\n---\n\n" +
		"# Some repo\n\nReal summary text.\n"
	if IsAppOwnedBodyless(content) {
		t.Fatal("a finished note was judged bodyless")
	}
}

func TestTheOwnersOwnFrontmatterIsSacred(t *testing.T) {
	if IsAppOwnedBodyless("---\ntitle: mine\n---\n\n# mine\n") {
		t.Fatal("the owner's frontmatter was claimed")
	}
}

func TestNoFrontmatterIsNotOurs(t *testing.T) {
	for _, c := range []string{"", "# just a heading\n"} {
		if IsAppOwnedBodyless(c) {
			t.Fatalf("%q claimed without frontmatter", c)
		}
	}
}

func TestTwoHeadingsIsARealNote(t *testing.T) {
	if IsAppOwnedBodyless("---\nmanaged_by: GitCurator\n---\n\n# One\n\n# Two\n") {
		t.Fatal("two headings judged bodyless")
	}
}

func TestFenceShapesNeverPanic(t *testing.T) {
	for _, c := range []string{"---\n", "---\n---\n", "---\nmanaged_by: x\n...\n\n# h\n"} {
		IsAppOwnedBodyless(c) // never panics
	}
}

// --- TestTheLinkedGhostDoor ----------------------------------------------------------

func TestALinkedGhostRetiresEvenWithoutTwinOrShape(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "linker.md"),
		realNoteText+"\nSee [[ghost-target]] for the twin.\n")
	writeT(t, filepath.Join(s.vault, "ghost-target.md"), "")
	res := s.run()
	found := false
	for _, n := range res.StubNames {
		if n == "ghost-target" {
			found = true
		}
	}
	if !found {
		t.Fatalf("ghost-target missing from %v", res.StubNames)
	}
	if _, err := os.Stat(filepath.Join(s.vault, "ghost-target.md")); err == nil {
		t.Fatal("the linked ghost stayed")
	}
	if res.RetiredLinked != 1 {
		t.Fatalf("retired_linked %d; want 1", res.RetiredLinked)
	}
	// the owner's own empty note — nobody links it — survives
	if _, err := os.Stat(filepath.Join(s.vault, "my-thoughts.md")); err != nil {
		t.Fatal("the owner's empty note was touched")
	}
}

func TestTheLinkMayLiveInAnAppOwnedNoteToo(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "real.md"),
		realNoteText+"\nRelated: [[some-ghost]]\n")
	writeT(t, filepath.Join(s.vault, "some-ghost.md"), "")
	res := s.run()
	found := false
	for _, n := range res.StubNames {
		if n == "some-ghost" {
			found = true
		}
	}
	if !found {
		t.Fatalf("some-ghost missing from %v", res.StubNames)
	}
}

func TestASelfLinkNeverVouchesForTheStubItself(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "lonely2.md"), "")
	res := s.run()
	for _, n := range res.StubNames {
		if n == "lonely2" {
			t.Fatal("a self-link vouched for the stub")
		}
	}
	if _, err := os.Stat(filepath.Join(s.vault, "lonely2.md")); err != nil {
		t.Fatal("lonely2 was touched")
	}
}

func TestTheAppOwnedBodylessStrayRetires(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "keeper.md"),
		realNoteText+"\nSee [[repo-note]].\n")
	writeT(t, filepath.Join(s.vault, "repo-note.md"),
		"---\n"+
			"title: repo-note\n"+
			"source: https://github.com/owner/repo\n"+
			"managed_by: GitCurator\n"+
			"tags: [LLM-Tools]\n"+
			"---\n"+
			"\n"+
			"# repo-note\n")
	res := s.run()
	if _, err := os.Stat(filepath.Join(s.vault, "repo-note.md")); err == nil {
		t.Fatal("the app-owned bodyless stray stayed")
	}
	if res.RetiredAppOwned != 1 {
		t.Fatalf("retired_app_owned %d; want 1", res.RetiredAppOwned)
	}
	if _, err := os.Stat(filepath.Join(s.vault, ".trash", "empty-stubs",
		"repo-note.md")); err != nil {
		t.Fatal("the bodyless stray is not recoverable")
	}
}

func TestAnAppOwnedBodylessUnlinkedStrayStays(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "quiet-placeholder.md"),
		"---\n"+
			"managed_by: GitCurator\n"+
			"fetch_status: failed\n"+
			"---\n"+
			"\n"+
			"# quiet-placeholder\n")
	RunVaultHygiene(s.vault, nil, nil, nil)
	if _, err := os.Stat(filepath.Join(s.vault, "quiet-placeholder.md")); err != nil {
		t.Fatal("the unlinked bodyless stray was touched (the sacred law)")
	}
}

func TestTheHandWrittenFrontmatterBodylessNoteIsSacred(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "linker.md"),
		realNoteText+"\nSee [[my-own-stub]].\n")
	writeT(t, filepath.Join(s.vault, "my-own-stub.md"),
		"---\ntitle: my own\n---\n\n# my own\n")
	RunVaultHygiene(s.vault, nil, nil, nil)
	if _, err := os.Stat(filepath.Join(s.vault, "my-own-stub.md")); err != nil {
		t.Fatal("a hand-written frontmatter bodyless note was touched")
	}
}

func TestTheRetirementIsRecoverableAndIdempotent(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "linker2.md"),
		realNoteText+"\n[[ghost-2]]\n")
	writeT(t, filepath.Join(s.vault, "ghost-2.md"), "")
	RunVaultHygiene(s.vault, nil, nil, nil)
	if _, err := os.Stat(filepath.Join(s.vault, ".trash", "empty-stubs",
		"ghost-2.md")); err != nil {
		t.Fatal("ghost-2 not recoverable")
	}
	again := RunVaultHygiene(s.vault, nil, nil, nil)
	if again.RetiredStubs != 0 {
		t.Fatalf("second run retired %d", again.RetiredStubs)
	}
}

func TestTheStorySpeaksTheLinkedGhosts(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "linker3.md"),
		realNoteText+"\n[[ghost-3]]\n")
	writeT(t, filepath.Join(s.vault, "ghost-3.md"), "")
	RunVaultHygiene(s.vault, s.logs.log, nil, nil)
	if !strings.Contains(s.logs.joined(), "linked ghost") {
		t.Fatalf("the story misses the linked ghosts:\n%s", s.logs.joined())
	}
}

func TestDryRunRehearsesTheLinkedGhost(t *testing.T) {
	s := newStrayVault(t)
	writeT(t, filepath.Join(s.vault, "AI-Domain", "LLM-Tools", "linker4.md"),
		realNoteText+"\n[[ghost-4]]\n")
	writeT(t, filepath.Join(s.vault, "ghost-4.md"), "")
	before := listNames(t, s.vault)
	dryrun.Enable()
	res := RunVaultHygiene(s.vault, nil, nil, nil)
	dryrun.Disable()
	dryrun.Clear()
	if res.RetiredStubs < 1 {
		t.Fatalf("rehearsed retired %d; want ≥ 1", res.RetiredStubs)
	}
	if _, err := os.Stat(filepath.Join(s.vault, "ghost-4.md")); err != nil {
		t.Fatal("the dry run moved the ghost")
	}
	if strings.Join(listNames(t, s.vault), "|") != strings.Join(before, "|") {
		t.Fatal("the dry run touched the vault")
	}
}

func TestTheOneWalkProgressSpeaks(t *testing.T) {
	s := newStrayVault(t)
	type progCall struct {
		label string
		value int
		max   int
	}
	var seen []progCall
	RunVaultHygiene(s.vault, nil,
		func(label string, value, maximum int) {
			seen = append(seen, progCall{label, value, maximum})
		}, nil)
	if len(seen) == 0 {
		t.Fatal("the walk never spoke its counts")
	}
	last := seen[len(seen)-1]
	if last.max < last.value {
		t.Fatalf("progress %d > maximum %d", last.value, last.max)
	}
	if !strings.HasPrefix(last.label, "healing ") {
		t.Fatalf("progress label %q misses the healing word", last.label)
	}
}

// --- TestHygienePass (test_vaulthygiene's selections, direct) -------------------------

type hygieneCase struct {
	t     *testing.T
	tmp   string
	vault string
	logs  *logSink
}

func newHygieneCase(t *testing.T) *hygieneCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "GitHubProjects")
	if err := os.MkdirAll(filepath.Join(vault, "Tools", "Automation"), 0o755); err != nil {
		t.Fatal(err)
	}
	return &hygieneCase{t: t, tmp: tmp, vault: vault, logs: &logSink{}}
}

func (h *hygieneCase) run() *logSink {
	RunVaultHygiene(h.vault, h.logs.log, nil, nil)
	return h.logs
}

// pollute mirrors VaultHygieneTestBase._pollute: the owner's machine,
// distilled.
func (h *hygieneCase) pollute() (ownerMocFile, stubRoot, stubFolder,
	oldReport, oldSummary, realNote string) {
	v := h.vault
	// 1. the app-owned scaffold (the ghost-link breeders)
	writeT(h.t, filepath.Join(v, "_index.md"),
		"---\ntype: master-index\ntotal_projects: 700\n---\n"+
			"# Projects Master Index\n- [[agent_Automation_python]]\n"+
			"- [[banished_repo_Automation_rust]]\n")
	writeT(h.t, filepath.Join(v, "_moc", "Automation.md"),
		"---\ntype: moc\ncategory: Automation\n---\n"+
			"# Automation\n- [[agent_Automation_python]]\n")
	// an owner file squatting in _moc — sacred, keeps the folder
	ownerMocFile = writeT(h.t, filepath.Join(v, "_moc", "my-own-notes.md"),
		"# My own map\n- something I wrote by hand\n")
	// 2. the ghost-born empty stubs (heading-only, no frontmatter)
	stubRoot = writeT(h.t, filepath.Join(v,
		"subtitle-translator_LLM-Tools_typescript.md"),
		"# subtitle-translator LLM-Tools typescript\n")
	stubFolder = writeT(h.t, filepath.Join(v, "Tools", "Automation",
		"ghost_Automation_python.md"), "\n\n")
	// 3. the legacy root reports
	oldReport = writeT(h.t, filepath.Join(v,
		"_processing_report_20260101_000001.md"),
		"# Processing Report — 2026-01-01\n| table |\n")
	oldSummary = writeT(h.t, filepath.Join(v,
		"processing_summary_20260101_000001.txt"), "SUMMARY\n")
	// 4. the survivors
	realNote = writeT(h.t, filepath.Join(v, "Tools", "Automation",
		"agent_Automation_python.md"),
		"---\nsource: \"https://github.com/o/agent\"\ncategory: "+
			"\"Automation\"\nmanaged_by: gitcurator\n---\n# agent\n\n"+
			"A real curated note with a body.\n")
	writeT(h.t, filepath.Join(v, "Tools", "Automation", "my-thoughts.md"),
		"# Just a heading I wrote myself\n")
	writeT(h.t, filepath.Join(v, "ideas.md"), "# Ideas\n\n- a real list of my own\n")
	// the manifest + undo list are working files, not notes
	writeT(h.t, filepath.Join(v, "links_manifest.json"), "{}")
	writeT(h.t, filepath.Join(v, "_undo_last_batch.txt"), "")
	return
}

func TestScaffoldRemovedAndBacklinksHealed(t *testing.T) {
	h := newHygieneCase(t)
	h.pollute()
	h.run()
	for _, gone := range []string{
		filepath.Join(h.vault, "_index.md"),
		filepath.Join(h.vault, "_moc", "Automation.md"),
	} {
		if _, err := os.Stat(gone); err == nil {
			t.Fatalf("%s still present", gone)
		}
	}
}

func TestOwnerFileInMocKeepsFolder(t *testing.T) {
	h := newHygieneCase(t)
	ownerMocFile, _, _, _, _, _ := h.pollute()
	h.run()
	if _, err := os.Stat(ownerMocFile); err != nil {
		t.Fatal("the owner's _moc file was touched")
	}
	if _, err := os.Stat(filepath.Join(h.vault, "_moc")); err != nil {
		t.Fatal("the _moc folder was pruned with an owner file inside")
	}
}

func TestEmptyStubsRetiredRecoverably(t *testing.T) {
	h := newHygieneCase(t)
	_, stubRoot, stubFolder, _, _, _ := h.pollute()
	h.run()
	if _, err := os.Stat(stubRoot); err == nil {
		t.Fatal("the root stub stayed")
	}
	if _, err := os.Stat(stubFolder); err == nil {
		t.Fatal("the folder stub stayed")
	}
	retiredRoot := filepath.Join(h.vault, ".trash", "empty-stubs",
		"subtitle-translator_LLM-Tools_typescript.md")
	retiredFolder := filepath.Join(h.vault, ".trash", "empty-stubs",
		"ghost_Automation_python.md")
	if _, err := os.Stat(retiredRoot); err != nil {
		t.Fatal("the root stub is not recoverable")
	}
	if _, err := os.Stat(retiredFolder); err != nil {
		t.Fatal("the folder stub is not recoverable")
	}
	if body := readT(t, retiredRoot); !strings.Contains(body, "subtitle-translator") {
		t.Fatal("bytes not preserved")
	}
}

func TestLegacyRootReportsRetired(t *testing.T) {
	h := newHygieneCase(t)
	_, _, _, oldReport, oldSummary, _ := h.pollute()
	h.run()
	if _, err := os.Stat(oldReport); err == nil {
		t.Fatal("the old report stayed")
	}
	if _, err := os.Stat(oldSummary); err == nil {
		t.Fatal("the old summary stayed")
	}
	if _, err := os.Stat(filepath.Join(h.vault, ".trash", "retired-reports",
		"_processing_report_20260101_000001.md")); err != nil {
		t.Fatal("the report is not recoverable")
	}
	if _, err := os.Stat(filepath.Join(h.vault, ".trash", "retired-reports",
		"processing_summary_20260101_000001.txt")); err != nil {
		t.Fatal("the summary is not recoverable")
	}
}

func TestRealNotesAndOwnerWritingSurvive(t *testing.T) {
	h := newHygieneCase(t)
	_, _, _, _, _, realNote := h.pollute()
	h.run()
	if _, err := os.Stat(realNote); err != nil {
		t.Fatal("the real note was touched")
	}
	if _, err := os.Stat(filepath.Join(h.vault, "Tools", "Automation",
		"my-thoughts.md")); err != nil {
		t.Fatal("the owner's note was touched")
	}
	if _, err := os.Stat(filepath.Join(h.vault, "ideas.md")); err != nil {
		t.Fatal("the owner's rich note was touched")
	}
	// working files stay
	for _, f := range []string{"links_manifest.json", "_undo_last_batch.txt"} {
		if _, err := os.Stat(filepath.Join(h.vault, f)); err != nil {
			t.Fatalf("%s was touched", f)
		}
	}
}

func TestAStubWithFrontmatterIsNeverTouched(t *testing.T) {
	h := newHygieneCase(t)
	guarded := writeT(t, filepath.Join(h.vault, "Tools", "Automation",
		"guarded_Automation_python.md"),
		"---\nsource: \"https://github.com/o/guarded\"\n---\n")
	h.run()
	if _, err := os.Stat(guarded); err != nil {
		t.Fatal("a frontmatter'd stub was touched")
	}
}

func TestContentFreeNoteWithANonAppNameIsNeverTouched(t *testing.T) {
	h := newHygieneCase(t)
	guarded := writeT(t, filepath.Join(h.vault, "Tools", "Automation",
		"my-thoughts.md"), "# Just a heading I wrote myself\n")
	blank := writeT(t, filepath.Join(h.vault, "Tools", "Automation",
		"scratchpad.md"), "")
	h.run()
	if _, err := os.Stat(guarded); err != nil {
		t.Fatal("the owner's heading-only note was touched")
	}
	if _, err := os.Stat(blank); err != nil {
		t.Fatal("the owner's blank note was touched")
	}
}

func TestAStubWithBodyTextIsNeverTouched(t *testing.T) {
	h := newHygieneCase(t)
	guarded := writeT(t, filepath.Join(h.vault, "Tools", "Automation",
		"words.md"), "# words\n\na single line of real writing\n")
	h.run()
	if _, err := os.Stat(guarded); err != nil {
		t.Fatal("a stub with body text was touched")
	}
}

func TestTwoHeadingsIsWritingNotAStub(t *testing.T) {
	h := newHygieneCase(t)
	guarded := writeT(t, filepath.Join(h.vault, "two_Automation.md"),
		"# one\n# two\n")
	h.run()
	if _, err := os.Stat(guarded); err != nil {
		t.Fatal("a two-heading note was touched")
	}
}

func TestCleanVaultReportsCleanAndTouchesNothing(t *testing.T) {
	h := newHygieneCase(t)
	_, stubRoot, stubFolder, oldReport, oldSummary, _ := h.pollute()
	// remove the pollution by hand → only survivors remain
	if err := os.RemoveAll(filepath.Join(h.vault, "_moc")); err != nil {
		t.Fatal(err)
	}
	for _, f := range []string{
		filepath.Join(h.vault, "_index.md"), stubRoot, stubFolder,
		oldReport, oldSummary,
	} {
		if err := os.Remove(f); err != nil {
			t.Fatal(err)
		}
	}
	before := listNames(t, h.vault)
	logs := h.run()
	after := listNames(t, h.vault)
	if strings.Join(before, "|") != strings.Join(after, "|") {
		t.Fatalf("the clean pass touched the vault: %v → %v", before, after)
	}
	if !strings.Contains(logs.joined(), "clean") {
		t.Fatalf("the clean story missing:\n%s", logs.joined())
	}
}

func TestIdempotentSecondRunIsANoop(t *testing.T) {
	h := newHygieneCase(t)
	_, _, _, _, _, realNote := h.pollute()
	h.run()
	logs := h.run()
	if !strings.Contains(logs.joined(), "clean") {
		t.Fatalf("the second run is not clean:\n%s", logs.joined())
	}
	if _, err := os.Stat(realNote); err != nil {
		t.Fatal("the real note was touched on the second run")
	}
}

func TestNoVaultConfiguredIsASilentNoop(t *testing.T) {
	res := RunVaultHygiene("", nil, nil, nil) // must not panic
	if res.RetiredStubs != 0 {
		t.Fatalf("empty vault path retired %d", res.RetiredStubs)
	}
	res = RunVaultHygiene(filepath.Join(t.TempDir(), "nope"), nil, nil, nil)
	if res.RetiredStubs != 0 {
		t.Fatalf("missing vault retired %d", res.RetiredStubs)
	}
}

func TestTheLogTellsTheOwnerTheStory(t *testing.T) {
	h := newHygieneCase(t)
	h.pollute()
	logs := h.run()
	j := logs.joined()
	for _, want := range []string{
		"master index", "stray empty note", ".trash/empty-stubs", "legacy report",
	} {
		if !strings.Contains(j, want) {
			t.Fatalf("the story misses %q:\n%s", want, j)
		}
	}
}

func TestDryRunRehearsesWithoutTouchingTheVault(t *testing.T) {
	h := newHygieneCase(t)
	h.pollute()
	dryrun.Enable()
	RunVaultHygiene(h.vault, nil, nil, nil)
	dryrun.Disable()
	dryrun.Clear()
	// nothing actually moved
	for _, f := range []string{
		filepath.Join(h.vault, "_index.md"),
		filepath.Join(h.vault, "subtitle-translator_LLM-Tools_typescript.md"),
		filepath.Join(h.vault, "_processing_report_20260101_000001.md"),
	} {
		if _, err := os.Stat(f); err != nil {
			t.Fatalf("%s moved in a rehearsal", f)
		}
	}
}

func TestAHandwrittenIndexIsNeverRemoved(t *testing.T) {
	h := newHygieneCase(t)
	handwritten := writeT(t, filepath.Join(h.vault, "_index.md"),
		"# My own index\n- [[something]]\n")
	h.run()
	if _, err := os.Stat(handwritten); err != nil {
		t.Fatal("a handwritten index was removed")
	}
}

func TestAHandwrittenMocIsNeverRemoved(t *testing.T) {
	h := newHygieneCase(t)
	handwritten := writeT(t, filepath.Join(h.vault, "_moc", "mine.md"),
		"# My own MOC\n")
	h.run()
	if _, err := os.Stat(handwritten); err != nil {
		t.Fatal("a handwritten MOC was removed")
	}
}

// --- TestStubClassifier (test_vaulthygiene's pure table) --------------------------

func TestEmptyStringIsAStub(t *testing.T) {
	if !IsEmptyStub("") {
		t.Fatal("'' is not a stub")
	}
}

func TestWhitespaceOnlyIsAStub(t *testing.T) {
	if !IsEmptyStub("\n\n \t\n") {
		t.Fatal("whitespace is not a stub")
	}
}

func TestSingleHeadingIsAStub(t *testing.T) {
	if !IsEmptyStub("# a name\n\n") {
		t.Fatal("a single heading is not a stub")
	}
}

func TestHeadingWithoutSpaceMarkerIsAStub(t *testing.T) {
	if !IsEmptyStub("#\n") {
		t.Fatal("a bare '#' line is not a stub")
	}
}

func TestFrontmatterIsNeverAStub(t *testing.T) {
	if IsEmptyStub("---\nsource: x\n---\n") {
		t.Fatal("frontmatter judged a stub")
	}
}

func TestBodyTextIsNeverAStub(t *testing.T) {
	if IsEmptyStub("# t\n\nreal words\n") {
		t.Fatal("body text judged a stub")
	}
}

func TestListIsNeverAStub(t *testing.T) {
	if IsEmptyStub("# t\n\n- an item\n") {
		t.Fatal("a list judged a stub")
	}
}

func TestTwoHeadingsAreWriting(t *testing.T) {
	if IsEmptyStub("# one\n# two\n") {
		t.Fatal("two headings judged a stub")
	}
}

func TestLoneHashIsTreatedAsEmpty(t *testing.T) {
	if !IsEmptyStub("#") {
		t.Fatal("'#' alone is not content-free")
	}
}

// --- TestNoteNameShape ------------------------------------------------------------

func TestCanonicalRepoNoteShapeMatches(t *testing.T) {
	if !HasNoteNameShape("subtitle-translator_LLM-Tools_typescript.md") {
		t.Fatal("the canonical shape did not match")
	}
}

func TestTwoSegmentShapeMatches(t *testing.T) {
	if !HasNoteNameShape("agent_Automation.md") {
		t.Fatal("the two-segment shape did not match")
	}
}

func TestOwnerFilenameDoesNotMatch(t *testing.T) {
	for _, n := range []string{"my-thoughts.md", "ideas.md",
		"subtitle-translator.md"} {
		if HasNoteNameShape(n) {
			t.Fatalf("%q matched the app shape", n)
		}
	}
}

func TestRandomUnderscoreNameDoesNotMatch(t *testing.T) {
	if HasNoteNameShape("my_automaton_notes.md") {
		t.Fatal("an underscore name matched without a category segment")
	}
}

func TestNoUnderscoreDoesNotMatch(t *testing.T) {
	if HasNoteNameShape("Automation.md") {
		t.Fatal("a bare category name matched")
	}
}

// --- test_picklock's locked-bodyless law (T-46's own leg) --------------------------

func TestTheHygienePassNeverRetiresALockedNote(t *testing.T) {
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "vault")
	if err := os.MkdirAll(filepath.Join(vault, "Design"), 0o755); err != nil {
		t.Fatal(err)
	}
	lockedBodyless := writeT(t,
		filepath.Join(vault, "Design", "locked_LLM-Tools_ts.md"),
		"---\nmanaged_by: \"GitCurator\"\nlocked: true\n"+
			"---\n\n# just a heading\n")
	freeBodyless := writeT(t,
		filepath.Join(vault, "Design", "free_LLM-Tools_ts.md"),
		"---\nmanaged_by: \"GitCurator\"\n---\n\n# head\n")
	// both answer the bodyless fingerprint…
	if !IsAppOwnedBodyless(readT(t, lockedBodyless)) {
		t.Fatal("the locked note is not app-owned bodyless")
	}
	if !IsAppOwnedBodyless(readT(t, freeBodyless)) {
		t.Fatal("the free note is not app-owned bodyless")
	}
	// …but only one carries the lock
	if !IsLockedFrontmatter(readT(t, lockedBodyless)) {
		t.Fatal("the locked note's lock was not seen")
	}
	if IsLockedFrontmatter(readT(t, freeBodyless)) {
		t.Fatal("the free note was seen locked")
	}
	// the tags door and the fold
	if !IsLockedFrontmatter("---\ntags: [design, Lock]\n---\n\nx\n") {
		t.Fatal("the flow-tag lock door did not fire")
	}
	if IsLockedFrontmatter("---\ntags: [design]\n---\n\nx\n") {
		t.Fatal("a design tag fired the lock door")
	}
	if IsLockedFrontmatter("") {
		t.Fatal("empty content fired the lock door")
	}
	healed := RunVaultHygiene(vault, nil, nil, nil)
	if healed.RetiredStubs != 1 {
		t.Fatalf("retired %d; want 1 (the free twin only)", healed.RetiredStubs)
	}
	if _, err := os.Stat(lockedBodyless); err != nil {
		t.Fatal("the locked note was retired")
	}
	if _, err := os.Stat(freeBodyless); err == nil {
		t.Fatal("the free note stayed")
	}
}

// --- helpers ------------------------------------------------------------------------

func listNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		out = append(out, e.Name())
	}
	sort.Strings(out)
	return out
}

func rootMDNames(t *testing.T, vault string) []string {
	t.Helper()
	entries, err := os.ReadDir(vault)
	if err != nil {
		t.Fatal(err)
	}
	out := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".md") {
			out = append(out, e.Name())
		}
	}
	sort.Strings(out)
	return out
}
