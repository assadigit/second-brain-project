// usermemory.go — THE USER CONFIG FILE (Documents/.gitcurator): the port
// of core/user_config.py — credentials + vault identity that survive
// version updates. THE FILL (read side) and THE MIRROR (write side).
// Never raises anywhere (a broken memory is {} on read, false on write).
//
// The tree representation rides storage.OrderedMap (the T-12 law: no
// second config representation) — the memory file's key order is behavior
// (the mirror writes through write_config_file, preserving insertion
// order), so reads decode via storage.DecodeJSON.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"unicode"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// ---------------------------------------------------------------------------
// The credential key sets (user_config.py:65-126).
// ---------------------------------------------------------------------------

// EnvOverride mirrors ENV_OVERRIDE: the tests/CI door (an explicit path
// is a promise — its parent IS created).
const EnvOverride = "GITCURATOR_USER_CONFIG"

// UserConfigFilename mirrors USER_CONFIG_FILENAME (the owner's own
// suggestion): the file's name inside Documents.
const UserConfigFilename = ".gitcurator"

// UserConfigMarker mirrors USER_CONFIG_MARKER: the self-describing header
// key — written into the file, never read back as a credential.
const UserConfigMarker = "_gitcurator_user_config"

// CredentialFlatKeys mirrors CREDENTIAL_FLAT_KEYS (14 keys — the spec's
// "15" was a counting slip, reality-fixed in-commit).
var CredentialFlatKeys = []string{
	"telegram_api_id",
	"telegram_api_hash",
	"telegram_phone",
	"telegram_user_id",
	"github_token",
	"bot_token",
	"bot_username",
	"cloud_api_url",
	"cloud_api_key",
	"cloud_model",
	"llamacpp_api_url",
	"llamacpp_api_key",
	"llamacpp_model",
	"embedding_model",
}

// CredentialNestedKeys mirrors CREDENTIAL_NESTED_KEYS: merged
// key-by-key, never wholesale.
var CredentialNestedKeys = []string{"proxy", "ollama", "models"}

// CredentialPrefixes mirrors CREDENTIAL_PREFIXES: paired setup + secrets
// other layers mint ride along automatically — a new cloudflare_* key
// added tomorrow is remembered with zero changes here.
var CredentialPrefixes = []string{"cloudflare_", "gdrive_", "hmac_"}

// VaultFlatKeys mirrors VAULT_FLAT_KEYS (v0.83.0 — the vault identity
// family; NOT credential keys: is_credential_key stays credentials-only).
var VaultFlatKeys = []string{
	"vault_path",
	"vaults_history",
	"website_vault_path",
	"manual_vault_path",
	"website_repo_name",
	"manual_repo_name",
}

// VaultNestedKeys mirrors VAULT_NESTED_KEYS.
var VaultNestedKeys = []string{"vaultseal", "goodrepos"}

// IsVaultKey mirrors is_vault_key.
func IsVaultKey(key any) bool {
	s, ok := key.(string)
	if !ok || strings.HasPrefix(s, "_") {
		return false
	}
	return keyIn(s, VaultFlatKeys) || keyIn(s, VaultNestedKeys)
}

// IsRememberedKey mirrors is_remembered_key: a credential key OR a vault
// identity key.
func IsRememberedKey(key any) bool {
	return IsCredentialKey(key) || IsVaultKey(key)
}

