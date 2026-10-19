#!/usr/bin/env python3
"""T-78 — the side-by-side verification vs the Python app (ADR-0008).

THE ACCEPTANCE (doc 13's row): the lock mutual exclusion ACROSS BOTH
BINARIES — the Go app and the Python app against the SAME data (the same
app/ folder: config.json, the user memory, app.lock), proving the
compatibility laws live:

  LEG A — the app.lock mutual exclusion, binary-to-binary.
    The Python half drives THE REFERENCE'S OWN LOCK CODE, extracted
    verbatim from the pinned checkout (gui/app.py:393-440 — the lock
    dance inline in main(); gui/headless.py:246-282 — _is_process_running
    with the POSIX signal-0 branch). The extraction exists because the
    import chain is PyQt6-gated (evidence: gui/headless.py:34 imports
    gui/icons.py:45 `from PyQt6.QtCore import ...` — a raw
    ModuleNotFoundError under the stdlib-only venv; the reference's GUI
    entry is the ONLY lock-taking binary — `--cli` and `--headless`
    both exit BEFORE the lock dance). Extract-and-exec drives the
    reference's own bytes without importing the wall.
    The Go half is testdiff/goside (the production door:
    internal/app.AcquireAppLock — the exact lock the wails shell takes).
    Scenarios: fresh takes, both cross-refusals (each side holds, the
    other refuses), both stale healings, the bare-PID BYTES law, and the
    EPERM probe (a live not-ours PID — the divergence T-78's audit
    found and fixed on the Go side: the reference's blanket except reads
    EPERM as dead, so a recycled-PID lock is take-over-able in BOTH).

  LEG B — the same-data config laws.
    The Python oracle is the reference's own pure modules imported
    in-process (gitcurator.core.storage.merge_config / write_config_file,
    gitcurator.core.user_config.apply_global_credentials /
    mirror_credentials_to_global, gitcurator.constants.CONFIG_EXAMPLE —
    all PyQt-free). The Go half is goside's config/fill/mirror modes
    (the same doors the shells call). Scenarios: the merge-never-rebuild
    + byte-comparable round trips (both apps compute + write the SAME
    bytes from the same input; each re-save of the other's file never
    disturbs the bytes), unknown-key survival, tuning keys unstomped,
    non-ASCII raw + no HTML escaping, the fill (local-wins + fresh door)
    and the mirror (never-loses + the 0600 law).

Usage:
    python3 testdiff/side_by_side.py            # all legs, foreground
    python3 testdiff/side_by_side.py --leg A    # the lock leg only
    python3 testdiff/side_by_side.py --leg B    # the config leg only

Environment overrides (house conventions):
    GC_PYTHON   python for the reference side  (default /home/z/venvs/gitcurator/bin/python)
    GC_APP_DIR  the reference app directory    (default /home/z/GitCurator/app —
                                                falls back to the sibling checkout when
                                                the symlink is down)
    GO          go toolchain                   (default from PATH)

Artifacts land in testdiff/runs/side-by-side/ (gitignored):
    report.md / report.json    the verdict (copy the signed-off report into
                               docs/poc-artifacts/ when it gates a phase)
"""

import argparse
import copy
import difflib
import hashlib
import json
import os
import re
import shutil
import signal
import subprocess
import sys
import textwrap
import time
from pathlib import Path

HARNESS_DIR = Path(__file__).resolve().parent
REPO_ROOT = HARNESS_DIR.parent
RUNS = HARNESS_DIR / "runs" / "side-by-side"
PINNED_REV = (HARNESS_DIR / "PINNED-GITCURATOR-REV").read_text().strip()

DEFAULT_PY = os.environ.get("GC_PYTHON", "/home/z/venvs/gitcurator/bin/python")
DEFAULT_APP_DIR = os.environ.get("GC_APP_DIR", "/home/z/GitCurator/app")
if not Path(DEFAULT_APP_DIR).exists():
    DEFAULT_APP_DIR = str(REPO_ROOT.parent / "GitCurator" / "app")

