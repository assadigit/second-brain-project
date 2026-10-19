package app

// syncdisplay.go — T-75 slice (b): THE VISUAL SYNC DISPLAY's engines
// (gui/sync_display.py — the v0.86.0 four-element dashboard model) +
// the picker's own vault walk (vault_pick_dialog.py's
// detect_scan_vaults, finally ported for the vault card). The model
// is PURE (no wails, no network — the network halves arrive as facts)
// and fully tested headlessly; the Qt layer's twin is the frontend's
// render (slice (c)).
//
// THE FOUR STATES (the connection truth first — never a zero that
// lies): ok = configured AND measured · not_set = not configured (the
// Settings door is spoken) · error = attempted and failed (spoken) ·
// never = configured, nothing has ever flowed (never sealed / never
// published — counted as "never", not failure).
//
// THE §5 ARITHMETIC (the owner confirmed "Yes. calculate."): V>P →
// "V−P need publishing"; V<P → "−N since the last publish (banished
// notes retire too)" — never a negative needs-publishing;
// never-published is "never"; an unreachable worker degrades to
// unknown, never a fabricated count; the calm line only when
// waiting==0 AND any_measured.

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/directorysync"
	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/vaultseal"
)

// The four elements, in flow order — the owner's own numbering.
const (
	ElementBot     = "bot"
	ElementVault   = "vault"
	ElementGitHub  = "github"
	ElementWebsite = "website"
)

// ElementOrder mirrors ELEMENT_ORDER (the Z-reading IS the pipeline).
var ElementOrder = []string{ElementBot, ElementVault, ElementGitHub,
	ElementWebsite}

// SyncElementChrome is one element's chrome: icon + name + role (the
// spec's own words).
type SyncElementChrome struct {
	Icon string
	Name string
	Role string
}

// ElementChrome mirrors ELEMENT_CHROME.
var ElementChrome = map[string]SyncElementChrome{
	ElementBot:     {"🤖", "Telegram Bot", "source of input"},
	ElementVault:   {"📓", "Obsidian Vault", "local backup"},
	ElementGitHub:  {"☁️", "GitHub Backups", "cloud backup of Obsidian"},
	ElementWebsite: {"🌐", "Website Directory", "main display and front-end"},
}

// The display states (connection truth first).
const (
	SyncStateOK     = "ok"
	SyncStateNotSet = "not_set"
	SyncStateError  = "error"
	SyncStateNever  = "never"
)

// ToneForState mirrors TONE_FOR_STATE (the dot tones).
var ToneForState = map[string]string{
	SyncStateOK:     "success",
	SyncStateNotSet: "muted",
	SyncStateError:  "error",
	SyncStateNever:  "muted",
}

// SyncLegend mirrors LEGEND: the one-line teaching of the flow.
const SyncLegend = "Links arrive at the bot → become notes in the vault → " +
	"seal to GitHub → publish to the website."

// ---------------------------------------------------------------------------
// The URL router's own split (classify_pending — the batch's grammar)
// ---------------------------------------------------------------------------

// PendingSplit is classify_pending's answer: repos + websites +
// unrouted == total, always.
type PendingSplit struct {
	Total    int `json:"total"`
	Repos    int `json:"repos"`
	Websites int `json:"websites"`
	Unrouted int `json:"unrouted"`
}

// ClassifyPending is classify_pending — PURE: the SAME router the
// batch itself routes by (a github.com/<owner>/<repo> link or an
// <owner>.github.io/<repo> page is a repo; every other link — gists
// included, the intake's own law — is a website). Junk rows (no
// parseable URL) count in the total and in neither bucket — spoken,
// never guessed.
func ClassifyPending(pending []any) PendingSplit {
	repos, websites, unrouted := 0, 0, 0
	for _, item := range pending {
		url := queueItemURL(item)
		if url == "" {
			unrouted++
		} else if links.IsGitHubURL(url) || links.MapGitHubIOURL(url) != "" {
			repos++
		} else {
			websites++
		}
	}
	return PendingSplit{Total: len(pending), Repos: repos,
		Websites: websites, Unrouted: unrouted}
}

// ---------------------------------------------------------------------------
// The picker's own vault walk (detect_scan_vaults — PURE, ported for
// the vault card; the labels/icons/statuses verbatim)
// ---------------------------------------------------------------------------

// The vault kinds the picker speaks, in display order.
const (
	VaultKindWebsites = "websites"
	VaultKindRepos    = "repos"
	VaultKindPersonal = "personal"
)

// vaultConfigKeys mirrors VAULT_CONFIG_KEYS (the Settings page's own
// keys), in display order.
var vaultConfigKeys = []struct{ Kind, Key string }{
	{VaultKindWebsites, "website_vault_path"},
	{VaultKindRepos, "vault_path"},
	{VaultKindPersonal, "manual_vault_path"},
}

