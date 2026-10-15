#!/usr/bin/env python3
"""testdiff — the differential (Python-vs-Go) harness skeleton.

Drives the UNMODIFIED Python app (assadigit/GitCurator, pinned clone) and
the Go port on the same inputs, in the same scratch working directory, then
compares stdout / stderr / exit code after the normalization rules from
docs/port-blueprint/09-testing-and-parity.md §3:

  - timestamps (HH:MM:SS and ISO shapes)        -> <TS>
  - the run's scratch directory path            -> <TMP>
  - durations (123ms / 1.2 s / 0.5s)            -> <DUR>
  - version strings inside UA-bearing logs      -> <VER>

Must match exactly after normalization: note file bytes (later phases), DB
contents, manifests, log verdict lines, exit codes — and for this Phase-0
skeleton: stdout/stderr of the CLI seed surface.

Usage:
    python3 testdiff/harness.py [cases/status-no-config ...]

Environment overrides:
    GC_PYTHON   python that runs the reference app
                (default /home/z/venvs/gitcurator/bin/python)
    GC_APP_DIR  the reference app directory      (default /home/z/GitCurator/app)
    GO_BIN      prebuilt Go CLI binary (skips the harness's own `go build`)
    GO          go toolchain                     (default from PATH)

The harness records the reference repo's git rev per run (and warns when it
drifts from testdiff/PINNED-GITCURATOR-REV). Reports land in
testdiff/runs/<case>/report.md (gitignored); copy the signed-off report
into docs/poc-artifacts/ when it is the evidence for a gate.
"""

import hashlib
import json
import os
import re
import subprocess
import sys
from datetime import datetime, timezone
from pathlib import Path

HARNESS_DIR = Path(__file__).resolve().parent
REPO_ROOT = HARNESS_DIR.parent
PINNED_REV = (HARNESS_DIR / "PINNED-GITCURATOR-REV").read_text().strip()

DEFAULT_PY = os.environ.get("GC_PYTHON", "/home/z/venvs/gitcurator/bin/python")
DEFAULT_APP_DIR = os.environ.get("GC_APP_DIR", "/home/z/GitCurator/app")

# ---------------------------------------------------------------------------
# normalization (doc 09 §3 — the documented rule set)
# ---------------------------------------------------------------------------

RULES = [
    ("timestamp", re.compile(r"\b\d{2}:\d{2}:\d{2}\b"), "<TS>"),
    ("timestamp-iso", re.compile(r"\b\d{4}-\d{2}-\d{2}[T ]\d{2}:\d{2}:\d{2}(?:\.\d+)?(?:Z|[+-]\d{2}:?\d{2})?\b"), "<TS>"),
    ("duration", re.compile(r"\b\d+(?:\.\d+)?\s?(?:ms|µs|us|s)\b"), "<DUR>"),
    ("ua-version", re.compile(r"\bGitCurator/v[0-9][0-9A-Za-z.\-+]*"), "<VER>"),
]


def normalize(text: str, scratch: str) -> tuple:
    """Apply the documented normalization rules; returns (normalized, rules_applied)."""
    applied = []
    for name, rx, repl in RULES:
        if rx.search(text):
            text = rx.sub(repl, text)
            applied.append(name)
    if scratch and scratch in text:
        text = text.replace(scratch, "<TMP>")
        applied.append("temp-path")
    return text, applied


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


# ---------------------------------------------------------------------------
# drivers
# ---------------------------------------------------------------------------

def build_go_binary() -> Path:
    out = HARNESS_DIR / "runs" / "bin" / "gitcurator-cli"
    out.parent.mkdir(parents=True, exist_ok=True)
    go = os.environ.get("GO", "go")
    cmd = [go, "build", "-o", str(out), "./cmd/gitcurator-cli"]
    proc = subprocess.run(cmd, cwd=REPO_ROOT, capture_output=True, text=True)
    if proc.returncode != 0:
        sys.exit(f"[harness] go build failed:\n{proc.stderr}")
    return out


def run_side(cmd, cwd: str, env: dict) -> dict:
    proc = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True)
    return {
        "cmd": cmd,
        "exit": proc.returncode,
        "stdout": proc.stdout.decode("utf-8", errors="replace"),
        "stderr": proc.stderr.decode("utf-8", errors="replace"),
    }


# ---------------------------------------------------------------------------
# comparison + report
# ---------------------------------------------------------------------------

def unified_diff(a: str, b: str, a_name: str, b_name: str) -> str:
    import difflib
    lines = list(difflib.unified_diff(
        a.splitlines(keepends=True), b.splitlines(keepends=True),
        fromfile=a_name, tofile=b_name, n=3))
    return "".join(lines)


