package connectioncheck

// worker.go — the Cloudflare bot Worker version check (v0.25.0
// staleness check), PROMOTED per ADR-0004 Am.3/R-0 from an auxiliary
// version check to the battery's queue-reachability leg: the queue
// rides the Worker HTTP API only, so GET {worker}/health covers both
// the version gate and queue reachability.

import (
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
)

// VersionCmp compares two dotted version strings numerically
// (0.22.0 vs 0.25.0): -1/0/1. Non-numeric segments compare equal to
// 0; a leading "v" is stripped; the tuple pads/truncates to 3 parts.
func VersionCmp(a, b string) int {
	pa, pb := versionParts(a), versionParts(b)
	switch {
	case pa[0] != pb[0]:
		if pa[0] < pb[0] {
			return -1
		}
		return 1
	case pa[1] != pb[1]:
		if pa[1] < pb[1] {
			return -1
		}
		return 1
	case pa[2] != pb[2]:
		if pa[2] < pb[2] {
			return -1
		}
		return 1
	}
	return 0
}

// versionParts is _version_cmp's parts(): strip leading "v" runes,
// split on ".", keep each chunk's ASCII digits (int or 0), pad to 3.
func versionParts(v string) [3]int {
	s := strings.TrimLeft(strings.TrimSpace(v), "v")
	var out [3]int
	if s == "" {
		return out
	}
	chunks := strings.Split(s, ".")
	for i, chunk := range chunks {
		if i >= 3 {
			break
		}
		digits := keepASCIIDigits(chunk)
		if digits != "" {
			if n, err := strconv.Atoi(digits); err == nil {
				out[i] = n
			}
		}
	}
	return out
}

func keepASCIIDigits(s string) string {
	var b strings.Builder
	for _, r := range s {
		if r >= '0' && r <= '9' {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// CheckWorkerVersion compares the deployed Worker's version (GET
// {url}/health → {"version": ...}) with EXPECTED_WORKER_VERSION so a
// stale bot is detectable from Test Connection. Returns a result
// pointer — nil when no worker URL is configured (the sync is
// optional: nothing to check, nothing to say). Never raises.
// httpGet ((url, timeout_s) → dict) is injectable for tests; nil uses
// the production proxy-bypassed fetch.
func CheckWorkerVersion(config any, httpGet WorkerHTTPGet) *Result {
	u := pyStr(pyOr(cfgValue(config, "cloudflare_worker_url"), ""))
	if !pyTruthy(u) {
		return nil
	}
	fetch := httpGet
	if fetch == nil {
		fetch = workerDirectGet
	}
	data, err := fetch(strings.TrimRight(u, "/")+"/health", 8.0)
	if err != nil {
		enabled := pyTruthy(cfgValue(config, "cloudflare_enabled"))
		level := LevelWarn
		if enabled {
			level = LevelError
		}
		return ptrResult(result("Bot Worker", level,
			"reachable check failed ("+errClass(err)+") — "+
				strings.TrimRight(u, "/")+"/health"))
	}
	deployed := pyStr(pyOr(dictValue(data, "version"), ""))
	if deployed == "" {
		return ptrResult(result("Bot Worker", LevelWarn,
			"no version reported by /health — redeploy the "+
				"Worker (deploy-latest.sh / deploy-latest.ps1 in "+
				"app/cloudflare-bot)"))
	}
	cmp := VersionCmp(deployed, constants.ExpectedWorkerVersion)
	switch {
	case cmp == 0:
		return ptrResult(result("Bot Worker", LevelOK,
			"v"+deployed+" — matches this app (expected v"+
				constants.ExpectedWorkerVersion+")"))
	case cmp < 0:
		return ptrResult(result("Bot Worker", LevelWarn,
			"v"+deployed+" deployed, this app expects v"+
				constants.ExpectedWorkerVersion+
				" — an older bot may miss newer link handling; "+
				"update it: cd app/cloudflare-bot && bash "+
				"deploy-latest.sh (Windows: .\\deploy-latest.ps1)"))
	default:
		return ptrResult(result("Bot Worker", LevelInfo,
			"v"+deployed+" deployed is NEWER than this app expects (v"+
				constants.ExpectedWorkerVersion+
				") — update the desktop app when convenient; the "+
				"bot stays backward compatible"))
	}
}

func ptrResult(r Result) *Result {
	return &r
}
