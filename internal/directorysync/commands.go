// commands.go — PHASE 2, THE COMMAND LEG (directory_sync.py lines
// 1297-1822, SPEC §10.2): pull the Worker's admin command queue, apply
// every command to the Obsidian vault (the master) idempotently and
// failure-isolated, mirror the vault to GitHub, publish the new index,
// ack each command — the ack only ever lands AFTER the world already
// agrees (the golden-rule-5 order: a apply, c mirror, d publish,
// e ack). Nothing is deleted permanently — delete moves the note to
// <vault>/.trash/ (SPEC §17.6).
package directorysync

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// commandsPath mirrors _COMMANDS_PATH.
const commandsPath = "/api/commands"

// trashDirname mirrors _TRASH_DIRNAME.
const trashDirname = ".trash"

// mirrorTimeout mirrors push_mirror_repo's per-command 120 s.
const mirrorTimeout = 120

// HTTPUserAgent mirrors HTTP_USER_AGENT (imported from
// cloudflare_sync in the reference; the inline fallback twin rides
// here — the cloudflare_sync port consolidates it later). THE 1010
// LAW: urllib's default signature is banned at Cloudflare's edge.
const HTTPUserAgent = "GitCurator/0.91.0 (+https://github.com/assadigit/GitCurator)"

// dictDoor is the present-aware dict reader the command leg needs
// (Python's `'key' in payload` presence checks — the belt floor's
// dictReader cannot tell absent from nil).
type dictDoor struct {
	get func(key string) (any, bool)
}

func dictDoorOf(v any) dictDoor {
	switch t := v.(type) {
	case *storage.OrderedMap:
		return dictDoor{get: t.Get}
	case map[string]any:
		return dictDoor{func(key string) (any, bool) {
			val, ok := t[key]
			return val, ok
		}}
	}
	return dictDoor{get: func(string) (any, bool) { return nil, false }}
}

func (d dictDoor) Has(key string) bool {
	_, ok := d.get(key)
	return ok
}

func (d dictDoor) Get(key string) any {
	v, _ := d.get(key)
	return v
}

// dictShape answers the dict-shaped door (or the empty reader when
// the value is not dict-shaped — Python's isinstance gate).
func dictShape(v any) (dictDoor, bool) {
	switch t := v.(type) {
	case *storage.OrderedMap:
		return dictDoor{get: t.Get}, true
	case map[string]any:
		return dictDoor{func(key string) (any, bool) {
			val, ok := t[key]
			return val, ok
		}}, true
	}
	return dictDoor{get: func(string) (any, bool) { return nil, false }}, false
}

// CommandResult mirrors one per-command result dict {command_id,
// status, error} for the ack leg.
type CommandResult struct {
	ID     string
	Status string
	Error  string
}

// ---------------------------------------------------------------------------
// The Worker legs (fetch / ack)
// ---------------------------------------------------------------------------

// FetchPendingCommands mirrors fetch_pending_commands: GET
// /api/commands?status=pending (Bearer). Returns the command list, or
// nil on any failure (the caller treats it as "no commands this run"
// — a down Worker must never look like an empty queue that was
// successfully drained).
func FetchPendingCommands(workerURL, token string, timeout int, log func(string)) any {
	if log == nil {
		log = func(string) {}
	}
	if workerURL == "" || token == "" {
		return nil
	}
	if timeout <= 0 {
		timeout = 30
	}
	url := strings.TrimRight(workerURL, "/") + commandsPath + "?status=pending"
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		log("[directory] command pull failed: " + err.Error())
		return nil
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", HTTPUserAgent)
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log("[directory] command pull failed: " + pyURLErrorText(err))
		return nil
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		log("[directory] command pull failed: " + err.Error())
		return nil
	}
	if resp.StatusCode >= 400 {
		// the HTTPError door
		log("[directory] command pull failed: HTTP " + strconv.Itoa(resp.StatusCode) +
			" " + helpers.CutRunes(helpers.DecodeUTF8Replace(raw), 200))
		return nil
	}
	data, err := storage.DecodeJSONValue(raw)
	if err != nil {
		log("[directory] command pull failed: " + err.Error())
		return nil
	}
	commands, ok := dictDoorOf(data).Get("commands").([]any)
	if !ok {
		return nil // `commands if isinstance(commands, list) else None`
	}
	return commands
}

