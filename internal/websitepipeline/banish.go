package websitepipeline

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/handdelivery"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// banish.go — v0.58.0 THE BANISHMENT (delete and never fetch again) +
// v0.60.0 THE CONFIRMATION GATE + v0.63.2 THE SETTLEMENT + the splits
// (website_pipeline.py:1788-3052). The vault is a garden, not an
// archive: the banishment is the DELIBERATE gesture — the marked notes
// leave the library for .trash/banished, their URLs are blacklisted
// (the never-fetch gate), and the record rows stay ♻️-revivable.

// BanishFM mirrors _parse_banish_frontmatter's dict (the notereads
// Banish shape — the cache and the local fallback speak the same
// type): source, managed_by (ownership), the tags LIST (flow style
// AND the block style Obsidian's property editor writes), and the
// banish boolean keys. Nil when the file has no frontmatter block or
// cannot be read. String-scan only — the house law.
type BanishFM = notereads.Banish

// ParseBanishFrontmatter mirrors _parse_banish_frontmatter.
func ParseBanishFrontmatter(path string) *BanishFM {
	head, ok := readHeadRunes(path, 200_000)
	if !ok {
		return nil
	}
	lines := pySplitLines(head)
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return nil
	}
	var tags []string
	var banish []notereads.KV
	out := &notereads.Banish{}
	inTagBlock := false
	for _, line := range lines[1:] {
		s := strings.TrimSpace(line)
		if s == "---" {
			break
		}
		if inTagBlock && strings.HasPrefix(s, "- ") {
			item := strings.TrimSpace(s[2:])
			item = strings.TrimSpace(stripQuotes(item))
			tags = append(tags, item)
			continue
		}
		inTagBlock = false
		if !strings.Contains(s, ":") {
			continue
		}
		key, val, _ := strings.Cut(s, ":")
		key = strings.ToLower(strings.TrimSpace(key))
		val = strings.TrimSpace(val)
		if key == "tags" {
			inner := val
			if strings.HasPrefix(inner, "[") && strings.HasSuffix(inner, "]") {
				inner = inner[1 : len(inner)-1]
			} else if inner == "" {
				inTagBlock = true // block style — items follow
			}
			for _, item := range strings.Split(inner, ",") {
				item = strings.TrimSpace(strings.TrimSpace(stripQuotes(item)))
				if item != "" {
					tags = append(tags, item)
				}
			}
			continue
		}
		clean := strings.TrimSpace(stripQuotes(val))
		if key == "decommission" || key == "banish" || key == "blacklist" {
			banish = append(banish, notereads.KV{Key: key, Value: clean})
		}
		if key == "source" || key == "managed_by" {
			if key == "source" {
				out.Source = clean
			} else {
				out.ManagedBy = clean
			}
		}
	}
	out.Tags = tags
	out.Banish = banish
	return out
}

// NoteReadsBanished mirrors _note_reads_banished: does a note's own
// frontmatter carry the owner's delete verdict? A tag counts when it
// CONTAINS the emoji (🗑️, "🗑️ delete") or IS a banish word exactly
// (the tag "deleted-files" never fires); a boolean key counts when its
// value reads true/yes/1/on. Returns (verdict, the marker that fired).
func NoteReadsBanished(fm *BanishFM) (bool, string) {
	if fm == nil {
		return false, ""
	}
	for _, t := range fm.Tags {
		s := strings.ToLower(strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(t), "#")))
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		if strings.Contains(s, decomtable.BanishEmoji) {
			return true, strings.TrimSpace(t)
		}
		for _, w := range decomtable.BanishWords {
			if s == w {
				return true, strings.TrimSpace(t)
			}
		}
	}
	for _, kv := range fm.Banish {
		key := strings.ToLower(strings.TrimSpace(kv.Key))
		isBanishKey := false
		for _, k := range decomtable.BanishKeys {
			if key == k {
				isBanishKey = true
				break
			}
		}
		if !isBanishKey {
			continue // a stranger's key is never our verdict
		}
		switch strings.ToLower(strings.TrimSpace(kv.Value)) {
		case "true", "yes", "1", "on":
			return true, fmt.Sprintf("%s: %s", kv.Key, kv.Value)
		}
	}
	return false, ""
}

// IsTrashFolder mirrors _is_trash_folder (v0.62.0): is this
// vault-relative folder path the owner's Trash? Root-level only (any
// spelling) and its subfolders. A SomeDir/Trash deep inside the tree
// is a category folder that happens to share the name — never the
// door. Pure predicate.
func IsTrashFolder(rel string) bool {
	rel = strings.ToLower(strings.TrimSpace(strings.ReplaceAll(strings.TrimSpace(rel), "\\", "/")))
	rel = strings.Trim(rel, "/")
	if rel == "" || rel == "." {
		return false
	}
	first := strings.SplitN(rel, "/", 2)[0]
	for _, name := range TrashFolderNames {
		if first == name {
			return true
		}
	}
	return false
}

