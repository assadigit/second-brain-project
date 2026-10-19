# Module Spec — GUI visual system: `gui/theme.py`, `gui/main_window/theme.py`, `gui/icons.py`, `gui/sound.py`, `gui/constants.py`

- **LOC**: 681 + 528 + 178 + 168 + 258 · **Go home**: `ui/frontend/src/theme/` (CSS custom
  properties + component classes) + `internal/app/sound.go` · **Status**: rehost (tokens
  become CSS variables; the Qt mechanics do not port)
- **Responsibility**: the single app theme (light/dark token tables + QSS builder + dynamic
  property grammar), the Lucide icon pack, the batch chime, and the legacy constants twin.

## 1. Theme token tables (source of truth — `gui/theme.py:58-237`)

Reproduced **in full** in this spec's companion table (also embedded in the frontend as CSS
variables). Headline values:

| Token | LIGHT | DARK |
|---|---|---|
| window | `#FBF8F2` cream | `#221E2E` plum-charcoal |
| sheet / sheet_alt / sheet_hover | `#FFFFFF` / `#F2EDE3` / `#EAE3D6` | `#2B2639` / `#352F4A` / `#352F4A` |
| border / border_input / border_hover | `#EAE3D6` / `#D8D0BE` / `#B5AC9C` | `#3B344F` / `#4A4263` / `#5C5378` |
| text / text_soft / text_muted / text_disabled | `#423A52` / `#514A63` / `#6C6480` / `#A79F92` | `#F2EEE7` / `#DDD7EC` / `#B7AFC9` / `#7E7794` |
| accent / focus | `#5F54B4` / `#8B80D6` | `#C4BCF5` / `#C4BCF5` |
| select_bg / select_fg / on_accent | `#5F54B4` / `#FFFFFF` / `#FFFFFF` | `#5F54B4` / `#FFFFFF` / `#221E2E` |
| success / warning / error / muted | `#1E6B4B` / `#75510A` / `#AE2237` / `#6C6480` | `#7CE2A9` / `#FFD37E` / `#FF9AAB` / `#B7AFC9` |
| log_well / log_tint / log_text / log_ts | `#FFFFFF` / `#F2EDE3` / `#57506B` / `#6C6480` | `#17131F` ×2 / `#C6BFE0` / `#8F89A3` |
| banner_warn_bg | `#FAF3E3` | `#322B20` |
| sync_ok_bg / border | `#ECF5EF` / `#CFE7D9` | `#25352B` / `#33503C` |
| plan_del bg/border; plan_grow; plan_mag | rose / mint / teal pairs | dark twins |
| progress_track / progress_chunk | `#E3DACA` / **`#7C3AED`** | `#17131F` / **`#A78BFA`** |
| bar_saved / bar_retry | `#22C55E` / `#F59E0B` | `#4ADE80` / `#FBBF24` |
| cta_fill / hover / text | `#B9E3C9` / `#A8DABA` / `#17402B` | same |
| err_fill / hover / text | `#F6C6CD` / `#F0B2BC` / `#5E1120` | same |
| hero_fill / hover / text | `#5F54B4` / `#514699` / `#FFFFFF` | `#C4BCF5` / `#B3A7F2` / `#221E2E` |
| danger_fill / hover / press | `#D63A24` / `#C43320` / `#B32D1D` | same |
| ghost_bg / hover_bg / hover_fg | `#FBF8F2` / `#F2EDE3` / `#57506B` | `#2B2639` / `#352F4A` / `#DDD7EC` |
| scroll / scroll_hover | `#DCD4C4` / `#B5AC9C` | `#7A7199` / `#8D84AD` |
| tab_selected / tab_hover | `#FFFFFF` / `#EAE3D6` | `#352F4A` / `#352F4A` |
| msg_success / error / warning | `#16A34A` / `#DC2626` / `#EA580C` | `#4ADE80` / `#F87171` / `#FB923C` |
| warn_box border/text | `#F5E3C0` / `#B45309` | `#4A3F28` / `#F2DCA8` |
| note_bg / note_fg / note_indigo | `#E9F5EE` / `#423A52` / `#EEEBFA` | `#26332D` / `#DDD7EC` / `#2E2A4A` |
| icon_border / hero2_bg | `#D8D0BE` / `#FFFFFF` | `#4A4263` / `#352F4A` |
| log_checked_hover | `#514699` | `#D3CDF9` |
| placeholder / toolbtn_press | `#7A7288` / `#514699` | `#A6A2AC` / `#514699` |

