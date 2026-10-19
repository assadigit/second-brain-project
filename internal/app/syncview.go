package app

// syncview.go — T-75 slice (b): THE DEFERRED LEDGER's view half — the
// stage-by-stage snapshot gather (sync_view.py's _gather_sync_snapshot
// + worker_jobs._sync_display_job), the matrix's shape tables, the
// differences modal's engines, and the Settings → Sync page's doors
// (one Refresh measures BOTH desks: the four-element flow above + the
// gap matrix below — v0.87.0's GRAND CONVERGENCE). The pure decisions
// live in internal/syncreview; this file composes them over the
// config + the queue state + the ledger, with every walk and network
// door injectable (the HERMETIC-SEAM law).
//
// THE QUIET-OPEN LAW: nothing here fetches at startup — the Refresh
// click is the consent for every network leg; the local walks always
// run. READ-ONLY BY LAW: the refresh never seals, never publishes,
// never scans, never mutates a file.

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/syncreview"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// ---------------------------------------------------------------------------
// The matrix's shape — the six real cells, their labels and fix doors
// ---------------------------------------------------------------------------

// SyncStageDef is one stage's own row of _STAGE_DEFS.
type SyncStageDef struct {
	Key         string `json:"key"`
	Stream      string `json:"stream"`
	Stage       string `json:"stage"`
	SourceLabel string `json:"source_label"`
	TargetLabel string `json:"target_label"`
	FixLabel    string `json:"fix_label"`
	FixAction   string `json:"fix_action"` // check_queue | publish | seal
}

// SyncStageDefs mirrors _STAGE_DEFS (the six real cells).
var SyncStageDefs = []SyncStageDef{
	{"repos:queue_to_vault", "Repos", "Queue → Vault",
		"the bot queue", "the vault", "Check the bot queue", "check_queue"},
	{"websites:queue_to_vault", "Websites", "Queue → Vault",
		"the bot queue", "the vault", "Check the bot queue", "check_queue"},
	{"websites:vault_to_site", "Websites", "Vault → Website",
		"the vault", "the website", "Publish the directory now", "publish"},
	{"repos:vault_to_backup", "Repos", "Vault → Backup",
		"the vault", "the backup mirror", "Seal the vault now", "seal"},
	{"websites:vault_to_backup", "Websites", "Vault → Backup",
		"the vault", "the backup mirror", "Seal the vault now", "seal"},
	{"personal:vault_to_backup", "Personal", "Vault → Backup",
		"the vault", "the backup mirror", "Seal the vault now", "seal"},
}

// SyncStageColumns mirrors _STAGE_COLUMNS.
var SyncStageColumns = []string{"Queue → Vault", "Vault → Website",
	"Vault → Backup"}

// SyncMatrixLayout mirrors _MATRIX_LAYOUT: the stream rows × stage
// columns of the painted matrix ("" = not a stage for that stream — a
// muted dash, never an Unknown).
var SyncMatrixLayout = map[string][]string{
	"Repos":    {"repos:queue_to_vault", "", "repos:vault_to_backup"},
	"Websites": {"websites:queue_to_vault", "websites:vault_to_site", "websites:vault_to_backup"},
	"Personal": {"", "", "personal:vault_to_backup"},
}

// SyncMatrixStreams is the rows' own order.
var SyncMatrixStreams = []string{"Repos", "Websites", "Personal"}

// DiffListCap mirrors _DIFF_LIST_CAP: the modal's list cap (the counts
// stay honest; the list paints at most this many rows, the cap spoken).
const DiffListCap = 400

// ---------------------------------------------------------------------------
// The gather (the worker's own half — every heavy thing off the GUI)
// ---------------------------------------------------------------------------

// SyncQueueState is the queue stages' data (the Bot page's own check
// feeds it; the Refresh poll refreshes it).
type SyncQueueState struct {
	Checked         bool     `json:"checked"`
	PendingRepos    []string `json:"pending_repos"`
	PendingWebsites []string `json:"pending_websites"`
	InVault         int      `json:"in_vault"`
	WebsitesInVault int      `json:"websites_in_vault"`
}

