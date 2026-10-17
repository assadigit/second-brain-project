// index.go — build_index (directory_sync.py lines 820-1008): the vault
// walk → directory.json. THE VAULT MIRROR LAW (v0.88.0): the published
// category tree IS the vault's folder tree — the config is a metadata
// registry only (ghost entries never publish; undeclared folders
// publish with prettified names). PRIVATE nodes never leave this
// machine inside the index (SPEC §12 — filtered here AND again in the
// Worker). The index tree is an *storage.OrderedMap so the publish
// body's bytes ride the reference's exact key order.
package directorysync

import (
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"golang.org/x/text/unicode/norm"
)

// DirectorySchema mirrors DIRECTORY_SCHEMA: the index's shape version
// (the Worker checks it).
const DirectorySchema = "gitcurator-directory/1"

// DirectoryNoteFilename mirrors DIRECTORY_NOTE_FILENAME: the app's
// directory-note filename (never indexed).
const DirectoryNoteFilename = "000 📚 Website Directory.md"

// NewDays mirrors new_days: how many days a site counts as new.
const NewDays = 14

// clockNow is the package's clock seam (the house pattern): the
// index's generated_at, the trash stamp, the banish stamps, and the
// mirror commit message ride it; tests freeze it.
var clockNow = time.Now

// skipDirs mirrors _SKIP_DIRS: folders the index walk never enters.
var skipDirs = map[string]bool{
	"_inbox": true, "_review": true, "_missing": true, "_moc": true,
	"_system": true, "attachments": true, ".obsidian": true, ".trash": true,
	".git": true, "node_modules": true,
}

// readTextLimit mirrors _read_text's 512_000-character cap.
const readTextLimit = 512_000

// utcISOSeconds mirrors datetime.now(timezone.utc).isoformat(
// timespec='seconds') — "2006-01-02T15:04:05+00:00".
func utcISOSeconds(t time.Time) string {
	return t.UTC().Format("2006-01-02T15:04:05+00:00")
}

// localStamp mirrors datetime.now().strftime('%Y%m%dT%H%M%S') — the
// trash filename stamp (local wall clock).
func localStamp(t time.Time) string {
	return t.Format("20060102T150405")
}

// slugify mirrors _slugify: NFKD-fold the lowercased text, strip the
// U+0300-036F combining marks, keep [a-z0-9] + the Persian block,
// collapse every other run to '-', strip the edges, 'site' when
// empty.
func slugify(text string) string {
	folded := norm.NFKD.String(helpers.PyLower(text))
	var nb strings.Builder
	for _, r := range folded {
		if r >= 0x0300 && r <= 0x036f {
			continue
		}
		nb.WriteRune(r)
	}
	var b strings.Builder
	prevSep := false
	for _, r := range nb.String() {
		if isSlugRune(r) {
			b.WriteRune(r)
			prevSep = false
		} else if !prevSep {
			b.WriteRune('-')
			prevSep = true
		}
	}
	base := strings.Trim(b.String(), "-")
	if base == "" {
		return "site"
	}
	return base
}

func isSlugRune(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') ||
		(r >= 0x0600 && r <= 0x06FF)
}

// readText mirrors _read_text: UTF-8 with errors='replace' (the
// maximal-subpart U+FFFD law), universal newlines, capped at
// readTextLimit CHARACTERS (Python text-mode read(n) counts
// characters post-translation).
func readText(path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return helpers.CutRunes(helpers.PyNewlines(helpers.DecodeUTF8Replace(data)), readTextLimit)
}

// iterNotes mirrors _iter_notes: yield (relative_path, abs_path) for
// every .md under the vault's Category/Subcategory tree — system
// folders and dot-dirs skipped, the directory note itself never
// indexed. Sorted for determinism (os.walk's files-before-subdirs
// order, the T-45 law).
func iterNotes(vault string, visit func(rel, absPath string)) {
	if vault == "" {
		return
	}
	if fi, err := os.Stat(vault); err != nil || !fi.IsDir() {
		return
	}
	walkNotes(vault, vault, visit)
}

