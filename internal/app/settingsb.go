package app

// settingsb.go — T-74 slice (b): the Settings pages B command surface
// (the engines the Bot, Dashboard, Sources and Backup pages drive; the
// specs matched BY NAME per doc 13's R-0 note: bot_queue.py +
// dashboard.py live in spec-gui-pages-a.md, backup_seal.py in
// spec-gui-pages-b.md).
//
// THE SHAPE (the T-73 settings pattern, extended): the pages ride the
// SAME live-config law (the session-held map, the in-place saves) and
// the SAME ONE merge save — the new groups (Bot, Backup, Quarantine)
// carry only their own keys (the group-pointer law). The job buttons
// (Check Queue, Process All, Backup Now, Seal Now, Resync, Publish,
// Detect Repos, Fast Sync) ride the Commands surface through the shell
// — this file owns their pure halves + the page-local engines:
//
//   · THE QUEUE REPORT RENDERING (check_bot_queue's display builder,
//     bot_queue.py 447-560 — the acceptance item): RenderQueueReport
//     builds the reference's exact labeled-count text from a QueueReport
//     (the canned-result law); QueueBadge answers the ⏳/✅ badge.
//   · THE DASHBOARD ENGINES (dashboard.py 61-311): update_dashboard's
//     first-1000-chars frontmatter scan + the exact report text;
//     refresh_quarantine_view's top-30 table + the threshold line;
//     clear_all_quarantine's reset; undo_last_batch's file list +
//     best-effort deletes + the processed-repos row removal.
//   · THE SOURCES ENGINES (bot_queue.py 72-140): fetch_from_sources'
//     GitHub-URL extraction over an RSS body or a Reddit .json walk,
//     over the injectable SourcesFetch seam (the hermetic law).
//   · THE BACKUP PAGE'S LADDERS + GROUPS (backup_seal.py 204-490 +
//     529-546 + 1175-1273): the three status-dot ladders (pure), the
//     compacted save group (backup_folder, backup_max, backup_enabled,
//     cloudflare_worker_url, vaultseal{…}, goodrepos{…}), and the
//     dashboard-opener door's URL law.
//
// OPTION-B notes (ADR-0004 Am.3 / R-0): the Check Queue door keeps the
// username guard (the page's own field — its refusal line is
// observable) but drops the Telethon credential + proxy-warning guards
// (no MTProto client exists); Process New and Mark All as Read fall
// away (the cursor and read-acknowledge are accepted losses — the
// ledger + the mirror push are the truth). The report line names, the
// badge law, and the settled lines port verbatim.
//
// THE HERMETIC LAW: SourcesFetch + OpenURL are injected seams; the
// dashboard/undo engines read the filesystem over fixture vaults (the
// scan-job tests' own shape); nil seams refuse honestly.

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// ---------------------------------------------------------------------------
// The deps' pages B extension (wired by the DI root; fakeable).
// ---------------------------------------------------------------------------

// pagesBSourcesFetch is the Sources page's network door (urllib's twin:
// a 15 s GET with the browser UA; the seam keeps the tests hermetic).
type pagesBSourcesFetch func(url string) (string, error)

// pagesBOpenURL opens the OS browser at a URL (webbrowser.open's twin;
// best-effort, never raises).
type pagesBOpenURL func(url string)

// defaultSourcesFetch is the production door: the reference's
// urllib.request.urlopen twin (the Mozilla UA, 15 s timeout).
func defaultSourcesFetch(url string) (string, error) {
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 20<<20))
	if err != nil {
		return "", err
	}
	return string(body), nil
}

// defaultOpenURL opens the OS browser (webbrowser.open's twin:
// xdg-open / open / start — best-effort, never raises).
func defaultOpenURL(url string) {
	defer func() { _ = recover() }()
	switch runtimeOS() {
	case "windows":
		_ = execOS("cmd", "/c", "start", "", url)
	case "darwin":
		_ = execOS("open", url)
	default:
		_ = execOS("xdg-open", url)
	}
}

// ---------------------------------------------------------------------------
// The snapshot's pages B fields (one payload builds every page).
// ---------------------------------------------------------------------------

// pagesBSnapshot carries the pages B's initial state.
type pagesBSnapshot struct {
	// Bot
	BotUsername string `json:"botUsername"`
	BotToken    string `json:"botToken"`

	// Dashboard
	StrikeThreshold int `json:"strikeThreshold"`

	// Backup — the Vault Backup group.
	BackupFolder  string `json:"backupFolder"`
	BackupMax     string `json:"backupMax"`
	BackupEnabled bool   `json:"backupEnabled"`
	// Backup — the VaultSeal group.
	VaultSealEnabled bool   `json:"vaultSealEnabled"`
	VaultSealPush    bool   `json:"vaultSealPush"`
	VaultSealRepo    string `json:"vaultSealRepo"`
	// Backup — the GoodRepos group.
	GoodReposEnabled bool   `json:"goodReposEnabled"`
	GoodReposPush    bool   `json:"goodReposPush"`
	GoodReposRepo    string `json:"goodReposRepo"`
	// Backup — the Dashboard group.
	WorkerURL string `json:"workerUrl"`

	// HasVaultToken / HasVaultPath feed the seal ladders' rungs.
	HasVaultPath bool `json:"hasVaultPath"`
	HasGHToken   bool `json:"hasGithubToken"`
}

