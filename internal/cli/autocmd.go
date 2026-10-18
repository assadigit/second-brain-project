package cli

// autocmd.go — T-62 slice 2: the batch drivers — cmd_auto (cli.py
// 1252-1440), cmd_retry_failed, and the import-file mode
// (cmd_manual_batch's --import-file branch).
//
// Option-B notes (recorded in spec-cli.md): the fetch is the Worker's
// /api/pending poll (the Telethon bot-chat read is the accepted loss);
// the missing-settings check keeps the reference's own five (the
// wizard's compat law — the Telegram keys are still collected, the
// engine never uses them); the fetch-failure tip that pointed at
// --login is dropped (the login door is itself an accepted loss); and
// bot_queue_max_id (the msg-id cursor) is not passed (mirror-push is
// the ack — the Phase-5 CLEAR re-spec).

import (
	"fmt"
	"os"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// cmdAuto mirrors cmd_auto: preflight → the Worker-queue fetch → the
// classification printout → the wrong-vault guard → the >10 confirm →
// run_batch_visual.
func (c *CLI) cmdAuto(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.CliPrint(fmt.Sprintf("No config at %s. Run %s first.",
			path, Paint("python main.py --cli --init", cBold)), "error")
		return 1
	}
	cfg.Set("__config_path__", path)

	if a.Strikes != nil {
		cfg.Set("notfound_strike_threshold", max2(*a.Strikes))
	}
	if a.Vault != "" {
		cfg.Set("vault_path", a.Vault)
	}

	vault := cfgOrStr(cfg, "vault_path")
	missing := []string{}
	if !(vault != "" && c.IsDir(vault)) {
		missing = append(missing, "vault path")
	}
	for _, kv := range [][2]string{
		{"telegram_api_id", "telegram_api_id"},
		{"telegram_api_hash", "telegram_api_hash"},
		{"telegram_phone", "telegram_phone"},
		{"bot_username", "bot_username"},
	} {
		if strings.TrimSpace(cfgOrStr(cfg, kv[0])) == "" {
			missing = append(missing, kv[1])
		}
	}
	if len(missing) > 0 {
		c.CliPrint("Missing settings: "+Paint(strings.Join(missing, ", "), cRed), "error")
		c.CliPrint("Fix with: "+Paint("python main.py --cli --init", cBold), "info")
		return 1
	}

	// v0.07.2 — the pre-flight BEFORE the fetch: a missing model is
	// caught and picked here, not 1 421 messages later.
	if !c.PreflightChecks(cfg, true) {
		return 1
	}

	fmt.Fprint(c.Out, Paint("Phase 1 — SYNC: fetching undone items from the bot queue…", cBold)+"\n")
	fmt.Fprint(c.Out, "\n")
	result := c.FetchWorkerQueue(cfg)
	fmt.Fprint(c.Out, "\n")

	if result.State != "ok" {
		c.CliPrint("Fetch failed: "+result.Error, "error")
		return 1
	}

	pending := result.PendingURLs
	inVault := result.InVaultCount
	decomm := result.Decommissioned
	nonGithub := result.NonGithub
	vaultIndexCount := result.VaultIndexCount
	pendingWeb := result.PendingWebsiteURLs
	webInVault := result.WebsitesInVault
	webProcessed := result.WebsitesProcessed

	// The queue card.
	pendingColor := cDim
	if len(pending) > 0 {
		pendingColor = cGreen
	}
	webColor := cDim
	if len(pendingWeb) > 0 {
		webColor = cGreen
	}
	fmt.Fprint(c.Out, Paint("Queue: ", cBold)+
		Paint(fmt.Sprintf("%d repo(s) pending", len(pending)), pendingColor)+
		Paint(" · ", cDim)+
		Paint(fmt.Sprintf("%d already in vault", inVault), cDim)+
		Paint(" · ", cDim)+
		Paint(fmt.Sprintf("%d decommissioned", decomm), cDim)+
		Paint(" · ", cDim)+
		Paint(fmt.Sprintf("%d website(s) pending", len(pendingWeb)), webColor)+
		Paint(fmt.Sprintf(" · %d non-GitHub", len(nonGithub)), cDim)+"\n")
	if result.WebsitesVaultIndex != nil {
		fmt.Fprint(c.Out, Paint(fmt.Sprintf(
			"Websites vault — %d note(s) indexed · %d already done",
			*result.WebsitesVaultIndex, webInVault+webProcessed), cDim)+"\n")
	} else if len(nonGithub) > 0 && result.WebsitesPipelineOff {
		c.CliPrint(fmt.Sprintf(
			"%d non-GitHub link(s) are waiting as _inbox rows — the Websites "+
				"pipeline is OFF (config.json: pipelines.websites), so they are "+
				"never curated into the Websites vault.", len(nonGithub)), "info")
	} else if len(nonGithub) > 0 && result.WebsitesNoVault {
		c.CliPrint("Websites pipeline is ON but website_vault_path is empty — "+
			"non-GitHub links stay in _inbox.", "info")
	}

	// v0.09.4 — the dedup ground-truth line.
	cacheProcessed := -1
	if cache, err := store.NewCacheDB(cacheDBPath()); err == nil {
		cacheProcessed = len(cache.GetAllProcessedURLs())
		_ = cache.Close()
	}
	if vaultIndexCount != nil || cacheProcessed >= 0 {
		var parts []string
		if vaultIndexCount != nil {
			parts = append(parts, fmt.Sprintf("vault index: %d note(s)", *vaultIndexCount))
		}
		if cacheProcessed >= 0 {
			parts = append(parts, fmt.Sprintf("cache: %d processed repo(s)", cacheProcessed))
		}
		fmt.Fprint(c.Out, Paint("Dedup ground truth — "+strings.Join(parts, " · "), cDim)+"\n")
	}

	// The wrong-vault guard (1374-1409).
	if c.wrongVaultGuard(cfg, vault, pending, inVault, vaultIndexCount,
		cacheProcessed, a.Yes, true) {
		return 1
	}

	if len(pending) == 0 && len(pendingWeb) == 0 {
		c.CliPrint("All caught up — nothing to process. 🎉", "success")
		return 0
	}

	// The >10 confirm (TTY-only, skipped by --yes).
	totalBatch := len(pending) + len(pendingWeb)
	if !a.Yes && totalBatch > 10 && c.IsATTYIn() {
		ans := c.askPainted(fmt.Sprintf(
			"? %d item(s) ready (%d repo(s) + %d website(s)) — process them all? ",
			totalBatch, len(pending), len(pendingWeb)), "[Y/n] ")
		if strings.HasPrefix(strings.ToLower(ans), "n") {
			c.CliPrint("Cancelled — nothing was processed.", "warning")
			return 0
		}
	}

	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint("Phase 2 — PROCESS: curating into the Obsidian vault…", cBold)+"\n")
	fmt.Fprint(c.Out, "\n")
	rawCount := result.RawURLCount
	if rawCount == 0 {
		rawCount = len(pending) + len(nonGithub)
	}
	return c.RunBatchVisual(cfg, RunBatchOptions{
		Mode:             "direct",
		URLs:             pending,
		BotSource:        true,
		NonGithub:        nonGithub,
		IntakeDuplicates: result.DuplicatesRemoved,
		RawURLCount:      rawCount,
		VaultArg:         a.Vault,
		DryRun:           a.DryRun,
	})
}

