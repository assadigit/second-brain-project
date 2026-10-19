// Package batch, file linktracker.go — the port of gui/link_tracker.py
// (the 5-phase no-link-lost manifest). Task T-25 (Phase 2); spec:
// spec-gui-cache-db.md §link_tracker.py.
//
// The manifest JSON is written in Python's json.dump(indent=2) byte
// shape (insertion-ordered keys, ", ": " separators, ensure_ascii
// escaping) — pinned byte-for-byte by the golden's manifest_bytes.
package batch

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/store"
)

// ManifestLink is one manifest row. Field order is the JSON write order
// (Python's insertion-ordered dict).
type ManifestLink struct {
	URL         string  `json:"url"`
	Normalized  string  `json:"normalized"`
	Type        string  `json:"type"`
	Status      string  `json:"status"`
	NotePath    *string `json:"note_path"`
	Error       *string `json:"error"`
	ProcessedAt *string `json:"processed_at"`
}

// Manifest is the links_manifest.json document.
type Manifest struct {
	BatchID    string          `json:"batch_id"`
	CreatedAt  string          `json:"created_at"`
	Source     string          `json:"source"`
	TotalLinks int             `json:"total_links"`
	Links      []*ManifestLink `json:"links"`
}

// LinkTracker tracks every link through a 5-phase pipeline to ensure no
// link is ever lost. The manifest file is the source of truth.
type LinkTracker struct {
	vaultPath    string
	manifestPath string
	Manifest     *Manifest

	injectedIndex *VaultIndex
	gtIndex       *VaultIndex

	// ReconciledVaultHits — how many links the LAST
	// GetReconciliationURLs pass healed against the vault.
	ReconciledVaultHits int
	// SetAsideWebsites — how many WEBSITE rows the same pass set aside.
	SetAsideWebsites int
	// SetAsideSettled — how many settled-repo rows the same pass set
	// aside (v0.63.3).
	SetAsideSettled int
}

// NewLinkTracker mirrors LinkTracker.__init__.
func NewLinkTracker(vaultPath string) *LinkTracker {
	return &LinkTracker{
		vaultPath:    vaultPath,
		manifestPath: filepath.Join(vaultPath, "links_manifest.json"),
		Manifest: &Manifest{
			BatchID:    time.Now().Format("20060102_150405"),
			CreatedAt:  store.PyISOFormat(time.Now()),
			Source:     "unknown",
			TotalLinks: 0,
			Links:      []*ManifestLink{},
		},
	}
}

// SetVaultIndex mirrors set_vault_index: share an ALREADY-BUILT
// VaultIndex (the worker builds one at batch start and keeps it
// incrementally updated).
func (t *LinkTracker) SetVaultIndex(index *VaultIndex) {
	t.injectedIndex = index
}

// groundTruth mirrors _ground_truth: injected index first, then a
// lazily-built one cached for the instance's lifetime; nil when the
// vault cannot be scanned (a failed scan degrades to the pure-manifest
// behavior).
func (t *LinkTracker) groundTruth() *VaultIndex {
	if t.injectedIndex != nil {
		return t.injectedIndex
	}
	if t.gtIndex != nil {
		return t.gtIndex
	}
	if t.vaultPath == "" {
		return nil
	}
	if fi, err := os.Stat(t.vaultPath); err != nil || !fi.IsDir() {
		return nil
	}
	idx := NewVaultIndex(t.vaultPath, nil)
	idx.Rebuild(nil)
	t.gtIndex = idx
	return idx
}

// SetSource mirrors set_source (bot/saved/rss/import).
func (t *LinkTracker) SetSource(source string) {
	t.Manifest.Source = source
}

