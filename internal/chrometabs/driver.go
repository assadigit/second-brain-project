// driver.go — the fifth door's engine (chrome_tabs.py:1832-2033):
// fetch_pages_via_chrome — the owner's own window when it answers,
// the right route when the door must launch its own Chrome. ATTACH
// first, then THE LEGS with escalation of direct-leg failures onto
// the proxy leg, per-site route memory, one tab per link in waves,
// the content read from the live DOM. Never raises: no Chrome found,
// a refused launch and every per-tab miss are honest per-URL
// failures; a both-legs failure carries both legs' sentences.
package chrometabs

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/webfetch"
)

// findChromeFn is the finder seam (the reference's driver imports
// hand_delivery lazily and its tests patch hd.find_chrome).
var findChromeFn = handdelivery.FindChrome

// FetchOptions mirrors fetch_pages_via_chrome's full parameter list
// (the reference's keyword defaults are DefaultFetchOptions' own).
type FetchOptions struct {
	Log            LogFn
	ChromeExe      string
	PageTimeoutS   float64
	Wave           int
	ShouldContinue func() bool
	// SessionFactory is the tests' _session_factory seam: the scripted
	// session IS the ladder's answer — no attach probe, one leg.
	SessionFactory func(exe string, log LogFn) (TabSession, error)
	ProfileDir     string
	Persist        bool
	Proxy          *ProxyConfig
	ProxyMode      string
	Attach         bool
	AttachPort     int
	RealProfile    bool
	MemoryPath     string
}

// DefaultFetchOptions carries the reference's keyword defaults.
func DefaultFetchOptions() FetchOptions {
	return FetchOptions{
		PageTimeoutS: DefaultPageTimeoutS,
		Wave:         DefaultWave,
		Persist:      true,
		ProxyMode:    "auto",
		Attach:       true,
		AttachPort:   DefaultAttachPort,
		RealProfile:  true,
	}
}

