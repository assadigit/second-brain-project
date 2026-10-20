// linking.go — linking.py (Phase 6, steps 3–4): candidate search, the
// LLM confirmation, the link store and the user-facing surfaces (the
// Suggestions note + the "Related (auto)" blocks written into MIRROR
// copies only). Pure stdlib; no GUI.
package mirror

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/prompts"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// The delimited "Related (auto)" block (mirror copies only).
const (
	RelatedStart = "<!-- gitcurator:related:start -->"
	RelatedEnd   = "<!-- gitcurator:related:end -->"
)

// SPEC §4.8: "Cap of about 5 to 7 links per note."
const (
	LinksCap    = 7
	LinksTarget = 5
)

// Candidate search defaults.
const (
	DefaultTopK     = 8
	DefaultMinScore = 0.30
)

// SuggestionsFilename is the Suggestions note's name under Library/.
const SuggestionsFilename = "Suggestions.md"

// Suggestion statuses.
const (
	StatusPending  = "pending"
	StatusApproved = "approved"
	StatusRejected = "rejected"
)

// ---------------------------------------------------------------------------
// 1) Collect the recall fields from both machine vaults
// ---------------------------------------------------------------------------

// VaultPath is one collect_note_fields input (the reference's dict is
// {'github': path, 'websites': path} — insertion order drives the
// first-wins dedupe, so the Go shape is an ordered slice).
type VaultPath struct {
	Vault string
	Path  string
}

// FieldsEntry is collect_note_fields' per-source value:
// {vault, path, stem, fields}.
type FieldsEntry struct {
	Vault  string
	Path   string
	Stem   string
	Fields RecallFields
}

// CollectedFields is the whole collect output (the reference's dict
// plus the __duplicate__ count entry, broken out honestly).
type CollectedFields struct {
	BySource   map[string]FieldsEntry
	Order      []string // insertion (walk) order — the dict's key order
	Duplicates int      // the __duplicate__ {'count': n} entry
}

// CollectNoteFields walks both machine vaults, reads every managed
// note, and extracts the recall fields keyed by the note's source URL.
// Managed notes only (a source: line — the identity everything else
// already keys on); the linking layer never touches unmanaged files.
// Duplicate sources keep the FIRST hit (walk-order dependent — the
// reference's own documented law) and count a warning entry.
func CollectNoteFields(vaults []VaultPath) *CollectedFields {
	out := &CollectedFields{BySource: map[string]FieldsEntry{}}
	for _, vp := range vaults {
		if vp.Path == "" {
			continue
		}
		if fi, err := os.Stat(vp.Path); err != nil || !fi.IsDir() {
			continue
		}
		for _, note := range notestate.ScanVault(vp.Path, nil) {
			if note.Unmanaged {
				continue
			}
			if note.SourceURL == "" {
				continue
			}
			if _, dup := out.BySource[note.SourceURL]; dup {
				out.Duplicates++
				continue
			}
			text, ok := readTextFile(note.Path)
			if !ok {
				continue
			}
			// v0.68.0 — THE RESOLVABLE LINK LAW: the stem (filename
			// without .md) rides along — the Related blocks and the
			// Suggestions note link by stem now (with the title as
			// alias), so a click resolves to the real note.
			stem, _ := storage.PySplitExt(filepath.Base(note.Path))
			out.BySource[note.SourceURL] = FieldsEntry{
				Vault:  vp.Vault,
				Path:   note.Path,
				Stem:   stem,
				Fields: RecallFieldsFor(text, vp.Vault),
			}
			out.Order = append(out.Order, note.SourceURL)
		}
	}
	return out
}

// ---------------------------------------------------------------------------
// 2) Candidate pairs (cosine neighbors on the stored vectors)
// ---------------------------------------------------------------------------

// VecEntry preserves Python dict insertion order — the pair
// canonicalization's tie-stability depends on it (the house law: Go
// map ranges are RANDOM).
type VecEntry struct {
	Key string
	Vec []float64
}

