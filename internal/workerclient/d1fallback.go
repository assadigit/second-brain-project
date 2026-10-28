package workerclient

// d1fallback.go — THE D1-API FALLBACK (the owner's SESSION-081
// directive): when the direct Worker request fails (the owner's ISP
// blocks workers.dev at the IP level — even the DoH tier can't reach a
// black-holed anycast), GetPendingViaD1() queries the D1 database
// directly via the Cloudflare REST API (api.cloudflare.com — a
// DIFFERENT edge the ISP does NOT block). The SQL mirrors the Worker's
// handleGetPending exactly: ledger LEFT JOIN vault_mirror, forgotten=0,
// mirror status NULL/'pending' (the items the desktop hasn't
// acknowledged yet), not decommissioned, not a blocked-domain dead
// letter, ORDER BY first_seen_at ASC, LIMIT 100. Read-only: no ack —
// the ack rides the direct path when the network recovers (the D1 REST
// API has no business writing the sync_state.desktop_last_poll side
// effect either; that's the Worker's own leg). The pending list
// returns in the SAME shape the Worker's /api/pending returns so the
// existing GetPendingResult consumer works unchanged.
//
// THE REACH-CLOUDFLARE LAW (carried): the D1 API call rides
// directClient — the SAME DoH-fallback dialer the sync channel uses —
// so a poisoned local resolver can't kill this leg either (the tier's
// answer for api.cloudflare.com, dialed directly with the original
// name's TLS). THE SAME-DECODER LAW: the envelope rides
// storage.DecodeJSONValue (the same decoder MakeRequest uses for the
// Worker response), so integral columns surface as int64 — the
// /api/pending shape the downstream consumers already handle.

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// d1APIBase is the Cloudflare REST API root — a DIFFERENT edge than
// workers.dev (the ISP's workers.dev IP block does not reach it). A
// var (not a const) so the hermetic suite swaps it to a loopback
// server (the seams law — every network leg rides an injectable var).
var d1APIBase = "https://api.cloudflare.com/client/v4/accounts"

// d1MaxBody caps the response read (a runaway result set must not hold
// the leg hostage; 100 rows × a generous README excerpt fits well
// within).
const d1MaxBody = 1 << 20 // 1 MiB

// d1FallbackConfig reads the three Cloudflare-API credentials from the
// live config (the Direct Pair UI's three optional fields; empty when
// the owner hasn't filled them — the fallback stays dormant).
func (s *Sync) d1FallbackConfig() (apiToken, accountID, databaseID string) {
	apiToken = pyStr(pyOr(cfgValue(s.config, "cloudflare_api_token"), ""))
	accountID = pyStr(pyOr(cfgValue(s.config, "cloudflare_account_id"), ""))
	databaseID = pyStr(pyOr(cfgValue(s.config, "cloudflare_d1_database_id"), ""))
	return
}

// d1FallbackMissing reports WHICH D1 credential is missing (the
// diagnostic law; empty when all three are set — the caller logs the
// reason, never a bare "fallback failed").
func (s *Sync) d1FallbackMissing() string {
	apiToken, accountID, databaseID := s.d1FallbackConfig()
	var missing []string
	if apiToken == "" {
		missing = append(missing, "cloudflare_api_token is empty")
	}
	if accountID == "" {
		missing = append(missing, "cloudflare_account_id is empty")
	}
	if databaseID == "" {
		missing = append(missing, "cloudflare_d1_database_id is empty")
	}
	if len(missing) == 0 {
		return ""
	}
	return strings.Join(missing, "; ")
}