// NoteBodyBanishTag mirrors _note_body_banish_tag (v0.60.1): the note
// BODY's own inline tags — a token counts when it is # + a non-space
// run (Obsidian's own law) preceded by a line start, whitespace, or an
// opening bracket, and the run contains 🗑️ or IS a banish word
// exactly (#deleted-files and #delete-me never fire). The frontmatter
// block and fenced code blocks are skipped. Pure string scan; never
// raises. Returns the marker as typed (#auto-delete) or ”.
func NoteBodyBanishTag(path string) string {
	head, ok := readHeadRunes(path, 200_000)
	if !ok {
		return ""
	}
	lines := pySplitLines(head)
	if len(lines) == 0 {
		return ""
	}
	// skip the frontmatter block (its own door already read it)
	start := len(lines) // unterminated block: all frontmatter
	if strings.TrimSpace(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				start = i + 1
				break
			}
		}
	}
	inFence := false
	for _, line := range lines[start:] {
		stripped := strings.TrimSpace(line)
		if strings.HasPrefix(stripped, "```") || strings.HasPrefix(stripped, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		runes := []rune(line)
		for i, ch := range runes {
			if ch != '#' {
				continue
			}
			if i+1 >= len(runes) {
				continue
			}
			nxt := runes[i+1]
			if nxt == ' ' || nxt == '\t' || nxt == '#' {
				continue // "# word" (a heading) or "##" — never a tag
			}
			prev := rune(0)
			if i > 0 {
				prev = runes[i-1]
			}
			if prev != 0 && prev != ' ' && prev != '\t' && prev != '[' && prev != '(' {
				continue // mid-word (C#, a URL's #fragment) — not a tag
			}
			j := i + 1
			for j < len(runes) && !strings.ContainsRune(BodyTagStops, runes[j]) && runes[j] != '#' {
				j++
			}
			token := strings.TrimSpace(string(runes[i+1 : j]))
			if token == "" {
				continue
			}
			low := strings.ToLower(token)
			if strings.Contains(low, decomtable.BanishEmoji) {
				return "#" + token
			}
			for _, w := range decomtable.BanishWords {
				if low == w {
					return "#" + token
				}
			}
		}
	}
	return ""
}

// BanishedNote mirrors scan_banished_notes' items:
// {'url', 'path', 'marker', 'door', 'app_owned'}.
type BanishedNote struct {
	URL      string
	Path     string
	Marker   string
	Door     string
	AppOwned bool
}

// ScanBanishedNotes mirrors scan_banished_notes (v0.58.0): every note
// in the library carrying the banish mark — the tag doors first (a
// tag's marker wins when the owner both marked and moved), the body's
// inline tags second, THE TRASH DOOR third (v0.62.0 — placement alone
// carries the verdict). Walks the vault the sweep walks (category
// folders, _review, _missing, _moc; _inbox record tables and
// dot-folders are skipped, so a note already resting in
// .trash/banished is never re-found). Hand-written marked notes ride
// the list with app_owned=false so the burial can keep + warn (the
// sacred law). v0.71.0 THE ONE-OPEN LAW: the optional readCache
// serves both parses from ONE bounded open per note. Pure file reads;
// no state DB, no network; never raises.
func ScanBanishedNotes(vaultPath string, log LogFn, readCache *notereads.Cache) []BanishedNote {
	if log == nil {
		log = noopLog
	}
	var out []BanishedNote
	if vaultPath == "" {
		return out
	}
	if _, err := os.Stat(vaultPath); err != nil || !isDirPath(vaultPath) {
		return out
	}
	walkVault(vaultPath, func(path, relRoot string) {
		fm := ParseBanishFrontmatter(path)
		if readCache != nil {
			func() {
				defer func() { recover() }()
				if cached := readCache.BanishFrontmatter(path); cached != nil {
					fm = cached
				}
			}()
		}
		if fm == nil {
			return // not a frontmattered note — a human's
		}
		marked, marker := NoteReadsBanished(fm)
		door := "note tag"
		if !marked {
			// v0.60.1 — the BODY's own inline tags
			marker = NoteBodyBanishTag(path)
			if readCache != nil {
				func() {
					defer func() { recover() }()
					if tag := readCache.BanishTag(path); tag != "" {
						marker = tag
					}
				}()
			}
			marked = marker != ""
		}
		inTrash := IsTrashFolder(relRoot)
		if !marked && inTrash {
			// v0.62.0 — THE TRASH DOOR: the move itself is the verdict
			marker = TrashMarker
			marked = true
			door = "trash folder"
		}
		if !marked {
			return
		}
		u := strings.TrimSpace(fm.Source)
		if !startsWithHTTP(u) {
			return // the directory note / tagless files
		}
		out = append(out, BanishedNote{
			URL: u, Path: path, Marker: marker, Door: door,
			AppOwned: strings.EqualFold(strings.TrimSpace(fm.ManagedBy), constants.ManagedByGitcurator),
		})
	})
	return out
}

// isDirPath reports a directory.
func isDirPath(p string) bool {
	st, err := os.Stat(p)
	return err == nil && st.IsDir()
}

// walkVault mirrors the reference's pruned os.walk: dirs sorted,
// '_inbox' and dot-folders skipped, files sorted, files of a directory
// handled BEFORE its subdirectories (os.walk's top-down order).
func walkVault(vaultPath string, visit func(path, relRoot string)) {
	_ = filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return nil // never raises — the walk's own law
		}
		if d.IsDir() {
			if path == vaultPath {
				return nil
			}
			name := d.Name()
			if name == "_inbox" || strings.HasPrefix(name, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(d.Name()), ".md") {
			return nil
		}
		if !isFilePath(path) {
			return nil
		}
		rel, relErr := filepath.Rel(vaultPath, filepath.Dir(path))
		if relErr != nil {
			rel = "."
		}
		rel = strings.ReplaceAll(rel, string(filepath.Separator), "/")
		if rel == "." {
			rel = ""
		}
		visit(path, rel)
		return nil
	})
}

func isFilePath(p string) bool {
	st, err := os.Stat(p)
	return err == nil && !st.IsDir()
}

