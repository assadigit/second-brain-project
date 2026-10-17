// index_test.go — T-49: the ported selections of the reference's
// TestBuildIndex + TestBanishWall (test_directorysync.py) — the
// end-to-end index build on real-shaped temp vaults: THE MIRROR LAW
// (ghosts drop, undeclared folders publish prettified), the folded
// private belt, the banish wall, the dedupe + slug suffix laws.
package directorysync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// vaultBuilder mirrors _VaultBuilder: a temp Websites vault with real
// note files.
type vaultBuilder struct{ dir string }

func newVaultBuilder(t *testing.T) *vaultBuilder {
	t.Helper()
	return &vaultBuilder{dir: t.TempDir()}
}

func (v *vaultBuilder) write(t *testing.T, rel, text string) {
	t.Helper()
	path := filepath.Join(v.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(text), 0o644); err != nil {
		t.Fatal(err)
	}
}

func (v *vaultBuilder) banished(t *testing.T, entries string) {
	t.Helper()
	v.write(t, "_system/banished.json", entries)
}

func (v *vaultBuilder) build() *storage.OrderedMap {
	return BuildIndex(v.dir, nil, nil)
}

// index accessors over the OrderedMap tree.
func idxSites(t *testing.T, index *storage.OrderedMap) []*storage.OrderedMap {
	t.Helper()
	raw, _ := index.Get("sites")
	items, _ := raw.([]any)
	out := make([]*storage.OrderedMap, 0, len(items))
	for _, s := range items {
		out = append(out, s.(*storage.OrderedMap))
	}
	return out
}

func idxCategories(index *storage.OrderedMap) []*storage.OrderedMap {
	raw, _ := index.Get("categories")
	items, _ := raw.([]any)
	out := make([]*storage.OrderedMap, 0, len(items))
	for _, s := range items {
		out = append(out, s.(*storage.OrderedMap))
	}
	return out
}

func omStr(om *storage.OrderedMap, key string) string {
	v, _ := om.Get(key)
	s, _ := v.(string)
	return s
}

func omAny(om *storage.OrderedMap, key string) any {
	v, _ := om.Get(key)
	return v
}

func siteNamed(t *testing.T, index *storage.OrderedMap, name string) *storage.OrderedMap {
	t.Helper()
	for _, s := range idxSites(t, index) {
		if omStr(s, "name") == name {
			return s
		}
	}
	t.Fatalf("site %q not found", name)
	return nil
}

func catNamed(t *testing.T, index *storage.OrderedMap, name string) *storage.OrderedMap {
	t.Helper()
	for _, c := range idxCategories(index) {
		if omStr(c, "name") == name {
			return c
		}
	}
	t.Fatalf("category %q not found", name)
	return nil
}

func catSubNames(cat *storage.OrderedMap) []string {
	raw, _ := cat.Get("subcategories")
	items, _ := raw.([]any)
	var out []string
	for _, s := range items {
		out = append(out, omStr(s.(*storage.OrderedMap), "name"))
	}
	return out
}

func idxCounts(t *testing.T, index *storage.OrderedMap) (sites, cats int) {
	t.Helper()
	raw, _ := index.Get("counts")
	counts, _ := raw.(*storage.OrderedMap)
	sv, _ := counts.Get("sites")
	cv, _ := counts.Get("categories")
	return int(pyIntOf(sv)), int(pyIntOf(cv))
}

