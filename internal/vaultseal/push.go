// push.go — v0.21.0 THE DIVERGENCE LADDER (the _push family): plain
// push → fetch + FETCH_HEAD token redaction → stale-rejection retry /
// rebase / unrelated-histories merge / _rescue branch — main is NEVER
// force-pushed.
package vaultseal

import (
	"os"
	"path/filepath"
	"strings"
)

// sanitizeFetchHead mirrors _sanitize_fetch_head: strip the access
// token out of .git/FETCH_HEAD after a fetch. `git fetch <token-url>`
// writes the full URL (token included) into FETCH_HEAD; the design
// rule is that the token is never persisted anywhere — best effort,
// failure tolerated (next fetch overwrites the file anyway).
func (s *VaultSeal) sanitizeFetchHead() {
	if s.vault == "" || s.token == "" {
		return
	}
	fh := filepath.Join(s.vault, ".git", "FETCH_HEAD")
	data, err := os.ReadFile(fh)
	if err != nil {
		return
	}
	text := string(data)
	if !strings.Contains(text, s.token) {
		return
	}
	_ = os.WriteFile(fh, []byte(strings.ReplaceAll(text, s.token, "***")), 0o644)
}

// rescue mirrors _rescue: last resort after an irreconcilable mirror
// — push the seal to a seal-rescue/<timestamp> branch. The local
// commit is never lost, the remote main is never force-pushed, and
// the error names the branch so the owner can reconcile at their
// leisure. Always answers ok=false (the seal did NOT land on main).
func (s *VaultSeal) rescue(pushURL, originalErr string) (bool, string) {
	branch := "seal-rescue/" + clockNow().Format("20060102-150405")
	rc, out := s.gitPush("push", pushURL, "HEAD:refs/heads/"+branch)
	if rc == 0 {
		return false, "mirror main rejected the seal and could not be reconciled " +
			"automatically — the commit was pushed to branch '" + branch + "' " +
			"instead (local commit kept). Reconcile on GitHub: open the mirror " +
			"repo → Compare & pull request '" + branch + "' → main, or clone the " +
			"mirror and merge locally. (" + cutRunes(originalErr, 150) + ")"
	}
	return false, originalErr + " (rescue push to '" + branch + "' also failed: " +
		cutRunes(out, 150) + ")"
}

// cutRunes is the Python str[:n] door (code-point slicing).
func cutRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	count := 0
	for i := range s {
		if count == n {
			return s[:i]
		}
		count++
	}
	return s
}

// push mirrors _push: push HEAD:refs/heads/main to the mirror (the
// v0.21.0 divergence ladder).
//
// Plain push first. When GitHub rejects it because the remote has
// commits we lack (a web-UI edit on the mirror, a second machine
// sealing, or the mirror repo having been used for something else
// mid-flight), the seal RECONCILES instead of failing forever:
//
//	fetch → remote already inside HEAD → retry the plain push
//	fetch → shared history → rebase our commit(s) onto it → push
//	fetch → unrelated histories (fresh git init / repurposed mirror)
//	         → merge --allow-unrelated-histories → push
//
// A reconciliation that cannot complete cleanly (conflicts) never
// destroys anything: the seal lands on a rescue branch and the error
// says exactly where. main is NEVER force-pushed — a backup tool must
// never discard remote history it cannot see. Returns (ok, error
// detail). Never errors.
func (s *VaultSeal) push(pushURL string) (bool, string) {
	rc, out := s.gitPush("push", pushURL, "HEAD:refs/heads/main")
	if rc == 0 {
		return true, ""
	}
	if !strings.Contains(out, "fetch first") &&
		!strings.Contains(strings.ToLower(out), "non-fast-forward") {
		// Auth / network / permission failure — reconciliation does not
		// apply; report the original error verbatim.
		return false, out
	}

	// The remote is ahead: bring its main in (FETCH_HEAD) and reconcile.
	frc, fout := s.gitPush("fetch", pushURL, "refs/heads/main")
	s.sanitizeFetchHead()
	if frc != 0 {
		return false, out + " — and fetching the mirror to reconcile failed: " +
			cutRunes(fout, 150)
	}

	// Remote already contained in HEAD (stale rejection) → plain retry.
	rcAnc, _ := s.gitf("merge-base", "--is-ancestor", "FETCH_HEAD", "HEAD")
	if rcAnc == 0 {
		rc2, out2 := s.gitPush("push", pushURL, "HEAD:refs/heads/main")
		return rc2 == 0, out2
	}

	rcBase, _ := s.gitf("merge-base", "HEAD", "FETCH_HEAD")
	if rcBase == 0 {
		// Shared history: replay our seal commit(s) on top of the remote.
		rrc, _ := s.gitf("rebase", "FETCH_HEAD")
		if rrc == 0 {
			rc3, out3 := s.gitPush("push", pushURL, "HEAD:refs/heads/main")
			if rc3 == 0 {
				return true, ""
			}
			return s.rescue(pushURL, out3)
		}
		s.gitf("rebase", "--abort")
		return s.rescue(pushURL, out)
	}

	// Unrelated histories: keep BOTH (merge) when the trees do not
	// collide; a collision (same path, different content) → rescue.
	mrc, _ := s.gitf("merge", "FETCH_HEAD", "--allow-unrelated-histories", "-m",
		"seal: reconcile mirror histories (v0.21.0)")
	if mrc == 0 {
		rc4, out4 := s.gitPush("push", pushURL, "HEAD:refs/heads/main")
		if rc4 == 0 {
			return true, ""
		}
		return s.rescue(pushURL, out4)
	}
	s.gitf("merge", "--abort")
	return s.rescue(pushURL, out)
}

// sealPushURL builds the one-time token URL — never written to
// .git/config, never logged.
func (s *VaultSeal) sealPushURL() string {
	return "https://x-access-token:" + s.token + "@github.com/" + s.repoFull + ".git"
}
