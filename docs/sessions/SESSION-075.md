# SESSION-075 — 2026-10-28 — Phase 7: THE OWNER'S GO.3 RESULTS CAME GREEN (path (a) of card v75's tree — the prompt carried a LIVE SCREENSHOT of the app running on their Windows machine: THE FIRST GREEN LIVE RUN EVER — the GUI launches, syncs, scans, the desk flows) + THE SIX PROBLEMS NAMED IN THE PROMPT → THE SIX FIXES LANDED + THE SIX-FIX RELEASE v0.92.0-go.4 CUT + LIVE; the battery 2403 + 95 = 2498 green at the close; go-ci + the release run GREEN

## Goal (from NEXT-SESSION.md at start)

SESSION-075 = THE POST-FIX SESSION — the decision tree in priority
order: (a) if the prompt carried the owner's go.3 re-run results or
their exact error text: the intake; if green → the Phase-7 exit
evaluation re-run; if an error names a machine gate or a Windows-only
bug → that's the next fix cycle's target; (b) otherwise the standing
delegation. **The prompt carried the owner's live results — path (a),
and they are GREEN: a live screenshot of the GUI mid-work on their
machine (the sentinel labels painted — Net: online, AI: on — local
(llama.cpp), Proxy: off; a real SYNC batch running; the vault scans
over both vaults; the desk flowing with 28 rows). The go.2 "none of
them works" era is OVER — the app runs on the owner's machine.**
The prompt then named SIX problems (the owner's own list,
verbatim-honest, recorded in the T-82 checklist §1's go.3 re-run
table + below) — the session's work was the six fixes, the battery,
and the six-fix release `v0.92.0-go.4`.

## The owner's six problems and the six fixes (the session's whole work)

1. **THE SENTINEL QUESTION** — *"In logs in shows, Sentinel:
   net=online latency=535ms · AI=on, what is sentinel does to this
   project? I've never requested it."* **The answer (recorded for
   the owner):** the sentinel is the connection-health monitor
   (the reference's v0.85.0 `net_guard` — the app's own, inherited
   behavior, not a new addition): one background pass every 60 s
   that (i) paints the header's **Net · AI · Proxy** cues (the
   labels the owner sees turning green), (ii) enforces the offline
   law (the one-time "No internet connection" advisory at launch,
   the fell-offline / back-online lines), (iii) probes the
   configured proxy port. Its per-pass log line — the noise the
   owner flagged — is GONE. **THE QUIET-HEARTBEAT LAW
   (internal/app/app.go):** a steady-state pass speaks NOTHING; the
   transition lines + the snapshot crossing (the cues' paint) stay
   exactly as they were; the dead `netWord` retired;
   `TestSentinelPassEmitsSnapshotAndOfflineLaw` pins the silence
   (a steady-state pass that speaks FAILS the test).
2. **THE \_review WARNING** — *"Got this error while syncing, though
   everything is set up correctly: ⚠️ Could not list
   G:\Docs\…\_review: The system cannot find the file specified."*
   **THE MISSING-FOLDER LAW (internal/websitepipeline/review.go):**
   a vault WITHOUT a `_review` folder is the HEALTHY normal (nothing
   waits for human eyes) — both review scans (`ScanReviewBacklog` +
   `ScanReviewNotes`) return nil SILENTLY on `fs.ErrNotExist`; every
   OTHER read error (permissions, a file squatting on the folder's
   path) keeps its warning line;
   `TestScanReviewMissingFolderIsSilent` pins both sides.
3. **THE UI + FONTS** — *"The overal UI and fonts are off. (see
   screenshot)"* (the screenshot showed: the mid-word truncation
   "co…", "a…", "th…"; the system font stack; the debug
   startup-frame badge in the header). **THE INTER LAW** (the UI
   face is the embedded **Inter variable font** — latin +
   latin-ext + cyrillic subsets, weights 100–900 from the one
   woff2; the system stacks stay the fallbacks) + **THE WRAP LAW**
   (the desk's rows WRAP — the reference's Qt NoWrap-elide was a
   widget constraint, not a design law; the HTML desk wraps, long
   paths and URLs included, full text always visible + on the
   tooltip) + the debug badge retired from the chrome (the number
   stays in `Stats()`).
