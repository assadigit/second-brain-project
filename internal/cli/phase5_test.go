package cli

// phase5_test.go — THE PHASE-5 CLEAR's unit battery (the Option-B
// re-spec's own laws): the all-clear verdict speech, the not-clear
// count, the tracker-nil honesty, the gate (success AND bot_source),
// the mirror push's three outcomes (ok / failed / disabled-silent),
// and the dry-run report writer (the path + the body + the frozen
// stamp). The differential bytes are the golden replay's job.

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// sha256Hex is the test's hash twin (the manager's own recipe).
func sha256Hex(s string) string {
	sum := sha256.Sum256([]byte(s))
	return hex.EncodeToString(sum[:])
}

// clearTracker builds a LinkTracker over a temp vault with N github
// links (all marked processed when clear).
func clearTracker(t *testing.T, n int, clear bool) *batch.LinkTracker {
	t.Helper()
	vault := t.TempDir()
	tr := batch.NewLinkTracker(vault)
	var urls []string
	for i := 0; i < n; i++ {
		urls = append(urls, "https://github.com/o/repo"+strings.Repeat("x", i+1))
	}
	tr.Intake(urls, nil)
	if clear {
		for _, u := range urls {
			tr.MarkProcessed(u, "note.md")
		}
	}
	return tr
}

func pushCLI(buf *bytes.Buffer, attempted, ok bool, resp string) *CLI {
	c := testCLI(buf, "")
	c.WorkerMirrorPush = func(_ *storage.OrderedMap, _, _ []any, _ string) (bool, bool, any) {
		return attempted, ok, resp
	}
	return c
}

// TestPhaseFiveClearVerdicts: the re-spec's verdict laws.
func TestPhaseFiveClearVerdicts(t *testing.T) {
	t.Run("all clear speaks the clear", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := pushCLI(buf, true, true, "")
		tr := clearTracker(t, 2, true)
		c.PhaseFiveClear(nil, true, true, tr, nil, nil)
		out := buf.String()
		if !strings.Contains(out, "✓ Bot queue synced — all links verified (Phase 5 CLEAR") {
			t.Fatalf("missing the clear line:\n%s", out)
		}
	})
	t.Run("not clear counts and stays pending", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := pushCLI(buf, true, true, "")
		tr := clearTracker(t, 2, false)
		c.PhaseFiveClear(nil, true, true, tr, nil, nil)
		out := buf.String()
		if !strings.Contains(out, "⏸️ 2 link(s) not verified — they stay PENDING in the Worker queue") {
			t.Fatalf("missing the pending line:\n%s", out)
		}
	})
	t.Run("tracker nil speaks the honest no-manifest line", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := pushCLI(buf, false, false, "")
		c.PhaseFiveClear(nil, true, true, nil, nil, nil)
		out := buf.String()
		if !strings.Contains(out, "No link manifest for this batch (empty vault path?)") {
			t.Fatalf("missing the no-manifest line:\n%s", out)
		}
	})
	t.Run("not bot source stays silent on the verdict", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := pushCLI(buf, false, false, "")
		tr := clearTracker(t, 2, true)
		c.PhaseFiveClear(nil, false, true, tr, nil, nil)
		out := buf.String()
		if strings.Contains(out, "Phase 5 CLEAR") || strings.Contains(out, "⏸️") {
			t.Fatalf("verdict spoken for a non-bot batch:\n%s", out)
		}
	})
	t.Run("failed batch stays silent on the verdict", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := pushCLI(buf, false, false, "")
		tr := clearTracker(t, 2, true)
		c.PhaseFiveClear(nil, true, false, tr, nil, nil)
		if strings.Contains(buf.String(), "Phase 5 CLEAR") {
			t.Fatalf("verdict spoken for a failed batch:\n%s", buf.String())
		}
	})
}

