// commands.go — T-71 slice (b): the command surface (doc 04 §5's
// table) — every Python worker job becomes one command method: takes
// a context, emits events on the bus, returns a result struct, and
// NEVER panics across the boundary (the recover + honest error event
// mirrors the BaseException-safe TestWorker wrapper,
// processing_worker.py:2093-2113).
//
// THE HERMETIC LAW: every engine dependency is an injected seam on
// Deps — the DI root (app.go, slice (d)) wires the production
// defaults (the real engine functions); the contract tests inject
// fakes and assert the TRANSPORT laws: the events, the ask-gates, the
// one-batch lock, the result shapes, the crash safety. A nil seam is
// a wiring bug — the commands refuse honestly, never improvise past
// a missing engine.
//
// T-71's scope (the mandate's table): Sync.FetchQueue, Sync.RunBatch,
// Sync.FastSync, Sync.DriftSeal, Scan.Run, Scan.ApplyPlan,
// Backup.SealVault/Resync/Now/Publish, Radar.Detect. The rest of doc
// 04 §5's table (SyncView.Measure, Connection.Test, LLM.*,
// Retry.*, HandDelivery, Maintenance.*) is later GUI-shell work.
package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/goodrepos"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaulthygiene"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
	"github.com/assadigit/second-brain-project/internal/vaultseal"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// LogFn mirrors the engine packages' optional log callback shape:
// log(msg, level).
type LogFn = func(msg, level string)

// ---------------------------------------------------------------------------
// The event adapters (the bus, seen through the engine seams' eyes).
// ---------------------------------------------------------------------------

// BatchEvents is the batch engine's event surface (the Runner's own
// callback fields, wired to the bus by the command).
type BatchEvents struct {
	Log         func(msg, level string)
	Status      func(url string)
	Progress    func(pos, total int)
	DiskFull    func(path string)
	ModelChange func(provider, model string)
	// LLMDecision is the 600 s failure door (Runner.LLMDecisionFn);
	// ModelPrompt is the CLI console model-picker (Runner.ModelPromptFn).
	LLMDecision func(repo string) string
	ModelPrompt func(configured string, available []string) string
}

// ScanEvents is the scan job's event surface (_vault_scan_job's
// injected callables): the desk log, the honest pipeline row, and the
// confirm door (the bus's AskScanConfirm gate).
type ScanEvents struct {
	Log      func(msg, level string)
	Progress func(label string, value, max int)
	Confirm  func(plan any) ScanAnswer
}

// ---------------------------------------------------------------------------
// The result structs (the jobs' result dicts — the GUI fallbacks and
// tests read these shapes).
// ---------------------------------------------------------------------------

// QueueReport is the _bot_queue_job result: the fetch's state word,
// the routed URL lists, and the classification counts (the ~12
// fields the sync display's fallbacks read).
type QueueReport struct {
	Success bool
	State   string // ok | disabled | error
	Error   string

	GitHubURLs []string
	NonGithub  []string

	PendingURLs     []string
	PendingWebsites []string

	// SettledList is the settled-repos ledger (the sync display's
	// dedupe reads it).
	SettledList map[string]bool

	InVault        int
	Decommissioned int
	Blocked        int
	Self           int
	Settled        int
	VaultIndex     *int

	// WebsitesVaultMissing marks the configured-but-missing websites
	// vault (every non-GitHub link reads as pending).
	WebsitesVaultMissing bool

	WebsitesInVault   int
	WebsitesProcessed int
	WebsitesDismissed int
	WebsitesBlocked   int
	WebsitesSelf      int
	WebsitesSettled   int
	WebsitesOff       bool
	WebsitesNoVault   bool

	RawCount int

	// DuplicatesRemoved is the fetch-side dedupe count (the Telethon
	// era's counter; the Worker pending list is unique by url — the
	// field stays zero under Option B, the render's door remains).
	DuplicatesRemoved int

	// WebsitesVaultIndex is the websites vault's note count (the 📚
	// line's own number, recorded by the classification half).
	WebsitesVaultIndex *int
}

