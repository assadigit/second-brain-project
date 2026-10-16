package notereads

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/store"
)

// v0.71.0 THE ONE-OPEN LAW — the scan's shared reads (the port of
// core/note_reads.py). See doc.go for the package laws and the
// documented divergences.

// DefaultTextBytes is the default bounded read (the banishment body
// door's own bound — the largest bound any current consumer asks for;
// smaller asks slice).
const DefaultTextBytes = 200_000

// DefaultRAMBudget is the RAM text layer's budget (32 MB — a 1000-note
// vault of 16 KB notes is ~16 MB; giant notes evict FIFO, never a
// crash). The port counts BYTES (the documented pre-classified fix);
// the reference counted characters.
const DefaultRAMBudget = 32 * 1024 * 1024

// ParseVersion is v0.83.0's THE PARSE VERSION: the unified grammar's
// revision marker. A row remembered by an older grammar lacks the
// newest keys — rows without this version are treated as misses and
// re-parsed ONCE (then remembered with the version).
const ParseVersion = 4

// BanishEmoji is the banish grammar's universal marker (mirrored from
// website_pipeline; the parse is a superset and the doors keep their
// own predicates).
const BanishEmoji = "🗑️"

// BanishWords is the grammar's own tight set (mirrored).
var BanishWords = []string{"delete", "banish", "blacklist", "purge",
	"auto_delete", "auto-delete"}

// BanishKeys are the banishment door's boolean keys (mirrored).
var BanishKeys = []string{"decommission", "banish", "blacklist"}

// bodyTagStops is the body-tag scan's stop set (mirrored): an Obsidian
// inline tag ends at whitespace, another '#', or any of these
// ('#delete.' is the tag 'delete' plus sentence punctuation).
const bodyTagStops = " \t\r\n,.;:!?)]}'\"»«…—>"

var banishWordSet = map[string]bool{}
var banishKeySet = map[string]bool{}

func init() {
	for _, w := range BanishWords {
		banishWordSet[w] = true
	}
	for _, k := range BanishKeys {
		banishKeySet[k] = true
	}
}

// The unified frontmatter's key families (the reference's
// _FM_STRING_KEYS / _REPOS_STRING_KEYS / _LOCK_KEYS / _CURATED_KEYS —
// all string-valued except the boolean folds).
var fmStringKeys = map[string]bool{
	"source": true, "category": true, "subcategory": true,
	"managed_by": true, "kind": true, "fetch_status": true,
	"primary_language": true, "org": true,
}

// reposStringKeys — THE VAULT-KIND DOOR's repos metadata (stars reads
// as a string here; the consumer's int fold is vault_scan's own).
var reposStringKeys = map[string]bool{"stars": true}

var lockKeys = map[string]bool{"locked": true, "category_locked": true}
var curatedKeys = map[string]bool{"curated": true}

// ---------------------------------------------------------------------------
// The pure parsers.
// ---------------------------------------------------------------------------

// KV is one ordered dict entry (Python dicts preserve encounter order;
// the row-JSON writer must reproduce it, so the banish map is ordered).
type KV struct{ Key, Value string }

// Unified is ONE parse of a note's frontmatter block — the union of the
// two doors' key sets (has_fm False when the text carries no
// frontmatter block at all; the string keys then read as ” and the
// lists empty). Never errors.
type Unified struct {
	HasFM           bool
	Source          string
	Category        string
	Subcategory     string
	ManagedBy       string
	Kind            string
	FetchStatus     string
	Tags            []string
	Banish          []KV
	Locked          bool
	CategoryLocked  bool
	Curated         bool
	PrimaryLanguage string
	Org             string
	Stars           string
}

