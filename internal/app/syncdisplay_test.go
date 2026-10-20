package app

// syncdisplay_test.go — T-75 slice (b)'s own battery: THE VISUAL SYNC
// DISPLAY's pure model (the acceptance item — the four-state rendering
// tests): the classify split, the friendly clock, the §5 arithmetic's
// every edge (shrink, never-published, unreachable, not-configured),
// the four cards' every state (ok / not_set / error / never — the
// connection truth first), the hero laws, the calm line, the flow
// order, and the total junk tolerance. Ported from
// tests/test_syncdisplay.py's model families.

import (
	"strings"
	"testing"
	"time"

	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/syncreview"
)

// ---------------------------------------------------------------------------
// classify_pending — the router's own split
// ---------------------------------------------------------------------------

func pendRow(url string) map[string]any {
	return map[string]any{"url_normalized": url}
}

func TestClassifyPendingTheOwnersOwnExample(t *testing.T) {
	// "the part of telegram bot shows 3 new repo 2 new websites"
	items := []any{
		pendRow("https://github.com/a/one"), pendRow("https://github.com/a/two"),
		pendRow("https://github.com/a/three"),
		pendRow("https://example.com/1"), pendRow("https://example.com/2"),
	}
	got := ClassifyPending(items)
	if got.Total != 5 || got.Repos != 3 || got.Websites != 2 || got.Unrouted != 0 {
		t.Fatalf("split = %+v", got)
	}
}

func TestClassifyPendingGitHubIOMapsToRepos(t *testing.T) {
	got := ClassifyPending([]any{pendRow("https://owner.github.io/repo")})
	if got.Repos != 1 || got.Websites != 0 {
		t.Fatalf("split = %+v", got)
	}
}

func TestClassifyPendingGistsAreWebsites(t *testing.T) {
	// the intake's own law: gists and every other link are websites
	got := ClassifyPending([]any{pendRow("https://gist.github.com/a/b")})
	if got.Websites != 1 || got.Repos != 0 {
		t.Fatalf("split = %+v", got)
	}
}

func TestClassifyPendingJunkRowsAreUnrouted(t *testing.T) {
	got := ClassifyPending([]any{
		42, map[string]any{"no_url": true}, nil, pendRow(""),
	})
	if got.Total != 4 || got.Unrouted != 4 {
		t.Fatalf("split = %+v (junk is spoken, never guessed)", got)
	}
}

func TestClassifyPendingNilIsZero(t *testing.T) {
	got := ClassifyPending(nil)
	if got.Total != 0 || got.Repos+got.Websites+got.Unrouted != 0 {
		t.Fatalf("split = %+v", got)
	}
}

// TestQueueItemURLOrderedMapPriorityTuple pins the reference's own
// _item_url law (sync_display.py:118-127) over the OrderedMap row
// shape: the FIXED key tuple ('url_normalized', 'url_original',
// 'url') speaks by priority — the row's own key insertion order NEVER
// decides. The Worker's row carries forward_count/ledger_id/…
// alongside the URL keys (spec-worker-contract), so any key order is
// a legal row; the first-flake (SESSION-045's pre-work repair: the
// botdoors rig's randomly-ordered rows dropped the URL whenever
// forward_count landed first) is pinned here deterministically.
func TestQueueItemURLOrderedMapPriorityTuple(t *testing.T) {
	// forward_count FIRST — the flake's own row shape.
	badOrder := storage.NewOrderedMap()
	badOrder.Set("forward_count", float64(4))
	badOrder.Set("url_normalized", "https://github.com/acme/x")
	if u := queueItemURL(badOrder); u != "https://github.com/acme/x" {
		t.Fatalf("forward_count-first row: queueItemURL = %q, want the url_normalized", u)
	}
	if n := queueItemForwardCount(badOrder); n != 4 {
		t.Fatalf("forward_count read = %d, want 4", n)
	}

	// the priority beats the insertion order: url_original first,
	// url_normalized second — Python's tuple answers url_normalized.
	prio := storage.NewOrderedMap()
	prio.Set("url_original", "https://example.com/orig")
	prio.Set("url_normalized", "https://example.com/norm")
	if u := queueItemURL(prio); u != "https://example.com/norm" {
		t.Fatalf("priority tuple: queueItemURL = %q, want url_normalized", u)
	}

	// the fallback ladder over insertion order: url_original only.
	fallback := storage.NewOrderedMap()
	fallback.Set("url_original", "https://example.com/orig")
	if u := queueItemURL(fallback); u != "https://example.com/orig" {
		t.Fatalf("fallback ladder: queueItemURL = %q, want url_original", u)
	}

	// a non-string at a higher-priority key falls through to the
	// next key (Python's isinstance ladder), and an all-junk row
	// stays unrouted.
	nonStr := storage.NewOrderedMap()
	nonStr.Set("url_normalized", 42)
	nonStr.Set("url", "https://example.com/plain")
	if u := queueItemURL(nonStr); u != "https://example.com/plain" {
		t.Fatalf("non-string fallthrough: queueItemURL = %q, want url", u)
	}
	junk := storage.NewOrderedMap()
	junk.Set("ledger_id", int64(9))
	if u := queueItemURL(junk); u != "" {
		t.Fatalf("junk row: queueItemURL = %q, want \"\"", u)
	}

	// and the routed split over the bad-order row end-to-end.
	githubURLs, nonGithub, raw := routeQueueItems([]any{badOrder})
	if raw != 1 || len(githubURLs) != 1 || len(nonGithub) != 0 {
		t.Fatalf("bad-order row routed: github=%v non=%v raw=%d", githubURLs, nonGithub, raw)
	}
}