// snapshotPagesB fills the pages B's fields off the live config.
func (s *Settings) snapshotPagesB(cfg *storage.OrderedMap) *pagesBSnapshot {
	snap := &pagesBSnapshot{
		BotUsername: pyStrCfgDefault(cfg, "bot_username", "githubfetcherbot"),
		BotToken:    pyStrCfg(cfg, "bot_token"),
		StrikeThreshold: func() int {
			n := store.DeadLinkThresholdFromConfig(cfgPlainMap(cfg))
			if n < 2 {
				n = 2
			}
			return n
		}(),
		BackupFolder:     pyStrCfg(cfg, "backup_folder"),
		BackupMax:        "10",
		BackupEnabled:    false,
		VaultSealEnabled: true,
		VaultSealPush:    true,
		VaultSealRepo:    "",
		GoodReposEnabled: true,
		GoodReposPush:    true,
		GoodReposRepo:    "",
		WorkerURL:        pyStrCfg(cfg, "cloudflare_worker_url"),
		HasVaultPath:     strings.TrimSpace(pyStrCfg(cfg, "vault_path")) != "",
		HasGHToken:       strings.TrimSpace(pyStrCfg(cfg, "github_token")) != "",
	}
	if v, ok := cfg.Get("backup_max"); ok {
		if n := pyIntOr(v, 10); n > 0 {
			snap.BackupMax = strconv.Itoa(n)
		}
	}
	if v, ok := cfg.Get("backup_enabled"); ok {
		if b, is := v.(bool); is {
			snap.BackupEnabled = b
		}
	}
	if vs := getMapCfg(cfg, "vaultseal"); vs != nil {
		if v, ok := vs.Get("enabled"); ok {
			if b, is := v.(bool); is {
				snap.VaultSealEnabled = b
			}
		}
		if v, ok := vs.Get("auto_push"); ok {
			if b, is := v.(bool); is {
				snap.VaultSealPush = b
			}
		}
		if v, ok := vs.Get("repo_name"); ok {
			if t, is := v.(string); is {
				snap.VaultSealRepo = t
			}
		}
	}
	if gr := getMapCfg(cfg, "goodrepos"); gr != nil {
		if v, ok := gr.Get("enabled"); ok {
			if b, is := v.(bool); is {
				snap.GoodReposEnabled = b
			}
		}
		if v, ok := gr.Get("auto_push"); ok {
			if b, is := v.(bool); is {
				snap.GoodReposPush = b
			}
		}
		if v, ok := gr.Get("repo_name"); ok {
			if t, is := v.(string); is {
				snap.GoodReposRepo = t
			}
		}
	}
	return snap
}

// pyStrCfgDefault reads a string with the reference's field default
// (ui.py's QLineEdit(self.config.get('bot_username', …)) shape).
func pyStrCfgDefault(cfg *storage.OrderedMap, key, def string) string {
	if cfg == nil {
		return def
	}
	v, ok := cfg.Get(key)
	if !ok || v == nil {
		return def
	}
	if s, is := v.(string); is && s != "" {
		return s
	}
	return def
}

// ---------------------------------------------------------------------------
// The save groups (the group-pointer law — a present group carries its
// own keys; an absent group keeps every config value).
// ---------------------------------------------------------------------------

// BotValues is the Bot page's group (the username + token fields).
type BotValues struct {
	Username string `json:"username"`
	Token    string `json:"token"`
}

// BackupValues is the Backup page's group (backup_seal.py's
// _backup_save_config — every persisted key in one commit).
type BackupValues struct {
	BackupFolder  string `json:"backupFolder"`
	BackupMax     string `json:"backupMax"` // digits → max(1, n); garbage → 10
	BackupEnabled bool   `json:"backupEnabled"`
	WorkerURL     string `json:"workerUrl"` // trailing '/' stripped

	VaultSealEnabled bool   `json:"vaultSealEnabled"`
	VaultSealPush    bool   `json:"vaultSealPush"`
	VaultSealRepo    string `json:"vaultSealRepo"`

	GoodReposEnabled bool   `json:"goodReposEnabled"`
	GoodReposPush    bool   `json:"goodReposPush"`
	GoodReposRepo    string `json:"goodReposRepo"`
}

// QuarantineValues is the Dashboard page's threshold group (the
// spinbox's valueChanged save — merge-safe, live for the next batch).
type QuarantineValues struct {
	StrikeThreshold int `json:"strikeThreshold"`
}

// SaveBotGroup commits the Bot page's fields (one group = one merge).
func (s *Settings) SaveBotGroup(vals *BotValues) *SaveResult {
	if vals == nil {
		vals = &BotValues{}
	}
	ui := storage.NewOrderedMap()
	ui.Set("log_level", "INFO")
	ui.Set("bot_username", strings.TrimPrefix(strings.TrimSpace(vals.Username), "@"))
	ui.Set("bot_token", strings.TrimSpace(vals.Token))
	return s.saveMerge(ui)
}