// SplitResult mirrors the split_* twins' (kept, dropped) pair.
type SplitResult struct {
	Kept    []string
	Dropped []string
}

// SplitDismissedLinks mirrors split_dismissed_links (v0.60.1): links
// whose verdict is already written leave the batch's COUNT before
// processing starts, with one honest line. The state probe is guarded
// (a broken DB never hides a link); never raises.
func SplitDismissedLinks(state *StateDB, urls []string, log LogFn) SplitResult {
	if log == nil {
		log = noopLog
	}
	out := SplitResult{}
	for _, u := range urls {
		dropped := false
		func() {
			defer func() { recover() }()
			c := links.NormalizeWebsiteURL(u)
			if c != "" && state != nil {
				if dismissed, err := state.IsDismissed(c); err == nil && dismissed {
					dropped = true
				}
			}
		}()
		if dropped {
			out.Dropped = append(out.Dropped, u)
			continue
		}
		out.Kept = append(out.Kept, u)
	}
	if len(out.Dropped) > 0 {
		log(fmt.Sprintf("🪦 %d link(s) already have their verdict (decommissioned / banished / auto-retired) — excluded from the run's count before processing: never fetched again, never counted as unprocessed (♻️ in the master table revives any of them)", len(out.Dropped)), "info")
	}
	return out
}

// SplitSettledLinks mirrors split_settled_links (v0.63.2 — THE
// SETTLED LEDGER's own honesty): links already in the settled ledger
// leave the batch's count BEFORE processing starts.
func SplitSettledLinks(state *StateDB, urls []string, log LogFn) SplitResult {
	if log == nil {
		log = noopLog
	}
	out := SplitResult{}
	for _, u := range urls {
		settledHit := false
		func() {
			defer func() { recover() }()
			c := links.NormalizeWebsiteURL(u)
			if c != "" && state != nil {
				if settled, err := state.IsSettled(c); err == nil && settled {
					settledHit = true
				}
			}
		}()
		if settledHit {
			out.Dropped = append(out.Dropped, u)
			continue
		}
		out.Kept = append(out.Kept, u)
	}
	if len(out.Dropped) > 0 {
		log(fmt.Sprintf("🤝 %d link(s) are settled — already sent to the bot and addressed (the owner's law) — excluded from the run's count before processing: never fetched again (♻️ revived in the master table un-settles any of them)", len(out.Dropped)), "info")
	}
	return out
}

// SettleLedgerReport mirrors settle_the_ledger's dict:
// {'settled': n, 'retries_cleared': m, 'already': bool}.
type SettleLedgerReport struct {
	Settled        int
	RetriesCleared int
	Already        bool
}

// SettleTheLedger mirrors settle_the_ledger (v0.63.2 — THE
// SETTLEMENT, run at every production door): ONE time per machine —
// guarded by the websites_settled_at meta key — every website the
// system already knows is settled (never machine-fetched again). The
// ONE SPELLING's healing pass runs at every door and BEFORE the meta
// guard. queueUrls (the queue door's full bot history) runs THE
// QUEUE-HISTORY SETTLEMENT after the first pass. Never raises on the
// caller's head.
func SettleTheLedger(state *StateDB, vaultPath string, log LogFn, queueURLs []string) SettleLedgerReport {
	if log == nil {
		log = noopLog
	}
	// v0.64.0 — THE ONE SPELLING's healing pass, at every door and
	// BEFORE the meta guard. A broken heal never blocks the settlement.
	func() {
		defer func() { recover() }()
		if rep, err := state.NormalizeLedgerKeys(); err == nil && rep.Rekeyed > 0 {
			merged := ""
			if rep.Merged > 0 {
				merged = fmt.Sprintf(" (%d duplicate spelling(s) merged)", rep.Merged)
			}
			log(fmt.Sprintf("🩹 The one spelling: %d ledger row(s) re-keyed to their canonical URL form%s", rep.Rekeyed, merged), "info")
		}
	}()
	firstAlready := false
	func() {
		defer func() { recover() }()
		if v, err := state.GetMeta(SettledMetaKey); err == nil && v != nil {
			firstAlready = *v != ""
		}
	}()
	if firstAlready && queueURLs == nil {
		// the old shape, unchanged: this door has nothing new to do
		return SettleLedgerReport{Already: true}
	}
	rep := SettleLedgerReport{Already: true}
	if !firstAlready {
		var extra []string
		if vaultPath != "" && isDirPath(vaultPath) {
			if lines, ok := readTableLines(DecommissionTablePath(vaultPath)); ok {
				for _, row := range ParseDecommissionRows(lines) {
					extra = append(extra, row.URL)
				}
			}
			for _, it := range ScanReviewNotes(vaultPath, nil) {
				extra = append(extra, it.URL)
			}
		}
		for i, u := range extra {
			if u != "" {
				extra[i] = links.NormalizeWebsiteURL(u)
			}
		}
		func() {
			defer func() { recover() }()
			if r, err := state.SettleExisting(extra); err == nil {
				rep = SettleLedgerReport{
					Settled: r.Settled, RetriesCleared: r.RetriesCleared, Already: r.Already,
				}
			} else {
				log(fmt.Sprintf("⚠️ Settlement skipped: %v", err), "warning")
			}
		}()
		if !rep.Already {
			log(fmt.Sprintf("🤝 THE SETTLEMENT: %d website link(s) the bot already delivered are settled — addressed and processed, never fetched again; %d queued retry(ies) cleared. Only websites added from now on are fetched (♻️ revived in the master table un-settles any of them)", rep.Settled, rep.RetriesCleared), "info")
		}
	}
	// v0.64.1 — THE QUEUE-HISTORY SETTLEMENT, the queue door only.
	if queueURLs != nil {
		func() {
			defer func() { recover() }()
			if qh, err := state.SettleQueueHistory(queueURLs); err == nil {
				if !qh.Already && qh.Settled > 0 {
					log(fmt.Sprintf("🤝 THE QUEUE-HISTORY SETTLEMENT (websites): %d link(s) the bot already delivered are settled — they were never batched into a state row, so the queue kept counting them as pending; addressed and processed now, never counted or fetched again. Only websites added from now on are counted (♻️ revived in the master table un-settles any of them)", qh.Settled), "info")
				}
			} else {
				log(fmt.Sprintf("⚠️ Queue-history settlement skipped: %v", err), "warning")
			}
		}()
	}
	if firstAlready {
		return SettleLedgerReport{Already: true}
	}
	return rep
}

