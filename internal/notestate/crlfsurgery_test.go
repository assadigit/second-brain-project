package notestate

// crlfsurgery_test.go — SESSION-061: the note_state SURGERY path's CRLF
// round-trip differential, the SESSION-008 recorded remainder closed with
// proof. The golden (notestate_crlf_roundtrip.golden.json) is GENERATED
// by re-riding the reference's own test_phase3 selections (the
// _github_note fixture + the moved/edited selections) over CRLF-written
// notes — the Windows norm — through the UNMODIFIED Python app
// (testdiff/goldens/dump_notestate.py §5); this test re-rides the same
// fixture bytes through the port and byte-matches the recorded law.
//
// THE LAW the golden pins (settled at byte level against
// note_state.py/storage.py before any port change was made): every
// reference read of a note file is TEXT MODE — open(path, 'r',
// encoding='utf-8', errors='replace') — whose universal-newlines
// translation (\r\n and lone \r become \n on read) feeds the surgery,
// and atomic_write_text writes with newline='' (raw, no translation) —
// so a CRLF note that rides the moves-as-corrections surgery comes out
// ALL-LF (the CRLF is eaten by the read, never restored by the write),
// and every fingerprint the module stores (record_note's own file read,
// the baseline, apply's row refresh, classify's current read) is the
// fingerprint of the TRANSLATED content. A port read that skips the
// translation leaks \r's into the surgery (mixed-endings output where
// the reference writes all-LF) and stores RAW fingerprints where the
// reference stores translated ones — a cross-binary cache.db
// disagreement (a Python-baselined CRLF row would read as perpetually
// edited to a raw-reading twin).

import (
	"encoding/hex"
	"os"
	"path/filepath"
	"testing"

	"github.com/assadigit/second-brain-project/internal/constants"
)

type crlfRoundtripGolden struct {
	Schema                 int      `json:"schema"`
	InputNoteAHex          string   `json:"input_note_a_hex"`
	InputNoteBHex          string   `json:"input_note_b_hex"`
	EditedAppendHex        string   `json:"edited_append_hex"`
	RecordNoteFingerprint  string   `json:"record_note_fingerprint"`
	BaselineRowFingerprint string   `json:"baseline_row_fingerprint"`
	MovedOutputHex         string   `json:"moved_output_hex"`
	MovedRowFingerprint    string   `json:"moved_row_fingerprint"`
	MovedFromCategory      string   `json:"moved_from_category"`
	MovedToCategory        string   `json:"moved_to_category"`
	MovedCount             int      `json:"moved_count"`
	MovedApplied           int      `json:"moved_applied"`
	EditedSources          []string `json:"edited_sources"`
}

func decodeHex(t *testing.T, name, s string) []byte {
	t.Helper()
	b, err := hex.DecodeString(s)
	if err != nil {
		t.Fatalf("%s: bad hex: %v", name, err)
	}
	return b
}

