// commands_test.go — the command CONTRACT tests (T-71 slice (b)):
// the transport laws over fake deps (the hermetic law) — the events,
// the ask-gates, the one-batch lock, the result shapes, and the
// crash-safety wrapper (the TestWorker contract's Go twin).
package app

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/githubpipeline"
	"github.com/assadigit/second-brain-project/internal/goodrepos"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
	"github.com/assadigit/second-brain-project/internal/vaultseal"
)

// ---------------------------------------------------------------------------
// The harness: a recording bus subscriber + tiny fake deps.
// ---------------------------------------------------------------------------

// busRec records everything the bus says.
type busRec struct {
	mu       sync.Mutex
	logs     []LogLine
	statuses []string
	progress []Progress
	batches  []BatchFinished
	jobs     []JobProgress
	disks    []string
	models   []ModelChanged
}

func (r *busRec) subscribe(b *Bus) {
	b.Subscribe(&Handlers{
		LogLine:       func(l LogLine) { r.mu.Lock(); r.logs = append(r.logs, l); r.mu.Unlock() },
		Status:        func(s Status) { r.mu.Lock(); r.statuses = append(r.statuses, s.Text); r.mu.Unlock() },
		Progress:      func(p Progress) { r.mu.Lock(); r.progress = append(r.progress, p); r.mu.Unlock() },
		BatchFinished: func(f BatchFinished) { r.mu.Lock(); r.batches = append(r.batches, f); r.mu.Unlock() },
		JobProgress:   func(j JobProgress) { r.mu.Lock(); r.jobs = append(r.jobs, j); r.mu.Unlock() },
		DiskFull:      func(d DiskFull) { r.mu.Lock(); r.disks = append(r.disks, d.Path); r.mu.Unlock() },
		ModelChanged:  func(m ModelChanged) { r.mu.Lock(); r.models = append(r.models, m); r.mu.Unlock() },
	})
}

func (r *busRec) findLog(part string) (LogLine, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, l := range r.logs {
		if strings.Contains(l.Msg, part) {
			return l, true
		}
	}
	return LogLine{}, false
}

func (r *busRec) countBatches() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.batches)
}

func (r *busRec) snapshotBatches() []BatchFinished {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]BatchFinished(nil), r.batches...)
}

// tinyTimeouts keeps the gates fast if a test exercises them.
func tinyTimeouts(b *Bus) {
	b.CodeTimeout = 30 * time.Millisecond
	b.ScanTimeout = 30 * time.Millisecond
	b.LLMTimeout = 30 * time.Millisecond
}

// ---------------------------------------------------------------------------
// Sync.FetchQueue.
// ---------------------------------------------------------------------------

func TestSyncFetchQueueHappyPath(t *testing.T) {
	b := NewBus()
	tinyTimeouts(b)
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		QueueFetch: func(cfg *storage.OrderedMap) (string, []any) {
			return "ok", []any{map[string]any{"url": "https://github.com/o/r"}}
		},
		QueueClassify: func(cfg *storage.OrderedMap, items []any, log LogFn) *QueueReport {
			log("🤝 THE SETTLEMENT (repos): 2 settled", "info")
			n := 5
			return &QueueReport{
				GitHubURLs:      []string{"https://github.com/o/r"},
				PendingURLs:     []string{"https://github.com/o/r"},
				InVault:         2,
				Settled:         1,
				VaultIndex:      &n,
				PendingWebsites: []string{"https://example.com"},
			}
		},
	})
	rep := cmds.SyncFetchQueue(context.Background())
	if !rep.Success || rep.State != "ok" {
		t.Fatalf("report = success=%v state=%q err=%q", rep.Success, rep.State, rep.Error)
	}
	if len(rep.PendingURLs) != 1 || rep.InVault != 2 || rep.VaultIndex == nil || *rep.VaultIndex != 5 {
		t.Fatalf("classification fields lost: %+v", rep)
	}
	if _, ok := rec.findLog("THE SETTLEMENT"); !ok {
		t.Fatalf("the classify seam's log lines never rode the bus: %v", rec.logs)
	}
}

