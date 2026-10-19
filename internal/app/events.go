// events.go — T-71 slice (a): the typed event bus + the ask-gates —
// doc 04 §2's composition seam, the production evolution of the
// guibridge/ pattern (T-70) and the transport replacement for the
// Python reference's 8 queued Qt signals
// (gui/processing_worker.py:101-112) and the TestWorker interactive
// doors' threading.Event round-trips (processing_worker.py:1959-2113,
// worker/llm.py:190-260).
//
// THE CONTRACTS (each mirrors a reference law):
//
//   - Fire-and-forget events (LogLine, Progress, Status, BatchFinished,
//     JobProgress, DiskFull, ModelChanged) replace the queued signals:
//     emitting NEVER blocks the command goroutine beyond the
//     subscribers' own marshaling — a subscriber must not block (the
//     GUI marshals to its main thread, exactly like a queued Qt
//     connection; the CLI prints). Fan-out is in subscription order,
//     panic-isolated per subscriber (a crashing subscriber is logged
//     past, never allowed to kill a command — the never-raise law).
//
//   - The ask-gates (AskCode, AskScanConfirm, LLMFailed) are channels
//     with the reference's IDENTICAL timeouts: 300 s for the login
//     code (request_code, processing_worker.py:2085), 300 s for the
//     scan plan (request_scan_confirm, :2078), 600 s for the LLM
//     failure decision (_wait_for_llm_decision, worker/llm.py:256).
//     The first answer wins; late answers are dropped without
//     blocking (the Answer helpers are non-blocking sends).
//
//   - THE NO-READER FAIL-FAST (the v30 hang-bomb law's Go twin): a
//     gate with no ask subscriber can never be answered — waiting the
//     full window would be the 10-minute-per-repo hang the reference's
//     headless fix killed, so the bus fails fast to the same defaults
//     the reference's headless mode returns. A channel with no reader
//     is a defer, never a stall.
//
//   - THE HEADLESS LAW (doc 04 §2 + gui/headless.py:219-227): in
//     headless mode every gate takes the fail-fast default immediately
//     ("" for the code, the scan's safe defer, "skip" for the LLM
//     decision) with the honest line on the bus. Composition-level
//     narration, not the reference's clipped `eadless]` stdout prints
//     (the reference's belt-and-suspenders receivers print those;
//     the sentences are mirrored with the unclipped [headless]
//     prefix — a recorded divergence, not a golden-tested string).
package app

import (
	"fmt"
	"strings"
	"sync"
	"time"
)

// ---------------------------------------------------------------------------
// The event types (the 8 queued signals' payloads + JobProgress).
// ---------------------------------------------------------------------------

// LogLine is log_message(str, str): one desk-log row.
type LogLine struct {
	Msg   string
	Level string
}

// Progress is progress_updated(int, int): pos of total items done.
type Progress struct {
	Pos   int
	Total int
}

// Status is status_updated(str): the pipeline row's text.
type Status struct {
	Text string
}

// BatchFinished is finished_signal(bool, str): the batch's outcome and
// its bottom-line summary (the lock-leak fix's "exactly once" law is
// the command's own; the bus just carries it).
type BatchFinished struct {
	OK      bool
	Summary string
}

// JobProgress is job_progress(str, int, int) — THE HONEST PIPELINE ROW
// (v0.70.0): any job's LIVE progress (label, value, maximum).
type JobProgress struct {
	Label string
	Value int
	Max   int
}

// DiskFull is disk_full_signal(str): the path that failed — the GUI
// opens the paused-spin modal and answers via the runner's resume
// door.
type DiskFull struct {
	Path string
}

// ModelChanged is model_changed(str, str) — the v30 persistence law:
// the provider + model the LLM-failure dialog (or the auto-switch)
// picked; the shell syncs Settings and saves config.
type ModelChanged struct {
	Provider string
	Model    string
}

// ---------------------------------------------------------------------------
// The ask-gate events (the Event round-trips' transport).
// ---------------------------------------------------------------------------

// The code door's kinds (code_requested's vocabulary).
const (
	KindCode     = "CODE"
	KindPassword = "PASSWORD"
)

// AskCodeEvent is code_requested(str): the worker needs a login code
// (or password). The subscriber shows the prompt and answers via
// Answer; Reply carries the code ("" = cancelled).
type AskCodeEvent struct {
	Kind string
	// Reply is buffered(1): the FIRST answer wins; Answer never blocks.
	Reply chan string
}

