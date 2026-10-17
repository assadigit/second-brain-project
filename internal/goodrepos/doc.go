// Package goodrepos ports integrations/goodrepos.py — the automatic
// post-run publisher of the curated vault as a PUBLIC "good-repos"
// GitHub directory (the emoji README + the notes mirrored into
// category folders; ADR-0009: git operations stay CLI-subprocess,
// never a library).
//
// T-48 (SESSION-023) landed the whole module:
//
//   - the construction (goodrepos.go): the v0.83.0 creation guard's
//     allowCreate opt-in, the autoPush AND token degradation, the
//     staging lifetime (mkdtemp OUTSIDE any repository, reuse across
//     publish() calls, the exported CleanupStaging — Go has no
//     atexit, the CLI's defer is the honest door), the deterministic
//     helpers (github_anchor / display_name / emoji_for /
//     format_stars), PublishResult (ok/describe/to_dict with the None
//     doors re-opened as null), and _git_env (the law's origin:
//     GIT_TERMINAL_PROMPT=0, the GIT_DIR-family stripped);
//   - the note machinery (notes.go): the tiny frontmatter subset
//     (parseScalar's space-keeping inline lists, the block-list
//     continuation), _to_int / _owner_repo_from_url / _as_tags (the
//     malformed-scalar tolerance) / _as_aliases, _parse_note (the
//     text-mode universal-newlines read, the title/TL;DR/"What is
//     it?" body mining, the display-name ladder url → aliases →
//     filename stem), and scan_vault (the os.walk category walk,
//     IGNORED_DIRS pruned, the vault root never a category);
//   - the README writer (readme.go): the emoji-rich categorized
//     directory — the stats line with the repo/category plural doors,
//     the Contents anchor lines, the display-name alphabetical orders
//     (the emoji must never decide), the stars-desc entry sort;
//   - the publisher (publisher.go): the _git runner (timeouts 60 s /
//     push+fetch 180 s, HideWindow, THE REDACTION LAW — the token
//     never survives in ANY captured output), the GitHub REST seam,
//     _owner_login (the 401 fresh-token hint), _ensure_repo (THE
//     CREATION GUARD: the empty-name fail-fast, the 404
//     never-creates-without-opt-in error, the POST /user/repos
//     private payload, the 422 race re-check), _fetch_history (+ the
//     FETCH_HEAD scrub), the staging tree builder (copy2 converges to
//     a mode-preserving content copy), and publish() (the skip-when-
//     unchanged door, the commit-message stamp, the one-time token
//     URL push with the defensive remote removes, the dry-run door
//     that runs NO git);
//   - the config bridge (bridges.go) and the standalone CLI (cli.go).
//
// The instance seams (the T-47 law): apiFn is the _api mock surface
// the reference's own creation-guard suite patches; tokenURLFn is the
// _token_url builder — the one door that lets the fetch/push ladder
// run against LOCAL BARE REPOS in the tests (the default builds the
// one-time https token URL; the token NEVER lands in .git/config).
package goodrepos
