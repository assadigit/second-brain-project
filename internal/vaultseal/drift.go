// drift.go — v0.89.0 THE DRIFT SEAL's pure half (gui/main_window/
// backup_seal.py lines 53-198: "no Qt, no GUI thread — the plan and
// the verdict are testable headlessly; the QThread above only runs
// them"). The owner's report that produced it: "Why it says 800
// something things are awaiting, but the app itself says everything
// is uptodate?" — the caught-up path now seals the dirty vaults.
//
// DriftSealPlan is PURE (the batch-end path's own guards, evaluated
// in the same order); RunDriftSeal runs the enabled legs + the
// directory publish — git and network, no Qt — with THE QUIET LAW: a
// leg that seals nothing and fails nothing stays SILENT (the
// caught-up empty state survives a clean vault); a leg that actually
// seals (or fails) speaks once.
//
// The publish door rides DriftSealDeps.PublishAfterBatch — T-49's
// directorysync surface; until it is wired the runner reports the
// not-wired warning on the websites-sealed path (the honest line, not
// a silent skip — the site's index drift must be spoken). The GUI
// worker task wires the real publish; the QThread wrapper is the
// future GUI task's own.
package vaultseal

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/directorysync"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// driftSealLeg mirrors one _DRIFT_SEAL_LEGS row: (key, human label,
// guard) — each guard mirrors the batch-end path's OWN condition in
// processing_control._on_batch_finished, verbatim.
type driftSealLeg struct {
	key   string
	label string
	guard func(config any) bool
}

// driftSealLegs mirrors _DRIFT_SEAL_LEGS: the repos seal rides
// vaultseal.enabled + vault_path; the websites seal rides
// pipelines.websites + website_vault_path; the personal seal rides
// manual_vault_path + manual_repo_name.
var driftSealLegs = []driftSealLeg{
	{"repos", "repos vault", func(c any) bool {
		return pyTruthyOr(cfgValue(driftDictDoor(c, "vaultseal"), "enabled"), true) &&
			strings.TrimSpace(driftStrDoor(c, "vault_path")) != ""
	}},
	{"websites", "websites vault", func(c any) bool {
		return pyTruthy(cfgValue(driftDictDoor(c, "pipelines"), "websites")) &&
			strings.TrimSpace(driftStrDoor(c, "website_vault_path")) != ""
	}},
	{"personal", "personal vault", func(c any) bool {
		return strings.TrimSpace(driftStrDoor(c, "manual_vault_path")) != "" &&
			strings.TrimSpace(driftStrDoor(c, "manual_repo_name")) != ""
	}},
}

// driftDictDoor mirrors (c.get(key) or {}): a FALSY value answers the
// empty dict (the .get defaults ride); a TRUTHY non-dict value raises
// in Python (AttributeError on .get) — the Go door panics the same
// way and DriftSealPlan's recover turns it into the silent [].
func driftDictDoor(c any, key string) any {
	v := cfgValue(c, key)
	if pyTruthy(v) {
		switch v.(type) {
		case *storage.OrderedMap, map[string]any:
			return v
		}
		panic(fmt.Sprintf("AttributeError: %T object has no attribute 'get'", v))
	}
	return nil
}

// driftStrDoor mirrors (c.get(key) or ”).strip(): a FALSY value
// answers ""; a TRUTHY non-string raises in Python (no .strip) — the
// Go door panics the same way.
func driftStrDoor(c any, key string) string {
	v := cfgValue(c, key)
	if pyTruthy(v) {
		if s, ok := v.(string); ok {
			return s
		}
		panic(fmt.Sprintf("AttributeError: %T object has no attribute 'strip'", v))
	}
	return ""
}

// DriftSealPlan mirrors _drift_seal_plan: the enabled drift-seal legs
// for this config (the batch-end path's own guards, evaluated in the
// same order). An empty plan is honest — nothing is sealed, nothing
// published. The reference's broken-config try/except → [] maps to
// the recover door: a guard that meets a truthy-but-wrong-typed value
// (Python's AttributeError on .get/.strip) is the same silence.
func DriftSealPlan(config any) (plan []string) {
	defer func() {
		if recover() != nil {
			plan = nil
		}
	}()
	for _, leg := range driftSealLegs {
		if leg.guard(config) {
			plan = append(plan, leg.key)
		}
	}
	return plan
}

// DriftSealDeps carries the drift seal's three legs + the directory
// publish door (the reference resolves them through module imports;
// the Go surface injects them — the tests stub all four, exactly like
// the reference's mock.patch.object legs).
type DriftSealDeps struct {
	SealRepos         func(config any, log LogFn) *SealResult
	SealWebsites      func(config any, log LogFn) *SealResult
	SealPersonal      func(config any, log LogFn) *SealResult
	PublishAfterBatch func(config any, log LogFn) bool // T-49's door
}

