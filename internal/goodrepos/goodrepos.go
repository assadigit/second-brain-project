// goodrepos.go — the construction's constants, the deterministic
// helpers (github_anchor / display_name / emoji_for / format_stars),
// PublishResult (ok/describe/to_dict), and the staging lifetime
// (mkdtemp OUTSIDE any repository, reuse across publish() calls,
// process-exit cleanup).
package goodrepos

import (
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// GITHUBAPI mirrors GITHUB_API.
const GITHUBAPI = "https://api.github.com"

// DefaultTimeout mirrors DEFAULT_TIMEOUT — seconds, per local git
// invocation.
const DefaultTimeout = 60

// PushTimeout mirrors PUSH_TIMEOUT — pushes/fetches can be slow on big
// vaults.
const PushTimeout = 180

// APITimeout mirrors API_TIMEOUT — seconds, per GitHub API call.
const APITimeout = 30

// DefaultRepoName mirrors DEFAULT_REPO_NAME: kept ONLY for backwards
// import compatibility (v0.83.0) — a GoodRepos with an EMPTY repo_name
// skips publishing. Nothing falls back to this constant anymore.
const DefaultRepoName = "good-repos"

// RepoDescription mirrors REPO_DESCRIPTION — the private-create
// payload's description.
const RepoDescription = "✨ A curated directory of good GitHub repositories — auto-maintained by GitCurator"

// IdentityName / IdentityEmail mirror IDENTITY_NAME / IDENTITY_EMAIL —
// the staging repo's local-only identity.
const IdentityName = "GoodRepos Publisher"

const IdentityEmail = "goodrepos@users.noreply.github.com"

// IgnoredDirs mirrors IGNORED_DIRS: machine-specific Obsidian state +
// scaffolding, never published.
var IgnoredDirs = map[string]bool{
	".obsidian": true, ".trash": true, "_inbox": true, "_review": true,
	"_moc": true, "__pycache__": true, ".git": true,
}

// CategoryEmoji mirrors CATEGORY_EMOJI: raw category folder component
// → emoji (fallback 📁).
var CategoryEmoji = map[string]string{
	"AI-Domain":        "🤖",
	"Agents":           "🧠",
	"Skills":           "✨",
	"MCP":              "🔌",
	"LLM-Tools":        "🧰",
	"Tools":            "🛠️",
	"Scraping":         "🕷️",
	"Automation":       "⚙️",
	"Dev-Tools":        "🧰",
	"Networking":       "🌐",
	"Media":            "🎨",
	"Documentation":    "📚",
	"Guides":           "📖",
	"References":       "📑",
	"Research":         "🔬",
	"Frameworks":       "🏗️",
	"Web-Frameworks":   "🌐",
	"Backend":          "🗄️",
	"Frontend":         "🎨",
	"Infrastructure":   "🏗️",
	"Deployment":       "🚀",
	"Cloud":            "☁️",
	"Containerization": "📦",
	"Uncategorized":    "📦",
}

// clockNow is the package's clock seam (the house pattern): the
// README's "Updated <date>" line, the commit message's UTC stamp, and
// Status()'s generated date ride it; tests freeze it.
var clockNow = time.Now

// LogFn mirrors LogFn = Callable[[str, str], None]: (message, level)
// with level in info/warning/success.
type LogFn func(msg, level string)

// noopLog mirrors _noop_log.
func noopLog(string, string) {}

// ---------------------------------------------------------------------------
// Small deterministic helpers (public — reused by the tests)
// ---------------------------------------------------------------------------

// GitHubAnchor mirrors github_anchor: the anchor GitHub generates for
// a heading — lowercase, drop anything that is not alphanumeric/space/
// hyphen (emoji, punctuation), spaces → hyphens ("🤖 AI Domain" →
// "-ai-domain" — the emoji leaves a leading hyphen).
func GitHubAnchor(text string) string {
	s := helpers.PyLower(helpers.Strip(text))
	var kept []rune
	for _, ch := range s {
		if pyIsAlnum(ch) || ch == ' ' || ch == '-' {
			kept = append(kept, ch)
		}
	}
	return strings.ReplaceAll(string(kept), " ", "-")
}

// DisplayName mirrors display_name: folder component → display name,
// hyphen → space ("AI-Domain" → "AI Domain").
func DisplayName(folder string) string {
	return helpers.Strip(strings.ReplaceAll(folder, "-", " "))
}

// EmojiFor mirrors emoji_for: look the components up from the deepest
// one backwards; fallback 📁.
func EmojiFor(components ...string) string {
	for i := len(components) - 1; i >= 0; i-- {
		comp := components[i]
		if comp != "" {
			if e, ok := CategoryEmoji[comp]; ok {
				return e
			}
		}
	}
	return "📁"
}

// FormatStars mirrors format_stars: thousands separators below 10k,
// compact "1.2k" style from 10k up (140232 → "140.2k", 10000 → "10k").
// nil = Python None → "".
func FormatStars(stars *int) string {
	if stars == nil {
		return ""
	}
	n := *stars
	if n >= 10_000 {
		s := strconv.FormatFloat(float64(n)/1000.0, 'f', 1, 64)
		if strings.HasSuffix(s, ".0") {
			s = s[:len(s)-2]
		}
		return s + "k"
	}
	return helpers.FormatComma(int64(n))
}

// FormatStarsAny mirrors format_stars over the parsed-note shapes —
// the Python try door: int(x) with TypeError/ValueError → "".
func FormatStarsAny(stars any) string {
	switch v := stars.(type) {
	case nil:
		return ""
	case *int:
		return FormatStars(v)
	case int:
		return FormatStars(&v)
	case int64:
		n := int(v)
		return FormatStars(&n)
	case bool:
		n := 0 // int(True) == 1 in Python
		if v {
			n = 1
		}
		return FormatStars(&n)
	case float64:
		n := int(v) // Python int(86.75) truncates
		return FormatStars(&n)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(v))
		if err != nil {
			return ""
		}
		return FormatStars(&n)
	default:
		return ""
	}
}

