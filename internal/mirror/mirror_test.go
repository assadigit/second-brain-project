// mirror_test.go — the ported test_phase5 selections (the core, minus
// the tool/CLI/GUI legs that belong to the Phase-5 CLI task):
// build_mirror_note, the safety refusals, _safe_join, the
// nothing-outside-Library acceptance, and the sync semantics.
package mirror

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// -- TestBuildMirrorNote -----------------------------------------------------

func TestBuildMirrorNoteMirrorOfInsertedBeforeFrontmatterClose(t *testing.T) {
	out := BuildMirrorNote(ghNote("https://github.com/a/b", "Repo", false),
		"https://github.com/a/b", MirrorBanners[GitHubTree])
	parts := strings.SplitN(out, "---", 3)
	head := parts[1] // the front-matter block
	if !strings.Contains(head, MirrorKey+": https://github.com/a/b") {
		t.Fatalf("mirror_of line missing from front matter:\n%s", head)
	}
	if !strings.Contains(head, "managed_by: gitcurator") {
		t.Fatalf("managed_by line missing:\n%s", head)
	}
	if !strings.HasSuffix(strings.TrimSpace(head),
		MirrorKey+": https://github.com/a/b") { // last FM key
		t.Fatalf("mirror_of is not the last front-matter key:\n%s", head)
	}
}

func TestBuildMirrorNoteBannerIsTheFirstBodyLine(t *testing.T) {
	out := BuildMirrorNote(ghNote("https://github.com/a/b", "Repo", false),
		"https://github.com/a/b", MirrorBanners[GitHubTree])
	parts := strings.SplitN(out, "---", 3)
	body := strings.TrimLeft(parts[2], "\n")
	if !strings.HasPrefix(body, MirrorBanners[GitHubTree]) {
		t.Fatalf("banner is not the first body line:\n%q", body[:min(len(body), 80)])
	}
	if !strings.Contains(out, "> [!info] Managed by GitCurator") {
		t.Fatal("body callout lost")
	}
}

func TestBuildMirrorNoteNoteWithoutFrontmatterGetsWrapped(t *testing.T) {
	out := BuildMirrorNote("# Just a heading\n\nbody\n",
		"https://example.com/x", MirrorBanners[WebsitesTree])
	if !strings.HasPrefix(out, "---\n") {
		t.Fatalf("no fresh front matter:\n%q", out[:min(len(out), 40)])
	}
	if !strings.Contains(out, MirrorKey+": https://example.com/x") {
		t.Fatal("mirror_of missing")
	}
	if !strings.Contains(out, MirrorBanners[WebsitesTree]) {
		t.Fatal("banner missing")
	}
	if !strings.Contains(out, "# Just a heading") {
		t.Fatal("body lost")
	}
}

func TestBuildMirrorNoteBannerReferencesAreStripped(t *testing.T) {
	out := BuildMirrorNote(
		ghNote("https://github.com/a/b", "Repo", true),
		"https://github.com/a/b", MirrorBanners[GitHubTree])
	for _, needle := range []string{"cover:", "![banner](", "attachments/banners"} {
		if strings.Contains(out, needle) {
			t.Fatalf("banner reference %q survived:\n%s", needle, out)
		}
	}
}

func TestBuildMirrorNoteBodyContentIsPreservedVerbatim(t *testing.T) {
	src := ghNote("https://github.com/a/b", "Repo", false)
	out := BuildMirrorNote(src, "https://github.com/a/b", MirrorBanners[GitHubTree])
	for _, needle := range []string{
		"# Repo", "What is it? Real body text.",
		"*Source: [https://github.com/a/b]", "tags: [python, tools]",
		"prompt_version: gh-v1",
	} {
		if !strings.Contains(out, needle) {
			t.Fatalf("body needle %q lost", needle)
		}
	}
}

func TestBuildMirrorNoteBodyWithoutBlankAfterFrontmatterStillGetsOne(t *testing.T) {
	src := "---\nsource: https://x/y\n---\n# Tight\n"
	out := BuildMirrorNote(src, "https://x/y", MirrorBanners[WebsitesTree])
	parts := strings.Split(out, "---\n")
	body := parts[len(parts)-1]
	if !strings.HasPrefix(body, "\n"+MirrorBanners[WebsitesTree]) {
		t.Fatalf("no blank line before the banner:\n%q", body)
	}
}

