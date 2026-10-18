package cli

// terminal.go — the CLI's paint box (cli_terminal.py, verbatim port).

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// ---------------------------------------------------------------------------
// The ANSI color system (respects NO_COLOR, non-TTY output and --no-color).
//
// C mirrors the reference's class attributes: the SGR codes are empty
// strings while colors are disabled. EnableColors mirrors C.enable()
// (force, NO_COLOR, TTY detection — the colorama/os.system Windows VT
// trick is unnecessary in Go: the console subsystem binary renders ANSI
// after hardenStdio's SetConsoleOutputCP).
// ---------------------------------------------------------------------------

var (
	cReset   = ""
	cBold    = ""
	cDim     = ""
	cItalic  = ""
	cRed     = ""
	cGreen   = ""
	cYellow  = ""
	cCyan    = ""
	cMagenta = ""
	cBlue    = ""
	cWhite   = ""
	cBGRed   = ""
	cBGGreen = ""

	cEnabled = false

	printMu sync.Mutex
)

// ColorsEnabled reports the C._ENABLED state.
func ColorsEnabled() bool { return cEnabled }

// DisableColors is the --no-color path: force off AND blank every code
// (cli_main's setattr loop).
func DisableColors() {
	cEnabled = false
	cReset, cBold, cDim, cItalic = "", "", "", ""
	cRed, cGreen, cYellow, cCyan = "", "", "", ""
	cMagenta, cBlue, cWhite = "", "", ""
	cBGRed, cBGGreen = "", ""
}

// EnableColors mirrors C.enable(force): honor NO_COLOR first, then
// stdout TTY detection (redirected/pipe output stays clean).
func EnableColors(force bool) {
	if force {
		cEnabled = true
	} else if os.Getenv("NO_COLOR") != "" {
		cEnabled = false
	} else if !isATTYStdout() {
		cEnabled = false
	} else {
		cEnabled = true
	}
	if cEnabled {
		cReset = "\x1b[0m"
		cBold = "\x1b[1m"
		cDim = "\x1b[2m"
		cItalic = "\x1b[3m"
		cRed = "\x1b[31m"
		cGreen = "\x1b[32m"
		cYellow = "\x1b[33m"
		cCyan = "\x1b[36m"
		cMagenta = "\x1b[35m"
		cBlue = "\x1b[34m"
		cWhite = "\x1b[37m"
		cBGRed = "\x1b[41m"
		cBGGreen = "\x1b[42m"
	} else {
		DisableColors()
	}
}

// Paint mirrors paint(): wrap text in the given ANSI codes (no-op when
// disabled or no code given).
func Paint(text string, codes ...string) string {
	any := false
	for _, c := range codes {
		if c != "" {
			any = true
			break
		}
	}
	if !any || !cEnabled {
		return text
	}
	var b strings.Builder
	for _, c := range codes {
		b.WriteString(c)
	}
	b.WriteString(text)
	b.WriteString(cReset)
	return b.String()
}

// levelStyle mirrors _LEVEL_STYLE + style_level: the level's color and
// icon; unknown levels render as info.
//
// GO NOTE (the dead-level-color quirk, pinned by the golden): the
// reference builds _LEVEL_STYLE at MODULE IMPORT, while C's codes are
// still the pre-enable empty strings — enable() mutates the class
// attributes but never the already-built dict, so the level COLORS are
// dead code in the reference (cli_print's icons and level text stay
// plain even with colors on; only the timestamp dims). The port freezes
// the same empty colors here — never the live codes.
func levelStyle(level string) (color, icon string) {
	switch level {
	case "info":
		return "", "•"
	case "success":
		return "", "✔"
	case "warning":
		return "", "▲"
	case "error":
		return "", "✖"
	case "dim":
		return "", "·"
	default:
		return "", "•"
	}
}

// stripANSIRe is _strip_ansi's pattern: one SGR escape.
var stripANSIRe = regexp.MustCompile("\x1b\\[[0-9;]*m")

// StripANSI mirrors _strip_ansi.
func StripANSI(s string) string {
	return stripANSIRe.ReplaceAllString(s, "")
}

// ---------------------------------------------------------------------------
// Terminal helpers
// ---------------------------------------------------------------------------

// termWidth mirrors term_width(): max(48, shutil.get_terminal_size(
// (80, 24)).columns) — the COLUMNS env var wins when a positive int,
// then the real stdout terminal size, then the (80, 24) fallback. Any
// failure answers 80.
func termWidth() int {
	columns := 0
	if v := os.Getenv("COLUMNS"); v != "" {
		if n, err := parseIntStrict(v); err == nil {
			columns = n
		}
	}
	if columns <= 0 {
		if w := stdoutWidth(); w > 0 {
			columns = w
		}
	}
	if columns <= 0 {
		columns = 80
	}
	if columns < 48 {
		return 48
	}
	return columns
}

