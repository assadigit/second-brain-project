// Package vaultseal ports integrations/vaultseal.py — the
// git-CLI-subprocess seal of an Obsidian vault into a PRIVATE GitHub
// mirror (ADR-0009: git operations stay CLI-subprocess, never a
// library).
//
// T-47 (SESSION-022) landed the whole module:
//
//   - the construction (vaultseal.go): the v0.83 creation guard's
//     allowCreate door, the v0.89 local-only belt, the autoPush AND
//     token degradation, the _git subprocess runner (env scrubbing —
//     GIT_TERMINAL_PROMPT=0, the GIT_DIR-family stripped — timeouts
//     60 s / push+fetch 180 s, HideWindow on Windows, THE REDACTION
//     LAW: the token never survives in any captured output), the
//     GitHub REST seam (30 s), SealResult (ok/describe/to_dict with
//     the None doors re-opened as null);
//   - the bootstrap family (bootstrap.go): the nested-repo guard →
//     git init -b main (the plain-init fallback + symbolic-ref), the
//     managed .gitignore block, THE LOCAL-ONLY BELT's second managed
//     block (rebuilt from the current list) + the untrack law (git rm
//     -r --cached — the files NEVER leave the disk), the local-only
//     identity, ensureRepo (the token-less remote add, the 404 →
//     creation-guard error, the opt-in POST /user/repos payload, the
//     422 race re-check);
//   - the seal (seal.go): the belt → add -A (pathspec-limited) → the
//     unchanged skip → the commit-message doors → the one-time token
//     URL push;
//   - THE DIVERGENCE LADDER (push.go): plain push → fetch +
//     FETCH_HEAD token redaction → stale-rejection retry / rebase /
//     unrelated-histories merge / the seal-rescue branch — main is
//     NEVER force-pushed;
//   - the resync (resync.go): write-tree delta counts (--no-renames),
//     the commit-tree reconciliation with both parents, the
//     empty-tree baseline 4b825dc…, the one fetch-first retry, never
//     a force;
//   - status (status.go) — pure introspection, no network;
//   - the four config bridges (bridges.go) + the local-only join;
//   - the standalone CLI (cli.go) and THE DRIFT SEAL's pure half
//     (drift.go — the plan + the runner with THE QUIET LAW; the
//     publish door rides the T-49 seam);
//   - CPython 3.12's fnmatch (fnmatch.go) — the belt's matcher.
//
// The belt floor landed with it in internal/directorysync (belt.go —
// local_only_note_patterns + private_exclusion_patterns + their pure
// helpers + the shipped directory-config embed at
// assets/directorysync), exactly the way T-45 landed the repobanish
// floor: the bridges consume it; T-49 completes the module.
//
// Documented convergences (the spec's Go notes carry the full list):
// the reference's never-raises wrappers are unreachable in Go; the
// ~user expanduser form is not ported (the config never carries it);
// the belt's import-failure door (try/except → []) has no Go analog;
// a non-JSON API body answers the parse door; the read of .gitignore
// decodes UTF-8 with the maximal-subpart U+FFFD law.
package vaultseal
