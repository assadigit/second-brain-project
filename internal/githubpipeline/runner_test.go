// runner_test.go — T-42's behavior battery: the gates in order, the
// 404-strike flow, the rate-limit laws (fake clock + sleep recorder),
// the 401-anonymous fallback, the cache dedupe, the fork→parent
// analysis, the review/quality routing, the disk-full protocol, the
// delays, and the undo diff.
package githubpipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// ---------------------------------------------------------------------------
// The fakes.
// ---------------------------------------------------------------------------

// fakeFetcher is the "fake repo" of the spec's test contract.
type fakeFetcher struct {
	mu          sync.Mutex
	rateLimit   *RateLimit
	rateErr     error
	repos       map[string]*Repo // by full name
	repoErrs    map[string]error // by full name (sticky)
	repoErrOnce map[string]error // by full name (fires once, clears)
	topics      map[string][]string
	commits     map[string]int
	readmes     map[string]*ReadmeResult
	languages   map[string][]string
	releases    map[string]*Release
	getRepoLog  []string
}

func newFakeFetcher() *fakeFetcher {
	return &fakeFetcher{
		repos:       map[string]*Repo{},
		repoErrs:    map[string]error{},
		repoErrOnce: map[string]error{},
		topics:      map[string][]string{},
		commits:     map[string]int{},
		readmes:     map[string]*ReadmeResult{},
		languages:   map[string][]string{},
		releases:    map[string]*Release{},
	}
}

func (f *fakeFetcher) GetRateLimit() (*RateLimit, error) {
	if f.rateErr != nil {
		return nil, f.rateErr
	}
	if f.rateLimit == nil {
		return &RateLimit{Remaining: 5000,
			Reset: time.Now().Add(time.Hour)}, nil
	}
	return f.rateLimit, nil
}

func (f *fakeFetcher) GetRepo(fullName string) (*Repo, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.getRepoLog = append(f.getRepoLog, fullName)
	if err, ok := f.repoErrOnce[fullName]; ok {
		delete(f.repoErrOnce, fullName)
		return nil, err
	}
	if err, ok := f.repoErrs[fullName]; ok {
		return nil, err
	}
	if r, ok := f.repos[fullName]; ok {
		return r, nil
	}
	return nil, &GHError{Status: 404, Message: "404 {'message': 'Not Found'}"}
}

func (f *fakeFetcher) GetTopics(fullName string) ([]string, error) {
	return f.topics[fullName], nil
}

func (f *fakeFetcher) CountCommitsSince(fullName string,
	since time.Time) (int, error) {
	return f.commits[fullName], nil
}

func (f *fakeFetcher) GetReadme(fullName string) (*ReadmeResult, error) {
	if r, ok := f.readmes[fullName]; ok {
		return r, nil
	}
	return nil, fmt.Errorf("no readme")
}

func (f *fakeFetcher) GetLanguages(fullName string) ([]string, error) {
	return f.languages[fullName], nil
}

func (f *fakeFetcher) LatestRelease(fullName string) (*Release, error) {
	if r, ok := f.releases[fullName]; ok {
		return r, nil
	}
	return nil, nil
}

func (f *fakeFetcher) repoCalls() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return len(f.getRepoLog)
}

// logSink records (level, msg) pairs.
type logSink struct {
	mu    sync.Mutex
	lines []logLine
}

type logLine struct {
	level string
	msg   string
}

func (l *logSink) emit(msg, level string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	l.lines = append(l.lines, logLine{level, msg})
}

func (l *logSink) all() []logLine {
	l.mu.Lock()
	defer l.mu.Unlock()
	return append([]logLine{}, l.lines...)
}

func (l *logSink) hasSub(substr string) bool {
	for _, ln := range l.all() {
		if strings.Contains(ln.msg, substr) {
			return true
		}
	}
	return false
}

// fakeClock is the frozen clock; sleeps advance it (the wait loops
// terminate instantly).
type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func newFakeClock(t time.Time) *fakeClock { return &fakeClock{now: t} }

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) sleep(d time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.now = c.now.Add(d)
}

// sleepRecorder counts the sleep durations without advancing anything.
type sleepRecorder struct {
	mu    sync.Mutex
	calls []time.Duration
}

func (s *sleepRecorder) sleep(d time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.calls = append(s.calls, d)
}

// fakeBanner is the banner seam (the "fake banner" of the spec).
type fakeBanner struct {
	calls  []string
	result string
}

func (b *fakeBanner) fn() func(owner, repoName, folder string) string {
	return func(owner, repoName, folder string) string {
		b.calls = append(b.calls, owner+"/"+repoName)
		return b.result
	}
}

// harness is the loop test rig: a tempdir vault, a real CacheDB, a
// real VaultIndex + LinkTracker, a fake fetcher, and the frozen seams.
type harness struct {
	t       *testing.T
	dir     string
	vault   string
	cache   *store.CacheDB
	fetcher *fakeFetcher
	runner  *Runner
	logs    *logSink
	clock   *fakeClock
}

