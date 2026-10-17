// runner.go — T-42: the GitHub per-repo loop
// (gui/processing_worker.py:947-1725) — the gates in their exact
// order, the get_repo error paths (403+rate wait / 401 anonymous
// fallback / 404 strike / other-failed), the two-layer dedupe, the
// fork→parent analysis, the enrichment (topics, 90-day commits,
// README ≤1500, languages, latest release, credibility), the banner
// goroutine (joined 15 s), the model re-read per URL, the note write
// with the disk-full protocol, the bookkeeping chain (cache → log →
// note_state → mark_failed_resolved → vault_index → manifest), and
// the delays (0.5 s + 1.5 s when total > 50).
//
// The Runner also owns the post-loop completion chain's GitHub pieces
// (reports.go / undo.go) and carries the seam surface the GUI shell
// (a later task) wires: events (log/status/progress/disk-full/model
// changed), the ask-gates (LLM decision, disk-full resume), the
// websites phase, and the hygiene pass.
package githubpipeline

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notebuilder"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// WebsiteResultEntry is one _website_summary['results'] row.
type WebsiteResultEntry struct {
	URL         string
	Outcome     string
	Category    string
	Subcategory string
	Error       string
}

// WebsiteSummary mirrors _website_summary: the counters dict, the
// per-link results, and the vault the phase worked in.
type WebsiteSummary struct {
	Counters map[string]int
	Results  []WebsiteResultEntry
	Vault    string
}

// ProcessedLogEntry mirrors the _processed_log row dict.
type ProcessedLogEntry struct {
	URL         string
	Repo        string
	Category    string
	NotePath    string
	Banner      bool
	Credibility float64
}

// Runner mirrors the ProcessingWorker's GitHub-side state + loop. The
// GUI shell builds one per batch; the tests drive it directly.
type Runner struct {
	// Config is the live config map (mutated IN PLACE by
	// ApplyModelChoice — the v30 persistence law).
	Config map[string]any
	// Headless mirrors _headless: no GUI to answer the doors — every
	// wait takes the fail-fast default instead of stalling.
	Headless bool

	// The pre-loop state the shell builds (the reference's _run_impl
	// setup): the cache, the note-state connection, the vault index,
	// the manifest tracker, and the run's note-state report.
	Cache       *store.CacheDB
	NoteStateDB *notestate.DB
	VaultIndex  *batch.VaultIndex
	LinkTracker *batch.LinkTracker

	// The GitHub seam: NewGHClient builds the fetcher for a token
	// ("" = anonymous). The loop swaps to the anonymous client on the
	// 401 drop.
	NewGHClient func(token string) RepoFetcher
	gh          RepoFetcher
	githubToken string

	// The LLM collaborators.
	LLMProvider  string
	OllamaClient llmclient.OllamaAPI
	Model        string

	// The websites-phase seam (the shell wires _run_website_phase; nil
	// mirrors the pipeline-off shape — the loop skips straight to the
	// completion chain).
	WebsitePhase func() *WebsiteSummary

	// The ask-gates.
	LLMDecisionFn func(repoName string, attempt int, err error,
		client llmclient.OllamaAPI, model string) string
	ModelPromptFn func(configured string, available []string) string

	// BannerFn is the banner-download seam (nil → the real
	// DownloadBanner; the tests inject the "fake banner").
	BannerFn func(owner, repoName, folder string) string

	// The events.
	Log            func(msg, level string)
	OnStatus       func(url string)
	OnProgress     func(pos, total int)
	OnDiskFull     func(path string)
	OnModelChanged func(provider, model string)
	Finished       func(ok bool, msg string)

	// The hygiene pass seam (core/vault_hygiene is a later task's
	// port; nil = skipped, the best-effort shape).
	HygienePass func()

	// The timing seams (the rate-limit law tests freeze them).
	Clock func() time.Time
	Sleep func(time.Duration)

	// writeNote is the atomic-write seam (the reference calls
	// _storage.atomic_write_text; the disk-full protocol's tests inject
	// failures through here — the same injection point the reference's
	// own suite uses via mock.patch).
	writeNote func(path, content string) error

	// The per-URL counters the aggregate lines speak from.
	DeadSkipped         int
	DismissedSkipped    int
	SettledReposSkipped int

	// The intake facts the final report carries.
	NonGithubURLs    []string
	IntakeDuplicates int
	RawURLCount      int
	// BotSource mirrors _bot_source (the manifest + report's source
	// label).
	BotSource bool
	// NoteStateRun mirrors _note_state_run (the run_start_check results
	// the final report's note-state section renders).
	NoteStateRun map[string]*notestate.RunResult

	// Totals.
	Total     int
	Processed int

	// The per-batch state.
	running             bool
	tokenDropped        bool
	llmHeadlessDeclined bool
	bannerCount         int
	position            int
	missingNotesWritten map[string]bool
	processedLog        []ProcessedLogEntry
	diskFullPaused      atomic.Bool
	batchFiles          map[string]bool
	websites            *WebsiteSummary
}

// NewRunner builds the runner with the default seams (the real GitHub
// client factory, the real clock and sleep) and no-op events.
func NewRunner(config map[string]any) *Runner {
	return &Runner{
		Config:              config,
		NewGHClient:         func(token string) RepoFetcher { return NewClient(token) },
		missingNotesWritten: map[string]bool{},
		Clock:               time.Now,
		Sleep:               time.Sleep,
		writeNote:           storage.AtomicWriteText,
		Log:                 func(msg, level string) {},
		OnStatus:            func(url string) {},
		OnProgress:          func(pos, total int) {},
		OnDiskFull:          func(path string) {},
		OnModelChanged:      func(provider, model string) {},
	}
}