// SyncGatherDeps carries the walk doors (the HERMETIC-SEAM law — every
// one nil → the production syncreview door).
type SyncGatherDeps struct {
	SiteItems func(vault string, log func(string), cap int) ([]syncreview.GapItem, bool)
	BackupGap func(vault string, log func(string), cap int) ([]syncreview.GapItem, bool)
	NoteCount func(vault string) *int
	Published func(workerURL string, timeoutS float64) *storage.OrderedMap
}

func (d *SyncGatherDeps) siteItems() func(string, func(string), int) ([]syncreview.GapItem, bool) {
	if d == nil || d.SiteItems == nil {
		return syncreview.VaultSiteItems
	}
	return d.SiteItems
}

func (d *SyncGatherDeps) backupGap() func(string, func(string), int) ([]syncreview.GapItem, bool) {
	if d == nil || d.BackupGap == nil {
		return syncreview.VaultBackupGap
	}
	return d.BackupGap
}

func (d *SyncGatherDeps) noteCount() func(string) *int {
	if d == nil || d.NoteCount == nil {
		return syncreview.VaultNoteCount
	}
	return d.NoteCount
}

func (d *SyncGatherDeps) published() func(string, float64) *storage.OrderedMap {
	if d == nil || d.Published == nil {
		return syncreview.FetchPublishedIndex
	}
	return d.Published
}

// SyncSnapshot is the gather's answer: the six stage records + the
// generated-at stamp.
type SyncSnapshot struct {
	Stages      map[string]*syncreview.StageRecord `json:"stages"`
	GeneratedAt string                             `json:"generated_at"`
}

// nowISOSecondsUTC is the snapshot's own stamp shape.
func nowISOSecondsUTC() string {
	return time.Now().UTC().Format("2006-01-02T15:04:05+00:00")
}

