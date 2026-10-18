package cli

// status.go — the --status command (cli.py cmd_status, the full card):
// the config summary rows, the taxonomy row, the sentinel probes, the
// Ollama/llama.cpp model checks, and the CacheDB / NoteStateDB /
// WebsiteStateDB stats blocks. Never creates state beyond CacheDB's own
// open (the reference's CacheDB() opens — and on a fresh install
// creates — APP_DIR/cache.db; the note-state and websites blocks only
// read when the file exists).

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// Default seam constructors (the net_guard / ollama import doors).

func defaultProbeInternet(cfg any, timeoutS float64) connectioncheck.ProbeInternetResult {
	return connectioncheck.ProbeInternet(cfg, timeoutS)
}

func defaultAIStatus(cfg any, timeoutS float64) connectioncheck.AIStatusResult {
	return connectioncheck.AIStatus(cfg, timeoutS)
}

func defaultOllamaModels(baseURL string, timeoutS float64) ([]string, error) {
	return llmclient.ListModelsWithTimeout(llmclient.NewOllamaClient(baseURL), timeoutS)
}

func defaultProbeLlama(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
	return llmclient.ProbeLlamaCpp(baseURL, apiKey, timeoutS)
}

func defaultDetectLlama(apiKey string) *llmclient.ProbeResult {
	return llmclient.DetectLlamaCpp(apiKey, nil, "127.0.0.1",
		llmclient.LLamaCppProbeTimeoutS, true)
}

// vaultCell mirrors cmd_status's _vault_cell: "(not set)", "  ✓
// exists", "  ◌ will be created" (websites) or "  ✗ not found".
func (c *CLI) vaultCell(raw any, kind string) string {
	p := strings.TrimSpace(orStr(raw))
	if p == "" {
		return cDim + "(not set)" + cReset
	}
	if c.IsDir(p) {
		return Paint("  ✓ exists", cGreen)
	}
	if kind == "websites" {
		return Paint("  ◌ will be created", cYellow)
	}
	return Paint("  ✗ not found", cRed)
}

