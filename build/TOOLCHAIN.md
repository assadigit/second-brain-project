# Toolchain record (required by ADR-0002: record the exact toolchain)

| When | Toolchain | Where |
|---|---|---|
| 2026-10-15 (SESSION-001, T-00) | go1.27.2 linux/amd64 | sandbox `c-6ad0379f-…` (Debian 13 trixie) |

- `go.mod` pins the minimum supported language version `go 1.22` per
  ADR-0002; the build toolchain used so far is go1.27.2.
- PoC modules (`poc/`) may declare newer `go` directives when their
  dependencies require them; the main module keeps the 1.22 floor.
