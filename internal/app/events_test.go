// events_test.go — the bus contract tests (T-71 slice (a)): fan-out
// order, the never-raise law, the ask-gate round-trips with the
// reference's timeout/headless/no-reader laws (tiny injected windows
// — the production constants are pinned by their own test).
package app

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
)

// tinyBus builds a bus with millisecond windows (the instrument; the
// production defaults are asserted separately).
func tinyBus() *Bus {
	return &Bus{
		CodeTimeout: 30 * time.Millisecond,
		ScanTimeout: 30 * time.Millisecond,
		LLMTimeout:  30 * time.Millisecond,
	}
}

// recLog collects every LogLine one subscriber sees, in order.
type recLog struct {
	mu    sync.Mutex
	lines []LogLine
}

func (r *recLog) log(l LogLine) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.lines = append(r.lines, l)
}

func (r *recLog) find(msgPart string) (LogLine, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, l := range r.lines {
		if strings.Contains(l.Msg, msgPart) {
			return l, true
		}
	}
	return LogLine{}, false
}

func TestNewPinsTheReferenceTimeouts(t *testing.T) {
	b := NewBus()
	if b.CodeTimeout != 300*time.Second {
		t.Fatalf("CodeTimeout = %v, want the reference's 300 s", b.CodeTimeout)
	}
	if b.ScanTimeout != 300*time.Second {
		t.Fatalf("ScanTimeout = %v, want the reference's 300 s", b.ScanTimeout)
	}
	if b.LLMTimeout != 600*time.Second {
		t.Fatalf("LLMTimeout = %v, want the reference's 600 s", b.LLMTimeout)
	}
	if b.Headless {
		t.Fatal("New() must not be headless")
	}
}

func TestBusFanOutOrderAndKinds(t *testing.T) {
	b := tinyBus()
	var got []string
	var mu sync.Mutex
	add := func(s string) { mu.Lock(); got = append(got, s); mu.Unlock() }
	cancel := b.Subscribe(&Handlers{
		LogLine:       func(LogLine) { add("log") },
		Progress:      func(Progress) { add("prog") },
		Status:        func(Status) { add("status") },
		BatchFinished: func(BatchFinished) { add("batch") },
		JobProgress:   func(JobProgress) { add("job") },
		DiskFull:      func(DiskFull) { add("disk") },
		ModelChanged:  func(ModelChanged) { add("model") },
	})
	b.EmitStatus("running")
	b.EmitProgress(1, 4)
	b.EmitLogLine("hello", "info")
	b.EmitJobProgress("walk", 1, 10)
	b.EmitDiskFull("/vault/x")
	b.EmitModelChanged("ollama", "llama3")
	b.EmitBatchFinished(&BatchResult{Success: true, Message: "done"})
	cancel()
	b.EmitStatus("nobody hears this")

	want := []string{"status", "prog", "log", "job", "disk", "model", "batch"}
	if len(got) != len(want) {
		t.Fatalf("got %d events, want %d: %v", len(got), len(want), got)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("event[%d] = %q, want %q (full: %v)", i, got[i], want[i], got)
		}
	}
}

func TestJobProgressJunkClamped(t *testing.T) {
	b := tinyBus()
	var jp JobProgress
	b.Subscribe(&Handlers{JobProgress: func(j JobProgress) { jp = j }})
	b.EmitJobProgress("", -1, -2)
	if jp.Value != 0 || jp.Max != 0 {
		t.Fatalf("junk JobProgress = %+v, want zeros (the never-kill law)", jp)
	}
}

func TestBusMultipleSubscribersAndCancel(t *testing.T) {
	b := tinyBus()
	var a, c int
	ca := b.Subscribe(&Handlers{LogLine: func(LogLine) { a++ }})
	b.Subscribe(&Handlers{LogLine: func(LogLine) { c++ }})
	b.EmitLogLine("one", "info")
	ca()
	b.EmitLogLine("two", "info")
	if a != 1 {
		t.Fatalf("cancelled subscriber heard %d, want 1", a)
	}
	if c != 2 {
		t.Fatalf("live subscriber heard %d, want 2", c)
	}
	ca() // idempotent cancel
	if a != 1 {
		t.Fatalf("double cancel changed delivery: %d", a)
	}
}

func TestBusSubscriberPanicIsolated(t *testing.T) {
	b := tinyBus()
	b.Subscribe(&Handlers{LogLine: func(LogLine) { panic("subscriber exploded") }})
	var heard bool
	b.Subscribe(&Handlers{Status: func(Status) { heard = true }})
	// The panicking LogLine subscriber must not kill the emit loop:
	// the NEXT event still lands, and no panic escapes the call.
	b.EmitLogLine("boom", "info")
	b.EmitStatus("still alive")
	if !heard {
		t.Fatal("the Status event never landed after a subscriber panic")
	}
}

