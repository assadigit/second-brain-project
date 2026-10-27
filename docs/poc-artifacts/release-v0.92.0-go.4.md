# Release v0.92.0-go.4 — THE SIX-FIX release (SESSION-075, 2026-10-28)

> The owner's go.3 re-run CAME GREEN on their machine (the live
> screenshot — THE FIRST GREEN LIVE RUN EVER) and named six problems;
> this release carries the six fixes. Cut under the standing fix-cycle
> directive, riding the proven ritual verbatim (T-80/T-81: the bump →
> the tag on the bump commit → the workflows green → the release
> verified THREE ways).

## The release (the verified numbers)

- **The tag**: `v0.92.0-go.4` riding the bump commit `14d715a`
  (VERSION + webui/wails.json Info.productVersion → `0.92.0-go.4` —
  THE ONE-VERSION-TRUTH: the exe's version resource, the zip's
  stamped VERSION member, and the tag all carry one number).
- **The release**: GitHub release **425875420**, LIVE, the asset
  `GitCurator-v0.92.0-go.4-windows.zip` (**27,421,748 bytes**,
  state uploaded).
- **The zip**: **14 files, 57,142,938 bytes in, 27,421,748 zipped**
  — `0-START-HERE.txt` FIRST in the folder (the portable-first law
  from go.3 holds).
- **The workflows**: go-ci run 40565884594 (the fixes commit
  `62e0997`) GREEN + go-ci run 40566051404 (the bump `14d715a`)
  GREEN + the release run 40566051363 GREEN (all watched to
  completion).
- **The batteries at the tag**: main plain **2403** top-level PASS
  over 36 ok (2398 + 5 new: 4 DoH + 1 missing-folder) / main race
  36 ok 0 races / webui **95** over 4 ok / gofmt+vet+build clean
  both modules — **2498 total**.

## The artifact table (sha256 — the notes' own numbers, verified)

| Artifact | sha256 |
|---|---|
| `GitCurator-cli.exe` | `7c66c6709d353d084be39c8637b1e2f1adf02f3e11de502189a18064c354f690` (A = B = cold) |
| `GitCurator.exe` (GUI) | `7418cc1d40fb71878a00dc0007836ef4d6f6b0e2f1f3ab328abff4a81c533bda` (A = B = cold) |
| `GitCurator-v0.92.0-go.4-windows.zip` | `11f9d85661073326a15205a56319982db849f900e513e77a57f66fe7ace9fb6a` (A = B) |

## THE THREE-WAY VERIFICATION (the release-record law, re-proven)

1. **The fresh download** — the live asset re-downloaded: sha256
   `11f9d85661073326a15205a56319982db849f900e513e77a57f66fe7ace9fb6a`
   **byte-equal to the notes' number**; the size 27,421,748 exact.
2. **The contents** — 14 members with `0-START-HERE.txt` FIRST; both
   exes extracted byte-equal to the table (`7c66c670…` CLI,
   `7418cc1d…` GUI); the `VERSION` member od-verified **12 bytes +
   LF** (`0.92.0-go.4\n`); the ASCII version stamp `0.92.0-go.4`
   present in BOTH exes (the mkzip guard's own law); the CLI's
   buildinfo carries `go1.27.2` + `mod … v0.92.0-go.4`; the GUI's
   PE version resource carries `ProductVersion 0.92.0-go.4`
   (UTF-16, from wails.json — the one-version-truth's GUI leg).
3. **The cross-machine byte-identity** — the full sandbox rebuild at
   the tagged checkout (`git clone --branch v0.92.0-go.4` →
   `./build/release.sh`, go1.27.2 + the wails CLI v2.16.0):
   CLI exe **A = B = cold = `7c66c670…`**; GUI exe **A = B = cold =
   `7418cc1d…`**; the zip **A = B = `11f9d856…`**, 14 files,
   57,142,938 in / 27,421,748 zipped — **every number reproduced,
   exit 0.**

## What the owner should feel on their machine (the go.4 re-run)

1. **The window ~20% smaller** — 1000×750 → **800×600** (the
   container law).
2. **Real icons everywhere, no emojis** — every emoji on every
   surface (the desk's rows, the buttons, the More menu, every
   dialog, the Settings pages) renders as a real Lucide line-art
   glyph (the emojicon law).
3. **The cleaner font + full lines** — the embedded Inter variable
   font; log lines WRAP instead of cutting off mid-word (the Inter
   + wrap laws); the debug startup badge gone from the header.
4. **The quiet log** — no more `Sentinel: net=online latency=535ms
   · AI=on` every minute (the quiet-heartbeat law; the header's
   Net/AI/Proxy cues still paint live — the sentinel's real job).
5. **The clean sync** — no more `⚠️ Could not list …\_review`
   warning when the folder simply doesn't exist (the missing-folder
   law).
6. **The proxy-off queue fetch** — the bot-queue poll (the Worker
   channel) survives a poisoned local resolver: on a DNS failure it
   asks Cloudflare's own DoH resolver (1.1.1.1, dialed by IP) and
   dials the answer directly (the reach-cloudflare law). With the
   proxy ON the channel was already direct (the v2rayN fix); with
   the proxy OFF the DNS class is now covered too.

The go.4 re-run slots: `t-82-clean-vm-checklist.md` §1.

## The first-run card (unchanged from go.3 — the portable-first law)

Extract All → `0-START-HERE.txt` (the 30-second card: the
unsigned-app SmartScreen walk-through, the AV guidance, 64-bit
Win10/11) → `GitCurator.bat`. Still failing? Screenshot the exact
error text — one picture pinpoints the cause.

## The AV note (unchanged, the standing call)

The exes are unsigned (the SmartScreen prompts are the expected
first-run shape — the START-HERE card walks them). The permanent
cure is doc 10 §9 rung 4 — the code-signing cert (annual fee +
validation); the owner's call.
