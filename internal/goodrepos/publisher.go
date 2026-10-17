// publisher.go — the GoodRepos engine: the construction (the v0.83
// creation guard's opt-in), the _git subprocess runner (env scrubbing,
// timeouts, THE REDACTION LAW), the GitHub REST seam, _ensure_repo
// (the duplicate-repo accident cure), _fetch_history (+ the FETCH_HEAD
// scrub), the staging tree builder, and publish() itself.
package goodrepos

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// apiFn mirrors the _api method's shape (the tests replace it on the
// instance exactly like the reference's creation-guard suite):
// (http_status, payload). Network failures answer status 0.
type apiFn func(method, apiPath string, body map[string]any) (int, map[string]any)

// RunSummary mirrors the run_summary dict publish() reads: processed
// and total are OPTIONAL (nil = the key was absent — Python None).
// The app wiring builds these from real ints.
type RunSummary struct {
	Processed *int
	Total     *int
}

// GoodRepos mirrors the GoodRepos class: publish one curated Obsidian
// vault as its public good-repos directory.
type GoodRepos struct {
	vault string // "" = Python None
	token string // "" = Python None

	repoName    string
	allowCreate bool // the v0.83 creation guard's one-time opt-in
	autoPush    bool // auto_push AND token
	enabled     bool
	timeout     int
	log         LogFn

	owner     string // resolved lazily from the token
	ownerSet  bool   // the _owner cache's "resolved" bit (failures retry)
	repoFull  string // "owner/repo" after ensureRepo
	remoteURL string
	staging   string // "" = Python None

	// The instance seams (the T-47 law: the reference's mock surface
	// maps to instance fields): apiFn is what the reference's
	// creation-guard suite patches; tokenURLFn is the _token_url
	// builder — the one door that lets the fetch/push ladder run
	// against a LOCAL bare repo in the tests (the default builds the
	// one-time https token URL; the token NEVER lands in .git/config).
	apiFn      apiFn
	tokenURLFn func() string
}

// NewGoodRepos mirrors __init__ (the same parameter surface, in the
// same order). vaultPath/token/repoName strip; autoPush degrades
// without a token; timeout 0 → DefaultTimeout; log defaults to the
// noop.
func NewGoodRepos(vaultPath, token, repoName string, autoPush, enabled bool,
	timeout int, log LogFn, allowCreate bool) *GoodRepos {
	if log == nil {
		log = noopLog
	}
	raw := strings.TrimSpace(vaultPath)
	vault := ""
	if raw != "" {
		vault = pyExpandUser(raw)
	}
	tok := strings.TrimSpace(token)
	s := &GoodRepos{
		vault:       vault,
		token:       tok,
		repoName:    strings.TrimSpace(repoName),
		allowCreate: allowCreate,
		autoPush:    autoPush && tok != "",
		enabled:     enabled,
		timeout:     pyOrInt(timeout, DefaultTimeout),
		log:         log,
	}
	s.apiFn = s.realAPI
	s.tokenURLFn = s.defaultTokenURL
	return s
}

// RepoName mirrors the repo_name field (the raw configured name — ""
// means NOT configured, v0.83.0).
func (s *GoodRepos) RepoName() string { return s.repoName }

// vaultRepr mirrors f"{self.vault}": Python renders None as "None" —
// the not-found error message keeps the quirk.
func (s *GoodRepos) vaultRepr() string {
	if s.vault == "" {
		return "None"
	}
	return s.vault
}

// ---------------------------------------------------------------------------
// _git — the subprocess runner (ADR-0009)
// ---------------------------------------------------------------------------