// BatchOptions is the shell's intake facts (the worker's constructor
// args + the v23/v25 pre-start injections).
type BatchOptions struct {
	// URLs is the batch's GitHub link list (already routed).
	URLs []string
	// Source is the intake source label: "bot" | "direct" | "import".
	Source string
	// Mode is the engine's dispatch mode: "direct" (the default) |
	// "import" | "review_retry" | "master_retry" (the T-74 wiring's
	// re-host; the Sync view's retry doors ride the retry modes).
	Mode string
	// ImportFile is the import-mode file path (the inputs fetcher's
	// subject — the Input page's PROCESS door hands it here).
	ImportFile string
	// NonGithub is the non-GitHub intake (inbox notes / the websites
	// phase's list).
	NonGithub []string
	// IntakeDuplicates / RawURLCount are the dedup stats the final
	// report carries (the bot queue's pre-drop).
	IntakeDuplicates int
	RawURLCount      int
	// DryRun flips the engine's dry-run law (every vault write logged,
	// not performed).
	DryRun bool
}

// BatchResult is the batch's finished_signal payload + the counts
// the hero/scorecard read. FailedLinks (T-77) is the tracker's failed
// count at finish — the retry chime's selector (the reference's
// summary['failed_links']). Natural/Stopped/WebsiteResults (T-79f)
// carry the reference's batch_summary facts to the finish crossing:
// Natural marks the batch that reached the end-of-batch stage
// (batch_summary's PRESENCE — the early returns never set it), Stopped
// is summary['stopped'] (the user's cooperative stop — no fanfare on a
// deliberate stop), WebsiteResults is _website_summary['results']
// (the end-of-run offer's data).
type BatchResult struct {
	Success     bool
	Message     string
	Processed   int
	Total       int
	FailedLinks int
	// Crashed marks the recover path (the TestWorker's
	// {'success': False, 'error': …} twin).
	Crashed bool
	Error   string

	Natural        bool
	Stopped        bool
	WebsiteResults []githubpipeline.WebsiteResultEntry
}

// ScanStory is one vault's scan story (the reference's per-vault story
// dict): the planned counts — spoken in PICKED units once a confirmed
// selection rides (the 'K of N' line never reads an unticked row as a
// failure) — the landed counts, and the failed moves BY NAME (v0.82.0's
// honesty ledger).
type ScanStory struct {
	Verdict            string                 `json:"verdict"`
	Deletions          int                    `json:"deletions"`
	Moves              int                    `json:"moves"`
	NewFolders         int                    `json:"new_folders"`
	Consolidations     int                    `json:"consolidations"`
	ConsolidationNotes int                    `json:"consolidation_notes"`
	Applied            int                    `json:"applied"`
	Moved              int                    `json:"moved"`
	FoldersCreated     int                    `json:"folders_created"`
	NotesConsolidated  int                    `json:"notes_consolidated"`
	FailedMoves        int                    `json:"failed_moves"`
	Failed             []vaultscan.FailedMove `json:"failed"`
	Vault              string                 `json:"vault"`
	VaultKind          string                 `json:"vault_kind"`
	VaultLabel         string                 `json:"vault_label"`
}