// CandidatePair is one (key_a, key_b, score) result.
type CandidatePair struct {
	A     string
	B     string
	Score float64
}

// FindCandidatePairs computes neighbors across ALL notes (both vaults —
// "across domains and within a domain"). Returns each unordered pair
// once (a<b lexicographically), sorted by score descending. topK bounds
// per-note candidates; minScore prunes weak matches.
func FindCandidatePairs(entries []VecEntry, topK int, minScore float64) []CandidatePair {
	var keys []string
	vecs := map[string][]float64{}
	for _, e := range entries {
		if len(e.Vec) == 0 {
			continue
		}
		if _, seen := vecs[e.Key]; !seen {
			keys = append(keys, e.Key)
		}
		vecs[e.Key] = e.Vec
	}
	type pairKey struct{ A, B string }
	var order []pairKey // best's insertion order — the stable-sort tie law
	best := map[pairKey]float64{}
	for _, keyA := range keys {
		vecA := vecs[keyA]
		type scored struct {
			Key   string
			Score float64
		}
		var scoredList []scored
		for _, keyB := range keys {
			if keyB == keyA {
				continue
			}
			score := Cosine(vecA, vecs[keyB])
			if score >= minScore {
				scoredList = append(scoredList, scored{keyB, score})
			}
		}
		sort.SliceStable(scoredList, func(x, y int) bool {
			return scoredList[x].Score > scoredList[y].Score
		})
		if topK > 0 && len(scoredList) > topK {
			scoredList = scoredList[:topK]
		}
		for _, s := range scoredList {
			// tuple(sorted((key_a, key_b))) — the canonical pair
			pair := pairKey{keyA, s.Key}
			if s.Key < keyA {
				pair = pairKey{s.Key, keyA}
			}
			if score, seen := best[pair]; !seen || s.Score > score {
				if !seen {
					order = append(order, pair)
				}
				best[pair] = s.Score
			}
		}
	}
	out := make([]CandidatePair, 0, len(order))
	for _, p := range order {
		out = append(out, CandidatePair{A: p.A, B: p.B, Score: best[p]})
	}
	sort.SliceStable(out, func(x, y int) bool {
		return out[x].Score > out[y].Score
	})
	return out
}

// ---------------------------------------------------------------------------
// 3) The LLM confirmation
// ---------------------------------------------------------------------------

// BuildConfirmPrompt fills l01 for one pair (the loader refuses an
// unfilled slot).
func BuildConfirmPrompt(fieldsA, fieldsB RecallFields) (string, error) {
	slot := func(v, fallback string) string {
		return pyOr(helpers.CutRunes(helpers.Strip(v), 300), fallback)
	}
	return prompts.LoadPrompt("l01_confirm",
		prompts.Slot{Name: "A_NAME", Value: slot(fieldsA.Name, "(unnamed)")},
		prompts.Slot{Name: "A_LINE", Value: slot(fieldsA.OneLine, "(no description)")},
		prompts.Slot{Name: "A_RECALL", Value: slot(fieldsA.Recall, "(not captured)")},
		prompts.Slot{Name: "A_TAGS", Value: slot(fieldsA.Tags, "(none)")},
		prompts.Slot{Name: "B_NAME", Value: slot(fieldsB.Name, "(unnamed)")},
		prompts.Slot{Name: "B_LINE", Value: slot(fieldsB.OneLine, "(no description)")},
		prompts.Slot{Name: "B_RECALL", Value: slot(fieldsB.Recall, "(not captured)")},
		prompts.Slot{Name: "B_TAGS", Value: slot(fieldsB.Tags, "(none)")},
	)
}