// SaveBackupGroup commits the Backup page's fields (the reference's
// int-law: max(1, int(text or '10')); a ValueError heals to 10).
func (s *Settings) SaveBackupGroup(vals *BackupValues) *SaveResult {
	if vals == nil {
		vals = &BackupValues{}
	}
	ui := storage.NewOrderedMap()
	ui.Set("log_level", "INFO")
	ui.Set("backup_enabled", vals.BackupEnabled)
	ui.Set("backup_folder", strings.TrimSpace(vals.BackupFolder))
	ui.Set("backup_max", backupMaxLaw(vals.BackupMax))
	ui.Set("cloudflare_worker_url", strings.TrimRight(strings.TrimSpace(vals.WorkerURL), "/"))

	vs := storage.NewOrderedMap()
	vs.Set("enabled", vals.VaultSealEnabled)
	vs.Set("auto_push", vals.VaultSealPush)
	vs.Set("repo_name", strings.TrimSpace(vals.VaultSealRepo))
	ui.Set("vaultseal", vs)

	gr := storage.NewOrderedMap()
	gr.Set("enabled", vals.GoodReposEnabled)
	gr.Set("auto_push", vals.GoodReposPush)
	gr.Set("repo_name", strings.TrimSpace(vals.GoodReposRepo))
	ui.Set("goodrepos", gr)
	return s.saveMerge(ui)
}

// SaveQuarantineGroup commits the Dashboard page's threshold spinbox.
func (s *Settings) SaveQuarantineGroup(vals *QuarantineValues) *SaveResult {
	if vals == nil {
		vals = &QuarantineValues{}
	}
	n := vals.StrikeThreshold
	if n < 2 {
		n = 2
	}
	if n > 10 {
		n = 10
	}
	ui := storage.NewOrderedMap()
	ui.Set("log_level", "INFO")
	ui.Set("notfound_strike_threshold", n)
	return s.saveMerge(ui)
}

// saveMerge is the four-laws' shared half (SaveSettings's tail):
// merge → write → in-place → the 250 ms sentinel re-arm.
func (s *Settings) saveMerge(ui *storage.OrderedMap) *SaveResult {
	s.mu.Lock()
	defer s.mu.Unlock()
	merged := storage.MergeConfig(s.cfg, ui)
	if s.Deps.Write != nil && !s.Deps.Write(merged) {
		return &SaveResult{Saved: false}
	}
	replaceInPlace(s.cfg, merged)
	if s.Deps.SentinelRefresh != nil {
		refresh := s.Deps.SentinelRefresh
		go func() {
			time.Sleep(250 * time.Millisecond)
			func() {
				defer func() { _ = recover() }()
				refresh()
			}()
		}()
	}
	return &SaveResult{Saved: true}
}

// backupMaxLaw mirrors max(1, int(text or '10')) with the ValueError
// heal (backup_seal.py 502-505).
func backupMaxLaw(text string) any {
	raw := strings.TrimSpace(text)
	if raw == "" {
		return 10
	}
	if n, err := strconv.Atoi(raw); err == nil {
		if n < 1 {
			return 1
		}
		return n
	}
	return 10
}

// ---------------------------------------------------------------------------
// THE QUEUE REPORT RENDERING (bot_queue.py 447-560 — the canned-result
// law: the report text builds from a QueueReport, verbatim).
// ---------------------------------------------------------------------------

// QueueBadgeState is the pending badge's two-state law (v31.1: zinc ⏳
// for a pending count — routine, not an error; mint ✅ for zero).
type QueueBadgeState struct {
	Text    string `json:"text"`    // "⏳ N pending" | "✅ 0 pending"
	Pending bool   `json:"pending"` // true = the zinc shape
	Visible bool   `json:"visible"` // the badge appears after Check Queue
}

// QueueBadge builds the badge off a report (the count covers BOTH
// pipelines' pending items — v0.24.1).
func QueueBadge(rep *QueueReport) QueueBadgeState {
	if rep == nil {
		return QueueBadgeState{}
	}
	total := len(rep.PendingURLs) + len(rep.PendingWebsites)
	if total > 0 {
		return QueueBadgeState{Text: fmt.Sprintf("⏳ %d pending", total),
			Pending: true, Visible: true}
	}
	return QueueBadgeState{Text: "✅ 0 pending", Pending: false, Visible: true}
}

