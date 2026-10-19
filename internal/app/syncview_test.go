package app

// syncview_test.go — T-75 slice (b)'s view-half battery: the gather's
// six stage records over hermetic fakes (THE HERMETIC-SEAM law), the
// differences modal's engines (the disabled-until law, the mark +
// reopen flows, the recompute-from-ledger), and the banner over the
// snapshot. Ported from tests/test_syncreview.py's UX families +
// sync_view.py's own laws.

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/syncreview"
)

// ---------------------------------------------------------------------------
// the harness — a Settings over a temp app dir with fake sync seams
// ---------------------------------------------------------------------------

// newSyncSettings builds a Settings whose sync doors all ride fakes.
func newSyncSettings(t *testing.T, cfg *storage.OrderedMap,
	queuePoll func(*storage.OrderedMap) (string, []any),
	published func(string, float64) (int, []byte, error)) (*Settings, *busRec) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	cfgPath := filepath.Join(appDir, "config.json")
	if err := storage.WriteConfigFile(cfgPath, cfg); err != nil {
		t.Fatal(err)
	}
	a := New(cfgPath, false)
	rec := &busRec{}
	rec.subscribe(a.Bus)
	a.Settings.SetSyncDeps(SyncDeps{
		QueuePoll:      queuePoll,
		PublishedFetch: published,
		LedgerOpen: func() (*syncreview.DeferredLog, error) {
			return syncreview.OpenLedger(filepath.Join(appDir, "sync_review.db"))
		},
	})
	return a.Settings, rec
}

// fakeGatherDeps answers canned walks for the gather (hermetic).
func fakeGatherDeps(siteItems []syncreview.GapItem, siteOK bool,
	backupGap map[string][]syncreview.GapItem, backupOK map[string]bool,
	noteCount map[string]int, published *storage.OrderedMap) *SyncGatherDeps {
	return &SyncGatherDeps{
		SiteItems: func(vault string, log func(string), cap int) ([]syncreview.GapItem, bool) {
			return siteItems, siteOK
		},
		BackupGap: func(vault string, log func(string), cap int) ([]syncreview.GapItem, bool) {
			gap, ok := backupGap[vault]
			if !ok {
				return nil, false
			}
			return gap, backupOK[vault]
		},
		NoteCount: func(vault string) *int {
			if n, ok := noteCount[vault]; ok {
				return &n
			}
			return nil
		},
		Published: func(workerURL string, timeoutS float64) *storage.OrderedMap {
			return published
		},
	}
}

func syncCfg(vaultPath, webVault, workerURL string) *storage.OrderedMap {
	cfg := storage.NewOrderedMap()
	if vaultPath != "" {
		cfg.Set("vault_path", vaultPath)
	}
	if webVault != "" {
		cfg.Set("website_vault_path", webVault)
	}
	if workerURL != "" {
		cfg.Set("directory_worker_url", workerURL)
	}
	return cfg
}

// ---------------------------------------------------------------------------
// the gather — the six stage records over fakes
// ---------------------------------------------------------------------------

func TestGatherTheSixStagesExist(t *testing.T) {
	snap := GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"),
		nil, nil, fakeGatherDeps(nil, true,
			map[string][]syncreview.GapItem{}, map[string]bool{},
			map[string]int{}, nil), nil)
	if len(snap.Stages) != 6 {
		t.Fatalf("stages = %d", len(snap.Stages))
	}
	for _, def := range SyncStageDefs {
		rec, ok := snap.Stages[def.Key]
		if !ok {
			t.Fatalf("stage %q missing", def.Key)
		}
		if rec.Stream != def.Stream || rec.Stage != def.Stage ||
			rec.FixAction != def.FixAction || rec.FixLabel != def.FixLabel {
			t.Fatalf("stage = %+v vs def %+v", rec, def)
		}
	}
	if snap.GeneratedAt == "" {
		t.Fatal("the generated-at stamp must ride")
	}
}