// ScanResult is the _vault_scan_job story dict: the hygiene pass, the
// per-scope verdict, and the closing report lines — plus the full story
// shape the scorecard reads (the reference's out dict: the top-level
// story belongs to the Websites vault, the repos sub-story rides Repos,
// and a repos-only scan promotes its story to the top).
type ScanResult struct {
	Success      bool
	Verdict      string // clean | confirmed | declined | timeout | error
	HygieneTotal int
	// Hygiene is the per-vault heal counts ('websites' / 'github'
	// keys — the scorecard's 🧹 rows read these).
	Hygiene     map[string]vaulthygiene.HygieneCounts `json:"hygiene"`
	WebsitesRan bool
	ReposRan    bool
	Lines       []string
	Crashed     bool
	Error       string

	// The story shape (v0.67/0.70/0.83 — the applied scorecard's
	// own data). Story is the TOP-LEVEL vault's story (the Websites
	// vault when both ride; the promoted repos story when repos
	// rides alone); Repos is the repos sub-story (nil when the
	// repos leg never ran).
	Story       *ScanStory `json:"story"`
	Repos       *ScanStory `json:"repos"`
	VaultsRan   []string   `json:"vaults"`
	Vault       string     `json:"vault"`
	GitHubVault string     `json:"github_vault"`
	VaultKind   string     `json:"vault_kind"`
	VaultLabel  string     `json:"vault_label"`
}

// ApplyResult is the apply door's report (apply_scan_plan's verdict,
// mapped over the vaultscan.ApplyReport shape the command returns).
type ApplyResult struct {
	Success bool
	Error   string
	Report  *vaultscan.ApplyReport
}

// SealResultRow is one Backup.* starter's outcome (the done(bool,
// str) signal every seal/publish/backup QThread emits).
type SealResultRow struct {
	Success bool
	Message string
	// Level is the log level the command line lands at (success |
	// warning | error).
	Level string
}

// RadarResult is the _repo_radar_job story: the detections, how many
// landed in the config copy, and the report lines.
type RadarResult struct {
	Success bool
	Applied int
	Results map[string]any
	Lines   []string
	Error   string
}

// ---------------------------------------------------------------------------
// The command surface.
// ---------------------------------------------------------------------------

