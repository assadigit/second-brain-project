// Package notebuilder ports app/gitcurator/core/note_builder.py — builds
// the GitHub-pipeline Obsidian note (YAML frontmatter + body) from
// LLM/GitHub data, treating EVERY value as hostile YAML input (the v30
// sanitization fix), plus the _missing/ 404 placeholder note. The
// sanitizers are reused later by the websites pipeline.
//
// Spec: docs/port-blueprint/module-specs/spec-core-note-builder.md;
// task T-15 (Phase 1). Parity is pinned by the generated golden in
// testdata/ (see notebuilder_test.go): full note formats for 6 synthetic
// repos, the sanitizer battery, the banner-gate traversal corpus, and the
// missing-repo notes.
//
// Port notes (documented structural decisions):
//   - Numeric fields (Stars, Forks, …) are int: GitHub returns ints and
//     the spec's migration note blesses the int field. List-or-scalar
//     inputs the LLM can misshape (Tags, Features) stay `any` (nil,
//     string, []string, []any) exactly like Python's isinstance paths;
//     json.Number values keep Python str() semantics via prompts.PyStr.
//   - Python's `\s` in the collapse regexes is UNICODE whitespace (nbsp,
//     U+2028 … collapse to a space) — pyWhitespaceCollapse reproduces it;
//     Go's ASCII-only `\s` would diverge (golden-pinned).
//   - `f"{stars:,}"` comma grouping rides dryrun.FormatComma (promoted
//     export; T-17's helpers package centralizes it, the formatComma
//     promotion law).
//   - date_processed / the missing-note dates are LOCAL dates
//     (datetime.now().strftime("%Y-%m-%d")); the golden normalizes them
//     to <DATE> and TestDateProcessedFormat pins the real format.
package notebuilder

import (
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
)

// ---------------------------------------------------------------------------
// Caps (note_builder.py:48-54).
// ---------------------------------------------------------------------------

// maxSeqItem mirrors _MAX_SEQ_ITEM (per tag/alias/language).
const maxSeqItem = 60

// MaxTags/MaxAliases/MaxLanguages/MaxScalar/MaxSummary mirror the caps.
const (
	MaxTags      = 12
	MaxAliases   = 4
	MaxLanguages = 8
	MaxScalar    = 120
	MaxSummary   = 2000
)

// seqItemOK mirrors _SEQ_ITEM_OK: the allowed charset inside a YAML flow
// sequence — no brackets, braces, commas, colons, quotes, hashes,
// backslashes or control characters (nothing that can escape the
// sequence or start a comment). Disallowed runes are replaced by a space.
var seqItemOK = regexp.MustCompile(`[^A-Za-z0-9 \-_.+/()]`)

// wsRunRE matches runs of ASCII whitespace (the post-charset-replacement
// corpus — non-ASCII whitespace was already substituted away).
var wsRunRE = regexp.MustCompile(`\s+`)

// bodyCtrlRE mirrors the body control-char strip [\x00-\x08\x0b-\x1f\x7f]
// (\n and \t survive).
var bodyCtrlRE = regexp.MustCompile("[\x00-\x08\x0b-\x1f\x7f]")

// bannerSafeRE mirrors the cover/banner fullmatch gate.
var bannerSafeRE = regexp.MustCompile(`^[A-Za-z0-9\-_.]+$`)

// ---------------------------------------------------------------------------
// Sanitizers (note_builder.py:57-167).
// ---------------------------------------------------------------------------

// SanitizeSeqItem mirrors sanitize_seq_item: disallowed → space,
// whitespace collapse, strip, strip '-' ends, truncate, strip. Returns ""
// for None/when nothing safe is left. Non-string values stringify with
// Python str() semantics (prompts.PyStr).
func SanitizeSeqItem(value any, maxLen int) string {
	if value == nil {
		return ""
	}
	text := helpers.PyStr(value)
	text = seqItemOK.ReplaceAllString(text, " ")
	text = helpers.CollapseWhitespace(text)
	text = strings.Trim(text, " ") // .strip() — spaces only remain
	text = strings.Trim(text, "-") // .strip('-')
	text = strings.Trim(text, " ") // .strip() again
	if len([]rune(text)) > maxLen {
		text = string([]rune(text)[:maxLen])
	}
	return strings.Trim(text, " ")
}

