// goodrepos_test.go — T-48: the ported selections of the reference's
// own suite (app/tests/test_goodrepos.py — the whole thing):
//
//   - TestPublish (§1-2: the staging tree, the README structure, the
//     entry-line format, status()'s scan-only snapshot, the v0.83.0
//     empty-name skip, the dry-run that runs NO git);
//   - TestAnchors (§3: GitHub's anchor math — the emoji's leading
//     hyphen, the counts' trailing digits, every Contents link
//     resolving);
//   - TestOrdering (§4: stars-desc entries, alphabetical categories,
//     the display-name order that beats the emoji code points);
//   - TestNoteParsing (§5-6: the TL;DR + "What is it?" fallback, the
//     display-name ladder url → aliases → filename, the frontmatter
//     edge cases, string stars);
//   - TestLifecycle (§7-8: publish-twice skips when unchanged, the
//     republish after a change, the disabled guard, the missing-vault
//     error, the empty-vault skip);
//   - TestConfigBridge (§9-10: publish_from_config's key wiring and
//     the disabled short-circuit that never touches the disk);
//   - TestVaultHygiene (§11: the root extras copied, the machine
//     state excluded, root notes never categories);
//   - TestHelpers (§12: format_stars, the emoji map, display_name,
//     the result dict contract);
//   - TestCreationGuardV0830 (the v0.83 creation guard — 404 never
//     creates without opt-in, the private-create payload POSTed once,
//     the empty-name NOT-CONFIGURED door with the zero-calls proof,
//     the bridge's no-default law);
//   - TestPublishFlow (the acceptance's fake-GitHub publish flow:
//     LOCAL BARE REPOS as the remote through the tokenURLFn seam —
//     the full cycle incl. the second-publish skip, the push-failure
//     that keeps the commit with the token REDACTED, the FETCH_HEAD
//     scrub through a real fetch).
//
// TEMP REPOS ONLY (never the owner's vault); NO real network (the
// GitHub API surface is the instance-seam fake exactly like the
// reference's own tests; the "remote" is a local bare repo).
package goodrepos

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// Fixtures (test_goodrepos.py's NOTE_TMPL / note_text / simple_note)
// ---------------------------------------------------------------------------

const noteTmpl = `---
source: "{source}"
aliases:
  - {alias}
  - {alias}/{repo}
tags: [{tags}]
category: "{category}"
stars: {stars}
org: ""
primary_language: {lang}
languages: [{lang}]
credibility_score: 78/100
date_processed: 2026-09-16
---

# {repo}

{tldr_line}

**` + "`{alias}/{repo}`" + `** · ⭐ {stars} · 🔧 {lang}

## What is it?
{what}
`

func noteText(source, repo, alias, category string, stars any, lang, tldr, tags, what string) string {
	tldrLine := ""
	if tldr != "" {
		tldrLine = "> **TL;DR:** " + tldr
	}
	return strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
		strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(
			strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(noteTmpl,
				"{source}", source),
				"{alias}", alias),
				"{repo}", repo),
			"{category}", category),
			"{stars}", pyStars(stars)),
			"{lang}", lang),
		"{tldr_line}", tldrLine),
		"{tags}", tags),
		"{what}", what)
}

func pyStars(v any) string {
	switch t := v.(type) {
	case int:
		return itoaGR(t)
	case string:
		return t
	default:
		return ""
	}
}

func itoaGR(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		b[i] = '-'
	}
	return string(b[i:])
}

type simpleNoteOpts struct {
	alias    string
	category string
	lang     string
	tags     string
	what     string
	stars    any
	tldr     *string // nil = Python None
}

func simpleNote(name string, opts *simpleNoteOpts) string {
	if opts == nil {
		opts = &simpleNoteOpts{}
	}
	if opts.alias == "" {
		opts.alias = name
	}
	if opts.category == "" {
		opts.category = "Agents"
	}
	if opts.lang == "" {
		opts.lang = "Python"
	}
	if opts.tags == "" {
		opts.tags = "alpha, beta"
	}
	if opts.what == "" {
		opts.what = "Fallback description sentence."
	}
	if opts.stars == nil {
		opts.stars = 100
	}
	tldr := "Does one thing well."
	if opts.tldr != nil {
		tldr = *opts.tldr
	}
	return noteText("https://github.com/"+name+"/"+name, name, opts.alias,
		opts.category, opts.stars, opts.lang, tldr, opts.tags, opts.what)
}

// grBase mirrors GoodReposBase: tiny in-memory vault + publisher
// factory (no network, no remotes).
type grBase struct {
	vault string
	t     *testing.T
}

func newGRBase(t *testing.T) *grBase {
	t.Helper()
	return &grBase{vault: t.TempDir(), t: t}
}

func (b *grBase) write(rel, text string) string {
	b.t.Helper()
	path := filepath.Join(b.vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		b.t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		b.t.Fatal(err)
	}
	return path
}

type pubOpts struct {
	repoName    string
	token       string
	allowCreate bool
	enabled     *bool // nil = true (the zero value must not disable)
	autoPush    *bool
}

// publisher mirrors GoodReposBase.publisher: auto_push=False ALWAYS
// (never push in the unit battery) and repo_name="good-repos".
func (b *grBase) publisher(opts *pubOpts) *GoodRepos {
	b.t.Helper()
	if opts == nil {
		opts = &pubOpts{}
	}
	repoName := opts.repoName
	if repoName == "" {
		repoName = "good-repos"
	}
	autoPush := false
	if opts.autoPush != nil {
		autoPush = *opts.autoPush
	}
	enabled := true
	if opts.enabled != nil {
		enabled = *opts.enabled
	}
	pub := NewGoodRepos(b.vault, opts.token, repoName, autoPush, enabled,
		0, nil, opts.allowCreate)
	b.t.Cleanup(func() { dropStaging(pub) })
	return pub
}