func newHarness(t *testing.T, config map[string]any) *harness {
	t.Helper()
	dir := t.TempDir()
	vault := filepath.Join(dir, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	cache, err := store.NewCacheDB(filepath.Join(dir, "cache.db"))
	if err != nil {
		t.Fatal(err)
	}
	if config == nil {
		config = map[string]any{}
	}
	config["vault_path"] = vault
	if _, ok := config["llm_provider"]; !ok {
		config["llm_provider"] = "cloud" // the tests fake the cloud path
	}
	logs := &logSink{}
	clock := newFakeClock(time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC))
	fetcher := newFakeFetcher()
	r := NewRunner(config)
	r.Cache = cache
	r.VaultIndex = batch.NewVaultIndex(vault, nil)
	r.LinkTracker = batch.NewLinkTracker(vault)
	r.NewGHClient = func(token string) RepoFetcher { return fetcher }
	r.Log = logs.emit
	r.Clock = clock.Now
	r.Sleep = clock.sleep
	r.LLMProvider = "cloud"
	r.BannerFn = func(owner, repoName, folder string) string {
		return "" // the fake banner: never fetched
	}
	r.running = true
	h := &harness{t: t, dir: dir, vault: vault, cache: cache,
		fetcher: fetcher, runner: r, logs: logs, clock: clock}
	t.Cleanup(func() { cache.Close() })
	return h
}

// fakeLLM installs a cloud-mode fake (the _call_cloud_llm
// monkeypatch pattern of the reference's own tests).
func fakeLLM(t *testing.T, fn func(messages []llmclient.Message,
	jsonMode bool) (string, error)) {
	t.Helper()
	orig := callCloudLLM
	callCloudLLM = func(apiURL, apiKey, model string,
		messages []llmclient.Message, timeoutS float64,
		opts *llmclient.ChatOptions) (string, error) {
		jsonMode := opts == nil || opts.JSONMode
		return fn(messages, jsonMode)
	}
	t.Cleanup(func() { callCloudLLM = orig })
}

// okResult is a canned valid LLM answer.
const okResult = `{"summary":"A solid tool for testing things with plenty of words to say here.",
	"how_it_works":"It works.","core_value":"Value.",
	"features":["a","b","c"],"difference":"Different.",
	"category":"Dev-Tools","confidence":80,"tags":["x"],
	"short_summary":"short"}`

// repoIDSeq gives the fakes unique IDs (the cache dedupe keys on them).
var repoIDSeq int64

// newRepo builds a plain successful repo.
func newRepo(fullName string, stars int) *Repo {
	owner, name, _ := strings.Cut(fullName, "/")
	repoIDSeq++
	return &Repo{
		ID: repoIDSeq, FullName: fullName, Name: name,
		StargazersCount: stars, ForksCount: 3,
		Description: "A " + name + " description long enough for docs.",
		OwnerLogin:  owner, Language: "Go",
		HasIssues: true, HasWiki: true, HasLicense: true,
	}
}

// ---------------------------------------------------------------------------
// The gates.
// ---------------------------------------------------------------------------

// TestLoopSkipGates pins the gate ORDER and the no-API-call law of the
// vault-index dedupe (v0.38.0).
func TestLoopSkipGates(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	deadURL := "https://github.com/dead/repo"
	h.cache.Record404(deadURL, "x")
	h.cache.ConfirmDead(deadURL, "x", 1)
	dismissedURL := "https://github.com/gone/repo"
	h.cache.AddProcessed(1, dismissedURL, "gone", "repo", "/old.md", "Dev-Tools")
	settledURL := "https://github.com/settled/repo"
	h.cache.SettleRepos([]string{settledURL})
	// The vault-index dedupe case: a note with source: <url>.
	existing := filepath.Join(h.vault, "Existing.md")
	os.WriteFile(existing, []byte("---\nsource: https://github.com/exist/repo\n---\n# x"),
		0o644)
	h.runner.VaultIndex.Rebuild(nil)

	urls := []string{
		"https://example.com/not-github", // non-GitHub
		deadURL,                          // 404 quarantine
		dismissedURL,                     // dismissed (note deleted)
		settledURL,                       // settled ledger
		"https://github.com/onlyowner",   // invalid (<2 parts)
		"https://github.com/exist/repo",  // vault dedupe
	}
	h.runner.Total = len(urls)
	dead := h.cache.GetDeadURLSet(1) // 1 strike + confirmed = dead at 1
	settled := h.cache.GetSettledRepoSet()
	dismissed := map[string]bool{links.NormalizeURL(dismissedURL): true}
	h.runner.RunGitHubLoop(urls, dead, dismissed, settled, 3)

	if calls := h.fetcher.repoCalls(); calls != 0 {
		t.Fatalf("the skip gates must never reach the API, got %d calls",
			calls)
	}
	if h.runner.DeadSkipped != 1 || h.runner.SettledReposSkipped != 1 ||
		h.runner.DismissedSkipped != 1 {
		t.Fatalf("skip counters: dead=%d settled=%d dismissed=%d",
			h.runner.DeadSkipped, h.runner.SettledReposSkipped,
			h.runner.DismissedSkipped)
	}
	if !h.logs.hasSub("Skipping non-GitHub URL") ||
		!h.logs.hasSub("Dismissed (you deleted its note)") ||
		!h.logs.hasSub("Invalid GitHub URL") ||
		!h.logs.hasSub("Already in vault") {
		t.Fatalf("gate logs missing:\n%v", h.logs.all())
	}
	// Every skipped URL's manifest row is terminal and its retry row
	// is settled.
	for _, u := range urls[:5] {
		if h.cache.GetFailedCount() != 0 {
			t.Fatalf("skips must not create failed rows")
		}
		_ = u
	}
	ml := h.runner.LinkTracker.Manifest.Links
	if len(ml) != 0 {
		t.Fatalf("no intake → no manifest rows")
	}
}

