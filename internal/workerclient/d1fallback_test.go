package workerclient

// d1fallback_test.go — THE D1-API FALLBACK LAWS (SESSION-081). Every
// law the d1fallback.go file carries rides a hermetic pinning test:
//   - THE DISABLED-WHEN-CREDS-MISSING LAW (no creds → no HTTP).
//   - THE DIAGNOSTIC LAW (d1FallbackMissing names WHICH key).
//   - THE ROW-SHAPING LAW (columns→dict + github_metadata parsed).
//   - THE ENVELOPE-FAILURE LAW (success:false → false,nil).
//   - THE QUERY-FAILURE LAW (result[0].success:false → false,nil).
//   - THE FALLBACK-WIRING LAW (GetPendingResult falls back to D1 when
//     the direct Worker leg fails — THE SESSION-081 LAW itself).
// The D1 REST API is faked by a loopback httptest server; the
// d1APIBase seam swaps to it. No network.

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// d1Envelope builds a canned Cloudflare D1 REST API response envelope
// (the SAME wire shape the live api.cloudflare.com returns): result[0].
// results is a LIST OF ROW-DICTS (each row an object keyed by column
// name — NOT a {columns,rows} pair). rows is the list of row-dicts.
func d1Envelope(rows []map[string]any, ok bool) map[string]any {
	rowAny := make([]any, len(rows))
	for i, r := range rows {
		rowAny[i] = r
	}
	return map[string]any{
		"result": []any{
			map[string]any{
				"results": rowAny,
				"success": ok,
			},
		},
		"success": ok,
		"errors":  []any{},
	}
}

// d1Server returns a loopback D1 REST API server that replies with the
// given canned envelope (or an envelope-level failure when ok=false +
// msg set). The handler captures the SQL it received.
func d1Server(t *testing.T, reply func(sql string) (map[string]any, int)) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var body struct {
			Sql string `json:"sql"`
		}
		_ = json.NewDecoder(r.Body).Decode(&body)
		env, code := reply(body.Sql)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(code)
		_ = json.NewEncoder(w).Encode(env)
	}))
	return srv
}

// d1Rig builds a Sync whose 3 D1 creds ride the loopback server.
func d1Rig(srv *httptest.Server) *Sync {
	cfg := map[string]any{
		"cloudflare_api_token":      "cfut_test_token",
		"cloudflare_account_id":     "acct_test",
		"cloudflare_d1_database_id": "db_test",
		"cloudflare_enabled":        true,
	}
	return NewSync(cfg)
}

// TestD1FallbackDisabledWhenCredsMissing: any of the 3 creds empty →
// (false, nil); NO HTTP leg runs (the law's own guard). Pinned by
// asserting against a closed listener (a connection-attempt would
// surface as a transport error in the debug log, not a skip).
func TestD1FallbackDisabledWhenCredsMissing(t *testing.T) {
	// Point the seam at a closed port so a stray HTTP leg would fail
	// loudly (the disabled law must short-circuit BEFORE the dial).
	withSeam(t, &d1APIBase, "http://127.0.0.1:1")
	cases := []struct {
		name string
		cfg  map[string]any
	}{
		{"api_token empty", map[string]any{
			"cloudflare_api_token":      "",
			"cloudflare_account_id":     "acct",
			"cloudflare_d1_database_id": "db",
			"cloudflare_enabled":        true,
		}},
		{"account_id empty", map[string]any{
			"cloudflare_api_token":      "tok",
			"cloudflare_account_id":     "",
			"cloudflare_d1_database_id": "db",
			"cloudflare_enabled":        true,
		}},
		{"database_id empty", map[string]any{
			"cloudflare_api_token":      "tok",
			"cloudflare_account_id":     "acct",
			"cloudflare_d1_database_id": "",
			"cloudflare_enabled":        true,
		}},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			s := NewSync(c.cfg)
			ok, pending := s.GetPendingViaD1()
			if ok || pending != nil {
				t.Fatalf("%s: GetPendingViaD1 = (%v, %v); want (false, nil)", c.name, ok, pending)
			}
		})
	}
}