// RenderQueueReport builds the Bot page's queue display text
// (check_bot_queue's display builder) from a classified report — the
// acceptance's canned-job-result rendering. websitesOn/websitesVault
// come from the live config (the shell passes them or the defaults
// read here).
func RenderQueueReport(rep *QueueReport) string {
	if rep == nil {
		return "Error: the queue check returned no report"
	}
	if rep.State == "error" {
		return "Error: " + rep.Error
	}
	if rep.State == "disabled" {
		return "Error: " + rep.Error
	}
	websitesOn := !rep.WebsitesOff
	websiteVault := ""
	// The websites twin's context rides the report's own flags.
	var b strings.Builder
	b.WriteString("📬 Bot Queue Results\n")
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString(fmt.Sprintf("Total GitHub URLs in bot:  %d\n", len(rep.GitHubURLs)))
	b.WriteString(fmt.Sprintf("✅ Already in vault:        %d\n", rep.InVault))
	b.WriteString(fmt.Sprintf("🗑️ Decommissioned (404):    %d\n", rep.Decommissioned))
	// v0.63.3 — the settled bucket's own honest line.
	if rep.Settled > 0 {
		b.WriteString(fmt.Sprintf("🤝 Repos settled:           %d  (addressed — never re-fetched)\n", rep.Settled))
	}
	if rep.Blocked > 0 {
		b.WriteString(fmt.Sprintf("🚫 Blocked domains:         %d\n", rep.Blocked))
	}
	if rep.Self > 0 {
		b.WriteString(fmt.Sprintf("🔒 Self domains (own bot):  %d\n", rep.Self))
	}
	b.WriteString(fmt.Sprintf("⏳ Pending (not in vault):  %d\n", len(rep.PendingURLs)))
	b.WriteString(fmt.Sprintf("🔗 Non-GitHub links:        %d\n", len(rep.NonGithub)))
	if websitesOn && websiteVault != "" && len(rep.NonGithub) > 0 {
		b.WriteString(strings.Repeat("=", 50) + "\n")
		b.WriteString(fmt.Sprintf("🌐 WEBSITES pipeline (vault: %s)\n", websiteVault))
		if rep.WebsitesVaultIndex != nil {
			b.WriteString(fmt.Sprintf("📚 Websites vault notes:    %d\n", *rep.WebsitesVaultIndex))
		}
		b.WriteString(fmt.Sprintf("✅ Websites in vault:       %d\n", rep.WebsitesInVault))
		b.WriteString(fmt.Sprintf("✔️ Websites processed:      %d\n", rep.WebsitesProcessed))
		if rep.WebsitesSettled > 0 {
			b.WriteString(fmt.Sprintf("🤝 Websites settled:        %d  (addressed — never re-fetched)\n", rep.WebsitesSettled))
		}
		if rep.WebsitesDismissed > 0 {
			b.WriteString(fmt.Sprintf("🗑️ Websites dismissed:      %d\n", rep.WebsitesDismissed))
		}
		if rep.WebsitesBlocked > 0 {
			b.WriteString(fmt.Sprintf("🚫 Blocked domains:         %d\n", rep.WebsitesBlocked))
		}
		if rep.WebsitesSelf > 0 {
			b.WriteString(fmt.Sprintf("🔒 Self domains (own bot):  %d\n", rep.WebsitesSelf))
		}
		b.WriteString(fmt.Sprintf("⏳ Websites pending:        %d\n", len(rep.PendingWebsites)))
	} else if len(rep.NonGithub) > 0 {
		if rep.WebsitesOff {
			b.WriteString(strings.Repeat("=", 50) + "\n")
			b.WriteString(fmt.Sprintf("⚠️ Websites pipeline is OFF — %d non-GitHub link(s) stay as _inbox rows.\n", len(rep.NonGithub)))
			b.WriteString("   Enable it in Settings → 📁 Vault to process them into the Websites vault.\n")
		} else if rep.WebsitesNoVault {
			b.WriteString(strings.Repeat("=", 50) + "\n")
			b.WriteString("⚠️ Websites pipeline is ON but no Websites vault is set — non-GitHub link(s) stay as _inbox rows.\n")
			b.WriteString("   Pick one in Settings → 📁 Vault.\n")
		}
	}
	if rep.DuplicatesRemoved > 0 {
		b.WriteString(fmt.Sprintf("🔄 Duplicates removed:      %d\n", rep.DuplicatesRemoved))
	}
	b.WriteString(strings.Repeat("=", 50) + "\n\n")

	if len(rep.PendingURLs) > 0 {
		b.WriteString("PENDING GITHUB REPOS (need processing):\n")
		for i, u := range rep.PendingURLs {
			b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, u))
		}
	} else {
		if rep.Settled > 0 {
			b.WriteString(fmt.Sprintf("🎉 All GitHub repos are addressed — %d settled (never re-fetched), the rest in the vault!\n", rep.Settled))
		} else {
			b.WriteString("🎉 All GitHub repos are already in the vault!\n")
		}
		b.WriteString("Click '✅ Verify All Processed' to confirm.\n")
	}

	if websitesOn && !rep.WebsitesNoVault {
		if len(rep.PendingWebsites) > 0 {
			b.WriteString(fmt.Sprintf("\nPENDING WEBSITES (%d — need processing):\n", len(rep.PendingWebsites)))
			shown := rep.PendingWebsites
			if len(shown) > 50 {
				shown = shown[:50]
			}
			for i, u := range shown {
				b.WriteString(fmt.Sprintf("  %d. %s\n", i+1, u))
			}
			if len(rep.PendingWebsites) > len(shown) {
				b.WriteString(fmt.Sprintf("  … and %d more\n", len(rep.PendingWebsites)-len(shown)))
			}
		} else if len(rep.NonGithub) > 0 {
			b.WriteString("\n🎉 All non-GitHub links are processed into the Websites vault!\n")
		}
	} else if len(rep.NonGithub) > 0 && rep.WebsitesOff {
		b.WriteString(fmt.Sprintf("\nNON-GITHUB LINKS (%d):\n", len(rep.NonGithub)))
		b.WriteString("(Recorded in _inbox/ per platform)\n")
	}
	return b.String()
}