// wrongVaultGuard mirrors cmd_auto's 1374-1409 signature stop: pending
// links, ZERO in vault, ZERO notes indexed, a cache that knows
// processed repos — the wrong-vault signature. Interactive mode asks
// y/N (EOF answers "n"); --yes never re-processes (returns true =
// stop). Returns true when the run must stop.
func (c *CLI) wrongVaultGuard(cfg *storage.OrderedMap, vault string,
	pending []string, inVault int, vaultIndexCount *int,
	cacheProcessed int, assumeYes, interactive bool) bool {

	if !(len(pending) > 0 && inVault == 0 &&
		(vaultIndexCount == nil || *vaultIndexCount == 0) &&
		cacheProcessed > 0) {
		return false
	}
	fmt.Fprint(c.Out, "\n")
	c.CliPrint("This looks like a WRONG VAULT PATH, so I stopped before "+
		"processing:", "warning")
	c.CliPrint(fmt.Sprintf(
		"  · the vault at %s indexed 0 notes (no 'source:' frontmatter found)",
		vault), "info")
	c.CliPrint(fmt.Sprintf(
		"  · the cache knows %d already-processed repos — your notes live "+
			"somewhere else", cacheProcessed), "info")
	c.CliPrint(fmt.Sprintf(
		"  · processing now would duplicate ~all %d link(s) into that empty vault",
		len(pending)), "info")
	fmt.Fprint(c.Out, "\n")
	c.CliPrint("Fix: point vault_path at the vault your notes are actually "+
		"written to —", "info")
	c.CliPrint("  · re-run "+Paint("python main.py --cli --init", cBold)+
		" (Enter keeps every current value; fix only the vault), or", "info")
	c.CliPrint("  · run once with "+
		Paint(`--vault C:\path\to\your\real\vault`, cBold)+" to test it, or", "info")
	c.CliPrint(`  · fix "vault_path" in config.json directly.`, "info")
	if assumeYes {
		fmt.Fprint(c.Out, "\n")
		c.CliPrint("(--yes mode never re-processes on this signature. If you "+
			"REALLY want a fresh vault re-run, start the same command "+
			"without --yes and confirm.)", "info")
		return true
	}
	if !interactive {
		return true
	}
	ans := c.askPainted(fmt.Sprintf(
		"Process all %d links into this empty vault anyway? ", len(pending)), "[y/N] ")
	if !strings.HasPrefix(strings.ToLower(ans), "y") {
		c.CliPrint("Cancelled — nothing was processed. Fix vault_path first.",
			"warning")
		return true
	}
	return false
}

