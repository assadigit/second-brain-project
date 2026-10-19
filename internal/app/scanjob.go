// scanjob.go — T-71 slice (d): the _vault_scan_job engine
// (worker_jobs.py:686-1287) — the app's production Scan seam. The
// whole story, one composition: the one-open cache (with its closing
// honesty line), the hygiene pass over BOTH machine vaults (the
// v0.68.0 stray-notes door), the vault picker's scope law (v0.84.0),
// the per-vault plan (the LLM's validated proposal), the ask (the
// GUI door rides the bus's confirm gate; the Telegram door rides the
// Worker's scan trio; an unpaired door defers safely — nothing moves,
// nothing deletes), the picking door's arithmetic, and the merged
// verdict. Never raises (the TestWorker contract).
package app

import (
	"fmt"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/vaulthygiene"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// The door config word (scan_confirm_door): 'gui' (the desktop's own
// launches — the bus's gate) or 'telegram' (the away-from-desk door —
// the Worker's scan trio).
const scanDoorTelegram = "telegram"

// scanLLMRouter builds the provider-routing llmCall with the
// vaultscan task override (worker_jobs.py:500-559 — _scan_llm_call's
// shape; the CLI's engine router's twin).
func scanLLMRouter(cfg *storage.OrderedMap, log func(msg, level string)) vaultscan.LLMCallFn {
	cfgPlain := cfgPlainMap(cfg)
	provider := pyStrCfg(cfg, "llm_provider")
	if provider == "" {
		provider = "ollama"
	}
	numCtx := 0
	if v, ok := cfg.Get("llm_num_ctx"); ok {
		if f, is := v.(float64); is {
			numCtx = int(f)
		}
	}
	timeoutS := 300.0
	if v, ok := cfg.Get("llm_timeout_s"); ok {
		if f, is := v.(float64); is && f > 0 {
			timeoutS = f
		}
	}
	warn := func(m string) { log(m, "warning") }
	return func(messages []llmclient.Message, task string) string {
		switch provider {
		case "cloud":
			model := llmclient.ResolveTaskModel(cfgPlain, task,
				pyStrCfg(cfg, "cloud_model"))
			out, _ := llmclient.CloudChat(pyStrCfg(cfg, "cloud_api_url"),
				pyStrCfg(cfg, "cloud_api_key"), model, messages, timeoutS,
				&llmclient.ChatOptions{JSONMode: true, NumCtx: numCtx,
					OnWarn: warn})
			return out
		case "llamacpp":
			model := llmclient.ResolveTaskModel(cfgPlain, task,
				pyStrCfg(cfg, "llamacpp_model"))
			out, _ := llmclient.CloudChat(
				llmclient.NormalizeLlamaCppAPIURL(pyStrCfg(cfg, "llamacpp_api_url")),
				pyStrCfg(cfg, "llamacpp_api_key"), model, messages, timeoutS,
				&llmclient.ChatOptions{JSONMode: true, NumCtx: numCtx,
					OnWarn: warn})
			return out
		default:
			// the reference's own reads (worker_jobs.py:541-549): the
			// NESTED ollama.base_url (default
			// http://localhost:11434) + ollama.model.
			base := strings.TrimRight(cfgNestedStr(cfg, "ollama", "base_url",
				"http://localhost:11434"), "/")
			model := llmclient.ResolveTaskModel(cfgPlain, task,
				cfgNestedStr(cfg, "ollama", "model", ""))
			out, _ := llmclient.OllamaChat(
				llmclient.NewOllamaClient(base),
				model, messages, timeoutS,
				&llmclient.OllamaChatOptions{NumCtx: numCtx, OnWarn: warn})
			return out
		}
	}
}

// RunVaultScan is the production Scan seam (the command's engine).
func RunVaultScan(cfg *storage.OrderedMap, vaults []string, ev ScanEvents) *ScanResult {
	if ev.Log == nil {
		ev.Log = func(string, string) {}
	}
	if ev.Progress == nil {
		ev.Progress = func(string, int, int) {}
	}
	log := ev.Log
	prog := ev.Progress
	cfgPlain := cfgPlainMap(cfg)
	story := &ScanResult{Verdict: "clean"}

	// v0.71.0 — THE ONE-OPEN CACHE: one shared NoteReadCache rides the
	// whole scan; under a dry run it rides the shadow cache.db.
	cachePath := cacheDBPathOf(cfg)
	if dryrun.IsEnabled() {
		cachePath = dryrun.ShadowCachePath(cachePath)
	}
	readCache := notereads.New(cachePath)
	closeCache := func() {
		stats := readCache.Close(false)
		if stats.Opens > 0 || stats.DbHits > 0 {
			log(fmt.Sprintf("⚡ The one-open law: %d note open(s) this scan "+
				"— %d unchanged note(s) served from the last scan's "+
				"memory, nothing re-opened", stats.Opens, stats.DbHits), "info")
		}
	}
	defer closeCache()

	// v0.68.0 — THE STRAY-NOTES DOOR: heal BOTH machine vaults first
	// (the heal is not the scan — the maintenance law).
	hygiene := vaulthygiene.HealMachineVaults(cfgPlain, log, func(label string, value, max int) {
		prog(label, value, max)
	}, readCache)
	hygieneChanged := vaulthygiene.HygieneTotal(hygiene) > 0
	story.HygieneTotal = vaulthygiene.HygieneTotal(hygiene)

	wvault := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	rvault := strings.TrimSpace(pyStrCfg(cfg, "vault_path"))
	hasWv := wvault != "" && fsIsDirPath(wvault)
	hasRv := rvault != "" && fsIsDirPath(rvault)

	// v0.84.0 — THE VAULT PICKER's scope: junk-tolerant (a string
	// rides as one choice; unknown kinds fall away; anything broken
	// degrades to the whole desk).
	scope := []string(nil)
	if vaults != nil {
		scope = []string{}
		raw := vaults
		for _, k := range raw {
			k = strings.ToLower(strings.TrimSpace(k))
			if (k == "websites" || k == "repos") && !containsStr(scope, k) {
				scope = append(scope, k)
			}
		}
	}
	wantWv := scope == nil || containsStr(scope, "websites")
	wantRv := scope == nil || containsStr(scope, "repos")
	websitesRan := wantWv && hasWv
	reposRan := wantRv && hasRv
	story.WebsitesRan = websitesRan
	story.ReposRan = reposRan

	if !websitesRan && !reposRan {
		// the honest per-scope miss (the picker's truth rides home).
		var okLine, missLine string
		switch {
		case wantWv && wantRv:
			okLine = "✅ Vault scan: the hygiene pass healed your machine vaults (see the lines above) — no vault is set, so there is nothing else to scan."
			missLine = "⚠️ Vault scan: no Websites vault and no GitHub Repos vault are set (Settings → 📁 Vault) — nothing to scan."
		case wantWv:
			okLine = "✅ Vault scan: the hygiene pass healed your machine vaults (see the lines above) — the Websites vault you chose is not set/found, so there is nothing else to scan."
			missLine = "⚠️ Vault scan: the Websites vault you chose is not set or not found (Settings → 📁 Vault) — nothing to scan."
		case wantRv:
			okLine = "✅ Vault scan: the hygiene pass healed your machine vaults (see the lines above) — the GitHub Repos vault you chose is not set/found, so there is nothing else to scan."
			missLine = "⚠️ Vault scan: the GitHub Repos vault you chose is not set or not found (Settings → 📁 Vault) — nothing to scan."
		default:
			okLine = "⚠️ Vault scan: no vault was chosen — nothing to scan."
			missLine = okLine
		}
		if hygieneChanged {
			log(okLine, "success")
			story.Success = true
			story.Verdict = "clean"
			return story
		}
		log(missLine, "warning")
		story.Success = true
		story.Verdict = "clean"
		return story
	}

	// The confirm door: 'gui' (the bus gate) is the desktop default;
	// 'telegram' rides the Worker trio; an unpaired telegram defers.
	door := strings.ToLower(strings.TrimSpace(pyStrCfg(cfg, "scan_confirm_door")))
	confirm := func(plan *vaultscan.Plan, label string) (string, map[string][]string) {
		if door == scanDoorTelegram {
			return telegramScanConfirm(cfg, plan, log)
		}
		ans := ev.Confirm(plan)
		return ans.Verdict, ans.Selection
	}

	// The per-vault stories.
	verdicts := []string{}
	if websitesRan {
		statePath := cacheDBPathOf(cfg)
		if dryrun.IsEnabled() {
			statePath = dryrun.ShadowCachePath(statePath)
		}
		state, serr := websitepipeline.NewWebsiteStateDB(statePath)
		if serr != nil {
			state = nil
		}
		// the Websites leg's own opening line, byte-for-byte.
		log(vaultscan.ScanPrefix+" reading the vault — the folder walk, "+
			"your auto-delete marks, and the LLM's filing eyes…", "info")
		v := scanOneVault(cfg, wvault, "websites", "Websites vault",
			scanLLMRouter(cfg, log), log, prog, readCache, state, nil, nil, confirm)
		verdicts = append(verdicts, v.Verdict)
		story.Lines = append(story.Lines, v.Lines...)
		if state != nil {
			func() {
				defer func() { recover() }()
				state.Close()
			}()
		}
	}
	if reposRan {
		var reposState *notestate.DB
		if nd, nerr := notestate.New(cacheDBPathOf(cfg)); nerr == nil {
			reposState = nd
			defer func() {
				defer func() { recover() }()
				_ = reposState.Close()
			}()
		}
		var reposCache *store.CacheDB
		if rc, rerr := store.NewCacheDB(cacheDBPathOf(cfg)); rerr == nil {
			reposCache = rc
			defer func() { _ = reposCache.Close() }()
		}
		v := scanOneVault(cfg, rvault, "repos", "GitHub Repos vault",
			scanLLMRouter(cfg, log), log, prog, readCache, nil, reposCache, reposState, confirm)
		verdicts = append(verdicts, v.Verdict)
		story.Lines = append(story.Lines, v.Lines...)
	}

	// The merged verdict: the worst word wins (error > timeout/defer >
	// declined > confirmed > clean).
	story.Verdict = mergeScanVerdicts(verdicts)
	story.Success = story.Verdict != "error"
	return story
}

// scanOneVault is _scan_one_vault (worker_jobs.py:1000-1160): the
// plan, the proposal lines, the gate, the picking arithmetic, the
// apply, the done line — each step speaking the reference's own
// sentences. Never raises (the TestWorker contract).
func scanOneVault(cfg *storage.OrderedMap, vaultPath, kind, label string,
	llmCall vaultscan.LLMCallFn, log func(msg, level string),
	prog func(label string, value, max int), readCache *notereads.Cache,
	state *websitepipeline.WebsiteStateDB, cacheDB *store.CacheDB,
	noteStateDB *notestate.DB,
	confirm func(plan *vaultscan.Plan, label string) (string, map[string][]string)) (story struct {
	Verdict string
	Lines   []string
}) {
	lines := []string{}
	say := func(msg, level string) {
		log(msg, level)
		lines = append(lines, level+"|"+msg)
	}
	defer func() {
		if r := recover(); r != nil {
			say(fmt.Sprintf("💥 Vault scan (%s) failed: %s", label, pyCrashText(r)), "error")
			story.Verdict = "error"
		}
	}()

	plan := vaultscan.BuildScanPlan(vaultPath, llmCall, log,
		cfgPlainMap(cfg), func(lb string, v, m int) { prog(lb, v, m) },
		readCache, kind, noteStateDB)

	// the deletions list over the vault's OWN kind (the repos vault
	// rides DeletionsRepos — the repobanish gate's items; the
	// websites vault rides Deletions).
	deletions := len(plan.Deletions)
	if kind == "repos" {
		deletions = len(plan.DeletionsRepos)
	}
	moves := len(plan.Moves)
	newFolders := len(plan.NewFolders)
	consolidations := len(plan.Consolidations)
	consolidationNotes := 0
	for _, c := range plan.Consolidations {
		consolidationNotes += len(c.Merge)
	}
	inv := plan.Inventory
	say(fmt.Sprintf("🔍 Vault scan (%s): the vault holds %d note(s) in "+
		"%d folder(s) — %d orphaned, %d uncategorized",
		label, inv.TotalNotes, inv.FolderCount, inv.RootNotes,
		inv.UncategorizedNotes), "info")
	hasDeletions := deletions > 0
	hasFiling := (moves + newFolders) > 0
	hasConsolidations := consolidations > 0
	if !hasDeletions && !hasFiling && !hasConsolidations {
		say(fmt.Sprintf("🔍 Vault scan (%s): the library is clean — no "+
			"auto-delete marks, no duplicate notes, no filing to propose. "+
			"Nothing to ask, nothing to do.", label), "info")
		story.Verdict = "clean"
		return story
	}
	if hasDeletions {
		say(fmt.Sprintf("🗑️ %d note(s) carry your delete mark (the v0.60.1 "+
			"grammar — frontmatter + body tags)", deletions), "info")
	}
	if hasConsolidations {
		for _, c := range plan.Consolidations {
			keep := c.Keep.File
			if keep == "" {
				keep = "?"
			}
			for _, m := range c.Merge {
				say(fmt.Sprintf("🧲 Consolidation proposed: %s — merges into "+
					"%s (site: %s)", m.File, keep, c.Site), "info")
			}
		}
	}
	if hasFiling {
		for _, nf := range plan.NewFolders {
			say(fmt.Sprintf("🌱 New folder proposed: %s", nf), "info")
		}
		for _, m := range plan.Moves {
			reason := ""
			if m.Reason != "" {
				reason = " (" + m.Reason + ")"
			}
			say(fmt.Sprintf("🚚 Move proposed: %s — %s → %s%s",
				m.Note, m.From, m.To, reason), "info")
		}
	}

	// ---- the gate: nothing moves or deletes without the owner ----
	verdict, selection := confirm(&plan, label)
	verdict = normalizeScanVerdictWord(verdict)
	story.Verdict = verdict
	if verdict != "confirmed" {
		why := map[string]string{
			"declined": "you said keep everything",
			"timeout":  "no answer in time",
			"defer":    "the confirm door could not carry the ask",
		}[verdict]
		if why == "" {
			why = verdict
		}
		say(fmt.Sprintf("👌 Vault scan (%s): nothing moved, nothing deleted "+
			"(%s) — the vault stays exactly as it is", label, why), "info")
		return story
	}

	// v0.73.0 — THE PICKING DOOR's arithmetic (the planned counts
	// speak the PICKED numbers from here on).
	if selection != nil {
		picked := vaultscan.PickedCounts(&plan, vaultscan.NormalizeSelection(selection))
		proposed := deletions + newFolders + moves + consolidationNotes
		pickedN := picked["deletions"] + picked["new_folders"] +
			picked["moves"] + picked["consolidation_notes"]
		say(fmt.Sprintf("🧹 You picked %d of %d proposal row(s) — applying "+
			"only the ticked ones (the rest wait for the next scan)",
			pickedN, proposed), "info")
	}
	say(fmt.Sprintf("✅ You confirmed the plan (%s) — applying it now "+
		"(nothing is rewritten, only moved)", label), "info")

	sel := vaultscan.NormalizeSelection(selection)
	rep := vaultscan.ApplyScanPlan(&plan, vaultPath, state, log, sel,
		kind, cacheDB, noteStateDB)
	say(fmt.Sprintf("🔍 Vault scan (%s) done — %d note(s) removed (resting "+
		"in .trash/banished, never to be fetched again), %d duplicate "+
		"note(s) merged into their site's note (resting in "+
		".trash/consolidated — one note per site), %d note(s) re-filed, "+
		"%d new folder(s) created",
		label, rep.Banished, rep.Consolidated, rep.NotesMoved,
		rep.FoldersCreated), "info")
	if story.Verdict == "" {
		story.Verdict = "confirmed"
	}
	return story
}

// normalizeScanVerdictWord maps the gate's answers onto the story's
// vocabulary (the telegram door's 'applied'/'defer' words fold into
// the reference's verdict set).
func normalizeScanVerdictWord(v string) string {
	v = strings.ToLower(strings.TrimSpace(v))
	switch v {
	case "confirmed":
		return "confirmed"
	case "declined":
		return "declined"
	case "timeout":
		return "timeout"
	case "defer", "":
		return "defer"
	default:
		return v
	}
}

// mergeScanVerdicts folds the per-vault verdicts into the story's
// one word (the reference's merged-verdict law: the worst wins).
func mergeScanVerdicts(vs []string) string {
	rank := map[string]int{"clean": 0, "confirmed": 1, "declined": 2,
		"timeout": 3, "defer": 3, "error": 5}
	worst, worstRank := "clean", 0
	for _, v := range vs {
		if r, ok := rank[v]; ok && r > worstRank {
			worst, worstRank = v, r
		}
	}
	if worst == "defer" {
		// the reference's story speaks 'timeout' for the safe defer
		// when no vault confirmed anything.
		return "timeout"
	}
	return worst
}

// telegramScanConfirm is the away-from-desk door
// (integrations/scan_confirm.py): propose → poll (3 s cadence, the
// configured window) → report. Unpaired: the safe defer.
func telegramScanConfirm(cfg *storage.OrderedMap, plan *vaultscan.Plan,
	log func(msg, level string)) (string, map[string][]string) {
	s := workerclient.NewSync(cfgPlainMap(cfg))
	if !s.IsEnabled() {
		return "defer", nil
	}
	timeoutS := 300.0
	if v, ok := cfg.Get("scan_confirm_timeout_s"); ok {
		if f, is := v.(float64); is && f > 0 {
			timeoutS = f
		}
	}
	// The ask's payload: the deletions + the moves (capped the way
	// the banish ask caps its list — 20 shown, the rest counted).
	items := []any{}
	moves := plan.Moves
	delPairs := scanDeletionPairs(plan)
	for _, d := range delPairs {
		if len(items) >= 20 {
			break
		}
		items = append(items, map[string]any{
			"kind": "delete", "title": d[0], "url": d[1],
		})
	}
	for _, m := range moves {
		if len(items) >= 20 {
			break
		}
		items = append(items, map[string]any{
			"kind": "move", "note": m.Note, "to": m.To,
		})
	}
	summary := fmt.Sprintf("%d deletion(s), %d move(s), %d new folder(s)",
		len(delPairs), len(moves), len(plan.NewFolders))
	id, why := s.ProposeScan(len(delPairs), len(moves), len(plan.NewFolders),
		items, summary, timeoutS)
	if id == nil {
		log("⚠️ The Telegram scan ask could not be proposed ("+why+
			") — the safe defer applies", "warning")
		return "defer", nil
	}
	deadline := time.Now().Add(time.Duration(timeoutS * float64(time.Second)))
	for {
		time.Sleep(3 * time.Second)
		if time.Now().After(deadline) {
			s.ReportScanResult(*id, "timeout", 0, 0, 0)
			return "timeout", nil
		}
		switch s.ScanStatus(*id) {
		case "confirmed":
			// The telegram door is all-or-nothing: no selection rides.
			return "confirmed", nil
		case "declined":
			return "declined", nil
		case "timeout", "expired":
			return "timeout", nil
		}
	}
}

// scanDeletionPairs flattens the plan's deletions into (title, url)
// pairs over whichever kind the plan rode.
func scanDeletionPairs(plan *vaultscan.Plan) [][2]string {
	out := [][2]string{}
	if plan == nil {
		return out
	}
	if plan.VaultKind == "repos" || len(plan.DeletionsRepos) > 0 {
		for _, d := range plan.DeletionsRepos {
			out = append(out, [2]string{d.Title, d.URL})
		}
		return out
	}
	for _, d := range plan.Deletions {
		out = append(out, [2]string{d.Title, d.URL})
	}
	return out
}

// cfgNestedStr reads a nested string with a default (the reference's
// (config.get(sec) or {}).get(key, def) shape).
func cfgNestedStr(cfg *storage.OrderedMap, section, key, def string) string {
	if cfg == nil {
		return def
	}
	v, _ := cfg.Get(section)
	om, ok := v.(*storage.OrderedMap)
	if !ok {
		return def
	}
	sv, present := om.Get(key)
	if !present {
		return def
	}
	if s, is := sv.(string); is && strings.TrimSpace(s) != "" {
		return s
	}
	return def
}

func containsStr(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}
