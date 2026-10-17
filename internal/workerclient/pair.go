package workerclient

// pair.go — the pairing leg (no HMAC: the code from Telegram /pair)
// with its error mapping (403 → the redeploy advice, 401 → the
// run-/pair-again advice), and unpair.

import (
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/storage"
)

// Pair pairs with the Worker using a code from the /pair command.
// Returns (success, message). On success the three pairing keys are
// written into the LIVE config object and the master switch turns on.
func (s *Sync) Pair(code string) (bool, string) {
	if s.WorkerURL() == "" {
		return false, "No Worker URL configured. Enter it in the " +
			"Cloudflare tab first."
	}
	u := s.WorkerURL() + "/api/pair"
	body := `{"code": ` + jsonStr(code) + `}`
	req, err := http.NewRequest("POST", u, strings.NewReader(body))
	if err != nil {
		return false, "Pairing error: " + errClass(err) + ": " + err.Error()
	}
	req.Header.Set("Content-Type", "application/json")
	// v0.60.2 — the 1010 law: the pairing leg introduces itself too.
	req.Header.Set("User-Agent", HTTPUserAgent)
	res, err := directClient(30.0).Do(req)
	if err != nil {
		return false, "Cannot reach Worker at " + s.WorkerURL() + ": " +
			urlErrorReason(err)
	}
	defer res.Body.Close()
	raw, rerr := io.ReadAll(res.Body)
	if rerr != nil {
		return false, "Pairing error: " + errClass(rerr) + ": " + rerr.Error()
	}
	if res.StatusCode >= 200 && res.StatusCode < 300 {
		data, perr := storage.DecodeJSONValue(raw)
		if perr != nil {
			return false, "Pairing error: JSONDecodeError: " + perr.Error()
		}
		if pyTruthy(dictValue(data, "success")) {
			cfgSet(s.config, "cloudflare_install_id",
				pyStr(dictValue(data, "install_id")))
			cfgSet(s.config, "cloudflare_shared_secret",
				pyStr(dictValue(data, "shared_secret")))
			cfgSet(s.config, "cloudflare_enabled", true)
			return true, "Paired successfully!"
		}
		errMsg := pyStr(pyOr(dictValue(data, "error"), ""))
		if errMsg == "" {
			errMsg = "Pairing failed"
		}
		return false, errMsg
	}
	// The HTTPError branches: read the actual response body for
	// debugging, then map the code to its advice.
	bodyText := string(raw)
	errMsg := ""
	if data, perr := storage.DecodeJSONValue(raw); perr == nil && isDictLike(data) {
		// JSON body: data.get('error', f'HTTP {code}')
		errMsg = pyStr(pyOr(dictValue(data, "error"), ""))
		if errMsg == "" {
			errMsg = "HTTP " + itoa(res.StatusCode)
		}
	} else {
		// non-JSON (or non-dict) body: the raw text, capped at 200
		errMsg = bodyText
		if len([]rune(errMsg)) > 200 {
			errMsg = string([]rune(errMsg)[:200])
		}
		if errMsg == "" {
			errMsg = "HTTP " + itoa(res.StatusCode)
		}
	}
	switch res.StatusCode {
	case 403:
		return false, "HTTP 403 from Worker.\n\nResponse: " + errMsg +
			"\n\nWorker URL: " + s.WorkerURL() +
			"\n\nThis usually means:\n" +
			"1. Worker needs redeploy: run 'npx wrangler deploy' in " +
			"the cloudflare folder\n" +
			"2. Wrong Worker URL (check spelling)\n" +
			"3. Cloudflare blocking the request"
	case 401:
		return false, "Invalid or expired pairing code (401).\n\n" +
			"Response: " + errMsg + "\n\nRun /pair again in Telegram " +
			"and enter the code within 10 minutes."
	default:
		return false, "HTTP " + itoa(res.StatusCode) + ": " + errMsg
	}
}

// Unpair removes the pairing (for reset/debug).
func (s *Sync) Unpair() {
	cfgSet(s.config, "cloudflare_install_id", "")
	cfgSet(s.config, "cloudflare_shared_secret", "")
	cfgSet(s.config, "cloudflare_enabled", false)
}

// jsonStr renders a Python json.dumps string literal (ASCII-escaped,
// the wire shape).
func jsonStr(s string) string {
	return storage.PyDumpsCompact(s)
}

func itoa(n int) string {
	return strconv.Itoa(n)
}
