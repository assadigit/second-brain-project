package cli

// settings.go — the CLI's settings plumbing (cli_settings.py): config.json
// load/save (the T-12 config port behind them), the interactive _ask,
// the token mask, the save-cleaner — and THE MODEL-MENU HALF
// (smart_model_pick / prompt_model_menu / apply_model_choice, the
// batch engine's model-picker bridge; the preflight's caller lands
// with T-62, the functions are the ported halves with their own
// differential section).

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/config"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// configPath mirrors _config_path: an explicit --config argument is
// made absolute; the default is APP_DIR/config.json.
func configPath(cliArg string) string {
	if cliArg != "" {
		if abs, err := filepath.Abs(cliArg); err == nil {
			return abs
		}
		return cliArg
	}
	return constants.ConfigFile()
}

// loadConfig mirrors load_config: nil when the file is missing or
// unreadable (the "Could not read config" error line printed first);
// the user-config fill rides internal/config (the same law — only
// still-blank keys).
func (c *CLI) loadConfig(path string) (*storage.OrderedMap, bool) {
	if _, err := os.Stat(path); err != nil {
		return nil, false
	}
	cfg, err := config.Load(path)
	if err != nil {
		c.CliPrint(fmt.Sprintf("Could not read config (%s)", err), "error")
		return nil, false
	}
	if cfg == nil {
		return nil, false
	}
	return cfg, true
}

// saveConfig mirrors save_config: the atomic write + the user-level
// credential mirror (both inside internal/config.Save).
func (c *CLI) saveConfig(path string, cfg *storage.OrderedMap) bool {
	return config.Save(path, cfg)
}

// Ask mirrors _ask: interactive prompt with visible default; Enter
// keeps the default; EOF answers the default. The paint box is off in
// tests (the plain "? … [default] → " shape).
func (c *CLI) Ask(prompt, defaultValue string) string {
	suffix := ""
	if defaultValue != "" {
		suffix = Paint(fmt.Sprintf(" [%s]", defaultValue), cDim)
	}
	fmt.Fprint(c.Out, Paint("? ", cCyan)+prompt+suffix+Paint(" → ", cCyan))
	line, err := c.readLine()
	if err != nil {
		line = ""
	}
	line = pyStrip(line)
	if line != "" {
		return line
	}
	return defaultValue
}

// readLine reads one \n-terminated line from In (EOF-safe).
func (c *CLI) readLine() (string, error) {
	if c.lineReader == nil {
		c.lineReader = bufio.NewReader(c.In)
	}
	return c.lineReader.ReadString('\n')
}

// pyStrip mirrors str.strip(): trims the leading/trailing whitespace
// set Python uses (strings.TrimSpace is the same class for the
// printable plane).
func pyStrip(s string) string { return strings.TrimSpace(s) }

// cleanConfigForSave mirrors _clean_config_for_save: strip CLI-private
// __…__ keys before writing config.json.
func cleanConfigForSave(cfg *storage.OrderedMap) *storage.OrderedMap {
	out := storage.NewOrderedMap()
	for _, k := range cfg.Keys() {
		if strings.HasPrefix(k, "__") && strings.HasSuffix(k, "__") {
			continue
		}
		v, _ := cfg.Get(k)
		out.Set(k, v)
	}
	return out
}

// ---------------------------------------------------------------------------
// THE MODEL MENU (cli_settings.py's model-menu half)
// ---------------------------------------------------------------------------

// SmartModelPick mirrors smart_model_pick: the best stand-in for a
// missing model — delegates to the SAME heuristic the pipeline uses
// (githubpipeline.PickBestModel, the ProcessingWorker's
// _pick_best_model twin) so the menu's recommendation and the headless
// auto-pick always agree. The reference's ImportError fallback (the
// GUI symbol unavailable) has no Go twin — the picker is compiled in.
func SmartModelPick(configured string, available []string) string {
	return githubpipeline.PickBestModel(configured, available)
}

