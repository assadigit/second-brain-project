package websitepipeline

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/directorysync"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/prompts"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/webextract"
	"github.com/assadigit/second-brain-project/internal/webfetch"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// pipeline_process.go — the per-link state machine (_process_link_inner)
// + the classify/analyze passes + the consolidation + the hand probes
// + the four batch drivers (website_pipeline.py:3858-5338).

// dryrunRemove wraps dryrun.Remove for the tolerance law.
func dryrunRemove(path string) error { return dryrun.Remove(path) }

// runePrefix mirrors Python's s[:n] (codepoint-wise).
func runePrefix(s string, n int) string {
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n])
	}
	return s
}

// llmJSON mirrors _llm_json: one LLM call + robust JSON extraction.
// task tags the pass ('classify' / 'analyze'). Raises on an
// empty/unparseable answer (callers retry, then fall back to _review).
func (p *Pipeline) llmJSON(messages []llmclient.Message, task string) (map[string]any, error) {
	content := p.llmCall(messages, task)
	if strings.TrimSpace(content) == "" {
		return nil, fmt.Errorf("model returned an empty response")
	}
	answer, err := llmclient.ExtractJSON(content)
	if err != nil {
		return nil, err
	}
	// ExtractJSON answers objects as *storage.OrderedMap (the Python
	// dict); the pipeline's consumers read it as a plain map.
	if om, ok := answer.(*storage.OrderedMap); ok {
		m := map[string]any{}
		for _, k := range om.Keys() {
			v, _ := om.Get(k)
			m[k] = v
		}
		return m, nil
	}
	if m, ok := answer.(map[string]any); ok {
		return m, nil
	}
	return nil, fmt.Errorf("answer was not a JSON object")
}

// correctionExamples mirrors _correction_examples (v0.13.0 — the
// deferred Phase-3 few-shot hook): this URL's own correction history
// first, then up to three of the owner's most recent moves in the
// Websites vault. '(none)' when there is nothing.
func (p *Pipeline) correctionExamples(u string) string {
	if p.noteStateDB == nil {
		return "(none)"
	}
	var lines []string
	if own, err := p.noteStateDB.CorrectionsFor(notestate.VaultWebsites, u); err == nil {
		for i, c := range own {
			if i >= 3 {
				break
			}
			lines = append(lines, fmt.Sprintf("- this exact website: the owner moved it %s -> %s",
				correctionCat(c.FromCategory), correctionCat(c.ToCategory)))
		}
	} else {
		p.log(fmt.Sprintf("⚠️ corrections lookup failed for %s: %v", u, err), "warning")
	}
	if recent, err := p.noteStateDB.RecentCorrections(notestate.VaultWebsites, 6); err == nil {
		seen := map[[2]string]bool{}
		for _, c := range recent {
			if c.SourceURL == u {
				continue
			}
			key := [2]string{correctionCat(c.FromCategory), correctionCat(c.ToCategory)}
			if seen[key] {
				continue
			}
			seen[key] = true
			lines = append(lines, fmt.Sprintf("- %s: the owner moved it %s -> %s",
				c.SourceURL, correctionCat(c.FromCategory), correctionCat(c.ToCategory)))
			if len(seen) >= 3 {
				break
			}
		}
	} else {
		p.log(fmt.Sprintf("⚠️ recent-corrections lookup failed: %v", err), "warning")
	}
	if len(lines) == 0 {
		return "(none)"
	}
	return strings.Join(lines, "\n")
}

func correctionCat(cat *string) string {
	if cat == nil || *cat == "" {
		return "(none)"
	}
	return *cat
}

// classifyCategory mirrors _classify_category (SPEC §4.6 pass 1):
// validates the answer; retries up to ClassifyRetries with a
// corrective nudge; returns (category, confidence) or (”, ”).
func (p *Pipeline) classifyCategory(u, title, description, text string) (string, string) {
	excerpt := runePrefix(text, LLMExcerptChars)
	if excerpt == "" {
		excerpt = "(no page text)"
	}
	if title == "" {
		title = "(unknown)"
	}
	if description == "" {
		description = "(none)"
	}
	rules := p.tax.JudgmentRules()
	if rules == "" {
		rules = "(none)"
	}
	prompt, err := prompts.LoadPrompt("w01_category",
		prompts.Slot{Name: "CATEGORY_NAMES_WITH_ONE_LINE_DEFINITIONS", Value: p.tax.CategoryOneLiner()},
		prompts.Slot{Name: "JUDGMENT_RULES", Value: rules},
		prompts.Slot{Name: "PAST_CORRECTIONS", Value: p.correctionExamples(u)},
		prompts.Slot{Name: "URL", Value: u},
		prompts.Slot{Name: "TITLE", Value: title},
		prompts.Slot{Name: "META_DESCRIPTION", Value: description},
		prompts.Slot{Name: "TEXT_EXCERPT", Value: excerpt},
	)
	if err != nil {
		p.log(fmt.Sprintf("⚠️ %s: category classification failed after %d attempts (%v) — filing under _review", u, 1+ClassifyRetries, err), "warning")
		return "", ""
	}
	messages := []llmclient.Message{{Role: "user", Content: prompt}}
	lastBad := ""
	for attempt := 0; attempt < 1+ClassifyRetries; attempt++ {
		answer, err := p.llmJSON(messages, "classify")
		if err != nil {
			lastBad = fmt.Sprintf("unparseable answer: %v", err)
			continue
		}
		cat := strings.TrimSpace(helpers.PyStr(answer["category"]))
		conf := strings.ToLower(strings.TrimSpace(helpers.PyStr(answer["confidence"])))
		if p.tax.IsCategory(cat) {
			if conf == "" {
				conf = "medium"
			}
			return cat, conf
		}
		lastBad = fmt.Sprintf("category %q is not in the taxonomy", cat)
		// Corrective retry: tell the model exactly what it did wrong.
		messages = []llmclient.Message{{Role: "user", Content: prompt + fmt.Sprintf(
			"\n\nYour previous answer was rejected: %s. Answer again with the EXACT name of one category from the list, as JSON.", lastBad)}}
	}
	p.log(fmt.Sprintf("⚠️ %s: category classification failed after %d attempts (%s) — filing under _review", u, 1+ClassifyRetries, lastBad), "warning")
	return "", ""
}

