// seal.go — the seal itself: _commit_message (the processed/total
// door vs the snapshot), _seal (the belt → stage → skip → commit →
// push flow), and the never-raises wrapper.
package vaultseal

import (
	"fmt"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// commitMessage mirrors _commit_message: the run's shape (processed/
// total ints) rides the message; anything else is a snapshot.
func (s *VaultSeal) commitMessage(runSummary *RunSummary) string {
	ts := clockNow().UTC().Format("2006-01-02 15:04") + " UTC"
	if runSummary != nil && runSummary.Processed != nil && runSummary.Total != nil &&
		*runSummary.Total > 0 {
		return fmt.Sprintf("seal: %d note(s) curated (%d/%d repos) — %s",
			*runSummary.Processed, *runSummary.Processed, *runSummary.Total, ts)
	}
	return "seal: vault snapshot — " + ts
}

// sealInternal mirrors _seal: the engine room (enabled → vault →
// ensureRepo → the belt → add -A → the unchanged skip → the commit →
// the one-time-token push with the reconciliation ladder).
func (s *VaultSeal) sealInternal(runSummary *RunSummary) *SealResult {
	if !s.enabled {
		return &SealResult{Private: true, SkippedReason: "VaultSeal is disabled in config"}
	}
	if !s.vaultIsDir() {
		return &SealResult{Private: true, Error: "vault directory not found: " + s.vaultRepr()}
	}

	ok, err := s.ensureFn()
	if !ok {
		return &SealResult{Private: true, Error: err}
	}

	metaRepoName, metaRepoURL := sealMeta(s)

	// v0.89.0 — THE LOCAL-ONLY BELT applies BEFORE the stage: the
	// managed ignore block keeps the belt's notes out of `git add -A`
	// forever, and the one-time untrack brings what an OLDER seal
	// pushed back off the mirror (the files never leave the disk —
	// "kept local", the owner's law).
	s.ApplyLocalOnly()

	// Stage everything — pathspec-limited to the vault subtree so a
	// hypothetical parent repo can never swallow unrelated files (the
	// managed .gitignore keeps machine state out either way).
	s.gitf("add", "-A", "--", ".")
	rc, porcelain := s.gitf("status", "--porcelain")
	rcHead, _ := s.gitf("rev-parse", "--verify", "-q", "HEAD")
	var dirty []string
	if rc == 0 {
		for _, ln := range splitNonEmpty(porcelain) {
			dirty = append(dirty, ln)
		}
	}

	if len(dirty) == 0 && rcHead == 0 {
		return &SealResult{Private: true, SkippedReason: "vault unchanged since the last seal",
			RepoName: metaRepoName, RepoURL: metaRepoURL}
	}

	message := s.commitMessage(runSummary)
	rc, out := s.gitf("commit", "-m", message, "--quiet")
	if rc != 0 {
		return &SealResult{Private: true, Error: "git commit failed: " + out,
			RepoName: metaRepoName, RepoURL: metaRepoURL}
	}
	_, sha := s.gitf("rev-parse", "--short=7", "HEAD")

	result := &SealResult{
		Private:       true,
		Sealed:        true,
		CommitSha:     sha,
		CommitMessage: message,
		FilesChanged:  len(dirty),
		RepoName:      metaRepoName,
		RepoURL:       metaRepoURL,
	}

	if !s.autoPush {
		s.log("VaultSeal: committed locally — no GitHub token configured, "+
			"push skipped (set the token in Settings to enable pushing)", "warning")
		return result
	}

	// One-time token URL: never written to .git/config, never logged.
	pushURL := s.sealPushURL()
	pushed, pushErr := s.pushFn(pushURL)
	if !pushed {
		result.Error = "push failed (local commit kept): " + cutRunes(pushErr, 300)
		s.log("VaultSeal: "+result.Error, "warning")
	} else {
		result.Pushed = true
		// The reconciliation path may have rebased/merged — re-read the
		// tip so the log line names the sha actually on the mirror.
		_, sha = s.gitf("rev-parse", "--short=7", "HEAD")
		if sha != "" {
			result.CommitSha = sha
		}
		s.log(fmt.Sprintf("VaultSeal: sealed %d file(s) → %s (%s)",
			len(dirty), s.repoFull, sha), "success")
	}
	return result
}

// Seal mirrors seal(): commit + best-effort push of the whole vault.
// The reference's try/except wrapper is unreachable in Go (no code
// path panics); the honest result is the answer — a documented
// convergence (the T-46 never-raises law).
func (s *VaultSeal) Seal(runSummary *RunSummary) *SealResult {
	started := time.Now()
	result := s.sealInternal(runSummary)
	result.DurationMS = int(time.Since(started).Milliseconds())
	return result
}

// splitNonEmpty mirrors [ln for ln in out.splitlines() if ln.strip()].
func splitNonEmpty(out string) []string {
	var lines []string
	for _, ln := range helpers.SplitLines(out) {
		if strings.TrimSpace(ln) != "" {
			lines = append(lines, ln)
		}
	}
	return lines
}

// sealMeta is the **meta kwargs pair every SealResult carries
// (repo_name + repo_url; private is always true).
func sealMeta(s *VaultSeal) (string, string) {
	return s.repoFull, s.remoteURL
}