// QueueLogLines speaks the check's own log lines (the pending line or
// the caught-up family + the settled lines — bot_queue.py 563-616).
func QueueLogLines(rep *QueueReport) []struct{ Msg, Level string } {
	out := []struct{ Msg, Level string }{}
	if rep == nil {
		return out
	}
	total := len(rep.PendingURLs) + len(rep.PendingWebsites)
	if total > 0 {
		out = append(out, struct{ Msg, Level string }{
			fmt.Sprintf("📬 Queue: %d new repo(s) + %d website link(s) pending (%d repos + %d websites already done)",
				len(rep.PendingURLs), len(rep.PendingWebsites), rep.InVault,
				rep.WebsitesInVault+rep.WebsitesProcessed), "success"})
		return out
	}
	if rep.Settled > 0 {
		out = append(out, struct{ Msg, Level string }{
			fmt.Sprintf("🤝 %d repo link(s) are settled — addressed and processed, never fetched or counted again (the owner's law).", rep.Settled), "info"})
	}
	if rep.WebsitesSettled > 0 {
		out = append(out, struct{ Msg, Level string }{
			fmt.Sprintf("🤝 %d website link(s) are settled — addressed and processed, never fetched or counted again (the owner's law; ♻️ revived in the master table un-settles any of them).", rep.WebsitesSettled), "info"})
	}
	out = append(out, struct{ Msg, Level string }{
		fmt.Sprintf("📬 Queue: all caught up — %d GitHub repos and %d non-GitHub link(s) are already processed.",
			len(rep.GitHubURLs), len(rep.NonGithub)), "success"})
	if rep.WebsitesOff && len(rep.NonGithub) > 0 {
		out = append(out, struct{ Msg, Level string }{
			fmt.Sprintf("ℹ️ %d non-GitHub link(s) are waiting as _inbox rows — turn the Websites pipeline ON (Settings → 📁 Vault) to curate them into the Websites vault.", len(rep.NonGithub)), "info"})
	} else if rep.WebsitesNoVault && len(rep.NonGithub) > 0 {
		out = append(out, struct{ Msg, Level string }{
			"ℹ️ Websites pipeline is ON but no Websites vault is set (Settings → 📁 Vault) — non-GitHub links stay in _inbox.", "info"})
	}
	return out
}

// BotUsernameGuard is the Check Queue door's own refusal (the page's
// field — Option B keeps the username guard, drops the Telethon
// credential guards). "" = the door opens.
func BotUsernameGuard(username string) string {
	if strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(username), "@")) == "" {
		return "❌ Please enter the bot username first."
	}
	return ""
}

// ---------------------------------------------------------------------------
// THE DASHBOARD ENGINES (dashboard.py).
// ---------------------------------------------------------------------------

// DashboardStats is update_dashboard's tally (the first-1000-chars
// frontmatter scan).
type DashboardStats struct {
	TotalNotes     int        `json:"totalNotes"`
	Categories     []CountRow `json:"categories"`
	Languages      []CountRow `json:"languages"`
	AvgCredibility float64    `json:"avgCredibility"`
	Banners        int        `json:"banners"`
	ReviewQueue    int        `json:"reviewQueue"`
}

// CountRow is one category/language line (sorted by count desc, then
// name — Python's sorted(key=lambda x: -x[1]) is stable over the
// insertion order; the port pins the count-desc order).
type CountRow struct {
	Name  string `json:"name"`
	Count int    `json:"count"`
}

// dashboardCatRe / dashboardCredRe / dashboardLangRe mirror the
// reference's frontmatter regexes.
var (
	dashboardCatRe  = regexp.MustCompile(`category:\s*(.+)`)
	dashboardCredRe = regexp.MustCompile(`credibility_score:\s*([\d.]+)`)
	dashboardLangRe = regexp.MustCompile(`primary_language:\s*(.+)`)
)

// ScanDashboardStats walks the vault (only the first 1KB of each .md —
// the speed law) and tallies the frontmatter fields.
func ScanDashboardStats(vault string) *DashboardStats {
	stats := &DashboardStats{}
	cats := map[string]int{}
	langs := map[string]int{}
	var credSum float64
	var credCount int
	if vault == "" {
		return stats
	}
	_ = filepath.Walk(vault, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		fname := info.Name()
		if strings.HasSuffix(fname, ".md") {
			stats.TotalNotes++
			if strings.Contains(p, "_review") {
				stats.ReviewQueue++
			}
			content := readFileHead(p, 1000)
			if m := dashboardCatRe.FindStringSubmatch(content); m != nil {
				cats[strings.TrimSpace(m[1])]++
			}
			if m := dashboardCredRe.FindStringSubmatch(content); m != nil {
				if f, err := strconv.ParseFloat(m[1], 64); err == nil {
					credSum += f
					credCount++
				}
			}
			if m := dashboardLangRe.FindStringSubmatch(content); m != nil {
				langs[strings.TrimSpace(m[1])]++
			}
		} else if strings.HasSuffix(fname, ".png") && strings.Contains(fname, "banner") {
			stats.Banners++
		}
		return nil
	})
	if credCount > 0 {
		stats.AvgCredibility = credSum / float64(credCount)
	}
	stats.Categories = sortedCountRows(cats)
	stats.Languages = sortedCountRows(langs)
	return stats
}

// readFileHead reads at most n bytes (the frontmatter window).
func readFileHead(path string, n int) string {
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer func() { _ = f.Close() }()
	buf := make([]byte, n)
	got, _ := f.Read(buf)
	return string(buf[:got])
}

// sortedCountRows sorts by count desc (stable over the map's walk
// order for ties — the display's own law).
func sortedCountRows(m map[string]int) []CountRow {
	rows := make([]CountRow, 0, len(m))
	for k, v := range m {
		rows = append(rows, CountRow{Name: k, Count: v})
	}
	sort.SliceStable(rows, func(i, j int) bool {
		return rows[i].Count > rows[j].Count
	})
	return rows
}

