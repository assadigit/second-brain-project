package cli

// engine.go — T-62 slice 2: run_batch_visual (cli_run.py 239-504) as
// the synchronous Go shell over the ported engines — no Qt anywhere.
//
// The composition mirrors the reference's two-layer shape:
//   - run_batch_visual (the CLI layer): the StatusLine, the signal
//     wiring, the SIGINT handler + the 100 ms heartbeat ticker, the
//     _finish handler (the seals in the reference's exact order + THE
//     PHASE-5 CLEAR + the dry-run report), and the exit-code law
//     (0 iff the batch's success flag).
//   - ProcessingWorker.run + _run_impl (the engine layer): the dry-run
//     switch flip/restore + the crash wrapper (finished exactly once),
//     the mode dispatch, the setup family (the LinkTracker intake with
//     the banned/self drops + the inbox tables, the cache open with the
//     dry-run shadow, note-state, the dead set, THE REPOS BANISHMENT
//     pass, the settled ledger, the vault index rebuild + SetVaultIndex,
//     the missing-note backfill, the §4.4 start-of-run pass, the
//     dismissed set, the batch snapshot), the warmup family (the Ollama
//     list/timeout + autostart + warmup + the auto-switch/stand-in
//     ladder over the model_prompt_callback bridge, the llama.cpp
//     re-probe, the cloud /models pre-flight), the GitHub loop over
//     githubpipeline.Runner, the websites phase over
//     websitepipeline.Pipeline (website_phase.py's wiring), and the
//     ported Finish chain.
//
// THE CONCURRENCY TWIN (Qt → Go): the reference runs the worker on a
// QThread while the main thread spins app.exec() and a 100 ms QTimer
// drives StatusLine.tick (the periodic return-to-Python that lets the
// SIGINT handler run). The Go twin runs the batch inline on the
// caller's goroutine and the heartbeat ticker + the signal drain on
// their own goroutines; the StatusLine's mutex serializes the paint
// (Qt's cross-thread signal marshaling, made explicit). A SIGINT logs
// the interrupt line and calls runner.Stop() — the loop breaks at the
// next poll and the batch still finishes through the normal chain
// (the reference's worker.stop sets is_running=False; the same law).
//
// Option-B laws (ADR-0004 Am.3): the code_requested signal never fires
// (no MTProto client exists — no Telegram login is ever prompted);
// _mark_bot_queue_read fell away (mirror-push is the read ack — the
// Phase-5 CLEAR re-spec, SESSION-029); the `telegram` fetch mode is
// unreachable (the flag surface answers refusal cards before the
// engine; an unknown mode reads as the missing-credentials shape the
// reference answers with no Telethon keys configured).

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/goodrepos"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/platformintake"
	"github.com/assadigit/second-brain-project/internal/repobanish"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/taxonomy"
	"github.com/assadigit/second-brain-project/internal/vaulthygiene"
	"github.com/assadigit/second-brain-project/internal/vaultseal"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// StatusPainter is the StatusLine door (the engine + the picker pause
// the live line through it; the golden replay injects a recorder so the
// log sequence is byte-stable without the \r repaints).
type StatusPainter interface {
	Start(label string, total int)
	Stop(final string)
	SetProgress(current, total int)
	SetRepo(repo string)
	Tick()
	Log(text, level string)
	EraseIfActive()
}

// RunBatchVisual mirrors run_batch_visual: run one batch with colors +
// spinner + progress bar; returns the process exit code (0 iff the
// batch's success flag).
type RunBatchOptions struct {
	Mode             string
	URLs             []string
	ImportFile       string
	BotSource        bool
	NonGithub        []string
	IntakeDuplicates int
	RawURLCount      int
	VaultArg         string
	DryRun           bool
}

func (c *CLI) RunBatchVisual(cfg *storage.OrderedMap,
	o RunBatchOptions) int {

	if o.VaultArg != "" {
		cfg.Set("vault_path", o.VaultArg)
	}

	status := c.statusLineDoor()
	e := &batchEngine{
		c:                c,
		cfg:              cfg,
		mode:             o.Mode,
		urls:             o.URLs,
		importFile:       o.ImportFile,
		botSource:        o.BotSource,
		nonGithub:        o.NonGithub,
		intakeDuplicates: o.IntakeDuplicates,
		rawURLCount:      o.RawURLCount,
		dryRun:           o.DryRun,
		status:           status,
		startedAt:        c.Now(),
	}

	// --- Ctrl+C: graceful stop (the runner sets its stop flag; the
	// batch still lands on the normal finish chain). ---
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)
	sigDone := make(chan struct{})
	go func() {
		select {
		case <-sigCh:
			status.Log("Interrupt received — stopping batch…", "warning")
			func() {
				defer func() { recover() }()
				if e.runner != nil {
					e.runner.Stop()
				}
			}()
		case <-sigDone:
		}
	}()

	// --- The heartbeat: a 100 ms ticker driving StatusLine.Tick (the
	// spinner under a real loop — T-61's leg; the SIGINT drain rides
	// its own goroutine natively in Go). ---
	ticker := time.NewTicker(100 * time.Millisecond)
	hbDone := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				status.Tick()
			case <-hbDone:
				return
			}
		}
	}()

	status.Start("Processing batch", 1)
	code := e.run()
	ticker.Stop()
	close(hbDone)
	close(sigDone)
	signal.Stop(sigCh)
	status.Stop("")
	// app.exec()'s own return code is always 0 after quit() — the
	// batch's success flag is what scripts and the launcher need.
	return code
}

// batchEngine carries the worker-side state of one batch.
type batchEngine struct {
	c          *CLI
	cfg        *storage.OrderedMap
	mode       string
	urls       []string
	urlsSet    bool
	importFile string

	botSource        bool
	nonGithub        []string
	intakeDuplicates int
	rawURLCount      int
	dryRun           bool

	status    StatusPainter
	runner    *githubpipeline.Runner
	startedAt time.Time

	// The outcome (written by finish; drives the exit code).
	success bool

	// The review-backlog modes' state (v0.42/v0.51).
	reviewBacklogItems []websitepipeline.BacklogItem
	masterRetryMode    bool
	nonGithubSet       bool

	// The model the warmup resolved (carries to the runner).
	warmModel    string
	warmOllama   llmclient.OllamaAPI
	warmProvider string
	vaultIndex   *batch.VaultIndex
	noteStateRun map[string]*notestate.RunResult
}

// run mirrors ProcessingWorker.run (199-237): the dry-run switch
// flip/restore + the crash wrapper (the finished signal fires exactly
// once even on a crash).
func (e *batchEngine) run() int {
	if e.dryRun {
		dryrun.Enable()
		e.log("🧪 DRY-RUN: every vault write this batch would perform is "+
			"logged and skipped. Nothing will be written, sealed or "+
			"marked read.", "warning")
	}
	func() {
		defer func() {
			if e.dryRun {
				dryrun.Disable()
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				// 💥 Batch crashed: {type}: {e} — the exception SHAPE is
				// the contract (the reference's traceback tail has no Go
				// twin; the class token + message are recorded here).
				func() {
					defer func() { recover() }()
					e.log(fmt.Sprintf("💥 Batch crashed: %s", pyCrashText(r)), "error")
				}()
				func() {
					defer func() { recover() }()
					e.finish(false, fmt.Sprintf("Batch crashed: %s", pyCrashText(r)))
				}()
			}
		}()
		e.runImpl()
	}()
	if e.success {
		return 0
	}
	return 1
}

// log routes a worker log line to the status line (the
// log_message → status.log wiring).
func (e *batchEngine) log(msg, level string) {
	e.status.Log(msg, level)
}