// Deps carries the engine seams (the hermetic law — see the file doc).
// Every field is optional at construction; a nil seam called by a
// command refuses honestly.
type Deps struct {
	// Config yields the live config (nil → an empty OrderedMap).
	Config func() *storage.OrderedMap

	// QueueFetch is the Worker /api/pending door (workerclient.Sync's
	// honest poll): (state, items) — state "ok" | "disabled" |
	// "error".
	QueueFetch func(cfg *storage.OrderedMap) (string, []any)
	// QueueClassify is the classification half (the vault filter +
	// the settlement + the websites ladder).
	QueueClassify func(cfg *storage.OrderedMap, items []any, log LogFn) *QueueReport

	// Batch is the batch engine composition (the DI root's wiring;
	// T-71's contract is the transport, T-72 binds the full GUI
	// composition).
	Batch func(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult

	// DriftSeal is the drift-seal engine (vaultseal.RunDriftSeal's
	// shape): returns (line, level) — ("", "info") is the quiet
	// automatic path.
	DriftSeal func(cfg *storage.OrderedMap, manual bool, log LogFn) (string, string)

	// Scan is the vault-scan job engine (heal + scope + plan + ask +
	// apply + verdict).
	Scan func(ctx context.Context, cfg *storage.OrderedMap, vaults []string, ev ScanEvents) *ScanResult

	// ApplyPlan is the apply door (apply_scan_plan over the config's
	// vault + DBs).
	ApplyPlan func(cfg *storage.OrderedMap, kind string, plan *vaultscan.Plan, selection map[string][]string, log LogFn) *ApplyResult

	// SealVault is the seal starter for kind repos|websites|personal
	// (vaultseal's Seal/WebsitesSeal/ManualSealFromConfig shape).
	SealVault func(cfg *storage.OrderedMap, kind string, log LogFn) *vaultseal.SealResult
	// Resync is the mirror re-sync (vaultseal.ResyncFromConfig).
	Resync func(cfg *storage.OrderedMap, log LogFn) *storage.OrderedMap
	// BackupNow is the vault zip job (the BackupWorker port).
	BackupNow func(cfg *storage.OrderedMap, log LogFn) *SealResultRow
	// Publish is the GoodRepos publish (goodrepos.PublishFromConfig).
	Publish func(cfg *storage.OrderedMap, log LogFn) *goodrepos.PublishResult

	// Radar is the repo-radar job (reporadar's composition).
	Radar func(cfg *storage.OrderedMap, log LogFn) *RadarResult
}

// Commands is the command surface both shells drive (doc 04 §5).
type Commands struct {
	Bus  *Bus
	Deps Deps

	// batchLock enforces one-batch-at-a-time (doc 04 §2's
	// TelegramLockManager law port — owner-scoped, watchdog-ready).
	batchLock OpLock
	// running is the _batch_running flag (the GUI mirror's signal).
	running bool
	// stopDoor is the active batch's cooperative-stop door, wired by
	// the App around each run (nil = no batch to stop).
	stopDoor func() bool
}

// NewCommands builds the surface over the bus + the injected deps.
func NewCommands(bus *Bus, deps Deps) *Commands {
	return &Commands{Bus: bus, Deps: deps}
}

// cfg snapshots the live config (nil-safe).
func (c *Commands) cfg() *storage.OrderedMap {
	if c.Deps.Config == nil {
		return storage.NewOrderedMap()
	}
	if cfg := c.Deps.Config(); cfg != nil {
		return cfg
	}
	return storage.NewOrderedMap()
}

// BatchBusy answers whether a batch holds the one-batch lock (the
// Scan flow's own guard reads it — the reference's "a batch is
// running" check).
func (c *Commands) BatchBusy() bool {
	return c.batchLock.Busy()
}

// guardCommand is the BaseException-safe wrapper: the seam runs
// recovered; a crash lands as the honest error line + the error
// result, never a panic across the boundary.
func (c *Commands) guardCommand(name string, fn func()) (crashed bool, crashMsg string) {
	defer func() {
		if r := recover(); r != nil {
			crashed = true
			crashMsg = fmt.Sprintf("%v", r)
			c.Bus.EmitLogLine(fmt.Sprintf("💥 Command '%s' crashed: %s — recovered, nothing left running", name, crashMsg), "error")
		}
	}()
	fn()
	return false, ""
}

// notWired is the honest refusal for a nil seam.
func notWired(name string) string {
	return "the " + name + " engine seam is not wired (a DI bug — nothing ran)"
}

// ---------------------------------------------------------------------------
// Sync.* — the queue door + the batch + the drift seals.
// ---------------------------------------------------------------------------

// SyncFetchQueue mirrors _bot_queue_job's fetch + classify halves:
// the Worker queue poll, the vault filtering, the settlement, and
// the websites ladder — the story lands on the bus, the counts in
// the report.
func (c *Commands) SyncFetchQueue(ctx context.Context) *QueueReport {
	cfg := c.cfg()
	if c.Deps.QueueFetch == nil {
		return &QueueReport{State: "error", Error: notWired("queue fetch")}
	}
	var out *QueueReport
	if crashed, _ := c.guardCommand("Sync.FetchQueue", func() {
		state, items := c.Deps.QueueFetch(cfg)
		if state != "ok" {
			errMsg := "the Worker queue poll failed (network down or the Worker answered an error)"
			if state == "disabled" {
				// THE DIAGNOSTIC LAW: check WHICH key is missing so the
				// owner sees exactly what to fix (not just "not configured").
				syncCheck := workerclient.NewSync(cfgPlainMap(cfg))
				reason := syncCheck.DisabledReason()
				errMsg = "the Worker queue is not configured — " + reason +
					". Open Settings → Bot → Direct Pair (advanced) and paste " +
					"the Worker URL + install_id + shared_secret, OR edit " +
					"config.json directly."
			} else if state == "error" {
				// THE D1-FALLBACK DIAGNOSTIC (SESSION-081): the direct
				// Worker request failed (the owner's ISP blocks workers.dev
				// at the IP level); tell the owner whether the D1-API
				// fallback is ARMED (the fallback ran + failed) or DORMANT
				// (the 3 Cloudflare-API creds are missing — the fallback
				// was skipped). Never a bare "queue poll failed".
				syncCheck := workerclient.NewSync(cfgPlainMap(cfg))
				errMsg = errMsg + " — " + syncCheck.D1FallbackStatus()
			}
			out = &QueueReport{State: state, Error: errMsg}
			return
		}
		if c.Deps.QueueClassify == nil {
			out = &QueueReport{State: "error", Error: notWired("queue classify")}
			return
		}
		if rep := c.Deps.QueueClassify(cfg, items, func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		}); rep != nil {
			rep.State = "ok"
			rep.Success = true
			out = rep
		} else {
			out = &QueueReport{State: "error", Error: "the classification returned no report"}
		}
	}); crashed {
		return &QueueReport{State: "error", Error: "the queue fetch crashed — see the desk log"}
	}
	return out
}

