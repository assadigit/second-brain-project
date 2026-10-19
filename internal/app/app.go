// app.go — T-71 slice (d): the DI root (doc 04 §2's App struct + §3's
// law: "app — the only package allowed to construct everything; no
// business logic"). New() wires the production seams over the real
// engine packages; every door stays injectable (the hermetic law —
// the contract tests in commands_test.go already proved the
// transport; a wiring smoke test proves the composition below).
//
// THE BATCH WIRING'S HONEST SCOPE (recorded per the mandate's own
// one-slice-properly law): the app's batch default is the direct
// engine composition — the pre-loop state (the manifest tracker, the
// vault index, the note-state DB, the cache), the runner with its
// events wired bus-ward (LLMDecision → the 600 s gate — the GUI's
// failure-modal path, NOT the CLI's headless warmup ladder), the
// GitHub loop, the closes, and the Finish chain. The CLI-parity
// shell wiring (the warmup family, the inputs fetchers, the websites
// phase, the Phase-5 CLEAR) remains the CLI's own composition and
// lands on the app's surface at T-72's binding (webui/ dissolving
// in) — until then the app's batch runs the runner's own laws, and
// the reference's GUI batch ran exactly this shape before its mixins
// grew.
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/config"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/goodrepos"
	"github.com/assadigit/second-brain-project/internal/llmclient"
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
// command surface, the lifecycle, and the injectable doors the
// engine compositions read.
type App struct {
	// ConfigPath is the live config file (the config store's home).
	ConfigPath string

	// Bus is the typed event bus (doc 04 §2).
	Bus *Bus
	// Commands is the command surface (doc 04 §5).
	Commands *Commands
	// Lifecycle owns the startup passes + the watchdog + shutdown.
	Lifecycle *Lifecycle

	// The engine doors (the hermetic law — New() wires the production
	// defaults; tests and staged shells may swap them BEFORE any
	// command runs).
	GHClientFactory func(token string) githubpipeline.RepoFetcher
	OllamaFactory   func(host string) llmclient.OllamaAPI
	BannerDownload  func(owner, repoName, folder string) string

	// appLock is the single-instance lock (nil until Acquire).
	appLock *AppLock
}