// d1PendingSQL mirrors the Worker's handleGetPending exactly (the
// documented /api/pending contract): ledger LEFT JOIN vault_mirror,
// forgotten=0, mirror status NULL/'pending', not decommissioned, not
// a blocked-domain dead letter, ORDER BY first_seen_at ASC, LIMIT 100.
// The github_metadata column rides as JSON text in D1 (the Worker
// stores it that way; shapeD1Row parses it back to the nested object
// the /api/pending response carries).
const d1PendingSQL = `SELECT
  l.url_normalized, l.url_original, l.url_type, l.github_owner,
  l.github_repo, l.github_metadata, l.first_seen_at, l.forward_count,
  l.ledger_id
FROM ledger l
LEFT JOIN vault_mirror vm ON vm.url_normalized = l.url_normalized
WHERE l.forgotten = 0
  AND (vm.status IS NULL OR vm.status = 'pending')
  AND l.decommissioned = 0
  AND NOT EXISTS (
    SELECT 1 FROM dead_letters d
    WHERE d.url_normalized = l.url_normalized AND d.reason = 'blocked_domain'
  )
ORDER BY l.first_seen_at ASC
LIMIT 100`

// GetPendingViaD1 is THE D1-API FALLBACK: queries the D1 database
// directly via api.cloudflare.com when the direct Worker request fails
// (the owner's ISP blocks workers.dev at the IP level). Returns
// (true, pending) on success; (false, nil) on any failure (the caller
// keeps the direct leg's own error as the honest first cause). Read-
// only: no ack, no sync_state side effect — the ack rides the direct
// path when the network recovers.
func (s *Sync) GetPendingViaD1() (bool, []any) {
	ok, pending := s.d1RawQuery(d1PendingSQL)
	if ok {
		debugf("[CloudflareSync] d1 fallback ok: %d pending", len(pending))
	}
	return ok, pending
}

// d1RawQuery runs a read-only SQL statement against the D1 database
// via the Cloudflare REST API and shapes the result rows into the
// same dict-list shape the Worker's JSON responses carry. THE D1
// REST API's result[0].results is a LIST OF ROW-DICTS (each row an
// object keyed by column name — NOT a {columns,rows} pair); this
// rides storage.DecodeJSONValue so integral columns surface as
// int64 (the /api/pending shape). The github_metadata column rides
// as JSON text in D1; shapeD1Row parses it back to a nested object.
// Returns (true, rows) on success; (false, nil) on any failure (the
// honest cause rides the debug log). The GetPendingViaD1 leg rides
// this; the hermetic suite rides it too.
func (s *Sync) d1RawQuery(sqlText string) (bool, []any) {
	apiToken, accountID, databaseID := s.d1FallbackConfig()
	if apiToken == "" || accountID == "" || databaseID == "" {
		debugf("[CloudflareSync] d1 fallback skipped: %s", s.d1FallbackMissing())
		return false, nil
	}
	u := d1APIBase + "/" + accountID + "/d1/database/" + databaseID + "/query"
	reqBody, _ := json.Marshal(struct {
		Sql string `json:"sql"`
	}{Sql: sqlText})
	req, err := http.NewRequest("POST", u, bytes.NewReader(reqBody))
	if err != nil {
		debugf("[CloudflareSync] d1 fallback request build failed: %s", err.Error())
		return false, nil
	}
	req.Header.Set("Authorization", "Bearer "+apiToken)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", HTTPUserAgent)
	res, err := directClient(30.0).Do(req)
	if err != nil {
		debugf("[CloudflareSync] d1 fallback transport failed: %s", err.Error())
		return false, nil
	}
	defer res.Body.Close()
	raw, rerr := io.ReadAll(io.LimitReader(res.Body, d1MaxBody))
	if rerr != nil {
		debugf("[CloudflareSync] d1 fallback read failed: %s", rerr.Error())
		return false, nil
	}
	data, perr := storage.DecodeJSONValue(raw)
	if perr != nil {
		debugf("[CloudflareSync] d1 fallback parse failed: %s", perr.Error())
		return false, nil
	}
	if !pyTruthy(dictValue(data, "success")) {
		debugf("[CloudflareSync] d1 fallback envelope failed: %s",
			firstD1Error(data, "errors"))
		return false, nil
	}
	resultSlice, _ := dictValue(data, "result").([]any)
	if len(resultSlice) == 0 {
		debugf("[CloudflareSync] d1 fallback: empty result slice")
		return false, nil
	}
	first := resultSlice[0]
	if !pyTruthy(dictValue(first, "success")) {
		debugf("[CloudflareSync] d1 fallback query failed: %s",
			firstD1Error(first, "errors"))
		return false, nil
	}
	// THE D1 REST API's result[0].results is a LIST OF ROW-DICTS
	// (each row an object keyed by column name). DecodeJSONValue
	// surfaces each row as a *storage.OrderedMap; shapeD1Row parses
	// the github_metadata JSON-text column back to a nested object.
	rowAny, _ := dictValue(first, "results").([]any)
	pending := make([]any, 0, len(rowAny))
	for _, r := range rowAny {
		switch row := r.(type) {
		case *storage.OrderedMap:
			if md, found := row.Get("github_metadata"); found {
				row.Set("github_metadata", shapeD1Row("github_metadata", md))
			}
			pending = append(pending, row)
		case map[string]any:
			if md, found := row["github_metadata"]; found {
				row["github_metadata"] = shapeD1Row("github_metadata", md)
			}
			pending = append(pending, row)
		default:
			pending = append(pending, r)
		}
	}
	return true, pending
}