4. **THE CONTAINER** — *"the app container is too big it must be
   like 20% smaller."* **THE CONTAINER LAW (webui/main.go):** the
   window 1000×750 → **800×600** (exactly −20%); the resize floor
   stays 760×380.
5. **THE ICONS** — *"Use real icons instead of emojies"*
   **THE EMOJICON LAW (webui/frontend/emojicons.js — NEW):** the
   EXTENDED lucide pack (the OFFICIAL lucide-static v0.544.0
   geometry, fetched verbatim from the package — ISC, the same
   license as the locked pack; **the golden-locked icons.js pack is
   untouched — the uitheme lock holds, the vocabulary grows**) +
   the emoji→icon map (**121 entries — the full production
   inventory**, the engines' Go strings audited emoji-by-emoji) +
   `iconify(root)` (walks a subtree's text nodes; every MAPPED emoji
   renders as a real currentColor-tinted SVG glyph, 1em square,
   baseline-aligned; FE0F-tolerant — both `⚠` and `⚠️` match;
   unknown emojis degrade to text — the graceful law; the no-match
   fast path is ONE regex test per text node, so the 10k flood
   pays nothing). Wired everywhere: the desk's rows (the LEADING
   emoji rides the level-glyph seat, LEVEL-COLORED — the v0.32
   one-muted-ink law survives), the chrome's buttons (Fast Sync /
   Scan / Test Connectivity / the hero machine's states / the
   terminal ✅/❌ marks), the More menu + its submenu, themedBox +
   every dialog builder (the battery modal, the scan picker / plan
   / scorecard, hand-deliver, chrome-retry, decommission,
   recategorize, bot-resolve, the settings mini-dialogs), the
   settings pages (at build + on the sync re-render).
6. **THE TELEGRAM FETCH / THE PROXY** — *"the app must be able to
   fetch from telegram bot, regardless of the windows is proxy on
   or off, because telegram bot is connected to cloudflare and
   system can try to reach cloudflare instead."* **THE
   REACH-CLOUDFLARE LAW (internal/workerclient/dohdial.go — NEW):**
   the Worker channel (the bot queue's own path — GET /api/pending
   and every other leg) was ALREADY proxy-independent
   (`directClient`'s `Proxy == nil`, the v2rayN-interception fix,
   byte-matching the Python reference) — BOTH proxy states ride the
   same direct sockets. The gap closed is the **DNS class**: with
   the system proxy OFF, a censored local resolver can poison the
   workers.dev name (NXDOMAIN or a black-hole address) and the
   queue poll dies before a byte crosses. The transport's dialer
   now carries the fallback: leg 1 — the plain direct dial (exactly
   as before); on failure — **Cloudflare's own DoH answer** (the
   JSON API on 1.1.1.1, dialed BY IP with the cloudflare-dns.com
   SNI — the question itself never touches the poisoned resolver),
   then the answer's address dialed directly (the ORIGINAL port;
   the TLS layer handshakes with the ORIGINAL hostname — SNI +
   certificate verification ride the transport's addr); the honest
   leg-1 error surfaces when both legs fail. Four hermetic tests
   (all loopback, zero network): the JSON grammar (A records;
   NXDOMAIN + junk answers yield nil), the selection law (the
   alive leg asks no question; the original error law), the port
   law, the endpoint-IP independence.

## What was done (task IDs, files, commits)

- **The open's bootstrap** (the 69th wipe — the FULL class this
  time: the toolchain, the venv, the repos all gone): absorbed by
  the standard recipe — go1.27.2 re-extracted to `/home/z/.local/go`
  (the tarball re-downloaded from go.dev, sha256 `ecbadb99091a…
  45fe5` verified against the TOOLCHAIN.md Am.2 pin), the venv
  rebuilt (Python 3.12.16), the work repo re-cloned to the recipe
  path @ `3895cad` (anonymously — the public repo; the remote
  scrubbed), the reference re-cloned @ `98f4fc7` (verified against
  `testdiff/PINNED-GITCURATOR-REV`), both modules build-verified
  before any work started.