// ---------------------------------------------------------------------------
// friendly_when — the Last-sync line's own grammar
// ---------------------------------------------------------------------------

func TestFriendlyWhenToday(t *testing.T) {
	now := time.Now()
	stamp := now.Format("2006-01-02T15:04:05")
	got := FriendlyWhen(stamp)
	if !strings.HasPrefix(got, "today ") || len(got) != len("today 00:00") {
		t.Fatalf("got %q, want the today HH:MM shape", got)
	}
}

func TestFriendlyWhenOtherDayKeepsTheDate(t *testing.T) {
	got := FriendlyWhen("2020-01-02T03:04:05")
	if got != "2020-01-02 03:04" {
		t.Fatalf("got %q", got)
	}
}

func TestFriendlyWhenJunkIsSpokenRaw(t *testing.T) {
	if got := FriendlyWhen("not a stamp"); got != "not a stamp" {
		t.Fatalf("got %q", got)
	}
}

func TestFriendlyWhenNothingIsNever(t *testing.T) {
	for _, stamp := range []string{"", "   "} {
		if got := FriendlyWhen(stamp); got != "never" {
			t.Fatalf("got %q, want never", got)
		}
	}
}

// ---------------------------------------------------------------------------
// THE ARITHMETIC + the four states (the website card's own battery)
// ---------------------------------------------------------------------------

func webFacts(vState string, vCount *int, pState string, pCount *int,
	gen string) *SyncFacts {
	return &SyncFacts{
		Website: &SyncWebsiteFacts{
			Vault:     SyncWebsiteVaultSide{State: vState, PublicReady: vCount},
			Published: SyncPublishedFacts{State: pState, Sites: pCount, GeneratedAt: gen},
		},
	}
}

func cardOf(t *testing.T, m *SyncDisplayModel, key string) SyncCard {
	t.Helper()
	for _, c := range m.Cards {
		if c.Key == key {
			return c
		}
	}
	t.Fatalf("no card %q", key)
	return SyncCard{}
}

func TestArithmeticTheOwnersOwnExample(t *testing.T) {
	// "website display shows need to be updated with 30 new websites"
	m := SyncDisplayModelOf(webFacts(SyncStateOK, ipApp(30), SyncStateOK,
		ipApp(0), "2026-10-01T00:00:00"))
	web := cardOf(t, m, ElementWebsite)
	if web.State != SyncStateOK || web.Pending != 30 {
		t.Fatalf("card = %+v", web)
	}
	joined := strings.Join(web.Lines, "\n")
	if !strings.Contains(joined, "vault has 30 public-ready notes") ||
		!strings.Contains(joined, "30 need publishing") {
		t.Fatalf("lines = %v", web.Lines)
	}
	if !strings.Contains(strings.Join(web.Foot, "\n"),
		"30 in the vault · 0 on the site — ⚡ Fast Sync publishes the gap") {
		t.Fatalf("foot = %v", web.Foot)
	}
}