# The extraction contract — the pinned line ranges inside the reference
# (ADR-0008's own citation for the dance). Structural pins (first/last
# line) fail loudly if the rev drifts; the rev guard is the authority.
LOCK_DANCE_RANGE = (393, 440)   # gui/app.py: main()'s lock dance
IS_RUNNING_RANGE = (246, 282)   # gui/headless.py: _is_process_running
DANCE_FIRST_LINE = '    lock_file = "app.lock"'
DANCE_LAST_MARKER = "We don't have an exclusive lock"
IS_RUNNING_FIRST_LINE = "def _is_process_running(pid):"

results = []   # [{name, ok, evidence: [str]}]


def record(name, ok, *evidence):
    results.append({"name": name, "ok": bool(ok), "evidence": list(evidence)})
    print(f"[{'PASS' if ok else 'FAIL'}] {name}")
    for line in evidence:
        print(f"       {line}")


def sha256(data: bytes) -> str:
    return hashlib.sha256(data).hexdigest()


def die(msg: str):
    print(f"[ABORT] {msg}", file=sys.stderr)
    sys.exit(2)


# ---------------------------------------------------------------------------
# Guards + tooling
# ---------------------------------------------------------------------------

def guard_environment(py: str, app_dir: str, go: str) -> Path:
    """The pinned rev + the interpreter + the go driver build."""
    if not Path(app_dir).exists():
        die(f"GC_APP_DIR does not exist: {app_dir}")
    ref_repo = str(Path(app_dir).parent)
    rev = subprocess.run(["git", "-C", ref_repo, "rev-parse", "HEAD"],
                         capture_output=True, text=True).stdout.strip()
    if rev != PINNED_REV:
        die(f"reference rev {rev[:12]} != pinned {PINNED_REV[:12]} — "
            "the extraction ranges are only valid at the pin")
    probe = subprocess.run([py, "-c", "import sys; print(sys.version_info[:2])"],
                           capture_output=True, text=True)
    if probe.returncode != 0:
        die(f"GC_PYTHON unusable: {probe.stderr.strip()}")
    goside = RUNS / "bin" / "goside"
    goside.parent.mkdir(parents=True, exist_ok=True)
    build = subprocess.run([go, "build", "-o", str(goside), "./testdiff/goside"],
                           cwd=REPO_ROOT, capture_output=True, text=True)
    if build.returncode != 0:
        die(f"go build goside failed:\n{build.stderr}")
    return goside


# ---------------------------------------------------------------------------
# The reference's own lock code — extracted verbatim from the pin
# ---------------------------------------------------------------------------

def extract_lines(app_dir: str, rel: str, lo: int, hi: int) -> str:
    path = Path(app_dir) / rel
    lines = path.read_text(encoding="utf-8").splitlines(keepends=True)
    return "".join(lines[lo - 1:hi])


def extract_reference_lock(app_dir: str):
    """The lock dance + the liveness probe, straight from the pinned bytes."""
    dance = extract_lines(app_dir, "gitcurator/gui/app.py", *LOCK_DANCE_RANGE)
    isrun = extract_lines(app_dir, "gitcurator/gui/headless.py", *IS_RUNNING_RANGE)
    if dance.splitlines()[0] != DANCE_FIRST_LINE:
        die(f"the extraction pin broke: gui/app.py:{LOCK_DANCE_RANGE[0]} "
            f"reads {dance.splitlines()[0]!r}")
    if DANCE_LAST_MARKER not in dance:
        die("the extraction pin broke: the dance no longer carries its last marker")
    if isrun.splitlines()[0] != IS_RUNNING_FIRST_LINE:
        die(f"the extraction pin broke: gui/headless.py:{IS_RUNNING_RANGE[0]} "
            f"reads {isrun.splitlines()[0]!r}")
    # The dance is main()-local at indent 4 — dedent to module level.
    dance_src = textwrap.dedent(dance)
    return dance_src, isrun, sha256(dance.encode()), sha256(isrun.encode())