// TestLoop404Strike pins the full 404 path: the strike is counted, the
// _missing note is written, the quarantine is confirmed on the spot,
// the notfound-links.md line lands, and the manifest row is skipped.
func TestLoop404Strike(t *testing.T) {
	h := newHarness(t, nil)
	url := "https://github.com/gone/repo"
	h.runner.LinkTracker.Intake([]string{url}, nil)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)

	// The strike is recorded (1 strike).
	rows := h.cache.GetUnconfirmed404s(3)
	if len(rows) != 0 {
		t.Fatalf("the 404 was CONFIRMED immediately, unconfirmed = %d",
			len(rows))
	}
	if !h.cache.IsDeadLink(url, 3) {
		t.Fatalf("confirm_dead must make the link dead at any threshold")
	}
	// The _missing note exists with the source key.
	notePath := filepath.Join(h.vault, "_missing",
		storage.SafeFilename("gone_repo_missing")+".md")
	data, err := os.ReadFile(notePath)
	if err != nil {
		t.Fatalf("missing-repo note: %v", err)
	}
	if !strings.Contains(string(data), "source: "+url) {
		t.Fatalf("note lacks the source key:\n%s", data)
	}
	// The notfound-links.md line.
	nf, err := os.ReadFile(filepath.Join(h.vault, "_inbox",
		"notfound-links", "notfound_links.md"))
	if err != nil {
		t.Fatalf("notfound links file: %v", err)
	}
	want := "| 2026-10-17 | " + url +
		" | 404 Not Found (missing-repo note) |"
	if !strings.Contains(string(nf), want) {
		t.Fatalf("notfound line missing: %q in %q", want, string(nf))
	}
	// The manifest row: skipped.
	if ml := h.runner.LinkTracker.Manifest.Links; len(ml) != 1 ||
		ml[0].Status != "skipped" {
		t.Fatalf("manifest: %+v", ml)
	}
	if !h.logs.hasSub("Repo not found (404)") {
		t.Fatalf("404 log missing")
	}
	// The idempotence: a second run writes no duplicate strike.
	h.runner.missingNotesWritten = map[string]bool{}
	h.cache.Record404(url, "again") // a fresh strike would be 2
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	nf2, _ := os.ReadFile(filepath.Join(h.vault, "_inbox",
		"notfound-links", "notfound_links.md"))
	if got := strings.Count(string(nf2), url); got != 1 {
		t.Fatalf("notfound-links appended twice: %d", got)
	}
}

// TestLoop404AfterRateLimitWait pins the v25 pre-flight 403 branch:
// the hour wait (chunked sleeps), then the 404 on the retry.
func TestLoop404AfterRateLimitWait(t *testing.T) {
	h := newHarness(t, nil)
	url := "https://github.com/late/dead"
	h.fetcher.repoErrOnce["late/dead"] = &GHError{Status: 403,
		Message: "403 {'message': \"API rate limit exceeded\"}"}
	// The retry after the wait 404s (no repo registered).
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if !h.logs.hasSub("Waiting 3600s (1 hour) for reset...") {
		t.Fatalf("rate-limit wait log missing")
	}
	if !h.logs.hasSub("Repo not found (404)") {
		t.Fatalf("post-wait 404 missing")
	}
	if calls := h.fetcher.repoCalls(); calls != 2 {
		t.Fatalf("the repo must be retried after the wait, calls = %d", calls)
	}
	// The frozen clock advanced past the wait.
	if h.clock.Now().Sub(time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)) < 3605*time.Second {
		t.Fatalf("the clock did not advance past the wait: %v", h.clock.Now())
	}
}

// TestLoopRateLimitWaitResumes pins the successful resume path.
func TestLoopRateLimitWaitResumes(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	url := "https://github.com/late/alive"
	h.fetcher.repoErrOnce["late/alive"] = &GHError{Status: 403,
		Message: "403 {'message': \"API rate limit exceeded\"}"}
	h.fetcher.repos["late/alive"] = newRepo("late/alive", 50)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if h.runner.Processed != 1 {
		t.Fatalf("the repo must process after the wait")
	}
	if !h.logs.hasSub("resuming with this repo") {
		t.Fatalf("resume log missing")
	}
}

// TestLoop401AnonymousFallback pins the v32.1 token-drop law: the
// once-logs, the anonymous retry, and the never-re-enter branch.
func TestLoop401AnonymousFallback(t *testing.T) {
	h := newHarness(t, map[string]any{"github_token": "tok"})
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	url := "https://github.com/auth/repo"
	// The token client 401s; the anonymous client succeeds.
	anon := newFakeFetcher()
	anon.repos["auth/repo"] = newRepo("auth/repo", 10)
	tokenFails := newFakeFetcher()
	tokenFails.repoErrs["auth/repo"] = &GHError{Status: 401,
		Message: "401 {'message': 'Bad credentials'}"}
	state := 0
	h.runner.NewGHClient = func(token string) RepoFetcher {
		if token == "" {
			state = 2
			return anon
		}
		state = 1
		return tokenFails
	}
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if state != 2 {
		t.Fatalf("the anonymous client must take over, state=%d", state)
	}
	if h.runner.Processed != 1 {
		t.Fatalf("the anonymous retry must process the repo")
	}
	if got := countLog(h.logs, "GitHub token rejected"); got != 1 {
		t.Fatalf("the rejection log must fire exactly once, got %d", got)
	}
	if got := countLog(h.logs, "Continuing this batch anonymously"); got != 1 {
		t.Fatalf("the anonymous notice must fire exactly once, got %d", got)
	}
	// A second 401 repo: no more logs (the token stays dropped).
	url2 := "https://github.com/auth/repo2"
	anon.repos["auth/repo2"] = newRepo("auth/repo2", 20)
	h.runner.Total = 2
	h.runner.RunGitHubLoop([]string{url2}, nil, nil, nil, 3)
	if got := countLog(h.logs, "GitHub token rejected"); got != 1 {
		t.Fatalf("the rejection log fired again: %d", got)
	}
}