// -- TestSafetyRefusals ------------------------------------------------------

func TestSafetyRefusalsManualNotSet(t *testing.T) {
	vc := newVaultCase(t)
	if err := CheckVaultPaths("", vc.gh, vc.ws, false); err == nil {
		t.Fatal("empty manual vault accepted")
	}
}

func TestSafetyRefusalsManualEqualsGithubVault(t *testing.T) {
	vc := newVaultCase(t)
	if _, err := RunMirror(vc.gh, vc.gh, vc.ws, false, nil); err == nil {
		t.Fatal("manual == github accepted")
	}
	if err := CheckVaultPaths(vc.gh, vc.gh, "", false); err == nil {
		t.Fatal("check_vault_paths accepted equality")
	}
}

func TestSafetyRefusalsManualContainsGithubVault(t *testing.T) {
	vc := newVaultCase(t)
	nested := filepath.Join(vc.manual, "gh-inside")
	if err := CheckVaultPaths(vc.manual, nested, vc.ws, false); err == nil {
		t.Fatal("manual containing the machine vault accepted")
	}
}

func TestSafetyRefusalsManualInsideGithubVault(t *testing.T) {
	vc := newVaultCase(t)
	inside := filepath.Join(vc.gh, "sub", "manual")
	if err := CheckVaultPaths(inside, vc.gh, vc.ws, false); err == nil {
		t.Fatal("manual inside the machine vault accepted")
	}
}

func TestSafetyRefusalsManualOverlapsWebsitesVaultToo(t *testing.T) {
	vc := newVaultCase(t)
	if err := CheckVaultPaths(vc.ws, vc.gh, vc.ws, false); err == nil {
		t.Fatal("manual == websites accepted")
	}
}

func TestSafetyRefusalsBothMachineVaultsUnset(t *testing.T) {
	vc := newVaultCase(t)
	if err := CheckVaultPaths(vc.manual, "", "", false); err == nil {
		t.Fatal("no machine vault accepted")
	}
}

func TestSafetyRefusalsApplyRequiresExistingManualFolder(t *testing.T) {
	vc := newVaultCase(t)
	missing := filepath.Join(vc.tmp, "not-created-yet")
	if err := CheckVaultPaths(missing, vc.gh, vc.ws, false); err != nil {
		t.Fatalf("dry run refused: %v", err)
	}
	if _, err := RunMirror(missing, vc.gh, vc.ws, true, nil); err == nil {
		t.Fatal("apply on a missing manual vault accepted")
	}
}

func TestSafetyRefusalsLibraryPathIsAFile(t *testing.T) {
	tmp := t.TempDir()
	writeFixture(t, filepath.Join(tmp, "m2", LibraryFolder), "a file")
	if _, err := RunMirror(filepath.Join(tmp, "m2"), filepath.Join(tmp, "gh"),
		filepath.Join(tmp, "ws"), false, nil); err == nil {
		t.Fatal("Library/ as a file accepted")
	}
}

// -- TestSafeJoin ------------------------------------------------------------

func TestSafeJoinDotdotRejected(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"../escape.md", "Agents/../../escape.md"} {
		if _, err := SafeJoin(root, rel); err == nil {
			t.Fatalf("traversal %q accepted", rel)
		}
	}
}

func TestSafeJoinAbsoluteRejected(t *testing.T) {
	root := t.TempDir()
	if _, err := SafeJoin(root, "/etc/passwd"); err == nil {
		t.Fatal("absolute path accepted")
	}
}

func TestSafeJoinBackslashAndDriveRejected(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{`C:\escape.md`, "C:/escape.md"} {
		if _, err := SafeJoin(root, rel); err == nil {
			t.Fatalf("drive-letter path %q accepted", rel)
		}
	}
}

func TestSafeJoinEmptyAndDotComponentsRejected(t *testing.T) {
	root := t.TempDir()
	for _, rel := range []string{"", ".", "Agents//x.md", "Agents/./x.md"} {
		if _, err := SafeJoin(root, rel); err == nil {
			t.Fatalf("component path %q accepted", rel)
		}
	}
}