// PendingBanishment mirrors scan_pending_banishments' items:
// {'url', 'canonical', 'title', 'marker', 'door', 'path'}.
type PendingBanishment struct {
	URL       string
	Canonical string
	Title     string
	Marker    string
	Door      string
	Path      string
}

// PendingBanishmentsReport mirrors scan_pending_banishments' dict:
// {'items': [...], 'kept_handwritten': int}.
type PendingBanishmentsReport struct {
	Items           []PendingBanishment
	KeptHandwritten int
}

// ScanPendingBanishments mirrors scan_pending_banishments (v0.60.0 —
// THE GATE'S EYES): what WOULD be banished this run — both doors, one
// list, deduped by canonical URL. App-owned marked notes only (a
// hand-written marked note is the owner's to delete by hand, counted
// separately) + the master-table door (🗑-marked rows that are not yet
// confirmed-stamped). Pure file reads; never raises.
func ScanPendingBanishments(vaultPath string, log LogFn, readCache *notereads.Cache) PendingBanishmentsReport {
	if log == nil {
		log = noopLog
	}
	out := PendingBanishmentsReport{}
	if vaultPath == "" || !isDirPath(vaultPath) {
		return out
	}
	seen := map[string]bool{}
	for _, it := range ScanBanishedNotes(vaultPath, log, readCache) {
		if !it.AppOwned {
			out.KeptHandwritten++
			continue
		}
		canonical := links.NormalizeWebsiteURL(it.URL)
		if canonical == "" || seen[canonical] {
			continue
		}
		seen[canonical] = true
		title := canonical
		if it.Path != "" {
			base := baseName(it.Path)
			title = strings.TrimSuffix(base, filepath.Ext(base))
		}
		marker := it.Marker
		if marker == "" {
			marker = "delete mark"
		}
		door := it.Door
		if door == "" {
			door = "note tag"
		}
		out.Items = append(out.Items, PendingBanishment{
			URL: it.URL, Canonical: canonical, Title: title,
			Marker: marker, Door: door, Path: it.Path,
		})
	}
	table := map[string]string{}
	func() {
		defer func() { recover() }()
		table = ScanDecommissionTable(vaultPath)
	}()
	for u, status := range table {
		if !decomtable.StatusIsBanished(status) {
			continue
		}
		if strings.Contains(strings.ToLower(status), "confirmed") {
			continue // history, not a pending ask
		}
		canonical := links.NormalizeWebsiteURL(u)
		if canonical == "" || seen[canonical] {
			continue
		}
		seen[canonical] = true
		domain := canonical
		if parsed, err := parseURLHost(canonical); err == nil && parsed != "" {
			domain = parsed
		}
		marker := status
		if marker == "" {
			marker = ""
		}
		out.Items = append(out.Items, PendingBanishment{
			URL: u, Canonical: canonical, Title: domain,
			Marker: marker, Door: "master-table gesture", Path: "",
		})
	}
	return out
}

// BanishURLReport mirrors _banish_url's dict: {'dismissed', 'moved',
// 'row_forgotten', 'review_swept'}.
type BanishURLReport struct {
	Dismissed    bool
	Moved        string
	RowForgotten bool
	ReviewSwept  int
}

// BanishNoteFile mirrors _banish_note_file (v0.58.0): move ONE note
// file into <vault>/.trash/banished/ (Obsidian's hidden trash — out of
// the library, recoverable by hand). Dry-run aware. Returns the
// destination path, or ” when the move failed (the dismissal still
// holds — the gate is the DB, the file is the furniture).
func BanishNoteFile(vaultPath, path string, log LogFn) string {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" || path == "" {
		return ""
	}
	if _, err := os.Stat(path); err != nil {
		return ""
	}
	content, err := os.ReadFile(path)
	if err != nil {
		log(fmt.Sprintf("⚠️ could not move %s to %s: %v", baseName(path), BanishQuarantineRelpath, err), "warning")
		return ""
	}
	quarantine := filepath.Join(vaultPath, filepath.FromSlash(BanishQuarantineRelpath))
	dst := storage.UniquePath(filepath.Join(quarantine, baseName(path)))
	_ = dryrun.MkdirAll(quarantine, true)
	if err := dryrun.WriteText(dst, string(content)); err != nil {
		log(fmt.Sprintf("⚠️ could not move %s to %s: %v", baseName(path), BanishQuarantineRelpath, err), "warning")
		return ""
	}
	if err := dryrun.Remove(path); err != nil {
		log(fmt.Sprintf("⚠️ could not move %s to %s: %v", baseName(path), BanishQuarantineRelpath, err), "warning")
		return ""
	}
	// honesty (the graveyard's law): the move only counts when the
	// bytes actually moved — a dry-run rehearsal leaves the note in
	// place, so it reports '' and the caller says so.
	if isFilePath(dst) {
		if _, err := os.Stat(path); os.IsNotExist(err) {
			return dst
		}
	}
	return ""
}

