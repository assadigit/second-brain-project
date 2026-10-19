package app

// batchwire.go — T-74 slice 1: THE CLI-PARITY SHELL WIRING, the debt
// T-73 carried and the state card made this session's first slice —
// engine.go's runImpl composition re-hosted on the app surface
// (internal/app), so the GUI shell drives exactly the composition the
// CLI drives over the same engine packages.
//
// THE COMPOSITION (processing_worker.py's run/run_impl + cli_run.py's
// run_batch_visual/_finish, the app's own transport):
//   - the dry-run switch flip/restore + the crash wrapper (the 💥
//     line + finish(false, …) — the finished verdict lands even on a
//     crash; the command's guardCommand is the belt outside);
//   - the pipeline-switches gate (both OFF consumes nothing);
//   - the mode dispatch: direct URLs | import file (the one-link
//     grammar + the _inbox tables + the counters) | review_retry |
//     master_retry (the Sync view's retry doors, T-75's UI) — the
//     Telethon `telegram` mode is Option B's accepted loss (the
//     missing-credentials shape, ADR-0004 Am.3);
//   - THE INTAKE (v23 Phase 1): the tracker with the source law, the
//     banned/self drops, the _inbox tables when the Websites pipeline
//     is OFF, the manifest line;
//   - THE WARMUP FAMILY (the provider ladder — the app's seams):
//     Ollama list/timeout + autostart + warmup + the adaptation
//     ladder (single-model auto-switch, the model-prompt door, the
//     best-stand-in), the llama.cpp re-probe/detect + model resolve,
//     the cloud /models pre-flight; a dead server aborts the batch
//     with the honest verdict;
//   - the setup family: the cache (the dry-run shadow), the
//     note-state connection, the CONFIRMED-dead set, THE REPOS
//     BANISHMENT pass, THE SETTLED REPOS LEDGER, the vault index +
//     SetVaultIndex, the missing-repo backfill, the §4.4 start-of-run
//     passes for BOTH vaults, the dismissed set, the batch snapshot;
//   - the GitHub loop over githubpipeline.Runner (the GUI shape: the
//     LLM failure door rides the bus's 600 s gate when a subscriber
//     answers; the model prompt rides the shell's door);
//   - THE WEBSITES PHASE (website_phase.py's wiring): the banned/self
//     drops, the settlement, the websites vault's own index, the LLM
//     router (both providers, the per-task overrides), the banish
//     gate, the pre-drop, the due retries / the backlog retry modes,
//     the manifest marking, the Website Directory rebuild, the
//     _inbox prune, the summary + the done line;
//   - THE FINISH CHAIN: the runner's own Finish (the undo diff, the
//     verification, the reports, the summary) → the seals in the
//     reference's exact order (VaultSeal → Good Repos → Websites →
//     Personal) → THE PHASE-5 CLEAR (the vault-mirror push is the
//     ack; the all-clear verdict gated by success AND bot_source) →
//     the dry-run report path (side-effect free everywhere).
//
// THE HERMETIC LAW: the composition's network/process doors are the
// App's seam fields (OllamaFactory, ProbeLlama, DetectLlama,
// AutostartServe, CloudListModels, WorkerMirrorPush,
// GHClientFactory, BannerDownload, WriteLiveConfig) — the tests fake
// them at construction; the engine packages underneath are
// golden-tested in their own right. The Worker round-trips
// (banish/mirror) ride workerclient.Sync over the config — a
// disabled Worker is the silent no-op, exactly like the reference.
//
// THE TRANSPORT TWIN: the reference runs the worker on a QThread with
// the GUI spinning app.exec(); the app runs the composition inline on
// the command's goroutine (the shell's own goroutine under wails) and
// every line rides the typed Bus — the same coalescing law T-72
// proved.

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/constants"
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

// banishPollIntervalS mirrors POLL_INTERVAL_S (the Worker banish
// round-trip's poll beat).
const banishPollIntervalS = 3.0

// batchStopHandle is one armed cooperative-stop door (the App's
// batchStop seat; the pointer identity is the arm token).
type batchStopHandle struct {
	fn func()
}

// batchEngine carries the worker-side state of one batch (the CLI's
// batchEngine twin, the app's transport).
type batchEngine struct {
	a    *App
	cfg  *storage.OrderedMap
	opts BatchOptions
	ev   BatchEvents

	nonGithub        []string
	intakeDuplicates int
	rawURLCount      int

	// stopRequested: the cooperative stop's flag (stop_processing's
	// engine half — v0.32). The runner's own flag covers the loops; this
	// seat keeps a stop that arrived during the warmup (before the
	// runner exists) from being lost.
	stopRequested bool

	// The review-backlog modes' state (v0.42/v0.51).
	reviewBacklogItems []websitepipeline.BacklogItem
	masterRetryMode    bool

	// The model the warmup resolved (carries to the runner).
	warmModel    string
	warmOllama   llmclient.OllamaAPI
	warmProvider string

	runner     *githubpipeline.Runner
	vaultIndex *batch.VaultIndex
	tracker    *batch.LinkTracker
	cache      *store.CacheDB

	// The outcome (written by finish; the command reads it).
	fin      *BatchResult
	finished bool
}

