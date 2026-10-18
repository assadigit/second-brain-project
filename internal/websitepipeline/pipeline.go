package websitepipeline

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/directorysync"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/taxonomy"
	"github.com/assadigit/second-brain-project/internal/webfetch"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// pipeline.go — the WebsitePipeline class (website_pipeline.py:3336-
// 5338): the per-link state machine on top of the T-23 state layer,
// the constructor's gate sequence (THE CONFIRMATION GATE → the
// graveyard consume → THE BANISHMENT → the twins sweep → THE TALLY),
// and the four batch drivers.

// LLMCallFn mirrors the injected llm_call(messages, task=None) -> str
// (the worker routes to Ollama or an OpenAI-compatible endpoint; the
// golden runner injects a fake). task is 'classify' for the w01/w02
// passes and 'analyze' for w03.
type LLMCallFn func(messages []llmclient.Message, task string) string

// VaultIndexHasFn mirrors vault_index_has(url) -> bool | path (the
// websites VaultIndex probe — the ground-truth dedupe layer).
type VaultIndexHasFn func(url string) any

// SiteNoteForFn mirrors site_note_for(url) -> path | None (v0.64.0 —
// ONE NOTE PER SITE's probe). nil leaves the law off: the pipeline
// behaves exactly as before, one note per link.
type SiteNoteForFn func(url string) string

// FetchFn mirrors the optional fetch_fn injection (tests + the offline
// golden run stub this so NO network is touched; production leaves it
// nil → webfetch.FetchURL).
type FetchFn func(rawURL string, opts *webfetch.Options) *webfetch.FetchResult

// BanishConfirmResult mirrors the banish_confirm channel's answer:
// {'verdict': ..., 'report': callable} — or a bare verdict string.
type BanishConfirmResult struct {
	Verdict string
	// Report is the confirmed ask's closing line (the count that
	// ACTUALLY left) — the Telegram round-trip's own edit.
	Report func(n int)
}

// BanishConfirmFn mirrors banish_confirm(items, log) — the injected
// confirmation channel (the Telegram round-trip the GUI worker and the
// CLI both wire in). nil = no channel (the safe default: defer).
type BanishConfirmFn func(items []PendingBanishment, log LogFn) *BanishConfirmResult

// BanishGate mirrors self.banish_gate: {'verdict', 'pending',
// 'banished'}.
type BanishGate struct {
	Verdict  string
	Pending  int
	Banished int
}

// Result mirrors process_link's result dict.
type Result struct {
	URL         string
	Canonical   string
	Outcome     string // processed | review | skipped | failed
	NotePath    string
	Category    string
	Subcategory string
	FetchStatus string
	Error       string
}

// ShouldContinueFn mirrors should_continue (the GUI Stop button's
// poll); OnProgressFn mirrors on_progress(url).
type ShouldContinueFn func() bool
type OnProgressFn func(url string)

// Pipeline mirrors WebsitePipeline: one instance per batch.
type Pipeline struct {
	config         map[string]any
	llmCall        LLMCallFn
	vaultIndexHas  VaultIndexHasFn
	state          *WebsiteStateDB
	log            LogFn
	noteStateDB    *notestate.DB
	siteNoteFor    SiteNoteForFn
	fetchFn        FetchFn
	tax            *taxonomy.Taxonomy
	taxonomyPath   string
	vaultPath      string
	vaultName      string
	graveyardUrls  map[string]bool
	banishedURLs   []string
	banishGate     BanishGate
	blockedDomains []string
	selfDomains    []string
	banishedCache  *directoryBanishedCache

	fetchTimeoutS       float64
	fetchMaxBytes       int
	impersonateFallback bool
	impersonateAutopip  bool
	archiveFallback     bool
	dohProbe            bool
	handDelivery        bool
	rateLimiter         *webfetch.DomainRateLimiter

	counters       map[string]int
	lastResults    []Result
	batchSiteNotes map[string]string

	// the hand probes' caches (mtime+size keyed)
	handTableKey   *tableCacheKey
	handTableRows  []decomtable.Row
	handReviewKey  *reviewCacheKey
	handReviewRows []ReviewNoteItem
}

type tableCacheKey struct {
	path  string
	mtime int64
	size  int64
}

type reviewCacheKey struct {
	mtime int64
	size  int64
}

