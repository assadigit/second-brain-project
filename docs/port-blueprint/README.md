# port-blueprint/ — GitCurator Python→Go Migration Blueprint

> Produced by the investigation session specified in `Port-investigation.md`, from a full read
> of the repository at **v0.91.0** (HEAD `8fd4179`, branch `main`, clean tree). This directory
> is the *only* artifact that session created. It contains no credentials and no application
> code — only specifications, decisions, and plans.

## What this is

A complete, evidence-based plan for migrating the **Python desktop application** (123
production modules, 75,605 LOC under `app/gitcurator/`) to **Go**, preserving behavior and data
while fixing the perceived sluggishness and producing a reliable Windows executable. The
Cloudflare Worker (JS) and the Next.js web apps are **not migrated** (ADR-0005) — their wire
contracts are pinned so the Go client stays compatible.

## Reading guide (ordered, for the implementation agent)

1. `00-executive-summary.md` — the whole picture in one page.
2. `12-migration-roadmap.md` — the phase plan and gates.
3. `13-implementation-task-list.md` — the ordered, actionable tasks (start at T-00).
4. `decisions/` — 10 ADRs (binding; three carry PoC gates).
5. `01-repository-inventory.md` → `02-current-architecture.md` → `03-behavioral-specification.md`
   — what exists and what it does.
6. `04-target-go-architecture.md` + `05-python-to-go-mapping.md` — what to build and where each
   Python module lands (incl. the 20 semantic hotspot table).
7. `module-specs/` — the per-module contracts (read the specs for whatever you implement that
   day; each cites `path:line` evidence).
8. `06-dependency-migration.md` — library replacements and their parity risks.
9. `07-ui-and-performance-assessment.md` — GUI strategy + confirmed-vs-hypothesis performance
   findings + the migration/redesign boundary.
10. `08-data-and-compatibility.md` — every file, format, and protocol contract (incl. the UA
    table and the Worker API).
11. `09-testing-and-parity.md` — the parity strategy and the differential harness design.
12. `10-build-release-packaging.md`, `11-security-reliability.md` — shipping and safety.
13. `14-risks-assumptions-open-questions.md`, `15-acceptance-criteria.md` — what can go wrong,
    and the definition of done.
14. `16-next-session-implementation-prompt.md` — the ready-to-copy briefing for the next agent.
15. `traceability-matrix.md` — behavior → evidence → spec → Go → test → task → AC.

## The recommended approach in one paragraph

CLI-first, GUI-second (ADR-0007): port the pure foundations byte-exactly (links, storage,
config, sanitizers), then the stores, the network/LLM layer, the pipelines and integrations —
proving each against the ported Python tests and a **differential harness** that runs the
unmodified Python app and the Go app on identical inputs — release a **CLI parity milestone**
the owner can use alongside the Python app, then build the **Wails GUI** replicating the
documented design (ADR-0001, no redesign), and finally package the deterministic Windows zip
(ADR-0010). Telethon is replaced by gotd/td (ADR-0004), curl_cffi by a Chrome-TLS Go client
(ADR-0006), git stays a CLI subprocess (ADR-0009), SQLite stays the same files via modernc
(ADR-0003), and the app opens the owner's real data in place (ADR-0008).

## Assumptions, limitations, unresolved decisions

- **The codebase is 3–4× the size the mission brief assumed** (<20k claimed; 75.6k measured) —
  the roadmap is phased accordingly.
- **Three PoCs gate the plan**: gotd/td login+queue (ADR-0004), Chrome-TLS impersonation
  (ADR-0006), Wails+WebView2 (ADR-0001). Failures have documented fallbacks.
- **This session could not run the GUI** (PyQt6 unavailable in the sandbox; pure-stdlib test
  modules pass: test_core 34/34, test_notereads 16/16, test_webnote_v2 10/10) and took **no
  performance measurements** — doc 07 separates in-repo measured facts from hypotheses and
  defines the baseline procedure; open questions with verification steps are registered in doc
  14.
- **Status**: complete and internally cross-checked; awaiting implementation. Update spec files
  in-repo whenever reality disagrees with them (the next-session prompt makes this a standing
  rule).

## Directory map

```
port-blueprint/
├── README.md                        ← you are here
├── 00-executive-summary.md … 16-next-session-implementation-prompt.md   (17 main docs)
├── module-specs/                    (28 grouped specs covering all 123 production modules)
│   ├── spec-core-*.md               (15 core-engine specs)
│   ├── spec-gui-*.md                (13 GUI specs)
│   ├── spec-integrations-*.md       (4)
│   ├── spec-cloud-*.md              (3, incl. the dormant quartet)
│   ├── spec-cli.md · spec-tools.md
│   └── spec-worker-contract.md      (JS Worker — contract only, not migrated)
├── decisions/                       (10 ADRs: 0001 GUI … 0010 packaging)
└── traceability-matrix.md
```

## Canonical copy note (2026-10)

This blueprint is mirrored to the private implementation repo
**`assadigit/second-brain-project`** at `docs/port-blueprint/` (owner decision — ADR-0002 Am.1);
that copy becomes canonical once implementation starts. The copy in this repository remains the
v0.91.0 audit-of-record snapshot, including the owner Q&A amendments (doc 14 §6) and the ADR
amendments (0001/0002/0007/0010 Am.1, 0010 Am.2).