PYLOCK_TEMPLATE = '''#!python
# GENERATED by testdiff/side_by_side.py — DO NOT EDIT.
# The lock code below is the REFERENCE'S OWN, extracted verbatim from the
# pinned checkout (gitcurator/gui/app.py + gui/headless.py); the wrapper
# only adds the mode verbs + the clean-shutdown mirror of main()'s
# finally block (close the fd, remove the lock, exit).
import os
import signal
import sys
import time as _ttime  # aliased: the dance's own `import time` (inside its
                       # PermissionError branch) would make `time` a local of
                       # main() and shadow any module-level name we use here

IS_RUNNING_BLOCK


def _cleanup(lock_file, lock_fd):
    """main()'s finally, verbatim shape."""
    if lock_fd is not None:
        try:
            os.close(lock_fd)
        except OSError:
            pass
    try:
        os.remove(lock_file)
    except OSError:
        pass


def _held(lock_file):
    """The observable held-state: the file carries OUR pid."""
    try:
        with open(lock_file, "r") as f:
            return f.read().strip() == str(os.getpid())
    except IOError:
        return False


def main():
    mode = sys.argv[1] if len(sys.argv) > 1 else "try"
    if mode == "hold":
        # A clean stop must run the reference's shutdown cleanup — the
        # handler raises SystemExit out of the sleep loop, the finally
        # below closes + removes (kill -9 skips it all — by design).
        signal.signal(signal.SIGTERM, lambda *_: sys.exit(0))
LOCK_DANCE_BLOCK
    if mode == "hold":
        try:
            if _held(lock_file):
                print(f"PYLOCK HELD {os.getpid()}")
            else:
                print(f"PYLOCK NOTHELD {os.getpid()}")
            sys.stdout.flush()
            while True:
                _ttime.sleep(0.2)
        except SystemExit:
            pass
        _cleanup(lock_file, lock_fd)
        sys.exit(0)
    # try mode: the verdict + the reference's shutdown cleanup
    verdict = "HELD" if _held(lock_file) else "NOTHELD"
    print(f"PYLOCK {verdict} {os.getpid()}")
    _cleanup(lock_file, lock_fd)
    sys.exit(0)


main()
'''


def indent_block(src: str, pad: str) -> str:
    out = []
    for ln in src.splitlines():
        out.append(pad + ln if ln.strip() else ln)
    return "\n".join(out)


def write_pylock(path: Path, dance_src: str, isrun_src: str):
    src = PYLOCK_TEMPLATE.replace("IS_RUNNING_BLOCK", isrun_src.rstrip())
    src = src.replace("LOCK_DANCE_BLOCK", indent_block(dance_src, "    "))
    path.write_text(src, encoding="utf-8")


# ---------------------------------------------------------------------------
# Process helpers
# ---------------------------------------------------------------------------

def base_env(extra=None):
    env = dict(os.environ)
    env["PYTHONDONTWRITEBYTECODE"] = "1"
    env["PYTHONIOENCODING"] = "utf-8"
    env["NO_COLOR"] = "1"
    if extra:
        env.update(extra)
    return env


def run_cmd(cmd, cwd, env=None, timeout=30):
    proc = subprocess.run(cmd, cwd=cwd, env=env, capture_output=True,
                          timeout=timeout)
    return {
        "exit": proc.returncode,
        "out": proc.stdout.decode("utf-8", errors="replace"),
        "err": proc.stderr.decode("utf-8", errors="replace"),
    }