// requestStop is the cooperative, between-items stop (v0.32): the
// runner's flag flips — the GitHub loop, the rate-limit waits, and the
// websites phase all check it between items; the finish chain still
// runs; the tracker's unfinished links surface for retry exactly like
// the reference's (no partial state, no confirmation dialog needed).
func (e *batchEngine) requestStop() {
	e.stopRequested = true
	if e.runner != nil {
		e.runner.Stop()
	}
}

// runBatchEngine is the App's batch engine door (runBatch delegates
// here): the dry-run flip + the crash wrapper around runImpl.
func (a *App) runBatchEngine(cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
	if cfg == nil {
		cfg = storage.NewOrderedMap()
	}
	e := &batchEngine{
		a:                a,
		cfg:              cfg,
		opts:             opts,
		ev:               ev,
		nonGithub:        opts.NonGithub,
		intakeDuplicates: opts.IntakeDuplicates,
		rawURLCount:      opts.RawURLCount,
		fin:              &BatchResult{},
	}
	// The cooperative-stop door: while this engine runs, the App's
	// StopBatch reaches requestStop (T-77 — hero.py's Stop leg). The
	// handle is the arm token — the deferred clear only disarms THIS
	// engine's door, never a newer batch's.
	stop := &batchStopHandle{fn: e.requestStop}
	a.setBatchStop(stop)
	defer a.clearBatchStop(stop)
	if opts.DryRun {
		dryrun.Enable()
		e.log("🧪 DRY-RUN: every vault write this batch would perform is "+
			"logged and skipped. Nothing will be written, sealed or "+
			"marked read.", "warning")
	}
	func() {
		defer func() {
			if opts.DryRun {
				dryrun.Disable()
			}
		}()
		defer func() {
			if r := recover(); r != nil {
				// 💥 Batch crashed: {type}: {e} — the SHAPE is the contract
				// (the class token + message; the reference's traceback
				// tail has no Go twin).
				func() {
					defer func() { recover() }()
					e.log(fmt.Sprintf("💥 Batch crashed: %s", pyCrashText(r)), "error")
				}()
				e.finish(false, fmt.Sprintf("Batch crashed: %s", pyCrashText(r)))
			}
		}()
		e.runImpl()
	}()
	return e.fin
}

// log routes a worker log line through the batch events (the
// log_message wiring; nil-safe).
func (e *batchEngine) log(msg, level string) {
	if e.ev.Log != nil {
		e.ev.Log(msg, level)
	}
}

// ---------------------------------------------------------------------------
// _run_impl — the engine body (processing_worker.py 239-1889).
// ---------------------------------------------------------------------------

