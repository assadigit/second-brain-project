#!/usr/bin/env python3
"""T-64 — the differential corpus FULL-RUN driver (doc 09 §3 authority).

Runs EVERY golden generator under testdiff/goldens/ against the UNMODIFIED
pinned reference, in TWO full passes, and proves the three corpus laws:

  1. PER-GENERATOR HEALTH — every generator exits 0 and reports the
     pinned rev (the [goldens] reference rev line).
  2. THE TWO-RUN BYTE-IDENTICAL PROOF — pass 1 and pass 2 produce
     byte-identical bytes for every file any generator writes (the
     full working-tree snapshot after pass 1 == after pass 2).
  3. COMMITTED == REGENERATED — after both passes the git tree is
     CLEAN: every committed golden equals the reference's current
     bytes at the pinned rev (git status --porcelain empty).

Usage (resumable; each pass is a separate foreground invocation):

    python3 testdiff/run_corpus.py --pass 1     # all generators once
    python3 testdiff/run_corpus.py --pass 2     # all generators again
    python3 testdiff/run_corpus.py --report     # compare + write the report

    python3 testdiff/run_corpus.py --pass 1 --only store,links   # subset
                                              (comma names, dump_ prefix optional)

Artifacts land in testdiff/runs/corpus/ (gitignored):
    pass1.json / pass2.json   per-generator results + the full tree snapshot
    report.json               the machine-readable verdict
    report.md                 the human report (embed into the archived
                              T-64 report in docs/poc-artifacts/)

The Go replay leg (`go test -race -count=1 ./...`) and the harness leg
(`python3 testdiff/harness.py`) are driven separately; their evidence is
assembled into the archived report by hand (the runner refuses to claim
the corpus green on the generators alone — the replays are the parity
proof, the regeneration is the provenance proof).

Environment overrides (same conventions as harness.py):
    GC_PYTHON / GC_APP_DIR — the reference python + app dir (only used
    for the guard checks; each generator re-reads them itself).
"""

import argparse
import hashlib
import json
import os
import re
import subprocess
import sys
import time
from pathlib import Path

HARNESS_DIR = Path(__file__).resolve().parent
REPO_ROOT = HARNESS_DIR.parent
GOLDENS_DIR = HARNESS_DIR / "goldens"
RUNS_DIR = HARNESS_DIR / "runs" / "corpus"
PINNED_REV = (HARNESS_DIR / "PINNED-GITCURATOR-REV").read_text().strip()

DEFAULT_APP_DIR = os.environ.get("GC_APP_DIR", "/home/z/GitCurator/app")

# The canonical corpus order. dump_store precedes dump_cacherepos (the
# atomic-pair fixture reuse); everything else is alphabetical. This list
# IS the corpus inventory — a generator missing from it is a bug.
CORPUS = [
    "store",                # T-20 — first: cacherepos reuses its fixture
    "cacherepos",           # T-24
    "chrometabs",           # T-42
    "cli",                  # T-60..T-63 (the Phase-5 CLI golden)
    "config",               # T-16
    "constants",            # T-10
    "directorynote",        # T-43
    "directorysync",        # T-44
    "githubpipeline",       # T-48
    "goldenwebsites",       # SESSION-026 — the 30-link offline stone
    "goodrepos",            # T-46
    "handdelivery",         # T-41
    "helpers",              # T-17
    "links",                # T-11
    "linktracker",          # T-25
    "llmclient",            # T-30
    "mirror",               # T-50
    "notebuilder",          # T-15
    "notereads",            # T-22
    "notestate",            # T-21
    "prompts",              # T-13
    "reporadar",            # T-47
    "storage",              # T-12
    "taxonomy",             # T-14
    "vaulthygiene",         # T-49
    "vaultscan",            # T-45
    "vaultseal",            # T-47
    "webextract",           # T-32
    "webfetch",             # T-31
    "websitepipeline",      # T-40
    "websitestate",         # T-23
]


