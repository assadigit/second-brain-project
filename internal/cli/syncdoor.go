package cli

// syncdoor.go — THE SYNC DOOR's pure halves (cli_sync.py v0.81.0):
// THE FAILURE LAW's one card, THE CHECK's pure half
// (sync_missing_essentials) and THE BUTTON (sync_confirm_prompt).
//
// The door's live half (the check battery + the fetch + the curate
// engine) is T-62's book per the task list; the T-60 slice wires beat
// 1a (the pure check) and renders the honest not-yet card after it.

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// GUIInstruction mirrors GUI_INSTRUCTION — THE FAILURE LAW's one
// remedy (the owner's own words, pinned by the reference's tests).
const GUIInstruction = "Open the GUI version (GitCurator.bat) -> Settings and add the " +
	"credentials and essentials again, then run Sync once more."

// SyncGuiInstruction mirrors sync_gui_instruction.
func SyncGuiInstruction() string { return GUIInstruction }

// SyncFailureCard mirrors sync_failure_card: the red title, each
// offending line, the GUI law. Pure rendering (print only) — every
// failure path in the door ends here, so the owner always meets the
// same door out.
func (c *CLI) SyncFailureCard(problems []string, title string) {
	if title == "" {
		title = "SYNC refused — the setup is not ready:"
	}
	fmt.Fprint(c.Out, "\n")
	c.CliPrint(title, "error")
	for _, line := range problems {
		c.CliPrint("  · "+line, "error")
	}
	fmt.Fprint(c.Out, "\n")
	c.CliPrint(GUIInstruction, "warning")
	fmt.Fprint(c.Out, "\n")
}

// SyncMissingEssentials mirrors sync_missing_essentials: the config's
// own word — what is structurally absent, fast. Pure and
// side-effect-free (isDir injectable for tests). Empty means the
// config has every essential this door needs before any probe runs.
func SyncMissingEssentials(cfg *storage.OrderedMap, isDir func(string) bool) []string {
	if isDir == nil {
		isDir = func(string) bool { return false }
	}
	var problems []string

	for _, kv := range [][2]string{
		{"telegram_api_id", "Telegram api id"},
		{"telegram_api_hash", "Telegram api hash"},
		{"telegram_phone", "Telegram phone"},
	} {
		if strings.TrimSpace(cfgOrStr(cfg, kv[0])) == "" {
			problems = append(problems, kv[1]+" is missing")
		}
	}
	if lstripAt(strings.TrimSpace(cfgOrStr(cfg, "bot_username"))) == "" {
		problems = append(problems, "the bot username is missing (the queue cannot be read)")
	}

	vault := strings.TrimSpace(cfgOrStr(cfg, "vault_path"))
	if vault == "" {
		problems = append(problems, "the vault path is empty")
	} else if !isDir(vault) {
		problems = append(problems, fmt.Sprintf("the vault is not on disk: %s", vault))
	}

	pipes := getMap(cfg, "pipelines")
	websitesOn := pyTruthy(getOr(pipes, "websites", false))
	webVault := strings.TrimSpace(cfgOrStr(cfg, "website_vault_path"))
	if websitesOn {
		if webVault == "" {
			problems = append(problems, "the websites pipeline is ON but its vault path is empty")
		} else if !isDir(webVault) {
			problems = append(problems, fmt.Sprintf("the websites vault is not on disk: %s", webVault))
		}
	}

	provider := cfgProvider(cfg)
	switch {
	case provider == "llamacpp":
		if strings.TrimSpace(cfgOrStr(cfg, "llamacpp_model")) == "" {
			problems = append(problems, "the llama.cpp model is not configured")
		}
	case provider != "ollama":
		if strings.TrimSpace(cfgOrStr(cfg, "cloud_model")) == "" {
			problems = append(problems, "the cloud LLM model is not configured")
		}
		if strings.TrimSpace(cfgOrStr(cfg, "cloud_api_url")) == "" {
			problems = append(problems, "the cloud LLM API URL is not configured")
		}
	default:
		oll := getMap(cfg, "ollama")
		if strings.TrimSpace(mapOrStr(oll, "model")) == "" {
			problems = append(problems, "the Ollama model is not configured")
		}
	}

	return problems
}

// SyncConfirmPrompt mirrors sync_confirm_prompt: THE SYNC BUTTON.
// Enter (an empty answer) syncs; anything else quits. --yes never asks
// (the cron law); a non-terminal stdin never asks (no finger, no
// button — the command's intent is the sync). EOF quits.
func (c *CLI) SyncConfirmPrompt(assumeYes, stdinIsTTY bool) bool {
	if assumeYes || !stdinIsTTY {
		return true
	}
	prompt := Paint("? ", cCyan) + "Press " + Paint("[Enter]", cBold) +
		" to SYNC now — anything else quits: "
	fmt.Fprint(c.Out, prompt)
	line, err := c.readLine()
	if err != nil && line == "" {
		return false // EOFError
	}
	return strings.ToLower(pyStrip(line)) == ""
}