func TestCRLFSurgeryRoundTripGolden(t *testing.T) {
	var g crlfRoundtripGolden
	decodeGolden(t, "notestate_crlf_roundtrip.golden.json", &g)
	if g.Schema != 1 {
		t.Fatalf("golden schema = %d; want 1 (regenerate goldens)", g.Schema)
	}
	inputA := decodeHex(t, "input_note_a_hex", g.InputNoteAHex)
	inputB := decodeHex(t, "input_note_b_hex", g.InputNoteBHex)
	appendB := decodeHex(t, "edited_append_hex", g.EditedAppendHex)

	tmp := t.TempDir()
	vault := filepath.Join(tmp, "gh")
	agents, _ := constants.CategoryFolder("Agents")
	scraping, _ := constants.CategoryFolder("Scraping")
	aPath := filepath.Join(vault, agents, "a.md")
	bPath := filepath.Join(vault, agents, "b.md")
	if err := os.MkdirAll(filepath.Dir(aPath), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(aPath, inputA, 0o666); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(bPath, inputB, 0o666); err != nil {
		t.Fatal(err)
	}

	// (a) RecordNote's own file read (content nil ⇒ the module reads the
	//     file, text mode in the reference) over the CRLF note — its own
	//     scratch DB so no stray row joins the classify flow below.
	probeDB, err := New(filepath.Join(tmp, "probe.db"))
	if err != nil {
		t.Fatal(err)
	}
	if err := probeDB.RecordNote(VaultGithub, "https://github.com/o/crlf-probe",
		aPath, nil, nil, nil, nil); err != nil {
		t.Fatal(err)
	}
	probeRow, err := probeDB.RowFor(VaultGithub, "https://github.com/o/crlf-probe")
	if err != nil || probeRow == nil {
		t.Fatalf("probe row_for: %v (%v)", probeRow, err)
	}
	if probeRow.Fingerprint != g.RecordNoteFingerprint {
		t.Errorf("record_note fingerprint over the CRLF file = %s; want %s "+
			"(the reference's own file read is TEXT MODE — the fingerprint "+
			"is of the translated content)", probeRow.Fingerprint,
			g.RecordNoteFingerprint)
	}
	if err := probeDB.Close(); err != nil {
		t.Fatal(err)
	}

	// (b) the baseline (test_phase3's own setUp shape) — the CRLF rows'
	// fingerprints are the translated ones.
	dbPath := filepath.Join(tmp, "cache.db")
	if _, err := RecordBaselineIfEmpty(VaultGithub, vault, dbPath, nil); err != nil {
		t.Fatal(err)
	}
	dbObj, err := New(dbPath)
	if err != nil {
		t.Fatal(err)
	}
	defer dbObj.Close()
	baseRow, err := dbObj.RowFor(VaultGithub, "https://github.com/o/repo1")
	if err != nil || baseRow == nil {
		t.Fatalf("baseline row_for: %v (%v)", baseRow, err)
	}
	if baseRow.Fingerprint != g.BaselineRowFingerprint {
		t.Errorf("baseline fingerprint over the CRLF file = %s; want %s",
			baseRow.Fingerprint, g.BaselineRowFingerprint)
	}

	// (c) the edited selection (the hand-written line appended, CRLF) +
	//     the moved selection (a.md → Tools/Scraping) — the reference's
	//     own test_phase3 moves.
	if err := os.WriteFile(bPath, append(inputB, appendB...), 0o666); err != nil {
		t.Fatal(err)
	}
	newA := filepath.Join(vault, scraping, "a.md")
	if err := os.MkdirAll(filepath.Dir(newA), 0o777); err != nil {
		t.Fatal(err)
	}
	if err := os.Rename(aPath, newA); err != nil {
		t.Fatal(err)
	}

	// (d) classify + apply — the same verdicts as the reference.
	changes, err := ClassifyChanges(VaultGithub, vault, dbPath, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(changes.Moved) != g.MovedCount {
		t.Fatalf("moved = %d; want %d", len(changes.Moved), g.MovedCount)
	}
	mv := changes.Moved[0]
	if mv.FromCategory == nil || *mv.FromCategory != g.MovedFromCategory {
		t.Errorf("moved from_category = %v; want %s", mv.FromCategory,
			g.MovedFromCategory)
	}
	if mv.ToCategory == nil || *mv.ToCategory != g.MovedToCategory {
		t.Errorf("moved to_category = %v; want %s", mv.ToCategory,
			g.MovedToCategory)
	}
	var editedSources []string
	for _, e := range changes.Edited {
		editedSources = append(editedSources, e.SourceURL)
	}
	if len(editedSources) != len(g.EditedSources) {
		t.Fatalf("edited sources = %v; want %v", editedSources, g.EditedSources)
	}
	for i := range g.EditedSources {
		if editedSources[i] != g.EditedSources[i] {
			t.Errorf("edited source[%d] = %s; want %s", i, editedSources[i],
				g.EditedSources[i])
		}
	}
	applied, err := ApplyCorrections(VaultGithub, vault, changes, dbObj, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if applied.MovedApplied != g.MovedApplied {
		t.Fatalf("moved_applied = %d; want %d", applied.MovedApplied,
			g.MovedApplied)
	}

	// (e) THE ROUND-TRIP BYTES — the all-LF law: the reference's surgery
	//     read is text mode and its write is raw, so the CRLF note comes
	//     out ALL-LF with the surgery applied. Byte-equality with the
	//     recorded reference output is the differential's own proof.
	out, err := os.ReadFile(newA)
	if err != nil {
		t.Fatal(err)
	}
	if got := hex.EncodeToString(out); got != g.MovedOutputHex {
		t.Errorf("the moved note's round-trip bytes diverge from the "+
			"reference:\n got (cr count %d): %q\nwant (cr count %d): %q",
			countCR(out), string(out),
			countCR(decodeHex(t, "moved_output_hex", g.MovedOutputHex)),
			string(decodeHex(t, "moved_output_hex", g.MovedOutputHex)))
	}

	// (f) apply's row refresh — the fingerprint of the translated +
	//     surgered content (the cross-binary cache.db contract).
	afterRow, err := dbObj.RowFor(VaultGithub, "https://github.com/o/repo1")
	if err != nil || afterRow == nil {
		t.Fatalf("post-apply row_for: %v (%v)", afterRow, err)
	}
	if afterRow.Fingerprint != g.MovedRowFingerprint {
		t.Errorf("post-apply fingerprint = %s; want %s (the reference's "+
			"own translated-content hash)", afterRow.Fingerprint,
			g.MovedRowFingerprint)
	}
}

func countCR(b []byte) int {
	n := 0
	for _, c := range b {
		if c == '\r' {
			n++
		}
	}
	return n
}
