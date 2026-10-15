# Module Spec — `app/gitcurator/core/prompts.py` (+ prompt file inventory)

- **LOC**: 98 · **Go home**: `internal/prompts` (+ `assets/prompts` embedded) · **Status**: port
- **Responsibility**: file-based prompt template loader with `{{SLOT}}` filling that **refuses to
  run** on unfilled slots, None/empty values, or values containing slot markers (untrusted-input
  guard).

## Symbols (`prompts.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `PROMPTS_DIR` | 29 | `APP_DIR/prompts` |
| `_SLOT_RE` | 32 | `r'\{\{([A-Z][A-Z0-9_]*)\}\}'` |
| `PromptError` | 35-36 | exception type (Go: sentinel error) |
| `slots_in(template)` | 39-45 | ordered, deduped slot names |
| `fill_template(template, **values)` | 48-80 | errors: `"no template given"`; missing/empty → `"prompt slot(s) left unfilled: {names}"`; None → `"prompt slot {n} got None"`; value matching `_SLOT_RE` → `"value for slot {n} contains a {{SLOT}} marker — refusing to fill (untrusted input)"`; literal `str.replace` per slot; leftover slots → unfilled error |
| `load_prompt_file(filename, **values)` | 83-93 | rejects path separators (`basename != filename` → `"bad prompt filename"`); reads `utf-8` (**not** `-sig` — quirk: a BOM would corrupt slot 1); OSError → `"cannot read prompt file …"` |
| `load_prompt(stem, **values)` | 96-98 | `stem + ".txt"` |

## Prompt inventory (consumers)

| File | Consumer | Notes |
|---|---|---|
| `w01_category.txt` | website_pipeline (`:3920`) | websites pass 1 |
| `w02_subcategory.txt` | website_pipeline (`:3955`) | pass 2 |
| `w03_analyze.txt` | website_pipeline (`:3991`) | analysis |
| `s01_vaultscan.txt` | **read directly** by `vault_scan.py:167` (bypasses loader) | scan system prompt |
| `s02_reposscan.txt` | `vault_scan.py:172` direct | repos variant |
| `l01_confirm.txt` | linking (`:161`) | pair confirmation |
| `r01_recall.txt` | recall (`:251`) | recall hooks |
| `01/02/03_*.txt` | **none** | dead (GitHub prompt lives in constants) — not shipped in Go |

## Go notes & tests

- Embed the 7 live prompt files via `go:embed`; keep the same refusal messages.
- The BOM quirk: Python uses plain utf-8; Go `os.ReadFile` + string ops match — do NOT add BOM
  stripping here (parity), but assert the shipped files have no BOM.
- Tests: slot enumeration order; every refusal error string; separator rejection; the
  marker-in-value guard.