// GatherSyncSnapshot mirrors _gather_sync_snapshot: compose the six
// stage snapshots (vault walks, one published-index GET, git status
// reads). Never raises; the caller owns the thread.
func GatherSyncSnapshot(cfg *storage.OrderedMap, queue *SyncQueueState,
	ledger *syncreview.DeferredLog, deps *SyncGatherDeps,
	log func(string)) *SyncSnapshot {
	if cfg == nil {
		cfg = storage.NewOrderedMap()
	}
	if log == nil {
		log = func(string) {}
	}
	stages := map[string]*syncreview.StageRecord{}
	now := nowISOSecondsUTC()

	finish := func(rec syncreview.StageRecord) {
		stages[rec.Key] = &rec
	}

	// ---- Websites · Vault → Website (the flagship stage) ------------
	webVault := asPathOf(cfg, "website_vault_path")
	workerURL := strings.TrimSpace(firstNonEmpty(
		strCfgOr(cfg, "directory_worker_url"),
		strCfgOr(cfg, "cloudflare_worker_url")))
	vtsSetUp := webVault != "" && workerURL != ""
	vtsHint, vtsReason, vtsNeeds := "", "", ""
	if webVault == "" {
		vtsHint = "set the Websites vault (Settings → Vault → Websites vault folder)"
		vtsReason = "Websites vault path not set"
		vtsNeeds = "setup"
	} else if workerURL == "" {
		vtsHint = "set the Worker URL (Settings → Credentials → Cloudflare Worker URL)"
		vtsReason = "no Worker URL set"
		vtsNeeds = "setup"
	}
	var sourceItems []syncreview.GapItem
	sourceOK := false
	if webVault != "" {
		sourceItems, sourceOK = deps.siteItems()(webVault, log, 0)
		if !sourceOK {
			sourceItems = nil
		}
	}
	var published *storage.OrderedMap
	targetURLs, targetOK := ([]string)(nil), false
	if vtsSetUp {
		published = deps.published()(workerURL, 0)
	}
	if published != nil {
		targetURLs, targetOK = syncreview.PublishedSiteUrls(published)
	}
	if !targetOK && vtsSetUp {
		vtsHint = "the published index did not answer (" + workerURL +
			syncreview.PublishedIndexPath + ")"
		vtsReason = "the site did not answer"
		vtsNeeds = "measure"
	}
	if targetOK && ledger != nil {
		ledger.ResolveSeen("websites:vault_to_site", targetURLs, "")
	}
	vtsDeferred := map[string]bool{}
	if ledger != nil {
		vtsDeferred = ledger.ReviewedIDs("websites:vault_to_site")
	}
	sourceIDs := make([]string, 0, len(sourceItems))
	for _, it := range sourceItems {
		sourceIDs = append(sourceIDs, it.ID)
	}
	var vtsTarget []string
	if targetOK {
		vtsTarget = targetURLs
	}
	vtsCell := syncreview.StageCell(vtsSetUp, sourceIDs, vtsTarget,
		keysOfSet(vtsDeferred))
	vtsItems := []syncreview.GapItem{}
	if sourceOK && targetOK {
		targetSet := map[string]bool{}
		for _, u := range targetURLs {
			targetSet[u] = true
		}
		for _, it := range sourceItems {
			if !targetSet[it.ID] {
				days := syncreview.AddedDaysAgo(it.AddedAt, time.Time{})
				it.AddedDaysAgo = days
				vtsItems = append(vtsItems, it)
			}
		}
		vtsItems = syncreview.SortOldestFirst(vtsItems)
		if len(vtsItems) > DiffListCap {
			vtsItems = vtsItems[:DiffListCap]
		}
	}
	finish(syncreview.StageRecord{
		Key: "websites:vault_to_site", Stream: "Websites",
		Stage: "Vault → Website", State: vtsCell.State,
		SetUp: vtsSetUp, SetUpHint: vtsHint, Reason: vtsReason, Needs: vtsNeeds,
		SourceLabel: "the vault", TargetLabel: "the website",
		FixLabel: "Publish the directory now", FixAction: "publish",
		DisplayNumber:    vtsCell.TargetCount,
		SourceCount:      vtsCell.SourceCount,
		TargetCount:      vtsCell.TargetCount,
		GapCount:         vtsCell.GapCount,
		UnexplainedCount: vtsCell.UnexplainedCount,
		DeferredGapCount: vtsCell.DeferredGapCount,
		UnexplainedIDs:   vtsCell.UnexplainedIDs,
		DeferredGapIDs:   vtsCell.DeferredGapIDs,
		Items:            vtsItems,
	})

	// ---- the two Queue → Vault stages ---------------------------------
	checked := false
	var pendingRepos, pendingWebsites []string
	inVault, websitesInVault := 0, 0
	if queue != nil {
		checked = queue.Checked
		pendingRepos = queue.PendingRepos
		pendingWebsites = queue.PendingWebsites
		inVault = queue.InVault
		websitesInVault = queue.WebsitesInVault
	}
	if pendingRepos == nil {
		pendingRepos = []string{} // list(...) — an empty list, never None
	}
	if pendingWebsites == nil {
		pendingWebsites = []string{}
	}
	for _, spec := range []struct {
		Key, Stream string
		Pending     []string
		Display     *int
	}{
		{"repos:queue_to_vault", "Repos", pendingRepos, nil},
		{"websites:queue_to_vault", "Websites", pendingWebsites, nil},
	} {
		if checked {
			if spec.Stream == "Repos" {
				spec.Display = intPtrLocal(inVault)
			} else {
				spec.Display = intPtrLocal(websitesInVault)
			}
		}
		if ledger != nil {
			ledger.ResolveGone(spec.Key, spec.Pending, "")
		}
		deferred := map[string]bool{}
		if ledger != nil {
			deferred = ledger.ReviewedIDs(spec.Key)
		}
		cell := syncreview.StageCell(checked, spec.Pending, []string{},
			keysOfSet(deferred))
		items := make([]syncreview.GapItem, 0, len(spec.Pending))
		for _, u := range spec.Pending {
			items = append(items, syncreview.GapItem{ID: u, Name: u})
		}
		hint, reason, needs := "", "", ""
		if !checked {
			hint = "the bot queue has not been checked this session — " +
				"\"Check the bot queue\" measures it (the quiet-open law: " +
				"the app never fetches on its own)"
			reason = "bot queue not checked this session"
			needs = "measure"
		}
		finish(syncreview.StageRecord{
			Key: spec.Key, Stream: spec.Stream, Stage: "Queue → Vault",
			State: cell.State, SetUp: checked, SetUpHint: hint,
			Reason: reason, Needs: needs,
			SourceLabel: "the bot queue", TargetLabel: "the vault",
			FixLabel: "Check the bot queue", FixAction: "check_queue",
			DisplayNumber:    spec.Display,
			SourceCount:      cell.SourceCount,
			TargetCount:      cell.TargetCount,
			GapCount:         cell.GapCount,
			UnexplainedCount: cell.UnexplainedCount,
			DeferredGapCount: cell.DeferredGapCount,
			UnexplainedIDs:   cell.UnexplainedIDs,
			DeferredGapIDs:   cell.DeferredGapIDs,
			Items:            items,
		})
	}

	// ---- the three Vault → Backup stages ------------------------------
	sealOn := true
	if vs := cfgNested(cfg, "vaultseal"); vs != nil {
		if v, ok := vs.Get("enabled"); ok {
			if b, is := v.(bool); is {
				sealOn = b
			}
		}
	}
	for _, spec := range []struct{ Key, Stream, PathKey string }{
		{"repos:vault_to_backup", "Repos", "vault_path"},
		{"websites:vault_to_backup", "Websites", "website_vault_path"},
		{"personal:vault_to_backup", "Personal", "manual_vault_path"},
	} {
		vault := asPathOf(cfg, spec.PathKey)
		setUp := sealOn && vault != ""
		hint, reason, needs := "", "", ""
		if !sealOn {
			hint = "VaultSeal is off (Settings → Backup → enable \"Seal after every run\")"
			reason = "sealing is off"
			needs = "setup"
		} else if vault == "" {
			hint = "set the " + spec.Stream + " vault path (Settings → Vault)"
			reason = spec.Stream + " vault path not set"
			needs = "setup"
		}
		var gap []syncreview.GapItem
		gapOK := false
		if vault != "" {
			gap, gapOK = deps.backupGap()(vault, log, 0)
		}
		if !gapOK && vault != "" && setUp {
			gap, gapOK = []syncreview.GapItem{}, true
		}
		gapIDs := make([]string, 0, len(gap))
		for _, it := range gap {
			gapIDs = append(gapIDs, it.ID)
		}
		if gapOK && ledger != nil {
			ledger.ResolveGone(spec.Key, gapIDs, "")
		}
		deferred := map[string]bool{}
		if ledger != nil {
			deferred = ledger.ReviewedIDs(spec.Key)
		}
		var cellSrc []string
		if gapOK {
			cellSrc = gapIDs
		}
		cell := syncreview.StageCell(setUp, cellSrc, []string{},
			keysOfSet(deferred))
		items := []syncreview.GapItem{}
		if gapOK {
			if len(gap) > DiffListCap {
				gap = gap[:DiffListCap]
			}
			items = gap
		}
		var display *int
		if vault != "" {
			if vaultCount := deps.noteCount()(vault); vaultCount != nil {
				d := *vaultCount - cell.GapCount
				if d < 0 {
					d = 0
				}
				display = intPtrLocal(d)
			}
		}
		finish(syncreview.StageRecord{
			Key: spec.Key, Stream: spec.Stream, Stage: "Vault → Backup",
			State: cell.State, SetUp: setUp, SetUpHint: hint,
			Reason: reason, Needs: needs,
			SourceLabel: "the vault", TargetLabel: "the backup mirror",
			FixLabel: "Seal the vault now", FixAction: "seal",
			DisplayNumber:    display,
			SourceCount:      cell.SourceCount,
			TargetCount:      cell.TargetCount,
			GapCount:         cell.GapCount,
			UnexplainedCount: cell.UnexplainedCount,
			DeferredGapCount: cell.DeferredGapCount,
			UnexplainedIDs:   cell.UnexplainedIDs,
			DeferredGapIDs:   cell.DeferredGapIDs,
			Items:            items,
		})
	}

	return &SyncSnapshot{Stages: stages, GeneratedAt: now}
}