// RenderDashboardText builds the dashboard panel's text (the exact
// lines, dashboard.py 127-155).
func RenderDashboardText(vault string, stats *DashboardStats) string {
	if stats == nil {
		stats = &DashboardStats{}
	}
	var b strings.Builder
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString("📊 VAULT DASHBOARD\n")
	b.WriteString(strings.Repeat("=", 50) + "\n")
	b.WriteString(fmt.Sprintf("Vault: %s\n", vault))
	b.WriteString(fmt.Sprintf("Total notes: %d\n", stats.TotalNotes))
	b.WriteString(fmt.Sprintf("Banners downloaded: %d\n", stats.Banners))
	b.WriteString(fmt.Sprintf("Review queue: %d\n", stats.ReviewQueue))
	b.WriteString(fmt.Sprintf("Average credibility: %.1f/100\n", stats.AvgCredibility))
	b.WriteString("\n")
	b.WriteString("📁 NOTES BY CATEGORY:\n")
	b.WriteString(strings.Repeat("-", 50) + "\n")
	for _, cat := range stats.Categories {
		b.WriteString(fmt.Sprintf("  %-40s %3d\n", cat.Name, cat.Count))
	}
	b.WriteString("\n")
	b.WriteString("💻 NOTES BY LANGUAGE:\n")
	b.WriteString(strings.Repeat("-", 50) + "\n")
	for _, lang := range stats.Languages {
		b.WriteString(fmt.Sprintf("  %-40s %3d\n", lang.Name, lang.Count))
	}
	b.WriteString("\n")
	b.WriteString(strings.Repeat("=", 50))
	return b.String()
}

// QuarantineView renders the 404 quarantine table (top 30 + the
// threshold footline — refresh_quarantine_view). Reads cache.db (never
// the vault), best-effort.
func QuarantineView(cfg *storage.OrderedMap) string {
	rows := []store.DeadURLRow{}
	if cache, err := store.NewCacheDB(cacheDBPathOf(cfg)); err == nil {
		func() {
			defer func() { recover() }()
			rows = cache.GetQuarantineStats()
		}()
		_ = cache.Close()
	}
	if len(rows) == 0 {
		return "✅ No 404 attempts on record. Deleted repos will appear here after a run reports them missing."
	}
	threshold := store.DeadLinkThresholdFromConfig(cfgPlainMap(cfg))
	var b strings.Builder
	b.WriteString(fmt.Sprintf("%-56s %8s  last seen\n", "URL", "attempts"))
	b.WriteString(strings.Repeat("-", 88) + "\n")
	shown := rows
	if len(shown) > 30 {
		shown = shown[:30]
	}
	for _, row := range shown {
		flag := ""
		if row.FailCount >= int64(threshold) {
			flag = "  ⛔ quarantined"
		}
		lastSeen := ""
		if row.DecommissionedAt != nil {
			lastSeen = pySliceRunes(*row.DecommissionedAt, 0, 16)
		}
		b.WriteString(fmt.Sprintf("%-56s %8d  %s%s\n", row.URL, row.FailCount, lastSeen, flag))
	}
	if len(rows) > 30 {
		b.WriteString(fmt.Sprintf("... and %d more (see CLI --status / --list-dead)\n", len(rows)-30))
	}
	b.WriteString("\n")
	b.WriteString(fmt.Sprintf(
		"Threshold: %d consecutive 404s → the repo is quarantined and skipped. Reset the quarantine to re-check it.", threshold))
	return b.String()
}

// countQuarantineLines counts the table's rendered data rows.
func countQuarantineLines(b *strings.Builder) int {
	return strings.Count(b.String(), "\n") - 2
}

// ClearQuarantine resets every 404 attempt counter
// (clear_all_quarantine's engine half — the confirm dialog is the
// frontend's). Returns the removed count + the log line's level.
func ClearQuarantine(cfg *storage.OrderedMap) (removed int, ok bool) {
	cache, err := store.NewCacheDB(cacheDBPathOf(cfg))
	if err != nil {
		return 0, false
	}
	defer func() { _ = cache.Close() }()
	func() {
		defer func() {
			if recover() != nil {
				ok = false
			}
		}()
		removed = cache.ResetDeadLinks("")
		ok = true
	}()
	return removed, ok
}

// ResetQuarantineLine is the success line (verbatim).
func ResetQuarantineLine(removed int) string {
	return fmt.Sprintf(
		"♻️ 404 quarantine reset — %d link(s) will be processed again (their settlement is gone too — fetched like new).",
		removed)
}

// ---------------------------------------------------------------------------
// undo_last_batch (dashboard.py 228-311).
// ---------------------------------------------------------------------------

// UndoBatchResult is the undo door's story (the dialogs' facts).
type UndoBatchResult struct {
	// Stage answers the early doors: "ok" (the list loaded), "no-file",
	// "empty", "gone" (every file already removed), "no-vault",
	// "read-error".
	Stage   string   `json:"stage"`
	Message string   `json:"message"`
	Files   []string `json:"files"` // the EXISTING files (the confirm list)
	Deleted int      `json:"deleted"`
	Errors  int      `json:"errors"`
	// ConfirmMessage is the themed confirm dialog's body ("" = no
	// confirm needed).
	ConfirmMessage string `json:"confirmMessage"`
}

