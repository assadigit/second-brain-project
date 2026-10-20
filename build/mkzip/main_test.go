package main

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
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

// stubExe writes a fake CLI exe (the laws under test don't care about
// the exe's bytes — only that it lands in the zip under its arcname and
// carries the version stamp, which the T-80 guard asserts).
func stubExe(t *testing.T, dir, version string) string {
	t.Helper()
	p := filepath.Join(dir, "GitCurator-cli.exe")
	data := append(bytes.Repeat([]byte{0x4D, 0x5A}, 512), []byte(version)...)
	if err := os.WriteFile(p, data, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// stubGuiExe writes a fake GUI exe carrying the version stamp AND the
// brand icon's bytes (the icon guard probes the exe for the committed
// icon's tail — a real wails build embeds it via the .syso).
func stubGuiExe(t *testing.T, dir, version string) string {
	t.Helper()
	ico, err := os.ReadFile(filepath.Join(repoRoot, "webui", "build",
		"windows", "icon.ico"))
	if err != nil {
		t.Fatalf("read the repo's brand icon (committed at webui/build/"+
			"windows/icon.ico): %v", err)
	}
	p := filepath.Join(dir, "GitCurator.exe")
	data := append(bytes.Repeat([]byte{0x4D, 0x5A}, 512), []byte(version)...)
	data = append(data, ico...)
	if err := os.WriteFile(p, data, 0o755); err != nil {
		t.Fatal(err)
	}
	return p
}

// buildFromRepo runs the builder against the real repo with stub exes.
func buildFromRepo(t *testing.T, outPath string) {
	t.Helper()
	if err := run(repoRoot, "0.92.0-go.1-test", stubExe(t, t.TempDir(), "0.92.0-go.1-test"),
		stubGuiExe(t, t.TempDir(), "0.92.0-go.1-test"), outPath); err != nil {
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

// TestZipComposition — the T-80 two-exe composition: the GUI exe, the
// CLI exe, the flattened CLI launchers, the GUI-side pair, the assets,
// the stamp.
func TestZipComposition(t *testing.T) {
	out := filepath.Join(t.TempDir(), "z.zip")
	buildFromRepo(t, out)
	r := openZip(t, out)

	names := []string{}
	for _, f := range r.File {
		names = append(names, f.Name)
	}
	must := []string{
		"GitCurator/GitCurator.exe",
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

// TestVersionStampGuard — a build that forgot -ldflags "-X
// main.version=…" (the var's "dev" default) never reaches the zip:
// BOTH exes are guarded.
func TestVersionStampGuard(t *testing.T) {
	ver := "9.9.9-guard-test"
	// The CLI exe without the stamp → refused.
	bare := filepath.Join(t.TempDir(), "bare.exe")
	if err := os.WriteFile(bare, bytes.Repeat([]byte{0x4D, 0x5A}, 64), 0o755); err != nil {
		t.Fatal(err)
	}
	gui := stubGuiExe(t, t.TempDir(), ver)
	if err := guardVersionStamp("GitCurator-cli.exe", nil, ver); err == nil {
		t.Fatal("nil exe data passed the stamp guard")
	}
	if err := run(repoRoot, ver, bare, gui, filepath.Join(t.TempDir(), "z.zip")); err == nil {
		t.Fatal("an unstamped CLI exe passed the guard")
	}
	// The GUI exe without the stamp → refused.
	cli := stubExe(t, t.TempDir(), ver)
	if err := run(repoRoot, ver, cli, bare, filepath.Join(t.TempDir(), "z.zip")); err == nil {
		t.Fatal("an unstamped GUI exe passed the guard")
	}
	// A stamped pair passes the stamp law (the icon guard rides its own
	// law below — build the GUI stub with the brand icon, as usual).
	out := filepath.Join(t.TempDir(), "ok.zip")
	if err := run(repoRoot, ver, cli, gui, out); err != nil {
		t.Fatalf("a stamped pair was refused: %v", err)
	}
}

// TestBrandIconGuard — a GUI exe built with the wails DEFAULT icon
// (icon.ico's bytes absent) never reaches the zip.
func TestBrandIconGuard(t *testing.T) {
	ver := "9.9.9-icon-test"
	cli := stubExe(t, t.TempDir(), ver)
	noIcon := filepath.Join(t.TempDir(), "noicon.exe")
	if err := os.WriteFile(noIcon, append(bytes.Repeat([]byte{0x4D, 0x5A}, 512),
		[]byte(ver)...), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := guardBrandIcon(repoRoot, nil); err == nil {
		t.Fatal("nil GUI data passed the icon guard")
	}
	if err := run(repoRoot, ver, cli, noIcon, filepath.Join(t.TempDir(), "z.zip")); err == nil {
		t.Fatal("a default-icon GUI exe passed the guard")
	}
}

// TestManifestLaws — doc 10 §5's manifest laws over the committed
// webui/build/windows/wails.exe.manifest: DPI-aware per-monitor-v2,
// supportedOS Win10, long-path aware.
func TestManifestLaws(t *testing.T) {
	data, err := os.ReadFile(filepath.Join(repoRoot, "webui", "build",
		"windows", "wails.exe.manifest"))
	if err != nil {
		t.Fatalf("read the manifest: %v", err)
	}
	m := string(data)
	for _, want := range []string{
		"<dpiAwareness xmlns=\"http://schemas.microsoft.com/SMI/2016/WindowsSettings\">permonitorv2,permonitor</dpiAwareness>",
		"<supportedOS Id=\"{8e0f7a12-bfb3-4fe8-b9a5-48fd50a15a9a}\"/>",
		"<longPathAware xmlns=\"http://schemas.microsoft.com/SMI/2016/WindowsSettings\">true</longPathAware>",
	} {
		if !strings.Contains(m, want) {
			t.Errorf("manifest law missing: %s", want)
		}
	}
}

// TestIconLaws — the exe's icon source is the BRAND icon: webui/build/
// windows/icon.ico is byte-identical to assets/icons/gitcurator.ico
// (the seed commit's brand icon, 7 PNG-compressed sizes), and the
// generic appicon.png is NOT the wails template default (the artwork
// drift the T-80 find exposed).
func TestIconLaws(t *testing.T) {
	ico, err := os.ReadFile(filepath.Join(repoRoot, "assets", "icons",
		"gitcurator.ico"))
	if err != nil {
		t.Fatalf("read the brand icon: %v", err)
	}
	shipped, err := os.ReadFile(filepath.Join(repoRoot, "webui", "build",
		"windows", "icon.ico"))
	if err != nil {
		t.Fatalf("read the shipped icon source: %v", err)
	}
	if !bytes.Equal(ico, shipped) {
		t.Fatalf("webui/build/windows/icon.ico is not the brand icon: "+
			"brand sha-tail %x vs shipped %x", ico[len(ico)-8:], shipped[len(shipped)-8:])
	}
	// The ICO grammar: magic 00 00 01 00 + the 7-entry directory.
	if len(ico) < 6 || ico[0] != 0 || ico[1] != 0 || ico[2] != 1 || ico[3] != 0 {
		t.Fatalf("icon.ico is not an ICO (magic bytes)")
	}
	if int(ico[5])<<8|int(ico[4]) != 7 {
		t.Fatalf("brand icon entry count = %d; want 7", int(ico[5])<<8|int(ico[4]))
	}
	// The generic appicon: the brand artwork, never the wails default
	// (sha256 28332a3b130d1970d208e40b7ab2c84f9ead10199f8c058e08e07fe21fe7ea49).
	appicon, err := os.ReadFile(filepath.Join(repoRoot, "webui", "build",
		"appicon.png"))
	if err != nil {
		t.Fatalf("read appicon.png: %v", err)
	}
	if !bytes.HasPrefix(appicon, []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1A, '\n'}) {
		t.Fatalf("appicon.png is not a PNG")
	}
	h := sha256.Sum256(appicon)
	got := fmt.Sprintf("%x", h)
	if got == "28332a3b130d1970d208e40b7ab2c84f9ead10199f8c058e08e07fe21fe7ea49" {
		t.Fatalf("appicon.png is the WAILS DEFAULT icon — the brand artwork " +
			"was lost (extract the 256px entry from assets/icons/gitcurator.ico)")
	}
}

// TestWailsInfoLaws — THE ONE-VERSION-TRUTH law: webui/wails.json's
// Info.productVersion (the exe's version resource source) equals the
// repo VERSION file's stamp; the product name is GitCurator.
func TestWailsInfoLaws(t *testing.T) {
	verBytes, err := os.ReadFile(filepath.Join(repoRoot, "VERSION"))
	if err != nil {
		t.Fatalf("read VERSION: %v", err)
	}
	ver := strings.TrimSpace(string(verBytes))
	data, err := os.ReadFile(filepath.Join(repoRoot, "webui", "wails.json"))
	if err != nil {
		t.Fatalf("read webui/wails.json: %v", err)
	}
	var w struct {
		Name string `json:"name"`
		Info struct {
			ProductName    string `json:"productName"`
			ProductVersion string `json:"productVersion"`
		} `json:"Info"`
	}
	if err := json.Unmarshal(data, &w); err != nil {
		t.Fatalf("parse webui/wails.json: %v", err)
	}
	if w.Info.ProductVersion != ver {
		t.Errorf("wails.json Info.productVersion = %q; want the VERSION file's %q "+
			"(the one-version-truth law — the exe's version resource and the "+
			"zip's stamped VERSION must agree)", w.Info.ProductVersion, ver)
	}
	if w.Info.ProductName != "GitCurator" {
		t.Errorf("wails.json Info.productName = %q; want GitCurator", w.Info.ProductName)
	}
	// The version var the -X flag stamps must exist in BOTH mains.
	for _, src := range []string{
		filepath.Join(repoRoot, "webui", "main.go"),
		filepath.Join(repoRoot, "cmd", "gitcurator-cli", "main.go"),
	} {
		b, err := os.ReadFile(src)
		if err != nil {
			t.Fatalf("read %s: %v", src, err)
		}
		if !bytes.Contains(b, []byte("var version = \"dev\"")) {
			t.Errorf("%s lost the \"var version = \"dev\"\" declaration — the -X "+
				"stamp lands nowhere (the linker drops the value under -trimpath "+
				"when the var is unreferenced)", src)
		}
	}
}
