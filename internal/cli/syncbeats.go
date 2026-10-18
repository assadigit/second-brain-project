package cli

// syncbeats.go — T-62 slice 2: cmd_sync's LIVE BEATS (cli_sync.py
// 232-381): beat 1b (the connection battery — the Worker leg IS the
// queue check, the Telegram leg is the Option-B accepted loss), beat 2
// (THE BUTTON — landed SESSION-027), beat 3 (the Worker-queue fetch +
// the NON-INTERACTIVE wrong-vault guard + the run). Every failure
// answers the GUI_INSTRUCTION failure card (never a credential ask,
// never a code prompt, never a model menu).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/store"
)

// cmdSyncLive is the full door: CHECK → BUTTON → SYNC.
func (c *CLI) cmdSyncLive(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.SyncFailureCard([]string{fmt.Sprintf(
			"no config at %s — the GUI's Settings writes it on its first save",
			path)}, "")
		return 1
	}
	cfg.Set("__config_path__", path)

	// ---- beat 1a: THE CHECK (the config's own word) ----
	problems := SyncMissingEssentials(cfg, c.IsDir)
	if len(problems) > 0 {
		c.SyncFailureCard(problems, "")
		return 1
	}

	// ---- beat 1b: the live battery — vaults · LLM · GitHub · the
	// queue (the Worker). Warnings speak but never block (the same
	// verdict law); ONE error ends the dance at the failure card. ----
	fmt.Fprint(c.Out, Paint("Sync — checking that everything is in place…", cBold)+"\n")
	fmt.Fprint(c.Out, Paint("config · vaults · LLM · GitHub · the sync queue (the Worker)", cDim)+"\n")
	fmt.Fprint(c.Out, "\n")

	cfgPlain := cfgPlainMap(cfg)
	sections := connectioncheck.RunLocalChecks(cfgPlain,
		func(title string, idx, total int) { c.section(title) },
		func(r connectioncheck.Result) { c.checkLine(cliLevels(r.Level), r.Name, r.Detail) })

	var errors []string
	for _, sec := range sections {
		for _, r := range sec.Results {
			if r.Level == "error" {
				errors = append(errors, strings.TrimSpace(
					fmt.Sprintf("%s: %s", r.Name, r.Detail)))
			}
		}
	}
	fmt.Fprint(c.Out, "\n")
	if len(errors) > 0 {
		c.SyncFailureCard(errors, "")
		return 1
	}

	// ---- the success message (the owner's beat) ----
	c.CliPrint("Everything checks out — ready to SYNC.", "success")
	fmt.Fprint(c.Out, "\n")

	// ---- beat 2: THE BUTTON ----
	if !c.SyncConfirmPrompt(a.Yes, c.IsATTYIn()) {
		c.CliPrint("Cancelled — nothing was synced.", "warning")
		return 0
	}

	// ---- beat 3: THE SYNC (the --auto engine's two phases). ----
	result := c.FetchWorkerQueue(cfg)
	fmt.Fprint(c.Out, "\n")
	if result.State != "ok" {
		c.SyncFailureCard([]string{"the queue fetch failed: " + result.Error}, "")
		return 1
	}

	pending := result.PendingURLs
	inVault := result.InVaultCount
	decomm := result.Decommissioned
	nonGithub := result.NonGithub
	pendingWeb := result.PendingWebsiteURLs

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

	// The wrong-vault guard, NON-INTERACTIVE here: the failure card
	// answers it (the GUI's vault picker fixes it), never a y/N.
	vault := cfgOrStr(cfg, "vault_path")
	cacheProcessed := -1
	if cache, err := store.NewCacheDB(cacheDBPath()); err == nil {
		cacheProcessed = len(cache.GetAllProcessedURLs())
		_ = cache.Close()
	}
	if len(pending) > 0 && inVault == 0 &&
		(result.VaultIndexCount == nil || *result.VaultIndexCount == 0) &&
		cacheProcessed > 0 {
		c.SyncFailureCard([]string{
			fmt.Sprintf("the vault at %s indexed 0 notes while the cache knows "+
				"%d processed repo(s)", vault, cacheProcessed),
			fmt.Sprintf("processing now would duplicate ~all %d link(s) into "+
				"that empty vault", len(pending)),
		}, "")
		return 1
	}

	if len(pending) == 0 && len(pendingWeb) == 0 {
		c.CliPrint("All caught up — nothing to sync. 🎉", "success")
		return 0
	}

	totalBatch := len(pending) + len(pendingWeb)
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint(fmt.Sprintf(
		"SYNC: curating %d new item(s) into the vaults…", totalBatch), cBold)+"\n")
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
		DryRun:           a.DryRun,
	})
}
