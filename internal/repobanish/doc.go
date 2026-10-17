// Package repobanish ports core/repo_banish.py — the GitHub Repos
// vault's banishment gate.
//
// Spec: docs/port-blueprint/module-specs/spec-core-repo-banish.md;
// task T-46 (Phase 4).
//
// LANDED SO FAR (T-45, SESSION-020): the surface vaultscan's
// vault-kind door consumes — the gate's eyes
// (ScanPendingRepoBanishments: both doors, one list, deduped by the
// canonical repo URL, the dismissed-set DB gate, the 🗑-marked
// master-table rows) and the burial (BanishMarkedRepos: the never-
// fetch gates — dismissed + settled + retry-resolved + row-forgotten
// — the .trash/banished move, the confirmed record rows, the
// handwritten-kept sacred law), plus the note-finder the table-gesture
// burial needs. These were ported with T-45 because vaultscan's own
// acceptance (the ported test_reposcan selections) exercises them end
// to end; their behavior battery rides internal/vaultscan's suites.
//
// STILL T-46's: revive_banished_repos (the ♻️ door back),
// run_repo_banishment_pass (the confirmation-gate orchestration), and
// the module's own test_repobanish suite.
package repobanish
