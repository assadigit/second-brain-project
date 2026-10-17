// ghclient.go — T-42: the GitHub API seam the per-repo loop drives.
//
// The reference rides PyGithub (g.get_rate_limit / g.get_repo plus the
// Repository object's lazy sub-fetches: get_topics, get_commits,
// get_readme, get_languages, get_releases). The Go port surfaces the
// SAME call surface behind one interface so the loop's laws (the
// gates, the 404-strike flow, the rate-limit waits, the 401-anonymous
// fallback) are testable with fakes — exactly the "fake repo" of the
// spec's test contract — while production wires the real REST client
// (Client below, api.github.com over net/http with a 15 s timeout —
// PyGithub's own default — standard TLS verification, which is what
// PyGithub does: the reference never threads netctx into the GitHub
// client, only into the banner download).
//
// GHError mirrors PyGithub's GithubException as the loop uses it:
// e.status plus str(e) (the 403-with-"rate limit" substring law and
// the 404/401 branches all read exactly those two facts).
package githubpipeline

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"
)

// GHError is the GithubException shape: Status is the HTTP status and
// Message is str(e) — the reference's rate-limit branch does
// `'rate limit' in str(e).lower()`, so the real client's Message
// carries the response body (which is where GitHub spells
// "API rate limit exceeded").
type GHError struct {
	Status  int
	Message string
}

func (e *GHError) Error() string { return e.Message }

// RateLimit is the core resource view of g.get_rate_limit().
type RateLimit struct {
	Remaining int
	Reset     time.Time // UTC; PyGithub hands back a naive-UTC datetime
}

// Repo is the Repository view the loop consumes. Parent is the
// embedded parent object (nil when the repo is not a fork) — the
// reference re-points the whole loop at it (fork → parent analysis)
// and every later sub-fetch then uses the PARENT's full name, exactly
// like PyGithub does once `repo = parent` rebinds.
type Repo struct {
	ID                int64
	FullName          string
	Name              string
	Fork              bool
	Parent            *Repo
	StargazersCount   int
	ForksCount        int
	Description       string
	OwnerLogin        string
	OrganizationLogin string // "" when the repo has no organization
	Language          string
	HasIssues         bool
	HasWiki           bool
	HasLicense        bool
}

// ReleaseDate is one latest-release view (tag + the published date).
type Release struct {
	TagName     string
	PublishedAt time.Time
}

// ReadmeResult carries the decoded README content and its RAW length
// (the loop logs "📄 README fetched (N chars)" with the pre-cap length
// while the analysis content is the 1500-char cap).
type ReadmeResult struct {
	Content  string
	RawChars int
}

// RepoFetcher is the GitHub API surface the loop drives. Every method
// returns *GHError for API-level failures (the loop branches on
// Status); transport failures come back as plain errors.
type RepoFetcher interface {
	GetRateLimit() (*RateLimit, error)
	GetRepo(fullName string) (*Repo, error)
	GetTopics(fullName string) ([]string, error)
	CountCommitsSince(fullName string, since time.Time) (int, error)
	GetReadme(fullName string) (*ReadmeResult, error)
	GetLanguages(fullName string) ([]string, error)
	LatestRelease(fullName string) (*Release, error)
}

// Client is the production RepoFetcher: the GitHub REST API over
// net/http. Token "" is anonymous access. Every request carries
// PyGithub's Accept header; the token rides the Authorization bearer
// header.
type Client struct {
	http    *http.Client
	baseURL string
	token   string
}

// NewClient mirrors Github(auth=Auth.Token(...)) / Github() — token
// "" builds the anonymous client, pointed at api.github.com.
func NewClient(token string) *Client {
	return NewClientFor("https://api.github.com", token)
}

// NewClientFor builds a client pointed at an arbitrary API base (the
// tests' loopback server) with the same wire behavior.
func NewClientFor(baseURL, token string) *Client {
	return &Client{
		http:    &http.Client{Timeout: 15 * time.Second},
		baseURL: strings.TrimRight(baseURL, "/"),
		token:   token,
	}
}