func dropStaging(pub *GoodRepos) {
	if pub.staging != "" {
		_ = os.RemoveAll(pub.staging)
		unregisterStaging(pub.staging)
	}
}

// defaultVault mirrors default_vault: 3 curated notes across 2
// categories + the two root extras.
func (b *grBase) defaultVault() {
	b.write("AI-Domain/Agents/aaa_Agents_x.md",
		simpleNote("aaa", &simpleNoteOpts{stars: 100, tldr: strPtr("First agent summary.")}))
	b.write("AI-Domain/Agents/bbb_Agents_y.md",
		simpleNote("bbb", &simpleNoteOpts{stars: 900, tldr: strPtr("Second agent summary.")}))
	b.write("Tools/Automation/ccc_Automation_z.md",
		simpleNote("ccc", &simpleNoteOpts{stars: 42, tldr: strPtr("Automation summary."),
			category: "Automation"}))
	b.write("_index.md", "# Index\n")
	b.write("links_manifest.json", `{"generated": "2026-09-16",
 "links": [{"url": "https://github.com/aaa/aaa", "status": "processed"}]}`)
}

func strPtr(s string) *string { return &s }

func boolPtr(v bool) *bool { return &v }

func (b *grBase) readme(result *PublishResult) string {
	b.t.Helper()
	data, err := os.ReadFile(filepath.Join(result.StagingDir, "README.md"))
	if err != nil {
		b.t.Fatal(err)
	}
	return string(data)
}

// ---------------------------------------------------------------------------
// 1-2. Scan, publish, staging tree, README structure (TestPublish)
// ---------------------------------------------------------------------------

func TestPublish_BasicTree(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	result := b.publisher(nil).Publish(nil, false)
	if !result.OK() {
		t.Fatalf("ok: %s", result.Describe())
	}
	if !result.Published {
		t.Fatal("published")
	}
	if result.Entries != 3 || result.Categories != 2 {
		t.Fatalf("entries/categories: %d/%d", result.Entries, result.Categories)
	}
	if result.Pushed { // no token -> local commit only
		t.Fatal("pushed without a token")
	}
	if result.CommitSha == "" {
		t.Fatal("commit_sha missing")
	}
	for _, rel := range []string{
		"AI-Domain/Agents/aaa_Agents_x.md",
		"AI-Domain/Agents/bbb_Agents_y.md",
		"Tools/Automation/ccc_Automation_z.md",
		"_index.md", "links_manifest.json", "README.md",
	} {
		if !isFile(filepath.Join(result.StagingDir, rel)) {
			t.Errorf("missing %s", rel)
		}
	}
}

func TestPublish_ReadmeStructure(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	readme := b.readme(b.publisher(nil).Publish(nil, false))
	if !strings.Contains(readme, "# 🗂️ Good Repos") {
		t.Error("title missing")
	}
	if !regexp.MustCompile(`📊 \*\*3 repos\*\* · 🗂️ \*\*2 categories\*\* · 🕰️ Updated \d{4}-\d{2}-\d{2}`).MatchString(readme) {
		t.Error("stats line missing")
	}
	if !strings.Contains(readme, "## 📇 Contents") {
		t.Error("contents missing")
	}
	if !strings.Contains(readme, "[GitCurator](https://github.com/assadigit/GitCurator)") {
		t.Error("GitCurator link missing")
	}
	if !strings.Contains(readme, "[aaa/aaa](https://github.com/aaa/aaa)") {
		t.Error("aaa entry link missing")
	}
	if !strings.Contains(readme, "🧠") { // Agents emoji
		t.Error("Agents emoji missing")
	}
	if !strings.Contains(readme, "🛠️") { // Tools emoji
		t.Error("Tools emoji missing")
	}
	if !strings.Contains(readme, "regenerated automatically after every GitCurator curation run") {
		t.Error("footer missing")
	}
}

func TestPublish_EntryLineFormat(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	readme := b.readme(b.publisher(nil).Publish(nil, false))
	// - 📦 [owner/repo](url) — TL;DR · ⭐ n · 🔧 Lang · `tag` `tag`
	want := regexp.MustCompile(
		"- 📦 \\[bbb/bbb\\]\\(https://github\\.com/bbb/bbb\\) — Second agent summary\\. " +
			"· ⭐ 900 · 🔧 Python · `alpha` `beta`")
	if !want.MatchString(readme) {
		t.Errorf("entry line format:\n%s", readme)
	}
}

func TestPublish_StatusScanOnly(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	info := b.publisher(nil).Status()
	if v, _ := info.Get("entries"); v.(int) != 3 {
		t.Errorf("entries: %v", v)
	}
	if v, _ := info.Get("categories"); v.(int) != 2 {
		t.Errorf("categories: %v", v)
	}
	// tree: {"AI-Domain/Agents": 2, "Tools/Automation": 1} (sorted keys)
	treeAny, _ := info.Get("tree")
	tree, ok := treeAny.(*storage.OrderedMap)
	if !ok {
		t.Fatalf("tree shape: %T", treeAny)
	}
	wantTree := map[string]int{"AI-Domain/Agents": 2, "Tools/Automation": 1}
	if tree.Len() != len(wantTree) {
		t.Fatalf("tree size: %d", tree.Len())
	}
	for _, k := range tree.Keys() {
		v, _ := tree.Get(k)
		if wantTree[k] != v.(int) {
			t.Errorf("tree[%s] = %v, want %d", k, v, wantTree[k])
		}
	}
	if v, _ := info.Get("repo_name"); v != "good-repos" {
		t.Errorf("repo_name: %v", v)
	}
	if v, _ := info.Get("exists"); v != true {
		t.Errorf("exists: %v", v)
	}
}

