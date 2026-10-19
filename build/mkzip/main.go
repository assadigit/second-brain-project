// Command mkzip builds the deterministic Windows release zip (doc 10 §2 —
// the port of the reference's tools/build_zip.py laws; T-65's minimal
// slice: the CLI-milestone composition. T-80/Phase 7 owns the full
// builder — the double-build reproducibility ritual + the two-exe
// composition + the release automation — per traceability F4).
//
// Composition (doc 10 §6's CLI slice, per SESSION-031's record):
//
//	GitCurator/GitCurator-cli.exe        the -exe input (built beforehand:
//	                                    GOOS=windows GOARCH=amd64
//	                                    CGO_ENABLED=0 go build -trimpath
//	                                    -ldflags "-X main.version=…")
//	GitCurator/<launcher>.bat|.sh       the CLI launcher set, tracked
//	                                    directly under cmd/gitcurator-cli/,
//	                                    FLATTENED to the zip root (the
//	                                    reference's v0.81.0 launcher-
//	                                    flattening law)
//	GitCurator/GitCurator.bat           the GUI-side pair at the repo
//	GitCurator/GitCurator-DRY-RUN.bat   root (T-65; the DRY-RUN one runs
//	                                    GitCurator-cli.exe --cli --auto
//	                                    --yes --dry-run — the safe first
//	                                    run; GitCurator.bat targets the
//	                                    GUI-milestone exe)
//	GitCurator/config.example.json      the settings template (the
//	                                    reference's own bytes)
//	GitCurator/WINDOWS-QUICKSTART.md    the first-run guide
//	GitCurator/VERSION                  STAMPED from -version (never
//	                                    copied as a tracked file)
//
// Determinism laws (build_zip.py:126-217, byte-for-byte discipline):
//   - fixed entry timestamps: 2026-09-29 00:00:00 UTC, every entry;
//   - sorted arcnames;
//   - 0o644 external attrs;
//   - DEFLATE level 9;
//   - tracked-files-only: every non-exe input must be git-tracked
//     (git ls-files — never local junk);
//   - the banned-secrets guard: the build REFUSES to run if any banned
//     file (config.json, config.local.json, installer.config.json,
//     cache.db, error_outbox.db, *.session*, .env, .env.local) is
//     tracked in the shipped surface (defense in depth against an
//     accidental `git add -f`);
//   - every .bat: pure ASCII (the build fails on non-ASCII — the
//     v0.09.1 codepage bug class) + CRLF (normalized idempotently).
//
// Same inputs + version → same bytes (sha256-stable): two builds from
// the same commit are byte-identical, so a released zip can always be
// reproduced from its tag.
//
// Usage:
//
//	go run ./build/mkzip -version "$(cat VERSION)" \
//	    -exe bin/GitCurator-cli.exe \
//	    -out reports/dist/GitCurator-v0.92.0-go.1-windows.zip
//
// Exit codes: 0 = built; 2 = fatal (bad inputs, banned file tracked,
// non-ASCII .bat, write error).
package main

