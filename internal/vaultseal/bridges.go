// bridges.go — the four config bridges (seal_from_config /
// websites_seal_from_config / manual_seal_from_config /
// resync_from_config) + _local_only_for (the belt join: the stems
// always, the directory's FULL belt for the websites vault, the
// config's own local_only extras — and manual_seal's deliberately
// NO-belt law: "a backup that filters the owner's own files is not a
// backup").
package vaultseal

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/directorysync"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// cfgGet reads a key out of a Python-dict-shaped config (nil /
// *storage.OrderedMap / map[string]any — the workerclient pattern).
func cfgGet(config any, key string) (any, bool) {
	switch t := config.(type) {
	case nil:
		return nil, false
	case *storage.OrderedMap:
		return t.Get(key)
	case map[string]any:
		v, ok := t[key]
		return v, ok
	default:
		return nil, false
	}
}

// cfgValue is cfgGet without the ok.
func cfgValue(config any, key string) any {
	v, _ := cfgGet(config, key)
	return v
}

// dictOfValue answers the dict-shaped value at key (or nil).
func dictOfValue(config any, key string) any {
	switch t := cfgValue(config, key).(type) {
	case *storage.OrderedMap:
		return t
	case map[string]any:
		return t
	default:
		return nil
	}
}

// pyStrOf mirrors str(v or "").
func pyStrOf(v any) string {
	switch t := v.(type) {
	case nil:
		return ""
	case string:
		return t
	case bool:
		if t {
			return "True"
		}
		return "False"
	default:
		return helpers.PyStr(v)
	}
}

// strList reads a list-of-strings value ([]any items through str()).
func strList(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	var out []string
	for _, x := range items {
		out = append(out, pyStrOf(x))
	}
	return out
}

// localOnlyFor mirrors _local_only_for: the local-only patterns for
// one vault's seal — the note-stem globs ALWAYS (a cobalt twin in the
// repos vault's _review stays local too); the private-category FOLDER
// forms as well when fullBelt (the Websites vault — the directory's
// own belt). The config's own vaultseal.local_only list joins both
// (the owner's extension point beyond the shipped law).
//
// The reference wraps the directory_sync import in try/except → []
// (a broken belt must never break a seal); Go's static import cannot
// fail — the functions themselves never error — so the failure door
// has no analog (a documented convergence).
func localOnlyFor(vaultPath string, fullBelt bool, config any) []string {
	cfg := dictOfValue(config, "vaultseal")
	var extra []string
	for _, x := range strList(cfgValue(cfg, "local_only")) {
		if strings.TrimSpace(x) != "" {
			extra = append(extra, x)
		}
	}
	var base []string
	if fullBelt {
		base = directorysync.PrivateExclusionPatterns(vaultPath, nil)
	} else {
		base = directorysync.LocalOnlyNotePatterns(nil)
	}
	seen := map[string]bool{}
	for _, p := range base {
		seen[p] = true
	}
	out := append([]string(nil), base...)
	for _, p := range extra {
		if p != "" && !seen[p] {
			seen[p] = true
			out = append(out, p)
		}
	}
	return out
}

// SealerFromConfig is the construction half of seal_from_config (the
// reference's tests capture the constructor kwargs; the Go surface
// inspects the built sealer).
func SealerFromConfig(config any, log LogFn) *VaultSeal {
	cfg := dictOfValue(config, "vaultseal")
	return NewVaultSeal(
		pyStrOf(cfgValue(config, "vault_path")),
		pyStrOf(cfgValue(config, "github_token")),
		pyStrOf(cfgValue(cfg, "repo_name")),
		pyTruthyOr(cfgValue(cfg, "auto_push"), true),
		pyTruthyOr(cfgValue(cfg, "enabled"), true),
		DefaultTimeout,
		log,
		pyTruthy(cfgValue(cfg, "create_repo")),
		localOnlyFor(pyStrOf(cfgValue(config, "vault_path")), false, config),
	)
}

// SealFromConfig mirrors seal_from_config: build a VaultSeal from the
// app's config and seal once. Reads vault_path, github_token, and the
// vaultseal section {enabled, repo_name, auto_push, create_repo}.
func SealFromConfig(config any, runSummary *RunSummary, log LogFn) *SealResult {
	return SealerFromConfig(config, log).Seal(runSummary)
}

