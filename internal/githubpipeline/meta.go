// meta.go — T-42: worker/github_meta.py (321 LOC) + the _missing
// placeholder from worker/notes.py:123-161 — the GitHub-side
// enrichment of a repo note.
//
//   - OrgReputation        _get_org_reputation (93-114)
//   - CalculateCredibility _calculate_credibility (236-281)
//   - RecordMissingRepo    _record_missing_repo (notes.py:123-161)
//   - BackfillMissingNotes _backfill_missing_notes (283-321)
package githubpipeline

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/notebuilder"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// orgMajor / orgMid mirror the mixin's literal reputation lists.
var orgMajor = []string{
	"microsoft", "google", "nvidia", "anthropic", "openai",
	"meta", "amazon", "apple", "ibm", "intel",
	"cloudflare", "aws", "azure", "googlecloud", "gcp",
	"netflix", "uber", "airbnb", "spotify", "twitter", "facebook",
	"github", "gitlab", "docker", "kubernetes", "linux", "redhat",
}

var orgMid = []string{
	"huggingface", "cohere", "together", "replit", "cursor",
	"vercel", "netlify", "railway", "flyio", "render",
	"supabase", "firebase", "mongodb", "elastic", "datadog",
}

// OrgReputation mirrors _get_org_reputation: 10 for the majors
// (substring match, case-insensitive), 7 for the mids, 5 for a
// multi-word org name longer than 3 chars, else 3.
func OrgReputation(org string) int {
	orgLower := strings.ToLower(org)
	for _, m := range orgMajor {
		if strings.Contains(orgLower, m) {
			return 10
		}
	}
	for _, m := range orgMid {
		if strings.Contains(orgLower, m) {
			return 7
		}
	}
	if strings.Contains(org, " ") && len(org) > 3 {
		return 5
	}
	return 3
}

// CredibilityDocSource is the repo view _calculate_credibility's doc
// score reads (has_issues / license / has_wiki / a >50-char
// description / a fetched README). The loop passes the already-fetched
// Repo; ReadmeFetchedAny carries the get_readme() probe result (the
// reference re-issues get_readme inside the score, best-effort and
// guarded — the loop's earlier README fetch is the same fact).
type CredibilityDocSource struct {
	HasIssues        bool
	HasLicense       bool
	HasWiki          bool
	Description      string
	ReadmeFetchedAny bool
}

// CalculateCredibility mirrors _calculate_credibility: the stars and
// activity ladders, the doc score (5 + up to five bonuses, capped 10),
// the weighted sum (0.4·org + 0.25·stars + 0.2·commits + 0.15·doc)
// ×10 rounded to one decimal, and the famous-org floor (org ≥ 10
// never scores below 90).
func CalculateCredibility(orgRep, stars, commits int, doc CredibilityDocSource) float64 {
	starsScore := 1
	switch {
	case stars > 10000:
		starsScore = 10
	case stars > 5000:
		starsScore = 8
	case stars > 1000:
		starsScore = 6
	case stars > 100:
		starsScore = 4
	case stars > 10:
		starsScore = 3
	}

	activityScore := 2
	switch {
	case commits > 50:
		activityScore = 10
	case commits > 20:
		activityScore = 7
	case commits > 5:
		activityScore = 4
	}

	docScore := 5
	if doc.HasIssues {
		docScore++
	}
	if doc.HasLicense {
		docScore++
	}
	if doc.HasWiki {
		docScore++
	}
	if doc.Description != "" && len(doc.Description) > 50 {
		docScore++
	}
	if doc.ReadmeFetchedAny {
		docScore++
	}
	if docScore > 10 {
		docScore = 10
	}

	raw := float64(orgRep)*0.4 + float64(starsScore)*0.25 +
		float64(activityScore)*0.2 + float64(docScore)*0.15
	score := pyRound1(raw * 10)
	if orgRep >= 10 && score < 90.0 {
		score = 90.0
	}
	return score
}