// ReadUndoList loads `<vault>/_undo_last_batch.txt` (one path per
// line) + the existing-files filter (the dialog's facts).
func ReadUndoList(vault string) *UndoBatchResult {
	if vault == "" || !fsIsDirPath(vault) {
		return &UndoBatchResult{Stage: "no-vault",
			Message: "Please select a valid Obsidian vault path."}
	}
	undoPath := filepath.Join(vault, "_undo_last_batch.txt")
	raw, err := os.ReadFile(undoPath)
	if err != nil {
		return &UndoBatchResult{Stage: "no-file",
			Message: "No _undo_last_batch.txt file found in the vault.\nNothing to undo."}
	}
	var fileList []string
	for _, line := range strings.Split(string(raw), "\n") {
		if line = strings.TrimSpace(line); line != "" {
			fileList = append(fileList, line)
		}
	}
	if len(fileList) == 0 {
		return &UndoBatchResult{Stage: "empty",
			Message: "The undo list is empty. Nothing to undo."}
	}
	var existing []string
	for _, f := range fileList {
		if fileExistsAt(f) {
			existing = append(existing, f)
		}
	}
	if len(existing) == 0 {
		return &UndoBatchResult{Stage: "gone",
			Message: "All files from the last batch have already been removed."}
	}
	// The confirm dialog's body (top 10 + the "..." law).
	var b strings.Builder
	fmt.Fprintf(&b, "This will DELETE %d note file(s) written by the last batch:\n\n", len(existing))
	for i, f := range existing {
		if i >= 10 {
			break
		}
		fmt.Fprintf(&b, "• %s\n", filepath.Base(f))
	}
	if len(existing) > 10 {
		b.WriteString("...\n")
	}
	b.WriteString("\nProceed?")
	return &UndoBatchResult{Stage: "ok", Files: existing, ConfirmMessage: b.String()}
}

// ApplyUndoLastBatch performs the undo (the confirm already answered
// yes): deletes each file best-effort, drops the processed_repos rows
// by note_path, removes the undo file so one batch undoes once, and
// answers the story for the log line.
func ApplyUndoLastBatch(vault string) *UndoBatchResult {
	res := ReadUndoList(vault)
	if res.Stage != "ok" {
		return res
	}
	cache, cerr := store.NewCacheDB(cacheDBPathOf(nil))
	deleted, errs := 0, 0
	for _, fpath := range res.Files {
		if err := os.Remove(fpath); err != nil {
			errs++
			continue
		}
		deleted++
		if cerr == nil {
			func() {
				defer func() { recover() }()
				cache.ForgetProcessedNotePath(fpath)
			}()
		}
	}
	if cerr == nil {
		_ = cache.Close()
	}
	// The undo file goes home either way (one batch, one undo).
	_ = os.Remove(filepath.Join(vault, "_undo_last_batch.txt"))
	res.Deleted = deleted
	res.Errors = errs
	return res
}

// UndoLogLine is the undo's own desk line (verbatim).
func UndoLogLine(deleted, errs int) (string, string) {
	level := "success"
	if errs > 0 {
		level = "warning"
	}
	return fmt.Sprintf("↩️ Undo: deleted %d file(s), %d error(s).", deleted, errs), level
}

// ---------------------------------------------------------------------------
// THE SOURCES ENGINES (bot_queue.py 72-140).
// ---------------------------------------------------------------------------

// SourcesResult is fetch_from_sources' story.
type SourcesResult struct {
	// Stage: "ok" | "no-url" | "error".
	Stage  string   `json:"stage"`
	Error  string   `json:"error"`
	URLs   []string `json:"urls"`
	Body   string   `json:"body"`   // the rendered results panel text
	Notice string   `json:"notice"` // the desk line
	Level  string   `json:"level"`
}

// sourcesGitHubRe mirrors the reference's extraction regex.
var sourcesGitHubRe = regexp.MustCompile(
	`https?://github\.com/([a-zA-Z0-9\-_]+)/([a-zA-Z0-9\-_]+)`)

// FetchFromSources fetches the feed URL + extracts the GitHub URLs
// (the .json Reddit walk when the URL carries .json, the plain body
// regex otherwise). The network door is the injectable seam (nil →
// the production GET).
func (s *Settings) FetchFromSources(url string) *SourcesResult {
	fetch := defaultSourcesFetch
	// the seam door (the hermetic law)
	if s.pagesBSourcesFetchDoor != nil {
		fetch = s.pagesBSourcesFetchDoor
	}
	url = strings.TrimSpace(url)
	if url == "" {
		return &SourcesResult{Stage: "no-url",
			Notice: "Please enter a URL.", Level: "warning"}
	}
	if s.Bus != nil {
		s.Bus.EmitLogLine("🔍 Fetching from: "+url, "info")
	}
	body, err := fetch(url)
	if err != nil {
		msg := pyErrString(err)
		res := &SourcesResult{Stage: "error", Error: msg,
			Body:   "Error: " + msg,
			Notice: fmt.Sprintf("❌ Failed to fetch: %s", msg), Level: "error"}
		s.speakSources(res)
		return res
	}
	urls := extractSourceGitHubURLs(url, body)
	if len(urls) == 0 {
		res := &SourcesResult{Stage: "ok", URLs: nil,
			Body:   "No GitHub URLs found in the source.",
			Notice: "⚠️ No GitHub URLs found in the source.", Level: "warning"}
		s.speakSources(res)
		return res
	}
	display := strings.Join(urls, "\n")
	res := &SourcesResult{Stage: "ok", URLs: urls,
		Body:   fmt.Sprintf("Found %d GitHub URLs:\n\n%s", len(urls), display),
		Notice: fmt.Sprintf("✅ Found %d GitHub URLs from source", len(urls)),
		Level:  "success"}
	s.speakSources(res)
	return res
}

// speakSources rides the result's own desk lines on the bus.
func (s *Settings) speakSources(res *SourcesResult) {
	if s.Bus != nil && res.Notice != "" {
		s.Bus.EmitLogLine(res.Notice, res.Level)
	}
}

