package cli

// detectllm.go — the --detect-llm command (cli.py cmd_detect_llm,
// 814-976): the CLI twin of the GUI's Detect & Set quick-switch
// buttons — probe the local engine, pick the model (the numbered menu
// when several are installed), set provider + model + URL and SAVE.
//
// The probe job is worker_jobs._quick_detect_job's composition
// (ollama: ONE /api/tags detect; llamacpp: the configured URL first,
// then the running llama-server process + the common ports), run
// SYNCHRONOUSLY over the CLI's probe seams — the reference rides a
// GUI thread + StatusLine spinner; the port needs no event loop, and
// the spinner paint box is already ported (terminal.go). --yes
// auto-picks the current (or first) model without the menu — the
// non-interactive path. Exit 0 = provider set + saved; 1 = engine not
// found / no model / choice aborted.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// DetectJobResult mirrors _quick_detect_job's result dict: success
// only means the probe RAN; found says whether the engine answered.
type DetectJobResult struct {
	Success    bool
	Provider   string
	Found      bool
	BaseURL    string
	Models     []string
	Detail     string
	PropsModel string
	Ready      *bool
	Via        string
}

// runDetectJob mirrors _quick_detect_job (the synchronous port): the
// job's log lines ride cli_print exactly as the reference's _LogShim
// pumped them through the StatusLine.
func (c *CLI) runDetectJob(provider string, cfg *storage.OrderedMap) DetectJobResult {
	provider = strings.ToLower(provider)
	out := DetectJobResult{Provider: provider, Detail: "not run"}
	if provider == "ollama" {
		oll := getMap(cfg, "ollama")
		base := strings.TrimSpace(orStr(getOr(oll, "base_url",
			"http://127.0.0.1:11434")))
		if base == "" {
			base = "http://127.0.0.1:11434"
		}
		c.CliPrint(fmt.Sprintf("🧠 Detecting Ollama at %s…", base), "info")
		r := c.detectOllamaDoor(base, 4.0)
		out.Success = true
		out.Found = r.Found
		out.BaseURL = r.BaseURL
		out.Models = r.Models
		out.Detail = r.Detail
		return out
	}
	key := cfgOrStr(cfg, "llamacpp_api_key")
	u := strings.TrimSpace(cfgOrStr(cfg, "llamacpp_api_url"))
	if u == "" {
		u = llmclient.LLamaCppDefaultBase
	}
	c.CliPrint(fmt.Sprintf("🦙 Detecting llama.cpp — %s first, then the running "+
		"llama-server process + common ports…", u), "info")
	probe := c.probeLlamaDoor(u, key, llmclient.LLamaCppProbeTimeoutS)
	via := "configured"
	if probe == nil || !probe.Found {
		if det := c.detectLlamaDoor(key); det != nil {
			probe = det
		}
		via = ""
		if probe != nil {
			via = probe.Via
		}
	}
	if probe != nil {
		out.Success = true
		out.Found = probe.Found
		out.BaseURL = probe.BaseURL
		if out.BaseURL == "" {
			out.BaseURL = u
		}
		out.Models = append([]string{}, probe.Models...)
		out.Detail = probe.Detail
		if probe.PropsModel != nil {
			out.PropsModel = *probe.PropsModel
		}
		out.Ready = probe.Ready
		out.Via = via
		if out.PropsModel != "" && !containsStr(out.Models, out.PropsModel) {
			out.Models = append(out.Models, out.PropsModel)
		}
	}
	return out
}

