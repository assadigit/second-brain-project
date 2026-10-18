package cli

// settings.go — the CLI's settings plumbing (cli_settings.py's config
// halves): config.json load/save (the T-12 config port behind them),
// the interactive _ask, the token mask and the save-cleaner.
//
// The model-menu half (smart_model_pick / prompt_model_menu /
// apply_model_choice) rides the batch engine's model-picker bridge —
// T-62's books, per the task list.

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/config"
	"github.com/assadigit/second-brain-project/internal/constants"
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
