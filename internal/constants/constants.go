// Package constants ports app/gitcurator/constants.py — the single source
// of app-wide constants: APP_DIR anchoring, ownership stamps, the legacy
// COLORS palette, the CONFIG_EXAMPLE defaults contract, the ordered category
// map, the default analysis prompt, and the expected Worker version.
//
// Spec: docs/port-blueprint/module-specs/spec-core-constants.md;
// task T-10 (Phase 1). Parity is pinned by the generated goldens in
// testdata/ (see constants_test.go).
package constants

import (
	"os"
	"path/filepath"
	"strings"
)

// ---------------------------------------------------------------------------
// APP_DIR anchoring (constants.py:28-30)
// ---------------------------------------------------------------------------

// AppDir mirrors constants.py APP_DIR: the application root that anchors
// every relative path. Python anchors to the folder containing the
// gitcurator package; the Go executable plays that role, with one ADDITIVE
// escape hatch Python never had: GITCURATOR_HOME (tests and staged installs).
// Default behavior is identical to Python (spec compat req 3). Resolved per
// call (cheap: one readlink + one env read) so tests can flip the hatch.
func AppDir() string {
	if home := strings.TrimSpace(os.Getenv("GITCURATOR_HOME")); home != "" {
		if abs, err := filepath.Abs(home); err == nil {
			return abs
		}
		return home
	}
	exe, err := os.Executable()
	if err != nil {
		wd, _ := os.Getwd()
		return wd
	}
	return filepath.Dir(exe)
}

// ConfigFile mirrors constants.py CONFIG_FILE = APP_DIR/config.json.
func ConfigFile() string { return filepath.Join(AppDir(), "config.json") }

// ---------------------------------------------------------------------------
// Ownership stamps (constants.py:37-43) — byte-exact.
// ---------------------------------------------------------------------------

const (
	// ManagedByGitcurator mirrors MANAGED_BY_GITCURATOR.
	ManagedByGitcurator = "gitcurator"
	// NoteSchemaVersion mirrors NOTE_SCHEMA_VERSION.
	NoteSchemaVersion = "1"
	// GitHubPromptVersion mirrors GITHUB_PROMPT_VERSION.
	GitHubPromptVersion = "gh-v1"
	// OwnershipBanner mirrors OWNERSHIP_BANNER (em-dash included).
	OwnershipBanner = "> [!info] Managed by GitCurator — machine-written note."
	// ExpectedWorkerVersion mirrors EXPECTED_WORKER_VERSION (the Worker
	// health-check version gate).
	ExpectedWorkerVersion = "0.40.0"
)

// ---------------------------------------------------------------------------
// Taxonomy default (constants.py:47-55)
// ---------------------------------------------------------------------------

// DefaultTaxonomyPath mirrors DEFAULT_TAXONOMY_PATH.
func DefaultTaxonomyPath() string {
	return filepath.Join(AppDir(), "taxonomy", "website-library-categories.md")
}

// ResolveTaxonomyPath mirrors resolve_taxonomy_path: the config override
// (whitespace-only = unset) or the bundled default. Pure helper.
func ResolveTaxonomyPath(config map[string]any) string {
	var raw string
	if config != nil {
		if v, ok := config["taxonomy_path"]; ok {
			if s, ok2 := v.(string); ok2 {
				raw = s
			}
		}
	}
	if s := strings.TrimSpace(raw); s != "" {
		return s
	}
	return DefaultTaxonomyPath()
}

// ---------------------------------------------------------------------------
// COLORS (constants.py:78-121) — the legacy QSS palette; the live UI theme
// lives in gui/theme.py (ported as CSS tokens in Phase 6). Kept for parity
// of any code reading it.
// ---------------------------------------------------------------------------

var Colors = map[string]string{
	// Primary action — PASTEL MINT fill with deep-forest text
	"cta":       "#B9E3C9",
	"cta_hover": "#A8DABA",
	"cta_text":  "#17402B",
	// Secondary/read action — PASTEL VIOLET family (OUTLINED buttons)
	"primary":       "#5F54B4",
	"primary_hover": "#514699",
	"primary_dark":  "#C4BCF5",
	// Neutral (warm mauve) — muted text, borders
	"neutral":       "#6C6480",
	"neutral_hover": "#57506B",
	// Semantic — pastel fills pair with deep text; *_deep = text on light
	"error":          "#F6C6CD",
	"error_hover":    "#F0B2BC",
	"error_text":     "#5E1120",
	"error_deep":     "#AE2237",
	"warning":        "#8A5B0B",
	"warning_pastel": "#F7E7BE",
	"success":        "#1E6B4B",
	// v0.07 main-screen tokens
	"hero_fill":         "#B3A7F2",
	"hero_fill_hover":   "#A296EC",
	"hero_text":         "#241D3F",
	"danger_fill":       "#D63A24",
	"danger_fill_hover": "#C43320",
	"danger_fill_press": "#B32D1D",
	"log_well_dark":     "#17131F",
	"log_well_light":    "#FFFFFF",
	"hint_dark":         "#A6A2AC",
	"hint_light":        "#7A7288",
	// Backgrounds — cream day / plum night
	"bg_light": "#FBF8F2",
	"bg_dark":  "#221E2E",
}