// AckCommand mirrors ack_command: POST /api/commands/<id>/ack —
// applied or failed (SPEC §10.2). Never raises; a lost ack simply
// means the command is re-served next run and re-applied
// idempotently.
func AckCommand(workerURL, token, commandID, status, errorText string,
	timeout int, log func(string)) bool {
	if log == nil {
		log = func(string) {}
	}
	if workerURL == "" || token == "" {
		return false
	}
	if timeout <= 0 {
		timeout = 30
	}
	url := strings.TrimRight(workerURL, "/") + commandsPath + "/" +
		pyQuoteAll(commandID) + "/ack"
	body := storage.NewOrderedMap()
	body.Set("status", status)
	if errorText == "" {
		body.Set("error", nil) // json.dumps({'error': error or None})
	} else {
		body.Set("error", errorText)
	}
	req, err := http.NewRequest(http.MethodPost, url,
		strings.NewReader(storage.PyDumpsCompact(body)))
	if err != nil {
		log("[directory] ack failed for " + commandID + ": " + err.Error())
		return false
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", HTTPUserAgent)
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log("[directory] ack failed for " + commandID + ": " + pyURLErrorText(err))
		return false
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		log("[directory] ack failed for " + commandID + ": HTTP " +
			strconv.Itoa(resp.StatusCode) + " " +
			helpers.CutRunes(helpers.DecodeUTF8Replace(raw), 200))
		return false
	}
	if err != nil {
		log("[directory] ack failed for " + commandID + ": " + err.Error())
		return false
	}
	if _, err := storage.DecodeJSONValue(raw); err != nil {
		// the reference parses the ack body (discarding it) — a non-JSON
		// 200 is a ValueError → False
		log("[directory] ack failed for " + commandID + ": " + err.Error())
		return false
	}
	return true
}

// pyQuoteAll mirrors urllib.parse.quote(str(s), safe=”) — every byte
// outside the unreserved set escapes as %XX.
func pyQuoteAll(s string) string {
	const hex = "0123456789ABCDEF"
	var b strings.Builder
	for i := 0; i < len(s); i++ {
		c := s[i]
		if (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z') ||
			(c >= '0' && c <= '9') || c == '_' || c == '-' || c == '.' || c == '~' {
			b.WriteByte(c)
		} else {
			b.WriteByte('%')
			b.WriteByte(hex[c>>4])
			b.WriteByte(hex[c&0xf])
		}
	}
	return b.String()
}

// pyURLErrorText renders a Go transport error the way str(e) renders
// Python's URLError (the cause's text, bare).
func pyURLErrorText(err error) string {
	return err.Error()
}

// ---------------------------------------------------------------------------
// The vault appliers (pure filesystem surgery — every one idempotent)
// ---------------------------------------------------------------------------

// frontmatterSource mirrors _frontmatter_source: the note's source:
// frontmatter value (light scan, no full parse).
func frontmatterSource(text string) string {
	if !strings.HasPrefix(text, "---") {
		return ""
	}
	lines := strings.Split(text, "\n")
	for i := 1; i < len(lines); {
		if helpers.Strip(lines[i]) == "---" {
			for _, line := range lines[1:i] {
				if m := fmLineRe.FindStringSubmatch(line); m != nil && m[1] == "source" {
					return helpers.Strip(unquote(m[2]))
				}
			}
			return ""
		}
		i++
	}
	return ""
}

// walkSiteNotes mirrors _walk_site_notes: site_id → absolute note
// path (the SAME id law as build_index: sha256(normalized
// URL)[:16]). First note wins on duplicates — the index builder's own
// rule, so the command leg can never disagree.
func walkSiteNotes(vault string) map[string]string {
	out := map[string]string{}
	iterNotes(vault, func(rel, absPath string) {
		source := frontmatterSource(readText(absPath))
		if source == "" {
			return
		}
		normalized := links.NormalizeWebsiteURL(source)
		if normalized == "" || !strings.HasPrefix(normalized, "http") {
			return
		}
		sum := sha256Sum16(normalized)
		if _, seen := out[sum]; !seen {
			out[sum] = absPath
		}
	})
	return out
}

// sha256Sum16 mirrors hashlib.sha256(url.encode('utf-8')).hexdigest()[:16].
func sha256Sum16(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])[:16]
}

