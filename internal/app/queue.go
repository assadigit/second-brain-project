// queue.go — T-71 slice (d): the _bot_queue_job classification half
// (worker_jobs.py:93-391's vault filter + the settlement + the
// websites ladder), composed over the same engine calls the CLI's
// ClassifyQueueLinks drives (each shell owns its composition; the
// engine laws live in the engine packages and are golden-tested
// there). The app's half speaks through the bus-oriented LogFn the
// command injects.
package app

import (
	"fmt"
	"os"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// cfgPlainMap flattens the OrderedMap config into the plain map the
// engine bridges read (nested dicts recursively — the CLI's own
// helper's shape).
func cfgPlainMap(cfg *storage.OrderedMap) map[string]any {
	out := map[string]any{}
	if cfg == nil {
		return out
	}
	for _, k := range cfg.Keys() {
		v, _ := cfg.Get(k)
		if om, ok := v.(*storage.OrderedMap); ok {
			out[k] = cfgPlainMap(om)
		} else {
			out[k] = v
		}
	}
	return out
}

// queueItemURL mirrors _item_url (sync_display.py:119-128): the URL a
// pending row carries — url_normalized, url_original, or url; only a
// string URL speaks.
func queueItemURL(item any) string {
	if m, ok := item.(map[string]any); ok {
		for _, key := range []string{"url_normalized", "url_original", "url"} {
			if raw, present := m[key]; present {
				if s, isStr := raw.(string); isStr && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
		return ""
	}
	if om, is := item.(*storage.OrderedMap); is {
		// The reference's FIXED key tuple speaks by priority
		// (sync_display.py:123) — the row's own key insertion
		// order never decides (the Worker row carries
		// forward_count/ledger_id/… alongside the URL keys; any
		// order is legal).
		for _, key := range []string{"url_normalized", "url_original", "url"} {
			if v, present := om.Get(key); present {
				if s, isStr := v.(string); isStr && strings.TrimSpace(s) != "" {
					return strings.TrimSpace(s)
				}
			}
		}
	}
	return ""
}

// routeQueueItems splits the pending items into the GitHub and
// website lists (classify_pending's grammar: a github.com/<owner>/
// <repo> link or an <owner>.github.io/<repo> page is a repo; every
// other link is a website).
func routeQueueItems(items []any) (githubURLs, nonGithub []string, rawCount int) {
	rawCount = len(items)
	for _, it := range items {
		u := queueItemURL(it)
		if u == "" {
			continue
		}
		if links.IsGitHubURL(u) || links.MapGitHubIOURL(u) != "" {
			githubURLs = append(githubURLs, u)
		} else {
			nonGithub = append(nonGithub, u)
		}
	}
	return githubURLs, nonGithub, rawCount
}

// ClassifyQueueLinks is the production QueueClassify seam: the
// classification over the fetched items, with every story line riding
// log (the command wires log → the bus).
func ClassifyQueueLinks(cfg *storage.OrderedMap, items []any, log LogFn) *QueueReport {
	if log == nil {
		log = func(string, string) {}
	}
	out := &QueueReport{}
	githubURLs, nonGithub, rawCount := routeQueueItems(items)
	out.GitHubURLs = githubURLs
	out.NonGithub = nonGithub
	out.RawCount = rawCount

	cfgPlain := cfgPlainMap(cfg)
	vaultPath := pyStrCfg(cfg, "vault_path")
	blockedDomains := links.PlatformDomainsFromConfig(cfgPlain)
	selfDomains := links.SelfDomainsFromConfig(cfgPlain)
	websiteBlocked := links.BlockedDomainsFromConfig(cfgPlain)
	websiteVault := strings.TrimSpace(pyStrCfg(cfg, "website_vault_path"))
	websitesOn := cfgNestedBool(cfg, "pipelines", "websites")

	// --- The GitHub half (vault_path given + a successful fetch). ---
	if vaultPath != "" {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// the reference's except: show all URLs
					log(fmt.Sprintf("⚠️ Vault filter failed in worker (%s); showing all URLs.", pyCrashText(r)), "warning")
					out.PendingURLs = append([]string(nil), githubURLs...)
				}
			}()
			vi := batch.NewVaultIndex(vaultPath, nil)
			vi.Rebuild(nil)
			count := vi.Count()
			out.VaultIndex = &count

			decommissioned := map[string]bool{}
			cache, cerr := store.NewCacheDB(cacheDBPathOf(cfg))
			if cerr == nil {
				defer func() { _ = cache.Close() }()
				decommissioned = cache.GetDeadURLSet(store.DeadLinkThreshold)
				// v0.63.3 — THE SETTLEMENT at the queue's own door.
				settlement, _ := cache.SettleExistingRepos(githubURLs)
				if !settlement.Already {
					log(fmt.Sprintf(
						"🤝 THE SETTLEMENT (repos): %d repo link(s) the bot already "+
							"delivered are settled — addressed and processed, never "+
							"fetched or counted again; %d queued retry(ies) cleared. "+
							"Only repos added from now on are counted (♻️ Reset 404 "+
							"Quarantine un-settles any of them)",
						settlement.Settled, settlement.RetriesCleared), "info")
				}
				out.SettledList = cache.GetSettledRepoSet()
			}
			pending, inVault, decomm, blocked, selfC, settledc := []string{}, 0, 0, 0, 0, 0
			for _, url := range githubURLs {
				norm := links.NormalizeURL(url)
				if len(blockedDomains) > 0 && links.DomainIsBlocked(url, blockedDomains) {
					blocked++
				} else if len(selfDomains) > 0 && links.DomainIsSelf(url, selfDomains) {
					selfC++
				} else if vi.HasURL(url) {
					inVault++
				} else if decommissioned[norm] {
					decomm++
				} else if out.SettledList != nil && out.SettledList[norm] {
					settledc++
				} else {
					pending = append(pending, url)
				}
			}
			out.PendingURLs = pending
			out.InVault = inVault
			out.Decommissioned = decomm
			out.Blocked = blocked
			out.Self = selfC
			out.Settled = settledc
		}()
	} else {
		out.PendingURLs = append([]string(nil), githubURLs...)
	}

	// --- The websites half (v0.24.1). ---
	if !websitesOn {
		out.PendingWebsites = nil
		if len(nonGithub) > 0 {
			out.WebsitesOff = true
		}
		return out
	}
	if websiteVault == "" {
		out.PendingWebsites = nil
		if len(nonGithub) > 0 {
			out.WebsitesNoVault = true
		}
		return out
	}
	if !fsIsDirPath(websiteVault) {
		log(fmt.Sprintf(
			"⚠️ Websites vault not found at %s — treating every non-GitHub link "+
				"as pending. Fix the path in Settings → 📁 Vault.", websiteVault), "warning")
		out.PendingWebsites = append([]string(nil), nonGithub...)
		out.WebsitesVaultMissing = true
		return out
	}
	if len(nonGithub) > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					log(fmt.Sprintf(
						"⚠️ Websites vault filter failed in worker (%s) — treating "+
							"every non-GitHub link as a pending website.",
						pyCrashText(r)), "warning")
					out.PendingWebsites = append([]string(nil), nonGithub...)
				}
			}()
			wvi := batch.NewVaultIndex(websiteVault,
				func(u string) string { return links.NormalizeWebsiteURL(u) })
			wvi.Rebuild(nil)
			wviCount := wvi.Count()
			out.WebsitesVaultIndex = &wviCount
			state, serr := websitepipeline.NewWebsiteStateDB(cacheDBPathOf(cfg))
			if serr != nil {
				log(fmt.Sprintf(
					"⚠️ Websites state DB unavailable (%s) — classifying by vault "+
						"index only.", pyErrString(serr)), "warning")
				state = nil
			}
			// v0.63.2/v0.64.1 — THE SETTLEMENT + the queue-history
			// settlement at the queue's own door.
			if state != nil {
				func() {
					defer func() { recover() }()
					websitepipeline.SettleTheLedger(state, websiteVault,
						func(msg, level string) { log(msg, level) }, nonGithub)
				}()
			}
			pendingWeb, webInVault, webDone, webDismissed := []string{}, 0, 0, 0
			webBlocked, webSelf, webSettled := 0, 0, 0
			for _, url := range nonGithub {
				canonical := links.NormalizeWebsiteURL(url)
				if len(websiteBlocked) > 0 && links.DomainIsBlocked(url, websiteBlocked) {
					webBlocked++
				} else if len(selfDomains) > 0 && links.DomainIsSelf(url, selfDomains) {
					webSelf++
				} else if state != nil && websiteIsDismissed(state, canonical) {
					webDismissed++
				} else if wvi.HasURL(url) {
					prior := websiteProcessedRow(state, canonical)
					priorFailed := prior != nil && prior.FetchStatus != nil &&
						*prior.FetchStatus == "failed"
					if priorFailed && state != nil {
						if websiteIsSettled(state, canonical) {
							webSettled++
						} else {
							pendingWeb = append(pendingWeb, url)
						}
					} else {
						webInVault++
					}
				} else if state != nil && websiteIsProcessed(state, canonical) {
					webDone++
				} else if state != nil && websiteIsSettled(state, canonical) {
					webSettled++
				} else {
					pendingWeb = append(pendingWeb, url)
				}
			}
			if state != nil {
				func() {
					defer func() { recover() }()
					state.Close()
				}()
			}
			out.PendingWebsites = pendingWeb
			out.WebsitesInVault = webInVault
			out.WebsitesProcessed = webDone
			out.WebsitesDismissed = webDismissed
			out.WebsitesBlocked = webBlocked
			out.WebsitesSelf = webSelf
			out.WebsitesSettled = webSettled
		}()
	} else {
		out.PendingWebsites = nil
	}
	return out
}

