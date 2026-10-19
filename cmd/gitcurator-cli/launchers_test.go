package main

// launchers_test.go — T-63's byte laws for the launcher set that ships
// beside the exe (spec-cli's Shims+launchers section + doc 10 §2-6):
// the .bat set is ASCII-ONLY with CRLF line endings (BOTH asserted at
// byte level), the .sh is LF-only UTF-8, the grammar skeleton is the
// reference's own (setlocal → title → cd /d "%~dp0" → the walk-up →
// the exe DIRECTLY — no python discovery — → EXITCODE echo → pause →
// endlocal), and the 1-INSTALL equivalent checks git and touches
// nothing else (no registry writes, no PATH mutations).
//
// The reference-anchored differential (the golden's launchers section:
// the reference's own bytes byte-compared + the divergences asserted)
// lives in internal/cli/goldencli_test.go — these are the self-side
// laws that must hold wherever the files travel.

import (
	"bytes"
	"os"
	"strings"
	"testing"
)

var batFiles = []string{
	"GitCurator-CLI.bat",
	"GitCurator-SYNC.bat",
	"GitCurator-CLI-Setup.bat",
	"Start-GitCurator-CLI.bat",
	"1-INSTALL.bat",
}

func readLauncher(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(name)
	if err != nil {
		t.Fatalf("read launcher %s: %v (the launchers live beside main.go)", name, err)
	}
	return raw
}

// TestBatLaunchersASCIICRLF — the .bat laws: pure ASCII, every line
// ends CRLF (no bare LF anywhere), the file ends with a final CRLF.
func TestBatLaunchersASCIICRLF(t *testing.T) {
	for _, name := range batFiles {
		raw := readLauncher(t, name)
		for i, b := range raw {
			if b > 0x7F {
				t.Errorf("%s: byte %d is non-ASCII (0x%02X) — the v0.09.2 law", name, i, b)
			}
		}
		if n := bytes.Count(raw, []byte("\r\n")); n == 0 {
			t.Errorf("%s: no CRLF line endings at all", name)
		}
		if bytes.Contains(raw, []byte("\n")) && bytes.Count(raw, []byte("\n")) != bytes.Count(raw, []byte("\r\n")) {
			t.Errorf("%s: bare LF present (every newline must be CRLF)", name)
		}
		if bytes.Count(raw, []byte("\r")) != bytes.Count(raw, []byte("\r\n")) {
			t.Errorf("%s: bare CR present (a CR must always precede an LF)", name)
		}
		if !bytes.HasSuffix(raw, []byte("\r\n")) {
			t.Errorf("%s: the file must end with a final CRLF (the reference's own shape)", name)
		}
	}
}

// TestShLauncherLFOnly — the .sh laws: UTF-8, LF-only (no CR
// anywhere), executable bit set, the shebang first.
func TestShLauncherLFOnly(t *testing.T) {
	raw := readLauncher(t, "gitcurator-cli.sh")
	if bytes.Contains(raw, []byte("\r")) {
		t.Errorf("gitcurator-cli.sh: CR byte present — Unix launchers are LF-only")
	}
	if !bytes.HasPrefix(raw, []byte("#!/usr/bin/env bash\n")) {
		t.Errorf("gitcurator-cli.sh: the shebang line is not first")
	}
	fi, err := os.Stat("gitcurator-cli.sh")
	if err != nil {
		t.Fatal(err)
	}
	if fi.Mode()&0111 == 0 {
		t.Errorf("gitcurator-cli.sh: the executable bit is not set")
	}
}

// launcherLines splits a launcher's bytes into lines (CRLF- or
// LF-terminated; the terminator is stripped).
func launcherLines(raw []byte) []string {
	s := strings.ReplaceAll(string(raw), "\r\n", "\n")
	return strings.Split(s, "\n")
}

// hasLine answers whether the exact line is present (full-line match,
// not substring — REM prose can never satisfy a command expectation).
func hasLine(lines []string, want string) bool {
	for _, l := range lines {
		if l == want {
			return true
		}
	}
	return false
}

