@echo off
REM ============================================================
REM  GitCurator.bat - double-click to start the desktop app.
REM  (Run 1-INSTALL.bat once first.)
REM
REM  A console window opens BEHIND the app - that is normal; it
REM  shows the app's log lines. If something breaks, screenshot
REM  that window: it has the exact error.
REM ============================================================
setlocal
title GitCurator
cd /d "%~dp0"

REM PURE ASCII, never changes the codepage (the v0.09.1 lesson).

REM The desktop app is GitCurator.exe (the GUI milestone); the
REM command line (GitCurator-cli.exe) is fully usable today.

if not exist "GitCurator.exe" (
    echo.
    echo  [GitCurator] The desktop app (GitCurator.exe) is not in this
    echo                folder yet - it ships with the GUI milestone.
    echo                The command line is fully usable today: run
    echo                GitCurator-CLI-Setup.bat to configure, then
    echo                GitCurator-DRY-RUN.bat for a safe first run
    echo                ^(it writes nothing^).
    echo.
    pause
    exit /b 1
)

:run
GitCurator.exe
if errorlevel 1 (
    echo.
    echo  [GitCurator] The app exited with an error - the message is above.
    echo                Screenshot this window if you need help with it.
    pause
)
endlocal
