// Package directorysync ports cloud/directory_sync.py — build_index +
// publish + admin command cycle.
//
// T-40 landed the banish gate floor (banishgate.go — load_banish_list,
// DEFAULT_SHARED_HOSTS, is_banished: the read side of
// Websites/_system/banished.json that WebsitePipeline probes before
// any link is fetched).
//
// T-47 (SESSION-022) landed THE LOCAL-ONLY BELT floor (belt.go) with
// vaultseal — local_only_note_patterns + private_exclusion_patterns +
// their pure helpers (_gitignore_escape, _belt_folds,
// _prettify_folder_name) + load_directory_config (the vault's
// _system/directory-config.json with the embedded shipped default at
// assets/directorysync) — because the vaultseal config bridges consume
// them (the T-45 repobanish-floor precedent).
//
// The REST of the module (build_index, publish, the command cycle,
// the config trio) is T-49's own task.
package directorysync
