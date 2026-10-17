// resync.go — v0.36.0 resync: the mirror follows the vault EXACTLY
// (write-tree delta counts, the commit-tree reconciliation with both
// parents, the empty-tree baseline, the one fetch-first retry, never
// a force — falls to the rescue branch).
package vaultseal

import (
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// isAncestor mirrors _is_ancestor.
func (s *VaultSeal) isAncestor(maybeAncestor, descendant string) bool {
	rc, _ := s.gitf("merge-base", "--is-ancestor", maybeAncestor, descendant)
	return rc == 0
}

// deltaCounts mirrors _delta_counts: (added, deleted, total) file
// counts between two tree-ishes. Rename detection is OFF
// (--no-renames): a deleted file plus an added file must count as one
// deletion AND one addition. An unreadable diff answers (0, 0, 0).
func (s *VaultSeal) deltaCounts(fromRef, toRef string) (int, int, int) {
	rc, out := s.gitf("diff", "--no-renames", "--name-status", fromRef, toRef)
	if rc != 0 {
		return 0, 0, 0
	}
	added, deleted, total := 0, 0, 0
	for _, ln := range helpers.SplitLines(out) {
		code := ln
		if i := strings.IndexByte(ln, '\t'); i >= 0 {
			code = ln[:i]
		}
		code = strings.TrimSpace(code)
		if code == "" {
			continue
		}
		total++
		if strings.HasPrefix(code, "D") {
			deleted++
		} else {
			added++
		}
	}
	return added, deleted, total
}

// resyncMessage mirrors _resync_message.
func (s *VaultSeal) resyncMessage(added, deleted int) string {
	ts := clockNow().UTC().Format("2006-01-02 15:04") + " UTC"
	return "resync: mirror follows the vault — " + itoa(deleted) +
		" file(s) deleted, " + itoa(added) + " added/updated (" + ts + ")"
}

// resyncPlain mirrors _resync_plain: the ordinary path (empty mirror,
// or the mirror already inside local history) — commit the staged
// state when dirty and push through the seal's own reconciliation.
// Deletions ride the commit.
func (s *VaultSeal) resyncPlain(pushURL, remoteHead, localHead string) *SealResult {
	rc, porcelain := s.gitf("status", "--porcelain")
	var dirty []string
	if rc == 0 {
		dirty = splitNonEmpty(porcelain)
	}
	if len(dirty) == 0 && localHead != "" && remoteHead == localHead {
		return &SealResult{Private: true, SkippedReason: "mirror already matches the vault",
			RepoName: s.repoFull, RepoURL: s.remoteURL}
	}
	msg := ""
	added, deleted := 0, 0
	if len(dirty) > 0 {
		baseline := remoteHead
		if baseline == "" {
			baseline = EmptyTree
		}
		_, treeOut := s.gitf("write-tree")
		tree := strings.TrimSpace(treeOut)
		if tree == "" {
			tree = baseline
		}
		added, deleted, _ = s.deltaCounts(baseline, tree)
		msg = s.resyncMessage(added, deleted)
		rc, out := s.gitf("commit", "-m", msg, "--quiet")
		if rc != 0 {
			return &SealResult{Private: true, Error: "git commit failed: " + cutRunes(out, 200),
				RepoName: s.repoFull, RepoURL: s.remoteURL}
		}
	}
	_, sha := s.gitf("rev-parse", "--short=7", "HEAD")
	pushed, pushErr := s.pushFn(pushURL)
	if !pushed {
		return &SealResult{
			Private:       true,
			Error:         "push failed (local commit kept): " + cutRunes(pushErr, 300),
			Sealed:        len(dirty) > 0,
			CommitSha:     sha,
			CommitMessage: msg,
			FilesChanged:  len(dirty),
			FilesAdded:    added,
			FilesDeleted:  deleted,
			RepoName:      s.repoFull,
			RepoURL:       s.remoteURL,
		}
	}
	_, sha = s.gitf("rev-parse", "--short=7", "HEAD")
	s.log("VaultSeal resync: the mirror now follows the vault — "+itoa(deleted)+
		" file(s) deleted, "+itoa(added)+" added/updated ("+sha+")", "success")
	commitMsg := msg
	if commitMsg == "" {
		commitMsg = "resync: mirror follows the vault"
	}
	return &SealResult{
		Private:       true,
		Sealed:        len(dirty) > 0,
		Pushed:        true,
		CommitSha:     sha,
		CommitMessage: commitMsg,
		FilesChanged:  len(dirty),
		FilesAdded:    added,
		FilesDeleted:  deleted,
		RepoName:      s.repoFull,
		RepoURL:       s.remoteURL,
	}
}

// Resync mirrors resync(push_url=None): the owner's recovery tool for
// "I deleted the wrongly-created notes in Obsidian, but the GitHub
// backup still shows them" — one owner-initiated action that commits
// the vault's CURRENT state (deletions included) and lands it on the
// mirror as a NEW commit on top of whatever the mirror has. The
// mirror's file tree ends up identical to the vault; its HISTORY is
// never rewritten. pushURL "" = None (the GitHub URL from the token);
// a non-empty pushURL is the hermetic test seam (a local bare repo).
// The reference's never-raises wrapper is unreachable in Go (the
// T-46 law).
func (s *VaultSeal) Resync(pushURL string) *SealResult {
	started := time.Now()
	result := s.resyncInternal(pushURL)
	result.DurationMS = int(time.Since(started).Milliseconds())
	return result
}

// resyncInternal mirrors _resync.
func (s *VaultSeal) resyncInternal(pushURL string) *SealResult {
	if !s.vaultIsDir() {
		return &SealResult{Private: true, Error: "vault directory not found: " + s.vaultRepr()}
	}
	if s.token == "" {
		return &SealResult{Private: true, Error: "the resync pushes to GitHub — no token " +
			"configured (set it in Settings → Credentials → GitHub Token)"}
	}

	if pushURL == "" {
		ok, err := s.ensureFn()
		if !ok {
			return &SealResult{Private: true, Error: err}
		}
		if s.repoFull == "" {
			return &SealResult{Private: true, Error: "could not resolve the mirror repository"}
		}
		pushURL = s.sealPushURL()
	} else {
		// Test seam / self-hosted mirror: skip the GitHub API, but a
		// commit still needs an identity.
		s.ensureIdentity()
	}

	// v0.89.0 — the belt applies here too: the resync is "the mirror
	// follows the vault EXACTLY", and the owner's vault keeps its
	// local-only notes — so the mirror must NOT carry them either.
	s.ApplyLocalOnly()

	// 1) Stage the CURRENT vault — deletions included.
	s.gitf("add", "-A", "--", ".")

	// 2) Bring the mirror's main in (an empty/unborn main just means
	//    "nothing to reconcile" — the plain path decides the rest).
	remoteHead := ""
	frc, _ := s.gitPush("fetch", pushURL, "refs/heads/main")
	s.sanitizeFetchHead()
	if frc == 0 {
		rc, out := s.gitf("rev-parse", "--verify", "-q", "FETCH_HEAD")
		if rc == 0 {
			remoteHead = strings.TrimSpace(out)
		}
	}
	rc, out := s.gitf("rev-parse", "--verify", "-q", "HEAD")
	localHead := ""
	if rc == 0 {
		localHead = strings.TrimSpace(out)
	}

	// 3) The ordinary path when the mirror is empty or already inside
	//    the local history (the push is then a plain fast-forward).
	if remoteHead == "" || (localHead != "" && s.isAncestor(remoteHead, localHead)) {
		return s.resyncPlain(pushURL, remoteHead, localHead)
	}

	// 4) The reconciliation path — the mirror is ahead or unrelated:
	//    one commit whose TREE is the vault exactly and whose parents
	//    keep both histories. write-tree reads the index we just
	//    staged; commit-tree + update-ref replace the merge machinery
	//    entirely (no conflicts, no resurrection of deleted files).
	rc, treeOut := s.gitf("write-tree")
	tree := strings.TrimSpace(treeOut)
	if rc != 0 || tree == "" {
		return &SealResult{Private: true, Error: "git write-tree failed: " + cutRunes(treeOut, 200),
			RepoName: s.repoFull, RepoURL: s.remoteURL}
	}

	sha := ""
	msg := ""
	added, deleted, total := 0, 0, 0
	pout := ""
	for attempt := 1; attempt <= 2; attempt++ {
		added, deleted, total = s.deltaCounts(remoteHead, tree)
		msg = s.resyncMessage(added, deleted)
		args := []string{"commit-tree", tree}
		if localHead != "" {
			args = append(args, "-p", localHead)
		}
		args = append(args, "-p", remoteHead, "-m", msg)
		rc, shaOut := s.gitf(args...)
		sha = strings.TrimSpace(shaOut)
		if rc != 0 || sha == "" {
			return &SealResult{Private: true, Error: "git commit-tree failed: " + cutRunes(shaOut, 200),
				RepoName: s.repoFull, RepoURL: s.remoteURL}
		}
		rc, out = s.gitf("update-ref", "HEAD", sha)
		if rc != 0 {
			return &SealResult{Private: true, Error: "git update-ref failed: " + cutRunes(out, 200),
				RepoName: s.repoFull, RepoURL: s.remoteURL}
		}
		rc, pout = s.gitPush("push", pushURL, "HEAD:refs/heads/main")
		if rc == 0 {
			s.log("VaultSeal resync: the mirror now follows the vault — "+itoa(deleted)+
				" file(s) deleted, "+itoa(added)+" added/updated ("+cutRunes(sha, 7)+")", "success")
			return &SealResult{
				Private:       true,
				Sealed:        true,
				Pushed:        true,
				CommitSha:     cutRunes(sha, 7),
				CommitMessage: msg,
				FilesChanged:  total,
				FilesAdded:    added,
				FilesDeleted:  deleted,
				RepoName:      s.repoFull,
				RepoURL:       s.remoteURL,
			}
		}
		if strings.Contains(pout, "fetch first") ||
			strings.Contains(strings.ToLower(pout), "non-fast-forward") {
			// The mirror moved between our fetch and the push — one
			// retry on the fresh tip (never a force-push).
			frc2, _ := s.gitPush("fetch", pushURL, "refs/heads/main")
			s.sanitizeFetchHead()
			if frc2 == 0 {
				rc2, out2 := s.gitf("rev-parse", "--verify", "-q", "FETCH_HEAD")
				fresh := ""
				if rc2 == 0 {
					fresh = strings.TrimSpace(out2)
				}
				if fresh != "" && fresh != remoteHead {
					remoteHead = fresh
					continue
				}
			}
		}
		break
	}

	// Never force-push — the seal's rescue-branch safety net instead
	// (the reconciliation commit stays local, named in the error).
	_, err := s.rescue(pushURL, cutRunes(pout, 300))
	return &SealResult{
		Private:       true,
		Error:         "push rejected (reconciliation commit kept locally): " + cutRunes(err, 300),
		Sealed:        true,
		Pushed:        false,
		CommitSha:     cutRunes(sha, 7),
		CommitMessage: msg,
		FilesChanged:  total,
		FilesAdded:    added,
		FilesDeleted:  deleted,
		RepoName:      s.repoFull,
		RepoURL:       s.remoteURL,
	}
}

// itoa is the local int renderer.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