// SyncRunBatch mirrors the batch start (_start_worker_with_urls → the
// worker): ONE batch at a time (the op lock + running flag — the
// refusal is honest, never a queue), the engine's events flow to the
// bus, and BatchFinished lands EXACTLY ONCE whatever happens (the
// lock-leak fix's own law).
func (c *Commands) SyncRunBatch(ctx context.Context, opts BatchOptions) *BatchResult {
	if !c.batchLock.Acquire("batch") {
		msg := fmt.Sprintf("⏳ A batch is already running (%s) — one batch at a time", c.batchLock.Describe())
		c.Bus.EmitLogLine(msg, "warning")
		return &BatchResult{Success: false, Message: msg}
	}
	c.running = true
	finished := false
	released := false
	defer func() {
		// the lock-leak fix: the finished signal fires exactly once,
		// and the lock always goes home (owner-scoped).
		if !finished {
			c.Bus.EmitBatchFinished(&BatchResult{Success: false,
				Message: "the batch ended without a verdict (crashed or cancelled)"})
		}
		if !released {
			c.running = false
			c.batchLock.Release("batch")
		}
	}()

	if c.Deps.Batch == nil {
		return &BatchResult{Success: false, Message: notWired("batch")}
	}

	c.Bus.EmitStatus("Processing…")
	c.Bus.EmitProgress(0, len(opts.URLs))

	var res *BatchResult
	if crashed, _ := c.guardCommand("Sync.RunBatch", func() {
		res = c.Deps.Batch(ctx, c.cfg(), opts, BatchEvents{
			Log:      func(msg, level string) { c.Bus.EmitLogLine(msg, level) },
			Status:   func(url string) { c.Bus.EmitStatus(url) },
			Progress: func(pos, total int) { c.Bus.EmitProgress(pos, total) },
			DiskFull: func(path string) { c.Bus.EmitDiskFull(path) },
			ModelChange: func(provider, model string) {
				c.Bus.EmitModelChanged(provider, model)
			},
			LLMDecision: func(repo string) string { return c.Bus.LLMFailed(repo) },
			ModelPrompt: nil, // the CLI console picker is the CLI shell's wiring
		})
	}); crashed {
		res = &BatchResult{Success: false, Message: "the batch crashed — see the desk log", Crashed: true, Error: "recovered panic"}
	}
	if res == nil {
		res = &BatchResult{Success: false, Message: "the batch engine returned no result"}
	}
	finished = true
	released = true
	// T-79f — THE FLAG-CLEAR LAW (the reference's processing_finished
	// opening: `self._batch_running = False` runs at the TOP of the
	// finish handler, BEFORE the fanfare chain): the running flag + the
	// lock go home BEFORE the finished signal crosses, so the chain
	// riding the crossing (the redo pass's own re-run batch, the
	// offer's answers) never sees a ghost batch.
	c.running = false
	c.batchLock.Release("batch")
	c.Bus.EmitBatchFinished(res)
	return res
}

// StopBatch is stop_processing's engine half (v0.32): the cooperative,
// between-items stop — the active batch's flag flips, the current item
// finishes, the finish chain still runs, and BatchFinished lands
// exactly once (the lock-leak fix's law). Returns false when no batch
// is running (an honest no-op, never an error).
func (c *Commands) StopBatch() bool {
	if c.stopDoor == nil {
		return false
	}
	return c.stopDoor()
}