func TestSafeJoinGoodPathResolvesInsideRoot(t *testing.T) {
	root := t.TempDir()
	target, err := SafeJoin(root, "Agents/Frameworks/Repo.md")
	if err != nil {
		t.Fatal(err)
	}
	if !isInside(target, root) {
		t.Fatalf("target %q not inside %q", target, root)
	}
	if target != filepath.Join(root, "Agents", "Frameworks", "Repo.md") {
		t.Fatalf("target %q unexpected", target)
	}
}

// -- TestNothingOutsideLibrary ----------------------------------------------

func TestNothingOutsideLibraryApplyTouchesOnlyLibrary(t *testing.T) {
	vc := newVaultCase(t)
	before := snapshot(t, vc.manual)
	outsideBefore := map[string]string{}
	for k, v := range before {
		if !strings.HasPrefix(k, LibraryFolder+"/") {
			outsideBefore[k] = v
		}
	}
	plan := vc.apply(t)
	if !plan.Applied {
		t.Fatal("plan not applied")
	}
	after := snapshot(t, vc.manual)
	outsideAfter := map[string]string{}
	for k, v := range after {
		if !strings.HasPrefix(k, LibraryFolder+"/") {
			outsideAfter[k] = v
		}
	}
	if len(outsideBefore) != len(outsideAfter) {
		t.Fatalf("outside-Library file set changed: %v vs %v",
			outsideBefore, outsideAfter)
	}
	for k, v := range outsideBefore {
		if outsideAfter[k] != v {
			t.Fatalf("outside-Library file %q changed", k)
		}
	}
	// nothing new outside Library/
	for k := range after {
		if _, old := before[k]; !old && !strings.HasPrefix(k, LibraryFolder+"/") {
			t.Fatalf("new file outside Library/: %q", k)
		}
	}
	// the owner files INSIDE Library/ survive untouched too
	if before["Library/owner-note.md"] != after["Library/owner-note.md"] {
		t.Fatal("owner file under Library/ changed")
	}
	if before["Library/"+GitHubTree+"/my-own-note.md"] !=
		after["Library/"+GitHubTree+"/my-own-note.md"] {
		t.Fatal("owner file inside the GitHub tree changed")
	}
}

func TestNothingOutsideLibraryDryRunWritesNothingAnywhere(t *testing.T) {
	vc := newVaultCase(t)
	before := snapshot(t, vc.manual)
	plan := vc.dry(t)
	if plan.Applied {
		t.Fatal("dry run claims applied")
	}
	if plan.TotalChanges() <= 0 {
		t.Fatalf("dry run found no changes: %d", plan.TotalChanges())
	}
	after := snapshot(t, vc.manual)
	if len(before) != len(after) {
		t.Fatalf("dry run changed the tree: %d vs %d files", len(before), len(after))
	}
	for k, v := range before {
		if after[k] != v {
			t.Fatalf("dry run changed %q", k)
		}
	}
}

func TestNothingOutsideLibraryDeletionsNeverRemoveUnmarkedFiles(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	orphan := vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md")
	if err := os.Remove(filepath.Join(vc.gh, "Agents", "Frameworks", "Repo.md")); err != nil {
		t.Fatal(err)
	}
	plan := vc.apply(t)
	if _, err := os.Stat(orphan); err == nil {
		t.Fatal("orphan mirror copy survived the delete pass")
	}
	// owner files survive the deletion pass
	if _, err := os.Stat(vc.lib(GitHubTree, "my-own-note.md")); err != nil {
		t.Fatal("owner file inside the GitHub tree deleted")
	}
	if _, err := os.Stat(vc.lib("owner-note.md")); err != nil {
		t.Fatal("owner file under Library/ deleted")
	}
	if !strings.Contains(plan.RenderReport(), "DELETE") {
		t.Fatal("report does not mention the deletion")
	}
}

func TestNothingOutsideLibraryConflictWhenUnmarkedFileIsInTheWay(t *testing.T) {
	vc := newVaultCase(t)
	blocked := vc.lib(GitHubTree, "Dev-Tools", "Cli_Tool.md")
	writeFixture(t, blocked, "an owner file exactly where a mirror would go\n")
	before, _ := os.ReadFile(blocked)
	plan := vc.apply(t)
	after, _ := os.ReadFile(blocked)
	if string(before) != string(after) {
		t.Fatal("the blocking owner file was touched")
	}
	tree := plan.Trees[GitHubTree]
	if len(tree.Conflicts) != 1 || tree.Conflicts[0].Rel != "Dev-Tools/Cli_Tool.md" {
		t.Fatalf("unexpected conflicts: %+v", tree.Conflicts)
	}
}