func (e *batchEngine) runImpl() {
	cfgPlain := cfgPlainMap(e.cfg)
	opts := e.opts

	// v0.10/v0.11 — the pipeline switches, checked BEFORE anything is
	// fetched so an OFF switch consumes nothing.
	pipes := getMapCfg(e.cfg, "pipelines")
	githubOn := pyTruthyCfg(getOrCfg(pipes, "github", true))
	websitesOn := pyTruthyCfg(getOrCfg(pipes, "websites", false))
	if !githubOn && !websitesOn {
		e.log("⛔ Both pipelines are switched OFF (Settings → 📁 Vault). "+
			"Nothing was fetched or processed.", "warning")
		e.finish(true, "Both pipelines are off — nothing to do.")
		return
	}

	// The mode dispatch.
	var urls []string
	mode := opts.Mode
	if mode == "" {
		mode = "direct"
	}
	switch mode {
	case "direct":
		urls = opts.URLs
	case "import":
		urls = e.fetchFromImport(opts.ImportFile)
	case "review_retry":
		// v0.42.0 — the _review backlog retry.
		if !websitesOn {
			e.log("⛔ Websites pipeline is OFF (Settings → 📁 Vault) — "+
				"the _review backlog cannot be retried.", "warning")
			e.finish(true, "Websites pipeline is off — nothing retried.")
			return
		}
		wv := strings.TrimSpace(pyStrCfg(e.cfg, "website_vault_path"))
		if wv == "" || !fsIsDirPath(wv) {
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
		wv := strings.TrimSpace(pyStrCfg(e.cfg, "website_vault_path"))
		if wv == "" || !fsIsDirPath(wv) {
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
					dryrun.ShadowCachePath(cacheDBPathOf(e.cfg)))
			} else {
				state, _ = websitepipeline.NewWebsiteStateDB(cacheDBPathOf(e.cfg))
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
		// The Telethon fetch — Option B's accepted loss (ADR-0004 Am.3):
		// the shell never routes here; the missing-keys shape is the
		// reference's own no-credentials answer.
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
	vaultPath := pyStrCfg(e.cfg, "vault_path")
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
			switch mode {
			case "direct":
				if opts.Source == "bot" {
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
					if opts.Source == "bot" {
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
	e.tracker = tracker

	// --- The warmup family (the provider ladder). A dead server aborts
	// the batch (the early-return law: finish already fired inside). ---
	if e.warmupFamily() {
		return
	}

	// v30 — the cache opens AFTER the Ollama early-return so the
	// "Ollama not available" exit can no longer leak the handle. A
	// dry-run batch reads the REAL cache through a throwaway copy.
	cache, cerr := store.NewCacheDB(func() string {
		if dryrun.IsEnabled() {
			return dryrun.ShadowCachePath(cacheDBPathOf(e.cfg))
		}
		return cacheDBPathOf(e.cfg)
	}())
	if cerr != nil {
		// The reference's constructor raises → the crash wrapper owns
		// the card; mirror its exact speech.
		panic(fmt.Errorf("CacheDB: %s", pyErrString(cerr)))
	}
	e.cache = cache

	// v0.10.0 — one note-state connection for the whole batch (None in
	// a dry-run — a rehearsal records no bookkeeping).
	var noteStateDB *notestate.DB
	if !dryrun.IsEnabled() {
		if ns, err := notestate.New(cacheDBPathOf(e.cfg)); err == nil {
			noteStateDB = ns
		}
	}

	// v0.08 — the CONFIRMED-dead set, loaded ONCE.
	deadThreshold := store.DeadLinkThresholdFromConfig(cfgPlain)
	deadURLs := cache.GetDeadURLSet(deadThreshold)

	// v0.65.0 — THE REPOS BANISHMENT, the batch's opening gate.
	if githubOn {
		rvault := strings.TrimSpace(pyStrCfg(e.cfg, "vault_path"))
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
			r := e.buildRunner(noteStateDB, tracker, total, urls)
			r.BackfillMissingNotes(cache, vaultPath, deadThreshold)
		}()
	}

	// v0.12.0 — the FULL start-of-run §4.4 pass for BOTH vaults.
	noteStateRun := map[string]*notestate.RunResult{}
	if githubOn && vaultPath != "" {
		noteStateRun["github"] = e.runStartCheckGithub(vaultPath, noteStateDB)
	}
	if websitesOn {
		if wv := strings.TrimSpace(pyStrCfg(e.cfg, "website_vault_path")); wv != "" {
			noteStateRun["websites"] = e.runStartCheckWebsites(wv, noteStateDB)
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

	runner := e.buildRunner(noteStateDB, tracker, total, urls)
	e.runner = runner
	runner.SetBatchFiles(snapshot)
	runner.NoteStateRun = noteStateRun

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
	// reports, the batch summary, the finished signal — the runner's
	// Finished callback fires e.finish: the seals + the Phase-5 CLEAR).
	runner.Finish()
}

// ---------------------------------------------------------------------------
// finish — run_batch_visual's _finish handler (the app's transport).
// ---------------------------------------------------------------------------

// finish records the verdict, then (outside a dry-run) runs the seals
// in the reference's exact order + THE PHASE-5 CLEAR. Exactly-once:
// the first verdict wins (the runner's Finish and an abort path never
// both speak).
func (e *batchEngine) finish(success bool, message string) {
	if e.finished {
		return
	}
	e.finished = true
	e.fin.Success = success
	e.fin.Message = message
	if e.runner != nil {
		e.fin.Processed = e.runner.Processed
		e.fin.Total = e.runner.Total
	}
	// The retry-chime selector (T-77): the tracker's failed count at
	// finish — a batch that finished WITH failed links plays the softer
	// descending retry voice (v0.40.0).
	if e.tracker != nil {
		e.fin.FailedLinks = len(e.tracker.GetFailedLinks())
	}
	if e.opts.DryRun {
		// A dry-run batch is side-effect free EVERYWHERE: no git seal,
		// no Good Repos publish, no mirror push; the withheld-writes
		// report is saved and the run stops.
		n := dryrun.EntryCount()
		e.log(fmt.Sprintf(
			"🧪 DRY-RUN COMPLETE — %d vault operation(s) were logged, "+
				"not performed. VaultSeal, Good Repos publish, bot-queue "+
				"mark-read and last_processed_msg_id were all skipped.",
			n), "warning")
		if reportPath := e.writeDryrunReport(); reportPath != "" {
			e.log("🧪 Dry-run report: "+reportPath, "info")
		}
		return
	}

	// The finish seals — best-effort (they run for failed batches too:
	// notes saved before a mid-run failure are exactly what we want
	// backed up).
	processed, total := e.processedTotal()
	e.sealVaultLeg(processed, total)
	e.publishGoodReposLeg(processed, total)
	e.sealWebsitesLeg(processed, total)
	e.sealManualLeg(processed, total)

	// THE PHASE-5 CLEAR (the re-spec: the mirror push is the ack; the
	// all-clear verdict rides the reference's success AND bot_source
	// gate).
	e.phaseFiveClear(e.opts.Source == "bot", success)
}

// processedTotal reads the runner's counters (nil-safe).
func (e *batchEngine) processedTotal() (int, int) {
	if e.runner == nil {
		return 0, 0
	}
	return e.runner.Processed, e.runner.Total
}

// writeDryrunReport mirrors _write_dryrun_report: the dry-run write
// log as Markdown OUTSIDE the vault (APP_DIR/reports/dry-runs/) so the
// owner can read exactly what a dry-run would have changed. A PLAIN
// open on purpose — this runs from the finished moment, which can fire
// a hair BEFORE the engine's defer turns the dry-run switch off.
func (e *batchEngine) writeDryrunReport() string {
	outDir := filepath.Join(constants.AppDir(), "reports", "dry-runs")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return ""
	}
	stamp := time.Now().Format("20060102-150405")
	path := filepath.Join(outDir, "dry_run_"+stamp+".md")
	body := dryrun.RenderReportMarkdown("GitCurator dry-run report")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return ""
	}
	return path
}

// The four seal legs (the reference's exact order: VaultSeal → Good
// Repos → Websites → Personal), each best-effort and crash-isolated.

func (e *batchEngine) sealVaultLeg(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.log(fmt.Sprintf("VaultSeal error: %s", pyCrashText(r)), "warning")
		}
	}()
	e.log("Sealing vault (git commit + push if configured)…", "info")
	summary := &vaultseal.RunSummary{Processed: &processed, Total: &total}
	vs := vaultseal.SealFromConfig(e.cfg, summary, e.silentSealLog())
	level := "warning"
	if vs.OK() {
		level = "success"
	}
	e.log("VaultSeal: "+vs.Describe(), level)
}

func (e *batchEngine) publishGoodReposLeg(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.log(fmt.Sprintf("Good Repos error: %s", pyCrashText(r)), "warning")
		}
	}()
	summary := &goodrepos.RunSummary{Processed: &processed, Total: &total}
	gr := goodrepos.PublishFromConfig(e.cfg, summary, e.silentSealLog())
	level := "warning"
	if gr.OK() {
		level = "success"
	}
	e.log("Good Repos: "+gr.Describe(), level)
}

