package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// botdoors.go — T-79c: the More menu's bot-queue tool doors (the
// pinned reference at 98f4fc7): ✅ Verify All Processed
// (bot_queue.py:1026 verify_all_bot_links) + 📋 Export All Links
// (bot_queue.py:929 export_all_bot_links) + 🔄 Retry Failed
// (dashboard.py:770 retry_failed_repos, the scan half — the batch
// itself rides the shell's SyncRunBatch like every batch door), plus
// the verify door's manual-resolve engines
// (_show_manual_resolve_dialog's Mark as Processed / Decommission /
// Mark ALL actions, bot_queue.py:1368-1496) and the resolve dialog's
// Process Selected leg (_process_missing_urls, bot_queue.py:1355 —
// the shell's own door over the bot-source batch).
//
// THE GUARD LAW (the reference's own, carried exactly — the task
// row's own wording): export/verify take NO busy guard. The
// reference guards only the TELEGRAM lock at these doors (the
// Telethon session's mutual exclusion — Option B's accepted loss, the
// same law that dropped the lock from the Check Queue door at T-74);
// the reference happily runs export/verify alongside a batch, and so
// does the port. The retry door's batch carries SyncRunBatch's own
// one-at-a-time refusal (the reference's _start_worker had NO busy
// check for 'direct' mode — the port's honest refusal is the
// already-recorded improvement, SESSION-035's one-batch law).
//
// THE OPTION-B RE-SCOPES (ADR-0004 Am.3 — recorded divergences, the
// full list in docs/poc-artifacts/t-79c-bot-queue-doors.md):
//   - The data source is the Worker /api/pending poll (the Telethon
//     ALL-messages fetch is gone). The export's "Total messages in
//     bot" renders Σ forward_count over the pending rows — the
//     Worker's own count of the bot's link-bearing deliveries (the
//     Telethon total counted every chat message; the queue is
//     URL-keyed now).
//   - The reference's save-before-verify rides the port's
//     save-on-change law (the config IS the live field; the Bot
//     page's group save would also rewrite the token — no redundant
//     save runs).
//   - The all-accounted verify's "mark all as read?" question is
//     DROPPED (read-acknowledge is the accepted loss — the same law
//     that dropped Mark All as Read from the Bot page; the mirror
//     push is the ack).
//   - The report renders in the More door's themed box (the
//     reference's queue_display is a page surface the More door
//     does not own) with the long-body scroll law; the fetching
//     placeholder line rides the box's completion rendering.
//   - The manual-resolve dialog opens through the report box's own
//     Resolve button (the one-overlay-at-a-time law) instead of
//     automatically on top of the page-surface report.
//   - The cache blocks open ONE connection per door (the reference
//     opens back-to-back connections; the T-79b precedent).
//   - The fuzzy match iterates MirrorIndex's SORTED map (the
//     reference rides its dict's insertion order — the recorded
//     sorted-iteration divergence class).

// ---------------------------------------------------------------------------
// The payloads (THE FIELD-TAG LAW — the tags mirror the keys the
// frontend reads; pinned by TestBotDoorPayloadFieldTags).
// ---------------------------------------------------------------------------

// BotVerifyResult is VerifyAllBotLinks' payload: the report box
// (Title "" = a refusal/failure that was log-only) plus the resolve
// dialog's missing list (empty when every link is accounted for).
type BotVerifyResult struct {
	Title   string   `json:"title"`
	Body    string   `json:"body"`
	Missing []string `json:"missing"`
}

// BotRetryScan is RetryFailedScan's verdict: the box (the vault
// error / the empty queue), the >10 confirm question, and the held
// URL list (the shell's own seat — the confirmed pass re-uses the
// scanned list, the reference's same-urls law).
type BotRetryScan struct {
	Box     *HeroBox `json:"box"`
	Confirm *HeroBox `json:"confirm"`
	URLs    []string `json:"urls"`
}

// BotResolveAction is one manual-resolve action's answer: AllResolved
// true = the dialog's list emptied (the reference's
// url_list.count() == 0 test).
type BotResolveAction struct {
	AllResolved bool `json:"allResolved"`
}