// splitNote mirrors _split_note: (frontmatter_lines, body_lines) or
// not-note-shaped.
func splitNote(text string) ([]string, []string, bool) {
	if !strings.HasPrefix(text, "---") {
		return nil, nil, false
	}
	lines := strings.Split(text, "\n")
	for i := 1; i < len(lines); i++ {
		if helpers.Strip(lines[i]) == "---" {
			return lines[1:i], lines[i+1:], true
		}
	}
	return nil, nil, false
}

// joinNote mirrors _join_note.
func joinNote(fmLines, bodyLines []string) string {
	return "---\n" + strings.Join(fmLines, "\n") + "\n---\n" +
		strings.Join(bodyLines, "\n")
}

// frontmatterSet mirrors _frontmatter_set: insert or replace one
// scalar frontmatter key, preserving order.
func frontmatterSet(fmLines []string, key, value string) []string {
	pattern := regexp.MustCompile(`(?i)^` + regexp.QuoteMeta(key) + `\s*:`)
	for i, line := range fmLines {
		if pattern.MatchString(line) {
			fmLines[i] = key + ": " + value
			return fmLines
		}
	}
	return append(fmLines, key+": "+value)
}

// pyValueError renders the reference's ValueError failures.
type pyValueError string

func (e pyValueError) Error() string { return "ValueError: " + string(e) }

// PyStr mirrors str(e) — the bare message (the reference's except-path
// log line renders str(e); the result's error renders the prefixed
// f'{type(e).__name__}: {e}').
func (e pyValueError) PyStr() string { return string(e) }

// pySilentFail is the reference's direct-append failure path (the
// unknown command type): no log line, no type prefix — the else
// branch, never the except path.
type pySilentFail string

func (e pySilentFail) Error() string { return string(e) }

// pyErrText renders an error the way str(e) renders the Python
// exception.
func pyErrText(err error) string {
	if pe, ok := err.(interface{ PyStr() string }); ok {
		return pe.PyStr()
	}
	return err.Error()
}

// pyOSError renders an OS failure the reference's way (OSError:
// <python-shaped text>).
func pyOSError(err error, path string) error {
	return fmt.Errorf("OSError: %s", helpers.PyOSErrorText(err, path))
}

// readNoteFile mirrors open(abs_path, 'r', encoding='utf-8').read().
func readNoteFile(absPath string) (string, error) {
	data, err := os.ReadFile(absPath)
	if err != nil {
		return "", pyOSError(err, absPath)
	}
	// Python text mode: universal newlines apply; errors are strict —
	// invalid UTF-8 raises (the golden fixtures are valid UTF-8).
	return helpers.PyNewlines(helpers.DecodeUTF8Replace(data)), nil
}

// writeNoteFile mirrors open(abs_path, 'w', encoding='utf-8').
func writeNoteFile(absPath, text string) error {
	if err := os.WriteFile(absPath, []byte(text), 0o644); err != nil {
		return pyOSError(err, absPath)
	}
	return nil
}

