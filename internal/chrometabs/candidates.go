// candidates.go — which links earn the offer (chrome_tabs.py:2342-
// 2383): collect_failed_fetch_links — the modal's data, pure.
package chrometabs

import (
	"strings"

	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// Link is the candidates' item shape: {'url', 'error'}.
type Link struct {
	URL   string
	Error string
}

// CollectFailedFetchLinks mirrors collect_failed_fetch_links: the
// links whose FETCH failed this run — outcome 'review' with
// fetch_status 'failed' (the placeholder-in-_review pile — the
// master table's " - " rows) or outcome 'failed' (a hard pipeline
// miss). Deduped by canonical URL (first wins), loopback dropped, and
// a link the fetcher itself retired (auto-verdict: dead / paywalled /
// refused — the ladder's own final answers, checked through
// isDismissed when the caller can) never nags: those have their doors
// already. Returns the links in run order.
func CollectFailedFetchLinks(results []websitepipeline.Result,
	isDismissed func(canonical string) bool) []Link {
	var out []Link
	seen := map[string]bool{}
	for _, r := range results {
		fetchFailed := r.Outcome == "review" && r.FetchStatus == "failed"
		hardFailed := r.Outcome == "failed"
		if !(fetchFailed || hardFailed) {
			continue
		}
		u := strings.TrimSpace(r.URL)
		lu := strings.ToLower(u)
		if !strings.HasPrefix(lu, "http://") && !strings.HasPrefix(lu, "https://") {
			continue
		}
		if IsLoopbackURL(u) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(u)
		if canonical == "" || seen[canonical] {
			continue
		}
		if isDismissed != nil {
			dismissed := false
			func() {
				defer func() { recover() }() // a state question
				// that cannot be asked: offer it
				dismissed = isDismissed(canonical)
			}()
			if dismissed {
				continue // the fetcher's own verdict already answered
			}
		}
		seen[canonical] = true
		out = append(out, Link{URL: canonical,
			Error: strings.TrimSpace(r.Error)})
	}
	return out
}
