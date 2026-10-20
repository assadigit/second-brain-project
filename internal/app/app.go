// app.go — T-71 slice (d): the DI root (doc 04 §2's App struct + §3's
// law: "app — the only package allowed to construct everything; no
// business logic"). New() wires the production seams over the real
// engine packages; every door stays injectable (the hermetic law —
// the contract tests in commands_test.go already proved the
// transport; a wiring smoke test proves the composition below).
//
// THE BATCH WIRING (T-74, slice 1 — the CLI-parity re-host): the app's
// batch engine is now engine.go's runImpl composition re-hosted on
// this surface (batchwire.go): the warmup family, the inputs fetchers,
// the setup family's full intake (the banned/self drops + the _inbox
// tables), the repos banishment + settled ledger + backfill + the §4.4
// both-vault passes, the websites phase, the finish seals + THE
// PHASE-5 CLEAR + the dry-run law. The honest-scope debt T-73 carried
// is closed; the shell (webui) drives exactly the composition the CLI
// drives, over the same engine packages.
package app

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/config"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/goodrepos"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/mirror"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/reporadar"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/vaulthygiene"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
	"github.com/assadigit/second-brain-project/internal/vaultseal"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// App is the composition root: the config source, the event bus, the
// command surface, the settings surface, the lifecycle, and the
// injectable doors the engine compositions read.
type App struct {
	// ConfigPath is the live config file (the config store's home).
	ConfigPath string

	// Bus is the typed event bus (doc 04 §2).
	Bus *Bus
	// Commands is the command surface (doc 04 §5).
	Commands *Commands
	// Settings is the Settings pages' command surface (T-73).
	Settings *Settings
	// Lifecycle owns the startup passes + the watchdog + shutdown.
	Lifecycle *Lifecycle

	// liveCfg is the SESSION-HELD config map (the reference's
	// self.config): loaded ONCE at New() through the GUI load path
	// (config.LoadAppConfig — fresh detection, the example base, THE
	// FILL, the credential strip heal), read by every door, mutated IN
	// PLACE by every save (vaults_config.py's MainWindow semantics).
	liveCfg *storage.OrderedMap

	// configFilled is the user-memory fill count (the one quiet
	// desk line after the UI exists — window.py:155-173).
	configFilled int

	// sentinelOnline is the sentinel's last observed verdict (nil =
	// never observed — the offline law's first-transition seat,
	// window.py's _sentinel_online). Guarded by sentinelMu: TWO passes
	// can run concurrently (the lifecycle's recurring tick + the save
	// law's 250 ms re-arm) and the read-transition-write must be one
	// unit (the race detector's own catch — two concurrent first
	// passes would both warn-now).
	sentinelMu     sync.Mutex
	sentinelOnline *bool

	// batchStop is the ACTIVE batch's cooperative-stop door (nil when
	// no batch runs) — Commands.StopBatch reaches it through
	// requestBatchStop. Guarded by batchStopMu (the hero's Stop click
	// can land on any goroutine). The handle's identity is the arm
	// token: a late clear from a finished batch never disarms a newer
	// one (the pointer compare is the whole point of the wrapper).
	batchStopMu sync.Mutex
	batchStop   *batchStopHandle

	// SentinelProbe is the sentinel pass's probe seam (the HERMETIC-SEAM
	// law: the tests fake the probes at construction; New() wires
	// connectioncheck.SentinelSnapshot).
	SentinelProbe func(cfg any, timeoutS float64) connectioncheck.Snapshot

	// The engine doors (the hermetic law — New() wires the production
	// defaults; tests and staged shells may swap them BEFORE any
	// command runs).
	GHClientFactory func(token string) githubpipeline.RepoFetcher
	OllamaFactory   func(host string) llmclient.OllamaAPI
	BannerDownload  func(owner, repoName, folder string) string

	// The batch wiring's own doors (T-74, slice 1 — batchwire.go's
	// seams; the warmup family + the Phase-5 CLEAR's network legs,
	// fakeable at construction per the hermetic law):
	// ProbeLlama is the one-address llama.cpp probe
	// (llmclient.ProbeLlamaCpp).
	ProbeLlama func(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult
	// DetectLlama is the process-ports + common-ports scan
	// (llmclient.DetectLlamaCpp).
	DetectLlama func(apiKey string) *llmclient.ProbeResult
	// AutostartServe spawns `ollama serve` detached ((pid, "") or
	// (0, errClass)).
	AutostartServe func() (int, string)
	// CloudListModels is the cloud /models pre-flight door (the
	// Anthropic/OpenAI flavor split by URL).
	CloudListModels func(apiURL, apiKey string, timeoutS float64) ([]string, error)
	// WorkerMirrorPush is THE PHASE-5 CLEAR's ack leg (the disabled
	// Worker is silent — the (false,false,nil) law).
	WorkerMirrorPush func(cfg *storage.OrderedMap, entries, decommissionEvents []any, vaultIndexHash string) (bool, bool, any)
	// WriteLiveConfig persists the live map (config.Save's atomic +
	// mirror law); the model-changed saves ride it.
	WriteLiveConfig func(cfg *storage.OrderedMap) bool

	// appLock is the single-instance lock (nil until Acquire).
	appLock *AppLock
}

// LiveConfig returns the session-held config map (the GUI load path's
// one load; save doors mutate this instance in place — the running
// workers see saves through it).
func (a *App) LiveConfig() *storage.OrderedMap {
	if a.liveCfg == nil {
		a.liveCfg = storage.NewOrderedMap()
	}
	return a.liveCfg
}

// New builds the whole composition over the config at configPath.
// headless mirrors the launch mode: no GUI will ever answer a gate.
func New(configPath string, headless bool) *App {
	a := &App{ConfigPath: configPath}
	a.Bus = NewBus()
	a.Bus.Headless = headless

	// THE LIVE-CONFIG LOAD (the GUI path, once per session):
	// config.LoadAppConfig — fresh detection, the example base, THE
	// FILL (the fill count for the one quiet desk line), the v0.07.1
	// credential strip heal. Every door reads this instance; every
	// save mutates it in place.
	loadRes := config.LoadAppConfig(configPath)
	a.liveCfg = loadRes.Config
	a.configFilled = loadRes.Filled

	a.GHClientFactory = func(token string) githubpipeline.RepoFetcher {
		return githubpipeline.NewClient(token)
	}
	a.OllamaFactory = func(host string) llmclient.OllamaAPI {
		return llmclient.NewOllamaClient(host)
	}
	a.BannerDownload = nil // the runner's real DownloadBanner

	// The batch wiring's production doors (T-74, slice 1).
	a.ProbeLlama = llmclient.ProbeLlamaCpp
	a.DetectLlama = func(apiKey string) *llmclient.ProbeResult {
		return llmclient.DetectLlamaCpp(apiKey, llmclient.LLamaCppScanPorts,
			"127.0.0.1", llmclient.LLamaCppProbeTimeoutS, true)
	}
	a.AutostartServe = spawnOllamaServe
	a.CloudListModels = func(apiURL, apiKey string, timeoutS float64) ([]string, error) {
		if llmclient.IsAnthropicURL(apiURL) {
			return llmclient.AnthropicListModels(apiURL, apiKey, timeoutS)
		}
		return llmclient.OpenAIListModels(apiURL, apiKey, timeoutS)
	}
	a.WorkerMirrorPush = func(cfg *storage.OrderedMap, entries,
		decommissionEvents []any, vaultIndexHash string) (bool, bool, any) {
		s := workerclient.NewSync(cfgPlainMap(cfg))
		if !s.IsEnabled() {
			return false, false, nil
		}
		ok, resp := s.PushVaultMirror(entries, decommissionEvents, vaultIndexHash)
		return true, ok, resp
	}
	a.WriteLiveConfig = func(cfg *storage.OrderedMap) bool {
		return config.Save(configPath, cfg)
	}
	a.Commands = NewCommands(a.Bus, Deps{
		Config:        a.loadConfig,
		QueueFetch:    a.queueFetch,
		QueueClassify: ClassifyQueueLinks,
		Batch:         a.runBatch,
		DriftSeal:     a.driftSeal,
		Scan:          a.scan,
		ApplyPlan:     a.applyScanPlan,
		SealVault:     a.sealVault,
		Resync:        a.resync,
		BackupNow:     RunBackupNow,
		Publish:       a.publish,
		Radar:         a.radar,
	})
	a.Lifecycle = NewLifecycle(a.Bus, &a.Commands.batchLock, StartupPasses{
		// The startup passes' engines: the sentinel's probe + the
		// llamacpp autodetect + the both-vaults heal + the review
		// backlog notice + the quiet radar (each engine's own gates
		// — vault_hygiene_on_start, radar_pending — live in the
		// engines, exactly as the reference's jobs check them).
		Sentinel:     a.sentinelPass,
		LlamaDetect:  a.llamaAutodetect,
		Heal:         a.startupHeal,
		BacklogCheck: a.backlogNotice,
		Radar:        a.quietRadar,
	})
	// The cooperative-stop door (T-77): the hero's Stop click reaches
	// the active batch's requestStop through the command surface.
	a.Commands.stopDoor = a.requestBatchStop
	// The sentinel's probe seam (the hermetic law).
	a.SentinelProbe = connectioncheck.SentinelSnapshot
	a.Settings = NewSettings(a.Bus, a.liveCfg, &a.Commands.batchLock, SettingsDeps{
		Config: a.loadConfig,
		Write: func(cfg *storage.OrderedMap) bool {
			return config.Save(configPath, cfg)
		},
		DiscoverVaults: batch.FindObsidianVaults,
		UserConfigStatus: func() string {
			return config.StatusLine("")
		},
		UserConfigPath: config.GlobalConfigPath,
		OpenFileLocation: func(path string) {
			openFileLocationOS(path)
		},
		GitHubGet:    connectioncheck.GitHubGetFn,
		DetectOllama: llmclient.DetectOllama,
		ProbeLlama:   llmclient.ProbeLlamaCpp,
		DetectLlama: func(apiKey string) *llmclient.ProbeResult {
			return llmclient.DetectLlamaCpp(apiKey, llmclient.LLamaCppScanPorts,
				"127.0.0.1", llmclient.LLamaCppProbeTimeoutS, true)
		},
		// The per-test doors' production seams (T-79a).
		OllamaChatPing: func(host, model, prompt string, opts *llmclient.OllamaChatOptions, timeoutS float64) (string, string) {
			reply, err := llmclient.OllamaChat(llmclient.NewOllamaClient(host), model,
				[]llmclient.Message{{Role: "user", Content: prompt}}, timeoutS, opts)
			if err != nil {
				return "", err.Error()
			}
			return strings.TrimSpace(reply), ""
		},
		CloudChatPing: func(apiURL, apiKey, model, prompt string, timeoutS float64) (string, string) {
			reply, err := llmclient.CloudChat(apiURL, apiKey, model,
				[]llmclient.Message{{Role: "user", Content: prompt}}, timeoutS, nil)
			if err != nil {
				return "", err.Error()
			}
			return strings.TrimSpace(reply), ""
		},
		PreflightCloud: llmclient.PreflightCloud,
		WorkerCheck: func(config any) *connectioncheck.Result {
			return connectioncheck.CheckWorkerVersion(config, nil)
		},
		ProxyCheck: connectioncheck.ProxyReachable,
		// The linking-tools doors' production seams (T-79b): the
		// tools' _build_llm_call chat lanes (JSON mode on; num_ctx only
		// when the config speaks one) + the embeddings router itself.
		LinkOllamaChat: func(host, model string, messages []llmclient.Message,
			numCtx int, timeoutS float64) (string, string) {
			opts := &llmclient.OllamaChatOptions{}
			if numCtx > 0 {
				opts.NumCtx = numCtx
			}
			reply, err := llmclient.OllamaChat(llmclient.NewOllamaClient(host),
				model, messages, timeoutS, opts)
			if err != nil {
				return "", err.Error()
			}
			return reply, ""
		},
		LinkCloudChat: func(apiURL, apiKey, model string, messages []llmclient.Message,
			numCtx int, timeoutS float64) (string, string) {
			co := &llmclient.ChatOptions{JSONMode: true}
			if numCtx > 0 {
				co.NumCtx = numCtx
			}
			reply, err := llmclient.CloudChat(apiURL, apiKey, model, messages,
				timeoutS, co)
			if err != nil {
				return "", err.Error()
			}
			return reply, ""
		},
		LinkEmbedCall:    mirror.BuildEmbeddingCall,
		SpawnOllamaServe: spawnOllamaServe,
		SoundDir: func() string {
			return filepath.Join(constants.AppDir(), "assets", "sounds")
		},
		SentinelRefresh: func() {
			if a.Lifecycle != nil {
				a.Lifecycle.RunSentinelPass()
			}
		},
	})
	return a
}

// setBatchStop/clearBatchStop register the active engine's stop door
// around each run.
func (a *App) setBatchStop(h *batchStopHandle) {
	a.batchStopMu.Lock()
	a.batchStop = h
	a.batchStopMu.Unlock()
}

func (a *App) clearBatchStop(h *batchStopHandle) {
	a.batchStopMu.Lock()
	if a.batchStop == h {
		a.batchStop = nil
	}
	a.batchStopMu.Unlock()
}

// requestBatchStop flips the active batch's cooperative stop flag.
// Returns false when no batch is running (the honest no-op).
func (a *App) requestBatchStop() bool {
	a.batchStopMu.Lock()
	h := a.batchStop
	a.batchStopMu.Unlock()
	if h == nil {
		return false
	}
	h.fn()
	return true
}

// loadConfig is the live-config door: the ONE session-held map (the
// GUI load at New()) — never a re-read; saves mutate it in place, so
// every caller (commands, passes, the settings surface) reads the
// same truth a running worker holds (vaults_config.py's law).
func (a *App) loadConfig() *storage.OrderedMap {
	if a.liveCfg == nil {
		loadRes := config.LoadAppConfig(a.ConfigPath)
		a.liveCfg = loadRes.Config
		a.configFilled = loadRes.Filled
	}
	return a.liveCfg
}

// TheFillLine is the one quiet desk line after the UI exists
// (window.py:155-173): when THIS launch restored credentials from the
// user memory, the owner sees that the fields filled themselves, and
// from where. Empty string = nothing was restored (the common case).
func (a *App) TheFillLine() string {
	if a.configFilled <= 0 {
		return ""
	}
	plural := "s"
	if a.configFilled == 1 {
		plural = ""
	}
	where := config.GlobalConfigPath()
	if where == "" {
		where = "the user config file"
	}
	return fmt.Sprintf("🔑 %d credential%s restored from %s — the user config file remembers them across versions.",
		a.configFilled, plural, where)
}

// openFileLocationOS opens the OS file manager at path (the file
// selected when the platform supports it) — open_user_config_location's
// platform legs. Best-effort: a failed opener is never an error.
func openFileLocationOS(path string) {
	defer func() { _ = recover() }()
	folder := filepath.Dir(path)
	switch runtime.GOOS {
	case "windows":
		if fileExistsAt(path) {
			_ = exec.Command("explorer", "/select,", path).Start()
		} else {
			_ = exec.Command("explorer", folder).Start()
		}
	case "darwin":
		if fileExistsAt(path) {
			_ = exec.Command("open", "-R", path).Start()
		} else {
			_ = exec.Command("open", folder).Start()
		}
	default:
		_ = exec.Command("xdg-open", folder).Start()
	}
}

// queueFetch is the Worker /api/pending poll (workerclient.Sync's
// honest door — the Option-B queue source).
func (a *App) queueFetch(cfg *storage.OrderedMap) (string, []any) {
	s := workerclient.NewSync(cfgPlainMap(cfg))
	return s.GetPendingResult()
}

// driftSeal adapts vaultseal.RunDriftSeal (the pure engine with its
// default deps).
func (a *App) driftSeal(cfg *storage.OrderedMap, manual bool, log LogFn) (string, string) {
	return vaultseal.RunDriftSeal(cfg, vaultseal.DefaultDriftSealDeps(), log, manual)
}

// scan adapts RunVaultScan (the composition's ctx flows through the
// command; the engine itself is cooperative per engine law).
func (a *App) scan(ctx context.Context, cfg *storage.OrderedMap, vaults []string, ev ScanEvents) *ScanResult {
	return RunVaultScan(cfg, vaults, ev)
}

// applyScanPlan opens the config's vault DBs and runs
// vaultscan.ApplyScanPlan (the apply door's engine).
func (a *App) applyScanPlan(cfg *storage.OrderedMap, kind string, plan *vaultscan.Plan, selection map[string][]string, log LogFn) *ApplyResult {
	vaultPath := pyStrCfg(cfg, "vault_path")
	if kind == "websites" {
		vaultPath = pyStrCfg(cfg, "website_vault_path")
	}
	var state *websitepipeline.WebsiteStateDB
	if kind == "websites" {
		if st, err := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(cfg)); err == nil {
			state = st
			defer func() {
				func() {
					defer func() { recover() }()
					state.Close()
				}()
			}()
		}
	}
	var cacheDB *store.CacheDB
	if rc, err := store.NewCacheDB(cacheDBPathOf(cfg)); err == nil {
		cacheDB = rc
		defer func() { _ = cacheDB.Close() }()
	}
	var noteStateDB *notestate.DB
	if nd, err := notestate.New(cacheDBPathOf(cfg)); err == nil {
		noteStateDB = nd
		defer func() {
			func() {
				defer func() { recover() }()
				_ = noteStateDB.Close()
			}()
		}()
	}
	rep := vaultscan.ApplyScanPlan(plan, vaultPath, state, log,
		vaultscan.NormalizeSelection(selection), kind, cacheDB, noteStateDB)
	return &ApplyResult{Success: true, Report: &rep}
}

