package app

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime/debug"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/mirror"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// linkdoors.go — T-79b: the More menu's linking-tools doors (the
// pinned reference at 98f4fc7): 🪝 Recall hooks (dry-run) +
// 🔗 Build link suggestions (suggest mode) over linking_tools.py's
// TestWorker runners, ✅ Validate Vault over connection_tests.py:219,
// 🔍 Verify Vault over dashboard.py's verify_vault, and 📁
// Recategorize Notes over dashboard.py's recategorize_notes — the
// reference's SAFE defaults only, over the ported engines
// (internal/mirror's recall/linking/embeddings, internal/batch's
// LinkTracker).
//
// THE GUARD LAW (the reference's own, widened per the task row's
// wording — the row names all five doors "the linking tools"): no
// tool while a batch runs. The reference guards only the two
// _run_phase6_tool runners (linking_tools.py:62-66); the port carries
// the same guard on all five doors — validate/verify/recategorize
// read/write the vaults a running batch may be mid-writing, and the
// owner's Path-B stamp ("the slow, more reliable and more rigorous
// method") prefers the safer family. Recorded divergence.
//
// THE TESTWORKER LAW (linking_tools.py:59-105): the two tool doors
// capture the tool's stdout+stderr, stream NOTHING while the tool
// runs, and land the LAST 40 LINES on the desk at completion — then
// the finish line ("✅ {label} finished." / "❌ {label} finished with
// exit code {rc} (see the lines above — the command-line tools print
// the exact next step)."). A tool crash renders the reference's
// "💥 …" line into the captured buffer and exits 2.
//
// THE SAFE-DEFAULT LAWS: recall hooks run DRY (apply=false — the diff
// lands on the desk + APP_DIR/reports/recall/recall-dryrun.md, the
// only write); link suggestions run in SUGGEST mode (the one write is
// the Suggestions note under the manual vault's Library/ — the
// collect/apply step stays on the command line, the SPEC flow).

// linkingBusyLine is the guard's own warning (linking_tools.py:63-65,
// verbatim).
const linkingBusyLine = "⏳ A batch is running — wait for it to finish " +
	"before running the linking tools."

// linkLabels are the two tool-runner labels (linking_tools.py:114/123).
const (
	labelRecallHooks = "Recall hooks (dry-run)"
	labelLinkSuggest = "Link suggestions"
)

// ---------------------------------------------------------------------------
// The TestWorker twin: the captured-output runner.
// ---------------------------------------------------------------------------

// linkToolOut is the StringIO buffer both redirects write into (the
// stdout + stderr capture — the reference redirects both into one
// buf).
type linkToolOut struct{ b strings.Builder }

// printf is print(…)'s twin (each call one line).
func (o *linkToolOut) printf(format string, args ...any) {
	fmt.Fprintf(&o.b, format+"\n", args...)
}

// out is the progress-callback shape the engines take (one plain
// line per call).
func (o *linkToolOut) out(msg string) {
	o.printf("%s", msg)
}

// stderrLine is print(…, file=sys.stderr)'s twin — the same buffer
// (redirect_stderr captures it too).
func (o *linkToolOut) stderrLine(msg string) {
	fmt.Fprintf(&o.b, "%s\n", msg)
}

// crashLine is the reference's generic-except shape: the 💥 line
// written INTO the buffer (so it rides the tail), rc 2.
func (o *linkToolOut) crashLine(v any) {
	fmt.Fprintf(&o.b, "\n💥 %v\n", v)
}

func (o *linkToolOut) String() string { return o.b.String() }

