// evidence.go — the radar's evidence pieces: parse_remote_repo +
// remote_origin_repo (evidence 1, the seal's signature), blob_sha +
// sample_vault_files (evidence 3's fingerprint math), and the
// candidate ranking + dominance ladder.
package reporadar

import (
	"crypto/sha1"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// remotePatterns mirrors _REMOTE_PATTERNS (the https and ssh forms —
// the [/:] class covers the colon form; ssh:// prefixes are tolerated
// by search).
var remotePatterns = []*regexp.Regexp{
	regexp.MustCompile(`github\.com[/:]([^/:\s]+)/([^/\s]+?)(?:\.git)?/?\s*$`),
}

// ParseRemoteRepo mirrors parse_remote_repo: owner/repo out of a
// GitHub remote URL (https, ssh, git@ forms), "" when the URL does
// not point at github.com (or the owner is git/www/api).
func ParseRemoteRepo(url string) string {
	u := strings.TrimSpace(url)
	if u == "" || !strings.Contains(helpers.PyLower(u), "github.com") {
		return ""
	}
	for _, pat := range remotePatterns {
		if m := pat.FindStringSubmatch(u); m != nil {
			owner := strings.TrimSpace(helpers.PyLower(m[1]))
			repo := strings.TrimSpace(m[2])
			if owner != "" && repo != "" && owner != "git" && owner != "www" && owner != "api" {
				return owner + "/" + repo
			}
		}
	}
	return ""
}

// RemoteOriginRepo mirrors remote_origin_repo: the origin remote of
// the vault's git repo, as owner/repo — the signature VaultSeal itself
// leaves behind. "" when the vault has no readable .git/config with an
// origin. The configparser read converges to a minimal INI scan
// ([remote "origin"] → url = …; full-line #/; comments skipped) — the
// reference's except door maps malformed files to "" and so does the
// lenient scan (a documented convergence: a malformed config never
// carries a findable origin in practice).
func RemoteOriginRepo(vaultPath string) string {
	raw := strings.TrimSpace(vaultPath)
	if raw == "" {
		return ""
	}
	cfgPath := filepath.Join(pyExpandUser(raw), ".git", "config")
	data, err := os.ReadFile(cfgPath)
	if err != nil {
		return ""
	}
	return parseOriginURL(string(data))
}

// parseOriginURL is the configparser subset: find the
// [remote "origin"] section, read its url key (case-insensitive —
// configparser lowercases option names).
func parseOriginURL(text string) string {
	const section = `remote "origin"`
	inSection := false
	for _, ln := range helpers.SplitLines(text) {
		line := strings.TrimSpace(ln)
		if line == "" || strings.HasPrefix(line, "#") || strings.HasPrefix(line, ";") {
			continue
		}
		if strings.HasPrefix(line, "[") && strings.HasSuffix(line, "]") {
			inSection = strings.TrimSpace(line[1:len(line)-1]) == section
			continue
		}
		if !inSection {
			continue
		}
		eq := strings.Index(line, "=")
		colon := strings.Index(line, ":")
		sep := -1
		if eq >= 0 && (colon < 0 || eq < colon) {
			sep = eq
		} else if colon >= 0 {
			sep = colon
		}
		if sep < 0 {
			continue
		}
		key := strings.TrimSpace(line[:sep])
		if strings.ToLower(key) != "url" {
			continue
		}
		return ParseRemoteRepo(strings.TrimSpace(line[sep+1:]))
	}
	return ""
}

// BlobSha mirrors blob_sha: the git blob SHA-1 of data — identical to
// what GitHub's tree API reports for the same content, whatever the
// file's path.
func BlobSha(data []byte) string {
	h := sha1.New()
	h.Write([]byte("blob " + strconv.Itoa(len(data)) + "\x00"))
	h.Write(data)
	return hex.EncodeToString(h.Sum(nil))
}

// SampleVaultFiles mirrors sample_vault_files: a deterministic,
// diversified sample of the vault's files for fingerprinting — sorted
// walk, .git/.trash/.DS_Store skipped, files grouped by their
// top-level folder and taken round-robin so one huge folder cannot
// monopolise the sample. Returns relative POSIX-style paths. A BLANK
// path is nothing — never the process's own working folder (the
// v0.84.0 dev catch; the test suite pins this law).
//
// Convergence (documented in the spec's Go notes): CPython's os.walk
// yields filenames in readdir (filesystem) order; Go's ReadDir sorts
// by name — the diversification law is order-insensitive (the
// round-robin takes one file per group per round either way), and the
// goldens are generated from trees written in sorted order so both
// sides agree.
func SampleVaultFiles(vaultPath string, maxFiles, maxBytes int) []string {
	raw := strings.TrimSpace(vaultPath)
	if raw == "" {
		return nil
	}
	root := pyExpandUser(raw)
	if !isDir(root) {
		return nil
	}
	if maxFiles <= 0 {
		maxFiles = MaxSampleFiles
	}
	if maxBytes <= 0 {
		maxBytes = MaxSampleBytes
	}
	groups := map[string][]string{}
	skipDirs := map[string]bool{".git": true, ".trash": true, ".DS_Store": true}
	walkSample(root, root, skipDirs, maxBytes, groups)
	orderedKeys := make([]string, 0, len(groups))
	for k := range groups {
		orderedKeys = append(orderedKeys, k)
	}
	sort.Strings(orderedKeys)
	var sample []string
	idx := 0
	for len(sample) < maxFiles {
		took := false
		for _, key := range orderedKeys {
			if idx < len(groups[key]) {
				sample = append(sample, groups[key][idx])
				took = true
				if len(sample) >= maxFiles {
					break
				}
			}
		}
		if !took {
			break
		}
		idx++
	}
	return sample
}

// walkSample is os.walk's pruned twin for the sampler (dirnames
// sorted+filtered in place, filenames in ReadDir order, size-capped).
func walkSample(root, dirpath string, skipDirs map[string]bool, maxBytes int,
	groups map[string][]string) {
	dirents, err := os.ReadDir(dirpath)
	if err != nil {
		return
	}
	var dirnames []string
	for _, e := range dirents {
		if e.IsDir() {
			if !skipDirs[e.Name()] {
				dirnames = append(dirnames, e.Name())
			}
			continue
		}
		fn := e.Name()
		rel, err := filepath.Rel(root, filepath.Join(dirpath, fn))
		if err != nil {
			continue
		}
		rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")
		top := "."
		if i := strings.Index(rel, "/"); i >= 0 {
			top = rel[:i]
		}
		if fi, err := os.Stat(filepath.Join(dirpath, fn)); err != nil || fi.Size() > int64(maxBytes) {
			continue
		}
		groups[top] = append(groups[top], rel)
	}
	sort.Strings(dirnames)
	for _, d := range dirnames {
		walkSample(root, filepath.Join(dirpath, d), skipDirs, maxBytes, groups)
	}
}

// localBlobShas mirrors _local_blob_shas: {blob_sha: 1} for the
// sampled files — a map (not a set) so the shape stays JSON-friendly
// for reports. Unreadable files are skipped.
func localBlobShas(vaultPath string) map[string]int {
	root := pyExpandUser(strings.TrimSpace(vaultPath))
	out := map[string]int{}
	for _, rel := range SampleVaultFiles(vaultPath, 0, 0) {
		data, err := os.ReadFile(filepath.Join(root, rel))
		if err != nil {
			continue
		}
		sha := BlobSha(data)
		if _, ok := out[sha]; !ok {
			out[sha] = 1
		}
	}
	return out
}

// fingerprintCandidates mirrors _fingerprint_candidates: candidate
// repos for fingerprinting, best-first — non-archived, non-fork, with
// a name; vault-ish names ahead of the rest.
//
// The live-account lesson (v0.84.0 dev): the repos API's “size“
// field is KILOBYTES ROUNDED DOWN — a repo holding one small README
// reports size=0 while being perfectly matchable, so size is NOT an
// emptiness signal here (a truly empty repo returns an empty tree and
// simply never matches).
func fingerprintCandidates(repos []map[string]any, vaultName string) []map[string]any {
	var live []map[string]any
	for _, r := range repos {
		if pyTruthy(r["archived"]) || pyTruthy(r["fork"]) {
			continue
		}
		name, _ := r["name"].(string)
		if name == "" {
			continue
		}
		live = append(live, r)
	}
	vaultTokens := tokenSet(helpers.PyLower(vaultName))

	type scored struct {
		repo map[string]any
		key  int
		name string
	}
	var ranked []scored
	for _, r := range live {
		name, _ := r["name"].(string)
		lower := helpers.PyLower(name)
		nameTokens := tokenSet(lower)
		overlap := 0
		for t := range vaultTokens {
			if nameTokens[t] {
				overlap++
			}
		}
		vaultish := false
		for _, t := range vaultishTokens {
			if strings.Contains(lower, t) {
				vaultish = true
				break
			}
		}
		score := overlap
		if vaultish {
			score += 2
		}
		ranked = append(ranked, scored{repo: r, key: -score, name: lower})
	}
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].key != ranked[j].key {
			return ranked[i].key < ranked[j].key
		}
		return ranked[i].name < ranked[j].name
	})
	out := make([]map[string]any, 0, len(ranked))
	for _, r := range ranked {
		out = append(out, r.repo)
	}
	if len(out) > MaxTreeFetches {
		out = out[:MaxTreeFetches]
	}
	return out
}