// git mirrors _git: run git inside the staging dir; returns
// (returncode, combined output). Never errors — missing git /
// timeouts come back as rc 127 / 124. THE REDACTION LAW: the token
// never survives in ANY captured output.
func (s *GoodRepos) git(args []string, timeout int) (int, string) {
	if s.staging == "" {
		return 127, "no staging directory configured"
	}
	if timeout <= 0 {
		timeout = s.timeout
	}
	full := append([]string{"-C", s.staging}, args...)
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(timeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", full...)
	cmd.Env = gitEnv()
	hideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		head := args
		if len(head) > 2 {
			head = head[:2]
		}
		return 124, fmt.Sprintf("git %s timed out after %ds",
			strings.Join(head, " "), timeout)
	}
	if err != nil {
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			// a normal non-zero exit: the output is the story
			return ee.ExitCode(), s.redact(strings.TrimSpace(stdout.String() + stderr.String()))
		}
		// missing git / spawn failure — Python's OSError door.
		return 127, fmt.Sprintf("git not available: %v", err)
	}
	return 0, s.redact(strings.TrimSpace(stdout.String() + stderr.String()))
}

// gitf is the variadic convenience over git.
func (s *GoodRepos) gitf(args ...string) (int, string) {
	return s.git(args, 0)
}

// gitPush is the push/fetch door (the 180 s timeout).
func (s *GoodRepos) gitPush(args ...string) (int, string) {
	return s.git(args, PushTimeout)
}

// redact mirrors _redact: remove the token from any text that will be
// stored or logged. A no-op when no token is configured.
func (s *GoodRepos) redact(text string) string {
	if s.token == "" || text == "" || !strings.Contains(text, s.token) {
		return text
	}
	return strings.ReplaceAll(text, s.token, "***")
}

// ---------------------------------------------------------------------------
// _api — the GitHub REST seam
// ---------------------------------------------------------------------------

// realAPI mirrors _api: GitHub REST call with the token. Returns
// (http_status, payload); network failures answer status 0 — never
// errors.
func (s *GoodRepos) realAPI(method, apiPath string, body map[string]any) (int, map[string]any) {
	var data []byte
	if body != nil {
		data, _ = json.Marshal(body)
	}
	req, err := http.NewRequest(method, GITHUBAPI+apiPath, bytes.NewReader(data))
	if err != nil {
		return 0, map[string]any{"message": err.Error()}
	}
	req.Header.Set("Authorization", "Bearer "+s.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "GitCurator-GoodRepos")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	client := &http.Client{Timeout: time.Duration(APITimeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return 0, map[string]any{"message": err.Error()}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, map[string]any{"message": err.Error()}
	}
	var payload map[string]any
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &payload); err != nil || payload == nil {
			// a non-JSON body (or a non-object root) falls through the
			// same door Python's json.loads-inside-try does: status 0
			// with the parse error as the message.
			return 0, map[string]any{"message": err.Error()}
		}
	} else {
		payload = map[string]any{}
	}
	return resp.StatusCode, payload
}

// apiMessageOr mirrors payload.get('message', fallback).
func apiMessageOr(status int, payload map[string]any, fallback string) string {
	if payload == nil {
		return fallback
	}
	if v, ok := payload["message"]; ok && v != nil {
		return helpers.PyStr(v)
	}
	return fallback
}

// ownerLogin mirrors _owner_login: resolve the token's login (cached
// on success; failures retry on the next call, like Python's
// None-stays-None cache).
func (s *GoodRepos) ownerLogin() string {
	if !s.ownerSet {
		if s.token == "" {
			return ""
		}
		status, payload := s.apiFn("GET", "/user", nil)
		if status != 200 {
			// v32.1: 401 = the token itself is bad — name the fix.
			hint := ""
			if status == 401 {
				hint = " — token invalid or expired; update it in " +
					"Settings → Credentials → GitHub Token (then 'Test GitHub Token')"
			}
			s.log(fmt.Sprintf(
				"GoodRepos: could not resolve the GitHub user for the token (%d: %s)%s",
				status, apiMessageOr(status, payload, "network error"), hint), "warning")
			return ""
		}
		login := ""
		if v, ok := payload["login"]; ok {
			login = helpers.PyStr(v)
		}
		if login == "" {
			return ""
		}
		s.owner = login
		s.ownerSet = true
	}
	return s.owner
}

