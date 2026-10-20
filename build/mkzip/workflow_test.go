package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// readWorkflow — the committed T-81 release workflow (the automation
// this file pins; the dry-run evidence lives in
// docs/poc-artifacts/t-81-release-workflow.md).
func readWorkflow(t *testing.T) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(repoRoot, ".github", "workflows",
		"release.yml"))
	if err != nil {
		t.Fatalf("read the release workflow: %v", err)
	}
	return string(data)
}

// TestReleaseWorkflowTrigger — the release door is a TAG PUSH ONLY
// (doc 10 §8: "tag push → build both exes → mkzip → attach"): the v*
// tag filter, no branch trigger, no pull_request door — a release never
// fires on a main push.
func TestReleaseWorkflowTrigger(t *testing.T) {
	w := readWorkflow(t)
	if !strings.Contains(w, "tags: ['v*']") {
		t.Error("the release trigger lost the tags: ['v*'] filter")
	}
	if strings.Contains(w, "branches:") {
		t.Error("the release workflow must not trigger on branch pushes " +
			"(a release fires on TAG PUSHES ONLY — remove the branches: filter)")
	}
	if strings.Contains(w, "pull_request") {
		t.Error("the release workflow must not fire on pull requests")
	}
}

// TestReleaseWorkflowToolchainPins — the runner builds with the COMMITTED
// toolchain: go-version is TOOLCHAIN.md Am.2's pin (go1.27.2 — the
// sha256 evidence numbers are its numbers) and the wails CLI is
// ADR-0001 Am.2's pin (v2.16.0); both pins cross-checked against
// build/TOOLCHAIN.md (the one-toolchain-truth).
func TestReleaseWorkflowToolchainPins(t *testing.T) {
	w := readWorkflow(t)
	for _, want := range []string{
		"go-version: '1.27.2'",
		"go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0",
	} {
		if !strings.Contains(w, want) {
			t.Errorf("the release workflow lost the toolchain pin: %s", want)
		}
	}
	tc, err := os.ReadFile(filepath.Join(repoRoot, "build", "TOOLCHAIN.md"))
	if err != nil {
		t.Fatalf("read build/TOOLCHAIN.md: %v", err)
	}
	for _, pin := range []string{"1.27.2", "v2.16.0"} {
		if !strings.Contains(string(tc), pin) {
			t.Errorf("build/TOOLCHAIN.md does not carry the %s pin — the "+
				"workflow and the committed recipe drifted", pin)
		}
	}
}

// TestReleaseWorkflowRitualWrap — THE SINGLE-OWNER law: the workflow
// wraps build/release.sh (T-80's double-build ritual) and never
// re-implements a build command itself; the ritual's legs are CGO-free
// windows cross-compiles, so the runner is plain ubuntu (no webkit, no
// windows runner).
func TestReleaseWorkflowRitualWrap(t *testing.T) {
	w := readWorkflow(t)
	if !strings.Contains(w, "runs-on: ubuntu-latest") {
		t.Error("the release workflow left ubuntu-latest — the ritual's " +
			"windows legs are CGO-free cross-compiles; no GUI runner is needed")
	}
	if !strings.Contains(w, "./build/release.sh") {
		t.Error("the release workflow must run ./build/release.sh (the " +
			"reusable core — the ritual owns the build commands)")
	}
	if strings.Contains(w, "go build") || strings.Contains(w, "wails build") {
		t.Error("the release workflow re-implements a build command — " +
			"build/release.sh is the single owner of the build legs")
	}
}

// TestReleaseWorkflowTagLaw — THE ONE-VERSION-TRUTH FOR TAGS: a real
// release tag must be exactly v<VERSION> (the repo VERSION file — the
// tag IS the version it ships); the reserved dry-run prefix v0.0.0-
// (the T-81 acceptance leg: a TEST TAG, never a real version bump)
// rides the same automation with the assertion waived.
func TestReleaseWorkflowTagLaw(t *testing.T) {
	w := readWorkflow(t)
	if !strings.Contains(w, `elif [ "$TAG" != "v$VERSION" ]; then`) {
		t.Error("the tag law lost the real-tag assertion " +
			`elif [ "$TAG" != "v$VERSION" ]; then`)
	}
	if !strings.Contains(w, `if [[ "$TAG" == v0.0.0-* ]]; then`) {
		t.Error("the tag law lost the reserved dry-run exemption v0.0.0-* " +
			"(the T-81 test-tag shape)")
	}
}

// TestReleaseWorkflowAssetLaw — THE OWNER'S LAW: the release carries the
// zip under the exact name GitCurator-v<version>-windows.zip, staged
// from the ritual's .a artifact (the ritual proved A == B before
// anything attached) with the AC-23 notes (the parity table's sha256
// record).
func TestReleaseWorkflowAssetLaw(t *testing.T) {
	w := readWorkflow(t)
	for _, want := range []string{
		`ASSET="GitCurator-v${VERSION}-windows.zip"`,
		`cp "reports/dist/${ASSET}.a" "$ASSET"`,
		`--notes-file notes.md "$ASSET"`,
		"| Artifact | sha256 |",
	} {
		if !strings.Contains(w, want) {
			t.Errorf("the asset law lost: %s", want)
		}
	}
}

// TestReleaseWorkflowPermissionsAndDryRun — the release+asset doors
// (contents: write) + the dry-run shape: a draft + prerelease titled
// [DRY-RUN] (a test release, deleted with its tag after the proof —
// formal releases happen only at roadmap milestones).
func TestReleaseWorkflowPermissionsAndDryRun(t *testing.T) {
	w := readWorkflow(t)
	for _, want := range []string{
		"contents: write",
		"--draft --prerelease",
		`TITLE="[DRY-RUN] ${TAG}"`,
	} {
		if !strings.Contains(w, want) {
			t.Errorf("the workflow lost: %s", want)
		}
	}
}
