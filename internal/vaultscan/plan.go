package vaultscan

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/repobanish"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// PlanSummary mirrors the plan's inventory summary cell.
type PlanSummary struct {
	TotalNotes         int `json:"total_notes"`
	RootNotes          int `json:"root_notes"`
	UncategorizedNotes int `json:"uncategorized_notes"`
	HandNotes          int `json:"hand_notes"`
	TrashNotes         int `json:"trash_notes"`
	CoreNotes          int `json:"core_notes"`
	LockedNotes        int `json:"locked_notes"`
	CuratedNotes       int `json:"curated_notes"`
	FolderCount        int `json:"folder_count"`
}

// Plan mirrors build_scan_plan's dict — the shape the confirm door and
// the apply pass both consume.
type Plan struct {
	Deletions        []websitepipeline.PendingBanishment
	KeptHandwritten  int
	Moves            []MoveRow
	NewFolders       []string
	Consolidations   []Consolidation
	Summary          string
	Inventory        PlanSummary
	RegexFiled       int
	LockedNotes      int
	LLMCandidates    int
	WeighedHomes     int
	CuratedNotes     int
	Reviewed         []string
	VaultKind        string
	VaultLabel       string
	FullInventory    *Inventory                         `json:"-"`
	DeletionsRepos   []repobanish.PendingRepoBanishment `json:"-"`
	KeptHandwrittenR int                                `json:"-"`
}

