// vaultseal.go — the construction, the result shape, and the two
// plumbing seams (_git and _api): constants, SealResult
// (ok/describe/to_dict), the VaultSeal struct (the constructor's
// doors — the v0.83 creation guard, the v0.89 belt), the git
// subprocess runner (env scrubbing + timeouts per ADR-0009), and the
// GitHub REST call.
package vaultseal

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
	"regexp"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// GITHUB_API mirrors GITHUB_API.
const GITHUBAPI = "https://api.github.com"

// DEFAULT_TIMEOUT mirrors DEFAULT_TIMEOUT — seconds, per git
// invocation.
const DefaultTimeout = 60

// PushTimeout mirrors PUSH_TIMEOUT — pushes can be slow on big
// vaults / slow links.
const PushTimeout = 180

// APITimeout mirrors API_TIMEOUT — seconds, per GitHub API call.
const APITimeout = 30

// IDENTITY_NAME / IDENTITY_EMAIL mirror the local-only identity
// (_ensure_identity): GitCurator VaultSeal <vaultseal@users.noreply.
// github.com> — only when unset, never the global config.
const IdentityName = "GitCurator VaultSeal"

const IdentityEmail = "vaultseal@users.noreply.github.com"

// DefaultIgnoreLines mirrors DEFAULT_IGNORE_LINES: machine-specific
// Obsidian state + OS noise, sealed OUT of the backup.
var DefaultIgnoreLines = []string{
	".obsidian/workspace.json",
	".obsidian/workspace-mobile.json",
	".obsidian/workspace.json.bak",
	".obsidian/cache",
	".trash/",
	"__pycache__/",
	"*.pyc",
	".DS_Store",
	"Thumbs.db",
	"desktop.ini",
	"*.tmp",
	"*.bak",
	"~$*",
}

// GitignoreHeader mirrors GITIGNORE_HEADER.
const GitignoreHeader = "# --- VaultSeal (machine-specific state — not backed up) ---"

// LocalOnlyHeader mirrors LOCAL_ONLY_HEADER — the second managed
// block (the belt).
const LocalOnlyHeader = "# --- VaultSeal (local-only notes — kept on this machine, never mirrored) ---"

// EmptyTree mirrors _EMPTY_TREE: the well-known empty git tree
// (content-addressed — stable across git versions); the delta
// baseline when the mirror has no commits.
const EmptyTree = "4b825dc642cb6eb9a060e54bf8d69288fbee4904"

// clockNow is the package's clock seam (the house pattern —
// websitepipeline/repobanish's twin): the commit/resync messages and
// the rescue branch name ride it; tests freeze it.
var clockNow = time.Now

// LogFn mirrors LogFn = Callable[[str, str], None]: (message,
// level) with level in info/warning/success.
type LogFn func(msg, level string)

// noopLog mirrors _noop_log.
func noopLog(string, string) {}

// slugifyRe collapses non-alphanumerics, slugDashes collapses runs.
var (
	slugifyRe  = regexp.MustCompile(`[^a-zA-Z0-9]+`)
	slugDashes = regexp.MustCompile(`-{2,}`)
)

// Slugify mirrors slugify: 'Github Projects(Automated)' →
// 'github-projects-automated'.
func Slugify(name string) string {
	s := slugDashes.ReplaceAllString(
		helpers.PyLower(strings.Trim(slugifyRe.ReplaceAllString(name, "-"), "-")), "-")
	return s
}

// ---------------------------------------------------------------------------
// Result
// ---------------------------------------------------------------------------

// SealResult mirrors the SealResult dataclass: ok == true means the
// vault is safe (a commit was created, or there was genuinely
// nothing to do). The Python None doors (skipped_reason, error,
// commit_sha, repo_name, repo_url) type-narrow to "" — ToMap re-opens
// them as null for the JSON shapes (the to_dict parity door).
type SealResult struct {
	Sealed        bool   // a commit was created
	SkippedReason string // set when nothing needed doing
	Error         string // set when the seal failed
	CommitSha     string
	CommitMessage string
	FilesChanged  int
	Pushed        bool
	RepoName      string // "owner/repo" when resolved
	RepoURL       string // https://github.com/owner/repo
	Private       bool
	DurationMS    int
	FilesAdded    int // v0.36.0 — resync mirror delta
	FilesDeleted  int // v0.36.0 — resync mirror delta
}

// NewSealResult mirrors SealResult()'s defaults (private=True is the
// dataclass default; the zero value would be false).
func NewSealResult() *SealResult {
	return &SealResult{Private: true}
}

// OK mirrors the ok property: error ⇒ false; else sealed or skipped.
func (r *SealResult) OK() bool {
	if r.Error != "" {
		return false
	}
	return r.Sealed || r.SkippedReason != ""
}