# The four ATOMIC-PAIR generators: their Python-created .db fixtures embed
# the reference's own now() timestamps, so the pair (.db + content golden)
# regenerates with fresh timestamps EVERY run — the golden's own docstring
# law ("necessarily change per run; the pair's internal consistency is what
# the golden proves, via fixture_sha256"). The corpus law for these four:
#   1. PAIR TIE — the regenerated .db's sha256 == the regenerated golden's
#      fixture_sha256 (both sides of the pair from the SAME run);
#   2. TWO-RUN STABILITY of the NORMALIZED content (every ISO-timestamp
#      masked to <TS>, fixture_sha256 dropped) — pass1 == pass2;
#   3. the COMMITTED pair stays canonical (restored after the corpus run —
#      the Go suite proves its own tie on every test run; committing fresh
#      timestamp noise every corpus run would make "committed == current"
#      meaningless for these files).
ATOMIC_PAIRS = {
    "store": ("internal/store/testdata/cache_python_created.db",
              "internal/store/testdata/cache_fixture_content.golden.json"),
    "notestate": ("internal/notestate/testdata/notestate_python_created.db",
                  "internal/notestate/testdata/notestate_fixture_content.golden.json"),
    "notereads": ("internal/notereads/testdata/notereads_python_created.db",
                  "internal/notereads/testdata/notereads_fixture_content.golden.json"),
    "websitestate": ("internal/websitepipeline/testdata/websitestate_python_created.db",
                     "internal/websitepipeline/testdata/websitestate_fixture_content.golden.json"),
}

# The ISO-timestamp mask (harness.py's timestamp-iso rule).
_TS_RX = re.compile(
    r"\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?\b")


def pair_facts(gen_name: str) -> dict:
    """Compute the tie + normalized digest for an atomic pair (post-run)."""
    db_rel, golden_rel = ATOMIC_PAIRS[gen_name]
    db_path = REPO_ROOT / db_rel
    golden_path = REPO_ROOT / golden_rel
    if not db_path.is_file() or not golden_path.is_file():
        return {"tie_ok": False, "error": "pair file missing"}
    doc = json.loads(golden_path.read_text())
    tie = sha256_file(db_path) == doc.get("fixture_sha256")
    norm = json.dumps(doc, sort_keys=True, ensure_ascii=False)
    norm = _TS_RX.sub("<TS>", norm)
    norm = norm.replace(doc.get("fixture_sha256", "<none>"), "<SHA>")
    return {
        "tie_ok": tie,
        "norm_digest": hashlib.sha256(norm.encode()).hexdigest(),
    }


def sha256_file(path: Path) -> str:
    h = hashlib.sha256()
    with open(path, "rb") as f:
        for chunk in iter(lambda: f.read(1 << 16), b""):
            h.update(chunk)
    return h.hexdigest()


def snapshot_tree() -> dict:
    """path -> sha256 for every tracked + untracked-non-ignored file."""
    proc = subprocess.run(
        ["git", "ls-files", "--cached", "--others", "--exclude-standard", "-z"],
        cwd=REPO_ROOT, capture_output=True)
    if proc.returncode != 0:
        sys.exit("[corpus] git ls-files failed: " + proc.stderr.decode())
    snap = {}
    for raw in proc.stdout.split(b"\0"):
        if not raw:
            continue
        rel = raw.decode()
        p = REPO_ROOT / rel
        if p.is_file():
            snap[rel] = sha256_file(p)
    return snap


def git_porcelain() -> str:
    proc = subprocess.run(["git", "status", "--porcelain"],
                          cwd=REPO_ROOT, capture_output=True, text=True)
    return proc.stdout


def reference_rev() -> str:
    app_dir = Path(DEFAULT_APP_DIR).resolve()
    proc = subprocess.run(["git", "-C", str(app_dir.parent), "rev-parse", "HEAD"],
                          capture_output=True, text=True)
    return proc.stdout.strip() if proc.returncode == 0 else "unknown"


