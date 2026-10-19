# Toolchain record (required by ADR-0002: record the exact toolchain)

| When | Toolchain | Where |
|---|---|---|
| 2026-10-15 (SESSION-001, T-00) | go1.27.2 linux/amd64 | sandbox `c-6ad0379f-…` (Debian 13 trixie) |

- `go.mod` pins the minimum supported language version `go 1.22` per
  ADR-0002; the build toolchain used so far is go1.27.2.
- PoC modules (`poc/`) may declare newer `go` directives when their
  dependencies require them; the main module keeps the 1.22 floor.

## Amendment 1 — T-31's dependency-forced go directive (2026-10-16, SESSION-012)

| When | Toolchain | Where |
|---|---|---|
| 2026-10-16 (SESSION-012, T-31) | go1.27.2 linux/amd64 (unchanged) | sandbox rebuild #7 |

- **Reality correction (found at implementation time, evidence at hash
  level)**: importing `github.com/bogdanfinn/tls-client v1.16.0` — the
  ADR-0006 Am.1+2 impersonation pin — into the MAIN module forces the
  `go` directive up: the pin's own published go.mod (module proxy
  `.../tls-client/@v/v1.16.0.mod`) declares `go 1.24.1`, and the Go
  toolchain (go1.27.2) refuses builds and rewrites the directive
  (`go mod tidy` from a hand-set `go 1.22` computes the graph minimum
  ≥1.24). The PoC had dodged this by living in the separate `poc/`
  module ("PoC modules may declare newer go directives"); T-31 imports
  the door into the main binary (ADR-0006 §4: "always compiled in"),
  making the collision unavoidable.
- **Resolution** (ADR-0002's own letter permits "Pin `go 1.22` (or
  newer)"): the directive rides the dependency-forced `go 1.24.1`;
  `modernc.org/sqlite` stays pinned at `v1.36.1` (untouched
  throughout); our own source keeps the 1.22 language floor. See
  ADR-0002 Am.2 for the full decision record and the rollback path.

## Amendment 2 — the exact tarball sha256 (2026-10-19, SESSION-032)

| When | Toolchain | Where |
|---|---|---|
| 2026-10-19 (SESSION-032, T-64) | go1.27.2 linux/amd64 (unchanged) | sandbox rebuild #26 (the twenty-sixth full wipe) |

- The exact sha256 of the go1.27.2.linux-amd64.tar.gz tarball, verified
  byte-equal to go.dev's published `?mode=json` listing AND to the
  locally-computed hash of the downloaded file:

  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`

- **The transcription-corruption fix (reality wins)**: the session
  cards and SESSION-031's report had carried a 63-character rendering
  (`ecadb99091…` — one dropped character). The verification procedure
  itself was always sound (download from go.dev over HTTPS + compare
  against go.dev's published sha); the recorded STRING was wrong. This
  amendment pins the 64-character truth so future rebuilds compare
  against the right bytes.

## Amendment 3 — the Wails build-time toolchain (2026-10-19, SESSION-033, T-70)

| When | Toolchain | Where |
|---|---|---|
| 2026-10-19 (SESSION-033, T-70) | go1.27.2 linux/amd64 + **wails CLI v2.16.0** (GOBIN=/home/z/.local/go/bin) | sandbox rebuild #27 (the twenty-seventh full wipe) |

- **Build tool**: `GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` (with go1.27.2). The pin is
  ADR-0001 Am.2's confirmed version; v2.17.0 exists on the proxy and is
  deliberately NOT taken. Build-time only — never a runtime dependency of the
  main module.
- **The runtime story**: the PoC shell is the NESTED module `webui/`
  (`go 1.25.0` — wails v2.16.0's own floor, verified from the proxy's
  published .mod; `require github.com/wailsapp/wails/v2 v2.16.0` +
  `replace` to the parent). The MAIN module's runtime set stays the
  authorized four, byte-level unchanged. Full decision record: ADR-0002 Am.3
  (surfaced BEFORE the first `wails` invocation, per the session mandate).
- **Linux build legs**: CGO + the rootless webkit2gtk-4.1 tree
  (~/.local/lib-wails) per docs/poc-artifacts/t-70-wails-poc.md's recipe
  (apt --download-only + dpkg -x + the .pc rewrites + the LD_PRELOAD
  libexec shim; WEBKIT_DISABLE_COMPOSITING_MODE=1 under Xvfb).
  **Windows exe**: CGO-free cross-compile (`wails build -platform
  windows/amd64` — the WebView2 loader is pure Go).
- **Frontend step**: none for the PoC (vanilla assets; wails.json's
  `frontend:install`/`frontend:build` empty). The production T-72+ frontend
  tooling question (vanilla vs vite) is OPEN for the owner.

### Amendment 3 correction (2026-10-19, SESSION-033, T-70's landing)

ADR-0002 Am.3's claim that Go's internal-package rule forces the
`guibridge/` seam is FALSE (the landing's boundary probe: a direct
`internal/` import from the nested webui module builds and vets clean —
exit 0 both). Full correction record: ADR-0002 Am.4. The seam stays for
its architectural value (doc 04 §2's event-seam shape, main-battery
tested); the toolchain facts in this amendment (the pin, the frontend
step, the CGO story) are unaffected.

## Amendment 4 — the 32nd wipe's symlink-syscall degradation (2026-10-20, SESSION-037)

| When | Toolchain | Where |
|---|---|---|
| 2026-10-20 (SESSION-037, T-74) | go1.27.2 linux/amd64 (unchanged) | sandbox rebuild #32 (the thirty-second full wipe) |

- **The deviation, recorded honestly**: during the 32nd rebuild the sandbox
  began blocking the `symlink(2)` syscall outright (`ln -s` and Python's
  `os.symlink` both refused, after working earlier in the same session for
  the repo convenience links). The webkit recipe's "relink the 26 dangling
  dev symlinks to the system runtime copies" step therefore landed as
  **hardlinks to the RESOLVED system targets** (`readlink -f` first — the
  system `.so.0` files are themselves versioned symlinks): same bytes, same
  link-time resolution, and the `wails build` CGO link over the tree is the
  proof the tree is sound (links clean, the gate leg runs).
- **The recipe itself stays the symlink form** — on a box where symlink(2)
  works, the original recipe's relinks are the fresher shape (they follow
  system package updates; a hardlink pins today's inode). This amendment is
  the fallback record for any future session that hits the same block.