// finish mirrors run_batch_visual's _finish handler: the outcome, the
// bottom-line stop, the BATCH card, the seals (the reference's exact
// order), THE PHASE-5 CLEAR, and the dry-run path.
func (e *batchEngine) finish(success bool, message string) {
	e.success = success
	e.status.Stop("")
	elapsed := int(e.c.Now().Sub(e.startedAt).Seconds())
	if elapsed < 1 {
		elapsed = 1 // max(1, int(...))
	}
	fmt.Fprint(e.c.Out, "\n")
	fmt.Fprint(e.c.Out, Rule()+"\n")
	if success {
		fmt.Fprint(e.c.Out, Paint("  ✔ BATCH DONE ", cGreen, cBold)+
			Paint(fmt.Sprintf("in %dm %02ds", elapsed/60, elapsed%60), cDim)+"\n")
		fmt.Fprint(e.c.Out, Paint("  "+message, cGreen)+"\n")
	} else {
		fmt.Fprint(e.c.Out, Paint("  ✖ BATCH FAILED ", cRed, cBold)+
			Paint(fmt.Sprintf("in %dm %02ds", elapsed/60, elapsed%60), cDim)+"\n")
		fmt.Fprint(e.c.Out, Paint("  "+message, cRed)+"\n")
	}
	fmt.Fprint(e.c.Out, Rule()+"\n")

	if e.dryRun {
		// A dry-run batch is side-effect free EVERYWHERE: no git seal,
		// no Good Repos publish, no mirror push; the withheld-writes
		// report is saved and the run stops.
		n := dryrun.EntryCount()
		e.status.Log(fmt.Sprintf(
			"🧪 DRY-RUN COMPLETE — %d vault operation(s) were logged, "+
				"not performed. VaultSeal, Good Repos publish, bot-queue "+
				"mark-read and last_processed_msg_id were all skipped.",
			n), "warning")
		if reportPath := e.c.writeDryrunReport(); reportPath != "" {
			e.status.Log("🧪 Dry-run report: "+reportPath, "info")
		}
		fmt.Fprint(e.c.Out, "\n")
		return
	}

	// The finish seals — best-effort (they run for failed batches too:
	// notes saved before a mid-run failure are exactly what we want
	// backed up).
	processed, total := e.processedTotal()
	e.sealVault(processed, total)
	e.publishGoodRepos(processed, total)
	e.sealWebsites(processed, total)
	e.sealManual(processed, total)

	// THE PHASE-5 CLEAR (the re-spec: the mirror push is the ack; the
	// all-clear verdict rides the reference's success AND bot_source
	// gate).
	var tracker *batch.LinkTracker
	var cache *store.CacheDB
	if e.runner != nil {
		tracker = e.runner.LinkTracker
		cache = e.runner.Cache
	}
	e.c.PhaseFiveClear(e.cfg, e.botSource, success, tracker,
		e.vaultIndex, cache)

	fmt.Fprint(e.c.Out, "\n")
}

// processedTotal reads the runner's counters (getattr(worker,
// "processed", 0) — nil-safe).
func (e *batchEngine) processedTotal() (int, int) {
	if e.runner == nil {
		return 0, 0
	}
	return e.runner.Processed, e.runner.Total
}

// sealVault mirrors the VaultSeal leg.
func (e *batchEngine) sealVault(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.status.Log(fmt.Sprintf("VaultSeal error: %s", pyCrashText(r)), "warning")
		}
	}()
	e.status.Log("Sealing vault (git commit + push if configured)…", "info")
	summary := &vaultseal.RunSummary{Processed: &processed, Total: &total}
	vs := vaultseal.SealFromConfig(e.cfg, summary, e.silentSealLog())
	level := "warning"
	if vs.OK() {
		level = "success"
	}
	e.status.Log("VaultSeal: "+vs.Describe(), level)
}

// publishGoodRepos mirrors the Good Repos publish leg.
func (e *batchEngine) publishGoodRepos(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.status.Log(fmt.Sprintf("Good Repos error: %s", pyCrashText(r)), "warning")
		}
	}()
	summary := &goodrepos.RunSummary{Processed: &processed, Total: &total}
	gr := goodrepos.PublishFromConfig(e.cfg, summary, e.silentSealLog())
	level := "warning"
	if gr.OK() {
		level = "success"
	}
	e.status.Log("Good Repos: "+gr.Describe(), level)
}

// sealWebsites mirrors the WEBSITES vault's own private mirror —
// silent no-op while the websites pipeline is OFF.
func (e *batchEngine) sealWebsites(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.status.Log(fmt.Sprintf("Websites vault seal error: %s", pyCrashText(r)), "warning")
		}
	}()
	summary := &vaultseal.RunSummary{Processed: &processed, Total: &total}
	ws := vaultseal.WebsitesSealFromConfig(e.cfg, summary, e.silentSealLog())
	if ws.Sealed || ws.Error != "" {
		level := "warning"
		if ws.OK() {
			level = "success"
		}
		e.status.Log("Websites vault seal: "+ws.Describe(), level)
	}
}

// sealManual mirrors the PERSONAL vault's own private mirror — silent
// no-op while either the vault path or the repo name is empty.
func (e *batchEngine) sealManual(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.status.Log(fmt.Sprintf("Personal vault seal error: %s", pyCrashText(r)), "warning")
		}
	}()
	summary := &vaultseal.RunSummary{Processed: &processed, Total: &total}
	mn := vaultseal.ManualSealFromConfig(e.cfg, summary, e.silentSealLog())
	if mn.Sealed || mn.Error != "" {
		level := "warning"
		if mn.OK() {
			level = "success"
		}
		e.status.Log("Personal vault seal: "+mn.Describe(), level)
	}
}

// silentSealLog is the sealer's internal log (the reference passes
// log=None to the config bridges — the CLI speaks the describe lines).
func (e *batchEngine) silentSealLog() func(string, string) {
	return func(string, string) {}
}

// onModelChanged mirrors the model_changed handler: the runner's
// in-place mutation is mirrored onto the engine's cfg (the reference's
// ONE shared dict — the Go runner holds a plain-map copy, so the write
// is synced here before the save), the switch line, and the config
// save (the v0.07.2 __-key strip law).
func (e *batchEngine) onModelChanged(provider, model string) {
	switch provider {
	case "cloud":
		e.cfg.Set("cloud_model", model)
	case "llamacpp":
		e.cfg.Set("llamacpp_model", model)
	default:
		if om := getMap(e.cfg, "ollama"); om != nil {
			om.Set("model", model)
		} else {
			om := storage.NewOrderedMap()
			om.Set("model", model)
			e.cfg.Set("ollama", om)
		}
	}
	e.status.Log(fmt.Sprintf(
		"LLM model switched to '%s' (%s) — saving config.", model, provider), "info")
	configPath := ""
	if v, ok := e.cfg.Get("__config_path__"); ok {
		if s, isStr := v.(string); isStr {
			configPath = s
		}
	}
	if configPath == "" {
		return
	}
	if err := storage.WriteConfigFile(configPath, cleanConfigForSave(e.cfg)); err != nil {
		e.status.Log("Could not save config: "+pyErrString(err), "warning")
	}
}

// applyModelChoiceInPlace mirrors WorkerLlmMixin._apply_model_choice:
// mutate the config IN PLACE (never rebind) + emit model_changed.
func (e *batchEngine) applyModelChoiceInPlace(newModel string) {
	if newModel == "" {
		return
	}
	provider := cfgOrStr(e.cfg, "llm_provider")
	if provider == "" {
		provider = "ollama"
	}
	switch provider {
	case "cloud":
		e.cfg.Set("cloud_model", newModel)
	case "llamacpp":
		e.cfg.Set("llamacpp_model", newModel)
	default:
		// Mutate the nested dict in place.
		var ollamaCfg *storage.OrderedMap
		if v, ok := e.cfg.Get("ollama"); ok {
			if om, is := v.(*storage.OrderedMap); is {
				ollamaCfg = om
			}
		}
		if ollamaCfg == nil {
			ollamaCfg = storage.NewOrderedMap()
			e.cfg.Set("ollama", ollamaCfg)
		}
		ollamaCfg.Set("model", newModel)
	}
	func() {
		defer func() { recover() }()
		e.onModelChanged(provider, newModel)
	}()
}

// workerModelPrompt mirrors _worker_model_prompt (the CLI's
// model_prompt_callback bridge): the numbered menu + the persist.
func (e *batchEngine) workerModelPrompt(configured string,
	available []string) string {

	choice := e.c.PromptModelMenu(configured, available, e.status, true)
	if choice != "" {
		configPath := ""
		if v, ok := e.cfg.Get("__config_path__"); ok {
			if s, isStr := v.(string); isStr {
				configPath = s
			}
		}
		e.c.ApplyModelChoice(e.cfg, choice, configPath)
	}
	return choice
}

// ---------------------------------------------------------------------------
// _run_impl — the engine body (processing_worker.py 239-1889)
// ---------------------------------------------------------------------------

