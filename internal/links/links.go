// Package links ports app/gitcurator/core/links.py — THE single source
// of link parsing & normalization: GitHub extraction, website
// canonicalization, THE LAW blocked-domain lists, platform
// classification, self-domain detection, secret scrubbing, and the
// faithful import-file grammar.
//
// Spec: docs/port-blueprint/module-specs/spec-core-links.md;
// task T-11 (Phase 1). Parity is pinned by the generated golden in
// testdata/ (see links_test.go) — the Phase-1 "import-file grammar"
// differential in its spec-defined form.
package links

import (
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"
)

// ---------------------------------------------------------------------------
// Patterns (spec-core-links §Public symbols). RE2-safe translations of
// the Python `re` patterns (doc 05 §3 item 8).
// ---------------------------------------------------------------------------

var (
	// githubURLPattern mirrors GITHUB_URL_PATTERN (unanchored).
	githubURLPattern = regexp.MustCompile(`https?://(?:www\.)?github\.com/([a-zA-Z0-9\-_.]+)/([a-zA-Z0-9\-_.]+)`)

	// allLinksPattern mirrors ALL_LINKS_PATTERN.
	allLinksPattern = regexp.MustCompile(`https?://[^\s<>"\)\]]+`)

	// githubIOPattern mirrors GITHUB_IO_PATTERN.
	githubIOPattern = regexp.MustCompile(`https?://([a-zA-Z0-9\-_.]+)\.github\.io/([a-zA-Z0-9\-_.]+)`)

	// gistURLPattern mirrors GIST_URL_PATTERN (IGNORECASE).
	gistURLPattern = regexp.MustCompile(`(?i)https?://(?:www\.)?gist\.github\.com/`)

	// validRepoChars mirrors _VALID_REPO_CHARS (fullmatch; defeats
	// traversal payloads).
	validRepoChars = regexp.MustCompile(`^[a-zA-Z0-9](?:[a-zA-Z0-9\-_.]*[a-zA-Z0-9])?$`)

	// secretURLTextRe mirrors _SECRET_URL_TEXT_RE (IGNORECASE).
	secretURLTextRe = regexp.MustCompile(`(?i)([?&](?:token|secret|access_token|api_key|apikey|password|passcode|signature|sig|auth)=)([A-Za-z0-9_\-.%]{16,})`)

	// trackingParamRe mirrors _TRACKING_PARAM_RE (IGNORECASE). The
	// optional trailing \n mirrors Python `$` (matches before a final
	// newline) so re.match-key semantics line up.
	trackingParamRe = regexp.MustCompile(`(?i)^(utm(?:_[a-z0-9_]+)?|fbclid|gclid|ref|ref_src|ref_url|igshid|mc_cid|mc_eid)\n?$`)

	// bareAddressRe mirrors _BARE_ADDRESS_RE.
	bareAddressRe = regexp.MustCompile(`^(?:[a-zA-Z0-9][a-zA-Z0-9\-.]*\.[a-zA-Z]{2,}|\d{1,3}(?:\.\d{1,3}){3})(?::\d+)?(?:[/?#].*)?$`)
)

// trailingPunct mirrors _TRAILING_PUNCT (the rstrip character set:
// . , ) ; : ! ' ").
const trailingPunct = `.,);:!'"`

// matchAtStart mirrors Python re.match: the match must start at index 0
// (returns the submatch groups like FindStringSubmatch, nil otherwise).
func matchAtStart(re *regexp.Regexp, s string) []string {
	idx := re.FindStringSubmatchIndex(s)
	if idx == nil || idx[0] != 0 {
		return nil
	}
	out := make([]string, 0, len(idx)/2)
	for i := 0; i < len(idx); i += 2 {
		if idx[i] < 0 {
			out = append(out, "")
			continue
		}
		out = append(out, s[idx[i]:idx[i+1]])
	}
	return out
}

func validRepoName(s string) bool { return validRepoChars.MatchString(s) }

// ---------------------------------------------------------------------------
// Extraction
// ---------------------------------------------------------------------------

