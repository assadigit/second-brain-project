package vaultscan

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/repobanish"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
)

// StampCurated mirrors stamp_curated (v0.78.0 — THE CENSUS STAMP's
// writer): 'curated: true' into the note's frontmatter, riding the
// house's own frontmatter surgery (notestate.Fmn) behind the atomic
// writer. A note without a well-formed frontmatter block is returned
// UNCHANGED (never invent a block the note never had). Idempotent.
// Dry-run aware. Never raises.
func StampCurated(path string, log LogFn) bool {
	if log == nil {
		log = noopLog
	}
	if path == "" || !fsIsFile(path) {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	content := helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
	stamped, changed := notestate.FmReplaceField(content, "curated", "curated: true")
	if !changed {
		return false // already stamped (or no frontmatter)
	}
	if err := storage.AtomicWriteText(path, stamped); err != nil {
		if !dryrun.IsEnabled() {
			log(fmt.Sprintf("⚠️ could not stamp %s as curated: %s",
				filepath.Base(path), err), "warning")
		}
		return false
	}
	return true
}

// RetireNoteFile mirrors _retire_note_file (v0.64.0): move ONE note
// file into <vault>/.trash/<subdir>/ (bytes read, written at a unique
// destination, source removed — dry-run aware, honest about what
// actually moved). Returns the destination path, or ” when the move
// did not land.
func RetireNoteFile(vaultPath, path, subdir string, log LogFn) string {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" || path == "" || !fsIsFile(path) {
		return ""
	}
	dstDir := filepath.Join(vaultPath, ".trash", subdir)
	if subdir == "" {
		dstDir = filepath.Join(vaultPath, ".trash", "retired")
	}
	result := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				log(fmt.Sprintf("⚠️ could not move %s to .trash/%s: %v",
					filepath.Base(path), subdir, r), "warning")
			}
		}()
		data, err := os.ReadFile(path)
		if err != nil {
			panic(err) // caught by the recover → the warning line
		}
		content := helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
		dst := storage.UniquePath(filepath.Join(dstDir, filepath.Base(path)))
		if err := dryrun.MkdirAll(dstDir, true); err != nil {
			panic(err)
		}
		if err := dryrun.WriteText(dst, content); err != nil {
			panic(err)
		}
		if err := dryrun.Remove(path); err != nil {
			panic(err)
		}
		if fsIsFile(dst) && !fsExists(path) {
			result = dst
		}
	}()
	return result
}

// Selection mirrors the picking door's checked ids: nil for a section
// means the section rides WHOLE (the Telegram door's all-or-nothing
// law); an explicitly EMPTY slice means nothing in that section rides.
type Selection struct {
	Deletions      map[string]bool
	Consolidations map[string]bool
	NewFolders     map[string]bool
	Moves          map[string]bool
}

// NormalizeSelection mirrors normalize_selection (v0.73.0 — THE
// PICKING DOOR's shape law, PURE): the checked ids of each section.
// A nil section rides WHOLE; junk reads as its section's nil — never a
// crash, never a surprise move. Never raises.
func NormalizeSelection(selection map[string][]string) Selection {
	out := Selection{}
	if selection == nil {
		return out
	}
	fill := func(val []string) map[string]bool {
		if val == nil {
			return nil
		}
		m := map[string]bool{}
		for _, v := range val {
			if helpers.Strip(v) != "" {
				m[v] = true
			}
		}
		return m
	}
	out.Deletions = fill(selection["deletions"])
	out.Consolidations = fill(selection["consolidations"])
	out.NewFolders = fill(selection["new_folders"])
	out.Moves = fill(selection["moves"])
	return out
}

// picked mirrors _picked: True when no selection rides for the
// section (nil — the whole plan) or the row's id is checked.
func picked(rowID string, sel map[string]bool) bool {
	if sel == nil {
		return true
	}
	return sel[rowID]
}