// ParseFrontmatterUnified mirrors parse_frontmatter_unified: the
// string-scan parser (no YAML dependency — the house law), first line
// '---', next '---'; flow tags 'tags: [a, b]' + block tags ('- item');
// the banish boolean keys keep their string values; the lock/curated
// keys fold through truthy. Never errors.
func ParseFrontmatterUnified(text string) Unified {
	out := Unified{Tags: []string{}, Banish: []KV{}}
	lines := helpers.SplitLines(text)
	if len(lines) == 0 || helpers.Strip(lines[0]) != "---" {
		return out
	}
	out.HasFM = true
	inTags := false
	for _, line := range lines[1:] {
		s := helpers.Strip(line)
		if s == "---" {
			break
		}
		if inTags && strings.HasPrefix(s, "- ") {
			out.Tags = append(out.Tags, stripTagItem(s[2:]))
			continue
		}
		inTags = false
		if !strings.Contains(s, ":") {
			continue
		}
		key, val, _ := strings.Cut(s, ":") // partition on the FIRST colon
		key = helpers.PyLower(helpers.Strip(key))
		val = helpers.Strip(val)
		if key == "tags" {
			inner := val
			if strings.HasPrefix(inner, "[") && strings.HasSuffix(inner, "]") {
				inner = inner[1 : len(inner)-1]
			} else if inner == "" {
				inTags = true // block style — items follow
			}
			for _, item := range strings.Split(inner, ",") {
				item = stripTagItem(item)
				if item != "" {
					out.Tags = append(out.Tags, item)
				}
			}
			continue
		}
		clean := stripValueQuotes(val)
		switch {
		case fmStringKeys[key] || reposStringKeys[key]:
			setUnifiedString(&out, key, clean)
		case banishKeySet[key]:
			setBanishKV(&out, key, clean)
		case lockKeys[key] || curatedKeys[key]:
			setUnifiedBool(&out, key, truthy(clean))
		}
	}
	return out
}

// BodyBanishTagOf mirrors body_banish_tag_of: the note BODY's own
// inline tags — a token counts when it is '#' + a non-space run
// preceded by a line start, whitespace, or an opening bracket, and the
// run contains 🗑️ or IS a banish word exactly. The frontmatter block
// and fenced code blocks are skipped — and a text whose first line is
// not '---' has NO scannable body at all (start = len(lines), the
// shared law of both doors: everything before a frontmatter block
// would be nothing). Returns the marker as typed ('#auto-delete') or
// ”.
func BodyBanishTagOf(text string) string {
	lines := helpers.SplitLines(text)
	if len(lines) == 0 {
		return ""
	}
	start := len(lines) // unterminated block: all frontmatter
	if helpers.Strip(lines[0]) == "---" {
		for i := 1; i < len(lines); i++ {
			if helpers.Strip(lines[i]) == "---" {
				start = i + 1
				break
			}
		}
	}
	inFence := false
	for _, line := range lines[start:] {
		stripped := helpers.Strip(line)
		if strings.HasPrefix(stripped, "```") || strings.HasPrefix(stripped, "~~~") {
			inFence = !inFence
			continue
		}
		if inFence {
			continue
		}
		rs := []rune(line)
		for i := 0; i < len(rs); i++ {
			if rs[i] != '#' {
				continue
			}
			if i+1 >= len(rs) {
				continue
			}
			nxt := rs[i+1]
			if nxt == ' ' || nxt == '\t' || nxt == '#' {
				continue // "# word" (a heading) or "##" — never a tag
			}
			if i > 0 {
				prev := rs[i-1]
				if prev != ' ' && prev != '\t' && prev != '(' && prev != '[' {
					continue // mid-word (C#, a URL's #fragment) — not a tag
				}
			}
			j := i + 1
			for j < len(rs) && !strings.ContainsRune(bodyTagStops, rs[j]) && rs[j] != '#' {
				j++
			}
			token := helpers.Strip(string(rs[i+1 : j]))
			if token == "" {
				continue
			}
			low := helpers.PyLower(token)
			if strings.Contains(low, BanishEmoji) || banishWordSet[low] {
				return "#" + token
			}
		}
	}
	return ""
}

// truthy mirrors _truthy — a frontmatter value's truth (the lock doors'
// own tolerance: quote-stripped, lowercased, exactly true/yes/1).
func truthy(val string) bool {
	v := helpers.Strip(helpers.StripChars(helpers.StripChars(helpers.Strip(val), `"`), `'`))
	switch helpers.PyLower(v) {
	case "true", "yes", "1":
		return true
	}
	return false
}

// stripTagItem mirrors item.strip().strip('"').strip("'").strip().
func stripTagItem(s string) string {
	return helpers.Strip(helpers.StripChars(helpers.StripChars(helpers.Strip(s), `"`), `'`))
}

// stripValueQuotes mirrors val.strip('"').strip("'").strip() (val is
// already whitespace-stripped at the partition).
func stripValueQuotes(val string) string {
	return helpers.Strip(helpers.StripChars(helpers.StripChars(val, `"`), `'`))
}

func setUnifiedString(out *Unified, key, val string) {
	switch key {
	case "source":
		out.Source = val
	case "category":
		out.Category = val
	case "subcategory":
		out.Subcategory = val
	case "managed_by":
		out.ManagedBy = val
	case "kind":
		out.Kind = val
	case "fetch_status":
		out.FetchStatus = val
	case "primary_language":
		out.PrimaryLanguage = val
	case "org":
		out.Org = val
	case "stars":
		out.Stars = val
	}
}