// SyncBannerOver computes the banner's verdict over a snapshot's
// stages (worst-first; the identity rides the records).
func SyncBannerOver(snap *SyncSnapshot) syncreview.Verdict {
	if snap == nil {
		return syncreview.BannerVerdict(nil)
	}
	cells := make([]*syncreview.Cell, 0, len(snap.Stages))
	for _, key := range syncStageOrder {
		if rec, ok := snap.Stages[key]; ok {
			cells = append(cells, &syncreview.Cell{
				State: rec.State, SourceCount: rec.SourceCount,
				TargetCount: rec.TargetCount, GapCount: rec.GapCount,
				UnexplainedCount: rec.UnexplainedCount,
				DeferredGapCount: rec.DeferredGapCount,
				Stream:           rec.Stream, Stage: rec.Stage,
				Reason: rec.Reason, Needs: rec.Needs,
			})
		}
	}
	return syncreview.BannerVerdict(cells)
}

// syncStageOrder is the defs' own order (the banner reads it
// deterministically).
var syncStageOrder = func() []string {
	out := make([]string, 0, len(SyncStageDefs))
	for _, d := range SyncStageDefs {
		out = append(out, d.Key)
	}
	return out
}()

// SyncRecomputeFromLedger mirrors _sync_recompute_from_ledger:
// recompute every stage cell from the LAST snapshot's items and the
// ledger's CURRENT reviewed sets (pure core calls — instant, honest
// between gathers; unmeasured or item-less stages keep their paint).
// Returns the refreshed banner.
func SyncRecomputeFromLedger(snap *SyncSnapshot,
	ledger *syncreview.DeferredLog) syncreview.Verdict {
	if snap == nil {
		return syncreview.BannerVerdict(nil)
	}
	for key, stage := range snap.Stages {
		if stage.State == syncreview.StateUnknown || len(stage.Items) == 0 {
			continue
		}
		gapIDs := make([]string, 0, len(stage.Items))
		for _, it := range stage.Items {
			gapIDs = append(gapIDs, it.ID)
		}
		var deferred []string
		if ledger != nil {
			deferred = keysOfSet(ledger.ReviewedIDs(key))
		}
		cell := syncreview.StageCell(true, gapIDs, []string{}, deferred)
		stage.UnexplainedIDs = cell.UnexplainedIDs
		stage.DeferredGapIDs = cell.DeferredGapIDs
		stage.UnexplainedCount = cell.UnexplainedCount
		stage.DeferredGapCount = cell.DeferredGapCount
		stage.GapCount = cell.GapCount
		stage.State = cell.State
	}
	return SyncBannerOver(snap)
}

