// embeddings.go — embeddings.py (Phase 6, step 2): local embeddings —
// the short embedding TEXT per note (recall + one-line + tags, NEVER
// full text), the Ollama / OpenAI-compatible endpoints, the SQLite
// store and the stale-only refresh. Vectors are stored as JSON (pure
// stdlib) and compared with a plain cosine.
package mirror

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/llmclient"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/store"
)

// Batching: Ollama and llama.cpp both accept an input list; 16 keeps
// single requests small and progress granular.
const EmbedBatch = 16

// OllamaEmbedDefault is the provider default embedding model (config
// embedding_model wins).
const OllamaEmbedDefault = "nomic-embed-text"

// EmbeddingError is any embedding failure — always carries a
// human-readable cause.
type EmbeddingError struct{ Msg string }

func (e *EmbeddingError) Error() string { return e.Msg }

// ---------------------------------------------------------------------------
// The embedding text (SPEC: recall field + one-line + tags, nothing else)
// ---------------------------------------------------------------------------

// EmbedTextFor is the exact text embedded for a note
// (recall_fields_for output): name, one-line description, the recall
// sentence and the tags — a few dozen words, NEVER the full note.
func EmbedTextFor(fields RecallFields) string {
	name := helpers.Strip(fields.Name)
	oneLine := helpers.Strip(fields.OneLine)
	recall := helpers.Strip(fields.Recall)
	tags := helpers.Strip(fields.Tags)
	var parts []string
	for _, p := range []string{name, oneLine, recall} {
		if p != "" {
			parts = append(parts, p)
		}
	}
	if tags != "" {
		parts = append(parts, "tags: "+tags)
	}
	return helpers.CutRunes(strings.Join(parts, "\n"), 1200)
}