// sealVault routes the kind to vaultseal's three bridges (the
// VaultSealWorker starters' engines).
func (a *App) sealVault(cfg *storage.OrderedMap, kind string, log LogFn) *vaultseal.SealResult {
	switch kind {
	case SealRepos:
		return vaultseal.SealFromConfig(cfg, &vaultseal.RunSummary{}, log)
	case SealWebsites:
		return vaultseal.WebsitesSealFromConfig(cfg, &vaultseal.RunSummary{}, log)
	case SealPersonal:
		return vaultseal.ManualSealFromConfig(cfg, &vaultseal.RunSummary{}, log)
	}
	r := vaultseal.NewSealResult()
	r.Error = "unknown seal kind '" + kind + "'"
	return r
}

// resync adapts vaultseal.ResyncFromConfig.
func (a *App) resync(cfg *storage.OrderedMap, log LogFn) *storage.OrderedMap {
	return vaultseal.ResyncFromConfig(cfg, log)
}

// publish adapts goodrepos.PublishFromConfig (the public directory).
func (a *App) publish(cfg *storage.OrderedMap, log LogFn) *goodrepos.PublishResult {
	return goodrepos.PublishFromConfig(cfg, &goodrepos.RunSummary{}, log)
}

// radar composes reporadar (the _repo_radar_job engine): the
// detection over the empty repo slots (READ-ONLY against GitHub),
// the apply to the LIVE config in place (the owner's typed word
// always won — apply never overwrites), and the report lines. The
// config's persistence stays with the shell's save door.
func (a *App) radar(cfg *storage.OrderedMap, log LogFn) *RadarResult {
	token := pyStrCfg(cfg, "github_token")
	api := reporadar.NewRadarAPI(token, reporadar.APITimeoutSec,
		reporadar.GITHUBAPI)
	dets := reporadar.RadarAllVaults(cfgPlainMap(cfg), api)
	results := map[string]any{}
	for k, v := range dets {
		if v != nil {
			results[k] = v.Map()
		}
	}
	applied := reporadar.ApplyDetections(cfg, results)
	lines := reporadar.ReportLines(results)
	return &RadarResult{Success: true, Applied: applied,
		Results: results, Lines: lines}
}