func TestBuildIndex_FullVaultShape(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.write(t, "Design/UI-UX/Fast.md", shortNote)
	v.write(t, "Random/Window Swap.md", unstructuredNote)
	index := v.build()
	if omStr(index, "schema") != DirectorySchema {
		t.Fatalf("schema: %q", omStr(index, "schema"))
	}
	sites, cats := idxCounts(t, index)
	if sites != 3 || cats != 2 {
		t.Fatalf("counts: %d sites, %d cats (Design + Random)", sites, cats)
	}
	coolors := siteNamed(t, index, "Coolors")
	if omStr(coolors, "slug") != "coolors" {
		t.Fatalf("slug: %q", omStr(coolors, "slug"))
	}
	if omStr(coolors, "host") != "coolors.co" {
		t.Fatalf("host: %q", omStr(coolors, "host"))
	}
	if omStr(coolors, "signup_required") != "no" {
		t.Fatalf("signup_required: %q", omStr(coolors, "signup_required"))
	}
	wd, _ := coolors.Get("what_it_does")
	if len(wd.([]any)) != 4 {
		t.Fatalf("what_it_does: %v", wd)
	}
	if omStr(coolors, "category") != "Design" {
		t.Fatalf("category: %q", omStr(coolors, "category"))
	}
	if omStr(coolors, "subcategory") != "Assets & Resources" {
		t.Fatalf("subcategory: %q", omStr(coolors, "subcategory"))
	}
	window := siteNamed(t, index, "Window Swap")
	if omStr(window, "detail_markdown") == "" {
		t.Fatal("the unstructured note lost its raw body")
	}
	// Categories carry the config metadata (fa names, icons, order).
	design := catNamed(t, index, "Design")
	if omStr(design, "name_fa") != "طراحی" {
		t.Fatalf("name_fa: %q", omStr(design, "name_fa"))
	}
	if omStr(design, "icon") != "palette" {
		t.Fatalf("icon: %q", omStr(design, "icon"))
	}
	// Popular tags computed from the notes.
	raw, _ := index.Get("popular_tags")
	items, _ := raw.([]any)
	tags := map[string]int{}
	for _, pt := range items {
		m := pt.(*storage.OrderedMap)
		tags[omStr(m, "tag")] = int(pyIntOf(omAny(m, "count")))
	}
	if tags["color-tool"] != 1 {
		t.Fatalf("popular tags: %v", tags)
	}
}

func TestBuildIndex_PrivateSubcategoryNeverInIndex(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	// v0.88.0 — the mirror law: a category node exists only when a
	// PUBLIC note lives under it, so the public sibling note is part
	// of the fixture (not a config ghost).
	gutenberg := refNote(fmFields(map[string]string{"source": "https://gutenberg.org/"}),
		"# Gutenberg\n\n> **TL;DR:** t\n"+refFooter("https://gutenberg.org/"))
	v.write(t, "Download Resources/Public Archives & Libraries/Gutenberg.md", gutenberg)
	v.write(t, "Download Resources/Torrent & Shadow Libraries/Annas.md", privateNote)
	index := v.build()
	blob := storage.PyDumpsUTF8(index)
	if strings.Contains(blob, "annas") || strings.Contains(blob, "Shadow") {
		t.Fatalf("the private node leaked: %s", blob)
	}
	// The PRIVATE node never leaves the machine at all — not even as a
	// category record (SPEC §12). The public sibling still rides along.
	dl := catNamed(t, index, "Download Resources")
	subs := catSubNames(dl)
	if !containsName(subs, "Public Archives & Libraries") {
		t.Fatalf("subs: %v", subs)
	}
	if containsName(subs, "Torrent & Shadow Libraries") {
		t.Fatalf("the private sub rode the tree: %v", subs)
	}
	if strings.Contains(storage.PyDumpsUTF8(dl), "PRIVATE") {
		t.Fatalf("the private mark leaked: %v", dl)
	}
}

func containsName(names []string, want string) bool {
	for _, n := range names {
		if n == want {
			return true
		}
	}
	return false
}

func TestBuildIndex_MirrorLawDropsConfigGhosts(t *testing.T) {
	// v0.88.0 — THE MIRROR LAW: the tree is the vault's folder tree. A
	// config-declared node with no notes on disk is a GHOST and never
	// publishes.
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	index := v.build()
	design := catNamed(t, index, "Design")
	subs := catSubNames(design)
	if !containsName(subs, "Assets & Resources") {
		t.Fatalf("subs: %v", subs)
	}
	if containsName(subs, "Print & Editorial Design") || containsName(subs, "Branding & Identity") {
		t.Fatalf("ghost subs published: %v", subs)
	}
	// Config categories with zero notes on disk drop whole.
	var catNames []string
	for _, c := range idxCategories(index) {
		catNames = append(catNames, omStr(c, "name"))
	}
	if containsName(catNames, "Geography & Urban Planning") || containsName(catNames, "Living Abroad") {
		t.Fatalf("ghost cats published: %v", catNames)
	}
	if !containsName(catNames, "Design") {
		t.Fatalf("Design itself must survive: %v", catNames)
	}
}