// websitesSealerFromConfig is the construction half (the reference's
// bridge builds the VaultSeal inline; the split gives the belt tests
// the captured-kwargs surface).
func websitesSealerFromConfig(config any, log LogFn) *VaultSeal {
	vaultPath := strings.TrimSpace(pyStrOf(cfgValue(config, "website_vault_path")))
	repoName := strings.TrimSpace(pyStrOf(cfgValue(config, "website_repo_name")))
	return NewVaultSeal(
		vaultPath,
		pyStrOf(cfgValue(config, "github_token")),
		repoName,
		true, true, DefaultTimeout, log, false,
		// v0.89.0 — THE FULL BELT (the directory's own): the private
		// folders + the local-only stems never ride this mirror.
		localOnlyFor(vaultPath, true, config),
	)
}

// WebsitesSealFromConfig mirrors websites_seal_from_config: seal the
// WEBSITES vault into its own private repo (Phase 1, v0.10.0) — a
// second, independent VaultSeal instance. A no-op unless the websites
// pipeline is ON and a websites vault path + repo name are configured.
func WebsitesSealFromConfig(config any, runSummary *RunSummary, log LogFn) *SealResult {
	pipelines := dictOfValue(config, "pipelines")
	if !pyTruthy(cfgValue(pipelines, "websites")) {
		return &SealResult{Private: true, SkippedReason: "websites pipeline is off"}
	}
	vaultPath := strings.TrimSpace(pyStrOf(cfgValue(config, "website_vault_path")))
	if vaultPath == "" {
		return &SealResult{Private: true, SkippedReason: "no websites vault configured"}
	}
	repoName := strings.TrimSpace(pyStrOf(cfgValue(config, "website_repo_name")))
	// v0.83.0 — no silent default: an unnamed websites repo means NOT
	// configured; nothing is created, nothing is pushed.
	if repoName == "" {
		return &SealResult{Private: true, SkippedReason: "no websites backup repo configured " +
			"(website_repo_name is empty)"}
	}
	return websitesSealerFromConfig(config, log).Seal(runSummary)
}

// manualSealerFromConfig is the construction half (the belt tests'
// captured-kwargs surface) — NO belt, by the owner's law.
func manualSealerFromConfig(config any, log LogFn) *VaultSeal {
	vaultPath := strings.TrimSpace(pyStrOf(cfgValue(config, "manual_vault_path")))
	repoName := strings.TrimSpace(pyStrOf(cfgValue(config, "manual_repo_name")))
	return NewVaultSeal(
		vaultPath,
		pyStrOf(cfgValue(config, "github_token")),
		repoName,
		true, true, DefaultTimeout, log, false, nil,
	)
}

// ManualSealFromConfig mirrors manual_seal_from_config: seal the
// PERSONAL vault (the Manual Notes vault) into its own private repo
// (v0.84.0 — the third vault, the owner's own writing). Either the
// path or the repo missing = a skipped no-op; nothing is ever created
// (the v0.83.0 creation guard).
//
// v0.89.0 — deliberately NO local-only belt here: the Personal vault
// is the owner's own writing, sealed WHOLE by design (a backup that
// filters the owner's own files is not a backup). The belt applies to
// the vaults the DIRECTORY's law governs.
func ManualSealFromConfig(config any, runSummary *RunSummary, log LogFn) *SealResult {
	vaultPath := strings.TrimSpace(pyStrOf(cfgValue(config, "manual_vault_path")))
	if vaultPath == "" {
		return &SealResult{Private: true, SkippedReason: "no personal vault configured"}
	}
	repoName := strings.TrimSpace(pyStrOf(cfgValue(config, "manual_repo_name")))
	if repoName == "" {
		return &SealResult{Private: true, SkippedReason: "no personal backup repo configured " +
			"(manual_repo_name is empty — the repo radar will detect it)"}
	}
	return manualSealerFromConfig(config, log).Seal(runSummary)
}

// vaultSealCtor is the class-patch seam the reference's resync tests
// ride (mock.patch.object(vs_mod, "VaultSeal", FakeSeal)): production
// binds NewVaultSeal; the tests record the constructions.
var vaultSealCtor = NewVaultSeal