// ---------------------------------------------------------------------------
// The batch composition — batchwire.go's CLI-parity re-host (T-74).
// ---------------------------------------------------------------------------

// runBatch is the app's batch engine: the runImpl composition
// re-hosted (see batchwire.go's file doc for the full law).
func (a *App) runBatch(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
	return a.runBatchEngine(cfg, opts, ev)
}

// ---------------------------------------------------------------------------
// The startup passes' engines.
// ---------------------------------------------------------------------------

// sentinelPass is the ONE 60 s background pass's each run (Net · AI ·
// Proxy — connectioncheck's probes; the snapshot crosses to the chrome
// as the Sentinel event for the labels' paint, and the offline law's
// verdict transitions speak their lines; read-only, never raises).
//
// T-77's reality fix, found re-reading the pass: the old line
// dereferenced net.LatencyMs unconditionally — ProbeInternet leaves it
// nil on the offline path, so an offline pass panicked into the
// crash-isolation and the labels never painted. The latency now reads
// nil-safe ("n/a" when absent).
func (a *App) sentinelPass() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[app-lifecycle] sentinel pass crashed (isolated): %v\n", r)
		}
	}()
	cfg := a.loadConfig()
	cfgPlain := cfgPlainMap(cfg)
	probe := a.SentinelProbe
	if probe == nil {
		probe = connectioncheck.SentinelSnapshot
	}
	snap := probe(cfgPlain, 5.0)

	// The offline law (window.py's _apply_sentinel half): the verdict
	// transition decides the lines — warn-now is ONE modal at open (the
	// frontend's themed box rides the verdict word), fell-offline and
	// back-online are log lines only. The seat's read-transition-write
	// is ONE unit under the mutex (concurrent passes serialize exactly
	// like the reference's GUI-thread queued signal).
	a.sentinelMu.Lock()
	verdict := connectioncheck.OfflineVerdict(a.sentinelOnline, snap.Net.Online)
	online := snap.Net.Online
	a.sentinelOnline = &online
	a.sentinelMu.Unlock()
	switch verdict {
	case "warn-now":
		a.Bus.EmitLogLine("⚠️ No internet connection detected — GitCurator "+
			"needs the internet for syncing (fetching the "+
			"Telegram bot queue, GitHub repo data, backup vault "+
			"seals and the cloud LLM all require it). Local work "+
			"— vault scans, hygiene, local LLM — still works. "+
			"The 'Net' cue turns green the moment you're back "+
			"online; no restart needed.", "warning")
	case "fell-offline":
		a.Bus.EmitLogLine("⚠️ Net: fell offline — syncing will fail until the "+
			"connection returns (the 'Net' cue is red).", "warning")
	case "back-online":
		a.Bus.EmitLogLine("✅ Net: back online — syncing is available again.", "success")
	}

	latency := "n/a"
	if snap.Net.LatencyMs != nil {
		latency = fmt.Sprintf("%dms", *snap.Net.LatencyMs)
	}
	a.Bus.EmitLogLine(fmt.Sprintf("🛰️ Sentinel: net=%s latency=%s · AI=%s",
		netWord(snap.Net.Online), latency, snap.AI.Short), "info")

	// THE SNAPSHOT CROSSING (T-77): the chrome's labels paint from this
	// event (Net · AI · Proxy — the honest placeholders retire).
	a.Bus.EmitSentinel(Sentinel{
		Ts:      snap.Ts,
		Net:     snap.Net,
		AI:      snap.AI,
		Proxy:   snap.Proxy,
		Verdict: verdict,
	})
}

