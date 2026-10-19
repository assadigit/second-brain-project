# SESSION-036 — 2026-10-19 — Phase 6: T-73 COMPLETE (Settings pages A — the six pages live on the production shell; the live-config law + the four save laws + the six pages' engines) — 2020 main-battery tests over 35 packages + 23 webui tests over 4; the thirty-first full wipe repaired BEFORE work

## Goal (from NEXT-SESSION.md at start)

SESSION-036 = **T-73 (L): SETTINGS PAGES A** per doc 13's row +
spec-gui-pages-a (READ FIRST — the authority): Vault, Credentials,
Proxy, LLM, Input, Sound — the six pages on the production shell
(webui/), driving internal/app's command surface; the save-on-change
law (editingFinished/toggled → the page savers); the golden-locked
theme tables + the 28-glyph icon pack ready. ACCEPTANCE: the page
workflow tests + the save-law (merge/in-place) tests. The CLI-parity
shell wiring (warmup family, inputs fetchers, websites phase, Phase-5
CLEAR) lands on the app surface at this binding. If capacity runs
short: close cleanly at a natural slice boundary.

## What was done (task IDs, files, commits — cite hashes)

- **The thirty-first full wipe repaired BEFORE work** (both repos, the
  toolchain, the venv, the module cache, AND the webkit tree all gone
  at boot; everything rebuilt from the committed recipes): both repos
  re-cloned with the ephemeral `GIT_CONFIG_COUNT` form (no PAT
  persisted anywhere; the push-side `-c http.extraheader` form used
  only at push time); go1.27.2 sha-verified THREE ways (local sha256 =
  go.dev's published JSON = the committed 64-char pin, `ecbadb99091a…`);
  the venv (Python 3.12.15); the wails CLI v2.16.0; the rootless
  webkit tree (144 debs → dpkg -x → the .pc prefix rewrites → **exactly
  26 dangling dev symlinks relinked to the system runtime copies** —
  the same count SESSION-035 recorded; the fresh base image again
  ships the gtk/atk runtimes system-wide). The re-verification law
  executed BEFORE any new work: MAIN go.mod byte-verified (the
  authorized four @ go 1.24.1; the nested webui @ 1.25.0 + wails +
  the parent replace), the full main battery (35 packages race-clean)
  + the webui battery (4 packages race-clean) + vet/gofmt + the
  clean-tree check all green pre-work.

- **T-73 slice (a) — commit `595a362`: internal/app/settings.go, the
  settings command surface.**
  - **THE LIVE-CONFIG LAW** (the session's biggest architectural
    decision, the reference's own MainWindow semantics):
    `App.loadConfig` previously re-read config.json on every call
    (the CLI's shape). The GUI reference loads ONCE at initUI
    (`self.config = self.load_config()` — fresh detection, the example
    base, THE FILL, the v0.07.1 credential strip heal) and holds the
    same dict for the session; `save_config` mutates it IN PLACE so
    running workers see saves. The port now matches: `App` holds
    `liveCfg` (loaded once via `config.LoadAppConfig` at `New()`),
    every door reads that instance, and every save mutates it
    clear+update (never a rebind). One prior wiring test codified the
    old semantics (missing file → the EMPTY map); the reference's
    actual behavior is `deepcopy(CONFIG_EXAMPLE)` — the assertion was
    corrected to pin the reference's behavior + the same-instance law
    (reality wins; the reference is the authority, not the old test).
    `App.TheFillLine()` carries window.py's v0.69 quiet desk line (the
    fill count) for the shell to speak after the UI exists.
  - **The six pages' engines** over injectable `SettingsDeps` (the
    hermetic law — every network/process/filesystem seam faked in the
    tests): THE ONE merge-based save with ALL FOUR v30 laws (MERGE
    never whitelist-rebuilds — unknown/tuning keys survive;
    a page group carries only its own keys, an absent group keeps the
    config value — the reference's defensive hasattr pattern kept as
    the group-pointer law; the LIVE map mutates IN PLACE; the atomic
    write + the credentials mirror ride `config.Save` BEFORE the
    mutation so a failed write leaves the live map untouched; the
    250 ms sentinel re-arm via the new `Lifecycle.RunSentinelPass`).
    `log_level` keeps the reference's one hardcoded quirk ("INFO" on
    every save). Plus: RemoveVault (the combo's history law), the
    five-rung `VaultPathStatus` ladder, the `Snapshot` builder
    (populate_vaults' dedupe + the no-pre-selection law + the
    blocked/self fields showing the LAW+extras union — ui.py reads
    `blocked_domains_from_config`, not the raw extras), TestGitHubToken
    (GET /user with the field AS TYPED; the 401/403/network lines
    verbatim; `GitHubAPIBase` = the PyGithub default), the About Me
    wizard's pure template + the CWD write (the same door the runner
    reads), RefreshOllamaModels (the v30 in-place `ollama.model` sync +
    the custom-name combo law + the log counting the SERVER's list),
    StartOllamaServer (the detached `ollama serve` spawn twins added to
    process_posix.go / process_windows.go — setsid / the Windows
    process-group flags, the FileNotFoundError token), llama.cpp
    Detect/Refresh (the `_fill_llamacpp_models` law — pick || current
    || first — + THE SERVER CONFIG GRAB riding every probe + the
    best-effort save), the Detect & Set fast lane (the batch-running
    guard over the OpLock; the live-URL snapshot law — a URL the user
    just typed is probed; the single-model auto-apply; the
    several-models menu door; the honest cancel line; the SET law —
    provider+URL+model in place + merge-save with the LIVE-FIELDS
    OVERLAY so the untouched engines keep their truth), PreviewSound
    (the exact batch gate's decision half: the switch, the clamped
    volume with the garbage/NaN→0.8 heal, BatchSound's one-notice
    missing-asset law; the wav bytes return as a data: URL for the
    frontend's audio element).
  - **settings_test.go — the acceptance battery, 26 tests**: the
    save-law tests (merge survives unknown/tuning keys; the
    RUNNING-WORKER FAKE watches the in-place mutation through the SAME
    map instance and the same-instance law; absent groups keep config —
    bot_token/backup_*/quarantine keys survive a sound-only save;
    the log_level quirk; the lenient readers; the failed-write
    isolation; the sentinel re-arm) + the page workflow tests (the
    token's five flows, the About Me template's three paths + the CWD
    write, the Ollama refresh/spawn flows, the llama.cpp
    detect/refresh/grab paths + the current-model law, the quick-switch
    guard/menu/apply/cancel paths + the live-overlay proof, the sound
    gate's three doors, the ladder's every rung, the snapshot's
    dedupe/defaults/heal, the page-commit group laws).

- **T-73 slice (b) — commit `48157da`: the Settings window on the
  production shell.**
  - webui/settingsui.go — the thin-shell binding wrapper (doc 04 §3:
  "ui — thin shells over app"): every door delegates to
  `internal/app`'s Settings surface; plus the two native pickers
  (the import-file dialog with the reference's filter family + the
  three vault-folder dialogs with the reference's own titles), bound
  as the second wails struct; the generated bindings land in
  frontend/wailsjs/go/main/SettingsUI.js (+ models.ts for the result
  types). Shell.js regenerated byte-identical (only the known 755
  mode flip — restored via chmod, per the WAILS-BUILD MODE FLIPS law;
  no stray webui/ binary left behind).
  - frontend/settings.js + index.html + style.css — the window's
  frontend half (dialogs.py's SettingsDialog + ui.py's six pages,
  the HTML twin): the gear door (the 34×34 chrome button, the
  reference's tooltip verbatim), the 880×640 window (header: title +
  the section hint + Close; the More-menu seat stays empty for T-77),
  the 190px sidebar (six rows in the reference's order — Credentials,
  Proxy, Vault, LLM, Input, Sound — the icon pack's own section
  glyphs with the two-mode tint, the TAB_INFO hints as both the nav
  tooltips and the following header hint), the master-detail stack.
  The six pages render from the snapshot with every label, tooltip
  and '?' help text verbatim: the Credentials fields (password echo
  on the token) + Test GitHub Token (as-typed) + the About Me Wizard
  (the six-field dialog) + the user-config quiet truth + the Show
  door; the Proxy form (the type combo, the use_for_web checkbox with
  its full v0.43 both-doors tooltip); the Vault page (the combo with
  the datalist + Browse/Remove with the themed confirm, the
  Websites/Manual groups with the LIVE status ladder per keystroke
  through the Go door, the blocked/self fields, the pipeline
  switches); the LLM page's three flat cards (the host/engine radios
  with the visibility law — the initial paint never saves; the limits
  fields; the quick-switch buttons; the Ollama URL/model + Refresh +
  Start Server with the 3000 ms singleShot twin; the llama.cpp
  URL/key/model + Detect (the page's ONE primary) + Refresh; the
  cloud card); the Input page's picker + the PROCESS hint; the Sound
  page (the toggle, the slider with the %-label-only drag law + save
  on release, both previews through the exact batch gate). THE
  SAVE-ON-CHANGE LAW: editingFinished/toggled → the page savers (one
  commit = one group = one merge save). The mini-dialog kit (the
  themed custom-question family's twin) carries the remove-confirm,
  the wizard, and the quick-detect model menu.
  - shell.go — the v0.69 fill line speaks after the startup passes
  (through the desk's own transport); main.go binds the second
  struct; the SB_POC_OPEN_SETTINGS autorun probe opens the window for
  the gate screenshot (the evidence door — same family as the flood
  probe).
  - settingsui_test.go — the wiring smoke (the snapshot/ladder/save
  doors through the wrapper + the import line on the bus) + the
  section-glyphs-in-pack law.

- **Spec updates riding the landing (reality wins, same commit)**:
  spec-gui-pages-a's vaults_config section gained the dated
  commit-cadence note (the port commits all six pages per edit where
  the reference saves creds/proxy/LLM only at closeEvent — every
  reference-persisted value persists, some a commit earlier); doc 13's
  T-73 row COMPLETE; doc 12's SESSION-036 entry; README's SESSION-036
  status; webui/README.md's map; internal/app/doc.go's file map.

## Gate evaluation (pass/fail/partial + EVIDENCE)

**PASS.** T-73's acceptance = the page workflow tests + the save-law
(merge/in-place) tests — both green with proof:

- **The save-law tests**: 26 settings tests in internal/app, the
  centerpiece being the RUNNING-WORKER FAKE (holds the same map the
  app holds, watches the in-place mutation through it, and the
  same-instance law proven — the map is never rebound); the merge
  never drops unknown/tuning keys (cloudflare_install_id, hmac_secret,
  user keys, timeout_per_repo survive); absent groups keep config;
  the failed-write isolation; the 250 ms sentinel re-arm observed.
- **The page workflow tests**: every engine door's flows tested over
  fake seams — the token (empty/200+login/401/403+payload/network),
  the wizard (the template's three paths + the CWD file write + the
  log lines), the Ollama refresh (custom-name survival + the v30
  in-place sync), the spawn (PID + the FileNotFoundError line), the
  llama.cpp detect/refresh/grab (the /props limits land in the config
  + the one success line + the current-model law), the quick-switch
  (guard/menu/cancel/SET + the live-overlay proof — cloud_api_key
  survives a llamacpp SET), the sound gate (off → 🔇; missing asset →
  one 🔕 notice then session silence; present asset → the play payload
  with the bytes round-tripped), the ladder's every rung.
- **The end-to-end evidence**: the wails build links clean over the
  rootless webkit tree; the gate leg re-run GREEN — first frame
  **963 ms ≤ the 1.5 s gate** (the box settling post-rebuild), the
  10k flood through the FULL production transport at **526,315
  lines/s producer-side, 0 drops, 170 ms drain**; the screenshot
  **VLM-verified**: the Settings window with all six sidebar rows
  (Credentials selected on the accent fill), the Credentials page's
  fields + buttons, the example-base values (API ID "0" — the str()
  render parity), and the user-config quiet-truth line VERBATIM
  ("Credentials are saved inside this app folder only — no Documents
  folder was found, so nothing carries over between versions.").
- **The battery at the close**: **2020 main-module tests race-clean
  over 35 packages** (1994 + 26 settings — arithmetic closed) + **23
  webui tests over 4 packages** (21 + 2 settingsui — arithmetic
  closed); gofmt/vet clean over both modules; the Windows cross-build
  of both modules compiles (GOOS=windows); the MAIN go.mod
  byte-unchanged all session; no mode flips in the landed tree (the
  one known wailsjs flip restored via chmod); the byte-level secret
  scan clean on every commit.
- **The carried debt, honestly recorded (the one partial)**: the
  CLI-parity shell wiring (warmup family, inputs fetchers, websites
  phase, Phase-5 CLEAR) did NOT land at this binding. It is the
  largest remaining composition port in Phase 6 (~engine.go's runImpl
  wiring + its own test battery — realistically a session of its own);
  the one-slice-properly law applies; and it gates T-77's hero batch
  composition, NOT the six pages (whose acceptance is fully green).
  Recorded in doc 13's T-73 row + the state card as the T-74+
  binding's first item.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the dated
  commit-cadence note at the vaults_config section (the port's
  per-page commit savers vs the reference's closeEvent-only save for
  creds/proxy/LLM; the owner's task brief words the law as
  editingFinished/toggled → page savers — the note records both).
- `docs/port-blueprint/13-implementation-task-list.md` — the T-73 row
  COMPLETE (the full record incl. the carried debt).
- `docs/port-blueprint/12-migration-roadmap.md` — the SESSION-036
  entry.
- `README.md` — the SESSION-036 status entry.
- `webui/README.md` — the settings surface in the map (settingsui.go +
  settings.js).
- `internal/app/doc.go` — the file map + the T-73 landing note.
- `internal/app/app_test.go` — the missing-config assertion corrected
  to the reference's own behavior (deepcopy(CONFIG_EXAMPLE) + the
  same-instance law) — reality wins over the old test's codified
  CLI-shape semantics.

## Blockers & decisions needed

- No new blockers. The authorized runtime set is UNCHANGED (the four;
  the MAIN go.mod byte-unchanged all session — verified at every
  commit). No new dependencies.
- Carried (no rush): the dissolution call (ADR-0002 Am.5 — keep the
  nested boundary vs amend the set); the vanilla-vs-vite call; the
  go-directive acknowledgement (twenty-seventh session carrying it);
  x/net/html keep-or-swap.
- **The CLI-parity shell wiring** — now the FIRST item for the T-74+
  binding (see Handoff).

## Owner actions requested

- None urgent. The two standing calls (dissolution, vanilla-vs-vite)
  remain open; T-74 proceeds on the nested boundary either way.

## Handoff

- **SESSION-037 = T-74 (L): SETTINGS PAGES B** per doc 13's row — Bot,
  Dashboard, Sources, Backup (+ radar, fast sync) — the pages by NAME
  against their spec files (bot_queue + dashboard live in
  spec-gui-pages-a; backup_seal in spec-gui-pages-b — the R-0 dated
  note). **The session's first item: the CLI-parity shell wiring on
  the app surface** (warmup family, inputs fetchers, websites phase,
  Phase-5 CLEAR — engine.go's runImpl composition re-hosted on
  internal/app with its own tests), landed as its own slice before or
  alongside the pages, closing the debt T-73 carried.
- The state card v37 (`docs/NEXT-SESSION.md`) carries the full
  read-first list + the SESSION-037 prompt.
- Push state: slices (a) and (b) + the close committed and pushed (the
  ephemeral PAT at each push; never stored, never written).