func TestNothingOutsideLibraryNoMarkerNoDeleteEvenForLookalikes(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	// an owner file that COPIES a mirror note but drops the marker
	lookalike := vc.lib(GitHubTree, "Agents", "Frameworks", "fake.md")
	original, _ := os.ReadFile(vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md"))
	writeFixture(t, lookalike, strings.ReplaceAll(string(original), MirrorKey+":", "xmirror:"))
	if err := os.Remove(filepath.Join(vc.gh, "Agents", "Frameworks", "Repo.md")); err != nil {
		t.Fatal(err)
	}
	vc.apply(t)
	if _, err := os.Stat(lookalike); err != nil {
		t.Fatal("unmarked lookalike deleted") // no marker → survives
	}
}

// -- TestSync ------------------------------------------------------------------

func TestSyncGithubNotesMirrorUnderLibraryGithubProjects(t *testing.T) {
	vc := newVaultCase(t)
	plan := vc.apply(t)
	tree := plan.Trees[GitHubTree]
	if tree.Sources != 3 { // 2 category + 1 _review
		t.Fatalf("sources = %d, want 3", tree.Sources)
	}
	if len(tree.Creates) != 3 {
		t.Fatalf("creates = %d, want 3", len(tree.Creates))
	}
	for _, rel := range []string{
		"Agents/Frameworks/Repo.md", "Dev-Tools/Cli_Tool.md", "_review/Pending.md",
	} {
		path := vc.libTree(GitHubTree, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("mirror missing: %s (%v)", rel, err)
		}
		content := string(data)
		if !strings.Contains(content, MirrorKey+": https://") {
			t.Fatalf("%s: no mirror_of marker", rel)
		}
		if !strings.Contains(content, MirrorBanners[GitHubTree]) {
			t.Fatalf("%s: no banner", rel)
		}
	}
}

func TestSyncWebsitesNotesMirrorUnderLibraryWebsites(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	for _, rel := range []string{"Design/UI/Tool.md", "Learning/Courses/Course.md"} {
		path := vc.libTree(WebsitesTree, rel)
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatalf("mirror missing: %s (%v)", rel, err)
		}
		if !strings.Contains(string(data), MirrorKey+": https://example.com/") {
			t.Fatalf("%s: no mirror_of marker", rel)
		}
	}
}

func TestSyncSkippedFoldersAndSourcelessNotesNeverMirror(t *testing.T) {
	vc := newVaultCase(t)
	plan := vc.apply(t)
	for _, rel := range []string{
		"_moc/map.md", "_inbox/inbox.md", ".obsidian/app.json",
		"Agents/no_source_note.md",
	} {
		if _, err := os.Stat(vc.libTree(GitHubTree, rel)); err == nil {
			t.Fatalf("skipped noise mirrored: %s", rel)
		}
	}
	tree := plan.Trees[GitHubTree]
	if tree.SkippedNoSource != 1 {
		t.Fatalf("skipped_no_source = %d, want 1", tree.SkippedNoSource)
	}
	if tree.Sources != 3 {
		t.Fatalf("sources = %d, want 3", tree.Sources)
	}
}

func TestSyncIdempotentSecondRunZeroChanges(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	before := snapshot(t, vc.manual)
	plan := vc.apply(t)
	if plan.TotalChanges() != 0 {
		t.Fatalf("second run found %d changes", plan.TotalChanges())
	}
	tree := plan.Trees[GitHubTree]
	if tree.Keeps != 3 {
		t.Fatalf("keeps = %d, want 3", tree.Keeps)
	}
	after := snapshot(t, vc.manual)
	if len(before) != len(after) {
		t.Fatal("second run changed the file set")
	}
	for k, v := range before {
		if after[k] != v {
			t.Fatalf("second run changed %q", k)
		}
	}
}

