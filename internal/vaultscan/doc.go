// Package vaultscan ports core/vault_scan.py (v0.61.0 → v0.83.0) —
// THE VAULT SCAN: the LLM-assisted refiling engine with the
// deterministic "regex brain", weighed-evidence scoring, chunked LLM
// filing, validated plan building, and the confirm-then-apply
// machinery. The `kind` door runs the same grammar for the Websites
// vault ('websites') and the GitHub Repos vault ('repos').
//
// Spec: docs/port-blueprint/module-specs/spec-core-vault-scan.md;
// task T-45 (Phase 4), landed in SESSION-020.
//
// The module's files (the reference's own layout, kept):
//
//   - scan.go       — the radar: constants, the kind door, the
//     truthy/lock/curated predicates, the direct frontmatter parser
//     (vault_scan's OWN grammar — flow tags REPLACE, the strip chain
//     without a final strip — preserved beside note_reads' unified
//     parse exactly as the reference keeps both), the CPython
//     urlparse netloc/path shapes (the inventory's domain cell and
//     the consolidation keeper rank — the \t\r\n removal, the
//     case-preserved netloc, the bracket law, _splitparams), the
//     core-note law, the inventory walk, and the move candidates.
//   - duplicates.go — ONE NOTE PER SITE: the consolidation magnet
//     (locked notes invisible to it), the locked-canonicals deletion
//     door, the folder-tree/digest lines, the candidate reason, the
//     fold laws (fold/foldSegment/foldTag), the sanitizer, the disk
//     spelling, and the shared disk-verification resolver (THE DISK
//     SPELLING WINS).
//   - regexbrain.go — the deterministic passes: the declared-home
//     pass (websites Cat/Sub, repos the CATEGORY_FOLDERS key), the
//     sibling-domain pass (one agreed home), and the taxonomy pass
//     (every matched tag agrees on one home) — each with the settled
//     teaching set.
//   - weighing.go   — THE WEIGHED EVIDENCE: the word-token identity,
//     the folder census, the weighing folder set, the scale itself
//     (domain votes / the repos language agreement, shared tags, the
//     taxonomy vocabulary, title words, the folder name, the declared
//     echoes, the popularity prior capped at 0.05, the too-broad
//     penalty — banker's rounding), the attachment with the CLOSE
//     CALL verdict, and the digest's homes cell.
//   - llmpass.go    — the chunked filing pass: the shared tree, the
//     exact user message (the three laws at the point of decision,
//     the confidence ask, the homes paragraph), per-chunk failure
//     dropping only that chunk, and the honest review list (the
//     census law).
//   - validate.go   — the guest's proposal meets the vault's laws:
//     the fresh-eyes absorption (incl. intra-plan dedupe), the depth
//     law, the taken-set, the landable move (the _v bend pre-computed
//     at plan time, the row carrying 'as'; a self-move never
//     proposed), the low-confidence flag (the guest's own word or the
//     weighed-evidence override check).
//   - plan.go       — BuildScanPlan: the gate's own deletions (the
//     lock filter), the consolidations before any filing, the
//     retiring-never-ride law, the three-pass residue chain, the
//     weighed evidence, the LLM pass, the validation, and the review
//     list.
//   - apply.go      — the confirmed plan's hands: the consolidations
//     (links appended, files retired to .trash/consolidated, ledgers
//     re-pointed), the folders, the moves (lock re-read at apply,
//     the 'as' bend honored, verified landings — failures SPOKEN in
//     the ledger), the census stamps (moved/declined/reviewed), the
//     deletions through the banishment machinery's own doors
//     (skip_canonicals / the repos items parameter), and the repos
//     ledger re-point.
//
// THE POINTER-THREADING LAW (a Go-reality note, documented in the
// spec's shipped record): the reference threads Python dict
// REFERENCES from the candidates list through attach_weighed_homes
// into _validate_moves — the validator reads the homes the scale
// attached. Go struct copies would break that aliasing, so the chain
// threads []*NoteRow (MoveCandidates' rows are re-taken by path from
// the inventory).
//
// The stat seam (fsExists/fsIsDir/fsIsFile) exists for the
// _windows_eyes test law: the reference suite patches os.path's
// exists/isdir/isfile to replay the owner's case-insensitive Windows
// disk on POSIX (the two-notes loop's second face).
package vaultscan
