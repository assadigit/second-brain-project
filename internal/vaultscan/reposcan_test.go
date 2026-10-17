// reposcan_test.go — T-45: the ported test_reposcan selections
// (v0.83.0 THE VAULT-KIND DOOR — the inventory's kind door, the plan's
// kind door, ONE NOTE PER REPO, the repos digest, the apply's kind
// door, and the repos regex brain + weighed evidence). The
// two-vault-worker and GUI-model legs stay with the future GUI task;
// the release-books legs are the Python repo's own CI facts.
package vaultscan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/repobanish"
	"github.com/assadigit/second-brain-project/internal/store"
)

// repoCase is the repos plumbing: temp repos vault + cache + note
// state.
type repoCase struct {
	t     *testing.T
	tmp   string
	vault string
	cache *store.CacheDB
	state *notestate.DB
	logs  *scanLog
}

func newRepoCase(t *testing.T) *repoCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "github-repos")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	cache, err := store.NewCacheDB(filepath.Join(tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	state, err := notestate.New(filepath.Join(tmp, "note_state.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		dryrun.Disable()
		dryrun.Clear()
		cache.Close()
		state.Close()
	})
	return &repoCase{t: t, tmp: tmp, vault: vault, cache: cache,
		state: state, logs: &scanLog{}}
}

// repoNote mirrors the reference suite's _repo_note (the frontmatter
// the scan grammar reads: source + managed_by + tags + the repos
// metadata).
func (rc *repoCase) repoNote(rel, url string, opts ...noteOpt) string {
	rc.t.Helper()
	n := noteSpec{managed: "gitcurator", fetchStatus: "full",
		tags: []string{"cli"}, category: "Agents", stars: "120",
		language: "Python", org: "some-org"}
	for _, o := range opts {
		o(&n)
	}
	path := filepath.Join(rc.vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		rc.t.Fatal(err)
	}
	body := n.body
	if body == "" {
		body = "A curated repo note. " + strings.Repeat("x", 200)
	}
	text := fmt.Sprintf(`---
source: %s
aliases: []
tags: [%s]
category: "%s"
stars: %s
org: %s
primary_language: %s
date_processed: 2026-10-09
managed_by: "%s"
%s
# %s

%s

---
*Source: [GitHub](%s)*
`, url, strings.Join(n.tags, ", "), n.category, n.stars, n.org,
		n.language, n.managed, n.extra,
		strings.TrimSuffix(filepath.Base(rel), ".md"), body, url)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		rc.t.Fatal(err)
	}
	return path
}

// missingNote mirrors the reference suite's _missing_note (a 404
// machinery placeholder, app-owned, bodyless).
func (rc *repoCase) missingNote(owner, repo string, strikes int) string {
	rc.t.Helper()
	path := filepath.Join(rc.vault, "_missing",
		owner+"_"+repo+"_missing.md")
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		rc.t.Fatal(err)
	}
	text := fmt.Sprintf(`---
source: "https://github.com/%s/%s"
aliases: []
tags: []
category: ""
date_processed: 2026-10-09
managed_by: "gitcurator"
schema_version: 2
---
# %s/%s — missing

Missing repo placeholder (404). Strikes: %d.
`, owner, repo, owner, repo, strikes)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		rc.t.Fatal(err)
	}
	return path
}

func (rc *repoCase) buildPlan(llm LLMCallFn, noteState *notestate.DB) Plan {
	return BuildScanPlan(rc.vault, llm, rc.logs.log, nil, nil, nil,
		VaultKindRepos, noteState)
}

func markRepoNote(rc *repoCase, path string) {
	text := readFileT(rc.t, path)
	text = strings.Replace(text, "tags: [cli]", "tags: [🗑️, cli]", 1)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		rc.t.Fatal(err)
	}
}

// --- TestTheInventoryKindDoor -----------------------------------------