// ---------------------------------------------------------------------------
// The published-index facts door (directory_sync's read-back law)
// ---------------------------------------------------------------------------

// publishedIndexHTTP is the production GET seam (code-aware): the
// Refresh door rides it; the tests inject canned (code, body, err)
// answers. Package-level but only swapped under the app package's own
// tests (the settings doors take their own injected copy — see
// SyncDeps.PublishedFetch).
func publishedIndexHTTP(url string, timeoutS float64) (int, []byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return 0, nil, err
	}
	req.Header.Set("User-Agent", workerclient.HTTPUserAgent)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: publishedTimeout(timeoutS)}
	res, err := client.Do(req)
	if err != nil {
		return 0, nil, err
	}
	defer res.Body.Close()
	body := make([]byte, 0, 4096)
	buf := make([]byte, 4096)
	for {
		n, rerr := res.Body.Read(buf)
		body = append(body, buf[:n]...)
		if rerr != nil {
			break
		}
	}
	return res.StatusCode, body, nil
}

func publishedTimeout(timeoutS float64) time.Duration {
	if timeoutS <= 0 {
		timeoutS = 15
	}
	return time.Duration(timeoutS * float64(time.Second))
}

// FetchPublishedFacts is directory_sync.fetch_published_index: the
// display's read-back door — one of four truths, never a silent zero:
// ok (sites + generated_at) · never_published (the Worker's own 404)
// · not_configured (no worker URL) · error (spoken, never swallowed).
func FetchPublishedFacts(cfg *storage.OrderedMap,
	fetch func(url string, timeoutS float64) (int, []byte, error),
	log func(string)) SyncPublishedFacts {
	if fetch == nil {
		fetch = publishedIndexHTTP
	}
	if log == nil {
		log = func(string) {}
	}
	out := SyncPublishedFacts{State: SyncStateError}
	url := strCfgOr(cfg, "directory_worker_url")
	if url == "" {
		return SyncPublishedFacts{State: SyncStateNotSet,
			Message: "no worker URL configured — Settings → " +
				"Credentials has the door"}
	}
	code, body, err := fetch(strings.TrimRight(url, "/")+
		syncreview.PublishedIndexPath, 0)
	if err != nil {
		out.Message = truncateStr(err.Error(), 200)
		return out
	}
	if code == 404 {
		// the Worker's own honest answer: the publish door exists but
		// nothing has ever been published through it.
		return SyncPublishedFacts{State: SyncStateNever,
			Message: "never published — the site is empty until the " +
				"first publish runs"}
	}
	if code != 200 {
		out.Message = "HTTP " + strconv.Itoa(code)
		return out
	}
	data, derr := storage.DecodeJSONValue(body)
	if derr != nil {
		out.Message = truncateStr(derr.Error(), 200)
		return out
	}
	om, ok := data.(*storage.OrderedMap)
	if !ok {
		out.Message = "the published index carried no site count"
		return out
	}
	countsV, _ := om.Get("counts")
	counts, _ := countsV.(*storage.OrderedMap)
	if counts == nil {
		out.Message = "the published index carried no site count"
		return out
	}
	sitesV, _ := counts.Get("sites")
	sites := -1
	switch t := sitesV.(type) {
	case int:
		sites = t
	case int64:
		sites = int(t)
	case float64:
		sites = int(t)
	}
	if sites < 0 {
		out.Message = "the published index carried no site count"
		return out
	}
	genV, _ := om.Get("generated_at")
	gen, _ := genV.(string)
	return SyncPublishedFacts{State: SyncStateOK, Sites: &sites,
		GeneratedAt: gen}
}