func TestSyncFetchQueueStateWords(t *testing.T) {
	for _, tc := range []struct {
		state string
		want  string
	}{
		{"disabled", "not configured"},
		{"error", "poll failed"},
	} {
		b := NewBus()
		cmds := NewCommands(b, Deps{
			QueueFetch: func(*storage.OrderedMap) (string, []any) { return tc.state, nil },
		})
		rep := cmds.SyncFetchQueue(context.Background())
		if rep.Success || rep.State != tc.state {
			t.Fatalf("state %q: report = %+v", tc.state, rep)
		}
		if !strings.Contains(rep.Error, tc.want) {
			t.Fatalf("state %q error = %q, want it to mention %q", tc.state, rep.Error, tc.want)
		}
	}
}

func TestSyncFetchQueueNilSeamRefusesHonestly(t *testing.T) {
	b := NewBus()
	cmds := NewCommands(b, Deps{})
	rep := cmds.SyncFetchQueue(context.Background())
	if rep.Success || rep.State != "error" {
		t.Fatalf("nil-seam report = %+v", rep)
	}
	if !strings.Contains(rep.Error, "not wired") {
		t.Fatalf("nil-seam error = %q", rep.Error)
	}
}

// ---------------------------------------------------------------------------
// Sync.RunBatch — the one-batch law, the event wiring, the crash law.
// ---------------------------------------------------------------------------

func TestSyncRunBatchEventWiringAndFinish(t *testing.T) {
	b := NewBus()
	tinyTimeouts(b)
	rec := &busRec{}
	rec.subscribe(b)
	var llmRepo string
	b.Subscribe(&Handlers{LLMFailed: func(e *LLMFailedEvent) {
		llmRepo = e.Repo
		e.Answer("skip")
	}})
	cmds := NewCommands(b, Deps{
		Batch: func(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
			ev.Log("engine says hi", "info")
			ev.Status("https://github.com/o/r")
			ev.Progress(1, 1)
			ev.DiskFull("/vault/full")
			ev.ModelChange("ollama", "llama3.1")
			if got := ev.LLMDecision("octo/r"); got != "skip" {
				t.Errorf("LLMDecision round-trip = %q, want skip", got)
			}
			if llmRepo != "octo/r" {
				t.Errorf("the LLM gate never saw the repo (got %q)", llmRepo)
			}
			return &BatchResult{Success: true, Message: "🏁 Done. Processed 1 repos.", Processed: 1, Total: 1}
		},
	})
	res := cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{"https://github.com/o/r"}, Source: "direct"})
	if !res.Success || res.Processed != 1 {
		t.Fatalf("result = %+v", res)
	}
	if rec.countBatches() != 1 {
		t.Fatalf("BatchFinished fired %d times, want exactly once", rec.countBatches())
	}
	fin := rec.snapshotBatches()[0]
	if !fin.OK || fin.Summary != "🏁 Done. Processed 1 repos." {
		t.Fatalf("BatchFinished = %+v", fin)
	}
	// The transport events all landed.
	if _, ok := rec.findLog("engine says hi"); !ok {
		t.Fatal("the engine's log never rode the bus")
	}
	if len(rec.statuses) == 0 || rec.statuses[len(rec.statuses)-1] != "https://github.com/o/r" {
		t.Fatalf("statuses = %v", rec.statuses)
	}
	if len(rec.progress) < 2 || rec.progress[len(rec.progress)-1] != (Progress{Pos: 1, Total: 1}) {
		t.Fatalf("progress = %v", rec.progress)
	}
	if len(rec.disks) != 1 || rec.disks[0] != "/vault/full" {
		t.Fatalf("disks = %v", rec.disks)
	}
	if len(rec.models) != 1 || rec.models[0].Model != "llama3.1" {
		t.Fatalf("models = %v", rec.models)
	}
	if rec.statuses[0] != "Processing…" {
		t.Fatalf("the start Status = %q, want Processing…", rec.statuses[0])
	}
}