func TestArithmeticTheHeroIsTheBigNumber(t *testing.T) {
	m := SyncDisplayModelOf(webFacts(SyncStateOK, ipApp(644), SyncStateOK,
		ipApp(640), ""))
	web := cardOf(t, m, ElementWebsite)
	if web.Hero.Value != "644" || web.Hero.Label != "public-ready in the vault" ||
		web.Hero.Tone != "success" {
		t.Fatalf("hero = %+v", web.Hero)
	}
}

func TestArithmeticShrinkingVaultIsNeverNegative(t *testing.T) {
	// V < P: "−N since the last publish (banished notes retire too)" —
	// never a negative needs-publishing
	m := SyncDisplayModelOf(webFacts(SyncStateOK, ipApp(700), SyncStateOK,
		ipApp(710), ""))
	web := cardOf(t, m, ElementWebsite)
	if web.Pending != 0 {
		t.Fatalf("pending = %d (never negative)", web.Pending)
	}
	joined := strings.Join(web.Lines, "\n")
	if !strings.Contains(joined, "−10 since the last publish (banished notes retire too)") {
		t.Fatalf("lines = %v", web.Lines)
	}
	found := false
	for _, st := range web.Stats {
		if st.Label == "retired since the last publish" && st.Value == "−10" {
			found = true
		}
	}
	if !found {
		t.Fatalf("stats = %+v", web.Stats)
	}
}

func TestArithmeticInSyncSpeaksEquality(t *testing.T) {
	m := SyncDisplayModelOf(webFacts(SyncStateOK, ipApp(720), SyncStateOK,
		ipApp(720), ""))
	web := cardOf(t, m, ElementWebsite)
	if !containsLine(web.Lines, "in sync with the vault") || web.Pending != 0 {
		t.Fatalf("lines = %v", web.Lines)
	}
}

func TestArithmeticNeverPublishedIsNeverNotFailure(t *testing.T) {
	// the Worker's own 404: the site is empty until the first publish
	m := SyncDisplayModelOf(webFacts(SyncStateOK, ipApp(30),
		SyncStateNever, nil, ""))
	web := cardOf(t, m, ElementWebsite)
	if web.State != SyncStateOK || web.Pending != 30 {
		t.Fatalf("card = %+v (never is not failure)", web)
	}
	joined := strings.Join(web.Lines, "\n")
	if !strings.Contains(joined, "never published — the site is empty until the first publish runs") ||
		!strings.Contains(joined, "30 would be new on the site") {
		t.Fatalf("lines = %v", web.Lines)
	}
}

func TestArithmeticUnreachableWorkerDegradesToUnknown(t *testing.T) {
	facts := webFacts(SyncStateOK, ipApp(30), SyncStateError, nil, "")
	facts.Website.Published.Message = "the poll failed"
	m := SyncDisplayModelOf(facts)
	web := cardOf(t, m, ElementWebsite)
	if web.State != SyncStateError || web.Pending != 0 {
		t.Fatalf("card = %+v (an unreachable worker is never a fabricated count)", web)
	}
	if !containsLine(web.Lines, "published count unknown — the poll failed") {
		t.Fatalf("lines = %v", web.Lines)
	}
}

func TestArithmeticNoWorkerURLIsNotConfigured(t *testing.T) {
	m := SyncDisplayModelOf(webFacts(SyncStateOK, ipApp(30),
		SyncStateNotSet, nil, ""))
	web := cardOf(t, m, ElementWebsite)
	if !containsLine(web.Lines, "published count unknown — no worker URL set") {
		t.Fatalf("lines = %v", web.Lines)
	}
	if !containsLine(web.Foot, "set the Cloudflare Worker URL (Settings → Credentials) — then this card speaks the site's own count") {
		t.Fatalf("foot = %v", web.Foot)
	}
}

func TestWebsiteVaultNotSetSpeaksItsDoor(t *testing.T) {
	m := SyncDisplayModelOf(webFacts(SyncStateNotSet, nil,
		SyncStateNotSet, nil, ""))
	web := cardOf(t, m, ElementWebsite)
	if web.State != SyncStateNotSet || web.DotTone != "muted" {
		t.Fatalf("card = %+v", web)
	}
	if !containsLine(web.Lines, "nothing to publish until the Websites vault is set") {
		t.Fatalf("lines = %v", web.Lines)
	}
	if web.Hero.Value != "—" {
		t.Fatalf("hero = %+v", web.Hero)
	}
}

