#!/usr/bin/env bash
# ============================================================
#  gitcurator-cli.sh — Unix launcher for the GitCurator visual CLI
#
#  ./gitcurator-cli.sh            → fully automatic run (--auto --yes)
#  ./gitcurator-cli.sh --sync     → the fast sync lane (check → confirm → sync)
#  ./gitcurator-cli.sh --init     → first-run wizard
#  ./gitcurator-cli.sh --status   → config + cache + strike summary
#  ./gitcurator-cli.sh ...        → any flag is forwarded to the CLI
#
#  v0.92.0-go.1: the Go port — the app is the native GitCurator-cli
#  binary sitting next to this file; no python3, no venv, nothing to
#  find. In the repo this file sits at cmd/gitcurator-cli/, two
#  levels under the root; in the zip it sits at the top, next to the
#  binary. VERSION is the app root's landmark (the reference's
#  landmark was main.py — the Python app it launched).
# ============================================================
set -euo pipefail
cd "$(dirname "$0")"
[ -f VERSION ] || cd ../..

if [ ! -x "./GitCurator-cli" ]; then
    echo "[GitCurator] GitCurator-cli not found next to this file — run me from the unzipped folder." >&2
    exit 1
fi

if [ "$#" -eq 0 ]; then
    exec ./GitCurator-cli --cli --auto --yes
fi

exec ./GitCurator-cli --cli "$@"
