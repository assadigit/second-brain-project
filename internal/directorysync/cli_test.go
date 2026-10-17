// cli_test.go — T-49: the ported selections of the reference's
// TestCli (test_directorysync.py) — the --out door writes the built
// index JSON; a missing vault answers exit 2.
package directorysync

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCli_MainPrintsAndWritesIndex(t *testing.T) {
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	outPath := filepath.Join(v.dir, "index.json")
	rc := Main([]string{"--vault", v.dir, "--out", outPath, "--config", "/nonexistent.json"},
		strings.NewReader(""), io.Discard, io.Discard)
	if rc != 0 {
		t.Fatalf("rc: %d", rc)
	}
	written, err := os.ReadFile(outPath)
	if err != nil {
		t.Fatal(err)
	}
	// the counts ride the written index
	if !strings.Contains(string(written), `"sites": 1`) {
		t.Fatalf("written index: %s", written[:200])
	}
}

func TestCli_MainWithoutVaultReturnsTwo(t *testing.T) {
	rc := Main([]string{"--vault", "", "--config", "/nonexistent.json"},
		strings.NewReader(""), io.Discard, io.Discard)
	if rc != 2 {
		t.Fatalf("rc: %d", rc)
	}
}