// BuildScanPlan mirrors build_scan_plan (v0.61.0 — THE SCAN'S OWN
// EYES): part one, the DELETIONS, is the banishment gate's own items
// (ONE grammar, both doors, deduped); part two, the FILING, is the
// regex brain's deterministic passes plus the LLM's validated
// proposal; the consolidations (ONE NOTE PER SITE) ride the plan
// before any filing decision. Pure reads + the injected llmCall;
// never touches a file; never raises.
func BuildScanPlan(vaultPath string, llmCall LLMCallFn, log LogFn,
	config map[string]any, progress ProgressFn,
	readCache *notereads.Cache, kind string,
	noteStateDB *notestate.DB) Plan {
	if log == nil {
		log = noopLog
	}
	if config == nil {
		config = map[string]any{}
	}
	tooBroad := cfgIntOr(config, "scan_too_broad_threshold", DefaultTooBroad)
	maxNotes := cfgIntOr(config, "scan_max_notes", DefaultMaxNotes)
	batch := cfgIntOr(config, "scan_llm_batch", DefaultLLMBatch)
	out := Plan{
		VaultKind:  vaultKindOf(kind),
		VaultLabel: kindLabel(kind),
	}
	if vaultPath == "" || !fsIsDir(vaultPath) {
		return out
	}
	// v0.71.0 — THE ONE-OPEN LAW: a caller's cache rides in; a bare
	// call gets a RAM-only one so even standalone callers stop paying
	// the triple-open tax.
	ownCache := false
	if readCache == nil {
		readCache = notereads.New("")
		ownCache = true
	}
	// the deletions: the gate's own eyes (ONE grammar).
	if isReposKind(kind) {
		gate := repobanish.ScanPendingRepoBanishments(vaultPath,
			noteStateDB, log)
		out.DeletionsRepos = gate.Items
		out.KeptHandwrittenR = gate.KeptHandwritten
	} else {
		gate := websitepipeline.ScanPendingBanishments(vaultPath, log,
			readCache)
		out.Deletions = gate.Items
		out.KeptHandwritten = gate.KeptHandwritten
	}
	inventory := ScanVaultInventory(vaultPath, log, progress, readCache, kind)
	out.FullInventory = &inventory
	out.Inventory = PlanSummary{
		TotalNotes:         inventory.TotalNotes,
		RootNotes:          inventory.RootNotes,
		UncategorizedNotes: inventory.UncategorizedNotes,
		HandNotes:          inventory.HandNotes,
		TrashNotes:         inventory.TrashNotes,
		CoreNotes:          inventory.CoreNotes,
		LockedNotes:        inventory.LockedNotes,
		CuratedNotes:       inventory.CuratedNotes,
		FolderCount:        len(inventory.Folders),
	}
	out.LockedNotes = inventory.LockedNotes
	out.CuratedNotes = inventory.CuratedNotes
	// v0.73.0 — THE LOCK LAW's deletion door: a locked note never rides
	// the scan's deletion list — whichever door marked it.
	if len(out.Deletions) > 0 || len(out.DeletionsRepos) > 0 {
		lockedCanon := LockedCanonicals(&inventory, kind)
		if len(lockedCanon) > 0 {
			if isReposKind(kind) {
				var kept []repobanish.PendingRepoBanishment
				for _, d := range out.DeletionsRepos {
					if lockedCanon[d.Canonical] {
						name := baseNoExt(d.Path)
						if name == "" {
							name = baseNoExt(d.Title)
						}
						if name == "" {
							name = "a note"
						}
						log(fmt.Sprintf("🔒 %s is locked — its delete "+
							"mark waits (the lock outranks the mark; "+
							"unlock the note to let it ride a future "+
							"scan)", name), "info")
						continue
					}
					kept = append(kept, d)
				}
				if len(kept) != len(out.DeletionsRepos) {
					log(fmt.Sprintf("🔒 %d locked note(s) left off the "+
						"deletion list — out of the radar entirely",
						len(out.DeletionsRepos)-len(kept)), "info")
				}
				out.DeletionsRepos = kept
			} else {
				var kept []websitepipeline.PendingBanishment
				for _, d := range out.Deletions {
					if lockedCanon[d.Canonical] {
						name := baseNoExt(d.Path)
						if name == "" {
							name = baseNoExt(d.Title)
						}
						if name == "" {
							name = "a note"
						}
						log(fmt.Sprintf("🔒 %s is locked — its delete "+
							"mark waits (the lock outranks the mark; "+
							"unlock the note to let it ride a future "+
							"scan)", name), "info")
						continue
					}
					kept = append(kept, d)
				}
				if len(kept) != len(out.Deletions) {
					log(fmt.Sprintf("🔒 %d locked note(s) left off the "+
						"deletion list — out of the radar entirely",
						len(out.Deletions)-len(kept)), "info")
				}
				out.Deletions = kept
			}
		}
	}
	// v0.64.0 — the regex half's newest eye: same-site notes become
	// consolidation proposals BEFORE any filing decision.
	out.Consolidations = ScanSiteDuplicates(&inventory, log, kind)
	var candidates []*NoteRow
	for _, n := range MoveCandidates(&inventory, tooBroad) {
		candidates = append(candidates, &n)
	}
	// v0.71.0 — THE RETIRING NEVER RIDE THE FILING CHAIN: a note a
	// consolidation proposal merges into its site's keeper is LEAVING
	// the vault — the keeper rides the chain alone.
	if len(out.Consolidations) > 0 {
		mergedPaths := map[string]bool{}
		for _, c := range out.Consolidations {
			for _, m := range c.Merge {
				mergedPaths[m.Path] = true
			}
		}
		var kept []*NoteRow
		for _, n := range candidates {
			if !mergedPaths[n.Path] {
				kept = append(kept, n)
			}
		}
		if len(kept) != len(candidates) {
			log(fmt.Sprintf("🧲 %d candidate note(s) ride a "+
				"consolidation proposal (they retire into their "+
				"site's keeper) — the filing chain never sees them",
				len(candidates)-len(kept)), "info")
		}
		candidates = kept
	}
	out.LLMCandidates = len(candidates)
	if len(candidates) == 0 {
		if ownCache {
			readCache.Close(false)
		}
		return out
	}
	// v0.70.0 — THE DECLARED-HOME PASS (deterministic moves ride FIRST
	// so the validator's taken-set gives them precedence).
	declaredMoves, extraFolders := DeclaredHomeMoves(candidates,
		vaultPath, log, kind)
	// v0.71.0 — THE RESIDUE CHAIN: declared home → the domain's settled
	// siblings → the taxonomy's own vocabulary. The repos chain is the
	// declared-home pass plus the LLM's genuine thinking pile (the
	// domain-shaped signals have no repos twin, honestly).
	var sibMoves, taxMoves []RawMove
	var sibExtra, taxExtra []string
	claimed := map[string]bool{}
	for _, m := range declaredMoves {
		claimed[m.Note] = true
	}
	var llmCandidates []*NoteRow
	if isReposKind(kind) {
		for _, n := range candidates {
			if !claimed[n.File] {
				llmCandidates = append(llmCandidates, n)
			}
		}
	} else {
		residue := make([]*NoteRow, 0, len(candidates))
		for _, n := range candidates {
			if !claimed[n.File] {
				residue = append(residue, n)
			}
		}
		sibMoves, sibExtra = SiblingDomainMoves(residue, &inventory,
			vaultPath, log)
		for _, m := range sibMoves {
			claimed[m.Note] = true
		}
		residue = residue[:0]
		for _, n := range candidates {
			if !claimed[n.File] {
				residue = append(residue, n)
			}
		}
		taxMoves, taxExtra = TaxonomyTagMoves(residue, vaultPath, log,
			config)
		for _, m := range taxMoves {
			claimed[m.Note] = true
		}
		for _, n := range candidates {
			if !claimed[n.File] {
				llmCandidates = append(llmCandidates, n)
			}
		}
	}
	for _, extra := range append(append([]string{}, sibExtra...), taxExtra...) {
		if extra != "" && !inList(extra, extraFolders) {
			extraFolders = append(extraFolders, extra)
		}
	}
	regexFiled := len(claimed)
	out.RegexFiled = regexFiled
	out.LLMCandidates = len(llmCandidates)
	if regexFiled > 0 {
		log(fmt.Sprintf("⚡ The regex brain filed %d of %d candidate "+
			"note(s) without asking the LLM (declared home %d · "+
			"siblings %d · taxonomy %d) — %d note(s) left for the "+
			"thinking pass", regexFiled, len(candidates), len(declaredMoves),
			len(sibMoves), len(taxMoves), len(llmCandidates)), "info")
	}
	// v0.74.0 — THE WEIGHED EVIDENCE: every note the LLM must THINK
	// about rides with a weighed shortlist.
	excludePaths := map[string]bool{}
	for _, n := range candidates {
		excludePaths[n.Path] = true
	}
	withHomes := AttachWeighedHomes(llmCandidates, &inventory,
		excludePaths, config, tooBroad, kind)
	out.WeighedHomes = withHomes
	if withHomes > 0 {
		log(fmt.Sprintf("⚖️ %d of %d thinking-pass note(s) ride the LLM "+
			"with a weighed shortlist of candidate homes — the vault's "+
			"own evidence (the notes already on those folders, the "+
			"domain's settled siblings, shared tags and title words, "+
			"the declared home, the taxonomy's vocabulary) scored each "+
			"home 0.0-1.0; the numbers advise, the LLM decides",
			withHomes, len(llmCandidates)), "info")
	}
	data := LLMProposal{NewFolders: []string{}, Moves: []RawMove{}}
	if llmCall != nil && len(llmCandidates) > 0 {
		data = llmPropose(&inventory, llmCandidates, llmCall, log,
			tooBroad, maxNotes, batch, progress, kind)
	}
	// the deterministic moves join the raw proposal BEFORE validation —
	// they meet the same laws through the same door as the guest's.
	data.Moves = append(append(append(append([]RawMove{},
		declaredMoves...), sibMoves...), taxMoves...), data.Moves...)
	moves, newFolders := ValidateMoves(data, candidates, &inventory, log,
		extraFolders, vaultPath, kind)
	out.Moves = moves
	out.NewFolders = newFolders
	out.Summary = cutRunesStr(data.Summary, 400)
	// v0.78.0 — the census's review list: the notes the LLM examined
	// and proposed nothing for, minus the notes that earned a
	// validated move.
	movePaths := map[string]bool{}
	for _, m := range moves {
		movePaths[m.Path] = true
	}
	for _, p := range data.Reviewed {
		if p != "" && !movePaths[p] {
			out.Reviewed = append(out.Reviewed, p)
		}
	}
	if ownCache {
		readCache.Close(false)
	}
	return out
}

// vaultKindOf mirrors the plan's vault_kind cell:
// str(kind or VAULT_KIND_WEBSITES).
func vaultKindOf(kind string) string {
	if helpers.Strip(kind) == "" {
		return VaultKindWebsites
	}
	return kind
}

// cfgIntOr mirrors int(cfg.get(key, default) or default): the config
// map's value may arrive as any JSON shape; a falsy value falls back
// to the default.
func cfgIntOr(config map[string]any, key string, def int) int {
	v, ok := config[key]
	if !ok || !pyTruthyAny(v) {
		return def
	}
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		if n, ok2 := pyIntLoose(helpers.Strip(t)); ok2 {
			return n
		}
		return def
	}
	return def
}

func baseNoExt(p string) string {
	if p == "" {
		return ""
	}
	base := filepathBase(p)
	return strings.TrimSuffix(base, pathExt(base))
}

func filepathBase(p string) string {
	if i := strings.LastIndexAny(p, `/\`); i >= 0 {
		p = p[i+1:]
	}
	return p
}

func pathExt(p string) string {
	i := strings.LastIndexByte(p, '.')
	if i <= strings.LastIndexAny(p, `/\`) {
		return ""
	}
	return p[i:]
}