// tailLines is (output or ”).splitlines()[-40:] — the last-40 law.
func tailLines(s string, n int) []string {
	if s == "" {
		return nil
	}
	s = strings.TrimSuffix(s, "\n")
	lines := strings.Split(s, "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return lines
}

// runLinkingTool is _run_phase6_tool's twin: the busy guard, the
// opening line, the captured run (crash-safe — the reference's
// TestWorker catches BaseException and always signals), the tail-40
// streaming, and the finish verdict. Log-only (no box).
func (s *Settings) runLinkingTool(label string, tool func(o *linkToolOut) int) {
	if s.linkBusy() {
		return
	}
	s.SpeakOnBus(fmt.Sprintf("⏳ %s running… (the GUI stays usable)", label), "info")
	o := &linkToolOut{}
	rc := 0
	func() {
		defer func() {
			if r := recover(); r != nil {
				o.crashLine(r)
				rc = 2
			}
		}()
		rc = tool(o)
	}()
	for _, line := range tailLines(o.String(), 40) {
		s.SpeakOnBus(line, "info")
	}
	if rc == 0 {
		s.SpeakOnBus(fmt.Sprintf("✅ %s finished.", label), "success")
	} else {
		s.SpeakOnBus(fmt.Sprintf(
			"❌ %s finished with exit code %d (see the lines above — "+
				"the command-line tools print the exact next step).",
			label, rc), "error")
	}
}

// linkBusy is the guard law's shared half: the warning line + the
// refusal.
func (s *Settings) linkBusy() bool {
	if s.lock != nil && s.lock.Busy() {
		s.SpeakOnBus(linkingBusyLine, "warning")
		return true
	}
	return false
}

// ---------------------------------------------------------------------------
// ✅ Validate Vault (connection_tests.py:219 test_vault — fast,
// inline, log-only).
// ---------------------------------------------------------------------------

// ValidateVault is test_vault's flow verbatim: the path from the LIVE
// config's vault_path (the reference reads the vault combo — the
// port's save-on-change combo law makes the config the same value),
// the .obsidian check, the five verdict lines. The busy guard rides
// the family law (the recorded widening).
func (s *Settings) ValidateVault() {
	if s.linkBusy() {
		return
	}
	s.SpeakOnBus("📁 Validating vault path...", "info")
	path := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if path == "" {
		s.SpeakOnBus("❌ No vault selected.", "error")
		return
	}
	if fi, err := os.Stat(path); err != nil || !fi.IsDir() {
		s.SpeakOnBus(fmt.Sprintf("❌ Vault path does not exist: %s", path), "error")
		return
	}
	if fi, err := os.Stat(filepath.Join(path, ".obsidian")); err == nil && fi.IsDir() {
		s.SpeakOnBus(fmt.Sprintf(
			"✅ Vault is valid (contains .obsidian folder): %s", path), "success")
		return
	}
	s.SpeakOnBus("⚠️ Path exists but does not appear to be an Obsidian vault "+
		"(no .obsidian folder).", "warning")
}

// ---------------------------------------------------------------------------
// 🔍 Verify Vault (dashboard.py:302-498 verify_vault).
// ---------------------------------------------------------------------------

// VerifyVault is verify_vault's flow: the manifest read, the tracker's
// full-accounting Verify over the vault ground truth (the websites
// vault's _inbox in the extra dirs), the summary line, the report (the
// reference renders it into the dashboard text area — the port's More
// door renders the themed box, the recorded divergence), and the
// failed-links enqueue legs (github links only, the settled-repo
// filter, the retry queue). The crash guard never takes the app down.
func (s *Settings) VerifyVault() (title, body string) {
	if s.linkBusy() {
		return "", ""
	}
	defer func() {
		if r := recover(); r != nil {
			s.SpeakOnBus(fmt.Sprintf("❌ Verify Vault crashed: %v", r), "error")
			title, body = "Verify Vault crashed",
				fmt.Sprintf("Verify Vault crashed: %v\n\n%s", r, debug.Stack())
		}
	}()
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		return "Verify Vault", "No vault selected."
	}
	tracker := batch.NewLinkTracker(vault)
	prev := tracker.LoadPreviousManifest()
	if prev == nil {
		s.SpeakOnBus("ℹ️ No manifest to verify — process a batch first.", "info")
		return "Verify Vault",
			"No previous manifest found in this vault.\n" +
				"Process a batch first — the manifest is created at the start of every batch."
	}
	tracker.Manifest = prev
	var extraInbox []string
	if wv := strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path")); wv != "" {
		extraInbox = append(extraInbox, filepath.Join(wv, "_inbox"))
	}
	report := tracker.Verify(nil, extraInbox)
	s.SpeakOnBus(fmt.Sprintf("🔍 Verify: %d processed, %d failed",
		report.GithubProcessed, report.GithubFailed), "info")

	// The report (the dashboard text's own rendering, verbatim).
	lines := []string{
		strings.Repeat("=", 60),
		"🔍 VAULT VERIFICATION REPORT",
		strings.Repeat("=", 60),
		"Batch:    " + pyOr(prev.BatchID, "unknown"),
		"Source:   " + pyOr(prev.Source, "unknown"),
		"Created:  " + pyOr(prev.CreatedAt, "unknown"),
		fmt.Sprintf("Total links: %d", report.Total),
		"",
		fmt.Sprintf("✅ GitHub processed:    %d", report.GithubProcessed),
		fmt.Sprintf("⏭️ GitHub skipped:      %d  (dedup — already in vault)",
			report.GithubSkipped),
	}
	if report.GithubPending > 0 {
		lines = append(lines, fmt.Sprintf(
			"⏳ GitHub pending:      %d  (unfinished — will retry)",
			report.GithubPending))
	}
	lines = append(lines, fmt.Sprintf("❌ GitHub failed:       %d", report.GithubFailed))
	if report.NonGithubRecorded > 0 {
		lines = append(lines, fmt.Sprintf(
			"✅ Non-GitHub recorded: %d  (inbox tables)", report.NonGithubRecorded))
	}
	if report.WebsitesProcessed > 0 || report.WebsitesReview > 0 || report.WebsitesSkipped > 0 {
		lines = append(lines,
			fmt.Sprintf("🌐 Websites notes:      %d", report.WebsitesProcessed),
			fmt.Sprintf("🗂️ Websites in _review: %d  (retry scheduled)", report.WebsitesReview),
			fmt.Sprintf("⏭️ Websites skipped:    %d  (dedup)", report.WebsitesSkipped))
	}
	if report.BlockedRecorded > 0 {
		lines = append(lines, fmt.Sprintf(
			"🚫 Blocked/self domains: %d  (omitted by design — never collected)",
			report.BlockedRecorded))
	}
	if report.NonGithubPending > 0 {
		lines = append(lines, fmt.Sprintf(
			"⏳ Non-GitHub pending:  %d  (websites pipeline off / no vault)",
			report.NonGithubPending))
	}
	if report.NonGithubFailed > 0 {
		lines = append(lines, fmt.Sprintf("❌ Non-GitHub failed:   %d",
			report.NonGithubFailed))
	}
	if report.AccountingOK {
		lines = append(lines, fmt.Sprintf(
			"🧮 Accounting:          %d/%d accounted for ✓",
			report.Accounted, report.Total))
	} else {
		lines = append(lines, fmt.Sprintf(
			"🧮 Accounting:          only %d/%d accounted for — %d escaped every bucket!",
			report.Accounted, report.Total, report.Unaccounted))
	}
	lines = append(lines, "")
	if report.VerificationPassed && report.AccountingOK {
		lines = append(lines, "🎉 ALL LINKS VERIFIED — NO DATA LOSS!")
	} else {
		lines = append(lines, fmt.Sprintf("⚠️ %d link(s) need retry:", len(report.FailedLinks)))
		lines = append(lines, "")
		for _, fl := range report.FailedLinks {
			errMsg := "unknown"
			if fl.Error != nil {
				errMsg = *fl.Error
			}
			lines = append(lines,
				fmt.Sprintf("  ❌ %s", fl.URL),
				fmt.Sprintf("     Type:   %s", fl.Type),
				fmt.Sprintf("     Status: %s", fl.Status),
				fmt.Sprintf("     Error:  %s", errMsg),
				"")
		}
	}
	lines = append(lines, strings.Repeat("=", 60))

	// The enqueue legs (v0.63.1: GITHUB LINKS ONLY; v0.63.3: settled
	// repos never re-enter the retry queue).
	wsFailed := 0
	failedURLs := []string{}
	for _, fl := range report.FailedLinks {
		if fl.URL == "" {
			continue
		}
		if fl.Type == "github" {
			failedURLs = append(failedURLs, fl.URL)
		} else {
			wsFailed++
		}
	}
	if wsFailed > 0 {
		plural := "s"
		if wsFailed == 1 {
			plural = ""
		}
		s.SpeakOnBus(fmt.Sprintf(
			"ℹ️ %d website link%s failed their fetches — the Websites pipeline's own "+
				"retry queue and master table own them (never the repos retry queue).",
			wsFailed, plural), "info")
	}
	settledSkips := 0
	if cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg)); err == nil {
		kept := failedURLs[:0]
		for _, u := range failedURLs {
			if cache.IsRepoSettled(u) {
				settledSkips++
			} else {
				kept = append(kept, u)
			}
		}
		failedURLs = kept
		cache.Close()
	}
	if settledSkips > 0 {
		s.SpeakOnBus(fmt.Sprintf(
			"🤝 %d settled repo link(s) kept out of the retry queue — addressed "+
				"(the owner's law), never re-fetched.", settledSkips), "info")
	}
	if len(failedURLs) > 0 {
		if cache, err := store.NewCacheDB(cacheDBPathOf(s.cfg)); err == nil {
			for _, u := range failedURLs {
				cache.AddFailed(u, "failed Phase 3 verification")
			}
			cache.Close()
			s.SpeakOnBus(fmt.Sprintf(
				"📥 Added %d failed link(s) to the retry queue. Click "+
					"'🔄 Retry Failed' in the Bot tab to reprocess them.",
				len(failedURLs)), "info")
		} else {
			s.SpeakOnBus(fmt.Sprintf("❌ Failed to enqueue retries: %v", err), "error")
		}
	}
	return "🔍 VAULT VERIFICATION REPORT", strings.Join(lines, "\n")
}

