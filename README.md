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

**SESSION-035 — T-72 COMPLETE: THE FRONTEND FOUNDATION (the
theme-token golden + the production log desk + the production shell
binding internal/app directly)**: the recovery first (the THIRTIETH
full wipe BEFORE work — everything rebuilt from the committed recipes
again: both repos re-cloned at the pinned revs, go1.27.2 sha-verified,
the venv, the wails CLI, and THE ROOTLESS WEBKIT TREE rebuilt + the
dev-symlink relinking for the fresh base image, PROVEN at both levels
— the CGO build + the headless gate re-run GREEN). The suite
re-verification then caught a REAL production race: deskbus's
post-Stop push was a ~50/50 select coin flip (a below-cap send and a
closed stop both ready) — a line could be silently swallowed after
Stop; fixed deterministically (the stop-check before the send, commit
`cc89cb0`; the test untouched, `-count=50 -race` green), plus hygiene
(52 wipe-era mode flips restored to 644, the orphan 5.7 MB wails
binary removed — `2502e7a`). Then T-72 in two slices, each
committed+pushed: **(a) `88e5c6b` — the theme-token CSS golden + the
icon-pack golden**: dump_theme.py + dump_icons.py (the reference-driven
generators; harness-side PyQt6 stubs — the venv is stdlib-only;
registered in the corpus, 33 generators) → TWO goldens locking THREE
twins: the Go tables (webui/uitheme — 76 LIGHT + 76 DARK tokens + the
28-glyph lucide-static pack), the CSS var dump (frontend/theme.css,
the FULL tables incl. the `[data-theme="dark"]` plum night), and the
JS pack (frontend/icons.js, currentColor tint, the reference's
_LOG_LEVEL_ICONS law) — a struct-marshal-order trap caught on the
first run (maps sort keys; structs emit declaration order); the spec's
"30 glyphs" corrected to the real 28; **(b) the production shell**:
webui/shell.go binds internal/app DIRECTLY (guibridge/ RETIRED,
deleted with its tests; the T-70 gate legs retired with it), the typed
Bus → the shell's one subscriber → the deskbus (120 ms/≤250) → the
cache-mirror desk, the startup passes speaking their real lines through
the same transport, the single-instance lock before the window
(AppLock.Refused added — the alive-PID refusal exits 1, every other
failure CONTINUES WITHOUT), the sync_card CTA band (the 56px SYNC
hero, honestly disabled until T-77's hero machine). The DISSOLUTION
probe ran FIRST per Am.4's law: importing wails into the MAIN module
forces x/net v0.56.0 + x/text v0.39.0 + the 1.25.0 directive —
BLOCKED by the authorized-pins law; fully reverted, the boundary stays
nested, the owner's call recorded in **ADR-0002 Am.5** (the direct
binding rides Am.4's probe; the nested go.mod gained the transitive
requires as indirect, the MAIN go.mod byte-unchanged). **The acceptance
GREEN: the token golden + the log perf test** — the headless transport
test (10k lines through the REAL internal/app Bus at ≥1.07M lines/s,
0 drops, every batch ≤250) + the end-to-end gate leg (first frame
283 ms ≤ 1.5 s; the 10k flood through the FULL production chain at
434,782 lines/s producer-side / 173 ms drain, 0 drops, 41 batches,
22 ms max handler — the main thread free ~93%, all 10,001 rows
rendered). **1994 main-battery tests race-clean over 35 packages
(1999 − 5 guibridge, arithmetic closed) + 21 webui tests over 4
packages (12 + 6 uitheme + 3 shell).** T-73 (Settings pages A) opened
next.
**SESSION-036 — T-73 COMPLETE: THE SETTINGS PAGES A (the six pages live
on the production shell)**: the thirty-first full wipe
absorbed before work (everything rebuilt from the committed recipes,
the full battery re-verified green pre-work), then two slices, each
committed+pushed: **(a) `595a362` — internal/app/settings.go, the
settings command surface**: THE LIVE-CONFIG LAW (config.json loads ONCE
at construction through the GUI load path — the fresh detection, the
example base, THE FILL, the credential strip heal — and the app holds
the SAME map for the session, the reference's self.config; saves mutate
it IN PLACE, never a rebind; the missing-file case loads the example
base like the reference's deepcopy(CONFIG_EXAMPLE), the old wiring
assertion corrected to pin it) + the six pages' engines over injectable
SettingsDeps (the hermetic law): the ONE merge-based save with ALL FOUR
v30 laws (MERGE never whitelist-rebuilds — unknown/tuning keys survive;
a page group carries only its keys, an absent group keeps config — the
hasattr pattern as the group-pointer law; the LIVE map mutates in
place, observed by a running-worker fake through the SAME instance; the
atomic write + the credentials mirror BEFORE the mutation; the 250 ms
sentinel re-arm; log_level keeps the hardcoded-INFO quirk), the
five-rung vault status ladder, the snapshot (populate_vaults' dedupe +
the no-pre-selection law + the blocked/self fields showing the
LAW+extras union), TestGitHubToken as-typed (the 401/403/network
lines verbatim), the About Me wizard's pure template + the CWD write,
RefreshOllamaModels (the v30 in-place ollama.model sync), StartOllamaServer
(the detached spawn twins on both platforms), llama.cpp Detect/Refresh +
THE SERVER CONFIG GRAB (the /props limits ride every probe), the Detect
& Set fast lane (the batch guard, the live-URL snapshot law, the
single-model auto-apply, the several-models menu door, the honest
cancel line, the SET law with the live-fields overlay), PreviewSound
(the exact batch gate's decision half + BatchSound's one-notice
missing-asset law) — **26 settings tests incl. the save-law battery**;
**(b) `48157da` — the Settings window on the production shell**:
webui/settingsui.go (the thin-shell binding wrapper + the two native
pickers with the reference's dialog titles), frontend/settings.js +
index.html + style.css (the gear door with the reference's tooltip
verbatim, the 880×640 window, the 190px sidebar — six rows in the
reference's order: Credentials, Proxy, Vault, LLM, Input, Sound — the
icon pack's section glyphs with the two-mode tint, the TAB_INFO hints;
the six pages rendered from the snapshot with every label/tooltip/
help text verbatim; THE SAVE-ON-CHANGE law editingFinished/toggled →
the page savers — the commit-cadence dated note in the spec: the port
commits all six pages per edit where the reference saves creds/proxy/
LLM only at closeEvent, every reference-persisted value persists, some
a commit earlier; the mini-dialog kit for the remove-confirm/wizard/
model-menu), the v0.69 fill line at startup, the SB_POC_OPEN_SETTINGS
evidence probe. **The acceptance GREEN: the page workflow tests + the
save-law (merge/in-place) tests** — 35 main packages + 4 webui
packages race-clean; the gate leg re-run GREEN (first frame 963 ms ≤
1.5 s; the 10k flood 526,315 lines/s, 0 drops); the screenshot
VLM-verified: all six sidebar rows (Credentials selected on the accent
fill), the Credentials page's fields + buttons, the example-base values
(API ID "0" — the str() render parity), the user-config quiet-truth
line VERBATIM. **The CLI-parity shell wiring (warmup family, inputs
fetchers, websites phase, Phase-5 CLEAR) carried honestly to the T-74+
binding** — the one-slice-properly law; it gates T-77's hero batch,
not the pages.
**SESSION-037 — T-74 COMPLETE: THE SETTINGS PAGES B + THE CLI-PARITY
SHELL WIRING (the carried debt closed)**: the thirty-second full wipe
absorbed before work (both repos re-cloned with the ephemeral
GIT_CONFIG_COUNT PAT form, go1.27.2 sha-triple-verified, the venv, the
wails CLI, and the rootless webkit tree rebuilt — the sandbox began
blocking symlink(2) mid-session, so the 26 dev .so links landed as
hardlinks to the resolved system runtime targets: same bytes, the
wails build verifying the tree), then three slices, each
committed+pushed: **(a) `a4619fe` — internal/app/batchwire.go, THE
CLI-PARITY SHELL WIRING**: engine.go's runImpl composition re-hosted
on the app surface — the warmup family (the Ollama list/autostart/
adaptation ladder, the llama.cpp re-probe/detect, the cloud /models
pre-flight), the inputs fetchers (the import file + the _inbox
tables), the full intake + banishment + settled ledger + backfill +
the §4.4 passes, THE WEBSITES PHASE (the full websitepipeline
composition), THE FINISH CHAIN (the seals in the reference's order +
the Phase-5 CLEAR's mirror-push ack + the all-clear verdict + the
dry-run law) — the debt T-73 recorded, closed with 19 wiring tests
over the App's six new seams; **(b) `6d479fb` — internal/app/
settingsb.go, the pages' engines** (the specs matched BY NAME per the
R-0 note): the QUEUE REPORT RENDERING from canned reports (every
labeled-count line verbatim + the settled lines + the websites
sub-table + the advisories + the 50-cap + the caught-up family) + the
zinc ⏳/mint ✅ badge law + the dashboard engines (the frontmatter
scan + the exact report text + the quarantine table/reset + the undo
engine with the top-10 confirm + the processed-repos row removal) +
the sources fetch over its seam + the backup ladders + the save
groups — 21 tests; **(c) `3091d76` — the shell**: webui/settingsuib.go
(the third bound struct) + frontend/settings.js's four pages at the
reference's own sidebar order (Dashboard after Input, Bot, Sources,
Backup — every label/tooltip/'?' help verbatim; the Restore button
honestly disabled with the landing note) + the section-probe evidence
door — 5 shell tests. **2060 main-battery tests race-clean over 35
packages (2020 + 19 + 21) + 28 webui tests over 4 (23 + 5); the gate
leg GREEN (first frame 283/942 ms; the 10k flood 476,190 lines/s,
0 drops); the screenshots VLM-verified (all TEN sidebar rows in the
reference's order; the Backup page's four groups with every button/
checkbox/placeholder/status dot).** T-75 (the Sync view: flow cards +
matrix + differences modal) opens next.

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