// ---------------------------------------------------------------------------
// CATEGORY_FOLDERS / CATEGORY_KEYS (constants.py:315-341) — insertion order
// is behavior (the classifier prompt embeds it); keep the ordered slice.
// ---------------------------------------------------------------------------

// CategoryEntry is one ordered CATEGORY_FOLDERS row.
type CategoryEntry struct {
	Key    string
	Folder string
}

// CategoryFolders mirrors CATEGORY_FOLDERS in insertion order (24 entries).
var CategoryFolders = []CategoryEntry{
	{"Agents", "AI-Domain/Agents"},
	{"Agents/Frameworks", "AI-Domain/Agents/Frameworks"},
	{"Agents/Implementations", "AI-Domain/Agents/Implementations"},
	{"Agents/Skills", "AI-Domain/Agents/Skills"},
	{"MCP", "AI-Domain/MCP"},
	{"Skills", "AI-Domain/Skills"},
	{"LLM-Tools", "AI-Domain/LLM-Tools"},
	{"Scraping", "Tools/Scraping"},
	{"Automation", "Tools/Automation"},
	{"Dev-Tools", "Tools/Dev-Tools"},
	{"Networking", "Tools/Networking"},
	{"Media", "Tools/Media"},
	{"Guides", "Documentation/Guides"},
	{"References", "Documentation/References"},
	{"Research", "Documentation/Research"},
	{"Web-Frameworks", "Frameworks/Web-Frameworks"},
	{"Backend", "Frameworks/Backend"},
	{"Frontend", "Frameworks/Frontend"},
	{"Infrastructure", "Infrastructure"},
	{"Infrastructure/Deployment", "Infrastructure/Deployment"},
	{"Infrastructure/Cloud", "Infrastructure/Cloud"},
	{"Infrastructure/Containerization", "Infrastructure/Containerization"},
	{"Infrastructure/Documentation", "Infrastructure/Documentation"},
	{"Uncategorized", "Uncategorized"},
}

// CategoryKeys mirrors CATEGORY_KEYS = list(CATEGORY_FOLDERS.keys()).
var CategoryKeys = func() []string {
	keys := make([]string, len(CategoryFolders))
	for i, e := range CategoryFolders {
		keys[i] = e.Key
	}
	return keys
}()

var categoryFolderIndex = func() map[string]string {
	m := make(map[string]string, len(CategoryFolders))
	for _, e := range CategoryFolders {
		m[e.Key] = e.Folder
	}
	return m
}()

// CategoryFolder translates an LLM category key to its vault subfolder
// (exact match, mirroring the Python dict lookup); ok=false for unknown keys.
func CategoryFolder(key string) (folder string, ok bool) {
	folder, ok = categoryFolderIndex[key]
	return folder, ok
}

// ---------------------------------------------------------------------------
// DEFAULT_SYSTEM_PROMPT (constants.py:344-365) — built like Python: the
// category list is joined into the prompt text.
// ---------------------------------------------------------------------------

const defaultSystemPromptHead = `You are an expert software engineer and technical writer. Your task is to analyze GitHub repositories and provide detailed, structured insights.

For each project, you must output a JSON object with the following keys:
- "short_summary": A single-sentence summary of the project in LESS THAN 140 characters. State the core utility/value directly.
- "summary": A 2-3 paragraph summary explaining what the project does, its main purpose, and its key features. (This is the "What is it?" section.)
- "how_it_works": A 1-2 paragraph explanation of the project's architecture, key technologies, and how it operates.
- "core_value": A 1-2 paragraph explanation of why this project is important, what problem it solves, and its unique value proposition. If user context is provided, explain how it specifically helps that user.
- "features": A list of 3-7 key features or technologies used (e.g., ["Uses Cloudflare Workers", "Integrates with Gmail API", "Self-hosted"]).
- "difference": A 1-2 paragraph comparison with similar projects, explaining what makes this one stand out.
- "category": Choose one category from this list: `

const defaultSystemPromptTail = `
- "confidence": A number between 0 and 100 indicating how confident you are in the chosen category.
- "tags": A list of up to 5 relevant tags (e.g., ["python", "llm", "agent", "framework", "email"]).

Guidelines:
- Be factual and use the provided data (name, description, topics, owner, stars, forks, README).
- If the project is from a major tech company (Microsoft, Google, Cloudflare, etc.), highlight this.
- Choose the best-fit category even if confidence is moderate. Only use "Uncategorized" if truly unknown.
- The short_summary MUST be under 140 characters and capture the essence of the project.

After generating your analysis, cross-check your output against the README and project data. If you find any factual errors or hallucinations, correct them before returning the JSON. Do not guess or invent information not present in the source data.

Return only valid JSON. Do not include any extra text.`