// Intake mirrors intake — Phase 1: record ALL links BEFORE any
// processing. Writes the manifest immediately so it survives crashes.
func (t *LinkTracker) Intake(githubURLs, nonGithubURLs []string) {
	for _, url := range githubURLs {
		t.Manifest.Links = append(t.Manifest.Links, &ManifestLink{
			URL: url, Normalized: links.NormalizeURL(url),
			Type: "github", Status: "pending",
		})
	}
	for _, url := range nonGithubURLs {
		t.Manifest.Links = append(t.Manifest.Links, &ManifestLink{
			URL: url, Normalized: links.NormalizeURL(url),
			Type: "non-github", Status: "pending",
		})
	}
	t.Manifest.TotalLinks = len(t.Manifest.Links)
	t.save()
}

func (t *LinkTracker) findLink(url string) *ManifestLink {
	norm := links.NormalizeURL(url)
	for _, link := range t.Manifest.Links {
		if link.Normalized == norm {
			return link
		}
	}
	return nil
}

func strPtr(s string) *string { return &s }

// MarkProcessing mirrors mark_processing — Phase 2: mark a link as
// being processed. The transient "processing" state is NO LONGER
// written to disk (the v0.06 perf law): a link left "pending" by a
// crash is treated exactly like one left "processing" (both are
// retried); the durable terminal states still save immediately.
func (t *LinkTracker) MarkProcessing(url string) {
	if link := t.findLink(url); link != nil {
		link.Status = "processing"
	}
}

// MarkProcessed mirrors mark_processed — Phase 2: mark a link as
// successfully processed.
func (t *LinkTracker) MarkProcessed(url, notePath string) {
	if link := t.findLink(url); link != nil {
		link.Status = "processed"
		link.NotePath = strPtr(notePath)
		now := store.PyISOFormat(time.Now())
		link.ProcessedAt = &now
		t.save()
	}
}

// MarkRecorded mirrors mark_recorded — Phase 2: mark a non-GitHub link
// as recorded in inbox.
func (t *LinkTracker) MarkRecorded(url string) {
	if link := t.findLink(url); link != nil {
		link.Status = "recorded"
		link.NotePath = strPtr("_inbox/non_github_links.md")
		now := store.PyISOFormat(time.Now())
		link.ProcessedAt = &now
		t.save()
	}
}

// MarkFailed mirrors mark_failed — Phase 2: mark a link as failed.
func (t *LinkTracker) MarkFailed(url, errMsg string) {
	if link := t.findLink(url); link != nil {
		link.Status = "failed"
		link.Error = strPtr(errMsg)
		t.save()
	}
}

// MarkSkipped mirrors mark_skipped (dedup, etc.).
func (t *LinkTracker) MarkSkipped(url, reason string) {
	if link := t.findLink(url); link != nil {
		link.Status = "skipped"
		link.Error = strPtr(reason)
		t.save()
	}
}

// MarkBlocked mirrors mark_blocked (v0.21.0 — blocked/self-domain,
// never fetched; the manifest entry ITSELF is the record).
func (t *LinkTracker) MarkBlocked(url, reason string) {
	if link := t.findLink(url); link != nil {
		link.Status = "blocked"
		link.Error = strPtr(reason)
		t.save()
	}
}

// VerifyReport mirrors verify()'s report dict.
type VerifyReport struct {
	Total              int
	GithubProcessed    int
	GithubFailed       int
	GithubSkipped      int
	GithubPending      int
	NonGithubRecorded  int
	NonGithubFailed    int
	WebsitesProcessed  int
	WebsitesReview     int
	WebsitesSkipped    int
	BlockedRecorded    int
	NonGithubPending   int
	VerificationPassed bool
	FailedLinks        []*ManifestLink
	Accounted          int
	Unaccounted        int
	AccountingOK       bool
}