// directoryBanishedCache mirrors self._banished_cache (v0.76.0 — the
// directory's banish gate state, SPEC §11.1): loaded lazily on first
// use, reloaded at most every 60s.
type directoryBanishedCache struct {
	loadedAt time.Time
	banished []directorysync.BanishEntry
	shared   []string
}

// NewPipeline mirrors WebsitePipeline.__init__ — the full gate
// sequence (the banishment gate → the graveyard consume → the marked
// notes' banishment → the twins sweep → the tally) runs HERE, before
// anything fetches. Bookkeeping never kills a batch.
func NewPipeline(config map[string]any, llmCall LLMCallFn,
	vaultIndexHas VaultIndexHasFn, state *WebsiteStateDB, log LogFn,
	tax *taxonomy.Taxonomy, noteStateDB *notestate.DB,
	rateLimiter *webfetch.DomainRateLimiter, fetchFn FetchFn,
	banishConfirm BanishConfirmFn, siteNoteFor SiteNoteForFn) (*Pipeline, error) {
	p := &Pipeline{
		config:        config,
		llmCall:       llmCall,
		vaultIndexHas: vaultIndexHas,
		state:         state,
		log:           log,
		noteStateDB:   noteStateDB,
		siteNoteFor:   siteNoteFor,
		graveyardUrls: map[string]bool{},
		counters: map[string]int{
			"processed": 0, "review": 0, "skipped": 0, "retried": 0,
			"failed": 0, "upgraded": 0, "consolidated": 0,
		},
		batchSiteNotes: map[string]string{},
	}
	if p.log == nil {
		p.log = noopLog
	}
	if p.config == nil {
		p.config = map[string]any{}
	}
	// v0.64.0 — ONE NOTE PER SITE's probe (nil = the law stays off).
	if fetchFn != nil {
		p.fetchFn = fetchFn
	}
	var proxyArmed bool
	if fetchFn == nil {
		p.fetchFn = func(rawURL string, opts *webfetch.Options) *webfetch.FetchResult {
			return webfetch.FetchURL(rawURL, opts)
		}
		// v0.19.0 — web fetches through the proxy (Settings → Proxy):
		// the blocked-web fix. NEVER applied to an injected fetch_fn
		// (the golden run must stay offline and tests hermetic).
		proxy := webfetch.ProxyFromConfig(p.config)
		if proxy != nil {
			if ok, why := webfetch.WebProxyPreflight(proxy, 5.0); ok {
				proxyArmed = true
				p.log(fmt.Sprintf("🌐 Web fetches via %s proxy (Settings → 🌐 Proxy) — v0.43.0 both doors: the direct line is still tried when the proxy path fails or is walled (403/405/429/451)", webfetch.ProxyLabel(proxy)), "info")
				base := p.fetchFn
				p.fetchFn = func(rawURL string, opts *webfetch.Options) *webfetch.FetchResult {
					if opts == nil {
						opts = &webfetch.Options{}
					}
					if opts.Proxy == nil {
						opts.Proxy = proxy
					}
					return base(rawURL, opts)
				}
			} else {
				p.log(fmt.Sprintf("⚠️ Web proxy %s NOT reachable — fetching DIRECT. %s x.com / YouTube links will keep failing until the proxy client is up.", webfetch.ProxyLabel(proxy), why), "warning")
			}
		}
	}
	if proxyArmed {
		p.maybeRearmRetries()
	}
	// v0.41.0 — an owner-set User-Agent (config.json "web_user_agent");
	// like the proxy wrap, NEVER touches an injected fetch_fn.
	customUA := strings.TrimSpace(cfgStr(p.config, "web_user_agent"))
	if customUA != "" && fetchFn == nil {
		base := p.fetchFn
		p.fetchFn = func(rawURL string, opts *webfetch.Options) *webfetch.FetchResult {
			if opts == nil {
				opts = &webfetch.Options{}
			}
			if opts.UserAgent == "" {
				opts.UserAgent = customUA
			}
			return base(rawURL, opts)
		}
	}
	if tax != nil {
		p.tax = tax
	} else {
		loaded, err := taxonomy.LoadTaxonomyFromConfig(p.config)
		if err != nil {
			return nil, err
		}
		p.tax = loaded
	}
	p.vaultPath = strings.TrimSpace(cfgStr(p.config, "website_vault_path"))
	// v0.35.0 — the display name of the Websites vault for the
	// per-item destination logs.
	p.vaultName = "Websites vault"
	if p.vaultPath != "" {
		if base := filepath.Base(p.vaultPath); base != "" && base != "." && base != string(filepath.Separator) {
			p.vaultName = base
		}
	}

	// ---- v0.60.0 THE CONFIRMATION GATE: the run opens with the vault
	// scan, the count is SPOKEN and asked over the injected channel,
	// and NOTHING is removed until the owner answers.
	var banishedRun []string
	gateVerdict := "auto"
	gatePending := 0
	var gateReport func(n int)
	if p.vaultPath != "" && isDirPath(p.vaultPath) {
		gateScan := PendingBanishmentsReport{}
		func() {
			defer func() { recover() }()
			gateScan = ScanPendingBanishments(p.vaultPath, p.log, nil)
		}()
		gatePending = len(gateScan.Items)
		if gatePending > 0 {
			// the number goes FIRST — "so I ensure that system
			// successfully detected them".
			p.log(fmt.Sprintf("%s %d note(s) marked for deletion (🗑️ / delete / auto_delete) — asking you to confirm before anything is removed", BanishGatePrefix, gatePending), "info")
			if banishConfirm != nil {
				var gateRes *BanishConfirmResult
				func() {
					defer func() {
						if r := recover(); r != nil {
							p.log(fmt.Sprintf("⚠️ Banish confirmation failed: %v — nothing deleted, the marks stay", r), "warning")
						}
					}()
					gateRes = banishConfirm(gateScan.Items, p.log)
				}()
				if gateRes == nil {
					gateVerdict = "defer"
				} else {
					gateVerdict = gateRes.Verdict
					if gateVerdict == "" {
						gateVerdict = "defer"
					}
					gateReport = gateRes.Report
				}
			} else {
				optOut := strings.ToLower(strings.TrimSpace(cfgStr(p.config, "banish_confirm")))
				// config banish_confirm defaults TRUE; the explicit
				// opt-out values restore the v0.58/v0.59 auto-delete.
				if optOut == "false" || optOut == "0" || optOut == "no" || optOut == "off" {
					gateVerdict = "auto"
				} else {
					// no channel injected and no opt-out — the SAFE
					// default: the marks stay, the next run asks again.
					gateVerdict = "defer"
					p.log(fmt.Sprintf("⚠️ %d marked note(s) found, but no confirmation channel is reachable (the Telegram worker is not paired/enabled) — nothing deleted; the marks stay, I'll ask again next run (config \"banish_confirm\": false restores the old auto-delete)", gatePending), "warning")
				}
			}
			switch gateVerdict {
			case "confirmed":
				p.log(fmt.Sprintf("✅ You confirmed the deletion — the %d marked note(s) leave the vault now (twins in the other vault too, if any); GitHub drops them on the next seal", gatePending), "info")
			case "declined":
				p.log(fmt.Sprintf("👌 You kept them — the %d marked note(s) stay in the vault; the marks stay too, I'll ask again next run", gatePending), "info")
			case "timeout":
				p.log(fmt.Sprintf("⌛ No answer in time — the %d marked note(s) stay; the marks stay, I'll ask again next run", gatePending), "info")
			}
		}
	}
	// ---- v0.44.0 the graveyard: consume the owner's burial decisions
	// BEFORE anything fetches. v0.60.0 — the table's own 🗑 pass 0 rides
	// THE GATE (dead / reviewed / hand verdicts run unchanged; the
	// DESTRUCTIVE door waits for the confirmation).
	if p.vaultPath != "" && isDirPath(p.vaultPath) {
		func() {
			defer func() { recover() }()
			consumeReport := ConsumeDecommissionTable(
				p.state, p.vaultPath, p.log,
				gateVerdict == "auto" || gateVerdict == "confirmed", nil)
			banishedRun = append(banishedRun, consumeReport.BanishedURLs...)
			for u, st := range ScanDecommissionTable(p.vaultPath) {
				if decomtable.StatusIsDead(st) {
					p.graveyardUrls[links.NormalizeWebsiteURL(u)] = true
				}
			}
		}()
	}
	// ---- v0.58.0 THE BANISHMENT: the owner's delete verdicts written
	// on the notes themselves — enforced BEFORE anything fetches. THE
	// GATE owns this door now: the pass runs only when the owner
	// confirmed (or opted into auto).
	if (gateVerdict == "auto" || gateVerdict == "confirmed") &&
		p.vaultPath != "" && isDirPath(p.vaultPath) {
		func() {
			defer func() { recover() }()
			noteReport := BanishMarkedNotes(p.state, p.vaultPath, p.log, nil)
			banishedRun = append(banishedRun, noteReport.URLs...)
		}()
		// v0.60.0 — the "both vaults" promise: sweep the OTHER vault
		// (config vault_path — the GitHub-projects vault) for twins.
		if len(banishedRun) > 0 {
			other := strings.TrimSpace(cfgStr(p.config, "vault_path"))
			if other != "" && isDirPath(other) && samePath(other, p.vaultPath) == false {
				func() {
					defer func() { recover() }()
					if twins := BanishTwinsInOtherVault(other, banishedRun, p.log); twins > 0 {
						p.log(fmt.Sprintf("🗑️ Both vaults: %d twin note(s) removed from the other vault; GitHub drops them on the next seal", twins), "info")
					}
				}()
			}
		}
	}
	// v0.59.0 — THE TALLY, spoken every run (the zero run answers too).
	p.banishedURLs = dedupeKeepOrder(banishedRun)
	p.banishGate = BanishGate{Verdict: gateVerdict, Pending: gatePending, Banished: len(p.banishedURLs)}
	if len(p.banishedURLs) > 0 {
		p.log(fmt.Sprintf("%s %d website(s) removed this run and never fetched again — you marked them for deletion (🗑️ / delete / auto_delete); the notes rest in %s, the URLs are blacklisted, and ♻️ revived on the record row undoes any of them", BanishTallyPrefix, len(p.banishedURLs), BanishQuarantineRelpath), "info")
	} else if gatePending > 0 && gateVerdict != "auto" {
		why := gateVerdict
		switch gateVerdict {
		case "declined":
			why = "you said keep"
		case "timeout":
			why = "no answer in time"
		case "defer":
			why = "no confirmation channel"
		}
		p.log(fmt.Sprintf("%s 0 websites removed this run — %d marked note(s) KEPT (%s; the marks stay, I'll ask again next run)", BanishTallyPrefix, gatePending, why), "info")
	} else {
		p.log(fmt.Sprintf("%s 0 websites removed this run — no 🗑️ / delete / auto_delete marks in the library", BanishTallyPrefix), "info")
	}
	// the gate's answer to Telegram: the confirmed ask gets its closing
	// line (the count that ACTUALLY left).
	if gateVerdict == "confirmed" && gateReport != nil {
		func() {
			defer func() { recover() }()
			gateReport(len(p.banishedURLs))
		}()
	}
	p.taxonomyPath = resolveTaxonomyPath(p.config)
	// v0.20.0 — blocked domains (the X fix): never fetched, never
	// noted, never retried.
	p.blockedDomains = links.BlockedDomainsFromConfig(p.config)
	// v0.21.0 — self domains (the app's own bot).
	p.selfDomains = links.SelfDomainsFromConfig(p.config)
	if (len(p.blockedDomains) > 0 || len(p.selfDomains) > 0) && fetchFn == nil {
		// Production only (an injected fetch_fn = the hermetic golden
		// run / tests). Purge queued retries + _review placeholders
		// ONCE so previously-queued blocked links stop coming back.
		p.enforceBlockedDomains()
		// v0.28.0 — THE LAW, disk edition: any app-owned note whose
		// source is on a banned domain leaves the library for
		// .trash/banned-domains (website_directory's sweep — T-41's
		// directorynote.go; the seam is gone, the sweep is real).
		// Runs on every batch; idempotent (the first run cleans
		// history, later runs are no-ops).
		if p.vaultPath != "" && isDirPath(p.vaultPath) {
			p.sweepBannedNotesGuarded()
		}
	}
	// Fetch politeness knobs are config-overridable.
	p.fetchTimeoutS = pyFloatOr(p.config, "web_fetch_timeout_s", FetchTimeoutS)
	p.fetchMaxBytes = int(pyFloatOr(p.config, "web_fetch_max_bytes", float64(FetchMaxBytes)))
	// v0.45.0 — THE THIRD DOOR: in the Go build the Chrome TLS door
	// (tls-client) is ALWAYS compiled in (ADR-0006 §4) — the reference's
	// curl_cffi_available()/autopip ladder collapses to "armed" (the
	// pip-install leg is obsolete by construction; documented in
	// doc.go).
	p.impersonateFallback = !cfgIsFalse(p.config, "web_impersonate_fallback")
	p.impersonateAutopip = !cfgIsFalse(p.config, "web_impersonate_autopip")
	if fetchFn == nil && p.impersonateFallback {
		p.log("🤝 v0.45.0 third door armed: a refusal-family wall on every door is re-asked with Chrome's own TLS handshake (curl_cffi) — the 'bot defense (Cloudflare: challenge)' class", "info")
	}
	p.archiveFallback = !cfgIsFalse(p.config, "web_archive_fallback")
	p.dohProbe = !cfgIsFalse(p.config, "web_doh_probe")
	if fetchFn == nil && (p.archiveFallback || p.dohProbe) {
		var rungs []string
		if p.archiveFallback {
			rungs = append(rungs, "dead pages climb variants then the Wayback Machine")
		}
		if p.dohProbe {
			rungs = append(rungs, "DNS failures get the DNS-over-HTTPS verdict")
		}
		p.log("🪜 v0.46.0 ladder armed: "+strings.Join(rungs, "; ")+
			"; 429/503 pay Retry-After into the domain limiter", "info")
	}
	// v0.48.0 — the fourth door (the owner's real Chrome).
	p.handDelivery = !cfgIsFalse(p.config, "web_hand_delivery")
	if rateLimiter != nil {
		p.rateLimiter = rateLimiter
	} else {
		delay := pyFloatOr(p.config, "web_domain_delay_s", DomainDelayS)
		p.rateLimiter = webfetch.NewDomainRateLimiter(delay)
	}
	return p, nil
}