// pyRound1 mirrors Python's round(x, 1): a correctly-rounded decimal
// conversion to one digit of the EXACT binary value (round(2.675, 1)
// is 2.67 — the binary value is 2.674999…; never round(x*10)/10,
// which shifts onto a different float). The credibility golden pins
// the battery; a tie divergence would surface there and be fixed
// here, never in the golden.
func pyRound1(f float64) float64 {
	s := strconv.FormatFloat(f, 'f', 1, 64)
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return f
	}
	return v
}

// RecordMissingRepo mirrors _record_missing_repo (notes.py:123-161):
// write the “_missing/“ placeholder note for a 404 repo and
// immediately confirm-dead the quarantine row. The note's “source:“
// frontmatter is the VaultIndex dedupe key — with it in the vault the
// repo stops counting as pending in every queue view and never
// reaches the GitHub API again. Idempotent (an existing note is kept;
// the runner's per-batch set short-circuits repeats); crash-guarded;
// dry-run records only. Returns the note path, "" when the vault is
// unset or the write failed.
func (r *Runner) RecordMissingRepo(url, owner, repoName string,
	strikes int, cache *store.CacheDB, threshold int) (notePath string) {
	defer func() {
		if recover() != nil {
			notePath = ""
		}
	}()
	if r.missingNotesWritten[url] {
		return ""
	}
	vaultPath := cfgStr(r.Config, "vault_path")
	path := ""
	if vaultPath != "" {
		folder := filepath.Join(vaultPath, "_missing")
		if err := dryrun.MkdirAll(folder, true); err != nil {
			return ""
		}
		fname := storage.SafeFilename(owner+"_"+repoName+"_missing") + ".md"
		path = filepath.Join(folder, fname)
		if _, err := os.Stat(path); err != nil || dryrun.IsEnabled() {
			if err := storage.AtomicWriteText(path,
				notebuilder.BuildMissingRepoNote(url, owner, repoName,
					strikes)); err != nil {
				return ""
			}
		}
		if r.VaultIndex != nil {
			r.VaultIndex.AddURL(url, path)
		}
	}
	cache.ConfirmDead(url, "404 Not Found (missing-repo note)", threshold)
	r.missingNotesWritten[url] = true
	return path
}

// BackfillMissingNotes mirrors _backfill_missing_notes (github_meta.py:
// 283-321): repos that struck out in earlier versions (below the
// threshold, no note, no confirmation) get their “_missing“ note
// NOW so they stop counting as pending and stop burning API calls —
// the owner's "8-9 github addresses that are 404 always count as
// remaining". Runs once per batch start; idempotent.
func (r *Runner) BackfillMissingNotes(cache *store.CacheDB,
	vaultPath string, threshold int) {
	rows := cache.GetUnconfirmed404s(threshold)
	if len(rows) == 0 {
		return
	}
	written := 0
	for _, row := range rows {
		url := row.URL
		if !strings.Contains(url, "github.com/") {
			continue // defensive: decommissioned rows are github links
		}
		if r.VaultIndex != nil && r.VaultIndex.HasURL(url) {
			// A note already covers it — just confirm the quarantine.
			cache.ConfirmDead(url, "404 Not Found (missing-repo note exists)",
				threshold)
			continue
		}
		tail := url
		if i := strings.Index(url, "github.com/"); i >= 0 {
			tail = url[i+len("github.com/"):]
		}
		parts := strings.Split(tail, "/")
		if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
			continue
		}
		owner, repoName := parts[0], parts[1]
		repoName = strings.TrimSuffix(repoName, ".git")
		if r.RecordMissingRepo(url, owner, repoName, int(row.FailCount), cache,
			threshold) != "" {
			written++
		} else {
			cache.ConfirmDead(url, "404 Not Found (backfill)", threshold)
		}
	}
	if written > 0 {
		r.log("🕳️ "+strconv.Itoa(written)+" missing-repo note(s) written "+
			"(past 404s) — those repos no longer count as pending. To "+
			"re-check one, delete its note in _missing/ and reset it in "+
			"More ▸ View 404 Quarantine.", "info")
	}
}
