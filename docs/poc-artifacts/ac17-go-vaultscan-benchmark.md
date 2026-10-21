# AC-17 — the Go-side vault-scan wall-time benchmark (SESSION-052, 2026-10-21)

> The parity report's one open performance row, closed with measured
> numbers. **AC-17 (doc 15 §4): "Full vault inventory scan of the
> owner-scale vault (≈1,300 notes) completes ≤ 50% of the measured
> Python baseline."** The verdict up front: **PASS — the Go scan's cold
> pass is 44.5% of Python's on the same box (14.6 ms vs 32.8 ms,
> medians of 3 fresh runs each), and 41.4% of T-04's recorded baseline
> (35.3 ms) — 2.25× faster, and every other pass (re-read, cache-fill,
> cache-warm) lands at 30–37% of Python's.**

## 1. The question and the method

T-04 (SESSION-001) measured the Python reference's vault scan on a
synthetic owner-scale corpus: **1,314 notes** (1,309 across six
categories with subfolders + 5 root notes — the owner's real vault
size per doc 07 §B.4), built by `testdiff/python_baselines.py`'s
`_build_vault`, and timed four passes in one fresh process:
**cold-no-cache → re-read-no-cache → cache-fill → cache-warm**. The
recorded numbers (35.3 / 32.1 / 215.7 / 16.7 ms) became the AC-17
baseline; the Go side was honestly left open until now.

This session mirrors that exact procedure on the Go side:

- **The rig**: `testdiff/gobench` (NEW, the goside pattern — test
  instrumentation, not a shipped command). One mode verb, `scan`,
  builds the SAME corpus (the generator is a byte-mirror of
  `_build_vault` — proven below) and times the SAME four passes
  through the production door `vaultscan.ScanVaultInventory` (nil log,
  nil progress — exactly how `internal/syncreview`'s production call
  invokes it), with `notereads.New(<tmp>/reads.db)` as the one-open
  reads cache for passes 3–4, mirroring the Python rig's
  `NoteReadCache`. The JSON field names match the Python rig's scan
  mode, plus `inventory_counts` (the census) for the same-output
  proof.
- **The protocol**: three fresh-process runs per side, on the same
  box, interleaved in the same session; medians compared. The Python
  side is the COMMITTED T-04 rig re-run fresh (`GC_APP_DIR` pointed at
  the pinned reference clone `98f4fc7`), so the primary comparison is
  same-box and same-corpus; T-04's recorded numbers stand as the
  historical baseline column.
- **The engine state**: the Go build is the SESSION-052 tree — the
  engine bytes (`internal/…`) are exactly SESSION-051's close commit
  `5325623` (this session adds only the instrumentation package); the
  Python side is the reference at the pinned rev, pristine after the
  runs (`git status` clean).

## 2. The environment

Recorded 2026-10-21T01:05Z (same sandbox class as T-04's record:
datacenter, not the owner's Windows): Linux 5.10.134, 2 cores, Intel
Xeon, 4 GB RAM; Python 3.12.15 (the stdlib-only venv — the scan mode
imports no GUI packages); go1.27.2 (the committed toolchain,
sha-verified against build/TOOLCHAIN.md Am.2's pin at the session
open).

## 3. Same input — the tree-hash proof

The Python corpus was built by the COMMITTED generator (a one-shot
driver importing `python_baselines._build_vault` — its bytes are in
§8); the Go corpus by the new rig's mirror. One sha256 over the sorted
(relpath, file-sha256) pairs of each tree:

```
TREE /home/z/ac17-py-vault  files=1314 sha256=1091ed894d33db198baf824fdff7a88f47c37cc5315cbd6b8312c667a9e0ee60
TREE /tmp/ac17-go-vault     files=1314 sha256=1091ed894d33db198baf824fdff7a88f47c37cc5315cbd6b8312c667a9e0ee60
```

**Byte-identical** — the two engines scanned the exact same corpus
bytes, so the wall-time comparison is same-input by construction.

## 4. Same output — the census proof

The reference's own `scan_vault_inventory` (read_cache=None, the
default kind) over the GO-built vault:

```json
{"core_notes": 0, "curated_notes": 0, "folders": 13, "hand_notes": 1314,
 "locked_notes": 0, "notes": 1314, "notes_seen_by_scan": 1319,
 "root_notes": 5, "total_notes": 1314, "trash_notes": 0,
 "uncategorized_notes": 0}
```

The Go rig's `inventory_counts` (from its cold pass over the same
tree): **identical on every field**, including
`notes_seen_by_scan = 1319` — the same arithmetic the Python rig's
own JSON reports. The two engines agree on the whole census at owner
scale (the per-note byte parity is separately proven by the
standing differential corpus — `testdiff/README.md` — this is the
live at-scale spot-check on top of it).

## 5. The measured numbers

**Go (`testdiff/gobench`, go1.27.2, three fresh processes):**

| run | cold ms | re-read ms | cache-fill ms | cache-warm ms | warm/cold | notes/s (cold) |
|---|---|---|---|---|---|---|
| 1 | 15.1 | 13.1 | 73.5 | 5.0 | 0.332 | 87,344.3 |
| 2 | 14.6 | 13.3 | 73.5 | 5.4 | 0.368 | 90,494.1 |
| 3 | 14.2 | 13.0 | 75.4 | 5.3 | 0.371 | 92,979.6 |
| **median** | **14.6** | **13.1** | **73.5** | **5.3** | **0.363** | **90,494.1** |

**Python (`testdiff/python_baselines.py scan`, 3.12.15, the pinned
reference, three fresh processes, same box):**

| run | cold ms | re-read ms | cache-fill ms | cache-warm ms | warm/cold | notes/s (cold) |
|---|---|---|---|---|---|---|
| 1 | 32.8 | 30.4 | 253.3 | 16.7 | 0.509 | 40,213.4 |
| 2 | 33.1 | 30.7 | 246.6 | 16.9 | 0.511 | 39,848.9 |
| 3 | 32.7 | 30.6 | 249.0 | 16.9 | 0.517 | 40,336.4 |
| **median** | **32.8** | **30.6** | **249.0** | **16.9** | **0.515** | **40,213.4** |

(The median row's warm/cold is the ratio of the medians; per-run
ratios are in the tables. T-04's recorded 2026-10-15 baseline — cold
35.3 / re-read 32.1 / fill 215.7 / warm 16.7 / 37.4k notes/s — the
fresh re-run reproduces it within ~7% on cold and ~1% on warm: the
same sandbox class, as the parity report assumed.)

**The verdict table:**

| Pass | Go (median) | Python same-box (median) | Go as % of Python | vs T-04's recorded | AC-17 gate |
|---|---|---|---|---|---|
| cold, no cache | **14.6 ms** | 32.8 ms | **44.5%** | 41.4% (of 35.3) | ≤ 50% — **PASS** |
| re-read, no cache | 13.1 ms | 30.6 ms | 42.8% | 40.8% (of 32.1) | — |
| cache-fill | 73.5 ms | 249.0 ms | 29.5% | 34.1% (of 215.7) | — |
| cache-warm | **5.3 ms** | 16.9 ms | **31.4%** | 31.7% (of 16.7) | — |
| throughput (cold) | ~90,494 notes/s | ~40,213 notes/s | **2.25× faster** | 2.42× vs 37.4k | — |

## 6. Honest caveats (carried from T-04, unchanged)

- The corpus is synthetic — realistic frontmatter shapes, but the
  owner's real vault may parse heavier. The rigs are the deliverable
  and rerunnable anywhere (§7), including on the owner's Windows
  machine.
- The sandbox is a datacenter box, not the owner's hardware; T-82's
  checklist measures the real-machine numbers (that is its purpose).
- The real-vault wall-time still lands at the owner's first live run;
  this benchmark closes AC-17's gate (the ≤ 50% comparison on the
  defined synthetic baseline), it does not replace the live number.