// Answer delivers the code/password. Non-blocking: a second answer
// (or an answer after the gate returned) is dropped silently.
func (e *AskCodeEvent) Answer(code string) {
	select {
	case e.Reply <- code:
	default:
	}
}

// The scan verdict vocabulary (provide_scan_verdict's law):
// 'confirmed' / 'declined' — anything else reads as 'timeout' (the
// safe defer: nothing moved, nothing deleted).
const (
	ScanConfirmed = "confirmed"
	ScanDeclined  = "declined"
	ScanTimeout   = "timeout"
)

// ScanAnswer is the scan door's reply: the owner's verdict plus — only
// on a CONFIRMED verdict — the picking door's selection (v0.73.0: the
// checked rows' ids; nil when the answer carries no selection, an
// all-or-nothing door, a decline, or a timeout).
type ScanAnswer struct {
	Verdict   string
	Selection map[string][]string
}

// AskScanConfirmEvent is scan_confirm_requested(dict): the plan
// awaiting the owner. The subscriber opens the plan dialog (or the
// Telegram door) and answers via Answer.
type AskScanConfirmEvent struct {
	// Plan is the vaultscan plan (a *vaultscan.Plan in production;
	// any — the bus is transport, not validator).
	Plan any
	// Reply is buffered(1): the FIRST answer wins.
	Reply chan ScanAnswer
}

// Answer delivers the verdict (and, on confirm, the picked rows).
// Non-blocking: a second answer is dropped silently.
func (e *AskScanConfirmEvent) Answer(verdict string, selection map[string][]string) {
	v := NormalizeScanVerdict(verdict)
	sel := selection
	if v != ScanConfirmed {
		sel = nil // the shape law: only a confirmed verdict carries one
	}
	select {
	case e.Reply <- ScanAnswer{Verdict: v, Selection: sel}:
	default:
	}
}

// NormalizeScanVerdict mirrors the verdict law
// (processing_worker.py:2089): strip + lower, and only
// 'confirmed'/'declined' pass; anything else is 'timeout'.
func NormalizeScanVerdict(verdict string) string {
	v := strings.ToLower(strings.TrimSpace(verdict))
	if v == ScanConfirmed || v == ScanDeclined {
		return v
	}
	return ScanTimeout
}

// LLMFailedEvent is llm_failed_signal(str) + resolve_llm_failure: the
// repo whose LLM analysis failed. The subscriber opens the failure
// modal and answers via Answer with 'skip', 'retry', 'stop', or a
// model name to retry with.
type LLMFailedEvent struct {
	Repo string
	// Reply is buffered(1): the FIRST answer wins.
	Reply chan string
}

// Answer delivers the decision. Non-blocking: a second answer is
// dropped silently.
func (e *LLMFailedEvent) Answer(decision string) {
	select {
	case e.Reply <- decision:
	default:
	}
}

// The reference's identical timeouts (doc 04 §2 — never bump without
// an owner decision; they are behavior).
const (
	// CodeAskTimeout is request_code's 300 s window
	// (processing_worker.py:2085).
	CodeAskTimeout = 300 * time.Second
	// ScanAskTimeout is request_scan_confirm's 300 s window
	// (processing_worker.py:2078).
	ScanAskTimeout = 300 * time.Second
	// LLMAskTimeout is _wait_for_llm_decision's 600 s window
	// (worker/llm.py:256).
	LLMAskTimeout = 600 * time.Second
)

// ---------------------------------------------------------------------------
// The bus.
// ---------------------------------------------------------------------------

// Handlers is one subscriber's typed handlers. Every field is
// optional; nil fields skip that event. Handlers must not block (the
// queued-connection law: marshal to your own thread and return).
type Handlers struct {
	LogLine        func(LogLine)
	Progress       func(Progress)
	Status         func(Status)
	BatchFinished  func(BatchFinished)
	JobProgress    func(JobProgress)
	DiskFull       func(DiskFull)
	ModelChanged   func(ModelChanged)
	AskCode        func(*AskCodeEvent)
	AskScanConfirm func(*AskScanConfirmEvent)
	LLMFailed      func(*LLMFailedEvent)
}

// Bus is the fan-out event bus + the ask-gate transport.
type Bus struct {
	mu   sync.Mutex
	subs []*Handlers

	// Headless mirrors the worker's _headless flag: no GUI will ever
	// answer a gate — every ask fails fast to its default (the
	// interactive-auth pattern's headless law).
	Headless bool

	// The gate windows. New() pins the reference's values; tests
	// construct with tiny windows (the fake-clock law's sibling —
	// the production constants are the law, the fields are the
	// instrument).
	CodeTimeout time.Duration
	ScanTimeout time.Duration
	LLMTimeout  time.Duration
}