// tokenSet mirrors {t for t in re.split(r"[^a-z0-9]+", s) if t}.
func tokenSet(s string) map[string]bool {
	out := map[string]bool{}
	for _, t := range regexp.MustCompile(`[^a-z0-9]+`).Split(s, -1) {
		if t != "" {
			out[t] = true
		}
	}
	return out
}

// fingerprintResult mirrors _fingerprint's tuple.
type fingerprintResult struct {
	repo      string // "" = None
	matched   int
	sampled   int
	candidate []string // candidates_seen (names)
	why       string
}

// fingerprint mirrors _fingerprint: a winner must clear
// MinDominantMatches and beat the runner-up by DominanceFactor —
// near-ties are ambiguity, sub-threshold matches are weakness, and
// BOTH answer no winner (a report, never a guess).
func fingerprint(vaultPath string, repos []map[string]any, api API, vaultName string) fingerprintResult {
	local := localBlobShas(vaultPath)
	if len(local) == 0 {
		return fingerprintResult{why: "empty"}
	}
	sampled := len(local)
	candidates := fingerprintCandidates(repos, vaultName)
	type pair struct {
		matched int
		name    string
	}
	var ranked []pair
	for _, repo := range candidates {
		fullName, _ := repo["full_name"].(string)
		ref, _ := repo["default_branch"].(string)
		if ref == "" {
			ref = "main"
		}
		shas := api.TreeBlobShas(fullName, ref)
		if len(shas) == 0 { // nil OR empty — the reference's `if not shas`
			continue
		}
		matched := 0
		for sha := range local {
			if shas[sha] {
				matched++
			}
		}
		if matched > 0 {
			name, _ := repo["name"].(string)
			ranked = append(ranked, pair{matched: matched, name: name})
		}
	}
	if len(ranked) == 0 {
		var seen []string
		for _, r := range candidates {
			seen = append(seen, helpers.PyStr(r["name"]))
		}
		return fingerprintResult{sampled: sampled, candidate: seen, why: "none"}
	}
	// ranked.sort(reverse=True) on (matched, name) tuples: matched
	// descending, then NAME descending for ties.
	sort.SliceStable(ranked, func(i, j int) bool {
		if ranked[i].matched != ranked[j].matched {
			return ranked[i].matched > ranked[j].matched
		}
		return ranked[i].name > ranked[j].name
	})
	bestN, bestRepo := ranked[0].matched, ranked[0].name
	runnerUp := 0
	if len(ranked) > 1 {
		runnerUp = ranked[1].matched
	}
	threshold := MinDominantMatches
	if sampled < MinDominantMatches {
		threshold = sampled
	}
	var names []string
	for _, r := range ranked {
		names = append(names, r.name)
	}
	if bestN >= threshold && bestN >= DominanceFactor*runnerUp {
		return fingerprintResult{repo: bestRepo, matched: bestN, sampled: sampled,
			candidate: names, why: "dominant"}
	}
	if bestN >= threshold {
		return fingerprintResult{matched: bestN, sampled: sampled, candidate: names,
			why: "ambiguous"}
	}
	return fingerprintResult{matched: bestN, sampled: sampled, candidate: names, why: "weak"}
}

// pyExpandUser mirrors Path(raw).expanduser() (the vaultseal twin).
func pyExpandUser(raw string) string {
	p := raw
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			p = filepath.Join(home, strings.TrimPrefix(p, "~"))
		}
	}
	return p
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
