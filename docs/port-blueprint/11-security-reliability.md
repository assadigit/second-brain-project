# 11 — Security and Reliability

> Only the concerns applicable to this actual application (a personal, local-first desktop tool
> with outbound integrations). No compliance theater.

## 1. Trust boundaries

1. **LLM output is untrusted input** (SPEC.md §2.6): every value landing in YAML frontmatter,
   filenames, or paths passes the sanitizers (`core/note_builder.py` caps/charset rules,
   `core/storage.safe_filename` traversal-proofing, banner-basename fullmatch gate). The Go port
   must keep the exact sanitizer batteries and their tests (the repo had a security audit fix
   here once — do not regress).
2. **Web page content is untrusted**: extraction only; no evaluation; regex-based thinness checks.
   The prompt pipeline wraps READMEs in `<<<README_BEGIN>>>` delimiters
   (`gui/worker/llm.py:477-488`).
3. **Worker responses are semi-trusted**: JSON parsed tolerantly; failures degrade honestly.
4. **Config contains secrets** — the hygiene laws below.

## 2. Secret handling (the token laws — all must port)

- GitHub token: used for API calls and **one-time push/fetch URLs only**
  (`https://x-access-token:<tok>@github.com/…`); never written to `.git/config`; never logged;
  `.git/FETCH_HEAD` rewritten with `***` after every fetch (`integrations/vaultseal.py:606-623`);
  all captured git output redacted (`goodrepos.py:531-535`).
- Telegram api_hash / phone: masked in UI (first 5 chars + bullets, `connection_check.py:555-614`);
  session file gitignored.
- Secret query values scrubbed from every stored URL (`core/links.py:547-576`, `…` replacement).
- `config.json` / `cache.db` / `*.session` / vault content gitignored (existing `.gitignore` — the
  Go repo additions must replicate these rules for `go-desktop/` artifacts: `bin/`, local test
  DBs).
- User memory file chmod 600 (best-effort on Windows).
- **Never print, log, or commit secrets** (AGENTS.md house rule; the repo was scrubbed once before
  going public — README §Security). The Go code must never log bearer URLs or keys; tests assert
  redaction.
- The blueprint itself contains **no credentials** (mission rule; verified in review).

## 3. Network security posture

- Websites fetcher: TLS **always verified** (`ssl.create_default_context()`,
  `web_fetch.py:1737`).
- LLM/cloud + connection battery: `verify_ssl` config default **false** (historical censored-
  network setting, `netctx.py`) — port the split exactly (per-transport config), and keep the
  config key.
- HMAC client: constant-time compare is server-side; client computes exactly per protocol.
  Timestamps ms-precision; replay window ±5 min server-side.
- Bearer directory token: constant-time compare server-side; token never logged.
- Admin/webhook surfaces belong to the Worker (out of scope).
- Proxy: SOCKS5 with credentials; never logged.

## 4. Resource cleanup and reliability (the patterns to keep)

- **Atomic writes only** for vault content (temp file + fsync + rename; cleanup on failure).
- **Never-raise boundaries**: batch item failures isolated; integration failures return result
  objects; the GUI never crashes on a background failure (BaseException-safe wrappers).
- **Subprocess hygiene**: Telethon child killed on app shutdown (registry + kill-all); stuck
  children killed with surfaced reasons; Chrome sessions closed gracefully
  (`Browser.close` before tree-kill so profile cookies flush).
- **Lock hygiene**: single-instance lock with PID liveness; telegram single-slot lock with
  owner-scoped release + watchdog.
- **Idempotency laws**: vault mirror push (sync_id), backfill chunks, `INSERT OR REPLACE` upserts,
  managed gitignore blocks rebuilt not appended, banish gate re-asks safely, dry-run shadow cache.
- **Fsync discipline**: file fsync before rename (no directory fsync — same as Python; acceptable
  on NTFS).
- **Timeout discipline**: every network call and LLM call has a wall-clock timeout (SPEC §2.7);
  Go contexts enforce them for real.

## 5. Resilience requirements (ported verbatim)

- One failing link never stops the batch; one failing vault never stops the others.
- Disk-full: honest modal + pause; headless skip.
- Worker down: degrade network halves, keep local halves measured; restart recovers numbers.
- Git divergence: reconciliation ladder; rescue branch; never force-push main.
- DB locked/broken: per-store best-effort with honest error lines; `record_404` fails CLOSED
  (returns threshold) so quarantine is fail-safe (`gui/cache_db.py:497-519`).

## 6. External input validation inventory

| Input | Validator | Go home |
|---|---|---|
| LLM JSON | `extract_json` (fences → brace-balance) + schema checks per prompt | `internal/llmclient` |
| Category answers | taxonomy exact-name validation, retry → `_review` | `internal/websitepipeline` |
| Filenames/paths | `safe_filename`, `_SAFE_FOLDER_NAME`, `_SAFE_NOTE_NAME`, `_bend_path` `_v` law | `internal/storage`, `internal/vaultscan` |
| Frontmatter surgery | bounded key scanning; never unterminated blocks | `internal/notestate` |
| URLs | both normalizers + `_VALID_REPO_CHARS` (defeats `..` traversal) | `internal/links` |
| Import files | UTF-8-sig/latin-1 tolerant read; per-line grammar; nothing silently dropped | `internal/links` |
| HTML | parser-based extraction; caps at 20k chars | `internal/webextract` |
| Chrome CDP frames | 64 MB frame guard; message-id skipping | `internal/chrometabs` |
| git output | porcelain parsing, redaction | `internal/vaultseal` |

## 7. Known risk notes

- `verify_ssl=false` default on the LLM path is the owner's deliberate censored-network posture;
  the websites fetcher verifies regardless. Do not "fix" silently.
- Windows junction creation (`chrome-real-link`) uses `cmd /c mklink /J` — keep the guarded
  removal path (never delete a real directory).
- The fifth door drives the owner's real Chrome profile via a junction link — the code already
  refuses debugging the default path directly; preserve that guard.
