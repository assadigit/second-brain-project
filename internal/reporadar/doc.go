// Package reporadar ports integrations/repo_radar.py — v0.84.0 THE
// REPO RADAR: the read-only detection that finds which existing repo
// backs up each of the three vaults (github / websites / personal)
// and fills ONLY the EMPTY config slots (the owner's typed word
// always wins; nothing is ever created — the v0.83.0 law stands).
//
// T-48 (SESSION-023) landed the whole module:
//
//   - the transport (radar.go): RadarAPI — the minimal read-only
//     GitHub client (UA GitCurator-RepoRadar, the four GitHub
//     headers, the 15 s timeout, me() / list_repos() (3 pages × 100,
//     sort=pushed) / repo_exists() (404 → nil, the _error door) /
//     tree_blob_shas() (recursive; a truncated tree still answers —
//     dominance, not completeness, is the criterion; the
//     str(None)→"None" sha quirk preserved). The tests' fake rides
//     the API interface — the reference's duck-typed FakeAPI seam;
//   - the evidence pieces (evidence.go): parse_remote_repo (the
//     https/ssh/git@ forms, the owner lowercased, git/www/api never
//     owners), remote_origin_repo (the configparser read converges to
//     a minimal INI scan of [remote "origin"] — a documented
//     convergence), blob_sha (git's content-addressed SHA-1 —
//     path-independent, matches GitHub's tree API), sample_vault_files
//     (the diversified round-robin sampler — Go's ReadDir order is
//     the documented convergence over CPython's readdir order), the
//     candidate ranking (non-archived/non-fork, vault-ish names
//     first, the MAX_TREE_FETCHES=24 crowding cure, size NEVER an
//     emptiness signal — the live lesson), and the dominance ladder
//     (≥3 matches AND ≥2× the runner-up, else ambiguous/weak = a
//     report, never a guess);
//   - the detection (detect.go): Detection (a dict with named
//     accessors — it crosses the GUI thread boundary as a plain
//     object), the slot grammar (dot = one nesting level),
//     detect_backup_repo (REMOTE → NAME → FINGERPRINT, every honest
//     closing line), radar_all_vaults (the never-raises doors: a
//     panicking API surface is recovered per slot and SPOKEN in the
//     detail — the T-47 raise-door law), apply_detections (only
//     filled slots land; owner confidence never overwrites; plain
//     dicts flow too), report_lines, radar_pending (the
//     quiet-startup gate).
package reporadar
