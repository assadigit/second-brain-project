@echo off
REM ============================================================
REM  GitCurator 1-INSTALL.bat - run me ONCE after unzipping.
REM  The Go app is a native exe - no Python, no .venv, no packages
REM  to install. The only tool it needs from your system is git
REM  (the vault seal, the good-repos scan and the vault mirror all
REM  shell out to it). This file only CHECKS that git is reachable
REM  and tells you what to do next - it never writes the registry,
REM  never touches PATH, never installs anything itself.
REM
REM  After this: double-click GitCurator.bat to start the app.
REM ============================================================
setlocal
title GitCurator - first install
cd /d "%~dp0"

REM v0.92.0-go.1: PURE ASCII, never changes the codepage (the v0.09.1
REM lesson - see GitCurator-CLI.bat's header).

echo.
where git >nul 2>nul
if errorlevel 1 (
    echo.
    echo  [GitCurator] git.exe was not found on this computer.
    echo.
    echo                GitCurator needs git for the vault seal, the good-repos
    echo                scan and the vault mirror.
    echo                1. Install it from https://git-scm.com/download/win
    echo                2. IMPORTANT: tick "Git from the command line and also
    echo                   from 3rd-party software" on the installer's
    echo                   adjusting-screen (that is the PATH option)
    echo                3. Double-click 1-INSTALL.bat again
    echo.
    pause
    exit /b 1
)

echo  [GitCurator] Using git:
git --version
echo.

echo  ------------------------------------------------------------
echo  [GitCurator] Everything the app needs is in place.
echo                Next: double-click GitCurator.bat
echo                (or GitCurator-SYNC.bat for the fast sync lane).
echo  ------------------------------------------------------------
echo.
pause
endlocal
