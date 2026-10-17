# Module Spec — `app/gitcurator/core/website_directory.py`

- **LOC**: 432 · **Go home**: `internal/websitepipeline/directorynote.go` · **Status**:
  port (T-41, SESSION-016 — complete; the pipeline's constructor calls the sweep,
  the GUI batch driver calls the rebuild)
- **Responsibility**: THE LAW's disk edition + the categorized Website Directory:
  every app-owned note on a banned domain leaves the vault for
  `.trash/banned-domains/`; the banned platforms' `_inbox` tables quarantine whole;
  `000 📚 Website Directory.md` regenerates after every curation run. No PyQt, no
  network, stdlib only — importable from the GUI worker, the headless CLI, the
  backfill tool, and the CI gate.

> **Provenance (SESSION-016, reality-wins):** this standalone spec file was
> CREATED at T-41's port commit. The task list (`13-implementation-task-list.md`
> T-41) and the session state card had referenced `spec-core-website-directory`
> all along, but the spec content lived inside
> `spec-core-website-pipeline.md` §`website_directory.py` — the standalone file
> never existed. The content below is that section, expanded with what T-41's
> port actually verified (the walk law, the universal-newlines law, the
> full-lowercase sort keys, the divergence records).

## Module constants

`DIRECTORY_FILENAME "000 📚 Website Directory.md"` (app-owned; sorts to the TOP of
the vault) · `QUARANTINE_RELPATH ".trash/banned-domains"` (inside Obsidian's
hidden `.trash`; VaultSeal never backs it up, VaultIndex/mirror/note_state never
read it) · `_SKIP_DIRS {_inbox, _review, _missing, _moc, attachments, .obsidian,
.trash, .git}` (the DIRECTORY walk) · `_SKIP_DIRS_SWEEP {_inbox, .obsidian,
.trash, .git}` (the SWEEP is broader — it ENTERS `_review/_missing/_moc` because
the legacy banned-domain pile lives in `_review`; the `_inbox` NOTE walk is
skipped because the tables are handled separately) · `_TLDR_MAX 140` /
`_TITLE_MAX 80` · the light regexes `_FM_LINE_RE` / `_TLDR_RE` / `_H1_RE`
(frontmatter key lines, the first `> **TL;DR:**` line, the first `# ` heading).

## The light parse (`_parse_note`)

Read the first **512,000 decoded characters** (text-mode read counts characters,
not bytes) with `errors='replace'` (invalid bytes → U+FFFD, maximal-subpart) and
**universal newlines** (`\r\n` and a lone `\r` both become `\n` BEFORE the read
count — every reference text-mode read of a note file translates identically).
A file with no frontmatter (`---` at position 0, a `\n---` terminator) is not a
GitCurator-shaped note → `None`, never an error. The frontmatter dict carries
`app_owned` (`managed_by: "gitcurator"` — EXACT match, case-sensitive),
`source`, `category`, `subcategory`, `pricing` (default `unknown`),
`fetch_status`; the body's first `> **TL;DR:**` line and first `# ` heading
become `tldr`/`title` (an em-dash, hyphen, or empty TL;DR reads as empty).

## The walk (`_iter_note_files`)

os.walk's top-down order, EXACTLY: a directory's own files (sorted) yield before
ANY of its subdirectories' files; subdirectories recurse in sorted order.
Dot-folders and the skip-set folders are pruned; dot-FILES ride the file list
(only directories carry the dot rule); the directory note itself is excluded by
name; `.md` matching is case-insensitive; symlinked directories are never
followed (a symlink-to-dir rides Python's `dirs` list and is never descended,
never yielded — `scanned` does not count it); a symlinked FILE is yielded like
any other.

## `sweep_banned_notes` (170-272) — THE LAW, disk edition

Every APP-OWNED note whose `source:` frontmatter is on a banned domain moves to
`<vault>/.trash/banned-domains/` (dry-run aware: the moves are recorded, not
performed; idempotent: the quarantine folder is skipped by the walk). Hand-written
notes with a banned source are the owner's — counted (`kept_handwritten`) and
reported, never touched. Every app-owned banned note leaves — `_review`
placeholders AND real notes alike (the law has no "but it processed fine"
exception; this is what cleans the legacy `x_com_i_status_*.md` review pile).
The move round-trips TEXT MODE: read `errors='replace'` + universal newlines →
write — a quarantined CRLF/CR note lands newline-normalized.

