// radar.go — the constants, the three vault slots, and RadarAPI (the
// real read-only GitHub transport; the tests inject the same-shape
// fake — the seam is the interface).
package reporadar

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// GITHUBAPI mirrors GITHUB_API.
const GITHUBAPI = "https://api.github.com"

// APITimeoutSec mirrors API_TIMEOUT = 15.0 (seconds).
const APITimeoutSec = 15.0

// MaxTreeFetches mirrors MAX_TREE_FETCHES = 24 — candidate repos
// fingerprinted, worst case (the v0.84.0 dev fix: was 12 — a real
// account's vault-ish names alone can crowd out the true backup).
const MaxTreeFetches = 24

// MaxSampleFiles mirrors MAX_SAMPLE_FILES — local files hashed for
// the fingerprint.
const MaxSampleFiles = 40

// MaxSampleBytes mirrors MAX_SAMPLE_BYTES — skip files larger than
// this when sampling.
const MaxSampleBytes = 512 * 1024

// MinDominantMatches mirrors MIN_DOMINANT_MATCHES — below this a
// fingerprint proves nothing.
const MinDominantMatches = 3

// DominanceFactor mirrors DOMINANCE_FACTOR — the winner must beat the
// runner-up by this factor.
const DominanceFactor = 2

// VaultSlot mirrors one VAULT_SLOTS row: (kind, config path key, repo
// slot key).
type VaultSlot struct {
	Kind    string
	PathKey string
	SlotKey string
}

// VaultSlots mirrors VAULT_SLOTS: the three vaults — config path key
// → repo slot (kind, setter path).
var VaultSlots = []VaultSlot{
	{"github", "vault_path", "vaultseal.repo_name"},
	{"websites", "website_vault_path", "website_repo_name"},
	{"personal", "manual_vault_path", "manual_repo_name"},
}

// vaultLabels mirrors _VAULT_LABELS.
var vaultLabels = map[string]string{
	"github": "GitHub vault", "websites": "Websites vault", "personal": "Personal vault",
}

// vaultishTokens mirrors _VAULTISH_TOKENS: repo names that could
// plausibly hold a vault — the candidate ordering hint for the
// fingerprint pass (detection quality never depends on it).
var vaultishTokens = []string{
	"vault", "obsidian", "directory", "library", "notes", "repos", "websites",
}

// API is the radar's GitHub surface — RadarAPI's exact call shape
// (the reference's tests inject a fake duck-typed to these four
// methods; the Go seam is the interface). Read-only by construction.
type API interface {
	// Me mirrors me(): the account login for the token, "" = None.
	Me() string
	// ListRepos mirrors list_repos(): the account's repositories (up
	// to 3 pages), newest-pushed first. Failures degrade to nil.
	ListRepos() []map[string]any
	// RepoExists mirrors repo_exists: the repo payload when it
	// exists, nil on 404, {"_error": …} when the check itself failed
	// (network / rate limit).
	RepoExists(name string) map[string]any
	// TreeBlobShas mirrors tree_blob_shas: the set of blob SHAs in a
	// repo tree (recursive), nil when the tree could not be read.
	TreeBlobShas(fullName, ref string) map[string]bool
}

// RadarAPI mirrors the RadarAPI class: the minimal read-only GitHub
// client for the radar. Never errors — failures come back as nil /
// empty with the reason spoken nowhere but the caller's report.
type RadarAPI struct {
	token   string
	timeout float64
	base    string
}

// NewRadarAPI mirrors __init__: token strip; timeout 0 → the 15 s
// default; base "" → api.github.com (trailing "/" dropped).
func NewRadarAPI(token string, timeout float64, base string) *RadarAPI {
	if timeout == 0 {
		timeout = APITimeoutSec
	}
	if base == "" {
		base = GITHUBAPI
	}
	return &RadarAPI{
		token:   strings.TrimSpace(token),
		timeout: timeout,
		base:    strings.TrimRight(base, "/"),
	}
}

// get mirrors _get: GET with the radar headers. Returns (status,
// payload); network failures answer (0, {}) — never errors.
func (r *RadarAPI) get(path string) (int, any) {
	if r.token == "" {
		return 0, map[string]any{"message": "no token"}
	}
	req, err := http.NewRequest("GET", r.base+path, nil)
	if err != nil {
		return 0, map[string]any{}
	}
	req.Header.Set("Authorization", "Bearer "+r.token)
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("X-GitHub-Api-Version", "2022-11-28")
	req.Header.Set("User-Agent", "GitCurator-RepoRadar")
	client := &http.Client{Timeout: time.Duration(r.timeout * float64(time.Second))}
	resp, err := client.Do(req)
	if err != nil {
		return 0, map[string]any{}
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return 0, map[string]any{}
	}
	body := "{}"
	if len(raw) > 0 {
		body = string(raw)
	}
	var payload any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		return 0, map[string]any{}
	}
	return resp.StatusCode, payload
}

// Me mirrors me(): the account login for the token, or "" (None).
func (r *RadarAPI) Me() string {
	status, payload := r.get("/user")
	if status != 200 {
		return ""
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return ""
	}
	login, _ := m["login"].(string)
	return login
}

// ListRepos mirrors list_repos: the account's repositories (up to 3
// pages), newest-pushed first. Failures degrade to nil.
func (r *RadarAPI) ListRepos() []map[string]any {
	var out []map[string]any
	for page := 1; page <= 3; page++ {
		status, payload := r.get("/user/repos?per_page=100&sort=pushed&page=" +
			itoa(page))
		if status != 200 {
			break
		}
		list, ok := payload.([]any)
		if !ok {
			break
		}
		for _, item := range list {
			if m, ok := item.(map[string]any); ok {
				out = append(out, m)
			}
		}
		if len(list) < 100 {
			break
		}
	}
	return out
}

// RepoExists mirrors repo_exists: the repo payload when it exists,
// nil on 404, {"_error": …} when the check itself failed.
func (r *RadarAPI) RepoExists(name string) map[string]any {
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	status, payload := r.get("/repos/" + name)
	if status == 200 {
		if m, ok := payload.(map[string]any); ok {
			return m
		}
	}
	if status == 404 {
		return nil
	}
	if m, ok := payload.(map[string]any); ok {
		msg := ""
		if v, ok := m["message"]; ok && v != nil {
			msg = helpers.PyStr(v)
		}
		if msg == "" {
			msg = "HTTP " + itoa(status)
		}
		return map[string]any{"_error": msg}
	}
	return map[string]any{"_error": "HTTP " + itoa(status)}
}

// TreeBlobShas mirrors tree_blob_shas: the set of blob SHAs in a repo
// tree (recursive), or nil when the tree could not be read. A
// truncated tree still answers with what it holds — dominance, not
// completeness, is the criterion. (The reference's str(None) → "None"
// quirk for a blob entry without a sha is preserved via PyStr.)
func (r *RadarAPI) TreeBlobShas(fullName, ref string) map[string]bool {
	status, payload := r.get("/repos/" + fullName + "/git/trees/" + ref + "?recursive=1")
	if status != 200 {
		return nil
	}
	m, ok := payload.(map[string]any)
	if !ok {
		return nil
	}
	tree, ok := m["tree"].([]any)
	if !ok {
		return nil
	}
	out := map[string]bool{}
	for _, item := range tree {
		t, ok := item.(map[string]any)
		if !ok {
			continue
		}
		if typ, _ := t["type"].(string); typ == "blob" {
			out[helpers.PyStr(t["sha"])] = true
		}
	}
	return out
}

func itoa(n int) string {
	return helpers.PyStr(n)
}