// TestD1FallbackMissingDiagnostic: d1FallbackMissing names WHICH key
// is empty (the diagnostic law — the owner sees exactly what to fix).
func TestD1FallbackMissingDiagnostic(t *testing.T) {
	s := NewSync(map[string]any{"cloudflare_enabled": true})
	if got := s.d1FallbackMissing(); got == "" {
		t.Fatalf("d1FallbackMissing with all-empty creds = empty; want the three-key list")
	}
	for _, c := range []struct {
		cfg  map[string]any
		want string
	}{
		{
			map[string]any{
				"cloudflare_api_token":      "tok",
				"cloudflare_account_id":     "",
				"cloudflare_d1_database_id": "db",
				"cloudflare_enabled":        true,
			},
			"cloudflare_account_id is empty",
		},
		{
			map[string]any{
				"cloudflare_api_token":      "",
				"cloudflare_account_id":     "acct",
				"cloudflare_d1_database_id": "",
				"cloudflare_enabled":        true,
			},
			"cloudflare_api_token is empty; cloudflare_d1_database_id is empty",
		},
	} {
		s := NewSync(c.cfg)
		if got := s.d1FallbackMissing(); got != c.want {
			t.Fatalf("d1FallbackMissing = %q; want %q", got, c.want)
		}
	}
	// All set → empty string.
	full := NewSync(map[string]any{
		"cloudflare_api_token":      "tok",
		"cloudflare_account_id":     "acct",
		"cloudflare_d1_database_id": "db",
		"cloudflare_enabled":        true,
	})
	if got := full.d1FallbackMissing(); got != "" {
		t.Fatalf("d1FallbackMissing with all creds = %q; want empty", got)
	}
}

// TestD1FallbackShapesRowsAndParsesMetadata: a canned D1 envelope
// surfaces as a list of OrderedMaps with github_metadata parsed back
// to a nested object (the SAME-DECODER + ROW-SHAPING law). Integers
// ride as int64 (the /api/pending shape).
func TestD1FallbackShapesRowsAndParsesMetadata(t *testing.T) {
	srv := d1Server(t, func(sql string) (map[string]any, int) {
		return d1Envelope([]map[string]any{
			{"url_normalized": "https://example.com", "github_metadata": `{"stars":42,"description":"d"}`, "forward_count": 5, "ledger_id": 100},
			{"url_normalized": "https://foo.com", "github_metadata": "", "forward_count": 0, "ledger_id": 101},
			{"url_normalized": "https://bar.com", "github_metadata": nil, "forward_count": 3, "ledger_id": 102},
		}, true), 200
	})
	defer srv.Close()
	withSeam(t, &d1APIBase, srv.URL)
	s := d1Rig(srv)

	ok, pending := s.GetPendingViaD1()
	if !ok || len(pending) != 3 {
		t.Fatalf("GetPendingViaD1 = (%v, %d items); want (true, 3)", ok, len(pending))
	}

	// Row 0: github_metadata parses to a nested object; forward_count int64.
	r0 := pending[0].(*storage.OrderedMap)
	if v, _ := r0.Get("url_normalized"); v != "https://example.com" {
		t.Fatalf("row0 url_normalized = %v; want https://example.com", v)
	}
	fc, _ := r0.Get("forward_count")
	if fc != int64(5) {
		t.Fatalf("row0 forward_count = %v (%T); want int64(5)", fc, fc)
	}
	md, _ := r0.Get("github_metadata")
	om, isOM := md.(*storage.OrderedMap)
	if !isOM {
		t.Fatalf("row0 github_metadata = %T; want *OrderedMap (parsed)", md)
	}
	stars, _ := om.Get("stars")
	if stars != int64(42) {
		t.Fatalf("row0 github_metadata.stars = %v; want int64(42)", stars)
	}

	// Row 1: empty github_metadata → nil.
	r1 := pending[1].(*storage.OrderedMap)
	if md, _ := r1.Get("github_metadata"); md != nil {
		t.Fatalf("row1 github_metadata = %v; want nil (empty string → nil)", md)
	}

	// Row 2: NULL github_metadata → nil.
	r2 := pending[2].(*storage.OrderedMap)
	if md, _ := r2.Get("github_metadata"); md != nil {
		t.Fatalf("row2 github_metadata = %v; want nil (NULL → nil)", md)
	}
}

