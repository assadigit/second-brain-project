package cli

// terminal_test.go — the paint box's behavior tests beyond the golden's
// byte shapes: the NO_COLOR/TTY color door, term_width's COLUMNS law,
// the StatusLine's erase/repaint grammar, and the rune helpers.

import (
	"bytes"
	"strings"
	"testing"
)

// TestEnableColorsDoor — NO_COLOR wins over TTY, force wins over
// everything, --no-color blanks every code.
func TestEnableColorsDoor(t *testing.T) {
	t.Setenv("NO_COLOR", "1")
	EnableColors(false)
	if ColorsEnabled() {
		t.Errorf("NO_COLOR: colors on, want off")
	}
	EnableColors(true)
	if !ColorsEnabled() {
		t.Errorf("force: colors off, want on")
	}
	if cBold == "" || cReset == "" {
		t.Errorf("force: codes stayed empty")
	}
	DisableColors()
	if ColorsEnabled() || cBold != "" {
		t.Errorf("--no-color: codes survived")
	}
}

// TestTermWidthColumnsLaw — a positive COLUMNS wins; junk falls back;
// a piped stdout falls back to 80; the 48 floor applies.
func TestTermWidthColumnsLaw(t *testing.T) {
	t.Setenv("COLUMNS", "100")
	if got := termWidth(); got != 100 {
		t.Errorf("COLUMNS=100: termWidth = %d, want 100", got)
	}
	t.Setenv("COLUMNS", "40")
	if got := termWidth(); got != 48 {
		t.Errorf("COLUMNS=40: termWidth = %d, want the 48 floor", got)
	}
	t.Setenv("COLUMNS", "junk")
	if got := termWidth(); got != 80 {
		t.Errorf("COLUMNS=junk: termWidth = %d, want the 80 fallback", got)
	}
	t.Setenv("COLUMNS", "")
	if got := termWidth(); got != 80 {
		t.Errorf("piped stdout: termWidth = %d, want 80", got)
	}
}

// TestRuleCap64 — the divider caps at 64 columns even on wide terms.
func TestRuleCap64(t *testing.T) {
	t.Setenv("COLUMNS", "200")
	got := Rule()
	if n := len([]rune(got)); n != 64 {
		t.Errorf("rule = %d runes, want 64", n)
	}
}

// TestStatusLineGrammar — start/tick paints the spinner line with \r
// + ljust; log erases first; stop erases + prints the final line.
func TestStatusLineGrammar(t *testing.T) {
	t.Setenv("COLUMNS", "")
	buf := &bytes.Buffer{}
	c := &CLI{Out: buf, In: strings.NewReader(""), Now: frozenClock}
	s := NewStatusLine(c)
	s.Start("Fetching bot queue", 0)
	first := buf.String()
	if !strings.HasPrefix(first, "\r") {
		t.Errorf("tick did not start with \\r: %q", first)
	}
	if !strings.Contains(first, "⠋ Fetching bot queue") {
		t.Errorf("tick lost the spinner + label: %q", first)
	}
	// the painted line is ljust'ed to term_width-1 RUNES (79; the
	// spinner glyph is 3 bytes, so the byte count is higher)
	if n := len([]rune(first)); n != 1+79 {
		t.Errorf("tick line = %d runes, want 80 (\\r + 79)", n)
	}
	buf.Reset()
	s.SetProgress(3, 4)
	s.SetRepo("owner/repo")
	s.Tick()
	second := buf.String()
	if !strings.Contains(second, "3/4 (75%)") {
		t.Errorf("tick lost the progress bar: %q", second)
	}
	if !strings.Contains(second, "owner/repo") {
		t.Errorf("tick lost the repo label: %q", second)
	}
	if !strings.Contains(second, "⠙") {
		t.Errorf("tick did not advance the spinner frame: %q", second)
	}
	buf.Reset()
	s.Log("a log line", "warning")
	logged := buf.String()
	// erase first: \r + 79 spaces + \r, then the log line, then the
	// repaint (\r + the spinner line again)
	lr := []rune(logged)
	if lr[0] != '\r' || lr[80] != '\r' {
		t.Errorf("log did not erase first (runes 0/80): %q", logged)
	}
	if !strings.Contains(logged, "▲ a log line") {
		t.Errorf("log lost the level icon: %q", logged)
	}
	if !strings.Contains(logged, "\r⠹ Fetching bot queue") {
		t.Errorf("log did not repaint underneath: %q", logged)
	}
	buf.Reset()
	s.Stop("done line")
	stopped := buf.String()
	sr := []rune(stopped)
	if sr[0] != '\r' || sr[80] != '\r' {
		t.Errorf("stop did not erase (runes 0/80): %q", stopped)
	}
	if !strings.HasSuffix(stopped, "done line\n") {
		t.Errorf("stop lost the final line: %q", stopped)
	}
	// after stop, ticks are inert
	buf.Reset()
	s.Tick()
	if buf.Len() != 0 {
		t.Errorf("tick after stop painted: %q", buf.String())
	}
}