func TestPublish_EmptyRepoNamePublishSkipsV0830(t *testing.T) {
	// v0.83.0 — THE NO-SILENT-DEFAULT LAW: a GoodRepos built without a
	// repo_name is NOT configured — publish skips instead of falling
	// back to 'good-repos' or ever touching GitHub.
	b := newGRBase(t)
	b.defaultVault()
	pub := NewGoodRepos(b.vault, "", "", false, true, 0, nil, false)
	t.Cleanup(func() { dropStaging(pub) })
	if pub.repoName != "" {
		t.Fatal("repo_name should be empty")
	}
	result := pub.Publish(nil, false)
	if result.SkippedReason == "" {
		t.Fatal("skipped_reason missing")
	}
	if !strings.Contains(result.SkippedReason, "no directory repo configured") {
		t.Errorf("skipped_reason: %s", result.SkippedReason)
	}
	if result.Pushed {
		t.Fatal("pushed")
	}
}

func TestPublish_DryRunBuildsTreeWithoutGit(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	pub := b.publisher(nil)
	result := pub.Publish(nil, true)
	t.Cleanup(func() { _ = os.RemoveAll(result.StagingDir) })
	if !result.OK() {
		t.Fatalf("ok: %s", result.Describe())
	}
	if !result.Published || !result.DryRun {
		t.Fatal("published/dry_run flags")
	}
	if result.Pushed {
		t.Fatal("pushed")
	}
	if result.CommitSha != "" {
		t.Fatal("commit_sha should be nil")
	}
	staging := result.StagingDir
	if isDir(filepath.Join(staging, ".git")) { // no git at all
		t.Fatal(".git exists in a dry-run staging")
	}
	if !isFile(filepath.Join(staging, "README.md")) {
		t.Fatal("README.md missing")
	}
	if !isFile(filepath.Join(staging, "AI-Domain/Agents/aaa_Agents_x.md")) {
		t.Fatal("note missing")
	}
}

// ---------------------------------------------------------------------------
// 3. Anchor correctness (GitHub math) (TestAnchors)
// ---------------------------------------------------------------------------

func TestAnchors_MathSpecExamples(t *testing.T) {
	cases := [][2]string{
		{"🤖 AI Domain", "-ai-domain"},
		{"⚙️ Automation", "-automation"},
		{"🧠 Agents (2)", "-agents-2"},
		{"📦 Uncategorized", "-uncategorized"},
		{"🧰 LLM Tools (1)", "-llm-tools-1"},
		{"🌐 Web Frameworks (3)", "-web-frameworks-3"},
	}
	for _, c := range cases {
		if got := GitHubAnchor(c[0]); got != c[1] {
			t.Errorf("github_anchor(%q) = %q, want %q", c[0], got, c[1])
		}
	}
}

func TestAnchors_ContentsLinksMatchHeadings(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	readme := b.readme(b.publisher(nil).Publish(nil, false))

	headingAnchors := map[string]bool{}
	for _, line := range strings.Split(readme, "\n") {
		if strings.HasPrefix(line, "## ") || strings.HasPrefix(line, "### ") {
			// independent reimplementation of GitHub's anchor algorithm
			text := strings.TrimSpace(strings.TrimLeft(line, "#"))
			text = strings.ToLower(text)
			var kept []rune
			for _, ch := range text {
				if pyIsAlnum(ch) || ch == ' ' || ch == '-' {
					kept = append(kept, ch)
				}
			}
			headingAnchors[strings.ReplaceAll(string(kept), " ", "-")] = true
		}
	}
	targets := map[string]bool{}
	for _, m := range regexp.MustCompile(`\]\(#([^)]+)\)`).FindAllStringSubmatch(readme, -1) {
		targets[m[1]] = true
	}
	if len(targets) == 0 {
		t.Fatal("no link targets")
	}
	for tgt := range targets { // every link target must be a real heading anchor
		if !headingAnchors[tgt] {
			t.Errorf("target %q is not a heading anchor", tgt)
		}
	}
	// … and every section heading is linked (except Contents itself)
	unlinked := map[string]bool{}
	for h := range headingAnchors {
		if !targets[h] {
			unlinked[h] = true
		}
	}
	if len(unlinked) != 1 || !unlinked["-contents"] {
		t.Errorf("unlinked headings: %v", unlinked)
	}
}

func TestAnchors_SpecificAnchorsPresent(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	readme := b.readme(b.publisher(nil).Publish(nil, false))
	for _, want := range []string{"(#-ai-domain)", "(#-agents-2)", "(#-automation-1)", "(#-tools)"} {
		if !strings.Contains(readme, want) {
			t.Errorf("anchor %s missing", want)
		}
	}
}

// ---------------------------------------------------------------------------
// 4. Sorting, ordering, counts (TestOrdering)
// ---------------------------------------------------------------------------

func TestOrdering_SortedByStarsDescAndCategoriesAlpha(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	readme := b.readme(b.publisher(nil).Publish(nil, false))

	// entries sorted by stars desc within their category
	aaaPos := strings.Index(readme, "[aaa/aaa]")
	bbbPos := strings.Index(readme, "[bbb/bbb]")
	if !(bbbPos < aaaPos) { // 900 stars before 100
		t.Error("bbb should precede aaa")
	}
	// category sections alphabetical: AI-Domain before Tools
	if !(strings.Index(readme, "## 🤖 AI Domain") < strings.Index(readme, "## 🛠️ Tools")) {
		t.Error("AI Domain should precede Tools")
	}
	// contents lists AI Domain before Tools too
	contents := strings.SplitN(strings.SplitN(readme, "## 📇 Contents", 2)[1], "\n\n## ", 2)[0]
	if !(strings.Index(contents, "AI Domain") < strings.Index(contents, "Tools")) {
		t.Error("contents order")
	}
}

func TestOrdering_PerCategoryCounts(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	readme := b.readme(b.publisher(nil).Publish(nil, false))
	for _, want := range []string{"### 🧠 Agents (2)", "### ⚙️ Automation (1)",
		"📊 **3 repos** · 🗂️ **2 categories**"} {
		if !strings.Contains(readme, want) {
			t.Errorf("count line %q missing", want)
		}
	}
}