// netWord renders the probe's verdict.
func netWord(on bool) string {
	if on {
		return "online"
	}
	return "offline"
}

// llamaAutodetect is the once-per-session probe (v0.15.1): the
// configured URL + the running llama-server's ports + the common
// ports — a find speaks one line; the config's own auto-switch
// policy stays with the engine.
func (a *App) llamaAutodetect() {
	cfg := a.loadConfig()
	if strings.ToLower(strings.TrimSpace(pyStrCfg(cfg, "llm_provider"))) != "llamacpp" {
		return // the pass only wakes for the llamacpp provider
	}
	apiKey := pyStrCfg(cfg, "llamacpp_api_key")
	ports := llmclient.LlamaCppProcessPorts()
	base := llmclient.NormalizeLlamaCppAPIURL(pyStrCfg(cfg, "llamacpp_api_url"))
	if pr := llmclient.ProbeLlamaCpp(base, apiKey, 2.0); pr != nil && pr.Found {
		a.Bus.EmitLogLine("🦙 llama.cpp detected at "+base+" — the provider is live", "success")
		return
	}
	for _, p := range ports {
		cand := fmt.Sprintf("http://127.0.0.1:%d", p)
		if pr := llmclient.ProbeLlamaCpp(cand, apiKey, 2.0); pr != nil && pr.Found {
			a.Bus.EmitLogLine("🦙 llama.cpp detected at "+cand+" (the running server's own port)", "success")
			return
		}
	}
	a.Bus.EmitLogLine("🦙 llama.cpp autodetect: no live server found on the configured URL or the running ports", "info")
}