// ParseConfirmReply: LLM reply → (related, reason). A nil result means
// unparseable → treated as NOT related (a wrong link costs more than a
// missing one).
func ParseConfirmReply(reply string) (*bool, string) {
	payload, err := llmclient.ExtractJSON(reply)
	if err != nil {
		return nil, ""
	}
	var relatedV any
	switch t := payload.(type) {
	case *storage.OrderedMap:
		relatedV, _ = t.Get("related")
	case map[string]any:
		relatedV = t["related"]
	default:
		return nil, ""
	}
	var reason string
	switch t := payload.(type) {
	case *storage.OrderedMap:
		if v, ok := t.Get("reason"); ok && v != nil {
			reason = helpers.CutRunes(helpers.Strip(helpers.PyStr(v)), 200)
		}
	case map[string]any:
		if v, ok := t["reason"]; ok && v != nil {
			reason = helpers.CutRunes(helpers.Strip(helpers.PyStr(v)), 200)
		}
	}
	switch v := relatedV.(type) {
	case bool:
		return &v, reason
	case string:
		lowered := strings.ToLower(helpers.Strip(v))
		switch lowered {
		case "true", "yes":
			t := true
			return &t, reason
		case "false", "no":
			f := false
			return &f, reason
		}
	}
	return nil, ""
}

// ConfirmPair sends one pair through the prompt → reply → parse
// pipeline.
func ConfirmPair(fieldsA, fieldsB RecallFields, llm RecallLLM) (*bool, string) {
	prompt, err := BuildConfirmPrompt(fieldsA, fieldsB)
	if err != nil {
		return nil, ""
	}
	reply, err := llm([]llmclient.Message{{Role: "user", Content: prompt}})
	if err != nil {
		return nil, ""
	}
	return ParseConfirmReply(reply)
}

// ---------------------------------------------------------------------------
// 4) The link store
// ---------------------------------------------------------------------------

// LinkStore persists suggestions in cache.db (own table, own
// connection — the NoteStateDB pattern). A pair (stored sorted, a<b) is
// suggested at most once: rejected never reappears, pending is not
// duplicated, approved is kept for the Related blocks.
type LinkStore struct {
	*store.DB
}

// NewLinkStore opens (creating if absent) the store; the literal
// default "cache.db" is anchored to APP_DIR (the v0.09.4 rule).
func NewLinkStore(dbPath string) (*LinkStore, error) {
	if dbPath == "cache.db" {
		dbPath = filepath.Join(constants.AppDir(), "cache.db")
	}
	base, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	s := &LinkStore{DB: base}
	s.Lock()
	defer s.Unlock()
	if _, err := s.SQL().Exec(`
                CREATE TABLE IF NOT EXISTS link_suggestions (
                        key_a TEXT NOT NULL,
                        key_b TEXT NOT NULL,
                        score REAL NOT NULL,
                        status TEXT NOT NULL DEFAULT 'pending',
                        reason TEXT NOT NULL DEFAULT '',
                        created_at TEXT NOT NULL,
                        decided_at TEXT,
                        PRIMARY KEY (key_a, key_b)
                )`); err != nil {
		base.Close()
		return nil, err
	}
	return s, nil
}

// linkPairKey is the sorted (a, b) pair.
type linkPairKey struct{ A, B string }

// pairOf mirrors LinkStore._pair: (a, b) if a <= b else (b, a).
func pairOf(a, b string) linkPairKey {
	if a <= b {
		return linkPairKey{a, b}
	}
	return linkPairKey{b, a}
}

// KnownPairs is {(key_a, key_b): status} — everything ever suggested.
func (s *LinkStore) KnownPairs() (map[linkPairKey]string, error) {
	s.RLock()
	defer s.RUnlock()
	rows, err := s.SQL().Query(
		"SELECT key_a, key_b, status FROM link_suggestions")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[linkPairKey]string{}
	for rows.Next() {
		var a, b, status string
		if err := rows.Scan(&a, &b, &status); err != nil {
			continue
		}
		out[linkPairKey{a, b}] = status
	}
	return out, rows.Err()
}