func TestAskCodeRoundTrip(t *testing.T) {
	b := tinyBus()
	b.Subscribe(&Handlers{AskCode: func(e *AskCodeEvent) {
		if e.Kind != KindPassword {
			t.Errorf("kind = %q, want PASSWORD", e.Kind)
		}
		e.Answer("12345")
	}})
	if got := b.AskCode(KindPassword); got != "12345" {
		t.Fatalf("AskCode = %q, want 12345", got)
	}
}

func TestAskCodeDefaultKindAndDoubleAnswerDropped(t *testing.T) {
	b := tinyBus()
	b.Subscribe(&Handlers{AskCode: func(e *AskCodeEvent) {
		if e.Kind != KindCode {
			t.Errorf("default kind = %q, want CODE", e.Kind)
		}
		e.Answer("first")
		e.Answer("second") // must not block (buffered 1, non-blocking)
	}})
	if got := b.AskCode(""); got != "first" {
		t.Fatalf("AskCode = %q, want first", got)
	}
}

func TestAskCodeTimeoutFailsFastToEmpty(t *testing.T) {
	b := tinyBus()
	rec := &recLog{}
	// An ask subscriber that never answers: the full window closes
	// (the no-reader guard does not apply — someone IS listening).
	b.Subscribe(&Handlers{LogLine: rec.log, AskCode: func(*AskCodeEvent) {}})
	start := time.Now()
	if got := b.AskCode(KindCode); got != "" {
		t.Fatalf("AskCode = %q, want empty string on timeout", got)
	}
	if d := time.Since(start); d > 5*time.Second {
		t.Fatalf("timeout took %v — the instrument window leaked", d)
	}
	// The reference's own line, byte-identical
	// (processing_worker.py:2090).
	l, ok := rec.find("Auth code input timed out")
	if !ok {
		t.Fatalf("the timeout warning never landed: %v", rec.lines)
	}
	if l.Level != "warning" {
		t.Fatalf("timeout line level = %q, want warning", l.Level)
	}
	if l.Msg != "⏰ Auth code input timed out (5 minutes)" {
		t.Fatalf("timeout line = %q, want the reference's bytes", l.Msg)
	}
}

func TestAskCodeHeadlessFailFast(t *testing.T) {
	b := tinyBus()
	b.Headless = true
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log})
	start := time.Now()
	if got := b.AskCode(KindCode); got != "" {
		t.Fatalf("headless AskCode = %q, want empty string", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("headless gate took %v — must fail fast", d)
	}
	if l, ok := rec.find("no GUI available; the worker fails fast"); !ok {
		t.Fatalf("the headless line never landed: %v", rec.lines)
	} else if l.Level != "warning" {
		t.Fatalf("headless line level = %q, want warning", l.Level)
	}
}

func TestAskCodeNoReaderFailFast(t *testing.T) {
	b := tinyBus()
	// A log-only subscriber: nobody can answer an ask.
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log})
	start := time.Now()
	if got := b.AskCode(KindCode); got != "" {
		t.Fatalf("no-reader AskCode = %q, want empty string", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("no-reader gate took %v — the hang-bomb guard failed", d)
	}
	if _, ok := rec.find("no ask subscriber is listening"); !ok {
		t.Fatalf("the no-reader line never landed: %v", rec.lines)
	}
}

func TestAskScanConfirmRoundTripWithSelection(t *testing.T) {
	b := tinyBus()
	b.Subscribe(&Handlers{AskScanConfirm: func(e *AskScanConfirmEvent) {
		if e.Plan == nil {
			t.Error("the plan never rode the event")
		}
		e.Answer("confirmed", map[string][]string{
			"deletions": {"d1", "d2"},
		})
	}})
	ans := b.AskScanConfirm(map[string]any{"kind": "repos"})
	if ans.Verdict != ScanConfirmed {
		t.Fatalf("verdict = %q, want confirmed", ans.Verdict)
	}
	if len(ans.Selection["deletions"]) != 2 {
		t.Fatalf("selection = %v, want the two picked rows", ans.Selection)
	}
}