// vaultLabelAndIcon mirrors VAULT_LABELS + VAULT_ICONS.
var vaultLabelAndIcon = map[string]struct{ Label, Icon string }{
	VaultKindWebsites: {"Websites vault", "🌐"},
	VaultKindRepos:    {"GitHub Repos vault", "🐙"},
	VaultKindPersonal: {"Personal vault", "✍️"},
}

// PersonalVaultWhy mirrors PERSONAL_VAULT_WHY: the Personal vault's
// own honest law, spoken on its row (never tickable, never scanned).
const PersonalVaultWhy = "Owner-owned — the curator mirrors Library/ into it " +
	"read-only and never proposes a move or a deletion in your own vault " +
	"(Settings → 📁 Vault)"

// The status vocabulary (the connection truth, never a guess).
const (
	VaultStatusConnected = "connected"
	VaultStatusNotFound  = "not_found"
	VaultStatusNotSet    = "not_set"
)

// SyncVaultRow is one vault's detection row (detect_scan_vaults's
// shape).
type SyncVaultRow struct {
	Kind       string `json:"kind"`
	Label      string `json:"label"`
	Icon       string `json:"icon"`
	Path       string `json:"path"`
	Status     string `json:"status"`
	NoteCount  *int   `json:"note_count"`
	Scannable  bool   `json:"scannable"`
	Selectable bool   `json:"selectable"`
	Why        string `json:"why"`
}

// countVaultNotes mirrors _count_notes: every .md under the tree,
// skipping the dot-directories (.trash, .obsidian, .git — the
// machine's own scaffolding is not the library). nil when the walk
// fails.
func countVaultNotes(path string) *int {
	total := 0
	ok := true
	func() {
		defer func() {
			if recover() != nil {
				ok = false
			}
		}()
		var walk func(dir string)
		walk = func(dir string) {
			dirents, err := os.ReadDir(dir)
			if err != nil {
				return
			}
			for _, e := range dirents {
				if e.IsDir() {
					if strings.HasPrefix(e.Name(), ".") {
						continue
					}
					walk(filepathJoin(dir, e.Name()))
				} else if strings.HasSuffix(strings.ToLower(e.Name()), ".md") {
					total++
				}
			}
		}
		walk(path)
	}()
	if !ok {
		return nil
	}
	return &total
}

// DetectScanVaults is detect_scan_vaults — the three vaults the
// Settings page names, each with its connection truth. PURE (no
// writes, never raises). withCounts=false skips the note counts (the
// quick walk).
func DetectScanVaults(cfg *storage.OrderedMap, withCounts bool) []SyncVaultRow {
	rows := []SyncVaultRow{}
	for _, vc := range vaultConfigKeys {
		path := asPathOf(cfg, vc.Key)
		var status string
		switch {
		case path == "":
			status = VaultStatusNotSet
		case fsIsDirPath(path):
			status = VaultStatusConnected
		default:
			status = VaultStatusNotFound
		}
		li := vaultLabelAndIcon[vc.Kind]
		scannable := vc.Kind != VaultKindPersonal
		selectable := scannable && status == VaultStatusConnected
		var noteCount *int
		if withCounts && status == VaultStatusConnected {
			noteCount = countVaultNotes(path)
		}
		why := ""
		switch {
		case vc.Kind == VaultKindPersonal:
			why = PersonalVaultWhy
		case status == VaultStatusNotSet:
			why = "Not set yet — Settings → 📁 Vault"
		case status == VaultStatusNotFound:
			why = "The folder is not found at this path"
		}
		rows = append(rows, SyncVaultRow{
			Kind: vc.Kind, Label: li.Label, Icon: li.Icon, Path: path,
			Status: status, NoteCount: noteCount,
			Scannable: scannable, Selectable: selectable, Why: why,
		})
	}
	return rows
}

// ---------------------------------------------------------------------------
// The local facts (collect_local_facts — the worker job's own half)
// ---------------------------------------------------------------------------

// SyncBackupRepoRow is one CONFIGURED backup repo's row (the GitHub
// card's own shape): VaultSeal.status()'s dict verbatim, or nil +
// Error when the walk broke (a broken walk is spoken).
type SyncBackupRepoRow struct {
	Kind      string              `json:"kind"`
	Label     string              `json:"label"`
	RepoName  string              `json:"repo_name"`
	VaultPath string              `json:"vault_path"`
	Status    *storage.OrderedMap `json:"status"`
	Error     string              `json:"error,omitempty"`
}

// SyncWebsiteVaultSide is the Websites vault's public-ready half.
type SyncWebsiteVaultSide struct {
	State       string `json:"state"`
	PublicReady *int   `json:"public_ready"`
	Message     string `json:"message"`
}