// The python-discovery command lines: the reference's own ladder
// (venv → python → py -3, with the 1-INSTALL.bat checks) — every one
// of them must be ABSENT from the Go launchers (the exe runs direct).
var pythonDiscoveryLines = []string{
	`if exist ".venv\Scripts\python.exe" goto :run`,
	`set "PY=python"`,
	`set "PY=py -3"`,
	`where python >nul 2>nul`,
	`%PY% --version >nul 2>nul`,
	`python -c "import sys; sys.exit(0 if sys.version_info>=(3,10) else 1)" >nul 2>nul`,
	`py -3 -c "import sys; sys.exit(0 if sys.version_info>=(3,10) else 1)" >nul 2>nul`,
	`%PY% -m venv .venv`,
	`".venv\Scripts\python.exe" -m pip install --upgrade pip`,
	`".venv\Scripts\python.exe" -m pip install -r requirements.txt`,
	`python --version`,
	`py --version`,
}

// TestBatGrammarSkeleton — every .bat follows the reference's own
// skeleton: @echo off → setlocal → title … → cd /d "%~dp0" → the
// walk-up `if not exist "VERSION" cd /d "%~dp0\..\.."` → :run → echo.
// → the exe line → set EXITCODE=%ERRORLEVEL% → the exit card → pause
// → endlocal; and NONE of the python-discovery lines appear.
func TestBatGrammarSkeleton(t *testing.T) {
	type want struct {
		title    string
		runLine  string
		walkUp   bool
		exitCode bool // carries the EXITCODE echo card
	}
	cases := map[string]want{
		"GitCurator-CLI.bat": {
			title:    `title GitCurator CLI - automatic run`,
			runLine:  `GitCurator-cli.exe --cli --auto --yes`,
			walkUp:   true,
			exitCode: true,
		},
		"GitCurator-SYNC.bat": {
			title:    `title GitCurator SYNC - the fast lane`,
			runLine:  `GitCurator-cli.exe --cli --sync %*`,
			walkUp:   true,
			exitCode: true,
		},
		"GitCurator-CLI-Setup.bat": {
			title:    `title GitCurator CLI - first-run setup`,
			runLine:  `GitCurator-cli.exe --cli --init`,
			walkUp:   true,
			exitCode: true,
		},
		"Start-GitCurator-CLI.bat": {
			title:    `title GitCurator CLI`,
			runLine:  `GitCurator-cli.exe --cli --auto --yes %*`,
			walkUp:   true,
			exitCode: true,
		},
		"1-INSTALL.bat": {
			title:    `title GitCurator - first install`,
			runLine:  "", // no exe run — the checker only checks
			walkUp:   false,
			exitCode: false,
		},
	}
	for name, w := range cases {
		lines := launcherLines(readLauncher(t, name))
		if !hasLine(lines, "@echo off") {
			t.Errorf("%s: @echo off missing", name)
		}
		if !hasLine(lines, "setlocal") {
			t.Errorf("%s: setlocal missing", name)
		}
		if !hasLine(lines, w.title) {
			t.Errorf("%s: the title line %q missing", name, w.title)
		}
		if !hasLine(lines, `cd /d "%~dp0"`) {
			t.Errorf("%s: the cd line missing", name)
		}
		if w.walkUp {
			if !hasLine(lines, `if not exist "VERSION" cd /d "%~dp0\..\.."`) {
				t.Errorf("%s: the walk-up line (landmark VERSION) missing", name)
			}
		}
		if w.runLine != "" {
			if !hasLine(lines, ":run") {
				t.Errorf("%s: the :run label missing", name)
			}
			if !hasLine(lines, "echo.") {
				t.Errorf("%s: the pre-run echo. missing", name)
			}
			if !hasLine(lines, w.runLine) {
				t.Errorf("%s: the exe run line %q missing", name, w.runLine)
			}
		}
		if w.exitCode {
			if !hasLine(lines, `set EXITCODE=%ERRORLEVEL%`) {
				t.Errorf("%s: the EXITCODE capture missing", name)
			}
			if !hasLine(lines, `    echo  [GitCurator] Finished. You can close this window.`) &&
				!hasLine(lines, `    echo  [GitCurator] Sync finished. You can close this window.`) &&
				!hasLine(lines, `    echo  [GitCurator] Setup complete! Next: double-click GitCurator-SYNC.bat`) {
				t.Errorf("%s: the success card line missing", name)
			}
			if !hasLine(lines, `    echo  [GitCurator] Exited with code %EXITCODE%.`) &&
				!hasLine(lines, `    echo  [GitCurator] Sync exited with code %EXITCODE%.`) &&
				!hasLine(lines, `    echo  [GitCurator] Setup exited with code %EXITCODE%.`) {
				t.Errorf("%s: the failure card line missing", name)
			}
		}
		if !hasLine(lines, "pause") {
			t.Errorf("%s: pause missing (the double-click law)", name)
		}
		if !hasLine(lines, "endlocal") {
			t.Errorf("%s: endlocal missing (the setlocal pairing law)", name)
		}
		for _, banned := range pythonDiscoveryLines {
			if hasLine(lines, banned) {
				t.Errorf("%s: python-discovery line leaked into the Go launcher: %q", name, banned)
			}
		}
		if hasLine(lines, "chcp 65001") {
			t.Errorf("%s: chcp 65001 leaked (the v0.09.1 lesson — the exe hardens its own console)", name)
		}
	}
}

