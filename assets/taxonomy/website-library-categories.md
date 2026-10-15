# Website Library — Category System

## How this works
- Two mirrored roots in the vault: `Websites/` (bookmarked links) and `Notes/` (your own writing, reflections, and evaluations) — same category names under both, so a bookmark and your own note on the same topic never conflict or compete for the same folder.
- Maximum two folder levels: **Category → Subcategory**. Anything more specific becomes a tag instead of a folder.
- GitHub repo links (exact domain `github.com`) are excluded entirely from this system — your separate GitHub Project Curator app handles those. `GitHub Projects/` is reserved as its own top-level folder in the same vault.

## The domain law (owner, 2026-10-01 — v0.28.0; second reading 2026-10-05 — v0.35.0)
**Banned from this Websites vault, always:** X/Twitter (`x.com`, `twitter.com`, `t.co`), the whole GitHub group (`github.com`, `gist.github.com`, `*.github.io`, `githubusercontent.com`), HuggingFace (`huggingface.co`, `hf.co`), Instagram, Facebook, and LinkedIn — **plus the v0.35.0 extension:** the whole YouTube group (`youtube.com`, `youtu.be`, `youtube-nocookie.com`), the Google share/drive/docs family (`share.google`, `drive.google.com`, `docs.google.com` — docs/sheets/slides, `forms.google.com`), and the social-media majors (TikTok, Threads, Snapchat, Pinterest, Twitch, Discord, Telegram `t.me`, WhatsApp, VK, Bluesky, Weibo). Only real websites (and GitHub, via its own pipeline) are ever curated. Banned links are never fetched, never turned into notes, never retried, and — v0.35.0 — never collected in the `_inbox` review-queue tables either: no note, no review, no row. Any legacy note or platform table is swept to `.trash/banned-domains` on the next run. This is law, not a setting: the "Blocked domains" field in Settings can only ADD domains. Deliberately still allowed: Reddit and Medium (content platforms), arXiv and the package registries.

---

## Categories

### Design 🌐 *(public-directory candidate — the folder tree below is the vault's own, evolved by the LLM + owner review)*
- **Assets & Resources** — general downloadable/usable design assets
- **Background Removal & Image Editors** — background removers, upscalers, and image editors
- **Backgrounds & SVG Generators** — background generators and SVG pattern tools
- **Branding & Identity** — logo tools, brand guidelines
- **Color Palettes & Tools** — palette generators, pickers, and color tools
- **Gradients & Color Assets** — gradient generators and color asset libraries
- **Icons & SVG Assets** — icon sets and SVG asset libraries
- **Image Edit & Background Tools** — general-purpose image editing utilities
- **Mockups & Templates PSD** — mockups, templates, and PSD resources
- **UI Design Inspiration** — galleries and inspiration for UI design
- **UI/UX & Product Design** — process: design systems, UX writing, wireframing/prototyping

*(v0.88.0 — the vault's Design tree grew from 4 declared subfolders to
these 11 real ones under LLM + owner review; "Print & Editorial Design"
never materialized on disk and was retired. The published directory's
category tree now mirrors the vault's folders exactly — a declared
folder with no notes never publishes, an undeclared folder publishes
with safe defaults until it is declared here.)*

Example tags: `#icons` `#mockup` `#illustration` `#vector` `#font` `#color-tool` `#gradient` `#ui-kit` `#free` `#paid` `#svg` `#3d-mockup` `#farsi-font` `#ai-powered` `#inspiration` `#design-system` `#ux-writing` `#wireframing` `#dark-mode` `#landing-page` `#mobile` `#dashboard` `#accessibility` `#brochure` `#catalogue` `#print-template` `#logo` `#brand-guideline`

### AI Tools for Web & App Development 🌐 *(public-directory candidate)*
- AI UI/Design Generation
- AI Coding Assistants
- AI Prototyping / No-Code

Example tags: `#ui-generation` `#code-gen` `#no-code`

### AI Tools (General)
- AI Writing & Content
- AI Research & Search Assistants
- AI Audio & Voice Generation

Example tags: `#tts` `#voice-cloning` `#research-ai`

### Developer Tools
*(no subcategories)*
Example tags: `#cli` `#testing` `#formatting` `#reference`

### Free Utilities & Everyday Tools
*(no subcategories)*
Example tags: `#pdf` `#converter` `#speed-test`

### Knowledge, Research & Reference
*(no subcategories)* — also the home for general, multi-topic learning platforms (Coursera, Udemy, Khan Academy). Tag with `#course`.
Example tags: `#data` `#academic` `#archive` `#course`

### Download Resources 🔒 *(the Torrent & Shadow Libraries subfolder is private)*
- **Public Archives & Libraries** — legitimate, public-domain (Gutenberg, Library of Congress, Internet Archive)
- **Torrent & Shadow Libraries** — Anna's Archive, libgen, sci-hub, Z-Library, torrent search engines

Example tags: `#ebook` `#torrent` `#shadow-library`

### Music & Audio Discovery
*(no subcategories)*
Example tags: `#radio` `#focus-music` `#sfx`

### Security & Privacy Tools
*(no subcategories)*
Example tags: `#breach-check` `#anon`

### Crypto & Markets
*(no subcategories)*
Example tags: `#tracker` `#exchange-rate`

### English Learning & Career Prep
General English-test prep (IELTS, etc.) and general career/salary research — **not** tied to a specific country. Country-specific relocation content goes under Living Abroad instead.
Example tags: `#ielts` `#immigration` `#salary`

### Geography & Urban Planning
*(empty for now — ready for future links)*
- Urban Planning Courses & Learning
- Maps, GIS & Data Tools
- Case Studies & Reference

Example tags: `#gis` `#zoning` `#transit` `#masterplan` `#course`

### Living Abroad
*(empty for now — ready for future links)*
Country-specific relocation content only: job boards, visa info, housing, cost of living for one particular country.
- Australia

Example tags: `#job-seeking` `#visa` `#housing` `#cost-of-living` `#salary`

### Random / Curiosities
Catch-all — use sparingly, only when nothing else genuinely fits.
*(no subcategories)*

---

## Judgment-call rules
*(these are also baked directly into `categorize_links.py`'s prompts, so the AI follows them automatically)*

- A single-purpose AI tool that does a design task (background removal, image upscaling, vectorizing) → **Design**, not either AI Tools category — its job is design, AI is just the mechanism. Tag `#ai-powered`.
- **AI Tools for Web & App Development** = tools that help *build* software (UI generation, code generation, no-code builders). **AI Tools (General)** = everything else AI-related with no build-a-product purpose.
- General course platforms (Coursera, Udemy) → **Knowledge, Research & Reference** + `#course`. A topic-specific course stays in its own topic folder + `#course`.
- Shadow libraries and piracy-adjacent sites → **Download Resources → Torrent & Shadow Libraries** (private). Legitimate public-domain archives → **Public Archives & Libraries**.

## Not yet covered
- ~~`*.github.io` sites and `gist.github.com` links~~ — **settled by the domain law (v0.28.0)**: the whole GitHub group is banned from the Websites vault; repo links keep going to the GitHub pipeline.
