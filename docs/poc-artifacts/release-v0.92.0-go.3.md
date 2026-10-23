# GitCurator 0.92.0-go.3 — the Windows release report (THE FIX RELEASE under the owner's directive)

> SESSION-074 (2026-10-24) — cut under the owner's explicit directive
> ("Please make sure the next version both CLI and GUI WORKS!" +
> "portable, without install"), after their live report that
> `v0.92.0-go.2` fails on their machine ("none of them works, neither
> cli or GUI" — the intake + the diagnosis record live in
> `t-82-clean-vm-checklist.md` §1). What this release changes: the
> **portable-first reframe** — `0-START-HERE.txt` ships FIRST in the
> zip (the 30-second first-run card: extract-ALL-first, the
> SmartScreen "More info → Run anyway" walk-through, the AV
> restore/allow guidance, the 64-bit-Windows-10/11 requirement, the
> CLI invocation card, the ask-the-exact-error ask), the QUICKSTART no
> longer frames an "install" step (`1-INSTALL.bat` repositioned as the
> OPTIONAL checker it always was), and the release notes themselves
> carry the First-run section. The binaries are UNCHANGED logic
> (byte-proven lineage from the verified go.2 artifacts — the
> version stamp + the new zip composition are the only byte
> differences); this release is verified three ways below, exactly as
> go.2 was.

- Tag: v0.92.0-go.3 at `5e0c09d6a69d53a79eb12b90e022755063d13c58`
  (the one-version-truth: tag == v$(cat VERSION); the tag rides the
  bumped-VERSION commit — THE TAG-STATE BUILD LAW)
- Zip: GitCurator-v0.92.0-go.3-windows.zip — **14 files, 56,935,578
  bytes in, 27,237,300 zipped**, sha256
  `65eb64b4f666459adefbdc7edd6938d04e5a9742fcb8b960e352defb29326e69`
  (== the GitHub release notes' table; == a local `./build/release.sh`
  at the tag — AC-22/AC-23; verified three ways this session: the
  runner's own ritual, the downloaded asset, and the sandbox rebuild)
- Built by: the release workflow (run **40022070760**, completed /
  success) wrapping build/release.sh (go1.27.2 + wails CLI v2.16.0;
  both exes + the zip double-built + one cold-GOCACHE leg per exe,
  every sha256 pair identical)
- Published: 2026-10-23T20:51:23Z — GitHub release id 422298475,
  "GitCurator v0.92.0-go.3 — the Windows release" (not a draft, not a
  prerelease); the asset carries the owner's law: a release without
  its zip is not done.

## The parity table (from the release notes — verified against the downloaded asset + the sandbox rebuild at the tag)

| Artifact | sha256 (A = B = cold) |
|---|---|
| GitCurator-cli.exe | `b9d18270a3aef3864980cb4c5681ab98654e9d3a794bd9c927a0e344465d6965` |
| GitCurator.exe (GUI) | `2e88e08a77abd42e29afb1f9cc51ea55e0a8240b365234dac0ba05416a967b4f` |
| GitCurator-v0.92.0-go.3-windows.zip | `65eb64b4f666459adefbdc7edd6938d04e5a9742fcb8b960e352defb29326e69` |

**Cross-machine byte-identity (the item-9 reproduce leg, run by the
release session):** a full sandbox rebuild at the tagged checkout
(`v0.92.0-go.3`, clean worktree, the same 8 ritual legs, exit 0 —
"PASS — byte-identical rebuilds: CLI exe + GUI exe + zip (warm x2 +
cold x1 per exe)") reproduces **every number in the table
byte-identically** — the GitHub runner and this sandbox agree on all
three artifacts. The CLI exe's buildinfo on both machines carries
`mod github.com/assadigit/second-brain-project v0.92.0-go.3` (the tag
IS the module version), and the zip's VERSION member is
`0.92.0-go.3\n` (od-verified 12 bytes + LF).

The zip's 14 files (the two-exe shape + the START-HERE card FIRST +
the launcher set + the optional git+WebView2 checker):