func (e *batchEngine) runImpl() {
	cfgPlain := cfgPlainMap(e.cfg)

	// v0.10/v0.11 — the pipeline switches, checked BEFORE anything is
	// fetched so an OFF switch consumes nothing.
	pipes := getMap(e.cfg, "pipelines")
	githubOn := pyTruthy(getOr(pipes, "github", true))
	websitesOn := pyTruthy(getOr(pipes, "websites", false))
	if !githubOn && !websitesOn {
		e.log("⛔ Both pipelines are switched OFF (Settings → 📁 Vault). "+
			"Nothing was fetched or processed.", "warning")
		e.finish(true, "Both pipelines are off — nothing to do.")
		return
	}

	// The mode dispatch.
	var urls []string
	switch e.mode {
	case "direct":
		urls = e.urls
	case "import":
		urls = e.fetchFromImport()
	case "review_retry":
		// v0.42.0 — the _review backlog retry.
		if !websitesOn {
			e.log("⛔ Websites pipeline is OFF (Settings → 📁 Vault) — "+
				"the _review backlog cannot be retried.", "warning")
			e.finish(true, "Websites pipeline is off — nothing retried.")
			return
		}
		wv := strings.TrimSpace(cfgOrStr(e.cfg, "website_vault_path"))
		if wv == "" || !e.c.IsDir(wv) {
			e.log("⚠️ No Websites vault is set (Settings → 📁 Vault) — "+
				"the _review backlog was not scanned.", "warning")
			e.finish(true, "No Websites vault — nothing retried.")
			return
		}
		scanned := websitepipeline.ScanReviewBacklog(wv, e.log, nil)
		if len(scanned) == 0 {
			e.log("✅ The _review backlog is empty — no app-owned "+
				"fetch-failed placeholders are waiting (anything else in "+
				"_review needs your eyes, not a retry).", "success")
			e.finish(true, "The _review backlog is empty — nothing to retry.")
			return
		}
		e.reviewBacklogItems = scanned
		e.setNonGithub(backlogURLs(scanned))
	case "master_retry":
		// v0.51.0 — the master table's " - " rows.
		if !websitesOn {
			e.log("⛔ Websites pipeline is OFF (Settings → 📁 Vault) — "+
				"the master table's ' - ' rows cannot be retried.", "warning")
			e.finish(true, "Websites pipeline is off — nothing retried.")
			return
		}
		wv := strings.TrimSpace(cfgOrStr(e.cfg, "website_vault_path"))
		if wv == "" || !e.c.IsDir(wv) {
			e.log("⚠️ No Websites vault is set (Settings → 📁 Vault) — "+
				"the master table was not checked.", "warning")
			e.finish(true, "No Websites vault — nothing retried.")
			return
		}
		var scanned []websitepipeline.WaitingRow
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Master-table waiting scan fell back to a "+
						"pure file read: %s", pyCrashText(r)), "warning")
					scanned = websitepipeline.ScanMasterWaitingRows(wv, nil, e.log)
				}
			}()
			var state *websitepipeline.WebsiteStateDB
			if dryrun.IsEnabled() {
				state, _ = websitepipeline.NewWebsiteStateDB(
					dryrun.ShadowCachePath(cacheDBPath()))
			} else {
				state, _ = websitepipeline.NewWebsiteStateDB("cache.db")
			}
			if state != nil {
				defer state.Close()
				scanned = websitepipeline.ScanMasterWaitingRows(wv, state, e.log)
			} else {
				scanned = websitepipeline.ScanMasterWaitingRows(wv, nil, e.log)
			}
		}()
		var waiting, eyes []websitepipeline.WaitingRow
		for _, it := range scanned {
			if it.Kind == "fetch" {
				waiting = append(waiting, it)
			} else if it.Kind == "eyes" {
				eyes = append(eyes, it)
			}
		}
		if len(waiting) == 0 {
			msg := "✅ The master table has no ' - ' row waiting for a retry — " +
				"everything is up to date"
			if len(eyes) > 0 {
				msg += fmt.Sprintf(" (%d row(s) wait for your eyes: set ✅ or 🪦 "+
					"in the table)", len(eyes))
			}
			e.log(msg, "success")
			e.finish(true, "Master table caught up — nothing to retry.")
			return
		}
		e.log(fmt.Sprintf("🔁 Retrying the master table's ' - ' rows: %d "+
			"link(s) — valid links whose fetches failed, no verdict set; "+
			"anything still walled keeps waiting", len(waiting)), "info")
		e.masterRetryMode = true
		var items []websitepipeline.BacklogItem
		for _, w := range waiting {
			items = append(items, websitepipeline.BacklogItem{URL: w.URL, Path: w.Path})
		}
		e.reviewBacklogItems = items
		e.setNonGithub(waitingURLs(waiting))
	default:
		// The Telethon fetch — Option B's accepted loss (ADR-0004
		// Am.3): the flag surface never routes here; the missing-keys
		// shape is the reference's own no-credentials answer.
		e.log("Telegram credentials missing.", "error")
		urls = nil
	}

	// v0.11.0 — with the GitHub pipeline off, GitHub links are logged
	// and skipped — never processed.
	if !githubOn && len(urls) > 0 {
		e.log(fmt.Sprintf("⛔ GitHub pipeline is OFF — %d GitHub link(s) "+
			"skipped this run.", len(urls)), "warning")
		urls = nil
	}

	websiteLinks := e.nonGithub
	if len(urls) == 0 && !(websitesOn && len(websiteLinks) > 0) {
		e.log("No URLs found to process.", "warning")
		e.finish(true, "No URLs found.")
		return
	}

	total := len(urls)
	if websitesOn {
		total += len(websiteLinks)
	}

	// v23 — Phase 1: INTAKE — record ALL links BEFORE any processing.
	vaultPath := cfgOrStr(e.cfg, "vault_path")
	var tracker *batch.LinkTracker
	if vaultPath != "" {
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Link tracker init failed (continuing "+
						"without manifest): %s", pyCrashText(r)), "warning")
					tracker = nil
				}
			}()
			tracker = batch.NewLinkTracker(vaultPath)
			source := "telegram"
			switch e.mode {
			case "direct":
				if e.botSource {
					source = "bot"
				} else {
					source = "import"
				}
			case "import":
				source = "import"
			}
			tracker.SetSource(source)
			nonGithub := e.nonGithub
			tracker.Intake(urls, nonGithub)

			// Non-GitHub links with the Websites pipeline ON are the
			// phase's input (logged); with it OFF they land in the
			// per-platform _inbox tables (banned never collected).
			lawBlocked := links.BlockedDomainsFromConfig(cfgPlain)
			if len(nonGithub) > 0 && !websitesOn {
				recordable := []string{}
				for _, ngURL := range nonGithub {
					if links.DomainIsBlocked(ngURL, lawBlocked) {
						tracker.MarkBlocked(ngURL, "blocked domain")
					} else {
						recordable = append(recordable, ngURL)
						tracker.MarkRecorded(ngURL)
					}
				}
				if len(recordable) > 0 {
					src := "Import"
					if e.botSource {
						src = "Bot"
					}
					e.createInboxNotes(recordable, src)
				}
			} else if len(nonGithub) > 0 {
				e.log(fmt.Sprintf("🌐 %d non-GitHub link(s) queued for the "+
					"Websites pipeline", len(nonGithub)), "info")
			}
			e.log(fmt.Sprintf("📋 Manifest created: %d GitHub + %d non-GitHub "+
				"links recorded", len(urls), len(nonGithub)), "info")
		}()
	}

	// --- The warmup family (the provider ladder). A dead server
	// aborts the batch (the early-return law: the finished signal
	// already fired inside). ---
	if e.warmupFamily() {
		return
	}

	// v30 — the cache opens AFTER the Ollama early-return so the
	// "Ollama not available" exit can no longer leak the handle. A
	// dry-run batch reads the REAL cache through a throwaway copy.
	cache, cerr := store.NewCacheDB(func() string {
		if dryrun.IsEnabled() {
			return dryrun.ShadowCachePath(cacheDBPath())
		}
		return cacheDBPath()
	}())
	if cerr != nil {
		// The reference's constructor raises → the crash wrapper owns
		// the card; mirror its exact speech.
		panic(fmt.Errorf("CacheDB: %s", pyErrString(cerr)))
	}

	// v0.10.0 — one note-state connection for the whole batch (None in
	// a dry-run — a rehearsal records no bookkeeping).
	var noteStateDB *notestate.DB
	if !dryrun.IsEnabled() {
		if ns, err := notestate.New("cache.db"); err == nil {
			noteStateDB = ns
		}
	}

	// v0.08 — the CONFIRMED-dead set, loaded ONCE.
	deadThreshold := store.DeadLinkThresholdFromConfig(cfgPlain)
	deadURLs := cache.GetDeadURLSet(deadThreshold)

	// v0.65.0 — THE REPOS BANISHMENT, the batch's opening gate.
	if githubOn {
		rvault := strings.TrimSpace(cfgOrStr(e.cfg, "vault_path"))
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Repos banishment pass skipped: %s",
						pyCrashText(r)), "warning")
				}
			}()
			repobanish.RunRepoBanishmentPass(cache, noteStateDB, rvault,
				e.log, cfgPlain, e.repoBanishFactory)
		}()
	}

	// v0.63.3 — THE SETTLED REPOS LEDGER, the run gate.
	settled := map[string]bool{}
	func() {
		defer func() { recover() }()
		settled = cache.GetSettledRepoSet()
	}()

	// The vault index for URL-based dedup (ground truth).
	if vaultPath != "" {
		e.vaultIndex = batch.NewVaultIndex(vaultPath, nil)
		e.vaultIndex.Rebuild(e.log)
		if tracker != nil {
			func() {
				defer func() { recover() }()
				tracker.SetVaultIndex(e.vaultIndex)
			}()
		}
	}

	// v0.20.0 — the missing-repo backfill.
	if githubOn && vaultPath != "" {
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Missing-repo backfill skipped: %s",
						pyCrashText(r)), "warning")
				}
			}()
			r := e.buildRunner(cache, noteStateDB, tracker, total, urls)
			r.BackfillMissingNotes(cache, vaultPath, deadThreshold)
		}()
	}

	// v0.12.0 — the FULL start-of-run §4.4 pass for BOTH vaults.
	e.noteStateRun = map[string]*notestate.RunResult{}
	if githubOn && vaultPath != "" {
		e.noteStateRun["github"] = e.runStartCheckGithub(vaultPath, noteStateDB)
	}
	if websitesOn {
		if wv := strings.TrimSpace(cfgOrStr(e.cfg, "website_vault_path")); wv != "" {
			e.noteStateRun["websites"] = e.runStartCheckWebsites(wv, noteStateDB)
		}
	}

	// v0.12.0 — the dismissed list for the GitHub loop.
	dismissed := map[string]bool{}
	if noteStateDB != nil {
		func() {
			defer func() { recover() }()
			if ds, err := noteStateDB.DismissedSet(notestate.VaultGithub); err == nil {
				dismissed = ds
			}
		}()
	}

	// v22 Feature 6 — the pre-batch snapshot.
	snapshot := githubpipeline.SnapshotBatchFiles(vaultPath)

	runner := e.buildRunner(cache, noteStateDB, tracker, total, urls)
	e.runner = runner
	runner.SetBatchFiles(snapshot)
	runner.NoteStateRun = e.noteStateRun

	// The GitHub loop.
	runner.RunGitHubLoop(urls, deadURLs, dismissed, settled, deadThreshold)

	// v0.11.0 — the Websites pipeline runs AFTER the GitHub loop.
	e.runWebsitePhase(cache, noteStateDB, e.warmOllama, e.warmModel)

	// The closes.
	_ = cache.Close()
	if noteStateDB != nil {
		func() {
			defer func() { recover() }()
			_ = noteStateDB.Close()
		}()
	}

	// The ported Finish chain (the undo diff, the verification, the
	// reports, the batch summary, the finished signal).
	runner.Finish()
}

