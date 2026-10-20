package app

// settings.go — T-73: the Settings pages A command surface (doc 04 §5's
// table extended with the settings doors; spec-gui-pages-a is the
// authority). The six pages — Credentials, Proxy, Vault, LLM, Input,
// Sound — drive THIS surface from the production shell (webui binds it
// directly, the Am.4-probed nested import).
//
// THE LIVE-CONFIG LAW (vaults_config.py's MainWindow semantics, the
// port's own composition): the app loads config.json ONCE at
// construction (config.LoadAppConfig — the GUI load path: fresh
// detection, the example base, THE FILL, the v0.07.1 credential strip
// heal) and holds the SAME *storage.OrderedMap for the session, exactly
// like the reference's self.config. Every door reads that instance;
// every save mutates it IN PLACE (clear+update, never rebind) so any
// holder — a running batch, a startup pass — sees saves live. The
// Python original rode the GIL; the Go twin serializes the settings'
// own doors (Settings.mu) and keeps the reference's shared-mutable
// shape for nested maps (documented — the engines snapshot values at
// composition boundaries, exactly where the reference reads them).
//
// THE SAVE LAW (vaults_config.py save_config, the v30 fix — all four
// laws pinned by tests):
//   1. MERGE, never whitelist-rebuild: unknown keys (cloudflare_*,
//      gdrive_*, hmac_secret, user-added) survive every save
//      (storage.MergeConfig); a page's group carries only its own keys
//      — an absent group keeps the config value (the reference's
//      defensive hasattr pattern, kept as the group-pointer law).
//   2. Tuning keys are only defaulted by the merge itself, never
//      stomped. log_level stays the reference's one hardcoded quirk:
//      every save writes "INFO" (a hand-edited DEBUG resets, exactly
//      like the reference's ui_values).
//   3. IN PLACE: merged values land in the LIVE map (clear+update),
//      never a rebind — the save-law test holds the pre-save map and
//      watches it change.
//   4. Atomic write (storage.WriteConfigFile) + the credentials mirror
//      (config.Save's composition) + the 250 ms sentinel re-arm
//      (QTimer.singleShot(250, _run_sentinel_pass)'s twin).
//
// THE PAGE WORKFLOWS (the button engines, every log line verbatim):
//   · Test GitHub Token (connection_tests.py:138-193) — GET /user with
//     the field AS TYPED, before any save.
//   · About Me Wizard (lifecycle.py:58-171) — the pure template build
//     + the CWD about_me.md write the runner reads.
//   · Ollama Refresh / Start Server (connection_tests.py:249-333).
//   · llama.cpp Detect / Refresh (llamacpp.py:297-375) + THE SERVER
//     CONFIG GRAB (_grab_llamacpp_limits — every probe carries the
//     launch limits home).
//   · Detect & Set (the quick-switch fast lane, llamacpp.py:445-687 +
//     worker_jobs._quick_detect_job): probe → the model menu when
//     several → provider + URL + model SET and merge-saved; cancel
//     keeps the provider untouched.
//   · The Sound preview gate (processing_control.py:2265-2307 +
//     sound.py's one-notice contract).
//   · The Vault status ladder (_vault_path_status) and the combo's
//     history law (populate_vaults / on_vault_changed / remove_vault).
//
// THE HERMETIC LAW: every network/process/filesystem seam is injected
// on SettingsDeps; the tests fake them at construction. A nil seam
// refuses honestly, never improvises.

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/assadigit/second-brain-project/assets"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/mirror"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/syncreview"
)

// GitHubAPIBase is the public GitHub REST root the token test probes
// (the reference's PyGithub default).
const GitHubAPIBase = "https://api.github.com"