// New builds the whole composition over the config at configPath.
// headless mirrors the launch mode: no GUI will ever answer a gate.
func New(configPath string, headless bool) *App {
	a := &App{ConfigPath: configPath}
	a.Bus = NewBus()
	a.Bus.Headless = headless
	a.GHClientFactory = func(token string) githubpipeline.RepoFetcher {
		return githubpipeline.NewClient(token)
	}
	a.OllamaFactory = func(host string) llmclient.OllamaAPI {
		return llmclient.NewOllamaClient(host)
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
	return a
}

// loadConfig is the live-config door (a missing file reads as the
// empty config — the reference's load_config(path) or {}).
func (a *App) loadConfig() *storage.OrderedMap {
	cfg, err := config.Load(a.ConfigPath)
	if err != nil || cfg == nil {
		return storage.NewOrderedMap()
	}
	return cfg
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
// The batch composition (the runner over the pre-loop state).
// ---------------------------------------------------------------------------

// runBatch is the app's batch engine: the manifest tracker, the
// vault index, the note-state DB, the cache, the runner with its
// events wired bus-ward (LLMDecision → the gate when a GUI answers;
// the runner's own headless ladder otherwise), the GitHub loop, the
// closes, the Finish chain.
func (a *App) runBatch(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
	cfgPlain := cfgPlainMap(cfg)
	vaultPath := pyStrCfg(cfg, "vault_path")
	urls := opts.URLs
	total := len(urls)

	// The manifest tracker (the batch's own bookkeeping).
	tracker := batch.NewLinkTracker(vaultPath)
	source := opts.Source
	if source == "" {
		source = "direct"
	}
	tracker.SetSource(source)
	tracker.Intake(urls, opts.NonGithub)
	if ev.Log != nil {
		ev.Log(fmt.Sprintf("📋 Manifest created: %d GitHub + %d non-GitHub "+
			"links recorded", len(urls), len(opts.NonGithub)), "info")
	}

	// The vault index + the note-state DB + the cache.
	vi := batch.NewVaultIndex(vaultPath, nil)
	vi.Rebuild(nil)
	var noteStateDB *notestate.DB
	if vaultPath != "" {
		if nd, err := notestate.New(cacheDBPathOf(cfg)); err == nil {
			noteStateDB = nd
		}
	}
	cachePath := cacheDBPathOf(cfg)
	var cache *store.CacheDB
	if cc, err := store.NewCacheDB(cachePath); err == nil {
		cache = cc
	}

	// The dead + settled sets (the loop's own gates).
	deadURLs := map[string]bool{}
	settled := map[string]bool{}
	if cache != nil {
		deadURLs = cache.GetDeadURLSet(store.DeadLinkThreshold)
		settled = cache.GetSettledRepoSet()
	}
	dismissed := map[string]bool{}
	if noteStateDB != nil {
		if ds, err := noteStateDB.DismissedSet(notestate.VaultGithub); err == nil {
			dismissed = ds
		}
	}

	// The runner (the GUI shape: NOT headless — the LLM failure door
	// rides the bus's gate when a subscriber answers).
	r := githubpipeline.NewRunner(cfgPlain)
	r.Start()
	// The LLM collaborators resolved from the config (the GUI shape —
	// no CLI warmup ladder: a dead server surfaces through the
	// runner's own LLM failure door, exactly like the reference's
	// ProcessingWorker).
	r.LLMProvider = strings.ToLower(strings.TrimSpace(pyStrCfg(cfg, "llm_provider")))
	if r.LLMProvider == "" {
		r.LLMProvider = "ollama"
	}
	if r.LLMProvider == "ollama" {
		base := strings.TrimRight(cfgNestedStr(cfg, "ollama", "base_url",
			"http://localhost:11434"), "/")
		if a.OllamaFactory != nil {
			r.OllamaClient = a.OllamaFactory(base)
		} else {
			r.OllamaClient = llmclient.NewOllamaClient(base)
		}
		r.Model = cfgNestedStr(cfg, "ollama", "model", "")
	}
	r.Cache = cache
	r.NoteStateDB = noteStateDB
	r.VaultIndex = vi
	r.LinkTracker = tracker
	r.Total = total
	r.NonGithubURLs = opts.NonGithub
	r.IntakeDuplicates = opts.IntakeDuplicates
	r.RawURLCount = opts.RawURLCount
	r.BotSource = source == "bot"
	r.NewGHClient = a.GHClientFactory
	if a.BannerDownload != nil {
		r.BannerFn = a.BannerDownload
	}
	r.Log = ev.Log
	r.OnStatus = ev.Status
	r.OnProgress = ev.Progress
	r.OnDiskFull = ev.DiskFull
	r.OnModelChanged = ev.ModelChange
	if ev.LLMDecision != nil {
		r.LLMDecisionFn = func(repoName string, attempt int, err error,
			client llmclient.OllamaAPI, model string) string {
			return ev.LLMDecision(repoName)
		}
	}
	if ev.ModelPrompt != nil {
		r.ModelPromptFn = ev.ModelPrompt
	}
	fin := &BatchResult{}
	r.Finished = func(ok bool, msg string) {
		fin.Success = ok
		fin.Message = msg
		fin.Processed = r.Processed
		fin.Total = r.Total
	}

	r.RunGitHubLoop(urls, deadURLs, dismissed, settled, store.DeadLinkThreshold)

	if cache != nil {
		_ = cache.Close()
	}
	if noteStateDB != nil {
		func() {
			defer func() { recover() }()
			_ = noteStateDB.Close()
		}()
	}
	r.Finish()
	if fin.Message == "" {
		fin.Success = false
		fin.Message = "the batch engine finished without a verdict"
	}
	return fin
}

// ---------------------------------------------------------------------------
// The startup passes' engines.
// ---------------------------------------------------------------------------

// sentinelPass is the ONE 60 s background pass's each run (Net · AI ·
// Proxy — connectioncheck's probes; the results land on the bus for
// the shell's status row; read-only, never raises).
func (a *App) sentinelPass() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[app-lifecycle] sentinel pass crashed (isolated): %v\n", r)
		}
	}()
	cfg := a.loadConfig()
	cfgPlain := cfgPlainMap(cfg)
	net := connectioncheck.ProbeInternet(cfgPlain, 5.0)
	aiState := connectioncheck.AIStatus(cfgPlain, 5.0)
	line := fmt.Sprintf("🛰️ Sentinel: net=%s latency=%dms · AI=%s",
		netWord(net.Online), *net.LatencyMs, aiState.Short)
	a.Bus.EmitLogLine(line, "info")
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
