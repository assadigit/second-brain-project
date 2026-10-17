// recall.go — recall.py (Phase 6, step 1): recall hooks — the delimited
// "Best used for" block on GitHub notes, the sanitizers, field
// extraction from both note formats, the plan/run pipeline and the
// report. The markers are note_state's (the fingerprint strips the
// block, so an app-added recall hook NEVER reads as a human edit).
package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/prompts"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// RecallPlaceholder is the fallback sentence (never a guess; SPEC: omit
// rather than guess).
const RecallPlaceholder = "Use when you need to… (not captured — see the source link)."

// RecallLinePrefix is the "Best used for" line rendered inside the block.
const RecallLinePrefix = "**Best used for:** "

// Sanitization bounds: one line, starting with the required words.
const (
	recallStartsWith = "use when you need to"
	recallMaxLen     = 220 // ~25 words plus slack for punctuation
)

// ---------------------------------------------------------------------------
// Block operations (pure string surgery)
// ---------------------------------------------------------------------------

var recallBlockRE = regexp.MustCompile(
	`(?s)` + regexp.QuoteMeta(notestate.RecallStart) + `.*?` +
		regexp.QuoteMeta(notestate.RecallEnd))

// StripRecallBlock is the note text WITHOUT the delimited recall block
// (and without the blank lines the block may have dragged along) — the
// view the fingerprint sees.
func StripRecallBlock(text string) string {
	return helpers.Rstrip(recallBlockRE.ReplaceAllString(text, ""))
}

// HasRecallBlock reports the delimited block's presence.
func HasRecallBlock(text string) bool {
	return recallBlockRE.MatchString(text)
}

// GetRecallSentence is the raw sentence inside the block (” when
// absent), with the **Best used for:** prefix and markers stripped.
func GetRecallSentence(text string) string {
	m := recallBlockRE.FindString(text)
	if m == "" {
		return ""
	}
	inner := m[len(notestate.RecallStart) : len(m)-len(notestate.RecallEnd)]
	inner = helpers.Strip(inner)
	if strings.HasPrefix(inner, RecallLinePrefix) {
		inner = inner[len(RecallLinePrefix):]
	}
	return helpers.Strip(inner)
}

// RenderRecallBlock is the full delimited block text for a (sanitized)
// sentence.
func RenderRecallBlock(sentence string) string {
	return notestate.RecallStart + "\n" + RecallLinePrefix + sentence + "\n" +
		notestate.RecallEnd
}

// InsertRecallBlock adds or REPLACES the recall block, touching nothing
// else. A missing block is appended at the end (after a blank line); an
// existing block is replaced in place (idempotent). Line endings follow
// the note's own dominant ending (CRLF vaults stay CRLF).
func InsertRecallBlock(text, sentence string) string {
	crlf := strings.Contains(text, "\r\n") &&
		float64(strings.Count(text, "\r\n")) > float64(strings.Count(text, "\n"))/2
	sep := "\n"
	if crlf {
		sep = "\r\n"
	}
	block := RenderRecallBlock(sentence)
	if crlf {
		block = strings.ReplaceAll(block, "\n", "\r\n")
	}
	if HasRecallBlock(text) {
		// literal replacement (no backslash-reference surprises)
		return replaceFirstBlock(text, block)
	}
	stripped := helpers.Rstrip(text)
	if stripped == "" {
		return block
	}
	return stripped + sep + sep + block + sep
}

// replaceFirstBlock swaps the FIRST delimited block occurrence for
// block, literally (the Python `sub(lambda _m: block, text, count=1)`
// shape).
func replaceFirstBlock(text, block string) string {
	loc := recallBlockRE.FindStringIndex(text)
	if loc == nil {
		return text
	}
	return text[:loc[0]] + block + text[loc[1]:]
}