// ---------------------------------------------------------------------------
// The Settings → Sync doors (the page's own command surface)
// ---------------------------------------------------------------------------

// SyncDeps carries the Sync page's seams (nil → the production doors).
type SyncDeps struct {
	// QueuePoll is GetPendingResult's shape ("ok"|"disabled"|"error",
	// the pending rows).
	QueuePoll func(cfg *storage.OrderedMap) (string, []any)
	// PublishedFetch is the code-aware published GET ((code, body,
	// err) — nil → the production net/http door).
	PublishedFetch func(url string, timeoutS float64) (int, []byte, error)
	// GatherDeps rides the gather's own walks.
	GatherDeps *SyncGatherDeps
	// LedgerOpen opens the DeferredLog (nil → the APP_DIR default).
	LedgerOpen func() (*syncreview.DeferredLog, error)
}

// SyncViewPayload is ONE Refresh's answer: the facts (the local +
// network halves), the display model (the four cards), the snapshot
// (the six stage records), the banner + the desk lines.
type SyncViewPayload struct {
	Facts    *SyncFacts         `json:"facts"`
	Model    *SyncDisplayModel  `json:"model"`
	Snapshot *SyncSnapshot      `json:"snapshot"`
	Banner   syncreview.Verdict `json:"banner"`
	Measured string             `json:"measured_at"`
	Error    string             `json:"error,omitempty"`
}

// SetSyncDeps installs the Sync page's seams (the DI root + the tests).
func (s *Settings) SetSyncDeps(d SyncDeps) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncDeps = d
}

func (s *Settings) syncQueuePoll() func(*storage.OrderedMap) (string, []any) {
	s.mu.Lock()
	d := s.syncDeps
	s.mu.Unlock()
	if d.QueuePoll != nil {
		return d.QueuePoll
	}
	return func(cfg *storage.OrderedMap) (string, []any) {
		return workerclient.NewSync(cfg).GetPendingResult()
	}
}

func (s *Settings) syncPublishedFetch() func(string, float64) (int, []byte, error) {
	s.mu.Lock()
	d := s.syncDeps
	s.mu.Unlock()
	if d.PublishedFetch != nil {
		return d.PublishedFetch
	}
	return publishedIndexHTTP
}

func (s *Settings) syncLedger() *syncreview.DeferredLog {
	s.mu.Lock()
	d := s.syncDeps
	inst := s.syncLedgerIns
	s.mu.Unlock()
	if inst != nil {
		return inst
	}
	var (
		log *syncreview.DeferredLog
		err error
	)
	if d.LedgerOpen != nil {
		log, err = d.LedgerOpen()
	} else {
		log, err = syncreview.OpenLedger("")
	}
	if err != nil {
		return nil // the honest zero — every ledger door no-ops on nil
	}
	s.mu.Lock()
	s.syncLedgerIns = log
	s.mu.Unlock()
	return log
}