func walkNotes(vault, dir string, visit func(rel, absPath string)) {
	dirents, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var dirnames, filenames []string
	for _, e := range dirents {
		if e.IsDir() {
			dirnames = append(dirnames, e.Name())
		} else {
			filenames = append(filenames, e.Name())
		}
	}
	// dirs[:] = sorted(d for d in dirs if d not in _SKIP_DIRS and not
	// d.startswith('.')) — pruned IN PLACE (never descended into).
	var kept []string
	for _, d := range dirnames {
		if skipDirs[d] || strings.HasPrefix(d, ".") {
			continue
		}
		kept = append(kept, d)
	}
	sort.Strings(kept)
	sort.Strings(filenames)
	for _, name := range filenames {
		if strings.HasSuffix(strings.ToLower(name), ".md") && name != DirectoryNoteFilename {
			abs := filepath.Join(dir, name)
			rel, err := filepath.Rel(vault, abs)
			if err != nil {
				continue
			}
			visit(filepath.ToSlash(rel), abs)
		}
	}
	for _, d := range kept {
		walkNotes(vault, filepath.Join(dir, d), visit)
	}
}

// catPair is one (category, subcategory) on-disk pair — Python's
// tuple set member (sub "" = Python None: top-level notes).
type catPair struct {
	cat string
	sub string
}

// publicCategoryRecords mirrors _public_category_records: THE MIRROR
// LAW's tree build — exactly the (category, subcategory) pairs that
// carry indexed notes on disk; the config's metadata (name_fa, icon,
// order, description, slugs) rides the nodes it knows, safe defaults
// for the ones it does not; private nodes never leave.
//
// The reference's on_disk_cats is a Python SET (hash-randomized
// iteration); the Go convergence is the walk's first-seen order (the
// T-48 ReadDir-order family) — the final (order, lower(name)) STABLE
// sort makes both sides identical whenever the keys are distinct,
// which every real vault (and the golden fixture) guarantees.
func publicCategoryRecords(config any, onDisk []catPair, private []string) []any {
	privFolded := map[string]bool{}
	for _, p := range private {
		privFolded[strings.ToLower(helpers.Strip(p))] = true
		privFolded[strings.ToLower(helpers.Strip(PrettifyFolderName(storage.SafeFilename(p))))] = true
	}

	get, hasConfig := dictOf(config)
	if !hasConfig {
		get = func(string) any { return nil }
	}

	// The metadata registries (name → config node), from the config
	// alone — insertion-ordered.
	catMeta := map[string]dictReader{}
	catMetaOrder := []string{}
	subMeta := map[string]map[string]dictReader{}
	subMetaOrder := map[string][]string{}
	for _, cat := range dictList(get("categories")) {
		name := strings.TrimSpace(pyStrOrEmpty(cat("name")))
		if name == "" {
			continue
		}
		if _, seen := catMeta[name]; !seen {
			catMetaOrder = append(catMetaOrder, name)
		}
		catMeta[name] = cat
		if subMeta[name] == nil {
			subMeta[name] = map[string]dictReader{}
		}
		for _, sub := range dictList(cat("subcategories")) {
			subName := strings.TrimSpace(pyStrOrEmpty(sub("name")))
			if subName == "" {
				continue
			}
			if _, seen := subMeta[name][subName]; !seen {
				subMetaOrder[name] = append(subMetaOrder[name], subName)
			}
			subMeta[name][subName] = sub
		}
	}

	// on_disk → the public pairs tree: category → the sub names that
	// actually carry notes (first-seen order; the '' sub key of
	// top-level notes never becomes a node).
	catOrder := []string{}
	catSubs := map[string]map[string]bool{}
	for _, pair := range onDisk {
		catName := pair.cat
		if catName == "" || privFolded[strings.ToLower(helpers.Strip(catName))] {
			continue
		}
		subName := helpers.Strip(pair.sub)
		if _, seen := catSubs[catName]; !seen {
			catSubs[catName] = map[string]bool{}
			catOrder = append(catOrder, catName)
		}
		if subName != "" && !privFolded[strings.ToLower(helpers.Strip(subName))] {
			catSubs[catName][subName] = true
		}
	}

	seenSlugs := map[string]bool{}
	uniqueSlug := func(base string) string {
		slug := strings.TrimSpace(base)
		if slug == "" {
			slug = "category"
		}
		for seenSlugs[slug] {
			slug = slug + "-2"
		}
		seenSlugs[slug] = true
		return slug
	}

	out := []any{}
	for _, catName := range catOrder {
		var meta dictReader = func(string) any { return nil }
		if m, ok := catMeta[catName]; ok {
			meta = m
		}
		subNodes := []any{} // Python's [] (never None — the publish body's bytes)
		// Subcategory order: the config's order where known, then the
		// unknown folders alphabetically (a stable, honest order).
		subs := catSubs[catName]
		var declared []string
		for _, s := range subMetaOrder[catName] {
			if subs[s] {
				declared = append(declared, s)
			}
		}
		var undeclared []string
		for s := range subs {
			if subMeta[catName] != nil {
				if _, ok := subMeta[catName][s]; ok {
					continue
				}
			}
			undeclared = append(undeclared, s)
		}
		sort.Strings(undeclared)
		for _, s := range append(declared, undeclared...) {
			if privFolded[strings.ToLower(helpers.Strip(s))] {
				continue
			}
			var sm dictReader = func(string) any { return nil }
			if m, ok := subMeta[catName][s]; ok {
				sm = m
			}
			node := storage.NewOrderedMap()
			node.Set("name", s)
			node.Set("name_fa", pyStrOrEmpty(pyOrValue(sm("name_fa"), "")))
			node.Set("slug", uniqueSlug(pyStrOr(sm("slug"), slugify(s))))
			node.Set("icon", pyOrValue(sm("icon"), nil))
			node.Set("visibility", "public")
			node.Set("description", pyStrOrEmpty(pyOrValue(sm("description"), "")))
			subNodes = append(subNodes, node)
		}
		node := storage.NewOrderedMap()
		node.Set("name", catName)
		node.Set("name_fa", pyStrOrEmpty(pyOrValue(meta("name_fa"), "")))
		node.Set("slug", uniqueSlug(pyStrOr(meta("slug"), slugify(catName))))
		node.Set("icon", pyOrValue(meta("icon"), "bookmark"))
		node.Set("order", orderOr999(meta("order")))
		node.Set("visibility", "public")
		node.Set("description", pyStrOrEmpty(pyOrValue(meta("description"), "")))
		node.Set("subcategories", subNodes)
		out = append(out, node)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i].(*storage.OrderedMap), out[j].(*storage.OrderedMap)
		oi, _ := a.Get("order")
		oj, _ := b.Get("order")
		ni, _ := a.Get("name")
		nj, _ := b.Get("name")
		if pyIntOf(oi) != pyIntOf(oj) {
			return pyIntOf(oi) < pyIntOf(oj)
		}
		return helpers.PyLower(ni.(string)) < helpers.PyLower(nj.(string))
	})
	return out
}