// Stop mirrors worker.stop(): the loop breaks at the next poll.
func (r *Runner) Stop() { r.running = false }

// RunningNow reports the is_running flag (atomic reads are fine; the
// loop is single-goroutine and the Stop door is the only writer).
func (r *Runner) RunningNow() bool { return r.running }

// ResumeAfterDiskFull mirrors the GUI's disk-full resume.
func (r *Runner) ResumeAfterDiskFull() { r.diskFullPaused.Store(false) }

func (r *Runner) log(msg, level string) {
	if r.Log != nil {
		func() {
			defer func() { recover() }()
			r.Log(msg, level)
		}()
	}
}

func (r *Runner) emitStatus(url string) {
	if r.OnStatus != nil {
		func() {
			defer func() { recover() }()
			r.OnStatus(url)
		}()
	}
}

func (r *Runner) emitProgress() {
	if r.OnProgress != nil {
		func() {
			defer func() { recover() }()
			r.OnProgress(r.position, r.Total)
		}()
	}
}

func (r *Runner) clockNow() time.Time {
	if r.Clock == nil {
		return time.Now()
	}
	return r.Clock()
}

func (r *Runner) sleep(d time.Duration) {
	if r.Sleep == nil {
		time.Sleep(d)
		return
	}
	r.Sleep(d)
}

// RunGitHubLoop is the per-repo loop (947-1725). deadURLs/dismissed/
// settled are the pre-loop sets the shell loaded; deadThreshold is
// the configured strike threshold.
func (r *Runner) RunGitHubLoop(urls []string, deadURLs, dismissed,
	settled map[string]bool, deadThreshold int) {

	r.githubToken = cfgStr(r.Config, "github_token")
	r.gh = r.NewGHClient(r.githubToken)
	if r.missingNotesWritten == nil {
		r.missingNotesWritten = map[string]bool{}
	}
	dismissedSkipped := 0

loop:
	for _, rawURL := range urls {
		if !r.running {
			break
		}
		url := rawURL
		r.position++ // increment for EVERY URL (including skips)
		r.emitStatus(url)
		r.emitProgress()

		// v23 — Phase 2: mark as processing (manifest is source of truth).
		if r.LinkTracker != nil {
			func() {
				defer func() { recover() }()
				r.LinkTracker.MarkProcessing(url)
			}()
		}

		if err := r.processURL(url, deadURLs, dismissed, settled,
			deadThreshold, &dismissedSkipped); err != nil {
			if err == errLoopBreak {
				break loop
			}
			// The per-URL except (1699-1717): the failure is recorded
			// for retry, the manifest marked failed, the bar repaints.
			r.log(fmt.Sprintf("❌ Error processing %s: %s", url, err),
				"error")
			func() {
				defer func() { recover() }()
				r.Cache.AddFailed(url, err.Error())
			}()
			if r.LinkTracker != nil {
				func() {
					defer func() { recover() }()
					r.LinkTracker.MarkFailed(url, err.Error())
				}()
			}
			r.emitProgress()
			continue
		}

		r.sleep(secs(pyFloatOr(r.Config, "delay_between_api_calls", 0.5)))
		// v25 pre-flight: large batches (>50 links) add an extra delay
		// so the GitHub rate limit is never hit mid-batch.
		if r.Total > 50 {
			r.sleep(secs(pyFloatOr(r.Config, "large_batch_extra_delay", 1.5)))
		}
	}
}

// errLoopBreak signals a stop that escaped a gate (the rate-limit
// waits and the disk-full pause break the loop after their wait).
var errLoopBreak = fmt.Errorf("loop break")

// processURL runs one URL through the full gate chain; an error is
// the per-URL except's payload (record + continue), errLoopBreak ends
// the loop.
func (r *Runner) processURL(url string, deadURLs, dismissed,
	settled map[string]bool, deadThreshold int,
	dismissedSkipped *int) error {

	// Check the GitHub rate limit (v25 pre-flight).
	if err := r.rateLimitPreflight(); err != nil {
		return err
	}

	result, err := r.processURLInner(url, deadURLs, dismissed, settled,
		deadThreshold, dismissedSkipped)
	if err != nil {
		return err
	}
	// A nil result with a nil error is the "already handled + continue"
	// shape (the skip gates); a non-nil result is the processed note.
	if result != nil {
		r.Processed++
	}
	return nil
}