// TestStatusLineBarWidth — the total rides render_bar (the 20-wide
// default).
func TestStatusLineBarWidth(t *testing.T) {
	t.Setenv("COLUMNS", "")
	buf := &bytes.Buffer{}
	c := &CLI{Out: buf, In: strings.NewReader(""), Now: frozenClock}
	s := NewStatusLine(c)
	s.Start("Processing batch", 1)
	s.SetProgress(1, 1)
	buf.Reset()
	s.Tick()
	if !strings.Contains(buf.String(), "1/1 (100%)") {
		t.Errorf("completed bar missing: %q", buf.String())
	}
}

// TestPyLjustAndSlice — the rune-exact helpers.
func TestPyLjustAndSlice(t *testing.T) {
	if got := pyLjust("••", 5); got != "••   " {
		t.Errorf("pyLjust runes = %q", got)
	}
	if got := pyLjust("abcdef", 3); got != "abcdef" {
		t.Errorf("pyLjust overflow = %q", got)
	}
	if got := pySliceRunes("héllo wörld", 0, 5); got != "héllo" {
		t.Errorf("pySliceRunes = %q", got)
	}
	if got := pySliceRunes("abc", 2, 99); got != "c" {
		t.Errorf("pySliceRunes clamp = %q", got)
	}
	if got := pySliceRunes("abc", 5, 2); got != "" {
		t.Errorf("pySliceRunes inverted = %q", got)
	}
}

// TestStripANSITable — the SGR-only law (other escapes survive).
func TestStripANSITable(t *testing.T) {
	cases := []struct{ in, want string }{
		{"", ""},
		{"plain text", "plain text"},
		{"\x1b[1mbold\x1b[0m", "bold"},
		{"\x1b[31;1mred\x1b[0m tail", "red tail"},
		{"keep \x1b[K cursor", "keep \x1b[K cursor"},
	}
	for _, tc := range cases {
		if got := StripANSI(tc.in); got != tc.want {
			t.Errorf("StripANSI(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// TestRenderBarBankersRounding — Python round()'s half-to-even at the
// exact 0.5 fill (1/40 → 20*0.025 = 0.5 → 0 blocks).
func TestRenderBarBankersRounding(t *testing.T) {
	got := RenderBar(1, 40, 20)
	if !strings.HasPrefix(got, "░") {
		t.Errorf("render_bar(1,40) filled blocks: %q", got)
	}
	if !strings.Contains(got, "1/40 (2%)") {
		t.Errorf("render_bar(1,40) counts: %q", got)
	}
	got2 := RenderBar(3, 40, 20) // 1.5 → 2 blocks
	if !strings.HasPrefix(got2, "██") || strings.HasPrefix(got2, "███") {
		t.Errorf("render_bar(3,40) fill: %q", got2)
	}
}
