package cli

// fetchqueue.go — T-62 slice 2: the Worker-queue fetch + the
// _bot_queue_job CLASSIFICATION half (worker_jobs.py 93-385), ported
// as its own coherent piece.
//
// THE OPTION-B SHAPE (ADR-0004 Am.3): the reference's fetch was the
// Telethon bot-chat read; the Go port's queue source is the Worker's
// GET /api/pending (the honest poll — workerclient.GetPendingResult).
// The CLASSIFICATION half runs identically over whatever links the
// fetch answers: the local ground truth (the VaultIndex, the dead set,
// THE SETTLEMENT's settled ledger, the platform-half ban list) decides
// what is actually pending — the Worker's ledger and the desktop's
// vault reconcile here, before any API call or vault write.
//
// The websites classification (v0.24.1 — "pending covers BOTH
// pipelines") is the exact per-URL ladder the reference's worker job
// runs, over the websites VaultIndex + the WebsiteStateDB's dedupe
// tables (dismissed → in-vault (failed placeholder with retries left
// → pending, settled → settled) → processed → settled → pending).

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// QueueFetchResult mirrors the _bot_queue_job result dict the drivers
// consume (plus the Worker poll's own state word).
type QueueFetchResult struct {
	State string // ok | disabled | error
	Error string

	// The fetch's own answer (the Worker's pending items, routed by the
	// batch's own URL grammar — the sync display's classify law).
	GitHubURLs []string
	NonGithub  []string

	// The GitHub classification (worker_jobs.py 168-230).
	SettledList       map[string]bool
	PendingURLs       []string
	InVaultCount      int
	Decommissioned    int
	BlockedCount      int
	SelfCount         int
	ReposSettled      int
	VaultIndexCount   *int
	DuplicatesRemoved int
	RawURLCount       int

	// The websites classification (worker_jobs.py 254-385).
	PendingWebsiteURLs   []string
	WebsitesInVault      int
	WebsitesProcessed    int
	WebsitesDismissed    int
	WebsitesBlocked      int
	WebsitesSelf         int
	WebsitesSettled      int
	WebsitesVaultIndex   *int
	WebsitesPipelineOff  bool
	WebsitesNoVault      bool
	WebsitesVaultMissing bool
}

// itemURL mirrors _item_url (sync_display.py 119-128): the URL a
// pending row carries — url_normalized, url_original, or url; only a
// string URL speaks, anything else is unrouted.
func itemURL(item any) string {
	m, ok := item.(map[string]any)
	if !ok {
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
			return ""
		}
		return ""
	}
	for _, key := range []string{"url_normalized", "url_original", "url"} {
		if raw, present := m[key]; present {
			if s, isStr := raw.(string); isStr && strings.TrimSpace(s) != "" {
				return strings.TrimSpace(s)
			}
		}
	}
	return ""
}

