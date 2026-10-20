package app

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/mirror"
	"github.com/assadigit/second-brain-project/internal/store"
)

// linkdoors_test.go — T-79b: the linking-tools doors' line grammar +
// the SAFE-DEFAULT laws, pinned against the reference's own flows
// (linking_tools.py:59-124, dashboard.py:302-660,
// connection_tests.py:219-233 at 98f4fc7). Every seam faked at
// construction; every line asserted by level + verbatim text.

// wantLogAt/wantLog ride settings_test.go's helpers (the shared rig).

// ---------------------------------------------------------------------------
// fixtures
// ---------------------------------------------------------------------------

// writeManagedNote drops one managed GitHub-vault note (a source:
// line + a # heading — the fields' name source).
func writeManagedNote(t *testing.T, vault, name, body string) string {
	t.Helper()
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(vault, name+".md")
	content := "---\nsource: https://github.com/acme/" + name + "\n" +
		"category: Agents\n---\n\n# " + name + "\n\n" + body + "\n"
	if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

// writeManifest drops the vault's links_manifest.json.
func writeManifest(t *testing.T, vault string, links []map[string]any) {
	t.Helper()
	doc := map[string]any{
		"batch_id":    "20261020_120000",
		"created_at":  "2026-10-20T12:00:00",
		"source":      "bot",
		"total_links": len(links),
		"links":       links,
	}
	raw, err := json.Marshal(doc)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vault, "links_manifest.json"), raw, 0o644); err != nil {
		t.Fatal(err)
	}
}

// fakeReplyJSON builds the tools' LLM JSON reply (best_used_for /
// related+reason shapes).
func fakeReplyJSON(fields string) string {
	return `{"` + fields + `"}`
}

// ---------------------------------------------------------------------------
// ✅ Validate Vault (connection_tests.py:219).
// ---------------------------------------------------------------------------

