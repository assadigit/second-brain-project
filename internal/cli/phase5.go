package cli

// phase5.go — T-62: the dry-run report writer (cli_run.py
// _write_dryrun_report 215-236) + THE PHASE-5 CLEAR, re-specced for
// Option B — the bounded open item ADR-0004 Am.3 left for T-62,
// RESOLVED HERE with evidence (never improvised):
//
// THE RE-SPEC. The reference's Phase-5 CLEAR (cli_run.py 425-469)
// advanced last_processed_msg_id to the batch's max message id and
// marked the Telethon bot queue read — BOTH legs die with the MTProto
// client: ADR-0004 Am.3 (owner stamp "OPTION B CONFIRMED", rider R-0)
// drops read-acknowledge entirely and rides the Worker HTTP API only;
// spec-worker-contract §8.4 pins the mechanics — "Mirror-push is the
// ack (idempotent by sync_id): /api/vault_mirror/push remains the one
// and only way items leave the pending queue; zero Worker changes are
// required for the core flow".
//
// What survives, on the evidence:
//  1. THE ALL-CLEAR GATE (tracker.get_all_clear()) survives as the
//     honest-completion law: only a batch whose every GitHub link
//     verified may speak the Phase-5 CLEAR line (the ✓ verdict).
//     A not-verified batch speaks the ⏸️ verdict — the queue is NOT
//     acknowledged for those links.
//  2. THE QUEUE-CONSUMPTION MECHANICS are per-link by LEDGER STATE: a
//     verified link has a note → it is in the vault index → the
//     mirror push tells the Worker it is in_vault → it leaves
//     /api/pending; an unverified link has no note → it is absent from
//     the pushed mirror → it stays pending → the next --auto re-offers
//     it. The anti-repeat law is preserved by the ledger itself, not
//     by a message cursor (a strictly finer-grained honesty than the
//     reference's all-or-nothing msg-id advance).
//  3. THE MIRROR PUSH rides the batch-complete moment — the
//     reference's own intended pattern: CloudflareManager.on_batch_complete
//     (the processing_finished handler's end, per the integration
//     snippet's law) schedules the debounced _do_batch_push after
//     EVERY batch. REALITY-WINS RECORD: the pinned rev's MainWindow
//     keeps cf_manager = None — the manager ships as a library plus
//     the documented integration pattern, never live-wired; the
//     entries builder (build_vault_mirror_entries) expects the
//     {url: {path, category, title}} dict the app's VaultIndex class
//     never was (dict(instance) raises TypeError inside the manager's
//     own try/except). The port takes the intended pattern as its law
//     and implements the push directly: the entries from the vault
//     index's url→path map (the builder's .get() defaults fill the
//     rest), the decommission events from the cache, the sha256 index
//     hash over the sorted urls — the manager's own recipe, ported.
//     Best-effort and failure-isolated: a down Worker logs one line
//     and never fails the batch; a disabled Worker is silent (not an
//     event — the reference's own is_enabled() early-return).

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/assadigit/second-brain-project/internal/batch"
	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/dryrun"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
	"github.com/assadigit/second-brain-project/internal/workerclient"
)

// MirrorPushOutcome carries the honest result of the batch-complete
// vault-mirror push (the Phase-5 CLEAR's ack leg).
type MirrorPushOutcome struct {
	// Attempted is false when the Worker is not configured (the silent
	// law) — never an error, never spoken.
	Attempted bool
	OK        bool
	Entries   int
	Decomm    int
	Error     string
}

// defaultWorkerMirrorPush builds workerclient.Sync over the config and
// pushes (the disabled Worker is silent; the attempted push answers
// honestly).
func defaultWorkerMirrorPush(cfg *storage.OrderedMap, entries,
	decommissionEvents []any, vaultIndexHash string) (bool, bool, any) {
	s := workerclient.NewSync(cfg)
	if !s.IsEnabled() {
		return false, false, nil
	}
	ok, resp := s.PushVaultMirror(entries, decommissionEvents, vaultIndexHash)
	return true, ok, resp
}