// pyOr is the Python `x or default` twin for the report's manifest
// fields.
func pyOr(s, def string) string {
	if strings.TrimSpace(s) == "" {
		return def
	}
	return s
}

// isDirPathApp is the shell's local twin (heroui.go's), at the
// engine's side.
func isDirPathApp(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}

// ---------------------------------------------------------------------------
// The LLM router (_build_llamml_call's twin — the tools' own shared
// composition).
// ---------------------------------------------------------------------------

// buildLinkLLM is _build_llm_call: the provider router over the LIVE
// config, returning the chat closure (mirror's llm(messages) shape,
// JSON mode on) + the "  LLM: …" line. abortMsg non-empty = the
// provider abort (the reference raises SystemExit with the message —
// the tool's own documented provider-error exit code is 2; the
// reference's GUI worker crashed on the string code and LOST the
// message, a display bug the port does not copy — the message rides
// the captured stderr instead. Recorded divergence).
func (s *Settings) buildLinkLLM(modelOverride, llamaAbort string) (
	llm mirror.RecallLLM, llmLine, abortMsg string) {
	provider := strings.ToLower(strings.TrimSpace(pyStrCfg(s.cfg, "llm_provider")))
	if provider == "" {
		provider = "ollama"
	}
	// num_ctx = int(config.get('llm_num_ctx', DEFAULT_NUM_CTX) or 0) or None
	numCtx := llmclient.DefaultNumCtx
	if v, ok := s.cfg.Get("llm_num_ctx"); ok && v != nil {
		numCtx = pyIntOr(v, 0)
	}
	suffix := ""
	if numCtx > 0 {
		suffix = fmt.Sprintf(" (num_ctx=%d)", numCtx)
	}
	timeoutS := pyFloatCfgOr(s.cfg, "llm_timeout_s", 300)

	switch provider {
	case "llamacpp":
		base := strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_api_url"))
		if base == "" {
			base = llmclient.LLamaCppDefaultBase + "/v1"
		}
		apiURL := llmclient.NormalizeLlamaCppAPIURL(base)
		key := strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_api_key"))
		model := strings.TrimSpace(pyStrCfg(s.cfg, "llamacpp_model"))
		if model == "" && s.Deps.ProbeLlama != nil {
			probe := s.Deps.ProbeLlama(apiURL, key, llmclient.LLamaCppProbeTimeoutS)
			if probe == nil || !probe.Found {
				if s.Deps.DetectLlama != nil {
					if det := s.Deps.DetectLlama(key); det != nil && det.Found {
						probe = det
					}
				}
			}
			if probe != nil && probe.Found {
				apiURL = probe.BaseURL + "/v1"
				if probe.Model != nil {
					model = *probe.Model
				}
			}
		}
		if model == "" {
			return nil, "", llamaAbort
		}
		if s.Deps.LinkCloudChat == nil {
			return nil, "", ""
		}
		line := fmt.Sprintf("  LLM: %s via llama.cpp %s%s", model, apiURL, suffix)
		return func(messages []llmclient.Message) (string, error) {
			reply, errText := s.Deps.LinkCloudChat(apiURL, key, model, messages,
				numCtx, timeoutS)
			return chatResult(reply, errText)
		}, line, ""

	case "cloud":
		apiURL := strings.TrimSpace(pyStrCfg(s.cfg, "cloud_api_url"))
		model := strings.TrimSpace(pyStrCfg(s.cfg, "cloud_model"))
		if apiURL == "" || model == "" {
			return nil, "", "cloud provider needs config.json cloud_api_url + cloud_model"
		}
		if s.Deps.LinkCloudChat == nil {
			return nil, "", ""
		}
		key := strings.TrimSpace(pyStrCfg(s.cfg, "cloud_api_key"))
		line := fmt.Sprintf("  LLM: %s via %s%s", model, apiURL, suffix)
		return func(messages []llmclient.Message) (string, error) {
			reply, errText := s.Deps.LinkCloudChat(apiURL, key, model, messages,
				numCtx, timeoutS)
			return chatResult(reply, errText)
		}, line, ""

	default: // ollama
		host := cfgNestedStr(s.cfg, "ollama", "base_url", "")
		if host == "" {
			host = "http://127.0.0.1:11434"
		}
		model := strings.TrimSpace(modelOverride)
		if model == "" {
			model = cfgNestedStr(s.cfg, "ollama", "model", "")
		}
		if model == "" {
			model = "llama3"
		}
		if s.Deps.LinkOllamaChat == nil {
			return nil, "", ""
		}
		line := fmt.Sprintf("  LLM: %s via Ollama (%s)%s", model, host, suffix)
		return func(messages []llmclient.Message) (string, error) {
			reply, errText := s.Deps.LinkOllamaChat(host, model, messages,
				numCtx, timeoutS)
			return chatResult(reply, errText)
		}, line, ""
	}
}

