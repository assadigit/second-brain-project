# T-70 — Wails PoC (ADR-0001) — evidence

Recorded 2026-10-15, SESSION-001. Environment: Debian 13 sandbox (no root),
go1.27.2, Wails **v2.16.0** (stable v2 line, pinned per ADR-0001's "pin a
stable Wails version" requirement), webkit2gtk-4.1 2.54.0 via a **rootless
extracted package tree** (~/.local/lib-wails — apt --download-only + dpkg -x;
no sudo in the sandbox), GTK 3.24.49. Screenshot:
`t-70-wails-poc-screenshot.png` (Xvfb 1100×750, ffmpeg x11grab).

## What the PoC contains (poc/wails)

- **Themed window**: GitCurator LIGHT tokens replicated from `gui/theme.py`
  (window `#FBF8F2` cream, text `#423A52` soft plum, accent `#5F54B4`,
  log tint `#F2EDE3`, level colors `#1E6B4B/#75510A/#AE2237`), 980×660,
  cream `BackgroundColour`, one filled primary per view (the btn_kind law).
- **Log desk**: Go-side coalescing bus — producer channel cap 1024,
  120 ms ticker flush, ≤250-row insert batches (the ui.py:65-86 numbers);
  backpressure instead of silent drops (the Python desk never lost lines);
  frontend batches DOM appends, caps DOM rows at 5000.
- **One worker job**: sequential 50-item loop, 20 ms/item, progress events
  to the progress bar, desk-log lines (the batch-sequential law, doc 04 §6).
- **Measurements**: process start → first frame (DOMContentLoaded round-trip),
  job wall time, 10k-line flood accept rate + drops. Autorun mode
  (`SB_POC_AUTORUN=1`) drives everything headlessly and writes
  `/tmp/t70-results.json`.

## Measured results (Xvfb run, 2026-10-15 04:50)

```json
{"last_flood":{"accept_lines_per_s":625000,"accept_ms":16,"dropped":0,
 "ingested":10011,"lines":10000,"ok":true},
 "last_job":{"items":50,"ms":1011,"ms_per_item":20.22,"ok":true},
 "startup_callback_ms":65,
 "startup_to_first_frame_ms":213}
```

| ADR-0001 PoC acceptance | Gate | Measured | Verdict |
|---|---|---|---|
| Cold start → first frame | < 1.5 s | **213 ms** (first run 301 ms) | **PASS** |
| Log desk sustains appends | ≥ 2,000 rows/s, interactive | **625,000 lines/s accepted, 0 drops**; desk rendered 5,000-row trim cap, scrolled, interactive | **PASS** |
| exe size | < 20 MB | Linux 9.1 MB / **Windows .exe 11.8 MB** (cross-compiled) | **PASS** |

Visual verification (VLM-read of the screenshot): cream/plum theme confirmed;
buttons + top bar + "startup→frame: 217.7ms" badge + progress bar "50/50"
rendered; desk filled with monospace rows ("flood line 9979 … the desk must
stay interactive", green success lines "flood: 10000 lines accepted in 17ms
(drops so far: 0)" and "autorun complete").

## Build & run recipe (rootless Linux, recorded for reproducibility)

```sh
# one-time: rootless webkit2gtk-4.1 + gtk3 dev closure
apt-get install --download-only -o Dir::Cache::archives=/tmp/wails-debs \
    -o Debug::NoLocking=1 -y libwebkit2gtk-4.1-dev libgtk-3-dev
for d in /tmp/wails-debs/*.deb; do dpkg -x "$d" ~/.local/lib-wails/; done
# rewrite .pc prefixes to the extracted tree (sed, see session log)
# fix broken dev symlinks -> system runtime copies (webkit libs were skipped
# by apt as "already installed")
wails build -tags webkit2_41   # trixie ships only webkit2gtk-4.1

# headless run (Xvfb + a 30-line LD_PRELOAD shim that rewrites WebKit's
# compile-time libexec path /usr/lib/x86_64-linux-gnu/webkit2gtk-4.1 → tree;
# WEBKIT_EXEC_PATH alone is not honored by WebKit 2.54)
Xvfb :99 -screen 0 1100x750x24 &
LD_PRELOAD=/tmp/webkit-path-shim.so \
LD_LIBRARY_PATH=$HOME/.local/lib-wails/usr/lib/x86_64-linux-gnu \
SB_POC_AUTORUN=1 DISPLAY=:99 WEBKIT_DISABLE_COMPOSITING_MODE=1 \
XDG_DATA_DIRS=$HOME/.local/lib-wails/usr/share:/usr/share \
./build/bin/wails-poc-t70
```

## The blocked leg (recorded as blocked-with-plan, per the task law)

**The silent WebView2 Evergreen-bootstrapper path on a Windows machine/VM
WITHOUT WebView2 could not be executed here — the sandbox has no Windows
environment.** Evidence of the constraint: only linux/amd64 binaries can
run in this sandbox (the cross-compiled `wails-poc-t70.exe` was built but
not executed).

Plan for the owner (or any Windows 10 VM without WebView2 — Edge deleted,
EOL, no updates — exactly ADR-0001 Am.1's scenario):

1. Produce the installer: `wails build -platform windows/amd64 -nsis`
   (Wails' NSIS target chains Microsoft's official **WebView2 Evergreen
   bootstrapper** with the silent flags `/silent /install`; NSIS is
   available via apt in this sandbox if installer *generation* proof is
   wanted in a later session).
2. On the VM, verify absence first: `winget list
   Microsoft.EdgeWebView2Runtime` (or registry probe
   `HKLM\SOFTWARE\WOW6432Node\Microsoft\EdgeUpdate\Clients\
   {F3017226-FE2A-4295-8BDF-00C3A3960767}`).
3. Run the setup.exe → measure: silent WebView2 install time → first
   launch → themed window on screen; record numbers.
4. Portable-zip variant: first-run check + one-click silent fetch of the
   bootstrapper (ADR-0010 Am.1's dual-artifact law).

This leg is registered as the owner action in NEXT-SESSION.md.
