package cli

// preflight.go — cli_run.py's pre-flight family (565-719), T-62's doors:
// _preflight_github (the token check), _preflight_proxy (the TCP
// reach), _preflight_llm (the provider probes + THE interactive model
// picker — the PromptModelMenu/ApplyModelChoice bridge, landed
// SESSION-028), and preflight_checks (the composition with the run
// configuration card). Every batch driver calls this BEFORE any work:
// a missing model is caught and picked here, not 1 421 messages later.
//
// The seams (the reference's own mock surface, the instance-seam law):
//   - PreflightGithub: the PyGithub get_user().login leg. The default
//     door rides the connectioncheck GitHub seam (the same GET /user
//     the battery's GitHub check makes — the twin call); the golden
//     replay injects the crafted answers (the fake PyGithub module the
//     generator raises).
//   - NetDial: socket.create_connection((host, port), timeout=5). The
//     class-name token rides the seam exactly like the worker check's
//     OSError stand-in (the SHAPE is the contract — the real path
//     renders Go's own honest shapes through the same token slot).
//   - OllamaModels: the existing seam; an error's string IS the
//     class-name token for the unreachable line (the golden crafts
//     "ConnectionError"-shaped errors; the real path maps Go's net
//     errors onto the Python class names).

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// PreflightGHReply is _preflight_github's PyGithub outcome as the
// seam's shape: Status 200 (Login carries get_user().login), 401 (the
// rejection), or 0 (transport failure — ErrClass the exception's class
// name).
type PreflightGHReply struct {
	Status   int
	Login    string
	ErrClass string
}

// defaultPreflightGithub rides the connectioncheck GitHub seam (the
// same GET /user the battery's check makes) and maps the reply onto
// the preflight's three outcomes.
func defaultPreflightGithub(token string) PreflightGHReply {
	r := connectioncheck.GitHubGetFn("/user", token, "", 15)
	switch {
	case r.Status == 200:
		login := "?"
		if om, ok := r.Payload.(*storage.OrderedMap); ok {
			if v, present := om.Get("login"); present {
				login = helpers.PyStr(v)
				if login == "" {
					login = "?"
				}
			}
		}
		return PreflightGHReply{Status: 200, Login: login}
	case r.Status == 401:
		return PreflightGHReply{Status: 401}
	default:
		return PreflightGHReply{Status: 0, ErrClass: errClassOfStr(r.Err)}
	}
}

// errClassOfStr derives the exception-class token from an error string
// ("ClassName: message" → "ClassName"; a bare string stays whole).
func errClassOfStr(s string) string {
	if s == "" {
		return "OSError"
	}
	if i := strings.IndexByte(s, ':'); i > 0 {
		return s[:i]
	}
	return s
}

// errClassOf maps a Go error onto the Python exception-class token the
// reference's lines render (socket/urllib family).
func errClassOf(err error) string {
	if err == nil {
		return "OSError"
	}
	if ne, ok := err.(net.Error); ok && ne.Timeout() {
		return "TimeoutError"
	}
	if strings.Contains(err.Error(), "connection refused") {
		return "ConnectionRefusedError"
	}
	return errClassOfStr(err.Error())
}

// preflightGithub mirrors _preflight_github: (level, detail) with the
// level in success/warning/error.
func (c *CLI) preflightGithub(cfg *storage.OrderedMap) (string, string) {
	token := cfgOrStr(cfg, "github_token")
	if token == "" {
		return "warning", "no token — running unauthenticated (60 req/h limit)"
	}
	var reply PreflightGHReply
	if c.PreflightGithub != nil {
		reply = c.PreflightGithub(token)
	} else {
		reply = defaultPreflightGithub(token)
	}
	switch reply.Status {
	case 200:
		return "success", "authenticated as " + reply.Login
	case 401:
		return "error", "token REJECTED (401) — make a new one at github.com/settings/tokens"
	default:
		return "error", "GitHub unreachable (" + reply.ErrClass +
			") — check internet"
	}
}