// Describe mirrors describe() — the one-line verdict for the CLI.
func (r *SealResult) Describe() string {
	if r.SkippedReason != "" {
		return "skipped — " + r.SkippedReason
	}
	if r.Error != "" {
		base := "failed — " + r.Error
		if r.Sealed {
			base += " (commit kept locally)"
		}
		return base
	}
	var bits []string
	if strings.HasPrefix(r.CommitMessage, "resync:") {
		bits = []string{fmt.Sprintf(
			"resync: mirror follows the vault — %d file(s) deleted, %d added/updated",
			r.FilesDeleted, r.FilesAdded)}
	} else {
		bits = []string{fmt.Sprintf("sealed %d file(s)", r.FilesChanged)}
	}
	if r.RepoName != "" {
		bits = append(bits, "→ "+r.RepoName)
	}
	if r.CommitSha != "" {
		bits = append(bits, "("+r.CommitSha+")")
	}
	if !r.Pushed {
		bits = append(bits, "[local only — no push]")
	}
	return strings.Join(bits, " ")
}

// ToMap mirrors to_dict() — the JSON shape with the None doors
// re-opened as nil (insertion-ordered for the CLI's json.dumps).
func (r *SealResult) ToMap() *storage.OrderedMap {
	m := storage.NewOrderedMap()
	m.Set("sealed", r.Sealed)
	m.Set("skipped_reason", pyNoneIfEmpty(r.SkippedReason))
	m.Set("error", pyNoneIfEmpty(r.Error))
	m.Set("commit_sha", pyNoneIfEmpty(r.CommitSha))
	m.Set("commit_message", r.CommitMessage)
	m.Set("files_changed", r.FilesChanged)
	m.Set("pushed", r.Pushed)
	m.Set("repo_name", pyNoneIfEmpty(r.RepoName))
	m.Set("repo_url", pyNoneIfEmpty(r.RepoURL))
	m.Set("private", r.Private)
	m.Set("duration_ms", r.DurationMS)
	m.Set("files_added", r.FilesAdded)
	m.Set("files_deleted", r.FilesDeleted)
	m.Set("ok", r.OK())
	return m
}

func pyNoneIfEmpty(s string) any {
	if s == "" {
		return nil
	}
	return s
}

// ---------------------------------------------------------------------------
// VaultSeal
// ---------------------------------------------------------------------------

// apiFn mirrors the _api method's shape (the tests replace it on the
// instance exactly like the reference's creation-guard suite):
// (http_status, payload). Network failures answer status 0.
type apiFn func(method, apiPath string, body map[string]any) (int, map[string]any)

// RunSummary mirrors the run_summary dict seal() reads: processed and
// total are OPTIONAL (nil = the key was absent — Python None); the
// snapshot message is chosen exactly like _commit_message's
// isinstance(int) door. The app wiring builds these from real ints.
type RunSummary struct {
	Processed *int
	Total     *int
}

// VaultSeal mirrors the VaultSeal class: seal one Obsidian vault into
// its private GitHub mirror.
type VaultSeal struct {
	vault string // "" = Python None
	token string // "" = Python None

	repoName    string
	allowCreate bool // the v0.83 creation guard's one-time opt-in
	localOnly   []string
	autoPush    bool // auto_push AND token
	enabled     bool
	timeout     int
	log         LogFn

	owner     string // resolved lazily from the token
	ownerSet  bool   // the _owner cache's "resolved" bit (failures retry)
	repoFull  string // "owner/repo" after ensureRepo
	remoteURL string

	// The instance seams the reference's tests patch on the object
	// (mock.patch.object(vs, "_push", …) / "ensure_repo"): bound in
	// NewVaultSeal, replaceable per test.
	apiFn    apiFn
	pushFn   func(pushURL string) (bool, string)
	ensureFn func() (bool, string)
}

// NewVaultSeal mirrors __init__ (the same parameter surface, in the
// same order). vaultPath/token/repoName strip; autoPush degrades to
// local-only without a token; timeout 0 → DefaultTimeout; localOnly
// keeps only non-blank patterns; log defaults to the noop.
func NewVaultSeal(vaultPath, token, repoName string, autoPush, enabled bool,
	timeout int, log LogFn, allowCreate bool, localOnly []string) *VaultSeal {
	if log == nil {
		log = noopLog
	}
	raw := strings.TrimSpace(vaultPath)
	vault := ""
	if raw != "" {
		vault = pyExpandUser(raw)
	}
	tok := strings.TrimSpace(token)
	s := &VaultSeal{
		vault:       vault,
		token:       tok,
		repoName:    strings.TrimSpace(repoName),
		allowCreate: allowCreate,
		autoPush:    autoPush && tok != "",
		enabled:     enabled,
		timeout:     pyOrInt(timeout, DefaultTimeout),
		log:         log,
	}
	for _, p := range localOnly {
		if p = strings.TrimSpace(p); p != "" {
			s.localOnly = append(s.localOnly, p)
		}
	}
	// Pushing requires a token — without one we degrade to local-only
	// (autoPush already carries the AND).
	s.apiFn = s.realAPI
	s.pushFn = s.push
	s.ensureFn = s.ensureRepo
	return s
}

// DefaultRepoName mirrors default_repo_name: slugify(vault.name).
func (s *VaultSeal) DefaultRepoName() string {
	if s.vault == "" {
		return ""
	}
	return Slugify(filepath.Base(s.vault))
}

