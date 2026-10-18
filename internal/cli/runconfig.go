package cli

// runconfig.go — cli_run.py run_config_card (524-562): the boxed
// "run configuration" snapshot shown before --auto runs — mirrors the
// GUI's settings summary. Ported as T-62's door (the preflight's card
// leg, cli_run.py 696-697).
//
// Exactness notes (the raw-comparison laws):
//   - the provider branches compare cfg.get("llm_provider", "ollama")
//     RAW against "ollama"/"llamacpp" — no strip, no lower (a "Ollama"
//     spelling falls to the OpenAI-compatible branch, byte-for-byte
//     like the reference);
//   - cloud_model / proxy host+port render through .get(key, default) —
//     a present-but-null value renders "None", an absent one the
//     default (str(v) in the f-string), never "";
//   - the token mask counts RUNES (Python len over str).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// rawGet mirrors dict.get(key, default) rendering: absent → the default,
// present → str(v) (None for null — the f-string's implicit str()).
func rawGet(cfg *storage.OrderedMap, key, def string) string {
	if cfg == nil {
		return def
	}
	if v, ok := cfg.Get(key); ok {
		return helpers.PyStr(v)
	}
	return def
}

// RunConfigCard mirrors run_config_card: five rows (Vault / Bot queue /
// Proxy / LLM / GitHub token) in a box-drawing frame, every value
// truncated with the … rune when wider than the inner box. The label
// column is 18 wide (labels ljust to 16); the box's inner width is
// min(max(64, term_width()-2), 118) - 2.
func (c *CLI) RunConfigCard(cfg *storage.OrderedMap) {
	px := getMap(cfg, "proxy")

	prov, _ := getOr(cfg, "llm_provider", "ollama").(string)
	var llmTxt string
	switch {
	case prov == "ollama":
		oll := getMap(cfg, "ollama")
		model := mapOrStr(oll, "model")
		if model == "" {
			model = "?"
		}
		llmTxt = "Ollama · " + model
	case prov == "llamacpp":
		m := cfgOrStr(cfg, "llamacpp_model")
		if m == "" {
			m = "(auto-detect)"
		}
		url := cfgOrStr(cfg, "llamacpp_api_url")
		if url == "" {
			url = "http://127.0.0.1:8080/v1"
		}
		llmTxt = "llama.cpp · " + m + " @ " + url
	default:
		llmTxt = "OpenAI-compatible · " + rawGet(cfg, "cloud_model", "?")
	}

	tok := cfgOrStr(cfg, "github_token")
	var tokTxt string
	switch {
	case len([]rune(tok)) > 10:
		r := []rune(tok)
		tokTxt = string(r[:4]) + strings.Repeat("•", 8) + string(r[len(r)-4:])
	case tok != "":
		tokTxt = strings.Repeat("•", len([]rune(tok)))
	default:
		tokTxt = "(not set)"
	}

	vault := cfgOrStr(cfg, "vault_path")
	if vault == "" {
		vault = "(not set)"
	}
	bot := "(not set)"
	if b := cfgOrStr(cfg, "bot_username"); b != "" {
		bot = "@" + b
	}
	proxy := "disabled"
	if pyTruthy(getOr(px, "enabled", nil)) {
		proxy = rawGet(px, "host", "?") + ":" + rawGet(px, "port", "?") + " (on)"
	}

	type row struct{ k, v string }
	rows := []row{
		{"Vault", vault},
		{"Bot queue", bot},
		{"Proxy", proxy},
		{"LLM", llmTxt},
		{"GitHub token", tokTxt},
	}

	labelW := 18
	w := termWidth() - 2
	if w < 64 {
		w = 64
	}
	if w > 118 {
		w = 118
	}
	avail := w - 2 // inner box width
	valW := avail - labelW - 1

	shown := make([]row, 0, len(rows))
	for _, r := range rows {
		v := r.v
		if len([]rune(v)) > valW {
			v = string([]rune(v)[:valW-1]) + "…"
		}
		shown = append(shown, row{r.k, v})
	}

	title := " run configuration "
	lead := (avail - len([]rune(title))) / 2
	if lead < 3 {
		lead = 3
	}
	fmt.Fprint(c.Out, Paint("┌"+strings.Repeat("─", lead)+title+
		strings.Repeat("─", avail-lead-len([]rune(title)))+"┐", cDim)+"\n")
	for _, r := range shown {
		vr := len([]rune(r.v))
		fmt.Fprint(c.Out, Paint("│ ", cDim)+
			Paint(pyLjust(r.k, labelW-2), cBold)+
			Paint(" "+r.v, cCyan)+
			Paint(strings.Repeat(" ", avail-labelW-1-vr)+"│", cDim)+"\n")
	}
	fmt.Fprint(c.Out, Paint("└"+strings.Repeat("─", avail)+"┘", cDim)+"\n")
}
