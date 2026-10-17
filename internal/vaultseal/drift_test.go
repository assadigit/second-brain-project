// drift_test.go — T-47: the ported selections of test_driftseal.py
// §1-§2 (the drift seal's pure half): the plan's legs mirror the
// batch-end guards, the defaults answer nothing, each leg gates
// exactly, a broken config is silence — and the runner's laws: THE
// QUIET LAW (the clean vault is quiet), a sealed leg speaks once
// with its file count, all three legs speak, the publish rides only
// the websites leg, errors speak as warnings without masking the
// seals, a raising leg is caught, a disabled plan publishes nothing.
package vaultseal

import (
	"strings"
	"testing"
)

// sealStub mirrors _SealStub.
type sealStub struct {
	sealed  bool
	files   int
	err     string
	skipped string
}

func (s *sealStub) result() *SealResult {
	r := NewSealResult()
	r.Sealed = s.sealed
	r.FilesChanged = s.files
	r.Error = s.err
	r.SkippedReason = s.skipped
	return r
}

func TestDriftSealPlan_TheThreeLegsMirrorTheBatchEndGuards(t *testing.T) {
	cfg := map[string]any{
		"vault_path": "/v", "vaultseal": map[string]any{"enabled": true},
		"pipelines":          map[string]any{"websites": true},
		"website_vault_path": "/w",
		"manual_vault_path":  "/p", "manual_repo_name": "pr",
	}
	got := DriftSealPlan(cfg)
	want := []string{"repos", "websites", "personal"}
	if len(got) != len(want) || got[0] != want[0] || got[1] != want[1] || got[2] != want[2] {
		t.Fatalf("plan: %v", got)
	}
}

func TestDriftSealPlan_TheDefaultsAnswerNothing(t *testing.T) {
	if got := DriftSealPlan(map[string]any{}); len(got) != 0 {
		t.Fatalf("empty config answered %v", got)
	}
	if got := DriftSealPlan(nil); len(got) != 0 {
		t.Fatalf("nil config answered %v", got)
	}
}

func TestDriftSealPlan_EachLegGatesExactlyLikeTheBatchEnd(t *testing.T) {
	// vaultseal disabled → no repos leg
	if inList(DriftSealPlan(map[string]any{
		"vault_path": "/v", "vaultseal": map[string]any{"enabled": false},
	}), "repos") {
		t.Fatal("the repos leg answered while vaultseal is disabled")
	}
	// websites pipeline off → no websites leg (the Phase 1 law)
	if inList(DriftSealPlan(map[string]any{
		"pipelines":          map[string]any{"websites": false},
		"website_vault_path": "/w",
	}), "websites") {
		t.Fatal("the websites leg answered while the pipeline is off")
	}
	// personal needs BOTH the vault and the repo (the v0.83.0 guard)
	if inList(DriftSealPlan(map[string]any{"manual_vault_path": "/p"}), "personal") {
		t.Fatal("the personal leg answered without a repo")
	}
	if inList(DriftSealPlan(map[string]any{"manual_repo_name": "pr"}), "personal") {
		t.Fatal("the personal leg answered without a vault")
	}
}

func TestDriftSealPlan_ABrokenConfigIsSilenceNeverACrash(t *testing.T) {
	// the reference's _Boom dict raises on .get → []; the Go guards
	// answer nil doors for junk shapes — the same silence.
	if got := DriftSealPlan(map[string]any{
		"vault_path": map[string]any{"junk": true},
		"vaultseal":  "not-a-dict",
		"pipelines":  []any{1, 2},
	}); len(got) != 0 {
		t.Fatalf("a broken config answered %v", got)
	}
}