// cfgStr reads a config string (Python str(v or ”).strip()).
func cfgStr(config map[string]any, key string) string {
	v, ok := config[key]
	if !ok || v == nil {
		return ""
	}
	return strings.TrimSpace(pyString(v))
}

// cfgIsFalse mirrors config.get(k) is not False (the Go view: an
// explicit false-y value opts out; anything else stays ON).
func cfgIsFalse(config map[string]any, key string) bool {
	v, ok := config[key]
	if !ok || v == nil {
		return false
	}
	if b, ok := v.(bool); ok {
		return !b
	}
	return false
}

// pyFloatOr reads a config float with Python's truthiness default.
func pyFloatOr(config map[string]any, key string, def float64) float64 {
	v, ok := config[key]
	if !ok || v == nil {
		return def
	}
	switch t := v.(type) {
	case float64:
		if t == 0 {
			return def
		}
		return t
	case int:
		if t == 0 {
			return def
		}
		return float64(t)
	case int64:
		if t == 0 {
			return def
		}
		return float64(t)
	case string:
		if strings.TrimSpace(t) == "" {
			return def
		}
		return def
	}
	return def
}

// pyString stringifies config values the Python str() way (floats
// keep their ".0").
func pyString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		return pyFloatStr(t)
	case bool:
		if t {
			return "True"
		}
		return "False"
	case nil:
		return "None"
	}
	return fmt.Sprintf("%v", v)
}