// SyncFastSync is the owner's FAST SYNC (the manual drift-seal door,
// v0.90.0): the triad + the directory publish, and the verdict ALWAYS
// speaks — even a clean vault gets its "already in sync" line.
func (c *Commands) SyncFastSync(ctx context.Context) *SealResultRow {
	return c.runDriftSeal(ctx, true)
}

// SyncDriftSeal is the AUTOMATIC drift seal (the batch-end path's own
// legs, v0.89.0): the quiet law — a leg that seals nothing and fails
// nothing stays silent.
func (c *Commands) SyncDriftSeal(ctx context.Context) *SealResultRow {
	return c.runDriftSeal(ctx, false)
}

func (c *Commands) runDriftSeal(ctx context.Context, manual bool) *SealResultRow {
	name := "Sync.DriftSeal"
	if manual {
		name = "Sync.FastSync"
	}
	if c.Deps.DriftSeal == nil {
		return &SealResultRow{Success: false, Message: notWired("drift seal"), Level: "error"}
	}
	var line, level string
	if crashed, _ := c.guardCommand(name, func() {
		line, level = c.Deps.DriftSeal(c.cfg(), manual, func(msg, lvl string) {
			c.Bus.EmitLogLine(msg, lvl)
		})
	}); crashed {
		return &SealResultRow{Success: false, Message: "the drift seal crashed — see the desk log", Level: "error"}
	}
	// ("", "info") is the quiet automatic path — nothing to say.
	if line == "" && !manual {
		return &SealResultRow{Success: true, Message: "", Level: "info"}
	}
	if line == "" && manual {
		// the manual door's own law: an explicit press ALWAYS gets its
		// verdict (the engine's clean-triad line lands here as empty
		// only if the seam under-spoke — refuse honestly).
		return &SealResultRow{Success: false, Message: "the fast sync returned no verdict", Level: "error"}
	}
	if level == "" {
		level = "info"
	}
	c.Bus.EmitLogLine(line, level)
	return &SealResultRow{Success: level == "success", Message: line, Level: level}
}

// ---------------------------------------------------------------------------
// Scan.* — the vault scan + the apply door.
// ---------------------------------------------------------------------------

// ScanRun mirrors _vault_scan_job: the hygiene pass over BOTH machine
// vaults, the scope picker, the per-vault plan, the ask (the bus's
// confirm gate), and the merged verdict — nothing moves until the
// owner says so.
func (c *Commands) ScanRun(ctx context.Context, vaults []string) *ScanResult {
	if c.Deps.Scan == nil {
		return &ScanResult{Success: false, Verdict: "error", Error: notWired("scan")}
	}
	var res *ScanResult
	if crashed, _ := c.guardCommand("Scan.Run", func() {
		res = c.Deps.Scan(ctx, c.cfg(), vaults, ScanEvents{
			Log: func(msg, level string) { c.Bus.EmitLogLine(msg, level) },
			Progress: func(label string, value, max int) {
				c.Bus.EmitJobProgress(label, value, max)
			},
			Confirm: func(plan any) ScanAnswer {
				return c.Bus.AskScanConfirm(plan)
			},
		})
	}); crashed {
		return &ScanResult{Success: false, Verdict: "error", Crashed: true, Error: "the scan crashed — see the desk log"}
	}
	if res == nil {
		res = &ScanResult{Success: false, Verdict: "error", Error: "the scan engine returned no result"}
	}
	return res
}