// New builds the bus with the reference's identical timeouts.
func New() *Bus {
	return &Bus{
		CodeTimeout: CodeAskTimeout,
		ScanTimeout: ScanAskTimeout,
		LLMTimeout:  LLMAskTimeout,
	}
}

// Subscribe registers h; the returned cancel removes it. Cancel is
// idempotent.
func (b *Bus) Subscribe(h *Handlers) (cancel func()) {
	if h == nil {
		return func() {}
	}
	b.mu.Lock()
	b.subs = append(b.subs, h)
	b.mu.Unlock()
	var once sync.Once
	return func() {
		once.Do(func() {
			b.mu.Lock()
			for i, s := range b.subs {
				if s == h {
					b.subs = append(b.subs[:i], b.subs[i+1:]...)
					break
				}
			}
			b.mu.Unlock()
		})
	}
}

// snapshot copies the subscriber list (emit must not hold the lock
// while calling out — a subscriber that (un)subscribes from a handler
// would deadlock otherwise).
func (b *Bus) snapshot() []*Handlers {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]*Handlers, len(b.subs))
	copy(out, b.subs)
	return out
}

// deliver calls fn recovered: a panicking subscriber is logged past
// on the bus itself (never raised across the command boundary).
func deliver(b *Bus, fn func()) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("[app-bus] subscriber panicked (isolated): %v\n", r)
		}
	}()
	fn()
}

// emit fans one event out in subscription order.
func emit[T any](b *Bus, pick func(*Handlers) func(T), ev T) {
	for _, h := range b.snapshot() {
		if fn := pick(h); fn != nil {
			deliver(b, func() { fn(ev) })
		}
	}
}

// EmitLogLine delivers log_message(msg, level).
func (b *Bus) EmitLogLine(msg, level string) {
	emit(b, func(h *Handlers) func(LogLine) { return h.LogLine }, LogLine{Msg: msg, Level: level})
}

// EmitProgress delivers progress_updated(pos, total).
func (b *Bus) EmitProgress(pos, total int) {
	emit(b, func(h *Handlers) func(Progress) { return h.Progress }, Progress{Pos: pos, Total: total})
}

// EmitStatus delivers status_updated(text).
func (b *Bus) EmitStatus(text string) {
	emit(b, func(h *Handlers) func(Status) { return h.Status }, Status{Text: text})
}

// EmitBatchFinished delivers finished_signal(ok, summary).
func (b *Bus) EmitBatchFinished(ok bool, summary string) {
	emit(b, func(h *Handlers) func(BatchFinished) { return h.BatchFinished }, BatchFinished{OK: ok, Summary: summary})
}

// EmitJobProgress delivers job_progress(label, value, maximum) —
// junk-tolerant by law (emit_job_progress: negative numbers clamp to
// 0; "a progress line must never kill the job that tried to speak
// it").
func (b *Bus) EmitJobProgress(label string, value, max int) {
	if value < 0 {
		value = 0
	}
	if max < 0 {
		max = 0
	}
	emit(b, func(h *Handlers) func(JobProgress) { return h.JobProgress }, JobProgress{Label: label, Value: value, Max: max})
}

// EmitDiskFull delivers disk_full_signal(path).
func (b *Bus) EmitDiskFull(path string) {
	emit(b, func(h *Handlers) func(DiskFull) { return h.DiskFull }, DiskFull{Path: path})
}

// EmitModelChanged delivers model_changed(provider, model).
func (b *Bus) EmitModelChanged(provider, model string) {
	emit(b, func(h *Handlers) func(ModelChanged) { return h.ModelChanged }, ModelChanged{Provider: provider, Model: model})
}

// askReady reports whether at least one live subscriber listens for
// the given ask-gate (nil handler lists mean nobody can ever answer —
// the no-reader fail-fast).
func (b *Bus) askReady(pick func(*Handlers) bool) bool {
	for _, h := range b.snapshot() {
		if pick(h) {
			return true
		}
	}
	return false
}

// hasAskCode etc. — the per-gate probes.
func hasAskCode(h *Handlers) bool   { return h.AskCode != nil }
func hasAskScan(h *Handlers) bool   { return h.AskScanConfirm != nil }
func hasLLMFailed(h *Handlers) bool { return h.LLMFailed != nil }