// chatResult adapts the seams' (reply, errText) to the closure's
// (string, error).
func chatResult(reply, errText string) (string, error) {
	if errText != "" {
		return "", fmt.Errorf("%s", errText)
	}
	return reply, nil
}

// ---------------------------------------------------------------------------
// 🪝 Recall hooks (dry-run) — add_recall_hooks.main(argv=[]) twin.
// ---------------------------------------------------------------------------

// RecallHooksDryRun is run_recall_hooks_dryrun: the SAFE default
// (apply=false, no sample, no force) through the TestWorker law.
func (s *Settings) RecallHooksDryRun() {
	s.runLinkingTool(labelRecallHooks, s.recallHooksTool)
}

func (s *Settings) recallHooksTool(o *linkToolOut) int {
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		o.stderrLine("GitHub vault not set or not found (config vault_path or --vault).")
		return 1
	}
	plan := mirror.PlanRecallHooks(vault, false)
	o.printf("GitHub vault: %s", vault)
	o.printf("Notes needing a hook: %d", len(plan))
	if len(plan) == 0 {
		o.printf("Nothing to do.")
		return 0
	}
	llm, llmLine, abort := s.buildLinkLLM("",
		"llama.cpp provider needs a running llama-server (its model is "+
			"detected automatically) or config.json llamacpp_api_url/llamacpp_model")
	if abort != "" {
		o.stderrLine(abort)
		return 2
	}
	if llm == nil {
		return 0
	}
	o.printf("%s", llmLine)
	o.printf("Mode: DRY RUN (nothing written)")
	results := mirror.RunRecallHooks(plan, llm, false, 0, o.out, nil)
	report := mirror.RecallReport(results, vault, false)
	outPath := filepath.Join(mirror.ReportDir(), "recall-dryrun.md")
	if err := writeReportFile(outPath, report); err != nil {
		o.crashLine(err)
		return 2
	}
	o.printf("\nReport: %s", outPath)
	o.printf("Dry run — re-run with --apply to write %d hook(s).", len(results))
	return 0
}

