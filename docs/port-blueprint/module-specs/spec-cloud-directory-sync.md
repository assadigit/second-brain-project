# Module Spec — `cloud/directory_sync.py`

- **LOC**: 1,826 · **Go home**: `internal/directorysync` · **Status**: port
- **Responsibility**: the Website Directory publish leg + the Phase-2 admin-command cycle —
  the one-way content flow Obsidian → GitHub → Website.

## Index build (`build_index` 820-1008) — THE VAULT MIRROR LAW (v0.88.0)

- The published category tree **IS the vault's folder tree**: the `_system/directory-config.json`
  is a metadata registry only (ghost entries never publish; undeclared folders publish with
  prettified names); folder-name translation `safe_filename(name) ↔ display name` is
  case-insensitive; the private belt is folded
  (`prettify_folder_name(safe_filename(name)).lower()`).
- Filters: banished hosts (host-level, with the shared-host exception list github.io,
  notion.site, vercel.app, blogspot.com, wordpress.com, netlify.app, pages.dev, gitlab.io,
  webflow.io, medium.com, substack.com — URL-level bans only on shared hosts); private
  categories; invalid sites; dedup.
- Site record: `{id: sha256(normalized)[:16], slug, name, url, host, category, subcategory,
  summary, tags, featured, needs_review, added_at, updated_at, what_it_does, best_used_for,
  practical_details, watch_out, similar_tools, extra_sections, signup_required, pricing,
  detail_markdown}`.
- Index: `{schema: "gitcurator-directory/1", generated_at, counts, categories, sites,
  popular_tags (top 12), tag_aliases, synonyms, banished_hosts, new_days: 14}`.

## Note parser (`parse_note` 216-383)

Frontmatter + drawer-section parser (TL;DR / What it does / Best used for / Practical details /
Watch out / Similar tools + extra sections); `signup_required` chip; pricing enum
(free/freemium/paid/unknown); `featured`; `needs_review` (v0.39.0); banish-tags stripped.

## Config trio + belts

- `_system/{directory-config.json, banished.json, search-synonyms.json}` in the vault;
  `ensure_system_files` creates when missing, never overwrites; shipped fallbacks at
  `app/taxonomy/`. *(Reality-wins fix, SESSION-024: this spec previously named
  `banished-sites.md` and `search-synonyms.md` — the reference's own constants are
  `BANISHED_FILENAME = "banished.json"` and `SYNONYMS_FILENAME =
  "search-synonyms.json"`; the code wins, the spec rides in the same commit.)*
- `local_only_note_patterns` (478-495; the six tombstoned stems as globs) +
  `private_exclusion_patterns(vault)` (498-567; private-category folders **discovered on disk**
  derived from belt names (folded), the gitignore escape law, the vault's own
  `_system/directory-config.json` `local_only` list as the owner's extension point, never
  raising) — feeds **both** the publish filter and VaultSeal's second gitignore block.
- `is_banished(url, banished, shared_hosts)` (619-637).

## Publish + read-back

- `publish_index(index, worker_url, token, timeout=30)` (1031-1065): `POST
  {worker}/api/directory/sync`; body `json.dumps(index, ensure_ascii=False)` UTF-8; headers
  `Content-Type: application/json`, `Authorization: Bearer <token>`, 1010 UA.
- `fetch_published_index(config)` (1068-1135): `GET /api/directory/index` → four truths:
  `ok (sites count + generated_at)`, `never_published` (404), `not_configured`, `error`.
- `run_sync` / `publish_after_batch` (the end-of-batch hook = full command cycle).
- Env overrides `DIRECTORY_WORKER_URL/TOKEN/VAULT`; config `directory_worker_url` (wins over
  `cloudflare_worker_url`), `directory_sync_token`, `directory_poll_interval` 300 s.

## Admin command cycle (Phase 2 — golden rule 5 order)

