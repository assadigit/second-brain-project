// Package githubpipeline ports the GitHub side of the GUI batch
// worker: gui/processing_worker.py's per-repo loop (947-1725), the
// banner download + the GitHub-side enrichment
// (gui/worker/github_meta.py), the LLM analysis ladder
// (gui/worker/llm.py:439-720 + the model-choice helpers), the
// missing-repo placeholder (gui/worker/notes.py:123-161), the reports
// (gui/worker/reports.py), and the batch undo (the snapshot + the
// _undo_last_batch.txt diff). Spec:
// docs/port-blueprint/module-specs/spec-gui-processing.md. Task T-42
// (Phase 4).
//
// The companion extension of internal/batch (T-25's LinkTracker +
// VaultIndex) is batch/summary.go — the v0.39.0 scorecard
// (_build_batch_summary).
//
// # Layout
//
//   - ghclient.go   the GitHub API seam (the RepoFetcher interface —
//     the "fake repo" of the spec's test contract) + the production
//     REST client (api.github.com over net/http, 15 s timeout,
//     PyGithub's Accept header, standard TLS verification — the
//     reference never threads netctx into the GitHub client, only
//     into the banner download)
//   - runner.go      the Runner: the per-repo loop with the gates in
//     their exact order (non-GitHub → dead-set → dismissed → settled
//     → invalid → vault-index dedupe BEFORE any API call), the get_repo
//     error paths (403+rate wait 3605 s → retry; 401 → drop token →
//     anonymous retry; 404 → the strike + the _missing note + the
//     notfound-links.md append; other → manifest failure), the
//     two-layer dedupe, fork→parent, the enrichment, the banner
//     goroutine (joined 15 s), the model re-read per URL, the
//     review/quality routing, the atomic write with the disk-full
//     protocol (headless skip vs the GUI pause), and the bookkeeping
//     chain
//   - meta.go        OrgReputation, CalculateCredibility,
//     RecordMissingRepo, BackfillMissingNotes
//   - banner.go      DownloadBanner (the throttle, the 429 backoff
//     ladder, the 24 h failed-marker cache, the atomic write, the
//     netctx verify_ssl split)
//   - llmanalyze.go  the 3-attempt ladder, the <<<README_BEGIN>>>
//     injection guard, the per-provider routing (models.analyze
//     overrides), the failure decision door (the headless half lives
//     here; the GUI half is the LLMDecisionFn seam), PickBestModel,
//     ApplyModelChoice, LooksLikeConnectionError
//   - reports.go     the summary log (+ the summary_keep_last
//     rotation) and the final Markdown report
//   - undo.go        the batch snapshot, the new-files diff, the
//     _undo_last_batch.txt writer, and the Finish completion chain
//
// # The seams (the GUI shell wires them; the tests freeze them)
//
// Runner.Clock / Runner.Sleep freeze time for the rate-limit laws;
// Runner.NewGHClient swaps the GitHub client (the 401-anonymous
// fallback re-enters through it); callCloudLLM is the _call_cloud_llm
// staticmethod seam (the reference's own tests monkeypatch the same
// spot); Runner.BannerFn is the fake banner; Runner.writeNote injects
// write failures for the disk-full protocol; Runner.writeNote and
// storage.AtomicWriteText are dry-run gated exactly like the
// reference's storage helpers.
//
// # Documented divergences (reality-wins record)
//
//   - The empty-attempt-2 lie: llm.py's comment says an empty attempt
//     2 goes "straight to Attempt 3", but the raise escapes the
//     except-handler into the OUTER failure door. The port follows
//     the CODE (the door), not the comment.
//   - A non-object JSON payload from the LLM is treated as a parse
//     failure (attempt 2). The reference would raise TypeError at
//     the defaults fill and land in the same failure door one step
//     later — the recovery family is identical, the timing differs
//     by one attempt.
//   - about_me.md / system_prompt.txt read errors other than
//     not-exist degrade to "absent" in the port (the reference only
//     swallows FileNotFoundError; a permission error would fail the
//     repo). Pathological in practice.
//   - The _calculate_credibility doc probe takes the loop's README
//     fetch result instead of re-issuing get_readme (the reference
//     re-probes, best-effort and guarded — the same fact, one fewer
//     API call).
//   - The disk-full substring law gains Go's own errno spelling
//     ("no space left on device", lowercase) beside Python's
//     strerror "No space" — the law is the shape, not the casing.
//   - notebuilder's CredScore became float64 (this task's
//     reality-wins fix): _calculate_credibility returns round(x, 1)
//     and every real note renders "90.0/100", never "90". The
//     notebuilder golden was regenerated with the production calling
//     convention.
//   - The primary-language tag merge keeps the reference's
//     case-sensitive membership quirk: the test is against the
//     ORIGINAL spelling ("Go") while the INSERTED tag is the
//     lowercase ("go") — a "Go" primary inserts "go" even when
//     "go" is already merged (the duplicate stays in all_tags; the
//     note's own sanitizer dedupes on the way into the frontmatter).
//   - A GitHub-level API error (the generic else branch) marks the
//     MANIFEST only — no retry-queue row (the reference's per-URL
//     except owns add_failed; the next launch's reconciliation owns
//     the retry).
//   - The vault-index layer is batch.VaultIndex (T-25) and the
//     manifest is batch.LinkTracker — extend, never duplicate.
//
// # The golden
//
// testdata/githubpipeline.golden.json is generated by
// testdiff/goldens/dump_githubpipeline.py from the unmodified
// reference (pinned rev) with the reports clock frozen at
// 2026-10-17 12:00:00. The GUI-layer imports (PyQt6 / PyGithub /
// ollama) are satisfied by minimal sys.modules stubs — the reference
// code is unmodified; the stubs only satisfy import machinery (the
// reference's own mock.patch discipline; the pinned venv is
// stdlib-only, so the GUI worker cannot import otherwise).
package githubpipeline