func TestOrdering_SubsAlphabeticalDespiteEmoji(t *testing.T) {
	// emoji code points (✨ U+2728 sorts before 🧠 U+1F9E0) must NOT
	// decide the order — the DISPLAY name does (demo-vault regression)
	b := newGRBase(t)
	for _, sub := range []string{"Skills", "MCP", "Agents", "LLM-Tools"} {
		b.write("AI-Domain/"+sub+"/"+strings.ToLower(sub)+"_note.md",
			simpleNote(strings.ToLower(sub), &simpleNoteOpts{category: sub}))
	}
	readme := b.readme(b.publisher(nil).Publish(nil, false))
	var order []int
	for _, sub := range []string{"Agents", "LLM-Tools", "MCP", "Skills"} {
		order = append(order, strings.Index(readme, "### "+EmojiFor(sub)+" "))
	}
	if !sort.IntsAreSorted(order) {
		t.Errorf("sub order: %v", order)
	}
	contents := strings.SplitN(strings.SplitN(readme, "## 📇 Contents", 2)[1], "\n\n## ", 2)[0]
	var cOrder []int
	for _, n := range []string{"Agents", "LLM Tools", "MCP", "Skills"} {
		cOrder = append(cOrder, strings.Index(contents, n))
	}
	if !sort.IntsAreSorted(cOrder) {
		t.Errorf("contents sub order: %v", cOrder)
	}
}

// ---------------------------------------------------------------------------
// 5-6. TL;DR extraction, display-name resolution (TestNoteParsing)
// ---------------------------------------------------------------------------

func TestNoteParsing_TldrExtractedAndFallback(t *testing.T) {
	b := newGRBase(t)
	b.write("AI-Domain/Agents/aaa_Agents_x.md",
		simpleNote("aaa", &simpleNoteOpts{tldr: strPtr("Primary summary.")}))
	b.write("AI-Domain/Agents/bbb_Agents_y.md",
		simpleNote("bbb", &simpleNoteOpts{tldr: strPtr("")})) // None → What is it? fallback
	b.write("AI-Domain/Agents/ccc_Agents_z.md",
		noteText("https://github.com/ccc/ccc", "ccc", "ccc", "Agents", 5, "Go", "", "solo", ""))
	readme := b.readme(b.publisher(nil).Publish(nil, false))

	if !strings.Contains(readme, "— Primary summary.") {
		t.Error("primary summary missing")
	}
	if !strings.Contains(readme, "— Fallback description sentence.") {
		t.Error("fallback summary missing")
	}
	// no TL;DR anywhere -> line renders without the "—" tail piece
	var cccLine string
	for _, ln := range strings.Split(readme, "\n") {
		if strings.Contains(ln, "[ccc/ccc]") {
			cccLine = ln
			break
		}
	}
	if cccLine == "" {
		t.Fatal("ccc line missing")
	}
	if strings.Contains(cccLine, " — ") {
		t.Errorf("ccc line should have no — tail: %s", cccLine)
	}
	want := regexp.MustCompile(
		"- 📦 \\[ccc/ccc\\]\\(https://github\\.com/ccc/ccc\\) · ⭐ 5 · 🔧 Go · `solo`")
	if !want.MatchString(cccLine) {
		t.Errorf("ccc line: %s", cccLine)
	}
}

func TestNoteParsing_DisplayNameURLAliasesFilename(t *testing.T) {
	b := newGRBase(t)
	// 1) parsed from the source url (with trailing .git + path noise)
	b.write("AI-Domain/Agents/aaa_Agents_x.md",
		strings.ReplaceAll(simpleNote("aaa", nil),
			"https://github.com/aaa/aaa", "https://github.com/aaa/aaa.git/tree/main"))
	// 2) no source -> alias containing "/" wins
	b.write("AI-Domain/Agents/bbb_Agents_y.md",
		strings.ReplaceAll(simpleNote("bbb", nil),
			`source: "https://github.com/bbb/bbb"`, `source: ""`))
	// 3) no source, bare alias only
	b.write("AI-Domain/Agents/ccc_Agents_z.md",
		strings.ReplaceAll(
			noteText("", "ccc", "ccc", "Agents", 7, "Go", "C summary.", "t",
				"Fallback description sentence."),
			"  - ccc\n  - ccc/ccc", "  - ccc"))
	// 4) nothing at all -> filename stem before the first "_"
	b.write("AI-Domain/Agents/ddd_Agents_w.md",
		"# ddd\n\nno frontmatter at all, just a body\n")

	readme := b.readme(b.publisher(nil).Publish(nil, false))
	if !strings.Contains(readme, "[aaa/aaa](https://github.com/aaa/aaa)") {
		t.Error("url display missing")
	}
	if !strings.Contains(readme, "- 📦 bbb/bbb —") { // alias "bbb/bbb" rescued the name
		t.Error("slash-alias display missing")
	}
	if !strings.Contains(readme, "- 📦 ccc —") { // bare alias
		t.Error("bare-alias display missing")
	}
	if !strings.Contains(readme, "- 📦 ddd") { // filename fallback "ddd"
		t.Error("filename fallback missing")
	}
}