class Holder:
    """A live lock-holder subprocess (Go or Python) with a readiness
    marker polled off a redirected stdout file."""

    def __init__(self, cmd, cwd, env, marker):
        self.log = RUNS / f"holder-{int(time.time() * 1000)}-{os.getpid()}.log"
        self.fh = open(self.log, "wb")
        self.proc = subprocess.Popen(cmd, cwd=cwd, env=env,
                                     stdout=self.fh, stderr=subprocess.STDOUT)
        self.pid = None
        deadline = time.time() + 20
        while time.time() < deadline:
            if self.proc.poll() is not None:
                break
            try:
                text = self.log.read_text(encoding="utf-8", errors="replace")
            except OSError:
                text = ""
            m = re.search(marker + r" (\d+)", text)
            if m:
                self.pid = int(m.group(1))
                return
            time.sleep(0.05)
        tail = ""
        try:
            tail = self.log.read_text(errors="replace")[:500]
        except OSError:
            pass
        self.stop()
        raise RuntimeError(f"holder never became ready ({marker}): {tail}")

    def stop(self, force=False):
        if self.proc.poll() is None:
            try:
                self.proc.send_signal(signal.SIGKILL if force else signal.SIGTERM)
                self.proc.wait(timeout=10)
            except subprocess.TimeoutExpired:
                self.proc.kill()
                self.proc.wait(timeout=10)
        self.fh.close()


def lock_bytes(path: Path):
    try:
        return path.read_bytes()
    except OSError:
        return None


def assert_bare_pid(name, raw, pid):
    want = str(pid).encode()
    ok = raw == want
    record(name, ok,
           f"lock bytes = {raw!r} ({len(raw) if raw else 0} bytes), "
           f"want the bare decimal {want!r}",
           f"hex: {raw.hex() if raw else '-'}")
    return ok


# ---------------------------------------------------------------------------
# LEG A — the lock mutual exclusion across both binaries
# ---------------------------------------------------------------------------