// banishedHostWall mirrors _banished_host_wall: the host-level banish
// wall for the payload (SPEC §11.2) — every banned host that is
// neither a shared-host platform nor a subdomain of one (those bans
// stay URL-level and must never block the platform).
func banishedHostWall(banished []BanishEntry, sharedHosts []string) []string {
	wall := map[string]bool{}
	for _, entry := range banished {
		host := helpers.Strip(strings.TrimLeft(helpers.PyLower(entry.Host), "."))
		if host == "" {
			continue
		}
		blocked := false
		for _, h := range sharedHosts {
			if host == h || strings.HasSuffix(host, "."+h) {
				blocked = true
				break
			}
		}
		if blocked {
			continue
		}
		wall[host] = true
	}
	out := make([]string, 0, len(wall))
	for h := range wall {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

// BuildIndex mirrors build_index: the vault → directory.json.
// PRIVATE nodes never leave (SPEC §12): filtered here AND again in the
// Worker. Banished hosts never leave. Hand-written notes ride along
// when they carry a source: URL — the vault is the owner's truth, not
// just the app's notes. privateNames nil → DefaultPrivateNames; an
// empty non-nil slice stays empty (Python None vs []). log may be nil.
func BuildIndex(vault string, privateNames []string, log func(string)) *storage.OrderedMap {
	if log == nil {
		log = func(string) {}
	}
	config := LoadDirectoryConfig(vault)
	synonyms := LoadSynonyms(vault)
	banished := LoadBanishList(vault)

	// shared_hosts = [str(h).lower() for h in
	// (config.get('shared_host_exceptions') or DEFAULT_SHARED_HOSTS)]
	sharedHosts := sharedHostsOf(config)

	private := map[string]bool{}
	if privateNames == nil {
		for _, n := range DefaultPrivateNames {
			private[n] = true
		}
	} else {
		for _, n := range privateNames {
			private[n] = true
		}
	}
	// The config file's own visibility marks join the belt.
	cfgGet, _ := dictOf(config)
	for _, cat := range dictList(cfgGet("categories")) {
		if vis, _ := cat("visibility").(string); vis == "private" {
			private[strings.TrimSpace(pyStrOrEmpty(cat("name")))] = true
		}
		for _, sub := range dictList(cat("subcategories")) {
			if vis, _ := sub("visibility").(string); vis == "private" {
				private[strings.TrimSpace(pyStrOrEmpty(sub("name")))] = true
			}
		}
	}
	privateFolded := map[string]bool{}
	for p := range private {
		privateFolded[strings.ToLower(helpers.Strip(p))] = true
		privateFolded[strings.ToLower(helpers.Strip(PrettifyFolderName(storage.SafeFilename(p))))] = true
	}

	// v0.72.14 — the folder-name law, v0.88.0 case-insensitive: only a
	// UNIQUE sanitized match is translated; an ambiguous name never
	// guesses.
	folderNames := map[string]map[string]bool{}
	for _, cat := range dictList(cfgGet("categories")) {
		name := strings.TrimSpace(pyStrOrEmpty(cat("name")))
		if name == "" {
			continue
		}
		names := []string{name}
		for _, sub := range dictList(cat("subcategories")) {
			subName := strings.TrimSpace(pyStrOrEmpty(sub("name")))
			if subName != "" {
				names = append(names, subName)
			}
		}
		for _, n := range names {
			key := strings.ToLower(storage.SafeFilename(n))
			if folderNames[key] == nil {
				folderNames[key] = map[string]bool{}
			}
			folderNames[key][n] = true
		}
	}
	folderToName := map[string]string{}
	for k, v := range folderNames {
		if len(v) == 1 {
			for n := range v {
				folderToName[k] = n
			}
		}
	}

	sites := []any{}
	seenIDs := map[string]bool{}
	seenSlugs := map[string]bool{}
	tagCounts := map[string]int{}
	tagCountsOrder := []string{} // insertion order for determinism (Python counts unordered)
	var onDisk []catPair

	iterNotes(vault, func(rel, absPath string) {
		parsed := ParseNote(readText(absPath))
		if parsed == nil {
			return
		}
		normalized := links.NormalizeWebsiteURL(parsed.Source)
		if normalized == "" || !strings.HasPrefix(normalized, "http") {
			return
		}
		if IsBanished(parsed.Source, banished, sharedHosts) {
			log("[directory] skipped: banished (" + normalized + ")")
			return
		}

		parts := strings.Split(strings.ReplaceAll(rel, "\\", "/"), "/")
		parts = parts[:len(parts)-1] // the filename
		rawCategory := ""
		if len(parts) > 0 {
			rawCategory = strings.TrimSpace(parts[0])
		}
		rawSubcategory := ""
		if len(parts) > 1 {
			rawSubcategory = strings.TrimSpace(parts[1])
		}
		// v0.88.0 — THE MIRROR LAW: the registry's declared display
		// name when the folder-name law matches (case-insensitively);
		// otherwise the PRETTIFIED folder name. The SAME string rides
		// the site rows and the category tree.
		category := folderToName[strings.ToLower(rawCategory)]
		if category == "" {
			category = PrettifyFolderName(rawCategory)
		}
		subcategory := ""
		if rawSubcategory != "" {
			subcategory = folderToName[strings.ToLower(rawSubcategory)]
			if subcategory == "" {
				subcategory = PrettifyFolderName(rawSubcategory)
			}
		}
		if category == "" {
			return
		}
		// The pair joins on_disk only AFTER the private gate (the
		// FOLDED belt) — the tree is the vault's PUBLIC folder tree.
		if privateFolded[strings.ToLower(helpers.Strip(category))] ||
			privateFolded[strings.ToLower(helpers.Strip(subcategory))] {
			log("[directory] skipped: private category (" + category + "/" + subcategory + ")")
			return
		}
		onDisk = append(onDisk, catPair{category, subcategory})

		sum := sha256.Sum256([]byte(normalized))
		siteID := hex.EncodeToString(sum[:])[:16]
		if seenIDs[siteID] {
			return
		}
		seenIDs[siteID] = true
		nameFallback := normHost(normalized)
		displayName := parsed.Name
		if displayName == "" {
			displayName = nameFallback
		}
		slug := slugify(displayName)
		for seenSlugs[slug] {
			slug = slug + "-2"
		}
		seenSlugs[slug] = true

		for _, tag := range parsed.Tags {
			if _, seen := tagCounts[tag]; !seen {
				tagCountsOrder = append(tagCountsOrder, tag)
			}
			tagCounts[tag]++
		}

		site := storage.NewOrderedMap()
		site.Set("id", siteID)
		site.Set("slug", slug)
		site.Set("name", displayName)
		site.Set("url", normalized)
		site.Set("host", normHost(normalized))
		site.Set("category", category)
		site.Set("subcategory", subcategory)
		site.Set("summary", parsed.Summary)
		site.Set("tags", strSliceAny(parsed.Tags))
		site.Set("featured", parsed.Featured)
		site.Set("needs_review", parsed.NeedsReview)
		site.Set("added_at", parsed.AddedAt)
		site.Set("updated_at", parsed.AddedAt)
		site.Set("what_it_does", strSliceAny(parsed.Sections.WhatItDoes))
		site.Set("best_used_for", parsed.Sections.BestUsedFor)
		site.Set("practical_details", strSliceAny(parsed.Sections.PracticalDetails))
		site.Set("watch_out", strSliceAny(parsed.Sections.WatchOut))
		site.Set("similar_tools", parsed.Sections.SimilarTools)
		extras := make([]any, 0, len(parsed.ExtraSections))
		for _, e := range parsed.ExtraSections {
			em := storage.NewOrderedMap()
			em.Set("title", e.Title)
			em.Set("bullets", strSliceAny(e.Bullets))
			em.Set("text", e.Text)
			extras = append(extras, em)
		}
		site.Set("extra_sections", extras)
		site.Set("signup_required", parsed.SignupRequired)
		site.Set("pricing", parsed.Pricing)
		site.Set("detail_markdown", parsed.DetailMarkdown)
		sites = append(sites, site)
	})

	var privateList []string
	for p := range private {
		privateList = append(privateList, p)
	}
	categories := publicCategoryRecords(config, onDisk, privateList)

	// tag_aliases = {str(k): str(v) for k, v in
	// (config.get('tag_aliases') or {}).items() if k and v} — the
	// config's own document order rides (OrderedMap).
	aliases := storage.NewOrderedMap()
	if aliasMap, ok := cfgGet("tag_aliases").(*storage.OrderedMap); ok {
		for _, k := range aliasMap.Keys() {
			v, _ := aliasMap.Get(k)
			if pyTruthyDS(k) && pyTruthyDS(v) {
				aliases.Set(helpers.PyStr(k), helpers.PyStr(v))
			}
		}
	} else if aliasMap, ok := cfgGet("tag_aliases").(map[string]any); ok {
		// a hand-built map config: Python's dict order is lost here —
		// sorted keys are the deterministic convergence (documented)
		keys := make([]string, 0, len(aliasMap))
		for k := range aliasMap {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			v := aliasMap[k]
			if pyTruthyDS(k) && pyTruthyDS(v) {
				aliases.Set(helpers.PyStr(k), helpers.PyStr(v))
			}
		}
	}

	// popular tags: top 12 by count desc, tag asc, the pricing words
	// excluded.
	type tagCount struct {
		tag   string
		count int
	}
	var popular []tagCount
	for _, tag := range tagCountsOrder {
		if tag == "free" || tag == "paid" || tag == "freemium" {
			continue
		}
		popular = append(popular, tagCount{tag, tagCounts[tag]})
	}
	sort.Slice(popular, func(i, j int) bool {
		if popular[i].count != popular[j].count {
			return popular[i].count > popular[j].count
		}
		return popular[i].tag < popular[j].tag
	})
	if len(popular) > 12 {
		popular = popular[:12]
	}
	popularTags := make([]any, 0, len(popular))
	for _, tc := range popular {
		pt := storage.NewOrderedMap()
		pt.Set("tag", tc.tag)
		pt.Set("count", tc.count)
		pt.Set("fa", "")
		if fa, ok := aliases.Get(tc.tag); ok {
			pt.Set("fa", helpers.PyStr(fa))
		}
		popularTags = append(popularTags, pt)
	}

	counts := storage.NewOrderedMap()
	counts.Set("sites", len(sites))
	counts.Set("categories", len(categories))

	index := storage.NewOrderedMap()
	index.Set("schema", DirectorySchema)
	index.Set("generated_at", utcISOSeconds(clockNow()))
	index.Set("counts", counts)
	index.Set("categories", categories)
	index.Set("sites", sites)
	index.Set("popular_tags", popularTags)
	// v0.72.14 — the FULL alias map rides the payload (UI spec §4).
	index.Set("tag_aliases", aliases)
	index.Set("synonyms", synonyms)
	// v0.76.0 — the durable banish wall rides the payload (SPEC §11.2).
	index.Set("banished_hosts", banishedHostWall(banished, sharedHosts))
	index.Set("new_days", NewDays)
	log("[directory] built index: " + strconv.Itoa(len(sites)) + " sites, " +
		strconv.Itoa(len(categories)) + " categories, " +
		strconv.Itoa(len(popularTags)) + " popular tags")
	return index
}

// sharedHostsOf mirrors the config read: [str(h).lower() for h in
// (config.get('shared_host_exceptions') or DEFAULT_SHARED_HOSTS)].
// A list value rides str()-lowered; a dict value iterates its KEYS
// (Python's dict-iteration quirk); anything else falls to the
// defaults (a documented narrowing of Python's string-iteration
// quirk).
func sharedHostsOf(config any) []string {
	get, _ := dictOf(config)
	v := get("shared_host_exceptions")
	if !pyTruthyDS(v) {
		return append([]string(nil), DefaultSharedHosts...)
	}
	if items, ok := v.([]any); ok {
		out := make([]string, 0, len(items))
		for _, h := range items {
			out = append(out, helpers.PyLower(helpers.PyStr(h)))
		}
		return out
	}
	if om, ok := v.(*storage.OrderedMap); ok {
		out := make([]string, 0, om.Len())
		for _, k := range om.Keys() {
			out = append(out, helpers.PyLower(helpers.PyStr(k)))
		}
		return out
	}
	if m, ok := v.(map[string]any); ok {
		out := make([]string, 0, len(m))
		for k := range m {
			out = append(out, helpers.PyLower(k))
		}
		sort.Strings(out)
		return out
	}
	return append([]string(nil), DefaultSharedHosts...)
}

// pyStrOrEmpty mirrors str(v or "").
func pyStrOrEmpty(v any) string {
	if !pyTruthyDS(v) {
		return ""
	}
	return helpers.PyStr(v)
}

// pyOrValue mirrors (v or fallback): a falsy value answers the
// fallback, a truthy value rides as-is.
func pyOrValue(v, fallback any) any {
	if !pyTruthyDS(v) {
		return fallback
	}
	return v
}

// pyStrOr mirrors str(v or fallback).
func pyStrOr(v any, fallback string) string {
	if !pyTruthyDS(v) {
		return fallback
	}
	return helpers.PyStr(v)
}

// orderOr999 mirrors meta.get('order') if isinstance(meta.get('order'),
// int) else 999 — Python's bool IS an int (True → 1, False → 0).
func orderOr999(v any) int {
	switch t := v.(type) {
	case bool:
		if t {
			return 1
		}
		return 0
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return 999 // a float is NOT an int in Python
	default:
		return 999
	}
}

// pyIntOf reads an int-shaped value (the sort keys).
func pyIntOf(v any) int {
	switch t := v.(type) {
	case bool:
		if t {
			return 1
		}
		return 0
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	default:
		return 999
	}
}

// pyTruthyDS mirrors Python truthiness for the config shapes (the
// vaultseal bridges' pyTruthy twin).
func pyTruthyDS(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case bool:
		return t
	case string:
		return t != ""
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case *storage.OrderedMap:
		return t.Len() > 0
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}