func TestNoteParsing_FrontmatterEdgeCases(t *testing.T) {
	b := newGRBase(t)
	weird := `---
source: "https://github.com/xxx/xxx"
aliases: []
tags: not-a-list, malformed]
category: ""
stars: "14k"
org: ""
primary_language: ""
languages: []
credibility_score: 78/100
date_processed: 2026-09-16
---

# xxx

> **TL;DR:** Weird frontmatter.

**` + "`xxx/xxx`" + `** · ⭐ ? · 🔧 ?
`
	b.write("Uncategorized/xxx_Uncategorized_y.md", weird)
	b.write("Uncategorized/yyy_Uncategorized_z.md",
		"# yyy\n\n> **TL;DR:** No frontmatter at all.\n")
	result := b.publisher(nil).Publish(nil, false)
	if result.Entries != 2 || result.Categories != 1 {
		t.Fatalf("entries/categories: %d/%d", result.Entries, result.Categories)
	}
	readme := b.readme(result)
	if !strings.Contains(readme, "## 📦 Uncategorized") {
		t.Error("Uncategorized heading missing")
	}
	// "14k" stars -> unparsable -> ⭐ piece skipped, no crash
	var xxxLine string
	for _, ln := range strings.Split(readme, "\n") {
		if strings.Contains(ln, "[xxx/xxx]") {
			xxxLine = ln
			break
		}
	}
	if xxxLine == "" {
		t.Fatal("xxx line missing")
	}
	if strings.Contains(xxxLine, "⭐") {
		t.Errorf("stars should be skipped: %s", xxxLine)
	}
	if strings.Contains(xxxLine, "🔧") {
		t.Errorf("language should be skipped: %s", xxxLine)
	}
	// malformed tag list degrades into individual tags
	if !strings.Contains(readme, "`not-a-list`") || !strings.Contains(readme, "`malformed`") {
		t.Error("malformed tags not degraded")
	}
	// note without frontmatter still publishes (filename fallback)
	if !strings.Contains(readme, "- 📦 yyy") {
		t.Error("yyy fallback missing")
	}
}

func TestNoteParsing_StarsStringValue(t *testing.T) {
	b := newGRBase(t)
	b.write("AI-Domain/Agents/aaa_Agents_x.md",
		simpleNote("aaa", &simpleNoteOpts{stars: "1234"}))
	readme := b.readme(b.publisher(nil).Publish(nil, false))
	if !strings.Contains(readme, "⭐ 1,234") {
		t.Errorf("comma stars missing:\n%s", readme)
	}
}

// ---------------------------------------------------------------------------
// 7-8. Idempotency + guards (TestLifecycle)
// ---------------------------------------------------------------------------

func TestLifecycle_PublishTwiceSkipsWhenUnchanged(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	pub := b.publisher(nil)
	first := pub.Publish(nil, false)
	if !first.Published {
		t.Fatal("first publish")
	}
	second := pub.Publish(nil, false)
	if second.Published {
		t.Fatal("second publish should skip")
	}
	if second.SkippedReason == "" || !strings.Contains(second.SkippedReason, "unchanged") {
		t.Errorf("skipped_reason: %s", second.SkippedReason)
	}
	if !second.OK() { // a skip is a success
		t.Error("a skip should be ok")
	}
	if second.Pushed {
		t.Error("pushed")
	}
}

func TestLifecycle_RepublishAfterChangeCommitsAgain(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	pub := b.publisher(nil)
	if !pub.Publish(nil, false).Published {
		t.Fatal("first publish")
	}
	b.write("AI-Domain/Agents/ddd_Agents_v.md",
		simpleNote("ddd", &simpleNoteOpts{stars: 1}))
	third := pub.Publish(nil, false)
	if !third.Published {
		t.Fatal("third publish should commit")
	}
	if third.Entries != 4 {
		t.Errorf("entries: %d", third.Entries)
	}
}

func TestLifecycle_DisabledShortCircuits(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	result := b.publisher(&pubOpts{enabled: boolPtr(false)}).Publish(nil, false)
	if result.Published {
		t.Fatal("published")
	}
	if result.SkippedReason == "" || !strings.Contains(result.SkippedReason, "disabled") {
		t.Errorf("skipped_reason: %s", result.SkippedReason)
	}
	if !result.OK() {
		t.Error("disabled should be ok")
	}
}

func TestLifecycle_MissingVaultIsAnError(t *testing.T) {
	b := newGRBase(t)
	pub := NewGoodRepos(filepath.Join(b.vault, "does-not-exist"), "", "",
		false, true, 0, nil, false)
	result := pub.Publish(nil, false)
	if result.OK() {
		t.Fatal("ok")
	}
	if result.Published {
		t.Fatal("published")
	}
	if result.Error == "" || !strings.Contains(result.Error, "vault directory not found") {
		t.Errorf("error: %s", result.Error)
	}
}

func TestLifecycle_EmptyVaultSkips(t *testing.T) {
	b := newGRBase(t)
	result := b.publisher(nil).Publish(nil, false)
	if result.Published {
		t.Fatal("published")
	}
	if result.SkippedReason != "vault contains no curated notes" {
		t.Errorf("skipped_reason: %s", result.SkippedReason)
	}
	if !result.OK() {
		t.Error("empty vault should be ok")
	}
}

// ---------------------------------------------------------------------------
// 9-10. Config bridge (TestConfigBridge)
// ---------------------------------------------------------------------------

func TestConfigBridge_PublishFromConfigBridgesKeys(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	config := map[string]any{
		"vault_path":   b.vault,
		"github_token": "",
		"goodrepos":    map[string]any{"enabled": true, "repo_name": "my-good-dir", "auto_push": true},
	}
	result := PublishFromConfig(config, nil, nil)
	if !result.OK() {
		t.Fatalf("ok: %s", result.Describe())
	}
	if !result.Published {
		t.Fatal("published")
	}
	if result.Entries != 3 {
		t.Errorf("entries: %d", result.Entries)
	}
	if result.RepoName != "my-good-dir" {
		t.Errorf("repo_name: %s", result.RepoName)
	}
	if result.Pushed { // no token -> auto_push degrades
		t.Fatal("pushed")
	}
}

func TestConfigBridge_DisabledShortCircuits(t *testing.T) {
	// enabled=False returns BEFORE any disk access: even a bogus
	// vault_path never turns into an error.
	config := map[string]any{
		"vault_path":   "/no/such/vault-anywhere",
		"github_token": "irrelevant-but-empty",
		"goodrepos":    map[string]any{"enabled": false},
	}
	result := PublishFromConfig(config, nil, nil)
	if result.Published {
		t.Fatal("published")
	}
	if result.SkippedReason == "" || !strings.Contains(result.SkippedReason, "disabled") {
		t.Errorf("skipped_reason: %s", result.SkippedReason)
	}
	if !result.OK() {
		t.Error("disabled should be ok")
	}
}