def run_pass(pass_no: int, only: list) -> dict:
    gens = only if only else CORPUS
    RUNS_DIR.mkdir(parents=True, exist_ok=True)
    out = RUNS_DIR / f"pass{pass_no}.json"
    # MERGE MODE: chunks accumulate into the same pass JSON — a chunked
    # re-run never loses earlier generators' results (the --only subset
    # updates only its own entries; a full re-run refreshes everything).
    if out.exists():
        doc = json.loads(out.read_text())
        results = [g for g in doc.get("generators", [])
                   if g["name"] not in set(gens)]
    else:
        results = []
    for name in gens:
        script = GOLDENS_DIR / f"dump_{name}.py"
        if not script.exists():
            results.append({"name": name, "exit": 127, "error": "generator missing"})
            continue
        before = snapshot_tree()
        t0 = time.time()
        proc = subprocess.run([sys.executable, str(script)],
                              cwd=REPO_ROOT, capture_output=True, text=True)
        duration = round(time.time() - t0, 2)
        after = snapshot_tree()
        changed = {k: after[k] for k in after if before.get(k) != after[k]}
        removed = [k for k in before if k not in after]
        # mtime evidence: a byte-identical rewrite still touches mtime —
        # "touched but sha-unchanged" is the PROOF the generator rewrote
        # its goldens to the committed bytes (vs wrote nothing at all).
        touched = []
        for rel in after:
            p = REPO_ROOT / rel
            try:
                if p.stat().st_mtime >= t0 - 0.5:
                    touched.append(rel)
            except OSError:
                pass
        results.append({
            "name": name,
            "exit": proc.returncode,
            "duration_s": duration,
            "rev_line": next((ln for ln in proc.stdout.splitlines()
                              if "reference rev" in ln), ""),
            "stdout_tail": "\n".join(proc.stdout.splitlines()[-6:]),
            "stderr_tail": "\n".join(proc.stderr.splitlines()[-6:]),
            "wrote": changed,
            "removed": removed,
            "touched": sorted(touched),
        })
        status = "OK" if proc.returncode == 0 else f"EXIT {proc.returncode}"
        extra = ""
        if name in ATOMIC_PAIRS:
            entry_pf = pair_facts(name)
            results[-1]["pair"] = entry_pf
            extra = f" tie={'OK' if entry_pf.get('tie_ok') else 'BROKEN'}"
        print(f"[corpus p{pass_no}] {name}: {status} ({duration}s) "
              f"touched {len(touched)} churned {len(changed)}{extra}", flush=True)
    doc = {
        "pass": pass_no,
        "finished_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "reference_rev": reference_rev(),
        "pinned_rev": PINNED_REV,
        "generators": results,
        "tree_snapshot": snapshot_tree(),
        "git_status_after": git_porcelain(),
    }
    RUNS_DIR.mkdir(parents=True, exist_ok=True)
    out = RUNS_DIR / f"pass{pass_no}.json"
    out.write_text(json.dumps(doc, sort_keys=True, indent=2) + "\n")
    print(f"[corpus p{pass_no}] wrote {out} "
          f"({len(doc['tree_snapshot'])} files snapshotted)")
    return doc


