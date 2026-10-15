# Module Spec — `app/gitcurator/core/taxonomy.py`

- **LOC**: 342 · **Go home**: `internal/taxonomy` · **Status**: port
- **Responsibility**: parses the owner-editable Websites category file into a `Taxonomy` model;
  validates LLM answers against parsed names before any path is touched.

## Symbols (`taxonomy.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `TaxonomyError` | 43-44 | error type |
| `_EMOJI_RE` | 52-54 | strips ranges U+1F000-1FAFF, U+2600-27BF, U+2190-21FF, U+2B00-2BFF, U+FE0F, U+200D |
| `_TRAILING_ITALIC_RE` / `_BULLET_RE` / `_EXAMPLE_TAGS_RE` / `_NO_SUBCATEGORIES_RE` / `_DASH_SPLIT_RE` | 57-69 | heading cleanup; `- ` bullets; `Example tags:`; `(no subcategories)`; em/en-dash split |
| `Category` / `Taxonomy` | 100-206 | ordered model; duplicate category names → error; queries (`category`, `is_subcategory_of`, one-liners for prompts, `tag_hints_of`); `folder_relpath(cat, sub)` via `safe_filename` — unknown names raise |
| `parse_taxonomy(text, source_path)` | 218-325 | grammar below |
| `load_taxonomy(path)` | 328-336 | `utf-8-sig` read; OSError → TaxonomyError |
| `load_taxonomy_from_config` | 339-342 | resolves path via constants |

## Parsing grammar (tolerant, line-oriented)

1. `### <Category>` → new category: strip trailing `*(…)*` italic note, strip emoji, strip; empty
   name → error.
2. `## <Section>` → section switch; **ends the current category block**.
3. `---` lines skipped; blank lines keep state.
4. Sections `How this works` and `Not yet covered` ignored wholesale.
5. `Judgment-call rules` → bullets' bodies and non-`*(` prose appended **verbatim** (joined `\n`)
   → `judgment_rules` passed to the classifier.
6. Category body: `Example tags:` line → `#([A-Za-z0-9\-_]+)` hints; `- …` bullet → subcategory
   (strip `**`, cut at first `—`/`–`); rejected when empty / starts `(empty for now` / contains
   `(no subcategories)`; subcategory definition = text after the dash; first prose paragraph
   becomes category definition; `*(no subcategories)*` inner-text markers skipped.
7. No `###` headings at all → error `no '### <Category>' headings found — not a taxonomy file`.

## Compatibility requirements

- Prompt inputs derived from the model (`category_one_liner()`, `subcategory_one_liner(cat)`,
  `judgment_rules`, tag hints) must render identically (they are prompt text → behavior).
- Folder mapping = `safe_filename(cat)[/safe_filename(sub)]`.

## Tests

Port the parser test set (from `test_phase*`/golden): emoji headings, italic notes, dash cuts,
`(no subcategories)`, judgment rules verbatim, empty categories, duplicate names, BOM file, and
the shipped taxonomy file parse snapshot (category/subcategory counts asserted).
