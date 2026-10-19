// deliver.go — the delivery (chrome_tabs.py:2391-2563): the fifth
// door end to end — queue the failed links (the same queue.json the
// fourth door reads, stamped door 'auto'), fetch every page through
// the owner's real Chrome, re-verify each page before writing, and
// write it as its suggested filename into <vault>/_review/
// hand-delivered/. Also THE WIRING: this package's init sets
// handdelivery.PageIsRealFn — the consume-side re-verification seam
// T-43 staged — so any binary that links the fifth door gets the
// real page_is_real verdict at consume time (the reference's lazy
// `from chrome_tabs import page_is_real` inside _auto_page_is_real,
// which never fails in the real app; handdelivery itself must not
// import chrometabs — the cycle the seam breaks).
package chrometabs

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/storage"
)

func init() {
	handdelivery.PageIsRealFn = func(pageHTML, href string) (bool, string) {
		return PageIsReal(pageHTML, href, "")
	}
}

// DeliverReport mirrors deliver_pages_via_chrome's report:
// {'delivered', 'failed', 'urls', 'failed_links', 'folder'}.
type DeliverReport struct {
	Delivered   int
	Failed      int
	URLs        []string
	FailedLinks []Link
	Folder      string
}

// FetcherFn is the delivery's _fetcher seam (urls → one result per
// url; the scripted tests' stand-in for the real Chrome fetch).
type FetcherFn func(urls []string, log LogFn) []PageResult

// DeliverPagesViaChrome mirrors deliver_pages_via_chrome. Never
// raises.
func DeliverPagesViaChrome(vaultPath string, linksList []Link, log LogFn,
	config map[string]any, chromeExe string, shouldContinue func() bool,
	fetcher FetcherFn) DeliverReport {
	if log == nil {
		log = noopLog
	}
	report := DeliverReport{
		Folder: handdelivery.HandDeliveryDir(vaultPath),
	}
	if vaultPath == "" || len(linksList) == 0 {
		return report
	}
	var candidates []Link
	for _, l := range linksList {
		lu := strings.ToLower(l.URL)
		if strings.HasPrefix(lu, "http://") || strings.HasPrefix(lu, "https://") {
			candidates = append(candidates, l)
		}
	}
	if len(candidates) == 0 {
		return report
	}
	if dryrun.IsEnabled() {
		log(fmt.Sprintf("🤖 Dry-run: the fifth door would open your real "+
			"Chrome with %d tab(s) and deliver each page into %s — "+
			"nothing was launched, nothing was written", len(candidates),
			report.Folder), "info")
		return report
	}
	if config == nil {
		config = map[string]any{}
	}
	pageTimeoutS := cfgFloatOr(config, "web_browser_retry_timeout_s",
		DefaultPageTimeoutS)
	wave := int(cfgFloatOr(config, "web_browser_retry_wave", DefaultWave))
	// the door's identity knobs: the persistent dedicated profile is
	// ON by default (the anti-bot law); the explicit dir override is
	// honored when set.
	persist := cfgBoolOrDefault(config, ConfigProfilePersist, true)
	profileDir := strings.TrimSpace(cfgStr(config, ConfigProfileDir))
	// the owner's own window + the right route.
	attach := cfgBoolOrDefault(config, ConfigAttach, true)
	attachPort := cfgIntOrDefault(config, ConfigAttachPort, DefaultAttachPort)
	realProfile := cfgBoolOrDefault(config, ConfigRealProfile, true)
	proxyMode := NormalizeProxyMode(cfgStr(config, ConfigProxyLegs))
	proxy := proxyFromConfigSafe(config)
	memoryPath := ProxyMemoryPath()
	if proxy != nil && proxyMode == "auto" {
		log(fmt.Sprintf("🤖 Fifth door, v0.55: your own Chrome when its "+
			"debug port answers (127.0.0.1:%d); otherwise the door's own "+
			"identity — DIRECT first, then through your proxy (%s) — and "+
			"every site remembers the route that worked", attachPort,
			ProxyLabelForLog(proxy)), "info")
	}
	var urls []string
	walls := map[string]string{}
	for _, l := range candidates {
		urls = append(urls, l.URL)
		walls[l.URL] = l.Error
	}
	// the queue row is written FIRST (door 'auto') so the consume path
	// finds the page; an enqueue failure is logged and tolerated — an
	// already-queued link's page still lands.
	func() {
		defer func() {
			if r := recover(); r != nil {
				log(fmt.Sprintf("⚠️ Fifth-door queue write "+
					"skipped: %v", r), "warning")
			}
		}()
		handdelivery.EnqueueHandDelivery(vaultPath, urls, walls, log,
			false, "auto")
	}()

	var pages []PageResult
	if fetcher != nil {
		pages = func() (out []PageResult) {
			defer func() {
				if r := recover(); r != nil {
					log(fmt.Sprintf("⚠️ The Chrome tab fetch failed "+
						"outright: %v", r), "warning")
					out = explodedPages(urls,
						fmt.Sprintf("chrome exploded: %v", r))
				}
			}()
			return fetcher(urls, log)
		}()
	} else {
		opts := DefaultFetchOptions()
		opts.Log = log
		opts.ChromeExe = chromeExe
		opts.PageTimeoutS = pageTimeoutS
		opts.Wave = wave
		opts.ShouldContinue = shouldContinue
		opts.ProfileDir = profileDir
		opts.Persist = persist
		opts.Proxy = proxy
		opts.ProxyMode = proxyMode
		opts.Attach = attach
		opts.AttachPort = attachPort
		opts.RealProfile = realProfile
		opts.MemoryPath = memoryPath
		pages = func() (out []PageResult) {
			defer func() {
				if r := recover(); r != nil {
					log(fmt.Sprintf("⚠️ The Chrome tab fetch failed "+
						"outright: %v", r), "warning")
					out = explodedPages(urls,
						fmt.Sprintf("chrome exploded: %v", r))
				}
			}()
			return FetchPagesViaChrome(urls, opts)
		}()
	}
	folder := report.Folder
	for _, page := range pages {
		u := page.URL
		if !page.OK || page.HTML == "" {
			report.Failed++
			errStr := strings.TrimSpace(page.Error)
			if errStr == "" {
				errStr = "no content"
			}
			report.FailedLinks = append(report.FailedLinks,
				Link{URL: u, Error: errStr})
			continue
		}
		// the verdict is the DELIVERY's law too (the fetch driver
		// already judges its own tabs; this guard keeps the delivered
		// count honest whatever the fetch path): a page that reads as
		// a Chrome error page, a crash or a challenge is a named
		// failure here, never a written file.
		ok, reason := PageIsReal(page.HTML, u, page.Title)
		if !ok {
			report.Failed++
			report.FailedLinks = append(report.FailedLinks,
				Link{URL: u, Error: reason})
			log(fmt.Sprintf("⚠️ %s: %s", u, reason), "warning")
			continue
		}
		suggested := handdelivery.SuggestedFilename(u)
		path := filepath.Join(folder, suggested)
		content := page.HTML
		if err := storage.AtomicWriteBytes(path, []byte(content)); err != nil {
			report.Failed++
			report.FailedLinks = append(report.FailedLinks,
				Link{URL: u,
					Error: "could not write the page: " + err.Error()})
			log(fmt.Sprintf("⚠️ Could not write the delivered page for "+
				"%s: %s", u, err.Error()), "warning")
			continue
		}
		report.Delivered++
		report.URLs = append(report.URLs, u)
		log(fmt.Sprintf("🤖 Delivered %s (%d KB of live DOM from your "+
			"Chrome) — it is a real fetch now", suggested,
			len([]rune(content))/1024), "success")
	}
	if report.Delivered > 0 {
		msg := fmt.Sprintf("🤖 Fifth door: %d page(s) delivered from your "+
			"real Chrome into %s", report.Delivered, folder)
		if report.Failed > 0 {
			msg += fmt.Sprintf(", %d failed", report.Failed)
		}
		msg += " — they will be processed as real fetches"
		log(msg, "success")
	} else if report.Failed > 0 {
		log(fmt.Sprintf("🤖 Fifth door: every tab failed (%d) — the links "+
			"keep waiting in the retry queue", report.Failed), "warning")
	}
	if len(report.FailedLinks) > 0 && report.Delivered > 0 {
		// a PARTIAL delivery is told too — the owner asked for the
		// numbers, not silence after the fanfare
		shown := report.FailedLinks
		if len(shown) > 5 {
			shown = shown[:5]
		}
		for _, fl := range shown {
			log(fmt.Sprintf("⚠️ %s: %s", fl.URL, fl.Error), "warning")
		}
	}
	return report
}