func TestBuildIndex_MirrorLawUnknownSubfolderPublishesWithDefaults(t *testing.T) {
	// v0.88.0 — a disk folder the registry never heard of still
	// publishes: the prettified name rides the SITE ROW and the TREE
	// NODE as the SAME string, so the site's subcategory filter
	// matches.
	v := newVaultBuilder(t)
	whim := refNote(fmFields(map[string]string{"source": "https://whimsical.com/"}),
		"# Whimsical\n\n> **TL;DR:** t\n"+refFooter("https://whimsical.com/"))
	v.write(t, "Design/Wireframe_Kits/Whimsical.md", whim)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	index := v.build()
	site := siteNamed(t, index, "Whimsical")
	if omStr(site, "category") != "Design" {
		t.Fatalf("category: %q", omStr(site, "category"))
	}
	if omStr(site, "subcategory") != "Wireframe Kits" {
		t.Fatalf("subcategory: %q", omStr(site, "subcategory"))
	}
	design := catNamed(t, index, "Design")
	nodes := map[string]*storage.OrderedMap{}
	raw, _ := design.Get("subcategories")
	items, _ := raw.([]any)
	for _, s := range items {
		m := s.(*storage.OrderedMap)
		nodes[omStr(m, "name")] = m
	}
	if nodes["Wireframe Kits"] == nil || nodes["Assets & Resources"] == nil {
		t.Fatalf("nodes: %v", catSubNames(design))
	}
	// Safe defaults for the unknown node; metadata for the known one.
	if omStr(nodes["Wireframe Kits"], "slug") != "wireframe-kits" {
		t.Fatalf("slug: %q", omStr(nodes["Wireframe Kits"], "slug"))
	}
	if icon, _ := nodes["Wireframe Kits"].Get("icon"); icon != nil {
		t.Fatalf("icon: %v", icon)
	}
	if omStr(nodes["Wireframe Kits"], "name_fa") != "" {
		t.Fatalf("name_fa: %q", omStr(nodes["Wireframe Kits"], "name_fa"))
	}
	if omStr(nodes["Assets & Resources"], "name_fa") != "دارایی\u200cها و منابع" {
		t.Fatalf("name_fa: %q", omStr(nodes["Assets & Resources"], "name_fa"))
	}
	// Order: the config's declared subs first, then the unknowns
	// alphabetically (a stable, honest order).
	ordered := catSubNames(design)
	if len(ordered) != 2 || ordered[0] != "Assets & Resources" || ordered[1] != "Wireframe Kits" {
		t.Fatalf("ordered subs: %v", ordered)
	}
}

func TestBuildIndex_FolderTranslationIsCaseInsensitive(t *testing.T) {
	// v0.88.0 — case drift between a registry display name and its
	// folder must never break the folder-name law — the folded key
	// translates to the declared display name on BOTH the site row and
	// the tree node.
	v := newVaultBuilder(t)
	icons := refNote(fmFields(map[string]string{"source": "https://icons8.com/"}),
		"# Icons8\n\n> **TL;DR:** t\n"+refFooter("https://icons8.com/"))
	v.write(t, "Design/Icons_Svg_Assets/Icons8.md", icons)
	index := v.build()
	site := siteNamed(t, index, "Icons8")
	if omStr(site, "subcategory") != "Icons & SVG Assets" {
		t.Fatalf("subcategory: %q", omStr(site, "subcategory"))
	}
	design := catNamed(t, index, "Design")
	subs := catSubNames(design)
	if !containsName(subs, "Icons & SVG Assets") {
		t.Fatalf("subs: %v", subs)
	}
	if containsName(subs, "Icons_Svg_Assets") {
		t.Fatalf("the raw folder name rode the tree: %v", subs)
	}
}

func TestBuildIndex_FoldedPrivateBeltMatchesFolderForms(t *testing.T) {
	// v0.88.0 — the FOLDED belt: a private name must also match its
	// folder form — a registry-less walk can never let the private
	// node slip through on a spelling difference.
	v := newVaultBuilder(t)
	v.write(t, "_system/directory-config.json", `{"categories": []}`)
	v.write(t, "Design/Assets_Resources/Coolors.md", longNote)
	v.write(t, "Download_Resources/Torrent_Shadow_Libraries/Annas.md", privateNote)
	index := v.build()
	blob := storage.PyDumpsUTF8(index)
	if strings.Contains(blob, "annas") || strings.Contains(blob, "Shadow") {
		t.Fatalf("the private node leaked: %s", blob)
	}
	// The public category rides alone (the mirror law: no ghost).
	var catNames []string
	for _, c := range idxCategories(index) {
		catNames = append(catNames, omStr(c, "name"))
	}
	if len(catNames) != 1 || catNames[0] != "Design" {
		t.Fatalf("cats: %v", catNames)
	}
}

