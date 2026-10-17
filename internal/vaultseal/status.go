// status.go — status(): the pure-introspection snapshot for UIs (repo
// state, dirty files, last commit) — NO network.
package vaultseal

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// Status mirrors status(): a snapshot dict for UIs. The Python None
// doors (dirty_count before the first rc==0, last_commit, remote_url,
// repo_name when unresolved) re-open as nil here — the dict keeps
// Python's insertion order for the CLI's json.dumps parity.
func (s *VaultSeal) Status() *storage.OrderedMap {
	info := storage.NewOrderedMap()
	info.Set("vault_path", pyNoneIfEmpty(s.vault))
	info.Set("exists", s.vaultIsDir())
	info.Set("is_git_repo", false)
	info.Set("has_commits", false)
	info.Set("dirty_count", nil)
	info.Set("last_commit", nil)
	info.Set("remote_url", nil)
	repoName := s.repoFull
	if repoName == "" {
		repoName = s.repoName
	}
	info.Set("repo_name", pyNoneIfEmpty(repoName))

	if !s.vaultIsDir() {
		return info
	}
	rc, out := s.gitf("rev-parse", "--show-toplevel")
	isGitRepo := rc == 0 && strings.TrimSpace(out) != "" &&
		pyRealpath(strings.TrimSpace(out)) == pyRealpath(s.vault)
	info.Set("is_git_repo", isGitRepo)
	if !isGitRepo {
		return info
	}
	rc, _ = s.gitf("rev-parse", "--verify", "-q", "HEAD")
	hasCommits := rc == 0
	info.Set("has_commits", hasCommits)
	rc, out = s.gitf("status", "--porcelain")
	if rc == 0 {
		info.Set("dirty_count", len(splitNonEmpty(out)))
	}
	if hasCommits {
		rc, out = s.gitf("log", "-1", "--format=%h%x00%s%x00%aI")
		if rc == 0 && out != "" {
			parts := strings.Split(out, "\x00")
			if len(parts) < 3 {
				// (out.split(\"\\x00\") + [\"\", \"\", \"\"])[:3] — pad short tails
				padded := make([]string, 3)
				copy(padded, parts)
				parts = padded
			}
			last := storage.NewOrderedMap()
			last.Set("sha", parts[0])
			last.Set("subject", parts[1])
			last.Set("date", parts[2])
			info.Set("last_commit", last)
		}
	}
	rc, out = s.gitf("remote", "get-url", "origin")
	if rc == 0 {
		info.Set("remote_url", out)
	}
	return info
}
