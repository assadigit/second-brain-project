package websitepipeline

import (
	"fmt"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notebuilder"
)

// notes.go — the website note builder (build_website_note /
// build_review_note, website_pipeline.py:97-264). Every value that
// reaches YAML goes through the note_builder sanitizers (non-
// negotiable #6: LLM output is untrusted input); category and
// subcategory are taxonomy-validated BEFORE these functions run —
// they are still re-sanitized here (defense in depth).

// clockNow is the package's clock seam: the golden runner freezes it
// (the reference generator freezes wp.datetime for the same runs);
// production reads the wall clock.
var clockNow = time.Now

// localDate mirrors datetime.now().strftime("%Y-%m-%d") (local time).
func localDate() string { return clockNow().Format("2006-01-02") }

// pyGet reads a key off a map[string]any with Python's .get(key)
// (missing → nil).
func pyGet(m map[string]any, key string) any {
	if m == nil {
		return nil
	}
	return m[key]
}

// getStr mirrors analysis.get(k) or ” → a clean string.
func getStr(m map[string]any, key string) string {
	v := pyGet(m, key)
	if v == nil {
		return ""
	}
	return helpers.PyStr(v)
}

// getStrList mirrors the `or []` + isinstance(str) → [str] ladder the
// note builder uses for its sequence fields.
func getStrList(m map[string]any, key string) []any {
	v := pyGet(m, key)
	if v == nil {
		return nil
	}
	if s, ok := v.(string); ok {
		return []any{s}
	}
	if list, ok := v.([]any); ok {
		return list
	}
	return []any{v}
}