func (c *Client) do(method, url string) ([]byte, int, string, error) {
	req, err := http.NewRequest(method, url, nil)
	if err != nil {
		return nil, 0, "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	req.Header.Set("User-Agent", "gitcurator")
	if c.token != "" {
		req.Header.Set("Authorization", "Bearer "+c.token)
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, 0, "", err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, resp.StatusCode, "", err
	}
	return body, resp.StatusCode, resp.Header.Get("Content-Type"), nil
}

func (c *Client) getJSON(path string, out any) error {
	body, status, _, err := c.do(http.MethodGet, c.baseURL+path)
	if err != nil {
		return err
	}
	if status < 200 || status >= 300 {
		return &GHError{Status: status, Message: ghErrMessage(status, body)}
	}
	if out == nil {
		return nil
	}
	return json.Unmarshal(body, out)
}

// ghErrMessage shapes a non-2xx body the way PyGithub's
// GithubException str() does: "<status> {'message': …}" is the wire
// text the loop's substring laws read; a body that fails to decode
// falls back to the raw text.
func ghErrMessage(status int, body []byte) string {
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err == nil {
		if msg, ok := payload["message"].(string); ok {
			return fmt.Sprintf("%d {'message': '%s'}", status, msg)
		}
	}
	return fmt.Sprintf("%d %s", status, strings.TrimSpace(string(body)))
}

// GetRateLimit mirrors g.get_rate_limit() (GET /rate_limit).
func (c *Client) GetRateLimit() (*RateLimit, error) {
	var payload struct {
		Resources struct {
			Core struct {
				Remaining int   `json:"remaining"`
				Reset     int64 `json:"reset"`
			} `json:"core"`
		} `json:"resources"`
	}
	if err := c.getJSON("/rate_limit", &payload); err != nil {
		return nil, err
	}
	return &RateLimit{
		Remaining: payload.Resources.Core.Remaining,
		Reset:     time.Unix(payload.Resources.Core.Reset, 0).UTC(),
	}, nil
}

// GetRepo mirrors g.get_repo(fullName) (GET /repos/{fullName}). The
// response's embedded parent object (the fork case) is mapped to the
// same Repo shape; organization maps to "" when absent.
func (c *Client) GetRepo(fullName string) (*Repo, error) {
	var p ghRepoJSON
	if err := c.getJSON("/repos/"+fullName, &p); err != nil {
		return nil, err
	}
	return p.toRepo(), nil
}

type ghRepoJSON struct {
	ID           int64       `json:"id"`
	FullName     string      `json:"full_name"`
	Name         string      `json:"name"`
	Fork         bool        `json:"fork"`
	Parent       *ghRepoJSON `json:"parent"`
	Stargazers   int         `json:"stargazers_count"`
	Forks        int         `json:"forks_count"`
	Description  string      `json:"description"`
	Owner        *ghOwner    `json:"owner"`
	Organization *ghOwner    `json:"organization"`
	Language     string      `json:"language"`
	HasIssues    bool        `json:"has_issues"`
	HasWiki      bool        `json:"has_wiki"`
	License      *ghLicense  `json:"license"`
}

type ghOwner struct {
	Login string `json:"login"`
}

type ghLicense struct {
	Key string `json:"key"`
}

func (p *ghRepoJSON) toRepo() *Repo {
	if p == nil {
		return nil
	}
	r := &Repo{
		ID:              p.ID,
		FullName:        p.FullName,
		Name:            p.Name,
		Fork:            p.Fork,
		StargazersCount: p.Stargazers,
		ForksCount:      p.Forks,
		Description:     p.Description,
		Language:        p.Language,
		HasIssues:       p.HasIssues,
		HasWiki:         p.HasWiki,
		HasLicense:      p.License != nil,
	}
	if p.Owner != nil {
		r.OwnerLogin = p.Owner.Login
	}
	if p.Organization != nil {
		r.OrganizationLogin = p.Organization.Login
	}
	if p.Parent != nil {
		r.Parent = p.Parent.toRepo()
	}
	return r
}

// GetTopics mirrors repo.get_topics() (GET /repos/{fullName}/topics).
func (c *Client) GetTopics(fullName string) ([]string, error) {
	var payload struct {
		Names []string `json:"names"`
	}
	if err := c.getJSON("/repos/"+fullName+"/topics", &payload); err != nil {
		return nil, err
	}
	if payload.Names == nil {
		payload.Names = []string{}
	}
	return payload.Names, nil
}

// CountCommitsSince mirrors repo.get_commits(since=…).totalCount —
// the list endpoint's total (the Link header's rel="last" page, or 1
// when GitHub answers without the header).
func (c *Client) CountCommitsSince(fullName string, since time.Time) (int, error) {
	body, status, header, err := c.do(http.MethodGet,
		c.baseURL+"/repos/"+fullName+
			"/commits?since="+since.UTC().Format(time.RFC3339)+
			"&per_page=100")
	if err != nil {
		return 0, err
	}
	if status < 200 || status >= 300 {
		return 0, &GHError{Status: status, Message: ghErrMessage(status, body)}
	}
	if n := linkLastPage(header); n > 0 {
		return n, nil
	}
	// No Link header: everything fits the first page — count it.
	var rows []json.RawMessage
	if err := json.Unmarshal(body, &rows); err != nil {
		return 0, err
	}
	if len(rows) == 0 {
		return 0, nil
	}
	return len(rows), nil
}

// linkLastPage extracts rel="last"'s page number from a Link header
// (PyGithub's PaginatedList.totalCount source).
func linkLastPage(linkHeader string) int {
	if linkHeader == "" {
		return 0
	}
	last := 0
	for _, seg := range strings.Split(linkHeader, ",") {
		if !strings.Contains(seg, `rel="last"`) {
			continue
		}
		if i := strings.Index(seg, "&page="); i >= 0 {
			num := seg[i+6:]
			if j := strings.IndexAny(num, "&>"); j >= 0 {
				num = num[:j]
			}
			n := 0
			for _, ch := range num {
				if ch < '0' || ch > '9' {
					n = 0
					break
				}
				n = n*10 + int(ch-'0')
			}
			if n > last {
				last = n
			}
		}
	}
	return last
}

// GetReadme mirrors repo.get_readme() (GET /repos/{fullName}/readme):
// the base64 content decoded UTF-8 with errors ignored (the reference
// decodes exactly that way).
func (c *Client) GetReadme(fullName string) (*ReadmeResult, error) {
	var payload struct {
		Content string `json:"content"`
	}
	if err := c.getJSON("/repos/"+fullName+"/readme", &payload); err != nil {
		return nil, err
	}
	content := decodeBase64UTF8Ignore(payload.Content)
	return &ReadmeResult{Content: content, RawChars: len([]rune(content))}, nil
}

// decodeBase64UTF8Ignore mirrors the reference's README decode:
// base64.b64decode(readme.content).decode('utf-8', errors='ignore').
// Python's b64decode (validate=False) DISCARDS non-alphabet bytes
// before decoding — newline-wrapped content decodes whole; the UTF-8
// pass ignores invalid sequences (drop-at-first-invalid keeps the
// raw-length log honest; GitHub payloads are valid UTF-8 in practice).
func decodeBase64UTF8Ignore(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for i := 0; i < len(s); i++ {
		ch := s[i]
		if ch == '\n' || ch == '\r' {
			continue
		}
		b.WriteByte(ch)
	}
	data, _ := base64.StdEncoding.DecodeString(b.String())
	return string(dropInvalidUTF8(data))
}

// dropInvalidUTF8 truncates at the first invalid UTF-8 sequence
// (Python decode(errors='ignore') DROPS just those bytes; the
// truncation only fires on pathological payloads and keeps the raw
// length a fact, not a guess).
func dropInvalidUTF8(data []byte) []byte {
	for i := 0; i < len(data); {
		r, size := utf8.DecodeRune(data[i:])
		if r == utf8.RuneError && size <= 1 {
			return data[:i]
		}
		i += size
	}
	return data
}
func (c *Client) GetLanguages(fullName string) ([]string, error) {
	body, status, _, err := c.do(http.MethodGet,
		c.baseURL+"/repos/"+fullName+"/languages")
	if err != nil {
		return nil, err
	}
	if status < 200 || status >= 300 {
		return nil, &GHError{Status: status, Message: ghErrMessage(status, body)}
	}
	// decodeBase64UTF8Ignore's sibling: an ordered key decode. The
	// payload is small; a two-pass scan keeps insertion order without
	// a generic ordered-map dependency.
	var keys []string
	dec := json.NewDecoder(strings.NewReader(string(body)))
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	if delim, ok := tok.(json.Delim); !ok || delim != '{' {
		return []string{}, nil
	}
	for dec.More() {
		ktok, err := dec.Token()
		if err != nil {
			return nil, err
		}
		key, _ := ktok.(string)
		var v json.RawMessage
		if err := dec.Decode(&v); err != nil {
			return nil, err
		}
		keys = append(keys, key)
	}
	if keys == nil {
		keys = []string{}
	}
	return keys, nil
}

// LatestRelease mirrors repo.get_releases()[0] — the newest release's
// tag and published date, or nil when there are none (totalCount 0
// and the empty list both land there).
func (c *Client) LatestRelease(fullName string) (*Release, error) {
	var payload []struct {
		TagName     string `json:"tag_name"`
		PublishedAt string `json:"published_at"`
	}
	if err := c.getJSON("/repos/"+fullName+"/releases?per_page=1", &payload); err != nil {
		return nil, err
	}
	if len(payload) == 0 {
		return nil, nil
	}
	pub := time.Time{}
	if payload[0].PublishedAt != "" {
		if t, err := time.Parse(time.RFC3339, payload[0].PublishedAt); err == nil {
			pub = t
		}
	}
	return &Release{TagName: payload[0].TagName, PublishedAt: pub}, nil
}