v0.35.0: after the note walk, the banned PLATFORMS' `_inbox` tables
(`_inbox/x_twitter_links.md`, `youtube_links.md`, …) quarantine to the same
`.trash/banned-domains/` folder — a platform whose EVERY domain is on the ban
list is never collected anymore. A table moves only when ALL its domains are
banned (partially-banned platforms — and `other`, which is not a platform — keep
their tables; their banned ROWS are pruned by the intake writer's prune pass).
`moved_tables` rides `PLATFORM_TABLE_FILES`' insertion order. The production
pipeline constructor runs the sweep (production only — an injected `fetch_fn`
never touches the vault; guarded by the blanket except so bookkeeping never
kills a batch). Returns `{'moved': [(url, src, dst)], 'kept_handwritten': n,
'scanned': n, 'moved_tables': [(platform, src, dst)]}`.

## `build_website_directory` (308-432) — the consolidated categorized index

Regenerates `<vault>/000 📚 Website Directory.md`: frontmatter
`managed_by: "gitcurator"` / `kind: "directory"` / `generated: "<local date>"`;
`# 📚 Website Directory`; the info callout with site + category counts; `## Cat
(n)` / `### Sub (n)` headings in TAXONOMY order (categories: taxonomy order then
leftovers alphabetically; subcategories: taxonomy order, then non-empty
leftovers sorted, then the empty-subcategory group last — except under
Uncategorized, where there is no taxonomy ordering at all); one line per site:
`- [[path|alias ≤80]] — TL;DR ≤140 — pricing — [↗](url)` joined with `" — "`
(the pricing word only when `free|freemium|paid`; the wiki-link path strips a
case-insensitive `.md` suffix; the alias strips `[]|^#` characters, falling back
to `Untitled`; the title falls back to the filename stem). Entries sort by
`(title.lower(), relpath.lower())` — Python's FULL lowercase (İ → `i`+U+0307,
final sigma), not Go's simple per-rune mapping. Only app-owned, source-bearing,
non-blocked-domain notes are indexed; `_review` notes are counted in the closing
`> [!todo]` callout so nothing looks silently missing; the file closes with the
regeneration line. The directory file itself is app-owned and has no `source:`
line, so the VaultIndex dedupe layer never sees it. Returns `{'path', 'sites',
'categories', 'review'}` or `None` when the vault is missing/empty (nothing to
index yet) or the write failed (logged, never raised). The taxonomy falls back
to `load_taxonomy_from_config(config)` when `taxonomy=None` and config is set
(a failed load degrades to no-taxonomy ordering).

## Tests

Output golden byte-compare (generated by `testdiff/goldens/dump_directorynote.py`
from the unmodified reference, frozen clock, no per-run facts): the light-parse
battery (quotes/whitespace/dashed keys, TL;DR/H1 adversarials, CRLF and CR-only
bodies, body rule-line reconstruction, invalid UTF-8, the 512,000-char cap);
the walk-order battery (files-before-subdirs, both skip sets, dot-files,
symlinks); the sweep flows (the legacy pile, hand-written kept, the platform
tables full/partial/order-pinned, collisions, idempotence, dry-run, missing
vault, empty ban list, the scanned counter); the directory regenerations (the
rich vault with taxonomy order + leftovers + the empty-sub group, the
full-lowercase sort pins, the trims, the pricing variants, the real bundled
taxonomy's own order, no-taxonomy and config-load fallbacks, regeneration
stability) — THE OUTPUT GOLDEN, byte-for-byte. Plus the ported reference legs
(`tests/test_lawfix.py`: the sweep + directory + production-constructor
integration; `tests/test_socialomit.py`'s sweep leg).

> **SESSION-016 (T-41) — what actually shipped, and the reality-wins fixes the
> golden forced.** `internal/websitepipeline/directorynote.go` landed the whole
> module (SweepBannedNotes + BuildWebsiteDirectory + the light parse + the
> walk), the constructor's seam (`sweepBannedNotesSeam`, a documented no-op
> since T-40) became the REAL guarded sweep, and the differential golden
> (testdata/directorynote.golden.json, 33 sections / 41,085 bytes, generated by
> testdiff/goldens/dump_directorynote.py with the directory clock frozen at
> 2026-10-17 12:00:00) pins it end-to-end. 28 new top-level tests; 938
> race-clean at the close. THE GOLDEN CAUGHT THREE REAL PORT DIVERGENCES, all
> fixed in the same commit: (1) `readHeadRunes` did not translate universal
> newlines — the reference's CR-only and CRLF notes parse through
> `open(..., newline=None)`'s `\r`→`\n` translation, so the port now rides
> `helpers.PyNewlines` before the rune cap (T-40's latent divergence, its
> fixtures were `\n`-only so nothing pinned it); (2) the sweep's content
> round-trip is TEXT MODE — a quarantined CRLF note lands newline-normalized;
> (3) `decodeUTF8Replace` (sitelinks.go) emitted one U+FFFD per BYTE where
> Python collapses each MAXIMAL SUBPART into one U+FFFD — now notestate's own
> law (b"\xe2\x82" → ONE U+FFFD). The walk also forced T-40's `walkVault`
> files-before-subdirs fix (WalkDir interleaves files and dirs lexically;
> os.walk yields a directory's files before any subdirectory's files — the
> banishment scans' list order rides the walk order; T-40's fixtures never hit
> the interleaving case). Sort keys ride `helpers.PyLower` (the İ and
> final-sigma pins); the regex classes carry Python's Unicode `\s` (U+00A0 and
> the space separators are whitespace to the reference's parses). `storage`
> gained `PySplitExt` (the exported os.path.splitext twin — the title
> fallback's primitive).
