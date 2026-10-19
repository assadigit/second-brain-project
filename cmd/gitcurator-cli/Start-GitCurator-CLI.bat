@echo off
REM ============================================================
REM  GitCurator CLI - one-click launcher (v0.09, merged lineage)
REM  Double-click this file. It:
REM    1. uses the config.json sitting NEXT TO this file,
REM    2. runs the full pipeline automatically:
REM       pre-flight checks -> bot queue -> process -> Obsidian
REM       notes -> seal -> publish.
REM  (Kept from v0.08 for muscle memory - it now drives the SAME
REM   native engine as GitCurator-CLI.bat: GitCurator-cli.exe.)
REM  v0.92.0-go.1: the Go port - the app is the native
REM  GitCurator-cli.exe sitting next to this file; no Python, no
REM  .venv, no installs.
REM ============================================================
setlocal
title GitCurator CLI
cd /d "%~dp0"
REM v0.81.0 - the folder law's walk-up: in the repo this file sits at
REM app\gitcurator\cli\, two levels under main.py; in the zip it sits
REM at the top, next to main.py. main.py IS the app root's landmark.
REM v0.92.0-go.1: the Go twin keeps the walk-up - VERSION is the
REM landmark (the exe embeds everything else the app needs).
if not exist "VERSION" cd /d "%~dp0\..\.."

REM v0.09.2: this file is PURE ASCII and never changes the codepage.
REM v0.09.1 ran "chcp 65001" here; a codepage switch INSIDE a batch
REM file makes cmd.exe re-parse the file at a shifted byte offset (the
REM dash characters in the header comments are 3 bytes in UTF-8 but 1
REM in ANSI), so fragments of those REM lines ran as commands - the
REM "'tlocal' / 'Double-click' / 'pre-flight' is not recognized" error
REM wall before the app started. The Go exe handles all console
REM encoding itself at startup (SetConsoleOutputCP(65001) + UTF-8 std
REM streams), so the chcp was redundant anyway. Launchers stay
REM ASCII-only and codepage-free.

REM ---- v0.92.0-go.1: run the native exe DIRECTLY - no Python
REM discovery, no .venv, no interpreter to find. ----
:run
echo.
GitCurator-cli.exe --cli --auto --yes %*
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