func TestGatherTheFlagshipStageSetupReasons(t *testing.T) {
	deps := fakeGatherDeps(nil, true, map[string][]syncreview.GapItem{},
		map[string]bool{}, map[string]int{}, nil)
	// no websites vault → setup + the reason
	snap := GatherSyncSnapshot(syncCfg("/v", "", "https://w.example"),
		nil, nil, deps, nil)
	rec := snap.Stages["websites:vault_to_site"]
	if rec.State != syncreview.StateUnknown || rec.Needs != "setup" ||
		rec.Reason != "Websites vault path not set" {
		t.Fatalf("rec = %+v", rec)
	}
	// a vault but no worker URL → setup + the other reason
	snap = GatherSyncSnapshot(syncCfg("/v", "/w", ""),
		nil, nil, deps, nil)
	rec = snap.Stages["websites:vault_to_site"]
	if rec.Needs != "setup" || rec.Reason != "no Worker URL set" {
		t.Fatalf("rec = %+v", rec)
	}
	// set up but the site did not answer → measure
	snap = GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"),
		nil, nil, deps, nil)
	rec = snap.Stages["websites:vault_to_site"]
	if rec.Needs != "measure" || rec.Reason != "the site did not answer" {
		t.Fatalf("rec = %+v", rec)
	}
}

func TestGatherTheFlagshipStageMeasuresTheGap(t *testing.T) {
	// the vault carries 3 sites, the site carries 1 → 2 to publish
	published := storage.NewOrderedMap()
	published.Set("sites", []any{
		siteMap("https://a.com"), siteMap("https://b.com")})
	deps := fakeGatherDeps([]syncreview.GapItem{
		{ID: "https://a.com", Name: "A", URL: "https://a.com", AddedAt: "2026-09-01"},
		{ID: "https://b.com", Name: "B", URL: "https://b.com", AddedAt: "2026-10-01"},
		{ID: "https://c.com", Name: "C", URL: "https://c.com", AddedAt: "2026-09-15"},
	}, true, map[string][]syncreview.GapItem{}, map[string]bool{},
		map[string]int{}, published)
	snap := GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"),
		nil, nil, deps, nil)
	rec := snap.Stages["websites:vault_to_site"]
	if rec.State != syncreview.StateBehind || rec.GapCount != 1 {
		t.Fatalf("rec = %+v", rec)
	}
	if len(rec.Items) != 1 || rec.Items[0].ID != "https://c.com" {
		t.Fatalf("items = %+v", rec.Items)
	}
	if rec.Items[0].AddedDaysAgo == nil {
		t.Fatal("the items carry their added-days-ago sort key")
	}
	// display number = the published count
	if rec.DisplayNumber == nil || *rec.DisplayNumber != 2 {
		t.Fatalf("display = %v", rec.DisplayNumber)
	}
	// the caption speaks the owner's own sentence
	cap := syncreview.CellCaption(rec)
	if !strings.Contains(cap.Cap, "1 to publish") ||
		!strings.Contains(cap.Cap, "3 in the vault") ||
		!strings.Contains(cap.Cap, "2 on the site") {
		t.Fatalf("cap = %+v", cap)
	}
}

func siteMap(url string) *storage.OrderedMap {
	m := storage.NewOrderedMap()
	m.Set("url", url)
	return m
}

