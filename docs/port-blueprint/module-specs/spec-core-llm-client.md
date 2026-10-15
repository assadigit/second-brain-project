# Module Spec — `app/gitcurator/core/llm_client.py`

- **LOC**: 1,459 · **Go home**: `internal/llmclient` · **Status**: port
- **Responsibility**: timeout-wrapped LLM calls for the three providers, JSON extraction,
  llama.cpp/Ollama auto-detection with process-port discovery, and the startup auto-switch policy.
  The `ollama` SDK client object is injected by callers; everything else is raw urllib.

## Symbols (selected; `llm_client.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `DEFAULT_TIMEOUT_S` / `DEFAULT_LIST_TIMEOUT_S` | 31-32 | 300 / 15 |
| `_is_loopback_url` / `_urlopen_direct` | 48-82 | 127.x/::1/localhost; loopback NEVER rides proxy (ProxyHandler({})) |
| `call_with_timeout(fn, timeout_s, …)` | 89-119 | 1-worker ThreadPoolExecutor; TimeoutError `"LLM call did not respond within {n}s"`; `shutdown(cancel_futures=True)` (Go: contexts) |
| `extract_json(text)` | 179-222 | `json.loads` → ```` ```json ```` fence regex → first balanced `{…}` (brace depth); else `ValueError("No valid JSON found in response")` |
| `DEFAULT_NUM_CTX` / `DEFAULT_MAX_OUTPUT_TOKENS` / `ANTHROPIC_FALLBACK_MAX_TOKENS` | 236-245 | 8192 / 0 / 4096 |
| `is_anthropic_url` | 264-268 | substring on lowered URL |
| `CloudLLMError` family | 276-294 | `CloudLLMHTTPError(code, body[:400])` message `endpoint answered HTTP {code}: {body[:200]}` |
| `json_mode_rejected` memo | 273, 297-304 | process-lifetime set keyed by normalized base URL |
| `estimate_tokens(messages)` | 307-315 | `sum(len(content)+16)//4` warning-only |
| `resolve_task_model(config, task, default)` | 318-330 | `models.classify/analyze` overrides |
| `openai_chat(...)` | 337-475 | see wire contract; socket timeout = wall+5 (or 120); JSON-mode 400 → one clean retry, memoized |
| `openai_list_models` / `preflight_openai` | 478-533 | `GET {base}/models`; `(endpoint_ok, msg, model_listed|None)` — never a gate |
| `anthropic_chat` / `anthropic_list_models` | 540-678 | Messages wire (below); models via `x-api-key` |
| `cloud_chat` / `preflight_cloud` | 681-709 | URL router |
| `ollama_chat(client, model, messages, …)` | 716-745 | SDK kwargs: `options.num_ctx/num_predict`, `format='json'` |
| llama.cpp block | 770-1285 | `LLAMACPP_SCAN_PORTS` (11 ports); `/props` markers; `llamacpp_limits_from_props` (n_ctx/n_predict from `default_generation_settings` incl. nested `params`; -1/0→None); fingerprints (`Server: llama.cpp`, `owned_by: llama.cpp`, `.gguf` id); `probe_llamacpp`; `detect_llamacpp` (process ports first: tasklist/netstat/ss parses); `ollama_reachable` (raw TCP 1.5 s); `detect_ollama` (`GET /api/tags`, UA `GitCurator-detect`) |
| `llamacpp_autodetect_decision` | 1370-1444 | pure policy: switch only when provider=ollama+down or provider=cloud+no key |
| `resolve_llamacpp_model` | 1447-1459 | config → `/v1/models[0]` → `/props` name |

## Wire contracts (exact)

**OpenAI-compatible** `POST {base}/chat/completions` (base must include `/v1`):
`{"model", "messages", "temperature" (default 0.7)}` + `"max_tokens": int` only when >0 +
`"response_format": {"type":"json_object"}` only in JSON mode (unless memoized-rejected).
Headers: `Content-Type: application/json`, `Authorization: Bearer` (when key). Response:
`choices[0].message.content` (or `''`); embedded `error` object → `CloudLLMBadResponse
"endpoint error object: {msg}"`.

**Anthropic** `POST {base}/messages`: `{"model", "messages" (user/assistant only), "max_tokens"`
(**required**; user value or 4096), `"temperature"` + top-level `"system"` when present; JSON mode
appends to system: `"\n\nReply with ONLY a valid JSON object — no prose, no markdown fences."`.
Headers: `anthropic-version: 2023-06-01`, `x-api-key`. Response: join `content[].text` where
`type=='text'`.

**TLS**: cloud paths honor `verify_ssl` (default false → `CERT_NONE` context); loopback always
direct.

**Detection**: process-table port discovery (`tasklist /FO CSV /NH` + `netstat -ano -p TCP` on
Windows; `ss -ltnp`/`netstat -ltnp` + `pgrep` names on POSIX) for names
`llama-server/llamafile/ik_llama_server/llama-server-bin`; probes `/props`, `/health`,
`/v1/models` all direct; 2 s timeouts; never raises.

## Go notes

- Replace the executor-timeout trick with `context.WithTimeout` on the HTTP request.
- Keep every user-visible error string byte-exact.
- Detection porting: Windows `tasklist`/`netstat` parsing vs Go-native enumeration — keep the
  subprocess parsing (parity, simplicity) with `CREATE_NO_WINDOW` equivalent.

## Tests

Fake-server payload goldens (3 providers), JSON-mode retry + memoization, `extract_json` corpus,
llama.cpp props/limits parsing (incl. nested params), detection decision table, UA assertions
(`GitCurator-detect`).