// SyncLastBatch is the desk's own Last-sync record (the vault card's
// memory fact).
type SyncLastBatch struct {
	When  string `json:"when"`
	Done  int    `json:"done"`
	Total int    `json:"total"`
}

// SyncLocalFacts is collect_local_facts's answer — the LOCAL half
// (no network; the job adds the bot + published halves).
type SyncLocalFacts struct {
	Vaults  []SyncVaultRow       `json:"vaults"`
	GitHub  []SyncBackupRepoRow  `json:"github_rows"`
	Website SyncWebsiteVaultSide `json:"website_vault"`
}

// backupRepoRows mirrors _backup_repo_rows: one row per CONFIGURED
// backup repo (the GitHub Repos vault's own backup + the Websites
// vault's backup). An unconfigured repo is SILENCE — no row, no error
// (the v0.83.0 creation guard's law). Local git reads only.
func backupRepoRows(cfg *storage.OrderedMap, log func(string)) []SyncBackupRepoRow {
	if log == nil {
		log = func(string) {}
	}
	rows := []SyncBackupRepoRow{}
	vsCfg := cfgNested(cfg, "vaultseal")
	spec := []struct{ Kind, Label, VaultPath, RepoName string }{
		{"repos", "GitHub Repos vault backup",
			asPathOf(cfg, "vault_path"), asPathOf(vsCfg, "repo_name")},
		{"websites", "Websites vault backup",
			asPathOf(cfg, "website_vault_path"), asPathOf(cfg, "website_repo_name")},
	}
	for _, s := range spec {
		if s.RepoName == "" {
			continue // silence — not configured
		}
		row := SyncBackupRepoRow{Kind: s.Kind, Label: s.Label,
			RepoName: s.RepoName, VaultPath: s.VaultPath}
		func() {
			defer func() {
				if r := recover(); r != nil {
					log("[sync-view] seal status failed for " + s.RepoName +
						": " + pyCrashText(r))
					row.Error = pyCrashText(r)
				}
			}()
			sealer := vaultseal.NewVaultSeal(s.VaultPath, "", s.RepoName,
				true, true, 0, nil, false, nil)
			row.Status = sealer.Status()
		}()
		rows = append(rows, row)
	}
	return rows
}

// websiteVaultSide mirrors _website_vault_side: the Websites vault's
// own public-ready count — the number a publish WOULD carry, measured
// locally by the publish leg's own parser (BuildIndex — reuse, never
// re-invented). Read-only (BuildIndex never writes).
func websiteVaultSide(cfg *storage.OrderedMap, log func(string)) SyncWebsiteVaultSide {
	if log == nil {
		log = func(string) {}
	}
	vault := asPathOf(cfg, "website_vault_path")
	if vault == "" || !fsIsDirPath(vault) {
		return SyncWebsiteVaultSide{State: SyncStateNotSet}
	}
	var idx *storage.OrderedMap
	func() {
		defer func() {
			if r := recover(); r != nil {
				log("[sync-view] vault walk failed: " + pyCrashText(r))
			}
		}()
		idx = directorysync.BuildIndex(vault, nil, func(string) {})
	}()
	sites := 0
	if idx != nil {
		if countsV, ok := idx.Get("counts"); ok {
			if counts, is := countsV.(*storage.OrderedMap); is {
				if v, ok := counts.Get("sites"); ok {
					sites = pyIntOfAny(v)
				}
			}
		}
	} else {
		return SyncWebsiteVaultSide{State: SyncStateError,
			Message: "the vault walk failed"}
	}
	return SyncWebsiteVaultSide{State: SyncStateOK, PublicReady: &sites}
}

// CollectLocalFacts mirrors collect_local_facts: the LOCAL half of
// the facts (no network — the worker job's first beat, drivable
// headlessly). The network halves (the bot queue, the published
// index) are the job's to add — never here.
func CollectLocalFacts(cfg *storage.OrderedMap, log func(string)) *SyncLocalFacts {
	if cfg == nil {
		cfg = storage.NewOrderedMap()
	}
	return &SyncLocalFacts{
		Vaults:  DetectScanVaults(cfg, true),
		GitHub:  backupRepoRows(cfg, log),
		Website: websiteVaultSide(cfg, log),
	}
}

// ---------------------------------------------------------------------------
// The friendly time (the Last-sync line's own grammar)
// ---------------------------------------------------------------------------