// KnownPairsSet is the composition driver's never-resuggest filter:
// every pair ever suggested (any status) as {(a, b) sorted} -- the
// external caller cannot build the unexported linkPairKey, so the set
// crosses as [2]string tuples (a <= b, pairOf's order).
func (s *LinkStore) KnownPairsSet() (map[[2]string]bool, error) {
	known, err := s.KnownPairs()
	if err != nil {
		return nil, err
	}
	out := make(map[[2]string]bool, len(known))
	for k := range known {
		out[[2]string{k.A, k.B}] = true
	}
	return out, nil
}

// Suggest inserts a NEW pending suggestion. Returns 'pending' on
// insert or the EXISTING status when the pair is already known (a
// rejected pair is never suggested again — that is the point).
func (s *LinkStore) Suggest(keyA, keyB string, score float64, reason string) (string, error) {
	pair := pairOf(keyA, keyB)
	s.Lock()
	defer s.Unlock()
	var status string
	row := s.SQL().QueryRow(
		"SELECT status FROM link_suggestions WHERE key_a=? AND key_b=?",
		pair.A, pair.B)
	if err := row.Scan(&status); err == nil {
		return status, nil
	}
	if _, err := s.SQL().Exec(
		"INSERT INTO link_suggestions (key_a, key_b, score, status, reason,"+
			" created_at) VALUES (?,?,?,?,?,?)",
		pair.A, pair.B, score, StatusPending, reason, pyStrftimeSeconds()); err != nil {
		return "", err
	}
	return StatusPending, nil
}

// Decide moves pending → approved / rejected (decided_at stamped).
// An unknown pair or an already-decided one is a no-op returning false.
// An invalid status is an error (the reference's ValueError).
func (s *LinkStore) Decide(keyA, keyB, status string) (bool, error) {
	if status != StatusApproved && status != StatusRejected {
		return false, fmt.Errorf("invalid status: %s", status)
	}
	pair := pairOf(keyA, keyB)
	s.Lock()
	defer s.Unlock()
	res, err := s.SQL().Exec(
		"UPDATE link_suggestions SET status=?, decided_at=?"+
			" WHERE key_a=? AND key_b=? AND status=?",
		status, pyStrftimeSeconds(), pair.A, pair.B, StatusPending)
	if err != nil {
		return false, err
	}
	n, _ := res.RowsAffected()
	return n > 0, nil
}

// SuggestionRow is one with_status row.
type SuggestionRow struct {
	KeyA   string
	KeyB   string
	Score  float64
	Reason string
}