func (e *batchEngine) sealWebsitesLeg(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.log(fmt.Sprintf("Websites vault seal error: %s", pyCrashText(r)), "warning")
		}
	}()
	summary := &vaultseal.RunSummary{Processed: &processed, Total: &total}
	ws := vaultseal.WebsitesSealFromConfig(e.cfg, summary, e.silentSealLog())
	if ws.Sealed || ws.Error != "" {
		level := "warning"
		if ws.OK() {
			level = "success"
		}
		e.log("Websites vault seal: "+ws.Describe(), level)
	}
}

func (e *batchEngine) sealManualLeg(processed, total int) {
	defer func() {
		if r := recover(); r != nil {
			e.log(fmt.Sprintf("Personal vault seal error: %s", pyCrashText(r)), "warning")
		}
	}()
	summary := &vaultseal.RunSummary{Processed: &processed, Total: &total}
	mn := vaultseal.ManualSealFromConfig(e.cfg, summary, e.silentSealLog())
	if mn.Sealed || mn.Error != "" {
		level := "warning"
		if mn.OK() {
			level = "success"
		}
		e.log("Personal vault seal: "+mn.Describe(), level)
	}
}

// silentSealLog is the sealer's internal log (the reference passes
// log=None to the config bridges — the app speaks the describe lines).
func (e *batchEngine) silentSealLog() func(string, string) {
	return func(string, string) {}
}

// ---------------------------------------------------------------------------
// THE PHASE-5 CLEAR (phase5.go's composition, the app's transport).
// ---------------------------------------------------------------------------

// MirrorPushOutcome carries the honest result of the batch-complete
// vault-mirror push (the Phase-5 CLEAR's ack leg).
type MirrorPushOutcome struct {
	// Attempted is false when the Worker is not configured (the silent
	// law) — never an error, never spoken.
	Attempted bool
	OK        bool
	Entries   int
	Decomm    int
	Error     string
}

