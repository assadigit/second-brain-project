// commands_test.go — T-49: the ported selections of the reference's
// TestCommandAppliers + TestCommandCycle (test_directorysync.py) — the
// idempotent vault appliers (trash-not-delete, the move collision
// suffix, the full-body edit, the banish upsert), the
// golden-rule-5 order proof (publish BEFORE ack), the down-Worker
// law, + the local-bare-repos mirror-push battery (the T-48
// technique: the vault as a git repo with a local bare remote).
package directorysync

import (
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
)

func siteIDOf(t *testing.T, url string) string {
	t.Helper()
	return sha256Sum16(links.NormalizeWebsiteURL(url))
}

var coolorsID string

func init() {
	coolorsID = sha256Sum16(links.NormalizeWebsiteURL("https://coolors.co/"))
}

// cmd builds one command dict (the JSON shape the Worker serves).
func cmd(id, ctype, siteID string, payload map[string]any) *storage.OrderedMap {
	c := storage.NewOrderedMap()
	c.Set("command_id", id)
	c.Set("type", ctype)
	c.Set("site_id", siteID)
	if payload == nil {
		payload = map[string]any{}
	}
	c.Set("payload", payload)
	return c
}

func cmdList(cs ...*storage.OrderedMap) []any {
	out := make([]any, len(cs))
	for i, c := range cs {
		out[i] = c
	}
	return out
}

func coolorsPath(v *vaultBuilder) string {
	return filepath.Join(v.dir, "Design", "Assets & Resources", "Coolors.md")
}

func readTextAt(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestCommandAppliers_DeleteMovesToTrashNeverPermanent(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	var logs []string
	results := ApplyAdminCommands(v.dir, cmdList(cmd("c1", "delete", coolorsID, nil)),
		func(msg string) { logs = append(logs, msg) })
	if len(results) != 1 || results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	if _, err := os.Stat(coolorsPath(v)); !os.IsNotExist(err) {
		t.Fatal("the note survived the delete")
	}
	trashed := listDirNames(t, filepath.Join(v.dir, ".trash"))
	if len(trashed) != 1 || !strings.Contains(trashed[0], "Coolors") {
		t.Fatalf("trashed: %v", trashed)
	}
	found := false
	for _, l := range logs {
		if strings.Contains(l, "trash") {
			found = true
		}
	}
	if !found {
		t.Fatalf("logs: %v", logs)
	}
}

func listDirNames(t *testing.T, dir string) []string {
	t.Helper()
	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatal(err)
	}
	var out []string
	for _, e := range entries {
		out = append(out, e.Name())
	}
	return out
}

func TestCommandAppliers_DeleteIsIdempotentAGoneNoteIsAppliedNotFailed(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	c := cmd("c1", "delete", coolorsID, nil)
	first := ApplyAdminCommands(v.dir, cmdList(c), nil)
	second := ApplyAdminCommands(v.dir, cmdList(c), nil)
	if first[0].Status != "applied" || second[0].Status != "applied" {
		t.Fatalf("first: %+v second: %+v (SPEC §10.2: already gone → applied)", first, second)
	}
}

