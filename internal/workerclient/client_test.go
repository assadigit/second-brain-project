package workerclient

// client_test.go — the client against the loopback rig: the honest
// poll's three truths, pair error mapping, the full endpoint
// round-trips, THE 1010 LAW proven from the wire, the HMAC window +
// install + signature rejections, the mirror idempotency, the
// proxy-bypass law, and the body-parity pin (json.dumps bytes).

import (
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// THE 1010 LAW: the channel's own name, pinned (kept in lockstep with
// the reference's VERSION at release).
func TestHTTPUserAgentConstant(t *testing.T) {
	want := "GitCurator/0.91.0 (+https://github.com/assadigit/GitCurator)"
	if HTTPUserAgent != want {
		t.Fatalf("UA = %q, want %q", HTTPUserAgent, want)
	}
}

func TestIsEnabledIsPairedGates(t *testing.T) {
	base := map[string]any{
		"cloudflare_worker_url":    "https://bot.example",
		"cloudflare_install_id":    "id",
		"cloudflare_shared_secret": "sec",
		"cloudflare_enabled":       true,
	}
	s := NewSync(base)
	if !s.IsEnabled() || !s.IsPaired() {
		t.Fatal("fully configured: enabled+paired expected")
	}
	for key, val := range map[string]any{
		"cloudflare_enabled":       false,
		"cloudflare_worker_url":    "",
		"cloudflare_install_id":    "",
		"cloudflare_shared_secret": "",
	} {
		cfg := map[string]any{}
		for k, v := range base {
			cfg[k] = v
		}
		cfg[key] = val
		if NewSync(cfg).IsEnabled() {
			t.Errorf("IsEnabled true with %s=%v", key, val)
		}
	}
	// paired ignores the switch
	cfg := map[string]any{"cloudflare_install_id": "id",
		"cloudflare_shared_secret": "sec", "cloudflare_enabled": false}
	if !NewSync(cfg).IsPaired() {
		t.Error("IsPaired must ignore the master switch")
	}
	if NewSync(cfg).IsEnabled() {
		t.Error("IsEnabled must honor the master switch")
	}
}

func TestPairSuccessWritesTheLiveConfig(t *testing.T) {
	fw := newFakeWorker(t)
	cfg := map[string]any{"cloudflare_worker_url": fw.url()}
	s := NewSync(cfg)
	ok, msg := s.Pair("ABC12345")
	if !ok {
		t.Fatalf("pair failed: %s", msg)
	}
	if msg != "Paired successfully!" {
		t.Errorf("message: %q", msg)
	}
	if cfg["cloudflare_install_id"] != rigInstall {
		t.Errorf("install_id = %v", cfg["cloudflare_install_id"])
	}
	if cfg["cloudflare_shared_secret"] != rigSecret {
		t.Errorf("shared_secret = %v", cfg["cloudflare_shared_secret"])
	}
	if cfg["cloudflare_enabled"] != true {
		t.Errorf("enabled = %v", cfg["cloudflare_enabled"])
	}
	// the pair leg wore the 1010 UA (the rig's server-side law)
	if r := fw.rejected(); len(r) != 0 {
		t.Fatalf("rejections: %+v", r)
	}
}

func TestPairNoURLAdvice(t *testing.T) {
	ok, msg := NewSync(map[string]any{}).Pair("X")
	if ok || msg != "No Worker URL configured. Enter it in the "+
		"Cloudflare tab first." {
		t.Fatalf("ok=%v msg=%q", ok, msg)
	}
}

func TestPair401RunsPairAgainAdvice(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(map[string]any{"cloudflare_worker_url": fw.url()})
	ok, msg := s.Pair("WRONGCODE")
	if ok {
		t.Fatal("wrong code paired")
	}
	if !strings.Contains(msg, "Invalid or expired pairing code (401)") {
		t.Errorf("message: %q", msg)
	}
	if !strings.Contains(msg, "Run /pair again in Telegram") {
		t.Errorf("message: %q", msg)
	}
}

func TestPairUnreachableAdvice(t *testing.T) {
	s := NewSync(map[string]any{
		"cloudflare_worker_url": deadWorkerURL(t)})
	ok, msg := s.Pair("X")
	if ok {
		t.Fatal("dead worker paired")
	}
	if !strings.Contains(msg, "Cannot reach Worker at ") {
		t.Errorf("message: %q", msg)
	}
}

func TestHonestPollThreeTruths(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))

	// truth 1: ok + the items
	row := storage.NewOrderedMap()
	row.Set("url_normalized", "https://example.com/x")
	row.Set("url_type", "website")
	fw.addPending(row)
	state, items := s.GetPendingResult()
	if state != "ok" {
		t.Fatalf("state = %s (%+v)", state, items)
	}
	if len(items) != 1 {
		t.Fatalf("items = %d", len(items))
	}
	if got := pyStr(dictValue(items[0], "url_normalized")); got != "https://example.com/x" {
		t.Errorf("item url = %q", got)
	}

	// truth 2: disabled — not enabled
	s2 := NewSync(map[string]any{
		"cloudflare_worker_url":    fw.url(),
		"cloudflare_install_id":    rigInstall,
		"cloudflare_shared_secret": rigSecret,
		"cloudflare_enabled":       false})
	state2, items2 := s2.GetPendingResult()
	if state2 != "disabled" || items2 != nil {
		t.Fatalf("state = %s items = %v", state2, items2)
	}

	// truth 3: error — attempted and failed (the dead worker)
	s3 := NewSync(map[string]any{
		"cloudflare_worker_url":    deadWorkerURL(t),
		"cloudflare_install_id":    rigInstall,
		"cloudflare_shared_secret": rigSecret,
		"cloudflare_enabled":       true})
	state3, items3 := s3.GetPendingResult()
	if state3 != "error" || items3 != nil {
		t.Fatalf("state = %s items = %v", state3, items3)
	}
}

