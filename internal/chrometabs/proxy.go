// proxy.go — the right route (chrome_tabs.py:1609-1755): the Chrome
// argv for one configured proxy, the per-site route memory beside
// the door's identity, and the static leg plan — direct first (the
// residential line the challenges like), the owner's proxy second
// (the escalation's landing pad).
package chrometabs

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/assadigit/second-brain-project/internal/constants"
	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
	"github.com/assadigit/second-brain-project/internal/webfetch"
)

// Config knob names (the config example's own keys).
const (
	// ConfigProfilePersist mirrors CONFIG_PROFILE_PERSIST.
	ConfigProfilePersist = "web_browser_profile_persist"
	// ConfigProfileDir mirrors CONFIG_PROFILE_DIR.
	ConfigProfileDir = "web_browser_profile_dir"
	// ConfigAttach mirrors CONFIG_ATTACH.
	ConfigAttach = "web_browser_attach"
	// ConfigAttachPort mirrors CONFIG_ATTACH_PORT.
	ConfigAttachPort = "web_browser_attach_port"
	// ConfigRealProfile mirrors CONFIG_REAL_PROFILE.
	ConfigRealProfile = "web_browser_real_profile"
	// ConfigProxyLegs mirrors CONFIG_PROXY_LEGS.
	ConfigProxyLegs = "web_browser_proxy"
	// ConfigGate mirrors CONFIG_GATE (the end-of-run modal + the door).
	ConfigGate = "web_browser_retry"
	// ProxyMemoryFilename mirrors PROXY_MEMORY_FILENAME.
	ProxyMemoryFilename = "chrome-proxy-memory.json"
)

// Route memory values.
const (
	// RouteDirect mirrors ROUTE_DIRECT.
	RouteDirect = "direct"
	// RouteProxy mirrors ROUTE_PROXY.
	RouteProxy = "proxy"
)

// directLegFlags mirrors _DIRECT_LEG_FLAGS: the direct leg's own
// flag — TRULY off, even where the OS has a system proxy set (the
// owner's exact words: "proxy caused problems").
var directLegFlags = []string{"--no-proxy-server"}

// ProxyConfig is the proxy dict the legs ride ({type, host, port}).
// The reference passes web_fetch's proxy_from_config dict; the port
// reuses webfetch.Proxy's shape.
type ProxyConfig = webfetch.Proxy

// ChromeProxyFlag mirrors chrome_proxy_flag: the Chrome argv for one
// configured proxy (['--proxy-server=socks5://127.0.0.1:10808']; an
// http proxy rides http://) — the 'proxy ON' leg. An absent proxy is
// an EMPTY list. Pure.
func ChromeProxyFlag(proxy *ProxyConfig) []string {
	if proxy == nil {
		return nil
	}
	ptype := strings.ToLower(strings.TrimSpace(proxy.Type))
	if ptype == "" {
		ptype = "socks5"
	}
	host := strings.TrimSpace(proxy.Host)
	if host == "" || proxy.Port == 0 {
		return nil
	}
	switch ptype {
	case "socks5", "socks4", "http":
		return []string{"--proxy-server=" + ptype + "://" + host +
			":" + strconv.Itoa(proxy.Port)}
	}
	return []string{"--proxy-server=http://" + host + ":" +
		strconv.Itoa(proxy.Port)}
}

// ProxyMemoryPath mirrors proxy_memory_path: <APP_DIR>/
// chrome-proxy-memory.json (beside the door's identity). The
// reference's import-failure fallback (next to the module) cannot
// happen in Go — constants.AppDir always resolves.
func ProxyMemoryPath() string {
	return filepath.Join(constants.AppDir(), ProxyMemoryFilename)
}

// LoadProxyMemory mirrors load_proxy_memory: the remembered routes,
// {domain: 'direct' | 'proxy'} — an OrderedMap (insertion order, the
// reference's dict semantics). A missing, unreadable or corrupt file
// reads as EMPTY. Never raises.
func LoadProxyMemory(path string) *storage.OrderedMap {
	out := storage.NewOrderedMap()
	if path == "" {
		return out
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return out
	}
	om, err := storage.DecodeJSON(data)
	if err != nil {
		return out
	}
	for _, k := range om.Keys() {
		v, _ := om.Get(k)
		s, ok := v.(string)
		if !ok {
			continue
		}
		if s == RouteDirect || s == RouteProxy {
			out.Set(helpers.PyStr(k), s)
		}
	}
	return out
}