func TestAskScanConfirmDeclineStripsSelection(t *testing.T) {
	b := tinyBus()
	b.Subscribe(&Handlers{AskScanConfirm: func(e *AskScanConfirmEvent) {
		// The shape law: even a decline sent WITH a selection carries
		// nothing (the gate strips it — provide_scan_verdict's own
		// `else None`).
		e.Answer("declined", map[string][]string{"moves": {"m1"}})
	}})
	ans := b.AskScanConfirm(nil)
	if ans.Verdict != ScanDeclined {
		t.Fatalf("verdict = %q, want declined", ans.Verdict)
	}
	if ans.Selection != nil {
		t.Fatalf("declined selection = %v, want nil", ans.Selection)
	}
}

func TestNormalizeScanVerdictJunk(t *testing.T) {
	cases := map[string]string{
		"confirmed":                  ScanConfirmed,
		"  CONFIRMED\n":              ScanConfirmed,
		"Declined":                   ScanDeclined,
		"yes":                        ScanTimeout,
		"":                           ScanTimeout,
		"timeout":                    ScanTimeout,
		"confirmed-with-extra-words": ScanTimeout,
	}
	for in, want := range cases {
		if got := NormalizeScanVerdict(in); got != want {
			t.Errorf("NormalizeScanVerdict(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestAskScanConfirmTimeoutSafeDefer(t *testing.T) {
	b := tinyBus()
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log, AskScanConfirm: func(*AskScanConfirmEvent) {}})
	ans := b.AskScanConfirm(nil)
	if ans.Verdict != ScanTimeout {
		t.Fatalf("verdict = %q, want timeout", ans.Verdict)
	}
	if ans.Selection != nil {
		t.Fatalf("timeout selection = %v, want nil", ans.Selection)
	}
	// The reference's own sentence, byte-identical
	// (processing_worker.py:2068-2071).
	l, ok := rec.find("answering window closed")
	if !ok {
		t.Fatalf("the safe-defer line never landed: %v", rec.lines)
	}
	if l.Msg != "⏰ The scan plan's answering window closed — nothing "+
		"moved, nothing deleted (the safe defer; I'll propose again on "+
		"the next scan)" {
		t.Fatalf("safe-defer line = %q, want the reference's bytes", l.Msg)
	}
	if l.Level != "warning" {
		t.Fatalf("safe-defer level = %q, want warning", l.Level)
	}
}

func TestAskScanConfirmHeadlessAndNoReader(t *testing.T) {
	// Headless: the safe defer, immediately.
	b := tinyBus()
	b.Headless = true
	start := time.Now()
	if ans := b.AskScanConfirm(nil); ans.Verdict != ScanTimeout || ans.Selection != nil {
		t.Fatalf("headless scan answer = %+v, want the empty defer", ans)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("headless scan gate took %v", d)
	}
	// No reader: the same defer (the hang-bomb guard).
	b2 := tinyBus()
	start = time.Now()
	if ans := b2.AskScanConfirm(nil); ans.Verdict != ScanTimeout {
		t.Fatalf("no-reader scan answer = %+v, want the defer", ans)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("no-reader scan gate took %v", d)
	}
}

func TestLLMFailedRoundTrip(t *testing.T) {
	b := tinyBus()
	b.Subscribe(&Handlers{LLMFailed: func(e *LLMFailedEvent) {
		if e.Repo != "octo-owner/curated-repo" {
			t.Errorf("repo = %q", e.Repo)
		}
		e.Answer("llama3.1")
	}})
	if got := b.LLMFailed("octo-owner/curated-repo"); got != "llama3.1" {
		t.Fatalf("LLMFailed = %q, want llama3.1 (the model-retry door)", got)
	}
	// Empty answers read as skip (the `or "skip"` law).
	b2 := tinyBus()
	b2.Subscribe(&Handlers{LLMFailed: func(e *LLMFailedEvent) { e.Answer("") }})
	if got := b2.LLMFailed("r"); got != "skip" {
		t.Fatalf("empty LLM answer = %q, want skip", got)
	}
}

func TestLLMFailedTimeoutSilentSkip(t *testing.T) {
	b := tinyBus()
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log, LLMFailed: func(*LLMFailedEvent) {}})
	if got := b.LLMFailed("r"); got != "skip" {
		t.Fatalf("timeout decision = %q, want skip", got)
	}
	// The reference has no timeout line (llm.py waits silently) — the
	// bus must not invent one.
	if len(rec.lines) != 0 {
		t.Fatalf("the timeout invented log lines: %v", rec.lines)
	}
}