func TestLinkDoorValidateVaultGrammar(t *testing.T) {
	s, rec, _ := newSettingsTest(t)

	// no vault selected.
	s.ValidateVault()
	wantLogAt(t, rec, "📁 Validating vault path...", "info", 1)
	wantLogAt(t, rec, "❌ No vault selected.", "error", 2)

	// missing path.
	vault := filepath.Join(t.TempDir(), "nope")
	writeCfgFile(t, s.cfg, `{"vault_path": "`+strings.ReplaceAll(vault, `\`, `\\`)+`"}`)
	s.ValidateVault()
	wantLogAt(t, rec, "❌ Vault path does not exist: "+vault, "error", 3)

	// valid Obsidian vault.
	ok := t.TempDir()
	if err := os.MkdirAll(filepath.Join(ok, ".obsidian"), 0o755); err != nil {
		t.Fatal(err)
	}
	s.cfg.Set("vault_path", ok)
	s.ValidateVault()
	wantLogAt(t, rec, "✅ Vault is valid (contains .obsidian folder): "+ok, "success", 4)

	// a plain folder: the warning.
	plain := t.TempDir()
	s.cfg.Set("vault_path", plain)
	s.ValidateVault()
	wantLogAt(t, rec, "⚠️ Path exists but does not appear to be an Obsidian vault "+
		"(no .obsidian folder).", "warning", 5)
}

func TestLinkDoorValidateVaultBusyGuard(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	s.cfg.Set("vault_path", t.TempDir())
	if !s.lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer s.lock.Release("batch")
	s.ValidateVault()
	wantLogAt(t, rec, linkingBusyLine, "warning", 1)
	if _, ok := rec.findLog("📁 Validating vault path..."); ok {
		t.Fatal("the guarded door never runs")
	}
}

// ---------------------------------------------------------------------------
// 🔍 Verify Vault (dashboard.py:302-498).
// ---------------------------------------------------------------------------

func TestLinkDoorVerifyVaultNoVault(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	title, body := s.VerifyVault()
	if title != "Verify Vault" || body != "No vault selected." {
		t.Fatalf("box = %q / %q", title, body)
	}
	if l, ok := rec.findLog("🔍 Verify:"); ok {
		t.Fatalf("no verify ran, log %q", l.Msg)
	}
}

func TestLinkDoorVerifyVaultNoManifest(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	title, body := s.VerifyVault()
	wantLogAt(t, rec, "ℹ️ No manifest to verify — process a batch first.", "info", 1)
	if title != "Verify Vault" {
		t.Fatalf("title = %q", title)
	}
	want := "No previous manifest found in this vault.\n" +
		"Process a batch first — the manifest is created at the start of every batch."
	if body != want {
		t.Fatalf("body = %q", body)
	}
}

func TestLinkDoorVerifyVaultReportAllVerified(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	note := writeManagedNote(t, vault, "alpha",
		"Long enough body to pass the >100 byte note-size probe. Padding padding padding.")
	writeManifest(t, vault, []map[string]any{
		{"url": "https://github.com/acme/alpha", "type": "github",
			"status": "processed", "note_path": note},
	})
	title, body := s.VerifyVault()
	wantLogAt(t, rec, "🔍 Verify: 1 processed, 0 failed", "info", 1)
	if title != "🔍 VAULT VERIFICATION REPORT" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{
		"Batch:    20261020_120000",
		"Source:   bot",
		"Created:  2026-10-20T12:00:00",
		"Total links: 1",
		"✅ GitHub processed:    1",
		"⏭️ GitHub skipped:      0  (dedup — already in vault)",
		"❌ GitHub failed:       0",
		"🧮 Accounting:          1/1 accounted for ✓",
		"🎉 ALL LINKS VERIFIED — NO DATA LOSS!",
		strings.Repeat("=", 60),
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("report missing %q\nbody:\n%s", want, body)
		}
	}
	if strings.Contains(body, "⚠️") {
		t.Fatalf("no failed links in the all-clear report:\n%s", body)
	}
	if l, ok := rec.findLog("📥 Added"); ok {
		t.Fatalf("no enqueue on the clean report: %q", l.Msg)
	}
}

func TestLinkDoorVerifyVaultFailedGitHubEnqueues(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManifest(t, vault, []map[string]any{
		{"url": "https://github.com/acme/gone", "type": "github",
			"status": "failed", "error": "fetch exploded"},
	})
	title, body := s.VerifyVault()
	wantLogAt(t, rec, "🔍 Verify: 0 processed, 1 failed", "info", 1)
	if title != "🔍 VAULT VERIFICATION REPORT" {
		t.Fatalf("title = %q", title)
	}
	for _, want := range []string{
		"❌ GitHub failed:       1",
		"⚠️ 1 link(s) need retry:",
		"  ❌ https://github.com/acme/gone",
		"     Type:   github",
		"     Status: failed",
		"     Error:  fetch exploded",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("report missing %q\nbody:\n%s", want, body)
		}
	}
	wantLogAt(t, rec, "📥 Added 1 failed link(s) to the retry queue. Click "+
		"'🔄 Retry Failed' in the Bot tab to reprocess them.", "info", 2)
	// the retry queue row landed in the (hermetic) cache.
	cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		t.Fatal(err)
	}
	defer cache.Close()
	if got := cache.GetFailedCount(); got != 1 {
		t.Fatalf("retry queue rows = %d", got)
	}
}

func TestLinkDoorVerifyVaultWebsitesFailedLine(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManifest(t, vault, []map[string]any{
		{"url": "https://example.com/one", "type": "non-github",
			"status": "failed", "error": "timeout"},
		{"url": "https://example.com/two", "type": "non-github",
			"status": "failed", "error": "timeout"},
	})
	_, body := s.VerifyVault()
	wantLogAt(t, rec, "ℹ️ 2 website links failed their fetches — the Websites "+
		"pipeline's own retry queue and master table own them (never the repos "+
		"retry queue).", "info", 1)
	if !strings.Contains(body, "❌ Non-GitHub failed:   2") {
		t.Fatalf("the non-github tally: %s", body)
	}
	if l, ok := rec.findLog("📥 Added"); ok {
		t.Fatalf("website failures never enter the repos queue: %q", l.Msg)
	}
}

