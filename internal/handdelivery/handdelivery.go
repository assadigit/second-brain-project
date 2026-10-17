// Package handdelivery ports core/hand_delivery.py — THE FOURTH DOOR:
// the owner's own Chrome (v0.48.0).
//
// T-40's SURFACE (SESSION-015): websitepipeline consumes this
// package's predicates, queue, and consume/stamp functions directly
// (the reference imports hand_delivery at module level). What ships
// here is the minimal surface that consumption needs:
//
//	StatusIsHand, IsWalledReason        the pure gestures
//	HandDeliveryDir/QueuePath/ReadmePath, ReadQueue,
//	SuggestedFilename, ReadmeText,
//	EnqueueHandDelivery                 the queue write side
//	CollectDelivered, HandFetchResult,
//	TakeHandDelivered                   the consume side
//	StampHandDeliveredRows              the harvest stamp
//
// What stays for T-43 (the module's own task): the Chrome launch
// family (find_chrome/open_in_chrome — T-44's browser leg), the
// batch-level consume_delivered + walled_retry_rows + pending_hand_
// links + stamp_hand_rows pickers, and the module's own test battery
// (test_handdelivery/test_handharvest). The v0.53 auto-page
// re-verification (chrome_tabs.page_is_real) degrades exactly the way
// the reference's own except-path defines ("a verdict that cannot be
// asked is not a wall") until T-44 lands — documented at
// autoPageIsReal.
//
// LAYERING: the reference's stamp functions import website_pipeline's
// table helpers via Python's function-level imports; Go has no
// deferred imports, so the shared table grammar lives in
// internal/websitepipeline/decomtable (below both packages — see that
// package's doc for the layering record).
//
// WIRE SHAPE: queue.json is Python's json.dumps(queue, indent=2,
// ensure_ascii=False) — insertion-ordered keys, raw UTF-8. The queue's
// link metas are *storage.OrderedMap values (never structs) so the
// house PyDumps encoder writes them byte-identically; the QueueLink
// helpers below rehydrate typed views without changing the stored
// shape.
package handdelivery

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"unicode"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/webfetch"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// LogFn mirrors the reference's optional log callback: log(msg, level).
type LogFn = func(msg, level string)

func noopLog(msg, level string) {}

// HandFolderName mirrors HAND_FOLDER_NAME: <vault>/_review/hand-delivered/.
const HandFolderName = "hand-delivered"

// QueueFilename mirrors QUEUE_FILENAME: the queue file inside the folder.
const QueueFilename = "queue.json"

// ReadmeFilename mirrors README_FILENAME: the instructions file.
const ReadmeFilename = "README.txt"

// HandMarkers mirrors HAND_MARKERS (v0.48.0 + v0.52.0): Status-cell
// markers meaning "queue this link for the fourth door" (substring +
// case-insensitive; deliberately AFTER the retirement markers in the
// check order — a cell that says both "dead" and "hand" is DEAD).
var HandMarkers = []string{"🖐", "✋", "🫱", "hand-deliver", "hand deliver", "hand", "chrome"}

// WallMarkers mirrors WALL_MARKERS: the reason classes that earn the
// fourth door (refusal family, bot-defense, challenge, TLS-handshake,
// the third door's report line). A 404/paywall/DNS truth never asks
// the owner's hand — it has its own door.
var WallMarkers = []string{
	"http 403", "http 405", "http 429", "http 451",
	"bot defense", "challenge", "blocked_bot", "refused",
	"ssl", "handshake", "chrome-impersonated", "curl_cffi",
	"fingerprint",
}

// slugStrip mirrors _SLUG_STRIP (suggested_filename's strip set).
const slugStrip = " /\\?%*:|\"<>.~#&=+,@"

// HandDeliveryDir mirrors hand_delivery_dir: the fourth door's folder.
func HandDeliveryDir(vaultPath string) string {
	if vaultPath == "" {
		return filepath.Join(decomtable.ReviewFolder, HandFolderName)
	}
	return filepath.Join(vaultPath, decomtable.ReviewFolder, HandFolderName)
}

// QueuePath mirrors queue_path: the queue.json path for one vault.
func QueuePath(vaultPath string) string {
	return filepath.Join(HandDeliveryDir(vaultPath), QueueFilename)
}

