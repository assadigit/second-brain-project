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
**Phase 4 CLOSED (SESSION-015–026)**: T-40 `internal/websitepipeline`
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
**T-49 `internal/directorysync` (SESSION-024) — THE
DIRECTORY SYNC COMPLETE** (spec-cloud-directory-sync.md, with the
SESSION-024 amendments): the whole of cloud/directory_sync.py — the
note-section parser (the drawer's grammar: the banner/footer/stray
strips, the loose headings, the sign-up chip, the banish-tag filter),
the config trio (ensure_system_files' create-once law + the embedded
shipped synonyms at assets/directorysync), build_index (THE VAULT
MIRROR LAW — the published tree IS the folder tree: ghosts drop,
undeclared folders publish prettified, the case-insensitive folder-name
translation, the folded private belt, the banish wall, the
phantom-category tuples), the publish leg (Bearer + 1010 UA +
json.dumps(ensure_ascii=False) wire bytes) + the read-back door's
four truths, the admin command cycle (fetch → the idempotent
failure-isolated vault appliers [the whole-body edit, the move
collision suffix, trash-not-delete, the banish upsert] → mirror →
publish → ack — the golden-rule-5 order; the down-Worker law), the
standalone CLI (incl. the hand-rolled PBKDF2 owner-setup helper), and
THE DRIFT SEAL'S PUBLISH DOOR WIRED
(vaultseal.DriftSealDeps.PublishAfterBatch →
directorysync.PublishAfterBatch — the honest not-wired warning
retires) — with the ported test_directorysync selections (the
reference's whole suite) + the read-back four truths + the loopback
stub battery (the publish-before-ack order proof) + the local-bare
mirror push + THE GOLDEN (22 sections, paths/clock/git-dates frozen —
the byte-exact publish body 10,514 bytes + the deterministic mirror
commit sha c28cc921; dump_directorysync.py imports the reference's own
fixtures). 1736 tests race-clean.
**T-50 `internal/mirror` (SESSION-025) — THE
LIBRARY MIRROR + THE LINKING LAYER COMPLETE** (spec-core-mirror.md,
with the SESSION-025 amendments + the reality-wins test_noteredo
mapping fix): the whole merged group — mirror.py + mirror_keys.py +
linking.py + embeddings.py + recall.py — the one-way Library/ mirror
(the containment guards: SafeJoin's every traversal defense, the
both-direction overlap refusal, nothing-outside-Library; the
RAW-source move matching; the dry-run default; idempotence), the
recall hooks (the delimited block's insert/replace/idempotent/CRLF
surgery, the sanitizers with the 220-rune cut, the field extractors,
the plan/run pipeline), the embeddings surface (the short embed text,
the Ollama/OpenAI-compatible endpoints over the loopback stub, the
SQLite store, the stale-only refresh), and the linking layer (the
candidate pairs, the LLM confirm, THE SUGGESTION LIFECYCLE —
pending→approved/rejected, never re-suggested, capped at 7 — the
Suggestions note with its checkbox/strike grammar, the Related (auto)
blocks in mirror copies only, the Phase-5↔6 carry-over) — with the
ported test_phase5 + test_phase6 core selections (79) + THE GOLDEN
(39 sections, generated from the unmodified reference's own fixtures
via importlib, run twice byte-identical — the exact mirror-note
bytes, the refusal messages, the full mirror flow, the resync
integration). 1826 tests race-clean at T-50's landing.
**THE PHASE-4 CLOSER (SESSION-026) — THE EXIT GATE'S LAST ARTIFACT,
THE 30-LINK OFFLINE GOLDEN RUNNER EQUIVALENT — LANDED GREEN**: the
generator `testdiff/goldens/dump_goldenwebsites.py` drives the
UNMODIFIED reference's own `tools/run_golden_websites.py --offline`
(CI's mode) over its own `tests/golden/websites.json` corpus (30
links; the fake LLM answering from the golden file; canned pages
feeding the fetch; a throwaway vault; clocks frozen; paths folded by
REPLACE) into `internal/websitepipeline/testdata/
goldenwebsites.golden.json` — and `internal/websitepipeline/
goldenrunner_test.go` replays the SAME corpus through the ported
pipeline with the same seams, byte-comparing the canned pages, the
90-call LLM choreography + prompts, the per-link results, the 30
produced notes, the report, the console, and the exit-code law (all
green first run; the mutation test bites; the generator's two-run
byte-identical proof is part of every regeneration). **Phase 4's exit
gate is CLOSED — all 10 named differential stones landed (the stone
table: SESSION-025.md §gate; the closer's stamp:
12-migration-roadmap.md §Phase 4). 1828 tests race-clean.**
**Phase 5 OPENED (SESSION-027) — T-60's first slice: THE CLI SHELL landed**:
`internal/cli` grew from the Phase-0 seed into the honest Option-B surface —
the paint box (the C ANSI system + NO_COLOR/TTY/--no-color, level styles,
term_width, mask, rule, cli_print over the frozen clock, banner, render_bar,
StatusLine, strip-ANSI, the Windows 65001 hardening), the settings halves
(config load/save over the T-12 port, _ask, the token mask, the save
cleaner), THE SYNC DOOR's pure halves (the failure card,
sync_missing_essentials, the Enter-button with the cron law) + beat 1a
wired, the full flag surface + dispatch + refusal cards for the Telethon
accepted losses (ADR-0004 Am.3) + in-development cards naming their landing
tasks, the FULL --status card (16 config rows + taxonomy + the sentinel
rows + the Ollama/llama.cpp checks + the CacheDB/NoteStateDB/
WebsiteStateDB stats), and --list-dead/--reset-dead. The differential
golden: dump_cli.py drives the reference's own CLI shapes (the gui.app door
stubbed to the REAL CacheDB, net_guard stubbed, a fake ollama module, APP_DIR
patched, the clock frozen) → goldencli.golden.json (two-run byte-identical)
→ the Go replay green on the first run; the mutation test bites.
**SESSION-028 — T-60's remainder: THE COMMAND WIRINGS landed**: the `--init`
wizard (the full _ask sequence + the repo radar over the RadarAll seam +
the provider branches with the probe seam + the honest Option-B closing
card), `--test-connection` (the battery over the ported connectioncheck
with the _section/_check_line renderers — the Worker leg IS the queue
check; every Option-B divergence asserted), `--detect-llm` (the probe job
+ the numbered menu + set-and-save) with cli_settings's model-menu half
(SmartModelPick/PromptModelMenu/ApplyModelChoice), `--hand-delivery`
complete (the walled listing + the enqueue + the folder card), and
`--chrome-retry`'s listing + full dry-run rehearsal (the live leg waits
for T-62's batch engine — an honest boundary card). The golden grew to 9
more sections (the wizard runs with their SAVED CONFIG bytes, the battery
with the stubbed net doors, the detect cards, the model-menu matrix over
the REAL mixin picker, the hand-delivery writes, the chrome-retry card) —
the two-run byte-identical proof + the mutation teeth stay part of every
regeneration; a reality-wins find fixed in-commit (the reference's
chrome-retry --dry-run forgot to flip the dry-run switch — the port flips
it). 1866 tests race-clean over 31 packages (+14).
**SESSION-030 — T-62 slice 2: THE ENGINE + THE DRIVERS landed — T-62
COMPLETE**: run_batch_visual's synchronous shell (the setup family +
the warmup family + the websites phase over the ported Pipeline + the
GitHub loop + the Finish chain + the seals in the reference's exact
order + THE PHASE-5 CLEAR + the dry-run path + the SIGINT drain + the
100 ms heartbeat ticker driving StatusLine.Tick (T-61's spinner under
a real loop) + the exit-code law), the Worker-queue fetch +
the _bot_queue_job classification half (its own coherent piece), the
batch drivers (cmd_auto with THE WRONG-VAULT GUARD + the >10 confirm,
cmd_sync's live beats 1b/2/3, cmd_retry_failed, the import-file mode,
the chrome-retry LIVE leg), and `internal/platformintake` (the
per-platform _inbox tables). The differential golden drove the
reference's own engine over a fake Qt event loop (the signals as
callback lists, the worker inline — the synchronous Go shell's twin):
the 29-line log sequence + the BATCH DONE card + the exit code + the
12-entry dry-run report byte-compared — 86,570 bytes, two-run
byte-identical; eight reality-wins finds fixed in-commit (the
tracker's processed level, the §4.4 db-path law, the hygiene
github-vault-only law, the "cache.db" default-path law, the is_running
Start law, the model-changed sync, the dry-run's empty-cache creation
quirk, the case-sensitive tag-insert quirk). **1911 tests race-clean
over 32 packages (+34).**
**SESSION-029 — T-62 slice 1: THE RUN ENGINE'S DOORS landed**: the
preflights (the PyGithub token check over the PreflightGithub seam, the
TCP proxy reach over the NetDial seam, the provider branches with THE
model-picker bridge wired — the not-pulled Ollama model opens the
landed PromptModelMenu and persists through ApplyModelChoice; the
GitHub-error continue-anyway door with the cron law), run_config_card
(the boxed snapshot with the raw-comparison laws), the dry-run report
writer (writeDryrunReport over the ported RenderReportMarkdown, the
new dryrun.NowFn frozen-clock seam), and **THE PHASE-5 CLEAR RE-SPEC —
the bounded open item of ADR-0004 Am.3 — RESOLVED ON EVIDENCE and
implemented**: the all-clear gate survives as the honest-completion law
(the ✓/⏸️ verdicts), the queue-consumption mechanics are per-link
by LEDGER STATE, and the mirror push (the ack, per spec-worker-contract
§8.4) rides the batch-complete moment — PhaseFiveClear +
PushVaultMirrorAfterBatch over the new batch.VaultIndex.MirrorIndex()
export, the decommission events from the cache, and
workerclient.Sync.PushVaultMirror. The golden grew to 10 more sections
(the cards, the preflight cases incl. two SAVED CONFIGS, the checks
composition + abort, the dry-run report) — 69,028 bytes, two-run
byte-identical; the fixed-length-roots law (the card's padding bakes
path lengths); two generator bugs caught by the differential itself
(the shallow-copy leak, the ms-clock arithmetic); the mutation teeth
grew. 1877 tests race-clean over 31 packages (+11).
**SESSION-031 — T-61 COMPLETE (the close-out audit) + T-63 COMPLETE
(THE LAUNCHERS)**: T-61's row closed with evidence (the paint box +
the NO_COLOR/TTY law + the spinner under the engine's heartbeat + the
65001 hardening all landed with tests + GOOS=windows build/vet clean;
the runtime Windows-console legs recorded as Phase 7's clean-VM items;
ADR-0007 Am.2 records the owner's re-instatement of the CLI-facing
legs). T-63: `cmd/gitcurator-cli` (the --cli strip law ported from
main.py's routing, the --version stamp door per doc 10 §5, the exit
pass-through) + VERSION (0.92.0-go.1) + THE LAUNCHER SET in
cmd/gitcurator-cli/ (GitCurator-CLI/SYNC/Setup/Start .bats +
gitcurator-cli.sh + 1-INSTALL.bat — ASCII+CRLF asserted at byte level,
the walk-up with landmark VERSION, the exe run DIRECT with the
reference's own --cli defaults, EXITCODE echo + pause, the git-only
install check with NO registry writes and NO PATH mutations) + THE
DISPATCH FLIP (the reality-wins find: the T-62 drivers had landed but
Run's dispatch still answered not-yet cards for
--auto/--sync/--retry-failed/--import-file — now flipped to the real
drivers, the stale cards retired). The differential stone grew the
launchers section — the reference's OWN launcher bytes recorded raw,
the shared grammar byte-compared, every divergence asserted both ways
(the python ladder gone, main.py → VERSION, the exe direct), mutation
teeth on three surfaces — 101,777 bytes, two-run byte-identical.
**1922 tests race-clean over 33 packages (+11).**
**SESSION-032 — T-64 COMPLETE: THE DIFFERENTIAL CORPUS FULL RUN**: all
31 golden generators under testdiff/goldens/ driven twice end-to-end
against the unmodified reference @ 98f4fc7 by the new corpus driver
(testdiff/run_corpus.py — per-generator health + THE TWO-RUN
BYTE-IDENTICAL PROOF + COMMITTED == REGENERATED, with the four atomic
pairs' own law: PAIR TIE + normalized-content stability + the
committed pairs stay canonical). The fix loop's five finds were ALL
generator-side determinism bugs — three generators recorded raw mkdtemp
roots (githubpipeline/websitepipeline/directorynote — folded to <TMP>,
the replays' own fold regexes make it a no-op), dump_mirror's duplicate
fixture had two stems (the os.walk-order quirk flipped the recorded
winner between machines — one stem now, the Go replay fixture
mirrored), and the pair-stability digest learned the notereads
mtime_ns/epoch/scratch masks + the websitestate dismissed-ledger tie
order (unspecified on BOTH sides — the reference purges a Python SET
under a tied seconds-precision dismissed_at; the Go port faithfully
mirrors with a map). ZERO Go-port divergences; 1922 tests race-clean
over 33 packages; the harness case PASS; the report archived at
docs/poc-artifacts/t-64-corpus-full-run-report.md.
**SESSION-032 part 2 — T-65 COMPLETE: THE CLI PARITY RELEASE
`v0.92.0-go.1`**: the zip-composition question resolved per the specs
(the minimal deterministic mkzip slice — build/mkzip with doc 10 §2's
laws + 5 law tests; T-80 keeps the full builder per F4) + the GUI-side
launcher pair at the repo root (GitCurator.bat with the honest
GUI-milestone card; GitCurator-DRY-RUN.bat — the safe first run via
the CLI exe; the golden's launchers section grown to the reference's
own app-root pair, two-run byte-identical, a fourth mutation surface)
+ config.example.json + the Go WINDOWS-QUICKSTART.md. The release: the
exe double-build identical (760942e4…), the zip double-build identical
— **GitCurator-v0.92.0-go.1-windows.zip, 12 files, 15,988,399 bytes,
sha256 d2679a1de9d600e11d2e590016ed5359ca3fe0f841802289506f4cdd43fec
ecebb**, tag `v0.92.0-go.1`, GitHub release 416838956 with the zip
asset (round-trip verified). **1927 tests race-clean over 34 packages
(+5). PHASE 5 CLOSED — the books are empty; Phase 6 (the GUI shell)
opens next.**
**SESSION-033 — PHASE 6'S ENTRY GATE GREEN: T-70 COMPLETE (the PoC C
re-run with ONE REAL JOB)**: the recovery first (the twenty-seventh full
wipe + a mid-session degradation — the toolchain, module cache, and
webkit tree wiped twice; everything rebuilt from the committed recipes;
a latent test-side TOCTOU in the webfetch fake SOCKS5 server fixed
in-commit — 1927/1927 verified at the pushed HEAD before work). Then
the Wails toolchain record surfaced BEFORE the first wails invocation
(ADR-0002 Am.3: the pinned CLI v2.16.0, the NESTED webui/ module —
the main module's runtime set stays the authorized four — the vanilla
frontend step, the CGO story; + Am.4: the internal-boundary claim
disproved by probe, reality wins, fixed in-commit). THE SHELL: webui/
with the themed window (the spec-gui-theme-icons LIGHT tokens), THE
COALESCED LOG full contract (the deskbus Go half + the cache-mirror
frontend half — cap 1000, filter from cache, per-row tooltips, batched
fragment appends), and ONE REAL JOB wired through the guibridge/ seam
(the note-state baseline — the real ScanVault walk + fingerprints +
SQLite upserts over a deterministic 800-note fixture vault; 5 main-
battery contract tests). THE MEASURED VERDICT over two runs: first
frame 266/679 ms (gate ≤ 1.5 s), the 10k flood 458,695/454,133 lines/s
with 0 drops and 23–25 ms max handler time per batch, the real job 800
notes in 94–101 ms, the Windows exe CGO-free at 14.95 MB; the desk's
cap law VLM-verified off the screenshot. Evidence:
docs/poc-artifacts/t-70-gate.md (+ two results JSONs + the screenshot);
the verdict: ADR-0001 Am.3. **1932 tests race-clean over 35 packages
(+5 guibridge) + 12 webui tests (deskbus + fixture, from the nested
module).**

