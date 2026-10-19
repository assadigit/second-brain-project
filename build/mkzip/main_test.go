package main

import (
	"archive/zip"
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// repoRoot — the work repo (tests run with cwd = build/mkzip).
var repoRoot = filepath.Join("..", "..")

// stubExe writes a fake exe (the laws under test don't care about the
// exe's bytes — only that it lands in the zip under its arcname).
func stubExe(t *testing.T, dir string) string {
	t.Helper()
	p := filepath.Join(dir, "GitCurator-cli.exe")
	if err := os.WriteFile(p, bytes.Repeat([]byte{0x4D, 0x5A}, 512), 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// buildFromRepo runs the builder against the real repo with a stub exe.
func buildFromRepo(t *testing.T, outPath string) {
	t.Helper()
	if err := run(repoRoot, "0.92.0-go.1-test", stubExe(t, t.TempDir()), outPath); err != nil {
		t.Fatalf("run: %v", err)
	}
}

func openZip(t *testing.T, path string) *zip.Reader {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { f.Close() })
	fi, err := f.Stat()
	if err != nil {
		t.Fatal(err)
	}
	r, err := zip.NewReader(f, fi.Size())
	if err != nil {
		t.Fatal(err)
	}
	return r
}

// TestZipComposition — the CLI-milestone composition: the exe, the
// flattened CLI launchers, the GUI-side pair, the assets, the stamp.
func TestZipComposition(t *testing.T) {
	out := filepath.Join(t.TempDir(), "z.zip")
	buildFromRepo(t, out)
	r := openZip(t, out)

	names := []string{}
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	must := []string{
		"GitCurator/GitCurator-cli.exe",
		"GitCurator/GitCurator-CLI.bat",
		"GitCurator/GitCurator-SYNC.bat",
		"GitCurator/GitCurator-CLI-Setup.bat",
		"GitCurator/Start-GitCurator-CLI.bat",
		"GitCurator/1-INSTALL.bat",
		"GitCurator/gitcurator-cli.sh",
		"GitCurator/GitCurator.bat",
		"GitCurator/GitCurator-DRY-RUN.bat",
		"GitCurator/config.example.json",
		"GitCurator/WINDOWS-QUICKSTART.md",
		"GitCurator/VERSION",
	}
	for _, want := range must {
		found := false
		for _, n := range names {
			if n == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("missing zip member: %s (have %v)", want, names)
		}
	}
	// The Go sources and tests never ship.
	for _, n := range names {
		if strings.HasSuffix(n, ".go") {
			t.Errorf("Go source shipped in the zip: %s", n)
		}
	}
	// VERSION is stamped, not copied.
	for _, f := range r.File {
		if f.Name == "GitCurator/VERSION" {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			data, _ := io.ReadAll(rc)
			rc.Close()
			if string(data) != "0.92.0-go.1-test\n" {
				t.Errorf("VERSION = %q; want the stamp", string(data))
			}
		}
	}
}

// TestZipDeterminismLaws — fixed timestamps, sorted arcnames, 0o644
// attrs, DEFLATE, .bat ASCII+CRLF.
func TestZipDeterminismLaws(t *testing.T) {
	out := filepath.Join(t.TempDir(), "z.zip")
	buildFromRepo(t, out)
	r := openZip(t, out)

	wantTime := time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)
	prev := ""
	for _, f := range r.File {
		if prev != "" && f.Name < prev {
			t.Errorf("arcnames not sorted: %s after %s", f.Name, prev)
		}
		prev = f.Name
		if !f.Modified.Equal(wantTime) {
			t.Errorf("%s: modified = %v; want %v", f.Name, f.Modified, wantTime)
		}
		if f.ExternalAttrs != fixedExternalAttrs {
			t.Errorf("%s: external attrs = %#o; want %#o",
				f.Name, f.ExternalAttrs, fixedExternalAttrs)
		}
		if f.Method != zip.Deflate {
			t.Errorf("%s: method = %d; want DEFLATE", f.Name, f.Method)
		}
		if strings.HasSuffix(f.Name, ".bat") {
			rc, err := f.Open()
			if err != nil {
				t.Fatal(err)
			}
			data, _ := io.ReadAll(rc)
			rc.Close()
			for i, b := range data {
				if b > 0x7F {
					t.Errorf("%s: byte %d non-ASCII", f.Name, i)
					break
				}
			}
			lf := bytes.Count(data, []byte("\n"))
			crlf := bytes.Count(data, []byte("\r\n"))
			if lf != crlf {
				t.Errorf("%s: bare LF in a .bat (CRLF law)", f.Name)
			}
			if !bytes.HasSuffix(data, []byte("\r\n")) {
				t.Errorf("%s: missing final CRLF", f.Name)
			}
		}
	}
}

// TestZipByteIdentity — the release law: same inputs + version → same
// bytes (two builds from the same commit are byte-identical).
func TestZipByteIdentity(t *testing.T) {
	dir := t.TempDir()
	a := filepath.Join(dir, "a.zip")
	b := filepath.Join(dir, "b.zip")
	buildFromRepo(t, a)
	buildFromRepo(t, b)
	shaA, shaB := sha256File(a), sha256File(b)
	if shaA == "" || shaA != shaB {
		t.Fatalf("two builds differ: %s vs %s", shaA, shaB)
	}
}

// TestBannedGuard — a tracked config.json in the shipped surface kills
// the build (defense in depth; a scratch git repo isolates the probe).
func TestBannedGuard(t *testing.T) {
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not on PATH")
	}
	dir := t.TempDir()
	for _, cmd := range [][]string{
		{"git", "init", "-q", dir},
	} {
		if out, err := exec.Command(cmd[0], cmd[1:]...).CombinedOutput(); err != nil {
			t.Fatalf("%v: %v\n%s", cmd, err, out)
		}
	}
	if err := os.WriteFile(filepath.Join(dir, "config.json"),
		[]byte("{\"github_token\": \"x\"}\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if out, err := exec.Command("git", "-C", dir, "add", "config.json").CombinedOutput(); err != nil {
		t.Fatalf("git add: %v\n%s", err, out)
	}
	tracked, err := collectTracked(dir)
	if err != nil {
		t.Fatal(err)
	}
	if err := guardBanned(tracked); err == nil {
		t.Fatal("a tracked config.json passed the banned guard")
	}
}

// TestNormalizeLauncherBatLaws — the .bat normalization is idempotent
// and refuses non-ASCII.
func TestNormalizeLauncherBatLaws(t *testing.T) {
	lf := []byte("@echo off\npause\n")
	got, err := normalizeLauncher("x.bat", lf)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("@echo off\r\npause\r\n")) {
		t.Fatalf("CRLF normalization = %q", got)
	}
	again, err := normalizeLauncher("x.bat", got)
	if err != nil || !bytes.Equal(again, got) {
		t.Fatalf("not idempotent: %q vs %q (%v)", again, got, err)
	}
	if _, err := normalizeLauncher("x.bat", []byte("echo caf\xe9\n")); err == nil {
		t.Fatal("non-ASCII .bat passed the ASCII guard")
	}
	if passthrough, err := normalizeLauncher("x.sh", lf); err != nil ||
		!bytes.Equal(passthrough, lf) {
		t.Fatalf(".sh must pass through untouched: %q (%v)", passthrough, err)
	}
}
