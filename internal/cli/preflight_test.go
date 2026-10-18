package cli

// preflight_test.go — T-62's doors' unit battery: the run-config card
// (the mask/truncation/provider laws), the three preflights over the
// seams, and the PreflightChecks composition (the abort door, the
// continue-anyway door, the cron law). The differential bytes are the
// golden replay's job (goldencli_test.go); these pin the behavioral
// laws.

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/assadigit/second-brain-project/internal/storage"
)

func om(pairs ...any) *storage.OrderedMap {
	m := storage.NewOrderedMap()
	for i := 0; i+1 < len(pairs); i += 2 {
		m.Set(pairs[i].(string), pairs[i+1])
	}
	return m
}

// TestRunConfigCardMask laws: the >10 mask, the short mask, the empty
// "(not set)", and the provider branches (the raw-comparison law — a
// capitalized provider spelling falls to the OpenAI-compatible row).
func TestRunConfigCardMask(t *testing.T) {
	cases := []struct {
		name string
		cfg  *storage.OrderedMap
		want []string
	}{
		{
			name: "long token masked",
			cfg:  om("github_token", "gh-token-1234567890abcdef"),
			want: []string{"GitHub token", "gh-t••••••••cdef"},
		},
		{
			name: "short token all-bulleted",
			cfg:  om("github_token", "abc"),
			want: []string{"GitHub token", "•••"},
		},
		{
			name: "no token",
			cfg:  om(),
			want: []string{"GitHub token", "(not set)"},
		},
		{
			name: "capitalized provider falls to cloud branch",
			cfg: om("llm_provider", "Ollama", "cloud_model", "gpt-x",
				"ollama", om("model", "qwen")),
			want: []string{"LLM", "OpenAI-compatible · gpt-x"},
		},
		{
			name: "ollama provider",
			cfg:  om("llm_provider", "ollama", "ollama", om("model", "qwen")),
			want: []string{"LLM", "Ollama · qwen"},
		},
		{
			name: "ollama missing model renders ?",
			cfg:  om("llm_provider", "ollama", "ollama", om()),
			want: []string{"LLM", "Ollama · ?"},
		},
		{
			name: "llamacpp defaults",
			cfg:  om("llm_provider", "llamacpp"),
			want: []string{"LLM", "llama.cpp · (auto-detect) @ http://127.0.0.1:8080/v1"},
		},
		{
			name: "bot + proxy on",
			cfg: om("bot_username", "MyCuratorBot", "proxy",
				om("enabled", true, "host", "127.0.0.1", "port", int64(1080))),
			want: []string{"Bot queue", "@MyCuratorBot", "Proxy", "127.0.0.1:1080 (on)"},
		},
		{
			name: "proxy off",
			cfg:  om("proxy", om("enabled", false)),
			want: []string{"Proxy", "disabled"},
		},
		{
			name: "vault unset",
			cfg:  om(),
			want: []string{"Vault", "(not set)"},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			c := testCLI(buf, "")
			c.RunConfigCard(tc.cfg)
			got := buf.String()
			for _, w := range tc.want {
				if !strings.Contains(got, w) {
					t.Fatalf("card missing %q:\n%s", w, got)
				}
			}
			// The frame: the title row + the closing corner.
			if !strings.Contains(got, " run configuration ") {
				t.Fatalf("card title missing:\n%s", got)
			}
			if !strings.Contains(got, "└") {
				t.Fatalf("card bottom missing:\n%s", got)
			}
		})
	}
}

// TestRunConfigCardTruncation: a value wider than the inner box is cut
// with the … rune (the COLUMNS law: 80-col terminal → avail 76 → val_w
// 57).
func TestRunConfigCardTruncation(t *testing.T) {
	t.Setenv("COLUMNS", "80")
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	long := strings.Repeat("v", 90)
	c.RunConfigCard(om("vault_path", long))
	got := buf.String()
	if !strings.Contains(got, "…") {
		t.Fatalf("expected the truncation rune:\n%s", got)
	}
	if strings.Contains(got, strings.Repeat("v", 58)) {
		t.Fatalf("value not truncated:\n%s", got)
	}
}

