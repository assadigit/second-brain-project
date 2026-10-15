# Module Spec — `app/gitcurator/constants.py`

- **LOC**: 366 · **Go home**: `internal/constants` · **Status**: port
- **Responsibility**: single source of app-wide constants — paths, ownership stamps, design-token
  legacy palette, the full default config, GitHub category→folder map, the default GitHub analysis
  prompt, expected Worker version. Pure stdlib (importable by GUI, CLI, subprocess, CI).

## Key symbols (evidence `constants.py:`)

| Symbol | Lines | Value / semantics |
|---|---|---|
| `APP_DIR` | 28 | dir *containing* the `gitcurator` package (i.e. `app/`); anchor for every relative path |
| `CONFIG_FILE` | 30 | `APP_DIR/config.json` |
| `MANAGED_BY_GITCURATOR` | 37 | `"gitcurator"` |
| `NOTE_SCHEMA_VERSION` / `GITHUB_PROMPT_VERSION` | 38-39 | `"1"` / `"gh-v1"` |
| `OWNERSHIP_BANNER` | 43 | `"> [!info] Managed by GitCurator — machine-written note."` |
| `DEFAULT_TAXONOMY_PATH` / `resolve_taxonomy_path` | 47-55 | `APP_DIR/taxonomy/website-library-categories.md`; config `taxonomy_path` override (whitespace = unset) |
| `COLORS` | 78-121 | 27 hex design tokens (legacy QSS palette; superseded by `gui/theme.py` for the UI — keep for parity of any code reading it) — *corrected 2026-10-15, T-10 golden audit (was "31")* |
| `CONFIG_EXAMPLE` | 122-305 | full default config — **the defaults contract** (see §Defaults below) |
| `EXPECTED_WORKER_VERSION` | 313 | `"0.40.0"` |
| `CATEGORY_FOLDERS` | 315-340 | 24-entry ordered map LLM category key → vault subfolder (e.g. `"Agents": "AI-Domain/Agents"`, `"Scraping": "Tools/Scraping"`, `"Uncategorized": "Uncategorized"`) |
| `CATEGORY_KEYS` | 341 | insertion-ordered key list (order = classifier prompt order!) |
| `DEFAULT_SYSTEM_PROMPT` | 344-365 | the live GitHub analysis prompt (JSON keys: `short_summary` <140, `summary`, `how_it_works`, `core_value`, `features` 3–7, `difference`, `category` from `', '.join(CATEGORY_KEYS)`, `confidence` 0–100, `tags` ≤5; "Return only valid JSON.") |

## Defaults contract (CONFIG_EXAMPLE highlights)

*(Corrected 2026-10-15, T-10 golden audit — the previous version listed several
read-site defaults from other modules as if they were CONFIG_EXAMPLE keys; the
goldens in the Go repo's `internal/constants/testdata/` are the binding record.)*

- `proxy`: `{enabled: False, type: "socks5", host: "127.0.0.1", port: 10808, use_for_web: True}`
- `web_blocked_domains`: the LAW list (41 domains, see spec-core-links §3); `web_self_domains`:
  `["github-to-obsidian-bot.aliassadi-plus.workers.dev"]`
- in CONFIG_EXAMPLE: `web_user_agent ""` and the four door booleans
  `web_archive_fallback/web_doh_probe/web_impersonate_fallback/web_impersonate_autopip`
  (all True). NOT CONFIG_EXAMPLE keys (read-site defaults in `web_fetch.py` /
  the banner throttles in the processing code): `web_fetch_timeout_s`,
  `web_fetch_max_bytes`, `web_domain_delay_s`, `banner_throttle_10`,
  `banner_throttle_50`, `summary_keep_last`, `notfound_strike_threshold`,
  `banish_confirm`, `banish_confirm_timeout_s`.
- LLM: `ollama {base_url http://localhost:11434, model "qwythos-9b"}`,
  `llm_provider "ollama"`,
  `cloud_api_url https://api.openai.com/v1`, `cloud_model gpt-4o-mini`,
  `llm_num_ctx 8192`, `llm_max_output_tokens 0`, `models {classify "", analyze ""}`,
  `llamacpp_api_url http://127.0.0.1:8080/v1`, `embedding_model ""`
- pacing: `timeout_per_repo 60`, `max_retries 3`, `delay_between_api_calls 0.5`
- vaults: `vault_path/website_vault_path/manual_vault_path ""`, `vaults_history []`,
  `taxonomy_path ""`, `website_repo_name/manual_repo_name ""`, `pipelines {github True,
  websites False}`, `vault_hygiene_on_start True`
- seals: `vaultseal {enabled True, repo_name "", auto_push True, create_repo False}`,
  `goodrepos {enabled True, repo_name "", auto_push True}`
- Telegram credentials: `telegram_api_id 0`, `telegram_api_hash/telegram_phone ""`
  (the Option-B compat law keeps them in the surface — ADR-0004 Am.3).

CONFIG_EXAMPLE is **53 top-level keys (71 incl nested)**. `app/config.example.json`
is a drifted human-facing template — 68 top-level / 84 incl nested keys, placeholder
credentials (`YOUR_*_HERE`), `proxy.enabled true`, a 17-domain snapshot of the law
list, and it OMITS 20 CONFIG_EXAMPLE keys (the whole gdrive block, `llm_num_ctx`,
`models`, `llamacpp_*`, `embedding_model`, `llm_max_output_tokens`, `cloudflare_*`
sync quartet) — it is NOT the defaults source and is never read at runtime (only
some Python tests read it as a packaging artifact). *(Corrected 2026-10-15, T-10
golden audit — was "Full 109-key template".)*

## Compatibility requirements

1. Defaults must match **value-for-value** (a config missing a key gets exactly these).
2. `CATEGORY_KEYS` order is behavior (prompt text + reverse map); Go must keep the ordered slice.
3. `APP_DIR` anchoring: Go resolves paths from the executable's directory with a
   `GITCURATOR_HOME` escape hatch for tests (Python had none — additive only, default identical).

## Tests

- Golden test: dump the Go defaults as sorted JSON and diff against a JSON dump of Python's
  `CONFIG_EXAMPLE` (generated once in the differential harness).
- Category map/folder translation tests.

## Notes / findings

- `gui/constants.py` (258 LOC) is a **stale duplicate** imported only by 5 test files — not ported.
- Consumers: ~40 modules (investigation 1-a).