// Mask mirrors mask(): keeps the first `keep` runes only.
func Mask(secret string, keep int) string {
	if secret == "" {
		return cDim + "(not set)" + cReset
	}
	r := []rune(secret)
	if len(r) <= keep {
		return strings.Repeat("•", len(r))
	}
	dots := len(r) - keep
	if dots > 24 {
		dots = 24
	}
	return string(r[:keep]) + strings.Repeat("•", dots)
}

// MaskToken mirrors _mask_token: ghp_••••••••DVb2 (head + tail kept).
func MaskToken(token string, keepHead, keepTail int) string {
	if token == "" {
		return cDim + "(not set)" + cReset
	}
	r := []rune(token)
	if len(r) <= keepHead+keepTail+2 {
		return strings.Repeat("•", len(r))
	}
	return string(r[:keepHead]) + strings.Repeat("•", 8) + string(r[len(r)-keepTail:])
}

// Rule mirrors rule(): the dim divider, capped at 64 columns.
func Rule() string {
	n := termWidth()
	if n > 64 {
		n = 64
	}
	return cDim + strings.Repeat("─", n) + cReset
}

// cliPrint mirrors cli_print: the level-colored, timestamped line
// printer (thread-safe; end is "\n" or ""). The timestamp dims when
// colors are on (the level text/icon never color — the dead-level-color
// quirk above).
func (c *CLI) cliPrint(text, level string, end string) {
	color, icon := levelStyle(level)
	ts := c.Now().Format("15:04:05")
	if cEnabled {
		ts = cDim + ts + cReset
	}
	body := Paint(text, "")
	if level != "info" {
		body = Paint(text, color)
	}
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Fprint(c.Out, ts+" "+Paint(icon, color)+" "+body+end)
}

// CliPrint is the one-line form (end="\n").
func (c *CLI) CliPrint(text, level string) { c.cliPrint(text, level, "\n") }

// bannerArt mirrors the art string in cli_terminal.py banner() byte for
// byte (raw string literal keeps the source shape 1:1).
const bannerArt = `
   ▄███████▄  ▄████████    ▄████████    ▄████████  ███▄▄▄▄      ▄████████
  ███    ███ ███    ███   ███    ███   ███    ███ ███▀▀▀██▄   ███    ███
  ███    ███ ███    ███   ███    ███   ███    █▀▀ ███    ███  ███    ███
 ▄███▄▄▄▄███▄███▄▄▄▄▄███▄ ███    ███  ▄███▄▄▄     ███    ███ ▄███▄▄▄▄███▄
▀▀▀▀▀▀▀▀▀▀▀ ███▀▀▀▀▀▀▀▀▀  ███    ███ ▀▀▀▀▀▀▀    ▀▀     ███▀ ▀▀▀▀▀   ███▀
▀███████████ ███    ███   ███    ███ ▀███████████ ▄██    ███ ▄██████████▄
  ███    ███ ███    ███   ███    ███          ███ ███    ███ ███    █████
  ███    ███ ███    ███ ████    ████ ▄███████████  ▀█████▀  ███    █████
`

// bannerVersionLine mirrors the second print of banner(): the stale
// "v0.09.5" stamp is a preserved quirk, not a bug to fix during the port.
const bannerVersionLine = "  GitHub Project Curator — CLI v0.09.5  ·  colored · animated · fully automatic"

// Banner mirrors banner(): the painted art, the version line, a blank
// line.
func (c *CLI) Banner() {
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Fprint(c.Out, Paint(bannerArt, cCyan, cBold))
	fmt.Fprint(c.Out, Paint("  GitHub Project Curator — CLI ", cBold)+
		Paint("v0.09.5", cMagenta)+
		Paint("  ·  colored · animated · fully automatic", cDim)+"\n")
	fmt.Fprint(c.Out, "\n")
}

// ---------------------------------------------------------------------------
// Live status line: spinner + progress bar (single owner of the bottom line)
// ---------------------------------------------------------------------------

// SpinFrames mirrors SPIN_FRAMES.
const SpinFrames = "⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏"