// writeReportFile is the tools' open(...,'w').write twin.
func writeReportFile(path, content string) error {
	return os.WriteFile(path, []byte(content), 0o644)
}

// ---------------------------------------------------------------------------
// 🔗 Build link suggestions (suggest mode) — build_links.main(argv=[])
// twin (the _suggest half).
// ---------------------------------------------------------------------------

// BuildLinkSuggestions is run_link_suggestions: the SAFE default
// (suggest mode — the one write is the Suggestions note) through the
// TestWorker law.
func (s *Settings) BuildLinkSuggestions() {
	s.runLinkingTool(labelLinkSuggest, s.buildLinksTool)
}

func (s *Settings) buildLinksTool(o *linkToolOut) int {
	manual := strings.TrimSpace(pyStrCfg(s.cfg, "manual_vault_path"))
	if manual == "" || !isDirPathApp(manual) {
		o.stderrLine("Manual Notes vault not set or not found (config " +
			"manual_vault_path or --manual-vault) — the Suggestions note lives " +
			"in its Library/.")
		return 1
	}
	library := filepath.Join(manual, "Library")
	if !isDirPathApp(library) {
		o.stderrLine("Library/ not found in the manual vault — run " +
			"tools/mirror_manual.py --apply first (the Related blocks live in " +
			"the mirror copies).")
		return 1
	}
	vaults := []mirror.VaultPath{
		{Vault: notestate.VaultGithub, Path: strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))},
		{Vault: notestate.VaultWebsites, Path: strings.TrimSpace(pyStrCfg(s.cfg, "website_vault_path"))},
	}
	parts := make([]string, 0, 2)
	for _, vp := range vaults {
		if vp.Path != "" {
			parts = append(parts, vp.Vault+": "+vp.Path)
		}
	}
	o.printf("Collecting recall fields %s", strings.Join(parts, " · "))
	collected := mirror.CollectNoteFields(vaults)
	if collected.Duplicates > 0 {
		o.printf("  (skipped %d duplicate-source note(s))", collected.Duplicates)
	}
	if len(collected.Order) == 0 {
		o.printf("No managed notes found — nothing to link.")
		return 1
	}
	o.printf("  %d managed note(s)", len(collected.Order))
	fieldsByKey := make(map[string]*mirror.FieldsEntry, len(collected.Order))
	for _, key := range collected.Order {
		e := collected.BySource[key]
		fieldsByKey[key] = &e
	}

	// The embeddings (stale only) — the router seam (faked at
	// construction; the production wiring IS mirror.BuildEmbeddingCall).
	if s.Deps.LinkEmbedCall == nil {
		return 0
	}
	embed, label, err := s.Deps.LinkEmbedCall(s.cfg, "")
	if err != nil {
		o.stderrLine(fmt.Sprintf("Embedding error: %v", err))
		o.stderrLine("  (llama-server needs --embeddings for /v1/embeddings; " +
			"Ollama needs the model pulled, e.g. 'ollama pull nomic-embed-text')")
		return 2
	}
	o.printf("  Embeddings: %s", label)
	modelKey := "default"
	if segs := strings.Split(label, "·"); len(segs) > 0 {
		if t := strings.TrimSpace(segs[len(segs)-1]); t != "" {
			modelKey = t
		}
	}
	estore, err := mirror.NewEmbeddingStore("cache.db")
	if err != nil {
		o.crashLine(err)
		return 2
	}
	defer estore.Close()
	linkStore, err := mirror.NewLinkStore("cache.db")
	if err != nil {
		o.crashLine(err)
		return 2
	}
	defer linkStore.Close()

	entries := make([]mirror.RefreshEntry, 0, len(collected.Order))
	for _, key := range collected.Order {
		entries = append(entries, mirror.RefreshEntry{
			SourceURL: key, Fields: collected.BySource[key].Fields})
	}
	stats, err := mirror.RefreshEmbeddings(entries, embed, estore,
		"linking", modelKey, false, o.out)
	if err != nil {
		o.crashLine(err)
		return 2
	}
	o.printf("  embeddings: %d embedded, %d fresh, %d failed",
		stats.Embedded, stats.SkippedFresh, stats.Failed)

	loaded, err := estore.Load("linking", modelKey)
	if err != nil {
		o.crashLine(err)
		return 2
	}
	// The vector space = EVERY stored vector (stale keys included, the
	// reference's own law). The iteration order: the collect order
	// first, then the stale keys sorted — the reference rides the
	// SQLite row order; the port pins the deterministic twin (the
	// tie-stability law's own demand). Recorded divergence.
	var vecEntries []mirror.VecEntry
	for _, key := range collected.Order {
		if v, ok := loaded[key]; ok && len(v.Vec) > 0 {
			vecEntries = append(vecEntries, mirror.VecEntry{Key: key, Vec: v.Vec})
		}
	}
	var stale []string
	for key, v := range loaded {
		if _, live := fieldsByKey[key]; !live && len(v.Vec) > 0 {
			stale = append(stale, key)
		}
	}
	sort.Strings(stale)
	for _, key := range stale {
		vecEntries = append(vecEntries, mirror.VecEntry{Key: key, Vec: loaded[key].Vec})
	}
	if len(vecEntries) < 2 {
		o.stderrLine("Fewer than 2 embeddings — nothing to compare (is the " +
			"embedding model pulled? e.g. ollama pull nomic-embed-text).")
		return 1
	}

	candidates := mirror.FindCandidatePairs(vecEntries, 8, 0.30)
	known, err := linkStore.KnownPairsSet()
	if err != nil {
		o.crashLine(err)
		return 2
	}
	fresh := make([]mirror.CandidatePair, 0, len(candidates))
	for _, c := range candidates {
		if !known[[2]string{c.A, c.B}] && !known[[2]string{c.B, c.A}] {
			fresh = append(fresh, c)
		}
	}
	o.printf("  candidates: %d above 0.30 — %d new (%d already suggested)",
		len(candidates), len(fresh), len(candidates)-len(fresh))

	llm, llmLine, abort := s.buildLinkLLM("",
		"llama.cpp provider needs a running llama-server or config.json "+
			"llamacpp_api_url/llamacpp_model")
	if abort != "" {
		o.stderrLine(abort)
		return 2
	}
	if llm == nil {
		return 0
	}
	o.printf("%s", llmLine)
	confirmed := 0
	asked := 0
	for _, c := range fresh {
		if asked >= 50 {
			o.printf("  … stopping at --max-confirm 50 for this run (re-run to continue)")
			break
		}
		fieldsA := mirror.RecallFields{}
		fieldsB := mirror.RecallFields{}
		nameA, nameB := c.A, c.B
		if e, ok := fieldsByKey[c.A]; ok {
			fieldsA = e.Fields
			nameA = e.Fields.Name
		}
		if e, ok := fieldsByKey[c.B]; ok {
			fieldsB = e.Fields
			nameB = e.Fields.Name
		}
		asked++
		related, reason := mirror.ConfirmPair(fieldsA, fieldsB, llm)
		if related != nil && *related {
			if _, err := linkStore.Suggest(c.A, c.B, c.Score, reason); err != nil {
				o.crashLine(err)
				return 2
			}
			confirmed++
			o.printf("  + related: %s ↔ %s — %s",
				helpers.CutRunes(nameA, 40), helpers.CutRunes(nameB, 40), reason)
		} else if related != nil {
			// a NO is recorded as rejected — never asked again.
			rejReason := "LLM: " + pyOr(reason, "not related")
			if _, err := linkStore.Suggest(c.A, c.B, c.Score, rejReason); err != nil {
				o.crashLine(err)
				return 2
			}
			if _, err := linkStore.Decide(c.A, c.B, "rejected"); err != nil {
				o.crashLine(err)
				return 2
			}
		} else {
			o.printf("  ? unparseable answer for %s ↔ %s — skipped (will retry next run)",
				c.A, c.B)
		}
	}

	// The Suggestions note — the ONLY write of the suggest step.
	note, err := mirror.BuildSuggestionsNote(linkStore, fieldsByKey)
	if err != nil {
		o.crashLine(err)
		return 2
	}
	out := mirror.SuggestionsPath(manual)
	if err := storage.AtomicWriteText(out, note); err != nil {
		o.crashLine(err)
		return 2
	}
	counts, err := linkStore.Counts()
	if err != nil {
		o.crashLine(err)
		return 2
	}
	o.printf("\nSuggestions note: %s", out)
	o.printf("  pending %d · approved %d · rejected %d (cap %d links/note)",
		counts["pending"], counts["approved"], counts["rejected"], mirror.LinksCap)
	o.printf("Tick a checkbox to approve, strike a line through to reject, " +
		"then run with --collect.")
	_ = confirmed
	return 0
}