// classifySubcategory mirrors _classify_subcategory (SPEC §4.6 pass
// 2): only the chosen category's subcategories. Low confidence drops
// the sub; invalid answers retry with the corrective nudge.
func (p *Pipeline) classifySubcategory(u, category, title, description, text string) (string, string) {
	excerpt := runePrefix(text, LLMExcerptChars)
	if excerpt == "" {
		excerpt = "(no page text)"
	}
	if title == "" {
		title = "(unknown)"
	}
	if description == "" {
		description = "(none)"
	}
	prompt, err := prompts.LoadPrompt("w02_subcategory",
		prompts.Slot{Name: "CATEGORY", Value: category},
		prompts.Slot{Name: "SUBCATEGORIES_WITH_DEFINITIONS", Value: p.tax.SubcategoryOneLiner(category)},
		prompts.Slot{Name: "URL", Value: u},
		prompts.Slot{Name: "TITLE", Value: title},
		prompts.Slot{Name: "META_DESCRIPTION", Value: description},
		prompts.Slot{Name: "TEXT_EXCERPT", Value: excerpt},
	)
	if err != nil {
		return "", ""
	}
	messages := []llmclient.Message{{Role: "user", Content: prompt}}
	for attempt := 0; attempt < 1+ClassifyRetries; attempt++ {
		answer, err := p.llmJSON(messages, "classify")
		if err != nil {
			continue
		}
		sub := strings.TrimSpace(helpers.PyStr(answer["subcategory"]))
		conf := strings.ToLower(strings.TrimSpace(helpers.PyStr(answer["confidence"])))
		if conf == "" {
			conf = "medium"
		}
		if strings.ToLower(sub) == "none" || sub == "" {
			return "", conf
		}
		if p.tax.IsSubcategoryOf(category, sub) {
			return sub, conf
		}
		messages = []llmclient.Message{{Role: "user", Content: prompt + fmt.Sprintf(
			"\n\nYour previous answer %q is not one of the listed subcategories. Answer again with the EXACT name of one subcategory from the list, or \"none\".", sub)}}
	}
	return "", "" // no subcategory — still file under the category
}

// analyze mirrors _analyze (w03: the note fields). Rule: omit rather
// than guess — enforced by the prompt itself; unparseable answers
// raise to the caller.
func (p *Pipeline) analyze(u, title, description, text, category string) (map[string]any, error) {
	// SPEC Appendix A note: when the page text is missing or clearly
	// partial, the model gets only the title and description.
	excerpt := ""
	if len([]rune(text)) < MinTextForAnalysis {
		excerpt = "(no usable page text — classify from title and description)"
	} else {
		excerpt = runePrefix(text, LLMExcerptChars)
	}
	if title == "" {
		title = "(unknown)"
	}
	if description == "" {
		description = "(none)"
	}
	hints := p.tax.TagHintsOf(category)
	if hints == "" {
		hints = "(none)"
	}
	prompt, err := prompts.LoadPrompt("w03_analyze",
		prompts.Slot{Name: "TAG_HINTS", Value: hints},
		prompts.Slot{Name: "URL", Value: u},
		prompts.Slot{Name: "TITLE", Value: title},
		prompts.Slot{Name: "META_DESCRIPTION", Value: description},
		prompts.Slot{Name: "TEXT_EXCERPT", Value: excerpt},
	)
	if err != nil {
		return nil, err
	}
	answer, err := p.llmJSON([]llmclient.Message{{Role: "user", Content: prompt}}, "analyze")
	if err != nil {
		return nil, err
	}
	return answer, nil
}

// -- writing -----------------------------------------------------------

// notePath mirrors _note_path: unique_path under the taxonomy folder.
func (p *Pipeline) notePath(category, subcategory, name string) string {
	rel, err := p.tax.FolderRelpath(category, subcategory)
	if err != nil {
		rel = ""
	}
	fname := storage.SafeFilename(name)
	if fname == "" {
		fname = "untitled-site"
	}
	if !strings.HasSuffix(strings.ToLower(fname), ".md") {
		fname += ".md"
	}
	full := filepath.Join(p.vaultPath, rel, fname)
	return storage.UniquePath(full)
}

// writeNote mirrors _write_note: dry-run records instead of creating
// the folder; the atomic write is gated the same way.
func (p *Pipeline) writeNote(path, content string) {
	_ = dryrun.MkdirAll(filepath.Dir(path), true)
	_ = storage.AtomicWriteText(path, content)
}

// record mirrors _record: persist dedupe + note-state bookkeeping. In
// a dry-run the state DB is the shadow cache, so nothing real is
// recorded.
func (p *Pipeline) record(u, path, content, category, subcategory, fetchStatus string) {
	_ = p.state.MarkProcessed(u, path, category, subcategory, fetchStatus)
	if p.noteStateDB != nil {
		func() {
			defer func() { recover() }() // bookkeeping must never break a run
			c := content
			var catPtr, subPtr *string
			if category != "" {
				catPtr = &category
			}
			if subcategory != "" {
				subPtr = &subcategory
			}
			_ = p.noteStateDB.RecordNote(notestate.VaultWebsites, u, path, &c, catPtr, subPtr, links.NormalizeWebsiteURL)
		}()
	}
}