// shapeD1Row renders one D1 cell into the /api/pending dict shape. The
// github_metadata column rides as JSON text in D1 (the Worker stores
// it + the Worker re-parses it); the fallback parses it back to the
// nested object the /api/pending response carries. A NULL/empty
// metadata stays nil (the Worker's own shape — github_metadata: null).
// Every other column rides as-is (the SAME-DECODER LAW: integers
// surface as int64, the /api/pending shape).
func shapeD1Row(col string, v any) any {
	if col == "github_metadata" {
		switch t := v.(type) {
		case string:
			if t == "" {
				return nil
			}
			if parsed, perr := storage.DecodeJSONValue([]byte(t)); perr == nil {
				return parsed
			}
			return nil
		case nil:
			return nil
		}
	}
	return v
}

// D1FallbackStatus reports the D1-API fallback's arming state for the
// command layer's diagnostic (THE DIAGNOSTIC LAW — the owner sees
// exactly whether the fallback is armed + what to fix, never a bare
// "queue poll failed"). Pure config read (no HTTP). Used by the
// queue-check error rendering when the direct Worker leg fails.
func (s *Sync) D1FallbackStatus() string {
	apiToken, accountID, databaseID := s.d1FallbackConfig()
	if apiToken == "" || accountID == "" || databaseID == "" {
		return "the D1-API fallback is DORMANT (" + s.d1FallbackMissing() +
			") — arm it ONE of two ways: (a) open Settings → Direct Pair + " +
			"fill the 3 Cloudflare-API fields (API Token + Account ID + " +
			"Database ID), OR (b) drop a d1creds.json file (with the 3 creds) " +
			"into the GitCurator folder (next to GitCurator.exe); when the " +
			"direct Worker request fails (the ISP blocks workers.dev), the " +
			"fallback queries the D1 database directly via api.cloudflare.com"
	}
	return "the D1-API fallback is ARMED (the 3 Cloudflare-API creds are set) — " +
		"the direct Worker request failed + the fallback should have run; if you " +
		"still see this error, the D1 API call itself failed (verify the API " +
		"token has D1 read access + the account_id + database_id are correct)"
}

// firstD1Error surfaces the first error message in a D1 envelope's
// `errors` list (each entry {code, message}); "d1 api failure" when
// the list is empty or missing (the honest cause rides the debug log,
// never a bare "failed").
func firstD1Error(env any, key string) string {
	errs, _ := dictValue(env, key).([]any)
	if len(errs) == 0 {
		return "d1 api failure"
	}
	if msg := pyStr(pyOr(dictValue(errs[0], "message"), "")); msg != "" {
		return msg
	}
	return "d1 api failure"
}