// TestSyncRunBatchFlagClearLawAndFanfareFacts (T-79f) pins the
// reference's processing_finished opening (the flag clears BEFORE the
// finish handler runs — the redo pass's own re-run batch must never
// see a ghost batch) + the fanfare facts' ride (Natural/Stopped/
// WebsiteResults built from the result, exactly as the reference's
// batch_summary carries them).
func TestSyncRunBatchFlagClearLawAndFanfareFacts(t *testing.T) {
	b := NewBus()
	var fin BatchFinished
	var busyAtCrossing bool
	b.Subscribe(&Handlers{BatchFinished: func(ev BatchFinished) {
		fin = ev
	}})
	cmds := NewCommands(b, Deps{
		Batch: func(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
			// the busy-at-emit question is answered INSIDE the
			// crossing's own handler below (the emit is synchronous
			// on this goroutine — the poll there reads the live flag).
			return &BatchResult{Success: true, Message: "done", Processed: 2, Total: 2,
				Natural: true,
				WebsiteResults: []githubpipeline.WebsiteResultEntry{
					{URL: "https://example.com/failed", Outcome: "failed", Error: "reset"},
				}}
		},
	})
	// the crossing reads the live flag DURING the emit (the emit is
	// synchronous on this goroutine): the flag-clear law says it must
	// already be false — the reference's `self._batch_running = False`
	// runs at processing_finished's top, before the fanfare chain.
	b.Subscribe(&Handlers{BatchFinished: func(ev BatchFinished) {
		busyAtCrossing = cmds.BatchBusy()
	}})
	res := cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{"u"}})
	if !res.Success {
		t.Fatalf("result = %+v", res)
	}
	if fin.OK != true || fin.Natural != true || fin.Stopped != false {
		t.Fatalf("the fanfare facts = %+v", fin)
	}
	if len(fin.WebsiteResults) != 1 ||
		fin.WebsiteResults[0].URL != "https://example.com/failed" {
		t.Fatalf("the website results = %+v", fin.WebsiteResults)
	}
	if busyAtCrossing {
		t.Fatal("THE FLAG-CLEAR LAW: BatchBusy() must be false inside the BatchFinished handler")
	}
	if cmds.BatchBusy() {
		t.Fatal("the flag must clear after the batch")
	}
	if fin.NeedsRetry {
		t.Fatal("no failed links — no retry chime")
	}
}

func TestSyncRunBatchOneAtATime(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	started := make(chan struct{})
	release := make(chan struct{})
	var callCount int
	var mu sync.Mutex
	var startedOnce sync.Once
	cmds := NewCommands(b, Deps{
		Batch: func(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
			mu.Lock()
			callCount++
			mu.Unlock()
			startedOnce.Do(func() { close(started) })
			<-release // closed on release: every later call returns at once
			return &BatchResult{Success: true, Message: "done"}
		},
	})
	go func() { _ = cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{"u"}}) }()
	<-started

	// A second batch while the first runs: refused, honestly, without
	// touching the engine (callCount stays 1).
	second := cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{"u2"}})
	if second.Success {
		t.Fatal("the second batch ran while the first held the lock")
	}
	if !strings.Contains(second.Message, "already running") {
		t.Fatalf("refusal = %q, want the already-running line", second.Message)
	}
	mu.Lock()
	if callCount != 1 {
		t.Fatalf("the engine ran %d times, want 1", callCount)
	}
	mu.Unlock()
	if l, ok := rec.findLog("already running"); !ok || l.Level != "warning" {
		t.Fatalf("the refusal line never landed: %v", rec.logs)
	}

	// Release; the lock goes home; a third batch runs.
	close(release)
	deadline := time.After(5 * time.Second)
	for cmds.batchLock.Busy() {
		select {
		case <-deadline:
			t.Fatal("the lock never went home after the batch finished")
		case <-time.After(5 * time.Millisecond):
		}
	}
	third := cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{"u3"}})
	if !third.Success {
		t.Fatalf("the third batch was refused after the lock went home: %+v", third)
	}
	if rec.countBatches() != 2 {
		t.Fatalf("BatchFinished count = %d, want 2", rec.countBatches())
	}
}

