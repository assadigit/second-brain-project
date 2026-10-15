# 14 — Risks, Assumptions, and Open Questions

Legend: **CONFIRMED** (evidence in repo) · **INFERRED** (evidence-based conclusion) ·
**RECOMMENDATION** (proposed handling) · **UNRESOLVED** (needs future verification, with the how).

## 1. Risk register (top risks with mitigations)

| # | Risk | Class | Likelihood × Impact | Mitigation |
|---|---|---|---|---|
| R1 | GUI replication drifts from the PyQt6 UX (missing states, dialogs, cadences) | CONFIRMED effort risk | M × H | workflow-preserving checklist per page in module specs; GUI workflow tests; owner preview checkpoints at M6 |
| R2 | gotd/td cannot replicate Telethon behaviors (auth UX, bot-chat iteration, read-ack) | UNRESOLVED (PoC T-01) | M × H | PoC before any telegram code; fallback = Worker-only queue (documented behavior change, new ADR) |
| R3 | Chrome-TLS fingerprint quality of tls-client below curl_cffi on real sites | UNRESOLVED (PoC T-02) | M × M | PoC on walled sites; utls fallback; worst case = hand-delivery/Chrome doors still solve those pages (the ladder exists) |
| R4 | Semantic drift (ordering, casing, rounding, timestamps, slicing) corrupts ledgers | CONFIRMED (Python semantics enumerated) | M × H | helper library + the 20-item hotspot table (05 §3) each with a dedicated parity test; differential corpus |
| R5 | Opening the owner's real `cache.db`/`config.json` corrupts or forks state | CONFIRMED hazard | L × H | open-in-place with same DDL; side-by-side lock; dry-run shadow cache; full backup before first Go run (owner ritual) |
| R6 | Scope explosion (75.6k LOC vs the mission's <20k assumption) | CONFIRMED | H × M | phased roadmap; CLI-first milestone; explicit out-of-scope register (ADR-0005, doc 07 §D) |
| R7 | WebView2 unavailable on the owner's Windows 10 | UNRESOLVED (PoC T-70) | L × H | bootstrapper; Fyne fallback (ADR-0001) |
| R8 | AV false positives on the unsigned Go exe disrupt the owner | INFERRED (known Go behavior) | M × L | mitigations doc 10 §9; signing later if needed |
| R9 | Wails v3 churn / v2 maintenance horizon | INFERRED | M × M | pin v2; the frontend is standard web tech — worst case the shell is replaceable without touching `internal/` |
| R10 | Owner depends on a quirk we classified as a bug and "fixed" | INFERRED | L × M | default preserve; fixes limited to the 5 pre-classified items (doc 09 §5) each with CHANGELOG-style notes |
| R11 | Determinism of the release zip breaks (timestamps, ordering) | CONFIRMED requirement | L × M | port build_zip laws exactly; double-build sha256 check in CI |
| R12 | Both apps running simultaneously (owner habit) | CONFIRMED hazard | M × H | same `app.lock` protocol across Go and Python (T-78) |

## 2. Assumptions (stated, so they can be challenged)

| # | Assumption | Basis | If false |
|---|---|---|---|
| A1 | The Worker (JS) and web apps remain live and unchanged during the whole migration | ADR-0005; "worker stays LIVE" in every recent changelog entry | contract tests catch drift; update `spec-worker-contract.md` + client |
| A2 | The owner's Windows 10 has (or can bootstrap) WebView2 | ADR-0001 PoC | Fyne fallback |
| A3 | git.exe is present (or installable) — already required by the Python release | Windows quickstart; VaultSeal design | installer check + guidance |
| A4 | One-time Telethon re-login is acceptable | INFERRED (single owner; login is a known ritual) | fallback R2 path |
| A5 | The Python app stays installable during transition (rollback path) | portability of the zip | keep the last Python zip release available |
| A6 | 3,192 Python tests correctly represent intended behavior (they are the spec of record) | repo culture; CI-gated | differential corpus + owner preview catch divergences |
| A7 | Sequential batch processing must be preserved (rate limits, ordering) | code evidence (delays everywhere) | none — verified in code |
| A8 | The repo's Go module can live at `go-desktop/` without breaking the Python zip tooling | build_zip only packs `app/` (verified: `git ls-files -- app`) | relocate module dir |

## 3. Open questions (with the verification procedure)

| # | Question | Why it matters | How to resolve |
|---|---|---|---|
| OQ-1 | gotd/td login + bot-chat iteration on the owner's account (R2) | the SYNC door's queue fetch | PoC T-01 |
| OQ-2 | tls-client license + fingerprint quality + SOCKS (R3) | the third door | PoC T-02 + license file read |
| OQ-3 | modernc DSN pragma syntax for WAL/busy_timeout | store correctness | 10-minute spike; docs of modernc.org/sqlite |
| OQ-4 | Wails v2 vs v3 pin + WebView2 on owner machine (R7/R9) | GUI shell | PoC T-70 + `winget`/registry check |
| OQ-5 | Exact PyGithub default User-Agent (parity nicety) | UA inventory completeness | read installed PyGithub version's source in the venv; decide explicit UA (recommended) |
| OQ-6 | Whether `requests` is truly unused (no hidden import) | dependency drop | `grep -rn "import requests" app/` in implementation (this session's searches found none) |
| OQ-7 | Windows `os.path.normcase` parity needs (only note_state used it) | moved-detection on Windows | test with mixed-case paths on Windows in Phase 2 |
| OQ-8 | Chrome CDP behavior against current Chrome versions (PUT /json/new law changed at Chrome 111) | fifth door | integration rig with local Chromium; keep GET fallback |
| OQ-9 | App icon asset for the .exe/.ico | packaging polish | ask the owner / reuse logo.svg |
| OQ-10 | `f"{stars:,}"` formatting for huge values | byte-parity of note bodies | golden tests with 6-7-digit stars |
| OQ-11 | Whether the owner wants the dormant GDrive/Cloudflare-manager features revived post-parity | scope | owner decision; register in Phase 8 |
| OQ-12 | Actual performance baselines (this session could not run the GUI) | before/after claims | T-04 procedure |

## 4. Findings requiring no further action (for the record)

- The five dead-code items and dormant modules (inventory §6) — documented, not ported.
- The stale README version claims — documented; not corrected (out of scope to edit).
- The Worker's vestigial `HMAC_SECRET` — Worker-side, untouched.
- `checkpoints` dead table — left in old DBs, not created by Go.

## 5. Session limitations (what this blueprint could NOT verify)

1. **No GUI execution**: PyQt6 is not installed in this sandbox; GUI behaviors come from code
   reading + the repo's own tests, not from running the app. Pure-stdlib test modules that do run
   here: `tests.test_core` 34/34 OK, `tests.test_notereads` 16/16 OK, `tests.test_webnote_v2`
   10/10 OK; GUI modules fail on `ModuleNotFoundError: PyQt6` (environment, not code defect).
2. **No external research** (no network): library versions/licenses for gotd, tls-client, Wails,
   modernc marked **[verify]** with exact checks (doc 06).
3. **No Windows execution**: Windows-only paths (Known-Folder, junctions, taskkill) are read from
   code, untested here.
4. **No performance measurements taken** — the mission forbids fabrication; doc 07 separates
   in-repo measured facts from hypotheses and defines the baseline procedure.

## 6. Owner decisions recorded (Q&A rounds, 2026-10, post-blueprint)

**Round 1 — 10 feasibility questions:**
- Q1 = **A** — replicate current design during the port; no redesign (redesign deferred).
- Q2 = **Yes** — one-time Telegram re-login acceptable (A4 answered; R2 mitigation holds,
  PoC T-01 still gates).
- Q3 = **In place** — the Go app opens the real `cache.db`/`config.json` where they are
  (ADR-0008 as written; backup ritual + cross-binary lock).
- Q4 = **Yes** — Windows-only v1.
- Q5 — CLI existed only to escape the slow Python GUI → **owner-facing CLI milestone CUT**;
  CLI stays as internal proving ground / test driver / emergency fallback (ADR-0007 Am.1,
  roadmap Am.1). GUI = first owner-visible release.
- Q7 = **Safest** — preserve all quirks byte-parity (matches the default; the 5 pre-classified
  fixes remain deferred).
- Q8 = **Tolerable** — unsigned exe first; code-sign later only if AV false positives appear.
- Q9 — owner **will upload a `.ico`** to the repo (path pending; OQ-9 → resolved-pending-file).
- Q10 — unanswered → default holds: dormant cloud quartet stays dormant (OQ-11 unchanged).

**Round 2 — WebView2 questionnaire:**
- Q1: Win10 not receiving updates for months (likely EOL). Q2: **Edge manually deleted**
  (Teams likely too) → **assume WebView2 absent on the owner machine** (A2 revised; R7
  likelihood M→H with the bootstrapper as the *expected* path — ADR-0001 Am.1).
- Q3: one-time silent install **acceptable**; **new non-blocking requirement registered:
  macOS/Linux flexibility** ("good to have"; v1 stays Windows-only).
- Q4: current design "acceptable, but not the polished and professional grade UI I am looking
  for"; absolute zero-dependency "tempting" → decision delegated: **Wails confirmed** (zero-dep
  benefit neutralized by the accepted one-time install; Fyne loses on fidelity AND on the
  professional-polish ceiling; the web stack makes the Phase-8 professional redesign cheap —
  upgraded from "maybe" to planned priority).
- Q5 = **Both** setup.exe and portable zip (ADR-0010 Am.1: dual artifact, bootstrapper chain).
- Q6 = **No offline installs** — always online at install time (Evergreen bootstrapper viable;
  fixed-version bundling rejected).

**Consequential updates:** ADR-0001 Am.1 · ADR-0007 Am.1 · ADR-0010 Am.1 · roadmap Am.1 ·
A2/A4/OQ-9 status above.

**Round 3 — icon and implementation repo:**
- Icon = **ready-made for now**: the repo's own brand mark (emerald tile + Lucide feather — the
  Website Directory favicon artwork) converted to a multi-size `.ico`; lives at
  `assets/icons/gitcurator.ico` in the implementation repo; owner may swap the file later
  (ADR-0010 Am.2; **OQ-9 fully resolved**).
- **New private repo `assadigit/second-brain-project`** created as the Go implementation home
  (codename; branded name later). Consequences in ADR-0002 Am.1: module path
  `github.com/assadigit/second-brain-project`; GitCurator stays the Python live-product repo;
  blueprint mirrored to `docs/port-blueprint/` as the implementation-canonical copy;
  differential reference pinned by git rev; A8 retired; public/branded-repo decision deferred to
  Phase 7/8.
