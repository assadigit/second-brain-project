// golden_flows_test.go — T-40: the generated-golden differential for
// the vault scans, the table writers, and THE BANISHMENT FLOWS (the
// same corpus as golden_test.go's pure logic, replayed against real
// temp state DBs — the reference generator's own pattern, frozen
// clock, no per-run facts inside).
package websitepipeline

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/assadigit/second-brain-project/internal/handdelivery"
)

var wpgoldenRootRE = regexp.MustCompile(`^/tmp/wpgolden-[^/]+`)

// cmpGoldenNorm byte-compares with both sides' tempdir roots folded
// to "<TMP>" (the golden's paths are positional, not per-run).
func cmpGoldenNorm(t *testing.T, what string, got, want any, root string) {
	t.Helper()
	cmpGoldenJSON(t, what, normPaths(normalizeGo(got), root), normPaths(want, ""))
}

// normPaths rewrites every string value under a map/list tree,
// replacing the leading tempdir root (the Go test's or the generator's
// baked-in /tmp/wpgolden-XXXX) with "<TMP>/" — the golden's paths are
// positional facts, not per-run ones, and this makes them comparable.
func normPaths(v any, root string) any {
	repl := func(s string) string {
		if root != "" && strings.HasPrefix(s, root) {
			return "<TMP>" + s[len(root):]
		}
		// the generator's baked tempdir (/tmp/wpgolden-XXXX)
		return wpgoldenRootRE.ReplaceAllString(s, "<TMP>")
	}

	switch t := v.(type) {
	case map[string]any:
		out := map[string]any{}
		for k, val := range t {
			if s, ok := val.(string); ok {
				out[k] = repl(s)
			} else {
				out[k] = normPaths(val, root)
			}
		}
		return out
	case []any:
		out := make([]any, len(t))
		for i, val := range t {
			if s, ok := val.(string); ok {
				out[i] = repl(s)
			} else {
				out[i] = normPaths(val, root)
			}
		}
		return out
	case []map[string]any:
		out := make([]any, len(t))
		for i, val := range t {
			out[i] = normPaths(val, root)
		}
		return out
	}
	return v
}

// copyTree mirrors shutil.copytree (files + dirs, recursively).
func copyTree(t *testing.T, src, dst string) {
	t.Helper()
	err := filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		return os.WriteFile(target, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copyTree %s -> %s: %v", src, dst, err)
	}
}

// --- the fixture vault (the generator's exact construction) ---------------

func buildFixtureVault(t *testing.T) string {
	t.Helper()
	vault := filepath.Join(t.TempDir(), "websites")
	os.MkdirAll(filepath.Join(vault, "_review"), 0o755)
	os.MkdirAll(filepath.Join(vault, "Design", "Assets & Resources"), 0o755)
	gwriteNote(t, vault, "Design/Assets & Resources/proper.md", "https://proper.example/", "[design]", "gitcurator", "full", "")
	gwriteNote(t, vault, "_review/failed-placeholder.md", "https://failed.example/", "[design]", "gitcurator", "failed",
		"> [!warning] Needs review — failed\n> Fetch failed: HTTP 404 — not found\n")
	gwriteNote(t, vault, "_review/low-confidence.md", "https://low.example/", "[design]", "gitcurator", "partial",
		"> [!warning] Needs review — partial\n> classification confidence was low\n")
	gwriteNote(t, vault, "_review/hand-written.md", "https://hand.example/", "[design]", "human", "full", "")
	gwriteNote(t, vault, "Trash/moved.md", "https://moved.example/", "[design]", "gitcurator", "full", "")
	gwriteNote(t, vault, "Design/Assets & Resources/marked.md", "https://marked.example/", "[delete, design]", "gitcurator", "full", "")
	gwriteNote(t, vault, "Design/Assets & Resources/body-marked.md", "https://bodymarked.example/", "[design]", "gitcurator", "full",
		"Some text.\n\n#auto-delete\n")
	os.MkdirAll(filepath.Join(vault, ".trash", "banished"), 0o755)
	gwriteNote(t, vault, ".trash/banished/already.md", "https://already.example/", "[delete]", "gitcurator", "full", "")
	gwriteNote(t, vault, "_inbox/platform-table.md", "https://inboxed.example/", "[delete]", "gitcurator", "full", "")
	return vault
}

func TestGoldenScanBanishedNotes(t *testing.T) {
	g := sectionList(t, "scan_banished_notes")
	withFrozenClock(t, func() {
		vault := buildFixtureVault(t)
		got := ScanBanishedNotes(vault, nil, nil)
		out := make([]map[string]any, 0, len(got))
		for _, it := range got {
			out = append(out, map[string]any{
				"url": it.URL, "path": it.Path, "marker": it.Marker,
				"door": it.Door, "app_owned": it.AppOwned,
			})
		}
		cmpGoldenNorm(t, "scan_banished_notes", out, g, filepath.Dir(vault))
	})
}

