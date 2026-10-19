package syncreview

// doors.go — the read-only doors: the ONE network door
// (fetch_published_index — GET-only, UA worn, never raises) and its
// URL extraction half. The net_guard law: method="GET", any failure
// answers nil, the honest verdict for an unmeasured side is UNKNOWN,
// never a guess.

import (
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// PublishedIndexPath mirrors PUBLISHED_INDEX_PATH: the Worker's public
// index endpoint (GET — no token, CORS-open).
const PublishedIndexPath = "/api/directory/index"

// httpDo is the seam the tests fake (the HERMETIC-SEAM law at the
// package level: the production door builds the request — GET, the UA
// worn, JSON accepted — and hands it to this seam; the tests inspect
// the request and answer canned responses without a socket).
var httpDo = func(req *http.Request, timeoutS float64) (*http.Response, error) {
	if timeoutS <= 0 {
		timeoutS = 15
	}
	client := &http.Client{
		Timeout: time.Duration(timeoutS * float64(time.Second)),
	}
	return client.Do(req)
}

// FetchPublishedIndex mirrors fetch_published_index: GET
// {worker_url}/api/directory/index — the published Website Directory
// (the target side of Websites · Vault → Website). Read-only,
// GET-only, never raises: any failure answers nil. Requires a dict
// with a 'sites' list.
func FetchPublishedIndex(workerURL string, timeoutS float64) *storage.OrderedMap {
	url := strings.TrimRight(strings.TrimSpace(workerURL), "/")
	if url == "" {
		return nil
	}
	lower := strings.ToLower(url)
	if !strings.HasPrefix(lower, "http://") && !strings.HasPrefix(lower, "https://") {
		return nil
	}
	req, err := http.NewRequest("GET", url+PublishedIndexPath, nil)
	if err != nil {
		return nil
	}
	// THE 1010 LAW: the UA literal lives in workerclient (the
	// reference imports it from cloudflare_sync so the version bump
	// stays one file's business).
	req.Header.Set("User-Agent", workerclient.HTTPUserAgent)
	req.Header.Set("Accept", "application/json")
	res, err := httpDo(req, timeoutS)
	if err != nil {
		return nil
	}
	raw, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return nil
	}
	data, derr := storage.DecodeJSONValue(raw)
	if derr != nil {
		return nil
	}
	om, ok := data.(*storage.OrderedMap)
	if !ok {
		return nil
	}
	sites, present := om.Get("sites")
	if !present {
		return nil
	}
	if _, isList := sites.([]any); !isList {
		return nil
	}
	return om
}

// PublishedSiteUrls mirrors published_site_urls: the published
// index's site URLs (normalized the build_index way). ok=false when
// the index is nil (unmeasured); the honest empty list when the index
// is live but empty. Junk site rows never crash the walk.
func PublishedSiteUrls(index *storage.OrderedMap) (urls []string, ok bool) {
	if index == nil {
		return nil, false
	}
	urls = []string{}
	sites, _ := index.Get("sites")
	list, _ := sites.([]any)
	for _, site := range list {
		var raw any
		switch entry := site.(type) {
		case *storage.OrderedMap:
			raw, _ = entry.Get("url")
		case map[string]any:
			raw = entry["url"]
		default:
			continue // a non-dict row is skipped, never a crash
		}
		normalized := links.NormalizeWebsiteURL(pyStr(raw))
		if normalized != "" {
			urls = append(urls, normalized)
		}
	}
	return urls, true
}

// pyStr mirrors str(v or ”) for the URL seat (junk-tolerant).
func pyStr(v any) string {
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}