func TestGetPendingDelegateAndShouldPoll(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))
	row := storage.NewOrderedMap()
	row.Set("url_normalized", "https://a.test/")
	fw.addPending(row)
	if got := s.GetPending(); len(got) != 1 {
		t.Fatalf("GetPending = %v", got)
	}
	if s.ShouldPoll() {
		t.Error("just polled: should_poll must be false")
	}
	withSeam(t, &nowSecondsFn, func() float64 { return nowSeconds() + 301 })
	if !s.ShouldPoll() {
		t.Error("301s later: should_poll must be true")
	}
	// the interval key is honored
	cfg := rigConfig(fw)
	cfg["cloudflare_poll_interval"] = 60
	s2 := NewSync(cfg)
	withSeam(t, &nowSecondsFn, func() float64 { return nowSeconds() })
	_, _ = s2.GetPendingResult()
	withSeam(t, &nowSecondsFn, func() float64 { return nowSeconds() + 61 })
	if !s2.ShouldPoll() {
		t.Error("61s later with a 60s interval: should_poll must be true")
	}
}

// THE 1010 LAW + THE HMAC LAW, proven from the wire: every leg the
// rig saw was HMAC-valid and wore the app's own name.
func TestEveryLegHMACValidAndUAWorn(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))
	fw.addPending(storage.NewOrderedMap())

	_, _ = s.GetPendingResult()
	_, _ = s.PushVaultMirror([]any{mirrorEntryFor("https://x.test/", "a/b.md")},
		nil, "hash1")
	_ = s.PushDecommission("https://x.test/", "404", "HTTP 404")
	_ = s.PushErrors([]any{map[string]any{"severity": "WARN"}})
	_ = s.PushBackupStatus(backupSample())
	_ = s.Verify()
	_ = s.GetDecommissions("")
	_, _ = s.BackfillChunk([]any{}, 0, 1)
	_, _ = s.ProposeBanish(2, []any{}, 300)
	_ = s.BanishStatus("ban-1")
	_ = s.ReportBanishResult("ban-1", "deleted", 2)
	_, _ = s.ProposeScan(1, 2, 3, []any{}, "summary", 300)
	_ = s.ScanStatus("scan-2")
	_ = s.ReportScanResult("scan-2", "applied", 1, 2, 3)
	_ = s.HealthCheck()

	if r := fw.rejected(); len(r) != 0 {
		t.Fatalf("the rig refused legs: %+v", r)
	}
	captured := fw.captured()
	if len(captured) < 15 {
		t.Fatalf("legs recorded = %d", len(captured))
	}
	for i, req := range captured {
		if req.UA != HTTPUserAgent {
			t.Errorf("leg %d (%s %s) UA = %q", i, req.Method, req.Path, req.UA)
		}
	}
	// health wears the UA too but carries no HMAC; every API leg does
	for i, req := range captured {
		if strings.HasPrefix(req.Path, "/api/") && !req.Authed &&
			req.Path != "/api/pair" {
			t.Errorf("leg %d (%s %s) not HMAC-authed", i, req.Method, req.Path)
		}
	}
}