func TestSyncSourceEditUpdatesTheMirror(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	writeFixture(t, filepath.Join(vc.gh, "Agents", "Frameworks", "Repo.md"),
		ghNote("https://github.com/owner/repo", "Repo", false)+
			"\nAn extra paragraph the LLM wrote.\n")
	plan := vc.apply(t)
	if len(plan.Trees[GitHubTree].Updates) != 1 {
		t.Fatalf("updates = %d, want 1", len(plan.Trees[GitHubTree].Updates))
	}
	mirrored, _ := os.ReadFile(vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md"))
	if !strings.Contains(string(mirrored), "An extra paragraph the LLM wrote.") {
		t.Fatal("edit did not reach the mirror")
	}
}

func TestSyncHumanEditToAMirrorCopyIsRestored(t *testing.T) {
	// Mirror files are policy read-only (rebuildable) — a sync repairs
	// them. The banner says so.
	vc := newVaultCase(t)
	vc.apply(t)
	path := vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md")
	current, _ := os.ReadFile(path)
	writeFixture(t, path, string(current)+"\nmy edit\n")
	plan := vc.apply(t)
	if len(plan.Trees[GitHubTree].Updates) != 1 {
		t.Fatalf("updates = %d, want 1", len(plan.Trees[GitHubTree].Updates))
	}
	restored, _ := os.ReadFile(path)
	if strings.Contains(string(restored), "my edit") {
		t.Fatal("human edit survived the repair")
	}
}

func TestSyncMovesPropagateMatchedBySource(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	old := vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md")
	new := vc.lib(GitHubTree, "Infrastructure", "Repo.md")
	_ = os.MkdirAll(filepath.Join(vc.gh, "Infrastructure"), 0o755)
	if err := os.Rename(
		filepath.Join(vc.gh, "Agents", "Frameworks", "Repo.md"),
		filepath.Join(vc.gh, "Infrastructure", "Repo.md")); err != nil {
		t.Fatal(err)
	}
	plan := vc.apply(t)
	tree := plan.Trees[GitHubTree]
	if len(tree.Moves) != 1 {
		t.Fatalf("moves = %d, want 1", len(tree.Moves))
	}
	if tree.Moves[0].Old != "Agents/Frameworks/Repo.md" ||
		tree.Moves[0].New != "Infrastructure/Repo.md" {
		t.Fatalf("unexpected move: %+v", tree.Moves[0])
	}
	if _, err := os.Stat(old); err == nil {
		t.Fatal("stale copy survived")
	}
	data, err := os.ReadFile(new)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(data), MirrorKey+": https://github.com/owner/repo") {
		t.Fatal("moved copy lost its marker")
	}
}

func TestSyncMoveSwapTwoNotesResolves(t *testing.T) {
	// A moves into B's old folder while B moves elsewhere — deletions
	// run before writes, so both end up correct.
	vc := newVaultCase(t)
	vc.apply(t)
	aSrc := filepath.Join(vc.gh, "Agents", "Frameworks", "Repo.md")
	bSrc := filepath.Join(vc.gh, "Dev-Tools", "Cli_Tool.md")
	aDst := filepath.Join(vc.gh, "Dev-Tools", "Repo.md")
	bDst := filepath.Join(vc.gh, "Agents", "Frameworks", "Cli_Tool.md")
	if err := os.Rename(aSrc, aDst); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(bSrc, bDst); err != nil {
		t.Fatal(err)
	}
	vc.apply(t)
	if _, err := os.Stat(vc.lib(GitHubTree, "Dev-Tools", "Repo.md")); err != nil {
		t.Fatal("swapped A missing")
	}
	if _, err := os.Stat(vc.lib(GitHubTree, "Agents", "Frameworks", "Cli_Tool.md")); err != nil {
		t.Fatal("swapped B missing")
	}
	// stale copies gone
	if _, err := os.Stat(vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md")); err == nil {
		t.Fatal("A's stale copy survived")
	}
	if _, err := os.Stat(vc.lib(GitHubTree, "Dev-Tools", "Cli_Tool.md")); err == nil {
		t.Fatal("B's stale copy survived")
	}
}

func TestSyncOrphanedMirrorRemovedOnlyWithMarker(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	orphan := vc.lib(WebsitesTree, "Design", "UI", "Tool.md")
	if err := os.Remove(filepath.Join(vc.ws, "Design", "UI", "Tool.md")); err != nil {
		t.Fatal(err)
	}
	plan := vc.apply(t)
	if _, err := os.Stat(orphan); err == nil {
		t.Fatal("orphaned mirror survived")
	}
	if plan.Trees[WebsitesTree].Sources != 1 {
		t.Fatalf("websites sources = %d, want 1", plan.Trees[WebsitesTree].Sources)
	}
}