// FindNoteForURL mirrors _find_note_for_url (v0.58.0): the note FILE
// for one canonical URL — the state ledger's row first (verified on
// disk), then a walk of the library for any app-owned note whose
// source normalizes to the canonical. ” when nothing is found (the
// burial proceeds DB-only). Never raises.
func FindNoteForURL(vaultPath, canonical string, state *StateDB, log LogFn) string {
	if log == nil {
		log = noopLog
	}
	if canonical == "" {
		return ""
	}
	if state != nil {
		var prior *ProcessedRow
		func() {
			defer func() { recover() }()
			prior, _ = state.ProcessedRow(canonical)
		}()
		if prior != nil && prior.NotePath != nil && isFilePath(*prior.NotePath) {
			path := *prior.NotePath
			fm := ParseReviewFrontmatter(path)
			if fm != nil && strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) &&
				SameSource(fm.Source, canonical) {
				return path
			}
		}
	}
	if vaultPath == "" || !isDirPath(vaultPath) {
		return ""
	}
	found := ""
	walkVault(vaultPath, func(path, relRoot string) {
		if found != "" {
			return
		}
		fm := ParseReviewFrontmatter(path)
		if fm == nil || !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
			return
		}
		if SameSource(fm.Source, canonical) {
			found = path
		}
	})
	return found
}

// BanishURL mirrors _banish_url (v0.58.0 — THE BURIAL CORE, both doors
// meet here): DB first — the never-fetch gate must hold even when
// every file operation below fails. Then the furniture: the note FILE
// leaves the library for .trash/banished, the processed row is
// forgotten (the ledger must not outlive the note it pointed at), and
// the link's half-fetched _review items go with it. Dry-run aware;
// never raises.
func BanishURL(state *StateDB, vaultPath, canonical, notePath, marker, gesture string, log LogFn) BanishURLReport {
	if log == nil {
		log = noopLog
	}
	report := BanishURLReport{}
	if canonical == "" {
		return report
	}
	reason := fmt.Sprintf("%s — 🗑️ %s (%s); the note left the library, the URL is blacklisted, never fetched again (v0.58.0)",
		BanishReasonPrefix, gesture, marker)
	if marker == "" {
		reason = fmt.Sprintf("%s — 🗑️ %s (delete mark); the note left the library, the URL is blacklisted, never fetched again (v0.58.0)",
			BanishReasonPrefix, gesture)
	}
	if err := state.Dismiss(canonical, reason); err == nil {
		report.Dismissed = true
	} else {
		log(fmt.Sprintf("⚠️ Banishment DB write skipped for %s: %v", canonical, err), "warning")
	}
	_ = state.ResolveRetry(canonical) // the dismissal is the gate; the queue row is furniture
	if notePath != "" && isFilePath(notePath) {
		dst := BanishNoteFile(vaultPath, notePath, log)
		report.Moved = dst
		if dst != "" {
			if _, err := os.Stat(notePath); os.IsNotExist(err) {
				if _, err := state.ForgetRow(canonical); err == nil {
					report.RowForgotten = true
				} else {
					log(fmt.Sprintf("⚠️ Could not forget the processed row for %s: %v", canonical, err), "warning")
				}
			}
		}
	} else {
		// the note is already gone (a hand deletion, an earlier
		// banishment) — a ledger row pointing at nothing is exactly the
		// false-success shape the owner reported; forget it.
		if _, err := state.ForgetRow(canonical); err == nil {
			report.RowForgotten = true
		}
	}
	report.ReviewSwept = SweepReviewLeftovers(vaultPath, canonical, "", log)
	return report
}

// BanishMarkedReport mirrors banish_marked_notes' dict.
type BanishMarkedReport struct {
	Marked          int
	Banished        int
	NotesMoved      int
	KeptHandwritten int
	ReviewSwept     int
	RowsWritten     int
	Skipped         int
	URLs            []string
}

