// detect.go — detect_ollama (llm_client.py v0.18.0): the Ollama twin
// of probe_llamacpp — ONE /api/tags probe answering "is the Ollama
// server up, and which models are installed?". Never raises; the
// app's own GitCurator-detect User-Agent reaches the wire.
package llmclient

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// DetectOllamaResult is the result dict shape.
type DetectOllamaResult struct {
	Found   bool     `json:"found"`
	BaseURL string   `json:"base_url"`
	Models  []string `json:"models"`
	Detail  string   `json:"detail"`
}

// DetectOllama probes base_url (every user spelling; empty → the
// Ollama default) with GET /api/tags. Raw HTTP, loopback targets
// bypass the system proxy, remote hosts honor it. The failure detail
// carries the Python exception-class prefix (URLError / HTTPError /
// JSONDecodeError / ValueError) — the runtime text after the colon is
// runtime-shaped and unpinned.
func DetectOllama(baseURL string, timeoutS float64) DetectOllamaResult {
	raw := strings.TrimSpace(baseURL)
	if raw == "" {
		raw = DefaultOllamaBase
	}
	if !strings.Contains(raw, "://") {
		raw = "http://" + raw
	}
	base := strings.TrimRight(raw, "/")
	res := DetectOllamaResult{Found: false, BaseURL: base, Models: []string{}}

	req, err := http.NewRequest("GET", base+"/api/tags", nil)
	if err != nil {
		res.Detail = fmt.Sprintf("URLError: <urlopen error %s>", err)
		return res
	}
	req.Header.Set("User-Agent", "GitCurator-detect")

	client := httpClient(false, IsLoopbackURL(base))
	if timeoutS > 0 {
		req = req.WithContext(timeoutCtx(req.Context(), timeoutS))
	}
	resp, err := client.Do(req)
	if err != nil {
		res.Detail = fmt.Sprintf("URLError: %s", pyURLOpenError(err))
		return res
	}
	defer resp.Body.Close()
	rawBody, _ := readAllBody(resp)
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		res.Detail = fmt.Sprintf("HTTPError: HTTP Error %d: %s",
			resp.StatusCode, reasonPhrase(resp))
		return res
	}

	body := rawBody
	if body == "" {
		body = "{}"
	}
	payload, jerr := storage.DecodeJSONValue([]byte(body))
	if jerr != nil {
		res.Detail = "JSONDecodeError: " + jerr.Error()
		return res
	}
	pm, ok := asOM(payload)
	if !ok {
		res.Detail = "ValueError: response is not a JSON object"
		return res
	}
	models := []string{}
	if modelsV, p := pm.Get("models"); p {
		if list, ok := modelsV.([]any); ok {
			models = extractTagNames(list)
		}
	}
	// De-duplicate, keep the server's order.
	ordered := []string{}
	seen := map[string]bool{}
	for _, n := range models {
		if !seen[n] {
			seen[n] = true
			ordered = append(ordered, n)
		}
	}
	// The server ANSWERED with a parsable dict — found is true even
	// with zero models installed (the Python law).
	res.Found = true
	res.Models = ordered
	if len(ordered) > 0 {
		res.Detail = fmt.Sprintf("up · %d model(s)", len(ordered))
	} else {
		res.Detail = "up but NO models installed"
	}
	return res
}
