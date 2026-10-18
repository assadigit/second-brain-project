package cli

// testconn.go — the --test-connection command (cli.py
// cmd_test_connection, 750-807) over the PORTED battery
// (connectioncheck.RunLocalChecks + Summarize) with cli_run.py's own
// renderers (_section 507-515, _check_line 518-521) ported into the
// paint box.
//
// OPTION-B (ADR-0004 Am.3 / rider R-0): the reference's live Telegram
// leg is dropped — the battery's Worker leg IS the queue-reachability
// check; no Telegram leg exists. The card's subtitle and section 4
// speak the Go truth (the reference's card stays in the golden as the
// pre-Option-B RECORD with the divergence asserted): the subtitle says
// "the sync queue (the Worker)" where the reference said "Telegram
// (live)", and the fourth section is titled "Worker" (its one row the
// Worker /health verdict) where the reference's was "Telegram" (its
// local rows + the live line). The exit-code law is unchanged: 0
// unless an error-level issue was found (warnings don't fail).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
)

// section mirrors cli_run.py _section: the full-width ─── title ───
// divider (capped at 118 columns; a title at least as wide as the
// terminal renders bare).
func (c *CLI) section(title string) {
	w := termWidth()
	if w > 118 {
		w = 118
	}
	t := " " + title + " "
	tr := len([]rune(t))
	if tr >= w {
		fmt.Fprint(c.Out, cDim+t+cReset+"\n")
		return
	}
	pad := (w - tr) / 2
	fmt.Fprint(c.Out, cDim+strings.Repeat("─", pad)+t+
		strings.Repeat("─", w-pad-tr)+cReset+"\n")
}

// checkLine mirrors cli_run.py _check_line: one aligned result line —
// "  {icon} {label.ljust(15)} {detail}" with the level's color on the
// icon and the detail (the dead-level-color quirk keeps the icon
// plain; only the detail's paint carries the color — levelStyle
// returns the reference's frozen empty colors).
func (c *CLI) checkLine(level, label, detail string) {
	color, icon := levelStyle(level)
	fmt.Fprintf(c.Out, "  %s %s %s\n",
		Paint(icon, color), Paint(pyLjust(label, 15), cBold),
		Paint(detail, color))
}

// cliLevels mirrors cc.CLI_LEVELS.get(level, "info") — the battery
// level onto cli_print/_check_line's vocabulary (identical to
// GUILevels, the reference keeps both spellings).
func cliLevels(level string) string {
	if v, ok := connectioncheck.GUILevels[level]; ok {
		return v
	}
	return "info"
}

// cmdTestConnection mirrors cmd_test_connection: the battery over the
// loaded config (each section's title divider rendered as it
// completes, each result row as its check line), then the one-line
// VERDICT and the exit-code law.
func (c *CLI) cmdTestConnection(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.CliPrint(fmt.Sprintf("No config at %s — run `--cli --init` first.", path), "warning")
		return 1
	}

	fmt.Fprint(c.Out, Paint("Test Connection — is everything up and ready?", cBold)+"\n")
	// OPTION-B divergence: the reference's subtitle names "Telegram
	// (live)" — the Go battery has no Telegram leg; the Worker leg IS
	// the queue check (the golden asserts the divergence).
	fmt.Fprint(c.Out, Paint("vaults · LLM · GitHub · the sync queue (the Worker)", cDim)+"\n")
	fmt.Fprint(c.Out, "\n")

	onSection := func(title string, idx, total int) {
		c.section(fmt.Sprintf("[%d/%d] %s", idx, total, title))
	}
	onResult := func(r connectioncheck.Result) {
		c.checkLine(cliLevels(r.Level), r.Name, r.Detail)
	}

	sections := connectioncheck.RunLocalChecks(cfg, onSection, onResult)

	// The reference appends the live Telegram leg's line under section
	// 4 here (probe or the credentials-skip line) — DROPPED per
	// ADR-0004 Am.3; the battery's Worker leg is the queue check.

	s := connectioncheck.Summarize(sections)
	fmt.Fprint(c.Out, "\n")
	c.checkLine(cliLevels(s.Level), "VERDICT", s.Headline)
	if s.Level == connectioncheck.LevelError {
		return 1
	}
	return 0
}
