# GitCurator 0.92.0-go.2 — the Windows release report (THE DESKTOP MILESTONE)

> SESSION-070 (2026-10-23) — the milestone release cut per the owner's
> post-close directive ("The work is continued until GUI is ready"),
> through the proven release workflow. The two-exe shape is LIVE:
> **GitCurator.exe** (the desktop GUI) + **GitCurator-cli.exe** (the full
> command line) in one portable zip. This report is the AC-23 scaffold
> (t-82-clean-vm-checklist.md §2) filled with the release's own verified
> numbers; the owner's live Windows legs (T-82 §1 items 1–8 + the AV
> scan) run against THIS release and land in the table below when done.

- Tag: v0.92.0-go.2 at `a74b1126ecfd06b7379809fce177823a9b02a95a`
  (the one-version-truth: tag == v$(cat VERSION); the tag rides the
  bumped-VERSION commit — THE TAG-STATE BUILD LAW)
- Zip: GitCurator-v0.92.0-go.2-windows.zip — **13 files, 56,931,140
  bytes in, 27,235,329 zipped**, sha256
  `78990ccbfed5f0a0b106b62bc6230f8aa1200a5a2682a34b2a315c2a43d6c3cd`
  (== the GitHub release notes' table; == a local `./build/release.sh`
  at the tag — AC-22/AC-23; verified three ways this session: the
  runner's own ritual, the downloaded asset, and the sandbox rebuild)
- Built by: the release workflow (run **39951431147**, all 8 steps
  green) wrapping build/release.sh (go1.27.2 + wails CLI v2.16.0; both
  exes + the zip double-built + one cold-GOCACHE leg per exe, every
  sha256 pair identical)
- Published: 2026-10-23T12:15:21Z — GitHub release id 421804097,
  "GitCurator v0.92.0-go.2 — the Windows release" (not a draft, not a
  prerelease); the asset carries the owner's law: a release without its
  zip is not done.

## The parity table (from the release notes — verified against the downloaded asset + the sandbox rebuild at the tag)

| Artifact | sha256 (A = B = cold) |
|---|---|
| GitCurator-cli.exe | `a6539a1dc52ef831414c5202a8c0c66e9299e423a9325c43ad4899a0a216355d` |
| GitCurator.exe (GUI) | `9d6c2c00d9d33b1890a005eb6afe3c65c69ea1aa287c398de8789f5a7322c400` |
| GitCurator-v0.92.0-go.2-windows.zip | `78990ccbfed5f0a0b106b62bc6230f8aa1200a5a2682a34b2a315c2a43d6c3cd` |

**Cross-machine byte-identity (the item-9 reproduce leg, run by the
release session):** a full sandbox rebuild at the tagged checkout
(`v0.92.0-go.2`, clean worktree, the same 8 ritual legs, exit 0 —
"PASS — byte-identical rebuilds: CLI exe + GUI exe + zip (warm x2 +
cold x1 per exe)") reproduces **every number in the table
byte-identically** — the GitHub runner and this sandbox agree on all
three artifacts. The CLI exe's buildinfo on both machines carries
`mod github.com/assadigit/second-brain-project v0.92.0-go.2` (the tag
IS the module version — one version, every home: the `--version`
stamp, the buildinfo, the exe's version resource, the zip's VERSION
member `0.92.0-go.2\n`, and the tag).

The zip's 13 files (the two-exe shape + the launcher set + the
git+WebView2 bootstrapper):

```
GitCurator/1-INSTALL.bat              3,918 bytes  (git + WebView2 Evergreen probe)
GitCurator/GitCurator-CLI-Setup.bat   2,334 bytes
GitCurator/GitCurator-CLI.bat         2,280 bytes
GitCurator/GitCurator-DRY-RUN.bat     1,197 bytes
GitCurator/GitCurator-SYNC.bat        2,210 bytes
GitCurator/GitCurator-cli.exe    29,209,088 bytes
GitCurator/GitCurator.bat             1,373 bytes
GitCurator/GitCurator.exe         27,696,128 bytes
GitCurator/Start-GitCurator-CLI.bat   2,373 bytes
GitCurator/VERSION                       12 bytes
GitCurator/WINDOWS-QUICKSTART.md      6,055 bytes
GitCurator/config.example.json        2,873 bytes
GitCurator/gitcurator-cli.sh          1,299 bytes
```

The honest-zip law is satisfied: the LIVE release list now carries the
desktop two-exe zip with the new git+WebView2 bootstrapper leg in
`1-INSTALL.bat` (the old live `v0.92.0-go.1` zip carried the CLI
milestone, 12 files, the git-only checker — both releases stay listed).

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
| 9 | sha256 + reproduce spot-check | **PASS (automation half — this session: the downloaded asset's sha256 verified == the notes' number; the reproduce-from-tag ritual run at the tagged checkout, every number byte-identical)** | 2026-10-23 |

Items 1–8 are the owner's live Windows legs (no haste — their standing
words); they run against THIS release (item 1 now carries the full
installer check: git + the WebView2 probe + the exe's first-launch
one-click offer). The fill-in surface: `t-82-clean-vm-checklist.md` §1.

## AV (doc 10 §9)

- Scanner/service used: ____   Date: ____
- GitCurator.exe verdict: ____ / GitCurator-cli.exe verdict: ____
- The `-w -s` comparison (ONLY if a stripped build was tested): ____
- Mitigation applied (if any): ____

## The reproduce recipe (any machine with go 1.27.2 + wails v2.16.0)

```bash
git clone --branch v0.92.0-go.2 https://github.com/assadigit/second-brain-project
cd second-brain-project && ./build/release.sh
# the ritual's table must equal this report's table (per-tag numbers —
# THE TAG-STATE BUILD LAW: always compare at the tagged checkout).
```

Signed (the automation half): SESSION-070, the release session —
2026-10-23. The live half (items 1–8 + AV): the owner's, whenever
they run the checklist.