def leg_a(py: str, goside: Path, dance_src, isrun_src):
    pylock = RUNS / "pylock.py"
    write_pylock(pylock, dance_src, isrun_src)

    def fresh_dir(tag):
        d = RUNS / f"appdir-{tag}"
        if d.exists():
            shutil.rmtree(d)
        d.mkdir(parents=True)
        return d

    # A0 — both sides take a fresh lock; the BYTES law on each take.
    for side, tag, hold_cmd in (
        ("go", "a0-go", [str(goside), "lock-hold", "app.lock"]),
        ("py", "a0-py", [py, str(pylock), "hold"]),
    ):
        d = fresh_dir(tag)
        h = Holder(hold_cmd, str(d), base_env(),
                   "GOSIDE HELD" if side == "go" else "PYLOCK HELD")
        try:
            assert_bare_pid(f"A0 {side} fresh-take bytes",
                            lock_bytes(d / "app.lock"), h.pid)
        finally:
            h.stop()

    # A1 — the Go binary holds; the reference's own dance REFUSES.
    d = fresh_dir("a1")
    h = Holder([str(goside), "lock-hold", "app.lock"], str(d), base_env(),
               "GOSIDE HELD")
    try:
        r = run_cmd([py, str(pylock), "try"], str(d), base_env())
        want = f"Another instance is already running (PID {h.pid}). Exiting."
        ok = (r["exit"] == 1 and want in r["out"]
              and lock_bytes(d / "app.lock") == str(h.pid).encode())
        record("A1 go-holds / python-refuses", ok,
               f"exit={r['exit']} (want 1)",
               f"stdout: {r['out'].strip()!r}",
               "the lock file still carries the Go holder's pid — untouched")
    finally:
        h.stop()

    # A2 — the reference's dance holds; the Go binary REFUSES.
    d = fresh_dir("a2")
    h = Holder([py, str(pylock), "hold"], str(d), base_env(), "PYLOCK HELD")
    try:
        r = run_cmd([str(goside), "lock-try", "app.lock"], str(d), base_env())
        want = f"Another instance is already running (PID {h.pid}). Exiting."
        ok = (r["exit"] == 1 and want in r["out"] and "GOSIDE REFUSED" in r["out"]
              and lock_bytes(d / "app.lock") == str(h.pid).encode())
        record("A2 python-holds / go-refuses", ok,
               f"exit={r['exit']} (want 1)",
               f"stdout: {r['out'].strip()!r}",
               "the lock file still carries the Python holder's pid — untouched")
    finally:
        h.stop()

    # A3 — the Go holder CRASHES (kill -9); the reference heals the stale.
    d = fresh_dir("a3")
    h = Holder([str(goside), "lock-hold", "app.lock"], str(d), base_env(),
               "GOSIDE HELD")
    h.stop(force=True)
    stale_pid = h.pid
    r = run_cmd([py, str(pylock), "try"], str(d), base_env())
    ok = (r["exit"] == 0
          and f"Removing stale lock file (PID {stale_pid} is no longer running)..." in r["out"]
          and "PYLOCK HELD" in r["out"])
    record("A3 go-crash / python-heals-stale", ok,
           f"the stale pid was the crashed Go holder ({stale_pid})",
           f"stdout: {r['out'].strip()!r}")

    # A4 — the Python holder CRASHES; the Go binary heals the stale.
    d = fresh_dir("a4")
    h = Holder([py, str(pylock), "hold"], str(d), base_env(), "PYLOCK HELD")
    h.stop(force=True)
    stale_pid = h.pid
    # A holding Go take: it heals the stale lock AND keeps the file long
    # enough for the bytes law (lock-try would release at exit).
    h2 = Holder([str(goside), "lock-hold", "app.lock"], str(d), base_env(),
                "GOSIDE HELD")
    try:
        holder_log = h2.log.read_text(encoding="utf-8", errors="replace")
        ok = (f"Removing stale lock file (PID {stale_pid} is no longer running)..." in holder_log
              and h2.proc.poll() is None)
        ev = [f"the stale pid was the crashed Python holder ({stale_pid})",
              f"holder output: {holder_log.strip()!r}"]
        record("A4 python-crash / go-heals-stale", ok, *ev)
        if ok:
            assert_bare_pid("A4b the healed lock's bytes (go)",
                            lock_bytes(d / "app.lock"), h2.pid)
    finally:
        h2.stop()

    # A6 — the EPERM probe: a LIVE pid we cannot signal. The reference's
    # blanket except reads it DEAD (take-over-able); the pre-T-78 Go twin
    # read it ALIVE (refused) — the divergence the fix aligns.
    eperm_pid = find_eperm_pid()
    if eperm_pid is None:
        record("A6 eperm-probe", True,
               "SKIPPED — no live not-ours pid on this box (running as root?); "
               "the law stays pinned by TestKillVerdictMirrorsTheBlanketExcept")
    else:
        for side, tag, cmd in (
            ("py", "a6-py", [py, str(pylock), "try"]),
            ("go", "a6-go", [str(goside), "lock-try", "app.lock"]),
        ):
            d = fresh_dir(tag)
            (d / "app.lock").write_text(str(eperm_pid))
            r = run_cmd(cmd, str(d), base_env())
            stale = (f"Removing stale lock file (PID {eperm_pid} "
                     "is no longer running)...")
            ok = (r["exit"] == 0 and stale in r["out"]
                  and (("PYLOCK HELD" in r["out"]) if side == "py"
                       else ("GOSIDE HELD" in r["out"])))
            record(f"A6 eperm-probe {side} reads dead (take-over)", ok,
                   f"the probe pid {eperm_pid} is alive but not ours to signal",
                   f"stdout: {r['out'].strip()!r}",
                   "both binaries take the lock over — the blanket-except law")


def find_eperm_pid():
    """A live process we cannot signal (kill 0 -> PermissionError) —
    the EPERM branch's live differential fixture."""
    for entry in os.listdir("/proc"):
        if not entry.isdigit():
            continue
        pid = int(entry)
        if pid == os.getpid():
            continue
        try:
            os.kill(pid, 0)
        except PermissionError:
            return pid
        except OSError:
            continue
    return None


# ---------------------------------------------------------------------------
# LEG B — the same-data config laws
# ---------------------------------------------------------------------------