// FriendlyWhen mirrors friendly_when: "today 14:02" for a same-day
// stamp, the plain date-time otherwise, the raw string when it does
// not parse, and "never" when there is nothing.
func FriendlyWhen(stamp string) string {
	if strings.TrimSpace(stamp) == "" {
		return "never"
	}
	raw := strings.TrimSpace(stamp)
	when, err := parsePyISODateTime(raw)
	if err != nil {
		return raw
	}
	// an aware stamp lands in the local zone (the main desk's voice)
	now := time.Now()
	if when.Format("2006-01-02") == now.Format("2006-01-02") {
		return "today " + when.Format("15:04")
	}
	return when.Format("2006-01-02 15:04")
}

// parsePyISODateTime parses datetime.fromisoformat's common shapes
// (date-only, date-time with 'T' or space, optional seconds and
// offset).
func parsePyISODateTime(raw string) (time.Time, error) {
	layouts := []string{
		"2006-01-02T15:04:05-07:00", "2006-01-02T15:04:05",
		"2006-01-02T15:04-07:00", "2006-01-02T15:04",
		"2006-01-02 15:04:05-07:00", "2006-01-02 15:04:05",
		"2006-01-02 15:04", "2006-01-02",
	}
	for _, layout := range layouts {
		if t, err := time.ParseInLocation(layout, raw, time.Local); err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unparseable stamp")
}

// ---------------------------------------------------------------------------
// The display model — PURE (the frontend only renders it)
// ---------------------------------------------------------------------------

// SyncHero is the card's ONE big number + its micro-copy label.
type SyncHero struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Tone  string `json:"tone"`
}

// SyncStat is one aligned stat row (bold value right, muted label
// left, an optional tiny note under it).
type SyncStat struct {
	Value string `json:"value"`
	Label string `json:"label"`
	Tone  string `json:"tone"`
	Note  string `json:"note"`
}

// SyncCard is one element's card (the display model's own shape).
type SyncCard struct {
	Key        string     `json:"key"`
	Icon       string     `json:"icon"`
	Name       string     `json:"name"`
	Role       string     `json:"role"`
	State      string     `json:"state"`
	DotTone    string     `json:"dot_tone"`
	StatusText string     `json:"status_text"`
	Lines      []string   `json:"lines"`
	Hint       string     `json:"hint"`
	Hero       SyncHero   `json:"hero"`
	Stats      []SyncStat `json:"stats"`
	Foot       []string   `json:"foot"`
	Pending    int        `json:"pending"`
}

// SyncBotFacts is the queue poll's half (the job's network addition).
type SyncBotFacts struct {
	State   string `json:"state"`
	Message string `json:"message"`
	Pending []any  `json:"pending"`
	Total   *int   `json:"total"`
}

// SyncPublishedFacts is the published index's half.
type SyncPublishedFacts struct {
	State       string `json:"state"`
	Sites       *int   `json:"sites"`
	GeneratedAt string `json:"generated_at"`
	Message     string `json:"message"`
}

// SyncVaultsFacts is the vaults' half (the local walk + the desk's
// last-batch memory).
type SyncVaultsFacts struct {
	Vaults    []SyncVaultRow `json:"vaults"`
	LastBatch *SyncLastBatch `json:"last_batch"`
}

// SyncWebsiteFacts is the website card's halves.
type SyncWebsiteFacts struct {
	Vault     SyncWebsiteVaultSide `json:"vault"`
	Published SyncPublishedFacts   `json:"published"`
}

// SyncGitHubFacts is the GitHub card's half.
type SyncGitHubFacts struct {
	Rows []SyncBackupRepoRow `json:"rows"`
}

// SyncFacts is the display model's full input: the local half + the
// job's network halves.
type SyncFacts struct {
	Bot     *SyncBotFacts     `json:"bot,omitempty"`
	Vaults  *SyncVaultsFacts  `json:"vaults,omitempty"`
	GitHub  *SyncGitHubFacts  `json:"github,omitempty"`
	Website *SyncWebsiteFacts `json:"website,omitempty"`
}

// SyncDisplayModel is sync_display_model's answer: the four cards in
// flow order + the calm line + the legend.
type SyncDisplayModel struct {
	Title        string     `json:"title"`
	Cards        []SyncCard `json:"cards"`
	AllSynced    bool       `json:"all_synced"`
	CalmLine     string     `json:"calm_line"`
	Legend       string     `json:"legend"`
	WaitingTotal int        `json:"waiting_total"`
}

// normalizeDisplayState mirrors _normalize_state: the doors' own
// vocabularies fold into the display's four states.
func normalizeDisplayState(state string) string {
	switch state {
	case SyncStateOK, SyncStateError:
		return state
	case "never_published", SyncStateNever:
		return SyncStateNever
	default:
		return SyncStateNotSet
	}
}

// pluralS is the tiny ” / 's' plural law.
func pluralS(n int) string {
	if n != 1 {
		return "s"
	}
	return ""
}