// PromptModelMenu mirrors prompt_model_menu (v0.07.2 — THE model
// picker the CLI was missing): a numbered menu of the models Ollama
// actually has, with the smart recommendation as the default (the
// recommendation is always entry 1; an embedding model's tag overrides
// the recommendation mark, exactly the reference's order). Returns the
// chosen model name, or "" when the user declines (q/quit/n) or EOF
// hits; a NON-TTY stdin never asks — the RECOMMENDED model is returned
// (non-interactive runs must not fail batch-wide over a missing
// model). Five tries, then "" (the give-up law).
func (c *CLI) PromptModelMenu(configured string, available []string,
	status StatusPainter, allowDecline bool) string {
	if len(available) == 0 {
		return ""
	}
	rec := SmartModelPick(configured, available)
	ordered := []string{}
	if containsStr(available, rec) {
		ordered = append(ordered, rec)
	}
	for _, m := range available {
		if m != rec {
			ordered = append(ordered, m)
		}
	}

	if !c.IsATTYIn() {
		return rec // piped / scheduled run: auto-pick, caller logs it
	}

	if status != nil {
		status.EraseIfActive()
	}
	fmt.Fprint(c.Out, "\n")
	c.CliPrint(fmt.Sprintf("Configured model '%s' is not installed.", configured), "warning")
	fmt.Fprint(c.Out, Paint("  Models available in Ollama:", cBold)+"\n")
	for i, m := range ordered {
		tag := ""
		if m == rec {
			tag = Paint("   ← recommended", cGreen)
		}
		if githubpipeline.IsEmbedModel(m) {
			tag = Paint("   (embedding — cannot analyze)", cDim)
		}
		fmt.Fprintf(c.Out, "    %s %s%s\n",
			Paint(strconv.Itoa(i+1)+")", cCyan), m, tag)
	}
	var hint string
	if allowDecline {
		hint = Paint("  (q = abort the run; tip: `ollama pull <model>` installs the original)", cDim)
	} else {
		hint = Paint(fmt.Sprintf("  (tip: `ollama pull %s` installs the original)", configured), cDim)
	}
	fmt.Fprint(c.Out, hint+"\n")
	for try := 0; try < 5; try++ {
		fmt.Fprint(c.Out, Paint("? ", cCyan)+"Select a model"+
			Paint(" [1]", cDim)+Paint(" → ", cCyan))
		raw, err := c.readLine()
		if err != nil && raw == "" {
			return "" // EOFError
		}
		ans := pyStrip(raw)
		if ans == "" {
			return rec
		}
		lowAns := strings.ToLower(ans)
		if allowDecline && (lowAns == "q" || lowAns == "quit" || lowAns == "n") {
			return ""
		}
		if n, perr := strconv.Atoi(ans); perr == nil && pyIsDigits(ans) &&
			n >= 1 && n <= len(ordered) {
			return ordered[n-1]
		}
		hits := []string{}
		for _, m := range ordered {
			if strings.Contains(strings.ToLower(m), lowAns) {
				hits = append(hits, m)
			}
		}
		if len(hits) == 1 {
			return hits[0]
		}
		fmt.Fprint(c.Out, Paint(fmt.Sprintf(
			"  ✖ Not a number or unique name — try again (1–%d or q).", len(ordered)), cRed)+"\n")
	}
	return ""
}

// ApplyModelChoice mirrors apply_model_choice: persist a model switch
// to the in-memory config AND config.json (private keys filtered — one
// source of truth for GUI + CLI). Provider-aware: 'cloud' writes
// cloud_model, 'llamacpp' writes llamacpp_model, else the Ollama dict
// (the raw provider comparison, no normalization — the reference's
// spelling).
func (c *CLI) ApplyModelChoice(cfg *storage.OrderedMap, model, configPath string) {
	provider, _ := getOr(cfg, "llm_provider", "ollama").(string)
	switch provider {
	case "cloud":
		cfg.Set("cloud_model", model)
	case "llamacpp":
		cfg.Set("llamacpp_model", model)
	default:
		oll := getMap(cfg, "ollama")
		oll.Set("model", model)
		cfg.Set("ollama", oll)
	}
	if configPath != "" {
		c.saveConfig(configPath, cleanConfigForSave(cfg))
	}
}