// Verify mirrors verify (link_tracker.py:206-449) — Phase 3: verify
// all links have their expected output, with FULL ACCOUNTING (every
// bucket sums to the total or something escaped every path — a loud
// error, never silent) and the v0.38.0 healing (a stale note_path is
// re-pointed via the vault ground truth — a MOVE is not a loss;
// pending rows whose note is in the vault heal to skipped).
func (t *LinkTracker) Verify(logSignal func(msg, level string),
	extraInboxDirs []string) *VerifyReport {
	rep := &VerifyReport{
		Total:              len(t.Manifest.Links),
		VerificationPassed: true,
	}
	inboxDirs := []string{filepath.Join(t.vaultPath, "_inbox")}
	for _, d := range extraInboxDirs {
		if d != "" {
			dup := false
			for _, x := range inboxDirs {
				if x == d {
					dup = true
					break
				}
			}
			if !dup {
				inboxDirs = append(inboxDirs, d)
			}
		}
	}

	// urlInInbox: true/false when known; ok=false when the tables
	// could not be read (tolerant — never false-fail on I/O).
	urlInInbox := func(url string) (found, readAny bool) {
		for _, inboxDir := range inboxDirs {
			if fi, err := os.Stat(inboxDir); err != nil || !fi.IsDir() {
				continue
			}
			fnames, err := os.ReadDir(inboxDir)
			if err != nil {
				continue
			}
			for _, e := range fnames {
				if !strings.HasSuffix(e.Name(), ".md") {
					continue
				}
				fpath := filepath.Join(inboxDir, e.Name())
				data, err := os.ReadFile(fpath)
				if err != nil {
					continue
				}
				if !utf8.Valid(data) {
					continue // the strict read would raise
				}
				readAny = true
				content := string(data)
				if strings.Contains(content, url) ||
					strings.Contains(content, links.NormalizeURL(url)) {
					found = true
					return found, readAny
				}
			}
		}
		return found, readAny
	}

	for _, link := range t.Manifest.Links {
		if link.Type == "github" {
			switch link.Status {
			case "processed":
				// Verify note exists on disk and is non-empty.
				// (Python: isfile(note_path) gates the size probe — a
				// missing file OR a directory falls to the ground-truth
				// branch; the getsize race's "cannot read note file"
				// shape is subsumed by Stat's error.)
				notePath := ""
				if link.NotePath != nil {
					notePath = *link.NotePath
				}
				isFile := false
				var size int64
				if notePath != "" {
					if fi, err := os.Stat(notePath); err == nil &&
						!fi.IsDir() {
						isFile = true
						size = fi.Size()
					}
				}
				if isFile {
					if size > 100 {
						rep.GithubProcessed++
					} else {
						link.Status = "failed"
						link.Error = strPtr("note file is empty")
						rep.GithubFailed++
						rep.FailedLinks = append(rep.FailedLinks, link)
						rep.VerificationPassed = false
					}
					continue
				}
				// A stale note_path is not automatically a failure: the
				// note may have MOVED inside the vault. The ground truth
				// is keyed by source URL — re-point and count processed.
				if newPath := t.vaultPathFor(link.URL); newPath != "" {
					link.NotePath = strPtr(newPath)
					rep.GithubProcessed++
				} else {
					link.Status = "failed"
					link.Error = strPtr("note file missing")
					rep.GithubFailed++
					rep.FailedLinks = append(rep.FailedLinks, link)
					rep.VerificationPassed = false
				}
			case "skipped":
				rep.GithubSkipped++
			case "failed":
				rep.GithubFailed++
				rep.FailedLinks = append(rep.FailedLinks, link)
				rep.VerificationPassed = false
			case "pending", "processing":
				// REAL unfinished work — unless the note is already IN
				// THE VAULT (heal the row to skipped and account it).
				if t.vaultHas(link.URL) {
					link.Status = "skipped"
					link.Error = strPtr("already in vault (reconciled)")
					if p := t.vaultPathFor(link.URL); p != "" {
						link.NotePath = strPtr(p)
					}
					now := store.PyISOFormat(time.Now())
					link.ProcessedAt = &now
					rep.GithubSkipped++
				} else {
					rep.GithubPending++
					rep.VerificationPassed = false
				}
			}
		} else if link.Type == "non-github" {
			switch link.Status {
			case "recorded":
				// Scan ALL .md files in _inbox/ for the URL (per-platform
				// files + the legacy non_github_links.md).
				found, readAny := urlInInbox(link.URL)
				if !found && readAny {
					link.Status = "failed"
					link.Error = strPtr(
						"URL not found in any _inbox/*.md file")
					rep.NonGithubFailed++
					rep.FailedLinks = append(rep.FailedLinks, link)
					rep.VerificationPassed = false
				} else {
					// found, or the tables were unreadable — never
					// false-fail on I/O.
					rep.NonGithubRecorded++
				}
			case "blocked":
				// v0.35.0 — banned links are NEVER collected; this
				// manifest entry IS the record.
				rep.BlockedRecorded++
			case "processed":
				// The Websites pipeline wrote a note (a _review
				// placeholder counts: it IS a note, with a retry
				// scheduled). Verify the file is really on disk.
				notePath := ""
				if link.NotePath != nil {
					notePath = *link.NotePath
				}
				okNote := false
				if notePath != "" {
					if fi, err := os.Stat(notePath); err == nil &&
						!fi.IsDir() && fi.Size() > 50 {
						okNote = true
					}
				}
				if okNote {
					if strings.Contains(
						strings.ReplaceAll(notePath, "\\", "/"),
						"_review") {
						rep.WebsitesReview++
					} else {
						rep.WebsitesProcessed++
					}
				} else {
					link.Status = "failed"
					link.Error = strPtr("websites note file missing")
					rep.NonGithubFailed++
					rep.FailedLinks = append(rep.FailedLinks, link)
					rep.VerificationPassed = false
				}
			case "skipped":
				rep.WebsitesSkipped++
			case "failed":
				rep.NonGithubFailed++
				rep.FailedLinks = append(rep.FailedLinks, link)
				rep.VerificationPassed = false
			case "pending", "processing":
				// Websites pipeline off / no vault — its own bucket, NOT
				// a failure (same semantics as GetAllClear).
				rep.NonGithubPending++
			}
		}
	}

	// Reconciliation: every bucket sums to the total, or something
	// escaped every path (a bug — loud, never silent).
	rep.Accounted = rep.GithubProcessed + rep.GithubFailed +
		rep.GithubSkipped + rep.GithubPending + rep.NonGithubRecorded +
		rep.NonGithubFailed + rep.WebsitesProcessed + rep.WebsitesReview +
		rep.WebsitesSkipped + rep.BlockedRecorded + rep.NonGithubPending
	rep.Unaccounted = rep.Total - rep.Accounted
	rep.AccountingOK = rep.Unaccounted == 0

	t.save()

	if logSignal != nil {
		logSignal(strings.Repeat("=", 50), "info")
		logSignal("🔍 VERIFICATION REPORT", "info")
		logSignal(strings.Repeat("=", 50), "info")
		logSignal(fmt.Sprintf("   Total links: %d", rep.Total), "info")
		processedLevel := "info" // the reference: success when any link processed
		if rep.GithubProcessed > 0 {
			processedLevel = "success"
		}
		logSignal(fmt.Sprintf("   GitHub processed: %d",
			rep.GithubProcessed), processedLevel)
		logSignal(fmt.Sprintf("   GitHub skipped (dedup): %d",
			rep.GithubSkipped), "info")
		if rep.GithubPending > 0 {
			logSignal(fmt.Sprintf("   GitHub pending (unfinished!): %d",
				rep.GithubPending), "warning")
		}
		logSignal(fmt.Sprintf("   GitHub failed: %d",
			rep.GithubFailed), "info")
		if rep.NonGithubRecorded > 0 {
			logSignal(fmt.Sprintf("   Non-GitHub recorded (inbox): %d",
				rep.NonGithubRecorded), "info")
		}
		if rep.WebsitesProcessed > 0 || rep.WebsitesReview > 0 ||
			rep.WebsitesSkipped > 0 {
			logSignal(fmt.Sprintf("   Websites notes: %d",
				rep.WebsitesProcessed), "info")
			logSignal(fmt.Sprintf(
				"   Websites in _review (retry scheduled): %d",
				rep.WebsitesReview), "info")
			logSignal(fmt.Sprintf("   Websites skipped (dedup): %d",
				rep.WebsitesSkipped), "info")
		}
		if rep.BlockedRecorded > 0 {
			logSignal(fmt.Sprintf(
				"   Blocked/self domains (omitted by design — never "+
					"collected): %d", rep.BlockedRecorded), "info")
		}
		if rep.NonGithubPending > 0 {
			logSignal(fmt.Sprintf(
				"   Non-GitHub pending (websites pipeline off / no "+
					"vault): %d", rep.NonGithubPending), "warning")
		}
		if rep.NonGithubFailed > 0 {
			logSignal(fmt.Sprintf("   Non-GitHub failed: %d",
				rep.NonGithubFailed), "error")
		}
		if rep.AccountingOK {
			logSignal(fmt.Sprintf(
				"   🧮 Accounting: %d/%d links accounted for ✓",
				rep.Accounted, rep.Total), "info")
		} else {
			logSignal(fmt.Sprintf(
				"   🧮 Accounting: only %d/%d accounted for — %d "+
					"escaped every bucket!",
				rep.Accounted, rep.Total, rep.Unaccounted), "error")
		}
		if rep.VerificationPassed && rep.AccountingOK {
			logSignal("✅ ALL LINKS VERIFIED — no data loss!", "success")
		} else {
			logSignal(fmt.Sprintf("❌ %d links need retry!",
				len(rep.FailedLinks)), "error")
			for _, fl := range rep.FailedLinks {
				errMsg := "unknown"
				if fl.Error != nil {
					errMsg = *fl.Error
				}
				logSignal(fmt.Sprintf("   • %s — %s", fl.URL, errMsg),
					"error")
			}
		}
		logSignal(strings.Repeat("=", 50), "info")
	}

	return rep
}