// SanitizeTags mirrors sanitize_tags: sanitize each item, case-
// insensitively dedupe (first casing kept — the post-charset corpus is
// ASCII, so strings.ToLower equals Python lower()), drop empties, cap.
// A string input becomes a one-item list; nil → nil.
func SanitizeTags(tags any) []string {
	if tags == nil {
		return nil
	}
	var items []any
	switch vv := tags.(type) {
	case string:
		items = []any{vv}
	case []string:
		items = make([]any, len(vv))
		for i, s := range vv {
			items[i] = s
		}
	case []any:
		items = vv
	default:
		items = []any{vv}
	}
	seen := map[string]bool{}
	var result []string
	for _, t := range items {
		clean := SanitizeSeqItem(t, maxSeqItem)
		if clean == "" {
			continue
		}
		key := strings.ToLower(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, clean)
		if len(result) >= MaxTags {
			break
		}
	}
	return result
}

// SanitizeTagsMax mirrors sanitize_tags(tags, max_items=N) — the
// reference's own parameterized cap (website_pipeline passes 5/6/8;
// the house SanitizeTags keeps the default 12). Same dedupe law:
// case-insensitive, first casing kept, cap applied to the DEDUPED
// result (a post-hoc slice of SanitizeTags' output is equivalent, but
// this names the reference's own shape). T-40's addition.
func SanitizeTagsMax(tags any, maxItems int) []string {
	if maxItems <= 0 {
		maxItems = MaxTags
	}
	if tags == nil {
		return nil
	}
	var items []any
	switch vv := tags.(type) {
	case string:
		items = []any{vv}
	case []string:
		items = make([]any, len(vv))
		for i, s := range vv {
			items[i] = s
		}
	case []any:
		items = vv
	default:
		items = []any{vv}
	}
	seen := map[string]bool{}
	var result []string
	for _, t := range items {
		clean := SanitizeSeqItem(t, maxSeqItem)
		if clean == "" {
			continue
		}
		key := strings.ToLower(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, clean)
		if len(result) >= maxItems {
			break
		}
	}
	return result
}

// SanitizeAliases mirrors sanitize_aliases: ≤80 chars per item, ≤4 items,
// case-insensitive dedupe.
func SanitizeAliases(names ...any) []string {
	seen := map[string]bool{}
	var result []string
	for _, n := range names {
		clean := SanitizeSeqItem(n, 80)
		if clean == "" {
			continue
		}
		key := strings.ToLower(clean)
		if seen[key] {
			continue
		}
		seen[key] = true
		result = append(result, clean)
		if len(result) >= MaxAliases {
			break
		}
	}
	return result
}

// SanitizeLanguages mirrors sanitize_languages: ≤40 chars per item, ≤8
// items, EXACT (case-sensitive) dedupe.
func SanitizeLanguages(languages any) []string {
	if languages == nil {
		return nil
	}
	var items []any
	switch vv := languages.(type) {
	case []string:
		items = make([]any, len(vv))
		for i, s := range vv {
			items[i] = s
		}
	case []any:
		items = vv
	default:
		items = []any{vv}
	}
	var out []string
	for _, lang := range items {
		clean := SanitizeSeqItem(lang, 40)
		if clean != "" && !contains(out, clean) {
			out = append(out, clean)
		}
		if len(out) >= MaxLanguages {
			break
		}
	}
	return out
}

func contains(list []string, s string) bool {
	for _, e := range list {
		if e == s {
			return true
		}
	}
	return false
}