// driftDeps mirrors _patch_legs: the three legs + the publish door.
func driftDeps(repos, websites, personal *sealStub, publish bool) (*DriftSealDeps, *int) {
	calls := new(int)
	stub := func(s *sealStub) func(any, LogFn) *SealResult {
		return func(any, LogFn) *SealResult { return s.result() }
	}
	skip := &sealStub{skipped: "unchanged"}
	if repos == nil {
		repos = skip
	}
	if websites == nil {
		websites = skip
	}
	if personal == nil {
		personal = skip
	}
	return &DriftSealDeps{
		SealRepos:         stub(repos),
		SealWebsites:      stub(websites),
		SealPersonal:      stub(personal),
		PublishAfterBatch: func(any, LogFn) bool { *calls++; return publish },
	}, calls
}

func TestRunDriftSeal_TheCleanVaultIsQuiet(t *testing.T) {
	// THE QUIET LAW: every leg skipped-unchanged, nothing sealed —
	// the empty state's honest title survives
	deps, _ := driftDeps(nil, nil, nil, false)
	line, level := RunDriftSeal(map[string]any{"vault_path": "/v"}, deps, nil, false)
	if line != "" || level != "info" {
		t.Fatalf("the clean vault was not quiet: %q / %q", line, level)
	}
}

func TestRunDriftSeal_ASealedLegSpeaksOnceWithItsFileCount(t *testing.T) {
	deps, _ := driftDeps(&sealStub{sealed: true, files: 69}, nil, nil, false)
	line, level := RunDriftSeal(map[string]any{"vault_path": "/v"}, deps, nil, false)
	if !strings.Contains(line, "69 file(s) → the repos vault mirror") {
		t.Fatalf("line: %q", line)
	}
	if level != "success" {
		t.Fatalf("level: %q", level)
	}
}

func TestRunDriftSeal_AllThreeLegsSpeakAndTheCountIsTheOwnersOwn(t *testing.T) {
	deps, _ := driftDeps(
		&sealStub{sealed: true, files: 69},
		&sealStub{sealed: true, files: 730},
		&sealStub{sealed: true, files: 32}, false)
	line, level := RunDriftSeal(map[string]any{
		"vault_path": "/v", "pipelines": map[string]any{"websites": true},
		"website_vault_path": "/w",
		"manual_vault_path":  "/p", "manual_repo_name": "r",
	}, deps, nil, false)
	for _, want := range []string{"69", "730", "32"} {
		if !strings.Contains(line, want) {
			t.Fatalf("line misses %s: %q", want, line)
		}
	}
	if level != "success" {
		t.Fatalf("level: %q", level)
	}
}

func TestRunDriftSeal_ThePublishRidesOnlyTheWebsitesLeg(t *testing.T) {
	// websites sealed → the directory publish runs
	deps, calls := driftDeps(nil, &sealStub{sealed: true, files: 5}, nil, true)
	line, _ := RunDriftSeal(map[string]any{
		"pipelines": map[string]any{"websites": true}, "website_vault_path": "/w",
	}, deps, nil, false)
	if !strings.Contains(line, "directory published") {
		t.Fatalf("line: %q", line)
	}
	if *calls != 1 {
		t.Fatalf("publish called %d times", *calls)
	}
	// repos-only seal → the publish NEVER runs (the site's index
	// derives from the websites vault alone)
	deps2, calls2 := driftDeps(&sealStub{sealed: true, files: 9}, nil, nil, true)
	line2, _ := RunDriftSeal(map[string]any{"vault_path": "/v"}, deps2, nil, false)
	if strings.Contains(line2, "directory published") {
		t.Fatalf("line: %q", line2)
	}
	if *calls2 != 0 {
		t.Fatalf("the publish ran for a repos-only seal (%d)", *calls2)
	}
}

func TestRunDriftSeal_ErrorsSpeakAsWarningsAndNeverMaskTheSeals(t *testing.T) {
	deps, _ := driftDeps(&sealStub{sealed: true, files: 3},
		&sealStub{err: "push failed"}, nil, false)
	line, level := RunDriftSeal(map[string]any{
		"vault_path": "/v", "pipelines": map[string]any{"websites": true},
		"website_vault_path": "/w",
	}, deps, nil, false)
	if level != "warning" {
		t.Fatalf("level: %q", level)
	}
	if !strings.Contains(line, "3 file(s)") {
		t.Fatalf("the seal count was masked: %q", line)
	}
	if !strings.Contains(line, "push failed") {
		t.Fatalf("the error was masked: %q", line)
	}
}