// repoBanishFactory mirrors _repo_banish_factory: the channel is
// built ONLY when the scan found marks — a clean vault never pays the
// Worker handshake. The Go shape returns the Worker-mediated round-trip
// over repobanish's own channel type; (nil, nil) is the reference's
// make_telegram_confirm() → None (the pass defers, marks stay).
func (e *batchEngine) repoBanishFactory() (repobanish.RepoBanishChannelFn, error) {
	s := workerclient.NewSync(cfgPlainMap(e.cfg))
	if !s.IsEnabled() {
		return nil, nil
	}
	return func(items []repobanish.PendingRepoBanishment,
		log repobanish.LogFn) *repobanish.RepoBanishConfirmResult {
		if log == nil {
			log = func(string, string) {}
		}
		timeoutS := pyFloatCfgOr(e.cfg, "banish_confirm_timeout_s", 300)
		if timeoutS == 0 {
			timeoutS = 300
		}
		payload := make([]any, 0, len(items))
		for _, it := range items {
			row := storage.NewOrderedMap()
			row.Set("url", it.URL)
			row.Set("title", it.Title)
			row.Set("marker", it.Marker)
			row.Set("door", it.Door)
			payload = append(payload, row)
		}
		log(fmt.Sprintf("📨 Deletion review sent to Telegram — %d note(s) marked "+
			"for deletion, waiting for your answer (buttons in the bot chat; up "+
			"to %ds)", len(payload), int(timeoutS)), "info")
		confID, why := s.ProposeBanish(len(payload), payload, timeoutS)
		if confID == nil {
			log(fmt.Sprintf("⚠️ The deletion review could not reach Telegram (%s) — "+
				"nothing deleted, the marks stay; I'll ask again next run", why), "warning")
			return &repobanish.RepoBanishConfirmResult{Verdict: "defer"}
		}
		verdict := "timeout"
		deadline := time.Now().Add(time.Duration(
			maxFloat(timeoutS, banishPollIntervalS) * float64(time.Second)))
		for time.Now().Before(deadline) {
			time.Sleep(time.Duration(banishPollIntervalS * float64(time.Second)))
			status := strings.TrimSpace(strings.ToLower(s.BanishStatus(*confID)))
			if status == "" {
				continue
			}
			switch status {
			case "confirmed", "declined":
				verdict = status
			case "superseded", "timeout", "done", "stale":
				verdict = "timeout"
			default:
				continue
			}
			break
		}
		if verdict == "timeout" {
			func() {
				defer func() { recover() }()
				s.ReportBanishResult(*confID, "timeout", 0)
			}()
		}
		out := &repobanish.RepoBanishConfirmResult{Verdict: verdict}
		if verdict == "confirmed" {
			id := *confID
			out.Report = func(deleted int) {
				defer func() { recover() }()
				s.ReportBanishResult(id, "deleted", deleted)
			}
		}
		return out
	}, nil
}

// setNonGithub stores the backlog modes' non-GitHub list.
func (e *batchEngine) setNonGithub(urls []string) {
	e.nonGithub = urls
	e.nonGithubSet = true
}

// buildRunner constructs the Runner with the engine's event wiring.
func (e *batchEngine) buildRunner(cache *store.CacheDB,
	noteStateDB *notestate.DB, tracker *batch.LinkTracker,
	total int, urls []string) *githubpipeline.Runner {

	r := githubpipeline.NewRunner(cfgPlainMap(e.cfg))
	r.Start()         // is_running = True (the __init__ law)
	r.Clock = e.c.Now // the reports' stamps ride the CLI's clock seam
	r.Headless = true // the CLI: non-blocking defaults everywhere
	r.Cache = cache
	r.NoteStateDB = noteStateDB
	r.VaultIndex = e.vaultIndex
	r.LinkTracker = tracker
	r.Total = total
	r.NonGithubURLs = e.nonGithub
	r.IntakeDuplicates = e.intakeDuplicates
	r.RawURLCount = e.rawURLCount
	r.BotSource = e.botSource
	r.LLMProvider = e.warmProvider
	r.OllamaClient = e.warmOllama
	r.Model = e.warmModel
	if e.c.GHClientFactory != nil {
		r.NewGHClient = e.c.GHClientFactory
	}
	if e.c.BannerDownload != nil {
		r.BannerFn = e.c.BannerDownload
	}
	r.ModelPromptFn = e.workerModelPrompt
	r.Log = e.log
	r.OnStatus = func(url string) {
		repo := url
		if i := strings.LastIndex(url, "/"); i >= 0 {
			repo = url[i+1:]
		}
		if repo == "" {
			repo = url
		}
		e.status.SetRepo(repo)
	}
	r.OnProgress = func(pos, tot int) {
		e.status.SetProgress(pos, tot)
		e.status.Tick()
	}
	r.OnDiskFull = func(p string) {
		e.status.Log(fmt.Sprintf(
			"Disk full at %s — repo skipped and recorded in retry queue.", p), "error")
	}
	r.OnModelChanged = e.onModelChanged
	r.Finished = e.finish
	r.HygienePass = e.hygienePass
	return r
}

// hygienePass wires the vaulthygiene module (T-46): the batch's own
// closing pass runs run_vault_hygiene over the GITHUB vault ONLY (the
// reference's _vault_hygiene_pass — NOT heal_machine_vaults, which is
// the startup heal's both-vaults door). Best-effort — never a killer.
func (e *batchEngine) hygienePass() {
	func() {
		defer func() { recover() }()
		vaulthygiene.RunVaultHygiene(cfgOrStr(e.cfg, "vault_path"),
			e.log, nil, nil)
	}()
}

// runStartCheckGithub mirrors the GitHub vault's §4.4 pass (the dry-run
// branch needs the db PATH even with no connection — the reference's
// run_start_check derives it when db is None).
func (e *batchEngine) runStartCheckGithub(vaultPath string,
	db *notestate.DB) *notestate.RunResult {
	res, err := notestate.RunStartCheck(notestate.VaultGithub, vaultPath,
		db, e.log, nil, cacheDBPath())
	if err != nil {
		return nil // never break a run
	}
	return res
}

// runStartCheckWebsites mirrors the websites vault's §4.4 pass (the
// taxonomy load failure is spoken + skipped, exactly like the
// reference).
func (e *batchEngine) runStartCheckWebsites(wv string,
	db *notestate.DB) *notestate.RunResult {
	cfgPlain := cfgPlainMap(e.cfg)
	tax, err := taxonomy.LoadTaxonomyFromConfig(cfgPlain)
	if err != nil {
		e.log(fmt.Sprintf("⚠️ Note state: websites taxonomy unavailable (%s) — "+
			"websites vault check skipped", pyErrString(err)), "warning")
		return nil
	}
	res, rerr := notestate.RunStartCheck(notestate.VaultWebsites, wv, db,
		e.log, tax, "")
	if rerr != nil {
		return nil
	}
	return res
}

// createInboxNotes mirrors _create_inbox_notes: bypassed (with the
// ON line) when the Websites pipeline is ON.
func (e *batchEngine) createInboxNotes(nonGithubURLs []string, source string) {
	pipes := getMap(e.cfg, "pipelines")
	if pyTruthy(getOr(pipes, "websites", false)) {
		e.log(fmt.Sprintf("🌐 Websites pipeline ON — %d link(s) will be "+
			"processed as websites (not written to _inbox).", len(nonGithubURLs)), "info")
		return
	}
	cfgPlain := cfgPlainMap(e.cfg)
	platformintake.WriteInboxLinksByPlatform(
		platformintake.InboxTableVault(cfgPlain), nonGithubURLs, source,
		e.log, links.BlockedDomainsFromConfig(cfgPlain), nil)
}