func TestTheReposMetadataRidesTheRows(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("AI-Domain/Agents/Note A.md",
		"https://github.com/some-org/repo-a",
		withStars("4200"), withLanguage("Rust"), withOrg("acme"))
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	var row NoteRow
	for _, n := range inv.Notes {
		if n.File == "Note A.md" {
			row = n
		}
	}
	if row.Stars != 4200 || row.PrimaryLang != "Rust" ||
		row.Org != "acme" || !row.AppOwned {
		t.Fatalf("row %+v", row)
	}
}

func TestMissingIsProtectedForReposOnly(t *testing.T) {
	rc := newRepoCase(t)
	rc.missingNote("ghost-owner", "ghost-repo", 1)
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	var row NoteRow
	for _, n := range inv.Notes {
		if strings.Contains(n.File, "ghost-repo") {
			row = n
		}
	}
	if !row.Protected {
		t.Fatal("not protected for repos")
	}
	for _, c := range MoveCandidates(&inv, 40) {
		if strings.Contains(c.File, "ghost-repo") {
			t.Fatal("the placeholder rode the candidates")
		}
	}
	// the websites kind keeps its own protected set (no _missing)
	invW := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindWebsites)
	var rowW NoteRow
	for _, n := range invW.Notes {
		if strings.Contains(n.File, "ghost-repo") {
			rowW = n
		}
	}
	if rowW.Protected {
		t.Fatal("protected for websites too")
	}
}

func TestRootAndUncategorizedAreCandidatesForRepos(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("Orphan.md", "https://github.com/o/orphan")
	rc.repoNote("Uncategorized/Parked.md", "https://github.com/o/parked")
	rc.repoNote("AI-Domain/Agents/Well.md", "https://github.com/o/well")
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	files := map[string]bool{}
	for _, c := range MoveCandidates(&inv, 40) {
		files[c.File] = true
	}
	if !files["Orphan.md"] || !files["Parked.md"] || files["Well.md"] {
		t.Fatalf("files %v", files)
	}
}

func TestTheKindHelpersSpeakTheTruth(t *testing.T) {
	if !isReposKind("repos") || !isReposKind("REPOS") ||
		isReposKind("websites") || isReposKind("") {
		t.Fatal("is_repos_kind")
	}
	if kindLabel("repos") != "GitHub Repos vault" ||
		kindLabel("websites") != "Websites vault" {
		t.Fatal("kind_label")
	}
	have := false
	for _, f := range protectedFoldersFor("repos") {
		if f == "_missing" {
			have = true
		}
	}
	if !have {
		t.Fatal("_missing missing from the repos set")
	}
	for _, f := range protectedFoldersFor("websites") {
		if f == "_missing" {
			t.Fatal("_missing in the websites set")
		}
	}
}

// --- TestThePlanKindDoor ----------------------------------------------

func TestThePlanCarriesTheKind(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("Orphan.md", "https://github.com/o/orphan")
	plan := rc.buildPlan(nil, nil)
	if plan.VaultKind != "repos" {
		t.Fatalf("kind %q", plan.VaultKind)
	}
	if plan.VaultLabel != "GitHub Repos vault" {
		t.Fatalf("label %q", plan.VaultLabel)
	}
	if plan.LLMCandidates != 1 {
		t.Fatalf("candidates %d", plan.LLMCandidates)
	}
}

func TestTheWebsitesDefaultIsUnchanged(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("Orphan.md", "https://github.com/o/orphan")
	plan := BuildScanPlan(rc.vault, nil, rc.logs.log, nil, nil, nil,
		"", nil)
	if plan.VaultKind != "websites" || plan.VaultLabel != "Websites vault" {
		t.Fatalf("plan %+v", plan)
	}
}

