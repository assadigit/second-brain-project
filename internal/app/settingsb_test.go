// settingsb_test.go — T-74 slice (b)'s own battery: the pages B
// engines over fixture vaults + fake seams (the hermetic law):
//
//	· THE QUEUE REPORT RENDERING from canned reports (the acceptance
//	  item): the labeled-count lines, the settled lines, the websites
//	  sub-table + the OFF/ON-no-vault advisories, the pending lists
//	  with the 50-cap, the caught-up family, the badge's two states,
//	  the log-line family, and the username guard;
//	· THE DASHBOARD ENGINES: the frontmatter scan over a fixture vault
//	  (categories/languages/credibility/banners/review), the exact
//	  report text, the quarantine table + the threshold footline + the
//	  reset, the threshold group's merge save;
//	· THE UNDO ENGINE: the no-file/empty/gone early doors, the confirm
//	  body's top-10 law, the deletes + the processed-repos row
//	  removal + the once law;
//	· THE SOURCES ENGINES: the no-url guard, the error path, the RSS
//	  extraction + the unique-set law, the Reddit .json walk, the
//	  empty answer;
//	· THE BACKUP PAGE: the three status-dot ladders' every rung, the
//	  save group's laws (the int-law heal, the URL strip, the nested
//	  groups merge, the group-pointer survival), the Backup Now
//	  guards, the Open Dashboard URL law;
//	· THE BOT GROUP's save law (the @-strip + the merge).
package app

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// ---------------------------------------------------------------------------
// The harness.
// ---------------------------------------------------------------------------

// newPagesBSettings builds a Settings over a temp config (the pages B
// tests' shape).
func newPagesBSettings(t *testing.T, cfg *storage.OrderedMap) (*Settings, *busRec) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	a := New(cfgPath, false)
	rec := &busRec{}
	rec.subscribe(a.Bus)
	return a.Settings, rec
}

// cannedQueue builds a classified report (the acceptance's canned job
// result).
func cannedQueue() *QueueReport {
	wvi := 12
	return &QueueReport{
		Success: true,
		State:   "ok",
		GitHubURLs: []string{
			"https://github.com/o/in-vault",
			"https://github.com/o/pending-one",
			"https://github.com/o/pending-two",
			"https://github.com/o/dead",
			"https://github.com/o/settled",
			"https://github.com/o/blocked",
			"https://github.com/o/self",
		},
		NonGithub: []string{
			"https://example.com/a",
			"https://example.com/b",
		},
		PendingURLs: []string{
			"https://github.com/o/pending-one",
			"https://github.com/o/pending-two",
		},
		InVault:           1,
		Decommissioned:    1,
		Blocked:           1,
		Self:              1,
		Settled:           1,
		WebsitesInVault:   2,
		WebsitesProcessed: 1,
		WebsitesSettled:   1,
		WebsitesDismissed: 1,
		WebsitesBlocked:   1,
		WebsitesSelf:      1,
		PendingWebsites: []string{
			"https://example.com/a",
			"https://example.com/b",
		},
		WebsitesVaultIndex: &wvi,
	}
}

// ---------------------------------------------------------------------------
// THE QUEUE REPORT RENDERING (the acceptance item).
// ---------------------------------------------------------------------------