// editNoteFields mirrors _edit_note_fields: apply an edit/review
// command's whitelisted fields to one note — frontmatter
// (source/tags/featured/needs_review/category/subcategory) + the
// body's H1 (name) and TL;DR (summary), or the note's WHOLE body when
// the payload carries body (SPEC §9 P1). Idempotent by construction.
func editNoteFields(absPath string, payload any, log func(string)) error {
	if log == nil {
		log = func(string) {}
	}
	text, err := readNoteFile(absPath)
	if err != nil {
		return err
	}
	fm, body, ok := splitNote(text)
	if !ok {
		return pyValueError("note has no frontmatter")
	}
	p := dictDoorOf(payload)

	if p.Has("url") {
		fm = frontmatterSet(fm, "source", helpers.PyStr(p.Get("url")))
	}
	if tags, ok := p.Get("tags").([]any); ok {
		var kept []string
		for _, t := range tags {
			if pyTruthyDS(t) {
				kept = append(kept, helpers.PyStr(t))
			}
		}
		fm = frontmatterSet(fm, "tags", "["+strings.Join(kept, ", ")+"]")
	}
	if p.Has("featured") {
		fm = frontmatterSet(fm, "featured", pyBoolLiteral(pyTruthyDS(p.Get("featured"))))
	}
	// v0.39.0 — Phase 4 (SPEC §9 P1): the needs-review flag rides the
	// frontmatter and flows back out through build_index on the next
	// publish.
	if p.Has("needs_review") {
		fm = frontmatterSet(fm, "needs_review", pyBoolLiteral(pyTruthyDS(p.Get("needs_review"))))
	}
	if p.Has("category") {
		fm = frontmatterSet(fm, "category", storage.PyDumps(helpers.PyStr(p.Get("category"))))
	}
	if p.Has("subcategory") {
		fm = frontmatterSet(fm, "subcategory", storage.PyDumps(helpers.PyStr(p.Get("subcategory"))))
	}

	// v0.39.0 — Phase 4: the whole-body edit. The admin's edit dialog
	// can send the note's complete markdown body (the frontmatter
	// stays — only what follows the second '---' is replaced).
	if bodyStr, ok := p.Get("body").(string); ok && helpers.Strip(bodyStr) != "" {
		newBody := strings.TrimSuffix(strings.ReplaceAll(bodyStr, "\r\n", "\n"), "\n")
		newBody = strings.TrimPrefix(newBody, "\n")
		body = strings.Split(newBody, "\n")
		log("[directory] replaced note body " + filepath.Base(absPath))
	}

	if p.Has("name") || p.Has("summary") {
		name := pyStrOrEmpty(p.Get("name"))
		summary := pyStrOrEmpty(p.Get("summary"))
		doneName := name == ""
		doneSummary := summary == ""
		for i, line := range body {
			if !doneName {
				if h1Re.MatchString(line) {
					body[i] = "# " + name
					doneName = true
					continue
				}
			}
			if !doneSummary {
				if tldrRe.MatchString(line) {
					body[i] = "> **TL;DR:** " + summary
					doneSummary = true
				}
			}
			if doneName && doneSummary {
				break
			}
		}
		if !doneName {
			body = insertString(body, 0, "# "+name)
		}
		if !doneSummary && summary != "" {
			// the reference's verbatim `1 if not done_name else 1` —
			// always index 1 (the quirk preserved)
			body = insertString(body, 1, "> **TL;DR:** "+summary)
		}
	}

	if err := writeNoteFile(absPath, joinNote(fm, body)); err != nil {
		return err
	}
	log("[directory] edited note " + filepath.Base(absPath))
	return nil
}

func insertString(s []string, idx int, val string) []string {
	out := make([]string, 0, len(s)+1)
	out = append(out, s[:idx]...)
	out = append(out, val)
	out = append(out, s[idx:]...)
	return out
}

func pyBoolLiteral(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// moveNoteFile mirrors _move_note_file: move a note into
// safe_filename(category[/subcategory]) — the exact inverse of the
// folder-name law. A name collision gets a numeric suffix (SPEC
// §10.2). Returns the new absolute path; moving onto the same spot is
// a no-op.
func moveNoteFile(absPath, vault, category, subcategory string,
	log func(string)) (string, error) {
	if log == nil {
		log = func(string) {}
	}
	parts := []string{storage.SafeFilename(category)}
	if subcategory != "" {
		parts = append(parts, storage.SafeFilename(subcategory))
	}
	targetDir := filepath.Join(append([]string{vault}, parts...)...)
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return "", pyOSError(err, targetDir)
	}
	filename := filepath.Base(absPath)
	target := filepath.Join(targetDir, filename)
	absTarget, err1 := filepath.Abs(target)
	absSource, err2 := filepath.Abs(absPath)
	if err1 == nil && err2 == nil && absTarget == absSource {
		return absPath, nil
	}
	stem, ext := storage.PySplitExt(filename)
	n := 2
	for fileExists(target) {
		target = filepath.Join(targetDir, stem+" "+strconv.Itoa(n)+ext)
		n++
	}
	if err := pyMove(absPath, target); err != nil {
		return "", pyOSError(err, absPath)
	}
	log("[directory] moved note → " + pyRelPathOf(target, vault))
	return target, nil
}