`fetch_pending_commands` (`GET /api/commands?status=pending`, Bearer) → `apply_admin_commands`
(types `delete, banish, feature, unfeature, review, move, edit, unban` — each idempotent,
failure-isolated; vault appliers `_edit_note_fields`, `_move_note_file` (safe_filename folders
+ numeric suffix on collision), `_trash_note` (→ `<vault>/.trash/<stem>-<ts>.<ext>` — never a
permanent delete), `_banish_add/_banish_remove`) → `push_mirror_repo` (plain
`status→add -A→commit "admin commands applied: <ts>"→push` when the vault is a repo with a
remote) → publish → `ack_command` (`POST /api/commands/<urlquoted-id>/ack {status: applied|
failed, error ≤500}`) — **ack only after publish (no resurrection window)**; held commands
(10 s undo window) invisible server-side; a down Worker returns None (commands stay un-acked).

## CLI (`main` 1228-1292)

`python -m gitcurator.cloud.directory_sync` with `--vault`, `--config` (default `./config.json`),
`--sync-only`, `--publish`, `--print-index`, `--out FILE`, `--make-admin-hash` (PBKDF2 helper).

## Tests

Index golden (tree-from-disk law, ghost/undeclared cases, folded belt); parser field matrix;
publish payload shape (Bearer + non-ASCII body); read-back four truths; command cycle (order,
idempotent applies, ack-after-publish, held-window invisibility, trash-not-delete); trio
creation law; belts builders (port `test_localonly` + `test_directorysync` selections).

## Belt floor shipped early — T-47 (SESSION-022)

`internal/directorysync/belt.go` carries the v0.89 LOCAL-ONLY BELT's
pattern builders (local_only_note_patterns, private_exclusion_
patterns, load_directory_config + the pure helpers) because T-47's
vaultseal config bridges consume them — the T-45 repobanish-floor
precedent. The shipped directory-config default is EMBEDDED at
`assets/directorysync/website-directory-config.json` (sha256-pinned
by vaultseal's golden) where the reference reads app/taxonomy/ from
disk — the doc 04 §1 assets layout. T-49 completes the module
(build_index, publish, the command cycle, the config trio).

## SESSION-024 shipped record — T-49 complete (commit in the session log)

The REST of the module landed at T-49: `internal/directorysync` now
carries the whole of `cloud/directory_sync.py` —

- **noteparse.go** — `parse_note` (the drawer's grammar; the JS twin
  kept identical on purpose) with `_unquote`/`_section_key`/
  `_parse_bullets`/`_parse_paragraph`/`_parse_tags`;
  `ParsedNote.ToMap()` rebuilds the reference's exact dict shape.
- **systemfiles.go** — `ensure_system_files` (create-once,
  never-overwrite; the shipped synonyms EMBEDDED at
  `assets/directorysync/search-synonyms.json`, sha256-pinned by the
  golden — the same convergence as the T-47 config embed) +
  `load_synonyms`.
- **index.go** — `build_index`: THE VAULT MIRROR LAW (the tree IS the
  folder tree; ghosts drop; undeclared folders publish prettified),
  the case-insensitive folder-name translation (only a UNIQUE
  sanitized match translates), the FOLDED private belt, the banish
  wall, the popular tags, the (cat, sub) tuple pairs (the
  phantom-category law), the ordered index tree.
- **publish.go** — `publish_index` (POST, Bearer, 1010 UA,
  `json.dumps(index, ensure_ascii=False)` wire bytes),
  `fetch_published_index` (the four truths), `_effective_config`
  (env overrides), `run_sync`, `publish_after_batch`.
- **commands.go** — `fetch_pending_commands`/`ack_command` (the
  urlquoted-id ack), the idempotent vault appliers
  (`_edit_note_fields` incl. the whole-body edit + the verbatim
  `insert(1 if not done_name else 1)` quirk, `_move_note_file` with
  the numeric collision suffix, `_trash_note` → `.trash/` never a
  permanent delete, `_banish_add`'s upsert keyed host+url with the
  existing entries riding verbatim, `_banish_remove`),
  `apply_admin_commands` (idempotent + failure-isolated),
  `push_mirror_repo`, `run_command_cycle` (the golden-rule-5 order:
  apply → mirror → publish → ack; the down-Worker law).
- **cli.go** — `main` (`--vault/--config/--sync-only/--publish/
  --print-index/--out/--make-admin-hash`; the PBKDF2 helper
  hand-rolled — `hashlib.pbkdf2_hmac('sha256', …)` over stdlib
  crypto/hmac, zero new dependencies).
- **the drift seal's publish door WIRED** —
  `vaultseal.DefaultDriftSealDeps().PublishAfterBatch` binds
  `directorysync.PublishAfterBatch` (the honest not-wired warning
  retires; the (msg, level)→msg adapter tags plain lines "info").

**The golden** (22 sections,
`internal/directorysync/testdata/directorysync.golden.json`,
GENERATED by `testdiff/goldens/dump_directorysync.py` from the
unmodified reference @ 98f4fc7 — the generator imports the
reference's own test fixtures via importlib, zero transcription risk;
run twice, byte-identical): the parser matrix, the pure helpers
(slugify/section_key/prettify/norm_host/is_banished incl. the
lstrip('www.') character-set quirk/the banish wall), the trio's
creation law + the shipped bytes + sha256, build_index on the
all-laws fixture, **the publish body's EXACT wire bytes**
(`json.dumps(index, ensure_ascii=False)` — key order + separators +
raw UTF-8, 10,514 bytes), the walk's id map, the appliers' exact file
surgery (frozen trash stamps + banish stamps), the
`_frontmatter_set` matrix, `_effective_config`'s override matrix, the
CLI shapes, and **the mirror push's deterministic commit** (a local
bare remote, frozen clock + frozen git dates → the reference's exact
sha `c28cc921`).

