// syncuib_test.go — T-75's Sync page binding wrapper wiring smoke:
// the ONE Refresh door's payload (both desks in one answer), the fix
// door's queue-check leg landing the session's queue state, the
// differences modal's doors round-trip, and the SECTIONS row's own
// laws (the sidebar order + the 'refresh' glyph — the engine laws
// live in internal/app's batteries; this is the shell-side proof the
// page's doors are real).
package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/app"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/syncreview"
	"github.com/assadigit/second-brain-project/webui/uitheme"
)

// newSyncUI builds a hermetic wrapper over fakes (the queue poll +
// the published GET answer canned data — the hermetic law at the
// shell).
func newSyncUI(t *testing.T) (*SyncUI, *app.Settings, *[]string) {
	t.Helper()
	appDir := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appDir)
	bus := app.NewBus()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", t.TempDir())
	cfg.Set("website_vault_path", t.TempDir())
	cfg.Set("directory_worker_url", "https://worker.example")
	settings := app.NewSettings(bus, cfg, &app.OpLock{}, app.SettingsDeps{})
	var spoken []string
	bus.Subscribe(&app.Handlers{LogLine: func(l app.LogLine) {
		spoken = append(spoken, l.Msg)
	}})
	settings.SetSyncDeps(app.SyncDeps{
		QueuePoll: func(*storage.OrderedMap) (string, []any) {
			return "ok", []any{
				map[string]any{"url": "https://github.com/octo/pending"},
				map[string]any{"url": "https://example.com/page"},
			}
		},
		PublishedFetch: func(url string, timeoutS float64) (int, []byte, error) {
			return 404, nil, nil // the Worker's honest "never published"
		},
		LedgerOpen: func() (*syncreview.DeferredLog, error) {
			return syncreview.OpenLedger(appDir + "/sync_review.db")
		},
		GatherDeps: &app.SyncGatherDeps{
			SiteItems: func(vault string, log func(string), cap int) ([]syncreview.GapItem, bool) {
				return []syncreview.GapItem{
					{ID: "https://a.com", Name: "A", URL: "https://a.com", AddedAt: "2026-09-01"},
					{ID: "https://c.com", Name: "C", URL: "https://c.com", AddedAt: "2026-09-15"},
				}, true
			},
			BackupGap: func(vault string, log func(string), cap int) ([]syncreview.GapItem, bool) {
				return []syncreview.GapItem{}, true
			},
			NoteCount: func(vault string) *int {
				n := 3
				return &n
			},
			Published: func(workerURL string, timeoutS float64) *storage.OrderedMap {
				idx := storage.NewOrderedMap()
				idx.Set("sites", []any{})
				return idx
			},
		},
	})
	// Prime the settlement (the one-time law: a FRESH cache would
	// settle the whole queue history — the pages B harness's own step).
	{
		cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db"))
		if err != nil {
			t.Fatalf("prime store: %v", err)
		}
		if _, err := cache.SettleExistingRepos(nil); err != nil {
			t.Fatal(err)
		}
		_ = cache.Close()
	}
	commands := app.NewCommands(bus, app.Deps{
		Config: func() *storage.OrderedMap { return cfg },
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) {
			return "ok", []any{
				map[string]any{"url": "https://github.com/octo/pending"},
				map[string]any{"url": "https://example.com/page"},
			}
		},
		QueueClassify: app.ClassifyQueueLinks,
	})
	return NewSyncUI(settings, commands), settings, &spoken
}

func TestSyncUIRefreshOnePayloadTwoDesks(t *testing.T) {
	ui, _, spoken := newSyncUI(t)
	payload := ui.Refresh()
	if payload == nil || payload.Model == nil || payload.Snapshot == nil {
		t.Fatalf("payload = %+v (one payload, both desks)", payload)
	}
	// the flow measured: the bot card carries the live queue, the
	// website card speaks the never-published truth
	if payload.Model.Cards[0].Key != app.ElementBot ||
		payload.Model.Cards[0].State != app.SyncStateOK ||
		payload.Model.Cards[0].Pending != 2 {
		t.Fatalf("bot card = %+v", payload.Model.Cards[0])
	}
	if payload.Model.Cards[3].State != app.SyncStateOK {
		t.Fatalf("website card = %+v (a live vault + never-published site is ok)", payload.Model.Cards[3])
	}
	// the matrix measured: all six stages + the banner speaks
	if len(payload.Snapshot.Stages) != 6 {
		t.Fatalf("stages = %d", len(payload.Snapshot.Stages))
	}
	if payload.Banner.Tone == "" || payload.Measured == "" {
		t.Fatalf("banner/measured = %+v / %q", payload.Banner, payload.Measured)
	}
	if !contains(joinMsgs(spoken), "Sync view refreshed") {
		t.Fatalf("the desk lines = %v", spoken)
	}
}

func TestSyncUICheckBotQueueLandsTheQueueState(t *testing.T) {
	ui, settings, _ := newSyncUI(t)
	// the queue stages start unmeasured (the quiet-open law)
	payload := ui.Refresh()
	if rec := payload.Snapshot.Stages["repos:queue_to_vault"]; rec.State != syncreview.StateUnknown {
		t.Fatalf("queue rec = %+v (not checked this session)", rec)
	}
	// the fix door: the check runs + the state lands
	res := ui.CheckBotQueue()
	if res.Refusal != "" || res.Error != "" {
		t.Fatalf("res = %+v", res)
	}
	payload = ui.Refresh()
	rec := payload.Snapshot.Stages["repos:queue_to_vault"]
	if rec.State != syncreview.StateBehind || rec.SetUp != true {
		t.Fatalf("queue rec after the check = %+v (the classified gap landed)", rec)
	}
	_ = settings
}