// trashNote mirrors _trash_note (SPEC §17.6): never a permanent
// delete — the note moves to <vault>/.trash/ with a timestamp suffix
// (that folder is never indexed and never mirrored).
func trashNote(absPath, vault string, log func(string)) (string, error) {
	if log == nil {
		log = func(string) {}
	}
	trashDir := filepath.Join(vault, trashDirname)
	if err := os.MkdirAll(trashDir, 0o755); err != nil {
		return "", pyOSError(err, trashDir)
	}
	filename := filepath.Base(absPath)
	stem, ext := storage.PySplitExt(filename)
	stamp := localStamp(clockNow())
	target := filepath.Join(trashDir, stem+"-"+stamp+ext)
	n := 2
	for fileExists(target) {
		target = filepath.Join(trashDir, stem+"-"+stamp+"-"+strconv.Itoa(n)+ext)
		n++
	}
	if err := pyMove(absPath, target); err != nil {
		return "", pyOSError(err, absPath)
	}
	log("[directory] trashed note → " + pyRelPathOf(target, vault))
	return target, nil
}

// pyMove is shutil.move: rename, with the copy+delete fallback when
// the rename crosses a filesystem boundary.
func pyMove(src, dst string) error {
	if err := os.Rename(src, dst); err == nil {
		return nil
	}
	data, err := os.ReadFile(src)
	if err != nil {
		return err
	}
	if err := os.WriteFile(dst, data, 0o644); err != nil {
		return err
	}
	if err := os.Remove(src); err != nil {
		return err
	}
	return nil
}

// banishAdd mirrors _banish_add: append to
// Websites/_system/banished.json (SPEC §7.4) — an idempotent upsert
// keyed on host+url. The existing entries ride VERBATIM (unknown keys
// preserved — the owner's hand edits are sacred).
func banishAdd(vault, host, url, reason string, log func(string)) {
	if log == nil {
		log = func(string) {}
	}
	path := filepath.Join(vault, SystemDirname, BanishedFilename)
	var entries []any
	if data, err := os.ReadFile(path); err == nil {
		if v, err := storage.DecodeJSONValue(data); err == nil {
			if list, ok := v.([]any); ok {
				entries = list
			}
		}
	}
	entry := storage.NewOrderedMap()
	entryHost := helpers.PyLower(pyStrOrEmpty(host))
	entry.Set("host", entryHost)
	entryURL := ""
	if url != "" {
		entryURL = url
		entry.Set("url", entryURL)
	}
	entryReason := reason
	if entryReason == "" {
		entryReason = "admin banish"
	}
	entry.Set("reason", entryReason)
	entry.Set("banished_at", utcISOSeconds(clockNow()))

	var kept []any
	for _, e := range entries {
		if ed, isDict := dictShape(e); isDict {
			// a dict entry with a matching host+url drops (the upsert);
			// non-dict junk rides verbatim (Python's isinstance gate)
			eHost := helpers.PyLower(pyStrOrEmpty(ed.Get("host")))
			eURL := pyStrOrEmpty(ed.Get("url"))
			if eHost == entryHost && eURL == entryURL {
				continue
			}
		}
		kept = append(kept, e)
	}
	kept = append(kept, entry)
	_ = os.MkdirAll(filepath.Dir(path), 0o755)
	_ = os.WriteFile(path, []byte(storage.PyDumps(kept)+"\n"), 0o644)
	log("[directory] banished host added: " + entryHost)
}

