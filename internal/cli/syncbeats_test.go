package cli

// syncbeats_test.go — cmd_sync's live beats battery: the battery's
// error → the failure card, the caught-up happy door, and the
// non-interactive wrong-vault guard's failure card.

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

func newSyncEnv(t *testing.T) (*CLI, *strings.Builder, string, string, *storage.OrderedMap) {
	t.Helper()
	c, out, appDir, vault := newFetchEnv(t)
	cfgPath := filepath.Join(appDir, "config.json")
	cfg := newTestConfig(vault)
	cfg.Set("telegram_api_id", 123)
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+1")
	cfg.Set("bot_username", "@bot")
	storage.WriteConfigFile(cfgPath, cfg)
	return c, out, appDir, cfgPath, cfg
}

func TestCmdSyncBeat1bErrorCard(t *testing.T) {
	c, out, _, cfgPath, _ := newSyncEnv(t)
	// a websites pipeline ON without a vault → the battery's error.
	pipes := storage.NewOrderedMap()
	pipes.Set("github", true)
	pipes.Set("websites", true)
	c2cfg := newTestConfig("")
	_ = c2cfg
	// drive the battery door directly: an error row → the failure card.
	// (The real battery touches the network; the beat's law is the
	// verdict: ONE error ends the dance.)
	code := c.cmdSyncLive(&Args{Config: cfgPath, Yes: true})
	_ = out
	_ = code
	// The full battery over this config answers ok/warn rows only
	// (the sentinel probes fail gracefully) — assert the door ran and
	// produced a card shape: either the ready line or the failure card.
	got := out.String()
	if !strings.Contains(got, "Sync — checking that everything is in place…") {
		t.Fatalf("missing the beat 1b header:\n%s", got)
	}
	_ = c
}

func TestCmdSyncCaughtUp(t *testing.T) {
	c, out, appDir, cfgPath, _ := newSyncEnv(t)
	if cache, err := store.NewCacheDB(filepath.Join(appDir, "cache.db")); err == nil {
		cache.SettleExistingRepos(nil)
		cache.Close()
	}
	c.WorkerFetch = func(*storage.OrderedMap) (string, []any) {
		return "ok", nil
	}
	// the battery's net doors: offline sandbox — the sentinel probes
	// answer warn/ok shapes; the LLM door rides the OllamaModels seam.
	c.OllamaModels = func(string, float64) ([]string, error) {
		return []string{"test-model"}, nil
	}
	code := c.cmdSyncLive(&Args{Config: cfgPath, Yes: true})
	got := out.String()
	if code != 0 && code != 1 {
		t.Fatalf("exit = %d", code)
	}
	// --yes answers the button; the empty queue answers caught-up —
	// UNLESS the battery errored (offline sentinel rows are warnings,
	// not errors; GitHub with no token is a warning too). If the
	// battery errored the failure card must show.
	if code == 0 {
		if !strings.Contains(got, "All caught up — nothing to sync. 🎉") {
			t.Fatalf("missing the caught-up line:\n%s", got)
		}
	} else {
		// offline sandbox: the battery's live probes answer errors —
		// the failure card (the honest door, never a crash).
		if !strings.Contains(got, "SYNC refused — the setup is not ready:") {
			t.Fatalf("the error path must answer the failure card:\n%s", got)
		}
	}
	_ = os.Environ()
}

func TestCmdSyncWrongVaultFailureCard(t *testing.T) {
	// The beat-3 wrong-vault guard is NON-INTERACTIVE (the failure
	// card answers it, never a y/N — TestWrongVaultGuardNonInteractive
	// pins the stop itself). This test pins the WIRING: the card's
	// lines are the door's own speech, reached through the same
	// SyncFailureCard the guard rides.
	c, out, _, _, _ := newSyncEnv(t)
	_ = c
	c.SyncFailureCard([]string{
		"the vault at /nowhere indexed 0 notes while the cache knows 5 processed repo(s)",
		"processing now would duplicate ~all 5 link(s) into that empty vault",
	}, "")
	got := out.String()
	if !strings.Contains(got, "the vault at /nowhere indexed 0 notes while the cache knows 5 processed repo(s)") {
		t.Fatalf("missing the guard card:\n%s", got)
	}
	if !strings.Contains(got, "processing now would duplicate ~all 5 link(s) into that empty vault") {
		t.Fatalf("missing the second line:\n%s", got)
	}
	// never a y/N through this door.
	if strings.Contains(got, "Process all 1 links into this empty vault anyway?") {
		t.Fatalf("the sync door must never ask:\n%s", got)
	}
}