func TestSyncRunBatchCrashSafetyAndFinishExactlyOnce(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		Batch: func(ctx context.Context, cfg *storage.OrderedMap, opts BatchOptions, ev BatchEvents) *BatchResult {
			ev.Log("about to explode", "info")
			panic("engine exploded")
		},
	})
	res := cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{"u"}})
	if res.Success || !res.Crashed {
		t.Fatalf("crash result = %+v, want crashed", res)
	}
	if l, ok := rec.findLog("Command 'Sync.RunBatch' crashed"); !ok || l.Level != "error" {
		t.Fatalf("the crash line never landed: %v", rec.logs)
	}
	if rec.countBatches() != 1 {
		t.Fatalf("BatchFinished fired %d times on the crash path, want exactly once", rec.countBatches())
	}
	if fin := rec.snapshotBatches()[0]; fin.OK {
		t.Fatalf("the crash's BatchFinished = %+v, want ok=false", fin)
	}
	// The lock went home despite the crash — the next batch is NOT
	// refused with the already-running line (it runs the same panicking
	// engine again and crashes again, honestly — the crash-freedom is
	// the proof the lock was released).
	next := cmds.SyncRunBatch(context.Background(), BatchOptions{URLs: []string{}})
	if strings.Contains(next.Message, "already running") {
		t.Fatalf("post-crash batch refused: %+v — the crash leaked the lock", next)
	}
	if !next.Crashed {
		t.Fatalf("post-crash batch = %+v, want the second honest crash", next)
	}
}

func TestSyncRunBatchNilSeamRefusedButLockClean(t *testing.T) {
	b := NewBus()
	cmds := NewCommands(b, Deps{})
	res := cmds.SyncRunBatch(context.Background(), BatchOptions{})
	if res.Success || !strings.Contains(res.Message, "not wired") {
		t.Fatalf("nil-seam batch = %+v", res)
	}
	if cmds.batchLock.Busy() {
		t.Fatal("the nil-seam refusal leaked the batch lock")
	}
}

// ---------------------------------------------------------------------------
// Sync.FastSync / Sync.DriftSeal — the manual door + the quiet law.
// ---------------------------------------------------------------------------

func TestDriftSealManualAndAutoDoors(t *testing.T) {
	var gotManual []bool
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		DriftSeal: func(cfg *storage.OrderedMap, manual bool, log LogFn) (string, string) {
			gotManual = append(gotManual, manual)
			if manual {
				return "✅ Fast Sync: the triad is already in sync — nothing to seal, nothing to publish", "success"
			}
			return "🛡️ Drift seal: 2 file(s) → the repos vault mirror", "success"
		},
	})
	if r := cmds.SyncFastSync(context.Background()); !r.Success || r.Level != "success" {
		t.Fatalf("fast sync = %+v", r)
	}
	if r := cmds.SyncDriftSeal(context.Background()); !r.Success || r.Level != "success" {
		t.Fatalf("drift seal = %+v", r)
	}
	if len(gotManual) != 2 || !gotManual[0] || gotManual[1] {
		t.Fatalf("manual door routing = %v, want [true false]", gotManual)
	}
	if _, ok := rec.findLog("triad is already in sync"); !ok {
		t.Fatal("the manual verdict never landed on the bus")
	}
}

func TestDriftSealQuietLawAndWarningPath(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		DriftSeal: func(cfg *storage.OrderedMap, manual bool, log LogFn) (string, string) {
			if manual {
				return "", "" // the manual door must never see this (engine bug)
			}
			return "", "info" // the quiet automatic path
		},
	})
	r := cmds.SyncDriftSeal(context.Background())
	if !r.Success || r.Message != "" {
		t.Fatalf("the quiet law broke: %+v", r)
	}
	if _, ok := rec.findLog("Drift seal"); ok {
		t.Fatal("the quiet automatic path spoke — the quiet law is silent")
	}
	// The manual door with an empty verdict refuses honestly (the
	// explicit press always deserves its verdict).
	b2 := NewBus()
	cmds2 := NewCommands(b2, Deps{
		DriftSeal: func(cfg *storage.OrderedMap, manual bool, log LogFn) (string, string) {
			return "", ""
		},
	})
	if r := cmds2.SyncFastSync(context.Background()); r.Success || r.Level != "error" {
		t.Fatalf("silent manual door = %+v, want the honest refusal", r)
	}
	// The warning verdict lands at warning level.
	b3 := NewBus()
	rec3 := &busRec{}
	rec3.subscribe(b3)
	cmds3 := NewCommands(b3, Deps{
		DriftSeal: func(cfg *storage.OrderedMap, manual bool, log LogFn) (string, string) {
			return "⚠️ Fast Sync: the directory publish did not land", "warning"
		},
	})
	if r := cmds3.SyncFastSync(context.Background()); r.Success || r.Level != "warning" {
		t.Fatalf("warning fast sync = %+v", r)
	}
	if l, ok := rec3.findLog("did not land"); !ok || l.Level != "warning" {
		t.Fatalf("the warning never landed at warning: %v", rec3.logs)
	}
}