// TestPreflightGithubSeam: the four outcomes.
func TestPreflightGithubSeam(t *testing.T) {
	cases := []struct {
		name  string
		reply PreflightGHReply
		level string
		want  string
	}{
		{"ok", PreflightGHReply{Status: 200, Login: "ali"},
			"success", "authenticated as ali"},
		{"rejected", PreflightGHReply{Status: 401},
			"error", "token REJECTED (401) — make a new one at github.com/settings/tokens"},
		{"unreachable", PreflightGHReply{Status: 0, ErrClass: "GatewayTimeout"},
			"error", "GitHub unreachable (GatewayTimeout) — check internet"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			buf := &bytes.Buffer{}
			c := testCLI(buf, "")
			c.PreflightGithub = func(string) PreflightGHReply { return tc.reply }
			lvl, txt := c.preflightGithub(om("github_token", "t"))
			if lvl != tc.level || txt != tc.want {
				t.Fatalf("got (%s, %q), want (%s, %q)", lvl, txt, tc.level, tc.want)
			}
		})
	}
	// the no-token warning (the seam never fires)
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	lvl, txt := c.preflightGithub(om())
	if lvl != "warning" || txt != "no token — running unauthenticated (60 req/h limit)" {
		t.Fatalf("no-token: got (%s, %q)", lvl, txt)
	}
}

// TestPreflightProxySeam: disabled / reachable / unreachable.
func TestPreflightProxySeam(t *testing.T) {
	t.Run("disabled", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.NetDial = func(string, int, int) (int, string) {
			t.Fatal("dial must not fire when disabled")
			return 0, ""
		}
		lvl, txt := c.preflightProxy(om("proxy", om("enabled", false)))
		if lvl != "dim" || txt != "disabled — connecting directly" {
			t.Fatalf("got (%s, %q)", lvl, txt)
		}
	})
	t.Run("reachable", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.NetDial = func(host string, port, _ int) (int, string) {
			if host != "127.0.0.1" || port != 1080 {
				t.Fatalf("dial args: %s:%d", host, port)
			}
			return 42, ""
		}
		lvl, txt := c.preflightProxy(om("proxy",
			om("enabled", true, "host", "127.0.0.1", "port", int64(1080))))
		want := "127.0.0.1:1080 reachable (socks5, 42 ms)"
		if lvl != "success" || txt != want {
			t.Fatalf("got (%s, %q), want (%s, %q)", lvl, txt, "success", want)
		}
	})
	t.Run("unreachable", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.NetDial = func(string, int, int) (int, string) { return 0, "TimeoutError" }
		lvl, txt := c.preflightProxy(om("proxy",
			om("enabled", true, "host", "proxy.local", "port", "1080")))
		want := "proxy.local:1080 UNREACHABLE (TimeoutError) — start the proxy app or disable it in config.json"
		if lvl != "warning" || txt != want {
			t.Fatalf("got (%s, %q), want (%s, %q)", lvl, txt, "warning", want)
		}
	})
}

// TestPreflightLLMOllama: ready / not-pulled (menu pick + save) /
// declined / no models / unreachable-continues.
func TestPreflightLLMOllama(t *testing.T) {
	// fresh cfg per subtest: ApplyModelChoice mutates the in-memory
	// config (the v30 persistence law) — a shared cfg would leak the
	// previous subtest's switch into the next verdict.
	fresh := func() *storage.OrderedMap {
		return om("llm_provider", "ollama",
			"ollama", om("base_url", "http://127.0.0.1:11434", "model", "qwen:7b"))
	}

	t.Run("ready", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.OllamaModels = func(string, float64) ([]string, error) {
			return []string{"qwen:7b", "llama3"}, nil
		}
		if !c.preflightLLM(fresh(), "") {
			t.Fatalf("expected continue")
		}
		want := "Ollama up @ http://127.0.0.1:11434 · 2 model(s) · 'qwen:7b' ready"
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("missing %q:\n%s", want, buf.String())
		}
	})
	t.Run("not pulled — menu pick saves", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "1\n")
		c.OllamaModels = func(string, float64) ([]string, error) {
			return []string{"llama3"}, nil
		}
		if !c.preflightLLM(fresh(), "") {
			t.Fatalf("expected continue")
		}
		out := buf.String()
		if !strings.Contains(out, "configured 'qwen:7b' is NOT pulled") {
			t.Fatalf("missing the not-pulled line:\n%s", out)
		}
		if !strings.Contains(out, "switched to 'llama3' — saved to config.json") {
			t.Fatalf("missing the switch line:\n%s", out)
		}
	})
	t.Run("declined aborts", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "q\n")
		c.IsATTYIn = func() bool { return true }
		c.OllamaModels = func(string, float64) ([]string, error) {
			return []string{"llama3"}, nil
		}
		if c.preflightLLM(fresh(), "") {
			t.Fatalf("expected abort")
		}
		if !strings.Contains(buf.String(), "No model selected — aborting") {
			t.Fatalf("missing the abort line:\n%s", buf.String())
		}
	})
	t.Run("no models aborts", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.OllamaModels = func(string, float64) ([]string, error) { return nil, nil }
		if c.preflightLLM(fresh(), "") {
			t.Fatalf("expected abort")
		}
		if !strings.Contains(buf.String(), "NO models installed") {
			t.Fatalf("missing the no-models line:\n%s", buf.String())
		}
	})
	t.Run("unreachable continues", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.OllamaModels = func(string, float64) ([]string, error) {
			return nil, errors.New("ConnectionError: refused")
		}
		if !c.preflightLLM(fresh(), "") {
			t.Fatalf("expected continue")
		}
		want := "Ollama not reachable at http://127.0.0.1:11434 (ConnectionError) — the pipeline will try to auto-start it"
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("missing %q:\n%s", want, buf.String())
		}
	})
}

