// fixtures_test.go — the reference suite's own note shapes
// (app/tests/test_directorysync.py's _fm/_note/_footer + LONG_NOTE &
// friends), transcribed verbatim. The golden differential
// (golden_test.go) byte-compares the Go parser's output against the
// reference's — any transcription drift fails there.
package directorysync

import (
	"strings"
)

// fmFields mirrors _fm(over): the thirteen base frontmatter fields in
// the reference's fixed order; over replaces in place (order kept)
// or appends (sorted — the reference's tests only ever append ONE new
// key).
func fmFields(over map[string]string) string {
	base := [][2]string{
		{"source", "https://coolors.co/"},
		{"aliases", "[]"},
		{"tags", "[color-tool, palette]"},
		{"category", `"Design"`},
		{"subcategory", `"Assets & Resources"`},
		{"fetch_status", `"ok"`},
		{"pricing", `"freemium"`},
		{"login_required", `"no"`},
		{"date_processed", "2026-09-28"},
		{"managed_by", `"gitcurator"`},
		{"schema_version", `"1"`},
		{"prompt_version", `"web-v2"`},
	}
	used := map[string]bool{}
	var lines []string
	for _, kv := range base {
		v := kv[1]
		if o, ok := over[kv[0]]; ok {
			v = o
			used[kv[0]] = true
		}
		lines = append(lines, kv[0]+": "+v)
	}
	var extra []string
	for k, v := range over {
		if !used[k] {
			extra = append(extra, k+": "+v)
		}
	}
	// sortStringsDS keeps a single extra key deterministic (the
	// reference's only multi-new-key case never appears in the suite)
	sortStringsDS(extra)
	lines = append(lines, extra...)
	return strings.Join(lines, "\n")
}

// refNote mirrors _note(fm_text, body).
func refNote(fmText, body string) string {
	return "---\n" + fmText + "\n---\n\n" +
		"> [!info] Managed by GitCurator — machine-written note.\n\n" + body
}

// refFooter mirrors _footer(url).
func refFooter(url string) string {
	return "\n---\n*Source: [" + url + "](" + url + ")*\n\n" +
		"*Not useful anymore? Tag this note 🗑️ or delete / auto_delete —\n" +
		"typed inline anywhere in the note (Obsidian's own tag syntax) or\n" +
		"added in the tags property — the next run counts it, asks you to\n" +
		"confirm on Telegram, and on your 🗑️ Delete it leaves the library and\n" +
		"never fetches this site again (v0.60.0).*\n"
}

var longNote = refNote(fmFields(nil), `# Coolors

> **TL;DR:** The super-fast color scheme generator for designers.

## What it does
- Generates color palettes with a single spacebar press
- Shows HEX, RGB and CSS values for every swatch
- Exports palettes as PNG, PDF, SVG and URL
- Contrast checking against WCAG standards

## Best used for
Use it when you need a quick, harmonious palette for a landing page or a brand exploration, and you want to lock it fast.

## Practical details
- Pricing: freemium — the free tier covers most work
- Sign-up required: no
- Export formats: PNG, SVG, PDF, URL
- Works fully in the browser

## Watch out
- The pro tier is required for some export formats and gradient tools

## Similar tools
Adobe Color, Khroma, Paletton
`+refFooter("https://coolors.co/"))

var shortNote = refNote(fmFields(map[string]string{
	"source":         "https://fast.com/",
	"tags":           "[speed-test]",
	"subcategory":    `""`,
	"pricing":        `"unknown"`,
	"login_required": `"unknown"`,
}), `# Fast.com

> **TL;DR:** Netflix's simple internet speed test.

## Best used for
When you want one number, right now.
`+refFooter("https://fast.com/"))

var missingNote = refNote(fmFields(map[string]string{
	"source":         "https://mobbin.com/",
	"pricing":        `"freemium"`,
	"login_required": `"yes"`,
}), `# Mobbin

> **TL;DR:** A browsable library of real mobile and web UI screens.

## Best used for
Use it when you need real-world reference screens for a flow you are designing.

## Similar tools
—
`+refFooter("https://mobbin.com/"))

var emptyBulletNote = refNote(fmFields(map[string]string{
	"source":         "https://v0.dev/",
	"tags":           "[ai-powered, ui-generation]",
	"category":       `"AI Tools for Web & App Development"`,
	"subcategory":    `"AI UI/Design Generation"`,
	"pricing":        `"freemium"`,
	"login_required": `"yes"`,
}), "# v0\n\n"+
	"> **TL;DR:** Vercel's generative UI — describe a screen, get React code.\n\n"+
	"## Watch out\n"+
	"- The free tier is limited\n"+
	"- \n"+ // the reference's empty bullet: dash + space + newline
	refFooter("https://v0.dev/"))

var strayNote = "---\n" +
	"source: \"https://icons8.com/\"\n" +
	"aliases: []\n" +
	"tags: [icons, illustration]\n" +
	"category: \"Design\"\n" +
	"subcategory: \"Assets & Resources\"\n" +
	"fetch_status: \"ok\"\n" +
	"pricing: \"freemium\"\n" +
	"login_required: \"no\"\n" +
	"date_processed: 2026-10-02\n" +
	"managed_by: \"gitcurator\"\n" +
	"schema_version: \"1\"\n" +
	"prompt_version: \"web-v2\"\n" +
	"---\n" +
	"\n" +
	"--- Descript\n" +
	"\n" +
	"> [!info] Managed by GitCurator — machine-written note.\n" +
	"\n" +
	"# Icons8\n" +
	"\n" +
	"> **TL;DR:** Icons, illustrations and photos with consistent styles.\n" +
	"\n" +
	"## What it does\n" +
	"- A large icon library across dozens of style families\n" +
	"- Consistent companion illustration and photo sets\n" +
	"\n" +
	"## Best used for\n" +
	"Use it when one visual family must carry a whole product.\n" +
	"\n" +
	"## Practical details\n" +
	"- Pricing: freemium — free tier with attribution\n" +
	refFooter("https://icons8.com/")

var unstructuredNote = "---\n" +
	"source: \"https://window-swap.com/\"\n" +
	"aliases: []\n" +
	"tags: [curiosity]\n" +
	"category: \"Random / Curiosities\"\n" +
	"subcategory: \"\"\n" +
	"fetch_status: \"ok\"\n" +
	"pricing: \"free\"\n" +
	"login_required: \"no\"\n" +
	"date_processed: 2026-10-10\n" +
	"managed_by: \"gitcurator\"\n" +
	"---\n" +
	"\n" +
	"> [!info] Managed by GitCurator — machine-written note.\n" +
	"\n" +
	"# Window Swap\n" +
	"\n" +
	"This is just prose. It has no sections at all, and the drawer must\n" +
	"show the full text as plain formatted content. See [the project](https://window-swap.com/).\n" +
	refFooter("https://window-swap.com/")

var privateNote = refNote(fmFields(map[string]string{
	"source":         "https://annas-archive.org/",
	"tags":           "[shadow-library, ebook]",
	"category":       `"Download Resources"`,
	"subcategory":    `"Torrent & Shadow Libraries"`,
	"pricing":        `"free"`,
	"login_required": `"no"`,
}), `# Anna's Archive

> **TL;DR:** A shadow-library aggregator — PRIVATE, never published.

## Watch out
- Never appears on the public site
`+refFooter("https://annas-archive.org/"))