// DefaultSystemPrompt mirrors DEFAULT_SYSTEM_PROMPT (the joined category
// list is part of the text, exactly as in Python).
var DefaultSystemPrompt = defaultSystemPromptHead + strings.Join(CategoryKeys, ", ") + defaultSystemPromptTail

// ---------------------------------------------------------------------------
// CONFIG_EXAMPLE (constants.py:122-305) — the defaults contract: a config
// missing a key gets exactly these values. Order in the literal mirrors the
// Python source for reviewability; consumers get a fresh deep copy.
// ---------------------------------------------------------------------------

var configExample = map[string]any{
	"telegram_api_id":   0,
	"telegram_api_hash": "",
	"telegram_phone":    "",
	"proxy": map[string]any{
		"enabled":     false,
		"type":        "socks5",
		"host":        "127.0.0.1",
		"port":        10808,
		"use_for_web": true,
	},
	"web_blocked_domains": []any{
		"x.com", "twitter.com", "t.co",
		"github.com", "gist.github.com", "github.io",
		"githubusercontent.com", "huggingface.co",
		"hf.co", "instagram.com", "instagr.am",
		"facebook.com", "fb.com", "fb.me", "fb.watch",
		"linkedin.com", "lnkd.in",
		"youtube.com", "youtu.be", "youtube-nocookie.com",
		"share.google", "drive.google.com",
		"docs.google.com", "forms.google.com",
		"tiktok.com", "threads.net", "threads.com",
		"snapchat.com", "pinterest.com", "pin.it",
		"twitch.tv", "discord.gg", "discord.com",
		"discordapp.com", "t.me", "telegram.me",
		"wa.me", "whatsapp.com", "vk.com", "bsky.app",
		"weibo.com",
	},
	"web_self_domains":         []any{"github-to-obsidian-bot.aliassadi-plus.workers.dev"},
	"web_user_agent":           "",
	"web_archive_fallback":     true,
	"web_doh_probe":            true,
	"web_impersonate_fallback": true,
	"web_impersonate_autopip":  true,
	"ollama":                   map[string]any{"base_url": "http://localhost:11434", "model": "qwythos-9b"},
	"llm_provider":             "ollama",
	"cloud_api_url":            "https://api.openai.com/v1",
	"cloud_api_key":            "",
	"cloud_model":              "gpt-4o-mini",
	"llm_max_output_tokens":    0,
	"llamacpp_api_url":         "http://127.0.0.1:8080/v1",
	"llamacpp_api_key":         "",
	"llamacpp_model":           "",
	"embedding_model":          "",
	"llm_num_ctx":              8192,
	"models":                   map[string]any{"classify": "", "analyze": ""},
	"github_token":             "",
	"vault_path":               "",
	"vaults_history":           []any{},
	"log_level":                "INFO",
	"timeout_per_repo":         60,
	"max_retries":              3,
	"delay_between_api_calls":  0.5,
	"cloudflare_enabled":       false,
	"cloudflare_worker_url":    "",
	"cloudflare_install_id":    "",
	"cloudflare_shared_secret": "",
	"cloudflare_poll_interval": 300,
	"gdrive_enabled":           false,
	"gdrive_client_id":         "",
	"gdrive_client_secret":     "",
	"gdrive_access_token":      "",
	"gdrive_refresh_token":     "",
	"gdrive_token_expiry":      0,
	"gdrive_folder_id":         "",
	"gdrive_max_backups":       10,
	"gdrive_redirect_port":     8765,
	"vaultseal": map[string]any{
		"enabled":     true,
		"repo_name":   "",
		"auto_push":   true,
		"create_repo": false,
	},
	"goodrepos": map[string]any{
		"enabled":   true,
		"repo_name": "",
		"auto_push": true,
	},
	"website_vault_path":     "",
	"manual_vault_path":      "",
	"website_repo_name":      "",
	"manual_repo_name":       "",
	"taxonomy_path":          "",
	"pipelines":              map[string]any{"github": true, "websites": false},
	"vault_hygiene_on_start": true,
}

// ConfigExample returns a fresh deep copy of the defaults contract
// (mirrors the Python consumers' CONFIG_EXAMPLE deep-copy base). The tree
// is finite and acyclic — verified by the golden — so no cycle guard.
func ConfigExample() map[string]any { return deepCopyConfig(configExample) }

func deepCopyConfig(src map[string]any) map[string]any {
	dst := make(map[string]any, len(src))
	for k, v := range src {
		switch vv := v.(type) {
		case map[string]any:
			dst[k] = deepCopyConfig(vv)
		case []any:
			lst := make([]any, len(vv))
			for i, e := range vv {
				if em, ok := e.(map[string]any); ok {
					lst[i] = deepCopyConfig(em)
				} else {
					lst[i] = e
				}
			}
			dst[k] = lst
		default:
			dst[k] = vv
		}
	}
	return dst
}