// GetFailedLinks mirrors get_failed_links — Phase 4: all links that
// need retry.
func (t *LinkTracker) GetFailedLinks() []*ManifestLink {
	var out []*ManifestLink
	for _, link := range t.Manifest.Links {
		if link.Status == "failed" || link.Status == "processing" {
			out = append(out, link)
		}
	}
	return out
}

func (t *LinkTracker) vaultHas(url string) bool {
	idx := t.groundTruth()
	if idx == nil {
		return false
	}
	return idx.HasURL(url)
}

// VaultHas mirrors vault_has — the public face (the finished-batch
// logger lists only GENUINELY unfinished links).
func (t *LinkTracker) VaultHas(url string) bool {
	return t.vaultHas(url)
}

func (t *LinkTracker) vaultPathFor(url string) string {
	idx := t.groundTruth()
	if idx == nil {
		return ""
	}
	return idx.GetPath(url)
}

// GetAllClear mirrors get_all_clear — Phase 5: are ALL links verified?
// Only GITHUB rows gate (v0.63.1 — the Websites pipeline's own doors
// own non-github rows); a failed/processing/pending GitHub row whose
// note IS in the vault does not block (v0.38.0 — the vault is the
// ground truth).
func (t *LinkTracker) GetAllClear() bool {
	for _, link := range t.Manifest.Links {
		if link.Type != "github" {
			continue
		}
		switch link.Status {
		case "failed", "processing", "pending":
			if !t.vaultHas(link.URL) {
				return false
			}
		}
	}
	return true
}