// SanitizeRecallSentence: LLM output → a shippable recall sentence. Must
// start with the required words (case-insensitive) AND name something
// after them, be a single short line; anything else (marketingese, a
// bare prefix, multi-line, empty, refusals) becomes the explicit
// placeholder — the app never ships a guess.
func SanitizeRecallSentence(raw string) string {
	sentence := helpers.Strip(raw)
	// strip surrounding quotes the model sometimes adds
	if r := []rune(sentence); len(r) >= 2 && r[0] == r[len(r)-1] &&
		(r[0] == '"' || r[0] == '\'' || r[0] == '`') {
		sentence = helpers.Strip(string(r[1 : len(r)-1]))
	}
	// one line only (Python str.splitlines' boundary set)
	sentence = firstPyLine(sentence)
	if sentence == "" {
		return RecallPlaceholder
	}
	lowered := helpers.PyLower(sentence)
	if !strings.HasPrefix(lowered, recallStartsWith) {
		return RecallPlaceholder
	}
	// a BARE prefix ("Use when you need to" and nothing after) names no
	// problem — that is a refusal in disguise
	if utf8.RuneCountInString(lowered) <= len(recallStartsWith)+1 {
		return RecallPlaceholder
	}
	if utf8.RuneCountInString(sentence) > recallMaxLen {
		// cut at the last word boundary inside the bound — a recall line
		// longer than the bound is a model rambling, not a hard error
		cut := helpers.RstripChars(
			lastSpaceCut(helpers.CutRunes(sentence, recallMaxLen)), ",;:")
		if strings.HasPrefix(helpers.PyLower(cut), recallStartsWith) &&
			utf8.RuneCountInString(cut) > len(recallStartsWith)+1 {
			sentence = cut
		} else {
			return RecallPlaceholder
		}
	}
	return sentence
}

// firstPyLine takes str.splitlines()[0] over Python's line-boundary set
// (\n, \r, \r\n as ONE, \v, \f, \x1c-\x1e, \x85, U+2028, U+2029).
func firstPyLine(s string) string {
	for i, r := range s {
		switch r {
		case '\n', '\r', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
			return helpers.Strip(s[:i])
		}
	}
	return helpers.Strip(s)
}

// lastSpaceCut mirrors `s.rsplit(' ', 1)[0]`: everything before the
// LAST space (the whole string when there is no space).
func lastSpaceCut(s string) string {
	if i := strings.LastIndexByte(s, ' '); i >= 0 {
		return s[:i]
	}
	return s
}

// ---------------------------------------------------------------------------
// Field extraction (best-effort parsers over the note formats in the wild)
// ---------------------------------------------------------------------------

var (
	frontTagsRE = regexp.MustCompile(`(?m)^tags:` + pyS + `*\[(.*?)\]` + pyS + `*$`)
	titleRE     = regexp.MustCompile(`(?m)^#` + pyS + `+(.+?)` + pyS + `*$`)
	tldrRE      = regexp.MustCompile(`>` + pyS + `*\*\*TL;DR:\*\*` + pyS + `*(.+)`)

	// The section twins: ^##\s*Heading\s*$\r?\n+(.*?)\r?\n^##\s with
	// MULTILINE|DOTALL.
	summaryRE = regexp.MustCompile(
		`(?ms)^##` + pyS + `*What is it\?` + pyS + `*$\r?\n+(.*?)\r?\n^##` + pyS)
	featuresRE = regexp.MustCompile(
		`(?ms)^##` + pyS + `*Key Features & Technologies` + pyS + `*$\r?\n+(.*?)\r?\n^##` + pyS)
	bestUsedRE = regexp.MustCompile(
		`(?ms)^##` + pyS + `*Best used for` + pyS + `*$\r?\n+(.*?)\r?\n^##` + pyS)
)

func sectionOf(re *regexp.Regexp, text string, maxLen int) string {
	m := re.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return helpers.CutRunes(helpers.Strip(m[1]), maxLen)
}