// defaultTokenURL mirrors _token_url: the one-time authenticated URL —
// used inline in git commands only, never stored in remotes or
// .git/config, never logged.
func (s *GoodRepos) defaultTokenURL() string {
	return "https://x-access-token:" + s.token + "@github.com/" + s.repoFull + ".git"
}

// ---------------------------------------------------------------------------
// _ensure_repo — v0.83.0 THE CREATION GUARD
// ---------------------------------------------------------------------------

// ensureRepo mirrors _ensure_repo: resolve the directory repository.
// A missing repo is an ERROR, never a silent POST /user/repos (the
// duplicate-repo accident cure). When creation IS explicitly allowed
// it creates PRIVATE (the owner's word: no public repos yet).
// Returns (ok, error_message). Safe to call repeatedly.
func (s *GoodRepos) ensureRepo() (bool, string) {
	// v0.83.0 — fail FAST on an unconfigured name: no network call,
	// no owner lookup, nothing happens.
	repo := strings.TrimSpace(s.repoName)
	if repo == "" {
		return false, "no directory repo configured — goodrepos.repo_name " +
			"is empty. Nothing was created or pushed."
	}
	owner := s.ownerLogin()
	if owner == "" {
		// v32.1: actionable — the token was rejected, not a mystery.
		return false, "could not resolve the GitHub account for the token " +
			"(invalid or expired — update the GitHub Token in " +
			"Settings → Credentials, then 'Test GitHub Token')"
	}

	status, payload := s.apiFn("GET", "/repos/"+owner+"/"+repo, nil)
	if status == 200 {
		if v, ok := payload["private"]; ok && pyTruthy(v) {
			s.log(fmt.Sprintf(
				"GoodRepos: %s/%s is PRIVATE — publishing anyway "+
					"(your word: no public repos yet; flip it on GitHub the "+
					"day you want it browsable)", owner, repo), "info")
		}
	} else if status == 404 {
		if !s.allowCreate {
			return false, fmt.Sprintf(
				"repository %s/%s not found — GitCurator does NOT "+
					"create repositories automatically (the duplicate-repo "+
					"accident cure, v0.83.0). Create it yourself on GitHub — "+
					"or set goodrepos.create_repo=true in config.json once, "+
					"or pass --create-repo to the CLI — then re-run.", owner, repo)
		}
		status, payload = s.apiFn("POST", "/user/repos", map[string]any{
			"name":        repo,
			"private":     true,
			"description": RepoDescription,
			"has_issues":  true,
			"has_wiki":    false,
			"auto_init":   false,
		})
		if status != 201 {
			// 422 usually means the name is taken (race) — re-check.
			reStatus, _ := s.apiFn("GET", "/repos/"+owner+"/"+repo, nil)
			if reStatus != 200 {
				return false, fmt.Sprintf("could not create the private repo %s/%s: %s",
					owner, repo, apiMessageOr(status, payload, fmt.Sprintf("%d", status)))
			}
		}
		s.log(fmt.Sprintf("GoodRepos: created PRIVATE repository %s/%s "+
			"(explicit opt-in only)", owner, repo), "success")
	} else {
		return false, fmt.Sprintf("GitHub API error checking %s/%s: %s",
			owner, repo, apiMessageOr(status, payload, fmt.Sprintf("%d", status)))
	}

	s.repoFull = owner + "/" + repo
	s.remoteURL = "https://github.com/" + owner + "/" + repo
	return true, ""
}

// ---------------------------------------------------------------------------
// _fetch_history — align the staging repo with the remote's main
// ---------------------------------------------------------------------------

// fetchHistory mirrors _fetch_history: fetch the remote's main into
// the staging repo and align HEAD/index with it (reset --mixed), so
// the next commit is a real diff on top of the existing history — and
// skip-when-unchanged works. True = the remote history was aligned;
// false = fresh directory.
func (s *GoodRepos) fetchHistory() bool {
	if s.repoFull == "" || s.token == "" {
		return false
	}
	rc, _ := s.gitPush("fetch", s.tokenURLFn(), "main")
	s.sanitizeFetchHead()
	if rc != 0 {
		s.log("GoodRepos: no remote history found — starting a fresh "+
			"directory", "info")
		return false
	}
	rc, out := s.gitf("reset", "--mixed", "FETCH_HEAD")
	if rc != 0 {
		s.log("GoodRepos: could not align with the remote history: "+
			out, "warning")
		return false
	}
	return true
}