// TestLoopOtherGitHubErrorIsFailed pins the generic error branch: the
// retry-queue row + the manifest failure.
func TestLoopOtherGitHubErrorIsFailed(t *testing.T) {
	h := newHarness(t, nil)
	url := "https://github.com/bad/repo"
	h.fetcher.repoErrs["bad/repo"] = &GHError{Status: 500,
		Message: "500 {'message': 'boom'}"}
	h.runner.LinkTracker.Intake([]string{url}, nil)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if !h.logs.hasSub("GitHub API error") {
		t.Fatalf("error log missing")
	}
	// The reference's GitHub-level error path marks the MANIFEST only
	// (the retry queue row is the per-URL except's job) — the next
	// launch's reconciliation owns the retry.
	if h.cache.GetFailedCount() != 0 {
		t.Fatalf("a GitHub-level error must not stack a queue row: %d",
			h.cache.GetFailedCount())
	}
	if ml := h.runner.LinkTracker.Manifest.Links; len(ml) != 1 ||
		ml[0].Status != "failed" {
		t.Fatalf("manifest: %+v", ml)
	}
}

func countLog(l *logSink, substr string) int {
	n := 0
	for _, ln := range l.all() {
		if strings.Contains(ln.msg, substr) {
			n++
		}
	}
	return n
}

// TestLoopCacheDedupe pins layer 2: the SQLite cache skip when the note
// exists, and the self-healing reprocess when it does not.
func TestLoopCacheDedupe(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	url := "https://github.com/dup/repo"
	repo := newRepo("dup/repo", 10)
	h.fetcher.repos["dup/repo"] = repo
	// A stale cache row whose note file still exists.
	note := filepath.Join(h.vault, "Dup.md")
	os.WriteFile(note, []byte(strings.Repeat("x", 200)), 0o644)
	h.cache.AddProcessed(repo.ID, url, "dup", "repo", note, "Dev-Tools")
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if h.runner.Processed != 0 {
		t.Fatalf("a valid note must skip processing")
	}
	if !h.logs.hasSub("Already processed (note exists)") {
		t.Fatalf("dedupe log missing")
	}
	// The vault index gained the URL (future runs dedupe on layer 1).
	if !h.runner.VaultIndex.HasURL(url) {
		t.Fatalf("the cache dedupe must backfill the vault index")
	}

	// The self-healing path: the note file is gone → re-process.
	h.runner.VaultIndex = batch.NewVaultIndex(h.vault, nil)
	h.runner.VaultIndex.Rebuild(nil)
	os.Remove(note)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if h.runner.Processed != 1 {
		t.Fatalf("a deleted note must re-process (got %d)",
			h.runner.Processed)
	}
	if !h.logs.hasSub("Note file was deleted, re-processing") {
		t.Fatalf("reprocess log missing")
	}
}

// TestLoopForkAnalyzesParent pins the fork→parent law.
func TestLoopForkAnalyzesParent(t *testing.T) {
	h := newHarness(t, nil)
	var seenName string
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		for _, msg := range m {
			if i := strings.Index(msg.Content, "Name: "); i >= 0 {
				rest := msg.Content[i+6:]
				if j := strings.Index(rest, "\n"); j >= 0 {
					seenName = rest[:j]
				}
			}
		}
		return okResult, nil
	})
	parent := newRepo("parent/upstream", 900)
	fork := newRepo("fork/repo", 5)
	fork.Fork = true
	fork.Parent = parent
	h.fetcher.repos["fork/repo"] = fork
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/fork/repo"},
		nil, nil, nil, 3)
	if seenName != "upstream" {
		t.Fatalf("the LLM must see the PARENT name, got %q", seenName)
	}
	if !h.logs.hasSub("🍴 Fork detected — analyzing parent: parent/upstream") {
		t.Fatalf("fork log missing")
	}
	// The note filename is the parent's.
	found := false
	_ = filepath.Walk(h.vault, func(p string, info os.FileInfo,
		err error) error {
		if err == nil && strings.HasPrefix(info.Name(), "upstream_") {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatalf("the note must be named after the parent")
	}
	// The cache row points at the PARENT's repo id.
	if dup, _ := h.cache.IsDuplicate(parent.ID); !dup {
		t.Fatalf("the cache must record the parent id")
	}
}

// TestLoopReviewAndQualityRouting pins the _review laws.
func TestLoopReviewAndQualityRouting(t *testing.T) {
	lowConf := strings.Replace(okResult, `"confidence":80`, `"confidence":45`, 1)
	veryLow := strings.Replace(okResult, `"confidence":80`, `"confidence":20`, 1)
	cases := []struct {
		name    string
		llm     string
		wantDir string
	}{
		{"low confidence (45) → _review", lowConf, "_review"},
		{"very low (20) → _review + quality", veryLow, "_review"},
		{"high confidence → the category folder", okResult, "Tools/Dev-Tools"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			h := newHarness(t, nil)
			fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
				return tc.llm, nil
			})
			h.fetcher.repos["o/r"] = newRepo("o/r", 10)
			h.runner.Total = 1
			h.runner.RunGitHubLoop([]string{"https://github.com/o/r"},
				nil, nil, nil, 3)
			if h.runner.Processed != 1 {
				t.Fatalf("processed = %d", h.runner.Processed)
			}
			found := false
			_ = filepath.Walk(h.vault, func(p string, info os.FileInfo,
				err error) error {
				if err == nil && strings.HasSuffix(p, ".md") &&
					strings.Contains(filepath.ToSlash(p), tc.wantDir) {
					found = true
				}
				return nil
			})
			if !found {
				t.Fatalf("note not filed under %s", tc.wantDir)
			}
		})
	}
	// A category outside CATEGORY_KEYS with confidence ≥50 →
	// Uncategorized.
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return strings.Replace(okResult, `"category":"Dev-Tools"`,
			`"category":"Made-Up-Category"`, 1), nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	found := false
	_ = filepath.Walk(h.vault, func(p string, info os.FileInfo,
		err error) error {
		if err == nil && strings.Contains(filepath.ToSlash(p),
			"Uncategorized") {
			found = true
		}
		return nil
	})
	if !found {
		t.Fatalf("an unknown category must fall to Uncategorized")
	}
}