// appOwnsOldNote mirrors _app_owns_old_note: True when the file at
// oldPath is EXACTLY the app-written note recorded in note_state
// (fingerprint unchanged — no human edit). Only then may the pipeline
// replace or remove it.
func (p *Pipeline) appOwnsOldNote(canonical, oldPath string) bool {
	if p.noteStateDB == nil || oldPath == "" {
		return false
	}
	row, err := p.noteStateDB.RowFor(notestate.VaultWebsites, canonical)
	if err != nil || row == nil || row.Fingerprint == "" {
		return false
	}
	if !samePath(row.Path, oldPath) {
		return false
	}
	data, err := os.ReadFile(oldPath)
	if err != nil {
		return false
	}
	return notestate.ComputeFingerprint(string(data)) == row.Fingerprint
}

// cleanupReplacedReviewNote mirrors _cleanup_replaced_review_note:
// after a successful upgrade, remove the old app-owned _review
// placeholder — a hand-edited placeholder is NEVER removed; Phase 3's
// duplicate detector will surface it instead.
func (p *Pipeline) cleanupReplacedReviewNote(canonical, oldPath string) {
	if oldPath == "" {
		return
	}
	if _, err := os.Stat(oldPath); err != nil {
		return
	}
	if p.appOwnsOldNote(canonical, oldPath) {
		_ = dryrun.Remove(oldPath)
		p.log(fmt.Sprintf("♻️ upgraded note replaced its _review placeholder (%s)", baseName(oldPath)), "info")
	} else {
		p.log(fmt.Sprintf("⚠️ kept the old _review note %q — it looks hand-edited; both files now carry the same source", oldPath), "warning")
	}
}

// consolidateIntoSiteNote mirrors _consolidate_into_site_note
// (v0.64.0 — ONE NOTE PER SITE, the gate's hands): the link's URL
// rides the site note's links list, the ledger row points at the site
// note, any app-owned failed placeholder is swept, and its retry row
// resolves. Returns the SUCCESS-shaped result; never raises — a
// broken consolidation falls back to the honest skip.
func (p *Pipeline) consolidateIntoSiteNote(u, canonical string, result *Result,
	siteNote string, prior *ProcessedRow, upgraded bool) Result {
	cat, sub, status := NotePlacement(siteNote)
	rep := AddSiteLinksToNote(siteNote, []string{canonical}, p.log)
	if !rep.Written && rep.Already == 0 {
		// the note could not take the link — the law cannot hold;
		// fall back to the plain skip so nothing is invented
		result.Outcome = "skipped"
		result.Error = "the site's note could not be extended — the link stays queued for the full flow"
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		return *result
	}
	if prior != nil && prior.FetchStatus != nil && *prior.FetchStatus == "failed" &&
		p.isReviewPath(ptrStr(prior.NotePath)) {
		// the sweep re-reads the file at removal time (the scan's own
		// law): an app-owned failed placeholder leaves, a hand-edited
		// one stays and warns
		p.removeScannedPlaceholder(ptrStr(prior.NotePath))
	}
	_ = p.state.ResolveRetry(canonical)
	content := ""
	if data, err := os.ReadFile(siteNote); err == nil {
		content = string(data)
	}
	p.record(canonical, siteNote, content, cat, sub, status)
	// the site's anchor — a later link of the same site in this batch
	// finds it in the overlay
	if key := links.SiteKeyOf(canonical); key != "" {
		if _, exists := p.batchSiteNotes[key]; !exists {
			p.batchSiteNotes[key] = siteNote
		}
	}
	// v0.64.0 — the 🖐 hand's delivery lands here too.
	p.harvestHandRow(canonical, siteNote, status)
	rel := siteNote
	if p.vaultPath != "" {
		if r, err := filepath.Rel(p.vaultPath, siteNote); err == nil {
			rel = strings.ReplaceAll(r, string(filepath.Separator), "/")
		}
	}
	p.counters["consolidated"]++
	result.Outcome = "processed"
	result.NotePath = siteNote
	result.Category = cat
	result.Subcategory = sub
	result.FetchStatus = status
	result.Error = ""
	p.lastResults = append(p.lastResults, *result)
	p.log(fmt.Sprintf("🧲 [%s] %s: the site's note already holds this domain — the link joined it (one note per site) → %s", p.vaultName, u, rel), "success")
	return *result
}

// -- v0.56.0: THE HAND'S HARVEST — the per-link probes ---------------

// tableRowsCached mirrors _table_rows_cached: the master table's rows
// parsed once per file STATE (mtime+size keyed).
func (p *Pipeline) tableRowsCached() []decomtableRowView {
	if p.vaultPath == "" {
		return nil
	}
	path := DecommissionTablePath(p.vaultPath)
	st, err := os.Stat(path)
	if err != nil {
		return nil
	}
	key := &tableCacheKey{path: path, mtime: st.ModTime().UnixNano(), size: st.Size()}
	if p.handTableKey == nil || *p.handTableKey != *key {
		if lines, ok := readTableLines(path); ok {
			p.handTableKey = key
			p.handTableRows = ParseDecommissionRows(lines)
		} else {
			p.handTableKey = key
			p.handTableRows = nil
		}
	}
	return p.handTableRows
}

// decomtableRowView is the cached-rows view alias.
type decomtableRowView = decomtable.Row