// cacheDBPathOf mirrors the CLI's cacheDBPath (APP_DIR/cache.db — the
// v0.09.4 split-state law: never CWD-relative).
func cacheDBPathOf(cfg *storage.OrderedMap) string {
	return constants.AppDir() + string(os.PathSeparator) + "cache.db"
}

// cfgNestedBool reads a nested boolean (getOr + pyTruthy's law).
func cfgNestedBool(cfg *storage.OrderedMap, section, key string) bool {
	if cfg == nil {
		return false
	}
	v, _ := cfg.Get(section)
	om, ok := v.(*storage.OrderedMap)
	if !ok {
		return false
	}
	sv, _ := om.Get(key)
	switch t := sv.(type) {
	case bool:
		return t
	case string:
		return t == "1" || strings.EqualFold(t, "true") || strings.EqualFold(t, "yes")
	case float64:
		return t != 0
	case int:
		return t != 0
	}
	return false
}

// pyErrString mirrors str(exc) — the error's message.
func pyErrString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

// pyErrorClass mirrors type(exc).__name__ — the class-name token law.
func pyErrorClass(err error) string {
	if err == nil {
		return ""
	}
	name := fmt.Sprintf("%T", err)
	name = strings.TrimPrefix(name, "*")
	if i := strings.LastIndexByte(name, '.'); i >= 0 {
		name = name[i+1:]
	}
	return name
}

