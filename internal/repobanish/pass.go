// pass.go — T-46: the completion of core/repo_banish.py's port —
// revive_banished_repos (475-526, THE ♻️ DOOR BACK) and
// run_repo_banishment_pass (529-659, the confirmation-gate
// orchestration). The gate's eyes, the burial and the note-finder
// landed with T-45 in repobanish.go.
package repobanish

import (
	"fmt"
	"strings"

	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/notestate"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/websitepipeline"
	"github.com/assadigit/second-brain-project/internal/websitepipeline/decomtable"
)

// RepoBanishGatePrefix mirrors REPO_BANISH_GATE_PREFIX: the log prefix
// the repos gate's count line carries (the twin of the websites gate's
// BANISH_GATE_PREFIX — its own prefix so the log tells which vault's
// door fired).
const RepoBanishGatePrefix = "🗑️ Repo vault scan:"

// RepoBanishConfirmResult mirrors the banish_confirm channel's answer:
// {'verdict': ..., 'report': callable} — the bare-string door is
// type-narrowed to the dict shape (the Telegram round-trip's own
// contract; the same narrowing the websites twin made).
type RepoBanishConfirmResult struct {
	Verdict string
	// Report is the confirmed ask's closing line (the count that
	// ACTUALLY left) — the Telegram round-trip's own edit.
	Report func(n int)
}

// RepoBanishChannelFn mirrors the channel the factory builds:
// banish_confirm(items, log) → the answer, or nil on failure (the
// reference's raising channel reads as a defer).
type RepoBanishChannelFn func(items []PendingRepoBanishment,
	log LogFn) *RepoBanishConfirmResult

// RepoBanishConfirmFactory mirrors banish_confirm_factory: builds the
// Telegram round-trip only when there is something to ask — the
// hermetic law: this module never touches the network itself. An error
// is the reference's raising factory (the warning line + the defer).
type RepoBanishConfirmFactory func() (RepoBanishChannelFn, error)

// RepoBanishPassReport mirrors run_repo_banishment_pass' dict:
// {'verdict', 'pending', 'banished', 'urls', 'revived'}.
type RepoBanishPassReport struct {
	Verdict  string
	Pending  int
	Banished int
	URLs     []string
	Revived  int
}

// ReviveBanishedRepos mirrors revive_banished_repos (v0.65.0 — THE ♻️
// DOOR BACK, the websites master table's revive twin, on the GitHub
// vault's own table). Every ♻-marked row in the vault's
// DECOMMISSIONED.md (decomtable.StatusIsRevived — ♻️ / revived /
// restored / un-decommissioned) revives its repo: the note_state
// dismissal is cleared (the §4.4 gate opens) and the settled ledger row
// is removed (fetched like new again). A revival is the owner's
// explicit hand — this pass runs BEFORE the confirmation gate,
// whatever the gate's verdict. Idempotent — a second run finds the
// same ♻️ row but nothing left to clear, and stays silent. Never
// raises; returns how many repos were actually revived (a DB row
// moved). The dismissal clear is skipped under dry-run (note_state is
// never written in a rehearsal); the unsettle rides the given cache
// (the shadow cache in a rehearsal) exactly as the reference gates it.
func ReviveBanishedRepos(cache *store.CacheDB, noteStateDB *notestate.DB,
	vaultPath string, log LogFn) int {
	if log == nil {
		log = noopLog
	}
	if vaultPath == "" || !isDir(vaultPath) {
		return 0
	}
	table := websitepipeline.ScanDecommissionTable(vaultPath)
	revived := 0
	for _, url := range sortedStringKeys(table) {
		status := table[url]
		if !decomtable.StatusIsRevived(status) {
			continue
		}
		if !IsRepoURL(url) {
			continue // a websites row — not this door's call
		}
		canonical := links.NormalizeURL(url)
		if canonical == "" {
			continue
		}
		var parts []string
		if noteStateDB != nil && !dryrun.IsEnabled() {
			if cleared, err := noteStateDB.ClearDismissed(
				notestate.VaultGithub, canonical); err == nil && cleared {
				parts = append(parts, "the dismissal is gone")
			}
		}
		if cache != nil && cache.UnsettleRepo(canonical) {
			parts = append(parts, "the settlement is gone")
		}
		if len(parts) > 0 {
			revived++
			log(fmt.Sprintf("♻️ %s: revived — %s, it will be fetched "+
				"like new", canonical, strings.Join(parts, " and ")), "info")
		}
	}
	return revived
}