func TestTheDeletionsRideTheReposGate(t *testing.T) {
	rc := newRepoCase(t)
	p := rc.repoNote("AI-Domain/Agents/Marked.md",
		"https://github.com/o/marked")
	markRepoNote(rc, p)
	plan := rc.buildPlan(nil, rc.state)
	if len(plan.DeletionsRepos) != 1 {
		t.Fatalf("deletions %d", len(plan.DeletionsRepos))
	}
	item := plan.DeletionsRepos[0]
	canon := links.NormalizeURL("https://github.com/o/marked")
	if item.Canonical != canon {
		t.Fatalf("canonical %q", item.Canonical)
	}
}

func TestALockedRepoNoteNeverRidesTheDeletions(t *testing.T) {
	rc := newRepoCase(t)
	p := rc.repoNote("AI-Domain/Agents/Locked.md",
		"https://github.com/o/locked", withExtra("locked: true\n"))
	markRepoNote(rc, p)
	plan := rc.buildPlan(nil, rc.state)
	if len(plan.DeletionsRepos) != 0 {
		t.Fatalf("deletions %d", len(plan.DeletionsRepos))
	}
	if !rc.logs.has("locked") {
		t.Fatal("missing lock log")
	}
}

func TestThePoisonedWellIsCapped(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("AI-Domain/Agents/Sib.md", "https://github.com/o/one")
	// a stale category key (its mapped folder does not exist on disk)
	rc.repoNote("Orphan.md", "https://github.com/o/orphan",
		withCategory("MCP"))
	plan := rc.buildPlan(nil, nil)
	if plan.LLMCandidates != 1 || plan.RegexFiled != 0 {
		t.Fatalf("plan %+v", plan)
	}
}

// --- TestOneNotePerRepo ------------------------------------------------

func TestTheTrailingSlashTwinIsOneProposal(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("AI-Domain/Agents/Twin.md", "https://github.com/o/twin")
	rc.repoNote("AI-Domain/Agents/Twin 2.md", "https://github.com/o/twin/")
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	props := ScanSiteDuplicates(&inv, rc.logs.log, VaultKindRepos)
	if len(props) != 1 {
		t.Fatalf("props %d", len(props))
	}
	// the keeper is the deterministic winner: equal URL paths, so the
	// walk's own order speaks — 'Twin 2.md' sorts before 'Twin.md'
	if props[0].Keep.File != "Twin 2.md" {
		t.Fatalf("keep %q", props[0].Keep.File)
	}
	if len(props[0].Merge) != 1 || props[0].Merge[0].File != "Twin.md" {
		t.Fatalf("merge %+v", props[0].Merge)
	}
	if !rc.logs.has("repo(s) hold more than one note") {
		t.Fatal("missing magnet log")
	}
}

func TestAMissingPlaceholderMergesIntoTheRealNote(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("Tools/Scraping/Real.md", "https://github.com/o/real")
	rc.missingNote("o", "real", 1)
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	props := ScanSiteDuplicates(&inv, rc.logs.log, VaultKindRepos)
	if len(props) != 1 {
		t.Fatalf("props %d", len(props))
	}
	if props[0].Keep.File != "Real.md" {
		t.Fatalf("keep %q", props[0].Keep.File)
	}
	if !strings.Contains(props[0].Merge[0].File, "real_missing") {
		t.Fatalf("merge %+v", props[0].Merge)
	}
}

func TestReposPlaceholdersAloneProduceNothing(t *testing.T) {
	rc := newRepoCase(t)
	rc.missingNote("ghost", "one", 1)
	rc.missingNote("ghost", "two", 1)
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	if props := ScanSiteDuplicates(&inv, rc.logs.log, VaultKindRepos); len(props) != 0 {
		t.Fatalf("props %+v", props)
	}
}

func TestWebsitesGroupingStillKeysOnTheDomain(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("AI-Domain/Agents/A.md", "https://github.com/o/a")
	rc.repoNote("AI-Domain/Agents/B.md", "https://gitlab.com/o/b")
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, "")
	if props := ScanSiteDuplicates(&inv, rc.logs.log, ""); len(props) != 0 {
		t.Fatalf("props %+v", props)
	}
}