// ReadmePath mirrors readme_path.
func ReadmePath(vaultPath string) string {
	return filepath.Join(HandDeliveryDir(vaultPath), ReadmeFilename)
}

// IsWalledReason mirrors is_walled_reason: True when a failure REASON
// names a wall the fourth door answers. Case-insensitive substring
// markers — the honest, over-inclusive side of the predicate.
func IsWalledReason(reason string) bool {
	s := strings.ToLower(strings.TrimSpace(reason))
	if s == "" {
		return false
	}
	for _, m := range WallMarkers {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}

// SuggestedFilename mirrors suggested_filename:
// hand-<domain-slug>-<hash8>.html — deterministic for one URL
// (v0.64.0: the hash rides the CANONICAL form, so https://site/ and
// https://site suggest the ONE name).
func SuggestedFilename(rawURL string) string {
	u := rawURL
	if c := links.NormalizeWebsiteURL(u); c != "" {
		u = c
	}
	netloc := "link"
	if parsed, err := url.Parse(u); err == nil && parsed.Host != "" {
		netloc = strings.ToLower(parsed.Host)
	}
	var b strings.Builder
	for _, c := range netloc {
		if strings.ContainsRune(slugStrip, c) || !unicode.IsLetter(c) && !unicode.IsNumber(c) {
			b.WriteByte('-')
		} else {
			b.WriteRune(c)
		}
	}
	slug := strings.Trim(b.String(), "-")
	if len(slug) > 40 {
		slug = slug[:40]
	}
	if slug == "" {
		slug = "link"
	}
	sum := md5.Sum([]byte(u))
	digest := hex.EncodeToString(sum[:])[:8]
	return fmt.Sprintf("hand-%s-%s.html", slug, digest)
}

// Queue mirrors the queue's shape: {"links": {url: meta}} — a
// URL-keyed object keeps re-enqueueing a no-op by construction.
// Corrupt/missing reads as the empty queue (the door never breaks on a
// bad file). Each meta is an *storage.OrderedMap (the on-disk JSON
// shape, insertion-ordered: wall, queued, suggested, door?, consumed?,
// requeued?).
type Queue struct {
	Links *storage.OrderedMap
}

// NewQueue returns the empty queue.
func NewQueue() *Queue { return &Queue{Links: storage.NewOrderedMap()} }

// ReadQueue mirrors _read_queue.
func ReadQueue(vaultPath string) *Queue {
	p := QueuePath(vaultPath)
	data, err := os.ReadFile(p)
	if err != nil {
		return NewQueue()
	}
	om, err := storage.DecodeJSON(data)
	if err != nil {
		return NewQueue()
	}
	raw, ok := om.Get("links")
	if !ok {
		return NewQueue()
	}
	linksOM, ok := raw.(*storage.OrderedMap)
	if !ok {
		return NewQueue()
	}
	return &Queue{Links: linksOM}
}

// Meta returns the ordered meta map for one URL (nil when absent).
func (q *Queue) Meta(u string) *storage.OrderedMap {
	if q == nil || q.Links == nil {
		return nil
	}
	raw, ok := q.Links.Get(u)
	if !ok {
		return nil
	}
	if om, ok := raw.(*storage.OrderedMap); ok {
		return om
	}
	return nil
}

// metaStr reads one string field off a meta map.
func metaStr(om *storage.OrderedMap, key string) string {
	if om == nil {
		return ""
	}
	v, ok := om.Get(key)
	if !ok || v == nil {
		return ""
	}
	if s, ok := v.(string); ok {
		return s
	}
	return ""
}

// metaTruthy reads one field with Python truthiness (a False/""/"0"
// consumed stamp reads NOT consumed).
func metaTruthy(om *storage.OrderedMap, key string) bool {
	if om == nil {
		return false
	}
	v, ok := om.Get(key)
	if !ok || v == nil {
		return false
	}
	switch t := v.(type) {
	case bool:
		return t
	case string:
		return t != ""
	case float64:
		return t != 0
	case int64:
		return t != 0
	default:
		return true
	}
}

// Dumps renders the queue as Python's json.dumps(queue, indent=2,
// ensure_ascii=False) — the wire shape the README contract and the
// owner's own hand-edits both ride.
func (q *Queue) Dumps() string {
	root := storage.NewOrderedMap()
	root.Set("links", q.Links)
	return storage.PyDumps(root)
}

// WriteAtomic persists the queue through the house atomic writer.
func (q *Queue) WriteAtomic(vaultPath string) error {
	return storage.AtomicWriteText(QueuePath(vaultPath), q.Dumps())
}

// Clock is the package's clock seam: the golden runner freezes it
// (the reference generator stubs hd.time for the same runs);
// production reads the wall clock.
var Clock = time.Now

// stampNow mirrors time.strftime('%Y-%m-%d %H:%M') (local time).
func stampNow() string { return Clock().Format("2006-01-02 15:04") }

// dateNow mirrors time.strftime('%Y-%m-%d') (local time).
func dateNow() string { return Clock().Format("2006-01-02") }

// ReadmeText mirrors _readme_text: the instructions file regenerated
// on every enqueue — the file IS the dialog.
func ReadmeText(q *Queue) string {
	lines := []string{
		"GitCurator - the fourth door: hand-delivered pages",
		strings.Repeat("=", 52),
		"",
		"These links were walled for every machine door (403 / bot",
		"defense / TLS fingerprint). Your real Chrome can open them.",
		"",
		"For each link below:",
		"  1. open it in Chrome (More > 'Hand-deliver walled links'",
		"     does it for you),",
		"  2. save the page here: Ctrl+S, format",
		"     'Webpage, HTML Only', filename = the suggested name,",
		"  3. run the next batch (SYNC) - the page is consumed as a",
		"     real fetch, the note is written, the retry clears.",
		"",
		"Queued links:",
	}
	type entry struct{ url, sug, wall string }
	var entries []entry
	for _, u := range q.Links.Keys() {
		m := q.Meta(u)
		sug, wall := "", ""
		if m != nil {
			sug = metaStr(m, "suggested")
			wall = metaStr(m, "wall")
		}
		if sug == "" {
			sug = SuggestedFilename(u)
		}
		if len(wall) > 120 {
			wall = wall[:120]
		}
		entries = append(entries, entry{u, sug, wall})
	}
	sort.SliceStable(entries, func(i, j int) bool { return entries[i].url < entries[j].url })
	for _, e := range entries {
		lines = append(lines, "  - "+e.url)
		lines = append(lines, "    save as: "+e.sug)
		if e.wall != "" {
			lines = append(lines, "    the wall: "+e.wall)
		}
	}
	lines = append(lines,
		"",
		"Saved pages already consumed are marked in queue.json",
		"(consumed: true) - do not delete them; they are the record.",
		"Delete this README freely - it regenerates on the next",
		"hand-delivery enqueue.",
		"",
	)
	return strings.Join(lines, "\n")
}

// EnqueueReport mirrors enqueue_hand_delivery's return shape.
type EnqueueReport struct {
	Added       int
	QueuedTotal int
	Opened      int
	Folder      string
}

// EnqueueHandDelivery mirrors enqueue_hand_delivery (the write side
// of the door): merges into queue.json (an already-queued URL is
// refreshed, never duplicated); rewrites README.txt with the full
// current queue. Dry-run aware. The open_chrome leg is T-43/T-44's
// (the pipeline never asks for it); door stamps which door queued the
// link. Never raises.
func EnqueueHandDelivery(vaultPath string, urls []string,
	walls map[string]string, log LogFn, door string) EnqueueReport {
	if log == nil {
		log = noopLog
	}
	report := EnqueueReport{Folder: HandDeliveryDir(vaultPath)}
	if vaultPath == "" || len(urls) == 0 {
		return report
	}
	folder := HandDeliveryDir(vaultPath)
	if err := dryrun.MkdirAll(folder, true); err != nil {
		log(fmt.Sprintf("⚠️ Hand-delivery folder could not be created (%s): %v", folder, err), "warning")
		return report
	}
	if walls == nil {
		walls = map[string]string{}
	}
	queue := ReadQueue(vaultPath)
	now := stampNow()
	var added []string
	for _, rawURL := range urls {
		u := strings.TrimSpace(rawURL)
		lu := strings.ToLower(u)
		if u == "" || !(strings.HasPrefix(lu, "http://") || strings.HasPrefix(lu, "https://")) {
			continue
		}
		// v0.64.0 — THE ONE SPELLING at the queue's own door: the
		// queue keys by the CANONICAL form; a legacy row still resting
		// under its raw spelling is re-keyed on sight (the heal).
		canon := links.NormalizeWebsiteURL(u)
		if canon == "" {
			canon = u
		}
		if canon != u {
			_, legacyExists := queue.Links.Get(u)
			_, canonExists := queue.Links.Get(canon)
			if legacyExists && !canonExists {
				if raw, ok := queue.Links.Get(u); ok {
					queue.Links.Delete(u)
					queue.Links.Set(canon, raw) // the heal: one key per link
				}
				u = canon
			} else if canonExists {
				queue.Links.Delete(u) // both spellings: one row wins
				u = canon
			} else {
				u = canon
			}
		}
		if m := queue.Meta(u); m != nil && metaTruthy(m, "consumed") {
			// a re-enqueue of a consumed link resets it (the owner may
			// hand-deliver again — e.g. the page changed)
			m.Set("consumed", false)
			m.Set("requeued", now)
			if door != "" {
				m.Set("door", door)
			}
			added = append(added, u)
			continue
		}
		if queue.Meta(u) != nil {
			continue // already queued — a no-op, never a duplicate
		}
		wall := walls[rawURL]
		if wall == "" {
			wall = walls[u]
		}
		meta := storage.NewOrderedMap()
		meta.Set("wall", wall)
		meta.Set("queued", now)
		meta.Set("suggested", SuggestedFilename(u))
		if door != "" {
			meta.Set("door", door)
		}
		queue.Links.Set(u, meta)
		added = append(added, u)
	}
	if len(added) > 0 {
		if err := dryrun.WriteText(QueuePath(vaultPath), queue.Dumps()); err == nil {
			_ = dryrun.WriteText(ReadmePath(vaultPath), ReadmeText(queue))
		} else {
			log(fmt.Sprintf("⚠️ Hand-delivery queue could not be written: %v", err), "warning")
			return report
		}
		report.Added = len(added)
		report.QueuedTotal = queue.Links.Len()
		log(fmt.Sprintf("🖐 Hand-delivery: %d link(s) queued — save each page (Ctrl+S, 'Webpage, HTML Only') as its suggested name into %s; the next batch consumes them", len(added), folder), "info")
	} else {
		report.QueuedTotal = queue.Links.Len()
	}
	return report
}

// DeliveredPage mirrors one collect_delivered item:
// {'url', 'wall', 'suggested', 'file', 'door'}.
type DeliveredPage struct {
	URL       string
	Wall      string
	Suggested string
	File      string
	Door      string
}

// CollectDelivered mirrors collect_delivered: the queued links whose
// suggested page file HAS been delivered (the owner saved it into the
// folder, or the fifth door wrote it). Orphan files (an .html in the
// folder matching no queue row) are IGNORED — the README is the
// contract, a stray file is the owner's business.
func CollectDelivered(vaultPath string) []DeliveredPage {
	queue := ReadQueue(vaultPath)
	folder := HandDeliveryDir(vaultPath)
	var out []DeliveredPage
	for _, u := range queue.Links.Keys() {
		m := queue.Meta(u)
		if m == nil || metaTruthy(m, "consumed") {
			continue
		}
		suggested := metaStr(m, "suggested")
		if suggested == "" {
			suggested = SuggestedFilename(u)
		}
		path := filepath.Join(folder, suggested)
		if st, err := os.Stat(path); err == nil && !st.IsDir() && st.Size() > 0 {
			out = append(out, DeliveredPage{
				URL: u, Wall: metaStr(m, "wall"), Suggested: suggested,
				File: path, Door: metaStr(m, "door"),
			})
		}
	}
	return out
}

// autoPageIsReal mirrors _auto_page_is_real's honest degradation: the
// reference asks chrome_tabs.page_is_real and answers (True, ”) when
// the verdict cannot be asked. Until T-44 lands the port has no
// chrome_tabs to ask — the same law, the same answer.
func autoPageIsReal(u string, body []byte) (bool, string) {
	_ = u
	_ = body
	return true, ""
}

// HandFetchResult mirrors hand_fetch_result: build the FetchResult a
// delivered page becomes — a REAL full fetch whose reason tells the
// story (the honest line the note, the master table, and the run
// report all carry). v0.50.0: door == "auto" words the fifth door's
// story; the default wording stays the fourth door's.
func HandFetchResult(u string, body []byte, wall string, door string) *webfetch.FetchResult {
	text := string(body)
	wall = strings.TrimSpace(wall)
	if wall == "" {
		wall = "the machine doors were walled"
	}
	var reason string
	if door == "auto" {
		reason = fmt.Sprintf("auto-delivered via the owner's own Chrome (the fifth door's tab retry — one tab per failed link, the live DOM taken after the page loaded) — the machine doors failed: %s", wall)
	} else {
		reason = fmt.Sprintf("hand-delivered via the owner's real Chrome — the machine doors were walled: %s", wall)
	}
	status200 := 200
	return &webfetch.FetchResult{
		URL: u, FinalURL: u, Status: "full", HTTPStatus: &status200,
		ContentType: "text/html", Charset: "utf-8", Body: body, Text: text,
		Reason: reason,
	}
}

// TakeHandDelivered mirrors take_hand_delivered: the per-link check
// the pipeline makes BEFORE every fetch — is there a delivered page
// for THIS link? Returns a FetchResult (status 'full', the story in
// the reason) or nil. Consumption is the one-movement law (queue
// stamped, file kept). v0.57.0: allowConsumed re-reads a page whose
// queue row was already stamped consumed (the redo-consume — the
// delivered page is the RECORD; no new tab, no re-stamp).
func TakeHandDelivered(vaultPath, canonical string, log LogFn, allowConsumed bool) *webfetch.FetchResult {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" || canonical == "" {
		return nil
	}
	queue := ReadQueue(vaultPath)
	meta := queue.Meta(canonical)
	if meta == nil {
		// v0.64.0 — THE ONE SPELLING's compat read: a queue row left
		// under a legacy spelling still answers the canonical probe —
		// found by folding every key, and healed in place so the next
		// read pays no scan.
		foundKey := ""
		for _, k := range queue.Links.Keys() {
			if links.NormalizeWebsiteURL(k) == canonical {
				foundKey = k
				break
			}
		}
		if foundKey != "" {
			if raw, ok := queue.Links.Get(foundKey); ok {
				meta = queue.Meta(foundKey)
				if meta != nil && !dryrun.IsEnabled() {
					queue.Links.Delete(foundKey)
					queue.Links.Set(canonical, raw)
					_ = queue.WriteAtomic(vaultPath)
				}
			}
		}
	}
	if meta == nil {
		return nil
	}
	if metaTruthy(meta, "consumed") && !allowConsumed {
		return nil
	}
	suggested := metaStr(meta, "suggested")
	if suggested == "" {
		suggested = SuggestedFilename(canonical)
	}
	path := filepath.Join(HandDeliveryDir(vaultPath), suggested)
	st, err := os.Stat(path)
	if err != nil || st.IsDir() || st.Size() == 0 {
		return nil
	}
	body, err := os.ReadFile(path)
	if err != nil {
		log(fmt.Sprintf("⚠️ Hand-delivered page unreadable for %s: %v", canonical, err), "warning")
		return nil
	}
	if len(strings.TrimSpace(string(body))) == 0 {
		return nil
	}
	if metaStr(meta, "door") == "auto" {
		if ok, reason := autoPageIsReal(canonical, body); !ok {
			_ = os.Remove(path) // the app's own file — discarded
			log(fmt.Sprintf("⚠️ The auto-delivered page for %s was NOT the site (%s) — discarded (the app's own file), the link keeps waiting for its doors", canonical, reason), "warning")
			return nil
		}
	}
	if !dryrun.IsEnabled() && !metaTruthy(meta, "consumed") {
		meta.Set("consumed", stampNow())
		if err := queue.WriteAtomic(vaultPath); err != nil {
			log(fmt.Sprintf("⚠️ Hand-delivery queue could not be stamped: %v", err), "warning")
		}
	}
	return HandFetchResult(canonical, body, metaStr(meta, "wall"), metaStr(meta, "door"))
}