// cmdDetectLLM mirrors cmd_detect_llm.
func (c *CLI) cmdDetectLLM(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.CliPrint(fmt.Sprintf("No config at %s — run `--cli --init` first.", path), "warning")
		return 1
	}
	provider := strings.ToLower(a.DetectLLM)
	label := "Ollama"
	icon := "🧠"
	if provider != "ollama" {
		label = "llama.cpp"
		icon = "🦙"
	}

	fmt.Fprint(c.Out, Paint(fmt.Sprintf("%s Detect & Set %s — probing the local engine…",
		icon, label), cBold)+"\n")
	fmt.Fprint(c.Out, "\n")
	result := c.runDetectJob(provider, cfg)

	if !result.Success {
		c.CliPrint(fmt.Sprintf("Detect & Set %s: the probe crashed — %s",
			label, result.Detail), "error")
		return 1
	}
	if !result.Found {
		if provider == "ollama" {
			c.CliPrint(fmt.Sprintf("Ollama is not running at %s (%s).",
				result.BaseURL, result.Detail), "error")
			c.CliPrint("Start the Ollama app (or `ollama serve`), then run this again.", "info")
		} else {
			c.CliPrint(fmt.Sprintf("No llama.cpp server found (%s).", result.Detail), "error")
			c.CliPrint("Start it with:  llama-server -m <model>.gguf --port 8080", "info")
			c.CliPrint("then run this again.", "info")
		}
		return 1
	}

	base := result.BaseURL
	var models []string
	for _, m := range result.Models {
		if m != "" {
			models = append(models, m)
		}
	}
	where := fmt.Sprintf("at %s", base)
	if provider == "llamacpp" && result.Via == "process" {
		where += " (found via the running llama-server process)"
	}
	fmt.Fprint(c.Out, Paint(fmt.Sprintf("✅ %s detected %s", label, where), cGreen)+"\n")
	if len(models) == 0 {
		if provider == "ollama" {
			c.CliPrint("Ollama is up but NO models are installed — pull one "+
				"with:  ollama pull <model>", "error")
		} else {
			c.CliPrint("llama.cpp is up but no model could be read — start "+
				"llama-server with -m <model>.gguf", "error")
		}
		return 1
	}

	// Pick the model: the only one directly; a numbered menu when
	// several (the CLI twin of the GUI's model-selection dialog).
	current := ""
	if provider == "ollama" {
		current = strings.TrimSpace(mapOrStr(getMap(cfg, "ollama"), "model"))
	} else {
		current = strings.TrimSpace(cfgOrStr(cfg, "llamacpp_model"))
	}
	var choice string
	if len(models) == 1 {
		choice = models[0]
		fmt.Fprintf(c.Out, "   model: %s (the only one installed)\n", Paint(choice, cCyan))
	} else {
		fmt.Fprintf(c.Out, "   %d models detected — pick one:\n", len(models))
		for i, name := range models {
			mark := ""
			if name == current {
				mark = "  ← current"
			}
			line := fmt.Sprintf("   %s. %s", Paint(strconv.Itoa(i+1), cCyan), name)
			if mark != "" {
				line += Paint(mark, cDim)
			}
			fmt.Fprint(c.Out, line+"\n")
		}
		if a.Yes {
			choice = current
			if !containsStr(models, current) {
				choice = models[0]
			}
			fmt.Fprint(c.Out, Paint(fmt.Sprintf("   --yes → using '%s'", choice), cDim)+"\n")
		} else {
			def := current
			if !containsStr(models, def) {
				def = models[0]
			}
			fmt.Fprint(c.Out, Paint("? ", cCyan)+
				fmt.Sprintf("Model number [1-%d] ", len(models))+
				fmt.Sprintf("(Enter = %s): ", def))
			raw, err := c.readLine()
			if err != nil && raw == "" {
				raw = "" // EOFError — the default
			}
			raw = pyStrip(raw)
			if raw == "" {
				choice = def
			} else if n, perr := strconv.Atoi(raw); perr == nil &&
				pyIsDigits(raw) && n >= 1 && n <= len(models) {
				choice = models[n-1]
			} else {
				c.CliPrint(fmt.Sprintf("'%s' is not a valid choice — aborted, "+
					"nothing changed.", raw), "warning")
				return 1
			}
		}
	}

	// SET + SAVE (the same keys the GUI buttons write — one source of
	// truth, read by the next batch on either side).
	if provider == "ollama" {
		cfg.Set("llm_provider", "ollama")
		oll := getMap(cfg, "ollama")
		oll.Set("base_url", base)
		oll.Set("model", choice)
		cfg.Set("ollama", oll)
	} else {
		cfg.Set("llm_provider", "llamacpp")
		cfg.Set("llamacpp_api_url", strings.TrimRight(base, "/")+"/v1")
		cfg.Set("llamacpp_model", choice)
	}
	if !c.saveConfig(path, cleanConfigForSave(cfg)) {
		return 1
	}
	shown := base
	if provider != "ollama" {
		shown = cfgOrStr(cfg, "llamacpp_api_url")
	}
	fmt.Fprint(c.Out, "\n")
	fmt.Fprint(c.Out, Paint(fmt.Sprintf("✅ LLM provider SET to %s — %s · model '%s'",
		label, shown, choice), cGreen)+"\n")
	fmt.Fprint(c.Out, Paint(fmt.Sprintf(
		"   saved to %s — the next batch uses it immediately.", path), cDim)+"\n")
	return 0
}