func TestSyncUIMarkAndReopenDoors(t *testing.T) {
	ui, _, _ := newSyncUI(t)
	// seed a behind stage via the refresh (the websites stage carries
	// the never-published gap: every public-ready note is the gap)
	payload := ui.Refresh()
	stage := payload.Snapshot.Stages["websites:vault_to_site"]
	if stage.State != syncreview.StateBehind {
		t.Fatalf("stage = %+v (the fixture vault carries site notes)", stage)
	}
	// the modal's data door
	diff := ui.StageDifferences("websites:vault_to_site")
	if !diff.Found || diff.Stage == nil || diff.Heading == "" {
		t.Fatalf("diff = %+v", diff)
	}
	if !strings.Contains(diff.Heading, "In the vault, not in the website") {
		t.Fatalf("heading = %q", diff.Heading)
	}
	// the unarmed mark refuses; the armed mark lands + confirms
	if res := ui.MarkIntentional("websites:vault_to_site", nil, "private", "", nil); res.Marked != 0 {
		t.Fatalf("unarmed mark = %+v", res)
	}
	ids := make([]string, 0, len(stage.Items))
	names := map[string]string{}
	for _, it := range stage.Items {
		ids = append(ids, it.ID)
		names[it.ID] = it.Name
	}
	if len(ids) == 0 {
		t.Fatal("the fixture stage carries no items")
	}
	res := ui.MarkIntentional("websites:vault_to_site", ids, "private", "the note", names)
	if res.Marked != len(ids) || !strings.Contains(res.Line, "marked — Private") {
		t.Fatalf("mark = %+v", res)
	}
	// the stage turned Reviewed (the recompute behind the modal)
	payload = ui.Refresh()
	if rec := payload.Snapshot.Stages["websites:vault_to_site"]; rec.State != syncreview.StateReviewed {
		t.Fatalf("rec after mark = %+v", rec)
	}
	// the reopen: one item back in the Differences tab
	rres := ui.ReopenItem("websites:vault_to_site", ids[0])
	if !rres.Ok || !strings.Contains(rres.Line, "reopened") {
		t.Fatalf("reopen = %+v", rres)
	}
	// the log doors
	if rows := ui.DeferredRows("websites:vault_to_site", ""); len(rows) != len(ids) {
		t.Fatalf("rows = %d", len(rows))
	}
	if ui.LedgerCount() != len(ids) {
		t.Fatalf("ledger count = %d", ui.LedgerCount())
	}
}

func TestSyncUIStageTablesShape(t *testing.T) {
	ui, _, _ := newSyncUI(t)
	tables := ui.StageTables()
	if len(tables.Defs) != 6 || len(tables.Streams) != 3 || len(tables.Columns) != 3 {
		t.Fatalf("tables = %+v", tables)
	}
	// the layout's rows carry the reference's own shape (the None
	// slots land as empty strings)
	if len(tables.Layout["Repos"]) != 3 || tables.Layout["Repos"][1] != "" {
		t.Fatalf("repos layout = %v", tables.Layout["Repos"])
	}
	if tables.Layout["Personal"][0] != "" || tables.Layout["Personal"][1] != "" ||
		tables.Layout["Personal"][2] != "personal:vault_to_backup" {
		t.Fatalf("personal layout = %v", tables.Layout["Personal"])
	}
}

func TestSyncSectionSidebarLaws(t *testing.T) {
	// the 'refresh' glyph is the pack's own (the T-72 golden locked
	// the pack; the Sync row renders from it)
	if _, ok := uitheme.Icons["refresh"]; !ok {
		t.Fatal("the Sync row needs the 'refresh' glyph — missing from the pack")
	}
	// the SECTIONS table carries the Sync row at the reference's own
	// order (between Dashboard and Bot — dialogs.py's TAB_INFO), with
	// the reference's own hint
	js, err := readFrontend("settings.js")
	if err != nil {
		t.Fatal(err)
	}
	for _, law := range []string{
		"{ id: 'dashboard', label: 'Dashboard'",
		"{ id: 'sync', label: 'Sync', icon: 'refresh'",
		"{ id: 'bot', label: 'Bot'",
		"The four elements' flow + the stage-by-stage gap matrix",
		"buildSync,",
	} {
		if !contains(js, law) {
			t.Fatalf("settings.js missing %q", law)
		}
	}
	dashAt := strings.Index(js, "{ id: 'dashboard'")
	syncAt := strings.Index(js, "{ id: 'sync'")
	botAt := strings.Index(js, "{ id: 'bot'")
	if !(dashAt < syncAt && syncAt < botAt) {
		t.Fatal("the Sync row must sit between Dashboard and Bot")
	}
	// the first-show door + the one-refresh law ride the page
	if !contains(js, "syncShownOnce") || !contains(js, "refreshSyncView") {
		t.Fatal("the first-show door law is missing")
	}
	// the page's own texts ride verbatim (the reference's own words)
	for _, law := range []string{
		"Sync — the pipeline at a glance.",
		"The chain, stage by stage — Telegram queue → vaults → backups → ",
		"the public website. Every cell speaks its number AND what the ",
		"How to read a cell:",
		"Mark as intentional",
		"Nothing is ever deleted from the log",
	} {
		if !contains(js, law) {
			t.Fatalf("settings.js missing the verbatim law %q", law)
		}
	}
}

// readFrontend reads one of the frontend's source files (the section
// laws live in the JS — the shell test reads them where they render).
func readFrontend(name string) (string, error) {
	data, err := os.ReadFile("frontend/" + name)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func joinMsgs(msgs *[]string) string {
	if msgs == nil {
		return ""
	}
	return strings.Join(*msgs, "\n")
}