// BanishMarkedNotes mirrors banish_marked_notes (v0.58.0 — THE
// BANISHMENT, the note-tag door): enforce the owner's delete verdicts
// written on the notes THEMSELVES. A hand-written note carrying the
// mark — or resting in the Trash folder — is KEPT with a warning (the
// sacred law). v0.73.0: skipCanonicals holds UNchecked/locked URLs out
// of the burial. Idempotent; dry-run aware; tolerated everywhere.
func BanishMarkedNotes(state *StateDB, vaultPath string, log LogFn, skipCanonicals map[string]bool) BanishMarkedReport {
	if log == nil {
		log = noopLog
	}
	report := BanishMarkedReport{}
	if vaultPath == "" || !isDirPath(vaultPath) {
		return report
	}
	items := ScanBanishedNotes(vaultPath, log, nil)
	if len(items) == 0 {
		return report // the cheap no-op — most runs
	}
	skip := skipCanonicals
	hasSkip := skip != nil
	dateStr := dateNow()
	seen := map[string]bool{}
	var recordURLs []string
	recordRowsTag := []string{}
	recordNotesTag := map[string]string{}
	recordRowsTrash := []string{}
	recordNotesTrash := map[string]string{}
	for _, it := range items {
		report.Marked++
		viaTrash := it.Door == "trash folder"
		if !it.AppOwned {
			report.KeptHandwritten++
			if viaTrash {
				log(fmt.Sprintf("✍️ kept %s — it rests in the Trash folder but is hand-written (yours); the app never deletes what it did not write — delete it by hand in Obsidian if you want it gone", baseName(it.Path)), "warning")
			} else {
				log(fmt.Sprintf("✍️ kept %s — it carries the 🗑️ mark but is hand-written (yours); delete it by hand in Obsidian if you want it gone", baseName(it.Path)), "warning")
			}
			continue
		}
		canonical := links.NormalizeWebsiteURL(it.URL)
		if canonical == "" || seen[canonical] {
			continue // two marked notes, one link — one burial
		}
		if hasSkip && skip[canonical] {
			// v0.73.0 — THE PICKING DOOR / THE LOCK LAW's skip
			seen[canonical] = true
			report.Skipped++
			log(fmt.Sprintf("🔒 %s: banishment skipped this run (unchecked or locked) — the mark stays, the note stays", canonical), "info")
			continue
		}
		seen[canonical] = true
		gesture := "note tag"
		if viaTrash {
			gesture = TrashGesture
		}
		marker := it.Marker
		b := BanishURL(state, vaultPath, canonical, it.Path, marker, gesture, log)
		report.ReviewSwept += b.ReviewSwept
		if !b.Dismissed {
			continue // the DB refused — the verdict is not law
		}
		report.Banished++
		recordURLs = append(recordURLs, canonical)
		if viaTrash {
			recordRowsTrash = append(recordRowsTrash, canonical)
			recordNotesTrash[canonical] = "🗑️ you moved it to the Trash folder"
		} else {
			recordRowsTag = append(recordRowsTag, canonical)
			m := marker
			if m == "" {
				m = "delete"
			}
			recordNotesTag[canonical] = fmt.Sprintf("🗑️ marked on the note: %s", m)
		}
		if b.Moved != "" {
			report.NotesMoved++
			if viaTrash {
				log(fmt.Sprintf("🗑️ %s: banished — you moved it to the Trash folder; it left the library for %s and the URL is blacklisted (never fetched again; ♻️ revived in the master table undoes it)", canonical, BanishQuarantineRelpath), "info")
			} else {
				log(fmt.Sprintf("🗑️ %s: banished — the note carried the delete mark; it left the library for %s and the URL is blacklisted (never fetched again; ♻️ revived in the master table undoes it)", canonical, BanishQuarantineRelpath), "info")
			}
		} else if dryrun.IsEnabled() {
			log(fmt.Sprintf("🗑️ %s: banishment REHEARSED (dry-run) — the URL is blacklisted in the shadow cache and the note's move to %s was recorded, not performed", canonical, BanishQuarantineRelpath), "info")
		} else {
			log(fmt.Sprintf("🗑️ %s: banished — the URL is blacklisted (never fetched again); the note file could not be moved, it stays where it is (an orphan now — the next run retries the move)", canonical), "warning")
		}
	}
	if len(recordURLs) > 0 {
		written := 0
		if len(recordRowsTag) > 0 {
			written += WriteDecommissionCandidates(vaultPath, recordRowsTag, "note tag", log,
				recordNotesTag, fmt.Sprintf("🗑️ banished — confirmed %s", dateStr))
		}
		if len(recordRowsTrash) > 0 {
			written += WriteDecommissionCandidates(vaultPath, recordRowsTrash, "Trash folder", log,
				recordNotesTrash, fmt.Sprintf("🗑️ banished — confirmed %s", dateStr))
		}
		report.RowsWritten = written
		log(fmt.Sprintf("🗑️ Banishment: %d note(s) removed from the library and blacklisted — %d hand-written marked note(s) kept (yours); the master table holds the record rows (♻️ revivable)", report.Banished, report.KeptHandwritten), "info")
	} else if report.KeptHandwritten > 0 {
		log(fmt.Sprintf("🗑️ Banishment: %d marked note(s) are hand-written and were KEPT (yours) — nothing banished", report.KeptHandwritten), "info")
	}
	report.URLs = recordURLs
	return report
}

// BanishTwinsInOtherVault mirrors banish_twins_in_other_vault
// (v0.60.0 — the "both vaults" promise of the confirmed deletion):
// after the owner CONFIRMS, walk the other vault (the config vault_path
// — the GitHub-projects vault) for APP-OWNED notes whose source
// normalizes into canonicals and move each to that vault's own
// .trash/banished (hand-written notes are never touched — the walk
// only ever moves notes the app wrote). Dry-run aware; never raises.
func BanishTwinsInOtherVault(otherVault string, canonicals []string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	moved := 0
	if otherVault == "" || !isDirPath(otherVault) || len(canonicals) == 0 {
		return moved
	}
	wanted := map[string]bool{}
	for _, u := range canonicals {
		c := links.NormalizeWebsiteURL(u)
		if c != "" {
			wanted[c] = true
		}
	}
	if len(wanted) == 0 {
		return moved
	}
	walkVault(otherVault, func(path, relRoot string) {
		fm := ParseReviewFrontmatter(path)
		if fm == nil || !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
			return
		}
		if !wanted[links.NormalizeWebsiteURL(fm.Source)] {
			return
		}
		if dst := BanishNoteFile(otherVault, path, log); dst != "" {
			moved++
			log(fmt.Sprintf("🗑️ %s: twin note removed from the other vault too (%s → %s) — the confirmed deletion reaches both vaults", fm.Source, baseName(path), BanishQuarantineRelpath), "info")
		}
	})
	return moved
}