// THE BODY-PARITY PIN: the wire bodies are json.dumps bytes — the GETs
// carry "{}" and the pushes carry the ", "/": " separators with the
// reference's key order.
func TestBodyParityOnTheWire(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))
	_, _ = s.GetPendingResult()
	_ = s.PushDecommission("https://Example.COM/Path/", "404", "HTTP 404")
	withSeam(t, &newSyncIDFn, func() string {
		return "0f0e0d0c-0b0a-4903-8802-1f2e3d4c5b6a"
	})
	_, _ = s.PushVaultMirror(
		[]any{mirrorEntryFor("https://example.com/x", "a/b.md")}, nil, "")

	captured := fw.captured()
	var getPending, decomm, mirror rigReq
	for _, req := range captured {
		switch {
		case req.Path == "/api/pending":
			getPending = req
		case req.Path == "/api/decommission":
			decomm = req
		case req.Path == "/api/vault_mirror/push":
			mirror = req
		}
	}
	if getPending.Body != "{}" {
		t.Errorf("pending GET body = %q, want {}", getPending.Body)
	}
	wantDecomm := `{"install_id": "` + rigInstall +
		`", "url_normalized": "https://Example.COM/Path/", ` +
		`"reason": "404", "details": "HTTP 404"}`
	if decomm.Body != wantDecomm {
		t.Errorf("decommission body:\n  got  %s\n  want %s", decomm.Body, wantDecomm)
	}
	wantMirror := `{"install_id": "` + rigInstall +
		`", "sync_id": "0f0e0d0c-0b0a-4903-8802-1f2e3d4c5b6a", ` +
		`"vault_index_hash": "", "entries": [{"url_normalized": ` +
		`"https://example.com/x", "vault_path": "a/b.md", "category": "", ` +
		`"title": "", "status": "in_vault", "first_processed_at": "", ` +
		`"last_updated_at": ""}], "decommission_events": []}`
	if mirror.Body != wantMirror {
		t.Errorf("mirror body:\n  got  %s\n  want %s", mirror.Body, wantMirror)
	}
}

func TestMirrorPushIdempotency(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))
	withSeam(t, &newSyncIDFn, func() string {
		return "fixed-sync-id-0001"
	})
	ok1, resp1 := s.PushVaultMirror([]any{}, nil, "h")
	if !ok1 {
		t.Fatalf("first push failed: %+v", resp1)
	}
	if pyTruthy(dictValue(resp1, "idempotent")) {
		t.Error("first push must not be idempotent")
	}
	// the Worker's sync_processed:<sync_id> state: the SAME sync_id
	// re-pushed is a no-op acknowledged as idempotent
	ok2, resp2 := s.PushVaultMirror([]any{}, nil, "h")
	if !ok2 {
		t.Fatalf("second push failed: %+v", resp2)
	}
	if !pyTruthy(dictValue(resp2, "idempotent")) {
		t.Error("second push with the same sync_id must be idempotent")
	}
	if n := pyIntDefault(dictValue(resp2, "vault_mirror_total"), 0); n != 1 {
		t.Errorf("vault_mirror_total = %d, want 1 (the sync_id is the unit)", n)
	}
}

