package cli

// wizard.go — the --init first-run wizard (cli.py cmd_init, 108-299):
// the _ask sequence over the existing config (Enter keeps the current
// value; the YOUR_ placeholder counts as unset — the v0.83.0
// empty-start law), THE REPO RADAR before the save (the radar legs ride
// the RadarAll seam; apply/report/pending are the ported pure halves),
// the LLM provider branches (ollama / llamacpp with the probe seam /
// cloud), the proxy block, the 2–10 threshold clamp, save_config and
// the saved box.
//
// OPTION-B (ADR-0004 Am.3 / rider R-0): the wizard COLLECTS the
// Telegram keys for compatibility (the compat law — the engine never
// uses them) but never offers the MTProto login the reference offers on
// a TTY; no login exists to offer. The closing card speaks the honest
// Go-port truth; the golden keeps the reference's non-interactive card
// as the pre-Option-B RECORD with the divergence asserted.
//
// GO NOTES (the int() edges, honest hardening — the happy paths are
// byte-identical): the reference's int(api_id) accepts any
// str.isdigit()-true answer (Unicode digits; superscripts then crash
// int() itself) and int(port) crashes the whole wizard on any
// non-integer answer; the port mirrors the ASCII-decimal paths and
// answers the or-door's default instead of crashing. int("1_0")-style
// underscore literals and arbitrary-precision answers are likewise not
// mirrored (int64); the golden scripts plain decimal answers.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/reporadar"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// defaultRadarAll is the production radar door: RadarAllVaults over
// the token's own read-only API (never raises — the package's own
// never-raises law).
func defaultRadarAll(cfg any) map[string]any {
	token := strings.TrimSpace(cfgOrStr(asOrderedMap(cfg), "github_token"))
	api := reporadar.NewRadarAPI(token, reporadar.APITimeoutSec, reporadar.GITHUBAPI)
	out := map[string]any{}
	for k, v := range reporadar.RadarAllVaults(cfg, api) {
		out[k] = v
	}
	return out
}

// asOrderedMap narrows a config value to the OrderedMap shape (nil
// when not one).
func asOrderedMap(cfg any) *storage.OrderedMap {
	if om, ok := cfg.(*storage.OrderedMap); ok {
		return om
	}
	return nil
}

