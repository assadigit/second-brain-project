// banner.go — T-42: _download_banner (worker/github_meta.py:116-234)
// — the GitHub social preview banner fetcher: opengraph.githubassets
// .com/1/{owner}/{repo} into the vault's central attachments/banners/
// folder, with the v25 pre-flight throttle (every 10 banners → 2 s,
// every 50 → 5 s), the 24 h failed-marker cache, the 429 backoff
// ladder (5/10/15 s over 3 attempts), the atomic write, and the
// netctx verify_ssl split (the banner path is an OUTBOUND path —
// connectioncheck.OutboundTLSConfig is the Go twin of
// netctx.outbound_ssl_context).
package githubpipeline

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/connectioncheck"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// bannerSafeName mirrors the mixin's own re.sub(r'[^a-zA-Z0-9\-_]+',
// '_', repo_name) — byte-identical to storage.SafeFilename (the same
// character class), kept here as the reference's own spelling.
var bannerSafeName = regexp.MustCompile(`[^a-zA-Z0-9\-_]+`)

// DownloadBanner mirrors _download_banner(owner, repo_name,
// folder_path): returns the local banner path, "" on failure. The
// folder_path argument only matters when the vault path is UNSET (the
// fallback vault root is its parent directory) — the central banners
// folder is always vault/attachments/banners otherwise.
func (r *Runner) DownloadBanner(owner, repoName, folderPath string) string {
	// v25 pre-flight: throttle banner downloads for large batches —
	// opengraph.githubassets.com 429s aggressively under a hammer.
	r.bannerCount++
	if r.bannerCount%50 == 0 {
		r.sleep(secs(pyFloatOr(r.Config, "banner_throttle_50", 5)))
	} else if r.bannerCount%10 == 0 {
		r.sleep(secs(pyFloatOr(r.Config, "banner_throttle_10", 2)))
	}

	vaultRoot := cfgStr(r.Config, "vault_path")
	if vaultRoot == "" {
		vaultRoot = filepath.Dir(folderPath)
	}
	bannersDir := filepath.Join(vaultRoot, "attachments", "banners")
	if err := dryrun.MkdirAll(bannersDir, true); err != nil {
		// best-effort — the reference's _dryrun.makedirs raises only
		// through real mode; the write below fails its own way.
		_ = err
	}

	url := fmt.Sprintf("https://opengraph.githubassets.com/1/%s/%s",
		owner, repoName)
	safeName := bannerSafeName.ReplaceAllString(repoName, "_")
	bannerPath := filepath.Join(bannersDir, safeName+"_banner.png")

	// Skip if already downloaded.
	if _, err := os.Stat(bannerPath); err == nil {
		return bannerPath
	}

	// Check if this repo previously failed (cache file marker).
	failedMarker := filepath.Join(bannersDir, safeName+"_failed.marker")
	if fi, err := os.Stat(failedMarker); err == nil {
		// Don't re-try known failures (marker auto-expires after 24h).
		if r.clockNow().Sub(fi.ModTime()) < 24*time.Hour {
			return ""
		}
		_ = dryrun.Remove(failedMarker)
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: connectioncheck.OutboundTLSConfig(r.Config),
		},
	}

	const maxRetries = 3
	for attempt := 0; attempt < maxRetries; attempt++ {
		req, err := http.NewRequest(http.MethodGet, url, nil)
		if err != nil {
			return ""
		}
		req.Header.Set("User-Agent", "Mozilla/5.0")
		req.Header.Set("Accept", "image/png,image/*")
		resp, err := client.Do(req)
		if err != nil {
			// The generic exception branch: 2 s, retry, then give up.
			if attempt < maxRetries-1 {
				r.sleep(2 * time.Second)
				continue
			}
			return ""
		}
		body, readErr := io.ReadAll(resp.Body)
		resp.Body.Close()
		if readErr != nil {
			// A read failure mid-body is the reference's generic
			// exception branch: 2 s, retry, then give up.
			if attempt < maxRetries-1 {
				r.sleep(2 * time.Second)
				continue
			}
			return ""
		}
		status := resp.StatusCode
		contentType := resp.Header.Get("Content-Type")
		switch {
		case status == http.StatusTooManyRequests:
			// Rate limited — wait and retry (5 s, 10 s, 15 s).
			if attempt < maxRetries-1 {
				wait := 5 * (attempt + 1)
				r.log(fmt.Sprintf(
					"   ⏳ Banner rate-limited (429), waiting %ds before "+
						"retry %d/%d...", wait, attempt+2, maxRetries),
					"warning")
				r.sleep(time.Duration(wait) * time.Second)
				continue
			}
			// Mark as failed to avoid re-trying for 24h.
			_ = dryrun.WriteText(failedMarker,
				fmt.Sprintf("%v", r.clockNow().Unix()))
			r.log(fmt.Sprintf(
				"   ⚠️ Banner rate-limited after %d attempts. Will skip "+
					"for 24h.", maxRetries), "warning")
			return ""
		case status == http.StatusNotFound:
			// No banner for this repo — mark as failed (permanent).
			_ = dryrun.WriteText(failedMarker, "404")
			return ""
		case status < 200 || status >= 300:
			if attempt < maxRetries-1 {
				r.sleep(2 * time.Second)
				continue
			}
			return ""
		default:
			// 2xx — the content-type gate, then the >1000-byte gate.
			if strings.Contains(contentType, "image") {
				if len(body) > 1000 {
					if err := storage.AtomicWriteBytes(bannerPath, body); err != nil {
						return ""
					}
					return bannerPath
				}
			}
			return ""
		}
	}
	return ""
}