// TestInstallBatChecksGitOnly — the 1-INSTALL equivalent: git.exe on
// PATH + the next-steps card; NO registry writes, NO PATH mutations,
// NO installs (the Go app is a native exe — nothing to install).
func TestInstallBatChecksGitOnly(t *testing.T) {
	lines := launcherLines(readLauncher(t, "1-INSTALL.bat"))
	if !hasLine(lines, "where git >nul 2>nul") {
		t.Errorf("the git presence check (where git) missing")
	}
	if !hasLine(lines, "git --version") {
		t.Errorf("the git version print missing")
	}
	if !hasLine(lines, `echo                Next: double-click GitCurator.bat`) {
		t.Errorf("the next-steps card line missing")
	}
	bannedSubstrings := []string{
		"reg add", "reg delete", "reg query", "setx ",
		"set PATH=", "set \"PATH=", "%PATH%;", "set PY=",
		"-m venv", "pip install", "msiexec", "powershell",
	}
	text := string(readLauncher(t, "1-INSTALL.bat"))
	for _, banned := range bannedSubstrings {
		if strings.Contains(text, banned) {
			t.Errorf("the checker must never mutate the system — found %q", banned)
		}
	}
}

// TestShLauncherGrammar — set -euo pipefail, the walk-up (landmark
// VERSION), the exe guard, and the reference's own default
// (--cli --auto --yes) + the flag pass-through.
func TestShLauncherGrammar(t *testing.T) {
	lines := launcherLines(readLauncher(t, "gitcurator-cli.sh"))
	for _, want := range []string{
		"set -euo pipefail",
		`cd "$(dirname "$0")"`,
		"[ -f VERSION ] || cd ../..",
		`if [ ! -x "./GitCurator-cli" ]; then`,
		`if [ "$#" -eq 0 ]; then`,
		`    exec ./GitCurator-cli --cli --auto --yes`,
		`fi`,
		`exec ./GitCurator-cli --cli "$@"`,
	} {
		if !hasLine(lines, want) {
			t.Errorf("gitcurator-cli.sh: line %q missing", want)
		}
	}
	for _, banned := range []string{
		"python3", "command -v python",
	} {
		for _, l := range lines {
			if strings.HasPrefix(l, "#") {
				continue // prose may name what the Go port dropped; commands may not
			}
			if strings.Contains(l, banned) {
				t.Errorf("gitcurator-cli.sh: python discovery leaked: %q", l)
			}
		}
	}
}

// TestLauncherRunLinesCarryCLISpelling — the reference's own run lines
// keep the --cli spelling verbatim (the exe strips it — main.py's
// routing law, ported to the main package): the launchers stay
// byte-close to the reference family.
func TestLauncherRunLinesCarryCLISpelling(t *testing.T) {
	for _, name := range []string{
		"GitCurator-CLI.bat", "GitCurator-SYNC.bat",
		"GitCurator-CLI-Setup.bat", "Start-GitCurator-CLI.bat",
	} {
		lines := launcherLines(readLauncher(t, name))
		found := false
		for _, l := range lines {
			if strings.HasPrefix(l, "GitCurator-cli.exe --cli ") {
				found = true
			}
		}
		if !found {
			t.Errorf("%s: no `GitCurator-cli.exe --cli …` run line", name)
		}
	}
}