// fetchFromImport mirrors _fetch_from_import (inputs.py 152-232): the
// one-link grammar + the inbox notes + the intake counters.
func (e *batchEngine) fetchFromImport() []string {
	if e.importFile == "" || !fileExists(e.importFile) {
		e.log("Import file not found.", "error")
		return nil
	}
	text, err := links.ReadImportFile(e.importFile)
	if err != nil {
		e.log("Could not read import file: "+pyErrString(err), "error")
		return nil
	}
	parsed := links.ParseImportText(text)
	githubURLs := parsed.GitHubURLs
	nonGithubURLs := parsed.WebsiteURLs

	if len(nonGithubURLs) > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Failed to record non-GitHub links from "+
						"import: %s", pyCrashText(r)), "warning")
				}
			}()
			e.createInboxNotes(nonGithubURLs, "Import")
		}()
		e.setNonGithub(nonGithubURLs)
	}

	e.intakeDuplicates = parsed.Duplicates
	e.rawURLCount = parsed.RawCount

	if parsed.Duplicates > 0 {
		e.log(fmt.Sprintf("🔄 %d duplicate address(es) in the file ignored "+
			"(%d total → %d unique)", parsed.Duplicates, parsed.RawCount,
			len(githubURLs)+len(nonGithubURLs)), "info")
	}
	if len(parsed.Unparsed) > 0 {
		e.log(fmt.Sprintf("⚠️ %d line(s) with no recognizable address skipped:",
			len(parsed.Unparsed)), "warning")
		for _, line := range parsed.Unparsed {
			if len(parsed.Unparsed) > 20 {
				// the reference lists at most 20 + the "… and N more" line
				break
			}
			e.log("     · "+pySliceRunes(line, 0, 100), "warning")
		}
		if len(parsed.Unparsed) > 20 {
			for _, line := range parsed.Unparsed[:20] {
				e.log("     · "+pySliceRunes(line, 0, 100), "warning")
			}
			e.log(fmt.Sprintf("     … and %d more", len(parsed.Unparsed)-20), "warning")
		}
	}

	e.log(fmt.Sprintf("📄 Loaded %d GitHub URLs + %d website URLs from import file.",
		len(githubURLs), len(nonGithubURLs)), "info")
	if len(githubURLs) == 0 && len(nonGithubURLs) == 0 {
		e.log("The import file contained no addresses — only blank lines and "+
			"# comments were found.", "warning")
	}
	return githubURLs
}

// backlogURLs extracts the scanned items' URLs.
func backlogURLs(items []websitepipeline.BacklogItem) []string {
	var out []string
	for _, it := range items {
		out = append(out, it.URL)
	}
	return out
}

// waitingURLs extracts the waiting rows' URLs.
func waitingURLs(rows []websitepipeline.WaitingRow) []string {
	var out []string
	for _, r := range rows {
		out = append(out, r.URL)
	}
	return out
}

// ---------------------------------------------------------------------------
// The warmup family (processing_worker.py 470-748)
// ---------------------------------------------------------------------------

// warmupFamily mirrors the provider ladder. Returns true when the
// batch must abort (the finished signal has already fired).
func (e *batchEngine) warmupFamily() bool {
	ollamaBase := e.nestedStrOr("ollama", "base_url",
		"http://localhost:11434")
	ollamaModel := e.nestedStrOr("ollama", "model", "qwythos-9b")
	llmProvider := cfgOrStr(e.cfg, "llm_provider")
	if llmProvider == "" {
		llmProvider = "ollama"
	}

	switch llmProvider {
	case "ollama":
		client := e.c.ollamaFactoryDoor(ollamaBase)
		// v30 — list() with a 15 s wall-clock timeout.
		if err := llmclient.CallWithTimeout(15, func(ctx context.Context) error {
			_, lerr := client.List(ctx)
			return lerr
		}); err != nil {
			e.log("❌ Ollama is not running: "+pyErrString(err), "error")
			if !e.autostartOllama(ollamaBase, client) {
				e.finish(false, "Ollama not available")
				return true
			}
		}
		// v22 Feature 8 — the warmup (best-effort).
		e.log(fmt.Sprintf("🔥 Warming up Ollama model '%s'...", ollamaModel), "info")
		if err := e.warmupChat(client, ollamaModel); err != nil {
			// v30 — the model adaptation ladder.
			e.modelAdaptationLadder(client, ollamaModel, &ollamaModel, err)
		} else {
			e.log("✅ Model warmed up", "success")
		}
		e.warmModel = ollamaModel
		e.warmOllama = client
		e.warmProvider = "ollama"
		return false

	case "llamacpp":
		// v0.15.0 — the llama.cpp engine detection.
		llamaURL := cfgOrStr(e.cfg, "llamacpp_api_url")
		if llamaURL == "" {
			llamaURL = llmclient.LLamaCppDefaultBase
		}
		llamaKey := cfgOrStr(e.cfg, "llamacpp_api_key")
		probe := e.c.probeLlamaDoor(llamaURL, llamaKey, llmclient.LLamaCppProbeTimeoutS)
		if !probe.Found {
			scan := e.c.detectLlamaDoor(llamaKey)
			if scan != nil && scan.BaseURL != "" {
				newURL := scan.BaseURL + "/v1"
				e.log(fmt.Sprintf("🔄 llama.cpp not at %s — detected at %s "+
					"(switching for this batch and saving to Settings).",
					llamaURL, scan.BaseURL), "warning")
				llamaURL = newURL
				e.cfg.Set("llamacpp_api_url", newURL)
				probe = scan
			}
		}
		if !probe.Found {
			e.log("❌ No llama.cpp server detected (no llama-server process, "+
				"nothing on the common ports). Start it with: llama-server -m "+
				"<model>.gguf --port 8080\n   (the app auto-detects llama.cpp at "+
				"launch and here in Settings → 🧠 LLM → 🔍 Detect, or switch "+
				"providers).", "error")
			e.finish(false, "llama.cpp server not detected")
			return true
		}
		llamaModelPtr := llmclient.ResolveLlamaCppModel(e.cfg, probe.Models,
			probe.PropsModel)
		llamaModel := ""
		if llamaModelPtr != nil {
			llamaModel = *llamaModelPtr
		}
		if llamaModel == "" {
			e.log("❌ llama.cpp server detected but no model is loaded (start "+
				"llama-server with -m <model>.gguf).", "error")
			e.finish(false, "llama.cpp has no model loaded")
			return true
		}
		configured := strings.TrimSpace(cfgOrStr(e.cfg, "llamacpp_model"))
		if llamaModel != configured {
			msg := fmt.Sprintf("🔄 llama.cpp model auto-detected: '%s'", llamaModel)
			if configured != "" {
				msg += fmt.Sprintf(" (was '%s')", configured)
			}
			msg += " — saved to Settings."
			e.log(msg, "success")
			e.applyModelChoiceInPlace(llamaModel)
		} else if len(probe.Models) > 0 {
			lowered := map[string]bool{}
			for _, n := range probe.Models {
				lowered[strings.ToLower(n)] = true
			}
			if !lowered[strings.ToLower(llamaModel)] {
				var first5 []string
				for i, n := range probe.Models {
					if i >= 5 {
						break
					}
					first5 = append(first5, n)
				}
				e.log(fmt.Sprintf("⚠️ '%s' is not in the server's /v1/models list "+
					"(%s) — llama-server serves the loaded model; trying anyway.",
					llamaModel, strings.Join(first5, ", ")), "warning")
			}
		}
		if probe.Ready != nil && !*probe.Ready {
			e.log("⏳ llama.cpp is still loading the model — first calls may wait "+
				"or fail; the batch proceeds.", "warning")
		}
		e.log(fmt.Sprintf("🦙 Using llama.cpp server %s · model '%s'",
			probe.BaseURL, llamaModel), "info")
		e.warmModel = llamaModel
		e.warmProvider = "llamacpp"
		return false

	default:
		// The cloud branch: no warmup, the /models pre-flight, and the
		// model name carried for the retry messages.
		cloudModel := cfgOrStrOr(e.cfg, "cloud_model", "gpt-4o-mini")
		cloudURL := cfgOrStrOr(e.cfg, "cloud_api_url", "https://api.openai.com/v1")
		cloudKey := cfgOrStr(e.cfg, "cloud_api_key")
		flavor := "OpenAI-compatible"
		if llmclient.IsAnthropicURL(cloudURL) {
			flavor = "Claude"
		}
		e.log(fmt.Sprintf("☁️ Using Cloud API (%s): %s / model '%s'",
			flavor, cloudURL, cloudModel), "info")
		modelsCfg := getMap(e.cfg, "models")
		type modelCheck struct{ label, m string }
		checks := []modelCheck{{"model", strings.TrimSpace(cloudModel)}}
		for _, t := range []string{"classify", "analyze"} {
			if m := strings.TrimSpace(getOr(modelsCfg, t, "").(string)); m != "" {
				checks = append(checks, modelCheck{"models." + t, m})
			}
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Endpoint pre-flight failed: %s — the "+
						"batch proceeds; per-link errors will surface any real "+
						"problem.", pyCrashText(r)), "warning")
				}
			}()
			var names []string
			var err error
			if llmclient.IsAnthropicURL(cloudURL) {
				names, err = llmclient.AnthropicListModels(cloudURL, cloudKey, 15)
			} else {
				names, err = llmclient.OpenAIListModels(cloudURL, cloudKey, 15)
			}
			if err != nil {
				// CloudLLMError → the info line; other exceptions → the
				// warning line (the reference's two excepts).
				if isCloudLLMError(err) {
					e.log(fmt.Sprintf("ℹ️ /models pre-flight unavailable (%s) — "+
						"skipping the model check; the batch proceeds.",
						pyErrString(err)), "info")
				} else {
					e.log(fmt.Sprintf("⚠️ Endpoint pre-flight failed: %s — the "+
						"batch proceeds; per-link errors will surface any real "+
						"problem.", pyErrString(err)), "warning")
				}
				return
			}
			lowered := map[string]bool{}
			for _, n := range names {
				lowered[strings.ToLower(n)] = true
			}
			var missing []modelCheck
			for _, ck := range checks {
				if ck.m != "" && !lowered[strings.ToLower(ck.m)] {
					missing = append(missing, ck)
				}
			}
			if len(missing) > 0 {
				for _, ck := range missing {
					e.log(fmt.Sprintf("⚠️ Endpoint pre-flight: '%s' (%s) is NOT "+
						"in the /models list (%d listed). Servers that load models "+
						"on demand may still work — the run continues; per-link "+
						"failures will name the model if it is wrong.",
						ck.m, ck.label, len(names)), "warning")
				}
			} else {
				e.log(fmt.Sprintf("✅ Endpoint pre-flight: %d model(s) listed; "+
					"every configured model is available.", len(names)), "success")
			}
		}()
		e.warmModel = cloudModel
		e.warmProvider = "cloud"
		return false
	}
}

