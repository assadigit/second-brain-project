// directorynote_test.go — T-41: the law-sweep + directory batteries
// ported from the reference's own test legs (tests/test_lawfix.py —
// v0.28.0 THE LAW + the Website Directory, and tests/test_socialomit.py's
// sweep leg), plus the walk-order law and the production-constructor
// integration (the seam that was a no-op until this task).
package websitepipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/taxonomy"
	"github.com/assadigit/second-brain-project/internal/webfetch"
)

// lawNote mirrors test_lawfix.py's _write_note: a v2-shaped website
// note (the pipeline's own format).
func lawNote(t *testing.T, path, url, category, subcategory, pricing,
	title, tldr string, managed bool, fetchStatus string) {
	t.Helper()
	managedBy := "gitcurator"
	if !managed {
		managedBy = "human"
	}
	does := tldr
	if does == "" {
		does = title
	}
	fm := []string{
		"---",
		fmt.Sprintf("source: %q", url),
		"aliases: []",
		"tags: []",
		fmt.Sprintf("category: %q", category),
		fmt.Sprintf("subcategory: %q", subcategory),
		fmt.Sprintf("fetch_status: %q", fetchStatus),
		fmt.Sprintf("pricing: %q", pricing),
		`login_required: "unknown"`,
		"date_processed: 2026-10-01",
		fmt.Sprintf("managed_by: %q", managedBy),
		`schema_version: "1"`,
		`prompt_version: "web-v2"`,
		"---",
		"",
		"> [!info] Managed by GitCurator — machine-written note.",
		"",
		fmt.Sprintf("# %s", title),
		"",
		fmt.Sprintf("> **TL;DR:** %s", tldr),
		"",
		"## What it does",
		"",
		fmt.Sprintf("- %s", does),
		"",
		"## Best used for",
		"Use when you need to…",
		"",
		"## Similar tools",
		"—",
		"",
		"---",
		fmt.Sprintf("*Source: [%s](%s)*", url, url),
		"",
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(strings.Join(fm, "\n")), 0o644); err != nil {
		t.Fatal(err)
	}
}

// lawCase is a temp websites vault + captured logs (test_lawfix's
// _VaultCase).
type lawCase struct {
	t     *testing.T
	tmp   string
	vault string
	logs  []logLine
}

func newLawCase(t *testing.T) *lawCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	c := &lawCase{t: t, tmp: t.TempDir()}
	c.vault = filepath.Join(c.tmp, "webvault")
	if err := os.MkdirAll(c.vault, 0o755); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		dryrun.Disable()
		dryrun.Clear()
	})
	return c
}

func (c *lawCase) logFn(msg, level string) {
	c.logs = append(c.logs, logLine{level: level, msg: msg})
}

func (c *lawCase) allLogs() string {
	var msgs []string
	for _, l := range c.logs {
		msgs = append(msgs, l.msg)
	}
	return strings.Join(msgs, "\n")
}

func (c *lawCase) note(rel string) string {
	return filepath.Join(c.vault, filepath.FromSlash(rel))
}

func (c *lawCase) quarantine() string {
	return filepath.Join(c.vault, ".trash", "banned-domains")
}

func (c *lawCase) law() []string {
	return links.BlockedDomainsFromConfig(map[string]any{})
}

// --- 1. The sweep — the legacy x_com_i_status_* pile leaves the vault ---