// cmdStatus mirrors cmd_status (the full card).
func (c *CLI) cmdStatus(a *Args) int {
	c.Banner()
	path := configPath(a.Config)
	cfg, ok := c.loadConfig(path)
	if !ok {
		c.CliPrint(fmt.Sprintf("No config at %s — run `--cli --init` first.", path), "warning")
		return 1
	}

	vault := helpers.PyStr(getOr(cfg, "vault_path", ""))
	fmt.Fprint(c.Out, Paint("Configuration ", cBold)+Paint(fmt.Sprintf("(%s)", path), cDim)+"\n")
	fmt.Fprint(c.Out, Rule()+"\n")

	pipes := getMap(cfg, "pipelines")
	ghOn := pyTruthy(getOr(pipes, "github", true))
	wsOn := pyTruthy(getOr(pipes, "websites", false))
	pipelinesCell := "github " + onOff(ghOn, cGreen, cRed) +
		" · websites " + onOff(wsOn, cGreen, cDim)

	manual := strings.TrimSpace(orStr(getOr(cfg, "manual_vault_path", nil)))
	libraryCell := cDim + "(needs the Manual vault)" + cReset
	if manual != "" {
		libraryCell = cDim + "tools/mirror_manual.py — dry-run first, then --apply" + cReset
	}

	provRaw := getOr(cfg, "llm_provider", "ollama")
	provStr := helpers.PyStr(provRaw)
	provIs := func(name string) bool { s, ok := provRaw.(string); return ok && s == name }
	var modelPart string
	switch {
	case provIs("ollama"):
		modelPart = fmt.Sprintf(" · %s", helpers.PyStr(getOr(getMap(cfg, "ollama"), "model", "?")))
	case provIs("llamacpp"):
		m := orStr(getOr(cfg, "llamacpp_model", ""))
		if m == "" {
			m = "(auto-detect)"
		}
		modelPart = fmt.Sprintf(" · %s", m)
	default:
		modelPart = fmt.Sprintf(" · %s", helpers.PyStr(getOr(cfg, "cloud_model", "?")))
	}
	models := getMap(cfg, "models")
	var llmExtra string
	if v := getOr(models, "classify", nil); pyTruthy(v) {
		llmExtra += fmt.Sprintf(" · classify='%s'", helpers.PyStr(v))
	}
	if v := getOr(models, "analyze", nil); pyTruthy(v) {
		llmExtra += fmt.Sprintf(" · analyze='%s'", helpers.PyStr(v))
	}
	llmCell := provStr + modelPart +
		cDim + fmt.Sprintf(" · num_ctx=%s", helpers.PyStr(getOr(cfg, "llm_num_ctx", 8192))) + llmExtra + cReset

	px := getMap(cfg, "proxy")
	var proxyCell string
	if pyTruthy(getOr(px, "enabled", nil)) {
		proxyCell = fmt.Sprintf("%s://%s:%s",
			helpers.PyStr(getOr(px, "type", nil)),
			helpers.PyStr(getOr(px, "host", nil)),
			helpers.PyStr(getOr(px, "port", nil)))
	} else {
		proxyCell = cDim + "disabled" + cReset
	}

	vaultVal := vault + Paint("  ✗ missing", cRed)
	if vault != "" && c.IsDir(vault) {
		vaultVal = vault + Paint("  ✓ exists", cGreen)
	}

	type row struct{ key, val string }
	rows := []row{
		{"Vault", vaultVal},
		{"Websites vault", c.vaultCell(getOr(cfg, "website_vault_path", ""), "websites")},
		{"Manual vault", c.vaultCell(getOr(cfg, "manual_vault_path", ""), "manual")},
		{"Library mirror", libraryCell},
		{"Vault repo", repoSlot(getMap(cfg, "vaultseal"), "repo_name")},
		{"Websites repo", repoSlot(cfg, "website_repo_name")},
		{"Personal repo", repoSlot(cfg, "manual_repo_name")},
		{"Pipelines", pipelinesCell},
		{"Telegram API ID", orStrCustom(getOr(cfg, "telegram_api_id", ""), "(not set)")},
		{"Telegram hash", Mask(helpers.PyStr(getOr(cfg, "telegram_api_hash", "")), 4)},
		{"Phone", Mask(helpers.PyStr(getOr(cfg, "telegram_phone", "")), 5)},
		{"Bot", botCell(getOr(cfg, "bot_username", ""))},
		{"GitHub token", Mask(helpers.PyStr(getOr(cfg, "github_token", "")), 4)},
		{"LLM", llmCell},
		{"Proxy", proxyCell},
		{"404 quarantine threshold", helpers.PyStr(getOr(cfg, "notfound_strike_threshold", 3)) +
			cDim + " (dead after N consecutive 404s)" + cReset},
	}
	for _, r := range rows {
		fmt.Fprintf(c.Out, "  %s %s\n", pyLjust(Paint(r.key, cBold), 22), r.val)
	}

	// The taxonomy row (resolve_taxonomy_path; never fails the card).
	tax := constants.ResolveTaxonomyPath(map[string]any{
		"taxonomy_path": getOr(cfg, "taxonomy_path", nil),
	})
	taxCell := Paint("  ✗ missing", cRed)
	if fileExists(tax) {
		taxCell = Paint("  ✓ exists", cGreen)
	}
	fmt.Fprintf(c.Out, "  %s %s%s\n", pyLjust(Paint("Taxonomy", cBold), 22),
		filepath.Base(tax), taxCell)

	// The sentinel rows (net_guard's probes; read-only, never failing).
	net := c.ProbeInternet(cfg, 5.0)
	var netCell string
	if net.Online {
		ep := net.Endpoint
		if i := strings.Index(ep, "//"); i >= 0 {
			ep = ep[i+2:]
		}
		ms := ""
		if net.LatencyMs != nil {
			ms = fmt.Sprintf(" · %d ms", *net.LatencyMs)
		}
		netCell = Paint("✓ online", cGreen) + fmt.Sprintf(" (%s%s)", ep, ms)
	} else {
		netCell = Paint("✗ offline", cRed) + cDim + " — syncing needs the internet" + cReset
	}
	fmt.Fprintf(c.Out, "  %s %s\n", pyLjust(Paint("Internet", cBold), 22), netCell)

	ai := c.AIStatus(cfg, 6.0)
	state := orStr(ai.State)
	var aiCell string
	switch state {
	case "on":
		eng := "cloud"
		if ai.Kind != "cloud" {
			e := orStr(ai.Engine)
			if e == "" {
				e = "local"
			}
			eng = fmt.Sprintf("local (%s)", e)
		}
		aiCell = Paint("✓ on", cGreen) + " — " + eng
	case "warn":
		s := orStr(ai.Short)
		if s == "" {
			s = "see detail"
		}
		aiCell = Paint("⚠ warn", cYellow) + " — " + s
	case "none":
		aiCell = cDim + "(not configured)" + cReset
	default:
		s := orStr(ai.Short)
		if s == "" {
			s = "unreachable"
		}
		aiCell = Paint("✗ off", cRed) + " — " + s
	}
	fmt.Fprintf(c.Out, "  %s %s\n", pyLjust(Paint("AI engine", cBold), 22), aiCell)
	if ai.Detail != "" {
		fmt.Fprint(c.Out, cDim+fmt.Sprintf("    %s", pySliceRunes(ai.Detail, 0, 96))+cReset+"\n")
	}

	// The Ollama / llama.cpp model checks.
	switch {
	case provIs("ollama"):
		oll := getMap(cfg, "ollama")
		base := helpers.PyStr(getOr(oll, "base_url", "http://127.0.0.1:11434"))
		modelsList, err := c.OllamaModels(base, 10)
		if err == nil {
			model := helpers.PyStr(getOr(oll, "model", ""))
			if model != "" && !containsStr(modelsList, model) {
				c.CliPrint(fmt.Sprintf("Configured model '%s' is NOT installed (%d model(s) available).", model, len(modelsList)), "warning")
				c.CliPrint("  → the next batch run opens the model picker (or change it via --init / config.json).", "info")
			} else if len(modelsList) > 0 {
				fmt.Fprintf(c.Out, "  %s %d model(s) installed · '%s' ready\n",
					pyLjust(Paint("Ollama", cBold), 22), len(modelsList), model)
			}
		}
	case provIs("llamacpp"):
		url := orStr(getOr(cfg, "llamacpp_api_url", ""))
		if url == "" {
			url = llmclient.LLamaCppDefaultBase
		}
		key := orStr(getOr(cfg, "llamacpp_api_key", ""))
		probe := c.ProbeLlama(url, key, llmclient.LLamaCppProbeTimeoutS)
		if probe != nil && !probe.Found {
			probe = c.DetectLlama(key)
		}
		if probe != nil && probe.Found {
			model := orStr(getOr(cfg, "llamacpp_model", ""))
			if model == "" && probe.Model != nil {
				model = *probe.Model
			}
			if model == "" {
				model = "(none loaded)"
			}
			fmt.Fprintf(c.Out, "  %s %s · configured '%s'\n",
				pyLjust(Paint("llama.cpp", cBold), 22), probe.Detail, model)
		} else if probe != nil {
			c.CliPrint(fmt.Sprintf("llama.cpp server not detected (%s) — start it with: llama-server -m <model>.gguf --port 8080", probe.Detail), "warning")
		}
	}
	fmt.Fprint(c.Out, "\n")

	// CacheDB stats (the engine is compiled in — the _gui_symbol door
	// maps to the ported store directly).
	func() {
		cache, err := store.NewCacheDB(filepath.Join(constants.AppDir(), "cache.db"))
		if err != nil {
			c.CliPrint(fmt.Sprintf("Cache stats unavailable (%s)", err), "warning")
			return
		}
		defer cache.Close()
		processed := len(cache.GetAllProcessedURLs())
		decommissioned := len(cache.GetAllDecommissioned())
		failed := cache.GetFailedCount()
		thr := thresholdFromConfig(cfg)
		quarantined := cache.GetDeadURLs(thr)
		allRows := cache.GetQuarantineStats()
		confirmed := map[string]bool{}
		for _, r := range quarantined {
			confirmed[r.URL] = true
		}
		fmt.Fprint(c.Out, Paint("Cache (cache.db)", cBold)+"\n")
		fmt.Fprint(c.Out, Rule()+"\n")
		fmt.Fprintf(c.Out, "  %s %d\n", pyLjust(Paint("Processed repos", cBold), 22), processed)
		fmt.Fprintf(c.Out, "  %s %d\n", pyLjust(Paint("Decommissioned (404)", cBold), 22), decommissioned)
		fmt.Fprintf(c.Out, "  %s %d\n", pyLjust(Paint("Retry queue", cBold), 22), failed)
		fmt.Fprintf(c.Out, "  %s %d confirmed / %d tracked\n",
			pyLjust(Paint("404 quarantine", cBold), 22), len(confirmed), len(allRows))
		limit := len(allRows)
		if limit > 10 {
			limit = 10
		}
		for _, r := range allRows[:limit] {
			flag := ""
			if confirmed[r.URL] {
				flag = Paint(" ⛔", cRed)
			}
			at := ""
			if r.DecommissionedAt != nil {
				at = *r.DecommissionedAt
			}
			fmt.Fprintf(c.Out, "      %s %s  %s%s\n",
				Paint("•", cYellow), r.URL,
				Paint(fmt.Sprintf("%d attempt(s) · %s", r.FailCount, pySliceRunes(at, 0, 16)), cDim),
				flag)
		}
		if len(allRows) > 10 {
			fmt.Fprint(c.Out, Paint(fmt.Sprintf("      … %d more", len(allRows)-10), cDim)+"\n")
		}
	}()
	fmt.Fprint(c.Out, "\n")

	// Note state (the moves-as-corrections record).
	dbFile := filepath.Join(constants.AppDir(), "cache.db")
	if fileExists(dbFile) {
		ns, err := notestate.New(dbFile)
		if err != nil {
			c.CliPrint(fmt.Sprintf("Note state unavailable (%s)", err), "warning")
		} else {
			ghCount, _ := ns.Count(notestate.VaultGithub)
			webCount, _ := ns.Count(notestate.VaultWebsites)
			ghCorr, _ := ns.CorrectionCount(notestate.VaultGithub)
			webCorr, _ := ns.CorrectionCount(notestate.VaultWebsites)
			ghDis, _ := ns.DismissedCount(notestate.VaultGithub)
			webDis, _ := ns.DismissedCount(notestate.VaultWebsites)
			ns.Close()
			fmt.Fprint(c.Out, Paint("Note state (moves-as-corrections record)", cBold)+"\n")
			fmt.Fprint(c.Out, Rule()+"\n")
			extra := ""
			if ghCount == 0 {
				extra = cDim + "  (the first real run records it)" + cReset
			}
			fmt.Fprintf(c.Out, "  %s %d note(s) recorded%s\n",
				pyLjust(Paint("GitHub vault baseline", cBold), 22), ghCount, extra)
			fmt.Fprintf(c.Out, "  %s %d note(s) recorded\n",
				pyLjust(Paint("Websites vault record", cBold), 22), webCount)
			fmt.Fprintf(c.Out, "  %s %d (github %d · websites %d)\n",
				pyLjust(Paint("Corrections applied", cBold), 22), ghCorr+webCorr, ghCorr, webCorr)
			fmt.Fprintf(c.Out, "  %s %d (github %d · websites %d)\n",
				pyLjust(Paint("Dismissed URLs", cBold), 22), ghDis+webDis, ghDis, webDis)
			fmt.Fprint(c.Out, "\n")
		}
	}

	// Websites pipeline state.
	if fileExists(dbFile) {
		ws, err := websitepipeline.NewWebsiteStateDB(dbFile)
		if err != nil {
			c.CliPrint(fmt.Sprintf("Websites state unavailable (%s)", err), "warning")
		} else {
			var proc, retryN, exhausted, dismissed int
			func() {
				defer ws.Close()
				row := ws.SQL().QueryRow("SELECT COUNT(*) FROM websites_processed")
				if err := row.Scan(&proc); err != nil {
					return
				}
				row = ws.SQL().QueryRow(
					"SELECT COUNT(*), "+
						"SUM(CASE WHEN attempts >= ? THEN 1 ELSE 0 END) "+
						"FROM website_retry_queue", websitepipeline.MaxFetchRetries)
				var n sql.NullInt64
				if err := row.Scan(&retryN, &n); err != nil {
					return
				}
				if n.Valid {
					exhausted = int(n.Int64)
				}
				row = ws.SQL().QueryRow("SELECT COUNT(*) FROM dismissed_urls")
				_ = row.Scan(&dismissed)
			}()
			fmt.Fprint(c.Out, Paint("Websites pipeline (cache.db)", cBold)+"\n")
			fmt.Fprint(c.Out, Rule()+"\n")
			fmt.Fprintf(c.Out, "  %s %d\n", pyLjust(Paint("Processed websites", cBold), 22), proc)
			pendingCell := fmt.Sprintf("%d pending", retryN)
			if exhausted != 0 {
				pendingCell += fmt.Sprintf(" (%d exhausted, kept in _review)", exhausted)
			}
			fmt.Fprintf(c.Out, "  %s %s\n", pyLjust(Paint("Fetch retry queue", cBold), 22), pendingCell)
			fmt.Fprintf(c.Out, "  %s %d\n", pyLjust(Paint("Dismissed URLs", cBold), 22), dismissed)
			fmt.Fprint(c.Out, "\n")
		}
	}
	return 0
}