// TestD1FallbackEnvelopeFailure: success:false at the envelope level
// → (false, nil); the honest cause rides the debug log.
func TestD1FallbackEnvelopeFailure(t *testing.T) {
	srv := d1Server(t, func(sql string) (map[string]any, int) {
		return map[string]any{
			"result":  []any{},
			"success": false,
			"errors": []any{
				map[string]any{"message": "bad token"},
			},
		}, 200
	})
	defer srv.Close()
	withSeam(t, &d1APIBase, srv.URL)
	s := d1Rig(srv)
	ok, pending := s.GetPendingViaD1()
	if ok || pending != nil {
		t.Fatalf("envelope-failure: GetPendingViaD1 = (%v, %v); want (false, nil)", ok, pending)
	}
}

// TestD1FallbackQueryFailure: result[0].success:false (a SQL error)
// → (false, nil); the per-statement error rides the debug log.
func TestD1FallbackQueryFailure(t *testing.T) {
	srv := d1Server(t, func(sql string) (map[string]any, int) {
		return map[string]any{
			"result": []any{
				map[string]any{
					"results": map[string]any{"columns": []any{}, "rows": []any{}},
					"success": false,
					"errors": []any{
						map[string]any{"message": "no such table: ledger"},
					},
				},
			},
			"success": true,
			"errors":  []any{},
		}, 200
	})
	defer srv.Close()
	withSeam(t, &d1APIBase, srv.URL)
	s := d1Rig(srv)
	ok, pending := s.GetPendingViaD1()
	if ok || pending != nil {
		t.Fatalf("query-failure: GetPendingViaD1 = (%v, %v); want (false, nil)", ok, pending)
	}
}

// TestGetPendingResultFallsBackToD1: THE SESSION-081 LAW itself —
// when the direct Worker request fails (the worker_url is dead), the
// honest poll falls back to GetPendingViaD1 and returns ("ok", list).
// This is the whole point of the D1-API fallback.
func TestGetPendingResultFallsBackToD1(t *testing.T) {
	srv := d1Server(t, func(sql string) (map[string]any, int) {
		return d1Envelope([]map[string]any{
			{"url_normalized": "https://fallback.example", "github_metadata": `{"stars":7}`, "forward_count": 2, "ledger_id": 200},
		}, true), 200
	})
	defer srv.Close()
	withSeam(t, &d1APIBase, srv.URL)

	// The Worker URL points at a DEAD host (the direct leg fails).
	// The 3 D1 creds point at the loopback fake D1.
	cfg := map[string]any{
		"cloudflare_worker_url":     "http://127.0.0.1:1",
		"cloudflare_install_id":     "install",
		"cloudflare_shared_secret":  "secret",
		"cloudflare_enabled":        true,
		"cloudflare_api_token":      "cfut_test",
		"cloudflare_account_id":     "acct_test",
		"cloudflare_d1_database_id": "db_test",
	}
	s := NewSync(cfg)

	verdict, pending := s.GetPendingResult()
	if verdict != "ok" {
		t.Fatalf("GetPendingResult verdict = %q; want %q (the D1 fallback should have rescued the dead direct leg)", verdict, "ok")
	}
	if len(pending) != 1 {
		t.Fatalf("GetPendingResult pending len = %d; want 1 (the fallback's row)", len(pending))
	}
	r0 := pending[0].(*storage.OrderedMap)
	if v, _ := r0.Get("url_normalized"); v != "https://fallback.example" {
		t.Fatalf("fallback row0 url_normalized = %v; want https://fallback.example", v)
	}
}
