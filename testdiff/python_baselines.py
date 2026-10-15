#!/usr/bin/env python3
"""python_baselines.py — the T-04 Phase-0 performance baseline driver.

Implements the doc 07 §B.4 procedure for the Python reference app
(assadigit/GitCurator @ pinned rev): timed cases for startup, scan
inventory (with the one-open cache), log flood (10k lines), queue
classification (1,000 fake links), CLI --status wall time, idle-GUI
memory — measured numbers only, never estimated.

Everything runs against the UNMODIFIED reference repo (a read-only clone);
scratch state lives outside it. GUI cases use QT_QPA_PLATFORM=offscreen
(the sandbox's PyQt6 needed a locally extracted libEGL — recorded in the
artifact). Each mode writes a JSON result file; "all" orchestrates the
modes as subprocesses and prints one combined JSON behind the
###BASELINE_JSON### sentinel (timing_runner style).

Usage:
    python3 testdiff/python_baselines.py gui|scan|queue|cli|golden|all
"""

import json
import os
import shutil
import subprocess
import sys
import tempfile
import time
from pathlib import Path

HERE = Path(__file__).resolve().parent
APP_DIR = Path(os.environ.get("GC_APP_DIR", "/home/z/GitCurator/app"))
N_NOTES = 1309   # the owner's real vault size per doc 07 §B.4
N_LINKS = 1000   # queue classification corpus size per doc 07 §B.4


def out_json(mode: str, payload: dict, path: str):
    payload = dict(payload)
    payload["mode"] = mode
    payload["python"] = sys.version.split()[0]
    with open(path, "w", encoding="utf-8") as f:
        json.dump(payload, f, indent=2)
    print("###BASELINE_JSON###" + json.dumps(payload))


# ---------------------------------------------------------------------------
# mode: gui — cold start → first frame, log flood (10k), idle memory
# ---------------------------------------------------------------------------

def mode_gui(json_path: str):
    # The driver process is spawned fresh: t0 here = post-interpreter-boot.
    t0 = time.perf_counter()
    os.environ.setdefault("QT_QPA_PLATFORM", "offscreen")
    sys.path.insert(0, str(APP_DIR))

    import PyQt6.QtWidgets as QtW  # noqa: E402  (timed import)
    t_qt_import = time.perf_counter()

    app = QtW.QApplication([])

    import gitcurator.gui.app as ga  # noqa: E402  (timed import — the heavy one)
    from gitcurator.gui.main_window import vaults_config as _vc  # noqa: E402
    from gitcurator.constants import APP_DIR as _APP_DIR  # noqa: E402
    t_ga_import = time.perf_counter()

    # the test_logbreathe pattern: temp CONFIG_FILE so nothing real is touched
    tmp = tempfile.mkdtemp(prefix="gc-t04-")
    _vc.CONFIG_FILE = os.path.join(tmp, "config.json")

    t_construct0 = time.perf_counter()
    win = ga.MainWindow()
    t_construct = time.perf_counter()

    win.show()
    app.processEvents()  # first paint processed (offscreen)
    t_first_frame = time.perf_counter()

    # idle memory after first frame
    def _rss_mb():
        with open("/proc/self/status", encoding="utf-8") as f:
            txt = f.read()
        rss = int(re_search(r"VmRSS:\s+(\d+) kB", txt))
        hwm = int(re_search(r"VmHWM:\s+(\d+) kB", txt))
        return round(rss / 1024.0, 1), round(hwm / 1024.0, 1)

    import re as _re

    def re_search(pat, txt):
        return _re.search(pat, txt).group(1)

    rss_after_frame, hwm_after_frame = _rss_mb()

    # 3 s of idle event loop (sentinel/cadence timers tick, nothing else)
    idle_until = time.perf_counter() + 3.0
    while time.perf_counter() < idle_until:
        app.processEvents()
        time.sleep(0.02)
    rss_idle, _ = _rss_mb()

    # --- the log flood: 10k real log_message calls through the real window
    t_flood0 = time.perf_counter()
    for i in range(10000):
        win.log_message("🚚 Some Note With A Rather Long Name %d.md: "
                        "Vault root → Category/Subfolder %d" % (i, i % 30), "info")
    t_producer = time.perf_counter()

    # drain (the test_logbreathe _drain pattern)
    t_drain0 = time.perf_counter()
    while True:
        app.processEvents()
        if not (getattr(win, "_log_pending", None) or []):
            timer = getattr(win, "_log_flush_timer", None)
            if not timer or not timer.isActive():
                break
        if time.perf_counter() - t_drain0 > 60:
            break
        time.sleep(0.01)
    t_drained = time.perf_counter()

    rss_after_flood, hwm_after_flood = _rss_mb()

    try:
        win._closing = True
        win.deleteLater()
    except Exception:
        pass
    app.processEvents()

    # keep the reference repo pristine (the repo's own tests do the same)
    for poison in ("config.json", "cache.db", "app.lock"):
        p = os.path.join(_APP_DIR, poison)
        if os.path.exists(p):
            os.unlink(p)
    shutil.rmtree(tmp, ignore_errors=True)

    producer_ms = (t_producer - t_flood0) * 1000.0
    out_json("gui", {
        "qt_import_ms": round((t_qt_import - t0) * 1000.0, 1),
        "gui_app_import_ms": round((t_ga_import - t_qt_import) * 1000.0, 1),
        "mainwindow_construct_ms": round((t_construct - t_construct0) * 1000.0, 1),
        "show_to_first_frame_ms": round((t_first_frame - t_construct) * 1000.0, 1),
        "driver_t0_to_first_frame_ms": round((t_first_frame - t0) * 1000.0, 1),
        "log_flood_producer_ms": round(producer_ms, 1),
        "log_flood_producer_us_per_line": round(producer_ms * 1000.0 / 10000, 2),
        "log_flood_drain_ms": round((t_drained - t_producer) * 1000.0, 1),
        "log_flood_total_ms": round((t_drained - t_flood0) * 1000.0, 1),
        "rss_mb_after_first_frame": rss_after_frame,
        "rss_mb_idle_3s": rss_idle,
        "rss_mb_after_10k_flood": rss_after_flood,
        "vmhwm_mb": hwm_after_flood,
    }, json_path)