// ---------------------------------------------------------------------------
// the bot card — the three truths of the honest poll
// ---------------------------------------------------------------------------

func TestBotDisabledIsNotSetNeverAnError(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{Bot: &SyncBotFacts{
		State: SyncStateNotSet}})
	bot := cardOf(t, m, ElementBot)
	if bot.State != SyncStateNotSet || bot.DotTone != "muted" {
		t.Fatalf("card = %+v", bot)
	}
	if !containsLine(bot.Lines, "queue not measured (the worker is not configured)") {
		t.Fatalf("lines = %v", bot.Lines)
	}
}

func TestBotErrorIsSpoken(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{Bot: &SyncBotFacts{
		State: SyncStateError, Message: "HTTP 500"}})
	bot := cardOf(t, m, ElementBot)
	if bot.State != SyncStateError || bot.DotTone != "error" {
		t.Fatalf("card = %+v", bot)
	}
	if !containsLine(bot.Lines, "queue unknown — HTTP 500") {
		t.Fatalf("lines = %v", bot.Lines)
	}
	if bot.Hero.Value != "—" || bot.Hero.Label != "queue unknown" {
		t.Fatalf("hero = %+v", bot.Hero)
	}
}

func TestBotOKSpeaksTheSplitAndTheWaitingHero(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{Bot: &SyncBotFacts{
		State: SyncStateOK,
		Pending: []any{pendRow("https://github.com/a/r"),
			pendRow("https://example.com/w")}}})
	bot := cardOf(t, m, ElementBot)
	if bot.State != SyncStateOK || bot.Pending != 2 {
		t.Fatalf("card = %+v", bot)
	}
	joined := strings.Join(bot.Lines, "\n")
	if !strings.Contains(joined, "2 links waiting to be synced") ||
		!strings.Contains(joined, "1 repo · 1 website") {
		t.Fatalf("lines = %v", bot.Lines)
	}
	if bot.Hero.Value != "2" || bot.Hero.Tone != "warning" {
		t.Fatalf("hero = %+v", bot.Hero)
	}
}

func TestBotCalmQueueIsTheSuccessHero(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{Bot: &SyncBotFacts{
		State: SyncStateOK, Pending: nil}})
	bot := cardOf(t, m, ElementBot)
	if !containsLine(bot.Lines, "Queue is calm — nothing is waiting") {
		t.Fatalf("lines = %v", bot.Lines)
	}
	if bot.Hero.Value != "0" || bot.Hero.Tone != "success" {
		t.Fatalf("hero = %+v", bot.Hero)
	}
}

// ---------------------------------------------------------------------------
// the vault card — the three vaults + the queue + the last batch
// ---------------------------------------------------------------------------

func vaultRow(kind, status string, count *int) SyncVaultRow {
	return SyncVaultRow{Kind: kind, Label: kind + " vault",
		Icon: "🗂️", Status: status, NoteCount: count,
		Scannable: kind != "personal"}
}

func TestVaultCardSpeaksTheThreeVaultsAndTheQueue(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{
		Vaults: &SyncVaultsFacts{Vaults: []SyncVaultRow{
			vaultRow("websites", VaultStatusConnected, ipApp(640)),
			vaultRow("repos", VaultStatusConnected, ipApp(1200)),
			vaultRow("personal", VaultStatusConnected, nil),
		}},
		Bot: &SyncBotFacts{State: SyncStateOK,
			Pending: []any{pendRow("https://github.com/a/r")}},
	})
	vault := cardOf(t, m, ElementVault)
	if vault.State != SyncStateOK {
		t.Fatalf("card = %+v", vault)
	}
	joined := strings.Join(vault.Lines, "\n")
	for _, want := range []string{
		"640 notes", "1200 notes", "owner-owned, read-only",
		"↳ 1 link waiting to become notes",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("lines missing %q: %v", want, vault.Lines)
		}
	}
	if vault.Hero.Value != "1840" {
		t.Fatalf("hero = %+v (the counted libraries' sum)", vault.Hero)
	}
}

func TestVaultCardLastBatchMemory(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{
		Vaults: &SyncVaultsFacts{
			Vaults:    []SyncVaultRow{vaultRow("repos", VaultStatusConnected, ipApp(5))},
			LastBatch: &SyncLastBatch{When: "2020-01-02T03:04:05", Done: 3, Total: 4},
		},
	})
	vault := cardOf(t, m, ElementVault)
	line := "last batch 2020-01-02 03:04 · 3 saved · 1 need retry"
	if !containsLine(vault.Lines, line) || !containsLine(vault.Foot, line) {
		t.Fatalf("lines = %v foot = %v", vault.Lines, vault.Foot)
	}
}