func TestLinkDoorVerifyVaultSettledSkipped(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManifest(t, vault, []map[string]any{
		{"url": "https://github.com/acme/settled", "type": "github",
			"status": "failed", "error": "old failure"},
	})
	// seed the settled ledger (the owner's law: addressed, never re-fetched).
	cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := cache.SettleRepos([]string{"https://github.com/acme/settled"}); err != nil {
		t.Fatal(err)
	}
	cache.Close()
	s.VerifyVault()
	wantLogAt(t, rec, "🤝 1 settled repo link(s) kept out of the retry queue — "+
		"addressed (the owner's law), never re-fetched.", "info", 1)
	if l, ok := rec.findLog("📥 Added"); ok {
		t.Fatalf("settled links never re-enter the queue: %q", l.Msg)
	}
}

func TestLinkDoorVerifyVaultBusyGuard(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	s.cfg.Set("vault_path", t.TempDir())
	if !s.lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer s.lock.Release("batch")
	title, _ := s.VerifyVault()
	wantLogAt(t, rec, linkingBusyLine, "warning", 1)
	if title != "" {
		t.Fatalf("the guarded door returns no box, got %q", title)
	}
}

// ---------------------------------------------------------------------------
// THE TESTWORKER LAW (the runner grammar).
// ---------------------------------------------------------------------------

func TestLinkDoorRecallHooksVaultMissing(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	s.cfg.Set("vault_path", filepath.Join(appDir, "missing"))
	s.RecallHooksDryRun()
	wantLogAt(t, rec, "⏳ Recall hooks (dry-run) running… (the GUI stays usable)", "info", 1)
	wantLogAt(t, rec, "GitHub vault not set or not found (config vault_path or "+
		"--vault).", "info", 2)
	wantLogAt(t, rec, "❌ Recall hooks (dry-run) finished with exit code 1 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 3)
}

func TestLinkDoorRecallHooksNothingToDo(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	f.linkOllamaChat = func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (string, string) {
		t.Errorf("the LLM is never called when there is nothing to do")
		return "", ""
	}
	s.RecallHooksDryRun()
	wantLogAt(t, rec, "GitHub vault: "+vault, "info", 1)
	wantLogAt(t, rec, "Notes needing a hook: 0", "info", 2)
	wantLogAt(t, rec, "Nothing to do.", "info", 3)
	wantLogAt(t, rec, "✅ Recall hooks (dry-run) finished.", "success", 4)
}

func TestLinkDoorRecallHooksDryRunGrammarAndLaw(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	note := writeManagedNote(t, vault, "alpha",
		"A CLI tool for doing very specific things with care and attention to detail.")
	before, err := os.ReadFile(note)
	if err != nil {
		t.Fatal(err)
	}
	f.linkOllamaChat = func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (string, string) {
		if host != "http://127.0.0.1:11434" || model != "llama3" {
			t.Errorf("chat seam = %s / %s", host, model)
		}
		if numCtx != llmclient.DefaultNumCtx {
			t.Errorf("numCtx = %d, want the default %d", numCtx, llmclient.DefaultNumCtx)
		}
		if len(messages) != 1 {
			t.Errorf("messages = %d", len(messages))
		}
		return fakeReplyJSON(`best_used_for": "use when you need to file things carefully`), ""
	}
	s.RecallHooksDryRun()
	wantLogAt(t, rec, "⏳ Recall hooks (dry-run) running… (the GUI stays usable)", "info", 1)
	wantLogAt(t, rec, "GitHub vault: "+vault, "info", 2)
	wantLogAt(t, rec, "Notes needing a hook: 1", "info", 3)
	wantLogAt(t, rec, "  LLM: llama3 via Ollama (http://127.0.0.1:11434) "+
		"(num_ctx=8192)", "info", 4)
	wantLogAt(t, rec, "Mode: DRY RUN (nothing written)", "info", 5)
	l := wantLogAt(t, rec, "  ~ alpha.md — would write: use when you need to "+
		"file things carefully", "info", 6)
	_ = l
	wantLogAt(t, rec, "Report: "+filepath.Join(appDir, "reports", "recall",
		"recall-dryrun.md"), "info", 7)
	wantLogAt(t, rec, "Dry run — re-run with --apply to write 1 hook(s).", "info", 8)
	wantLogAt(t, rec, "✅ Recall hooks (dry-run) finished.", "success", 9)

	// THE DRY-RUN LAW: the note file is byte-unchanged; the report
	// carries the planned diff.
	after, err := os.ReadFile(note)
	if err != nil {
		t.Fatal(err)
	}
	if string(before) != string(after) {
		t.Fatalf("THE DRY-RUN LAW broken — the note changed")
	}
	report, err := os.ReadFile(filepath.Join(appDir, "reports", "recall", "recall-dryrun.md"))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(report), "PLANNED  alpha.md") {
		t.Fatalf("the report's planned row: %s", report)
	}
	if !strings.Contains(string(report), "GitCurator — recall hooks — DRY RUN (nothing written)") {
		t.Fatalf("the report's mode line: %s", report)
	}
}