// PickedCounts mirrors picked_counts (v0.73.0 — the picking door's
// honest arithmetic, PURE): how many rows of each section the owner
// checked. A nil selection = the whole plan.
func PickedCounts(plan *Plan, sel Selection) map[string]int {
	out := map[string]int{
		"deletions": 0, "new_folders": 0, "moves": 0,
		"consolidation_notes": 0, "consolidations": 0,
	}
	if plan == nil {
		return out
	}
	for _, d := range plan.Deletions {
		id := d.Canonical
		if id == "" {
			id = d.URL
		}
		if picked(id, sel.Deletions) {
			out["deletions"]++
		}
	}
	for _, f := range plan.NewFolders {
		if picked(f, sel.NewFolders) {
			out["new_folders"]++
		}
	}
	for _, m := range plan.Moves {
		if picked(m.Path, sel.Moves) {
			out["moves"]++
		}
	}
	for _, c := range plan.Consolidations {
		anyPicked := false
		for _, m := range c.Merge {
			if picked(m.Path, sel.Consolidations) {
				out["consolidation_notes"]++
				anyPicked = true
			}
		}
		if anyPicked {
			out["consolidations"]++
		}
	}
	return out
}

// FailedMove mirrors the report's failed ledger rows:
// {'note', 'to', 'why'}.
type FailedMove struct {
	Note string
	To   string
	Why  string
}

// MovedRow mirrors the report's moved rows:
// {'note', 'from', 'to', 'as'}.
type MovedRow struct {
	Note string
	From string
	To   string
	As   string
}

// ConsolidationRecord mirrors the report's consolidations rows:
// {'site', 'kept', 'retired'}.
type ConsolidationRecord struct {
	Site    string
	Kept    string
	Retired string
}

// ApplyReport mirrors apply_scan_plan's report:
// {'folders_created', 'notes_moved', 'moved', 'failed_moves',
// 'banished', 'consolidated', 'consolidations', 'stamped', 'failed'}.
type ApplyReport struct {
	FoldersCreated int
	NotesMoved     int
	Moved          []MovedRow
	FailedMoves    int
	Banished       int
	Consolidated   int
	Consolidations []ConsolidationRecord
	Stamped        int
	Failed         []FailedMove
}

// rememberFailed mirrors _remember_failed (v0.82.0 — the landable
// move's honesty ledger): every failed move is remembered BY NAME and
// destination. Pure; never raises.
func rememberFailed(report *ApplyReport, m *MoveRow, destRel, why string) {
	note := filepath.Base(m.Path)
	if note == "" || note == "." {
		note = m.Note
	}
	report.Failed = append(report.Failed, FailedMove{
		Note: note, To: destRel, Why: why})
}

// repointRepoLedger mirrors _repoint_repo_ledger (v0.83.0 — THE REPOS
// LEDGER RE-POINT): one moved or merged repo note's processed_repos
// row follows its file. The row is fetched by canonical URL and
// re-written with the new note path and the destination's CATEGORY
// key. No row, no re-point. Returns true when a row actually moved.
func repointRepoLedger(cacheDB *store.CacheDB, canonical, newPath,
	folderOrKey string) bool {
	result := false
	func() {
		defer func() { recover() }() // the reference's try/except
		row := cacheDB.ProcessedRowFor(canonical)
		if row == nil {
			return
		}
		key := helpers.Strip(folderOrKey)
		if key != "" && (strings.Contains(key, "/") ||
			strings.Contains(key, "\\")) {
			key = reposCategoryReverse()[strings.Trim(
				strings.ReplaceAll(key, "\\", "/"), "/")]
			// an unmapped folder keeps the row's own word (the note's
			// frontmatter is its own)
		}
		if key == "" {
			if row.Category != nil {
				key = *row.Category
			}
		}
		repoID := row.RepoID
		url := canonical
		if row.URL != nil && *row.URL != "" {
			url = *row.URL
		}
		owner := ""
		if row.Owner != nil {
			owner = *row.Owner
		}
		repoName := ""
		if row.RepoName != nil {
			repoName = *row.RepoName
		}
		cacheDB.AddProcessed(repoID, url, owner, repoName, newPath, key)
		result = true
	}()
	return result
}