func (t *LinkTracker) save() {
	t.saveManifestDict(t.Manifest)
}

// saveManifestDict mirrors _save_manifest_dict (v0.38.0 — the atomic
// writer factored out so reconciliation can persist a HEALED previous
// manifest without touching the live one). Dry-run records, never
// writes. Atomic: temp file in the vault dir, then rename.
func (t *LinkTracker) saveManifestDict(m *Manifest) {
	if dryrun.IsEnabled() {
		dryrun.Record("write", t.manifestPath, nil,
			"links_manifest.json (link tracker)")
		return
	}
	tmp, err := os.CreateTemp(t.vaultPath, "*.tmp")
	if err != nil {
		return // best-effort — never crash the batch
	}
	tmpPath := tmp.Name()
	if _, err := tmp.WriteString(pyDumpManifest(m)); err != nil {
		tmp.Close()
		_ = os.Remove(tmpPath)
		return
	}
	if err := tmp.Close(); err != nil {
		_ = os.Remove(tmpPath)
		return
	}
	if err := os.Rename(tmpPath, t.manifestPath); err != nil {
		_ = os.Remove(tmpPath)
	}
}

// LoadPreviousManifest mirrors load_previous_manifest: the previous
// manifest for reconciliation, or nil.
func (t *LinkTracker) LoadPreviousManifest() *Manifest {
	data, err := os.ReadFile(t.manifestPath)
	if err != nil {
		return nil
	}
	var m Manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil
	}
	for _, l := range m.Links {
		if l == nil {
			return nil
		}
	}
	return &m
}