// sanitizeFetchHead mirrors _sanitize_fetch_head: git strips
// credentials when it writes .git/FETCH_HEAD, but belt and braces — if
// the token ever leaked in there, redact it. Best effort.
func (s *GoodRepos) sanitizeFetchHead() {
	if s.staging == "" || s.token == "" {
		return
	}
	fh := filepath.Join(s.staging, ".git", "FETCH_HEAD")
	data, err := os.ReadFile(fh)
	if err != nil {
		return
	}
	text := string(data)
	if !strings.Contains(text, s.token) {
		return
	}
	_ = os.WriteFile(fh, []byte(strings.ReplaceAll(text, s.token, "***")), 0o644)
}

// ---------------------------------------------------------------------------
// staging
// ---------------------------------------------------------------------------

// newStaging mirrors _new_staging: a fresh mkdtemp OUTSIDE any
// repository (the sandbox workspace being a git repo can never leak
// into the directory). register=true appends to the process-exit
// cleanup list.
func (s *GoodRepos) newStaging(register bool) string {
	staging, err := os.MkdirTemp("", "goodrepos-")
	if err != nil {
		s.log(fmt.Sprintf("GoodRepos: could not create a staging directory: %v", err), "warning")
		return ""
	}
	if register {
		registerStaging(staging)
	}
	return staging
}

// cleanStaging mirrors _clean_staging: remove everything except .git
// so the tree can be regenerated without stale files surviving from a
// previous publish.
func (s *GoodRepos) cleanStaging(staging string) {
	entries, err := os.ReadDir(staging)
	if err != nil {
		return
	}
	for _, e := range entries {
		if e.Name() == ".git" {
			continue
		}
		if e.Type()&os.ModeSymlink != 0 {
			_ = os.Remove(filepath.Join(staging, e.Name()))
		} else if e.IsDir() {
			_ = os.RemoveAll(filepath.Join(staging, e.Name()))
		} else {
			_ = os.Remove(filepath.Join(staging, e.Name()))
		}
	}
}

// buildTree mirrors _build_tree: copy the curated notes at their
// vault-relative paths, the root _index.md + links_manifest.json when
// present, and write README.md. Returns the number of files written.
// copy2 converges to a mode-preserving content copy (git only ever
// sees content + the exec bit).
func (s *GoodRepos) buildTree(staging string, cats []*CategoryNotes) int {
	written := 0
	for _, c := range cats {
		for _, note := range c.Notes {
			dest := filepath.Join(staging, filepath.FromSlash(note.RelPath))
			if err := os.MkdirAll(filepath.Dir(dest), 0o755); err == nil {
				if err := copyFile(note.Path, dest); err == nil {
					written++
					continue
				} else {
					s.log(fmt.Sprintf("GoodRepos: could not copy %s: %v", note.RelPath, err), "warning")
					continue
				}
			} else {
				s.log(fmt.Sprintf("GoodRepos: could not copy %s: %v", note.RelPath, err), "warning")
			}
		}
	}
	for _, name := range []string{"_index.md", "links_manifest.json"} {
		if s.vault == "" {
			continue
		}
		src := filepath.Join(s.vault, name)
		if isFile(src) {
			if err := copyFile(src, filepath.Join(staging, name)); err == nil {
				written++
			} else {
				s.log(fmt.Sprintf("GoodRepos: could not copy %s: %v", name, err), "warning")
			}
		}
	}
	if err := os.WriteFile(filepath.Join(staging, "README.md"),
		[]byte(s.buildReadme(cats)), 0o644); err == nil {
		written++
	} else {
		s.log(fmt.Sprintf("GoodRepos: could not write README.md: %v", err), "warning")
	}
	return written
}