```
GitCurator/0-START-HERE.txt                3,338 bytes  (FIRST — the 30-second card)
GitCurator/1-INSTALL.bat                    3,918 bytes  (OPTIONAL checker: git + WebView2 Evergreen probe)
GitCurator/GitCurator-CLI-Setup.bat         2,334 bytes
GitCurator/GitCurator-CLI.bat               2,280 bytes
GitCurator/GitCurator-DRY-RUN.bat           1,197 bytes
GitCurator/GitCurator-SYNC.bat              2,210 bytes
GitCurator/GitCurator-cli.exe          29,209,088 bytes
GitCurator/GitCurator.bat                  1,373 bytes
GitCurator/GitCurator.exe               27,696,128 bytes
GitCurator/Start-GitCurator-CLI.bat        2,373 bytes
GitCurator/VERSION                            12 bytes
GitCurator/WINDOWS-QUICKSTART.md            7,155 bytes  (the reframed guide)
GitCurator/config.example.json              2,873 bytes
GitCurator/gitcurator-cli.sh                1,299 bytes
```

## The clean-VM checklist results (doc 10 §7 — the nine items)

The go.2 attempt's results (the owner's first live run — FAIL on
their machine, recorded in `t-82-clean-vm-checklist.md` §1) stand as
the go.2 report's table records them. **This release's re-run slots**
are prepared in the checklist §1 (the owner's legs against go.3;
item 1 now reads: unzip (Extract All) → 0-START-HERE → GUI launch).

| # | Item | RESULT | DATE |
|---|---|---|---|
| 1 | unzip → 0-START-HERE → GUI launch | ____ (the owner's re-run against go.3) | ____ |
| 2 | cold start ≤ 1.5 s / 1000×750 / theme | ____ | ____ |
| 3 | CLI --status ≤ 0.5 s / config card | ____ | ____ |
| 4 | DRY-RUN / zero vault writes | ____ | ____ |
| 5 | Test Connection modal | ____ | ____ |
| 6 | 5-link batch / seal / publish | ____ | ____ |
| 7 | no console / sounds / theme toggle / log flood | ____ | ____ |
| 8 | AV verdict (the report's §AV) | ____ | ____ |
| 9 | sha256 + reproduce spot-check | **PASS (automation half — this session: the downloaded asset's sha256 verified == the notes' number; the reproduce-from-tag ritual run at the tagged checkout, every number byte-identical)** | 2026-10-24 |

## AV (doc 10 §9)

- Scanner/service used: not yet provided   Date: —
- GitCurator.exe verdict: not yet provided / GitCurator-cli.exe verdict: not yet provided
- The `-w -s` comparison (ONLY if a stripped build was tested): not tested
- Mitigation applied (if any): the ladder stands as recorded — rung 3
  (the sha256 in the release notes) is the law; **rung 4 (an EV/OV
  code-signing cert — the only rung that quiets SmartScreen at the
  source) is the owner's call**: the go.2 failure's leading candidates
  are SmartScreen/AV on the unsigned exes, and a cert would remove
  that entire class (cost: an annual fee + a validation process;
  benefit: no blue box for every new user — the owner decides).

## The reproduce recipe (any machine with go 1.27.2 + wails v2.16.0)

```bash
git clone --branch v0.92.0-go.3 https://github.com/assadigit/second-brain-project
cd second-brain-project && ./build/release.sh
# the ritual's table must equal this report's table (per-tag numbers —
# THE TAG-STATE BUILD LAW: always compare at the tagged checkout).
```

Signed (the automation half): SESSION-074, the fix-release session —
2026-10-24. The live half (items 1–8 + AV): the owner's, against THIS
release (the re-run slots above); their exact error text from any
remaining failure is the single most valuable next input — the
SESSION-074 report asks for it.
