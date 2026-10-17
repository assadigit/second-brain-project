// bootstrap.go — the repository bootstrap family: the managed
// .gitignore block (ensureGitignore), the LOCAL-ONLY belt's second
// block + the untrack law (applyLocalOnly / ensureLocalOnlyGitignore
// / untrackLocalOnly), the local-only identity, and ensureRepo (the
// nested-repo guard → git init -b main, the token-less remote, the
// v0.83 creation guard, the 422 race re-check).
package vaultseal

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// pyRealpath mirrors os.path.realpath (non-strict): resolve as much
// as exists; the unresolvable tail stays literal. Python never errors
// here, so neither do we.
func pyRealpath(p string) string {
	if resolved, err := filepath.EvalSymlinks(p); err == nil {
		return resolved
	}
	// resolve the longest existing prefix, keep the tail literal
	clean := filepath.Clean(p)
	dir, tail := filepath.Split(clean)
	if resolved, err := filepath.EvalSymlinks(dir); err == nil {
		return filepath.Join(resolved, tail)
	}
	return clean
}

// readGitignore reads the vault's .gitignore as text (an unreadable
// file answers "" — the except OSError door). Python's read_text
// decodes UTF-8 AND translates universal newlines (\r\n → \n — the
// T-41 law, caught live by the golden's CRLF tail case); a decode
// error would crash the reference's seal through its never-raises
// wrapper — the house law: decode with the maximal-subpart U+FFFD (a
// documented convergence).
func readGitignore(vault string) string {
	data, err := os.ReadFile(filepath.Join(vault, ".gitignore"))
	if err != nil {
		return ""
	}
	return helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
}

// writeGitignoreAtomic mirrors the .gitignore.tmp + os.replace dance
// (atomic — readers see old or new, never half).
func writeGitignoreAtomic(vault, text string, log LogFn, warnPrefix string) bool {
	tmp := filepath.Join(vault, ".gitignore.tmp")
	if err := os.WriteFile(tmp, []byte(text), 0o644); err != nil {
		log(warnPrefix+": "+helpers.PyOSErrorText(err, tmp), "warning")
		return false
	}
	if err := os.Rename(tmp, filepath.Join(vault, ".gitignore")); err != nil {
		log(warnPrefix+": "+helpers.PyOSErrorText(err, tmp), "warning")
		return false
	}
	return true
}

// EnsureGitignore mirrors ensure_gitignore: merge the managed ignore
// block into the vault's .gitignore. Idempotent — returns the lines
// actually added (nil when clean).
func (s *VaultSeal) EnsureGitignore() []string {
	if s.vault == "" {
		return nil
	}
	current := readGitignore(s.vault)
	lines := splitGitignoreLines(current)
	have := map[string]bool{}
	for _, ln := range lines {
		have[ln] = true
	}
	var added []string
	for _, ln := range DefaultIgnoreLines {
		if !have[ln] {
			added = append(added, ln)
		}
	}
	if len(added) == 0 {
		return nil
	}
	block := strings.Join(append([]string{GitignoreHeader}, added...), "\n")
	newText := block + "\n"
	if strings.TrimSpace(current) != "" {
		newText = helpers.RstripChars(current, "\n") + "\n\n" + block + "\n"
	}
	writeGitignoreAtomic(s.vault, newText, s.log, "VaultSeal: could not update .gitignore")
	return added
}

// splitGitignoreLines mirrors [ln.rstrip("\r\n") for ln in
// current.splitlines()] — helpers.SplitLines is Python's full
// boundary set; the rstrip is belt-and-braces exactly as the
// reference has it.
func splitGitignoreLines(current string) []string {
	lines := helpers.SplitLines(current)
	for i, ln := range lines {
		lines[i] = helpers.RstripChars(ln, "\r\n")
	}
	return lines
}

// ApplyLocalOnly mirrors _apply_local_only: one door for the whole
// belt — merge the managed ignore block, then untrack whatever an
// older seal already pushed. Returns the number of paths untracked
// (0, no block, for a vault with no local-only set).
func (s *VaultSeal) ApplyLocalOnly() int {
	if len(s.localOnly) == 0 {
		return 0
	}
	s.EnsureLocalOnlyGitignore()
	return s.untrackLocalOnly()
}