// runeSlice mirrors Python's s[:n] (codepoint-wise).
func runeSlice(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

// sanitizeSeqList runs SanitizeBodyText over a sequence and drops
// blank items (the does/practical/watch ladders).
func sanitizeSeqList(items []any, maxLen int, cap int) []string {
	var out []string
	for _, it := range items {
		clean := notebuilder.SanitizeBodyText(it, maxLen)
		if strings.TrimSpace(clean) != "" {
			out = append(out, clean)
		}
		if cap > 0 && len(out) >= cap {
			break
		}
	}
	return out
}

// BuildWebsiteNote mirrors build_website_note: build a website note
// (frontmatter + body) from SANITIZED-safe inputs. web-v2: only what
// is known — no "unknown" filler lines; "Best used for" always
// renders (the recall/linking layer regex-parses it and needs a
// heading after it) and "Similar tools" always closes.
func BuildWebsiteNote(u string, analysis map[string]any, category, subcategory, fetchStatus string, tags []string) string {
	name := runeSlice(notebuilder.SanitizeShortSummary(getStr(analysis, "name")), 120)
	if name == "" {
		name = "Untitled site"
	}
	oneLine := notebuilder.SanitizeShortSummary(getStr(analysis, "one_line"))
	// web-v2: "what_it_does" replaces "core_offerings" (old-shape
	// dicts — the offline golden runner, any in-flight analyses —
	// still render).
	does := getStrList(analysis, "what_it_does")
	if len(does) == 0 {
		does = getStrList(analysis, "core_offerings")
	}
	doesLines := sanitizeSeqList(does, 200, 0)
	bestUsedFor := notebuilder.SanitizeBodyText(getStr(analysis, "best_used_for"), 400)
	pricing := strings.ToLower(strings.TrimSpace(getStr(analysis, "pricing")))
	if pricing == "" {
		pricing = "unknown"
	}
	switch pricing {
	case "free", "freemium", "paid", "unknown":
	default:
		pricing = "unknown"
	}
	pricingDetail := notebuilder.SanitizeBodyText(getStr(analysis, "pricing_detail"), 200)
	loginRequired := strings.ToLower(strings.TrimSpace(getStr(analysis, "login_required")))
	if loginRequired == "" {
		loginRequired = "unknown"
	}
	switch loginRequired {
	case "yes", "no", "unknown":
	default:
		loginRequired = "unknown"
	}
	practicalLines := sanitizeSeqList(getStrList(analysis, "practical_details"), 200, 4)
	watchLines := sanitizeSeqList(getStrList(analysis, "watch_out"), 200, 2)
	similar := notebuilder.SanitizeTagsMax(pyGet(analysis, "similar_tools"), 5)

	noteTags := append([]string{}, tags...)
	noteTags = append(noteTags, notebuilder.SanitizeTagsMax(pyGet(analysis, "tags"), 8)...)
	noteTags = notebuilder.SanitizeTagsMax(noteTags, 8)
	if links.IsGistURL(u) {
		hasSnippet := false
		for _, t := range noteTags {
			if strings.ToLower(t) == "snippet" {
				hasSnippet = true
				break
			}
		}
		if !hasSnippet {
			noteTags = append(noteTags, "snippet") // SPEC §4.2: gists get #snippet
		}
	}

	catYAML := notebuilder.YamlScalar(category, 120)
	subYAML := `""`
	if subcategory != "" {
		subYAML = notebuilder.YamlScalar(subcategory, 120)
	}
	urlYAML := notebuilder.YamlScalar(u, 120)
	tagsYAML := "tags: []"
	if len(noteTags) > 0 {
		tagsYAML = "tags: [" + strings.Join(noteTags, ", ") + "]"
	}

	// --- body (web-v2: only what is known) ---
	bestUsedLine := bestUsedFor
	if bestUsedLine == "" {
		bestUsedLine = "Use when you need to… (not captured — see the source link)."
	}
	similarMD := "—"
	if len(similar) > 0 {
		similarMD = strings.Join(similar, ", ")
	}

	var detailLines []string
	if pricingDetail != "" {
		lowerDetail := strings.ToLower(pricingDetail)
		if (pricing == "free" || pricing == "freemium" || pricing == "paid") &&
			!strings.HasPrefix(lowerDetail, pricing) {
			detailLines = append(detailLines, fmt.Sprintf("Pricing: %s — %s", pricing, pricingDetail))
		} else {
			detailLines = append(detailLines, "Pricing: "+pricingDetail)
		}
	} else if pricing == "free" || pricing == "freemium" || pricing == "paid" {
		detailLines = append(detailLines, "Pricing: "+pricing)
	}
	if loginRequired == "yes" || loginRequired == "no" {
		if loginRequired == "yes" {
			detailLines = append(detailLines, "Sign-up required: yes")
		} else {
			detailLines = append(detailLines, "Sign-up required: no")
		}
	}
	detailLines = append(detailLines, practicalLines...)

	var sections []string
	if len(doesLines) > 0 {
		var bullets []string
		for _, o := range doesLines {
			bullets = append(bullets, "- "+o)
		}
		sections = append(sections, "## What it does\n"+strings.Join(bullets, "\n"))
	}
	sections = append(sections, "## Best used for\n"+bestUsedLine)
	if len(detailLines) > 0 {
		var bullets []string
		for _, d := range detailLines {
			bullets = append(bullets, "- "+d)
		}
		sections = append(sections, "## Practical details\n"+strings.Join(bullets, "\n"))
	}
	if len(watchLines) > 0 {
		var bullets []string
		for _, w := range watchLines {
			bullets = append(bullets, "- "+w)
		}
		sections = append(sections, "## Watch out\n"+strings.Join(bullets, "\n"))
	}
	sections = append(sections, "## Similar tools\n"+similarMD)
	bodyMD := strings.Join(sections, "\n\n")

	tldr := oneLine
	if tldr == "" {
		tldr = "—"
	}

	return fmt.Sprintf(`---
source: %s
aliases: []
%s
category: %s
subcategory: %s
fetch_status: "%s"
pricing: "%s"
login_required: "%s"
date_processed: %s
managed_by: "%s"
schema_version: "%s"
prompt_version: "%s"
---

%s

# %s

> **TL;DR:** %s

%s

---
*Source: [%s](%s)*

*Not useful anymore? Tag this note 🗑️ or delete / auto_delete —
typed inline anywhere in the note (Obsidian's own tag syntax) or
added in the tags property — the next run counts it, asks you to
confirm on Telegram, and on your 🗑️ Delete it leaves the library and
never fetches this site again (v0.60.0).*
`,
		urlYAML, tagsYAML, catYAML, subYAML, fetchStatus, pricing,
		loginRequired, localDate(), constants.ManagedByGitcurator,
		constants.NoteSchemaVersion, WebsitePromptVersion,
		constants.OwnershipBanner, name, tldr, bodyMD, u, u)
}

// BuildReviewNote mirrors build_review_note: minimal _review note for
// links that could not be fully processed (SPEC §4.3: "a link that
// cannot be fetched still gets a minimal note in _review with
// fetch_status: failed").
func BuildReviewNote(u, fetchStatus, reason, title string, tags []string) string {
	noteTags := notebuilder.SanitizeTagsMax(tags, 6)
	if links.IsGistURL(u) {
		hasSnippet := false
		for _, t := range noteTags {
			if strings.ToLower(t) == "snippet" {
				hasSnippet = true
				break
			}
		}
		if !hasSnippet {
			noteTags = append(noteTags, "snippet")
		}
	}
	tagsYAML := "tags: []"
	if len(noteTags) > 0 {
		tagsYAML = "tags: [" + strings.Join(noteTags, ", ") + "]"
	}
	title = runeSlice(notebuilder.SanitizeShortSummary(title), 120)
	if title == "" {
		title = u
	}
	if reason == "" {
		reason = "The page could not be processed fully."
	}
	reason = notebuilder.SanitizeBodyText(reason, 400)

	return fmt.Sprintf(`---
source: %s
aliases: []
%s
category: ""
subcategory: ""
fetch_status: "%s"
pricing: "unknown"
login_required: "unknown"
date_processed: %s
managed_by: "%s"
schema_version: "%s"
prompt_version: "%s"
---

%s

# %s

> [!warning] Needs review — %s
> %s

This note is a placeholder created automatically. The link was recorded so
it is never lost; it will be retried automatically.

---
*Source: [%s](%s)*

*Not useful anymore? Tag this note 🗑️ or delete / auto_delete —
typed inline anywhere in the note (Obsidian's own tag syntax) or
added in the tags property — the next run counts it, asks you to
confirm on Telegram, and on your 🗑️ Delete it leaves the library and
never fetches this site again (v0.60.0).*
`,
		notebuilder.YamlScalar(u, 120), tagsYAML, fetchStatus, localDate(),
		constants.ManagedByGitcurator, constants.NoteSchemaVersion,
		WebsitePromptVersion, constants.OwnershipBanner, title,
		fetchStatus, reason, u, u)
}