// cmdInit mirrors cmd_init: the wizard writes config.json (vault,
// Telegram api_id/hash/phone — collected for compatibility only, bot
// username, GitHub token, the other two vault paths, the LLM provider
// branch, proxy, the 404 threshold), runs the repo radar BEFORE saving
// so the config lands once already wired, saves, and prints the
// summary card.
func (c *CLI) cmdInit(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	existing, ok := c.loadConfig(path)
	if !ok {
		existing = storage.NewOrderedMap()
	}
	fmt.Fprint(c.Out, Paint("First-run setup — answers are saved to ", cBold)+Paint(path, cCyan)+"\n")
	fmt.Fprint(c.Out, Paint("Press Enter to keep the [current] value. Secrets stay on this machine.\n", cDim)+"\n")

	cfg := shallowCopyConfig(existing)

	// -- vault (the v0.83.0 empty-start law: a YOUR_ placeholder counts
	// as unset, so the wizard starts EMPTY) --------------------------
	curVault := orStr(getOr(cfg, "vault_path", ""))
	if curVault != "" {
		s := strings.TrimSpace(curVault)
		if s == "" || strings.HasPrefix(strings.ToUpper(s), "YOUR_") {
			curVault = ""
		}
	}
	cfg.Set("vault_path", c.Ask("Obsidian vault folder", curVault))
	for cfgOrStr(cfg, "vault_path") != "" && !c.IsDir(cfgOrStr(cfg, "vault_path")) {
		c.CliPrint("That folder does not exist — try again (or leave empty to skip).", "warning")
		cfg.Set("vault_path", c.Ask("Obsidian vault folder", ""))
	}

	// -- telegram (COLLECTED for compatibility; the engine never uses
	// them per ADR-0004 Am.3) ----------------------------------------
	fmt.Fprint(c.Out, Paint("\n— Telegram (api.telegram.org credentials) —", cBold)+"\n")
	apiID := c.Ask("API ID", orStr(getOr(cfg, "telegram_api_id", "")))
	cfg.Set("telegram_api_id", pyIntOrZero(apiID))
	cfg.Set("telegram_api_hash", c.Ask("API hash", cfgOrStr(cfg, "telegram_api_hash")))
	cfg.Set("telegram_phone", c.Ask("Phone (+98…)", cfgOrStr(cfg, "telegram_phone")))

	// -- bot ----------------------------------------------------------
	fmt.Fprint(c.Out, Paint("\n— Bot queue (your dedicated bot chat) —", cBold)+"\n")
	cfg.Set("bot_username", lstripAt(c.Ask("Bot username (without @)", cfgOrStr(cfg, "bot_username"))))

	// -- github -------------------------------------------------------
	fmt.Fprint(c.Out, Paint("\n— GitHub (higher API rate limit with a token) —", cBold)+"\n")
	fmt.Fprint(c.Out, Paint("    create one at github.com/settings/tokens (public_repo read is enough)", cDim)+"\n")
	cfg.Set("github_token", c.Ask("GitHub token", cfgOrStr(cfg, "github_token")))

	// -- the other two vaults (Enter = skip; the Websites folder may
	// not exist yet — its pipeline creates it) -----------------------
	fmt.Fprint(c.Out, Paint("\n— The other two vaults (Enter to skip) —", cBold)+"\n")
	cfg.Set("website_vault_path", c.Ask("Websites vault folder", orStr(getOr(cfg, "website_vault_path", ""))))
	cfg.Set("manual_vault_path", c.Ask("Personal (Manual Notes) vault folder", orStr(getOr(cfg, "manual_vault_path", ""))))

	// THE REPO RADAR — detect the backup repo for every vault whose
	// slot is still EMPTY (read-only; typed names always win; nothing
	// is ever created). Runs BEFORE the save so the wizard's config
	// lands once, already wired.
	if strings.TrimSpace(cfgOrStr(cfg, "github_token")) != "" {
		c.runRepoRadar(cfg)
	}

	// -- llm provider -------------------------------------------------
	fmt.Fprint(c.Out, Paint("\n— LLM provider —", cBold)+"\n")
	savedRaw := getOr(cfg, "llm_provider", "ollama")
	defProv := "cloud"
	if s, ok := savedRaw.(string); ok {
		switch s {
		case "llamacpp":
			defProv = "llamacpp"
		case "ollama":
			defProv = "ollama"
		}
	}
	prov := c.Ask("Provider: 1) ollama  2) OpenAI-compatible endpoint "+
		"(llama.cpp, vLLM, LM Studio, cloud)  3) llama.cpp "+
		"(local, auto-detected)", defProv)
	low := strings.ToLower(prov)
	switch {
	case strings.Contains(prov, "3") || strings.Contains(low, "llama"):
		prov = "llamacpp"
	case strings.Contains(prov, "2") || strings.Contains(low, "cloud") ||
		strings.Contains(low, "openai"):
		prov = "cloud"
	default:
		prov = "ollama"
	}
	cfg.Set("llm_provider", prov)
	switch prov {
	case "ollama":
		oll := getMap(cfg, "ollama")
		nb := storage.NewOrderedMap()
		nb.Set("base_url", c.Ask("Ollama base URL",
			orStr(getOr(oll, "base_url", "http://127.0.0.1:11434"))))
		nb.Set("model", c.Ask("Ollama model",
			orStr(getOr(oll, "model", "qwen2.5-coder:7b"))))
		cfg.Set("ollama", nb)
	case "llamacpp":
		cfg.Set("llamacpp_api_url", c.Ask(
			"llama.cpp server URL (llama-server default http://127.0.0.1:8080)",
			orStr(getOr(cfg, "llamacpp_api_url", "http://127.0.0.1:8080/v1"))))
		cfg.Set("llamacpp_api_key", c.Ask(
			"API key (empty unless llama-server was started with --api-key)",
			orStr(getOr(cfg, "llamacpp_api_key", ""))))
		// "its model detected automatically": when the server runs, its
		// loaded model is offered as the default right here (the probe
		// rides the CLI seam — the ported llmclient underneath).
		probe := c.probeLlamaDoor(cfgOrStr(cfg, "llamacpp_api_url"),
			cfgOrStr(cfg, "llamacpp_api_key"), llmclient.LLamaCppProbeTimeoutS)
		defaultModel := ""
		if probe != nil && probe.Found {
			c.CliPrint(fmt.Sprintf("llama.cpp detected: %s", probe.Detail), "success")
			if probe.Model != nil {
				defaultModel = *probe.Model
			}
		} else if probe != nil {
			c.CliPrint(fmt.Sprintf("llama.cpp not detected yet (%s) "+
				"— the model is auto-detected when the server runs.", probe.Detail), "warning")
		}
		label := " (empty = auto-detect)"
		if defaultModel != "" {
			label = fmt.Sprintf(" (detected: %s)", defaultModel)
		}
		curModel := orStr(getOr(cfg, "llamacpp_model", ""))
		if curModel == "" {
			curModel = defaultModel
		}
		cfg.Set("llamacpp_model", c.Ask("Model"+label, curModel))
	default:
		cfg.Set("cloud_api_url", c.Ask(
			"Endpoint base URL (e.g. http://localhost:8080/v1 for llama.cpp)",
			orStr(getOr(cfg, "cloud_api_url", "https://api.openai.com/v1"))))
		cfg.Set("cloud_api_key", c.Ask("API key (empty for local servers)",
			orStr(getOr(cfg, "cloud_api_key", ""))))
		cfg.Set("cloud_model", c.Ask("Model name",
			orStr(getOr(cfg, "cloud_model", "gpt-4o-mini"))))
	}

	// -- proxy ----------------------------------------------------------
	fmt.Fprint(c.Out, Paint("\n— Proxy (leave disabled if Telegram is reachable directly) —", cBold)+"\n")
	px := getMap(cfg, "proxy")
	enAns := c.Ask("Enable proxy? (y/N)", pyBoolWord(pyTruthy(getOr(px, "enabled", nil))))
	en := strings.HasPrefix(strings.ToLower(enAns), "y")
	np := storage.NewOrderedMap()
	np.Set("enabled", en)
	if en {
		np.Set("type", c.Ask("Proxy type (socks5/http)",
			orStr(getOr(px, "type", "socks5"))))
		np.Set("host", c.Ask("Proxy host",
			orStr(getOr(px, "host", "127.0.0.1"))))
		np.Set("port", pyIntOrDoor(c.Ask("Proxy port",
			orStr(getOr(px, "port", 10808)))))
	} else {
		np.Set("type", getOr(px, "type", "socks5"))
		np.Set("host", getOr(px, "host", "127.0.0.1"))
		np.Set("port", getOr(px, "port", 10808))
	}
	cfg.Set("proxy", np)

	// -- 404 quarantine -------------------------------------------------
	fmt.Fprint(c.Out, Paint("\n— Deleted-repo policy —", cBold)+"\n")
	thr := c.Ask("Quarantine a repo as dead after N consecutive 404s (2–10)",
		orStr(getOr(cfg, "notfound_strike_threshold", 3)))
	if n, err := strconv.Atoi(thr); err != nil {
		cfg.Set("notfound_strike_threshold", 3)
	} else {
		if n < 2 {
			n = 2
		}
		if n > 10 {
			n = 10
		}
		cfg.Set("notfound_strike_threshold", n)
	}

	if !c.saveConfig(path, cfg) {
		return 1
	}

	// Summary card (the reference's middle line: the bold filename +
	// the green " written — …" text — the .ljust(64) rides the PAINTED
	// string, escapes count as characters; the quirk is preserved by
	// pyLjust).
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint("┌─ saved ─────────────────────────────────────────┐", cGreen)+"\n")
	fmt.Fprint(c.Out, Paint("│ ", cGreen)+Paint("config.json", cBold)+
		pyLjust(Paint(" written — credentials stay local", cGreen), 64)+
		Paint("│", cGreen)+"\n")
	fmt.Fprint(c.Out, Paint("└──────────────────────────────────────────────────┘", cGreen)+"\n")

	// The interactive Telegram login offer is DROPPED per ADR-0004
	// Am.3 — no MTProto client ships, so no login is ever offered. The
	// closing card speaks the honest Go-port truth (the golden's
	// reference card keeps the login lines as the pre-Option-B record).
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, "\n")
	c.CliPrint("Setup complete — no Telegram login is needed in the Go port "+
		"(the bot queue rides the Worker HTTP API).", "success")
	c.CliPrint("Double-click "+Paint(`gitcurator\cli\GitCurator-CLI.bat`, cBold)+
		" (Windows) or run:", "info")
	fmt.Fprint(c.Out, Paint("      gitcurator-cli --auto", cCyan)+"\n")
	return 0
}