// ResyncFromConfig mirrors resync_from_config: the owner's "make
// GitHub follow the vault" action for EVERY configured vault mirror —
// the GitHub vault + the Websites vault + the Personal vault (each
// into its own repo, exactly like the post-run seals). The resync is
// an explicit owner action, so the vaultseal.enabled / auto_push
// switches do not gate it (a token is still required — checked per
// vault). Returns the insertion-ordered {github, websites, manual}
// map.
func ResyncFromConfig(config any, log LogFn) *storage.OrderedMap {
	out := storage.NewOrderedMap()

	vsCfg := dictOfValue(config, "vaultseal")
	vaultPath := pyStrOf(cfgValue(config, "vault_path"))
	githubSealer := vaultSealCtor(
		vaultPath,
		pyStrOf(cfgValue(config, "github_token")),
		pyStrOf(cfgValue(vsCfg, "repo_name")),
		true, true, DefaultTimeout, log, false,
		localOnlyFor(vaultPath, false, config),
	)
	if strings.TrimSpace(vaultPath) == "" {
		res := &SealResult{Private: true, SkippedReason: "no GitHub vault configured"}
		if githubSealer.repoName != "" {
			res.RepoName = githubSealer.repoName
		}
		out.Set("github", res)
	} else if strings.TrimSpace(pyStrOf(cfgValue(vsCfg, "repo_name"))) == "" {
		// v0.83.0 — an unnamed repo is NOT a default-seal anymore.
		out.Set("github", &SealResult{Private: true,
			SkippedReason: "no backup repo configured (vaultseal.repo_name is empty — " +
				"nothing created, nothing pushed)"})
	} else {
		out.Set("github", githubSealer.Resync(""))
	}

	pipelines := dictOfValue(config, "pipelines")
	webPath := strings.TrimSpace(pyStrOf(cfgValue(config, "website_vault_path")))
	webRepo := strings.TrimSpace(pyStrOf(cfgValue(config, "website_repo_name")))
	if pyTruthy(cfgValue(pipelines, "websites")) && webPath != "" && webRepo != "" {
		webSealer := vaultSealCtor(
			webPath,
			pyStrOf(cfgValue(config, "github_token")),
			webRepo,
			true, true, DefaultTimeout, log, false,
			localOnlyFor(webPath, true, config),
		)
		out.Set("websites", webSealer.Resync(""))
	} else if pyTruthy(cfgValue(pipelines, "websites")) && webPath != "" && webRepo == "" {
		out.Set("websites", &SealResult{Private: true,
			SkippedReason: "no websites backup repo configured (website_repo_name is empty)"})
	} else {
		out.Set("websites", &SealResult{Private: true,
			SkippedReason: "websites pipeline off or no websites vault"})
	}

	// v0.84.0 — the PERSONAL vault's mirror joins the resync (the
	// third vault, the owner's own writing). No pipeline gate (like
	// the seal); the same not-configured skip vocabulary.
	personalPath := strings.TrimSpace(pyStrOf(cfgValue(config, "manual_vault_path")))
	personalRepo := strings.TrimSpace(pyStrOf(cfgValue(config, "manual_repo_name")))
	if personalPath != "" && personalRepo != "" {
		personalSealer := vaultSealCtor(
			personalPath,
			pyStrOf(cfgValue(config, "github_token")),
			personalRepo,
			true, true, DefaultTimeout, log, false, nil,
		)
		out.Set("manual", personalSealer.Resync(""))
	} else if personalPath != "" && personalRepo == "" {
		out.Set("manual", &SealResult{Private: true,
			SkippedReason: "no personal backup repo configured (manual_repo_name is empty)"})
	} else {
		out.Set("manual", &SealResult{Private: true,
			SkippedReason: "no personal vault configured"})
	}
	return out
}

// pyTruthy mirrors Python truthiness for the config value shapes.
func pyTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case *storage.OrderedMap:
		return t.Len() > 0
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}

// pyTruthyOr mirrors cfg.get(key, default) + Python truthiness.
func pyTruthyOr(v any, fallback bool) bool {
	switch t := v.(type) {
	case nil:
		return fallback
	case bool:
		return t
	case string:
		if t == "" {
			return fallback
		}
		return true
	default:
		if pyTruthy(v) {
			return true
		}
		return fallback
	}
}