func setUnifiedBool(out *Unified, key string, val bool) {
	switch key {
	case "locked":
		out.Locked = val
	case "category_locked":
		out.CategoryLocked = val
	case "curated":
		out.Curated = val
	}
}

// setBanishKV keeps Python dict semantics: a repeated key updates its
// value IN PLACE (the first position wins).
func setBanishKV(out *Unified, key, val string) {
	for i := range out.Banish {
		if out.Banish[i].Key == key {
			out.Banish[i].Value = val
			return
		}
	}
	out.Banish = append(out.Banish, KV{key, val})
}

// ---------------------------------------------------------------------------
// The text-mode read law (the ONLY file-open site).
// ---------------------------------------------------------------------------

// readBounded mirrors _read_bounded(path, max_bytes): one bounded text
// read — open(..., encoding='utf-8', errors='replace').read(n) with
// universal newlines (\r\n and lone \r translate to \n; \x85 does NOT),
// the maximal-subpart replacement decode, and the CHARACTER cut at n.
// Any error returns ”.
//
// Bounded byte read: at most 4n+7 bytes are read, which provably yields
// the same first n runes as decoding the whole file — every multibyte
// sequence that could influence a rune within the first n (≤ 4n bytes)
// starts at ≤ 4n-1 and needs at most 3 more bytes (≤ 4n+2 < 4n+7), so
// no sequence is falsely truncated at the boundary; and the newline
// translation is safe at any cut (\r\n and lone \r both collapse to \n
// — a \r at the boundary translates identically either way).
func readBounded(path string, maxBytes int) string {
	// Python's read(negative) would read everything, but text() clamps
	// to >= 1 before calling — unreachable through the cache; read(0)
	// is ''.
	if maxBytes <= 0 {
		return ""
	}
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	data, err := io.ReadAll(io.LimitReader(f, int64(4*maxBytes+7)))
	if err != nil {
		return ""
	}
	text := helpers.PyNewlines(helpers.DecodeUTF8Replace(data))
	return helpers.CutRunes(text, maxBytes)
}

// Fingerprint is the note's identity for one scan: (mtime_ns, size).
type Fingerprint struct {
	MtimeNS int64
	Size    int64
}

// statFingerprint mirrors _stat_fingerprint: (st_mtime_ns, st_size), or
// no-match when the path cannot be stat'd (gone mid-scan — the doors
// already treat that as the empty shapes).
func statFingerprint(path string) (Fingerprint, bool) {
	info, err := os.Stat(path)
	if err != nil {
		return Fingerprint{}, false
	}
	// st_mtime_ns: Linux gives the timespec's own nanoseconds (exactly
	// ModTime's UnixNano); Windows' NTFS timestamps are 100ns units in
	// both worlds (the Phase-7 clean-machine leg re-verifies).
	return Fingerprint{MtimeNS: info.ModTime().UnixNano(), Size: info.Size()}, true
}

// ---------------------------------------------------------------------------
// The row-JSON layer (the persisted parse's serialization).
// ---------------------------------------------------------------------------

// rowFM is the stored row's JSON shape (json.loads' target on the read
// path; the write path is marshalRowJSON's hand-rolled Python-shape
// serializer).
type rowFM struct {
	HasFM           bool              `json:"has_fm"`
	Source          string            `json:"source"`
	Category        string            `json:"category"`
	Subcategory     string            `json:"subcategory"`
	ManagedBy       string            `json:"managed_by"`
	Kind            string            `json:"kind"`
	FetchStatus     string            `json:"fetch_status"`
	Tags            []string          `json:"tags"`
	Banish          map[string]string `json:"banish"`
	Locked          bool              `json:"locked"`
	CategoryLocked  bool              `json:"category_locked"`
	Curated         bool              `json:"curated"`
	PrimaryLanguage string            `json:"primary_language"`
	Org             string            `json:"org"`
	Stars           string            `json:"stars"`
	V               int               `json:"v"`
}

// rowIsCurrent mirrors _row_is_current composed with the caller's
// non-empty-dict check: a remembered row's parse grammar is the current
// one (rows written before ParseVersion lack the newest keys — a row
// without the marker answers yesterday's question and re-reads once,
// then rides the new marker forever). Undecodable/empty rows read as
// stale, exactly like the reference's json.loads failure → {} → falsy.
func rowIsCurrent(fmJSON []byte) bool {
	var row rowFM
	if err := json.Unmarshal(fmJSON, &row); err != nil {
		return false
	}
	return row.V == ParseVersion
}

