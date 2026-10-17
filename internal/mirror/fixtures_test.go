// fixtures_test.go — the reference's own test_phase5 + test_phase6
// fixtures, transcribed byte-faithfully (the golden differential pins
// every byte; a transcription error surfaces as a golden diff, never
// silently).
package mirror

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// llmMessage aliases the LLM seam's message shape for the tests.
type llmMessage = llmclient.Message

// errBoom is the scripted LLM failure.
var errBoom = errors.New("boom")

// frozenTime is the frozen clock instant the golden pins.
func frozenTime() time.Time {
	return time.Date(2026, 1, 1, 12, 34, 56, 0, time.UTC)
}

// ghNote mirrors test_phase5._gh_note (the GitHub note shape).
func ghNote(src, title string, withBanner bool) string {
	cover := ""
	embed := ""
	if withBanner {
		cover = fmt.Sprintf("cover: attachments/banners/%s_banner.png\n", title)
		embed = fmt.Sprintf("![banner](attachments/banners/%s_banner.png)\n\n", title)
	}
	return "---\n" +
		"source: " + src + "\n" +
		"aliases: []\n" +
		"tags: [python, tools]\n" +
		"category: Agents\n" +
		cover +
		"managed_by: gitcurator\n" +
		"schema_version: 1\n" +
		"prompt_version: gh-v1\n" +
		"---\n" +
		"\n" +
		embed +
		"> [!info] Managed by GitCurator — machine-written note.\n" +
		"\n" +
		"# " + title + "\n" +
		"\n" +
		"What is it? Real body text.\n" +
		"\n" +
		"---\n" +
		"*Source: [" + src + "](" + src + ")*\n"
}

// wsNote mirrors test_phase5._ws_note (the website note shape).
func wsNote(src, name string) string {
	return "---\n" +
		"source: \"" + src + "\"\n" +
		"aliases: []\n" +
		"tags: [design]\n" +
		"category: \"Design\"\n" +
		"subcategory: \"UI\"\n" +
		"fetch_status: \"full\"\n" +
		"managed_by: \"gitcurator\"\n" +
		"schema_version: \"1\"\n" +
		"prompt_version: \"web-v1\"\n" +
		"---\n" +
		"\n" +
		"> [!info] Managed by GitCurator — machine-written note.\n" +
		"\n" +
		"# " + name + "\n" +
		"\n" +
		"Standout feature text.\n" +
		"\n" +
		"---\n" +
		"*Source: [" + src + "](" + src + ")*\n"
}

// githubNote6 mirrors test_phase6._github_note (the richer Phase-6
// GitHub shape with all the r01 sections).
func githubNote6(name, url, tags, tldr string) string {
	return "---\n" +
		"source: " + url + "\n" +
		"aliases: []\n" +
		"tags: [" + tags + "]\n" +
		"category: Developer Tools\n" +
		"stars: 120\n" +
		"org: acme\n" +
		"credibility_score: 80/100\n" +
		"date_processed: 2026-09-01\n" +
		"managed_by: gitcurator\n" +
		"schema_version: 2\n" +
		"prompt_version: g3\n" +
		"---\n" +
		"\n" +
		"# " + name + "\n" +
		"\n" +
		"> **TL;DR:** " + tldr + "\n" +
		"\n" +
		"**`acme/" + name + "`** · ⭐ 120 · 🔧 Python\n" +
		"\n" +
		"## What is it?\n" +
		"It automates the boring parts of demo datasets so tests stay small.\n" +
		"\n" +
		"## How does it work?\n" +
		"A single command walks a folder and emits JSON fixtures.\n" +
		"\n" +
		"## Why is it important? (Core Value)\n" +
		"Tests stop depending on hand-written data.\n" +
		"\n" +
		"## Key Features & Technologies\n" +
		"- fixture generation\n" +
		"- deterministic output\n" +
		"- zero dependencies\n" +
		"\n" +
		"## Difference from Others\n" +
		"Smaller and faster than the big frameworks.\n" +
		"\n" +
		"## 🏢 Organization & Credibility\n" +
		"- **Developer:** acme\n" +
		"\n" +
		"---\n" +
		"*Source: [GitHub](" + url + ")*\n"
}

// websiteNote6 mirrors test_phase6._website_note.
func websiteNote6(name, url, best string) string {
	return "---\n" +
		"source: " + url + "\n" +
		"aliases: []\n" +
		"tags: [design, colors]\n" +
		"category: Design\n" +
		"subcategory: Color Tools\n" +
		"date_processed: 2026-09-01\n" +
		"managed_by: gitcurator\n" +
		"schema_version: 2\n" +
		"prompt_version: w3\n" +
		"---\n" +
		"\n" +
		"# " + name + "\n" +
		"\n" +
		"> **TL;DR:** Generates color palettes from an image.\n" +
		"\n" +
		"## One-line description\n" +
		"Generates color palettes from any image.\n" +
		"\n" +
		"## Best used for\n" +
		best + "\n" +
		"\n" +
		"## Pricing & sign-up\n" +
		"- Pricing: freemium\n" +
		"- Login required: no\n" +
		"\n" +
		"---\n" +
		"*Source: [" + url + "](" + url + ")*\n"
}

// mirrorNote6 mirrors test_phase6._mirror_note.
func mirrorNote6(source, name, relatedBlock string) string {
	body := "---\nsource: " + source + "\nmirror_of: " + source + "\n" +
		"managed_by: gitcurator\n---\n\n" +
		"> 🔒 Read-only library copy.\n\n# " + name + "\n\nBody.\n"
	if relatedBlock != "" {
		body += "\n" + relatedBlock + "\n"
	}
	return body
}