def build_report() -> int:
    p1 = json.loads((RUNS_DIR / "pass1.json").read_text())
    p2 = json.loads((RUNS_DIR / "pass2.json").read_text())

    pair_files = set()
    for db_rel, golden_rel in ATOMIC_PAIRS.values():
        pair_files.add(db_rel)
        pair_files.add(golden_rel)

    failures = []
    per_gen = []
    for g1 in p1["generators"]:
        g2 = next((g for g in p2["generators"] if g["name"] == g1["name"]), None)
        if g2 is None:
            failures.append(f"{g1['name']}: missing from pass 2")
            continue
        entry = {
            "name": g1["name"],
            "exit1": g1["exit"], "exit2": g2["exit"],
            "dur1": g1.get("duration_s"), "dur2": g2.get("duration_s"),
            "rev_line": g1.get("rev_line") or g2.get("rev_line", ""),
        }
        if g1["exit"] != 0 or g2["exit"] != 0:
            failures.append(f"{g1['name']}: exit {g1['exit']}/{g2['exit']}")
        # For the four pair generators the pair files' shas legitimately
        # differ per pass (fresh timestamps); their two-run proof is the
        # NORMALIZED digest below. Everything else they wrote must still
        # be byte-identical across passes.
        def nonpair_wrote(g):
            return {k: v for k, v in g.get("wrote", {}).items()
                    if k not in pair_files} | {"__removed__": g.get("removed", [])}
        if nonpair_wrote(g1) != nonpair_wrote(g2):
            failures.append(f"{g1['name']}: pass1/pass2 wrote DIFFERENT bytes "
                            f"(two-run proof broken)")
            entry["two_run"] = "DIVERGED"
        else:
            entry["two_run"] = "identical"
            entry["wrote_n"] = len(g1.get("wrote", {}))
        if g1["exit"] == 0 and g2["exit"] == 0 and not g1.get("wrote") \
                and not g1.get("removed") and g2.get("wrote") == {}:
            # no churn on either pass: the generator rewrote its files
            # byte-identically to the COMMITTED state.
            entry["committed"] = "byte-identical (no tree churn)"
        t1, t2 = g1.get("touched", []), g2.get("touched", [])
        entry["touched_n"] = len(t1)
        if not t1 or not t2:
            # exit 0 but nothing touched: the generator silently wrote
            # nothing — flag it (a healthy generator rewrites its goldens).
            # (mtime artifacts can inflate a touched set, but never empty
            # it — the silent case is unambiguous.)
            if g1["exit"] == 0 and g2["exit"] == 0:
                failures.append(f"{g1['name']}: exited 0 but touched NO files "
                                f"(silent generator — wrote nothing)")
        if g1["name"] in ATOMIC_PAIRS:
            pf1, pf2 = g1.get("pair", {}), g2.get("pair", {})
            entry["pair_tie"] = ("OK" if pf1.get("tie_ok") and pf2.get("tie_ok")
                                 else "BROKEN")
            if not pf1.get("tie_ok") or not pf2.get("tie_ok"):
                failures.append(f"{g1['name']}: atomic-pair TIE broken "
                                f"(regenerated .db sha != golden's "
                                f"fixture_sha256)")
            elif pf1.get("norm_digest") != pf2.get("norm_digest"):
                failures.append(f"{g1['name']}: atomic-pair NORMALIZED content "
                                f"differs between passes (two-run proof "
                                f"broken beyond timestamps)")
                entry["pair_two_run"] = "DIVERGED"
            else:
                entry["pair_two_run"] = "identical (normalized)"
        per_gen.append(entry)

    snap1, snap2 = p1["tree_snapshot"], p2["tree_snapshot"]
    # The two-run snapshot proof covers every file EXCEPT the four atomic
    # pairs (whose timestamp churn is the documented pair law; their
    # stability is proven by the per-pair normalized digest above).
    snap_diff = {
        "only_in_1": sorted(k for k in snap1
                            if k not in snap2 and k not in pair_files),
        "only_in_2": sorted(k for k in snap2
                            if k not in snap1 and k not in pair_files),
        "sha_mismatch": sorted(k for k in snap1 if k in snap2
                               and snap1[k] != snap2[k]
                               and k not in pair_files),
    }
    if any(snap_diff.values()):
        failures.append(f"two-run tree snapshot differs: {snap_diff}")

    porcelain2 = p2["git_status_after"]
    # The documented exception: the four atomic pairs are EXPECTED to churn
    # (fresh now() timestamps every run — the pair law). Everything else
    # must leave the tree clean.
    unexpected = [ln for ln in porcelain2.splitlines()
                  if ln[3:] not in pair_files]
    if unexpected:
        failures.append("tree NOT clean after pass 2 beyond the four atomic "
                        "pairs (committed goldens diverge from regenerated): "
                        + " | ".join(unexpected))

    rev_ok = p1["reference_rev"] == PINNED_REV == p2["reference_rev"]

    doc = {
        "generated_at": time.strftime("%Y-%m-%dT%H:%M:%S%z"),
        "reference_rev": p1["reference_rev"],
        "pinned_rev": PINNED_REV,
        "rev_guard": "OK" if rev_ok else "MISMATCH",
        "generators": per_gen,
        "two_run_snapshot_diff": snap_diff,
        "git_status_after_pass2": porcelain2,
        "verdict": "GREEN" if not failures and rev_ok else "DIVERGED",
        "failures": failures,
    }
    (RUNS_DIR / "report.json").write_text(
        json.dumps(doc, sort_keys=True, indent=2) + "\n")

    lines = [
        "# T-64 corpus full-run — machine report",
        "",
        f"- verdict: **{doc['verdict']}**",
        f"- reference rev: {doc['reference_rev']} (pinned {PINNED_REV}, "
        f"guard {doc['rev_guard']})",
        f"- generators driven: {len(per_gen)} × two passes",
        f"- two-run tree diff: {snap_diff or 'none'}",
        f"- git status after pass 2: {porcelain2 or 'CLEAN'} "
        f"(the four atomic pairs' timestamp churn is the documented "
        f"exception; the committed pairs stay canonical)",
        "",
        "| generator | exit p1/p2 | two-run | touched | churned |",
        "|---|---|---|---|---|",
    ]
    for e in per_gen:
        lines.append(f"| {e['name']} | {e['exit1']}/{e['exit2']} "
                     f"| {e.get('two_run', '?')} | {e.get('touched_n', '?')} "
                     f"| {e.get('wrote_n', 0)} |")
    if failures:
        lines += ["", "## Failures", ""]
        lines += [f"- {f}" for f in failures]
    (RUNS_DIR / "report.md").write_text("\n".join(lines) + "\n")
    print("\n".join(lines))
    return 0 if doc["verdict"] == "GREEN" else 1


def main() -> int:
    ap = argparse.ArgumentParser()
    ap.add_argument("--pass", dest="pass_no", type=int, choices=[1, 2])
    ap.add_argument("--report", action="store_true")
    ap.add_argument("--only", default="",
                    help="comma-separated generator names (subset)")
    args = ap.parse_args()

    if args.report:
        return build_report()
    if args.pass_no:
        only = [g for g in args.only.split(",") if g] if args.only else []
        # tolerate both bare names and dump_-prefixed script names
        only = [g[5:] if g.startswith("dump_") else g for g in only]
        unknown = [g for g in only if g not in CORPUS]
        if unknown:
            sys.exit(f"[corpus] unknown generator(s): {unknown}")
        run_pass(args.pass_no, only)
        return 0
    ap.print_help()
    return 2


if __name__ == "__main__":
    raise SystemExit(main())