// rateLimitPreflight mirrors 961-994: pause when <50 remain until the
// exact reset time (chunked 5 s sleeps), continue-anyway when the
// wait is out of the window, the %50 info line, all best-effort.
func (r *Runner) rateLimitPreflight() error {
	defer func() { recover() }() // rate limit check is best-effort
	rl, err := r.gh.GetRateLimit()
	if err != nil {
		return nil
	}
	remaining := rl.Remaining
	if remaining < 50 {
		resetLocal := rl.Reset.In(time.Local)
		waitSeconds := rl.Reset.Sub(r.clockNow().UTC()).Seconds()
		if waitSeconds > 0 && waitSeconds < 3700 { // less than ~1 hour
			r.log(fmt.Sprintf(
				"⏳ GitHub rate limit reached. Pausing until %s (%s minutes, "+
					"%d remaining)...",
				resetLocal.Format("15:04"),
				fmt.Sprintf("%.1f", waitSeconds/60), remaining), "warning")
			// Sleep in small chunks so Stop is responsive.
			end := r.clockNow().Add(time.Duration(waitSeconds*float64(time.Second)) +
				5*time.Second)
			for r.clockNow().Before(end) && r.running {
				rem := end.Sub(r.clockNow())
				if rem > 5*time.Second {
					rem = 5 * time.Second
				}
				r.sleep(rem)
			}
			if !r.running {
				return errLoopBreak
			}
			r.log("✅ Rate limit wait complete — resuming.", "success")
		} else {
			r.log(fmt.Sprintf(
				"⚠️ GitHub rate limit low (%d remaining). Reset at %s. "+
					"Continuing anyway...", remaining,
				resetLocal.Format("15:04")), "warning")
		}
	} else if remaining%50 == 0 {
		r.log(fmt.Sprintf("📊 GitHub API: %d requests remaining", remaining),
			"info")
	}
	return nil
}

// processURLInner is the loop body inside the reference's per-URL try.
func (r *Runner) processURLInner(url string, deadURLs, dismissed,
	settled map[string]bool, deadThreshold int,
	dismissedSkipped *int) (*ProcessedLogEntry, error) {

	url = links.CleanURL(url)
	// v0.29.0 — canonicalize through the single link grammar so http://,
	// www. and decorated lines all resolve to the same canonical repo URL.
	if canon := links.ExtractGitHubURLs(url); len(canon) > 0 {
		url = canon[0]
	}
	if !strings.HasPrefix(url, "https://github.com/") {
		r.log(fmt.Sprintf("Skipping non-GitHub URL: %s", url), "warning")
		r.markSkipped(url, "non-GitHub URL")
		r.markFailedResolved(url)
		r.emitProgress()
		return nil, nil
	}

	// v0.08 — 404 QUARANTINE skip: confirmed-dead links never reach the
	// GitHub API. Counted silently; ONE aggregate line at batch end.
	if deadURLs[links.NormalizeURL(url)] {
		r.DeadSkipped++
		r.markSkipped(url, "404 quarantine (confirmed dead)")
		r.markFailedResolved(url)
		r.emitProgress()
		return nil, nil
	}

	// v0.12.0 — Phase 3 (§4.4 "deleted"): a URL whose note the owner
	// deleted is NEVER re-added — BEFORE any GitHub API call.
	if len(dismissed) > 0 && dismissed[links.NormalizeURL(url)] {
		r.log(fmt.Sprintf("🚫 Dismissed (you deleted its note): %s", url), "info")
		*dismissedSkipped++
		r.DismissedSkipped = *dismissedSkipped
		r.markSkipped(url, "dismissed (note deleted)")
		r.markFailedResolved(url)
		r.emitProgress()
		return nil, nil
	}

	// v0.63.3 — THE SETTLED REPOS LEDGER's skip: a settled repo is never
	// fetched, never counted as this batch's work.
	if len(settled) > 0 && settled[links.NormalizeURL(url)] {
		r.SettledReposSkipped++
		r.markSkipped(url, "settled — addressed (the owner's law)")
		r.markFailedResolved(url)
		r.emitProgress()
		return nil, nil
	}

	parts := strings.Split(strings.TrimPrefix(url, "https://github.com/"), "/")
	if len(parts) < 2 {
		r.log(fmt.Sprintf("Invalid GitHub URL: %s", url), "warning")
		r.markSkipped(url, "invalid GitHub URL")
		r.markFailedResolved(url)
		r.emitProgress()
		return nil, nil
	}
	owner, repoName := parts[0], parts[1]

	// v0.38.0 — VAULT DEDUP BEFORE ANY GITHUB API CALL: the vault is
	// the ground truth — note present → skip, no API, no AI.
	if r.VaultIndex != nil && r.VaultIndex.HasURL(url) {
		notePath := r.VaultIndex.GetPath(url)
		r.log(fmt.Sprintf("⏭️ Already in vault: %s/%s → %s", owner, repoName,
			filepath.Base(notePath)), "info")
		r.markSkipped(url, "already in vault")
		r.markFailedResolved(url)
		r.emitProgress()
		return nil, nil
	}

	repo, err := r.gh.GetRepo(owner + "/" + repoName)
	if err != nil {
		return r.getRepoErrorPath(url, owner, repoName, err, deadThreshold)
	}

	// v0.09 (merge fix) — 404-quarantine RESET on success: the repo
	// EXISTS again, so its attempt counter starts fresh. All get_repo
	// success paths (direct, post-rate-limit retry, anonymous post-401
	// retry) converge here.
	func() {
		defer func() { recover() }()
		r.Cache.ResetDeadLinks(url)
	}()

	// === DEDUP CHECK, layer 2: the SQLite cache ===
	if dup, _ := func() (bool, error) {
		defer func() { recover() }()
		return r.Cache.IsDuplicate(repo.ID)
	}(); dup {
		if valid, _ := func() (bool, error) {
			defer func() { recover() }()
			return r.Cache.IsNoteValid(repo.ID)
		}(); valid {
			notePath := ""
			if p, err := r.Cache.GetNotePath(repo.ID); err == nil && p != nil {
				notePath = *p
			}
			r.log(fmt.Sprintf("⏭️ Already processed (note exists): %s -> %s",
				url, notePath), "info")
			// Also add to vault index for future runs.
			if r.VaultIndex != nil {
				r.VaultIndex.AddURL(url, notePath)
			}
			r.markSkipped(url, "already in cache")
			r.markFailedResolved(url)
			r.emitProgress()
			return nil, nil
		}
		r.log(fmt.Sprintf("♻️ Note file was deleted, re-processing: %s", url),
			"info")
		func() {
			defer func() { recover() }()
			r.Cache.RemoveEntry(repo.ID)
		}()
		// Fall through to re-process.
	}

	return r.analyzeAndStore(url, owner, repoName, repo, "")
}

