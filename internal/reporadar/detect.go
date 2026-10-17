// detect.go — the detection itself: the Detection dict (named
// accessors so it crosses the GUI thread boundary as a plain object),
// the slot grammar (dot = one nesting level), detect_backup_repo (the
// evidence ladder), radar_all_vaults / apply_detections (only EMPTY
// slots are ever filled), report_lines, and radar_pending (the
// quiet-startup gate).
package reporadar

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// Detection mirrors the Detection(dict) class: one vault's radar
// outcome — a dict with named accessors. NewDetection takes
// alternating key/value pairs, mirroring Detection(**kwargs) exactly
// (only the passed keys exist; absent keys read as the accessor's
// default).
type Detection struct {
	m *storage.OrderedMap
}

// NewDetection mirrors Detection(kind=…, vault_path=…, …).
func NewDetection(pairs ...any) *Detection {
	m := storage.NewOrderedMap()
	for i := 0; i+1 < len(pairs); i += 2 {
		if k, ok := pairs[i].(string); ok {
			m.Set(k, pairs[i+1])
		}
	}
	return &Detection{m: m}
}

// Map answers the underlying dict (the worker payload shape —
// Detection IS a dict).
func (d *Detection) Map() *storage.OrderedMap { return d.m }

// Kind mirrors the kind property.
func (d *Detection) Kind() string { return detStr(d, "kind") }

// VaultPath mirrors the vault_path property.
func (d *Detection) VaultPath() string { return detStr(d, "vault_path") }

// RepoName mirrors the repo_name property.
func (d *Detection) RepoName() string { return detStr(d, "repo_name") }

// Confidence mirrors the confidence property.
func (d *Detection) Confidence() string { return detStr(d, "confidence") }

// Method mirrors the method property.
func (d *Detection) Method() string { return detStr(d, "method") }

// Matched mirrors the matched property (int(get("matched") or 0)).
func (d *Detection) Matched() int { return detInt(d, "matched") }

// Sampled mirrors the sampled property.
func (d *Detection) Sampled() int { return detInt(d, "sampled") }

// Detail mirrors the detail property.
func (d *Detection) Detail() string { return detStr(d, "detail") }

// detStr mirrors str(self.get(key) or "").
func detStr(d *Detection, key string) string {
	v, _ := d.m.Get(key)
	if v == nil || !pyTruthy(v) {
		return ""
	}
	return helpers.PyStr(v)
}

// detInt mirrors int(self.get(key) or 0) over the dict shapes (int,
// float64 from JSON, …).
func detInt(d *Detection, key string) int {
	v, _ := d.m.Get(key)
	switch t := v.(type) {
	case nil:
		return 0
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		return 0
	default:
		if !pyTruthy(v) {
			return 0
		}
		return 0
	}
}

// pyTruthy mirrors Python's truthiness over the report shapes.
func pyTruthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return t != ""
	case bool:
		return t
	case int:
		return t != 0
	case int64:
		return t != 0
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	default:
		return true
	}
}

// asDetDict normalizes a results value to its dict form (*Detection,
// *storage.OrderedMap, or a plain map) — nil when the shape is not a
// dict (the reference's isinstance(det, dict) door).
func asDetDict(v any) *storage.OrderedMap {
	switch t := v.(type) {
	case *Detection:
		return t.m
	case *storage.OrderedMap:
		return t
	case map[string]any:
		m := storage.NewOrderedMap()
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			m.Set(k, t[k])
		}
		return m
	default:
		return nil
	}
}

// detGet reads a key off any dict-shaped detection value.
func detGet(v any, key string) any {
	if m := asDetDict(v); m != nil {
		got, _ := m.Get(key)
		return got
	}
	return nil
}

// slotValue mirrors _slot_value: the current repo-name value for a
// slot key like 'vaultseal.repo_name' (dot = one nesting level, the
// config grammar). config is the Python-dict-shaped config (nil /
// *storage.OrderedMap / map[string]any).
func slotValue(config any, slotKey string) string {
	if i := strings.Index(slotKey, "."); i >= 0 {
		head, tail := slotKey[:i], slotKey[i+1:]
		nested := cfgDictValue(config, head)
		if nested == nil {
			return ""
		}
		return strings.TrimSpace(cfgStrOf(cfgValue(nested, tail)))
	}
	return strings.TrimSpace(cfgStrOf(cfgValue(config, slotKey)))
}