// ApplyScanPlan mirrors apply_scan_plan (v0.61.0 — the confirmed
// plan's hands): the consolidations (links appended, files retired to
// .trash/consolidated, ledgers re-pointed), the picked new folders,
// the moves (lock re-read at apply; the 'as' bend honored; verified
// landings — a vanished source or a did-not-land is SPOKEN in the
// failed ledger), the census stamps (curated: true on moved notes at
// their new home, on declined rows when a selection spoke, on
// reviewed), and the deletions through the banishment machinery's own
// doors (skip_canonicals / the repos items parameter). Dry-run aware;
// never raises.
func ApplyScanPlan(plan *Plan, vaultPath string,
	state *websitepipeline.WebsiteStateDB, log LogFn,
	sel Selection, kind string, cacheDB *store.CacheDB,
	noteStateDB *notestate.DB) ApplyReport {
	if log == nil {
		log = noopLog
	}
	report := ApplyReport{}
	if vaultPath == "" || !fsIsDir(vaultPath) {
		return report
	}
	if plan == nil {
		plan = &Plan{}
	}
	// ---- v0.64.0 — the consolidations (ONE NOTE PER SITE) -----------
	type consJob struct {
		c         Consolidation
		keep      KeepRow
		keepPath  string
		mergeRows []MergeRow
	}
	var consolidations []consJob
	for _, c := range plan.Consolidations {
		keepPath := c.Keep.Path
		var mergeRows []MergeRow
		for _, m := range c.Merge {
			if !picked(m.Path, sel.Consolidations) {
				continue // unchecked — the duplicate stays put
			}
			mergeRows = append(mergeRows, m)
		}
		if len(mergeRows) == 0 {
			continue // nothing picked — the whole consolidation waits
		}
		consolidations = append(consolidations, consJob{
			c: c, keep: c.Keep, keepPath: keepPath, mergeRows: mergeRows})
	}
	for _, job := range consolidations {
		c, keep, keepPath, mergeRows := job.c, job.keep, job.keepPath, job.mergeRows
		site := c.Site
		if site == "" {
			site = "the site"
		}
		if keepPath == "" || !fsIsFile(keepPath) {
			log(fmt.Sprintf("⚠️ Scan plan: the consolidation's kept note "+
				"is gone — nothing merged for %s", site), "warning")
			continue
		}
		// v0.73.0 — THE LOCK LAW's apply-time guard: a keeper locked
		// while the plan stood open is never rewritten.
		if noteLocked(frontmatterOf(keepPath)) {
			log(fmt.Sprintf("🔒 %s (the keeper for %s) is locked — the "+
				"consolidation waits (the locked note is never "+
				"rewritten; unlock it to let a future scan merge)",
				filepath.Base(keepPath), site), "info")
			continue
		}
		var urls []string
		for _, m := range mergeRows {
			urls = append(urls, m.Source)
		}
		canonOf := func(u string) string { return links.NormalizeWebsiteURL(u) }
		if isReposKind(kind) {
			canonOf = repoURLKey
		}
		kept := []string{}
		for _, u := range urls {
			if canonOf(u) != canonOf(keep.Source) {
				kept = append(kept, u)
			}
		}
		urls = kept
		func() {
			defer func() { recover() }() // the reference's try/except
			websitepipeline.AddSiteLinksToNote(keepPath, urls, log)
		}()
		for _, m := range mergeRows {
			mpath := m.Path
			if mpath == "" || !fsIsFile(mpath) {
				continue // already gone — a hand deletion mid-plan
			}
			// the lock law's apply-time guard: a duplicate locked while
			// the plan stood open never retires.
			if noteLocked(frontmatterOf(mpath)) {
				log(fmt.Sprintf("🔒 %s is locked — the merge is dropped "+
					"(the duplicate stays; unlock it to let a future "+
					"scan merge)", filepath.Base(mpath)), "info")
				continue
			}
			dst := RetireNoteFile(vaultPath, mpath, "consolidated", log)
			if dst != "" || dryrun.IsEnabled() {
				report.Consolidated++
				s := c.Site
				if s == "" {
					s = ""
				}
				report.Consolidations = append(report.Consolidations,
					ConsolidationRecord{
						Site:    s,
						Kept:    filepath.Base(keepPath),
						Retired: filepath.Base(mpath)})
				log(fmt.Sprintf("🧲 %s merged into %s (site: %s) — the "+
					"duplicate rests in .trash/consolidated, its link "+
					"rides the kept note", filepath.Base(mpath),
					filepath.Base(keepPath), siteName(c.Site)), "info")
			}
		}
		// the ledger: every merged note's URL points at the kept note
		// and its retry row resolves.
		for _, m := range mergeRows {
			if isReposKind(kind) {
				if cacheDB == nil {
					break
				}
				func() {
					defer func() { recover() }() // bookkeeping never fails
					canon := repoURLKey(m.Source)
					if canon == "" {
						return
					}
					repointRepoLedger(cacheDB, canon, keepPath, keep.Category)
					cacheDB.ResolveFailedURLs([]string{canon})
				}()
				continue
			}
			if state == nil {
				break
			}
			func() {
				defer func() { recover() }() // bookkeeping never fails
				canon := links.NormalizeWebsiteURL(m.Source)
				if canon == "" {
					return
				}
				_ = state.MarkProcessed(canon, keepPath, keep.Category,
					keep.Subcategory, "full")
				_ = state.ResolveRetry(canon)
			}()
		}
	}
	// ---- the filing: folders first, then the moves -------------------
	for _, rel := range plan.NewFolders {
		if !picked(rel, sel.NewFolders) {
			continue // unchecked, not born (a CHECKED move that files
			// into it still creates it)
		}
		clean := SanitizeFolder(rel)
		if clean == "" {
			continue
		}
		dst := filepath.Join(append([]string{vaultPath},
			strings.Split(clean, "/")...)...)
		func() {
			defer func() { recover() }()
			if err := dryrun.MkdirAll(dst, true); err == nil {
				if fsIsDir(dst) || dryrun.IsEnabled() {
					report.FoldersCreated++
					would := " "
					if dryrun.IsEnabled() {
						would = " would be "
					}
					log(fmt.Sprintf("📁 New folder%screated: %s", would,
						clean), "info")
				}
			} else {
				log(fmt.Sprintf("⚠️ Could not create the folder %s: %v",
					clean, err), "warning")
			}
		}()
	}
	for i := range plan.Moves {
		m := &plan.Moves[i]
		if !picked(m.Path, sel.Moves) {
			// v0.78.0 — the census stamp's decline door: a move row the
			// owner UNCHECKED received its verdict — "no" — and is
			// stamped in place so the next scan skips it. Only when the
			// picking door actually spoke.
			if sel.Moves != nil {
				dpath := m.Path
				if dpath != "" && fsIsFile(dpath) &&
					!noteLocked(frontmatterOf(dpath)) {
					if StampCurated(dpath, log) {
						report.Stamped++
						if dryrun.IsEnabled() {
							log(fmt.Sprintf("🏷️ %s would be stamped "+
								"curated — the declined move is "+
								"remembered, the next scan skips it "+
								"(dry-run rehearsal)", filepath.Base(dpath)),
								"info")
						} else {
							log(fmt.Sprintf("🏷️ %s stamped curated — "+
								"the declined move is remembered, the "+
								"next scan skips it (remove the line by "+
								"hand to ask again)", filepath.Base(dpath)),
								"info")
						}
					}
				}
			}
			continue // unchecked, never moved
		}
		src := m.Path
		destRel := SanitizeFolder(m.To)
		if src == "" || !fsIsFile(src) || destRel == "" {
			// v0.82.0 — a vanished source is SPOKEN.
			name := filepath.Base(src)
			if name == "" || name == "." {
				name = m.Note
				if name == "" {
					name = "a note"
				}
			}
			log(fmt.Sprintf("⚠️ Scan plan: %s is gone — the move is "+
				"skipped (the file was moved or deleted by hand while "+
				"the plan stood open)", name), "warning")
			report.FailedMoves++
			rememberFailed(&report, m, destRel, "gone")
			continue
		}
		// the lock law's apply-time guard.
		if noteLocked(frontmatterOf(src)) {
			log(fmt.Sprintf("🔒 %s is locked — the move is dropped "+
				"(unlock the note to let a future scan file it)",
				filepath.Base(src)), "info")
			continue
		}
		dstDir := filepath.Join(append([]string{vaultPath},
			strings.Split(destRel, "/")...)...)
		base := filepath.Base(src)
		// v0.82.0 — the landing name: the plan may bend the filename
		// past a same-named resident (the 'as' field).
		asName := helpers.Strip(m.As)
		var dst string
		if asName != "" && asName != base {
			dst = filepath.Join(dstDir, asName)
		} else {
			dst = filepath.Join(dstDir, base)
			asName = ""
		}
		if fsExists(dst) {
			log(fmt.Sprintf("⚠️ A note named %s already lives in %s — "+
				"the move is skipped (nothing is overwritten)",
				filepath.Base(src), destRel), "warning")
			report.FailedMoves++
			rememberFailed(&report, m, destRel, "name taken")
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					log(fmt.Sprintf("⚠️ Could not move %s: %v",
						filepath.Base(src), r), "warning")
					report.FailedMoves++
					rememberFailed(&report, m, destRel, fmt.Sprint(r))
				}
			}()
			if err := dryrun.MkdirAll(dstDir, true); err != nil {
				panic(err)
			}
			// the move itself rides the dry-run stand-in (bytes are
			// never rewritten — shutil.move semantics, rehearsed only
			// under a dry run).
			if err := dryrunMoveSeam(src, dst); err != nil {
				panic(err)
			}
			if dryrun.IsEnabled() {
				// a rehearsal: nothing actually moved, but the plan is
				// sound — count it as the move it would be.
				report.NotesMoved++
				note := m.Note
				if note == "" {
					note = filepath.Base(src)
				}
				report.Moved = append(report.Moved, MovedRow{
					Note: note, From: m.From, To: destRel, As: asName})
				suffix := ""
				if asName != "" {
					suffix = fmt.Sprintf(" (as %s)", asName)
				}
				log(fmt.Sprintf("📦 %s would move → %s%s (dry-run "+
					"rehearsal)", filepath.Base(src), destRel, suffix),
					"info")
				// the census rides the rehearsal's own law: counted,
				// never written.
				report.Stamped++
				log(fmt.Sprintf("🏷️ %s would be stamped curated at its "+
					"new home — the next scan skips it (dry-run "+
					"rehearsal)", filepath.Base(src)), "info")
			} else if fsIsFile(dst) && !fsExists(src) {
				report.NotesMoved++
				note := m.Note
				if note == "" {
					note = filepath.Base(src)
				}
				report.Moved = append(report.Moved, MovedRow{
					Note: note, From: m.From, To: destRel, As: asName})
				suffix := ""
				if asName != "" {
					suffix = fmt.Sprintf(" (as %s)", asName)
				}
				log(fmt.Sprintf("📦 %s moved → %s%s", filepath.Base(src),
					destRel, suffix), "info")
				// v0.78.0 — the census stamp's move door: stamped AT ITS
				// NEW HOME.
				if StampCurated(dst, log) {
					report.Stamped++
					log(fmt.Sprintf("🏷️ %s stamped curated at its new "+
						"home — the next scan skips it", filepath.Base(dst)),
						"info")
				}
				// the folder is the category — re-point the state row.
				if isReposKind(kind) {
					if cacheDB != nil {
						func() {
							defer func() { recover() }() // bookkeeping
							fm := frontmatterOf(dst)
							url := helpers.Strip(fm.Source)
							if url != "" {
								if canonical := repoURLKey(url); canonical != "" {
									repointRepoLedger(cacheDB, canonical,
										dst, destRel)
								}
							}
						}()
					}
				} else if state != nil {
					func() {
						defer func() { recover() }() // bookkeeping never
						// fails a move
						fm := frontmatterOf(dst)
						url := helpers.Strip(fm.Source)
						if url != "" {
							if canonical := links.NormalizeWebsiteURL(url); canonical != "" {
								segs := strings.Split(destRel, "/")
								cat := segs[0]
								sub := strings.Join(segs[1:], "/")
								_ = state.MarkProcessed(canonical, dst,
									cat, sub, "full")
							}
						}
					}()
				}
			} else {
				// v0.82.0 — a move that did not land is SPOKEN, never a
				// bare silent count.
				log(fmt.Sprintf("⚠️ Could not verify the move of %s → "+
					"%s (the file did not land) — it rides the next scan",
					filepath.Base(src), destRel), "warning")
				report.FailedMoves++
				rememberFailed(&report, m, destRel, "did not land")
			}
		}()
	}
	// ---- v0.78.0 — the census: the reviewed notes ---------------------
	for _, rpath := range plan.Reviewed {
		if rpath == "" || !fsIsFile(rpath) {
			continue // gone mid-plan — nothing to stamp
		}
		if noteLocked(frontmatterOf(rpath)) {
			continue // the lock is the stronger word
		}
		if StampCurated(rpath, log) {
			report.Stamped++
			if dryrun.IsEnabled() {
				log(fmt.Sprintf("🏷️ %s would be stamped curated — the "+
					"LLM reviewed it and proposed nothing; the next scan "+
					"skips it (dry-run rehearsal)", filepath.Base(rpath)),
					"info")
			} else {
				log(fmt.Sprintf("🏷️ %s stamped curated — the LLM "+
					"reviewed it and proposed nothing; the next scan "+
					"skips it (remove the line by hand to ask again)",
					filepath.Base(rpath)), "info")
			}
		}
	}
	if report.Stamped > 0 && !dryrun.IsEnabled() {
		log(fmt.Sprintf("🏷️ The census stamped %d note(s) curated — "+
			"the next scan skips them all (the scan remembers its own "+
			"work; only NEW notes ride the radar)", report.Stamped), "info")
	}
	// ---- the deletions: the banishment's own machinery ---------------
	if isReposKind(kind) {
		func() {
			defer func() { recover() }()
			var pickedItems []repobanish.PendingRepoBanishment
			for _, d := range plan.DeletionsRepos {
				canon := helpers.Strip(d.Canonical)
				if canon == "" {
					continue
				}
				if !picked(canon, sel.Deletions) {
					continue // the owner unchecked the row — the mark
					// stays, the note stays, the next scan asks again
				}
				dpath := d.Path
				if dpath != "" && fsIsFile(dpath) &&
					noteLocked(frontmatterOf(dpath)) {
					log(fmt.Sprintf("🔒 %s is locked — its deletion "+
						"waits (locked while the plan stood open; "+
						"unlock it to let a future scan banish)",
						filepath.Base(dpath)), "info")
					continue
				}
				pickedItems = append(pickedItems, d)
			}
			if len(pickedItems) > 0 {
				rep := repobanish.BanishMarkedRepos(cacheDB, noteStateDB,
					vaultPath, log, pickedItems)
				report.Banished += len(rep.URLs)
			}
		}()
		return report
	}
	skip := map[string]bool{}
	for _, d := range plan.Deletions {
		canon := helpers.Strip(d.Canonical)
		if canon == "" {
			continue
		}
		if !picked(canon, sel.Deletions) {
			skip[canon] = true // the owner unchecked the row
			continue
		}
		// the mid-plan lock guard.
		dpath := d.Path
		if dpath != "" && fsIsFile(dpath) &&
			noteLocked(frontmatterOf(dpath)) {
			skip[canon] = true
			log(fmt.Sprintf("🔒 %s is locked — its deletion waits "+
				"(locked while the plan stood open; unlock it to let "+
				"a future scan banish)", filepath.Base(dpath)), "info")
		}
	}
	var skipArg map[string]bool
	if len(skip) > 0 {
		skipArg = skip
	}
	func() {
		defer func() { recover() }()
		consume := websitepipeline.ConsumeDecommissionTable(state,
			vaultPath, log, true, skipArg)
		report.Banished += len(consume.BanishedURLs)
		notes := websitepipeline.BanishMarkedNotes(state, vaultPath, log,
			skipArg)
		report.Banished += len(notes.URLs)
	}()
	return report
}

// frontmatterOf reads a note's frontmatter through the scan's own
// parser (the lock/curated guards' read).
func frontmatterOf(path string) *notereads.Inventory {
	fm := parseNoteFrontmatter(path)
	return &fm
}

// siteName renders a consolidation's site for the log ('?' when the
// plan's own cell is empty).
func siteName(site string) string {
	if site == "" {
		return "?"
	}
	return site
}

// dryrunMoveSeam wraps dryrun.Move — the test law's patch point (the
// reference suite patches _dryrun.move to rehearse the did-not-land
// honesty branch; a nil-src move that silently no-ops must be SPOKEN).
var dryrunMoveSeam = dryrun.Move