// banishRemove mirrors _banish_remove: remove every entry for a host
// from banished.json (the admin's Unban, SPEC §9) — idempotent: an
// absent host is a no-op. The rewrite happens only when something was
// actually removed.
func banishRemove(vault, host string, log func(string)) {
	if log == nil {
		log = func(string) {}
	}
	path := filepath.Join(vault, SystemDirname, BanishedFilename)
	if _, err := os.Stat(path); err != nil {
		return
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	v, err := storage.DecodeJSONValue(data)
	if err != nil {
		return
	}
	entries, ok := v.([]any)
	if !ok {
		return
	}
	wanted := helpers.PyLower(pyStrOrEmpty(host))
	var kept []any
	for _, e := range entries {
		if ed, isDict := dictShape(e); isDict {
			if helpers.PyLower(pyStrOrEmpty(ed.Get("host"))) == wanted {
				continue
			}
		}
		kept = append(kept, e)
	}
	if len(kept) == len(entries) {
		return
	}
	_ = os.WriteFile(path, []byte(storage.PyDumps(kept)+"\n"), 0o644)
	log("[directory] banished host removed: " + wanted)
}

// noteCategoryOf mirrors _note_category_of: the note's CURRENT folder
// category (verbatim folder name).
func noteCategoryOf(absPath, vault string) string {
	return relSegment(absPath, vault, 0)
}

// noteSubcategoryOf mirrors _note_subcategory_of.
func noteSubcategoryOf(absPath, vault string) string {
	return relSegment(absPath, vault, 1)
}

func relSegment(absPath, vault string, idx int) string {
	rel, err := filepath.Rel(vault, absPath)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.ReplaceAll(filepath.ToSlash(rel), "\\", "/"), "/")
	// category needs len(parts) > 1; subcategory len(parts) > 2
	if len(parts) > idx+1 {
		return parts[idx]
	}
	return ""
}

