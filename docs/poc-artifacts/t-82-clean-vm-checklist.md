# T-82 — the clean-VM checklist + the release report + the AV note (SESSION-050, 2026-10-20)

> Doc 10 §7's nine items as the owner's fill-in execution record + AC-23's
> release-report scaffold + doc 10 §9's AV mitigation ladder. **The live
> Windows execution is the owner's** (this sandbox has no Windows box —
> the standing constraint since SESSION-001); this document is the
> prepared surface the owner's signed results land into. T-81's dry-run
> (evidence doc `t-81-release-workflow.md` §5) already proved the
> AUTOMATION half of AC-23 (the release carries the zip + the sha256 +
> the parity table in its notes); this document carries the LIVE half.

## 1. THE CLEAN-VM CHECKLIST (doc 10 §7 — the owner's execution record)

Each item: the do, the verify, and the record slot. Fill RESULT
(PASS/FAIL/BLOCKED + one line), DATE, and NOTES (what you saw). Item 9
is already half-automated (the workflow's notes carry the sha256; the
spot-check recipe is below).

**The v0.92.0-go.2 attempt (the owner's own machine, reported in chat
2026-10-24: "I have downloaded the latest github release. none of them
works, neither cli or GUI") — recorded below verbatim-honest; no error
text was provided, so the rows carry what is known and mark what is
not. The v0.92.0-go.3 re-run slots follow the table.**

| # | The do | The verify | RESULT | DATE | NOTES |
|---|---|---|---|---|---|
| 1 | Fresh Windows 10 VM (no dev tools); unzip the release zip; run `1-INSTALL.bat`; launch `GitCurator.exe` | GUI launches; installer checks WebView2 + git and bootstraps/points as designed | **FAIL** | 2026-10-24 | The owner's own machine (not a clean VM): "none of them works, neither cli or GUI" — the GUI does not launch. No error text provided; whether 1-INSTALL.bat ran is not specified. |
| 2 | Cold start the GUI | First frame ≤ 1.5 s; window 1000×750; theme correct | **NOT REACHED** | 2026-10-24 | The GUI never launches on the owner's machine (item 1's FAIL). |
| 3 | `GitCurator-cli.exe --status` in a terminal | Completes ≤ 0.5 s; the config card prints | **FAIL** | 2026-10-24 | "neither cli or GUI" — the CLI does not work either. No error text; how it was invoked (terminal vs double-click) is not specified. |
| 4 | `GitCurator-DRY-RUN.bat` | Completes; the dry-run report writes; **zero** vault writes (snapshot diff before/after) | NOT RUN | — | Not reported. |
| 5 | The Test Connection modal | Against the owner's real services (or the VM's fakes) | NOT RUN | — | Not reported. |
| 6 | A real 5-link batch against a temp vault | Seals against a temp GitHub repo; publish against the staging Worker door (or the loopback rig) | NOT RUN | — | Not reported. |
| 7 | The GUI's live behaviors | No console window on launch; sounds play; dark/light toggle; a 10k-line log flood stays interactive | NOT RUN | — | Not reported. |
| 8 | AV check (§3 below) | The unsigned exe through a scanner service; the verdict + any mitigations recorded in §3 | **UNKNOWN** | 2026-10-24 | No scanner verdict provided. AV interference is a LIVE hypothesis for the simultaneous double failure (a dependency-free static CLI exe failing too) — unverified pending the owner's error details. |
| 9 | sha256 + reproducibility spot-check | The downloaded zip's sha256 equals the release notes' number; `./build/release.sh` from the tag reproduces it (the recipe below) | **PASS (automation half)** | 2026-10-24 | SESSION-070's legs stand (the downloaded asset's sha256 == the notes' number; the cross-machine reproduction). SESSION-074 re-verified the live download's sha256 byte-equal to the recorded number. The owner's own spot-check: not performed. |

### The v0.92.0-go.3 re-run — THE OWNER'S RESULTS (SESSION-075's intake, 2026-10-28: the prompt carried a LIVE SCREENSHOT of the app running on their machine — **THE FIRST GREEN LIVE RUN EVER**: the GUI launches, syncs, scans, the desk flows; six problems named and fixed as v0.92.0-go.4)