// ExtractGitHubURLs mirrors extract_github_urls.
func ExtractGitHubURLs(text string, dedupe ...bool) []string {
	dedupeOn := true
	if len(dedupe) > 0 {
		dedupeOn = dedupe[0]
	}
	if text == "" {
		return []string{}
	}
	urls := []string{}
	seen := map[string]bool{}
	for _, m := range githubURLPattern.FindAllStringSubmatch(text, -1) {
		url := "https://github.com/" + m[1] + "/" + m[2]
		if dedupeOn {
			if seen[url] {
				continue
			}
			seen[url] = true
		}
		urls = append(urls, url)
	}
	return urls
}

// ExtractAllLinks mirrors extract_all_links: trailing-punct rstrip
// (character-SET semantics), duplicates KEPT, order preserved.
func ExtractAllLinks(text string) []string {
	if text == "" {
		return []string{}
	}
	out := []string{}
	for _, lnk := range allLinksPattern.FindAllString(text, -1) {
		out = append(out, strings.TrimRight(lnk, trailingPunct))
	}
	return out
}

// MapGitHubIOURL mirrors map_github_io_url: ” when not a
// <owner>.github.io/<repo> page (a bare Pages site stays a website).
func MapGitHubIOURL(url string) string {
	if url == "" {
		return ""
	}
	m := matchAtStart(githubIOPattern, strings.TrimSpace(url))
	if m == nil {
		return ""
	}
	if !validRepoName(m[1]) || !validRepoName(m[2]) {
		return ""
	}
	return "https://github.com/" + m[1] + "/" + m[2]
}

// IsGistURL mirrors is_gist_url.
func IsGistURL(url string) bool {
	return url != "" && gistURLPattern.MatchString(strings.TrimSpace(url))
}

// SplitLinks mirrors split_links → (github, non-github, raw count).
func SplitLinks(text string) (githubURLs, nonGithub []string, rawCount int) {
	raw := ExtractAllLinks(text)
	githubURLs = []string{}
	nonGithub = []string{}
	seen := map[string]bool{}
	for _, link := range raw {
		if seen[link] {
			continue
		}
		seen[link] = true
		m := matchAtStart(githubURLPattern, link)
		if m != nil {
			canonical := "https://github.com/" + m[1] + "/" + m[2]
			if !contains(githubURLs, canonical) {
				githubURLs = append(githubURLs, canonical)
			}
			continue
		}
		ghIO := MapGitHubIOURL(link)
		if ghIO != "" {
			if !contains(githubURLs, ghIO) {
				githubURLs = append(githubURLs, ghIO)
			}
			continue
		}
		nonGithub = append(nonGithub, link)
	}
	return githubURLs, nonGithub, len(raw)
}

