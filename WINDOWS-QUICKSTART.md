# GitCurator — first run on Windows (v0.92.0-go.1, the CLI milestone)

Three steps and you're in. This is the Go port's CLI milestone: the
command line is fully usable today; the desktop app (GitCurator.exe)
arrives with the GUI milestone.

## 1. Install (once)

Double-click **`1-INSTALL.bat`**.

- It checks that **git** is on PATH (the only thing the app needs —
  no Python, no packages, nothing else is installed on your computer).
- **If it says git.exe was not found:** install git from
  <https://git-scm.com/download/win> (the "Git for Windows" setup,
  default options are fine), then double-click `1-INSTALL.bat` again.
- Safe to re-run any time.

## 2. Configure (the wizard)

Double-click **`GitCurator-CLI-Setup.bat`**.

It asks for the same things the app's Settings page will: your GitHub
token, the LLM (local Ollama / llama.cpp or any OpenAI-compatible
endpoint), and your vault paths — and saves them to `config.json`
(kept private, never shared).

## 3. Safe first run (writes NOTHING)

Double-click **`GitCurator-DRY-RUN.bat`**.

It performs the same fully-automatic run as the real thing — checks,
queue, processing — but every write is only **reported**. No notes are
created, nothing touches your vaults, no backups are pushed. Read what
it says it *would* do. When you're happy: the real run is
**`GitCurator-CLI.bat`**, and the fast lane for every day after is
**`GitCurator-SYNC.bat`** — it checks everything is in place, waits
for one Enter, and syncs.

## Is everything up and ready? — Test Connection

In a terminal:

`GitCurator-cli.exe --cli --test-connection`

It checks and shows, one line each: **vaults** (found + writable),
**LLM** (Ollama / llama.cpp / API — whichever is active), **GitHub**
(token + backup repos) and the **Worker queue**. Ends with
`🏁 ALL SYSTEMS READY` or a list of what needs attention.

## What each file is

| File | What it is |
|---|---|
| `1-INSTALL.bat` | one-time check (git on PATH; safe to re-run) |
| `GitCurator-CLI-Setup.bat` | first-run credential wizard |
| `GitCurator-DRY-RUN.bat` | full automatic run, writes nothing |
| `GitCurator-CLI.bat` | full automatic run (real) |
| `Start-GitCurator-CLI.bat` | same as GitCurator-CLI.bat (kept for muscle memory) |
| `GitCurator-SYNC.bat` | **the fast sync lane** — check → one Enter → sync |
| `GitCurator.bat` | the desktop app door — runs `GitCurator.exe` when the GUI milestone ships (today it prints an honest card and points you at the CLI) |
| `gitcurator-cli.sh` | the Linux/macOS twin of the CLI launcher |
| `GitCurator-cli.exe` | the command line itself |
| `config.json` | created when you save settings — your credentials, keep it private |
| `config.example.json` | the settings template (safe to look at) |
| `cache.db` | the app's memory (created on first run) |
| `reports\` | run reports, dry-run plans, vault scans land here |
| `VERSION` | the build's version stamp |

## Good to know

- **LLM:** the default is local Ollama (`http://localhost:11434`).
  Don't have it? Install from <https://ollama.com>, then run
  `ollama pull llama3.1` — or use a local **llama.cpp** server
  (llama-server): the app detects it AUTOMATICALLY at launch (whatever
  port it runs on) and switches to it when Ollama isn't running.
  Terminal twin of the detect buttons:
  `GitCurator-cli.exe --cli --detect-llm ollama` (or `llamacpp`).
- Everything the app writes stays inside this folder (`config.json`,
  `cache.db`, `reports\`) or inside the vaults you point it at.
- **Updating later:** unzip the new version and copy your old
  `config.json` (and `cache.db` if you want the app to remember
  processed links) into the new folder.

Version 0.92.0-go.1 — the Go port's CLI milestone: the full command
line (status, wizard, test connection, dry-run, automatic runs, sync,
retry, hand delivery, the LLM detect tools) with byte-parity to the
Python app's CLI, proven by the differential corpus. The desktop app
follows with the GUI milestone.
