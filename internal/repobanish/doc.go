// Package repobanish ports core/repo_banish.py — the GitHub Repos
// vault's banishment gate.
//
// Spec: docs/port-blueprint/module-specs/spec-core-repo-banish.md;
// task T-46 (Phase 4). COMPLETE:
//
//   - T-45 (SESSION-020) landed the surface vaultscan's vault-kind door
//     consumes — the gate's eyes (ScanPendingRepoBanishments: both
//     doors, one list, deduped by the canonical repo URL, the
//     dismissed-set DB gate, the 🗑-marked master-table rows) and the
//     burial (BanishMarkedRepos: the never-fetch gates — dismissed +
//     settled + retry-resolved + row-forgotten — the .trash/banished
//     move, the confirmed record rows, the handwritten-kept sacred
//     law), plus the note-finder the table-gesture burial needs
//     (repobanish.go). Their behavior battery rode internal/vaultscan's
//     suites.
//
//   - T-46 (SESSION-021) completed the module (pass.go):
//     ReviveBanishedRepos (the ♻️ door back — the dismissal cleared, the
//     settlement removed, dry-run gating the note_state write exactly as
//     the reference does) and RunRepoBanishmentPass (the
//     confirmation-gate orchestration — the ungated revive first, the
//     detection scan, the count SPOKEN before the ask over the injected
//     channel factory, the verdict gates ['auto'/'confirmed' bury;
//     'declined'/'timeout'/'defer' keep], THE TALLY with the zero-run
//     answer), plus the module's own test_repobanish suite
//     (repobanish_test.go).
package repobanish