// warmupChat sends the tiny "Hi" prompt (num_predict 1).
func (e *batchEngine) warmupChat(client llmclient.OllamaAPI,
	model string) error {
	return llmclient.CallWithTimeout(120, func(ctx context.Context) error {
		_, cerr := client.Chat(ctx, &llmclient.OllamaChatRequest{
			Model:    model,
			Messages: []llmclient.Message{{Role: "user", Content: "Hi"}},
			Options:  &llmclient.OllamaOptions{NumPredict: 1},
		})
		return cerr
	})
}

// autostartOllama mirrors _autostart_ollama: spawn 'ollama serve'
// detached, poll 16 × 1.5 s (4 s probe timeout).
func (e *batchEngine) autostartOllama(baseURL string,
	client llmclient.OllamaAPI) bool {

	e.log("🚀 Ollama server not reachable — trying to start it automatically "+
		"('ollama serve')...", "info")
	pid, spawnErr := func() (int, string) {
		if e.c.AutostartServe != nil {
			return e.c.AutostartServe()
		}
		return spawnOllamaServe()
	}()
	if spawnErr != "" {
		if spawnErr == "FileNotFoundError" {
			e.log("❌ 'ollama' was not found on PATH — Ollama is not installed "+
				"(or not in PATH).", "error")
			e.log("   Install it from https://ollama.com/download, then run "+
				"'ollama pull <model>' and click SYNC again.", "error")
			return false
		}
		e.log("❌ Could not start Ollama: "+spawnErr, "error")
		return false
	}
	e.log(fmt.Sprintf("   started 'ollama serve' (PID %d) — waiting for it to "+
		"answer...", pid), "info")
	for attempt := 0; attempt < 16; attempt++ {
		time.Sleep(1500 * time.Millisecond)
		if err := llmclient.CallWithTimeout(4, func(ctx context.Context) error {
			_, lerr := client.List(ctx)
			return lerr
		}); err == nil {
			e.log(fmt.Sprintf("✅ Ollama is up (probe %d/16) — continuing with "+
				"the batch.", attempt+1), "success")
			return true
		}
	}
	e.log("❌ Ollama did not come up within ~25s. Start it manually (Settings → "+
		"LLM → 🚀 Start Server, or run 'ollama serve' in a terminal), then click "+
		"SYNC again.", "error")
	return false
}

// modelAdaptationLadder mirrors the warmup-failure adaptation: the
// single-model auto-switch, the console model picker (the
// model_prompt_callback bridge), the best-stand-in auto-pick.
func (e *batchEngine) modelAdaptationLadder(client llmclient.OllamaAPI,
	ollamaModel string, modelOut *string, warmupErr error) {

	func() {
		defer func() { recover() }() // the reference's outer except-pass
		e.log(fmt.Sprintf("⚠️ Model warmup failed: %s", pyErrString(warmupErr)), "warning")
		available := []string{}
		if models, err := llmclient.ListModelsWithTimeout(client, 15); err == nil {
			available = models
		}
		containsModel := func(m string) bool {
			for _, a := range available {
				if a == m {
					return true
				}
			}
			return false
		}
		if !containsModel(ollamaModel) && len(available) == 1 {
			newModel := available[0]
			e.log(fmt.Sprintf("🔄 Configured model '%s' not found — Ollama has "+
				"only '%s'. Auto-switching to it (saved to Settings).",
				ollamaModel, newModel), "success")
			*modelOut = newModel
			e.applyModelChoiceInPlace(newModel)
		} else if len(available) > 0 && !containsModel(ollamaModel) {
			// Resolve a stand-in BEFORE the batch starts.
			origModel := ollamaModel
			choice := ""
			func() {
				defer func() { recover() }()
				choice = e.workerModelPrompt(ollamaModel, available)
			}()
			if choice != "" && containsModel(choice) {
				*modelOut = choice
				e.applyModelChoiceInPlace(choice)
				e.log(fmt.Sprintf("🔄 Switched to '%s' (your choice — applied to "+
					"the rest of the batch and saved).", choice), "success")
				// Verify the stand-in actually warms up.
				if err := e.warmupChat(client, *modelOut); err == nil {
					e.log(fmt.Sprintf("✅ Replacement model '%s' warmed up",
						*modelOut), "success")
				} else {
					choice = ""
					e.log(fmt.Sprintf("⚠️ '%s' also failed to warm up — trying the "+
						"best available stand-in…", *modelOut), "warning")
				}
			}
			if choice == "" {
				// No console / user declined / stand-in failed: auto-pick.
				best := githubpipeline.PickBestModel(origModel, available)
				if best != "" && best != *modelOut {
					*modelOut = best
					e.applyModelChoiceInPlace(best)
					e.log(fmt.Sprintf("🔄 Auto-selected '%s' as the closest "+
						"installed stand-in for '%s' (applied to the rest of the "+
						"batch and saved). Pull the original with: ollama pull <model>",
						best, origModel), "warning")
				}
			}
		}
	}()
}

// isCloudLLMError answers whether the error is the CloudLLMError family
// (the reference's first except names the class; the SHAPE is the
// contract).
func isCloudLLMError(err error) bool {
	if err == nil {
		return false
	}
	name := pyErrorClass(err)
	return name == "CloudLLMError" || name == "CloudLLMHTTPError" ||
		name == "CloudLLMBadResponse"
}

// ---------------------------------------------------------------------------
// The websites phase (worker/website_phase.py 95-515)
// ---------------------------------------------------------------------------

// websitePhaseSummary mirrors _website_summary (the richer shape the
// run's done message + the manifest marking consume).
type websitePhaseSummary struct {
	counters     map[string]int
	results      []websitepipeline.Result
	banished     int
	banishedURLs []string
	banishGate   websitepipeline.BanishGate
	excludedVerd int
	vault        string
}