// TestPreflightLLMCloud: the pure success line.
func TestPreflightLLMCloud(t *testing.T) {
	buf := &bytes.Buffer{}
	c := testCLI(buf, "")
	c.OllamaModels = func(string, float64) ([]string, error) {
		t.Fatal("the cloud provider never probes Ollama")
		return nil, nil
	}
	cfg := om("llm_provider", "cloud", "cloud_api_url", "http://x/v1",
		"cloud_model", "gpt-4o-mini")
	if !c.preflightLLM(cfg, "") {
		t.Fatalf("expected continue")
	}
	want := "OpenAI-compatible endpoint http://x/v1 · model 'gpt-4o-mini'"
	if !strings.Contains(buf.String(), want) {
		t.Fatalf("missing %q:\n%s", want, buf.String())
	}
}

// TestPreflightChecksComposition: the card + the section + the abort
// door (a GitHub error on a non-TTY never asks) + the happy path.
func TestPreflightChecksComposition(t *testing.T) {
	t.Run("github error aborts on non-tty", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.PreflightGithub = func(string) PreflightGHReply {
			return PreflightGHReply{Status: 401}
		}
		c.OllamaModels = func(string, float64) ([]string, error) {
			return []string{"qwen:7b"}, nil
		}
		cfg := om("github_token", "tok", "llm_provider", "ollama",
			"ollama", om("model", "qwen:7b"))
		if c.PreflightChecks(cfg, false) {
			t.Fatalf("expected abort")
		}
		out := buf.String()
		if !strings.Contains(out, "─── pre-flight checks ") &&
			!strings.Contains(out, "pre-flight checks") {
			t.Fatalf("missing the section divider:\n%s", out)
		}
		if !strings.Contains(out, "Aborting — fix the token (config.json / --init) and rerun.") {
			t.Fatalf("missing the abort line:\n%s", out)
		}
	})
	t.Run("github error continues on tty yes", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "y\n")
		c.IsATTYIn = func() bool { return true }
		c.PreflightGithub = func(string) PreflightGHReply {
			return PreflightGHReply{Status: 401}
		}
		c.OllamaModels = func(string, float64) ([]string, error) {
			return []string{"qwen:7b"}, nil
		}
		cfg := om("github_token", "tok", "llm_provider", "ollama",
			"ollama", om("model", "qwen:7b"))
		if !c.PreflightChecks(cfg, true) {
			t.Fatalf("expected continue")
		}
		out := buf.String()
		if !strings.Contains(out, "run configuration") {
			t.Fatalf("missing the card:\n%s", out)
		}
		if !strings.Contains(out, "continue anyway?") {
			t.Fatalf("missing the continue door:\n%s", out)
		}
	})
	t.Run("happy path with card", func(t *testing.T) {
		buf := &bytes.Buffer{}
		c := testCLI(buf, "")
		c.PreflightGithub = func(string) PreflightGHReply {
			return PreflightGHReply{Status: 200, Login: "ali"}
		}
		c.OllamaModels = func(string, float64) ([]string, error) {
			return []string{"qwen:7b"}, nil
		}
		cfg := om("github_token", "tok", "llm_provider", "ollama",
			"ollama", om("model", "qwen:7b"), "vault_path", "/tmp/v")
		if !c.PreflightChecks(cfg, true) {
			t.Fatalf("expected continue")
		}
		out := buf.String()
		for _, want := range []string{
			"authenticated as ali",
			"disabled — connecting directly",
			"'qwen:7b' ready",
		} {
			if !strings.Contains(out, want) {
				t.Fatalf("missing %q:\n%s", want, out)
			}
		}
	})
}