// ---------------------------------------------------------------------------
// Scan.Run + Scan.ApplyPlan — the confirm-gate round-trip.
// ---------------------------------------------------------------------------

func TestScanRunConfirmGateRoundTrip(t *testing.T) {
	b := NewBus()
	tinyTimeouts(b)
	rec := &busRec{}
	rec.subscribe(b)
	// The GUI side: receives the plan, answers confirmed with a
	// selection.
	b.Subscribe(&Handlers{AskScanConfirm: func(e *AskScanConfirmEvent) {
		plan, ok := e.Plan.(map[string]any)
		if !ok || plan["kind"] != "repos" {
			t.Errorf("the plan never rode the gate: %v", e.Plan)
		}
		e.Answer("CONFIRMED", map[string][]string{"deletions": {"d1"}})
	}})
	var gateAnswer ScanAnswer
	cmds := NewCommands(b, Deps{
		Scan: func(ctx context.Context, cfg *storage.OrderedMap, vaults []string, ev ScanEvents) *ScanResult {
			ev.Log("📋 the plan", "info")
			ev.Progress("inventory walk", 3, 9)
			gateAnswer = ev.Confirm(map[string]any{"kind": "repos"})
			return &ScanResult{Success: true, Verdict: gateAnswer.Verdict, HygieneTotal: 2}
		},
	})
	res := cmds.ScanRun(context.Background(), []string{"repos"})
	if !res.Success || res.Verdict != ScanConfirmed {
		t.Fatalf("scan result = %+v", res)
	}
	if gateAnswer.Selection == nil || len(gateAnswer.Selection["deletions"]) != 1 {
		t.Fatalf("the picked rows never rode the gate answer: %+v", gateAnswer)
	}
	if _, ok := rec.findLog("the plan"); !ok {
		t.Fatal("the scan's log lines never rode the bus")
	}
	if len(rec.jobs) == 0 || rec.jobs[0].Label != "inventory walk" {
		t.Fatalf("the honest pipeline row never rode the bus: %v", rec.jobs)
	}
}

func TestScanRunNilSeamAndCrash(t *testing.T) {
	b := NewBus()
	cmds := NewCommands(b, Deps{})
	if res := cmds.ScanRun(context.Background(), nil); res.Success || res.Verdict != "error" {
		t.Fatalf("nil-seam scan = %+v", res)
	}
	cmds2 := NewCommands(b, Deps{
		Scan: func(context.Context, *storage.OrderedMap, []string, ScanEvents) *ScanResult {
			panic("scan exploded")
		},
	})
	res := cmds2.ScanRun(context.Background(), nil)
	if res.Success || !res.Crashed {
		t.Fatalf("crash scan = %+v, want crashed", res)
	}
}

func TestScanApplyPlanCarriesThePickingDoor(t *testing.T) {
	b := NewBus()
	var gotKind string
	var gotSel map[string][]string
	var gotPlan *vaultscan.Plan
	cmds := NewCommands(b, Deps{
		ApplyPlan: func(cfg *storage.OrderedMap, kind string, plan *vaultscan.Plan, sel map[string][]string, log LogFn) *ApplyResult {
			gotKind, gotSel, gotPlan = kind, sel, plan
			log("applied", "success")
			return &ApplyResult{Success: true}
		},
	})
	res := cmds.ScanApplyPlan(context.Background(), "repos", &vaultscan.Plan{}, map[string][]string{"moves": {"m1"}})
	if !res.Success {
		t.Fatalf("apply = %+v", res)
	}
	if gotKind != "repos" || gotPlan == nil || len(gotSel["moves"]) != 1 {
		t.Fatalf("apply carried kind=%q plan=%v sel=%v", gotKind, gotPlan, gotSel)
	}
}

