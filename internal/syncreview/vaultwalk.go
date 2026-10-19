package syncreview

// vaultwalk.go — the two local walks + the count door: the Websites
// vault's public site-note walk (build_index's own filter laws,
// keeping the note path + added_at), the vault's git backup gap (the
// read-only `git status --porcelain` walk), and the cheap note count.
// Read-only everywhere: never a write verb, never a byte of the vault
// touched.

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/directorysync"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
)

func quietLog(log func(string)) func(string) {
	if log == nil {
		return func(string) {}
	}
	return log
}

// VaultSiteItems mirrors vault_site_items: the Websites vault's PUBLIC
// site notes — the source side of the flagship stage, one walk,
// mirroring build_index's own filter laws (banished hosts out, private
// categories out, the folder-name translation) while keeping what
// build_index sends away: each item carries its NOTE PATH (the "Open
// note" door) and its added_at (the relative time + the oldest-first
// sort). ok=false when the vault is not set up (the honest Unknown);
// never raises; reads only.
func VaultSiteItems(vault string, log func(string), cap int) (items []GapItem, ok bool) {
	log = quietLog(log)
	if cap <= 0 {
		cap = 2000
	}
	vault = strings.TrimSpace(vault)
	if vault == "" {
		return nil, false
	}
	if fi, err := os.Stat(vault); err != nil || !fi.IsDir() {
		return nil, false
	}

	config := directorysync.LoadDirectoryConfig(vault)
	banished := directorysync.LoadBanishList(vault)
	sharedHosts := directorysync.SharedHostsOf(config)
	private := map[string]bool{}
	for _, n := range directorysync.DefaultPrivateNames {
		private[n] = true
	}
	for _, cat := range dictList(config) {
		if catVisibility(cat) == "private" {
			private[strings.TrimSpace(catName(cat))] = true
		}
		for _, sub := range catSubcategories(cat) {
			if dictValueStr(sub, "visibility") == "private" {
				private[strings.TrimSpace(dictValueStr(sub, "name"))] = true
			}
		}
	}

	// The folder-name law (build_index's own translation): disk's
	// 'Assets_Resources' answers to the taxonomy's 'Assets & Resources'.
	folderNames := map[string]map[string]bool{}
	for _, cat := range dictList(config) {
		name := catName(cat)
		if name == "" {
			continue
		}
		names := []string{name}
		for _, sub := range catSubcategories(cat) {
			if n := dictValueStr(sub, "name"); n != "" {
				names = append(names, n)
			}
		}
		for _, n := range names {
			key := storage.SafeFilename(n)
			if folderNames[key] == nil {
				folderNames[key] = map[string]bool{}
			}
			folderNames[key][n] = true
		}
	}
	folderToName := map[string]string{}
	for key, set := range folderNames {
		if len(set) == 1 {
			for n := range set {
				folderToName[key] = n
			}
		}
	}

	items = []GapItem{}
	seen := map[string]bool{}
	capped := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				// never raise on a walk — the items gathered so far are
				// the honest answer
				log("[sync] vault walk failed")
			}
		}()
		directorysync.IterNotes(vault, func(rel, absPath string) {
			if capped {
				return
			}
			if len(items) >= cap {
				log("[sync] vault walk capped at " + itoa(cap) + " site notes")
				capped = true
				return
			}
			parsed := directorysync.ParseNote(directorysync.ReadText(absPath))
			if parsed == nil {
				return
			}
			normalized := links.NormalizeWebsiteURL(parsed.Source)
			if normalized == "" || !strings.HasPrefix(normalized, "http") {
				return
			}
			if directorysync.IsBanished(parsed.Source, banished, sharedHosts) {
				return
			}
			parts := strings.FieldsFunc(strings.ReplaceAll(rel, "\\", "/"),
				func(r rune) bool { return r == '/' })
			parts = parts[:len(parts)-1] // the filename
			rawCategory := strings.TrimSpace(firstOr(parts, ""))
			rawSub := ""
			if len(parts) > 1 {
				rawSub = strings.TrimSpace(parts[1])
			}
			category := rawCategory
			if n, ok := folderToName[rawCategory]; ok {
				category = n
			}
			subcategory := rawSub
			if rawSub != "" {
				if n, ok := folderToName[rawSub]; ok {
					subcategory = n
				}
			}
			if private[category] || private[subcategory] {
				return
			}
			if seen[normalized] {
				return
			}
			seen[normalized] = true
			name := parsed.Name
			if name == "" {
				name = directorysync.NormHost(normalized)
			}
			items = append(items, GapItem{
				ID: normalized, Name: name, URL: normalized,
				AddedAt: parsed.AddedAt, NotePath: absPath,
			})
		})
	}()
	return items, true
}