// the status-predicate views over the grammar floor.
func statusDeadView(s string) bool     { return decomtable.StatusIsDead(s) }
func statusReviewedView(s string) bool { return decomtable.StatusIsReviewed(s) }
func statusRevivedView(s string) bool  { return decomtable.StatusIsRevived(s) }

// rowReadsHand mirrors _row_reads_hand: does the master table carry
// the 🖐 gesture for this link? The table's own grammar decides (the
// precedence law; first row wins on a hand-added duplicate). A pure
// probe — never a gate, and a table that cannot be read answers NO.
func (p *Pipeline) rowReadsHand(canonical string) bool {
	if canonical == "" {
		return false
	}
	for _, row := range p.tableRowsCached() {
		if links.NormalizeWebsiteURL(row.URL) != canonical {
			continue
		}
		s := row.Status
		if statusDeadView(s) || statusReviewedView(s) || statusRevivedView(s) {
			return false
		}
		return handdelivery.StatusIsHand(s)
	}
	return false
}

// hasAppReviewNote mirrors _has_app_review_note: does this link own
// an app-written note in _review (ANY fetch_status — the half-fetched
// classes)? The folder scan is cached per folder STATE.
func (p *Pipeline) hasAppReviewNote(canonical string) bool {
	if canonical == "" || p.vaultPath == "" {
		return false
	}
	reviewDir := filepath.Join(p.vaultPath, ReviewFolder)
	st, err := os.Stat(reviewDir)
	if err != nil {
		return false
	}
	key := &reviewCacheKey{mtime: st.ModTime().UnixNano(), size: st.Size()}
	if p.handReviewKey == nil || *p.handReviewKey != *key {
		p.handReviewKey = key
		p.handReviewRows = ScanReviewNotes(p.vaultPath, nil)
	}
	for _, it := range p.handReviewRows {
		if SameSource(it.URL, canonical) {
			return true
		}
	}
	return false
}

// harvestHandRow mirrors _harvest_hand_row (v0.56.0 — THE HAND'S
// HARVEST, the delivery-time half): a link whose master-table row
// carries the 🖐 gesture just landed a PROPER, categorized note, so
// the gesture is retired into the green checkbox and every app-owned
// half-fetched _review item for this source is swept. Never raises;
// never a gate. v0.57.0 — the stamp itself is EARNED by
// NoteIsProperlyStored.
func (p *Pipeline) harvestHandRow(canonical, notePath, fetchStatus string) {
	if p.vaultPath == "" || canonical == "" {
		return
	}
	if !p.rowReadsHand(canonical) {
		return
	}
	proper, _ := NoteIsProperlyStored(&ProcessedRow{
		URL: canonical, NotePath: &notePath, FetchStatus: &fetchStatus,
	})
	if !proper {
		p.log(fmt.Sprintf("🖐 [%s] %s: the note could not be proven properly stored — the row keeps its gesture (the redo pass asks the LLM again; a ✅ is only earned by the note itself)", p.vaultName, canonical), "warning")
		return
	}
	stamped := handdelivery.StampHandDeliveredRows(p.vaultPath, []string{canonical}, p.log)
	swept := SweepReviewLeftovers(p.vaultPath, canonical, notePath, p.log)
	if stamped > 0 || swept > 0 {
		p.handTableKey = nil // the stamp changed the table
		sweptLine := ""
		if swept > 0 {
			sweptLine = fmt.Sprintf(", %d half-fetched _review item(s) swept", swept)
		}
		p.log(fmt.Sprintf("✅ [%s] the hand's harvest: the 🖐 row retired to the green checkbox (✅ hand-delivered)%s — the proper, categorized note is in the vault", p.vaultName, sweptLine), "success")
	}
}

// -- the per-link flow ---------------------------------------------------

// ProcessLink mirrors process_link: run the full per-link pipeline.
// Returns a result (never raises — one failing link never stops the
// batch). The unexpected-exception path leaves a _review note behind
// (SPEC §4.3: nothing is silently dropped) unless this link already
// got one.
func (p *Pipeline) ProcessLink(u string) Result {
	result := &Result{URL: u, Outcome: "failed"}
	panicked := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicked = true
				result.Error = fmt.Sprintf("%v", r)
				result.Outcome = "failed"
			}
		}()
		*result = p.processLinkInner(u, result)
	}()
	if !panicked {
		return *result
	}
	p.log(fmt.Sprintf("❌ Website pipeline error for %s: %s", u, result.Error), "error")
	if result.NotePath == "" {
		func() {
			defer func() { recover() }()
			canonical := result.Canonical
			if canonical == "" {
				canonical = links.NormalizeWebsiteURL(u)
			}
			note := BuildReviewNote(canonical, "failed",
				fmt.Sprintf("Pipeline error: %s", result.Error), "", nil)
			path := p.reviewPath(canonical, nil)
			p.writeNote(path, note)
			p.record(canonical, path, note, "", "", "failed")
			result.Outcome = "review"
			result.NotePath = path
			result.Canonical = canonical
			p.counters["review"]++
		}()
	}
	p.counters["failed"]++
	p.lastResults = append(p.lastResults, *result)
	return *result
}