// extractSourceGitHubURLs walks the body (the Reddit .json children
// when the URL carries .json; the raw regex otherwise).
func extractSourceGitHubURLs(url, body string) []string {
	var urls []string
	seen := map[string]bool{}
	add := func(owner, repo string) {
		u := fmt.Sprintf("https://github.com/%s/%s", owner, repo)
		if !seen[u] {
			seen[u] = true
			urls = append(urls, u)
		}
	}
	if strings.Contains(url, ".json") {
		// Reddit JSON: children → selftext + url + title.
		walkRedditChildren(body, add)
		return urls
	}
	for _, m := range sourcesGitHubRe.FindAllStringSubmatch(body, -1) {
		add(m[1], m[2])
	}
	return urls
}

// walkRedditChildren walks data.children[].data.{selftext,url,title}
// (a tolerant scan — the reference reads the parsed JSON; the port
// extracts the same three fields from the raw document via the same
// regex, applied to every string value the walker visits).
func walkRedditChildren(body string, add func(owner, repo string)) {
	// The tolerant twin: run the reference's regex over the whole JSON
	// body (selftext/url/title all carry the same link grammar; the
	// unique-set law deduplicates across the three fields).
	for _, m := range sourcesGitHubRe.FindAllStringSubmatch(body, -1) {
		add(m[1], m[2])
	}
}

// ---------------------------------------------------------------------------
// THE BACKUP PAGE'S LADDERS (backup_seal.py — pure).
// ---------------------------------------------------------------------------

// StatusDot is one group's status row (the ● ladder).
type StatusDot struct {
	Text  string `json:"text"`
	State string `json:"state"` // success | warning | error | muted
}

// BackupFolderStatus is the Vault Backup group's dot
// (_backup_refresh_status 529-546).
func BackupFolderStatus(folder string, enabled bool) StatusDot {
	folder = strings.TrimSpace(folder)
	switch {
	case enabled && folder != "":
		if fsIsDirPath(folder) {
			return StatusDot{"● Ready", "success"}
		}
		return StatusDot{"● Folder not found", "error"}
	case enabled:
		return StatusDot{"● No folder selected", "warning"}
	default:
		return StatusDot{"● Disabled", "error"}
	}
}

// VaultSealStatus is the VaultSeal group's dot
// (_vaultseal_refresh_status 1175-1193).
func VaultSealStatus(enabled, hasVault, hasToken bool) StatusDot {
	switch {
	case !enabled:
		return StatusDot{"● Disabled", "error"}
	case !hasVault:
		return StatusDot{"● No vault selected", "warning"}
	case !hasToken:
		return StatusDot{"● Local-only (no GitHub token — commits, no push)", "warning"}
	default:
		return StatusDot{"● Ready — auto-seal after every run", "success"}
	}
}

// GoodReposStatus is the GoodRepos group's dot
// (_goodrepos_refresh_status 1253-1273).
func GoodReposStatus(enabled, hasVault, hasToken bool) StatusDot {
	switch {
	case !enabled:
		return StatusDot{"● Disabled", "error"}
	case !hasVault:
		return StatusDot{"● No vault selected", "warning"}
	case !hasToken:
		return StatusDot{"● Local-only (no GitHub token — README built, no push)", "warning"}
	default:
		return StatusDot{"● Ready — auto-publish after every run", "success"}
	}
}

// BackupNowGuards are the Backup Now button's honest refusals
// (_backup_now 548-560): "" = the door opens; otherwise the themed
// dialog's title + body.
func BackupNowGuards(cfg *storage.OrderedMap) (title, body string) {
	if strings.TrimSpace(pyStrCfg(cfg, "vault_path")) == "" {
		return "No Vault", "Set a vault path first (📁 Vault tab)."
	}
	if strings.TrimSpace(pyStrCfg(cfg, "backup_folder")) == "" {
		return "No Backup Folder", "Select a backup folder first."
	}
	return "", ""
}

// OpenDashboardURL is the Open Dashboard door's URL law
// (_open_dashboard_from_backup_tab): the field's value first, the
// config's second; empty → the honest refusal pair.
func OpenDashboardURL(fieldURL string, cfg *storage.OrderedMap) (url string, refusedTitle, refusedBody string) {
	url = strings.TrimRight(strings.TrimSpace(fieldURL), "/")
	if url == "" {
		url = strings.TrimRight(strings.TrimSpace(pyStrCfg(cfg, "cloudflare_worker_url")), "/")
	}
	if url == "" {
		return "", "No Worker URL",
			"Enter your Worker URL first.\n\nExample: https://github-to-obsidian-bot.your-subdomain.workers.dev"
	}
	return url + "/dashboard", "", ""
}

// runtimeOS / execOS isolate the platform calls for the browser door.
func runtimeOS() string { return runtime.GOOS }

func execOS(name string, args ...string) error {
	return exec.Command(name, args...).Start()
}

// OpenDashboardBrowser is the Open Dashboard button's door: the URL
// law (the field first, the config second, the honest refusal) + the
// browser open (best-effort). Returns the refusal TITLE when the door
// refuses (the shell renders the themed dialog); "" = opened.
func (s *Settings) OpenDashboardBrowser(fieldURL string) (refusedTitle, refusedBody string) {
	url, title, body := OpenDashboardURL(fieldURL, s.cfg)
	if title != "" {
		return title, body
	}
	door := s.pagesBOpenURLDoor
	if door == nil {
		door = defaultOpenURL
	}
	door(url)
	if s.Bus != nil {
		s.Bus.EmitLogLine("📊 Opening the dashboard: "+url, "info")
	}
	return "", ""
}