// askPainted mirrors the input("? " + text + "[y/N] ") prompt shape
// (EOF answers "").
func (c *CLI) askPainted(text, suffix string) string {
	fmt.Fprint(c.Out, Paint("? ", cCyan)+text+Paint(suffix, cCyan))
	raw, err := c.readLine()
	if err != nil && raw == "" {
		return "" // EOFError → the caller's default
	}
	return strings.ToLower(pyStrip(raw))
}

// cmdRetryFailed mirrors cmd_retry_failed: reprocess every unresolved
// URL in the retry queue (failed_repos).
func (c *CLI) cmdRetryFailed(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.CliPrint(fmt.Sprintf("No config at %s — run --cli --init first.", path), "error")
		return 1
	}
	cfg.Set("__config_path__", path)
	if a.Strikes != nil {
		cfg.Set("notfound_strike_threshold", max2(*a.Strikes))
	}

	if !c.PreflightChecks(cfg, false) {
		return 1
	}

	cache, err := store.NewCacheDB(cacheDBPath())
	if err != nil {
		c.CliPrint(fmt.Sprintf("Cache unavailable: %s", pyErrString(err)), "error")
		return 1
	}
	var failed []string
	for _, row := range cache.GetFailedURLs() {
		failed = append(failed, row.URL)
	}
	_ = cache.Close()
	if len(failed) == 0 {
		c.CliPrint("Retry queue is empty — nothing to reprocess.", "success")
		return 0
	}

	c.CliPrint(fmt.Sprintf("Retrying %d failed URL(s)…", len(failed)), "info")
	return c.RunBatchVisual(cfg, RunBatchOptions{
		Mode:     "direct",
		URLs:     failed,
		VaultArg: a.Vault,
		DryRun:   a.DryRun,
	})
}

// cmdManualBatch mirrors cmd_manual_batch's import-file branch (the
// Telegram ID modes answer refusal cards at the dispatch; the mode
// list card closes the no-mode run).
func (c *CLI) cmdManualBatch(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.CliPrint(fmt.Sprintf("No config at %s — run --cli --init first.", path), "error")
		return 1
	}
	cfg.Set("__config_path__", path)
	if a.Strikes != nil {
		cfg.Set("notfound_strike_threshold", max2(*a.Strikes))
	}

	if !c.PreflightChecks(cfg, false) {
		return 1
	}

	if a.ImportFile != "" {
		fi, err := os.Stat(a.ImportFile)
		if err != nil || fi.IsDir() {
			c.CliPrint(fmt.Sprintf("Import file not found: %s", a.ImportFile), "error")
			return 1
		}
		c.CliPrint(fmt.Sprintf("Mode: import from %s", a.ImportFile), "info")
		return c.RunBatchVisual(cfg, RunBatchOptions{
			Mode:       "import",
			ImportFile: a.ImportFile,
			VaultArg:   a.Vault,
			DryRun:     a.DryRun,
		})
	}

	c.CliPrint("Specify a mode: --auto, --retry-failed, --import-file, "+
		"--single-id, --from-id/--to-id, --offset-start/--count (see --help).", "warning")
	return 1
}

// max2 mirrors max(2, int(x)).
func max2(n int) int {
	if n < 2 {
		return 2
	}
	return n
}