// TestPushVaultMirrorAfterBatch: the ack leg's three outcomes + the
// entries/decommissions built from the index + the cache.
func TestPushVaultMirrorAfterBatch(t *testing.T) {
	t.Run("ok", func(t *testing.T) {
		buf := &bytes.Buffer{}
		var gotEntries []any
		var gotHash string
		c := testCLI(buf, "")
		c.WorkerMirrorPush = func(_ *storage.OrderedMap, entries, _ []any, hash string) (bool, bool, any) {
			gotEntries, gotHash = entries, hash
			return true, true, ""
		}
		vault := t.TempDir()
		idx := batch.NewVaultIndex(vault, nil)
		idx.AddURL("https://github.com/b/x", filepath.Join(vault, "b.md"))
		idx.AddURL("https://github.com/a/y", filepath.Join(vault, "a.md"))
		out := c.PushVaultMirrorAfterBatch(nil, idx, nil)
		if !out.Attempted || !out.OK || out.Entries != 2 {
			t.Fatalf("outcome: %+v", out)
		}
		if !strings.Contains(buf.String(), "☁️ Worker synced: 2 vault entry(ies) and 0 decommission event(s) pushed") {
			t.Fatalf("missing the push line:\n%s", buf.String())
		}
		// SORTED url order (the deterministic-wire law).
		if len(gotEntries) != 2 {
			t.Fatalf("entries: %d", len(gotEntries))
		}
		// the hash: sha256('|'.join(sorted(urls)))
		wantHash := sha256Hex("https://github.com/a/y|https://github.com/b/x")
		if gotHash != wantHash {
			t.Fatalf("hash %s != %s", gotHash, wantHash)
		}
	})
	t.Run("failed", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.WorkerMirrorPush = func(*storage.OrderedMap, []any, []any, string) (bool, bool, any) {
			return true, false, "{'error': 'down'}"
		}
		vault := t.TempDir()
		idx := batch.NewVaultIndex(vault, nil)
		idx.AddURL("https://github.com/a/y", filepath.Join(vault, "a.md"))
		out := c.PushVaultMirrorAfterBatch(nil, idx, nil)
		if !out.Attempted || out.OK {
			t.Fatalf("outcome: %+v", out)
		}
		if !strings.Contains(buf.String(), "⚠️ Vault-mirror push failed: \"{'error': 'down'}\"") {
			t.Fatalf("missing the failure line:\n%s", buf.String())
		}
	})
	t.Run("disabled is silent", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.WorkerMirrorPush = func(*storage.OrderedMap, []any, []any, string) (bool, bool, any) {
			return false, false, nil
		}
		vault := t.TempDir()
		idx := batch.NewVaultIndex(vault, nil)
		idx.AddURL("https://github.com/a/y", filepath.Join(vault, "a.md"))
		out := c.PushVaultMirrorAfterBatch(nil, idx, nil)
		if out.Attempted || buf.Len() != 0 {
			t.Fatalf("disabled push spoke: %+v %q", out, buf.String())
		}
	})
	t.Run("decommissions from the cache", func(t *testing.T) {
		buf := &bytes.Buffer{}
		var gotDecomm []any
		c := testCLI(buf, "")
		c.WorkerMirrorPush = func(_ *storage.OrderedMap, _, decomm []any, _ string) (bool, bool, any) {
			gotDecomm = decomm
			return true, true, ""
		}
		cache, err := store.NewCacheDB(filepath.Join(t.TempDir(), "cache.db"))
		if err != nil {
			t.Fatalf("cache: %v", err)
		}
		defer cache.Close()
		if _, err := cache.SQL().Exec(
			"INSERT INTO decommissioned_repos (url, reason, decommissioned_at, fail_count)" +
				" VALUES ('https://github.com/gone/r', '404', '2026-10-17T12:00:00', 3)"); err != nil {
			t.Fatalf("insert: %v", err)
		}
		vault := t.TempDir()
		idx := batch.NewVaultIndex(vault, nil)
		idx.AddURL("https://github.com/a/y", filepath.Join(vault, "a.md"))
		out := c.PushVaultMirrorAfterBatch(nil, idx, cache)
		if out.Decomm != 1 || len(gotDecomm) != 1 {
			t.Fatalf("decommissions: %d / %d", out.Decomm, len(gotDecomm))
		}
	})
}

// TestWriteDryrunReport: the path, the frozen stamp, and the body over
// recorded entries.
func TestWriteDryrunReport(t *testing.T) {
	appTmp := t.TempDir()
	t.Setenv("GITCURATOR_HOME", appTmp)
	dryrun.Clear()
	defer dryrun.Clear()
	// the frozen clock: both the entry stamps and the Generated line
	// ride the dryrun package's seam (the golden freezes the reference
	// module's datetime the same way).
	oldNow := dryrun.NowFn
	dryrun.NowFn = frozenClock
	defer func() { dryrun.NowFn = oldNow }()
	dryrun.Enable()
	dryrun.Record("write", "/vault/note.md", "the content", "")
	dryrun.Record("makedirs", "/vault/sub", nil, "create folder")
	dryrun.Disable()

	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	path := c.writeDryrunReport()
	if path == "" {
		t.Fatalf("no report path")
	}
	wantDir := filepath.Join(appTmp, "reports", "dry-runs")
	if filepath.Dir(path) != wantDir {
		t.Fatalf("dir %s != %s", filepath.Dir(path), wantDir)
	}
	if !strings.Contains(path, "dry_run_20261017-120000.md") {
		t.Fatalf("stamp missing from %s (the frozen clock 2026-10-17 12:00:00)", path)
	}
	body, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read: %v", err)
	}
	s := string(body)
	for _, want := range []string{
		"# GitCurator dry-run report",
		"**Nothing below was written.**",
		"| 1 | 12:00:00 | write | `/vault/note.md` |",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("body missing %q:\n%s", want, s)
		}
	}
}