// ---------------------------------------------------------------------------
// Backup.* — the starters.
// ---------------------------------------------------------------------------

func TestBackupSealVaultKindRouting(t *testing.T) {
	var kinds []string
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		SealVault: func(cfg *storage.OrderedMap, kind string, log LogFn) *vaultseal.SealResult {
			kinds = append(kinds, kind)
			r := vaultseal.NewSealResult()
			r.Sealed = true
			r.FilesChanged = 1
			r.RepoName = "octo/" + kind
			return r
		},
	})
	for _, k := range []string{"repos", "websites", "personal"} {
		if r := cmds.BackupSealVault(context.Background(), k); !r.Success {
			t.Fatalf("seal %s = %+v", k, r)
		}
	}
	if len(kinds) != 3 || kinds[0] != "repos" || kinds[1] != "websites" || kinds[2] != "personal" {
		t.Fatalf("kind routing = %v", kinds)
	}
	// Unknown kind: the honest refusal, engine untouched.
	if r := cmds.BackupSealVault(context.Background(), "bogus"); r.Success || !strings.Contains(r.Message, "unknown seal kind") {
		t.Fatalf("bogus kind = %+v", r)
	}
	if len(kinds) != 3 {
		t.Fatalf("the bogus kind reached the engine: %v", kinds)
	}
}

func TestBackupResyncNowPublishVerdicts(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		Resync: func(cfg *storage.OrderedMap, log LogFn) *storage.OrderedMap {
			return storage.NewOrderedMap()
		},
		BackupNow: func(cfg *storage.OrderedMap, log LogFn) *SealResultRow {
			return &SealResultRow{Success: true, Message: "💾 Backup created: vault-x.zip (12 files, 3.4 KB)"}
		},
		Publish: func(cfg *storage.OrderedMap, log LogFn) *goodrepos.PublishResult {
			r := goodrepos.NewPublishResult()
			r.Published = true
			r.Entries = 42
			return r
		},
	})
	if r := cmds.BackupResync(context.Background()); !r.Success {
		t.Fatalf("resync = %+v", r)
	}
	if r := cmds.BackupNow(context.Background()); !r.Success {
		t.Fatalf("backup now = %+v", r)
	}
	if l, ok := rec.findLog("Backup created"); !ok || l.Level != "success" {
		t.Fatalf("the backup verdict line = %v", rec.logs)
	}
	if r := cmds.BackupPublish(context.Background()); !r.Success {
		t.Fatalf("publish = %+v", r)
	}
}

func TestBackupNilSeamsRefuseHonestly(t *testing.T) {
	b := NewBus()
	cmds := NewCommands(b, Deps{})
	for name, r := range map[string]*SealResultRow{
		"seal":   cmds.BackupSealVault(context.Background(), "repos"),
		"resync": cmds.BackupResync(context.Background()),
		"now":    cmds.BackupNow(context.Background()),
		"pub":    cmds.BackupPublish(context.Background()),
	} {
		if r.Success || !strings.Contains(r.Message, "not wired") {
			t.Fatalf("%s with nil seam = %+v, want the honest refusal", name, r)
		}
	}
}

// ---------------------------------------------------------------------------
// Radar.Detect — the starter's guards + the job's story.
// ---------------------------------------------------------------------------

func TestRadarDetectStarterGuards(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cmds := NewCommands(b, Deps{
		Radar: func(cfg *storage.OrderedMap, log LogFn) *RadarResult {
			t.Fatal("the radar engine ran past a guard")
			return nil
		},
	})
	// No token.
	res := cmds.RadarDetect(context.Background())
	if res.Success || !strings.Contains(res.Error, "set the token first") {
		t.Fatalf("no-token guard = %+v", res)
	}
	// Token but no vaults.
	cmds2 := NewCommands(b, Deps{
		Config: func() *storage.OrderedMap {
			cfg := storage.NewOrderedMap()
			cfg.Set("github_token", "tok")
			return cfg
		},
		Radar: func(cfg *storage.OrderedMap, log LogFn) *RadarResult {
			t.Fatal("the radar engine ran past the no-vaults guard")
			return nil
		},
	})
	if res := cmds2.RadarDetect(context.Background()); res.Success || !strings.Contains(res.Error, "vault path") {
		t.Fatalf("no-vault guard = %+v", res)
	}
	if _, ok := rec.findLog("Repo radar: scanning"); ok {
		t.Fatal("the opening line fired past a guard")
	}
}

