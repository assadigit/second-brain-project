// belt_test.go — T-47: the ported selections of test_localonly.py §1
// (the belt's pure pattern builders: the stem globs, the owner's
// extras, the gitignore escape, the belt folds, the full websites
// belt's disk discovery + config joins + the note-glob floor).
package directorysync

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNotePatterns_TheShippedStemsRideAsGlobs(t *testing.T) {
	pats := LocalOnlyNotePatterns(nil)
	for _, want := range []string{"CakhiaTV*.md", "cobalt*.md", "Flac_Downloader*.md"} {
		found := false
		for _, p := range pats {
			if p == want {
				found = true
			}
		}
		if !found {
			t.Fatalf("%s missing from %v", want, pats)
		}
	}
	// the six the round removed from the live site, no more no less
	if len(pats) != len(DefaultLocalOnlyStems) {
		t.Fatalf("pattern count: %d", len(pats))
	}
}

func TestNotePatterns_TheOwnersExtrasJoinWithoutDuplicating(t *testing.T) {
	pats := LocalOnlyNotePatterns([]string{"MyPrivate*.md", "cobalt*.md"})
	found := 0
	for _, p := range pats {
		if p == "MyPrivate*.md" || p == "cobalt*.md" {
			found++
		}
	}
	if found != 2 {
		t.Fatalf("extras did not join uniquely: %v", pats)
	}
	if len(pats) != len(DefaultLocalOnlyStems)+1 {
		t.Fatalf("pattern count: %d (%v)", len(pats), pats)
	}
}

func TestNotePatterns_TheDefaultStemsAreTheRoundsOwnList(t *testing.T) {
	want := []string{"CakhiaTV", "1Music_cc", "cobalt", "Prowlarr", "Max_Effect", "Flac_Downloader"}
	got := append([]string(nil), DefaultLocalOnlyStems...)
	sortStrings(got)
	sortStrings(want)
	if len(got) != len(want) {
		t.Fatalf("stems: %v", got)
	}
	for i := range got {
		if got[i] != want[i] {
			t.Fatalf("stems: %v", got)
		}
	}
}

func TestGitignoreEscape_PlainFolderNamesPassThrough(t *testing.T) {
	if got := gitignoreEscape("Torrent_Shadow_Libraries"); got != "Torrent_Shadow_Libraries" {
		t.Fatalf("escape: %q", got)
	}
}

func TestGitignoreEscape_GlobAndSpecialCharsAreEscaped(t *testing.T) {
	out := gitignoreEscape("a[weird]?name#")
	if strings.Contains(strings.ReplaceAll(out, `\[`, ""), "[") {
		t.Fatalf("an unescaped bracket survived: %q", out)
	}
	if !strings.HasPrefix(out, `a\[`) {
		t.Fatalf("escape shape: %q", out)
	}
}

func TestGitignoreEscape_ASlashInANameBecomesAnUnderscore(t *testing.T) {
	// 'Random / Curiosities' as a pattern would mean a PATH in
	// gitignore — the safe translation folds it to one folder.
	if got := gitignoreEscape("Random / Curiosities"); got != "Random _ Curiosities" {
		t.Fatalf("escape: %q", got)
	}
}

func TestBeltFolds_ThePrivateNameAndItsFolderFormFoldTogether(t *testing.T) {
	folds := beltFolds([]string{"Torrent & Shadow Libraries"})
	if !folds[strings.ToLower("torrent & shadow libraries")] {
		t.Fatalf("the name's fold is missing: %v", folds)
	}
	if !folds["torrent shadow libraries"] {
		t.Fatalf("the folder form's fold is missing: %v", folds)
	}
}

// beltVault mirrors TestPrivateExclusionPatterns._vault.
func beltVault(t *testing.T) string {
	t.Helper()
	tmp := t.TempDir()
	cat := filepath.Join(tmp, "Download_Resources")
	if err := os.MkdirAll(filepath.Join(cat, "Torrent_Shadow_Libraries"), 0o755); err != nil {
		t.Fatal(err)
	}
	for _, p := range []string{
		filepath.Join(cat, "CakhiaTV.md"),
		filepath.Join(cat, "Torrent_Shadow_Libraries", "Z-Library.md"),
	} {
		if err := os.WriteFile(p, []byte("x"), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	return tmp
}

func hasPattern(pats []string, want string) bool {
	for _, p := range pats {
		if p == want {
			return true
		}
	}
	return false
}

func TestPrivateExclusionPatterns_TheDiskFolderIsDiscoveredAndTheStemsRide(t *testing.T) {
	vault := beltVault(t)
	pats := PrivateExclusionPatterns(vault, nil)
	if !hasPattern(pats, "Torrent_Shadow_Libraries/") {
		t.Fatalf("the disk folder form is missing: %v", pats)
	}
	if !hasPattern(pats, "CakhiaTV*.md") {
		t.Fatalf("the stem glob is missing: %v", pats)
	}
	// the safe_filename form of the belt name joins too (a folder
	// renamed later still matches by name)
	if !hasPattern(pats, "Torrent_Shadow_Libraries/") {
		t.Fatalf("the name-derived form is missing: %v", pats)
	}
}

func TestPrivateExclusionPatterns_TheVaultsOwnLocalOnlyListJoins(t *testing.T) {
	vault := beltVault(t)
	cfgDir := filepath.Join(vault, "_system")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(cfgDir, DirectoryConfigFilename),
		[]byte(`{"local_only": ["OwnersOwn*.md"]}`), 0o644); err != nil {
		t.Fatal(err)
	}
	pats := PrivateExclusionPatterns(vault, nil)
	if !hasPattern(pats, "OwnersOwn*.md") {
		t.Fatalf("the vault's own pattern is missing: %v", pats)
	}
}

func TestPrivateExclusionPatterns_AConfigDeclaredPrivateSubcategoryIsDiscovered(t *testing.T) {
	vault := beltVault(t)
	// a private sub with a DIFFERENT on-disk spelling
	if err := os.MkdirAll(filepath.Join(vault, "Download_Resources", "Piracy_Wall"), 0o755); err != nil {
		t.Fatal(err)
	}
	cfgDir := filepath.Join(vault, "_system")
	if err := os.MkdirAll(cfgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	cfg := `{"categories": [{"name": "Download Resources", "subcategories": [
		{"name": "Piracy & Wall", "visibility": "private"}]}]}`
	if err := os.WriteFile(filepath.Join(cfgDir, DirectoryConfigFilename), []byte(cfg), 0o644); err != nil {
		t.Fatal(err)
	}
	pats := PrivateExclusionPatterns(vault, nil)
	joined := strings.Join(pats, "\n")
	if !strings.Contains(joined, "Piracy_Wall/") {
		t.Fatalf("the disk spelling is missing: %v", pats)
	}
	if !strings.Contains(joined, "Piracy_Wall/") {
		t.Fatalf("the safe_filename spelling is missing: %v", pats)
	}
}

func TestPrivateExclusionPatterns_AMissingVaultStillAnswersTheNoteGlobs(t *testing.T) {
	pats := PrivateExclusionPatterns("/no/such/vault", nil)
	if !hasPattern(pats, "cobalt*.md") {
		t.Fatalf("the note globs did not answer: %v", pats)
	}
}

func TestPrivateExclusionPatterns_NeverRaisesOnAJunkVault(t *testing.T) {
	// a file where a directory should be, a broken config — the
	// floor (the note globs) always answers
	tmp := t.TempDir()
	pats := PrivateExclusionPatterns(tmp, nil)
	if !hasPattern(pats, "Prowlarr*.md") {
		t.Fatalf("the floor did not answer: %v", pats)
	}
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