def run_case(case_dir: Path, py: str, app_dir: str, go_bin: Path):
    case = json.loads((case_dir / "case.json").read_text())
    name = case.get("name", case_dir.name)

    runs_dir = HARNESS_DIR / "runs" / case_dir.name
    runs_dir.mkdir(parents=True, exist_ok=True)
    scratch = runs_dir / "scratch"
    if scratch.exists():
        import shutil
        shutil.rmtree(scratch)
    scratch.mkdir(parents=True)

    env = dict(os.environ)
    env.update(case.get("env", {}))
    env["PYTHONDONTWRITEBYTECODE"] = "1"  # keep the reference repo pristine
    env.setdefault("PYTHONIOENCODING", "utf-8")
    env.setdefault("NO_COLOR", "1")

    py_args = case["python"]["args"]
    go_args = case["go"]["args"]

    py_run = run_side([py, str(Path(app_dir) / "main.py")] + py_args, str(scratch), env)
    go_run = run_side([str(go_bin)] + go_args, str(scratch), env)

    norm_py_out, rules_a = normalize(py_run["stdout"], str(scratch))
    norm_go_out, rules_b = normalize(go_run["stdout"], str(scratch))
    norm_py_err, _ = normalize(py_run["stderr"], str(scratch))
    norm_go_err, _ = normalize(go_run["stderr"], str(scratch))

    diffs = []
    if norm_py_out != norm_go_out:
        diffs.append(("stdout", unified_diff(norm_py_out, norm_go_out, "python", "go")))
    if norm_py_err != norm_go_err:
        diffs.append(("stderr", unified_diff(norm_py_err, norm_go_err, "python.stderr", "go.stderr")))
    if py_run["exit"] != go_run["exit"]:
        diffs.append(("exit-code", "python exited %d, go exited %d" % (py_run["exit"], go_run["exit"])))
    ok = not diffs

    # reference repo rev record (ADR-0002 Am.1: exact git rev per run)
    try:
        ref_rev = subprocess.run(
            ["git", "-C", str(Path(app_dir).parent), "rev-parse", "HEAD"],
            capture_output=True, text=True).stdout.strip()
    except Exception:
        ref_rev = "<unknown>"
    rev_warning = "" if ref_rev == PINNED_REV else (
        "\n> ⚠️ reference rev %s != pinned %s — results are not comparable to the pinned corpus\n"
        % (ref_rev[:12], PINNED_REV[:12]))

    now = datetime.now(timezone.utc).strftime("%Y-%m-%d %H:%M:%SZ")
    rules_used = sorted(set(rules_a) | set(rules_b))
    lines = [
        "# Differential report — %s" % name,
        "",
        "- Run: %s (UTC)" % now,
        "- Reference: %s (pinned: %s)%s" % (ref_rev, PINNED_REV, rev_warning),
        "- Go binary: %s (sha256 %s…)" % (go_bin, sha256(go_bin.read_bytes())[:16]),
        "- Python: %s" % py,
        "- Scratch cwd: `%s` (normalized to `<TMP>` where it appears)" % scratch,
        "- Normalization rules applied: %s" % (", ".join(rules_used) or "(none needed)"),
        "",
        "## Commands (same scratch cwd, same env)",
        "",
        "```",
        "python: %s" % " ".join(py_run["cmd"]),
        "go:     %s" % " ".join(go_run["cmd"]),
        "```",
        "",
        "## Raw outputs (sha256)",
        "",
        "- python stdout: `%s` (%d bytes)" % (sha256(py_run["stdout"].encode()), len(py_run["stdout"])),
        "- go stdout:     `%s` (%d bytes)" % (sha256(go_run["stdout"].encode()), len(go_run["stdout"])),
        "- python stderr: `%s` (%d bytes)" % (sha256(py_run["stderr"].encode()), len(py_run["stderr"])),
        "- go stderr:     `%s` (%d bytes)" % (sha256(go_run["stderr"].encode()), len(go_run["stderr"])),
        "",
        "## Verdict",
        "",
        "**%s**" % ("PASS — outputs identical after normalization; exit codes match" if ok else "FAIL"),
        "",
    ]
    if ok:
        lines += [
            "Normalized stdout (python side shown; go side byte-identical after the rules above):",
            "",
            "```",
            norm_py_out,
            "```",
            "",
            "Normalized stderr (both sides):",
            "",
            "```",
            norm_py_err or "(empty)",
            "```",
        ]
    else:
        for what, d in diffs:
            lines += ["## Diff — " + what, "", "```diff", d, "```", ""]
    report = runs_dir / "report.md"
    report.write_text("\n".join(lines))
    return ok, report


def main() -> int:
    case_roots = sys.argv[1:] or [str(HARNESS_DIR / "cases" / "status-no-config")]
    py = DEFAULT_PY
    app_dir = DEFAULT_APP_DIR
    go_bin = Path(os.environ["GO_BIN"]) if os.environ.get("GO_BIN") else build_go_binary()

    all_ok = True
    for c in case_roots:
        case_dir = Path(c)
        ok, report = run_case(case_dir, py, app_dir, go_bin)
        all_ok &= ok
        print("[testdiff] %s: %s  → %s" % (case_dir.name, "PASS" if ok else "FAIL", report))
    return 0 if all_ok else 1


if __name__ == "__main__":
    raise SystemExit(main())