// omInt reads an int out of an OrderedMap seat (nil-safe).
func omInt(om *storage.OrderedMap, key string) *int {
	if om == nil {
		return nil
	}
	v, ok := om.Get(key)
	if !ok {
		return nil
	}
	switch t := v.(type) {
	case int:
		return &t
	case int64:
		n := int(t)
		return &n
	case float64:
		n := int(t)
		return &n
	}
	return nil
}

// omBool reads a bool out of an OrderedMap seat (nil-safe).
func omBool(om *storage.OrderedMap, key string) bool {
	if om == nil {
		return false
	}
	v, _ := om.Get(key)
	b, _ := v.(bool)
	return b
}

// omStr reads a string out of an OrderedMap seat (nil-safe).
func omStr(om *storage.OrderedMap, key string) string {
	if om == nil {
		return ""
	}
	v, _ := om.Get(key)
	s, _ := v.(string)
	return s
}

func heroOf(value, label, tone string) SyncHero {
	return SyncHero{Value: value, Label: label, Tone: tone}
}

func statOf(value, label, tone, note string) SyncStat {
	return SyncStat{Value: value, Label: label, Tone: tone, Note: note}
}

func numStr(n int) string { return strconv.Itoa(n) }

// botCard is _bot_card: the queue — N links waiting, split repos ·
// websites, the split the router itself makes.
func botCard(bot *SyncBotFacts) SyncCard {
	chrome := ElementChrome[ElementBot]
	state := SyncStateNotSet
	data := bot
	if data == nil {
		data = &SyncBotFacts{}
	} else {
		state = normalizeDisplayState(data.State)
	}
	split := ClassifyPending(data.Pending)
	total := split.Total
	if data.Total != nil {
		total = *data.Total
	}
	var lines []string
	var stats []SyncStat
	var hero SyncHero
	var statusText, hint string
	if state == SyncStateOK {
		if total > 0 {
			lines = append(lines,
				numStr(total)+" link"+pluralS(total)+" waiting to be synced",
				numStr(split.Repos)+" repo"+pluralS(split.Repos)+
					" · "+numStr(split.Websites)+" website"+
					pluralS(split.Websites))
			stats = append(stats,
				statOf(numStr(split.Repos), "repos waiting", "", ""),
				statOf(numStr(split.Websites), "websites waiting", "", ""))
			hero = heroOf(numStr(total), "links waiting", "warning")
		} else {
			lines = append(lines, "Queue is calm — nothing is waiting")
			hero = heroOf("0", "links waiting", "success")
		}
		statusText = "connected · queue is live"
		hint = "The SYNC button (main window) fetches and processes " +
			"the queue into notes"
	} else if state == SyncStateError {
		statusText = "could not reach the worker"
		msg := data.Message
		if msg == "" {
			msg = "the poll failed"
		}
		lines = []string{"queue unknown — " + msg}
		hero = heroOf("—", "queue unknown", "muted")
		hint = "The worker may be down or the pairing stale — " +
			"Settings → Credentials has the door"
	} else {
		statusText = "not set — Settings → Credentials has the door"
		lines = []string{"queue not measured (the worker is not configured)"}
		hero = heroOf("—", "queue not measured", "muted")
		hint = "Settings → Credentials — the worker URL and the pairing"
	}
	pending := 0
	if state == SyncStateOK {
		pending = total
	}
	return SyncCard{Key: ElementBot, Icon: chrome.Icon, Name: chrome.Name,
		Role: chrome.Role, State: state,
		DotTone: ToneForState[state], StatusText: statusText,
		Lines: lines, Hint: hint, Hero: hero, Stats: stats, Pending: pending}
}