func TestLinkDoorRecallHooksProviderAbort(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManagedNote(t, vault, "alpha", "body text")
	// the cloud provider without its endpoint — the abort path.
	s.cfg.Set("llm_provider", "cloud")
	s.RecallHooksDryRun()
	wantLogAt(t, rec, "cloud provider needs config.json cloud_api_url + "+
		"cloud_model", "info", 4)
	wantLogAt(t, rec, "❌ Recall hooks (dry-run) finished with exit code 2 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 5)
}

func TestLinkDoorRecallHooksCrashLine(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManagedNote(t, vault, "alpha", "body text")
	f.linkOllamaChat = func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (string, string) {
		panic("the model exploded")
	}
	s.RecallHooksDryRun()
	wantLogAt(t, rec, "💥 the model exploded", "info", 6)
	wantLogAt(t, rec, "❌ Recall hooks (dry-run) finished with exit code 2 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 7)
}

func TestLinkDoorRecallHooksBusyGuard(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	if !s.lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer s.lock.Release("batch")
	s.RecallHooksDryRun()
	wantLogAt(t, rec, linkingBusyLine, "warning", 1)
	if _, ok := rec.findLog("⏳ Recall hooks (dry-run) running…"); ok {
		t.Fatal("the guarded tool never starts")
	}
}

// TestLinkDoorTailFortyLaw pins the last-40-lines law end-to-end: a
// 45-note vault makes 50+ captured lines; only the LAST 40 land.
func TestLinkDoorTailFortyLaw(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	for i := 0; i < 45; i++ {
		writeManagedNote(t, vault, fmt.Sprintf("note-%02d", i),
			fmt.Sprintf("body number %d with enough words to plan a hook", i))
	}
	f.linkOllamaChat = func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (string, string) {
		return fakeReplyJSON(`best_used_for": "use when you need thing number nine`), ""
	}
	s.RecallHooksDryRun()
	wantLogAt(t, rec, "✅ Recall hooks (dry-run) finished.", "success", 1)
	// the HEAD of the capture (the opening lines) is past the 40-line
	// window — it never lands on the desk.
	if _, ok := rec.findLog("GitHub vault: "); ok {
		t.Fatal("the capture's head lines never stream (the last-40 law)")
	}
	if _, ok := rec.findLog("Mode: DRY RUN (nothing written)"); ok {
		t.Fatal("the capture's head lines never stream (the last-40 law)")
	}
	// the TAIL does: note-44 is the last planned line; the capture
	// holds 4 header + 45 progress + blank + Report + Dry run = 52
	// lines, so the 40-line window opens at progress line 8.
	wantLogAt(t, rec, "  ~ note-44.md — would write:", "info", 2)
	wantLogAt(t, rec, "  ~ note-08.md — would write:", "info", 3)
	if l, ok := rec.findLog("  ~ note-07.md — would write:"); ok {
		t.Fatalf("progress line 7 of 45 is outside the window: %q", l.Msg)
	}
}

// ---------------------------------------------------------------------------
// 🔗 Build link suggestions (build_links.py _suggest).
// ---------------------------------------------------------------------------

func TestLinkDoorLinkSuggestionsManualMissing(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	s.cfg.Set("manual_vault_path", filepath.Join(appDir, "nope"))
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, "⏳ Link suggestions running… (the GUI stays usable)", "info", 1)
	wantLogAt(t, rec, "Manual Notes vault not set or not found (config "+
		"manual_vault_path or --manual-vault) — the Suggestions note lives "+
		"in its Library/.", "info", 2)
	wantLogAt(t, rec, "❌ Link suggestions finished with exit code 1 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 3)
}

func TestLinkDoorLinkSuggestionsLibraryMissing(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	manual := t.TempDir() // exists, no Library/
	s.cfg.Set("manual_vault_path", manual)
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, "Library/ not found in the manual vault — run "+
		"tools/mirror_manual.py --apply first (the Related blocks live in "+
		"the mirror copies).", "info", 2)
}

func TestLinkDoorLinkSuggestionsNoManagedNotes(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	manual := t.TempDir()
	if err := os.MkdirAll(filepath.Join(manual, "Library"), 0o755); err != nil {
		t.Fatal(err)
	}
	github := t.TempDir() // exists, empty
	s.cfg.Set("manual_vault_path", manual)
	s.cfg.Set("vault_path", github)
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, "Collecting recall fields github: "+github, "info", 1)
	wantLogAt(t, rec, "No managed notes found — nothing to link.", "info", 2)
	wantLogAt(t, rec, "❌ Link suggestions finished with exit code 1 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 3)
}

func TestLinkDoorLinkSuggestionsEmbedError(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	manual := t.TempDir()
	if err := os.MkdirAll(filepath.Join(manual, "Library"), 0o755); err != nil {
		t.Fatal(err)
	}
	github := t.TempDir()
	writeManagedNote(t, github, "alpha", "body")
	s.cfg.Set("manual_vault_path", manual)
	s.cfg.Set("vault_path", github)
	f.linkEmbedCall = func(cfg any, modelOverride string) (mirror.EmbedFn, string, error) {
		return nil, "", fmt.Errorf("connection refused")
	}
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, "  1 managed note(s)", "info", 1)
	wantLogAt(t, rec, "Embedding error: connection refused", "info", 2)
	wantLogAt(t, rec, "  (llama-server needs --embeddings for /v1/embeddings; "+
		"Ollama needs the model pulled, e.g. 'ollama pull nomic-embed-text')", "info", 3)
	wantLogAt(t, rec, "❌ Link suggestions finished with exit code 2 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 4)
}

// TestLinkDoorLinkSuggestionsFullFlow — the suggest grammar + THE
// ONE-WRITE LAW (the Suggestions note is the only write; the machine
// vaults stay untouched).
func TestLinkDoorLinkSuggestionsFullFlow(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	manual := t.TempDir()
	library := filepath.Join(manual, "Library")
	if err := os.MkdirAll(library, 0o755); err != nil {
		t.Fatal(err)
	}
	github := t.TempDir()
	websites := t.TempDir()
	noteA := writeManagedNote(t, github, "alpha", "The same body.")
	noteB := writeManagedNote(t, websites, "beta", "The same body.")
	beforeA, _ := os.ReadFile(noteA)
	beforeB, _ := os.ReadFile(noteB)
	s.cfg.Set("manual_vault_path", manual)
	s.cfg.Set("vault_path", github)
	s.cfg.Set("website_vault_path", websites)

	// the embeddings: identical texts → identical vectors → cosine 1.0.
	f.linkEmbedCall = func(cfg any, modelOverride string) (mirror.EmbedFn, string, error) {
		embed := func(texts []string) ([][]float64, error) {
			out := make([][]float64, len(texts))
			for i := range texts {
				out[i] = []float64{1, 0, 0}
			}
			return out, nil
		}
		return embed, "Ollama http://127.0.0.1:11434 · nomic-embed-text", nil
	}
	f.linkOllamaChat = func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (string, string) {
		return fakeReplyJSON(`related": true, "reason": "the same body`), ""
	}
	s.BuildLinkSuggestions()

	wantLogAt(t, rec, "⏳ Link suggestions running… (the GUI stays usable)", "info", 1)
	wantLogAt(t, rec, "Collecting recall fields github: "+github+
		" · websites: "+websites, "info", 2)
	wantLogAt(t, rec, "  2 managed note(s)", "info", 3)
	wantLogAt(t, rec, "  Embeddings: Ollama http://127.0.0.1:11434 · "+
		"nomic-embed-text", "info", 4)
	wantLogAt(t, rec, "  embeddings: 2 embedded, 0 fresh, 0 failed", "info", 5)
	wantLogAt(t, rec, "  candidates: 1 above 0.30 — 1 new (0 already suggested)", "info", 6)
	wantLogAt(t, rec, "  LLM: llama3 via Ollama (http://127.0.0.1:11434) "+
		"(num_ctx=8192)", "info", 7)
	wantLogAt(t, rec, "  + related: alpha ↔ beta — the same body", "info", 8)
	wantLogAt(t, rec, "Suggestions note: "+filepath.Join(library,
		mirror.SuggestionsFilename), "info", 9)
	wantLogAt(t, rec, "  pending 1 · approved 0 · rejected 0 (cap 7 links/note)", "info", 10)
	wantLogAt(t, rec, "Tick a checkbox to approve, strike a line through to "+
		"reject, then run with --collect.", "info", 11)
	wantLogAt(t, rec, "✅ Link suggestions finished.", "success", 12)

	// THE ONE-WRITE LAW: the Suggestions note under Library/, the
	// machine vaults byte-untouched.
	note, err := os.ReadFile(mirror.SuggestionsPath(manual))
	if err != nil {
		t.Fatalf("the Suggestions note: %v", err)
	}
	if !strings.Contains(string(note), "- [ ] ") {
		t.Fatalf("the note's pending checkbox: %s", note)
	}
	afterA, _ := os.ReadFile(noteA)
	afterB, _ := os.ReadFile(noteB)
	if string(beforeA) != string(afterA) || string(beforeB) != string(afterB) {
		t.Fatal("THE ONE-WRITE LAW broken — a machine vault note changed")
	}
	// the store row: pending.
	ls, err := mirror.NewLinkStore("cache.db")
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()
	counts, err := ls.Counts()
	if err != nil {
		t.Fatal(err)
	}
	if counts["pending"] != 1 {
		t.Fatalf("pending = %d", counts["pending"])
	}
}

func TestLinkDoorLinkSuggestionsRejectedRecorded(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	manual := t.TempDir()
	if err := os.MkdirAll(filepath.Join(manual, "Library"), 0o755); err != nil {
		t.Fatal(err)
	}
	github := t.TempDir()
	websites := t.TempDir()
	writeManagedNote(t, github, "alpha", "The same body.")
	writeManagedNote(t, websites, "beta", "The same body.")
	s.cfg.Set("manual_vault_path", manual)
	s.cfg.Set("vault_path", github)
	s.cfg.Set("website_vault_path", websites)
	f.linkEmbedCall = func(cfg any, modelOverride string) (mirror.EmbedFn, string, error) {
		return func(texts []string) ([][]float64, error) {
			out := make([][]float64, len(texts))
			for i := range texts {
				out[i] = []float64{1, 0, 0}
			}
			return out, nil
		}, "Ollama http://127.0.0.1:11434 · nomic-embed-text", nil
	}
	f.linkOllamaChat = func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (string, string) {
		return fakeReplyJSON(`related": false, "reason": "different worlds`), ""
	}
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, "  pending 0 · approved 0 · rejected 1 (cap 7 links/note)", "info", 1)
	if l, ok := rec.findLog("  + related:"); ok {
		t.Fatalf("a NO never prints the related line: %q", l.Msg)
	}
	// the never-resuggest rule: the pair is recorded rejected.
	ls, err := mirror.NewLinkStore("cache.db")
	if err != nil {
		t.Fatal(err)
	}
	defer ls.Close()
	known, err := ls.KnownPairsSet()
	if err != nil {
		t.Fatal(err)
	}
	if len(known) != 1 {
		t.Fatalf("known pairs = %d", len(known))
	}
}