// ---------------------------------------------------------------------------
// 11. Root extras copied, machine state excluded (TestVaultHygiene)
// ---------------------------------------------------------------------------

func TestVaultHygiene_ExtrasCopiedAndNoiseExcluded(t *testing.T) {
	b := newGRBase(t)
	b.write("AI-Domain/Agents/aaa_Agents_x.md", simpleNote("aaa", nil))
	b.write("_index.md", "# Index\n")
	b.write("links_manifest.json", `{"generated": "2026-09-16", "links": []}`)
	b.write("README.md", "# the vault's own README (must NOT win)\n")
	b.write(".obsidian/app.json", "{}")
	b.write(".obsidian/workspace.json", "{}")
	b.write("_inbox/stray.md", simpleNote("stray", nil))
	b.write("_review/pending.md", simpleNote("pending", nil))
	b.write("_moc/map.md", simpleNote("map", nil))
	b.write(".trash/deleted.md", simpleNote("deleted", nil))
	b.write("__pycache__/junk.md", simpleNote("junk", nil))
	b.write("AI-Domain/Agents/notes.txt", "not a note")

	result := b.publisher(nil).Publish(nil, false)
	if result.Entries != 1 {
		t.Fatalf("entries: %d", result.Entries)
	}
	staging := result.StagingDir
	if !isFile(filepath.Join(staging, "_index.md")) {
		t.Error("_index.md missing")
	}
	if !isFile(filepath.Join(staging, "links_manifest.json")) {
		t.Error("links_manifest.json missing")
	}
	for _, excluded := range []string{
		".obsidian", "_inbox", "_review", "_moc", ".trash", "__pycache__",
		"AI-Domain/Agents/notes.txt",
	} {
		if _, err := os.Stat(filepath.Join(staging, excluded)); err == nil {
			t.Errorf("%s leaked into the directory", excluded)
		}
	}
	readme := b.readme(result)
	if strings.Contains(readme, "stray") {
		t.Error("stray leaked")
	}
	if strings.Contains(readme, "vault's own README") {
		t.Error("the vault's own README won")
	}
	// the generated README is the directory's README
	if !strings.Contains(readme, "# 🗂️ Good Repos") {
		t.Error("generated README missing")
	}
}

func TestVaultHygiene_RootNotesAreNotCategories(t *testing.T) {
	// .md files directly in the vault root are not curated entries
	b := newGRBase(t)
	b.write("loose.md", simpleNote("loose", nil))
	b.write("AI-Domain/Agents/aaa_Agents_x.md", simpleNote("aaa", nil))
	result := b.publisher(nil).Publish(nil, false)
	if result.Entries != 1 {
		t.Fatalf("entries: %d", result.Entries)
	}
	readme := b.readme(result)
	if strings.Contains(readme, "loose") {
		t.Error("loose root note leaked")
	}
}

// ---------------------------------------------------------------------------
// 12. Star formatting + emoji map (TestHelpers)
// ---------------------------------------------------------------------------

func TestHelpers_FormatStars(t *testing.T) {
	cases := []struct {
		value any
		want  string
	}{
		{0, "0"}, {999, "999"}, {1234, "1,234"}, {9999, "9,999"},
		{10000, "10k"}, {14023, "14k"}, {42987, "43k"}, {86750, "86.8k"},
		{140232, "140.2k"},
	}
	for _, c := range cases {
		if got := FormatStarsAny(c.value); got != c.want {
			t.Errorf("format_stars(%v) = %q, want %q", c.value, got, c.want)
		}
	}
	if got := FormatStarsAny(nil); got != "" {
		t.Errorf("format_stars(None) = %q", got)
	}
	if got := FormatStarsAny("14k"); got != "" {
		t.Errorf(`format_stars("14k") = %q`, got)
	}
}

func TestHelpers_EmojiMapAndFallback(t *testing.T) {
	if got := EmojiFor("AI-Domain"); got != "🤖" {
		t.Errorf("AI-Domain: %s", got)
	}
	if got := EmojiFor("LLM-Tools"); got != "🧰" {
		t.Errorf("LLM-Tools: %s", got)
	}
	if got := EmojiFor("Uncategorized"); got != "📦" {
		t.Errorf("Uncategorized: %s", got)
	}
	if got := EmojiFor("No-Such-Category"); got != "📁" {
		t.Errorf("fallback: %s", got)
	}
	// deepest component wins for nested category paths
	if got := EmojiFor("AI-Domain", "Agents", "Frameworks"); got != "🏗️" {
		t.Errorf("nested: %s", got)
	}
	if _, ok := CategoryEmoji["AI-Domain"]; !ok {
		t.Error("AI-Domain not in CATEGORY_EMOJI")
	}
}

func TestHelpers_DisplayName(t *testing.T) {
	if got := DisplayName("AI-Domain"); got != "AI Domain" {
		t.Errorf("AI-Domain: %s", got)
	}
	if got := DisplayName("Web-Frameworks"); got != "Web Frameworks" {
		t.Errorf("Web-Frameworks: %s", got)
	}
}