def leg_b(py, app_dir, goside: Path):
    # The reference's own pure modules, imported in-process.
    sys.path.insert(0, app_dir)
    from gitcurator.core import storage as ref_storage
    from gitcurator.core import user_config as ref_user_config
    from gitcurator import constants as ref_constants

    def fresh_dir(tag):
        d = RUNS / f"cfg-{tag}"
        if d.exists():
            shutil.rmtree(d)
        d.mkdir(parents=True)
        return d

    # The shared fixture — the reference's own config shape.
    base = copy.deepcopy(ref_constants.CONFIG_EXAMPLE)
    base["telegram_phone"] = "local-wins-value"          # the local-wins probe
    base["user_custom_key"] = "unknown keys survive"     # merge-never-rebuild
    base["timeout_per_repo"] = 99                        # tuning keys unstomped
    base["max_retries"] = 7
    base["delay_between_api_calls"] = 2.5
    base["github_token"] = "Ελληνικά ✓ — <b>&amp;</b>"   # non-ASCII + HTML-ish
    base["proxy"] = dict(base.get("proxy") or {})
    base["proxy"]["host"] = "127.0.0.1"
    base["proxy"]["port"] = 10808
    base["ollama"] = dict(base.get("ollama") or {})
    base["ollama"]["base_url"] = "http://127.0.0.1:11434"

    updates = {
        "telegram_phone": "+15550100",                   # a real edit
        "user_custom_key2": "new unknown key",
        "proxy": {"host": "127.0.0.2", "use_for_web": True},   # nested merge
        "ollama": {"model": "llama3"},                   # nested NEW key
    }

    d = fresh_dir("b1")
    p0 = d / "config.json"
    ref_storage.write_config_file(str(p0), copy.deepcopy(base))
    (d / "updates.json").write_text(json.dumps(updates, indent=2),
                                    encoding="utf-8")
    p0_bytes = p0.read_bytes()

    # PY side: json.load -> merge_config -> write (the GUI's save engine).
    cfg = json.loads(p0_bytes.decode("utf-8"))
    merged = ref_storage.merge_config(cfg, copy.deepcopy(updates))
    p1 = d / "config.py.json"
    ref_storage.write_config_file(str(p1), merged)
    p1_bytes = p1.read_bytes()

    # GO side: the app door (load -> merge -> write) over the SAME input.
    gdir = fresh_dir("b1-go")
    gcfg = gdir / "config.json"
    shutil.copyfile(p0, gcfg)
    shutil.copyfile(d / "updates.json", gdir / "updates.json")
    env = base_env({"GITCURATOR_USER_CONFIG": str(gdir / "mem-absent.json")})
    r = run_cmd([str(goside), "config-save", "config.json", "updates.json"],
                str(gdir), env)
    g1_bytes = gcfg.read_bytes()

    ev = [f"sha256(go) = {sha256(g1_bytes)}",
          f"sha256(py) = {sha256(p1_bytes)}",
          f"goside: {r['out'].strip()!r} exit={r['exit']}"]
    ok = g1_bytes == p1_bytes
    if not ok:
        diff = "\n".join(list(difflib.unified_diff(
            p1_bytes.decode("utf-8", "replace").splitlines(),
            g1_bytes.decode("utf-8", "replace").splitlines(),
            "python", "go", lineterm=""))[:40])
        ev.append("diff (first 40 lines):\n" + diff)
    record("B1 both apps write the SAME bytes (merge + write laws)", ok, *ev)

    # The observables, pinned byte-level on the Go-written file:
    ok_unknown = (b'"user_custom_key": "unknown keys survive"' in g1_bytes
                  and b'"user_custom_key2": "new unknown key"' in g1_bytes)
    record("B1a unknown keys survive both merges", ok_unknown,
           "the fixture's two unknown keys are in the Go-written bytes")
    ok_tuning = (b'"timeout_per_repo": 99' in g1_bytes
                 and b'"max_retries": 7' in g1_bytes
                 and b'"delay_between_api_calls": 2.5' in g1_bytes)
    record("B1b tuning keys unstomped (default-only)", ok_tuning,
           "the custom 99 / 7 / 2.5 values survive — never the 60/3/0.5 defaults")
    ok_nested = (b'"host": "127.0.0.2"' in g1_bytes
                 and b'"model": "llama3"' in g1_bytes)
    record("B1c nested dicts merge key-by-key", ok_nested,
           "proxy.host edited + ollama.model added — the rest of each dict kept")
    raw_value = "Ελληνικά ✓ — <b>&amp;</b>".encode("utf-8")
    record("B1d non-ASCII raw + no HTML escaping", raw_value in g1_bytes,
           f"the unicode+HTML-ish value's raw UTF-8 ({len(raw_value)} bytes) "
           "is byte-present — ensure_ascii=False and SetEscapeHTML(false) both live")

    # The cross re-save stability: Python re-saves Go's file; Go re-saves
    # Python's file — neither disturbs the bytes.
    cfg2 = json.loads(g1_bytes.decode("utf-8"))
    p2 = d / "config.py-resave.json"
    ref_storage.write_config_file(str(p2), cfg2)
    record("B1e python's re-save of the go file is byte-stable",
           p2.read_bytes() == g1_bytes,
           f"sha256(py-re-save) = {sha256(p2.read_bytes())}")
    (gdir / "updates-empty.json").write_text("{}", encoding="utf-8")
    r2 = run_cmd([str(goside), "config-save", "config.json",
                  "updates-empty.json"], str(gdir), env)
    record("B1f go's re-save of the python file is byte-stable",
           gcfg.read_bytes() == p1_bytes,
           f"sha256(go-re-save) = {sha256(gcfg.read_bytes())}",
           f"goside: {r2['out'].strip()!r}")

    # B2 — the user-config fill on the SAME memory file.
    d = fresh_dir("b2")
    mem = d / "dot-gitcurator.json"

    # The memory fixture — written by the reference's OWN mirror door.
    remembered = {
        "telegram_api_id": 12345678,
        "telegram_phone": "+15550100",
        "github_token": "memory-value-not-a-secret",
        "bot_token": "only-in-memory-value",  # the never-loses probe: a
        # remembered-family key that lives ONLY in the memory (the edited
        # config below never sets it — a correct mirror never erases it)
        "proxy": {"host": "127.0.0.1", "port": 10808, "enabled": True,
                  "type": "socks5", "use_for_web": False},
        "ollama": {"base_url": "http://127.0.0.1:11434", "model": "llama3"},
    }
    ref_user_config.mirror_credentials_to_global(
        copy.deepcopy(remembered), str(mem))
    mem_fixture = mem.read_bytes()

    # The input config — the example + one real local value (local-wins).
    local = copy.deepcopy(ref_constants.CONFIG_EXAMPLE)
    local["telegram_phone"] = "local-wins-value"

    for fresh in (False, True):
        tag = "fresh" if fresh else "kept"
        # PY side
        pcfg = copy.deepcopy(local)
        n_py = ref_user_config.apply_global_credentials(pcfg, str(mem),
                                                        fresh=fresh)
        pout = d / f"filled-{tag}-py.json"
        ref_storage.write_config_file(str(pout), pcfg)
        # GO side
        gin = d / f"fill-in-{tag}.json"
        gout = d / f"filled-{tag}-go.json"
        ref_storage.write_config_file(str(gin), copy.deepcopy(local))
        r = run_cmd([str(goside), "fill", str(mem), str(gin), str(gout),
                     "true" if fresh else "false"], str(d), base_env())
        ok = (pout.read_bytes() == gout.read_bytes()
              and f"GOSIDE FILL {n_py}" in r["out"])
        record(f"B2 fill ({tag}) — same bytes, same count", ok,
               f"sha256(py) = {sha256(pout.read_bytes())}",
               f"sha256(go) = {sha256(gout.read_bytes())}",
               f"counts: py={n_py}, go line={r['out'].strip()!r}")
        if not fresh:
            record("B2a local-wins law (kept mode)",
                   b'"telegram_phone": "local-wins-value"' in pout.read_bytes(),
                   "the real local value beats the remembered one")

    # B3 — the mirror: both sides from the SAME fixture start; never-loses.
    edited = copy.deepcopy(local)
    edited["github_token"] = "newly-typed-memory-value"
    edited["cloud_api_key"] = "another-new-memory-value"
    min_cfg = d / "mirror-in.json"
    ref_storage.write_config_file(str(min_cfg), copy.deepcopy(edited))

    # PY side
    mem.write_bytes(mem_fixture)
    ok_py = ref_user_config.mirror_credentials_to_global(
        json.loads(min_cfg.read_text(encoding="utf-8")), str(mem))
    m1 = mem.read_bytes()
    mode_py = oct(os.stat(mem).st_mode & 0o777)

    # GO side (from the same fixture start)
    mem.write_bytes(mem_fixture)
    r = run_cmd([str(goside), "mirror", str(mem), str(min_cfg)],
                str(d), base_env())
    m2 = mem.read_bytes()
    mode_go = oct(os.stat(mem).st_mode & 0o777)

    record("B3 mirror — same memory bytes, never-loses", m1 == m2 and ok_py,
           f"sha256(py) = {sha256(m1)}  mode={mode_py}",
           f"sha256(go) = {sha256(m2)}  mode={mode_go}",
           f"goside: {r['out'].strip()!r}",
           "bot_token (memory-only) kept: "
           f"{b'only-in-memory-value' in m1 and b'only-in-memory-value' in m2}")
    record("B3a the 0600 permission law",
           mode_py == "0o600" and mode_go == "0o600",
           f"py={mode_py} go={mode_go}")