**SESSION-038 — T-75 COMPLETE: THE SYNC VIEW (the flow cards + the matrix
+ the differences modal), the placeholder seat filled**: the
thirty-third full wipe absorbed before work (both repos re-cloned with
the ephemeral GIT_CONFIG_COUNT PAT form, go1.27.2 sha-verified against
the committed 64-char pin, the venv, the wails CLI, and the rootless
webkit tree rebuilt — the 6th webkit rebuild, Am.4's hardlink fallback
again, symlink(2) still blocked). Three slices, each committed+pushed:
**(a) `6eea74d` — internal/syncreview, core/sync_review.py PORTED** (the
placeholder seat T-71 recorded, finally filled): the four-state truth
table, the banner verdict with v0.91.0's named verdict + the
setup/measure split, the micro-copy captions (the owner's own 720/710
example), the relative clock, the chip grammar, the DeferredLog over
the SQLITE LAW (never a delete), the three read-only doors (the ONE
network door with the 1010 UA + the hermetic seam, the vault site walk
over build_index's own filters, the read-only git backup gap) — 101
tests. **(b) `04768e7` — internal/app's syncdisplay.go + syncview.go**:
THE VISUAL SYNC DISPLAY's pure model (the four cards in the owner's
flow order with every line verbatim, the §5 arithmetic's every edge —
shrink/never-published/unreachable/not-configured — the calm line law,
total junk tolerance) + the view half (the six-stage gather over
injectable walks with the queue stages riding the Bot page's OWN
classified check, the recompute-from-ledger, the code-aware published
door with the Worker's 404 = never_published, the Settings doors with
ONE Refresh payload for both desks — the shared single gather the
reference defers to v2, taken now) — 50 tests. **(c) `478101c` — the
shell**: the fourth bound struct + the Sync page at the reference's
own sidebar order (between Dashboard and Bot, the 'refresh' glyph):
the 2×2 flow grid, the banner + the four-state matrix, the differences
modal (the chips + the disabled-until law + the ONE inline
confirmation + the log tab with Reopen), the first-show door, the fix
doors — 5 shell tests. **The acceptance GREEN (the four-state
rendering tests + the deferred-ledger UX tests): 2211 main-battery
tests race-clean over 36 packages (2060 + 101 + 50) + 33 webui tests
over 4 (28 + 5); the gate leg GREEN over the rebuilt tree (first
frame 292/730 ms ≤ 1.5 s; the 10k flood 588,235 lines/s, 0 drops);
the screenshots VLM-verified** (the four cards' honest not-set grammar
— queue not measured / no vault connected / no backup repos / vault
not set; "measured at 18:49:32 — the numbers are as fresh as this
line"; the dashed unknown matrix; the desk living underneath). The
MAIN go.mod byte-unchanged all session.

**SESSION-039 — T-76 COMPLETE: THE SCAN FLOW (the vault picker → the
plan modal → the applied scorecard)**: the thirty-fourth full wipe
absorbed before work (both repos re-cloned with the ephemeral
GIT_CONFIG_COUNT PAT form, go1.27.2 sha-verified against BOTH the
committed 64-char pin AND go.dev's published JSON, the venv, the wails
CLI, and the rootless webkit tree rebuilt — the 7th webkit rebuild,
Am.4's hardlink fallback again: symlink(2) blocked at boot, the 26
dev links as hardlinks to the resolved system targets; the FULL
battery re-verified green pre-work, exactly the recorded 2211 + 33).
Two slices: **(a) `internal/app/scandialogs.go` + the scanjob story
extension — the three dialogs' pure engines**: the ScanStory (the
reference's per-vault story dict: the planned counts spoken in PICKED
units once a confirmed selection rides, the landed counts, the failed
moves BY NAME; ScanResult grows the out-dict shape — Story/Repos/
VaultsRan/Hygiene + the vault seats; the GUI door's ask line
byte-for-byte) + PickDisplayModelOf (the status lines in the owner's
voice, the hint law) + PickScanButtonLabel (the live arithmetic) +
PlanDisplayModelOf (the four sections with the reference's own washes,
every row's stable id + one-line grammar + full tooltip story — the
deletion door words, the one-note-per-site grammar, THE LANDABLE
MOVE's "(as …)", the inventory block, the MAX_ROWS 500 freeze guard)
+ **PlanPickState** (the selection semantics as a pure state machine:
the low-confidence pre-uncheck, the master sweep, the three-state
master mirror, the live Apply label, THE ARM-GUARD'S ENABLE LAW, the
four-key Selection) + ResultDisplayModelOf (the LANDED counts, the
'K of N' warning tones, the named failures cap-3, the three-way
heading law, the per-vault sections; **the reference's repos-only
double-section quirk preserved — verified by RUNNING the reference's
result_display_model at 98f4fc7**) — 19 tests. **(b) the shell**:
webui/scanui.go the FIFTH bound struct (StartScan THE CTA's whole
flow — the guards verbatim, the quiet cancel, the start line, the
blocking job over the bus gate, the finished routing + the scorecard
event; ScanAnswer the held gate's answer door; the two evidence
doors) + shell.go's AskScanConfirm wiring (the scan:ask event with
the answering window; the bus's 300 s window the belt) +
frontend/scanflow.js's three modals (the picker nothing-pre-checked
with the live-arithmetic button; THE PLAN MODAL with the 350 ms
fresh-click door, the masters with the indeterminate partial, the
char-budget elideMiddle twin, the m:ss countdown, the four-key
selection riding ONLY confirmed; THE SCORECARD with the per-vault
sections, the 12 s self-close) + the Scan CTA's own machine (the
reference's scan_btn tooltip verbatim — ui.py wires it in
vault_scan_ui.py, NOT the hero machine; the SYNC/Fast/Test buttons
stay T-77's) — 7 shell tests. **The acceptance GREEN (workflow tests
+ selection semantics): 2230 main-battery tests race-clean over 36
packages (2211 + 19) + 40 webui tests over 4 (33 + 7); the wails
build links clean over the 7th webkit rebuild; the gate leg GREEN
(first frame 339/283/280 ms ≤ 1.5 s; the 10k flood 0 drops); the
screenshots VLM-verified** (the plan modal: the four wash sections +
the masters + the LOW-CONFIDENCE ROW PRE-UNCHECKED + the landable
"(as C_v1.md)" row + the pinned verdict row + the ticking window;
the picker + the scorecard legs). The MAIN go.mod byte-unchanged all
session.

