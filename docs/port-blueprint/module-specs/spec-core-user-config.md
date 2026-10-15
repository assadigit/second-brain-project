# Module Spec — `app/gitcurator/core/user_config.py`

- **LOC**: 481 · **Go home**: `internal/config` (user memory) · **Status**: port
- **Responsibility**: credentials + vault identity that survive version updates in one user-level
  file `Documents/.gitcurator`: THE FILL (read side) and THE MIRROR (write side). Never raises.

## Symbols (`user_config.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `ENV_OVERRIDE` / `USER_CONFIG_FILENAME` / `USER_CONFIG_MARKER` | 65-72 | `GITCURATOR_USER_CONFIG` / `.gitcurator` / `"_gitcurator_user_config": 1` |
| `CREDENTIAL_FLAT_KEYS` | 75-90 | 15 keys: `telegram_api_id, telegram_api_hash, telegram_phone, telegram_user_id, github_token, bot_token, bot_username, cloud_api_url, cloud_api_key, cloud_model, llamacpp_api_url, llamacpp_api_key, llamacpp_model, embedding_model` |
| `CREDENTIAL_NESTED_KEYS` / `CREDENTIAL_PREFIXES` | 93-103 | `proxy, ollama, models` / `("cloudflare_","gdrive_","hmac_")` |
| `VAULT_FLAT_KEYS` / `VAULT_NESTED_KEYS` | 112-126 | 6 vault keys / `vaultseal, goodrepos` |
| `is_unset_value(v)` | 155-179 | None; blank str; `YOUR…_HERE`; **numeric 0**; empty list/dict → unset; booleans never unset |
| `_windows_documents_dir()` | 190-221 | ctypes `SHGetKnownFolderPath` GUID `{FDD39AD0-238F-46AF-ADB4-6C85480369C7}` + `CoTaskMemFree`; failure → None |
| `_documents_dir()` | 224-234 | win32 known-folder → `~/Documents` **only if exists** |
| `global_config_path()` | 237-253 | env override (abspath+expanduser, **creates parent**) else documents/.gitcurator; else None (feature sleeps) |
| `apply_global_credentials(config, fresh=False)` | 286-374 | THE FILL: flat credential → flat vault → nested merged key-by-key (caller's dict copied first) → credential-prefix keys → unknown remembered keys added; real local values always win unless `fresh` |
| `mirror_credentials_to_global(config)` | 395-460 | THE MIRROR: real values gain/update; unset local never erases a real remembered value; file created only when something real; chmod 600 best-effort; written via atomic `write_config_file` (dry-run aware) |
| `status_line(path)` | 467-481 | 3 human variants for Settings → Credentials |

## Compatibility requirements

1. Same file name/location/marker; same fill/mirror algorithms (order matters: credentials →
   vault → prefix → unknown).
2. Unset grammar including **numeric 0 = unset** (telegram_api_id 0 is "no credential").
3. Nested merge never mutates the caller's dict; blanks never flow; booleans always real.
4. Windows Known-Folder via `golang.org/x/sys/windows.KnownFolderPath` (same GUID); POSIX
   `~/Documents` only when it exists.

## Tests

Port `test_userconfig` (83 cases): fill/mirror matrices, unset grammar (incl. 0), nested merge
aliasing law, env override creation, sleep mode, chmod advisory, never-erase law.