func TestVaultCardNotSetSpeaksTheDoor(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{
		Vaults: &SyncVaultsFacts{Vaults: []SyncVaultRow{
			vaultRow("repos", VaultStatusNotSet, nil)}}})
	vault := cardOf(t, m, ElementVault)
	if vault.State != SyncStateNotSet ||
		!strings.Contains(strings.Join(vault.Lines, "\n"), "— not set") {
		t.Fatalf("card = %+v", vault)
	}
}

func TestVaultCardNotFoundSpeaksHonestly(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{
		Vaults: &SyncVaultsFacts{Vaults: []SyncVaultRow{
			vaultRow("repos", VaultStatusNotFound, nil)}}})
	vault := cardOf(t, m, ElementVault)
	if !strings.Contains(strings.Join(vault.Lines, "\n"), "— folder not found") {
		t.Fatalf("lines = %v", vault.Lines)
	}
}

// ---------------------------------------------------------------------------
// the GitHub card — one row per CONFIGURED repo, the seal truth
// ---------------------------------------------------------------------------

func sealStatus(isGit, hasCommits bool, dirty *int, date string) *storage.OrderedMap {
	st := storage.NewOrderedMap()
	st.Set("is_git_repo", isGit)
	st.Set("has_commits", hasCommits)
	if dirty != nil {
		st.Set("dirty_count", *dirty)
	}
	if date != "" {
		last := storage.NewOrderedMap()
		last.Set("date", date)
		st.Set("last_commit", last)
	}
	return st
}

func TestGitHubCardOneRowPerRepoWithTheSealTruth(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{GitHub: &SyncGitHubFacts{
		Rows: []SyncBackupRepoRow{
			{Kind: "repos", RepoName: "vault-backup",
				Status: sealStatus(true, true, ipApp(3), "2020-01-02T03:04:05")},
			{Kind: "websites", RepoName: "web-backup",
				Status: sealStatus(true, true, ipApp(0), "2020-01-02T03:04:05")},
		},
	}})
	gh := cardOf(t, m, ElementGitHub)
	if gh.State != SyncStateOK || gh.Pending != 3 {
		t.Fatalf("card = %+v", gh)
	}
	if gh.StatusText != "2 backup repos configured" {
		t.Fatalf("status = %q", gh.StatusText)
	}
	if gh.Hero.Value != "3" || gh.Hero.Tone != "warning" {
		t.Fatalf("hero = %+v", gh.Hero)
	}
	joined := strings.Join(gh.Lines, "\n")
	for _, want := range []string{"📦 vault-backup", "3 files changed since",
		"📦 web-backup", "no changes since — sealed clean"} {
		if !strings.Contains(joined, want) {
			t.Fatalf("lines missing %q: %v", want, gh.Lines)
		}
	}
}

func TestGitHubCardNeverSealedAndBrokenStatusGrammars(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{GitHub: &SyncGitHubFacts{
		Rows: []SyncBackupRepoRow{
			{RepoName: "never", Status: sealStatus(true, false, nil, "")},
			{RepoName: "notgit", Status: sealStatus(false, false, nil, "")},
			{RepoName: "broke", Status: nil, Error: "the git read failed"},
		},
	}})
	gh := cardOf(t, m, ElementGitHub)
	joined := strings.Join(gh.Lines, "\n")
	for _, want := range []string{
		"never sealed — no commit has ever landed",
		"not sealed yet — the folder is not a git repo",
		"status unknown — the git read failed",
	} {
		if !strings.Contains(joined, want) {
			t.Fatalf("lines missing %q: %v", want, gh.Lines)
		}
	}
	if gh.Hero.Value != "—" || gh.Hero.Label != "no seal measured yet" {
		t.Fatalf("hero = %+v", gh.Hero)
	}
}

func TestGitHubCardSilenceWhenUnconfigured(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{GitHub: &SyncGitHubFacts{}})
	gh := cardOf(t, m, ElementGitHub)
	if gh.State != SyncStateNotSet ||
		!containsLine(gh.Lines, "nothing to show — a backup repo is set in Settings → Backup") {
		t.Fatalf("card = %+v (an unconfigured repo is silence, never an error)", gh)
	}
}