// ---------------------------------------------------------------------------
// 📁 Recategorize Notes (dashboard.py:500-660) — the two-door flow:
// the open door walks + answers the table payload; the apply door
// performs the moves.
// ---------------------------------------------------------------------------

// recategorizeLockedRE is the locked-note check (v0.12.0 Phase 3
// §4.4: the owner's own moves are corrections — never re-categorized,
// not even listed).
var recategorizeLockedRE = regexp.MustCompile(
	`(?m)^category_locked:\s*true`)

// recategorizeCatRE is the category frontmatter reader.
var recategorizeCatRE = regexp.MustCompile(`category:\s*(.+)`)

// RecategorizeRow is one dialog row (the note, its current category).
type RecategorizeRow struct {
	Path     string `json:"path"`
	Name     string `json:"name"`
	Category string `json:"category"`
}

// RecategorizeModel is the open door's payload: the box (non-empty
// title = the reference's custom message box instead of the table) or
// the table itself (notes + the category list for the dropdowns).
type RecategorizeModel struct {
	BoxTitle   string            `json:"boxTitle"`
	BoxBody    string            `json:"boxBody"`
	Notes      []RecategorizeRow `json:"notes"`
	Categories []string          `json:"categories"`
}

// RecategorizeChange is one apply row: the note's path + the captured
// current category (the dialog's own row — the reference compares the
// new pick against the CAPTURED current, the open-time walk's value)
// + the new pick.
type RecategorizeChange struct {
	Path        string `json:"path"`
	OldCategory string `json:"oldCategory"`
	Category    string `json:"category"`
}