// --- TestTheReposDigest ------------------------------------------------

func TestTheDigestSpeaksTheReposOwnEvidence(t *testing.T) {
	lines := DigestLines([]NoteRow{{
		Title: "whizzy", File: "Orphan.md", Org: "acme",
		Stars: 4200, PrimaryLang: "Rust",
		Tags: []string{"rust", "cli"}, Folder: "", InRoot: true,
	}}, VaultKindRepos)
	line := lines[0]
	for _, needle := range []string{"by acme", "⭐ 4,200", "🔧 Rust",
		"tags: rust,cli"} {
		if !strings.Contains(line, needle) {
			t.Fatalf("line %q missing %q", line, needle)
		}
	}
	if strings.Contains(line, "github.com") {
		t.Fatal("the domain stayed home")
	}
}

func TestTheWebsitesDigestKeepsItsDomain(t *testing.T) {
	lines := DigestLines([]NoteRow{{
		Title: "A Site", File: "A.md", Domain: "a.example",
		Tags: []string{}, Folder: "", InRoot: true,
	}}, VaultKindWebsites)
	if !strings.Contains(lines[0], "a.example") {
		t.Fatalf("line %q", lines[0])
	}
}

func TestTheReposPromptExistsAndSpeaksRepos(t *testing.T) {
	prompt := ReadPrompt(VaultKindRepos)
	if !strings.Contains(prompt, "GitHub repository notes") ||
		!strings.Contains(prompt, "star count") {
		t.Fatalf("repos prompt %q", prompt[:80])
	}
	websitesPrompt := ReadPrompt(VaultKindWebsites)
	if !strings.Contains(websitesPrompt, "website notes") {
		t.Fatal("websites prompt")
	}
}

// --- TestTheApplyKindDoor ---------------------------------------------

func (rc *repoCase) reposPlanWithMove() (*Plan, string, string) {
	src := rc.repoNote("Orphan.md", "https://github.com/o/mover")
	if err := os.MkdirAll(filepath.Join(rc.vault, "Tools", "Scraping"),
		0o755); err != nil {
		rc.t.Fatal(err)
	}
	rc.repoNote("Tools/Scraping/Resident.md",
		"https://github.com/o/resident")
	canon := links.NormalizeURL("https://github.com/o/mover")
	rc.cache.AddProcessed(101, "https://github.com/o/mover", "o", "mover",
		src, "Uncategorized")
	plan := &Plan{
		VaultKind: "repos",
		Moves: []MoveRow{{
			Note: "Orphan.md", Path: src, Title: "Orphan",
			From: "(vault root)", To: "Tools/Scraping", As: "",
			Reason: "test",
		}},
	}
	return plan, src, canon
}

func TestTheMoveRepointsTheLedgerAndStampsCurated(t *testing.T) {
	rc := newRepoCase(t)
	plan, src, canon := rc.reposPlanWithMove()
	rep := ApplyScanPlan(plan, rc.vault, nil, rc.logs.log,
		Selection{}, VaultKindRepos, rc.cache, nil)
	if rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeGone(t, src)
	dst := filepath.Join(rc.vault, "Tools", "Scraping", "Orphan.md")
	mustBeFile(t, dst)
	if rep.Stamped != 1 {
		t.Fatalf("stamped %d", rep.Stamped)
	}
	if !strings.Contains(readFileT(t, dst), "curated: true") {
		t.Fatal("no stamp")
	}
	row := rc.cache.ProcessedRowFor(canon)
	if row == nil || row.NotePath == nil ||
		!strings.HasSuffix(*row.NotePath,
			filepath.Join("Tools", "Scraping", "Orphan.md")) {
		t.Fatalf("row %+v", row)
	}
	if row.Category == nil || *row.Category != "Scraping" {
		t.Fatalf("category %v (reverse-mapped)", row.Category)
	}
}

