package platformintake

// platformintake_test.go — the writer + the pruner battery: the row
// grammar (byte-exact), the header template, the insert-after-last-row
// law, the idempotence, the banned filter, the stored filter, the
// token scrub, and the prune's drop rules (the frozen clock).

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func freezeClock(t *testing.T, at time.Time) {
	t.Helper()
	orig := Clock
	Clock = func() time.Time { return at }
	t.Cleanup(func() { Clock = orig })
}

func TestWriteInboxLinksByPlatformRowGrammar(t *testing.T) {
	freezeClock(t, time.Date(2026, 10, 18, 12, 34, 0, 0, time.UTC))
	vault := t.TempDir()
	logs := []string{}
	n := WriteInboxLinksByPlatform(vault,
		[]string{"https://www.reddit.com/r/golang/comments/abc123/a_post/"},
		"Import", func(msg, level string) { logs = append(logs, level+"|"+msg) },
		nil, nil)
	if n != 1 {
		t.Fatalf("new rows = %d, want 1", n)
	}
	data, err := os.ReadFile(filepath.Join(vault, "_inbox", "reddit_links.md"))
	if err != nil {
		t.Fatalf("the reddit table never landed: %v", err)
	}
	want := `# 👽 Reddit — Review Queue

> Auto-updated. DO NOT delete rows — only update the Status column.
> Edit Status to: ✅ reviewed / ❌ ignored
> Last updated: 2026-10-18 12:34

| # | Date | URL | Domain | Source | Status | Notes |
|---|------|-----|--------|--------|--------|-------|
| - | 2026-10-18 | https://www.reddit.com/r/golang/comments/abc123/a_post/ | www.reddit.com | Import | unreviewed | |
`
	if string(data) != want {
		t.Fatalf("row grammar broke:\n--- got ---\n%s\n--- want ---\n%s", string(data), want)
	}
	// the per-platform + total log lines.
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "📥 👽 Reddit: 1 new links → reddit_links.md") {
		t.Fatalf("missing the per-platform line: %v", logs)
	}
	if !strings.Contains(joined, "📥 Total: 1 non-GitHub links added to _inbox/ (by platform)") {
		t.Fatalf("missing the total line: %v", logs)
	}
}

func TestWriteInboxLinksIdempotentAndAppend(t *testing.T) {
	freezeClock(t, time.Date(2026, 10, 18, 12, 34, 0, 0, time.UTC))
	vault := t.TempDir()
	first := []string{
		"https://www.reddit.com/r/golang/",
		"https://news.ycombinator.com/item?id=1",
	}
	if n := WriteInboxLinksByPlatform(vault, first, "Import",
		func(string, string) {}, nil, nil); n != 2 {
		t.Fatalf("first write = %d, want 2", n)
	}
	// idempotence: the same URLs again → 0 new.
	if n := WriteInboxLinksByPlatform(vault,
		[]string{"https://www.reddit.com/r/golang/"},
		"Import", func(string, string) {}, nil, nil); n != 0 {
		t.Fatalf("re-write = %d, want 0 (dedupe by normalized URL)", n)
	}
	// THE QUERY-BLIND DEDUPE QUIRK: the writer dedupes by normalize_url
	// (the query stripped) — item?id=2 reads as a duplicate of
	// item?id=1. A different PATH is a new row (appends AFTER the last
	// data row).
	if n := WriteInboxLinksByPlatform(vault,
		[]string{"https://news.ycombinator.com/item?id=2"},
		"Bot", func(string, string) {}, nil, nil); n != 0 {
		t.Fatalf("the query-blind dedupe quirk broke: %d", n)
	}
	if n := WriteInboxLinksByPlatform(vault,
		[]string{"https://news.ycombinator.com/item/2"},
		"Bot", func(string, string) {}, nil, nil); n != 1 {
		t.Fatalf("append = %d, want 1", n)
	}
	data, _ := os.ReadFile(filepath.Join(vault, "_inbox", "other_links.md"))
	lines := strings.Split(string(data), "\n")
	hn1, hn2 := -1, -1
	for i, l := range lines {
		if strings.Contains(l, "item?id=1") {
			hn1 = i
		}
		if strings.Contains(l, "item/2") {
			hn2 = i
		}
	}
	if hn1 < 0 || hn2 < 0 || hn2 < hn1 {
		t.Fatalf("the append order broke (hn1@%d hn2@%d):\n%s", hn1, hn2, string(data))
	}
}

func TestWriteInboxLinksBannedNeverCollected(t *testing.T) {
	freezeClock(t, time.Date(2026, 10, 18, 12, 34, 0, 0, time.UTC))
	vault := t.TempDir()
	logs := []string{}
	n := WriteInboxLinksByPlatform(vault,
		[]string{"https://www.youtube.com/watch?v=x", "https://x.com/someone/status/1"},
		"Import", func(msg, level string) { logs = append(logs, msg) },
		[]string{"youtube.com", "x.com"}, nil)
	if n != 0 {
		t.Fatalf("banned links collected: %d", n)
	}
	// an all-banned batch leaves no empty _inbox dir.
	if _, err := os.Stat(filepath.Join(vault, "_inbox")); !os.IsNotExist(err) {
		t.Fatalf("the empty _inbox dir leaked")
	}
	if !strings.Contains(strings.Join(logs, "\n"),
		"🚫 2 banned-domain link(s) omitted — never collected") {
		t.Fatalf("missing the omission line: %v", logs)
	}
}