// ---------------------------------------------------------------------------
// the model itself — the flow order, the calm line, junk tolerance
// ---------------------------------------------------------------------------

func TestTheFlowOrderIsTheOwnersNumbering(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{})
	if len(m.Cards) != 4 {
		t.Fatalf("cards = %d", len(m.Cards))
	}
	want := []string{ElementBot, ElementVault, ElementGitHub, ElementWebsite}
	for i, key := range want {
		if m.Cards[i].Key != key {
			t.Fatalf("card[%d] = %q, want %q", i, m.Cards[i].Key, key)
		}
	}
	if m.Title != "Sync — the pipeline at a glance" || m.Legend != SyncLegend {
		t.Fatalf("title/legend = %q / %q", m.Title, m.Legend)
	}
}

func TestAllSyncedRendersCalm(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{
		Bot: &SyncBotFacts{State: SyncStateOK, Pending: nil}})
	if !m.AllSynced || m.CalmLine != "All synced ✅ — nothing is waiting anywhere" ||
		m.WaitingTotal != 0 {
		t.Fatalf("model = %+v", m)
	}
}

func TestCalmRequiresSomethingMeasured(t *testing.T) {
	// nothing measured at all → no calm line (the honest unknown)
	m := SyncDisplayModelOf(&SyncFacts{})
	if m.AllSynced || m.CalmLine != "" {
		t.Fatalf("model = %+v (calm requires something measured)", m)
	}
}

func TestWaitingTotalSumsEveryCardsPending(t *testing.T) {
	m := SyncDisplayModelOf(&SyncFacts{
		Bot: &SyncBotFacts{State: SyncStateOK, Pending: []any{
			pendRow("https://github.com/a/r")}},
		Website: &SyncWebsiteFacts{
			Vault:     SyncWebsiteVaultSide{State: SyncStateOK, PublicReady: ipApp(9)},
			Published: SyncPublishedFacts{State: SyncStateNever},
		},
	})
	if m.WaitingTotal != 10 {
		t.Fatalf("waiting = %d, want 1 (bot) + 9 (website)", m.WaitingTotal)
	}
	if m.AllSynced {
		t.Fatal("waiting items must kill the calm line")
	}
}

func TestJunkFactsNeverCrash(t *testing.T) {
	m := SyncDisplayModelOf(nil)
	if m == nil || len(m.Cards) != 4 {
		t.Fatalf("model = %+v", m)
	}
	for _, c := range m.Cards {
		if c.State != SyncStateNotSet {
			t.Fatalf("junk card %q = %+v (never a fabricated count)", c.Key, c)
		}
	}
	// a broken bot facts shape degrades to its not-set grammar too
	m2 := SyncDisplayModelOf(&SyncFacts{Bot: &SyncBotFacts{State: "garbage"}})
	if bot := cardOf(t, m2, ElementBot); bot.State != SyncStateNotSet {
		t.Fatalf("bot = %+v", bot)
	}
}

func TestDetectScanVaultsTheThreeRows(t *testing.T) {
	dir := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", dir)
	cfg.Set("website_vault_path", "/no/such/vault")
	cfg.Set("manual_vault_path", dir)
	rows := DetectScanVaults(cfg, true)
	if len(rows) != 3 {
		t.Fatalf("rows = %d", len(rows))
	}
	// Websites: not found at the path
	if rows[0].Kind != "websites" || rows[0].Status != VaultStatusNotFound ||
		rows[0].Selectable {
		t.Fatalf("websites row = %+v", rows[0])
	}
	// Repos: connected + counted + selectable
	if rows[1].Kind != "repos" || rows[1].Status != VaultStatusConnected ||
		!rows[1].Selectable || rows[1].NoteCount == nil {
		t.Fatalf("repos row = %+v", rows[1])
	}
	// Personal: connected but NEVER selectable (owner-owned)
	if rows[2].Kind != "personal" || rows[2].Status != VaultStatusConnected ||
		rows[2].Selectable || rows[2].Scannable {
		t.Fatalf("personal row = %+v", rows[2])
	}
	if rows[2].Why != PersonalVaultWhy {
		t.Fatalf("why = %q", rows[2].Why)
	}
	// an unset path reads not set, never a bogus path
	cfg2 := storage.NewOrderedMap()
	cfg2.Set("vault_path", 42) // junk in the seat
	if rows2 := DetectScanVaults(cfg2, false); rows2[1].Status != VaultStatusNotSet {
		t.Fatalf("junk seat row = %+v", rows2[1])
	}
}