// PushVaultMirrorAfterBatch is the batch-complete leg: build the
// entries from the vault index + the decommission events from the
// cache, hash the index (sha256 over '|'.join(sorted(urls)) — the
// manager's own recipe), push (idempotent by sync_id), and speak the
// honest outcome. A nil index or cache pushes nothing (an empty batch
// has no state to sync).
func (c *CLI) PushVaultMirrorAfterBatch(cfg *storage.OrderedMap,
	index *batch.VaultIndex, cache *store.CacheDB) *MirrorPushOutcome {
	if index == nil {
		return &MirrorPushOutcome{}
	}
	idxData := index.MirrorIndex()
	entries := workerclient.BuildVaultMirrorEntries(idxData)
	var decomm []any
	if cache != nil {
		rows := cache.GetAllDecommissioned()
		dr := make([]workerclient.DecommissionEventRow, 0, len(rows))
		for _, r := range rows {
			dr = append(dr, workerclient.DecommissionEventRow{
				URL: r.URL, Reason: r.Error})
		}
		decomm = workerclient.BuildDecommissionEvents(dr)
	}
	urls := idxData.Keys()
	sum := sha256.Sum256([]byte(strings.Join(urls, "|")))
	hash := hex.EncodeToString(sum[:])

	var attempted, ok bool
	var resp any
	if c.WorkerMirrorPush != nil {
		attempted, ok, resp = c.WorkerMirrorPush(cfg, entries, decomm, hash)
	} else {
		attempted, ok, resp = defaultWorkerMirrorPush(cfg, entries, decomm, hash)
	}
	out := &MirrorPushOutcome{
		Attempted: attempted, OK: ok,
		Entries: len(entries), Decomm: len(decomm),
	}
	if attempted {
		if ok {
			c.CliPrint(fmt.Sprintf(
				"☁️ Worker synced: %d vault entry(ies) and %d decommission event(s) pushed — the acknowledged links leave the pending queue.",
				out.Entries, out.Decomm), "info")
		} else {
			out.Error = storage.PyDumpsCompact(resp)
			c.CliPrint(fmt.Sprintf(
				"⚠️ Vault-mirror push failed: %s — the queue keeps its state (the links stay pending; the next --auto re-offers them).",
				out.Error), "warning")
		}
	}
	return out
}

// PhaseFiveClear is THE PHASE-5 CLEAR, Option-B re-spec: the
// batch-complete mirror push (the ack — every batch, best-effort) plus
// the all-clear verdict speech (gated exactly like the reference:
// success AND bot_source). The reference's cursor advance + mark-read
// legs are the accepted losses the re-spec replaces — see the file
// header's evidence chain and spec-cli.md's dated note.
func (c *CLI) PhaseFiveClear(cfg *storage.OrderedMap, botSource, success bool,
	tracker *batch.LinkTracker, index *batch.VaultIndex, cache *store.CacheDB) {

	// The ack rides the batch-complete moment — unconditional in the
	// caller's finish chain (the reference's on_batch_complete law).
	c.PushVaultMirrorAfterBatch(cfg, index, cache)

	// The verdict speech — the reference's `if success and bot_source`
	// gate.
	if !success || !botSource {
		return
	}
	if tracker == nil {
		c.CliPrint("⏸️ No link manifest for this batch (empty vault path?) — "+
			"the Worker queue is not acknowledged; dedup still protects "+
			"against duplicates.", "warning")
		return
	}
	if tracker.GetAllClear() {
		c.CliPrint("✓ Bot queue synced — all links verified (Phase 5 CLEAR; "+
			"the vault-mirror push is the ack).", "success")
		return
	}
	unverified := 0
	if tracker.Manifest != nil {
		for _, l := range tracker.Manifest.Links {
			if l.Type == "github" &&
				(l.Status == "failed" || l.Status == "processing" ||
					l.Status == "pending") {
				unverified++
			}
		}
	}
	c.CliPrint(fmt.Sprintf(
		"⏸️ %d link(s) not verified — they stay PENDING in the Worker queue "+
			"and are not acknowledged; the next --auto re-fetches and retries "+
			"them.", unverified), "warning")
}

// writeDryrunReport mirrors _write_dryrun_report (215-236): save the
// dry-run write log as Markdown OUTSIDE the vault (APP_DIR/reports/
// dry-runs/dry_run_<stamp>.md) so the owner can read exactly what a
// dry-run would have changed. A PLAIN open on purpose — this runs from
// the finished moment, which can fire a hair BEFORE the engine's
// defer turns the dry-run switch off; a gated write here could end up
// recorded instead of saved. Returns the report path, or "" on failure.
func (c *CLI) writeDryrunReport() string {
	outDir := filepath.Join(constants.AppDir(), "reports", "dry-runs")
	if err := os.MkdirAll(outDir, 0o755); err != nil {
		return ""
	}
	stamp := c.Now().Format("20060102-150405")
	path := filepath.Join(outDir, "dry_run_"+stamp+".md")
	body := dryrun.RenderReportMarkdown("GitCurator dry-run report")
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return ""
	}
	return path
}
