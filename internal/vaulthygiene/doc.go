// Package vaulthygiene ports core/vault_hygiene.py — v0.68.0 THE
// STRAY-NOTES DOOR (the shared pass), widened by v0.70.0 (THE
// LINKED-GHOST DOOR + THE APP-OWNED BODYLESS FINGERPRINT + THE ONE
// WALK), v0.71.0 (THE ONE-OPEN LAW's shared cache) and v0.73.0 (THE
// LOCK LAW's hygiene predicate).
//
// T-46 (SESSION-021) landed the whole module in hygiene.go:
//
//   - the classifiers: IsEmptyStub, IsAppOwnedBodyless,
//     IsLockedFrontmatter, HasNoteNameShape, NormalizeIdentity;
//   - the harvesters: HarvestLinkTargets (the [[wiki-link]] eyes) and
//     the one walk's head-identity harvest (stems, first H1,
//     frontmatter aliases);
//   - RunVaultHygiene — ONE walk, ONE bounded read per file (or the
//     shared notereads cache), the three fingerprints (shape /
//     name-twin / linked-ghost), the app-owned bodyless door, the lock
//     guard, the master-index scaffold removal, the legacy root
//     reports, the honest summary;
//   - HealMachineVaults (both machine vaults) + HygieneTotal.
//
// The reference's collect_real_note_identities is dead upstream and is
// not ported; its behavior lives in the one walk. The reference's
// try/except "pass skipped" wrapper is unreachable in Go (no path can
// abort mid-walk); the bounded direct read keeps the OSError door (an
// unreadable file is skipped, exactly as the reference's except
// OSError: continue) while the CACHE read is the house API (an
// unreadable file reads ” through it — a documented convergence; the
// reference's own tests never exercise the door).
package vaulthygiene