// getRepoErrorPath mirrors the GithubException branches (1165-1305).
func (r *Runner) getRepoErrorPath(url, owner, repoName string,
	err error, deadThreshold int) (*ProcessedLogEntry, error) {

	ghErr, ok := err.(*GHError)
	if !ok {
		// A transport failure is the reference's generic error path.
		r.log(fmt.Sprintf("GitHub API error: %s", err), "error")
		r.markFailed(url, fmt.Sprintf("GitHub API: %s", err))
		r.emitProgress()
		return nil, nil
	}
	switch {
	case ghErr.Status == 403 && strings.Contains(
		strings.ToLower(ghErr.Message), "rate limit"):
		// v25 pre-flight: wait the full hour, then retry THIS repo.
		r.log("⏳ GitHub rate limit reached. Waiting 3600s (1 hour) for reset...",
			"warning")
		rlEnd := r.clockNow().Add(3605 * time.Second)
		for r.clockNow().Before(rlEnd) && r.running {
			rem := rlEnd.Sub(r.clockNow())
			if rem > 5*time.Second {
				rem = 5 * time.Second
			}
			r.sleep(rem)
		}
		if !r.running {
			return nil, errLoopBreak
		}
		repo, err2 := r.gh.GetRepo(owner + "/" + repoName)
		if err2 == nil {
			r.log("✅ Rate limit wait complete — resuming with this repo.",
				"success")
			func() {
				defer func() { recover() }()
				r.Cache.ResetDeadLinks(url)
			}()
			return r.analyzeAndStore(url, owner, repoName, repo, "")
		}
		if gh2, ok2 := err2.(*GHError); ok2 && gh2.Status == 404 {
			r.record404Path(url, owner, repoName, deadThreshold)
		} else {
			r.log(fmt.Sprintf("GitHub API error after rate-limit wait: %s",
				err2), "error")
			r.markFailed(url, fmt.Sprintf("GitHub API: %s", err2))
		}
		r.emitProgress()
		return nil, nil

	case ghErr.Status == 401 && r.githubToken != "":
		// v32.1 — the saved GitHub token was rejected: drop it for the
		// rest of the batch (anonymous access, 60 req/h), retry THIS
		// repo, tell the user how to fix it. Logged once.
		if !r.tokenDropped {
			r.tokenDropped = true
			r.log("🔑 GitHub token rejected (401 Bad credentials) — it is "+
				"invalid, expired, or was rotated.", "error")
			r.log("   Continuing this batch anonymously (60 requests/hour). "+
				"Fix: Settings → Credentials → paste a fresh token "+
				"(github.com/settings/tokens) → 'Test GitHub Token'.",
				"warning")
			r.githubToken = "" // never re-enter this branch
		}
		r.gh = r.NewGHClient("") // anonymous client from here on
		repo, err2 := r.gh.GetRepo(owner + "/" + repoName)
		if err2 != nil {
			r.log(fmt.Sprintf("GitHub API error: %s", err2), "error")
			r.markFailed(url, fmt.Sprintf("GitHub API: %s", err2))
			r.emitProgress()
			return nil, nil
		}
		func() {
			defer func() { recover() }()
			r.Cache.ResetDeadLinks(url)
		}()
		return r.analyzeAndStore(url, owner, repoName, repo, "")

	case ghErr.Status == 404:
		// v0.20.0 — a GitHub 404 is definitive: the strike counts, the
		// _missing placeholder note is written, the quarantine row is
		// confirmed on the spot, and the permanent record lands in
		// _inbox/notfound-links/ (once).
		r.record404Path(url, owner, repoName, deadThreshold)
		r.emitProgress()
		return nil, nil

	default:
		r.log(fmt.Sprintf("GitHub API error: %s", err), "error")
		r.markFailed(url, fmt.Sprintf("GitHub API: %s", err))
		r.emitProgress()
		return nil, nil
	}
}

// record404Path is the 404 branch's shared body (the direct hit and
// the post-rate-limit retry): the strike, the _missing note, the
// notfound-links.md append, the manifest skip.
func (r *Runner) record404Path(url, owner, repoName string,
	deadThreshold int) {
	strikes := 0
	func() {
		defer func() { recover() }()
		strikes = r.Cache.Record404(url, "404 Not Found")
	}()
	notePath := r.RecordMissingRepo(url, owner, repoName, strikes,
		r.Cache, deadThreshold)
	suffix := ""
	if notePath != "" {
		suffix = " → " + filepath.Base(notePath)
	}
	r.log("🗑️ Repo not found (404) — missing-repo note written, never "+
		"counted again: "+url+suffix, "warning")
	// Write to _inbox/notfound-links/ ONCE.
	func() {
		defer func() { recover() }()
		vaultPath := cfgStr(r.Config, "vault_path")
		if vaultPath == "" {
			return
		}
		nfFolder := filepath.Join(vaultPath, "_inbox", "notfound-links")
		if err := dryrun.MkdirAll(nfFolder, true); err != nil {
			return
		}
		nfPath := filepath.Join(nfFolder, "notfound_links.md")
		line := "| " + r.clockNow().Format("2006-01-02") + " | " + url +
			" | 404 Not Found (missing-repo note) |\n"
		_ = dryrun.AppendText(nfPath, line)
	}()
	r.markSkipped(url, "404 Not Found — missing-repo note")
}