func ipApp(n int) *int { return &n }

func containsLine(lines []string, want string) bool {
	for _, l := range lines {
		if l == want {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// the local facts over a fixture vault (the seal rows + the web side)
// ---------------------------------------------------------------------------

func TestCollectLocalFactsSilenceAndSides(t *testing.T) {
	// no configured repos → the GitHub half is silence; no website
	// vault → the website side is not_set
	cfg := storage.NewOrderedMap()
	facts := CollectLocalFacts(cfg, nil)
	if len(facts.GitHub) != 0 {
		t.Fatalf("github rows = %+v (silence, never an error)", facts.GitHub)
	}
	if facts.Website.State != SyncStateNotSet || facts.Website.PublicReady != nil {
		t.Fatalf("website side = %+v", facts.Website)
	}
	if len(facts.Vaults) != 3 {
		t.Fatalf("vault rows = %d", len(facts.Vaults))
	}
}

func TestCollectLocalFactsTheConfiguredRepoRows(t *testing.T) {
	dir := t.TempDir()
	cfg := storage.NewOrderedMap()
	cfg.Set("vault_path", dir)
	vs := storage.NewOrderedMap()
	vs.Set("repo_name", "the-backup")
	cfg.Set("vaultseal", vs)
	facts := CollectLocalFacts(cfg, nil)
	if len(facts.GitHub) != 1 {
		t.Fatalf("github rows = %+v", facts.GitHub)
	}
	row := facts.GitHub[0]
	if row.RepoName != "the-backup" || row.Kind != "repos" ||
		row.Label != "GitHub Repos vault backup" {
		t.Fatalf("row = %+v", row)
	}
	if row.Status == nil {
		t.Fatal("the seal status must be measured (local git read)")
	}
	// a never-sealed folder: not a git repo yet — spoken, never an error
	if isGit, _ := row.Status.Get("is_git_repo"); isGit != false {
		t.Fatalf("status = %v", isGit)
	}
}

// ---------------------------------------------------------------------------
// the published facts door — the four truths (the code-aware GET)
// ---------------------------------------------------------------------------

func TestFetchPublishedFactsTheFourTruths(t *testing.T) {
	cfg := storage.NewOrderedMap()
	// 1. no worker URL → not_configured
	got := FetchPublishedFacts(cfg, nil, nil)
	if got.State != SyncStateNotSet ||
		!strings.Contains(got.Message, "no worker URL configured") {
		t.Fatalf("got = %+v", got)
	}
	cfg.Set("directory_worker_url", "https://w.example")
	// 2. the Worker's 404 → never_published
	got = FetchPublishedFacts(cfg, func(u string, to float64) (int, []byte, error) {
		return 404, nil, nil
	}, nil)
	if got.State != SyncStateNever ||
		!strings.Contains(got.Message, "never published") {
		t.Fatalf("got = %+v", got)
	}
	// 3. other codes → error, spoken
	got = FetchPublishedFacts(cfg, func(u string, to float64) (int, []byte, error) {
		return 500, nil, nil
	}, nil)
	if got.State != SyncStateError || got.Message != "HTTP 500" {
		t.Fatalf("got = %+v", got)
	}
	// 4. ok — the counts + the stamp
	body := `{"counts": {"sites": 644}, "generated_at": "2026-10-01T00:00:00"}`
	got = FetchPublishedFacts(cfg, func(u string, to float64) (int, []byte, error) {
		if !strings.HasSuffix(u, syncreview.PublishedIndexPath) {
			t.Fatalf("url = %q", u)
		}
		return 200, []byte(body), nil
	}, nil)
	if got.State != SyncStateOK || got.Sites == nil || *got.Sites != 644 ||
		got.GeneratedAt != "2026-10-01T00:00:00" {
		t.Fatalf("got = %+v", got)
	}
	// junk payloads → error, never a crash
	got = FetchPublishedFacts(cfg, func(u string, to float64) (int, []byte, error) {
		return 200, []byte(`{"no": "counts"}`), nil
	}, nil)
	if got.State != SyncStateError ||
		!strings.Contains(got.Message, "no site count") {
		t.Fatalf("got = %+v", got)
	}
}