// DefaultDriftSealDeps binds the production legs (the vaultseal
// bridges) + the directory publish door (T-49's directorysync
// surface — publish_after_batch, the full golden-rule-5 cycle).
func DefaultDriftSealDeps() *DriftSealDeps {
	return &DriftSealDeps{
		SealRepos:    func(config any, log LogFn) *SealResult { return SealFromConfig(config, nil, log) },
		SealWebsites: func(config any, log LogFn) *SealResult { return WebsitesSealFromConfig(config, nil, log) },
		SealPersonal: func(config any, log LogFn) *SealResult { return ManualSealFromConfig(config, nil, log) },
		PublishAfterBatch: func(config any, log LogFn) bool {
			// the reference's backup_seal passes its (msg, level) log
			// into publish_after_batch's (msg) log; the plain lines tag
			// "info" (the documented adapter)
			if log == nil {
				return directorysync.PublishAfterBatch(config, nil)
			}
			return directorysync.PublishAfterBatch(config, func(msg string) {
				log(msg, "info")
			})
		},
	}
}

// RunDriftSeal mirrors _run_drift_seal(config, log, manual): run the
// enabled seal legs + the directory publish (the site's leg).
// Returns (line, level) — ("", "info") is the quiet leg: nothing to
// say. manual opens v0.90.0's FAST SYNC door (the publish ALWAYS runs
// when the websites leg is enabled; the verdict ALWAYS SPEAKS).
func RunDriftSeal(config any, deps *DriftSealDeps, log LogFn, manual bool) (string, string) {
	if log == nil {
		log = noopLog
	}
	var parts, errors []string
	websitesSealed := false
	if deps == nil {
		deps = DefaultDriftSealDeps()
	}
	fns := map[string]func(any, LogFn) *SealResult{
		"repos":    deps.SealRepos,
		"websites": deps.SealWebsites,
		"personal": deps.SealPersonal,
	}
	labels := map[string]string{}
	for _, leg := range driftSealLegs {
		labels[leg.key] = leg.label
	}
	plan := DriftSealPlan(config)
	for _, key := range plan {
		fn := fns[key]
		if fn == nil {
			continue
		}
		result := runDriftLeg(fn, config, log, labels[key], &errors)
		if result == nil {
			continue
		}
		if result.Error != "" {
			errors = append(errors, labels[key]+": "+cutRunes(result.Error, 200))
			continue
		}
		if result.Sealed {
			parts = append(parts, itoa(result.FilesChanged)+" file(s) → the "+
				labels[key]+" mirror")
			if key == "websites" {
				websitesSealed = true
			}
		}
	}
	published := false
	// v0.90.0 — the manual door asks for the site's leg whenever the
	// websites pipeline is on; the automatic path keeps the
	// sealed-only law (an unchanged vault publishes an unchanged
	// index — the worker's own freshness law would skip it anyway).
	planHasWebsites := false
	for _, key := range plan {
		if key == "websites" {
			planHasWebsites = true
		}
	}
	if websitesSealed || (manual && planHasWebsites) {
		if deps.PublishAfterBatch == nil {
			errors = append(errors,
				"directory publish: publish_after_batch is not wired in this build "+
					"(the directorysync publish door is T-49's own task)")
		} else {
			published = deps.PublishAfterBatch(config, log)
		}
		if manual && !published && !hasDirPublishError(errors) {
			errors = append(errors,
				"the directory publish did not land — the desk log above carries "+
					"the Worker's own reason (no worker URL? a down Worker?)")
		}
	}
	name := "Drift seal"
	glyph := "🛡️"
	if manual {
		name = "Fast Sync"
		glyph = "⚡"
	}
	if len(parts) > 0 {
		line := glyph + " " + name + ": " + strings.Join(parts, "; ")
		if published {
			line += "; 🌐 directory published → the website follows"
		}
		if len(errors) > 0 {
			line += " — but " + strings.Join(errors, "; ")
			return line, "warning"
		}
		return line, "success"
	}
	if published {
		// manual-only in practice (the automatic path publishes only
		// after a seal): the mirrors were already current and the
		// site's leg re-landed — the whole triad, said in one line.
		line := glyph + " " + name + ": the mirrors were already current — " +
			"nothing to seal; 🌐 directory published → the website follows"
		if len(errors) > 0 {
			line += " — but " + strings.Join(errors, "; ")
			return line, "warning"
		}
		return line, "success"
	}
	if len(errors) > 0 {
		return "⚠️ " + name + ": " + strings.Join(errors, "; "), "warning"
	}
	if manual {
		// the manual door's own law: an explicit press ALWAYS gets its
		// verdict — the clean triad is a success line, not the
		// automatic path's silence.
		return "✅ " + name + ": the triad is already in sync — nothing to " +
			"seal, nothing to publish", "success"
	}
	return "", "info" // the quiet leg: nothing to say
}

func hasDirPublishError(errors []string) bool {
	for _, e := range errors {
		if strings.HasPrefix(e, "directory publish:") {
			return true
		}
	}
	return false
}

// runDriftLeg runs one leg with the reference's raising-door belt:
// `except Exception as e: errors.append(f"{labels[key]}: {e}")` — a
// panicking leg (a coding bug) is caught and spoken, never fatal to
// the drift seal. The result flows through otherwise.
func runDriftLeg(fn func(any, LogFn) *SealResult, config any, log LogFn,
	label string, errors *[]string) (result *SealResult) {
	defer func() {
		if r := recover(); r != nil {
			*errors = append(*errors, label+": "+fmt.Sprint(r))
			result = nil
		}
	}()
	return fn(config, log)
}