// startupHeal is the quiet both-vaults heal (v0.68.0) with the
// one-open cache; the story lands in the desk log.
func (a *App) startupHeal() {
	cfg := a.loadConfig()
	rc := notereads.New(cacheDBPathOf(cfg))
	defer func() {
		stats := rc.Close(false)
		if stats.Opens > 0 || stats.DbHits > 0 {
			a.Bus.EmitLogLine(fmt.Sprintf("⚡ The one-open law: %d note open(s) this pass — %d unchanged note(s) served from the last scan's memory", stats.Opens, stats.DbHits), "info")
		}
	}()
	// the config's own gate (vault_hygiene_on_start, default on —
	// the reference's _startup_vault_heal check).
	if v, ok := cfg.Get("vault_hygiene_on_start"); ok {
		if b, is := v.(bool); is && !b {
			return // the owner turned the quiet heal off
		}
	}
	healed := vaulthygiene.HealMachineVaults(cfgPlainMap(cfg),
		func(msg, level string) { a.Bus.EmitLogLine(msg, level) },
		func(label string, value, max int) { a.Bus.EmitJobProgress(label, value, max) },
		rc)
	if total := vaulthygiene.HygieneTotal(healed); total > 0 {
		a.Bus.EmitLogLine(fmt.Sprintf("🧹 Startup heal: %d item(s) healed across the machine vaults", total), "success")
	}
}