func TestGoldenScanPendingBanishments(t *testing.T) {
	g := section(t, "scan_pending_banishments")
	withFrozenClock(t, func() {
		vault := buildFixtureVault(t)
		rep := ScanPendingBanishments(vault, nil, nil)
		items := make([]map[string]any, 0, len(rep.Items))
		for _, it := range rep.Items {
			items = append(items, map[string]any{
				"url": it.URL, "canonical": it.Canonical, "title": it.Title,
				"marker": it.Marker, "door": it.Door, "path": it.Path,
			})
		}
		cmpGoldenNorm(t, "scan_pending_banishments.items", items, glist(t, g, "items"), filepath.Dir(vault))
		if rep.KeptHandwritten != gint(t, g, "kept_handwritten") {
			t.Errorf("kept_handwritten: got %d", rep.KeptHandwritten)
		}
	})
}

func TestGoldenScanReviewNotesAndBacklog(t *testing.T) {
	gn := sectionList(t, "scan_review_notes")
	gb := sectionList(t, "scan_review_backlog")
	withFrozenClock(t, func() {
		vault := buildFixtureVault(t)
		notes := ScanReviewNotes(vault, nil)
		outN := make([]map[string]any, 0, len(notes))
		for _, it := range notes {
			outN = append(outN, map[string]any{
				"url": it.URL, "path": it.Path, "fetch_status": it.FetchStatus,
			})
		}
		cmpGoldenNorm(t, "scan_review_notes", outN, gn, filepath.Dir(vault))
		backlog := ScanReviewBacklog(vault, nil, nil)
		outB := make([]map[string]any, 0, len(backlog))
		for _, it := range backlog {
			outB = append(outB, map[string]any{"url": it.URL, "path": it.Path})
		}
		cmpGoldenNorm(t, "scan_review_backlog", outB, gb, filepath.Dir(vault))
	})
}

func TestGoldenScanDecommissionTable(t *testing.T) {
	g := section(t, "scan_decommission_table")
	withFrozenClock(t, func() {
		vault := buildFixtureVault(t)
		gtable(t, filepath.Join(vault, "_review", "DECOMMISSIONED.md"), masterTableRows())
		got := ScanDecommissionTable(vault)
		cmpGoldenJSON(t, "scan_decommission_table", map[string]string(got), g)
	})
}

func masterTableRows() []string {
	return []string{
		"| - | 2026-10-01 | https://wait.example/w | wait.example | waiting | unreviewed | |",
		"| - | 2026-10-01 | https://low.example/ | low.example | review | unreviewed | |",
		"| - | 2026-10-01 | https://stored.example/s | stored.example | waiting | unreviewed | |",
		"| - | 2026-10-01 | https://failed.example/ | failed.example | waiting | unreviewed | |",
		"| 🖐 | 2026-10-01 | https://handrow.example/h | handrow.example | waiting | unreviewed | the wall |",
		"| 🪦 | 2026-10-01 | https://deadrow.example/d | deadrow.example | waiting |  | |",
		"| - | 2026-10-01 | https://settled.example/t | settled.example | waiting | - | |",
	}
}