// TestLoopDiskFullHeadlessSkips pins the v30 headless hang-bomb fix.
func TestLoopDiskFullHeadlessSkips(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.runner.Headless = true
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	diskFull := fmt.Errorf("write %s: No space left on device",
		filepath.Join("x", "y"))
	h.runner.writeNote = func(path, content string) error {
		return diskFull
	}
	url := "https://github.com/o/r"
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if !h.logs.hasSub("DISK FULL!") ||
		!h.logs.hasSub("Headless mode: skipping this repo (disk full)") {
		t.Fatalf("disk-full logs missing:\n%v", h.logs.all())
	}
	failed := false
	for _, row := range h.cache.GetFailedURLs() {
		if strings.Contains(row.URL, "o/r") ||
			strings.Contains(fmt.Sprint(row.URL), "github.com/o/r") {
			failed = true
		}
	}
	if !failed {
		t.Fatalf("the disk-full repo must land in the retry queue")
	}
}

// TestLoopDiskFullGUIPausesAndRetries pins the GUI door: the pause
// until the resume, then the successful retry write.
func TestLoopDiskFullGUIPausesAndRetries(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	failures := 0
	h.runner.writeNote = func(path, content string) error {
		failures++
		if failures <= 1 {
			return fmt.Errorf("write %s: No space left on device", path)
		}
		return storage.AtomicWriteText(path, content)
	}
	// The GUI answers the disk-full signal by clearing the pause.
	diskFullEvents := 0
	h.runner.OnDiskFull = func(path string) {
		diskFullEvents++
		h.runner.ResumeAfterDiskFull()
	}
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if diskFullEvents != 1 {
		t.Fatalf("disk full signal count = %d", diskFullEvents)
	}
	if h.runner.Processed != 1 {
		t.Fatalf("the retried write must process the repo")
	}
}

// TestLoopStillFullAfterResumeSkips pins the still-full retry path.
func TestLoopStillFullAfterResumeSkips(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.writeNote = func(path, content string) error {
		return fmt.Errorf("write %s: No space left on device", path)
	}
	h.runner.OnDiskFull = func(path string) { h.runner.ResumeAfterDiskFull() }
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if !h.logs.hasSub("Disk still full — skipping") {
		t.Fatalf("still-full log missing")
	}
	if h.runner.Processed != 0 {
		t.Fatalf("the repo must stay unprocessed")
	}
}

// TestLoopDelays pins the inter-repo delay law: 0.5 s always, the
// extra 1.5 s only when total > 50.
func TestLoopDelays(t *testing.T) {
	rec := &sleepRecorder{}
	h := newHarness(t, nil)
	h.runner.Sleep = rec.sleep
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if len(rec.calls) == 0 || rec.calls[0] != 500*millis {
		t.Fatalf("the 0.5s delay missing: %v", rec.calls)
	}
	// total > 50 → the extra delay fires after each repo.
	rec2 := &sleepRecorder{}
	h2 := newHarness(t, nil)
	h2.runner.Sleep = rec2.sleep
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h2.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h2.runner.Total = 51
	h2.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if len(rec2.calls) != 2 || rec2.calls[0] != 500*millis ||
		rec2.calls[1] != 1500*millis {
		t.Fatalf("large-batch delays: %v", rec2.calls)
	}
}

const millis = time.Millisecond