// ---------------------------------------------------------------------------
// The shared helpers.
// ---------------------------------------------------------------------------

// botQueueFetchErr maps the honest poll's state word to the same
// honest message SyncFetchQueue speaks (the doors' "❌ … failed:
// {err}" lines read it).
func botQueueFetchErr(state string) string {
	if state == "disabled" {
		return "the Worker queue is not configured (no worker_base_url or pairing)"
	}
	return "the Worker queue poll failed (network down or the Worker answered an error)"
}

// queueItemForwardCount reads forward_count off a Worker pending row
// (the ledger's own count of the bot's deliveries for the URL; 0 when
// the field is absent — the Worker contract's row shape).
func queueItemForwardCount(item any) int {
	readNum := func(raw any) int {
		switch t := raw.(type) {
		case float64:
			return int(t)
		case int64:
			return int(t)
		case int:
			return t
		}
		return 0
	}
	if m, ok := item.(map[string]any); ok {
		if raw, present := m["forward_count"]; present {
			return readNum(raw)
		}
		return 0
	}
	if om, is := item.(*storage.OrderedMap); is {
		if raw, present := om.Get("forward_count"); present {
			return readNum(raw)
		}
	}
	return 0
}

// fuzzyOwnerRepoRE is the reference's own pair-extraction regex
// (bot_queue.py:1332 / 1422 — no IGNORECASE; the classes carry the
// case; re.match anchors the start).
var fuzzyOwnerRepoRE = regexp.MustCompile(
	`^https?://(?:www\.)?github\.com/([a-zA-Z0-9\-_.]+)/([a-zA-Z0-9\-_.]+)`)

// fuzzyGitHubPair extracts the lowercase owner/repo pair off a URL
// (the .git suffix stripped — the reference's exact law); ok false =
// not a GitHub owner/repo URL.
func fuzzyGitHubPair(url string) (owner, repo string, ok bool) {
	m := fuzzyOwnerRepoRE.FindStringSubmatch(url)
	if m == nil {
		return "", "", false
	}
	owner = strings.ToLower(m[1])
	repo = strings.ToLower(m[2])
	if strings.HasSuffix(repo, ".git") {
		repo = repo[:len(repo)-4]
	}
	return owner, repo, true
}

// fuzzyMatchGitHubURL is _fuzzy_match_github_url (v29.9): extract
// owner/repo off the URL, then check whether ANY vault note's source
// URL carries the same pair (catches renames + normalization
// mismatches — the reference's own example: embedchain/embedchain →
// mem0ai/mem0). Iterates MirrorIndex's sorted map (the recorded
// divergence); guarded — trouble reads as no match, exactly like the
// reference's except-pass.
func fuzzyMatchGitHubURL(vi *batch.VaultIndex, url string) string {
	defer func() {
		if recover() != nil {
			// the probe reads "" on trouble
		}
	}()
	owner, repo, ok := fuzzyGitHubPair(url)
	if !ok {
		return ""
	}
	index := vi.MirrorIndex()
	for _, key := range index.Keys() {
		vOwner, vRepo, vOK := fuzzyGitHubPair(key)
		if vOK && vOwner == owner && vRepo == repo {
			if info, _ := index.Get(key); info != nil {
				if om, is := info.(*storage.OrderedMap); is {
					if p, _ := om.Get("path"); p != nil {
						if s, isStr := p.(string); isStr {
							return s
						}
					}
				}
			}
		}
	}
	return ""
}

// pyStampFile is datetime.now().strftime('%Y%m%d_%H%M%S').
func pyStampFile() string { return time.Now().Format("20060102_150405") }

// pyStampMinute is strftime('%Y-%m-%d %H:%M').
func pyStampMinute() string { return time.Now().Format("2006-01-02 15:04") }

// pyStampSeconds is strftime('%Y-%m-%d %H:%M:%S').
func pyStampSeconds() string { return time.Now().Format("2006-01-02 15:04:05") }