func tagsOf(text string) string {
	m := frontTagsRE.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	var parts []string
	for _, t := range strings.Split(m[1], ",") {
		if helpers.Strip(t) != "" {
			parts = append(parts, helpers.Strip(t))
		}
	}
	return strings.Join(parts, ", ")
}

func titleOf(text string) string {
	m := titleRE.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return helpers.Strip(m[1])
}

func tldrOf(text string) string {
	m := tldrRE.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return helpers.Strip(m[1])
}

// GithubFields is extract_github_fields' output shape (the r01 slots).
type GithubFields struct {
	RepoName string
	TLDR     string
	Summary  string
	Features string
	Tags     string
}

// ExtractGithubFields: GitHub note → the r01 prompt slots (best-effort,
// never raises).
func ExtractGithubFields(text string) GithubFields {
	return GithubFields{
		RepoName: titleOf(text),
		TLDR:     helpers.CutRunes(tldrOf(text), 300),
		Summary:  sectionOf(summaryRE, text, 600),
		Features: sectionOf(featuresRE, text, 600),
		Tags:     tagsOf(text),
	}
}

// WebsiteFields is extract_website_fields' output shape.
type WebsiteFields struct {
	Name        string
	OneLine     string
	BestUsedFor string
	Tags        string
}

// ExtractWebsiteFields: website note → the recall fields (the body
// already carries them — no LLM needed for the field itself).
func ExtractWebsiteFields(text string) WebsiteFields {
	return WebsiteFields{
		Name:        titleOf(text),
		OneLine:     helpers.CutRunes(tldrOf(text), 300),
		BestUsedFor: sectionOf(bestUsedRE, text, 400),
		Tags:        tagsOf(text),
	}
}

// RecallFields is the uniform recall view of a note from either vault:
// {name, one_line, recall, tags} (recall = the Best-used-for sentence —
// the delimited block on GitHub notes, the body section on website
// notes).
type RecallFields struct {
	Name    string
	OneLine string
	Recall  string
	Tags    string
}

// RecallFieldsFor mirrors recall_fields_for.
func RecallFieldsFor(text, vault string) RecallFields {
	if vault == notestate.VaultGithub {
		gh := ExtractGithubFields(text)
		return RecallFields{
			Name: gh.RepoName, OneLine: gh.TLDR,
			Recall: GetRecallSentence(text), Tags: gh.Tags,
		}
	}
	web := ExtractWebsiteFields(text)
	return RecallFields{
		Name: web.Name, OneLine: web.OneLine,
		Recall: web.BestUsedFor, Tags: web.Tags,
	}
}

// ---------------------------------------------------------------------------
// Planning + running (the tool layer drives these)
// ---------------------------------------------------------------------------

// BuildRecallPrompt fills r01 from the extracted fields (the loader
// refuses a missing slot — a half-filled prompt is never sent).
func BuildRecallPrompt(fields GithubFields) (string, error) {
	orFallback := func(v, fallback string) string {
		if v == "" {
			return fallback
		}
		return v
	}
	return prompts.LoadPrompt("r01_recall",
		prompts.Slot{Name: "REPO_NAME", Value: orFallback(fields.RepoName, "(unknown repository)")},
		prompts.Slot{Name: "TLDR", Value: orFallback(fields.TLDR, "(not captured)")},
		prompts.Slot{Name: "SUMMARY", Value: orFallback(fields.Summary, "(not captured)")},
		prompts.Slot{Name: "FEATURES", Value: orFallback(fields.Features, "(not captured)")},
		prompts.Slot{Name: "TAGS", Value: orFallback(fields.Tags, "(none)")},
	)
}

// RecallPlanEntry is one planned note (plan_recall_hooks' dict shape).
type RecallPlanEntry struct {
	Path      string
	SourceURL string
	HasBlock  bool
	Current   string
}