// EnsureLocalOnlyGitignore mirrors ensure_local_only_gitignore: merge
// THE LOCAL-ONLY block (its own header, the belt's patterns) into the
// vault's .gitignore. Idempotent — returns the FULL current list (the
// reference returns list(self.local_only)) — and a pattern the OWNER
// removes from their belt is also removed here (the block is REBUILT
// from the current list, never appended to twice).
func (s *VaultSeal) EnsureLocalOnlyGitignore() []string {
	if s.vault == "" || len(s.localOnly) == 0 {
		return nil
	}
	current := readGitignore(s.vault)
	lines := splitGitignoreLines(current)
	// rebuild the managed block in place (keep everything else): the
	// block starts at its header and ends at the first BLANK line,
	// the other VaultSeal header, or EOF — the patterns never contain
	// blanks, so anything past a blank is the owner's own.
	kept := make([]string, 0, len(lines)+len(s.localOnly)+2)
	inBlock := false
	for _, ln := range lines {
		if strings.TrimSpace(ln) == LocalOnlyHeader {
			inBlock = true
			continue
		}
		if inBlock {
			if strings.TrimSpace(ln) == "" || strings.TrimSpace(ln) == GitignoreHeader {
				inBlock = false
				kept = append(kept, ln) // the owner's blank / the other block
				continue
			}
			continue // an old pattern line — replaced
		}
		kept = append(kept, ln)
	}
	// trim the trailing blanks the removal may leave
	for len(kept) > 0 && strings.TrimSpace(kept[len(kept)-1]) == "" {
		kept = kept[:len(kept)-1]
	}
	block := strings.Join(append([]string{LocalOnlyHeader}, s.localOnly...), "\n")
	base := strings.Join(kept, "\n")
	newText := block + "\n"
	if strings.TrimSpace(base) != "" {
		newText = base + "\n\n" + block + "\n"
	}
	writeGitignoreAtomic(s.vault, newText, s.log,
		"VaultSeal: could not update .gitignore (local-only)")
	return append([]string(nil), s.localOnly...)
}

// untrackLocalOnly mirrors _untrack_local_only: untrack the
// local-only notes an OLDER seal may have already pushed — `git rm
// --cached` (NEVER a disk delete: the files stay exactly where they
// are; only the index forgets them, so the next commit records their
// removal from the mirror). Returns the number of paths untracked
// (0 = clean).
func (s *VaultSeal) untrackLocalOnly() int {
	if s.vault == "" || len(s.localOnly) == 0 {
		return 0
	}
	rc, out := s.gitf("ls-files")
	if rc != 0 || strings.TrimSpace(out) == "" {
		return 0
	}
	patterns := make([]string, 0, len(s.localOnly))
	for _, p := range s.localOnly {
		patterns = append(patterns, strings.ReplaceAll(p, "\\", "/"))
	}
	var victims []string
	for _, path := range helpers.SplitLines(out) {
		p := strings.ReplaceAll(
			strings.Trim(strings.TrimSpace(path), `"`), "\\", "/")
		if p == "" {
			continue
		}
		folded := helpers.PyLower(p)
		for _, pat := range patterns {
			patL := helpers.PyLower(pat)
			matched := false
			if strings.HasSuffix(patL, "/") {
				// a folder pattern 'A/B/' matches any path under A/B
				matched = strings.HasPrefix(folded, patL) || FnMatch(folded, "*"+patL+"*")
			} else {
				// a stem pattern 'Cobalt*.md' matches any basename
				base := folded
				if i := strings.LastIndex(folded, "/"); i >= 0 {
					base = folded[i+1:]
				}
				matched = FnMatch(base, patL) || FnMatch(folded, "*/"+patL) || folded == patL
			}
			if matched {
				victims = append(victims, p)
				break
			}
		}
	}
	if len(victims) == 0 {
		return 0
	}
	args := append([]string{"rm", "-r", "--cached", "--ignore-unmatch", "--"}, victims...)
	rc, out = s.git(args, 0)
	if rc == 0 {
		s.log(strconv.Itoa(len(victims))+" local-only note(s) untracked — kept on "+
			"this machine, removed from the mirror on this seal", "info")
	} else {
		s.log("VaultSeal: could not untrack the local-only notes ("+
			helpers.CutRunes(out, 150)+")", "warning")
	}
	if rc != 0 {
		return 0
	}
	return len(victims)
}

// ensureIdentity mirrors _ensure_identity: local git identity — only
// when unset, never the global config.
func (s *VaultSeal) ensureIdentity() {
	rc, out := s.gitf("config", "user.email")
	if rc != 0 || out == "" {
		s.gitf("config", "user.name", IdentityName)
		s.gitf("config", "user.email", IdentityEmail)
	}
}