// explodedPages mirrors the both-except paths' per-URL failure rows.
func explodedPages(urls []string, msg string) []PageResult {
	out := make([]PageResult, 0, len(urls))
	for _, u := range urls {
		out = append(out, PageResult{URL: u, OK: false,
			Error: msg})
	}
	return out
}

// cfgStr reads one config knob as Python's str(config.get(k) or ”).
func cfgStr(config map[string]any, key string) string {
	v, ok := config[key]
	if !ok || v == nil {
		return ""
	}
	switch t := v.(type) {
	case string:
		return t
	case bool:
		if t {
			return "True"
		}
		return "False"
	case json.Number:
		return t.String()
	default:
		return storage.PyDumpsCompact(v)
	}
}

// cfgFloatOr mirrors float(config.get(key, def) or def): a falsy or
// unparseable value reads as the default.
func cfgFloatOr(config map[string]any, key string, def float64) float64 {
	v, ok := config[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case json.Number:
		if f, err := t.Float64(); err == nil && f != 0 {
			return f
		}
	case float64:
		if t != 0 {
			return t
		}
	case int:
		if t != 0 {
			return float64(t)
		}
	case int64:
		if t != 0 {
			return float64(t)
		}
	case string:
		if f, err := strconv.ParseFloat(strings.TrimSpace(t), 64); err == nil && f != 0 {
			return f
		}
	}
	return def
}

// cfgIntOrDefault mirrors int(config.get(key, def) or def) with the
// reference's try/except → def.
func cfgIntOrDefault(config map[string]any, key string, def int) int {
	f := cfgFloatOr(config, key, float64(def))
	return int(f)
}

// cfgBoolOrDefault mirrors config.get(key, True) with None→True and
// Python truthiness.
func cfgBoolOrDefault(config map[string]any, key string, def bool) bool {
	v, ok := config[key]
	if !ok {
		return def
	}
	if v == nil {
		return true
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != ""
	case float64:
		return t != 0
	case int:
		return t != 0
	case int64:
		return t != 0
	case json.Number:
		if f, err := t.Float64(); err == nil {
			return f != 0
		}
		return true
	default:
		return true
	}
}
