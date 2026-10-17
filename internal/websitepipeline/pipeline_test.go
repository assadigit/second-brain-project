// pipeline_test.go — T-40's state-machine table tests + THE
// CONFIRMATION GATE flows (the reference's _BanishCase/_MasterCase
// pattern: temp vault + real state DB + injected fakes — zero network).
package websitepipeline

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/webfetch"
)

// fakeFetch mirrors the reference's _FakeFetch: canned pages keyed by
// URL path; whole-path failures simulate the wall.
type fakeFetch struct {
	pages     map[string]string
	failPaths map[string]string // path → failure reason
	calls     []string
}

func newFakeFetch() *fakeFetch {
	return &fakeFetch{
		pages:     map[string]string{},
		failPaths: map[string]string{},
	}
}

func (f *fakeFetch) fail(path, reason, category string) {
	f.failPaths[path] = reason + "\x00" + category
}

func (f *fakeFetch) fetch(u string, _ *webfetch.Options) *webfetch.FetchResult {
	f.calls = append(f.calls, u)
	parsed, _ := url.Parse(u)
	path := "/"
	if parsed != nil && parsed.Path != "" {
		path = parsed.Path
	}
	if reasonCat, ok := f.failPaths[path]; ok {
		reason, cat, _ := strings.Cut(reasonCat, "\x00")
		return &webfetch.FetchResult{URL: u, Status: "failed", Reason: reason, Category: cat}
	}
	html, ok := f.pages[path]
	if !ok {
		html = "<html><head><title>Test Site</title>" +
			"<meta name=\"description\" content=\"A test page.\"></head>" +
			"<body><p>Body text about design tools and resources for " +
			"building websites and applications, long enough to " +
			"classify confidently.</p></body></html>"
	}
	status200 := 200
	return &webfetch.FetchResult{
		URL: u, FinalURL: u, Status: "full", HTTPStatus: &status200,
		ContentType: "text/html", Charset: "utf-8",
		Body: []byte(html), Text: html,
	}
}

// fakeLLM mirrors the reference's _FakeLLM: scriptable
// classify/analyze answers (valid taxonomy names).
type fakeLLM struct {
	category    string
	subcategory string
	answers     []llmclient.Message // captured
}

func (l *fakeLLM) call(messages []llmclient.Message, task string) string {
	l.answers = append(l.answers, messages...)
	text := messages[0].Content
	if strings.Contains(text, "filing a website into a personal library") {
		return fmt.Sprintf(`{"category": %q, "confidence": "high", "reason": "testing"}`, l.category)
	}
	if strings.Contains(text, "was filed under") {
		return fmt.Sprintf(`{"subcategory": %q, "confidence": "high"}`, l.subcategory)
	}
	return `{"name": "Test Site", "one_line": "A test page about design.",
		"what_it_does": ["One", "Two"],
		"best_used_for": "Use when you need to test the pipeline.",
		"pricing": "free", "login_required": "no",
		"similar_tools": [], "tags": ["design"],
		"confidence": "high"}`
}

// pipeCase is the shared plumbing (the _BanishCase pattern).
type pipeCase struct {
	t        *testing.T
	tmp      string
	vault    string
	db       *WebsiteStateDB
	inVault  map[string]bool
	logs     []logLine
	fake     *fakeFetch
	llm      *fakeLLM
	gateFn   BanishConfirmFn
	extraCfg map[string]any
}

type logLine struct {
	level string
	msg   string
}

func newPipeCase(t *testing.T) *pipeCase {
	t.Helper()
	dryrun.Disable()
	dryrun.Clear()
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "websites")
	os.MkdirAll(filepath.Join(vault, "_review"), 0o755)
	db, err := NewWebsiteStateDB(filepath.Join(tmp, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		dryrun.Disable()
		db.Close()
	})
	return &pipeCase{
		t: t, tmp: tmp, vault: vault, db: db,
		inVault: map[string]bool{},
		fake:    newFakeFetch(),
		llm:     &fakeLLM{category: "Design", subcategory: "Assets & Resources"},
	}
}