// setSlotValue mirrors _set_slot_value: write a repo name into its
// slot; true when the value changed. Creates the nested dict when
// missing (never dropping neighbouring keys).
func setSlotValue(config any, slotKey, value string) bool {
	if i := strings.Index(slotKey, "."); i >= 0 {
		head, tail := slotKey[:i], slotKey[i+1:]
		switch t := config.(type) {
		case *storage.OrderedMap:
			nested, _ := t.Get(head)
			m, ok := nested.(*storage.OrderedMap)
			if !ok {
				m = storage.NewOrderedMap()
				t.Set(head, m)
			}
			if strings.TrimSpace(cfgStrOf(cfgValue(m, tail))) != value {
				m.Set(tail, value)
				return true
			}
			return false
		case map[string]any:
			nested, _ := t[head].(map[string]any)
			if nested == nil {
				nested = map[string]any{}
				t[head] = nested
			}
			if strings.TrimSpace(cfgStrOf(cfgValue(nested, tail))) != value {
				nested[tail] = value
				return true
			}
			return false
		default:
			return false
		}
	}
	switch t := config.(type) {
	case *storage.OrderedMap:
		if strings.TrimSpace(cfgStrOf(cfgValue(t, slotKey))) != value {
			t.Set(slotKey, value)
			return true
		}
		return false
	case map[string]any:
		if strings.TrimSpace(cfgStrOf(cfgValue(t, slotKey))) != value {
			t[slotKey] = value
			return true
		}
		return false
	default:
		return false
	}
}

// cfgGet / cfgValue / cfgDictValue / cfgStrOf — the Python-dict config
// readers (the vaultseal bridges' twins).
func cfgGet(config any, key string) (any, bool) {
	switch t := config.(type) {
	case nil:
		return nil, false
	case *storage.OrderedMap:
		return t.Get(key)
	case map[string]any:
		v, ok := t[key]
		return v, ok
	default:
		return nil, false
	}
}

func cfgValue(config any, key string) any {
	v, _ := cfgGet(config, key)
	return v
}

func cfgDictValue(config any, key string) any {
	switch t := cfgValue(config, key).(type) {
	case *storage.OrderedMap:
		return t
	case map[string]any:
		return t
	default:
		return nil
	}
}

func cfgStrOf(v any) string {
	if v == nil || !pyTruthy(v) {
		return ""
	}
	return helpers.PyStr(v)
}

