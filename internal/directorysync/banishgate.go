// banishgate.go — T-40's surface of cloud/directory_sync.py: the
// directory's banish gate (SPEC §11) that WebsitePipeline's
// _directory_banish_gate probes before any link is fetched.
//
// What ships here: load_banish_list, DEFAULT_SHARED_HOSTS, is_banished
// (+ the _norm_host floor) — the read side of
// Websites/_system/banished.json. The REST of the module
// (build_index, publish, the command cycle, the config trio) is
// T-49's own task.
package directorysync

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/links"
)

// DefaultSharedHosts mirrors DEFAULT_SHARED_HOSTS: shared-host
// platforms — a banish on these hosts bans the FULL URL, not the host
// (SPEC §11). The live list rides in directory-config.json
// ('shared_host_exceptions'); this is the floor.
var DefaultSharedHosts = []string{
	"github.io", "notion.site", "vercel.app", "blogspot.com",
	"wordpress.com", "netlify.app", "pages.dev", "gitlab.io",
	"webflow.io", "medium.com", "substack.com",
}

// SystemDirname mirrors SYSTEM_DIRNAME: the vault's own system folder
// (SPEC §7.4 banished.json + §11.3 files).
const SystemDirname = "_system"

// BanishedFilename mirrors BANISHED_FILENAME.
const BanishedFilename = "banished.json"

// BanishEntry mirrors one banished.json entry: host, url?, reason?,
// banished_at? (SPEC §7.4).
type BanishEntry struct {
	Host       string `json:"host"`
	URL        string `json:"url,omitempty"`
	Reason     string `json:"reason,omitempty"`
	BanishedAt string `json:"banished_at,omitempty"`
}

// SystemDir mirrors system_dir: <vault>/_system.
func SystemDir(vault string) string {
	if vault == "" {
		return SystemDirname
	}
	return filepath.Join(vault, SystemDirname)
}

// LoadBanishList mirrors load_banish_list: reads
// Websites/_system/banished.json. Missing file or a non-list shape =
// the empty list (never raises).
func LoadBanishList(vault string) []BanishEntry {
	path := filepath.Join(SystemDir(vault), BanishedFilename)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil
	}
	var entries []BanishEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil
	}
	return entries
}

// normHost mirrors _norm_host: the host out of a raw URL string, www.
// stripped, lowercased.
func normHost(rawURL string) string {
	host := rawURL
	if i := strings.Index(host, "//"); i >= 0 {
		host = host[i+2:]
	}
	if i := strings.Index(host, "/"); i >= 0 {
		host = host[:i]
	}
	if i := strings.Index(host, ":"); i >= 0 {
		host = host[:i]
	}
	if i := strings.Index(host, "?"); i >= 0 {
		host = host[:i]
	}
	host = strings.ToLower(host)
	host = strings.TrimPrefix(host, "www.")
	return host
}

// IsBanished mirrors is_banished: True when the url's HOST (or the
// exact normalized URL on a shared host) is on the banish list. Pure —
// the intake calls this BEFORE fetching anything.
func IsBanished(rawURL string, banished []BanishEntry, sharedHosts []string) bool {
	if len(banished) == 0 {
		return false
	}
	var hosts []string
	urlSet := map[string]bool{}
	for _, b := range banished {
		if b.Host != "" {
			hosts = append(hosts, strings.TrimPrefix(strings.ToLower(b.Host), "www."))
		}
		if b.URL != "" {
			urlSet[links.NormalizeWebsiteURL(b.URL)] = true
		}
	}
	shared := sharedHosts
	if shared == nil {
		shared = DefaultSharedHosts
	}
	host := normHost(rawURL)
	if host == "" {
		return false
	}
	hostHit := false
	for _, h := range hosts {
		if h != "" && (host == h || strings.HasSuffix(host, "."+h)) {
			hostHit = true
			break
		}
	}
	if hostHit {
		for _, s := range shared {
			if host == s || strings.HasSuffix(host, "."+s) {
				// Shared platform: only the exact URL is banned.
				return urlSet[links.NormalizeWebsiteURL(rawURL)]
			}
		}
		return true
	}
	return urlSet[links.NormalizeWebsiteURL(rawURL)]
}