func TestGoldenMasterWaitingRowsAndHandQueue(t *testing.T) {
	gw := sectionList(t, "scan_master_waiting_rows")
	gh := sectionList(t, "scan_master_hand_rows")
	gc := sectionList(t, "collect_delivered")
	gsf := section(t, "suggested_filename")
	gr := loadGolden(t)["readme_text"].(string)
	gq := section(t, "queue_json")
	gt := section(t, "take_hand_delivered")
	gtr := section(t, "take_hand_delivered_redo")
	gth := section(t, "take_hand_delivered_heal")

	withFrozenClock(t, func() {
		tmp := t.TempDir()
		vault := buildFixtureVault(t)
		db, err := NewWebsiteStateDB(filepath.Join(tmp, "cache.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		_ = db.EnqueueRetry("https://wait.example/w", "HTTP 404 — not found")
		_ = db.EnqueueRetry("https://failed.example/", "HTTP 403 — refused")
		storedNote := filepath.Join(vault, "Design", "Assets & Resources", "stored.md")
		_ = db.MarkProcessed("https://stored.example/s", storedNote, "Design", "Assets & Resources", "full")
		_ = db.Dismiss("https://dismissed.example/x", "auto-verdict: dead — HTTP 404")
		if _, err := db.SettleExisting([]string{"https://settled.example/t"}); err != nil {
			t.Fatal(err)
		}
		gtable(t, filepath.Join(vault, "_review", "DECOMMISSIONED.md"), masterTableRows())

		got := ScanMasterWaitingRows(vault, db, nil)
		out := make([]map[string]any, 0, len(got))
		for _, it := range got {
			out = append(out, map[string]any{
				"url": it.URL, "status": it.Status, "notes": it.Notes,
				"path": it.Path, "kind": it.Kind,
			})
		}
		cmpGoldenNorm(t, "scan_master_waiting_rows", out, gw, filepath.Dir(vault))

		// the hand queue: one delivered page for the hand row
		walls := map[string]string{"https://handrow.example/h": "HTTP 403 — refused"}
		handdelivery.EnqueueHandDelivery(vault, []string{"https://handrow.example/h"}, walls, nil, false, "")
		page := filepath.Join(handdelivery.HandDeliveryDir(vault),
			handdelivery.SuggestedFilename("https://handrow.example/h"))
		os.WriteFile(page, []byte("<html><head><title>Hand</title></head><body>delivered</body></html>"), 0o644)

		handRows := ScanMasterHandRows(vault, nil)
		outH := make([]map[string]any, 0, len(handRows))
		for _, it := range handRows {
			outH = append(outH, map[string]any{"url": it.URL, "wall": it.Wall, "status": it.Status})
		}
		cmpGoldenJSON(t, "scan_master_hand_rows", outH, gh)

		delivered := handdelivery.CollectDelivered(vault)
		outC := make([]map[string]any, 0, len(delivered))
		for _, it := range delivered {
			outC = append(outC, map[string]any{
				"url": it.URL, "wall": it.Wall, "suggested": it.Suggested, "door": it.Door,
			})
		}
		cmpGoldenJSON(t, "collect_delivered", outC, gc)

		for u, want := range gsf {
			if got := handdelivery.SuggestedFilename(u); got != want.(string) {
				t.Errorf("suggested_filename(%q): got %q want %q", u, got, want.(string))
			}
		}
		queue := handdelivery.ReadQueue(vault)
		if got := handdelivery.ReadmeText(queue); got != gr {
			t.Errorf("readme_text diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gr)
		}
		qj := map[string]any{}
		if err := json.Unmarshal([]byte(readFileT(t, handdelivery.QueuePath(vault))), &qj); err != nil {
			t.Fatal(err)
		}
		cmpGoldenJSON(t, "queue_json", qj, gq)

		// the take: consume the delivered page
		taken := handdelivery.TakeHandDelivered(vault, "https://handrow.example/h", nil, false)
		if taken == nil {
			t.Fatal("take_hand_delivered: got nil")
		}
		after := map[string]any{}
		json.Unmarshal([]byte(readFileT(t, handdelivery.QueuePath(vault))), &after)
		cmpGoldenJSON(t, "take_hand_delivered", map[string]any{
			"status":       taken.Status,
			"http_status":  derefInt(taken.HTTPStatus),
			"content_type": taken.ContentType,
			"charset":      taken.Charset,
			"reason":       taken.Reason,
			"body":         string(taken.Body),
			"queue_after":  after,
		}, gt)

		// redo-consume: allowConsumed re-reads
		taken2 := handdelivery.TakeHandDelivered(vault, "https://handrow.example/h", nil, true)
		after2 := map[string]any{}
		json.Unmarshal([]byte(readFileT(t, handdelivery.QueuePath(vault))), &after2)
		cmpGoldenJSON(t, "take_hand_delivered_redo", map[string]any{
			"status":      taken2.Status,
			"queue_after": after2,
		}, gtr)

		// the canonical re-key heal: a legacy spelling row
		qraw := readFileT(t, handdelivery.QueuePath(vault))
		healed := strings.Replace(qraw,
			`"https://handrow.example/h": {`, `"https://handrow.example/h/": {`, 1)
		os.WriteFile(handdelivery.QueuePath(vault), []byte(healed), 0o644)
		taken3 := handdelivery.TakeHandDelivered(vault, "https://handrow.example/h", nil, false)
		after3 := map[string]any{}
		json.Unmarshal([]byte(readFileT(t, handdelivery.QueuePath(vault))), &after3)
		var status3 any
		if taken3 != nil {
			status3 = taken3.Status
		}
		cmpGoldenJSON(t, "take_hand_delivered_heal", map[string]any{
			"status":      status3,
			"queue_after": after3,
		}, gth)
	})
}

func derefInt(p *int) any {
	if p == nil {
		return nil
	}
	return float64(*p)
}

func TestGoldenNoteIsProperlyStoredAndFind(t *testing.T) {
	g := sectionList(t, "note_is_properly_stored")
	gf := section(t, "find_note_for_url")
	withFrozenClock(t, func() {
		tmp := t.TempDir()
		vault := buildFixtureVault(t)
		proper := filepath.Join(vault, "Design", "Assets & Resources", "proper.md")
		siteNote := filepath.Join(tmp, "nips", "site.md")
		gnote(t, siteNote, "https://site.example/")
		AddSiteLinksToNote(siteNote, []string{"https://site.example/other"}, nil)

		np := func(prior *ProcessedRow) []any {
			proper, why := NoteIsProperlyStored(prior)
			return []any{proper, why}
		}
		got := []map[string]any{
			{"prior": nil, "proper": false, "why": np(nil)},
			{"prior": map[string]any{
				"url": "https://proper.example/", "note_path": proper, "fetch_status": "full",
			}, "res": np(&ProcessedRow{URL: "https://proper.example/", NotePath: strPtr(proper), FetchStatus: strPtr("full")})},
			{"res": np(&ProcessedRow{URL: "https://proper.example/", NotePath: strPtr(proper), FetchStatus: strPtr("failed")})},
			{"res": np(&ProcessedRow{URL: "https://proper.example/", NotePath: strPtr(filepath.Join(vault, "_review", "failed-placeholder.md")), FetchStatus: strPtr("partial")})},
			{"res": np(&ProcessedRow{URL: "https://proper.example/", NotePath: strPtr(""), FetchStatus: strPtr("full")})},
			{"res": np(&ProcessedRow{URL: "https://proper.example/", NotePath: strPtr(filepath.Join(vault, "gone", "missing.md")), FetchStatus: strPtr("full")})},
			{"res": np(&ProcessedRow{URL: "https://hand.example/", NotePath: strPtr(filepath.Join(vault, "_review", "hand-written.md")), FetchStatus: strPtr("full")})},
			{"res": np(&ProcessedRow{URL: "https://site.example/other", NotePath: strPtr(siteNote), FetchStatus: strPtr("full")})},
			{"res": np(&ProcessedRow{URL: "https://stranger.example/x", NotePath: strPtr(siteNote), FetchStatus: strPtr("full")})},
		}
		cmpGoldenNorm(t, "note_is_properly_stored", got, g, filepath.Dir(vault))

		db, err := NewWebsiteStateDB(filepath.Join(tmp, "find.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer db.Close()
		norm1 := func(p string) string {
			if r, ok := normPaths(p, filepath.Dir(vault)).(string); ok {
				return r
			}
			return p
		}
		normW := func(p string) string {
			if r, ok := normPaths(p, "").(string); ok {
				return r
			}
			return p
		}
		if got, want := norm1(FindNoteForURL(vault, "https://proper.example/", db, nil)), normW(gstr(t, gf, "state_row")); got != want {
			t.Errorf("find_note_for_url.state_row: got %q want %q", got, want)
		}
		if got, want := norm1(FindNoteForURL(vault, "https://marked.example/", nil, nil)), normW(gstr(t, gf, "walk")); got != want {
			t.Errorf("find_note_for_url.walk: got %q want %q", got, want)
		}
		if got, want := norm1(FindNoteForURL(vault, "https://nothing.example/", nil, nil)), normW(gstr(t, gf, "gone")); got != want {
			t.Errorf("find_note_for_url.gone: got %q want %q", got, want)
		}
	})
}

func strPtr(s string) *string { return &s }

// --- the table writers --------------------------------------------------------

func TestGoldenTableWriters(t *testing.T) {
	gf := section(t, "write_candidates_fresh")
	ga := section(t, "write_candidates_append")
	gs := section(t, "write_candidates_status")
	gm := section(t, "mark_urls_dead")
	ge := section(t, "ensure_table_shape")
	gei := loadGolden(t)["ensure_table_shape_idempotent"]

	withFrozenClock(t, func() {
		tmp := t.TempDir()
		v2 := filepath.Join(tmp, "websites2")
		os.MkdirAll(filepath.Join(v2, "_review"), 0o755)
		table := filepath.Join(v2, "_review", "DECOMMISSIONED.md")

		n1 := WriteDecommissionCandidates(v2, []string{
			"https://n1.example/a", "https://n2.example/b", "https://n1.example/a",
		}, "waiting", nil, map[string]string{"https://n1.example/a": "the last error | with pipe"}, "unreviewed")
		if n1 != gint(t, gf, "written") {
			t.Errorf("fresh written: got %d", n1)
		}
		if got := readFileT(t, table); got != gstr(t, gf, "file") {
			t.Errorf("fresh table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, gf, "file"))
		}

		n2 := WriteDecommissionCandidates(v2, []string{
			"https://n1.example/a", "https://n3.example/c",
		}, "waiting", nil, map[string]string{"https://n3.example/c": "note c"}, "unreviewed")
		if n2 != gint(t, ga, "written") {
			t.Errorf("append written: got %d", n2)
		}
		if got := readFileT(t, table); got != gstr(t, ga, "file") {
			t.Errorf("append table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, ga, "file"))
		}

		n3 := WriteDecommissionCandidates(v2, []string{"https://n4.example/d"},
			"retry backlog", nil, map[string]string{"https://n4.example/d": strings.Repeat("x", 200)}, "🪦 auto — dead")
		if n3 != gint(t, gs, "written") {
			t.Errorf("status written: got %d", n3)
		}
		if got := readFileT(t, table); got != gstr(t, gs, "file") {
			t.Errorf("status table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, gs, "file"))
		}

		m1 := MarkURLsDeadInTable(v2, []string{"https://n5.example/e", "https://n1.example/a"}, nil)
		if m1 != gint(t, gm, "marked") {
			t.Errorf("marked: got %d", m1)
		}
		if got := readFileT(t, table); got != gstr(t, gm, "file") {
			t.Errorf("mark_urls_dead table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, gm, "file"))
		}

		v3 := filepath.Join(tmp, "websites3")
		os.MkdirAll(filepath.Join(v3, "_review"), 0o755)
		tpath := filepath.Join(v3, "_review", "DECOMMISSIONED.md")
		os.WriteFile(tpath, []byte(`# header

| # | Date | URL | Domain |
|---|------|-----|--------|
| 💀 | 2026-10-01 | https://trim.example/t | trim.example |
| - | 2026-10-01 | https://trim2.example/t2 | trim2.example |
`), 0o644)
		restored := EnsureTableShape(tpath, nil)
		if restored != gbool(t, ge, "restored") {
			t.Errorf("ensure_table_shape.restored: got %v", restored)
		}
		if got := readFileT(t, tpath); got != gstr(t, ge, "file") {
			t.Errorf("ensure_table_shape table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, ge, "file"))
		}
		restored2 := EnsureTableShape(tpath, nil)
		if restored2 != gei.(bool) {
			t.Errorf("ensure_table_shape.idempotent: got %v", restored2)
		}
	})
}

// --- THE BANISHMENT FLOWS ------------------------------------------------------

func TestGoldenBanishMarkedNotes(t *testing.T) {
	g := section(t, "banish_marked_notes")
	gi := section(t, "banish_marked_notes_idempotent")
	withFrozenClock(t, func() {
		tmp := t.TempDir()
		// the generator's flow: the fixture vault (with its master
		// table) copied to flowA, the proper note re-marked in the copy.
		vault := buildFixtureVault(t)
		gtable(t, filepath.Join(vault, "_review", "DECOMMISSIONED.md"), masterTableRows())
		va := filepath.Join(tmp, "flowA")
		copyTree(t, vault, va)
		dbA, err := NewWebsiteStateDB(filepath.Join(tmp, "flowA.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbA.Close()
		_ = dbA.MarkProcessed("https://proper.example/",
			filepath.Join(va, "Design", "Assets & Resources", "proper.md"),
			"Design", "Assets & Resources", "full")
		_ = dbA.EnqueueRetry("https://marked.example/", "HTTP 403 — refused")

		rep := BanishMarkedNotes(dbA, va, nil, nil)
		cmpGoldenNorm(t, "banish_marked_notes", map[string]any{
			"marked":           float64(rep.Marked),
			"banished":         float64(rep.Banished),
			"notes_moved":      float64(rep.NotesMoved),
			"kept_handwritten": float64(rep.KeptHandwritten),
			"review_swept":     float64(rep.ReviewSwept),
			"rows_written":     float64(rep.RowsWritten),
			"urls":             rep.URLs,
			"table":            readFileT(t, filepath.Join(va, "_review", "DECOMMISSIONED.md")),
			"trash":            listDirT(t, filepath.Join(va, ".trash", "banished")),
			"db":               dumpDB(t, filepath.Join(tmp, "flowA.db")),
		}, g, tmp)

		rep2 := BanishMarkedNotes(dbA, va, nil, nil)
		cmpGoldenJSON(t, "banish_marked_notes_idempotent", map[string]any{
			"marked":   float64(rep2.Marked),
			"banished": float64(rep2.Banished),
			"urls":     rep2.URLs,
		}, gi)
	})
}

func TestGoldenConsumeConfirmed(t *testing.T) {
	g := section(t, "consume_confirmed")
	withFrozenClock(t, func() {
		tmp := t.TempDir()
		vb := filepath.Join(tmp, "flowB")
		os.MkdirAll(filepath.Join(vb, "_review"), 0o755)
		os.MkdirAll(filepath.Join(vb, "Design", "Assets & Resources"), 0o755)
		pb := gwriteNote(t, vb, "Design/Assets & Resources/bnote.md", "https://banish.example/", "[design]", "gitcurator", "full", "")
		dbB, err := NewWebsiteStateDB(filepath.Join(tmp, "flowB.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbB.Close()
		_ = dbB.MarkProcessed("https://banish.example/", pb, "Design", "Assets & Resources", "full")
		gwriteNote(t, vb, "_review/bplaceholder.md", "https://reviewed.example/", "[design]", "gitcurator", "failed", "")
		gwriteNote(t, vb, "_review/bhand.md", "https://handc.example/", "[design]", "gitcurator", "failed", "")
		tb := filepath.Join(vb, "_review", "DECOMMISSIONED.md")
		gtable(t, tb, []string{
			"| 🗑️ | 2026-10-01 | https://banish.example/ | banish.example | gesture |  | |",
			"| - | 2026-10-01 | https://reviewed.example/ | reviewed.example | gesture | ✅ reviewed | |",
			"| - | 2026-10-01 | https://deadb.example/ | deadb.example | gesture | 🪦 | |",
			"| 🖐 | 2026-10-01 | https://handc.example/ | handc.example | gesture |  | |",
			"| ♻️ | 2026-10-01 | https://revive.example/ | revive.example | gesture |  | |",
			"| - | 2026-10-01 | https://gone.example/ | gone.example | gesture | 🗑️ delete | |",
		})
		_ = dbB.Dismiss("https://revive.example/", "banished by owner — x")
		if _, err := dbB.SettleExisting([]string{"https://revive.example/"}); err != nil {
			t.Fatal(err)
		}

		rep := ConsumeDecommissionTable(dbB, vb, nil, true, nil)
		queue := any(nil)
		if _, err := os.Stat(handdelivery.QueuePath(vb)); err == nil {
			queue = map[string]any{}
			json.Unmarshal([]byte(readFileT(t, handdelivery.QueuePath(vb))), &queue)
		}
		trash := []string{}
		if entries, err := os.ReadDir(filepath.Join(vb, ".trash", "banished")); err == nil {
			for _, e := range entries {
				trash = append(trash, e.Name())
			}
			sort.Strings(trash)
		}
		cmpGoldenNorm(t, "consume_confirmed", map[string]any{
			"dead":               float64(rep.Dead),
			"reviewed":           float64(rep.Reviewed),
			"banished":           float64(rep.Banished),
			"revived":            float64(rep.Revived),
			"handed":             float64(rep.Handed),
			"placeholders_swept": float64(rep.PlaceholdersSwept),
			"notes_moved":        float64(rep.NotesMoved),
			"rows_confirmed":     float64(rep.RowsConfirmed),
			"banished_urls":      rep.BanishedURLs,
			"table":              readFileT(t, tb),
			"trash":              trash,
			"queue":              queue,
			"db":                 dumpDB(t, filepath.Join(tmp, "flowB.db")),
			"review_left":        listDirT(t, filepath.Join(vb, "_review")),
		}, g, tmp)
	})
}

func TestGoldenConsumeHeld(t *testing.T) {
	g := section(t, "consume_held")
	withFrozenClock(t, func() {
		tmp := t.TempDir()
		// the generator's flow: consume_confirmed runs on VB first; the
		// held case copies the POST-CONSUME vault (stamped table, swept
		// placeholders) and appends the held row.
		vB := filepath.Join(tmp, "flowB")
		os.MkdirAll(filepath.Join(vB, "_review"), 0o755)
		os.MkdirAll(filepath.Join(vB, "Design", "Assets & Resources"), 0o755)
		pb := gwriteNote(t, vB, "Design/Assets & Resources/bnote.md", "https://banish.example/", "[design]", "gitcurator", "full", "")
		dbB, err := NewWebsiteStateDB(filepath.Join(tmp, "flowB.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbB.Close()
		_ = dbB.MarkProcessed("https://banish.example/", pb, "Design", "Assets & Resources", "full")
		gwriteNote(t, vB, "_review/bplaceholder.md", "https://reviewed.example/", "[design]", "gitcurator", "failed", "")
		gwriteNote(t, vB, "_review/bhand.md", "https://handc.example/", "[design]", "gitcurator", "failed", "")
		tb := filepath.Join(vB, "_review", "DECOMMISSIONED.md")
		gtable(t, tb, []string{
			"| 🗑️ | 2026-10-01 | https://banish.example/ | banish.example | gesture |  | |",
			"| - | 2026-10-01 | https://reviewed.example/ | reviewed.example | gesture | ✅ reviewed | |",
			"| - | 2026-10-01 | https://deadb.example/ | deadb.example | gesture | 🪦 | |",
			"| 🖐 | 2026-10-01 | https://handc.example/ | handc.example | gesture |  | |",
			"| ♻️ | 2026-10-01 | https://revive.example/ | revive.example | gesture |  | |",
			"| - | 2026-10-01 | https://gone.example/ | gone.example | gesture | 🗑️ delete | |",
		})
		_ = dbB.Dismiss("https://revive.example/", "banished by owner — x")
		if _, err := dbB.SettleExisting([]string{"https://revive.example/"}); err != nil {
			t.Fatal(err)
		}
		ConsumeDecommissionTable(dbB, vB, nil, true, nil)

		vc := filepath.Join(tmp, "flowC")
		copyTree(t, vB, vc)
		gwriteNote(t, vc, "Design/Assets & Resources/cnote.md", "https://held.example/", "[design]", "gitcurator", "full", "")
		tc := filepath.Join(vc, "_review", "DECOMMISSIONED.md")
		f, err := os.OpenFile(tc, os.O_APPEND|os.O_WRONLY, 0o644)
		if err != nil {
			t.Fatal(err)
		}
		f.WriteString("| 🗑️ | 2026-10-01 | https://held.example/ | held.example | gesture | delete | |\n")
		f.Close()
		rep := ConsumeDecommissionTable(dbB, vc, nil, false, nil)
		cmpGoldenJSON(t, "consume_held", map[string]any{
			"pending_banish":   float64(rep.PendingBanish),
			"banished":         float64(rep.Banished),
			"banished_urls":    rep.BanishedURLs,
			"table":            readFileT(t, tc),
			"held_note_exists": fileExists(filepath.Join(vc, "Design", "Assets & Resources", "cnote.md")),
		}, g)
	})
}

func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func TestGoldenBanishTwins(t *testing.T) {
	g := section(t, "banish_twins")
	withFrozenClock(t, func() {
		tmp := t.TempDir()
		vd := filepath.Join(tmp, "other-vault")
		os.MkdirAll(filepath.Join(vd, "Design"), 0o755)
		gwriteNote(t, vd, "Design/twin.md", "https://banish.example/", "[design]", "gitcurator", "full", "")
		gwriteNote(t, vd, "Design/human-twin.md", "https://banish.example/", "[design]", "human", "full", "")
		twins := BanishTwinsInOtherVault(vd, []string{"https://banish.example/"}, nil)
		cmpGoldenJSON(t, "banish_twins", map[string]any{
			"moved": float64(twins),
			"trash": listDirT(t, filepath.Join(vd, ".trash", "banished")),
			"left":  listDirT(t, filepath.Join(vd, "Design")),
		}, g)
	})
}

func TestGoldenSplitsAndSettlement(t *testing.T) {
	g := section(t, "splits")
	gl := section(t, "settle_the_ledger")
	withFrozenClock(t, func() {
		tmp := t.TempDir()
		dbE, err := NewWebsiteStateDB(filepath.Join(tmp, "flowE.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbE.Close()
		_ = dbE.Dismiss("https://split-a.example/", "auto-verdict: dead")
		if _, err := dbE.SettleExisting([]string{"https://split-b.example/"}); err != nil {
			t.Fatal(err)
		}
		s1 := SplitDismissedLinks(dbE, []string{"https://split-a.example/", "https://keep.example/"}, nil)
		s2 := SplitSettledLinks(dbE, []string{"https://split-b.example/", "https://keep2.example/"}, nil)
		cmpGoldenJSON(t, "splits", map[string]any{
			"dismissed_kept": s1.Kept, "dismissed_dropped": s1.Dropped,
			"settled_kept": s2.Kept, "settled_dropped": s2.Dropped,
		}, g)

		// the settlement (the generator's flow: VF is the POST-REFRESH
		// vault — its table + notes are the settlement's extras)
		vf := filepath.Join(tmp, "flowF")
		os.MkdirAll(filepath.Join(vf, "_review"), 0o755)
		dbF2, err := NewWebsiteStateDB(filepath.Join(tmp, "flowF2.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbF2.Close()
		_ = dbF2.EnqueueRetry("https://refresh-wait.example/r", "HTTP 404 — not found")
		gwriteNote(t, vf, "_review/lost-placeholder.md", "https://lost.example/", "[design]", "gitcurator", "failed", "")
		gwriteNote(t, vf, "_review/eyes.md", "https://eyes.example/", "[design]", "gitcurator", "partial",
			"> [!warning] Needs review — partial\n> classification confidence was low\n")
		RefreshMasterTable(dbF2, vf, nil)
		dbL, err := NewWebsiteStateDB(filepath.Join(tmp, "flowL.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbL.Close()
		_ = dbL.MarkProcessed("https://ledger.example/a", "", "Design", "", "full")
		rep := SettleTheLedger(dbL, vf, nil, nil)
		rep2 := SettleTheLedger(dbL, vf, nil, nil)
		rep3 := SettleTheLedger(dbL, vf, nil, []string{"https://qhistory.example/q"})
		cmpGoldenJSON(t, "settle_the_ledger", map[string]any{
			"first": map[string]any{
				"settled": float64(rep.Settled), "retries_cleared": float64(rep.RetriesCleared),
				"already": rep.Already,
			},
			"second": map[string]any{
				"settled": float64(rep2.Settled), "retries_cleared": float64(rep2.RetriesCleared),
				"already": rep2.Already,
			},
			"queue_history": map[string]any{
				"settled": float64(rep3.Settled), "retries_cleared": float64(rep3.RetriesCleared),
				"already": rep3.Already,
			},
			"db": dumpDB(t, filepath.Join(tmp, "flowL.db")),
		}, gl)
	})
}

func TestGoldenStampHarvestPruneRefreshSweep(t *testing.T) {
	gs := section(t, "stamp_stored_rows")
	gh := section(t, "harvest_hand_rows")
	gp := section(t, "prune")
	gr := section(t, "refresh_master_table")
	gsw := section(t, "sweep_review_leftovers")
	grr := section(t, "review_note_reason")

	withFrozenClock(t, func() {
		tmp := t.TempDir()
		ve := filepath.Join(tmp, "flowE")
		os.MkdirAll(filepath.Join(ve, "_review"), 0o755)
		os.MkdirAll(filepath.Join(ve, "Design", "Assets & Resources"), 0o755)
		pstored := gwriteNote(t, ve, "Design/Assets & Resources/stored-note.md", "https://stamp.example/", "[design]", "gitcurator", "full", "")
		dbS, err := NewWebsiteStateDB(filepath.Join(tmp, "flowS.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbS.Close()
		_ = dbS.MarkProcessed("https://stamp.example/", pstored, "Design", "Assets & Resources", "full")
		phand := gwriteNote(t, ve, "Design/Assets & Resources/hand-note.md", "https://harvest.example/", "[design]", "gitcurator", "full", "")
		_ = dbS.MarkProcessed("https://harvest.example/", phand, "Design", "Assets & Resources", "full")
		te := filepath.Join(ve, "_review", "DECOMMISSIONED.md")
		gtable(t, te, []string{
			"| - | 2026-10-01 | https://stamp.example/ | stamp.example | waiting | - | |",
			"| 🖐 | 2026-10-01 | https://harvest.example/ | harvest.example | waiting | hand | |",
			"| - | 2026-10-01 | https://terminal.example/ | terminal.example | w | 🪦 confirmed — decommissioned 2026-10-01 | |",
			"| - | 2026-10-01 | https://pending.example/ | pending.example | w | unreviewed | |",
		})

		stamped := StampStoredRows(dbS, ve, nil)
		if stamped != gint(t, gs, "stamped") {
			t.Errorf("stamp_stored_rows: got %d want %d", stamped, gint(t, gs, "stamped"))
		}
		if got := readFileT(t, te); got != gstr(t, gs, "table") {
			t.Errorf("stamp_stored_rows table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, gs, "table"))
		}

		harvested := HarvestHandRows(dbS, ve, nil)
		if harvested != gint(t, gh, "harvested") {
			t.Errorf("harvest_hand_rows: got %d want %d", harvested, gint(t, gh, "harvested"))
		}
		if got := readFileT(t, te); got != gstr(t, gh, "table") {
			t.Errorf("harvest table diverges:\n--- go ---\n%s\n--- reference ---\n%s", got, gstr(t, gh, "table"))
		}

		pruned := PruneDecommissionTable(ve, nil)
		cmpGoldenJSON(t, "prune", map[string]any{
			"pruned": float64(pruned.Pruned), "kept": float64(pruned.Kept),
			"pruned_urls": pruned.PrunedURLs,
			"table":       readFileT(t, te),
		}, gp)

		// the refresh (flow F's own vault)
		vf := filepath.Join(tmp, "flowF")
		os.MkdirAll(filepath.Join(vf, "_review"), 0o755)
		dbF, err := NewWebsiteStateDB(filepath.Join(tmp, "flowF.db"))
		if err != nil {
			t.Fatal(err)
		}
		defer dbF.Close()
		_ = dbF.EnqueueRetry("https://refresh-wait.example/r", "HTTP 404 — not found")
		gwriteNote(t, vf, "_review/lost-placeholder.md", "https://lost.example/", "[design]", "gitcurator", "failed", "")
		gwriteNote(t, vf, "_review/eyes.md", "https://eyes.example/", "[design]", "gitcurator", "partial",
			"> [!warning] Needs review — partial\n> classification confidence was low\n")
		repF := RefreshMasterTable(dbF, vf, nil)
		cmpGoldenJSON(t, "refresh_master_table", map[string]any{
			"waiting": float64(repF.Waiting), "retired": float64(repF.Retired),
			"written": float64(repF.Written),
			"table":   readFileT(t, filepath.Join(vf, "_review", "DECOMMISSIONED.md")),
		}, gr)

		// sweep_review_leftovers (flow G)
		vg := filepath.Join(tmp, "flowG")
		os.MkdirAll(filepath.Join(vg, "_review"), 0o755)
		gwriteNote(t, vg, "_review/sweep-me.md", "https://sweep.example/", "[design]", "gitcurator", "failed", "")
		gwriteNote(t, vg, "_review/keep-hand.md", "https://sweep.example/", "[design]", "human", "full", "")
		gwriteNote(t, vg, "_review/other-link.md", "https://other.example/", "[design]", "gitcurator", "failed", "")
		swept := SweepReviewLeftovers(vg, "https://sweep.example/", "", nil)
		cmpGoldenJSON(t, "sweep_review_leftovers", map[string]any{
			"swept": float64(swept),
			"left":  listDirT(t, filepath.Join(vg, "_review")),
		}, gsw)

		// the reason reader over the fixture vault
		vault := buildFixtureVault(t)
		for name, path := range map[string]string{
			"built":        filepath.Join(vault, "_review", "failed-placeholder.md"),
			"eyes":         filepath.Join(vault, "_review", "low-confidence.md"),
			"hand_written": filepath.Join(vault, "_review", "hand-written.md"),
			"missing":      filepath.Join(tmp, "gone.md"),
		} {
			if got := ReviewNoteReason(path); got != gstr(t, grr, name) {
				t.Errorf("review_note_reason.%s: got %q want %q", name, got, gstr(t, grr, name))
			}
		}
	})
}

// --- the golden dumper (direct SQL, mirroring the generator's SELECTs) ----

// dumpDB reads the temp state DB directly — the SAME SELECTs the
// generator's dump_db()/dump_db_path() run, timestamp-free (the *_at
// columns are T-23's own golden's territory).
func dumpDB(t *testing.T, path string) map[string]any {
	t.Helper()
	db, err := sql.Open("sqlite", "file:"+path+"?mode=ro")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	out := map[string]any{}
	readRows := func(query string, cols int) []any {
		rows, err := db.Query(query)
		if err != nil {
			t.Fatalf("dumpDB query %q: %v", query, err)
		}
		defer rows.Close()
		var list []any
		for rows.Next() {
			vals := make([]any, cols)
			ptrs := make([]any, cols)
			for i := range vals {
				ptrs[i] = &vals[i]
			}
			if err := rows.Scan(ptrs...); err != nil {
				t.Fatal(err)
			}
			item := make([]any, cols)
			for i, v := range vals {
				if v == nil {
					item[i] = nil
				} else if b, ok := v.([]byte); ok {
					item[i] = string(b)
				} else {
					item[i] = v
				}
			}
			list = append(list, item)
		}
		sort.Slice(list, func(i, j int) bool {
			a := list[i].([]any)[0]
			b := list[j].([]any)[0]
			return fmt.Sprintf("%v", a) < fmt.Sprintf("%v", b)
		})
		return list
	}
	out["dismissed"] = readRows("SELECT url, reason FROM dismissed_urls", 2)
	out["processed"] = readRows("SELECT url, note_path, fetch_status FROM websites_processed", 3)
	out["retry"] = readRows("SELECT url, attempts, last_error FROM website_retry_queue", 3)
	// settled is a FLAT sorted string list in the generator
	if rows, err := db.Query("SELECT url FROM websites_settled"); err == nil {
		var settled []any
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err != nil {
				t.Fatal(err)
			}
			settled = append(settled, u)
		}
		rows.Close()
		sort.Slice(settled, func(i, j int) bool {
			return settled[i].(string) < settled[j].(string)
		})
		out["settled"] = settled
	}
	out["meta"] = readRows("SELECT key, value FROM website_state_meta", 2)
	return out
}