// YamlScalar mirrors yaml_scalar: None → '""'; \r and \n → space; Unicode
// whitespace collapse; strip; truncate at maxLen; escape backslashes
// FIRST, then double quotes; ALWAYS double-quoted.
func YamlScalar(value any, maxLen int) string {
	if value == nil {
		return `""`
	}
	text := helpers.PyStr(value)
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = helpers.CollapseWhitespace(text)
	text = strings.Trim(text, " ")
	if len([]rune(text)) > maxLen {
		text = string([]rune(text)[:maxLen])
	}
	text = strings.ReplaceAll(text, `\`, `\\`)
	text = strings.ReplaceAll(text, `"`, `\"`)
	return `"` + text + `"`
}

// SanitizeBodyText mirrors sanitize_body_text: CRLF→LF, lone \r→LF, strip
// control chars (keep \n \t), strip, truncate (no re-strip). Body text is
// plain markdown; the frontmatter sits above it, so it cannot re-open
// YAML — but runaway lengths are capped.
func SanitizeBodyText(value any, maxLen int) string {
	if value == nil {
		return ""
	}
	text := helpers.PyStr(value)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	text = bodyCtrlRE.ReplaceAllString(text, "")
	text = strings.Trim(text, " \t\n\r\v\f") // str.strip() set (post-CRLF corpus)
	if maxLen > 0 {
		if len([]rune(text)) > maxLen {
			text = string([]rune(text)[:maxLen])
		}
	}
	return text
}

// SanitizeShortSummary mirrors sanitize_short_summary: one line (the
// Unicode collapse), capped at 2000.
func SanitizeShortSummary(value any) string {
	if value == nil {
		return ""
	}
	text := helpers.PyStr(value)
	text = strings.ReplaceAll(text, "\r", " ")
	text = strings.ReplaceAll(text, "\n", " ")
	text = helpers.CollapseWhitespace(text)
	text = strings.Trim(text, " ")
	if len([]rune(text)) > MaxSummary {
		text = string([]rune(text)[:MaxSummary])
	}
	return text
}

// ---------------------------------------------------------------------------
// Reputation / rating helpers (note_builder.py:174-193).
// ---------------------------------------------------------------------------

// RepToStr mirrors rep_to_str.
func RepToStr(rep int) string {
	switch {
	case rep >= 8:
		return "High (Major tech company)"
	case rep >= 6:
		return "Medium (Well-known organization)"
	case rep >= 4:
		return "Low (Individual or unknown)"
	default:
		return "Unknown"
	}
}

// ScoreToRating mirrors score_to_rating.
func ScoreToRating(score int) string {
	switch {
	case score >= 90:
		return "Excellent"
	case score >= 75:
		return "Good"
	case score >= 60:
		return "Average"
	default:
		return "Low"
	}
}

// ---------------------------------------------------------------------------
// Note builder (note_builder.py:200-345).
// ---------------------------------------------------------------------------

// NoteArgs mirrors build_note's keyword arguments. The list-or-scalar
// fields the LLM can misshape stay any; numeric fields are int (GitHub
// API contract — spec migration note).
type NoteArgs struct {
	URL               string `json:"url"`
	RepoName          string `json:"repo_name"`
	Owner             string `json:"owner"`
	OrgName           string `json:"org_name"`
	Stars             int    `json:"stars"`
	Forks             int    `json:"forks"`
	CommitCount       int    `json:"commit_count"`
	CredScore         int    `json:"cred_score"`
	OrgRep            int    `json:"org_rep"`
	Summary           string `json:"summary"`
	Tags              any    `json:"tags"`
	CategoryKey       string `json:"category_key"`
	Confidence        int    `json:"confidence"`
	HowItWorks        string `json:"how_it_works"`
	CoreValue         string `json:"core_value"`
	Features          any    `json:"features"`
	Difference        string `json:"difference"`
	BannerPath        string `json:"banner_path"`
	PrimaryLanguage   string `json:"primary_language"`
	Languages         any    `json:"languages"`
	ShortSummary      string `json:"short_summary"`
	LatestReleaseDate string `json:"latest_release_date"`
	// QualityIssues nil = auto-detect (Python None); a non-nil slice —
	// even empty — means the caller supplied them and auto-detect is
	// skipped, exactly like Python's `if quality_issues is None`.
	QualityIssues []string `json:"quality_issues"`
	IsLowQuality  bool     `json:"is_low_quality"`
}

// bannerRefs mirrors the banner security gate: backslashes fold to '/',
// the last segment must fullmatch [A-Za-z0-9\-_.]+ or the reference is
// dropped entirely. (".." fullmatches — benign, the reference stays
// inside attachments/banners/; pinned as-is by the golden.)
func bannerRefs(bannerPath string) (coverLine, section string) {
	if bannerPath == "" {
		return "", ""
	}
	normalized := strings.ReplaceAll(bannerPath, "\\", "/")
	rawName := normalized
	if i := strings.LastIndex(normalized, "/"); i >= 0 {
		rawName = normalized[i+1:]
	}
	if !bannerSafeRE.MatchString(rawName) {
		return "", ""
	}
	return "cover: attachments/banners/" + rawName,
		"![banner](attachments/banners/" + rawName + ")\n\n"
}

// localDateString mirrors datetime.now().strftime("%Y-%m-%d") — the LOCAL
// date, no timezone.
func localDateString() string { return time.Now().Format("2006-01-02") }

// BuildNote mirrors build_note: the full note content (frontmatter +
// markdown body) with every YAML-bound value routed through the
// sanitizers. The exact key order, the blank-line quirks (an absent
// optional field emits its template line as a blank line), and the
// aliases block's trailing newline are pinned byte-exact by the golden.
func BuildNote(args NoteArgs) string {
	tags := SanitizeTags(args.Tags)
	if tags == nil {
		tags = []string{}
	}
	aliases := SanitizeAliases(args.RepoName, args.Owner+"/"+args.RepoName)
	languages := SanitizeLanguages(args.Languages)
	if languages == nil {
		languages = []string{}
	}
	orgYAML := YamlScalar(args.OrgName, MaxScalar)
	urlYAML := YamlScalar(args.URL, MaxScalar)
	categoryYAML := YamlScalar(args.CategoryKey, MaxScalar)
	langScalar := SanitizeSeqItem(args.PrimaryLanguage, 40)
	releaseScalar := SanitizeSeqItem(args.LatestReleaseDate, 32)

	summary := SanitizeBodyText(args.Summary, 0)
	howItWorks := SanitizeBodyText(args.HowItWorks, 0)
	coreValue := SanitizeBodyText(args.CoreValue, 0)
	difference := SanitizeBodyText(args.Difference, 0)
	shortSummary := SanitizeShortSummary(args.ShortSummary)
	repoTitle := SanitizeShortSummary(args.RepoName)
	if len(repoTitle) > 120 {
		repoTitle = repoTitle[:120]
	}
	if repoTitle == "" {
		repoTitle = "Untitled"
	}
	ownerLabel := SanitizeShortSummary(args.Owner)
	if len(ownerLabel) > 120 {
		ownerLabel = ownerLabel[:120]
	}

	var featuresList string
	switch args.Features.(type) {
	case []string, []any:
		items := toAnySlice(args.Features)
		var feats []string
		for _, f := range items {
			if f == nil || f == "" {
				continue // `if f` — falsy entries skipped pre-sanitize
			}
			feats = append(feats, SanitizeBodyText(f, 300))
		}
		feats = filterNonBlank(feats)
		if len(feats) == 0 {
			featuresList = "- (none listed)"
		} else {
			featuresList = joinBulletList(feats)
		}
	default:
		featuresList = SanitizeBodyText(args.Features, 0)
		if featuresList == "" {
			featuresList = "(none listed)"
		}
	}

	// Quality issues (v22 Feature 5) — auto-detect ONLY when none were
	// supplied (Python `if quality_issues is None`).
	qualityIssues := args.QualityIssues
	if qualityIssues == nil {
		qualityIssues = []string{}
		if len(summary) < 50 {
			qualityIssues = append(qualityIssues, "summary too short")
		}
		if isList(args.Features) && len(toAnySlice(args.Features)) < 3 {
			qualityIssues = append(qualityIssues, "fewer than 3 features")
		}
		if args.Confidence < 30 {
			qualityIssues = append(qualityIssues, "low confidence")
		}
		if args.CategoryKey == "Uncategorized" {
			qualityIssues = append(qualityIssues, "uncategorized")
		}
		if !args.IsLowQuality {
			args.IsLowQuality = len(qualityIssues) > 0
		}
	}
	var cleanIssues []string
	for _, q := range qualityIssues {
		if clean := SanitizeSeqItem(q, 40); clean != "" {
			cleanIssues = append(cleanIssues, clean)
		}
	}

	// Banner — the security gate.
	bannerCoverLine, bannerSection := bannerRefs(args.BannerPath)

	// Frontmatter pieces (the empty-string quirks are pinned by the
	// golden: absent optionals emit a BLANK template line).
	aliasesYAML := "aliases: []"
	if len(aliases) > 0 {
		var b strings.Builder
		b.WriteString("aliases:\n")
		for _, a := range aliases {
			b.WriteString("  - " + a + "\n")
		}
		aliasesYAML = b.String()
	}
	tagsYAML := "tags: []"
	if len(tags) > 0 {
		tagsYAML = "tags: [" + strings.Join(tags, ", ") + "]"
	}
	languagesYAML := ""
	if len(languages) > 0 {
		languagesYAML = "languages: [" + strings.Join(languages, ", ") + "]"
	}
	langLine := ""
	if langScalar != "" {
		langLine = "primary_language: " + langScalar
	}
	lastReleaseLine := ""
	if releaseScalar != "" {
		lastReleaseLine = "last_release: " + releaseScalar
	}
	qualityLines := ""
	if args.IsLowQuality {
		qualityLines = "quality: low\nquality_issues: [" +
			strings.Join(cleanIssues, ", ") + "]"
	}
	shortSummarySection := ""
	if shortSummary != "" {
		shortSummarySection = "> **TL;DR:** " + shortSummary + "\n\n"
	}

	langLabel := langScalar
	if langLabel == "" {
		langLabel = "N/A"
	}
	langsLine := "N/A"
	if len(languages) > 0 {
		langsLine = strings.Join(languages, ", ")
	}
	releaseLine := args.LatestReleaseDate
	if releaseLine == "" {
		releaseLine = "No releases"
	}
	qualityLine := "✅ good"
	if args.IsLowQuality {
		qualityLine = "⚠️ low (" + strings.Join(cleanIssues, ", ") + ")"
	}

	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("source: " + urlYAML + "\n")
	b.WriteString(aliasesYAML + "\n")
	b.WriteString(tagsYAML + "\n")
	b.WriteString("category: " + categoryYAML + "\n")
	b.WriteString("stars: " + strconv.Itoa(args.Stars) + "\n")
	b.WriteString("org: " + orgYAML + "\n")
	b.WriteString(langLine + "\n")
	b.WriteString(languagesYAML + "\n")
	b.WriteString("credibility_score: " +
		strconv.Itoa(args.CredScore) + "/100\n")
	b.WriteString("date_processed: " + localDateString() + "\n")
	b.WriteString(lastReleaseLine + "\n")
	b.WriteString(bannerCoverLine + "\n")
	b.WriteString(qualityLines + "\n")
	b.WriteString("managed_by: " + constants.ManagedByGitcurator + "\n")
	b.WriteString("schema_version: " + constants.NoteSchemaVersion + "\n")
	b.WriteString("prompt_version: " + constants.GitHubPromptVersion + "\n")
	b.WriteString("---\n\n")
	b.WriteString(bannerSection)
	b.WriteString(constants.OwnershipBanner + "\n\n")
	b.WriteString("# " + repoTitle + "\n\n")
	b.WriteString(shortSummarySection)
	b.WriteString("**`" + ownerLabel + "/" + repoTitle + "`** · ⭐ " +
		helpers.FormatComma(int64(args.Stars)) + " · 🔧 " + langLabel + "\n\n")
	b.WriteString("## What is it?\n" + summary + "\n\n")
	b.WriteString("## How does it work?\n" + howItWorks + "\n\n")
	b.WriteString("## Why is it important? (Core Value)\n" + coreValue + "\n\n")
	b.WriteString("## Key Features & Technologies\n" + featuresList + "\n\n")
	b.WriteString("## Difference from Others\n" + difference + "\n\n")
	b.WriteString("## 🏢 Organization & Credibility\n")
	b.WriteString("- **Developer:** " + args.OrgName + "\n")
	b.WriteString("- **Reputation:** " + RepToStr(args.OrgRep) + "\n")
	b.WriteString("- **Stars:** " + helpers.FormatComma(int64(args.Stars)) + "\n")
	b.WriteString("- **Forks:** " + strconv.Itoa(args.Forks) + "\n")
	b.WriteString("- **Recent Activity:** " +
		strconv.Itoa(args.CommitCount) + " commits in 3 months\n")
	b.WriteString("- **Credibility Score:** " + strconv.Itoa(args.CredScore) +
		"/100 (" + ScoreToRating(args.CredScore) + ")\n")
	b.WriteString("- **Languages:** " + langsLine + "\n")
	b.WriteString("- **Last Release:** " + releaseLine + "\n")
	b.WriteString("- **Quality:** " + qualityLine + "\n\n")
	b.WriteString("---\n")
	b.WriteString("*Source: [GitHub](" + args.URL + ")*\n\n")
	b.WriteString(trashFooter)
	return b.String()
}

// trashFooter is the v0.65.0 trash-tag footer paragraph (verbatim).
const trashFooter = `*Not useful anymore? Tag this note 🗑️ or delete / auto_delete —
typed inline anywhere in the note (Obsidian's own tag syntax), added
in the tags property, or moved into the Trash folder — the next run
counts it, asks you to confirm on Telegram, and on your 🗑️ Delete it
leaves the library and this repo is never fetched or counted again
(v0.65.0).*
`

func isList(v any) bool {
	switch v.(type) {
	case []string, []any:
		return true
	}
	return false
}

func toAnySlice(v any) []any {
	switch vv := v.(type) {
	case []any:
		return vv
	case []string:
		out := make([]any, len(vv))
		for i, s := range vv {
			out[i] = s
		}
		return out
	default:
		return nil
	}
}

func filterNonBlank(feats []string) []string {
	var out []string
	for _, f := range feats {
		if strings.TrimSpace(f) != "" {
			out = append(out, f)
		}
	}
	return out
}

func joinBulletList(feats []string) string {
	var b strings.Builder
	for i, f := range feats {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString("- " + f)
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Missing-repo placeholder note (note_builder.py:352-400).
// ---------------------------------------------------------------------------

// BuildMissingRepoNote mirrors build_missing_repo_note: the _missing/
// placeholder for a 404 GitHub repo. Deliberately NOT a managed content
// note — no ownership banner, no category — so note_state skips the
// _missing folder (deleting this note is the re-check trigger, never a
// dismissal) and the Library mirror never copies it. The source: key IS
// the VaultIndex dedupe key.
func BuildMissingRepoNote(url, owner, repo string, strikes int) string {
	today := localDateString()
	cleanRepo := strings.TrimSpace(repo)
	cleanRepo = strings.TrimSuffix(cleanRepo, ".git")
	var b strings.Builder
	b.WriteString("---\n")
	b.WriteString("source: " + url + "\n")
	b.WriteString("repo: " + owner + "/" + cleanRepo + "\n")
	b.WriteString("status: missing\n")
	b.WriteString("reason: 404 Not Found (deleted or private)\n")
	b.WriteString("strikes: " + strconv.Itoa(strikes) + "\n")
	b.WriteString("date_recorded: " + today + "\n")
	b.WriteString("placeholder: missing-repo\n")
	b.WriteString("---\n\n")
	b.WriteString("# " + owner + "/" + cleanRepo + " — missing (404)\n\n")
	b.WriteString("The repository **" + url + "** returned **404 Not Found** — it was deleted or\n")
	b.WriteString("made private. GitCurator recorded this placeholder so the link:\n\n")
	b.WriteString("- no longer counts as *remaining to be processed* in any queue view,\n")
	b.WriteString("- is never fetched from the GitHub API again (404 quarantine confirmed).\n\n")
	b.WriteString("**To re-check this repo later** (restored / renamed / made public):\n\n")
	b.WriteString("1. Delete this note (the file you are reading).\n")
	b.WriteString("2. More ▸ View 404 Quarantine → reset this URL\n")
	b.WriteString("   (or the CLI: `python main.py --cli --reset-dead \"" + url + "\"`).\n")
	b.WriteString("3. Run the next batch — the repo is processed like new.\n\n")
	b.WriteString("*Recorded " + today + " after " + strconv.Itoa(strikes) +
		" confirmed 404(s). The x/y links of this\n")
	b.WriteString("repo, if any, live on in the bot chat; nothing was lost.*\n")
	return b.String()
}