func (c *pipeCase) logFn(msg, level string) {
	c.logs = append(c.logs, logLine{level: level, msg: msg})
}

func (c *pipeCase) allLogs() string {
	var msgs []string
	for _, l := range c.logs {
		msgs = append(msgs, l.msg)
	}
	return strings.Join(msgs, "\n")
}

func (c *pipeCase) hasLog(level, msg string) bool {
	for _, l := range c.logs {
		if l.level == level && l.msg == msg {
			return true
		}
	}
	return false
}

func (c *pipeCase) hasLogContaining(part string) bool {
	for _, l := range c.logs {
		if strings.Contains(l.msg, part) {
			return true
		}
	}
	return false
}

func (c *pipeCase) makePipeline() *Pipeline {
	cfg := map[string]any{
		"website_vault_path": c.vault,
		"web_domain_delay_s": 0,
		"taxonomy_path":      "../taxonomy/testdata/website-library-categories.md",
	}
	for k, v := range c.extraCfg {
		cfg[k] = v
	}
	p, err := NewPipeline(cfg,
		func(messages []llmclient.Message, task string) string { return c.llm.call(messages, task) },
		func(u string) any {
			if c.inVault[u] {
				return true
			}
			return nil
		},
		c.db, c.logFn, nil, nil, nil,
		func(rawURL string, opts *webfetch.Options) *webfetch.FetchResult { return c.fake.fetch(rawURL, opts) },
		c.gateFn, nil)
	if err != nil {
		c.t.Fatal(err)
	}
	return p
}

func (c *pipeCase) storeProperNote(u string) (string, string) {
	c.t.Helper()
	p := c.makePipeline()
	r := p.ProcessLink(u)
	if r.Outcome != "processed" {
		c.t.Fatalf("storeProperNote(%s): outcome %q err %q", u, r.Outcome, r.Error)
	}
	c.inVault[r.Canonical] = true
	return r.Canonical, r.NotePath
}

func (c *pipeCase) writeTable(rows [][2]string) string {
	c.t.Helper()
	path := DecommissionTablePath(c.vault)
	var lines []string
	for _, row := range rows {
		domain := "unknown"
		if parsed, err := url.Parse(row[0]); err == nil && parsed.Host != "" {
			domain = parsed.Host
		}
		lines = append(lines, fmt.Sprintf("| - | 2026-10-09 | %s | %s | test | %s | |", row[0], domain, row[1]))
	}
	content := fmt.Sprintf(GraveyardHeader, "2026-10-09 12:00") + strings.Join(lines, "\n") + "\n"
	os.MkdirAll(filepath.Dir(path), 0o755)
	os.WriteFile(path, []byte(content), 0o644)
	return path
}

func (c *pipeCase) tagNote(path, marker string) {
	c.t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		c.t.Fatal(err)
	}
	text := string(data)
	if strings.Contains(text, "tags: [") {
		text = strings.Replace(text, "tags: [", "tags: ["+marker+", ", 1)
	} else {
		text = strings.Replace(text, "tags:\n", "tags:\n  - "+marker+"\n", 1)
	}
	os.WriteFile(path, []byte(text), 0o644)
}

// --- THE STATE-MACHINE TABLE (every gate/outcome) --------------------------