**SESSION-040 — T-77 COMPLETE: THE MAIN WINDOW CHROME (the hero button
machine, the progress row, the More menu, the sentinel labels, the
battery modal, the sounds)**: the thirty-fifth full wipe absorbed
before work (the repos re-cloned with the ephemeral GIT_CONFIG_COUNT
PAT form, go1.27.2 sha-verified against both the committed pin and
go.dev's published JSON, the venv, the wails CLI, and the rootless
webkit tree rebuilt — the 8th webkit rebuild, Am.4's hardlink fallback
again; the FULL battery re-verified green pre-work: exactly 2230 +
40). Two slices: **(a) `9cd7586` — the chrome's pure engines**:
internal/app/herostate.go — HeroMachine, hero.py's GUI-only state
machine (the four states sync→fetching→process→running; the click
routing's guards as data incl. the offline advisory + the
nothing-to-sync split RE-WORDED for Option B's Worker queue; the fetch
routing with the both-pipelines arithmetic + the caught-up law (the
waiting-rows report, the drift seal, the closing-line choice); the
mirror's defensive law; the stopping render; Render's exact
labels/glyphs/fills/tooltips; ConfirmLargeBatch; the progress row's
live formats; the sentinel labels' paint laws) — 21 tests. **THE
RETROACTIVE FIELD-TAG FIX** — a REAL display bug family: the T-71 bus
payloads crossed tagless, so the progress bar NEVER filled during a
batch and every finish rendered ❌ Failed; every crossed payload
re-tagged + pinned by its own test; BatchFinished gains needs_retry.
**The cooperative StopBatch door** (the runner's between-items law +
the warmup lost-stop window closed + the arm-token handle seat).
**The sentinel pass's full emission** (the net+ai+proxy snapshot +
the offline verdict transitions + the Sentinel event) with TWO reality
fixes the re-read caught: the nil-latency deref (an offline pass
PANICKED into the crash-isolation) + the sentinelOnline DATA RACE (the
tick + the save's 250 ms re-arm — the seat now one unit under
sentinelMu). **The WAVs SHIPPED** (the deferred-debt note closes:
success.wav/retry.wav embedded, disk-first, the quiet batch gate split
from the preview's lines). **(b) `c14fb8f` — the shell**: the SIXTH
bound struct webui/heroui.go (HeroClick the one click door with both
stages + the 200 ms mirror + TestConnection the battery door with the
live events + the More menu's live doors incl. the retry modes +
ChromeBoot) + frontend/heroflow.js (the hero renders, the guard boxes,
the sentinel paint, THE BATTERY MODAL — the four rows + Start Syncing
gated on all-ok, THE MORE MENU — the reference's own action list
verbatim with the honest-disabled law, THE BATCH CHIME) — 16 shell
tests. **The acceptance GREEN (hero state machine tests + the startup
cadence battery): 2252 main-battery tests race-clean over 36 packages
(2230 + 22) + 56 webui over 4 (40 + 16), stable across three
consecutive full -race runs**; the gate leg GREEN (first frame 908 ms
≤ 1.5 s; the 10k flood 555,555 lines/s, 0 drops, 26 ms max handler);
the screenshot VLM-verified (the sentinels painting LIVE: Net: online
green / AI: off — unreachable red / Proxy: off gray; the SYNC hero
enabled; the honest placeholders retired). The MAIN go.mod
byte-unchanged all session.

**SESSION-041 — T-78 COMPLETE: THE SIDE-BY-SIDE VERIFICATION VS THE PYTHON
APP (the LAST Phase 6 task — the task list is now EMPTY).** The
thirty-sixth full wipe absorbed before work (both repos, the toolchain,
the venv, the module cache gone; everything re-cloned/rebuilt from the
committed recipes — the webkit tree deliberately NOT rebuilt, no wails
build was needed for the data-side legs; the FULL battery re-verified
green pre-work: exactly 2252 + 56). **The machine (committed):
`testdiff/side_by_side.py` + `testdiff/goside`** — 21/21 GREEN, the
evidence archived at `docs/poc-artifacts/t-78-side-by-side.md`. LEG A
(the acceptance — the lock mutual exclusion ACROSS BOTH BINARIES): the
Python half drives THE REFERENCE'S OWN LOCK CODE, extracted verbatim by
pinned line range (the PyQt6 import wall probed + pinned:
headless.py:34 → icons.py:45 → ModuleNotFoundError under the
stdlib-only venv; `--cli`/`--headless` both exit BEFORE the lock — the
GUI entry is the reference's only lock-taking binary) — the fresh-take
BYTES law (bare decimal PID) on both sides; A1 Go holds → the
reference's own dance REFUSES (exit 1, the exact sentence); A2 Python
holds → Go REFUSES; A3/A4 both crash-heal paths with the exact stale
sentences; A6 the EPERM probe — **A REAL DIVERGENCE THE AUDIT CAUGHT
AND FIXED** (the port read a live not-ours process (EPERM) as ALIVE;
the reference's blanket except reads it DEAD — a recycled-PID lock was
take-over-able in Python but refused by Go; `process_posix.go` fixed,
pinned hermetically, both binaries now take the PID-1 lock over
identically, LIVE). LEG B (the same-data laws, the reference's own
pure modules as the oracle): both apps write BYTE-IDENTICAL config.json
through their real doors (merge-never-rebuild, tuning keys
default-only, nested key-by-key, unknown keys preserved,
ensure_ascii=False + SetEscapeHTML(false) raw) + the cross re-save
stability (all four sha256 equal) + the fill (same bytes, same counts,
local-wins + the fresh door) + the mirror (identical .gitcurator bytes,
never-loses, 0o600). **The acceptance GREEN: lock test across both
binaries ✓ — 2253 main-battery tests race-clean over 36 packages
(2252 + 1) + 56 webui over 4**; the app package 3× clean post-change;
the MAIN go.mod byte-unchanged. The phase-exit question (opening
Phase 7 per the protocol's gate laws) is the owner's to decide —
surfaced in the session report.

**SESSION-042 — THE OWNER'S PATH-B STAMP: T-79 (the GUI doors
addendum) + T-79a COMPLETE (the five per-test doors).** The
phase-exit call answered in the owner's own words ("the slow, more
reliable and more rigorous method"): Phase 6 stays OPEN past T-78 —
the More menu's honestly-disabled tail (the 19 doors from
SESSION-040's record) became the NAMED rows T-79a…T-79e on doc 13
(the export zip rides T-80's packaging composition per the standing
assignment); Phase 7 opens only when the tail is empty. The
thirty-seventh full wipe absorbed before work (both repos, the
toolchain, the venv, the module cache, the wails CLI gone; the repair
all from the committed recipes — the venv as `--without-pip --copies`
(this boot lacks ensurepip; symlink(2) still blocked), the webkit
tree's NINTH rebuild with Am.4's hardlink fallback again; the FULL
battery re-verified green pre-work: exactly 2253 + 56). **T-79a: the
five per-test doors LIVE** (`internal/app/pertests.go` over five new
hermetic `SettingsDeps` seams + the `HeroUI` wrappers + the
`heroflow.js` wiring + the `SB_POC_OPEN_MORE` evidence probe): 🔑
Test Telegram & GitHub (the Option-B re-scope — the GitHub half
verbatim + the Worker queue leg in the Telegram half's seat), 🌐 Test
Proxy (the sentinel's ProxyReachable TCP verdict — the re-scope), 🧠
Test Ollama (models + the real "What is 2 + 3?" inference ping + the
options retry), 🔌 Test Cloud API (the /models pre-flight +
the "Say hello" ping + the three verdict boxes), 🦙 Test llama.cpp
(the /props probe + the model pick + the ping through the same
OpenAI-compatible path + the three verdict boxes). **The acceptance
GREEN: every door live ✓ — 2280 main-battery tests race-clean over 36
packages (2253 + 27) + 59 webui over 4 (56 + 3)**; the app package 3×
race-clean; the wails build links clean over the 9th-rebuilt webkit
tree (no mode flips; the regenerated bindings content-only); the gate
leg GREEN (first frame 779/284 ms ≤ 1.5 s; the 10k flood 0 drops);
the screenshots VLM-verified (the five doors LIVE-dark beside the
remaining tail's honest-disabled wash + the end-to-end click leg
speaking the reference's own honest Ollama-failure lines). The full
record: `docs/poc-artifacts/t-79a-per-test-doors.md`. **The remaining
tail: T-79b (the linking tools), T-79c (the bot-queue tool dialogs),
T-79d (the Chrome doors), T-79e (the decommission picker).**

**SESSION-043 — T-79b COMPLETE: THE FIVE LINKING-TOOLS DOORS LIVE.**
A two-wipe session: the thirty-eighth absorbed before work (both
repos, the toolchain, the venv, the module cache, the wails CLI, AND
the webkit tree — all rebuilt from the committed recipes, the TENTH
webkit rebuild; the FULL battery re-verified green pre-work: exactly
2280 + 59) and the thirty-ninth at the context resumption (the first
context died after the implementation + evidence but before any
commit; the continuation recipe executed verbatim — the repos
survived WITH the in-flight draft + tracked-file MODE FLIPS 644→755
(content byte-intact), repaired via chmod off the parsed
`git diff --raw` list, never git checkout; the toolchain rebuilt; the
draft AUDITED line-by-line against the reference's own bytes, never
blindly trusted; the one audit gap fixed — webui/README.md's map
entries). **T-79b: `internal/app/linkdoors.go` — the five doors over
THREE new hermetic `SettingsDeps` seams**
(`LinkOllamaChat`/`LinkCloudChat`/`LinkEmbedCall` — the production
wiring IS `mirror.BuildEmbeddingCall`) + the `HeroUI` wrappers + the
`heroflow.js` wiring + the recategorize table modal + the generalized
probe click leg: 🪝 Recall hooks (dry-run — the TestWorker law: the
busy guard → the "⏳ … running…" opening → the captured
stdout+stderr run → the LAST 40 LINES at completion → the ✅/❌
verdict; DRY-RUN pinned — recall-dryrun.md is the only write) + 🔗
Build link suggestions (the suggest grammar + THE ONE-WRITE law — the
Suggestions note under Library/, the machine vaults byte-untouched) +
✅ Validate Vault (connection_tests.py:219 verbatim) + 🔍 Verify Vault
(the report box + the enqueue legs with the settled-repo filter) +
📁 Recategorize Notes (the two-door flow: the open walk + the table
modal + the apply with the cache row moves). THE GUARD LAW widened
per the row's own naming: all five doors carry the batch-busy guard
(the recorded divergence — the reference guards only the two tool
runners). **The acceptance GREEN: the five doors live with their
safe-default laws ✓ — 2312 main-battery tests race-clean over 36
packages (2280 + 32) + 63 webui over 4 (59 + 4)**; the touched
packages 3× race-clean; gofmt/vet clean over both modules; the MAIN
go.mod byte-unchanged all session; the wails build links clean over
the 10th-rebuilt webkit tree (no mode flips; the regenerated bindings
content-only); the gate legs GREEN (first frame 726/269 ms; the 10k
flood 0 drops); the screenshots VLM-verified (the five doors LIVE +
the recall-door click leg's whole runner grammar on the desk — the
lower-region misreads settled per the 042 law). The full record:
`docs/poc-artifacts/t-79b-linking-doors.md` (the nine recorded
divergences inside). **The remaining tail: T-79c (the bot-queue tool
dialogs), T-79d (the Chrome doors), T-79e (the decommission
picker).**