func TestWriteInboxLinksStoredNotRecollected(t *testing.T) {
	freezeClock(t, time.Date(2026, 10, 18, 12, 34, 0, 0, time.UTC))
	vault := t.TempDir()
	// a note already in the vault (source: the link).
	note := "---\nsource: https://news.ycombinator.com/item?id=1\ncategory: tools\n---\n\n# Note\n"
	if err := os.WriteFile(filepath.Join(vault, "note.md"), []byte(note), 0o644); err != nil {
		t.Fatal(err)
	}
	logs := []string{}
	n := WriteInboxLinksByPlatform(vault,
		[]string{"https://news.ycombinator.com/item?id=1"},
		"Import", func(msg, level string) { logs = append(logs, msg) }, nil, nil)
	if n != 0 {
		t.Fatalf("a stored link re-collected: %d", n)
	}
	if !strings.Contains(strings.Join(logs, "\n"),
		"🔗 1 link(s) already stored as note(s) in the vault — not re-added to _inbox") {
		t.Fatalf("missing the stored line: %v", logs)
	}
}

func TestInboxTableVaultLaw(t *testing.T) {
	cfg := map[string]any{"vault_path": "/gh", "website_vault_path": "/web"}
	if got := InboxTableVault(cfg); got != "/web" {
		t.Fatalf("table vault = %q, want /web", got)
	}
	cfg2 := map[string]any{"vault_path": "/gh"}
	if got := InboxTableVault(cfg2); got != "/gh" {
		t.Fatalf("fallback = %q, want /gh", got)
	}
	if got := InboxTableVault(nil); got != "" {
		t.Fatalf("nil config = %q", got)
	}
}

func TestPruneInboxTables(t *testing.T) {
	freezeClock(t, time.Date(2026, 10, 18, 12, 34, 0, 0, time.UTC))
	vault := t.TempDir()
	inbox := filepath.Join(vault, "_inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	table := `# 🔗 Other — Review Queue

> Auto-updated. DO NOT delete rows — only update the Status column.
> Edit Status to: ✅ reviewed / ❌ ignored
> Last updated: 2026-10-01 10:00

| # | Date | URL | Domain | Source | Status | Notes |
|---|------|-----|--------|--------|--------|-------|
| - | 2026-10-01 | https://x.com/someone/status/1 | x.com | Import | unreviewed | |
| - | 2026-10-01 | https://www.youtube.com/watch?v=x | www.youtube.com | Import | unreviewed | |
| - | 2026-10-01 | https://www.keepme.dev/page | www.keepme.dev | Import | ✅ reviewed | |
`
	if err := os.WriteFile(filepath.Join(inbox, "other_links.md"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	dropped := PruneInboxTables(vault, []string{"x.com", "youtube.com"},
		func(url string) bool { return false }, func(string, string) {}, nil)
	if dropped != 2 {
		t.Fatalf("dropped = %d, want 2 (the banned rows)", dropped)
	}
	data, _ := os.ReadFile(filepath.Join(inbox, "other_links.md"))
	got := string(data)
	if strings.Contains(got, "x.com/someone") || strings.Contains(got, "youtube.com/watch") {
		t.Fatalf("the banned rows survived:\n%s", got)
	}
	if !strings.Contains(got, "https://www.keepme.dev/page") {
		t.Fatalf("the kept row vanished:\n%s", got)
	}
	if !strings.Contains(got, "✅ reviewed") {
		t.Fatalf("the owner's status mark must survive VERBATIM:\n%s", got)
	}
	if !strings.Contains(got, "> Last updated: 2026-10-18 12:34") {
		t.Fatalf("the stamp refresh broke:\n%s", got)
	}
}

func TestPruneInboxTablesStoredURLs(t *testing.T) {
	vault := t.TempDir()
	inbox := filepath.Join(vault, "_inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	table := "| - | 2026-10-01 | https://www.done.dev/ | www.done.dev | Bot | unreviewed | |\n"
	if err := os.WriteFile(filepath.Join(inbox, "other_links.md"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	// the stored_urls set drops the processed row (the caller passes
	// NORMALIZED urls — the reference's set comprehension).
	dropped := PruneInboxTables(vault, nil, nil, nil,
		[]string{"https://done.dev"})
	if dropped != 1 {
		t.Fatalf("dropped = %d, want 1", dropped)
	}
	data, _ := os.ReadFile(filepath.Join(inbox, "other_links.md"))
	if strings.Contains(string(data), "done.dev") {
		t.Fatalf("the stored row survived:\n%s", string(data))
	}
}

func TestScrubTokenRewrite(t *testing.T) {
	freezeClock(t, time.Date(2026, 10, 18, 12, 34, 0, 0, time.UTC))
	vault := t.TempDir()
	inbox := filepath.Join(vault, "_inbox")
	if err := os.MkdirAll(inbox, 0o755); err != nil {
		t.Fatal(err)
	}
	// a legacy row with a live token gets scrubbed on the next write.
	table := "| - | 2026-10-01 | https://bot.example/auth/?token=0123456789abcdef0123456789abcdef | bot.example | Bot | unreviewed | |\n"
	if err := os.WriteFile(filepath.Join(inbox, "other_links.md"), []byte(table), 0o644); err != nil {
		t.Fatal(err)
	}
	logs := []string{}
	_ = WriteInboxLinksByPlatform(vault, []string{"https://www.new.dev/"},
		"Import", func(msg, level string) { logs = append(logs, msg) }, nil, nil)
	data, _ := os.ReadFile(filepath.Join(inbox, "other_links.md"))
	if !strings.Contains(string(data), "token=…") {
		t.Fatalf("the token scrub never ran:\n%s", string(data))
	}
	if !strings.Contains(strings.Join(logs, "\n"),
		"🔒 other_links.md: scrubbed secret token(s) out of stored URL(s)") {
		t.Fatalf("missing the scrub line: %v", logs)
	}
}
