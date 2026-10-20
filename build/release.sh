#!/usr/bin/env bash
# build/release.sh — THE DETERMINISTIC RELEASE RITUAL (T-80; doc 10 §2/§5,
# AC-22): build BOTH exes + the release zip TWICE each (plus one COLD-cache
# leg per exe — a fresh machine must reproduce the tag) and prove
# byte-identity by sha256. The release automation (T-81) wraps this script;
# the reproduce-from-tag recipe is exactly these steps from a clean checkout
# of the tag.
#
# Usage (from anywhere; the script finds the repo root):
#     PATH=/path/to/go/bin:$PATH ./build/release.sh
#     PATH=… ./build/release.sh --allow-dirty   # dev probes only
#
# Exit codes: 0 = the ritual passed (every pair byte-identical);
#             2 = a guard refused (dirty tree, version drift, build failure)
#             3 = byte-identity FAILED (the release is not reproducible).
#
# The ritual's laws:
#   - the CLI exe: GOOS=windows GOARCH=amd64 CGO_ENABLED=0
#     go build -trimpath -ldflags "-X main.version=$(cat VERSION)" over
#     ./cmd/gitcurator-cli (doc 10 §2's CLI command);
#   - the GUI exe: wails build -platform windows/amd64 -trimpath
#     -ldflags "-X main.version=$(cat VERSION)" over the NESTED webui
#     module (the windowsgui subsystem, the brand icon + the T-80 manifest
#     + the version resource ride the .syso; CGO-free — the WebView2
#     loader is pure Go);
#   - the zip: go run ./build/mkzip -version "$(cat VERSION)" -exe …
#     -gui-exe … (the two-exe composition + the determinism laws + the
#     version-stamp + brand-icon guards);
#   - THE ONE-VERSION-TRUTH: webui/wails.json's Info.productVersion must
#     equal the repo VERSION file (the exe's version resource and the
#     zip's stamped VERSION agree);
#   - THE CLEAN-TREE LAW: a release builds from a committed tree (the
#     exes embed the working tree's source; a dirty tree cannot be
#     reproduced from its tag). --allow-dirty overrides for dev probes.
#   - THE TAG-STATE BUILD LAW (T-81's find, 2026-10-20): go1.24+ embeds
#     the VCS-derived module version in buildinfo — a TAGGED checkout
#     (the release state) and an untagged one of the SAME commit produce
#     DIFFERENT exe bytes (the buildinfo mod line carries the tag name vs
#     a pseudo-version; byte-settled at T-81: same commit, sha moved with
#     the tag). The identity claims are per-tree-state; the release
#     numbers must be derived at the tagged checkout (the
#     reproduce-from-tag recipe builds exactly there). The wails-built
#     GUI exe is exempt (its buildinfo rides "(devel)" + a v0.0.0 parent
#     dep — byte-stable across commits and tag states).
#
# Outputs land under reports/dist/ (gitignored — binaries never enter git;
# the sha256 evidence is archived in docs/poc-artifacts/).

set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
ALLOW_DIRTY=0
for arg in "$@"; do
    case "$arg" in
        --allow-dirty) ALLOW_DIRTY=1 ;;
        *) echo "[release] fatal: unknown argument: $arg" >&2; exit 2 ;;
    esac
done

GO_BIN="${GO:-go}"
WAILS_BIN="${WAILS:-wails}"
VERSION="$(cat "$ROOT/VERSION")"
DIST="$ROOT/reports/dist"
mkdir -p "$DIST"

say()  { printf '[release] %s\n' "$*"; }
die2() { printf '[release] fatal: %s\n' "$*" >&2; exit 2; }
die3() { printf '[release] REPRODUCIBILITY FAILURE: %s\n' "$*" >&2; exit 3; }
sha()  { sha256sum "$1" | cut -d' ' -f1; }

# --- Preflight guards ------------------------------------------------------

command -v "$GO_BIN"   >/dev/null 2>&1 || die2 "go not on PATH (export PATH=…/go/bin:\$PATH)"
command -v "$WAILS_BIN" >/dev/null 2>&1 || die2 "wails CLI not on PATH (GOBIN=… go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0)"

if [ "$(git -C "$ROOT" status --porcelain | wc -l)" -ne 0 ] && [ "$ALLOW_DIRTY" -eq 0 ]; then
    die2 "the working tree is DIRTY — a release builds from a committed tree (the exes embed the source; use --allow-dirty for dev probes)"
fi

WAILS_VER="$(python3 - "$ROOT/webui/wails.json" <<'PY'
import json, sys
print(json.load(open(sys.argv[1]))["Info"]["productVersion"])
PY
)"
if [ "$WAILS_VER" != "$VERSION" ]; then
    die2 "version drift: webui/wails.json Info.productVersion = $WAILS_VER but VERSION = $VERSION (the one-version-truth law)"