func TestRunDriftSeal_OnlyErrorsIsAWarningLine(t *testing.T) {
	deps, _ := driftDeps(nil, nil, &sealStub{err: "no token"}, false)
	line, level := RunDriftSeal(map[string]any{
		"manual_vault_path": "/p", "manual_repo_name": "r",
	}, deps, nil, false)
	if level != "warning" {
		t.Fatalf("level: %q", level)
	}
	if !strings.Contains(line, "personal vault") {
		t.Fatalf("line: %q", line)
	}
	if !strings.Contains(line, "no token") {
		t.Fatalf("line: %q", line)
	}
}

func TestRunDriftSeal_ARaisingLegIsCaughtNotFatal(t *testing.T) {
	deps, _ := driftDeps(nil, nil, nil, false)
	deps.SealRepos = func(any, LogFn) *SealResult { panic("git exploded") }
	line, level := RunDriftSeal(map[string]any{"vault_path": "/v"}, deps, nil, false)
	if level != "warning" {
		t.Fatalf("level: %q", level)
	}
	if !strings.Contains(line, "git exploded") {
		t.Fatalf("line: %q", line)
	}
}

func TestRunDriftSeal_ADisabledPlanPublishesNothing(t *testing.T) {
	deps, calls := driftDeps(nil, nil, nil, true)
	line, level := RunDriftSeal(map[string]any{}, deps, nil, false)
	if line != "" || level != "info" {
		t.Fatalf("the disabled plan spoke: %q / %q", line, level)
	}
	if *calls != 0 {
		t.Fatalf("the publish ran for a disabled plan (%d)", *calls)
	}
}

func TestRunDriftSeal_TheManualDoorAlwaysSpeaks(t *testing.T) {
	// v0.90.0 — the manual door's own law: an explicit press ALWAYS
	// gets its verdict; the publish ALWAYS runs when the websites leg
	// is enabled (a clean vault with a re-landing publish says the
	// whole triad in one line).
	deps, calls := driftDeps(nil, nil, nil, true)
	line, level := RunDriftSeal(map[string]any{
		"pipelines": map[string]any{"websites": true}, "website_vault_path": "/w",
	}, deps, nil, true)
	if level != "success" {
		t.Fatalf("level: %q (%s)", level, line)
	}
	if !strings.Contains(line, "the mirrors were already current") {
		t.Fatalf("line: %q", line)
	}
	if *calls != 1 {
		t.Fatalf("the manual door did not ask for the publish (%d)", *calls)
	}
	// the clean triad with NO websites leg: the ✅ already-in-sync
	// verdict, no publish attempt
	deps2, calls2 := driftDeps(nil, nil, nil, true)
	line2, level2 := RunDriftSeal(map[string]any{
		"manual_vault_path": "/p", "manual_repo_name": "r",
	}, deps2, nil, true)
	if level2 != "success" || !strings.Contains(line2, "already in sync") {
		t.Fatalf("line: %q / %q", line2, level2)
	}
	if *calls2 != 0 {
		t.Fatalf("a publish ran without the websites leg (%d)", *calls2)
	}
	// a publish that did NOT land speaks its own warning (the manual
	// door's honesty: the desk log carries the Worker's reason)
	deps3, _ := driftDeps(nil, nil, nil, false)
	line3, level3 := RunDriftSeal(map[string]any{
		"pipelines": map[string]any{"websites": true}, "website_vault_path": "/w",
	}, deps3, nil, true)
	if level3 != "warning" {
		t.Fatalf("level: %q (%s)", level3, line3)
	}
	if !strings.Contains(line3, "the directory publish did not land") {
		t.Fatalf("line: %q", line3)
	}
}