func TestSyncDuplicateSourceFilesReportedOnceMirrored(t *testing.T) {
	vc := newVaultCase(t)
	writeFixture(t, filepath.Join(vc.gh, "Dev-Tools", "Repo_copy.md"),
		ghNote("https://github.com/owner/repo", "Repo", false))
	plan := vc.dry(t)
	tree := plan.Trees[GitHubTree]
	if tree.Sources != 3 { // dedup: one per source
		t.Fatalf("sources = %d, want 3", tree.Sources)
	}
	if len(tree.Duplicates) != 1 || tree.Duplicates[0] != "Dev-Tools/Repo_copy.md" {
		t.Fatalf("unexpected duplicates: %v", tree.Duplicates)
	}
	if len(tree.Creates) != 3 {
		t.Fatalf("creates = %d, want 3", len(tree.Creates))
	}
}

func TestSyncDuplicateMarkerCopiesCollapseToOne(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	good := vc.lib(GitHubTree, "Agents", "Frameworks", "Repo.md")
	extra := vc.lib(GitHubTree, "zz-dup", "Repo.md")
	data, _ := os.ReadFile(good)
	writeFixture(t, extra, string(data))
	plan := vc.apply(t)
	if _, err := os.Stat(extra); err == nil {
		t.Fatal("duplicate marker copy survived") // marker dup removed
	}
	if _, err := os.Stat(good); err != nil {
		t.Fatal("the good copy was removed")
	}
	found := false
	for _, d := range plan.Trees[GitHubTree].Deletes {
		if d == "zz-dup/Repo.md" {
			found = true
		}
	}
	if !found {
		t.Fatalf("zz-dup/Repo.md not in deletes: %v", plan.Trees[GitHubTree].Deletes)
	}
}

func TestSyncEmptyDirsCleanedButOwnerContentDirsStay(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	if err := os.Remove(filepath.Join(vc.gh, "Agents", "Frameworks", "Repo.md")); err != nil {
		t.Fatal(err)
	}
	vc.apply(t)
	if _, err := os.Stat(vc.lib(GitHubTree, "Agents", "Frameworks")); err == nil {
		t.Fatal("emptied mirror dir survived")
	}
	// Library root survives: it still holds the owner's file
	if fi, err := os.Stat(vc.lib()); err != nil || !fi.IsDir() {
		t.Fatal("Library root removed")
	}
	if _, err := os.Stat(vc.lib("owner-note.md")); err != nil {
		t.Fatal("owner file under Library/ removed")
	}
}

func TestSyncMissingMachineVaultReportsSkipped(t *testing.T) {
	vc := newVaultCase(t)
	plan, err := RunMirror(vc.manual, vc.gh, filepath.Join(vc.tmp, "nope"), false, nil)
	if err != nil {
		t.Fatal(err)
	}
	tree := plan.Trees[WebsitesTree]
	if tree.VaultPresent {
		t.Fatal("missing vault claimed present")
	}
	if tree.Sources != 0 {
		t.Fatalf("sources = %d, want 0", tree.Sources)
	}
	if !strings.Contains(plan.RenderReport(), "folder not found — skipped") {
		t.Fatal("report does not mention the skip")
	}
}

func TestSyncUnmarkedFilesCountedAndLeftAlone(t *testing.T) {
	vc := newVaultCase(t)
	vc.apply(t)
	writeFixture(t, vc.lib(WebsitesTree, "my-note.md"), "mine\n")
	plan := vc.dry(t)
	if plan.Trees[WebsitesTree].UnmarkedFiles != 1 {
		t.Fatalf("unmarked = %d, want 1", plan.Trees[WebsitesTree].UnmarkedFiles)
	}
	if plan.TotalChanges() != 0 {
		t.Fatalf("changes = %d, want 0", plan.TotalChanges())
	}
	data, _ := os.ReadFile(vc.lib(WebsitesTree, "my-note.md"))
	if string(data) != "mine\n" {
		t.Fatal("unmarked file touched")
	}
}
