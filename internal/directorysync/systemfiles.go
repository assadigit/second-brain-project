// systemfiles.go — the vault's own _system trio (SPEC §7.4 + §11.3):
// ensure_system_files creates the three files from the app's shipped
// defaults when missing, NEVER overwrites (the owner's edits are
// sacred); load_synonyms reads _system/search-synonyms.json with the
// shipped fallback. load_directory_config + load_banish_list already
// landed with the T-47 belt floor / the T-40 banish gate.
package directorysync

import (
	"os"
	"path/filepath"

	"github.com/assadigit/second-brain-project/assets"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// SynonymsFilename mirrors SYNONYMS_FILENAME.
const SynonymsFilename = "search-synonyms.json"

// ShippedSynonyms is the SHIPPED search-synonyms default
// (app/taxonomy/search-synonyms.json in the reference), copied
// verbatim — never hand-edited; sha256-pinned by the directorysync
// golden (system_files section).
//
// The reference's load_synonyms reads this file from disk at runtime
// (app/taxonomy/); the Go binary embeds the same bytes as the shipped
// default — the same documented convergence the shipped
// directory-config embed (T-47) records.
var ShippedSynonyms = assets.SearchSynonyms

// LoadSynonyms mirrors load_synonyms: _system/search-synonyms.json
// (owner-extendable) with the shipped fallback; the empty dict on
// every miss. Never errors.
func LoadSynonyms(vault string) any {
	if data, err := os.ReadFile(filepath.Join(SystemDir(vault), SynonymsFilename)); err == nil {
		if v, err := storage.DecodeJSONValue(data); err == nil {
			// `isinstance(data, dict) and not isinstance(data, list)`
			// — dict-shaped and not list-shaped (lists never decode
			// dict-shaped here, so the dict check is the whole law).
			if _, ok := dictOf(v); ok {
				return v
			}
		}
	}
	if v, err := storage.DecodeJSONValue(ShippedSynonyms); err == nil {
		if _, ok := dictOf(v); ok {
			return v
		}
	}
	return storage.NewOrderedMap()
}

// EnsureSystemFiles mirrors ensure_system_files: create the _system
// trio from the shipped defaults when missing. The vault copy always
// wins afterwards — this never overwrites an existing file. Returns
// the list of files it created (absolute paths). log may be nil.
func EnsureSystemFiles(vault string, log func(string)) []string {
	if log == nil {
		log = func(string) {}
	}
	var created []string
	if vault == "" {
		return created
	}
	if fi, err := os.Stat(vault); err != nil || !fi.IsDir() {
		return created
	}
	sdir := SystemDir(vault)
	_ = os.MkdirAll(sdir, 0o755)
	banishedPayload := jsonDumpsIndent2List() + "\n"
	pairs := []struct {
		filename string
		payload  []byte
	}{
		{BanishedFilename, []byte(banishedPayload)},
		{DirectoryConfigFilename, assets.WebsiteDirectoryConfig},
		{SynonymsFilename, ShippedSynonyms},
	}
	for _, pair := range pairs {
		target := filepath.Join(sdir, pair.filename)
		if _, err := os.Stat(target); err == nil {
			continue
		}
		if err := os.WriteFile(target, pair.payload, 0o644); err != nil {
			continue
		}
		created = append(created, target)
		log("[directory] created " + filepath.Join(SystemDirname, pair.filename))
	}
	return created
}

// jsonDumpsIndent2List mirrors json.dumps([], indent=2,
// ensure_ascii=False): "[]".
func jsonDumpsIndent2List() string { return "[]" }