func TestStateMachineTable(t *testing.T) {
	type gate struct {
		name string
		run  func(c *pipeCase) Result
	}
	gates := []gate{
		{"fresh_link_processes",
			func(c *pipeCase) Result { return c.makePipeline().ProcessLink("https://fresh.example/x") }},
		{"directory_banish_gate_skips",
			func(c *pipeCase) Result {
				os.MkdirAll(filepath.Join(c.vault, "_system"), 0o755)
				os.WriteFile(filepath.Join(c.vault, "_system", "banished.json"),
					[]byte(`[{"host": "banished.example"}]`), 0o644)
				return c.makePipeline().ProcessLink("https://banished.example/page")
			}},
		{"blocked_domain_skips",
			func(c *pipeCase) Result {
				c.extraCfg = map[string]any{"web_blocked_domains": []any{"blocked.example"}}
				return c.makePipeline().ProcessLink("https://blocked.example/x")
			}},
		{"self_domain_skips",
			func(c *pipeCase) Result {
				c.extraCfg = map[string]any{"web_self_domains": []any{"self.example"}}
				return c.makePipeline().ProcessLink("https://self.example/auth?token=abc")
			}},
		{"graveyard_dismissed_skips",
			func(c *pipeCase) Result {
				c.db.Dismiss("https://grave.example/g", "decommissioned by owner — graveyard table (v0.44.0)")
				c.writeTable([][2]string{{"https://grave.example/g", "🪦 dead"}})
				return c.makePipeline().ProcessLink("https://grave.example/g")
			}},
		{"banished_dismissed_skips",
			func(c *pipeCase) Result {
				c.db.Dismiss("https://gone.example/g", "banished by owner — 🗑️ note tag (delete); the note left the library, the URL is blacklisted, never fetched again (v0.58.0)")
				return c.makePipeline().ProcessLink("https://gone.example/g")
			}},
		{"auto_verdict_dismissed_skips",
			func(c *pipeCase) Result {
				c.db.Dismiss("https://auto.example/g", "auto-verdict: dead — HTTP 404")
				return c.makePipeline().ProcessLink("https://auto.example/g")
			}},
		{"plain_dismissed_skips",
			func(c *pipeCase) Result {
				c.db.Dismiss("https://plain.example/g", "whatever")
				return c.makePipeline().ProcessLink("https://plain.example/g")
			}},
		{"already_in_vault_skips",
			func(c *pipeCase) Result {
				canonical, _ := c.storeProperNote("https://stored.example/s")
				return c.makePipeline().ProcessLink(canonical)
			}},
		{"failed_review_prior_upgrades",
			func(c *pipeCase) Result {
				// first run fails → a failed _review note + retry row
				c.fake.fail("/w", "HTTP 404 — not found", "dead")
				// dead never requeues — use a healable failure instead
				c.fake.failPaths = map[string]string{"/w": "connection reset\x00network"}
				r1 := c.makePipeline().ProcessLink("https://up.example/w")
				if r1.Outcome != "review" {
					t.Fatalf("setup: outcome %q", r1.Outcome)
				}
				// the page heals: the next run must UPGRADE, not skip
				c.fake.failPaths = map[string]string{}
				return c.makePipeline().ProcessLink("https://up.example/w")
			}},
		{"fetch_failure_reviews_and_enqueues",
			func(c *pipeCase) Result {
				c.fake.fail("/f", "connection reset", "network")
				return c.makePipeline().ProcessLink("https://fail.example/f")
			}},
		{"no_heal_verdict_dismisses",
			func(c *pipeCase) Result {
				c.fake.fail("/d", "HTTP 404 — not found", "dead")
				return c.makePipeline().ProcessLink("https://dead.example/d")
			}},
		{"walled_reason_gains_the_door_hint",
			func(c *pipeCase) Result {
				c.fake.fail("/w", "HTTP 403 — bot defense (Cloudflare: challenge)", "refused")
				return c.makePipeline().ProcessLink("https://wall.example/w")
			}},
		{"archived_rescue_files_partial",
			func(c *pipeCase) Result {
				c.fake.failPaths = map[string]string{"/a": "archived copy\x00archived"}
				return c.makePipeline().ProcessLink("https://arch.example/a")
			}},
	}
	for _, g := range gates {
		t.Run(g.name, func(t *testing.T) {
			c := newPipeCase(t)
			r := g.run(c)
			// every outcome is one of the four; nothing panics, nothing
			// silently drops (a note exists for review/processed).
			switch r.Outcome {
			case "processed":
				if r.NotePath == "" {
					t.Error("processed without a note path")
				}
			case "review":
				if r.NotePath == "" {
					t.Error("review without a placeholder note")
				}
				if _, err := os.Stat(r.NotePath); err != nil {
					t.Errorf("review note missing on disk: %v", err)
				}
			case "skipped":
				if r.Error == "" {
					t.Error("skipped without a reason")
				}
			default:
				t.Errorf("unexpected outcome %q (err %q)", r.Outcome, r.Error)
			}
		})
	}
}