// botQueueFetch runs the Worker pending poll through the door seam
// (the hermetic law: the seam is faked at construction; the
// production wiring IS a.queueFetch). state "ok" returns the items;
// every other word returns the honest error.
func (s *Settings) botQueueFetch() (items []any, errText string) {
	if s.Deps.BotQueueFetch == nil {
		return nil, notWired("queue fetch")
	}
	state, items := s.Deps.BotQueueFetch(s.cfg)
	if state != "ok" {
		return nil, botQueueFetchErr(state)
	}
	return items, ""
}

// botDoorUsernameGuard is the two doors' shared guard (the reference
// reads the page's username field; the port's save-on-change law
// makes the config the same value). Returns true when the door may
// run.
func (s *Settings) botDoorUsernameGuard() bool {
	username := pyStrCfg(s.cfg, "bot_username")
	if line := BotUsernameGuard(username); line != "" {
		s.SpeakOnBus(line, "error")
		return false
	}
	return true
}

// ---------------------------------------------------------------------------
// 📋 Export All Links (bot_queue.py:929 export_all_bot_links).
// ---------------------------------------------------------------------------

// ExportAllBotLinks is export_all_bot_links' flow over the Worker
// pending queue: the username + vault guards, the "no limit" export
// line, the poll, the routed lists, the export file under the vault
// (the reference's exact format), the two verdict lines, and the
// 📋 Links Exported box. Title "" = a refusal or failure that was
// log-only.
func (s *Settings) ExportAllBotLinks() (title, body string) {
	if !s.botDoorUsernameGuard() {
		return "", ""
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" {
		s.SpeakOnBus("❌ No vault selected — need a place to save the export.", "error")
		return "", ""
	}

	s.SpeakOnBus("📋 Exporting ALL links from bot (no limit)...", "info")

	items, errText := s.botQueueFetch()
	if errText != "" {
		s.SpeakOnBus("❌ Export failed: "+errText, "error")
		return "", ""
	}
	urls, nonGithub, rawCount := routeQueueItems(items)
	totalMsgs := 0
	for _, it := range items {
		totalMsgs += queueItemForwardCount(it)
	}
	dups := 0 // the Worker pending list is unique by URL (the render's door remains)

	lines := []string{
		"# 📋 Bot Links Export — " + pyStampMinute(),
		"",
		"## 📊 Statistics",
		fmt.Sprintf("- 📬 Total messages in bot: **%d**", totalMsgs),
		fmt.Sprintf("- 🔗 Total raw links found: **%d**", rawCount),
		fmt.Sprintf("- 🔄 Duplicates removed: **%d**", dups),
		fmt.Sprintf("- ✅ Unique GitHub repos: **%d**", len(urls)),
		fmt.Sprintf("- ✅ Unique non-GitHub links: **%d**", len(nonGithub)),
		fmt.Sprintf("- 📝 Total unique links: **%d**", len(urls)+len(nonGithub)),
		"",
		"## 🐙 GitHub Repos",
		"",
	}
	for i, u := range urls {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, u))
	}
	lines = append(lines, "",
		"## 🔗 Non-GitHub Links",
		"")
	for i, u := range nonGithub {
		lines = append(lines, fmt.Sprintf("%d. %s", i+1, u))
	}
	lines = append(lines, "",
		"---",
		"*Export generated: "+pyStampSeconds()+"*",
		"*Compare this list with what you forwarded to the bot to verify no links are missing.*")

	exportPath := filepath.Join(vault, "_bot_links_export_"+pyStampFile()+".md")
	// The reference writes with a plain open('w') — a bad vault path
	// fails honestly (the atomic wrapper's MkdirAll would paper over
	// the failure the reference exposes at this door).
	if err := os.WriteFile(exportPath, []byte(strings.Join(lines, "\n")), 0o644); err != nil {
		s.SpeakOnBus(fmt.Sprintf("❌ Failed to save export: %v", err), "error")
		return "", ""
	}
	s.SpeakOnBus("📋 Export saved: "+exportPath, "success")
	s.SpeakOnBus(fmt.Sprintf(
		"📊 Total messages: %d | GitHub: %d | Non-GitHub: %d | Duplicates: %d",
		totalMsgs, len(urls), len(nonGithub), dups), "info")
	boxBody := fmt.Sprintf(
		"Total messages: %d\n"+
			"GitHub repos: %d\n"+
			"Non-GitHub: %d\n"+
			"Duplicates removed: %d\n\n"+
			"Export saved to:\n%s\n\n"+
			"Open this file in Obsidian to verify all links are accounted for.",
		totalMsgs, len(urls), len(nonGithub), dups, filepath.Base(exportPath))
	return "📋 Links Exported", boxBody
}