// directoryBanishGate mirrors _directory_banish_gate (v0.76.0, SPEC
// §11.1): True when the URL hits the directory's banish list (host
// match; shared-host platforms match the full URL). Fail-open on
// every error (a broken list never blocks intake).
func (p *Pipeline) directoryBanishGate(u string) bool {
	defer func() { recover() }() // fail-open, the reference's own law
	vault := strings.TrimSpace(p.vaultPath)
	if vault == "" || !isDirPath(vault) {
		return false
	}
	now := time.Now()
	if p.banishedCache == nil || now.Sub(p.banishedCache.loadedAt) > 60*time.Second {
		banished := directorysync.LoadBanishList(vault)
		p.banishedCache = &directoryBanishedCache{
			loadedAt: now, banished: banished, shared: directorysync.DefaultSharedHosts,
		}
	}
	if len(p.banishedCache.banished) == 0 {
		return false
	}
	return directorysync.IsBanished(u, p.banishedCache.banished, p.banishedCache.shared)
}

// processLinkInner mirrors _process_link_inner — THE STATE MACHINE.
// The result is pointer-passed so the unexpected-panic path (Process-
// Link's catch-all) still sees the canonical this run computed — the
// reference's by-reference dict.
func (p *Pipeline) processLinkInner(u string, result *Result) Result {
	canonical := links.NormalizeWebsiteURL(u)
	result.Canonical = canonical

	// ---- 1a. the DIRECTORY banish gate (v0.76.0, SPEC §11.1) -------
	if p.directoryBanishGate(u) {
		result.Outcome = "skipped"
		result.Error = "banished site (the directory's admin list) — skipped: banished, never fetched"
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		p.log(fmt.Sprintf("🗑️ [%s] %s: skipped: banished (the directory's banish list — Websites/_system/banished.json)", p.vaultName, links.ScrubURLToken(u)), "info")
		return *result
	}

	// ---- 1b. never-fetch domains (v0.20.0 blocked + v0.21.0 self) ---
	if len(p.blockedDomains) > 0 && links.DomainIsBlocked(u, p.blockedDomains) {
		result.Outcome = "skipped"
		result.Error = "blocked domain — omitted (never collected)"
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		p.log(fmt.Sprintf("🚫 [%s] %s: banned domain — omitted (never fetched, never noted, never collected)", p.vaultName, u), "info")
		return *result
	}
	if len(p.selfDomains) > 0 && links.DomainIsSelf(u, p.selfDomains) {
		result.Outcome = "skipped"
		result.Error = "self domain (the app's own bot) — recorded in _inbox only"
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		p.log(fmt.Sprintf("🔒 %s: self domain (the app's own bot) — never fetched, the _inbox row is the record", links.ScrubURLToken(u)), "info")
		return *result
	}

	// ---- 2. dedupe (SPEC §4.3 step 2) --------------------------------
	autoVerdict := ""
	if dismissed, err := p.state.IsDismissed(canonical); err == nil && dismissed {
		result.Outcome = "skipped"
		dreason := ""
		if row, err := p.state.DismissedRow(canonical); err == nil && row != nil {
			dreason = ptrStr(row.Reason)
		}
		if strings.HasPrefix(dreason, "auto-verdict:") {
			autoVerdict = strings.TrimSpace(dreason[len("auto-verdict:"):])
		}
		if p.graveyardUrls[canonical] {
			result.Error = "decommissioned by owner (graveyard) — never fetched again"
		} else if strings.HasPrefix(dreason, BanishReasonPrefix) {
			// v0.58.0 — the banishment: the note left the library AND
			// the URL is blacklisted.
			result.Error = "banished by owner (🗑️) — the note was removed and the URL blacklisted; never fetched again"
		} else if autoVerdict != "" {
			// v0.46.0 — the fetcher's own verdict retired this link.
			result.Error = fmt.Sprintf("retired by auto-verdict (%s) — never fetched again (revive via the graveyard if it returns)", autoVerdict)
		} else {
			result.Error = "dismissed (note was deleted)"
		}
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		if p.graveyardUrls[canonical] {
			p.log(fmt.Sprintf("🪦 %s: decommissioned by owner — skipped (the graveyard table's verdict)", u), "info")
		} else if strings.HasPrefix(dreason, BanishReasonPrefix) {
			p.log(fmt.Sprintf("🗑️ %s: banished by owner — skipped (the note is gone, the URL is blacklisted; ♻️ revived in the master table brings it back)", u), "info")
		} else if autoVerdict != "" {
			p.log(fmt.Sprintf("🪦 %s: skipped — auto-verdict: %s (revive via the graveyard if the link returns)", u, autoVerdict), "info")
		}
		return *result
	}

	inVault := p.vaultIndexHas(canonical) != nil && !isFalsey(p.vaultIndexHas(canonical))
	var prior *ProcessedRow
	if pr, err := p.state.ProcessedRow(canonical); err == nil {
		prior = pr
	}
	// v0.56.0 — THE HAND'S HARVEST, the gate's half-fetched exemption.
	handGesture := p.rowReadsHand(canonical)
	halfFetchedHand := inVault && handGesture && p.hasAppReviewNote(canonical)
	if inVault && !((prior != nil && prior.FetchStatus != nil && *prior.FetchStatus == "failed" &&
		p.isReviewPath(ptrStr(prior.NotePath))) || halfFetchedHand) {
		// Already a real note (or a non-failed _review note): skip.
		result.Outcome = "skipped"
		result.Error = "already in the websites vault"
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		return *result
	}

	// ---- 2b. ONE NOTE PER SITE (v0.64.0) ------------------------------
	siteNote := ""
	if p.siteNoteFor != nil {
		func() {
			defer func() { recover() }()
			siteNote = p.siteNoteFor(canonical)
		}()
	}
	if siteNote == "" {
		// the batch-local overlay (a site this very batch noted)
		if key := links.SiteKeyOf(canonical); key != "" {
			siteNote = p.batchSiteNotes[key]
		}
	}
	if siteNote != "" && !isFilePath(siteNote) {
		siteNote = "" // a stale probe never lies
	}
	if siteNote != "" && p.isReviewPath(siteNote) {
		siteNote = "" // a placeholder is not the site's note
	}
	if siteNote != "" && !handGesture {
		return p.consolidateIntoSiteNote(u, canonical, result, siteNote, prior, inVault && prior != nil)
	}

	var retry *RetryRow
	if r, err := p.state.RetryRow(canonical); err == nil {
		retry = r
	}
	if retry != nil && retry.Attempts >= MaxFetchRetries {
		// v0.56.0 — THE HAND'S REBORN COUNTER: the gesture — or a
		// delivered page already sitting there — reborns ONE counter.
		handReborn := handGesture
		if !handReborn && p.handDelivery && p.vaultPath != "" {
			func() {
				defer func() { recover() }()
				for _, d := range handdelivery.CollectDelivered(p.vaultPath) {
					if SameSource(d.URL, canonical) {
						handReborn = true
						break
					}
				}
			}()
		}
		if handReborn {
			if _, err := p.state.ResetRetryAttempts(canonical); err == nil {
				if r, err := p.state.RetryRow(canonical); err == nil {
					retry = r
				}
				p.log(fmt.Sprintf("🖐 %s: the hand reborn the burned-out retry counter — the gesture (or the delivered page) says finish this one", u), "info")
			} else {
				p.log(fmt.Sprintf("⚠️ could not reborn the retry counter for %s: %v", u, err), "warning")
			}
		}
	}
	if retry != nil && retry.Attempts >= MaxFetchRetries {
		// A failed-fetch _review note already exists and retries are
		// exhausted: leave it, report, never drop silently.
		result.Outcome = "skipped"
		result.Error = fmt.Sprintf("fetch failed %d times — kept in _review, no more retries", retry.Attempts)
		p.counters["skipped"]++
		p.lastResults = append(p.lastResults, *result)
		return *result
	}

	upgraded := inVault && prior != nil &&
		((prior.FetchStatus != nil && *prior.FetchStatus == "failed") ||
			(halfFetchedHand && p.isReviewPath(ptrStr(prior.NotePath))))

	// v0.12.0 — Phase 3 (locked-skip, SPEC §4.4/§6): a note the owner
	// moved by hand is LOCKED — its folder placement beats the
	// classifier.
	var lockedRow *notestate.Row
	if p.noteStateDB != nil {
		func() {
			defer func() { recover() }()
			if row, err := p.noteStateDB.RowFor(notestate.VaultWebsites, canonical); err == nil &&
				row != nil && row.Locked && row.Category != nil && *row.Category != "" {
				lockedRow = row
			}
		}()
	}

	// ---- 3. fetch ------------------------------------------------------
	// v0.48.0 — the fourth door answers FIRST; v0.57.0 — the
	// redo-consume: a gestured link may re-read an ALREADY CONSUMED
	// page (allow_consumed).
	var fetch *webfetch.FetchResult
	if p.handDelivery {
		func() {
			defer func() { recover() }()
			fetch = handdelivery.TakeHandDelivered(p.vaultPath, canonical, p.log, handGesture)
		}()
	}
	if fetch == nil {
		opts := &webfetch.Options{
			TimeoutS:           p.fetchTimeoutS,
			MaxBytes:           p.fetchMaxBytes,
			RateLimiter:        p.rateLimiter,
			NoImpersonate:      !p.impersonateFallback,
			NoArchive:          !p.archiveFallback,
			NoDoHProbe:         !p.dohProbe,
			ImpersonateAutopip: p.impersonateAutopip,
		}
		fetch = p.fetchFn(u, opts)
	} else {
		p.log(fmt.Sprintf("🖐 [%s] %s: %s — the hand-delivered page is the fetch answer", p.vaultName, u, fetch.Reason), "success")
	}
	result.FetchStatus = fetch.Status

	if !fetch.OK() {
		// ---- failure handling: minimal _review note + retry queue -----
		// v0.46.0 — the CATEGORY decides whether time can heal.
		noHeal := fetch.Category == "dead" || fetch.Category == "paywalled" ||
			fetch.Category == "refused" || fetch.Category == "bad_url"
		// v0.48.0 — a WALL the three machine doors could not open names
		// the fourth door in the reason itself.
		if p.handDelivery && handdelivery.IsWalledReason(fetch.Reason) {
			fetch.Reason = fetch.Reason + HandDoorHint
		}
		if noHeal {
			if retry != nil {
				_ = p.state.ResolveRetry(canonical)
			}
			_ = p.state.Dismiss(canonical, fmt.Sprintf("auto-verdict: %s — %s", fetch.Category, fetch.Reason))
		} else {
			_ = p.state.EnqueueRetry(canonical, fetch.Reason)
		}
		note := BuildReviewNote(canonical, "failed", fmt.Sprintf("Fetch failed: %s", fetch.Reason), "", nil)
		// Re-failure: overwrite the app-owned placeholder in place when
		// possible (same path, atomic write) instead of stacking
		// _v1/_v2 duplicates.
		path := p.reviewPath(canonical, prior)
		p.writeNote(path, note)
		p.record(canonical, path, note, "", "", "failed")
		result.Outcome = "review"
		result.NotePath = path
		result.Error = fmt.Sprintf("fetch failed: %s", fetch.Reason)
		if upgraded {
			p.counters["upgraded"]++
		} else {
			p.counters["review"]++
		}
		p.lastResults = append(p.lastResults, *result)
		if noHeal {
			p.log(fmt.Sprintf("🪦 [%s] %s: fetch failed (%s) — minimal note in _review; no retry scheduled (category: %s — the verdict does not heal with time; revive via the graveyard if the link returns)", p.vaultName, u, fetch.Reason, fetch.Category), "warning")
		} else {
			p.log(fmt.Sprintf("📥 [%s] %s: fetch failed (%s) — minimal note in _review, retry scheduled", p.vaultName, u, fetch.Reason), "warning")
		}
		return *result
	}

	if fetch.Category == "archived" {
		// ---- v0.46.0 — an archived rescue files under _review ---------
		page := webextract.ExtractFromBytes(fetch.Body, fetch.Charset)
		reason := fetch.Reason
		if reason == "" {
			reason = "archived copy of a dead page"
		}
		note := BuildReviewNote(canonical, "partial", reason, page.Title, nil)
		path := p.reviewPath(canonical, prior)
		p.writeNote(path, note)
		p.record(canonical, path, note, "", "", "partial")
		if retry != nil {
			_ = p.state.ResolveRetry(canonical)
		}
		result.Outcome = "review"
		result.NotePath = path
		result.Error = reason
		result.FetchStatus = "partial"
		if upgraded {
			p.counters["upgraded"]++
		} else {
			p.counters["review"]++
		}
		p.lastResults = append(p.lastResults, *result)
		p.log(fmt.Sprintf("🗄️ [%s] %s: %s — filed under _review (the live page is gone)", p.vaultName, u, reason), "warning")
		return *result
	}

	// ---- 4. extract ----------------------------------------------------
	page := webextract.ExtractFromBytes(fetch.Body, fetch.Charset)
	fetchStatus := fetch.Status // full | partial (pdf/size)
	if fetchStatus == "full" {
		if page.IsJsShell {
			fetchStatus = "partial"
		} else if page.IsPaywall {
			fetchStatus = "partial"
		}
	}
	title := page.Title
	description := page.MetaDescription

	// ---- 5. classify ---------------------------------------------------
	// v0.64.0 — ONE NOTE PER SITE's soft lock: a 🖐 hand link whose
	// site already holds a real note does not need the classifier
	// either. A locked row (the owner's own placement for THIS link)
	// still wins first.
	siteLock := siteNote != "" && lockedRow == nil
	var category, subcategory string
	if lockedRow != nil {
		// The owner already placed this note — no model call, no
		// _review: their correction IS the classification.
		category = ptrStr(lockedRow.Category)
		subcategory = ptrStr(lockedRow.Subcategory)
		sub := ""
		if subcategory != "" {
			sub = " / " + subcategory
		}
		p.log(fmt.Sprintf("🔒 %s: locked note — owner's placement (%s%s) kept, classifier skipped", u, category, sub), "info")
	} else if siteLock {
		category, subcategory, _ = NotePlacement(siteNote)
		sub := ""
		if subcategory != "" {
			sub = " / " + subcategory
		}
		p.log(fmt.Sprintf("🔗 %s: the site's note is the answer — its placement (%s%s) kept, classifier and analysis skipped (one note per site)", u, category, sub), "info")
	} else {
		var conf1 string
		category, conf1 = p.classifyCategory(canonical, title, description, page.Text)
		if category == "" || conf1 == LowConfidence {
			// Low confidence is NOT a fetch failure — no fetch retry.
			// The note waits in _review for the owner's move (SPEC §4.4).
			reason := "no valid category after retries"
			if category != "" {
				reason = "classification confidence was low"
			}
			note := BuildReviewNote(canonical, fetchStatus, reason, title, nil)
			path := p.reviewPath(canonical, nil)
			p.writeNote(path, note)
			p.record(canonical, path, note, "", "", fetchStatus)
			result.Outcome = "review"
			result.NotePath = path
			result.Error = reason
			result.FetchStatus = fetchStatus
			p.counters["review"]++
			p.lastResults = append(p.lastResults, *result)
			// v0.57.0 — THE NOTE IS THE SUCCESS: for a 🖐 hand link
			// this is NOT a success — the note is half-fetched.
			handLine := ""
			if handGesture {
				handLine = " — NOT properly stored yet: the 🖐 row keeps its gesture and the redo pass will re-read the delivered page and ask the LLM again"
			}
			p.log(fmt.Sprintf("🗂️ [%s] %s: %s — filed under _review%s", p.vaultName, u, reason, handLine), "warning")
			return *result
		}
		var conf2 string
		subcategory, conf2 = p.classifySubcategory(canonical, category, title, description, page.Text)
		if conf2 == LowConfidence {
			subcategory = "" // low-confidence subcategory: category only
		}
	}

	// ---- 6. analyze ----------------------------------------------------
	var analysis map[string]any
	if !siteLock {
		var err error
		analysis, err = p.analyze(canonical, title, description, page.Text, category)
		if err != nil {
			// Analysis failed but classification succeeded: still write
			// a review note — the link must never be silently dropped.
			note := BuildReviewNote(canonical, fetchStatus, fmt.Sprintf("Analysis failed: %v", err), title, nil)
			path := p.reviewPath(canonical, nil)
			p.writeNote(path, note)
			p.record(canonical, path, note, category, subcategory, fetchStatus)
			result.Outcome = "review"
			result.NotePath = path
			result.Error = fmt.Sprintf("analysis failed: %v", err)
			result.Category = category
			result.Subcategory = subcategory
			result.FetchStatus = fetchStatus
			p.counters["review"]++
			p.lastResults = append(p.lastResults, *result)
			if handGesture {
				p.log(fmt.Sprintf("🗂️ [%s] %s: analysis failed — filed under _review — NOT properly stored yet: the 🖐 row keeps its gesture and the redo pass will re-read the delivered page and ask the LLM again", p.vaultName, u), "warning")
			}
			return *result
		}
	}

	// ---- 7. build & write ------------------------------------------------
	if siteNote != "" {
		// v0.64.0 — ONE NOTE PER SITE, the 🖐 hand's delivery: the
		// fetched page answered, the site's note is where the link
		// belongs — it joins the note, no second note is defined.
		return p.consolidateIntoSiteNote(u, canonical, result, siteNote, prior, upgraded)
	}
	name := ""
	if analysis != nil {
		name = helpers.PyStr(analysis["name"])
	}
	if name == "" {
		name = title
	}
	if name == "" {
		name = canonical
	}
	note := BuildWebsiteNote(canonical, analysis, category, subcategory, fetchStatus, nil)
	path := p.notePath(category, subcategory, name)
	p.writeNote(path, note)
	// Upgrade cleanup BEFORE record: the ownership proof compares the
	// note_state row's path with the OLD path.
	if upgraded && prior != nil {
		p.cleanupReplacedReviewNote(canonical, ptrStr(prior.NotePath))
	}
	p.record(canonical, path, note, category, subcategory, fetchStatus)
	// v0.64.0 — ONE NOTE PER SITE: this note is now the site's anchor.
	if key := links.SiteKeyOf(canonical); key != "" {
		if _, exists := p.batchSiteNotes[key]; !exists {
			p.batchSiteNotes[key] = path
		}
	}
	// Full success: any pending fetch-retry for this link is resolved.
	if retry != nil {
		_ = p.state.ResolveRetry(canonical)
	}
	// v0.56.0 — THE HAND'S HARVEST: the proper, categorized note just
	// landed — a link whose row carries the gesture retires it.
	p.harvestHandRow(canonical, path, fetchStatus)

	result.Outcome = "processed"
	result.NotePath = path
	result.Category = category
	result.Subcategory = subcategory
	result.FetchStatus = fetchStatus
	result.Error = ""
	if upgraded {
		p.counters["upgraded"]++
	} else {
		p.counters["processed"]++
	}
	p.lastResults = append(p.lastResults, *result)
	// v0.35.0 — the owner asked the log to SHOW where each link goes.
	rel := path
	if p.vaultPath != "" {
		if r, err := filepath.Rel(p.vaultPath, path); err == nil {
			rel = strings.ReplaceAll(r, string(filepath.Separator), "/")
		}
	}
	statusLine := ""
	if fetchStatus != "full" {
		statusLine = fmt.Sprintf(" (%s)", fetchStatus)
	}
	p.log(fmt.Sprintf("✅ [%s] %s processed and stored → %s%s", p.vaultName, name, rel, statusLine), "success")
	return *result
}

