// Package directorysync ports cloud/directory_sync.py — THE WEBSITE
// DIRECTORY'S PUBLISH LEG + the Phase-2 admin-command cycle: the
// one-way content flow Obsidian → GitHub → Website.
//
// T-40 landed the banish gate floor (banishgate.go — load_banish_list,
// DEFAULT_SHARED_HOSTS, is_banished: the read side of
// Websites/_system/banished.json that WebsitePipeline probes before
// any link is fetched).
//
// T-47 (SESSION-022) landed THE LOCAL-ONLY BELT floor (belt.go) with
// vaultseal — local_only_note_patterns + private_exclusion_patterns +
// their pure helpers + load_directory_config (the embedded shipped
// default at assets/directorysync) — because the vaultseal config
// bridges consume them (the T-45 repobanish-floor precedent).
//
// T-49 (SESSION-024) completed the module: the note-section parser
// (noteparse.go — parse_note, the drawer's grammar with the JS twin
// kept identical on purpose), the config trio (systemfiles.go —
// ensure_system_files' never-overwrite law + load_synonyms with the
// embedded shipped default), build_index (index.go — THE VAULT
// MIRROR LAW: the published category tree IS the vault's folder tree;
// the folder-name translation; the folded private belt; the banish
// wall), the publish leg + the read-back door's four truths +
// run_sync/publish_after_batch (publish.go), the admin command cycle
// (commands.go — fetch/ack, the idempotent vault appliers, the
// trash-not-delete law, push_mirror_repo, run_command_cycle), and the
// standalone CLI (cli.go, incl. the PBKDF2 owner-setup helper). The
// drift seal's publish door (vaultseal.DriftSealDeps.PublishAfterBatch)
// wires to PublishAfterBatch.
package directorysync