// unifiedFromJSON mirrors _unified_from_json — a stored row's JSON →
// the unified shape (missing keys read as their empties; the version
// marker never leaks into the shape) — WITH THE DOCUMENTED
// PRE-CLASSIFIED FIX (spec quirk 1): the reference iterates only
// _FM_STRING_KEYS and so drops `stars` on DB hits; the port restores
// the row's own value (regression: TestDBHitRowsCarryStars).
func unifiedFromJSON(fmJSON []byte) (Unified, bool) {
	var row rowFM
	if err := json.Unmarshal(fmJSON, &row); err != nil {
		return Unified{}, false
	}
	out := Unified{
		HasFM:           row.HasFM,
		Source:          row.Source,
		Category:        row.Category,
		Subcategory:     row.Subcategory,
		ManagedBy:       row.ManagedBy,
		Kind:            row.Kind,
		FetchStatus:     row.FetchStatus,
		Tags:            append([]string{}, row.Tags...),
		Locked:          row.Locked,
		CategoryLocked:  row.CategoryLocked,
		Curated:         row.Curated,
		PrimaryLanguage: row.PrimaryLanguage,
		Org:             row.Org,
		Stars:           row.Stars, // THE FIX (the reference drops it)
	}
	// The banish dict's order is unobservable on the read path (hit
	// rows are never re-stored; dict comparisons are orderless) — the
	// keys are sorted for determinism.
	keys := make([]string, 0, len(row.Banish))
	for k := range row.Banish {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	out.Banish = make([]KV, 0, len(keys))
	for _, k := range keys {
		out.Banish = append(out.Banish, KV{k, row.Banish[k]})
	}
	if out.Tags == nil {
		out.Tags = []string{}
	}
	return out, true
}

// marshalRowJSON mirrors json.dumps(row_fm, ensure_ascii=False) for the
// unified shape: INSERTION ORDER (the unified dict's own — has_fm,
// source, …, stars, then the version marker last, exactly what
// _db_store appends), Python's default separators (", " and ": "),
// Python's minimal string escaping. Pinned byte-for-byte by the
// fixture's corpus (TestRowJSONByteParity).
func marshalRowJSON(fm Unified) string {
	var b strings.Builder
	b.WriteString(`{"has_fm": `)
	writePyJSONBool(&b, fm.HasFM)
	writePyJSONStringField(&b, "source", fm.Source)
	writePyJSONStringField(&b, "category", fm.Category)
	writePyJSONStringField(&b, "subcategory", fm.Subcategory)
	writePyJSONStringField(&b, "managed_by", fm.ManagedBy)
	writePyJSONStringField(&b, "kind", fm.Kind)
	writePyJSONStringField(&b, "fetch_status", fm.FetchStatus)
	// tags: ["a", "b"] (", " separator); [] when empty
	b.WriteString(`, "tags": [`)
	for i, t := range fm.Tags {
		if i > 0 {
			b.WriteString(", ")
		}
		writePyJSONString(&b, t)
	}
	b.WriteString(`]`)
	// banish: {"k": "v", …} in encounter order; {} when empty
	b.WriteString(`, "banish": {`)
	for i, kv := range fm.Banish {
		if i > 0 {
			b.WriteString(", ")
		}
		writePyJSONString(&b, kv.Key)
		b.WriteString(": ")
		writePyJSONString(&b, kv.Value)
	}
	b.WriteString(`}`)
	writePyJSONBoolField(&b, "locked", fm.Locked)
	writePyJSONBoolField(&b, "category_locked", fm.CategoryLocked)
	writePyJSONBoolField(&b, "curated", fm.Curated)
	writePyJSONStringField(&b, "primary_language", fm.PrimaryLanguage)
	writePyJSONStringField(&b, "org", fm.Org)
	writePyJSONStringField(&b, "stars", fm.Stars)
	// v0.73.0 — the parse version rides the remembered row (added last
	// by _db_store, exactly this position).
	b.WriteString(fmt.Sprintf(`, "v": %d}`, ParseVersion))
	return b.String()
}

func writePyJSONStringField(b *strings.Builder, key, val string) {
	b.WriteString(", ")
	writePyJSONString(b, key)
	b.WriteString(": ")
	writePyJSONString(b, val)
}

func writePyJSONBoolField(b *strings.Builder, key string, val bool) {
	b.WriteString(", ")
	writePyJSONString(b, key)
	b.WriteString(": ")
	writePyJSONBool(b, val)
}

func writePyJSONBool(b *strings.Builder, val bool) {
	if val {
		b.WriteString("true")
	} else {
		b.WriteString("false")
	}
}

// writePyJSONString mirrors Python json.dumps(s, ensure_ascii=False):
// '"' → \", '\' → \\, the \b \f \n \r \t shortcuts, \u00xx (lowercase
// hex) for the other control chars; EVERYTHING else raw — no HTML
// escaping, no \u2028/\u2029 escaping, 0x7f raw.
func writePyJSONString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\b':
			b.WriteString(`\b`)
		case '\f':
			b.WriteString(`\f`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			if r < 0x20 {
				fmt.Fprintf(b, `\u%04x`, r)
			} else {
				b.WriteRune(r)
			}
		}
	}
	b.WriteByte('"')
}