// PlanRecallHooks — read-only: which GitHub-vault notes need a recall
// hook? Without force, notes that already carry a block are marked done
// (the caller skips them); with it they are re-planned for a refresh.
// Unmanaged notes (no source:) are skipped — the app does not write
// into files it does not own.
func PlanRecallHooks(vaultPath string, force bool) []RecallPlanEntry {
	var planned []RecallPlanEntry
	for _, note := range notestate.ScanVault(vaultPath, nil) {
		if note.Unmanaged || note.Path == "" {
			continue
		}
		text, ok := readTextFile(note.Path)
		if !ok {
			continue
		}
		planned = append(planned, RecallPlanEntry{
			Path:      note.Path,
			SourceURL: note.SourceURL,
			HasBlock:  HasRecallBlock(text),
			Current:   GetRecallSentence(text),
		})
	}
	if !force {
		var kept []RecallPlanEntry
		for _, p := range planned {
			if !p.HasBlock {
				kept = append(kept, p)
			}
		}
		return kept
	}
	return planned
}

// RecallLLM is the reference's llm(messages) closure shape.
type RecallLLM func(messages []llmclient.Message) (string, error)

// RecallRunResult is one run_recall_hooks result entry.
type RecallRunResult struct {
	Path      string
	SourceURL string
	HasBlock  bool
	Current   string

	Sentence  string `json:",omitempty"`
	OldBlock  string `json:",omitempty"`
	NewBlock  string `json:",omitempty"`
	Unchanged bool
	Written   bool   `json:",omitempty"`
	Error     string `json:",omitempty"`
}

// RunRecallHooks runs the plan: extract fields → prompt → LLM →
// sanitize → (write). Dry-run by default — every entry in the result
// carries the before/after block text for the diff report. Apply writes
// through storage.AtomicWriteText (or the writer hook the tests
// inject). Never fails on a single note's failure: the entry gets Error
// and the run continues.
//
// Divergence note (the T-40 house law): the reference's error strings
// render f"{type(e).__name__}: {e}" (Python exception classes); the Go
// port renders the error's own text. The pinned contract is the SHAPE
// (an error entry + the run continues), never the class names.
func RunRecallHooks(plan []RecallPlanEntry, llm RecallLLM,
	apply bool, limit int, progress func(string),
	writer func(path, content string)) []RecallRunResult {

	say := func(msg string) {
		if progress != nil {
			progress(msg)
		}
	}
	write := writer
	if write == nil {
		write = func(path, content string) {
			_ = storage.AtomicWriteText(path, content)
		}
	}
	var results []RecallRunResult
	todo := plan
	if limit > 0 && limit < len(plan) {
		todo = plan[:limit]
	}
	for _, entry := range todo {
		out := RecallRunResult{
			Path: entry.Path, SourceURL: entry.SourceURL,
			HasBlock: entry.HasBlock, Current: entry.Current,
		}
		results = append(results, out)
		res := &results[len(results)-1]
		if err := runOneRecallHook(res, entry, llm, apply, say, write); err != nil {
			res.Error = err.Error()
			say(fmt.Sprintf("  ! %s — %s", filepath.Base(entry.Path), res.Error))
		}
	}
	return results
}