// TestLoopProcessedLogAndBookkeeping pins the chain: the cache row, the
// log line with the vault-relative path, the note-state row, the
// resolved retry row, the vault index, the manifest, the processed
// count, and the _processed_log entry.
func TestLoopProcessedLogAndBookkeeping(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	// The banner lands.
	banner := &fakeBanner{result: filepath.Join(h.vault, "attachments",
		"banners", "r_banner.png")}
	os.MkdirAll(filepath.Dir(banner.result), 0o755)
	os.WriteFile(banner.result, []byte("png"), 0o644)
	h.runner.BannerFn = banner.fn()
	repo := newRepo("o/r", 10)
	repo.StargazersCount = 1200
	h.fetcher.repos["o/r"] = repo
	h.fetcher.topics["o/r"] = []string{"cli", "go"}
	h.fetcher.commits["o/r"] = 30
	h.fetcher.readmes["o/r"] = &ReadmeResult{
		Content: strings.Repeat("R", 3000), RawChars: 3000}
	h.fetcher.languages["o/r"] = []string{"Go", "Shell", "C", "D", "E",
		"F", "G"}
	h.fetcher.releases["o/r"] = &Release{TagName: "v1.0",
		PublishedAt: time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC)}
	url := "https://github.com/o/r"
	h.runner.LinkTracker.Intake([]string{url}, nil)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)

	if h.runner.Processed != 1 {
		t.Fatalf("processed = %d", h.runner.Processed)
	}
	if !h.logs.hasSub("✅ [vault] o/r processed and stored → ") {
		t.Fatalf("the stored log line missing:\n%v", h.logs.all())
	}
	if !h.logs.hasSub("📄 README fetched (3000 chars)") {
		t.Fatalf("README log missing")
	}
	if !h.logs.hasSub("📦 Latest release: v1.0 (2026-09-01)") {
		t.Fatalf("release log missing")
	}
	if !h.logs.hasSub("   ✅ Banner: r_banner.png") {
		t.Fatalf("banner log missing")
	}
	// The processed log entry.
	if len(h.runner.processedLog) != 1 {
		t.Fatalf("processed log: %+v", h.runner.processedLog)
	}
	pl := h.runner.processedLog[0]
	if pl.Repo != "o/r" || pl.Category != "Dev-Tools" || !pl.Banner {
		t.Fatalf("processed log entry: %+v", pl)
	}
	// The manifest row.
	if ml := h.runner.LinkTracker.Manifest.Links; len(ml) != 1 ||
		ml[0].Status != "processed" || ml[0].NotePath == nil {
		t.Fatalf("manifest: %+v", ml)
	}
	// The retry queue is empty (no failures + the row resolved).
	if h.cache.GetFailedCount() != 0 {
		t.Fatalf("failed rows: %d", h.cache.GetFailedCount())
	}
	// The note content: frontmatter + the release date + the tags cap.
	var note []byte
	_ = filepath.Walk(h.vault, func(p string, info os.FileInfo,
		err error) error {
		if err == nil && strings.HasSuffix(p, ".md") &&
			!strings.Contains(p, "_missing") {
			note, _ = os.ReadFile(p)
		}
		return nil
	})
	s := string(note)
	if !strings.Contains(s, "⭐ 1,200 ·") ||
		!strings.Contains(s, "- **Stars:** 1,200") {
		t.Fatalf("the comma-grouped stars lines missing:\n%s", s)
	}
	if !strings.Contains(s, "last_release: 2026-09-01") ||
		!strings.Contains(s, "- **Last Release:** 2026-09-01") {
		t.Fatalf("the release lines missing:\n%s", s)
	}
	// Tags: the LLM tag + topics + languages, capped at 12; the
	// primary language's lowercase inserts at the head (the merged
	// list carries the case-quirk dup, and the note's own sanitizer
	// dedupes on the way into the frontmatter — the reference's law).
	if !strings.Contains(s, "tags: [go, x, cli, shell, c, d, e]") {
		t.Fatalf("the tags line missing:\n%s", s)
	}
}

// TestLoopNonDiskFullWriteErrorFailsRepo pins the raise path: an
// OSError that is NOT disk-full fails the repo into the retry queue.
func TestLoopNonDiskFullWriteErrorFailsRepo(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.writeNote = func(path, content string) error {
		return fmt.Errorf("permission denied")
	}
	url := "https://github.com/o/r"
	h.runner.LinkTracker.Intake([]string{url}, nil)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	if !h.logs.hasSub("❌ Error processing") {
		t.Fatalf("the per-URL except log missing")
	}
	if h.cache.GetFailedCount() != 1 {
		t.Fatalf("the write failure must land in the retry queue")
	}
}

// TestLoopStopsBetweenRepos pins the Stop door.
func TestLoopStopsBetweenRepos(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/a"] = newRepo("o/a", 10)
	h.fetcher.repos["o/b"] = newRepo("o/b", 10)
	// The reference's shape: the is_running poll sits at the TOP of each
	// iteration — a Stop that lands during /a's status emit lets /a
	// finish (the checks inside the body are the waits) and breaks the
	// loop at /b's poll.
	h.runner.OnStatus = func(url string) {
		if strings.HasSuffix(url, "/a") {
			h.runner.Stop()
		}
	}
	h.runner.Total = 2
	h.runner.RunGitHubLoop([]string{"https://github.com/o/a",
		"https://github.com/o/b"}, nil, nil, nil, 3)
	if h.runner.Processed != 1 {
		t.Fatalf("the second repo must not process after Stop (got %d)",
			h.runner.Processed)
	}
}

// TestLoopBannerJoinWithMove pins the banner join + the Uncategorized
// → category move when the banner landed in the Uncategorized folder.
func TestLoopBannerJoinWithMove(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	// The banner seam returns a path INSIDE the Uncategorized folder —
	// the move branch then relocates it into Dev-Tools.
	uncatFolder := filepath.Join(h.vault, "Uncategorized")
	os.MkdirAll(uncatFolder, 0o755)
	bannerPath := filepath.Join(uncatFolder, "r_banner.png")
	os.WriteFile(bannerPath, []byte("png"), 0o644)
	banner := &fakeBanner{result: bannerPath}
	h.runner.BannerFn = banner.fn()
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	// The banner moved into the Dev-Tools folder.
	if _, err := os.Stat(filepath.Join(h.vault, "Tools", "Dev-Tools",
		"r_banner.png")); err != nil {
		t.Fatalf("the banner was not moved: %v", err)
	}
}