func TestMirrorPushNotEnabled(t *testing.T) {
	s := NewSync(map[string]any{})
	ok, resp := s.PushVaultMirror(nil, nil, "")
	if ok {
		t.Fatal("pushed while disabled")
	}
	if pyStr(dictValue(resp, "error")) != "Not enabled" {
		t.Errorf("error = %v", dictValue(resp, "error"))
	}
}

func TestTimestampWindowRejected(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))
	// six minutes stale — outside the Worker's ±5-minute window
	withSeam(t, &nowMS, func() int64 {
		return time.Now().Add(-6 * time.Minute).UnixMilli()
	})
	state, _ := s.GetPendingResult()
	if state != "error" {
		t.Fatalf("state = %s, want error", state)
	}
	rej := fw.rejected()
	if len(rej) != 1 || rej[0].Why != "Timestamp out of range" {
		t.Fatalf("rejections: %+v", rej)
	}
}

func TestUnknownInstallRejected(t *testing.T) {
	fw := newFakeWorker(t)
	cfg := rigConfig(fw)
	cfg["cloudflare_install_id"] = "99999999-9999-9999-9999-999999999999"
	s := NewSync(cfg)
	state, _ := s.GetPendingResult()
	if state != "error" {
		t.Fatalf("state = %s", state)
	}
	rej := fw.rejected()
	if len(rej) != 1 || rej[0].Why != "Unknown install" {
		t.Fatalf("rejections: %+v", rej)
	}
}

func TestInvalidSignatureRejected(t *testing.T) {
	fw := newFakeWorker(t)
	cfg := rigConfig(fw)
	// the client signs with a DIFFERENT secret than the rig knows
	cfg["cloudflare_shared_secret"] = strings.Repeat("b", 64)
	s := NewSync(cfg)
	state, _ := s.GetPendingResult()
	if state != "error" {
		t.Fatalf("state = %s", state)
	}
	rej := fw.rejected()
	if len(rej) != 1 || rej[0].Why != "Invalid signature" {
		t.Fatalf("rejections: %+v", rej)
	}
}

func TestMissingAuthHeadersRejectedByTheRig(t *testing.T) {
	// a RAW request with no headers — the rig's own law (the Worker's
	// first 401 reason), pinned from the bare wire
	fw := newFakeWorker(t)
	res, err := http.Get(fw.url() + "/api/pending")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != 401 {
		t.Fatalf("status = %d, want 401", res.StatusCode)
	}
	_, _ = io.ReadAll(res.Body)
	rej := fw.rejected()
	if len(rej) != 1 || rej[0].Why != "Missing auth headers" {
		t.Fatalf("rejections: %+v", rej)
	}
}

// THE PROXY-BYPASS LAW: a bogus environment proxy must never intercept
// the channel (workers.dev was intercepted by local v2rayN into 403s).
func TestProxyBypassFakeEnvProxyIgnored(t *testing.T) {
	fw := newFakeWorker(t)
	t.Setenv("HTTP_PROXY", "http://127.0.0.1:9")
	t.Setenv("HTTPS_PROXY", "http://127.0.0.1:9")
	t.Setenv("http_proxy", "http://127.0.0.1:9")
	t.Setenv("https_proxy", "http://127.0.0.1:9")
	s := NewSync(rigConfig(fw))
	state, _ := s.GetPendingResult()
	if state != "ok" {
		t.Fatal("the bogus env proxy broke the channel")
	}
	// structural: the channel's client carries NO proxy hook
	tr, ok := directClient(1.0).Transport.(*http.Transport)
	if !ok || tr.Proxy != nil {
		t.Error("the channel's transport has a proxy hook — the bypass law is broken")
	}
}