// preflightProxy mirrors _preflight_proxy: TCP-reach the configured
// proxy. The int() edge is guarded in the reference itself (a
// non-integer port falls to 0).
func (c *CLI) preflightProxy(cfg *storage.OrderedMap) (string, string) {
	px := getMap(cfg, "proxy")
	if !pyTruthy(getOr(px, "enabled", nil)) {
		return "dim", "disabled — connecting directly"
	}
	host := mapOrStr(px, "host")
	if host == "" {
		host = "127.0.0.1"
	}
	port := pyIntOrStr(getOr(px, "port", nil))
	var ms int
	var errClass string
	if c.NetDial != nil {
		ms, errClass = c.NetDial(host, port, 5)
	} else {
		ms, errClass = defaultNetDial(host, port, 5)
	}
	if errClass == "" {
		typ := mapOrStr(px, "type")
		if typ == "" {
			typ = "socks5"
		}
		return "success", fmt.Sprintf("%s:%d reachable (%s, %d ms)",
			host, port, typ, ms)
	}
	return "warning", fmt.Sprintf(
		"%s:%d UNREACHABLE (%s) — start the proxy app or disable it in config.json",
		host, port, errClass)
}

// pyIntOrStr mirrors int(px.get("port", 0) or 0) with its except door
// (TypeError/ValueError → 0).
func pyIntOrStr(v any) int {
	switch t := v.(type) {
	case int64:
		return int(t)
	case float64:
		return int(t)
	case int:
		return t
	case string:
		if t == "" {
			return 0
		}
		if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil {
			return n
		}
		return 0
	default:
		return 0
	}
}

// defaultNetDial is socket.create_connection((host, port), timeout)
// with the elapsed-ms measure.
func defaultNetDial(host string, port, timeoutS int) (int, string) {
	t0 := time.Now()
	conn, err := net.DialTimeout("tcp",
		net.JoinHostPort(host, strconv.Itoa(port)),
		time.Duration(timeoutS)*time.Second)
	if err != nil {
		return 0, errClassOf(err)
	}
	conn.Close()
	return int(time.Since(t0).Milliseconds()), ""
}

// preflightLLM mirrors _preflight_llm: the provider probes + THE
// interactive model picker. Returns false only when the run cannot
// proceed (no server / no model at all / user declined).
func (c *CLI) preflightLLM(cfg *storage.OrderedMap, configPath string) bool {
	provider := cfgProvider(cfg)

	// -- llama.cpp -------------------------------------------------------
	if provider == "llamacpp" {
		url := cfgOrStr(cfg, "llamacpp_api_url")
		if url == "" {
			url = llmclient.LLamaCppDefaultBase
		}
		key := cfgOrStr(cfg, "llamacpp_api_key")
		probe := c.probeLlamaDoor(url, key, 15)
		if !probe.Found {
			// The reference's `detect_llamacpp(key) or probe` keeps the
			// ORIGINAL probe when the scan answers nothing; the Go door may
			// return nil — the same keep-the-original law.
			if scan := c.detectLlamaDoor(key); scan != nil {
				probe = scan
			}
		}
		if !probe.Found {
			c.checkLine("error", "LLM provider",
				fmt.Sprintf("llama.cpp server not detected (%s) — start it with: llama-server -m <model>.gguf --port 8080",
					probe.Detail))
			return false
		}
		if probe.BaseURL != "" && url != probe.BaseURL+"/v1" {
			// The configured URL was stale — the scan found the server
			// elsewhere; persist the working URL.
			cfg.Set("llamacpp_api_url", probe.BaseURL+"/v1")
			url = probe.BaseURL + "/v1"
		}
		model := llmclient.ResolveLlamaCppModel(cfg, probe.Models, probe.PropsModel)
		if model == nil || *model == "" {
			c.checkLine("error", "LLM provider",
				"llama.cpp up but no model loaded — start llama-server with -m <model>.gguf")
			return false
		}
		configured := pyStrip(cfgOrStr(cfg, "llamacpp_model"))
		if configured == "" {
			c.ApplyModelChoice(cfg, *model, configPath)
			detail := fmt.Sprintf(
				"llama.cpp up @ %s · model '%s' (auto-detected — saved to config.json)",
				probe.BaseURL, *model)
			if probe.Ready != nil && !*probe.Ready {
				detail += " · still loading"
			}
			c.checkLine("success", "LLM provider", detail)
			return true
		}
		listed := len(probe.Models) == 0
		if !listed {
			for _, n := range probe.Models {
				if strings.EqualFold(n, cfgOrStr(cfg, "llamacpp_model")) {
					listed = true
					break
				}
			}
		}
		level := "success"
		detail := fmt.Sprintf("llama.cpp up @ %s · model '%s'",
			probe.BaseURL, cfgOrStr(cfg, "llamacpp_model"))
		if !listed {
			level = "warning"
			detail += " (NOT in the /v1/models list — trying anyway)"
		}
		if probe.Ready != nil && !*probe.Ready {
			detail += " · still loading"
		}
		c.checkLine(level, "LLM provider", detail)
		return true
	}

	// -- cloud (any other provider) --------------------------------------
	if provider != "ollama" {
		c.checkLine("success", "LLM provider", fmt.Sprintf(
			"OpenAI-compatible endpoint %s · model '%s'",
			rawGet(cfg, "cloud_api_url", "?"), rawGet(cfg, "cloud_model", "?")))
		return true
	}

	// -- Ollama -----------------------------------------------------------
	oll := getMap(cfg, "ollama")
	base := rawGet(oll, "base_url", "http://127.0.0.1:11434")
	model := rawGet(oll, "model", "")
	models, err := c.ollamaModelsDoor(base, 15)
	if err != nil {
		c.checkLine("warning", "LLM provider", fmt.Sprintf(
			"Ollama not reachable at %s (%s) — the pipeline will try to auto-start it",
			base, errClassOf(err)))
		return true
	}
	if len(models) == 0 {
		c.checkLine("error", "LLM provider", fmt.Sprintf(
			"Ollama up @ %s but NO models installed — run `ollama pull <model>` first", base))
		return false
	}
	if model != "" && containsStr(models, model) {
		c.checkLine("success", "LLM provider", fmt.Sprintf(
			"Ollama up @ %s · %d model(s) · '%s' ready", base, len(models), model))
		return true
	}
	c.checkLine("warning", "LLM provider", fmt.Sprintf(
		"Ollama up @ %s · %d model(s) — configured '%s' is NOT pulled",
		base, len(models), model))
	choice := c.PromptModelMenu(model, models, nil, true)
	if choice == "" {
		c.CliPrint("No model selected — aborting. Run `ollama pull <name>` "+
			"(or fix config.json) and rerun.", "error")
		return false
	}
	c.ApplyModelChoice(cfg, choice, configPath)
	c.checkLine("success", "LLM provider",
		fmt.Sprintf("switched to '%s' — saved to config.json", choice))
	return true
}