func TestLinkDoorLinkSuggestionsFewerThanTwo(t *testing.T) {
	s, rec, f := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	manual := t.TempDir()
	if err := os.MkdirAll(filepath.Join(manual, "Library"), 0o755); err != nil {
		t.Fatal(err)
	}
	github := t.TempDir()
	writeManagedNote(t, github, "alpha", "one note only")
	s.cfg.Set("manual_vault_path", manual)
	s.cfg.Set("vault_path", github)
	f.linkEmbedCall = func(cfg any, modelOverride string) (mirror.EmbedFn, string, error) {
		return func(texts []string) ([][]float64, error) {
			out := make([][]float64, len(texts))
			for i := range texts {
				out[i] = []float64{1, 0, 0}
			}
			return out, nil
		}, "Ollama http://127.0.0.1:11434 · nomic-embed-text", nil
	}
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, "Fewer than 2 embeddings — nothing to compare (is the "+
		"embedding model pulled? e.g. ollama pull nomic-embed-text).", "info", 2)
	wantLogAt(t, rec, "❌ Link suggestions finished with exit code 1 "+
		"(see the lines above — the command-line tools print the exact next "+
		"step).", "error", 3)
}

func TestLinkDoorLinkSuggestionsBusyGuard(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	if !s.lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer s.lock.Release("batch")
	s.BuildLinkSuggestions()
	wantLogAt(t, rec, linkingBusyLine, "warning", 1)
	if _, ok := rec.findLog("⏳ Link suggestions running…"); ok {
		t.Fatal("the guarded tool never starts")
	}
}