func TestGatherTheQueueStagesRideTheBotsOwnCheck(t *testing.T) {
	deps := fakeGatherDeps(nil, true, map[string][]syncreview.GapItem{},
		map[string]bool{}, map[string]int{}, nil)
	// not checked this session → unknown + measure
	snap := GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"),
		nil, nil, deps, nil)
	rec := snap.Stages["repos:queue_to_vault"]
	if rec.State != syncreview.StateUnknown || rec.Needs != "measure" ||
		rec.Reason != "bot queue not checked this session" {
		t.Fatalf("rec = %+v", rec)
	}
	// checked with a pending gap → behind, display = the in-vault count
	snap = GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"),
		&SyncQueueState{Checked: true,
			PendingRepos: []string{"https://github.com/a/one"},
			InVault:      631}, nil, deps, nil)
	rec = snap.Stages["repos:queue_to_vault"]
	if rec.State != syncreview.StateBehind || rec.GapCount != 1 {
		t.Fatalf("rec = %+v", rec)
	}
	if rec.DisplayNumber == nil || *rec.DisplayNumber != 631 {
		t.Fatalf("display = %v", rec.DisplayNumber)
	}
	cap := syncreview.CellCaption(rec)
	if !strings.Contains(cap.Cap, "1 waiting in the queue") ||
		!strings.Contains(cap.Cap, "631 already in the vault") {
		t.Fatalf("cap = %+v", cap)
	}
	// the websites queue stage carries its own list
	wrec := snap.Stages["websites:queue_to_vault"]
	if wrec.State != syncreview.StateMatch {
		t.Fatalf("websites queue rec = %+v (no pending = match)", wrec)
	}
}

func TestGatherTheBackupStagesSetupLaw(t *testing.T) {
	deps := fakeGatherDeps(nil, true, map[string][]syncreview.GapItem{},
		map[string]bool{}, map[string]int{}, nil)
	// sealing off → setup
	cfg := syncCfg("/v", "/w", "https://w.example")
	vs := storage.NewOrderedMap()
	vs.Set("enabled", false)
	cfg.Set("vaultseal", vs)
	snap := GatherSyncSnapshot(cfg, nil, nil, deps, nil)
	rec := snap.Stages["repos:vault_to_backup"]
	if rec.Needs != "setup" || rec.Reason != "sealing is off" {
		t.Fatalf("rec = %+v", rec)
	}
	// a sealed-clean vault → match + the sealed pair caption
	cfg2 := syncCfg("/v", "/w", "https://w.example")
	deps2 := fakeGatherDeps(nil, true,
		map[string][]syncreview.GapItem{"/v": {}},
		map[string]bool{"/v": true},
		map[string]int{"/v": 678}, nil)
	snap = GatherSyncSnapshot(cfg2, nil, nil, deps2, nil)
	rec = snap.Stages["repos:vault_to_backup"]
	if rec.State != syncreview.StateMatch {
		t.Fatalf("rec = %+v", rec)
	}
	if rec.DisplayNumber == nil || *rec.DisplayNumber != 678 {
		t.Fatalf("display = %v", rec.DisplayNumber)
	}
	cap := syncreview.CellCaption(rec)
	if cap.Num != "678" || cap.Cap != "678 in the vault = 678 in the mirror" {
		t.Fatalf("cap = %+v", cap)
	}
	// an unsealed gap → behind + the unsealed caption
	deps3 := fakeGatherDeps(nil, true,
		map[string][]syncreview.GapItem{"/v": {
			{ID: "New.md", Name: "New.md", NotePath: "/v/New.md"}}},
		map[string]bool{"/v": true},
		map[string]int{"/v": 679}, nil)
	snap = GatherSyncSnapshot(cfg2, nil, nil, deps3, nil)
	rec = snap.Stages["repos:vault_to_backup"]
	if rec.State != syncreview.StateBehind {
		t.Fatalf("rec = %+v", rec)
	}
	cap = syncreview.CellCaption(rec)
	if !strings.Contains(cap.Cap, "1 unsealed") ||
		!strings.Contains(cap.Cap, "679 in the vault") ||
		!strings.Contains(cap.Cap, "678 sealed in the mirror") {
		t.Fatalf("cap = %+v", cap)
	}
}