func TestTheConsolidationRepointsAndResolves(t *testing.T) {
	rc := newRepoCase(t)
	keep := rc.repoNote("Tools/Scraping/Keep.md",
		"https://github.com/o/dupe")
	twin := rc.repoNote("Tools/Scraping/Keep 2.md",
		"https://github.com/o/dupe/")
	canon := links.NormalizeURL("https://github.com/o/dupe/")
	rc.cache.AddProcessed(202, canon, "o", "dupe", twin, "Scraping")
	rc.cache.AddFailed(canon, "a 404 once")
	plan := &Plan{
		VaultKind: "repos",
		Consolidations: []Consolidation{{
			Site: canon,
			Keep: KeepRow{File: "Keep.md", Path: keep,
				Source: "https://github.com/o/dupe", Title: "Keep",
				Category: "Scraping"},
			Merge: []MergeRow{{File: "Keep 2.md", Path: twin,
				Source: "https://github.com/o/dupe/",
				Title:  "Keep 2"}},
			URLs: []string{"https://github.com/o/dupe/"},
		}},
	}
	rep := ApplyScanPlan(plan, rc.vault, nil, rc.logs.log,
		Selection{}, VaultKindRepos, rc.cache, nil)
	if rep.Consolidated != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeGone(t, twin)
	mustBeFile(t, keep)
	// the twin's canonical IS the keeper's own source — NO link is
	// appended; the keeper stays byte-identical
	if strings.Contains(readFileT(t, keep), "site_links") {
		t.Fatal("a redundant link was appended")
	}
	row := rc.cache.ProcessedRowFor(canon)
	if row == nil || row.NotePath == nil || *row.NotePath != keep {
		t.Fatalf("row %+v", row)
	}
}

func TestTheDeletionsEnforceTheReposBurial(t *testing.T) {
	rc := newRepoCase(t)
	src := rc.repoNote("AI-Domain/Agents/Doomed.md",
		"https://github.com/o/doomed")
	markRepoNote(rc, src)
	canon := links.NormalizeURL("https://github.com/o/doomed")
	plan := &Plan{
		VaultKind: "repos",
		DeletionsRepos: []repobanish.PendingRepoBanishment{{
			URL: "https://github.com/o/doomed", Canonical: canon,
			Title: "Doomed", Marker: "🗑️", Door: "note tag", Path: src,
		}},
	}
	rep := ApplyScanPlan(plan, rc.vault, nil, rc.logs.log,
		Selection{}, VaultKindRepos, rc.cache, rc.state)
	if rep.Banished != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeGone(t, src)
	mustBeDir(t, filepath.Join(rc.vault, ".trash", "banished"))
	// the never-fetch promise: dismissed + settled
	dismissed, err := rc.state.IsDismissed(notestate.VaultGithub, canon)
	if err != nil || !dismissed {
		t.Fatalf("dismissed %v %v", dismissed, err)
	}
	if !rc.cache.GetSettledRepoSet()[canon] {
		t.Fatal("not settled")
	}
	// the record row landed (♻️ revivable)
	table := filepath.Join(rc.vault, "_review", "DECOMMISSIONED.md")
	mustBeFile(t, table)
	if !strings.Contains(readFileT(t, table), "github.com/o/doomed") {
		t.Fatal("no record row")
	}
}

func TestAnUncheckedDeletionRowStays(t *testing.T) {
	rc := newRepoCase(t)
	src := rc.repoNote("AI-Domain/Agents/Kept.md",
		"https://github.com/o/kept")
	markRepoNote(rc, src)
	canon := links.NormalizeURL("https://github.com/o/kept")
	plan := &Plan{
		VaultKind: "repos",
		DeletionsRepos: []repobanish.PendingRepoBanishment{{
			URL: "https://github.com/o/kept", Canonical: canon,
			Title: "Kept", Marker: "🗑️", Door: "note tag", Path: src,
		}},
	}
	sel := Selection{Deletions: map[string]bool{}} // ALL unchecked
	rep := ApplyScanPlan(plan, rc.vault, nil, rc.logs.log, sel,
		VaultKindRepos, rc.cache, rc.state)
	if rep.Banished != 0 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, src)
	dismissed, err := rc.state.IsDismissed(notestate.VaultGithub, canon)
	if err != nil || dismissed {
		t.Fatal("the unchecked row was dismissed")
	}
}

