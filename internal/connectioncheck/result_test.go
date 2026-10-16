package connectioncheck

// result_test.go — the Result line shapes: render_line's icon/name/
// detail law and short_line's v0.37.0 compact one-line battery
// (path-like parentheticals dropped, facts kept, hard budget,
// mid-ellipsis, always exactly one line). Ported from the reference's
// TestShortLine + the render law.

import (
	"strings"
	"testing"
	"unicode/utf8"
)

func TestRenderLine(t *testing.T) {
	cases := []struct {
		r    Result
		want string
	}{
		{result("GitHub vault", LevelOK, "found"),
			"✅ GitHub vault — found"},
		{result("X", LevelWarn, "careful"), "⚠️ X — careful"},
		{result("X", LevelError, "broken"), "❌ X — broken"},
		{result("X", LevelInfo, "note"), "ℹ️ X — note"},
		{result("X", "bogus-level", "d"), "• X — d"},
		{result("", LevelOK, "d"), "✅ ? — d"},
	}
	for _, c := range cases {
		if got := RenderLine(c.r); got != c.want {
			t.Errorf("RenderLine(%+v) = %q, want %q", c.r, got, c.want)
		}
	}
}

func TestShortLinePathParensDroppedFactsKept(t *testing.T) {
	r := result("GitHub vault", "ok",
		"found · writable — ready to receive notes · "+
			"Obsidian vault (G:\\Docs\\GitHub Projects\\Obsidian Vault)")
	s := ShortLine(r)
	if !strings.Contains(s, "GitHub vault") {
		t.Errorf("name missing: %q", s)
	}
	if !strings.Contains(s, "found · writable") {
		t.Errorf("detail missing: %q", s)
	}
	if strings.Contains(s, "G:\\") {
		t.Errorf("pathy paren survived: %q", s)
	}
	if !strings.Contains(s, "Obsidian vault") {
		t.Errorf("fact paren dropped: %q", s)
	}
	// fact parens survive: rate limits are facts, not locations
	r2 := result("GitHub token", "ok",
		"valid — account assadigit (5000 req/h)")
	if !strings.Contains(ShortLine(r2), "(5000 req/h)") {
		t.Errorf("rate-limit fact dropped: %q", ShortLine(r2))
	}
}

func TestShortLineForwardSlashDrivePathsDropped(t *testing.T) {
	r := result("Websites vault", "ok",
		"found · writable — ready to receive notes · "+
			"Obsidian vault (G:/Docs/Documents/Obsidian Website Directory/)")
	s := ShortLine(r)
	if strings.Contains(s, "G:/") {
		t.Errorf("forward-slash drive path survived: %q", s)
	}
	if !strings.Contains(s, "found · writable") {
		t.Errorf("detail missing: %q", s)
	}
}

func TestShortLineLongDetailsMidEllipsizedToOneLine(t *testing.T) {
	r := result("Hint", "info", strings.Repeat("x", 200))
	s := ShortLine(r)
	// the budget is on the WHOLE visible line (icon + name + detail)
	if utf8.RuneCountInString(s) > shortLineBudget+4 {
		t.Errorf("line over budget: %d runes (%q)",
			utf8.RuneCountInString(s), s)
	}
	if !strings.Contains(s, "…") {
		t.Errorf("no mid-ellipsis: %q", s)
	}
	if strings.Contains(s, "\n") {
		t.Errorf("multi-line: %q", s)
	}
}

func TestShortLineEveryRealCheckFitsOneLine(t *testing.T) {
	// The owner's real screenshot lines (v0.37.0 pass) — every one of
	// them must collapse to a single budgeted line; long-named checks
	// get a tighter detail budget so the TOTAL line never clips.
	real := []struct{ name, detail string }{
		{"llama.cpp", "up @ http://127.0.0.1:8080 · model " +
			"'Ternary-Bonsai-2-7B-Instruct-vision'"},
		{"Account session",
			"no session file — the account is not logged in yet " +
				"(GUI: Test Telegram & GitHub · CLI: --login)"},
		{"GitHub token",
			"REJECTED (401 Bad credentials) — make a fresh token at " +
				"github.com/settings/tokens (classic, 'repo' scope)"},
		{"Proxy", "127.0.0.1:10808 reachable (socks5, 4 ms)"},
		{"Live connection",
			"connected — account login OK · bot queue readable " +
				"(12 link(s) waiting · newest message id 991234)"},
		{"Bot", "@githubfetcherbot configured — the queue is read " +
			"through your account session"},
	}
	for _, c := range real {
		s := ShortLine(result(c.name, "warn", c.detail))
		if strings.Contains(s, "\n") {
			t.Errorf("%s: multi-line %q", c.name, s)
		}
		if utf8.RuneCountInString(s) > shortLineBudget+4 {
			t.Errorf("%s: over budget %d (%q)", c.name,
				utf8.RuneCountInString(s), s)
		}
	}
}

func TestShortLineEmptyDetailJustName(t *testing.T) {
	if got := ShortLine(result("X", "ok", "")); got != "✅ X" {
		t.Errorf("empty detail: got %q, want %q", got, "✅ X")
	}
}

func TestShortLineAllPathyDetailKeepsOriginal(t *testing.T) {
	// the reference's guard: when dropping every pathy paren would
	// empty the text, the original stays (never an empty detail).
	r := result("X", "ok", "(G:\\only\\a\\path)")
	s := ShortLine(r)
	if !strings.Contains(s, "X") {
		t.Errorf("name missing: %q", s)
	}
	if s == "✅ X" {
		t.Errorf("all-pathy detail collapsed to empty: %q", s)
	}
}

func TestShortLineUnicodeBudgetIsRuneWise(t *testing.T) {
	// the budget counts runes, not bytes (é is 2 bytes, 1 rune)
	detail := strings.Repeat("é", 200)
	r := result("Hint", "info", detail)
	s := ShortLine(r)
	if utf8.RuneCountInString(s) > shortLineBudget+4 {
		t.Errorf("rune budget broken: %d", utf8.RuneCountInString(s))
	}
}