// GetReconciliationURLs mirrors get_reconciliation_urls
// (link_tracker.py:566-692) — Phase 4: URLs from the previous manifest
// that need retry, healed against the vault ground truth:
//
//   - pending/processing/failed rows whose note IS in the vault →
//     healed to skipped, never retried, and the healed manifest is
//     written back;
//   - processed rows with a stale note_path → re-pointed;
//   - non-github rows are NEVER this queue's retry (set aside with an
//     honest tally — the Websites pipeline's own doors own them);
//   - settled github rows (the v0.63.3 ledger, probed through the
//     default-anchored CacheDB) are never a retry either;
//   - only links absent from BOTH the manifest's good graces and the
//     vault are returned.
func (t *LinkTracker) GetReconciliationURLs() []string {
	t.ReconciledVaultHits = 0
	t.SetAsideWebsites = 0
	t.SetAsideSettled = 0
	prev := t.LoadPreviousManifest()
	if prev == nil {
		return nil
	}
	// The settled probe: one read, before the loop. Guarded — a broken
	// cache.db never resurrects a settled repo's cry.
	settledRepos := map[string]bool{}
	if cache, err := store.NewCacheDB("cache.db"); err == nil {
		settledRepos = cache.GetSettledRepoSet()
		cache.Close()
	}
	var failed []string
	healed := false
	for _, link := range prev.Links {
		if link.Type != "github" {
			// A WEBSITE row: its retries belong to the Websites
			// pipeline's own doors. Count the still-waiting ones; never
			// cry them here.
			switch link.Status {
			case "failed", "processing", "pending":
				t.SetAsideWebsites++
			case "processed":
				np := ""
				if link.NotePath != nil {
					np = *link.NotePath
				}
				if np == "" {
					t.SetAsideWebsites++
				} else if fi, err := os.Stat(np); err != nil || fi.IsDir() {
					t.SetAsideWebsites++
				}
			}
			continue
		}
		switch link.Status {
		case "failed", "processing", "pending":
			if t.vaultHas(link.URL) {
				// Stored — the ledger just never got its terminal
				// update. Heal the row; it is nobody's retry.
				link.Status = "skipped"
				link.Error = strPtr("already in vault (reconciled)")
				if p := t.vaultPathFor(link.URL); p != "" {
					link.NotePath = strPtr(p)
				}
				now := store.PyISOFormat(time.Now())
				link.ProcessedAt = &now
				healed = true
				t.ReconciledVaultHits++
			} else if settledRepos[link.Normalized] ||
				settledRepos[link.URL] {
				// THE SETTLED REPOS LEDGER: this row predates the
				// settlement, and the repo it points at is addressed
				// now — the owner's law. Never a retry, never a cry.
				t.SetAsideSettled++
			} else {
				failed = append(failed, link.URL)
			}
		case "processed":
			// Check if note still exists — at the recorded path, or
			// anywhere in the vault (a move is not a loss).
			notePath := ""
			if link.NotePath != nil {
				notePath = *link.NotePath
			}
			if notePath == "" {
				if newPath := t.vaultPathFor(link.URL); newPath != "" {
					link.NotePath = strPtr(newPath)
					healed = true
				} else {
					failed = append(failed, link.URL)
				}
				continue
			}
			if fi, err := os.Stat(notePath); err != nil || fi.IsDir() {
				if newPath := t.vaultPathFor(link.URL); newPath != "" {
					link.NotePath = strPtr(newPath)
					healed = true
				} else {
					failed = append(failed, link.URL)
				}
			}
		}
	}
	if healed {
		// Persist the HEALED previous manifest. The live manifest stays
		// untouched — it belongs to the next batch's intake.
		t.saveManifestDict(prev)
	}
	return failed
}