// backlogNotice is the review-backlog check (v0.42.0): scan the
// Websites vault's _review folder once per launch — one notice only
// when there is something to retry, never during a live batch.
func (a *App) backlogNotice() {
	if a.Commands != nil && a.Commands.batchLock.Busy() {
		return // never during a live batch
	}
	cfg := a.loadConfig()
	wv := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	if wv == "" || !fsIsDirPath(wv) {
		return
	}
	items := websitepipeline.ScanReviewBacklog(wv, func(msg, level string) {
		a.Bus.EmitLogLine(msg, level)
	}, func(string) bool { return false })
	if len(items) > 0 {
		a.Bus.EmitLogLine(fmt.Sprintf(
			"📌 The review backlog holds %d fetch-failed placeholder(s) in the "+
				"Websites vault's _review folder — the retry door can pick them up",
			len(items)), "info")
	}
}

// quietRadar is the quiet startup radar (v0.84.0): only when
// radar_pending says there is work — a fully-wired config never
// wakes it.
func (a *App) quietRadar() {
	cfg := a.loadConfig()
	cfgPlain := cfgPlainMap(cfg)
	if !reporadar.RadarPending(cfgPlain) {
		return
	}
	res := a.Commands.RadarDetect(context.Background())
	if !res.Success {
		return
	}
	for _, line := range res.Lines {
		a.Bus.EmitLogLine(line, "info")
	}
	if res.Applied > 0 {
		a.Bus.EmitLogLine(fmt.Sprintf(
			"📡 The quiet startup radar filled %d empty repo slot(s) — the detected names are remembered like credentials", res.Applied), "success")
	}
}
