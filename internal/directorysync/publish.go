// publish.go — publish_index + fetch_published_index + run_sync +
// publish_after_batch + _effective_config (directory_sync.py lines
// 1015-1174): the Worker leg (1010-law UA, Bearer token,
// all-or-nothing) and THE VISUAL SYNC DISPLAY's read-back door (the
// four truths, v0.86.0).
package directorysync

import (
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// DefaultWorkerURL mirrors DEFAULT_WORKER_URL ("").
const DefaultWorkerURL = ""

// DefaultSyncToken mirrors DEFAULT_SYNC_TOKEN ("").
const DefaultSyncToken = ""

// DefaultVault mirrors DEFAULT_VAULT ("").
const DefaultVault = ""

// DefaultRepo mirrors DEFAULT_REPO ("").
const DefaultRepo = ""

// DefaultPollInterval mirrors DEFAULT_POLL_INTERVAL (seconds).
const DefaultPollInterval = 300

// publishTimeout mirrors publish_index's default timeout (seconds).
const publishTimeout = 30

// readbackTimeout mirrors fetch_published_index's default timeout.
const readbackTimeout = 15

// effectiveConfig mirrors _effective_config: dict(config or {}) with
// the three resolved keys — directory_worker_url (env
// DIRECTORY_WORKER_URL → config directory_worker_url → config
// cloudflare_worker_url, all stripped), directory_sync_token (env
// DIRECTORY_SYNC_TOKEN → config), website_vault_path (env
// DIRECTORY_VAULT → config). An env var set-but-empty falls through
// (Python's `or`).
func effectiveConfig(config any) *storage.OrderedMap {
	out := storage.NewOrderedMap()
	switch t := config.(type) {
	case *storage.OrderedMap:
		for _, k := range t.Keys() {
			v, _ := t.Get(k)
			out.Set(k, v)
		}
	case map[string]any:
		for k, v := range t {
			out.Set(k, v)
		}
	}
	if _, ok := out.Get("cloudflare_worker_url"); !ok {
		out.Set("cloudflare_worker_url", DefaultWorkerURL) // setdefault
	}
	cf, _ := out.Get("cloudflare_worker_url")

	dwu := os.Getenv("DIRECTORY_WORKER_URL")
	if dwu == "" {
		dwu = strings.TrimSpace(pyStrOrEmpty(mustCfgGet(out, "directory_worker_url")))
		if dwu == "" {
			dwu = strings.TrimSpace(pyStrOrEmpty(cf))
		}
	}
	out.Set("directory_worker_url", dwu)

	token := os.Getenv("DIRECTORY_SYNC_TOKEN")
	if token == "" {
		token = strings.TrimSpace(pyStrOrEmpty(mustCfgGet(out, "directory_sync_token")))
	}
	out.Set("directory_sync_token", token)

	vault := os.Getenv("DIRECTORY_VAULT")
	if vault == "" {
		vault = strings.TrimSpace(pyStrOrEmpty(mustCfgGet(out, "website_vault_path")))
	}
	out.Set("website_vault_path", vault)
	return out
}

// PublishIndex mirrors publish_index: POST directory.json to the
// Worker. Returns the Worker's response on success, nil on failure
// (the caller logs; the intake NEVER crashes on a publish failure —
// SPEC §14: "the automation retries next run and the old site content
// stays live").
func PublishIndex(index any, workerURL, token string, timeout int,
	log func(string)) any {
	if log == nil {
		log = func(string) {}
	}
	if timeout <= 0 {
		timeout = publishTimeout
	}
	if workerURL == "" || token == "" {
		log("[directory] publish skipped: worker URL or token not configured")
		return nil
	}
	url := strings.TrimRight(workerURL, "/") + "/api/directory/sync"
	// body = json.dumps(index, ensure_ascii=False).encode('utf-8')
	body := storage.PyDumpsUTF8(index)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(body))
	if err != nil {
		log("[directory] publish failed: " + err.Error())
		return nil
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("User-Agent", HTTPUserAgent)
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log("[directory] publish failed: " + pyURLErrorText(err))
		return nil
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		// the HTTPError door: e.read()[:200]
		detail := helpers.CutRunes(helpers.DecodeUTF8Replace(raw), 200)
		log("[directory] publish failed: HTTP " + strconv.Itoa(resp.StatusCode) + " " + detail)
		return nil
	}
	if err != nil {
		log("[directory] publish failed: " + err.Error())
		return nil
	}
	data, err := storage.DecodeJSONValue(raw)
	if err != nil {
		// the ValueError door (json.loads)
		log("[directory] publish failed: " + err.Error())
		return nil
	}
	log("[directory] published: " + pyReprValue(dictDoorOf(data).Get("stats")))
	return data
}