// vaultCard is _vault_card: the three vaults + the queue's waiting
// line + the last batch's memory.
func vaultCard(vaults *SyncVaultsFacts, bot *SyncBotFacts) SyncCard {
	chrome := ElementChrome[ElementVault]
	rows := []SyncVaultRow{}
	lastBatch := (*SyncLastBatch)(nil)
	if vaults != nil {
		rows = vaults.Vaults
		lastBatch = vaults.LastBatch
	}
	anyConnected := false
	for _, r := range rows {
		if r.Status == VaultStatusConnected {
			anyConnected = true
		}
	}
	var lines, foot []string
	var stats []SyncStat
	counted := 0
	for _, row := range rows {
		icon := row.Icon
		if icon == "" {
			icon = "🗂️"
		}
		name := row.Label
		if name == "" {
			name = "Vault"
		}
		label := icon + " " + name
		short := name
		switch {
		case row.Kind == VaultKindPersonal:
			lines = append(lines, label+" — owner-owned, read-only")
			stats = append(stats, statOf("—", short+" (read-only)", "", ""))
		case row.NoteCount != nil:
			count := *row.NoteCount
			lines = append(lines, label+" — "+numStr(count)+" note"+
				pluralS(count))
			stats = append(stats, statOf(numStr(count), short+" notes", "", ""))
			counted += count
		case row.Status == VaultStatusConnected:
			lines = append(lines, label+" — connected")
			stats = append(stats, statOf("—", short, "", ""))
		case row.Status == VaultStatusNotFound:
			lines = append(lines, label+" — folder not found")
			stats = append(stats, statOf("—", short+" (not found)", "", ""))
		default:
			lines = append(lines, label+" — not set")
			stats = append(stats, statOf("—", short+" (not set)", "", ""))
		}
	}
	if bot != nil && normalizeDisplayState(bot.State) == SyncStateOK {
		split := ClassifyPending(bot.Pending)
		if split.Total > 0 {
			lines = append(lines, "↳ "+numStr(split.Total)+" link"+
				pluralS(split.Total)+" waiting to become notes")
			stats = append(stats,
				statOf(numStr(split.Total), "waiting to become notes",
					"warning", ""))
		}
	}
	if lastBatch != nil && lastBatch.When != "" {
		tail := ""
		if lastBatch.Total > 0 {
			if lastBatch.Done >= lastBatch.Total {
				tail = " · " + numStr(lastBatch.Done) + " saved"
			} else {
				tail = " · " + numStr(lastBatch.Done) + " saved · " +
					numStr(lastBatch.Total-lastBatch.Done) + " need retry"
			}
		}
		lines = append(lines, "last batch "+FriendlyWhen(lastBatch.When)+tail)
		foot = append(foot, "last batch "+FriendlyWhen(lastBatch.When)+tail)
	}
	var state, statusText, hint string
	var hero SyncHero
	if anyConnected {
		state = SyncStateOK
		statusText = "connected — the local library"
		if counted > 0 {
			hero = heroOf(numStr(counted), "notes in the local library",
				"success")
		} else {
			hero = heroOf("—", "no notes counted", "muted")
		}
		hint = "Settings → Vault has the folders; the Scan button " +
			"curates what is already filed"
	} else {
		state = SyncStateNotSet
		statusText = "not set — Settings → Vault has the door"
		hero = heroOf("—", "no vault connected", "muted")
		hint = "Settings → Vault — the three folders this desk walks"
	}
	return SyncCard{Key: ElementVault, Icon: chrome.Icon, Name: chrome.Name,
		Role: chrome.Role, State: state, DotTone: ToneForState[state],
		StatusText: statusText, Lines: lines, Hint: hint, Hero: hero,
		Stats: stats, Foot: foot}
}

// sealRowLines is _seal_row_lines: one backup repo's honest lines.
func sealRowLines(row SyncBackupRepoRow) []string {
	status := row.Status
	if status == nil {
		reason := row.Error
		if reason == "" {
			reason = "the git read failed"
		}
		return []string{"status unknown — " + reason}
	}
	if !omBool(status, "is_git_repo") {
		return []string{"not sealed yet — the folder is not a git repo"}
	}
	if !omBool(status, "has_commits") {
		return []string{"never sealed — no commit has ever landed"}
	}
	var lines []string
	lastV, _ := status.Get("last_commit")
	date := ""
	if last, ok := lastV.(*storage.OrderedMap); ok {
		date = omStr(last, "date")
	}
	lines = append(lines, "last seal "+FriendlyWhen(date))
	if dirty := omInt(status, "dirty_count"); dirty != nil {
		if *dirty > 0 {
			lines = append(lines, numStr(*dirty)+" file"+pluralS(*dirty)+
				" changed since")
		} else {
			lines = append(lines, "no changes since — sealed clean")
		}
	}
	return lines
}

// lastCommitDateOf pulls the status's last-commit date.
func lastCommitDateOf(status *storage.OrderedMap) string {
	if status == nil {
		return ""
	}
	lastV, _ := status.Get("last_commit")
	if last, ok := lastV.(*storage.OrderedMap); ok {
		return omStr(last, "date")
	}
	return ""
}

