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

| # | The do | The verify | RESULT | DATE | NOTES |
|---|---|---|---|---|---|
| 1 | Fresh Windows 10 VM (no dev tools); unzip the release zip; run `1-INSTALL.bat`; launch `GitCurator.exe` | GUI launches; installer checks WebView2 + git and bootstraps/points as designed | ____ | ____ | ____ |
| 2 | Cold start the GUI | First frame ≤ 1.5 s; window 1000×750; theme correct | ____ | ____ | ____ |
| 3 | `GitCurator-cli.exe --status` in a terminal | Completes ≤ 0.5 s; the config card prints | ____ | ____ | ____ |
| 4 | `GitCurator-DRY-RUN.bat` | Completes; the dry-run report writes; **zero** vault writes (snapshot diff before/after) | ____ | ____ | ____ |
| 5 | The Test Connection modal | Against the owner's real services (or the VM's fakes) | ____ | ____ | ____ |
| 6 | A real 5-link batch against a temp vault | Seals against a temp GitHub repo; publish against the staging Worker door (or the loopback rig) | ____ | ____ | ____ |
| 7 | The GUI's live behaviors | No console window on launch; sounds play; dark/light toggle; a 10k-line log flood stays interactive | ____ | ____ | ____ |
| 8 | AV check (§3 below) | The unsigned exe through a scanner service; the verdict + any mitigations recorded in §3 | ____ | ____ | ____ |
| 9 | sha256 + reproducibility spot-check | The downloaded zip's sha256 equals the release notes' number; `./build/release.sh` from the tag reproduces it (the recipe below) | ____ | ____ | ____ |

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
