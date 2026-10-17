package workerclient

// helpers.go — the URL normalizer (must byte-match the Worker's own),
// the vault-mirror entry builders, the hand-rolled uuid4 (no new
// dependencies), and the health GET.

import (
	"crypto/rand"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// NormalizeURL normalizes a URL for dedup — MUST match the Worker's
// normalizeUrl() (utils.js): strip whitespace + trailing .,); (all of
// them), twitter.com → x.com (every occurrence — see doc.go's
// divergence record), drop fragment, drop query, drop trailing slashes
// (a bare "scheme:" regains "//"), lowercase the domain ONLY.
func NormalizeURL(rawURL string) string {
	u := helpers.RstripChars(helpers.Strip(rawURL), ".,);")
	u = strings.ReplaceAll(u, "twitter.com", "x.com")
	if i := strings.IndexByte(u, '#'); i >= 0 {
		u = u[:i]
	}
	if i := strings.IndexByte(u, '?'); i >= 0 {
		u = u[:i]
	}
	u = strings.TrimRight(u, "/")
	if strings.HasSuffix(u, ":") {
		u += "//"
	}
	if i := strings.Index(u, "://"); i >= 0 {
		scheme := u[:i]
		rest := u[i+3:]
		if j := strings.IndexByte(rest, '/'); j >= 0 {
			domain := strings.ToLower(rest[:j])
			u = scheme + "://" + domain + rest[j:]
		} else {
			u = scheme + "://" + strings.ToLower(rest)
		}
	}
	return u
}

// BuildVaultMirrorEntries builds vault_mirror entries from the
// desktop's VaultIndex ({url: {path, category, title,
// first_processed, last_updated}} — insertion order preserved).
func BuildVaultMirrorEntries(vaultIndex any) []any {
	var entries []any
	if om, ok := vaultIndex.(*storage.OrderedMap); ok {
		for _, url := range om.Keys() {
			info, _ := om.Get(url)
			entries = append(entries, mirrorEntry(url, info))
		}
		return entries
	}
	if m, ok := vaultIndex.(map[string]any); ok {
		// Go maps are unordered; sorted keys keep the wire bytes
		// deterministic (documented divergence — Python dicts are
		// insertion-ordered by construction).
		keys := make([]string, 0, len(m))
		for k := range m {
			keys = append(keys, k)
		}
		sortStrings(keys)
		for _, url := range keys {
			entries = append(entries, mirrorEntry(url, m[url]))
		}
	}
	return entries
}

func mirrorEntry(url string, info any) *storage.OrderedMap {
	om := storage.NewOrderedMap()
	om.Set("url_normalized", url)
	om.Set("vault_path", pyStr(pyOr(dictValue(info, "path"), "")))
	om.Set("category", pyStr(pyOr(dictValue(info, "category"), "")))
	om.Set("title", pyStr(pyOr(dictValue(info, "title"), "")))
	om.Set("status", "in_vault")
	om.Set("first_processed_at", pyStr(pyOr(dictValue(info, "first_processed"), "")))
	om.Set("last_updated_at", pyStr(pyOr(dictValue(info, "last_updated"), nowISO())))
	return om
}

// DecommissionEventRow is the settled decommissioned row shape
// (store.CacheDB.GetAllDecommissioned's FailedRow: URL + reason).
type DecommissionEventRow struct {
	URL    string
	Reason string
}

// BuildDecommissionEvents builds the decommission_events list from the
// desktop's decommissioned rows. (The reference calls .get() on the
// (url, reason) tuples get_all_decommissioned returns — a latent
// AttributeError its caller's try/except swallows; the port maps the
// settled shape directly, per doc.go's reality-wins record.)
func BuildDecommissionEvents(rows []DecommissionEventRow) []any {
	var events []any
	for _, row := range rows {
		reason := row.Reason
		if reason == "" {
			reason = "404"
		}
		om := storage.NewOrderedMap()
		om.Set("url_normalized", row.URL)
		om.Set("reason", reason)
		om.Set("details", "")
		om.Set("decommissioned_at", nowISO())
		events = append(events, om)
	}
	return events
}

// nowISO mirrors _now_iso: datetime.now(timezone.utc).isoformat()
// (timespec='auto' — the microsecond fraction IFF nonzero).
func nowISO() string {
	t := time.Now().UTC()
	base := t.Format("2006-01-02T15:04:05")
	if us := t.Nanosecond() / 1000; us != 0 {
		return fmt.Sprintf("%s.%06d+00:00", base, us)
	}
	return base + "+00:00"
}

// newSyncIDFn is the sync_id seam (uuid.uuid4() in the reference —
// the idempotency tests pin it).
var newSyncIDFn = newSyncID

// newSyncID is uuid.uuid4() — hand-rolled (crypto/rand + the v4
// version/variant bits; zero new dependencies).
func newSyncID() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand failing is process-fatal territory; the UUID
		// shape stays honest even on the degraded path.
		return "00000000-0000-4000-8000-000000000000"
	}
	b[6] = (b[6] & 0x0f) | 0x40
	b[8] = (b[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// getJSON is the health GET: UA worn (the 1010 law), proxy bypassed,
// no auth, JSON-decoded (a parse failure surfaces as the error).
func getJSON(u string, timeoutS float64) (any, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", HTTPUserAgent)
	res, err := directClient(timeoutS).Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()
	raw, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return nil, rerr
	}
	return storage.DecodeJSONValue(raw)
}

func sortStrings(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