| # | Item | RESULT | DATE | NOTES |
|---|---|---|---|---|
| 1 | unzip (Extract All) → 0-START-HERE → GUI launch | **PASS** | 2026-10-28 | The screenshot shows the live app mid-work on the owner's machine (the GitCurator window: the sentinel labels painted — Net: online, AI: on — local (llama.cpp), Proxy: off; the SYNC batch running; the vault scans reporting 628/673 notes). |
| 2 | cold start ≤ 1.5 s / 1000×750 / theme | **PASS** | 2026-10-28 | The header's startup-frame badge reads **1.206129 s** (≤ 1.5 s, the gate); the window/theme correct. (go.4 note: the badge itself was retired from the chrome at SESSION-075 — the number stays in Stats().) |
| 3 | CLI --status ≤ 0.5 s / config card | NOT REPORTED | 2026-10-28 | Not evidenced in the prompt; the GUI's own sync flow ran (the engine works end to end on their machine). |
| 4 | DRY-RUN / zero vault writes | NOT RUN | — | Not reported. |
| 5 | Test Connection modal | NOT RUN | 2026-10-28 | The sentinel's own probes ran (Net/AI/Proxy painted); the modal itself not evidenced. |
| 6 | 5-link batch / seal / publish | PARTIAL | 2026-10-28 | A real SYNC ran (the screenshot's progress bar: "reading Github Projects Obsidian Vault — 673 / 673"; the vault scans over both vaults; hygiene clean; the scan plan proposed + answered). The full seal/publish legs not evidenced. |
| 7 | no console / sounds / theme toggle / log flood | PARTIAL | 2026-10-28 | The desk's live behaviors proven (28 rows cached, the filters/search live); sounds/toggle/flood not evidenced. |
| 8 | AV verdict (§3) | **UNKNOWN** | 2026-10-28 | The app RUNS — the go.2 double failure is resolved in practice; the specific gate (SmartScreen vs AV vs extract-first) never named. The go.2 candidates (a)-(c) are all now moot in effect; (d) pre-Win10/32-bit is disproven by the run itself. |
| 9 | sha256 + reproduce spot-check | **PASS (automation half)** | 2026-10-28 | The owner downloaded and ran the live go.3 asset (the run itself is the proof the bytes landed whole); their own hash spot-check not reported. |

**The owner's six problems (the prompt's own list, verbatim-honest) —
all six FIXED in v0.92.0-go.4 (SESSION-075):**

1. *"In logs in shows, Sentinel: net=online latency=535ms · AI=on,
   what is sentinel does to this project? I've never requested it."* →
   THE QUIET-HEARTBEAT LAW: the sentinel is the connection-health
   monitor (the reference's v0.85.0 net_guard — its job: paint the
   Net/AI/Proxy cues, enforce the offline warn-once law); its
   per-pass log line was noise — now silent on steady state
   (transitions still speak).
2. *"Got this error while syncing, though everything is set up
   correctly: ⚠️ Could not list G:\Docs\…\_review: The system
   cannot find the file specified."* → THE MISSING-FOLDER LAW: a
   vault without a `_review` folder is the healthy normal — ENOENT
   returns silently; other read errors keep their warning.
3. *"The overal UI and fonts are off. (see screenshot)"* → THE INTER
   LAW (the embedded Inter variable font) + THE WRAP LAW (log lines
   wrap — the NoWrap elide cut mid-word: "co…", "a…", "th…") + the
   debug startup-frame badge retired from the chrome.
4. *"the app container is too big it must be like 20% smaller."* →
   THE CONTAINER LAW: the window 1000×750 → **800×600** (exactly −20%).
5. *"Use real icons instead of emojies"* → THE EMOJICON LAW: every
   emoji in every surface renders as a real Lucide SVG glyph
   (emojicons.js — the official lucide-static v0.544.0 geometry, the
   golden-locked pack untouched).
6. *"the app must be able to fetch from telegram bot, regardless of
   the windows is proxy on or off, because telegram bot is connected
   to cloudflare and system can try to reach cloudflare instead."* →
   THE REACH-CLOUDFLARE LAW: the Worker channel (the bot queue's own
   path) was already proxy-independent; the DNS-poisoning class
   (proxy OFF, a lying local resolver killing the workers.dev name)
   now falls through to **Cloudflare's own DoH answer** (1.1.1.1,
   dialed by IP — the question never touches the poisoned resolver),
   the answer's address dialed directly with the original hostname's
   TLS.