// pyFloatStr mirrors Python repr for floats (integral keeps ".0").
func pyFloatStr(f float64) string {
	if f == float64(int64(f)) && f < 1e16 && f > -1e16 {
		return fmt.Sprintf("%d.0", int64(f))
	}
	return fmt.Sprintf("%v", f)
}

// resolveTaxonomyPath mirrors constants.resolve_taxonomy_path.
func resolveTaxonomyPath(config map[string]any) string {
	if v := cfgStr(config, "taxonomy_path"); v != "" {
		return v
	}
	return "taxonomy/website-library-categories.md"
}

// sweepBannedNotesGuarded runs the v0.28.0 disk sweep wrapped in the
// reference's blanket except ("⚠️ Banned-domain sweep skipped: …") —
// bookkeeping never kills a batch. The pinned shape is the fallback
// (a warning line, the batch continues), never the rendered exception
// text (the package's standing divergence record: Go renders the
// recovered value, Python renders the exception class + message).
func (p *Pipeline) sweepBannedNotesGuarded() {
	defer func() {
		if r := recover(); r != nil {
			p.log(fmt.Sprintf("⚠️ Banned-domain sweep skipped: %v", r), "warning")
		}
	}()
	SweepBannedNotes(p.vaultPath, p.blockedDomains, p.log)
}

