package app

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// botdoors_test.go — T-79c: the bot-queue tool doors' line grammar +
// payload laws, pinned against the reference's own flows
// (bot_queue.py:929-1324 + dashboard.py:770-837 at 98f4fc7). Every
// network seam faked at construction; every line asserted by level +
// verbatim text; the file outputs byte-checked against the
// reference's own formats.

// wantLogAt/wantLog/noLog ride settings_test.go's helpers;
// writeManagedNote rides linkdoors_test.go's (the shared rig).

// newBotDoorRig builds the Settings over a temp APP_DIR (the cache
// lands under it) + the canned queue seam.
func newBotDoorRig(t *testing.T) (*Settings, *busRec, *fakeSettingsDeps) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	s, rec, f := newSettingsTest(t)
	f.botQueueState = "ok"
	return s, rec, f
}

// seedBotDoorCache opens the rig's cache (the temp APP_DIR's own
// cache.db).
func seedBotDoorCache(t *testing.T, s *Settings) *store.CacheDB {
	t.Helper()
	cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		t.Fatal(err)
	}
	return cache
}

// ---------------------------------------------------------------------------
// 📋 Export All Links (bot_queue.py:929).
// ---------------------------------------------------------------------------

func TestBotDoorExportGrammar(t *testing.T) {
	s, rec, f := newBotDoorRig(t)
	vault := t.TempDir()

	// the username guard (the page's own field — Option B keeps it).
	s.ExportAllBotLinks()
	wantLogAt(t, rec, "❌ Please enter the bot username first.", "error", 1)
	s.cfg.Set("bot_username", "@mybot")

	// the vault guard.
	s.ExportAllBotLinks()
	wantLogAt(t, rec, "❌ No vault selected — need a place to save the export.", "error", 2)
	s.cfg.Set("vault_path", vault)

	// the fetch error states.
	s.SpeakOnBus("—", "info") // the separator so the indexes stay honest
	f.botQueueState = "error"
	title, body := s.ExportAllBotLinks()
	wantLogAt(t, rec, "📋 Exporting ALL links from bot (no limit)...", "info", 4)
	wantLogAt(t, rec, "❌ Export failed: the Worker queue poll failed "+
		"(network down or the Worker answered an error)", "error", 5)
	if title != "" || body != "" {
		t.Fatalf("a failed export is log-only: %q / %q", title, body)
	}
	s.SpeakOnBus("—", "info")
	f.botQueueState = "disabled"
	s.ExportAllBotLinks()
	wantLogAt(t, rec, "❌ Export failed: the Worker queue is not configured "+
		"(no worker_base_url or pairing)", "error", 7)

	// the happy path: the routed lists + the Σ forward_count message
	// count + the export file + the two verdict lines + the box.
	s.SpeakOnBus("—", "info")
	f.botQueueState = "ok"
	f.botQueueItems = []any{
		map[string]any{"url_normalized": "https://github.com/acme/one",
			"forward_count": float64(2)},
		map[string]any{"url_normalized": "https://github.com/acme/two"},
		map[string]any{"url_normalized": "https://example.com/page",
			"forward_count": float64(3)},
	}
	title, body = s.ExportAllBotLinks()
	wantLogAt(t, rec, "📋 Exporting ALL links from bot (no limit)...", "info", 9)
	if title != "📋 Links Exported" {
		t.Fatalf("the export box title: %q", title)
	}
	matches, err := filepath.Glob(filepath.Join(vault, "_bot_links_export_*.md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("the export file: %v (%d)", err, len(matches))
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	stampRE := regexp.MustCompile(`\d{4}-\d{2}-\d{2} \d{2}:\d{2}(:\d{2})?`)
	content := string(stampRE.ReplaceAll(raw, []byte("<TS>")))
	want := "# 📋 Bot Links Export — <TS>\n" +
		"\n" +
		"## 📊 Statistics\n" +
		"- 📬 Total messages in bot: **5**\n" +
		"- 🔗 Total raw links found: **3**\n" +
		"- 🔄 Duplicates removed: **0**\n" +
		"- ✅ Unique GitHub repos: **2**\n" +
		"- ✅ Unique non-GitHub links: **1**\n" +
		"- 📝 Total unique links: **3**\n" +
		"\n" +
		"## 🐙 GitHub Repos\n" +
		"\n" +
		"1. https://github.com/acme/one\n" +
		"2. https://github.com/acme/two\n" +
		"\n" +
		"## 🔗 Non-GitHub Links\n" +
		"\n" +
		"1. https://example.com/page\n" +
		"\n" +
		"---\n" +
		"*Export generated: <TS>*\n" +
		"*Compare this list with what you forwarded to the bot to verify no links are missing.*"
	if content != want {
		t.Fatalf("the export file:\n%s\nwant:\n%s", content, want)
	}
	wantLogAt(t, rec, "📋 Export saved: "+matches[0], "success", 10)
	wantLogAt(t, rec, "📊 Total messages: 5 | GitHub: 2 | Non-GitHub: 1 | Duplicates: 0", "info", 11)
	wantBody := "Total messages: 5\n" +
		"GitHub repos: 2\n" +
		"Non-GitHub: 1\n" +
		"Duplicates removed: 0\n\n" +
		"Export saved to:\n" + filepath.Base(matches[0]) + "\n\n" +
		"Open this file in Obsidian to verify all links are accounted for."
	if body != wantBody {
		t.Fatalf("the export box body:\n%q\nwant:\n%q", body, wantBody)
	}

	// the write failure (a nonexistent vault dir — the reference's
	// guard only checks the combo's text; the write speaks).
	s.SpeakOnBus("—", "info")
	s.cfg.Set("vault_path", filepath.Join(t.TempDir(), "missing"))
	f.botQueueItems = nil
	s.ExportAllBotLinks()
	if l, ok := rec.findLog("❌ Failed to save export:"); !ok {
		t.Fatalf("the save-failure line never spoke (have %d lines)", len(rec.logs))
	} else if l.Level != "error" {
		t.Fatalf("the save-failure level: %q", l.Level)
	}
}

// TestBotDoorExportOrderedMapItems pins the seam's other row shape
// (the workerclient's OrderedMap rows) + the message-count law over
// it.
func TestBotDoorExportOrderedMapItems(t *testing.T) {
	s, rec, f := newBotDoorRig(t)
	s.cfg.Set("bot_username", "mybot")
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	row := newOrderedItem(t, map[string]any{
		"url_normalized": "https://github.com/acme/x",
		"forward_count":  float64(4),
	})
	f.botQueueItems = []any{row}
	title, _ := s.ExportAllBotLinks()
	if title != "📋 Links Exported" {
		t.Fatalf("the OrderedMap row exported: %q", title)
	}
	wantLogAt(t, rec, "📊 Total messages: 4 | GitHub: 1 | Non-GitHub: 0 | Duplicates: 0", "info", 2)
}

// ---------------------------------------------------------------------------
// ✅ Verify All Processed (bot_queue.py:1026).
// ---------------------------------------------------------------------------

func TestBotDoorVerifyGuardsAndFetchStates(t *testing.T) {
	s, rec, f := newBotDoorRig(t)

	// the username guard.
	res := s.VerifyAllBotLinks()
	wantLogAt(t, rec, "❌ Please enter the bot username first.", "error", 1)
	if res.Title != "" || res.Body != "" || res.Missing != nil {
		t.Fatalf("a refusal is log-only: %+v", res)
	}
	s.cfg.Set("bot_username", "mybot")

	// the vault guard — the empty path.
	res = s.VerifyAllBotLinks()
	wantLogAt(t, rec, "❌ No vault selected — cannot verify links.", "error", 2)
	if res.Title != "" {
		t.Fatalf("a refusal is log-only: %+v", res)
	}

	// the vault guard — a missing dir (the reference's isdir check).
	s.cfg.Set("vault_path", filepath.Join(t.TempDir(), "missing"))
	res = s.VerifyAllBotLinks()
	wantLogAt(t, rec, "❌ No vault selected — cannot verify links.", "error", 3)

	// the fetch error — the line + the display's Error text rides
	// the box (the recorded divergence).
	s.cfg.Set("vault_path", t.TempDir())
	f.botQueueState = "error"
	res = s.VerifyAllBotLinks()
	wantLogAt(t, rec, "✅ Verifying all bot links against vault...", "info", 4)
	wantLogAt(t, rec, "❌ Verify All failed: the Worker queue poll failed "+
		"(network down or the Worker answered an error)", "error", 5)
	if res.Title != "Verify All" || res.Body != "Error: the Worker queue poll failed "+
		"(network down or the Worker answered an error)" {
		t.Fatalf("the failure box: %+v", res)
	}
}

func TestBotDoorVerifyLadder(t *testing.T) {
	s, rec, f := newBotDoorRig(t)
	s.cfg.Set("bot_username", "mybot")
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManagedNote(t, vault, "invault", "the body")

	// The cache fixture: a dead row, a processed row (the renamed
	// case's bucket), and the settlement PRIMED (the meta guard) —
	// so the door's own settlement call is a no-op and the fetched
	// urls read by their own bucket truth.
	cache := seedBotDoorCache(t, s)
	cache.Decommission("https://github.com/acme/dead", "HTTP 404")
	_ = cache.AddProcessed(0, "https://github.com/acme/cached",
		"acme", "cached", "note.md", "Tools")
	_, _ = cache.SettleRepos([]string{"https://github.com/acme/settled-one"})
	if _, err := cache.SettleExistingRepos(nil); err != nil {
		t.Fatal(err)
	}
	_ = cache.Close()

	f.botQueueItems = []any{
		map[string]any{"url_normalized": "https://github.com/acme/invault"},
		map[string]any{"url_normalized": "https://github.com/acme/cached"},
		map[string]any{"url_normalized": "https://github.com/Acme/InVault.git"},
		map[string]any{"url_normalized": "https://github.com/acme/dead"},
		map[string]any{"url_normalized": "https://github.com/acme/settled-one"},
		map[string]any{"url_normalized": "https://github.com/acme/missing-one"},
		map[string]any{"url_normalized": "https://example.com/extra",
			"forward_count": float64(2)},
	}

	res := s.VerifyAllBotLinks()
	wantLogAt(t, rec, "✅ Verifying all bot links against vault...", "info", 1)
	wantLogAt(t, rec, "📚 Vault index: 1 notes indexed", "info", 2)
	// the primed settlement: the door's call reads already — NO line.
	noLog(t, rec, "THE SETTLEMENT (repos)")
	wantLogAt(t, rec, "✅ Found in cache (processed before): "+
		"https://github.com/acme/cached", "info", 3)
	// v29.9 — the fuzzy match: the same owner/repo pair under the
	// vault note's source URL (the path-case + .git drift the URL
	// normalizer does not fold).
	wantLogAt(t, rec, "✅ Fuzzy match: https://github.com/Acme/InVault.git → invault.md", "info", 4)
	wantLogAt(t, rec, "⏳ 1 missing GitHub link(s):", "warning", 5)
	wantLogAt(t, rec, "   1. https://github.com/acme/missing-one", "info", 6)

	if res.Title != "✅ VERIFICATION REPORT" {
		t.Fatalf("the report box title: %q", res.Title)
	}
	for _, want := range []string{
		"✅ VERIFICATION REPORT",
		"============================================================",
		"Total GitHub links in bot:      6",
		"✅ Found in vault:              3",
		"🗑️ Decommissioned (404):        1",
		"🤝 Settled (addressed):         1  (never re-fetched)",
		"⏳ Missing from vault:          1",
		"🔗 Non-GitHub links:            1",
		"ALL GITHUB LINKS — DETAILED CHECK:",
		"  1. ✅ https://github.com/acme/invault",
		"       → invault.md",
		"  2. ⏳ https://github.com/acme/cached",
		"  3. ⏳ https://github.com/Acme/InVault.git",
		"  4. ⏳ https://github.com/acme/dead",
		"  5. ⏳ https://github.com/acme/settled-one",
		"  6. ⏳ https://github.com/acme/missing-one",
		"SUMMARY: 1 link(s) need processing!",
		"To process the missing links:",
		"  1. Click '📬 Check Queue' to load them",
		"  2. Click '🚀 Process All' to process",
	} {
		if !strings.Contains(res.Body, want) {
			t.Fatalf("the report missing %q:\n%s", want, res.Body)
		}
	}
	// the resolve payload.
	if len(res.Missing) != 1 || res.Missing[0] != "https://github.com/acme/missing-one" {
		t.Fatalf("the missing payload: %v", res.Missing)
	}

	// the summary line (the missing path).
	wantLogAt(t, rec, "✅ Verified: 3/6 GitHub links in vault (1 missing)", "warning", 7)

	// the permanent report file: the summary + the detailed tables.
	matches, err := filepath.Glob(filepath.Join(vault, "_verification_report_*.md"))
	if err != nil || len(matches) != 1 {
		t.Fatalf("the report file: %v (%d)", err, len(matches))
	}
	raw, err := os.ReadFile(matches[0])
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"# ✅ Verification Report — ",
		"## 📊 Summary",
		"| Metric | Count |\n|--------|-------|",
		"| 📬 Total GitHub links in bot | 6 |",
		"| ✅ Found in vault | 3 |",
		"| ⏳ Missing from vault | 1 |",
		"| 🔗 Non-GitHub links | 1 |",
		"## 📋 Detailed Check",
		"| # | Status | URL | Note |",
		"|---|--------|-----|------|",
		"| 1 | ✅ | https://github.com/acme/invault | invault.md |",
		"| 2 | ⏳ | https://github.com/acme/cached | — |",
		"| 6 | ⏳ | https://github.com/acme/missing-one | — |",
		"*Report generated: ",
	} {
		if !strings.Contains(string(raw), want) {
			t.Fatalf("the report file missing %q:\n%s", want, raw)
		}
	}
	wantLogAt(t, rec, "📄 Verification report saved to vault: "+
		filepath.Base(matches[0]), "success", 8)
}

// TestBotDoorVerifyFreshCacheSettlement pins the fresh-cache path:
// the door's own settlement speaks THE SETTLEMENT line, every fetched
// repo rides the settled bucket (the owner's "older ones are
// processed" law), and the all-accounted summary lands (the
// mark-all-read question is Option B's accepted loss — nothing asks).
func TestBotDoorVerifyFreshCacheSettlement(t *testing.T) {
	s, rec, f := newBotDoorRig(t)
	s.cfg.Set("bot_username", "mybot")
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	f.botQueueItems = []any{
		map[string]any{"url_normalized": "https://github.com/acme/new-one"},
	}
	res := s.VerifyAllBotLinks()
	wantLogAt(t, rec, "🤝 THE SETTLEMENT (repos): 1 repo link(s) the "+
		"bot already delivered are settled — addressed and processed, "+
		"never fetched or counted again; 0 queued retry(ies) cleared. "+
		"Only repos added from now on are counted (♻️ Reset 404 "+
		"Quarantine un-settles any of them)", "info", 2)
	wantLogAt(t, rec, "✅ Verified: 0 in vault + 0 decommissioned + 1 "+
		"settled = 1/1 accounted for", "success", 3)
	if len(res.Missing) != 0 {
		t.Fatalf("the fresh-cache verify misses nothing: %v", res.Missing)
	}
	if !strings.Contains(res.Body, "🎉 All GitHub links are in the vault!") {
		t.Fatalf("the all-accounted report: %s", res.Body)
	}
	// the mark-all-read question is dropped (read-acknowledge is the
	// accepted loss — nothing in the payload asks).
	noLog(t, rec, "mark all bot messages as read")
}

// ---------------------------------------------------------------------------
// 🔄 Retry Failed (dashboard.py:770 — the scan half).
// ---------------------------------------------------------------------------

func TestBotDoorRetryScan(t *testing.T) {
	s, rec, _ := newBotDoorRig(t)

	// the vault guard box.
	s.cfg.Set("vault_path", "")
	out := s.RetryFailedScan()
	if out.Box == nil || out.Box.Title != "Error" ||
		out.Box.Body != "Please select a valid Obsidian vault path." {
		t.Fatalf("the vault-error box: %+v", out.Box)
	}
	if len(rec.logs) != 0 {
		t.Fatalf("the vault guard is box-only: %v", rec.logs)
	}
	s.cfg.Set("vault_path", t.TempDir())

	// the empty queue.
	out = s.RetryFailedScan()
	wantLogAt(t, rec, "✓ No failed repos to retry.", "success", 1)
	if out.Box == nil || out.Box.Title != "Retry Queue Empty" ||
		out.Box.Body != "No failed repos to retry. 🎉" {
		t.Fatalf("the empty box: %+v", out.Box)
	}
	if out.Confirm != nil || len(out.URLs) != 0 {
		t.Fatalf("the empty scan carries nothing else: %+v", out)
	}

	// the read failure (a corrupt cache.db).
	if err := os.WriteFile(cacheDBPathOf(s.cfg), []byte("not a sqlite file"), 0o644); err != nil {
		t.Fatal(err)
	}
	out = s.RetryFailedScan()
	if l, ok := rec.findLog("❌ Failed to read retry queue:"); !ok {
		t.Fatalf("the read-failure line never spoke")
	} else if l.Level != "error" {
		t.Fatalf("the read-failure level: %q", l.Level)
	}
	if out.Box != nil || out.Confirm != nil || len(out.URLs) != 0 {
		t.Fatalf("a failed read carries nothing: %+v", out)
	}

	// a fresh cache: the settled skip + the kept rows.
	os.Remove(cacheDBPathOf(s.cfg))
	cache := seedBotDoorCache(t, s)
	cache.AddFailed("https://github.com/acme/stale", "boom")
	cache.AddFailed("https://github.com/acme/live-one", "boom")
	cache.AddFailed("https://github.com/acme/live-two", "boom")
	_, _ = cache.SettleRepos([]string{"https://github.com/acme/stale"})
	_ = cache.Close()
	out = s.RetryFailedScan()
	wantLogAt(t, rec, "🤝 1 settled repo link(s) left the retry queue — "+
		"addressed (the owner's law), never re-fetched (♻️ Reset 404 "+
		"Quarantine un-settles any of them).", "info", 3)
	// the kept set (the rows' order is the cache's own — failed_at
	// DESC with equal stamps rides the rowid; the DOOR's contract is
	// the set, not the order).
	kept := map[string]bool{}
	for _, u := range out.URLs {
		kept[u] = true
	}
	if len(out.URLs) != 2 || !kept["https://github.com/acme/live-one"] ||
		!kept["https://github.com/acme/live-two"] {
		t.Fatalf("the kept urls: %v", out.URLs)
	}
	if out.Box != nil || out.Confirm != nil {
		t.Fatalf("a ready scan asks nothing: %+v", out)
	}
}

// TestBotDoorRetryScanConfirmGate pins the >10 confirm (v31.1's safety
// gate over "the retry queue") + the boundary (exactly 10 runs).
func TestBotDoorRetryScanConfirmGate(t *testing.T) {
	s, _, _ := newBotDoorRig(t)
	s.cfg.Set("vault_path", t.TempDir())
	cache := seedBotDoorCache(t, s)
	for i := 0; i < 11; i++ {
		cache.AddFailed("https://github.com/acme/r"+string(rune('a'+i)), "boom")
	}
	_ = cache.Close()
	out := s.RetryFailedScan()
	if out.Confirm == nil || out.Confirm.Title != "Confirm Large Batch" ||
		out.Confirm.Body != "This will process 11 items (the retry queue).\n\nContinue?" ||
		!out.Confirm.Question {
		t.Fatalf("the confirm box: %+v", out.Confirm)
	}
	if len(out.URLs) != 11 {
		t.Fatalf("the scanned count: %d", len(out.URLs))
	}

	// exactly 10: no gate.
	os.Remove(cacheDBPathOf(s.cfg))
	cache = seedBotDoorCache(t, s)
	for i := 0; i < 10; i++ {
		cache.AddFailed("https://github.com/acme/r"+string(rune('a'+i)), "boom")
	}
	_ = cache.Close()
	out = s.RetryFailedScan()
	if out.Confirm != nil {
		t.Fatalf("10 items never ask: %+v", out.Confirm)
	}
	if len(out.URLs) != 10 {
		t.Fatalf("the scanned count: %d", len(out.URLs))
	}
}

// ---------------------------------------------------------------------------
// The manual-resolve engines (bot_queue.py:1368-1496).
// ---------------------------------------------------------------------------

func TestBotDoorResolveEngines(t *testing.T) {
	s, rec, _ := newBotDoorRig(t)
	urlA := "https://github.com/acme/alpha"
	urlB := "https://github.com/acme/beta"

	// Mark as Processed: the row + the settle + the line; the
	// all-resolved arithmetic over the dialog's list.
	res := s.BotResolveMarkProcessed(urlA, []string{urlA, urlB})
	wantLogAt(t, rec, "✅ Marked as processed (manual): "+urlA, "success", 1)
	if res.AllResolved {
		t.Fatal("one row left — the dialog stays")
	}
	// the dialog's list SHRINKS as rows resolve (the frontend's own
	// state — the reference's url_list.takeItem); the second call
	// passes the current list.
	res = s.BotResolveMarkProcessed(urlB, []string{urlB})
	wantLogAt(t, rec, "✅ Marked as processed (manual): "+urlB, "success", 2)
	wantLogAt(t, rec, "✅ All missing links resolved!", "success", 3)
	if !res.AllResolved {
		t.Fatal("the list emptied — the dialog closes")
	}
	cache := seedBotDoorCache(t, s)
	processed := map[string]bool{}
	for _, u := range cache.GetAllProcessedURLs() {
		processed[u] = true
	}
	_ = cache.Close()
	// THE CROSS-BINARY DISAGREEMENT LAW: processed_repos keys on
	// repo_id (the reference's own schema, cache_db.py:151) — two
	// manual marks at repo_id 0 REPLACE: only the LAST row lands,
	// exactly like the reference's own dialog. The SETTLEMENT is
	// the durable record (both urls settle, keyed by URL).
	if !processed[urlB] {
		t.Fatalf("the last processed row: %v", processed)
	}
	if processed[urlA] {
		t.Fatalf("repo_id 0 replaces — the reference's own law: %v", processed)
	}
	cache = seedBotDoorCache(t, s)
	if !cache.IsRepoSettled(urlA) || !cache.IsRepoSettled(urlB) {
		t.Fatal("the owner's verdict settles the ledger too (v0.63.3)")
	}
	_ = cache.Close()

	// Decommission: the quarantine row + the settle + the line (info).
	s.SpeakOnBus("—", "info")
	urlC := "https://github.com/acme/gamma"
	res = s.BotResolveDecommission(urlC, []string{urlC})
	wantLogAt(t, rec, "🗑️ Decommissioned (manual): "+urlC, "info", 5)
	if !res.AllResolved {
		t.Fatal("the single row resolved")
	}
	cache = seedBotDoorCache(t, s)
	if !cache.IsDecommissioned(urlC) || !cache.IsRepoSettled(urlC) {
		t.Fatal("the decommission row + the settle")
	}
	_ = cache.Close()

	// Mark ALL: the loop + the bulk settle + the count line.
	s.SpeakOnBus("—", "info")
	res = s.BotResolveMarkAll([]string{urlA, urlB})
	wantLogAt(t, rec, "✅ Marked 2 URL(s) as processed (manual)", "success", 7)
	if !res.AllResolved {
		t.Fatal("mark-all resolves everything")
	}

	// the pair extraction: a non-GitHub url seeds empty owner/repo
	// (the reference's m-or-"" law) — the row still lands.
	s.BotResolveMarkProcessed("https://example.com/plain", []string{"https://example.com/plain"})
	wantLogAt(t, rec, "✅ Marked as processed (manual): https://example.com/plain", "success", 8)
}

// ---------------------------------------------------------------------------
// THE GUARD LAW's own pin: export/verify take NO busy guard (the
// reference guards only the TELEGRAM lock — Option B's accepted
// loss); the retry scan reads freely too (the batch refuses at its
// own door, never the scan).
// ---------------------------------------------------------------------------

func TestBotDoorNoBusyGuard(t *testing.T) {
	s, rec, f := newBotDoorRig(t)
	s.cfg.Set("bot_username", "mybot")
	s.cfg.Set("vault_path", t.TempDir())
	f.botQueueState = "ok"
	if !s.lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer s.lock.Release("batch")

	s.ExportAllBotLinks()
	res := s.VerifyAllBotLinks()
	s.RetryFailedScan()
	noLog(t, rec, "A batch is running")
	if res.Title != "✅ VERIFICATION REPORT" {
		t.Fatalf("the verify ran under the lock: %+v", res)
	}
	wantLog(t, rec, "📋 Exporting ALL links from bot (no limit)...", "info")
	wantLog(t, rec, "✅ Verifying all bot links against vault...", "info")
	wantLog(t, rec, "✓ No failed repos to retry.", "success")
}

// ---------------------------------------------------------------------------
// THE FIELD-TAG LAW (the payload structs' tags mirror the keys the
// frontend reads).
// ---------------------------------------------------------------------------

func TestBotDoorPayloadFieldTags(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"BotVerifyResult", BotVerifyResult{Title: "t", Body: "b",
			Missing: []string{"u"}},
			`{"title":"t","body":"b","missing":["u"]}`},
		{"BotVerifyResultEmpty", BotVerifyResult{},
			`{"title":"","body":"","missing":null}`},
		{"BotRetryScan", BotRetryScan{
			Box:     &HeroBox{Title: "E", Body: "B"},
			Confirm: &HeroBox{Title: "C", Body: "Q", Question: true},
			URLs:    []string{"u"}},
			`{"box":{"title":"E","body":"B"},` +
				`"confirm":{"title":"C","body":"Q","question":true},` +
				`"urls":["u"]}`},
		{"BotResolveAction", BotResolveAction{AllResolved: true},
			`{"allResolved":true}`},
	}
	for _, tc := range cases {
		raw, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatalf("%s marshal: %v", tc.name, err)
		}
		if string(raw) != tc.want {
			t.Fatalf("%s = %s, want %s", tc.name, raw, tc.want)
		}
	}
}

// newOrderedItem builds one workerclient-shaped row (the OrderedMap
// form the honest poll's callers also produce).
func newOrderedItem(t *testing.T, fields map[string]any) any {
	t.Helper()
	out := storage.NewOrderedMap()
	for k, v := range fields {
		out.Set(k, v)
	}
	return out
}