func TestRadarDetectHappyPath(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	cfg := storage.NewOrderedMap()
	cfg.Set("github_token", "tok")
	cfg.Set("vault_path", "/vault/repos")
	cmds := NewCommands(b, Deps{
		Config: func() *storage.OrderedMap { return cfg },
		Radar: func(cfg *storage.OrderedMap, log LogFn) *RadarResult {
			log("📡 repos → octo/backup (fingerprint, 12/12)", "info")
			return &RadarResult{Success: true, Applied: 1, Results: map[string]any{
				"repos": map[string]any{"repo_name": "octo/backup"},
			}, Lines: []string{"📡 repos → octo/backup"}}
		},
	})
	res := cmds.RadarDetect(context.Background())
	if !res.Success || res.Applied != 1 || len(res.Lines) != 1 {
		t.Fatalf("radar = %+v", res)
	}
	if _, ok := rec.findLog("Repo radar: scanning your GitHub account"); !ok {
		t.Fatal("the opening line never landed")
	}
	if _, ok := rec.findLog("fingerprint"); !ok {
		t.Fatal("the radar's own log lines never rode the bus")
	}
}

// ---------------------------------------------------------------------------
// The crash-safety law across the surface.
// ---------------------------------------------------------------------------

func TestEveryCommandRecoversPanics(t *testing.T) {
	b := NewBus()
	rec := &busRec{}
	rec.subscribe(b)
	boom := func() { panic("boom") }
	cmds := NewCommands(b, Deps{
		QueueFetch: func(*storage.OrderedMap) (string, []any) { boom(); return "", nil },
		DriftSeal:  func(*storage.OrderedMap, bool, LogFn) (string, string) { boom(); return "", "" },
		Scan:       func(context.Context, *storage.OrderedMap, []string, ScanEvents) *ScanResult { boom(); return nil },
		ApplyPlan: func(*storage.OrderedMap, string, *vaultscan.Plan, map[string][]string, LogFn) *ApplyResult {
			boom()
			return nil
		},
		SealVault: func(*storage.OrderedMap, string, LogFn) *vaultseal.SealResult { boom(); return nil },
		Resync:    func(*storage.OrderedMap, LogFn) *storage.OrderedMap { boom(); return nil },
		BackupNow: func(*storage.OrderedMap, LogFn) *SealResultRow { boom(); return nil },
		Publish:   func(*storage.OrderedMap, LogFn) *goodrepos.PublishResult { boom(); return nil },
		Radar:     func(*storage.OrderedMap, LogFn) *RadarResult { boom(); return nil },
	})
	ctx := context.Background()
	cfg := storage.NewOrderedMap()
	cfg.Set("github_token", "t")
	cfg.Set("vault_path", "/v")

	// None of these may panic across the boundary; each lands its
	// honest crash line instead.
	check := func(name string, fn func()) {
		defer func() {
			if r := recover(); r != nil {
				t.Fatalf("%s panicked across the boundary: %v", name, r)
			}
		}()
		fn()
		if _, ok := rec.findLog("crashed"); !ok {
			t.Fatalf("%s: no crash line landed (%v)", name, rec.logs)
		}
	}
	check("FetchQueue", func() { cmds.SyncFetchQueue(ctx) })
	check("FastSync", func() { cmds.SyncFastSync(ctx) })
	check("Scan", func() { cmds.ScanRun(ctx, nil) })
	check("ApplyPlan", func() { cmds.ScanApplyPlan(ctx, "repos", nil, nil) })
	check("SealVault", func() { cmds.BackupSealVault(ctx, "repos") })
	check("Resync", func() { cmds.BackupResync(ctx) })
	check("BackupNow", func() { cmds.BackupNow(ctx) })
	check("Publish", func() { cmds.BackupPublish(ctx) })
	check("Radar", func() { cmds.RadarDetect(ctx) })
	if cmds.batchLock.Busy() {
		t.Fatal("a crashed command leaked the batch lock")
	}
}