// BanishedURLs / BanishGate expose the run's verdict + tally.
func (p *Pipeline) BanishedURLs() []string { return p.banishedURLs }
func (p *Pipeline) Gate() BanishGate       { return p.banishGate }

// Counters exposes the per-batch counters map.
func (p *Pipeline) Counters() map[string]int { return p.counters }

// LastResults exposes the run's per-link results.
func (p *Pipeline) LastResults() []Result { return p.lastResults }

// maybeRearmRetries mirrors _maybe_rearm_retries (v0.19.0): when the
// ACTIVE web proxy differs from the last one this state DB saw, reset
// the whole retry queue — those failures queued while fetching direct
// deserve an immediate retry through the tunnel. Once per proxy epoch.
func (p *Pipeline) maybeRearmRetries() {
	defer func() { recover() }() // bookkeeping must never kill the batch
	epoch := webfetch.ProxyLabel(webfetch.ProxyFromConfig(p.config))
	if v, err := p.state.GetMeta(ProxyEpochKey); err == nil && v != nil && *v == epoch {
		return
	}
	count, err := p.state.RearmRetries()
	if err != nil {
		return
	}
	_ = p.state.SetMeta(ProxyEpochKey, epoch)
	if count > 0 {
		p.log(fmt.Sprintf("🔁 Web proxy active — re-armed %d queued retry(ies) for an immediate retry through the proxy", count), "info")
	}
}