// introName mirrors _intro_name: the top-category intro sentence's
// display name — lowercase, keeping short ALL-CAPS acronyms ("AI
// Domain" → "AI domain", "Tools" → "tools").
func introName(display string) string {
	var words []string
	for _, w := range strings.Fields(display) {
		if len([]rune(w)) <= 4 && pyIsAlpha(w) && pyIsUpper(w) {
			words = append(words, w)
		} else {
			words = append(words, helpers.PyLower(w))
		}
	}
	return strings.Join(words, " ")
}

// pyIsAlpha mirrors str.isalpha: non-empty, every char alphabetic.
func pyIsAlpha(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsLetter(r) {
			return false
		}
	}
	return true
}

// pyIsUpper mirrors str.isupper: at least one cased char, every cased
// char uppercase.
func pyIsUpper(s string) bool {
	cased := 0
	for _, r := range s {
		if isCasedRune(r) {
			cased++
			if !unicode.IsUpper(r) {
				return false
			}
		}
	}
	return cased > 0
}

// isCasedRune mirrors the Unicode Lowercase|Uppercase|Titlecase
// property (Python's str cased chars).
func isCasedRune(r rune) bool {
	return unicode.IsLower(r) || unicode.IsUpper(r) || unicode.IsTitle(r)
}

// pyIsAlnum mirrors str.isalnum for one char (letter or number).
func pyIsAlnum(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsNumber(r)
}

// ---------------------------------------------------------------------------
// PublishResult
// ---------------------------------------------------------------------------

// PublishResult mirrors the PublishResult dataclass: ok == true means
// the public directory is up to date (a commit was pushed/kept, there
// was genuinely nothing to do, or a dry-run built the tree). The
// Python None doors type-narrow to "" — ToMap re-opens them as null.
type PublishResult struct {
	Published     bool   // a directory was built/committed
	SkippedReason string // set when nothing needed doing
	Error         string // set when the publish failed
	CommitSha     string
	CommitMessage string
	FilesChanged  int
	Pushed        bool
	RepoName      string // "owner/repo" or the raw configured name
	RepoURL       string // https://github.com/owner/repo
	Entries       int    // curated repos found
	Categories    int    // category folders found
	DurationMS    int
	DryRun        bool
	StagingDir    string // where the tree was built
}