func TestSweepMovesTheLegacyReviewPile(t *testing.T) {
	c := newLawCase(t)
	pile := []string{}
	for i, sid := range []string{"2011934783064826", "2011943712111707", "2011947543668256"} {
		p := c.note(fmt.Sprintf("_review/x_com_i_status_%s.md", sid))
		lawNote(t, p, fmt.Sprintf("https://x.com/i/status/%s", sid), "", "", "unknown", fmt.Sprintf("tweet %d", i), "", true, "failed")
		pile = append(pile, p)
	}
	lawNote(t, c.note("Design/Assets_Resources/Real_Site.md"), "https://reverseui.com/", "Design", "Assets & Resources", "unknown", "Real Site", "", true, "full")

	report := SweepBannedNotes(c.vault, c.law(), c.logFn)

	if len(report.Moved) != 3 {
		t.Fatalf("moved = %d, want 3 (report %+v)", len(report.Moved), report)
	}
	for _, p := range pile {
		if _, err := os.Stat(p); !os.IsNotExist(err) {
			t.Errorf("%s still exists", p)
		}
		q := filepath.Join(c.quarantine(), filepath.Base(p))
		if _, err := os.Stat(q); err != nil {
			t.Errorf("%s missing from quarantine: %v", q, err)
		}
	}
	// the real site is untouched
	if _, err := os.Stat(c.note("Design/Assets_Resources/Real_Site.md")); err != nil {
		t.Errorf("real site touched: %v", err)
	}
	if !strings.Contains(c.allLogs(), "Law sweep") {
		t.Errorf("log missing the Law sweep line: %q", c.allLogs())
	}
}

func TestSweepMovesRealCategoryNotesOnLawDomains(t *testing.T) {
	c := newLawCase(t)
	for rel, url := range map[string]string{
		"Design/Tools/tco_note.md":       "https://t.co/abc",
		"AI_Tools_(General)/hf_note.md":  "https://huggingface.co/org/model",
		"Random__Curiosities/ig_note.md": "https://www.instagram.com/p/xyz/",
	} {
		lawNote(t, c.note(rel), url, "", "", "unknown", "x", "", true, "full")
	}
	report := SweepBannedNotes(c.vault, c.law(), c.logFn)
	if len(report.Moved) != 3 {
		t.Fatalf("moved = %d, want 3 (report %+v)", len(report.Moved), report)
	}
	if report.KeptHandwritten != 0 {
		t.Errorf("kept_handwritten = %d, want 0", report.KeptHandwritten)
	}
}

func TestSweepKeepsHandwrittenNotesAndCountsThem(t *testing.T) {
	c := newLawCase(t)
	lawNote(t, c.note("Notes/my_own_x_thread.md"), "https://x.com/thread", "", "", "unknown", "My notes", "", false, "full")
	report := SweepBannedNotes(c.vault, c.law(), c.logFn)
	if len(report.Moved) != 0 {
		t.Fatalf("moved = %v, want none", report.Moved)
	}
	if report.KeptHandwritten != 1 {
		t.Fatalf("kept_handwritten = %d, want 1", report.KeptHandwritten)
	}
	if _, err := os.Stat(c.note("Notes/my_own_x_thread.md")); err != nil {
		t.Errorf("the owner's own writing was touched: %v", err)
	}
	if !strings.Contains(c.allLogs(), "hand-written") {
		t.Errorf("log missing the hand-written line: %q", c.allLogs())
	}
}