// pushVaultMirrorAfterBatch is the batch-complete leg: build the
// entries from the vault index + the decommission events from the
// cache, hash the index (sha256 over '|'.join(sorted(urls)) — the
// manager's own recipe), push (idempotent by sync_id), and speak the
// honest outcome. A nil index or cache pushes nothing.
func (e *batchEngine) pushVaultMirrorAfterBatch() *MirrorPushOutcome {
	if e.vaultIndex == nil {
		return &MirrorPushOutcome{}
	}
	idxData := e.vaultIndex.MirrorIndex()
	entries := workerclient.BuildVaultMirrorEntries(idxData)
	var decomm []any
	if e.cache != nil {
		rows := e.cache.GetAllDecommissioned()
		dr := make([]workerclient.DecommissionEventRow, 0, len(rows))
		for _, r := range rows {
			dr = append(dr, workerclient.DecommissionEventRow{
				URL: r.URL, Reason: r.Error})
		}
		decomm = workerclient.BuildDecommissionEvents(dr)
	}
	urls := idxData.Keys()
	sort.Strings(urls)
	sum := sha256.Sum256([]byte(strings.Join(urls, "|")))
	hash := hex.EncodeToString(sum[:])

	var attempted, ok bool
	var resp any
	if e.a.WorkerMirrorPush != nil {
		attempted, ok, resp = e.a.WorkerMirrorPush(e.cfg, entries, decomm, hash)
	}
	out := &MirrorPushOutcome{
		Attempted: attempted, OK: ok,
		Entries: len(entries), Decomm: len(decomm),
	}
	if attempted {
		if ok {
			e.log(fmt.Sprintf(
				"☁️ Worker synced: %d vault entry(ies) and %d decommission event(s) pushed — the acknowledged links leave the pending queue.",
				out.Entries, out.Decomm), "info")
		} else {
			out.Error = storage.PyDumpsCompact(resp)
			e.log(fmt.Sprintf(
				"⚠️ Vault-mirror push failed: %s — the queue keeps its state (the links stay pending; the next --auto re-offers them).",
				out.Error), "warning")
		}
	}
	return out
}

// phaseFiveClear is THE PHASE-5 CLEAR, Option-B re-spec: the
// batch-complete mirror push (the ack — every batch, best-effort) plus
// the all-clear verdict speech (gated exactly like the reference:
// success AND bot_source).
func (e *batchEngine) phaseFiveClear(botSource, success bool) {
	// The ack rides the batch-complete moment — unconditional in the
	// finish chain (the reference's on_batch_complete law).
	e.pushVaultMirrorAfterBatch()

	// The verdict speech — the reference's `if success and bot_source`
	// gate.
	if !success || !botSource {
		return
	}
	if e.tracker == nil {
		e.log("⏸️ No link manifest for this batch (empty vault path?) — "+
			"the Worker queue is not acknowledged; dedup still protects "+
			"against duplicates.", "warning")
		return
	}
	if e.tracker.GetAllClear() {
		e.log("✓ Bot queue synced — all links verified (Phase 5 CLEAR; "+
			"the vault-mirror push is the ack).", "success")
		return
	}
	unverified := 0
	if e.tracker.Manifest != nil {
		for _, l := range e.tracker.Manifest.Links {
			if l.Type == "github" &&
				(l.Status == "failed" || l.Status == "processing" ||
					l.Status == "pending") {
				unverified++
			}
		}
	}
	e.log(fmt.Sprintf(
		"⏸️ %d link(s) not verified — they stay PENDING in the Worker queue "+
			"and are not acknowledged; the next --auto re-fetches and retries "+
			"them.", unverified), "warning")
}

// ---------------------------------------------------------------------------
// The model-changed law (the runner's in-place mutation → the live
// config + the save).
// ---------------------------------------------------------------------------

// onModelChanged mirrors the model_changed handler: the runner's
// in-place mutation is mirrored onto the LIVE config map (the
// reference's ONE shared dict — the Go runner holds a plain-map copy,
// so the write is synced here before the save), the switch line, and
// the config save (config.Save's atomic + mirror law).
func (e *batchEngine) onModelChanged(provider, model string) {
	switch provider {
	case "cloud":
		e.cfg.Set("cloud_model", model)
	case "llamacpp":
		e.cfg.Set("llamacpp_model", model)
	default:
		if om := getMapCfg(e.cfg, "ollama"); om != nil {
			om.Set("model", model)
		} else {
			om := storage.NewOrderedMap()
			om.Set("model", model)
			e.cfg.Set("ollama", om)
		}
	}
	e.log(fmt.Sprintf(
		"LLM model switched to '%s' (%s) — saving config.", model, provider), "info")
	if e.a.WriteLiveConfig == nil {
		return
	}
	if !e.a.WriteLiveConfig(e.cfg) {
		e.log("Could not save config: the write failed", "warning")
	}
}

