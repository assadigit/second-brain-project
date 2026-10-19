#!/usr/bin/env python3
"""goldens — T-72: the theme-token golden generator.

Dumps the reference's ONE app theme kit's token tables (gui/theme.py:58-237,
LIGHT + DARK — every color the app paints with) as canonical JSON for the
frontend foundation's golden tests (spec-gui-theme-icons §1; the Go/CSS
twins byte-diff against this golden).

Usage (from the repo root):

    python3 testdiff/goldens/dump_theme.py [--out PATH]

Environment overrides (same conventions as ../harness.py):
    GC_PYTHON   python that runs the reference app
                (default /home/z/venvs/gitcurator/bin/python)
    GC_APP_DIR  the reference app directory (default /home/z/GitCurator/app)

The import stub (harness-side only — the reference repo stays UNMODIFIED):
gui/theme.py imports QColor/QPalette from PyQt6.QtGui at import time but
only USES them inside its QSS/palette functions; the venv is stdlib-only
(PyQt6 is the GUI runtime the port replaces). The payload therefore
satisfies the import with attribute stubs before loading the module; the
tables themselves are pure data and dump byte-faithfully.
"""

import json
import os
import subprocess
import sys
from pathlib import Path

HARNESS_DIR = Path(__file__).resolve().parent.parent
REPO_ROOT = HARNESS_DIR.parent
PINNED_REV = (HARNESS_DIR / "PINNED-GITCURATOR-REV").read_text().strip()

DEFAULT_PY = os.environ.get("GC_PYTHON", "/home/z/venvs/gitcurator/bin/python")
DEFAULT_APP_DIR = os.environ.get("GC_APP_DIR", "/home/z/GitCurator/app")

# The payload executed BY the reference python (unmodified app on sys.path).
# Dumps the LIGHT + DARK token tables as one canonical JSON document.
PAYLOAD = r'''
import json, os, sys

def _stub(name, **attrs):
    mod = type(sys)(name)
    for k, v in attrs.items():
        setattr(mod, k, v)
    sys.modules[name] = mod
    return mod

# gui/theme.py imports these at import time (used only inside functions).
pkg = _stub("PyQt6")
_stub("PyQt6.QtGui", QColor=object, QPalette=object)

sys.path.insert(0, os.environ["GC_APP_DIR"])
from gitcurator.gui import theme as t

def canon(obj):
    return json.dumps(obj, sort_keys=True, ensure_ascii=False, indent=2) + "\n"

doc = {
    "schema": 1,
    "light": t.LIGHT,
    "dark": t.DARK,
}
out_dir = os.environ["GOLDEN_OUT_DIR"]
with open(os.path.join(out_dir, "theme_tokens.golden.json"), "w", encoding="utf-8") as f:
    f.write(canon(doc))

print(json.dumps({
    "light_tokens": len(t.LIGHT),
    "dark_tokens": len(t.DARK),
}))
'''

def reference_rev(app_dir: str) -> str:
    try:
        proc = subprocess.run(
            ["git", "-C", str(Path(app_dir).parent), "rev-parse", "HEAD"],
            capture_output=True, text=True)
        if proc.returncode == 0:
            return proc.stdout.strip()
    except OSError:
        pass
    return "unknown"


def main() -> int:
    out_default = REPO_ROOT / "webui" / "uitheme" / "testdata"
    args = sys.argv[1:]
    out_dir = out_default
    if args[:2] == ["--out"]:
        out_dir = Path(args[2])
    out_dir.mkdir(parents=True, exist_ok=True)

    app_dir = Path(DEFAULT_APP_DIR).resolve()
    if not (app_dir / "gitcurator" / "gui" / "theme.py").exists():
        sys.exit(f"[goldens] reference app not found at {app_dir}")
    rev = reference_rev(str(app_dir))
    if rev != PINNED_REV:
        print(f"[goldens] WARNING: reference rev {rev} != pinned {PINNED_REV}",
              file=sys.stderr)

    env = dict(os.environ)
    env["GC_APP_DIR"] = str(app_dir)
    env["GOLDEN_OUT_DIR"] = str(out_dir)
    env["PYTHONDONTWRITEBYTECODE"] = "1"  # keep the reference repo pristine
    proc = subprocess.run([DEFAULT_PY, "-c", PAYLOAD], env=env,
                           capture_output=True, text=True)
    if proc.returncode != 0:
        sys.exit(f"[goldens] reference run failed:\n{proc.stderr}")
    summary = json.loads(proc.stdout)
    print(f"[goldens] generated {out_dir}/theme_tokens.golden.json")
    print(f"[goldens] reference rev: {rev} (pinned: {PINNED_REV})")
    print(f"[goldens] theme tokens: {summary['light_tokens']} LIGHT / "
          f"{summary['dark_tokens']} DARK")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