// IsCredentialKey mirrors is_credential_key.
func IsCredentialKey(key any) bool {
	s, ok := key.(string)
	if !ok || strings.HasPrefix(s, "_") {
		return false
	}
	if keyIn(s, CredentialFlatKeys) || keyIn(s, CredentialNestedKeys) {
		return true
	}
	for _, p := range CredentialPrefixes {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

func keyIn(key string, list []string) bool {
	for _, k := range list {
		if k == key {
			return true
		}
	}
	return false
}

// ---------------------------------------------------------------------------
// Unset / real value semantics (user_config.py:155-183).
// ---------------------------------------------------------------------------

// IsUnsetValue mirrors is_unset_value: the fill/mirror grammar for one
// value. Unset (the memory may speak): nil, blank string, the
// "YOUR_*_HERE" example placeholder, zero numbers, empty lists/dicts.
// Real (the owner's word): everything else — including every boolean,
// because a deliberate OFF is a choice, not a blank. Python numbers
// compare numerically across int/float (json.Number literals included).
func IsUnsetValue(value any) bool {
	switch vv := value.(type) {
	case nil:
		return true
	case bool:
		return false
	case string:
		s := pyTrimSpace(vv)
		if s == "" {
			return true
		}
		return strings.HasPrefix(s, "YOUR_") && strings.HasSuffix(s, "_HERE")
	case int:
		return vv == 0
	case int64:
		return vv == 0
	case float64:
		return vv == 0
	case json.Number:
		// decoded memory values keep their literal shapes
		if f, err := vv.Float64(); err == nil {
			return f == 0
		}
		return false
	case []any:
		return len(vv) == 0
	case map[string]any:
		return len(vv) == 0
	case *storage.OrderedMap:
		return vv.Len() == 0
	default:
		return false
	}
}

// pyTrimSpace mirrors Python str.strip()'s whitespace set (Unicode,
// including U+00A0/\x1c — the same law the taxonomy port pins).
func pyTrimSpace(s string) string {
	return strings.TrimFunc(s, pyIsSpaceRune)
}

func pyIsSpaceRune(r rune) bool {
	switch r {
	case '\x1c', '\x1d', '\x1e', '\x1f':
		return true
	}
	return unicode.IsSpace(r)
}

// isReal mirrors _real.
func isReal(value any) bool { return !IsUnsetValue(value) }

// pyEqual compares two config values with Python == semantics: numeric
// cross-type equality (0 == 0.0; bool is an int subclass — True == 1),
// string equality, order-INSENSITIVE map equality, order-sensitive list
// equality. Used by the fill/mirror change detection
// (config.get(key) != gv) — a value that already equals the memory's
// word does not count as a change.
func pyEqual(a, b any) bool {
	if a == nil || b == nil {
		return a == nil && b == nil
	}
	if an, aok := pyNumber(a); aok {
		if bn, bok := pyNumber(b); bok {
			return an == bn
		}
		return false
	}
	switch av := a.(type) {
	case string:
		bs, ok := b.(string)
		return ok && av == bs
	case bool:
		// bools compare with numbers as 0/1
		if bn, bok := pyNumber(b); bok {
			return float64(boolToInt(av)) == bn
		}
		bs, ok := b.(bool)
		return ok && av == bs
	case *storage.OrderedMap:
		bm, ok := b.(*storage.OrderedMap)
		if !ok || av.Len() != bm.Len() {
			return false
		}
		for _, k := range av.Keys() {
			bv, exists := bm.Get(k)
			if !exists {
				return false
			}
			avv, _ := av.Get(k)
			if !pyEqual(avv, bv) {
				return false
			}
		}
		return true
	case []any:
		bl, ok := b.([]any)
		if !ok || len(av) != len(bl) {
			return false
		}
		for i := range av {
			if !pyEqual(av[i], bl[i]) {
				return false
			}
		}
		return true
	default:
		return false
	}
}

// pyNumber extracts a float64 from Python-numeric shapes (int, int64,
// float64, json.Number).
func pyNumber(v any) (float64, bool) {
	switch vv := v.(type) {
	case int:
		return float64(vv), true
	case int64:
		return float64(vv), true
	case float64:
		return vv, true
	case bool:
		return float64(boolToInt(vv)), true
	case json.Number:
		if f, err := vv.Float64(); err == nil {
			return f, true
		}
		return 0, false
	default:
		return 0, false
	}
}

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// ---------------------------------------------------------------------------
// The path law (user_config.py:190-253).
// ---------------------------------------------------------------------------

// GlobalConfigPath mirrors global_config_path: the user-level
// .gitcurator path, or "" when the feature sleeps (no Documents folder,
// no env override). The env override pins the file to an exact path and
// CREATES its parent (an explicit path is a promise, not a guess).
func GlobalConfigPath() string {
	env := strings.TrimSpace(os.Getenv(EnvOverride))
	if env != "" {
		path := absExpandUser(env)
		parent := filepath.Dir(path)
		if parent != "" && parent != "." {
			if _, err := os.Stat(parent); err != nil {
				if err := os.MkdirAll(parent, 0o755); err != nil {
					return ""
				}
			}
		}
		return path
	}
	docs := documentsDir()
	if docs == "" {
		return ""
	}
	return filepath.Join(docs, UserConfigFilename)
}

// absExpandUser mirrors os.path.abspath(os.path.expanduser(env)) for the
// shapes an env value takes: a leading "~" expands to the home dir; the
// result is made absolute (relative paths resolve against the cwd).
func absExpandUser(p string) string {
	if p == "~" || strings.HasPrefix(p, "~/") || strings.HasPrefix(p, "~\\") {
		home, err := os.UserHomeDir()
		if err == nil && home != "" {
			p = filepath.Join(home, p[1:])
		}
	}
	if abs, err := filepath.Abs(p); err == nil {
		return abs
	}
	return p
}

// ---------------------------------------------------------------------------
// Read (user_config.py:260-279).
// ---------------------------------------------------------------------------

// LoadGlobalConfig mirrors load_global_config: the user file as an
// ordered tree — empty when absent, broken or not a dict. Never raises.
// path "" resolves through GlobalConfigPath().
func LoadGlobalConfig(path string) *storage.OrderedMap {
	if path == "" {
		path = GlobalConfigPath()
	}
	if path == "" {
		return storage.NewOrderedMap()
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return storage.NewOrderedMap()
	}
	om, err := storage.DecodeJSON(raw)
	if err != nil {
		return storage.NewOrderedMap()
	}
	return om
}

// RememberedCount mirrors remembered_count: how many real values the
// memory currently holds (credentials + vault identity).
func RememberedCount(path string) int {
	g := LoadGlobalConfig(path)
	n := 0
	for _, k := range g.Keys() {
		v, _ := g.Get(k)
		if IsRememberedKey(k) && isReal(v) {
			n++
		}
	}
	return n
}

// ---------------------------------------------------------------------------
// THE FILL (user_config.py:286-374).
// ---------------------------------------------------------------------------

// ApplyGlobalCredentials mirrors apply_global_credentials: fill UNSET
// credential + vault identity keys in config from the user file. fresh
// marks the fresh-version case (config.json missing/unreadable — the
// whole local config is example placeholders): every remembered key with
// a real value restores. fresh=false: only genuinely unset keys fill; a
// real local value always wins. Returns the number of values that
// actually changed. Never raises; a nil config returns 0. path ""
// resolves through GlobalConfigPath().
func ApplyGlobalCredentials(config *storage.OrderedMap, path string, fresh bool) int {
	if config == nil {
		return 0
	}
	g := LoadGlobalConfig(path)
	if g.Len() == 0 {
		return 0
	}
	filled := 0

	// Flat keys — credentials first, then the vault identity family.
	for _, key := range append(append([]string{}, CredentialFlatKeys...), VaultFlatKeys...) {
		gv, ok := g.Get(key)
		if !ok || !isReal(gv) {
			continue
		}
		cur, _ := config.Get(key)
		if fresh || IsUnsetValue(cur) {
			if !pyEqual(cur, gv) {
				config.Set(key, gv)
				filled++
			}
		}
	}

	// Nested keys — key-by-key, never wholesale. The local dict is
	// COPIED before any write (the v0.69.0 aliasing law): a
	// shallow-copied caller (the fresh-version path copies the example
	// tree) must never have its shared nested dicts corrupted.
	for _, key := range append(append([]string{}, CredentialNestedKeys...), VaultNestedKeys...) {
		gd, ok := g.Get(key)
		if !ok {
			continue
		}
		gdm, ok := gd.(*storage.OrderedMap)
		if !ok {
			continue
		}
		var ld *storage.OrderedMap
		if cur, ok := config.Get(key); ok {
			if cm, isMap := cur.(*storage.OrderedMap); isMap {
				ld = cm.Clone() // never mutate the caller's dict
			}
		}
		if ld == nil {
			ld = storage.NewOrderedMap()
		}
		changedHere := false
		for _, sk := range gdm.Keys() {
			sv, _ := gdm.Get(sk)
			if !isReal(sv) {
				continue
			}
			cur, exists := ld.Get(sk)
			if fresh || !exists || IsUnsetValue(cur) {
				if !exists || !pyEqual(cur, sv) {
					ld.Set(sk, sv)
					filled++
					changedHere = true
				}
			}
		}
		if changedHere {
			config.Set(key, ld)
		}
	}

	// Prefix keys already in config (snapshot of the key order — Python
	// iterates list(config.keys())).
	for _, key := range config.Keys() {
		if keyIn(key, CredentialFlatKeys) || keyIn(key, CredentialNestedKeys) {
			continue
		}
		if !IsCredentialKey(key) {
			continue
		}
		gv, ok := g.Get(key)
		if !ok || !isReal(gv) {
			continue
		}
		cur, _ := config.Get(key)
		if fresh || IsUnsetValue(cur) {
			if !pyEqual(cur, gv) {
				config.Set(key, gv)
				filled++
			}
		}
	}

	// Remembered keys the memory holds that config doesn't know at all
	// (a key retired here, or minted by a newer version) — in the
	// memory file's key order.
	for _, key := range g.Keys() {
		if _, exists := config.Get(key); exists {
			continue
		}
		gv, _ := g.Get(key)
		if !IsRememberedKey(key) || !isReal(gv) {
			continue
		}
		config.Set(key, gv)
		filled++
	}

	return filled
}

// ---------------------------------------------------------------------------
// THE MIRROR (user_config.py:381-460).
// ---------------------------------------------------------------------------

// hasRemembered mirrors _has_remembered.
func hasRemembered(data *storage.OrderedMap) bool {
	for _, k := range data.Keys() {
		v, _ := data.Get(k)
		if IsRememberedKey(k) && isReal(v) {
			return true
		}
	}
	return false
}

// restrictPermissions mirrors _restrict_permissions: best-effort chmod
// 600 — owner-only on POSIX, a no-op on Windows (the profile folder
// already scopes the file to the user).
func restrictPermissions(path string) {
	_ = os.Chmod(path, 0o600)
}

// MirrorCredentialsToGlobal mirrors mirror_credentials_to_global: write
// REAL credential + vault identity values from config into the user
// file. The memory never loses: an unset local value never erases a real
// remembered one; keys only gain or update. The write is atomic and
// dry-run aware (storage.WriteConfigFile), then chmod 600 best-effort.
// The file is created only when there is something real to remember (or
// it already exists). True when the file on disk is (now) current; false
// when the feature sleeps or the write fails. Never raises.
func MirrorCredentialsToGlobal(config *storage.OrderedMap, path string) bool {
	if path == "" {
		path = GlobalConfigPath()
	}
	if path == "" || config == nil {
		return false
	}

	out := LoadGlobalConfig(path)
	changed := false

	for _, key := range append(append([]string{}, CredentialFlatKeys...), VaultFlatKeys...) {
		v, ok := config.Get(key)
		if !ok {
			v = nil
		}
		cur, _ := out.Get(key)
		if isReal(v) && !pyEqual(cur, v) {
			out.Set(key, v)
			changed = true
		}
	}

	for _, key := range append(append([]string{}, CredentialNestedKeys...), VaultNestedKeys...) {
		v, ok := config.Get(key)
		if !ok {
			continue
		}
		vm, isMap := v.(*storage.OrderedMap)
		if !isMap {
			continue
		}
		curAny, exists := out.Get(key)
		var cur *storage.OrderedMap
		if exists {
			if cm, ok := curAny.(*storage.OrderedMap); ok {
				cur = cm
			}
		}
		if cur == nil {
			cur = storage.NewOrderedMap()
		}
		subChanged := false
		for _, sk := range vm.Keys() {
			sv, _ := vm.Get(sk)
			curV, curExists := cur.Get(sk)
			var curAny any
			if curExists {
				curAny = curV
			}
			if isReal(sv) && (!curExists || !pyEqual(curAny, sv)) {
				cur.Set(sk, sv)
				subChanged = true
			}
		}
		if subChanged {
			out.Set(key, cur)
			changed = true
		}
	}

	for _, key := range config.Keys() {
		if keyIn(key, CredentialFlatKeys) || keyIn(key, CredentialNestedKeys) ||
			keyIn(key, VaultFlatKeys) || keyIn(key, VaultNestedKeys) {
			continue
		}
		if !IsCredentialKey(key) {
			continue
		}
		v, _ := config.Get(key)
		cur, _ := out.Get(key)
		if isReal(v) && !pyEqual(cur, v) {
			out.Set(key, v)
			changed = true
		}
	}

	fileExists := fileExistsAt(path)
	if !changed {
		return fileExists // already current (or nothing to say)
	}
	if !fileExists && !hasRemembered(out) {
		return false // nothing real to remember yet
	}

	out.Set(UserConfigMarker, 1)
	if err := storage.WriteConfigFile(path, out); err != nil {
		return false
	}
	restrictPermissions(path)
	return true
}

func fileExistsAt(path string) bool {
	fi, err := os.Stat(path)
	return err == nil && !fi.IsDir()
}

// ---------------------------------------------------------------------------
// The status line (user_config.py:467-481).
// ---------------------------------------------------------------------------

// StatusLine mirrors status_line: one human line for the Settings →
// Credentials page. path "" resolves through GlobalConfigPath().
func StatusLine(path string) string {
	if path == "" {
		path = GlobalConfigPath()
	}
	if path == "" {
		return "Credentials are saved inside this app folder only — " +
			"no Documents folder was found, so nothing carries " +
			"over between versions."
	}
	n := RememberedCount(path)
	if n > 0 {
		plural := "s"
		if n == 1 {
			plural = ""
		}
		return "🔑 " + strconv.Itoa(n) + " credential" + plural + " remembered in " +
			path + " — a new app version restores them automatically " +
			"on first launch."
	}
	return "Credentials will be remembered in " + path + " once you enter " +
		"them — a new app version restores them automatically " +
		"on first launch."
}