func (e *batchEngine) runWebsitePhase(cache *store.CacheDB,
	noteStateDB *notestate.DB, ollamaClient llmclient.OllamaAPI,
	ollamaModel string) *websitePhaseSummary {

	summary := func() *websitePhaseSummary {
		defer func() {
			if r := recover(); r != nil {
				// One failing vault/pipeline never stops the rest of the
				// batch — but it IS reported loudly.
				func() {
					defer func() { recover() }()
					e.log(fmt.Sprintf("⚠️ Websites pipeline failed (GitHub results "+
						"above are unaffected): %s", pyCrashText(r)), "error")
				}()
			}
		}()
		cfgPlain := cfgPlainMap(e.cfg)
		pipes := getMap(e.cfg, "pipelines")
		if !pyTruthy(getOr(pipes, "websites", false)) {
			return nil
		}
		websiteVault := strings.TrimSpace(cfgOrStr(e.cfg, "website_vault_path"))
		linksList := append([]string(nil), e.nonGithub...)

		// v0.20/v0.21/v0.35 — the intake drops: banned + self domains.
		blocked := links.BlockedDomainsFromConfig(cfgPlain)
		selfDomains := links.SelfDomainsFromConfig(cfgPlain)
		if (len(blocked) > 0 || len(selfDomains) > 0) && len(linksList) > 0 {
			kept, dropBlocked, dropSelf := []string{}, []string{}, []string{}
			for _, u := range linksList {
				if len(blocked) > 0 && links.DomainIsBlocked(u, blocked) {
					dropBlocked = append(dropBlocked, u)
				} else if len(selfDomains) > 0 && links.DomainIsSelf(u, selfDomains) {
					dropSelf = append(dropSelf, u)
				} else {
					kept = append(kept, u)
				}
			}
			if len(dropBlocked) > 0 {
				e.log(fmt.Sprintf("🚫 %d link(s) on banned domains (%s) — omitted "+
					"entirely: never fetched, never noted, never collected",
					len(dropBlocked), strings.Join(blocked, ", ")), "info")
				if e.runner != nil && e.runner.LinkTracker != nil {
					for _, u := range dropBlocked {
						func() {
							defer func() { recover() }()
							e.runner.LinkTracker.MarkBlocked(u, "blocked domain")
						}()
					}
				}
			}
			if len(dropSelf) > 0 {
				e.log(fmt.Sprintf("🔒 %d link(s) on self domains (%s — the app's own "+
					"bot) — never fetched; the _inbox row keeps the record (tokens "+
					"scrubbed)", len(dropSelf), strings.Join(selfDomains, ", ")), "info")
				if e.runner != nil && e.runner.LinkTracker != nil {
					for _, u := range dropSelf {
						func() {
							defer func() { recover() }()
							e.runner.LinkTracker.MarkBlocked(u,
								"self domain (the app's own bot)")
						}()
					}
				}
			}
			linksList = kept
		}
		if websiteVault == "" {
			if len(linksList) > 0 {
				e.log("⚠️ Websites pipeline is ON but no Websites vault is set "+
					"(Settings → 📁 Vault) — the non-GitHub links of this batch are "+
					"kept in the manifest only.", "warning")
				// v0.37.0 — give the denominator back.
				if e.runner != nil {
					e.runner.Total -= len(linksList)
					if e.runner.Total < 0 {
						e.runner.Total = 0
					}
				}
			}
			return nil
		}

		// The state DB: the shadow cache during a dry-run.
		statePath := ""
		if dryrun.IsEnabled() {
			statePath = dryrun.ShadowCachePath(cacheDBPath())
		}
		state, serr := websitepipeline.NewWebsiteStateDB(statePath)
		if serr != nil {
			panic(serr)
		}
		defer state.Close()

		// v0.63.2 — THE SETTLEMENT at the batch's own door.
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Settlement pass skipped: %s",
						pyCrashText(r)), "warning")
				}
			}()
			websitepipeline.SettleTheLedger(state, websiteVault, e.log, nil)
		}()

		// The websites vault gets its OWN VaultIndex (the website
		// normalizer — query params are part of the identity).
		index := batch.NewVaultIndex(websiteVault,
			func(u string) string { return links.NormalizeWebsiteURL(u) })
		index.Rebuild(e.log)

		// The LLM router — both providers through the shared helpers,
		// honoring the per-task model overrides.
		llmProvider := cfgOrStr(e.cfg, "llm_provider")
		if llmProvider == "" {
			llmProvider = "ollama"
		}
		numCtx := pyIntDefaultCfg(e.cfg, "llm_num_ctx", llmclient.DefaultNumCtx)
		if numCtx == 0 {
			numCtx = 0 // None when falsy
		}
		warn := func(m string) { e.log(m, "warning") }
		llmCall := func(messages []llmclient.Message, task string) string {
			timeoutS := pyFloatCfgOr(e.cfg, "llm_timeout_s", 300)
			switch llmProvider {
			case "cloud":
				model := llmclient.ResolveTaskModel(e.cfg, task,
					cfgOrStr(e.cfg, "cloud_model"))
				out, _ := llmclient.CloudChat(cfgOrStr(e.cfg, "cloud_api_url"),
					cfgOrStr(e.cfg, "cloud_api_key"), model, messages, timeoutS,
					&llmclient.ChatOptions{JSONMode: true, NumCtx: numCtx,
						OnWarn: warn, VerifyTLS: connectioncheck.VerifySSLEnabled(e.cfg)})
				return out
			case "llamacpp":
				model := llmclient.ResolveTaskModel(e.cfg, task,
					cfgOrStr(e.cfg, "llamacpp_model"))
				out, _ := llmclient.CloudChat(
					llmclient.NormalizeLlamaCppAPIURL(cfgOrStr(e.cfg, "llamacpp_api_url")),
					cfgOrStr(e.cfg, "llamacpp_api_key"), model, messages, timeoutS,
					&llmclient.ChatOptions{JSONMode: true, NumCtx: numCtx,
						OnWarn: warn, VerifyTLS: connectioncheck.VerifySSLEnabled(e.cfg)})
				return out
			default:
				model := llmclient.ResolveTaskModel(e.cfg, task, ollamaModel)
				out, _ := llmclient.OllamaChat(ollamaClient, model, messages,
					timeoutS, &llmclient.OllamaChatOptions{NumCtx: numCtx, OnWarn: warn})
				return out
			}
		}

		// v0.60.0 — the banish gate's channel (the Telegram
		// round-trip; None when not paired/enabled — the defer
		// default).
		banishFn := e.makeBanishConfirm()
		tax, terr := taxonomy.LoadTaxonomyFromConfig(cfgPlain)
		if terr != nil {
			panic(terr)
		}
		pipeline, perr := websitepipeline.NewPipeline(cfgPlain, llmCall,
			func(url string) any { return index.HasURL(url) },
			state, e.log, tax, noteStateDB, nil, nil, banishFn,
			func(url string) string { return index.SiteNoteFor(url) })
		if perr != nil {
			panic(perr)
		}

		// v0.60.1 — the pre-drop: links whose verdict is already
		// written leave the batch BEFORE the count.
		split := websitepipeline.SplitDismissedLinks(state, linksList, e.log)
		if len(split.Dropped) > 0 {
			linksList = split.Kept
			if e.runner != nil && e.runner.LinkTracker != nil {
				for _, u := range split.Dropped {
					func() {
						defer func() { recover() }()
						e.runner.LinkTracker.MarkSkipped(u,
							"already decommissioned/banished (excluded before "+
								"processing — never fetched again)")
					}()
				}
			}
		}

		var due []string
		if dl, derr := state.DueRetries(); derr == nil {
			due = dl
		}

		// v0.37.0 — the bar's own honesty: re-base the denominator.
		githubDone := 0
		if e.runner != nil {
			githubDone = e.runner.Position()
		}
		if e.runner != nil {
			e.runner.SetTotal(githubDone + len(linksList) + len(due))
		}

		wpAdvance := func(u string) {
			if e.runner != nil {
				e.runner.AdvancePosition(u)
			}
		}
		shouldContinue := func() bool {
			if e.runner == nil {
				return true
			}
			return e.runner.RunningNow()
		}

		var results []websitepipeline.Result
		backlogItems := e.reviewBacklogItems
		if len(backlogItems) > 0 {
			unique := map[string]bool{}
			for _, it := range backlogItems {
				unique[it.URL] = true
			}
			if e.runner != nil {
				e.runner.SetTotal(githubDone + len(unique))
			}
			if e.masterRetryMode {
				e.log(fmt.Sprintf("🔁 Master-table waiting pass: %d ' - ' link(s) "+
					"— valid links whose fetches failed and no verdict is set; each "+
					"is fetched again before 'everything is up to date' may be said",
					len(unique)), "info")
				results = pipeline.RetryMasterWaiting(shouldContinue, wpAdvance)
			} else {
				e.log(fmt.Sprintf("🔁 Retrying the _review backlog: %d app-owned "+
					"placeholder(s) (%d link(s)) — v0.41.0 fetches as a browser now; "+
					"the bot-defense wall may be down for them",
					len(backlogItems), len(unique)), "info")
				results = pipeline.RetryReviewBacklog(backlogItems,
					shouldContinue, wpAdvance)
			}
		} else {
			dueTxt := ""
			if len(due) > 0 {
				dueTxt = fmt.Sprintf(" + %d due retry(ies)", len(due))
			}
			e.log(fmt.Sprintf("🌐 Websites pipeline: %d link(s)%s",
				len(linksList), dueTxt), "info")
			if len(due) > 0 {
				pipeline.RunDueRetries(shouldContinue, wpAdvance)
			}
			results = pipeline.Run(linksList, shouldContinue, wpAdvance)
		}

		// Manifest marking (the intake left website links pending).
		if e.runner != nil && e.runner.LinkTracker != nil {
			for _, r := range results {
				func() {
					defer func() { recover() }()
					switch r.Outcome {
					case "processed", "review":
						e.runner.LinkTracker.MarkProcessed(r.URL, r.NotePath)
					case "skipped":
						e.runner.LinkTracker.MarkSkipped(r.URL, r.Error)
					case "failed":
						e.runner.LinkTracker.MarkFailed(r.URL, r.Error)
					}
				}()
			}
		}

		// v0.28.0 — the Website Directory rebuild.
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ Website Directory rebuild skipped: %s",
						pyCrashText(r)), "warning")
				}
			}()
			if dir := websitepipeline.BuildWebsiteDirectory(websiteVault,
				pipeline.Taxonomy(), cfgPlain, e.log); dir != nil {
				reviewTxt := ""
				if dir.Review > 0 {
					reviewTxt = fmt.Sprintf(" (+%d in review)", dir.Review)
				}
				e.log(fmt.Sprintf("📚 Website Directory rebuilt: %d site(s) in %d "+
					"categor(ies)%s → %s", dir.Sites, dir.Categories, reviewTxt,
					filepath.Base(dir.Path)), "success")
			}
		}()

		// v0.35.0 — prune the _inbox tables now that the batch is done.
		func() {
			defer func() {
				if r := recover(); r != nil {
					e.log(fmt.Sprintf("⚠️ _inbox table prune skipped: %s",
						pyCrashText(r)), "warning")
				}
			}()
			tblVault := platformintake.InboxTableVault(cfgPlain)
			if tblVault != "" {
				var storedURLs []string
				for _, r := range results {
					if r.Outcome == "processed" || r.Outcome == "review" {
						storedURLs = append(storedURLs,
							links.NormalizeWebsiteURL(r.URL))
						if r.Canonical != "" {
							storedURLs = append(storedURLs, r.Canonical)
						}
					}
				}
				platformintake.PruneInboxTables(tblVault, blocked,
					func(url string) bool { return index.HasURL(url) },
					e.log, storedURLs)
			}
		}()

		// The summary + the done line.
		banishedURLs := pipeline.BanishedURLs()
		gate := pipeline.Gate()
		sum := &websitePhaseSummary{
			counters:     pipeline.Counters(),
			results:      results,
			banishedURLs: banishedURLs,
			banishGate:   gate,
			excludedVerd: len(split.Dropped),
			vault:        websiteVault,
		}
		sum.banished = len(banishedURLs)
		c := pipeline.Counters()
		tally := ""
		if sum.banished > 0 {
			tally = fmt.Sprintf(" 🗑️ %d banished — removed and never fetched "+
				"again (you confirmed).", sum.banished)
		}
		kept := gate.Pending
		verdict := gate.Verdict
		if verdict == "" {
			verdict = "auto"
		}
		if sum.banished == 0 && kept > 0 && verdict != "auto" {
			tally = fmt.Sprintf(" 🗑️ %d marked note(s) KEPT — waiting on your "+
				"confirmation (I'll ask again next run).", kept)
		}
		e.log(fmt.Sprintf("🌐 Websites done: %d processed, %d to review, %d "+
			"skipped, %d retried, %d upgraded, %d failed.%s",
			c["processed"], c["review"], c["skipped"], c["retried"],
			c["upgraded"], c["failed"], tally), "success")

		if e.runner != nil {
			e.runner.SetWebsiteSummary(&githubpipeline.WebsiteSummary{
				Counters: c, Results: nil, Vault: websiteVault,
			})
		}
		return sum
	}()
	if summary == nil {
		return nil
	}
	return summary
}

