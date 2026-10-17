# second-brain-project (codename)

> **Private workspace for the Go port of GitCurator** — the personal second-brain library
> builder (Python/PyQt6 → Go). Codename until the branded name is chosen; stays private
> until launch.

**Status: Phase 0 CLOSED** — all three PoC verdicts final (Wails ✅ · tls-client ✅ ·
gotd/td ✅ proven on the owner's account, then **demoted to dormant by the owner's
Option-B stamp**: the queue rides the Worker only — ADR-0004 Am.3); harness skeleton +
Python baselines archived; SESSION-001/002 closed. **Phase 1 CLOSED** (SESSION-003–005):
T-10…T-17 ALL DONE with generated-golden parity — the characterization suites
(test_core incl. prompts/usermemory, taxonomy, note formats), the import-grammar
differential, and the byte-stable config round-trip are green. **Phase 2 CLOSED
(SESSION-006–010; all six tasks)**: T-20 `internal/store` (modernc.org/sqlite
v1.36.1 pinned to the go 1.22 floor; the ADR-0003 DSN pragmas; **ADR-0003 Am.1's
CAST read law**); T-21 `internal/notestate` (the §4.4 moves-as-corrections engine;
+ the SESSION-008 CRLF fingerprint fidelity fix); T-22 `internal/notereads` (v0.71.0
THE ONE-OPEN LAW — the second scan of unchanged notes opens NOTHING:
reference-counted opens 4→0, db_hits 6); **T-23 `internal/websitepipeline/state`
(SESSION-010) — WebsiteStateDB: THE ONE SPELLING boundary, the 2×attempts backoff,
both settlements, the v0.64.0 healing pass; T-24 `internal/store/repos.go` —
CacheDB: the retry queue's no-stacking/normalized-key laws, THE SETTLED REPOS
LEDGER + the ♻️ revive door, the fail-closed record_404, the open-time migrations
(and the CAST read law GENUINELY LIVE on the TIMESTAMP-declared columns);
T-25 `internal/batch` — LinkTracker (the 5-phase manifest with FULL ACCOUNTING +
the json.dump(indent=2) byte-parity writer) + VaultIndex (the reconciliation
ground truth: skip sets, the 4000-char read, the site-map gates). SESSION-009's
T-23/T-24 were lost at an environment wipe while push-pending — re-ported
independently in SESSION-010, which now pushes after every task. 425 tests
race-clean.**
**Phase 3 CLOSED (SESSION-011–014)**: T-30 `internal/llmclient`
(commit `f2073b5`) — the three providers on Python-json.dumps-identical
wire bytes (the payload goldens), the JSON-mode memoized 400-fallback,
extract_json, the llama.cpp detection engine + the /props grab, the
autodetect policy; T-31 `internal/webfetch` (commit `4d93877`) — the
whole fetch ladder: the routes, the OWN redirect loop, the Chrome-131
header set, the rate limiter with penalties, the caps + capped
decompression, the failure taxonomy, the DoH verdicts, the dead-page
rescue + the Wayback courtesy, and the tls-client impersonation door;
T-32 `internal/webextract` (commit `833b21b`) — the HTML extraction
port on the HAND-ROLLED CPython-3.12 html.parser tokenizer (option b,
zero new imports), the charset ladder with four GENERATED data tables;
and **T-33 `internal/connectioncheck` (commit `e5846bd`;
SESSION-014) — the RE-SCOPED Test Connection battery (Vaults, LLM,
GitHub, and the Worker leg promoted to queue-reachability per
ADR-0004 Am.3 — the Telegram section is dropped), the sentinel trio
(internet/AI/proxy with the any-HTTP-answer law and the warn-once
offline verdict), netctx's verify_ssl split, the UA table pinned from
the wire, and the 22-case short_detail differential; and T-34
`internal/workerclient` (commit `6ffc4e9`; SESSION-014) — the
desktop↔Worker REST client: pairing, the HMAC signing law (the
sha256hex-key), all endpoints, the 1010 UA, the proxy bypass, the
honest poll, sync_id idempotency (a hand-rolled uuid4 — zero new
dependencies), plus THE LOOPBACK RIG (test_synclive's _FakeWorker
ported: verifyHmac byte-for-byte, the golden signature vectors, the
directory doors T-49 reuses). 872 tests race-clean; the exit gate's
five stones all laid (payload goldens, ladder/redirect parity,
fixture parity, result shapes + UA table, HMAC loopback rig).**
**Phase 4 OPENED (SESSION-015)**: T-40 `internal/websitepipeline`
(the engine room, commit below) — the per-link state machine on top of
T-23's state layer: the note builders + the review scans + the
consolidation (ONE NOTE PER SITE) + the master table's full grammar
(parse/writers/stamp/harvest/prune/refresh) + THE BANISHMENT with its
confirmation gate (confirmed/declined/timeout/defer/auto) and the
run tally + the settlement + the four batch drivers; the table
grammar floor (decomtable), the handdelivery consumption surface, and
the directorysync banish-gate helpers carved as the dependency
floors T-43/T-49 complete; the DIFFERENTIAL GOLDEN (51 sections,
frozen clocks, generated from the unmodified reference). **T-41
`internal/websitepipeline/directorynote.go` (SESSION-016) — THE LAW
sweep + the Website Directory regeneration** (spec-core-website-
directory.md, the standalone spec the task list had always
referenced): sweep_banned_notes (app-owned notes on banned domains →
.trash/banned-domains; the banned platforms' _inbox tables quarantine
whole) + build_website_directory (the `000 📚 Website Directory.md`
regeneration — taxonomy-ordered categories/subcategories, the ≤80
alias / ≤140 TL;DR trims, the full-lowercase sort keys, the _review
todo callout) + the constructor's real sweep (the T-40 seam closed);
THE OUTPUT GOLDEN (33 sections, byte-compared, generated from the
unmodified reference) — and three reality-wins fixes the golden forced
in the shared primitives (universal newlines in readHeadRunes,
maximal-subpart U+FFFD in decodeUTF8Replace, files-before-subdirs in
walkVault). **T-42
`internal/githubpipeline` + `internal/batch/summary.go` (SESSION-017)**
— the GUI processing worker's GitHub loop per spec-gui-processing:
the per-repo gate chain (non-GitHub → dead-set → dismissed → settled
ledger → invalid → vault-index dedupe BEFORE any API call), the
get_repo error paths (403+rate-limit 3605 s wait + retry, the
401-anonymous token drop, the 404 strike → the `_missing` note +
notfound-links.md, the manifest-failure else), the two-layer dedupe,
fork→parent analysis, the enrichment + the banner goroutine, the
LLM 3-attempt ladder with the README injection guard, the disk-full
protocol, the reports (the summary log + rotation, the final Markdown
report), the batch-undo diff, and the v0.39.0 scorecard — THE GOLDEN
(12 sections, frozen clocks, generated from the unmodified reference
through the GUI-layer import stubs) + the behavior battery (the
rate-limit laws, the 404-strike flow, the 401 fallback, the undo
laws). Plus the reality-wins float fix: notebuilder's cred_score is
now the round(x,1) FLOAT the reference renders ("90.0/100"). **T-43
`internal/handdelivery` (SESSION-018) — THE FOURTH DOOR COMPLETE**
(spec-core-hand-delivery.md): the module T-40 floored is finished —
the Chrome launch family (FindChrome/OpenInChrome with the patchable
launch seams; loopback never opens), WalledRetryRows (the picker's
data over the RetryRowsState seam) + PendingHandLinks (the fifth
door's to-do list), the batch ConsumeDelivered (stamps the queue,
keeps the file; dry-run takes without stamping) with the PageIsRealFn
seam staged for T-44 (the honest except-path degrade), StampHandRows
(the master-table queue stamp), and the enqueue's openChrome leg —
with THE GOLDEN (5 sections, frozen clocks, generated from the
unmodified reference; hand_delivery is core/ and stdlib-only, no
stubs needed) + the ported test_handdelivery/test_handharvest
selections (54 module tests + 20 pipeline legs); two reality-wins
fixes: the blank-body law (Python bytes.strip is ASCII-only — a
NBSP-only page IS delivered) and a latent T-42 flake (the report's
category table iterated a Go map — ties flipped randomly; now rides
the insertion-order slice). **T-44 `internal/chrometabs`
(SESSION-019) — THE FIFTH DOOR COMPLETE** (spec-core-chrome-tabs.md,
with the SESSION-019 amendment: the CDP WebSocket is hand-rolled
stdlib net + crypto/rand + encoding/base64 exactly as the reference
hand-rolls it on Python's socket module — the old gorilla/websocket
Go-note is superseded, zero new dependencies): the RFC 6455 frame
codec (masked client frames, all three length encodings,
fragmentation + control frames, the 64 MB runaway guard), the
DevTools socket (the handshake shape, the call/event round-trip,
ping→pong, the error/stall/timeout laws), page_is_real wired into
handdelivery.PageIsRealFn (the seam T-43 staged) with the deferred
auto-door discard battery landed (an auto-delivered error/challenge
page is discarded, the link re-asks; the owner's saves stay
trusted), the load watch (complete at an http(s) URL, still for the
settle, real by the verdict), the candidates, the session ladders
(adoption through DevToolsActivePort, container/headless retries),
the real-profile rung (the attach link, chrome-attach.bat), the
proxy legs with per-site route memory, and the fetch driver honesty
(no Chrome, loopback, dry-run) — with THE GOLDEN (9 sections, mask
bytes folded, generated from the unmodified reference; the pyurl
lenient urlsplit is CPython 3.12's exact shapes incl. the bracket
law) + the ported test_fifthdoor selections (63 module tests + 2
pipeline legs). **T-45 `internal/vaultscan` (SESSION-020) — THE VAULT
SCAN COMPLETE** (spec-core-vault-scan.md, with the SESSION-020
amendments): the radar (the inventory walk with the lock/curated
counts, the move candidates, the CPython urlparse shapes
hand-rolled), ONE NOTE PER SITE (the consolidation magnet, the keeper
rank), the regex brain (declared home → the domain's settled siblings
→ the taxonomy's vocabulary, the repos twin through CATEGORY_FOLDERS),
THE WEIGHED EVIDENCE (the folder census, the 0.0-1.0 scale with
banker's rounding, the CLOSE CALL verdict, the alignment door), the
chunked LLM filing (the exact user message, per-chunk failure
isolation, the honest review list), the validation door (the
fresh-eyes absorption, the depth law, the landable move with the `_v`
bend pre-computed as `as`), the picking door + the lock law, the
census stamp, and the apply pass (verified landings, the spoken
failure ledger, the deletions through the banishment machinery's own
skip doors, the repos ledger re-point) — with THE GOLDEN (21
sections, paths folded, generated from the unmodified reference) +
the ported test_vaultscan/test_censusstamp/test_picklock/
test_weighhomes/test_reposcan selections (215 tests; the repos
banishment floor — the gate's eyes + the burial — landed with T-45 in
`internal/repobanish`, its remaining doors stay T-46's); the
pointer-threading law ([]*NoteRow — Go struct copies would break the
reference's dict aliasing, found live). **T-46 `internal/vaulthygiene`
+ the completion of `internal/repobanish` (SESSION-021) — THE
STRAY-NOTES DOOR + THE REPOS BANISHMENT COMPLETE** (spec-core-vault-
hygiene.md + spec-core-repo-banish.md, with the SESSION-021
amendments): the whole hygiene module — the classifiers (empty stub,
app-owned bodyless with the fence laws, the lock doors, the note-name
shape, the identity fold), the [[wiki-link]] eyes + the head-identity
harvest, and THE ONE WALK (each .md opened exactly once — identities +
stub candidacy + link targets from the same read, the shared
notereads cache or the direct bounded read, the three fingerprints:
shape / name-twin / linked-ghost, the lock guard, the master-index
scaffold, the legacy root reports, heal_machine_vaults over both
vaults) — with THE GOLDEN (9 sections, paths + walk-order folded,
generated from the unmodified reference) + the ported
test_strayheal/test_vaulthygiene selections + test_picklock's
locked-bodyless law + test_vaultscan's one-open hygiene-sharing leg;
and the repobanish completion — revive_banished_repos (the ♻️ door
back) + run_repo_banishment_pass (the confirmation-gate orchestration:
the ungated revive, the count spoken before the ask over the injected
channel factory, the verdict gates, the tally with the zero-run
answer) + the ported test_repobanish selections incl. the never-fetch
promise and the no-op-settle locking pin. **T-47 `internal/vaultseal`
(SESSION-022) — THE SEAL COMPLETE** (spec-integrations-vaultseal.md +
ADR-0009, with the SESSION-022 amendments): the git-CLI-subprocess
seal of a vault into a PRIVATE GitHub mirror — the construction (the
v0.83 creation guard, the local-only belt, the autoPush/token
degradation), status(), seal() (ensure_repo: the nested-repo guard →
`git init -b main`, the managed .gitignore block, the local-only
identity, the token-less remote add, the 422 race re-check; the belt:
the second managed block REBUILT + the untrack law — `git rm -r
--cached`, NEVER a disk delete; the unchanged skip; the commit-message
doors), THE DIVERGENCE LADDER under push (plain push → fetch +
FETCH_HEAD token redaction → stale-rejection retry / rebase /
unrelated-histories merge / the seal-rescue branch — main is NEVER
force-pushed), resync() (write-tree delta counts, the commit-tree
reconciliation with both parents, the empty-tree baseline), the four
config bridges (manual_seal's deliberately-NO-belt law), the standalone
CLI, THE DRIFT SEAL's pure half (the plan + the runner with THE QUIET
LAW; the publish door rides T-49's seam), CPython 3.12's fnmatch,
env scrubbing + HideWindow + THE REDACTION LAW (the token masked in
any captured output) — with THE GOLDEN (16 sections, paths + clock +
git-dates frozen, generated from the unmodified reference; it caught
the universal-newlines find live) + the ported
test_sealfix/test_resyncfix/test_localonly/test_driftseal selections
+ the safety-law battery; the belt floor landed with it in
`internal/directorysync` (the pattern builders + the shipped-config
embed — T-49 completes the module). **T-48 `internal/goodrepos` +
`internal/reporadar` (SESSION-023) — THE PUBLISHER + THE RADAR
COMPLETE** (spec-integrations-goodrepos.md + ADR-0009, with the
SESSION-023 amendments): the curated vault published as its PUBLIC
good-repos directory — the tiny frontmatter subset + the note parser
(the display-name ladder url → aliases → filename stem, the TL;DR +
"What is it?" fallback), the emoji README writer (the GitHub anchor
math, the plural doors, the stars-desc sort, the display-name orders
that beat the emoji code points), the staging lifetime (mkdtemp
OUTSIDE any repository, reuse across publishes, the atexit →
CleanupStaging convergence), THE CREATION GUARD (the empty-name
fail-fast, the 404 never-creates-without-opt-in error, the private
POST payload, the 422 race re-check), the fetch-alignment + the
skip-when-unchanged cycle, the one-time-token push with the defensive
remote removes, THE REDACTION LAW + the FETCH_HEAD scrub, the config
bridge + the standalone CLI — and the read-only repo radar: the
remote-URL grammar (configparser → the minimal INI scan), git's
content-addressed blob SHA, the diversified round-robin sampler, the
candidate ranking (archived/fork filtered, vault-ish first, size
NEVER an emptiness signal — the live lesson), THE DOMINANCE LADDER
(≥3 matches AND ≥2× the runner-up, else ambiguity/weakness = a
report, never a guess), the DETECTION ladder (REMOTE → NAME →
FINGERPRINT), the only-EMPTY-slots law, the quiet-startup gate, and
the never-raises doors (a panicking API surface recovered and spoken)
— with the ported test_goodrepos/test_reporadar selections (the
reference's whole suites) + the local-bare-repos publish-flow battery
+ the safety-law battery + THE GOLDEN (25 sections, paths/clock/
git-dates frozen — the Go port produces the reference's exact
byte-deterministic publish sha `a300991`) + the 12-section radar
golden. 1658 tests race-clean.
Next: T-49 directorysync.

## Map

| What | Where |
|---|---|
| **HOW SESSIONS WORK (start here)** | [`docs/SESSION-PROTOCOL.md`](docs/SESSION-PROTOCOL.md) — the loop every session follows |
| **The state card / next prompt** | [`docs/NEXT-SESSION.md`](docs/NEXT-SESSION.md) — rewritten at every session close |
| **Session reports (the history)** | [`docs/sessions/`](docs/sessions/) — numbered SESSION-NNN.md files |
| The complete migration blueprint (73 docs) | [`docs/port-blueprint/`](docs/port-blueprint/) — start at its `README.md`, then `00-executive-summary.md` |
| The mission spec that produced the blueprint | [`docs/Port-investigation.md`](docs/Port-investigation.md) |
| App icon (ready-made brand mark, multi-size) | [`assets/icons/gitcurator.ico`](assets/icons/gitcurator.ico) — swap the file later to rebrand |
| Owner decisions (all Q&A rounds) | `docs/port-blueprint/14-risks-assumptions-open-questions.md` §6 |

> **Starting a new work session?** Open a fresh AI chat and paste the prompt from
> `docs/NEXT-SESSION.md` §5. That's the whole workflow.

## The goal in one paragraph

A single Windows executable (Go engine + Wails GUI) that behaves byte-identically to the
Python app, opens the owner's real data in place, and fixes the sluggishness: cold start
≤ 1.5 s, vault scan ≤ 50% of Python, idle memory ≤ 200 MB. Delivered as both a
`setup.exe` (silent WebView2 bootstrap when missing) and a deterministic portable zip.
The unmodified Python reference repo (`assadigit/GitCurator`, pinned rev — at blueprint
time v0.91.0 / `8fd4179`) stays the live product until Go GUI parity, and is never
modified by this migration; differential tests run both apps side by side.

## Laws (binding, from the blueprint)

- Reality wins over specs: when source code contradicts a spec, fix the spec in the same commit.
- No visual redesign during the port — replicate the documented design (ADR-0001).
- Preserve observable behavior, including quirks; only the five pre-classified fixes are allowed.
- Never commit credentials, session files, or the owner's real `config.json` / `cache.db`.
- Tests are never weakened, skipped, or deleted to get green.