func TestGatherResolvesTheLedgerAutomatically(t *testing.T) {
	dir := t.TempDir()
	ledger, err := syncreview.OpenLedger(filepath.Join(dir, "l.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer ledger.Close()
	// a deferred queue item that got processed → RESOLVED (gap semantics)
	ledger.Mark("repos:queue_to_vault",
		[]string{"https://github.com/a/processed"}, "other", "", nil, "")
	published := storage.NewOrderedMap()
	published.Set("sites", []any{})
	deps := fakeGatherDeps([]syncreview.GapItem{
		{ID: "https://a.com", Name: "A", URL: "https://a.com"},
	}, true, map[string][]syncreview.GapItem{}, map[string]bool{},
		map[string]int{}, published)
	GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"),
		&SyncQueueState{Checked: true,
			PendingRepos: []string{"https://github.com/a/new"}},
		ledger, deps, nil)
	// the processed item resolved; the reviewed set no longer carries it
	rows := ledger.Rows("repos:queue_to_vault", "")
	if len(rows) != 1 || rows[0].Status != syncreview.StatusResolved {
		t.Fatalf("rows = %+v (resolve_gone at the gather's own door)", rows)
	}
	// the site item deferred then seen on the site → RESOLVED (target
	// semantics: present in the target resolves)
	ledger.Mark("websites:vault_to_site", []string{"https://a.com"},
		"private", "", nil, "")
	publishedWithA := storage.NewOrderedMap()
	publishedWithA.Set("sites", []any{siteMap("https://a.com")})
	deps.Published = func(workerURL string, timeoutS float64) *storage.OrderedMap {
		return publishedWithA
	}
	GatherSyncSnapshot(syncCfg("/v", "/w", "https://w.example"), nil,
		ledger, deps, nil)
	rows = ledger.Rows("websites:vault_to_site", "")
	if len(rows) != 1 || rows[0].Status != syncreview.StatusResolved {
		t.Fatalf("rows = %+v (resolve_seen when the item appears)", rows)
	}
}

func TestSyncBannerOverTheSnapshot(t *testing.T) {
	deps := fakeGatherDeps(nil, true, map[string][]syncreview.GapItem{},
		map[string]bool{}, map[string]int{}, nil)
	snap := GatherSyncSnapshot(syncCfg("", "", ""), nil, nil, deps, nil)
	v := SyncBannerOver(snap)
	// every stage unknown + named → the setup/measure split speaks
	if v.Tone != "muted" ||
		!strings.Contains(v.Text, "not set up") {
		t.Fatalf("verdict = %+v", v)
	}
}

// ---------------------------------------------------------------------------
// the differences modal's engines (the deferred-ledger UX flows)
// ---------------------------------------------------------------------------

func TestSyncDifferencesHeadingTheSplit(t *testing.T) {
	stage := &syncreview.StageRecord{
		SourceLabel: "the vault", TargetLabel: "the website",
		Items:            []syncreview.GapItem{{ID: "a"}, {ID: "b"}, {ID: "c"}},
		UnexplainedCount: 1}
	if got := SyncDiffHeading(stage); got !=
		"In the vault, not in the website — 3 items (2 already explained, 1 waiting)" {
		t.Fatalf("heading = %q", got)
	}
	// plain shape (all unexplained)
	stage.UnexplainedCount = 3
	if got := SyncDiffHeading(stage); got !=
		"In the vault, not in the website — 3 items" {
		t.Fatalf("heading = %q", got)
	}
}

func TestSyncMarkIntentionalFlowAndRecompute(t *testing.T) {
	s, _ := newSyncSettings(t, syncCfg("", "", ""),
		func(*storage.OrderedMap) (string, []any) { return "disabled", nil },
		func(string, float64) (int, []byte, error) {
			return 404, nil, nil
		})
	// seed the snapshot with a behind stage (the modal's fuel)
	s.mu.Lock()
	s.syncSnapshot = &SyncSnapshot{Stages: map[string]*syncreview.StageRecord{
		"websites:vault_to_site": {
			Key: "websites:vault_to_site", Stream: "Websites",
			Stage: "Vault → Website", State: syncreview.StateBehind,
			Items: []syncreview.GapItem{
				{ID: "https://a.com", Name: "A"},
				{ID: "https://b.com", Name: "B"}},
			UnexplainedCount: 2, GapCount: 2,
		},
	}}
	s.mu.Unlock()

	// the disabled-until law rides syncreview.MarkAllowed (tested in
	// its own battery); the door refuses an unarmed mark
	if n, line := s.SyncMarkIntentional("websites:vault_to_site",
		nil, "private", "", nil); n != 0 || line != "" {
		t.Fatalf("unarmed mark = %d %q", n, line)
	}
	// the armed mark: both items, one shared reason, the confirm line
	n, line := s.SyncMarkIntentional("websites:vault_to_site",
		[]string{"https://a.com", "https://b.com"}, "private", "the note",
		map[string]string{"https://a.com": "A", "https://b.com": "B"})
	if n != 2 || line != "✓ 2 items marked — Private (the note)" {
		t.Fatalf("mark = %d %q", n, line)
	}
	// the cell turned Reviewed behind the modal (the recompute law)
	s.mu.Lock()
	rec := s.syncSnapshot.Stages["websites:vault_to_site"]
	s.mu.Unlock()
	if rec.State != syncreview.StateReviewed || rec.DeferredGapCount != 2 {
		t.Fatalf("rec = %+v", rec)
	}
	// the banner flipped healthy with the calm deferred count
	if v := SyncBannerOver(s.currentSnapshotForTest()); v.Text != "In sync · 2 deferred" {
		t.Fatalf("verdict = %+v", v)
	}
	// the differences payload carries the log rows + the count
	payload := s.SyncDifferences("websites:vault_to_site")
	if !payload.Found || payload.LogCount != 2 || len(payload.Rows) != 2 {
		t.Fatalf("payload = %+v", payload)
	}
}

func TestSyncReopenItemFlow(t *testing.T) {
	s, _ := newSyncSettings(t, syncCfg("", "", ""),
		func(*storage.OrderedMap) (string, []any) { return "disabled", nil },
		func(string, float64) (int, []byte, error) {
			return 404, nil, nil
		})
	s.mu.Lock()
	s.syncSnapshot = &SyncSnapshot{Stages: map[string]*syncreview.StageRecord{
		"repos:queue_to_vault": {
			Key: "repos:queue_to_vault", Stream: "Repos",
			Stage: "Queue → Vault", State: syncreview.StateReviewed,
			Items:            []syncreview.GapItem{{ID: "https://github.com/a/one"}},
			DeferredGapCount: 1, GapCount: 1,
		},
	}}
	s.mu.Unlock()
	// mark then reopen: the item is back in the Differences tab
	if _, line := s.SyncMarkIntentional("repos:queue_to_vault",
		[]string{"https://github.com/a/one"}, "not_ready", "", nil); line == "" {
		t.Fatal("the mark must land")
	}
	ok, line := s.SyncReopenItem("repos:queue_to_vault",
		"https://github.com/a/one")
	if !ok || line != "↩ 1 item reopened — back in the Differences tab" {
		t.Fatalf("reopen = %v %q", ok, line)
	}
	// the recompute: the deferral withdrawn → behind again
	s.mu.Lock()
	rec := s.syncSnapshot.Stages["repos:queue_to_vault"]
	s.mu.Unlock()
	if rec.State != syncreview.StateBehind || rec.UnexplainedCount != 1 {
		t.Fatalf("rec = %+v", rec)
	}
	// a second reopen of the same row must not happen (only ACTIVE reopens)
	if ok, _ := s.SyncReopenItem("repos:queue_to_vault",
		"https://github.com/a/one"); ok {
		t.Fatal("a reopened row must not reopen again")
	}
}

func TestSyncDeferredRowsSearchAndCount(t *testing.T) {
	s, _ := newSyncSettings(t, syncCfg("", "", ""),
		func(*storage.OrderedMap) (string, []any) { return "disabled", nil },
		func(string, float64) (int, []byte, error) {
			return 404, nil, nil
		})
	s.SyncMarkIntentional("s", []string{"https://alpha.com"}, "private",
		"secret", nil)
	s.SyncMarkIntentional("s", []string{"https://beta.com"}, "duplicate",
		"", nil)
	if got := len(s.SyncDeferredRows("s", "alpha")); got != 1 {
		t.Fatalf("search = %d", got)
	}
	if got := len(s.SyncDeferredRows("s", "Private")); got != 1 {
		t.Fatalf("label search = %d", got)
	}
	if got := s.SyncLedgerCount(); got != 2 {
		t.Fatalf("count = %d", got)
	}
}

// currentSnapshotForTest reads the held snapshot (the test's own eye).
func (s *Settings) currentSnapshotForTest() *SyncSnapshot {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.syncSnapshot
}

// ---------------------------------------------------------------------------
// THE ONE CONSENT DOOR — the Refresh's own flow (hermetic)
// ---------------------------------------------------------------------------

func TestRefreshSyncViewOnePayloadTwoDesks(t *testing.T) {
	// the flow's facts + the matrix's snapshot in ONE payload (v0.87.0's
	// convergence — the Go port rides the shared single gather the
	// reference records as its v2 follow-up)
	body := `{"counts": {"sites": 2}, "generated_at": "2026-10-01T00:00:00"}`
	s, rec := newSyncSettings(t, syncCfg("", "", ""),
		func(*storage.OrderedMap) (string, []any) {
			return "ok", []any{pendRow("https://github.com/a/one")}
		},
		func(string, float64) (int, []byte, error) {
			return 200, []byte(body), nil
		})
	payload := s.RefreshSyncView()
	if payload.Error != "" || payload.Model == nil || payload.Snapshot == nil {
		t.Fatalf("payload = %+v", payload)
	}
	// the flow cards measured: the bot card carries the live queue
	bot := payload.Model.Cards[0]
	if bot.Key != ElementBot || bot.State != SyncStateOK || bot.Pending != 1 {
		t.Fatalf("bot card = %+v", bot)
	}
	// the matrix measured: the queue stage is unknown (the Bot page's
	// own check has not run — the poll never lies into it)
	qrec := payload.Snapshot.Stages["repos:queue_to_vault"]
	if qrec.State != syncreview.StateUnknown || qrec.Needs != "measure" {
		t.Fatalf("queue rec = %+v", qrec)
	}
	// the published side measured: the website card speaks the site
	web := payload.Model.Cards[3]
	if web.State != SyncStateNotSet {
		// no websites vault configured in this config → not_set is honest
		t.Fatalf("web card = %+v", web)
	}
	// the banner + the measured-at stamp ride
	if payload.Banner.Tone == "" || payload.Measured == "" {
		t.Fatalf("banner/measured = %+v / %q", payload.Banner, payload.Measured)
	}
	// the desk speaks the refresh
	if _, ok := rec.findLog("Sync view refreshed"); !ok {
		t.Fatalf("the desk lines = %v", rec.logs)
	}
}

func TestRefreshSyncViewTheQueueCheckFeedsTheStages(t *testing.T) {
	// the Bot page's own check lands the queue state → the queue stages
	// carry the classified gap (the reference's own hook)
	body := `{"counts": {"sites": 0}}`
	s, _ := newSyncSettings(t, syncCfg("/v", "/w", "https://w.example"),
		func(*storage.OrderedMap) (string, []any) { return "disabled", nil },
		func(string, float64) (int, []byte, error) {
			return 200, []byte(body), nil
		})
	s.SetSyncQueueState(&SyncQueueState{Checked: true,
		PendingRepos:    []string{"https://github.com/a/one"},
		PendingWebsites: []string{"https://example.com/w"},
		InVault:         10, WebsitesInVault: 5})
	payload := s.RefreshSyncView()
	rrec := payload.Snapshot.Stages["repos:queue_to_vault"]
	wrec := payload.Snapshot.Stages["websites:queue_to_vault"]
	if rrec.State != syncreview.StateBehind || rrec.DisplayNumber == nil ||
		*rrec.DisplayNumber != 10 {
		t.Fatalf("repos queue rec = %+v", rrec)
	}
	if wrec.State != syncreview.StateBehind || *wrec.DisplayNumber != 5 {
		t.Fatalf("websites queue rec = %+v", wrec)
	}
}