// FetchPublishedIndex mirrors fetch_published_index (v0.86.0 — THE
// VISUAL SYNC DISPLAY's read-back door): what the public site
// ACTUALLY carries, straight from the Worker's own GET
// /api/directory/index. Read-only, honest, never raises. Returns a
// state dict — one of four truths, never a silent zero:
// {state, sites, generated_at, message} with sites nil unless ok.
func FetchPublishedIndex(config any, timeout int, log func(string)) *storage.OrderedMap {
	if log == nil {
		log = func(string) {}
	}
	if timeout <= 0 {
		timeout = readbackTimeout
	}
	out := storage.NewOrderedMap()
	out.Set("state", "error")
	out.Set("sites", nil)
	out.Set("generated_at", "")
	out.Set("message", "")

	cfg := effectiveConfig(config)
	url := strings.TrimSpace(pyStrOrEmpty(mustCfgGet(cfg, "directory_worker_url")))
	if url == "" {
		out.Set("state", "not_configured")
		out.Set("message", "no worker URL configured — Settings → "+
			"Credentials has the door")
		return out
	}
	req, err := http.NewRequest(http.MethodGet,
		strings.TrimRight(url, "/")+"/api/directory/index", nil)
	if err != nil {
		out.Set("message", helpers.CutRunes(err.Error(), 200))
		return out
	}
	req.Header.Set("User-Agent", HTTPUserAgent)
	req.Header.Set("Accept", "application/json")
	client := &http.Client{Timeout: time.Duration(timeout) * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		out.Set("message", helpers.CutRunes(pyURLErrorText(err), 200))
		return out
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if resp.StatusCode == 404 {
		// the Worker's own honest answer: the publish door exists but
		// nothing has ever been published through it.
		out.Set("state", "never_published")
		out.Set("message", "never published — the site is empty until "+
			"the first publish runs")
		return out
	}
	if resp.StatusCode >= 400 {
		out.Set("message", "HTTP "+strconv.Itoa(resp.StatusCode))
		return out
	}
	if err != nil {
		out.Set("message", helpers.CutRunes(err.Error(), 200))
		return out
	}
	data, err := storage.DecodeJSONValue(raw)
	if err != nil {
		// the ValueError door
		out.Set("message", helpers.CutRunes(err.Error(), 200))
		return out
	}
	d := dictDoorOf(data)
	var sites any
	if counts, ok := d.Get("counts").(*storage.OrderedMap); ok {
		sites, _ = counts.Get("sites")
	} else if counts, ok := d.Get("counts").(map[string]any); ok {
		sites = counts["sites"]
	}
	switch t := sites.(type) {
	case int64:
		// an int literal is the only shape Python's isinstance(int)
		// accepts (bools excluded there; floats rejected)
		out.Set("state", "ok")
		out.Set("sites", t)
		out.Set("generated_at", pyStrOrEmpty(d.Get("generated_at")))
	default:
		out.Set("message", "the published index carried no site count")
	}
	return out
}

// RunSync mirrors run_sync: the whole Phase-1 leg — ensure the
// _system trio → build the index → publish. Never raises. Returns
// True only on a confirmed publish.
func RunSync(config any, log func(string)) bool {
	if log == nil {
		log = func(string) {}
	}
	cfg := effectiveConfig(config)
	vault := pyStrOrEmpty(mustCfgGet(cfg, "website_vault_path"))
	if vault == "" || !isDir(vault) {
		log("[directory] sync skipped: website_vault_path not configured")
		return false
	}
	EnsureSystemFiles(vault, log)
	index := BuildIndex(vault, nil, log)
	result := PublishIndex(index,
		pyStrOrEmpty(mustCfgGet(cfg, "directory_worker_url")),
		pyStrOrEmpty(mustCfgGet(cfg, "directory_sync_token")), 0, log)
	return result != nil && pyTruthyDS(dictDoorOf(result).Get("success"))
}

// PublishAfterBatch mirrors publish_after_batch: the end-of-batch
// hook (SPEC §6 golden rule 5) — the FULL cycle (commands first, then
// mirror, then publish, then ack) because golden rule 5's order
// demands the command apply precede the publish even on the
// end-of-batch path. The drift seal's publish door wires HERE
// (vaultseal.DriftSealDeps.PublishAfterBatch).
func PublishAfterBatch(config any, log func(string)) bool {
	summary := RunCommandCycle(config, log)
	return pyTruthyDS(mustCfgGet(summary, "published"))
}

// pyReprValue renders a decoded JSON value the way Python's f-string
// renders it (None/True/False, ints, floats, quoted strings, dicts
// with single quotes, lists) — the "published: {stats}" log line.
func pyReprValue(v any) string {
	switch t := v.(type) {
	case nil:
		return "None"
	case bool:
		if t {
			return "True"
		}
		return "False"
	case string:
		return helpers.PyRepr(t)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	case float64:
		return helpers.PyFloatRepr(t)
	case *storage.OrderedMap:
		var b strings.Builder
		b.WriteByte('{')
		for i, k := range t.Keys() {
			if i != 0 {
				b.WriteString(", ")
			}
			val, _ := t.Get(k)
			b.WriteString(helpers.PyRepr(k))
			b.WriteString(": ")
			b.WriteString(pyReprValue(val))
		}
		b.WriteByte('}')
		return b.String()
	case map[string]any:
		keys := make([]string, 0, len(t))
		for k := range t {
			keys = append(keys, k)
		}
		sortStringsDS(keys)
		var b strings.Builder
		b.WriteByte('{')
		for i, k := range keys {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(helpers.PyRepr(k))
			b.WriteString(": ")
			b.WriteString(pyReprValue(t[k]))
		}
		b.WriteByte('}')
		return b.String()
	case []any:
		var b strings.Builder
		b.WriteByte('[')
		for i, val := range t {
			if i != 0 {
				b.WriteString(", ")
			}
			b.WriteString(pyReprValue(val))
		}
		b.WriteByte(']')
		return b.String()
	default:
		return helpers.PyStr(v)
	}
}

func sortStringsDS(s []string) {
	for i := 1; i < len(s); i++ {
		for j := i; j > 0 && s[j] < s[j-1]; j-- {
			s[j], s[j-1] = s[j-1], s[j]
		}
	}
}
