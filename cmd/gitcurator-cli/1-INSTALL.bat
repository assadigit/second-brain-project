@echo off
REM ============================================================
REM  GitCurator 1-INSTALL.bat - run me ONCE after unzipping.
REM  The Go app is a native exe - no Python, no .venv, no packages
REM  to install. The two things this file CHECKS:
REM    1. git on PATH (the vault seal, the good-repos scan and the
REM       vault mirror all shell out to it) - a hard requirement;
REM    2. the WebView2 Runtime (the desktop app's rendering engine;
REM       a READ-ONLY registry probe, per-user first then
REM       per-machine). If it is missing, nothing is broken: the
REM       desktop app offers the one-click download at its first
REM       launch.
REM  This file only CHECKS and points - it never writes the
REM  registry, never touches PATH, never installs anything itself.
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

REM ---- the desktop app's runtime check (the SESSION-063 bootstrapper
REM leg): a READ-ONLY probe of the WebView2 Evergreen Runtime's own
REM registry key - per-user first, then per-machine. This file never
REM installs anything itself: when the runtime is missing, the desktop
REM app offers the one-click install at its first launch.
set "WV2VER="
for /f "tokens=2,*" %%a in ('reg query "HKCU\Software\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" /v pv 2^>nul ^| findstr /i "REG_SZ"') do set "WV2VER=%%b"
if not defined WV2VER for /f "tokens=2,*" %%a in ('reg query "HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}" /v pv 2^>nul ^| findstr /i "REG_SZ"') do set "WV2VER=%%b"
echo.
if defined WV2VER (
    echo  [GitCurator] Using the WebView2 Runtime:
    echo                found ^(version %WV2VER%^) - the desktop app is ready.
) else (
    echo  [GitCurator] The WebView2 Runtime was not found on this computer.
    echo                The desktop app ^(GitCurator.exe^) needs it. Nothing is
    echo                broken: at the first launch of the desktop app it
    echo                offers to download and install it automatically
    echo                ^(one internet connection, then never again^). You can
    echo                also install it now from the official page:
    echo                https://developer.microsoft.com/en-us/microsoft-edge/webview2/
    echo                then double-click 1-INSTALL.bat again to re-check.
)
echo.
echo  ------------------------------------------------------------
if defined WV2VER (
    echo  [GitCurator] Everything the app needs is in place.
) else (
    echo  [GitCurator] git is ready. The desktop app will offer the
    echo                WebView2 Runtime install at its first launch.
)
echo                Next: double-click GitCurator.bat
echo                (or GitCurator-SYNC.bat for the fast sync lane).
echo  ------------------------------------------------------------
echo.
pause
endlocal
