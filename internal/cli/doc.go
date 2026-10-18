// Package cli is the command-line surface of the Go port (spec:
// docs/port-blueprint/module-specs/spec-cli.md; task T-60, Phase 5 —
// re-scoped per ADR-0004 Am.3 "OPTION B CONFIRMED" / rider R-0: the
// direct Telegram (MTProto) leg is gone, the bot queue rides the
// Cloudflare Worker HTTP API only, and mirror-push is the read
// acknowledgement).
//
// T-60's first slice (SESSION-027): THE CLI SHELL — the paint box
// (terminal.go: the C ANSI system with the NO_COLOR/TTY/--no-color
// law, level styles, term_width, mask, rule, cli_print, banner,
// render_bar, StatusLine, _strip_ansi, the UTF-8 console hardening),
// the settings plumbing (settings.go: _config_path / load_config /
// save_config over the T-12 config port, _ask, _mask_token,
// _clean_config_for_save), THE SYNC DOOR's pure halves (syncdoor.go:
// the failure card, sync_missing_essentials, sync_confirm_prompt) +
// beat 1a wired, the honest Option-B flag surface + dispatch order +
// the friendly help card (cli.go), the FULL --status card (status.go:
// the 16 config rows, the taxonomy row, the sentinel rows, the
// Ollama/llama.cpp checks, the CacheDB/NoteStateDB/WebsiteStateDB
// stats), and --list-dead/--reset-dead (dead.go). Verified by the
// generated differential golden testdiff/goldens/dump_cli.py →
// testdata/goldencli.golden.json (driving the reference's own CLI
// shapes with frozen clocks + folded paths) + internal/cli tests.
//
// The honest flag surface (T-60's decision per the spec's banner,
// never a silent no-op): implemented commands run; the Telethon-only
// flags (--login, --mark-read, --from-id/--to-id,
// --offset-start/--count, --single-id) answer explicit refusal cards
// (accepted losses, ADR-0004 Am.3); the wizard (--init),
// --test-connection, --detect-llm, --hand-delivery, --chrome-retry
// and the batch engines (--auto, --sync's live beats, --retry-failed,
// --import-file) answer explicit in-development cards naming their
// landing task (the rest of T-60 / T-62). The Telethon-only flags
// stay RECOGNIZED so old scripts and launchers get the honest card,
// never an unknown-flag crash.
//
// The banner's stale "v0.09.5" version string is a preserved Python
// quirk (cli_terminal.py banner(); the house rule is quirks-preserving
// parity, see 09-testing-and-parity.md §5). The Phase-0 seed's
// --status "no config" path (verified by testdiff/cases/
// status-no-config) is preserved byte-exactly inside the full card.
package cli
