package cli

// syncdoor_test.go — THE SYNC DOOR's own laws beyond the golden's byte
// shapes: the GUI-instruction law (the owner's own words), the or-door
// unit shapes, and the save-cleaner.

import (
	"bytes"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// TestGuiInstructionLaw mirrors TheGuiInstructionLaw: one remedy, the
// owner's own words.
func TestGuiInstructionLaw(t *testing.T) {
	text := SyncGuiInstruction()
	for _, want := range []string{
		"GUI version", "credentials and essentials again", "Sync once more",
	} {
		if !strings.Contains(text, want) {
			t.Errorf("GUI instruction lost %q", want)
		}
	}
}

// TestSyncEssentialsOrDoors — the or-door shapes the golden's matrix
// exercises only through strings.
func TestSyncEssentialsOrDoors(t *testing.T) {
	// telegram_api_id 0 (int) is falsy → missing
	cfg := storage.NewOrderedMap()
	cfg.Set("telegram_api_id", int64(0))
	cfg.Set("telegram_api_hash", "hash")
	cfg.Set("telegram_phone", "+98")
	cfg.Set("bot_username", "bot")
	cfg.Set("vault_path", "/vault")
	cfg.Set("llm_provider", "ollama")
	oll := storage.NewOrderedMap()
	oll.Set("model", "llama3")
	cfg.Set("ollama", oll)
	problems := SyncMissingEssentials(cfg, func(string) bool { return true })
	if len(problems) != 1 || problems[0] != "Telegram api id is missing" {
		t.Errorf("or-door: problems = %v, want the falsy api id named", problems)
	}
	// a bare-at bot is missing
	cfg2 := storage.NewOrderedMap()
	cfg2.Set("telegram_api_id", int64(1))
	cfg2.Set("telegram_api_hash", "hash")
	cfg2.Set("telegram_phone", "+98")
	cfg2.Set("bot_username", "@")
	cfg2.Set("vault_path", "/vault")
	cfg2.Set("llm_provider", "ollama")
	cfg2.Set("ollama", oll)
	problems2 := SyncMissingEssentials(cfg2, func(string) bool { return true })
	if len(problems2) != 1 ||
		problems2[0] != "the bot username is missing (the queue cannot be read)" {
		t.Errorf("bare-at bot: problems = %v", problems2)
	}
	// llm_provider "" (empty string) → the or-door → ollama branch
	cfg3 := storage.NewOrderedMap()
	cfg3.Set("telegram_api_id", int64(1))
	cfg3.Set("telegram_api_hash", "hash")
	cfg3.Set("telegram_phone", "+98")
	cfg3.Set("bot_username", "bot")
	cfg3.Set("vault_path", "/vault")
	cfg3.Set("llm_provider", "")
	cfg3.Set("ollama", oll)
	if got := SyncMissingEssentials(cfg3, func(string) bool { return true }); len(got) != 0 {
		t.Errorf("empty provider: problems = %v, want none (the or-door → ollama)", got)
	}
}

// TestSyncConfirmPromptNonTty — the cron law through the CLI shell.
func TestSyncConfirmPromptNonTty(t *testing.T) {
	buf := &bytes.Buffer{}
	c := testCLI(buf, "q\n")
	if !c.SyncConfirmPrompt(false, false) {
		t.Errorf("non-TTY: refused, want the cron-law auto-sync")
	}
	if buf.Len() != 0 {
		t.Errorf("non-TTY: the button rendered: %q", buf.String())
	}
	if !c.SyncConfirmPrompt(true, true) {
		t.Errorf("--yes: refused, want the cron-law auto-sync")
	}
}

// TestCleanConfigForSave — the __…__ keys strip.
func TestCleanConfigForSave(t *testing.T) {
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", "/v")
	cfg.Set("__config_path__", "/tmp/x.json")
	cfg.Set("github_token", "t")
	inner := storage.NewOrderedMap()
	cfg.Set("proxy", inner) // nested maps survive untouched
	got := cleanConfigForSave(cfg)
	if _, ok := got.Get("__config_path__"); ok {
		t.Errorf("the private key survived the cleaner")
	}
	if v, _ := got.Get("vault_path"); v != "/v" {
		t.Errorf("vault_path lost: %v", v)
	}
	if _, ok := got.Get("proxy"); !ok {
		t.Errorf("proxy lost")
	}
}

// TestMaskTokenLaw — the head+tail mask.
func TestMaskTokenLaw(t *testing.T) {
	if got := MaskToken("abcdefghij", 4, 4); got != "••••••••••" {
		t.Errorf("short token = %q, want all dots", got)
	}
	if got := MaskToken("abcdefghijk", 4, 4); got != "abcd••••••••hijk" {
		t.Errorf("token = %q", got)
	}
}
