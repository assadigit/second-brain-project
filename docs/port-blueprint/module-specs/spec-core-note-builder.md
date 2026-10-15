# Module Spec — `app/gitcurator/core/note_builder.py`

- **LOC**: 400 · **Go home**: `internal/notebuilder` · **Status**: port (security-critical)
- **Responsibility**: builds the GitHub-pipeline Obsidian note text (YAML frontmatter + body) from
  LLM/GitHub data, treating every value as hostile YAML input. Also the `_missing/` 404
  placeholder note. Its sanitizers are reused by the websites pipeline.

## Symbols (`note_builder.py:`)

Caps (49-54): `_MAX_SEQ_ITEM 60`, `_MAX_TAGS 12`, `_MAX_ALIASES 4`, `_MAX_LANGUAGES 8`,
`_MAX_SCALAR 120`, `_MAX_SUMMARY 2000`. Charset gate `_SEQ_ITEM_OK = [^A-Za-z0-9 \-_.+/()]`
(48).

| Symbol | Lines | Contract |
|---|---|---|
| `sanitize_seq_item(v, max_len=60)` | 57-68 | disallowed→space, collapse ws, strip `-` ends, truncate, strip; None→`''` |
| `sanitize_tags(tags, max_items=12)` | 71-93 | case-insensitive dedupe (first casing kept), drop empties, cap |
| `sanitize_aliases(*names)` | 96-108 | ≤80 chars, ≤4, ci-dedupe |
| `sanitize_languages(langs)` | 111-122 | ≤40, ≤8, exact dedupe |
| `yaml_scalar(value, max_len=120)` | 125-139 | None→`'""'`; `\r`/`\n`→space; collapse; truncate; escape `\` then `"`; **always double-quoted** |
| `sanitize_body_text(v, max_len=None)` | 142-158 | CRLF→LF; strip control `[\x00-\x08\x0b-\x1f\x7f]`; strip; truncate |
| `sanitize_short_summary` | 161-167 | one line, ≤2000 |
| `rep_to_str(rep)` | 174-182 | ≥8 High(Major…), ≥6 Medium…, ≥4 Low…, else Unknown |
| `score_to_rating(score)` | 185-193 | ≥90 Excellent, ≥75 Good, ≥60 Average, else Low |
| `build_note(...)` | 200-345 | full note — exact key order & format below |
| `build_missing_repo_note(url, owner, repo, strikes=1)` | 352-400 | unquoted `source:`/`repo:`/`status: missing`/`reason: 404…`/`strikes:`/`date_recorded:`/`placeholder: missing-repo`; strips trailing `.git`; **no banner, no category** (so note_state skips `_missing`) |

## GitHub note frontmatter (exact order, `:284-301`)

```yaml
---
source: "<yaml_scalar>"
aliases:            # block list when present else "aliases: []"
tags: [a, b]        # flow, sanitized; else "tags: []"
category: "<key>"
stars: <int>        # int(stars) if int/float else 0
org: "<scalar>"
primary_language: <seq-item>   # only when non-empty
languages: [..]                # only when non-empty
credibility_score: <n>/100
date_processed: YYYY-MM-DD     # LOCAL date
last_release: <scalar>         # only when non-empty
cover: attachments/banners/<name>   # only when banner basename fullmatch [A-Za-z0-9\-_.]+
quality: low                   # only when is_low_quality
quality_issues: [..]
managed_by: gitcurator
schema_version: 1             # UNQUOTED (reality-fixed 2026-10-15, T-15: NOTE_SCHEMA_VERSION
                              # interpolates raw — the golden pins `schema_version: 1`)
prompt_version: gh-v1
---
```

Body: optional `![banner](attachments/banners/<n>)` line → `OWNERSHIP_BANNER` → `# {title}` →
optional `> **TL;DR:** …` → `**owner/repo** · ⭐ {stars:,} · 🔧 {lang}` line → sections
"What is it? / How does it work? / Why is it important? (Core Value) / Key Features &
Technologies / Difference from Others / 🏢 Organization & Credibility" → `---` →
`*Source: [GitHub](url)*` → v0.65.0 trash-tag footer paragraph.

Auto quality-issues (243-257): summary <50 chars; <3 features; confidence <30; category
Uncategorized → `is_low_quality=True`.

## Migration-sensitive quirks

- `stars:` frontmatter coerces via `int(...)` (289) but the body `f"⭐ {stars:,}"` runs on the raw
  value — works because GitHub API returns ints; Go struct field is `int` (safe).
- `f"{stars:,}"` comma grouping — Go helper must match.
- Date: `datetime.now()` LOCAL, no timezone.
- Banner reference gate is a security control (fullmatch) — port exactly.

## Tests

- YAML sanitizer battery (port `test_core` cases): quoting, escaping order, caps, dedupe casing.
- Note-format goldens for 6 synthetic repos (diff against Python-rendered strings via harness).
- `safe banner basename` traversal corpus.
- Missing-repo note golden.