func TestHelpers_ResultDictContract(t *testing.T) {
	result := &PublishResult{Published: true, Entries: 3, Categories: 2}
	if !result.OK() {
		t.Fatal("ok")
	}
	d := result.ToMap()
	for _, key := range []string{
		"published", "skipped_reason", "error", "commit_sha",
		"files_changed", "pushed", "repo_name", "repo_url",
		"entries", "categories", "duration_ms", "ok",
	} {
		if _, ok := d.Get(key); !ok {
			t.Errorf("key %s missing", key)
		}
	}
	if !strings.Contains(result.Describe(), "3 repos across 2 categories") {
		t.Errorf("describe: %s", result.Describe())
	}
	skipped := &PublishResult{SkippedReason: "nothing to do"}
	if !skipped.OK() {
		t.Error("skipped should be ok")
	}
	if skipped.Describe() != "skipped — nothing to do" {
		t.Errorf("describe: %s", skipped.Describe())
	}
	failed := &PublishResult{Error: "boom"}
	if failed.OK() {
		t.Error("failed should not be ok")
	}
}

// ---------------------------------------------------------------------------
// v0.83.0 — THE CREATION GUARD (TestCreationGuardV0830)
// ---------------------------------------------------------------------------

// fakeAPI mirrors _stub_api: a table-driven _api replacement that
// records every call.
func fakeAPI(pub *GoodRepos, responses map[[2]string][2]any) *[][3]any {
	calls := &[][3]any{}
	pub.apiFn = func(method, apiPath string, body map[string]any) (int, map[string]any) {
		*calls = append(*calls, [3]any{method, apiPath, body})
		key := [2]string{method, apiPath}
		if r, ok := responses[key]; ok {
			payload, _ := r[1].(map[string]any)
			if r[1] == nil {
				payload = map[string]any{}
			}
			status, _ := r[0].(int)
			return status, payload
		}
		return 404, map[string]any{"message": "Not Found"}
	}
	return calls
}

func TestCreationGuardV0830_404NeverCreatesWithoutOptIn(t *testing.T) {
	b := newGRBase(t)
	b.write("AI-Domain/Agents/aaa_Agents_python.md", simpleNote("aaa", nil))
	pub := b.publisher(&pubOpts{repoName: "curated-dir", token: "tok"})
	calls := fakeAPI(pub, map[[2]string][2]any{
		{"GET", "/user"}: {200, map[string]any{"login": "testuser"}},
	})
	ok, err := pub.ensureRepo()
	if ok {
		t.Fatal("ok should be false")
	}
	if err == "" || !strings.Contains(err, "does NOT create repositories automatically") {
		t.Errorf("err: %s", err)
	}
	for _, c := range *calls {
		if c[0] == "POST" {
			t.Fatalf("unexpected POST: %v", c)
		}
	}
}

func TestCreationGuardV0830_AllowCreatePostsPrivateOnce(t *testing.T) {
	b := newGRBase(t)
	b.write("AI-Domain/Agents/aaa_Agents_python.md", simpleNote("aaa", nil))
	pub := b.publisher(&pubOpts{repoName: "curated-dir", token: "tok", allowCreate: true})
	calls := fakeAPI(pub, map[[2]string][2]any{
		{"GET", "/user"}:        {200, map[string]any{"login": "testuser"}},
		{"POST", "/user/repos"}: {201, map[string]any{"full_name": "testuser/curated-dir"}},
	})
	ok, err := pub.ensureRepo()
	if !ok {
		t.Fatalf("ok: %s", err)
	}
	var posts int
	for _, c := range *calls {
		if c[0] == "POST" {
			posts++
			body, _ := c[2].(map[string]any)
			if body["private"] != true { // v0.83.0 — the owner's word: no public repos yet
				t.Errorf("POST body not private: %v", body)
			}
		}
	}
	if posts != 1 {
		t.Fatalf("posts: %d", posts)
	}
}

func TestCreationGuardV0830_EmptyNameIsNotConfigured(t *testing.T) {
	b := newGRBase(t)
	b.write("AI-Domain/Agents/aaa_Agents_python.md", simpleNote("aaa", nil))
	pub := b.publisher(&pubOpts{token: "tok", repoName: "zzz"}) // named, then blanked
	pub.repoName = ""
	if pub.repoName != "" {
		t.Fatal("repo_name should be empty")
	}
	calls := fakeAPI(pub, map[[2]string][2]any{})
	ok, err := pub.ensureRepo()
	if ok {
		t.Fatal("ok should be false")
	}
	if err == "" || !strings.Contains(err, "no directory repo configured") {
		t.Errorf("err: %s", err)
	}
	if len(*calls) != 0 {
		t.Fatalf("calls: %v", *calls)
	}
}

func TestCreationGuardV0830_ConfigBridgeNoDefaultAnymore(t *testing.T) {
	// publish_from_config with an UNNAMED goodrepos section skips —
	// the old `or DEFAULT_REPO_NAME` silent fallback is gone.
	b := newGRBase(t)
	cfg := map[string]any{
		"vault_path":   b.vault,
		"github_token": "tok",
		"goodrepos":    map[string]any{"enabled": true, "auto_push": false},
	}
	result := PublishFromConfig(cfg, nil, nil)
	if result.SkippedReason == "" {
		t.Fatal("skipped_reason missing")
	}
	if !strings.Contains(result.SkippedReason, "no directory repo configured") {
		t.Errorf("skipped_reason: %s", result.SkippedReason)
	}
	if result.Pushed {
		t.Fatal("pushed")
	}
}

// ---------------------------------------------------------------------------
// The acceptance's fake-GitHub publish flow — LOCAL BARE REPOS as the
// remote (through the tokenURLFn instance seam; the reference's own
// test_sealfix technique)
// ---------------------------------------------------------------------------

func runGitT(t *testing.T, cwd string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", cwd}, args...)...).CombinedOutput()
	if err != nil {
		t.Fatalf("git -C %s %v: %v\n%s", cwd, args, err, out)
	}
	return string(out)
}

func bareRepoT(t *testing.T, path string) string {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
	runGitT(t, path, "init", "--bare", "-b", "main")
	return path
}