**The batteries** (78 new top-level tests — 1736 total): the ported
selections of the reference's whole test_directorysync.py
(TestParser/TestBanish/TestSystemFiles/TestBuildIndex/TestPublish/
TestCli/TestCommandAppliers/TestBanishWall/TestCommandCycle — every
test function), the read-back door's four truths, the loopback stub
worker (the reference's _StubWorker/_StubWorker2 ports with the
ORDERED event log — the publish-before-ack proof), the local-bare
mirror-push battery, and the golden differential.

## Go notes (the convergences and laws found at port time)

- **`storage.PyDumpsUTF8`** (new, T-49): `json.dumps(v,
  ensure_ascii=False)` — the publish leg's wire shape (raw UTF-8, ", "
  / ": " separators). The latent `\b`/`\f` fidelity gap in the house
  `pyEncodeString` (Python emits the two-char escapes; the old code
  emitted `\u0008`/`\u000c`) was fixed in the same session — the full
  suite proves no golden ever depended on the quirk.
- **The `[]`-vs-`null` byte law**: a Go nil slice dumps as `null`
  (Python None); the reference's `sub_nodes = []` is an EMPTY LIST.
  Every list the index tree carries must be initialized `[]any{}` —
  the publish-body golden caught this live (the tree comparison
  alone was blind to it: norm maps compare equal).
- **`dictDoor` over the T-48 instance seams**: the command leg needs
  PRESENCE checks (`'key' in payload`) the belt floor's dictReader
  could not answer (absent vs nil); the new present-aware door rides
  commands.go. CAUTION: pass the RAW value into `editNoteFields` —
  double-wrapping a door is a silent no-op (found live, fixed live).
- **The unknown command type is a SILENT failure**: the reference's
  `else` branch appends the failed result with the bare error (no
  `ValueError:` prefix) and NO log line — only the `except` path logs
  (`f"[directory] command {ctype} {id} failed: {e}"` with str(e),
  while the result's error rides the `f'{type(e).__name__}: {e}'`
  prefix). The port's `pySilentFail`/`pyValueError` pair carries both
  renderings.