// ScanApplyPlan is the apply door (the ScanPlanDialog's Apply): the
// picked rows ride the selection; nothing else moves.
func (c *Commands) ScanApplyPlan(ctx context.Context, kind string, plan *vaultscan.Plan, selection map[string][]string) *ApplyResult {
	if c.Deps.ApplyPlan == nil {
		return &ApplyResult{Success: false, Error: notWired("apply plan")}
	}
	var res *ApplyResult
	if crashed, _ := c.guardCommand("Scan.ApplyPlan", func() {
		res = c.Deps.ApplyPlan(c.cfg(), kind, plan, selection, func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		})
	}); crashed {
		return &ApplyResult{Success: false, Error: "the apply crashed — see the desk log"}
	}
	if res == nil {
		res = &ApplyResult{Success: false, Error: "the apply engine returned no result"}
	}
	return res
}

// ---------------------------------------------------------------------------
// Backup.* — the seal starters + the mirror resync + the zip + the
// GoodRepos publish.
// ---------------------------------------------------------------------------

// The seal kinds (the drift legs' own vocabulary).
const (
	SealRepos    = "repos"
	SealWebsites = "websites"
	SealPersonal = "personal"
)

// BackupSealVault mirrors the VaultSealWorker starters: kind
// repos|websites|personal, sealed in the background, the verdict in
// the log (never the GUI thread).
func (c *Commands) BackupSealVault(ctx context.Context, kind string) *SealResultRow {
	if kind != SealRepos && kind != SealWebsites && kind != SealPersonal {
		return &SealResultRow{Success: false, Message: "unknown seal kind '" + kind + "' (repos | websites | personal)", Level: "error"}
	}
	if c.Deps.SealVault == nil {
		return &SealResultRow{Success: false, Message: notWired("seal vault"), Level: "error"}
	}
	var res *vaultseal.SealResult
	if crashed, _ := c.guardCommand("Backup.SealVault", func() {
		res = c.Deps.SealVault(c.cfg(), kind, func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		})
	}); crashed {
		return &SealResultRow{Success: false, Message: "the seal crashed — see the desk log", Level: "error"}
	}
	if res == nil {
		return &SealResultRow{Success: false, Message: "the seal engine returned no result", Level: "error"}
	}
	level := "success"
	if !res.OK() {
		level = "warning"
	}
	c.Bus.EmitLogLine(res.Describe(), level)
	return &SealResultRow{Success: res.OK(), Message: res.Describe(), Level: level}
}

// BackupResync mirrors the mirror re-sync starter (the Backup tab's
// Resync button): pull the mirror's state back into the honest map.
func (c *Commands) BackupResync(ctx context.Context) *SealResultRow {
	if c.Deps.Resync == nil {
		return &SealResultRow{Success: false, Message: notWired("resync"), Level: "error"}
	}
	var m *storage.OrderedMap
	if crashed, _ := c.guardCommand("Backup.Resync", func() {
		m = c.Deps.Resync(c.cfg(), func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		})
	}); crashed {
		return &SealResultRow{Success: false, Message: "the resync crashed — see the desk log", Level: "error"}
	}
	if m == nil {
		return &SealResultRow{Success: false, Message: "the resync returned no result", Level: "error"}
	}
	// vaultseal.ResyncFromConfig returns the honest map; the shell
	// renders it. The command's verdict: it ran.
	c.Bus.EmitLogLine("🔄 Mirror re-sync complete — the status card carries the state", "info")
	return &SealResultRow{Success: true, Message: "resync complete", Level: "info"}
}

// BackupNow mirrors _backup_now (the BackupWorker port): the vault
// zip with the exclude rules + the FIFO rotation + the deferred
// ledger rider, in the background, the verdict in the log.
func (c *Commands) BackupNow(ctx context.Context) *SealResultRow {
	if c.Deps.BackupNow == nil {
		return &SealResultRow{Success: false, Message: notWired("backup now"), Level: "error"}
	}
	var res *SealResultRow
	if crashed, _ := c.guardCommand("Backup.Now", func() {
		res = c.Deps.BackupNow(c.cfg(), func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		})
	}); crashed {
		return &SealResultRow{Success: false, Message: "the backup crashed — see the desk log", Level: "error"}
	}
	if res == nil {
		return &SealResultRow{Success: false, Message: "the backup engine returned no result", Level: "error"}
	}
	if res.Message != "" {
		c.Bus.EmitLogLine(res.Message, res.level())
	}
	return res
}