// publishFlowFor builds the auto-push publisher with the fake API and
// the local-bare tokenURLFn seam.
func publishFlowFor(t *testing.T, vault, bare, token string) *GoodRepos {
	t.Helper()
	pub := NewGoodRepos(vault, token, "mirror", true, true, 0, nil, false)
	fakeAPI(pub, map[[2]string][2]any{
		{"GET", "/user"}:                  {200, map[string]any{"login": "testuser"}},
		{"GET", "/repos/testuser/mirror"}: {200, map[string]any{"private": true}},
	})
	pub.tokenURLFn = func() string { return bare }
	t.Cleanup(func() { dropStaging(pub) })
	return pub
}

func TestPublishFlow_WithLocalBareRemote(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	bare := bareRepoT(t, filepath.Join(t.TempDir(), "remote.git"))
	pub := publishFlowFor(t, b.vault, bare, "SUPERSECRETTOKEN")

	first := pub.Publish(nil, false)
	if !first.OK() || !first.Published {
		t.Fatalf("first publish: %s", first.Describe())
	}
	if !first.Pushed {
		t.Fatalf("first publish should push: %s", first.Describe())
	}
	// the directory landed on the bare remote's main
	tree := strings.Fields(runGitT(t, bare, "ls-tree", "-r", "--name-only", "main"))
	for _, want := range []string{"README.md", "_index.md",
		"AI-Domain/Agents/aaa_Agents_x.md", "Tools/Automation/ccc_Automation_z.md"} {
		found := false
		for _, p := range tree {
			if p == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("remote tree missing %s", want)
		}
	}
	// the first publish's base repo_name is the RAW name (the frozen-
	// before-ensure quirk); the SECOND publish sees owner/repo
	if first.RepoName != "mirror" {
		t.Errorf("first repo_name: %s", first.RepoName)
	}
	// THE TOKEN LAW: the token never reached .git/config
	cfgData, err := os.ReadFile(filepath.Join(pub.staging, ".git", "config"))
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(cfgData), "SUPERSECRETTOKEN") {
		t.Fatalf("THE TOKEN LAW BROKEN — the token reached .git/config:\n%s", cfgData)
	}

	// the full cycle: publish again → the fetch aligns with the remote
	// → skip-when-unchanged works WITHOUT a real remote round trip
	second := pub.Publish(nil, false)
	if second.Published {
		t.Fatalf("second publish should skip: %s", second.Describe())
	}
	if !strings.Contains(second.SkippedReason, "unchanged") {
		t.Errorf("skipped_reason: %s", second.SkippedReason)
	}
	if second.RepoName != "testuser/mirror" {
		t.Errorf("second repo_name (owner-qualified after ensure): %s", second.RepoName)
	}

	// the fetch succeeded through the token-URL door: FETCH_HEAD exists
	// and carries NO token (the sanitize door ran)
	fhData, err := os.ReadFile(filepath.Join(pub.staging, ".git", "FETCH_HEAD"))
	if err != nil {
		t.Fatal("FETCH_HEAD should exist after the aligned fetch")
	}
	if strings.Contains(string(fhData), "SUPERSECRETTOKEN") {
		t.Fatalf("FETCH_HEAD carries the token:\n%s", fhData)
	}
}

func TestPublishFlow_PushFailureKeepsCommitWithRedaction(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	// a NONEXISTENT local path with the token embedded: git echoes the
	// path in its failure output — the redaction door must mask the
	// token before it reaches the result (deterministic, offline).
	badRemote := filepath.Join(t.TempDir(), "xSUPERSECRETTOKENx", "repo.git")
	pub := publishFlowFor(t, b.vault, badRemote, "SUPERSECRETTOKEN")

	result := pub.Publish(nil, false)
	if !result.Published {
		t.Fatal("the local commit should exist (commit kept locally)")
	}
	if result.Pushed {
		t.Fatal("pushed")
	}
	if result.Error == "" || !strings.Contains(result.Error, "push failed:") {
		t.Fatalf("error: %s", result.Error)
	}
	if strings.Contains(result.Error, "SUPERSECRETTOKEN") {
		t.Fatalf("THE REDACTION LAW BROKEN — the token survived the push error: %s", result.Error)
	}
	if !strings.Contains(result.Error, "***") {
		t.Fatalf("the mask is missing: %s", result.Error)
	}
	// the reference's law: error set ⇒ ok is FALSE even with the kept
	// commit (the exit code is 1; describe speaks the kept commit)
	if result.OK() {
		t.Fatal("a push failure should not be ok")
	}
	if !strings.Contains(result.Describe(), "(commit kept locally)") {
		t.Errorf("describe: %s", result.Describe())
	}
}

func TestPublishFlow_FetchHeadScrubThroughRealFetch(t *testing.T) {
	b := newGRBase(t)
	b.defaultVault()
	tmp := t.TempDir()
	// the bare repo lives INSIDE a directory whose name carries the
	// token — the fetch URL (a local path) deterministically embeds
	// it, FETCH_HEAD records it, and the sanitize door must rewrite.
	tokenDir := filepath.Join(tmp, "tok-SUPERSECRETTOKEN")
	bare := bareRepoT(t, filepath.Join(tokenDir, "remote.git"))
	pub := publishFlowFor(t, b.vault, bare, "SUPERSECRETTOKEN")

	first := pub.Publish(nil, false)
	if !first.Pushed {
		t.Fatalf("first publish: %s", first.Describe())
	}
	fh := filepath.Join(pub.staging, ".git", "FETCH_HEAD")
	// the first fetch failed (empty remote) — force a SUCCEEDING fetch
	// through the token-embedding path, then run the sanitize door.
	if rc, _ := pub.gitPush("fetch", pub.tokenURLFn(), "main"); rc != 0 {
		t.Fatalf("fetch through the token path failed")
	}
	pub.sanitizeFetchHead()
	data, err := os.ReadFile(fh)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "SUPERSECRETTOKEN") {
		t.Fatalf("FETCH_HEAD carries the token:\n%s", data)
	}
	if !strings.Contains(string(data), "***") {
		t.Fatalf("the mask is missing:\n%s", data)
	}
}
