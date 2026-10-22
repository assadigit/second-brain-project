# SESSION-063 — 2026-10-22 — Phase 7: the Phase-7 finish, owner-input driven — path (c) a sixth time under the standing delegation: the 1-INSTALL BOOTSTRAPPER LEG landed with proof (the last named test-adjacent debt; doc 10 §4's [verify] resolved; the t-70 plan's recorded GUID corruption corrected with teeth)

## Goal (from NEXT-SESSION.md at start)

SESSION-063 = the Phase-7 finish, OWNER-INPUT DRIVEN — the decision
tree in priority order: (a) T-82 results → the intake + the flip +
the Phase-7 exit evaluation; (b) "ship it" + the number → the release
session; (c) otherwise the standing delegation ("you choose; no haste
to start the app; only tell me if you really need me to test the
app") → ONE slice properly with proof: the bootstrapper leg (doc 10
§7/§4 — the last named test-adjacent debt, Phase-7 checklist flavor)
or the Phase-8 menu surfacing. The owner's input this session brought
no T-82 results and no "ship it" — path (c), and the chosen slice:
**the bootstrapper leg** (the meatier autonomous candidate; the Phase-
8 menu rides the report as the honest what's-next statement, unchanged
for the owner's pick).

## What was done (task IDs, files, commits — cite hashes)

- **The 58th wipe absorbed before work** (both repos, the toolchain,
  the venv, the wails CLI — the sandbox held only the Next.js
  scaffold). Rebuilt per the committed recipes: the work repo
  re-cloned (HEAD `d2fa36c` — SESSION-062's close, tree clean; the
  clone anonymous — the repo is public), the reference re-cloned via
  the **ephemeral GIT_CONFIG_COUNT PAT form** (checked out at
  `98f4fc71cfb8…`, byte-equal to testdiff/PINNED-GITCURATOR-REV; zero
  token hits in `.git/config`), go1.27.2 (70,590,635 bytes, the `-L`
  follow) sha-verified against the **committed pin** in
  build/TOOLCHAIN.md Am.2 — MATCH — plus the second verify side (the
  `?mode=json&include=all` archive listing: MATCH ×2), the venv
  (plain form; ensurepip present; Python 3.12.15), and the wails CLI
  v2.16.0 reinstalled (GOBIN=/home/z/.local/go/bin; no GUI build this
  session — no ritual change was needed, see below). **No webkit
  rebuild** (no Linux GUI build; the pending rebuild stays the
  SIXTEENTH).
- **go-ci verified GREEN at the open**: the run at `d2fa36c`
  (SESSION-062's close push — run 39833368088 — the latest) —
  success, read through the public API (nothing persisted).
- **The battery re-verified green pre-work — the exact committed
  shape**: main plain **2397 PASS / 0 FAIL / 0 SKIP over 36 ok**
  (exit 0), race **36 ok / 0 data races** (exit 0), webui **95 / 0 /
  0 over 4 ok** (exit 0), `gofmt -l` empty + vet + build clean in
  both modules. The pinned states hold: VERSION = `0.92.0-go.1`
  (12 bytes + LF), the MAIN go.mod/go.sum byte-unchanged, T-82 §1's
  nine slots still `____`.
- **THE SLICE, settled at byte level BEFORE anything was "fixed" —
  then landed test-first, then proven** (the house chain):
  - **The settle**: the debt = the 1-INSTALL equivalent's WebView2
    leg (doc 10 §4's runtime table: "the installer/`1-INSTALL`
    equivalent checks and bootstraps the evergreen installer
    **[verify]** the Wails bootstrapper path"; T-82 item 1's verify
    column: "installer checks WebView2 + git and bootstraps/points as
    designed" — UNTRUE until this session: the shipped checker tested
    only git, so on ADR-0001 Am.1's own machine (WebView2 assumed
    ABSENT — updates stopped + Edge deleted) it printed "Everything
    the app needs is in place" while the desktop app's rendering
    runtime was missing). The Evergreen GUID settled against the
    authority (learn.microsoft.com's distribution page, fetched:
    **4 hits for `{F3017226-FE2A-4295-8BDF-00C3A9A7E4C5}`, 0 for the
    corrupted form**) — the per-user key
    `HKCU\Software\Microsoft\EdgeUpdate\Clients\{GUID}`, the
    per-machine key `HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\
    Clients\{GUID}`, value `pv` (REG_SZ). **The t-70 poc plan's
    recorded GUID (`{F3017226-FE2A-4295-8BDF-00C3A3960767}`) was a
    CORRUPTION** — a scrambled last group, not even a valid GUID
    shape, standing wrong in the artifact since SESSION-033; a probe
    carrying it reads a nonexistent key and reports a FALSE absent on
    machines that have the runtime present.
  - **The runtime half settled at source level (the record; no code
    change)**: the built GUI exe carries wails v2.16.0's DEFAULT
    `download` strategy (byte-settled in the pinned CLI's source:
    `cmd/wails/flags/build.go:71` sets `WebView2: "download"`;
    `internal/wv2installer/download.go`'s build tags make it the
    compiled default): at first launch on a WebView2-less machine the
    exe asks once (a confirm MessageBox), then fetches Microsoft's
    Evergreen bootstrapper
    (`https://go.microsoft.com/fwlink/p/?LinkId=2124703`) and runs it.
    No ritual change needed — the flag equals the default under the
    pin; the proven release command stands.
  - **The landing, TEST-FIRST**: the new assertions written and run
    **RED** on their first run (the empirical proof — exactly the
    missing leg): 8 failures on the self-side (`the probe variable
    reset missing`, both probe lines, all four cards, the official
    page, the re-check line, `the Evergreen GUID … found 0`) + 8 on
    the golden differential (the same lines through the goOnly
    contract) — 16 predicted failures, zero unexpected. Then the
    .bat: `cmd/gitcurator-cli/1-INSTALL.bat` rewritten via **python
    on the real bytes** (CRLF explicit; the Write tool mangles line
    endings) — the WebView2 leg after the git leg: `set "WV2VER="`,
    the per-user probe, the per-machine fallback probe (both
    `for /f "tokens=2,*" … reg query … /v pv 2^>nul ^| findstr /i
    "REG_SZ"'), the found card (`Using the WebView2 Runtime: found
    (version %WV2VER%) - the desktop app is ready.`), the absent card
    (the honest pointing: nothing is broken — the desktop app offers
    the one-click download at its first launch; or install now from
    the official page, then re-run 1-INSTALL.bat to re-check), and
    the final card made conditional (`Everything the app needs is in
    place.` only when the runtime was found; the honest git-ready
    summary otherwise). Byte-verified: **3918 bytes, 82 CRLF, 0 bare
    LF/CR, pure ASCII, final CRLF, the GUID exactly twice, the
    corrupted shape absent**. GREEN post-fix on all surfaces.
  - **The pins, BOTH sides**: `launchers_test.go`'s
    `TestInstallBatChecksGitOnly` → **`TestInstallBatChecksGitAndWebview2`**
    (a strict SHARPENING, never a weakening: every old assertion
    stands; `reg query` moved from the over-broad read-ban to the
    required read-only verb — the one allowed reg verb; the mutating
    verbs reg add/reg delete/setx/msiexec/powershell + the PATH/venv
    verbs all stay banned; the GUID byte-count == 2 + the corrupted
    shape banned outright) and the golden differential's 1-INSTALL
    goOnly contract (`internal/cli/goldencli_test.go` — the same
    lines through the reference-anchored stone; the "Everything … in
    place" diverge pin updated to its new in-block indent).
  - **THE MUTATION TEETH proven on the corrupted-GUID shape** (the
    exact bug class the t-70 plan carried): both probes flipped to
    `{F3017226-FE2A-4295-8BDF-00C3A3960767}` → the self-side pin
    FAILS exactly (`the WebView2 Evergreen registry probe missing` ×2
    + `the Evergreen GUID … found 0` + `the corrupted t-70 poc GUID
    shape … is present`) AND the differential FAILS (both probe
    goOnly lines) → reverted byte-identical (cmp-verified) → GREEN
    again.
  - **The exe-strategy + docs surface**: the QUICKSTART §1 + the file
    table updated (the read-only probe; the one-click first-launch
    offer); doc 10 §4's dated note (the [verify] resolution, both
    halves); the t-70 poc artifact's dated correction; the t-82
    checklist's honest-zip note (the new leg rides the NEXT release's
    zip — the current live `v0.92.0-go.1` carries the old git-only
    checker).
  - **The reference's own module green in the venv** (the authority):
    `tests.test_packaging` — **14 tests, OK** (the reference's own
    launcher/zip bytes and laws hold on its untouched tree; the
    golden's authority bytes stand).
- **THE MULTIEDIT-TAB corruption struck a FIFTH time, both Go files**
  (launchers_test.go: 0 tab-lines vs 274 space-indented;
  goldencli_test.go: 0 vs 2427 — the tools re-expanded whole files to
  spaces): caught by the standing `gofmt -l` after each edit batch,
  repaired by `gofmt -w` (HEAD is gofmt-clean), the post-repair diffs
  verified to be EXACTLY the edit regions (67/8 and 13+2 lines). One
  **anchor-asserted python patch refused on a wrong indent guess**
  (5 tabs written, 4 in the file) — nothing was patched, exactly the
  belt's purpose; the corrected anchor applied cleanly.
- **The full battery at close**: main plain **2397 PASS / 0 FAIL / 0
  SKIP over 36 ok** (exit 0 — the count UNCHANGED: the slice extended
  one test + the differential contract, not the rows; the arithmetic
  closes exactly), race **36 ok, 0 data races** (exit 0), webui **95
  / 0 / 0 over 4 ok** (exit 0), `gofmt -l` empty + vet + build clean
  in both modules, the MAIN go.mod/go.sum + VERSION byte-unchanged
  (zero new dependencies), the byte-level secret scan clean (zero
  token-shaped strings), the reference tree untouched (read-only).

## Gate evaluation (pass/fail/partial + EVIDENCE)

**The chosen path's own acceptance (the standing delegation's ONE
slice properly with proof): PASS.** The last named test-adjacent debt
is closed with the full chain: the byte-level settle first (the
authority fetch — 4/0 GUID hits; the t-70 corruption identified and
corrected), the test-first RED (16 exact failures), the landing
(python on the real bytes, CRLF byte-verified), GREEN post-fix, the
mutation teeth on the corrupted-GUID shape (both pins bite, revert
verified byte-identical), the pins on BOTH sides (self-side +
differential), the exe-strategy settle at source level, and the
battery green around it (2397 + 95, race-clean, gofmt/vet/build
clean, zero new dependencies).

**The re-verification gate (the open-of-session law after a wipe):
PASS** — the battery numbers at the open (2397 + 95, race-clean,
gofmt/vet/build clean) and at the close (the same shape; the slice
added no rows — the counts close exactly).

**The parity claim's honesty: the recorded test-debt inventory is now
EMPTY of autonomous items** — §9's list ends with the bootstrapper
leg CLOSED (preparable half) and names only the LIVE halves that are
the owner's by design (T-82 item 1 for this leg; the launchers'
clean-VM smoke rides T-82 the same way). What remains in Phase 7 is
the OWNER's, unchanged: T-82's live checklist (item 1 now executable
as designed — no haste, the owner's own standing words) + the
milestone release decision. The Phase-8 menu (doc 12's register)
stands for the owner's pick whenever wanted — surfaced here as the
honest what's-next statement; no Phase-8 work was started (stay
inside the phase).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/10-build-release-packaging.md` — §4's dated
  note: the [verify] on the WebView2 row RESOLVED (both halves: the
  installer half landed; the runtime half settled at source level,
  the live leg the owner's).
- `docs/port-blueprint/module-specs/spec-cli.md` — the dated
  SESSION-063 shipped record (the launcher spec's trail).
- `docs/poc-artifacts/t-70-wails-poc.md` — the dated correction: the
  plan's registry GUID fixed to the authority's form (reality wins).
- `docs/poc-artifacts/t-82-clean-vm-checklist.md` — §4's dated note:
  the new leg's honest-zip boundary (rides the NEXT release's zip).
- `docs/PARITY-REPORT.md` — §4.2's SESSION-063 provenance (the counts
  unchanged, the arithmetic closing exactly) + §9 (the bootstrapper
  leg CLOSED; the inventory empty of autonomous items).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-063 tail
  entry (the house pattern).
- `README.md` — the front page's "What's left (honest)" (the debt
  books empty) + the SESSION-063 status entry.
- `WINDOWS-QUICKSTART.md` — §1 (the read-only probe; the one-click
  first-launch offer) + the file table's 1-INSTALL row.
- `cmd/gitcurator-cli/1-INSTALL.bat` — the WebView2 leg (the slice).
- `cmd/gitcurator-cli/launchers_test.go` — the extended + renamed
  self-side test (a strict sharpening; the GUID guard).
- `internal/cli/goldencli_test.go` — the 1-INSTALL goOnly contract
  extension + the diverge pin's in-block indent.
- `docs/sessions/SESSION-063.md` — NEW (this report).
- `docs/NEXT-SESSION.md` — REWRITTEN (the state card v64 + the
  SESSION-064 prompt; the wipe count 57 → 58; the debt books empty).

## Blockers & decisions needed

- **NONE blocking.**
- Carried, unchanged: the milestone release decision (the owner's —
  explicitly no haste); T-82's live execution (the owner's — same);
  the dissolution call (ADR-0002 Am.5); vanilla-vs-vite; the
  go-directive acknowledgement; x/net/html keep-or-swap; the
  answering-window belt + the T-71 merged-verdict question; the
  batch-scorecard modal question; the v0.91.0 zip-parity reading
  (minor); the Phase-8 menu (the owner's pick).

## Owner actions requested

- **NOTHING this session needs you** (the standing answer under your
  delegation). The only owner-testing items remain T-82's live
  Windows legs — item 1 (the installer check + the first launch on a
  WebView2-less machine) now has its full designed shape and rides
  the NEXT milestone release's zip (the current live zip carries the
  old git-only checker; running item 1 against it still exercises
  the git leg + the exe's own first-launch confirm/download). No
  hurry: the port is proven without them, and nothing blocks on them.
- The menu, unchanged, whenever you want it: (1) T-82's checklist
  (path (a)); (2) "ship it" + the number (path (b)); (3) the Phase-8
  pick (doc 12's register — path (c) now has no named slice left; a
  next path-(c) session would surface the menu for your pick and
  close cleanly at a natural boundary).

## Handoff

- **Next session: SESSION-064 = the Phase-7 finish, owner-input
  driven — the same decision tree, with the owner's standing
  delegation recorded**: (a) T-82 intake if the results arrive; (b)
  the release session if "ship it" + the number arrive; (c)
  otherwise — with the debt books now EMPTY — the Phase-8 menu
  surfacing is the honest move (doc 12's register, the owner's pick);
  close cleanly at a natural boundary when capacity runs out.
- NEXT-SESSION.md rewritten: yes (the state card v64, the
  SESSION-064 prompt inside).
- Push state: CLEAN — this session's close commit pushed (the
  ephemeral PAT form; never stored); go-ci to verify green at
  SESSION-064's open (a red is a REAL failure now).