// TestLoopModelRereadPerURL pins the v30 persistence law: a config
// mutation (ApplyModelChoice) changes what the NEXT URL's analysis
// uses.
func TestLoopModelRereadPerURL(t *testing.T) {
	h := newHarness(t, map[string]any{"cloud_model": "first-model"})
	var seen []string
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	orig := callCloudLLM
	callCloudLLM = func(apiURL, apiKey, model string,
		messages []llmclient.Message, timeoutS float64,
		opts *llmclient.ChatOptions) (string, error) {
		seen = append(seen, model)
		return okResult, nil
	}
	t.Cleanup(func() { callCloudLLM = orig })
	h.fetcher.repos["o/a"] = newRepo("o/a", 10)
	h.fetcher.repos["o/b"] = newRepo("o/b", 10)
	h.runner.Total = 2
	h.runner.RunGitHubLoop([]string{"https://github.com/o/a",
		"https://github.com/o/b"}, nil, nil, nil, 3)
	// ApplyModelChoice after the first repo → the second uses it.
	h.runner.ApplyModelChoice("second-model")
	h.runner.Total = 3
	h.fetcher.repos["o/c"] = newRepo("o/c", 10)
	h.runner.RunGitHubLoop([]string{"https://github.com/o/c"}, nil,
		nil, nil, 3)
	if len(seen) != 3 || seen[0] != "first-model" ||
		seen[1] != "first-model" || seen[2] != "second-model" {
		t.Fatalf("models: %v", seen)
	}
}

// ---------------------------------------------------------------------------
// The rate-limit pre-flight laws (961-994).
// ---------------------------------------------------------------------------

func TestRateLimitPreflightPauseUntilReset(t *testing.T) {
	h := newHarness(t, nil)
	reset := h.clock.Now().Add(300 * time.Second)
	h.fetcher.rateLimit = &RateLimit{Remaining: 49, Reset: reset}
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if !h.logs.hasSub("GitHub rate limit reached. Pausing until") {
		t.Fatalf("pause log missing:\n%v", h.logs.all())
	}
	if !h.logs.hasSub("(5.0 minutes, 49 remaining)") {
		t.Fatalf("the wait minutes line missing")
	}
	if !h.logs.hasSub("✅ Rate limit wait complete — resuming.") {
		t.Fatalf("resume log missing")
	}
	if h.runner.Processed != 1 {
		t.Fatalf("the repo must process after the pause")
	}
}

func TestRateLimitPreflightContinueAnyway(t *testing.T) {
	h := newHarness(t, nil)
	reset := h.clock.Now().Add(4000 * time.Second)
	h.fetcher.rateLimit = &RateLimit{Remaining: 10, Reset: reset}
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if !h.logs.hasSub("GitHub rate limit low (10 remaining)") {
		t.Fatalf("the continue-anyway warning missing")
	}
	if !h.logs.hasSub("Continuing anyway...") {
		t.Fatalf("the ellipsis line missing")
	}
	// No pause happened: the clock did not move 4000s.
	if h.clock.Now().Sub(time.Date(2026, 10, 17, 12, 0, 0, 0, time.UTC)) > time.Hour {
		t.Fatalf("the continue-anyway branch must not pause")
	}
}

func TestRateLimitPreflightMultipleOfFifty(t *testing.T) {
	h := newHarness(t, nil)
	h.fetcher.rateLimit = &RateLimit{Remaining: 100,
		Reset: h.clock.Now().Add(time.Hour)}
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if !h.logs.hasSub("📊 GitHub API: 100 requests remaining") {
		t.Fatalf("the modulo-50 info line missing")
	}
}

func TestRateLimitPreflightErrorIsSilent(t *testing.T) {
	h := newHarness(t, nil)
	h.fetcher.rateErr = fmt.Errorf("boom")
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if h.logs.hasSub("rate limit") {
		t.Fatalf("a rate-limit probe error must be silent")
	}
	if h.runner.Processed != 1 {
		t.Fatalf("the batch must continue")
	}
}

func TestRateLimitPreflightStopDuringWait(t *testing.T) {
	h := newHarness(t, nil)
	h.fetcher.rateLimit = &RateLimit{Remaining: 1,
		Reset: h.clock.Now().Add(300 * time.Second)}
	// The sleep seam stops the batch mid-wait.
	h.runner.Sleep = func(d time.Duration) {
		h.clock.sleep(d)
		h.runner.Stop()
	}
	h.runner.Total = 1
	h.runner.RunGitHubLoop([]string{"https://github.com/o/r"}, nil,
		nil, nil, 3)
	if h.runner.Processed != 0 {
		t.Fatalf("the repo must not process when stopped mid-wait")
	}
}

// ---------------------------------------------------------------------------
// The undo machinery.
// ---------------------------------------------------------------------------

func TestSnapshotAndUndoDiff(t *testing.T) {
	dir := t.TempDir()
	vault := filepath.Join(dir, "vault")
	mk := func(rel string) string {
		p := filepath.Join(vault, rel)
		os.MkdirAll(filepath.Dir(p), 0o755)
		os.WriteFile(p, []byte("x"), 0o644)
		return p
	}
	mk("old.md")
	mk("Sub/old2.md")
	mk(".obsidian/config.md")
	mk("attachments/banner/x.md")

	snap := SnapshotBatchFiles(vault)
	if len(snap) != 2 {
		t.Fatalf("snapshot = %d, want 2 (.obsidian/attachments skipped)",
			len(snap))
	}

	// The batch writes new notes; .trash must not count as new.
	mk("New.md")
	mk("Sub/New2.md")
	mk(".trash/retired.md")
	mk("attachments/banner/new.md")

	newFiles := ComputeNewFiles(vault, snap)
	if len(newFiles) != 2 {
		t.Fatalf("new files = %v", newFiles)
	}
	// The walk order: files before subdirs at every level (old.md and
	// New.md before Sub/New2.md).
	if len(newFiles) == 2 && newFiles[0] != filepath.Join(vault, "New.md") {
		t.Fatalf("order: %v", newFiles)
	}

	// The empty-snapshot quirk: no snapshot → no diff at all.
	mk("More.md")
	if got := ComputeNewFiles(vault, nil); len(got) != 0 {
		t.Fatalf("an empty snapshot must yield no diff, got %v", got)
	}

	// The substring law: a vault path containing a skip name skips
	// whole.
	weird := filepath.Join(dir, "attachments-demo", "v")
	os.MkdirAll(weird, 0o755)
	os.WriteFile(filepath.Join(weird, "n.md"), []byte("x"), 0o644)
	if got := SnapshotBatchFiles(weird); len(got) != 0 {
		t.Fatalf("the substring law must skip the whole vault, got %v", got)
	}
}