func TestStateMachineOutcomeDetails(t *testing.T) {
	t.Run("fresh_link_processes_and_logs_destination", func(t *testing.T) {
		c := newPipeCase(t)
		r := c.makePipeline().ProcessLink("https://fresh.example/x")
		if r.Outcome != "processed" || r.Category != "Design" || r.Subcategory != "Assets & Resources" {
			t.Fatalf("outcome %+v", r)
		}
		if !strings.Contains(filepath.ToSlash(r.NotePath), "Design/Assets_Resources/") {
			t.Errorf("note not filed under the taxonomy folder: %s", r.NotePath)
		}
		if !c.hasLog("success", "✅ [websites] Test Site processed and stored → Design/Assets_Resources/Test_Site.md") {
			t.Errorf("the destination log line missing:\n%s", c.allLogs())
		}
		if !strings.Contains(readFileT(t, r.NotePath), "## Best used for") {
			t.Error("the note lacks the Best used for heading")
		}
	})

	t.Run("blocked_domain_skips_with_the_law", func(t *testing.T) {
		c := newPipeCase(t)
		c.extraCfg = map[string]any{"web_blocked_domains": []any{"blocked.example"}}
		r := c.makePipeline().ProcessLink("https://blocked.example/x")
		if r.Outcome != "skipped" || r.Error != "blocked domain — omitted (never collected)" {
			t.Fatalf("result %+v", r)
		}
		if len(c.fake.calls) != 0 {
			t.Error("a blocked domain was fetched")
		}
	})

	t.Run("banished_skip_names_the_door", func(t *testing.T) {
		c := newPipeCase(t)
		c.db.Dismiss("https://gone.example/g", BanishReasonPrefix+" — 🗑️ note tag (delete); the note left the library, the URL is blacklisted, never fetched again (v0.58.0)")
		r := c.makePipeline().ProcessLink("https://gone.example/g")
		if !strings.Contains(r.Error, "banished by owner") {
			t.Errorf("error %q", r.Error)
		}
		if !c.hasLogContaining("banished by owner — skipped") {
			t.Errorf("the door's log line missing:\n%s", c.allLogs())
		}
	})

	t.Run("fetch_failure_writes_review_note_and_enqueues", func(t *testing.T) {
		c := newPipeCase(t)
		c.fake.fail("/f", "connection reset", "network")
		r := c.makePipeline().ProcessLink("https://fail.example/f")
		if r.Outcome != "review" || !strings.Contains(r.Error, "fetch failed") {
			t.Fatalf("result %+v", r)
		}
		if !strings.Contains(readFileT(t, r.NotePath), "Needs review — failed") {
			t.Error("the placeholder lacks the warning callout")
		}
		rr, err := c.db.RetryRow("https://fail.example/f")
		if err != nil || rr == nil {
			t.Fatalf("no retry row: %v", err)
		}
		if rr.Attempts != 1 {
			t.Errorf("attempts: %d", rr.Attempts)
		}
		if ptrStr(rr.LastError) != "connection reset" {
			t.Errorf("last_error: %q", ptrStr(rr.LastError))
		}
	})

	t.Run("no_heal_verdict_dismisses_and_never_requeues", func(t *testing.T) {
		c := newPipeCase(t)
		c.fake.fail("/d", "HTTP 404 — not found", "dead")
		r := c.makePipeline().ProcessLink("https://dead.example/d")
		if r.Outcome != "review" {
			t.Fatalf("result %+v", r)
		}
		dismissed, _ := c.db.IsDismissed("https://dead.example/d")
		if !dismissed {
			t.Error("the auto-verdict did not dismiss")
		}
		rr, _ := c.db.RetryRow("https://dead.example/d")
		if rr != nil {
			t.Error("a dead verdict requeued")
		}
	})

	t.Run("walled_reason_carries_the_fourth_door_hint", func(t *testing.T) {
		c := newPipeCase(t)
		c.fake.fail("/w", "HTTP 403 — bot defense (Cloudflare: challenge)", "refused")
		r := c.makePipeline().ProcessLink("https://wall.example/w")
		if !strings.Contains(r.Error, HandDoorHint) {
			t.Errorf("the wall's reason lacks the door hint: %q", r.Error)
		}
	})

	t.Run("no_more_retries_skips", func(t *testing.T) {
		c := newPipeCase(t)
		c.fake.fail("/b", "connection reset", "network")
		p := c.makePipeline()
		p.ProcessLink("https://burn.example/b")
		// burn out the counter (attempts >= 3)
		for i := 0; i < 2; i++ {
			c.db.EnqueueRetry("https://burn.example/b", "again")
		}
		r := c.makePipeline().ProcessLink("https://burn.example/b")
		if r.Outcome != "skipped" || !strings.Contains(r.Error, "no more retries") {
			t.Fatalf("result %+v", r)
		}
	})

	t.Run("locked_row_overrides_the_classifier", func(t *testing.T) {
		c := newPipeCase(t)
		// a locked note_state row with the owner's own placement
		db := openNoteState(t, c.tmp)
		cat := "AI Tools for Web & App Development"
		db.RecordNote("websites", "https://locked.example/l", "", nil, &cat, nil, nil)
		db.SetLocked("websites", "https://locked.example/l", true)
		p := c.makePipeline()
		p.noteStateDB = db
		r := p.ProcessLink("https://locked.example/l")
		if r.Outcome != "processed" || r.Category != cat {
			t.Fatalf("result %+v", r)
		}
		if !c.hasLogContaining("locked note — owner's placement") {
			t.Errorf("the locked log line missing:\n%s", c.allLogs())
		}
	})

	t.Run("consolidation_into_the_site_note", func(t *testing.T) {
		c := newPipeCase(t)
		_, note := c.storeProperNote("https://site.example/first")
		p := c.makePipeline()
		p.siteNoteFor = func(u string) string { return note }
		r := p.ProcessLink("https://site.example/second")
		if r.Outcome != "processed" {
			t.Fatalf("result %+v", r)
		}
		if r.NotePath != note {
			t.Errorf("the link defined a second note: %s", r.NotePath)
		}
		content := readFileT(t, note)
		// the fresh note had no site_links line: the consolidation adds
		// only the NEW link to the list (the note's own source is the
		// anchor, not a list member)
		if !strings.Contains(content, "site_links: ["+"https://site.example/second]") {
			t.Errorf("the site_links line missing")
		}
		if !strings.Contains(content, "## Links on this site") {
			t.Error("the body section missing")
		}
		if !c.hasLogContaining("the site's note already holds this domain") {
			t.Errorf("the consolidation log missing:\n%s", c.allLogs())
		}
		if p.counters["consolidated"] != 1 {
			t.Errorf("counters: %+v", p.counters)
		}
	})

	t.Run("settled_link_is_never_refetched", func(t *testing.T) {
		c := newPipeCase(t)
		c.db.SettleExisting([]string{"https://settled.example/s"})
		p := c.makePipeline()
		results := p.Run([]string{"https://settled.example/s"}, nil, nil)
		if results[0].Outcome != "skipped" || !strings.Contains(results[0].Error, "settled") {
			t.Fatalf("result %+v", results[0])
		}
		if len(c.fake.calls) != 0 {
			t.Error("a settled link was fetched")
		}
	})

	t.Run("duplicate_within_batch_is_reported", func(t *testing.T) {
		c := newPipeCase(t)
		p := c.makePipeline()
		results := p.Run([]string{"https://dup.example/a", "https://dup.example/a"}, nil, nil)
		if results[1].Outcome != "skipped" || results[1].Error != "duplicate within batch" {
			t.Fatalf("second result %+v", results[1])
		}
	})
}

// openNoteState opens a temp notestate DB (the locked-row test's seam).
func openNoteState(t *testing.T, tmp string) *notestate.DB {
	t.Helper()
	db, err := notestate.New(filepath.Join(tmp, "notes.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { db.Close() })
	return db
}