func TestSweepQuarantinesBannedPlatformTables(t *testing.T) {
	c := newLawCase(t)
	inbox := c.note("_inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	for _, fname := range []string{"x_twitter_links.md", "youtube_links.md",
		"linkedin_links.md", "huggingface_links.md", "reddit_links.md",
		"medium_links.md", "other_links.md"} {
		content := fmt.Sprintf("# %s\n\n| - | 2026-10-05 | https://%s.example/x | x | Bot | unreviewed | |\n",
			fname, strings.SplitN(fname, "_", 2)[0])
		if err := os.WriteFile(filepath.Join(inbox, fname), []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	report := SweepBannedNotes(c.vault, c.law(), c.logFn)

	var moved []string
	for _, tb := range report.MovedTables {
		moved = append(moved, tb.Platform)
	}
	sort.Strings(moved)
	want := []string{"huggingface", "linkedin", "x_twitter", "youtube"}
	if len(moved) != len(want) {
		t.Fatalf("moved_tables = %v, want %v", moved, want)
	}
	for i := range want {
		if moved[i] != want[i] {
			t.Fatalf("moved_tables = %v, want %v", moved, want)
		}
	}
	for _, platform := range []string{"x_twitter", "youtube", "linkedin", "huggingface"} {
		if _, err := os.Stat(filepath.Join(inbox, platform+"_links.md")); !os.IsNotExist(err) {
			t.Errorf("_inbox/%s_links.md still exists", platform)
		}
		if _, err := os.Stat(filepath.Join(c.quarantine(), platform+"_links.md")); err != nil {
			t.Errorf("%s_links.md missing from quarantine: %v", platform, err)
		}
	}
	for _, fname := range []string{"reddit_links.md", "medium_links.md", "other_links.md"} {
		if _, err := os.Stat(filepath.Join(inbox, fname)); err != nil {
			t.Errorf("%s was touched: %v", fname, err)
		}
	}
	if !strings.Contains(c.allLogs(), "never collected") {
		t.Errorf("log missing the never-collected line: %q", c.allLogs())
	}
}

func TestSweepKeepsAPartiallyBannedPlatformsTable(t *testing.T) {
	c := newLawCase(t)
	inbox := c.note("_inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	table := filepath.Join(inbox, "reddit_links.md")
	if err := os.WriteFile(table, []byte("# reddit\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	report := SweepBannedNotes(c.vault, []string{"x.com", "t.co", "twitter.com", "youtube.com"}, c.logFn)
	if len(report.MovedTables) != 0 {
		t.Fatalf("moved_tables = %v, want none", report.MovedTables)
	}
	if _, err := os.Stat(table); err != nil {
		t.Errorf("reddit table touched: %v", err)
	}
}

func TestSweepIsIdempotent(t *testing.T) {
	c := newLawCase(t)
	lawNote(t, c.note("_review/x_com_1.md"), "https://x.com/1", "", "", "unknown", "x", "", true, "full")
	SweepBannedNotes(c.vault, c.law(), c.logFn)
	second := SweepBannedNotes(c.vault, c.law(), c.logFn)
	if len(second.Moved) != 0 {
		t.Fatalf("second sweep moved %v, want none", second.Moved)
	}
	entries, err := os.ReadDir(c.quarantine())
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 1 {
		t.Fatalf("quarantine holds %d files, want 1", len(entries))
	}
}

func TestSweepRespectsDryRun(t *testing.T) {
	c := newLawCase(t)
	lawNote(t, c.note("_review/x_com_1.md"), "https://x.com/1", "", "", "unknown", "x", "", true, "full")
	dryrun.Enable()
	dryrun.Clear()
	defer func() {
		dryrun.Disable()
		dryrun.Clear()
	}()
	report := SweepBannedNotes(c.vault, c.law(), c.logFn)
	if len(report.Moved) != 1 {
		t.Fatalf("moved = %d, want 1 (the rehearsal records)", len(report.Moved))
	}
	// a rehearsal records, never performs: the file survives
	if _, err := os.Stat(c.note("_review/x_com_1.md")); err != nil {
		t.Errorf("dry-run performed the move: %v", err)
	}
	if _, err := os.Stat(c.quarantine()); !os.IsNotExist(err) {
		t.Errorf("quarantine dir materialized under dry-run")
	}
	if dryrun.EntryCount() <= 0 {
		t.Errorf("dry-run recorded nothing")
	}
}

func TestSweepNameCollisionGetsUniqueNames(t *testing.T) {
	c := newLawCase(t)
	lawNote(t, c.note("_review/x_com_1.md"), "https://x.com/1", "", "", "unknown", "x", "", true, "full")
	lawNote(t, c.note("Design/x_com_1.md"), "https://x.com/1b", "", "", "unknown", "x", "", true, "full")
	report := SweepBannedNotes(c.vault, c.law(), c.logFn)
	if len(report.Moved) != 2 {
		t.Fatalf("moved = %d, want 2", len(report.Moved))
	}
	entries, err := os.ReadDir(c.quarantine())
	if err != nil {
		t.Fatal(err)
	}
	names := []string{}
	for _, e := range entries {
		names = append(names, e.Name())
	}
	sort.Strings(names)
	if len(names) != 2 || names[0] == names[1] {
		t.Fatalf("quarantine names %v, want two distinct", names)
	}
}

func TestSweepOnMissingVaultIsANoOp(t *testing.T) {
	c := newLawCase(t)
	report := SweepBannedNotes(filepath.Join(c.tmp, "nope"), []string{"x.com"}, c.logFn)
	if len(report.Moved) != 0 || report.KeptHandwritten != 0 || report.Scanned != 0 || len(report.MovedTables) != 0 {
		t.Fatalf("report = %+v, want the zero report", report)
	}
}

func TestSweepEmptyBlockedListIsANoOp(t *testing.T) {
	c := newLawCase(t)
	lawNote(t, c.note("_review/x_com_1.md"), "https://x.com/1", "", "", "unknown", "x", "", true, "full")
	report := SweepBannedNotes(c.vault, nil, c.logFn)
	if len(report.Moved) != 0 || report.Scanned != 0 {
		t.Fatalf("report = %+v, want nothing scanned (empty ban list)", report)
	}
	if _, err := os.Stat(c.note("_review/x_com_1.md")); err != nil {
		t.Errorf("file touched under an empty ban list: %v", err)
	}
}

// --- the walk's own law: files before subdirs, both skip sets ---------

func TestIterNoteFilesOrderAndSkips(t *testing.T) {
	c := newLawCase(t)
	w := c.vault
	mkNote := func(rel string) {
		t.Helper()
		if err := os.MkdirAll(filepath.Dir(filepath.Join(w, filepath.FromSlash(rel))), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(w, filepath.FromSlash(rel)), []byte("---\nx: 1\n---\n"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, rel := range []string{"zNote.md", "m-note.md", "a-dir/inner.md",
		"a-dir/z-inner.md", "b-dir/n1.md", "b-dir/N2.MD", "_review/r.md",
		"_missing/m.md", "_moc/m.md", "attachments/a.md", "_inbox/i.md",
		".obsidian/o.md", ".trash/t.md", ".git/g.md", ".hidden.md",
		"sub/_inbox/deep.md", "sub/deep.md", DirectoryFilename} {
		mkNote(rel)
	}
	// a file that sorts lexically AFTER a directory: os.walk yields the
	// root's files before ANY subdirectory's files (WalkDir interleaves).
	collect := func(forSweep bool) []string {
		var out []string
		iterNoteFiles(w, forSweep, func(path string) {
			rel, err := filepath.Rel(w, path)
			if err != nil {
				t.Fatal(err)
			}
			out = append(out, filepath.ToSlash(rel))
		})
		return out
	}
	got := collect(false)
	want := []string{".hidden.md", "m-note.md", "zNote.md",
		"a-dir/inner.md", "a-dir/z-inner.md", "b-dir/N2.MD", "b-dir/n1.md",
		"sub/deep.md"}
	if len(got) != len(want) {
		t.Fatalf("directory walk = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("directory walk[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
	gotSweep := collect(true)
	wantSweep := []string{".hidden.md", "m-note.md", "zNote.md",
		"_missing/m.md", "_moc/m.md", "_review/r.md",
		"a-dir/inner.md", "a-dir/z-inner.md", "attachments/a.md",
		"b-dir/N2.MD", "b-dir/n1.md", "sub/deep.md"}
	if len(gotSweep) != len(wantSweep) {
		t.Fatalf("sweep walk = %v, want %v", gotSweep, wantSweep)
	}
	for i := range wantSweep {
		if gotSweep[i] != wantSweep[i] {
			t.Fatalf("sweep walk[%d] = %q, want %q (full: %v)", i, gotSweep[i], wantSweep[i], gotSweep)
		}
	}
}

func TestIterNoteFilesSkipsSymlinkedDirs(t *testing.T) {
	c := newLawCase(t)
	w := c.vault
	if err := os.MkdirAll(filepath.Join(w, "a-dir"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w, "a-dir", "inner.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(w, "real.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	// a symlinked DIRECTORY named *.md: rides Python's dirs list, never
	// descended, never yielded; a symlinked FILE is yielded like any file.
	if err := os.Symlink(filepath.Join(w, "a-dir"), filepath.Join(w, "dirlink.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	if err := os.Symlink(filepath.Join(w, "real.md"), filepath.Join(w, "filelink.md")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}
	var got []string
	iterNoteFiles(w, false, func(path string) {
		rel, _ := filepath.Rel(w, path)
		got = append(got, filepath.ToSlash(rel))
	})
	want := []string{"filelink.md", "real.md", "a-dir/inner.md"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("walk = %v, want %v", got, want)
	}
}

// --- 2. The Website Directory — consolidated, categorized, clickable ---

func (c *lawCase) seedDirectory(t *testing.T) *taxonomy.Taxonomy {
	t.Helper()
	lawNote(t, c.note("Design/Assets_Resources/Reverse_UI.md"), "https://reverseui.com/",
		"Design", "Assets & Resources", "free", "Reverse UI",
		"68 animated web UI components, copy-paste Tailwind", true, "full")
	lawNote(t, c.note("Design/Assets_Resources/Coolors.md"), "https://coolors.co/",
		"Design", "Assets & Resources", "freemium", "Coolors",
		"Color palette maker with generators and export", true, "full")
	lawNote(t, c.note("Security_Privacy_Tools/HaveIBeenPwned.md"), "https://haveibeenpwned.com/",
		"Security & Privacy Tools", "", "paid", "HaveIBeenPwned",
		"Breach checker for your accounts", true, "full")
	lawNote(t, c.note("Random__Curiosities/Curiosity.md"), "https://curiosity.example/",
		"Random / Curiosities", "", "unknown", "A Curiosity",
		"Odd but lovely", true, "full")
	// NOT in the directory: review placeholder, hand-written, law note
	lawNote(t, c.note("_review/Waiting_Site.md"), "https://waitingsite.com/",
		"", "", "unknown", "Waiting Site", "", true, "failed")
	lawNote(t, c.note("Notes/my_own.md"), "https://myblog.example/",
		"", "", "unknown", "My own writing", "", false, "full")
	lawNote(t, c.note("Design/Assets_Resources/x_note.md"), "https://x.com/thread",
		"Design", "Assets & Resources", "unknown", "An X thread", "", true, "full")
	tax, err := taxonomy.LoadTaxonomy("../taxonomy/testdata/website-library-categories.md")
	if err != nil {
		t.Fatal(err)
	}
	return tax
}

func TestDirectoryGroupsByTaxonomyOrderNotAlphabetical(t *testing.T) {
	c := newLawCase(t)
	tax := c.seedDirectory(t)
	out := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn)
	if out == nil {
		t.Fatal("out = nil, want a report")
	}
	text := readFileT(t, out.Path)
	// 'Security & Privacy Tools' (taxonomy position 9) must come BEFORE
	// 'Random / Curiosities' (taxonomy position 14, last) — alphabetical
	// order would flip them.
	iSec := strings.Index(text, "## Security & Privacy Tools")
	iRnd := strings.Index(text, "## Random / Curiosities")
	iDes := strings.Index(text, "## Design")
	if iDes < 0 || iSec < 0 || iRnd < 0 {
		t.Fatalf("missing headings: %d %d %d\n%s", iDes, iSec, iRnd, text)
	}
	if !(iDes < iSec && iSec < iRnd) {
		t.Fatalf("taxonomy order broken: Design %d, Security %d, Random %d", iDes, iSec, iRnd)
	}
}

func TestDirectoryEntriesAreClickableWithDescription(t *testing.T) {
	c := newLawCase(t)
	tax := c.seedDirectory(t)
	out := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn)
	text := readFileT(t, out.Path)
	for _, want := range []string{
		"[[Design/Assets_Resources/Reverse_UI|Reverse UI]] — 68 animated web UI components, copy-paste Tailwind — free — [↗](https://reverseui.com/)",
		"[[Design/Assets_Resources/Coolors|Coolors]] — Color palette maker with generators and export — freemium — [↗](https://coolors.co/)",
		"### Assets & Resources (2)",
		"## Design (2)",
		"paid",
		"[↗](https://curiosity.example/)",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("directory missing %q:\n%s", want, text)
		}
	}
}

func TestDirectoryExcludesReviewHandwrittenAndLawNotes(t *testing.T) {
	c := newLawCase(t)
	tax := c.seedDirectory(t)
	out := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn)
	if out == nil {
		t.Fatal("out = nil")
	}
	text := readFileT(t, out.Path)
	for _, banned := range []string{"Waiting Site", "My own writing", "An X thread"} {
		if strings.Contains(text, banned) {
			t.Errorf("directory lists %q (it must not):\n%s", banned, text)
		}
	}
	// but the review count is reported — nothing silently missing
	if !strings.Contains(text, "> [!todo]") || !strings.Contains(text, "1 link(s) still waiting") {
		t.Errorf("review callout missing:\n%s", text)
	}
	if out.Sites != 4 {
		t.Errorf("sites = %d, want 4", out.Sites)
	}
	if out.Review != 1 {
		t.Errorf("review = %d, want 1", out.Review)
	}
}

func TestDirectoryRegeneratesIdentically(t *testing.T) {
	c := newLawCase(t)
	tax := c.seedDirectory(t)
	one := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn)
	text1 := readFileT(t, one.Path)
	two := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn)
	text2 := readFileT(t, two.Path)
	if text1 != text2 {
		t.Fatalf("regeneration not identical:\n--- one ---\n%s\n--- two ---\n%s", text1, text2)
	}
}

func TestDirectoryOnEmptyVaultReturnsNone(t *testing.T) {
	c := newLawCase(t)
	tax, err := taxonomy.LoadTaxonomy("../taxonomy/testdata/website-library-categories.md")
	if err != nil {
		t.Fatal(err)
	}
	if out := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn); out != nil {
		t.Fatalf("out = %+v, want nil", out)
	}
}

func TestDirectoryLongTldrIsTrimmedAtAWordBoundary(t *testing.T) {
	c := newLawCase(t)
	tax, err := taxonomy.LoadTaxonomy("../taxonomy/testdata/website-library-categories.md")
	if err != nil {
		t.Fatal(err)
	}
	lawNote(t, c.note("Design/Tools/Wordy.md"), "https://wordy.example/",
		"Design", "", "unknown", "Wordy", strings.Repeat("word ", 60), true, "full")
	out := BuildWebsiteDirectory(c.vault, tax, nil, c.logFn)
	text := readFileT(t, out.Path)
	var line string
	for _, ln := range strings.Split(text, "\n") {
		if strings.Contains(ln, "Wordy") {
			line = ln
			break
		}
	}
	if line == "" {
		t.Fatalf("Wordy line missing:\n%s", text)
	}
	if !strings.Contains(line, "…") {
		t.Errorf("line not trimmed with an ellipsis: %q", line)
	}
	if len([]rune(line)) >= 260 {
		t.Errorf("line too long (%d runes): %q", len([]rune(line)), line)
	}
}

// --- 3. The pipeline constructor runs the sweep (production only) -----

func TestProductionConstructorSweepsAndPurges(t *testing.T) {
	c := newLawCase(t)
	dryrun.Disable()
	dryrun.Clear()
	db, err := NewWebsiteStateDB(filepath.Join(c.tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tax, err := taxonomy.LoadTaxonomy("../taxonomy/testdata/website-library-categories.md")
	if err != nil {
		t.Fatal(err)
	}
	// (a) DB-known legacy placeholder — purged + deleted
	review := c.note("_review/x_com_i_status_2011934783064826.md")
	lawNote(t, review, "https://x.com/i/status/2011934783064826", "", "", "unknown", "tweet", "", true, "failed")
	if err := db.EnqueueRetry("https://x.com/i/status/2011934783064826", "refused"); err != nil {
		t.Fatal(err)
	}
	if err := db.MarkProcessed("https://x.com/i/status/2011934783064826", review, "", "", "failed"); err != nil {
		t.Fatal(err)
	}
	// (b) DB-unknown legacy note (cache was cleared once) — swept
	orphan := c.note("Design/Assets_Resources/hf_note.md")
	lawNote(t, orphan, "https://huggingface.co/org/model", "Design", "Assets & Resources", "unknown", "An HF model", "", true, "full")

	cfg := map[string]any{
		"website_vault_path": c.vault,
		"web_domain_delay_s": 0,
	}
	_, err = NewPipeline(cfg,
		func(messages []llmclient.Message, task string) string { return "{}" },
		func(u string) any { return nil },
		db, c.logFn, tax, nil, nil,
		nil, // fetchFn = nil → PRODUCTION path (the sweep runs)
		nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	// (a) gone — purged + deleted by the bookkeeping pass
	if _, err := os.Stat(review); !os.IsNotExist(err) {
		t.Errorf("DB-known placeholder still on disk: %v", err)
	}
	// (b) gone — moved to the quarantine by the law sweep
	if _, err := os.Stat(orphan); !os.IsNotExist(err) {
		t.Errorf("orphan still on disk: %v", err)
	}
	if _, err := os.Stat(filepath.Join(c.quarantine(), "hf_note.md")); err != nil {
		t.Errorf("hf_note.md missing from quarantine: %v", err)
	}
	// …the bookkeeping agrees on both…
	if ok, err := db.IsDismissed("https://x.com/i/status/2011934783064826"); err != nil || !ok {
		t.Errorf("DB-known URL not dismissed: %v %v", ok, err)
	}
	if row, err := db.RetryRow("https://x.com/i/status/2011934783064826"); err != nil || row != nil {
		t.Errorf("retry row survived: %+v %v", row, err)
	}
	// …and the run log says so in plain words.
	if !strings.Contains(c.allLogs(), "Law sweep") {
		t.Errorf("log missing the Law sweep line: %q", c.allLogs())
	}
	if !strings.Contains(c.allLogs(), "purged 1 queued retry") {
		t.Errorf("log missing the purge line: %q", c.allLogs())
	}
}

func TestHermeticConstructorDoesNotSweep(t *testing.T) {
	c := newLawCase(t)
	dryrun.Disable()
	dryrun.Clear()
	db, err := NewWebsiteStateDB(filepath.Join(c.tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	tax, err := taxonomy.LoadTaxonomy("../taxonomy/testdata/website-library-categories.md")
	if err != nil {
		t.Fatal(err)
	}
	review := c.note("_review/x_com_1.md")
	lawNote(t, review, "https://x.com/1", "", "", "unknown", "x", "", true, "full")

	cfg := map[string]any{
		"website_vault_path": c.vault,
		"web_domain_delay_s": 0,
	}
	_, err = NewPipeline(cfg,
		func(messages []llmclient.Message, task string) string { return "{}" },
		func(u string) any { return nil },
		db, c.logFn, tax, nil, nil,
		func(rawURL string, opts *webfetch.Options) *webfetch.FetchResult { // injected = hermetic
			panic("never fetched")
		},
		nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(review); err != nil {
		t.Errorf("hermetic constructor touched the vault: %v", err)
	}
	if _, err := os.Stat(c.quarantine()); !os.IsNotExist(err) {
		t.Errorf("quarantine materialized under a hermetic run")
	}
}