// isFalsey mirrors Python truthiness for the vault_index_has probe's
// answer (False/None/"" → not in vault; a path string or True → in).
func isFalsey(v any) bool {
	switch t := v.(type) {
	case nil:
		return true
	case bool:
		return !t
	case string:
		return t == ""
	}
	return false
}

// reviewPath mirrors _review_path: when prior points at an existing
// app-owned _review placeholder for the same URL, that SAME path is
// returned so the atomic write replaces it (no _v1/_v2 stacking).
func (p *Pipeline) reviewPath(canonical string, prior *ProcessedRow) string {
	if prior != nil && prior.NotePath != nil && *prior.NotePath != "" &&
		p.isReviewPath(*prior.NotePath) && isFilePath(*prior.NotePath) &&
		p.appOwnsOldNote(canonical, *prior.NotePath) {
		return *prior.NotePath
	}
	host := "unknown"
	if parsed, err := parseURLHost(canonical); err == nil {
		host = parsed
	}
	pathPart := ""
	if parsed, err := urlParsePath(canonical); err == nil {
		pathPart = parsed
	}
	fname := storage.SafeFilename(host + pathPart)
	if len(fname) > 120 {
		fname = fname[:120]
	}
	if fname == "" {
		fname = "review"
	}
	if !strings.HasSuffix(strings.ToLower(fname), ".md") {
		fname += ".md"
	}
	return storage.UniquePath(filepath.Join(p.vaultPath, ReviewFolder, fname))
}