- **The reference's `on_disk_cats` is a Python SET** (hash-randomized
  iteration); the Go convergence is the walk's first-seen order — the
  final `(order, lower(name))` STABLE sort makes both identical
  whenever the keys are distinct (every real vault and the golden
  fixture guarantee this; the ReadDir-order law's family).
- **`push_mirror_repo`'s runner** inherits the parent env and never
  redacts — its module has no `_git_env` (the vaultseal/goodrepos
  law family does NOT extend here; reality wins). HideWindow rides
  ADR-0009's Windows quirk list. The first-ever push needs upstream
  tracking configured (plain `git push` with push.default=simple
  refuses otherwise) — the owner's real vault has it; the reference
  shares the failure mode.
- **CPython `re` vs Go regexp**: `\s`/`\S`/`\b` are Unicode-aware on
  str in Python, ASCII-only in Go — the note grammar's fixtures are
  ASCII-whitespace throughout, so the engines agree on every
  reachable input; the `(yes|no)\b` tail check is hand-rolled with
  Python's Unicode `\w` semantics.
- **The `lstrip('www.')` character-set quirk** (T-40's banish gate):
  Python strips any leading `w`/`.` characters (`'wwww.example.com'`
  → `'example.com'`), not the prefix — TrimLeft, not TrimPrefix (a
  T-49 reality-wins fix over the landed floor; the golden pins it).
- **`getpass` → plain stdin read**: Python's getpass itself falls
  back to a plain read (with a warning) where it cannot hide the
  terminal; Go's stdlib has no cross-platform getpass — the
  `--make-admin-hash` leg reads plainly (the reference's own fallback
  door).
- **`banished.json` reads/writes ride the RAW tree** (not the typed
  `BanishEntry`): the owner's hand-edited extra keys survive the
  banish add/remove round-trip verbatim (the reference preserves the
  full dicts).
- **The drift-seal log adapter**: `publish_after_batch`'s
  `Callable[[str], None]` log meets the seal's `(msg, level)` log —
  the plain lines tag "info" (documented convergence).

## SESSION-063 dated record — the CI incident: git's version-dependent fatal wording normalized (the push_mirror_plain/logs pin)

go-ci run 39837881142 (on the SESSION-063 close push `f9c81ee`) went
RED on `TestGolden_PushMirror` — `push_mirror_plain/logs` diverged:

- **got** (the runner): `[directory] git rev-parse
  --is-inside-work-tree failed: fatal: not a git repository (or any
  parent up to mount point /)\nStopping at filesystem boundary
  (GIT_DISCOVERY_ACROSS_FILESYSTEM not set).`
- **want** (the golden): the 2.47-era wording (`or any of the parent
  directories): .git`).

Root cause (byte-settled from the two runs' logs): the runner's git
is 2.55.0 on BOTH runs (the green `d2fa36c` run at 18:41 and the red
`f9c81ee` run at 19:11 — both fresh executions, not cache), but the
two runs rode DIFFERENT ubuntu26 image releases (20261021.171 green
vs 20261012.163 red): on one filesystem layout the walk-up from
`t.TempDir()` to the root crosses a mount boundary (git emits the
"up to mount point /" + "Stopping at filesystem boundary" form), on
the other it does not (the old wording). The golden had accidentally
recorded ONE machine's git wording as if it were the app's contract —
the reference emits whatever ITS local git says, so the wording is
environment-owned, never a cross-machine contract.

The fix (tests never weakened — the pin SHARPENED to the app's real
contract): `internal/directorysync/golden_test.go` gained
`normGitWording` — both known git fatal variants normalize to one
canonical token (`fatal: not a git repository (<git-version-dependent
wording>)`) on BOTH sides of the push_mirror_plain/logs comparison
(the run's captured text AND the golden's recorded text; the golden's
bytes stay untouched — goldens are never hand-edited). Everything
else stays byte-exact. The teeth, proven twice:

1. The fixture test `TestNormGitWordingEnvironmentVariants` pins the
   exact CI failure bytes: both real-world shapes (the sandbox's
   2.47 form AND the runner's 2.55 form) normalize to the SAME token
   while the app's own lines pass through untouched.
2. The mutation: the app's skip line reworded ("vault is not a
   repository") — the comparison FAILS exactly (the normalized token
   visible on both sides, the app line differing) — reverted
   byte-identical, GREEN.

The battery after the fix: main 2398 PASS / 0 FAIL over 36 ok (2397
+ the fixture test), race-clean, gofmt/vet/build clean. The incident
class is now a standing law: environment-owned text (a tool's
version-dependent wording) must never ride a golden pin — normalize
at comparison time on both sides, or stub the tool.