// IsGitHubURL mirrors is_github_url.
func IsGitHubURL(url string) bool {
	if url == "" {
		return false
	}
	m := matchAtStart(githubURLPattern, strings.TrimSpace(url))
	return m != nil && validRepoName(m[1]) && validRepoName(m[2])
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Normalization
// ---------------------------------------------------------------------------

// CleanURL mirrors clean_url.
func CleanURL(url string) string {
	u := strings.TrimSpace(url)
	if u != "" && !strings.HasPrefix(u, "http") {
		u = "https://" + u
	}
	return u
}

// NormalizeURL mirrors normalize_url (Normalizer A): twitter.com→x.com
// substring replace (anywhere — domain OR path), drop #fragment then
// ?query, rstrip ALL trailing '/', lowercase the domain ONLY (pyLower);
// path case preserved (GitHub paths are case-sensitive).
func NormalizeURL(url string) string {
	if url == "" {
		return ""
	}
	u := strings.TrimSpace(url)
	u = strings.ReplaceAll(u, "twitter.com", "x.com")
	if i := strings.IndexByte(u, '#'); i >= 0 {
		u = u[:i]
	}
	if i := strings.IndexByte(u, '?'); i >= 0 {
		u = u[:i]
	}
	if strings.HasSuffix(u, "/") && !strings.HasSuffix(u, "://") {
		u = strings.TrimRight(u, "/")
	}
	if i := strings.Index(u, "://"); i >= 0 {
		scheme, rest := u[:i], u[i+3:]
		if j := strings.IndexByte(rest, '/'); j >= 0 {
			domain, path := rest[:j], rest[j+1:]
			return scheme + "://" + pyLower(domain) + "/" + path
		}
		return scheme + "://" + pyLower(rest)
	}
	return u
}

// pyNetloc mirrors Python urlparse().netloc for the URL shapes this
// module processes (manual extraction — Go's net/url rejects host
// characters like spaces that Python's urlparse tolerates).
func pyNetloc(u string) string {
	i := strings.Index(u, "://")
	if i < 0 {
		return ""
	}
	rest := u[i+3:]
	if j := strings.IndexAny(rest, "/?#"); j >= 0 {
		return rest[:j]
	}
	return rest
}

// DomainOf mirrors domain_of: lowercased netloc (userinfo and port
// kept); scheme injected when missing — THE LAW cannot be bypassed by
// dropping the scheme.
func DomainOf(url string) string {
	u := strings.TrimSpace(url)
	if u == "" {
		return ""
	}
	if !strings.Contains(u, "://") {
		u = "https://" + u
	}
	return pyLower(pyNetloc(u))
}

// SiteKeyOf mirrors site_key_of: the host with a leading www. dropped —
// ONE NOTE PER SITE. Deliberately NOT the registrable domain.
func SiteKeyOf(url string) string {
	host := DomainOf(url)
	return strings.TrimPrefix(host, "www.")
}

// ---------------------------------------------------------------------------
// THE LAW — blocked domains (the 41-entry floor; config can only ADD)
// ---------------------------------------------------------------------------

// LawBlockedDomains mirrors LAW_BLOCKED_DOMAINS (order preserved).
var LawBlockedDomains = []string{
	// X / Twitter (the original v0.20.0 fix)
	"x.com", "twitter.com", "t.co",
	// GitHub — the whole group
	"github.com", "gist.github.com", "github.io", "githubusercontent.com",
	// HuggingFace
	"huggingface.co", "hf.co",
	// Instagram
	"instagram.com", "instagr.am",
	// Facebook
	"facebook.com", "fb.com", "fb.me", "fb.watch",
	// LinkedIn
	"linkedin.com", "lnkd.in",
	// v0.35.0 — YouTube
	"youtube.com", "youtu.be", "youtube-nocookie.com",
	// v0.35.0 — Google share/drive/docs family
	"share.google", "drive.google.com", "docs.google.com", "forms.google.com",
	// v0.35.0 — the social-media majors
	"tiktok.com",
	"threads.net", "threads.com",
	"snapchat.com",
	"pinterest.com", "pin.it",
	"twitch.tv",
	"discord.gg", "discord.com", "discordapp.com",
	"t.me", "telegram.me",
	"wa.me", "whatsapp.com",
	"vk.com",
	"bsky.app",
	"weibo.com",
}

// LawPlatformDomains mirrors LAW_PLATFORM_DOMAINS: the law minus the
// GitHub group (repo links must keep flowing to the GitHub pipeline).
var LawPlatformDomains = func() []string {
	isGitHubGroup := func(d string) bool {
		return d == "github.com" || strings.HasSuffix(d, ".github.com") ||
			d == "github.io" || strings.HasSuffix(d, ".github.io") ||
			d == "githubusercontent.com" || strings.HasSuffix(d, ".githubusercontent.com")
	}
	out := []string{}
	for _, d := range LawBlockedDomains {
		if !isGitHubGroup(d) {
			out = append(out, d)
		}
	}
	return out
}()

// DefaultSelfDomains mirrors DEFAULT_SELF_DOMAINS.
var DefaultSelfDomains = []string{"github-to-obsidian-bot.aliassadi-plus.workers.dev"}

// pyStr mirrors Python str() for the realistic scalar types a config
// list can carry (strings, ints, floats, bools; dicts/None never occur
// in real configs).
func pyStr(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "True"
		}
		return "False"
	case int:
		return strconv.Itoa(t)
	case float64:
		s := strconv.FormatFloat(t, 'g', -1, 64)
		if !strings.ContainsAny(s, ".eE") {
			s += ".0"
		}
		return s
	default:
		return ""
	}
}

// normalizeDomainEntry mirrors the per-entry normalization:
// str(item).strip().lower().lstrip('.').
func normalizeDomainEntry(item string) string {
	d := pyLower(strings.TrimSpace(item))
	return strings.TrimPrefix(d, ".")
}