fi

say "version $VERSION — the double-build ritual begins"

# --- The CLI exe: build twice + one cold-cache leg -------------------------

cd "$ROOT"
CLI_FLAGS=(-trimpath -ldflags "-X main.version=$VERSION")
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 "$GO_BIN" build "${CLI_FLAGS[@]}" \
    -o "$DIST/GitCurator-cli.exe.a" ./cmd/gitcurator-cli
GOOS=windows GOARCH=amd64 CGO_ENABLED=0 "$GO_BIN" build "${CLI_FLAGS[@]}" \
    -o "$DIST/GitCurator-cli.exe.b" ./cmd/gitcurator-cli
COLD_CACHE="$(mktemp -d)"
GOCACHE="$COLD_CACHE" GOOS=windows GOARCH=amd64 CGO_ENABLED=0 "$GO_BIN" build "${CLI_FLAGS[@]}" \
    -o "$DIST/GitCurator-cli.exe.cold" ./cmd/gitcurator-cli
CLI_A="$(sha "$DIST/GitCurator-cli.exe.a")"
CLI_B="$(sha "$DIST/GitCurator-cli.exe.b")"
CLI_C="$(sha "$DIST/GitCurator-cli.exe.cold")"
say "CLI exe  A=$CLI_A"
say "CLI exe  B=$CLI_B"
say "CLI exe cold=$CLI_C"
[ "$CLI_A" = "$CLI_B" ] || die3 "the CLI exe differs across two builds ($CLI_A vs $CLI_B)"
[ "$CLI_A" = "$CLI_C" ] || die3 "the CLI exe differs across cache states ($CLI_A vs $CLI_C)"

# --- The GUI exe: build twice + one cold-cache leg --------------------------

cd "$ROOT/webui"
GUI_FLAGS=(-platform windows/amd64 -trimpath -ldflags "-X main.version=$VERSION")
CGO_ENABLED=0 "$WAILS_BIN" build "${GUI_FLAGS[@]}" 1>&2
cp build/bin/gitcurator-shell.exe "$DIST/GitCurator.exe.a"
CGO_ENABLED=0 "$WAILS_BIN" build "${GUI_FLAGS[@]}" 1>&2
cp build/bin/gitcurator-shell.exe "$DIST/GitCurator.exe.b"
GOCACHE="$COLD_CACHE" CGO_ENABLED=0 "$WAILS_BIN" build "${GUI_FLAGS[@]}" 1>&2
cp build/bin/gitcurator-shell.exe "$DIST/GitCurator.exe.cold"
GUI_A="$(sha "$DIST/GitCurator.exe.a")"
GUI_B="$(sha "$DIST/GitCurator.exe.b")"
GUI_C="$(sha "$DIST/GitCurator.exe.cold")"
say "GUI exe  A=$GUI_A"
say "GUI exe  B=$GUI_B"
say "GUI exe cold=$GUI_C"
[ "$GUI_A" = "$GUI_B" ] || die3 "the GUI exe differs across two builds ($GUI_A vs $GUI_B)"
[ "$GUI_A" = "$GUI_C" ] || die3 "the GUI exe differs across cache states ($GUI_A vs $GUI_C)"

# --- The zip: build twice over the A exes -----------------------------------

cd "$ROOT"
ZIP_A="$DIST/GitCurator-v$VERSION-windows.zip.a"
ZIP_B="$DIST/GitCurator-v$VERSION-windows.zip.b"
"$GO_BIN" run ./build/mkzip -version "$VERSION" \
    -exe "$DIST/GitCurator-cli.exe.a" -gui-exe "$DIST/GitCurator.exe.a" -out "$ZIP_A"
"$GO_BIN" run ./build/mkzip -version "$VERSION" \
    -exe "$DIST/GitCurator-cli.exe.a" -gui-exe "$DIST/GitCurator.exe.a" -out "$ZIP_B"
ZIP_SHA_A="$(sha "$ZIP_A")"
ZIP_SHA_B="$(sha "$ZIP_B")"
say "zip      A=$ZIP_SHA_A"
say "zip      B=$ZIP_SHA_B"
[ "$ZIP_SHA_A" = "$ZIP_SHA_B" ] || die3 "the zip differs across two builds ($ZIP_SHA_A vs $ZIP_SHA_B)"

# --- The verdict -------------------------------------------------------------

say "PASS — byte-identical rebuilds: CLI exe + GUI exe + zip (warm x2 + cold x1 per exe)"
say "the release zip: $ZIP_A"
say "  files + bytes + sha256 above are the release-notes numbers (AC-23's sha256 record)"