// NewPublishResult mirrors PublishResult()'s defaults (all zero).
func NewPublishResult() *PublishResult { return &PublishResult{} }

// OK mirrors the ok property: error ⇒ false; else published or
// skipped.
func (r *PublishResult) OK() bool {
	if r.Error != "" {
		return false
	}
	return r.Published || r.SkippedReason != ""
}

// Describe mirrors describe() — the one-line verdict for the CLI.
func (r *PublishResult) Describe() string {
	if r.SkippedReason != "" {
		return "skipped — " + r.SkippedReason
	}
	if r.Error != "" {
		base := "failed — " + r.Error
		if r.Published {
			base += " (commit kept locally)"
		}
		return base
	}
	if r.DryRun {
		staging := r.StagingDir
		if staging == "" {
			staging = "None" // the f"{None}" quirk (unreachable in practice)
		}
		return "built a directory of " + strconv.Itoa(r.Entries) + " repos across " +
			strconv.Itoa(r.Categories) + " categories (dry-run — no git, no push) — staged at " +
			staging
	}
	bits := []string{"published " + strconv.Itoa(r.Entries) + " repos across " +
		strconv.Itoa(r.Categories) + " categories"}
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
func (r *PublishResult) ToMap() *storage.OrderedMap {
	m := storage.NewOrderedMap()
	m.Set("published", r.Published)
	m.Set("skipped_reason", pyNoneIfEmpty(r.SkippedReason))
	m.Set("error", pyNoneIfEmpty(r.Error))
	m.Set("commit_sha", pyNoneIfEmpty(r.CommitSha))
	m.Set("commit_message", r.CommitMessage)
	m.Set("files_changed", r.FilesChanged)
	m.Set("pushed", r.Pushed)
	m.Set("repo_name", pyNoneIfEmpty(r.RepoName))
	m.Set("repo_url", pyNoneIfEmpty(r.RepoURL))
	m.Set("entries", r.Entries)
	m.Set("categories", r.Categories)
	m.Set("duration_ms", r.DurationMS)
	m.Set("dry_run", r.DryRun)
	m.Set("staging_dir", pyNoneIfEmpty(r.StagingDir))
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
// Staging lifetime
// ---------------------------------------------------------------------------

// Staging dirs created by real publishes are reused across Publish()
// calls of the same instance (that is what makes skip-when-unchanged
// work without a remote) and removed when the process exits. Dry-run
// dirs are NEVER removed — the CLI reports their path so a human can
// inspect the built directory.
//
// Go has no atexit: the reference's atexit.register maps to the
// exported CleanupStaging(), which the standalone CLI defers and the
// app wiring calls at shutdown (a documented convergence — the honest
// door instead of a silent leak).
var (
	stagingMu      sync.Mutex
	stagingCleanup []string
)

// registerStaging appends a staging dir to the process-exit cleanup
// list (mirrors _STAGING_CLEANUP.append).
func registerStaging(dir string) {
	stagingMu.Lock()
	defer stagingMu.Unlock()
	stagingCleanup = append(stagingCleanup, dir)
}

// unregisterStaging removes a staging dir from the cleanup list
// (mirrors the test helper's _STAGING_CLEANUP.remove).
func unregisterStaging(dir string) {
	stagingMu.Lock()
	defer stagingMu.Unlock()
	for i, d := range stagingCleanup {
		if d == dir {
			stagingCleanup = append(stagingCleanup[:i], stagingCleanup[i+1:]...)
			return
		}
	}
}

// CleanupStaging mirrors _cleanup_staging: remove every registered
// staging dir, best effort (ignore_errors=True). Idempotent; safe to
// call from the CLI's defer and the tests' cleanups.
func CleanupStaging() {
	stagingMu.Lock()
	dirs := stagingCleanup
	stagingCleanup = nil
	stagingMu.Unlock()
	for _, d := range dirs {
		_ = os.RemoveAll(d)
	}
}

// ---------------------------------------------------------------------------
// _git_env — the subprocess environment (ADR-0009; goodrepos is the
// law's origin — vaultseal's twin)
// ---------------------------------------------------------------------------

// gitEnv mirrors _git_env: strip repo-hijacking vars, never prompt
// (GIT_TERMINAL_PROMPT=0).
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