// RenderBar mirrors render_bar: ████████░░░░░░░░ 12/40 (30%).
// filled uses Python round() — banker's rounding
// (helpers.RoundHalfEven).
func RenderBar(current, total, width int) string {
	if total < 1 {
		total = 1
	}
	frac := float64(current) / float64(total)
	if frac < 0 {
		frac = 0
	}
	if frac > 1 {
		frac = 1
	}
	filled := int(helpers.RoundHalfEven(float64(width)*frac, 0))
	barTxt := strings.Repeat("█", filled) + strings.Repeat("░", width-filled)
	pct := int(frac * 100)
	color := cCyan
	if frac >= 1.0 {
		color = cGreen
	}
	return Paint(barTxt, color) + Paint(fmt.Sprintf(" %d/%d (%d%%)", current, total, pct), cBold)
}

// StatusLine mirrors cli_terminal.StatusLine: owns the terminal's
// bottom line while a phase runs. update/tick repaint with "\r"; log
// erases, prints a full log line, then repaints underneath.
type StatusLine struct {
	c       *CLI
	frame   int
	active  bool
	label   string
	current int
	total   int
	repo    string
}

// NewStatusLine builds one bound to the CLI's output.
func NewStatusLine(c *CLI) *StatusLine { return &StatusLine{c: c} }

// Start mirrors start().
func (s *StatusLine) Start(label string, total int) {
	s.label = label
	s.total = total
	s.current = 0
	s.repo = ""
	s.active = true
	s.Tick()
}

// Stop mirrors stop().
func (s *StatusLine) Stop(final string) {
	if s.active {
		s.erase()
	}
	s.active = false
	if final != "" {
		fmt.Fprint(s.c.Out, final+"\n")
	}
}

// SetProgress mirrors set_progress().
func (s *StatusLine) SetProgress(current, total int) {
	s.current, s.total = current, total
}

// SetRepo mirrors set_repo().
func (s *StatusLine) SetRepo(repo string) { s.repo = repo }

func (s *StatusLine) erase() {
	w := termWidth()
	fmt.Fprint(s.c.Out, "\r"+strings.Repeat(" ", w-1)+"\r")
}

// Tick mirrors tick(): one spinner frame (Python's slice on the PAINTED
// line — escapes count as characters — is preserved via rune slicing).
func (s *StatusLine) Tick() {
	if !s.active {
		return
	}
	frames := []rune(SpinFrames)
	frame := string(frames[s.frame%len(frames)])
	s.frame++
	spin := Paint(frame, cMagenta, cBold)
	line := spin + " " + s.label
	if s.total != 0 {
		line += "  " + RenderBar(s.current, s.total, 20)
	}
	if s.repo != "" {
		line += "  " + Paint(s.repo, cDim)
	}
	w := termWidth()
	line = pySliceRunes(line, 0, w-1)
	printMu.Lock()
	defer printMu.Unlock()
	fmt.Fprint(s.c.Out, "\r"+pyLjust(line, w-1))
}

// Log mirrors log(): a log line ABOVE the live status line.
func (s *StatusLine) Log(text, level string) {
	printMu.Lock()
	if s.active {
		s.erase()
	}
	printMu.Unlock()
	s.c.CliPrint(text, level)
	if s.active {
		s.Tick()
	}
}

// ---------------------------------------------------------------------------
// UTF-8 hardening (the _harden_stdio twin)
// ---------------------------------------------------------------------------

// hardenStdio mirrors _harden_stdio's Windows half: switch the console
// codepage to 65001 so the glyphs RENDER. The stdout/stderr
// errors="replace" half is Go-native (invalid UTF-8 in output never
// raises). Runs at CLI startup — before any print.
func hardenStdio() { setConsoleUTF8() }

// ---------------------------------------------------------------------------
// Small Python-shape helpers (rune-exact)
// ---------------------------------------------------------------------------

// pyLjust pads s to n runes (Python str.ljust).
func pyLjust(s string, n int) string {
	r := []rune(s)
	if len(r) >= n {
		return s
	}
	return s + strings.Repeat(" ", n-len(r))
}

// pySliceRunes slices s by runes (Python str[a:b], negative-safe only
// for the forms used here).
func pySliceRunes(s string, a, b int) string {
	r := []rune(s)
	if a < 0 {
		a = 0
	}
	if b > len(r) {
		b = len(r)
	}
	if a >= b {
		return ""
	}
	return string(r[a:b])
}

// parseIntStrict mirrors int(s) for plain decimal ASCII (the COLUMNS
// env parse inside shutil.get_terminal_size).
func parseIntStrict(s string) (int, error) {
	n := 0
	if s == "" {
		return 0, fmt.Errorf("empty")
	}
	for _, ch := range s {
		if ch < '0' || ch > '9' {
			return 0, fmt.Errorf("not an int")
		}
		n = n*10 + int(ch-'0')
	}
	return n, nil
}

var _ = time.Time{} // keep time imported for the seam docs