### The v0.92.0-go.4 re-run slots (the owner's legs against the six-fix release)

| # | Item | RESULT | DATE | NOTES |
|---|---|---|---|---|
| 1 | Extract All → 0-START-HERE → GUI launch (800×600 now) | ____ | ____ | ____ |
| 2 | cold start ≤ 1.5 s / **800×600** / Inter / wrapped log lines | ____ | ____ | ____ |
| 3 | CLI --status ≤ 0.5 s / config card | ____ | ____ | ____ |
| 4 | DRY-RUN / zero vault writes | ____ | ____ | ____ |
| 5 | Test Connection modal (real icons everywhere now) | ____ | ____ | ____ |
| 6 | 5-link batch / seal / publish | ____ | ____ | ____ |
| 7 | no console / sounds / theme toggle / log flood | ____ | ____ | ____ |
| 8 | AV verdict (§3) | ____ | ____ | ____ |
| 9 | sha256 + reproduce spot-check | ____ | ____ | ____ |

**The SESSION-074 diagnosis record (the evidence at hand):** the live
asset re-downloaded fresh — sha256 `78990ccb…d6c3cd` byte-equal to the
recorded number; both exes extracted byte-equal to the parity table
(`9d6c2c00…`, `a6539a1d…`); PE-inspected: both PE32+ x86-64
(AMD64), `GitCurator.exe` subsystem **WINDOWS_GUI**, `GitCurator-cli.exe`
subsystem **WINDOWS_CUI**, both stamped for Windows 10+, the CLI's
buildinfo carries `go1.27.2` + `mod … v0.92.0-go.2` — **the shipped
bytes are exactly the verified, reproducible release artifacts.**
Go 1.21+ requires 64-bit Windows 10 or later; Windows 7/8/8.1 or a
32-bit system refuses both exes at load. The candidates for the
owner's failure, unranked: (a) SmartScreen blocking the unsigned
exes ("Windows protected your PC" — the default reading of "does not
work" for an unsigned app); (b) an antivirus quarantining/blocking both
exes; (c) running from inside the zip window (no extraction);
(d) pre-Windows-10 or 32-bit Windows; (e) a Windows-only runtime
bug — **not excludable from this sandbox** (no Windows box since
SESSION-001; the exes had never been executed on real Windows before
this report). Distinguishing (a)/(b)/(c)/(d) from (e) needs the
owner's exact error text — requested in the SESSION-074 report.

### The item-9 spot-check recipe (any machine with go 1.27.2 + wails v2.16.0)

```bash
# 1. verify the download (PowerShell: Get-FileHash <zip> -Algorithm SHA256;
#    or: certutil -hashfile <zip> SHA256) — must equal the release notes' sha256.
# 2. reproduce from the tag (the same ritual the release workflow runs):
git clone --branch <tag> https://github.com/assadigit/second-brain-project
cd second-brain-project && ./build/release.sh
# the ritual's table must equal the release notes' table (per-tag numbers —
# THE TAG-STATE BUILD LAW: always compare at the tagged checkout).
```

## 2. THE RELEASE-REPORT SCAFFOLD (AC-23 — fill per release)

```markdown
# GitCurator <VERSION> — the Windows release report

- Tag: v<VERSION> at <commit sha>   (the one-version-truth: tag == v$(cat VERSION))
- Zip: GitCurator-v<VERSION>-windows.zip — <N> files, <bytes-in> bytes in,
  <bytes-zipped> zipped, sha256 <sha256>  (== the GitHub release notes' table;
  == a local `./build/release.sh` at the tag — AC-22/AC-23)
- Built by: the release workflow (run <run-id>) wrapping build/release.sh
  (go1.27.2 + wails CLI v2.16.0; both exes + the zip double-built + one
  cold-GOCACHE leg per exe, every sha256 pair identical)

## The parity table (from the release notes)
| Artifact | sha256 (A = B = cold) |
|---|---|
| GitCurator-cli.exe | <cli-sha256> |
| GitCurator.exe (GUI) | <gui-sha256> |
| GitCurator-v<VERSION>-windows.zip | <zip-sha256> |

## The clean-VM checklist results (doc 10 §7 — the nine items)
| # | Item | RESULT | DATE |
|---|---|---|---|
| 1 | unzip → 1-INSTALL → GUI launch | ____ | ____ |
| 2 | cold start ≤ 1.5 s / 1000×750 / theme | ____ | ____ |
| 3 | CLI --status ≤ 0.5 s / config card | ____ | ____ |
| 4 | DRY-RUN / zero vault writes | ____ | ____ |
| 5 | Test Connection modal | ____ | ____ |
| 6 | 5-link batch / seal / publish | ____ | ____ |
| 7 | no console / sounds / theme toggle / log flood | ____ | ____ |
| 8 | AV verdict (the report's §AV) | ____ | ____ |
| 9 | sha256 + reproduce spot-check | ____ | ____ |

## AV (doc 10 §9)
- Scanner/service used: ____   Date: ____
- GitCurator.exe verdict: ____ / GitCurator-cli.exe verdict: ____
- The `-w -s` comparison (ONLY if a stripped build was tested): ____
- Mitigation applied (if any): ____

Signed: ______________  Date: ______
```

## 3. THE AV NOTE (doc 10 §9 — the mitigation ladder + the verdict slots)

Go exes occasionally trip heuristic scanners (packed strings). The
ladder, in order (each step only if the previous is insufficient):

1. **Nothing** — a clean scan needs no mitigation; record the verdict.
2. **Keep the build simple** — already the law: `-trimpath`, no
   obfuscation, **no `-s -w` strip** (the ritual's exes are unstripped;
   stripping shrinks the exe but can worsen heuristic scores — test
   BEFORE adopting: build once with `-ldflags "-s -w -X main.version=…"`
   and compare the scanner verdicts; the sha256 will differ from the
   release table — a stripped build is a DIFFERENT artifact, never the
   release one).
3. **Document the sha256 in the release notes** — already the law (the
   workflow's notes carry it): the owner verifies every download against
   it (`Get-FileHash` / `certutil`).
4. **An EV/OV code-signing cert** — the owner's call (cost/benefit);
   would also quiet SmartScreen. Noted here as the last rung, per doc 10
   §9.

The verdict slots live in §2's scaffold (the AV section) — the owner
fills them per release; a heavily-flagged verdict triggers the ladder
from rung 2.

## 4. The honest boundary

- Items 1–8 need a live Windows machine — the owner's; this sandbox
  cannot execute them (standing since SESSION-001).
- Item 9's automation half is DONE and green (T-81's dry-run: the
  release notes' table + the cross-machine byte-identity proof).
- The scaffold's `<…>` slots are per-release; the CURRENT tagged-state
  numbers for `v0.92.0-go.1` (should the owner ever re-cut it) live in
  the T-81 evidence doc §5's table.
- **Dated note (2026-10-22, SESSION-063):** item 1's installer check
  gained its second leg — `1-INSTALL.bat` now ALSO probes (read-only)
  for the WebView2 Evergreen Runtime (the Evergreen runtime's own
  registry key, per-user then per-machine; the found/absent cards; the
  one-click pointing) — so the verify column's "installer checks
  WebView2 + git and bootstraps/points as designed" is true by design
  from the next release on. The honest-zip boundary: the CURRENT live
  `v0.92.0-go.1` zip carries the OLD git-only checker (12 files); the
  new leg rides the NEXT milestone release's zip (the two-exe shape).
  Running item 1 against the current zip exercises the git leg +
  the exe's own first-launch confirm/download (wails' default
  strategy) — the desktop app's zip rides the next release either
  way, so the natural time to run item 1 is after it.
- **Dated note (2026-10-24, SESSION-074 — the owner's first live run
  + the go.2 FAIL):** the owner downloaded the live `v0.92.0-go.2`
  and reported "none of them works, neither cli or GUI" (no error
  text). The intake rides §1 above (items 1/3 FAIL, item 2 not
  reached, items 4–7 not run, item 8 unknown) + the SESSION-074
  diagnosis record (the shipped bytes verified correct; the root
  cause on the owner's machine undetermined from this sandbox —
  SmartScreen/AV/extract-first/pre-Win10 are the candidates, a
  Windows-only runtime bug not excludable). The owner's directive:
  the next version must WORK and be portable without install —
  `v0.92.0-go.3` ships the portable-first reframe (`0-START-HERE.txt`
  first in the zip + the QUICKSTART reframe), and the re-run slots
  above target it. The AV ladder's rung 4 (a code-signing cert —
  the only rung that quiets SmartScreen at the source) is surfaced
  to the owner as their call.