// SettingsDeps carries the seams (the hermetic law — see the file doc).
// Every field is wired by the DI root (app.go); every one is fakeable.
type SettingsDeps struct {
	// Config yields the LIVE config map (the session-held instance).
	Config func() *storage.OrderedMap

	// Write persists a config map atomically + mirrors the
	// credentials (config.Save's composition). False = write failed.
	Write func(cfg *storage.OrderedMap) bool

	// DiscoverVaults is find_obsidian_vaults (batch.FindObsidianVaults)
	// — the vault combo's discovery half.
	DiscoverVaults func() []string

	// UserConfigStatus is user_config.status_line (the Credentials
	// page's quiet truth).
	UserConfigStatus func() string
	// UserConfigPath is user_config.global_config_path (the Show
	// button's enabled state + the open door's target).
	UserConfigPath func() string
	// OpenFileLocation opens the OS file manager AT the file (selected
	// when the platform supports it) or at its folder — the "Show the
	// credentials file" door; best-effort, never raises.
	OpenFileLocation func(path string)

	// GitHubGet is connectioncheck.GitHubGetFn — the token test's
	// GET /user seam.
	GitHubGet func(path, token, apiBase string, timeoutS float64) connectioncheck.GitHubReply

	// DetectOllama is llmclient.DetectOllama — GET /api/tags.
	DetectOllama func(base string, timeoutS float64) llmclient.DetectOllamaResult
	// ProbeLlama is llmclient.ProbeLlamaCpp at ONE address.
	ProbeLlama func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult
	// DetectLlama is llmclient.DetectLlamaCpp — the process ports +
	// the common-port scan (the CLI's detectLlamaDoor wiring).
	DetectLlama func(apiKey string) *llmclient.ProbeResult

	// The per-test doors' seams (T-79a) — every one faked at
	// construction per THE HERMETIC-SEAM law; the production wiring
	// lives in App.New.
	// OllamaChatPing is the Ollama inference ping (OllamaChat over a
	// fresh client; the options pass through; errText "" = no error).
	OllamaChatPing func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (reply, errText string)
	// CloudChatPing is the cloud chat ping (CloudChat's router: the
	// URL decides the wire format; errText "" = no error).
	CloudChatPing func(apiURL, apiKey, model, prompt string, timeoutS float64) (reply, errText string)
	// PreflightCloud is llmclient.PreflightCloud — the /models
	// pre-flight (warn-never-block; listed nil = no verdict).
	PreflightCloud func(apiURL, apiKey, model string, timeoutS float64) (ok bool, message string, listed *bool)
	// WorkerCheck is connectioncheck.CheckWorkerVersion (nil httpGet =
	// the production proxy-bypassed fetch; a nil Result = the queue is
	// not configured).
	WorkerCheck func(config any) *connectioncheck.Result
	// ProxyCheck is connectioncheck.ProxyReachable — the sentinel's own
	// TCP verdict (state disabled/connected/unreachable).
	ProxyCheck func(config any, timeoutS float64) connectioncheck.ProxyResult

	// The linking-tools doors' seams (T-79b) — every one faked at
	// construction per THE HERMETIC-SEAM law; the production wiring
	// lives in App.New. Message-list-shaped (the tools' llm(messages)
	// closures), JSON mode on.
	// LinkOllamaChat is the tools' Ollama chat lane (numCtx 0 = the
	// option rides unset, exactly like the reference's None; errText ""
	// = no error).
	LinkOllamaChat func(host, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (reply, errText string)
	// LinkCloudChat is the tools' OpenAI-compatible chat lane (cloud +
	// llama.cpp /v1 alike; errText "" = no error).
	LinkCloudChat func(apiURL, apiKey, model string, messages []llmclient.Message,
		numCtx int, timeoutS float64) (reply, errText string)
	// LinkEmbedCall is mirror.BuildEmbeddingCall — the embeddings
	// router (config → the provider's embed closure + label). The seam
	// returns the BUILT call, so tests inject a deterministic embed fn
	// and the production wiring IS the engine itself.
	LinkEmbedCall func(cfg any, modelOverride string) (
		mirror.EmbedFn, string, error)

	// The bot-queue tool doors' seam (T-79c) — faked at
	// construction per THE HERMETIC-SEAM law; the production
	// wiring is the SAME a.queueFetch composition the queue
	// command rides (the Worker /api/pending honest poll).
	// BotQueueFetch returns the state word ("ok" | "disabled" |
	// "error") + the pending row dicts.
	BotQueueFetch func(cfg *storage.OrderedMap) (string, []any)

	// SpawnOllamaServe starts `ollama serve` detached
	// (start_ollama_server's Popen half): (pid, "") or (0, errClass).
	SpawnOllamaServe func() (int, string)

	// SoundDir is the chime assets' folder
	// (APP_DIR/assets/sounds — sound.py's DEFAULT_SOUND_DIR).
	SoundDir func() string

	// SentinelRefresh re-arms the sentinel pass (the save law's
	// 250 ms singleShot; wired to the lifecycle's sentinel runner).
	SentinelRefresh func()

	// SourcesFetch is the Sources page's network door (urllib's twin;
	// nil → the production GET). T-74.
	SourcesFetch func(url string) (string, error)
	// OpenURL opens the OS browser at a URL (webbrowser.open's twin;
	// nil → the platform door). T-74.
	OpenURL func(url string)
}

// Settings is the Settings pages' command surface (one per app).
type Settings struct {
	Bus  *Bus
	Deps SettingsDeps
	lock *OpLock // the batch-busy guard (the quick-switch law)

	// cfg is the LIVE config map (the session-held instance; the
	// save-law's in-place target).
	cfg *storage.OrderedMap

	// mu serializes the settings' own doors (a second save while one
	// merges is queued, not interleaved) + the sound one-notice state.
	mu sync.Mutex

	// soundUnavailable is BatchSound's one-notice state (sound.py's
	// _unavailable_reason: set once, then silence for the session).
	soundUnavailable string

	// pagesBSourcesFetchDoor / pagesBOpenURLDoor are the pages B's
	// injected doors (T-74 — settingsb.go; nil → the production doors).
	pagesBSourcesFetchDoor func(url string) (string, error)
	pagesBOpenURLDoor      func(url string)

	// The Sync page's own state (T-75 — syncview.go): the injected
	// seams, the lazy ledger instance, the session's queue state (the
	// Bot page's own check lands it), and the last snapshot (the
	// modal's + the recompute's fuel).
	syncDeps      SyncDeps
	syncLedgerIns *syncreview.DeferredLog
	syncQueue     *SyncQueueState
	syncSnapshot  *SyncSnapshot
}

// NewSettings builds the surface. cfg is the LIVE map the app holds
// (nil → an empty map; the DI root always passes the real one).
func NewSettings(bus *Bus, cfg *storage.OrderedMap, lock *OpLock, deps SettingsDeps) *Settings {
	if cfg == nil {
		cfg = storage.NewOrderedMap()
	}
	return &Settings{Bus: bus, Deps: deps, lock: lock, cfg: cfg,
		pagesBSourcesFetchDoor: deps.SourcesFetch,
		pagesBOpenURLDoor:      deps.OpenURL,
	}
}

// logf speaks one line on the desk (the pages' engine lines ride the
// same transport the shell subscribes to).
func (s *Settings) logf(level, format string, args ...any) {
	if s.Bus != nil {
		s.Bus.EmitLogLine(fmt.Sprintf(format, args...), level)
	}
}

// ---------------------------------------------------------------------------
// The snapshot (one payload builds every page — the reference's
// initUI reads self.config once for all pages).
// ---------------------------------------------------------------------------

// SettingsSnapshot carries every page's initial state to the frontend.
type SettingsSnapshot struct {
	// Credentials
	TelegramAPIID      string `json:"telegramApiId"`
	TelegramAPIHash    string `json:"telegramApiHash"`
	TelegramPhone      string `json:"telegramPhone"`
	GitHubToken        string `json:"githubToken"`
	UserConfigStatus   string `json:"userConfigStatus"`
	UserConfigShowable bool   `json:"userConfigShowable"`

	// Proxy
	ProxyEnabled   bool   `json:"proxyEnabled"`
	ProxyType      string `json:"proxyType"`
	ProxyHost      string `json:"proxyHost"`
	ProxyPort      string `json:"proxyPort"`
	ProxyUseForWeb bool   `json:"proxyUseForWeb"`

	// Vault
	VaultPath        string   `json:"vaultPath"`
	VaultChoices     []string `json:"vaultChoices"`
	WebsiteVaultPath string   `json:"websiteVaultPath"`
	WebsiteRepoName  string   `json:"websiteRepoName"`
	ManualVaultPath  string   `json:"manualVaultPath"`
	ManualRepoName   string   `json:"manualRepoName"`
	WebBlocked       string   `json:"webBlocked"`
	WebSelf          string   `json:"webSelf"`
	PipelineGithub   bool     `json:"pipelineGithub"`
	PipelineWebsites bool     `json:"pipelineWebsites"`

	// LLM (the two-level radios + every engine's fields)
	LLMProvider        string `json:"llmProvider"` // ollama | llamacpp | cloud
	LLMNumCtx          string `json:"llmNumCtx"`
	LLMMaxOutputTokens string `json:"llmMaxOutputTokens"`
	OllamaURL          string `json:"ollamaUrl"`
	OllamaModel        string `json:"ollamaModel"`
	LlamaCppURL        string `json:"llamaCppUrl"`
	LlamaCppKey        string `json:"llamaCppKey"`
	LlamaCppModel      string `json:"llamaCppModel"`
	CloudURL           string `json:"cloudUrl"`
	CloudKey           string `json:"cloudKey"`
	CloudModel         string `json:"cloudModel"`

	// Sound
	SoundEnabled bool `json:"soundEnabled"`
	SoundVolume  int  `json:"soundVolume"` // the slider's 0-100 int

	// The always-banned law list (the Vault page's '?' help text).
	BlockedLawDomains []string `json:"blockedLawDomains"`

	// The pages B's fields (T-74 — settingsb.go's own struct, embedded
	// so one payload still builds every page).
	pagesBSnapshot
}

// Snapshot builds the payload: the live config + the vault discovery +
// the user-config quiet truth (populate_vaults' dedupe law: current +
// history + discovered, in order, no pre-selection — the combo OFFERS,
// the current value selects, an empty vault_path leaves it unselected).
func (s *Settings) Snapshot() *SettingsSnapshot {
	cfg := s.cfg
	if cfg == nil {
		cfg = storage.NewOrderedMap()
	}
	cfgPlain := cfgPlainMap(cfg)
	snap := &SettingsSnapshot{
		TelegramAPIID:   pyStrAny(cfgVal(cfg, "telegram_api_id")),
		TelegramAPIHash: pyStrCfg(cfg, "telegram_api_hash"),
		TelegramPhone:   pyStrCfg(cfg, "telegram_phone"),
		GitHubToken:     pyStrCfg(cfg, "github_token"),

		ProxyType: "socks5",
		ProxyHost: "127.0.0.1",
		ProxyPort: "10808",

		WebsiteVaultPath: strings.TrimSpace(pyStrCfg(cfg, "website_vault_path")),
		WebsiteRepoName:  strings.TrimSpace(pyStrCfg(cfg, "website_repo_name")),
		ManualVaultPath:  strings.TrimSpace(pyStrCfg(cfg, "manual_vault_path")),
		ManualRepoName:   strings.TrimSpace(pyStrCfg(cfg, "manual_repo_name")),
		// The fields show the LAW+extras union / the DEFAULT+config
		// union (ui.py reads blocked_domains_from_config /
		// self_domains_from_config — the save writes the text back).
		WebBlocked:       strings.Join(links.BlockedDomainsFromConfig(cfgPlain), ", "),
		WebSelf:          strings.Join(links.SelfDomainsFromConfig(cfgPlain), ", "),
		PipelineGithub:   true,
		PipelineWebsites: false,

		LLMProvider:        "ollama",
		LLMNumCtx:          strconv.Itoa(llmclient.DefaultNumCtx),
		LLMMaxOutputTokens: strconv.Itoa(llmclient.DefaultMaxOutputTokens),
		OllamaURL:          "http://localhost:11434",
		OllamaModel:        "qwythos-9b",
		LlamaCppURL:        llmclient.LLamaCppDefaultBase + "/v1",
		CloudURL:           "https://api.openai.com/v1",
		CloudModel:         "gpt-4o-mini",

		SoundEnabled: true,
		SoundVolume:  80,
		VaultChoices: []string{},
	}

	// -- proxy (the reference's per-field defaults) --
	if px, ok := cfg.Get("proxy"); ok {
		if pxm, is := px.(*storage.OrderedMap); is {
			if v, ok2 := pxm.Get("enabled"); ok2 {
				snap.ProxyEnabled, _ = v.(bool)
			}
			if v, ok2 := pxm.Get("type"); ok2 {
				if t, is := v.(string); is && t != "" {
					snap.ProxyType = t
				}
			}
			if v, ok2 := pxm.Get("host"); ok2 {
				if h, is := v.(string); is && h != "" {
					snap.ProxyHost = h
				}
			}
			if v, ok2 := pxm.Get("port"); ok2 {
				switch p := v.(type) {
				case int:
					snap.ProxyPort = strconv.Itoa(p)
				case int64:
					snap.ProxyPort = strconv.Itoa(int(p))
				case float64:
					snap.ProxyPort = strconv.Itoa(int(p))
				}
			}
			if v, ok2 := pxm.Get("use_for_web"); ok2 {
				snap.ProxyUseForWeb, _ = v.(bool)
			}
		}
	}

	// -- vault: the combo (populate_vaults) --
	current := strings.TrimSpace(pyStrCfg(cfg, "vault_path"))
	snap.VaultPath = current
	seen := map[string]bool{}
	add := func(v string) {
		v = strings.TrimSpace(v)
		if v == "" || seen[v] {
			return
		}
		seen[v] = true
		snap.VaultChoices = append(snap.VaultChoices, v)
	}
	add(current)
	if v, ok := cfg.Get("vaults_history"); ok {
		if arr, is := v.([]any); is {
			for _, h := range arr {
				if s2, is2 := h.(string); is2 {
					add(s2)
				}
			}
		}
	}
	if s.Deps.DiscoverVaults != nil {
		for _, v := range s.Deps.DiscoverVaults() {
			add(v)
		}
	}

	// -- pipelines --
	if pipes, ok := cfg.Get("pipelines"); ok {
		if pm, is := pipes.(*storage.OrderedMap); is {
			if v, ok2 := pm.Get("github"); ok2 {
				if b, is2 := v.(bool); is2 {
					snap.PipelineGithub = b
				}
			}
			if v, ok2 := pm.Get("websites"); ok2 {
				if b, is2 := v.(bool); is2 {
					snap.PipelineWebsites = b
				}
			}
		}
	}

	// -- LLM: the radios + the limits + every engine's fields --
	if p := strings.ToLower(strings.TrimSpace(pyStrCfg(cfg, "llm_provider"))); p == "cloud" || p == "llamacpp" || p == "ollama" {
		snap.LLMProvider = p
	}
	if v, ok := cfg.Get("llm_num_ctx"); ok {
		if n := pyIntOr(v, llmclient.DefaultNumCtx); n >= 0 {
			snap.LLMNumCtx = strconv.Itoa(n)
		}
	}
	if v, ok := cfg.Get("llm_max_output_tokens"); ok {
		if n := pyIntOr(v, llmclient.DefaultMaxOutputTokens); n >= 0 {
			snap.LLMMaxOutputTokens = strconv.Itoa(n)
		}
	}
	if oll, ok := cfg.Get("ollama"); ok {
		if om, is := oll.(*storage.OrderedMap); is {
			if v, ok2 := om.Get("base_url"); ok2 {
				if u, is2 := v.(string); is2 && u != "" {
					snap.OllamaURL = u
				}
			}
			if v, ok2 := om.Get("model"); ok2 {
				if m, is2 := v.(string); is2 && m != "" {
					snap.OllamaModel = m
				}
			}
		}
	}
	if u := strings.TrimSpace(pyStrCfg(cfg, "llamacpp_api_url")); u != "" {
		snap.LlamaCppURL = u
	}
	snap.LlamaCppKey = pyStrCfg(cfg, "llamacpp_api_key")
	snap.LlamaCppModel = strings.TrimSpace(pyStrCfg(cfg, "llamacpp_model"))
	if u := strings.TrimSpace(pyStrCfg(cfg, "cloud_api_url")); u != "" {
		snap.CloudURL = u
	}
	snap.CloudKey = pyStrCfg(cfg, "cloud_api_key")
	if m := strings.TrimSpace(pyStrCfg(cfg, "cloud_model")); m != "" {
		snap.CloudModel = m
	}

	// -- sound (the slider's int(round(v*100)); garbage/NaN heals to
	// the 0.8 default exactly like the reference's widget init) --
	if v, ok := cfg.Get("sound_enabled"); ok {
		if b, is := v.(bool); is {
			snap.SoundEnabled = b
		}
	}
	snap.SoundVolume = int(sliderPercent(soundVolumeValue(cfg)))

	// -- the user-config quiet truth --
	if s.Deps.UserConfigStatus != nil {
		snap.UserConfigStatus = s.Deps.UserConfigStatus()
	}
	if s.Deps.UserConfigPath != nil {
		snap.UserConfigShowable = s.Deps.UserConfigPath() != ""
	}

	snap.BlockedLawDomains = links.LawBlockedDomains
	snap.pagesBSnapshot = *s.snapshotPagesB(cfg)
	return snap
}

// pyIntOr reads a config number leniently (int/int64/float64 → int;
// junk → def).
func pyIntOr(v any, def int) int {
	switch n := v.(type) {
	case int:
		return n
	case int64:
		return int(n)
	case float64:
		return int(n)
	}
	return def
}

// pyStrAny renders a config value the way Python's str() does for the
// field-display types (numbers → their decimal form; strings as-is;
// anything else → "").
func pyStrAny(v any) string {
	switch n := v.(type) {
	case string:
		return n
	case int:
		return strconv.Itoa(n)
	case int64:
		return strconv.FormatInt(n, 10)
	case float64:
		return strconv.FormatFloat(n, 'f', -1, 64)
	}
	return ""
}

// sliderPercent mirrors int(round(v * 100)) for the volume's 0-100
// (Python's round-half-away-from-zero on the non-negative window).
func sliderPercent(v float64) float64 {
	return float64(int64(v*100 + 0.5))
}

// soundVolumeValue is the clamped [0,1] volume with the 0.8 heal
// (garbage/NaN defaults — the widget-init + _sound_volume_value law).
func soundVolumeValue(cfg *storage.OrderedMap) float64 {
	v := 0.8
	if raw, ok := cfg.Get("sound_volume"); ok {
		switch n := raw.(type) {
		case float64:
			v = n
		case int:
			v = float64(n)
		case int64:
			v = float64(n)
		default:
			return 0.8
		}
	}
	if v != v { // NaN
		return 0.8
	}
	if v < 0 {
		return 0
	}
	if v > 1 {
		return 1
	}
	return v
}

// ---------------------------------------------------------------------------
// The vault status ladder (_vault_path_status, ui.py's live cue).
// ---------------------------------------------------------------------------

// VaultStatus is one picker's live status row.
type VaultStatus struct {
	Text  string `json:"text"`  // "● found" …
	State string `json:"state"` // success | warning | error (the theme's semantic)
}

// VaultPathStatus is the ladder: an empty path warns "not set"; an
// existing folder is "found"; a missing Websites folder is "will be
// created" (its pipeline creates it); a missing GitHub vault is an
// error; a missing Manual vault is the owner's own to create.
// kind = "github" | "websites" | "manual".
func VaultPathStatus(path, kind string) VaultStatus {
	path = strings.TrimSpace(path)
	if path == "" {
		return VaultStatus{"● not set", "warning"}
	}
	if fsIsDirPath(path) {
		return VaultStatus{"● found", "success"}
	}
	switch kind {
	case "websites":
		return VaultStatus{"● will be created (folder does not exist yet)", "warning"}
	case "github":
		return VaultStatus{"● missing on disk", "error"}
	}
	return VaultStatus{"● not found — you create this vault yourself", "warning"}
}

// ---------------------------------------------------------------------------
// The ONE save door (the four laws — see the file doc).
// ---------------------------------------------------------------------------

// CredentialsValues is the Credentials page's group.
type CredentialsValues struct {
	APIID       string `json:"apiId"`
	APIHash     string `json:"apiHash"`
	Phone       string `json:"phone"`
	GitHubToken string `json:"githubToken"`
}

// ProxyValues is the Proxy page's group.
type ProxyValues struct {
	Enabled   bool   `json:"enabled"`
	Type      string `json:"type"`
	Host      string `json:"host"`
	Port      string `json:"port"` // the field's text; digits → int, else 10808
	UseForWeb bool   `json:"useForWeb"`
}

// VaultValues is the Vault page's group.
type VaultValues struct {
	// VaultPath is the combo's current text ("" = unselected — the
	// no-pre-selection law).
	VaultPath string `json:"vaultPath"`
	// History is the combo's law the frontend maintains: the current
	// path rides at the front (on_vault_changed's insert-at-0).
	History          []string `json:"history"`
	WebsiteVaultPath string   `json:"websiteVaultPath"`
	WebsiteRepoName  string   `json:"websiteRepoName"`
	ManualVaultPath  string   `json:"manualVaultPath"`
	ManualRepoName   string   `json:"manualRepoName"`
	WebBlocked       string   `json:"webBlocked"` // comma text → clean list (the field shows the law+extras union)
	WebSelf          string   `json:"webSelf"`
	PipelineGithub   *bool    `json:"pipelineGithub"`
	PipelineWebsites *bool    `json:"pipelineWebsites"`
}

// LLMValues is the LLM page's group (the two-level radios flatten to
// the same three-value llm_provider string the reference stores).
type LLMValues struct {
	Provider        string `json:"provider"` // ollama | llamacpp | cloud
	NumCtx          string `json:"numCtx"`   // lenient: digits ≥ 0 or keep config
	MaxOutputTokens string `json:"maxOutputTokens"`
	OllamaURL       string `json:"ollamaUrl"`
	OllamaModel     string `json:"ollamaModel"`
	LlamaCppURL     string `json:"llamaCppUrl"`
	LlamaCppKey     string `json:"llamaCppKey"`
	LlamaCppModel   string `json:"llamaCppModel"`
	CloudURL        string `json:"cloudUrl"`
	CloudKey        string `json:"cloudKey"`
	CloudModel      string `json:"cloudModel"`
}

// SoundValues is the Sound page's group.
type SoundValues struct {
	Enabled bool `json:"enabled"`
	Volume  int  `json:"volume"` // the slider's 0-100
}

// SettingsValues carries the per-page groups (the defensive hasattr
// pattern's Go shape: a nil group = "the page's widgets were not
// touched" = every key in it KEEPS the config value — save_config's
// own `else self.config.get(...)` law).
type SettingsValues struct {
	Credentials *CredentialsValues `json:"credentials,omitempty"`
	Proxy       *ProxyValues       `json:"proxy,omitempty"`
	Vault       *VaultValues       `json:"vault,omitempty"`
	LLM         *LLMValues         `json:"llm,omitempty"`
	Sound       *SoundValues       `json:"sound,omitempty"`
}

// SaveResult is the save door's verdict.
type SaveResult struct {
	Saved bool `json:"saved"`
}

// SaveSettings is THE ONE merge-based save (vaults_config.py
// save_config): each PRESENT group's keys land in the ui_values merge
// (absent groups untouched), the merged config is written atomically
// and mirrored to the user memory, the LIVE map is updated IN PLACE
// (holders see it — never a rebind), and the sentinel re-arms after
// the 250 ms singleShot beat.
func (s *Settings) SaveSettings(vals *SettingsValues) *SaveResult {
	if vals == nil {
		vals = &SettingsValues{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	ui := storage.NewOrderedMap()

	// The one hardcoded pass-through quirk: every reference save writes
	// log_level "INFO" (a hand-edited value resets — ui_values verbatim).
	ui.Set("log_level", "INFO")

	if vals.Credentials != nil {
		ui.Set("telegram_api_id", pyDigitsOr(vals.Credentials.APIID, 0))
		ui.Set("telegram_api_hash", strings.TrimSpace(vals.Credentials.APIHash))
		ui.Set("telegram_phone", strings.TrimSpace(vals.Credentials.Phone))
		ui.Set("github_token", strings.TrimSpace(vals.Credentials.GitHubToken))
	}

	if vals.Proxy != nil {
		px := storage.NewOrderedMap()
		px.Set("enabled", vals.Proxy.Enabled)
		px.Set("type", vals.Proxy.Type)
		px.Set("host", strings.TrimSpace(vals.Proxy.Host))
		px.Set("port", pyDigitsOr(vals.Proxy.Port, 10808))
		px.Set("use_for_web", vals.Proxy.UseForWeb)
		ui.Set("proxy", px)
	}

	if vals.Vault != nil {
		ui.Set("vault_path", vals.Vault.VaultPath)
		ui.Set("vaults_history", strSliceAny(vals.Vault.History))
		ui.Set("website_vault_path", strings.TrimSpace(vals.Vault.WebsiteVaultPath))
		ui.Set("website_repo_name", strings.TrimSpace(vals.Vault.WebsiteRepoName))
		ui.Set("manual_vault_path", strings.TrimSpace(vals.Vault.ManualVaultPath))
		ui.Set("manual_repo_name", strings.TrimSpace(vals.Vault.ManualRepoName))
		ui.Set("web_blocked_domains", commaDomains(vals.Vault.WebBlocked))
		ui.Set("web_self_domains", commaDomains(vals.Vault.WebSelf))
		pipes := storage.NewOrderedMap()
		if vals.Vault.PipelineGithub != nil {
			pipes.Set("github", *vals.Vault.PipelineGithub)
		}
		if vals.Vault.PipelineWebsites != nil {
			pipes.Set("websites", *vals.Vault.PipelineWebsites)
		}
		ui.Set("pipelines", pipes)
	}

	if vals.LLM != nil {
		switch strings.ToLower(strings.TrimSpace(vals.LLM.Provider)) {
		case "cloud", "llamacpp":
			ui.Set("llm_provider", strings.ToLower(strings.TrimSpace(vals.LLM.Provider)))
		default:
			ui.Set("llm_provider", "ollama")
		}
		ui.Set("llm_num_ctx", lenientIntField(vals.LLM.NumCtx, s.cfg, "llm_num_ctx", llmclient.DefaultNumCtx))
		ui.Set("llm_max_output_tokens", lenientIntField(vals.LLM.MaxOutputTokens, s.cfg, "llm_max_output_tokens", llmclient.DefaultMaxOutputTokens))
		oll := storage.NewOrderedMap()
		oll.Set("base_url", vals.LLM.OllamaURL)
		oll.Set("model", vals.LLM.OllamaModel)
		ui.Set("ollama", oll)
		ui.Set("llamacpp_api_url", strings.TrimSpace(vals.LLM.LlamaCppURL))
		ui.Set("llamacpp_api_key", strings.TrimSpace(vals.LLM.LlamaCppKey))
		ui.Set("llamacpp_model", strings.TrimSpace(vals.LLM.LlamaCppModel))
		ui.Set("cloud_api_url", strings.TrimSpace(vals.LLM.CloudURL))
		ui.Set("cloud_api_key", strings.TrimSpace(vals.LLM.CloudKey))
		ui.Set("cloud_model", strings.TrimSpace(vals.LLM.CloudModel))
	}

	if vals.Sound != nil {
		ui.Set("sound_enabled", vals.Sound.Enabled)
		ui.Set("sound_volume", clampSoundVolume(vals.Sound.Volume))
	}

	// Law 1: MERGE into the live config (unknown keys survive; nested
	// dicts merge key-by-key; tuning keys only defaulted).
	merged := storage.MergeConfig(s.cfg, ui)

	// Law 4: the atomic write + the credentials mirror, BEFORE the
	// in-place mutation — a failed write leaves the live map untouched
	// (the reference's own order).
	if s.Deps.Write != nil && !s.Deps.Write(merged) {
		return &SaveResult{Saved: false}
	}

	// Law 3: IN PLACE — clear + update, never rebind. Every holder of
	// the live map (a running batch, a startup pass) sees the save.
	replaceInPlace(s.cfg, merged)

	// The 250 ms sentinel re-arm (QTimer.singleShot's twin): the
	// top-bar cues (Net · AI · Proxy) refresh within a second of a
	// provider/proxy/URL edit. Best-effort, never blocks the save.
	if s.Deps.SentinelRefresh != nil {
		refresh := s.Deps.SentinelRefresh
		go func() {
			time.Sleep(250 * time.Millisecond)
			func() {
				defer func() { _ = recover() }()
				refresh()
			}()
		}()
	}
	return &SaveResult{Saved: true}
}

// replaceInPlace is the v30 law's mutation half: dst.clear() +
// dst.update(src) — dst keeps its identity so every holder sees src's
// values (never a rebind).
func replaceInPlace(dst, src *storage.OrderedMap) {
	for _, k := range dst.Keys() {
		dst.Delete(k)
	}
	for _, k := range src.Keys() {
		v, _ := src.Get(k)
		dst.Set(k, v)
	}
}

// pyDigitsOr mirrors the reference's int-if-digits field law:
// "123" → 123; anything else (empty, garbage) → def.
func pyDigitsOr(raw string, def int) any {
	raw = strings.TrimSpace(raw)
	if raw != "" && isAllDigits(raw) {
		if n, err := strconv.Atoi(raw); err == nil {
			return n
		}
	}
	return def
}

// lenientIntField is _llm_num_ctx_value / _llm_max_output_tokens_value:
// digits ≥ 0 are taken as-is; anything else keeps the previous config
// value, defaulting to def — a bad field can never break a save.
func lenientIntField(raw string, cfg *storage.OrderedMap, key string, def int) any {
	raw = strings.TrimSpace(raw)
	if raw != "" && isAllDigits(raw) {
		if n, err := strconv.Atoi(raw); err == nil && n >= 0 {
			return n
		}
	}
	if cfg != nil {
		if v, ok := cfg.Get(key); ok {
			if n := pyIntOr(v, def); n >= 0 {
				return n
			}
		}
	}
	return def
}

// isAllDigits mirrors str.isdigit (ASCII digits — the config fields'
// own law).
func isAllDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

// commaDomains is the blocked/self lists' text→list contract:
// comma-separated, each trimmed + lowered, empties dropped. An empty
// field is a deliberate "just the law" (blocked) / "none" (self) — []
// both ways.
func commaDomains(text string) []any {
	var out []any
	for _, d := range strings.Split(text, ",") {
		d = strings.ToLower(strings.TrimSpace(d))
		if d != "" {
			out = append(out, d)
		}
	}
	if out == nil {
		out = []any{}
	}
	return out
}

// strSliceAny converts for the config tree.
func strSliceAny(in []string) []any {
	out := make([]any, 0, len(in))
	for _, s := range in {
		out = append(out, s)
	}
	return out
}

// clampSoundVolume maps the slider's 0-100 to the file's 0.0-1.0
// (QSoundEffect's [0, 1] range — clamped on read AND write).
func clampSoundVolume(v int) float64 {
	f := float64(v) / 100.0
	if f < 0 {
		return 0
	}
	if f > 1 {
		return 1
	}
	return f
}

// RemoveVault is the combo's Remove door (remove_vault): the path
// leaves the history list, and vault_path clears when it WAS the
// current one. The folder itself is never touched (the confirm dialog
// lives in the frontend; this door only edits config).
func (s *Settings) RemoveVault(path string) *SaveResult {
	path = strings.TrimSpace(path)
	if path == "" {
		return &SaveResult{Saved: true}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	// history minus the removed path (in place on the live map).
	kept := []any{}
	if v, ok := s.cfg.Get("vaults_history"); ok {
		if arr, is := v.([]any); is {
			for _, h := range arr {
				if s2, is2 := h.(string); is2 && strings.TrimSpace(s2) == path {
					continue
				}
				kept = append(kept, h)
			}
		}
	}
	ui := storage.NewOrderedMap()
	ui.Set("vaults_history", kept)
	if strings.TrimSpace(pyStrCfg(s.cfg, "vault_path")) == path {
		ui.Set("vault_path", "")
	}
	merged := storage.MergeConfig(s.cfg, ui)
	if s.Deps.Write != nil && !s.Deps.Write(merged) {
		return &SaveResult{Saved: false}
	}
	replaceInPlace(s.cfg, merged)
	return &SaveResult{Saved: true}
}

// ---------------------------------------------------------------------------
// Credentials page: the token test + the About Me Wizard + the user
// config's show door.
// ---------------------------------------------------------------------------

// TokenTestResult is the token test's verdict for the field's own cue.
type TokenTestResult struct {
	OK    bool   `json:"ok"`
	Level string `json:"level"`
	Login string `json:"login"`
}

// TestGitHubToken validates the token AS TYPED (connection_tests.py
// test_github_token — reads the field, before any save): GET /user;
// 200 → the account login; 401 → the actionable rotation line; 403 →
// the forbidden line; transport → the network line. Never raises.
func (s *Settings) TestGitHubToken(token string) *TokenTestResult {
	token = strings.TrimSpace(token)
	if token == "" {
		s.logf("warning", "🔑 No GitHub token entered — the app will use anonymous access (60 requests/hour, 5000 with a token).")
		return &TokenTestResult{OK: false, Level: "warning"}
	}
	s.logf("info", "🔑 Testing GitHub token...")
	if s.Deps.GitHubGet == nil {
		return &TokenTestResult{OK: false, Level: "error"}
	}
	r := s.Deps.GitHubGet("/user", token, GitHubAPIBase, 15.0)
	switch {
	case r.Status == 0:
		s.logf("error", "❌ GitHub token test failed (network): %s", r.Err)
		return &TokenTestResult{OK: false, Level: "error"}
	case r.Status == 401:
		s.logf("error", "❌ GitHub token REJECTED (401 Bad credentials) — it is invalid, expired, or was rotated. Create a fresh token at github.com/settings/tokens (classic, 'repo' scope) and paste it here.")
		return &TokenTestResult{OK: false, Level: "error"}
	case r.Status == 403:
		s.logf("error", "❌ GitHub token forbidden (403): %s", githubPayloadString(r.Payload, r.Err))
		return &TokenTestResult{OK: false, Level: "error"}
	case r.Status != 200:
		s.logf("error", "❌ GitHub token test failed: HTTP %d", r.Status)
		return &TokenTestResult{OK: false, Level: "error"}
	}
	login := "?"
	if om, ok := r.Payload.(*storage.OrderedMap); ok {
		if v, ok2 := om.Get("login"); ok2 {
			if s2, is := v.(string); is && s2 != "" {
				login = s2
			}
		}
	}
	s.logf("success", "✅ GitHub token valid — account: %s (5000 requests/hour enabled)", login)
	return &TokenTestResult{OK: true, Level: "success", Login: login}
}

// githubPayloadString renders a 403 payload the way the reference
// printed e.data (the dict's message when shaped, else the raw).
func githubPayloadString(payload any, fallback string) string {
	if om, ok := payload.(*storage.OrderedMap); ok {
		if v, ok2 := om.Get("message"); ok2 {
			if s2, is := v.(string); is && s2 != "" {
				return s2
			}
		}
	}
	return fallback
}

// AboutMeFields is the wizard's six answers (empty = skipped).
type AboutMeFields struct {
	Name       string `json:"name"`
	Role       string `json:"role"`
	Interests  string `json:"interests"`
	Objectives string `json:"objectives"`
	Tech       string `json:"tech"`
	VaultUse   string `json:"vaultUse"`
}

// AboutMeResult is the wizard's verdict.
type AboutMeResult struct {
	OK      bool `json:"ok"`
	Chars   int  `json:"chars"`
	Written bool `json:"written"`
}

// BuildAboutMe is the wizard's pure half (lifecycle.py:133-152): the
// markdown template — every non-empty answer lands under its heading;
// ALL empty → the two default sections. Exported for the test.
func BuildAboutMe(f AboutMeFields) string {
	f = AboutMeFields{
		Name:       strings.TrimSpace(f.Name),
		Role:       strings.TrimSpace(f.Role),
		Interests:  strings.TrimSpace(f.Interests),
		Objectives: strings.TrimSpace(f.Objectives),
		Tech:       strings.TrimSpace(f.Tech),
		VaultUse:   strings.TrimSpace(f.VaultUse),
	}
	var b strings.Builder
	b.WriteString("# About Me\n\n")
	if f.Name != "" {
		b.WriteString("## Who I am\nI am a " + f.Name + ".\n\n")
	}
	if f.Role != "" {
		b.WriteString("## My role\n" + f.Role + "\n\n")
	}
	if f.Interests != "" {
		b.WriteString("## My interests\n" + f.Interests + "\n\n")
	}
	if f.Objectives != "" {
		b.WriteString("## My objectives\n" + f.Objectives + "\n\n")
	}
	if f.Tech != "" {
		b.WriteString("## Technologies I use\n" + f.Tech + "\n\n")
	}
	if f.VaultUse != "" {
		b.WriteString("## How I use this knowledge base\n" + f.VaultUse + "\n\n")
	}
	if f.Name == "" && f.Role == "" && f.Interests == "" && f.Objectives == "" && f.Tech == "" && f.VaultUse == "" {
		b.WriteString("## Who I am\nI am a software engineer who curates GitHub projects.\n\n")
		b.WriteString("## My objectives\nDiscover useful tools and frameworks for my development workflow.\n\n")
	}
	return b.String()
}

// GenerateAboutMe writes the file to the process working directory
// (open("about_me.md") — exactly where the runner reads it back,
// llmanalyze.go's CWD-relative door).
func (s *Settings) GenerateAboutMe(f AboutMeFields) *AboutMeResult {
	content := BuildAboutMe(f)
	err := os.WriteFile("about_me.md", []byte(content), 0o644)
	if err != nil {
		s.logf("error", "Failed to write about_me.md: %v", err)
		return &AboutMeResult{OK: false, Chars: len(content)}
	}
	s.logf("success", "📝 Generated about_me.md (%d chars)", len(content))
	s.logf("info", "The LLM will now personalize analyses based on your profile.")
	return &AboutMeResult{OK: true, Chars: len(content), Written: true}
}

// UserConfigStatusLine repaints the Credentials page's quiet truth
// (called after saves; cheap local read).
func (s *Settings) UserConfigStatusLine() string {
	if s.Deps.UserConfigStatus == nil {
		return ""
	}
	return s.Deps.UserConfigStatus()
}

// EmitImportFileSelected speaks the Input page's picker line
// (select_import_file's own words — the shell's picker door calls it).
func (s *Settings) EmitImportFileSelected(path string) {
	s.logf("info", "📄 Import file selected: %s", path)
}

// ShowUserConfigFile is the "Show the credentials file" door
// (open_user_config_location): open the OS file manager at the user
// config's folder, creating it when missing (the reference makedirs
// first — the file must be findable, not theoretical), the FILE
// selected when the platform supports it and it exists.
func (s *Settings) ShowUserConfigFile() bool {
	if s.Deps.UserConfigPath == nil {
		return false
	}
	path := s.Deps.UserConfigPath()
	if path == "" {
		return false
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		abs = path
	}
	folder := filepath.Dir(abs)
	_ = os.MkdirAll(folder, 0o755)
	if s.Deps.OpenFileLocation == nil {
		return false
	}
	if fileExistsAt(abs) {
		s.Deps.OpenFileLocation(abs)
	} else {
		s.Deps.OpenFileLocation(folder)
	}
	return true
}

// ---------------------------------------------------------------------------
// LLM page: the Ollama + llama.cpp engines.
// ---------------------------------------------------------------------------

// OllamaModelsResult is the Refresh door's payload.
type OllamaModelsResult struct {
	OK     bool     `json:"ok"`
	URL    string   `json:"url"`
	Models []string `json:"models"`
	// Model is the combo's new current text (the v30 persistence fix:
	// the typed custom name survives when it is not on the server).
	Model string `json:"model"`
}

// RefreshOllamaModels pulls the model list from the server at the
// field's URL (connection_tests.py refresh_ollama_models) and syncs
// config['ollama']['model'] IN PLACE (the v30 persistence fix — the
// next batch reads what the combo shows, even before a save). current
// is the combo's live text (a typed custom name is preserved at the
// top when the server does not list it).
func (s *Settings) RefreshOllamaModels(url, current string) *OllamaModelsResult {
	url = strings.TrimSpace(url)
	if url == "" {
		s.logf("error", "❌ Ollama URL is empty.")
		return &OllamaModelsResult{OK: false}
	}
	s.logf("info", "🔄 Refreshing models from %s...", url)
	if s.Deps.DetectOllama == nil {
		return &OllamaModelsResult{OK: false, URL: url}
	}
	r := s.Deps.DetectOllama(url, 8.0)
	if !r.Found {
		s.logf("error", "❌ Could not list models. Is Ollama running? Click '🚀 Start Ollama Server' first. Error: %s", r.Detail)
		return &OllamaModelsResult{OK: false, URL: url}
	}
	names := r.Models
	if len(names) == 0 {
		s.logf("warning", "⚠️ Ollama is running but no models are installed. Pull one with: ollama pull <model>")
		return &OllamaModelsResult{OK: true, URL: url, Models: []string{}}
	}
	// The combo's law: the current text survives (custom names kept at
	// the top when the server does not list them).
	model := strings.TrimSpace(current)
	serverNames := append([]string{}, names...)
	if !containsString(names, model) {
		if model == "" {
			model = names[0]
		} else {
			names = append([]string{model}, names...)
		}
	}
	// The in-place sync (ollama.model) — held maps see it immediately.
	s.mu.Lock()
	oll := storage.NewOrderedMap()
	if cur, ok := s.cfg.Get("ollama"); ok {
		if om, is := cur.(*storage.OrderedMap); is {
			oll = om
		}
	}
	oll.Set("model", model)
	s.cfg.Set("ollama", oll)
	s.mu.Unlock()
	// The log counts the SERVER's list (refresh_ollama_models' own
	// line — the combo's custom-name prepend is UI-only).
	s.logf("success", "✅ Found %d model(s): %s", len(serverNames), strings.Join(serverNames, ", "))
	return &OllamaModelsResult{OK: true, URL: url, Models: names, Model: model}
}

func containsString(haystack []string, needle string) bool {
	for _, h := range haystack {
		if h == needle {
			return true
		}
	}
	return false
}

// StartServerResult is the Start Server door's verdict.
type StartServerResult struct {
	OK  bool `json:"ok"`
	PID int  `json:"pid"`
}

// StartOllamaServer spawns `ollama serve` detached
// (connection_tests.py start_ollama_server): new session (POSIX) /
// new process group (Windows) so it survives the app; stdout+stderr to
// DEVNULL. The 3000 ms auto-refresh is the frontend's setTimeout
// (QTimer.singleShot(3000)'s twin).
func (s *Settings) StartOllamaServer() *StartServerResult {
	s.logf("info", "🚀 Starting Ollama server...")
	if s.Deps.SpawnOllamaServe == nil {
		s.logf("error", "❌ Failed to start Ollama: not wired")
		return &StartServerResult{}
	}
	pid, spawnErr := s.Deps.SpawnOllamaServe()
	switch {
	case spawnErr == "FileNotFoundError":
		s.logf("error", "❌ 'ollama' not found on PATH. Install it from https://ollama.com/download and make sure it's in your PATH.")
		return &StartServerResult{}
	case spawnErr != "":
		s.logf("error", "❌ Failed to start Ollama: %s", spawnErr)
		return &StartServerResult{}
	}
	s.logf("success", "✅ Ollama server starting in background (PID %d). Wait a moment, then click '🔄 Refresh Models'.", pid)
	return &StartServerResult{OK: true, PID: pid}
}

// LlamaDetectResult is the Detect/Refresh doors' payload.
type LlamaDetectResult struct {
	OK     bool     `json:"ok"`
	Found  bool     `json:"found"`
	URL    string   `json:"url"`
	Models []string `json:"models"`
	Model  string   `json:"model"`
	Saved  bool     `json:"saved"`
}

// DetectLlamaCpp is the Detect button (llamacpp.py
// detect_llamacpp_service): scan the running llama-server's process
// ports first, then the common ports; a find fills the URL + the model
// combo IN PLACE, THE SERVER CONFIG GRAB rides the probe, and the save
// is best-effort — never breaks Detect.
func (s *Settings) DetectLlamaCpp(liveKey string) *LlamaDetectResult {
	key := strings.TrimSpace(liveKey)
	s.logf("info", "🦙 Detecting llama.cpp server (running llama-server processes + ports %s)…",
		joinInts(llmclient.LLamaCppScanPorts))
	if s.Deps.DetectLlama == nil {
		return &LlamaDetectResult{}
	}
	probe := s.Deps.DetectLlama(key)
	if probe == nil || !probe.Found {
		s.logf("error", "❌ No llama.cpp server found (no llama-server process, nothing on the common ports). Start it with:\n   llama-server -m <model>.gguf --port 8080\n   (llama-server ships with llama.cpp — 'winget install ggml.llamacpp' or build from source), then click Detect again — or type a custom URL in the Server URL field.")
		return &LlamaDetectResult{OK: true, Found: false}
	}
	url := probe.BaseURL + "/v1"
	model := s.applyLlamaProbe(url, probe, probe.Model)
	s.logf("success", "✅ %s", probe.Detail)
	saved := s.saveLlamaGrab(probe)
	if model == "" {
		s.logf("warning", "⚠️ Server detected but no model is loaded — start it with -m <model>.gguf.")
	} else if probe.Ready != nil && !*probe.Ready {
		s.logf("warning", "⏳ The model is still loading — first real calls may wait until it is ready.")
	}
	return &LlamaDetectResult{OK: true, Found: true, URL: url,
		Models: probe.Models, Model: model, Saved: saved}
}

// RefreshLlamaCppModels is the Refresh button (llamacpp.py
// refresh_llamacpp_models): probe the URL in the field (no port scan),
// fill the combo, ride the props grab, best-effort save.
func (s *Settings) RefreshLlamaCppModels(liveURL, liveKey string) *LlamaDetectResult {
	url := strings.TrimSpace(liveURL)
	key := strings.TrimSpace(liveKey)
	if url == "" {
		s.logf("error", "❌ llama.cpp Server URL is empty.")
		return &LlamaDetectResult{OK: false}
	}
	s.logf("info", "🦙 Probing %s…", url)
	if s.Deps.ProbeLlama == nil {
		return &LlamaDetectResult{OK: false, URL: url}
	}
	probe := s.Deps.ProbeLlama(url, key, llmclient.LLamaCppProbeTimeoutS)
	if probe == nil || !probe.Found {
		detail := ""
		if probe != nil {
			detail = probe.Detail
		}
		s.logf("error", "❌ No llama.cpp server at %s (%s). Start llama-server, click '🔍 Detect' to scan the common ports, or fix the URL.", url, detail)
		return &LlamaDetectResult{OK: true, Found: false, URL: url}
	}
	model := s.applyLlamaProbe(url, probe, probe.Model)
	s.logf("success", "✅ %s", probe.Detail)
	saved := s.saveLlamaGrab(probe)
	return &LlamaDetectResult{OK: true, Found: true, URL: url,
		Models: probe.Models, Model: model, Saved: saved}
}

// applyLlamaProbe is the shared fill half (_fill_llamacpp_models + the
// URL's in-place sync): names filtered, the props model appended when
// missing, choice = pick || current || first; config['llamacpp_model']
// and config['llamacpp_api_url'] sync IN PLACE so the next batch sees
// them even before the save.
func (s *Settings) applyLlamaProbe(url string, probe *llmclient.ProbeResult, pick *string) string {
	names := []string{}
	for _, n := range probe.Models {
		if strings.TrimSpace(n) != "" {
			names = append(names, n)
		}
	}
	propsModel := ""
	if probe.PropsModel != nil {
		propsModel = strings.TrimSpace(*probe.PropsModel)
	}
	if propsModel != "" && !containsString(names, propsModel) {
		names = append(names, propsModel)
	}
	current := strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_model"))
	choice := ""
	if pick != nil {
		choice = strings.TrimSpace(*pick)
	}
	if choice == "" {
		choice = current
	}
	if choice == "" && len(names) > 0 {
		choice = names[0]
	}
	s.mu.Lock()
	s.cfg.Set("llamacpp_api_url", url)
	s.cfg.Set("llamacpp_model", choice)
	s.mu.Unlock()
	return choice
}

// saveLlamaGrab is THE SERVER CONFIG GRAB (_grab_llamacpp_limits):
// n_ctx → llm_num_ctx, n_predict → llm_max_output_tokens — the
// server's own launch values, only when it said them (nil touches
// nothing); the one success line lands when something changed; the
// save is best-effort (never breaks Detect/Refresh).
func (s *Settings) saveLlamaGrab(probe *llmclient.ProbeResult) bool {
	if probe.PropsLimits == nil {
		return false
	}
	changed := false
	var bits []string
	if probe.PropsLimits.NCtx != nil && *probe.PropsLimits.NCtx > 0 {
		nCtx := *probe.PropsLimits.NCtx
		if pyIntOr(cfgVal(s.cfg, "llm_num_ctx"), 0) != nCtx {
			s.cfg.Set("llm_num_ctx", nCtx)
			changed = true
		}
		bits = append(bits, fmt.Sprintf("context window %d tokens", nCtx))
	}
	if probe.PropsLimits.NPredict != nil && *probe.PropsLimits.NPredict > 0 {
		nPred := *probe.PropsLimits.NPredict
		if pyIntOr(cfgVal(s.cfg, "llm_max_output_tokens"), 0) != nPred {
			s.cfg.Set("llm_max_output_tokens", nPred)
			changed = true
		}
		bits = append(bits, fmt.Sprintf("output cap %d tokens", nPred))
	}
	if !changed {
		return false
	}
	s.logf("success", "🦙 llama.cpp server config grabbed: %s — Settings → 🧠 LLM → Limits now match the server's own launch values.", strings.Join(bits, " · "))
	// The best-effort save (llamacpp.py's try/except around save_config).
	s.SaveSettings(nil)
	return true
}

// cfgVal is the nil-safe Get.
func cfgVal(cfg *storage.OrderedMap, key string) any {
	if cfg == nil {
		return nil
	}
	v, _ := cfg.Get(key)
	return v
}

func joinInts(nums []int) string {
	parts := make([]string, len(nums))
	for i, n := range nums {
		parts[i] = strconv.Itoa(n)
	}
	return strings.Join(parts, ", ")
}

// ---------------------------------------------------------------------------
// The quick-switch fast lane (Detect & Set — llamacpp.py:445-687).
// ---------------------------------------------------------------------------

// QuickDetectResult is the probe half's payload; when Models holds
// several the frontend opens the model menu, else the SET half already
// ran with the single model.
type QuickDetectResult struct {
	OK         bool     `json:"ok"`
	Found      bool     `json:"found"`
	Provider   string   `json:"provider"`
	BaseURL    string   `json:"baseUrl"`
	Models     []string `json:"models"`
	Detail     string   `json:"detail"`
	PropsModel string   `json:"propsModel"`
	Ready      *bool    `json:"ready"`
	Via        string   `json:"via"`
	// Applied is true when the SET half already ran (the single-model
	// case); the several-models case returns the list for the menu, and
	// the frontend calls ApplyQuickDetect with the choice.
	Applied bool   `json:"applied"`
	Choice  string `json:"choice"`
}

// QuickDetect is the fast lane's probe half (_run_quick_detect's job):
// the LIVE widget URL overrides the config before the probe (a URL the
// user just typed is probed, not the stale one); ollama = ONE /api/tags
// detect; llamacpp = the configured URL first, then the process ports +
// common-port scan. The batch-running guard refuses first.
func (s *Settings) QuickDetect(provider, liveURL, liveKey string) *QuickDetectResult {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "ollama" && provider != "llamacpp" {
		provider = "ollama"
	}
	if s.lock != nil && s.lock.Busy() {
		s.logf("warning", "⏳ A batch is running — wait for it to finish before switching the LLM engine.")
		return &QuickDetectResult{OK: false, Provider: provider}
	}
	out := &QuickDetectResult{Provider: provider}
	switch provider {
	case "ollama":
		base := strings.TrimSpace(liveURL)
		if base == "" {
			base = cfgNestedStr(s.cfg, "ollama", "base_url", "http://127.0.0.1:11434")
		}
		if base == "" {
			base = "http://127.0.0.1:11434"
		}
		s.logf("info", "🧠 Detecting Ollama at %s…", base)
		if s.Deps.DetectOllama == nil {
			return out
		}
		r := s.Deps.DetectOllama(base, 4.0)
		out.OK = true
		out.Found = r.Found
		out.BaseURL = r.BaseURL
		out.Models = r.Models
		out.Detail = r.Detail
	default:
		key := strings.TrimSpace(liveKey)
		url := strings.TrimSpace(liveURL)
		if url == "" {
			url = strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_api_url"))
		}
		if url == "" {
			url = llmclient.LLamaCppDefaultBase
		}
		s.logf("info", "🦙 Detecting llama.cpp — %s first, then the running llama-server process + common ports…", url)
		if s.Deps.ProbeLlama == nil || s.Deps.DetectLlama == nil {
			return out
		}
		probe := s.Deps.ProbeLlama(url, key, llmclient.LLamaCppProbeTimeoutS)
		via := "configured"
		if probe == nil || !probe.Found {
			if det := s.Deps.DetectLlama(key); det != nil {
				probe = det
				via = det.Via
			}
		}
		if probe == nil {
			return out
		}
		out.OK = true
		out.Found = probe.Found
		out.BaseURL = probe.BaseURL
		if out.BaseURL == "" {
			out.BaseURL = url
		}
		out.Models = append([]string{}, probe.Models...)
		out.Detail = probe.Detail
		if probe.PropsModel != nil {
			out.PropsModel = *probe.PropsModel
		}
		out.Ready = probe.Ready
		out.Via = via
		if out.PropsModel != "" && !containsString(out.Models, out.PropsModel) {
			out.Models = append(out.Models, out.PropsModel)
		}
	}
	if !out.OK {
		return out
	}
	if !out.Found {
		if out.Provider == "ollama" {
			s.logf("error", "❌ 🧠 Detect & Set Ollama: Ollama is not running at %s (%s). Start the Ollama app, or Settings → 🧠 LLM → 🚀 Start Server, then click Detect & Set Ollama again.", out.BaseURL, out.Detail)
		} else {
			s.logf("error", "❌ 🦙 Detect & Set llama.cpp: No llama.cpp server found (%s). Start it with:\n   llama-server -m <model>.gguf --port 8080\n   then click Detect & Set llama.cpp again.", out.Detail)
		}
		return out
	}
	base := out.BaseURL
	models := []string{}
	for _, m := range out.Models {
		if strings.TrimSpace(m) != "" {
			models = append(models, m)
		}
	}
	loading := ""
	if out.Provider == "llamacpp" && out.Ready != nil && !*out.Ready {
		loading = " · model still LOADING…"
	}
	if len(models) == 0 {
		if out.Provider == "ollama" {
			s.logf("warning", "⚠️ 🧠 Ollama is up at %s but NO models are installed. Pull one with: ollama pull <model>, then click Detect & Set Ollama again.", base)
		} else {
			s.logf("warning", "⚠️ 🦙 llama.cpp is up at %s but no model could be read%s. Start llama-server with -m <model>.gguf (or wait for it to finish loading), then click Detect & Set llama.cpp again.", base, loading)
		}
		return out
	}
	if len(models) == 1 {
		s.logf("success", "✅ %s %s detected at %s · model '%s'%s", quickIcon(out.Provider), quickLabel(out.Provider), base, models[0], loading)
		s.ApplyQuickDetect(out.Provider, base, models, models[0], nil)
		out.Applied = true
		out.Choice = models[0]
		return out
	}
	// Several models: the menu is the frontend's half ("a menu like the
	// current one should help user to select their desired model").
	s.logf("success", "✅ %s %s detected at %s · %d models — pick one:", quickIcon(out.Provider), quickLabel(out.Provider), base, len(models))
	return out
}

func quickIcon(provider string) string {
	if provider == "ollama" {
		return "🧠"
	}
	return "🦙"
}

func quickLabel(provider string) string {
	if provider == "ollama" {
		return "Ollama"
	}
	return "llama.cpp"
}

// ApplyQuickDetectResult is the SET half's verdict.
type ApplyQuickDetectResult struct {
	Applied bool `json:"applied"`
}

// ApplyQuickDetect is the SET half (_apply_quick_detect): provider +
// URL + model land in the LIVE config (in place) and merge-save, so
// the next batch uses them immediately. choice "" = the menu's Cancel
// — the provider was NOT changed (the honest line, nothing else
// touched). live is the frontend's current LLM field state (the
// snapshot law: unsaved edits count) — the SET values overlay it, so
// the save carries the other engines' current truth instead of
// clobbering it; nil builds the group from the live config.
func (s *Settings) ApplyQuickDetect(provider, base string, models []string, choice string, live *LLMValues) *ApplyQuickDetectResult {
	provider = strings.ToLower(strings.TrimSpace(provider))
	if provider != "ollama" {
		provider = "llamacpp"
	}
	choice = strings.TrimSpace(choice)
	if choice == "" {
		s.logf("info", "⏭️ Cancelled — the LLM provider was NOT changed (still %s).", strings.ToLower(strings.TrimSpace(pyStrCfg(s.cfg, "llm_provider"))))
		return &ApplyQuickDetectResult{}
	}
	isOllama := provider == "ollama"

	// The group: the live fields (or the config's current truth) with
	// the SET values overlaid — the reference's save reads every
	// widget; the untouched engines keep what they show.
	var llm *LLMValues
	if live != nil {
		cp := *live
		llm = &cp
	} else {
		llm = &LLMValues{
			NumCtx:          strconv.Itoa(pyIntOr(cfgVal(s.cfg, "llm_num_ctx"), llmclient.DefaultNumCtx)),
			MaxOutputTokens: strconv.Itoa(pyIntOr(cfgVal(s.cfg, "llm_max_output_tokens"), llmclient.DefaultMaxOutputTokens)),
			OllamaURL:       cfgNestedStr(s.cfg, "ollama", "base_url", "http://localhost:11434"),
			OllamaModel:     cfgNestedStr(s.cfg, "ollama", "model", ""),
			LlamaCppURL:     strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_api_url")),
			LlamaCppKey:     pyStrCfg(s.cfg, "llamacpp_api_key"),
			LlamaCppModel:   strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_model")),
			CloudURL:        strings.TrimSpace(pyStrCfg(s.cfg, "cloud_api_url")),
			CloudKey:        pyStrCfg(s.cfg, "cloud_api_key"),
			CloudModel:      strings.TrimSpace(pyStrCfg(s.cfg, "cloud_model")),
		}
		if cur := strings.ToLower(strings.TrimSpace(pyStrCfg(s.cfg, "llm_provider"))); cur == "cloud" || cur == "llamacpp" || cur == "ollama" {
			llm.Provider = cur
		}
	}
	llm.Provider = provider
	if isOllama {
		llm.OllamaURL = base
		llm.OllamaModel = choice
	} else {
		llm.LlamaCppURL = strings.TrimRight(base, "/") + "/v1"
		llm.LlamaCppModel = choice
	}
	res := s.SaveSettings(&SettingsValues{LLM: llm})
	if !res.Saved {
		return &ApplyQuickDetectResult{}
	}
	shown := base
	if !isOllama {
		shown = strings.TrimRight(base, "/") + "/v1"
	}
	s.logf("success", "✅ %s LLM provider SET to %s — %s · model '%s' — saved. The next batch uses it immediately.", quickIcon(provider), quickLabel(provider), shown, choice)
	return &ApplyQuickDetectResult{Applied: true}
}

// ---------------------------------------------------------------------------
// The Sound page: the preview gate (the exact batch path's decision
// half; the playback itself is the frontend's audio element).
// ---------------------------------------------------------------------------

// SoundPreviewResult is the preview door's verdict + payload.
type SoundPreviewResult struct {
	Play   bool    `json:"play"`
	Volume float64 `json:"volume"`
	// DataURL is the wav as a data: URL when playable (the frontend's
	// <audio> src); empty when not.
	DataURL string `json:"dataUrl"`
}

// soundFor is THE EXACT BATCH GATE (quiet): the switch, the clamped
// volume, the one-notice missing-asset law (sound.py's contract). The
// preview's 🔇 line is the preview wrapper's own; this door stays
// silent on a switched-off chime (the reference's _play_batch_sound
// returns False without a word). The lookup is DISK FIRST (the
// APP_DIR/assets/sounds copy — an owner can swap the chime by dropping
// a file), the EMBEDDED default second (T-77 ships the files).
func (s *Settings) soundFor(needsRetry bool) *SoundPreviewResult {
	wav := "success.wav"
	if needsRetry {
		wav = "retry.wav"
	}
	if v, ok := s.cfg.Get("sound_enabled"); ok {
		if b, is := v.(bool); is && !b {
			return &SoundPreviewResult{}
		}
	}
	// (an absent key defaults TRUE — the reference's config.get law.)
	s.mu.Lock()
	unavailable := s.soundUnavailable
	s.mu.Unlock()
	if unavailable != "" {
		// BatchSound already said it once — silence (rule 5).
		return &SoundPreviewResult{}
	}
	var data []byte
	if s.Deps.SoundDir != nil {
		if dir := s.Deps.SoundDir(); dir != "" {
			if raw, err := os.ReadFile(filepath.Join(dir, wav)); err == nil {
				data = raw
			}
		}
	}
	if data == nil {
		// The embedded default (T-77): the shipped chime backs the
		// missing disk copy up.
		if needsRetry {
			data = assets.SoundRetryWav
		} else {
			data = assets.SoundSuccessWav
		}
	}
	if len(data) == 0 {
		s.mu.Lock()
		s.soundUnavailable = fmt.Sprintf("%s chime missing (neither the disk copy nor the embedded default answered)", wav)
		reason := s.soundUnavailable
		s.mu.Unlock()
		s.logf("warning", "🔕 Success chime disabled — %s", reason)
		return &SoundPreviewResult{}
	}
	return &SoundPreviewResult{
		Play:    true,
		Volume:  soundVolumeValue(s.cfg),
		DataURL: "data:audio/wav;base64," + base64.StdEncoding.EncodeToString(data),
	}
}

// BatchChime is the batch-finish door (T-77): the exact batch gate,
// quiet — the shell enriches the batch-finished crossing with this
// payload and the frontend's <audio> plays it (the QSoundEffect twin).
func (s *Settings) BatchChime(needsRetry bool) *SoundPreviewResult {
	return s.soundFor(needsRetry)
}

// PreviewSound is _preview_batch_sound's decision half: the exact batch
// gate + the preview's OWN lines — a switched-off chime logs the 🔇
// line (the batch path stays silent); a missing asset logs BatchSound's
// one 🔕 notice then stays silent for the session; a playable asset
// returns the payload for the frontend's <audio> (a machine without an
// audio backend accepts the play and stays silent, never an error).
func (s *Settings) PreviewSound(needsRetry bool) *SoundPreviewResult {
	res := s.soundFor(needsRetry)
	if !res.Play {
		if v, ok := s.cfg.Get("sound_enabled"); ok {
			if b, is := v.(bool); is && !b {
				s.logf("info", "🔇 Preview silent — the chime is switched off (the switch above).")
				return res
			}
		}
		// BatchSound already said the 🔕 once — the preview does not
		// repeat it (the branch exists so the log never blames the
		// switch when the machine is audio-less).
	}
	return res
}

func fileExistsAt(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}