// pyRelPathOf mirrors os.path.relpath(target, vault) with os.sep →
// '/' (the log lines' shape).
func pyRelPathOf(target, vault string) string {
	rel, err := filepath.Rel(vault, target)
	if err != nil {
		return target
	}
	return filepath.ToSlash(rel)
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// ---------------------------------------------------------------------------
// apply_admin_commands — the vault surgery, idempotent + isolated
// ---------------------------------------------------------------------------

// ApplyAdminCommands mirrors apply_admin_commands: apply every
// command, IN ORDER, to the vault. Returns the per-command results
// for the ack leg. Idempotent (a gone note → applied),
// failure-isolated (one bad command never stops the others).
func ApplyAdminCommands(vault string, commands any, log func(string)) []CommandResult {
	if log == nil {
		log = func(string) {}
	}
	var results []CommandResult
	notes := walkSiteNotes(vault)
	items, _ := commands.([]any)
	for _, item := range items {
		c := dictDoorOf(item)
		commandID := pyStrOrEmpty(c.Get("command_id"))
		ctype := pyStrOrEmpty(c.Get("type"))
		p := dictDoorOf(c.Get("payload"))
		siteID := pyStrOrEmpty(c.Get("site_id"))

		res := CommandResult{ID: commandID, Status: "applied"}
		fail := func(err error) {
			if _, silent := err.(pySilentFail); !silent {
				log("[directory] command " + ctype + " " + commandID + " failed: " +
					pyErrText(err))
			}
			results = append(results, CommandResult{
				ID: commandID, Status: "failed", Error: err.Error(),
			})
		}
		applyOne := func() error {
			if ctype == "unban" {
				banishRemove(vault, pyStrOrEmpty(p.Get("host")), log)
				return nil
			}

			absPath := ""
			if siteID != "" {
				absPath = notes[siteID]
			}
			switch ctype {
			case "delete", "banish":
				if absPath != "" && fileExists(absPath) {
					if _, err := trashNote(absPath, vault, log); err != nil {
						return err
					}
				}
				// else: already gone — applied, never failed (SPEC §10.2)
				if ctype == "banish" {
					banishAdd(vault, pyStrOrEmpty(c.Get("host")),
						pyStrOrEmpty(p.Get("url")),
						pyStrOrEmpty(p.Get("reason")), log)
				}
				return nil
			case "feature", "unfeature":
				if absPath != "" && fileExists(absPath) {
					payload := storage.NewOrderedMap()
					payload.Set("featured", ctype == "feature")
					if err := editNoteFields(absPath, payload, log); err != nil {
						return err
					}
				}
				return nil
			case "review":
				// v0.39.0 — Phase 4 (SPEC §9 P1): the manual
				// "Needs review (maybe not free anymore)" flag.
				if absPath != "" && fileExists(absPath) {
					payload := storage.NewOrderedMap()
					payload.Set("needs_review", pyTruthyDS(p.Get("needs_review")))
					if err := editNoteFields(absPath, payload, log); err != nil {
						return err
					}
				}
				return nil
			case "move":
				if absPath != "" && fileExists(absPath) {
					category := pyStrOrEmpty(p.Get("category"))
					subcategory := pyStrOrEmpty(p.Get("subcategory"))
					if category == "" {
						return pyValueError("move command without a category")
					}
					newPath, err := moveNoteFile(absPath, vault, category, subcategory, log)
					if err != nil {
						return err
					}
					payload := storage.NewOrderedMap()
					payload.Set("category", category)
					payload.Set("subcategory", subcategory)
					return editNoteFields(newPath, payload, log)
				}
				return nil
			case "edit":
				if absPath != "" && fileExists(absPath) {
					if err := editNoteFields(absPath, c.Get("payload"), log); err != nil {
						return err
					}
					if p.Has("category") || p.Has("subcategory") {
						category := pyStrOrEmpty(p.Get("category"))
						if category == "" {
							category = noteCategoryOf(absPath, vault)
						}
						subcategory := pyStrOrEmpty(p.Get("subcategory"))
						if subcategory == "" {
							subcategory = noteSubcategoryOf(absPath, vault)
						}
						if category != "" {
							newPath, err := moveNoteFile(absPath, vault, category, subcategory, log)
							if err != nil {
								return err
							}
							absNew, e1 := filepath.Abs(newPath)
							absOld, e2 := filepath.Abs(absPath)
							if e1 == nil && e2 == nil && absNew != absOld {
								payload := storage.NewOrderedMap()
								payload.Set("category", category)
								payload.Set("subcategory", subcategory)
								if err := editNoteFields(newPath, payload, log); err != nil {
									return err
								}
							}
						}
					}
				}
				return nil
			default:
				return pySilentFail("unknown command type: " + helpers.PyRepr(ctype))
			}
		}
		if err := applyOne(); err != nil {
			fail(err)
			continue
		}
		results = append(results, res)
	}
	return results
}

// ---------------------------------------------------------------------------
// push_mirror_repo + run_command_cycle
// ---------------------------------------------------------------------------

// mirrorGit mirrors push_mirror_repo's inner run(): one git command
// in the vault, captured, 120 s timeout; (nil, failDetail) on
// failure, (stdout, "") on success. NOTE the honest convergence:
// the reference's runner inherits the parent env and never redacts
// (its module has no _git_env — the vaultseal/goodrepos law family
// does not extend here; reality wins). HideWindow rides ADR-0009's
// Windows quirk list.
func mirrorGit(vault string, args ...string) (*string, string) {
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(mirrorTimeout)*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = vault
	hideWindow(cmd)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	if ctx.Err() == context.DeadlineExceeded {
		return nil, "timed out" // a hung child is killed; the run failed
	}
	if err != nil {
		// (proc.stderr or proc.stdout or '').strip()[:200]
		detail := strings.TrimSpace(stderr.String())
		if detail == "" {
			detail = strings.TrimSpace(stdout.String())
		}
		return nil, helpers.CutRunes(detail, 200)
	}
	out := strings.TrimSpace(stdout.String())
	return &out, ""
}

// mirrorRun is the logging wrapper over mirrorGit (the reference's
// run() logs the failure line with the captured output).
func mirrorRun(vault string, log func(string), args ...string) *string {
	res, failDetail := mirrorGit(vault, args...)
	if res == nil {
		log("[directory] git " + strings.Join(args, " ") + " failed: " + failDetail)
	}
	return res
}

// PushMirrorRepo mirrors push_mirror_repo (golden rule 5c): when the
// vault sits inside a git repo with a remote, every change is
// committed and pushed (GitHub history IS the audit trail, SPEC
// §10.2). Not a repo / no remote / git missing → a logged no-op,
// never a raise.
func PushMirrorRepo(vault string, log func(string)) bool {
	if log == nil {
		log = func(string) {}
	}
	if _, err := exec.LookPath("git"); err != nil {
		log("[directory] mirror push skipped: git not found")
		return false
	}
	if mirrorRun(vault, log, "rev-parse", "--is-inside-work-tree") == nil {
		log("[directory] mirror push skipped: vault is not a git repo")
		return false
	}
	remotes := mirrorRun(vault, log, "remote")
	if remotes == nil {
		return false
	}
	if *remotes == "" {
		log("[directory] mirror push skipped: no git remote configured")
		return false
	}
	status := mirrorRun(vault, log, "status", "--porcelain")
	if status == nil {
		return false
	}
	if *status == "" {
		log("[directory] mirror push: nothing to push")
		return true
	}
	if mirrorRun(vault, log, "add", "-A") == nil {
		return false
	}
	stamp := utcISOSeconds(clockNow())
	if mirrorRun(vault, log, "commit", "-m", "admin commands applied: "+stamp) == nil {
		return false
	}
	if mirrorRun(vault, log, "push") == nil {
		return false
	}
	log("[directory] mirror repo pushed to GitHub")
	return true
}

// RunCommandCycle mirrors run_command_cycle: the FULL Phase-2 leg in
// the golden-rule-5 order — (a) pull + apply pending admin commands
// to the vault, (c) mirror the vault to GitHub, (d) build + publish
// the index, (e) ack every command. Never raises. Returns the summary
// {pulled, applied, failed, acked, published} (insertion-ordered for
// the CLI's dict repr).
func RunCommandCycle(config any, log func(string)) *storage.OrderedMap {
	if log == nil {
		log = func(string) {}
	}
	summary := storage.NewOrderedMap()
	summary.Set("pulled", 0)
	summary.Set("applied", 0)
	summary.Set("failed", 0)
	summary.Set("acked", 0)
	summary.Set("published", false)

	cfg := effectiveConfig(config)
	vault, _ := cfg.Get("website_vault_path")
	vaultStr := pyStrOrEmpty(vault)
	workerURL := pyStrOrEmpty(mustCfgGet(cfg, "directory_worker_url"))
	token := pyStrOrEmpty(mustCfgGet(cfg, "directory_sync_token"))
	if vaultStr == "" || !isDir(vaultStr) {
		log("[directory] command cycle skipped: website_vault_path not configured")
		return summary
	}

	// (a) pull + apply
	commands := FetchPendingCommands(workerURL, token, 30, log)
	var results []CommandResult
	if commands == nil {
		if workerURL != "" && token != "" {
			log("[directory] command pull unavailable — commands wait, " +
				"nothing is acked (a down Worker must never look like an empty queue)")
		}
	} else if items, ok := commands.([]any); ok && len(items) > 0 {
		log("[directory] pulled " + strconv.Itoa(len(items)) + " pending command(s)")
		results = ApplyAdminCommands(vaultStr, items, log)
		summary.Set("pulled", len(items))
		applied, failed := 0, 0
		for _, r := range results {
			if r.Status == "applied" {
				applied++
			} else if r.Status == "failed" {
				failed++
			}
		}
		summary.Set("applied", applied)
		summary.Set("failed", failed)
	}

	// (c) mirror — only when commands were applied (the reference's
	// `if results:` door; push_mirror_repo itself no-ops on a clean
	// tree).
	if len(results) > 0 {
		PushMirrorRepo(vaultStr, log)
	}

	// (d) publish the world as it now stands
	EnsureSystemFiles(vaultStr, log)
	index := BuildIndex(vaultStr, nil, log)
	published := PublishIndex(index, workerURL, token, 30, log)
	ok := published != nil && pyTruthyDS(dictDoorOf(published).Get("success"))
	summary.Set("published", ok)

	// (e) ack — only after the publish, so the Worker's overlay and
	// the stored index already agree (no resurrection window).
	if len(results) > 0 {
		acked := 0
		for _, r := range results {
			if AckCommand(workerURL, token, r.ID, r.Status, r.Error, 30, log) {
				acked++
			}
		}
		summary.Set("acked", acked)
	}
	return summary
}

func mustCfgGet(cfg *storage.OrderedMap, key string) any {
	v, _ := cfg.Get(key)
	return v
}

func isDir(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && fi.IsDir()
}