# ---------------------------------------------------------------------------
# mode: scan — synthetic 1,309-note vault inventory + one-open cache
# ---------------------------------------------------------------------------

def _build_vault(root: Path, n_notes: int) -> int:
    categories = [
        ("Tech", ["AI", "DevOps", "Security"]),
        ("Science", ["Physics", "Biology"]),
        ("Design", ["Typography", "UX"]),
        ("Business", ["Strategy", "Marketing"]),
        ("Culture", ["Film", "Books"]),
        ("Health", ["Nutrition", "Sleep"]),
    ]
    note_tmpl = (
        "---\n"
        "title: {title}\n"
        "url: https://example-{dom}.com/articles/{slug}\n"
        "domain: example-{dom}.com\n"
        "category: {cat}\n"
        "subcategory: {sub}\n"
        "date_processed: 2025-0{m}-1{d}\n"
        "stars: {stars}\n"
        "kind: website\n"
        "---\n\n"
        "# {title}\n\nA synthetic baseline note with a realistic body length.\n"
    )
    made = 0
    per_cat = n_notes // len(categories)
    extra = n_notes - per_cat * len(categories)
    for ci, (cat, subs) in enumerate(categories):
        count = per_cat + (extra if ci == 0 else 0)
        for i in range(count):
            sub = subs[i % len(subs)]
            folder = root / cat / sub
            folder.mkdir(parents=True, exist_ok=True)
            title = "Note %s %s %04d" % (cat, sub, i)
            body = note_tmpl.format(title=title, dom=(ci % 5), slug="n%04d" % i,
                                    cat=cat, sub=sub, m=(i % 9) + 1, d=i % 10, stars=i % 5)
            (folder / ("%s.md" % title.replace(" ", "-"))).write_text(body, encoding="utf-8")
            made += 1
    for i in range(5):  # a few root notes
        (root / ("Root Note %d.md" % i)).write_text(
            note_tmpl.format(title="Root Note %d" % i, dom=9, slug="r%d" % i,
                             cat="", sub="", m=1, d=1, stars=0), encoding="utf-8")
        made += 1
    return made


def mode_scan(json_path: str):
    os.environ.setdefault("PYTHONDONTWRITEBYTECODE", "1")
    sys.path.insert(0, str(APP_DIR))
    from gitcurator.core.vault_scan import scan_vault_inventory
    from gitcurator.core.note_reads import NoteReadCache

    tmp = Path(tempfile.mkdtemp(prefix="gc-t04-scan-"))
    vault = tmp / "vault"
    made = _build_vault(vault, N_NOTES)

    def timed(label, **kw):
        t0 = time.perf_counter()
        inv = scan_vault_inventory(str(vault), **kw)
        dt = (time.perf_counter() - t0) * 1000.0
        return label, round(dt, 1), inv

    # pass 1: no cache (the old direct 16 KB head reads) — COLD, fresh process
    _, ms_cold, inv1 = timed("cold-nocache")
    # pass 2: no cache again — the re-read cost (no cache in-process either)
    _, ms_reread, inv2 = timed("reread-nocache")
    # pass 3+4: the ONE-OPEN LAW — NoteReadCache populated, then warm
    cache = NoteReadCache(db_path=str(tmp / "reads.db"))
    _, ms_cache_fill, inv3 = timed("cache-fill", read_cache=cache)
    _, ms_cache_warm, inv4 = timed("cache-warm", read_cache=cache)

    notes_seen = len(inv1.get("notes", [])) + inv1.get("root_notes", 0)
    shutil.rmtree(tmp, ignore_errors=True)

    out_json("scan", {
        "synthetic_notes_created": made,
        "notes_seen_by_scan": notes_seen,
        "scan_cold_nocache_ms": ms_cold,
        "scan_reread_nocache_ms": ms_reread,
        "scan_cache_fill_ms": ms_cache_fill,
        "scan_cache_warm_ms": ms_cache_warm,
        "cache_warm_vs_cold_ratio": round(ms_cache_warm / ms_cold, 3) if ms_cold else None,
        "notes_per_second_cold": round(notes_seen / (ms_cold / 1000.0), 1) if ms_cold else None,
    }, json_path)