// vaultRepr mirrors f"{self.vault}": Python renders None as "None" —
// the not-found error message keeps the quirk.
func (s *VaultSeal) vaultRepr() string {
	if s.vault == "" {
		return "None"
	}
	return s.vault
}

// ---------------------------------------------------------------------------
// _git — the subprocess runner (ADR-0009)
// ---------------------------------------------------------------------------

// gitEnv mirrors goodrepos._git_env: strip repo-hijacking vars, never
// prompt (GIT_TERMINAL_PROMPT=0).
func gitEnv() []string {
	stripped := map[string]bool{
		"GIT_DIR": true, "GIT_WORK_TREE": true, "GIT_INDEX_FILE": true,
		"GIT_CONFIG": true, "GIT_CONFIG_NOSYSTEM": true,
	}
	env := os.Environ()
	out := make([]string, 0, len(env)+1)
	for _, kv := range env {
		if i := strings.IndexByte(kv, '='); i > 0 {
			if stripped[kv[:i]] {
				continue
			}
		}
		out = append(out, kv)
	}
	out = append(out, "GIT_TERMINAL_PROMPT=0")
	return out
}

// git mirrors _git: run git inside the vault; returns (returncode,
// combined output). Never errors — missing git / timeouts come back
// as rc 127 / 124 (Python's FileNotFoundError / TimeoutExpired
// doors). timeout 0 → s.timeout. THE REDACTION LAW (the spec's Go
// hardening over the reference): the token never survives in ANY
// captured output — git may echo the one-time push/fetch URL in
// errors, so every output pass through the redaction door.
func (s *VaultSeal) git(args []string, timeout int) (int, string) {
	if s.vault == "" {
		return 127, "no vault path configured"
	}
	if timeout <= 0 {
		timeout = s.timeout
	}
	full := append([]string{"-C", s.vault}, args...)
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
			// (Python's (proc.returncode, out)).
			return ee.ExitCode(), s.redact(strings.TrimSpace(stdout.String() + stderr.String()))
		}
		// missing git / spawn failure — Python's OSError door.
		return 127, fmt.Sprintf("git not available: %v", err)
	}
	return 0, s.redact(strings.TrimSpace(stdout.String() + stderr.String()))
}

// redact masks the token in any captured output ("***" — the
// FETCH_HEAD mask). A no-op when no token is configured.
func (s *VaultSeal) redact(out string) string {
	if s.token == "" || !strings.Contains(out, s.token) {
		return out
	}
	return strings.ReplaceAll(out, s.token, "***")
}

// gitf is the variadic convenience over git.
func (s *VaultSeal) gitf(args ...string) (int, string) {
	return s.git(args, 0)
}

// gitPush is the push/fetch door (the 180 s timeout).
func (s *VaultSeal) gitPush(args ...string) (int, string) {
	return s.git(args, PushTimeout)
}

// ---------------------------------------------------------------------------
// _api — the GitHub REST seam
// ---------------------------------------------------------------------------

// realAPI mirrors _api: GitHub REST call with the token. Returns
// (http_status, payload); network failures answer status 0 — never
// errors.
func (s *VaultSeal) realAPI(method, apiPath string, body map[string]any) (int, map[string]any) {
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
	req.Header.Set("User-Agent", "GitCurator-VaultSeal")
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

// apiMessage renders payload.get('message', status) — the int status
// when the message key is absent.
func apiMessage(status int, payload map[string]any) string {
	if v, ok := payload["message"]; ok && v != nil {
		return helpers.PyStr(v)
	}
	return fmt.Sprintf("%d", status)
}

// ownerLogin mirrors _owner_login: resolve the token's login (cached
// on success; failures retry on the next call, like Python's
// None-stays-None cache).
func (s *VaultSeal) ownerLogin() string {
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
				"VaultSeal: could not resolve the GitHub user for the token (%d: %s)%s",
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

// pyExpandUser mirrors Path(raw).expanduser() for the shapes a config
// vault path takes: a leading "~" expands to the home dir. Python's
// "~user" form (a pwd lookup) is not ported — the config never
// carries it (the house usermemory precedent; a documented
// convergence).
func pyExpandUser(raw string) string {
	p := pyPathNormalize(raw)
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			p = pyPathNormalize(filepath.Join(home, strings.TrimPrefix(p, "~")))
		}
	}
	return p
}

// pyPathNormalize mirrors PurePosixPath's lexical parse: empty
// segments and "." drop, "//" collapses, trailing "/" drops — but ".."
// stays a literal segment (Python does NOT resolve it).
func pyPathNormalize(p string) string {
	if p == "" {
		return "."
	}
	abs := strings.HasPrefix(p, "/")
	var parts []string
	for _, seg := range strings.Split(p, "/") {
		if seg == "" || seg == "." {
			continue
		}
		parts = append(parts, seg)
	}
	if len(parts) == 0 {
		if abs {
			return "/"
		}
		return "."
	}
	out := strings.Join(parts, "/")
	if abs {
		out = "/" + out
	}
	return out
}

func pyOrInt(v, fallback int) int {
	if v == 0 {
		return fallback
	}
	return v
}
