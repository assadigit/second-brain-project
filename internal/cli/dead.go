package cli

// dead.go — the 404-quarantine commands (cli.py cmd_list_dead /
// cmd_reset_dead, ported from the v0.08 companion CLI at the v0.09
// merge): every confirmed-dead link and every in-progress attempt row,
// then the reset door.

import (
	"fmt"
	"path/filepath"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// cacheDBPath is the CacheDB() default: APP_DIR/cache.db (the v0.09.4
// split-state fix — never CWD-relative).
func cacheDBPath() string {
	return filepath.Join(constants.AppDir(), "cache.db")
}

// cmdListDead mirrors cmd_list_dead: list every CONFIRMED-dead link
// (attempts >= threshold) and every in-progress attempt row, then
// remind how to reset. A missing config file continues with an EMPTY
// config (load_config(path) or {}).
func (c *CLI) cmdListDead(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		cfg = storage.NewOrderedMap()
	}
	cache, err := store.NewCacheDB(cacheDBPath())
	if err != nil {
		c.CliPrint(fmt.Sprintf("Quarantine unavailable (%s)", err), "error")
		return 1
	}
	thr := thresholdFromConfig(cfg)
	dead := cache.GetDeadURLs(thr)
	allRows := cache.GetQuarantineStats()
	cache.Close()

	fmt.Fprint(c.Out, Paint(fmt.Sprintf("404 quarantine (threshold %d)", thr), cBold)+"\n")
	fmt.Fprint(c.Out, Rule()+"\n")
	if len(allRows) == 0 {
		fmt.Fprint(c.Out, Paint("  ✓ empty — no 404 attempts recorded.", cGreen)+"\n")
		return 0
	}
	confirmed := map[string]bool{}
	for _, r := range dead {
		confirmed[r.URL] = true
	}
	for _, r := range allRows {
		flag := ""
		if confirmed[r.URL] {
			flag = Paint("  ⛔ QUARANTINED", cRed)
		}
		at := ""
		if r.DecommissionedAt != nil {
			at = *r.DecommissionedAt
		}
		reason := "None"
		if r.Reason != nil {
			reason = *r.Reason
		}
		fmt.Fprintf(c.Out, "  %s %s\n", Paint("•", cYellow), r.URL)
		fmt.Fprintf(c.Out, "      %s%s\n",
			Paint(fmt.Sprintf("attempt %d/%d · since %s · %s",
				r.FailCount, thr, pySliceRunes(at, 0, 10), reason), cDim),
			flag)
	}
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint("Reset with: ", cDim)+Paint("--cli --reset-dead", cCyan)+"\n")
	return 0
}

// cmdResetDead mirrors cmd_reset_dead: clear the whole 404 quarantine
// so every link gets a fresh set of attempts. No config is loaded.
func (c *CLI) cmdResetDead(a *Args) int {
	c.Banner()
	cache, err := store.NewCacheDB(cacheDBPath())
	if err != nil {
		c.CliPrint(fmt.Sprintf("Quarantine reset failed (%s)", err), "error")
		return 1
	}
	removed := cache.ResetDeadLinks("")
	cache.Close()
	c.CliPrint(fmt.Sprintf("♻️ 404 quarantine reset — %d link(s) will be processed again.", removed), "success")
	return 0
}