func TestAMidplanLockHoldsTheBurial(t *testing.T) {
	rc := newRepoCase(t)
	src := rc.repoNote("AI-Domain/Agents/Locked.md",
		"https://github.com/o/late-lock", withExtra("locked: true\n"))
	markRepoNote(rc, src)
	canon := links.NormalizeURL("https://github.com/o/late-lock")
	plan := &Plan{
		VaultKind: "repos",
		DeletionsRepos: []repobanish.PendingRepoBanishment{{
			URL: "https://github.com/o/late-lock", Canonical: canon,
			Title: "Locked", Marker: "🗑️", Door: "note tag", Path: src,
		}},
	}
	rep := ApplyScanPlan(plan, rc.vault, nil, rc.logs.log,
		Selection{}, VaultKindRepos, rc.cache, rc.state)
	if rep.Banished != 0 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, src)
	if !rc.logs.has("locked") {
		t.Fatal("missing lock log")
	}
}

func TestTheReposLedgerRepointReverseMaps(t *testing.T) {
	rc := newRepoCase(t)
	if repointRepoLedger(rc.cache, "https://github.com/o/x",
		filepath.Join(rc.vault, "AI-Domain", "Agents", "X.md"),
		"AI-Domain/Agents") {
		t.Fatal("no row — no re-point")
	}
	rc.cache.AddProcessed(303, "https://github.com/o/x", "o", "x",
		"nowhere.md", "Uncategorized")
	if !repointRepoLedger(rc.cache, "https://github.com/o/x",
		filepath.Join(rc.vault, "AI-Domain", "Agents", "X.md"),
		"AI-Domain/Agents") {
		t.Fatal("the row did not move")
	}
	row := rc.cache.ProcessedRowFor("https://github.com/o/x")
	if row == nil || row.Category == nil || *row.Category != "Agents" {
		t.Fatalf("row %+v", row)
	}
	// an unmapped folder keeps the row's own word
	if !repointRepoLedger(rc.cache, "https://github.com/o/x",
		"elsewhere.md", "Some/New/Place") {
		t.Fatal("the row did not move")
	}
	row2 := rc.cache.ProcessedRowFor("https://github.com/o/x")
	if row2 == nil || row2.Category == nil || *row2.Category != "Agents" {
		t.Fatalf("row2 %+v", row2)
	}
}

// --- TestTheReposRegexBrain -------------------------------------------

func TestTheDeclaredCategoryKeyFilesDeterministically(t *testing.T) {
	rc := newRepoCase(t)
	if err := os.MkdirAll(filepath.Join(rc.vault, "AI-Domain", "Agents"),
		0o755); err != nil {
		t.Fatal(err)
	}
	rc.repoNote("AI-Domain/Agents/Resident.md",
		"https://github.com/o/resident-brain")
	rc.repoNote("Orphan.md", "https://github.com/o/declared",
		withCategory("Agents"))
	plan := rc.buildPlan(nil, nil)
	if plan.RegexFiled != 1 || plan.LLMCandidates != 0 {
		t.Fatalf("plan %+v", plan)
	}
	if len(plan.Moves) != 1 || plan.Moves[0].To != "AI-Domain/Agents" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	if !rc.logs.has("regex brain") {
		t.Fatal("missing regex-brain log")
	}
}

