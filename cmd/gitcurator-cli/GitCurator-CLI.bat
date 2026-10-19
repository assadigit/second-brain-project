@echo off
REM ============================================================
REM  GitCurator-CLI.bat - one double-click, fully automatic run
REM  (bot SYNC -> PROCESS -> VaultSeal -> Good Repos)
REM
REM  First time? Run GitCurator-CLI-Setup.bat once to save your
REM  credentials locally, then use this file for every run.
REM  v0.92.0-go.1: the Go port - the app is the native
REM  GitCurator-cli.exe sitting next to this file; no Python, no
REM  .venv, no installs (1-INSTALL.bat only checks that git is on
REM  PATH).
REM  v0.81.0: lives in gitcurator\cli\ now (the owner's folder law:
REM  everything CLI in the CLI folder); the zip still unpacks it at
REM  the top level, and the cd below finds the exe in both homes.
REM ============================================================
setlocal
title GitCurator CLI - automatic run
cd /d "%~dp0"
REM v0.81.0 - the folder law's walk-up: in the repo this file sits at
REM app\gitcurator\cli\, two levels under main.py; in the zip it sits
REM at the top, next to main.py. main.py IS the app root's landmark.
REM v0.92.0-go.1: the Go twin keeps the walk-up - VERSION is the
REM landmark (the exe embeds everything else the app needs).
if not exist "VERSION" cd /d "%~dp0\..\.."

REM v0.09.2: this file is PURE ASCII and never changes the codepage.
REM v0.09.1 ran "chcp 65001" here; a codepage switch INSIDE a batch
REM file makes cmd.exe re-parse the file at a shifted byte offset, so
REM fragments of the REM header ran as commands ("'GitCurator' is not
REM recognized" error wall). The Go exe sets its own console codepage
REM at startup (SetConsoleOutputCP(65001) + UTF-8 std streams), so the
REM chcp was redundant anyway. ASCII-only, codepage-free.

REM ---- v0.92.0-go.1: run the native exe DIRECTLY - no Python
REM discovery, no .venv, no interpreter to find. ----
:run
echo.
GitCurator-cli.exe --cli --auto --yes
set EXITCODE=%ERRORLEVEL%

echo.
if "%EXITCODE%"=="0" (
    echo  [GitCurator] Finished. You can close this window.
) else (
    echo  [GitCurator] Exited with code %EXITCODE%.
    echo                If credentials are missing, run GitCurator-CLI-Setup.bat first.
    echo                For the fast sync lane, run GitCurator-SYNC.bat.
)
echo.
pause
endlocal