// copyFile is shutil.copy2's content+mode half.
func copyFile(src, dst string) error {
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	mode := os.FileMode(0o644)
	if fi, err := os.Stat(src); err == nil {
		mode = fi.Mode().Perm()
	}
	return os.WriteFile(dst, data, mode)
}

// ---------------------------------------------------------------------------
// publish
// ---------------------------------------------------------------------------

// publishInternal mirrors _publish: the engine room (enabled → vault →
// the v0.83 name guard → the scan → staging → the git ladder → the
// one-time-token push).
func (s *GoodRepos) publishInternal(runSummary *RunSummary, dryRun bool) *PublishResult {
	if !s.enabled {
		return &PublishResult{SkippedReason: "GoodRepos is disabled in config"}
	}
	if s.vault == "" || !isDir(s.vault) {
		return &PublishResult{Error: "vault directory not found: " + s.vaultRepr()}
	}
	// v0.83.0 — an empty repo name means NOT configured: publish skips
	// (no silent default, no auto-creation — the duplicate-repo cure).
	if strings.TrimSpace(s.repoName) == "" && !dryRun {
		return &PublishResult{SkippedReason: "no directory repo configured — goodrepos.repo_name " +
			"is empty (nothing published, nothing created)"}
	}

	if runSummary != nil && runSummary.Processed != nil && runSummary.Total != nil &&
		*runSummary.Total > 0 {
		s.log(fmt.Sprintf("GoodRepos: publishing after a curation run that "+
			"processed %d/%d repos", *runSummary.Processed, *runSummary.Total), "info")
	}

	cats := s.ScanVault()
	entries := 0
	for _, c := range cats {
		entries += len(c.Notes)
	}
	// base is FROZEN here — BEFORE ensureRepo runs (the reference's
	// `self._repo_full or self.repo_name` snapshot quirk: a pushed
	// result still reports the RAW name on the first publish).
	baseRepoName := s.repoName
	if s.repoFull != "" {
		baseRepoName = s.repoFull
	}
	base := func(r *PublishResult) *PublishResult {
		r.Entries = entries
		r.Categories = len(cats)
		r.RepoName = baseRepoName
		r.RepoURL = s.remoteURL
		return r
	}
	if len(cats) == 0 {
		return base(&PublishResult{SkippedReason: "vault contains no curated notes"})
	}

	// -- staging tree (fresh mkdtemp, OUTSIDE any repo) ----------------
	var staging string
	if dryRun {
		staging = s.newStaging(false)
	} else {
		if s.staging == "" {
			s.staging = s.newStaging(true)
		}
		staging = s.staging
		s.cleanStaging(staging)
	}
	if staging == "" {
		return base(&PublishResult{Error: "could not create a staging directory"})
	}

	written := s.buildTree(staging, cats)

	if dryRun {
		return base(&PublishResult{Published: true, FilesChanged: written,
			DryRun: true, StagingDir: staging})
	}

	// -- git: init + identity --------------------------------------------
	rc, out := s.gitf("init", "-b", "main")
	if rc != 0 {
		// older git without -b: plain init + move HEAD to main
		rc2, out2 := s.gitf("init")
		if rc2 == 0 {
			s.gitf("symbolic-ref", "HEAD", "refs/heads/main")
		} else {
			return base(&PublishResult{Error: "git init failed: " + orStr(out2, out)})
		}
	}
	s.gitf("config", "user.name", IdentityName)
	s.gitf("config", "user.email", IdentityEmail)

	// -- remote prep (only when pushing is possible) ---------------------
	if s.autoPush {
		ok, err := s.ensureRepo()
		if !ok {
			return base(&PublishResult{Error: err})
		}
		s.fetchHistory()
	}

	// -- stage + skip-when-unchanged ---------------------------------------
	s.gitf("add", "-A", "--", ".")
	rc, porcelain := s.gitf("status", "--porcelain")
	rcHead, _ := s.gitf("rev-parse", "--verify", "-q", "HEAD")
	var dirty []string
	if rc == 0 {
		for _, ln := range helpers.SplitLines(porcelain) {
			if strings.TrimSpace(ln) != "" {
				dirty = append(dirty, ln)
			}
		}
	}

	if len(dirty) == 0 && rcHead == 0 {
		return base(&PublishResult{
			SkippedReason: "directory unchanged since the last publish",
			StagingDir:    staging,
		})
	}

	ts := clockNow().UTC().Format("2006-01-02 15:04") + " UTC"
	message := fmt.Sprintf("publish: %d repos across %d categories — %s",
		entries, len(cats), ts)
	rc, out = s.gitf("-c", "commit.gpgsign=false", "commit", "-m", message, "--quiet")
	if rc != 0 {
		return base(&PublishResult{Error: "git commit failed: " + out, StagingDir: staging})
	}
	_, sha := s.gitf("rev-parse", "--short=7", "HEAD")

	result := base(&PublishResult{
		Published:     true,
		CommitSha:     sha,
		CommitMessage: message,
		FilesChanged:  len(dirty),
		StagingDir:    staging,
	})

	if !s.autoPush {
		s.log("GoodRepos: committed locally — no GitHub token configured, "+
			"push skipped (set the token in Settings to enable pushing)", "warning")
		return result
	}

	// -- push via one-time token URL (never in .git/config, never logged) --
	s.gitf("remote", "remove", "origin") // defensive: no token-less remotes
	rc, out = s.gitPush("push", s.tokenURLFn(), "HEAD:refs/heads/main")
	s.gitf("remote", "remove", "origin")
	if rc != 0 {
		result.Error = "push failed: " + helpers.CutRunes(out, 300)
		s.log("GoodRepos: "+result.Error, "warning")
	} else {
		result.Pushed = true
		s.log(fmt.Sprintf("GoodRepos: published %d repos across %d "+
			"categories → %s (%s)", entries, len(cats), s.repoFull, sha), "success")
	}
	return result
}