func TestBuildIndex_BanishedHostSkipped(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.banished(t, `[{"host": "coolors.co", "reason": "gone"}]`)
	index := v.build()
	sites, _ := idxCounts(t, index)
	if sites != 0 {
		t.Fatalf("sites: %d", sites)
	}
}

func TestBuildIndex_SharedHostURLBanOnlyBlocksThatURL(t *testing.T) {
	shared := refNote(fmFields(map[string]string{
		"source":      "https://someone.github.io/good/",
		"category":    `"Developer Tools"`,
		"subcategory": `""`,
	}), "# Shared Good\n\n> **TL;DR:** t\n"+refFooter("https://someone.github.io/good/"))
	v := newVaultBuilder(t)
	v.write(t, "Developer Tools/Shared Good.md", shared)
	v.banished(t, `[{"host": "github.io", "url": "https://someone.github.io/bad/"}]`)
	index := v.build()
	sites, _ := idxCounts(t, index)
	if sites != 1 {
		t.Fatalf("sites: %d", sites)
	}
}

func TestBuildIndex_DuplicateIDsCollapseAndSlugSuffix(t *testing.T) {
	a := refNote(fmFields(map[string]string{"source": "https://a.example/"}),
		"# Same Name\n\n> **TL;DR:** t\n"+refFooter("https://a.example/"))
	b := refNote(fmFields(map[string]string{"source": "https://b.example/"}),
		"# Same Name\n\n> **TL;DR:** t\n"+refFooter("https://b.example/"))
	v := newVaultBuilder(t)
	v.write(t, "Developer Tools/A.md", a)
	v.write(t, "Developer Tools/B.md", b)
	index := v.build()
	sites := idxSites(t, index)
	if len(sites) != 2 {
		t.Fatalf("sites: %d", len(sites))
	}
	var slugs []string
	for _, s := range sites {
		slugs = append(slugs, omStr(s, "slug"))
	}
	sortStringsDS(slugs)
	if slugs[0] != "same-name" || slugs[1] != "same-name-2" {
		t.Fatalf("slugs: %v", slugs)
	}
}

func TestBuildIndex_SystemAndSkipFoldersNeverIndexed(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.write(t, "_system/directory-config.json", "{}")
	v.write(t, "_review/Pending.md", shortNote)
	v.write(t, ".trash/Gone.md", privateNote)
	v.write(t, "000 📚 Website Directory.md", "# The directory note")
	index := v.build()
	sites, _ := idxCounts(t, index)
	if sites != 1 {
		t.Fatalf("sites: %d", sites)
	}
}

func TestBuildIndex_OnDiskStragglerCategoryGetsDefaults(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Mystery Bucket/Thing.md", shortNote)
	index := v.build()
	var cats map[string]*storage.OrderedMap
	raw, _ := index.Get("categories")
	items, _ := raw.([]any)
	cats = map[string]*storage.OrderedMap{}
	for _, c := range items {
		m := c.(*storage.OrderedMap)
		cats[omStr(m, "name")] = m
	}
	if cats["Mystery Bucket"] == nil {
		t.Fatalf("cats: %v", catNames(index))
	}
	if omStr(cats["Mystery Bucket"], "icon") != "bookmark" {
		t.Fatalf("icon: %q", omStr(cats["Mystery Bucket"], "icon"))
	}
	if omStr(cats["Mystery Bucket"], "visibility") != "public" {
		t.Fatalf("visibility: %q", omStr(cats["Mystery Bucket"], "visibility"))
	}
}

func catNames(index *storage.OrderedMap) []string {
	var out []string
	for _, c := range idxCategories(index) {
		out = append(out, omStr(c, "name"))
	}
	return out
}