// runOneRecallHook is the per-note pipeline; a returned error is the
// reference's caught-exception path (the entry degrades, never the run).
func runOneRecallHook(res *RecallRunResult, entry RecallPlanEntry,
	llm RecallLLM, apply bool, say func(string),
	write func(path, content string)) error {

	text, ok := readTextFile(entry.Path)
	if !ok {
		return fmt.Errorf("cannot read %s", entry.Path)
	}
	prompt, err := BuildRecallPrompt(ExtractGithubFields(text))
	if err != nil {
		return err
	}
	reply, err := llm([]llmclient.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return err
	}
	raw := ""
	if payload, jerr := llmclient.ExtractJSON(reply); jerr == nil {
		if s, ok := payloadGet(payload, "best_used_for"); ok {
			raw = s
		}
	} else {
		raw = reply
	}
	sentence := SanitizeRecallSentence(raw)
	newBlock := RenderRecallBlock(sentence)
	oldBlock := ""
	if HasRecallBlock(text) {
		oldBlock = recallBlockRE.FindString(text)
	}
	res.Sentence = sentence
	res.OldBlock = oldBlock
	res.NewBlock = newBlock
	if entry.HasBlock && entry.Current == sentence {
		res.Unchanged = true
		say(fmt.Sprintf("  = %s — unchanged", filepath.Base(entry.Path)))
		return nil
	}
	res.Unchanged = false
	if apply {
		write(entry.Path, InsertRecallBlock(text, sentence))
		res.Written = true
		replaced := ""
		if entry.HasBlock {
			replaced = " (replaced)"
		}
		say(fmt.Sprintf("  + %s — hook written%s",
			filepath.Base(entry.Path), replaced))
	} else {
		say(fmt.Sprintf("  ~ %s — would write: %s",
			filepath.Base(entry.Path), helpers.CutRunes(sentence, 70)))
	}
	return nil
}

// payloadGet reads a key from an ExtractJSON payload dict (either
// representation), Python str()ing non-strings on the way out (None →
// the empty string, mirroring str(None or ”) on the sanitize path).
func payloadGet(payload any, key string) (string, bool) {
	var v any
	present := false
	switch t := payload.(type) {
	case *storage.OrderedMap:
		v, present = t.Get(key)
	case map[string]any:
		v, present = t[key]
	default:
		return "", false
	}
	if !present || v == nil {
		if present {
			return "", true
		}
		return "", false
	}
	return helpers.PyStr(v), true
}

// RecallReport is the human-readable report (written under
// app/reports/recall/).
func RecallReport(results []RecallRunResult, vaultPath string, applied bool) string {
	mode := "DRY RUN (nothing written)"
	if applied {
		mode = "APPLIED"
	}
	lines := []string{
		"GitCurator — recall hooks — " + mode,
		"GitHub vault: " + vaultPath,
		fmt.Sprintf("Notes in this run: %d", len(results)),
		"",
	}
	for _, r := range results {
		p := r.Path
		if p == "" {
			p = "?"
		}
		name := filepath.Base(p)
		switch {
		case r.Error != "":
			lines = append(lines, fmt.Sprintf("ERROR    %s — %s", name, r.Error))
		case r.Unchanged:
			lines = append(lines, fmt.Sprintf("KEEP     %s — already current", name))
		case r.Written:
			lines = append(lines, fmt.Sprintf("WRITTEN  %s — %s", name, r.Sentence))
		default:
			lines = append(lines, fmt.Sprintf("PLANNED  %s — %s", name, r.Sentence))
		}
	}
	placeholder := 0
	for _, r := range results {
		if r.Sentence == RecallPlaceholder {
			placeholder++
		}
	}
	if placeholder > 0 {
		lines = append(lines, "")
		lines = append(lines, fmt.Sprintf(
			"%d note(s) got the 'not captured' placeholder — the source text "+
				"was too thin to name a problem (deliberate: omit rather than "+
				"guess).", placeholder))
	}
	lines = append(lines, "")
	lines = append(lines, "The block is delimited and fingerprint-invisible: "+
		"note_state will not read it as a human edit.")
	return strings.Join(lines, "\n") + "\n"
}

// ReportDir mirrors report_dir: APP_DIR/reports/recall, created.
func ReportDir() string {
	path := filepath.Join(constants.AppDir(), "reports", "recall")
	_ = os.MkdirAll(path, 0o755)
	return path
}

// pyStrftimeSeconds is time.strftime('%Y-%m-%d %H:%M:%S') on the
// package clock seam.
func pyStrftimeSeconds() string {
	return clockNow().Format("2006-01-02 15:04:05")
}
