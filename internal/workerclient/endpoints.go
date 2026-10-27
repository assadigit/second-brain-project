package workerclient

// endpoints.go — every desktop endpoint: the honest poll, the mirror
// push (sync_id idempotency), the decommission/errors/backup pushes,
// the two confirmation-gate round-trips (banish + scan, v0.60/v0.61),
// verify, the decommissions pull, backfill, and health.

import (
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// GetPendingResult is THE HONEST POLL (v0.86.0): the same GET
// /api/pending leg GetPending rides, but the verdict distinguishes the
// three truths a display must never merge — ("ok", items) the live
// queue, ("disabled", nil) the worker is not configured, ("error",
// nil) the poll was attempted and failed. Read-only: no state
// mutates, no mirror pushes. Never raises.
func (s *Sync) GetPendingResult() (string, []any) {
	if !s.IsEnabled() {
		// THE DIAGNOSTIC LAW: tell the owner WHICH key is missing so
		// they know exactly what to fix (not just "not configured").
		debugf("isEnabled: disabled — missing: %s", s.DisabledReason())
		return "disabled", nil
	}
	var r reply
	func() {
		defer func() {
			if recover() != nil {
				r = reply{0, errMap("poll panicked")}
			}
		}()
		s.mu.lock()
		defer s.mu.unlock()
		s.lastPoll = nowSecondsFn()
		r = s.MakeRequest("GET", "/api/pending", nil)
	}()
	if r.status == 200 && pyTruthy(dictValue(r.data, "success")) {
		pending := dictValue(r.data, "pending")
		if list, ok := pending.([]any); ok {
			return "ok", list
		}
		return "ok", nil
	}
	debugf("[CloudflareSync] get_pending failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return "error", nil
}

// GetPending fetches the pending link dicts (the byte-compat delegate
// riding the honest poll).
func (s *Sync) GetPending() []any {
	_, pending := s.GetPendingResult()
	return pending
}

// ShouldPoll reports whether enough time has passed since the last
// poll (cloudflare_poll_interval, default 300 s).
func (s *Sync) ShouldPoll() bool {
	interval := pyIntDefault(cfgValue(s.config, "cloudflare_poll_interval"), 300)
	return (nowSecondsFn() - s.lastPoll) >= float64(interval)
}

// PushVaultMirror pushes vault_mirror entries to the Worker
// (idempotent via sync_id — the Worker's sync_processed:<sync_id>
// state). entries: [{url_normalized, vault_path, category, title,
// status, ...}]; decommissionEvents: [{url_normalized, reason,
// details}]; vaultIndexHash: the local index sha256. Returns
// (success, response).
func (s *Sync) PushVaultMirror(entries, decommissionEvents []any,
	vaultIndexHash string) (bool, any) {
	if !s.IsEnabled() {
		return false, errMap("Not enabled")
	}
	syncID := newSyncIDFn()
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("sync_id", syncID)
	body.Set("vault_index_hash", pyOr(vaultIndexHash, ""))
	body.Set("entries", pyOr(entries, []any{}))
	body.Set("decommission_events", pyOr(decommissionEvents, []any{}))

	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("POST", "/api/vault_mirror/push", body)
	}()
	if successLaw(r) {
		return true, r.data
	}
	debugf("[CloudflareSync] push_vault_mirror failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return false, r.data
}

// PushDecommission pushes a real-time decommission event (e.g., a
// detected 404).
func (s *Sync) PushDecommission(urlNormalized, reason, details string) bool {
	if !s.IsEnabled() {
		return false
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("url_normalized", urlNormalized)
	body.Set("reason", reason)
	body.Set("details", details)
	r := s.MakeRequest("POST", "/api/decommission", body)
	if successLaw(r) {
		return true
	}
	debugf("[CloudflareSync] push_decommission failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return false
}

// ProposeBanish asks the owner to confirm the pending deletions
// (v0.60.0): the Worker sends the Telegram message with the two
// buttons. Returns (id, "") on success, (nil, reason) otherwise.
func (s *Sync) ProposeBanish(count int, items []any,
	timeoutS float64) (*string, string) {
	if !s.IsEnabled() {
		return nil, "not enabled"
	}
	if timeoutS == 0 {
		timeoutS = 300.0
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("count", count)
	body.Set("items", pyOr(items, []any{}))
	body.Set("timeout_s", timeoutS)
	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("POST", "/api/banish/propose", body)
	}()
	if successLaw(r) && pyTruthy(dictValue(r.data, "id")) {
		id := pyStr(dictValue(r.data, "id"))
		return &id, ""
	}
	return nil, pyStr(pyOr(dictValue(r.data, "error"), "HTTP "+itoa(r.status)))
}

// BanishStatus polls the ask's status ('pending' / 'confirmed' /
// 'declined' / 'timeout' / 'superseded'); empty when the Worker could
// not answer (the caller treats it as still-pending and keeps polling
// until its own deadline).
func (s *Sync) BanishStatus(confirmID string) string {
	if !s.IsEnabled() || confirmID == "" {
		return ""
	}
	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("GET", "/api/banish/status?id="+confirmID, nil)
	}()
	if r.status == 200 && pyTruthy(dictValue(r.data, "success")) {
		return pyStr(pyOr(dictValue(r.data, "status"), "pending"))
	}
	return ""
}

// ReportBanishResult closes the ask: outcome 'deleted' (the confirmed
// run's final count) or 'timeout'. Best-effort: a failed report never
// fails the run.
func (s *Sync) ReportBanishResult(confirmID, outcome string, deleted int) bool {
	if !s.IsEnabled() || confirmID == "" {
		return false
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("id", confirmID)
	body.Set("outcome", outcome)
	body.Set("deleted", deleted)
	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("POST", "/api/banish/result", body)
	}()
	return successLaw(r)
}

// ProposeScan asks the owner to confirm the VAULT SCAN plan (v0.61.0 —
// the banish trio's twin). Returns (id, "") or (nil, reason).
func (s *Sync) ProposeScan(deletions, moves, newFolders int, items []any,
	summary string, timeoutS float64) (*string, string) {
	if !s.IsEnabled() {
		return nil, "not enabled"
	}
	if timeoutS == 0 {
		timeoutS = 300.0
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("deletions", deletions)
	body.Set("moves", moves)
	body.Set("new_folders", newFolders)
	body.Set("items", pyOr(items, []any{}))
	body.Set("summary", truncateRunes(pyStr(pyOr(summary, "")), 400))
	body.Set("timeout_s", timeoutS)
	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("POST", "/api/scan/propose", body)
	}()
	if successLaw(r) && pyTruthy(dictValue(r.data, "id")) {
		id := pyStr(dictValue(r.data, "id"))
		return &id, ""
	}
	return nil, pyStr(pyOr(dictValue(r.data, "error"), "HTTP "+itoa(r.status)))
}

// ScanStatus polls the scan ask's status; empty when the Worker could
// not answer.
func (s *Sync) ScanStatus(confirmID string) string {
	if !s.IsEnabled() || confirmID == "" {
		return ""
	}
	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("GET", "/api/scan/status?id="+confirmID, nil)
	}()
	if r.status == 200 && pyTruthy(dictValue(r.data, "success")) {
		return pyStr(pyOr(dictValue(r.data, "status"), "pending"))
	}
	return ""
}

// ReportScanResult closes the scan ask: outcome 'applied' (the counts
// that actually landed) or 'timeout'. Best-effort.
func (s *Sync) ReportScanResult(confirmID, outcome string,
	applied, moved, folders int) bool {
	if !s.IsEnabled() || confirmID == "" {
		return false
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("id", confirmID)
	body.Set("outcome", outcome)
	body.Set("applied", applied)
	body.Set("moved", moved)
	body.Set("folders", folders)
	var r reply
	func() {
		s.mu.lock()
		defer s.mu.unlock()
		r = s.MakeRequest("POST", "/api/scan/result", body)
	}()
	return successLaw(r)
}

// PushErrors pushes desktop errors to the Worker
// ([{severity, error_code, message, details, occurred_at}]).
func (s *Sync) PushErrors(errors []any) bool {
	if !s.IsEnabled() {
		return false
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("errors", pyOr(errors, []any{}))
	r := s.MakeRequest("POST", "/api/errors", body)
	if successLaw(r) {
		return true
	}
	debugf("[CloudflareSync] push_errors failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return false
}

// PushBackupStatus pushes GDrive backup metadata to the Worker
// (install_id first, then the snapshot's own keys in order).
func (s *Sync) PushBackupStatus(backupData *storage.OrderedMap) bool {
	if !s.IsEnabled() {
		return false
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	if backupData != nil {
		for _, k := range backupData.Keys() {
			v, _ := backupData.Get(k)
			body.Set(k, v)
		}
	}
	r := s.MakeRequest("POST", "/api/gdrive/backup_status", body)
	if successLaw(r) {
		return true
	}
	debugf("[CloudflareSync] push_backup_status failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return false
}

// Verify gets the full reconciliation data from the Worker (nil on
// failure).
func (s *Sync) Verify() *storage.OrderedMap {
	if !s.IsEnabled() {
		return nil
	}
	r := s.MakeRequest("GET", "/api/verify", nil)
	if successLaw(r) {
		if om, ok := r.data.(*storage.OrderedMap); ok {
			return om
		}
		return nil
	}
	debugf("[CloudflareSync] verify failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return nil
}

// GetDecommissions pulls bot-originated decommission events since the
// given ISO instant ([] on failure — the reference's list law).
func (s *Sync) GetDecommissions(since string) []any {
	if !s.IsEnabled() {
		return nil
	}
	if since == "" {
		since = "1970-01-01T00:00:00Z"
	}
	r := s.MakeRequest("GET", "/api/decommissions?since="+since, nil)
	if r.status == 200 && pyTruthy(dictValue(r.data, "success")) {
		if list, ok := dictValue(r.data, "decommissions").([]any); ok {
			return list
		}
		return nil
	}
	debugf("[CloudflareSync] get_decommissions failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return nil
}

// BackfillChunk pushes a backfill chunk to the Worker (idempotent per
// URL). Returns (success, response with progress).
func (s *Sync) BackfillChunk(entries []any, chunkNum,
	totalChunks int) (bool, any) {
	if !s.IsEnabled() {
		return false, errMap("Not enabled")
	}
	body := storage.NewOrderedMap()
	body.Set("install_id", s.InstallID())
	body.Set("entries", pyOr(entries, []any{}))
	body.Set("chunk_num", chunkNum)
	body.Set("total_chunks", totalChunks)
	r := s.MakeRequest("POST", "/api/backfill", body)
	if successLaw(r) {
		return true, r.data
	}
	debugf("[CloudflareSync] backfill_chunk failed: %d %s", r.status,
		storage.PyDumpsCompact(r.data))
	return false, r.data
}

// HealthCheck checks Worker health (no auth required; the 1010 UA worn
// all the same). Nil on failure.
func (s *Sync) HealthCheck() *storage.OrderedMap {
	u := s.WorkerURL() + "/health"
	r, err := getJSON(u, 10.0)
	if err != nil {
		debugf("[CloudflareSync] health_check failed: %s", err.Error())
		return nil
	}
	if om, ok := r.(*storage.OrderedMap); ok {
		return om
	}
	return nil
}

// nowSecondsFn is the wall-clock seam (time.time() in the reference —
// the should_poll tests pin it).
var nowSecondsFn = nowSeconds

func nowSeconds() float64 {
	return float64(time.Now().UnixNano()) / 1e9
}

// pyIntDefault mirrors int(value) with the fallback: bool → 1/0,
// float truncates, decimal strings parse, anything else → def.
func pyIntDefault(v any, def int) int {
	switch t := v.(type) {
	case nil:
		return def
	case bool:
		if t {
			return 1
		}
		return 0
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		if n, err := strconv.Atoi(strings.TrimSpace(t)); err == nil {
			return n
		}
		return def
	default:
		return def
	}
}

// truncateRunes is Python's s[:n].
func truncateRunes(s string, n int) string {
	rs := []rune(s)
	if len(rs) <= n {
		return s
	}
	return string(rs[:n])
}