# ---------------------------------------------------------------------------
# mode: queue — 1,000 fake links through the worker's classification loop
# ---------------------------------------------------------------------------

def mode_queue(json_path: str):
    sys.path.insert(0, str(APP_DIR))
    from gitcurator.core import links as _links

    # the corpus: 1,000 links across ~820 messages (some dupes, mixed kinds)
    messages = []
    for i in range(N_LINKS):
        if i % 7 == 3:
            messages.append("check this out https://www.pexels.com/photo/%d/ nice" % i)
        elif i % 7 == 5:
            messages.append("read https://news.ycombinator.com/item?id=%d today" % i)
        else:
            owner = ["torvalds", "golang", "assadigit", "python", "rust-lang"][i % 5]
            messages.append("repo https://github.com/%s/repo-%d star it" % (owner, i))
    # 15% duplicates (raw link count > unique count — the dedup path)
    messages += messages[: len(messages) // 7]

    github_pattern = _links.GITHUB_URL_PATTERN
    all_links_pattern = _links.ALL_LINKS_PATTERN

    # the telegram_fetch_worker.py:503-566 loop, verbatim shape
    def classify(messages):
        urls = []
        non_github_urls = []
        _seen_github = set()
        _seen_links = set()
        raw_link_count = 0
        for msg_text in messages:
            if msg_text:
                for link in all_links_pattern.findall(msg_text):
                    raw_link_count += 1
                    link = link.rstrip('.,);')
                    if link not in _seen_links and link not in _seen_github:
                        _seen_links.add(link)
                        non_github_urls.append(link)
                for owner, repo in github_pattern.findall(msg_text):
                    url = "https://github.com/%s/%s" % (owner, repo)
                    if url not in _seen_github:
                        _seen_github.add(url)
                        urls.append(url)
        filtered_non_github = []
        for u in non_github_urls:
            is_github = u in _seen_github
            if not is_github and "github.com" in u:
                match = github_pattern.search(u)
                if match:
                    constructed = "https://github.com/%s/%s" % (match.group(1), match.group(2))
                    if constructed not in _seen_github:
                        _seen_github.add(constructed)
                        urls.append(constructed)
                    is_github = True
            if not is_github:
                filtered_non_github.append(u)
        return urls, filtered_non_github, raw_link_count

    # warm import caches first (imports are not the measured case)
    classify(messages[:5])
    t0 = time.perf_counter()
    runs = []
    for _ in range(5):
        t1 = time.perf_counter()
        urls, non_gh, raw = classify(messages)
        runs.append((time.perf_counter() - t1) * 1000.0)
    total_ms = (time.perf_counter() - t0) * 1000.0

    out_json("queue", {
        "messages": len(messages),
        "raw_links_classified": raw,
        "github_urls": len(urls),
        "non_github_urls": len(non_gh),
        "runs_ms": [round(r, 2) for r in runs],
        "median_ms": round(sorted(runs)[len(runs) // 2], 2),
        "links_per_second": round(raw / (sorted(runs)[len(runs) // 2] / 1000.0), 1),
    }, json_path)


# ---------------------------------------------------------------------------
# mode: cli — --status wall time (missing config + empty config)
# ---------------------------------------------------------------------------

def mode_cli(json_path: str):
    tmp = Path(tempfile.mkdtemp(prefix="gc-t04-cli-"))
    env = dict(os.environ)
    env["PYTHONIOENCODING"] = "utf-8"
    env["NO_COLOR"] = "1"
    env["PYTHONDONTWRITEBYTECODE"] = "1"
    sub_py = os.environ.get("GC_PYTHON", sys.executable)

    def timeit(args, cwd):
        runs = []
        for _ in range(3):
            t0 = time.perf_counter()
            subprocess.run([sub_py, str(APP_DIR / "main.py")] + args,
                           cwd=cwd, env=env, capture_output=True)
            runs.append((time.perf_counter() - t0) * 1000.0)
        return round(sorted(runs)[1], 1)  # median of 3

    ms_missing = timeit(["--cli", "--status", "--config", str(tmp / "missing.json")], str(tmp))
    (tmp / "config.json").write_text("{}", encoding="utf-8")
    ms_empty = timeit(["--cli", "--status", "--config", str(tmp / "config.json")], str(tmp))
    # interpreter boot reference (the floor of any CLI wall time)
    t0 = time.perf_counter()
    subprocess.run([sub_py, "-c", "pass"], capture_output=True)
    ms_boot = round((time.perf_counter() - t0) * 1000.0, 1)

    shutil.rmtree(tmp, ignore_errors=True)
    out_json("cli", {
        "cli_status_missing_config_ms": ms_missing,
        "cli_status_empty_config_ms": ms_empty,
        "python_interpreter_boot_ms": ms_boot,
    }, json_path)


# ---------------------------------------------------------------------------
# mode: golden — the offline golden websites batch (fake fetch+LLM)
# ---------------------------------------------------------------------------

def mode_golden(json_path: str):
    tmp = Path(tempfile.mkdtemp(prefix="gc-t04-golden-"))
    env = dict(os.environ)
    env["PYTHONDONTWRITEBYTECODE"] = "1"
    env["PYTHONIOENCODING"] = "utf-8"
    env["NO_COLOR"] = "1"
    cmd = [sys.executable, str(APP_DIR / "gitcurator" / "tools" / "run_golden_websites.py"),
           "--offline", "--out", str(tmp / "out")]
    t0 = time.perf_counter()
    proc = subprocess.run(cmd, cwd=str(APP_DIR), env=env, capture_output=True, timeout=900)
    dt = round((time.perf_counter() - t0) * 1000.0, 1)
    ok = proc.returncode == 0
    tail = proc.stderr.decode("utf-8", errors="replace").strip().splitlines()[-5:]
    shutil.rmtree(tmp, ignore_errors=True)
    out_json("golden", {
        "cmd": "run_golden_websites.py --offline",
        "ok": ok,
        "wall_ms": dt,
        "stderr_tail": tail,
    }, json_path)


# ---------------------------------------------------------------------------
# mode: all — orchestrates subprocesses (fresh processes = honest cold runs)
# ---------------------------------------------------------------------------

def mode_all(json_path: str):
    env = dict(os.environ)
    env["PYTHONDONTWRITEBYTECODE"] = "1"
    env["QT_QPA_PLATFORM"] = "offscreen"
    # the sandbox's rootless libEGL for PyQt6 (recorded in the artifact)
    egl = os.path.expanduser("~/.local/lib-opengl/usr/lib/x86_64-linux-gnu")
    if os.path.isdir(egl):
        env["LD_LIBRARY_PATH"] = egl + (":" + env["LD_LIBRARY_PATH"] if env.get("LD_LIBRARY_PATH") else "")
    env.setdefault("GC_APP_DIR", str(APP_DIR))
    sub_py = os.environ.get("GC_PYTHON", sys.executable)

    results = {}
    for mode in ("cli", "queue", "scan", "gui", "golden"):
        sub_json = json_path.replace(".json", "-" + mode + ".json")
        cmd = [sub_py, str(Path(__file__).resolve()), mode]
        print("[baselines] running mode=%s" % mode, file=sys.stderr, flush=True)
        proc = subprocess.run(cmd, env=env, capture_output=True, timeout=900)
        if proc.returncode != 0:
            results[mode] = {"ok": False,
                             "stderr_tail": proc.stderr.decode("utf-8", errors="replace").strip().splitlines()[-8:]}
            continue
        # each mode prints one sentinel line
        line = [l for l in proc.stdout.decode("utf-8", errors="replace").splitlines()
                if l.startswith("###BASELINE_JSON###")]
        results[mode] = json.loads(line[-1][len("###BASELINE_JSON###"):]) if line else {"ok": False}
        results[mode]["ok"] = True
    with open(json_path, "w", encoding="utf-8") as f:
        json.dump({"baselines": results}, f, indent=2)
    print("###BASELINE_JSON###" + json.dumps({"baselines": results}))


def main() -> int:
    mode = sys.argv[1] if len(sys.argv) > 1 else "all"
    json_path = (sys.argv[sys.argv.index("--out") + 1] if "--out" in sys.argv
                 else "/tmp/t04-%s.json" % mode)
    if mode == "gui":
        mode_gui(json_path)
    elif mode == "scan":
        mode_scan(json_path)
    elif mode == "queue":
        mode_queue(json_path)
    elif mode == "cli":
        mode_cli(json_path)
    elif mode == "golden":
        mode_golden(json_path)
    elif mode == "all":
        mode_all(json_path)
    else:
        print("unknown mode %r" % mode, file=sys.stderr)
        return 2
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