import (
	"archive/zip"
	"bytes"
	"compress/flate"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// zipRoot — the single root folder inside the zip.
const zipRoot = "GitCurator"

// fixedDateTime — every entry's timestamp (build byte-identity law).
var fixedDateTime = time.Date(2026, 9, 29, 0, 0, 0, 0, time.UTC)

// fixedExternalAttrs — 0o644 (the build_zip.py law).
const fixedExternalAttrs = 0o644 << 16

// launcherDir — the CLI launcher home (the owner's folder law: the
// launchers live with the CLI exe; they FLATTEN to the zip root).
const launcherDir = "cmd/gitcurator-cli"

// rootAssets — the repo-root files that ship (relative to the repo
// root, forward slashes).
var rootAssets = []string{
	"GitCurator.bat",
	"GitCurator-DRY-RUN.bat",
	"config.example.json",
	"WINDOWS-QUICKSTART.md",
}

// bannedExact — tracked paths (repo-root-relative) that kill the build.
var bannedExact = map[string]bool{
	"config.json":                true,
	"config.local.json":          true,
	"installer.config.json":      true,
	"cache.db":                   true,
	"error_outbox.db":            true,
	launcherDir + "/config.json": true,
	launcherDir + "/cache.db":    true,
	launcherDir + "/.env":        true,
	launcherDir + "/.env.local":  true,
}

// bannedSuffixes / bannedNames — per-name rules on the shipped surface
// (the repo root and the launcher dir).
var bannedSuffixes = []string{".session", ".session-journal"}
var bannedNames = map[string]bool{".env": true, ".env.local": true}

// entry — one zip member (arcname + bytes).
type entry struct {
	arcname string
	data    []byte
}

func main() {
	root := flag.String("root", ".", "the repo root (a git checkout)")
	version := flag.String("version", "", "the version stamp (required)")
	exe := flag.String("exe", "", "the built GitCurator-cli.exe path (required)")
	out := flag.String("out", "", "the output zip path (default reports/dist/GitCurator-v<version>-windows.zip)")
	flag.Parse()

	if strings.TrimSpace(*version) == "" {
		fatal("missing -version (usually: -version \"$(cat VERSION)\")")
	}
	if strings.TrimSpace(*exe) == "" {
		fatal("missing -exe (build it first: GOOS=windows GOARCH=amd64 " +
			"CGO_ENABLED=0 go build -trimpath -ldflags \"-X main.version=" +
			*version + "\" -o bin/GitCurator-cli.exe ./cmd/gitcurator-cli)")
	}
	outPath := *out
	if outPath == "" {
		outPath = filepath.Join("reports", "dist",
			fmt.Sprintf("GitCurator-v%s-windows.zip", strings.TrimSpace(*version)))
	}
	if err := run(*root, strings.TrimSpace(*version), *exe, outPath); err != nil {
		fatal(err.Error())
	}
}

func fatal(msg string) {
	fmt.Fprintf(os.Stderr, "[mkzip] fatal: %s\n", msg)
	os.Exit(2)
}

// run builds the zip. Deterministic: same inputs + version → same bytes.
func run(root, version, exePath, outPath string) error {
	tracked, err := collectTracked(root)
	if err != nil {
		return err
	}
	if err := guardBanned(tracked); err != nil {
		return err
	}

	entries := []entry{}

	// 1. The exe.
	exeData, err := os.ReadFile(exePath)
	if err != nil {
		return fmt.Errorf("read -exe: %w", err)
	}
	entries = append(entries, entry{zipRoot + "/GitCurator-cli.exe", exeData})

	// 2. The CLI launchers — tracked .bat/.sh directly under the
	//    launcher dir, FLATTENED to the zip root.
	for _, rel := range tracked {
		if !strings.HasPrefix(rel, launcherDir+"/") {
			continue
		}
		tail := rel[len(launcherDir)+1:]
		if tail == "" || strings.Contains(tail, "/") {
			continue
		}
		if !strings.HasSuffix(tail, ".bat") && !strings.HasSuffix(tail, ".sh") {
			continue
		}
		data, err := readTracked(root, rel, tracked)
		if err != nil {
			return err
		}
		data, err = normalizeLauncher(tail, data)
		if err != nil {
			return err
		}
		entries = append(entries, entry{zipRoot + "/" + tail, data})
	}

	// 3. The root assets.
	for _, rel := range rootAssets {
		data, err := readTracked(root, rel, tracked)
		if err != nil {
			return err
		}
		if strings.HasSuffix(rel, ".bat") {
			data, err = normalizeLauncher(rel, data)
			if err != nil {
				return err
			}
		}
		entries = append(entries, entry{zipRoot + "/" + rel, data})
	}

	// 4. VERSION — stamped from the flag, never copied as a tracked file.
	entries = append(entries, entry{zipRoot + "/VERSION",
		[]byte(version + "\n")})

	// 5. The determinism laws: sorted arcnames, fixed timestamps,
	//    fixed attrs, DEFLATE level 9.
	sort.Slice(entries, func(i, j int) bool {
		return entries[i].arcname < entries[j].arcname
	})
	if err := writeZip(entries, outPath); err != nil {
		return err
	}

	totalIn := 0
	for _, e := range entries {
		totalIn += len(e.data)
	}
	sha := sha256File(outPath)
	fi, _ := os.Stat(outPath)
	fmt.Printf("[mkzip] %s\n", outPath)
	fmt.Printf("[mkzip] %d files, %d bytes in, %d bytes zipped\n",
		len(entries), totalIn, fi.Size())
	fmt.Printf("[mkzip] sha256 %s\n", sha)
	return nil
}

// collectTracked returns the repo-root-relative tracked paths (git
// ls-files — only committed files, never local junk), sorted.
func collectTracked(root string) ([]string, error) {
	proc, err := exec.Command("git", "-C", root,
		"ls-files", "-z").Output()
	if err != nil {
		return nil, fmt.Errorf("git ls-files failed (git installed? "+
			"%q a git checkout?): %w", root, err)
	}
	paths := []string{}
	for _, raw := range bytes.Split(proc, []byte{0}) {
		if len(raw) > 0 {
			paths = append(paths, string(raw))
		}
	}
	sort.Strings(paths)
	return paths, nil
}

// guardBanned — defense in depth: refuse to build if any banned file
// is tracked in the shipped surface (the root or the launcher dir).
func guardBanned(tracked []string) error {
	for _, rel := range tracked {
		if bannedExact[rel] {
			return fmt.Errorf("banned file tracked at %q (secrets / live "+
				"state never ship; remove it from git)", rel)
		}
		atRoot := !strings.Contains(rel, "/")
		inLauncherDir := strings.HasPrefix(rel, launcherDir+"/")
		if !atRoot && !inLauncherDir {
			continue
		}
		name := rel
		if i := strings.LastIndex(rel, "/"); i >= 0 {
			name = rel[i+1:]
		}
		if bannedNames[name] {
			return fmt.Errorf("banned file tracked at %q", rel)
		}
		for _, suf := range bannedSuffixes {
			if strings.HasSuffix(name, suf) {
				return fmt.Errorf("banned file tracked at %q", rel)
			}
		}
	}
	return nil
}

// readTracked reads a tracked file, refusing anything not in the
// tracked set (the tracked-files-only law).
func readTracked(root, rel string, tracked []string) ([]byte, error) {
	found := false
	for _, t := range tracked {
		if t == rel {
			found = true
			break
		}
	}
	if !found {
		return nil, fmt.Errorf("%q is not git-tracked (tracked-files-only "+
			"law — commit it first)", rel)
	}
	return os.ReadFile(filepath.Join(root, filepath.FromSlash(rel)))
}

// normalizeLauncher enforces the .bat laws: pure ASCII (fail on
// non-ASCII — the v0.09.1 codepage bug class) + CRLF, normalized
// idempotently. The .sh files pass through untouched.
func normalizeLauncher(name string, data []byte) ([]byte, error) {
	if !strings.HasSuffix(name, ".bat") {
		return data, nil
	}
	for i, b := range data {
		if b > 0x7F {
			return nil, fmt.Errorf("%s: byte %d is non-ASCII (0x%02X) — "+
				"the codepage bug class can never ship", name, i, b)
		}
	}
	text := string(data)
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	lines := strings.Split(text, "\n")
	// A trailing newline keeps the final line from gluing onto the next.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	return []byte(strings.Join(lines, "\r\n") + "\r\n"), nil
}

// writeZip writes the archive with the fixed metadata laws.
func writeZip(entries []entry, outPath string) error {
	if err := os.MkdirAll(filepath.Dir(outPath), 0o755); err != nil {
		return err
	}
	f, err := os.Create(outPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	zw.RegisterCompressor(zip.Deflate, func(out io.Writer) (io.WriteCloser, error) {
		return flate.NewWriter(out, flate.BestCompression)
	})
	for _, e := range entries {
		hdr := &zip.FileHeader{
			Name:           e.arcname,
			Method:         zip.Deflate,
			Modified:       fixedDateTime,
			ExternalAttrs:  fixedExternalAttrs,
			CreatorVersion: 20, // fixed — no environment leakage
			ReaderVersion:  20,
		}
		w, err := zw.CreateHeader(hdr)
		if err != nil {
			return err
		}
		if _, err := w.Write(e.data); err != nil {
			return err
		}
	}
	return zw.Close()
}

func sha256File(path string) string {
	h := sha256.New()
	f, err := os.Open(path)
	if err != nil {
		return ""
	}
	defer f.Close()
	if _, err := io.Copy(h, f); err != nil {
		return ""
	}
	return fmt.Sprintf("%x", h.Sum(nil))
}