// runRepoRadar wires the radar into the wizard: the pending gate, the
// 📡 header, the seam-driven detection, apply + report, the applied
// note. The reference wraps the whole block in try/except — a panic is
// recovered and spoken as the skip warning.
func (c *CLI) runRepoRadar(cfg *storage.OrderedMap) {
	defer func() {
		if r := recover(); r != nil {
			c.CliPrint(fmt.Sprintf("Repo radar skipped: %v", r), "warning")
		}
	}()
	if !reporadar.RadarPending(cfg) {
		return
	}
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint("📡 Repo radar — looking for the vaults' backup repos "+
		"on your GitHub account (read-only)…", cBold)+"\n")
	results := c.radarAllDoor(cfg)
	applied := reporadar.ApplyDetections(cfg, results)
	for _, line := range reporadar.ReportLines(results) {
		color := cDim
		if strings.HasPrefix(line, "🔗") {
			color = cGreen
		}
		fmt.Fprint(c.Out, Paint("  "+line, color)+"\n")
	}
	if applied > 0 {
		fmt.Fprint(c.Out, Paint(fmt.Sprintf(
			"  → %d backup repo name(s) set automatically — "+
				"nothing to type, nothing created", applied), cGreen)+"\n")
	} else {
		fmt.Fprint(c.Out, Paint(
			"  → nothing conclusive — empty slots stay empty "+
				"(create the repo on GitHub once and the next radar pass finds it)",
			cDim)+"\n")
	}
}