// writeFixture is test_phase5._write.
func writeFixture(t *testing.T, path, text string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// makeVault mirrors test_phase6._make_vault: notes = (rel, content).
func makeVault(t *testing.T, root string, notes [][2]string) string {
	t.Helper()
	for _, n := range notes {
		writeFixture(t, filepath.Join(root, filepath.FromSlash(n[0])), n[1])
	}
	return root
}

// buildGhVault mirrors test_phase5._build_gh_vault: 2 notes + a _review
// note + noise that must be skipped.
func buildGhVault(t *testing.T, root string, banner bool) string {
	t.Helper()
	writeFixture(t, filepath.Join(root, "Agents", "Frameworks", "Repo.md"),
		ghNote("https://github.com/owner/repo", "Repo", banner))
	writeFixture(t, filepath.Join(root, "Dev-Tools", "Cli_Tool.md"),
		ghNote("https://github.com/other/cli", "Cli_Tool", false))
	writeFixture(t, filepath.Join(root, "_review", "Pending.md"),
		ghNote("https://github.com/owner/pending", "Pending", false))
	// noise: skipped folders + a note without a source line
	writeFixture(t, filepath.Join(root, "_moc", "map.md"), "moc\n")
	writeFixture(t, filepath.Join(root, "_inbox", "inbox.md"), "inbox\n")
	writeFixture(t, filepath.Join(root, "attachments", "banners", "x.png"), "png")
	writeFixture(t, filepath.Join(root, ".obsidian", "app.json"), "{}")
	writeFixture(t, filepath.Join(root, "Agents", "no_source_note.md"), "no source\n")
	return root
}

// buildWsVault mirrors test_phase5._build_ws_vault.
func buildWsVault(t *testing.T, root string) string {
	t.Helper()
	writeFixture(t, filepath.Join(root, "Design", "UI", "Tool.md"),
		wsNote("https://example.com/tool", "Tool"))
	writeFixture(t, filepath.Join(root, "Learning", "Courses", "Course.md"),
		wsNote("https://example.com/course", "Course"))
	return root
}

// buildManual mirrors test_phase5._build_manual: an owner's manual
// vault with precious files everywhere.
func buildManual(t *testing.T, root string) string {
	t.Helper()
	writeFixture(t, filepath.Join(root, "My Idea.md"), "my precious idea\n")
	writeFixture(t, filepath.Join(root, "Journal", "2026-09-29.md"), "journal\n")
	_ = os.MkdirAll(filepath.Join(root, "Library"), 0o755)
	writeFixture(t, filepath.Join(root, "Library", "owner-note.md"),
		"an owner file directly under Library (never touched)\n")
	writeFixture(t, filepath.Join(root, "Library", GitHubTree, "my-own-note.md"),
		"owner file inside the GitHub tree (no marker — never touched)\n")
	return root
}

// snapshot is test_phase5._snapshot: rel path → bytes for EVERY file.
func snapshot(t *testing.T, root string) map[string]string {
	t.Helper()
	out := map[string]string{}
	_ = filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		data, readErr := os.ReadFile(path)
		if readErr != nil {
			return nil
		}
		rel, relErr := filepath.Rel(root, path)
		if relErr != nil {
			return nil
		}
		out[strings.ReplaceAll(rel, "\\", "/")] = string(data)
		return nil
	})
	return out
}

// vaultCase is test_phase5._VaultCase: three temp vaults wired the way
// the real ones are.
type vaultCase struct {
	tmp    string
	gh     string
	ws     string
	manual string
}

func newVaultCase(t *testing.T) *vaultCase {
	t.Helper()
	tmp := t.TempDir()
	return &vaultCase{
		tmp:    tmp,
		gh:     buildGhVault(t, filepath.Join(tmp, "ghvault"), true),
		ws:     buildWsVault(t, filepath.Join(tmp, "wsvault")),
		manual: buildManual(t, filepath.Join(tmp, "manualvault")),
	}
}

func (vc *vaultCase) lib(parts ...string) string {
	return filepath.Join(append([]string{vc.manual, LibraryFolder}, parts...)...)
}

func (vc *vaultCase) libTree(tree, rel string) string {
	return vc.lib(append([]string{tree}, strings.Split(rel, "/")...)...)
}

func (vc *vaultCase) apply(t *testing.T) *MirrorPlan {
	t.Helper()
	plan, err := RunMirror(vc.manual, vc.gh, vc.ws, true, nil)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

func (vc *vaultCase) dry(t *testing.T) *MirrorPlan {
	t.Helper()
	plan, err := RunMirror(vc.manual, vc.gh, vc.ws, false, nil)
	if err != nil {
		t.Fatal(err)
	}
	return plan
}

// frozenClock freezes the package clock seam for a test.
func frozenClock(t *testing.T, at time.Time) {
	t.Helper()
	old := clockNow
	clockNow = func() time.Time { return at }
	t.Cleanup(func() { clockNow = old })
}

// pyNowStr is the frozen strftime the golden pins
// ('2026-01-01 12:34:56' seconds / '2026-01-01 12:34' minutes).
const (
	frozenStampSeconds = "2026-01-01 12:34:56"
	frozenStampMinutes = "2026-01-01 12:34"
)

// mustPyDumps is the golden comparison writer.
func mustPyDumps(v any) string { return storage.PyDumps(v) }