// enforceBlockedDomains mirrors _enforce_blocked_domains (v0.20.0 +
// v0.21.0): purge the state DB of blocked/self-domain links and delete
// the placeholder FILES. Idempotent; dry-run aware; tolerated.
func (p *Pipeline) enforceBlockedDomains() {
	enforceDomains := func(domains []string, isSelf bool) {
		defer func() { recover() }()
		var report PurgeReport
		var err error
		if isSelf {
			report, err = p.state.PurgeBlockedDomains(func(u string) bool {
				return links.DomainIsSelf(u, domains)
			})
		} else {
			report, err = p.state.PurgeBlockedDomains(func(u string) bool {
				return links.DomainIsBlocked(u, domains)
			})
		}
		if err != nil {
			if isSelf {
				p.log(fmt.Sprintf("⚠️ Self-domain purge skipped: %v", err), "warning")
			} else {
				p.log(fmt.Sprintf("⚠️ Blocked-domain purge skipped: %v", err), "warning")
			}
			return
		}
		deletedFiles := 0
		for _, ph := range report.Placeholders {
			if _, statErr := os.Stat(ph.NotePath); statErr == nil {
				if err := dryrunRemove(ph.NotePath); err == nil {
					deletedFiles++
				}
			}
		}
		if report.Retries > 0 || len(report.Placeholders) > 0 {
			if isSelf {
				p.log(fmt.Sprintf("🔒 Self domains (%s — the app's own bot): purged %d queued retry(ies) + %d _review placeholder(s) (%d file(s) deleted, tokens gone with them) — never fetched, the _inbox row is the record",
					strings.Join(domains, ", "), report.Retries, len(report.Placeholders), deletedFiles), "info")
			} else {
				p.log(fmt.Sprintf("🚫 Blocked domains (%s): purged %d queued retry(ies) + %d _review placeholder(s) (%d file(s) deleted) — banned links are never collected (no note, no _review, no _inbox row)",
					strings.Join(domains, ", "), report.Retries, len(report.Placeholders), deletedFiles), "info")
			}
		}
	}
	enforceDomains(p.blockedDomains, false)
	if len(p.selfDomains) == 0 {
		return
	}
	enforceDomains(p.selfDomains, true)
}

// Taxonomy exposes the loaded taxonomy (the CLI engine's Website
// Directory rebuild reads it — pipeline.taxonomy). Additive seam.
func (p *Pipeline) Taxonomy() *taxonomy.Taxonomy { return p.tax }