// urlParsePath extracts the URL's path component.
func urlParsePath(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	return parsed.Path, nil
}

// isReviewPath mirrors _is_review_path.
func (p *Pipeline) isReviewPath(path string) bool {
	return path != "" && strings.Contains(strings.ReplaceAll(path, "\\", "/"), ReviewFolder)
}

// removeScannedPlaceholder mirrors _remove_scanned_placeholder: remove
// one scanned _review placeholder — but ONLY while it still reads as
// an app-owned fetch-failed note (re-read at removal time). Dry-run
// aware.
func (p *Pipeline) removeScannedPlaceholder(path string) {
	if path == "" {
		return
	}
	if _, err := os.Stat(path); err != nil {
		return
	}
	if !p.isReviewPath(path) {
		return
	}
	fm := ParseReviewFrontmatter(path)
	if fm == nil || !strings.EqualFold(fm.ManagedBy, "gitcurator") ||
		!strings.EqualFold(fm.FetchStatus, "failed") {
		p.log(fmt.Sprintf("⚠️ kept %s — it no longer reads as an app-owned failed placeholder (hand-edited?)", baseName(path)), "warning")
		return
	}
	if err := dryrun.Remove(path); err == nil {
		p.log(fmt.Sprintf("🧹 removed stale _review placeholder %s", baseName(path)), "info")
	} else {
		p.log(fmt.Sprintf("⚠️ could not remove %s: %v", path, err), "warning")
	}
}