// makeBanishConfirm mirrors make_telegram_confirm: the Telegram
// round-trip over the Worker's banish trio; nil (the defer default)
// when the Worker is not paired/enabled. Never raises.
func (e *batchEngine) makeBanishConfirm() websitepipeline.BanishConfirmFn {
	cfgPlain := cfgPlainMap(e.cfg)
	s := workerclient.NewSync(cfgPlain)
	if !s.IsEnabled() {
		return nil
	}
	return func(items []websitepipeline.PendingBanishment,
		log websitepipeline.LogFn) *websitepipeline.BanishConfirmResult {
		if log == nil {
			log = func(string, string) {}
		}
		timeoutS := pyFloatCfgOr(e.cfg, "banish_confirm_timeout_s", 300)
		if timeoutS == 0 {
			timeoutS = 300
		}
		payload := make([]any, 0, len(items))
		for _, it := range items {
			row := storage.NewOrderedMap()
			row.Set("url", it.URL)
			row.Set("title", it.Title)
			row.Set("marker", it.Marker)
			row.Set("door", it.Door)
			payload = append(payload, row)
		}
		log(fmt.Sprintf("📨 Deletion review sent to Telegram — %d note(s) marked "+
			"for deletion, waiting for your answer (buttons in the bot chat; up "+
			"to %ds)", len(payload), int(timeoutS)), "info")
		confID, why := s.ProposeBanish(len(payload), payload, timeoutS)
		if confID == nil {
			log(fmt.Sprintf("⚠️ The deletion review could not reach Telegram (%s) — "+
				"nothing deleted, the marks stay; I'll ask again next run", why), "warning")
			return &websitepipeline.BanishConfirmResult{Verdict: "defer"}
		}
		verdict := "timeout"
		deadline := time.Now().Add(time.Duration(
			maxFloat(timeoutS, banishPollIntervalS) * float64(time.Second)))
		for time.Now().Before(deadline) {
			time.Sleep(time.Duration(banishPollIntervalS * float64(time.Second)))
			status := strings.TrimSpace(strings.ToLower(s.BanishStatus(*confID)))
			switch status {
			case "confirmed", "declined":
				verdict = status
			case "":
				continue // a missed poll is not an answer — keep waiting
			case "superseded", "timeout", "done", "stale":
				verdict = "timeout"
			default:
				continue
			}
			if verdict == "confirmed" || verdict == "declined" {
				break
			}
			if status == "superseded" || status == "timeout" ||
				status == "done" || status == "stale" {
				break
			}
		}
		if verdict == "timeout" {
			func() {
				defer func() { recover() }()
				s.ReportBanishResult(*confID, "timeout", 0)
			}()
		}
		out := &websitepipeline.BanishConfirmResult{Verdict: verdict}
		if verdict == "confirmed" {
			id := *confID
			out.Report = func(deleted int) {
				defer func() { recover() }()
				s.ReportBanishResult(id, "deleted", deleted)
			}
		}
		return out
	}
}

// banishPollIntervalS mirrors POLL_INTERVAL_S.
const banishPollIntervalS = 3.0

// maxFloat mirrors max(a, b).
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// cfgOrStrOr mirrors cfg.get(key, default) through str() — a present
// value renders even when falsy ("" stays "").
func cfgOrStrOr(cfg *storage.OrderedMap, key, def string) string {
	v, ok := cfg.Get(key)
	if !ok || v == nil {
		return def
	}
	if s, is := v.(string); is {
		return s
	}
	return fmt.Sprintf("%v", v)
}

// nestedStrOr reads cfg[section][key] through str() with a default.
func (e *batchEngine) nestedStrOr(section, key, def string) string {
	sub := getMap(e.cfg, section)
	if sub == nil {
		return def
	}
	v, ok := sub.Get(key)
	if !ok || v == nil {
		return def
	}
	if s, is := v.(string); is {
		if s == "" {
			return def
		}
		return s
	}
	return fmt.Sprintf("%v", v)
}

// pyFloatCfgOr mirrors float(cfg.get(key, default) or default).
func pyFloatCfgOr(cfg *storage.OrderedMap, key string, def float64) float64 {
	v, ok := cfg.Get(key)
	if !ok || v == nil {
		return def
	}
	if f, is := v.(float64); is {
		if f == 0 {
			return def
		}
		return f
	}
	if i, is := v.(int); is {
		if i == 0 {
			return def
		}
		return float64(i)
	}
	return def
}

// pyIntDefaultCfg mirrors int(cfg.get(key, default) or 0) or 0.
func pyIntDefaultCfg(cfg *storage.OrderedMap, key string, def int) int {
	v, ok := cfg.Get(key)
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case int:
		return t
	case float64:
		return int(t)
	case string:
		if n, err := parseIntStrict(t); err == nil {
			return n
		}
	}
	return 0
}

// pyErrString mirrors str(exc) — the error's message.
func pyErrString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// pyErrorClass mirrors type(exc).__name__ — the class-name token law
// (the Go type's bare name).
func pyErrorClass(err error) string {
	if err == nil {
		return ""
	}
	name := fmt.Sprintf("%T", err)
	name = strings.TrimPrefix(name, "*")
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// pyCrashText renders a panic value / error the way the reference
// speaks a crashed exception ("{type}: {e}" — the SHAPE is the
// contract; Go panic values carry no Python class).
func pyCrashText(r any) string {
	if err, isErr := r.(error); isErr {
		return pyErrorClass(err) + ": " + pyErrString(err)
	}
	return fmt.Sprintf("%v", r)
}
