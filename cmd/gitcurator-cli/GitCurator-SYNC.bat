@echo off
REM ============================================================
REM  GitCurator-SYNC.bat - THE FAST SYNC LANE (v0.81.0)
REM
REM  The simplistic door: check everything (config, connection,
REM  vaults) -> one Enter -> the bot queue syncs and the LLM
REM  curates every new item into the vaults. Nothing else.
REM
REM  If anything is not right, this window tells you to open the
REM  GUI (GitCurator.bat) and fix it there - the sync door never
REM  asks for credentials itself.
REM
REM  For the cron / Task Scheduler future:
REM     GitCurator-SYNC.bat --yes   (no Enter needed - auto-sync)
REM
REM  v0.92.0-go.1: the Go port - the app is the native
REM  GitCurator-cli.exe sitting next to this file; no Python, no
REM  .venv, no installs.
REM  v0.81.0: lives in gitcurator\cli\ now (the owner's folder law:
REM  everything CLI in the CLI folder); the zip still unpacks it at
REM  the top level, and the cd below finds the exe in both homes.
REM ============================================================
setlocal
title GitCurator SYNC - the fast lane
cd /d "%~dp0"
REM v0.81.0 - the folder law's walk-up: in the repo this file sits at
REM app\gitcurator\cli\, two levels under main.py; in the zip it sits
REM at the top, next to main.py. main.py IS the app root's landmark.
REM v0.92.0-go.1: the Go twin keeps the walk-up - VERSION is the
REM landmark (the exe embeds everything else the app needs).
if not exist "VERSION" cd /d "%~dp0\..\.."

REM v0.09.2: this file is PURE ASCII and never changes the codepage
REM (the v0.09.1 lesson - see GitCurator-CLI.bat's header).

REM ---- v0.92.0-go.1: run the native exe DIRECTLY - no Python
REM discovery, no .venv, no interpreter to find. ----
:run
echo.
GitCurator-cli.exe --cli --sync %*
set EXITCODE=%ERRORLEVEL%

echo.
if "%EXITCODE%"=="0" (
    echo  [GitCurator] Sync finished. You can close this window.
) else (
    echo  [GitCurator] Sync exited with code %EXITCODE%.
    echo                If a check failed, open the GUI ^(GitCurator.bat^),
    echo                fix the credentials and essentials in Settings,
    echo                then run this file again.
)
echo.
pause
endlocal