// RecategorizeResult is the apply door's verdict.
type RecategorizeResult struct {
	Moved int `json:"moved"`
}

// RecategorizeOpen is recategorize_notes' walk half: the vault check,
// the walk (locked notes counted + skipped with the 🔒 line), the
// no-notes box, and the table payload.
func (s *Settings) RecategorizeOpen() *RecategorizeModel {
	if s.linkBusy() {
		return &RecategorizeModel{}
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		return &RecategorizeModel{
			BoxTitle: "Error",
			BoxBody:  "Please select a valid vault first.",
		}
	}
	var notes []RecategorizeRow
	locked := 0
	_ = filepath.WalkDir(vault, func(path string, d os.DirEntry, err error) error {
		if err != nil || d == nil {
			return nil
		}
		if d.IsDir() {
			return nil
		}
		if !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}
		data, rerr := os.ReadFile(path)
		if rerr != nil {
			return nil
		}
		head := helpers.CutRunes(string(data), 1000)
		if recategorizeLockedRE.MatchString(head) {
			locked++
			return nil
		}
		cat := "Unknown"
		if m := recategorizeCatRE.FindStringSubmatch(head); m != nil {
			cat = strings.TrimSpace(m[1])
		}
		notes = append(notes, RecategorizeRow{Path: path, Name: d.Name(), Category: cat})
		return nil
	})
	if locked > 0 {
		s.SpeakOnBus(fmt.Sprintf(
			"🔒 %d locked note(s) skipped — your own moves are never "+
				"re-categorized.", locked), "info")
	}
	if len(notes) == 0 {
		return &RecategorizeModel{
			BoxTitle: "Recategorize",
			BoxBody:  "No notes found in vault.",
		}
	}
	return &RecategorizeModel{
		Notes:      notes,
		Categories: constants.CategoryKeys,
	}
}