Full table incl. disabled pair, outline pair, pipeline_caption: `gui/theme.py:58-237`.

## 2. Type scale & families

Base `'Segoe UI','SF Pro Display','Helvetica Neue',Arial` 13px; mono `'Consolas','Monaco',
'Menlo','Courier New'` 12px (log); Inter-Bold registered from `assets/fonts/`.
ObjectName scale: settings_title 20/800 · sync_hero 24/800 · sync_cell_num 20/800 ·
plan_heading 15/800 · card_heading 14/700 · msg_heading 16/700 · cc_row_name 13/700 ·
sync_stat_value 14/700 · cc_item/status 12 (11 in plan rows) · micro 11/10 · msg_glyph 32 ·
logo_title 14/800 · progress_count mono 15/700. Layout tokens: spacing 4/8/16/24/32/48; focus
2px outline + 2px offset; help '?' 22×22.

## 3. Button grammar (`btn_kind` property; 601-659)

`primary` (mint fill, 8×24 padding, r6, 13 bold) · `secondary` (accent outline) · `danger`
(rose fill) · `ghost` · `icon` (16 glyph, r8) · `help` · `hero_primary` (14/800) · `hero_danger`
· `hero_secondary`. **Exactly one filled primary per view** (Backup Now / Process All / Apply /
Mark-as-intentional / OK / Scan).

## 4. icons.py

Lucide-static **v0.544.0** (ISC) embedded, **28 glyphs** (layers, settings, moon, sun, refresh-cw,
stop, play, loader-circle, activity, search, trash-2, dot, circle, check, triangle-alert,
circle-x, info, key-round, globe, folder, brain, inbox, list, bar-chart-3, bot,
satellite-dish, archive, volume-2 — the pack's real count; the "30" here earlier was a
drafting slip, fixed at T-72's landing when the golden counted them). Render-time tint via
`{color}` substitution; 2× DPR; explicit target rect (the v0.08 clipping fix); two-mode nav
icons (Selected state); graceful degradation to empty icons. **Go**: embed the same SVGs; tint
server-side or via CSS `currentColor`. **Landed (T-72)**: the pack is golden-locked
byte-for-byte (testdiff/goldens/dump_icons.py → webui/uitheme/testdata/icons.golden.json);
the Go twin (webui/uitheme/icons.go) and the frontend twin (webui/frontend/icons.js, tinted
via currentColor) both diff against that one golden.

## 5. sound.py

`QSoundEffect` lazy QtMultimedia; `success.wav` natural completion / `retry.wav` failures;
five laws (batch never waits; lazy import; one cached effect per asset; no-backend ≠ error;
missing asset poisons silently after one notice). Config `sound_enabled`, `sound_volume`
(0–100 → 0–1, triple-clamped). **Go**: `oto`/`beep` or OS-native play; keep the laws and
defaults (0.8).

## 6. main_window/theme.py (router)

Thin routers onto the kit + Inter loading + help button factory + scroll wrapper
(`_wrap_scroll` vertical-only, `tab_sheet` objectName, trailing stretch) + 200 ms dialog fade
(parented to dialog or GC collects it; 10 ms start delay; 500 ms opacity failsafe) + custom
message/question modals (zombie guards) + theme-toggle re-tint pass (baked pixmaps + log
re-render + Backup dots). Window geometry: main 1000×750; Settings fixed 880×640.

## 7. gui/constants.py

Stale duplicate (258 LOC; older palette) — **not ported**; test-only importers noted.

## Go parity requirements

1. CSS custom properties generated from the token tables (light + dark `[data-theme]`) —
   a golden test diffs the CSS var dump against this table.
2. Type scale + spacing + focus laws as CSS.
3. One-primary-per-view enforced by component usage (checklist in UI workflow tests).
4. Icon pack embedded with tint support; sounds embedded with the five laws.
5. Dialog fades 200 ms; no zombie modals after close (guard in the UI layer).