// FetchPagesViaChrome mirrors fetch_pages_via_chrome. Returns one
// PageResult per URL.
func FetchPagesViaChrome(urls []string, opts FetchOptions) []PageResult {
	log := opts.Log
	if log == nil {
		log = noopLog
	}
	var results []PageResult
	var work []string
	for _, raw := range urls {
		u := strings.TrimSpace(raw)
		lu := strings.ToLower(u)
		if !strings.HasPrefix(lu, "http://") && !strings.HasPrefix(lu, "https://") {
			results = append(results, PageResult{URL: u, OK: false,
				Error: "not an http(s) link"})
			continue
		}
		if IsLoopbackURL(u) {
			results = append(results, PageResult{URL: u, OK: false,
				Error: "loopback URL — no browser opens (the v0.15.1 rule)"})
			continue
		}
		work = append(work, u)
	}
	if len(work) == 0 {
		return results
	}
	if dryrun.IsEnabled() { // belt and braces — the delivery layer guards
		for _, u := range work {
			results = append(results, PageResult{URL: u, OK: false,
				Error: "dry-run — Chrome is never launched"})
		}
		return results
	}
	chrome := opts.ChromeExe
	if chrome == "" {
		chrome = findChromeFn()
	}
	if chrome == "" {
		for _, u := range work {
			results = append(results, PageResult{URL: u, OK: false,
				Error: "Google Chrome was not found on this machine " +
					"(the fifth door needs the real browser)"})
		}
		log("🤖 Fifth door: no Chrome found — nothing was auto-fetched",
			"warning")
		return results
	}
	wave := opts.Wave
	if wave == 0 {
		wave = DefaultWave
	}
	if wave < 1 {
		wave = 1
	}
	pageTimeoutS := opts.PageTimeoutS
	if pageTimeoutS == 0 {
		pageTimeoutS = DefaultPageTimeoutS
	}

	// THE OWNER'S OWN WINDOW: his running Chrome's DevTools port
	// answers → the tabs open in HIS browser, and the door never
	// closes it. The scripted tests' factory bypasses the probe
	// (hermetic by law).
	if opts.Attach && opts.SessionFactory == nil && opts.AttachPort != 0 {
		if probeVersionEndpoint(opts.AttachPort, 3.0) {
			log(fmt.Sprintf("🤖 ATTACHED to your own running Chrome (DevTools "+
				"on 127.0.0.1:%d) — the tabs open in YOUR window: your "+
				"profile, your cookies, your logins, your network path. "+
				"GitCurator closes only its own tabs.", opts.AttachPort), "info")
			session := NewAttachedChromeSession(opts.AttachPort, log)
			pages := gatherThroughSession(session, work, pageTimeoutS, wave,
				opts.ShouldContinue, log)
			session.Close() // gentle — only the door's own tabs
			results = append(results, pages...)
			return results
		}
		log(fmt.Sprintf("🤖 Your Chrome's debug port (127.0.0.1:%d) did not "+
			"answer — the door drives its own identities (More ▸ 🪄 Attach "+
			"to my Chrome teaches your Chrome to answer; it needs to be "+
			"running with the port open, or closed so the door can launch "+
			"your real profile)", opts.AttachPort), "info")
	}

	// THE RIGHT ROUTE: the legs. Direct first (the local line the
	// challenges like), then the owner's proxy; a link that fails the
	// first leg is escalated onto the second in the same delivery, and
	// every site remembers the route that delivered.
	memory := storage.NewOrderedMap()
	if opts.MemoryPath != "" {
		memory = LoadProxyMemory(opts.MemoryPath)
	}
	memoryDirty := false
	legs := PlanProxyLegs(work, memory, opts.Proxy, opts.ProxyMode)
	mode := NormalizeProxyMode(opts.ProxyMode)
	escalate := mode == "auto" && len(ChromeProxyFlag(opts.Proxy)) > 0
	if mode == "auto" && len(ChromeProxyFlag(opts.Proxy)) == 0 {
		log("🤖 No proxy is configured (Settings → Proxy) — the proxy leg "+
			"is skipped; every page rides the direct line", "info")
	}
	outcomes := map[string]PageResult{}
	legErrors := map[string][]string{}
	stopped := false
	for _, leg := range legs {
		legURLs := leg.URLs
		if escalate && leg.Route == RouteProxy {
			var escalated []string
			for _, u := range work {
				_, done := outcomes[u]
				_, failed := legErrors[u]
				inLeg := containsString(legURLs, u)
				if !done && failed && !inLeg {
					escalated = append(escalated, u)
				}
			}
			if len(escalated) > 0 {
				log(fmt.Sprintf("🤖 %d link(s) that failed the direct leg "+
					"are retried through your proxy now (the same delivery, "+
					"the other route — the owner's law: some sites need the "+
					"proxy, some are broken by it)", len(escalated)), "info")
				legURLs = append(append([]string{}, legURLs...), escalated...)
			}
		}
		if len(legURLs) == 0 {
			continue // the landing pad stayed empty — nothing to ride
		}
		flags := chromeProxyFlagOrDirect(leg.Route, opts.Proxy)
		if leg.Route == RouteProxy && len(flags) == 0 {
			continue // nothing to turn on — already logged
		}
		if leg.Route == RouteDirect {
			log(fmt.Sprintf("🤖 Chrome leg: DIRECT (no proxy — the local "+
				"line) for %d link(s)", len(legURLs)), "info")
		} else {
			log(fmt.Sprintf("🤖 Chrome leg: through YOUR PROXY (%s) for %d "+
				"link(s)", ProxyLabelForLog(opts.Proxy), len(legURLs)), "info")
		}
		session, err := safeLaunchIdentity(chrome, log, pageTimeoutS,
			opts.ProfileDir, opts.Persist, opts.RealProfile, flags,
			opts.SessionFactory)
		if err != nil {
			log(fmt.Sprintf("🤖 Fifth door could not open a Chrome session: "+
				"%s", err.Error()), "warning")
			for _, u := range legURLs {
				legErrors[u] = append(legErrors[u], err.Error())
			}
			continue
		}
		if session == nil {
			for _, u := range legURLs {
				legErrors[u] = append(legErrors[u],
					"no Chrome session could be opened")
			}
			continue
		}
		pages := gatherThroughSession(session, legURLs, pageTimeoutS, wave,
			opts.ShouldContinue, log)
		func() {
			defer func() { recover() }() // the reference's belt-and-braces
			session.Close()
		}()
		for _, page := range pages {
			u := page.URL
			if page.OK {
				outcomes[u] = page
				if opts.MemoryPath != "" {
					RememberRoute(memory, u, leg.Route)
					memoryDirty = true
				}
			} else {
				errStr := strings.TrimSpace(page.Error)
				if errStr == "" {
					errStr = "no content"
				}
				prefix := "proxy leg: "
				if leg.Route == RouteDirect {
					prefix = "direct leg: "
				}
				legErrors[u] = append(legErrors[u], prefix+errStr)
			}
		}
		if opts.ShouldContinue != nil && !opts.ShouldContinue() {
			stopped = true
			break
		}
	}
	if memoryDirty {
		SaveProxyMemory(opts.MemoryPath, memory)
	}
	// the final story, in work order: the leg that delivered, or every
	// leg's honest sentence together
	for _, u := range work {
		if page, ok := outcomes[u]; ok {
			results = append(results, page)
			continue
		}
		if errs := legErrors[u]; len(errs) > 0 {
			results = append(results, PageResult{URL: u, OK: false,
				Error: strings.Join(errs, " | ")})
		} else {
			errStr := "the link was not tried this run"
			if stopped {
				errStr = "stopped by the user"
			}
			results = append(results, PageResult{URL: u, OK: false,
				Error: errStr})
		}
	}
	return results
}

// safeLaunchIdentity wraps launchIdentitySessionFn with the
// reference's try/except (a raised constructor error is caught by the
// driver and named per-url).
func safeLaunchIdentity(exe string, log LogFn, pageTimeoutS float64,
	profileDir string, persist, realProfile bool, flags []string,
	factory func(exe string, log LogFn) (TabSession, error)) (session TabSession, err error) {
	defer func() {
		if r := recover(); r != nil {
			session, err = nil, newCDPError("%v", r)
		}
	}()
	return launchIdentitySessionFn(exe, log, pageTimeoutS, profileDir,
		persist, realProfile, flags, factory)
}

func chromeProxyFlagOrDirect(route string, proxy *ProxyConfig) []string {
	if route == RouteDirect {
		return directLegFlags
	}
	return ChromeProxyFlag(proxy)
}

func containsString(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

// proxyFromConfigSafe mirrors the reference's try/except around
// proxy_from_config (an unreadable proxy block rides no leg).
func proxyFromConfigSafe(config map[string]any) (proxy *webfetch.Proxy) {
	defer func() {
		if r := recover(); r != nil {
			proxy = nil
		}
	}()
	return webfetch.ProxyFromConfig(config)
}