// DetectBackupRepo mirrors detect_backup_repo: detect the backup repo
// for ONE vault. Never errors; repo_name is "" when nothing was found
// with confidence.
func DetectBackupRepo(kind, vaultPath string, api API, owner string) *Detection {
	path := strings.TrimSpace(vaultPath)
	label := vaultLabels[kind]
	if label == "" {
		label = kind
	}

	if path == "" {
		return NewDetection("kind", kind, "vault_path", "", "repo_name", "",
			"detail", "no vault path configured — nothing to detect for the "+
				helpers.PyLower(label))
	}

	if owner == "" {
		return NewDetection("kind", kind, "vault_path", path, "repo_name", "",
			"detail", "could not resolve the GitHub account for the token — "+
				"nothing was detected")
	}

	// Evidence 1 — the vault's own remote (the seal's signature).
	remote := RemoteOriginRepo(path)
	if remote != "" {
		payload := api.RepoExists(remote)
		if payload != nil && !pyTruthy(payload["_error"]) {
			name := remote
			if v, ok := payload["name"].(string); ok && v != "" {
				name = v
			}
			return NewDetection("kind", kind, "vault_path", path,
				"repo_name", name, "confidence", "certain", "method", "remote",
				"detail", "the vault's own git remote points at "+remote)
		}
		if payload != nil && pyTruthy(payload["_error"]) {
			return NewDetection("kind", kind, "vault_path", path, "repo_name", "",
				"detail", "the vault's remote points at "+remote+" but it could not "+
					"be verified ("+helpers.PyStr(payload["_error"])+")")
		}
	}

	// Evidence 2 — the vault folder's name IS a repo name (the cloned
	// vault case; also the not-yet-created Websites vault, whose
	// promised path names the repo).
	vaultName := vaultBaseName(path)
	if vaultName != "" {
		payload := api.RepoExists(owner + "/" + vaultName)
		if payload != nil && !pyTruthy(payload["_error"]) && pyTruthy(payload["name"]) {
			name, _ := payload["name"].(string)
			return NewDetection("kind", kind, "vault_path", path,
				"repo_name", name, "confidence", "certain", "method", "name",
				"detail", "a repo named "+vaultName+" exists on the account — "+
					"the vault folder's own name")
		}
	}

	// Evidence 3 — the blob fingerprint (only when the folder exists
	// on disk with files to hash).
	root := pyExpandUser(path)
	if isDir(root) {
		repos := api.ListRepos()
		if len(repos) > 0 {
			fp := fingerprint(path, repos, api, vaultName)
			if fp.repo != "" {
				return NewDetection("kind", kind, "vault_path", path,
					"repo_name", fp.repo, "confidence", "strong", "method", "fingerprint",
					"matched", fp.matched, "sampled", fp.sampled,
					"detail", fmt.Sprintf("%d/%d sampled files already live in %s — "+
						"the vault's content fingerprints to its mirror",
						fp.matched, fp.sampled, fp.repo))
			}
			best := "?"
			if len(fp.candidate) > 0 {
				best = fp.candidate[0]
			}
			if fp.why == "ambiguous" {
				return NewDetection("kind", kind, "vault_path", path, "repo_name", "",
					"matched", fp.matched, "sampled", fp.sampled,
					"detail", fmt.Sprintf("ambiguous — the best candidate (%s) matches "+
						"%d/%d sampled files, too close to another repo to call it; "+
						"type the repo name once in Settings and it is remembered",
						best, fp.matched, fp.sampled))
			}
			if fp.why == "weak" {
				return NewDetection("kind", kind, "vault_path", path, "repo_name", "",
					"matched", fp.matched, "sampled", fp.sampled,
					"detail", fmt.Sprintf("the best candidate (%s) matches only "+
						"%d/%d sampled files — not enough evidence to call it; "+
						"type the repo name once in Settings and it is remembered",
						best, fp.matched, fp.sampled))
			}
			if fp.sampled > 0 {
				return NewDetection("kind", kind, "vault_path", path, "repo_name", "",
					"sampled", fp.sampled,
					"detail", fmt.Sprintf("no repo on the account shares the vault's "+
						"content fingerprint (%d files sampled) — no backup repo "+
						"exists yet", fp.sampled))
			}
		}
	}

	return NewDetection("kind", kind, "vault_path", path, "repo_name", "",
		"detail", "no backup repo found for the "+helpers.PyLower(label)+
			" — nothing was created (the v0.83.0 law stands); create the repo on "+
			"GitHub once and the next radar pass will find it")
}

// vaultBaseName mirrors os.path.basename(path.rstrip("/\\")) or "".
func vaultBaseName(path string) string {
	rstripped := strings.TrimRight(path, "/\\")
	if rstripped == "" {
		return ""
	}
	base := filepath.Base(rstripped)
	if base == "." || base == "/" {
		return ""
	}
	return base
}

// RadarAllVaults mirrors radar_all_vaults: run the radar for every
// vault whose path is set and whose repo slot is EMPTY (the owner's
// typed word always wins — detection never overwrites). Returns
// {kind: Detection}; vaults with nothing to do carry a skipped
// Detection with the reason. Never errors — a panicking API surface
// (the reference's BoomAPI raise doors) is recovered per slot and
// spoken in the detail (the T-47 raise-door law).
func RadarAllVaults(config any, api API) map[string]*Detection {
	token := strings.TrimSpace(cfgStrOf(cfgValue(config, "github_token")))
	out := map[string]*Detection{}

	if token == "" {
		for _, slot := range VaultSlots {
			path := strings.TrimSpace(cfgStrOf(cfgValue(config, slot.PathKey)))
			if path != "" {
				out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", path,
					"repo_name", "",
					"detail", "no GitHub token — the radar sleeps (set it in "+
						"Settings → Credentials)")
			} else {
				out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", "",
					"repo_name", "", "detail", "no vault path configured")
			}
		}
		return out
	}

	radar := api
	if radar == nil {
		radar = NewRadarAPI(token, 0, "")
	}
	var meError string
	owner := ""
	func() {
		defer func() {
			if r := recover(); r != nil {
				meError = fmt.Sprintf(" (the account probe failed: %v)", r)
			}
		}()
		owner = radar.Me()
	}()
	if owner == "" {
		for _, slot := range VaultSlots {
			path := strings.TrimSpace(cfgStrOf(cfgValue(config, slot.PathKey)))
			if path != "" {
				out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", path,
					"repo_name", "",
					"detail", "the GitHub token was rejected (invalid or expired) "+
						"— nothing was detected"+meError)
			} else {
				out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", "",
					"repo_name", "", "detail", "no vault path configured")
			}
		}
		return out
	}

	for _, slot := range VaultSlots {
		path := strings.TrimSpace(cfgStrOf(cfgValue(config, slot.PathKey)))
		if path == "" {
			out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", "",
				"repo_name", "",
				"detail", "no vault path configured — set it once and the radar "+
					"does the rest")
			continue
		}
		current := slotValue(config, slot.SlotKey)
		if current != "" {
			out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", path,
				"repo_name", current, "confidence", "owner", "method", "config",
				"detail", "already set by the owner: "+current+" — never overwritten")
			continue
		}
		func() {
			defer func() {
				if r := recover(); r != nil {
					out[slot.Kind] = NewDetection("kind", slot.Kind, "vault_path", path,
						"repo_name", "", "detail", fmt.Sprintf("detection failed: %v", r))
				}
			}()
			out[slot.Kind] = DetectBackupRepo(slot.Kind, path, radar, owner)
		}()
	}
	return out
}