// ---------------------------------------------------------------------------
// The shape adapters (each door's own contract, byte-for-byte — a copy
// carrying ONLY the door's keys, so a consumer mutating its shape never
// corrupts the cache's).
// ---------------------------------------------------------------------------

// Inventory is the inventory door's shape (the 13 keys — source,
// category, subcategory, managed_by, kind, fetch_status, tags, locked,
// category_locked, curated, primary_language, org, stars).
type Inventory struct {
	Source          string   `json:"source"`
	Category        string   `json:"category"`
	Subcategory     string   `json:"subcategory"`
	ManagedBy       string   `json:"managed_by"`
	Kind            string   `json:"kind"`
	FetchStatus     string   `json:"fetch_status"`
	Tags            []string `json:"tags"`
	Locked          bool     `json:"locked"`
	CategoryLocked  bool     `json:"category_locked"`
	Curated         bool     `json:"curated"`
	PrimaryLanguage string   `json:"primary_language"`
	Org             string   `json:"org"`
	Stars           string   `json:"stars"`
}

// Banish is the banishment door's shape (source, managed_by, tags,
// banish) — nil when the file carries no frontmatter block (a human's
// note).
type Banish struct {
	Source    string   `json:"source"`
	ManagedBy string   `json:"managed_by"`
	Tags      []string `json:"tags"`
	Banish    []KV     `json:"banish"`
}

func inventoryShape(fm Unified) Inventory {
	return Inventory{
		Source:          fm.Source,
		Category:        fm.Category,
		Subcategory:     fm.Subcategory,
		ManagedBy:       fm.ManagedBy,
		Kind:            fm.Kind,
		FetchStatus:     fm.FetchStatus,
		Tags:            append([]string{}, fm.Tags...),
		Locked:          fm.Locked,
		CategoryLocked:  fm.CategoryLocked,
		Curated:         fm.Curated,
		PrimaryLanguage: fm.PrimaryLanguage,
		Org:             fm.Org,
		Stars:           fm.Stars,
	}
}

func banishShape(fm Unified) *Banish {
	if !fm.HasFM {
		return nil
	}
	banish := make([]KV, len(fm.Banish))
	copy(banish, fm.Banish)
	return &Banish{
		Source:    fm.Source,
		ManagedBy: fm.ManagedBy,
		Tags:      append([]string{}, fm.Tags...),
		Banish:    banish,
	}
}

// ---------------------------------------------------------------------------
// NoteReadCache — the one-open cache: RAM text layer + persisted parse
// layer.
// ---------------------------------------------------------------------------

type textEntry struct {
	fp    Fingerprint
	bound int
	text  string
}

type fmEntry struct {
	fp Fingerprint
	fm Unified
}

type tagEntry struct {
	fp  Fingerprint
	tag string
}

type dbRowResult struct {
	fp     Fingerprint
	fmJSON string
	tag    string
}

// Stats is the closing log line's own numbers (the one-open law's
// proof, on the record).
type Stats struct {
	Opens       int `json:"opens"`        // real file opens this cache performed
	RamHits     int `json:"ram_hits"`     // text slices served from the RAM layer
	DbHits      int `json:"db_hits"`      // parses served from the last scan's rows
	DbWrites    int `json:"db_writes"`    // parses remembered for the next scan
	NotesParsed int `json:"notes_parsed"` // notes with a remembered parse
}

// Cache is the one-open cache. Construct with a dbPath to persist
// parses across scans (the app's own cache.db — a new table, own
// connection, the store plumbing law) or with "" for a RAM-only
// scan-local cache. Thread-safe: one mutex around the layers (the Go
// shape of Python's single RLock; the DB statements additionally ride
// the store's embedded RWMutex).
type Cache struct {
	mu         sync.Mutex
	dbPath     string
	textBudget int

	texts     map[string]textEntry // path → (fingerprint, bound, text)
	textOrder []string             // FIFO eviction order (duplicates tolerated, Python's own)
	textBytes int                  // BYTES (the documented fix; the reference counted chars)

	fm  map[string]fmEntry
	tag map[string]tagEntry

	db       *store.DB
	dbBroken bool

	opens, ramHits, dbHits, dbWrites int
}