func TestLLMFailedHeadlessLine(t *testing.T) {
	b := tinyBus()
	b.Headless = true
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log})
	start := time.Now()
	if got := b.LLMFailed("octo/ repo"); got != "skip" {
		t.Fatalf("headless decision = %q, want skip", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("headless LLM gate took %v", d)
	}
	l, ok := rec.find("Headless mode: LLM failed for 'octo/ repo'")
	if !ok {
		t.Fatalf("the headless LLM line never landed: %v", rec.lines)
	}
	if l.Msg != "⏭️ Headless mode: LLM failed for 'octo/ repo' — using "+
		"fallback values for the note (same as the GUI 'Skip' button) "+
		"and continuing." {
		t.Fatalf("headless LLM line = %q, want the reference's bytes", l.Msg)
	}
}

func TestLLMFailedNoReaderFailsFast(t *testing.T) {
	b := tinyBus()
	rec := &recLog{}
	b.Subscribe(&Handlers{LogLine: rec.log})
	start := time.Now()
	if got := b.LLMFailed("r"); got != "skip" {
		t.Fatalf("no-reader decision = %q, want skip", got)
	}
	if d := time.Since(start); d > time.Second {
		t.Fatalf("no-reader LLM gate took %v — the hang-bomb guard failed", d)
	}
	if _, ok := rec.find("no ask subscriber is listening"); !ok {
		t.Fatalf("the no-reader line never landed: %v", rec.lines)
	}
}

func TestBusConcurrentEmitAndSubscribe(t *testing.T) {
	b := tinyBus()
	var count int64
	var mu sync.Mutex
	b.Subscribe(&Handlers{
		LogLine: func(LogLine) { mu.Lock(); count++; mu.Unlock() },
		Status:  func(Status) { mu.Lock(); count++; mu.Unlock() },
	})
	var wg sync.WaitGroup
	for w := 0; w < 4; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for i := 0; i < 50; i++ {
				b.EmitLogLine("line", "info")
				b.EmitStatus("tick")
				// A concurrent subscriber churning the list must not
				// deadlock emit (snapshot copies) or race.
				cancel := b.Subscribe(&Handlers{})
				cancel()
			}
		}()
	}
	wg.Wait()
	if count != 400 {
		t.Fatalf("delivered %d, want 400", count)
	}
}

// TestCrossedPayloadFieldTags pins THE FIELD-TAG LAW on every payload
// the shell crosses to the frontend as a wails event (T-77's retroactive
// fix: the tagless T-71 shapes marshaled as "OK"/"Pos"/"Text" while
// app.js read ok/pos/text — the progress row never filled and every
// finish rendered ❌). The tags mirror the keys the frontend reads.
func TestCrossedPayloadFieldTags(t *testing.T) {
	cases := []struct {
		name string
		v    any
		want string
	}{
		{"LogLine", LogLine{Msg: "m", Level: "info"}, `{"msg":"m","level":"info"}`},
		{"Progress", Progress{Pos: 2, Total: 5}, `{"pos":2,"total":5}`},
		{"Status", Status{Text: "Processing…"}, `{"text":"Processing…"}`},
		{"BatchFinished", BatchFinished{OK: true, Summary: "done", NeedsRetry: true},
			`{"ok":true,"summary":"done","needs_retry":true,"natural":false,"stopped":false,"website_results":null}`},
		{"JobProgress", JobProgress{Label: "walk", Value: 1, Max: 10},
			`{"label":"walk","value":1,"max":10}`},
		{"DiskFull", DiskFull{Path: "/v"}, `{"path":"/v"}`},
		{"ModelChanged", ModelChanged{Provider: "ollama", Model: "llama3"},
			`{"provider":"ollama","model":"llama3"}`},
	}
	for _, tc := range cases {
		raw, err := json.Marshal(tc.v)
		if err != nil {
			t.Fatalf("%s marshal: %v", tc.name, err)
		}
		if string(raw) != tc.want {
			t.Fatalf("%s = %s, want %s", tc.name, raw, tc.want)
		}
	}
	// The sentinel snapshot's probe shapes ride their own tags (the
	// connectioncheck dict keys — THE FIELD-TAG LAW's source of truth).
	raw, err := json.Marshal(Sentinel{
		Net:     connectioncheck.ProbeInternetResult{Online: true},
		AI:      connectioncheck.AIStatusResult{State: "on", Kind: "cloud"},
		Proxy:   connectioncheck.ProxyResult{State: "disabled"},
		Verdict: "first-online",
	})
	if err != nil {
		t.Fatalf("sentinel marshal: %v", err)
	}
	for _, key := range []string{
		`"net":{"online":true`, `"ai":{"state":"on","kind":"cloud"`,
		`"proxy":{"state":"disabled"`, `"verdict":"first-online"`,
	} {
		if !strings.Contains(string(raw), key) {
			t.Fatalf("sentinel payload missing %s: %s", key, raw)
		}
	}
}