// SetSyncQueueState lands the session's queue state (the Bot page's
// own check feeds it — the queue stages' data is fresh).
func (s *Settings) SetSyncQueueState(state *SyncQueueState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.syncQueue = state
}

func (s *Settings) currentSyncQueue() *SyncQueueState {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.syncQueue
}

// RefreshSyncView is THE ONE CONSENT DOOR (v0.87.0's convergence: one
// Refresh measures BOTH desks): the local walks always run; the live
// doors (the bot queue, the published site) are asked only because
// the owner clicked. Read-only by law. Never raises.
func (s *Settings) RefreshSyncView() *SyncViewPayload {
	now := nowISOSecondsUTC()
	defer func() {
		if r := recover(); r != nil {
			s.logf("warning", "⚠️ Sync view refresh failed: %s", pyCrashText(r))
		}
	}()
	cfg := s.PlainConfig()
	facts := &SyncFacts{
		Vaults: &SyncVaultsFacts{Vaults: DetectScanVaults(cfg, true)},
		GitHub: &SyncGitHubFacts{Rows: backupRepoRows(cfg, nil)},
	}
	localWeb := websiteVaultSide(cfg, nil)
	facts.Website = &SyncWebsiteFacts{Vault: localWeb}

	// -- the bot queue (network, consented by the Refresh click) -----
	state, pending := s.syncQueuePoll()(cfg)
	botFacts := &SyncBotFacts{State: state, Pending: pending}
	if state == "disabled" {
		botFacts.State = SyncStateNotSet
	} else if state != "ok" {
		botFacts.State = SyncStateError
	}
	facts.Bot = botFacts

	// -- the published index (network, consented by the same click) --
	s.mu.Lock()
	deps := s.syncDeps
	s.mu.Unlock()
	facts.Website.Published = FetchPublishedFacts(cfg,
		s.syncPublishedFetch(), nil)

	// -- the display model (PURE) + the gather ------------------------
	model := SyncDisplayModelOf(facts)
	// the queue stages ride the Bot page's OWN check (the classified
	// gap after the vault filter — the raw poll would lie); the
	// Refresh measures everything else
	queue := s.currentSyncQueue()
	snap := GatherSyncSnapshot(cfg, queue, s.syncLedger(),
		deps.GatherDeps, func(string) {})
	banner := SyncBannerOver(snap)
	s.mu.Lock()
	s.syncSnapshot = snap
	s.mu.Unlock()
	s.speakSyncRefresh(model, banner)
	return &SyncViewPayload{Facts: facts, Model: model, Snapshot: snap,
		Banner: banner, Measured: now}
}

// speakSyncRefresh speaks the desk lines (the flow's own verdict).
func (s *Settings) speakSyncRefresh(model *SyncDisplayModel, banner syncreview.Verdict) {
	if model != nil && model.CalmLine != "" {
		s.SpeakOnBus("✅ "+model.CalmLine, "success")
		return
	}
	if model != nil {
		s.SpeakOnBus(fmt.Sprintf(
			"🔄 Sync view refreshed — %d item(s) waiting in the flow",
			model.WaitingTotal), "info")
	}
	if banner.Tone == "warning" {
		s.SpeakOnBus("⚠️ "+banner.Text, "warning")
	}
}

// SyncDifferencesPayload is the differences modal's data: the stage
// record (from the last snapshot) + the deferred-log rows + the
// heading + the log count.
type SyncDifferencesPayload struct {
	Stage    *syncreview.StageRecord `json:"stage"`
	Heading  string                  `json:"heading"`
	Rows     []syncreview.LedgerRow  `json:"rows"`
	LogCount int                     `json:"log_count"`
	Found    bool                    `json:"found"`
}

// SyncDiffHeading is _diff_heading: "In {src}, not in {tgt} — N items"
// (+ the already-explained split when it applies).
func SyncDiffHeading(stage *syncreview.StageRecord) string {
	if stage == nil {
		return ""
	}
	n := len(stage.Items)
	src := stage.SourceLabel
	if src == "" {
		src = "the source"
	}
	tgt := stage.TargetLabel
	if tgt == "" {
		tgt = "the target"
	}
	text := "In " + src + ", not in " + tgt + " — " + fmt.Sprint(n) +
		" item" + pluralS(n)
	if stage.UnexplainedCount > 0 && stage.UnexplainedCount < n {
		text += fmt.Sprintf(" (%d already explained, %d waiting)",
			n-stage.UnexplainedCount, stage.UnexplainedCount)
	}
	return text
}