- **The intake** (T-82 checklist §1's go.3 re-run table filled +
  the six problems recorded verbatim + the go.4 re-run slots
  prepared; this file).
- **THE SIX FIXES** — commit `62e0997` (17 files, +1009/−55):
  `internal/app/app.go` + `herostate_test.go` (the quiet
  heartbeat), `internal/websitepipeline/review.go` +
  `pipeline_test.go` (the missing-folder law),
  `internal/workerclient/dohdial.go` + `dohdial_test.go` +
  `sign.go` (the reach-cloudflare law),
  `webui/frontend/{emojicons.js, app.js, heroflow.js, scanflow.js,
  settings.js, style.css}` + `fonts/` (the emojicon + inter + wrap
  laws), `webui/main.go` (the container law). GO FILES edited
  python-on-bytes only (the sixth-strike law held); `icons.js` +
  `theme.css` untouched (the golden locks hold — the uitheme
  battery green proves it).
- **THE VISUAL VERIFICATION** (the owner's screenshot answered with
  the sandbox's own): the real frontend served standalone (a
  TEST-ONLY wails shim — never in the repo) at the 800×600
  viewport — Inter loads (`document.fonts.check` true, regular +
  800 weights), 13 mock log rows pushed through the REAL
  `log:batch` pipeline (the leading emojis ride the glyph seat,
  the long lines wrap — the Windows-path warning line and the
  300-char scan lines render whole), the More menu's 30 rows
  VLM-verified **zero emojis anywhere** (twice — the second pass
  after the 🪝/🌙/✕ audit closed the last three), the settings
  window (11 nav rows, all SVG) + the battery modal (4 rows, 8
  glyphs) render live.
- **THE RELEASE** — the bump `14d715a` (VERSION + wails.json →
  0.92.0-go.4 + the QUICKSTART's live-usage refs) carrying the tag
  `v0.92.0-go.4` → go-ci run 40565884594 (the fixes commit) GREEN +
  go-ci run 40566051404 (the bump) GREEN + the release run
  40566051363 → the release LIVE (the numbers in the go.4 release
  report: `docs/poc-artifacts/release-v0.92.0-go.4.md`).

## The battery at the close (the honest numbers)

- **Main plain: 2403 top-level PASS over 36 ok packages, 0 FAILs**
  (2398 + 5 new: 4 DoH + 1 missing-folder).
- **Main race: 36 ok, 0 data races** (exit 0).
- **Webui: 95 top-level PASS over 4 ok packages** (the exact
  committed shape — the golden locks held through the frontend
  work).
- **gofmt -l empty + go vet + go build clean in BOTH modules; the
  tree clean** — **2498 total.**

## THE PHASE-7 EXIT EVALUATION (honestly, against doc 12's gate)

The gate: the clean-VM demonstration + the attached release + the
owner sign-off. **Closer than ever — but still NOT GREEN, honestly:**
the owner's go.3 legs prove the GUI RUNS on their machine (items
1/2/7 partial/6 partial GREEN — the first live legs ever), and the
six problems they named are fixed + released as go.4; **but the
sign-off is not yet given** — the honest gate needs their go.4
re-run (the slots are prepared) confirming the six fixes land right
on their machine (the window size, the icons, the fonts, the quiet
logs, the clean sync, the proxy-off queue fetch). The fix cycle
narrows to exactly that.

## Handoff (the next prompt rides NEXT-SESSION.md §5)

SESSION-076 = the post-fix session: (a) the owner's go.4 results →
the intake (the checklist's go.4 slots) + if green the Phase-7 exit
evaluation re-run (the sign-off question — the owner's call); (b)
the standing delegation otherwise.