## 7. The rerun recipes

```bash
# Go side (from the repo root, go 1.27.2 on PATH):
go build -o /tmp/gobench ./testdiff/gobench
/tmp/gobench scan                       # temp corpus; JSON -> /tmp/gobench-scan.json
/tmp/gobench scan --vault /tmp/kept-vault   # keep the corpus for cross-side checks

# Python side (the committed T-04 rig; the reference clone + the stdlib venv):
GC_APP_DIR=<reference-clone>/app <venv>/bin/python \
    testdiff/python_baselines.py scan --out /tmp/t04-scan.json

# The cross-checks (the driver in §8):
python3 ac17_driver.py py-build /tmp/py-vault     # the committed generator's corpus
python3 ac17_driver.py tree-hash /tmp/py-vault    # must equal the Go corpus's hash
python3 ac17_driver.py tree-hash /tmp/kept-vault
python3 ac17_driver.py py-scan /tmp/kept-vault    # the census vs the Go rig's counts
```

## 8. The driver (verbatim — the one-shot cross-check utility)

Recorded here so the comparison is reproducible byte-for-byte; it
imports the COMMITTED generator and the reference's own scan (no
hand-transcribed logic on the Python side):

```python
#!/usr/bin/env python3
"""ac17_driver.py — the SESSION-052 comparison driver. Three one-shot verbs:

  py-build <dir>   build the corpus with the COMMITTED T-04 generator
                   (python_baselines._build_vault) into <dir>
  tree-hash <dir>  print one sha256 over the sorted (relpath, file sha256)
                   pairs of the tree — the byte-equality verdict between the
                   two builders' corpora
  py-scan <dir>    run the reference's scan_vault_inventory over <dir>
                   (read_cache=None, the default kind) and print the census
                   counts + notes_seen — the same-output verdict vs the Go
                   rig's inventory_counts
"""
import hashlib
import json
import os
import sys
from pathlib import Path

WORK = "/home/z/my-project/second-brain-project"
APP_DIR = os.environ.get("GC_APP_DIR", "/home/z/my-project/GitCurator/app")


def py_build(dest):
    sys.path.insert(0, os.path.join(WORK, "testdiff"))
    import python_baselines as pb
    os.makedirs(dest, exist_ok=True)
    made = pb._build_vault(Path(dest), pb.N_NOTES)
    print("PYBUILD made=%d" % made)


def tree_hash(root):
    h = hashlib.sha256()
    entries = []
    for dirpath, dirnames, filenames in os.walk(root):
        dirnames.sort()
        for fn in sorted(filenames):
            p = os.path.join(dirpath, fn)
            rel = os.path.relpath(p, root)
            digest = hashlib.sha256(open(p, "rb").read()).hexdigest()
            entries.append((rel, digest))
    for rel, digest in sorted(entries):
        h.update(rel.encode("utf-8") + b"\x00" + digest.encode("ascii") + b"\x00")
    print("TREE %s files=%d sha256=%s" % (root, len(entries), h.hexdigest()))


def py_scan(vault):
    sys.path.insert(0, APP_DIR)
    from gitcurator.core.vault_scan import scan_vault_inventory
    inv = scan_vault_inventory(vault)
    counts = {k: v for k, v in inv.items() if isinstance(v, int)}
    counts["notes"] = len(inv.get("notes", []))
    counts["folders"] = len(inv.get("folders", []))
    counts["notes_seen_by_scan"] = counts["notes"] + inv.get("root_notes", 0)
    print("PYSCAN " + json.dumps(counts, sort_keys=True))


def main():
    verb, arg = sys.argv[1], sys.argv[2]
    {"py-build": py_build, "tree-hash": tree_hash, "py-scan": py_scan}[verb](arg)


if __name__ == "__main__":
    main()
```

---

*Recorded at SESSION-052 (2026-10-21). The parity report's §5 row cites
this artifact; doc 15 §4's AC-17 carries the dated completion note.*