// ---------------------------------------------------------------------------
// ✅ Verify All Processed (bot_queue.py:1026 verify_all_bot_links).
// ---------------------------------------------------------------------------

// VerifyAllBotLinks is verify_all_bot_links' whole flow: the guards,
// the poll, the vault-index ground truth, the ONE-connection cache
// block (the dead set + the processed set + THE SETTLEMENT + the
// settled set), the per-URL ladder (in-vault → decommissioned →
// cache-processed → fuzzy → settled → missing), the missing lines,
// the report (the box body), the permanent _verification_report_*.md
// under the vault, and the summary line. Never takes the app down
// (the reference's own whole-body try/except — the v26 Fix 6 law).
func (s *Settings) VerifyAllBotLinks() (res *BotVerifyResult) {
	res = &BotVerifyResult{}
	defer func() {
		if r := recover(); r != nil {
			// never let Verify All take down the app (the reference's
			// outer except) — the crash speaks + the display gets the
			// crash text (the box's twin).
			s.SpeakOnBus(fmt.Sprintf("❌ Verify All crashed: %v", r), "error")
			res = &BotVerifyResult{
				Title: "Verify All",
				Body:  fmt.Sprintf("Verify All crashed: %v\n\n%s", r, debug.Stack()),
			}
		}
	}()
	if !s.botDoorUsernameGuard() {
		return res
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		s.SpeakOnBus("❌ No vault selected — cannot verify links.", "error")
		return res
	}

	s.SpeakOnBus("✅ Verifying all bot links against vault...", "info")

	items, errText := s.botQueueFetch()
	if errText != "" {
		s.SpeakOnBus("❌ Verify All failed: "+errText, "error")
		res.Title, res.Body = "Verify All", "Error: "+errText
		return res
	}
	urls, nonGithub, _ := routeQueueItems(items)

	// Build the vault index (the ground truth for "is this URL
	// processed?").
	vi := batch.NewVaultIndex(vault, nil)
	vi.Rebuild(nil)
	s.SpeakOnBus(fmt.Sprintf("📚 Vault index: %d notes indexed", vi.Count()), "info")

	// ONE CacheDB connection for the decommissioned set, the
	// processed-URL set, THE SETTLEMENT, and the settled set (the
	// reference's v0.06 perf law). A broken cache reads as the
	// empty sets — the classification still runs (the reference's
	// except).
	decommissioned := map[string]bool{}
	cacheProcessed := map[string]bool{}
	settledRepos := map[string]bool{}
	if cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg)); err == nil {
		decommissioned = cache.GetDeadURLSet(store.DeadLinkThreshold)
		for _, u := range cache.GetAllProcessedURLs() {
			cacheProcessed[u] = true
		}
		// v0.63.3 — THE SETTLED REPOS LEDGER at the verify door too
		// (this door reads the FULL bot history). One meta-guarded
		// batch, then the settled set.
		if rep, serr := cache.SettleExistingRepos(urls); serr == nil && !rep.Already {
			s.SpeakOnBus(fmt.Sprintf(
				"🤝 THE SETTLEMENT (repos): %d repo link(s) the "+
					"bot already delivered are settled — "+
					"addressed and processed, never fetched "+
					"or counted again; %d queued "+
					"retry(ies) cleared. Only repos added "+
					"from now on are counted (♻️ Reset 404 "+
					"Quarantine un-settles any of them)",
				rep.Settled, rep.RetriesCleared), "info")
		}
		settledRepos = cache.GetSettledRepoSet()
		_ = cache.Close()
	}

	inVault := 0
	decomm := 0
	settled := 0
	var missing []string
	for _, url := range urls {
		// the reference's per-URL try/except → missing.
		func() {
			defer func() {
				if recover() != nil {
					missing = append(missing, url)
				}
			}()
			norm := links.NormalizeURL(url)
			if vi.HasURL(url) {
				inVault++
			} else if decommissioned[norm] {
				decomm++
			} else if cacheProcessed[norm] || cacheProcessed[url] {
				// v29.10 — found in CacheDB (processed before, even
				// when the note's source: field drifted after a
				// rename).
				inVault++
				s.SpeakOnBus("✅ Found in cache (processed before): "+url, "info")
			} else if fuzzy := fuzzyMatchGitHubURL(vi, url); fuzzy != "" {
				// v29.9 — the fuzzy match: the same owner/repo pair
				// under a vault note's source URL.
				inVault++
				s.SpeakOnBus(fmt.Sprintf("✅ Fuzzy match: %s → %s", url,
					filepath.Base(fuzzy)), "info")
			} else if settledRepos[norm] {
				// v0.63.3 — addressed without a vault note (the
				// owner's law): accounted for, never re-fetched.
				settled++
			} else {
				missing = append(missing, url)
			}
		}()
	}

	// v29.7/v31.1 — the missing list, so the owner knows exactly what
	// to process (⏳ — pending work, not a failure).
	if len(missing) > 0 {
		s.SpeakOnBus(fmt.Sprintf("⏳ %d missing GitHub link(s):", len(missing)), "warning")
		for i, u := range missing {
			s.SpeakOnBus(fmt.Sprintf("   %d. %s", i+1, u), "info")
		}
	}

	// The human-readable report (the queue_display text's own
	// grammar, byte-parity).
	report := []string{
		"✅ VERIFICATION REPORT",
		strings.Repeat("=", 60),
		fmt.Sprintf("Total GitHub links in bot:      %d", len(urls)),
		fmt.Sprintf("✅ Found in vault:              %d", inVault),
		fmt.Sprintf("🗑️ Decommissioned (404):        %d", decomm),
	}
	if settled > 0 {
		report = append(report, fmt.Sprintf(
			"🤝 Settled (addressed):         %d  (never re-fetched)", settled))
	}
	report = append(report,
		fmt.Sprintf("⏳ Missing from vault:          %d", len(missing)),
		fmt.Sprintf("🔗 Non-GitHub links:            %d", len(nonGithub)),
		"",
		"ALL GITHUB LINKS — DETAILED CHECK:",
		strings.Repeat("-", 60))
	for i, url := range urls {
		func() {
			defer func() {
				if recover() != nil {
					report = append(report, fmt.Sprintf("  %3d. ❌ %s (check error)", i+1, url))
				}
			}()
			if vi.HasURL(url) {
				notePath := vi.GetPath(url)
				noteName := "?"
				if notePath != "" {
					noteName = filepath.Base(notePath)
				}
				report = append(report,
					fmt.Sprintf("  %3d. ✅ %s", i+1, url),
					fmt.Sprintf("       → %s", noteName))
			} else {
				report = append(report, fmt.Sprintf("  %3d. ⏳ %s", i+1, url))
			}
		}()
	}
	report = append(report, strings.Repeat("-", 60), "")
	if len(missing) > 0 {
		report = append(report,
			fmt.Sprintf("SUMMARY: %d link(s) need processing!", len(missing)),
			"",
			"To process the missing links:",
			"  1. Click '📬 Check Queue' to load them",
			"  2. Click '🚀 Process All' to process")
	} else {
		report = append(report, "🎉 All GitHub links are in the vault!")
	}
	report = append(report, strings.Repeat("=", 60))
	res.Title = "✅ VERIFICATION REPORT"
	res.Body = strings.Join(report, "\n")
	res.Missing = missing

	// Also save to the vault for the permanent record (best-effort —
	// a failed write never disturbs the report).
	func() {
		defer func() {
			if recover() != nil {
				// the reference's except: pass
			}
		}()
		reportPath := filepath.Join(vault, "_verification_report_"+pyStampFile()+".md")
		var b strings.Builder
		fmt.Fprintf(&b, "# ✅ Verification Report — %s\n\n", pyStampMinute())
		fmt.Fprintf(&b, "## 📊 Summary\n\n")
		fmt.Fprintf(&b, "| Metric | Count |\n|--------|-------|\n")
		fmt.Fprintf(&b, "| 📬 Total GitHub links in bot | %d |\n", len(urls))
		fmt.Fprintf(&b, "| ✅ Found in vault | %d |\n", inVault)
		fmt.Fprintf(&b, "| ⏳ Missing from vault | %d |\n", len(missing))
		fmt.Fprintf(&b, "| 🔗 Non-GitHub links | %d |\n\n", len(nonGithub))
		fmt.Fprintf(&b, "## 📋 Detailed Check\n\n")
		fmt.Fprintf(&b, "| # | Status | URL | Note |\n")
		fmt.Fprintf(&b, "|---|--------|-----|------|\n")
		for i, url := range urls {
			if vi.HasURL(url) {
				notePath := vi.GetPath(url)
				noteName := "?"
				if notePath != "" {
					noteName = filepath.Base(notePath)
				}
				fmt.Fprintf(&b, "| %d | ✅ | %s | %s |\n", i+1, url, noteName)
			} else {
				fmt.Fprintf(&b, "| %d | ⏳ | %s | — |\n", i+1, url)
			}
		}
		fmt.Fprintf(&b, "\n---\n*Report generated: %s*\n", pyStampSeconds())
		// The plain-open twin (the same law as the export file).
		if err := os.WriteFile(reportPath, []byte(b.String()), 0o644); err != nil {
			return
		}
		s.SpeakOnBus("📄 Verification report saved to vault: "+filepath.Base(reportPath), "success")
	}()

	// The summary line. The reference's mark-all-read question after
	// a clean verify is DROPPED under Option B (read-acknowledge is
	// the accepted loss — the mirror push is the ack).
	if len(missing) > 0 {
		s.SpeakOnBus(fmt.Sprintf(
			"✅ Verified: %d/%d GitHub links in vault (%d missing)",
			inVault, len(urls), len(missing)), "warning")
	} else {
		s.SpeakOnBus(fmt.Sprintf(
			"✅ Verified: %d in vault + %d decommissioned + %d settled = %d/%d accounted for",
			inVault, decomm, settled, inVault+decomm+settled, len(urls)), "success")
	}
	return res
}