// vaultIsDir mirrors self.vault.is_dir().
func (s *VaultSeal) vaultIsDir() bool {
	if s.vault == "" {
		return false
	}
	fi, err := os.Stat(s.vault)
	return err == nil && fi.IsDir()
}

// EnsureRepo mirrors ensure_repo: bootstrap the vault as a git repo
// with identity + remote. Returns (ok, error message). Safe to call
// repeatedly.
func (s *VaultSeal) ensureRepo() (bool, string) {
	if !s.vaultIsDir() {
		return false, "vault directory not found: " + s.vaultRepr()
	}

	// 1) The vault must be its OWN repository (git add -A is
	//    repo-wide from a subdirectory in git ≥ 2.0): compare the
	//    resolved toplevel with the vault and bootstrap a nested repo
	//    on mismatch — which also covers the plain no-repo-yet case.
	vaultReal := pyRealpath(s.vault)
	rc, out := s.gitf("rev-parse", "--show-toplevel")
	toplevel := ""
	if rc == 0 && strings.TrimSpace(out) != "" {
		toplevel = pyRealpath(strings.TrimSpace(out))
	}
	if toplevel != vaultReal {
		rc, out = s.gitf("init", "-b", "main")
		if rc != 0 {
			// older git without -b: plain init + move HEAD to main
			rc2, out2 := s.gitf("init")
			if rc2 == 0 {
				s.gitf("symbolic-ref", "HEAD", "refs/heads/main")
			} else {
				initErr := out2
				if initErr == "" {
					initErr = out
				}
				return false, "git init failed: " + initErr
			}
		}
		s.log("VaultSeal: initialised a git repository inside the vault", "info")
	}

	// 2) machine-state hygiene (idempotent)
	s.EnsureGitignore()

	// 3) local identity
	s.ensureIdentity()

	// 4) remote (only when pushing is possible)
	if !s.autoPush {
		return true, ""
	}
	// v0.83.0 — no slug fallback: the repo name must be the owner's
	// explicit word. An empty name means NOT CONFIGURED — fail FAST,
	// before any network call.
	repo := strings.TrimSpace(s.repoName)
	if repo == "" {
		return false, "no backup repo configured — vaultseal.repo_name is empty. " +
			"Nothing was created or pushed. Set the repo name in Settings → Backup " +
			"(it is remembered like credentials once you type it)."
	}
	owner := s.ownerLogin()
	if owner == "" {
		// v32.1: actionable — the token was rejected, not a mystery.
		return false, "could not resolve the GitHub account for the token " +
			"(invalid or expired — update the GitHub Token in Settings → " +
			"Credentials, then 'Test GitHub Token')"
	}

	status, payload := s.apiFn("GET", "/repos/"+owner+"/"+repo, nil)
	if status == 404 {
		if !s.allowCreate {
			return false, "repository " + owner + "/" + repo + " not found — " +
				"GitCurator does NOT create repositories automatically (the " +
				"duplicate-repo accident cure, v0.83.0). Create it yourself on " +
				"GitHub — or set vaultseal.create_repo=true in config.json once, " +
				"or pass --create-repo to the CLI — then re-run."
		}
		status, payload = s.apiFn("POST", "/user/repos", map[string]any{
			"name":         repo,
			"private":      true,
			"description":  "VaultSeal — automatic Obsidian vault backup (GitCurator)",
			"has_issues":   false,
			"has_wiki":     false,
			"has_projects": false,
			"auto_init":    false,
		})
		if status != 201 {
			// 422 usually means the name is taken (race) — re-check.
			reStatus, _ := s.apiFn("GET", "/repos/"+owner+"/"+repo, nil)
			if reStatus != 200 {
				return false, "could not create the private repo " + owner + "/" +
					repo + ": " + apiMessage(status, payload)
			}
		}
		s.log("VaultSeal: created private repository "+owner+"/"+repo, "success")
	} else if status != 200 {
		return false, "GitHub API error checking " + owner + "/" + repo + ": " +
			apiMessage(status, payload)
	}

	url := "https://github.com/" + owner + "/" + repo + ".git"
	rc, _ = s.gitf("remote", "get-url", "origin")
	if rc != 0 {
		s.gitf("remote", "add", "origin", url) // token-less remote
	}
	s.repoFull = owner + "/" + repo
	s.remoteURL = "https://github.com/" + owner + "/" + repo
	return true, ""
}