func TestAStaleCategoryKeyStaysOnTheLLMPlate(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("Orphan.md", "https://github.com/o/stale",
		withCategory("MCP"))
	plan := rc.buildPlan(nil, nil)
	if plan.RegexFiled != 0 || plan.LLMCandidates != 1 {
		t.Fatalf("plan %+v", plan)
	}
}

func TestTheWeighedEvidenceSpeaksTheLanguage(t *testing.T) {
	rc := newRepoCase(t)
	if err := os.MkdirAll(filepath.Join(rc.vault, "Tools", "Scraping"),
		0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		rc.repoNote(fmt.Sprintf("Tools/Scraping/Py Res %d.md", i),
			fmt.Sprintf("https://github.com/o/py%d", i),
			withLanguage("Python"), withTags("scraping"))
	}
	if err := os.MkdirAll(filepath.Join(rc.vault, "Frameworks",
		"Backend"), 0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		rc.repoNote(fmt.Sprintf("Frameworks/Backend/Go Res %d.md", i),
			fmt.Sprintf("https://github.com/o/go%d", i),
			withLanguage("Go"))
	}
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	if cands := MoveCandidates(&inv, 40); len(cands) != 0 {
		t.Fatalf("all well-filed — cands %d", len(cands))
	}
	cand := NoteRow{Title: "py-tool", File: "Orphan.md", Org: "o",
		Stars: 10, PrimaryLang: "Python", Tags: []string{"scraping"},
		InRoot: true}
	ptrs := []*NoteRow{&cand}
	withHomes := AttachWeighedHomes(ptrs, &inv, nil, nil, 40,
		VaultKindRepos)
	if withHomes != 1 {
		t.Fatalf("with homes %d", withHomes)
	}
	homes := ptrs[0].Homes
	if len(homes) == 0 || homes[0].Folder != "Tools/Scraping" {
		t.Fatalf("homes %+v", homes)
	}
	reasons := strings.Join(homes[0].Reasons, " ")
	if !strings.Contains(reasons, "written in Python") {
		t.Fatalf("reasons %q", reasons)
	}
	if !strings.Contains(reasons, "shared tags: scraping") {
		t.Fatalf("reasons %q", reasons)
	}
}

func TestTheDeclaredCategoryEchoWeighsItsOwnFolder(t *testing.T) {
	rc := newRepoCase(t)
	if err := os.MkdirAll(filepath.Join(rc.vault, "Tools", "Scraping"),
		0o755); err != nil {
		t.Fatal(err)
	}
	rc.repoNote("Tools/Scraping/Res.md", "https://github.com/o/scrape",
		withLanguage("Python"))
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, VaultKindRepos)
	cand := NoteRow{Title: "scraper", File: "Orphan.md", Org: "o",
		Stars: 5, PrimaryLang: "", Category: "Scraping", InRoot: true}
	ptrs := []*NoteRow{&cand}
	withHomes := AttachWeighedHomes(ptrs, &inv, nil, nil, 40,
		VaultKindRepos)
	if withHomes != 1 {
		t.Fatalf("with homes %d", withHomes)
	}
	if len(ptrs[0].Homes) == 0 ||
		ptrs[0].Homes[0].Folder != "Tools/Scraping" {
		t.Fatalf("homes %+v", ptrs[0].Homes)
	}
	if !strings.Contains(strings.Join(ptrs[0].Homes[0].Reasons, " "),
		"declared category's own folder") {
		t.Fatalf("reasons %+v", ptrs[0].Homes[0].Reasons)
	}
}