// New mirrors NoteReadCache.__init__: dbPath "" is a RAM-only cache;
// the text budget is floored at DefaultTextBytes (every custom budget
// included — Python's max(int(text_budget or 0), DEFAULT_TEXT_BYTES)).
// A broken DB degrades to the RAM layer (never an error out of the
// constructor — the scan must not die on a bad cache.db).
func New(dbPath string, textBudget ...int) *Cache {
	budget := DefaultRAMBudget
	if len(textBudget) > 0 {
		budget = textBudget[0]
	}
	if budget < DefaultTextBytes {
		budget = DefaultTextBytes
	}
	c := &Cache{
		dbPath:     dbPath,
		textBudget: budget,
		texts:      map[string]textEntry{},
		fm:         map[string]fmEntry{},
		tag:        map[string]tagEntry{},
	}
	if dbPath != "" {
		c.openDB()
	}
	return c
}

// noteReadsDDL is the reference's own CREATE statement, byte-identical
// inside (sqlite_master stores the NORMALIZED form — "IF NOT EXISTS "
// dropped — so the executed guard must not change the inner text; the
// schema golden pins the stored form).
const noteReadsDDL = `
                    CREATE TABLE IF NOT EXISTS note_reads (
                        path TEXT PRIMARY KEY,
                        mtime_ns INTEGER NOT NULL,
                        size INTEGER NOT NULL,
                        frontmatter_json TEXT NOT NULL DEFAULT '',
                        banish_tag TEXT NOT NULL DEFAULT '',
                        scanned_at REAL NOT NULL DEFAULT 0
                    )`

func (c *Cache) openDB() {
	db, err := store.Open(c.dbPath)
	if err != nil {
		c.dbBroken = true
		return
	}
	if _, err := db.SQL().Exec(noteReadsDDL); err != nil {
		db.Close()
		c.dbBroken = true
		return
	}
	c.db = db
}

// DBBroken reports the broken-DB degradation flag (the reference's
// _db_broken — an unopenable path or a failing statement turns the
// cache RAM-only, never raising).
func (c *Cache) DBBroken() bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.dbBroken
}

// dbRow mirrors _db_row: (fingerprint, frontmatter_json, banish_tag)
// for a path, or no-match. A real statement error marks the DB broken
// (a failing DB never breaks the scan); a missing row is a plain miss.
func (c *Cache) dbRow(path string) (dbRowResult, bool) {
	if c.db == nil {
		return dbRowResult{}, false
	}
	var mtimeNS, size int64
	var fmJSON, tag string
	c.db.RLock()
	err := c.db.SQL().QueryRow(
		"SELECT mtime_ns, size, frontmatter_json, banish_tag "+
			"FROM note_reads WHERE path = ?", path).
		Scan(&mtimeNS, &size, &fmJSON, &tag)
	c.db.RUnlock()
	if err == sql.ErrNoRows {
		return dbRowResult{}, false
	}
	if err != nil {
		c.dbBroken = true
		return dbRowResult{}, false
	}
	return dbRowResult{fp: Fingerprint{mtimeNS, size}, fmJSON: fmJSON, tag: tag}, true
}

// dbStore mirrors _db_store: remember one parse (the upsert), the
// version marker riding the row, scanned_at as the REAL time.time()
// float. A failing statement marks the DB broken — never breaks the
// scan.
func (c *Cache) dbStore(path string, fp Fingerprint, fm Unified, tag string) {
	if c.db == nil || c.dbBroken {
		return
	}
	rowJSON := marshalRowJSON(fm)
	now := float64(time.Now().UnixNano()) / 1e9 // the time.time() float shape
	c.db.Lock()
	_, err := c.db.SQL().Exec(
		"INSERT INTO note_reads "+
			"(path, mtime_ns, size, frontmatter_json, banish_tag, "+
			"scanned_at) VALUES (?, ?, ?, ?, ?, ?) "+
			"ON CONFLICT(path) DO UPDATE SET mtime_ns=excluded."+
			"mtime_ns, size=excluded.size, frontmatter_json=excluded."+
			"frontmatter_json, banish_tag=excluded.banish_tag, "+
			"scanned_at=excluded.scanned_at",
		path, fp.MtimeNS, fp.Size, rowJSON, tag, now)
	c.db.Unlock()
	if err != nil {
		c.dbBroken = true // a failing DB never breaks the scan
		return
	}
	c.dbWrites++
}

