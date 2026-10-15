# Module Spec — `gitcurator/tools/` (developer utilities)

- **LOC**: ~1,300 total · **Go home**: selectively ported as `go-desktop/tools/` commands ·
  **Status**: port selectively (dev utilities, not runtime dependencies)
- **Responsibility**: developer/owner utilities — release packaging, golden runners, safety
  scanners, backfill, mirror CLI, link builder, recall hooks, diagnostics.

## Inventory

| Tool | Role | Go disposition |
|---|---|---|
| `build_zip.py` (285) | the deterministic release zip (laws: tracked-files-only via `git ls-files -- app`, banned-secrets guard, launcher flattening, `.bat` CRLF+ASCII, fixed timestamps, sorted entries, DEFLATE 9, sha256 print) | **Port** — the Go packaging owns its own deterministic zip (ADR-0010); the algorithm is the spec (`10-build-release-packaging.md` §2) |
| `run_golden_websites.py` (~120) | the offline golden run: full Websites pipeline (prompt loading, slot filling, taxonomy validation, note building, report rendering) against `tests/golden/websites.json` (30 entries) with fake fetch + fake LLM answering from the golden file; exit 0 iff every answer was a valid taxonomy name; `--live` mode for side-by-side backend comparison reports | **Port** — becomes the Go differential golden driver (a first-class citizen of `09-testing-and-parity.md`) |
| `pick_golden_links.py` | deterministic 30-link pick from the owner's bookmark export | Reference only (data already committed) |
| `scan_vault_edits.py` | read-only vault safety scanner (pre-run snapshot diffing aid) | Port (S) — same CLI contract |
| `snapshot_vault.py` | read-only vault snapshotter | Port (S) |
| `backfill_websites.py` | one-time websites backfill driver | Port (S) — Worker client already ported |
| `build_links.py` | Phase-6 linking batch builder (embeddings → candidates → confirm → suggestions → related blocks) | Port (M) — drives `internal/mirror/linking` |
| `mirror_manual.py` | the apply entry for `run_mirror` (`--apply` flag) | Port (S) |
| `add_recall_hooks.py` | recall-hook pass runner (dry-run first) | Port (S) |
| `diagnose_code.py` | offline Telegram code-request diagnostic | Optional (S) |
| `integration_snippet.py` | documentation-only example (never imported) | Not ported |
| `test.py` | Telethon credential checker (manual) | Not ported (superseded by the gotd login flow) |

## Notes

- Tools are standalone CLIs (argparse) run from `app/`; they share the core modules.
- In Go they become small `main` packages under `go-desktop/tools/<name>/` sharing
  `internal/…` — same flags, same outputs (differential-comparable where applicable).