func TestBuildIndex_SafeFilenameFoldersMapBackToTaxonomyNames(t *testing.T) {
	// v0.72.14 — the folder-name law: the pipeline writes notes into
	// safe_filename(category) folders; the walk translates them back
	// so SPEC §7.1 holds and the UI's counts / name_fa / grouping
	// survive.
	fig := refNote(fmFields(map[string]string{"source": "https://figma.com/"}),
		"# Figma\n\n> **TL;DR:** t\n"+refFooter("https://figma.com/"))
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets_Resources/Coolors.md", longNote)
	v.write(t, "Design/UI_UX_Product_Design/Figma.md", fig)
	index := v.build()
	coolors := siteNamed(t, index, "Coolors")
	if omStr(coolors, "category") != "Design" {
		t.Fatalf("category: %q", omStr(coolors, "category"))
	}
	if omStr(coolors, "subcategory") != "Assets & Resources" {
		t.Fatalf("subcategory: %q", omStr(coolors, "subcategory"))
	}
	figma := siteNamed(t, index, "Figma")
	if omStr(figma, "subcategory") != "UI/UX & Product Design" {
		t.Fatalf("subcategory: %q", omStr(figma, "subcategory"))
	}
	names := catNames(index)
	if containsName(names, "Assets_Resources") || containsName(names, "UI_UX_Product_Design") {
		t.Fatalf("raw folder names rode the tree: %v", names)
	}
	design := catNamed(t, index, "Design")
	if !containsName(catSubNames(design), "UI/UX & Product Design") {
		t.Fatalf("subs: %v", catSubNames(design))
	}
}

func TestBuildIndex_FullTagAliasesRideThePayload(t *testing.T) {
	// v0.72.14 — the Farsi search corpus law: the FULL alias map rides
	// the index, not just the popular top-12.
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets_Resources/Coolors.md", longNote)
	index := v.build()
	raw, _ := index.Get("tag_aliases")
	aliases, _ := raw.(*storage.OrderedMap)
	fa, _ := aliases.Get("color-tool")
	if fa != "ابزار رنگ" {
		t.Fatalf("color-tool alias: %v", fa)
	}
	if aliases.Len() < 10 {
		t.Fatalf("aliases: %d", aliases.Len())
	}
}

func TestBuildIndex_TopLevelCategoryNamedWithSlashNeverSplitsIntoPhantom(t *testing.T) {
	// v0.72.14 — the phantom-category bug: a top-level category NAMED
	// 'Random / Curiosities' once split into a phantom 'Random '
	// straggler. (cat, sub) tuples kill it.
	swap := refNote(fmFields(map[string]string{"source": "https://www.window-swap.com/"}),
		"# Window Swap\n\n> **TL;DR:** t\n"+refFooter("https://www.window-swap.com/"))
	v := newVaultBuilder(t)
	v.write(t, "Random_Curiosities/Window_Swap.md", swap)
	index := v.build()
	names := catNames(index)
	if !containsName(names, "Random / Curiosities") {
		t.Fatalf("cats: %v", names)
	}
	if containsName(names, "Random") || containsName(names, "Random ") {
		t.Fatalf("the phantom straggler rode: %v", names)
	}
	site := siteNamed(t, index, "Window Swap")
	if omStr(site, "category") != "Random / Curiosities" {
		t.Fatalf("category: %q", omStr(site, "category"))
	}
	if omStr(site, "subcategory") != "" {
		t.Fatalf("subcategory: %q", omStr(site, "subcategory"))
	}
	// v0.88.0 — the mirror law: only the on-disk category publishes
	// (1, not the old config-order 14) — and no 15th phantom.
	if len(idxCategories(index)) != 1 {
		t.Fatalf("cats: %v", names)
	}
}

// --- TestBanishWall -------------------------------------------------------

func TestBanishWall_SharedHostPlatformsNeverWallTheWholeHost(t *testing.T) {
	wall := banishedHostWall([]BanishEntry{
		{Host: "spam.example"},
		{Host: "someone.github.io"}, // shared platform → URL-level
		{Host: "vercel.app"},
		{Host: "realbad.org"},
	}, []string{"github.io", "vercel.app"})
	if len(wall) != 2 || wall[0] != "realbad.org" || wall[1] != "spam.example" {
		t.Fatalf("wall: %v", wall)
	}
}

func TestBanishWall_BuildIndexCarriesTheWall(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.banished(t, `[{"host": "spam.example", "reason": "r", "banished_at": "x"}]`)
	index := v.build()
	raw, _ := index.Get("banished_hosts")
	wall, ok := raw.([]string)
	if !ok || len(wall) != 1 || wall[0] != "spam.example" {
		t.Fatalf("wall: %v (%T)", raw, raw)
	}
}

func TestBanishWall_BanishedHostSitesLeaveTheIndex(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.banished(t, `[{"host": "coolors.co", "reason": "r", "banished_at": "x"}]`)
	index := v.build()
	sites, _ := idxCounts(t, index)
	if sites != 0 {
		t.Fatalf("sites: %d", sites)
	}
}