// level maps a row's verdict to the log level (the _backup_result
// lines: success → "💾 Backup created: …" success; failure → "❌
// Backup failed: …" error — the seam's message carries its own glyph).
func (r *SealResultRow) level() string {
	if r.Level != "" {
		return r.Level
	}
	if r.Success {
		return "success"
	}
	return "error"
}

// BackupPublish mirrors _start_goodrepos_publish: the public curated
// directory's publish, best-effort, the verdict in the log.
func (c *Commands) BackupPublish(ctx context.Context) *SealResultRow {
	if c.Deps.Publish == nil {
		return &SealResultRow{Success: false, Message: notWired("publish"), Level: "error"}
	}
	var res *goodrepos.PublishResult
	if crashed, _ := c.guardCommand("Backup.Publish", func() {
		res = c.Deps.Publish(c.cfg(), func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		})
	}); crashed {
		return &SealResultRow{Success: false, Message: "the publish crashed — see the desk log", Level: "error"}
	}
	if res == nil {
		return &SealResultRow{Success: false, Message: "the publish engine returned no result", Level: "error"}
	}
	level := "success"
	if !res.OK() {
		level = "warning"
	}
	c.Bus.EmitLogLine(res.Describe(), level)
	return &SealResultRow{Success: res.OK(), Message: res.Describe(), Level: level}
}

// ---------------------------------------------------------------------------
// Radar.Detect — the backup-repo radar.
// ---------------------------------------------------------------------------

// RadarDetect mirrors _radar_detect → _repo_radar_job: read-only
// detection of the vaults' backup repos over the empty slots, the
// owner's typed names never overwritten, ambiguous cases reported.
func (c *Commands) RadarDetect(ctx context.Context) *RadarResult {
	cfg := c.cfg()
	// The starter's own guards (the Detect button's honest refusals —
	// _radar_detect 944-973): no token, no vaults.
	if strings.TrimSpace(pyStrCfg(cfg, "github_token")) == "" {
		msg := "The repo radar reads your GitHub account to find the repos that already back up the vaults — set the token first (Settings → 🔑 Credentials → GitHub Token)."
		c.Bus.EmitLogLine("📡 "+msg, "warning")
		return &RadarResult{Success: false, Error: msg}
	}
	hasVault := false
	for _, k := range []string{"vault_path", "website_vault_path", "manual_vault_path"} {
		if strings.TrimSpace(pyStrCfg(cfg, k)) != "" {
			hasVault = true
			break
		}
	}
	if !hasVault {
		msg := "Set at least one vault path first (📁 Vault tab) — the radar finds the backup repo FOR a vault."
		c.Bus.EmitLogLine("📡 "+msg, "warning")
		return &RadarResult{Success: false, Error: msg}
	}
	if c.Deps.Radar == nil {
		return &RadarResult{Success: false, Error: notWired("radar")}
	}
	c.Bus.EmitLogLine("📡 Repo radar: scanning your GitHub account for the vaults' existing backup repos…", "info")
	var res *RadarResult
	if crashed, _ := c.guardCommand("Radar.Detect", func() {
		res = c.Deps.Radar(cfg, func(msg, level string) {
			c.Bus.EmitLogLine(msg, level)
		})
	}); crashed {
		return &RadarResult{Success: false, Error: "the radar crashed — see the desk log"}
	}
	if res == nil {
		res = &RadarResult{Success: false, Error: "the radar engine returned no result"}
	}
	return res
}

// pyStrCfg reads a config string the way the engine bridges do
// (truthy strings only; junk reads as "").
func pyStrCfg(cfg *storage.OrderedMap, key string) string {
	if cfg == nil {
		return ""
	}
	if v, ok := cfg.Get(key); ok {
		if s, is := v.(string); is {
			return s
		}
	}
	return ""
}