// SaveProxyMemory mirrors save_proxy_memory: write the remembered
// routes back (atomic, json.dumps(indent=2, ensure_ascii=False) —
// PyDumps over the OrderedMap keeps the insertion order). Best-effort,
// never raises.
func SaveProxyMemory(path string, memory *storage.OrderedMap) bool {
	if path == "" {
		return false
	}
	defer func() {
		if r := recover(); r != nil {
			// best-effort by law
		}
	}()
	if memory == nil {
		memory = storage.NewOrderedMap()
	}
	return storage.AtomicWriteText(path, storage.PyDumps(memory)) == nil
}

// DomainOfURL mirrors domain_of_url: the route memory's key — the
// host, lowercased, 'www.' stripped. Pure.
func DomainOfURL(rawURL string) string {
	_, hostname, _ := pyURLParts(rawURL)
	host := helpers.PyLower(hostname)
	if strings.HasPrefix(host, "www.") {
		host = host[4:]
	}
	return host
}

// NormalizeProxyMode mirrors normalize_proxy_mode: 'auto' (the
// both-ways ladder) | 'off' (never the proxy) | 'on' (always the
// proxy). Anything unreadable reads as 'auto'.
func NormalizeProxyMode(mode string) string {
	m := strings.ToLower(strings.TrimSpace(mode))
	switch m {
	case "auto", "off", "on":
		return m
	}
	return "auto"
}

// Leg is one planned leg: the route name and its urls, in run order.
type Leg struct {
	Route string
	URLs  []string
}

// PlanProxyLegs mirrors plan_proxy_legs: the static leg plan (pure) —
// which urls ride which leg BEFORE anyone fails anything. Direct
// first, proxy second. The laws: mode 'off' → every url on the direct
// leg; mode 'on' → every url on the proxy leg (a missing proxy
// degrades honestly to the direct leg); 'auto' → unknown +
// remembered-direct urls on the direct leg, remembered-proxy urls on
// the proxy leg, and the proxy leg is ALWAYS planned (even empty) —
// it is the ESCALATION's landing pad. Never raises.
func PlanProxyLegs(urls []string, memory *storage.OrderedMap, proxy *ProxyConfig, mode string) []Leg {
	var work []string
	for _, u := range urls {
		if u != "" {
			work = append(work, u)
		}
	}
	mode = NormalizeProxyMode(mode)
	haveProxy := len(ChromeProxyFlag(proxy)) > 0
	if mode == "off" || (mode == "auto" && !haveProxy) || (mode == "on" && !haveProxy) {
		if len(work) == 0 {
			return nil
		}
		return []Leg{{Route: RouteDirect, URLs: work}}
	}
	if mode == "on" {
		if len(work) == 0 {
			return nil
		}
		return []Leg{{Route: RouteProxy, URLs: work}}
	}
	// auto, with a proxy: the remembered split — and the proxy leg is
	// ALWAYS planned (even empty): it is the ESCALATION's landing pad.
	var directLeg, proxyLeg []string
	for _, u := range work {
		route := ""
		if memory != nil {
			if v, ok := memory.Get(DomainOfURL(u)); ok {
				if s, ok := v.(string); ok {
					route = s
				}
			}
		}
		if route == RouteProxy {
			proxyLeg = append(proxyLeg, u)
		} else {
			directLeg = append(directLeg, u)
		}
	}
	var legs []Leg
	if len(directLeg) > 0 {
		legs = append(legs, Leg{Route: RouteDirect, URLs: directLeg})
	}
	legs = append(legs, Leg{Route: RouteProxy, URLs: proxyLeg})
	return legs
}

// RememberRoute mirrors remember_route: the memory write law (pure) —
// a page that actually LANDED on a leg writes that leg's name for its
// domain. A url with no host is ignored.
func RememberRoute(memory *storage.OrderedMap, rawURL, route string) *storage.OrderedMap {
	if memory == nil {
		return memory
	}
	defer func() { recover() }()
	if host := DomainOfURL(rawURL); host != "" &&
		(route == RouteDirect || route == RouteProxy) {
		memory.Set(host, route)
	}
	return memory
}

// ProxyLabelForLog mirrors proxy_label_for_log: "socks5
// 127.0.0.1:10808" — the leg's log label (never a secret).
func ProxyLabelForLog(proxy *ProxyConfig) string {
	if proxy == nil {
		return "none"
	}
	t := strings.ToUpper(proxy.Type)
	if t == "" {
		t = "SOCKS5"
	}
	return t + " " + proxy.Host + ":" + strconv.Itoa(proxy.Port)
}