// HashEmbedText is the staleness key: the embedding is refreshed only
// when the embedded text (not the whole note) changes. Python's
// .encode('utf-8', 'replace') maps unencodable code points to '?'; the
// Go twin validates UTF-8 the same way (every text here came through
// the replace-decoder, so the corner is defensive only).
func HashEmbedText(text string) string {
	b := []byte(text)
	if !utf8.ValidString(text) {
		b = []byte(strings.ToValidUTF8(text, "?"))
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// ---------------------------------------------------------------------------
// Vector math (pure float64 — no numpy in the dependency tree)
// ---------------------------------------------------------------------------

// Cosine is cosine similarity; 0.0 for empty or mismatched vectors
// (never raises — a broken row degrades to 'unrelated', it does not
// kill the run).
func Cosine(a, b []float64) float64 {
	if len(a) == 0 || len(b) == 0 || len(a) != len(b) {
		return 0.0
	}
	dot, normA, normB := 0.0, 0.0, 0.0
	for i := range a {
		dot += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}
	if normA <= 0.0 || normB <= 0.0 {
		return 0.0
	}
	return dot / (math.Sqrt(normA) * math.Sqrt(normB))
}

// ---------------------------------------------------------------------------
// Endpoint calls (both never-proxy for loopback — the v0.15.1 rule)
// ---------------------------------------------------------------------------

// postJSON mirrors _post_json: POST the payload, read the body, decode
// the JSON. Loopback targets never ride the system proxy; remote hosts
// honor it. A non-JSON body is an EmbeddingError carrying the repr of
// its first 120 characters.
func postJSON(url string, payload any, timeoutS float64) (any, error) {
	body := []byte(storage.PyDumpsCompact(payload))
	req, err := http.NewRequest("POST", url, strings.NewReader(string(body)))
	if err != nil {
		return nil, &EmbeddingError{Msg: fmt.Sprintf(
			"URLError: <urlopen error %s>", err)}
	}
	req.Header.Set("Content-Type", "application/json")
	client := embedHTTPClient(url, timeoutS)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		// urllib raises HTTPError — carried through untouched (the
		// reference does not wrap it here).
		return nil, fmt.Errorf("HTTPError: HTTP Error %d: %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	decoded, jerr := storage.DecodeJSONValue(raw)
	if jerr != nil {
		return nil, &EmbeddingError{Msg: fmt.Sprintf(
			"%s returned a non-JSON body: %s", url,
			helpers.PyRepr(helpers.CutRunes(string(raw), 120)))}
	}
	return decoded, nil
}

// embedHTTPClient is the per-call client: direct (no proxy) for
// loopback — a proxy can only break it; the environment proxy for
// everything else. TLS is verified (the reference's plain urlopen).
func embedHTTPClient(url string, timeoutS float64) *http.Client {
	tr := &http.Transport{}
	if !llmclient.IsLoopbackURL(url) {
		tr.Proxy = http.ProxyFromEnvironment
	}
	client := &http.Client{Transport: tr}
	if timeoutS > 0 {
		client.Timeout = time.Duration(timeoutS * float64(time.Second))
	}
	return client
}

// toFloats converts a decoded JSON list of numbers into []float64.
func toFloats(v any) ([]float64, bool) {
	list, ok := v.([]any)
	if !ok {
		return nil, false
	}
	out := make([]float64, 0, len(list))
	for _, item := range list {
		f, ok := asFloat(item)
		if !ok {
			return nil, false
		}
		out = append(out, f)
	}
	return out, true
}

// payloadMap exposes a decoded JSON object as a getter (either
// representation).
func payloadMap(payload any) func(key string) (any, bool) {
	return func(key string) (any, bool) {
		switch t := payload.(type) {
		case *storage.OrderedMap:
			return t.Get(key)
		case map[string]any:
			v, ok := t[key]
			return v, ok
		}
		return nil, false
	}
}

// OllamaEmbed embeds via a local Ollama. Modern servers: one batched
// /api/embed call; older ones: /api/embeddings per text (the fallback
// is detected, not configured — a 404/405 on /api/embed means 'old
// server', not 'dead server'). Returns one vector per text, input
// order.
func OllamaEmbed(host, model string, texts []string, timeoutS float64) ([][]float64, error) {
	host = strings.TrimRight(pyOr(host, "http://127.0.0.1:11434"), "/")
	payload := storage.NewOrderedMap()
	payload.Set("model", model)
	inputs := make([]any, 0, len(texts))
	for _, t := range texts {
		inputs = append(inputs, t)
	}
	payload.Set("input", inputs)
	resp, err := postJSON(host+"/api/embed", payload, timeoutS)
	if err == nil {
		get := payloadMap(resp)
		if vectorsV, ok := get("embeddings"); ok {
			if list, isList := vectorsV.([]any); isList && len(list) == len(texts) {
				vectors := make([][]float64, 0, len(list))
				good := true
				for _, v := range list {
					vec, ok := toFloats(v)
					if !ok {
						good = false
						break
					}
					vectors = append(vectors, vec)
				}
				if good {
					return vectors, nil
				}
			}
		}
	}
	// fall through to legacy — one prompt per call
	var vectors [][]float64
	for _, text := range texts {
		req := storage.NewOrderedMap()
		req.Set("model", model)
		req.Set("prompt", text)
		resp, err := postJSON(host+"/api/embeddings", req, timeoutS)
		if err != nil {
			return nil, err
		}
		get := payloadMap(resp)
		vecV, _ := get("embedding")
		vec, ok := toFloats(vecV)
		if !ok {
			return nil, &EmbeddingError{Msg: fmt.Sprintf(
				"Ollama %s returned no embedding for the model '%s' "+
					"(pull it first: ollama pull %s)", host, model, model)}
		}
		vectors = append(vectors, vec)
	}
	return vectors, nil
}

// OpenAIEmbed embeds via any OpenAI-compatible /embeddings endpoint —
// llama-server with --embeddings, LM Studio, vLLM, or a cloud API.
// Returns one vector per text, input order (the response's index field
// is honored, not assumed).
func OpenAIEmbed(apiURL, model string, texts []string, apiKey string, timeoutS float64) ([][]float64, error) {
	base := strings.TrimRight(apiURL, "/")
	if base == "" {
		return nil, &EmbeddingError{Msg: "no embeddings API URL configured"}
	}
	payload := storage.NewOrderedMap()
	payload.Set("model", model)
	inputs := make([]any, 0, len(texts))
	for _, t := range texts {
		inputs = append(inputs, t)
	}
	payload.Set("input", inputs)
	body := []byte(storage.PyDumpsCompact(payload))
	req, err := http.NewRequest("POST", base+"/embeddings", strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("URLError: <urlopen error %s>", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+apiKey)
	}
	client := embedHTTPClient(base, timeoutS)
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("HTTPError: HTTP Error %d: %s",
			resp.StatusCode, http.StatusText(resp.StatusCode))
	}
	decoded, jerr := storage.DecodeJSONValue(raw)
	if jerr != nil {
		return nil, &EmbeddingError{Msg: fmt.Sprintf(
			"%s/embeddings returned a non-JSON body", base)}
	}
	get := payloadMap(decoded)
	dataV, _ := get("data")
	data, isList := dataV.([]any)
	if !isList || len(data) != len(texts) {
		n := 0
		if isList {
			n = len(data)
		}
		return nil, &EmbeddingError{Msg: fmt.Sprintf(
			"%s/embeddings returned %d embedding(s) for %d text(s)",
			base, n, len(texts))}
	}
	byIndex := map[int][]float64{}
	for _, item := range data {
		obj, ok := item.(map[string]any)
		if !ok {
			if om, isOM := item.(*storage.OrderedMap); isOM {
				idxV, _ := om.Get("index")
				vecV, _ := om.Get("embedding")
				if vec, ok := toFloats(vecV); ok {
					byIndex[intOf(idxV)] = vec
				}
				continue
			}
			continue
		}
		if vecV, isVec := obj["embedding"].([]any); isVec {
			if vec, ok := toFloats(vecV); ok {
				byIndex[intOf(obj["index"])] = vec
			}
		}
	}
	out := make([][]float64, 0, len(texts))
	for i := 0; i < len(texts); i++ {
		vec, ok := byIndex[i]
		if !ok {
			return nil, &EmbeddingError{Msg: fmt.Sprintf(
				"%s/embeddings response missing index %d", base, i)}
		}
		out = append(out, vec)
	}
	return out, nil
}

// intOf is Python int() over a decoded JSON number-ish value.
func intOf(v any) int {
	switch t := v.(type) {
	case nil:
		return 0
	case float64:
		return int(t)
	case int:
		return t
	case int64:
		return int(t)
	case json.Number:
		f, err := t.Float64()
		if err != nil {
			return 0
		}
		return int(f)
	case string:
		n, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return 0
		}
		return int(n)
	}
	return 0
}

// EmbedFn is the embed-texts closure shape (one vector per text, input
// order).
type EmbedFn func(texts []string) ([][]float64, error)

// BuildEmbeddingCall mirrors build_embedding_call: config → (embed,
// label). Same provider routing as the chat paths: llamacpp → the
// detected server's /v1/embeddings; cloud → the configured endpoint;
// ollama (default) → /api/embed. The model comes from config
// embedding_model or the provider default.
func BuildEmbeddingCall(config any, modelOverride string) (EmbedFn, string, error) {
	get := func(key string) (any, bool) {
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
	// str(cfg.get(key, def) or def) — falsy (absent/None/''/0/false)
	// falls back to the default, exactly like Python's `or`.
	strOf := func(key string) string {
		v, ok := get(key)
		if !ok || v == nil {
			return ""
		}
		switch t := v.(type) {
		case string:
			return t
		case bool:
			if t {
				return "True"
			}
			return ""
		default:
			if f, isNum := asFloat(v); isNum && f == 0 {
				return ""
			}
			return helpers.PyStr(v)
		}
	}

	provider := strings.ToLower(pyOr(strOf("llm_provider"), "ollama"))
	model := helpers.Strip(pyOr(modelOverride, strOf("embedding_model")))
	// float(cfg.get('llm_timeout_s', 300) or 300) — the falsy door again;
	// the trailing `or 60.0` is unreachable in the reference (300 is
	// truthy) and kept only for the letter.
	timeout := 300.0
	if v, ok := get("llm_timeout_s"); ok && v != nil {
		if f, isNum := asFloat(v); isNum && f != 0 {
			timeout = f
		}
	}
	if timeout == 0 {
		timeout = 60.0
	}

	switch provider {
	case "llamacpp":
		base := strOf("llamacpp_api_url")
		if base == "" {
			base = llmclient.LLamaCppDefaultBase + "/v1"
		}
		apiURL := llmclient.NormalizeLlamaCppAPIURL(base)
		key := strOf("llamacpp_api_key")
		if model == "" {
			model = helpers.Strip(strOf("llamacpp_model"))
		}
		label := fmt.Sprintf("llama.cpp %s · %s", apiURL,
			pyOr(model, "(server model)"))
		return func(texts []string) ([][]float64, error) {
			return OpenAIEmbed(apiURL, model, texts, key, timeout)
		}, label, nil
	case "cloud", "custom":
		apiURL := strings.TrimRight(strOf("cloud_api_url"), "/")
		key := strOf("cloud_api_key")
		if apiURL == "" {
			return nil, "", &EmbeddingError{
				Msg: "cloud provider needs cloud_api_url for embeddings"}
		}
		if model == "" {
			model = helpers.Strip(strOf("cloud_model"))
		}
		if model == "" {
			return nil, "", &EmbeddingError{Msg: "cloud provider needs an " +
				"embedding model (config embedding_model, or --embed-model)"}
		}
		label := fmt.Sprintf("%s · %s", apiURL, model)
		return func(texts []string) ([][]float64, error) {
			return OpenAIEmbed(apiURL, model, texts, key, timeout)
		}, label, nil
	}

	// ollama (default)
	host := ""
	if oll, ok := get("ollama"); ok {
		if om, isOM := oll.(*storage.OrderedMap); isOM {
			if v, p := om.Get("base_url"); p {
				host = helpers.PyStr(v)
			}
		} else if m, isM := oll.(map[string]any); isM {
			host = helpers.PyStr(m["base_url"])
		}
	}
	if host == "" {
		host = strOf("ollama_url")
	}
	if host == "" {
		host = "http://127.0.0.1:11434"
	}
	if model == "" {
		if oll, ok := get("ollama"); ok {
			if om, isOM := oll.(*storage.OrderedMap); isOM {
				if v, p := om.Get("model"); p {
					model = helpers.PyStr(v)
				}
			} else if m, isM := oll.(map[string]any); isM {
				model = helpers.PyStr(m["model"])
			}
		}
	}
	if model == "" {
		model = OllamaEmbedDefault
	}
	label := fmt.Sprintf("Ollama %s · %s", host, model)
	return func(texts []string) ([][]float64, error) {
		return OllamaEmbed(host, model, texts, timeout)
	}, label, nil
}

// ---------------------------------------------------------------------------
// SQLite store (same cache.db, own table — the NoteStateDB pattern)
// ---------------------------------------------------------------------------

// EmbeddingStore holds vectors keyed by (vault, source_url, model),
// JSON-serialized — pure stdlib. Own connection + lock; CacheDB is not
// touched.
type EmbeddingStore struct {
	*store.DB
}

// NewEmbeddingStore opens (creating if absent) the store; the literal
// default "cache.db" is anchored to APP_DIR (the v0.09.4 rule).
func NewEmbeddingStore(dbPath string) (*EmbeddingStore, error) {
	if dbPath == "cache.db" {
		dbPath = filepath.Join(constants.AppDir(), "cache.db")
	}
	base, err := store.Open(dbPath)
	if err != nil {
		return nil, err
	}
	s := &EmbeddingStore{DB: base}
	s.Lock()
	defer s.Unlock()
	if _, err := s.SQL().Exec(`
                CREATE TABLE IF NOT EXISTS embeddings (
                        vault TEXT NOT NULL,
                        source_url TEXT NOT NULL,
                        model TEXT NOT NULL,
                        dim INTEGER NOT NULL,
                        vec TEXT NOT NULL,
                        text_hash TEXT NOT NULL,
                        updated_at TEXT NOT NULL,
                        PRIMARY KEY (vault, source_url, model)
                )`); err != nil {
		base.Close()
		return nil, err
	}
	return s, nil
}

// EmbeddingValue is one loaded row: (vector, text_hash).
type EmbeddingValue struct {
	Vec      []float64
	TextHash string
}

// Put upserts one vector (vec stored as Python json.dumps bytes).
func (s *EmbeddingStore) Put(vault, sourceURL, model string, vector []float64, textHash string) error {
	s.Lock()
	defer s.Unlock()
	_, err := s.SQL().Exec(`
                INSERT INTO embeddings (vault, source_url, model, dim, vec,
                        text_hash, updated_at) VALUES (?,?,?,?,?,?,?)
                ON CONFLICT(vault, source_url, model) DO UPDATE SET
                        dim=excluded.dim, vec=excluded.vec,
                        text_hash=excluded.text_hash,
                        updated_at=excluded.updated_at`,
		vault, sourceURL, model, len(vector),
		storage.PyDumpsCompact(vecAny(vector)), textHash,
		pyStrftimeSeconds())
	return err
}

// vecAny converts for the dumps writer.
func vecAny(vector []float64) any {
	out := make([]any, 0, len(vector))
	for _, v := range vector {
		out = append(out, v)
	}
	return out
}

// Load reads {source_url: (vector, text_hash)} for one vault+model; a
// broken row degrades to absent.
func (s *EmbeddingStore) Load(vault, model string) (map[string]EmbeddingValue, error) {
	s.RLock()
	defer s.RUnlock()
	rows, err := s.SQL().Query(
		"SELECT source_url, vec, text_hash FROM embeddings WHERE vault=? AND model=?",
		vault, model)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := map[string]EmbeddingValue{}
	for rows.Next() {
		var url, vecJSON, textHash string
		if err := rows.Scan(&url, &vecJSON, &textHash); err != nil {
			continue
		}
		v, err := storage.DecodeJSONValue([]byte(vecJSON))
		if err != nil {
			continue
		}
		vec, ok := toFloats(v)
		if !ok {
			continue
		}
		out[url] = EmbeddingValue{Vec: vec, TextHash: textHash}
	}
	return out, rows.Err()
}

// DropModel deletes one vault+model's rows; returns the row count.
func (s *EmbeddingStore) DropModel(vault, model string) (int, error) {
	s.Lock()
	defer s.Unlock()
	res, err := s.SQL().Exec(
		"DELETE FROM embeddings WHERE vault=? AND model=?", vault, model)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// Count rows (all vaults when vault is nil).
func (s *EmbeddingStore) Count(vault *string) (int, error) {
	s.RLock()
	defer s.RUnlock()
	var (
		rows *sql.Rows
		err  error
	)
	if vault == nil {
		rows, err = s.SQL().Query("SELECT COUNT(*) FROM embeddings")
	} else {
		rows, err = s.SQL().Query(
			"SELECT COUNT(*) FROM embeddings WHERE vault=?", *vault)
	}
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	if rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			return 0, err
		}
		return n, nil
	}
	return 0, rows.Err()
}

// ---------------------------------------------------------------------------
// Refresh (the tool layer drives this)
// ---------------------------------------------------------------------------

// RefreshEntry is one {source_url, fields} input.
type RefreshEntry struct {
	SourceURL string
	Fields    RecallFields
}

// RefreshStats is the {embedded, skipped_fresh, failed} result.
type RefreshStats struct {
	Embedded     int
	SkippedFresh int
	Failed       int
}

// RefreshEmbeddings embeds the stale notes only. Failures never abort
// the batch — the note simply keeps its old vector (or stays absent).
func RefreshEmbeddings(entries []RefreshEntry, embed EmbedFn,
	s *EmbeddingStore, vault, model string, force bool,
	progress func(string)) (RefreshStats, error) {

	say := func(msg string) {
		if progress != nil {
			progress(msg)
		}
	}
	stats := RefreshStats{}
	existing, err := s.Load(vault, model)
	if err != nil {
		return stats, err
	}
	type batchEntry struct {
		SourceURL string
		EmbedText string
		TextHash  string
	}
	flush := func(batch []batchEntry) error {
		if len(batch) == 0 {
			return nil
		}
		texts := make([]string, 0, len(batch))
		for _, e := range batch {
			texts = append(texts, e.EmbedText)
		}
		vectors, err := embed(texts)
		if err == nil && len(vectors) != len(batch) {
			err = &EmbeddingError{Msg: fmt.Sprintf(
				"endpoint returned %d vector(s) for %d text(s)",
				len(vectors), len(batch))}
		}
		if err != nil {
			stats.Failed += len(batch)
			say(fmt.Sprintf("  ! batch of %d failed: %s", len(batch), err))
			return nil
		}
		for i, e := range batch {
			if err := s.Put(vault, e.SourceURL, model, vectors[i], e.TextHash); err != nil {
				return err
			}
			stats.Embedded++
		}
		return nil
	}
	var batch []batchEntry
	for _, entry := range entries {
		text := EmbedTextFor(entry.Fields)
		textHash := HashEmbedText(text)
		old, hadOld := existing[entry.SourceURL]
		if !force && hadOld && old.TextHash == textHash {
			stats.SkippedFresh++
			continue
		}
		batch = append(batch, batchEntry{
			SourceURL: entry.SourceURL, EmbedText: text, TextHash: textHash})
		if len(batch) >= EmbedBatch {
			if err := flush(batch); err != nil {
				return stats, err
			}
			batch = batch[:0]
		}
	}
	if err := flush(batch); err != nil {
		return stats, err
	}
	return stats, nil
}

// asFloat converts a decoded JSON number (json.Number / float64 /
// int64 / int — the house decoder's shapes) to float64.
func asFloat(v any) (float64, bool) {
	switch t := v.(type) {
	case float64:
		return t, true
	case float32:
		return float64(t), true
	case int:
		return float64(t), true
	case int64:
		return float64(t), true
	case json.Number:
		f, err := t.Float64()
		return f, err == nil
	case string:
		f, err := strconv.ParseFloat(t, 64)
		return f, err == nil
	}
	return 0, false
}