// WithStatus lists the suggestions with one status, score DESC.
func (s *LinkStore) WithStatus(status string) ([]SuggestionRow, error) {
	s.RLock()
	defer s.RUnlock()
	rows, err := s.SQL().Query(
		"SELECT key_a, key_b, score, reason FROM link_suggestions"+
			" WHERE status=? ORDER BY score DESC", status)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []SuggestionRow
	for rows.Next() {
		var r SuggestionRow
		if err := rows.Scan(&r.KeyA, &r.KeyB, &r.Score, &r.Reason); err != nil {
			continue
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ApprovedLink is one approved edge as seen from a note:
// {other, reason, score}.
type ApprovedLink struct {
	Other  string
	Reason string
	Score  float64
}

// ApprovedLinks is {key: [{other, reason, score}…]} — the
// Related-block input, capped per note (strongest first, then
// alphabetical).
func (s *LinkStore) ApprovedLinks() (map[string][]ApprovedLink, error) {
	rows, err := s.WithStatus(StatusApproved)
	if err != nil {
		return nil, err
	}
	links := map[string][]ApprovedLink{}
	for _, row := range rows {
		links[row.KeyA] = append(links[row.KeyA], ApprovedLink{
			Other: row.KeyB, Reason: row.Reason, Score: row.Score})
		links[row.KeyB] = append(links[row.KeyB], ApprovedLink{
			Other: row.KeyA, Reason: row.Reason, Score: row.Score})
	}
	for me, items := range links {
		sort.SliceStable(items, func(x, y int) bool {
			if items[x].Score != items[y].Score {
				return items[x].Score > items[y].Score
			}
			return items[x].Other < items[y].Other
		})
		if len(items) > LinksCap {
			links[me] = items[:LinksCap]
		}
	}
	return links, nil
}

// Counts is {pending, approved, rejected}.
func (s *LinkStore) Counts() (map[string]int, error) {
	s.RLock()
	defer s.RUnlock()
	rows, err := s.SQL().Query(
		"SELECT status, COUNT(*) FROM link_suggestions GROUP BY status")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]int{
		StatusPending: 0, StatusApproved: 0, StatusRejected: 0,
	}
	for rows.Next() {
		var status string
		var n int
		if err := rows.Scan(&status, &n); err != nil {
			continue
		}
		out[status] = n
	}
	return out, rows.Err()
}

// ---------------------------------------------------------------------------
// 5a) The Suggestions note (written under Library/, read back for decisions)
// ---------------------------------------------------------------------------

// noteTitleFor mirrors _note_title_for: the name or '(unnamed)'.
func noteTitleFor(fields *RecallFields) string {
	if fields == nil {
		return "(unnamed)"
	}
	name := helpers.Strip(fields.Name)
	if name == "" {
		return "(unnamed)"
	}
	return name
}

// linkStemSafeRE mirrors _LINK_STEM_SAFE_RE: a stem may ride a
// wiki-link only when it cannot break one (no '|' alias separator, no
// brackets, no heading anchors) — the v0.68.0 RESOLVABLE LINK LAW's
// defensive gate.
var linkStemSafeRE = regexp.MustCompile(`^[^|\[\]#\r\n]+$`)

// LinkTargetFor mirrors _link_target_for: what goes inside [[…]] for
// one note — its FILENAME STEM (which a click actually resolves), with
// the readable title as the Obsidian alias. Unknown/unsafe stems fall
// back to the plain title — the old behavior, never worse.
func LinkTargetFor(entry *FieldsEntry) string {
	if entry == nil {
		return "(unnamed)"
	}
	stem := helpers.Strip(entry.Stem)
	title := noteTitleFor(&entry.Fields)
	if stem != "" && linkStemSafeRE.MatchString(stem) {
		if title != "" && title != stem && title != "(unnamed)" {
			return stem + "|" + title
		}
		return stem
	}
	return title
}

// SuggestionsPath is the Suggestions note's path under the manual
// vault's Library/.
func SuggestionsPath(manualVault string) string {
	return filepath.Join(manualVault, "Library", SuggestionsFilename)
}

// BuildSuggestionsNote renders the Suggestions note: pending
// suggestions as Obsidian checkboxes (tick = approve, strike through =
// reject), then the approved and rejected history. Written ONLY under
// <manual>/Library/ (the caller owns that law).
func BuildSuggestionsNote(s *LinkStore, fieldsByKey map[string]*FieldsEntry) (string, error) {
	now := clockNow().Format("2006-01-02 15:04")
	entryFor := func(key string) *FieldsEntry {
		return fieldsByKey[key]
	}
	lines := []string{
		"---",
		"managed_by: gitcurator",
		"purpose: link suggestions",
		"---",
		"",
		"# Link suggestions",
		"",
		fmt.Sprintf("Generated %s. Related notes are linked automatically once", now),
		"you approve them here.",
		"",
		"**How to decide:** tick a checkbox (`- [x]`) to APPROVE the",
		"pair; strike the whole line through (`~~…~~`) to REJECT it",
		"forever; leave it as `- [ ]` to keep it pending for later.",
		"Then run the collect step (Settings → More menu, or",
		"`python gitcurator/tools/build_links.py --collect`) — the",
		"Related (auto) blocks in the Library mirror update.",
		"",
	}
	pending, err := s.WithStatus(StatusPending)
	if err != nil {
		return "", err
	}
	lines = append(lines, fmt.Sprintf("## Pending (%d)", len(pending)), "")
	if len(pending) == 0 {
		lines = append(lines, "_(nothing waiting for your decision)_")
	}
	for _, row := range pending {
		a := LinkTargetFor(entryFor(row.KeyA))
		b := LinkTargetFor(entryFor(row.KeyB))
		reason := pyOr(row.Reason, "similar purpose")
		lines = append(lines, fmt.Sprintf(
			"- [ ] [[%s]] ↔ [[%s]] — %s (score %s)",
			a, b, reason, pyScoreRepr(row.Score)))
		lines = append(lines, fmt.Sprintf("  - `%s` / `%s`", row.KeyA, row.KeyB))
	}
	lines = append(lines, "")

	approved, err := s.WithStatus(StatusApproved)
	if err != nil {
		return "", err
	}
	lines = append(lines, fmt.Sprintf("## Approved (%d)", len(approved)), "")
	if len(approved) == 0 {
		lines = append(lines, "_(none yet)_")
	}
	for _, row := range approved {
		a := LinkTargetFor(entryFor(row.KeyA))
		b := LinkTargetFor(entryFor(row.KeyB))
		lines = append(lines, fmt.Sprintf("- [[%s]] ↔ [[%s]] — %s", a, b, row.Reason))
	}
	lines = append(lines, "")

	rejected, err := s.WithStatus(StatusRejected)
	if err != nil {
		return "", err
	}
	lines = append(lines, fmt.Sprintf("## Rejected (%d) — never suggested again",
		len(rejected)), "")
	if len(rejected) == 0 {
		lines = append(lines, "_(none)_")
	}
	for _, row := range rejected {
		a := LinkTargetFor(entryFor(row.KeyA))
		b := LinkTargetFor(entryFor(row.KeyB))
		lines = append(lines, fmt.Sprintf("- ~~[[%s]] ↔ [[%s]]~~ — %s", a, b, row.Reason))
	}
	lines = append(lines, "")
	return strings.Join(lines, "\n"), nil
}

// pyScoreRepr is f"{score:.2f}".
func pyScoreRepr(score float64) string {
	return fmt.Sprintf("%.2f", score)
}

// suggestCheckLineRE mirrors _SUGGEST_CHECK_RE (Python \s → the
// Unicode class; re.match = start-anchored, the trailing $ pinned by
// \z — the note's lines carry no trailing newline).
var suggestCheckLineRE = regexp.MustCompile(
	`\A` + pyS + `*` + `-` + pyS + `*\[([ xX])\]` + pyS + `*(.+?)` + pyS + `*\z`)

var backtickRE = regexp.MustCompile("`([^`]+)`")

// Decision is one read-back decision {key_a, key_b, status}.
type Decision struct {
	KeyA   string
	KeyB   string
	Status string
}

// CollectDecisions reads the Suggestions note back → decisions. A
// TICKED checkbox (- [x]) approves the pair on that line; a
// STRUCK-THROUGH line (~~) rejects it; anything else stays pending.
// Only lines that carry both source URLs (the indented `url` / `url`
// line pairs the builder writes) are read — free text is never
// misparsed as a decision.
func CollectDecisions(notePath string) []Decision {
	var decisions []Decision
	if notePath == "" {
		return decisions
	}
	if fi, err := os.Stat(notePath); err != nil || fi.IsDir() {
		return decisions
	}
	data, err := os.ReadFile(notePath)
	if err != nil {
		return decisions
	}
	lines := pySplitLines(helpers.PyNewlines(helpers.DecodeUTF8Replace(data)))
	for i, line := range lines {
		m := suggestCheckLineRE.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		mark, body := m[1], m[2]
		// the URLs ride the next non-empty indented line
		var urls []string
		for j := i + 1; j < len(lines) && j <= i+2; j++ {
			found := backtickRE.FindAllStringSubmatch(lines[j], -1)
			if len(found) >= 2 {
				urls = []string{found[0][1], found[1][1]}
				break
			}
		}
		if len(urls) != 2 {
			continue
		}
		struck := strings.Contains(body, "~~")
		status := ""
		switch {
		case strings.ToLower(mark) == "x" && !struck:
			status = StatusApproved
		case struck:
			status = StatusRejected
		default:
			continue // still pending — not a decision
		}
		decisions = append(decisions, Decision{
			KeyA: urls[0], KeyB: urls[1], Status: status})
	}
	return decisions
}

// pySplitLines is Python str.splitlines over the full boundary set.
func pySplitLines(s string) []string {
	var out []string
	start := 0
	for i, r := range s {
		isBoundary := false
		switch r {
		case '\n', '\r', '\v', '\f', 0x1c, 0x1d, 0x1e, 0x85, 0x2028, 0x2029:
			isBoundary = true
		}
		if isBoundary {
			if r == '\r' && i+1 < len(s) && s[i+1] == '\n' {
				out = append(out, s[start:i])
				start = i + 2
				continue
			}
			out = append(out, s[start:i])
			start = i + utf8.RuneLen(r)
		}
	}
	out = append(out, s[start:])
	return out
}

// ApplyDecisions applies collected decisions to the store. Returns
// {approved, rejected, ignored} (ignored = already decided or unknown
// pair).
func ApplyDecisions(s *LinkStore, decisions []Decision) (approved, rejected, ignored int, err error) {
	for _, dec := range decisions {
		changed, err := s.Decide(dec.KeyA, dec.KeyB, dec.Status)
		if err != nil {
			return approved, rejected, ignored, err
		}
		if !changed {
			ignored++
		} else if dec.Status == StatusApproved {
			approved++
		} else {
			rejected++
		}
	}
	return approved, rejected, ignored, nil
}

// ---------------------------------------------------------------------------
// 5b) The "Related (auto)" block — mirror copies ONLY
// ---------------------------------------------------------------------------

var relatedBlockRE = regexp.MustCompile(
	`(?s)` + regexp.QuoteMeta(RelatedStart) + `.*?` + regexp.QuoteMeta(RelatedEnd))

// HasRelatedBlock reports the delimited block's presence.
func HasRelatedBlock(text string) bool {
	return relatedBlockRE.MatchString(text)
}

// ExtractRelatedBlock is the block's text (” when absent).
func ExtractRelatedBlock(text string) string {
	return relatedBlockRE.FindString(text)
}

// RenderRelatedBlock renders the delimited block for one note: a small
// bulleted list of wiki-links (capped by the caller through
// ApprovedLinks).
func RenderRelatedBlock(items []ApprovedLink, titleFor func(string) string) string {
	lines := []string{RelatedStart, "**Related (auto):**"}
	for _, item := range items {
		title := titleFor(item.Other)
		reason := helpers.Strip(item.Reason)
		suffix := ""
		if reason != "" {
			suffix = " — " + reason
		}
		lines = append(lines, fmt.Sprintf("- [[%s]]%s", title, suffix))
	}
	lines = append(lines, RelatedEnd)
	return strings.Join(lines, "\n")
}

// InsertRelatedBlock adds or REPLACES the Related block at the end of a
// mirror note (idempotent; the same surgery rules as the recall block).
func InsertRelatedBlock(text, block string) string {
	if HasRelatedBlock(text) {
		// literal replacement of the FIRST occurrence only
		loc := relatedBlockRE.FindStringIndex(text)
		if loc == nil {
			return text
		}
		return text[:loc[0]] + block + text[loc[1]:]
	}
	stripped := helpers.Rstrip(text)
	if stripped == "" {
		return block
	}
	sep := "\n"
	if strings.Contains(text, "\r\n") &&
		float64(strings.Count(text, "\r\n")) > float64(strings.Count(text, "\n"))/2 {
		sep = "\r\n"
	}
	return stripped + sep + sep + block + sep
}

// ApplyRelatedBlocks writes the approved Related (auto) blocks into the
// MIRROR copies under <manual>/Library/ — never the machine vaults,
// never the owner's notes. Mirror copies are found by their mirror_of
// frontmatter value (the linking layer's key). Returns the relative
// paths updated/added.
func ApplyRelatedBlocks(libraryRoot string, approved map[string][]ApprovedLink,
	fieldsByKey map[string]*FieldsEntry,
	writer func(path, content string)) []string {

	write := writer
	if write == nil {
		write = func(path, content string) {
			_ = storage.AtomicWriteText(path, content)
		}
	}
	var updated []string
	if libraryRoot == "" {
		return updated
	}
	if fi, err := os.Stat(libraryRoot); err != nil || !fi.IsDir() {
		return updated
	}
	// os.walk order; ReadDir's sorted order is the house's deterministic
	// stand-in (the reference's own tests sort the result before use).
	var walk func(dir string)
	walk = func(dir string) {
		if !isInside(dir, libraryRoot) {
			return
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range entries {
			if e.IsDir() || !strings.HasSuffix(e.Name(), ".md") {
				continue
			}
			fpath := filepath.Join(dir, e.Name())
			text, ok := readTextFile(fpath)
			if !ok {
				continue
			}
			head := helpers.CutRunes(text, 800)
			m := mirrorRE.FindStringSubmatchIndex(head)
			if m == nil {
				continue // not a mirror copy — NEVER touched
			}
			key := links.NormalizeURL(helpers.Strip(unquote(head[m[2]:m[3]])))
			items := approved[key]
			// v0.68.0 — THE RESOLVABLE LINK LAW: the block links by
			// FILENAME STEM (with the title as alias).
			targetFor := func(other string) string {
				return LinkTargetFor(fieldsByKey[other])
			}
			newBlock := ""
			if len(items) > 0 {
				newBlock = RenderRelatedBlock(items, targetFor)
			}
			var newText string
			if newBlock != "" {
				newText = InsertRelatedBlock(text, newBlock)
			} else {
				newText = helpers.Rstrip(relatedBlockRE.ReplaceAllString(text, ""))
				if strings.HasSuffix(text, "\n") {
					newText += "\n"
				}
				if !HasRelatedBlock(text) {
					continue // nothing to add, nothing to remove
				}
			}
			if newText != text {
				write(fpath, newText)
				rel, relErr := filepath.Rel(libraryRoot, fpath)
				if relErr == nil {
					updated = append(updated, strings.ReplaceAll(rel, "\\", "/"))
				}
			}
		}
		for _, e := range entries {
			if e.IsDir() {
				walk(filepath.Join(dir, e.Name()))
			}
		}
	}
	walk(libraryRoot)
	return updated
}

// ---------------------------------------------------------------------------
// The mirror carry-over hook (Phase 5 ↔ Phase 6 integration)
// ---------------------------------------------------------------------------

// PreserveRelatedBlock carries an existing Related (auto) block into a
// rebuilt mirror note: mirror's planner rebuilds mirror copies from the
// machine vault on every sync — without this hook the rebuild would
// silently wipe the Related blocks the owner approved.
func PreserveRelatedBlock(oldText, newText string) string {
	block := ExtractRelatedBlock(oldText)
	if block == "" {
		return newText
	}
	return InsertRelatedBlock(newText, block)
}

// clockNow is the package's clock seam (the house pattern): every
// time.strftime the reference's linking/embeddings/recall modules write
// goes through here, so tests and the golden can freeze the clock.
var clockNow = time.Now