// RecategorizeApply is _apply's twin: for every row whose pick differs
// from the captured current — rewrite the category frontmatter (every
// occurrence, the reference's global re.sub), write to the category
// folder, remove the old file when the path changed, and move the
// cache row. Per-note failures speak; the run continues. The finish
// line carries the moved count.
//
// Divergences recorded: the cache connection opens ONCE per apply (the
// reference opens per note — same end state); update_dashboard() rides
// the port's Settings-window Dashboard page (it renders on open — no
// main-window page to refresh).
func (s *Settings) RecategorizeApply(changes []RecategorizeChange) *RecategorizeResult {
	if s.linkBusy() {
		return &RecategorizeResult{}
	}
	vault := strings.TrimSpace(pyStrCfg(s.cfg, "vault_path"))
	if vault == "" || !isDirPathApp(vault) {
		s.SpeakOnBus("❌ Recategorize: no valid vault selected.", "error")
		return &RecategorizeResult{}
	}
	var cache *store.CacheDB
	if c, err := store.NewCacheDB(cacheDBPathOf(s.cfg)); err == nil {
		cache = c
		defer cache.Close()
	}
	moved := 0
	for _, ch := range changes {
		if ch.Category == ch.OldCategory {
			continue
		}
		name := filepath.Base(ch.Path)
		data, err := os.ReadFile(ch.Path)
		if err != nil {
			s.SpeakOnBus(fmt.Sprintf("Failed to recategorize %s: %v", name, err), "error")
			continue
		}
		full := string(data)
		replacement := "category: " + ch.Category
		newContent := recategorizeCatRE.ReplaceAllStringFunc(full,
			func(string) string { return replacement })
		folder := "Uncategorized"
		for _, e := range constants.CategoryFolders {
			if e.Key == ch.Category {
				folder = e.Folder
				break
			}
		}
		newFolder := filepath.Join(vault, folder)
		if err := os.MkdirAll(newFolder, 0o755); err != nil {
			s.SpeakOnBus(fmt.Sprintf("Failed to recategorize %s: %v", name, err), "error")
			continue
		}
		newPath := filepath.Join(newFolder, name)
		if err := os.WriteFile(newPath, []byte(newContent), 0o644); err != nil {
			s.SpeakOnBus(fmt.Sprintf("Failed to recategorize %s: %v", name, err), "error")
			continue
		}
		oldAbs, _ := filepath.Abs(ch.Path)
		newAbs, _ := filepath.Abs(newPath)
		if oldAbs != newAbs {
			if err := os.Remove(ch.Path); err != nil {
				s.SpeakOnBus(fmt.Sprintf("Failed to recategorize %s: %v", name, err), "error")
				continue
			}
		}
		if cache != nil {
			if _, err := cache.RecategorizeNote(ch.Path, newPath, ch.Category); err != nil {
				s.SpeakOnBus(fmt.Sprintf(
					"   ⚠️ Cache update failed for %s: %v", name, err), "warning")
			}
		}
		moved++
	}
	s.SpeakOnBus(fmt.Sprintf("📁 Recategorized %d notes", moved), "success")
	return &RecategorizeResult{Moved: moved}
}