// Text mirrors text(path, max_bytes): the note's bounded text — ONE
// open per note per scan, sliced to the caller's own bound. A stat
// fingerprint guards every hit (a note changed mid-scan is re-read,
// honestly).
func (c *Cache) Text(path string, maxBytes int) string {
	if maxBytes < 1 {
		maxBytes = 1
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.textLocked(path, maxBytes)
}

func (c *Cache) textLocked(path string, maxBytes int) string {
	fp, ok := statFingerprint(path)
	if !ok {
		delete(c.texts, path)
		return ""
	}
	if hit, ok := c.texts[path]; ok && hit.fp == fp && hit.bound >= maxBytes {
		c.ramHits++
		return helpers.CutRunes(hit.text, maxBytes)
	}
	body := readBounded(path, maxBytes)
	c.opens++
	c.rememberText(path, fp, maxBytes, body)
	return body
}

// rememberText mirrors _remember_text: store the largest read we have
// (a bigger ask replaces a smaller cached text); evict FIFO over the
// budget — BYTES here (the documented fix), where the reference
// counted characters.
func (c *Cache) rememberText(path string, fp Fingerprint, maxBytes int, body string) {
	prev, had := c.texts[path]
	if had && prev.fp == fp && prev.bound >= maxBytes {
		return
	}
	if had {
		c.textBytes -= len(prev.text)
		if i := indexOf(c.textOrder, path); i >= 0 {
			c.textOrder = append(c.textOrder[:i], c.textOrder[i+1:]...)
		}
	}
	c.texts[path] = textEntry{fp: fp, bound: maxBytes, text: body}
	c.textOrder = append(c.textOrder, path)
	c.textBytes += len(body)
	for c.textBytes > c.textBudget && len(c.textOrder) > 0 {
		victim := c.textOrder[0]
		c.textOrder = c.textOrder[1:]
		if v, ok := c.texts[victim]; ok {
			delete(c.texts, victim)
			c.textBytes -= len(v.text)
		}
	}
}

func indexOf(list []string, s string) int {
	for i, v := range list {
		if v == s {
			return i
		}
	}
	return -1
}

// Frontmatter mirrors frontmatter(path): the inventory door's shape,
// served from RAM, then the last scan's row (fingerprint match AND the
// current parse grammar — the version marker), else ONE bounded read +
// ONE unified parse, remembered (and the body tag computed from the
// same read, then both stored as one row).
func (c *Cache) Frontmatter(path string) Inventory {
	c.mu.Lock()
	defer c.mu.Unlock()
	fp, ok := statFingerprint(path)
	if !ok {
		delete(c.fm, path)
		return inventoryShape(ParseFrontmatterUnified(""))
	}
	if hit, ok := c.fm[path]; ok && hit.fp == fp {
		return inventoryShape(hit.fm)
	}
	if !c.dbBroken {
		if row, ok := c.dbRow(path); ok && row.fp == fp && rowIsCurrent([]byte(row.fmJSON)) {
			c.dbHits++
			fm, _ := unifiedFromJSON([]byte(row.fmJSON))
			c.fm[path] = fmEntry{fp: fp, fm: fm}
			return inventoryShape(fm)
		}
	}
	fm := ParseFrontmatterUnified(c.textLocked(path, DefaultTextBytes))
	c.fm[path] = fmEntry{fp: fp, fm: fm}
	if !c.dbBroken {
		tag := c.tagOf(path, fp)
		c.dbStore(path, fp, fm, tag)
	}
	return inventoryShape(fm)
}

// BanishFrontmatter mirrors banish_frontmatter(path): the banishment
// door's shape — nil when the file carries no frontmatter block (its
// own contract for "a human's note"); the version marker gates the
// remembered row exactly the inventory door's law.
func (c *Cache) BanishFrontmatter(path string) *Banish {
	c.mu.Lock()
	defer c.mu.Unlock()
	fp, ok := statFingerprint(path)
	if !ok {
		delete(c.fm, path)
		return nil
	}
	if hit, ok := c.fm[path]; ok && hit.fp == fp {
		return banishShape(hit.fm)
	}
	if !c.dbBroken {
		if row, ok := c.dbRow(path); ok && row.fp == fp && rowIsCurrent([]byte(row.fmJSON)) {
			c.dbHits++
			fm, _ := unifiedFromJSON([]byte(row.fmJSON))
			c.fm[path] = fmEntry{fp: fp, fm: fm}
			return banishShape(fm)
		}
	}
	fm := ParseFrontmatterUnified(c.textLocked(path, DefaultTextBytes))
	c.fm[path] = fmEntry{fp: fp, fm: fm}
	if !c.dbBroken {
		tag := c.tagOf(path, fp)
		c.dbStore(path, fp, fm, tag)
	}
	return banishShape(fm)
}

// BanishTag mirrors banish_tag(path): the body door's inline tag
// ('#auto-delete' or ”) — parsed once, remembered across scans with
// the fingerprint. The DB path is NOT version-gated (the reference's
// own law — the body grammar is stable; preserve it), unlike the
// frontmatter doors.
func (c *Cache) BanishTag(path string) string {
	c.mu.Lock()
	defer c.mu.Unlock()
	fp, ok := statFingerprint(path)
	if !ok {
		delete(c.tag, path)
		return ""
	}
	if hit, ok := c.tag[path]; ok && hit.fp == fp {
		return hit.tag
	}
	if !c.dbBroken {
		if row, ok := c.dbRow(path); ok && row.fp == fp {
			c.dbHits++
			c.tag[path] = tagEntry{fp: fp, tag: row.tag}
			return row.tag
		}
	}
	if hit, ok := c.fm[path]; ok && hit.fp == fp {
		_ = hit // the RAM parse cache is already warm (Python's own flow)
	} else {
		seeded := false
		if !c.dbBroken {
			// only a CURRENT-grammar row may seed the RAM parse cache
			// (a stale row's tags would ride a parse the lock keys
			// never saw)
			if row, ok := c.dbRow(path); ok && row.fp == fp && rowIsCurrent([]byte(row.fmJSON)) {
				fm, _ := unifiedFromJSON([]byte(row.fmJSON))
				c.fm[path] = fmEntry{fp: fp, fm: fm}
				seeded = true
			}
		}
		if !seeded {
			c.fm[path] = fmEntry{fp: fp,
				fm: ParseFrontmatterUnified(c.textLocked(path, DefaultTextBytes))}
		}
	}
	return c.tagOf(path, fp)
}

// tagOf mirrors _tag_of: the body tag behind the frontmatter block —
// the text layer's own tail (the frontmatter always lives in the head
// of the same bounded read), parsed once and remembered. A
// budget-evicted text re-reads (the honest open — the tag must be
// computed).
func (c *Cache) tagOf(path string, fp Fingerprint) string {
	var text string
	if hit, ok := c.texts[path]; ok && hit.fp == fp {
		text = hit.text
	} else {
		text = c.textLocked(path, DefaultTextBytes)
	}
	tag := BodyBanishTagOf(text)
	c.tag[path] = tagEntry{fp: fp, tag: tag}
	return tag
}

// Stats mirrors stats(): the closing log line's own numbers.
func (c *Cache) Stats() Stats {
	c.mu.Lock()
	defer c.mu.Unlock()
	return Stats{Opens: c.opens, RamHits: c.ramHits, DbHits: c.dbHits,
		DbWrites: c.dbWrites, NotesParsed: len(c.fm)}
}

// Close mirrors close(prune=True): prune dead-path rows (the memory
// never outlives the vault it describes), close the connection, return
// the stats for the closing honesty line. Never errors (the reference's
// bare excepts); idempotent.
func (c *Cache) Close(prune bool) Stats {
	stats := c.Stats()
	if c.db != nil && prune {
		c.pruneDead()
	}
	if c.db != nil {
		c.db.Close()
		c.db = nil
	}
	return stats
}

// pruneDead mirrors the prune half: select every remembered path, drop
// the rows whose files no longer stat. Rows are collected before the
// deletes (the store law — collect before nested statements); the
// deletes ride one transaction (Python's executemany + commit).
func (c *Cache) pruneDead() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.db.RLock()
	rows, err := c.db.SQL().Query("SELECT path FROM note_reads")
	if err != nil {
		c.db.RUnlock()
		return
	}
	var paths []string
	for rows.Next() {
		var p string
		if err := rows.Scan(&p); err != nil {
			rows.Close()
			c.db.RUnlock()
			return
		}
		paths = append(paths, p)
	}
	rows.Close()
	c.db.RUnlock()

	var dead []string
	for _, p := range paths {
		if _, ok := statFingerprint(p); !ok {
			dead = append(dead, p)
		}
	}
	if len(dead) == 0 {
		return
	}
	c.db.Lock()
	tx, err := c.db.SQL().Begin()
	if err != nil {
		c.db.Unlock()
		return
	}
	for _, p := range dead {
		if _, err := tx.Exec("DELETE FROM note_reads WHERE path = ?", p); err != nil {
			tx.Rollback()
			c.db.Unlock()
			return
		}
	}
	if err := tx.Commit(); err != nil {
		c.db.Unlock()
		return
	}
	c.db.Unlock()
}