// ---------------------------------------------------------------------------
// 🔄 Retry Failed (dashboard.py:770 retry_failed_repos — the scan
// half; the batch is the shell's).
// ---------------------------------------------------------------------------

// RetryFailedScan is retry_failed_repos' everything-but-the-batch
// flow: the vault guard (the box), the retry-queue read (the honest
// read-failure line), the empty verdict (the line + the box), the
// settled-skip layer (the ledger's second guard — the owner's law),
// the re-empty verdict, and the >10 confirm question. URLs carries
// the kept list for the shell's batch.
func (s *Settings) RetryFailedScan() *BotRetryScan {
	out := &BotRetryScan{}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		out.Box = &HeroBox{Title: "Error",
			Body: "Please select a valid Obsidian vault path."}
		return out
	}

	var failed []store.FailedRow
	if cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg)); err == nil {
		failed = cache.GetFailedURLs()
		// v0.63.3 — THE SETTLED REPOS LEDGER's door on the retry
		// button: a settled repo (addressed — the owner's law) is
		// never retried, whatever stale row sits in the queue. ONE
		// connection for both reads (the recorded divergence — the
		// reference opens a second).
		var kept []string
		var settledSkips []string
		for _, row := range failed {
			if row.URL == "" {
				continue
			}
			if cache.IsRepoSettled(row.URL) {
				settledSkips = append(settledSkips, row.URL)
			} else {
				kept = append(kept, row.URL)
			}
		}
		_ = cache.Close()
		if len(settledSkips) > 0 {
			s.SpeakOnBus(fmt.Sprintf(
				"🤝 %d settled repo link(s) left "+
					"the retry queue — addressed (the owner's law), never "+
					"re-fetched (♻️ Reset 404 Quarantine un-settles any of "+
					"them).", len(settledSkips)), "info")
		}
		out.URLs = kept
	} else {
		s.SpeakOnBus(fmt.Sprintf("❌ Failed to read retry queue: %v", err), "error")
		return out
	}

	if len(out.URLs) == 0 {
		s.SpeakOnBus("✓ No failed repos to retry.", "success")
		out.Box = &HeroBox{Title: "Retry Queue Empty",
			Body: "No failed repos to retry. 🎉"}
		return out
	}

	// v31.1 safety gate: confirm before large batches (>10 items).
	if ask, title, body := ConfirmLargeBatch(len(out.URLs), "the retry queue"); ask {
		out.Confirm = &HeroBox{Title: title, Body: body, Question: true}
	}
	return out
}