// analyzeAndStore is the success path: fork→parent, the enrichment,
// the banner goroutine, the LLM analysis, the note build, the atomic
// write with the disk-full protocol, and the bookkeeping chain.
func (r *Runner) analyzeAndStore(url, owner, repoName string,
	repo *Repo, parentInfo string) (*ProcessedLogEntry, error) {

	// Check if this is a fork — analyze the parent instead.
	func() {
		defer func() { recover() }() // a broken parent access keeps the fork
		if repo.Fork && repo.Parent != nil {
			parent := repo.Parent
			parentInfo = "This is a fork of " + parent.FullName + ". "
			r.log(fmt.Sprintf("🍴 Fork detected — analyzing parent: %s",
				parent.FullName), "info")
			// Use the parent repo for analysis (the owner re-read below
			// follows the rebound repo, like the reference's 1393).
			repo = parent
			repoName = parent.Name
		}
	}()

	description := repo.Description
	if parentInfo != "" {
		description = repo.Description + "\n\n" + parentInfo
	}
	var topics []string
	if t, err := r.gh.GetTopics(repo.FullName); err == nil {
		topics = t
	} else {
		return nil, err // the reference lets get_topics() failures raise
	}
	ownerLogin := repo.OwnerLogin
	orgName := ownerLogin
	if repo.OrganizationLogin != "" {
		orgName = repo.OrganizationLogin
	}
	orgRep := OrgReputation(orgName)

	commitCount := 0
	if n, err := r.gh.CountCommitsSince(repo.FullName,
		r.clockNow().AddDate(0, 0, -90)); err == nil {
		commitCount = n
	} else {
		r.log(fmt.Sprintf("   ⚠️ Could not fetch commits: %s", err), "warning")
	}

	// Fetch README content (first 1500 chars) for the LLM context.
	readmeContent := ""
	readmeFetched := false
	if rd, err := r.gh.GetReadme(repo.FullName); err == nil && rd != nil {
		readmeFetched = true
		readmeContent = truncateRunes(rd.Content, 1500)
		r.log(fmt.Sprintf("   📄 README fetched (%d chars)", rd.RawChars),
			"info")
	} else {
		r.log("   ⚠️ No README found", "warning")
	}

	// Start the banner download in parallel (the daemon thread twin)
	// while the LLM analyzes.
	uncategorizedFolder := ""
	vaultPathForBanner := cfgStr(r.Config, "vault_path")
	if vaultPathForBanner != "" {
		folder, _ := constants.CategoryFolder("Uncategorized")
		if folder == "" {
			folder = "Uncategorized"
		}
		uncategorizedFolder = filepath.Join(vaultPathForBanner, folder)
		_ = dryrun.MkdirAll(uncategorizedFolder, true)
	}
	bannerFn := r.DownloadBanner
	if r.BannerFn != nil {
		bannerFn = func(owner, repoName, folder string) string {
			defer func() { recover() }()
			return r.BannerFn(owner, repoName, folder)
		}
	}
	bannerCh := make(chan string, 1)
	go func() {
		defer func() { recover() }()
		if vaultPathForBanner == "" {
			bannerCh <- ""
			return
		}
		bannerCh <- bannerFn(ownerLogin, repoName, uncategorizedFolder)
	}()

	// v30 — re-read the model from the config on EVERY url.
	r.Model = r.rereadModel()

	// LLM analysis with README + about_me context.
	llmResult, err := r.LLMAnalyze(r.Model, repoName, description, topics,
		ownerLogin, repo.StargazersCount, repo.ForksCount, readmeContent)
	if err != nil {
		return nil, err
	}
	summary := mapStrDef(llmResult, "summary", "No summary available.")
	howItWorks := mapStrDef(llmResult, "how_it_works", "No explanation provided.")
	coreValue := mapStrDef(llmResult, "core_value", "No core value provided.")
	features := mapDef(llmResult, "features", []any{"Feature 1", "Feature 2"})
	difference := mapStrDef(llmResult, "difference", "No comparison provided.")
	categoryGuess := mapStrDef(llmResult, "category", "Uncategorized")
	confidence, confErr := mapFloatDef(llmResult, "confidence", 50)
	if confErr != nil {
		return nil, confErr // float() of garbage → the per-URL except
	}
	tags := mapList(llmResult, "tags")

	// Lower threshold to 50% — always pick a category, just flag
	// low-confidence.
	categoryKey := "Uncategorized"
	if confidence >= 50 {
		if inCategoryKeys(categoryGuess) {
			categoryKey = categoryGuess
		}
	}

	// Merge GitHub topics with LLM tags (topics are authoritative).
	allTags := append([]any{}, tags...)
	for _, t := range topics {
		if t == "" {
			continue
		}
		found := false
		for _, a := range allTags {
			if helpers.PyStr(a) == t {
				found = true
				break
			}
		}
		if !found {
			allTags = append(allTags, t)
		}
	}

	// Fetch the primary language + all languages from the GitHub API.
	primaryLanguage := ""
	var languagesList []string
	func() {
		defer func() { recover() }()
		primaryLanguage = repo.Language
		if primaryLanguage != "" {
			// The reference's membership test is CASE-SENSITIVE against
			// the ORIGINAL spelling while the INSERTED tag is the
			// lowercase — a "Go" primary inserts "go" even when "go"
			// is already there (the duplicate stays; the quirk is
			// preserved).
			if !containsPyStr(allTags, primaryLanguage) {
				allTags = append([]any{helpers.PyLower(primaryLanguage)},
					allTags...)
			}
		}
		langs, err := r.gh.GetLanguages(repo.FullName)
		if err != nil {
			return // the try/except swallows it
		}
		for i, langName := range langs {
			if i >= 5 {
				break
			}
			if !containsStr(languagesList, langName) {
				languagesList = append(languagesList, langName)
			}
			langLower := helpers.PyLower(langName)
			if !containsPyStr(allTags, langLower) {
				allTags = append(allTags, langLower)
			}
		}
	}()

	tagStrs := make([]string, 0, len(allTags))
	for _, t := range allTags {
		tagStrs = append(tagStrs, helpers.PyStr(t))
	}
	if len(tagStrs) > 12 {
		tagStrs = tagStrs[:12] // cap at 12
	}

	// Fetch the latest release date.
	latestReleaseDate := ""
	func() {
		defer func() { recover() }()
		if rel, err := r.gh.LatestRelease(repo.FullName); err == nil && rel != nil {
			latestReleaseDate = rel.PublishedAt.Format("2006-01-02")
			r.log(fmt.Sprintf("   📦 Latest release: %s (%s)", rel.TagName,
				latestReleaseDate), "info")
		}
	}()

	credScore := CalculateCredibility(orgRep, repo.StargazersCount, commitCount,
		CredibilityDocSource{
			HasIssues: repo.HasIssues, HasLicense: repo.HasLicense,
			HasWiki: repo.HasWiki, Description: repo.Description,
			ReadmeFetchedAny: readmeFetched,
		})

	vaultPath := cfgStr(r.Config, "vault_path")
	if vaultPath == "" {
		// The reference closes the cache and finishes the batch False —
		// the runner surfaces the shape for the shell to drive.
		r.log("Vault path not configured", "error")
		return nil, errVaultMissing
	}

	folder, _ := constants.CategoryFolder(categoryKey)
	if folder == "" {
		folder = "Uncategorized"
	}
	folderPath := filepath.Join(vaultPath, folder)
	_ = dryrun.MkdirAll(folderPath, true)

	// Review queue: low-confidence notes go to _review/.
	reviewMode := confidence < 60
	if reviewMode {
		reviewFolder := filepath.Join(vaultPath, "_review")
		_ = dryrun.MkdirAll(reviewFolder, true)
		folderPath = reviewFolder
		r.log(fmt.Sprintf("⚠️ Low confidence (%s%%) — note sent to _review/ folder",
			helpers.PyFloatRepr(confidence)), "warning")
	}

	// v22 Feature 5: Note Quality Score — flag low-quality notes.
	var qualityIssues []string
	if len(summary) < 50 {
		qualityIssues = append(qualityIssues, "summary too short")
	}
	if isListLen(features, 3) {
		qualityIssues = append(qualityIssues, "fewer than 3 features")
	}
	if confidence < 30 {
		qualityIssues = append(qualityIssues, "low confidence")
	}
	if categoryKey == "Uncategorized" {
		qualityIssues = append(qualityIssues, "uncategorized")
	}
	isLowQuality := len(qualityIssues) > 0
	if isLowQuality && !reviewMode {
		reviewFolder := filepath.Join(vaultPath, "_review")
		_ = dryrun.MkdirAll(reviewFolder, true)
		folderPath = reviewFolder
		r.log(fmt.Sprintf("⚠️ Low quality (%s) — note sent to _review/ folder",
			strings.Join(qualityIssues, ", ")), "warning")
	}

	// Wait for the banner download to finish (join, 15 s timeout).
	bannerPath := ""
	select {
	case bannerPath = <-bannerCh:
	case <-time.After(15 * time.Second):
	}

	// If the banner was downloaded to Uncategorized but the category
	// differs, move it.
	if bannerPath != "" && categoryKey != "Uncategorized" {
		if uncategorizedFolder != "" && strings.Contains(bannerPath,
			uncategorizedFolder) {
			newBanner := filepath.Join(folderPath, filepath.Base(bannerPath))
			if err := dryrun.Move(bannerPath, newBanner); err == nil {
				bannerPath = newBanner
			}
		}
	}

	if bannerPath != "" {
		r.log(fmt.Sprintf("   ✅ Banner: %s", filepath.Base(bannerPath)),
			"success")
	} else {
		// v26 — Fix 5: a missing banner is NOT a problem.
		r.log("   ⚠️ No banner (will retry later) — note still written", "info")
	}

	shortSummary := mapStrDef(llmResult, "short_summary", "")

	noteContent := notebuilder.BuildNote(notebuilder.NoteArgs{
		URL:               url,
		RepoName:          repoName,
		Owner:             ownerLogin,
		OrgName:           orgName,
		Stars:             repo.StargazersCount,
		Forks:             repo.ForksCount,
		CommitCount:       commitCount,
		CredScore:         credScore,
		OrgRep:            orgRep,
		Summary:           summary,
		Tags:              tagStrs,
		CategoryKey:       categoryKey,
		Confidence:        int(confidence),
		HowItWorks:        howItWorks,
		CoreValue:         coreValue,
		Features:          features,
		Difference:        difference,
		BannerPath:        bannerPath,
		PrimaryLanguage:   primaryLanguage,
		Languages:         languagesList,
		ShortSummary:      shortSummary,
		LatestReleaseDate: latestReleaseDate,
		QualityIssues:     qualityIssues,
		IsLowQuality:      isLowQuality,
	})

	// v30 — the atomic write + collision handling.
	filename := storage.BuildNoteFilename(repoName, categoryKey, tagStrs)
	fullPath := storage.UniquePath(filepath.Join(folderPath, filename))

	writeErr := r.writeNote(fullPath, noteContent)
	if writeErr != nil {
		if isDiskFull(writeErr) {
			r.log(fmt.Sprintf("💾 DISK FULL! Cannot write: %s", fullPath),
				"error")
			// v30 — the headless hang-bomb fix: skip the repo, record it
			// in the retry queue, keep the batch moving.
			if r.Headless {
				r.log("⏭️ Headless mode: skipping this repo (disk full). "+
					"Free space, then use 'Retry Failed' in the GUI.",
					"error")
				func() {
					defer func() { recover() }()
					r.Cache.AddFailed(url, "disk full: "+writeErr.Error())
				}()
				r.markFailed(url, "disk full: "+writeErr.Error())
				r.emitProgress()
				return nil, nil
			}
			// The pause flag goes up BEFORE the event (the reference
			// emits a queued signal then sets the flag — the GUI can
			// only ever clear a flag that is already set).
			r.diskFullPaused.Store(true)
			if r.OnDiskFull != nil {
				func() {
					defer func() { recover() }()
					r.OnDiskFull(fullPath)
				}()
			}
			// Wait for resume (the GUI clears the flag after freeing
			// space; Stop still wins).
			for r.diskFullPaused.Load() && r.running {
				r.sleep(1 * time.Second)
			}
			if !r.running {
				return nil, errLoopBreak
			}
			// Retry the write (the disk may still be full).
			if err2 := r.writeNote(fullPath, noteContent); err2 != nil {
				r.log(fmt.Sprintf("❌ Disk still full — skipping %s", url),
					"error")
				r.emitProgress()
				return nil, nil
			}
		} else {
			return nil, writeErr // raise → the per-URL except
		}
	}

	// The bookkeeping chain.
	r.Cache.AddProcessed(repo.ID, url, ownerLogin, repoName, fullPath,
		categoryKey)
	// v0.35.0 — the log shows where each link goes.
	vaultName := "GitHub vault"
	if vaultPath != "" {
		vaultName = filepath.Base(vaultPath)
	}
	relPath := fullPath
	if vaultPath != "" {
		if rel, err := filepath.Rel(vaultPath, fullPath); err == nil {
			relPath = strings.ReplaceAll(rel, string(os.PathSeparator), "/")
		}
	}
	r.log(fmt.Sprintf("✅ [%s] %s/%s processed and stored → %s", vaultName,
		ownerLogin, repoName, relPath), "success")
	// v0.10.0 — Phase 1: record the note the app just WROTE (dry-runs
	// never reach here — no connection).
	if r.NoteStateDB != nil {
		func() {
			defer func() { recover() }() // best-effort by design
			content := noteContent
			cat := categoryKey
			_ = r.NoteStateDB.RecordNote(notestate.VaultGithub, url, fullPath,
				&content, &cat, nil, nil)
		}()
	}
	// v22 Feature 4: mark any previous failure resolved.
	func() {
		defer func() { recover() }()
		r.Cache.MarkFailedResolved(url)
	}()
	// Update the vault index incrementally so the next URL dedupes
	// against this note.
	if r.VaultIndex != nil {
		r.VaultIndex.AddURL(url, fullPath)
	}
	if r.LinkTracker != nil {
		func() {
			defer func() { recover() }()
			r.LinkTracker.MarkProcessed(url, fullPath)
		}()
	}

	entry := ProcessedLogEntry{
		URL:         url,
		Repo:        ownerLogin + "/" + repoName,
		Category:    categoryKey,
		NotePath:    fullPath,
		Banner:      bannerPath != "",
		Credibility: credScore,
	}
	r.processedLog = append(r.processedLog, entry)
	return &entry, nil
}