**SESSION-034 — T-71 COMPLETE: `internal/app` IS REAL (the composition
layer — doc 04 §2/§5's Qt-replacement seam)**: the recovery first (the
twenty-ninth full wipe BEFORE work — both repos, the toolchain, the
venv, and the module cache all gone; everything rebuilt from the
committed recipes: go1.27.2 sha-verified, both repos re-cloned at the
pinned revs with the ephemeral GIT_CONFIG_COUNT form, the venv, the
wails CLI v2.16.0; the rootless webkit tree DEFERRED — T-71 is pure
Go, the committed recipe rebuilds it when T-72's build legs need it).
Then T-71 in four slices, each committed+pushed: **(a)** events.go —
the typed fan-out Bus (the 8 queued Qt signals' transport) + the three
ask-gates with the reference's IDENTICAL timeouts (300/300/600 s), the
headless + NO-READER fail-fasts (the hang-bomb law's Go twin); **(b)**
commands.go + lock.go — the command surface (Sync.FetchQueue/RunBatch/
FastSync/DriftSeal, Scan.Run/ApplyPlan, Backup.SealVault/Resync/Now/
Publish, Radar.Detect) over injected seams, every command crash-safe
with BatchFinished exactly once, + the OpLock (the TelegramLockManager
port) and the one-batch law; **(c)** lifecycle.go — the single-instance
app.lock with PID-liveness healing, the watchdog, the startup passes
at the reference's exact delays, the shutdown graces; **(d)** app.go +
queue.go + scanjob.go + backupjob.go — the DI root wiring every
production seam (the queue classification with the settlement at the
door, the scan job with BOTH confirm doors — the bus gate + the Worker
scan trio, the backup zip's exclude/FIFO laws, the batch composition
with LLMDecision→the 600 s gate). **The acceptance (the command
contract tests + the lock/watchdog tests) GREEN: 67 internal/app tests
race-clean; the full battery at the close.** doc 04 §2 gained the
landed-shape note (the deltas the port's own laws forced — recorded,
not improvised). T-72 (the frontend foundation: the theme-token CSS
golden + the production log desk) is the next session's primary task.

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