// applyModelChoiceInPlace mirrors WorkerLlmMixin._apply_model_choice:
// mutate the config IN PLACE (never rebind) + emit model_changed.
func (e *batchEngine) applyModelChoiceInPlace(newModel string) {
	if newModel == "" {
		return
	}
	provider := strings.ToLower(strings.TrimSpace(pyStrCfg(e.cfg, "llm_provider")))
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

// ---------------------------------------------------------------------------
// The warmup family (processing_worker.py 470-748).
// ---------------------------------------------------------------------------

// warmupFamily mirrors the provider ladder. Returns true when the
// batch must abort (finish has already fired).
func (e *batchEngine) warmupFamily() bool {
	ollamaBase := e.nestedStrOr("ollama", "base_url", "http://localhost:11434")
	ollamaModel := e.nestedStrOr("ollama", "model", "qwythos-9b")
	llmProvider := strings.ToLower(strings.TrimSpace(pyStrCfg(e.cfg, "llm_provider")))
	if llmProvider == "" {
		llmProvider = "ollama"
	}

	switch llmProvider {
	case "ollama":
		client := e.a.ollamaFactoryDoor(ollamaBase)
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
		llamaURL := pyStrCfg(e.cfg, "llamacpp_api_url")
		if llamaURL == "" {
			llamaURL = llmclient.LLamaCppDefaultBase
		}
		llamaKey := pyStrCfg(e.cfg, "llamacpp_api_key")
		probe := e.a.probeLlamaDoor(llamaURL, llamaKey, llmclient.LLamaCppProbeTimeoutS)
		if !probe.Found {
			scan := e.a.detectLlamaDoor(llamaKey)
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
		configured := strings.TrimSpace(pyStrCfg(e.cfg, "llamacpp_model"))
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
		cloudKey := pyStrCfg(e.cfg, "cloud_api_key")
		flavor := "OpenAI-compatible"
		if llmclient.IsAnthropicURL(cloudURL) {
			flavor = "Claude"
		}
		e.log(fmt.Sprintf("☁️ Using Cloud API (%s): %s / model '%s'",
			flavor, cloudURL, cloudModel), "info")
		modelsCfg := getMapCfg(e.cfg, "models")
		type modelCheck struct{ label, m string }
		checks := []modelCheck{{"model", strings.TrimSpace(cloudModel)}}
		for _, t := range []string{"classify", "analyze"} {
			if raw := getOrCfg(modelsCfg, t, ""); raw != nil {
				if m, isStr := raw.(string); isStr && strings.TrimSpace(m) != "" {
					checks = append(checks, modelCheck{"models." + t, strings.TrimSpace(m)})
				}
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
			listDoor := e.a.cloudListModelsDoor()
			names, err := listDoor(cloudURL, cloudKey, 15)
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
	pid, spawnErr := e.a.autostartServeDoor()
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
// single-model auto-switch, the model-prompt door (the shell's picker
// when one answers; the reference's console menu), the best-stand-in
// auto-pick.
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
			if e.ev.ModelPrompt != nil {
				func() {
					defer func() { recover() }()
					choice = e.ev.ModelPrompt(ollamaModel, available)
				}()
			}
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
				// No picker / the user declined / the stand-in failed: auto-pick.
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
// The inputs fetchers (the import file — inputs.py 152-232).
// ---------------------------------------------------------------------------

// fetchFromImport mirrors _fetch_from_import: the one-link grammar +
// the inbox notes + the intake counters.
func (e *batchEngine) fetchFromImport(importFile string) []string {
	if importFile == "" || !fileExistsAt(importFile) {
		e.log("Import file not found.", "error")
		return nil
	}
	text, err := links.ReadImportFile(importFile)
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
		if len(parsed.Unparsed) > 20 {
			for _, line := range parsed.Unparsed[:20] {
				e.log("     · "+pySliceRunes(line, 0, 100), "warning")
			}
			e.log(fmt.Sprintf("     … and %d more", len(parsed.Unparsed)-20), "warning")
		} else {
			for _, line := range parsed.Unparsed {
				e.log("     · "+pySliceRunes(line, 0, 100), "warning")
			}
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

// createInboxNotes mirrors _create_inbox_notes: bypassed (with the
// ON line) when the Websites pipeline is ON.
func (e *batchEngine) createInboxNotes(nonGithubURLs []string, source string) {
	pipes := getMapCfg(e.cfg, "pipelines")
	if pyTruthyCfg(getOrCfg(pipes, "websites", false)) {
		e.log(fmt.Sprintf("🌐 Websites pipeline ON — %d link(s) will be "+
			"processed as websites (not written to _inbox).", len(nonGithubURLs)), "info")
		return
	}
	cfgPlain := cfgPlainMap(e.cfg)
	platformintake.WriteInboxLinksByPlatform(
		platformintake.InboxTableVault(cfgPlain), nonGithubURLs, source,
		e.log, links.BlockedDomainsFromConfig(cfgPlain), nil)
}

// setNonGithub stores the backlog modes' non-GitHub list.
func (e *batchEngine) setNonGithub(urls []string) {
	e.nonGithub = urls
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
// The runner construction (the GUI shape).
// ---------------------------------------------------------------------------

// buildRunner constructs the Runner with the engine's event wiring.
func (e *batchEngine) buildRunner(noteStateDB *notestate.DB,
	tracker *batch.LinkTracker, total int, urls []string) *githubpipeline.Runner {

	r := githubpipeline.NewRunner(cfgPlainMap(e.cfg))
	r.Start() // is_running = True (the __init__ law)
	// A stop that arrived during the warmup reaches the runner now (the
	// lost-stop window closes — requestStop's own law).
	if e.stopRequested {
		r.Stop()
	}
	r.Cache = e.cache
	r.NoteStateDB = noteStateDB
	r.VaultIndex = e.vaultIndex
	r.LinkTracker = tracker
	r.Total = total
	r.NonGithubURLs = e.nonGithub
	r.IntakeDuplicates = e.intakeDuplicates
	r.RawURLCount = e.rawURLCount
	r.BotSource = false
	if e.optsBotSource() {
		r.BotSource = true
	}
	r.LLMProvider = e.warmProvider
	r.OllamaClient = e.warmOllama
	r.Model = e.warmModel
	if e.a.GHClientFactory != nil {
		r.NewGHClient = e.a.GHClientFactory
	}
	if e.a.BannerDownload != nil {
		r.BannerFn = e.a.BannerDownload
	}
	r.ModelPromptFn = e.ev.ModelPrompt
	r.Log = e.ev.Log
	r.OnStatus = e.ev.Status
	r.OnProgress = e.ev.Progress
	r.OnDiskFull = e.ev.DiskFull
	r.OnModelChanged = e.onModelChanged
	if e.ev.LLMDecision != nil {
		r.LLMDecisionFn = func(repoName string, attempt int, err error,
			client llmclient.OllamaAPI, model string) string {
			return e.ev.LLMDecision(repoName)
		}
	}
	r.Finished = func(ok bool, msg string) {
		e.finish(ok, msg)
	}
	r.HygienePass = e.hygienePass
	return r
}

// opts carries the batch's own options through the runner callbacks
// (finish needs the dry-run + source flags after the runner holds the
// stage).
func (e *batchEngine) optsBotSource() bool { return e.opts.Source == "bot" }

// hygienePass wires the vaulthygiene module (T-46): the batch's own
// closing pass runs run_vault_hygiene over the GITHUB vault ONLY.
// Best-effort — never a killer.
func (e *batchEngine) hygienePass() {
	func() {
		defer func() { recover() }()
		vaulthygiene.RunVaultHygiene(pyStrCfg(e.cfg, "vault_path"),
			e.log, nil, nil)
	}()
}

// runStartCheckGithub mirrors the GitHub vault's §4.4 pass.
func (e *batchEngine) runStartCheckGithub(vaultPath string,
	db *notestate.DB) *notestate.RunResult {
	res, err := notestate.RunStartCheck(notestate.VaultGithub, vaultPath,
		db, e.log, nil, cacheDBPathOf(e.cfg))
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

// ---------------------------------------------------------------------------
// The websites phase (worker/website_phase.py 95-515).
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
		pipes := getMapCfg(e.cfg, "pipelines")
		if !pyTruthyCfg(getOrCfg(pipes, "websites", false)) {
			return nil
		}
		websiteVault := strings.TrimSpace(pyStrCfg(e.cfg, "website_vault_path"))
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
			statePath = dryrun.ShadowCachePath(cacheDBPathOf(e.cfg))
		} else {
			statePath = cacheDBPathOf(e.cfg)
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
		llmProvider := strings.ToLower(strings.TrimSpace(pyStrCfg(e.cfg, "llm_provider")))
		if llmProvider == "" {
			llmProvider = "ollama"
		}
		numCtx := pyIntDefaultCfg(e.cfg, "llm_num_ctx", llmclient.DefaultNumCtx)
		warn := func(m string) { e.log(m, "warning") }
		llmCall := func(messages []llmclient.Message, task string) string {
			timeoutS := pyFloatCfgOr(e.cfg, "llm_timeout_s", 300)
			switch llmProvider {
			case "cloud":
				model := llmclient.ResolveTaskModel(e.cfg, task,
					pyStrCfg(e.cfg, "cloud_model"))
				out, _ := llmclient.CloudChat(pyStrCfg(e.cfg, "cloud_api_url"),
					pyStrCfg(e.cfg, "cloud_api_key"), model, messages, timeoutS,
					&llmclient.ChatOptions{JSONMode: true, NumCtx: numCtx,
						OnWarn: warn})
				return out
			case "llamacpp":
				model := llmclient.ResolveTaskModel(e.cfg, task,
					pyStrCfg(e.cfg, "llamacpp_model"))
				out, _ := llmclient.CloudChat(
					llmclient.NormalizeLlamaCppAPIURL(pyStrCfg(e.cfg, "llamacpp_api_url")),
					pyStrCfg(e.cfg, "llamacpp_api_key"), model, messages, timeoutS,
					&llmclient.ChatOptions{JSONMode: true, NumCtx: numCtx,
						OnWarn: warn})
				return out
			default:
				model := llmclient.ResolveTaskModel(e.cfg, task, ollamaModel)
				out, _ := llmclient.OllamaChat(ollamaClient, model, messages,
					timeoutS, &llmclient.OllamaChatOptions{NumCtx: numCtx, OnWarn: warn})
				return out
			}
		}

		// v0.60.0 — the banish gate's channel (the Telegram round-trip;
		// nil when not paired/enabled — the defer default).
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

// repoBanishFactory mirrors _repo_banish_factory: the channel is
// built ONLY when the scan found marks — a clean vault never pays the
// Worker handshake. (nil, nil) is the reference's
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

// ---------------------------------------------------------------------------
// The seam doors (nil-safe — the CLI's door law on the App).
// ---------------------------------------------------------------------------

func (a *App) ollamaFactoryDoor(host string) llmclient.OllamaAPI {
	if a.OllamaFactory != nil {
		return a.OllamaFactory(host)
	}
	return llmclient.NewOllamaClient(host)
}

func (a *App) probeLlamaDoor(baseURL, apiKey string, timeoutS float64) *llmclient.ProbeResult {
	if a.ProbeLlama != nil {
		return a.ProbeLlama(baseURL, apiKey, timeoutS)
	}
	return llmclient.ProbeLlamaCpp(baseURL, apiKey, timeoutS)
}

func (a *App) detectLlamaDoor(apiKey string) *llmclient.ProbeResult {
	if a.DetectLlama != nil {
		return a.DetectLlama(apiKey)
	}
	return llmclient.DetectLlamaCpp(apiKey, llmclient.LLamaCppScanPorts,
		"127.0.0.1", llmclient.LLamaCppProbeTimeoutS, true)
}

func (a *App) autostartServeDoor() (int, string) {
	if a.AutostartServe != nil {
		return a.AutostartServe()
	}
	return spawnOllamaServe()
}

func (a *App) cloudListModelsDoor() func(apiURL, apiKey string, timeoutS float64) ([]string, error) {
	if a.CloudListModels != nil {
		return a.CloudListModels
	}
	return func(apiURL, apiKey string, timeoutS float64) ([]string, error) {
		if llmclient.IsAnthropicURL(apiURL) {
			return llmclient.AnthropicListModels(apiURL, apiKey, timeoutS)
		}
		return llmclient.OpenAIListModels(apiURL, apiKey, timeoutS)
	}
}

// ---------------------------------------------------------------------------
// The config-reading helpers (the CLI's own law, the app's copies).
// ---------------------------------------------------------------------------

// getMapCfg reads a nested OrderedMap (nil-safe).
func getMapCfg(cfg *storage.OrderedMap, key string) *storage.OrderedMap {
	if cfg == nil {
		return nil
	}
	v, _ := cfg.Get(key)
	if om, ok := v.(*storage.OrderedMap); ok {
		return om
	}
	return nil
}

// getOrCfg reads a key with a default through any's truth.
func getOrCfg(om *storage.OrderedMap, key string, def any) any {
	if om == nil {
		return def
	}
	v, ok := om.Get(key)
	if !ok || v == nil {
		return def
	}
	return v
}

// pyTruthyCfg mirrors the reference's truthiness law.
func pyTruthyCfg(v any) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != "" && t != "0" && t != "false" && t != "False"
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	}
	return false
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
	sub := getMapCfg(e.cfg, section)
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
	}
	return 0
}

// maxFloat mirrors max(a, b).
func maxFloat(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// pySliceRunes mirrors Python's s[:n] over runes.
func pySliceRunes(s string, start, end int) string {
	runes := []rune(s)
	if start < 0 {
		start = 0
	}
	if end > len(runes) {
		end = len(runes)
	}
	if start >= end {
		return ""
	}
	return string(runes[start:end])
}