// ---------------------------------------------------------------------------
// The Python json.dump(indent=2) byte-shape writer.
// ---------------------------------------------------------------------------

// pyDumpManifest serializes a Manifest exactly like Python's
// json.dump(m, f, indent=2, default=str): insertion-ordered keys, the
// ",\n<indent>" item separator, the ": " key separator, ensure_ascii
// escaping (non-ASCII as \uXXXX, lowercase hex), None as null.
func pyDumpManifest(m *Manifest) string {
	var b strings.Builder
	b.WriteString("{\n")
	b.WriteString(`  "batch_id": ` + pyJSONString(m.BatchID) + ",\n")
	b.WriteString(`  "created_at": ` + pyJSONString(m.CreatedAt) + ",\n")
	b.WriteString(`  "source": ` + pyJSONString(m.Source) + ",\n")
	b.WriteString(`  "total_links": ` + fmt.Sprintf("%d", m.TotalLinks) +
		",\n")
	b.WriteString(`  "links": [`)
	if len(m.Links) == 0 {
		b.WriteString("]\n}")
		return b.String()
	}
	b.WriteString("\n")
	for i, l := range m.Links {
		b.WriteString("    {\n")
		b.WriteString(`      "url": ` + pyJSONString(l.URL) + ",\n")
		b.WriteString(`      "normalized": ` + pyJSONString(l.Normalized) +
			",\n")
		b.WriteString(`      "type": ` + pyJSONString(l.Type) + ",\n")
		b.WriteString(`      "status": ` + pyJSONString(l.Status) + ",\n")
		b.WriteString(`      "note_path": ` + pyJSONNullString(l.NotePath) +
			",\n")
		b.WriteString(`      "error": ` + pyJSONNullString(l.Error) + ",\n")
		b.WriteString(`      "processed_at": ` +
			pyJSONNullString(l.ProcessedAt) + "\n")
		b.WriteString("    }")
		if i != len(m.Links)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString("  ]\n}")
	return b.String()
}

// pyJSONNullString renders a *string as null or a JSON string.
func pyJSONNullString(s *string) string {
	if s == nil {
		return "null"
	}
	return pyJSONString(*s)
}

// pyJSONString renders a Go string as Python's json.dumps scalar
// (ensure_ascii=True): every non-ASCII rune escapes as \uXXXX (surrogates
// for astral runes), the standard JSON short escapes apply.
func pyJSONString(s string) string {
	var b strings.Builder
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		default:
			if r < 0x20 {
				b.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else if r < 0x7f {
				b.WriteRune(r)
			} else if r <= 0xffff {
				b.WriteString(fmt.Sprintf(`\u%04x`, r))
			} else {
				// astral plane: the surrogate pair Python emits
				r -= 0x10000
				hi := 0xd800 + (r >> 10)
				lo := 0xdc00 + (r & 0x3ff)
				b.WriteString(fmt.Sprintf(`\u%04x\u%04x`, hi, lo))
			}
		}
	}
	b.WriteByte('"')
	return b.String()
}
