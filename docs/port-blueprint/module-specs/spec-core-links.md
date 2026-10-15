# Module Spec — `app/gitcurator/core/links.py`

- **LOC**: 798 · **Go home**: `internal/links` · **Status**: port (byte-exact grammar)
- **Responsibility**: THE single source of link parsing/normalization: GitHub URL extraction,
  website canonicalization, THE LAW blocked-domain lists, platform classification, self-domain
  detection, secret scrubbing, faithful import-file parsing. Pure leaf module (~40 importers).

## Public symbols (evidence `links.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `GITHUB_URL_PATTERN` | 39-41 | `r'https?://(?:www\.)?github\.com/([a-zA-Z0-9\-_.]+)/([a-zA-Z0-9\-_.]+)'` — unanchored |
| `ALL_LINKS_PATTERN` | 44 | `r'https?://[^\s<>"\)\]]+'` |
| `GITHUB_IO_PATTERN` | 50-52 | `r'https?://([a-zA-Z0-9\-_.]+)\.github\.io/([a-zA-Z0-9\-_.]+)'` |
| `GIST_URL_PATTERN` | 56-57 | gist host, IGNORECASE |
| `_VALID_REPO_CHARS` | 65 | `^[a-zA-Z0-9](?:[a-zA-Z0-9\-_.]*[a-zA-Z0-9])?$` (fullmatch; defeats traversal) |
| `extract_github_urls(text, dedupe=True)` | 72-90 | findall → canonical `https://github.com/{o}/{r}`, order-preserved, deduped |
| `extract_all_links(text)` | 93-103 | rstrip trailing `'.,);:!\'"'` **character-set**; duplicates KEPT |
| `map_github_io_url(url)` | 106-121 | → `https://github.com/{owner}/{repo}` or `''` (bare pages site stays website) |
| `split_links(text)` | 129-163 | → `(github_urls, non_github, raw_count)`; per-link raw dedupe; github.io mapped |
| `is_github_url` / `is_gist_url` | 166-126 | match + validity |
| `clean_url` | 179-184 | strip; prepend `https://` unless starts `"http"` |
| **`normalize_url(url)`** | 187-223 | GitHub/VaultIndex identity — see §Normalizer A |
| `domain_of(url)` | 226-242 | lowercased netloc (port+userinfo kept); scheme injected; never raises |
| `site_key_of(url)` | 245-263 | domain minus leading `www.` — ONE NOTE PER SITE key |
| **`LAW_BLOCKED_DOMAINS`** | 298-335 | 41-domain tuple (see §3) |
| `LAW_PLATFORM_DOMAINS` | 342-347 | law minus GitHub group (37) |
| `blocked_domains_from_config` / `platform_domains_from_config` | 391-412 | law ∪ config extras (extras can only add) |
| `domain_is_blocked(url, blocked)` | 415-432 | host (port-stripped) == entry or `endswith('.'+entry)` |
| `PLATFORM_DOMAINS` / `PLATFORM_TABLE_FILES` / `classify_platform` | 446-495 | 9 platforms, first-match-wins dict order; `_inbox` table filenames |
| `DEFAULT_SELF_DOMAINS` / `self_domains_from_config` / `domain_is_self` | 509-540 | bot worker host; empty config = opt-out |
| `_SECRET_URL_TEXT_RE` / `scrub_url_token` / `scrub_urls_in_text` | 547-576 | secret-named query values (≥16 chars) → `…` U+2026 |
| `_TRACKING_PARAM_RE` | 584-587 | `^(utm(?:_[a-z0-9_]+)?\|fbclid\|gclid\|ref\|ref_src\|ref_url\|igshid\|mc_cid\|mc_eid)$` IGNORECASE |
| **`normalize_website_url(url)`** | 590-649 | website identity — see §Normalizer B |
| `dedupe_urls(urls)` | 652-665 | keyed by `normalize_url` (test-only in production) |
| `_BARE_ADDRESS_RE` / `read_import_file` / `parse_import_text` | 678-798 | import grammar (see §Import) |

## Normalizer A — `normalize_url` (ordered steps, `links.py:187-223`)

1. empty → `""`; else `strip()`.
2. **substring replace** `'twitter.com'` → `'x.com'` (case-sensitive, anywhere — domain *or*
   path!) — before anything else.