// AskCode mirrors request_code: the ask rides the bus, the caller
// blocks until a subscriber answers, the window closes, or the bus is
// headless / has no ask reader. Returns the code, or "" on cancel,
// timeout, or fail-fast (the reference's own empty-string contract).
func (b *Bus) AskCode(kind string) string {
	if kind == "" {
		kind = KindCode
	}
	if b.Headless {
		b.EmitLogLine("[headless] Telegram "+kind+
			" requested — no GUI available; the worker fails fast.", "warning")
		return ""
	}
	if !b.askReady(hasAskCode) {
		b.EmitLogLine("[app-bus] Telegram "+kind+
			" requested — no ask subscriber is listening; failing fast (the no-reader law).", "warning")
		return ""
	}
	reply := make(chan string, 1)
	ev := &AskCodeEvent{Kind: kind, Reply: reply}
	for _, h := range b.snapshot() {
		if h.AskCode != nil {
			hh := h
			deliver(b, func() { hh.AskCode(ev) })
		}
	}
	select {
	case code := <-reply:
		return code
	case <-time.After(b.CodeTimeout):
		// The reference's own line, byte-identical
		// (processing_worker.py:2090).
		b.EmitLogLine("⏰ Auth code input timed out (5 minutes)", "warning")
		return ""
	}
}

// AskScanConfirm mirrors request_scan_confirm: the plan rides the bus,
// the caller blocks for the owner's verdict (window: ScanTimeout).
// Returns the normalized ScanAnswer — confirmed (with the picked
// rows), declined, or timeout (the safe defer: nothing moved, nothing
// deleted).
func (b *Bus) AskScanConfirm(plan any) ScanAnswer {
	if b.Headless {
		b.scanDeferLine()
		return ScanAnswer{Verdict: ScanTimeout}
	}
	if !b.askReady(hasAskScan) {
		b.EmitLogLine("[app-bus] the scan plan asked — no ask subscriber is listening; deferring safely (the no-reader law).", "warning")
		return ScanAnswer{Verdict: ScanTimeout}
	}
	reply := make(chan ScanAnswer, 1)
	ev := &AskScanConfirmEvent{Plan: plan, Reply: reply}
	for _, h := range b.snapshot() {
		if h.AskScanConfirm != nil {
			hh := h
			deliver(b, func() { hh.AskScanConfirm(ev) })
		}
	}
	select {
	case ans := <-reply:
		ans.Verdict = NormalizeScanVerdict(ans.Verdict)
		if ans.Verdict != ScanConfirmed {
			ans.Selection = nil
		}
		return ans
	case <-time.After(b.ScanTimeout):
		b.scanDeferLine()
		return ScanAnswer{Verdict: ScanTimeout}
	}
}

// scanDeferLine is the reference's own answering-window-closed
// sentence, byte-identical (processing_worker.py:2068-2071).
func (b *Bus) scanDeferLine() {
	b.EmitLogLine("⏰ The scan plan's answering window closed — nothing "+
		"moved, nothing deleted (the safe defer; I'll propose again on "+
		"the next scan)", "warning")
}

// LLMFailed mirrors _wait_for_llm_decision's GUI half: the failure ask
// rides the bus, the caller blocks for the decision (window:
// LLMTimeout). Returns the decision — 'skip', 'retry', 'stop', or a
// model name — with "skip" as the timeout default (the reference's
// `return self._llm_retry_response or "skip"`).
func (b *Bus) LLMFailed(repo string) string {
	if b.Headless {
		// The reference's own headless line, byte-identical
		// (worker/llm.py:442-445).
		b.EmitLogLine("⏭️ Headless mode: LLM failed for '"+repo+
			"' — using fallback values for the note (same as the GUI "+
			"'Skip' button) and continuing.", "warning")
		return "skip"
	}
	if !b.askReady(hasLLMFailed) {
		// No failure modal can ever open: the batch must not park 600 s
		// per repo — the v30 hang-bomb law.
		b.EmitLogLine("[app-bus] LLM failed for '"+repo+
			"' — no ask subscriber is listening; skipping with fallback values (the no-reader law).", "warning")
		return "skip"
	}
	reply := make(chan string, 1)
	ev := &LLMFailedEvent{Repo: repo, Reply: reply}
	for _, h := range b.snapshot() {
		if h.LLMFailed != nil {
			hh := h
			deliver(b, func() { hh.LLMFailed(ev) })
		}
	}
	select {
	case decision := <-reply:
		if strings.TrimSpace(decision) == "" {
			return "skip"
		}
		return decision
	case <-time.After(b.LLMTimeout):
		// The reference has no timeout line here (llm.py:256-258 waits
		// silently and falls back to "skip") — mirrored exactly.
		return "skip"
	}
}
