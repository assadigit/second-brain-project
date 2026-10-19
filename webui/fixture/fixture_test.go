// fixture_test.go — the fixture's own laws: determinism (same bytes every
// run), the managed-law preconditions (source: line inside the head-800
// window), the stub law (no source line), and the negative-notes guard.
// No engine imports (the webui module cannot reach internal/ — the
// internal-package rule, ADR-0002 Am.3).
package fixture

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func countMD(t *testing.T, root string) int {
	t.Helper()
	n := 0
	err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			n++
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	return n
}

func TestWriteVaultShape(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	v, err := WriteVault(root, 300)
	if err != nil {
		t.Fatal(err)
	}
	if v.Notes != 300 {
		t.Fatalf("notes = %d, want 300", v.Notes)
	}
	if v.Unmanaged != 2+300/200 {
		t.Fatalf("unmanaged = %d, want %d", v.Unmanaged, 2+300/200)
	}
	if got := countMD(t, root); got != v.Notes+v.Unmanaged {
		t.Fatalf("tree holds %d .md files, want %d", got, v.Notes+v.Unmanaged)
	}
}

func TestWriteVaultDeterministic(t *testing.T) {
	// same input → same bytes: the corpus law's fixture twin (a fixture
	// that flips between runs would make the job's numbers dishonest).
	rootA := filepath.Join(t.TempDir(), "a")
	rootB := filepath.Join(t.TempDir(), "b")
	if _, err := WriteVault(rootA, 50); err != nil {
		t.Fatal(err)
	}
	if _, err := WriteVault(rootB, 50); err != nil {
		t.Fatal(err)
	}
	snapshot := func(root string) []string {
		t.Helper()
		var files []string
		err := filepath.WalkDir(root, func(path string, d os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if !d.IsDir() {
				rel, _ := filepath.Rel(root, path)
				data, err := os.ReadFile(path)
				if err != nil {
					return err
				}
				files = append(files, rel+"|"+string(data))
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		return files
	}
	filesA, filesB := snapshot(rootA), snapshot(rootB)
	if len(filesA) != len(filesB) {
		t.Fatalf("file counts differ: %d vs %d", len(filesA), len(filesB))
	}
	for i := range filesA {
		if filesA[i] != filesB[i] {
			t.Fatalf("file %d differs between runs (non-deterministic fixture)", i)
		}
	}
}

func TestManagedNotesCarrySourceLine(t *testing.T) {
	root := filepath.Join(t.TempDir(), "vault")
	if _, err := WriteVault(root, 24); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 24; i++ {
		folder := filepath.Join(root, realFolders[i%len(realFolders)])
		p := filepath.Join(folder, fmt.Sprintf("fixture-note-%04d.md", i))
		data, err := os.ReadFile(p)
		if err != nil {
			t.Fatalf("note %d missing: %v", i, err)
		}
		head := string(data)
		if len([]rune(head)) > 800 {
			head = string([]rune(head)[:800])
		}
		if !strings.Contains(head, "source: https://github.com/fixture-org/repo-") {
			t.Fatalf("note %d's head-800 lacks its source line", i)
		}
	}
	// the stubs never carry one.
	stub := filepath.Join(root, "Documentation/Guides", "fixture-stub-00.md")
	data, err := os.ReadFile(stub)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "source:") {
		t.Fatal("a stub carries a source line — it would be adopted (the never-adopted law broken)")
	}
}

func TestNegativeNotesRefused(t *testing.T) {
	if _, err := WriteVault(t.TempDir(), -1); err == nil {
		t.Fatal("negative notes must be refused")
	}
}