// githubCard is _github_card: one row per CONFIGURED backup repo.
func githubCard(github *SyncGitHubFacts) SyncCard {
	chrome := ElementChrome[ElementGitHub]
	rows := []SyncBackupRepoRow{}
	if github != nil {
		rows = github.Rows
	}
	if len(rows) == 0 {
		return SyncCard{Key: ElementGitHub, Icon: chrome.Icon,
			Name: chrome.Name, Role: chrome.Role, State: SyncStateNotSet,
			DotTone: "muted", StatusText: "no backup repos configured",
			Lines: []string{"nothing to show — a backup repo is set in " +
				"Settings → Backup"},
			Hint: "Settings → Backup — VaultSeal's repo names " +
				"(never created silently)",
			Hero: heroOf("—", "no backup repos", "muted")}
	}
	var lines []string
	var stats []SyncStat
	measured, pending := 0, 0
	for _, row := range rows {
		repo := row.RepoName
		if repo == "" {
			repo = "backup repo"
		}
		lines = append(lines, "📦 "+repo)
		for _, ln := range sealRowLines(row) {
			lines = append(lines, "   "+ln)
		}
		note, value, tone := "", "—", ""
		status := row.Status
		if status != nil && omBool(status, "has_commits") {
			measured++
			note = "last seal " + FriendlyWhen(lastCommitDateOf(status))
			if dirty := omInt(status, "dirty_count"); dirty != nil {
				value = numStr(*dirty)
				if *dirty > 0 {
					tone = "warning"
					pending += *dirty
				} else {
					tone = "success"
				}
			}
		} else if status != nil && !omBool(status, "is_git_repo") {
			note = "not sealed yet — not a git repo"
		} else if status != nil && !omBool(status, "has_commits") {
			note = "never sealed — no commit yet"
		} else {
			note = "status unknown — the git read failed"
		}
		stats = append(stats, statOf(value, repo+" · files changed", tone, note))
	}
	var hero SyncHero
	switch {
	case pending > 0:
		hero = heroOf(numStr(pending), "files changed since the seals", "warning")
	case measured > 0:
		hero = heroOf("0", "files changed — all mirrors clean", "success")
	default:
		hero = heroOf("—", "no seal measured yet", "muted")
	}
	return SyncCard{Key: ElementGitHub, Icon: chrome.Icon, Name: chrome.Name,
		Role: chrome.Role, State: SyncStateOK, DotTone: "success",
		StatusText: numStr(len(rows)) + " backup repo" +
			pluralS(len(rows)) + " configured",
		Lines: lines, Hero: hero, Stats: stats,
		Hint: "The Seal door (Settings → Backup) seals the " +
			"changes; a seal also rides every batch's end " +
			"and the ⚡ Fast Sync button (main window)",
		Pending: pending}
}

// websiteCard is _website_card: the §5 arithmetic (published P ·
// vault has V · V−P need publishing), the edge cases spoken honestly.
func websiteCard(website *SyncWebsiteFacts) SyncCard {
	chrome := ElementChrome[ElementWebsite]
	vaultSide := SyncWebsiteVaultSide{}
	pub := SyncPublishedFacts{}
	if website != nil {
		vaultSide = website.Vault
		pub = website.Published
	}
	vState := normalizeDisplayState(vaultSide.State)
	vCount := vaultSide.PublicReady
	pState := normalizeDisplayState(pub.State)
	pCount := pub.Sites

	if vState == SyncStateNotSet {
		return SyncCard{Key: ElementWebsite, Icon: chrome.Icon,
			Name: chrome.Name, Role: chrome.Role, State: SyncStateNotSet,
			DotTone: "muted", StatusText: "websites vault not set",
			Lines: []string{"nothing to publish until the Websites vault is set"},
			Hint:  "Settings → Vault — the Websites vault folder",
			Hero:  heroOf("—", "vault not set", "muted")}
	}

	var lines, foot []string
	var stats []SyncStat
	var hero SyncHero
	if vState == SyncStateOK && vCount != nil {
		lines = append(lines, "vault has "+numStr(*vCount)+" public-ready note"+
			pluralS(*vCount))
		hero = heroOf(numStr(*vCount), "public-ready in the vault", "success")
	} else {
		msg := vaultSide.Message
		if msg == "" {
			msg = "unknown"
		}
		lines = append(lines, "vault walk failed — "+msg)
		hero = heroOf("—", "vault walk failed", "muted")
	}

	pending := 0
	switch {
	case pState == SyncStateOK && pCount != nil:
		lines = append(lines, "published "+numStr(*pCount)+" · "+
			FriendlyWhen(pub.GeneratedAt))
		pubNote := "measured " + FriendlyWhen(pub.GeneratedAt)
		if vCount != nil {
			v, p := *vCount, *pCount
			switch {
			case v > p:
				pending = v - p
				lines = append(lines, numStr(pending)+" need publishing")
				stats = append(stats,
					statOf(numStr(p), "published on the site", "success", pubNote),
					statOf(numStr(pending), "need publishing", "warning", ""))
				foot = append(foot, numStr(v)+" in the vault · "+numStr(p)+
					" on the site — ⚡ Fast Sync publishes the gap")
			case v < p:
				lines = append(lines, "\u2212"+numStr(p-v)+
					" since the last publish (banished notes retire too)")
				stats = append(stats,
					statOf(numStr(p), "published on the site", "success", pubNote),
					statOf("\u2212"+numStr(p-v), "retired since the last publish",
						"", ""))
			default:
				lines = append(lines, "in sync with the vault")
				stats = append(stats,
					statOf(numStr(p), "published on the site", "success", pubNote),
					statOf("0", "need publishing", "success", ""))
			}
		} else {
			stats = append(stats,
				statOf(numStr(*pCount), "published on the site", "success", pubNote))
		}
	case pState == SyncStateNever:
		lines = append(lines, "never published — the site is empty until the "+
			"first publish runs")
		stats = append(stats, statOf("never", "published on the site", "", ""))
		if vCount != nil {
			pending = *vCount
			lines = append(lines, numStr(*vCount)+" would be new on the site")
			stats = append(stats, statOf(numStr(*vCount),
				"would be new on the site", "warning", ""))
		}
	case pState == SyncStateNotSet:
		lines = append(lines, "published count unknown — no worker URL set")
		stats = append(stats, statOf("—", "published on the site (no Worker URL)",
			"muted", ""))
		foot = append(foot, "set the Cloudflare Worker URL (Settings → "+
			"Credentials) — then this card speaks the site's own count")
	default:
		msg := pub.Message
		if msg == "" {
			msg = "not checked"
		}
		lines = append(lines, "published count unknown — "+msg)
		stats = append(stats, statOf("—", "published on the site (unknown)",
			"muted", ""))
	}

	var state, tone string
	switch {
	case vState == SyncStateError || pState == SyncStateError:
		state, tone = SyncStateError, "error"
	case pState == SyncStateOK || pState == SyncStateNever:
		state, tone = SyncStateOK, "success"
	case vState == SyncStateOK:
		state, tone = SyncStateOK, "success"
	default:
		state, tone = SyncStateNotSet, "muted"
	}
	return SyncCard{Key: ElementWebsite, Icon: chrome.Icon, Name: chrome.Name,
		Role: chrome.Role, State: state, DotTone: tone,
		StatusText: "the public site", Lines: lines, Hero: hero,
		Stats: stats, Foot: foot,
		Hint: "Publishing rides every batch's end AND the " +
			"⚡ Fast Sync button (main window) — the " +
			"triad's no-AI mirror on demand",
		Pending: pending}
}