// ---------------------------------------------------------------------------
// The manual-resolve engines (bot_queue.py:1368-1496 — the verify
// dialog's actions; every write settles the ledger too, v0.63.3's
// "the owner's verdict settles" law).
// ---------------------------------------------------------------------------

// BotResolveMarkProcessed is the Mark as Processed button: add the
// processed row (the "(manual resolve)" note + the "Manual" category)
// + settle the ledger + the verdict line. missing is the dialog's
// CURRENT list (the reference removes the row, then the empty list
// speaks the all-resolved line).
func (s *Settings) BotResolveMarkProcessed(url string, missing []string) *BotResolveAction {
	return s.botResolveSettle(url, missing, "processed")
}

// BotResolveDecommission is the Decommission button: the
// manual-decommission record + the settle + the verdict line.
func (s *Settings) BotResolveDecommission(url string, missing []string) *BotResolveAction {
	return s.botResolveSettle(url, missing, "decommission")
}

func (s *Settings) botResolveSettle(url string, missing []string, kind string) *BotResolveAction {
	out := &BotResolveAction{}
	cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		// the reference's except: pass — the write never disturbs
		// the dialog.
		return out
	}
	defer func() { _ = cache.Close() }()
	if kind == "decommission" {
		cache.Decommission(url, "Manual decommission by user")
	} else {
		owner, repo, _ := fuzzyGitHubPair(url)
		_ = cache.AddProcessed(0, url, owner, repo, "(manual resolve)", "Manual")
	}
	_, _ = cache.SettleRepos([]string{url})
	if kind == "decommission" {
		s.SpeakOnBus("🗑️ Decommissioned (manual): "+url, "info")
	} else {
		s.SpeakOnBus("✅ Marked as processed (manual): "+url, "success")
	}
	remaining := 0
	for _, u := range missing {
		if u != url {
			remaining++
		}
	}
	if remaining == 0 {
		out.AllResolved = true
		s.SpeakOnBus("✅ All missing links resolved!", "success")
	}
	return out
}

// BotResolveMarkAll is the Mark ALL as Processed button: the loop +
// the bulk settle + the count line (the dialog closes — the whole
// list resolved).
func (s *Settings) BotResolveMarkAll(missing []string) *BotResolveAction {
	out := &BotResolveAction{}
	cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg))
	if err != nil {
		return out
	}
	defer func() { _ = cache.Close() }()
	for _, url := range missing {
		owner, repo, _ := fuzzyGitHubPair(url)
		_ = cache.AddProcessed(0, url, owner, repo, "(manual resolve)", "Manual")
	}
	_, _ = cache.SettleRepos(missing)
	s.SpeakOnBus(fmt.Sprintf(
		"✅ Marked %d URL(s) as processed (manual)", len(missing)), "success")
	out.AllResolved = true
	return out
}