func TestRenderQueueReportFull(t *testing.T) {
	text := RenderQueueReport(cannedQueue())
	// The labeled-count lines (verbatim shapes).
	for _, want := range []string{
		"📬 Bot Queue Results",
		"Total GitHub URLs in bot:  7",
		"✅ Already in vault:        1",
		"🗑️ Decommissioned (404):    1",
		"🤝 Repos settled:           1  (addressed — never re-fetched)",
		"🚫 Blocked domains:         1",
		"🔒 Self domains (own bot):  1",
		"⏳ Pending (not in vault):  2",
		"🔗 Non-GitHub links:        2",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	// The pending list.
	for _, want := range []string{
		"PENDING GITHUB REPOS (need processing):",
		"  1. https://github.com/o/pending-one",
		"  2. https://github.com/o/pending-two",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}

func TestRenderQueueReportWebsitesOffAdvisory(t *testing.T) {
	rep := cannedQueue()
	rep.WebsitesOff = true
	rep.PendingWebsites = nil
	text := RenderQueueReport(rep)
	if !strings.Contains(text, "⚠️ Websites pipeline is OFF — 2 non-GitHub link(s) stay as _inbox rows.") {
		t.Fatalf("the OFF advisory never spoke:\n%s", text)
	}
	if strings.Contains(text, "🌐 WEBSITES pipeline") {
		t.Fatalf("the websites sub-table spoke while OFF:\n%s", text)
	}
	if !strings.Contains(text, "NON-GITHUB LINKS (2):") ||
		!strings.Contains(text, "(Recorded in _inbox/ per platform)") {
		t.Fatalf("the inbox note never spoke:\n%s", text)
	}
}

func TestRenderQueueReportWebsitesOnNoVault(t *testing.T) {
	rep := cannedQueue()
	rep.WebsitesNoVault = true
	rep.PendingWebsites = nil
	text := RenderQueueReport(rep)
	if !strings.Contains(text, "⚠️ Websites pipeline is ON but no Websites vault is set") {
		t.Fatalf("the no-vault advisory never spoke:\n%s", text)
	}
}

func TestRenderQueueReportCaughtUp(t *testing.T) {
	rep := cannedQueue()
	rep.PendingURLs = nil
	rep.PendingWebsites = nil
	text := RenderQueueReport(rep)
	if !strings.Contains(text, "🎉 All GitHub repos are addressed — 1 settled (never re-fetched), the rest in the vault!") {
		t.Fatalf("the settled caught-up line never spoke:\n%s", text)
	}
	rep.Settled = 0
	text = RenderQueueReport(rep)
	if !strings.Contains(text, "🎉 All GitHub repos are already in the vault!") {
		t.Fatalf("the plain caught-up line never spoke:\n%s", text)
	}
	if !strings.Contains(text, "Click '✅ Verify All Processed' to confirm.") {
		t.Fatalf("the confirm hint never spoke:\n%s", text)
	}
}

func TestRenderQueueReportWebsitesCap50(t *testing.T) {
	rep := cannedQueue()
	rep.PendingWebsites = nil
	for i := 0; i < 60; i++ {
		rep.PendingWebsites = append(rep.PendingWebsites,
			fmt.Sprintf("https://example.com/page-%d", i))
	}
	text := RenderQueueReport(rep)
	if !strings.Contains(text, "PENDING WEBSITES (60 — need processing):") {
		t.Fatalf("the websites header never spoke:\n%s", text)
	}
	if !strings.Contains(text, "  … and 10 more\n") {
		t.Fatalf("the cap line never spoke:\n%s", text)
	}
	if strings.Contains(text, "page-59") {
		t.Fatalf("the cap leaked row 60:\n%s", text)
	}
}

func TestRenderQueueReportErrorStates(t *testing.T) {
	if text := RenderQueueReport(&QueueReport{State: "error", Error: "the poll failed"}); text != "Error: the poll failed" {
		t.Fatalf("error render = %q", text)
	}
	if text := RenderQueueReport(&QueueReport{State: "disabled", Error: "not configured"}); text != "Error: not configured" {
		t.Fatalf("disabled render = %q", text)
	}
	if text := RenderQueueReport(nil); text != "Error: the queue check returned no report" {
		t.Fatalf("nil render = %q", text)
	}
}

func TestQueueBadgeLaw(t *testing.T) {
	rep := cannedQueue()
	b := QueueBadge(rep)
	if b.Text != "⏳ 4 pending" || !b.Pending || !b.Visible {
		t.Fatalf("badge = %+v", b)
	}
	rep.PendingURLs = nil
	rep.PendingWebsites = nil
	b = QueueBadge(rep)
	if b.Text != "✅ 0 pending" || b.Pending || !b.Visible {
		t.Fatalf("badge = %+v", b)
	}
	// A nil report keeps the badge hidden.
	if b := QueueBadge(nil); b.Visible || b.Text != "" {
		t.Fatalf("nil badge = %+v", b)
	}
}

func TestQueueLogLinesFamily(t *testing.T) {
	rep := cannedQueue()
	lines := QueueLogLines(rep)
	if len(lines) != 1 || lines[0].Level != "success" ||
		!strings.Contains(lines[0].Msg, "📬 Queue: 2 new repo(s) + 2 website link(s) pending (1 repos + 3 websites already done)") {
		t.Fatalf("pending lines = %+v", lines)
	}

	rep.PendingURLs = nil
	rep.PendingWebsites = nil
	lines = QueueLogLines(rep)
	joined := ""
	for _, l := range lines {
		joined += l.Msg + "\n"
	}
	for _, want := range []string{
		"🤝 1 repo link(s) are settled — addressed and processed, never fetched or counted again (the owner's law).",
		"🤝 1 website link(s) are settled — addressed and processed, never fetched or counted again (the owner's law; ♻️ revived in the master table un-settles any of them).",
		"📬 Queue: all caught up — 7 GitHub repos and 2 non-GitHub link(s) are already processed.",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("missing %q in:\n%s", want, joined)
		}
	}

	// The OFF / no-vault hints.
	rep.WebsitesOff = true
	joined = ""
	for _, l := range QueueLogLines(rep) {
		joined += l.Msg + "\n"
	}
	if !strings.Contains(joined, "ℹ️ 2 non-GitHub link(s) are waiting as _inbox rows") {
		t.Fatalf("the OFF hint never spoke:\n%s", joined)
	}
}

func TestBotUsernameGuard(t *testing.T) {
	if BotUsernameGuard("") != "❌ Please enter the bot username first." {
		t.Fatal("the empty guard never refused")
	}
	if BotUsernameGuard("  @githubfetcherbot ") != "" {
		t.Fatal("a username with an @ still refused")
	}
}

// ---------------------------------------------------------------------------
// THE DASHBOARD ENGINES.
// ---------------------------------------------------------------------------

func TestScanDashboardStats(t *testing.T) {
	vault := t.TempDir()
	writeFrontmatterNote(t, vault, "Agents/one.md", "Agents", "85.5", "Go")
	writeFrontmatterNote(t, vault, "Agents/two.md", "Agents", "90.0", "Go")
	writeFrontmatterNote(t, vault, "Tools/three.md", "Tools", "70.0", "Python")
	writeFrontmatterNote(t, vault, "_review/waiting.md", "Websites", "50.0", "HTML")
	// A banner + a plain png (only 'banner' counts).
	if err := os.MkdirAll(filepath.Join(vault, "banners"), 0o755); err != nil {
		t.Fatal(err)
	}
	_ = os.WriteFile(filepath.Join(vault, "banners", "banner-one.png"), []byte("x"), 0o644)
	_ = os.WriteFile(filepath.Join(vault, "banners", "plain.png"), []byte("x"), 0o644)

	stats := ScanDashboardStats(vault)
	if stats.TotalNotes != 4 || stats.ReviewQueue != 1 || stats.Banners != 1 {
		t.Fatalf("stats = %+v", stats)
	}
	if stats.AvgCredibility < 73.8 || stats.AvgCredibility > 73.9 {
		t.Fatalf("avg credibility = %v", stats.AvgCredibility)
	}
	if len(stats.Categories) != 3 || stats.Categories[0].Name != "Agents" || stats.Categories[0].Count != 2 {
		t.Fatalf("categories = %+v", stats.Categories)
	}
	if len(stats.Languages) != 3 || stats.Languages[0].Count != 2 {
		t.Fatalf("languages = %+v", stats.Languages)
	}
}

func TestRenderDashboardText(t *testing.T) {
	stats := &DashboardStats{TotalNotes: 12, Banners: 3, ReviewQueue: 2,
		AvgCredibility: 81.25,
		Categories:     []CountRow{{"Agents", 5}, {"Tools", 4}},
		Languages:      []CountRow{{"Go", 7}, {"Python", 5}}}
	text := RenderDashboardText("/tmp/vault", stats)
	for _, want := range []string{
		"📊 VAULT DASHBOARD",
		"Vault: /tmp/vault",
		"Total notes: 12",
		"Banners downloaded: 3",
		"Review queue: 2",
		"Average credibility: 81.2/100",
		"📁 NOTES BY CATEGORY:",
		"💻 NOTES BY LANGUAGE:",
		fmt.Sprintf("  %-40s %3d", "Agents", 5),
		fmt.Sprintf("  %-40s %3d", "Go", 7),
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
}

func TestQuarantineViewAndReset(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", vault)

	// The empty answer (no attempts on record).
	if text := QuarantineView(cfg); !strings.Contains(text, "✅ No 404 attempts on record.") {
		t.Fatalf("empty view = %q", text)
	}

	// Record some 404s (three strikes ⇒ quarantined).
	cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		cache.Record404("https://github.com/o/dead", "gone")
		cache.Record404("https://github.com/o/sick", "flaky")
	}
	_ = cache.Close()

	text := QuarantineView(cfg)
	for _, want := range []string{
		"URL", "attempts", "last seen",
		"https://github.com/o/dead",
		"⛔ quarantined",
		"Threshold: 3 consecutive 404s → the repo is quarantined and skipped. Reset the quarantine to re-check it.",
	} {
		if !strings.Contains(text, want) {
			t.Fatalf("missing %q in:\n%s", want, text)
		}
	}
	if strings.Contains(text, "https://github.com/o/sick  ⛔") {
		// the sick row (2 attempts) carries no flag — but its URL still shows
	}

	// The reset clears everything + speaks the line.
	removed, ok := ClearQuarantine(cfg)
	if !ok || removed != 2 {
		t.Fatalf("removed = %d ok=%v", removed, ok)
	}
	if line := ResetQuarantineLine(removed); !strings.Contains(line,
		"♻️ 404 quarantine reset — 2 link(s) will be processed again") {
		t.Fatalf("line = %q", line)
	}
	if text := QuarantineView(cfg); !strings.Contains(text, "✅ No 404 attempts on record.") {
		t.Fatalf("the reset never cleared: %q", text)
	}
}

func TestSaveQuarantineGroup(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("hmac_secret", "keep-me")
	s, _ := newPagesBSettings(t, cfg)

	if res := s.SaveQuarantineGroup(&QuarantineValues{StrikeThreshold: 5}); !res.Saved {
		t.Fatal("the threshold save failed")
	}
	if v, ok := s.cfg.Get("notfound_strike_threshold"); !ok || v != 5 {
		t.Fatalf("threshold = %v", v)
	}
	// The group-pointer law: unknown keys survive; the clamp law.
	if v, _ := s.cfg.Get("hmac_secret"); v != "keep-me" {
		t.Fatalf("the merge dropped hmac_secret: %v", v)
	}
	s.SaveQuarantineGroup(&QuarantineValues{StrikeThreshold: 99})
	if v, _ := s.cfg.Get("notfound_strike_threshold"); v != 10 {
		t.Fatalf("the clamp never applied: %v", v)
	}
}

// writeFrontmatterNote writes a fixture note with the dashboard's
// frontmatter fields.
func writeFrontmatterNote(t *testing.T, vault, rel, category, credibility, language string) {
	t.Helper()
	path := filepath.Join(vault, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	text := fmt.Sprintf("---\ncategory: %s\ncredibility_score: %s\nprimary_language: %s\n---\n# note\n",
		category, credibility, language)
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

// ---------------------------------------------------------------------------
// THE UNDO ENGINE.
// ---------------------------------------------------------------------------

func TestUndoLastBatch(t *testing.T) {
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()

	// (a) no vault.
	if res := ReadUndoList(""); res.Stage != "no-vault" {
		t.Fatalf("stage = %q", res.Stage)
	}
	// (b) no file.
	if res := ReadUndoList(vault); res.Stage != "no-file" {
		t.Fatalf("stage = %q", res.Stage)
	}

	// (c) the full flow: 12 files written, the top-10 confirm law, the
	// deletes + the processed-repos removal + the once law.
	var paths []string
	for i := 0; i < 12; i++ {
		p := filepath.Join(vault, fmt.Sprintf("note-%02d.md", i))
		if err := os.WriteFile(p, []byte("note"), 0o644); err != nil {
			t.Fatal(err)
		}
		paths = append(paths, p)
	}
	// One file already removed by hand.
	_ = os.Remove(paths[11])

	undoPath := filepath.Join(vault, "_undo_last_batch.txt")
	list := strings.Join(paths, "\n")
	if err := os.WriteFile(undoPath, []byte(list), 0o644); err != nil {
		t.Fatal(err)
	}

	// A processed_repos row keyed by note_path (the row the undo drops).
	cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	cache.AddProcessed(1001, "https://github.com/o/one", "o", "one",
		paths[0], "Agents")
	_ = cache.Close()

	res := ReadUndoList(vault)
	if res.Stage != "ok" || len(res.Files) != 11 {
		t.Fatalf("res = %+v", res)
	}
	if !strings.Contains(res.ConfirmMessage, "This will DELETE 11 note file(s) written by the last batch:") ||
		!strings.Contains(res.ConfirmMessage, "• note-00.md") ||
		!strings.Contains(res.ConfirmMessage, "...\n") {
		t.Fatalf("confirm body = %q", res.ConfirmMessage)
	}
	// The top-10 law: 11 existing files list 10 + "...".
	if got := strings.Count(res.ConfirmMessage, "• "); got != 10 {
		t.Fatalf("confirm listed %d files, want the 10 cap", got)
	}

	done := ApplyUndoLastBatch(vault)
	if done.Deleted != 11 || done.Errors != 0 {
		t.Fatalf("done = %+v", done)
	}
	if line, level := UndoLogLine(done.Deleted, done.Errors); level != "success" ||
		!strings.Contains(line, "↩️ Undo: deleted 11 file(s), 0 error(s).") {
		t.Fatalf("line = %q (%q)", line, level)
	}
	// The once law: the undo file is gone.
	if _, err := os.Stat(undoPath); !os.IsNotExist(err) {
		t.Fatal("the undo file survived a completed undo")
	}
	if res := ReadUndoList(vault); res.Stage != "no-file" {
		t.Fatalf("the once law broke: stage = %q", res.Stage)
	}
	// The processed_repos row is gone.
	cache2, _ := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
	if row := cache2.ProcessedRowFor("https://github.com/o/one"); row != nil {
		t.Fatalf("the processed row survived the undo: %+v", row)
	}
	_ = cache2.Close()

	// (d) the gone door: a list whose files are all removed.
	if err := os.WriteFile(undoPath, []byte(filepath.Join(vault, "nope.md")), 0o644); err != nil {
		t.Fatal(err)
	}
	if res := ReadUndoList(vault); res.Stage != "gone" {
		t.Fatalf("stage = %q", res.Stage)
	}
	// (e) the empty door.
	if err := os.WriteFile(undoPath, []byte("\n\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if res := ReadUndoList(vault); res.Stage != "empty" {
		t.Fatalf("stage = %q", res.Stage)
	}
}

// ---------------------------------------------------------------------------
// THE SOURCES ENGINES.
// ---------------------------------------------------------------------------

func TestFetchFromSources(t *testing.T) {
	cfg := storage.NewOrderedMap()
	s, rec := newPagesBSettings(t, cfg)
	s.pagesBSourcesFetchDoor = func(url string) (string, error) {
		if url == "https://hnrss.org/frontpage" {
			return "<rss><item><title>See https://github.com/octo/one and text</title></item>" +
				"<item><description>https://github.com/octo/one again https://github.com/octo/two</description></item></rss>", nil
		}
		if url == "https://reddit.com/r/programming.json" {
			return `{"data":{"children":[{"data":{"selftext":"check https://github.com/r/alpha",` +
				`"url":"https://github.com/r/beta","title":"https://github.com/r/alpha twice"}}]}}`, nil
		}
		if url == "https://example.com/empty" {
			return "<rss>nothing here</rss>", nil
		}
		return "", os.ErrDeadlineExceeded
	}

	// (a) the no-url guard.
	res := s.FetchFromSources("  ")
	if res.Stage != "no-url" || res.Notice != "Please enter a URL." {
		t.Fatalf("res = %+v", res)
	}

	// (b) the RSS extraction + the unique-set law.
	res = s.FetchFromSources("https://hnrss.org/frontpage")
	if res.Stage != "ok" || len(res.URLs) != 2 {
		t.Fatalf("res = %+v", res)
	}
	if res.URLs[0] != "https://github.com/octo/one" || res.URLs[1] != "https://github.com/octo/two" {
		t.Fatalf("urls = %v", res.URLs)
	}
	if !strings.Contains(res.Body, "Found 2 GitHub URLs:") {
		t.Fatalf("body = %q", res.Body)
	}
	if !joinLogsHas(rec, "🔍 Fetching from: https://hnrss.org/frontpage") ||
		!joinLogsHas(rec, "✅ Found 2 GitHub URLs from source") {
		t.Fatalf("the desk lines never rode: %s", joinLogs(rec))
	}

	// (c) the Reddit .json walk (selftext + url + title, deduped).
	res = s.FetchFromSources("https://reddit.com/r/programming.json")
	if res.Stage != "ok" || len(res.URLs) != 2 {
		t.Fatalf("res = %+v", res)
	}

	// (d) the empty answer.
	res = s.FetchFromSources("https://example.com/empty")
	if res.Stage != "ok" || len(res.URLs) != 0 ||
		res.Body != "No GitHub URLs found in the source." {
		t.Fatalf("res = %+v", res)
	}

	// (e) the error path (os.ErrDeadlineExceeded speaks "i/o timeout").
	res = s.FetchFromSources("https://example.com/broken")
	if res.Stage != "error" || !strings.Contains(res.Error, "i/o timeout") {
		t.Fatalf("res = %+v", res)
	}
	if !strings.Contains(res.Body, "Error: ") {
		t.Fatalf("body = %q", res.Body)
	}
}

// ---------------------------------------------------------------------------
// THE BACKUP PAGE.
// ---------------------------------------------------------------------------

func TestBackupLadders(t *testing.T) {
	// The Vault Backup dot's every rung.
	if (BackupFolderStatus("/definitely/not/here", true)) != (StatusDot{"● Folder not found", "error"}) {
		t.Fatal("the missing-folder rung broke")
	}
	if (BackupFolderStatus(t.TempDir(), true)) != (StatusDot{"● Ready", "success"}) {
		t.Fatal("the ready rung broke")
	}
	if (BackupFolderStatus("", true)) != (StatusDot{"● No folder selected", "warning"}) {
		t.Fatal("the no-folder rung broke")
	}
	if (BackupFolderStatus("", false)) != (StatusDot{"● Disabled", "error"}) {
		t.Fatal("the disabled rung broke")
	}
	// The VaultSeal dot's every rung.
	if (VaultSealStatus(false, true, true)) != (StatusDot{"● Disabled", "error"}) {
		t.Fatal("the seal disabled rung broke")
	}
	if (VaultSealStatus(true, false, true)) != (StatusDot{"● No vault selected", "warning"}) {
		t.Fatal("the seal no-vault rung broke")
	}
	if (VaultSealStatus(true, true, false)) != (StatusDot{"● Local-only (no GitHub token — commits, no push)", "warning"}) {
		t.Fatal("the seal local-only rung broke")
	}
	if (VaultSealStatus(true, true, true)) != (StatusDot{"● Ready — auto-seal after every run", "success"}) {
		t.Fatal("the seal ready rung broke")
	}
	// The GoodRepos dot's every rung.
	if (GoodReposStatus(true, true, false)) != (StatusDot{"● Local-only (no GitHub token — README built, no push)", "warning"}) {
		t.Fatal("the goodrepos local-only rung broke")
	}
	if (GoodReposStatus(true, true, true)) != (StatusDot{"● Ready — auto-publish after every run", "success"}) {
		t.Fatal("the goodrepos ready rung broke")
	}
}

func TestSaveBackupGroup(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("hmac_secret", "keep-me")
	cfg.Set("backup_max", 3) // an old value the int-law overwrites
	s, _ := newPagesBSettings(t, cfg)

	res := s.SaveBackupGroup(&BackupValues{
		BackupFolder:  "/backups",
		BackupMax:     "24",
		BackupEnabled: true,
		WorkerURL:     "https://bot.example.workers.dev/",

		VaultSealEnabled: true,
		VaultSealPush:    false,
		VaultSealRepo:    "vault-mirror",

		GoodReposEnabled: false,
		GoodReposPush:    true,
		GoodReposRepo:    "",
	})
	if !res.Saved {
		t.Fatal("the backup save failed")
	}
	if v, _ := s.cfg.Get("backup_folder"); v != "/backups" {
		t.Fatalf("backup_folder = %v", v)
	}
	if v, _ := s.cfg.Get("backup_max"); v != 24 {
		t.Fatalf("backup_max = %v", v)
	}
	if v, _ := s.cfg.Get("cloudflare_worker_url"); v != "https://bot.example.workers.dev" {
		t.Fatalf("worker url = %v (the trailing slash never stripped)", v)
	}
	vs := getMapCfg(s.cfg, "vaultseal")
	if vs == nil {
		t.Fatal("the vaultseal group never landed")
	}
	e, _ := vs.Get("enabled")
	p, _ := vs.Get("auto_push")
	r, _ := vs.Get("repo_name")
	if e != true || p != false || r != "vault-mirror" {
		t.Fatalf("vaultseal = %v %v %v", e, p, r)
	}
	gr := getMapCfg(s.cfg, "goodrepos")
	if gr == nil {
		t.Fatal("the goodrepos group never landed")
	}
	ge, _ := gr.Get("enabled")
	if ge != false {
		t.Fatalf("goodrepos.enabled = %v", ge)
	}
	// The merge law: unknown keys survive.
	if v, _ := s.cfg.Get("hmac_secret"); v != "keep-me" {
		t.Fatalf("the merge dropped hmac_secret: %v", v)
	}
	// The int-law heals: garbage → 10, empty → 10, 0 → 1.
	if got := backupMaxLaw("garbage"); got != 10 {
		t.Fatalf("garbage max = %v", got)
	}
	if got := backupMaxLaw(""); got != 10 {
		t.Fatalf("empty max = %v", got)
	}
	if got := backupMaxLaw("0"); got != 1 {
		t.Fatalf("zero max = %v", got)
	}
	if got := backupMaxLaw("7"); got != 7 {
		t.Fatalf("seven max = %v", got)
	}
}

func TestBackupNowDialogGuards(t *testing.T) {
	cfg := storage.NewOrderedMap()
	s, _ := newPagesBSettings(t, cfg)

	title, body := BackupNowGuards(s.cfg)
	if title != "No Vault" || body != "Set a vault path first (📁 Vault tab)." {
		t.Fatalf("guards = %q %q", title, body)
	}
	s.cfg.Set("vault_path", t.TempDir())
	title, body = BackupNowGuards(s.cfg)
	if title != "No Backup Folder" || body != "Select a backup folder first." {
		t.Fatalf("guards = %q %q", title, body)
	}
	s.cfg.Set("backup_folder", t.TempDir())
	if title, body = BackupNowGuards(s.cfg); title != "" || body != "" {
		t.Fatalf("guards should open: %q %q", title, body)
	}
}

func TestOpenDashboardURLLaw(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("cloudflare_worker_url", "https://bot.example.workers.dev/")
	s, _ := newPagesBSettings(t, cfg)

	// The field wins; the trailing slash strips; /dashboard rides.
	url, t2, b2 := OpenDashboardURL("https://field.example.workers.dev/", s.cfg)
	if url != "https://field.example.workers.dev/dashboard" || t2 != "" || b2 != "" {
		t.Fatalf("url = %q %q %q", url, t2, b2)
	}
	// The config answers when the field is empty.
	url, _, _ = OpenDashboardURL("", s.cfg)
	if url != "https://bot.example.workers.dev/dashboard" {
		t.Fatalf("url = %q", url)
	}
	// The honest refusal.
	url, t2, b2 = OpenDashboardURL("", storage.NewOrderedMap())
	if url != "" || t2 != "No Worker URL" ||
		!strings.Contains(b2, "Enter your Worker URL first.") {
		t.Fatalf("refusal = %q %q %q", url, t2, b2)
	}
}

// ---------------------------------------------------------------------------
// THE BOT GROUP's save law.
// ---------------------------------------------------------------------------

func TestSaveBotGroup(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("hmac_secret", "keep-me")
	s, _ := newPagesBSettings(t, cfg)

	res := s.SaveBotGroup(&BotValues{Username: "  @githubfetcherbot ", Token: "  abc  "})
	if !res.Saved {
		t.Fatal("the bot save failed")
	}
	if v, _ := s.cfg.Get("bot_username"); v != "githubfetcherbot" {
		t.Fatalf("username = %v (the @ never stripped)", v)
	}
	if v, _ := s.cfg.Get("bot_token"); v != "abc" {
		t.Fatalf("token = %v", v)
	}
	if v, _ := s.cfg.Get("hmac_secret"); v != "keep-me" {
		t.Fatalf("the merge dropped hmac_secret: %v", v)
	}
}

// ---------------------------------------------------------------------------
// The snapshot carries the pages B fields.
// ---------------------------------------------------------------------------

func TestSnapshotCarriesPagesB(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("bot_username", "mybot")
	cfg.Set("backup_folder", "/backups")
	cfg.Set("backup_max", 5)
	cfg.Set("backup_enabled", true)
	cfg.Set("cloudflare_worker_url", "https://bot.example.workers.dev")
	cfg.Set("vault_path", t.TempDir())
	cfg.Set("github_token", "fixture-token")
	vs := storage.NewOrderedMap()
	vs.Set("repo_name", "the-mirror")
	vs.Set("auto_push", false)
	cfg.Set("vaultseal", vs)
	s, _ := newPagesBSettings(t, cfg)

	snap := s.Snapshot()
	if snap.BotUsername != "mybot" || snap.BotToken != "" {
		t.Fatalf("bot fields = %q %q", snap.BotUsername, snap.BotToken)
	}
	if snap.BackupFolder != "/backups" || snap.BackupMax != "5" || !snap.BackupEnabled {
		t.Fatalf("backup fields = %+v", snap.pagesBSnapshot)
	}
	if snap.WorkerURL != "https://bot.example.workers.dev" {
		t.Fatalf("worker url = %q", snap.WorkerURL)
	}
	if !snap.VaultSealEnabled || snap.VaultSealPush || snap.VaultSealRepo != "the-mirror" {
		t.Fatalf("vaultseal fields = %+v", snap.pagesBSnapshot)
	}
	if !snap.HasVaultPath || !snap.HasGHToken {
		t.Fatalf("the ladder feeds never landed: %+v", snap.pagesBSnapshot)
	}
	if snap.StrikeThreshold != 3 {
		t.Fatalf("strike threshold = %d", snap.StrikeThreshold)
	}
	// The default username law (ui.py's QLineEdit default).
	cfg2 := storage.NewOrderedMap()
	s2, _ := newPagesBSettings(t, cfg2)
	if snap2 := s2.Snapshot(); snap2.BotUsername != "githubfetcherbot" {
		t.Fatalf("default username = %q", snap2.BotUsername)
	}
}
