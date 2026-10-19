#!/bin/sh
# run-gate.sh — the T-70 gate's headless evidence leg: builds nothing,
# runs the already-built shell under Xvfb with the rootless webkit tree
# (ADR-0002 Am.3's CGO story; SESSION-001's recipe, committed), in autorun
# mode: first frame → the real baseline job → the 10k flood → the results
# JSON. Usage: run-gate.sh [seconds-to-wait-after-autorun]
#
# Environment (all local to this script):
#   SB_POC_RESULTS   results JSON path (default /tmp/t70-gate-results.json)
#   SB_SCREENSHOT    screenshot path (default /tmp/t70-gate-screenshot.png)
set -eu

TREE=${HOME}/.local/lib-wails
HERE=$(cd "$(dirname "$0")" && pwd)
BIN=$(cd "$HERE/../build/bin" && pwd)/gitcurator-shell
RESULTS=${SB_POC_RESULTS:-/tmp/t70-gate-results.json}
SHOT=${SB_SCREENSHOT:-/tmp/t70-gate-screenshot.png}
WAIT=${1:-6}

# the shim (rebuild if missing — the recipe is self-contained)
if [ ! -f "$HERE/webkit-path-shim.so" ]; then
    gcc -shared -fPIC -o "$HERE/webkit-path-shim.so" "$HERE/webkit-path-shim.c" -ldl
fi

mkdir -p /tmp/t70-xdg-runtime && chmod 700 /tmp/t70-xdg-runtime
rm -f "$RESULTS"

Xvfb :99 -screen 0 1100x750x24 -nolisten tcp >/tmp/t70-xvfb.log 2>&1 &
XVFB_PID=$!
trap 'kill $XVFB_PID 2>/dev/null || true' EXIT
sleep 1

export DISPLAY=:99
export GDK_BACKEND=x11
export WEBKIT_DISABLE_COMPOSITING_MODE=1
export LD_PRELOAD=$HERE/webkit-path-shim.so
export LD_LIBRARY_PATH=$TREE/usr/lib/x86_64-linux-gnu
export WEBKIT_TREE_LIBEXEC=$TREE/usr/lib/x86_64-linux-gnu/webkit2gtk-4.1
export XDG_DATA_DIRS=$TREE/usr/share:/usr/share
export XDG_RUNTIME_DIR=/tmp/t70-xdg-runtime
export SB_POC_AUTORUN=1
export SB_POC_RESULTS=$RESULTS

"$BIN" >/tmp/t70-gate-stdout.log 2>&1 &
APP_PID=$!

# wait for the autorun verdict to land (bounded)
i=0
while [ ! -s "$RESULTS" ] && [ $i -lt 60 ]; do
    sleep 1
    i=$((i + 1))
done
sleep "$WAIT"

# the evidence screenshot (ffmpeg x11grab — the whole themed window)
ffmpeg -y -loglevel error -f x11grab -video_size 1100x750 -i :99 \
    -frames:v 1 "$SHOT" || true

kill $APP_PID 2>/dev/null || true
wait $APP_PID 2>/dev/null || true
echo "results: $RESULTS"
cat "$RESULTS" 2>/dev/null || echo "NO RESULTS FILE — check /tmp/t70-gate-stdout.log"