// StatusIsHand mirrors _status_is_hand: True when a gesture reads as
// the fourth door's (the owner's 🖐 hand). The precedence law: revived
// is checked first; then ANY hand marker wins (v0.52.0's combined
// gesture: a 🖐 icon over an 'unreviewed' Status cell is a HAND row).
// Death and reviewed are the CALLER's first checks (they retire, the
// stronger sentences).
func StatusIsHand(status string) bool {
	s := strings.ToLower(strings.TrimSpace(status))
	if s == "" {
		return false
	}
	for _, m := range []string{"♻️", "revived", "restored"} {
		if strings.Contains(s, m) {
			return false
		}
	}
	for _, m := range HandMarkers {
		if strings.Contains(s, m) {
			return true
		}
	}
	return false
}

// StampHandDeliveredRows mirrors stamp_hand_delivered_rows (v0.56.0 —
// THE HAND'S HARVEST): retire the 🖐 gesture into the green checkbox
// (✅ hand-delivered — fetched <date>) when its link's PROPER note has
// landed. Only a row whose Status still reads HAND (with death /
// reviewed / revived winning first — the precedence law) is
// rewritten. URLs are matched by canonical form. Pure file edit,
// atomic; never raises. Returns rows retired.
func StampHandDeliveredRows(vaultPath string, urls []string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" {
		return 0
	}
	path := decomtable.DecommissionTablePath(vaultPath)
	if _, err := os.Stat(path); err != nil {
		return 0
	}
	wanted := map[string]bool{}
	for _, u := range urls {
		c := links.NormalizeWebsiteURL(u)
		if c != "" {
			wanted[c] = true
		} else if u != "" {
			wanted[u] = true
		}
	}
	if len(wanted) == 0 {
		return 0
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log(fmt.Sprintf("⚠️ Hand-harvest stamp skipped — the table could not be read: %v", err), "warning")
		return 0
	}
	lines := pySplitLines(string(data))
	rows := decomtable.ParseDecommissionRows(lines)
	dateStr := dateNow()
	stamped := 0
	for _, row := range rows {
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			canonical = row.URL
		}
		if !wanted[canonical] && !wanted[row.URL] {
			continue
		}
		s := row.Status
		if decomtable.StatusIsDead(s) || decomtable.StatusIsReviewed(s) ||
			decomtable.StatusIsRevived(s) {
			continue // a stronger verdict owns the cell — never touched
		}
		if !StatusIsHand(s) {
			continue // not the gesture — the waiting family's writers own it
		}
		parts := strings.Split(row.Raw, "|")
		if len(parts) < 8 {
			continue
		}
		parts[6] = fmt.Sprintf(" ✅ hand-delivered — fetched %s ", dateStr)
		if row.Line < len(lines) {
			lines[row.Line] = strings.Join(parts, "|")
			stamped++
		}
	}
	if stamped == 0 {
		return 0
	}
	now := stampNow()
	for i, ln := range lines {
		if strings.HasPrefix(ln, "> Last updated:") {
			lines[i] = "> Last updated: " + now
		}
	}
	if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(lines, "\n"), "\n")+"\n"); err != nil {
		log(fmt.Sprintf("⚠️ Hand-harvest stamp could not be written: %v", err), "warning")
		return 0
	}
	log(fmt.Sprintf("✅ %d hand row(s) retired to the green checkbox (✅ hand-delivered) — their proper, categorized notes are in the vault", stamped), "info")
	return stamped
}

// pySplitLines mirrors Python str.splitlines(): splits on \n, \r\n
// (one boundary), \r, \v, \f, \x1c-\x1e, \x85, \u2028, \u2029; keeps
// empty interior lines; no trailing empty line.
func pySplitLines(s string) []string {
	var out []string
	start := 0
	i := 0
	for i < len(s) {
		r, size := rune(s[i]), 1
		if s[i] >= 0x80 {
			for _, rr := range string(s[i:]) {
				r, size = rr, len(string(rr))
				break
			}
		}
		switch {
		case r == '\r' && i+1 < len(s) && s[i+1] == '\n':
			out = append(out, s[start:i])
			i += 2
			start = i
		case r == '\n' || r == '\r' || r == '\v' || r == '\f' ||
			r == 0x1c || r == 0x1d || r == 0x1e || r == 0x85 ||
			r == 0x2028 || r == 0x2029:
			out = append(out, s[start:i])
			i += size
			start = i
		default:
			i += size
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}