// pyCrashText renders a panic value the way the reference speaks a
// crashed exception ("{type}: {e}" — the SHAPE is the contract).
func pyCrashText(r any) string {
	if err, isErr := r.(error); isErr {
		return pyErrorClass(err) + ": " + pyErrString(err)
	}
	return fmt.Sprintf("%v", r)
}

// fsIsDirPath is the IsDir probe (guarded — trouble reads as false).
func fsIsDirPath(p string) bool {
	fi, err := os.Stat(p)
	return err == nil && fi.IsDir()
}

// The state probes (guarded — a state error reads as the no-answer).
func websiteIsDismissed(state *websitepipeline.WebsiteStateDB, canonical string) bool {
	defer func() {
		if recover() != nil {
			// the probe reads false on trouble
		}
	}()
	v, err := state.IsDismissed(canonical)
	return err == nil && v
}

func websiteIsSettled(state *websitepipeline.WebsiteStateDB, canonical string) bool {
	defer func() { recover() }()
	v, err := state.IsSettled(canonical)
	return err == nil && v
}

func websiteIsProcessed(state *websitepipeline.WebsiteStateDB, canonical string) bool {
	defer func() { recover() }()
	v, err := state.IsProcessed(canonical)
	return err == nil && v
}

func websiteProcessedRow(state *websitepipeline.WebsiteStateDB,
	canonical string) *websitepipeline.ProcessedRow {
	defer func() { recover() }()
	row, err := state.ProcessedRow(canonical)
	if err != nil {
		return nil
	}
	return row
}