// SyncDifferences answers the modal's data for one stage (the stage
// record from the LAST snapshot + the ledger's rows — pure local,
// instant; no network).
func (s *Settings) SyncDifferences(key string) *SyncDifferencesPayload {
	s.mu.Lock()
	snap := s.syncSnapshot
	s.mu.Unlock()
	if snap == nil {
		return &SyncDifferencesPayload{}
	}
	stage, ok := snap.Stages[key]
	if !ok {
		return &SyncDifferencesPayload{}
	}
	ledger := s.syncLedger()
	rows := []syncreview.LedgerRow{}
	count := 0
	if ledger != nil {
		rows = ledger.Rows(key, "")
		count = ledger.RowCount(key)
	}
	return &SyncDifferencesPayload{Stage: stage,
		Heading: SyncDiffHeading(stage), Rows: rows, LogCount: count,
		Found: true}
}

// SyncMarkIntentional is the modal's Mark button: bulk marking with
// ONE shared reason; the items leave the Differences list, the cell
// turns Reviewed, the reason is remembered. Returns the marked count
// + the inline confirmation line (ONE line, never a second dialog).
func (s *Settings) SyncMarkIntentional(stageKey string,
	itemIDs []string, reason, note string,
	displayNames map[string]string) (int, string) {
	ledger := s.syncLedger()
	if ledger == nil || !syncreview.MarkAllowed(len(itemIDs), reason) {
		return 0, ""
	}
	n := ledger.Mark(stageKey, itemIDs, reason, note, displayNames, "")
	s.mu.Lock()
	snap := s.syncSnapshot
	s.mu.Unlock()
	if snap != nil {
		SyncRecomputeFromLedger(snap, ledger)
	}
	label := syncreview.ReasonLabels[reason]
	line := fmt.Sprintf("✓ %d item%s marked — %s", n, pluralS(n), label)
	if note != "" {
		line += " (" + note + ")"
	}
	s.SpeakOnBus(line, "success")
	return n, line
}

// SyncReopenItem is the modal's Reopen: the item returns to the
// Differences tab (the deferral is withdrawn, the row stays in the
// log, the reopen counted). Answers the confirmation line.
func (s *Settings) SyncReopenItem(stageKey, itemID string) (bool, string) {
	ledger := s.syncLedger()
	if ledger == nil {
		return false, ""
	}
	if !ledger.Reopen(stageKey, itemID, "") {
		return false, ""
	}
	s.mu.Lock()
	snap := s.syncSnapshot
	s.mu.Unlock()
	if snap != nil {
		SyncRecomputeFromLedger(snap, ledger)
	}
	const line = "↩ 1 item reopened — back in the Differences tab"
	s.SpeakOnBus(line, "info")
	return true, line
}

// SyncDeferredRows answers the log tab's rows (search included, the
// cap 500 the frontend paints).
func (s *Settings) SyncDeferredRows(stage, query string) []syncreview.LedgerRow {
	ledger := s.syncLedger()
	if ledger == nil {
		return nil
	}
	return ledger.Rows(stage, query)
}

// SyncLedgerCount answers the ledger's own row count (the log tab's
// badge + the last-measured line).
func (s *Settings) SyncLedgerCount() int {
	ledger := s.syncLedger()
	if ledger == nil {
		return 0
	}
	return ledger.RowCount("")
}

// ---------------------------------------------------------------------------
// tiny helpers
// ---------------------------------------------------------------------------

func keysOfSet(set map[string]bool) []string {
	out := make([]string, 0, len(set))
	for k := range set {
		out = append(out, k)
	}
	return out
}

func intPtrLocal(n int) *int { return &n }

func strCfgOr(cfg *storage.OrderedMap, key string) string {
	if cfg == nil {
		return ""
	}
	v, _ := cfg.Get(key)
	s, _ := v.(string)
	return strings.TrimSpace(s)
}

func firstNonEmpty(vals ...string) string {
	for _, v := range vals {
		if v != "" {
			return v
		}
	}
	return ""
}

func truncateStr(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