// Publish mirrors publish(): build + commit + best-effort push of the
// public directory. The reference's try/except wrapper is unreachable
// in Go (no code path panics — the T-46 never-raises convergence);
// the honest result is the answer.
func (s *GoodRepos) Publish(runSummary *RunSummary, dryRun bool) *PublishResult {
	started := time.Now()
	result := s.publishInternal(runSummary, dryRun)
	result.DurationMS = int(time.Since(started).Milliseconds())
	return result
}

// Status mirrors status(): the pure vault scan snapshot for UIs — no
// git calls, no network. The dict's insertion order is the
// reference's literal.
func (s *GoodRepos) Status() *storage.OrderedMap {
	cats := s.ScanVault()
	tree := storage.NewOrderedMap()
	rels := make([]string, 0, len(cats))
	for _, c := range cats {
		rels = append(rels, c.Rel)
	}
	sortStrings(rels)
	for _, rel := range rels {
		for _, c := range cats {
			if c.Rel == rel {
				tree.Set(rel, len(c.Notes))
				break
			}
		}
	}
	entries := 0
	for _, c := range cats {
		entries += len(c.Notes)
	}
	st := storage.NewOrderedMap()
	st.Set("vault_path", pyNoneIfEmpty(s.vault))
	st.Set("exists", s.vault != "" && isDir(s.vault))
	st.Set("enabled", s.enabled)
	st.Set("auto_push", s.autoPush)
	repoName := s.repoName
	if s.repoFull != "" {
		repoName = s.repoFull
	}
	st.Set("repo_name", pyNoneIfEmpty(repoName))
	st.Set("repo_url", pyNoneIfEmpty(s.remoteURL))
	st.Set("entries", entries)
	st.Set("categories", tree.Len())
	st.Set("tree", tree)
	st.Set("generated", clockNow().UTC().Format("2006-01-02"))
	return st
}

// orStr mirrors Python's `a or b` for the two git outputs.
func orStr(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

// pyExpandUser mirrors Path(raw).expanduser() (the vaultseal twin).
func pyExpandUser(raw string) string {
	p := raw
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

func pyOrInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}