// ollamaModelsDoor is the nil-safe OllamaModels seam accessor.
func (c *CLI) ollamaModelsDoor(base string, timeoutS float64) ([]string, error) {
	if c.OllamaModels != nil {
		return c.OllamaModels(base, timeoutS)
	}
	return defaultOllamaModels(base, timeoutS)
}

// PreflightChecks mirrors preflight_checks: the (optional) run
// configuration card, the section divider, the GitHub token + proxy +
// LLM checks. Returns false when the run must abort. Never touches the
// network twice when a check is irrelevant (proxy is checked only when
// enabled, cloud provider skips the Ollama probe). The GitHub-error
// door asks "continue anyway?" — only a TTY ever asks (the cron law).
func (c *CLI) PreflightChecks(cfg *storage.OrderedMap, card bool) bool {
	path := ""
	if v, ok := cfg.Get("__config_path__"); ok {
		// The reference passes cfg.get("__config_path__") through
		// verbatim (None when absent); only a real string ever names a
		// file — a null must NOT become a literal "None" path.
		if s, isStr := v.(string); isStr {
			path = s
		}
	}
	if card {
		c.RunConfigCard(cfg)
	}
	c.section("pre-flight checks")

	lvl, txt := c.preflightGithub(cfg)
	c.checkLine(lvl, "GitHub token", txt)
	if lvl == "error" {
		cont := false
		if c.IsATTYIn() {
			ans := c.Ask("GitHub check failed — continue anyway? (y/N)", "n")
			cont = strings.HasPrefix(strings.ToLower(ans), "y")
		}
		if !cont {
			c.CliPrint("Aborting — fix the token (config.json / --init) and rerun.", "error")
			return false
		}
	}

	lvl, txt = c.preflightProxy(cfg)
	c.checkLine(lvl, "Proxy", txt)

	if !c.preflightLLM(cfg, path) {
		return false
	}
	fmt.Fprint(c.Out, "\n")
	return true
}