3. drop `#fragment` (split at first `#`), then `?query` (split at first `?`).
4. strip trailing `/` — via `rstrip('/')` (ALL trailing slashes) unless the result would end
   `://`.
5. lowercase **domain only**: if `"://"` present, split once; if rest contains `/`, split once
   more → `scheme://domain.lower()/path` else `scheme://rest.lower()`. **Path case preserved**
   (GitHub is case-sensitive).

## Normalizer B — `normalize_website_url` (`links.py:590-649`)

1. empty → `""`; strip; drop `#fragment` (first `#`).
2. `urlparse`; on exception → return the stripped url unchanged.
3. `scheme = 'https'` **unconditionally**; `host = netloc.lower()` with leading `www.` dropped.
4. query: split `&`; drop empty pairs; drop a pair iff its key (before first `=`) matches
   `_TRACKING_PARAM_RE` (case-insensitive); **all other params kept, order preserved**.
5. path folding (v0.64.0 one-spelling): `path=='/' and no kept` → `''`; `path=='' and kept` →
   `'/'`; `path.endswith('/') and len>1` → `rstrip('/')` (all).
6. `f"https://{host}{path}"` + `('?'+'&'.join(kept))` if kept.

## THE LAW (blocked domains) — exact 41 entries

`x.com, twitter.com, t.co, github.com, gist.github.com, github.io, githubusercontent.com,
huggingface.co, hf.co, instagram.com, instagr.am, facebook.com, fb.com, fb.me, fb.watch,
linkedin.com, lnkd.in, youtube.com, youtu.be, youtube-nocookie.com, share.google,
drive.google.com, docs.google.com, forms.google.com, tiktok.com, threads.net, threads.com,
snapchat.com, pinterest.com, pin.it, twitch.tv, discord.gg, discord.com, discordapp.com, t.me,
telegram.me, wa.me, whatsapp.com, vk.com, bsky.app, weibo.com`

`PLATFORM_DOMAINS` (ordered): x_twitter (x.com, twitter.com, t.co) · youtube (…, youtu.be,
youtube-nocookie.com) · reddit (reddit.com, redd.it) · linkedin · medium · huggingface (…, hf.co)
· arxiv · package_registry (npmjs.com, pypi.org).

## Import-file grammar (`parse_import_text`, `links.py:708-798`)

- read: bytes → `utf-8-sig`; `UnicodeDecodeError` → `latin-1` `errors='replace'`
  (`read_import_file:693-705`).
- per line: strip + `lstrip('\ufeff')`; blank or `#`-comment → `skipped_comment_lines`;
  `extract_all_links` per line; no links → `_BARE_ADDRESS_RE` match
  (`^(?:[a-zA-Z0-9][a-zA-Z0-9\-.]*\.[a-zA-Z]{2,}|\d{1,3}(?:\.\d{1,3}){3})(?::\d+)?(?:[/?#].*)?$`)
  → `clean_url` + route; else `unparsed`.
- routing: GitHub match → canonical (duplicates counted); github.io mapped; websites keyed by
  `normalize_website_url` for dedupe but the **raw link** is kept in the output list.
- returns `{github_urls, website_urls, unparsed, duplicates, raw_count, skipped_comment_lines}`.

## Migration-sensitive semantics

- `re.match` = start-anchored (Go `^`); `findall` with groups → tuples (Go: `FindAllStringSubmatch`).
- Insertion-ordered dedupe everywhere (`seen` sets + ordered lists) — Go: ordered-set helper.
- `rstrip(chars)` character-set semantics; `…`/`🗑️` byte-exact emoji checks.
- `domain_of` keeps ports; only blocked/platform checks strip them.

## Go mapping & tests

- Package `internal/links` with `NormalizeURL`, `NormalizeWebsiteURL`, `SiteKeyOf`, `ClassifyPlatform`,
  `BlockedDomainsFromConfig`, `ScrubURLToken`, `ParseImportText`.
- Tests: port `test_core` grammar cases + NEW Unicode edge fixtures (İ, ß, emoji URLs), the 41-law
  golden, tracking-param matrix, import-file corpus (markdown links, `127.0.0.1:8901/x`, `www.`).
- Differential case: run Python `parse_import_text` vs Go on the same corpus, compare JSON.

## Evidence

Investigation report 1-a (full symbol table with line ranges verified). Consumers: note_state,
linking, website_pipeline, ~40 GUI modules, integrations, CLI.
