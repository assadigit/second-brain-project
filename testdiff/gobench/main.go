// gobench — the Go half of the performance baselines' scan case: the
// AC-17 vault-scan wall-time rig, mirroring testdiff/python_baselines.py's
// mode_scan (T-04) on the Go side.
//
// This is test instrumentation, not a shipped command (the goside
// pattern): it exposes the production door — internal/vaultscan's
// ScanVaultInventory, the exact engine the sync/auto scan passes and
// the GUI's scan views call (nil log, nil progress, exactly like
// internal/syncreview's production call) — behind one mode verb:
//
//	gobench scan [--vault DIR] [--out FILE]
//
// It builds the SAME synthetic owner-scale corpus the committed T-04
// rig builds (1,309 notes per doc 07 §B.4 + 5 root notes, six
// categories with subfolders, the same frontmatter template — the
// generator below is a byte-mirror of python_baselines.py's
// _build_vault; the artifact's tree-hash proof pins the two builders
// equal), then times the four passes T-04 timed, in one fresh process:
//
//	pass 1  cold-nocache    ScanVaultInventory(vault, nil, nil, nil, kind)
//	pass 2  reread-nocache  the same call again (no cache in-process either)
//	pass 3  cache-fill      notereads.New(<tmp>/reads.db) + the scan through it
//	pass 4  cache-warm      the same cache object again
//
// The JSON carries the same field names as the Python rig's scan mode
// (plus inventory_counts — the scalar census, so the artifact can pin
// the two engines' outputs equal on the same tree) behind the same
// ###BASELINE_JSON### sentinel. With --vault DIR the corpus is built
// into DIR and KEPT (the Python side then scans the exact same bytes);
// the default builds into a fresh temp dir and removes it.
//
// The recorded runs, the verdict, and the rerun recipe live in
// docs/poc-artifacts/ac17-go-vaultscan-benchmark.md.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/notereads"
	"github.com/assadigit/second-brain-project/internal/vaultscan"
)

// benchNNotes is the owner's real vault size per doc 07 §B.4 — the
// same 1,309 python_baselines.py builds (its N_NOTES).
const benchNNotes = 1309

// benchCategories mirrors _build_vault's category table, in order.
var benchCategories = []struct {
	Cat  string
	Subs []string
}{
	{"Tech", []string{"AI", "DevOps", "Security"}},
	{"Science", []string{"Physics", "Biology"}},
	{"Design", []string{"Typography", "UX"}},
	{"Business", []string{"Strategy", "Marketing"}},
	{"Culture", []string{"Film", "Books"}},
	{"Health", []string{"Nutrition", "Sleep"}},
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: gobench scan [--vault DIR] [--out FILE]  (see the file header)")
	os.Exit(2)
}

func main() {
	if len(os.Args) < 2 || os.Args[1] != "scan" {
		usage()
	}
	vault, out := "", ""
	args := os.Args[2:]
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--vault":
			if i+1 >= len(args) {
				usage()
			}
			i++
			vault = args[i]
		case "--out":
			if i+1 >= len(args) {
				usage()
			}
			i++
			out = args[i]
		default:
			usage()
		}
	}
	if out == "" {
		out = "/tmp/gobench-scan.json"
	}
	scanMode(vault, out)
}

// scanOut is the rig's JSON — the Python rig's scan-mode field names
// (so the two artifacts line up column for column), plus mode/go in
// place of the Python rig's mode/python, plus inventory_counts.
type scanOut struct {
	Mode                  string         `json:"mode"`
	Go                    string         `json:"go"`
	SyntheticNotesCreated int            `json:"synthetic_notes_created"`
	NotesSeenByScan       int            `json:"notes_seen_by_scan"`
	ScanColdNocacheMs     float64        `json:"scan_cold_nocache_ms"`
	ScanRereadNocacheMs   float64        `json:"scan_reread_nocache_ms"`
	ScanCacheFillMs       float64        `json:"scan_cache_fill_ms"`
	ScanCacheWarmMs       float64        `json:"scan_cache_warm_ms"`
	CacheWarmVsColdRatio  float64        `json:"cache_warm_vs_cold_ratio"`
	NotesPerSecondCold    float64        `json:"notes_per_second_cold"`
	InventoryCounts       map[string]int `json:"inventory_counts"`
}