// onOff renders the ON/OFF word pair with the given colors.
func onOff(on bool, onColor, offColor string) string {
	if on {
		return Paint("ON", onColor)
	}
	return Paint("OFF", offColor)
}

// repoSlot renders a backup-repo slot: the trimmed name or "(not set —
// the radar detects it)".
func repoSlot(m *storage.OrderedMap, key string) string {
	name := strings.TrimSpace(mapOrStr(m, key))
	if name != "" {
		return name
	}
	return cDim + "(not set — the radar detects it)" + cReset
}

// orStrCustom mirrors str(value or fallback).
func orStrCustom(v any, fallback string) string {
	if !pyTruthy(v) {
		return fallback
	}
	return helpers.PyStr(v)
}

// botCell mirrors the Bot row.
func botCell(v any) string {
	if pyTruthy(v) {
		return "@" + helpers.PyStr(v)
	}
	return cDim + "(not set)" + cReset
}

// containsStr mirrors `x in list`.
func containsStr(list []string, x string) bool {
	for _, s := range list {
		if s == x {
			return true
		}
	}
	return false
}

// thresholdFromConfig bridges the OrderedMap config into
// store.DeadLinkThresholdFromConfig (dead_link_threshold's law).
func thresholdFromConfig(cfg *storage.OrderedMap) int {
	m := map[string]any{}
	if cfg != nil {
		if v, ok := cfg.Get("notfound_strike_threshold"); ok {
			m["notfound_strike_threshold"] = v
		}
	}
	return store.DeadLinkThresholdFromConfig(m)
}

// fileExists answers os.path.exists for a regular file.
func fileExists(p string) bool {
	if fi, err := os.Stat(p); err == nil && !fi.IsDir() {
		return true
	}
	return false
}