// ApplyDetections mirrors apply_detections: write the detected repo
// names into their config slots. Only detections WITH a repo_name
// land; empty slots keep their emptiness. Accepts Detection objects
// OR plain dicts (the worker payload shape — Detection IS a dict).
// Returns how many slots were set. config is mutated in place.
func ApplyDetections(config any, results map[string]any) int {
	applied := 0
	for _, slot := range VaultSlots {
		det, ok := results[slot.Kind]
		if !ok {
			continue
		}
		m := asDetDict(det)
		if m == nil {
			continue
		}
		repo := strings.TrimSpace(cfgStrOf(cfgValue(m, "repo_name")))
		if repo == "" {
			continue
		}
		if v, _ := m.Get("confidence"); v != nil && helpers.PyStr(v) == "owner" {
			continue
		}
		if setSlotValue(config, slot.SlotKey, repo) {
			applied++
		}
	}
	return applied
}

// ReportLines mirrors report_lines: one human line per vault — the
// radar's honest report. Accepts Detection objects or plain dicts.
func ReportLines(results map[string]any) []string {
	var lines []string
	for _, slot := range VaultSlots {
		det, ok := results[slot.Kind]
		if !ok {
			continue
		}
		m := asDetDict(det)
		if m == nil {
			continue
		}
		label := vaultLabels[slot.Kind]
		if label == "" {
			label = slot.Kind
		}
		repo := cfgStrOf(cfgValue(m, "repo_name"))
		confidence := cfgStrOf(cfgValue(m, "confidence"))
		if repo != "" && confidence == "owner" {
			lines = append(lines, "🔗 "+label+": "+repo+" (already set — kept)")
		} else if repo != "" {
			how := "detection"
			switch cfgStrOf(cfgValue(m, "method")) {
			case "remote":
				how = "its own git remote"
			case "name":
				how = "the vault folder's name"
			case "fingerprint":
				how = "a content fingerprint"
			}
			lines = append(lines, "🔗 "+label+": "+repo+" — detected via "+how)
		} else if cfgStrOf(cfgValue(m, "vault_path")) != "" {
			lines = append(lines, "⚪ "+label+": no backup repo found — "+
				cfgStrOf(cfgValue(m, "detail")))
		} else {
			lines = append(lines, "⚪ "+label+": not configured (no vault path)")
		}
	}
	return lines
}

// RadarPending mirrors radar_pending: true when the radar has work to
// do — a token and at least one EMPTY repo slot among the SET paths.
// The quiet-startup gate: a fully-wired config never wakes the radar.
func RadarPending(config any) bool {
	if strings.TrimSpace(cfgStrOf(cfgValue(config, "github_token"))) == "" {
		return false
	}
	for _, slot := range VaultSlots {
		if strings.TrimSpace(cfgStrOf(cfgValue(config, slot.PathKey))) != "" {
			if slotValue(config, slot.SlotKey) == "" {
				return true
			}
		}
	}
	return false // every set path has its repo slot filled — nothing pending
}