// shallowCopyConfig mirrors dict(existing): a new map sharing the
// nested values (apply_detections' nested writes land in the shared
// dicts — the same object identity law the reference rides).
func shallowCopyConfig(src *storage.OrderedMap) *storage.OrderedMap {
	out := storage.NewOrderedMap()
	if src == nil {
		return out
	}
	for _, k := range src.Keys() {
		v, _ := src.Get(k)
		out.Set(k, v)
	}
	return out
}

// pyIntOrZero mirrors int(s) if s.isdigit() else 0 (ASCII digits; the
// Unicode-digit and crash edges are the documented Go note).
func pyIntOrZero(s string) int64 {
	if !pyIsDigits(s) {
		return 0
	}
	n, _ := strconv.ParseInt(s, 10, 64)
	return n
}

// pyIsDigits mirrors str.isdigit() over the ASCII class.
func pyIsDigits(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// pyIntOrDoor mirrors int(x) or 10808: the parsed value, or the
// or-door's default when the answer parses to 0 / fails (the
// reference's crash edge is honestly hardened to the default).
func pyIntOrDoor(s string) any {
	n, err := strconv.Atoi(s)
	if err != nil || n == 0 {
		return 10808
	}
	return n
}

// pyBoolWord renders a bool as Python's str(): "True"/"False" — the
// wizard's y/N default ("y" if px.get("enabled") else "n") only needs
// the y/n word, built here directly.
func pyBoolWord(b bool) string {
	if b {
		return "y"
	}
	return "n"
}
