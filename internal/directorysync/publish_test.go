// publish_test.go — T-49: the ported selections of the reference's
// TestPublish (test_directorysync.py) against the loopback stub
// (success, wrong token, Worker offline — the failure case from SPEC
// §14: the old site content stays live, nothing crashes) + THE
// READ-BACK DOOR's four truths + run_sync end-to-end.
package directorysync

import (
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

func stubIndex() *storage.OrderedMap {
	index := storage.NewOrderedMap()
	index.Set("schema", DirectorySchema)
	index.Set("categories", []any{})
	index.Set("sites", []any{})
	index.Set("popular_tags", []any{})
	index.Set("synonyms", storage.NewOrderedMap())
	counts := storage.NewOrderedMap()
	counts.Set("sites", 0)
	counts.Set("categories", 0)
	index.Set("counts", counts)
	index.Set("generated_at", "2026-10-12T00:00:00Z")
	index.Set("new_days", 14)
	return index
}

func TestPublish_SuccessPostsWithToken(t *testing.T) {
	server := newStubWorker(t, "")
	result := PublishIndex(stubIndex(), server.url(), stubToken, 0, nil)
	if result == nil {
		t.Fatal("publish returned nil")
	}
	if !pyTruthyDS(dictDoorOf(result).Get("success")) {
		t.Fatalf("result: %v", result)
	}
	received := server.receivedIndexes()
	if len(received) != 1 {
		t.Fatalf("received: %d", len(received))
	}
	if omStr(received[0], "schema") != DirectorySchema {
		t.Fatalf("schema: %q", omStr(received[0], "schema"))
	}
}

func TestPublish_WrongTokenReturnsNilNeverRaises(t *testing.T) {
	server := newStubWorker(t, "right-token")
	var logs []string
	result := PublishIndex(stubIndex(), server.url(), "wrong-token", 0, func(msg string) {
		logs = append(logs, msg)
	})
	if result != nil {
		t.Fatalf("result: %v", result)
	}
	found := false
	for _, l := range logs {
		if strings.Contains(l, "401") {
			found = true
		}
	}
	if !found {
		t.Fatalf("logs: %v", logs)
	}
	if len(server.receivedIndexes()) != 0 {
		t.Fatal("the stub recorded a body on a 401")
	}
}

func TestPublish_WorkerOfflineReturnsNilNeverRaises(t *testing.T) {
	// Port 1 on localhost: closed with certainty, connection refused.
	var logs []string
	result := PublishIndex(stubIndex(), "http://127.0.0.1:1", "any-token", 0,
		func(msg string) { logs = append(logs, msg) })
	if result != nil {
		t.Fatalf("result: %v", result)
	}
	found := false
	for _, l := range logs {
		if strings.Contains(l, "failed") {
			found = true
		}
	}
	if !found {
		t.Fatalf("logs: %v", logs)
	}
}

func TestPublish_NonASCIIBodyRidesRawUTF8(t *testing.T) {
	// The spec's publish-payload shape law: the body is
	// json.dumps(index, ensure_ascii=False) — the Farsi rides RAW, the
	// Bearer header rides the wire (pinned byte-for-byte by the
	// golden's publish_body section; this leg proves the transport
	// carries it).
	server := newStubWorker(t, "")
	index := stubIndex()
	aliases := storage.NewOrderedMap()
	aliases.Set("color-tool", "ابزار رنگ")
	index.Set("tag_aliases", aliases)
	result := PublishIndex(index, server.url(), stubToken, 0, nil)
	if result == nil {
		t.Fatal("publish returned nil")
	}
	received := server.receivedIndexes()
	if len(received) != 1 {
		t.Fatalf("received: %d", len(received))
	}
	raw, _ := received[0].Get("tag_aliases")
	aliasesOut, _ := raw.(*storage.OrderedMap)
	fa, _ := aliasesOut.Get("color-tool")
	if fa != "ابزار رنگ" {
		t.Fatalf("the Farsi alias did not survive the wire: %v", fa)
	}
}

func TestPublish_RunSyncEndToEndPublishes(t *testing.T) {
	v := newVaultBuilder(t)
	server := newStubWorker(t, "")
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	config := map[string]any{
		"website_vault_path":   v.dir,
		"directory_worker_url": server.url(),
		"directory_sync_token": stubToken,
	}
	if !RunSync(config, nil) {
		t.Fatal("run_sync returned false")
	}
	received := server.receivedIndexes()
	if len(received) != 1 {
		t.Fatalf("received: %d", len(received))
	}
	sites, _ := received[0].Get("sites")
	if len(sites.([]any)) != 1 {
		t.Fatalf("sites: %v", sites)
	}
}

func TestPublish_RunSyncMissingVaultIsQuietFailure(t *testing.T) {
	if RunSync(map[string]any{"website_vault_path": ""}, nil) {
		t.Fatal("run_sync returned true on a missing vault")
	}
}

// --- the read-back door's four truths (spec: read-back four truths) ---

func TestFetchPublishedIndex_FourTruths(t *testing.T) {
	// 1. not_configured: no worker URL in the config.
	out := FetchPublishedIndex(map[string]any{}, 0, nil)
	if omStr(out, "state") != "not_configured" {
		t.Fatalf("state: %q (%v)", omStr(out, "state"), out)
	}
	if omAny(out, "sites") != nil {
		t.Fatalf("sites: %v", omAny(out, "sites"))
	}

	// 2. never_published: the Worker's own honest 404.
	server := newStubWorker(t, "")
	out = FetchPublishedIndex(map[string]any{
		"directory_worker_url": server.url(),
	}, 0, nil)
	if omStr(out, "state") != "never_published" {
		t.Fatalf("state: %q", omStr(out, "state"))
	}

	// 3. ok: the published index answers its site count + generated_at.
	v := newVaultBuilder(t)
	v.write(t, "Design/Assets & Resources/Coolors.md", longNote)
	if !RunSync(map[string]any{
		"website_vault_path":   v.dir,
		"directory_worker_url": server.url(),
		"directory_sync_token": stubToken,
	}, nil) {
		t.Fatal("run_sync failed")
	}
	out = FetchPublishedIndex(map[string]any{
		"directory_worker_url": server.url(),
	}, 0, nil)
	if omStr(out, "state") != "ok" {
		t.Fatalf("state: %q (%v)", omStr(out, "state"), out)
	}
	if int(pyIntOf(omAny(out, "sites"))) != 1 {
		t.Fatalf("sites: %v", omAny(out, "sites"))
	}
	if omStr(out, "generated_at") == "" {
		t.Fatalf("generated_at: %q", omStr(out, "generated_at"))
	}

	// 4. error: a down Worker is spoken, never swallowed.
	out = FetchPublishedIndex(map[string]any{
		"directory_worker_url": "http://127.0.0.1:1",
	}, 0, nil)
	if omStr(out, "state") != "error" {
		t.Fatalf("state: %q", omStr(out, "state"))
	}
	if omStr(out, "message") == "" {
		t.Fatal("the error state carries no message")
	}
}

func TestFetchPublishedIndex_ANonIntSiteCountIsTheHonestError(t *testing.T) {
	// counts.sites must be a true int (a float or bool is rejected) —
	// the reference's isinstance door.
	server := newStubWorker(t, "")
	// pre-seed the stub's published index with a float site count
	fake := storage.NewOrderedMap()
	counts := storage.NewOrderedMap()
	counts.Set("sites", 1.5)
	fake.Set("counts", counts)
	fake.Set("generated_at", "x")
	server.mu.Lock()
	server.publishedIndex = fake
	server.mu.Unlock()
	out := FetchPublishedIndex(map[string]any{
		"directory_worker_url": server.url(),
	}, 0, nil)
	if omStr(out, "state") != "error" {
		t.Fatalf("state: %q", omStr(out, "state"))
	}
	if omStr(out, "message") != "the published index carried no site count" {
		t.Fatalf("message: %q", omStr(out, "message"))
	}
}