// errVaultMissing surfaces the reference's early-return shape (cache
// close + finished_signal(False, "Vault path missing")) to the shell.
var errVaultMissing = fmt.Errorf("Vault path missing")

// rereadModel mirrors the v30 model persistence re-read.
func (r *Runner) rereadModel() string {
	switch r.LLMProvider {
	case "ollama":
		if m := nestedCfgStr(r.Config, "ollama", "model"); m != "" {
			return m
		}
		return r.Model
	case "llamacpp":
		if m := cfgStr(r.Config, "llamacpp_model"); m != "" {
			return m
		}
		return r.Model
	default:
		if m := cfgStr(r.Config, "cloud_model"); m != "" {
			return m
		}
		return r.Model
	}
}

func (r *Runner) markSkipped(url, reason string) {
	if r.LinkTracker == nil {
		return
	}
	defer func() { recover() }()
	r.LinkTracker.MarkSkipped(url, reason)
}

func (r *Runner) markFailed(url, errMsg string) {
	if r.LinkTracker == nil {
		return
	}
	defer func() { recover() }()
	r.LinkTracker.MarkFailed(url, errMsg)
}

func (r *Runner) markFailedResolved(url string) {
	if r.Cache == nil {
		return
	}
	defer func() { recover() }()
	r.Cache.MarkFailedResolved(url)
}