// SyncDisplayModelOf is sync_display_model — PURE (no network, never
// raises): the four cards in flow order, the calm line when nothing
// is waiting anywhere. Junk tolerance is total: a missing or broken
// element degrades to its not-set grammar, never a crash, never a
// fabricated count.
func SyncDisplayModelOf(facts *SyncFacts) *SyncDisplayModel {
	if facts == nil {
		facts = &SyncFacts{}
	}
	cards := []SyncCard{
		botCard(facts.Bot),
		vaultCard(facts.Vaults, facts.Bot),
		githubCard(facts.GitHub),
		websiteCard(facts.Website),
	}
	waiting := 0
	anyMeasured := false
	for _, c := range cards {
		waiting += c.Pending
		if c.State == SyncStateOK || c.State == SyncStateError {
			anyMeasured = true
		}
	}
	calm := ""
	if waiting == 0 && anyMeasured {
		calm = "All synced ✅ — nothing is waiting anywhere"
	}
	return &SyncDisplayModel{
		Title:        "Sync — the pipeline at a glance",
		Cards:        cards,
		AllSynced:    waiting == 0 && anyMeasured,
		CalmLine:     calm,
		Legend:       SyncLegend,
		WaitingTotal: waiting,
	}
}

// ---------------------------------------------------------------------------
// tiny config readers (the _as_path / nested-dict laws)
// ---------------------------------------------------------------------------

// asPathOf mirrors _as_path: only a real string is a path (stripped) —
// a number or a list in a config seat reads as not set, never as a
// bogus path.
func asPathOf(cfg *storage.OrderedMap, key string) string {
	if cfg == nil {
		return ""
	}
	v, _ := cfg.Get(key)
	s, ok := v.(string)
	if !ok {
		return ""
	}
	return strings.TrimSpace(s)
}

// cfgNested reads a nested OrderedMap (nil-safe).
func cfgNested(cfg *storage.OrderedMap, section string) *storage.OrderedMap {
	if cfg == nil {
		return nil
	}
	v, _ := cfg.Get(section)
	om, _ := v.(*storage.OrderedMap)
	return om
}

// pyIntOfAny reads an int out of an any seat (int/int64/float64).
func pyIntOfAny(v any) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	}
	return 0
}

// filepathJoin is filepath.Join without the import dance at every
// call site of the tiny walks.
func filepathJoin(parts ...string) string {
	return filepath.Join(parts...)
}
