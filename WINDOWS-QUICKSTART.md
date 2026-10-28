# GitCurator — first run on Windows (v0.92.0-go.11)

**Nothing is installed on your computer** — this is a portable
app; everything it writes stays inside its folder. Three steps and
you're in. This build carries BOTH faces of the app: the desktop app
(**GitCurator.exe**, double-click **`GitCurator.bat`**) and the full
command line (**GitCurator-cli.exe**) — same engine, same config,
same vaults.

**First: open `0-START-HERE.txt`** (it sorts first in the folder) —
the 30-second version of this guide, including what to do when
Windows shows its "protected your PC" box.

**Needs 64-bit Windows 10 or 11** (it will not start on Windows 7/8
or a 32-bit system).

## 0. Get the app (once)

Download the zip from
**[github.com/assadigit/second-brain-project/releases](https://github.com/assadigit/second-brain-project/releases)**.

- **Verify the download before unzipping** (the release notes list the
  sha256 for every zip): in a terminal,
  `certutil -hashfile GitCurator-v0.92.0-go.11-windows.zip SHA256`
  — the number must match the release notes exactly. If it doesn't,
  don't use the file; re-download.
- **Which zip you get depends on the release**: the current live
  release (`v0.92.0-go.11`, the desktop milestone line) is the
  **two-exe shape** this guide describes — GitCurator.exe +
  GitCurator-cli.exe + the launchers + the optional git+WebView2
  checker. (The older releases stay listed.)
- **Extract ALL files first**: right-click the zip → "Extract All...",
  then work from the extracted folder. **Do not run anything from
  inside the zip window itself** — Windows runs it from a temporary
  copy and things break. Extract first, always.
- Unzip anywhere you like (Desktop is fine). Nothing is installed
  outside the folder you unzipped into (and the vaults you point it
  at).

## 1. If Windows says "protected your PC" (read once, 20 seconds)

GitCurator is a personal project and is **not code-signed**, so the
first time you run it Windows may show the blue SmartScreen box:
"Windows protected your PC". That is normal for every unsigned app.

- Click **"More info"** → **"Run anyway"** — once per exe, then
  Windows remembers.
- If your **antivirus** removes or blocks `GitCurator.exe` /
  `GitCurator-cli.exe`, restore and allow them — they are plain Go
  binaries, and the sha256 numbers in the release notes prove exactly
  which bytes you downloaded.

## 2. Start the desktop app

Double-click **`GitCurator.bat`** (or `GitCurator.exe` directly).

The themed window opens: the SYNC desk (the log + the batch machine),
the scan views, and Settings (every page the app has). A console window
opens BEHIND the app if you launched it from the .bat — that is normal;
if something breaks, screenshot that window: it has the exact error.

**The desktop app needs the WebView2 Runtime** (preinstalled on
updated Windows 10/11). If your Windows doesn't have it, nothing is
broken — **at first launch the app offers to download and install it
automatically** (one internet connection, about a minute, then never
again). You can also install it yourself from
<https://developer.microsoft.com/en-us/microsoft-edge/webview2/>.

**git** is needed only for sealing/publishing into vaults (the app
tells you if it's missing — get it from <https://git-scm.com/download/win>,
default options are fine). **`1-INSTALL.bat` is an optional checker**
that verifies both at once: it only checks and points, never installs
anything itself, and is safe to re-run any time. You do not need it
to run the app.

## 3. Configure (the wizard — or the Settings pages)

Double-click **`GitCurator-CLI-Setup.bat`**, or just fill the Settings
pages in the desktop app.

It asks for the same things the app's Settings pages do: your GitHub
token, the LLM (local Ollama / llama.cpp or any OpenAI-compatible
endpoint), and your vault paths — and saves them to `config.json`
(kept private, never shared).

## 4. Safe first run (writes NOTHING)

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
| `0-START-HERE.txt` | the 30-second first-run card (extract first, the "protected your PC" box, how to start) |
| `1-INSTALL.bat` | OPTIONAL checker (git on PATH + a read-only WebView2 probe; not an installer — you never need it to run the app) |
| `GitCurator-CLI-Setup.bat` | first-run credential wizard |
| `GitCurator-DRY-RUN.bat` | full automatic run, writes nothing |
| `GitCurator-CLI.bat` | full automatic run (real) |
| `Start-GitCurator-CLI.bat` | same as GitCurator-CLI.bat (kept for muscle memory) |
| `GitCurator-SYNC.bat` | **the fast sync lane** — check → one Enter → sync |
| `GitCurator.bat` | the desktop app door — runs `GitCurator.exe` |
| `GitCurator.exe` | the desktop app itself (the themed window) |
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
- **What is the sentinel?** The **Net · AI · Proxy** cues at the top
  of the window are painted by a small background monitor (the
  Python original had it too — it is not a new addition): once a
  minute it checks your connection and the configured AI endpoint,
  and the offline warning speaks once at launch. It is silent by
  design — no heartbeat lines in the log (the old
  "Sentinel: net=online latency=… · AI=on" noise was removed in
  v0.92.0-go.10); a line appears only when something CHANGES
  (fell offline / back online). There is nothing to configure —
  it just keeps the little lights honest.
- **Updating later:** unzip the new version and copy your old
  `config.json` (and `cache.db` if you want the app to remember
  processed links) into the new folder.

Version 0.92.0-go.11 — the Go port's desktop milestone: the Wails GUI
(the SYNC desk, the scan views, every Settings page, the More tools)
plus the full command line (status, wizard, test connection, dry-run,
automatic runs, sync, retry, hand delivery, the LLM detect tools) with
byte-parity to the Python app, proven by the differential corpus.
