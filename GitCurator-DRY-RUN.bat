@echo off
REM ============================================================
REM  GitCurator-DRY-RUN.bat - the SAFE first run.
REM  The same fully-automatic run as GitCurator-CLI.bat, but it
REM  writes NOTHING: no notes, no vault changes, no backup pushes,
REM  no queue consumption. Everything it WOULD have done is
REM  printed so you can read the plan first.
REM
REM  Needs config first: double-click GitCurator-CLI-Setup.bat
REM  (the desktop app arrives with the GUI milestone).
REM ============================================================
setlocal
title GitCurator - dry run (writes nothing)
cd /d "%~dp0"

REM PURE ASCII, never changes the codepage (the v0.09.1 lesson).

if not exist "GitCurator-cli.exe" (
    echo.
    echo  [GitCurator] GitCurator-cli.exe was not found next to this
    echo                file - run me from the unzipped folder.
    echo.
    pause
    exit /b 1
)

:run
GitCurator-cli.exe --cli --auto --yes --dry-run
echo.
echo  [GitCurator] Dry run finished - nothing was written. Read the plan
echo                above; when you are happy with it, the real run is
echo                GitCurator-CLI.bat.
echo.
pause
endlocal