// ---------------------------------------------------------------------------
// Small config/map helpers (the house cfgStr pattern).
// ---------------------------------------------------------------------------

// cfgStr reads a config string (Python str(v or ”).strip()).
func cfgStr(config map[string]any, key string) string {
	v, ok := config[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(helpers.PyStr(v))
}

// cfgStrOr reads a config string with a fallback (the .get(k, d)
// shape: only a MISSING key takes the default).
func cfgStrOr(config map[string]any, key, def string) string {
	v, ok := config[key]
	if !ok {
		return def
	}
	if v == nil {
		return ""
	}
	s := helpers.PyStr(v)
	if s == "" {
		return s
	}
	return s
}

// nestedCfgStr reads config["section"]["key"] as a string.
func nestedCfgStr(config map[string]any, section, key string) string {
	m, _ := config[section].(map[string]any)
	if m == nil {
		return ""
	}
	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(helpers.PyStr(v))
}

// pyFloatOr reads a config float with Python's truthiness default
// (0 / missing → the default; the banner throttle + delay keys).
func pyFloatOr(config map[string]any, key string, def float64) float64 {
	v, ok := config[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		if t == 0 {
			return def
		}
		return t
	case int:
		if t == 0 {
			return def
		}
		return float64(t)
	case int64:
		if t == 0 {
			return def
		}
		return float64(t)
	case string:
		return def
	}
	return def
}

// pyIntRaw mirrors int(config.get(key, default) or 0) or None's INT
// half: a MISSING key takes the default; a present number coerces
// with 0 kept as 0 (the "unset" signal the LLM keys use).
func pyIntRaw(config map[string]any, key string, def int) int {
	v, ok := config[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	}
	return 0 // int(garbage) — the reference raises; the port clamps
}

// secs converts Python-seconds to a duration.
func secs(f float64) time.Duration {
	return time.Duration(f * float64(time.Second))
}

// ---------------------------------------------------------------------------
// Result-map readers (the .get defaults).
// ---------------------------------------------------------------------------

func mapStrDef(m map[string]any, key, def string) string {
	if v, ok := m[key]; ok && v != nil {
		if s, ok := v.(string); ok {
			return s
		}
		return helpers.PyStr(v)
	}
	return def
}

func mapDef(m map[string]any, key string, def any) any {
	if v, ok := m[key]; ok {
		return v
	}
	return def
}

// mapFloatDef mirrors float(llm_result.get(key, default)): a number
// passes through, a numeric string coerces, anything else is the
// Python float() TypeError → the per-URL except.
func mapFloatDef(m map[string]any, key string, def float64) (float64, error) {
	v, ok := m[key]
	if !ok {
		return def, nil
	}
	if v == nil {
		return 0, fmt.Errorf("float() argument must be a string or a " +
			"number, not 'NoneType'")
	}
	switch t := v.(type) {
	case float64:
		return t, nil
	case int:
		return float64(t), nil
	case int64:
		return float64(t), nil
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("float() argument must be a string or a real "+
			"number, not '%s'", string(t))
	case string:
		var f float64
		if _, err := fmt.Sscanf(strings.TrimSpace(t), "%g", &f); err == nil {
			return f, nil
		}
		return 0, fmt.Errorf("float() argument must be a string or a real "+
			"number, not '%s'", t)
	}
	return 0, fmt.Errorf("float() argument must be a string or a real number")
}

// mapList mirrors list(llm_result.get(key, [])): a JSON array passes
// through; a STRING becomes its character list (Python's list() on a
// str — the pathological LLM-output quirk, preserved).
func mapList(m map[string]any, key string) []any {
	v, ok := m[key]
	if !ok || v == nil {
		return nil
	}
	switch t := v.(type) {
	case []any:
		return t
	case []string:
		out := make([]any, len(t))
		for i, s := range t {
			out[i] = s
		}
		return out
	case string:
		out := make([]any, 0, len(t))
		for _, ch := range t {
			out = append(out, string(ch))
		}
		return out
	}
	return nil
}

func containsStr(list []string, s string) bool {
	for _, x := range list {
		if x == s {
			return true
		}
	}
	return false
}

func containsPyStr(list []any, s string) bool {
	for _, x := range list {
		if helpers.PyStr(x) == s {
			return true
		}
	}
	return false
}

func inCategoryKeys(cat string) bool {
	for _, k := range constants.CategoryKeys {
		if k == cat {
			return true
		}
	}
	return false
}

// isListLen mirrors isinstance(features, list) and len(features) < 3 —
// the quality check fires ONLY on real lists; a scalar features value
// is not a quality issue (the note's own sanitizers shape it).
func isListLen(v any, n int) bool {
	switch t := v.(type) {
	case []any:
		return len(t) < n
	case []string:
		return len(t) < n
	}
	return false
}

// isDiskFull mirrors the OSError substring law: "No space" (Python's
// strerror spelling, capital N) or "disk" (case-insensitive) — plus
// Go's own errno spelling ("no space left on device", lowercase),
// which is what a real modernc/ENOSPC write failure renders on this
// side of the port. The reference's law is the SHAPE (a disk-full
// OSError), not the casing.
func isDiskFull(err error) bool {
	if err == nil {
		return false
	}
	s := err.Error()
	return strings.Contains(s, "No space") ||
		strings.Contains(s, "no space left on device") ||
		strings.Contains(strings.ToLower(s), "disk")
}

// defaultSystemPrompt reads the constants twin.
func defaultSystemPrompt() string { return constants.DefaultSystemPrompt }