func TestWriteUndoFileBytes(t *testing.T) {
	h := newHarness(t, nil)
	files := []string{filepath.Join(h.vault, "A.md"),
		filepath.Join(h.vault, "B.md")}
	if !h.runner.WriteUndoFile(h.vault, files) {
		t.Fatal("write failed")
	}
	data, err := os.ReadFile(filepath.Join(h.vault, "_undo_last_batch.txt"))
	if err != nil {
		t.Fatal(err)
	}
	want := files[0] + "\n" + files[1] + "\n"
	if string(data) != want {
		t.Fatalf("undo bytes: %q want %q", data, want)
	}
	// The empty case writes an empty file.
	if !h.runner.WriteUndoFile(h.vault, nil) {
		t.Fatal("write failed")
	}
	data, _ = os.ReadFile(filepath.Join(h.vault, "_undo_last_batch.txt"))
	if len(data) != 0 {
		t.Fatalf("empty undo: %q", data)
	}
}

func TestFinishChainSummaryAndLines(t *testing.T) {
	h := newHarness(t, nil)
	fakeLLM(t, func(m []llmclient.Message, jm bool) (string, error) {
		return okResult, nil
	})
	h.runner.Config["reports_dir"] = filepath.Join(h.dir, "reports")
	h.fetcher.repos["o/r"] = newRepo("o/r", 10)
	url := "https://github.com/o/r"
	h.runner.LinkTracker.Intake([]string{url}, nil)
	h.runner.Total = 1
	// A pre-existing note (the realistic vault; also dodges the
	// reference's empty-snapshot falsiness quirk, which records NO
	// undo lines for a batch on an empty vault).
	os.WriteFile(filepath.Join(h.vault, "Preexisting.md"),
		[]byte("# old"), 0o644)
	h.runner.batchFiles = SnapshotBatchFiles(h.vault)
	h.runner.RunGitHubLoop([]string{url}, nil, nil, nil, 3)
	finished := ""
	h.runner.Finished = func(ok bool, msg string) { finished = msg }
	summary := h.runner.Finish()
	if !strings.Contains(finished, "Processed 1 out of 1 repos.") {
		t.Fatalf("finished message: %q", finished)
	}
	if summary == nil || summary.Stopped || summary.GithubProcessed != 1 ||
		summary.GithubTotal != 1 || summary.NewNotes != 1 ||
		summary.FailedLinks != 0 {
		t.Fatalf("summary: %+v", summary)
	}
	if summary.SummaryPath == "" || summary.ReportPath == "" {
		t.Fatalf("the report paths must be captured: %+v", summary)
	}
	if !h.logs.hasSub("↩️ Batch undo saved: 1 new files") {
		t.Fatalf("undo log missing")
	}
	if !h.logs.hasSub("🏁 Done. Processed 1 repos.") {
		t.Fatalf("done log missing")
	}
	// The summary log + final report files exist.
	for _, p := range []string{summary.SummaryPath, summary.ReportPath} {
		if fi, err := os.Stat(p); err != nil || fi.Size() == 0 {
			t.Fatalf("report file %s: %v", p, err)
		}
	}
	// The undo file lists the new note.
	data, _ := os.ReadFile(filepath.Join(h.vault, "_undo_last_batch.txt"))
	if !strings.Contains(string(data), ".md\n") {
		t.Fatalf("undo list: %q", data)
	}
}

func TestFinishChainAggregateLines(t *testing.T) {
	h := newHarness(t, nil)
	h.runner.Config["reports_dir"] = filepath.Join(h.dir, "reports")
	h.runner.DeadSkipped = 2
	h.runner.SettledReposSkipped = 1
	h.runner.DismissedSkipped = 1
	h.runner.batchFiles = map[string]bool{}
	summary := h.runner.Finish()
	if summary == nil {
		t.Fatal("summary nil")
	}
	if !h.logs.hasSub("2 dead link(s) skipped — 404 quarantine") {
		t.Fatalf("dead aggregate line missing")
	}
	if !h.logs.hasSub("1 settled repo link(s) skipped") {
		t.Fatalf("settled aggregate line missing")
	}
	if !h.logs.hasSub("1 dismissed link(s) skipped") {
		t.Fatalf("dismissed aggregate line missing")
	}
}

func TestFinishChainWebsitesTally(t *testing.T) {
	h := newHarness(t, nil)
	h.runner.Config["reports_dir"] = filepath.Join(h.dir, "reports")
	h.runner.websites = &WebsiteSummary{
		Counters: map[string]int{
			"processed": 7, "review": 2, "skipped": 1, "failed": 0},
		Vault: "/wv"}
	h.runner.batchFiles = map[string]bool{}
	var msg string
	h.runner.Finished = func(ok bool, m string) { msg = m }
	summary := h.runner.Finish()
	if !strings.Contains(msg, "Websites: 7 processed, 2 to review, 1 skipped.") {
		t.Fatalf("websites tally missing: %q", msg)
	}
	if summary.Websites == nil || summary.Websites.Processed != 7 {
		t.Fatalf("websites summary: %+v", summary.Websites)
	}
}