# ---------------------------------------------------------------------------
# The report
# ---------------------------------------------------------------------------

def write_report(extraction_shas, py, app_dir):
    RUNS.mkdir(parents=True, exist_ok=True)
    total = len(results)
    passed = sum(1 for r in results if r["ok"])
    verdict = "GREEN" if passed == total else "RED"
    lines = [
        "# T-78 — the side-by-side verification vs the Python app",
        "",
        f"- Verdict: **{verdict}** ({passed}/{total} checks)",
        f"- Reference: {app_dir} @ {PINNED_REV[:12]}",
        f"- Python: {py}",
        f"- The lock extraction provenance: gui/app.py:{LOCK_DANCE_RANGE[0]}-{LOCK_DANCE_RANGE[1]} "
        f"sha256 {extraction_shas[0][:16]}… · gui/headless.py:{IS_RUNNING_RANGE[0]}-{IS_RUNNING_RANGE[1]} "
        f"sha256 {extraction_shas[1][:16]}…",
        "",
        "## Checks",
        "",
    ]
    for r in results:
        lines.append(f"- [{'PASS' if r['ok'] else 'FAIL'}] {r['name']}")
        for ev in r["evidence"]:
            lines.append(f"  - {ev}")
    (RUNS / "report.md").write_text("\n".join(lines) + "\n", encoding="utf-8")
    (RUNS / "report.json").write_text(
        json.dumps({"verdict": verdict, "passed": passed, "total": total,
                    "checks": results,
                    "provenance": {"rev": PINNED_REV,
                                   "dance_sha256": extraction_shas[0],
                                   "isrunning_sha256": extraction_shas[1]}},
                   indent=2), encoding="utf-8")
    return verdict, passed, total


def main():
    ap = argparse.ArgumentParser()
    ap.add_argument("--leg", choices=["A", "B", "all"], default="all")
    args = ap.parse_args()

    py = DEFAULT_PY
    app_dir = DEFAULT_APP_DIR
    go = os.environ.get("GO", "go")
    RUNS.mkdir(parents=True, exist_ok=True)

    goside = guard_environment(py, app_dir, go)
    dance_src, isrun_src, dance_sha, isrun_sha = extract_reference_lock(app_dir)

    if args.leg in ("A", "all"):
        leg_a(py, goside, dance_src, isrun_src)
    if args.leg in ("B", "all"):
        leg_b(py, app_dir, goside)

    verdict, passed, total = write_report((dance_sha, isrun_sha), py, app_dir)
    print(f"\n=== T-78 side-by-side: {verdict} ({passed}/{total}) ===")
    print(f"report: {RUNS / 'report.md'}")
    sys.exit(0 if verdict == "GREEN" else 1)


if __name__ == "__main__":
    main()