// scanMode builds the corpus and times the four passes (the T-04
// procedure, mirrored: same passes, same order, same rounding).
func scanMode(vault, outPath string) {
	tmp, err := os.MkdirTemp("", "gc-gobench-scan-")
	if err != nil {
		fail(err)
	}
	defer os.RemoveAll(tmp)
	keep := vault != ""
	target := filepath.Join(tmp, "vault")
	if keep {
		target = vault
		if err := os.MkdirAll(target, 0o755); err != nil {
			fail(err)
		}
	}
	made, err := buildVault(target)
	if err != nil {
		fail(err)
	}

	// pass 1: no cache (the old direct head reads) — COLD, fresh process.
	t0 := time.Now()
	inv1 := vaultscan.ScanVaultInventory(target, nil, nil, nil,
		vaultscan.VaultKindWebsites)
	msCold := msSince(t0)
	// pass 2: no cache again — the re-read cost (no cache in-process either).
	t0 = time.Now()
	_ = vaultscan.ScanVaultInventory(target, nil, nil, nil,
		vaultscan.VaultKindWebsites)
	msReread := msSince(t0)
	// pass 3+4: the ONE-OPEN LAW — the reads cache populated, then warm.
	cache := notereads.New(filepath.Join(tmp, "reads.db"))
	t0 = time.Now()
	_ = vaultscan.ScanVaultInventory(target, nil, nil, cache,
		vaultscan.VaultKindWebsites)
	msCacheFill := msSince(t0)
	t0 = time.Now()
	_ = vaultscan.ScanVaultInventory(target, nil, nil, cache,
		vaultscan.VaultKindWebsites)
	msCacheWarm := msSince(t0)
	cache.Close(false)

	notesSeen := len(inv1.Notes) + inv1.RootNotes
	res := scanOut{
		Mode:                  "scan",
		Go:                    runtime.Version(),
		SyntheticNotesCreated: made,
		NotesSeenByScan:       notesSeen,
		ScanColdNocacheMs:     round1(msCold),
		ScanRereadNocacheMs:   round1(msReread),
		ScanCacheFillMs:       round1(msCacheFill),
		ScanCacheWarmMs:       round1(msCacheWarm),
		CacheWarmVsColdRatio:  round3(msCacheWarm / msCold),
		NotesPerSecondCold:    round1(float64(notesSeen) / (msCold / 1000.0)),
		InventoryCounts: map[string]int{
			"notes":               len(inv1.Notes),
			"root_notes":          inv1.RootNotes,
			"total_notes":         inv1.TotalNotes,
			"uncategorized_notes": inv1.UncategorizedNotes,
			"hand_notes":          inv1.HandNotes,
			"core_notes":          inv1.CoreNotes,
			"locked_notes":        inv1.LockedNotes,
			"curated_notes":       inv1.CuratedNotes,
			"trash_notes":         inv1.TrashNotes,
			"folders":             len(inv1.Folders),
		},
	}
	b, err := json.MarshalIndent(res, "", "  ")
	if err != nil {
		fail(err)
	}
	if err := os.WriteFile(outPath, append(b, '\n'), 0o644); err != nil {
		fail(err)
	}
	fmt.Println("###BASELINE_JSON###" + string(b))
	if keep {
		fmt.Fprintf(os.Stderr, "[gobench] corpus kept at %s\n", target)
	}
}

// buildVault is _build_vault's byte-mirror: the same categories, the
// same counts (per_cat=218, cat 0 takes the extra 1; +5 root notes),
// the same frontmatter template and the same filenames. The artifact's
// tree-hash proof pins this builder and the Python one equal.
func buildVault(root string) (int, error) {
	const noteTmpl = "---\n" +
		"title: %s\n" +
		"url: https://example-%d.com/articles/%s\n" +
		"domain: example-%d.com\n" +
		"category: %s\n" +
		"subcategory: %s\n" +
		"date_processed: 2025-0%d-1%d\n" +
		"stars: %d\n" +
		"kind: website\n" +
		"---\n\n" +
		"# %s\n\nA synthetic baseline note with a realistic body length.\n"
	made := 0
	perCat := benchNNotes / len(benchCategories)
	extra := benchNNotes - perCat*len(benchCategories)
	for ci, cc := range benchCategories {
		count := perCat
		if ci == 0 {
			count += extra
		}
		for i := 0; i < count; i++ {
			sub := cc.Subs[i%len(cc.Subs)]
			folder := filepath.Join(root, cc.Cat, sub)
			if err := os.MkdirAll(folder, 0o755); err != nil {
				return 0, err
			}
			title := fmt.Sprintf("Note %s %s %04d", cc.Cat, sub, i)
			slug := fmt.Sprintf("n%04d", i)
			body := fmt.Sprintf(noteTmpl, title, ci%5, slug, ci%5,
				cc.Cat, sub, (i%9)+1, i%10, i%5, title)
			name := strings.ReplaceAll(title, " ", "-") + ".md"
			if err := os.WriteFile(filepath.Join(folder, name),
				[]byte(body), 0o644); err != nil {
				return 0, err
			}
			made++
		}
	}
	for i := 0; i < 5; i++ { // a few root notes
		title := fmt.Sprintf("Root Note %d", i)
		slug := fmt.Sprintf("r%d", i)
		body := fmt.Sprintf(noteTmpl, title, 9, slug, 9, "", "", 1, 1, 0, title)
		if err := os.WriteFile(filepath.Join(root,
			fmt.Sprintf("Root Note %d.md", i)), []byte(body), 0o644); err != nil {
			return 0, err
		}
		made++
	}
	return made, nil
}

func msSince(t time.Time) float64 {
	return float64(time.Since(t).Nanoseconds()) / 1e6
}

func round1(v float64) float64 {
	return math.Round(v*10) / 10
}

func round3(v float64) float64 {
	return math.Round(v*1000) / 1000
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "[gobench]", err)
	os.Exit(1)
}