// RunRepoBanishmentPass mirrors run_repo_banishment_pass (v0.65.0 —
// THE REPOS BANISHMENT PASS, one orchestration): (1) the ♻️ revive
// pass, the owner's explicit hand, ungated; (2) the detection scan —
// both doors, deduped, hand-written marks counted but never asked;
// (3) the count SPOKEN first, then the ask over the INJECTED channel
// factory (the hermetic law: this module never imports the network
// itself); (4) the verdict gates the burial: 'confirmed' (and 'auto' —
// the config banish_confirm: false opt-out) → BanishMarkedRepos;
// 'declined' / 'timeout' / 'defer' → the marks and the notes stay, the
// next run asks again; (5) THE TALLY, spoken every run (the zero run
// answers too — 0 is a number). Never fails loud; a broken scan is one
// warning line and a defer.
func RunRepoBanishmentPass(cache *store.CacheDB, noteStateDB *notestate.DB,
	vaultPath string, log LogFn, config map[string]any,
	banishConfirmFactory RepoBanishConfirmFactory) RepoBanishPassReport {
	if log == nil {
		log = noopLog
	}
	out := RepoBanishPassReport{Verdict: "auto"}
	if vaultPath == "" || !isDir(vaultPath) {
		return out
	}
	// 1 — the ♻️ door back, the owner's explicit hand, ungated.
	out.Revived = ReviveBanishedRepos(cache, noteStateDB, vaultPath, log)
	// 2 — the gate's eyes.
	gate := ScanPendingRepoBanishments(vaultPath, noteStateDB, log)
	items := gate.Items
	pending := len(items)
	out.Pending = pending
	if pending == 0 {
		return out // the cheap no-op — most runs
	}
	// 3 — the number FIRST, then the ask.
	log(fmt.Sprintf("%s %d repo note(s) marked for deletion (🗑️ / "+
		"delete / auto_delete) — asking you to confirm before anything is "+
		"removed", RepoBanishGatePrefix, pending), "info")
	verdict := "defer"
	var channel RepoBanishChannelFn
	if banishConfirmFactory != nil {
		var err error
		channel, err = banishConfirmFactory()
		if err != nil {
			log(fmt.Sprintf("⚠️ Repo banish gate channel unavailable: %v — "+
				"nothing deleted, the marks stay; I'll ask again next run",
				err), "warning")
			channel = nil
		}
	}
	if channel != nil {
		var res *RepoBanishConfirmResult
		func() {
			defer func() {
				if r := recover(); r != nil {
					log(fmt.Sprintf("⚠️ Repo banish confirmation failed: %v — "+
						"nothing deleted, the marks stay", r), "warning")
				}
			}()
			res = channel(items, log)
		}()
		if res != nil {
			verdict = res.Verdict
			if verdict == "" {
				verdict = "defer"
			}
		}
	} else if repoBanishConfirmOptOut(config) {
		verdict = "auto" // the owner's explicit opt-out
	} else {
		// no channel injected and no opt-out — the SAFE default: the
		// marks stay, the next run asks again.
		verdict = "defer"
		log(fmt.Sprintf("⚠️ %d marked repo note(s) found, but no "+
			"confirmation channel is reachable (the Telegram worker is not "+
			"paired/enabled) — nothing deleted; the marks stay, I'll ask "+
			"again next run (config \"banish_confirm\": false restores the "+
			"old auto-delete)", pending), "warning")
	}
	out.Verdict = verdict
	if verdict == "confirmed" {
		log(fmt.Sprintf("✅ You confirmed the deletion — the %d marked repo "+
			"note(s) leave the GitHub vault now; the repos are dismissed + "+
			"settled (never fetched, never counted, never retried again); "+
			"GitHub drops them on the next seal", pending), "info")
	} else if verdict == "declined" {
		log(fmt.Sprintf("👌 You kept them — the %d marked repo note(s) stay "+
			"in the vault; the marks stay too, I'll ask again next run",
			pending), "info")
	} else if verdict == "timeout" {
		log(fmt.Sprintf("⌛ No answer in time — the %d marked repo note(s) "+
			"stay; the marks stay, I'll ask again next run", pending), "info")
	}
	// 4 — the burial (the gate's hold: only confirmed / auto).
	if verdict == "auto" || verdict == "confirmed" {
		report := BanishMarkedRepos(cache, noteStateDB, vaultPath, log, items)
		out.URLs = report.URLs
		out.Banished = len(out.URLs)
	}
	// 5 — THE TALLY (the zero run answers too).
	if out.Banished > 0 {
		log(fmt.Sprintf("%s %d repo(s) removed this run and never fetched "+
			"again — you marked their notes for deletion (🗑️ / delete / "+
			"auto_delete); the notes rest in %s, the repos are dismissed + "+
			"settled, and ♻️ on the record row undoes any of them",
			RepoBanishTallyPrefix, out.Banished,
			websitepipeline.BanishQuarantineRelpath), "info")
	} else if pending > 0 && verdict != "auto" {
		why := map[string]string{
			"declined": "you said keep",
			"timeout":  "no answer in time",
			"defer":    "no confirmation channel",
		}[verdict]
		if why == "" {
			why = verdict
		}
		log(fmt.Sprintf("%s 0 repos removed this run — %d marked repo "+
			"note(s) KEPT (%s; the marks stay, I'll ask again next run)",
			RepoBanishTallyPrefix, pending, why), "info")
	}
	return out
}

// repoBanishConfirmOptOut mirrors the config door
// str(config.get('banish_confirm', True)).strip().lower() in
// ('false', '0', 'no', 'off') — a nil config (and a missing key)
// defaults True (no opt-out). str() is helpers.PyStr (str(0) == '0'
// fires the opt-out; str(0.0) == '0.0' does not — Python's own word).
func repoBanishConfirmOptOut(config map[string]any) bool {
	if config == nil {
		return false
	}
	v, ok := config["banish_confirm"]
	if !ok || v == nil {
		return false
	}
	s := helpers.PyStr(v)
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "false", "0", "no", "off":
		return true
	}
	return false
}