func TestTheWebsitesWeighingKeepsItsDomainLaw(t *testing.T) {
	rc := newRepoCase(t)
	if err := os.MkdirAll(filepath.Join(rc.vault, "Design", "UI_UX"),
		0o755); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		rc.repoNote(fmt.Sprintf("Design/UI_UX/Res %d.md", i),
			fmt.Sprintf("https://d.example/p%d", i),
			withCategory(""), withStars(""), withLanguage(""),
			withOrg(""))
	}
	inv := ScanVaultInventory(rc.vault, nil, nil, nil, "")
	cand := NoteRow{Title: "a page", File: "Orphan.md",
		Domain: "d.example", InRoot: true}
	ptrs := []*NoteRow{&cand}
	withHomes := AttachWeighedHomes(ptrs, &inv, nil, nil, 40, "")
	if withHomes != 1 {
		t.Fatalf("with homes %d", withHomes)
	}
	if !strings.Contains(strings.Join(ptrs[0].Homes[0].Reasons, " "),
		"settled sibling") {
		t.Fatalf("reasons %+v", ptrs[0].Homes[0].Reasons)
	}
}

// --- TestTheRoundTrip --------------------------------------------------

func TestAFullReposPlanBuildsAndApplies(t *testing.T) {
	rc := newRepoCase(t)
	rc.repoNote("Orphan.md", "https://github.com/o/round",
		withStars("900"), withLanguage("Go"), withOrg("round-org"))
	if err := os.MkdirAll(filepath.Join(rc.vault, "Tools", "Automation"),
		0o755); err != nil {
		t.Fatal(err)
	}
	rc.repoNote("Tools/Automation/Resident.md",
		"https://github.com/o/resident2",
		withTags("go", "automation"), withLanguage("Go"))
	llm := newFakeLLM(map[string]any{
		"new_folders": []any{},
		"moves": []any{map[string]any{"note": "Orphan.md",
			"to":     "Tools/Automation",
			"reason": "a Go automation repo", "confidence": "high"}},
		"summary": "one move"})
	plan := rc.buildPlan(llm.call, rc.state)
	if len(plan.Moves) != 1 || plan.Moves[0].To != "Tools/Automation" {
		t.Fatalf("moves %+v", plan.Moves)
	}
	// the ask rode the repos prompt
	system := llm.calls[0][0].Content
	if !strings.Contains(system, "GitHub repository notes") {
		t.Fatal("the ask did not ride the repos prompt")
	}
	user := llm.calls[0][1].Content
	if !strings.Contains(user, "⭐ 900") || !strings.Contains(user, "🔧 Go") {
		t.Fatalf("user message missing the repos evidence")
	}
	rep := ApplyScanPlan(&plan, rc.vault, nil, rc.logs.log,
		Selection{}, VaultKindRepos, rc.cache, nil)
	if rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	dst := filepath.Join(rc.vault, "Tools", "Automation", "Orphan.md")
	mustBeFile(t, dst)
	if !strings.Contains(readFileT(t, dst), "curated: true") {
		t.Fatal("no stamp")
	}
	// the second scan's radar skips the stamped note
	plan2 := rc.buildPlan(nil, rc.state)
	if plan2.LLMCandidates != 0 {
		t.Fatalf("candidates %d", plan2.LLMCandidates)
	}
}

func TestReposDryRunRehearsesWithoutTouchingTheDisk(t *testing.T) {
	rc := newRepoCase(t)
	src := rc.repoNote("Orphan.md", "https://github.com/o/rehearse")
	plan := &Plan{
		VaultKind:  "repos",
		NewFolders: []string{"Tools/New_Sub"},
		Moves: []MoveRow{{
			Note: "Orphan.md", Path: src, Title: "Orphan",
			From: "(vault root)", To: "Tools/New_Sub", As: "",
			Reason: "test",
		}},
	}
	dryrun.Enable()
	rep := ApplyScanPlan(plan, rc.vault, nil, rc.logs.log,
		Selection{}, VaultKindRepos, rc.cache, nil)
	dryrun.Disable()
	dryrun.Clear()
	if rep.NotesMoved != 1 {
		t.Fatalf("report %+v", rep)
	}
	mustBeFile(t, src)
	if _, err := os.Stat(filepath.Join(rc.vault, "Tools", "New_Sub")); err == nil {
		t.Fatal("the rehearsal created the folder")
	}
}
