// bridge_test.go — the interim seam's contract tests (T-70): the real-job
// composition must behave like the engine's own record_baseline_if_empty
// (the one-time law, the managed-only law, the reference's final sentence),
// the progress/log instrumentation must fire in order, and the
// DSN-empty-path hazard must be refused.
package guibridge

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/notestate"
)

// writeNote drops one managed fixture note (a source: line in the head —
// the managed law) into dir/sub and returns its path.
func writeNote(t *testing.T, dir, sub, name, source string) string {
	t.Helper()
	full := filepath.Join(dir, sub)
	if err := os.MkdirAll(full, 0o755); err != nil {
		t.Fatal(err)
	}
	p := filepath.Join(full, name)
	body := "---\ntitle: " + name + "\n---\nsource: " + source + "\n\n# " + name +
		"\n\nbody text for " + name + "\n"
	if err := os.WriteFile(p, []byte(body), 0o644); err != nil {
		t.Fatal(err)
	}
	return p
}

func TestBaselineRecordsManagedNotes(t *testing.T) {
	vaultDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "cache.db")

	// three managed notes across real category folders + one at the vault
	// root (category nil, still managed), plus one unmanaged stub.
	writeNote(t, vaultDir, "AI-Domain/Agents", "agent-note.md", "https://github.com/owner/agents")
	writeNote(t, vaultDir, "Tools/Scraping", "scrape-note.md", "https://github.com/owner/scrapers")
	writeNote(t, vaultDir, "", "root-note.md", "https://github.com/owner/root")
	stub := filepath.Join(vaultDir, "stub.md")
	if err := os.WriteFile(stub, []byte("# no source line\n\nunmanaged\n"), 0o644); err != nil {
		t.Fatal(err)
	}

	var logs []string
	var prog []int
	res, err := RunNoteStateBaseline("github", vaultDir, dbPath,
		func(msg, level string) { logs = append(logs, level+"|"+msg) },
		func(pos, total int) { prog = append(prog, pos) })
	if err != nil {
		t.Fatal(err)
	}
	if res.Existing {
		t.Fatal("first run must record, not no-op")
	}
	if res.Managed != 3 || res.Recorded != 3 {
		t.Fatalf("managed/recorded = %d/%d, want 3/3", res.Managed, res.Recorded)
	}
	if res.Unmanaged != 1 {
		t.Fatalf("unmanaged = %d, want 1", res.Unmanaged)
	}
	// progress fires once per managed note, in order 1..3.
	if len(prog) != 3 || prog[0] != 1 || prog[1] != 2 || prog[2] != 3 {
		t.Fatalf("progress = %v", prog)
	}
	// the final line is the reference's own sentence, byte-identical.
	want := "info|🗂️ Note-state baseline recorded: 3 notes (github vault)"
	if logs[len(logs)-1] != want {
		t.Fatalf("final log = %q, want %q", logs[len(logs)-1], want)
	}
	// the DB actually holds the rows (identity = vault + source_url).
	db, err := notestate.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	n, err := db.Count("github")
	if err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("db count = %d, want 3", n)
	}
	for _, src := range []string{
		"https://github.com/owner/agents",
		"https://github.com/owner/scrapers",
		"https://github.com/owner/root",
	} {
		ok, err := db.IsDismissed("github", src)
		if err != nil {
			t.Fatal(err)
		}
		if ok {
			t.Fatalf("row %q unexpectedly dismissed", src)
		}
	}
}

func TestBaselineIsOneTime(t *testing.T) {
	vaultDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "cache.db")
	writeNote(t, vaultDir, "Documentation/Guides", "g1.md", "https://github.com/owner/g1")
	writeNote(t, vaultDir, "Documentation/Guides", "g2.md", "https://github.com/owner/g2")

	if _, err := RunNoteStateBaseline("github", vaultDir, dbPath, nil, nil); err != nil {
		t.Fatal(err)
	}
	db, err := notestate.New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	before, err := db.Count("github")
	if err != nil {
		t.Fatal(err)
	}

	// add MORE notes, then re-run: the one-time law says no-op.
	writeNote(t, vaultDir, "Documentation/Guides", "g3.md", "https://github.com/owner/g3")
	var logs []string
	res, err := RunNoteStateBaseline("github", vaultDir, dbPath,
		func(msg, level string) { logs = append(logs, msg) }, nil)
	if err != nil {
		t.Fatal(err)
	}
	if !res.Existing || res.Recorded != 0 || res.Managed != 0 {
		t.Fatalf("second run = %+v, want the existing/no-op shape", res)
	}
	if len(logs) != 0 {
		t.Fatalf("the no-op path must not log, got %v", logs)
	}
	after, err := db.Count("github")
	if err != nil {
		t.Fatal(err)
	}
	if after != before {
		t.Fatalf("count drifted %d -> %d on the no-op path", before, after)
	}
}

func TestEmptyDBPathRefused(t *testing.T) {
	vaultDir := t.TempDir()
	writeNote(t, vaultDir, "Uncategorized", "u.md", "https://github.com/owner/u")
	_, err := RunNoteStateBaseline("github", vaultDir, "", nil, nil)
	if err == nil {
		t.Fatal("the empty db path must be refused (the DSN-empty-path hazard)")
	}
	if !strings.Contains(err.Error(), "DSN-empty-path") {
		t.Fatalf("err = %v (must name the hazard)", err)
	}
}

func TestNilCallbacksAreNoop(t *testing.T) {
	vaultDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "cache.db")
	writeNote(t, vaultDir, "Infrastructure", "i1.md", "https://github.com/owner/i1")
	res, err := RunNoteStateBaseline("github", vaultDir, dbPath, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if res.Recorded != 1 {
		t.Fatalf("recorded = %d, want 1", res.Recorded)
	}
}

func TestProgressLogCadence(t *testing.T) {
	// 60 managed notes: per-50 cadence + the final line + the walk line.
	vaultDir := t.TempDir()
	dbPath := filepath.Join(t.TempDir(), "cache.db")
	for i := 0; i < 60; i++ {
		writeNote(t, vaultDir, "Frameworks/Backend", fmt.Sprintf("b%02d.md", i),
			fmt.Sprintf("https://github.com/owner/b%02d", i))
	}
	var msgs []string
	if _, err := RunNoteStateBaseline("github", vaultDir, dbPath,
		func(msg, level string) { msgs = append(msgs, msg) }, nil); err != nil {
		t.Fatal(err)
	}
	// walk line + baseline 50/60 + baseline 60/60 + the reference's final
	// sentence = 4 lines.
	if len(msgs) != 4 {
		t.Fatalf("log lines = %d (%v), want 4", len(msgs), msgs)
	}
	if msgs[1] != "baseline 50/60 notes recorded" || msgs[2] != "baseline 60/60 notes recorded" {
		t.Fatalf("cadence lines = %q, %q", msgs[1], msgs[2])
	}
}