// VaultBackupGap mirrors vault_backup_gap: the Vault → Backup stage's
// gap — notes in the vault the mirror does not carry yet, from the
// vault's own local git status (read-only: `git status --porcelain`,
// never a write verb; the repo_radar law). A vault with NO .git was
// never sealed: the honest gap is every note (the first seal fixes
// it). ok=false when not set up (no vault). Never raises.
func VaultBackupGap(vault string, log func(string), cap int) (items []GapItem, ok bool) {
	log = quietLog(log)
	if cap <= 0 {
		cap = 4000
	}
	vault = strings.TrimSpace(vault)
	if vault == "" {
		return nil, false
	}
	if fi, err := os.Stat(vault); err != nil || !fi.IsDir() {
		return nil, false
	}
	if fi, err := os.Stat(filepath.Join(vault, ".git")); err != nil || !fi.IsDir() {
		// never sealed — every note is pending (the first seal fixes it)
		out := []GapItem{}
		var walk func(dir string)
		walk = func(dir string) {
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
			sort.Strings(dirnames)
			keep := dirnames[:0]
			for _, d := range dirnames {
				if strings.HasPrefix(d, ".") || d == "node_modules" {
					continue
				}
				keep = append(keep, d)
			}
			sort.Strings(filenames)
			for _, name := range filenames {
				if !strings.HasSuffix(strings.ToLower(name), ".md") {
					continue
				}
				if len(out) >= cap {
					log("[sync] backup gap capped at " + itoa(cap) + " notes")
					return
				}
				absPath := filepath.Join(dir, name)
				rel, _ := filepath.Rel(vault, absPath)
				rel = filepath.ToSlash(rel)
				out = append(out, GapItem{ID: rel, Name: rel, NotePath: absPath})
			}
			for _, d := range keep {
				walk(filepath.Join(dir, d))
			}
		}
		walk(vault)
		return out, true
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, "git", "-C", vault, "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		msg := ""
		if ee, isEE := err.(*exec.ExitError); isEE {
			msg = strings.TrimSpace(string(ee.Stderr))
		} else {
			msg = err.Error()
		}
		if len(msg) > 120 {
			msg = msg[:120]
		}
		log("[sync] git status failed: " + msg)
		return []GapItem{}, true
	}
	items = []GapItem{}
	for _, line := range strings.Split(string(out), "\n") {
		line = strings.TrimRight(line, "\r")
		if len(line) < 4 {
			continue
		}
		path := strings.Trim(strings.TrimSpace(line[3:]), `"`)
		// renames carry "old -> new": the NEW path is the item
		if idx := strings.Index(path, " -> "); idx >= 0 {
			path = strings.Trim(strings.TrimSpace(path[idx+4:]), `"`)
		}
		if !strings.HasSuffix(strings.ToLower(path), ".md") {
			continue
		}
		// the seal's own noise never counts as a pending note
		folded := strings.ToLower(strings.ReplaceAll(path, "\\", "/"))
		if strings.HasPrefix(folded, ".obsidian/") || strings.HasPrefix(folded, ".trash/") {
			continue
		}
		if len(items) >= cap {
			log("[sync] backup gap capped at " + itoa(cap) + " notes")
			break
		}
		slashPath := strings.ReplaceAll(path, "\\", "/")
		items = append(items, GapItem{
			ID: slashPath, Name: slashPath,
			NotePath: filepath.Join(vault, path),
		})
	}
	return items, true
}

// VaultNoteCount mirrors vault_note_count: the cheap vault-side count
// (scan_vault_inventory's own totals — the inventory walk, pure local
// reads). nil when not set up.
func VaultNoteCount(vault string) *int {
	vault = strings.TrimSpace(vault)
	if vault == "" {
		return nil
	}
	if fi, err := os.Stat(vault); err != nil || !fi.IsDir() {
		return nil
	}
	count := 0
	scanned := false
	func() {
		defer func() { recover() }()
		inv := vaultscan.ScanVaultInventory(vault, nil, nil, nil,
			vaultscan.VaultKindWebsites)
		count = inv.TotalNotes
		scanned = true
	}()
	if scanned {
		return &count
	}
	// the fallback count: the walk the inventory itself walks
	total := 0
	var walk func(dir string)
	walk = func(dir string) {
		dirents, err := os.ReadDir(dir)
		if err != nil {
			return
		}
		for _, e := range dirents {
			if e.IsDir() {
				if e.Name() == "_inbox" || strings.HasPrefix(e.Name(), ".") {
					continue
				}
				walk(filepath.Join(dir, e.Name()))
			} else if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
				total++
			}
		}
	}
	walk(vault)
	return &total
}

// ---------------------------------------------------------------------------
// tiny junk-tolerant config readers (the Python dict guards' twins)
// ---------------------------------------------------------------------------

// dictList answers the config's 'categories' list as OrderedMaps.
func dictList(config any) []*storage.OrderedMap {
	om, ok := config.(*storage.OrderedMap)
	if !ok {
		return nil
	}
	v, _ := om.Get("categories")
	list, _ := v.([]any)
	var out []*storage.OrderedMap
	for _, item := range list {
		if m, ok := item.(*storage.OrderedMap); ok {
			out = append(out, m)
		}
	}
	return out
}

func catName(cat *storage.OrderedMap) string {
	v, _ := cat.Get("name")
	return pyStr(v)
}

func catVisibility(cat *storage.OrderedMap) string {
	v, _ := cat.Get("visibility")
	return pyStr(v)
}

func catSubcategories(cat *storage.OrderedMap) []*storage.OrderedMap {
	v, _ := cat.Get("subcategories")
	list, _ := v.([]any)
	var out []*storage.OrderedMap
	for _, item := range list {
		if m, ok := item.(*storage.OrderedMap); ok {
			out = append(out, m)
		}
	}
	return out
}

func dictValueStr(om *storage.OrderedMap, key string) string {
	v, _ := om.Get(key)
	return pyStr(v)
}

func firstOr(parts []string, def string) string {
	if len(parts) == 0 {
		return def
	}
	return parts[0]
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	neg := n < 0
	if neg {
		n = -n
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	if neg {
		i--
		buf[i] = '-'
	}
	return string(buf[i:])
}