func TestConfirmationGateTrios(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))

	id, why := s.ProposeBanish(2, []any{}, 300)
	if id == nil {
		t.Fatalf("propose failed: %s", why)
	}
	if status := s.BanishStatus(*id); status != "pending" {
		t.Errorf("status = %q", status)
	}
	if !s.ReportBanishResult(*id, "deleted", 2) {
		t.Error("report failed")
	}

	scanID, why2 := s.ProposeScan(1, 2, 3, []any{}, "the plan", 300)
	if scanID == nil {
		t.Fatalf("scan propose failed: %s", why2)
	}
	if status := s.ScanStatus(*scanID); status != "pending" {
		t.Errorf("scan status = %q", status)
	}
	if !s.ReportScanResult(*scanID, "applied", 1, 2, 3) {
		t.Error("scan report failed")
	}
	// not enabled → (nil, "not enabled")
	s2 := NewSync(map[string]any{})
	if id2, why3 := s2.ProposeBanish(1, nil, 0); id2 != nil || why3 != "not enabled" {
		t.Errorf("disabled propose = (%v, %q)", id2, why3)
	}
	if s2.BanishStatus("x") != "" || s2.ReportBanishResult("x", "y", 0) {
		t.Error("disabled status/report must be zero values")
	}
}

func TestPushErrorsBackupVerifyDecommissionsBackfillHealth(t *testing.T) {
	fw := newFakeWorker(t)
	s := NewSync(rigConfig(fw))

	errRow := storage.NewOrderedMap()
	errRow.Set("severity", "CRITICAL")
	errRow.Set("error_code", "X")
	errRow.Set("message", "boom")
	if !s.PushErrors([]any{errRow}) {
		t.Fatal("push_errors failed")
	}
	fw.mu.Lock()
	errs := fw.errorsSeen
	fw.mu.Unlock()
	if len(errs) != 1 {
		t.Fatalf("errors seen = %d", len(errs))
	}

	if !s.PushBackupStatus(backupSample()) {
		t.Fatal("push_backup_status failed")
	}
	fw.mu.Lock()
	backs := fw.backupSeen
	fw.mu.Unlock()
	if len(backs) != 1 {
		t.Fatalf("backups seen = %d", len(backs))
	}
	// install_id first, then the snapshot's keys in order
	keys := backs[0].Keys()
	if keys[0] != "install_id" || keys[1] != "gdrive_last_backup" {
		t.Errorf("backup body keys = %v", keys)
	}

	v := s.Verify()
	if v == nil {
		t.Fatal("verify failed")
	}
	if got := pyIntDefault(dictValue(dictValue(v, "stats"), "total_ledger"), -1); got != 3 {
		t.Errorf("total_ledger = %d", got)
	}

	if !s.PushDecommission("https://gone.test/", "404", "") {
		t.Fatal("push_decommission failed")
	}
	decomms := s.GetDecommissions("")
	if len(decomms) != 1 {
		t.Fatalf("decommissions = %d", len(decomms))
	}

	entry := storage.NewOrderedMap()
	entry.Set("url_normalized", "https://old.test/")
	ok, resp := s.BackfillChunk([]any{entry}, 0, 1)
	if !ok {
		t.Fatalf("backfill failed: %+v", resp)
	}
	if got := pyIntDefault(dictValue(resp, "inserted"), -1); got != 1 {
		t.Errorf("inserted = %d", got)
	}

	h := s.HealthCheck()
	if h == nil {
		t.Fatal("health failed")
	}
	if got := pyStr(dictValue(h, "version")); got != "0.40.0" {
		t.Errorf("health version = %q", got)
	}
}