// extraDomainsFromConfig mirrors _extra_domains_from_config (never
// raises; list/tuple or comma string; deduped in order).
func extraDomainsFromConfig(config map[string]any, key string) []string {
	var raw any
	if config != nil {
		raw = config[key]
	}
	var items []any
	switch v := raw.(type) {
	case nil:
		return []string{}
	case string:
		for _, it := range strings.Split(v, ",") {
			items = append(items, it)
		}
	case []any:
		items = v
	default:
		return []string{}
	}
	out, seen := []string{}, map[string]bool{}
	for _, item := range items {
		d := normalizeDomainEntry(pyStr(item))
		if d != "" && !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// mergeDomainLists mirrors _merge_domain_lists: base first (the law),
// then the extras; ordered dedupe union.
func mergeDomainLists(base, extras []string) []string {
	out, seen := []string{}, map[string]bool{}
	for _, d := range append(append([]string{}, base...), extras...) {
		e := normalizeDomainEntry(d)
		if e != "" && !seen[e] {
			seen[e] = true
			out = append(out, e)
		}
	}
	return out
}

// BlockedDomainsFromConfig mirrors blocked_domains_from_config: the LAW
// floor + config extras (extras can only add; empty = just the law).
func BlockedDomainsFromConfig(config map[string]any) []string {
	return mergeDomainLists(LawBlockedDomains, extraDomainsFromConfig(config, "web_blocked_domains"))
}

// PlatformDomainsFromConfig mirrors platform_domains_from_config: the
// LAW's social-platform subset + config extras.
func PlatformDomainsFromConfig(config map[string]any) []string {
	return mergeDomainLists(LawPlatformDomains, extraDomainsFromConfig(config, "web_blocked_domains"))
}

// SelfDomainsFromConfig mirrors self_domains_from_config: missing key →
// DEFAULT; empty value → opt-out (nothing is self).
func SelfDomainsFromConfig(config map[string]any) []string {
	var raw any
	if config != nil {
		raw = config["web_self_domains"]
	}
	var items []any
	switch v := raw.(type) {
	case nil:
		return append([]string{}, DefaultSelfDomains...)
	case string:
		for _, it := range strings.Split(v, ",") {
			items = append(items, it)
		}
	case []any:
		items = v
	default:
		return append([]string{}, DefaultSelfDomains...)
	}
	out, seen := []string{}, map[string]bool{}
	for _, item := range items {
		d := normalizeDomainEntry(pyStr(item))
		if d != "" && !seen[d] {
			seen[d] = true
			out = append(out, d)
		}
	}
	return out
}

// DomainIsBlocked mirrors domain_is_blocked: host (everything before the
// first colon — the Python port-strip quirk) == entry or
// endswith('.'+entry).
func DomainIsBlocked(url string, blockedDomains []string) bool {
	if len(blockedDomains) == 0 {
		return false
	}
	host := DomainOf(url)
	if host == "" {
		return false
	}
	host, _, _ = strings.Cut(host, ":")
	for _, entry := range blockedDomains {
		e := normalizeDomainEntry(entry)
		if e == "" {
			continue
		}
		if host == e || strings.HasSuffix(host, "."+e) {
			return true
		}
	}
	return false
}

// DomainIsSelf mirrors domain_is_self (same suffix-anchored matching).
func DomainIsSelf(url string, selfDomains []string) bool {
	return DomainIsBlocked(url, selfDomains)
}

// ---------------------------------------------------------------------------
// Platform recognition (v0.35.0) — first-match-wins, suffix-anchored
// ---------------------------------------------------------------------------

// PlatformEntry is one ordered PLATFORM_DOMAINS row.
type PlatformEntry struct {
	Key     string
	Domains []string
}

// PlatformDomains mirrors PLATFORM_DOMAINS in insertion order.
var PlatformDomains = []PlatformEntry{
	{"x_twitter", []string{"x.com", "twitter.com", "t.co"}},
	{"youtube", []string{"youtube.com", "youtu.be", "youtube-nocookie.com"}},
	{"reddit", []string{"reddit.com", "redd.it"}},
	{"linkedin", []string{"linkedin.com", "lnkd.in"}},
	{"medium", []string{"medium.com"}},
	{"huggingface", []string{"huggingface.co", "hf.co"}},
	{"arxiv", []string{"arxiv.org"}},
	{"package_registry", []string{"npmjs.com", "pypi.org"}},
}

// PlatformTableFiles mirrors PLATFORM_TABLE_FILES.
var PlatformTableFiles = map[string]string{
	"x_twitter":        "x_twitter_links.md",
	"youtube":          "youtube_links.md",
	"reddit":           "reddit_links.md",
	"linkedin":         "linkedin_links.md",
	"medium":           "medium_links.md",
	"huggingface":      "huggingface_links.md",
	"arxiv":            "arxiv_links.md",
	"package_registry": "package_registry_links.md",
	"other":            "other_links.md",
}

// ClassifyPlatform mirrors classify_platform.
func ClassifyPlatform(url string) string {
	host := DomainOf(url)
	if host == "" {
		return "other"
	}
	host, _, _ = strings.Cut(host, ":")
	if host == "github.com" || strings.HasSuffix(host, ".github.com") {
		return "github"
	}
	for _, p := range PlatformDomains {
		for _, d := range p.Domains {
			if host == d || strings.HasSuffix(host, "."+d) {
				return p.Key
			}
		}
	}
	return "other"
}

// ---------------------------------------------------------------------------
// Secret scrubbing (v0.21.0)
// ---------------------------------------------------------------------------

// ScrubURLToken mirrors scrub_url_token: secret-named query values
// (≥16 token-ish chars) → literal "…" (U+2026).
func ScrubURLToken(url string) string {
	raw := url
	if !strings.Contains(raw, "=") || !strings.Contains(raw, "?") {
		return raw
	}
	return secretURLTextRe.ReplaceAllString(raw, "${1}…")
}

// ScrubURLsInText mirrors scrub_urls_in_text.
func ScrubURLsInText(text string) string {
	if text == "" {
		return text
	}
	return secretURLTextRe.ReplaceAllString(text, "${1}…")
}

// ---------------------------------------------------------------------------
// Normalizer B — normalize_website_url (THE ONE SPELLING, v0.64.0)
// ---------------------------------------------------------------------------

// NormalizeWebsiteURL mirrors normalize_website_url.
func NormalizeWebsiteURL(url string) string {
	if url == "" {
		return ""
	}
	u := strings.TrimSpace(url)
	if i := strings.IndexByte(u, '#'); i >= 0 {
		u = u[:i]
	}
	netloc, path, query := pyURLParse(u)
	host := strings.TrimPrefix(pyLower(netloc), "www.")
	var kept []string
	if query != "" {
		for _, pair := range strings.Split(query, "&") {
			if pair == "" {
				continue
			}
			key, _, _ := strings.Cut(pair, "=")
			if key != "" && !trackingParamRe.MatchString(key) {
				kept = append(kept, pair)
			}
		}
	}
	switch {
	case path == "/" && len(kept) == 0:
		path = ""
	case path == "" && len(kept) > 0:
		path = "/"
	case strings.HasSuffix(path, "/") && len(path) > 1:
		path = strings.TrimRight(path, "/")
	}
	out := "https://" + host + path
	if len(kept) > 0 {
		out += "?" + strings.Join(kept, "&")
	}
	return out
}

// pyURLParse mirrors the urlparse pieces normalize_website_url uses
// (netloc/path/query; Python's tolerant semantics — the fragment is
// already stripped by the caller).
func pyURLParse(u string) (netloc, path, query string) {
	if i := strings.IndexByte(u, '#'); i >= 0 {
		u = u[:i]
	}
	rest := u
	if i := strings.Index(rest, "://"); i >= 0 {
		rest = rest[i+3:]
		if j := strings.IndexAny(rest, "/?#"); j >= 0 {
			netloc, rest = rest[:j], rest[j:]
		} else {
			return rest, "", ""
		}
	} // scheme-less: netloc stays "" and the whole string is the path
	if k := strings.IndexByte(rest, '?'); k >= 0 {
		return netloc, rest[:k], rest[k+1:]
	}
	return netloc, rest, ""
}

// DedupeURLs mirrors dedupe_urls: order-preserving dedup by NormalizeURL.
func DedupeURLs(urls []string) (unique []string, duplicates int) {
	seen := map[string]bool{}
	unique = []string{}
	for _, u := range urls {
		key := NormalizeURL(u)
		if seen[key] {
			continue
		}
		seen[key] = true
		unique = append(unique, u)
	}
	return unique, len(urls) - len(unique)
}

// ---------------------------------------------------------------------------
// Faithful batch imports (v0.29.0)
// ---------------------------------------------------------------------------

// ImportResult mirrors the parse_import_text result dict.
type ImportResult struct {
	// Field order is alphabetical by JSON name so a canonical dump of a
	// result matches Python json.dumps(sort_keys=True) byte-for-byte.
	Duplicates          int      `json:"duplicates"`
	GitHubURLs          []string `json:"github_urls"`
	RawCount            int      `json:"raw_count"`
	SkippedCommentLines int      `json:"skipped_comment_lines"`
	Unparsed            []string `json:"unparsed"`
	WebsiteURLs         []string `json:"website_urls"`
}

// ReadImportFile mirrors read_import_file: UTF-8 with BOM first, then
// latin-1 with errors='replace' (never raises for encoding).
func ReadImportFile(path string) (string, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}
	return decodeImportBytes(raw), nil
}

func decodeImportBytes(raw []byte) string {
	// Python: raw.decode('utf-8-sig') — succeeds iff the bytes are valid
	// UTF-8 (a leading BOM is stripped when present); on
	// UnicodeDecodeError → latin-1, errors='replace' (latin-1 maps every
	// byte, so nothing is ever replaced).
	if utf8.Valid(raw) {
		return strings.TrimPrefix(string(raw), "\ufeff")
	}
	var b strings.Builder
	b.Grow(len(raw))
	for _, by := range raw {
		b.WriteByte(by) // latin-1: every byte 0x00-0xFF is its own rune
	}
	return b.String()
}

// ParseImportText mirrors parse_import_text (the import grammar).
func ParseImportText(text string) ImportResult {
	res := ImportResult{GitHubURLs: []string{}, WebsiteURLs: []string{}, Unparsed: []string{}}
	seenGitHub := map[string]bool{}
	seenWebsite := map[string]bool{}

	route := func(link string) {
		res.RawCount++
		if m := matchAtStart(githubURLPattern, link); m != nil {
			canonical := "https://github.com/" + m[1] + "/" + m[2]
			if seenGitHub[canonical] {
				res.Duplicates++
				return
			}
			seenGitHub[canonical] = true
			res.GitHubURLs = append(res.GitHubURLs, canonical)
			return
		}
		if ghIO := MapGitHubIOURL(link); ghIO != "" {
			if seenGitHub[ghIO] {
				res.Duplicates++
				return
			}
			seenGitHub[ghIO] = true
			res.GitHubURLs = append(res.GitHubURLs, ghIO)
			return
		}
		key := NormalizeWebsiteURL(link)
		if seenWebsite[key] {
			res.Duplicates++
			return
		}
		seenWebsite[key] = true
		res.WebsiteURLs = append(res.WebsiteURLs, link)
	}

	for _, line := range splitLines(text) {
		line = strings.TrimSpace(line)
		line = strings.TrimLeft(line, "\ufeff")
		if line == "" || strings.HasPrefix(line, "#") {
			res.SkippedCommentLines++
			continue
		}
		found := ExtractAllLinks(line)
		if len(found) == 0 {
			if bareAddressRe.MatchString(line) {
				route(CleanURL(line))
			} else {
				res.Unparsed = append(res.Unparsed, line)
			}
			continue
		}
		for _, link := range found {
			route(link)
		}
	}
	return res
}

// splitLines mirrors Python str.splitlines(): splits on \n, \r\n (one
// boundary), \r, \v, \f, \x1c-\x1e, \x85, \u2028, \u2029; keeps empty
// interior lines; no trailing empty line.
func splitLines(s string) []string {
	var out []string
	start := 0
	i := 0
	for i < len(s) {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch {
		case r == '\r' && i+1 < len(s) && s[i+1] == '\n':
			out = append(out, s[start:i])
			i += 2
			start = i
		case r == '\n' || r == '\r' || r == '\v' || r == '\f' ||
			r == 0x1c || r == 0x1d || r == 0x1e ||
			r == 0x85 || r == 0x2028 || r == 0x2029:
			out = append(out, s[start:i])
			i += size
			start = i
		default:
			i += size
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