// routeQueueItems splits the pending items into the GitHub and website
// lists (classify_pending's grammar: a github.com/<owner>/<repo> link
// or an <owner>.github.io/<repo> page is a repo; every other link is a
// website).
func routeQueueItems(items []any) (githubURLs, nonGithub []string, rawCount int) {
	rawCount = len(items)
	for _, it := range items {
		u := itemURL(it)
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

// FetchWorkerQueue mirrors fetch_bot_queue's shape (the spinner-wrapped
// fetch): the honest poll over the WorkerFetch door. The state word
// answers the caller's failure card; "ok" carries the classified
// result.
func (c *CLI) FetchWorkerQueue(cfg *storage.OrderedMap) *QueueFetchResult {
	status := c.statusLineDoor()
	status.Start("Fetching bot queue", 0)
	defer status.Stop("")

	state, items := func() (string, []any) {
		if c.WorkerFetch != nil {
			return c.WorkerFetch(cfg)
		}
		return defaultWorkerFetch(cfg)
	}()

	if state != "ok" {
		out := &QueueFetchResult{State: state}
		if state == "error" {
			out.Error = "the Worker queue poll failed (network down or the Worker answered an error)"
		} else {
			out.Error = "the Worker queue is not configured — set the Worker URL + pair the app in Settings → Bot (the pairing generates the install_id + shared_secret the HMAC signature requires)"
		}
		return out
	}

	out := c.ClassifyQueueLinks(cfg, items)
	out.State = "ok"
	return out
}

// defaultWorkerFetch is the production /api/pending poll (the
// workerclient.Sync honest poll).
func defaultWorkerFetch(cfg *storage.OrderedMap) (string, []any) {
	s := workerclient.NewSync(cfgPlainMap(cfg))
	return s.GetPendingResult()
}

// ClassifyQueueLinks is the _bot_queue_job classification half (its
// own coherent piece — the vault filter + the settlement + the
// websites ladder), driven over the fetched items.
func (c *CLI) ClassifyQueueLinks(cfg *storage.OrderedMap,
	items []any) *QueueFetchResult {

	out := &QueueFetchResult{}
	githubURLs, nonGithub, rawCount := routeQueueItems(items)
	out.GitHubURLs = githubURLs
	out.NonGithub = nonGithub
	out.RawURLCount = rawCount

	cfgPlain := cfgPlainMap(cfg)
	vaultPath := cfgOrStr(cfg, "vault_path")
	blockedDomains := links.PlatformDomainsFromConfig(cfgPlain)
	selfDomains := links.SelfDomainsFromConfig(cfgPlain)
	websiteBlocked := links.BlockedDomainsFromConfig(cfgPlain)
	websiteVault := strings.TrimSpace(cfgOrStr(cfg, "website_vault_path"))
	pipes := getMap(cfg, "pipelines")
	websitesOn := pyTruthy(getOr(pipes, "websites", false))

	// --- The GitHub half (vault_path given + a successful fetch). ---
	if vaultPath != "" {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// the reference's except: show all URLs
					c.CliPrint(fmt.Sprintf(
						"⚠️ Vault filter failed in worker (%s); showing all URLs.",
						pyCrashText(r)), "warning")
					out.PendingURLs = append([]string(nil), githubURLs...)
					out.InVaultCount = 0
					out.Decommissioned = 0
				}
			}()
			vi := batch.NewVaultIndex(vaultPath, nil)
			vi.Rebuild(nil)
			count := vi.Count()
			out.VaultIndexCount = &count

			decommissioned := map[string]bool{}
			cache, cerr := store.NewCacheDB(cacheDBPath())
			if cerr == nil {
				defer func() { _ = cache.Close() }()
				decommissioned = cache.GetDeadURLSet(store.DeadLinkThreshold)
				// v0.63.3 — THE SETTLEMENT at the queue's own door.
				settlement, _ := cache.SettleExistingRepos(githubURLs)
				if !settlement.Already {
					c.CliPrint(fmt.Sprintf(
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
			out.InVaultCount = inVault
			out.Decommissioned = decomm
			out.BlockedCount = blocked
			out.SelfCount = selfC
			out.ReposSettled = settledc
		}()
	} else {
		out.PendingURLs = append([]string(nil), githubURLs...)
	}

	// --- The websites half (v0.24.1). ---
	if !websitesOn {
		out.PendingWebsiteURLs = nil
		if len(nonGithub) > 0 {
			out.WebsitesPipelineOff = true
		}
		return out
	}
	if websiteVault == "" {
		out.PendingWebsiteURLs = nil
		if len(nonGithub) > 0 {
			out.WebsitesNoVault = true
		}
		return out
	}
	if !c.IsDir(websiteVault) {
		// a configured-but-missing Websites vault path is a config error:
		// say so, and treat every non-GitHub link as pending.
		c.CliPrint(fmt.Sprintf(
			"⚠️ Websites vault not found at %s — treating every non-GitHub link "+
				"as pending. Fix the path in Settings → 📁 Vault.", websiteVault), "warning")
		out.PendingWebsiteURLs = append([]string(nil), nonGithub...)
		out.WebsitesVaultMissing = true
		return out
	}
	if len(nonGithub) > 0 {
		func() {
			defer func() {
				if r := recover(); r != nil {
					c.CliPrint(fmt.Sprintf(
						"⚠️ Websites vault filter failed in worker (%s) — treating "+
							"every non-GitHub link as a pending website.",
						pyCrashText(r)), "warning")
					out.PendingWebsiteURLs = append([]string(nil), nonGithub...)
				}
			}()
			wvi := batch.NewVaultIndex(websiteVault,
				func(u string) string { return links.NormalizeWebsiteURL(u) })
			wvi.Rebuild(nil)
			count := wvi.Count()
			out.WebsitesVaultIndex = &count
			state, serr := websitepipeline.NewWebsiteStateDB("cache.db")
			if serr != nil {
				c.CliPrint(fmt.Sprintf(
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
						c.CliPrintLog(), nonGithub)
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
			out.PendingWebsiteURLs = pendingWeb
			out.WebsitesInVault = webInVault
			out.WebsitesProcessed = webDone
			out.WebsitesDismissed = webDismissed
			out.WebsitesBlocked = webBlocked
			out.WebsitesSelf = webSelf
			out.WebsitesSettled = webSettled
		}()
	} else {
		out.PendingWebsiteURLs = nil
	}
	return out
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

// CliPrintLog adapts CliPrint to the LogFn shape.
func (c *CLI) CliPrintLog() func(string, string) {
	return func(msg, level string) { c.CliPrint(msg, level) }
}