// THE DIRECTORY DOORS (the rig's other half — T-49's reuse): the
// honest 404, the Bearer publish (constant-time), the commands door.
func TestRigDirectoryDoors(t *testing.T) {
	fw := newFakeWorker(t)

	// never-published: the honest 404 — a STATE, not an error
	res, err := httpGet(fw.url() + "/api/directory/index")
	if err != nil {
		t.Fatal(err)
	}
	if res.status != 404 {
		t.Fatalf("index status = %d, want the honest 404", res.status)
	}
	if got := pyStr(dictValue(res.data, "error")); got != "index not published yet" {
		t.Errorf("error = %q", got)
	}

	// the publish door: wrong Bearer → 401; right Bearer → published
	bad := publishRequest(fw, "wrong-token")
	if bad.status != 401 {
		t.Fatalf("wrong token status = %d, want 401", bad.status)
	}
	payload := storage.NewOrderedMap()
	payload.Set("schema", "gitcurator-directory/1")
	payload.Set("generated_at", "2026-10-16T00:00:00+00:00")
	counts := storage.NewOrderedMap()
	counts.Set("sites", 590)
	counts.Set("categories", 13)
	payload.Set("counts", counts)
	good := publishRequestWith(fw, rigToken, payload)
	if good.status != 200 {
		t.Fatalf("publish status = %d (%v)", good.status, good.data)
	}
	// and the index now serves it back
	after, err := httpGet(fw.url() + "/api/directory/index")
	if err != nil {
		t.Fatal(err)
	}
	if after.status != 200 {
		t.Fatalf("index-after status = %d", after.status)
	}
	if got := pyIntDefault(dictValue(dictValue(after.data, "counts"), "sites"), -1); got != 590 {
		t.Errorf("sites = %d", got)
	}

	// the commands door: Bearer-only
	cmds := commandsRequest(fw, "wrong-token")
	if cmds.status != 401 {
		t.Fatalf("wrong token commands status = %d", cmds.status)
	}
	cmdsOK := commandsRequest(fw, rigToken)
	if cmdsOK.status != 200 {
		t.Fatalf("commands status = %d", cmdsOK.status)
	}
}

// --- small test helpers ---------------------------------------------------

func mirrorEntryFor(url, path string) *storage.OrderedMap {
	om := storage.NewOrderedMap()
	om.Set("url_normalized", url)
	om.Set("vault_path", path)
	om.Set("category", "")
	om.Set("title", "")
	om.Set("status", "in_vault")
	om.Set("first_processed_at", "")
	om.Set("last_updated_at", "")
	return om
}

func backupSample() *storage.OrderedMap {
	om := storage.NewOrderedMap()
	om.Set("gdrive_last_backup", "2026-10-16T00:00:00+00:00")
	om.Set("gdrive_backup_count", 3)
	return om
}

func deadWorkerURL(t *testing.T) string {
	t.Helper()
	return "http://127.0.0.1:9"
}

type httpResult struct {
	status int
	data   any
}

func httpGet(u string) (httpResult, error) {
	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		return httpResult{}, err
	}
	req.Header.Set("User-Agent", HTTPUserAgent)
	res, err := directClient(10.0).Do(req)
	if err != nil {
		return httpResult{}, err
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return httpResult{res.StatusCode, mustDecode(body)}, nil
}

func publishRequest(fw *fakeWorker, token string) httpResult {
	payload := storage.NewOrderedMap()
	counts := storage.NewOrderedMap()
	counts.Set("sites", 1)
	payload.Set("counts", counts)
	return publishRequestWith(fw, token, payload)
}

func publishRequestWith(fw *fakeWorker, token string,
	payload *storage.OrderedMap) httpResult {
	req, err := http.NewRequest("POST", fw.url()+"/api/directory/sync",
		strings.NewReader(storage.PyDumpsCompact(payload)))
	if err != nil {
		return httpResult{0, nil}
	}
	req.Header.Set("User-Agent", HTTPUserAgent)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	res, err := directClient(10.0).Do(req)
	if err != nil {
		return httpResult{0, nil}
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return httpResult{res.StatusCode, mustDecode(body)}
}

func commandsRequest(fw *fakeWorker, token string) httpResult {
	req, err := http.NewRequest("GET", fw.url()+"/api/commands", nil)
	if err != nil {
		return httpResult{0, nil}
	}
	req.Header.Set("User-Agent", HTTPUserAgent)
	req.Header.Set("Authorization", "Bearer "+token)
	res, err := directClient(10.0).Do(req)
	if err != nil {
		return httpResult{0, nil}
	}
	defer res.Body.Close()
	body, _ := io.ReadAll(res.Body)
	return httpResult{res.StatusCode, mustDecode(body)}
}
