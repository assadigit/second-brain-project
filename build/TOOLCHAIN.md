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