// ---------------------------------------------------------------------------
// 📁 Recategorize Notes (dashboard.py:500-660).
// ---------------------------------------------------------------------------

func TestLinkDoorRecategorizeNoVault(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	m := s.RecategorizeOpen()
	if m.BoxTitle != "Error" || m.BoxBody != "Please select a valid vault first." {
		t.Fatalf("box = %q / %q", m.BoxTitle, m.BoxBody)
	}
	if m.Notes != nil {
		t.Fatalf("no walk on the invalid vault")
	}
}

func TestLinkDoorRecategorizeWalk(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	writeManagedNote(t, vault, "alpha", "body a")
	// a locked note (never listed, never re-categorized).
	if err := os.WriteFile(filepath.Join(vault, "locked.md"), []byte(
		"---\ncategory: Agents\ncategory_locked: true\n---\n\n# Locked\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	// an unmanaged note with no category frontmatter → Unknown.
	if err := os.WriteFile(filepath.Join(vault, "wild.md"), []byte(
		"---\nsource: https://example.com/x\n---\n\n# Wild\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	m := s.RecategorizeOpen()
	wantLogAt(t, rec, "🔒 1 locked note(s) skipped — your own moves are never "+
		"re-categorized.", "info", 1)
	if m.BoxTitle != "" {
		t.Fatalf("the table rides, no box: %+v", m)
	}
	if len(m.Notes) != 2 {
		t.Fatalf("notes = %d (want alpha + wild)", len(m.Notes))
	}
	cats := map[string]string{}
	for _, n := range m.Notes {
		cats[n.Name] = n.Category
	}
	if cats["alpha.md"] != "Agents" || cats["wild.md"] != "Unknown" {
		t.Fatalf("categories = %v", cats)
	}
	if len(m.Categories) != len(constants.CategoryKeys) {
		t.Fatalf("the category list = %d", len(m.Categories))
	}
}

func TestLinkDoorRecategorizeNoNotes(t *testing.T) {
	s, _, _ := newSettingsTest(t)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	m := s.RecategorizeOpen()
	if m.BoxTitle != "Recategorize" || m.BoxBody != "No notes found in vault." {
		t.Fatalf("box = %q / %q", m.BoxTitle, m.BoxBody)
	}
}

func TestLinkDoorRecategorizeApplyMovesAndRewrites(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	oldPath := writeManagedNote(t, vault, "alpha", "the body stays")
	// seed the cache row (the dedup twin must find the note at the NEW path).
	cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		t.Fatal(err)
	}
	if err := cache.AddProcessed(1, "https://github.com/acme/alpha", "acme",
		"alpha", oldPath, "Agents"); err != nil {
		t.Fatal(err)
	}
	cache.Close()

	res := s.RecategorizeApply([]RecategorizeChange{{
		Path: oldPath, OldCategory: "Agents", Category: "Tools",
	}})
	if res == nil || res.Moved != 1 {
		t.Fatalf("moved = %+v", res)
	}
	wantLogAt(t, rec, "📁 Recategorized 1 notes", "success", 1)

	// the frontmatter rewritten (EVERY occurrence — the reference's global sub).
	newPath := filepath.Join(vault, categoryFolderOf("Tools"), "alpha.md")
	data, err := os.ReadFile(newPath)
	if err != nil {
		t.Fatalf("the new-path note: %v", err)
	}
	if strings.Count(string(data), "category: Tools") != 1 ||
		strings.Contains(string(data), "category: Agents") {
		t.Fatalf("the rewritten frontmatter: %s", data)
	}
	// the old file is gone.
	if _, err := os.Stat(oldPath); !os.IsNotExist(err) {
		t.Fatalf("the old file must be gone: %v", err)
	}
	// the cache row moved.
	cache2, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		t.Fatal(err)
	}
	defer cache2.Close()
	row := cache2.ProcessedRowFor("https://github.com/acme/alpha")
	if row == nil || row.NotePath == nil || *row.NotePath != newPath {
		t.Fatalf("the cache row after the move: %+v", row)
	}
	if row.Category == nil || *row.Category != "Tools" {
		t.Fatalf("the cache category: %+v", row.Category)
	}
}

func TestLinkDoorRecategorizeApplyUnchangedSkipped(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	note := writeManagedNote(t, vault, "alpha", "body")
	res := s.RecategorizeApply([]RecategorizeChange{{
		Path: note, OldCategory: "Agents", Category: "Agents",
	}})
	if res.Moved != 0 {
		t.Fatalf("moved = %d", res.Moved)
	}
	wantLogAt(t, rec, "📁 Recategorized 0 notes", "success", 1)
	if _, err := os.Stat(note); err != nil {
		t.Fatalf("the untouched note: %v", err)
	}
}

func TestLinkDoorRecategorizeApplyMissingFileSpeaks(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	vault := t.TempDir()
	s.cfg.Set("vault_path", vault)
	res := s.RecategorizeApply([]RecategorizeChange{{
		Path: filepath.Join(vault, "gone.md"), OldCategory: "Agents",
		Category: "Tools",
	}})
	if res.Moved != 0 {
		t.Fatalf("moved = %d", res.Moved)
	}
	l := wantLogAt(t, rec, "Failed to recategorize gone.md:", "error", 1)
	_ = l
	wantLogAt(t, rec, "📁 Recategorized 0 notes", "success", 2)
}

func TestLinkDoorRecategorizeBusyGuard(t *testing.T) {
	s, rec, _ := newSettingsTest(t)
	s.cfg.Set("vault_path", t.TempDir())
	if !s.lock.Acquire("batch") {
		t.Fatal("acquire")
	}
	defer s.lock.Release("batch")
	m := s.RecategorizeOpen()
	if m.BoxTitle != "" || len(m.Notes) != 0 {
		t.Fatalf("the guarded open answers the empty model: %+v", m)
	}
	res := s.RecategorizeApply(nil)
	if res == nil || res.Moved != 0 {
		t.Fatalf("the guarded apply answers the empty result: %+v", res)
	}
	wantLogAt(t, rec, linkingBusyLine, "warning", 1)
	if l, _ := rec.findLog("📁 Recategorized"); strings.Contains(l.Msg, "Recategorized") {
		if len(rec.logs) > 1 {
			t.Fatalf("the guarded doors speak only the guard line")
		}
	}
}

// categoryFolderOf resolves one CATEGORY_FOLDERS entry (the test's
// local twin — the engine walks the table itself).
func categoryFolderOf(key string) string {
	for _, e := range constants.CategoryFolders {
		if e.Key == key {
			return e.Folder
		}
	}
	return "Uncategorized"
}

// ---------------------------------------------------------------------------
// THE FIELD-TAG LAW — the new crossed payloads ride their tags.
// ---------------------------------------------------------------------------

func TestLinkDoorRecategorizePayloadFieldTags(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"RecategorizeRow", RecategorizeRow{Path: "/v/a.md", Name: "a.md", Category: "Agents"},
			`{"path":"/v/a.md","name":"a.md","category":"Agents"}`},
		{"RecategorizeModel", RecategorizeModel{BoxTitle: "Error", BoxBody: "b", Notes: nil, Categories: nil},
			`{"boxTitle":"Error","boxBody":"b","notes":null,"categories":null}`},
		{"RecategorizeChange", RecategorizeChange{Path: "/v/a.md", OldCategory: "Agents", Category: "Tools"},
			`{"path":"/v/a.md","oldCategory":"Agents","category":"Tools"}`},
		{"RecategorizeResult", RecategorizeResult{Moved: 2}, `{"moved":2}`},
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