// ConsumeReport mirrors consume_decommission_table's dict.
type ConsumeReport struct {
	Dead              int
	Reviewed          int
	Banished          int
	Revived           int
	Handed            int
	PlaceholdersSwept int
	NotesMoved        int
	RowsConfirmed     int
	BanishedURLs      []string
	PendingBanish     int
	Skipped           int
}

// ConsumeDecommissionTable mirrors consume_decommission_table
// (v0.44.0/v0.47.0/v0.48.0/v0.58.0 — the owner's master-table
// decisions enforced): pass 0 the banishments (THE STRONGEST
// retirement; with applyBanish=false the 🗑 rows are only COUNTED — the
// confirmation gate's hold), pass 1 the retirements (dead + reviewed:
// dismiss + sweep + confirm), pass 1.5 the fourth door's gestures
// (queue for hand-delivery), pass 2 the revivals (undismiss +
// unsettle), pass 3 the placeholder sweep, pass 4 the write-back.
// Idempotent; tolerated everywhere. banished_urls is the list of
// verdicts WRITTEN this run (the v0.60.2 honesty law).
func ConsumeDecommissionTable(state *StateDB, vaultPath string, log LogFn,
	applyBanish bool, skipCanonicals map[string]bool) ConsumeReport {
	if log == nil {
		log = noopLog
	}
	report := ConsumeReport{}
	path := DecommissionTablePath(vaultPath)
	if vaultPath == "" {
		return report
	}
	if _, err := os.Stat(path); err != nil {
		return report
	}
	hasSkip := skipCanonicals != nil
	// v0.52.0 — a trimmed table gets its full grammar back first.
	func() {
		defer func() { recover() }()
		EnsureTableShape(path, log)
	}()
	lines, ok := readTableLines(path)
	if !ok {
		return report
	}
	rows := ParseDecommissionRows(lines)
	if len(rows) == 0 {
		return report
	}
	outLines := lines
	haveLines := false
	dateStr := dateNow()
	var retiredCanonicals []string

	// ---- pass 0 (v0.58.0): the banishments ---------------------------
	var banishedCanonicals []string
	var banishedConsumed []string
	for _, row := range rows {
		if !decomtable.StatusIsBanished(row.Status) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			continue
		}
		already := false
		for _, c := range banishedCanonicals {
			if c == canonical {
				already = true
				break
			}
		}
		if already {
			continue
		}
		banishedCanonicals = append(banishedCanonicals, canonical)
		if hasSkip && skipCanonicals[canonical] {
			report.PendingBanish++
			report.Skipped++
			log(fmt.Sprintf("🔒 %s: the master-table banishment waits (unchecked or locked) — the mark stays, the note stays", row.URL), "info")
			continue
		}
		if !applyBanish {
			report.PendingBanish++
			continue // the gate's hold — nothing consumed, nothing moved
		}
		report.Banished++
		banishedConsumed = append(banishedConsumed, canonical)
		notePath := FindNoteForURL(vaultPath, canonical, state, log)
		b := BanishURL(state, vaultPath, canonical, notePath, row.Status, "master-table gesture", log)
		if b.Moved != "" {
			report.NotesMoved++
		}
		// Confirm the row (byte-stable rewrite — the marker stays).
		parts := strings.Split(row.Raw, "|")
		if len(parts) >= 8 && !strings.Contains(strings.ToLower(row.Status), "confirmed") {
			parts[6] = fmt.Sprintf(" 🗑️ banished — confirmed %s ", dateStr)
			if !haveLines {
				if fresh, ok := readTableLines(path); ok {
					outLines = fresh
				} else {
					outLines = nil
				}
				haveLines = true
			}
			if outLines != nil && row.Line < len(outLines) {
				outLines[row.Line] = strings.Join(parts, "|")
				report.RowsConfirmed++
			}
		}
	}

	// ---- pass 1: the retirements (DB first — the gate must hold) ----
	for _, row := range rows {
		if decomtable.StatusIsBanished(row.Status) {
			continue // v0.58.0 — the banishment owns this row
		}
		dead := decomtable.StatusIsDead(row.Status)
		reviewed := !dead && decomtable.StatusIsReviewed(row.Status)
		if !dead && !reviewed {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			continue
		}
		retiredCanonicals = append(retiredCanonicals, canonical)
		var reason, stamp string
		if dead {
			report.Dead++
			reason = "decommissioned by owner — graveyard table (v0.44.0)"
			stamp = fmt.Sprintf(" 🪦 confirmed — decommissioned %s ", dateStr)
		} else {
			report.Reviewed++
			reason = "retired by owner (reviewed ✅) — the master table (v0.47.0); never fetched again"
			stamp = fmt.Sprintf(" ✅ confirmed — reviewed %s ", dateStr)
		}
		if err := state.Dismiss(canonical, reason); err != nil {
			log(fmt.Sprintf("⚠️ Master-table DB write skipped for %s: %v", row.URL, err), "warning")
		}
		_ = state.ResolveRetry(canonical)
		_, _ = state.ForgetFailedRow(canonical)
		// Confirm the row (byte-stable rewrite — the marker stays).
		parts := strings.Split(row.Raw, "|")
		if len(parts) >= 8 && !strings.Contains(strings.ToLower(row.Status), "confirmed") {
			parts[6] = stamp
			if !haveLines {
				if fresh, ok := readTableLines(path); ok {
					outLines = fresh
				} else {
					outLines = nil
				}
				haveLines = true
			}
			if outLines != nil && row.Line < len(outLines) {
				outLines[row.Line] = strings.Join(parts, "|")
				report.RowsConfirmed++
			}
		}
	}

	// ---- pass 1.5 (v0.48.0): the fourth door's gestures -------------
	var handURLs []string
	for _, row := range rows {
		s := row.Status
		if decomtable.StatusIsDead(s) || decomtable.StatusIsRevived(s) || decomtable.StatusIsReviewed(s) {
			continue
		}
		if !handStatusIsHand(s) {
			continue
		}
		if strings.Contains(strings.ToLower(s), "queued") {
			continue // already stamped by a previous pass
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			continue
		}
		handURLs = append(handURLs, canonical)
		report.Handed++
		parts := strings.Split(row.Raw, "|")
		if len(parts) >= 8 {
			parts[6] = fmt.Sprintf(" 🖐 hand — queued %s ", dateStr)
			if !haveLines {
				if fresh, ok := readTableLines(path); ok {
					outLines = fresh
				} else {
					outLines = nil
				}
				haveLines = true
			}
			if outLines != nil && row.Line < len(outLines) {
				outLines[row.Line] = strings.Join(parts, "|")
			}
		}
	}
	if len(handURLs) > 0 {
		handdelivery.EnqueueHandDelivery(vaultPath, handURLs, nil, log, "")
		log(fmt.Sprintf("🖐 %d hand row(s) queued — your real Chrome opens for them at the end of this run (one tab per link, the live pages taken as real fetches; saving a page into the hand-delivered folder yourself still works too)", report.Handed), "info")
	}

	// ---- pass 2: the revivals ----------------------------------------
	for _, row := range rows {
		if !decomtable.StatusIsRevived(row.Status) {
			continue
		}
		canonical := links.NormalizeWebsiteURL(row.URL)
		if canonical == "" {
			continue
		}
		report.Revived++
		reviveParts := []string{}
		if gone, err := state.Undismiss(canonical); err == nil && gone {
			reviveParts = append(reviveParts, "the dismissal is gone")
		}
		if gone, err := state.Unsettle(canonical); err == nil && gone {
			reviveParts = append(reviveParts, "the settlement is gone")
		}
		if len(reviveParts) > 0 {
			log(fmt.Sprintf("♻️ %s: revived — %s, it will be fetched like new", row.URL, strings.Join(reviveParts, " and ")), "info")
		}
	}

	// ---- pass 3: sweep the retired links' placeholder FILES ----------
	if len(retiredCanonicals) > 0 {
		retiredSet := map[string]bool{}
		for _, c := range retiredCanonicals {
			retiredSet[c] = true
		}
		reviewDir := filepath.Join(vaultPath, ReviewFolder)
		entries, err := os.ReadDir(reviewDir)
		if err == nil {
			names := make([]string, 0, len(entries))
			for _, e := range entries {
				names = append(names, e.Name())
			}
			sort.Strings(names)
			for _, name := range names {
				if !strings.HasSuffix(strings.ToLower(name), ".md") || name == DecommissionTable {
					continue
				}
				p := filepath.Join(reviewDir, name)
				if !isFilePath(p) {
					continue
				}
				fm := ParseReviewFrontmatter(p)
				if fm == nil || !strings.EqualFold(fm.ManagedBy, constants.ManagedByGitcurator) {
					continue // a human's note — never our call
				}
				src := links.NormalizeWebsiteURL(strings.TrimSpace(fm.Source))
				if !retiredSet[src] {
					continue
				}
				if err := dryrun.Remove(p); err == nil {
					// dry-run records the sweep without deleting — the
					// count stays honest (a rehearsed burial sweeps nothing)
					if _, statErr := os.Stat(p); os.IsNotExist(statErr) {
						report.PlaceholdersSwept++
					}
				} else {
					log(fmt.Sprintf("⚠️ Could not sweep placeholder %s: %v", name, err), "warning")
				}
			}
		}
	}

	// ---- pass 4: write the confirmed rows back ----------------------
	if haveLines && outLines != nil {
		now := stampNow()
		for i, line := range outLines {
			if strings.HasPrefix(line, "> Last updated:") {
				outLines[i] = "> Last updated: " + now
			}
		}
		if err := storage.AtomicWriteText(path, strings.TrimRight(strings.Join(outLines, "\n"), "\n")+"\n"); err != nil {
			log(fmt.Sprintf("⚠️ Could not write the master table: %v", err), "warning")
		}
	}

	if report.Banished > 0 {
		log(fmt.Sprintf("🗑️ Master table: %d link(s) banished — %d note(s) removed from the library (.trash/banished), URLs blacklisted, never fetched again (♻️ revived undoes any)", report.Banished, report.NotesMoved), "info")
	}
	if report.Dead > 0 {
		log(fmt.Sprintf("🪦 Master table: %d link(s) decommissioned — dismissed, %d placeholder(s) swept, never fetched again", report.Dead, report.PlaceholdersSwept), "info")
	}
	if report.Reviewed > 0 {
		log(fmt.Sprintf("✅ Master table: %d link(s) retired as reviewed — dismissed, never fetched again (♻️ revived brings any back)", report.Reviewed), "info")
	}
	report.BanishedURLs = banishedConsumed
	return report
}

// parseURLHost extracts the netloc (URL.Host) for the scan's title
// fallback (urlparse(canonical).netloc or canonical).
func parseURLHost(u string) (string, error) {
	parsed, err := url.Parse(u)
	if err != nil {
		return "", err
	}
	if parsed.Host == "" {
		return "", fmt.Errorf("no host")
	}
	return parsed.Host, nil
}