func TestCommandAppliers_BanishTrashesAndWritesBanishedJSON(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	c := storage.NewOrderedMap()
	c.Set("command_id", "c2")
	c.Set("type", "banish")
	c.Set("site_id", coolorsID)
	c.Set("host", "coolors.co")
	c.Set("payload", map[string]any{"reason": "gone paid"})
	results := ApplyAdminCommands(v.dir, cmdList(c), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	entries := LoadBanishList(v.dir)
	if len(entries) != 1 {
		t.Fatalf("entries: %+v", entries)
	}
	if entries[0].Host != "coolors.co" || entries[0].Reason != "gone paid" {
		t.Fatalf("entry: %+v", entries[0])
	}
	if _, err := os.Stat(coolorsPath(v)); !os.IsNotExist(err) {
		t.Fatal("the note survived the banish")
	}
}

func TestCommandAppliers_BanishThenTheIntakeWallBlocksTheHost(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	c := storage.NewOrderedMap()
	c.Set("command_id", "c2")
	c.Set("type", "banish")
	c.Set("site_id", coolorsID)
	c.Set("host", "coolors.co")
	c.Set("payload", map[string]any{})
	ApplyAdminCommands(v.dir, cmdList(c), nil)
	banished := LoadBanishList(v.dir)
	if !IsBanished("https://coolors.co/", banished, DefaultSharedHosts) {
		t.Fatal("the host wall")
	}
	if !IsBanished("https://www.coolors.co/app", banished, DefaultSharedHosts) {
		t.Fatal("the www subdomain wall")
	}
}

func TestCommandAppliers_MoveRelocatesTheNoteAndUpdatesFrontmatter(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(cmd("c3", "move", coolorsID,
		map[string]any{"category": "Developer Tools", "subcategory": "Command Line"})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	moved := filepath.Join(v.dir, "Developer_Tools", "Command_Line", "Coolors.md")
	if _, err := os.Stat(moved); err != nil {
		t.Fatal("the move did not land")
	}
	text := readTextAt(t, moved)
	if !strings.Contains(text, `category: "Developer Tools"`) {
		t.Fatalf("category frontmatter: %s", text[:200])
	}
	if !strings.Contains(text, `subcategory: "Command Line"`) {
		t.Fatalf("subcategory frontmatter: %s", text[:200])
	}
	if _, err := os.Stat(coolorsPath(v)); !os.IsNotExist(err) {
		t.Fatal("the original survived the move")
	}
}

func TestCommandAppliers_MoveNameCollisionGetsANumericSuffix(t *testing.T) {
	// Two notes with the SAME filename, different URLs — both moved
	// into one target folder: the second lands as 'Coolors 2.md'.
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.write(t, "OtherCat/Coolors.md", strings.ReplaceAll(longNote,
		"https://coolors.co/", "https://twin.example/"))
	ApplyAdminCommands(v.dir, cmdList(cmd("c4", "move", coolorsID,
		map[string]any{"category": "Fresh"})), nil)
	ApplyAdminCommands(v.dir, cmdList(cmd("c5", "move", siteIDOf(t, "https://twin.example/"),
		map[string]any{"category": "Fresh"})), nil)
	var names []string
	for _, n := range listDirNames(t, filepath.Join(v.dir, "Fresh")) {
		if strings.HasSuffix(n, ".md") {
			names = append(names, n)
		}
	}
	sort.Strings(names)
	if len(names) != 2 || names[0] != "Coolors 2.md" || names[1] != "Coolors.md" {
		t.Fatalf("names: %v", names)
	}
}

func TestCommandAppliers_EditUpdatesH1TLDRTagsAndFeatured(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(cmd("c6", "edit", coolorsID,
		map[string]any{
			"name":     "Coolors Two",
			"summary":  "A newer summary line.",
			"tags":     []any{"color", "palette", "new"},
			"featured": true,
		})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	text := readTextAt(t, coolorsPath(v))
	if !strings.Contains(text, "# Coolors Two") {
		t.Fatal("the H1 did not update")
	}
	if !strings.Contains(text, "> **TL;DR:** A newer summary line.") {
		t.Fatal("the TL;DR did not update")
	}
	if !strings.Contains(text, "tags: [color, palette, new]") {
		t.Fatal("the tags did not update")
	}
	if !strings.Contains(text, "featured: true") {
		t.Fatal("featured did not update")
	}
	// Idempotent: applying the same edit again writes the same bytes.
	ApplyAdminCommands(v.dir, cmdList(cmd("c6", "edit", coolorsID,
		map[string]any{
			"name":     "Coolors Two",
			"summary":  "A newer summary line.",
			"tags":     []any{"color", "palette", "new"},
			"featured": true,
		})), nil)
	if readTextAt(t, coolorsPath(v)) != text {
		t.Fatal("the second edit wrote different bytes")
	}
}

func TestCommandAppliers_EditWithCategoryMovesTheNote(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(cmd("c7", "edit", coolorsID,
		map[string]any{"category": "Dev Tools"})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	// A PARTIAL edit: the category changes, the subcategory is kept.
	if _, err := os.Stat(filepath.Join(v.dir, "Dev_Tools", "Assets_Resources",
		"Coolors.md")); err != nil {
		t.Fatal("the partial-category move did not land")
	}
}

func TestCommandAppliers_FeatureAndUnfeatureFlipTheFrontmatterFlag(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	ApplyAdminCommands(v.dir, cmdList(cmd("c8", "feature", coolorsID, nil)), nil)
	if !strings.Contains(readTextAt(t, coolorsPath(v)), "featured: true") {
		t.Fatal("feature did not flip")
	}
	ApplyAdminCommands(v.dir, cmdList(cmd("c9", "unfeature", coolorsID, nil)), nil)
	if !strings.Contains(readTextAt(t, coolorsPath(v)), "featured: false") {
		t.Fatal("unfeature did not flip")
	}
}

func TestCommandAppliers_UnbanRemovesTheHostIdempotently(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	v.banished(t, `[{"host": "spam.example", "reason": "spam", "banished_at": "x"},
{"host": "coolors.co", "reason": "test", "banished_at": "x"}]`)
	c := storage.NewOrderedMap()
	c.Set("command_id", "c10")
	c.Set("type", "unban")
	c.Set("site_id", nil)
	c.Set("host", "coolors.co")
	c.Set("payload", map[string]any{"host": "coolors.co"})
	ApplyAdminCommands(v.dir, cmdList(c), nil)
	entries := LoadBanishList(v.dir)
	var hosts []string
	for _, e := range entries {
		hosts = append(hosts, e.Host)
	}
	if len(hosts) != 1 || hosts[0] != "spam.example" {
		t.Fatalf("hosts: %v", hosts)
	}
	// Un-banning an absent host is a quiet applied (idempotent).
	results := ApplyAdminCommands(v.dir, cmdList(cmd("c11", "unban", "",
		map[string]any{"host": "ghost.example"})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
}

func TestCommandAppliers_OneFailureNeverStopsTheQueue(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(
		cmd("bad", "explode", coolorsID, nil),
		cmd("good", "feature", coolorsID, nil)), nil)
	if results[0].Status != "failed" {
		t.Fatalf("results[0]: %+v", results[0])
	}
	if !strings.Contains(results[0].Error, "unknown command type") {
		t.Fatalf("error: %q", results[0].Error)
	}
	if results[1].Status != "applied" {
		t.Fatalf("results[1]: %+v (SPEC §10.2: one failure → continue)", results[1])
	}
}

func TestCommandAppliers_UnknownSiteIsAppliedNotFailed(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(cmd("c12", "feature", "ffff", nil)), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v (idempotent by law)", results)
	}
}

// ── v0.39.0 — Phase 4: the needs-review flag (SPEC §9 P1) ──────────

func TestCommandAppliers_ReviewCommandFlagsAndUnflagsTheNote(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(cmd("r1", "review", coolorsID,
		map[string]any{"needs_review": true})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	text := readTextAt(t, coolorsPath(v))
	if !strings.Contains(text, "needs_review: true") {
		t.Fatal("needs_review did not flip")
	}
	if !strings.Contains(text, "# Coolors") {
		t.Fatal("the review command touched the body")
	}
	// The flag rides the next index (the site's public badge).
	index := v.build()
	site := siteNamed(t, index, "Coolors")
	if omAny(site, "needs_review") != true {
		t.Fatalf("needs_review: %v", omAny(site, "needs_review"))
	}
	// Un-flag → false on both layers.
	ApplyAdminCommands(v.dir, cmdList(cmd("r2", "review", coolorsID,
		map[string]any{"needs_review": false})), nil)
	if !strings.Contains(readTextAt(t, coolorsPath(v)), "needs_review: false") {
		t.Fatal("needs_review did not unflip")
	}
	index2 := v.build()
	site2 := siteNamed(t, index2, "Coolors")
	if omAny(site2, "needs_review") != false {
		t.Fatalf("needs_review: %v", omAny(site2, "needs_review"))
	}
}

func TestCommandAppliers_ReviewCommandIsIdempotent(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	c := cmd("r3", "review", coolorsID, map[string]any{"needs_review": true})
	ApplyAdminCommands(v.dir, cmdList(c), nil)
	first := readTextAt(t, coolorsPath(v))
	ApplyAdminCommands(v.dir, cmdList(c), nil)
	if readTextAt(t, coolorsPath(v)) != first {
		t.Fatal("the second review wrote different bytes")
	}
	if !strings.Contains(first, "needs_review: true") {
		t.Fatal("needs_review missing")
	}
}

func TestCommandAppliers_IndexDefaultsNeedsReviewToFalseWhenAbsent(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	index := v.build()
	site := siteNamed(t, index, "Coolors")
	if omAny(site, "needs_review") != false {
		t.Fatalf("needs_review: %v", omAny(site, "needs_review"))
	}
}

// ── v0.39.0 — Phase 4: the full-body edit (SPEC §9 P1) ─────────────

func TestCommandAppliers_EditBodyReplacesTheWholeBodyKeepsFrontmatter(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	markdown := "# Coolors Prime\n\n" +
		"> **TL;DR:** rewritten from the admin area.\n\n" +
		"## What it does\n- one new fact\n- another\n\n" +
		"## Watch out\n- nothing"
	results := ApplyAdminCommands(v.dir, cmdList(cmd("b1", "edit", coolorsID,
		map[string]any{"body": markdown})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	text := readTextAt(t, coolorsPath(v))
	if !strings.HasPrefix(text, "---") {
		t.Fatal("the frontmatter survived a body edit — it must (but start with ---)")
	}
	if !strings.Contains(text, "source: https://coolors.co/") {
		t.Fatal("the source frontmatter did not survive")
	}
	if !strings.Contains(text, "# Coolors Prime") {
		t.Fatal("the new H1 did not land")
	}
	if !strings.Contains(text, "> **TL;DR:** rewritten from the admin area.") {
		t.Fatal("the new TL;DR did not land")
	}
	if !strings.Contains(text, "## What it does\n- one new fact") {
		t.Fatal("the new section did not land")
	}
	if strings.Contains(text, "## Similar tools") {
		t.Fatal("the OLD body is gone wholesale")
	}
	// The next index parses the NEW body.
	index := v.build()
	site := siteNamed(t, index, "Coolors Prime")
	if omStr(site, "summary") != "rewritten from the admin area." {
		t.Fatalf("summary: %q", omStr(site, "summary"))
	}
	wd, _ := site.Get("what_it_does")
	if len(wd.([]any)) != 2 {
		t.Fatalf("what_it_does: %v", wd)
	}
	wo, _ := site.Get("watch_out")
	if len(wo.([]any)) != 1 || wo.([]any)[0] != "nothing" {
		t.Fatalf("watch_out: %v", wo)
	}
}

func TestCommandAppliers_EditBodyNormalizesCRLFAndIsIdempotent(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	c := cmd("b2", "edit", coolorsID, map[string]any{"body": "# X\r\n\r\n> **TL;DR:** y\r\n"})
	ApplyAdminCommands(v.dir, cmdList(c), nil)
	first := readTextAt(t, coolorsPath(v))
	if strings.Contains(first, "\r") {
		t.Fatal("the CRLF survived")
	}
	ApplyAdminCommands(v.dir, cmdList(c), nil)
	if readTextAt(t, coolorsPath(v)) != first {
		t.Fatal("the second body edit wrote different bytes")
	}
}

func TestCommandAppliers_EditBodyAlongsideFieldEditsInOneCommand(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	results := ApplyAdminCommands(v.dir, cmdList(cmd("b3", "edit", coolorsID,
		map[string]any{
			"featured": true,
			"body":     "# Zed\n\n> **TL;DR:** combined",
		})), nil)
	if results[0].Status != "applied" {
		t.Fatalf("results: %+v", results)
	}
	text := readTextAt(t, coolorsPath(v))
	if !strings.Contains(text, "featured: true") {
		t.Fatal("featured did not land")
	}
	if !strings.Contains(text, "# Zed") {
		t.Fatal("the combined body did not land")
	}
}

// --- TestCommandCycle -----------------------------------------------------

func cycleConfig(url, vault string) map[string]any {
	return map[string]any{
		"website_vault_path":   vault,
		"directory_worker_url": url,
		"directory_sync_token": stubToken,
	}
}

func TestCommandCycle_FullCyclePullApplyPublishAckInOrder(t *testing.T) {
	v := newVaultBuilder(t)
	server := newStubWorker(t, "")
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	happy := strings.ReplaceAll(
		strings.ReplaceAll(longNote, "https://coolors.co/", "https://happyhues.example/"),
		"Coolors", "Happy Hues")
	v.write(t, "Design/Assets & Resources/Happy Hues.md", happy)
	c2 := storage.NewOrderedMap()
	c2.Set("command_id", "cmd-2")
	c2.Set("type", "delete")
	c2.Set("site_id", siteIDOf(t, "https://happyhues.example/"))
	c2.Set("slug", "happy-hues")
	c2.Set("host", "happyhues.example")
	c2.Set("payload", map[string]any{})
	c2.Set("created_at", "t")
	c1 := storage.NewOrderedMap()
	c1.Set("command_id", "cmd-1")
	c1.Set("type", "feature")
	c1.Set("site_id", coolorsID)
	c1.Set("slug", "coolors")
	c1.Set("host", "coolors.co")
	c1.Set("payload", map[string]any{})
	c1.Set("created_at", "t")
	server.setCommands(cmdList(c1, c2))
	summary := RunCommandCycle(cycleConfig(server.url(), v.dir), nil)
	if int(pyIntOf(mustCfgGet(summary, "pulled"))) != 2 ||
		int(pyIntOf(mustCfgGet(summary, "applied"))) != 2 ||
		int(pyIntOf(mustCfgGet(summary, "failed"))) != 0 ||
		int(pyIntOf(mustCfgGet(summary, "acked"))) != 2 {
		t.Fatalf("summary: %v", summary)
	}
	if !pyTruthyDS(mustCfgGet(summary, "published")) {
		t.Fatal("published")
	}
	// The vault: one featured, one trashed.
	if !strings.Contains(readTextAt(t, coolorsPath(v)), "featured: true") {
		t.Fatal("the feature did not land")
	}
	if len(listDirNames(t, filepath.Join(v.dir, ".trash"))) == 0 {
		t.Fatal("the trash is empty")
	}
	// The publish reflects the post-apply vault (1 site left).
	received := server.receivedIndexes()
	if len(received) != 1 {
		t.Fatalf("received: %d", len(received))
	}
	sites, _ := received[0].Get("sites")
	if len(sites.([]any)) != 1 {
		t.Fatalf("sites: %v", sites)
	}
	// THE ORDER LAW: publish BEFORE ack (golden rule 5: d before e —
	// the stored index and the overlay agree before the ack).
	events := server.eventLog()
	syncIdx, ackIdx := -1, -1
	for i, e := range events {
		if e.kind == "sync" && syncIdx == -1 {
			syncIdx = i
		}
		if e.kind == "ack" && ackIdx == -1 {
			ackIdx = i
		}
	}
	if syncIdx == -1 || ackIdx == -1 || syncIdx >= ackIdx {
		t.Fatalf("the order law broke: events=%v", events)
	}
	var acks [][2]string
	for _, e := range events {
		if e.kind == "ack" {
			acks = append(acks, [2]string{e.command, e.status})
		}
	}
	if len(acks) != 2 || acks[0][0] != "cmd-1" || acks[0][1] != "applied" ||
		acks[1][0] != "cmd-2" || acks[1][1] != "applied" {
		t.Fatalf("acks: %v", acks)
	}
}

func TestCommandCycle_WorkerDownCommandsWaitNeverAcked(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	summary := RunCommandCycle(cycleConfig("http://127.0.0.1:1", v.dir), nil)
	if int(pyIntOf(mustCfgGet(summary, "pulled"))) != 0 ||
		int(pyIntOf(mustCfgGet(summary, "acked"))) != 0 {
		t.Fatalf("summary: %v", summary)
	}
	if pyTruthyDS(mustCfgGet(summary, "published")) {
		t.Fatal("a down Worker is a quiet failure, never a false publish")
	}
}

func TestCommandCycle_PullWithWrongTokenIsNil(t *testing.T) {
	server := newStubWorker(t, "right-token")
	if FetchPendingCommands(server.url(), "wrong-token", 0, nil) != nil {
		t.Fatal("the wrong token pulled")
	}
	if FetchPendingCommands(server.url(), "right-token", 0, nil) == nil {
		t.Fatal("the right token did not pull")
	}
}

func TestCommandCycle_AckNeverRaisesOnADeadWorker(t *testing.T) {
	if AckCommand("http://127.0.0.1:1", "tok", "x", "applied", "", 0, nil) {
		t.Fatal("the dead-worker ack answered true")
	}
}

func TestCommandCycle_PushMirrorRepoOnAPlainFolderIsAQuietNoop(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	if PushMirrorRepo(v.dir, nil) {
		t.Fatal("a plain folder answered true")
	}
}

// --- the local-bare-repos mirror-push battery (the T-48 technique) --------

// mirrorRepo builds a vault repo + a local bare remote (the golden
// fixture's twin: identity, tracking, frozen git dates via the
// caller's t.Setenv).
func mirrorRepo(t *testing.T) (vault, bare string) {
	t.Helper()
	tmp := t.TempDir()
	vault = filepath.Join(tmp, "vault")
	bare = filepath.Join(tmp, "bare.git")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	runGit := func(dir string, args ...string) {
		t.Helper()
		out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	runGit(tmp, "init", "--bare", "-b", "main", "bare.git")
	runGit(vault, "init", "-b", "main")
	runGit(vault, "config", "user.name", "Golden Mirror")
	runGit(vault, "config", "user.email", "golden@example.com")
	runGit(vault, "remote", "add", "origin", bare)
	runGit(vault, "config", "branch.main.remote", "origin")
	runGit(vault, "config", "branch.main.merge", "refs/heads/main")
	if err := os.MkdirAll(filepath.Join(vault, "Design"), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(vault, "Design", "Note.md"),
		[]byte(longNote), 0o644); err != nil {
		t.Fatal(err)
	}
	return vault, bare
}

func gitOut(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := exec.Command("git", append([]string{"-C", dir}, args...)...).Output()
	if err != nil {
		t.Fatalf("git %v: %v", args, err)
	}
	return strings.TrimSpace(string(out))
}

func TestMirrorPush_TheFullLadderAgainstALocalBareRemote(t *testing.T) {
	freezeClock(t)
	freezeGitDates(t)
	vault, bare := mirrorRepo(t)
	var logs []string
	if !PushMirrorRepo(vault, func(msg string) { logs = append(logs, msg) }) {
		t.Fatalf("logs: %v", logs)
	}
	// The bare remote carries the committed tree on main.
	tree := gitOut(t, bare, "ls-tree", "-r", "--name-only", "main")
	if !strings.Contains(tree, "Design/Note.md") {
		t.Fatalf("remote tree: %s", tree)
	}
	// The frozen clock + frozen git dates pin the commit message.
	logLine := gitOut(t, bare, "log", "--format=%s", "main")
	if logLine != "admin commands applied: 2026-01-01T12:34:00+00:00" {
		t.Fatalf("commit message: %q", logLine)
	}
	// The clean-tree leg: nothing to push answers True quietly.
	if !PushMirrorRepo(vault, nil) {
		t.Fatal("the clean tree answered false")
	}
}

func TestMirrorPush_NoRemoteIsAQuietSkip(t *testing.T) {
	freezeClock(t)
	freezeGitDates(t)
	tmp := t.TempDir()
	vault := filepath.Join(tmp, "vault")
	if err := os.MkdirAll(vault, 0o755); err != nil {
		t.Fatal(err)
	}
	gitOut(t, vault, "init", "-b", "main")
	gitOut(t, vault, "config", "user.name", "Golden Mirror")
	gitOut(t, vault, "config", "user.email", "golden@example.com")
	if err := os.WriteFile(filepath.Join(vault, "Note.md"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	var logs []string
	if PushMirrorRepo(vault, func(msg string) { logs = append(logs, msg) }) {
		t.Fatal("a repo with no remote answered true")
	}
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "no git remote configured") {
		t.Fatalf("logs: %v", logs)
	}
}
